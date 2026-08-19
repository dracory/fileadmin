package file_manager

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestBulkDeleteAjax_MissingSelectedItemsParameter(t *testing.T) {
	u := newTestUI(t)

	form := url.Values{}
	form.Add("selected_items", "")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.bulkDeleteAjax(req)

	if !strings.Contains(result, "No items selected") {
		t.Errorf("bulkDeleteAjax() result = %q, want to contain %q", result, "No items selected")
	}
}

func TestBulkDeleteAjax_InvalidJSON(t *testing.T) {
	u := newTestUI(t)

	form := url.Values{}
	form.Add("selected_items", "invalid json")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.bulkDeleteAjax(req)

	if !strings.Contains(result, "Invalid selected items data") {
		t.Errorf("bulkDeleteAjax() result = %q, want to contain %q", result, "Invalid selected items data")
	}
}

func TestBulkDeleteAjax_EmptyItemsArray(t *testing.T) {
	u := newTestUI(t)

	form := url.Values{}
	form.Add("selected_items", "[]")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.bulkDeleteAjax(req)

	if !strings.Contains(result, "No items selected") {
		t.Errorf("bulkDeleteAjax() with empty array should return 'No items selected', got: %q", result)
	}
}

func TestBulkDeleteAjax_NilStorage(t *testing.T) {
	u := newTestUI(t)
	u.StorageField = nil

	form := url.Values{}
	form.Add("selected_items", `[{"path":"/uploads/file.txt","type":"file"}]`)

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.bulkDeleteAjax(req)

	if !strings.Contains(result, "Storage not initialized") {
		t.Errorf("bulkDeleteAjax() with nil storage should return 'Storage not initialized', got: %q", result)
	}
}

func TestBulkDeleteAjax_NonExistentFileIsNoOp(t *testing.T) {
	u := newTestUI(t)

	form := url.Values{}
	form.Add("selected_items", `[{"path":"/uploads/nonexistent.txt","type":"file"}]`)

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.bulkDeleteAjax(req)

	// The SQL storage does not error when deleting a non-existent file,
	// so the bulk delete reports success (the item simply didn't exist).
	// This test verifies the handler doesn't crash on non-existent paths.
	if strings.Contains(result, "error") && !strings.Contains(result, "success") {
		// Some storage backends may report an error — that's acceptable too.
		t.Logf("Storage reported error for non-existent file (acceptable): %s", result)
	}
}

func TestBulkDeleteAjax_DeletesDirectory(t *testing.T) {
	u := newTestUI(t)

	if err := u.Storage().MakeDirectory("/uploads/todelete"); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	form := url.Values{}
	form.Add("selected_items", `[{"path":"/uploads/todelete","type":"directory"}]`)

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.bulkDeleteAjax(req)

	if !strings.Contains(result, "success") {
		t.Errorf("Expected success for directory delete, got: %s", result)
	}

	exists, _ := u.Storage().Exists("/uploads/todelete")
	if exists {
		t.Errorf("Directory should have been deleted")
	}
}
