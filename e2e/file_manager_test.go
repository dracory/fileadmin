package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestFileManagerLoadsSeeded(t *testing.T) {
	body := waitTable(t, newTab(t), adminURL("file-manager"))
	if !strings.Contains(body, seedDir) {
		t.Fatalf("expected seeded directory %q in file manager", seedDir)
	}
	if !strings.Contains(body, seedFile) {
		t.Fatalf("expected seeded file %q in file manager", seedFile)
	}
}

// TestNavigateIntoDirectory verifies clicking a directory name
// navigates into it and lists its files.
func TestNavigateIntoDirectory(t *testing.T) {
	ctx := newTab(t)

	err := chromedp.Run(ctx,
		chromedp.Navigate(adminURL("file-manager")),
		chromedp.WaitReady("table tbody tr:nth-of-type(2)", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("file manager did not load: %v", err)
	}

	// Directories render before files; the first row is the
	// alphabetically-first directory ("documents"). Its name is a link
	// in the third column.
	err = chromedp.Run(ctx,
		chromedp.Click(`table tbody tr:nth-of-type(1) td:nth-of-type(3) a`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		t.Fatalf("directory navigation failed: %v", err)
	}

	var body string
	if err := chromedp.Run(ctx, chromedp.Text("body", &body)); err != nil {
		t.Fatalf("reading file manager failed: %v", err)
	}
	if !strings.Contains(body, "readme.txt") {
		t.Fatal("expected 'readme.txt' after navigating into documents")
	}
}

// TestCreateDirectoryFlow verifies the "New Directory" button opens the
// create modal, and that submitting it creates a directory that then
// appears in the file table.
func TestCreateDirectoryFlow(t *testing.T) {
	ctx := newTab(t)
	dirName := "e2e-new-dir"

	err := chromedp.Run(ctx,
		chromedp.Navigate(adminURL("file-manager")),
		chromedp.WaitReady("table tbody tr:nth-of-type(2)", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("file manager did not load: %v", err)
	}

	// "New Directory" is the only .btn-info on the toolbar.
	err = chromedp.Run(ctx,
		chromedp.Click(`button.btn.btn-info`, chromedp.ByQuery),
		chromedp.WaitReady(`input[placeholder="directory-name"]`, chromedp.ByQuery),
		chromedp.SendKeys(`input[placeholder="directory-name"]`, dirName, chromedp.ByQuery),
		chromedp.Click(`.modal-footer button[type="submit"]`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		t.Fatalf("create directory flow failed: %v", err)
	}

	var body string
	if err := chromedp.Run(ctx, chromedp.Text("body", &body)); err != nil {
		t.Fatalf("reading file manager failed: %v", err)
	}
	if !strings.Contains(body, dirName) {
		t.Fatalf("expected newly created directory %q in file manager", dirName)
	}
}

// TestDeleteFileFlow verifies deleting a file via its row's gear
// dropdown and the confirm modal. It creates its own file first so the
// test is independent of shared seed state and safe to re-run.
func TestDeleteFileFlow(t *testing.T) {
	ctx := newTab(t)

	fileName := "e2e-delete-" + time.Now().Format("150405.000000") + ".txt"
	if err := testStorage.Put(rootDir+"/"+fileName, []byte("delete me")); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	err := chromedp.Run(ctx,
		chromedp.Navigate(adminURL("file-manager")),
		chromedp.WaitReady("table tbody tr:nth-of-type(2)", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("file manager did not load: %v", err)
	}

	// Find the row for the seed file by text — row positions shift as
	// other tests create directories in the shared store. Its Actions
	// cell has a Bootstrap dropdown; the Delete item is
	// .dropdown-item.text-danger.
	var clicked bool
	err = chromedp.Run(ctx,
		chromedp.Evaluate(rowFindJS(fileName, `querySelector('.dropdown-toggle').click(); true`), &clicked),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Evaluate(rowFindJS(fileName, `querySelector('.dropdown-item.text-danger').click(); true`), &clicked),
		chromedp.WaitReady(`.modal .modal-title`, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("delete modal did not open: %v", err)
	}

	var body, title string
	if err := chromedp.Run(ctx,
		chromedp.Text("body", &body),
		chromedp.Text(`.modal .modal-title`, &title, chromedp.ByQuery),
	); err != nil {
		t.Fatalf("reading delete modal failed: %v", err)
	}
	if !strings.Contains(body, "Confirm File Delete") {
		t.Fatalf("expected 'Confirm File Delete' modal title, got %q", title)
	}

	err = chromedp.Run(ctx,
		chromedp.Click(`.modal-footer .btn-danger`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		t.Fatalf("delete confirmation failed: %v", err)
	}

	if err := chromedp.Run(ctx, chromedp.Text("body", &body)); err != nil {
		t.Fatalf("reading file manager failed: %v", err)
	}
	if strings.Contains(body, fileName) {
		t.Fatalf("expected file %q to be deleted", fileName)
	}
}

// rowFindJS returns JS that finds the table row whose text contains
// name and evaluates inner against it. Row positions shift as tests
// mutate the shared store, so rows are located by text, not index.
func rowFindJS(name, inner string) string {
	return `[...document.querySelectorAll('table tbody tr')]` +
		`.find(tr => tr.textContent.includes('` + name + `'))` +
		`?.` + inner
}

// dropdownItemJS returns JS that clicks the dropdown item containing
// label inside the row for name (used after the gear dropdown opens).
func dropdownItemJS(name, label string) string {
	return `(() => {` +
		` const tr = [...document.querySelectorAll('table tbody tr')]` +
		`.find(tr => tr.textContent.includes('` + name + `'));` +
		` if (!tr) return false;` +
		` const item = [...tr.querySelectorAll('.dropdown-item')]` +
		`.find(i => i.textContent.includes('` + label + `'));` +
		` if (!item) return false;` +
		` item.click(); return true })()`
}

// setVueInputJS returns JS that sets an input's value and dispatches
// the input event so Vue's v-model picks it up (SendKeys would append
// to the prefilled value instead of replacing it).
func setVueInputJS(sel, value string) string {
	return `(() => { const el = document.querySelector('` + sel + `');` +
		` el.value = '` + value + `';` +
		` el.dispatchEvent(new Event('input', {bubbles: true})); return true })()`
}

// TestUploadFileFlow verifies the "Upload File" modal uploads a file
// that then appears in the file table.
func TestUploadFileFlow(t *testing.T) {
	ctx := newTab(t)

	fileName := "e2e-upload-" + time.Now().Format("150405.000000") + ".txt"
	tmpFile := filepath.Join(t.TempDir(), fileName)
	if err := os.WriteFile(tmpFile, []byte("e2e upload content"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	err := chromedp.Run(ctx,
		chromedp.Navigate(adminURL("file-manager")),
		chromedp.WaitReady("table tbody tr:nth-of-type(2)", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("file manager did not load: %v", err)
	}

	// "Upload File" is the only .btn-secondary on the toolbar. The file
	// input is populated via CDP's DOM.setFileInputFiles.
	err = chromedp.Run(ctx,
		chromedp.Click(`button.btn.btn-secondary`, chromedp.ByQuery),
		chromedp.WaitReady(`.modal input[type="file"]`, chromedp.ByQuery),
		chromedp.SetUploadFiles(`.modal input[type="file"]`, []string{tmpFile}),
		chromedp.Click(`.modal-footer .btn-primary`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		t.Fatalf("upload flow failed: %v", err)
	}

	var body string
	if err := chromedp.Run(ctx, chromedp.Text("body", &body)); err != nil {
		t.Fatalf("reading file manager failed: %v", err)
	}
	if !strings.Contains(body, fileName) {
		t.Fatalf("expected uploaded file %q in file manager", fileName)
	}
}

// TestRenameFileFlow verifies renaming a file via its row's gear
// dropdown and the rename modal.
func TestRenameFileFlow(t *testing.T) {
	ctx := newTab(t)

	fileName := "e2e-rename-" + time.Now().Format("150405.000000") + ".txt"
	newName := "e2e-renamed-" + time.Now().Format("150405.000000") + ".txt"
	if err := testStorage.Put(rootDir+"/"+fileName, []byte("rename me")); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	err := chromedp.Run(ctx,
		chromedp.Navigate(adminURL("file-manager")),
		chromedp.WaitReady("table tbody tr:nth-of-type(2)", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("file manager did not load: %v", err)
	}

	var clicked bool
	err = chromedp.Run(ctx,
		chromedp.Evaluate(rowFindJS(fileName, `querySelector('.dropdown-toggle').click(); true`), &clicked),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Evaluate(dropdownItemJS(fileName, "Rename"), &clicked),
		chromedp.WaitReady(`.modal input[type="text"]`, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("rename modal did not open: %v", err)
	}

	var body, title string
	if err := chromedp.Run(ctx,
		chromedp.Text("body", &body),
		chromedp.Text(`.modal .modal-title`, &title, chromedp.ByQuery),
	); err != nil {
		t.Fatalf("reading rename modal failed: %v", err)
	}
	if !strings.Contains(title, "Rename File") {
		t.Fatalf("expected 'Rename File' modal title, got %q", title)
	}

	err = chromedp.Run(ctx,
		chromedp.Evaluate(setVueInputJS(`.modal input[type="text"]`, newName), &clicked),
		chromedp.Click(`.modal-footer button[type="submit"]`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		t.Fatalf("rename submit failed: %v", err)
	}

	if err := chromedp.Run(ctx, chromedp.Text("body", &body)); err != nil {
		t.Fatalf("reading file manager failed: %v", err)
	}
	if !strings.Contains(body, newName) {
		t.Fatalf("expected renamed file %q in file manager", newName)
	}
	if strings.Contains(body, fileName) {
		t.Fatalf("expected old name %q to be gone", fileName)
	}
}

// TestRenameDirectoryFlow verifies renaming a directory via its row's
// gear dropdown and the rename modal.
func TestRenameDirectoryFlow(t *testing.T) {
	ctx := newTab(t)

	dirName := "e2e-rename-dir-" + time.Now().Format("150405.000000")
	newName := "e2e-renamed-dir-" + time.Now().Format("150405.000000")
	if err := testStorage.MakeDirectory(rootDir + "/" + dirName); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	err := chromedp.Run(ctx,
		chromedp.Navigate(adminURL("file-manager")),
		chromedp.WaitReady("table tbody tr:nth-of-type(2)", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("file manager did not load: %v", err)
	}

	var clicked bool
	err = chromedp.Run(ctx,
		chromedp.Evaluate(rowFindJS(dirName, `querySelector('.dropdown-toggle').click(); true`), &clicked),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Evaluate(dropdownItemJS(dirName, "Rename"), &clicked),
		chromedp.WaitReady(`.modal input[type="text"]`, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("rename modal did not open: %v", err)
	}

	var title string
	if err := chromedp.Run(ctx,
		chromedp.Text(`.modal .modal-title`, &title, chromedp.ByQuery),
	); err != nil {
		t.Fatalf("reading rename modal failed: %v", err)
	}
	if !strings.Contains(title, "Rename Directory") {
		t.Fatalf("expected 'Rename Directory' modal title, got %q", title)
	}

	err = chromedp.Run(ctx,
		chromedp.Evaluate(setVueInputJS(`.modal input[type="text"]`, newName), &clicked),
		chromedp.Click(`.modal-footer button[type="submit"]`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		t.Fatalf("rename submit failed: %v", err)
	}

	var body string
	if err := chromedp.Run(ctx, chromedp.Text("body", &body)); err != nil {
		t.Fatalf("reading file manager failed: %v", err)
	}
	if !strings.Contains(body, newName) {
		t.Fatalf("expected renamed directory %q in file manager", newName)
	}
}
