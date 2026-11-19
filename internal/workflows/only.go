package workflows

import (
	"io/fs"
	"path"
)

type chosen struct {
	fs.FS
	files map[string]bool
	dirs  map[string]bool
}

func only(whole fs.FS, wanted func(name string) bool) (fs.FS, error) {
	picked := &chosen{FS: whole, files: map[string]bool{}, dirs: map[string]bool{".": true}}
	err := fs.WalkDir(whole, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !wanted(name) {
			return nil
		}
		picked.files[name] = true
		for dir := path.Dir(name); dir != "."; dir = path.Dir(dir) {
			picked.dirs[dir] = true
		}
		return nil
	})
	return picked, err
}

func (c *chosen) Open(name string) (fs.File, error) {
	if !c.files[name] && !c.dirs[name] {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return c.FS.Open(name)
}

func (c *chosen) Stat(name string) (fs.FileInfo, error) {
	if !c.files[name] && !c.dirs[name] {
		return nil, &fs.PathError{Op: "stat", Path: name, Err: fs.ErrNotExist}
	}
	return fs.Stat(c.FS, name)
}

func (c *chosen) ReadDir(name string) ([]fs.DirEntry, error) {
	if !c.dirs[name] {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrNotExist}
	}
	entries, err := fs.ReadDir(c.FS, name)
	if err != nil {
		return nil, err
	}
	kept := entries[:0]
	for _, entry := range entries {
		full := path.Join(name, entry.Name())
		if c.files[full] || c.dirs[full] {
			kept = append(kept, entry)
		}
	}
	return kept, nil
}
