package onboarding

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/quiqr/chatcms-app/internal/gitsync"
	"github.com/quiqr/chatcms-app/internal/store"
)

// SourceKind is where a site came from.
type SourceKind string

const (
	SourceNone       SourceKind = "none"
	SourceRepository SourceKind = "repository"
	SourceExport     SourceKind = "export"
	SourceTemplate   SourceKind = "template"
)

// PublishTo is where publishing sends work.
type PublishTo string

const (
	PublishToMainLine   PublishTo = "main-line"
	PublishToItsOwnLine PublishTo = "its-own-line"
)

// PublishForm is what publishing sends.
type PublishForm string

const (
	PublishSource PublishForm = "source"
	PublishBuilt  PublishForm = "built"
)

// DefaultPublishBranch is what the designs offer when a site publishes to a
// line of its own.
const DefaultPublishBranch = "quiqr/publish"

var (
	// ErrUnusableBranchName says a remote would not accept the name.
	ErrUnusableBranchName = errors.New(
		"a name may hold letters, digits, hyphens, underscores and slashes, and may not begin or end with one")

	// ErrNoSuchTemplate says the template is not one ChatCMS knows.
	ErrNoSuchTemplate = errors.New("that is not a template ChatCMS knows")

	// ErrExportUnreadable says the directory given as an export cannot be
	// read.
	ErrExportUnreadable = errors.New("that directory could not be read")

	ErrRepositoryNotEmpty = errors.New("that repository is not empty; a template needs an empty one")
	ErrNeedsARepository   = errors.New("a template needs an empty repository to be copied into")
)

// branchShape is deliberately stricter than git, because a name that works
// everywhere is worth more than one that works here.
var branchShape = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/-]*[A-Za-z0-9]$`)

// Importer brings a site into a project.
type Importer struct {
	db        *store.DB
	sites     *gitsync.Sites
	keys      *Keys
	who       gitsync.Identity
	templates *Catalogue
}

func NewImporter(db *store.DB, sites *gitsync.Sites, keys *Keys, who gitsync.Identity) *Importer {
	return &Importer{db: db, sites: sites, keys: keys, who: who}
}

func (i *Importer) UseTemplates(catalogue *Catalogue) {
	i.templates = catalogue
}

// Request says what to import and from where.
type Request struct {
	Kind SourceKind
	// Address is the repository, the directory holding an export, or the
	// template's name.
	Address string
	// Remote is where the project's work is sent. For a repository it is
	// usually the same as the address.
	Remote string
}

// Result is what happened, and what an implementer should do next.
type Result struct {
	Inspection Inspection `json:"inspection"`
	Remote     string     `json:"remote"`
	Kind       SourceKind `json:"kind"`
}

// Import brings a site in and says what arrived.
func (i *Importer) Import(ctx context.Context, projectID uuid.UUID, request Request) (Result, error) {
	siteID := projectID.String()

	remote := strings.TrimSpace(request.Remote)
	address := strings.TrimSpace(request.Address)

	var repository gitsync.Repo
	var err error

	switch request.Kind {
	case SourceRepository:
		if remote == "" {
			remote = address
		}
		repository, err = i.sites.Clone(ctx, siteID, address)

	case SourceTemplate:
		repository, err = i.startFromTemplate(ctx, projectID, remote, address)

	case SourceExport:
		content, openErr := exportAt(address)
		if openErr != nil {
			return Result{}, openErr
		}
		repository, err = i.sites.Create(ctx, siteID, remote, content, i.who,
			"Import the site from Quiqr Desktop")

	default:
		return Result{}, fmt.Errorf("%q is not a way to import a site", request.Kind)
	}

	if err != nil {
		return Result{}, err
	}
	defer repository.Close()

	if err := i.db.Queries.SetProjectSource(ctx, store.SetProjectSourceParams{
		ID:            pgtype.UUID{Bytes: projectID, Valid: true},
		Remote:        remote,
		SourceKind:    string(request.Kind),
		SourceAddress: address,
	}); err != nil {
		return Result{}, fmt.Errorf("recording where the site came from failed: %w", err)
	}

	return Result{
		Inspection: Inspect(repository.FS()),
		Remote:     remote,
		Kind:       request.Kind,
	}, nil
}

func (i *Importer) startFromTemplate(ctx context.Context, projectID uuid.UUID, remote, id string) (gitsync.Repo, error) {
	template, err := i.templates.Find(ctx, id)
	if err != nil {
		return nil, err
	}
	if remote == "" {
		return nil, ErrNeedsARepository
	}

	siteID := projectID.String()
	empty, err := i.sites.RemoteIsEmpty(ctx, siteID, remote)
	if err != nil {
		return nil, err
	}
	if !empty {
		return nil, ErrRepositoryNotEmpty
	}

	scratch, err := os.MkdirTemp("", "chatcms-template-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(scratch)

	files := filepath.Join(scratch, "files")
	if err := i.sites.Snapshot(ctx, template.SourceURL, files); err != nil {
		return nil, fmt.Errorf("fetching the template %s failed: %w", template.Name, err)
	}

	repository, err := i.sites.Create(ctx, siteID, remote, os.DirFS(files), i.who, "Start from "+template.Name)
	if err != nil {
		return nil, err
	}

	if _, err := repository.Publish(ctx); err != nil {
		repository.Close()
		if dir, dirErr := i.sites.Dir(siteID); dirErr == nil {
			_ = os.RemoveAll(dir)
		}
		return nil, fmt.Errorf("sending the template to the repository failed: %w", err)
	}

	if err := i.db.Queries.SetProjectHugo(ctx, store.SetProjectHugoParams{
		ID:           pgtype.UUID{Bytes: projectID, Valid: true},
		HugoVersion:  template.HugoVersion,
		HugoExtended: template.Extended,
	}); err != nil {
		repository.Close()
		return nil, err
	}
	return repository, nil
}

// Inspect says what a project's working copy holds, without importing.
func (i *Importer) Inspect(ctx context.Context, projectID uuid.UUID) (Inspection, error) {
	repository, err := i.sites.Open(ctx, projectID.String())
	if err != nil {
		return Inspection{}, err
	}
	defer repository.Close()

	return Inspect(repository.FS()), nil
}

// Connection is what a test of the key against the remote found.
type Connection struct {
	// Good says the remote took the key.
	Good bool `json:"good"`
	// Message says what to do next, in words an implementer can act on.
	Message string `json:"message"`
	// PublicKey is shown again when the key still has to be added.
	PublicKey string `json:"publicKey,omitempty"`
}

// TestConnection asks whether the project's key reaches its remote.
func (i *Importer) TestConnection(ctx context.Context, projectID uuid.UUID, remote string) (Connection, error) {
	if strings.TrimSpace(remote) == "" {
		return Connection{
			Message: "This project has no remote yet, so there is nothing to reach.",
		}, nil
	}

	// Whether there is a key at all is asked first, so that a remote which
	// refuses for want of one is not reported as refusing the key.
	hasKey, err := i.keys.Has(ctx, projectID)
	if err != nil {
		return Connection{}, err
	}
	if !hasKey {
		return Connection{
			Message: "This project has no key yet. Make one, then add it to the git host.",
		}, nil
	}

	err = i.sites.Reachable(ctx, projectID.String(), remote)

	switch {
	case err == nil:
		return Connection{Good: true, Message: "The connection is good."}, nil

	case errors.Is(err, ErrNoKeyYet):
		return Connection{
			Message: "This project has no key yet. Make one, then add it to the git host.",
		}, nil

	case errors.Is(err, gitsync.ErrKeyRefused):
		public, keyErr := i.keys.PublicKey(ctx, projectID)
		if keyErr != nil {
			public = ""
		}
		return Connection{
			Message: "The host answered but would not accept this key. Add the key below to " +
				"the repository's deploy keys, with write access, and test again.",
			PublicKey: public,
		}, nil

	default:
		return Connection{
			Message: "The host could not be reached. Check the address, then test again.",
		}, nil
	}
}

// Publishing is where and in what form a project publishes.
type Publishing struct {
	To     PublishTo   `json:"to"`
	Branch string      `json:"branch,omitempty"`
	Form   PublishForm `json:"form"`
}

// SetPublishing records where and in what form a project publishes.
func (i *Importer) SetPublishing(ctx context.Context, projectID uuid.UUID, publishing Publishing) error {
	branch := strings.TrimSpace(publishing.Branch)

	switch publishing.To {
	case PublishToMainLine:
		branch = ""
	case PublishToItsOwnLine:
		if branch == "" {
			branch = DefaultPublishBranch
		}
		if !branchShape.MatchString(branch) || strings.Contains(branch, "//") {
			return fmt.Errorf("%w: %q", ErrUnusableBranchName, branch)
		}
	default:
		return fmt.Errorf("%q is not a place to publish to", publishing.To)
	}

	switch publishing.Form {
	case PublishSource, PublishBuilt:
	default:
		return fmt.Errorf("%q is not a form to publish in", publishing.Form)
	}

	return i.db.Queries.SetProjectPublishing(ctx, store.SetProjectPublishingParams{
		ID:            pgtype.UUID{Bytes: projectID, Valid: true},
		PublishTo:     string(publishing.To),
		PublishBranch: branch,
		PublishForm:   string(publishing.Form),
	})
}

// exportAt reads a directory holding a site exported from the Quiqr desktop
// app.
func exportAt(path string) (fs.FS, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("%w: no directory was named", ErrExportUnreadable)
	}

	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("%w: %s", ErrExportUnreadable, path)
	}
	return os.DirFS(path), nil
}
