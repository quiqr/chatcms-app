package gitsync

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestHostMakesAnEmptyBareRepository(t *testing.T) {
	cfg := testConfig(t)
	state := t.TempDir()
	sites := NewSites(cfg, filepath.Join(state, "sites"), nil)

	path, err := sites.Host(t.Context(), "korenaar")
	if err != nil {
		t.Fatalf("Host() error = %v, want nil", err)
	}
	if want := filepath.Join(state, "hosted", "korenaar.git"); path != want {
		t.Errorf("Host() = %q, want %q", path, want)
	}
	if got := strings.TrimSpace(run(t, cfg, path, "rev-parse", "--is-bare-repository")); got != "true" {
		t.Errorf("bare = %q, want true", got)
	}
	if got := strings.TrimSpace(run(t, cfg, path, "for-each-ref")); got != "" {
		t.Errorf("refs = %q, want none", got)
	}

	if _, err := sites.Host(t.Context(), "korenaar"); !errors.Is(err, ErrAlreadyAWorkingCopy) {
		t.Errorf("second Host() error = %v, want ErrAlreadyAWorkingCopy", err)
	}
	if _, err := sites.Host(t.Context(), "../x"); !errors.Is(err, ErrUnusableSiteID) {
		t.Errorf("Host(../x) error = %v, want ErrUnusableSiteID", err)
	}

	if err := sites.Unhost("korenaar"); err != nil {
		t.Fatalf("Unhost() error = %v", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("after Unhost the repository is still there: %v", err)
	}
}

func hostedSite(t *testing.T) (Repo, *Sites, Config) {
	t.Helper()
	cfg := testConfig(t)
	sites := NewSites(cfg, filepath.Join(t.TempDir(), "sites"), nil)

	kept, err := sites.Host(t.Context(), "korenaar")
	if err != nil {
		t.Fatal(err)
	}
	repository, err := sites.Create(t.Context(), "korenaar", kept,
		fstest.MapFS{"content/_index.md": {Data: []byte("# Home\n")}}, editor("Anna"), "Start")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { repository.Close() })
	if _, err := repository.Publish(t.Context()); err != nil {
		t.Fatal(err)
	}
	save(t, repository, "Anna", "Second", map[string]string{"content/a.md": "a\n"})
	save(t, repository, "Anna", "Third", map[string]string{"content/b.md": "b\n"})
	return repository, sites, cfg
}

func TestMoveToSendsEveryVersion(t *testing.T) {
	repository, _, cfg := hostedSite(t)
	target := emptyBareRemote(t, cfg)

	if err := repository.MoveTo(t.Context(), target); err != nil {
		t.Fatalf("MoveTo() error = %v, want nil", err)
	}

	subjects := strings.Fields(run(t, cfg, target, "log", "--format=%s", "main"))
	if strings.Join(subjects, ",") != "Third,Second,Start" {
		t.Errorf("target holds %v, want Third, Second, Start", subjects)
	}

	status, err := repository.Status(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if status.Unpublished != 0 {
		t.Errorf("unpublished after moving = %d, want 0", status.Unpublished)
	}
}

func TestMoveToAnUnreachableRemoteLeavesOriginAlone(t *testing.T) {
	repository, sites, cfg := hostedSite(t)
	dir, _ := sites.Dir("korenaar")
	before := run(t, cfg, dir, "remote", "get-url", "origin")

	err := repository.MoveTo(t.Context(), filepath.Join(t.TempDir(), "nowhere.git"))
	if err == nil {
		t.Fatal("MoveTo() error = nil, want a failure")
	}
	if after := run(t, cfg, dir, "remote", "get-url", "origin"); after != before {
		t.Errorf("origin = %q after a failed move, want %q", after, before)
	}
}

func TestMoveToGivesASiteWithoutOriginOne(t *testing.T) {
	cfg := testConfig(t)
	sites := NewSites(cfg, filepath.Join(t.TempDir(), "sites"), nil)
	repository, err := sites.Create(t.Context(), "korenaar", "",
		fstest.MapFS{"a.md": {Data: []byte("a\n")}}, editor("Anna"), "Start")
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	dir, _ := sites.Dir("korenaar")

	if err := repository.MoveTo(t.Context(), filepath.Join(t.TempDir(), "nowhere.git")); err == nil {
		t.Fatal("MoveTo() error = nil, want a failure")
	}
	if out := run(t, cfg, dir, "remote"); strings.TrimSpace(out) != "" {
		t.Errorf("remotes = %q after a failed move, want none", out)
	}

	target := emptyBareRemote(t, cfg)
	if err := repository.MoveTo(t.Context(), target); err != nil {
		t.Fatalf("MoveTo() error = %v", err)
	}
	if got := strings.TrimSpace(run(t, cfg, target, "log", "--format=%s", "main")); got != "Start" {
		t.Errorf("target holds %q, want Start", got)
	}
}
