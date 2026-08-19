package shared

import "net/http"

// Links provides URL helpers for fileadmin controllers.
// The base URL is read from request context (injected by Handle()),
// not hardcoded. This follows the shopadmin/blogadmin/logadmin pattern.
type Links struct {
	baseURL string
}

// NewLinks creates a Links helper with the given base URL.
// If baseURL is empty, defaults to "/admin/file-manager".
func NewLinks(baseURL string) *Links {
	if baseURL == "" {
		baseURL = "/admin/file-manager"
	}
	return &Links{baseURL: baseURL}
}

// NewLinksFromRequest creates a Links helper using the file admin URL
// from the request context.
func NewLinksFromRequest(r *http.Request) *Links {
	return NewLinks(FileAdminURL(r))
}

// FileManager builds the URL for the file manager controller
func (l *Links) FileManager(params map[string]string) string {
	return l.url(CONTROLLER_FILE_MANAGER, params)
}

// url builds a URL for the given controller. The params map is copied
// before mutation (does not modify caller's map).
func (l *Links) url(controller string, params map[string]string) string {
	return URL(l.baseURL, controller, params)
}
