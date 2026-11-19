package nixpin

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

type memory struct {
	mu   sync.Mutex
	pkgs map[string]Package
}

func (m *memory) Get(ctx context.Context, name, version, system string) (Package, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	pkg, found := m.pkgs[name+version+system]
	return pkg, found, nil
}

func (m *memory) Put(ctx context.Context, pkg Package) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pkgs[pkg.Name+pkg.Version+pkg.System] = pkg
	return nil
}

const rev = "93dc9803a1ee435e590b02cde9589038d5cc3a4e"

func nixhub(t *testing.T, asked *atomic.Int32, answer func(w http.ResponseWriter, r *http.Request)) string {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		asked.Add(1)
		answer(w, r)
	}))
	t.Cleanup(server.Close)
	return server.URL
}

func hugoAnswer(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/v2/resolve" || r.URL.Query().Get("name") != "hugo" {
		http.NotFound(w, r)
		return
	}
	if r.URL.Query().Get("version") != "0.139.0" {
		http.Error(w, "no package found", http.StatusNotFound)
		return
	}
	w.Write([]byte(`{"name":"hugo","systems":{"x86_64-linux":{"flake_installable":{"ref":{"type":"github","owner":"NixOS","repo":"nixpkgs","rev":"` + rev + `"},"attr_path":"hugo"}}}}`))
}

func fakeNix(t *testing.T, script string) (string, string) {
	t.Helper()
	dir := t.TempDir()
	log := filepath.Join(dir, "asked")
	bin := filepath.Join(dir, "nix")
	body := "#!/bin/sh\necho \"$@\" >> " + log + "\n" + script + "\n"
	if err := os.WriteFile(bin, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	return bin, log
}

func pins(t *testing.T, hub, nix string, store Store, exists func(string) bool) *Pins {
	t.Helper()
	made, err := New(store, Settings{Nixhub: hub, NixBin: nix, Roots: t.TempDir(), System: "x86_64-linux", Exists: exists})
	if err != nil {
		t.Fatal(err)
	}
	return made
}

func always(string) bool { return true }

func TestAVersionIsResolvedRealisedAndRemembered(t *testing.T) {
	var asked atomic.Int32
	hub := nixhub(t, &asked, hugoAnswer)
	nix, log := fakeNix(t, `echo /nix/store/abc-hugo-0.139.0`)
	store := &memory{pkgs: map[string]Package{}}
	p := pins(t, hub, nix, store, always)

	path, err := p.Path(t.Context(), "hugo", "0.139.0", "")
	if err != nil || path != "/nix/store/abc-hugo-0.139.0" {
		t.Fatalf("Path() = %q, %v", path, err)
	}
	called, _ := os.ReadFile(log)
	if !strings.Contains(string(called), "github:NixOS/nixpkgs/"+rev+"#hugo") || !strings.Contains(string(called), "--out-link") {
		t.Errorf("nix was asked %q", called)
	}

	again, err := p.Path(t.Context(), "hugo", "0.139.0", "")
	if err != nil || again != path || asked.Load() != 1 {
		t.Errorf("second Path() = %q, %v, asked the index %d times, want once", again, err, asked.Load())
	}
	if remembered, found, _ := store.Get(t.Context(), "hugo", "0.139.0", "x86_64-linux"); !found || remembered.Revision != rev {
		t.Errorf("remembered %+v, %v", remembered, found)
	}
}

func TestAPathThatWentAwayIsRealisedAgainWithoutAsking(t *testing.T) {
	var asked atomic.Int32
	hub := nixhub(t, &asked, hugoAnswer)
	nix, log := fakeNix(t, `echo /nix/store/abc-hugo-0.139.0`)
	store := &memory{pkgs: map[string]Package{"hugo0.139.0x86_64-linux": {
		Name: "hugo", Version: "0.139.0", System: "x86_64-linux", Revision: rev, AttrPath: "hugo", StorePath: "/nix/store/gone",
	}}}
	p := pins(t, hub, nix, store, func(path string) bool { return path != "/nix/store/gone" })

	path, err := p.Path(t.Context(), "hugo", "0.139.0", "")
	if err != nil || path != "/nix/store/abc-hugo-0.139.0" || asked.Load() != 0 {
		t.Fatalf("Path() = %q, %v, asked %d, want realised again without the index", path, err, asked.Load())
	}
	if called, _ := os.ReadFile(log); !strings.Contains(string(called), rev) {
		t.Errorf("nix was asked %q", called)
	}
}

func TestAnUnknownVersionIsNotFound(t *testing.T) {
	var asked atomic.Int32
	p := pins(t, nixhub(t, &asked, hugoAnswer), "/bin/false", &memory{pkgs: map[string]Package{}}, always)

	if _, err := p.Path(t.Context(), "hugo", "0.0.1", ""); !errors.Is(err, ErrNotFound) {
		t.Errorf("Path() error = %v, want ErrNotFound", err)
	}
}

func TestAnUnusableNameOrVersionIsRefusedBeforeAsking(t *testing.T) {
	var asked atomic.Int32
	p := pins(t, nixhub(t, &asked, hugoAnswer), "/bin/false", &memory{pkgs: map[string]Package{}}, always)

	for _, pair := range [][2]string{{"hugo", "1.0; rm -rf"}, {"../hugo", "1"}, {"hugo", ""}, {"", "1"}} {
		if _, err := p.Path(t.Context(), pair[0], pair[1], ""); !errors.Is(err, ErrUnusable) {
			t.Errorf("Path(%q, %q) error = %v, want ErrUnusable", pair[0], pair[1], err)
		}
	}
	if asked.Load() != 0 {
		t.Errorf("the index was asked about an unusable version")
	}
}

func TestAnIndexThatNamesSomethingElseIsRefused(t *testing.T) {
	var asked atomic.Int32
	hub := nixhub(t, &asked, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"systems":{"x86_64-linux":{"flake_installable":{"ref":{"owner":"evil","repo":"nixpkgs","rev":"` + rev + `"},"attr_path":"hugo"}}}}`))
	})
	p := pins(t, hub, "/bin/false", &memory{pkgs: map[string]Package{}}, always)

	if _, err := p.Path(t.Context(), "hugo", "0.139.0", ""); !errors.Is(err, ErrUnavailable) {
		t.Errorf("Path() error = %v, want ErrUnavailable", err)
	}
}

func TestAnotherSystemOnlyIsNotFound(t *testing.T) {
	var asked atomic.Int32
	hub := nixhub(t, &asked, hugoAnswer)
	made, err := New(&memory{pkgs: map[string]Package{}}, Settings{Nixhub: hub, NixBin: "/bin/false", Roots: t.TempDir(), System: "riscv64-linux"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := made.Path(t.Context(), "hugo", "0.139.0", ""); !errors.Is(err, ErrNotFound) {
		t.Errorf("Path() error = %v, want ErrNotFound", err)
	}
}

func TestTroubleWithTheIndexOrNixIsUnavailable(t *testing.T) {
	var asked atomic.Int32
	broken := nixhub(t, &asked, func(w http.ResponseWriter, r *http.Request) { http.Error(w, "down", http.StatusBadGateway) })
	if _, err := pins(t, broken, "/bin/false", &memory{pkgs: map[string]Package{}}, always).Path(t.Context(), "hugo", "0.139.0", ""); !errors.Is(err, ErrUnavailable) {
		t.Errorf("a failing index error = %v, want ErrUnavailable", err)
	}

	garbled := nixhub(t, &asked, func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("{")) })
	if _, err := pins(t, garbled, "/bin/false", &memory{pkgs: map[string]Package{}}, always).Path(t.Context(), "hugo", "0.139.0", ""); !errors.Is(err, ErrUnavailable) {
		t.Errorf("a garbled answer error = %v, want ErrUnavailable", err)
	}

	failing, _ := fakeNix(t, `echo "error: cannot build" >&2; exit 1`)
	_, err := pins(t, nixhub(t, &asked, hugoAnswer), failing, &memory{pkgs: map[string]Package{}}, always).Path(t.Context(), "hugo", "0.139.0", "")
	if !errors.Is(err, ErrUnavailable) || !strings.Contains(err.Error(), "cannot build") {
		t.Errorf("a failing nix error = %v, want ErrUnavailable with its words", err)
	}

	silent, _ := fakeNix(t, `echo nothing`)
	if _, err := pins(t, nixhub(t, &asked, hugoAnswer), silent, &memory{pkgs: map[string]Package{}}, always).Path(t.Context(), "hugo", "0.139.0", ""); !errors.Is(err, ErrUnavailable) {
		t.Errorf("a nix naming no path error = %v, want ErrUnavailable", err)
	}

	if _, err := pins(t, "http://127.0.0.1:1", "/bin/false", &memory{pkgs: map[string]Package{}}, always).Path(t.Context(), "hugo", "0.139.0", ""); !errors.Is(err, ErrUnavailable) {
		t.Errorf("an unreachable index error = %v, want ErrUnavailable", err)
	}
}

func TestOneVersionIsAskedForOnceAtATime(t *testing.T) {
	var asked atomic.Int32
	hub := nixhub(t, &asked, hugoAnswer)
	nix, _ := fakeNix(t, `sleep 0.2; echo /nix/store/abc-hugo-0.139.0`)
	p := pins(t, hub, nix, &memory{pkgs: map[string]Package{}}, always)

	var wg sync.WaitGroup
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := p.Path(t.Context(), "hugo", "0.139.0", ""); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if asked.Load() != 1 {
		t.Errorf("the index was asked %d times, want once", asked.Load())
	}
}

func TestNewNeedsItsParts(t *testing.T) {
	if _, err := New(nil, Settings{}); !errors.Is(err, ErrIncomplete) {
		t.Errorf("New() error = %v, want ErrIncomplete", err)
	}
	made, err := New(&memory{}, Settings{NixBin: "nix", Roots: "/tmp"})
	if err != nil || made.settings.Nixhub != DefaultNixhub || made.settings.System == "" {
		t.Errorf("New() defaults = %+v, %v", made, err)
	}
	if !strings.Contains(System(), "-") {
		t.Errorf("System() = %q", System())
	}
}

const pinned = "02c087f3fe673c9e317f206a627cb6b910336dbe"

func TestAPinnedRevisionIsBuiltWithoutTheIndex(t *testing.T) {
	var asked atomic.Int32
	nix, log := fakeNix(t, "echo /nix/store/abc-quarto-1.10.18")
	store := &memory{pkgs: map[string]Package{}}
	p := pins(t, nixhub(t, &asked, hugoAnswer), nix, store, func(path string) bool { return path == "/nix/store/abc-quarto-1.10.18" })

	path, err := p.Path(t.Context(), "quarto", "1.10.18", pinned)
	if err != nil || path != "/nix/store/abc-quarto-1.10.18" || asked.Load() != 0 {
		t.Fatalf("Path() = %q, %v, the index asked %d times", path, err, asked.Load())
	}
	said, _ := os.ReadFile(log)
	if !strings.Contains(string(said), "github:NixOS/nixpkgs/"+pinned+"#quarto") {
		t.Errorf("nix was asked %s", said)
	}
	if _, found, _ := store.Get(t.Context(), "quarto", "1.10.18@"+pinned, "x86_64-linux"); !found {
		t.Error("the pinned build was not remembered")
	}
	if again, _ := p.Path(t.Context(), "quarto", "1.10.18", pinned); again != path || strings.Count(readAll(t, log), "\n") != 1 {
		t.Errorf("the second ask built again")
	}
}

func readAll(t *testing.T, name string) string {
	t.Helper()
	data, _ := os.ReadFile(name)
	return string(data)
}

func TestARevisionWithAnotherVersionIsRefused(t *testing.T) {
	var asked atomic.Int32
	nix, _ := fakeNix(t, "echo /nix/store/abc-quarto-1.9.37")
	p := pins(t, nixhub(t, &asked, hugoAnswer), nix, &memory{pkgs: map[string]Package{}}, always)
	_, err := p.Path(t.Context(), "quarto", "1.10.18", pinned)
	if !errors.Is(err, ErrNotFound) || !strings.Contains(err.Error(), "quarto-1.9.37") {
		t.Errorf("Path() error = %v", err)
	}
	if _, err := p.Path(t.Context(), "quarto", "1.10.18", "main"); !errors.Is(err, ErrUnusable) {
		t.Errorf("a branch name error = %v", err)
	}
}

func TestCellEnvironmentsAreBuiltFromAnExpression(t *testing.T) {
	var asked atomic.Int32
	nix, log := fakeNix(t, "echo /nix/store/env-python3-3.13-env")
	p := pins(t, nixhub(t, &asked, hugoAnswer), nix, &memory{pkgs: map[string]Package{}}, always)

	path, err := p.Environment(t.Context(), "python", pinned, []string{"pandas", "jupyter"})
	if err != nil || path != "/nix/store/env-python3-3.13-env" {
		t.Fatalf("Environment() = %q, %v", path, err)
	}
	said := readAll(t, log)
	for _, want := range []string{"--impure --expr", `getFlake "github:NixOS/nixpkgs/` + pinned + `"`, "python3.withPackages (ps: [ ps.ipykernel ps.jupyter ps.pandas ])"} {
		if !strings.Contains(said, want) {
			t.Errorf("nix was asked %s, missing %s", said, want)
		}
	}

	if _, err := p.Environment(t.Context(), "r", pinned, []string{"ggplot2"}); err != nil || !strings.Contains(readAll(t, log), "p.rWrapper.override { packages = [ p.rPackages.ggplot2 p.rPackages.knitr p.rPackages.rmarkdown ]; }") {
		t.Errorf("R environment: %v\n%s", err, readAll(t, log))
	}
	for _, bad := range [][]string{{"pandas\"; x"}, {"a b"}} {
		if _, err := p.Environment(t.Context(), "python", pinned, bad); !errors.Is(err, ErrUnusable) {
			t.Errorf("Environment(%q) error = %v", bad, err)
		}
	}
	if _, err := p.Environment(t.Context(), "julia", pinned, nil); !errors.Is(err, ErrUnusable) {
		t.Errorf("an unknown kind error = %v", err)
	}
	if _, err := p.Environment(t.Context(), "python", "", nil); !errors.Is(err, ErrUnusable) {
		t.Errorf("no revision error = %v", err)
	}
}
