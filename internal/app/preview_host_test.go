package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/quiqr/chatcms-app/internal/preview"
	"github.com/quiqr/chatcms-app/internal/workspace"
)

type noViews struct{}

func (noViews) ByView(string) (workspace.Key, workspace.Endpoint, bool) {
	return workspace.Key{}, workspace.Endpoint{}, false
}

func TestThePreviewHostReachesOnlyTheOrigin(t *testing.T) {
	app := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusTeapot) })
	alone := (&Server{}).byHost(app)
	if answer := serve(alone, "preview.example", "/"); answer != http.StatusTeapot {
		t.Errorf("without an origin = %d", answer)
	}

	s := &Server{cfg: Config{PreviewURL: "https://Preview.Example"}}
	s.previewOrigin = preview.NewOrigin(nil, noViews{}, func(context.Context, workspace.Key) (bool, error) { return true, nil })
	routed := s.byHost(app)
	for _, path := range []string{"/", "/api/me", "/v/nothing/"} {
		if answer := serve(routed, "preview.example", path); answer != http.StatusNotFound {
			t.Errorf("preview host %s = %d, want 404", path, answer)
		}
	}
	if answer := serve(routed, "app.example", "/api/me"); answer != http.StatusTeapot {
		t.Errorf("app host = %d", answer)
	}
	if s.appPreview() != nil {
		t.Error("the app still serves previews beside itself")
	}
}

func serve(handler http.Handler, host, path string) int {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, path, nil)
	request.Host = host
	handler.ServeHTTP(recorder, request)
	return recorder.Code
}
