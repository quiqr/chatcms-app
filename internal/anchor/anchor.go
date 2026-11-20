package anchor

import (
	"fmt"
	"html"
	"io/fs"
	"regexp"
	"sort"
	"strings"

	"github.com/quiqr/chatcms-app/internal/content"
	"github.com/quiqr/chatcms-app/internal/forms"
	"github.com/quiqr/chatcms-app/internal/frontmatter"
	"github.com/quiqr/chatcms-app/internal/model"
)

type Kind string

const (
	InAField   Kind = "field"
	SharedText Kind = "shared"
	TheDesign  Kind = "design"
)

type Where struct {
	Kind      Kind   `json:"kind"`
	Page      string `json:"page,omitempty"`
	PageTitle string `json:"pageTitle,omitempty"`
	Field     string `json:"field,omitempty"`
	Matches   int    `json:"matches,omitempty"`
}

type Asked struct {
	Text   string
	Before string
	After  string
	Page   string
}

const mostPages = 1000

var (
	tags        = regexp.MustCompile(`<[^>]*>`)
	shortcodes  = regexp.MustCompile(`\{\{[<%].*?[>%]\}\}`)
	links       = regexp.MustCompile(`!?\[([^\]]*)\]\([^)]*\)`)
	references  = regexp.MustCompile(`\[([^\]]*)\]\[[^\]]*\]`)
	definitions = regexp.MustCompile(`(?m)^\s*\[[^\]]+\]:\s*\S+.*$`)
	headings    = regexp.MustCompile(`(?m)^\s{0,3}(#{1,6}\s+|>\s?|[-*+]\s+|\d+[.)]\s+)`)
	marks       = regexp.MustCompile("[*_~`]+")
	spaces      = regexp.MustCompile(`\s+`)
	dataLine    = regexp.MustCompile(`^\s*(?:-\s+)?([\w-]+)\s*[:=]\s*(.+)$`)
)

func Plain(text string) string {
	text = shortcodes.ReplaceAllString(text, " ")
	text = definitions.ReplaceAllString(text, " ")
	text = links.ReplaceAllString(text, "$1")
	text = references.ReplaceAllString(text, "$1")
	text = tags.ReplaceAllString(text, " ")
	text = headings.ReplaceAllString(text, "")
	text = marks.ReplaceAllString(text, "")
	text = html.UnescapeString(text)
	return strings.ToLower(strings.TrimSpace(spaces.ReplaceAllString(text, " ")))
}

type candidate struct {
	page  string
	title string
	field string
	plain string
}

type page struct {
	key    string
	title  string
	file   string
	fields []model.Field
	shared bool
}

func pagesOf(site fs.FS, m *model.Model) []page {
	var pages []page
	for _, single := range m.Singles {
		if single.File == "" {
			continue
		}
		pages = append(pages, page{
			key:    single.Key,
			title:  titleOr(single.Title, single.Key),
			file:   single.File,
			fields: single.Fields,
			shared: !strings.HasPrefix(strings.TrimPrefix(single.File, "/"), "content/"),
		})
	}
	for _, collection := range m.Collections {
		items, _ := model.Items(site, collection)
		for _, item := range items {
			if len(pages) >= mostPages {
				return pages
			}
			pages = append(pages, page{
				key:    collection.Key + "/" + item.Key,
				title:  titleOr(item.Title, item.Key),
				file:   item.File,
				fields: collection.Fields,
			})
		}
	}
	return pages
}

func titleOr(title, key string) string {
	if title != "" {
		return title
	}
	return key
}

func candidatesOf(site fs.FS, p page) []candidate {
	data, err := fs.ReadFile(site, strings.TrimPrefix(p.file, "/"))
	if err != nil {
		return nil
	}
	var found []candidate
	add := func(field string, value string) {
		if plain := Plain(value); plain != "" {
			found = append(found, candidate{page: p.key, title: p.title, field: field, plain: plain})
		}
	}
	document, err := frontmatter.Read(p.file, data)
	if err != nil {
		return nil
	}
	if document.Format == frontmatter.FormatNone && isData(p.file) {
		for _, line := range strings.Split(string(data), "\n") {
			if held := dataLine.FindStringSubmatch(line); held != nil {
				add(held[1], strings.Trim(strings.TrimSpace(held[2]), `"'`))
			}
		}
		return found
	}
	form := forms.New(p.fields, document)
	for _, binding := range form.Bindings() {
		if binding.IsBody {
			add(binding.FieldPath, form.Body())
			continue
		}
		value, ok := form.Get(binding.FieldPath)
		if !ok {
			continue
		}
		walk(binding.FieldPath, content.Plain(value), add)
	}
	return found
}

func isData(file string) bool {
	for _, extension := range []string{".yaml", ".yml", ".json", ".toml"} {
		if strings.HasSuffix(strings.ToLower(file), extension) {
			return true
		}
	}
	return false
}

func walk(at string, value any, add func(string, string)) {
	switch v := value.(type) {
	case string:
		add(at, v)
	case []any:
		for i, one := range v {
			walk(fmt.Sprintf("%s[%d]", at, i), one, add)
		}
	case map[string]any:
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			walk(at+"."+key, v[key], add)
		}
	}
}

func best(found []candidate, text, before, after string) (candidate, int) {
	var hits []candidate
	for _, one := range found {
		if strings.Contains(one.plain, text) {
			hits = append(hits, one)
		}
	}
	if len(hits) == 0 {
		return candidate{}, 0
	}
	if len(hits) > 1 {
		for _, one := range hits {
			if (before != "" && strings.Contains(one.plain, before+" "+text)) || (before != "" && strings.Contains(one.plain, before+text)) ||
				(after != "" && strings.Contains(one.plain, text+" "+after)) || (after != "" && strings.Contains(one.plain, text+after)) {
				return one, len(hits)
			}
		}
	}
	return hits[0], len(hits)
}

func tail(text string, words int) string {
	parts := strings.Fields(text)
	if len(parts) > words {
		parts = parts[len(parts)-words:]
	}
	return strings.Join(parts, " ")
}

func head(text string, words int) string {
	parts := strings.Fields(text)
	if len(parts) > words {
		parts = parts[:words]
	}
	return strings.Join(parts, " ")
}

func Find(site fs.FS, m *model.Model, asked Asked) Where {
	text := Plain(asked.Text)
	if m == nil || len([]rune(text)) < 2 {
		return Where{Kind: TheDesign}
	}
	before := tail(Plain(asked.Before), 3)
	after := head(Plain(asked.After), 3)
	pages := pagesOf(site, m)

	look := func(keep func(page) bool) (candidate, int) {
		var found []candidate
		for _, p := range pages {
			if keep(p) {
				found = append(found, candidatesOf(site, p)...)
			}
		}
		return best(found, text, before, after)
	}

	if asked.Page != "" {
		if hit, n := look(func(p page) bool { return p.key == asked.Page }); n > 0 {
			return Where{Kind: InAField, Page: hit.page, PageTitle: hit.title, Field: hit.field, Matches: n}
		}
	}
	if hit, n := look(func(p page) bool { return p.shared }); n > 0 {
		return Where{Kind: SharedText, Page: hit.page, PageTitle: hit.title, Field: hit.field, Matches: n}
	}
	if hit, n := look(func(p page) bool { return !p.shared && p.key != asked.Page }); n > 0 {
		return Where{Kind: InAField, Page: hit.page, PageTitle: hit.title, Field: hit.field, Matches: n}
	}
	return Where{Kind: TheDesign}
}
