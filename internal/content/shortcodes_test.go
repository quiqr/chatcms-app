package content_test

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/content"
	"github.com/quiqr/chatcms-app/internal/model"
)

func TestTheDialectFollowsTheModelThenTheRuntime(t *testing.T) {
	quarto := fstest.MapFS{"quiqr/runtime.yaml": {Data: []byte("packages:\n  quarto: \"1.10.18\"\npreview:\n  command: [quarto, preview, \"{source}\"]\n")}}
	hugo := fstest.MapFS{"quiqr/runtime.yaml": {Data: []byte("packages:\n  hugo: \"0.140.0\"\npreview:\n  command: [hugo, server]\n")}}
	for name, c := range map[string]struct {
		model *model.Model
		site  fstest.MapFS
		want  string
	}{
		"no runtime":           {&model.Model{}, fstest.MapFS{}, "hugo"},
		"a Hugo runtime":       {&model.Model{}, hugo, "hugo"},
		"a Quarto runtime":     {&model.Model{}, quarto, "quarto"},
		"another runtime":      {&model.Model{}, fstest.MapFS{"quiqr/runtime.yaml": {Data: []byte("packages:\n  eleventy: \"3.0.0\"\npreview:\n  command: [eleventy, --serve]\n")}}, ""},
		"the model names Hugo": {&model.Model{ShortcodeDialect: "hugo"}, quarto, "hugo"},
	} {
		t.Run(name, func(t *testing.T) {
			if got := content.ShortcodeDialect(c.model, c.site); got != c.want {
				t.Fatalf("dialect = %q, want %q", got, c.want)
			}
		})
	}
}

func TestAPageCarriesTheDeclaredShortcodes(t *testing.T) {
	zero := 0
	project := uuid.New()
	src := &sources{site: aSite()}
	src.site.model.Shortcodes = []model.Shortcode{
		{Key: "figure", Title: "Picture", Fields: []model.ShortcodeField{
			{Field: model.Field{Key: "src", Type: "image-select"}},
			{Field: model.Field{Key: "caption", Type: "string", Title: "Caption"}},
		}},
		{Key: "notice", Inner: "markdown", Fields: []model.ShortcodeField{
			{Field: model.Field{Key: "type", Type: "string"}, Position: &zero},
		}},
	}

	page, err := service(t, src, &syncer{}).Read(context.Background(), anEditor(project), project, "home")
	if err != nil {
		t.Fatal(err)
	}
	if page.Dialect != "hugo" || len(page.Shortcodes) != 2 {
		t.Fatalf("dialect %q, shortcodes %+v", page.Dialect, page.Shortcodes)
	}
	figure := page.Shortcodes[0]
	if figure.Title != "Picture" || len(figure.Controls) != 2 || figure.Controls[1].Title != "Caption" || figure.Controls[0].Value != nil {
		t.Fatalf("figure = %+v", figure)
	}
	notice := page.Shortcodes[1]
	if notice.Title != "Notice" || notice.Inner != "markdown" || notice.Positions["type"] != 0 || len(notice.Positions) != 1 {
		t.Fatalf("notice = %+v", notice)
	}
}

func TestAQuartoPageCarriesNoShortcodes(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSite()}
	src.site.files["quiqr/runtime.yaml"] = &fstest.MapFile{Data: []byte("packages:\n  quarto: \"1.10.18\"\npreview:\n  command: [quarto, preview, \"{source}\"]\n")}
	src.site.model.Shortcodes = []model.Shortcode{{Key: "figure"}}

	page, err := service(t, src, &syncer{}).Read(context.Background(), anEditor(project), project, "home")
	if err != nil {
		t.Fatal(err)
	}
	if page.Dialect != "quarto" || page.Shortcodes != nil {
		t.Fatalf("dialect %q, shortcodes %+v", page.Dialect, page.Shortcodes)
	}
}
