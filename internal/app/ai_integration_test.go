//go:build integration

package app_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/aigateway"
	"github.com/quiqr/chatcms-app/internal/store"
	"github.com/quiqr/chatcms-app/internal/workspace"
)

func TestTheServerAnswersOnTheGatewaysOwnListener(t *testing.T) {
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer sk-the-real-one" {
			http.Error(w, "wrong key", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"choices":[{"message":{"content":"Hello from the provider"}}],"usage":{"prompt_tokens":3,"completion_tokens":4}}`)
	}))
	defer provider.Close()

	cfg := testConfig(t)
	cfg.AIListen = "127.0.0.1:0"
	cfg.AIProviderURL = provider.URL
	cfg.AIKey = "sk-the-real-one"
	cfg.AIModel = "deepseek-chat"
	server, _ := startServer(t, cfg)
	if server.AIAddr() == "" || server.AIAddr() == server.Addr() {
		t.Fatalf("the gateway listens on %q, want a listener of its own", server.AIAddr())
	}
	waitForListener(t, server.AIAddr())

	db, err := store.Open(t.Context(), cfg.DatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	organisation, project, person := uuid.New(), uuid.New(), uuid.New()
	for _, statement := range []struct {
		sql  string
		args []any
	}{
		{`INSERT INTO organisations (id, slug, name) VALUES ($1, $2, 'Studio')`, []any{organisation, "studio-" + organisation.String()[:8]}},
		{`INSERT INTO projects (id, organisation_id, slug, name) VALUES ($1, $2, 'site', 'Site')`, []any{project, organisation}},
		{`INSERT INTO people (id, email) VALUES ($1, $2)`, []any{person, person.String() + "@example.org"}},
	} {
		if _, err := db.Pool().Exec(t.Context(), statement.sql, statement.args...); err != nil {
			t.Fatal(err)
		}
	}
	token, err := aigateway.NewTokens(aigateway.NewPostgres(db.Queries)).Issue(t.Context(), workspace.Key{Person: person, Project: project})
	if err != nil {
		t.Fatal(err)
	}

	ask := func(token string) (int, string) {
		request, _ := http.NewRequest(http.MethodPost, "http://"+server.AIAddr()+"/v1/chat/completions", strings.NewReader(`{"messages":[]}`))
		if token != "" {
			request.Header.Set("Authorization", "Bearer "+token)
		}
		answer, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		defer answer.Body.Close()
		body, _ := io.ReadAll(answer.Body)
		return answer.StatusCode, string(body)
	}

	if status, body := ask(token); status != http.StatusOK || !strings.Contains(body, "Hello from the provider") {
		t.Errorf("with a token = %d %s", status, body)
	}
	if status, _ := ask(""); status != http.StatusUnauthorized {
		t.Errorf("without a token = %d", status)
	}

	answer, err := http.Post("http://"+server.Addr()+"/v1/chat/completions", "application/json", strings.NewReader(`{}`))
	if err == nil {
		answer.Body.Close()
		if answer.StatusCode == http.StatusOK {
			t.Errorf("the app's own address answers the gateway route")
		}
	}

	var used int64
	db.Pool().QueryRow(t.Context(), `SELECT coalesce(sum(tokens_in + tokens_out), 0) FROM ai_usage WHERE organisation_id = $1`, organisation).Scan(&used)
	if used != 7 {
		t.Errorf("recorded %d tokens, want 7", used)
	}
}
