package sitebuild

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/quiqr/chatcms-app/internal/runner"
	"github.com/quiqr/chatcms-app/internal/runtime"
)

const (
	Output             = "public"
	DefaultTimeout     = 10 * time.Minute
	DefaultMemoryBytes = 2 << 30
	cacheInside        = "/tmp/hugo-cache"
	defaultEnvironment = "production"
)

type Limits struct {
	Timeout     time.Duration
	MemoryBytes int64
}

func (l Limits) withDefaults() Limits {
	if l.Timeout <= 0 {
		l.Timeout = DefaultTimeout
	}
	if l.MemoryBytes <= 0 {
		l.MemoryBytes = DefaultMemoryBytes
	}
	return l
}

type Options struct {
	BaseURL     string
	Environment string
	Drafts      bool
	Project     uuid.UUID
}

type Builder struct {
	runner     runner.Runner
	hugo       string
	limits     Limits
	runtimeFor func(ctx context.Context, project uuid.UUID, site fs.FS) (runtime.Resolved, error)
}

func (b *Builder) UseRuntimeOf(runtimeFor func(ctx context.Context, project uuid.UUID, site fs.FS) (runtime.Resolved, error)) {
	b.runtimeFor = runtimeFor
}

var (
	ErrNoHugo  = errors.New("sitebuild: no Hugo was named")
	ErrNoBuild = errors.New("sitebuild: this site's runtime definition declares no build and no Hugo")
)

func New(r runner.Runner, hugoBin string, limits Limits) (*Builder, error) {
	if r == nil {
		return nil, errors.New("sitebuild: no runner")
	}
	if hugoBin == "" {
		return nil, ErrNoHugo
	}
	return &Builder{runner: r, hugo: hugoBin, limits: limits.withDefaults()}, nil
}

func (b *Builder) resolved(ctx context.Context, site fs.FS, project uuid.UUID) (runtime.Resolved, error) {
	if b.runtimeFor != nil && project != uuid.Nil {
		return b.runtimeFor(ctx, project, site)
	}
	return runtime.Resolved{Path: []string{filepath.Dir(b.hugo)}, Preview: runtime.Command{Program: b.hugo}}, nil
}

func findHugo(resolved runtime.Resolved, fallback string) string {
	if filepath.Base(resolved.Preview.Program) == "hugo" {
		return resolved.Preview.Program
	}
	for _, dir := range resolved.Path {
		candidate := filepath.Join(dir, "hugo")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
	}
	if !resolved.Declared {
		return fallback
	}
	return ""
}

func (b *Builder) Build(ctx context.Context, site fs.FS, opts Options) (*runner.Result, error) {
	if site == nil {
		return nil, errors.New("sitebuild: no site to build")
	}
	resolved, err := b.resolved(ctx, site, opts.Project)
	if err != nil {
		return nil, err
	}

	environment := opts.Environment
	if environment == "" {
		environment = defaultEnvironment
	}

	task := runner.Task{
		Inputs: site,
		Environment: map[string]string{
			"HUGO_ENV":         environment,
			"HUGO_ENVIRONMENT": environment,
		},
		Tools:       append([]string{}, resolved.Path...),
		Timeout:     b.limits.Timeout,
		MemoryBytes: b.limits.MemoryBytes,
	}
	root := Output

	if resolved.Build != nil {
		root = resolved.Output
		task.Command = resolved.Build.Program
		task.Arguments = runtime.Fill(resolved.Build.Args, map[string]string{
			"{output}": resolved.Output,
			"{cache}":  cacheInside,
			"{url}":    opts.BaseURL,
			"{source}": ".",
		})
	} else {
		hugo := findHugo(resolved, b.hugo)
		if hugo == "" {
			return nil, ErrNoBuild
		}
		task.Command = hugo
		task.Arguments = []string{
			"--destination", Output,
			"--cleanDestinationDir",
			"--cacheDir", cacheInside,
			"--environment", environment,
			"--logLevel", "info",
			"--noBuildLock",
		}
		if opts.BaseURL != "" {
			task.Arguments = append(task.Arguments, "--baseURL", opts.BaseURL)
		}
		if opts.Drafts {
			task.Arguments = append(task.Arguments, "--buildDrafts")
		}
	}
	task.Tools = append(task.Tools, task.Command)
	task.Outputs = []string{root}

	result, err := b.runner.Run(ctx, task)
	var missing *runner.MissingOutputError
	if errors.As(err, &missing) {
		return nil, fmt.Errorf("sitebuild: the build reported success but wrote no %s", root)
	}
	if result != nil {
		result.Root = root
	}
	return result, err
}

// Pages is the built site itself rather than the folder holding it. A caller
// publishing what was built wants the pages at the root, which is where they
// land on a server.
func Pages(result *runner.Result) fs.FS {
	if result == nil || result.Outputs == nil {
		return nil
	}
	root := result.Root
	if root == "" {
		root = Output
	}
	pages, err := fs.Sub(result.Outputs, root)
	if err != nil {
		return result.Outputs
	}
	return pages
}
