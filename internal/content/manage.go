package content

import (
	"context"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"strings"
	"unicode"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/frontmatter"
	"github.com/quiqr/chatcms-app/internal/gitsync"
	"github.com/quiqr/chatcms-app/internal/httpapi"
	"github.com/quiqr/chatcms-app/internal/model"
	"github.com/quiqr/chatcms-app/internal/policy"
)

// Slug turns what somebody typed into a file name, the way the old Quiqr
// does: small letters, digits and hyphens, with accents folded away.
func Slug(said string) string {
	lowered := strings.ToLower(strings.TrimSpace(said))

	out := strings.Builder{}
	previousHyphen := false
	for _, letter := range lowered {
		switch {
		case unicode.IsLetter(letter) && letter < unicode.MaxASCII,
			unicode.IsDigit(letter) && letter < unicode.MaxASCII:
			out.WriteRune(letter)
			previousHyphen = false
		case folded[letter] != "":
			out.WriteString(folded[letter])
			previousHyphen = false
		case !previousHyphen && out.Len() > 0:
			out.WriteRune('-')
			previousHyphen = true
		}
	}
	return strings.Trim(out.String(), "-")
}

// folded is the handful of letters a Dutch or German site actually uses. A
// letter nothing here knows becomes a hyphen, which is what the old Quiqr
// does with it too.
var folded = map[rune]string{
	'á': "a", 'à': "a", 'ä': "a", 'â': "a", 'å': "a",
	'é': "e", 'è': "e", 'ë': "e", 'ê': "e",
	'í': "i", 'ì': "i", 'ï': "i", 'î': "i",
	'ó': "o", 'ò': "o", 'ö': "o", 'ô': "o",
	'ú': "u", 'ù': "u", 'ü': "u", 'û': "u",
	'ñ': "n", 'ç': "c", 'ß': "ss",
}

var usableSlug = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// NewPage is what somebody asked to make.
type NewPage struct {
	// Section is the collection's key. Only a collection can grow.
	Section string
	// Title is what they typed. The file name comes from it.
	Title  string
	Folder string
	Slug   string
}

// Create makes a page in a collection, with the title filled in and nothing
// else, so that opening it is the next thing that happens.
func (s *Service) Create(ctx context.Context, actor policy.Actor, projectID uuid.UUID, asked NewPage) (Editable, error) {
	if err := s.allowed(ctx, actor, policy.ActionEditContent, projectID); err != nil {
		return Editable{}, err
	}

	site, done, err := s.sources.Site(ctx, actor.PersonID, projectID)
	if err != nil {
		return Editable{}, err
	}
	defer done()

	collection, err := collectionOf(site, asked.Section)
	if err != nil {
		return Editable{}, err
	}

	slug, folder, err := addressOf(site.FS(), collection, asked)
	if err != nil {
		return Editable{}, err
	}

	file := newItemFile(collection, folder, slug)
	if taken(site.FS(), collection, folder, slug) {
		return Editable{}, &httpapi.RefusedError{
			Key:    "refused.page-name-taken",
			Reason: fmt.Sprintf("there is already a page called %q here", asked.Title),
			Detail: map[string]any{"title": asked.Title},
		}
	}

	first, err := firstContent(asked.Title, file)
	if err != nil {
		return Editable{}, err
	}

	who, err := s.who(ctx, actor)
	if err != nil {
		return Editable{}, err
	}

	if _, err := site.Save(ctx, gitsync.Save{
		Message: "Add " + asked.Title,
		Author:  who,
		Writes:  map[string][]byte{file: first},
	}); err != nil {
		return Editable{}, err
	}

	s.syncSoon(ctx, actor, projectID)
	return s.after(ctx, site, collection.Key+"/"+path.Join(folder, slug))
}

var usableFolder = regexp.MustCompile(`^[A-Za-z0-9_-]+(?:/[A-Za-z0-9_-]+)*$`)

func addressOf(fsys fs.FS, collection model.Collection, asked NewPage) (string, string, error) {
	said := asked.Slug
	if strings.TrimSpace(said) == "" {
		said = asked.Title
	}
	slug := Slug(said)
	if !usableSlug.MatchString(slug) {
		return "", "", &httpapi.RefusedError{
			Key:    "refused.page-needs-name-with-some",
			Reason: "a page needs a name with some letters or numbers in it",
		}
	}

	folder := strings.Trim(strings.TrimSpace(asked.Folder), "/")
	if folder == "" {
		return slug, "", nil
	}
	if !usableFolder.MatchString(folder) {
		return "", "", httpapi.ErrMissing
	}
	if info, err := fs.Stat(fsys, path.Join(collection.Root(), folder)); err != nil || !info.IsDir() {
		return "", "", httpapi.ErrMissing
	}
	return slug, folder, nil
}

func firstContent(title, file string) ([]byte, error) {
	fields := frontmatter.NewMap()
	fields.Set("title", title)
	header, err := frontmatter.EmitYAML(fields)
	if err != nil {
		return nil, err
	}
	if isContentFile(file) {
		return []byte("---\n" + string(header) + "---\n\n"), nil
	}
	return header, nil
}

func isContentFile(file string) bool {
	switch strings.ToLower(strings.TrimPrefix(path.Ext(file), ".")) {
	case "md", "html", "markdown", "qmd":
		return true
	}
	return false
}

func newItemFile(collection model.Collection, folder, slug string) string {
	if collection.UnderContent() {
		return path.Join(collection.Root(), folder, slug, "index."+extensionOf(collection))
	}
	return path.Join(collection.Root(), folder, slug+"."+extensionOf(collection))
}

func taken(fsys fs.FS, collection model.Collection, folder, slug string) bool {
	for _, name := range []string{
		path.Join(collection.Root(), folder, slug+"."+extensionOf(collection)),
		path.Join(collection.Root(), folder, slug),
	} {
		if _, err := fs.Stat(fsys, name); err == nil {
			return true
		}
	}
	return false
}

func extensionOf(collection model.Collection) string {
	if collection.Extension != "" {
		return collection.Extension
	}
	return "md"
}

func collectionOf(site Site, key string) (model.Collection, error) {
	m, err := site.Model()
	if err != nil {
		return model.Collection{}, ErrNoModel
	}
	for _, collection := range m.Collections {
		if collection.Key == key && !collection.Hidden {
			return collection, nil
		}
	}
	return model.Collection{}, httpapi.ErrMissing
}

// Rename gives a page a new name, which means a new file. The old Quiqr names
// a file after its title, so renaming is what keeps the two in step.
func (s *Service) Rename(ctx context.Context, actor policy.Actor, projectID uuid.UUID, key, title string) (Editable, error) {
	return s.moveOrCopy(ctx, actor, projectID, key, title, false)
}

// Duplicate makes a copy of a page under a new name.
func (s *Service) Duplicate(ctx context.Context, actor policy.Actor, projectID uuid.UUID, key, title string) (Editable, error) {
	return s.moveOrCopy(ctx, actor, projectID, key, title, true)
}

func (s *Service) moveOrCopy(
	ctx context.Context,
	actor policy.Actor,
	projectID uuid.UUID,
	key, title string,
	keepTheOldOne bool,
) (Editable, error) {
	if err := s.allowed(ctx, actor, policy.ActionEditContent, projectID); err != nil {
		return Editable{}, err
	}

	site, done, err := s.sources.Site(ctx, actor.PersonID, projectID)
	if err != nil {
		return Editable{}, err
	}
	defer done()

	section, _, nested := strings.Cut(key, "/")
	if !nested {
		return Editable{}, &httpapi.RefusedError{
			Key:    "refused.page-model-names-cannot-renamed",
			Reason: "a page the model names cannot be renamed here; its name is in the model",
		}
	}

	collection, err := collectionOf(site, section)
	if err != nil {
		return Editable{}, err
	}

	page, form, err := load(ctx, site, key)
	if err != nil {
		return Editable{}, err
	}

	slug := Slug(title)
	if !usableSlug.MatchString(slug) {
		return Editable{}, &httpapi.RefusedError{
			Key:    "refused.page-needs-name-with-some",
			Reason: "a page needs a name with some letters or numbers in it",
		}
	}

	within := path.Dir(strings.TrimPrefix(key, collection.Key+"/"))
	if within == "." {
		within = ""
	}
	parent := path.Dir(page.File)
	to := path.Join(parent, slug+path.Ext(page.File))
	if page.Bundle {
		parent = path.Dir(parent)
		to = path.Join(parent, slug, path.Base(page.File))
	}
	if to == page.File {
		return s.after(ctx, site, key)
	}
	if nameTaken(site.FS(), parent, slug, path.Ext(page.File)) {
		return Editable{}, &httpapi.RefusedError{
			Key:    "refused.page-name-taken",
			Reason: fmt.Sprintf("there is already a page called %q here", title),
			Detail: map[string]any{"title": title},
		}
	}

	// The title inside the page follows its name, which is what the old
	// Quiqr does and what stops a renamed page still calling itself the
	// old thing.
	if _, has := form.Get("title"); has {
		if err := form.Set("title", title); err != nil {
			return Editable{}, &httpapi.RefusedError{Reason: err.Error()}
		}
	}
	written, err := form.Bytes()
	if err != nil {
		return Editable{}, &httpapi.RefusedError{Reason: unwritable(err)}
	}

	who, err := s.who(ctx, actor)
	if err != nil {
		return Editable{}, err
	}

	save := gitsync.Save{
		Author: who,
		Writes: map[string][]byte{to: written},
	}
	moved := []string{page.File}
	if page.Bundle {
		alongside, err := bundleFiles(site.FS(), page.File)
		if err != nil {
			return Editable{}, err
		}
		for _, file := range alongside {
			data, err := fs.ReadFile(site.FS(), file)
			if err != nil {
				return Editable{}, err
			}
			save.Writes[path.Join(path.Dir(to), strings.TrimPrefix(file, path.Dir(page.File)+"/"))] = data
			moved = append(moved, file)
		}
	}
	if keepTheOldOne {
		save.Message = "Copy " + page.Title + " to " + title
	} else {
		save.Message = "Rename " + page.Title + " to " + title
		save.Deletes = moved
	}

	if _, err := site.Save(ctx, save); err != nil {
		return Editable{}, err
	}

	s.syncSoon(ctx, actor, projectID)
	return s.after(ctx, site, collection.Key+"/"+path.Join(within, slug))
}

func nameTaken(fsys fs.FS, parent, slug, extension string) bool {
	for _, name := range []string{path.Join(parent, slug+extension), path.Join(parent, slug)} {
		if _, err := fs.Stat(fsys, name); err == nil {
			return true
		}
	}
	return false
}

func bundleFiles(fsys fs.FS, index string) ([]string, error) {
	var files []string
	err := fs.WalkDir(fsys, path.Dir(index), func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && name != index {
			files = append(files, name)
		}
		return nil
	})
	return files, err
}

// Delete removes a page. Only an item of a collection can go: a page the model
// names is part of the site's shape.
func (s *Service) Delete(ctx context.Context, actor policy.Actor, projectID uuid.UUID, key string) error {
	if err := s.allowed(ctx, actor, policy.ActionEditContent, projectID); err != nil {
		return err
	}

	site, done, err := s.sources.Site(ctx, actor.PersonID, projectID)
	if err != nil {
		return err
	}
	defer done()

	section, _, nested := strings.Cut(key, "/")
	if !nested {
		return &httpapi.RefusedError{
			Key:    "refused.page-model-names-cannot-deleted",
			Reason: "a page the model names cannot be deleted here; it is part of how the site is built",
		}
	}
	if _, err := collectionOf(site, section); err != nil {
		return err
	}

	page, _, err := load(ctx, site, key)
	if err != nil {
		return err
	}

	who, err := s.who(ctx, actor)
	if err != nil {
		return err
	}

	removed := []string{page.File}
	if page.Bundle {
		alongside, err := bundleFiles(site.FS(), page.File)
		if err != nil {
			return err
		}
		removed = append(removed, alongside...)
	}
	if _, err := site.Save(ctx, gitsync.Save{
		Message: "Remove " + page.Title,
		Author:  who,
		Deletes: removed,
	}); err != nil {
		return err
	}

	s.syncSoon(ctx, actor, projectID)
	return nil
}

// List is the content browser: every section with its items, and a mark on
// anything not yet published.
func (s *Service) List(ctx context.Context, actor policy.Actor, projectID uuid.UUID) ([]Section, error) {
	resource, err := s.reachable(ctx, actor, policy.ActionReadContent, projectID)
	if err != nil {
		return nil, err
	}

	site, done, err := s.sources.Site(ctx, actor.PersonID, projectID)
	if err != nil {
		return nil, err
	}
	defer done()

	if _, err := site.Model(); err != nil {
		return nil, noModel(err)
	}

	changes, err := site.Changes(ctx)
	if err != nil {
		return nil, err
	}
	return Browse(ctx, site, waitingFrom(changes), policy.Can(actor, policy.ActionSeeTechnicalDetail, resource)), nil
}

func noModel(err error) error {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), model.ErrNoModel.Error()) {
		return ErrNoModel
	}
	return &httpapi.RefusedError{
		Key:    "refused.model-cannot-be-read",
		Reason: "this site's model cannot be read: " + err.Error(),
		Detail: map[string]any{"why": err.Error()},
	}
}

type Address struct {
	Slug    string
	Folder  string
	Address string
	Free    bool
	File    string
}

func (s *Service) Address(ctx context.Context, actor policy.Actor, projectID uuid.UUID, asked NewPage) (Address, error) {
	resource, err := s.reachable(ctx, actor, policy.ActionEditContent, projectID)
	if err != nil {
		return Address{}, err
	}

	site, done, err := s.sources.Site(ctx, actor.PersonID, projectID)
	if err != nil {
		return Address{}, err
	}
	defer done()

	collection, err := collectionOf(site, asked.Section)
	if err != nil {
		return Address{}, err
	}
	slug, folder, err := addressOf(site.FS(), collection, asked)
	if err != nil {
		return Address{}, err
	}

	answer := Address{
		Slug:   slug,
		Folder: folder,
		Free:   !taken(site.FS(), collection, folder, slug),
	}
	if RuntimeOf(site.FS()) == "hugo" {
		answer.Address = webAddress(collection, folder, slug)
	}
	if policy.Can(actor, policy.ActionSeeTechnicalDetail, resource) {
		answer.File = newItemFile(collection, folder, slug)
	}
	return answer, nil
}

func webAddress(collection model.Collection, folder, slug string) string {
	base := strings.TrimSpace(collection.PreviewURLBase)
	if base == "" {
		base = "/" + strings.TrimPrefix(strings.TrimPrefix(collection.Root(), "content"), "/")
	}
	return path.Join("/", base, folder, slug) + "/"
}
