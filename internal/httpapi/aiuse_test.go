package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"testing"
)

type stubAIUse struct {
	view     AIUseView
	err      error
	requests *int
	tokens   *int64
}

func (s *stubAIUse) AIUse(context.Context, Actor, string) (AIUseView, error) {
	return s.view, s.err
}

func (s *stubAIUse) SetAILimits(_ context.Context, _ Actor, _ string, requests *int, tokens *int64) (AIUseView, error) {
	s.requests, s.tokens = requests, tokens
	return s.view, s.err
}

func withAIUse(service AIUseService, signedIn bool) http.Handler {
	known := map[string]Actor{}
	if signedIn {
		known["a-session-value"] = Actor{ID: "person-1", Email: "boss@example.org"}
	}
	return New(Options{
		Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Health: func(context.Context) Health { return Health{Database: DatabaseReachable} },
		Auth:   &stubAuth{known: known},
		AIUse:  service,
	})
}

const aiPath = "/api/administration/organisations/bakery/ai"

func TestAIUseNeedsAPerson(t *testing.T) {
	handler := withAIUse(&stubAIUse{}, false)
	for _, method := range []string{http.MethodGet, http.MethodPut} {
		if got := call(t, handler, method, aiPath, map[string]any{}); got.Code != http.StatusUnauthorized {
			t.Errorf("%s = %d, want 401", method, got.Code)
		}
	}
}

func TestAIUseIsReadAndSet(t *testing.T) {
	rate := 30
	service := &stubAIUse{view: AIUseView{RequestsPerMinute: &rate, RequestsThisMonth: 4, TokensThisMonth: 1200}}
	handler := withAIUse(service, true)

	got := signedInCall(t, handler, http.MethodGet, aiPath, nil)
	var view AIUseView
	json.NewDecoder(got.Body).Decode(&view)
	if got.Code != http.StatusOK || *view.RequestsPerMinute != 30 || view.TokensThisMonth != 1200 || view.TokensPerMonth != nil {
		t.Errorf("GET = %d %+v", got.Code, view)
	}

	put := signedInCall(t, handler, http.MethodPut, aiPath, map[string]any{"requestsPerMinute": 10, "tokensPerMonth": nil})
	if put.Code != http.StatusOK || service.requests == nil || *service.requests != 10 || service.tokens != nil {
		t.Errorf("PUT = %d, passed %v %v", put.Code, service.requests, service.tokens)
	}

	if got := signedInCall(t, handler, http.MethodDelete, aiPath, nil); got.Code != http.StatusMethodNotAllowed {
		t.Errorf("DELETE = %d", got.Code)
	}
}

func TestAIUseRefusalsArePassedOn(t *testing.T) {
	handler := withAIUse(&stubAIUse{err: ErrNotAllowed}, true)
	if got := signedInCall(t, handler, http.MethodGet, aiPath, nil); got.Code != http.StatusForbidden {
		t.Errorf("GET not allowed = %d", got.Code)
	}
	if got := signedInCall(t, handler, http.MethodPut, aiPath, map[string]any{}); got.Code != http.StatusForbidden {
		t.Errorf("PUT not allowed = %d", got.Code)
	}
}
