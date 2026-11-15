package e2e

import (
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
	rowJS := func(inner string) string {
		return `[...document.querySelectorAll('table tbody tr')]` +
			`.find(tr => tr.textContent.includes('` + fileName + `'))` +
			`?.` + inner
	}
	var clicked bool
	err = chromedp.Run(ctx,
		chromedp.Evaluate(rowJS(`querySelector('.dropdown-toggle').click(); true`), &clicked),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Evaluate(rowJS(`querySelector('.dropdown-item.text-danger').click(); true`), &clicked),
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
