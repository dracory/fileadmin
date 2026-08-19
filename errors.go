package fileadmin

import "errors"

// Common errors
var (
	// ErrStorageRequired is returned when Storage is not provided
	ErrStorageRequired = errors.New("storage is required")
)
