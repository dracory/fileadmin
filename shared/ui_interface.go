package shared

import (
	"net/http"

	"github.com/dracory/filesystem"
)

// UiInterface defines the methods every subcontroller UI must implement.
// This follows the shopadmin/blogadmin/logadmin pattern.
type UiInterface interface {
	Storage() filesystem.StorageInterface
	RootDirPath() string

	Layout(w http.ResponseWriter, r *http.Request, webpageTitle, webpageHtml string, options struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}) string
}
