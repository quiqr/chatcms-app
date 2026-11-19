package workspace_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/runtime"
	"github.com/quiqr/chatcms-app/internal/workspace"
	"github.com/quiqr/chatcms-app/internal/workspace/workspacetest"
)

type harness struct {
	backend *workspacetest.Backend
	store   *workspacetest.Store
	copies  *workspacetest.Copies
	allowed map[workspace.Key]bool
	mu      sync.Mutex
	clock   time.Time
	w       *workspace.Workspaces
}

func newHarness(t *testing.T, most int) *harness {
	t.Helper()

	h := &harness{
		backend: workspacetest.NewBackend(),
		store:   workspacetest.NewStore(),
		copies:  workspacetest.NewCopies(),
		allowed: map[workspace.Key]bool{},
		clock:   time.Date(2026, 11, 18, 9, 0, 0, 0, time.UTC),
	}
	h.w = h.manager(t, most)
	return h
}

func (h *harness) manager(t *testing.T, most int) *workspace.Workspaces {
	t.Helper()

	made, err := workspace.New(h.backend, h.store, h.copies, h.isAllowed, workspace.Settings{Most: most, Idle: 15 * time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	made.PretendTimeIs(h.now)
	t.Cleanup(func() { made.Close(context.Background()) })
	return made
}

func (h *harness) isAllowed(ctx context.Context, key workspace.Key) (bool, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	allowed, found := h.allowed[key]
	return !found || allowed, nil
}

func (h *harness) now() time.Time {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.clock
}

func (h *harness) pass(d time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clock = h.clock.Add(d)
}

func aKey() workspace.Key {
	return workspace.Key{Person: uuid.New(), Project: uuid.New()}
}

func endpoint(t *testing.T, w *workspace.Workspaces, key workspace.Key) workspace.Endpoint {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	found, err := w.Endpoint(ctx, key, "http://chatcms.test/preview/o/p/")
	if err != nil {
		t.Fatalf("Endpoint() error = %v", err)
	}
	return found
}

func TestTwoPeopleGetAWorkspaceEach(t *testing.T) {
	h := newHarness(t, 8)
	project := uuid.New()
	anna := workspace.Key{Person: uuid.New(), Project: project}
	bert := workspace.Key{Person: uuid.New(), Project: project}

	if endpoint(t, h.w, anna) == endpoint(t, h.w, bert) {
		t.Errorf("two people share one workspace")
	}
	if h.backend.Running() != 2 {
		t.Errorf("running = %d, want 2", h.backend.Running())
	}
	if h.backend.Started[0].Dir == h.backend.Started[1].Dir {
		t.Errorf("two people were given the same copy of the site")
	}
}

func TestAskingTwiceUsesTheSameWorkspace(t *testing.T) {
	h := newHarness(t, 8)
	key := aKey()

	first := endpoint(t, h.w, key)
	second := endpoint(t, h.w, key)
	if first != second || h.backend.StartCount() != 1 {
		t.Errorf("asking twice started %d workspaces, want 1", h.backend.StartCount())
	}
}

func TestOpenStartsWithoutWaiting(t *testing.T) {
	h := newHarness(t, 8)
	h.backend.Hold = make(chan struct{})
	key := aKey()

	state, err := h.w.Open(t.Context(), key, "u")
	if err != nil || state != workspace.Starting {
		t.Fatalf("Open() = %v, %v, want starting at once", state, err)
	}
	var report workspace.Report
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		report, _ = h.w.Report(key)
		if report.Steps[2].State == workspace.StepRunning {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if report.State != workspace.Starting || report.Steps[2].State != workspace.StepRunning || report.Steps[0].State != workspace.StepDone {
		t.Errorf("Report() = %+v, want the preview being built", report)
	}

	close(h.backend.Hold)
	endpoint(t, h.w, key)
	report, _ = h.w.Report(key)
	if report.State != workspace.Warm {
		t.Errorf("Report() state = %v, want warm", report.State)
	}
	for _, step := range report.Steps {
		if step.State != workspace.StepDone {
			t.Errorf("step %s = %s, want done", step.Step, step.State)
		}
	}
	if len(report.Log) == 0 {
		t.Errorf("Report() has no log")
	}
	if record, _ := h.store.Get(key); record.State != workspace.Warm || record.Handle == "" {
		t.Errorf("stored = %+v, want warm with a handle", record)
	}
}

func TestAReportSaysOnWhatAndSinceWhen(t *testing.T) {
	h := newHarness(t, 8)
	key := aKey()
	if report, _ := h.w.Report(key); report.Backend == "" || !report.Since.IsZero() {
		t.Errorf("asleep Report() = %+v, want the backend and no since", report)
	}
	endpoint(t, h.w, key)
	report, _ := h.w.Report(key)
	if !report.Since.Equal(h.now()) || report.Backend != h.w.Backend() {
		t.Errorf("Report() = %+v, want since %v on %s", report, h.now(), h.w.Backend())
	}
}

func TestRestartStopsAndStartsAgain(t *testing.T) {
	h := newHarness(t, 8)
	key := aKey()
	endpoint(t, h.w, key)
	if _, err := h.w.Restart(t.Context(), key, "http://chatcms.test/preview/o/p/"); err != nil {
		t.Fatal(err)
	}
	endpoint(t, h.w, key)
	if h.backend.StartCount() != 2 || h.backend.Running() != 1 {
		t.Errorf("started %d, running %d, want 2 and 1", h.backend.StartCount(), h.backend.Running())
	}
}

func TestAFailedStartIsSleepingWithAReason(t *testing.T) {
	h := newHarness(t, 8)
	h.backend.Fail = workspace.ErrNotReady
	key := aKey()

	_, err := h.w.Endpoint(t.Context(), key, "u")
	if !errors.Is(err, workspace.ErrNotReady) {
		t.Fatalf("Endpoint() error = %v, want ErrNotReady", err)
	}
	report, _ := h.w.Report(key)
	if report.State != workspace.Sleeping || report.Reason != "did-not-start" || report.Steps[2].State != workspace.StepFailed {
		t.Errorf("Report() = %+v, want a failed build", report)
	}
	if record, _ := h.store.Get(key); record.State != workspace.Sleeping || record.Reason == "" {
		t.Errorf("stored = %+v, want sleeping with the reason", record)
	}

	h.backend.Fail = nil
	endpoint(t, h.w, key)
	if h.backend.StartCount() != 2 {
		t.Errorf("asking again after a failure started %d times, want a fresh start", h.backend.StartCount())
	}
}

func TestTheCapRefusesAndLeavesTheOthers(t *testing.T) {
	h := newHarness(t, 2)
	endpoint(t, h.w, aKey())
	endpoint(t, h.w, aKey())

	if _, err := h.w.Open(t.Context(), aKey(), "u"); !errors.Is(err, workspace.ErrTooMany) {
		t.Errorf("Open() error = %v, want ErrTooMany", err)
	}
	if h.backend.Running() != 2 || h.w.Running() != 2 {
		t.Errorf("running = %d, want the two untouched", h.backend.Running())
	}
}

func TestAWorkspaceNobodyUsesGoesToSleep(t *testing.T) {
	h := newHarness(t, 8)
	idle, busy := aKey(), aKey()
	endpoint(t, h.w, idle)
	endpoint(t, h.w, busy)

	h.pass(10 * time.Minute)
	endpoint(t, h.w, busy)
	h.pass(10 * time.Minute)
	h.w.SweepNow()

	if report, _ := h.w.Report(idle); report.State != workspace.Sleeping {
		t.Errorf("idle workspace is %v, want sleeping", report.State)
	}
	if record, _ := h.store.Get(idle); record.State != workspace.Sleeping {
		t.Errorf("idle workspace stored as %v, want sleeping", record.State)
	}
	if report, _ := h.w.Report(busy); report.State != workspace.Warm {
		t.Errorf("used workspace is %v, want warm", report.State)
	}
	if h.backend.Running() != 1 {
		t.Errorf("running = %d, want 1", h.backend.Running())
	}

	endpoint(t, h.w, idle)
	if h.backend.Started[len(h.backend.Started)-1].Dir != h.backend.Started[0].Dir {
		t.Errorf("woken workspace was given a different copy")
	}
}

func TestAnotherAddressStartsAfresh(t *testing.T) {
	h := newHarness(t, 8)
	key := aKey()
	if _, err := h.w.Endpoint(t.Context(), key, "http://a/"); err != nil {
		t.Fatal(err)
	}
	if _, err := h.w.Endpoint(t.Context(), key, "http://b/"); err != nil {
		t.Fatal(err)
	}
	if h.backend.StartCount() != 2 || h.backend.Running() != 1 {
		t.Errorf("started %d, running %d, want a fresh start and one running", h.backend.StartCount(), h.backend.Running())
	}
}

func TestReconcileReattaches(t *testing.T) {
	h := newHarness(t, 8)
	key := aKey()
	endpoint(t, h.w, key)
	again := h.manager(t, 8)
	if err := again.Reconcile(t.Context()); err != nil {
		t.Fatal(err)
	}
	if again.Running() != 1 {
		t.Fatalf("reconciled running = %d, want the warm one adopted", again.Running())
	}
	endpoint(t, again, key)
	if h.backend.StartCount() != 1 {
		t.Errorf("started %d times, want the running one served again", h.backend.StartCount())
	}
}

func TestReconcileStopsWhatIsGone(t *testing.T) {
	h := newHarness(t, 8)
	key := aKey()
	endpoint(t, h.w, key)
	record, _ := h.store.Get(key)
	h.backend.Forget(record.Handle)

	starting := aKey()
	h.store.Put(workspace.Record{ID: uuid.New(), Key: starting, Backend: "fake", State: workspace.Starting})
	other := aKey()
	h.store.Put(workspace.Record{ID: uuid.New(), Key: other, Backend: "elsewhere", State: workspace.Warm, Handle: "x"})

	again := h.manager(t, 8)
	if err := again.Reconcile(t.Context()); err != nil {
		t.Fatal(err)
	}
	if again.Running() != 0 {
		t.Errorf("reconciled running = %d, want none", again.Running())
	}
	for _, k := range []workspace.Key{key, starting, other} {
		if stored, _ := h.store.Get(k); stored.State != workspace.Sleeping {
			t.Errorf("stored %v, want sleeping", stored.State)
		}
	}
}

func TestClosingStopsWhatRunsAsPartOfTheApp(t *testing.T) {
	h := newHarness(t, 8)
	key := aKey()
	endpoint(t, h.w, key)

	h.w.Close(t.Context())
	h.w.Close(t.Context())

	if h.backend.Running() != 0 {
		t.Errorf("running = %d after closing, want none", h.backend.Running())
	}
	if stored, _ := h.store.Get(key); stored.State != workspace.Sleeping {
		t.Errorf("stored %v, want sleeping", stored.State)
	}
}

func TestClosingLeavesWhatRunsApart(t *testing.T) {
	h := newHarness(t, 8)
	h.backend.ApartFromApp = true
	key := aKey()
	endpoint(t, h.w, key)

	h.w.Close(t.Context())

	if h.backend.Running() != 1 {
		t.Errorf("running = %d after closing, want it left running", h.backend.Running())
	}
	again := h.manager(t, 8)
	if err := again.Reconcile(t.Context()); err != nil {
		t.Fatal(err)
	}
	endpoint(t, again, key)
	if h.backend.StartCount() != 1 {
		t.Errorf("started %d times, want it served again after the restart", h.backend.StartCount())
	}
}

func TestCleanupRemovesWhatNobodyMayUse(t *testing.T) {
	h := newHarness(t, 8)
	gone, waiting, fresh, stale := aKey(), aKey(), aKey(), aKey()
	for _, key := range []workspace.Key{gone, waiting, fresh, stale} {
		endpoint(t, h.w, key)
		h.w.Sleep(t.Context(), key)
	}
	h.allowed[gone] = false
	h.copies.Waiting[waiting] = 2
	h.pass(31 * 24 * time.Hour)
	h.store.Put(func() workspace.Record { r, _ := h.store.Get(fresh); r.LastUsed = h.now(); return r }())

	if err := h.w.CleanupNow(t.Context()); err != nil {
		t.Fatal(err)
	}

	removed := map[workspace.Key]bool{}
	for _, key := range h.copies.Removed {
		removed[key] = true
	}
	if !removed[gone] || !removed[stale] {
		t.Errorf("removed = %v, want the one nobody may use and the stale one", h.copies.Removed)
	}
	if removed[waiting] || removed[fresh] {
		t.Errorf("removed a copy with unshared work or one in recent use")
	}
	if _, found := h.store.Get(gone); found {
		t.Errorf("the record of a removed workspace is still there")
	}
}

func TestCleanupLeavesAWorkspaceInUse(t *testing.T) {
	h := newHarness(t, 8)
	key := aKey()
	endpoint(t, h.w, key)
	h.allowed[key] = false
	h.store.Put(func() workspace.Record { r, _ := h.store.Get(key); r.State = workspace.Sleeping; return r }())

	if err := h.w.CleanupNow(t.Context()); err != nil {
		t.Fatal(err)
	}
	if len(h.copies.Removed) != 0 {
		t.Errorf("removed a copy that is in use")
	}
}

func TestSleepProjectSleepsEveryPersonsWorkspace(t *testing.T) {
	h := newHarness(t, 8)
	project := uuid.New()
	anna := workspace.Key{Person: uuid.New(), Project: project}
	bert := workspace.Key{Person: uuid.New(), Project: project}
	other := aKey()
	for _, key := range []workspace.Key{anna, bert, other} {
		endpoint(t, h.w, key)
	}

	h.w.SleepProject(t.Context(), project.String())
	h.w.Sleep(t.Context(), anna)

	if h.backend.Running() != 1 || h.w.Running() != 1 {
		t.Errorf("running = %d, want only the other project's", h.backend.Running())
	}
}

func TestACopyThatCannotBeMadeIsAnError(t *testing.T) {
	h := newHarness(t, 8)
	h.copies.Fail = errors.New("nothing brought in")

	if _, err := h.w.Open(t.Context(), aKey(), "u"); err == nil {
		t.Errorf("Open() error = nil, want the copy's error")
	}
	if h.backend.StartCount() != 0 {
		t.Errorf("a workspace started without a copy")
	}
}

func TestEndpointGivesUpWithTheRequest(t *testing.T) {
	h := newHarness(t, 8)
	h.backend.Hold = make(chan struct{})
	defer close(h.backend.Hold)

	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()
	if _, err := h.w.Endpoint(ctx, aKey(), "u"); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Endpoint() error = %v, want the request's deadline", err)
	}
}

func TestNewRefusesMissingParts(t *testing.T) {
	if _, err := workspace.New(nil, nil, nil, nil, workspace.Settings{}); !errors.Is(err, workspace.ErrIncomplete) {
		t.Errorf("New() error = %v, want ErrIncomplete", err)
	}
}

func TestEndpointAddress(t *testing.T) {
	if got := (workspace.Endpoint{Host: "10.0.0.2", Port: 1313}).Address(); got != "10.0.0.2:1313" {
		t.Errorf("Address() = %q", got)
	}
}

func TestATouchIsRecordedAtMostOnceAMinute(t *testing.T) {
	h := newHarness(t, 8)
	key := aKey()
	endpoint(t, h.w, key)
	before, _ := h.store.Get(key)

	h.pass(30 * time.Second)
	endpoint(t, h.w, key)
	time.Sleep(20 * time.Millisecond)
	if after, _ := h.store.Get(key); !after.LastUsed.Equal(before.LastUsed) {
		t.Errorf("use recorded within the minute")
	}

	h.pass(time.Minute)
	endpoint(t, h.w, key)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if after, _ := h.store.Get(key); after.LastUsed.After(before.LastUsed) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Errorf("use not recorded after a minute")
}

func TestTheProjectsHugoReachesTheBackend(t *testing.T) {
	h := newHarness(t, 8)
	made, err := workspace.New(h.backend, h.store, h.copies, h.isAllowed, workspace.Settings{
		RuntimeFor: func(ctx context.Context, key workspace.Key, dir string) (runtime.Resolved, error) {
			if dir == "" {
				t.Errorf("the runtime was asked for without the person's copy")
			}
			return runtime.Resolved{Preview: runtime.Command{Program: "/nix/store/x-hugo-0.139.0/bin/hugo"}}, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { made.Close(context.Background()) })

	endpoint(t, made, aKey())
	if got := h.backend.Started[0].Run.Preview.Program; got != "/nix/store/x-hugo-0.139.0/bin/hugo" {
		t.Errorf("backend was given Hugo %q", got)
	}
}

func TestAHugoThatCannotBeHadStopsTheStart(t *testing.T) {
	h := newHarness(t, 8)
	unknown := errors.New("no such Hugo")
	made, err := workspace.New(h.backend, h.store, h.copies, h.isAllowed, workspace.Settings{
		RuntimeFor: func(ctx context.Context, key workspace.Key, dir string) (runtime.Resolved, error) {
			return runtime.Resolved{}, unknown
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { made.Close(context.Background()) })

	key := aKey()
	if _, err := made.Endpoint(t.Context(), key, "u"); !errors.Is(err, unknown) {
		t.Errorf("Endpoint() error = %v, want the Hugo's error", err)
	}
	if h.backend.StartCount() != 0 {
		t.Errorf("a workspace started without its Hugo")
	}
	if report, _ := made.Report(key); !errors.Is(report.Err, unknown) || report.State != workspace.Sleeping {
		t.Errorf("Report() = %+v, want sleeping with the Hugo's error", report)
	}
}

type tokens struct {
	mu      sync.Mutex
	issued  map[workspace.Key]int
	revoked map[workspace.Key]int
	fail    error
}

func (t *tokens) Issue(ctx context.Context, key workspace.Key) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.fail != nil {
		return "", t.fail
	}
	t.issued[key]++
	return "token-" + key.Person.String(), nil
}

func (t *tokens) Revoke(ctx context.Context, key workspace.Key) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.revoked[key]++
	return nil
}

func (t *tokens) counts(key workspace.Key) (int, int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.issued[key], t.revoked[key]
}

func withTokens(t *testing.T, h *harness) (*workspace.Workspaces, *tokens) {
	t.Helper()
	issued := &tokens{issued: map[workspace.Key]int{}, revoked: map[workspace.Key]int{}}
	made, err := workspace.New(h.backend, h.store, h.copies, h.isAllowed, workspace.Settings{Tokens: issued, Idle: 15 * time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	made.PretendTimeIs(h.now)
	t.Cleanup(func() { made.Close(context.Background()) })
	return made, issued
}

func TestAStartingWorkspaceIsGivenAToken(t *testing.T) {
	h := newHarness(t, 8)
	made, issued := withTokens(t, h)
	key := aKey()

	endpoint(t, made, key)
	if got := h.backend.Started[0].Token; got != "token-"+key.Person.String() {
		t.Errorf("the backend was given token %q", got)
	}
	if n, _ := issued.counts(key); n != 1 {
		t.Errorf("issued %d tokens, want one", n)
	}
}

func TestATokenStopsWorkingWhenItsWorkspaceSleeps(t *testing.T) {
	h := newHarness(t, 8)
	made, issued := withTokens(t, h)
	key := aKey()
	endpoint(t, made, key)

	h.pass(20 * time.Minute)
	made.SweepNow()
	if _, revoked := issued.counts(key); revoked != 1 {
		t.Errorf("revoked %d times after sleeping, want once", revoked)
	}
}

func TestATokenStopsWorkingWhenAStartFails(t *testing.T) {
	h := newHarness(t, 8)
	h.backend.Fail = workspace.ErrNotReady
	made, issued := withTokens(t, h)
	key := aKey()

	made.Endpoint(t.Context(), key, "u")
	if issuedN, revoked := issued.counts(key); issuedN != 1 || revoked != 1 {
		t.Errorf("issued %d, revoked %d, want the token revoked after the failure", issuedN, revoked)
	}
}

func TestATokenThatCannotBeIssuedStopsTheStart(t *testing.T) {
	h := newHarness(t, 8)
	made, issued := withTokens(t, h)
	issued.fail = errors.New("no tokens today")

	if _, err := made.Endpoint(t.Context(), aKey(), "u"); err == nil {
		t.Errorf("a workspace started without its token")
	}
	if h.backend.StartCount() != 0 {
		t.Errorf("the backend was started")
	}
}

func TestReconcileRevokesTheTokensOfWhatIsGone(t *testing.T) {
	h := newHarness(t, 8)
	key := aKey()
	h.store.Put(workspace.Record{ID: uuid.New(), Key: key, Backend: "fake", State: workspace.Warm, Handle: "gone"})

	made, issued := withTokens(t, h)
	if err := made.Reconcile(t.Context()); err != nil {
		t.Fatal(err)
	}
	if _, revoked := issued.counts(key); revoked != 1 {
		t.Errorf("revoked %d times, want the gone workspace's token revoked", revoked)
	}
}

func TestTheAgentIsFoundWithItsSecret(t *testing.T) {
	h := newHarness(t, 8)
	made, _ := withTokens(t, h)
	key := aKey()

	endpoint, secret, err := made.Agent(t.Context(), key, "u")
	if err != nil || secret != "token-"+key.Person.String() || endpoint.Host == "" {
		t.Errorf("Agent() = %+v, %q, %v", endpoint, secret, err)
	}

	h.backend.Fail = workspace.ErrNotReady
	if _, _, err := made.Agent(t.Context(), aKey(), "u"); !errors.Is(err, workspace.ErrNotReady) {
		t.Errorf("Agent() of a workspace that does not start = %v", err)
	}
}

func TestAWarmWorkspaceIsFoundByItsViewKeyOnly(t *testing.T) {
	h := newHarness(t, 8)
	key := aKey()
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	if _, err := h.w.Endpoint(ctx, key, "https://preview.test/v/{view}/"); err != nil {
		t.Fatal(err)
	}
	under := h.backend.Started[0].Under
	view := strings.TrimSuffix(strings.TrimPrefix(under, "https://preview.test/v/"), "/")
	if view == "" || strings.Contains(view, "{") {
		t.Fatalf("the backend was given %q", under)
	}
	found, at, ok := h.w.ByView(view)
	if !ok || found != key || at.Host == "" {
		t.Errorf("ByView(%q) = %v, %+v, %v", view, found, at, ok)
	}
	for _, wrong := range []string{"", view + "x", view[1:], "{view}"} {
		if _, _, ok := h.w.ByView(wrong); ok {
			t.Errorf("ByView(%q) found a workspace", wrong)
		}
	}
	if report, _ := h.w.Report(key); report.Address != under {
		t.Errorf("Report().Address = %q, want %q", report.Address, under)
	}
}

func TestAWorkspaceOnTheAppsOwnAddressHasNoViewKey(t *testing.T) {
	h := newHarness(t, 8)
	endpoint(t, h.w, aKey())
	if _, _, ok := h.w.ByView(""); ok {
		t.Error("an empty key found a workspace")
	}
}

func TestAChangedRuntimeDefinitionStartsAfresh(t *testing.T) {
	h := newHarness(t, 8)
	h.copies.Root = t.TempDir()
	key := aKey()
	endpoint(t, h.w, key)
	endpoint(t, h.w, key)
	if len(h.backend.Started) != 1 {
		t.Fatalf("started %d times before any change", len(h.backend.Started))
	}

	dir, _ := h.copies.Dir(t.Context(), key)
	if err := os.MkdirAll(filepath.Join(dir, "quiqr"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "quiqr", "runtime.yaml"), []byte("packages: {hugo: 0.139.0}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	endpoint(t, h.w, key)
	endpoint(t, h.w, key)
	if len(h.backend.Started) != 2 || h.backend.Running() != 1 {
		t.Errorf("started %d, running %d, want 2 and 1", len(h.backend.Started), h.backend.Running())
	}
}

func TestTheEndpointSaysWhetherThePreviewIsRooted(t *testing.T) {
	h := newHarness(t, 8)
	made, err := workspace.New(h.backend, h.store, h.copies, h.isAllowed, workspace.Settings{
		RuntimeFor: func(context.Context, workspace.Key, string) (runtime.Resolved, error) {
			return runtime.Resolved{Preview: runtime.Command{Program: "/nix/store/q-quarto/bin/quarto"}, Rooted: true}, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { made.Close(context.Background()) })
	if got := endpoint(t, made, aKey()); !got.Rooted {
		t.Errorf("Endpoint() = %+v, want it rooted", got)
	}
}

func TestAPreviewThatIsNotHugoHasPagesOfItsOwn(t *testing.T) {
	h := newHarness(t, 8)
	made, err := workspace.New(h.backend, h.store, h.copies, h.isAllowed, workspace.Settings{
		RuntimeFor: func(context.Context, workspace.Key, string) (runtime.Resolved, error) {
			return runtime.Resolved{Preview: runtime.Command{Program: "/nix/store/q-quarto/bin/quarto"}, Rooted: true}, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { made.Close(context.Background()) })
	key := aKey()
	endpoint(t, made, key)
	if report, _ := made.Report(key); !report.OwnPages {
		t.Errorf("Report() = %+v", report)
	}
	endpoint(t, h.w, aKey())
}

func TestARootedPreviewIsAddressedTheWayQuartoExpects(t *testing.T) {
	h := newHarness(t, 8)
	made, err := workspace.New(h.backend, h.store, h.copies, h.isAllowed, workspace.Settings{
		RuntimeFor: func(context.Context, workspace.Key, string) (runtime.Resolved, error) {
			return runtime.Resolved{Preview: runtime.Command{Program: "/nix/store/q-quarto/bin/quarto"}, Rooted: true}, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { made.Close(context.Background()) })
	key := aKey()
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	if _, err := made.Endpoint(ctx, key, "https://preview.test/v/{view}/"); err != nil {
		t.Fatal(err)
	}
	if report, _ := made.Report(key); !strings.HasSuffix(report.Address, "/p/site/") {
		t.Errorf("Address = %q", report.Address)
	}
}
