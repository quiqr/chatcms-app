package content_test

import (
	"bytes"
	"context"
	"io"
	"errors"
	"testing"
	"testing/fstest"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/content"
	"github.com/quiqr/chatcms-app/internal/httpapi"
)

func aSiteWithFiles() *site {
	return &site{files: fstest.MapFS{
		"hugo.toml":                   {Data: []byte("title = \"Korenaar\"\n")},
		"layouts/index.html":          {Data: []byte("<h1>{{ .Title }}</h1>\n")},
		"layouts/partials/head.html":  {Data: []byte("<title>{{ .Site.Title }}</title>\n")},
		"static/images/loaf.png":      {Data: []byte("\x89PNG\r\n\x1a\n\x00\x00")},
		"static/film.MP4":             {Data: bytes.Repeat([]byte{0}, 1<<20+1)},
		"static/data.zip":             {Data: []byte("PK\x03\x04\x00")},
		"public/index.html":           {Data: []byte("built")},
		"resources/_gen/images/a.png": {Data: []byte("built")},
		"node_modules/x/index.js":     {Data: []byte("x")},
		"outside":                     {Data: []byte("/etc"), Mode: 0o777 | 1<<27},
	}}
}

func TestTheTreeLeavesOutVersionControlAndBuildOutput(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSiteWithFiles()}
	entries, err := service(t, src, &syncer{}).Files(context.Background(), anImplementer(project), project)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]content.FileEntry{}
	for _, entry := range entries {
		got[entry.Path] = entry
	}
	for _, want := range []string{"hugo.toml", "layouts", "layouts/index.html", "layouts/partials/head.html", "static/images/loaf.png"} {
		if _, ok := got[want]; !ok {
			t.Errorf("the tree lacks %s", want)
		}
	}
	for _, unwanted := range []string{"public", "public/index.html", "resources/_gen", "node_modules"} {
		if _, ok := got[unwanted]; ok {
			t.Errorf("the tree lists %s", unwanted)
		}
	}
	if !got["layouts"].Dir || got["hugo.toml"].Dir || !got["outside"].Link {
		t.Errorf("kinds = %+v", got)
	}
}

func TestOnlySomebodyWhoSeesTechnicalDetailGetsTheFiles(t *testing.T) {
	project := uuid.New()
	files := service(t, &sources{site: aSiteWithFiles()}, &syncer{})
	if _, err := files.Files(context.Background(), anEditor(project), project); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("Files() for an editor error = %v", err)
	}
	if _, err := files.ReadFile(context.Background(), anEditor(project), project, "hugo.toml"); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("ReadFile() for an editor error = %v", err)
	}
	if _, err := files.WriteFile(context.Background(), anEditor(project), project, "hugo.toml", "x", ""); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("WriteFile() for an editor error = %v", err)
	}
}

func TestAPathThatLeavesTheProjectIsRefused(t *testing.T) {
	project := uuid.New()
	files := service(t, &sources{site: aSiteWithFiles()}, &syncer{})
	for _, name := range []string{"../other/hugo.toml", "/etc/passwd", ".git/config", "outside", "layouts/../hugo.toml", "", "."} {
		if _, err := files.ReadFile(context.Background(), anImplementer(project), project, name); !errors.Is(err, httpapi.ErrRefused) {
			t.Errorf("ReadFile(%q) error = %v, want a refusal", name, err)
		}
	}
	if _, err := files.ReadFile(context.Background(), anImplementer(project), project, "layouts/nothing.html"); !errors.Is(err, httpapi.ErrMissing) {
		t.Errorf("ReadFile(missing) error = %v", err)
	}
}

func TestAFileIsReadAndSavedAsAChange(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSiteWithFiles()}
	sync := &syncer{}
	files := service(t, src, sync)
	implementer := anImplementer(project)

	read, err := files.ReadFile(context.Background(), implementer, project, "layouts/index.html")
	if err != nil || read.Text != "<h1>{{ .Title }}</h1>\n" || read.Sum == "" || read.Unusable != "" {
		t.Fatalf("ReadFile() = %+v, %v", read, err)
	}
	saved, err := files.WriteFile(context.Background(), implementer, project, "layouts/index.html", "<h1 class=\"big\">{{ .Title }}</h1>\n", read.Sum)
	if err != nil {
		t.Fatal(err)
	}
	if len(src.site.saved) != 1 || string(src.site.saved[0].Writes["layouts/index.html"]) != "<h1 class=\"big\">{{ .Title }}</h1>\n" || src.site.saved[0].Message != "Edit index.html" {
		t.Errorf("saved = %+v", src.site.saved)
	}
	if saved.Sum == read.Sum {
		t.Error("the sum did not change with the text")
	}
	if _, err := files.WriteFile(context.Background(), implementer, project, "layouts/index.html", "late", read.Sum); !errors.Is(err, httpapi.ErrRefused) {
		t.Errorf("writing over a newer version error = %v, want a refusal", err)
	}
}

func TestAPictureIsNotOpenedAsText(t *testing.T) {
	project := uuid.New()
	files := service(t, &sources{site: aSiteWithFiles()}, &syncer{})
	read, err := files.ReadFile(context.Background(), anImplementer(project), project, "static/images/loaf.png")
	if err != nil || read.Unusable != "binary" || read.Text != "" || read.Preview != "image" {
		t.Errorf("ReadFile(png) = %+v, %v", read, err)
	}
	if _, err := files.WriteFile(context.Background(), anImplementer(project), project, "static/images/loaf.png", "x", read.Sum); !errors.Is(err, httpapi.ErrRefused) {
		t.Errorf("WriteFile(png) error = %v", err)
	}
}

func TestAFileTheBrowserCanShowSaysHow(t *testing.T) {
	project := uuid.New()
	files := service(t, &sources{site: aSiteWithFiles()}, &syncer{})
	for name, want := range map[string]content.FileText{
		"static/film.MP4":    {Unusable: "too-large", Preview: "video"},
		"static/data.zip":    {Unusable: "binary"},
		"layouts/index.html": {},
	} {
		read, err := files.ReadFile(context.Background(), anImplementer(project), project, name)
		if err != nil || read.Unusable != want.Unusable || read.Preview != want.Preview {
			t.Errorf("ReadFile(%s) = %+v, %v", name, read, err)
		}
	}
}

func TestAPictureIsServedAsItself(t *testing.T) {
	project := uuid.New()
	files := service(t, &sources{site: aSiteWithFiles()}, &syncer{})
	opened, err := files.OpenFile(context.Background(), anImplementer(project), project, "static/images/loaf.png")
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(opened.Content)
	if err != nil || string(data) != "\x89PNG\r\n\x1a\n\x00\x00" {
		t.Errorf("content = %q, %v", data, err)
	}
	if opened.Name != "loaf.png" || opened.Type != "image/png" || opened.Preview != "image" {
		t.Errorf("opened = %+v", opened)
	}
	if err := opened.Close(); err != nil {
		t.Error(err)
	}
}

func TestOnlyAKnownKindIsServed(t *testing.T) {
	project := uuid.New()
	files := service(t, &sources{site: aSiteWithFiles()}, &syncer{})
	implementer := anImplementer(project)
	for _, name := range []string{"layouts/index.html", "static/data.zip", "hugo.toml"} {
		if _, err := files.OpenFile(context.Background(), implementer, project, name); !errors.Is(err, httpapi.ErrRefused) {
			t.Errorf("OpenFile(%s) error = %v, want a refusal", name, err)
		}
	}
	for _, name := range []string{"../etc/x.png", ".git/a.png", "outside.png"} {
		if _, err := files.OpenFile(context.Background(), implementer, project, name); err == nil {
			t.Errorf("OpenFile(%s) succeeded", name)
		}
	}
	if _, err := files.OpenFile(context.Background(), implementer, project, "static/missing.png"); !errors.Is(err, httpapi.ErrMissing) {
		t.Errorf("a missing picture error = %v", err)
	}
	if _, err := files.OpenFile(context.Background(), anEditor(project), project, "static/images/loaf.png"); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("OpenFile() for an editor error = %v", err)
	}
}
