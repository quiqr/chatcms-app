package content_test

import (
	"context"
	"errors"
	"testing"
	"testing/fstest"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/content"
	"github.com/quiqr/chatcms-app/internal/httpapi"
	"github.com/quiqr/chatcms-app/internal/model"
)

func aDataSite() *site {
	return &site{
		files: fstest.MapFS{
			"content/post/hello/index.md": {Data: []byte("---\ntitle: Hello\n---\n")},
			"data/team/anna.yaml":         {Data: []byte("title: Anna\n")},
		},
		model: &model.Model{Collections: []model.Collection{
			{Key: "posts", Folder: "content/post/", Extension: "md"},
			{Key: "team", Folder: "data/team/", Extension: "yaml"},
		}},
	}
}

func TestAPageOutsideContentIsAFile(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aDataSite()}

	page, err := service(t, src, &syncer{}).Create(context.Background(), anEditor(project), project,
		content.NewPage{Section: "team", Title: "Bram"})
	if err != nil {
		t.Fatal(err)
	}
	written, wrote := src.site.saved[0].Writes["data/team/bram.yaml"]
	if !wrote || string(written) != "title: Bram\n" || page.Key != "team/bram" {
		t.Errorf("wrote %q (%v), key %q", written, src.site.saved[0].Writes, page.Key)
	}
}

func TestAPageCannotBeMadeOverABundle(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aDataSite()}

	_, err := service(t, src, &syncer{}).Create(context.Background(), anEditor(project), project,
		content.NewPage{Section: "posts", Title: "Hello"})
	if !errors.Is(err, httpapi.ErrRefused) || len(src.site.saved) != 0 {
		t.Errorf("Create() error = %v, saved %d", err, len(src.site.saved))
	}
}

func TestAPageCannotBeMadeOverAFlatFile(t *testing.T) {
	project := uuid.New()
	working := aDataSite()
	working.files["content/post/flat.md"] = &fstest.MapFile{Data: []byte("---\ntitle: Flat\n---\n")}
	src := &sources{site: working}

	_, err := service(t, src, &syncer{}).Create(context.Background(), anEditor(project), project,
		content.NewPage{Section: "posts", Title: "Flat"})
	if !errors.Is(err, httpapi.ErrRefused) || len(src.site.saved) != 0 {
		t.Errorf("Create() error = %v, saved %d", err, len(src.site.saved))
	}
}

func TestANewPageHoldsWhatQuiqrDesktopWrites(t *testing.T) {
	for title, want := range map[string]string{
		"Spring prices": "---\ntitle: Spring prices\n---\n\n",
		"Spring: sale":  "---\ntitle: 'Spring: sale'\n---\n\n",
		"2024":          "---\ntitle: '2024'\n---\n\n",
		"Café":          "---\ntitle: Café\n---\n\n",
	} {
		project := uuid.New()
		src := &sources{site: aDataSite()}
		page, err := service(t, src, &syncer{}).Create(context.Background(), anEditor(project), project,
			content.NewPage{Section: "posts", Title: title})
		if err != nil {
			t.Fatalf("Create(%q) error = %v", title, err)
		}
		file := "content/post/" + page.Key[len("posts/"):] + "/index.md"
		if got := string(src.site.saved[0].Writes[file]); got != want {
			t.Errorf("Create(%q) wrote %q to %s, want %q", title, got, file, want)
		}
	}
}

func TestASiteConfigurationAndDataAreEditedWhole(t *testing.T) {
	project := uuid.New()
	working := &site{
		files: fstest.MapFS{
			"config.toml":     {Data: []byte("title = \"UILite\"\n\n[params]\ncolor = \"blue\"\n")},
			"data/hours.yaml": {Data: []byte("note: Closed on Sundays\n")},
		},
		model: &model.Model{Singles: []model.Single{
			{Key: "mainConfig", File: "config.toml", Fields: []model.Field{{Key: "title", Type: "string"}}},
			{Key: "hours", File: "data/hours.yaml", Fields: []model.Field{{Key: "note", Type: "string"}}},
		}},
	}
	src := &sources{site: working}
	editing := service(t, src, &syncer{})

	for key, want := range map[string]string{"mainConfig": "UILite", "hours": "Closed on Sundays"} {
		page, err := editing.Read(context.Background(), anEditor(project), project, key)
		if err != nil || len(page.Controls) == 0 || page.Controls[0].Value != want {
			t.Errorf("Read(%s) = %+v, %v, want %q", key, page.Controls, err, want)
		}
	}

	if _, err := editing.Save(context.Background(), anEditor(project), project, "mainConfig", content.Edit{Values: map[string]any{"title": "My portfolio"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := editing.Save(context.Background(), anEditor(project), project, "hours", content.Edit{Values: map[string]any{"note": "Open on Sundays"}}); err != nil {
		t.Fatal(err)
	}
	if got := string(src.site.saved[0].Writes["config.toml"]); got != "title = \"My portfolio\"\n\n[params]\ncolor = \"blue\"" {
		t.Errorf("config.toml = %q", got)
	}
	if got := string(src.site.saved[1].Writes["data/hours.yaml"]); got != "note: Open on Sundays\n" {
		t.Errorf("data/hours.yaml = %q", got)
	}
}
