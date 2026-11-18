package app

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/gitsync"
)

func copiesSites(t *testing.T) (*gitsync.Sites, uuid.UUID) {
	t.Helper()

	bin := os.Getenv("CHATCMS_GIT_BIN")
	if bin == "" {
		found, err := exec.LookPath("git")
		if err != nil {
			t.Skipf("git is not available: %v", err)
		}
		bin = found
	}
	cfg := gitsync.Config{GitBin: bin, System: gitsync.Identity{Name: "Quiqr ChatCMS", Email: "chatcms@example.org"}}
	project := uuid.New()
	sites := gitsync.NewSites(cfg, filepath.Join(t.TempDir(), "sites"), nil)
	repository, err := sites.Create(t.Context(), project.String(), "", os.DirFS(seedDir(t)), gitsync.Identity{Name: "Seed", Email: "seed@example.org"}, "First version")
	if err != nil {
		t.Fatal(err)
	}
	repository.Close()
	return sites, project
}

func seedDir(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "content"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "content", "_index.md"), []byte("---\ntitle: Home\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func saveAs(t *testing.T, sites *gitsync.Sites, project uuid.UUID, person, name, body string) {
	t.Helper()

	repository, err := openFor(t.Context(), sites, person, project)
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	if _, err := repository.Save(t.Context(), gitsync.Save{
		Message: "Edit " + name,
		Author:  gitsync.Identity{Name: person, Email: person + "@example.org"},
		Writes:  map[string][]byte{name: []byte(body)},
	}); err != nil {
		t.Fatal(err)
	}
}

func standingFor(t *testing.T, sites *gitsync.Sites, project uuid.UUID, person string) standingCopy {
	t.Helper()

	site, done, err := syncSources{publishSources: publishSources{sites: sites}}.Site(t.Context(), person, project)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { done() })
	return site.(standingCopy)
}

func TestStandingCountsWhatIsNotShared(t *testing.T) {
	sites, project := copiesSites(t)
	saveAs(t, sites, project, "anna", "content/about.md", "anna\n")

	status, err := standingFor(t, sites, project, "anna").Status(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if status.Unpublished < 2 {
		t.Errorf("Unpublished = %d, want the unshared save and the unpublished first version", status.Unpublished)
	}
}

func TestStandingBringsInWhatOthersSharedWhileOffline(t *testing.T) {
	sites, project := copiesSites(t)
	standingFor(t, sites, project, "anna")
	saveAs(t, sites, project, "bert", "content/contact.md", "bert\n")
	if err := (publishSources{sites: sites}).Share(t.Context(), "bert", project); err != nil {
		t.Fatal(err)
	}

	anna := standingFor(t, sites, project, "anna")
	result, err := anna.Sync(t.Context())
	if !errors.Is(err, gitsync.ErrRemoteUnreachable) {
		t.Fatalf("Sync() error = %v, want the missing remote reported", err)
	}
	if result.Conflict != nil {
		t.Fatalf("Sync() conflict = %v, want none", result.Conflict)
	}
	if data, err := fs.ReadFile(anna.FS(), "content/contact.md"); err != nil || string(data) != "bert\n" {
		t.Errorf("anna's copy holds %q, %v, want bert's shared save", data, err)
	}
}

func TestAnsweringSettlesThePersonsCopyAndShares(t *testing.T) {
	sites, project := copiesSites(t)
	saveAs(t, sites, project, "anna", "content/_index.md", "---\ntitle: Anna\n---\n")
	saveAs(t, sites, project, "bert", "content/_index.md", "---\ntitle: Bert\n---\n")
	shares := publishSources{sites: sites}
	if err := shares.Share(t.Context(), "anna", project); err != nil {
		t.Fatal(err)
	}
	if err := shares.Share(t.Context(), "bert", project); err != errAnswerFirst {
		t.Fatalf("Share() error = %v, want errAnswerFirst", err)
	}

	bert := standingFor(t, sites, project, "bert")
	status, err := bert.Status(t.Context())
	if err != nil || !status.ExchangeInProgress {
		t.Fatalf("Status() = %+v, %v, want a question waiting", status, err)
	}

	result, err := bert.Resolve(t.Context(), map[string][]byte{"content/_index.md": []byte("---\ntitle: Both\n---\n")})
	if err != nil || result.Conflict != nil {
		t.Fatalf("Resolve() = %+v, %v, want it settled", result, err)
	}

	shared, err := sites.Open(t.Context(), project.String())
	if err != nil {
		t.Fatal(err)
	}
	defer shared.Close()
	if data, _ := fs.ReadFile(shared.FS(), "content/_index.md"); string(data) != "---\ntitle: Both\n---\n" {
		t.Errorf("shared version holds %q, want the answer shared", data)
	}
}

func TestGivingUpRestoresThePersonsCopy(t *testing.T) {
	sites, project := copiesSites(t)
	saveAs(t, sites, project, "anna", "content/_index.md", "---\ntitle: Anna\n---\n")
	saveAs(t, sites, project, "bert", "content/_index.md", "---\ntitle: Bert\n---\n")
	shares := publishSources{sites: sites}
	_ = shares.Share(t.Context(), "anna", project)
	_ = shares.Share(t.Context(), "bert", project)

	if err := standingFor(t, sites, project, "bert").Abandon(t.Context()); err != nil {
		t.Fatalf("Abandon() error = %v", err)
	}
	status, err := standingFor(t, sites, project, "bert").Status(t.Context())
	if err != nil || status.ExchangeInProgress {
		t.Errorf("Status() = %+v, %v, want no question left", status, err)
	}
}

func TestShareNeedsASite(t *testing.T) {
	sites, _ := copiesSites(t)
	if err := (publishSources{sites: sites}).Share(t.Context(), "anna", uuid.New()); !errors.Is(err, errNothingBroughtIn) {
		t.Errorf("Share() error = %v, want errNothingBroughtIn", err)
	}
}

func TestOpenForRefusesSomebodyWithoutAName(t *testing.T) {
	sites, project := copiesSites(t)
	if _, err := openFor(t.Context(), sites, "", project); err == nil {
		t.Errorf("openFor() with no person succeeded, want it refused")
	}
}

func TestShareSubjectRoundTrips(t *testing.T) {
	project := uuid.New()
	gotProject, gotPerson, ok := parseShareSubject(shareSubject("anna", project))
	if !ok || gotProject != project.String() || gotPerson != "anna" {
		t.Errorf("parseShareSubject() = %q, %q, %v", gotProject, gotPerson, ok)
	}
	for _, bad := range []string{"", "x", "/anna", "x/"} {
		if _, _, ok := parseShareSubject(bad); ok {
			t.Errorf("parseShareSubject(%q) accepted", bad)
		}
	}
}

func TestAnsweringInALaterRequestFinishesTheExchange(t *testing.T) {
	sites, project := copiesSites(t)
	saveAs(t, sites, project, "anna", "content/_index.md", "---\ntitle: Anna\n---\n")
	saveAs(t, sites, project, "bert", "content/_index.md", "---\ntitle: Bert\n---\n")
	shares := publishSources{sites: sites}
	_ = shares.Share(t.Context(), "anna", project)
	_ = shares.Share(t.Context(), "bert", project)

	later := standingFor(t, sites, project, "bert")
	asked, err := later.Sync(t.Context())
	if err != nil || asked.Conflict == nil || asked.Conflict.Paths[0].Path != "content/_index.md" {
		t.Fatalf("Sync() in a later request = %+v, %v, want the open question", asked, err)
	}

	finished, err := later.Resolve(t.Context(), map[string][]byte{"content/_index.md": asked.Conflict.Paths[0].Local})
	if err != nil || finished.Conflict != nil {
		t.Fatalf("Resolve() = %+v, %v, want it finished", finished, err)
	}

	shared, err := sites.Open(t.Context(), project.String())
	if err != nil {
		t.Fatal(err)
	}
	defer shared.Close()
	if data, _ := fs.ReadFile(shared.FS(), "content/_index.md"); string(data) != "---\ntitle: Bert\n---\n" {
		t.Errorf("shared version holds %q, want bert's kept answer", data)
	}
	if status, _ := standingFor(t, sites, project, "bert").Status(t.Context()); status.ExchangeInProgress {
		t.Errorf("a question is still open after answering")
	}
}
