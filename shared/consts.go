package shared

// Context keys for config values injected by Handle()
const KeyEndpoint = "endpoint"
const KeyAdminHomeURL = "admin_home_url"
const KeyFileAdminURL = "file_admin_url"

// Controller names used in the ?controller= query parameter
const (
	CONTROLLER_FILE_MANAGER = "file-manager"
)

// CatchAll is the catch-all route suffix
const CatchAll = "/*"
