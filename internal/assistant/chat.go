package assistant

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/quiqr/chatcms-app/internal/aigateway"
	"github.com/quiqr/chatcms-app/internal/workspace"
)

var (
	ErrUnavailable = errors.New("assistant: there is no assistant on this installation")
	ErrNoAgent     = errors.New("assistant: the workspace's assistant did not answer")
)

type Agents interface {
	Agent(ctx context.Context, key workspace.Key, under string) (workspace.Endpoint, string, error)
}

type Event struct {
	Type    string  `json:"type"`
	Text    string  `json:"text,omitempty"`
	Tool    string  `json:"tool,omitempty"`
	Change  *Change `json:"change,omitempty"`
	Message string  `json:"message,omitempty"`
}

type Chat struct {
	agents Agents
	client *http.Client

	mu        sync.Mutex
	sessions  map[workspace.Key]string
	listeners map[workspace.Key]map[chan Event]bool
}

func NewChat(agents Agents) *Chat {
	return &Chat{
		agents:    agents,
		client:    &http.Client{Timeout: 30 * time.Second},
		sessions:  map[workspace.Key]string{},
		listeners: map[workspace.Key]map[chan Event]bool{},
	}
}

func (c *Chat) Available() bool {
	return c != nil && c.agents != nil
}

func (c *Chat) agent(ctx context.Context, key workspace.Key, under string) (string, string, error) {
	if !c.Available() {
		return "", "", ErrUnavailable
	}
	endpoint, secret, err := c.agents.Agent(ctx, key, under)
	if err != nil {
		return "", "", err
	}
	if endpoint.AgentPort == 0 || secret == "" {
		return "", "", ErrUnavailable
	}
	return "http://" + workspace.Endpoint{Host: endpoint.Host, Port: endpoint.AgentPort}.Address(), secret, nil
}

func (c *Chat) call(ctx context.Context, method, url, secret string, body any) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(data)
	}
	request, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, err
	}
	request.SetBasicAuth("opencode", secret)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	return c.client.Do(request)
}

func (c *Chat) session(ctx context.Context, key workspace.Key, base, secret string) (string, error) {
	c.mu.Lock()
	known, found := c.sessions[key]
	c.mu.Unlock()
	if found {
		return known, nil
	}

	var answer *http.Response
	var err error
	deadline := time.Now().Add(3 * time.Minute)
	for {
		answer, err = c.call(ctx, http.MethodPost, base+"/api/session", secret, map[string]any{"title": "ChatCMS"})
		starting := err == nil && answer.StatusCode == http.StatusServiceUnavailable
		if (err == nil && !starting) || time.Now().After(deadline) || ctx.Err() != nil {
			break
		}
		if starting {
			answer.Body.Close()
		}
		time.Sleep(250 * time.Millisecond)
	}
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrNoAgent, err)
	}
	defer answer.Body.Close()
	if answer.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: it answered %d", ErrNoAgent, answer.StatusCode)
	}
	var made struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(answer.Body).Decode(&made); err != nil || made.Data.ID == "" {
		return "", fmt.Errorf("%w: no session", ErrNoAgent)
	}
	c.awaitTools(ctx, base, secret)

	c.mu.Lock()
	defer c.mu.Unlock()
	if known, found := c.sessions[key]; found {
		return known, nil
	}
	c.sessions[key] = made.Data.ID
	return made.Data.ID, nil
}

const toolsWithin = 15 * time.Second

func (c *Chat) awaitTools(ctx context.Context, base, secret string) {
	deadline := time.Now().Add(toolsWithin)
	for time.Now().Before(deadline) && ctx.Err() == nil {
		if c.toolsConnected(ctx, base, secret) {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func (c *Chat) toolsConnected(ctx context.Context, base, secret string) bool {
	answer, err := c.call(ctx, http.MethodGet, base+"/api/mcp", secret, nil)
	if err != nil {
		return false
	}
	defer answer.Body.Close()
	var listed struct {
		Data []struct {
			Name   string `json:"name"`
			Status struct {
				Status string `json:"status"`
			} `json:"status"`
		} `json:"data"`
	}
	if answer.StatusCode != http.StatusOK || json.NewDecoder(answer.Body).Decode(&listed) != nil {
		return false
	}
	for _, server := range listed.Data {
		if server.Name == "quiqr" && server.Status.Status == "connected" {
			return true
		}
	}
	return false
}

func (c *Chat) Send(ctx context.Context, key workspace.Key, under, text string) error {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	base, secret, err := c.agent(ctx, key, under)
	if err != nil {
		return err
	}
	session, err := c.session(ctx, key, base, secret)
	if err != nil {
		return err
	}
	answer, err := c.call(ctx, http.MethodPost, base+"/api/session/"+session+"/prompt", secret, map[string]any{"text": text})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrNoAgent, err)
	}
	answer.Body.Close()
	if answer.StatusCode >= 300 {
		c.forget(key)
		return fmt.Errorf("%w: it answered %d", ErrNoAgent, answer.StatusCode)
	}
	return nil
}

func (c *Chat) Stop(ctx context.Context, key workspace.Key, under string) error {
	c.mu.Lock()
	session, known := c.sessions[key]
	c.mu.Unlock()
	if !known {
		return nil
	}
	base, secret, err := c.agent(ctx, key, under)
	if err != nil {
		return err
	}
	answer, err := c.call(ctx, http.MethodPost, base+"/api/session/"+session+"/interrupt", secret, nil)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrNoAgent, err)
	}
	answer.Body.Close()
	if answer.StatusCode >= 300 {
		return fmt.Errorf("%w: it answered %d", ErrNoAgent, answer.StatusCode)
	}
	return nil
}

func (c *Chat) forget(key workspace.Key) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.sessions, key)
}

func (c *Chat) Announce(caller aigateway.Caller, change Change) {
	key := workspace.Key{Person: caller.Person, Project: caller.Project}
	c.mu.Lock()
	defer c.mu.Unlock()
	for listener := range c.listeners[key] {
		select {
		case listener <- Event{Type: "change", Change: &change}:
		default:
		}
	}
}

func (c *Chat) listen(key workspace.Key) chan Event {
	listener := make(chan Event, 64)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.listeners[key] == nil {
		c.listeners[key] = map[chan Event]bool{}
	}
	c.listeners[key][listener] = true
	return listener
}

func (c *Chat) unlisten(key workspace.Key, listener chan Event) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.listeners[key], listener)
	if len(c.listeners[key]) == 0 {
		delete(c.listeners, key)
	}
}

type agentEvent struct {
	Type string `json:"type"`
	Data struct {
		SessionID string `json:"sessionID"`
		ID        string `json:"id"`
		Name      string `json:"name"`
		Delta     string `json:"delta"`
		Attempt   int    `json:"attempt"`
		Error     *struct {
			Type    string `json:"type"`
			Message string `json:"message"`
			Status  int    `json:"status"`
		} `json:"error"`
	} `json:"data"`
}

const retriesBeforeGivingUp = 3

func (c *Chat) interrupt(ctx context.Context, base, secret, session string) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if answer, err := c.call(ctx, http.MethodPost, base+"/api/session/"+session+"/interrupt", secret, nil); err == nil {
		answer.Body.Close()
	}
}

func (c *Chat) Events(ctx context.Context, key workspace.Key, under string, emit func(Event) error) error {
	if !c.Available() {
		return ErrUnavailable
	}
	changes := c.listen(key)
	defer c.unlisten(key, changes)

	base, secret, err := c.agent(ctx, key, under)
	if err != nil {
		return err
	}
	session, err := c.session(ctx, key, base, secret)
	if err != nil {
		return err
	}

	streamCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	request, err := http.NewRequestWithContext(streamCtx, http.MethodGet, base+"/api/event", nil)
	if err != nil {
		return err
	}
	request.SetBasicAuth("opencode", secret)
	request.Header.Set("Accept", "text/event-stream")
	answer, err := (&http.Client{}).Do(request)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrNoAgent, err)
	}
	defer answer.Body.Close()
	if answer.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: it answered %d", ErrNoAgent, answer.StatusCode)
	}

	from := make(chan Event, 64)
	go func() {
		defer close(from)
		relay(answer.Body, session, func(event Event) {
			select {
			case from <- event:
			case <-streamCtx.Done():
			}
		}, func() {
			go c.interrupt(context.WithoutCancel(ctx), base, secret, session)
		})
	}()

	if err := emit(Event{Type: "ready"}); err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case change := <-changes:
			if err := emit(change); err != nil {
				return err
			}
		case event, open := <-from:
			if !open {
				return nil
			}
			if err := emit(event); err != nil {
				return err
			}
		}
	}
}

func relay(stream io.Reader, session string, emit func(Event), halt func()) {
	tools := map[string]string{}
	halted := false

	reader := bufio.NewReaderSize(stream, 256<<10)
	for {
		line, err := reader.ReadBytes('\n')
		if payload, found := bytes.CutPrefix(bytes.TrimSpace(line), []byte("data:")); found {
			var event agentEvent
			if json.Unmarshal(bytes.TrimSpace(payload), &event) == nil && event.Data.SessionID == session {
				if gaveUp(event) {
					if !halted {
						halted = true
						emit(Event{Type: "error", Message: messageOf(event, "the assistant could not answer")})
						halt()
					}
				} else {
					translate(event, tools, emit)
				}
				if strings.HasPrefix(event.Type, "session.execution.") {
					halted = false
				}
			}
		}
		if err != nil {
			return
		}
	}
}

func messageOf(event agentEvent, otherwise string) string {
	if event.Data.Error != nil && event.Data.Error.Message != "" {
		return event.Data.Error.Message
	}
	return otherwise
}

func gaveUp(event agentEvent) bool {
	if event.Type != "session.retry.scheduled" {
		return false
	}
	refused := event.Data.Error != nil && event.Data.Error.Status >= 400 && event.Data.Error.Status < 500
	return refused || event.Data.Attempt >= retriesBeforeGivingUp
}

func translate(event agentEvent, tools map[string]string, emit func(Event)) {
	data := event.Data
	switch event.Type {
	case "session.text.delta":
		if data.Delta != "" {
			emit(Event{Type: "text", Text: data.Delta})
		}
	case "session.tool.input.started":
		tools[data.ID] = strings.TrimPrefix(data.Name, "quiqr_")
		emit(Event{Type: "tool", Tool: tools[data.ID], Message: "running"})
	case "session.tool.failed":
		emit(Event{Type: "tool", Tool: tools[data.ID], Message: "error"})
	case "session.execution.succeeded", "session.execution.interrupted":
		emit(Event{Type: "done"})
	case "session.execution.failed":
		emit(Event{Type: "error", Message: messageOf(event, "the assistant could not answer")})
	}
}
