package file_manager

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestFileDeleteAjax_MissingDeleteFileParameter(t *testing.T) {
	u := newTestUI(t)

	form := url.Values{}
	form.Add("delete_file", "")
	form.Add("current_dir", "/uploads")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.fileDeleteAjax(req)

	if !strings.Contains(result, "delete_file is required") {
		t.Errorf("fileDeleteAjax() result = %q, want to contain %q", result, "delete_file is required")
	}
}
