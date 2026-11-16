package content_test

import (
	"context"
	"errors"
	"testing"
	"testing/fstest"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/content"
	"github.com/quiqr/chatcms-app/internal/gitsync"
	"github.com/quiqr/chatcms-app/internal/httpapi"
	"github.com/quiqr/chatcms-app/internal/model"
	"github.com/quiqr/chatcms-app/internal/policy"
)

func aMenuSite() *site {
	no := false
	return &site{
		files: fstest.MapFS{
			"content/_index.md":                       {Data: []byte("---\ntitle: Home\n---\n")},
			"quiqr/home/index.md":                     {Data: []byte("read me\n")},
			"content/post/hello.md":                   {Data: []byte("---\ntitle: Hello\n---\n")},
			"content/post/2024/spring-sale/index.md":  {Data: []byte("---\ntitle: Spring sale\n---\n")},
			"content/post/2024/spring-sale/cover.jpg": {Data: []byte("jpg")},
			"content/post/en/hi.md":                   {Data: []byte("---\ntitle: Hi\n---\n")},
			"content/about.md":                        {Data: []byte("---\ntitle: About\n---\n")},
		},
		model: &model.Model{
			Singles: []model.Single{
				{Key: "home", Title: "Home", File: "content/_index.md"},
				{Key: "readme", Title: "README", File: "quiqr/home/index.md"},
			},
			Collections: []model.Collection{
				{Key: "posts", Title: "Posts", ItemTitle: "post", Folder: "content/post/"},
				{Key: "pages", Title: "Pages", Folder: "content/", HideIndex: true, IncludeSubdirs: &no},
			},
			Menu: []model.MenuSection{
				{Key: "content", Title: "Content", MenuItems: []model.MenuItem{{Key: "home"}, {Key: "pages"}, {Key: "posts"}}},
				{Key: "dev", Title: "Developer", MatchRole: "siteDeveloper", MenuItems: []model.MenuItem{{Key: "readme"}}},
				{Key: "editors", Title: "Editors", MatchRole: "contentEditor", MenuItems: []model.MenuItem{{Key: "home"}}},
				{Key: "off", Title: "Off", Disabled: true, MenuItems: []model.MenuItem{{Key: "posts"}}},
			},
		},
	}
}

func anImplementer(project uuid.UUID) policy.Actor {
	return policy.Actor{
		PersonID: uuid.NewString(),
		Projects: map[string]policy.Role{project.String(): policy.RoleImplementer},
	}
}

func groupKeys(menu content.Menu) []string {
	keys := []string{}
	for _, group := range menu.Groups {
		keys = append(keys, group.Key)
	}
	return keys
}

func TestTheMenuFollowsTheModelForAnEditor(t *testing.T) {
	project := uuid.New()
	menu, err := service(t, &sources{site: aMenuSite()}, &syncer{}).Menu(context.Background(), anEditor(project), project)
	if err != nil {
		t.Fatal(err)
	}

	if got := groupKeys(menu); len(got) != 2 || got[0] != "content" || got[1] != "editors" {
		t.Fatalf("groups = %v, want content and editors", got)
	}
	entries := menu.Groups[0].Entries
	if len(entries) != 3 || entries[0].Kind != "single" || entries[1].Key != "pages" || entries[1].Kind != "collection" {
		t.Fatalf("entries = %+v", entries)
	}
	if entries[1].Count != 1 || entries[2].Count != 3 {
		t.Errorf("counts = %d and %d, want 1 and 3", entries[1].Count, entries[2].Count)
	}
	if menu.Singles != 1 || menu.Collections != 2 {
		t.Errorf("totals = %d singles, %d collections", menu.Singles, menu.Collections)
	}
}

func TestADeveloperSectionIsForThoseWhoSeeTechnicalDetail(t *testing.T) {
	project := uuid.New()
	menu, err := service(t, &sources{site: aMenuSite()}, &syncer{}).Menu(context.Background(), anImplementer(project), project)
	if err != nil {
		t.Fatal(err)
	}
	if got := groupKeys(menu); len(got) != 3 || got[1] != "dev" {
		t.Errorf("groups = %v, want the developer section second", got)
	}
}

func TestASingleInADroppedSectionStillOpens(t *testing.T) {
	project := uuid.New()
	page, err := service(t, &sources{site: aMenuSite()}, &syncer{}).Read(context.Background(), anEditor(project), project, "readme")
	if err != nil || page.Title != "README" {
		t.Errorf("Read(readme) = %+v, %v", page, err)
	}
}

func TestTheMenuMarksWhatWaitsInsideABundle(t *testing.T) {
	project := uuid.New()
	working := aMenuSite()
	working.changes = []gitsync.Change{{Path: "content/post/2024/spring-sale/cover.jpg", Kind: gitsync.Added}}

	menu, err := service(t, &sources{site: working}, &syncer{}).Menu(context.Background(), anEditor(project), project)
	if err != nil {
		t.Fatal(err)
	}
	entries := menu.Groups[0].Entries
	if entries[0].HasChanges || entries[1].HasChanges || !entries[2].HasChanges {
		t.Errorf("entries = %+v, want only posts marked", entries)
	}
}

func TestTheMenuNeedsAModel(t *testing.T) {
	project := uuid.New()
	_, err := service(t, &sources{site: &site{files: fstest.MapFS{}}}, &syncer{}).Menu(context.Background(), anEditor(project), project)
	if !errors.Is(err, content.ErrNoModel) {
		t.Errorf("Menu() error = %v, want ErrNoModel", err)
	}
}

func TestTheMenuIsNotThereForAStranger(t *testing.T) {
	project := uuid.New()
	_, err := service(t, &sources{site: aMenuSite()}, &syncer{}).Menu(context.Background(), policy.Actor{PersonID: uuid.NewString()}, project)
	if !errors.Is(err, httpapi.ErrMissing) {
		t.Errorf("Menu() error = %v, want ErrMissing", err)
	}
}

func TestACollectionListsItsItemsWithStatusAndFolders(t *testing.T) {
	project := uuid.New()
	working := aMenuSite()
	working.changes = []gitsync.Change{
		{Path: "content/post/hello.md", Kind: gitsync.Edited, By: gitsync.Identity{Name: "Anna"}},
		{Path: "content/post/en/hi.md", Kind: gitsync.Added},
	}

	listing, err := service(t, &sources{site: working}, &syncer{}).Items(context.Background(), anEditor(project), project, "posts")
	if err != nil {
		t.Fatal(err)
	}
	if listing.ShowsPaths {
		t.Error("an editor is shown paths")
	}
	status := map[string]string{}
	for _, item := range listing.Items {
		status[item.Key] = item.Status
	}
	if status["posts/hello"] != content.StatusChanged || status["posts/en/hi"] != content.StatusNew ||
		status["posts/2024/spring-sale"] != content.StatusPublished {
		t.Errorf("status = %v", status)
	}
	if len(listing.Folders) != 2 || listing.Folders[0].Name != "2024" || listing.Folders[1].Count != 1 {
		t.Errorf("folders = %+v", listing.Folders)
	}

	implementer, err := service(t, &sources{site: working}, &syncer{}).Items(context.Background(), anImplementer(project), project, "posts")
	if err != nil || !implementer.ShowsPaths {
		t.Errorf("an implementer: %v, shows paths %v", err, implementer.ShowsPaths)
	}

	if _, err := service(t, &sources{site: working}, &syncer{}).Items(context.Background(), anEditor(project), project, "nothing"); !errors.Is(err, httpapi.ErrMissing) {
		t.Errorf("Items(nothing) error = %v, want ErrMissing", err)
	}
}

func TestWithoutAMenuTheGroupsAreTheFallback(t *testing.T) {
	project := uuid.New()
	working := aMenuSite()
	working.model.Menu = nil

	menu, err := service(t, &sources{site: working}, &syncer{}).Menu(context.Background(), anEditor(project), project)
	if err != nil {
		t.Fatal(err)
	}
	if !menu.Fallback || len(menu.Groups) != 2 || menu.Groups[0].Key != "pages" || menu.Groups[1].Key != "collections" {
		t.Errorf("menu = %+v", menu)
	}
	if len(menu.Groups[0].Entries) != 2 {
		t.Errorf("the fallback shows %d singles, want both", len(menu.Groups[0].Entries))
	}
}
