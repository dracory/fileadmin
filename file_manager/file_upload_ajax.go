package file_manager

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/dracory/api"
	"github.com/dracory/base/files"
	"github.com/dracory/req"
)

// fileUploadAjax handles file upload requests
func (u *ui) fileUploadAjax(r *http.Request) string {
	// Fast-path rejection for honest clients that send Content-Length.
	// r.ContentLength is -1 for chunked transfer encoding, so this alone
	// is not sufficient — the actual file size is checked below using
	// fileHeader.Size after parsing the multipart form.
	if r.ContentLength > MAX_UPLOAD_SIZE {
		return api.Error("The uploaded file is too big. Please use a file less than 50MB in size").ToString()
	}

	currentDir := req.GetStringTrimmed(r, "current_dir")
	// When current_dir is empty (root view), default to RootDirPath so
	// uploads land in the same directory that loadFiles lists from.
	if strings.TrimSpace(currentDir) == "" {
		currentDir = u.RootDirPath()
	}

	// The argument to FormFile must match the name attribute
	// of the file input on the frontend
	file, fileHeader, err := r.FormFile("upload_file")
	if err != nil {
		return api.Error(err.Error()).ToString()
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Warning: failed to close uploaded file: %v", err)
		}
	}()

	// Enforce the size limit using the actual uploaded file size, which
	// is reliable even when Content-Length is missing or spoofed (e.g.
	// chunked transfer encoding).
	if fileHeader.Size > MAX_UPLOAD_SIZE {
		return api.Error("The uploaded file is too big. Please use a file less than 50MB in size").ToString()
	}

	filePath, errSave := files.SaveToTempDir(fileHeader.Filename, file)
	if errSave != nil {
		log.Println(errSave.Error())
		return api.Error(errSave.Error()).ToString()
	}
	defer func() {
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) { // #nosec G703 -- filePath is from SaveToTempDir, not user-controlled
			log.Printf("Warning: failed to remove temp file %s: %v", filePath, err)
		}
	}()

	remoteFilePath, err := verifyAndNormalizePathOrError(currentDir, fileHeader.Filename)
	if err != nil {
		return api.Error("invalid file path: " + err.Error()).ToString()
	}

	data, err := os.ReadFile(filePath) // #nosec G703 G304 -- filePath is from SaveToTempDir, not user-controlled
	if err != nil {
		return api.Error(err.Error()).ToString()
	}

	if u.Storage() == nil {
		return api.Error("Storage not initialized").ToString()
	}

	err = u.Storage().Put(remoteFilePath, data)

	if err != nil {
		return api.Error(err.Error()).ToString()
	}

	return api.Success("File uploaded successfully").ToString()
}
