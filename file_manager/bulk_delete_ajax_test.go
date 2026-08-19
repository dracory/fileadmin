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
