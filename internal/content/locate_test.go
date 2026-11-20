package content

import (
	"testing"
	"testing/fstest"

	"github.com/quiqr/chatcms-app/internal/model"
)

func locateSite() (*model.Model, fstest.MapFS) {
	m := &model.Model{
		Singles:     []model.Single{{Key: "home", Title: "Home", File: "content/_index.md"}},
		Collections: []model.Collection{{Key: "posts", Title: "Posts", Folder: "content/post/"}},
	}
	files := fstest.MapFS{
		"content/_index.md":                      {Data: []byte("---\ntitle: Home\n---\n")},
		"content/post/hello.md":                  {Data: []byte("---\ntitle: Hello\n---\n")},
		"content/post/2024/spring-sale/index.md": {Data: []byte("---\ntitle: Spring sale\n---\n")},
		"content/post/both.md":                   {Data: []byte("---\ntitle: Flat\n---\n")},
		"content/post/both/index.md":             {Data: []byte("---\ntitle: Bundle\n---\n")},
	}
	return m, files
}

func TestLocateFindsAFlatFile(t *testing.T) {
	m, files := locateSite()
	page, ok := Locate(m, files, "posts/hello")
	if !ok || page.File != "content/post/hello.md" || page.Bundle || page.Title != "Hello" {
		t.Errorf("Locate() = %+v, %v", page, ok)
	}
}

func TestLocateFindsABundleInASubfolder(t *testing.T) {
	m, files := locateSite()
	page, ok := Locate(m, files, "posts/2024/spring-sale")
	if !ok || page.File != "content/post/2024/spring-sale/index.md" || !page.Bundle {
		t.Errorf("Locate() = %+v, %v", page, ok)
	}
}

func TestLocatePrefersTheFlatFile(t *testing.T) {
	m, files := locateSite()
	page, ok := Locate(m, files, "posts/both")
	if !ok || page.File != "content/post/both.md" {
		t.Errorf("Locate() = %+v, %v", page, ok)
	}
}

func TestLocateFindsNothingOutsideTheListing(t *testing.T) {
	m, files := locateSite()
	for _, key := range []string{"posts/../_index", "posts/nothing", "nothing/hello", "posts/", ""} {
		if page, ok := Locate(m, files, key); ok {
			t.Errorf("Locate(%q) = %+v, want nothing", key, page)
		}
	}
	if page, ok := Locate(m, files, "home"); !ok || page.File != "content/_index.md" {
		t.Errorf("Locate(home) = %+v, %v", page, ok)
	}
}

func TestLocateSaysWhatThePreviewShows(t *testing.T) {
	m := &model.Model{
		Singles: []model.Single{
			{Key: "home", File: "content/_index.md"},
			{Key: "book", File: "index.qmd"},
			{Key: "services", File: "data/services.json"},
			{Key: "mainConfig", File: "config.toml"},
			{Key: "sidebar", File: "data/sidebar.yaml", PreviewURL: " / "},
			{Key: "secret", File: "content/secret.md", HidePreviewIcon: true},
		},
		Collections: []model.Collection{
			{Key: "posts", Folder: "content/post/"},
			{Key: "team", Folder: "data/team/"},
		},
	}
	files := fstest.MapFS{
		"content/post/hello.md": {Data: []byte("---\ntitle: Hello\n---\n")},
		"data/team/anna.yaml":   {Data: []byte("name: Anna\n")},
	}
	for key, want := range map[string]Preview{
		"home":        {Kind: PreviewOwnPage},
		"book":        {Kind: PreviewOwnPage},
		"services":    {Kind: PreviewNoPage},
		"mainConfig":  {Kind: PreviewNoPage},
		"sidebar":     {Kind: PreviewDeclared, Address: "/"},
		"secret":      {Kind: PreviewHidden},
		"posts/hello": {Kind: PreviewOwnPage},
		"team/anna":   {Kind: PreviewNoPage},
	} {
		page, ok := Locate(m, files, key)
		if !ok || page.Preview != want {
			t.Errorf("Locate(%s).Preview = %+v, %v, want %+v", key, page.Preview, ok, want)
		}
	}
}
