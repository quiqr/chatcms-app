package app

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/runner"
	"github.com/quiqr/chatcms-app/internal/runtime"
	"github.com/quiqr/chatcms-app/internal/workspace"
)

const listing = `WARN  some warning first
path,slug,title,date,expiryDate,publishDate,draft,permalink,kind,section
content/_index.md,,Korenaar,,,,false,https://korenaar.example/,home,
content/post/hello.md,,Hello,2024-02-01,,,false,https://korenaar.example/post/hello/,page,post
content/post/2024/spring/index.md,,"Spring, sale",,,,false,https://korenaar.example/post/2024/spring/,page,post
content/about.md,,About,,,,false,https://korenaar.example/sub/about/,page,
`

func TestFileAtMatchesTheAddressAPageIsServedAt(t *testing.T) {
	for asked, want := range map[string]string{
		"/":                  "content/_index.md",
		"/post/hello/":       "content/post/hello.md",
		"/post/hello":        "content/post/hello.md",
		"/post/2024/spring/": "content/post/2024/spring/index.md",
		"/about/":            "content/about.md",
		"/nowhere/":          "",
	} {
		if got := FileAt(listing, asked); got != want {
			t.Errorf("FileAt(%q) = %q, want %q", asked, got, want)
		}
	}
	if FileAt("Error: no site", "/") != "" || FileAt("path,title\nx,y\n", "/") != "" {
		t.Error("a listing without its columns matched")
	}
}

type listingRunner struct {
	log  string
	code int
	err  error
	task runner.Task
}

func (r *listingRunner) Run(_ context.Context, task runner.Task) (*runner.Result, error) {
	r.task = task
	if r.err != nil {
		return nil, r.err
	}
	return &runner.Result{ExitCode: r.code, Log: r.log}, nil
}

func aCopy(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"quiqr/model/base.yaml": "build: []\nserve: []\nhugover: ''\nmenu: []\ncollections:\n  - key: posts\n    title: Posts\n    folder: content/post/\n    extension: md\n    dataformat: yaml\n    itemtitle: Post\n    fields: []\nsingles:\n  - key: home\n    title: Home\n    file: content/_index.md\n    fields: []\n",
		"content/_index.md":     "---\ntitle: Korenaar\n---\n",
		"content/post/hello.md": "---\ntitle: Hello\n---\n",
	}
	for name, text := range files {
		full := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestLocateGivesThePageKey(t *testing.T) {
	dir := aCopy(t)
	run := &listingRunner{log: listing}
	locator := pageLocator{
		dir: func(context.Context, workspace.Key) (string, error) { return dir, nil },
		runtimeOf: func(context.Context, uuid.UUID, fs.FS) (runtime.Resolved, error) {
			return runtime.Resolved{Path: []string{"/store/hugo/bin"}, Preview: runtime.Command{Program: "/store/hugo/bin/hugo"}}, nil
		},
		runner: run,
	}
	key := workspace.Key{Person: uuid.New(), Project: uuid.New()}

	for asked, want := range map[string]string{"/post/hello/": "posts/hello", "/": "home", "/nowhere/": ""} {
		if got, err := locator.Locate(t.Context(), key, asked); err != nil || got != want {
			t.Errorf("Locate(%q) = %q, %v, want %q", asked, got, err, want)
		}
	}
	if run.task.Command != "/store/hugo/bin/hugo" || run.task.Network || len(run.task.Tools) != 1 || run.task.Arguments[0] != "list" {
		t.Errorf("task = %+v", run.task)
	}

	run.code = 1
	if got, err := locator.Locate(t.Context(), key, "/post/hello/"); got != "" || err == nil {
		t.Errorf("a failed listing = %q, %v", got, err)
	}
	run.err = errors.New("no sandbox")
	if _, err := locator.Locate(t.Context(), key, "/"); err == nil {
		t.Error("a runner error was swallowed")
	}
}

func TestLocateLeavesOtherRuntimesAlone(t *testing.T) {
	run := &listingRunner{log: listing}
	locator := pageLocator{
		dir: func(context.Context, workspace.Key) (string, error) { return t.TempDir(), nil },
		runtimeOf: func(context.Context, uuid.UUID, fs.FS) (runtime.Resolved, error) {
			return runtime.Resolved{Preview: runtime.Command{Program: "/store/quarto/bin/quarto"}}, nil
		},
		runner: run,
	}
	if got, err := locator.Locate(t.Context(), workspace.Key{}, "/"); got != "" || err != nil || run.task.Command != "" {
		t.Errorf("Locate() = %q, %v, ran %q", got, err, run.task.Command)
	}
}
