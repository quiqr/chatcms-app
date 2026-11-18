package app

import (
	"context"
	"errors"
	"testing"

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
