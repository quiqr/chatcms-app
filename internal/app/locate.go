package app

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/model"
	"github.com/quiqr/chatcms-app/internal/runner"
	"github.com/quiqr/chatcms-app/internal/runtime"
	"github.com/quiqr/chatcms-app/internal/workspace"
)

const listWithin = 30 * time.Second

type pageLocator struct {
	dir       func(ctx context.Context, key workspace.Key) (string, error)
	runtimeOf func(ctx context.Context, project uuid.UUID, site fs.FS) (runtime.Resolved, error)
	runner    runner.Runner
	memory    int64
}

func (l pageLocator) Locate(ctx context.Context, key workspace.Key, asked string) (string, error) {
	dir, err := l.dir(ctx, key)
	if err != nil {
		return "", err
	}
	site := os.DirFS(dir)
	resolved, err := l.runtimeOf(ctx, key.Project, site)
	if err != nil || filepath.Base(resolved.Preview.Program) != "hugo" {
		return "", err
	}
	result, err := l.runner.Run(ctx, runner.Task{
		Inputs:      site,
		Command:     resolved.Preview.Program,
		Arguments:   []string{"list", "all", "--source", "."},
		Tools:       append([]string{}, resolved.Path...),
		Timeout:     listWithin,
		MemoryBytes: l.memory,
	})
	if err != nil {
		return "", err
	}
	defer result.Release()
	if !result.Succeeded() {
		return "", fmt.Errorf("hugo list all ended with %d: %s", result.ExitCode, lastLines(result.Log, 5))
	}
	file := FileAt(result.Log, asked)
	if file == "" {
		return "", nil
	}
	found, err := model.Load(site)
	if err != nil {
		return "", err
	}
	return found.KeyOf(file), nil
}

func lastLines(text string, most int) string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	if len(lines) > most {
		lines = lines[len(lines)-most:]
	}
	return strings.Join(lines, " / ")
}

func FileAt(listing, asked string) string {
	start := strings.Index(listing, "path,")
	if start < 0 {
		return ""
	}
	reader := csv.NewReader(strings.NewReader(listing[start:]))
	reader.FieldsPerRecord = -1
	header, err := reader.Read()
	if err != nil {
		return ""
	}
	pathAt, linkAt := -1, -1
	for i, name := range header {
		switch name {
		case "path":
			pathAt = i
		case "permalink":
			linkAt = i
		}
	}
	if pathAt < 0 || linkAt < 0 {
		return ""
	}
	asked = "/" + strings.Trim(path.Clean("/"+asked), "/")
	best, bestLength := "", -1
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(row) <= pathAt || len(row) <= linkAt {
			continue
		}
		link, err := url.Parse(row[linkAt])
		if err != nil {
			continue
		}
		served := "/" + strings.Trim(path.Clean("/"+link.Path), "/")
		switch {
		case served == asked:
			return row[pathAt]
		case asked != "/" && strings.HasSuffix(served, asked) && (bestLength < 0 || len(served) < bestLength):
			best, bestLength = row[pathAt], len(served)
		}
	}
	return best
}
