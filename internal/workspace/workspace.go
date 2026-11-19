package workspace

import (
	"context"
	"errors"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/runtime"
)

var (
	ErrTooMany    = errors.New("workspace: too many workspaces are running at once")
	ErrNotReady   = errors.New("workspace: the workspace did not start")
	ErrIncomplete = errors.New("workspace: the manager is missing a part it needs")
	ErrNotFound   = errors.New("workspace: no such workspace")
)

type Key struct {
	Person  uuid.UUID
	Project uuid.UUID
}

type State string

const (
	Starting State = "starting"
	Warm     State = "warm"
	Sleeping State = "sleeping"
)

type Spec struct {
	Key   Key
	Dir   string
	Under string
	Run   runtime.Resolved
	Token string
}

type Endpoint struct {
	Host      string
	Port      int
	AgentPort int
	Rooted    bool
}

func (e Endpoint) Address() string {
	return net.JoinHostPort(e.Host, strconv.Itoa(e.Port))
}

func (e Endpoint) Preview() Endpoint {
	return Endpoint{Host: e.Host, Port: e.Port}
}

type Handle string

type Backend interface {
	Name() string
	Start(ctx context.Context, spec Spec, log io.Writer) (Handle, Endpoint, error)
	Alive(ctx context.Context, handle Handle) (Endpoint, bool)
	Stop(ctx context.Context, handle Handle) error
	Apart() bool
}

type Record struct {
	ID       uuid.UUID
	Key      Key
	Backend  string
	State    State
	Handle   Handle
	Reason   string
	LastUsed time.Time
}

type Store interface {
	Reserve(ctx context.Context, key Key, backend string, most int, at time.Time) (Record, error)
	Warm(ctx context.Context, id uuid.UUID, handle Handle) error
	Sleep(ctx context.Context, id uuid.UUID, reason string) error
	Touch(ctx context.Context, id uuid.UUID, at time.Time) error
	Busy(ctx context.Context) ([]Record, error)
	Sleeping(ctx context.Context) ([]Record, error)
	Remove(ctx context.Context, id uuid.UUID) error
}

type Copies interface {
	Dir(ctx context.Context, key Key) (string, error)
	Unshared(ctx context.Context, key Key) (int, error)
	Remove(ctx context.Context, key Key) error
}

type Allowed func(ctx context.Context, key Key) (bool, error)

type Tokens interface {
	Issue(ctx context.Context, key Key) (string, error)
	Revoke(ctx context.Context, key Key) error
}
