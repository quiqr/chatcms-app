package model

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func itemKeys(items []Item) string {
	keys := make([]string, len(items))
	for i, item := range items {
		keys[i] = item.Key
	}
	return strings.Join(keys, ",")
}

func aSite() fstest.MapFS {
	return fstest.MapFS{
		"content/post/_index.md":                  {Data: []byte("---\ntitle: All\n---\n")},
		"content/post/hello.md":                   {Data: []byte("---\ntitle: Hello\ndate: 2024-02-01\n---\n")},
		"content/post/2024/spring-sale/index.md":  {Data: []byte("---\ntitle: Spring sale\ndate: 2024-03-01\n---\n")},
		"content/post/2024/spring-sale/cover.jpg": {Data: []byte("jpg")},
		"content/post/nl/hallo.markdown":          {Data: []byte("---\ntitle: Hallo\n---\n")},
		"content/post/.draft.md":                  {Data: []byte("hidden")},
		"content/post/notes.txt":                  {Data: []byte("not content")},
	}
}

func TestItemsLooksBelowTheFolder(t *testing.T) {
	items, _ := Items(aSite(), Collection{Key: "posts", Folder: "content/post/"})

	if got := itemKeys(items); got != "2024/spring-sale,_index,hello,nl/hallo" {
		t.Errorf("keys = %s", got)
	}
}

func TestItemsStayInTheFolderWhenAsked(t *testing.T) {
	no := false
	items, _ := Items(aSite(), Collection{Key: "posts", Folder: "content/post", IncludeSubdirs: &no})

	if got := itemKeys(items); got != "_index,hello" {
		t.Errorf("keys = %s", got)
	}
}

func TestAPageBundleIsOneItemNamedAfterItsFolder(t *testing.T) {
	items, _ := Items(aSite(), Collection{Key: "posts", Folder: "content/post", HideIndex: true})

	bundle := items[0]
	if bundle.Key != "2024/spring-sale" || !bundle.Bundle || bundle.File != "content/post/2024/spring-sale/index.md" {
		t.Errorf("bundle = %+v", bundle)
	}
	if bundle.Folder != "2024" || bundle.Title != "Spring sale" {
		t.Errorf("folder %q title %q", bundle.Folder, bundle.Title)
	}
	for _, item := range items {
		if strings.HasSuffix(item.File, ".jpg") {
			t.Errorf("a picture is listed: %+v", item)
		}
	}
}

func TestHideIndexHidesOnlyTheListPage(t *testing.T) {
	items, _ := Items(aSite(), Collection{Key: "posts", Folder: "content/post", HideIndex: true})

	if got := itemKeys(items); got != "2024/spring-sale,hello,nl/hallo" {
		t.Errorf("keys = %s", got)
	}
}

func TestADataFolderListsEveryDataFile(t *testing.T) {
	site := fstest.MapFS{
		"data/team/anna.yaml":        {Data: []byte("name: Anna\n")},
		"data/team/sub/bram.json":    {Data: []byte("{}")},
		"data/team/index/index.yaml": {Data: []byte("x: 1\n")},
		"data/team/readme.md":        {Data: []byte("# no")},
	}
	items, _ := Items(site, Collection{Key: "team", Folder: "data/team", Extension: "yaml"})

	if got := itemKeys(items); got != "anna,index/index,sub/bram" {
		t.Errorf("keys = %s", got)
	}
	for _, item := range items {
		if item.Bundle {
			t.Errorf("%s is a bundle outside content", item.Key)
		}
	}
}

func TestItemsAreSortedByTheSortKeyThenByPath(t *testing.T) {
	items, _ := Items(aSite(), Collection{Key: "posts", Folder: "content/post", HideIndex: true, SortKey: "date"})

	if got := itemKeys(items); got != "hello,2024/spring-sale,nl/hallo" {
		t.Errorf("keys = %s", got)
	}
}

func TestATitleFallsBackToTheName(t *testing.T) {
	site := fstest.MapFS{"content/post/spring-prices.md": {Data: []byte("no front matter\n")}}
	items, _ := Items(site, Collection{Key: "posts", Folder: "content/post"})

	if len(items) != 1 || items[0].Title != "Spring prices" {
		t.Errorf("items = %+v", items)
	}
}

func TestAFlatFileWinsOverABundleOfTheSameName(t *testing.T) {
	site := fstest.MapFS{
		"content/post/hello.md":       {Data: []byte("---\ntitle: Flat\n---\n")},
		"content/post/hello/index.md": {Data: []byte("---\ntitle: Bundle\n---\n")},
	}
	items, warnings := Items(site, Collection{Key: "posts", Folder: "content/post"})

	if len(items) != 1 || items[0].File != "content/post/hello.md" {
		t.Fatalf("items = %+v", items)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0].Message, "hello") {
		t.Errorf("warnings = %+v", warnings)
	}
}

func TestAMissingFolderHasNoItems(t *testing.T) {
	for _, folder := range []string{"", "content/absent", "/"} {
		if items, _ := Items(aSite(), Collection{Key: "x", Folder: folder}); len(items) != 0 {
			t.Errorf("Items(%q) = %+v, want none", folder, items)
		}
	}
	if items, _ := Items(nil, Collection{Key: "x", Folder: "content/post"}); items != nil {
		t.Errorf("Items(nil) = %+v", items)
	}
}

func TestFindItemFindsABundleByItsKey(t *testing.T) {
	item, ok := FindItem(aSite(), Collection{Key: "posts", Folder: "content/post"}, "2024/spring-sale")
	if !ok || item.File != "content/post/2024/spring-sale/index.md" {
		t.Errorf("FindItem() = %+v, %v", item, ok)
	}
	if _, ok := FindItem(aSite(), Collection{Key: "posts", Folder: "content/post"}, "nothing"); ok {
		t.Error("FindItem(nothing) found something")
	}
}

func TestIncludeSubdirsIsRead(t *testing.T) {
	for _, tc := range []struct {
		yaml string
		want bool
	}{
		{"", true},
		{"    includeSubdirs: true\n", true},
		{"    includeSubdirs: false\n", false},
	} {
		m := load(t, map[string]string{
			"quiqr/model/base.yaml": "collections:\n  - key: posts\n    folder: content/post/\n" + tc.yaml,
		})
		if got := m.Collections[0].LooksBelow(); got != tc.want {
			t.Errorf("%q: LooksBelow() = %v, want %v", tc.yaml, got, tc.want)
		}
	}
}

type listedItem struct {
	Key    string `json:"key"`
	File   string `json:"file"`
	Bundle bool   `json:"bundle,omitempty"`
	Folder string `json:"folder,omitempty"`
	Title  string `json:"title"`
}

func TestThePageBundleFixtureListsAsQuiqrDesktopDoes(t *testing.T) {
	root := "../../test/fixtures/sites/page-bundles"
	fsys := os.DirFS(root)
	m, err := Load(fsys)
	if err != nil {
		t.Fatal(err)
	}

	listed := map[string][]listedItem{}
	for _, collection := range m.Collections {
		items, _ := Items(fsys, collection)
		listed[collection.Key] = []listedItem{}
		for _, item := range items {
			listed[collection.Key] = append(listed[collection.Key], listedItem{
				Key: item.Key, File: item.File, Bundle: item.Bundle, Folder: item.Folder, Title: item.Title,
			})
		}
	}
	got, err := json.MarshalIndent(map[string]any{"items": listed, "navigation": m.Navigate(fsys)}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	got = append(got, '\n')

	golden := filepath.Join(root, "items.golden.json")
	if os.Getenv(updateGoldenVar) == "1" {
		if err := os.WriteFile(golden, got, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("reading the golden failed: %v; run with %s=1 to create it", err, updateGoldenVar)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("the listing differs from %s:\n%s", golden, got)
	}
}

func TestKeyOfGivesTheContentBrowsersKeyForAFile(t *testing.T) {
	m := &Model{
		Singles: []Single{{Key: "home", File: "content/_index.md"}, {Key: "about", File: "/content/about.md"}},
		Collections: []Collection{
			{Key: "posts", Folder: "content/post/"},
			{Key: "news", Folder: "content/post/news"},
			{Key: "authors", Folder: "data/authors", Extension: "yaml"},
		},
	}
	for file, want := range map[string]string{
		"content/_index.md":                      "home",
		"content/about.md":                       "about",
		"content/post/hello.md":                  "posts/hello",
		"content/post/2024/spring-sale/index.md": "posts/2024/spring-sale",
		"content/post/news/today.md":             "news/today",
		"data/authors/index.yaml":                "authors/index",
		"layouts/index.html":                     "",
		"content/postal.md":                      "",
	} {
		if got := m.KeyOf(file); got != want {
			t.Errorf("KeyOf(%q) = %q, want %q", file, got, want)
		}
	}
	var none *Model
	if none.KeyOf("content/_index.md") != "" {
		t.Error("a missing model found a key")
	}
}
