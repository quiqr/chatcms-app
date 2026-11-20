package assistant

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/aigateway"
	"github.com/quiqr/chatcms-app/internal/httpapi"
)

type callers map[string]aigateway.Caller

func (c callers) Caller(ctx context.Context, token string) (aigateway.Caller, bool, error) {
	if token == "broken" {
		return aigateway.Caller{}, false, errors.New("database down")
	}
	caller, found := c[token]
	return caller, found, nil
}

type pages struct {
	updated []map[string]any
	refuse  error
}

func (p *pages) List(ctx context.Context, caller aigateway.Caller) ([]PageRef, error) {
	return []PageRef{{Key: "home", Title: "Home page", Section: "Pages"}}, nil
}

func (p *pages) Read(ctx context.Context, caller aigateway.Caller, key string) (Page, error) {
	if key != "home" {
		return Page{}, httpapi.ErrMissing
	}
	text := "Come in."
	return Page{Key: "home", Title: "Home page", Fields: []Field{{Path: "title", Label: "Page title", Value: "Korenaar"}}, Text: &text}, nil
}

func (p *pages) Update(ctx context.Context, caller aigateway.Caller, key string, fields map[string]any, text *string) (Change, error) {
	if p.refuse != nil {
		return Change{}, p.refuse
	}
	p.updated = append(p.updated, fields)
	return Change{Page: key, Title: "Home page", Before: "v1", After: "v2"}, nil
}

type announced struct {
	changes []Change
}

func (a *announced) Announce(caller aigateway.Caller, change Change) {
	a.changes = append(a.changes, change)
}

func toolsRig() (*Tools, *pages, *announced) {
	found := &pages{}
	heard := &announced{}
	people := callers{"live": {Person: uuid.New(), Project: uuid.New()}}
	return NewTools(people, found, heard), found, heard
}

func rpc(t *testing.T, tools *Tools, token, body string) (int, map[string]any) {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	recorder := httptest.NewRecorder()
	tools.ServeHTTP(recorder, request)
	var answer map[string]any
	json.Unmarshal(recorder.Body.Bytes(), &answer)
	return recorder.Code, answer
}

func callTool(t *testing.T, tools *Tools, name string, arguments any) (string, bool) {
	t.Helper()
	params, _ := json.Marshal(map[string]any{"name": name, "arguments": arguments})
	_, answer := rpc(t, tools, "live", `{"jsonrpc":"2.0","id":7,"method":"tools/call","params":`+string(params)+`}`)
	result := answer["result"].(map[string]any)
	content := result["content"].([]any)[0].(map[string]any)
	return content["text"].(string), result["isError"].(bool)
}

func TestTheToolsSpeakMCP(t *testing.T) {
	tools, _, _ := toolsRig()

	_, initialised := rpc(t, tools, "live", `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`)
	result := initialised["result"].(map[string]any)
	if result["protocolVersion"] == "" || result["serverInfo"].(map[string]any)["name"] != "quiqr" {
		t.Errorf("initialize = %v", initialised)
	}

	if code, _ := rpc(t, tools, "live", `{"jsonrpc":"2.0","method":"notifications/initialized"}`); code != http.StatusAccepted {
		t.Errorf("a notification = %d, want 202", code)
	}

	_, listed := rpc(t, tools, "live", `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`)
	var names []string
	for _, tool := range listed["result"].(map[string]any)["tools"].([]any) {
		names = append(names, tool.(map[string]any)["name"].(string))
	}
	if strings.Join(names, ",") != "list_pages,read_page,update_page" {
		t.Errorf("tools = %v", names)
	}

	if _, pinged := rpc(t, tools, "live", `{"jsonrpc":"2.0","id":3,"method":"ping"}`); pinged["result"] == nil {
		t.Errorf("ping = %v", pinged)
	}
	if _, unknown := rpc(t, tools, "live", `{"jsonrpc":"2.0","id":4,"method":"resources/list"}`); unknown["error"] == nil {
		t.Errorf("an unknown method = %v", unknown)
	}
	if _, garbled := rpc(t, tools, "live", `{`); garbled["error"] == nil {
		t.Errorf("garbled JSON = %v", garbled)
	}
}

func TestTheToolsWantALiveToken(t *testing.T) {
	tools, _, _ := toolsRig()
	for token, want := range map[string]int{"": http.StatusUnauthorized, "stale": http.StatusUnauthorized, "broken": http.StatusServiceUnavailable} {
		if code, _ := rpc(t, tools, token, `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`); code != want {
			t.Errorf("token %q = %d, want %d", token, code, want)
		}
	}
	for _, method := range []string{http.MethodGet, http.MethodDelete, http.MethodPut} {
		recorder := httptest.NewRecorder()
		tools.ServeHTTP(recorder, httptest.NewRequest(method, "/mcp", nil))
		if recorder.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s = %d", method, recorder.Code)
		}
	}
}

func TestListingAndReadingPages(t *testing.T) {
	tools, _, _ := toolsRig()

	listed, failed := callTool(t, tools, "list_pages", map[string]any{})
	if failed || !strings.Contains(listed, `"home"`) {
		t.Errorf("list_pages = %s", listed)
	}
	read, failed := callTool(t, tools, "read_page", map[string]any{"key": "home"})
	if failed || !strings.Contains(read, "Page title") || !strings.Contains(read, "Come in.") {
		t.Errorf("read_page = %s", read)
	}
	if missing, failed := callTool(t, tools, "read_page", map[string]any{"key": "nope"}); !failed || !strings.Contains(missing, "no such page") {
		t.Errorf("read_page of a missing page = %s, %v", missing, failed)
	}
	if empty, failed := callTool(t, tools, "read_page", map[string]any{}); !failed || !strings.Contains(empty, "which page") {
		t.Errorf("read_page with no key = %s", empty)
	}
}

func TestUpdatingAPageSavesItAndSaysSo(t *testing.T) {
	tools, found, heard := toolsRig()

	said, failed := callTool(t, tools, "update_page", map[string]any{"key": "home", "fields": map[string]any{"title": "Korenaar by night"}})
	if failed || !strings.Contains(said, "Saved") {
		t.Fatalf("update_page = %s, %v", said, failed)
	}
	if len(found.updated) != 1 || found.updated[0]["title"] != "Korenaar by night" {
		t.Errorf("updated %v", found.updated)
	}
	if len(heard.changes) != 1 || heard.changes[0].Before != "v1" || heard.changes[0].After != "v2" {
		t.Errorf("announced %v", heard.changes)
	}

	if nothing, failed := callTool(t, tools, "update_page", map[string]any{"key": "home"}); !failed || !strings.Contains(nothing, "what to change") {
		t.Errorf("update_page with nothing to change = %s", nothing)
	}
}

func TestARefusedUpdateIsAToolErrorInPlainWords(t *testing.T) {
	for refusal, want := range map[error]string{
		&httpapi.RefusedError{Reason: `this page has no field called "colour"`}: `no field called "colour"`,
		httpapi.ErrNotAllowed:   "may not change this site",
		errors.New("disk full"): "disk full",
	} {
		tools, found, heard := toolsRig()
		found.refuse = refusal
		said, failed := callTool(t, tools, "update_page", map[string]any{"key": "home", "fields": map[string]any{"colour": "red"}})
		if !failed || !strings.Contains(said, want) {
			t.Errorf("refused with %v = %s, %v", refusal, said, failed)
		}
		if len(heard.changes) != 0 {
			t.Errorf("a refused update was announced")
		}
	}
	tools, _, _ := toolsRig()
	if unknown, failed := callTool(t, tools, "delete_site", map[string]any{}); !failed || !strings.Contains(unknown, "no tool") {
		t.Errorf("an unknown tool = %s", unknown)
	}
}

func TestTheAgentsConfigurationKeepsItToTheTools(t *testing.T) {
	var config map[string]any
	if err := json.Unmarshal([]byte(OpenCodeConfig("http://10.231.0.1:8091/", "tok")), &config); err != nil {
		t.Fatal(err)
	}
	provider := config["providers"].(map[string]any)[Provider].(map[string]any)["settings"].(map[string]any)
	if provider["baseURL"] != "http://10.231.0.1:8091/v1" || provider["apiKey"] != "tok" {
		t.Errorf("provider = %v", provider)
	}
	mcp := config["mcp"].(map[string]any)["servers"].(map[string]any)["quiqr"].(map[string]any)
	if mcp["url"] != "http://10.231.0.1:8091/mcp" || mcp["headers"].(map[string]any)["Authorization"] != "Bearer tok" || mcp["codemode"] != false {
		t.Errorf("mcp = %v", mcp)
	}
	effects := map[string]string{}
	for _, rule := range config["permissions"].([]any) {
		rule := rule.(map[string]any)
		if rule["resource"] != "*" {
			t.Errorf("rule %v is not for every resource", rule)
		}
		effects[rule["action"].(string)] = rule["effect"].(string)
	}
	for _, denied := range []string{"edit", "shell", "execute", "webfetch", "websearch", "subagent", "external_directory"} {
		if effects[denied] != "deny" {
			t.Errorf("%s is %q, want deny", denied, effects[denied])
		}
	}
	if effects["read"] != "allow" {
		t.Errorf("read is %q, want allow", effects["read"])
	}
	build := config["agents"].(map[string]any)["build"].(map[string]any)
	if build["system"] != Prompt {
		t.Errorf("the build agent's prompt = %v", build["system"])
	}

	environment := strings.Join(OpenCodeEnvironment("/run/agent", "http://h:1", "tok"), "\n")
	for _, want := range []string{"OPENCODE_DISABLE_PROJECT_CONFIG=1", "OPENCODE_DISABLE_MODELS_FETCH=1", "OPENCODE_SERVER_PASSWORD=tok", "HOME=/run/agent"} {
		if !strings.Contains(environment, want) {
			t.Errorf("the environment lacks %s", want)
		}
	}
}

func TestTheAgentsEnvironmentIsGivenOnlyForALiveToken(t *testing.T) {
	served := NewAgentEnvironment(callers{"live": {}}, "/run/agent")
	ask := func(token string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodGet, "http://10.231.3.1:8091/agent/env", nil)
		if token != "" {
			request.Header.Set("Authorization", "Bearer "+token)
		}
		recorder := httptest.NewRecorder()
		served.ServeHTTP(recorder, request)
		return recorder
	}

	answer := ask("live")
	if answer.Code != http.StatusOK || !bytes.Contains(answer.Body.Bytes(), []byte("OPENCODE_CONFIG_CONTENT='")) || !bytes.Contains(answer.Body.Bytes(), []byte("http://10.231.3.1:8091/v1")) {
		t.Errorf("env = %d %s", answer.Code, answer.Body)
	}
	if ask("").Code != http.StatusUnauthorized || ask("stale").Code != http.StatusUnauthorized || ask("broken").Code != http.StatusServiceUnavailable {
		t.Errorf("a missing, stale or unreadable token was given the environment")
	}
	recorder := httptest.NewRecorder()
	served.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/agent/env", nil))
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST = %d", recorder.Code)
	}
	if shellQuote("it's") != `'it'\''s'` {
		t.Errorf("shellQuote = %s", shellQuote("it's"))
	}
}
