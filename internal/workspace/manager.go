package workspace

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/quiqr/chatcms-app/internal/runtime"
)

const (
	DefaultMost   = 8
	DefaultIdle   = 15 * time.Minute
	DefaultForget = 30 * 24 * time.Hour

	touchEvery   = time.Minute
	sweepEvery   = time.Minute
	cleanupEvery = time.Hour
	startWithin  = 2 * time.Minute
)

type Settings struct {
	Most       int
	Idle       time.Duration
	Forget     time.Duration
	Logger     *slog.Logger
	RuntimeFor func(ctx context.Context, key Key, dir string) (runtime.Resolved, error)
	Tokens     Tokens
}

func (s Settings) withDefaults() Settings {
	if s.Most <= 0 {
		s.Most = DefaultMost
	}
	if s.Idle <= 0 {
		s.Idle = DefaultIdle
	}
	if s.Forget <= 0 {
		s.Forget = DefaultForget
	}
	if s.Logger == nil {
		s.Logger = slog.New(slog.DiscardHandler)
	}
	return s
}

type Report struct {
	State    State        `json:"state"`
	Reason   string       `json:"reason,omitempty"`
	Address  string       `json:"address,omitempty"`
	Type     string       `json:"type,omitempty"`
	Open     string       `json:"open,omitempty"`
	OwnPages bool         `json:"ownPages,omitempty"`
	Err      error        `json:"-"`
	Steps    []StepReport `json:"steps,omitempty"`
	Log      []string     `json:"log,omitempty"`
}

type live struct {
	record   Record
	spec     Spec
	state    State
	endpoint Endpoint
	err      error
	done     chan struct{}
	progress *progress
	used     time.Time
	written  time.Time
	under    string
	view     string
	made     string
}

type Workspaces struct {
	backend  Backend
	store    Store
	copies   Copies
	allowed  Allowed
	settings Settings

	mu      sync.Mutex
	now     func() time.Time
	live    map[Key]*live
	stop    chan struct{}
	loops   sync.WaitGroup
	running bool
	closed  bool
}

func New(backend Backend, store Store, copies Copies, allowed Allowed, settings Settings) (*Workspaces, error) {
	if backend == nil || store == nil || copies == nil || allowed == nil {
		return nil, ErrIncomplete
	}
	return &Workspaces{
		backend:  backend,
		store:    store,
		copies:   copies,
		allowed:  allowed,
		settings: settings.withDefaults(),
		now:      time.Now,
		live:     map[Key]*live{},
		stop:     make(chan struct{}),
	}, nil
}

func (w *Workspaces) Backend() string {
	return w.backend.Name()
}

func (w *Workspaces) Most() int {
	return w.settings.Most
}

func (w *Workspaces) clock() time.Time {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.now()
}

func (w *Workspaces) Reconcile(ctx context.Context) error {
	busy, err := w.store.Busy(ctx)
	if err != nil {
		return err
	}
	for _, record := range busy {
		if record.Backend == w.backend.Name() && record.State == Warm && record.Handle != "" {
			if endpoint, alive := w.backend.Alive(ctx, record.Handle); alive {
				w.adopt(record, endpoint)
				continue
			}
			_ = w.backend.Stop(ctx, record.Handle)
		}
		w.revoke(ctx, record.Key)
		if err := w.store.Sleep(ctx, record.ID, ""); err != nil {
			return err
		}
	}
	return nil
}

func (w *Workspaces) adopt(record Record, endpoint Endpoint) {
	w.mu.Lock()
	defer w.mu.Unlock()

	done := make(chan struct{})
	close(done)
	record.State = Warm
	w.live[record.Key] = &live{
		record:   record,
		spec:     Spec{Key: record.Key},
		state:    Warm,
		endpoint: endpoint,
		done:     done,
		progress: newProgress(),
		used:     w.now(),
		written:  w.now(),
	}
}

func (w *Workspaces) Run() {
	w.mu.Lock()
	if w.running || w.closed {
		w.mu.Unlock()
		return
	}
	w.running = true
	w.mu.Unlock()

	w.loops.Add(1)
	go func() {
		defer w.loops.Done()
		sweep := time.NewTicker(sweepEvery)
		cleanup := time.NewTicker(cleanupEvery)
		defer sweep.Stop()
		defer cleanup.Stop()
		for {
			select {
			case <-w.stop:
				return
			case <-sweep.C:
				w.SweepNow()
			case <-cleanup.C:
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
				if err := w.CleanupNow(ctx); err != nil {
					w.settings.Logger.Warn("cleaning up workspaces failed", "error", err)
				}
				cancel()
			}
		}
	}()
}

func (w *Workspaces) Open(ctx context.Context, key Key, under string) (State, error) {
	held, err := w.open(ctx, key, under)
	if err != nil {
		return "", err
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	return held.state, nil
}

func (w *Workspaces) Endpoint(ctx context.Context, key Key, under string) (Endpoint, error) {
	held, err := w.open(ctx, key, under)
	if err != nil {
		return Endpoint{}, err
	}

	select {
	case <-held.done:
	case <-ctx.Done():
		return Endpoint{}, ctx.Err()
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	if held.err != nil {
		return Endpoint{}, held.err
	}
	return held.endpoint, nil
}

func (w *Workspaces) open(ctx context.Context, key Key, under string) (*live, error) {
	w.mu.Lock()
	held, found := w.live[key]
	if found && held.state != Sleeping {
		if held.under == "" {
			held.under = under
			held.spec.Under = strings.ReplaceAll(under, ViewPlaceholder, held.view)
		}
		if held.under == under && !held.changed() {
			w.touchLocked(ctx, held)
			w.mu.Unlock()
			return held, nil
		}
		delete(w.live, key)
		w.mu.Unlock()
		<-held.done
		w.stopOne(ctx, held)
	} else {
		w.mu.Unlock()
	}

	dir, err := w.copies.Dir(ctx, key)
	if err != nil {
		return nil, err
	}
	return w.begin(ctx, Spec{Key: key, Dir: dir, Under: under})
}

func definitionIn(dir string) string {
	if dir == "" {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(runtime.File)))
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func (held *live) changed() bool {
	return held.state == Warm && held.spec.Dir != "" && definitionIn(held.spec.Dir) != held.made
}

func (w *Workspaces) begin(ctx context.Context, spec Spec) (*live, error) {
	now := w.clock()
	record, err := w.store.Reserve(ctx, spec.Key, w.backend.Name(), w.settings.Most, now)
	if err != nil {
		return nil, err
	}

	w.mu.Lock()
	if held, found := w.live[spec.Key]; found && held.state != Sleeping {
		w.mu.Unlock()
		return held, nil
	}
	view := randomKey()
	under := spec.Under
	spec.Under = strings.ReplaceAll(under, ViewPlaceholder, view)
	held := &live{
		record:   record,
		spec:     spec,
		state:    Starting,
		done:     make(chan struct{}),
		progress: newProgress(),
		used:     now,
		written:  now,
		under:    under,
		view:     view,
		made:     definitionIn(spec.Dir),
	}
	w.live[spec.Key] = held
	w.mu.Unlock()

	held.progress.begin(StepStarting)
	go w.start(held)
	return held, nil
}

func (w *Workspaces) start(held *live) {
	ctx, cancel := context.WithTimeout(context.Background(), startWithin)
	defer cancel()

	held.progress.begin(StepFetching)
	held.progress.logf("fetching " + held.spec.Dir)
	var err error
	if w.settings.RuntimeFor != nil {
		held.spec.Run, err = w.settings.RuntimeFor(ctx, held.spec.Key, held.spec.Dir)
		if held.spec.Run.Preview.Program != "" {
			held.progress.logf("preview with " + held.spec.Run.Preview.Program)
		}
	}
	if err == nil && w.settings.Tokens != nil {
		held.spec.Token, err = w.settings.Tokens.Issue(ctx, held.spec.Key)
	}
	var handle Handle
	var endpoint Endpoint
	if err == nil {
		held.progress.begin(StepBuilding)
		handle, endpoint, err = w.backend.Start(ctx, held.spec, held.progress)
	}

	if err == nil {
		err = w.store.Warm(ctx, held.record.ID, handle)
		if err != nil {
			_ = w.backend.Stop(ctx, handle)
		}
	}

	w.mu.Lock()
	if err != nil {
		held.state = Sleeping
		held.err = err
		held.record.Reason = err.Error()
	} else {
		held.state = Warm
		held.endpoint = endpoint
		held.endpoint.Rooted = held.spec.Run.Rooted
		held.record.Handle = handle
		held.record.State = Warm
	}
	w.mu.Unlock()

	held.progress.finish(err != nil)
	if err != nil {
		w.revoke(ctx, held.spec.Key)
		held.progress.logf(err.Error())
		if sleepErr := w.store.Sleep(ctx, held.record.ID, reasonFor(err)); sleepErr != nil {
			w.settings.Logger.Warn("recording a failed start failed", "error", sleepErr)
		}
	}
	close(held.done)
}

func reasonFor(err error) string {
	if errors.Is(err, ErrNotReady) {
		return "did-not-start"
	}
	return "failed"
}

func (w *Workspaces) touchLocked(ctx context.Context, held *live) {
	now := w.now()
	held.used = now
	if now.Sub(held.written) < touchEvery {
		return
	}
	held.written = now
	id := held.record.ID
	go func() {
		touchCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		if err := w.store.Touch(touchCtx, id, now); err != nil {
			w.settings.Logger.Warn("recording a workspace's use failed", "error", err)
		}
	}()
}

func (w *Workspaces) Report(key Key) (Report, bool) {
	w.mu.Lock()
	held, found := w.live[key]
	if !found {
		w.mu.Unlock()
		return Report{State: Sleeping}, false
	}
	program := held.spec.Run.Preview.Program
	report := Report{State: held.state, Err: held.err, Type: held.spec.Run.Type, Open: held.spec.Run.Open, OwnPages: program != "" && filepath.Base(program) != "hugo"}
	if held.state == Warm {
		report.Address = held.spec.Under
		if held.spec.Run.Rooted {
			report.Address += RootedPath
		}
	}
	if held.err != nil {
		report.Reason = reasonFor(held.err)
	}
	w.mu.Unlock()

	report.Steps, report.Log = held.progress.report()
	return report, true
}

func (w *Workspaces) Running() int {
	w.mu.Lock()
	defer w.mu.Unlock()

	count := 0
	for _, held := range w.live {
		if held.state != Sleeping {
			count++
		}
	}
	return count
}

func (w *Workspaces) Sleep(ctx context.Context, key Key) {
	w.mu.Lock()
	held, found := w.live[key]
	delete(w.live, key)
	w.mu.Unlock()

	if found {
		<-held.done
		w.stopOne(ctx, held)
	}
}

func (w *Workspaces) SleepProject(ctx context.Context, project string) {
	w.mu.Lock()
	var keys []Key
	for key := range w.live {
		if key.Project.String() == project {
			keys = append(keys, key)
		}
	}
	w.mu.Unlock()

	for _, key := range keys {
		w.Sleep(ctx, key)
	}
}

func (w *Workspaces) stopOne(ctx context.Context, held *live) {
	w.mu.Lock()
	handle := held.record.Handle
	state := held.state
	held.state = Sleeping
	w.mu.Unlock()

	if state != Warm {
		return
	}
	w.revoke(ctx, held.spec.Key)
	if err := w.backend.Stop(ctx, handle); err != nil {
		w.settings.Logger.Warn("stopping a workspace failed", "error", err)
	}
	if err := w.store.Sleep(ctx, held.record.ID, ""); err != nil {
		w.settings.Logger.Warn("recording a workspace asleep failed", "error", err)
	}
}

func (w *Workspaces) SweepNow() {
	w.mu.Lock()
	now := w.now()
	var idle []*live
	for key, held := range w.live {
		finished := isClosed(held.done)
		if finished && now.Sub(held.used) > w.settings.Idle {
			idle = append(idle, held)
			delete(w.live, key)
		}
	}
	w.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	for _, held := range idle {
		w.stopOne(ctx, held)
	}
}

func isClosed(done chan struct{}) bool {
	select {
	case <-done:
		return true
	default:
		return false
	}
}

func (w *Workspaces) CleanupNow(ctx context.Context) error {
	sleeping, err := w.store.Sleeping(ctx)
	if err != nil {
		return err
	}
	now := w.clock()

	for _, record := range sleeping {
		w.mu.Lock()
		_, inUse := w.live[record.Key]
		w.mu.Unlock()
		if inUse {
			continue
		}

		allowed, err := w.allowed(ctx, record.Key)
		if err != nil {
			return err
		}
		forget := !allowed
		if allowed && now.Sub(record.LastUsed) > w.settings.Forget {
			unshared, err := w.copies.Unshared(ctx, record.Key)
			if err != nil {
				return err
			}
			forget = unshared == 0
		}
		if !forget {
			continue
		}

		if err := w.copies.Remove(ctx, record.Key); err != nil {
			return fmt.Errorf("removing a workspace's copy failed: %w", err)
		}
		if err := w.store.Remove(ctx, record.ID); err != nil {
			return err
		}
	}
	return nil
}

func (w *Workspaces) Close(ctx context.Context) {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return
	}
	w.closed = true
	close(w.stop)
	held := make([]*live, 0, len(w.live))
	if !w.backend.Apart() {
		for key, one := range w.live {
			held = append(held, one)
			delete(w.live, key)
		}
	}
	w.mu.Unlock()

	w.loops.Wait()
	for _, one := range held {
		select {
		case <-one.done:
		case <-ctx.Done():
			continue
		}
		w.stopOne(ctx, one)
	}
}

func (w *Workspaces) PretendTimeIs(now func() time.Time) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.now = now
}

func (w *Workspaces) revoke(ctx context.Context, key Key) {
	if w.settings.Tokens == nil {
		return
	}
	if err := w.settings.Tokens.Revoke(ctx, key); err != nil {
		w.settings.Logger.Warn("revoking a workspace's token failed", "error", err)
	}
}

func (w *Workspaces) Agent(ctx context.Context, key Key, under string) (Endpoint, string, error) {
	endpoint, err := w.Endpoint(ctx, key, under)
	if err != nil {
		return Endpoint{}, "", err
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	held, found := w.live[key]
	if !found {
		return Endpoint{}, "", ErrNotFound
	}
	return endpoint, held.spec.Token, nil
}

const ViewPlaceholder = "{view}"

func randomKey() string {
	raw := make([]byte, 32)
	rand.Read(raw)
	return base64.RawURLEncoding.EncodeToString(raw)
}

func (w *Workspaces) ByView(view string) (Key, Endpoint, bool) {
	if view == "" {
		return Key{}, Endpoint{}, false
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	for key, held := range w.live {
		if held.state == Warm && held.view != "" && subtle.ConstantTimeCompare([]byte(held.view), []byte(view)) == 1 && strings.Contains(held.under, ViewPlaceholder) {
			w.touchLocked(context.Background(), held)
			return key, held.endpoint, true
		}
	}
	return Key{}, Endpoint{}, false
}
