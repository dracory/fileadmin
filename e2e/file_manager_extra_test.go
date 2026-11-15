package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

// TestDeleteDirectoryFlow verifies deleting a directory via its row's
// gear dropdown and the confirm modal.
func TestDeleteDirectoryFlow(t *testing.T) {
	ctx := newTab(t)

	dirName := "e2e-delete-dir-" + time.Now().Format("150405.000000")
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
		chromedp.Evaluate(dropdownItemJS(dirName, "Delete"), &clicked),
		chromedp.WaitReady(`.modal .modal-title`, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("delete modal did not open: %v", err)
	}

	var title string
	if err := chromedp.Run(ctx,
		chromedp.Text(`.modal .modal-title`, &title, chromedp.ByQuery),
	); err != nil {
		t.Fatalf("reading delete modal failed: %v", err)
	}
	if !strings.Contains(title, "Confirm Directory Delete") {
		t.Fatalf("expected 'Confirm Directory Delete' modal title, got %q", title)
	}

	err = chromedp.Run(ctx,
		chromedp.Click(`.modal-footer .btn-danger`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		t.Fatalf("delete confirmation failed: %v", err)
	}

	var body string
	if err := chromedp.Run(ctx, chromedp.Text("body", &body)); err != nil {
		t.Fatalf("reading file manager failed: %v", err)
	}
	if strings.Contains(body, dirName) {
		t.Fatalf("expected directory %q to be deleted", dirName)
	}
}

// TestCloneFileFlow verifies cloning a file via its row's gear
// dropdown and the clone modal.
func TestCloneFileFlow(t *testing.T) {
	ctx := newTab(t)

	fileName := "e2e-clone-" + time.Now().Format("150405.000000") + ".txt"
	cloneName := "e2e-cloned-" + time.Now().Format("150405.000000") + ".txt"
	if err := testStorage.Put(rootDir+"/"+fileName, []byte("clone me")); err != nil {
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
		chromedp.Evaluate(dropdownItemJS(fileName, "Clone"), &clicked),
		chromedp.WaitReady(`.modal .modal-title`, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("clone modal did not open: %v", err)
	}

	var title string
	if err := chromedp.Run(ctx,
		chromedp.Text(`.modal .modal-title`, &title, chromedp.ByQuery),
	); err != nil {
		t.Fatalf("reading clone modal failed: %v", err)
	}
	if !strings.Contains(title, "Clone File") {
		t.Fatalf("expected 'Clone File' modal title, got %q", title)
	}

	err = chromedp.Run(ctx,
		chromedp.Evaluate(setVueInputJS(`.modal input[type="text"]`, cloneName), &clicked),
		chromedp.Click(`.modal-footer button[type="submit"]`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		t.Fatalf("clone submit failed: %v", err)
	}

	var body string
	if err := chromedp.Run(ctx, chromedp.Text("body", &body)); err != nil {
		t.Fatalf("reading file manager failed: %v", err)
	}
	if !strings.Contains(body, cloneName) {
		t.Fatalf("expected cloned file %q in file manager", cloneName)
	}
}

// TestViewFileFlow verifies the file view modal opens via the row's
// gear dropdown.
func TestViewFileFlow(t *testing.T) {
	ctx := newTab(t)

	fileName := "e2e-view-" + time.Now().Format("150405.000000") + ".txt"
	if err := testStorage.Put(rootDir+"/"+fileName, []byte("view me")); err != nil {
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
		chromedp.Evaluate(dropdownItemJS(fileName, "View"), &clicked),
		chromedp.WaitReady(`.modal .modal-title`, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("view modal did not open: %v", err)
	}

	var title string
	if err := chromedp.Run(ctx,
		chromedp.Text(`.modal .modal-title`, &title, chromedp.ByQuery),
	); err != nil {
		t.Fatalf("reading view modal failed: %v", err)
	}
	if !strings.Contains(title, "File View") {
		t.Fatalf("expected 'File View' modal title, got %q", title)
	}
}

// TestUploadMultipleFlow verifies the "Upload Multiple" modal uploads
// several files that then appear in the file table.
func TestUploadMultipleFlow(t *testing.T) {
	ctx := newTab(t)

	stamp := time.Now().Format("150405.000000")
	file1 := "e2e-multi-1-" + stamp + ".txt"
	file2 := "e2e-multi-2-" + stamp + ".txt"
	tmpDir := t.TempDir()
	tmp1 := filepath.Join(tmpDir, file1)
	tmp2 := filepath.Join(tmpDir, file2)
	for _, p := range []string{tmp1, tmp2} {
		if err := os.WriteFile(p, []byte("e2e multi upload"), 0644); err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
	}

	err := chromedp.Run(ctx,
		chromedp.Navigate(adminURL("file-manager")),
		chromedp.WaitReady("table tbody tr:nth-of-type(2)", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("file manager did not load: %v", err)
	}

	// "Upload Multiple" is the toolbar .btn-primary without .btn-sm.
	// The hidden multi file input accepts several paths.
	err = chromedp.Run(ctx,
		chromedp.Click(`button.btn.btn-primary`, chromedp.ByQuery),
		chromedp.WaitReady(`.modal input[type="file"][multiple]`, chromedp.ByQuery),
		chromedp.SetUploadFiles(`.modal input[type="file"][multiple]`, []string{tmp1, tmp2}),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Click(`.modal-footer .btn-primary`, chromedp.ByQuery),
		chromedp.Sleep(3*time.Second),
	)
	if err != nil {
		t.Fatalf("multi upload flow failed: %v", err)
	}

	var body string
	if err := chromedp.Run(ctx, chromedp.Text("body", &body)); err != nil {
		t.Fatalf("reading file manager failed: %v", err)
	}
	for _, name := range []string{file1, file2} {
		if !strings.Contains(body, name) {
			t.Fatalf("expected uploaded file %q in file manager", name)
		}
	}
}

// clickBulkToggleJS returns JS that clicks the "Bulk Actions" dropdown
// toggle. The menu lives outside the table rows, so it is located by
// button text across all dropdown toggles.
func clickBulkToggleJS() string {
	return `(() => {` +
		` const btn = [...document.querySelectorAll('.dropdown-toggle')]` +
		`.find(b => b.textContent.includes('Bulk Actions'));` +
		` if (!btn) return false; btn.click(); return true })()`
}

// clickDropdownItemJS returns JS that clicks the .dropdown-item whose
// text contains label (used for the open Bulk Actions menu).
func clickDropdownItemJS(label string) string {
	return `(() => {` +
		` const item = [...document.querySelectorAll('.dropdown-item')]` +
		`.find(i => i.textContent.includes('` + label + `'));` +
		` if (!item) return false; item.click(); return true })()`
}

// TestBulkDeleteSelectedFlow verifies selecting rows and deleting them
// via the Bulk Actions dropdown and the confirm modal.
func TestBulkDeleteSelectedFlow(t *testing.T) {
	ctx := newTab(t)

	stamp := time.Now().Format("150405.000000")
	file1 := "e2e-bulk-1-" + stamp + ".txt"
	file2 := "e2e-bulk-2-" + stamp + ".txt"
	for _, name := range []string{file1, file2} {
		if err := testStorage.Put(rootDir+"/"+name, []byte("bulk delete")); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	err := chromedp.Run(ctx,
		chromedp.Navigate(adminURL("file-manager")),
		chromedp.WaitReady("table tbody tr:nth-of-type(2)", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("file manager did not load: %v", err)
	}

	// Tick both rows' checkboxes; the Bulk Actions dropdown appears
	// once selectedItems is non-empty.
	var clicked bool
	err = chromedp.Run(ctx,
		chromedp.Evaluate(clickRowCheckboxJS(file1), &clicked),
		chromedp.Evaluate(clickRowCheckboxJS(file2), &clicked),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Evaluate(clickBulkToggleJS(), &clicked),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Evaluate(clickDropdownItemJS("Delete Selected"), &clicked),
		chromedp.WaitReady(`.modal .modal-title`, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("bulk delete modal did not open: %v", err)
	}

	var title string
	if err := chromedp.Run(ctx,
		chromedp.Text(`.modal .modal-title`, &title, chromedp.ByQuery),
	); err != nil {
		t.Fatalf("reading bulk delete modal failed: %v", err)
	}
	if !strings.Contains(title, "Delete Selected Items") {
		t.Fatalf("expected 'Delete Selected Items' modal title, got %q", title)
	}

	err = chromedp.Run(ctx,
		chromedp.Click(`.modal-footer .btn-danger`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		t.Fatalf("bulk delete confirmation failed: %v", err)
	}

	var body string
	if err := chromedp.Run(ctx, chromedp.Text("body", &body)); err != nil {
		t.Fatalf("reading file manager failed: %v", err)
	}
	for _, name := range []string{file1, file2} {
		if strings.Contains(body, name) {
			t.Fatalf("expected file %q to be deleted", name)
		}
	}
}

// TestMoveSelectedFlow verifies selecting a file and moving it into a
// directory via the Bulk Actions dropdown and the move modal.
func TestMoveSelectedFlow(t *testing.T) {
	ctx := newTab(t)

	stamp := time.Now().Format("150405.000000")
	dirName := "e2e-move-dir-" + stamp
	fileName := "e2e-move-" + stamp + ".txt"
	if err := testStorage.MakeDirectory(rootDir + "/" + dirName); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}
	if err := testStorage.Put(rootDir+"/"+fileName, []byte("move me")); err != nil {
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
		chromedp.Evaluate(clickRowCheckboxJS(fileName), &clicked),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Evaluate(clickBulkToggleJS(), &clicked),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Evaluate(clickDropdownItemJS("Move Selected"), &clicked),
		chromedp.WaitReady(`.modal .modal-title`, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("move modal did not open: %v", err)
	}

	var title string
	if err := chromedp.Run(ctx,
		chromedp.Text(`.modal .modal-title`, &title, chromedp.ByQuery),
	); err != nil {
		t.Fatalf("reading move modal failed: %v", err)
	}
	if !strings.Contains(title, "Move Selected Items") {
		t.Fatalf("expected 'Move Selected Items' modal title, got %q", title)
	}

	// Destination options are dir paths root-relative ("/<dir>").
	err = chromedp.Run(ctx,
		chromedp.Evaluate(setVueSelectJS(`.modal select.form-select`, "/"+dirName), &clicked),
		chromedp.Click(`.modal-footer button[type="submit"]`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		t.Fatalf("move submit failed: %v", err)
	}

	var body string
	if err := chromedp.Run(ctx, chromedp.Text("body", &body)); err != nil {
		t.Fatalf("reading file manager failed: %v", err)
	}
	if strings.Contains(body, fileName) {
		t.Fatalf("expected file %q moved out of root", fileName)
	}

	// The file should now live inside the destination directory.
	exists, err := testStorage.Exists(rootDir + "/" + dirName + "/" + fileName)
	if err != nil || !exists {
		t.Fatalf("expected file at %s/%s (exists=%v, err=%v)", dirName, fileName, exists, err)
	}
}

// TestSelectAllFlow verifies the header checkbox selects every row and
// the bulk actions bar reflects the count.
func TestSelectAllFlow(t *testing.T) {
	ctx := newTab(t)

	err := chromedp.Run(ctx,
		chromedp.Navigate(adminURL("file-manager")),
		chromedp.WaitReady("table tbody tr:nth-of-type(2)", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("file manager did not load: %v", err)
	}

	err = chromedp.Run(ctx,
		chromedp.Click(`table thead input[type="checkbox"]`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("select all failed: %v", err)
	}

	var body string
	if err := chromedp.Run(ctx, chromedp.Text("body", &body)); err != nil {
		t.Fatalf("reading file manager failed: %v", err)
	}
	if !strings.Contains(body, "Bulk Actions") {
		t.Fatal("expected 'Bulk Actions' bar after selecting all rows")
	}
}
