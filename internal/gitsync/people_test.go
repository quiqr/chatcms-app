package gitsync

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func sharedSite(t *testing.T) (*Sites, string, Config) {
	t.Helper()

	cfg := testConfig(t)
	remote := bareRemote(t, cfg)
	sites := NewSites(cfg, filepath.Join(t.TempDir(), "sites"), nil)
	repository, err := sites.Clone(t.Context(), "korenaar", remote)
	if err != nil {
		t.Fatalf("Clone() error = %v, want nil", err)
	}
	repository.Close()
	return sites, remote, cfg
}

func openFor(t *testing.T, sites *Sites, person string) Repo {
	t.Helper()

	repository, err := sites.OpenFor(t.Context(), "korenaar", person)
	if err != nil {
		t.Fatalf("OpenFor(%q) error = %v, want nil", person, err)
	}
	t.Cleanup(func() { repository.Close() })
	return repository
}

func readSite(t *testing.T, repository Repo, name string) string {
	t.Helper()

	data, err := fs.ReadFile(repository.FS(), name)
	if errors.Is(err, fs.ErrNotExist) {
		return ""
	}
	if err != nil {
		t.Fatalf("reading %s: %v", name, err)
	}
	return string(data)
}

func share(t *testing.T, sites *Sites, person string) SyncResult {
	t.Helper()

	result, err := sites.Share(t.Context(), "korenaar", person)
	if err != nil {
		t.Fatalf("Share(%q) error = %v, want nil", person, err)
	}
	return result
}

func TestOpenForMakesACopyOnce(t *testing.T) {
	sites, _, _ := sharedSite(t)

	first := openFor(t, sites, "anna")
	save(t, first, "Anna", "Edit about", map[string]string{"content/about.md": "anna\n"})

	again := openFor(t, sites, "anna")
	if got := readSite(t, again, "content/about.md"); got != "anna\n" {
		t.Errorf("reopened copy holds %q, want the unshared save kept", got)
	}

	shared, _ := sites.Dir("korenaar")
	mine, _ := sites.DirFor("korenaar", "anna")
	if shared == mine {
		t.Fatalf("a person's copy is the shared copy")
	}
	if !strings.HasPrefix(mine, filepath.Dir(shared)) {
		t.Errorf("DirFor() = %q, want it under the site's own directory", mine)
	}
}

func TestOpenForRefusesUnusableNames(t *testing.T) {
	sites, _, _ := sharedSite(t)

	for _, person := range []string{"", "..", "a/b"} {
		if _, err := sites.OpenFor(t.Context(), "korenaar", person); !errors.Is(err, ErrUnusableSiteID) {
			t.Errorf("OpenFor(%q) error = %v, want ErrUnusableSiteID", person, err)
		}
	}
}

func TestOpenForNeedsASharedCopy(t *testing.T) {
	sites := NewSites(testConfig(t), filepath.Join(t.TempDir(), "sites"), nil)

	if _, err := sites.OpenFor(t.Context(), "korenaar", "anna"); !errors.Is(err, ErrNotAWorkingCopy) {
		t.Errorf("OpenFor() error = %v, want ErrNotAWorkingCopy", err)
	}
}

func TestPersonCopyNeedsNoKey(t *testing.T) {
	cfg := testConfig(t)
	cfg.SSHKeyPath = "/nonexistent/key"
	remote := bareRemote(t, testConfig(t))
	sites := NewSites(cfg, filepath.Join(t.TempDir(), "sites"), nil)
	repository, err := sites.Clone(t.Context(), "korenaar", remote)
	if err != nil {
		t.Fatalf("Clone() error = %v", err)
	}
	repository.Close()

	mine := openFor(t, sites, "anna")
	if got := mine.(*repo).cfg.SSHKeyPath; got != "" {
		t.Errorf("a person's copy carries key %q, want none", got)
	}
}

func TestSharedCopyTreeFollowsASend(t *testing.T) {
	sites, _, _ := sharedSite(t)

	mine := openFor(t, sites, "anna")
	save(t, mine, "Anna", "Edit about", map[string]string{"content/about.md": "anna\n"})
	share(t, sites, "anna")

	shared, _ := sites.Dir("korenaar")
	data, err := os.ReadFile(filepath.Join(shared, "content", "about.md"))
	if err != nil || string(data) != "anna\n" {
		t.Errorf("shared tree holds %q, %v, want the shared save on disk", data, err)
	}
}

func TestShareReachesTheSharedVersion(t *testing.T) {
	sites, _, _ := sharedSite(t)

	mine := openFor(t, sites, "anna")
	save(t, mine, "Anna", "Edit about", map[string]string{"content/about.md": "anna\n"})

	shared, err := sites.Open(t.Context(), "korenaar")
	if err != nil {
		t.Fatal(err)
	}
	defer shared.Close()
	if got := readSite(t, shared, "content/about.md"); got != "" {
		t.Fatalf("shared version holds %q before sharing, want nothing", got)
	}

	share(t, sites, "anna")

	if got := readSite(t, shared, "content/about.md"); got != "anna\n" {
		t.Errorf("shared version holds %q, want the save", got)
	}
	versions, err := shared.History(t.Context(), 1)
	if err != nil || len(versions) == 0 || versions[0].Author.Name != "Anna" {
		t.Errorf("shared history = %v, %v, want Anna's version on top", versions, err)
	}

	unshared, err := sites.Unshared(t.Context(), "korenaar", "anna")
	if err != nil || unshared != 0 {
		t.Errorf("Unshared() = %d, %v, want 0", unshared, err)
	}
}

func TestShareKeepsAnotherPagesSave(t *testing.T) {
	sites, _, _ := sharedSite(t)

	anna := openFor(t, sites, "anna")
	bert := openFor(t, sites, "bert")
	save(t, anna, "Anna", "Edit about", map[string]string{"content/about.md": "anna\n"})
	save(t, bert, "Bert", "Edit contact", map[string]string{"content/contact.md": "bert\n"})

	share(t, sites, "anna")
	result := share(t, sites, "bert")
	if result.Conflict != nil {
		t.Fatalf("Share() conflict = %v, want none for different pages", result.Conflict)
	}

	shared, _ := sites.Open(t.Context(), "korenaar")
	defer shared.Close()
	if readSite(t, shared, "content/about.md") != "anna\n" || readSite(t, shared, "content/contact.md") != "bert\n" {
		t.Errorf("shared version does not hold both saves")
	}
	if readSite(t, bert, "content/about.md") != "anna\n" {
		t.Errorf("bert's copy does not hold anna's save after sharing")
	}
}

func TestTwoSharesRacing(t *testing.T) {
	sites, _, _ := sharedSite(t)

	people := []string{"anna", "bert", "cees", "dora"}
	for _, person := range people {
		mine := openFor(t, sites, person)
		save(t, mine, person, "Edit "+person, map[string]string{"content/" + person + ".md": person + "\n"})
	}

	var wg sync.WaitGroup
	errs := make(chan error, len(people))
	for _, person := range people {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := sites.Share(t.Context(), "korenaar", person); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Errorf("Share() error = %v, want nil", err)
	}

	shared, _ := sites.Open(t.Context(), "korenaar")
	defer shared.Close()
	for _, person := range people {
		if readSite(t, shared, "content/"+person+".md") != person+"\n" {
			t.Errorf("shared version misses %s's save", person)
		}
	}
}

func TestShareStopsOnTheSameField(t *testing.T) {
	sites, _, _ := sharedSite(t)

	anna := openFor(t, sites, "anna")
	bert := openFor(t, sites, "bert")
	save(t, anna, "Anna", "Edit home", map[string]string{"content/_index.md": "---\ntitle: Anna\n---\n"})
	save(t, bert, "Bert", "Edit home", map[string]string{"content/_index.md": "---\ntitle: Bert\n---\n"})

	share(t, sites, "anna")
	result := share(t, sites, "bert")
	if result.Conflict == nil || len(result.Conflict.Paths) != 1 {
		t.Fatalf("Share() conflict = %v, want one path", result.Conflict)
	}

	shared, _ := sites.Open(t.Context(), "korenaar")
	defer shared.Close()
	if got := readSite(t, shared, "content/_index.md"); got != "---\ntitle: Anna\n---\n" {
		t.Errorf("shared version holds %q, want anna's save untouched", got)
	}

	unshared, err := sites.Unshared(t.Context(), "korenaar", "bert")
	if err != nil || unshared == 0 {
		t.Errorf("Unshared() = %d, %v, want bert's waiting change counted", unshared, err)
	}

	again, err := sites.Share(t.Context(), "korenaar", "bert")
	if err != nil || again.Conflict == nil || again.Conflict.Paths[0].Path != "content/_index.md" {
		t.Errorf("Share() during a question = %+v, %v, want the same question again", again, err)
	}
}

func TestRemoteArrivalReachesACopyOnOpen(t *testing.T) {
	sites, remote, cfg := sharedSite(t)
	openFor(t, sites, "anna")

	other := elsewhere(t, cfg, remote)
	pushFrom(t, cfg, other, "content/news.md", "news\n", "News from elsewhere")

	shared, _ := sites.Open(t.Context(), "korenaar")
	if _, err := shared.Sync(t.Context()); err != nil {
		t.Fatal(err)
	}
	shared.Close()

	again := openFor(t, sites, "anna")
	if got := readSite(t, again, "content/news.md"); got != "news\n" {
		t.Errorf("anna's copy holds %q, want the arrival", got)
	}
}

func TestOpenForLeavesUnsharedWorkAlone(t *testing.T) {
	sites, _, _ := sharedSite(t)

	anna := openFor(t, sites, "anna")
	bert := openFor(t, sites, "bert")
	save(t, anna, "Anna", "Edit about", map[string]string{"content/about.md": "anna\n"})
	save(t, bert, "Bert", "Edit about", map[string]string{"content/about.md": "bert\n"})
	share(t, sites, "anna")

	again := openFor(t, sites, "bert")
	if got := readSite(t, again, "content/about.md"); got != "bert\n" {
		t.Errorf("opening bert's copy changed it to %q, want bert's own work", got)
	}
}

func TestRemoveForRemovesOnlyThatCopy(t *testing.T) {
	sites, _, _ := sharedSite(t)
	openFor(t, sites, "anna")
	openFor(t, sites, "bert")

	if err := sites.RemoveFor("korenaar", "anna"); err != nil {
		t.Fatalf("RemoveFor() error = %v", err)
	}
	anna, _ := sites.DirFor("korenaar", "anna")
	bert, _ := sites.DirFor("korenaar", "bert")
	if isWorkingCopy(anna) {
		t.Errorf("anna's copy is still there")
	}
	if !isWorkingCopy(bert) {
		t.Errorf("bert's copy was removed too")
	}
	if n, err := sites.Unshared(t.Context(), "korenaar", "anna"); err != nil || n != 0 {
		t.Errorf("Unshared() of a removed copy = %d, %v, want 0", n, err)
	}
	if _, err := sites.Open(t.Context(), "korenaar"); err != nil {
		t.Errorf("the shared copy no longer opens: %v", err)
	}
}

func TestUnsharedCountsUnsavedWork(t *testing.T) {
	sites, _, _ := sharedSite(t)
	openFor(t, sites, "anna")
	dir, _ := sites.DirFor("korenaar", "anna")
	writeFile(t, filepath.Join(dir, "content", "_index.md"), "---\ntitle: Half saved\n---\n")

	if n, err := sites.Unshared(t.Context(), "korenaar", "anna"); err != nil || n == 0 {
		t.Errorf("Unshared() = %d, %v, want unsaved work counted", n, err)
	}
}

func TestWhatAPreviewWritesIsNotWork(t *testing.T) {
	sites, remote, cfg := sharedSite(t)
	anna := openFor(t, sites, "anna")
	dir, _ := sites.DirFor("korenaar", "anna")
	writeFile(t, filepath.Join(dir, ".quarto", "xref", "index.json"), "{}")
	writeFile(t, filepath.Join(dir, "_book", "index.html"), "<h1>old render</h1>")

	if n, err := sites.Unshared(t.Context(), "korenaar", "anna"); err != nil || n != 0 {
		t.Errorf("Unshared() = %d, %v, want a preview's output left out", n, err)
	}
	if status, err := anna.Status(t.Context()); err != nil || status.UnsavedWork {
		t.Errorf("Status() = %+v, %v, want no unsaved work", status, err)
	}

	other := elsewhere(t, cfg, remote)
	pushFrom(t, cfg, other, "content/news.md", "news\n", "News from elsewhere")
	pushFrom(t, cfg, other, "_book/index.html", "<h1>committed render</h1>", "Someone commits the output")
	shared, _ := sites.Open(t.Context(), "korenaar")
	if _, err := shared.Sync(t.Context()); err != nil {
		t.Fatal(err)
	}
	shared.Close()

	again := openFor(t, sites, "anna")
	if got := readSite(t, again, "content/news.md"); got != "news\n" {
		t.Errorf("anna's copy holds %q, want the arrival despite the preview's output", got)
	}
	if got := readSite(t, again, "_book/index.html"); got != "<h1>committed render</h1>" {
		t.Errorf("the committed file is %q", got)
	}
	if got := readSite(t, again, ".quarto/xref/index.json"); got != "{}" {
		t.Errorf("output nobody committed was touched: %q", got)
	}
}

func TestOpeningForTheFirstTimeTwiceAtOnce(t *testing.T) {
	sites, _, _ := sharedSite(t)

	var wg sync.WaitGroup
	errs := make(chan error, 6)
	for range 6 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			repository, err := sites.OpenFor(t.Context(), "korenaar", "anna")
			if err != nil {
				errs <- err
				return
			}
			if _, err := repository.History(t.Context(), 1); err != nil {
				errs <- err
			}
			repository.Close()
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Errorf("a first open at the same time failed: %v", err)
	}

	dir, _ := sites.DirFor("korenaar", "anna")
	leftovers, _ := filepath.Glob(filepath.Join(filepath.Dir(dir), ".making-*"))
	if len(leftovers) != 0 {
		t.Errorf("a copy being made was left behind: %v", leftovers)
	}
}

func TestAPersonWaitingOnAQuestionStillSeesTheirOwnWork(t *testing.T) {
	sites, _, _ := sharedSite(t)

	anna := openFor(t, sites, "anna")
	bert := openFor(t, sites, "bert")
	save(t, anna, "Anna", "Edit home", map[string]string{"content/_index.md": "---\ntitle: Anna\n---\n"})
	save(t, bert, "Bert", "Edit home", map[string]string{"content/_index.md": "---\ntitle: Bert\n---\n"})
	share(t, sites, "anna")
	if result := share(t, sites, "bert"); result.Conflict == nil {
		t.Fatalf("Share() asked nothing, want a question")
	}

	dir, _ := sites.DirFor("korenaar", "bert")
	data, err := os.ReadFile(filepath.Join(dir, "content", "_index.md"))
	if err != nil || string(data) != "---\ntitle: Bert\n---\n" {
		t.Errorf("bert's copy holds %q, %v, want his own work while the question waits", data, err)
	}

	openFor(t, sites, "bert")
	data, err = os.ReadFile(filepath.Join(dir, "content", "_index.md"))
	if err != nil || string(data) != "---\ntitle: Bert\n---\n" {
		t.Errorf("reopened, bert's copy holds %q, %v, want his own work", data, err)
	}
}

func TestAPersonsCopyComparesWithThePublishedSite(t *testing.T) {
	sites, _, _ := sharedSite(t)

	anna := openFor(t, sites, "anna")
	save(t, anna, "Anna", "Add prices", map[string]string{"content/prices.md": "prices\n"})
	share(t, sites, "anna")

	changes, err := anna.Changes(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 1 || changes[0].Path != "content/prices.md" || changes[0].Kind != Added {
		t.Errorf("Changes() = %+v, want the shared but unpublished page as added", changes)
	}
	if _, err := anna.ReadPublished(t.Context(), "content/prices.md"); !errors.Is(err, ErrNotPublished) {
		t.Errorf("ReadPublished() error = %v, want ErrNotPublished", err)
	}
	if data, err := anna.ReadPublished(t.Context(), "content/_index.md"); err != nil || len(data) == 0 {
		t.Errorf("ReadPublished() of a published page = %q, %v", data, err)
	}

	shared, _ := sites.Open(t.Context(), "korenaar")
	if _, err := shared.Publish(t.Context()); err != nil {
		t.Fatal(err)
	}
	shared.Close()

	again := openFor(t, sites, "anna")
	if changes, err := again.Changes(t.Context()); err != nil || len(changes) != 0 {
		t.Errorf("Changes() after publishing = %+v, %v, want nothing", changes, err)
	}
}

func TestWhatIsNotPublishedCountsEveryonesSharedSaves(t *testing.T) {
	sites, _, _ := sharedSite(t)

	anna := openFor(t, sites, "anna")
	bert := openFor(t, sites, "bert")
	save(t, bert, "Bert", "Edit contact", map[string]string{"content/contact.md": "bert\n"})
	share(t, sites, "bert")
	save(t, anna, "Anna", "Edit about", map[string]string{"content/about.md": "anna\n"})

	status, err := openFor(t, sites, "anna").Status(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if status.Unpublished != 2 {
		t.Errorf("Unpublished = %d, want anna's save and bert's shared one", status.Unpublished)
	}
	if n, err := sites.Unshared(t.Context(), "korenaar", "anna"); err != nil || n != 1 {
		t.Errorf("Unshared() = %d, %v, want anna's one save", n, err)
	}
}

func TestACopyWithNothingOfItsOwnFollowsARewrittenSharedVersion(t *testing.T) {
	sites, remote, cfg := sharedSite(t)
	shared, _ := sites.Open(t.Context(), "korenaar")
	save(t, shared, "Seed", "Edit about", map[string]string{"content/about.md": "waiting to be published\n"})
	shared.Close()

	anna := openFor(t, sites, "anna")
	if got := readSite(t, anna, "content/about.md"); got != "waiting to be published\n" {
		t.Fatalf("anna's copy holds %q", got)
	}

	other := elsewhere(t, cfg, remote)
	pushFrom(t, cfg, other, "content/news.md", "news\n", "News from elsewhere")
	shared, _ = sites.Open(t.Context(), "korenaar")
	if _, err := shared.Sync(t.Context()); err != nil {
		t.Fatal(err)
	}
	shared.Close()

	again := openFor(t, sites, "anna")
	if got := readSite(t, again, "content/news.md"); got != "news\n" {
		t.Errorf("anna's copy did not follow the shared version: news is %q", got)
	}
	if got := readSite(t, again, "content/about.md"); got != "waiting to be published\n" {
		t.Errorf("about is %q", got)
	}
}

func TestACopyWithWorkOfItsOwnIsNotThrownAway(t *testing.T) {
	sites, remote, cfg := sharedSite(t)
	shared, _ := sites.Open(t.Context(), "korenaar")
	save(t, shared, "Seed", "Edit about", map[string]string{"content/about.md": "waiting to be published\n"})
	shared.Close()

	anna := openFor(t, sites, "anna")
	save(t, anna, "Anna", "Anna's page", map[string]string{"content/anna.md": "mine\n"})

	other := elsewhere(t, cfg, remote)
	pushFrom(t, cfg, other, "content/news.md", "news\n", "News from elsewhere")
	shared, _ = sites.Open(t.Context(), "korenaar")
	if _, err := shared.Sync(t.Context()); err != nil {
		t.Fatal(err)
	}
	shared.Close()

	again := openFor(t, sites, "anna")
	if got := readSite(t, again, "content/anna.md"); got != "mine\n" {
		t.Errorf("anna's own work is %q", got)
	}
}
