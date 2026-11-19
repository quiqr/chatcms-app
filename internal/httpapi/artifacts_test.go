package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"
)

type stubArtifacts struct {
	views []ArtifactView
	err   error
	asked string
}

type readSeekCloser struct{ *strings.Reader }

func (readSeekCloser) Close() error { return nil }

func (s *stubArtifacts) Artifacts(context.Context, Actor, string, string) ([]ArtifactView, error) {
	return s.views, s.err
}

func (s *stubArtifacts) OpenArtifact(_ context.Context, _ Actor, _, _ string, id string) (ArtifactView, io.ReadSeekCloser, time.Time, error) {
	s.asked = id
	if s.err != nil {
		return ArtifactView{}, nil, time.Time{}, s.err
	}
	return ArtifactView{ID: id, Name: "downloads/Korenaar Recipes.pdf", Size: 5}, readSeekCloser{strings.NewReader("%PDF-")}, time.Date(2026, 11, 19, 9, 0, 0, 0, time.UTC), nil
}

func withArtifacts(service ArtifactService, signedIn bool) http.Handler {
	known := map[string]Actor{}
	if signedIn {
		known["a-session-value"] = Actor{ID: "person-1", Email: "anna@example.org"}
	}
	return New(Options{
		Logger:    slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Health:    func(context.Context) Health { return Health{Database: DatabaseReachable} },
		Auth:      &stubAuth{known: known},
		Artifacts: service,
	})
}

const artifactsPath = "/api/organisations/technative/projects/korenaar/artifacts"

func TestArtifactsNeedAPerson(t *testing.T) {
	handler := withArtifacts(&stubArtifacts{}, false)
	for _, path := range []string{artifactsPath, artifactsPath + "/a1"} {
		if got := call(t, handler, http.MethodGet, path, nil); got.Code != http.StatusUnauthorized {
			t.Errorf("GET %s = %d", path, got.Code)
		}
	}
}

func TestArtifactsAreListed(t *testing.T) {
	handler := withArtifacts(&stubArtifacts{views: []ArtifactView{{ID: "a1", Name: "Korenaar.pdf", Size: 5}}}, true)
	got := signedInCall(t, handler, http.MethodGet, artifactsPath, nil)
	if got.Code != http.StatusOK || !strings.Contains(got.Body.String(), `"name":"Korenaar.pdf"`) {
		t.Errorf("GET = %d %s", got.Code, got.Body)
	}
}

func TestAnArtifactDownloadsUnderItsName(t *testing.T) {
	service := &stubArtifacts{}
	got := signedInCall(t, withArtifacts(service, true), http.MethodGet, artifactsPath+"/a1", nil)
	if got.Code != http.StatusOK || got.Body.String() != "%PDF-" || service.asked != "a1" {
		t.Fatalf("GET = %d %q", got.Code, got.Body)
	}
	if got.Header().Get("Content-Type") != "application/pdf" || got.Header().Get("Content-Disposition") != `attachment; filename="Korenaar Recipes.pdf"` || got.Header().Get("Content-Security-Policy") != "sandbox" {
		t.Errorf("headers = %v", got.Header())
	}
}

func TestAnArtifactThatIsNotThereIsMissing(t *testing.T) {
	got := signedInCall(t, withArtifacts(&stubArtifacts{err: ErrMissing}, true), http.MethodGet, artifactsPath+"/nope", nil)
	if got.Code != http.StatusNotFound {
		t.Errorf("GET = %d", got.Code)
	}
	if got := signedInCall(t, withArtifacts(&stubArtifacts{}, true), http.MethodPost, artifactsPath, nil); got.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST = %d", got.Code)
	}
}
