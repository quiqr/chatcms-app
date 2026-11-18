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
		answer, err = c.call(ctx, http.MethodPost, base+"/session", secret, map[string]any{"title": "ChatCMS"})
		if err == nil || time.Now().After(deadline) || ctx.Err() != nil {
			break
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
		ID string `json:"id"`
	}
	if err := json.NewDecoder(answer.Body).Decode(&made); err != nil || made.ID == "" {
		return "", fmt.Errorf("%w: no session", ErrNoAgent)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if known, found := c.sessions[key]; found {
		return known, nil
	}
	c.sessions[key] = made.ID
	return made.ID, nil
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
	answer, err := c.call(ctx, http.MethodPost, base+"/session/"+session+"/prompt_async", secret, map[string]any{
		"parts": []map[string]string{{"type": "text", "text": text}},
	})
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
	Type       string `json:"type"`
	Properties struct {
		SessionID string `json:"sessionID"`
		MessageID string `json:"messageID"`
		PartID    string `json:"partID"`
		Field     string `json:"field"`
		Delta     string `json:"delta"`
		Info      struct {
			ID   string `json:"id"`
			Role string `json:"role"`
		} `json:"info"`
		Part struct {
			ID        string `json:"id"`
			MessageID string `json:"messageID"`
			Type      string `json:"type"`
			Tool      string `json:"tool"`
			State     struct {
				Status string `json:"status"`
			} `json:"state"`
		} `json:"part"`
		Error *struct {
			Name string `json:"name"`
			Data struct {
				Message string `json:"message"`
			} `json:"data"`
		} `json:"error"`
	} `json:"properties"`
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
	request, err := http.NewRequestWithContext(streamCtx, http.MethodGet, base+"/event", nil)
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

func relay(stream io.Reader, session string, emit func(Event)) {
	users := map[string]bool{}
	texts := map[string]bool{}
	tools := map[string]string{}

	reader := bufio.NewReaderSize(stream, 256<<10)
	for {
		line, err := reader.ReadBytes('\n')
		if payload, found := bytes.CutPrefix(bytes.TrimSpace(line), []byte("data:")); found {
			var event agentEvent
			if json.Unmarshal(bytes.TrimSpace(payload), &event) == nil {
				translate(event, session, users, texts, tools, emit)
			}
		}
		if err != nil {
			return
		}
	}
}

func translate(event agentEvent, session string, users, texts map[string]bool, tools map[string]string, emit func(Event)) {
	properties := event.Properties
	switch event.Type {
	case "message.updated":
		if properties.Info.Role == "user" {
			users[properties.Info.ID] = true
		}
	case "message.part.updated":
		if properties.SessionID != session {
			return
		}
		part := properties.Part
		switch part.Type {
		case "text":
			texts[part.ID] = !users[part.MessageID]
		case "tool":
			status := part.State.Status
			if status != tools[part.ID] && (status == "running" || status == "error") {
				emit(Event{Type: "tool", Tool: strings.TrimPrefix(part.Tool, "quiqr_"), Message: status})
			}
			tools[part.ID] = status
		}
	case "message.part.delta":
		if properties.SessionID == session && properties.Field == "text" && texts[properties.PartID] && !users[properties.MessageID] {
			emit(Event{Type: "text", Text: properties.Delta})
		}
	case "session.idle":
		if properties.SessionID == session {
			emit(Event{Type: "done"})
		}
	case "session.error":
		if properties.SessionID == session {
			message := "the assistant could not answer"
			if properties.Error != nil && properties.Error.Data.Message != "" {
				message = properties.Error.Data.Message
			}
			emit(Event{Type: "error", Message: message})
		}
	}
}
