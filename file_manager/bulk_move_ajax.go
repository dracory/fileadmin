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

// bulkMoveAjax handles bulk file/folder move requests
func (u *ui) bulkMoveAjax(r *http.Request) string {
	if r == nil || r.URL == nil {
		return api.Error("invalid request").ToString()
	}

	destinationDir := req.GetStringTrimmed(r, "destination_dir")

	// Resolve root-relative destination to full storage path, then
	// normalize to prevent path traversal.
	resolvedDest := u.resolveCurrentDir(destinationDir)
	normalizedDestDir, err := verifyAndNormalizeDirPath("", strings.Trim(resolvedDest, "/"))
	if err != nil {
		return api.Error("invalid destination directory: " + err.Error()).ToString()
	}
	destinationDir = normalizedDestDir

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

		// Resolve root-relative path (from frontend) to full storage
		// path, then validate to prevent path traversal attacks.
		fullSrcPath := u.resolveCurrentDir(item.Path)
		normalizedSrcPath, err := verifyAndNormalizePathOrError("", strings.TrimPrefix(fullSrcPath, "/"))
		if err != nil {
			errors = append(errors, "Invalid source path: "+err.Error())
			continue
		}

		// Extract the filename/directory name from the normalized path
		itemName := filepath.Base(normalizedSrcPath)

		// Build the new path
		var newPath string
		if destinationDir == "" || destinationDir == "/" {
			newPath = "/" + itemName
		} else {
			newPath = strings.TrimRight(destinationDir, "/") + "/" + itemName
		}

		// Check if trying to move a directory into itself or one of its
		// subdirectories. Use path-boundary-aware prefix check to avoid
		// false positives (e.g. /foo vs /foobar).
		if item.Type == "directory" && (destinationDir == normalizedSrcPath || strings.HasPrefix(destinationDir, normalizedSrcPath+"/")) {
			errors = append(errors, "Cannot move directory into itself: "+itemName)
			continue
		}

		// Perform the move using storage.Move
		err = u.Storage().Move(normalizedSrcPath, newPath)
		if err != nil {
			log.Printf("Error moving %s to %s: %v", normalizedSrcPath, newPath, err)
			errors = append(errors, "Failed to move "+itemName+": "+err.Error())
			continue
		}

		successCount++
	}

	// Return appropriate response
	if successCount == 0 {
		return api.Error("Failed to move items: " + strings.Join(errors, "; ")).ToString()
	}

	message := "Successfully moved " + strconv.Itoa(successCount) + " item(s)"
	if len(errors) > 0 {
		message += ". Some items failed: " + strings.Join(errors, "; ")
	}

	return api.Success(message).ToString()
}
