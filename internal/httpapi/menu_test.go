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

func (s *stubMenu) CollectionItems(_ context.Context, _ Actor, organisation, project, collection string) (CollectionItemsView, error) {
	s.asked = organisation + "/" + project + "/" + collection
	if s.err != nil {
		return CollectionItemsView{}, s.err
	}
	return CollectionItemsView{Key: collection, Title: "Posts", ItemName: "post", Total: 1,
		Items:   []CollectionItemView{{Key: "posts/hello", Title: "Hello", Status: "changed"}},
		Folders: []CollectionFolderView{{Name: "en", Label: "en", Count: 1}},
	}, nil
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
		if recorder := signedInCall(t, withMenu(&stubMenu{}, true), http.MethodPost, path, nil); recorder.Code != http.StatusMethodNotAllowed {
			t.Errorf("POST %s = %d, want 405", path, recorder.Code)
		}
	}
}
