package file_manager

import (
	"net/http"
	"strings"

	"github.com/dracory/api"
	"github.com/dracory/req"
)

// fileDeleteAjax handles file deletion requests
func (u *ui) fileDeleteAjax(r *http.Request) string {
	selectedFileName := req.GetStringTrimmed(r, "delete_file")
	if selectedFileName == "" {
		return api.Error("delete_file is required").ToString()
	}
	currentDir := req.GetStringTrimmed(r, "current_dir")
	if strings.TrimSpace(currentDir) == "" {
		currentDir = u.RootDirPath()
	}

	filePath, err := verifyAndNormalizePathOrError(currentDir, selectedFileName)
	if err != nil {
		return api.Error("invalid file path: " + err.Error()).ToString()
	}

	if u.Storage() == nil {
		return api.Error("Storage not initialized").ToString()
	}
	errDeleted := u.Storage().DeleteFile([]string{filePath})

	if errDeleted == nil {
		return api.Success("file deleted successfully").ToString()
	}

	return api.Error(errDeleted.Error()).ToString()
}
