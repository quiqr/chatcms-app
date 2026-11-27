package content

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/httpapi"
	"github.com/quiqr/chatcms-app/internal/policy"
	"github.com/quiqr/chatcms-app/internal/tenancy"
)

// Web answers the HTTP layer's questions, turning slugs into a project and a
// signed-in person into an actor.
type Web struct {
	service *Service
	tenancy *tenancy.Store
}

func NewWeb(service *Service, tenancyStore *tenancy.Store) (*Web, error) {
	if service == nil || tenancyStore == nil {
		return nil, ErrIncomplete
	}
	return &Web{service: service, tenancy: tenancyStore}, nil
}

func (w *Web) reach(ctx context.Context, who httpapi.Actor, organisation, project string) (policy.Actor, uuid.UUID, error) {
	if who.Anonymous || who.ID == "" {
		return policy.Actor{}, uuid.Nil, httpapi.ErrNotAllowed
	}
	personID, err := uuid.Parse(who.ID)
	if err != nil {
		return policy.Actor{}, uuid.Nil, httpapi.ErrNotAllowed
	}

	actor, err := w.tenancy.ActorFor(ctx, personID)
	if err != nil {
		return policy.Actor{}, uuid.Nil, err
	}

	found, err := w.tenancy.ProjectBySlug(ctx, organisation, project)
	if errors.Is(err, tenancy.ErrNotFound) {
		return policy.Actor{}, uuid.Nil, httpapi.ErrMissing
	}
	if err != nil {
		return policy.Actor{}, uuid.Nil, err
	}
	return actor, found.ID, nil
}

func (w *Web) ReadPage(ctx context.Context, who httpapi.Actor, organisation, project, page string) (httpapi.PageView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.PageView{}, err
	}

	editable, err := w.service.Read(ctx, actor, id, page)
	if err != nil {
		return httpapi.PageView{}, noModelIsRefusal(err)
	}
	return pageView(editable), nil
}

func (w *Web) SavePage(ctx context.Context, who httpapi.Actor, organisation, project, page string, edit httpapi.EditRequest) (httpapi.PageView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.PageView{}, err
	}

	saved, err := w.service.Save(ctx, actor, id, page, Edit{Values: edit.Values, Body: edit.Body})
	if err != nil {
		return httpapi.PageView{}, noModelIsRefusal(err)
	}
	return pageView(saved), nil
}

// noModelIsRefusal turns a site with no model into something worth showing
// rather than a failure nobody can act on.
func noModelIsRefusal(err error) error {
	if errors.Is(err, ErrNoModel) {
		return &httpapi.RefusedError{
			Key:    "refused.site-has-no-quiqr-model",
			Reason: "this site has no Quiqr model yet, so there are no forms to fill in",
		}
	}
	return err
}

func pageView(editable Editable) httpapi.PageView {
	return httpapi.PageView{
		Key:        editable.Key,
		Title:      editable.Title,
		Section:    editable.Section,
		Fields:     controlViews(editable.Controls),
		Body:       editable.Body,
		HasBody:    editable.HasBody,
		Dialect:    editable.Dialect,
		Shortcodes: shortcodeViews(editable.Shortcodes),
		Preview:    httpapi.PagePreview{Kind: editable.Preview.Kind, Address: editable.Preview.Address},
	}
}

func shortcodeViews(shapes []ShortcodeShape) []httpapi.ShortcodeView {
	if len(shapes) == 0 {
		return nil
	}
	views := make([]httpapi.ShortcodeView, 0, len(shapes))
	for _, shape := range shapes {
		views = append(views, httpapi.ShortcodeView{
			Key:       shape.Key,
			Title:     shape.Title,
			Inner:     shape.Inner,
			Fields:    controlViews(shape.Controls),
			Positions: shape.Positions,
		})
	}
	return views
}

func controlViews(controls []Control) []httpapi.ControlView {
	views := make([]httpapi.ControlView, 0, len(controls))
	for _, control := range controls {
		view := httpapi.ControlView{
			Path:    control.Path,
			Type:    control.Type,
			Title:   control.Title,
			Tip:     control.Tip,
			Value:   control.Value,
			Config:  control.Config,
			Unknown: control.Unknown,
		}
		for _, option := range control.Options {
			view.Options = append(view.Options, httpapi.OptionView{Value: option.Value, Label: option.Label})
		}
		if len(control.Controls) > 0 {
			view.Fields = controlViews(control.Controls)
		}
		views = append(views, view)
	}
	return views
}

var _ httpapi.ContentService = (*Web)(nil)

func (w *Web) ListContent(ctx context.Context, who httpapi.Actor, organisation, project string) ([]httpapi.SectionView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return nil, err
	}

	sections, err := w.service.List(ctx, actor, id)
	if err != nil {
		return nil, noModelIsRefusal(err)
	}

	views := make([]httpapi.SectionView, 0, len(sections))
	for _, section := range sections {
		view := httpapi.SectionView{
			Key:      section.Key,
			Title:    section.Title,
			ItemName: section.ItemName,
			Count:    section.Count,
			Items:    make([]httpapi.ItemView, 0, len(section.Items)),
		}
		for _, item := range section.Items {
			shown := httpapi.ItemView{
				Key:     item.Key,
				Title:   item.Title,
				Waiting: item.Waiting,
				By:      item.By,
			}
			if !item.At.IsZero() {
				shown.At = item.At.Format(time.RFC3339)
			}
			view.Items = append(view.Items, shown)
		}
		views = append(views, view)
	}
	return views, nil
}

func (w *Web) CreatePage(ctx context.Context, who httpapi.Actor, organisation, project, section, title string) (httpapi.PageView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.PageView{}, err
	}

	made, err := w.service.Create(ctx, actor, id, NewPage{Section: section, Title: title})
	if err != nil {
		return httpapi.PageView{}, noModelIsRefusal(err)
	}
	return pageView(made), nil
}

func (w *Web) RenamePage(ctx context.Context, who httpapi.Actor, organisation, project, page, title string) (httpapi.PageView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.PageView{}, err
	}

	renamed, err := w.service.Rename(ctx, actor, id, page, title)
	if err != nil {
		return httpapi.PageView{}, noModelIsRefusal(err)
	}
	return pageView(renamed), nil
}

func (w *Web) DuplicatePage(ctx context.Context, who httpapi.Actor, organisation, project, page, title string) (httpapi.PageView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.PageView{}, err
	}

	copied, err := w.service.Duplicate(ctx, actor, id, page, title)
	if err != nil {
		return httpapi.PageView{}, noModelIsRefusal(err)
	}
	return pageView(copied), nil
}

func (w *Web) DeletePage(ctx context.Context, who httpapi.Actor, organisation, project, page string) error {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return err
	}
	return noModelIsRefusal(w.service.Delete(ctx, actor, id, page))
}

var _ httpapi.BrowseService = (*Web)(nil)

func (w *Web) ContentMenu(ctx context.Context, who httpapi.Actor, organisation, project string) (httpapi.ContentMenuView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.ContentMenuView{}, err
	}
	menu, err := w.service.Menu(ctx, actor, id)
	if err != nil {
		return httpapi.ContentMenuView{}, noModelIsRefusal(err)
	}

	view := httpapi.ContentMenuView{
		Groups:   make([]httpapi.MenuGroupView, 0, len(menu.Groups)),
		Totals:   httpapi.MenuTotalsView{Singles: menu.Singles, Collections: menu.Collections},
		Fallback: menu.Fallback,
	}
	for _, group := range menu.Groups {
		shown := httpapi.MenuGroupView{Key: group.Key, Title: group.Title, Items: make([]httpapi.MenuEntryView, 0, len(group.Entries))}
		for _, entry := range group.Entries {
			item := httpapi.MenuEntryView{
				Key: entry.Key, Kind: entry.Kind, Title: entry.Title,
				ItemName: entry.ItemName, HasChanges: entry.HasChanges,
			}
			if entry.Kind == "collection" {
				count := entry.Count
				item.Count = &count
			}
			shown.Items = append(shown.Items, item)
		}
		view.Groups = append(view.Groups, shown)
	}
	if menu.Technical {
		for _, warning := range menu.Warnings {
			view.Warnings = append(view.Warnings, warning.Message)
		}
	}
	return view, nil
}

func (w *Web) CollectionItems(ctx context.Context, who httpapi.Actor, organisation, project, collection string, query httpapi.ItemsQuery) (httpapi.CollectionItemsView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.CollectionItemsView{}, err
	}
	listing, err := w.service.Items(ctx, actor, id, collection, Query{
		Q: query.Q, Folder: query.Folder, Sort: query.Sort,
		Page: query.Page, PageSize: query.PageSize, Around: query.Around,
	})
	if err != nil {
		return httpapi.CollectionItemsView{}, noModelIsRefusal(err)
	}

	view := httpapi.CollectionItemsView{
		Key:        listing.Collection.Key,
		Title:      titleOr(listing.Collection.Title, listing.Collection.Key),
		ItemName:   itemNameOf(listing.Collection),
		Items:      make([]httpapi.CollectionItemView, 0, len(listing.Items)),
		Total:      listing.Total,
		All:        listing.All,
		Sort:       listing.Sort,
		Folders:    make([]httpapi.CollectionFolderView, 0, len(listing.Folders)),
		ShowsPaths: listing.ShowsPaths,
		HasImages:  listing.Collection.HasThumbnails(),
	}
	if listing.Around != nil {
		view.Around = &httpapi.AroundView{
			Index: listing.Around.Index, Previous: listing.Around.Previous, Next: listing.Around.Next,
		}
	}
	for _, item := range listing.Items {
		shown := httpapi.CollectionItemView{
			Key: item.Key, Title: item.Title, Folder: item.Folder, Status: item.Status, By: item.By,
			Date: item.Date, Author: item.Author, Summary: item.Summary, Thumbnail: item.Thumbnail,
		}
		if !item.Edited.IsZero() {
			shown.Edited = item.Edited.Format(time.RFC3339)
		}
		if listing.ShowsPaths {
			shown.Path = item.File
		}
		view.Items = append(view.Items, shown)
	}
	for _, folder := range listing.Folders {
		view.Folders = append(view.Folders, httpapi.CollectionFolderView{Name: folder.Name, Label: folder.Name, Count: folder.Count})
	}
	return view, nil
}

func (w *Web) CreateItem(ctx context.Context, who httpapi.Actor, organisation, project, collection string, asked httpapi.NewItemRequest) (httpapi.PageView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.PageView{}, err
	}
	made, err := w.service.Create(ctx, actor, id, NewPage{
		Section: collection, Title: asked.Title, Folder: asked.Folder, Slug: asked.Slug,
	})
	var refused *httpapi.RefusedError
	if errors.As(err, &refused) && refused.Key == "refused.page-name-taken" {
		return httpapi.PageView{}, &httpapi.RefusedError{
			Key:    "refused.item-address-taken",
			Reason: "there is already an item at this address; change the title or the address",
		}
	}
	if err != nil {
		return httpapi.PageView{}, noModelIsRefusal(err)
	}
	return pageView(made), nil
}

func (w *Web) ItemAddress(ctx context.Context, who httpapi.Actor, organisation, project, collection string, asked httpapi.NewItemRequest) (httpapi.AddressView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.AddressView{}, err
	}
	address, err := w.service.Address(ctx, actor, id, NewPage{
		Section: collection, Title: asked.Title, Folder: asked.Folder, Slug: asked.Slug,
	})
	if err != nil {
		return httpapi.AddressView{}, noModelIsRefusal(err)
	}
	return httpapi.AddressView{
		Slug: address.Slug, Folder: address.Folder, Address: address.Address, Free: address.Free, File: address.File,
	}, nil
}

func (w *Web) Search(ctx context.Context, who httpapi.Actor, organisation, project, q string) (httpapi.SearchView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.SearchView{}, err
	}
	groups, err := w.service.Search(ctx, actor, id, q)
	if err != nil {
		return httpapi.SearchView{}, noModelIsRefusal(err)
	}
	view := httpapi.SearchView{Groups: make([]httpapi.SearchGroupView, 0, len(groups))}
	for _, group := range groups {
		shown := httpapi.SearchGroupView{Title: group.Title, Items: make([]httpapi.SearchItemView, 0, len(group.Items))}
		for _, found := range group.Items {
			shown.Items = append(shown.Items, httpapi.SearchItemView{Key: found.Key, Title: found.Title, Kind: found.Kind})
		}
		view.Groups = append(view.Groups, shown)
	}
	return view, nil
}

func (w *Web) ProjectFiles(ctx context.Context, who httpapi.Actor, organisation, project string) (httpapi.FilesView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.FilesView{}, err
	}
	entries, err := w.service.Files(ctx, actor, id)
	if err != nil {
		return httpapi.FilesView{}, err
	}
	view := httpapi.FilesView{Entries: make([]httpapi.FileEntryView, 0, len(entries))}
	for _, entry := range entries {
		view.Entries = append(view.Entries, httpapi.FileEntryView{Path: entry.Path, Dir: entry.Dir, Link: entry.Link, Size: entry.Size})
	}
	return view, nil
}

func fileView(text FileText) httpapi.FileView {
	return httpapi.FileView{Path: text.Path, Text: text.Text, Sum: text.Sum, Unusable: text.Unusable, Preview: text.Preview}
}

func (w *Web) OpenProjectFile(ctx context.Context, who httpapi.Actor, organisation, project, name string) (httpapi.FileContent, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.FileContent{}, err
	}
	file, err := w.service.OpenFile(ctx, actor, id, name)
	if err != nil {
		return httpapi.FileContent{}, err
	}
	return httpapi.FileContent{Name: file.Name, Preview: file.Preview, Type: file.Type, Modified: file.Modified, Content: file.Content, Close: file.Close}, nil
}

func (w *Web) ProjectFile(ctx context.Context, who httpapi.Actor, organisation, project, name string) (httpapi.FileView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.FileView{}, err
	}
	text, err := w.service.ReadFile(ctx, actor, id, name)
	if err != nil {
		return httpapi.FileView{}, err
	}
	return fileView(text), nil
}

func (w *Web) SaveProjectFile(ctx context.Context, who httpapi.Actor, organisation, project, name string, save httpapi.FileSave) (httpapi.FileView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.FileView{}, err
	}
	text, err := w.service.WriteFile(ctx, actor, id, name, save.Text, save.Sum)
	if err != nil {
		return httpapi.FileView{}, err
	}
	return fileView(text), nil
}

var _ httpapi.FilesService = (*Web)(nil)
