package gvisor

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/quiqr/chatcms-app/internal/runtime"
	"github.com/quiqr/chatcms-app/internal/workspace"
)

func inStore(t *testing.T, name string) string {
	t.Helper()
	found, err := exec.LookPath(name)
	if err != nil {
		t.Skipf("no %s", name)
	}
	real, err := filepath.EvalSymlinks(found)
	if err != nil || !strings.HasPrefix(real, "/nix/store/") {
		t.Skipf("%s is not in the store", name)
	}
	return real
}

func aSpec(t *testing.T, token string) workspace.Spec {
	shell := inStore(t, "sh")
	return workspace.Spec{
		Dir:   t.TempDir(),
		Under: "http://127.0.0.1:9/",
		Token: token,
		Run:   runtime.Resolved{Preview: runtime.Command{Program: shell, Args: []string{"-c", "true"}}, Path: []string{filepath.Dir(shell)}},
	}
}

func TestABackendNeedsItsParts(t *testing.T) {
	if _, err := New(Settings{}); !errors.Is(err, ErrIncomplete) {
		t.Fatalf("New() = %v", err)
	}
}

func TestTheBundleSeesOnlyWhatTheWorkspaceNeedsAndNoNetwork(t *testing.T) {
	shell := inStore(t, "sh")
	b, err := New(Settings{Runsc: "runsc", Relay: shell, State: t.TempDir(), Command: func(string, ...string) *exec.Cmd { return exec.Command("true") }})
	if err != nil {
		t.Fatal(err)
	}
	spec := aSpec(t, "tok")
	config, err := b.Bundle("ws1", spec, "RUN")
	if err != nil {
		t.Fatal(err)
	}
	var mounts []string
	for _, mount := range config["mounts"].([]map[string]any) {
		mounts = append(mounts, mount["destination"].(string)+" "+mount["type"].(string))
	}
	if got := strings.Join(mounts, ","); got != "/nix/store bind,/site bind,/run tmpfs,/run/chatcms bind,/tmp tmpfs,/proc proc" {
		t.Fatalf("mounts = %s", got)
	}
	for _, mount := range config["mounts"].([]map[string]any) {
		if mount["destination"] == "/nix/store" && strings.Join(mount["options"].([]string), ",") != "rbind,ro" {
			t.Fatalf("the store is not read-only: %v", mount["options"])
		}
	}
	for _, namespace := range config["linux"].(map[string]any)["namespaces"].([]map[string]any) {
		if namespace["type"] == "network" {
			t.Fatal("the bundle asks for a network namespace; runsc --network=none gives none")
		}
	}
	env := strings.Join(config["process"].(map[string]any)["env"].([]string), " ")
	if !strings.Contains(env, "CHATCMS_RUN=RUN") || !strings.Contains(env, "CHATCMS_AI_TOKEN=tok") {
		t.Fatalf("env = %s", env)
	}
}

func TestRunscIsAskedForNoNetworkAndHostSockets(t *testing.T) {
	var asked []string
	b, err := New(Settings{Runsc: "/x/runsc", Relay: "r", State: t.TempDir(), Command: func(name string, arguments ...string) *exec.Cmd {
		asked = append(asked, name+" "+strings.Join(arguments, " "))
		return exec.Command("true")
	}})
	if err != nil {
		t.Fatal(err)
	}
	b.runsc("run", "--bundle", "b", "ws1").Run()
	got := asked[len(asked)-1]
	for _, want := range []string{"--rootless", "--network=none", "--host-uds=all", "run --bundle b ws1"} {
		if !strings.Contains(got, want) {
			t.Fatalf("runsc was asked %q, lacking %q", got, want)
		}
	}
	if !strings.HasPrefix(asked[0], "/x/runsc ") || !strings.Contains(asked[0], "list -quiet") {
		t.Fatalf("leftovers were not looked for first: %v", asked)
	}
}

func TestASandboxThatEndsEarlyIsReportedAndLeavesNothing(t *testing.T) {
	relay := inStore(t, "sh")
	state := t.TempDir()
	b, err := New(Settings{Runsc: "runsc", Relay: relay, State: state, Command: func(name string, arguments ...string) *exec.Cmd {
		return exec.Command("false")
	}})
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = b.Start(context.Background(), aSpec(t, ""), io.Discard)
	if !errors.Is(err, workspace.ErrNotReady) {
		t.Fatalf("Start() = %v", err)
	}
	entries, _ := os.ReadDir(state)
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "ws") {
			t.Fatalf("%s was left behind", entry.Name())
		}
	}
	if left, _ := os.ReadDir(filepath.Join(state, "s")); len(left) != 0 {
		t.Fatalf("sockets were left behind: %v", left)
	}
}

func TestAHandleNamesItsSandbox(t *testing.T) {
	if id, ok := parseHandle("ws0a1b:8000:0"); !ok || id != "ws0a1b" {
		t.Fatalf("parseHandle = %q %v", id, ok)
	}
	if _, ok := parseHandle("12:1:2"); ok {
		t.Fatal("a micro VM's handle was taken for a sandbox")
	}
	b := &Backend{running: map[string]*sandbox{}}
	if _, alive := b.Alive(context.Background(), "ws9:1:2"); alive {
		t.Fatal("a sandbox this backend does not hold is alive")
	}
}
