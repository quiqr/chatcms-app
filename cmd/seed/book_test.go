package main

import (
	"io/fs"
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

func TestTheRecipesBookPreviewsAsAPDF(t *testing.T) {
	book := fstest.MapFS{
		"quiqr/runtime.yaml": {Data: []byte("preview:\n  type: web\n  command: [quarto, preview, --to, html]\n")},
		"index.qmd":          {Data: []byte("# Preface\n")},
	}
	content, err := previewedAsPDF(book)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := fs.ReadFile(content, "quiqr/runtime.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if want := "preview:\n  type: pdf\n  open: /Korenaar-Recipes.pdf\n  command: [quarto, preview, --to, typst]\n"; string(runtime) != want {
		t.Fatalf("runtime = %q, want %q", runtime, want)
	}
	if index, err := fs.ReadFile(content, "index.qmd"); err != nil || string(index) != "# Preface\n" {
		t.Fatalf("index = %q, %v", index, err)
	}
	var names []string
	if err := fs.WalkDir(content, ".", func(name string, entry fs.DirEntry, err error) error {
		if err == nil && !entry.IsDir() {
			names = append(names, name)
		}
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(names, ","); got != "index.qmd,quiqr/runtime.yaml" {
		t.Fatalf("walked %s", got)
	}
}

func TestTheRecipesBookSaysWhenItsRuntimeChanged(t *testing.T) {
	book := fstest.MapFS{"quiqr/runtime.yaml": {Data: []byte("preview:\n  type: pdf\n")}}
	if _, err := previewedAsPDF(book); err == nil {
		t.Fatal("a runtime that no longer matches was taken without a word")
	}
}

func TestTheFixtureBookStillMatches(t *testing.T) {
	if _, err := previewedAsPDF(os.DirFS("../../test/fixtures/quarto-book")); err != nil {
		t.Fatal(err)
	}
}
