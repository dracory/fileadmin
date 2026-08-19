package file_manager

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestFileUploadAjax_MissingUploadFile(t *testing.T) {
	u := newTestUI(t)

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	form := url.Values{}
	form.Add("current_dir", "/uploads")
	req.PostForm = form

	result := u.fileUploadAjax(req)

	if !strings.Contains(result, "multipart/form-data") {
		t.Errorf("fileUploadAjax() result = %q, want to contain %q", result, "multipart/form-data")
	}
}

func TestFileUploadAjax_InvalidMultipartBoundary(t *testing.T) {
	u := newTestUI(t)

	req, err := http.NewRequest("POST", "/file-manager", strings.NewReader("invalid body"))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "multipart/form-data")

	form := url.Values{}
	form.Add("current_dir", "/uploads")
	req.PostForm = form

	result := u.fileUploadAjax(req)

	if !strings.Contains(result, "boundary") {
		t.Errorf("fileUploadAjax() result = %q, want to contain %q", result, "boundary")
	}
}

func TestFileUploadAjaxWithFile(t *testing.T) {
	u := newTestUI(t)

	// Create a proper multipart request with a file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("upload_file", "test.txt")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	part.Write([]byte("test content"))
	writer.WriteField("current_dir", "/uploads")
	writer.Close()

	req, err := http.NewRequest("POST", "/file-manager", body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	result := u.fileUploadAjax(req)

	// The result should contain success or error, but not a boundary error
	if strings.Contains(result, "boundary") {
		t.Errorf("fileUploadAjax() should not have boundary error with proper multipart data, got: %q", result)
	}
}

func TestFileUploadAjax_NilStorage(t *testing.T) {
	u := newTestUI(t)
	u.StorageField = nil

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("upload_file", "test.txt")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	part.Write([]byte("test content"))
	writer.WriteField("current_dir", "/uploads")
	writer.Close()

	req, err := http.NewRequest("POST", "/file-manager", body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	result := u.fileUploadAjax(req)

	if !strings.Contains(result, "Storage not initialized") {
		t.Errorf("fileUploadAjax() with nil storage should return 'Storage not initialized', got: %q", result)
	}
}

func TestFileUploadAjax_PathTraversalInFilename(t *testing.T) {
	u := newTestUI(t)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("upload_file", "../../etc/passwd")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	part.Write([]byte("test content"))
	writer.WriteField("current_dir", "/uploads")
	writer.Close()

	req, err := http.NewRequest("POST", "/file-manager", body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	result := u.fileUploadAjax(req)

	// The multipart parser sanitizes the filename via filepath.Base(),
	// so "../../etc/passwd" becomes "passwd". The upload should succeed
	// with the sanitized name, proving path traversal is prevented.
	if strings.Contains(result, "invalid file path") {
		t.Errorf("fileUploadAjax() should not see path traversal (multipart sanitizes filename), got: %q", result)
	}
	// Verify the file was saved with the sanitized base name
	exists, _ := u.Storage().Exists("/uploads/passwd")
	if !exists {
		t.Errorf("Expected /uploads/passwd to exist after upload with sanitized filename")
	}
}
