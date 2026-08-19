package file_manager

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestDirectoryDeleteAjax_MissingDeleteDirParameter(t *testing.T) {
	u := newTestUI(t)

	form := url.Values{}
	form.Add("delete_dir", "")
	form.Add("current_dir", "/uploads")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.directoryDeleteAjax(req)

	if !strings.Contains(result, "delete_dir is required") {
		t.Errorf("directoryDeleteAjax() result = %q, want to contain %q", result, "delete_dir is required")
	}
}

func TestDirectoryDeleteAjax_InvalidCurrentDir(t *testing.T) {
	u := newTestUI(t)

	form := url.Values{}
	form.Add("delete_dir", "test")
	form.Add("current_dir", ".")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.directoryDeleteAjax(req)

	if !strings.Contains(result, "invalid directory name") {
		t.Errorf("directoryDeleteAjax() result = %q, want to contain %q", result, "invalid directory name")
	}
}
