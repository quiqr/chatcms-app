package content

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"path"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/gitsync"
	"github.com/quiqr/chatcms-app/internal/httpapi"
	"github.com/quiqr/chatcms-app/internal/policy"
)

const (
	mostFiles    = 5000
	largestFile  = 1 << 20
	fileTooLarge = "too-large"
	fileBinary   = "binary"
	fileLink     = "link"
)

var leftOut = map[string]bool{".git": true, "node_modules": true, "public": true, "resources/_gen": true}

type FileEntry struct {
	Path string
	Dir  bool
	Link bool
	Size int64
}

type FileText struct {
	Path     string
	Text     string
	Sum      string
	Unusable string
}

var refusedPath = &httpapi.RefusedError{Key: "refused.file-path", Reason: "this path is not a file of the project"}

func refusePath() error {
	copied := *refusedPath
	return &copied
}

func checkPath(fsys fs.FS, name string) error {
	if name == "" || !fs.ValidPath(name) || name == "." {
		return refusePath()
	}
	parts := strings.Split(name, "/")
	if parts[0] == ".git" {
		return refusePath()
	}
	for i := range parts {
		info, err := fs.Lstat(fsys, strings.Join(parts[:i+1], "/"))
		if errors.Is(err, fs.ErrNotExist) {
			return httpapi.ErrMissing
		}
		if err != nil {
			return err
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			return refusePath()
		}
	}
	return nil
}

func sumOf(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func (s *Service) Files(ctx context.Context, actor policy.Actor, projectID uuid.UUID) ([]FileEntry, error) {
	if err := s.allowed(ctx, actor, policy.ActionSeeTechnicalDetail, projectID); err != nil {
		return nil, err
	}
	site, done, err := s.sources.Site(ctx, actor.PersonID, projectID)
	if err != nil {
		return nil, err
	}
	defer done()

	entries := []FileEntry{}
	err = fs.WalkDir(site.FS(), ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || name == "." {
			return nil
		}
		if leftOut[name] {
			if entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if len(entries) >= mostFiles {
			return fs.SkipAll
		}
		item := FileEntry{Path: name, Dir: entry.IsDir(), Link: entry.Type()&fs.ModeSymlink != 0}
		if info, err := entry.Info(); err == nil && !item.Dir {
			item.Size = info.Size()
		}
		entries = append(entries, item)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	return entries, nil
}

func (s *Service) ReadFile(ctx context.Context, actor policy.Actor, projectID uuid.UUID, name string) (FileText, error) {
	if err := s.allowed(ctx, actor, policy.ActionSeeTechnicalDetail, projectID); err != nil {
		return FileText{}, err
	}
	site, done, err := s.sources.Site(ctx, actor.PersonID, projectID)
	if err != nil {
		return FileText{}, err
	}
	defer done()
	return readText(site.FS(), name)
}

func readText(fsys fs.FS, name string) (FileText, error) {
	if err := checkPath(fsys, name); err != nil {
		return FileText{}, err
	}
	info, err := fs.Stat(fsys, name)
	if err != nil {
		return FileText{}, httpapi.ErrMissing
	}
	if info.IsDir() {
		return FileText{}, refusePath()
	}
	if info.Size() > largestFile {
		return FileText{Path: name, Unusable: fileTooLarge}, nil
	}
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return FileText{}, err
	}
	if bytes.IndexByte(data, 0) >= 0 || !utf8.Valid(data) {
		return FileText{Path: name, Unusable: fileBinary, Sum: sumOf(data)}, nil
	}
	return FileText{Path: name, Text: string(data), Sum: sumOf(data)}, nil
}

func (s *Service) WriteFile(ctx context.Context, actor policy.Actor, projectID uuid.UUID, name, text, sum string) (FileText, error) {
	if err := s.allowed(ctx, actor, policy.ActionSeeTechnicalDetail, projectID); err != nil {
		return FileText{}, err
	}
	if err := s.allowed(ctx, actor, policy.ActionEditContent, projectID); err != nil {
		return FileText{}, err
	}
	site, done, err := s.sources.Site(ctx, actor.PersonID, projectID)
	if err != nil {
		return FileText{}, err
	}
	defer done()

	current, err := readText(site.FS(), name)
	if err != nil {
		return FileText{}, err
	}
	if current.Unusable != "" {
		return FileText{}, &httpapi.RefusedError{Key: "refused.file-not-text", Reason: "this file cannot be edited as text"}
	}
	if current.Sum != sum {
		return FileText{}, &httpapi.RefusedError{Key: "refused.file-changed", Reason: "this file changed since it was opened"}
	}

	who, err := s.who(ctx, actor)
	if err != nil {
		return FileText{}, err
	}
	if _, err := site.Save(ctx, gitsync.Save{
		Message: "Edit " + path.Base(name),
		Author:  who,
		Writes:  map[string][]byte{name: []byte(text)},
	}); err != nil && !errors.Is(err, gitsync.ErrNothingChanged) {
		return FileText{}, err
	}
	s.syncSoon(ctx, actor, projectID)
	return FileText{Path: name, Text: text, Sum: sumOf([]byte(text))}, nil
}
