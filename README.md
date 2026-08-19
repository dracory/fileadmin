# fileadmin

<img src="https://opengraph.githubassets.com/dracory/fileadmin" />

[![Tests Status](https://github.com/dracory/fileadmin/actions/workflows/tests.yml/badge.svg?branch=main)](https://github.com/dracory/fileadmin/actions/workflows/tests.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dracory/fileadmin)](https://goreportcard.com/report/github.com/dracory/fileadmin)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/dracory/fileadmin)](https://pkg.go.dev/github.com/dracory/fileadmin)

## License

This project is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0). You can find a copy of the license at [https://www.gnu.org/licenses/agpl-3.0.en.html](https://www.gnu.org/licenses/agpl-3.0.txt)

For commercial use, please use my [contact page](https://lesichkov.co.uk/contact) to obtain a commercial license.

## Introduction

Admin interface for [`github.com/dracory/filesystem`](https://github.com/dracory/filesystem).
Provides a ready-to-use admin panel for managing files and directories
stored in a SQL-backed filesystem.

Modeled after [`github.com/dracory/shopadmin`](https://github.com/dracory/shopadmin),
[`github.com/dracory/blogadmin`](https://github.com/dracory/blogadmin), and
[`github.com/dracory/logadmin`](https://github.com/dracory/logadmin)
— same folder-per-controller pattern, same `UiConfig`/`UiBase` conventions.

## Features

- **File browsing** — navigate directories, view file metadata
- **File upload** — upload files up to 50MB
- **File operations** — rename, clone (duplicate), delete
- **Directory operations** — create, delete directories
- **Bulk actions** — bulk move and bulk delete with item selection
- **Move destinations** — filtered list of valid move targets (excludes
  moving a directory into itself or its subdirectories)
- **Path traversal protection** — all paths are validated to prevent
  directory traversal attacks
- **Custom layouts** — bring your own layout via `FuncLayout`
- **Bootstrap + Vue CDN** — default UI works out of the box

## Installation

```bash
go get github.com/dracory/fileadmin
```

## Quick Start

```go
package main

import (
    "log/slog"
    "net/http"
    "os"

    "github.com/dracory/fileadmin"
    "github.com/dracory/filesystem"
)

func main() {
    storage, err := filesystem.NewStorage(filesystem.Disk{
        DiskName:  filesystem.DRIVER_SQL,
        Driver:    filesystem.DRIVER_SQL,
        Url:       "/files",
        DB:        yourDB,
        TableName: "snv_files_file",
    })
    if err != nil {
        log.Fatal(err)
    }

    admin, err := fileadmin.New(fileadmin.AdminOptions{
        Storage:      storage,
        RootDirPath:  "/uploads",
        AdminHomeURL: "/admin",
        FileAdminURL: "/admin/file-manager",
    })
    if err != nil {
        log.Fatal(err)
    }

    http.Handle("/admin/file-manager", http.HandlerFunc(admin.Handle))
    http.ListenAndServe(":8080", nil)
}
```

See [`example/`](example/) for a complete runnable server with
in-memory SQLite.

## Integration with a Router

`fileadmin.AdminInterface` exposes `Handle(w, r)`, which is an
`http.HandlerFunc`-compatible method. Wire it into any router that
accepts standard `http.Handler`:

```go
// stdlib
mux.Handle("/admin/file-manager", http.HandlerFunc(admin.Handle))

// github.com/dracory/rtr
route := rtr.NewRoute().
    SetName("Admin > File Manager").
    SetPath("/admin/file-manager").
    SetHTMLHandler(admin.Handle)
```

## Custom Layout

By default, fileadmin renders a bare-bones HTML page with Bootstrap and
Vue from CDN. To embed the admin inside your own layout (branding, menus,
etc.), provide `FuncLayout`:

```go
admin, _ := fileadmin.New(fileadmin.AdminOptions{
    Storage:     storage,
    RootDirPath: "/uploads",
    FuncLayout: func(w http.ResponseWriter, r *http.Request, title, body string, opts struct {
        Styles     []string
        StyleURLs  []string
        Scripts    []string
        ScriptURLs []string
    }) string {
        return myLayout(w, r, title, body, opts)
    },
})
```

The anonymous struct matches shopadmin/blogadmin/logadmin exactly, so you
can reuse your existing layout function.

## Testing

```bash
go test ./...
```

Tests use an in-memory SQLite database via `modernc.org/sqlite` — no
external services required.
