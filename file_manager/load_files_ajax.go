package file_manager

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/dracory/api"
	"github.com/dracory/req"
	"github.com/dromara/carbon/v2"
	"github.com/samber/lo"
)

// handleLoadFilesAjax returns directory contents as JSON for the Vue app
func (u *ui) handleLoadFilesAjax(r *http.Request) string {
	if u.Storage() == nil {
		return api.Error("storage is required").ToString()
	}

	currentDirectory := req.GetStringTrimmed(r, "current_dir")
	currentDirectory = strings.Trim(currentDirectory, "/")

	// Validate the directory path to prevent path traversal attacks.
	// strings.Trim only strips leading/trailing characters and does not
	// catch embedded ".." sequences (e.g. "docs/../secret"). Use the
	// same validator as all other handlers.
	if currentDirectory != "" {
		normalized, err := verifyAndNormalizeDirPath("", currentDirectory)
		if err != nil {
			return api.Error("invalid current directory: " + err.Error()).ToString()
		}
		currentDirectory = normalized
	}

	parentDirectory := ""
	if currentDirectory != "" && currentDirectory != "/" {
		parentDirectory = filepath.Dir(currentDirectory)
	}

	parentDirectory = strings.Trim(parentDirectory, "/")

	if currentDirectory == "" {
		currentDirectory = u.RootDirPath()
	}

	directories, err := u.Storage().Directories(currentDirectory)
	if err != nil {
		return api.Error(err.Error()).ToString()
	}

	files, err := u.Storage().Files(currentDirectory)
	if err != nil {
		return api.Error(err.Error()).ToString()
	}

	directoryList := []FileEntry{}
	for _, dir := range directories {
		size, _ := u.Storage().Size(dir)
		hSize := lo.If(size > 0, u.HumanFilesize(size)).Else("-")
		modified, _ := u.Storage().LastModified(dir)
		hModified := lo.If(lo.IsEmpty(modified), "-").Else(carbon.CreateFromStdTime(modified).ToDateTimeString())
		directoryList = append(directoryList, FileEntry{
			Path:              dir,
			Name:              filepath.Base(dir),
			Size:              size,
			SizeHuman:         hSize,
			LastModified:      modified,
			LastModifiedHuman: hModified,
		})
	}

	fileList := []FileEntry{}
	for _, file := range files {
		size, _ := u.Storage().Size(file)
		hSize := u.HumanFilesize(size)
		modified, _ := u.Storage().LastModified(file)
		hModified := carbon.CreateFromStdTime(modified).ToDateTimeString()
		url, _ := u.Storage().Url(file)

		fileList = append(fileList, FileEntry{
			Path:              file,
			URL:               url,
			Name:              filepath.Base(file),
			Size:              size,
			SizeHuman:         hSize,
			LastModified:      modified,
			LastModifiedHuman: hModified,
		})
	}

	return api.SuccessWithData("Files loaded successfully", map[string]any{
		"current_directory": currentDirectory,
		"parent_directory":  parentDirectory,
		"directories":       directoryList,
		"files":             fileList,
	}).ToString()
}
