package microvm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/quiqr/chatcms-app/internal/workspace"
)

const (
	DefaultReady  = 60 * time.Second
	DefaultPort   = 1313
	DefaultPrefix = "chv"
	DefaultSubnet = "10.231"
)

var (
	ErrIncomplete = errors.New("microvm: the backend is missing a part it needs")
	ErrNoHugo     = errors.New("microvm: Hugo is not in the package store, so the VM cannot see it")
)

type Guest struct {
	Kernel  string `json:"kernel"`
	Initrd  string `json:"initrd"`
	Cmdline string `json:"cmdline"`
	Memory  int    `json:"memory"`
	VCPU    int    `json:"vcpu"`
}

func LoadGuest(path string) (Guest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Guest{}, fmt.Errorf("microvm: reading the guest failed: %w", err)
	}
	var guest Guest
	if err := json.Unmarshal(data, &guest); err != nil {
		return Guest{}, fmt.Errorf("microvm: the guest is not readable: %w", err)
	}
	if guest.Kernel == "" || guest.Initrd == "" || guest.Cmdline == "" {
		return Guest{}, fmt.Errorf("%w: the guest names no kernel, initrd or command line", ErrIncomplete)
	}
	if guest.Memory <= 0 {
		guest.Memory = 1024
	}
	if guest.VCPU <= 0 {
		guest.VCPU = 1
	}
	return guest, nil
}

type Settings struct {
	CloudHypervisor string
	Virtiofsd       string
	Guest           Guest
	RunDir          string
	Taps            int
	TapPrefix       string
	Subnet          string
	Port            int
	Ready           time.Duration
	HugoBin         string
	Store           string
	GatewayPort     int
}

type vm struct {
	slot    int
	dir     string
	pids    []int
	exited  []chan struct{}
	console string
}

type Backend struct {
	settings Settings

	mu      sync.Mutex
	taken   map[int]bool
	running map[int]*vm
}

func New(settings Settings) (*Backend, error) {
	if settings.CloudHypervisor == "" || settings.Virtiofsd == "" || settings.RunDir == "" || settings.Taps <= 0 || settings.Guest.Kernel == "" {
		return nil, ErrIncomplete
	}
	if settings.TapPrefix == "" {
		settings.TapPrefix = DefaultPrefix
	}
	if settings.Subnet == "" {
		settings.Subnet = DefaultSubnet
	}
	if settings.Port <= 0 {
		settings.Port = DefaultPort
	}
	if settings.Ready <= 0 {
		settings.Ready = DefaultReady
	}
	if settings.Store == "" {
		settings.Store = "/nix/store"
	}
	return &Backend{settings: settings, taken: map[int]bool{}, running: map[int]*vm{}}, nil
}

func (b *Backend) Name() string { return "microvm" }

func (b *Backend) Apart() bool { return false }

func (b *Backend) tap(slot int) string { return b.settings.TapPrefix + strconv.Itoa(slot) }

func (b *Backend) hostAddress(slot int) string {
	return fmt.Sprintf("%s.%d.1", b.settings.Subnet, slot)
}

func (b *Backend) guestAddress(slot int) string {
	return fmt.Sprintf("%s.%d.2", b.settings.Subnet, slot)
}

func mac(slot int) string {
	return fmt.Sprintf("02:00:0a:e7:%02x:02", slot)
}

func (b *Backend) endpoint(slot int) workspace.Endpoint {
	return workspace.Endpoint{Host: b.guestAddress(slot), Port: b.settings.Port}
}

func (b *Backend) take() (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for slot := 0; slot < b.settings.Taps; slot++ {
		if !b.taken[slot] {
			b.taken[slot] = true
			return slot, nil
		}
	}
	return 0, workspace.ErrTooMany
}

func (b *Backend) release(slot int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.taken, slot)
	delete(b.running, slot)
}

func (b *Backend) Free() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.settings.Taps - len(b.taken)
}

func (b *Backend) hugo(spec workspace.Spec) (string, error) {
	hugo := spec.Hugo
	if hugo == "" {
		hugo = b.settings.HugoBin
	}
	if hugo == "" {
		return "", ErrNoHugo
	}
	if !filepath.IsAbs(hugo) {
		found, err := exec.LookPath(hugo)
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrNoHugo, err)
		}
		hugo = found
	}
	real, err := filepath.EvalSymlinks(hugo)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrNoHugo, err)
	}
	if !strings.HasPrefix(real, b.settings.Store+"/") {
		return "", fmt.Errorf("%w: %s", ErrNoHugo, real)
	}
	return real, nil
}

func (b *Backend) Cmdline(slot int, hugo, under, token string) string {
	parts := []string{
		b.settings.Guest.Cmdline,
		"chatcms.ip=" + b.guestAddress(slot) + "/30",
		"chatcms.gateway=" + b.hostAddress(slot),
		"chatcms.hugo=" + hugo,
		"chatcms.under=" + under,
	}
	if port := browserPort(under); port > 0 {
		parts = append(parts, "chatcms.reload="+strconv.Itoa(port))
	}
	if token != "" && b.settings.GatewayPort > 0 {
		parts = append(parts,
			"chatcms.ai="+fmt.Sprintf("http://%s:%d/v1", b.hostAddress(slot), b.settings.GatewayPort),
			"chatcms.token="+token)
	}
	return strings.Join(parts, " ")
}

func (b *Backend) hypervisorArguments(slot int, dir, hugo, under, token string) []string {
	guest := b.settings.Guest
	return []string{
		"--cpus", "boot=" + strconv.Itoa(guest.VCPU),
		"--memory", "size=" + strconv.Itoa(guest.Memory) + "M,shared=on",
		"--kernel", guest.Kernel,
		"--initramfs", guest.Initrd,
		"--cmdline", strings.Replace(b.Cmdline(slot, hugo, under, token), "console=ttyS0", "console=hvc0", 1),
		"--seccomp", "true",
		"--fs", "tag=ro-store,socket=" + filepath.Join(dir, "s"), "tag=site,socket=" + filepath.Join(dir, "t"),
		"--net", "tap=" + b.tap(slot) + ",mac=" + mac(slot),
		"--serial", "off",
		"--console", "file=" + filepath.Join(dir, "console.log"),
	}
}

func shareArguments(socket, source, cache string) []string {
	return []string{
		"--socket-path=" + socket,
		"--shared-dir=" + source,
		"--readonly",
		"--sandbox=none",
		"--cache=" + cache,
	}
}

func (b *Backend) Start(ctx context.Context, spec workspace.Spec, log io.Writer) (workspace.Handle, workspace.Endpoint, error) {
	if strings.ContainsAny(spec.Under, " \t\n") || strings.ContainsAny(spec.Token, " \t\n") || strings.ContainsAny(spec.Dir, "\n") {
		return "", workspace.Endpoint{}, fmt.Errorf("microvm: an address or directory this cannot pass on")
	}
	hugo, err := b.hugo(spec)
	if err != nil {
		return "", workspace.Endpoint{}, err
	}
	slot, err := b.take()
	if err != nil {
		return "", workspace.Endpoint{}, err
	}

	dir := filepath.Join(b.settings.RunDir, strconv.Itoa(slot))
	machine := &vm{slot: slot, dir: dir, console: filepath.Join(dir, "console.log")}
	fail := func(err error) (workspace.Handle, workspace.Endpoint, error) {
		b.stopMachine(machine)
		if tail := lastLines(machine.console, 20); tail != "" {
			fmt.Fprintln(log, tail)
		}
		return "", workspace.Endpoint{}, err
	}

	if err := os.RemoveAll(dir); err != nil {
		return fail(fmt.Errorf("microvm: %w", err))
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fail(fmt.Errorf("microvm: %w", err))
	}

	store := filepath.Join(dir, "s")
	site := filepath.Join(dir, "t")
	for _, share := range []struct{ socket, source, cache string }{
		{store, b.settings.Store, "auto"},
		{site, spec.Dir, "never"},
	} {
		if err := b.launch(machine, log, b.settings.Virtiofsd, shareArguments(share.socket, share.source, share.cache)); err != nil {
			return fail(err)
		}
	}
	if err := waitForSockets(ctx, machine, []string{store, site}, 10*time.Second); err != nil {
		return fail(err)
	}

	fmt.Fprintf(log, "starting a micro VM on %s\n", b.tap(slot))
	if err := b.launch(machine, log, b.settings.CloudHypervisor, b.hypervisorArguments(slot, dir, hugo, spec.Under, spec.Token)); err != nil {
		return fail(err)
	}

	b.mu.Lock()
	b.running[slot] = machine
	b.mu.Unlock()

	endpoint := b.endpoint(slot)
	if err := waitFor(ctx, machine, endpoint, b.settings.Ready); err != nil {
		return fail(err)
	}
	return b.handleOf(machine), endpoint, nil
}

func (b *Backend) launch(machine *vm, log io.Writer, binary string, arguments []string) error {
	command := exec.Command(binary, arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Stdout = log
	command.Stderr = log
	command.Env = []string{"PATH=/usr/bin:/bin", "TZ=UTC"}
	if err := command.Start(); err != nil {
		return fmt.Errorf("microvm: starting %s failed: %w", filepath.Base(binary), err)
	}
	exited := make(chan struct{})
	go func() {
		command.Wait()
		close(exited)
	}()
	machine.pids = append(machine.pids, command.Process.Pid)
	machine.exited = append(machine.exited, exited)
	return nil
}

func anyExited(machine *vm) bool {
	for _, exited := range machine.exited {
		select {
		case <-exited:
			return true
		default:
		}
	}
	return false
}

func waitForSockets(ctx context.Context, machine *vm, sockets []string, within time.Duration) error {
	deadline := time.Now().Add(within)
	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if anyExited(machine) {
			return fmt.Errorf("%w: a share stopped before it was ready", workspace.ErrNotReady)
		}
		ready := true
		for _, socket := range sockets {
			if _, err := os.Stat(socket); err != nil {
				ready = false
			}
		}
		if ready {
			return nil
		}
		time.Sleep(20 * time.Millisecond)
	}
	return fmt.Errorf("%w: the shares did not start", workspace.ErrNotReady)
}

func waitFor(ctx context.Context, machine *vm, endpoint workspace.Endpoint, within time.Duration) error {
	deadline := time.Now().Add(within)
	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if anyExited(machine) {
			return workspace.ErrNotReady
		}
		connection, err := net.DialTimeout("tcp", endpoint.Address(), 500*time.Millisecond)
		if err == nil {
			connection.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return workspace.ErrNotReady
}

func (b *Backend) handleOf(machine *vm) workspace.Handle {
	parts := []string{strconv.Itoa(machine.slot)}
	for _, pid := range machine.pids {
		parts = append(parts, strconv.Itoa(pid))
	}
	return workspace.Handle(strings.Join(parts, ":"))
}

func parseHandle(handle workspace.Handle) (int, []int, bool) {
	parts := strings.Split(string(handle), ":")
	if len(parts) != 4 {
		return 0, nil, false
	}
	slot, err := strconv.Atoi(parts[0])
	if err != nil || slot < 0 {
		return 0, nil, false
	}
	pids := make([]int, 0, 3)
	for _, part := range parts[1:] {
		pid, err := strconv.Atoi(part)
		if err != nil || pid <= 1 {
			return 0, nil, false
		}
		pids = append(pids, pid)
	}
	return slot, pids, true
}

func runs(pid int, binary string) bool {
	cmdline, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "cmdline"))
	if err != nil {
		return false
	}
	first, _, _ := strings.Cut(string(cmdline), "\x00")
	return first == binary
}

func (b *Backend) Alive(ctx context.Context, handle workspace.Handle) (workspace.Endpoint, bool) {
	slot, pids, ok := parseHandle(handle)
	if !ok || slot >= b.settings.Taps || !runs(pids[2], b.settings.CloudHypervisor) {
		return workspace.Endpoint{}, false
	}
	endpoint := b.endpoint(slot)
	connection, err := net.DialTimeout("tcp", endpoint.Address(), time.Second)
	if err != nil {
		return workspace.Endpoint{}, false
	}
	connection.Close()

	b.mu.Lock()
	b.taken[slot] = true
	if _, ours := b.running[slot]; !ours {
		b.running[slot] = &vm{slot: slot, dir: filepath.Join(b.settings.RunDir, strconv.Itoa(slot)), pids: pids}
	}
	b.mu.Unlock()
	return endpoint, true
}

func (b *Backend) Stop(ctx context.Context, handle workspace.Handle) error {
	slot, pids, ok := parseHandle(handle)
	if !ok {
		return nil
	}

	b.mu.Lock()
	machine, ours := b.running[slot]
	b.mu.Unlock()

	if ours {
		if samePids(machine.pids, pids) {
			b.stopMachine(machine)
		}
		return nil
	}

	left := &vm{slot: slot, dir: filepath.Join(b.settings.RunDir, strconv.Itoa(slot))}
	binaries := []string{b.settings.Virtiofsd, b.settings.Virtiofsd, b.settings.CloudHypervisor}
	for i, pid := range pids {
		if runs(pid, binaries[i]) {
			left.pids = append(left.pids, pid)
		}
	}
	for i := len(left.pids) - 1; i >= 0; i-- {
		end(left.pids[i], nil)
	}
	os.RemoveAll(left.dir)
	return nil
}

func samePids(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (b *Backend) stopMachine(machine *vm) {
	for i := len(machine.pids) - 1; i >= 0; i-- {
		var exited chan struct{}
		if i < len(machine.exited) {
			exited = machine.exited[i]
		}
		end(machine.pids[i], exited)
	}
	os.RemoveAll(machine.dir)
	b.release(machine.slot)
}

func end(pid int, exited chan struct{}) {
	syscall.Kill(-pid, syscall.SIGTERM)
	if gone(pid, exited, 5*time.Second) {
		return
	}
	syscall.Kill(-pid, syscall.SIGKILL)
	gone(pid, exited, 5*time.Second)
}

func gone(pid int, exited chan struct{}, within time.Duration) bool {
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

var _ workspace.Backend = (*Backend)(nil)
