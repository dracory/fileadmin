package shared

import (
	"net/http"

	"github.com/dracory/filesystem"
)

// UiBase is a base struct that implements shared.UiInterface.
// Subcontroller ui structs can embed this to get the Storage(),
// RootDirPath(), and Layout() methods for free, following the
// shopadmin/blogadmin/logadmin pattern.
type UiBase struct {
	StorageField     filesystem.StorageInterface
	RootDirPathField string
	LayoutField      func(w http.ResponseWriter, r *http.Request, webpageTitle, webpageHtml string, options struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}) string
}

func (u UiBase) Storage() filesystem.StorageInterface { return u.StorageField }
func (u UiBase) RootDirPath() string                  { return u.RootDirPathField }

func (u UiBase) Layout(w http.ResponseWriter, r *http.Request, webpageTitle, webpageHtml string, options struct {
	Styles     []string
	StyleURLs  []string
	Scripts    []string
	ScriptURLs []string
}) string {
	return u.LayoutField(w, r, webpageTitle, webpageHtml, options)
}

// NewUiBase creates a UiBase from a UiConfig
func NewUiBase(config UiConfig) UiBase {
	return UiBase{
		StorageField:     config.Storage,
		RootDirPathField: config.RootDirPath,
		LayoutField:      config.Layout,
	}
}
