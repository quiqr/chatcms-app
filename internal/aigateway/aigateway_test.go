package aigateway

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/workspace"
)

type memory struct {
	mu      sync.Mutex
	tokens  map[string]Caller
	live    map[string]bool
	usage   []Usage
	limits  map[uuid.UUID]Limits
	orgOf   uuid.UUID
	lookErr error
}

func newMemory() *memory {
	return &memory{tokens: map[string]Caller{}, live: map[string]bool{}, limits: map[uuid.UUID]Limits{}, orgOf: uuid.New()}
}

func (m *memory) Issue(ctx context.Context, person, project uuid.UUID, hash []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokens[string(hash)] = Caller{Person: person, Project: project, Organisation: m.orgOf}
	m.live[string(hash)] = true
	return nil
}

func (m *memory) Revoke(ctx context.Context, person, project uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for hash, caller := range m.tokens {
		if caller.Person == person && caller.Project == project {
			m.live[hash] = false
		}
	}
	return nil
}

func (m *memory) LookUp(ctx context.Context, hash []byte) (Caller, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.lookErr != nil {
		return Caller{}, false, m.lookErr
	}
	caller, found := m.tokens[string(hash)]
	return caller, found && m.live[string(hash)], nil
}

func (m *memory) Record(ctx context.Context, usage Usage) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.usage = append(m.usage, usage)
	return nil
}

func (m *memory) UsedSince(ctx context.Context, organisation uuid.UUID, since time.Time) (int64, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var requests, tokens int64
	for _, usage := range m.usage {
		if usage.Caller.Organisation == organisation {
			requests++
			tokens += usage.TokensIn + usage.TokensOut
		}
	}
	return requests, tokens, nil
}

func (m *memory) Limits(ctx context.Context, organisation uuid.UUID) (Limits, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.limits[organisation], nil
}

func (m *memory) SetLimits(ctx context.Context, organisation uuid.UUID, limits Limits) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.limits[organisation] = limits
	return nil
}

func (m *memory) recorded() []Usage {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]Usage{}, m.usage...)
}

type provider struct {
	mu   sync.Mutex
	seen []map[string]any
	auth []string
}

func (p *provider) server(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		p.mu.Lock()
		p.seen = append(p.seen, body)
		p.auth = append(p.auth, r.Header.Get("Authorization"))
		p.mu.Unlock()

		if r.URL.Path != "/v1/chat/completions" {
			http.NotFound(w, r)
			return
		}
		if stream, _ := body["stream"].(bool); stream {
			w.Header().Set("Content-Type", "text/event-stream")
			flusher := w.(http.Flusher)
			for _, piece := range []string{"Hel", "lo"} {
				fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":%q}}]}\n\n", piece)
				flusher.Flush()
			}
			fmt.Fprint(w, "data: {\"choices\":[],\"usage\":{\"prompt_tokens\":12,\"completion_tokens\":3}}\n\n")
			fmt.Fprint(w, "data: [DONE]\n\n")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"choices":[{"message":{"content":"Hello"}}],"usage":{"prompt_tokens":10,"completion_tokens":5}}`)
	}))
	t.Cleanup(server.Close)
	return server
}

type rig struct {
	store    *memory
	provider *provider
	gateway  *Gateway
	tokens   *Tokens
	clock    time.Time
}

func newRig(t *testing.T) *rig {
	t.Helper()
	r := &rig{store: newMemory(), provider: &provider{}, clock: time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)}
	server := r.provider.server(t)
	gateway, err := New(r.store, Settings{
		ProviderURL: server.URL + "/v1/",
		Key:         "the-real-key",
		Model:       "deepseek-flash",
		Now:         func() time.Time { return r.clock },
	})
	if err != nil {
		t.Fatal(err)
	}
	r.gateway = gateway
	r.tokens = NewTokens(r.store)
	return r
}

func (r *rig) token(t *testing.T) (string, workspace.Key) {
	t.Helper()
	key := workspace.Key{Person: uuid.New(), Project: uuid.New()}
	token, err := r.tokens.Issue(t.Context(), key)
	if err != nil {
		t.Fatal(err)
	}
	return token, key
}

func (r *rig) ask(t *testing.T, method, path, token, body, from string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.RemoteAddr = from
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	recorder := httptest.NewRecorder()
	r.gateway.ServeHTTP(recorder, request)
	return recorder
}

const workspaceAddress = "10.231.0.2:40000"

func code(t *testing.T, recorder *httptest.ResponseRecorder) string {
	t.Helper()
	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	json.Unmarshal(recorder.Body.Bytes(), &body)
	if strings.Contains(recorder.Body.String(), "the-real-key") {
		t.Errorf("an answer carried the provider's key")
	}
	return body.Error.Code
}

func TestAWorkspaceTokenReachesTheModelWithTheRealKey(t *testing.T) {
	r := newRig(t)
	token, key := r.token(t)

	answer := r.ask(t, http.MethodPost, "/v1/chat/completions", token,
		`{"model":"gpt-something-expensive","messages":[{"role":"user","content":"hi"}]}`, workspaceAddress)
	if answer.Code != http.StatusOK || !strings.Contains(answer.Body.String(), "Hello") {
		t.Fatalf("answer = %d %s", answer.Code, answer.Body)
	}
	if r.provider.auth[0] != "Bearer the-real-key" {
		t.Errorf("the provider was sent %q, want the real key", r.provider.auth[0])
	}
	if r.provider.seen[0]["model"] != "deepseek-flash" {
		t.Errorf("the provider was asked for model %v, want the configured one", r.provider.seen[0]["model"])
	}
	used := r.store.recorded()
	if len(used) != 1 || used[0].TokensIn != 10 || used[0].TokensOut != 5 || used[0].Caller.Person != key.Person || used[0].Status != 200 {
		t.Errorf("recorded %+v", used)
	}
}

func TestAStreamedAnswerPassesThroughAndIsCounted(t *testing.T) {
	r := newRig(t)
	token, _ := r.token(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		req.RemoteAddr = workspaceAddress
		r.gateway.ServeHTTP(w, req)
	}))
	defer server.Close()

	request, _ := http.NewRequest(http.MethodPost, server.URL+"/v1/chat/completions",
		strings.NewReader(`{"stream":true,"messages":[{"role":"user","content":"hi"}]}`))
	request.Header.Set("Authorization", "Bearer "+token)
	answer, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer answer.Body.Close()

	var pieces []string
	scanner := bufio.NewScanner(answer.Body)
	for scanner.Scan() {
		if line := scanner.Text(); strings.HasPrefix(line, "data:") {
			pieces = append(pieces, line)
		}
	}
	if len(pieces) != 4 || !strings.Contains(pieces[0], "Hel") || pieces[3] != "data: [DONE]" {
		t.Errorf("pieces = %v", pieces)
	}
	options, _ := r.provider.seen[0]["stream_options"].(map[string]any)
	if options["include_usage"] != true {
		t.Errorf("the provider was not asked for usage: %v", r.provider.seen[0])
	}
	used := r.store.recorded()
	if len(used) != 1 || used[0].TokensIn != 12 || used[0].TokensOut != 3 {
		t.Errorf("recorded %+v", used)
	}
}

func TestOnlyALiveTokenIsAccepted(t *testing.T) {
	r := newRig(t)
	token, key := r.token(t)

	if got := r.ask(t, http.MethodPost, "/v1/chat/completions", "", `{}`, workspaceAddress); got.Code != http.StatusUnauthorized || code(t, got) != "token" {
		t.Errorf("no token = %d %s", got.Code, got.Body)
	}
	if got := r.ask(t, http.MethodPost, "/v1/chat/completions", "made-up", `{}`, workspaceAddress); got.Code != http.StatusUnauthorized {
		t.Errorf("an unknown token = %d", got.Code)
	}

	r.tokens.Revoke(t.Context(), key)
	if got := r.ask(t, http.MethodPost, "/v1/chat/completions", token, `{}`, workspaceAddress); got.Code != http.StatusUnauthorized {
		t.Errorf("a revoked token = %d", got.Code)
	}

	again, _ := r.tokens.Issue(t.Context(), key)
	if got := r.ask(t, http.MethodPost, "/v1/chat/completions", token, `{}`, workspaceAddress); got.Code != http.StatusUnauthorized {
		t.Errorf("the token before a fresh one = %d", got.Code)
	}
	if got := r.ask(t, http.MethodPost, "/v1/chat/completions", again, `{}`, workspaceAddress); got.Code != http.StatusOK {
		t.Errorf("the fresh token = %d %s", got.Code, got.Body)
	}
	if len(r.provider.seen) != 1 {
		t.Errorf("the provider was asked %d times, want only for the live token", len(r.provider.seen))
	}
}

func TestOnlyWorkspacesMayAsk(t *testing.T) {
	r := newRig(t)
	token, _ := r.token(t)

	for from, want := range map[string]int{
		"203.0.113.5:1000": http.StatusForbidden,
		"192.168.1.10:999": http.StatusForbidden,
		"nonsense":         http.StatusForbidden,
		"127.0.0.1:5000":   http.StatusOK,
		"10.231.7.2:5000":  http.StatusOK,
	} {
		if got := r.ask(t, http.MethodPost, "/v1/chat/completions", token, `{}`, from); got.Code != want {
			t.Errorf("from %s = %d, want %d", from, got.Code, want)
		}
	}
}

func TestOnlyTwoRoutesAreAnswered(t *testing.T) {
	r := newRig(t)
	token, _ := r.token(t)

	models := r.ask(t, http.MethodGet, "/v1/models", token, "", workspaceAddress)
	if models.Code != http.StatusOK || !strings.Contains(models.Body.String(), "deepseek-flash") {
		t.Errorf("models = %d %s", models.Code, models.Body)
	}
	if got := r.ask(t, http.MethodGet, "/v1/models", "", "", workspaceAddress); got.Code != http.StatusUnauthorized {
		t.Errorf("models without a token = %d", got.Code)
	}
	for _, path := range []string{"/v1/embeddings", "/v1/files", "/"} {
		if got := r.ask(t, http.MethodPost, path, token, `{}`, workspaceAddress); got.Code != http.StatusNotFound {
			t.Errorf("%s = %d, want 404", path, got.Code)
		}
	}
}

func TestABadRequestIsRefusedBeforeTheProvider(t *testing.T) {
	r := newRig(t)
	token, _ := r.token(t)

	if got := r.ask(t, http.MethodPost, "/v1/chat/completions", token, `[1,2]`, workspaceAddress); got.Code != http.StatusBadRequest {
		t.Errorf("a JSON array = %d", got.Code)
	}
	huge := `{"x":"` + strings.Repeat("a", maxBody) + `"}`
	if got := r.ask(t, http.MethodPost, "/v1/chat/completions", token, huge, workspaceAddress); got.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("a huge body = %d", got.Code)
	}
	if len(r.provider.seen) != 0 {
		t.Errorf("the provider was asked")
	}
}

func TestOverTheMonthlyBudgetIsRefusedPlainly(t *testing.T) {
	r := newRig(t)
	token, _ := r.token(t)
	budget := int64(20)
	r.store.SetLimits(t.Context(), r.store.orgOf, Limits{TokensPerMonth: &budget})

	if got := r.ask(t, http.MethodPost, "/v1/chat/completions", token, `{}`, workspaceAddress); got.Code != http.StatusOK {
		t.Fatalf("the first request = %d", got.Code)
	}
	r.ask(t, http.MethodPost, "/v1/chat/completions", token, `{}`, workspaceAddress)
	refused := r.ask(t, http.MethodPost, "/v1/chat/completions", token, `{}`, workspaceAddress)
	if refused.Code != http.StatusTooManyRequests || code(t, refused) != "budget" || !strings.Contains(refused.Body.String(), "budget for this month") {
		t.Errorf("over budget = %d %s", refused.Code, refused.Body)
	}
	if len(r.provider.seen) != 2 {
		t.Errorf("the provider was asked %d times, want twice", len(r.provider.seen))
	}
}

func TestTooManyAMinuteIsRefusedUntilTheMinutePasses(t *testing.T) {
	r := newRig(t)
	token, _ := r.token(t)
	rate := 2
	r.store.SetLimits(t.Context(), r.store.orgOf, Limits{RequestsPerMinute: &rate})

	for range 2 {
		if got := r.ask(t, http.MethodPost, "/v1/chat/completions", token, `{}`, workspaceAddress); got.Code != http.StatusOK {
			t.Fatalf("within the rate = %d", got.Code)
		}
	}
	refused := r.ask(t, http.MethodPost, "/v1/chat/completions", token, `{}`, workspaceAddress)
	if refused.Code != http.StatusTooManyRequests || code(t, refused) != "rate" || !strings.Contains(refused.Body.String(), "wait a moment") {
		t.Errorf("over the rate = %d %s", refused.Code, refused.Body)
	}

	r.clock = r.clock.Add(61 * time.Second)
	if got := r.ask(t, http.MethodPost, "/v1/chat/completions", token, `{}`, workspaceAddress); got.Code != http.StatusOK {
		t.Errorf("a minute later = %d", got.Code)
	}
}

func TestAProviderThatCannotBeReachedIsSaidAndRecorded(t *testing.T) {
	store := newMemory()
	gateway, _ := New(store, Settings{ProviderURL: "http://127.0.0.1:1", Key: "k", Model: "m"})
	tokens := NewTokens(store)
	token, _ := tokens.Issue(t.Context(), workspace.Key{Person: uuid.New(), Project: uuid.New()})

	request := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{}`))
	request.RemoteAddr = workspaceAddress
	request.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()
	gateway.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadGateway {
		t.Errorf("an unreachable provider = %d", recorder.Code)
	}
	if used := store.recorded(); len(used) != 1 || used[0].Status != http.StatusBadGateway {
		t.Errorf("recorded %+v", used)
	}
}

func TestTroubleCheckingATokenIsNotAYes(t *testing.T) {
	r := newRig(t)
	token, _ := r.token(t)
	r.store.lookErr = errors.New("database down")

	if got := r.ask(t, http.MethodPost, "/v1/chat/completions", token, `{}`, workspaceAddress); got.Code != http.StatusServiceUnavailable {
		t.Errorf("a failing lookup = %d", got.Code)
	}
}

func TestTheBooksSayThisMonthsUseAndKeepLimits(t *testing.T) {
	r := newRig(t)
	token, _ := r.token(t)
	r.ask(t, http.MethodPost, "/v1/chat/completions", token, `{}`, workspaceAddress)

	books := NewBooks(r.store, func() time.Time { return r.clock })
	rate, budget := 30, int64(2_000_000)
	if err := books.SetLimits(t.Context(), r.store.orgOf, Limits{RequestsPerMinute: &rate, TokensPerMonth: &budget}); err != nil {
		t.Fatal(err)
	}
	limits, requests, tokens, err := books.Use(t.Context(), r.store.orgOf)
	if err != nil || *limits.RequestsPerMinute != 30 || *limits.TokensPerMonth != 2_000_000 || requests != 1 || tokens != 15 {
		t.Errorf("Use() = %+v, %d, %d, %v", limits, requests, tokens, err)
	}

	zero := 0
	if err := books.SetLimits(t.Context(), r.store.orgOf, Limits{RequestsPerMinute: &zero}); !errors.Is(err, ErrBelowOne) {
		t.Errorf("a limit of zero error = %v", err)
	}
}

func TestNewNeedsItsParts(t *testing.T) {
	if _, err := New(nil, Settings{}); !errors.Is(err, ErrIncomplete) {
		t.Errorf("New() error = %v", err)
	}
	if _, err := New(newMemory(), Settings{ProviderURL: "http://x", Key: "k"}); !errors.Is(err, ErrIncomplete) {
		t.Errorf("New() without a model error = %v", err)
	}
	if monthStart(time.Date(2026, 3, 17, 23, 0, 0, 0, time.FixedZone("x", 3*3600))) != time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC) {
		t.Errorf("monthStart is not the first of the month in UTC")
	}
}
