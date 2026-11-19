package local

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/runtime"
	"github.com/quiqr/chatcms-app/internal/workspace"
)

func hugoBin(t *testing.T) string {
	t.Helper()

	if named := os.Getenv("CHATCMS_HUGO_BIN"); named != "" {
		return named
	}
	found, err := exec.LookPath("hugo")
	if err != nil {
		t.Skipf("no Hugo on this machine: %v", err)
	}
	return found
}

func aSite(t *testing.T, title string, broken bool) string {
	t.Helper()

	root := t.TempDir()
	dir := filepath.Join(root, "working-copy")
	layout := "<!doctype html><html><body><h1>{{ .Title }}</h1>{{ .Content }}</body></html>\n"
	config := "baseURL = \"https://example.org/\"\ntitle = \"" + title + "\"\n"
	if broken {
		config = "this is not = [valid\n"
	}
	files := map[string]string{
		"hugo.toml":          config,
		"content/_index.md":  "---\ntitle: " + title + "\n---\n\nThe body of " + title + ".\n",
		"layouts/index.html": layout,
	}
	for name, content := range files {
		where := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(where), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(where, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (l *lockedBuffer) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.buf.Write(p)
}

func backend(t *testing.T) *Backend {
	t.Helper()

	made, err := New(Settings{HugoBin: hugoBin(t), Ready: 60 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	return made
}

func start(t *testing.T, b *Backend, dir string, log io.Writer) (workspace.Handle, workspace.Endpoint) {
	t.Helper()

	handle, endpoint, err := b.Start(t.Context(), workspace.Spec{
		Key:   workspace.Key{Person: uuid.New(), Project: uuid.New()},
		Dir:   dir,
		Under: "http://127.0.0.1:8080/preview/o/p/",
	}, log)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	t.Cleanup(func() { b.Stop(context.Background(), handle) })
	return handle, endpoint
}

func TestALocalWorkspaceServesTheSite(t *testing.T) {
	b := backend(t)
	log := &lockedBuffer{}
	dir := aSite(t, "Korenaar", false)
	handle, endpoint := start(t, b, dir, log)

	response, err := http.Get("http://" + endpoint.Address() + "/preview/o/p/")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	if !strings.Contains(string(body), "Korenaar") {
		t.Errorf("served %q, want the site", body)
	}
	if found, alive := b.Alive(t.Context(), handle); !alive || found != endpoint {
		t.Errorf("Alive() = %v, %v, want the running one", found, alive)
	}
	if _, err := os.Stat(filepath.Join(dir, ".cache")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Hugo wrote into the working copy")
	}
}

func TestALocalWorkspaceEndsWithTheApp(t *testing.T) {
	b := backend(t)
	if b.Apart() {
		t.Errorf("Apart() = true, want the local backend's workspaces to end with the app")
	}
	handle, _ := start(t, b, aSite(t, "Korenaar", false), io.Discard)

	if err := b.Stop(t.Context(), handle); err != nil {
		t.Fatal(err)
	}
	if err := b.Stop(t.Context(), handle); err != nil {
		t.Errorf("stopping twice error = %v, want nil", err)
	}
	if _, alive := b.Alive(t.Context(), handle); alive {
		t.Errorf("Alive() after Stop() = true")
	}
}

func TestAnotherProcessCanReattachAndStop(t *testing.T) {
	first := backend(t)
	handle, _ := start(t, first, aSite(t, "Korenaar", false), io.Discard)

	second := backend(t)
	if _, alive := second.Alive(t.Context(), handle); !alive {
		t.Fatalf("Alive() from a second backend = false, want the running Hugo found")
	}
	if err := second.Stop(t.Context(), handle); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if _, alive := first.Alive(t.Context(), handle); !alive {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Errorf("a reattached Hugo was not stopped")
}

func TestLocalAliveRefusesAReusedPid(t *testing.T) {
	b := backend(t)
	reused := handleOf(os.Getpid(), 1, hugoBin(t))
	if _, alive := b.Alive(t.Context(), reused); alive {
		t.Errorf("Alive() = true for a pid that is not Hugo")
	}
	if err := b.Stop(t.Context(), reused); err != nil {
		t.Errorf("Stop() of a reused pid error = %v", err)
	}
}

func TestASiteThatWillNotBuildIsNotReady(t *testing.T) {
	b := backend(t)
	log := &lockedBuffer{}
	_, _, err := b.Start(t.Context(), workspace.Spec{Dir: aSite(t, "Broken", true), Under: "http://127.0.0.1/"}, log)
	if !errors.Is(err, workspace.ErrNotReady) {
		t.Fatalf("Start() error = %v, want ErrNotReady", err)
	}
	log.mu.Lock()
	defer log.mu.Unlock()
	if log.buf.Len() == 0 {
		t.Errorf("Hugo's complaint did not reach the log")
	}
}

func TestNewNeedsHugo(t *testing.T) {
	if _, err := New(Settings{}); !errors.Is(err, ErrNoHugo) {
		t.Errorf("New() error = %v, want ErrNoHugo", err)
	}
}

func TestHandlesRoundTripAndRefuseNonsense(t *testing.T) {
	pid, port, hugo, ok := parseHandle(handleOf(4242, 1313, "/nix/store/x/bin/hugo"))
	if !ok || pid != 4242 || port != 1313 || hugo != "/nix/store/x/bin/hugo" {
		t.Errorf("parseHandle() = %d, %d, %q, %v", pid, port, hugo, ok)
	}
	for _, bad := range []string{"", "1:2", "x:1:h", "1:x:h", "1:2:", "1:2:h", fmt.Sprintf("%d:0:h", 5)} {
		if _, _, _, ok := parseHandle(workspace.Handle(bad)); ok {
			t.Errorf("parseHandle(%q) accepted", bad)
		}
	}
}

func TestBrowserPort(t *testing.T) {
	for under, want := range map[string]int{
		"http://localhost:5173/preview/": 5173,
		"https://chatcms.eu/preview/":    443,
		"http://chatcms.eu/":             80,
		"ftp://x/":                       0,
		"http://x:nope/":                 0,
		"::":                             0,
	} {
		if got := workspace.BrowserPort(under); got != want {
			t.Errorf("workspace.BrowserPort(%q) = %d, want %d", under, got, want)
		}
	}
}

func TestALocalWorkspaceRunsTheProjectsHugo(t *testing.T) {
	made, err := New(Settings{HugoBin: "/nonexistent/hugo", Ready: 60 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	run, err := runtime.Resolver{InstalledHugo: hugoBin(t)}.Resolve(t.Context(), runtime.Default(""), false)
	if err != nil {
		t.Fatal(err)
	}
	handle, _, err := made.Start(t.Context(), workspace.Spec{Dir: aSite(t, "Korenaar", false), Under: "http://127.0.0.1/", Run: run}, io.Discard)
	if err != nil {
		t.Fatalf("Start() with the project's Hugo error = %v", err)
	}
	defer made.Stop(context.Background(), handle)
	if _, _, hugo, _ := parseHandle(handle); hugo != hugoBin(t) {
		t.Errorf("handle names %q, want the Hugo it runs", hugo)
	}
}

func TestALocalWorkspaceStartsItsAgentWhenItHasOne(t *testing.T) {
	stubDir := t.TempDir()
	told := filepath.Join(stubDir, "told")
	stub := filepath.Join(stubDir, "opencode")
	os.WriteFile(stub, []byte("#!/bin/sh\n{ echo \"$PWD $*\"; export -p; } > "+told+"\n"), 0o755)

	made, err := New(Settings{HugoBin: hugoBin(t), Ready: 60 * time.Second, OpenCodeBin: stub, Gateway: "http://127.0.0.1:8091"})
	if err != nil {
		t.Fatal(err)
	}
	dir := aSite(t, "Korenaar", false)
	handle, endpoint, err := made.Start(t.Context(), workspace.Spec{Dir: dir, Under: "http://127.0.0.1/", Token: "tok"}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if endpoint.AgentPort == 0 || endpoint.AgentPort == endpoint.Port {
		t.Errorf("AgentPort = %d", endpoint.AgentPort)
	}

	deadline := time.Now().Add(5 * time.Second)
	var said []byte
	for time.Now().Before(deadline) {
		if said, _ = os.ReadFile(told); strings.Contains(string(said), "TZ") {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	for _, want := range []string{dir + " serve --hostname 127.0.0.1 --port", "OPENCODE_SERVER_PASSWORD=", "tok", "OPENCODE_DISABLE_PROJECT_CONFIG=", "http://127.0.0.1:8091/v1"} {
		if !strings.Contains(string(said), want) {
			t.Errorf("the agent was not told %q", want)
		}
	}

	made.Stop(t.Context(), handle)
	if conn, err := net.DialTimeout("tcp", endpoint.Address(), 200*time.Millisecond); err == nil {
		conn.Close()
		t.Errorf("the preview still answers after Stop()")
	}

	without, _ := New(Settings{HugoBin: hugoBin(t), Ready: 60 * time.Second, OpenCodeBin: stub, Gateway: "http://127.0.0.1:8091"})
	other, plain, err := without.Start(t.Context(), workspace.Spec{Dir: aSite(t, "Plain", false), Under: "http://127.0.0.1/"}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	defer without.Stop(context.Background(), other)
	if plain.AgentPort != 0 {
		t.Errorf("a workspace without a token was given an agent")
	}
}

func TestALocalWorkspaceRunsTheDeclaredPreviewCommand(t *testing.T) {
	definition, err := runtime.Parse([]byte("packages: {hugo: \"0.139.0\"}\npreview:\n  command: [hugo, server, --source, \"{source}\", --bind, \"{host}\", --port, \"{port}\", --baseURL, \"{url}\", --appendPort=false, --renderToMemory]\n"))
	if err != nil {
		t.Fatal(err)
	}
	run, err := runtime.Resolver{InstalledHugo: hugoBin(t)}.Resolve(t.Context(), definition, true)
	if err != nil {
		t.Fatal(err)
	}
	made := backend(t)
	handle, endpoint, err := made.Start(t.Context(), workspace.Spec{Dir: aSite(t, "Declared", false), Under: "http://127.0.0.1:8080/preview/o/p/", Run: run}, io.Discard)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer made.Stop(context.Background(), handle)
	answer, err := http.Get("http://" + endpoint.Address() + "/preview/o/p/")
	if err != nil {
		t.Fatal(err)
	}
	defer answer.Body.Close()
	body, _ := io.ReadAll(answer.Body)
	if !strings.Contains(string(body), "Declared") {
		t.Errorf("served %q", body)
	}
}

func TestALocalPreviewIsGivenItsCellEnvironment(t *testing.T) {
	stubDir := t.TempDir()
	told := filepath.Join(stubDir, "told")
	stub := filepath.Join(stubDir, "quarto")
	os.WriteFile(stub, []byte("#!/bin/sh\nexport -p > "+told+"\nexec sleep 30\n"), 0o755)

	made, err := New(Settings{HugoBin: hugoBin(t), Ready: 2 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	dir := aSite(t, "Korenaar", false)
	run := runtime.Resolved{
		Path:    []string{stubDir},
		Preview: runtime.Command{Program: stub, Args: []string{"preview"}},
		Env:     map[string]string{"QUARTO_PYTHON": "/nix/store/py/bin/python3"},
	}
	made.Start(t.Context(), workspace.Spec{Dir: dir, Under: "http://127.0.0.1/", Run: run}, io.Discard)

	deadline := time.Now().Add(5 * time.Second)
	var said []byte
	for time.Now().Before(deadline) {
		if said, _ = os.ReadFile(told); strings.Contains(string(said), "QUARTO_PYTHON") {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if !strings.Contains(string(said), `QUARTO_PYTHON="/nix/store/py/bin/python3"`) && !strings.Contains(string(said), "QUARTO_PYTHON='/nix/store/py/bin/python3'") {
		t.Errorf("the preview was given %s", said)
	}
}
