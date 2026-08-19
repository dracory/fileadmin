package file_manager

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// --- Load Files ---

func TestHandleLoadFilesAjax_EmptyCurrentDirDefaultsToRoot(t *testing.T) {
	u := newTestUI(t)

	// Create a subdirectory so we can verify we're listing /uploads
	if err := u.Storage().MakeDirectory("/uploads/photos"); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	form := url.Values{}
	form.Add("current_dir", "")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.handleLoadFilesAjax(req)

	if !strings.Contains(result, "success") {
		t.Errorf("Expected success, got: %s", result)
	}
	if !strings.Contains(result, "photos") {
		t.Errorf("Expected 'photos' directory in listing (proving root defaults to /uploads), got: %s", result)
	}
}

func TestHandleLoadFilesAjax_PathTraversalRejected(t *testing.T) {
	u := newTestUI(t)

	form := url.Values{}
	form.Add("current_dir", "../../etc")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.handleLoadFilesAjax(req)

	if !strings.Contains(result, "invalid current directory") {
		t.Errorf("Expected path traversal rejection, got: %s", result)
	}
}

// --- Directory Create ---

func TestDirectoryCreateAjax_SuccessWithEmptyCurrentDir(t *testing.T) {
	u := newTestUI(t)

	form := url.Values{}
	form.Add("create_dir", "mydir")
	form.Add("current_dir", "")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.directoryCreateAjax(req)

	if !strings.Contains(result, "success") {
		t.Errorf("Expected success, got: %s", result)
	}

	// Verify directory was created at /uploads/mydir (not /mydir)
	dirs, err := u.Storage().Directories("/uploads")
	if err != nil {
		t.Fatalf("Failed to list directories: %v", err)
	}
	found := false
	for _, d := range dirs {
		if strings.HasSuffix(d, "/uploads/mydir") || d == "/uploads/mydir" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected /uploads/mydir to exist, got directories: %v", dirs)
	}
}

func TestDirectoryCreateAjax_PathTraversalRejected(t *testing.T) {
	u := newTestUI(t)

	form := url.Values{}
	form.Add("create_dir", "../../etc")
	form.Add("current_dir", "/uploads")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.directoryCreateAjax(req)

	if !strings.Contains(result, "invalid directory name") {
		t.Errorf("Expected path traversal rejection, got: %s", result)
	}
}

// --- Directory Delete ---

func TestDirectoryDeleteAjax_SuccessWithEmptyCurrentDir(t *testing.T) {
	u := newTestUI(t)

	// Create a directory to delete
	if err := u.Storage().MakeDirectory("/uploads/todelete"); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	form := url.Values{}
	form.Add("delete_dir", "todelete")
	form.Add("current_dir", "")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.directoryDeleteAjax(req)

	if !strings.Contains(result, "success") {
		t.Errorf("Expected success, got: %s", result)
	}

	// Verify directory is gone
	dirs, _ := u.Storage().Directories("/uploads")
	for _, d := range dirs {
		if strings.Contains(d, "todelete") {
			t.Errorf("Directory should have been deleted, but found: %s", d)
		}
	}
}

func TestDirectoryDeleteAjax_RootDirectoryRejected(t *testing.T) {
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

	// Empty delete_dir should be caught by the "required" check
	if !strings.Contains(result, "delete_dir is required") {
		t.Errorf("Expected 'delete_dir is required', got: %s", result)
	}
}

// --- File Delete ---

func TestFileDeleteAjax_SuccessWithEmptyCurrentDir(t *testing.T) {
	u := newTestUI(t)

	// Create a file to delete
	if err := u.Storage().Put("/uploads/todelete.txt", []byte("test")); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	form := url.Values{}
	form.Add("delete_file", "todelete.txt")
	form.Add("current_dir", "")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.fileDeleteAjax(req)

	if !strings.Contains(result, "success") {
		t.Errorf("Expected success, got: %s", result)
	}

	// Verify file is gone
	exists, _ := u.Storage().Exists("/uploads/todelete.txt")
	if exists {
		t.Errorf("File should have been deleted")
	}
}

func TestFileDeleteAjax_PathTraversalRejected(t *testing.T) {
	u := newTestUI(t)

	form := url.Values{}
	form.Add("delete_file", "../../etc/passwd")
	form.Add("current_dir", "/uploads")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.fileDeleteAjax(req)

	if !strings.Contains(result, "invalid file path") {
		t.Errorf("Expected path traversal rejection, got: %s", result)
	}
}

// --- File Clone ---

func TestFileCloneAjax_SuccessWithEmptyCurrentDir(t *testing.T) {
	u := newTestUI(t)

	// Create a file to clone
	if err := u.Storage().Put("/uploads/original.txt", []byte("test content")); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	form := url.Values{}
	form.Add("clone_file", "original.txt")
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

	// Verify clone was created at /uploads/original_copy.txt
	exists, _ := u.Storage().Exists("/uploads/original_copy.txt")
	if !exists {
		t.Errorf("Expected /uploads/original_copy.txt to exist after clone")
	}
}

func TestFileCloneAjax_PathTraversalRejected(t *testing.T) {
	u := newTestUI(t)

	form := url.Values{}
	form.Add("clone_file", "../../etc/passwd")
	form.Add("current_dir", "/uploads")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.fileCloneAjax(req)

	if !strings.Contains(result, "invalid file path") {
		t.Errorf("Expected path traversal rejection, got: %s", result)
	}
}

// --- File Rename ---

func TestFileRenameAjax_SuccessWithEmptyCurrentDir(t *testing.T) {
	u := newTestUI(t)

	// Create a file to rename
	if err := u.Storage().Put("/uploads/oldname.txt", []byte("test")); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	form := url.Values{}
	form.Add("rename_file", "oldname.txt")
	form.Add("new_file", "newname.txt")
	form.Add("current_dir", "")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.fileRenameAjax(req)

	if !strings.Contains(result, "success") {
		t.Errorf("Expected success, got: %s", result)
	}

	// Verify old file is gone and new file exists
	oldExists, _ := u.Storage().Exists("/uploads/oldname.txt")
	if oldExists {
		t.Errorf("Old file should have been renamed")
	}
	newExists, _ := u.Storage().Exists("/uploads/newname.txt")
	if !newExists {
		t.Errorf("New file should exist after rename")
	}
}

// --- File Upload ---

func TestFileUploadAjax_SuccessWithEmptyCurrentDir(t *testing.T) {
	u := newTestUI(t)

	// Build a multipart form with a file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("upload_file", "uploaded.txt")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	part.Write([]byte("uploaded content"))
	writer.WriteField("current_dir", "")
	writer.Close()

	req, err := http.NewRequest("POST", "/file-manager", body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	result := u.fileUploadAjax(req)

	if !strings.Contains(result, "success") {
		t.Errorf("Expected success, got: %s", result)
	}

	// Verify file was uploaded to /uploads/uploaded.txt
	exists, _ := u.Storage().Exists("/uploads/uploaded.txt")
	if !exists {
		t.Errorf("Expected /uploads/uploaded.txt to exist after upload")
	}
}

func TestFileUploadAjax_OversizedFileRejected(t *testing.T) {
	u := newTestUI(t)

	// Create a file that exceeds MAX_UPLOAD_SIZE (50MB)
	// We can't actually send 50MB, but we can set ContentLength > MAX_UPLOAD_SIZE
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("upload_file", "big.txt")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	part.Write([]byte("small"))
	writer.WriteField("current_dir", "/uploads")
	writer.Close()

	req, err := http.NewRequest("POST", "/file-manager", body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	// Spoof ContentLength to trigger the fast-path rejection
	req.ContentLength = MAX_UPLOAD_SIZE + 1

	result := u.fileUploadAjax(req)

	if !strings.Contains(result, "too big") {
		t.Errorf("Expected 'too big' rejection, got: %s", result)
	}
}

// --- Bulk Delete ---

func TestBulkDeleteAjax_SuccessWithValidPaths(t *testing.T) {
	u := newTestUI(t)

	// Create files to delete
	if err := u.Storage().Put("/uploads/file1.txt", []byte("1")); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	if err := u.Storage().Put("/uploads/file2.txt", []byte("2")); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	items := []map[string]string{
		{"path": "/uploads/file1.txt", "type": "file"},
		{"path": "/uploads/file2.txt", "type": "file"},
	}
	itemsJSON, _ := json.Marshal(items)

	form := url.Values{}
	form.Add("selected_items", string(itemsJSON))

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.bulkDeleteAjax(req)

	if !strings.Contains(result, "success") {
		t.Errorf("Expected success, got: %s", result)
	}

	// Verify both files are gone
	exists1, _ := u.Storage().Exists("/uploads/file1.txt")
	exists2, _ := u.Storage().Exists("/uploads/file2.txt")
	if exists1 || exists2 {
		t.Errorf("Both files should have been deleted, exists1=%v exists2=%v", exists1, exists2)
	}
}

func TestBulkDeleteAjax_PathTraversalRejected(t *testing.T) {
	u := newTestUI(t)

	items := []map[string]string{
		{"path": "../../etc/passwd", "type": "file"},
	}
	itemsJSON, _ := json.Marshal(items)

	form := url.Values{}
	form.Add("selected_items", string(itemsJSON))

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.bulkDeleteAjax(req)

	// Should not succeed — path traversal should be caught
	if strings.Contains(result, "\"status\":\"success\"") {
		t.Errorf("Path traversal should not succeed, got: %s", result)
	}
}

// --- Bulk Move ---

func TestBulkMoveAjax_SuccessWithValidPaths(t *testing.T) {
	u := newTestUI(t)

	// Create a file and a destination directory
	if err := u.Storage().Put("/uploads/moveme.txt", []byte("test")); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	if err := u.Storage().MakeDirectory("/uploads/dest"); err != nil {
		t.Fatalf("Failed to create dest directory: %v", err)
	}

	items := []map[string]string{
		{"path": "/uploads/moveme.txt", "type": "file"},
	}
	itemsJSON, _ := json.Marshal(items)

	form := url.Values{}
	form.Add("destination_dir", "/uploads/dest")
	form.Add("selected_items", string(itemsJSON))

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.bulkMoveAjax(req)

	if !strings.Contains(result, "success") {
		t.Errorf("Expected success, got: %s", result)
	}

	// Verify file moved
	oldExists, _ := u.Storage().Exists("/uploads/moveme.txt")
	newExists, _ := u.Storage().Exists("/uploads/dest/moveme.txt")
	if oldExists {
		t.Errorf("File should have been moved from /uploads/moveme.txt")
	}
	if !newExists {
		t.Errorf("File should have been moved to /uploads/dest/moveme.txt")
	}
}

func TestBulkMoveAjax_MoveIntoItselfRejected(t *testing.T) {
	u := newTestUI(t)

	// Create a directory with a subdirectory
	if err := u.Storage().MakeDirectory("/uploads/parent"); err != nil {
		t.Fatalf("Failed to create parent directory: %v", err)
	}

	items := []map[string]string{
		{"path": "/uploads/parent", "type": "directory"},
	}
	itemsJSON, _ := json.Marshal(items)

	form := url.Values{}
	form.Add("destination_dir", "/uploads/parent/sub")
	form.Add("selected_items", string(itemsJSON))

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.bulkMoveAjax(req)

	// Should report an error about moving into itself
	if !strings.Contains(result, "Cannot move directory into itself") {
		t.Errorf("Expected 'Cannot move directory into itself', got: %s", result)
	}
}

func TestBulkMoveAjax_PathTraversalInSourceRejected(t *testing.T) {
	u := newTestUI(t)

	items := []map[string]string{
		{"path": "../../etc/passwd", "type": "file"},
	}
	itemsJSON, _ := json.Marshal(items)

	form := url.Values{}
	form.Add("destination_dir", "/uploads")
	form.Add("selected_items", string(itemsJSON))

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.bulkMoveAjax(req)

	if strings.Contains(result, "\"status\":\"success\"") {
		t.Errorf("Path traversal in source should not succeed, got: %s", result)
	}
}

// --- Get Move Destinations ---

func TestGetMoveDestinationsAjax_ExcludesCurrentAndSelectedDirs(t *testing.T) {
	u := newTestUI(t)

	// Create directories
	if err := u.Storage().MakeDirectory("/uploads/dirA"); err != nil {
		t.Fatalf("Failed to create dirA: %v", err)
	}
	if err := u.Storage().MakeDirectory("/uploads/dirB"); err != nil {
		t.Fatalf("Failed to create dirB: %v", err)
	}

	items := []map[string]string{
		{"path": "/uploads/dirA", "type": "directory"},
	}
	itemsJSON, _ := json.Marshal(items)

	form := url.Values{}
	form.Add("current_dir", "/uploads")
	form.Add("selected_items", string(itemsJSON))

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.getMoveDestinationsAjax(req)

	if !strings.Contains(result, "success") {
		t.Errorf("Expected success, got: %s", result)
	}
	// dirA should be excluded (it's selected), dirB should be included
	if strings.Contains(result, "dirA") {
		t.Errorf("dirA should be excluded from move destinations (it's selected), got: %s", result)
	}
	if !strings.Contains(result, "dirB") {
		t.Errorf("dirB should be included in move destinations, got: %s", result)
	}
}

// --- Controller / Content-Type ---

func TestHandler_AjaxActionSetsJSONContentType(t *testing.T) {
	u := newTestUI(t)

	form := url.Values{}
	form.Add("action", "load-files")
	form.Add("current_dir", "")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	w := httptest.NewRecorder()
	_ = u.Handler(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Expected Content-Type 'application/json' for AJAX action, got %q", ct)
	}
}

func TestHandler_PageRenderDoesNotSetJSONContentType(t *testing.T) {
	u := newTestUI(t)

	// No action parameter → page render
	req, err := http.NewRequest("GET", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	_ = u.Handler(w, req)

	ct := w.Header().Get("Content-Type")
	if ct == "application/json" {
		t.Errorf("Page render should not set Content-Type to application/json, got %q", ct)
	}
}

func TestFileManager_PageRenderSetsHTMLContentType(t *testing.T) {
	u := newTestUI(t)

	req, err := http.NewRequest("GET", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	u.FileManager(w, req)

	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("Page render should set Content-Type to text/html, got %q", ct)
	}
}

func TestFileManager_AjaxDoesNotOverwriteJSONContentType(t *testing.T) {
	u := newTestUI(t)

	form := url.Values{}
	form.Add("action", "load-files")
	form.Add("current_dir", "")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	w := httptest.NewRecorder()
	u.FileManager(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("AJAX response should keep Content-Type 'application/json', got %q", ct)
	}
}
