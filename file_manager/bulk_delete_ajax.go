package file_manager

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dracory/api"
	"github.com/dracory/req"
)

// bulkDeleteAjax handles bulk file/folder delete requests
func (u *ui) bulkDeleteAjax(r *http.Request) string {
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
		log.Printf("Error parsing selected items JSON: %v", err)
		return api.Error("Invalid selected items data").ToString()
	}

	if len(selectedItems) == 0 {
		return api.Error("No items selected").ToString()
	}

	if u.Storage() == nil {
		return api.Error("Storage not initialized").ToString()
	}

	// Track success and failures
	successCount := 0
	var errors []string

	for _, item := range selectedItems {
		if item.Path == "" {
			continue
		}

		// Validate path to prevent path traversal attacks.
		// item.Path comes from user-submitted JSON and must be normalized
		// before being passed to storage, like all other handlers.
		normalizedPath, err := verifyAndNormalizePathOrError("", strings.TrimPrefix(item.Path, "/"))
		if err != nil {
			errors = append(errors, "Invalid path: "+err.Error())
			continue
		}

		itemName := filepath.Base(normalizedPath)

		if item.Type == "directory" {
			err = u.Storage().DeleteDirectory(normalizedPath)
		} else {
			err = u.Storage().DeleteFile([]string{normalizedPath})
		}

		if err != nil {
			log.Printf("Error deleting %s: %v", normalizedPath, err)
			errors = append(errors, "Failed to delete "+itemName+": "+err.Error())
			continue
		}

		successCount++
	}

	// Return appropriate response
	if successCount == 0 {
		return api.Error("Failed to delete items: " + strings.Join(errors, "; ")).ToString()
	}

	message := "Successfully deleted " + strconv.Itoa(successCount) + " item(s)"
	if len(errors) > 0 {
		message += ". Some items failed: " + strings.Join(errors, "; ")
	}

	return api.Success(message).ToString()
}
