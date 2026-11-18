package assistant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/aigateway"
	"github.com/quiqr/chatcms-app/internal/workspace"
)

type fakeOpenCode struct {
	mu       sync.Mutex
	sessions int
	prompts  []string
	auth     []string
	events   chan string
	refuse   bool
}

func (f *fakeOpenCode) server(t *testing.T) *httptest.Server {
	t.Helper()
	f.events = make(chan string, 64)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, password, _ := r.BasicAuth()
		f.mu.Lock()
		f.auth = append(f.auth, user+":"+password)
		f.mu.Unlock()
		if password != "secret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/session":
			f.mu.Lock()
			f.sessions++
			f.mu.Unlock()
			if f.refuse {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			fmt.Fprint(w, `{"id":"ses_one"}`)
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/prompt_async"):
			var body struct {
				Parts []struct{ Text string } `json:"parts"`
			}
			json.NewDecoder(r.Body).Decode(&body)
			f.mu.Lock()
			f.prompts = append(f.prompts, body.Parts[0].Text)
			f.mu.Unlock()
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodGet && r.URL.Path == "/event":
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)
			w.(http.Flusher).Flush()
			for {
				select {
				case <-r.Context().Done():
					return
				case line := <-f.events:
					fmt.Fprintf(w, "data: %s\n\n", line)
					w.(http.Flusher).Flush()
				}
			}
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

type agents struct {
	endpoint workspace.Endpoint
	secret   string
	err      error
}

func (a agents) Agent(ctx context.Context, key workspace.Key, under string) (workspace.Endpoint, string, error) {
	return a.endpoint, a.secret, a.err
}

func agentsFor(t *testing.T, server *httptest.Server) agents {
	t.Helper()
	host, port, _ := net.SplitHostPort(strings.TrimPrefix(server.URL, "http://"))
	number, _ := strconv.Atoi(port)
	return agents{endpoint: workspace.Endpoint{Host: host, Port: 1, AgentPort: number}, secret: "secret"}
}

func aKey() workspace.Key { return workspace.Key{Person: uuid.New(), Project: uuid.New()} }

func TestASentMessageGoesToOneSessionWithTheSecret(t *testing.T) {
	fake := &fakeOpenCode{}
	chat := NewChat(agentsFor(t, fake.server(t)))
	key := aKey()

	for _, text := range []string{"first", "second", "   "} {
		if err := chat.Send(t.Context(), key, "u", text); err != nil {
			t.Fatalf("Send(%q) error = %v", text, err)
		}
	}
	if fake.sessions != 1 || len(fake.prompts) != 2 || fake.prompts[1] != "second" {
		t.Errorf("sessions %d, prompts %v, want one session and two prompts", fake.sessions, fake.prompts)
	}
	for _, auth := range fake.auth {
		if auth != "opencode:secret" {
			t.Errorf("OpenCode was asked with %q", auth)
		}
	}
}

func TestWithoutAnAssistantNothingIsTried(t *testing.T) {
	var none *Chat
	if none.Available() {
		t.Errorf("a missing chat is available")
	}
	empty := NewChat(nil)
	if err := empty.Send(t.Context(), aKey(), "u", "hi"); !errors.Is(err, ErrUnavailable) {
		t.Errorf("Send() error = %v, want ErrUnavailable", err)
	}
	if err := empty.Events(t.Context(), aKey(), "u", func(Event) error { return nil }); !errors.Is(err, ErrUnavailable) {
		t.Errorf("Events() error = %v, want ErrUnavailable", err)
	}
	noAgent := NewChat(agents{endpoint: workspace.Endpoint{Host: "127.0.0.1", Port: 1}, secret: "s"})
	if err := noAgent.Send(t.Context(), aKey(), "u", "hi"); !errors.Is(err, ErrUnavailable) {
		t.Errorf("Send() to a workspace without an agent error = %v", err)
	}
	failing := NewChat(agents{err: errors.New("no workspace")})
	if err := failing.Send(t.Context(), aKey(), "u", "hi"); err == nil || err.Error() != "no workspace" {
		t.Errorf("Send() when the workspace fails = %v", err)
	}
}

func TestAnAgentThatWillNotMakeASessionIsSaid(t *testing.T) {
	fake := &fakeOpenCode{refuse: true}
	chat := NewChat(agentsFor(t, fake.server(t)))
	if err := chat.Send(t.Context(), aKey(), "u", "hi"); !errors.Is(err, ErrNoAgent) {
		t.Errorf("Send() error = %v, want ErrNoAgent", err)
	}
}

const (
	userMessage      = `{"type":"message.updated","properties":{"info":{"id":"msg_user","role":"user","sessionID":"ses_one"}}}`
	userText         = `{"type":"message.part.updated","properties":{"sessionID":"ses_one","part":{"id":"prt_user","messageID":"msg_user","type":"text","text":"hi"}}}`
	userDelta        = `{"type":"message.part.delta","properties":{"sessionID":"ses_one","messageID":"msg_user","partID":"prt_user","field":"text","delta":"hi"}}`
	assistantMessage = `{"type":"message.updated","properties":{"info":{"id":"msg_a","role":"assistant","sessionID":"ses_one"}}}`
	reasoning        = `{"type":"message.part.updated","properties":{"sessionID":"ses_one","part":{"id":"prt_think","messageID":"msg_a","type":"reasoning"}}}`
	reasoningDelta   = `{"type":"message.part.delta","properties":{"sessionID":"ses_one","messageID":"msg_a","partID":"prt_think","field":"text","delta":"hmm"}}`
	tool             = `{"type":"message.part.updated","properties":{"sessionID":"ses_one","part":{"id":"prt_tool","messageID":"msg_a","type":"tool","tool":"quiqr_update_page","state":{"status":"running"}}}}`
	toolAgain        = `{"type":"message.part.updated","properties":{"sessionID":"ses_one","part":{"id":"prt_tool","messageID":"msg_a","type":"tool","tool":"quiqr_update_page","state":{"status":"running"}}}}`
	text             = `{"type":"message.part.updated","properties":{"sessionID":"ses_one","part":{"id":"prt_text","messageID":"msg_a","type":"text","text":""}}}`
	textOne          = `{"type":"message.part.delta","properties":{"sessionID":"ses_one","messageID":"msg_a","partID":"prt_text","field":"text","delta":"Hel"}}`
	textTwo          = `{"type":"message.part.delta","properties":{"sessionID":"ses_one","messageID":"msg_a","partID":"prt_text","field":"text","delta":"lo"}}`
	elsewhere        = `{"type":"message.part.delta","properties":{"sessionID":"ses_other","messageID":"msg_x","partID":"prt_text","field":"text","delta":"nope"}}`
	failed           = `{"type":"session.error","properties":{"sessionID":"ses_one","error":{"name":"APIError","data":{"message":"this organisation has used its AI budget"}}}}`
	idle             = `{"type":"session.idle","properties":{"sessionID":"ses_one"}}`
)

func TestTheReplyStreamsAndChangesArrive(t *testing.T) {
	fake := &fakeOpenCode{}
	chat := NewChat(agentsFor(t, fake.server(t)))
	key := aKey()

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	got := make(chan Event, 64)
	go chat.Events(ctx, key, "u", func(event Event) error {
		got <- event
		return nil
	})

	if first := <-got; first.Type != "ready" {
		t.Fatalf("first event = %+v, want ready", first)
	}
	for _, line := range []string{userMessage, userText, userDelta, assistantMessage, reasoning, reasoningDelta, tool, toolAgain, text, textOne, elsewhere, textTwo, "not json", failed, idle} {
		fake.events <- line
	}
	chat.Announce(aigateway.Caller{Person: key.Person, Project: key.Project}, Change{Page: "home", Title: "Home page", Before: "a", After: "b"})
	chat.Announce(aigateway.Caller{Person: uuid.New(), Project: key.Project}, Change{Page: "other"})

	var seen []Event
	deadline := time.After(5 * time.Second)
	for len(seen) < 6 {
		select {
		case event := <-got:
			seen = append(seen, event)
		case <-deadline:
			t.Fatalf("saw only %+v", seen)
		}
	}

	var text strings.Builder
	var kinds []string
	for _, event := range seen {
		kinds = append(kinds, event.Type)
		if event.Type == "text" {
			text.WriteString(event.Text)
		}
		if event.Type == "change" && event.Change.Page != "home" {
			t.Errorf("another person's change arrived: %+v", event.Change)
		}
		if event.Type == "tool" && event.Tool != "update_page" {
			t.Errorf("tool = %q", event.Tool)
		}
		if event.Type == "error" && !strings.Contains(event.Message, "budget") {
			t.Errorf("error = %q", event.Message)
		}
	}
	if text.String() != "Hello" {
		t.Errorf("the reply read %q, want only the assistant's text", text.String())
	}
	joined := strings.Join(kinds, ",")
	for _, want := range []string{"tool", "change", "error", "done"} {
		if !strings.Contains(joined, want) {
			t.Errorf("events %s, want %s among them", joined, want)
		}
	}
	if strings.Count(joined, "tool") != 1 {
		t.Errorf("a running tool was reported more than once: %s", joined)
	}
}

func TestEventsStopWhenThePersonLeaves(t *testing.T) {
	fake := &fakeOpenCode{}
	chat := NewChat(agentsFor(t, fake.server(t)))
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() {
		done <- chat.Events(ctx, aKey(), "u", func(Event) error { return nil })
	}()
	time.Sleep(100 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Events() error = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("Events() did not stop")
	}
	chat.mu.Lock()
	defer chat.mu.Unlock()
	if len(chat.listeners) != 0 {
		t.Errorf("a listener was left behind")
	}
}

func TestAnEmitThatFailsEndsTheStream(t *testing.T) {
	fake := &fakeOpenCode{}
	chat := NewChat(agentsFor(t, fake.server(t)))
	gone := errors.New("browser gone")
	if err := chat.Events(t.Context(), aKey(), "u", func(Event) error { return gone }); !errors.Is(err, gone) {
		t.Errorf("Events() error = %v", err)
	}
}
