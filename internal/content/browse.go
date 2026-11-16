package content

import (
	"context"
	"time"

	"github.com/quiqr/chatcms-app/internal/gitsync"
	"github.com/quiqr/chatcms-app/internal/model"
)

// Section is a group in the content browser: the site's pages, or one
// collection of them.
type Section struct {
	Key   string
	Title string
	// ItemName is the singular word for one of its items, used in "New
	// article". It is empty for the pages section.
	ItemName string
	Count    int
	Items    []Item
}

// Item is one page in the browser.
type Item struct {
	Key   string
	Title string
	// Waiting says this page has a change that has not been published.
	Waiting bool
	By      string
	At      time.Time
}

// Browse lists what a content creator can open, in the order the model's
// menu asks for, with a mark on anything not yet published.
func Browse(ctx context.Context, site Site, waiting map[string]changeBy, technical bool) []Section {
	m, err := site.Model()
	if err != nil {
		return nil
	}

	menu := menuFor(m, site.FS(), waiting, technical)
	sections := []Section{}
	pages := Section{Key: singlesSection, Title: "Pages"}
	var collections []Section
	seen := map[string]bool{}

	for _, group := range menu.Groups {
		for _, entry := range group.Entries {
			if seen[entry.Kind+"/"+entry.Key] {
				continue
			}
			seen[entry.Kind+"/"+entry.Key] = true

			if entry.Kind == "single" {
				single, _ := m.Single(entry.Key)
				pages.Items = append(pages.Items, itemOf(single.Key, titleOr(single.Title, single.Key), single.File, waiting))
				continue
			}
			collection, ok := m.Collection(entry.Key)
			if !ok {
				continue
			}
			section := Section{
				Key:      collection.Key,
				Title:    titleOr(collection.Title, collection.Key),
				ItemName: itemNameOf(collection),
			}
			items, _ := model.Items(site.FS(), collection)
			for _, found := range items {
				item := Item{Key: collection.Key + "/" + found.Key, Title: found.Title}
				if change, ok := waitingFor(found, waiting); ok {
					item.Waiting, item.By, item.At = true, change.By, change.At
				}
				section.Items = append(section.Items, item)
			}
			section.Count = len(section.Items)
			collections = append(collections, section)
		}
	}

	pages.Count = len(pages.Items)
	if pages.Count > 0 {
		sections = append(sections, pages)
	}
	return append(sections, collections...)
}

const singlesSection = "singles"

func itemNameOf(collection model.Collection) string {
	if collection.ItemTitle != "" {
		return collection.ItemTitle
	}
	return model.DefaultItemName
}

func itemOf(key, title, file string, waiting map[string]changeBy) Item {
	item := Item{Key: key, Title: title}
	if change, found := waiting[file]; found {
		item.Waiting = true
		item.By = change.By
		item.At = change.At
	}
	return item
}

// changeBy is who last changed a file and when, for the mark on an item.
type changeBy struct {
	By string
	At time.Time
}

func waitingFrom(changes []gitsync.Change) map[string]changeBy {
	waiting := map[string]changeBy{}
	for _, change := range changes {
		waiting[change.Path] = changeBy{By: change.By.Name, At: change.At}
	}
	return waiting
}
