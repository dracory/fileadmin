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
