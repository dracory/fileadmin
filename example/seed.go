package main

import (
	"fmt"
	"log/slog"

	"github.com/dracory/filesystem"
)

// seedStorage populates the storage with sample directories and files so
// the file admin panel has content to browse on first launch.
//
// Creates:
//   - /uploads/documents (with 3 sample text files)
//   - /uploads/images    (with 2 sample "image" files)
//   - /uploads/projects  (with 2 subdirectories and files)
//   - /uploads/notes.txt (root-level file)
func seedStorage(storage filesystem.StorageInterface, rootDir string, logger *slog.Logger) {
	seedDir(storage, logger, rootDir+"/documents")
	seedFile(storage, logger, rootDir+"/documents/readme.txt", "This is a sample readme file.\nIt demonstrates text file browsing.")
	seedFile(storage, logger, rootDir+"/documents/todo.md", "# Todo\n\n- [x] Create fileadmin\n- [ ] Conquer the world")
	seedFile(storage, logger, rootDir+"/documents/notes.txt", "Some quick notes for testing the file manager.")

	seedDir(storage, logger, rootDir+"/images")
	seedFile(storage, logger, rootDir+"/images/logo.txt", "placeholder for logo.png")
	seedFile(storage, logger, rootDir+"/images/banner.txt", "placeholder for banner.jpg")

	seedDir(storage, logger, rootDir+"/projects")
	seedDir(storage, logger, rootDir+"/projects/alpha")
	seedDir(storage, logger, rootDir+"/projects/beta")
	seedFile(storage, logger, rootDir+"/projects/alpha/main.go", "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n")
	seedFile(storage, logger, rootDir+"/projects/alpha/config.json", `{"name":"alpha","version":"1.0.0"}`)
	seedFile(storage, logger, rootDir+"/projects/beta/README.md", "# Beta Project\n\nA sample project for testing nested directories.")

	seedFile(storage, logger, rootDir+"/welcome.txt", "Welcome to fileadmin!\nThis is a root-level sample file.")

	logger.Info("seedStorage complete",
		"directories", 5,
		"files", 9,
	)
}

func seedDir(storage filesystem.StorageInterface, logger *slog.Logger, path string) {
	if err := storage.MakeDirectory(path); err != nil {
		logger.Error("seedDir: failed to create directory", "path", path, "error", err)
	}
}

func seedFile(storage filesystem.StorageInterface, logger *slog.Logger, path, content string) {
	if err := storage.Put(path, []byte(content)); err != nil {
		logger.Error("seedFile: failed to create file", "path", path, "error", err)
	}
}

// seedSummary returns a human-readable summary of seeded content for the
// landing page.
func seedSummary() string {
	return fmt.Sprintf("Seeded with %d directories and %d files", 5, 9)
}
