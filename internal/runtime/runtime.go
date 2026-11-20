package runtime

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/goccy/go-yaml"
)

const File = "quiqr/runtime.yaml"

const (
	Web   = "web"
	PDF   = "pdf"
	Other = "other"
)

var (
	ErrInvalid     = errors.New("runtime: the runtime definition is not usable")
	ErrNoProgram   = errors.New("runtime: no declared package provides this program")
	ErrNoPackages  = errors.New("runtime: this installation cannot provide packages")
	ErrNoHugo      = errors.New("runtime: no Hugo to fall back to")
	placeholders   = map[string]bool{"{host}": true, "{port}": true, "{url}": true, "{reload}": true, "{source}": true, "{output}": true, "{cache}": true}
	placeholderish = regexp.MustCompile(`^\{[a-z]+\}$`)
	packageName    = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._+-]*$`)
	versionName    = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._+-]*$`)
	revisionName   = regexp.MustCompile(`^[0-9a-f]{40}$`)
	cellPackage    = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
)

type Package struct {
	Version string `yaml:"version"`
	Nixpkgs string `yaml:"nixpkgs"`
}

func (p *Package) UnmarshalYAML(data []byte) error {
	var version string
	if err := yaml.Unmarshal(data, &version); err == nil {
		*p = Package{Version: version}
		return nil
	}
	type plain Package
	var whole plain
	if err := yaml.UnmarshalWithOptions(data, &whole, yaml.DisallowUnknownField()); err != nil {
		return err
	}
	*p = Package(whole)
	return nil
}

type Environment struct {
	Nixpkgs  string   `yaml:"nixpkgs"`
	Packages []string `yaml:"packages"`
}

const (
	Python = "python"
	R      = "r"
)

type Preview struct {
	Type    string   `yaml:"type"`
	Open    string   `yaml:"open"`
	Command []string `yaml:"command"`
}

type Build struct {
	Output  string   `yaml:"output"`
	Command []string `yaml:"command"`
}

type Definition struct {
	Packages map[string]Package `yaml:"packages"`
	Python   *Environment       `yaml:"python"`
	R        *Environment       `yaml:"r"`
	Preview  Preview            `yaml:"preview"`
	Build    *Build             `yaml:"build"`
}

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalid, fmt.Sprintf(format, args...))
}

func Parse(data []byte) (Definition, error) {
	var definition Definition
	if err := yaml.UnmarshalWithOptions(data, &definition, yaml.DisallowUnknownField()); err != nil {
		return Definition{}, invalid("%s cannot be read: %v", File, err)
	}
	if err := definition.check(); err != nil {
		return Definition{}, err
	}
	return definition, nil
}

func checkCommand(name string, command []string) error {
	if len(command) == 0 {
		return invalid("the %s command is empty", name)
	}
	if command[0] == "" || strings.ContainsAny(command[0], "/{} \t\n") {
		return invalid("the %s command must start with a program's name, not %q", name, command[0])
	}
	for _, word := range command {
		if placeholderish.MatchString(word) && !placeholders[word] {
			return invalid("the %s command uses %s, which is not one of {host}, {port}, {url}, {reload}, {source}, {output} or {cache}", name, word)
		}
		if strings.ContainsRune(word, 0) {
			return invalid("the %s command holds a word with a NUL", name)
		}
	}
	return nil
}

func (d *Definition) check() error {
	if len(d.Packages) == 0 {
		return invalid("no packages are named")
	}
	for name, pkg := range d.Packages {
		if !packageName.MatchString(name) {
			return invalid("%q is not a package name", name)
		}
		if !versionName.MatchString(pkg.Version) {
			return invalid("%q is not a version of %s", pkg.Version, name)
		}
		if pkg.Nixpkgs != "" && !revisionName.MatchString(pkg.Nixpkgs) {
			return invalid("%q is not a nixpkgs revision for %s; give the whole 40-character commit", pkg.Nixpkgs, name)
		}
	}
	for kind, env := range map[string]*Environment{Python: d.Python, R: d.R} {
		if env == nil {
			continue
		}
		if !revisionName.MatchString(env.Nixpkgs) {
			return invalid("%s needs the nixpkgs revision to build its packages from, as the whole 40-character commit", kind)
		}
		for _, name := range env.Packages {
			if !cellPackage.MatchString(name) {
				return invalid("%q is not a %s package name", name, kind)
			}
		}
	}
	if d.Preview.Type == "" {
		d.Preview.Type = Web
	}
	switch d.Preview.Type {
	case Web, PDF, Other:
	default:
		return invalid("the preview type %q is not one of web, pdf or other", d.Preview.Type)
	}
	if d.Preview.Open == "" {
		d.Preview.Open = "/"
	}
	if !strings.HasPrefix(d.Preview.Open, "/") || strings.Contains(d.Preview.Open, "..") {
		return invalid("the page to open %q must start with / and stay inside the preview", d.Preview.Open)
	}
	if err := checkCommand("preview", d.Preview.Command); err != nil {
		return err
	}
	if d.Build != nil {
		if err := checkCommand("build", d.Build.Command); err != nil {
			return err
		}
		output := filepath.Clean(d.Build.Output)
		if d.Build.Output == "" || filepath.IsAbs(output) || output == "." || strings.HasPrefix(output, "..") {
			return invalid("the build output %q must be a directory inside the site", d.Build.Output)
		}
		d.Build.Output = filepath.ToSlash(output)
	}
	return nil
}

func Default(hugoVersion string) Definition {
	packages := map[string]Package{"hugo": {Version: hugoVersion}}
	return Definition{
		Packages: packages,
		Preview: Preview{
			Type: Web,
			Open: "/",
			Command: []string{
				"hugo", "server",
				"--source", "{source}",
				"--bind", "{host}",
				"--port", "{port}",
				"--baseURL", "{url}",
				"--appendPort=false",
				"--renderToMemory",
				"--buildDrafts",
				"--noBuildLock",
				"--cacheDir", "{cache}",
				"--liveReloadPort", "{reload}",
				"--poll", "500ms",
			},
		},
	}
}

func Read(site fs.FS, hugoVersion string) (Definition, bool, error) {
	data, err := fs.ReadFile(site, File)
	if errors.Is(err, fs.ErrNotExist) {
		return Default(hugoVersion), false, nil
	}
	if err != nil {
		return Definition{}, false, invalid("%s cannot be read: %v", File, err)
	}
	definition, err := Parse(data)
	return definition, true, err
}

type Command struct {
	Program string   `json:"program"`
	Args    []string `json:"args"`
}

type Resolved struct {
	Path     []string          `json:"path"`
	Preview  Command           `json:"preview"`
	Type     string            `json:"type"`
	Open     string            `json:"open"`
	Build    *Command          `json:"build,omitempty"`
	Output   string            `json:"output,omitempty"`
	Declared bool              `json:"declared"`
	Env      map[string]string `json:"env,omitempty"`
	Rooted   bool              `json:"rooted,omitempty"`
	Hugo     *HugoChoice       `json:"hugo,omitempty"`
}

type HugoChoice struct {
	Asked   string `json:"asked"`
	Program string `json:"program"`
	Because string `json:"because"`
}

const (
	HugoPinned     = "pinned"
	HugoPinningOff = "pinning-off"
	HugoNoVersion  = "no-version"
	HugoProvided   = "provided"
)

type Pins interface {
	Path(ctx context.Context, name, version, revision string) (string, error)
	Environment(ctx context.Context, kind, revision string, packages []string) (string, error)
}

type Resolver struct {
	Pins          Pins
	InstalledHugo string
	Provided      map[string]string
	Exists        func(path string) bool
}

func providedAs(path, version string) bool {
	return version != "" && strings.HasSuffix(filepath.Base(path), "-"+version)
}

func (r Resolver) exists(path string) bool {
	if r.Exists != nil {
		return r.Exists(path)
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func (r Resolver) Resolve(ctx context.Context, definition Definition, declared bool) (Resolved, error) {
	names := make([]string, 0, len(definition.Packages))
	for name := range definition.Packages {
		names = append(names, name)
	}
	sort.Strings(names)

	var dirs []string
	var hugo *HugoChoice
	chose := func(name, version, dir, because string) {
		if name == "hugo" {
			hugo = &HugoChoice{Asked: version, Program: filepath.Join(dir, "hugo"), Because: because}
		}
	}
	for _, name := range names {
		pkg := definition.Packages[name]
		version := pkg.Version
		if provided, found := r.Provided[name]; found && providedAs(provided, version) {
			dirs = append(dirs, filepath.Join(provided, "bin"))
			chose(name, version, filepath.Join(provided, "bin"), HugoProvided)
			continue
		}
		if name == "hugo" && (version == "" || r.Pins == nil) {
			if r.InstalledHugo == "" {
				return Resolved{}, ErrNoHugo
			}
			dirs = append(dirs, filepath.Dir(r.InstalledHugo))
			because := HugoPinningOff
			if version == "" {
				because = HugoNoVersion
			}
			hugo = &HugoChoice{Asked: version, Program: r.InstalledHugo, Because: because}
			continue
		}
		if r.Pins == nil {
			return Resolved{}, fmt.Errorf("%w: %s %s", ErrNoPackages, name, version)
		}
		path, err := r.Pins.Path(ctx, name, version, pkg.Nixpkgs)
		if err != nil {
			return Resolved{}, err
		}
		dirs = append(dirs, filepath.Join(path, "bin"))
		chose(name, version, filepath.Join(path, "bin"), HugoPinned)
	}

	env := map[string]string{}
	for _, cells := range []struct {
		kind, variable, program string
		env                     *Environment
	}{{Python, "QUARTO_PYTHON", "python3", definition.Python}, {R, "QUARTO_R", "R", definition.R}} {
		if cells.env == nil {
			continue
		}
		if r.Pins == nil {
			return Resolved{}, fmt.Errorf("%w: %s packages", ErrNoPackages, cells.kind)
		}
		path, err := r.Pins.Environment(ctx, cells.kind, cells.env.Nixpkgs, cells.env.Packages)
		if err != nil {
			return Resolved{}, err
		}
		env[cells.variable] = filepath.Join(path, "bin", cells.program)
	}

	find := func(program string) (string, error) {
		for _, dir := range dirs {
			candidate := filepath.Join(dir, program)
			if r.exists(candidate) {
				return candidate, nil
			}
		}
		return "", fmt.Errorf("%w: %s", ErrNoProgram, program)
	}

	resolved := Resolved{Path: dirs, Type: definition.Preview.Type, Open: definition.Preview.Open, Declared: declared, Rooted: !slices.Contains(definition.Preview.Command, "{url}"), Hugo: hugo}
	if len(env) > 0 {
		resolved.Env = env
	}
	program, err := find(definition.Preview.Command[0])
	if err != nil {
		return Resolved{}, err
	}
	resolved.Preview = Command{Program: program, Args: append([]string{}, definition.Preview.Command[1:]...)}

	if definition.Build != nil {
		program, err := find(definition.Build.Command[0])
		if err != nil {
			return Resolved{}, err
		}
		resolved.Build = &Command{Program: program, Args: append([]string{}, definition.Build.Command[1:]...)}
		resolved.Output = definition.Build.Output
	}
	return resolved, nil
}

func Fill(words []string, values map[string]string) []string {
	filled := make([]string, 0, len(words))
	for i := 0; i < len(words); i++ {
		word := words[i]
		value, known := values[word]
		if !placeholders[word] {
			filled = append(filled, word)
			continue
		}
		if known && value != "" {
			filled = append(filled, value)
			continue
		}
		if len(filled) > 0 && strings.HasPrefix(filled[len(filled)-1], "-") {
			filled = filled[:len(filled)-1]
		}
	}
	return filled
}
