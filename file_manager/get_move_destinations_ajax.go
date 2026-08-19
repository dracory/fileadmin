package file_manager

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dracory/api"
	"github.com/dracory/req"
)

// getMoveDestinationsAjax returns a filtered list of directories valid for moving selected items
func (u *ui) getMoveDestinationsAjax(r *http.Request) string {
	rawCurrentDir := req.GetStringTrimmed(r, "current_dir")
	fullCurrentDir := u.resolveCurrentDir(rawCurrentDir)

	// Parse selected items JSON
	selectedItemsJSON := req.GetStringTrimmed(r, "selected_items")
	if selectedItemsJSON == "" {
		return api.Error("No items selected").ToString()
	}

	var selectedItems []struct {
		Path string `json:"path"`
		Type string `json:"type"`
	}

	if err := json.Unmarshal([]byte(selectedItemsJSON), &selectedItems); err != nil {
		return api.Error("Invalid selected items data").ToString()
	}

	if len(selectedItems) == 0 {
		return api.Error("No items selected").ToString()
	}

	if u.Storage() == nil {
		return api.Error("Storage not initialized").ToString()
	}

	allDirs, err := u.allDirectories(u.RootDirPath())
	if err != nil {
		return api.Error(err.Error()).ToString()
	}

	// Build set of full storage paths to exclude:
	// - current directory itself
	// - any selected directory (resolve root-relative to full)
	// - any subdirectory of a selected directory
	excludePaths := map[string]bool{}
	for _, item := range selectedItems {
		if item.Path == "" {
			continue
		}
		fullItemPath := u.resolveCurrentDir(item.Path)
		excludePaths[strings.TrimRight(fullItemPath, "/")] = true
	}

	filtered := []FileEntry{}
	for _, dir := range allDirs {
		dirPath := strings.TrimRight(dir.Path, "/")

		// Exclude current directory
		if dirPath == strings.TrimRight(fullCurrentDir, "/") {
			continue
		}

		// Exclude selected directories and their subdirectories
		excluded := false
		for excludedPath := range excludePaths {
			if dirPath == excludedPath || strings.HasPrefix(dirPath+"/", excludedPath+"/") {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		// Return root-relative paths to the frontend
		dir.Path = u.stripRootPrefix(dir.Path)
		filtered = append(filtered, dir)
	}

	return api.SuccessWithData("", map[string]interface{}{
		"directories": filtered,
	}).ToString()
}
