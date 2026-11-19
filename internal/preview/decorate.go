package preview

import (
	"bytes"
	"html"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/quiqr/chatcms-app/internal/workspace"
)

const Assets = "__chatcms/"

func (s *Service) ServeAssets(assets fs.FS) {
	s.assets = assets
}

func (s *Service) appOrigin() string {
	parsed, err := url.Parse(s.baseURL)
	if err != nil || parsed.Host == "" {
		return ""
	}
	return parsed.Scheme + "://" + parsed.Host
}

func (s *Service) decorate(where *url.URL, prefix string, rooted bool) http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(where)
	tag := []byte(`<script type="module" src="` + html.EscapeString(prefix+Assets+"select.js") + `" data-app="` + html.EscapeString(s.appOrigin()) + `"></script>`)
	director := proxy.Director
	proxy.Director = func(r *http.Request) {
		if rooted {
			r.URL.Path = "/" + inside(r.URL.Path, prefix)
			if r.URL.RawPath != "" {
				r.URL.RawPath = "/" + inside(r.URL.RawPath, prefix)
			}
		}
		director(r)
		r.Header.Del("Accept-Encoding")
	}
	proxy.ModifyResponse = func(answer *http.Response) error {
		return inject(answer, tag)
	}
	assets := s.assets
	meta := []byte(`<meta name="chatcms-app" content="` + html.EscapeString(s.appOrigin()) + `">`)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		within := strings.TrimPrefix(r.URL.Path, prefix)
		if rooted {
			within = inside(r.URL.Path, prefix)
		}
		if rest, found := strings.CutPrefix(within, Assets); found && strings.HasPrefix(r.URL.Path, prefix) {
			serveAsset(w, r, assets, rest, meta)
			return
		}
		proxy.ServeHTTP(w, r)
	})
}

func inside(path, prefix string) string {
	return strings.TrimPrefix(strings.TrimPrefix(path, prefix), workspace.RootedPath)
}

func serveAsset(w http.ResponseWriter, r *http.Request, assets fs.FS, name string, meta []byte) {
	if assets == nil || (r.Method != http.MethodGet && r.Method != http.MethodHead) {
		http.NotFound(w, r)
		return
	}
	if name == "" || strings.HasSuffix(name, "/") {
		name += "index.html"
	}
	if !fs.ValidPath(name) {
		http.NotFound(w, r)
		return
	}
	data, err := fs.ReadFile(assets, name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if strings.HasSuffix(name, ".html") {
		data = withMeta(data, meta)
	}
	kind := mime.TypeByExtension(name[strings.LastIndex(name, "."):])
	if strings.HasSuffix(name, ".mjs") {
		kind = "text/javascript; charset=utf-8"
	}
	if kind != "" {
		w.Header().Set("Content-Type", kind)
	}
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(w, r, name, timeZero, bytes.NewReader(data))
}

func withMeta(page, meta []byte) []byte {
	at := bytes.Index(bytes.ToLower(page), []byte("<head>"))
	if at < 0 {
		return append(append([]byte{}, meta...), page...)
	}
	at += len("<head>")
	out := make([]byte, 0, len(page)+len(meta))
	out = append(out, page[:at]...)
	out = append(out, meta...)
	return append(out, page[at:]...)
}

func inject(answer *http.Response, tag []byte) error {
	if answer.StatusCode != http.StatusOK || answer.Header.Get("Content-Encoding") != "" {
		return nil
	}
	kind, _, _ := mime.ParseMediaType(answer.Header.Get("Content-Type"))
	if kind != "text/html" {
		return nil
	}
	page, err := io.ReadAll(answer.Body)
	answer.Body.Close()
	if err != nil {
		return err
	}
	at := bytes.LastIndex(bytes.ToLower(page), []byte("</body>"))
	var out []byte
	if at < 0 {
		out = append(page, tag...)
	} else {
		out = make([]byte, 0, len(page)+len(tag))
		out = append(out, page[:at]...)
		out = append(out, tag...)
		out = append(out, page[at:]...)
	}
	answer.Body = io.NopCloser(bytes.NewReader(out))
	answer.ContentLength = int64(len(out))
	answer.Header.Set("Content-Length", strconv.Itoa(len(out)))
	return nil
}

var timeZero time.Time
