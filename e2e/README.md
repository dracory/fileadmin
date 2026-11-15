# e2e

End-to-end tests that drive the fileadmin panel in a real headless Chrome
via [chromedp](https://github.com/chromedp/chromedp) (CDP). Unlike go-rod,
chromedp has no `leakless.exe` helper, so it avoids the Windows Defender
false positive entirely.

## How it works

- `TestMain` mounts the admin on `httptest.NewServer` (in-process, random
  port) backed by a seeded in-memory SQL storage via
  `testutils.InitStorage` — no binary build needed. Three directories
  (documents, images, projects) and four files are seeded under
  `/uploads`.
- One shared browser is launched via `chromedp.NewExecAllocator` for the
  whole run; each test opens a fresh tab with a 30s deadline (`newTab`).
- Chrome binary resolution order: `CHROME_BIN` env → Windows Playwright
  `chromium_headless_shell` → WSL-side Windows shell → Linux Playwright
  shell → chromedp's default lookup (Chrome on PATH / install dirs).

## Run

```bash
go test -v ./e2e/     # or: task e2e
```

Note: on WSL, run tests with the Windows Go toolchain (`go.exe`) —
Windows-side Chrome binds 127.0.0.1, which WSL cannot reach (no
WSL→Windows localhost forwarding). The Linux Playwright shell also
requires system libs (libnspr4 etc.) that may be missing.

## Coverage

- `file-manager` — seeded rows + directory navigation + create-directory,
  upload, delete, file-rename and directory-rename modal flows (the only
  registered controller)

## Gotchas learned

- Prefer `chromedp.WaitReady` over `chromedp.WaitVisible`: on some pages
  `WaitVisible` polls forever even though the DOM is fully loaded.
- Routing is query-param based: `?controller=file-manager`
  (see `shared/consts.go`).
- The file manager renders via Vue after an AJAX fetch — wait for the
  second `tbody tr` before asserting (v-if/v-else and v-for mean the
  data rows only exist after Vue mounts).
- `v-model` is compiled away by Vue — it does NOT appear in the rendered
  DOM. Match form inputs by `type` or `placeholder`, never by the
  `v-model` attribute.
- Tests share one store/browser across the whole run — don't assert on
  fixed row indices (other tests create rows) and don't delete seeded
  fixtures. Create uniquely-named fixtures inside the test and locate
  rows by text via `Evaluate` + `Array.find`, e.g.:
  `[...document.querySelectorAll('table tbody tr')].find(tr => tr.textContent.includes(name))`.
- Row action menus are real Bootstrap dropdowns — click the
  `.dropdown-toggle` (a small sleep is enough) before clicking the
  `.dropdown-item` inside the opened menu.
- File upload works via `chromedp.SetUploadFiles` on the modal's
  `input[type="file"]` — no real file picker needed; use `t.TempDir()`
  for fixture files.
- The rename input is prefilled with the old name via `v-model` —
  `SendKeys` would append. Set `el.value` and dispatch an `input` event
  so Vue picks up the new value (see `setVueInputJS`).
- Modals are Vue-rendered `v-if` blocks (not Bootstrap JS modals) — they
  appear as `.modal.fade.show` once the `modal` data property is set;
  wait on `.modal .modal-title`.
- "New Directory" is the only `.btn-info` button; disambiguate toolbar
  buttons by color class.
- chromedp v0.23 rewrote the API (generic `Run[T]`); this suite pins the
  stable v0.14 line with matching `cdproto`/`go-json-experiment` versions.
