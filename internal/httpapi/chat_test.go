package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type stubChat struct {
	view   ChatView
	err    error
	asked  []string
	events []any
}

func (s *stubChat) Chat(context.Context, Actor, string, string) (ChatView, error) {
	return s.view, s.err
}

func (s *stubChat) Ask(_ context.Context, _ Actor, _, _ string, text string) error {
	s.asked = append(s.asked, text)
	return s.err
}

func (s *stubChat) Listen(_ context.Context, _ Actor, _, _ string, emit func(any) error) error {
	if s.err != nil {
		return s.err
	}
	for _, event := range s.events {
		if err := emit(event); err != nil {
			return err
		}
	}
	return nil
}

func withChat(service ChatService, signedIn bool) http.Handler {
	known := map[string]Actor{}
	if signedIn {
		known["a-session-value"] = Actor{ID: "person-1", Email: "anna@example.org"}
	}
	return New(Options{
		Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Health: func(context.Context) Health { return Health{Database: DatabaseReachable} },
		Auth:   &stubAuth{known: known},
		Chat:   service,
	})
}

const chatPath = "/api/organisations/technative/projects/korenaar/chat"

func TestChatNeedsAPerson(t *testing.T) {
	handler := withChat(&stubChat{}, false)
	for _, pair := range [][2]string{{http.MethodGet, chatPath}, {http.MethodPost, chatPath}, {http.MethodGet, chatPath + "/events"}} {
		if got := call(t, handler, pair[0], pair[1], map[string]any{}); got.Code != http.StatusUnauthorized {
			t.Errorf("%s %s = %d", pair[0], pair[1], got.Code)
		}
	}
}

func TestAskingAndListening(t *testing.T) {
	service := &stubChat{view: ChatView{Available: true}, events: []any{map[string]string{"type": "ready"}, map[string]string{"type": "text", "text": "Hi"}}}
	handler := withChat(service, true)

	if got := signedInCall(t, handler, http.MethodGet, chatPath, nil); got.Code != http.StatusOK || !strings.Contains(got.Body.String(), `"available":true`) {
		t.Errorf("GET = %d %s", got.Code, got.Body)
	}
	if got := signedInCall(t, handler, http.MethodPost, chatPath, map[string]any{"text": "change the title"}); got.Code != http.StatusAccepted || service.asked[0] != "change the title" {
		t.Errorf("POST = %d, asked %v", got.Code, service.asked)
	}

	request := httptest.NewRequest(http.MethodGet, chatPath+"/events", nil)
	request.AddCookie(&http.Cookie{Name: SessionCookie, Value: "a-session-value"})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Header().Get("Content-Type") != "text/event-stream" || !strings.Contains(recorder.Body.String(), "data: {\"text\":\"Hi\",\"type\":\"text\"}") {
		t.Errorf("events = %q %q", recorder.Header().Get("Content-Type"), recorder.Body)
	}
}

func TestAChatRefusalIsPassedOn(t *testing.T) {
	handler := withChat(&stubChat{err: &RefusedError{Key: "refused.assistant-not-available", Reason: "no"}}, true)
	if got := signedInCall(t, handler, http.MethodPost, chatPath, map[string]any{"text": "hi"}); got.Code < 400 {
		t.Errorf("POST refused = %d", got.Code)
	}
	if got := signedInCall(t, handler, http.MethodGet, chatPath+"/events", nil); got.Code < 400 {
		t.Errorf("events refused = %d", got.Code)
	}
	if got := signedInCall(t, handler, http.MethodDelete, chatPath, nil); got.Code != http.StatusMethodNotAllowed {
		t.Errorf("DELETE = %d", got.Code)
	}
}
