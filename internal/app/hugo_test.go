package app

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/httpapi"
	"github.com/quiqr/chatcms-app/internal/nixpin"
)

type fakePins struct {
	path string
	err  error
}

func (f fakePins) Path(ctx context.Context, name, version string) (string, error) {
	return f.path, f.err
}

func versionIs(version string) func(context.Context, uuid.UUID) (string, error) {
	return func(context.Context, uuid.UUID) (string, error) { return version, nil }
}

func TestAProjectWithoutAVersionUsesTheInstallationsHugo(t *testing.T) {
	h := hugos{installed: "/usr/bin/hugo", versionOf: versionIs(""), pins: fakePins{path: "/nix/store/x"}}
	if got, err := h.For(t.Context(), uuid.New()); err != nil || got != "/usr/bin/hugo" {
		t.Errorf("For() = %q, %v", got, err)
	}
}

func TestAProjectWithAVersionUsesThatHugo(t *testing.T) {
	h := hugos{installed: "/usr/bin/hugo", versionOf: versionIs("0.139.0"), pins: fakePins{path: "/nix/store/x-hugo-0.139.0"}}
	if got, err := h.For(t.Context(), uuid.New()); err != nil || got != "/nix/store/x-hugo-0.139.0/bin/hugo" {
		t.Errorf("For() = %q, %v", got, err)
	}
}

func TestAVersionThatCannotBeHadIsRefusedNamingIt(t *testing.T) {
	for err, key := range map[error]string{
		nixpin.ErrNotFound:    "refused.hugo-version-not-found",
		nixpin.ErrUnusable:    "refused.hugo-version-not-found",
		nixpin.ErrUnavailable: "refused.hugo-version-unavailable",
	} {
		h := hugos{installed: "/usr/bin/hugo", versionOf: versionIs("0.0.1"), pins: fakePins{err: err}}
		_, got := h.For(t.Context(), uuid.New())
		var refused *httpapi.RefusedError
		if !errors.As(got, &refused) || refused.Key != key || refused.Detail["version"] != "0.0.1" {
			t.Errorf("For() with %v = %v, want %s naming the version", err, got, key)
		}
	}
}

func TestAProjectThatCannotBeReadIsAnError(t *testing.T) {
	broken := errors.New("no project")
	h := hugos{versionOf: func(context.Context, uuid.UUID) (string, error) { return "", broken }}
	if _, err := h.For(t.Context(), uuid.New()); !errors.Is(err, broken) {
		t.Errorf("For() error = %v", err)
	}
}

func TestASiteWithoutARuntimeDefinitionGetsTheProjectsHugo(t *testing.T) {
	store := t.TempDir()
	hugo := filepath.Join(store, "x-hugo", "bin", "hugo")
	os.MkdirAll(filepath.Dir(hugo), 0o755)
	os.WriteFile(hugo, []byte("#!/bin/sh\n"), 0o755)

	h := hugos{installed: "/usr/bin/hugo", versionOf: versionIs("0.139.0"), pins: fakePins{path: filepath.Dir(filepath.Dir(hugo))}}
	resolved, err := h.Runtime(t.Context(), uuid.New(), fstest.MapFS{})
	if err != nil || resolved.Preview.Program != hugo || resolved.Declared {
		t.Errorf("Runtime() = %+v, %v", resolved, err)
	}
}

func TestADeclaredRuntimeIsResolvedFromTheSite(t *testing.T) {
	store := t.TempDir()
	quarto := filepath.Join(store, "q-quarto", "bin", "quarto")
	os.MkdirAll(filepath.Dir(quarto), 0o755)
	os.WriteFile(quarto, []byte("#!/bin/sh\n"), 0o755)
	site := fstest.MapFS{"quiqr/runtime.yaml": {Data: []byte("packages: {quarto: \"1.6.40\"}\npreview: {type: pdf, open: /book.pdf, command: [quarto, preview]}\n")}}

	h := hugos{installed: "/usr/bin/hugo", versionOf: versionIs(""), pins: fakePins{path: filepath.Dir(filepath.Dir(quarto))}}
	resolved, err := h.Runtime(t.Context(), uuid.New(), site)
	if err != nil || resolved.Preview.Program != quarto || resolved.Type != "pdf" || resolved.Open != "/book.pdf" || !resolved.Declared {
		t.Errorf("Runtime() = %+v, %v", resolved, err)
	}
}

func TestAWrongRuntimeIsRefusedSayingWhy(t *testing.T) {
	broken := fstest.MapFS{"quiqr/runtime.yaml": {Data: []byte("packages: {quarto: \"1\"}\npreview: {type: video, command: [quarto]}\n")}}
	quarto := fstest.MapFS{"quiqr/runtime.yaml": {Data: []byte("packages: {quarto: \"1\"}\npreview: {command: [quarto]}\n")}}
	for name, case_ := range map[string]struct {
		site fs.FS
		h    hugos
		key  string
	}{
		"an unknown type":    {broken, hugos{installed: "/usr/bin/hugo", versionOf: versionIs("")}, "refused.runtime-definition"},
		"no Nix for quarto":  {quarto, hugos{installed: "/usr/bin/hugo", versionOf: versionIs("")}, "refused.runtime-definition"},
		"an unknown version": {quarto, hugos{installed: "/usr/bin/hugo", versionOf: versionIs(""), pins: fakePins{err: nixpin.ErrNotFound}}, "refused.package-not-found"},
	} {
		_, err := case_.h.Runtime(t.Context(), uuid.New(), case_.site)
		var refused *httpapi.RefusedError
		if !errors.As(err, &refused) || refused.Key != case_.key {
			t.Errorf("%s: Runtime() error = %v, want %s", name, err, case_.key)
		}
	}
}
