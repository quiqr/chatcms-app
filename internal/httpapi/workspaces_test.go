package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
)

type stubStatistics struct{ asked int }

func (s *stubStatistics) WorkspaceStatistics(context.Context) (any, error) {
	s.asked++
	return map[string]any{"limit": 8, "now": map[string]int{"warm": 2}}, nil
}

func withStatistics(administers, signedIn bool, stats WorkspaceStatistics) http.Handler {
	known := map[string]Actor{}
	if signedIn {
		known["a-session-value"] = Actor{ID: "person-1"}
	}
	return New(Options{
		Logger:         slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Auth:           &stubAuth{known: known},
		Administration: &stubAdministration{administers: administers},
		Workspaces:     stats,
	})
}

func TestAPlatformAdminReadsWorkspaceStatistics(t *testing.T) {
	stats := &stubStatistics{}
	recorder := signedInCall(t, withStatistics(true, true, stats), http.MethodGet, "/api/administration/workspaces", nil)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"limit":8`) {
		t.Fatalf("status = %d: %s", recorder.Code, recorder.Body)
	}
}

func TestNobodyElseReadsWorkspaceStatistics(t *testing.T) {
	stats := &stubStatistics{}
	if got := signedInCall(t, withStatistics(false, true, stats), http.MethodGet, "/api/administration/workspaces", nil); got.Code != http.StatusForbidden {
		t.Errorf("somebody who is not a platform admin = %d, want 403", got.Code)
	}
	if got := call(t, withStatistics(true, false, stats), http.MethodGet, "/api/administration/workspaces", nil); got.Code != http.StatusUnauthorized {
		t.Errorf("nobody signed in = %d, want 401", got.Code)
	}
	if got := signedInCall(t, withStatistics(true, true, stats), http.MethodPost, "/api/administration/workspaces", nil); got.Code != http.StatusMethodNotAllowed {
		t.Errorf("a POST = %d, want 405", got.Code)
	}
	if stats.asked != 0 {
		t.Error("the statistics were read for somebody who may not see them")
	}
}
