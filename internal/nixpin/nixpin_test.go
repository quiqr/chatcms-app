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

	path, err := p.Path(t.Context(), "hugo", "0.139.0")
	if err != nil || path != "/nix/store/abc-hugo-0.139.0" {
		t.Fatalf("Path() = %q, %v", path, err)
	}
	called, _ := os.ReadFile(log)
	if !strings.Contains(string(called), "github:NixOS/nixpkgs/"+rev+"#hugo") || !strings.Contains(string(called), "--out-link") {
		t.Errorf("nix was asked %q", called)
	}

	again, err := p.Path(t.Context(), "hugo", "0.139.0")
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

	path, err := p.Path(t.Context(), "hugo", "0.139.0")
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

	if _, err := p.Path(t.Context(), "hugo", "0.0.1"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Path() error = %v, want ErrNotFound", err)
	}
}

func TestAnUnusableNameOrVersionIsRefusedBeforeAsking(t *testing.T) {
	var asked atomic.Int32
	p := pins(t, nixhub(t, &asked, hugoAnswer), "/bin/false", &memory{pkgs: map[string]Package{}}, always)

	for _, pair := range [][2]string{{"hugo", "1.0; rm -rf"}, {"../hugo", "1"}, {"hugo", ""}, {"", "1"}} {
		if _, err := p.Path(t.Context(), pair[0], pair[1]); !errors.Is(err, ErrUnusable) {
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

	if _, err := p.Path(t.Context(), "hugo", "0.139.0"); !errors.Is(err, ErrUnavailable) {
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
	if _, err := made.Path(t.Context(), "hugo", "0.139.0"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Path() error = %v, want ErrNotFound", err)
	}
}

func TestTroubleWithTheIndexOrNixIsUnavailable(t *testing.T) {
	var asked atomic.Int32
	broken := nixhub(t, &asked, func(w http.ResponseWriter, r *http.Request) { http.Error(w, "down", http.StatusBadGateway) })
	if _, err := pins(t, broken, "/bin/false", &memory{pkgs: map[string]Package{}}, always).Path(t.Context(), "hugo", "0.139.0"); !errors.Is(err, ErrUnavailable) {
		t.Errorf("a failing index error = %v, want ErrUnavailable", err)
	}

	garbled := nixhub(t, &asked, func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("{")) })
	if _, err := pins(t, garbled, "/bin/false", &memory{pkgs: map[string]Package{}}, always).Path(t.Context(), "hugo", "0.139.0"); !errors.Is(err, ErrUnavailable) {
		t.Errorf("a garbled answer error = %v, want ErrUnavailable", err)
	}

	failing, _ := fakeNix(t, `echo "error: cannot build" >&2; exit 1`)
	_, err := pins(t, nixhub(t, &asked, hugoAnswer), failing, &memory{pkgs: map[string]Package{}}, always).Path(t.Context(), "hugo", "0.139.0")
	if !errors.Is(err, ErrUnavailable) || !strings.Contains(err.Error(), "cannot build") {
		t.Errorf("a failing nix error = %v, want ErrUnavailable with its words", err)
	}

	silent, _ := fakeNix(t, `echo nothing`)
	if _, err := pins(t, nixhub(t, &asked, hugoAnswer), silent, &memory{pkgs: map[string]Package{}}, always).Path(t.Context(), "hugo", "0.139.0"); !errors.Is(err, ErrUnavailable) {
		t.Errorf("a nix naming no path error = %v, want ErrUnavailable", err)
	}

	if _, err := pins(t, "http://127.0.0.1:1", "/bin/false", &memory{pkgs: map[string]Package{}}, always).Path(t.Context(), "hugo", "0.139.0"); !errors.Is(err, ErrUnavailable) {
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
			if _, err := p.Path(t.Context(), "hugo", "0.139.0"); err != nil {
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
