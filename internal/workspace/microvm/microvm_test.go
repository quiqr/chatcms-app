package microvm

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/workspace"
)

func TestMain(m *testing.M) {
	switch filepath.Base(os.Args[0]) {
	case "virtiofsd":
		fakeShare()
	case "cloud-hypervisor":
		fakeHypervisor()
	default:
		os.Exit(m.Run())
	}
}

func argument(prefix string) string {
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, prefix) {
			return strings.TrimPrefix(arg, prefix)
		}
	}
	return ""
}

func untilTold() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop
	os.Exit(0)
}

func fakeShare() {
	socket := argument("--socket-path=")
	if _, err := os.Stat(argument("--shared-dir=")); err != nil {
		os.Exit(3)
	}
	listener, err := net.Listen("unix", socket)
	if err != nil {
		os.Exit(2)
	}
	defer listener.Close()
	untilTold()
}

func fakeHypervisor() {
	here := filepath.Dir(os.Args[0])
	if _, err := os.Stat(filepath.Join(here, "refuse")); err == nil {
		fmt.Println("cloud-hypervisor: refusing")
		os.Exit(1)
	}
	var cmdline string
	for i, arg := range os.Args {
		if arg == "--cmdline" && i+1 < len(os.Args) {
			cmdline = os.Args[i+1]
		}
	}
	var address, hugo string
	for _, part := range strings.Fields(cmdline) {
		if value, found := strings.CutPrefix(part, "chatcms.ip="); found {
			address, _, _ = strings.Cut(value, "/")
		}
		if value, found := strings.CutPrefix(part, "chatcms.hugo="); found {
			hugo = value
		}
	}
	if _, err := os.Stat(filepath.Join(here, "hang")); err == nil {
		untilTold()
	}
	port, _ := os.ReadFile(filepath.Join(here, "port"))
	listener, err := net.Listen("tcp", net.JoinHostPort(address, strings.TrimSpace(string(port))))
	if err != nil {
		os.Exit(2)
	}
	go http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hugo=%s cmdline=%s", hugo, cmdline)
	}))
	untilTold()
}

type lockedLog struct {
	mu   sync.Mutex
	text strings.Builder
}

func (l *lockedLog) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.text.Write(p)
}

func (l *lockedLog) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.text.String()
}

type rig struct {
	dir     string
	backend *Backend
	port    int
	hugo    string
}

func freePort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func newRig(t *testing.T, taps int) *rig {
	t.Helper()

	dir := t.TempDir()
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"virtiofsd", "cloud-hypervisor"} {
		if err := os.Symlink(self, filepath.Join(dir, name)); err != nil {
			t.Fatal(err)
		}
	}
	port := freePort(t)
	if err := os.WriteFile(filepath.Join(dir, "port"), []byte(strconv.Itoa(port)), 0o644); err != nil {
		t.Fatal(err)
	}

	store := filepath.Join(dir, "store")
	hugo := filepath.Join(store, "abc-hugo-0.139.0", "bin", "hugo")
	if err := os.MkdirAll(filepath.Dir(hugo), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(hugo, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	backend, err := New(Settings{
		CloudHypervisor: filepath.Join(dir, "cloud-hypervisor"),
		Virtiofsd:       filepath.Join(dir, "virtiofsd"),
		Guest:           Guest{Kernel: "/k", Initrd: "/i", Cmdline: "console=ttyS0 init=/init", Memory: 512, VCPU: 2},
		RunDir:          filepath.Join(dir, "w"),
		Taps:            taps,
		Subnet:          "127.231",
		Port:            port,
		Ready:           5 * time.Second,
		HugoBin:         hugo,
		Store:           store,
		GatewayPort:     8091,
	})
	if err != nil {
		t.Fatal(err)
	}
	return &rig{dir: dir, backend: backend, port: port, hugo: hugo}
}

func (r *rig) spec(t *testing.T) workspace.Spec {
	t.Helper()
	site := filepath.Join(t.TempDir(), "working-copy")
	if err := os.MkdirAll(site, 0o755); err != nil {
		t.Fatal(err)
	}
	return workspace.Spec{
		Key:   workspace.Key{Person: uuid.New(), Project: uuid.New()},
		Dir:   site,
		Under: "http://127.0.0.1:8080/preview/o/p/",
	}
}

func (r *rig) start(t *testing.T, spec workspace.Spec) (workspace.Handle, workspace.Endpoint) {
	t.Helper()
	handle, endpoint, err := r.backend.Start(t.Context(), spec, io.Discard)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	t.Cleanup(func() { r.backend.Stop(context.Background(), handle) })
	return handle, endpoint
}

func get(t *testing.T, endpoint workspace.Endpoint) string {
	t.Helper()
	answer, err := http.Get("http://" + endpoint.Address() + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer answer.Body.Close()
	body, _ := io.ReadAll(answer.Body)
	return string(body)
}

func pidsOf(t *testing.T, handle workspace.Handle) []int {
	t.Helper()
	_, pids, ok := parseHandle(handle)
	if !ok {
		t.Fatalf("handle %q does not parse", handle)
	}
	return pids
}

func alive(pid int) bool { return syscall.Kill(pid, 0) == nil }

func TestAWorkspaceStartsInAVMOfItsOwn(t *testing.T) {
	r := newRig(t, 2)
	handle, endpoint := r.start(t, r.spec(t))

	if endpoint.Host != "127.231.0.2" || endpoint.Port != r.port {
		t.Errorf("endpoint = %+v, want the first slot's guest address", endpoint)
	}
	body := get(t, endpoint)
	for _, want := range []string{
		"hugo=" + r.hugo,
		"chatcms.ip=127.231.0.2/30",
		"chatcms.gateway=127.231.0.1",
		"chatcms.under=http://127.0.0.1:8080/preview/o/p/",
		"chatcms.reload=8080",
		"console=hvc0",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("the VM was not given %q: %s", want, body)
		}
	}
	if len(pidsOf(t, handle)) != 3 {
		t.Errorf("handle %q, want two shares and a hypervisor", handle)
	}
	if found, ok := r.backend.Alive(t.Context(), handle); !ok || found != endpoint {
		t.Errorf("Alive() = %v, %v", found, ok)
	}
}

func TestTheProjectsHugoMustBeInTheStore(t *testing.T) {
	r := newRig(t, 1)
	spec := r.spec(t)
	outside := filepath.Join(t.TempDir(), "hugo")
	os.WriteFile(outside, []byte("#!/bin/sh\n"), 0o755)
	spec.Hugo = outside

	if _, _, err := r.backend.Start(t.Context(), spec, io.Discard); !errors.Is(err, ErrNoHugo) {
		t.Errorf("Start() with a Hugo outside the store error = %v, want ErrNoHugo", err)
	}
	if r.backend.Free() != 1 {
		t.Errorf("a refused start kept its slot")
	}
}

func TestEachVMHasItsOwnAddressAndTheyRunOut(t *testing.T) {
	r := newRig(t, 2)
	_, first := r.start(t, r.spec(t))
	handle, second := r.start(t, r.spec(t))
	if first.Host == second.Host {
		t.Errorf("two VMs share %s", first.Host)
	}

	if _, _, err := r.backend.Start(t.Context(), r.spec(t), io.Discard); !errors.Is(err, workspace.ErrTooMany) {
		t.Errorf("a third Start() error = %v, want ErrTooMany", err)
	}

	r.backend.Stop(t.Context(), handle)
	_, again := r.start(t, r.spec(t))
	if again.Host != second.Host {
		t.Errorf("a freed address was not given out again: %s", again.Host)
	}
}

func TestStoppingLeavesNothingBehind(t *testing.T) {
	r := newRig(t, 1)
	handle, _ := r.start(t, r.spec(t))
	pids := pidsOf(t, handle)

	if err := r.backend.Stop(t.Context(), handle); err != nil {
		t.Fatal(err)
	}
	if err := r.backend.Stop(t.Context(), handle); err != nil {
		t.Errorf("stopping twice error = %v", err)
	}
	for _, pid := range pids {
		if alive(pid) {
			t.Errorf("process %d is still running", pid)
		}
	}
	if _, err := os.Stat(filepath.Join(r.dir, "w", "0")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the VM's directory is still there")
	}
	if r.backend.Free() != 1 {
		t.Errorf("the address was not freed")
	}
	if _, ok := r.backend.Alive(t.Context(), handle); ok {
		t.Errorf("Alive() after Stop() = true")
	}
}

func TestAHypervisorThatStopsIsNotReady(t *testing.T) {
	r := newRig(t, 1)
	os.WriteFile(filepath.Join(r.dir, "refuse"), nil, 0o644)

	log := &lockedLog{}
	if _, _, err := r.backend.Start(t.Context(), r.spec(t), log); !errors.Is(err, workspace.ErrNotReady) {
		t.Fatalf("Start() error = %v, want ErrNotReady", err)
	}
	if !strings.Contains(log.String(), "refusing") {
		t.Errorf("the hypervisor's words did not reach the log: %q", log.String())
	}
	if r.backend.Free() != 1 {
		t.Errorf("a failed start kept its address")
	}
}

func TestAVMThatNeverAnswersIsStoppedInTime(t *testing.T) {
	r := newRig(t, 1)
	r.backend.settings.Ready = 500 * time.Millisecond
	os.WriteFile(filepath.Join(r.dir, "hang"), nil, 0o644)

	if _, _, err := r.backend.Start(t.Context(), r.spec(t), io.Discard); !errors.Is(err, workspace.ErrNotReady) {
		t.Errorf("Start() error = %v, want ErrNotReady", err)
	}
	if r.backend.Free() != 1 {
		t.Errorf("a VM that never answered kept its address")
	}
}

func TestAShareThatCannotStartIsNotReady(t *testing.T) {
	r := newRig(t, 1)
	spec := r.spec(t)
	spec.Dir = filepath.Join(t.TempDir(), "missing")

	if _, _, err := r.backend.Start(t.Context(), spec, io.Discard); !errors.Is(err, workspace.ErrNotReady) {
		t.Errorf("Start() error = %v, want ErrNotReady", err)
	}
}

func TestAnotherBackendCanFindAndStopARunningVM(t *testing.T) {
	r := newRig(t, 2)
	handle, endpoint, err := r.backend.Start(t.Context(), r.spec(t), io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	pids := pidsOf(t, handle)

	again, err := New(r.backend.settings)
	if err != nil {
		t.Fatal(err)
	}
	if found, ok := again.Alive(t.Context(), handle); !ok || found != endpoint {
		t.Fatalf("Alive() from another backend = %v, %v", found, ok)
	}
	if again.Free() != 1 {
		t.Errorf("an adopted VM did not take its address back")
	}
	if err := again.Stop(t.Context(), handle); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for _, pid := range pids {
		for alive(pid) && time.Now().Before(deadline) {
			time.Sleep(20 * time.Millisecond)
		}
		if alive(pid) {
			t.Errorf("process %d survived", pid)
		}
	}

	stranger, _ := New(r.backend.settings)
	if err := stranger.Stop(t.Context(), handle); err != nil {
		t.Errorf("stopping a VM that is gone error = %v", err)
	}
}

func TestAStaleHandleLeavesTheVMNowInItsPlace(t *testing.T) {
	r := newRig(t, 1)
	first, _, err := r.backend.Start(t.Context(), r.spec(t), io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	r.backend.Stop(t.Context(), first)
	second, endpoint := r.start(t, r.spec(t))

	r.backend.Stop(t.Context(), first)
	if found, ok := r.backend.Alive(t.Context(), second); !ok || found != endpoint {
		t.Errorf("a stale handle stopped the VM now in its slot")
	}
}

func TestHandlesRefuseNonsense(t *testing.T) {
	for _, bad := range []string{"", "0:1:2", "x:2:3:4", "0:1:3:4", "0:2:x:4", "-1:2:3:4"} {
		if _, _, ok := parseHandle(workspace.Handle(bad)); ok {
			t.Errorf("parseHandle(%q) accepted", bad)
		}
	}
	r := newRig(t, 1)
	if _, ok := r.backend.Alive(t.Context(), "9:2:3:4"); ok {
		t.Errorf("Alive() accepted a slot outside the pool")
	}
	if err := r.backend.Stop(t.Context(), "nonsense"); err != nil {
		t.Errorf("Stop() of nonsense error = %v", err)
	}
}

func TestAnAddressWithSpacesIsRefused(t *testing.T) {
	r := newRig(t, 1)
	spec := r.spec(t)
	spec.Under = "http://x/ chatcms.hugo=/bin/sh"
	if _, _, err := r.backend.Start(t.Context(), spec, io.Discard); err == nil {
		t.Errorf("Start() passed an address with a space on to the kernel")
	}
}

func TestLoadGuest(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "good.json")
	os.WriteFile(good, []byte(`{"kernel":"/k","initrd":"/i","cmdline":"init=/init"}`), 0o644)
	guest, err := LoadGuest(good)
	if err != nil || guest.Memory != 1024 || guest.VCPU != 1 {
		t.Errorf("LoadGuest() = %+v, %v, want defaults filled in", guest, err)
	}

	empty := filepath.Join(dir, "empty.json")
	os.WriteFile(empty, []byte(`{}`), 0o644)
	if _, err := LoadGuest(empty); !errors.Is(err, ErrIncomplete) {
		t.Errorf("LoadGuest() of an empty guest error = %v", err)
	}
	garbled := filepath.Join(dir, "garbled.json")
	os.WriteFile(garbled, []byte(`{`), 0o644)
	if _, err := LoadGuest(garbled); err == nil {
		t.Errorf("LoadGuest() of garbled JSON succeeded")
	}
	if _, err := LoadGuest(filepath.Join(dir, "missing.json")); err == nil {
		t.Errorf("LoadGuest() of a missing file succeeded")
	}
}

func TestNewNeedsItsParts(t *testing.T) {
	if _, err := New(Settings{}); !errors.Is(err, ErrIncomplete) {
		t.Errorf("New() error = %v", err)
	}
	made, err := New(Settings{CloudHypervisor: "ch", Virtiofsd: "v", RunDir: "/run/x", Taps: 1, Guest: Guest{Kernel: "k"}})
	if err != nil || made.Name() != "microvm" || made.Apart() || made.settings.Port != DefaultPort || made.settings.Subnet != DefaultSubnet {
		t.Errorf("New() = %+v, %v", made, err)
	}
	if made.tap(3) != "chv3" || mac(3) != "02:00:0a:e7:03:02" {
		t.Errorf("tap(3) = %s, mac(3) = %s", made.tap(3), mac(3))
	}
}

func TestAVMIsGivenItsTokenAndTheGatewayOnItsHostAddress(t *testing.T) {
	r := newRig(t, 2)
	r.start(t, r.spec(t))
	spec := r.spec(t)
	spec.Token = "a-token"
	_, endpoint := r.start(t, spec)

	body := get(t, endpoint)
	if !strings.Contains(body, "chatcms.ai=http://127.231.1.1:8091/v1") || !strings.Contains(body, "chatcms.token=a-token") {
		t.Errorf("the VM was not given its gateway and token: %s", body)
	}

	spec = r.spec(t)
	spec.Token = "two words"
	r.backend.Stop(t.Context(), workspace.Handle(""))
	if _, _, err := r.backend.Start(t.Context(), spec, io.Discard); err == nil {
		t.Errorf("a token with a space was passed on to the kernel")
	}
}

func TestAVMWithoutATokenIsToldNoGateway(t *testing.T) {
	r := newRig(t, 1)
	_, endpoint := r.start(t, r.spec(t))
	if body := get(t, endpoint); strings.Contains(body, "chatcms.ai=") || strings.Contains(body, "chatcms.token=") {
		t.Errorf("a VM without a token was told about the gateway: %s", body)
	}
}
