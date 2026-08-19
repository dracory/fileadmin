package shared

import "net/http"

// FileAdminURL returns the file admin base URL from request context
func FileAdminURL(r *http.Request) string {
	value, ok := r.Context().Value(KeyFileAdminURL).(string)
	if !ok {
		return ""
	}
	return value
}
