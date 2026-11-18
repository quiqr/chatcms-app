package assistant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/quiqr/chatcms-app/internal/aigateway"
	"github.com/quiqr/chatcms-app/internal/httpapi"
)

const protocolVersion = "2025-06-18"

type PageRef struct {
	Key     string `json:"key"`
	Title   string `json:"title"`
	Section string `json:"section,omitempty"`
}

type Field struct {
	Path  string `json:"path"`
	Label string `json:"label"`
	Value any    `json:"value"`
}

type Page struct {
	Key    string  `json:"key"`
	Title  string  `json:"title"`
	Fields []Field `json:"fields"`
	Text   *string `json:"text,omitempty"`
}

type Change struct {
	Page   string `json:"page"`
	Title  string `json:"title"`
	Before string `json:"before"`
	After  string `json:"after"`
}

type Pages interface {
	List(ctx context.Context, caller aigateway.Caller) ([]PageRef, error)
	Read(ctx context.Context, caller aigateway.Caller, key string) (Page, error)
	Update(ctx context.Context, caller aigateway.Caller, key string, fields map[string]any, text *string) (Change, error)
}

type Callers interface {
	Caller(ctx context.Context, token string) (aigateway.Caller, bool, error)
}

type Announcer interface {
	Announce(caller aigateway.Caller, change Change)
}

type Tools struct {
	callers   Callers
	pages     Pages
	announcer Announcer
}

func NewTools(callers Callers, pages Pages, announcer Announcer) *Tools {
	return &Tools{callers: callers, pages: pages, announcer: announcer}
}

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func answer(w http.ResponseWriter, id json.RawMessage, result any, failure *rpcError) {
	body := map[string]any{"jsonrpc": "2.0", "id": id}
	if failure != nil {
		body["error"] = failure
	} else {
		body["result"] = result
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(body)
}

var toolList = []map[string]any{
	{
		"name":        "list_pages",
		"description": "List every page of the site you may edit, with its key and title.",
		"inputSchema": map[string]any{"type": "object", "properties": map[string]any{}},
	},
	{
		"name":        "read_page",
		"description": "Read one page: each field with its path, label and value, and the page's text.",
		"inputSchema": map[string]any{
			"type":       "object",
			"properties": map[string]any{"key": map[string]any{"type": "string", "description": "The page's key, from list_pages."}},
			"required":   []string{"key"},
		},
	},
	{
		"name":        "update_page",
		"description": "Change a page's fields and, if it has one, its text. Fields are given by path, as read_page names them. The change is saved like an edit in the form.",
		"inputSchema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"key":    map[string]any{"type": "string", "description": "The page's key, from list_pages."},
				"fields": map[string]any{"type": "object", "description": "Field paths and their new values.", "additionalProperties": true},
				"text":   map[string]any{"type": "string", "description": "The page's whole new text, in Markdown. Leave out to keep it."},
			},
			"required": []string{"key"},
		},
	},
}

func (t *Tools) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet || r.Method == http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	token, _ := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	caller, known, err := t.callers.Caller(r.Context(), token)
	if err != nil {
		http.Error(w, "the tools cannot check tokens right now", http.StatusServiceUnavailable)
		return
	}
	if !known {
		http.Error(w, "this token is not, or no longer, valid", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "unreadable", http.StatusBadRequest)
		return
	}
	var asked request
	if err := json.Unmarshal(body, &asked); err != nil {
		answer(w, nil, nil, &rpcError{Code: -32700, Message: "not JSON"})
		return
	}
	if len(asked.ID) == 0 {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	switch asked.Method {
	case "initialize":
		answer(w, asked.ID, map[string]any{
			"protocolVersion": protocolVersion,
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]any{"name": "quiqr", "version": "1"},
		}, nil)
	case "ping":
		answer(w, asked.ID, map[string]any{}, nil)
	case "tools/list":
		answer(w, asked.ID, map[string]any{"tools": toolList}, nil)
	case "tools/call":
		answer(w, asked.ID, t.call(r.Context(), caller, asked.Params), nil)
	default:
		answer(w, asked.ID, nil, &rpcError{Code: -32601, Message: "no such method"})
	}
}

func result(text string, failed bool) map[string]any {
	return map[string]any{
		"content": []map[string]any{{"type": "text", "text": text}},
		"isError": failed,
	}
}

func asJSON(value any) string {
	data, _ := json.MarshalIndent(value, "", "  ")
	return string(data)
}

func plainly(err error) string {
	var refused *httpapi.RefusedError
	switch {
	case errors.As(err, &refused):
		return refused.Reason
	case errors.Is(err, httpapi.ErrNotAllowed):
		return "the person you are working for may not change this site"
	case errors.Is(err, httpapi.ErrMissing):
		return "there is no such page"
	default:
		return "that did not work: " + err.Error()
	}
}

func (t *Tools) call(ctx context.Context, caller aigateway.Caller, params json.RawMessage) map[string]any {
	var asked struct {
		Name      string `json:"name"`
		Arguments struct {
			Key    string         `json:"key"`
			Fields map[string]any `json:"fields"`
			Text   *string        `json:"text"`
		} `json:"arguments"`
	}
	if err := json.Unmarshal(params, &asked); err != nil {
		return result("the arguments are not readable", true)
	}

	switch asked.Name {
	case "list_pages":
		pages, err := t.pages.List(ctx, caller)
		if err != nil {
			return result(plainly(err), true)
		}
		return result(asJSON(pages), false)
	case "read_page":
		if asked.Arguments.Key == "" {
			return result("say which page, by its key", true)
		}
		page, err := t.pages.Read(ctx, caller, asked.Arguments.Key)
		if err != nil {
			return result(plainly(err), true)
		}
		return result(asJSON(page), false)
	case "update_page":
		if asked.Arguments.Key == "" {
			return result("say which page, by its key", true)
		}
		if len(asked.Arguments.Fields) == 0 && asked.Arguments.Text == nil {
			return result("say what to change: fields, text or both", true)
		}
		change, err := t.pages.Update(ctx, caller, asked.Arguments.Key, asked.Arguments.Fields, asked.Arguments.Text)
		if err != nil {
			return result(plainly(err), true)
		}
		if change.After != "" && change.After != change.Before && t.announcer != nil {
			t.announcer.Announce(caller, change)
		}
		return result(fmt.Sprintf("Saved %q. It shows in the preview within seconds.", change.Title), false)
	default:
		return result("there is no tool called "+asked.Name, true)
	}
}
