package artifacts

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/workflows"
)

const (
	DefaultMost  = 200 << 20
	DefaultLimit = 100
)

var (
	ErrMissing    = errors.New("artifacts: no such artifact")
	ErrIncomplete = errors.New("artifacts: missing a part it needs")
)

type Artifact struct {
	ID        uuid.UUID  `json:"id"`
	ProjectID uuid.UUID  `json:"-"`
	PublishID *uuid.UUID `json:"publishId,omitempty"`
	Name      string     `json:"name"`
	Size      int64      `json:"size"`
	SHA256    string     `json:"sha256"`
	CreatedAt time.Time  `json:"createdAt"`
}

type Records interface {
	Insert(ctx context.Context, artifact Artifact) error
	List(ctx context.Context, projectID uuid.UUID, limit int32) ([]Artifact, error)
	Get(ctx context.Context, projectID, id uuid.UUID) (Artifact, error)
	OfPublish(ctx context.Context, publishID uuid.UUID) ([]Artifact, error)
	Any(ctx context.Context, projectID uuid.UUID) (bool, error)
}

type Store struct {
	records Records
	dir     string
	most    int64
	now     func() time.Time
}

func New(records Records, dir string, most int64) (*Store, error) {
	if records == nil || dir == "" {
		return nil, ErrIncomplete
	}
	if most <= 0 {
		most = DefaultMost
	}
	return &Store{records: records, dir: dir, most: most, now: time.Now}, nil
}

func (s *Store) fileOf(projectID, id uuid.UUID) string {
	return filepath.Join(s.dir, projectID.String(), id.String())
}

func (s *Store) Keep(ctx context.Context, run workflows.Run, files fs.FS, names []string) ([]workflows.Kept, error) {
	for _, name := range names {
		info, err := fs.Stat(files, name)
		if err != nil {
			return nil, err
		}
		if info.Size() > s.most {
			return nil, fmt.Errorf("%w: %s is %d bytes", workflows.ErrTooBig, name, info.Size())
		}
	}
	if err := os.MkdirAll(filepath.Join(s.dir, run.ProjectID.String()), 0o750); err != nil {
		return nil, err
	}
	kept := make([]workflows.Kept, 0, len(names))
	for _, name := range names {
		artifact, err := s.keepOne(ctx, run, files, name)
		if err != nil {
			return nil, err
		}
		kept = append(kept, workflows.Kept{ID: artifact.ID, Name: artifact.Name, Size: artifact.Size})
	}
	return kept, nil
}

func (s *Store) keepOne(ctx context.Context, run workflows.Run, files fs.FS, name string) (Artifact, error) {
	from, err := files.Open(name)
	if err != nil {
		return Artifact{}, err
	}
	defer from.Close()

	id := uuid.New()
	final := s.fileOf(run.ProjectID, id)
	to, err := os.CreateTemp(filepath.Dir(final), ".keeping-*")
	if err != nil {
		return Artifact{}, err
	}
	defer os.Remove(to.Name())

	sum := sha256.New()
	size, err := io.Copy(io.MultiWriter(to, sum), io.LimitReader(from, s.most+1))
	if closeErr := to.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return Artifact{}, err
	}
	if size > s.most {
		return Artifact{}, fmt.Errorf("%w: %s", workflows.ErrTooBig, name)
	}
	if err := os.Rename(to.Name(), final); err != nil {
		return Artifact{}, err
	}

	artifact := Artifact{ID: id, ProjectID: run.ProjectID, Name: path.Clean(name), Size: size, SHA256: hex.EncodeToString(sum.Sum(nil)), CreatedAt: s.now()}
	if run.PublishID != uuid.Nil {
		publish := run.PublishID
		artifact.PublishID = &publish
	}
	if err := s.records.Insert(ctx, artifact); err != nil {
		os.Remove(final)
		return Artifact{}, err
	}
	return artifact, nil
}

func (s *Store) List(ctx context.Context, projectID uuid.UUID) ([]Artifact, error) {
	return s.records.List(ctx, projectID, DefaultLimit)
}

func (s *Store) OfPublish(ctx context.Context, publishID uuid.UUID) ([]Artifact, error) {
	return s.records.OfPublish(ctx, publishID)
}

func (s *Store) Any(ctx context.Context, projectID uuid.UUID) (bool, error) {
	return s.records.Any(ctx, projectID)
}

func (s *Store) Open(ctx context.Context, projectID, id uuid.UUID) (Artifact, *os.File, error) {
	artifact, err := s.records.Get(ctx, projectID, id)
	if err != nil {
		return Artifact{}, nil, err
	}
	file, err := os.Open(s.fileOf(projectID, id))
	if errors.Is(err, fs.ErrNotExist) {
		return Artifact{}, nil, ErrMissing
	}
	if err != nil {
		return Artifact{}, nil, err
	}
	return artifact, file, nil
}

var _ workflows.Keeper = (*Store)(nil)

type OfPublish struct {
	Store *Store
}

func (o OfPublish) OfPublish(ctx context.Context, publishID uuid.UUID) ([]workflows.Kept, error) {
	list, err := o.Store.OfPublish(ctx, publishID)
	if err != nil {
		return nil, err
	}
	kept := make([]workflows.Kept, 0, len(list))
	for _, artifact := range list {
		kept = append(kept, workflows.Kept{ID: artifact.ID, Name: artifact.Name, Size: artifact.Size})
	}
	return kept, nil
}
