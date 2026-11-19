package aigateway

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/workspace"
)

const (
	maxBody         = 1 << 20
	providerTimeout = 5 * time.Minute
	DefaultListen   = "127.0.0.1:8091"
)

var (
	ErrIncomplete = errors.New("aigateway: missing a part it needs")
	ErrNotFound   = errors.New("aigateway: no such organisation")
)

type Caller struct {
	Person       uuid.UUID
	Project      uuid.UUID
	Organisation uuid.UUID
}

type Limits struct {
	RequestsPerMinute *int
	TokensPerMonth    *int64
}

type Usage struct {
	Caller    Caller
	Model     string
	TokensIn  int64
	TokensOut int64
	Status    int
}

type Store interface {
	Issue(ctx context.Context, person, project uuid.UUID, hash []byte) error
	Revoke(ctx context.Context, person, project uuid.UUID) error
	LookUp(ctx context.Context, hash []byte) (Caller, bool, error)
	Record(ctx context.Context, usage Usage) error
	UsedSince(ctx context.Context, organisation uuid.UUID, since time.Time) (int64, int64, error)
	Limits(ctx context.Context, organisation uuid.UUID) (Limits, error)
	SetLimits(ctx context.Context, organisation uuid.UUID, limits Limits) error
}

func hash(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

type Tokens struct {
	store Store
}

func NewTokens(store Store) *Tokens {
	return &Tokens{store: store}
}

func (t *Tokens) Issue(ctx context.Context, key workspace.Key) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	if err := t.store.Revoke(ctx, key.Person, key.Project); err != nil {
		return "", err
	}
	if err := t.store.Issue(ctx, key.Person, key.Project, hash(token)); err != nil {
		return "", err
	}
	return token, nil
}

func (t *Tokens) Revoke(ctx context.Context, key workspace.Key) error {
	return t.store.Revoke(ctx, key.Person, key.Project)
}

var _ workspace.Tokens = (*Tokens)(nil)

type Settings struct {
	ProviderURL   string
	Key           string
	Model         string
	SessionHeader string
	UserAgent     string
	Client        *http.Client
	Logger        *slog.Logger
	Now           func() time.Time
	Allowed       []*net.IPNet
}

func (s Settings) userAgent() string {
	if s.UserAgent != "" {
		return s.UserAgent
	}
	return "quiqr-chatcms"
}

func sessionOf(caller Caller) string {
	sum := sha256.Sum256([]byte("chatcms-session\x00" + caller.Person.String() + "\x00" + caller.Project.String()))
	return hex.EncodeToString(sum[:16])
}

type Gateway struct {
	settings Settings
	store    Store

	mu     sync.Mutex
	recent map[uuid.UUID][]time.Time
}

func DefaultAllowed() []*net.IPNet {
	var nets []*net.IPNet
	for _, cidr := range []string{"127.0.0.0/8", "::1/128", "10.231.0.0/16"} {
		_, network, _ := net.ParseCIDR(cidr)
		nets = append(nets, network)
	}
	return nets
}

func New(store Store, settings Settings) (*Gateway, error) {
	if store == nil || settings.ProviderURL == "" || settings.Key == "" || settings.Model == "" {
		return nil, ErrIncomplete
	}
	settings.ProviderURL = strings.TrimSuffix(settings.ProviderURL, "/")
	if settings.Client == nil {
		settings.Client = &http.Client{Timeout: providerTimeout}
	}
	if settings.Logger == nil {
		settings.Logger = slog.New(slog.DiscardHandler)
	}
	if settings.Now == nil {
		settings.Now = time.Now
	}
	if settings.Allowed == nil {
		settings.Allowed = DefaultAllowed()
	}
	return &Gateway{settings: settings, store: store, recent: map[uuid.UUID][]time.Time{}}, nil
}

func refuse(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"message": message, "type": "chatcms_gateway", "code": code},
	})
}

func (g *Gateway) from(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	for _, network := range g.settings.Allowed {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !g.from(r) {
		refuse(w, http.StatusForbidden, "address", "this gateway answers workspaces only")
		return
	}
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/v1/models":
		g.models(w, r)
	case r.Method == http.MethodPost && r.URL.Path == "/v1/chat/completions":
		g.complete(w, r)
	default:
		refuse(w, http.StatusNotFound, "not_found", "there is nothing here")
	}
}

func (g *Gateway) caller(w http.ResponseWriter, r *http.Request) (Caller, bool) {
	token, found := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !found || token == "" {
		refuse(w, http.StatusUnauthorized, "token", "this workspace has no token for the gateway")
		return Caller{}, false
	}
	caller, known, err := g.store.LookUp(r.Context(), hash(token))
	if err != nil {
		g.settings.Logger.Warn("looking up a gateway token failed", "error", err)
		refuse(w, http.StatusServiceUnavailable, "unavailable", "the gateway cannot check tokens right now")
		return Caller{}, false
	}
	if !known {
		refuse(w, http.StatusUnauthorized, "token", "this token is not, or no longer, valid")
		return Caller{}, false
	}
	return caller, true
}

func (g *Gateway) models(w http.ResponseWriter, r *http.Request) {
	if _, ok := g.caller(w, r); !ok {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"object": "list",
		"data":   []map[string]any{{"id": g.settings.Model, "object": "model", "owned_by": "chatcms"}},
	})
}

func monthStart(now time.Time) time.Time {
	now = now.UTC()
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func (g *Gateway) admit(ctx context.Context, organisation uuid.UUID) (int, string, string) {
	limits, err := g.store.Limits(ctx, organisation)
	if err != nil {
		return http.StatusServiceUnavailable, "unavailable", "the gateway cannot check limits right now"
	}
	now := g.settings.Now()

	if limits.TokensPerMonth != nil {
		_, tokens, err := g.store.UsedSince(ctx, organisation, monthStart(now))
		if err != nil {
			return http.StatusServiceUnavailable, "unavailable", "the gateway cannot check limits right now"
		}
		if tokens >= *limits.TokensPerMonth {
			return http.StatusTooManyRequests, "budget", "this organisation has used its AI budget for this month; it starts again on the first of next month"
		}
	}

	g.mu.Lock()
	defer g.mu.Unlock()
	window := g.recent[organisation][:0]
	for _, at := range g.recent[organisation] {
		if now.Sub(at) < time.Minute {
			window = append(window, at)
		}
	}
	if limits.RequestsPerMinute != nil && len(window) >= *limits.RequestsPerMinute {
		g.recent[organisation] = window
		return http.StatusTooManyRequests, "rate", "too many requests from this organisation this minute; wait a moment and try again"
	}
	g.recent[organisation] = append(window, now)
	return 0, "", ""
}

type usageBody struct {
	Usage *struct {
		PromptTokens     int64 `json:"prompt_tokens"`
		CompletionTokens int64 `json:"completion_tokens"`
	} `json:"usage"`
}

func (g *Gateway) complete(w http.ResponseWriter, r *http.Request) {
	caller, ok := g.caller(w, r)
	if !ok {
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, maxBody+1))
	if err != nil || len(body) > maxBody {
		refuse(w, http.StatusRequestEntityTooLarge, "too_large", "this request is too large")
		return
	}
	var request map[string]any
	if err := json.Unmarshal(body, &request); err != nil || request == nil {
		refuse(w, http.StatusBadRequest, "invalid", "this request is not a JSON object")
		return
	}
	if status, code, message := g.admit(r.Context(), caller.Organisation); status != 0 {
		refuse(w, status, code, message)
		return
	}

	request["model"] = g.settings.Model
	streamed, _ := request["stream"].(bool)
	if streamed {
		options, _ := request["stream_options"].(map[string]any)
		if options == nil {
			options = map[string]any{}
		}
		options["include_usage"] = true
		request["stream_options"] = options
	}
	forward, _ := json.Marshal(request)

	ctx, cancel := context.WithTimeout(r.Context(), providerTimeout)
	defer cancel()
	outgoing, err := http.NewRequestWithContext(ctx, http.MethodPost, g.settings.ProviderURL+"/chat/completions", bytes.NewReader(forward))
	if err != nil {
		refuse(w, http.StatusInternalServerError, "internal", "the gateway could not make the request")
		return
	}
	outgoing.Header.Set("Content-Type", "application/json")
	outgoing.Header.Set("Authorization", "Bearer "+g.settings.Key)
	outgoing.Header.Set("User-Agent", g.settings.userAgent())
	if g.settings.SessionHeader != "" {
		outgoing.Header.Set(g.settings.SessionHeader, sessionOf(caller))
	}
	if streamed {
		outgoing.Header.Set("Accept", "text/event-stream")
	}

	answer, err := g.settings.Client.Do(outgoing)
	if err != nil {
		g.record(caller, 0, 0, http.StatusBadGateway)
		refuse(w, http.StatusBadGateway, "provider", "the model provider could not be reached")
		return
	}
	defer answer.Body.Close()

	if streamed && answer.StatusCode == http.StatusOK {
		g.stream(w, answer, caller)
		return
	}

	data, _ := io.ReadAll(io.LimitReader(answer.Body, 64<<20))
	var counted usageBody
	json.Unmarshal(data, &counted)
	in, out := int64(0), int64(0)
	if counted.Usage != nil {
		in, out = counted.Usage.PromptTokens, counted.Usage.CompletionTokens
	}
	g.record(caller, in, out, answer.StatusCode)

	w.Header().Set("Content-Type", answer.Header.Get("Content-Type"))
	w.WriteHeader(answer.StatusCode)
	w.Write(data)
}

func (g *Gateway) stream(w http.ResponseWriter, answer *http.Response, caller Caller) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	flusher, _ := w.(http.Flusher)

	in, out := int64(0), int64(0)
	reader := bufio.NewReaderSize(answer.Body, 64<<10)
	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			if payload, found := bytes.CutPrefix(bytes.TrimSpace(line), []byte("data:")); found {
				var counted usageBody
				if json.Unmarshal(bytes.TrimSpace(payload), &counted) == nil && counted.Usage != nil {
					in, out = counted.Usage.PromptTokens, counted.Usage.CompletionTokens
				}
			}
			w.Write(line)
			if flusher != nil {
				flusher.Flush()
			}
		}
		if err != nil {
			break
		}
	}
	g.record(caller, in, out, http.StatusOK)
}

func (g *Gateway) record(caller Caller, in, out int64, status int) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	g.settings.Logger.Info("ai call", "model", g.settings.Model, "tokens_in", in, "tokens_out", out, "status", status, "project", caller.Project.String())
	if err := g.store.Record(ctx, Usage{Caller: caller, Model: g.settings.Model, TokensIn: in, TokensOut: out, Status: status}); err != nil {
		g.settings.Logger.Warn("recording AI use failed", "error", err)
	}
}

type Books struct {
	store Store
	now   func() time.Time
}

var ErrBelowOne = errors.New("aigateway: a limit is at least one")

func NewBooks(store Store, now func() time.Time) *Books {
	if now == nil {
		now = time.Now
	}
	return &Books{store: store, now: now}
}

func (b *Books) Use(ctx context.Context, organisation uuid.UUID) (Limits, int64, int64, error) {
	limits, err := b.store.Limits(ctx, organisation)
	if err != nil {
		return Limits{}, 0, 0, err
	}
	requests, tokens, err := b.store.UsedSince(ctx, organisation, monthStart(b.now()))
	return limits, requests, tokens, err
}

func (b *Books) SetLimits(ctx context.Context, organisation uuid.UUID, limits Limits) error {
	if (limits.RequestsPerMinute != nil && *limits.RequestsPerMinute < 1) || (limits.TokensPerMonth != nil && *limits.TokensPerMonth < 1) {
		return ErrBelowOne
	}
	return b.store.SetLimits(ctx, organisation, limits)
}

func (t *Tokens) Caller(ctx context.Context, token string) (Caller, bool, error) {
	if token == "" {
		return Caller{}, false, nil
	}
	return t.store.LookUp(ctx, hash(token))
}

func (g *Gateway) FromAWorkspace(r *http.Request) bool {
	return g.from(r)
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

var (
	ErrLimited    = errors.New("aigateway: the organisation's limit is reached")
	ErrNoAnswer   = errors.New("aigateway: the model did not answer")
	ErrNotEnabled = errors.New("aigateway: no model is configured")
)

func (g *Gateway) Ask(ctx context.Context, caller Caller, messages []Message) (string, error) {
	if g == nil || g.settings.ProviderURL == "" {
		return "", ErrNotEnabled
	}
	if status, _, reason := g.admit(ctx, caller.Organisation); status != 0 {
		return "", fmt.Errorf("%w: %s", ErrLimited, reason)
	}
	body, _ := json.Marshal(map[string]any{"model": g.settings.Model, "messages": messages})
	ctx, cancel := context.WithTimeout(ctx, providerTimeout)
	defer cancel()
	outgoing, err := http.NewRequestWithContext(ctx, http.MethodPost, g.settings.ProviderURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	outgoing.Header.Set("Content-Type", "application/json")
	outgoing.Header.Set("Authorization", "Bearer "+g.settings.Key)
	outgoing.Header.Set("User-Agent", g.settings.userAgent())
	if g.settings.SessionHeader != "" {
		outgoing.Header.Set(g.settings.SessionHeader, sessionOf(caller))
	}
	answer, err := g.settings.Client.Do(outgoing)
	if err != nil {
		g.record(caller, 0, 0, http.StatusBadGateway)
		return "", fmt.Errorf("%w: %v", ErrNoAnswer, err)
	}
	defer answer.Body.Close()
	var said struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		usageBody
	}
	decodeErr := json.NewDecoder(io.LimitReader(answer.Body, maxBody)).Decode(&said)
	var in, out int64
	if said.Usage != nil {
		in, out = said.Usage.PromptTokens, said.Usage.CompletionTokens
	}
	g.record(caller, in, out, answer.StatusCode)
	if answer.StatusCode != http.StatusOK || decodeErr != nil || len(said.Choices) == 0 || strings.TrimSpace(said.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("%w: it answered %d", ErrNoAnswer, answer.StatusCode)
	}
	return said.Choices[0].Message.Content, nil
}
