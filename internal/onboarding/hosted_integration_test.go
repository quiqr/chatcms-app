//go:build integration

package onboarding_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/auth"
	"github.com/quiqr/chatcms-app/internal/gitsync"
	"github.com/quiqr/chatcms-app/internal/httpapi"
	"github.com/quiqr/chatcms-app/internal/onboarding"
	"github.com/quiqr/chatcms-app/internal/policy"
	"github.com/quiqr/chatcms-app/internal/tenancy"
)

func keptPath(h *harness, projectID uuid.UUID) string {
	return filepath.Join(h.stateDir, "hosted", projectID.String()+".git")
}

func hosted(t *testing.T, h *harness, projectID uuid.UUID) bool {
	t.Helper()
	row, err := h.db.Queries.GetProject(t.Context(), pgUUID(projectID))
	if err != nil {
		t.Fatal(err)
	}
	return row.Hosted
}

func keptTemplate(t *testing.T, h *harness, projectID uuid.UUID) {
	t.Helper()
	result, err := h.importer.Import(t.Context(), projectID, onboarding.Request{
		Kind: onboarding.SourceTemplate, Address: "Bakery",
	})
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if !result.Hosted || result.Remote != "" || !result.Inspection.Ready() {
		t.Fatalf("result = %+v, want a ready site kept by ChatCMS", result)
	}
}

func TestATemplateWithoutAnAddressIsKept(t *testing.T) {
	h := newHarness(t)
	templateCatalogue(t, h)
	project := h.project(t, "bakery")

	keptTemplate(t, h, project.ID)

	if has, _ := h.keys.Has(t.Context(), project.ID); has {
		t.Error("a key was made, want none needed")
	}
	if !hosted(t, h, project.ID) {
		t.Error("hosted = false, want the project to say it is kept")
	}
	subjects, files, _ := onTheRemote(t, h, keptPath(h, project.ID))
	if len(subjects) != 1 || strings.TrimSpace(subjects[0]) != "Start from Bakery" {
		t.Errorf("the kept repository holds %q, want one version named after the template", subjects)
	}
	if !strings.Contains(files, "content/about.md") {
		t.Errorf("the kept repository lacks the template's files:\n%s", files)
	}
}

func TestAnExportWithoutAnAddressIsKept(t *testing.T) {
	h := newHarness(t)
	project := h.project(t, "korenaar")

	result, err := h.importer.Import(t.Context(), project.ID, onboarding.Request{
		Kind: onboarding.SourceExport, Address: aSite(t),
	})
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if !result.Hosted || !result.Inspection.Ready() {
		t.Errorf("result = %+v, want a ready site kept by ChatCMS", result)
	}
	if !hosted(t, h, project.ID) {
		t.Error("hosted = false, want the project to say it is kept")
	}
	if subjects, _, _ := onTheRemote(t, h, keptPath(h, project.ID)); len(subjects) != 1 {
		t.Errorf("the kept repository holds %q, want the one version", subjects)
	}
}

func TestAnUnreadableExportKeepsNothing(t *testing.T) {
	h := newHarness(t)
	project := h.project(t, "korenaar")

	_, err := h.importer.Import(t.Context(), project.ID, onboarding.Request{
		Kind: onboarding.SourceExport, Address: filepath.Join(t.TempDir(), "absent"),
	})
	if !errors.Is(err, onboarding.ErrExportUnreadable) {
		t.Fatalf("Import() error = %v, want ErrExportUnreadable", err)
	}
	if _, err := os.Stat(keptPath(h, project.ID)); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("a kept repository was left behind: %v", err)
	}
}

func TestARepositoryImportNeedsAnAddress(t *testing.T) {
	h := newHarness(t)
	project := h.project(t, "korenaar")

	_, err := h.importer.Import(t.Context(), project.ID, onboarding.Request{Kind: onboarding.SourceRepository})
	if !errors.Is(err, onboarding.ErrNeedsAnAddress) {
		t.Errorf("Import() error = %v, want ErrNeedsAnAddress", err)
	}
}

func TestConnectingAKeptSiteToAnEmptyRepository(t *testing.T) {
	h := newHarness(t)
	templateCatalogue(t, h)
	project := h.project(t, "bakery")
	keptTemplate(t, h, project.ID)

	repository, err := h.sites.Open(t.Context(), project.ID.String())
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"one", "two"} {
		if _, err := repository.Save(t.Context(), gitsync.Save{
			Message: "Edit " + name,
			Author:  gitsync.Identity{Name: "Anna", Email: "anna@example.org"},
			Writes:  map[string][]byte{"content/" + name + ".md": []byte(name + "\n")},
		}); err != nil {
			t.Fatal(err)
		}
	}
	repository.Close()

	remote := bareRemote(t, h.gitCfg)
	if err := h.importer.Connect(t.Context(), project.ID, remote); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	subjects, _, _ := onTheRemote(t, h, remote)
	if len(subjects) != 3 {
		t.Errorf("the repository holds %q, want three versions", subjects)
	}
	if hosted(t, h, project.ID) {
		t.Error("hosted = true after connecting, want false")
	}
	row, _ := h.db.Queries.GetProject(t.Context(), pgUUID(project.ID))
	if row.Remote != remote {
		t.Errorf("remote = %q, want %q", row.Remote, remote)
	}
	if _, err := os.Stat(keptPath(h, project.ID)); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the kept repository is still there: %v", err)
	}

	if err := h.importer.Connect(t.Context(), project.ID, bareRemote(t, h.gitCfg)); !errors.Is(err, onboarding.ErrNotKept) {
		t.Errorf("a second Connect() error = %v, want ErrNotKept", err)
	}
}

func TestConnectingRefusesARepositoryThatHoldsWork(t *testing.T) {
	h := newHarness(t)
	templateCatalogue(t, h)
	project := h.project(t, "bakery")
	keptTemplate(t, h, project.ID)
	remote := remoteHolding(t, h.gitCfg, aSite(t))
	_, _, before := onTheRemote(t, h, remote)

	if err := h.importer.Connect(t.Context(), project.ID, remote); !errors.Is(err, onboarding.ErrConnectNotEmpty) {
		t.Fatalf("Connect() error = %v, want ErrConnectNotEmpty", err)
	}
	if _, _, after := onTheRemote(t, h, remote); after != before {
		t.Errorf("the repository moved from %s to %s", before, after)
	}
	if !hosted(t, h, project.ID) {
		t.Error("hosted = false, want the site still kept")
	}
}

func TestConnectingRefusesARepositoryTheKeyDoesNotReach(t *testing.T) {
	h := newHarness(t)
	templateCatalogue(t, h)
	project := h.project(t, "bakery")
	keptTemplate(t, h, project.ID)

	err := h.importer.Connect(t.Context(), project.ID, filepath.Join(t.TempDir(), "absent.git"))
	if !errors.Is(err, onboarding.ErrConnectKeyRefused) && !errors.Is(err, onboarding.ErrConnectUnreachable) {
		t.Fatalf("Connect() error = %v, want it refused", err)
	}
	if !hosted(t, h, project.ID) {
		t.Error("hosted = false, want the site still kept")
	}
	if _, err := os.Stat(keptPath(h, project.ID)); err != nil {
		t.Errorf("the kept repository is gone: %v", err)
	}
	if err := h.importer.Connect(t.Context(), project.ID, " "); !errors.Is(err, onboarding.ErrConnectNeedsAnAddress) {
		t.Errorf("Connect(blank) error = %v, want ErrConnectNeedsAnAddress", err)
	}
}

func TestConnectingIsForThoseWhoMayChangeSettings(t *testing.T) {
	h := newHarness(t)
	templateCatalogue(t, h)
	service, project, actors := serviceFor(t, h)
	keptTemplate(t, h, project.ID)

	_, err := service.ConnectSite(t.Context(), actors[policy.RoleEditor], "technative", "korenaar", bareRemote(t, h.gitCfg))
	if !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("an editor connecting: error = %v, want ErrNotAllowed", err)
	}

	_, err = service.ConnectSite(t.Context(), actors[policy.RoleImplementer], "technative", "korenaar",
		remoteHolding(t, h.gitCfg, aSite(t)))
	var refused *httpapi.RefusedError
	if !errors.As(err, &refused) || refused.Key != "refused.connect-not-empty" {
		t.Errorf("connecting to a full repository: error = %v, want refused.connect-not-empty", err)
	}

	view, err := service.InspectSite(t.Context(), actors[policy.RoleImplementer], "technative", "korenaar")
	if err != nil || !view.Hosted {
		t.Errorf("InspectSite() = %+v, %v, want it kept", view, err)
	}
	connection, err := service.TestConnection(t.Context(), actors[policy.RoleImplementer], "technative", "korenaar")
	if err != nil || connection.Good || !strings.Contains(connection.Message, "kept in ChatCMS") {
		t.Errorf("TestConnection() = %+v, %v, want it to say the site is kept", connection, err)
	}

	remote := bareRemote(t, h.gitCfg)
	view, err = service.ConnectSite(t.Context(), actors[policy.RoleImplementer], "technative", "korenaar", remote)
	if err != nil {
		t.Fatalf("ConnectSite() error = %v", err)
	}
	if view.Hosted || view.Remote != remote {
		t.Errorf("view = %+v, want connected to %s", view, remote)
	}
}

func TestRemovingAKeptProjectRemovesWhatWasKept(t *testing.T) {
	h := newHarness(t)
	templateCatalogue(t, h)
	project := h.project(t, "bakery")
	keptTemplate(t, h, project.ID)

	service := tenancy.NewService(h.store, auth.NewService(h.db, quiet(), "https://chatcms.example"))
	service.ForgetsKeptSites(func(_ context.Context, projectID uuid.UUID) error {
		return h.sites.Unhost(projectID.String())
	})
	owner := httpapi.Actor{ID: h.owner.ID.String(), Email: h.owner.Email}
	if err := service.RemoveProject(t.Context(), owner, "technative", "bakery"); err != nil {
		t.Fatalf("RemoveProject() error = %v", err)
	}
	if _, err := os.Stat(keptPath(h, project.ID)); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the kept repository is still there: %v", err)
	}
}
