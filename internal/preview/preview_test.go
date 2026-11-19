package preview_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/httpapi"
	"github.com/quiqr/chatcms-app/internal/policy"
	"github.com/quiqr/chatcms-app/internal/preview"
	"github.com/quiqr/chatcms-app/internal/workspace"
	"github.com/quiqr/chatcms-app/internal/workspace/local"
	"github.com/quiqr/chatcms-app/internal/workspace/workspacetest"
)

const under = "/preview/an/org/"

func hugoBin(t *testing.T) string {
	t.Helper()

	if named := os.Getenv("CHATCMS_HUGO_BIN"); named != "" {
		return named
	}
	found, err := exec.LookPath("hugo")
	if err != nil {
		t.Skipf("no Hugo on this machine: %v", err)
	}
	return found
}

func writeSite(t *testing.T, dir, title, layout string) {
	t.Helper()

	files := map[string]string{
		"hugo.toml":          "baseURL = \"https://example.org/\"\ntitle = \"" + title + "\"\n",
		"content/_index.md":  "---\ntitle: " + title + "\n---\n\nThe body of " + title + ".\n",
		"layouts/index.html": layout,
	}
	for name, content := range files {
		where := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(where), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(where, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

const goodLayout = "<!doctype html><html><body><h1>{{ .Title }}</h1>{{ .Content }}</body></html>\n"

type copies struct {
	mu   sync.Mutex
	root string
	made map[workspace.Key]string
	t    *testing.T
}

func (c *copies) Dir(ctx context.Context, key workspace.Key) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if dir, found := c.made[key]; found {
		return dir, nil
	}
	dir := filepath.Join(c.root, key.Project.String(), key.Person.String(), "working-copy")
	writeSite(c.t, dir, "Korenaar Bakery", goodLayout)
	c.made[key] = dir
	return dir, nil
}

func (c *copies) Unshared(ctx context.Context, key workspace.Key) (int, error) { return 0, nil }

func (c *copies) Remove(ctx context.Context, key workspace.Key) error { return nil }

type sources struct{ organisation string }

func (s sources) Resource(ctx context.Context, projectID uuid.UUID) (policy.Resource, error) {
	return policy.Project(s.organisation, projectID.String()), nil
}

type harness struct {
	copies  *copies
	w       *workspace.Workspaces
	service *preview.Service
	project uuid.UUID
}

func newHarness(t *testing.T, most int) *harness {
	t.Helper()

	backend, err := local.New(local.Settings{HugoBin: hugoBin(t), Ready: 60 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	h := &harness{
		copies:  &copies{root: t.TempDir(), made: map[workspace.Key]string{}, t: t},
		project: uuid.New(),
	}
	allowed := func(ctx context.Context, key workspace.Key) (bool, error) { return true, nil }
	h.w, err = workspace.New(backend, workspacetest.NewStore(), h.copies, allowed, workspace.Settings{Most: most, Idle: 15 * time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { h.w.Close(context.Background()) })

	h.service, err = preview.NewService(sources{organisation: uuid.NewString()}, h.w, "http://example.org/")
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func (h *harness) editor() policy.Actor {
	return policy.Actor{PersonID: uuid.NewString(), Projects: map[string]policy.Role{h.project.String(): policy.RoleEditor}}
}

func (h *harness) viewer() policy.Actor {
	return policy.Actor{PersonID: uuid.NewString(), Projects: map[string]policy.Role{h.project.String(): policy.RoleViewer}}
}

func (h *harness) dirOf(t *testing.T, actor policy.Actor) string {
	t.Helper()
	dir, _ := h.copies.Dir(t.Context(), workspace.Key{Person: uuid.MustParse(actor.PersonID), Project: h.project})
	return dir
}

func (h *harness) handler(t *testing.T, actor policy.Actor) http.Handler {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	handler, err := h.service.Handler(ctx, actor, h.project, under)
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	return handler
}

func asked(t *testing.T, handler http.Handler, path string) (int, string) {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, path, nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	body, _ := io.ReadAll(recorder.Body)
	return recorder.Code, string(body)
}

func waitToSee(t *testing.T, h *harness, actor policy.Actor, want string) string {
	t.Helper()

	deadline := time.Now().Add(20 * time.Second)
	for {
		_, body := asked(t, h.handler(t, actor), "http://example.org"+under)
		if strings.Contains(body, want) {
			return body
		}
		if time.Now().After(deadline) {
			t.Fatalf("%q did not show within 20 seconds; the preview says %q", want, body)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func TestAPreviewShowsTheSite(t *testing.T) {
	h := newHarness(t, 8)

	status, body := asked(t, h.handler(t, h.editor()), "http://example.org"+under)
	if status != http.StatusOK || !strings.Contains(body, "Korenaar Bakery") {
		t.Fatalf("status = %d, body %q, want the site", status, body)
	}
}

func TestAChangeShowsInThePreviewWithoutRestartingIt(t *testing.T) {
	h := newHarness(t, 8)
	anna := h.editor()
	waitToSee(t, h, anna, "The body of Korenaar Bakery")

	if err := os.WriteFile(filepath.Join(h.dirOf(t, anna), "content/_index.md"),
		[]byte("---\ntitle: Korenaar Bakery\n---\n\nNow with plums.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	waitToSee(t, h, anna, "Now with plums")

	if h.w.Running() != 1 {
		t.Errorf("%d previews are running, want the same one", h.w.Running())
	}
}

func TestTwoPeopleSeeTheirOwnChange(t *testing.T) {
	h := newHarness(t, 8)
	anna, bert := h.editor(), h.editor()
	for name, actor := range map[string]policy.Actor{"Anna's plums": anna, "Bert's pears": bert} {
		writeSite(t, h.dirOf(t, actor), name, goodLayout)
	}

	annaSees := waitToSee(t, h, anna, "Anna&rsquo;s plums")
	bertSees := waitToSee(t, h, bert, "Bert&rsquo;s pears")
	if strings.Contains(annaSees, "Bert") || strings.Contains(bertSees, "Anna") {
		t.Errorf("one person sees the other's change")
	}
	if h.w.Running() != 2 {
		t.Errorf("%d previews are running, want one each", h.w.Running())
	}
}

func TestAskingTwiceUsesTheSamePreview(t *testing.T) {
	h := newHarness(t, 8)
	anna := h.editor()
	h.handler(t, anna)
	h.handler(t, anna)

	if h.w.Running() != 1 {
		t.Errorf("%d previews are running, want one", h.w.Running())
	}
}

func TestAPreviewNobodyHasLookedAtIsStopped(t *testing.T) {
	h := newHarness(t, 8)
	anna, bert := h.editor(), h.editor()
	h.handler(t, anna)
	kept := h.handler(t, bert)

	later := time.Now().Add(20 * time.Minute)
	h.w.PretendTimeIs(func() time.Time { return later })
	h.handler(t, bert)
	h.w.PretendTimeIs(func() time.Time { return later.Add(time.Minute) })
	h.w.SweepNow()

	if h.w.Running() != 1 {
		t.Errorf("%d previews are running, want only the one somebody is looking at", h.w.Running())
	}
	if status, _ := asked(t, kept, "http://example.org"+under); status != http.StatusOK {
		t.Errorf("the kept preview answers %d", status)
	}
}

func TestOnlySoManyPreviewsRunAtOnce(t *testing.T) {
	h := newHarness(t, 2)
	h.handler(t, h.editor())
	h.handler(t, h.editor())

	_, err := h.service.Handler(t.Context(), h.editor(), h.project, under)
	var refused *httpapi.RefusedError
	if !errors.As(err, &refused) || refused.Key != "refused.too-many-sites-being-previewed" {
		t.Fatalf("Handler() error = %v, want the cap's refusal", err)
	}
	if h.w.Running() != 2 {
		t.Errorf("%d previews are running, want the two allowed", h.w.Running())
	}
}

func TestStoppingEverythingLeavesNothingRunning(t *testing.T) {
	h := newHarness(t, 8)
	h.handler(t, h.editor())

	h.w.Close(t.Context())
	h.w.Close(t.Context())

	if h.w.Running() != 0 {
		t.Errorf("%d previews survived", h.w.Running())
	}
}

func TestStopSleepsEveryPreviewOfTheSite(t *testing.T) {
	h := newHarness(t, 8)
	h.handler(t, h.editor())
	h.handler(t, h.editor())

	h.service.Stop(h.project)

	if h.w.Running() != 0 {
		t.Errorf("%d previews are running after stopping the site's", h.w.Running())
	}
}

func TestAPreviewOfSomethingThatWillNotBuildSaysSo(t *testing.T) {
	h := newHarness(t, 8)
	anna := h.editor()
	if err := os.WriteFile(filepath.Join(h.dirOf(t, anna), "hugo.toml"), []byte("this is not = [valid\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := h.service.Handler(t.Context(), anna, h.project, under)
	var refused *httpapi.RefusedError
	if !errors.As(err, &refused) || refused.Key != "refused.preview-did-not-start-site" {
		t.Fatalf("Handler() error = %v, want the did-not-start refusal", err)
	}
	if h.w.Running() != 0 {
		t.Errorf("%d previews are running after one that could not start", h.w.Running())
	}
}

func TestAViewerMayPreview(t *testing.T) {
	h := newHarness(t, 8)
	if _, body := asked(t, h.handler(t, h.viewer()), "http://example.org"+under); !strings.Contains(body, "Korenaar") {
		t.Errorf("a viewer was not shown the preview")
	}
}

func TestSomebodyWithNoPartInTheProjectIsToldThereIsNothing(t *testing.T) {
	h := newHarness(t, 8)
	stranger := policy.Actor{PersonID: uuid.NewString()}

	if _, err := h.service.Handler(t.Context(), stranger, h.project, under); !errors.Is(err, httpapi.ErrMissing) {
		t.Errorf("Handler() error = %v, want ErrMissing", err)
	}
	if h.w.Running() != 0 {
		t.Errorf("a preview started for somebody with no part in the project")
	}
}

func TestSomebodyWithoutAnIdentityIsRefused(t *testing.T) {
	h := newHarness(t, 8)
	nameless := policy.Actor{PersonID: "", Projects: map[string]policy.Role{h.project.String(): policy.RoleEditor}}

	if _, err := h.service.Handler(t.Context(), nameless, h.project, under); err == nil {
		t.Errorf("Handler() for somebody nameless succeeded, want it refused")
	}
}

func TestNewServiceRefusesMissingParts(t *testing.T) {
	if _, err := preview.NewService(nil, nil, ""); !errors.Is(err, preview.ErrIncomplete) {
		t.Errorf("NewService() error = %v, want ErrIncomplete", err)
	}
}

func TestRefusalLeavesOtherErrorsAlone(t *testing.T) {
	other := errors.New("other")
	if preview.Refusal(other) != other {
		t.Errorf("Refusal() changed an unrelated error")
	}
}

func (h *harness) implementer() policy.Actor {
	return policy.Actor{PersonID: uuid.NewString(), Projects: map[string]policy.Role{h.project.String(): policy.RoleImplementer}}
}

func TestOpeningStartsTheWorkspaceAndAnImplementerSeesTheSteps(t *testing.T) {
	h := newHarness(t, 8)
	implementer := h.implementer()

	standing, err := h.service.Open(t.Context(), implementer, h.project, under)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if standing.State == workspace.Sleeping {
		t.Fatalf("Open() state = %v, want it starting or warm", standing.State)
	}

	deadline := time.Now().Add(60 * time.Second)
	for standing.State != workspace.Warm && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
		standing, _ = h.service.Standing(t.Context(), implementer, h.project)
	}
	if standing.State != workspace.Warm {
		t.Fatalf("state = %v, want warm", standing.State)
	}
	if len(standing.Steps) != 3 || len(standing.Log) == 0 {
		t.Errorf("an implementer sees %+v, want the steps and the log", standing)
	}
}

func TestAnImplementerSeesTheBackendAndCanStartAgain(t *testing.T) {
	h := newHarness(t, 8)
	implementer := h.implementer()
	if _, err := h.service.Open(t.Context(), implementer, h.project, under); err != nil {
		t.Fatal(err)
	}
	standing, err := h.service.Standing(t.Context(), implementer, h.project)
	if err != nil || standing.Backend == "" || standing.Since.IsZero() {
		t.Fatalf("Standing() = %+v, %v, want the backend and since when", standing, err)
	}
	again, err := h.service.Restart(t.Context(), implementer, h.project, under)
	if err != nil || again.State == workspace.Sleeping || again.Since.Before(standing.Since) {
		t.Errorf("Restart() = %+v, %v, want it started again", again, err)
	}
}

func TestOnlyAnImplementerMayStartAgain(t *testing.T) {
	h := newHarness(t, 8)
	editor := h.editor()
	if _, err := h.service.Open(t.Context(), editor, h.project, under); err != nil {
		t.Fatal(err)
	}
	if _, err := h.service.Restart(t.Context(), editor, h.project, under); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("Restart() error = %v, want ErrNotAllowed", err)
	}
	standing, err := h.service.Standing(t.Context(), editor, h.project)
	if err != nil || standing.Backend != "" || standing.Since.IsZero() {
		t.Errorf("an editor was given %+v, %v, want since when but no backend", standing, err)
	}
}

func TestAContentCreatorIsNotGivenTheSteps(t *testing.T) {
	h := newHarness(t, 8)
	for name, actor := range map[string]policy.Actor{"somebody who edits": h.editor(), "somebody who looks": h.viewer()} {
		if _, err := h.service.Open(t.Context(), actor, h.project, under); err != nil {
			t.Fatalf("%s Open() error = %v", name, err)
		}
		standing, err := h.service.Standing(t.Context(), actor, h.project)
		if err != nil {
			t.Fatal(err)
		}
		if standing.Steps != nil || standing.Log != nil {
			t.Errorf("a %s was given %+v, want no steps or log", name, standing)
		}
	}
}

func TestAFailedStartGivesAReasonWorthShowing(t *testing.T) {
	h := newHarness(t, 8)
	anna := h.editor()
	if err := os.WriteFile(filepath.Join(h.dirOf(t, anna), "hugo.toml"), []byte("this is not = [valid\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	h.service.Handler(t.Context(), anna, h.project, under)

	standing, err := h.service.Standing(t.Context(), anna, h.project)
	if err != nil || standing.State != workspace.Sleeping || standing.Reason != "refused.preview-did-not-start-site" {
		t.Errorf("Standing() = %+v, %v, want sleeping with the did-not-start reason", standing, err)
	}
}

func TestOpeningAtTheCapIsRefused(t *testing.T) {
	h := newHarness(t, 1)
	h.handler(t, h.editor())

	_, err := h.service.Open(t.Context(), h.editor(), h.project, under)
	var refused *httpapi.RefusedError
	if !errors.As(err, &refused) || refused.Key != "refused.too-many-sites-being-previewed" {
		t.Errorf("Open() error = %v, want the cap's refusal", err)
	}
}

func TestStandingNeedsAPartInTheProject(t *testing.T) {
	h := newHarness(t, 8)
	stranger := policy.Actor{PersonID: uuid.NewString()}
	if _, err := h.service.Standing(t.Context(), stranger, h.project); !errors.Is(err, httpapi.ErrMissing) {
		t.Errorf("Standing() error = %v, want ErrMissing", err)
	}
	if _, err := h.service.Open(t.Context(), stranger, h.project, under); !errors.Is(err, httpapi.ErrMissing) {
		t.Errorf("Open() error = %v, want ErrMissing", err)
	}
}
