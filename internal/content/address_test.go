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

func aFolderedSite() *site {
	return &site{
		files: fstest.MapFS{
			"content/post/en/hello/index.md": {Data: []byte("---\ntitle: Hello\n---\n")},
			"content/post/nl/.keep":          {Data: []byte("")},
			"content/_index.md":              {Data: []byte("---\ntitle: Home\n---\n")},
			"quiqr/home/index.md":            {Data: []byte("notes\n")},
		},
		model: &model.Model{
			Singles: []model.Single{
				{Key: "home", Title: "Home", File: "content/_index.md"},
				{Key: "readme", Title: "Hello developers", File: "quiqr/home/index.md"},
			},
			Collections: []model.Collection{
				{Key: "posts", Title: "Posts", Folder: "content/post/"},
				{Key: "news", Title: "News", Folder: "content/news/", PreviewURLBase: "/nieuws"},
			},
			Menu: []model.MenuSection{
				{Key: "site", Title: "Site", MenuItems: []model.MenuItem{{Key: "home"}, {Key: "posts"}}},
				{Key: "dev", Title: "Developer", MatchRole: "siteDeveloper", MenuItems: []model.MenuItem{{Key: "readme"}}},
			},
		},
	}
}

func TestTheAddressIsCheckedBeforeMaking(t *testing.T) {
	project := uuid.New()
	made := service(t, &sources{site: aFolderedSite()}, &syncer{})

	free, err := made.Address(context.Background(), anEditor(project), project, content.NewPage{Section: "posts", Title: "Spring sale", Folder: "nl"})
	if err != nil {
		t.Fatal(err)
	}
	if !free.Free || free.Slug != "spring-sale" || free.Address != "/post/nl/spring-sale/" || free.File != "" {
		t.Errorf("free = %+v", free)
	}

	taken, err := made.Address(context.Background(), anImplementer(project), project, content.NewPage{Section: "posts", Title: "Hello", Folder: "en"})
	if err != nil {
		t.Fatal(err)
	}
	if taken.Free || taken.File != "content/post/en/hello/index.md" {
		t.Errorf("taken = %+v", taken)
	}

	edited, err := made.Address(context.Background(), anEditor(project), project, content.NewPage{Section: "news", Title: "x", Slug: "Spring Prices!"})
	if err != nil || edited.Slug != "spring-prices" || edited.Address != "/nieuws/spring-prices/" {
		t.Errorf("an edited address = %+v, %v", edited, err)
	}

	for _, folder := range []string{"../etc", "absent"} {
		if _, err := made.Address(context.Background(), anEditor(project), project, content.NewPage{Section: "posts", Title: "x", Folder: folder}); !errors.Is(err, httpapi.ErrMissing) {
			t.Errorf("folder %q: %v, want ErrMissing", folder, err)
		}
	}
	if _, err := made.Address(context.Background(), aViewer(project), project, content.NewPage{Section: "posts", Title: "x"}); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("a viewer: %v, want ErrNotAllowed", err)
	}
}

func TestAnItemIsMadeInAFolder(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aFolderedSite()}

	page, err := service(t, src, &syncer{}).Create(context.Background(), anEditor(project), project,
		content.NewPage{Section: "posts", Title: "Spring sale", Folder: "nl", Slug: "lente"})
	if err != nil {
		t.Fatal(err)
	}
	if _, wrote := src.site.saved[0].Writes["content/post/nl/lente/index.md"]; !wrote || page.Key != "posts/nl/lente" {
		t.Errorf("wrote %v, key %q", src.site.saved[0].Writes, page.Key)
	}
}

func TestQuickFindSearchesTitlesThePersonMaySee(t *testing.T) {
	project := uuid.New()
	made := service(t, &sources{site: aFolderedSite()}, &syncer{})

	groups, err := made.Search(context.Background(), anEditor(project), project, "hel")
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 || groups[0].Title != "Posts" || groups[0].Items[0].Key != "posts/en/hello" {
		t.Errorf("an editor finds %+v", groups)
	}

	groups, err = made.Search(context.Background(), anImplementer(project), project, "hel")
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 2 || groups[1].Title != "Developer" {
		t.Errorf("an implementer finds %+v", groups)
	}

	if groups, _ := made.Search(context.Background(), anEditor(project), project, "  "); len(groups) != 0 {
		t.Errorf("an empty search finds %+v", groups)
	}
}

func TestQuickFindStopsAtTwenty(t *testing.T) {
	project := uuid.New()
	working := aFolderedSite()
	for i := 0; i < 30; i++ {
		working.files["content/post/many-"+string(rune('a'+i%26))+string(rune('a'+i/26))+".md"] = &fstest.MapFile{Data: []byte("---\ntitle: Many\n---\n")}
	}
	groups, err := service(t, &sources{site: working}, &syncer{}).Search(context.Background(), anEditor(project), project, "many")
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 || len(groups[0].Items) != content.SearchLimit {
		t.Errorf("found %d groups, %d items", len(groups), len(groups[0].Items))
	}
}
