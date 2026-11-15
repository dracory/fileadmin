package shared

import (
	"net/http"

	"github.com/dracory/cdn"
	"github.com/dracory/filesystem"
	"github.com/dracory/hb"
)

// Layout is the default layout renderer. It builds a complete HTML page
// using hb.NewWebpage() with Bootstrap CDN, matching the
// shopadmin/blogadmin/logadmin pattern. If a FuncLayout is provided in
// AdminOptions, it takes precedence over this default.
func Layout(w http.ResponseWriter, r *http.Request, webpageTitle, webpageHtml string, options struct {
	Styles     []string
	StyleURLs  []string
	Scripts    []string
	ScriptURLs []string
}) string {
	_ = r // kept for layout signature compatibility
	_ = w
	return webpageComplete(webpageTitle, webpageHtml, options).ToHTML()
}

// webpageComplete builds the default webpage template
func webpageComplete(title, content string, options struct {
	Styles     []string
	StyleURLs  []string
	Scripts    []string
	ScriptURLs []string
}) *hb.HtmlWebpage {
	webpage := hb.NewWebpage()
	webpage.SetTitle(title)

	webpage.AddStyleURLs([]string{
		cdn.BootstrapCss_5_3_3(),
		"https://cdn.jsdelivr.net/npm/bootstrap-icons@1.11.3/font/bootstrap-icons.min.css",
	})
	webpage.AddStyleURLs(options.StyleURLs)
	// Vue is intentionally NOT loaded here. Pages that need it (e.g. the
	// file manager) pass cdn.VueJs_3() via options.ScriptURLs, so it is
	// loaded exactly once regardless of whether a custom FuncLayout or
	// this default layout is used.
	webpage.AddScriptURLs([]string{
		cdn.BootstrapJs_5_3_3(),
		cdn.Jquery_3_7_1(),
		cdn.Sweetalert2_11(),
	})
	webpage.AddScriptURLs(options.ScriptURLs)
	webpage.AddScripts(options.Scripts)
	webpage.AddStyle(`html,body{height:100%;font-family: Ubuntu, sans-serif;}`)
	webpage.AddStyle(`body {
		font-family: "Nunito", sans-serif;
		font-size: 0.9rem;
		font-weight: 400;
		line-height: 1.6;
		color: #212529;
		text-align: left;
		background-color: #f8fafc;
	}`)
	// The [v-cloak]{display:none} style is intentionally NOT set here.
	// Vue pages pass it via options.Styles so it applies regardless of
	// which layout renders the page.
	webpage.AddStyles(options.Styles)
	webpage.AddChild(hb.NewHTML(content))
	return webpage
}

// UiConfig holds the dependencies passed to subcontroller UI factories.
// The Layout function uses an anonymous struct for options to match
// shopadmin/blogadmin/logadmin exactly, allowing consumers to reuse their
// existing layout function for fileadmin.
type UiConfig struct {
	Storage     filesystem.StorageInterface
	RootDirPath string
	Layout      func(w http.ResponseWriter, r *http.Request, webpageTitle, webpageHtml string, options struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}) string
}
