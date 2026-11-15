// Package e2e contains end-to-end tests that drive the fileadmin panel
// with a real headless Chrome via chromedp (CDP). The admin is served
// in-process by httptest on a random port — no binary build needed.
package e2e

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/dracory/fileadmin"
	"github.com/dracory/fileadmin/testutils"
	"github.com/dracory/filesystem"
	_ "modernc.org/sqlite"
)

var (
	baseURL     string
	allocCtx    context.Context
	testStorage filesystem.StorageInterface
	rootDir     = "/uploads"
	seedDir     = "documents"
	seedFile    = "welcome.txt"
)

const adminPath = "/admin/file-manager"

func TestMain(m *testing.M) {
	// In-process admin server backed by a seeded in-memory SQL storage.
	storage, _, err := testutils.InitStorage(":memory:")
	if err != nil {
		panic("e2e: failed to init storage: " + err.Error())
	}
	testStorage = storage

	seedStorage(storage)

	panel, err := fileadmin.New(fileadmin.AdminOptions{
		Storage:      storage,
		RootDirPath:  rootDir,
		AdminHomeURL: "/admin",
		FileAdminURL: adminPath,
	})
	if err != nil {
		panic("e2e: failed to create fileadmin: " + err.Error())
	}

	mux := http.NewServeMux()
	mux.Handle(adminPath, http.HandlerFunc(panel.Handle))
	mux.Handle(adminPath+"/", http.HandlerFunc(panel.Handle))
	srv := httptest.NewServer(mux)
	baseURL = srv.URL

	// One shared browser process for the entire test run.
	opts := append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Headless)
	if bin := chromeBin(); bin != "" {
		opts = append(opts, chromedp.ExecPath(bin))
	}
	var cancel context.CancelFunc
	allocCtx, cancel = chromedp.NewExecAllocator(context.Background(), opts...)

	code := m.Run()

	cancel()
	srv.Close()
	os.Exit(code)
}

// seedStorage populates the storage with directories and files so the
// file manager page has rows to render.
func seedStorage(storage filesystem.StorageInterface) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// The SQL driver requires parents to exist — create the root first.
	if err := storage.MakeDirectory(rootDir); err != nil {
		logger.Error("seedStorage: failed to create root directory", "dir", rootDir, "error", err)
	}

	for _, dir := range []string{"documents", "images", "projects"} {
		if err := storage.MakeDirectory(rootDir + "/" + dir); err != nil {
			logger.Error("seedStorage: failed to create directory", "dir", dir, "error", err)
		}
	}

	files := map[string]string{
		rootDir + "/documents/readme.txt": "Sample readme file for e2e tests.",
		rootDir + "/documents/notes.txt": "Some quick notes for e2e tests.",
		rootDir + "/images/logo.txt":     "placeholder for logo.png",
		rootDir + "/welcome.txt":         "Welcome to fileadmin! Root-level seed file.",
	}
	for path, content := range files {
		if err := storage.Put(path, []byte(content)); err != nil {
			logger.Error("seedStorage: failed to create file", "path", path, "error", err)
		}
	}
}

// newTab opens a fresh tab in the shared browser and returns its
// context. Each tab gets a 30s deadline; cleanup closes the tab.
func newTab(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := chromedp.NewContext(allocCtx)
	ctx, cancelTimeout := context.WithTimeout(ctx, 30*time.Second)
	t.Cleanup(func() { cancelTimeout(); cancel() })
	return ctx
}

// chromeBin resolves a headless Chromium binary. Order: CHROME_BIN env,
// Playwright's downloaded headless shell, empty (chromedp finds Chrome
// on PATH / standard install locations).
func chromeBin() string {
	if b := os.Getenv("CHROME_BIN"); b != "" {
		return b
	}
	patterns := []string{
		// Windows
		filepath.Join(os.Getenv("LOCALAPPDATA"),
			"ms-playwright/chromium_headless_shell-*/chrome-headless-shell-win64/chrome-headless-shell.exe"),
		// WSL: Windows-side Playwright shell via interop
		"/mnt/*/Users/*/AppData/Local/ms-playwright/chromium_headless_shell-*/chrome-headless-shell-win64/chrome-headless-shell.exe",
		// Linux
		filepath.Join(os.Getenv("HOME"),
			".cache/ms-playwright/chromium_headless_shell-*/chrome-headless-shell-linux64/chrome-headless-shell"),
	}
	for _, pattern := range patterns {
		if matches, _ := filepath.Glob(pattern); len(matches) > 0 {
			return matches[len(matches)-1]
		}
	}
	return ""
}
