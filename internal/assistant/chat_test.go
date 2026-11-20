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
	aborts   int
	looks    int
	connects int
	starting int
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
		case r.Method == http.MethodPost && r.URL.Path == "/api/session":
			f.mu.Lock()
			f.sessions++
			still := f.sessions <= f.starting
			f.mu.Unlock()
			if still {
				w.WriteHeader(http.StatusServiceUnavailable)
				fmt.Fprint(w, `{"code":"service_starting"}`)
				return
			}
			if f.refuse {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			fmt.Fprint(w, `{"data":{"id":"ses_one","title":"ChatCMS"}}`)
		case r.Method == http.MethodGet && r.URL.Path == "/api/mcp":
			f.mu.Lock()
			f.looks++
			connected := f.looks > f.connects
			f.mu.Unlock()
			if connected {
				fmt.Fprint(w, `{"data":[{"name":"quiqr","status":{"status":"connected"}}]}`)
			} else {
				fmt.Fprint(w, `{"data":[]}`)
			}
		case r.Method == http.MethodPost && r.URL.Path == "/api/session/ses_one/prompt":
			var body struct {
				Text string `json:"text"`
			}
			json.NewDecoder(r.Body).Decode(&body)
			f.mu.Lock()
			f.prompts = append(f.prompts, body.Text)
			f.mu.Unlock()
			fmt.Fprint(w, `{"data":{"id":"msg_1","type":"user"}}`)
		case r.Method == http.MethodPost && r.URL.Path == "/api/session/ses_one/interrupt":
			f.mu.Lock()
			f.aborts++
			f.mu.Unlock()
			fmt.Fprint(w, `{"interrupted":true}`)
		case r.Method == http.MethodGet && r.URL.Path == "/api/event":
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
	connected   = `{"id":"evt_0","type":"server.connected","data":{}}`
	enqueued    = `{"type":"session.inbox.enqueued","data":{"sessionID":"ses_one","item":{"type":"user","payload":{"text":"hi"}}}}`
	started     = `{"type":"session.execution.started","data":{"sessionID":"ses_one"}}`
	toolStarted = `{"type":"session.tool.input.started","data":{"sessionID":"ses_one","assistantMessageID":"msg_a","id":"call_1","name":"quiqr_update_page"}}`
	toolCalled  = `{"type":"session.tool.called","data":{"sessionID":"ses_one","id":"call_1","input":{}}}`
	toolDone    = `{"type":"session.tool.success","data":{"sessionID":"ses_one","id":"call_1","content":[]}}`
	textOne     = `{"type":"session.text.delta","data":{"sessionID":"ses_one","assistantMessageID":"msg_a","ordinal":0,"delta":"Hel"}}`
	textTwo     = `{"type":"session.text.delta","data":{"sessionID":"ses_one","assistantMessageID":"msg_a","ordinal":0,"delta":"lo"}}`
	textEnded   = `{"type":"session.text.ended","data":{"sessionID":"ses_one","text":"Hello"}}`
	elsewhere   = `{"type":"session.text.delta","data":{"sessionID":"ses_other","delta":"nope"}}`
	failed      = `{"type":"session.execution.failed","data":{"sessionID":"ses_one","error":{"type":"provider","message":"this organisation has used its AI budget"}}}`
	succeeded   = `{"type":"session.execution.succeeded","data":{"sessionID":"ses_one"}}`
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
	for _, line := range []string{connected, enqueued, started, toolStarted, toolCalled, toolDone, textOne, elsewhere, textTwo, textEnded, "not json", failed, succeeded} {
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

func TestStoppingAbortsTheSessionAndNothingWithoutOne(t *testing.T) {
	fake := &fakeOpenCode{}
	chat := NewChat(agentsFor(t, fake.server(t)))
	key := aKey()
	if err := chat.Stop(t.Context(), key, "u"); err != nil {
		t.Fatalf("Stop() without a session error = %v", err)
	}
	if err := chat.Send(t.Context(), key, "u", "change it"); err != nil {
		t.Fatal(err)
	}
	if err := chat.Stop(t.Context(), key, "u"); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	fake.mu.Lock()
	defer fake.mu.Unlock()
	if fake.aborts != 1 {
		t.Fatalf("aborts = %d, want 1", fake.aborts)
	}
}

func eventsOf(t *testing.T, chat *Chat, key workspace.Key) chan Event {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)
	got := make(chan Event, 64)
	go chat.Events(ctx, key, "u", func(event Event) error {
		got <- event
		return nil
	})
	if first := <-got; first.Type != "ready" {
		t.Fatalf("first event = %+v, want ready", first)
	}
	return got
}

func waitFor(t *testing.T, got chan Event, kind string) Event {
	t.Helper()
	deadline := time.After(5 * time.Second)
	for {
		select {
		case event := <-got:
			if event.Type == kind {
				return event
			}
		case <-deadline:
			t.Fatalf("no %s event arrived", kind)
		}
	}
}

func TestARefusalEndsTheAnswerAndInterruptsTheSession(t *testing.T) {
	fake := &fakeOpenCode{}
	chat := NewChat(agentsFor(t, fake.server(t)))
	got := eventsOf(t, chat, aKey())

	fake.events <- `{"type":"session.retry.scheduled","data":{"sessionID":"ses_one","attempt":2,"error":{"type":"provider.rate","message":"Your organisation has used its AI allowance for now","status":429}}}`
	if said := waitFor(t, got, "error"); !strings.Contains(said.Message, "allowance") {
		t.Errorf("error = %q", said.Message)
	}
	fake.events <- `{"type":"session.execution.interrupted","data":{"sessionID":"ses_one","reason":"user"}}`
	waitFor(t, got, "done")
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		fake.mu.Lock()
		aborts := fake.aborts
		fake.mu.Unlock()
		if aborts == 1 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Errorf("the session was not interrupted")
}

func TestAFailingModelIsRetriedTwiceAndThenGivenUp(t *testing.T) {
	fake := &fakeOpenCode{}
	chat := NewChat(agentsFor(t, fake.server(t)))
	got := eventsOf(t, chat, aKey())

	retry := func(attempt int) string {
		return fmt.Sprintf(`{"type":"session.retry.scheduled","data":{"sessionID":"ses_one","attempt":%d,"error":{"type":"provider.internal","message":"the model broke","status":500}}}`, attempt)
	}
	fake.events <- retry(2)
	fake.events <- textOne
	if first := waitFor(t, got, "text"); first.Text != "Hel" {
		t.Fatalf("text = %+v", first)
	}
	fake.events <- retry(3)
	fake.events <- retry(4)
	if said := waitFor(t, got, "error"); said.Message != "the model broke" {
		t.Errorf("error = %q", said.Message)
	}
	time.Sleep(200 * time.Millisecond)
	fake.events <- succeeded
	deadline := time.After(5 * time.Second)
	for done := false; !done; {
		select {
		case event := <-got:
			if event.Type == "error" {
				t.Errorf("a second error arrived: %+v", event)
			}
			done = event.Type == "done"
		case <-deadline:
			t.Fatal("no done event arrived")
		}
	}
	fake.mu.Lock()
	defer fake.mu.Unlock()
	if fake.aborts != 1 {
		t.Errorf("interrupts = %d, want 1", fake.aborts)
	}
}

func TestTheFirstMessageWaitsForThePageTools(t *testing.T) {
	fake := &fakeOpenCode{connects: 3}
	chat := NewChat(agentsFor(t, fake.server(t)))
	if err := chat.Send(t.Context(), aKey(), "u", "change it"); err != nil {
		t.Fatal(err)
	}
	fake.mu.Lock()
	defer fake.mu.Unlock()
	if fake.looks != 4 || len(fake.prompts) != 1 {
		t.Errorf("looked %d times, prompts %v; want the prompt after the tools connected", fake.looks, fake.prompts)
	}
}

func TestAnAgentThatIsStillStartingIsWaitedFor(t *testing.T) {
	fake := &fakeOpenCode{starting: 2}
	chat := NewChat(agentsFor(t, fake.server(t)))
	if err := chat.Send(t.Context(), aKey(), "u", "hi"); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	fake.mu.Lock()
	defer fake.mu.Unlock()
	if fake.sessions != 3 || len(fake.prompts) != 1 {
		t.Errorf("sessions asked %d, prompts %v; want the third to succeed", fake.sessions, fake.prompts)
	}
}
