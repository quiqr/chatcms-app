package preview

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/httpapi"
	"github.com/quiqr/chatcms-app/internal/policy"
	"github.com/quiqr/chatcms-app/internal/workspace"
)

type Sources interface {
	Resource(ctx context.Context, projectID uuid.UUID) (policy.Resource, error)
}

type Workspaces interface {
	Endpoint(ctx context.Context, key workspace.Key, under string) (workspace.Endpoint, error)
	Open(ctx context.Context, key workspace.Key, under string) (workspace.State, error)
	Report(key workspace.Key) (workspace.Report, bool)
	SleepProject(ctx context.Context, project string)
}

var ErrIncomplete = errors.New("preview: the service is missing a part it needs")

var (
	refusedTooMany = &httpapi.RefusedError{
		Key:    "refused.too-many-sites-being-previewed",
		Reason: "too many sites are being previewed at once; try again in a few minutes",
	}
	refusedNotReady = &httpapi.RefusedError{
		Key:    "refused.preview-did-not-start-site",
		Reason: "the preview did not start. Your site may have something in it Hugo cannot build.",
	}
)

type Service struct {
	sources    Sources
	workspaces Workspaces
	baseURL    string
	origin     string
	assets     fs.FS
}

func (s *Service) ServeFrom(origin string) {
	s.origin = strings.TrimSuffix(origin, "/")
}

func (s *Service) Apart() bool {
	return s.origin != ""
}

func NewService(sources Sources, workspaces Workspaces, baseURL string) (*Service, error) {
	if sources == nil || workspaces == nil {
		return nil, ErrIncomplete
	}
	return &Service{sources: sources, workspaces: workspaces, baseURL: strings.TrimSuffix(baseURL, "/")}, nil
}

func (s *Service) Under(under string) string {
	if s.origin != "" {
		return s.origin + "/v/" + workspace.ViewPlaceholder + "/"
	}
	return s.baseURL + under
}

func (s *Service) Handler(ctx context.Context, actor policy.Actor, projectID uuid.UUID, under string) (http.Handler, error) {
	resource, err := s.sources.Resource(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if !policy.Can(actor, policy.ActionReadContent, resource) {
		return nil, httpapi.ErrMissing
	}
	person, err := uuid.Parse(actor.PersonID)
	if err != nil {
		return nil, httpapi.ErrNotAllowed
	}

	endpoint, err := s.workspaces.Endpoint(ctx, workspace.Key{Person: person, Project: projectID}, s.Under(under))
	if err != nil {
		return nil, Refusal(err)
	}

	where, err := url.Parse("http://" + endpoint.Address())
	if err != nil {
		return nil, err
	}
	return s.decorate(where, under, endpoint.Rooted), nil
}

func Refusal(err error) error {
	switch {
	case errors.Is(err, workspace.ErrTooMany):
		return refusedTooMany
	case errors.Is(err, workspace.ErrNotReady):
		return refusedNotReady
	default:
		return err
	}
}

func (s *Service) Stop(projectID uuid.UUID) {
	s.workspaces.SleepProject(context.Background(), projectID.String())
}

type Standing struct {
	State    workspace.State
	Address  string
	Type     string
	Open     string
	OwnPages bool
	Reason   string
	Steps    []workspace.StepReport
	Log      []string
}

func (s *Service) Open(ctx context.Context, actor policy.Actor, projectID uuid.UUID, under string) (Standing, error) {
	key, technical, err := s.keyFor(ctx, actor, projectID)
	if err != nil {
		return Standing{}, err
	}
	if _, err := s.workspaces.Open(ctx, key, s.Under(under)); err != nil {
		return Standing{}, Refusal(err)
	}
	return s.standing(key, technical), nil
}

func (s *Service) Standing(ctx context.Context, actor policy.Actor, projectID uuid.UUID) (Standing, error) {
	key, technical, err := s.keyFor(ctx, actor, projectID)
	if err != nil {
		return Standing{}, err
	}
	return s.standing(key, technical), nil
}

func (s *Service) keyFor(ctx context.Context, actor policy.Actor, projectID uuid.UUID) (workspace.Key, bool, error) {
	resource, err := s.sources.Resource(ctx, projectID)
	if err != nil {
		return workspace.Key{}, false, err
	}
	if !policy.Can(actor, policy.ActionReadContent, resource) {
		return workspace.Key{}, false, httpapi.ErrMissing
	}
	person, err := uuid.Parse(actor.PersonID)
	if err != nil {
		return workspace.Key{}, false, httpapi.ErrNotAllowed
	}
	technical := policy.Can(actor, policy.ActionSeeTechnicalDetail, resource)
	return workspace.Key{Person: person, Project: projectID}, technical, nil
}

func (s *Service) standing(key workspace.Key, technical bool) Standing {
	report, _ := s.workspaces.Report(key)
	standing := Standing{State: report.State, Address: s.addressOf(report.Address), Type: report.Type, Open: report.Open, OwnPages: report.OwnPages}
	if report.Reason != "" {
		standing.Reason = refusedNotReady.Key
		var refused *httpapi.RefusedError
		if errors.As(report.Err, &refused) && refused.Key != "" {
			standing.Reason = refused.Key
		}
	}
	if technical {
		standing.Steps = report.Steps
		standing.Log = report.Log
	}
	return standing
}

type Viewed interface {
	ByView(view string) (workspace.Key, workspace.Endpoint, bool)
}

type Origin struct {
	workspaces Viewed
	allowed    func(ctx context.Context, key workspace.Key) (bool, error)
	service    *Service
}

func NewOrigin(service *Service, workspaces Viewed, allowed func(ctx context.Context, key workspace.Key) (bool, error)) *Origin {
	if service == nil {
		service = &Service{}
	}
	return &Origin{workspaces: workspaces, allowed: allowed, service: service}
}

func (o *Origin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rest, found := strings.CutPrefix(r.URL.Path, "/v/")
	view, _, _ := strings.Cut(rest, "/")
	if !found || view == "" {
		http.NotFound(w, r)
		return
	}
	key, endpoint, live := o.workspaces.ByView(view)
	if !live {
		http.NotFound(w, r)
		return
	}
	allowed, err := o.allowed(r.Context(), key)
	if err != nil {
		http.Error(w, "the preview cannot be checked right now", http.StatusServiceUnavailable)
		return
	}
	if !allowed {
		http.NotFound(w, r)
		return
	}
	where, err := url.Parse("http://" + endpoint.Address())
	if err != nil {
		http.NotFound(w, r)
		return
	}
	o.service.decorate(where, "/v/"+view+"/", endpoint.Rooted).ServeHTTP(w, r)
}

func (s *Service) addressOf(address string) string {
	if s.Apart() || address == "" {
		return address
	}
	parsed, err := url.Parse(address)
	if err != nil {
		return ""
	}
	return parsed.EscapedPath()
}
