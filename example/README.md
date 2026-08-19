# fileadmin example

A minimal, runnable server that mounts the fileadmin panel on an
in-memory SQLite database. No external services required.

## Run

From the `fileadmin` module root:

```bash
go run ./example
```

Then open <http://localhost:8080/> in your browser and click
**Open File Admin**, or go directly to <http://localhost:8080/admin/file-manager>.

## What you get

- `/` — landing page with a link into the admin
- `/admin/file-manager` — file manager with directory browsing, file
  upload, rename, clone, delete, bulk move/delete, and directory
  create/delete

## Configuration

Edit the constants at the top of [`main.go`](main.go) to change:

- `addr` — listen address (default `:8080`)
- `dbFile` — `:memory:` for an ephemeral DB, or a file path to persist
  data across restarts
- `adminURL`, `homeURL`, `rootDir` — mount points and root directory

## Notes

- `AuthUserID` is intentionally nil, so the panel is open. Provide a
  real implementation in production.
- The in-memory database is reset on every restart. Pass a file path
  to `openDB` to persist data.
