package gvisor

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/quiqr/chatcms-app/internal/relay"
	"github.com/quiqr/chatcms-app/internal/workspace"
)

const (
	DefaultReady  = 2 * time.Minute
	InsidePreview = 8080
	InsideAgent   = 4096
	InsideGateway = 8091
	socketsWithin = 30 * time.Second
)

var ErrIncomplete = errors.New("gvisor: the backend is missing a part it needs")

type Settings struct {
	Runsc    string
	Relay    string
	OpenCode string
	HugoBin  string
	State    string
	Store    string
	Gateway  string
	Ready    time.Duration
	Command  func(name string, arguments ...string) *exec.Cmd
}

type sandbox struct {
	id      string
	dir     string
	sockets string
	preview int
	agent   int
	stop    context.CancelFunc
	runsc   *exec.Cmd
	exited  chan struct{}
}

type Backend struct {
	settings Settings
	mu       sync.Mutex
	running  map[string]*sandbox
}

func New(settings Settings) (*Backend, error) {
	if settings.Runsc == "" || settings.Relay == "" || settings.State == "" {
		return nil, ErrIncomplete
	}
	if settings.Store == "" {
		settings.Store = "/nix/store"
	}
	state, err := filepath.Abs(settings.State)
	if err != nil {
		return nil, fmt.Errorf("gvisor: %w", err)
	}
	settings.State = state
	if settings.Ready <= 0 {
		settings.Ready = DefaultReady
	}
	if settings.Command == nil {
		settings.Command = exec.Command
	}
	for _, dir := range []string{settings.State, filepath.Join(settings.State, "s"), filepath.Join(settings.State, "runsc")} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return nil, fmt.Errorf("gvisor: %w", err)
		}
	}
	b := &Backend{settings: settings, running: map[string]*sandbox{}}
	b.clearLeftovers()
	return b, nil
}

func (b *Backend) Name() string { return "gvisor" }

func (b *Backend) Apart() bool { return false }

func (b *Backend) runsc(arguments ...string) *exec.Cmd {
	base := []string{"--rootless", "--root=" + filepath.Join(b.settings.State, "runsc"), "--network=none", "--host-uds=all"}
	return b.settings.Command(b.settings.Runsc, append(base, arguments...)...)
}

func (b *Backend) clearLeftovers() {
	out, err := b.runsc("list", "-quiet").Output()
	if err == nil {
		for _, id := range strings.Fields(string(out)) {
			b.runsc("delete", "--force", id).Run()
		}
	}
	entries, _ := os.ReadDir(b.settings.State)
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "ws") {
			os.RemoveAll(filepath.Join(b.settings.State, entry.Name()))
		}
	}
	os.RemoveAll(filepath.Join(b.settings.State, "s"))
	os.MkdirAll(filepath.Join(b.settings.State, "s"), 0o700)
}

func newID() string {
	var raw [6]byte
	rand.Read(raw[:])
	return "ws" + hex.EncodeToString(raw[:])
}

func (b *Backend) Bundle(id string, spec workspace.Spec, run string) (map[string]any, error) {
	opencode := ""
	if spec.Token != "" && b.settings.OpenCode != "" {
		found, err := workspace.InStore(b.settings.OpenCode, b.settings.Store)
		if err != nil {
			return nil, err
		}
		opencode = found
	}
	relayBin, err := workspace.InStore(b.settings.Relay, b.settings.Store)
	if err != nil {
		return nil, err
	}
	env := []string{
		"CHATCMS_RUN=" + run,
		"CHATCMS_PREVIEW_PORT=" + strconv.Itoa(InsidePreview),
		"CHATCMS_AGENT_PORT=" + strconv.Itoa(InsideAgent),
		"CHATCMS_GATEWAY_PORT=" + strconv.Itoa(InsideGateway),
	}
	if spec.Token != "" {
		env = append(env, "CHATCMS_AI_TOKEN="+spec.Token, "CHATCMS_OPENCODE="+opencode)
	}
	sockets := b.socketsOf(id)
	return map[string]any{
		"ociVersion": "1.0.0",
		"hostname":   "workspace",
		"process": map[string]any{
			"terminal": false,
			"user":     map[string]any{"uid": 0, "gid": 0},
			"args":     []string{relayBin, "launch"},
			"env":      env,
			"cwd":      "/site",
		},
		"root": map[string]any{"path": "rootfs", "readonly": true},
		"mounts": []map[string]any{
			{"destination": "/nix/store", "type": "bind", "source": b.settings.Store, "options": []string{"rbind", "ro"}},
			{"destination": "/site", "type": "bind", "source": spec.Dir, "options": []string{"rbind", "rw"}},
			{"destination": "/run", "type": "tmpfs", "source": "tmpfs", "options": []string{"nosuid", "nodev"}},
			{"destination": "/run/chatcms", "type": "bind", "source": sockets, "options": []string{"rbind", "rw"}},
			{"destination": "/tmp", "type": "tmpfs", "source": "tmpfs", "options": []string{"nosuid", "nodev"}},
			{"destination": "/proc", "type": "proc", "source": "proc"},
		},
		"linux": map[string]any{
			"namespaces": []map[string]any{{"type": "pid"}, {"type": "ipc"}, {"type": "uts"}, {"type": "mount"}},
		},
	}, nil
}

func (b *Backend) socketsOf(id string) string {
	return filepath.Join(b.settings.State, "s", id)
}

func (b *Backend) Start(ctx context.Context, spec workspace.Spec, log io.Writer) (workspace.Handle, workspace.Endpoint, error) {
	run, err := workspace.EncodeRun(spec, workspace.RunOptions{
		HugoBin: b.settings.HugoBin,
		Store:   b.settings.Store,
		Host:    "127.0.0.1",
		Port:    InsidePreview,
		Source:  "/site",
		Cache:   "/run/preview/cache",
	})
	if err != nil {
		return "", workspace.Endpoint{}, err
	}
	site, err := filepath.Abs(spec.Dir)
	if err != nil {
		return "", workspace.Endpoint{}, fmt.Errorf("gvisor: %w", err)
	}
	spec.Dir = site
	id := newID()
	box := &sandbox{id: id, dir: filepath.Join(b.settings.State, id), sockets: b.socketsOf(id), exited: make(chan struct{})}
	relays, cancel := context.WithCancel(context.Background())
	box.stop = cancel
	fail := func(err error) (workspace.Handle, workspace.Endpoint, error) {
		if tail := lastLines(filepath.Join(box.dir, "console.log"), 20); tail != "" {
			fmt.Fprintln(log, tail)
		}
		b.end(box)
		return "", workspace.Endpoint{}, err
	}

	config, err := b.Bundle(id, spec, run)
	if err != nil {
		return fail(err)
	}
	bundle := filepath.Join(box.dir, "bundle")
	for _, dir := range []string{filepath.Join(bundle, "rootfs"), box.sockets} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return fail(fmt.Errorf("gvisor: %w", err))
		}
	}
	data, _ := json.MarshalIndent(config, "", " ")
	if err := os.WriteFile(filepath.Join(bundle, "config.json"), data, 0o600); err != nil {
		return fail(fmt.Errorf("gvisor: %w", err))
	}

	if spec.Token != "" && b.settings.Gateway != "" {
		if _, err := relay.Forward(relays, "unix", filepath.Join(box.sockets, "ai.sock"), "tcp", b.settings.Gateway); err != nil {
			return fail(fmt.Errorf("gvisor: %w", err))
		}
	}

	console, err := os.Create(filepath.Join(box.dir, "console.log"))
	if err != nil {
		return fail(fmt.Errorf("gvisor: %w", err))
	}
	box.runsc = b.runsc("run", "--bundle", bundle, id)
	box.runsc.Stdout = console
	box.runsc.Stderr = console
	fmt.Fprintf(log, "starting a gVisor sandbox %s\n", id)
	if err := box.runsc.Start(); err != nil {
		console.Close()
		return fail(fmt.Errorf("gvisor: %w", err))
	}
	go func() {
		box.runsc.Wait()
		console.Close()
		close(box.exited)
	}()

	b.mu.Lock()
	b.running[id] = box
	b.mu.Unlock()

	if err := waitForSocket(ctx, box, filepath.Join(box.sockets, "preview.sock"), socketsWithin); err != nil {
		return fail(err)
	}
	preview, err := relay.Forward(relays, "tcp", "127.0.0.1:0", "unix", filepath.Join(box.sockets, "preview.sock"))
	if err != nil {
		return fail(fmt.Errorf("gvisor: %w", err))
	}
	box.preview = preview.Addr().(*net.TCPAddr).Port
	endpoint := workspace.Endpoint{Host: "127.0.0.1", Port: box.preview}
	if spec.Token != "" && b.settings.OpenCode != "" {
		if err := waitForSocket(ctx, box, filepath.Join(box.sockets, "agent.sock"), socketsWithin); err != nil {
			return fail(err)
		}
		agent, err := relay.Forward(relays, "tcp", "127.0.0.1:0", "unix", filepath.Join(box.sockets, "agent.sock"))
		if err != nil {
			return fail(fmt.Errorf("gvisor: %w", err))
		}
		box.agent = agent.Addr().(*net.TCPAddr).Port
		endpoint.AgentPort = box.agent
	}
	if err := waitForAnswer(ctx, box, endpoint, b.settings.Ready); err != nil {
		return fail(err)
	}
	return handleOf(box), endpoint, nil
}

func waitForSocket(ctx context.Context, box *sandbox, path string, within time.Duration) error {
	deadline := time.Now().Add(within)
	for {
		if info, err := os.Stat(path); err == nil && info.Mode()&os.ModeSocket != 0 {
			return nil
		}
		select {
		case <-box.exited:
			return fmt.Errorf("%w: the sandbox ended before %s appeared", workspace.ErrNotReady, filepath.Base(path))
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%w: %s did not appear", workspace.ErrNotReady, filepath.Base(path))
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func waitForAnswer(ctx context.Context, box *sandbox, endpoint workspace.Endpoint, within time.Duration) error {
	deadline := time.Now().Add(within)
	for {
		connection, err := net.DialTimeout("tcp", endpoint.Address(), time.Second)
		if err == nil {
			connection.SetDeadline(time.Now().Add(2 * time.Second))
			fmt.Fprint(connection, "HEAD / HTTP/1.0\r\n\r\n")
			buffer := make([]byte, 4)
			n, _ := connection.Read(buffer)
			connection.Close()
			if n > 0 {
				return nil
			}
		}
		select {
		case <-box.exited:
			return fmt.Errorf("%w: the sandbox ended before the preview answered", workspace.ErrNotReady)
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%w: the preview did not answer within %s", workspace.ErrNotReady, within)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func handleOf(box *sandbox) workspace.Handle {
	return workspace.Handle(fmt.Sprintf("%s:%d:%d", box.id, box.preview, box.agent))
}

func parseHandle(handle workspace.Handle) (string, bool) {
	id, _, ok := strings.Cut(string(handle), ":")
	return id, ok && strings.HasPrefix(id, "ws")
}

func (b *Backend) Alive(ctx context.Context, handle workspace.Handle) (workspace.Endpoint, bool) {
	id, ok := parseHandle(handle)
	if !ok {
		return workspace.Endpoint{}, false
	}
	b.mu.Lock()
	box, ours := b.running[id]
	b.mu.Unlock()
	if !ours {
		return workspace.Endpoint{}, false
	}
	select {
	case <-box.exited:
		return workspace.Endpoint{}, false
	default:
	}
	return workspace.Endpoint{Host: "127.0.0.1", Port: box.preview, AgentPort: box.agent}, true
}

func (b *Backend) Stop(ctx context.Context, handle workspace.Handle) error {
	id, ok := parseHandle(handle)
	if !ok {
		return nil
	}
	b.mu.Lock()
	box, ours := b.running[id]
	b.mu.Unlock()
	if !ours {
		b.runsc("delete", "--force", id).Run()
		os.RemoveAll(filepath.Join(b.settings.State, id))
		os.RemoveAll(b.socketsOf(id))
		return nil
	}
	b.end(box)
	return nil
}

func (b *Backend) end(box *sandbox) {
	if box.stop != nil {
		box.stop()
	}
	b.runsc("kill", box.id, "KILL").Run()
	b.runsc("delete", "--force", box.id).Run()
	if box.runsc != nil && box.runsc.Process != nil {
		select {
		case <-box.exited:
		case <-time.After(10 * time.Second):
			box.runsc.Process.Kill()
			<-box.exited
		}
	}
	b.mu.Lock()
	delete(b.running, box.id)
	b.mu.Unlock()
	os.RemoveAll(box.dir)
	os.RemoveAll(box.sockets)
}

func lastLines(path string, count int) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) > count {
		lines = lines[len(lines)-count:]
	}
	return strings.Join(lines, "\n")
}

var _ workspace.Backend = (*Backend)(nil)
