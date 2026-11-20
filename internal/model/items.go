package model

import (
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/quiqr/chatcms-app/internal/frontmatter"
)

type Item struct {
	Key       string
	File      string
	Bundle    bool
	Folder    string
	Title     string
	SortValue string
	Sorted    bool
	Date      string
	Author    string
	Summary   string
	Draft     bool
	Thumbnail string
}

var contentExtensions = []string{"md", "html", "markdown", "qmd"}

var dataExtensions = []string{"json", "toml", "yaml", "yml"}

func (c Collection) Root() string {
	return path.Clean(strings.TrimSuffix(strings.TrimSpace(c.Folder), "/"))
}

func (c Collection) UnderContent() bool {
	root := c.Root()
	return root == "content" || strings.HasPrefix(root, "content/")
}

func (c Collection) contentStyle() bool {
	if c.UnderContent() {
		return true
	}
	for _, extension := range contentExtensions {
		if strings.TrimPrefix(c.Extension, ".") == extension {
			return true
		}
	}
	return false
}

func Items(fsys fs.FS, c Collection) ([]Item, []Warning) {
	root := c.Root()
	if fsys == nil || root == "" || root == "." || root == "/" {
		return nil, nil
	}
	if info, err := fs.Stat(fsys, root); err != nil || !info.IsDir() {
		return nil, nil
	}

	extensions := dataExtensions
	recursive := true
	if c.contentStyle() {
		extensions = contentExtensions
		recursive = c.LooksBelow()
	}

	byKey := map[string]Item{}
	var order []string
	var warnings []Warning

	_ = fs.WalkDir(fsys, root, func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if name == root {
			return nil
		}
		if strings.HasPrefix(entry.Name(), ".") {
			if entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			if !recursive {
				return fs.SkipDir
			}
			return nil
		}

		extension := strings.TrimPrefix(path.Ext(entry.Name()), ".")
		if !hasExtension(extensions, extension) {
			return nil
		}
		stem := strings.TrimSuffix(entry.Name(), path.Ext(entry.Name()))
		if c.HideIndex && stem == "_index" {
			return nil
		}

		rel := strings.TrimPrefix(name, root+"/")
		item := Item{File: name, Key: strings.TrimSuffix(rel, path.Ext(rel))}
		if c.contentStyle() && stem == "index" && strings.Contains(rel, "/") {
			item.Bundle = true
			item.Key = path.Dir(rel)
		}
		if dir := path.Dir(item.Key); dir != "." {
			item.Folder = dir
		}

		if existing, taken := byKey[item.Key]; taken {
			kept, dropped := existing, item
			if existing.Bundle && !item.Bundle {
				kept, dropped = item, existing
			}
			byKey[item.Key] = kept
			warnings = append(warnings, Warning{
				File:    dropped.File,
				Path:    "collections." + c.Key,
				Message: fmt.Sprintf("%s and %s are both called %q; only %s is listed", kept.File, dropped.File, item.Key, kept.File),
			})
			return nil
		}
		byKey[item.Key] = item
		order = append(order, item.Key)
		return nil
	})

	items := make([]Item, 0, len(order))
	for _, key := range order {
		items = append(items, describeItem(fsys, c, byKey[key]))
	}

	sort.SliceStable(items, func(i, j int) bool {
		a, b := items[i], items[j]
		if a.Sorted != b.Sorted {
			return a.Sorted
		}
		if a.Sorted && a.SortValue != b.SortValue {
			return a.SortValue < b.SortValue
		}
		return a.Key < b.Key
	})
	return items, warnings
}

func describeItem(fsys fs.FS, c Collection, item Item) Item {
	item.Title = Readable(path.Base(item.Key))

	data, err := fs.ReadFile(fsys, item.File)
	if err != nil {
		return item
	}
	document, err := frontmatter.Read(item.File, data)
	if err != nil || document.Fields == nil {
		return item
	}

	if title, ok := document.Fields.Get("title"); ok {
		if text, ok := title.(string); ok && strings.TrimSpace(text) != "" {
			item.Title = strings.TrimSpace(text)
		}
	}
	if c.SortKey != "" {
		if value, ok := document.Fields.Get(c.SortKey); ok && value != nil {
			item.SortValue = sortable(value)
			item.Sorted = true
		}
	}
	if value, ok := document.Fields.Get("date"); ok && value != nil {
		item.Date = sortable(value)
	}
	item.Author = firstText(document.Fields, "author")
	item.Summary = firstText(document.Fields, "summary", "description")
	if value, ok := document.Fields.Get("draft"); ok {
		item.Draft = value == true
	}
	if field := c.thumbnailField(); field != "" {
		if text := firstText(document.Fields, field); text != "" {
			item.Thumbnail = path.Base(text)
		}
	}
	return item
}

func (c Collection) thumbnailField() string {
	for _, field := range c.Fields {
		if field.Type == "image-select" {
			return field.Key
		}
	}
	return ""
}

func (c Collection) HasThumbnails() bool {
	return c.thumbnailField() != ""
}

func (c Collection) HasDraft() bool {
	for _, field := range c.Fields {
		if field.Key == "draft" {
			return true
		}
	}
	return false
}

func firstText(fields *frontmatter.Map, keys ...string) string {
	for _, key := range keys {
		value, ok := fields.Get(key)
		if !ok || value == nil {
			continue
		}
		switch typed := value.(type) {
		case string:
			if strings.TrimSpace(typed) != "" {
				return strings.TrimSpace(typed)
			}
		case []any:
			if len(typed) > 0 {
				if text, ok := typed[0].(string); ok {
					return text
				}
			}
		}
	}
	return ""
}

func sortable(value any) string {
	switch typed := value.(type) {
	case time.Time:
		return typed.UTC().Format(time.RFC3339Nano)
	case string:
		return typed
	default:
		return fmt.Sprint(typed)
	}
}

func hasExtension(extensions []string, extension string) bool {
	for _, candidate := range extensions {
		if strings.EqualFold(candidate, extension) {
			return true
		}
	}
	return false
}

func Readable(name string) string {
	spaced := strings.NewReplacer("-", " ", "_", " ").Replace(name)
	spaced = strings.TrimSpace(spaced)
	if spaced == "" {
		return name
	}
	return strings.ToUpper(spaced[:1]) + spaced[1:]
}

func FindItem(fsys fs.FS, c Collection, key string) (Item, bool) {
	items, _ := Items(fsys, c)
	for _, item := range items {
		if item.Key == key {
			return item, true
		}
	}
	return Item{}, false
}

func (m *Model) KeyOf(file string) string {
	if m == nil {
		return ""
	}
	file = path.Clean(strings.TrimPrefix(file, "/"))
	for _, single := range m.Singles {
		if path.Clean(strings.TrimPrefix(single.File, "/")) == file {
			return single.Key
		}
	}
	best, key := "", ""
	for _, c := range m.Collections {
		root := c.Root()
		if root == "" || root == "." || root == "/" || len(root) <= len(best) {
			continue
		}
		rel, found := strings.CutPrefix(file, root+"/")
		if !found {
			continue
		}
		name := path.Base(rel)
		stem := strings.TrimSuffix(name, path.Ext(name))
		item := strings.TrimSuffix(rel, path.Ext(rel))
		if c.contentStyle() && stem == "index" && strings.Contains(rel, "/") {
			item = path.Dir(rel)
		}
		best, key = root, c.Key+"/"+item
	}
	return key
}
