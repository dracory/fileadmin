package main

import (
	"fmt"
	"log/slog"

	"github.com/dracory/filesystem"
)

// seedStats tracks the outcome of seed operations so seedSummary can
// report accurate counts on the landing page.
type seedStats struct {
	dirs  int
	files int
}

// seedStorage populates the storage with sample directories and files so
// the file admin panel has content to browse on first launch.
//
// Creates:
//   - /uploads/documents (with 3 sample text files)
//   - /uploads/images    (with 2 sample "image" files)
//   - /uploads/projects  (with 2 subdirectories and files)
//   - /uploads/notes.txt (root-level file)
func seedStorage(storage filesystem.StorageInterface, rootDir string, logger *slog.Logger) seedStats {
	stats := seedStats{}

	seedDir(storage, logger, &stats, rootDir+"/documents")
	seedFile(storage, logger, &stats, rootDir+"/documents/readme.txt", "This is a sample readme file.\nIt demonstrates text file browsing.")
	seedFile(storage, logger, &stats, rootDir+"/documents/todo.md", "# Todo\n\n- [x] Create fileadmin\n- [ ] Conquer the world")
	seedFile(storage, logger, &stats, rootDir+"/documents/notes.txt", "Some quick notes for testing the file manager.")

	seedDir(storage, logger, &stats, rootDir+"/images")
	seedFile(storage, logger, &stats, rootDir+"/images/logo.txt", "placeholder for logo.png")
	seedFile(storage, logger, &stats, rootDir+"/images/banner.txt", "placeholder for banner.jpg")

	seedDir(storage, logger, &stats, rootDir+"/projects")
	seedDir(storage, logger, &stats, rootDir+"/projects/alpha")
	seedDir(storage, logger, &stats, rootDir+"/projects/beta")
	seedFile(storage, logger, &stats, rootDir+"/projects/alpha/main.go", "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n")
	seedFile(storage, logger, &stats, rootDir+"/projects/alpha/config.json", `{"name":"alpha","version":"1.0.0"}`)
	seedFile(storage, logger, &stats, rootDir+"/projects/beta/README.md", "# Beta Project\n\nA sample project for testing nested directories.")

	seedFile(storage, logger, &stats, rootDir+"/welcome.txt", "Welcome to fileadmin!\nThis is a root-level sample file.")

	logger.Info("seedStorage complete",
		"directories", stats.dirs,
		"files", stats.files,
	)

	return stats
}

func seedDir(storage filesystem.StorageInterface, logger *slog.Logger, stats *seedStats, path string) {
	if err := storage.MakeDirectory(path); err != nil {
		logger.Error("seedDir: failed to create directory", "path", path, "error", err)
		return
	}
	stats.dirs++
}

func seedFile(storage filesystem.StorageInterface, logger *slog.Logger, stats *seedStats, path, content string) {
	if err := storage.Put(path, []byte(content)); err != nil {
		logger.Error("seedFile: failed to create file", "path", path, "error", err)
		return
	}
	stats.files++
}

// seedSummary returns a human-readable summary of seeded content for the
// landing page. It receives the actual counts from seedStorage so the
// numbers are accurate even if some operations failed.
func seedSummary(stats seedStats) string {
	return fmt.Sprintf("Seeded with %d directories and %d files.", stats.dirs, stats.files)
}
