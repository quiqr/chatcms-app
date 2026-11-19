package workspace

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/quiqr/chatcms-app/internal/runtime"
)

var ErrNotInStore = errors.New("workspace: the preview's program is not in the package store, so the workspace cannot see it")

type RunOptions struct {
	HugoBin string
	Store   string
	Host    string
	Port    int
	Source  string
	Cache   string
}

func InStore(path, store string) (string, error) {
	if !filepath.IsAbs(path) {
		found, err := exec.LookPath(path)
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrNotInStore, err)
		}
		path = found
	}
	real, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrNotInStore, err)
	}
	if !strings.HasPrefix(real, store+"/") {
		return "", fmt.Errorf("%w: %s", ErrNotInStore, real)
	}
	return real, nil
}

func EncodeRun(spec Spec, options RunOptions) (string, error) {
	resolved := spec.Run
	if resolved.Preview.Program == "" {
		if options.HugoBin == "" {
			return "", ErrNotInStore
		}
		hugo, err := InStore(options.HugoBin, options.Store)
		if err != nil {
			return "", err
		}
		resolved, err = runtime.Resolver{InstalledHugo: hugo, Exists: func(string) bool { return true }}.Resolve(context.Background(), runtime.Default(""), false)
		if err != nil {
			return "", err
		}
	}
	program, err := InStore(resolved.Preview.Program, options.Store)
	if err != nil {
		return "", err
	}
	path := make([]string, 0, len(resolved.Path))
	for _, dir := range resolved.Path {
		real, err := filepath.EvalSymlinks(dir)
		if err != nil || !strings.HasPrefix(real, options.Store+"/") {
			return "", fmt.Errorf("%w: %s", ErrNotInStore, dir)
		}
		path = append(path, real)
	}

	values := map[string]string{
		"{host}":   options.Host,
		"{port}":   strconv.Itoa(options.Port),
		"{url}":    spec.Under,
		"{source}": options.Source,
		"{cache}":  options.Cache,
	}
	if port := BrowserPort(spec.Under); port > 0 {
		values["{reload}"] = strconv.Itoa(port)
	}
	env := map[string]string{}
	for name, value := range resolved.Env {
		real, err := InStore(value, options.Store)
		if err != nil {
			return "", err
		}
		env[name] = real
	}
	data, err := json.Marshal(map[string]any{
		"cmd":  append([]string{program}, runtime.Fill(resolved.Preview.Args, values)...),
		"path": path,
		"env":  env,
	})
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

func BrowserPort(under string) int {
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
