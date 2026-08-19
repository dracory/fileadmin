package file_manager

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/dracory/fileadmin/shared"
	"github.com/dracory/fileadmin/testutils"
	_ "modernc.org/sqlite"
)

func newTestUI(t *testing.T) *ui {
	t.Helper()
	storage, db, err := testutils.InitStorage(":memory:")
	if err != nil {
		t.Fatalf("Failed to init storage: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	// Create root directory so listing works
	if err := storage.MakeDirectory("/uploads"); err != nil {
		t.Fatalf("Failed to create root directory: %v", err)
	}

	return &ui{UiBase: shared.NewUiBase(shared.UiConfig{
		Storage:     storage,
		RootDirPath: "/uploads",
		Layout: func(w http.ResponseWriter, r *http.Request, title, body string, opts struct {
			Styles     []string
			StyleURLs  []string
			Scripts    []string
			ScriptURLs []string
		}) string {
			return body
		},
	})}
}

func TestFileManagerController_LoadFilesAction(t *testing.T) {
	u := newTestUI(t)

	form := url.Values{}
	form.Add("action", "load-files")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	w := httptest.NewRecorder()
	_ = u.Handler(w, req)
}

func TestFileManagerController_FileCloneAction(t *testing.T) {
	u := newTestUI(t)

	form := url.Values{}
	form.Add("action", "file_clone")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	w := httptest.NewRecorder()
	result := u.Handler(w, req)

	if !strings.Contains(result, "clone_file is required") {
		t.Errorf("Handler() result = %q, want to contain %q", result, "clone_file is required")
	}
}

func TestFileManagerController_FileRenameAction(t *testing.T) {
	u := newTestUI(t)

	form := url.Values{}
	form.Add("action", "file_rename")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	w := httptest.NewRecorder()
	result := u.Handler(w, req)

	if !strings.Contains(result, "rename_file is required") {
		t.Errorf("Handler() result = %q, want to contain %q", result, "rename_file is required")
	}
}

func TestFileManagerController_FileDeleteAction(t *testing.T) {
	u := newTestUI(t)

	form := url.Values{}
	form.Add("action", "file_delete")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	w := httptest.NewRecorder()
	result := u.Handler(w, req)

	if !strings.Contains(result, "delete_file is required") {
		t.Errorf("Handler() result = %q, want to contain %q", result, "delete_file is required")
	}
}

func TestFileManagerController_DirectoryCreateAction(t *testing.T) {
	u := newTestUI(t)

	form := url.Values{}
	form.Add("action", "directory_create")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	w := httptest.NewRecorder()
	result := u.Handler(w, req)

	if !strings.Contains(result, "create_dir is required") {
		t.Errorf("Handler() result = %q, want to contain %q", result, "create_dir is required")
	}
}

func TestFileManagerController_DirectoryDeleteAction(t *testing.T) {
	u := newTestUI(t)

	form := url.Values{}
	form.Add("action", "directory_delete")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	w := httptest.NewRecorder()
	result := u.Handler(w, req)

	if !strings.Contains(result, "delete_dir is required") {
		t.Errorf("Handler() result = %q, want to contain %q", result, "delete_dir is required")
	}
}

func TestFileManagerController_BulkMoveAction(t *testing.T) {
	u := newTestUI(t)

	form := url.Values{}
	form.Add("action", "bulk_move")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	w := httptest.NewRecorder()
	result := u.Handler(w, req)

	if !strings.Contains(result, "No items selected") {
		t.Errorf("Handler() result = %q, want to contain %q", result, "No items selected")
	}
}

func TestFileManagerController_BulkDeleteAction(t *testing.T) {
	u := newTestUI(t)

	form := url.Values{}
	form.Add("action", "bulk_delete")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	w := httptest.NewRecorder()
	result := u.Handler(w, req)

	if !strings.Contains(result, "No items selected") {
		t.Errorf("Handler() result = %q, want to contain %q", result, "No items selected")
	}
}

func TestFileManagerController_GetMoveDestinationsAction(t *testing.T) {
	u := newTestUI(t)

	form := url.Values{}
	form.Add("action", "get_move_destinations")

	req, err := http.NewRequest("POST", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.PostForm = form

	w := httptest.NewRecorder()
	result := u.Handler(w, req)

	if !strings.Contains(result, "No items selected") {
		t.Errorf("Handler() result = %q, want to contain %q", result, "No items selected")
	}
}

func TestUI_FactoryReturnsUiInterface(t *testing.T) {
	storage, db, err := testutils.InitStorage(":memory:")
	if err != nil {
		t.Fatalf("Failed to init storage: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := storage.MakeDirectory("/uploads"); err != nil {
		t.Fatalf("Failed to create root directory: %v", err)
	}

	config := shared.UiConfig{
		Storage:     storage,
		RootDirPath: "/uploads",
		Layout: func(w http.ResponseWriter, r *http.Request, title, body string, opts struct {
			Styles     []string
			StyleURLs  []string
			Scripts    []string
			ScriptURLs []string
		}) string {
			return body
		},
	}

	u := UI(config)
	if u == nil {
		t.Fatalf("UI() returned nil")
	}

	// Verify it implements UiInterface by calling FileManager
	req := httptest.NewRequest(http.MethodGet, "/file-manager", nil)
	w := httptest.NewRecorder()
	u.FileManager(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRenderPage_NilStorageReturnsError(t *testing.T) {
	u := newTestUI(t)
	u.StorageField = nil

	req, err := http.NewRequest("GET", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	result := u.renderPage(req)

	if !strings.Contains(result, "storage is required") {
		t.Errorf("renderPage() with nil storage should return 'storage is required', got: %q", result)
	}
}

func TestAnyIndex_DefaultActionRendersPage(t *testing.T) {
	u := newTestUI(t)

	req, err := http.NewRequest("GET", "/file-manager", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	result := u.anyIndex(w, req)

	if result == "" {
		t.Errorf("Expected non-empty page render for default action")
	}
	if !strings.Contains(result, "container") {
		t.Errorf("Expected rendered page to contain 'container', got: %q", result)
	}
}
