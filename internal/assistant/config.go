package assistant

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
)

const (
	Provider = "chatcms"
	Model    = "chatcms"
	Prompt   = `You help a person edit the content of their website in Quiqr ChatCMS.
Change pages only with the quiqr tools: list_pages to find a page, read_page to see its fields and text, update_page to change them. Never try to write files or run commands; that is not possible here.
Answer briefly, in the language the person writes in. Never use the words commit, branch, push, pull, merge, repository or conflict.`
)

func OpenCodeConfig(gateway, token string) string {
	gateway = strings.TrimSuffix(gateway, "/")
	config := map[string]any{
		"$schema":    "https://opencode.ai/config.json",
		"autoupdate": false,
		"share":      "disabled",
		"snapshot":   false,
		"provider": map[string]any{
			Provider: map[string]any{
				"npm":  "@ai-sdk/openai-compatible",
				"name": "ChatCMS",
				"options": map[string]any{
					"baseURL": gateway + "/v1",
					"apiKey":  token,
				},
				"models": map[string]any{Model: map[string]any{"name": "ChatCMS"}},
			},
		},
		"model":       Provider + "/" + Model,
		"small_model": Provider + "/" + Model,
		"mcp": map[string]any{
			"quiqr": map[string]any{
				"type":    "remote",
				"url":     gateway + "/mcp",
				"enabled": true,
				"oauth":   false,
				"headers": map[string]string{"Authorization": "Bearer " + token},
			},
		},
		"permission": map[string]any{
			"read":               "allow",
			"glob":               "allow",
			"grep":               "allow",
			"list":               "allow",
			"todowrite":          "allow",
			"edit":               "deny",
			"bash":               "deny",
			"task":               "deny",
			"external_directory": "deny",
			"webfetch":           "deny",
			"websearch":          "deny",
			"repo_clone":         "deny",
			"repo_overview":      "deny",
			"lsp":                "deny",
			"skill":              "deny",
			"question":           "deny",
			"doom_loop":          "deny",
		},
		"agent": map[string]any{
			"build": map[string]any{"prompt": Prompt},
		},
	}
	data, _ := json.Marshal(config)
	return string(data)
}

func OpenCodeEnvironment(home, gateway, token string) []string {
	return []string{
		"HOME=" + home,
		"XDG_CONFIG_HOME=" + filepath.Join(home, "config"),
		"XDG_DATA_HOME=" + filepath.Join(home, "data"),
		"XDG_STATE_HOME=" + filepath.Join(home, "state"),
		"XDG_CACHE_HOME=" + filepath.Join(home, "cache"),
		"OPENCODE_CONFIG_CONTENT=" + OpenCodeConfig(gateway, token),
		"OPENCODE_DISABLE_PROJECT_CONFIG=1",
		"OPENCODE_DISABLE_CLAUDE_CODE=1",
		"OPENCODE_DISABLE_EXTERNAL_SKILLS=1",
		"OPENCODE_DISABLE_DEFAULT_PLUGINS=1",
		"OPENCODE_DISABLE_MODELS_FETCH=1",
		"OPENCODE_DISABLE_AUTOUPDATE=1",
		"OPENCODE_DISABLE_LSP_DOWNLOAD=1",
		"OPENCODE_DISABLE_SHARE=1",
		"OPENCODE_SERVER_PASSWORD=" + token,
		"TZ=UTC",
	}
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

type AgentEnvironment struct {
	callers Callers
	home    string
}

func NewAgentEnvironment(callers Callers, home string) *AgentEnvironment {
	return &AgentEnvironment{callers: callers, home: home}
}

func (a *AgentEnvironment) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	token, _ := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	_, known, err := a.callers.Caller(r.Context(), token)
	if err != nil {
		http.Error(w, "the gateway cannot check tokens right now", http.StatusServiceUnavailable)
		return
	}
	if !known {
		http.Error(w, "this token is not, or no longer, valid", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	for _, pair := range OpenCodeEnvironment(a.home, "http://"+r.Host, token) {
		name, value, _ := strings.Cut(pair, "=")
		fmt.Fprintf(w, "%s=%s\n", name, shellQuote(value))
	}
}
