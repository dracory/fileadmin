// Package fileadmin provides a standalone file admin interface following
// the folder-per-controller pattern. Each controller is in its own
// subfolder and handles its own views and AJAX data.
//
// This module is modeled on github.com/dracory/shopadmin,
// github.com/dracory/blogadmin, and github.com/dracory/logadmin.
package fileadmin

import (
	"context"
	"net/http"
	"strings"

	"github.com/dracory/fileadmin/file_manager"
	"github.com/dracory/fileadmin/shared"
	"github.com/dracory/filesystem"
	"github.com/dracory/req"
)

// AdminOptions contains all dependencies and configuration for the file admin.
//
// Storage and RootDirPath replace the in-repo version's Registry field
// (which was app.AppInterface). This matches the shopadmin/blogadmin/logadmin
// convention where dependencies are passed directly.
//
// FuncLayout is an optional function to render the admin interface
// inside your own layout (branding, menus, etc.). It receives the
// request and response writer so the host project can access request
// context (auth user, locale, etc.) when rendering the layout.
// If nil, a default bare-bones HTML page is used (Bootstrap + Vue CDN).
// Uses anonymous struct to match shopadmin/blogadmin/logadmin exactly, so
// consumers can reuse their existing layout function for fileadmin.
type AdminOptions struct {
	// Storage is the filesystem.StorageInterface (required)
	Storage filesystem.StorageInterface

	// RootDirPath is the root directory for file operations (required).
	// Typically derived from the host project's media root config.
	// e.g. "/uploads"
	RootDirPath string

	// FuncLayout is an optional function to render the admin interface
	// inside your own layout (branding, menus, etc.). It receives the
	// request and response writer so the host project can access request
	// context (auth user, locale, etc.) when rendering the layout.
	FuncLayout func(w http.ResponseWriter, r *http.Request, title string, body string, options struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}) string

	// AdminHomeURL is the URL for the admin home page (default: "/admin")
	AdminHomeURL string

	// FileAdminURL is the base URL for file admin (default: "/admin/file-manager")
	FileAdminURL string

	// AuthUserID returns the authenticated user ID from the request.
	// If it returns "", the user is treated as unauthenticated.
	AuthUserID func(r *http.Request) string
}

// AdminInterface defines the interface for the file admin
type AdminInterface interface {
	Handle(w http.ResponseWriter, r *http.Request)
}

// admin implements AdminInterface
type admin struct {
	storage     filesystem.StorageInterface
	rootDirPath string
	funcLayout  func(w http.ResponseWriter, r *http.Request, title string, body string, options struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}) string
	adminHomeURL string
	fileAdminURL string
	authUserID   func(r *http.Request) string
	routes       map[string]func(w http.ResponseWriter, r *http.Request)
}

// New creates a new file admin instance.
// Returns ErrStorageRequired if Storage is nil.
func New(opts AdminOptions) (AdminInterface, error) {
	if opts.Storage == nil {
		return nil, ErrStorageRequired
	}

	// Set defaults
	if opts.AdminHomeURL == "" {
		opts.AdminHomeURL = "/admin"
	}
	if opts.FileAdminURL == "" {
		opts.FileAdminURL = "/admin/file-manager"
	}

	// Normalize root dir path
	rootDirPath := strings.TrimSpace(opts.RootDirPath)
	rootDirPath = strings.Trim(rootDirPath, "/")
	rootDirPath = strings.Trim(rootDirPath, ".")
	rootDirPath = "/" + rootDirPath

	a := &admin{
		storage:      opts.Storage,
		rootDirPath:  rootDirPath,
		funcLayout:   opts.FuncLayout,
		adminHomeURL: opts.AdminHomeURL,
		fileAdminURL: opts.FileAdminURL,
		authUserID:   opts.AuthUserID,
	}

	// Build routes once at construction time
	a.routes = a.buildRoutes()

	return a, nil
}

// Handle processes all file admin requests.
// Config values are injected into the request context (following the
// shopadmin/blogadmin/logadmin pattern). Route lookup is map-based.
func (a *admin) Handle(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	if a.authUserID != nil && a.authUserID(r) == "" {
		http.Redirect(w, r, a.adminHomeURL, http.StatusSeeOther)
		return
	}

	// Inject config into request context (like shopadmin/blogadmin/logadmin)
	ctx := context.WithValue(r.Context(), shared.KeyEndpoint, r.URL.Path)
	ctx = context.WithValue(ctx, shared.KeyAdminHomeURL, a.adminHomeURL)
	ctx = context.WithValue(ctx, shared.KeyFileAdminURL, a.fileAdminURL)
	r = r.WithContext(ctx)

	// Map-based route lookup
	controller := req.GetStringTrimmed(r, "controller")
	if controller == "" {
		controller = shared.CONTROLLER_FILE_MANAGER
	}

	handler, ok := a.routes[controller]
	if !ok {
		handler = a.routes[shared.CONTROLLER_FILE_MANAGER]
	}

	handler(w, r)
}

// buildRoutes creates the handler dispatch map once at construction time.
func (a *admin) buildRoutes() map[string]func(w http.ResponseWriter, r *http.Request) {
	uiConfig := shared.UiConfig{
		Storage:     a.storage,
		RootDirPath: a.rootDirPath,
		Layout:      a.render,
	}

	return map[string]func(w http.ResponseWriter, r *http.Request){
		shared.CONTROLLER_FILE_MANAGER: func(w http.ResponseWriter, r *http.Request) {
			file_manager.UI(uiConfig).FileManager(w, r)
		},
	}
}

// render wraps content in the layout. If FuncLayout is provided and
// returns non-empty, it is used; otherwise the default shared.Layout
// is used (following the shopadmin/blogadmin/logadmin pattern).
//
// When FuncLayout is set, the default shared.Layout is NOT computed
// (avoids wasted work).
func (a *admin) render(w http.ResponseWriter, r *http.Request, webpageTitle, webpageHtml string, options struct {
	Styles     []string
	StyleURLs  []string
	Scripts    []string
	ScriptURLs []string
}) string {
	// If a custom layout is provided, try it first
	if a.funcLayout != nil {
		custom := a.funcLayout(w, r, webpageTitle, webpageHtml, options)
		if custom != "" {
			if w != nil {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				_, _ = w.Write([]byte(custom))
				return ""
			}
			return custom
		}
	}

	webpage := shared.Layout(w, r, webpageTitle, webpageHtml, options)

	if w != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(webpage))
		return ""
	}

	return webpage
}
