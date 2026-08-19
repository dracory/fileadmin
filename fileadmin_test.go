package fileadmin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dracory/fileadmin/testutils"
	_ "modernc.org/sqlite"
)

func TestNew_ValidOptions(t *testing.T) {
	storage, db, err := testutils.InitStorage(":memory:")
	if err != nil {
		t.Fatalf("Failed to init storage: %v", err)
	}
	defer db.Close()

	options := AdminOptions{
		Storage:      storage,
		RootDirPath:  "/uploads",
		AdminHomeURL: "/admin",
		FileAdminURL: "/admin/file-manager",
	}
	a, err := New(options)
	if err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}
	if a == nil {
		t.Errorf("Expected admin to be created, got nil")
	}
}

func TestNew_MissingStorage(t *testing.T) {
	options := AdminOptions{
		RootDirPath: "/uploads",
	}
	a, err := New(options)
	if err == nil {
		t.Errorf("Expected error when storage is missing")
	}
	if a != nil {
		t.Errorf("Expected nil admin when storage is missing")
	}
	if !strings.Contains(err.Error(), ErrStorageRequired.Error()) {
		t.Errorf("Expected error to contain '%s', got '%s'", ErrStorageRequired.Error(), err.Error())
	}
}

func TestNew_Defaults(t *testing.T) {
	storage, db, err := testutils.InitStorage(":memory:")
	if err != nil {
		t.Fatalf("Failed to init storage: %v", err)
	}
	defer db.Close()

	options := AdminOptions{
		Storage:     storage,
		RootDirPath: "/uploads",
	}
	a, err := New(options)
	if err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	// Verify defaults were set
	admin := a.(*admin)
	if admin.adminHomeURL != "/admin" {
		t.Errorf("Expected default adminHomeURL '/admin', got '%s'", admin.adminHomeURL)
	}
	if admin.fileAdminURL != "/admin/file-manager" {
		t.Errorf("Expected default fileAdminURL '/admin/file-manager', got '%s'", admin.fileAdminURL)
	}
}

func TestNew_RootDirPathNormalization(t *testing.T) {
	storage, db, err := testutils.InitStorage(":memory:")
	if err != nil {
		t.Fatalf("Failed to init storage: %v", err)
	}
	defer db.Close()

	tests := []struct {
		input    string
		expected string
	}{
		{"/uploads", "/uploads"},
		{"uploads", "/uploads"},
		{"/uploads/", "/uploads"},
		{".", "/"},
		{"", "/"},
	}

	for _, tt := range tests {
		a, err := New(AdminOptions{
			Storage:     storage,
			RootDirPath: tt.input,
		})
		if err != nil {
			t.Fatalf("Failed to create admin for input %q: %v", tt.input, err)
		}
		admin := a.(*admin)
		if admin.rootDirPath != tt.expected {
			t.Errorf("RootDirPath(%q) = %q, want %q", tt.input, admin.rootDirPath, tt.expected)
		}
	}
}

func TestHandle_FileManagerController(t *testing.T) {
	storage, db, err := testutils.InitStorage(":memory:")
	if err != nil {
		t.Fatalf("Failed to init storage: %v", err)
	}
	defer db.Close()

	// Create root directory so listing works
	if err := storage.MakeDirectory("/uploads"); err != nil {
		t.Fatalf("Failed to create root directory: %v", err)
	}

	a, err := New(AdminOptions{
		Storage:     storage,
		RootDirPath: "/uploads",
	})
	if err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/file-manager", nil)
	rr := httptest.NewRecorder()

	a.Handle(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "File Manager") {
		t.Errorf("Expected body to contain 'File Manager'")
	}
}

func TestHandle_UnknownControllerFallsBackToFileManager(t *testing.T) {
	storage, db, err := testutils.InitStorage(":memory:")
	if err != nil {
		t.Fatalf("Failed to init storage: %v", err)
	}
	defer db.Close()

	if err := storage.MakeDirectory("/uploads"); err != nil {
		t.Fatalf("Failed to create root directory: %v", err)
	}

	a, err := New(AdminOptions{
		Storage:     storage,
		RootDirPath: "/uploads",
	})
	if err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/file-manager?controller=nonexistent", nil)
	rr := httptest.NewRecorder()

	a.Handle(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "File Manager") {
		t.Errorf("Expected fallback to file manager page with 'File Manager'")
	}
}

func TestHandle_AuthUserIDRedirect(t *testing.T) {
	storage, db, err := testutils.InitStorage(":memory:")
	if err != nil {
		t.Fatalf("Failed to init storage: %v", err)
	}
	defer db.Close()

	a, err := New(AdminOptions{
		Storage:     storage,
		RootDirPath: "/uploads",
		AuthUserID:  func(r *http.Request) string { return "" },
	})
	if err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/file-manager", nil)
	rr := httptest.NewRecorder()

	a.Handle(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Expected redirect status 303, got %d", rr.Code)
	}
}

func TestHandle_CustomLayoutUsedWhenProvided(t *testing.T) {
	storage, db, err := testutils.InitStorage(":memory:")
	if err != nil {
		t.Fatalf("Failed to init storage: %v", err)
	}
	defer db.Close()

	if err := storage.MakeDirectory("/uploads"); err != nil {
		t.Fatalf("Failed to create root directory: %v", err)
	}

	layoutCalled := false
	a, err := New(AdminOptions{
		Storage:     storage,
		RootDirPath: "/uploads",
		FuncLayout: func(w http.ResponseWriter, r *http.Request, title, body string, options struct {
			Styles     []string
			StyleURLs  []string
			Scripts    []string
			ScriptURLs []string
		}) string {
			layoutCalled = true
			return "<html><head><title>" + title + "</title></head><body>" + body + "</body></html>"
		},
	})
	if err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/file-manager", nil)
	rr := httptest.NewRecorder()

	a.Handle(rr, req)

	if !layoutCalled {
		t.Errorf("Expected custom layout to be called")
	}

	body := rr.Body.String()
	if !strings.Contains(body, "<html>") {
		t.Errorf("Expected custom layout HTML wrapper, got: %s", body)
	}
}

func TestHandle_CustomLayoutEmptyFallsBackToDefault(t *testing.T) {
	storage, db, err := testutils.InitStorage(":memory:")
	if err != nil {
		t.Fatalf("Failed to init storage: %v", err)
	}
	defer db.Close()

	if err := storage.MakeDirectory("/uploads"); err != nil {
		t.Fatalf("Failed to create root directory: %v", err)
	}

	a, err := New(AdminOptions{
		Storage:     storage,
		RootDirPath: "/uploads",
		FuncLayout: func(w http.ResponseWriter, r *http.Request, title, body string, options struct {
			Styles     []string
			StyleURLs  []string
			Scripts    []string
			ScriptURLs []string
		}) string {
			return "" // Return empty to trigger fallback
		},
	})
	if err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/file-manager", nil)
	rr := httptest.NewRecorder()

	a.Handle(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	body := rr.Body.String()
	// Default layout should contain the page title
	if !strings.Contains(body, "File Manager") {
		t.Errorf("Expected default layout fallback to contain 'File Manager', got: %s", body)
	}
}

func TestRender_WithWriterWritesToResponse(t *testing.T) {
	storage, db, err := testutils.InitStorage(":memory:")
	if err != nil {
		t.Fatalf("Failed to init storage: %v", err)
	}
	defer db.Close()

	a, err := New(AdminOptions{
		Storage:     storage,
		RootDirPath: "/uploads",
	})
	if err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	admin := a.(*admin)
	req := httptest.NewRequest(http.MethodGet, "/admin/file-manager", nil)
	w := httptest.NewRecorder()

	result := admin.render(w, req, "Test Title", "<p>body</p>", struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}{})

	// When w is non-nil, render writes to w and returns ""
	if result != "" {
		t.Errorf("Expected empty string return when writer is provided, got: %q", result)
	}

	body := w.Body.String()
	if !strings.Contains(body, "<p>body</p>") {
		t.Errorf("Expected body to be written to response writer, got: %s", body)
	}
}
