package content

import (
	"context"
	"io/fs"
	"path"
	"strings"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/gitsync"
	"github.com/quiqr/chatcms-app/internal/httpapi"
	"github.com/quiqr/chatcms-app/internal/model"
	"github.com/quiqr/chatcms-app/internal/policy"
)

const developerRole = "siteDeveloper"

type Menu struct {
	Groups      []Group
	Singles     int
	Collections int
	Warnings    []model.Warning
	Technical   bool
	Fallback    bool
}

type Group struct {
	Key     string
	Title   string
	Entries []Entry
}

type Entry struct {
	Key        string
	Kind       string
	Title      string
	ItemName   string
	Count      int
	HasChanges bool
}

type CollectionItem struct {
	Key       string
	Title     string
	Folder    string
	File      string
	Bundle    bool
	Status    string
	Waiting   bool
	By        string
	SortValue string
}

type Folder struct {
	Name  string
	Count int
}

type Listing struct {
	Collection model.Collection
	Items      []CollectionItem
	Folders    []Folder
	ShowsPaths bool
}

const (
	StatusPublished = "published"
	StatusChanged   = "changed"
	StatusNew       = "new"
)

func (s *Service) Menu(ctx context.Context, actor policy.Actor, projectID uuid.UUID) (Menu, error) {
	resource, err := s.reachable(ctx, actor, policy.ActionReadContent, projectID)
	if err != nil {
		return Menu{}, err
	}

	site, done, err := s.sources.Site(ctx, projectID)
	if err != nil {
		return Menu{}, err
	}
	defer done()

	m, err := site.Model()
	if err != nil {
		return Menu{}, noModel(err)
	}
	changes, err := site.Changes(ctx)
	if err != nil {
		return Menu{}, err
	}

	technical := policy.Can(actor, policy.ActionSeeTechnicalDetail, resource)
	menu := menuFor(m, site.FS(), waitingFrom(changes), technical)
	menu.Technical = technical
	return menu, nil
}

func menuFor(m *model.Model, fsys fs.FS, waiting map[string]changeBy, technical bool) Menu {
	navigation := m.Navigate(fsys)
	menu := Menu{Warnings: navigation.Warnings, Groups: []Group{}, Fallback: len(m.Menu) == 0}
	seenSingles := map[string]bool{}
	seenCollections := map[string]bool{}

	for _, section := range navigation.Sections {
		if section.MatchRole == developerRole && !technical {
			continue
		}
		group := Group{Key: section.Key, Title: section.Title, Entries: []Entry{}}
		if group.Key == "" {
			group.Key = strings.ToLower(section.Title)
		}
		for _, entry := range section.Entries {
			shown := Entry{Key: entry.Key, Kind: entry.Kind, Title: entry.Title, ItemName: entry.ItemName}
			switch entry.Kind {
			case model.KindPage:
				shown.Kind = "single"
				if single, ok := m.Single(entry.Key); ok {
					_, shown.HasChanges = waiting[single.File]
				}
				seenSingles[entry.Key] = true
			case model.KindCollection:
				if entry.Count != nil {
					shown.Count = *entry.Count
				}
				if collection, ok := m.Collection(entry.Key); ok {
					items, _ := model.Items(fsys, collection)
					for _, item := range items {
						if _, changed := waitingFor(item, waiting); changed {
							shown.HasChanges = true
							break
						}
					}
				}
				seenCollections[entry.Key] = true
			}
			group.Entries = append(group.Entries, shown)
		}
		menu.Groups = append(menu.Groups, group)
	}
	menu.Singles = len(seenSingles)
	menu.Collections = len(seenCollections)
	return menu
}

func waitingFor(item model.Item, waiting map[string]changeBy) (changeBy, bool) {
	if change, ok := waiting[item.File]; ok {
		return change, true
	}
	if !item.Bundle {
		return changeBy{}, false
	}
	folder := path.Dir(item.File) + "/"
	var latest changeBy
	found := false
	for file, change := range waiting {
		if strings.HasPrefix(file, folder) && (!found || change.At.After(latest.At)) {
			latest, found = change, true
		}
	}
	return latest, found
}

func (s *Service) Items(ctx context.Context, actor policy.Actor, projectID uuid.UUID, collectionKey string) (Listing, error) {
	resource, err := s.reachable(ctx, actor, policy.ActionReadContent, projectID)
	if err != nil {
		return Listing{}, err
	}

	site, done, err := s.sources.Site(ctx, projectID)
	if err != nil {
		return Listing{}, err
	}
	defer done()

	if _, err := site.Model(); err != nil {
		return Listing{}, noModel(err)
	}
	collection, err := collectionOf(site, collectionKey)
	if err != nil {
		return Listing{}, err
	}

	changes, err := site.Changes(ctx)
	if err != nil {
		return Listing{}, err
	}
	added := map[string]bool{}
	for _, change := range changes {
		if change.Kind == gitsync.Added {
			added[change.Path] = true
		}
	}

	listing := Listing{
		Collection: collection,
		Items:      []CollectionItem{},
		Folders:    []Folder{},
		ShowsPaths: policy.Can(actor, policy.ActionSeeTechnicalDetail, resource),
	}
	counted := map[string]int{}
	var order []string
	waiting := waitingFrom(changes)

	items, _ := model.Items(site.FS(), collection)
	for _, item := range items {
		shown := CollectionItem{
			Key:       collection.Key + "/" + item.Key,
			Title:     item.Title,
			Folder:    item.Folder,
			File:      item.File,
			Bundle:    item.Bundle,
			Status:    StatusPublished,
			SortValue: item.SortValue,
		}
		if change, waiting := waitingFor(item, waiting); waiting {
			shown.Waiting = true
			shown.By = change.By
			shown.Status = StatusChanged
			if added[item.File] {
				shown.Status = StatusNew
			}
		}
		listing.Items = append(listing.Items, shown)

		if item.Folder != "" {
			top, _, _ := strings.Cut(item.Folder, "/")
			if _, seen := counted[top]; !seen {
				order = append(order, top)
			}
			counted[top]++
		}
	}
	for _, name := range order {
		listing.Folders = append(listing.Folders, Folder{Name: name, Count: counted[name]})
	}
	return listing, nil
}

func (s *Service) reachable(ctx context.Context, actor policy.Actor, action policy.Action, projectID uuid.UUID) (policy.Resource, error) {
	resource, err := s.sources.Resource(ctx, projectID)
	if err != nil {
		return policy.Resource{}, err
	}
	if !policy.Can(actor, policy.ActionReadContent, resource) {
		return policy.Resource{}, httpapi.ErrMissing
	}
	if !policy.Can(actor, action, resource) {
		return policy.Resource{}, httpapi.ErrNotAllowed
	}
	return resource, nil
}
