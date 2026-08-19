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

	// Read current_dir from request. This may be root-relative (e.g.
	// "/1234") or a full storage path (e.g. "/uploads/1234") for
	// backward compatibility. resolveCurrentDir normalizes it.
	rawCurrentDir := req.GetStringTrimmed(r, "current_dir")

	// Validate the raw path to prevent path traversal attacks before
	// resolving. Use the root-relative portion only for validation.
	validatePath := strings.Trim(rawCurrentDir, "/")
	if validatePath != "" {
		normalized, err := verifyAndNormalizeDirPath("", validatePath)
		if err != nil {
			return api.Error("invalid current directory: " + err.Error()).ToString()
		}
		// Keep the validated, trimmed form for resolution below
		rawCurrentDir = normalized
	}

	// Resolve to full storage path (prepend root if needed)
	fullCurrentDir := u.resolveCurrentDir(rawCurrentDir)

	// Compute parent directory (root-relative for the frontend).
	// When the parent is root, return "" so the frontend's parent link
	// navigates to root (empty current_dir = root). For non-root
	// parents, return with a leading slash (e.g. "/parent") to match
	// the format of current_directory and directory click paths.
	parentDirectory := ""
	rootRelativeCurrent := u.stripRootPrefix(fullCurrentDir)
	if rootRelativeCurrent != "" && rootRelativeCurrent != "/" {
		parent := filepath.Dir(rootRelativeCurrent)
		parent = strings.ReplaceAll(parent, "\\", "/")
		parent = strings.Trim(parent, "/")
		// filepath.Dir("/1234") returns "/" → trimmed to "" → root.
		// filepath.Dir("/a/b") returns "/a" → trimmed to "a".
		if parent == "." || parent == "" {
			parentDirectory = ""
		} else {
			parentDirectory = "/" + parent
		}
	}

	directories, err := u.Storage().Directories(fullCurrentDir)
	if err != nil {
		return api.Error(err.Error()).ToString()
	}

	files, err := u.Storage().Files(fullCurrentDir)
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
			Path:              u.stripRootPrefix(dir),
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
			Path:              u.stripRootPrefix(file),
			URL:               url,
			Name:              filepath.Base(file),
			Size:              size,
			SizeHuman:         hSize,
			LastModified:      modified,
			LastModifiedHuman: hModified,
		})
	}

	return api.SuccessWithData("Files loaded successfully", map[string]any{
		"current_directory": rootRelativeCurrent,
		"parent_directory":  parentDirectory,
		"root_directory":    u.RootDirPath(),
		"directories":       directoryList,
		"files":             fileList,
	}).ToString()
}
