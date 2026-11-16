package content

import (
	"context"
	"io/fs"
	"path"
	"strings"

	"github.com/google/uuid"

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

const (
	StatusPublished = "published"
	StatusChanged   = "changed"
	StatusNew       = "new"
	StatusDraft     = "draft"
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

const SearchLimit = 20

type Found struct {
	Key   string
	Title string
	Kind  string
}

type FoundGroup struct {
	Title string
	Items []Found
}

func (s *Service) Search(ctx context.Context, actor policy.Actor, projectID uuid.UUID, q string) ([]FoundGroup, error) {
	resource, err := s.reachable(ctx, actor, policy.ActionReadContent, projectID)
	if err != nil {
		return nil, err
	}
	looking := strings.ToLower(strings.TrimSpace(q))
	if looking == "" {
		return []FoundGroup{}, nil
	}

	site, done, err := s.sources.Site(ctx, projectID)
	if err != nil {
		return nil, err
	}
	defer done()

	m, err := site.Model()
	if err != nil {
		return nil, noModel(err)
	}

	menu := menuFor(m, site.FS(), map[string]changeBy{}, policy.Can(actor, policy.ActionSeeTechnicalDetail, resource))
	groups := []FoundGroup{}
	seen := map[string]bool{}
	left := SearchLimit

	add := func(title string, found Found) {
		for i := range groups {
			if groups[i].Title == title {
				groups[i].Items = append(groups[i].Items, found)
				return
			}
		}
		groups = append(groups, FoundGroup{Title: title, Items: []Found{found}})
	}

	for _, group := range menu.Groups {
		for _, entry := range group.Entries {
			if left == 0 {
				return groups, nil
			}
			if seen[entry.Kind+"/"+entry.Key] {
				continue
			}
			seen[entry.Kind+"/"+entry.Key] = true

			if entry.Kind == "single" {
				if strings.Contains(strings.ToLower(entry.Title), looking) {
					add(group.Title, Found{Key: entry.Key, Title: entry.Title, Kind: "single"})
					left--
				}
				continue
			}
			collection, ok := m.Collection(entry.Key)
			if !ok {
				continue
			}
			items, _ := s.itemsOf(ctx, projectID, site, collection)
			for _, item := range items {
				if left == 0 {
					break
				}
				if strings.Contains(strings.ToLower(item.Title), looking) {
					add(entry.Title, Found{Key: collection.Key + "/" + item.Key, Title: item.Title, Kind: "item"})
					left--
				}
			}
		}
	}
	return groups, nil
}
