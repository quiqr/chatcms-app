//go:build integration

package onboarding_test

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/quiqr/chatcms-app/internal/gitsync"
	"github.com/quiqr/chatcms-app/internal/httpapi"
	"github.com/quiqr/chatcms-app/internal/onboarding"
	"github.com/quiqr/chatcms-app/internal/policy"
	"github.com/quiqr/chatcms-app/internal/tenancy"
)

func onTheRemote(t *testing.T, h *harness, remote string) (subjects []string, files string, head string) {
	t.Helper()

	repository, err := gitsync.Clone(t.Context(), h.gitCfg, remote, filepath.Join(t.TempDir(), "look"))
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()

	versions, err := repository.History(t.Context(), 10)
	if err != nil {
		t.Fatal(err)
	}
	for _, version := range versions {
		subjects = append(subjects, version.Message)
	}
	var names []string
	fs.WalkDir(repository.FS(), ".", func(name string, entry fs.DirEntry, _ error) error {
		if entry != nil && !entry.IsDir() {
			names = append(names, name)
		}
		return nil
	})
	if len(versions) > 0 {
		head = versions[0].ID
	}
	return subjects, strings.Join(names, "\n"), head
}

func templateCatalogue(t *testing.T, h *harness) *onboarding.Catalogue {
	t.Helper()

	repository := remoteHolding(t, h.gitCfg, aSite(t))
	seed := filepath.Join(t.TempDir(), "seed")
	gitRun(t, h.gitCfg, "", "clone", "--", repository, seed)
	write(t, filepath.Join(seed, "content", "about.md"), "---\ntitle: About\n---\n")
	gitRun(t, h.gitCfg, seed, "add", "--all")
	gitRun(t, h.gitCfg, seed, "commit", "--message=A second version of the template")
	gitRun(t, h.gitCfg, seed, "push", "origin", "HEAD:refs/heads/main")
	gitRun(t, h.gitCfg, repository, "update-server-info")

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.Handle("/bakery.git/", http.StripPrefix("/bakery.git/", http.FileServer(http.Dir(repository))))
	mux.HandleFunc("/templates.json", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, `[{"QuiqrEtalageName": "Bakery", "NormalizedName": "Bakery",
			"HugoVersion": "extended_0.101.0", "SourceLink": "%s/bakery.git"}]`, server.URL)
	})

	catalogue := onboarding.NewCatalogue(h.db, server.URL+"/templates.json", h.stateDir, nil)
	if err := catalogue.Refresh(t.Context()); err != nil {
		t.Fatal(err)
	}
	h.importer.UseTemplates(catalogue)
	return catalogue
}

func (h *harness) projectWithKey(t *testing.T, slug string) tenancy.Project {
	t.Helper()

	project := h.project(t, slug)
	if _, err := h.keys.Generate(t.Context(), project.ID, "chatcms"); err != nil {
		t.Fatal(err)
	}
	return project
}

func TestStartingFromATemplateCopiesItIntoAnEmptyRepository(t *testing.T) {
	h := newHarness(t)
	templateCatalogue(t, h)
	project := h.projectWithKey(t, "bakery")
	remote := bareRemote(t, h.gitCfg)

	result, err := h.importer.Import(t.Context(), project.ID, onboarding.Request{
		Kind: onboarding.SourceTemplate, Address: "Bakery", Remote: remote,
	})
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if !result.Inspection.Ready() || result.Remote != remote {
		t.Errorf("result = %+v", result)
	}

	subjects, files, _ := onTheRemote(t, h, remote)
	if len(subjects) != 1 || strings.TrimSpace(subjects[0]) != "Start from Bakery" {
		t.Errorf("the remote holds %q, want one version named after the template", subjects)
	}
	for _, want := range []string{"hugo.toml", "content/about.md", "quiqr/model/base.yaml"} {
		if !strings.Contains(files, want) {
			t.Errorf("the remote lacks %s:\n%s", want, files)
		}
	}

	repository, err := h.sites.Open(t.Context(), project.ID.String())
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	if _, err := fs.ReadFile(repository.FS(), "content/about.md"); err != nil {
		t.Errorf("the working copy lacks the template's files: %v", err)
	}

	row, err := h.db.Queries.GetProject(t.Context(), pgtype.UUID{Bytes: project.ID, Valid: true})
	if err != nil {
		t.Fatal(err)
	}
	if row.HugoVersion != "0.101.0" || !row.HugoExtended {
		t.Errorf("hugo = %q extended %v, want the template's 0.101.0 extended", row.HugoVersion, row.HugoExtended)
	}
}

func TestATemplateIsNotCopiedIntoARepositoryThatHoldsWork(t *testing.T) {
	h := newHarness(t)
	templateCatalogue(t, h)
	project := h.projectWithKey(t, "bakery")
	remote := remoteHolding(t, h.gitCfg, aSite(t))
	_, _, before := onTheRemote(t, h, remote)

	_, err := h.importer.Import(t.Context(), project.ID, onboarding.Request{
		Kind: onboarding.SourceTemplate, Address: "Bakery", Remote: remote,
	})
	if !errors.Is(err, onboarding.ErrRepositoryNotEmpty) {
		t.Fatalf("Import() error = %v, want ErrRepositoryNotEmpty", err)
	}
	if _, _, after := onTheRemote(t, h, remote); after != before {
		t.Errorf("the remote moved from %s to %s", before, after)
	}
}

func TestATemplateThatIsNotInTheList(t *testing.T) {
	h := newHarness(t)
	templateCatalogue(t, h)
	project := h.projectWithKey(t, "bakery")

	_, err := h.importer.Import(t.Context(), project.ID, onboarding.Request{
		Kind: onboarding.SourceTemplate, Address: "No-Such", Remote: bareRemote(t, h.gitCfg),
	})
	if !errors.Is(err, onboarding.ErrNoSuchTemplate) {
		t.Errorf("Import() error = %v, want ErrNoSuchTemplate", err)
	}
}

func TestWithoutAGalleryNoTemplateCanBeChosen(t *testing.T) {
	h := newHarness(t)
	project := h.projectWithKey(t, "bakery")

	_, err := h.importer.Import(t.Context(), project.ID, onboarding.Request{
		Kind: onboarding.SourceTemplate, Address: "Bakery", Remote: bareRemote(t, h.gitCfg),
	})
	if !errors.Is(err, onboarding.ErrGalleryOff) {
		t.Errorf("Import() error = %v, want ErrGalleryOff", err)
	}
}

func TestOnlyThoseWhoMayCreateProjectsSeeTheGallery(t *testing.T) {
	h := newHarness(t)
	catalogue := templateCatalogue(t, h)
	service, _, actors := serviceFor(t, h)
	service.UseTemplates(catalogue, onboarding.Hugo{Version: "0.139.0", Extended: true})
	owner := httpapi.Actor{ID: h.owner.ID.String(), Email: h.owner.Email}

	view, err := service.ListTemplates(t.Context(), owner, "technative")
	if err != nil {
		t.Fatalf("the organisation's admin was refused: %v", err)
	}
	if !view.Offered || !view.Available || len(view.Templates) != 1 || view.Hugo.Version != "0.139.0" {
		t.Errorf("view = %+v", view)
	}
	if got := view.Templates[0].ScreenshotURL; got != "" {
		t.Errorf("a template without a screenshot type got %q", got)
	}

	for _, role := range []policy.Role{policy.RoleImplementer, policy.RoleEditor, policy.RoleViewer} {
		if _, err := service.ListTemplates(t.Context(), actors[role], "technative"); !errors.Is(err, httpapi.ErrNotAllowed) {
			t.Errorf("a project %s got %v, want ErrNotAllowed", role, err)
		}
		if _, _, err := service.TemplateScreenshot(t.Context(), actors[role], "technative", "Bakery"); !errors.Is(err, httpapi.ErrNotAllowed) {
			t.Errorf("a project %s got the screenshot: %v", role, err)
		}
	}

	if _, err := service.ListTemplates(t.Context(), owner, "no-such-organisation"); !errors.Is(err, httpapi.ErrMissing) {
		t.Errorf("an unknown organisation gave %v, want ErrMissing", err)
	}
}

func TestTheGallerySaysWhenItIsOffOrUnavailable(t *testing.T) {
	h := newHarness(t)
	service, _, _ := serviceFor(t, h)
	owner := httpapi.Actor{ID: h.owner.ID.String(), Email: h.owner.Email}

	service.UseTemplates(onboarding.NewCatalogue(h.db, "", h.stateDir, nil), onboarding.Hugo{})
	off, err := service.ListTemplates(t.Context(), owner, "technative")
	if err != nil || off.Offered {
		t.Errorf("a gallery that is off: %+v, %v", off, err)
	}

	unreachable := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(unreachable.Close)
	service.UseTemplates(onboarding.NewCatalogue(h.db, unreachable.URL+"/templates.json", h.stateDir, nil), onboarding.Hugo{})
	unavailable, err := service.ListTemplates(t.Context(), owner, "technative")
	if err != nil || !unavailable.Offered || unavailable.Available || unavailable.Templates == nil {
		t.Errorf("a gallery never fetched: %+v, %v", unavailable, err)
	}
}

func TestTemplateRefusalsCarryTheirKeys(t *testing.T) {
	h := newHarness(t)
	templateCatalogue(t, h)
	service, project, actors := serviceFor(t, h)
	if _, err := h.keys.Generate(t.Context(), project.ID, "chatcms"); err != nil {
		t.Fatal(err)
	}

	_, err := service.ImportSite(t.Context(), actors[policy.RoleImplementer], "technative", "korenaar",
		"template", "Bakery", remoteHolding(t, h.gitCfg, aSite(t)))
	var refused *httpapi.RefusedError
	if !errors.As(err, &refused) || refused.Key != "refused.repository-not-empty" {
		t.Errorf("ImportSite() error = %v, want the not-empty refusal", err)
	}
}

func TestAProjectsHugoVersionCanBeChanged(t *testing.T) {
	h := newHarness(t)
	service, _, actors := serviceFor(t, h)
	service.UseTemplates(nil, onboarding.Hugo{Version: "0.139.0", Extended: true})
	implementer := actors[policy.RoleImplementer]

	view, err := service.SetHugo(t.Context(), implementer, "technative", "korenaar", "v0.101.0", true)
	if err != nil {
		t.Fatal(err)
	}
	if view.Version != "0.101.0" || !view.Extended || view.Installed.Version != "0.139.0" {
		t.Errorf("view = %+v", view)
	}

	read, err := service.ReadHugo(t.Context(), implementer, "technative", "korenaar")
	if err != nil || read.Version != "0.101.0" {
		t.Errorf("read back %+v, %v", read, err)
	}

	var refused *httpapi.RefusedError
	if _, err := service.SetHugo(t.Context(), implementer, "technative", "korenaar", "latest", false); !errors.As(err, &refused) {
		t.Errorf("a version that is not numbers gave %v", err)
	}
	if _, err := service.SetHugo(t.Context(), actors[policy.RoleEditor], "technative", "korenaar", "0.1.0", false); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("an editor changed the Hugo version: %v", err)
	}
}
