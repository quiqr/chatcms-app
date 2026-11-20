package preview_test

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/preview"
	"github.com/quiqr/chatcms-app/internal/workspace"
)

type idle struct{}

func (idle) Endpoint(context.Context, workspace.Key, string) (workspace.Endpoint, error) {
	return workspace.Endpoint{}, workspace.ErrNotReady
}
func (idle) Open(context.Context, workspace.Key, string) (workspace.State, error)    { return "", nil }
func (idle) Report(workspace.Key) (workspace.Report, bool)                           { return workspace.Report{}, false }
func (idle) Restart(context.Context, workspace.Key, string) (workspace.State, error) { return "", nil }
func (idle) SleepProject(context.Context, string)                                    {}

func decorated(t *testing.T, site http.HandlerFunc) (http.Handler, *[]http.Header) {
	t.Helper()
	return decoratedAs(t, false, site)
}

func decoratedAs(t *testing.T, rooted bool, site http.HandlerFunc) (http.Handler, *[]http.Header) {
	t.Helper()
	return decoratedWith(t, rooted, true, site)
}

func decoratedWith(t *testing.T, rooted, apart bool, site http.HandlerFunc) (http.Handler, *[]http.Header) {
	t.Helper()
	var asked []http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		asked = append(asked, r.Header.Clone())
		site(w, r)
	}))
	t.Cleanup(server.Close)
	host, port, _ := net.SplitHostPort(strings.TrimPrefix(server.URL, "http://"))
	number, _ := strconv.Atoi(port)

	service, err := preview.NewService(sources{organisation: "o"}, idle{}, "https://app.example/")
	if err != nil {
		t.Fatal(err)
	}
	if apart {
		service.ServeFrom("https://preview.example")
	}
	service.ServeAssets(fstest.MapFS{
		"select.js":          {Data: []byte("console.log('select')")},
		"pdf/index.html":     {Data: []byte("<!doctype html><head><title>viewer</title>")},
		"pdf/assets/view.js": {Data: []byte("view()")},
	})
	known := viewed{view: "k", key: workspace.Key{Person: uuid.New(), Project: uuid.New()}, endpoint: workspace.Endpoint{Host: host, Port: number, Rooted: rooted}}
	return preview.NewOrigin(service, known, func(context.Context, workspace.Key) (bool, error) { return true, nil }), &asked
}

func TestAnHTMLPageGetsTheSelectionScript(t *testing.T) {
	origin, asked := decorated(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, "<html><BODY><p>Fresh bread</p></BODY></html>")
	})
	answer := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "http://preview.example/v/k/posts/one/", nil)
	request.Header.Set("Accept-Encoding", "br, zstd")
	origin.ServeHTTP(answer, request)
	body := answer.Body.String()
	want := `<p>Fresh bread</p><script type="module" src="/v/k/__chatcms/select.js" data-app="https://app.example"></script></BODY>`
	if !strings.Contains(body, want) {
		t.Errorf("page = %s", body)
	}
	if answer.Header().Get("Content-Length") != strconv.Itoa(len(body)) {
		t.Errorf("Content-Length = %s for %d bytes", answer.Header().Get("Content-Length"), len(body))
	}
	if got := (*asked)[0].Get("Accept-Encoding"); strings.Contains(got, "br") || strings.Contains(got, "zstd") {
		t.Errorf("the workspace was asked for %q", got)
	}
}

func TestAPageWithoutABodyGetsTheScriptAtItsEnd(t *testing.T) {
	origin, _ := decorated(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, "<p>bare")
	})
	if body := ask(origin, "/v/k/").Body.String(); !strings.HasPrefix(body, "<p>bare<script ") || !strings.HasSuffix(body, "</script>") {
		t.Errorf("page = %s", body)
	}
}

func TestEverythingElsePassesUnchanged(t *testing.T) {
	for _, tc := range []struct {
		kind   string
		status int
	}{{"text/css", 200}, {"application/pdf", 200}, {"text/html", 404}} {
		origin, _ := decorated(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", tc.kind)
			w.WriteHeader(tc.status)
			io.WriteString(w, "body{}")
		})
		if body := ask(origin, "/v/k/x").Body.String(); body != "body{}" {
			t.Errorf("%s %d became %q", tc.kind, tc.status, body)
		}
	}
}

func TestTheAssetsAreServedNotProxied(t *testing.T) {
	origin, asked := decorated(t, func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, "from the site") })
	for path, want := range map[string]string{
		"/v/k/__chatcms/select.js":          "console.log('select')",
		"/v/k/__chatcms/pdf/":               `<!doctype html><head><meta name="chatcms-app" content="https://app.example"><title>viewer</title>`,
		"/v/k/__chatcms/pdf/assets/view.js": "view()",
	} {
		answer := ask(origin, path)
		if answer.Code != http.StatusOK || answer.Body.String() != want {
			t.Errorf("%s = %d %q", path, answer.Code, answer.Body)
		}
	}
	if answer := ask(origin, "/v/k/__chatcms/nothing.js"); answer.Code != http.StatusNotFound {
		t.Errorf("a missing asset = %d", answer.Code)
	}
	if ct := ask(origin, "/v/k/__chatcms/select.js").Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/javascript") {
		t.Errorf("select.js Content-Type = %q", ct)
	}
	if len(*asked) != 0 {
		t.Errorf("the workspace was asked %d times", len(*asked))
	}
}

func TestARootedPreviewIsAskedWithoutThePrefix(t *testing.T) {
	var paths []string
	origin, _ := decoratedAs(t, true, func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		io.WriteString(w, "ok")
	})
	ask(origin, "/v/k/")
	ask(origin, "/v/k/chapters/bread.html")
	ask(origin, "/v/k/a%20b.html")
	ask(origin, "/v/k/p/site/chapters/rye.html")
	if len(paths) != 4 || paths[0] != "/" || paths[1] != "/chapters/bread.html" || paths[2] != "/a b.html" || paths[3] != "/chapters/rye.html" {
		t.Errorf("the preview was asked %v", paths)
	}
	for _, path := range []string{"/v/k/__chatcms/select.js", "/v/k/p/site/__chatcms/select.js"} {
		if body := ask(origin, path).Body.String(); body != "console.log('select')" {
			t.Errorf("%s: the assets moved: %q", path, body)
		}
	}
}

func TestAPreviewToldItsAddressKeepsThePrefix(t *testing.T) {
	var asked string
	origin, _ := decoratedAs(t, false, func(w http.ResponseWriter, r *http.Request) { asked = r.URL.Path })
	ask(origin, "/v/k/posts/one/")
	if asked != "/v/k/posts/one/" {
		t.Errorf("the preview was asked %q", asked)
	}
}

func TestOnTheAppsOwnAddressTheScriptAnswersItsOwnPage(t *testing.T) {
	origin, _ := decoratedWith(t, false, false, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, "<body>x</body>")
	})
	if body := ask(origin, "/v/k/").Body.String(); !strings.Contains(body, `data-app=""`) {
		t.Errorf("page = %s", body)
	}
	if body := ask(origin, "/v/k/__chatcms/pdf/").Body.String(); !strings.Contains(body, `<meta name="chatcms-app" content="">`) {
		t.Errorf("viewer = %s", body)
	}
}

func TestTheAppsOwnAbsoluteAddressesBecomePaths(t *testing.T) {
	for name, test := range map[string]struct {
		apart bool
		base  string
	}{
		"on the app's address":         {false, "https://app.example"},
		"on an origin of its own":      {true, "https://preview.example"},
		"the app's address when apart": {true, "https://app.example"},
	} {
		page := `<link href="` + test.base + `/v/k/css/styles.css"><img src="` + test.base + `/v/k/images/sim.jpg"><a href="https://elsewhere.example/v/k/x">x</a>`
		origin, _ := decoratedWith(t, false, test.apart, func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/v/k/index.json", "/index.json":
				w.Header().Set("Content-Type", "application/json")
				io.WriteString(w, `{"url":"`+strings.ReplaceAll(test.base+"/v/k/post/", "/", `\/`)+`"}`)
			case "/v/k/logo.png", "/logo.png":
				w.Header().Set("Content-Type", "image/png")
				io.WriteString(w, test.base+"/v/k/")
			default:
				w.Header().Set("Content-Type", "text/html")
				io.WriteString(w, page)
			}
		})
		body := ask(origin, "/v/k/").Body.String()
		for _, want := range []string{`<link href="/v/k/css/styles.css">`, `<img src="/v/k/images/sim.jpg">`, `https://elsewhere.example/v/k/x`} {
			if !strings.Contains(body, want) {
				t.Errorf("%s: page lacks %s:\n%s", name, want, body)
			}
		}
		if json := ask(origin, "/v/k/index.json").Body.String(); json != `{"url":"\/v\/k\/post\/"}` {
			t.Errorf("%s: json = %s", name, json)
		}
		if png := ask(origin, "/v/k/logo.png").Body.String(); png != test.base+"/v/k/" {
			t.Errorf("%s: a picture was changed: %s", name, png)
		}
	}
}
