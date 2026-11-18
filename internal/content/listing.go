package content

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/gitsync"
	"github.com/quiqr/chatcms-app/internal/model"
	"github.com/quiqr/chatcms-app/internal/policy"
)

const (
	SortNewest = "newest"
	SortOldest = "oldest"
	SortTitle  = "title"
	SortEdited = "edited"
	SortModel  = "model"

	DefaultPageSize = 20
	historyDepth    = 2000
)

type Query struct {
	Q        string
	Folder   string
	Sort     string
	Page     int
	PageSize int
	Around   string
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
	Date      string
	Author    string
	Summary   string
	Thumbnail string
	Edited    time.Time
	recency   int
}

type Folder struct {
	Name  string
	Count int
}

type Neighbours struct {
	Index    int
	Previous string
	Next     string
}

type Listing struct {
	Collection model.Collection
	Items      []CollectionItem
	Total      int
	All        int
	Folders    []Folder
	Sort       string
	ShowsPaths bool
	Around     *Neighbours
}

type historied interface {
	History(ctx context.Context, limit int) ([]gitsync.Version, error)
}

type touched struct {
	at   time.Time
	rank int
}

type cachedListing struct {
	version string
	items   []model.Item
	edited  map[string]touched
}

type listings struct {
	mu    sync.Mutex
	saved map[string]cachedListing
}

func (l *listings) get(key, version string) (cachedListing, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	found, ok := l.saved[key]
	return found, ok && version != "" && found.version == version
}

func (l *listings) put(key string, listing cachedListing) {
	if listing.version == "" {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.saved == nil {
		l.saved = map[string]cachedListing{}
	}
	l.saved[key] = listing
}

func (s *Service) itemsOf(ctx context.Context, projectID uuid.UUID, site Site, collection model.Collection) ([]model.Item, map[string]touched) {
	version := ""
	var versions []gitsync.Version
	if history, ok := site.(historied); ok {
		if latest, err := history.History(ctx, 1); err == nil && len(latest) > 0 {
			version = latest[0].ID
		}
	}

	key := projectID.String() + "/" + collection.Key
	if found, ok := s.cache.get(key, version); ok {
		return found.items, found.edited
	}

	items, _ := model.Items(site.FS(), collection)
	edited := map[string]touched{}
	if history, ok := site.(historied); ok {
		versions, _ = history.History(ctx, historyDepth)
	}
	for rank, version := range versions {
		for _, file := range version.Paths {
			if _, seen := edited[file]; !seen {
				edited[file] = touched{at: version.Time, rank: rank + 1}
			}
		}
	}
	s.cache.put(key, cachedListing{version: version, items: items, edited: edited})
	return items, edited
}

func (s *Service) Items(ctx context.Context, actor policy.Actor, projectID uuid.UUID, collectionKey string, query Query) (Listing, error) {
	resource, err := s.reachable(ctx, actor, policy.ActionReadContent, projectID)
	if err != nil {
		return Listing{}, err
	}

	site, done, err := s.sources.Site(ctx, actor.PersonID, projectID)
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
	waiting := waitingFrom(changes)
	added := map[string]bool{}
	for _, change := range changes {
		if change.Kind == gitsync.Added {
			added[change.Path] = true
		}
	}

	found, edited := s.itemsOf(ctx, projectID, site, collection)
	all := make([]CollectionItem, 0, len(found))
	counted := map[string]int{}
	var folders []string
	for _, item := range found {
		shown := CollectionItem{
			Key:       collection.Key + "/" + item.Key,
			Title:     item.Title,
			Folder:    item.Folder,
			File:      item.File,
			Bundle:    item.Bundle,
			Status:    StatusPublished,
			Date:      item.Date,
			Author:    item.Author,
			Summary:   item.Summary,
			Thumbnail: item.Thumbnail,
		}
		last := lastEdited(item, edited)
		shown.Edited, shown.recency = last.at, last.rank
		if change, ok := waitingFor(item, waiting); ok {
			shown.Waiting, shown.By, shown.Status = true, change.By, StatusChanged
			if added[item.File] {
				shown.Status = StatusNew
			}
		}
		if item.Draft && shown.Status == StatusPublished {
			shown.Status = StatusDraft
		}
		all = append(all, shown)

		if item.Folder != "" {
			top, _, _ := strings.Cut(item.Folder, "/")
			if _, seen := counted[top]; !seen {
				folders = append(folders, top)
			}
			counted[top]++
		}
	}

	listing := Listing{
		Collection: collection,
		All:        len(all),
		Folders:    []Folder{},
		ShowsPaths: policy.Can(actor, policy.ActionSeeTechnicalDetail, resource),
		Sort:       sortOf(query.Sort, collection),
	}
	for _, name := range folders {
		listing.Folders = append(listing.Folders, Folder{Name: name, Count: counted[name]})
	}

	filtered := filter(all, query)
	order(filtered, listing.Sort)
	listing.Total = len(filtered)

	if query.Around != "" {
		for i, item := range filtered {
			if item.Key != query.Around {
				continue
			}
			around := &Neighbours{Index: i + 1}
			if i > 0 {
				around.Previous = filtered[i-1].Key
			}
			if i+1 < len(filtered) {
				around.Next = filtered[i+1].Key
			}
			listing.Around = around
			break
		}
		listing.Items = []CollectionItem{}
		return listing, nil
	}

	listing.Items = page(filtered, query)
	return listing, nil
}

func lastEdited(item model.Item, edited map[string]touched) touched {
	if !item.Bundle {
		return edited[item.File]
	}
	folder := item.File[:strings.LastIndex(item.File, "/")+1]
	var latest touched
	for file, at := range edited {
		if strings.HasPrefix(file, folder) && (latest.rank == 0 || at.rank < latest.rank) {
			latest = at
		}
	}
	return latest
}

func sortOf(asked string, collection model.Collection) string {
	switch asked {
	case SortNewest, SortOldest, SortTitle, SortEdited:
		return asked
	}
	if collection.SortKey != "" {
		return SortModel
	}
	return SortNewest
}

func filter(items []CollectionItem, query Query) []CollectionItem {
	looking := strings.ToLower(strings.TrimSpace(query.Q))
	folder := strings.Trim(query.Folder, "/")
	kept := make([]CollectionItem, 0, len(items))
	for _, item := range items {
		if folder != "" && item.Folder != folder && !strings.HasPrefix(item.Folder, folder+"/") {
			continue
		}
		if looking != "" &&
			!strings.Contains(strings.ToLower(item.Title), looking) &&
			!strings.Contains(strings.ToLower(item.Summary), looking) {
			continue
		}
		kept = append(kept, item)
	}
	return kept
}

func order(items []CollectionItem, how string) {
	switch how {
	case SortModel:
		return
	case SortTitle:
		sort.SliceStable(items, func(i, j int) bool {
			return strings.ToLower(items[i].Title) < strings.ToLower(items[j].Title)
		})
	case SortEdited:
		sort.SliceStable(items, func(i, j int) bool {
			a, b := items[i].recency, items[j].recency
			if (a == 0) != (b == 0) {
				return a != 0
			}
			return a < b
		})
	case SortOldest:
		sort.SliceStable(items, func(i, j int) bool { return dated(items[i], items[j], false) })
	default:
		sort.SliceStable(items, func(i, j int) bool { return dated(items[i], items[j], true) })
	}
}

func dated(a, b CollectionItem, newestFirst bool) bool {
	if (a.Date == "") != (b.Date == "") {
		return a.Date != ""
	}
	if a.Date == b.Date {
		return a.Key < b.Key
	}
	if newestFirst {
		return a.Date > b.Date
	}
	return a.Date < b.Date
}

func page(items []CollectionItem, query Query) []CollectionItem {
	size := query.PageSize
	if size <= 0 {
		if query.Page <= 0 {
			return items
		}
		size = DefaultPageSize
	}
	if size > 200 {
		size = 200
	}
	number := query.Page
	if number <= 0 {
		number = 1
	}
	start := (number - 1) * size
	if start >= len(items) {
		return []CollectionItem{}
	}
	end := start + size
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}
