package file_manager

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// FileEntry represents a file or directory entry
type FileEntry struct {
	IsDir             bool
	Path              string
	URL               string
	Name              string
	Size              int64
	SizeHuman         string
	LastModified      time.Time
	LastModifiedHuman string
	Depth             int
}

// HumanFilesize converts bytes to human readable format
func (u *ui) HumanFilesize(size int64) string {
	const unit = 1000
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(size)/float64(div), "kMGTPE"[exp])
}

// verifyAndNormalizePathOrError normalizes a directory path by handling root directory representation, removing double slashes, and preventing path traversal
func verifyAndNormalizePathOrError(dir, filename string) (string, error) {
	if dir == "/" {
		dir = ""
	}

	// Check for path traversal attempts BEFORE cleaning
	// This prevents any attempt to use ".." to escape the directory structure
	if dir == "." || dir == ".." {
		return "", errors.New("path traversal detected")
	}
	if filename == "." || filename == ".." {
		return "", errors.New("path traversal detected")
	}
	if strings.HasPrefix(dir, "~") || strings.HasPrefix(filename, "~") {
		return "", errors.New("path traversal detected")
	}
	if strings.Contains(dir, "..") {
		return "", errors.New("path traversal detected")
	}
	if strings.Contains(filename, "..") {
		return "", errors.New("path traversal detected")
	}

	// Clean the filename to resolve any other relative components
	filename = filepath.Clean(filename)

	path := dir + "/" + filename
	path = strings.ReplaceAll(path, "//", "/")

	// Clean the final path to resolve any remaining relative components
	path = filepath.Clean(path)

	// Ensure forward slashes are used consistently (Windows compatibility)
	path = strings.ReplaceAll(path, "\\", "/")

	return path, nil
}

// verifyAndNormalizeDirPath normalizes a directory path by handling root directory representation, removing double slashes, and trimming trailing slashes
func verifyAndNormalizeDirPath(dir, filename string) (string, error) {
	path, err := verifyAndNormalizePathOrError(dir, filename)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(path, "/"), nil
}

// allDirectories recursively lists all directories starting from the given path
func (u *ui) allDirectories(rootPath string) ([]FileEntry, error) {
	return u.allDirectoriesDepth(rootPath, 0)
}

// resolveCurrentDir resolves a current_dir parameter (which may be
// root-relative, e.g. "/1234" or "1234") to a full storage path
// (e.g. "/uploads/1234"). Empty string or "/" resolves to the root
// directory. Paths already starting with the root directory are
// returned as-is (backward compatibility).
func (u *ui) resolveCurrentDir(currentDir string) string {
	currentDir = strings.Trim(currentDir, "/")
	if currentDir == "" {
		return u.RootDirPath()
	}

	rootDir := strings.Trim(u.RootDirPath(), "/")
	if currentDir == rootDir {
		return u.RootDirPath()
	}

	// Already includes root prefix (backward compat)
	if strings.HasPrefix(currentDir, rootDir+"/") {
		return "/" + currentDir
	}

	// Root-relative path — prepend root
	return u.RootDirPath() + "/" + currentDir
}

// stripRootPrefix removes the root directory prefix from a full storage
// path, returning a root-relative path. e.g. "/uploads/1234" → "/1234".
// If the path IS the root directory, returns "" (root). If the path
// doesn't start with the root prefix, it is returned unchanged.
func (u *ui) stripRootPrefix(path string) string {
	rootDir := u.RootDirPath()
	if path == rootDir {
		return ""
	}
	if strings.HasPrefix(path, rootDir+"/") {
		return strings.TrimPrefix(path, rootDir)
	}
	return path
}

func (u *ui) allDirectoriesDepth(rootPath string, depth int) ([]FileEntry, error) {
	result := []FileEntry{}
	dirs, err := u.Storage().Directories(rootPath)
	if err != nil {
		return nil, err
	}

	for _, dir := range dirs {
		if dir == "." || dir == ".." {
			continue
		}
		size, _ := u.Storage().Size(dir)
		hSize := "-"
		if size > 0 {
			hSize = u.HumanFilesize(size)
		}
		modified, _ := u.Storage().LastModified(dir)
		result = append(result, FileEntry{
			Path:              dir,
			Name:              filepath.Base(dir),
			Size:              size,
			SizeHuman:         hSize,
			LastModified:      modified,
			LastModifiedHuman: "",
			Depth:             depth,
		})

		subDirs, err := u.allDirectoriesDepth(dir, depth+1)
		if err != nil {
			return nil, err
		}
		result = append(result, subDirs...)
	}

	// Sort by path for consistent ordering
	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i].Path) < strings.ToLower(result[j].Path)
	})

	return result, nil
}
