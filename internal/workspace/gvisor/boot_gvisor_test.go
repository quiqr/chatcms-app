package gvisor

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/quiqr/chatcms-app/internal/runtime"
	"github.com/quiqr/chatcms-app/internal/workspace"
)

const probe = `
import http.server, socket, sys
host_port = int(sys.argv[1])
def reach(address):
    try:
        socket.create_connection(address, timeout=2).close()
        return "reached"
    except Exception:
        return "blocked"
class Answer(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        body = ("internet " + reach(("1.1.1.1", 80)) + "; host " + reach(("127.0.0.1", host_port)) + "; root " + ",".join(sorted(__import__("os").listdir("/")))).encode()
        self.send_response(200); self.end_headers(); self.wfile.write(body)
    def do_HEAD(self):
        self.send_response(200); self.end_headers()
http.server.HTTPServer(("127.0.0.1", 8080), Answer).serve_forever()
`

func TestAWorkspaceRunsInAGVisorSandboxThatReachesNothingButItsSockets(t *testing.T) {
	runsc, relayBin := os.Getenv("CHATCMS_RUNSC"), os.Getenv("CHATCMS_RELAY")
	if runsc == "" || relayBin == "" {
		t.Skip("CHATCMS_RUNSC and CHATCMS_RELAY name a runsc and a relay in the store")
	}
	python := inStore(t, "python3")

	hostService, _ := net.Listen("tcp", "127.0.0.1:0")
	defer hostService.Close()
	hostPort := hostService.Addr().(*net.TCPAddr).Port

	site := t.TempDir()
	os.WriteFile(filepath.Join(site, "probe.py"), []byte(probe), 0o644)
	state, err := os.MkdirTemp("", "gv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(state)

	b, err := New(Settings{Runsc: runsc, Relay: relayBin, State: state, Ready: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	spec := workspace.Spec{
		Dir:   site,
		Under: "http://127.0.0.1:9/",
		Run: runtime.Resolved{
			Preview: runtime.Command{Program: python, Args: []string{"/site/probe.py", strconv.Itoa(hostPort)}},
			Path:    []string{filepath.Dir(python)},
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	handle, endpoint, err := b.Start(ctx, spec, os.Stderr)
	if err != nil {
		t.Fatalf("Start() = %v", err)
	}

	answer, err := http.Get("http://" + endpoint.Address() + "/")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(answer.Body)
	answer.Body.Close()
	said := string(body)
	if !strings.Contains(said, "internet blocked") || !strings.Contains(said, "host blocked") {
		t.Fatalf("the sandbox reached out: %s", said)
	}
	for _, unexpected := range []string{"home", "etc", "var", "root"} {
		if strings.Contains(said, ","+unexpected+",") || strings.HasSuffix(said, ","+unexpected) {
			t.Fatalf("the sandbox sees %s: %s", unexpected, said)
		}
	}
	if _, alive := b.Alive(ctx, handle); !alive {
		t.Fatal("a running sandbox is not alive")
	}

	if err := b.Stop(ctx, handle); err != nil {
		t.Fatal(err)
	}
	if _, alive := b.Alive(ctx, handle); alive {
		t.Fatal("a stopped sandbox is alive")
	}
	if left, _ := os.ReadDir(filepath.Join(state, "s")); len(left) != 0 {
		t.Fatalf("sockets were left: %v", left)
	}
	if out, _ := b.runsc("list", "-quiet").Output(); strings.TrimSpace(string(out)) != "" {
		t.Fatalf("runsc still lists %s", out)
	}
}

func TestTheAgentInASandboxReachesTheGatewayThroughItsSocket(t *testing.T) {
	runsc, relayBin, opencode := os.Getenv("CHATCMS_RUNSC"), os.Getenv("CHATCMS_RELAY"), os.Getenv("CHATCMS_OPENCODE_BIN")
	if runsc == "" || relayBin == "" || opencode == "" {
		t.Skip("CHATCMS_RUNSC, CHATCMS_RELAY and CHATCMS_OPENCODE_BIN name a runsc, a relay and OpenCode in the store")
	}
	python := inStore(t, "python3")
	asked := make(chan string, 4)
	gateway := http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		asked <- r.URL.Path + " " + r.Header.Get("Authorization")
		if r.URL.Path == "/agent/env" {
			fmt.Fprint(w, "HOME='/run/agent'\nOPENCODE_SERVER_PASSWORD='tok'\n")
		}
	})}
	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	go gateway.Serve(listener)
	defer gateway.Close()

	site := t.TempDir()
	os.WriteFile(filepath.Join(site, "probe.py"), []byte(probe), 0o644)
	state, err := os.MkdirTemp("", "gv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(state)
	b, err := New(Settings{Runsc: runsc, Relay: relayBin, OpenCode: opencode, State: state, Gateway: listener.Addr().String(), Ready: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	spec := workspace.Spec{
		Dir:   site,
		Under: "http://127.0.0.1:9/",
		Token: "tok",
		Run: runtime.Resolved{
			Preview: runtime.Command{Program: python, Args: []string{"/site/probe.py", "9"}},
			Path:    []string{filepath.Dir(python)},
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	handle, endpoint, err := b.Start(ctx, spec, os.Stderr)
	if err != nil {
		t.Fatalf("Start() = %v", err)
	}
	defer b.Stop(ctx, handle)
	if got := <-asked; got != "/agent/env Bearer tok" {
		t.Fatalf("the gateway was asked %q", got)
	}
	if endpoint.AgentPort == 0 {
		t.Fatal("no agent port")
	}
	deadline := time.Now().Add(time.Minute)
	for {
		answer, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/doc", endpoint.AgentPort))
		if err == nil {
			answer.Body.Close()
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("the agent never answered: %v", err)
		}
		time.Sleep(time.Second)
	}
}
