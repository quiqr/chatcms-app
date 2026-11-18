//go:build microvm

package microvm

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/aigateway"
	"github.com/quiqr/chatcms-app/internal/assistant"
	"github.com/quiqr/chatcms-app/internal/workspace"
)

type acceptAny struct{}

func (acceptAny) Caller(ctx context.Context, token string) (aigateway.Caller, bool, error) {
	return aigateway.Caller{}, token != "", nil
}

func needs(t *testing.T, name string) string {
	t.Helper()
	value := os.Getenv(name)
	if value == "" {
		t.Skipf("%s is not set; see scripts/dev-taps.sh", name)
	}
	return value
}

func TestARealMicroVMServesThePersonsSite(t *testing.T) {
	guest, err := LoadGuest(needs(t, "CHATCMS_WORKSPACE_GUEST"))
	if err != nil {
		t.Fatal(err)
	}
	backend, err := New(Settings{
		CloudHypervisor: needs(t, "CHATCMS_CLOUD_HYPERVISOR_BIN"),
		Virtiofsd:       needs(t, "CHATCMS_VIRTIOFSD_BIN"),
		Guest:           guest,
		RunDir:          filepath.Join("/tmp", "chatcms-w-"+uuid.NewString()[:8]),
		Taps:            1,
		HugoBin:         needs(t, "CHATCMS_HUGO_BIN"),
		Ready:           2 * time.Minute,
		GatewayPort:     8091,
	})
	if err != nil {
		t.Fatal(err)
	}

	site := filepath.Join(t.TempDir(), "working-copy")
	for name, content := range map[string]string{
		"hugo.toml":          "baseURL = \"https://example.org/\"\ntitle = \"Korenaar\"\n",
		"content/_index.md":  "---\ntitle: Korenaar in a VM\n---\n\nHello from inside.\n",
		"layouts/index.html": "<!doctype html><html><body><h1>{{ .Title }}</h1>{{ .Content }}</body></html>\n",
	} {
		os.MkdirAll(filepath.Dir(filepath.Join(site, name)), 0o755)
		os.WriteFile(filepath.Join(site, name), []byte(content), 0o644)
	}

	gateway, err := net.Listen("tcp", "10.231.0.1:8091")
	if err != nil {
		t.Skipf("no tap device to stand a gateway on: %v", err)
	}
	served := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/agent/env" {
			assistant.NewAgentEnvironment(acceptAny{}, "/run/agent").ServeHTTP(w, r)
			return
		}
		http.NotFound(w, r)
	})}
	go served.Serve(gateway)
	defer served.Close()

	started := time.Now()
	handle, endpoint, err := backend.Start(t.Context(), workspace.Spec{
		Key: workspace.Key{Person: uuid.New(), Project: uuid.New()}, Dir: site,
		Under: "http://127.0.0.1:8080/preview/o/p/",
		Token: "a-boot-test-token",
	}, os.Stderr)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	t.Logf("cold start answered after %v", time.Since(started))
	saved := time.Time{}
	defer backend.Stop(context.Background(), handle)

	client := &http.Client{Timeout: 2 * time.Second}
	fetch := func() string {
		answer, err := client.Get("http://" + endpoint.Address() + "/preview/o/p/")
		if err != nil {
			return ""
		}
		defer answer.Body.Close()
		body, _ := io.ReadAll(answer.Body)
		return string(body)
	}
	if body := fetch(); !strings.Contains(body, "Korenaar in a VM") {
		t.Fatalf("the VM served %q", body)
	}

	os.WriteFile(filepath.Join(site, "content/_index.md"), []byte("---\ntitle: Korenaar after a save\n---\n"), 0o644)
	saved = time.Now()
	deadline := time.Now().Add(20 * time.Second)
	for !strings.Contains(fetch(), "Korenaar after a save") {
		if time.Now().After(deadline) {
			t.Fatalf("a save did not show in the VM's preview within 20 seconds")
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Logf("a save showed after %v", time.Since(saved))

	agent := "http://" + workspace.Endpoint{Host: endpoint.Host, Port: endpoint.AgentPort}.Address() + "/session"
	reached := false
	agentDeadline := time.Now().Add(90 * time.Second)
	for time.Now().Before(agentDeadline) {
		request, _ := http.NewRequest(http.MethodGet, agent, nil)
		request.SetBasicAuth("opencode", "a-boot-test-token")
		if answer, err := client.Do(request); err == nil {
			answer.Body.Close()
			if answer.StatusCode == http.StatusOK {
				reached = true
				break
			}
		}
		time.Sleep(time.Second)
	}
	if !reached {
		t.Errorf("the agent in the VM did not answer on %s", agent)
		t.Log(lastLines(filepath.Join(backend.settings.RunDir, "0", "console.log"), 2000))
	} else {
		t.Logf("the agent answered after %v", time.Since(started))
		if answer, err := client.Get(agent); err == nil {
			answer.Body.Close()
			if answer.StatusCode != http.StatusUnauthorized {
				t.Errorf("the agent answered %d without its secret", answer.StatusCode)
				t.Log(lastLines(filepath.Join(backend.settings.RunDir, "0", "console.log"), 2000))
			}
		}
	}

	if err := backend.Stop(t.Context(), handle); err != nil {
		t.Fatal(err)
	}
	if fetch() != "" {
		t.Errorf("the VM still answers after Stop()")
	}
}
