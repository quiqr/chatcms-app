package onboarding

import (
	"context"
	"errors"
	"net/url"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/quiqr/chatcms-app/internal/httpapi"
	"github.com/quiqr/chatcms-app/internal/policy"
	"github.com/quiqr/chatcms-app/internal/store"
	"github.com/quiqr/chatcms-app/internal/tenancy"
)

type Hugo struct {
	Version  string
	Extended bool
}

var hugoVersionOutput = regexp.MustCompile(`v([0-9]+\.[0-9]+(?:\.[0-9]+)?)\S*`)

func ParseHugoVersion(output string) Hugo {
	match := hugoVersionOutput.FindStringSubmatch(output)
	if match == nil {
		return Hugo{}
	}
	return Hugo{Version: match[1], Extended: regexp.MustCompile(`\+extended\b`).MatchString(match[0])}
}

func (s *Service) UseTemplates(catalogue *Catalogue, installed Hugo) {
	s.templates = catalogue
	s.hugo = installed
}

func (s *Service) mayCreateProjectsIn(ctx context.Context, who httpapi.Actor, organisationSlug string) error {
	if who.Anonymous || who.ID == "" {
		return httpapi.ErrNotAllowed
	}
	personID, err := uuid.Parse(who.ID)
	if err != nil {
		return httpapi.ErrNotAllowed
	}
	actor, err := s.tenancy.ActorFor(ctx, personID)
	if err != nil {
		return err
	}
	organisation, err := s.tenancy.OrganisationBySlug(ctx, organisationSlug)
	if errors.Is(err, tenancy.ErrNotFound) {
		return httpapi.ErrMissing
	}
	if err != nil {
		return err
	}

	resource := policy.Organisation(organisation.ID.String())
	if policy.Can(actor, policy.ActionCreateProject, resource) {
		return nil
	}
	_, member := actor.Organisations[organisation.ID.String()]
	if member || policy.Can(actor, policy.ActionSeeEveryOrganisation, policy.Resource{}) {
		return httpapi.ErrNotAllowed
	}
	projects, err := s.tenancy.ProjectsIn(ctx, organisation.ID)
	if err != nil {
		return err
	}
	for _, project := range projects {
		if _, holds := actor.Projects[project.ID.String()]; holds {
			return httpapi.ErrNotAllowed
		}
	}
	return httpapi.ErrMissing
}

func (s *Service) ListTemplates(ctx context.Context, who httpapi.Actor, organisation string) (httpapi.TemplatesView, error) {
	if err := s.mayCreateProjectsIn(ctx, who, organisation); err != nil {
		return httpapi.TemplatesView{}, err
	}

	view := httpapi.TemplatesView{
		Offered:   s.templates.Enabled(),
		Templates: []httpapi.TemplateView{},
		Hugo:      httpapi.HugoView{Version: s.hugo.Version, Extended: s.hugo.Extended},
	}
	if !view.Offered {
		return view, nil
	}

	templates, err := s.templates.List(ctx)
	if errors.Is(err, ErrGalleryUnavailable) {
		return view, nil
	}
	if err != nil {
		return httpapi.TemplatesView{}, err
	}

	view.Available = true
	for _, template := range templates {
		view.Templates = append(view.Templates, templateView(organisation, template))
	}
	return view, nil
}

func (s *Service) TemplateScreenshot(ctx context.Context, who httpapi.Actor, organisation, id string) ([]byte, string, error) {
	if err := s.mayCreateProjectsIn(ctx, who, organisation); err != nil {
		return nil, "", err
	}
	data, kind, err := s.templates.Screenshot(ctx, id)
	if errors.Is(err, ErrNoSuchTemplate) || errors.Is(err, ErrNoScreenshot) ||
		errors.Is(err, ErrGalleryOff) || errors.Is(err, ErrGalleryUnavailable) {
		return nil, "", httpapi.ErrMissing
	}
	return data, kind, err
}

func templateView(organisation string, template Template) httpapi.TemplateView {
	screenshot := ""
	if template.ScreenshotType != "" {
		screenshot = "/api/organisations/" + url.PathEscape(organisation) +
			"/templates/" + url.PathEscape(template.ID) + "/screenshot"
	}
	return httpapi.TemplateView{
		ID:            template.ID,
		Name:          template.Name,
		Description:   template.Description,
		Author:        template.Author,
		AuthorURL:     template.AuthorURL,
		License:       template.License,
		LicenseURL:    template.LicenseURL,
		Homepage:      template.Homepage,
		DemoURL:       template.DemoURL,
		HugoTheme:     template.HugoTheme,
		HugoVersion:   template.HugoVersion,
		Extended:      template.Extended,
		Model:         append([]string{}, template.Model...),
		Forms:         template.Forms,
		SourceURL:     template.SourceURL,
		ScreenshotURL: screenshot,
	}
}

var ErrUnusableHugoVersion = errors.New("a Hugo version is written as numbers, such as 0.101.0")

func (s *Service) ReadHugo(ctx context.Context, who httpapi.Actor, organisation, project string) (httpapi.ProjectHugoView, error) {
	found, err := s.reach(ctx, who, organisation, project, policy.ActionChangeSettings)
	if err != nil {
		return httpapi.ProjectHugoView{}, err
	}
	return s.hugoView(ctx, found.ID)
}

func (s *Service) SetHugo(ctx context.Context, who httpapi.Actor, organisation, project, version string, extended bool) (httpapi.ProjectHugoView, error) {
	found, err := s.reach(ctx, who, organisation, project, policy.ActionChangeSettings)
	if err != nil {
		return httpapi.ProjectHugoView{}, err
	}

	version = strings.TrimPrefix(strings.TrimSpace(version), "v")
	if version != "" && !hugoVersion.MatchString(version) {
		return httpapi.ProjectHugoView{}, &httpapi.RefusedError{
			Key: "refused.unusable-hugo-version", Reason: ErrUnusableHugoVersion.Error(),
		}
	}
	if err := s.db.Queries.SetProjectHugo(ctx, store.SetProjectHugoParams{
		ID:           pgtype.UUID{Bytes: found.ID, Valid: true},
		HugoVersion:  version,
		HugoExtended: extended && version != "",
	}); err != nil {
		return httpapi.ProjectHugoView{}, err
	}
	return s.hugoView(ctx, found.ID)
}

func (s *Service) hugoView(ctx context.Context, projectID uuid.UUID) (httpapi.ProjectHugoView, error) {
	row, err := s.db.Queries.GetProject(ctx, pgtype.UUID{Bytes: projectID, Valid: true})
	if err != nil {
		return httpapi.ProjectHugoView{}, err
	}
	return httpapi.ProjectHugoView{
		Version:   row.HugoVersion,
		Extended:  row.HugoExtended,
		Installed: httpapi.HugoView{Version: s.hugo.Version, Extended: s.hugo.Extended},
	}, nil
}
