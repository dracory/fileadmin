package file_manager

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/dracory/api"
	"github.com/dracory/req"
)

// fileCloneAjax handles file clone (duplicate) requests
func (u *ui) fileCloneAjax(r *http.Request) string {
	selectedFileName := req.GetStringTrimmed(r, "clone_file")
	if selectedFileName == "" {
		return api.Error("clone_file is required").ToString()
	}

	currentDir := req.GetStringTrimmed(r, "current_dir")

	filePath, err := verifyAndNormalizePathOrError(currentDir, selectedFileName)
	if err != nil {
		return api.Error("invalid file path: " + err.Error()).ToString()
	}

	if u.Storage() == nil {
		return api.Error("Storage not initialized").ToString()
	}

	// Use provided new filename or auto-generate one
	newFileName := req.GetStringTrimmed(r, "new_file")
	if newFileName == "" {
		ext := filepath.Ext(selectedFileName)
		base := strings.TrimSuffix(selectedFileName, ext)
		newFileName = base + "_copy" + ext
	}
	newFilePath, err := verifyAndNormalizePathOrError(currentDir, newFileName)
	if err != nil {
		return api.Error("invalid file path: " + err.Error()).ToString()
	}

	// Check if target already exists and append number if needed.
	// Use the extension of newFileName (not selectedFileName) so the
	// generated copy name preserves the correct extension even when the
	// user provides a new_file with a different extension.
	exists, err := u.Storage().Exists(newFilePath)
	if err != nil {
		return api.Error("Failed to check if file exists: " + err.Error()).ToString()
	}
	const maxCopyAttempts = 1000
	counter := 2
	ext := filepath.Ext(newFileName)
	base := strings.TrimSuffix(newFileName, ext)
	for exists {
		if counter > maxCopyAttempts {
			return api.Error("Failed to find a unique file name after " + fmt.Sprintf("%d", maxCopyAttempts) + " attempts").ToString()
		}
		newFileName = fmt.Sprintf("%s_copy_%d%s", base, counter, ext)
		newFilePath, err = verifyAndNormalizePathOrError(currentDir, newFileName)
		if err != nil {
			return api.Error("invalid file path: " + err.Error()).ToString()
		}
		exists, err = u.Storage().Exists(newFilePath)
		if err != nil {
			return api.Error("Failed to check if file exists: " + err.Error()).ToString()
		}
		counter++
	}

	errClone := u.Storage().Copy(filePath, newFilePath)

	if errClone == nil {
		return api.SuccessWithData("file cloned successfully", map[string]any{
			"new_file_name": newFileName,
		}).ToString()
	}

	return api.Error(errClone.Error()).ToString()
}
