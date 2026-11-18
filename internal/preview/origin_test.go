package preview_test

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/preview"
	"github.com/quiqr/chatcms-app/internal/workspace"
)

type viewed struct {
	view     string
	key      workspace.Key
	endpoint workspace.Endpoint
}

func (v viewed) ByView(view string) (workspace.Key, workspace.Endpoint, bool) {
	if view != v.view {
		return workspace.Key{}, workspace.Endpoint{}, false
	}
	return v.key, v.endpoint, true
}

func originRig(t *testing.T, allowed bool, check error) (*preview.Origin, *[]string) {
	t.Helper()
	var asked []string
	site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		asked = append(asked, r.URL.Path)
		io.WriteString(w, "the site at "+r.URL.Path)
	}))
	t.Cleanup(site.Close)
	host, port, _ := net.SplitHostPort(strings.TrimPrefix(site.URL, "http://"))
	number, _ := strconv.Atoi(port)
	known := viewed{view: "the-key", key: workspace.Key{Person: uuid.New(), Project: uuid.New()}, endpoint: workspace.Endpoint{Host: host, Port: number}}
	origin := preview.NewOrigin(nil, known, func(ctx context.Context, key workspace.Key) (bool, error) {
		if key != known.key {
			t.Errorf("asked about %v", key)
		}
		return allowed, check
	})
	return origin, &asked
}

func ask(origin http.Handler, path string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	origin.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "http://preview.example"+path, nil))
	return recorder
}

func TestTheOriginServesAWorkspaceByItsKey(t *testing.T) {
	origin, asked := originRig(t, true, nil)
	answer := ask(origin, "/v/the-key/posts/one/")
	if answer.Code != http.StatusOK || !strings.Contains(answer.Body.String(), "the site at /v/the-key/posts/one/") {
		t.Errorf("answer = %d %s", answer.Code, answer.Body)
	}
	if len(*asked) != 1 {
		t.Errorf("the workspace was asked %d times", len(*asked))
	}
}

func TestTheOriginServesNothingElse(t *testing.T) {
	origin, asked := originRig(t, true, nil)
	for _, path := range []string{"/", "/api/organisations", "/v/", "/v/wrong-key/", "/preview/o/p/", "/sign-in"} {
		if answer := ask(origin, path); answer.Code != http.StatusNotFound {
			t.Errorf("%s = %d, want 404", path, answer.Code)
		}
	}
	if len(*asked) != 0 {
		t.Errorf("a workspace was asked for %v", *asked)
	}
}

func TestTheOriginChecksTheRightToReadOnEveryRequest(t *testing.T) {
	origin, asked := originRig(t, false, nil)
	if answer := ask(origin, "/v/the-key/"); answer.Code != http.StatusNotFound || len(*asked) != 0 {
		t.Errorf("without the right = %d, asked %v", answer.Code, *asked)
	}
	broken, _ := originRig(t, true, errors.New("database down"))
	if answer := ask(broken, "/v/the-key/"); answer.Code != http.StatusServiceUnavailable {
		t.Errorf("when the right cannot be checked = %d", answer.Code)
	}
}

func TestAPreviewOnItsOwnOriginIsAddressedByAKey(t *testing.T) {
	h := newHarness(t, 1)
	if h.service.Apart() || h.service.Under("/preview/o/p/") != "http://example.org/preview/o/p/" {
		t.Errorf("without an origin Under() = %q", h.service.Under("/preview/o/p/"))
	}
	h.service.ServeFrom("https://preview.example/")
	if !h.service.Apart() || h.service.Under("/preview/o/p/") != "https://preview.example/v/{view}/" {
		t.Errorf("with an origin Under() = %q", h.service.Under("/preview/o/p/"))
	}
}
