package onboarding

import (
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

func site(files map[string]string) fstest.MapFS {
	out := fstest.MapFS{}
	for name, contents := range files {
		out[name] = &fstest.MapFile{Data: []byte(contents)}
	}
	return out
}

const aModel = `singles:
  - key: home
    title: Homepage
    file: content/_index.md
    fields:
      - key: title
        type: string
collections:
  - key: posts
    title: Posts
    folder: content/posts/
    fields:
      - key: title
        type: string
`

func TestDetectsHugoConfigurationWhereverItIs(t *testing.T) {
	for _, where := range hugoConfigurations {
		t.Run(where, func(t *testing.T) {
			inspection := Inspect(site(map[string]string{
				where:                   "baseURL = 'https://example.org'\n",
				"quiqr/model/base.yaml": aModel,
			}))

			if !inspection.IsHugoSite {
				t.Error("a Hugo configuration was not recognised")
			}
			if inspection.HugoConfiguration != where {
				t.Errorf("configuration = %q, want %q", inspection.HugoConfiguration, where)
			}
		})
	}
}

func TestDetectsAModel(t *testing.T) {
	inspection := Inspect(site(map[string]string{
		"hugo.toml":             "baseURL = 'https://example.org'\n",
		"quiqr/model/base.yaml": aModel,
	}))

	if !inspection.Ready() {
		t.Fatalf("inspection = %+v, want a site that can be edited", inspection)
	}
	if inspection.Pages != 1 || inspection.Collections != 1 {
		t.Errorf("pages = %d, collections = %d, want 1 and 1", inspection.Pages, inspection.Collections)
	}
	if inspection.ModelBaseFile != "quiqr/model/base.yaml" {
		t.Errorf("base file = %q", inspection.ModelBaseFile)
	}
	if inspection.Problem != "" {
		t.Errorf("problem = %q, want none", inspection.Problem)
	}
}

func TestHugoWithNoModelIsUsableAndSaysWhatToDo(t *testing.T) {
	inspection := Inspect(site(map[string]string{
		"hugo.toml": "baseURL = 'https://example.org'\n",
	}))

	if !inspection.IsHugoSite {
		t.Error("the site was not recognised as Hugo")
	}
	if inspection.HasModel {
		t.Error("a model was reported where there is none")
	}
	if inspection.Ready() {
		t.Error("a site with no model is reported as ready to edit")
	}
	if !strings.Contains(inspection.Problem, "quiqr/model") {
		t.Errorf("problem = %q, want it to say where a model goes", inspection.Problem)
	}
	if len(inspection.LookedFor) == 0 {
		t.Error("nothing was named as looked for")
	}
}

func TestNotAHugoSite(t *testing.T) {
	inspection := Inspect(site(map[string]string{"README.md": "just a readme"}))

	if inspection.IsHugoSite {
		t.Error("a site with no configuration was called a Hugo site")
	}
	if !strings.Contains(inspection.Problem, "Hugo") {
		t.Errorf("problem = %q", inspection.Problem)
	}
	if len(inspection.LookedFor) < 4 {
		t.Errorf("looked for %v, want the list of places that were checked", inspection.LookedFor)
	}
}

func TestAModelThatDoesNotParse(t *testing.T) {
	inspection := Inspect(site(map[string]string{
		"hugo.toml":             "baseURL = 'https://example.org'\n",
		"quiqr/model/base.yaml": "collections:\n  - key: posts\n",
	}))

	if inspection.HasModel {
		t.Error("a model that cannot be read was reported as found")
	}
	if !strings.Contains(inspection.Problem, "cannot be read") {
		t.Errorf("problem = %q, want it to say the model cannot be read", inspection.Problem)
	}
	if !strings.Contains(inspection.Problem, "folder") {
		t.Errorf("problem = %q, want the model's own reason in it", inspection.Problem)
	}
}

func TestADirectoryNamedLikeAConfigurationIsNotOne(t *testing.T) {
	fsys := fstest.MapFS{
		"hugo.toml/placeholder": &fstest.MapFile{Data: []byte("x")},
		"README.md":             &fstest.MapFile{Data: []byte("readme")},
	}

	if Inspect(fsys).IsHugoSite {
		t.Error("a directory named like a configuration file was taken for one")
	}
}

func TestASiteThatDeclaresItsRuntimeNeedsNoHugoConfiguration(t *testing.T) {
	site := fstest.MapFS{
		"_quarto.yml":           {Data: []byte("project:\n  type: book\n")},
		"quiqr/runtime.yaml":    {Data: []byte("packages: {quarto: \"1.10.18\"}\npreview: {command: [quarto, preview, \"{source}\"]}\n")},
		"quiqr/model/base.yaml": {Data: []byte("singles:\n  - key: preface\n    title: Preface\n    file: index.qmd\n    fields: []\ncollections:\n  - key: chapters\n    title: Chapters\n    folder: chapters/\n    extension: qmd\n    fields: []\n")},
		"index.qmd":             {Data: []byte("---\ntitle: Preface\n---\n")},
	}
	inspection := Inspect(site)
	if inspection.Runtime != "quarto" || inspection.IsHugoSite || !inspection.HasModel || !inspection.Ready() || inspection.Pages != 1 || inspection.Collections != 1 {
		t.Errorf("Inspect() = %+v", inspection)
	}
}

func TestARuntimeDefinitionThatCannotBeUsedIsTheProblem(t *testing.T) {
	inspection := Inspect(fstest.MapFS{"quiqr/runtime.yaml": {Data: []byte("preview: {type: video}\n")}})
	if inspection.Ready() || !strings.Contains(inspection.Problem, "runtime definition") {
		t.Errorf("Inspect() = %+v", inspection)
	}
}

func TestTheQuartoBookFixtureIsASiteWithAModel(t *testing.T) {
	inspection := Inspect(os.DirFS("../../test/fixtures/quarto-book"))
	if !inspection.Ready() || inspection.Runtime != "quarto" || inspection.Pages != 1 || inspection.Collections != 1 {
		t.Errorf("Inspect() = %+v", inspection)
	}
}
