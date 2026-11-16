//go:build integration

package content_test

import (
	"errors"
	"io/fs"
	"os"
	"testing"

	"github.com/quiqr/chatcms-app/internal/gitsync"
	"github.com/quiqr/chatcms-app/internal/httpapi"
)

func bundleHarness(t *testing.T) *webHarness {
	t.Helper()
	return newWebHarnessFrom(t, os.DirFS("../../test/fixtures/sites/page-bundles"))
}

func sectionKeys(sections []httpapi.SectionView) map[string]httpapi.SectionView {
	byKey := map[string]httpapi.SectionView{}
	for _, section := range sections {
		byKey[section.Key] = section
	}
	return byKey
}

func TestTheBrowserFollowsTheMenuAndQuiqrDesktop(t *testing.T) {
	h := bundleHarness(t)

	sections, err := h.web.ListContent(t.Context(), h.editor, "technative", "korenaar")
	if err != nil {
		t.Fatal(err)
	}
	byKey := sectionKeys(sections)
	if len(sections) != 5 {
		t.Errorf("got %d sections, want singles, pages, posts, notes and team: %+v", len(sections), sections)
	}

	singles := byKey["singles"]
	if len(singles.Items) != 1 || singles.Items[0].Key != "home" {
		t.Errorf("an editor's singles = %+v, want only home", singles.Items)
	}
	if pages := byKey["pages"]; pages.Count != 1 || pages.Items[0].Key != "pages/about" || pages.Items[0].Title != "About us" {
		t.Errorf("the pages collection = %+v", pages)
	}
	posts := byKey["posts"]
	if posts.Count != 5 || posts.Items[2].Key != "posts/2024/spring-sale" || posts.Items[2].Title != "Spring sale" {
		t.Errorf("posts = %+v", posts)
	}

	implementer, err := h.web.ListContent(t.Context(), h.builder, "technative", "korenaar")
	if err != nil {
		t.Fatal(err)
	}
	if got := sectionKeys(implementer)["singles"].Items; len(got) != 2 || got[1].Key != "readme" {
		t.Errorf("an implementer's singles = %+v, want home and readme", got)
	}

	page, err := h.web.ReadPage(t.Context(), h.editor, "technative", "korenaar", "posts/2024/spring-sale")
	if err != nil || page.Title != "Spring sale" {
		t.Errorf("opening the bundle: %+v, %v", page, err)
	}
}

func (h *webHarness) publishEverything(t *testing.T) {
	t.Helper()
	repository, err := h.sites.Open(t.Context(), h.project.ID.String())
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	if _, err := repository.Publish(t.Context()); err != nil {
		t.Fatal(err)
	}
}

func TestTheMenuAndItemsThroughTheWholeStack(t *testing.T) {
	h := bundleHarness(t)
	h.publishEverything(t)

	title := "Spring sale, again"
	if _, err := h.web.SavePage(t.Context(), h.editor, "technative", "korenaar", "posts/2024/spring-sale",
		httpapi.EditRequest{Values: map[string]any{"title": title}}); err != nil {
		t.Fatal(err)
	}
	if _, err := h.web.CreatePage(t.Context(), h.editor, "technative", "korenaar", "posts", "Brand new"); err != nil {
		t.Fatal(err)
	}

	menu, err := h.web.ContentMenu(t.Context(), h.viewer, "technative", "korenaar")
	if err != nil {
		t.Fatal(err)
	}
	if len(menu.Groups) != 2 || menu.Groups[0].Key != "content" || menu.Groups[1].Key != "people" {
		t.Fatalf("a viewer's groups = %+v", menu.Groups)
	}
	if menu.Warnings != nil {
		t.Errorf("a viewer is shown model warnings: %v", menu.Warnings)
	}
	entries := map[string]httpapi.MenuEntryView{}
	for _, entry := range menu.Groups[0].Items {
		entries[entry.Key] = entry
	}
	if !entries["posts"].HasChanges || entries["notes"].HasChanges || *entries["posts"].Count != 6 {
		t.Errorf("entries = %+v", entries)
	}

	items, err := h.web.CollectionItems(t.Context(), h.editor, "technative", "korenaar", "posts")
	if err != nil {
		t.Fatal(err)
	}
	status := map[string]string{}
	for _, item := range items.Items {
		status[item.Key] = item.Status
		if item.Path != "" {
			t.Errorf("an editor is sent the path %s", item.Path)
		}
	}
	if status["posts/2024/spring-sale"] != "changed" || status["posts/brand-new"] != "new" || status["posts/autumn"] != "published" {
		t.Errorf("status = %v", status)
	}
	if len(items.Folders) != 2 {
		t.Errorf("folders = %+v", items.Folders)
	}

	implementer, err := h.web.CollectionItems(t.Context(), h.builder, "technative", "korenaar", "posts")
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range implementer.Items {
		if item.Key == "posts/2024/spring-sale" && item.Path != "content/post/2024/spring-sale/index.md" {
			t.Errorf("an implementer's path = %q", item.Path)
		}
	}

	if _, err := h.web.CollectionItems(t.Context(), h.editor, "technative", "korenaar", "nothing"); err == nil {
		t.Error("a missing collection was listed")
	}
}

func (h *webHarness) holds(t *testing.T, name string) bool {
	t.Helper()
	repository, err := h.sites.Open(t.Context(), h.project.ID.String())
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	_, err = fs.Stat(repository.FS(), name)
	return err == nil
}

const bundle = "content/post/2024/spring-sale/"

func TestRenamingABundleMovesItsFolder(t *testing.T) {
	h := bundleHarness(t)

	page, err := h.web.RenamePage(t.Context(), h.editor, "technative", "korenaar", "posts/2024/spring-sale", "Summer sale")
	if err != nil {
		t.Fatal(err)
	}
	if page.Key != "posts/2024/summer-sale" || page.Title != "Summer sale" {
		t.Errorf("renamed = %+v", page)
	}
	for _, name := range []string{"content/post/2024/summer-sale/index.md", "content/post/2024/summer-sale/cover.jpg"} {
		if !h.holds(t, name) {
			t.Errorf("%s is missing", name)
		}
	}
	for _, name := range []string{bundle + "index.md", bundle + "cover.jpg"} {
		if h.holds(t, name) {
			t.Errorf("%s is still there", name)
		}
	}

	if _, err := h.web.RenamePage(t.Context(), h.editor, "technative", "korenaar", "posts/2024/summer-sale", "Summer sale"); err != nil {
		t.Errorf("renaming to the same name: %v", err)
	}
}

func TestRenamingAFileInASubfolderKeepsItThere(t *testing.T) {
	h := bundleHarness(t)

	page, err := h.web.RenamePage(t.Context(), h.editor, "technative", "korenaar", "posts/en/hello", "Hi there")
	if err != nil {
		t.Fatal(err)
	}
	if page.Key != "posts/en/hi-there" || !h.holds(t, "content/post/en/hi-there.md") {
		t.Errorf("renamed = %+v", page)
	}
}

func TestRenamingOntoATakenNameIsRefused(t *testing.T) {
	h := bundleHarness(t)

	_, err := h.web.RenamePage(t.Context(), h.editor, "technative", "korenaar", "posts/undated", "Autumn")
	if !errors.Is(err, httpapi.ErrRefused) {
		t.Errorf("RenamePage() error = %v, want refused", err)
	}
	if !h.holds(t, "content/post/undated.md") {
		t.Error("the page moved anyway")
	}
}

func TestCopyingABundleCopiesEverythingInIt(t *testing.T) {
	h := bundleHarness(t)

	page, err := h.web.DuplicatePage(t.Context(), h.editor, "technative", "korenaar", "posts/2024/spring-sale", "Spring sale two")
	if err != nil {
		t.Fatal(err)
	}
	if page.Key != "posts/2024/spring-sale-two" {
		t.Errorf("copied = %+v", page)
	}
	for _, name := range []string{
		"content/post/2024/spring-sale-two/index.md", "content/post/2024/spring-sale-two/cover.jpg",
		bundle + "index.md", bundle + "cover.jpg",
	} {
		if !h.holds(t, name) {
			t.Errorf("%s is missing", name)
		}
	}
}

func TestRemovingABundleRemovesItsFolderAndCanBeUndone(t *testing.T) {
	h := bundleHarness(t)

	repository, err := h.sites.Open(t.Context(), h.project.ID.String())
	if err != nil {
		t.Fatal(err)
	}
	before, err := repository.History(t.Context(), 1)
	repository.Close()
	if err != nil {
		t.Fatal(err)
	}

	if err := h.web.DeletePage(t.Context(), h.editor, "technative", "korenaar", "posts/2024/spring-sale"); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{bundle + "index.md", bundle + "cover.jpg"} {
		if h.holds(t, name) {
			t.Errorf("%s is still there", name)
		}
	}

	repository, err = h.sites.Open(t.Context(), h.project.ID.String())
	if err != nil {
		t.Fatal(err)
	}
	after, err := repository.History(t.Context(), 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != 2 {
		t.Errorf("removing made %d versions, want one", len(after)-1)
	}
	if _, err := repository.GoBackTo(t.Context(), before[0].ID, gitsync.Identity{Name: "Anna", Email: "anna@example.org"}, "Bring it back"); err != nil {
		t.Fatal(err)
	}
	repository.Close()
	for _, name := range []string{bundle + "index.md", bundle + "cover.jpg"} {
		if !h.holds(t, name) {
			t.Errorf("%s was not brought back", name)
		}
	}
}

func TestANewPostIsABundle(t *testing.T) {
	h := bundleHarness(t)

	page, err := h.web.CreatePage(t.Context(), h.editor, "technative", "korenaar", "posts", "Brand new")
	if err != nil {
		t.Fatal(err)
	}
	if page.Key != "posts/brand-new" || !h.holds(t, "content/post/brand-new/index.md") {
		t.Errorf("made = %+v", page)
	}
	if _, err := h.web.CreatePage(t.Context(), h.editor, "technative", "korenaar", "posts", "Autumn"); !errors.Is(err, httpapi.ErrRefused) {
		t.Errorf("making over a bundle: %v, want refused", err)
	}
}
