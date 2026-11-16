package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"testing"
)

type stubMenu struct {
	asked string
	query ItemsQuery
	err   error
}

func (s *stubMenu) ContentMenu(_ context.Context, _ Actor, organisation, project string) (ContentMenuView, error) {
	s.asked = organisation + "/" + project
	if s.err != nil {
		return ContentMenuView{}, s.err
	}
	count := 3
	return ContentMenuView{
		Groups: []MenuGroupView{{Key: "content", Title: "Content", Items: []MenuEntryView{
			{Key: "home", Kind: "single", Title: "Home"},
			{Key: "posts", Kind: "collection", Title: "Posts", Count: &count, HasChanges: true},
		}}},
		Totals: MenuTotalsView{Singles: 1, Collections: 1},
	}, nil
}

func (s *stubMenu) CollectionItems(_ context.Context, _ Actor, organisation, project, collection string, query ItemsQuery) (CollectionItemsView, error) {
	s.asked = organisation + "/" + project + "/" + collection
	s.query = query
	if s.err != nil {
		return CollectionItemsView{}, s.err
	}
	return CollectionItemsView{Key: collection, Title: "Posts", ItemName: "post", Total: 1,
		Items:   []CollectionItemView{{Key: "posts/hello", Title: "Hello", Status: "changed"}},
		Folders: []CollectionFolderView{{Name: "en", Label: "en", Count: 1}},
	}, nil
}

func (s *stubMenu) CreateItem(_ context.Context, _ Actor, organisation, project, collection string, asked NewItemRequest) (PageView, error) {
	s.asked = organisation + "/" + project + "/" + collection + "=" + asked.Title + "|" + asked.Folder + "|" + asked.Slug
	if s.err != nil {
		return PageView{}, s.err
	}
	return PageView{Key: collection + "/" + asked.Slug, Title: asked.Title}, nil
}

func (s *stubMenu) ItemAddress(_ context.Context, _ Actor, organisation, project, collection string, asked NewItemRequest) (AddressView, error) {
	s.asked = organisation + "/" + project + "/" + collection + "?" + asked.Title + "|" + asked.Folder
	if s.err != nil {
		return AddressView{}, s.err
	}
	return AddressView{Slug: "spring-sale", Address: "/post/spring-sale/", Free: true}, nil
}

func (s *stubMenu) Search(_ context.Context, _ Actor, organisation, project, q string) (SearchView, error) {
	s.asked = organisation + "/" + project + "?" + q
	if s.err != nil {
		return SearchView{}, s.err
	}
	return SearchView{Groups: []SearchGroupView{{Title: "Posts", Items: []SearchItemView{{Key: "posts/hello", Title: "Hello", Kind: "item"}}}}}, nil
}

func withMenu(service MenuService, signedIn bool) http.Handler {
	known := map[string]Actor{}
	if signedIn {
		known["a-session-value"] = Actor{ID: "person-1", Email: "editor@example.org"}
	}
	return New(Options{
		Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Health: func(context.Context) Health { return Health{Database: DatabaseReachable} },
		Auth:   &stubAuth{known: known},
		Menu:   service,
	})
}

const menuBase = "/api/organisations/technative/projects/korenaar"

func TestTheMenuRoutesNeedAPerson(t *testing.T) {
	for _, path := range []string{menuBase + "/content-menu", menuBase + "/collections/posts/items"} {
		if recorder := call(t, withMenu(&stubMenu{}, false), http.MethodGet, path, nil); recorder.Code != http.StatusUnauthorized {
			t.Errorf("GET %s = %d, want 401", path, recorder.Code)
		}
	}
}

func TestTheContentMenu(t *testing.T) {
	service := &stubMenu{}
	recorder := signedInCall(t, withMenu(service, true), http.MethodGet, menuBase+"/content-menu", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	var view ContentMenuView
	if err := json.Unmarshal(recorder.Body.Bytes(), &view); err != nil {
		t.Fatal(err)
	}
	items := view.Groups[0].Items
	if service.asked != "technative/korenaar" || items[0].Count != nil || *items[1].Count != 3 || !items[1].HasChanges {
		t.Errorf("view = %+v, asked %q", view, service.asked)
	}
}

func TestACollectionsItems(t *testing.T) {
	service := &stubMenu{}
	recorder := signedInCall(t, withMenu(service, true), http.MethodGet, menuBase+"/collections/posts/items", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	var view CollectionItemsView
	if err := json.Unmarshal(recorder.Body.Bytes(), &view); err != nil {
		t.Fatal(err)
	}
	if service.asked != "technative/korenaar/posts" || view.Total != 1 || view.Items[0].Path != "" || view.Folders[0].Name != "en" {
		t.Errorf("view = %+v", view)
	}
}

func TestAMissingCollectionIsNotFound(t *testing.T) {
	recorder := signedInCall(t, withMenu(&stubMenu{err: ErrMissing}, true), http.MethodGet, menuBase+"/collections/nothing/items", nil)
	if recorder.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", recorder.Code)
	}
}

func TestTheMenuRoutesOnlyRead(t *testing.T) {
	for _, path := range []string{menuBase + "/content-menu", menuBase + "/collections/posts/items"} {
		if recorder := signedInCall(t, withMenu(&stubMenu{}, true), http.MethodDelete, path, nil); recorder.Code != http.StatusMethodNotAllowed {
			t.Errorf("DELETE %s = %d, want 405", path, recorder.Code)
		}
	}
}

func TestTheItemsRoutePassesTheQueryOn(t *testing.T) {
	service := &stubMenu{}
	recorder := signedInCall(t, withMenu(service, true), http.MethodGet,
		menuBase+"/collections/posts/items?q=spring&folder=en&sort=title&page=2&pageSize=10&around=posts%2Fhello", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	want := ItemsQuery{Q: "spring", Folder: "en", Sort: "title", Page: 2, PageSize: 10, Around: "posts/hello"}
	if service.query != want {
		t.Errorf("query = %+v, want %+v", service.query, want)
	}

	signedInCall(t, withMenu(service, true), http.MethodGet, menuBase+"/collections/posts/items?page=-3&pageSize=x", nil)
	if service.query.Page != 0 || service.query.PageSize != 0 {
		t.Errorf("bad numbers became %+v", service.query)
	}
}

func TestMakingAnItem(t *testing.T) {
	service := &stubMenu{}
	recorder := signedInCall(t, withMenu(service, true), http.MethodPost, menuBase+"/collections/posts/items",
		map[string]string{"title": "Spring sale", "folder": "en", "slug": "spring-sale"})
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d", recorder.Code)
	}
	if service.asked != "technative/korenaar/posts=Spring sale|en|spring-sale" {
		t.Errorf("asked = %q", service.asked)
	}

	refused := &stubMenu{err: &RefusedError{Key: "refused.item-address-taken", Reason: "taken"}}
	if recorder := signedInCall(t, withMenu(refused, true), http.MethodPost, menuBase+"/collections/posts/items",
		map[string]string{"title": "Spring sale"}); recorder.Code != http.StatusConflict {
		t.Errorf("a taken address = %d, want 409", recorder.Code)
	}
}

func TestAskingForAnAddress(t *testing.T) {
	service := &stubMenu{}
	recorder := signedInCall(t, withMenu(service, true), http.MethodGet, menuBase+"/collections/posts/address?title=Spring+sale&folder=en", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	var view AddressView
	if err := json.Unmarshal(recorder.Body.Bytes(), &view); err != nil {
		t.Fatal(err)
	}
	if !view.Free || view.Address != "/post/spring-sale/" || service.asked != "technative/korenaar/posts?Spring sale|en" {
		t.Errorf("view = %+v, asked %q", view, service.asked)
	}
}

func TestSearching(t *testing.T) {
	service := &stubMenu{}
	recorder := signedInCall(t, withMenu(service, true), http.MethodGet, menuBase+"/search?q=hel", nil)
	if recorder.Code != http.StatusOK || service.asked != "technative/korenaar?hel" {
		t.Fatalf("status = %d, asked %q", recorder.Code, service.asked)
	}
	if recorder := call(t, withMenu(service, false), http.MethodGet, menuBase+"/search?q=hel", nil); recorder.Code != http.StatusUnauthorized {
		t.Errorf("searching signed out = %d", recorder.Code)
	}
}
