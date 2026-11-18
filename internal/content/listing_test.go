package content_test

import (
	"context"
	"io/fs"
	"strings"
	"sync/atomic"
	"testing"
	"testing/fstest"
	"time"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/content"
	"github.com/quiqr/chatcms-app/internal/gitsync"
	"github.com/quiqr/chatcms-app/internal/model"
)

type countingFS struct {
	fs.FS
	reads *atomic.Int64
}

func (c countingFS) Open(name string) (fs.File, error) {
	if strings.HasSuffix(name, ".md") {
		c.reads.Add(1)
	}
	return c.FS.Open(name)
}

type versionedSite struct {
	*site
	reads    *atomic.Int64
	versions []gitsync.Version
}

func (v *versionedSite) FS() fs.FS { return countingFS{FS: v.site.files, reads: v.reads} }

func (v *versionedSite) History(ctx context.Context, limit int) ([]gitsync.Version, error) {
	if limit < len(v.versions) {
		return v.versions[:limit], nil
	}
	return v.versions, nil
}

type versionedSources struct {
	sources
	site *versionedSite
}

func (s *versionedSources) Site(ctx context.Context, personID string, projectID uuid.UUID) (content.Site, func() error, error) {
	return s.site, func() error { return nil }, nil
}

func aJournal() *site {
	return &site{
		files: fstest.MapFS{
			"content/post/a-spring-sale.md": {Data: []byte("---\ntitle: Spring sale\ndate: 2024-03-01\nsummary: Plums are in\n---\n")},
			"content/post/b-autumn.md":      {Data: []byte("---\ntitle: Autumn walk\ndate: 2023-10-01\nauthor: Anna\n---\n")},
			"content/post/en/c-hello.md":    {Data: []byte("---\ntitle: Hello\ndate: 2024-05-01\nimage: /images/loaf.png\n---\n")},
			"content/post/en/d-old.md":      {Data: []byte("---\ntitle: Zebra\ndate: 2020-01-01\ndraft: true\n---\n")},
			"content/post/nl/e-hallo.md":    {Data: []byte("---\ntitle: Hallo\n---\n")},
		},
		model: &model.Model{Collections: []model.Collection{{
			Key: "posts", Title: "Posts", ItemTitle: "post", Folder: "content/post/",
			Fields: []model.Field{{Key: "title", Type: "string"}, {Key: "image", Type: "image-select"}, {Key: "draft", Type: "boolean"}},
		}}},
	}
}

func keysOf(listing content.Listing) string {
	keys := []string{}
	for _, item := range listing.Items {
		keys = append(keys, strings.TrimPrefix(item.Key, "posts/"))
	}
	return strings.Join(keys, ",")
}

func listed(t *testing.T, query content.Query) content.Listing {
	t.Helper()
	project := uuid.New()
	listing, err := service(t, &sources{site: aJournal()}, &syncer{}).
		Items(context.Background(), anEditor(project), project, "posts", query)
	if err != nil {
		t.Fatal(err)
	}
	return listing
}

func TestItemsAreSortedAsAsked(t *testing.T) {
	for sort, want := range map[string]string{
		"":       "en/c-hello,a-spring-sale,b-autumn,en/d-old,nl/e-hallo",
		"newest": "en/c-hello,a-spring-sale,b-autumn,en/d-old,nl/e-hallo",
		"oldest": "en/d-old,b-autumn,a-spring-sale,en/c-hello,nl/e-hallo",
		"title":  "b-autumn,nl/e-hallo,en/c-hello,a-spring-sale,en/d-old",
	} {
		if got := keysOf(listed(t, content.Query{Sort: sort})); got != want {
			t.Errorf("sort %q = %s, want %s", sort, got, want)
		}
	}
}

func TestItemsAreSearchedByTitleAndSummary(t *testing.T) {
	if got := keysOf(listed(t, content.Query{Q: "plums"})); got != "a-spring-sale" {
		t.Errorf("summary search = %s", got)
	}
	listing := listed(t, content.Query{Q: "HAL"})
	if got := keysOf(listing); got != "nl/e-hallo" || listing.Total != 1 || listing.All != 5 {
		t.Errorf("title search = %s, total %d of %d", got, listing.Total, listing.All)
	}
}

func TestItemsAreFilteredByFolder(t *testing.T) {
	listing := listed(t, content.Query{Folder: "en"})
	if got := keysOf(listing); got != "en/c-hello,en/d-old" || listing.Total != 2 {
		t.Errorf("folder en = %s", got)
	}
	if len(listing.Folders) != 2 || listing.Folders[0].Name != "en" || listing.Folders[0].Count != 2 {
		t.Errorf("folders = %+v", listing.Folders)
	}
}

func TestItemsArePaged(t *testing.T) {
	listing := listed(t, content.Query{Page: 2, PageSize: 2})
	if got := keysOf(listing); got != "b-autumn,en/d-old" || listing.Total != 5 {
		t.Errorf("page 2 = %s, total %d", got, listing.Total)
	}
	if got := keysOf(listed(t, content.Query{Page: 4, PageSize: 2})); got != "" {
		t.Errorf("a page past the end = %s", got)
	}
	if got := len(listed(t, content.Query{Page: 1}).Items); got != 5 {
		t.Errorf("a page of the default size holds %d", got)
	}
}

func TestNeighboursFollowTheQuery(t *testing.T) {
	middle := listed(t, content.Query{Around: "posts/a-spring-sale"})
	if middle.Around == nil || middle.Around.Index != 2 || middle.Around.Previous != "posts/en/c-hello" || middle.Around.Next != "posts/b-autumn" {
		t.Errorf("around = %+v", middle.Around)
	}
	first := listed(t, content.Query{Around: "posts/en/c-hello"})
	if first.Around.Previous != "" || first.Around.Index != 1 {
		t.Errorf("first = %+v", first.Around)
	}
	last := listed(t, content.Query{Around: "posts/nl/e-hallo"})
	if last.Around.Next != "" || last.Around.Index != 5 {
		t.Errorf("last = %+v", last.Around)
	}
	filtered := listed(t, content.Query{Folder: "en", Around: "posts/en/d-old"})
	if filtered.Around.Index != 2 || filtered.Total != 2 || filtered.Around.Previous != "posts/en/c-hello" {
		t.Errorf("filtered = %+v total %d", filtered.Around, filtered.Total)
	}
	if missing := listed(t, content.Query{Around: "posts/nothing"}); missing.Around != nil {
		t.Errorf("around a missing item = %+v", missing.Around)
	}
}

func TestADraftIsMarkedAndItemsCarryWhatTheModelHas(t *testing.T) {
	byKey := map[string]content.CollectionItem{}
	for _, item := range listed(t, content.Query{}).Items {
		byKey[item.Key] = item
	}
	if byKey["posts/en/d-old"].Status != content.StatusDraft {
		t.Errorf("draft status = %q", byKey["posts/en/d-old"].Status)
	}
	if byKey["posts/en/c-hello"].Thumbnail != "loaf.png" || byKey["posts/b-autumn"].Author != "Anna" {
		t.Errorf("items = %+v", byKey)
	}
}

func TestTheListingIsReadOncePerVersion(t *testing.T) {
	project := uuid.New()
	reads := &atomic.Int64{}
	journal := &versionedSite{site: aJournal(), reads: reads, versions: []gitsync.Version{
		{ID: "v2", Time: time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC), Paths: []string{"content/post/b-autumn.md"}},
		{ID: "v1", Time: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), Paths: []string{"content/post/a-spring-sale.md"}},
	}}
	made, err := content.NewService(&versionedSources{site: journal}, &syncer{})
	if err != nil {
		t.Fatal(err)
	}

	var first content.Listing
	first, err = made.Items(context.Background(), anEditor(project), project, "posts", content.Query{Sort: "edited"})
	if err != nil {
		t.Fatal(err)
	}
	if got := keysOf(first); !strings.HasPrefix(got, "b-autumn,a-spring-sale") {
		t.Errorf("recently edited = %s", got)
	}
	afterFirst := reads.Load()
	if afterFirst == 0 {
		t.Fatal("the first listing read no files")
	}

	if _, err := made.Items(context.Background(), anEditor(project), project, "posts", content.Query{Q: "spring"}); err != nil {
		t.Fatal(err)
	}
	if reads.Load() != afterFirst {
		t.Errorf("a second listing at the same version read %d files", reads.Load()-afterFirst)
	}

	journal.versions = append([]gitsync.Version{{ID: "v3"}}, journal.versions...)
	if _, err := made.Items(context.Background(), anEditor(project), project, "posts", content.Query{}); err != nil {
		t.Fatal(err)
	}
	if reads.Load() == afterFirst {
		t.Error("a new version did not read the files again")
	}
}
