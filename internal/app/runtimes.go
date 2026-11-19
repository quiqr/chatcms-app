package app

import (
	"errors"
	"io/fs"
	"os"
	"path"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/gitsync"
	"github.com/quiqr/chatcms-app/internal/runtime"
)

type siteRuntimes struct {
	sites *gitsync.Sites
}

func runtimeOfSite(site fs.FS) string {
	data, err := fs.ReadFile(site, runtime.File)
	if errors.Is(err, fs.ErrNotExist) {
		if _, err := fs.Stat(site, "."); err != nil {
			return ""
		}
		entries, _ := fs.ReadDir(site, ".")
		if len(entries) == 0 {
			return ""
		}
		return "hugo"
	}
	if err != nil {
		return ""
	}
	definition, err := runtime.Parse(data)
	if err != nil {
		return ""
	}
	return path.Base(definition.Preview.Command[0])
}

func (r siteRuntimes) RuntimeOf(projectID uuid.UUID) string {
	dir, err := r.sites.Dir(projectID.String())
	if err != nil {
		return ""
	}
	return runtimeOfSite(os.DirFS(dir))
}
