package file_manager

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestHandleLoadFilesAjax_NilStorage(t *testing.T) {
	u := newTestUI(t)
	u.StorageField = nil

	form := url.Values{}
	form.Add("current_dir", "/uploads")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.handleLoadFilesAjax(req)

	if !strings.Contains(result, "storage is required") {
		t.Errorf("handleLoadFilesAjax() result = %q, want to contain %q", result, "storage is required")
	}
}

func TestHandleLoadFilesAjax_LoadsFromRoot(t *testing.T) {
	u := newTestUI(t)

	form := url.Values{}
	form.Add("current_dir", "")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.handleLoadFilesAjax(req)

	if !strings.Contains(result, "Files loaded successfully") {
		t.Errorf("handleLoadFilesAjax() result = %q, want to contain %q", result, "Files loaded successfully")
	}
}

func TestHandleLoadFilesAjax_LoadsWithFilesAndDirectories(t *testing.T) {
	u := newTestUI(t)

	// Create a subdirectory and a file
	if err := u.Storage().MakeDirectory("/uploads/photos"); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := u.Storage().Put("/uploads/readme.txt", []byte("hello")); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	form := url.Values{}
	form.Add("current_dir", "")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.handleLoadFilesAjax(req)

	if !strings.Contains(result, "Files loaded successfully") {
		t.Errorf("Expected success, got: %s", result)
	}
	if !strings.Contains(result, "photos") {
		t.Errorf("Expected 'photos' directory in listing, got: %s", result)
	}
	if !strings.Contains(result, "readme.txt") {
		t.Errorf("Expected 'readme.txt' file in listing, got: %s", result)
	}
}

func TestHandleLoadFilesAjax_SubdirectoryWithParent(t *testing.T) {
	u := newTestUI(t)

	if err := u.Storage().MakeDirectory("/uploads/docs"); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := u.Storage().Put("/uploads/docs/note.txt", []byte("note")); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	form := url.Values{}
	form.Add("current_dir", "/uploads/docs")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.handleLoadFilesAjax(req)

	if !strings.Contains(result, "Files loaded successfully") {
		t.Errorf("Expected success, got: %s", result)
	}
	// parent_directory should be set when browsing a subdirectory
	if !strings.Contains(result, "parent_directory") {
		t.Errorf("Expected 'parent_directory' in result for subdirectory, got: %s", result)
	}
	if !strings.Contains(result, "note.txt") {
		t.Errorf("Expected 'note.txt' in listing, got: %s", result)
	}
}
