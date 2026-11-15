package e2e

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/dracory/fileadmin"
)

// TestVueAppMountsNoRawTemplate is the regression test for the bug where
// the file manager rendered raw Vue template ({{ }}, all modals stacked)
// because Vue never loaded. It asserts the app mounts: v-cloak is
// removed, Vue is defined, no {{ }} interpolations remain, and no modal
// markup is present in the rendered DOM.
func TestVueAppMountsNoRawTemplate(t *testing.T) {
	ctx := newTab(t)

	err := chromedp.Run(ctx,
		chromedp.Navigate(adminURL("file-manager")),
		chromedp.WaitReady("table tbody tr:nth-of-type(2)", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("file manager did not load: %v", err)
	}

	var body string
	var vueDefined, cloakPresent bool
	var modalCount int
	err = chromedp.Run(ctx,
		chromedp.Text("body", &body),
		chromedp.Evaluate(`typeof Vue !== 'undefined'`, &vueDefined),
		chromedp.Evaluate(`document.getElementById('files-app').hasAttribute('v-cloak')`, &cloakPresent),
		chromedp.Evaluate(`document.querySelectorAll('#files-app .modal').length`, &modalCount),
	)
	if err != nil {
		t.Fatalf("evaluating page state failed: %v", err)
	}

	if !vueDefined {
		t.Fatal("Vue is not defined — the app cannot have mounted")
	}
	if cloakPresent {
		t.Fatal("v-cloak attribute still present — Vue did not mount")
	}
	if strings.Contains(body, "{{") {
		t.Fatal("raw {{ }} interpolation visible — Vue template was not compiled")
	}
	if modalCount != 0 {
		t.Fatalf("expected 0 modals in DOM, found %d", modalCount)
	}
}

// TestVueFallbackWhenLayoutOmitsIt reproduces the dogpace scenario: a
// custom FuncLayout that renders the page WITHOUT loading Vue (ignoring
// options.ScriptURLs). files.js must detect the missing Vue and inject
// the CDN build itself, so the app still mounts.
func TestVueFallbackWhenLayoutOmitsIt(t *testing.T) {
	layoutWithoutVue := func(w http.ResponseWriter, r *http.Request,
		title, body string, options struct {
			Styles     []string
			StyleURLs  []string
			Scripts    []string
			ScriptURLs []string
		}) string {
		// Intentionally renders a bare page: no Vue, no Bootstrap,
		// ignoring everything in options.
		return "<!doctype html><html><head><title>" + title +
			"</title></head><body>" + body + "</body></html>"
	}

	panel, err := fileadmin.New(fileadmin.AdminOptions{
		Storage:      testStorage,
		RootDirPath:  rootDir,
		FuncLayout:   layoutWithoutVue,
		AdminHomeURL: "/admin",
		FileAdminURL: adminPath,
	})
	if err != nil {
		t.Fatalf("failed to create fileadmin: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle(adminPath, http.HandlerFunc(panel.Handle))
	mux.Handle(adminPath+"/", http.HandlerFunc(panel.Handle))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	url := srv.URL + adminPath + "?controller=file-manager"
	ctx := newTab(t)

	// The fallback loads Vue from the CDN asynchronously, so mounting
	// takes longer than the default path — give it room via WaitReady.
	err = chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("table tbody tr:nth-of-type(2)", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("app did not mount via Vue fallback loader: %v", err)
	}

	var mounted bool
	if err := chromedp.Run(ctx,
		chromedp.Evaluate(`document.querySelector('#files-app').__vue_app__ !== undefined`, &mounted),
	); err != nil {
		t.Fatalf("evaluating mount state failed: %v", err)
	}
	if !mounted {
		t.Fatal("Vue app not mounted despite fallback loader")
	}
}
