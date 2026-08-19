package testutils

import (
	"database/sql"
	"os"

	"github.com/dracory/filesystem"
)

// InitStorage creates an in-memory SQL file storage for testing.
//
// Note: The SQLite driver (modernc.org/sqlite) must be imported with a
// side-effect import in _test.go files of the consuming package, NOT here.
// This avoids double driver registration panics when a host project that
// already imports modernc.org/sqlite also imports this testutils package.
//
// Example (in your _test.go file):
//
//	import (
//	    "github.com/dracory/fileadmin/testutils"
//	    _ "modernc.org/sqlite"
//	)
func InitStorage(filepath string) (filesystem.StorageInterface, *sql.DB, error) {
	db, err := initDB(filepath)
	if err != nil {
		return nil, nil, err
	}

	storage, err := filesystem.NewStorage(filesystem.Disk{
		DiskName:  filesystem.DRIVER_SQL,
		Driver:    filesystem.DRIVER_SQL,
		Url:       "/files",
		DB:        db,
		TableName: "snv_files_file",
	})
	if err != nil {
		return nil, nil, err
	}

	return storage, db, nil
}

func initDB(filepath string) (*sql.DB, error) {
	if filepath != ":memory:" {
		err := os.Remove(filepath)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}

	dsn := filepath + "?parseTime=true"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	return db, nil
}
