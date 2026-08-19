package file_manager

import (
	"embed"
	"net/http"
	"strings"

	"github.com/dracory/api"
	"github.com/dracory/hb"
	"github.com/dracory/req"

	"github.com/dracory/fileadmin/shared"
)

//go:embed *.html
//go:embed *.js
var filesEmbed embed.FS

const (
	actionLoadFiles                   = "load-files"
	JSON_ACTION_FILE_CLONE            = "file_clone"
	JSON_ACTION_FILE_RENAME           = "file_rename"
	JSON_ACTION_FILE_DELETE           = "file_delete"
	JSON_ACTION_FILE_UPLOAD           = "file_upload"
	JSON_ACTION_DIRECTORY_CREATE      = "directory_create"
	JSON_ACTION_DIRECTORY_DELETE      = "directory_delete"
	JSON_ACTION_BULK_MOVE             = "bulk_move"
	JSON_ACTION_BULK_DELETE           = "bulk_delete"
	JSON_ACTION_GET_MOVE_DESTINATIONS = "get_move_destinations"
	MAX_UPLOAD_SIZE                   = 50 * 1024 * 1024 // 50MB
)

// UiInterface defines the file manager controller's UI interface
type UiInterface interface {
	shared.UiInterface
	FileManager(w http.ResponseWriter, r *http.Request)
}

// ui implements UiInterface
type ui struct {
	shared.UiBase
}

// UI creates a new file manager controller UI from the given config
func UI(config shared.UiConfig) UiInterface {
	return &ui{UiBase: shared.NewUiBase(config)}
}

// FileManager handles the file manager controller requests
func (u *ui) FileManager(w http.ResponseWriter, r *http.Request) {
	html := u.Handler(w, r)
	if html != "" {
		// Don't overwrite Content-Type if Handler already set it (e.g.
		// application/json for AJAX responses). Only default to text/html
		// for page renders.
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		}
		_, _ = w.Write([]byte(html))
	}
}

// Handler processes the file manager controller request and returns the
// response body as a string. For AJAX actions it also sets the
// Content-Type header to application/json on the response writer.
func (u *ui) Handler(w http.ResponseWriter, r *http.Request) string {
	return u.anyIndex(w, r)
}

// anyIndex routes to the appropriate action handler
func (u *ui) anyIndex(w http.ResponseWriter, r *http.Request) string {
	action := strings.TrimSpace(req.GetStringTrimmed(r, "action"))

	// AJAX actions return JSON
	if isAjaxAction(action) {
		w.Header().Set("Content-Type", "application/json")
	}

	switch action {
	case actionLoadFiles:
		return u.handleLoadFilesAjax(r)
	case JSON_ACTION_FILE_CLONE:
		return u.fileCloneAjax(r)
	case JSON_ACTION_FILE_RENAME:
		return u.fileRenameAjax(r)
	case JSON_ACTION_FILE_DELETE:
		return u.fileDeleteAjax(r)
	case JSON_ACTION_DIRECTORY_CREATE:
		return u.directoryCreateAjax(r)
	case JSON_ACTION_DIRECTORY_DELETE:
		return u.directoryDeleteAjax(r)
	case JSON_ACTION_FILE_UPLOAD:
		return u.fileUploadAjax(r)
	case JSON_ACTION_BULK_MOVE:
		return u.bulkMoveAjax(r)
	case JSON_ACTION_BULK_DELETE:
		return u.bulkDeleteAjax(r)
	case JSON_ACTION_GET_MOVE_DESTINATIONS:
		return u.getMoveDestinationsAjax(r)
	default:
		return u.renderPage(r)
	}
}

// isAjaxAction returns true if the given action returns a JSON response
func isAjaxAction(action string) bool {
	switch action {
	case actionLoadFiles,
		JSON_ACTION_FILE_CLONE,
		JSON_ACTION_FILE_RENAME,
		JSON_ACTION_FILE_DELETE,
		JSON_ACTION_FILE_UPLOAD,
		JSON_ACTION_DIRECTORY_CREATE,
		JSON_ACTION_DIRECTORY_DELETE,
		JSON_ACTION_BULK_MOVE,
		JSON_ACTION_BULK_DELETE,
		JSON_ACTION_GET_MOVE_DESTINATIONS:
		return true
	}
	return false
}

// renderPage renders the file manager Vue.js application
func (u *ui) renderPage(r *http.Request) string {
	if u.Storage() == nil {
		return api.Error("storage is required").ToString()
	}

	htmlContent, err := filesEmbed.ReadFile("files.html")
	if err != nil {
		return api.Error("Failed to read files HTML template: " + err.Error()).ToString()
	}

	jsContent, err := filesEmbed.ReadFile("files.js")
	if err != nil {
		return api.Error("Failed to read files JavaScript file: " + err.Error()).ToString()
	}

	linksHelper := shared.NewLinksFromRequest(r)
	initScript := hb.Script(`
		const urlFileManager = '` + linksHelper.FileManager(nil) + `';
	`)

	htmlTemplate := hb.Wrap().HTML(string(htmlContent))
	componentScript := hb.Script(string(jsContent))

	// Note: Vue CDN is loaded by shared.Layout via cdn.VueJs_3(), so we
	// don't add it here to avoid loading it twice.
	vueContainer := hb.Div().
		Child(htmlTemplate).
		Child(initScript).
		Child(componentScript)

	content := hb.Div().
		Class("container").
		Child(vueContainer)

	return u.Layout(nil, r, "File Manager", content.ToHTML(), struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}{})
}
