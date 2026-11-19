package nixpin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	ErrNotFound    = errors.New("nixpin: no such version")
	ErrUnusable    = errors.New("nixpin: this name or version cannot be used")
	ErrUnavailable = errors.New("nixpin: the package could not be made available")
	ErrIncomplete  = errors.New("nixpin: missing a part it needs")
)

const DefaultNixhub = "https://search.devbox.sh"

type Package struct {
	Name      string
	Version   string
	System    string
	Revision  string
	AttrPath  string
	StorePath string
}

type Store interface {
	Get(ctx context.Context, name, version, system string) (Package, bool, error)
	Put(ctx context.Context, pkg Package) error
}

type Settings struct {
	Nixhub string
	NixBin string
	Roots  string
	System string
	Client *http.Client
	Exists func(path string) bool
}

type Pins struct {
	settings Settings
	store    Store

	mu      sync.Mutex
	pending map[string]*sync.Mutex
}

func New(store Store, settings Settings) (*Pins, error) {
	if store == nil || settings.NixBin == "" || settings.Roots == "" {
		return nil, ErrIncomplete
	}
	if settings.Nixhub == "" {
		settings.Nixhub = DefaultNixhub
	}
	if settings.System == "" {
		settings.System = System()
	}
	if settings.Client == nil {
		settings.Client = &http.Client{Timeout: 30 * time.Second}
	}
	if settings.Exists == nil {
		settings.Exists = func(path string) bool {
			_, err := os.Stat(path)
			return err == nil
		}
	}
	return &Pins{settings: settings, store: store, pending: map[string]*sync.Mutex{}}, nil
}

func System() string {
	arch := map[string]string{"amd64": "x86_64", "arm64": "aarch64"}[runtime.GOARCH]
	if arch == "" {
		arch = runtime.GOARCH
	}
	return arch + "-" + runtime.GOOS
}

var usable = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._+-]*$`)

func (p *Pins) lockFor(key string) *sync.Mutex {
	p.mu.Lock()
	defer p.mu.Unlock()
	lock, found := p.pending[key]
	if !found {
		lock = &sync.Mutex{}
		p.pending[key] = lock
	}
	return lock
}

func (p *Pins) Path(ctx context.Context, name, version, rev string) (string, error) {
	if !usable.MatchString(name) || !usable.MatchString(version) || (rev != "" && !revision.MatchString(rev)) {
		return "", fmt.Errorf("%w: %s %s", ErrUnusable, name, version)
	}
	if rev == "" {
		return p.kept(ctx, name, version, func() (Package, error) { return p.resolve(ctx, name, version) }, p.realise)
	}
	path, err := p.kept(ctx, name, version+"@"+rev, func() (Package, error) {
		return Package{Name: name, Version: version + "@" + rev, System: p.settings.System, Revision: rev, AttrPath: name}, nil
	}, p.realise)
	if err != nil {
		return "", err
	}
	if !strings.HasSuffix(filepath.Base(path), "-"+version) {
		_, held, _ := strings.Cut(filepath.Base(path), "-")
		return "", fmt.Errorf("%w: nixpkgs %s holds %s, not %s %s", ErrNotFound, rev[:12], held, name, version)
	}
	return path, nil
}

var cellPackage = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

var cellBasics = map[string][]string{
	"python": {"jupyter", "ipykernel"},
	"r":      {"knitr", "rmarkdown"},
}

func cellExpression(kind, rev, system string, packages []string) (string, error) {
	wanted := append(append([]string{}, cellBasics[kind]...), packages...)
	sort.Strings(wanted)
	wanted = slices.Compact(wanted)
	for _, name := range wanted {
		if !cellPackage.MatchString(name) {
			return "", fmt.Errorf("%w: %s package %q", ErrUnusable, kind, name)
		}
	}
	pkgs := fmt.Sprintf("(builtins.getFlake \"github:NixOS/nixpkgs/%s\").legacyPackages.%s", rev, system)
	switch kind {
	case "python":
		return fmt.Sprintf("%s.python3.withPackages (ps: [ ps.%s ])", pkgs, strings.Join(wanted, " ps.")), nil
	case "r":
		return fmt.Sprintf("let p = %s; in p.rWrapper.override { packages = [ p.rPackages.%s ]; }", pkgs, strings.Join(wanted, " p.rPackages.")), nil
	}
	return "", fmt.Errorf("%w: %q is not python or r", ErrUnusable, kind)
}

func (p *Pins) Environment(ctx context.Context, kind, rev string, packages []string) (string, error) {
	if !revision.MatchString(rev) {
		return "", fmt.Errorf("%w: %s needs a whole nixpkgs revision", ErrUnusable, kind)
	}
	expression, err := cellExpression(kind, rev, p.settings.System, packages)
	if err != nil {
		return "", err
	}
	listed := append([]string{}, packages...)
	sort.Strings(listed)
	version := strings.Join(listed, "+") + "@" + rev
	return p.kept(ctx, kind+"-cells", version, func() (Package, error) {
		return Package{Name: kind + "-cells", Version: version, System: p.settings.System, Revision: rev, AttrPath: expression}, nil
	}, p.realiseExpression)
}

func (p *Pins) kept(ctx context.Context, name, version string, find func() (Package, error), realise func(context.Context, Package) (string, error)) (string, error) {
	system := p.settings.System
	lock := p.lockFor(name + "\x00" + version)
	lock.Lock()
	defer lock.Unlock()

	known, found, err := p.store.Get(ctx, name, version, system)
	if err != nil {
		return "", err
	}
	if found && p.settings.Exists(known.StorePath) {
		return known.StorePath, nil
	}

	pkg := known
	if !found {
		pkg, err = find()
		if err != nil {
			return "", err
		}
	}
	pkg.StorePath, err = realise(ctx, pkg)
	if err != nil {
		return "", err
	}
	if err := p.store.Put(ctx, pkg); err != nil {
		return "", err
	}
	return pkg.StorePath, nil
}

type resolved struct {
	Systems map[string]struct {
		FlakeInstallable struct {
			Ref struct {
				Owner string `json:"owner"`
				Repo  string `json:"repo"`
				Rev   string `json:"rev"`
			} `json:"ref"`
			AttrPath string `json:"attr_path"`
		} `json:"flake_installable"`
	} `json:"systems"`
}

var revision = regexp.MustCompile(`^[0-9a-f]{40}$`)

func (p *Pins) resolve(ctx context.Context, name, version string) (Package, error) {
	where := strings.TrimSuffix(p.settings.Nixhub, "/") + "/v2/resolve?" + url.Values{"name": {name}, "version": {version}}.Encode()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, where, nil)
	if err != nil {
		return Package{}, err
	}
	request.Header.Set("User-Agent", "chatcms")
	answer, err := p.settings.Client.Do(request)
	if err != nil {
		return Package{}, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer answer.Body.Close()
	if answer.StatusCode == http.StatusNotFound {
		return Package{}, fmt.Errorf("%w: %s %s", ErrNotFound, name, version)
	}
	if answer.StatusCode != http.StatusOK {
		return Package{}, fmt.Errorf("%w: the package index answered %d", ErrUnavailable, answer.StatusCode)
	}

	var body resolved
	if err := json.NewDecoder(answer.Body).Decode(&body); err != nil {
		return Package{}, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	found, ok := body.Systems[p.settings.System]
	if !ok {
		return Package{}, fmt.Errorf("%w: %s %s for %s", ErrNotFound, name, version, p.settings.System)
	}
	ref := found.FlakeInstallable.Ref
	if ref.Owner != "NixOS" || ref.Repo != "nixpkgs" || !revision.MatchString(ref.Rev) || !usable.MatchString(found.FlakeInstallable.AttrPath) {
		return Package{}, fmt.Errorf("%w: the package index named something unexpected", ErrUnavailable)
	}
	return Package{
		Name:     name,
		Version:  version,
		System:   p.settings.System,
		Revision: ref.Rev,
		AttrPath: found.FlakeInstallable.AttrPath,
	}, nil
}

func (p *Pins) realise(ctx context.Context, pkg Package) (string, error) {
	return p.build(ctx, pkg, "github:NixOS/nixpkgs/"+pkg.Revision+"#"+pkg.AttrPath)
}

func (p *Pins) realiseExpression(ctx context.Context, pkg Package) (string, error) {
	return p.build(ctx, pkg, "--impure", "--expr", pkg.AttrPath)
}

func (p *Pins) build(ctx context.Context, pkg Package, what ...string) (string, error) {
	if err := os.MkdirAll(p.settings.Roots, 0o755); err != nil {
		return "", err
	}
	link := filepath.Join(p.settings.Roots, strings.ReplaceAll(pkg.Name+"-"+pkg.Version, "/", "_"))
	args := append([]string{"--extra-experimental-features", "nix-command flakes",
		"build", "--print-out-paths", "--out-link", link}, what...)
	command := exec.CommandContext(ctx, p.settings.NixBin, args...)
	out, err := command.Output()
	if err != nil {
		detail := ""
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			detail = strings.TrimSpace(string(exit.Stderr))
		}
		return "", fmt.Errorf("%w: %s %s: %v %s", ErrUnavailable, pkg.Name, pkg.Version, err, detail)
	}
	lines := strings.Fields(strings.TrimSpace(string(out)))
	if len(lines) == 0 || !strings.HasPrefix(lines[0], "/nix/store/") {
		return "", fmt.Errorf("%w: nix named no store path", ErrUnavailable)
	}
	return lines[0], nil
}
