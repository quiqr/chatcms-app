//go:build integration

package content_test

import (
	"errors"
	"testing"

	"github.com/quiqr/chatcms-app/internal/httpapi"
)

func TestTheCollectionViewThroughTheWholeStack(t *testing.T) {
	h := bundleHarness(t)

	if _, err := h.web.SavePage(t.Context(), h.editor, "technative", "korenaar", "posts/autumn",
		httpapi.EditRequest{Values: map[string]any{"title": "Autumn, edited"}}); err != nil {
		t.Fatal(err)
	}

	edited, err := h.web.CollectionItems(t.Context(), h.editor, "technative", "korenaar", "posts", httpapi.ItemsQuery{Sort: "edited", Page: 1, PageSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(edited.Items) != 2 || edited.Items[0].Key != "posts/autumn" || edited.Total != 5 || edited.Sort != "edited" {
		t.Errorf("recently edited = %+v", edited)
	}

	around, err := h.web.CollectionItems(t.Context(), h.editor, "technative", "korenaar", "posts", httpapi.ItemsQuery{Around: "posts/2024/spring-sale"})
	if err != nil {
		t.Fatal(err)
	}
	if around.Around == nil || around.Around.Index != 3 || around.Around.Previous != "posts/autumn" {
		t.Errorf("around = %+v", around.Around)
	}

	address, err := h.web.ItemAddress(t.Context(), h.builder, "technative", "korenaar", "posts", httpapi.NewItemRequest{Title: "Spring sale", Folder: "2024"})
	if err != nil {
		t.Fatal(err)
	}
	if address.Free || address.File != "content/post/2024/spring-sale/index.md" {
		t.Errorf("address = %+v", address)
	}

	_, err = h.web.CreateItem(t.Context(), h.editor, "technative", "korenaar", "posts", httpapi.NewItemRequest{Title: "Spring sale", Folder: "2024"})
	var refused *httpapi.RefusedError
	if !errors.As(err, &refused) || refused.Key != "refused.item-address-taken" {
		t.Errorf("a taken address: %v", err)
	}

	made, err := h.web.CreateItem(t.Context(), h.editor, "technative", "korenaar", "posts", httpapi.NewItemRequest{Title: "Summer", Folder: "en"})
	if err != nil {
		t.Fatal(err)
	}
	if made.Key != "posts/en/summer" || !h.holds(t, "content/post/en/summer/index.md") {
		t.Errorf("made = %+v", made)
	}

	found, err := h.web.Search(t.Context(), h.editor, "technative", "korenaar", "summer")
	if err != nil {
		t.Fatal(err)
	}
	if len(found.Groups) != 1 || found.Groups[0].Items[0].Key != "posts/en/summer" {
		t.Errorf("search = %+v", found)
	}
	if found, _ := h.web.Search(t.Context(), h.editor, "technative", "korenaar", "readme"); len(found.Groups) != 0 {
		t.Errorf("an editor found the developer's page: %+v", found)
	}
}
