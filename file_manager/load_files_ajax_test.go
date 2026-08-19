package file_manager

import (
	"encoding/json"
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

func TestHandleLoadFilesAjax_ReturnsRootDirectory(t *testing.T) {
	u := newTestUI(t)

	form := url.Values{}
	form.Add("current_dir", "")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.handleLoadFilesAjax(req)

	if !strings.Contains(result, "root_directory") {
		t.Errorf("Expected 'root_directory' in response, got: %s", result)
	}
	if !strings.Contains(result, "/uploads") {
		t.Errorf("Expected root_directory to contain '/uploads', got: %s", result)
	}
}

func TestHandleLoadFilesAjax_TopLevelDirHasEmptyParent(t *testing.T) {
	u := newTestUI(t)

	if err := u.Storage().MakeDirectory("/uploads/myfolder"); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	form := url.Values{}
	form.Add("current_dir", "/uploads/myfolder")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.handleLoadFilesAjax(req)

	// Parse the JSON to check parent_directory is empty (root-relative)
	var resp struct {
		Data struct {
			ParentDirectory string `json:"parent_directory"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v\nResponse: %s", err, result)
	}

	// Parent is root, which is represented as "" (empty string).
	// The frontend shows the parent link when currentDirectory != "",
	// and navigateTo("") goes to root.
	if resp.Data.ParentDirectory != "" {
		t.Errorf("Expected empty parent_directory for top-level folder (parent is root), got: %q", resp.Data.ParentDirectory)
	}
}

func TestHandleLoadFilesAjax_DeepDirHasNonEmptyParent(t *testing.T) {
	u := newTestUI(t)

	if err := u.Storage().MakeDirectory("/uploads/parent"); err != nil {
		t.Fatalf("Failed to create parent directory: %v", err)
	}
	if err := u.Storage().MakeDirectory("/uploads/parent/child"); err != nil {
		t.Fatalf("Failed to create child directory: %v", err)
	}

	form := url.Values{}
	form.Add("current_dir", "/uploads/parent/child")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.handleLoadFilesAjax(req)

	// Parse the JSON to check parent_directory is set
	var resp struct {
		Data struct {
			ParentDirectory  string `json:"parent_directory"`
			CurrentDirectory string `json:"current_directory"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v\nResponse: %s", err, result)
	}

	if resp.Data.ParentDirectory != "/parent" {
		t.Errorf("Expected parent_directory '/parent' (with leading slash), got: %q", resp.Data.ParentDirectory)
	}
	// current_directory should be root-relative (no "/uploads" prefix)
	if resp.Data.CurrentDirectory != "/parent/child" {
		t.Errorf("Expected current_directory '/parent/child' (root-relative), got: %q", resp.Data.CurrentDirectory)
	}
}

func TestHandleLoadFilesAjax_ReturnsRootRelativePaths(t *testing.T) {
	u := newTestUI(t)

	if err := u.Storage().MakeDirectory("/uploads/myfolder"); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := u.Storage().Put("/uploads/myfolder/myfile.txt", []byte("test")); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	form := url.Values{}
	form.Add("current_dir", "/myfolder")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	result := u.handleLoadFilesAjax(req)

	var resp struct {
		Data struct {
			CurrentDirectory string      `json:"current_directory"`
			Directories      []FileEntry `json:"directories"`
			Files            []FileEntry `json:"files"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v\nResponse: %s", err, result)
	}

	// current_directory should be root-relative
	if resp.Data.CurrentDirectory != "/myfolder" {
		t.Errorf("Expected current_directory '/myfolder', got: %q", resp.Data.CurrentDirectory)
	}

	// Directory paths should be root-relative (no "/uploads" prefix)
	if len(resp.Data.Directories) != 0 {
		// myfolder itself won't appear, but if there were subdirs they'd be root-relative
	}

	// File paths should be root-relative
	if len(resp.Data.Files) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(resp.Data.Files))
	}
	if resp.Data.Files[0].Path != "/myfolder/myfile.txt" {
		t.Errorf("Expected file path '/myfolder/myfile.txt' (root-relative), got: %q", resp.Data.Files[0].Path)
	}
	if resp.Data.Files[0].Name != "myfile.txt" {
		t.Errorf("Expected file name 'myfile.txt', got: %q", resp.Data.Files[0].Name)
	}
}
