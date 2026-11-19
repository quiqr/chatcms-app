package local

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/quiqr/chatcms-app/internal/assistant"
	"github.com/quiqr/chatcms-app/internal/runtime"
	"github.com/quiqr/chatcms-app/internal/workspace"
)

var ErrNoHugo = errors.New("local: no Hugo was named")

const DefaultReady = 30 * time.Second

type Settings struct {
	HugoBin     string
	Ready       time.Duration
	OpenCodeBin string
	Gateway     string
}

type agent struct {
	pid    int
	port   int
	exited chan struct{}
}

type Backend struct {
	settings Settings

	mu       sync.Mutex
	children map[int]chan struct{}
	agents   map[int]agent
}

func New(settings Settings) (*Backend, error) {
	if settings.HugoBin == "" {
		return nil, ErrNoHugo
	}
	if settings.Ready <= 0 {
		settings.Ready = DefaultReady
	}
	return &Backend{settings: settings, children: map[int]chan struct{}{}, agents: map[int]agent{}}, nil
}

func (b *Backend) Name() string { return "local" }

func (b *Backend) Apart() bool { return false }

func (b *Backend) resolved(spec workspace.Spec) (runtime.Resolved, error) {
	if spec.Run.Preview.Program != "" {
		return spec.Run, nil
	}
	return runtime.Resolver{InstalledHugo: b.settings.HugoBin, Exists: func(string) bool { return true }}.Resolve(context.Background(), runtime.Default(""), false)
}

func (b *Backend) Start(ctx context.Context, spec workspace.Spec, log io.Writer) (workspace.Handle, workspace.Endpoint, error) {
	run, err := b.resolved(spec)
	if err != nil {
		return "", workspace.Endpoint{}, err
	}
	port, err := freePort()
	if err != nil {
		return "", workspace.Endpoint{}, err
	}

	home := filepath.Join(filepath.Dir(spec.Dir), ".hugohome")
	cache := filepath.Join(home, "cache")
	if err := os.MkdirAll(cache, 0o700); err != nil {
		return "", workspace.Endpoint{}, fmt.Errorf("local: %w", err)
	}

	values := map[string]string{
		"{host}":   "127.0.0.1",
		"{port}":   strconv.Itoa(port),
		"{url}":    spec.Under,
		"{source}": spec.Dir,
		"{cache}":  cache,
	}
	if browser := browserPort(spec.Under); browser > 0 {
		values["{reload}"] = strconv.Itoa(browser)
	}

	hugo := run.Preview.Program
	command := exec.Command(hugo, runtime.Fill(run.Preview.Args, values)...)
	command.Dir = spec.Dir
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Env = []string{"HOME=" + home, "PATH=" + strings.Join(append(append([]string{}, run.Path...), "/usr/bin", "/bin"), ":"), "TZ=UTC"}
	for _, name := range slices.Sorted(maps.Keys(run.Env)) {
		command.Env = append(command.Env, name+"="+run.Env[name])
	}
	command.Stdout = log
	command.Stderr = log

	if err := command.Start(); err != nil {
		return "", workspace.Endpoint{}, fmt.Errorf("local: %w", err)
	}

	pid := command.Process.Pid
	exited := make(chan struct{})
	b.mu.Lock()
	b.children[pid] = exited
	b.mu.Unlock()

	go func() {
		command.Wait()
		close(exited)
	}()

	endpoint := workspace.Endpoint{Host: "127.0.0.1", Port: port}
	handle := handleOf(pid, port, hugo)
	if err := waitFor(ctx, endpoint, b.settings.Ready, exited); err != nil {
		b.Stop(context.Background(), handle)
		return "", workspace.Endpoint{}, err
	}

	if started, ok := b.startAgent(spec, log); ok {
		b.mu.Lock()
		b.agents[pid] = started
		b.mu.Unlock()
		endpoint.AgentPort = started.port
	}
	return handle, endpoint, nil
}

func (b *Backend) startAgent(spec workspace.Spec, log io.Writer) (agent, bool) {
	if b.settings.OpenCodeBin == "" || b.settings.Gateway == "" || spec.Token == "" {
		return agent{}, false
	}
	port, err := freePort()
	if err != nil {
		fmt.Fprintf(log, "the assistant could not start: %v\n", err)
		return agent{}, false
	}
	home := filepath.Join(filepath.Dir(spec.Dir), ".agenthome")
	if err := os.MkdirAll(home, 0o700); err != nil {
		fmt.Fprintf(log, "the assistant could not start: %v\n", err)
		return agent{}, false
	}

	command := exec.Command(b.settings.OpenCodeBin, "serve", "--hostname", "127.0.0.1", "--port", strconv.Itoa(port))
	command.Dir = spec.Dir
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Env = append(assistant.OpenCodeEnvironment(home, b.settings.Gateway, spec.Token), "PATH=/usr/bin:/bin")
	command.Stdout = log
	command.Stderr = log
	if err := command.Start(); err != nil {
		fmt.Fprintf(log, "the assistant could not start: %v\n", err)
		return agent{}, false
	}
	exited := make(chan struct{})
	go func() {
		command.Wait()
		close(exited)
	}()
	fmt.Fprintf(log, "assistant on port %d\n", port)
	return agent{pid: command.Process.Pid, port: port, exited: exited}, true
}

func handleOf(pid, port int, hugo string) workspace.Handle {
	return workspace.Handle(fmt.Sprintf("%d:%d:%s", pid, port, hugo))
}

func parseHandle(handle workspace.Handle) (int, int, string, bool) {
	parts := strings.SplitN(string(handle), ":", 3)
	if len(parts) != 3 {
		return 0, 0, "", false
	}
	pid, err := strconv.Atoi(parts[0])
	if err != nil || pid <= 1 {
		return 0, 0, "", false
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil || port <= 0 {
		return 0, 0, "", false
	}
	return pid, port, parts[2], parts[2] != ""
}

func (b *Backend) Alive(ctx context.Context, handle workspace.Handle) (workspace.Endpoint, bool) {
	pid, port, hugo, ok := parseHandle(handle)
	if !ok || !runs(pid, hugo) {
		return workspace.Endpoint{}, false
	}
	endpoint := workspace.Endpoint{Host: "127.0.0.1", Port: port}
	connection, err := net.DialTimeout("tcp", endpoint.Address(), time.Second)
	if err != nil {
		return workspace.Endpoint{}, false
	}
	connection.Close()
	return endpoint, true
}

func runs(pid int, hugo string) bool {
	cmdline, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "cmdline"))
	if err != nil {
		return false
	}
	first, _, _ := strings.Cut(string(cmdline), "\x00")
	return first == hugo
}

func (b *Backend) Stop(ctx context.Context, handle workspace.Handle) error {
	pid, _, hugo, ok := parseHandle(handle)
	if !ok {
		return nil
	}

	b.mu.Lock()
	exited, ours := b.children[pid]
	delete(b.children, pid)
	helper, helped := b.agents[pid]
	delete(b.agents, pid)
	b.mu.Unlock()

	if helped {
		syscall.Kill(-helper.pid, syscall.SIGTERM)
		if !gone(helper.pid, helper.exited, 5*time.Second) {
			syscall.Kill(-helper.pid, syscall.SIGKILL)
			gone(helper.pid, helper.exited, 5*time.Second)
		}
	}

	if !ours && !runs(pid, hugo) {
		return nil
	}

	syscall.Kill(-pid, syscall.SIGTERM)
	if gone(pid, exited, 5*time.Second) {
		return nil
	}
	syscall.Kill(-pid, syscall.SIGKILL)
	gone(pid, exited, 5*time.Second)
	return nil
}

func gone(pid int, exited <-chan struct{}, within time.Duration) bool {
	deadline := time.Now().Add(within)
	for time.Now().Before(deadline) {
		if exited != nil {
			select {
			case <-exited:
				return true
			default:
			}
		} else if syscall.Kill(pid, 0) != nil {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return false
}

func waitFor(ctx context.Context, endpoint workspace.Endpoint, within time.Duration, exited <-chan struct{}) error {
	deadline := time.Now().Add(within)
	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		select {
		case <-exited:
			return workspace.ErrNotReady
		default:
		}
		connection, err := net.DialTimeout("tcp", endpoint.Address(), time.Second)
		if err == nil {
			connection.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return workspace.ErrNotReady
}

func browserPort(under string) int {
	where, err := url.Parse(under)
	if err != nil {
		return 0
	}
	if named := where.Port(); named != "" {
		if port, err := strconv.Atoi(named); err == nil {
			return port
		}
		return 0
	}
	switch where.Scheme {
	case "https":
		return 443
	case "http":
		return 80
	default:
		return 0
	}
}

func freePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("local: %w", err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

var _ workspace.Backend = (*Backend)(nil)
