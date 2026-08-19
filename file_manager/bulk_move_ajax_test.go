package file_manager

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestBulkMoveAjax_MissingSelectedItemsParameter(t *testing.T) {
	u := newTestUI(t)

	form := url.Values{}
	form.Add("destination_dir", "/uploads")
	form.Add("selected_items", "")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.bulkMoveAjax(req)

	if !strings.Contains(result, "No items selected") {
		t.Errorf("bulkMoveAjax() result = %q, want to contain %q", result, "No items selected")
	}
}

func TestBulkMoveAjax_InvalidJSON(t *testing.T) {
	u := newTestUI(t)

	form := url.Values{}
	form.Add("destination_dir", "/uploads")
	form.Add("selected_items", "invalid json")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.bulkMoveAjax(req)

	if !strings.Contains(result, "Invalid selected items data") {
		t.Errorf("bulkMoveAjax() result = %q, want to contain %q", result, "Invalid selected items data")
	}
}

func TestBulkMoveAjax_NilRequest(t *testing.T) {
	u := newTestUI(t)

	result := u.bulkMoveAjax(nil)

	if !strings.Contains(result, "invalid request") {
		t.Errorf("bulkMoveAjax(nil) should return 'invalid request', got: %q", result)
	}
}

func TestBulkMoveAjax_EmptyItemsArray(t *testing.T) {
	u := newTestUI(t)

	form := url.Values{}
	form.Add("destination_dir", "/uploads")
	form.Add("selected_items", "[]")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.bulkMoveAjax(req)

	if !strings.Contains(result, "No items selected") {
		t.Errorf("bulkMoveAjax() with empty array should return 'No items selected', got: %q", result)
	}
}

func TestBulkMoveAjax_NilStorage(t *testing.T) {
	u := newTestUI(t)
	u.StorageField = nil

	form := url.Values{}
	form.Add("destination_dir", "/uploads")
	form.Add("selected_items", `[{"path":"/uploads/file.txt","type":"file"}]`)

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.bulkMoveAjax(req)

	if !strings.Contains(result, "Storage not initialized") {
		t.Errorf("bulkMoveAjax() with nil storage should return 'Storage not initialized', got: %q", result)
	}
}

func TestBulkMoveAjax_InvalidDestinationDir(t *testing.T) {
	u := newTestUI(t)

	form := url.Values{}
	form.Add("destination_dir", "../../etc")
	form.Add("selected_items", `[{"path":"/uploads/file.txt","type":"file"}]`)

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.bulkMoveAjax(req)

	if !strings.Contains(result, "invalid destination directory") {
		t.Errorf("bulkMoveAjax() with path traversal destination should return 'invalid destination directory', got: %q", result)
	}
}

func TestBulkMoveAjax_MoveToRoot(t *testing.T) {
	u := newTestUI(t)

	if err := u.Storage().MakeDirectory("/uploads/sub"); err != nil {
		t.Fatalf("Failed to create sub directory: %v", err)
	}
	if err := u.Storage().Put("/uploads/sub/moveme.txt", []byte("test")); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	form := url.Values{}
	form.Add("destination_dir", "")
	form.Add("selected_items", `[{"path":"/sub/moveme.txt","type":"file"}]`)

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.bulkMoveAjax(req)

	if !strings.Contains(result, "success") {
		t.Errorf("Expected success for move to root, got: %s", result)
	}

	// File should now be at /uploads/moveme.txt (root directory)
	newExists, _ := u.Storage().Exists("/uploads/moveme.txt")
	if !newExists {
		t.Errorf("Expected /uploads/moveme.txt to exist after move to root")
	}
}
