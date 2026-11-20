package workspacetest

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/workspace"
)

type Backend struct {
	ApartFromApp bool
	Fail         error
	Hold         chan struct{}

	mu      sync.Mutex
	next    int
	running map[workspace.Handle]workspace.Spec
	Started []workspace.Spec
	Says    []string
	Stopped []workspace.Handle
}

func NewBackend() *Backend {
	return &Backend{running: map[workspace.Handle]workspace.Spec{}}
}

func (b *Backend) Name() string { return "fake" }

func (b *Backend) Apart() bool { return b.ApartFromApp }

func (b *Backend) Start(ctx context.Context, spec workspace.Spec, log io.Writer) (workspace.Handle, workspace.Endpoint, error) {
	if b.Hold != nil {
		select {
		case <-b.Hold:
		case <-ctx.Done():
			return "", workspace.Endpoint{}, ctx.Err()
		}
	}
	fmt.Fprintf(log, "starting %s\n", spec.Dir)
	for _, line := range b.Says {
		fmt.Fprintln(log, line)
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	b.Started = append(b.Started, spec)
	if b.Fail != nil {
		return "", workspace.Endpoint{}, b.Fail
	}
	b.next++
	handle := workspace.Handle(fmt.Sprintf("fake-%d", b.next))
	b.running[handle] = spec
	return handle, workspace.Endpoint{Host: "127.0.0.1", Port: 40000 + b.next}, nil
}

func (b *Backend) Alive(ctx context.Context, handle workspace.Handle) (workspace.Endpoint, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.running[handle]; !ok {
		return workspace.Endpoint{}, false
	}
	return workspace.Endpoint{Host: "127.0.0.1", Port: 39999}, true
}

func (b *Backend) Stop(ctx context.Context, handle workspace.Handle) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.running, handle)
	b.Stopped = append(b.Stopped, handle)
	return nil
}

func (b *Backend) Running() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.running)
}

func (b *Backend) Forget(handle workspace.Handle) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.running, handle)
}

func (b *Backend) StartCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.Started)
}

var _ workspace.Backend = (*Backend)(nil)

type Store struct {
	mu      sync.Mutex
	records map[workspace.Key]*workspace.Record
}

func NewStore() *Store {
	return &Store{records: map[workspace.Key]*workspace.Record{}}
}

func (s *Store) Reserve(ctx context.Context, key workspace.Key, backend string, most int, at time.Time) (workspace.Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	busy := 0
	for other, record := range s.records {
		if other != key && record.State != workspace.Sleeping {
			busy++
		}
	}
	if busy >= most {
		return workspace.Record{}, workspace.ErrTooMany
	}
	record, found := s.records[key]
	if !found {
		record = &workspace.Record{ID: uuid.New(), Key: key}
		s.records[key] = record
	}
	record.Backend = backend
	record.State = workspace.Starting
	record.Handle = ""
	record.Reason = ""
	record.LastUsed = at
	return *record, nil
}

func (s *Store) find(id uuid.UUID) (*workspace.Record, error) {
	for _, record := range s.records {
		if record.ID == id {
			return record, nil
		}
	}
	return nil, workspace.ErrNotFound
}

func (s *Store) Warm(ctx context.Context, id uuid.UUID, handle workspace.Handle) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, err := s.find(id)
	if err != nil {
		return err
	}
	record.State = workspace.Warm
	record.Handle = handle
	return nil
}

func (s *Store) Sleep(ctx context.Context, id uuid.UUID, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, err := s.find(id)
	if err != nil {
		return err
	}
	record.State = workspace.Sleeping
	record.Handle = ""
	record.Reason = reason
	return nil
}

func (s *Store) Touch(ctx context.Context, id uuid.UUID, at time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, err := s.find(id)
	if err != nil {
		return err
	}
	if record.LastUsed.Before(at) {
		record.LastUsed = at
	}
	return nil
}

func (s *Store) list(want func(workspace.State) bool) []workspace.Record {
	s.mu.Lock()
	defer s.mu.Unlock()
	var records []workspace.Record
	for _, record := range s.records {
		if want(record.State) {
			records = append(records, *record)
		}
	}
	return records
}

func (s *Store) Busy(ctx context.Context) ([]workspace.Record, error) {
	return s.list(func(state workspace.State) bool { return state != workspace.Sleeping }), nil
}

func (s *Store) Sleeping(ctx context.Context) ([]workspace.Record, error) {
	return s.list(func(state workspace.State) bool { return state == workspace.Sleeping }), nil
}

func (s *Store) Remove(ctx context.Context, id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, record := range s.records {
		if record.ID == id {
			delete(s.records, key)
		}
	}
	return nil
}

func (s *Store) Put(record workspace.Record) {
	s.mu.Lock()
	defer s.mu.Unlock()
	copied := record
	s.records[record.Key] = &copied
}

func (s *Store) Get(key workspace.Key) (workspace.Record, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, found := s.records[key]
	if !found {
		return workspace.Record{}, false
	}
	return *record, true
}

var _ workspace.Store = (*Store)(nil)

type Copies struct {
	mu      sync.Mutex
	Waiting map[workspace.Key]int
	Removed []workspace.Key
	Fail    error
	Root    string
}

func NewCopies() *Copies {
	return &Copies{Waiting: map[workspace.Key]int{}}
}

func (c *Copies) Dir(ctx context.Context, key workspace.Key) (string, error) {
	if c.Fail != nil {
		return "", c.Fail
	}
	return c.Root + "/sites/" + key.Project.String() + "/people/" + key.Person.String(), nil
}

func (c *Copies) Unshared(ctx context.Context, key workspace.Key) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Waiting[key], nil
}

func (c *Copies) Remove(ctx context.Context, key workspace.Key) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Removed = append(c.Removed, key)
	return nil
}

var _ workspace.Copies = (*Copies)(nil)
