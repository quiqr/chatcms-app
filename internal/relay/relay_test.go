package relay

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func shortDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "rl")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func echo(t *testing.T, network, address string) {
	t.Helper()
	listener, err := net.Listen(network, address)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { listener.Close() })
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func() {
				line, _ := bufio.NewReader(conn).ReadString('\n')
				fmt.Fprintf(conn, "heard %s", line)
				conn.Close()
			}()
		}
	}()
}

func ask(t *testing.T, network, address string) string {
	t.Helper()
	conn, err := net.Dial(network, address)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	fmt.Fprint(conn, "hello\n")
	answer, _ := io.ReadAll(conn)
	return string(answer)
}

func TestASocketIsJoinedToAPortBothWays(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dir := shortDir(t)

	server, _ := net.Listen("tcp", "127.0.0.1:0")
	port := server.Addr().String()
	server.Close()
	echo(t, "tcp", port)
	if _, err := Forward(ctx, "unix", filepath.Join(dir, "out.sock"), "tcp", port); err != nil {
		t.Fatal(err)
	}
	if got := ask(t, "unix", filepath.Join(dir, "out.sock")); got != "heard hello\n" {
		t.Fatalf("through the socket: %q", got)
	}

	echo(t, "unix", filepath.Join(dir, "in.sock"))
	listener, err := Forward(ctx, "tcp", "127.0.0.1:0", "unix", filepath.Join(dir, "in.sock"))
	if err != nil {
		t.Fatal(err)
	}
	if got := ask(t, "tcp", listener.Addr().String()); got != "heard hello\n" {
		t.Fatalf("through the port: %q", got)
	}
}

func TestARunIsReadAsTheMicroVMWritesIt(t *testing.T) {
	data, _ := json.Marshal(map[string]any{"cmd": []string{"/nix/store/x-hugo/bin/hugo", "server"}, "path": []string{"/nix/store/x/bin"}, "env": map[string]string{"A": "b"}})
	run, err := DecodeRun(base64.RawURLEncoding.EncodeToString(data))
	if err != nil {
		t.Fatal(err)
	}
	if run.Cmd[1] != "server" || run.Path[0] != "/nix/store/x/bin" || run.Env["A"] != "b" {
		t.Fatalf("run = %+v", run)
	}
	if _, err := DecodeRun(base64.RawURLEncoding.EncodeToString([]byte(`{"cmd":[]}`))); err == nil {
		t.Fatal("a run without a command was taken")
	}
}

func TestTheAgentsSettingsAreReadAsTheGatewayWritesThem(t *testing.T) {
	got := parseEnv("HOME='/run/agent'\nOPENCODE_CONFIG_CONTENT='{\"a\":\"it'\\''s\"}'\n\n# a note\n")
	want := []string{"HOME=/run/agent", `OPENCODE_CONFIG_CONTENT={"a":"it's"}`}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("parseEnv = %q", got)
	}
}

func TestLaunchServesThePreviewAndStopsWhenItEnds(t *testing.T) {
	shell, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("no shell")
	}
	dir := shortDir(t)
	server, _ := net.Listen("tcp", "127.0.0.1:0")
	port := server.Addr().(*net.TCPAddr).Port
	server.Close()
	echo(t, "tcp", fmt.Sprintf("127.0.0.1:%d", port))

	sleep, err := exec.LookPath("sleep")
	if err != nil {
		t.Skip("no sleep")
	}
	data, _ := json.Marshal(map[string]any{"cmd": []string{shell, "-c", "sleep 1"}, "path": []string{filepath.Dir(shell), filepath.Dir(sleep)}})
	plan := Plan{Run: base64.RawURLEncoding.EncodeToString(data), Site: dir, Home: dir, Sockets: dir, PreviewPort: port}
	done := make(chan error, 1)
	go func() { done <- Launch(context.Background(), plan, io.Discard) }()

	for i := 0; i < 50; i++ {
		if _, err := os.Stat(filepath.Join(dir, "preview.sock")); err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if got := ask(t, "unix", filepath.Join(dir, "preview.sock")); got != "heard hello\n" {
		t.Fatalf("the preview through its socket: %q", got)
	}
	select {
	case err := <-done:
		if err == nil || !strings.Contains(err.Error(), "the preview ended") {
			t.Fatalf("Launch() = %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Launch did not stop when the preview ended")
	}
}

func TestTheAgentGetsItsSettingsThroughTheGatewaySocket(t *testing.T) {
	dir := shortDir(t)
	gateway, err := net.Listen("unix", filepath.Join(dir, "ai.sock"))
	if err != nil {
		t.Fatal(err)
	}
	asked := make(chan string, 1)
	go http.Serve(gateway, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		asked <- r.Header.Get("Authorization") + " " + r.URL.Path
		fmt.Fprint(w, "HOME='/run/agent'\n")
	}))
	defer gateway.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	inside, _ := net.Listen("tcp", "127.0.0.1:0")
	port := inside.Addr().(*net.TCPAddr).Port
	inside.Close()
	if _, err := Forward(ctx, "tcp", fmt.Sprintf("127.0.0.1:%d", port), "unix", filepath.Join(dir, "ai.sock")); err != nil {
		t.Fatal(err)
	}
	env, err := agentEnvironment(ctx, Plan{GatewayPort: port, Token: "tok"}, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(env, "|") != "HOME=/run/agent" || <-asked != "Bearer tok /agent/env" {
		t.Fatalf("env = %q", env)
	}
}
