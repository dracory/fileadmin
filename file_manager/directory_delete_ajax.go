package file_manager

import (
	"net/http"
	"strings"

	"github.com/dracory/api"
	"github.com/dracory/req"
)

// directoryDeleteAjax handles directory deletion requests
func (u *ui) directoryDeleteAjax(r *http.Request) string {
	selectedDirName := req.GetStringTrimmed(r, "delete_dir")

	if selectedDirName == "" {
		return api.Error("delete_dir is required").ToString()
	}

	currentDir := req.GetStringTrimmed(r, "current_dir")
	if strings.TrimSpace(currentDir) == "" {
		currentDir = u.RootDirPath()
	}

	dirPath, err := verifyAndNormalizeDirPath(currentDir, selectedDirName)
	if err != nil {
		return api.Error("invalid directory name: " + err.Error()).ToString()
	}

	if dirPath == "" || dirPath == "/" {
		return api.Error("root directory can not be deleted").ToString()
	}

	if u.Storage() == nil {
		return api.Error("Storage not initialized").ToString()
	}

	errDeleted := u.Storage().DeleteDirectory(dirPath)

	if errDeleted == nil {
		return api.Success("directory deleted successfully").ToString()
	}

	return api.Error(errDeleted.Error()).ToString()
}
