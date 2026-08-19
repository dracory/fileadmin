package file_manager

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestFileCloneAjax_MissingCloneFileParameter(t *testing.T) {
	u := newTestUI(t)

	form := url.Values{}
	form.Add("clone_file", "")
	form.Add("current_dir", "/uploads")
	form.Add("new_file", "")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.fileCloneAjax(req)

	if !strings.Contains(result, "clone_file is required") {
		t.Errorf("fileCloneAjax() result = %q, want to contain %q", result, "clone_file is required")
	}
}

func TestFileCloneAjax_NilStorage(t *testing.T) {
	u := newTestUI(t)
	u.StorageField = nil

	form := url.Values{}
	form.Add("clone_file", "test.txt")
	form.Add("current_dir", "/uploads")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.fileCloneAjax(req)

	if !strings.Contains(result, "Storage not initialized") {
		t.Errorf("fileCloneAjax() with nil storage should return 'Storage not initialized', got: %q", result)
	}
}

func TestFileCloneAjax_WithExplicitNewFileName(t *testing.T) {
	u := newTestUI(t)

	if err := u.Storage().Put("/uploads/original.txt", []byte("content")); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	form := url.Values{}
	form.Add("clone_file", "original.txt")
	form.Add("new_file", "custom_name.txt")
	form.Add("current_dir", "")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.fileCloneAjax(req)

	if !strings.Contains(result, "success") {
		t.Errorf("Expected success, got: %s", result)
	}

	exists, _ := u.Storage().Exists("/uploads/custom_name.txt")
	if !exists {
		t.Errorf("Expected /uploads/custom_name.txt to exist after clone with explicit name")
	}
}

func TestFileCloneAjax_CopyNameCollisionAppendsCounter(t *testing.T) {
	u := newTestUI(t)

	// Create original file and its first copy so the counter loop runs
	if err := u.Storage().Put("/uploads/doc.txt", []byte("content")); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := u.Storage().Put("/uploads/doc_copy.txt", []byte("content")); err != nil {
		t.Fatalf("Failed to create first copy: %v", err)
	}

	form := url.Values{}
	form.Add("clone_file", "doc.txt")
	form.Add("new_file", "")
	form.Add("current_dir", "")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.fileCloneAjax(req)

	if !strings.Contains(result, "success") {
		t.Errorf("Expected success, got: %s", result)
	}

	// Should create doc_copy_copy_2.txt since doc_copy.txt already exists.
	// The counter loop uses base="doc_copy" (from newFileName="doc_copy.txt"),
	// producing "doc_copy_copy_2.txt".
	exists, _ := u.Storage().Exists("/uploads/doc_copy_copy_2.txt")
	if !exists {
		t.Errorf("Expected /uploads/doc_copy_copy_2.txt to exist after collision-handling clone")
	}
}

func TestFileCloneAjax_InvalidNewFilePath(t *testing.T) {
	u := newTestUI(t)

	form := url.Values{}
	form.Add("clone_file", "test.txt")
	form.Add("new_file", "../../etc/passwd")
	form.Add("current_dir", "/uploads")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.fileCloneAjax(req)

	if !strings.Contains(result, "invalid file path") {
		t.Errorf("Expected 'invalid file path' for traversal in new_file, got: %s", result)
	}
}
