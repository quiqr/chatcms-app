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
