package admin

import (
	"context"
	"errors"
	"strconv"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/aigateway"
	"github.com/quiqr/chatcms-app/internal/httpapi"
	"github.com/quiqr/chatcms-app/internal/policy"
)

type AIBooks interface {
	Use(ctx context.Context, organisation uuid.UUID) (aigateway.Limits, int64, int64, error)
	SetLimits(ctx context.Context, organisation uuid.UUID, limits aigateway.Limits) error
}

func (s *Service) UseAIBooks(books AIBooks) {
	s.ai = books
}

func (s *Service) AIUse(ctx context.Context, who httpapi.Actor, slug string) (httpapi.AIUseView, error) {
	if _, err := s.mayAdminister(ctx, who, policy.ActionSetLimits); err != nil {
		return httpapi.AIUseView{}, err
	}
	if s.ai == nil {
		return httpapi.AIUseView{}, httpapi.ErrMissing
	}
	organisation, err := s.tenancy.OrganisationBySlug(ctx, slug)
	if err != nil {
		return httpapi.AIUseView{}, err
	}
	limits, requests, tokens, err := s.ai.Use(ctx, organisation.ID)
	if errors.Is(err, aigateway.ErrNotFound) {
		return httpapi.AIUseView{}, httpapi.ErrMissing
	}
	if err != nil {
		return httpapi.AIUseView{}, err
	}
	return httpapi.AIUseView{
		RequestsPerMinute: limits.RequestsPerMinute,
		TokensPerMonth:    limits.TokensPerMonth,
		RequestsThisMonth: requests,
		TokensThisMonth:   tokens,
	}, nil
}

func (s *Service) SetAILimits(ctx context.Context, who httpapi.Actor, slug string, requests *int, tokens *int64) (httpapi.AIUseView, error) {
	if _, err := s.mayAdminister(ctx, who, policy.ActionSetLimits); err != nil {
		return httpapi.AIUseView{}, err
	}
	if s.ai == nil {
		return httpapi.AIUseView{}, httpapi.ErrMissing
	}
	organisation, err := s.tenancy.OrganisationBySlug(ctx, slug)
	if err != nil {
		return httpapi.AIUseView{}, err
	}
	err = s.ai.SetLimits(ctx, organisation.ID, aigateway.Limits{RequestsPerMinute: requests, TokensPerMonth: tokens})
	if errors.Is(err, aigateway.ErrBelowOne) {
		return httpapi.AIUseView{}, &httpapi.RefusedError{
			Key:    "refused.limit-below-one",
			Reason: "a limit is a number of at least one, or none at all",
		}
	}
	if err != nil {
		return httpapi.AIUseView{}, err
	}
	s.record(ctx, who, SetTheAILimits, organisation.ID.String(), organisation.Name, map[string]any{
		"requests a minute": said(requests),
		"tokens a month":    saidBig(tokens),
	})
	return s.AIUse(ctx, who, slug)
}

func saidBig(limit *int64) string {
	if limit == nil {
		return "no limit"
	}
	return strconv.FormatInt(*limit, 10)
}

func (w *Web) AIUse(ctx context.Context, who httpapi.Actor, slug string) (httpapi.AIUseView, error) {
	return w.service.AIUse(ctx, who, slug)
}

func (w *Web) SetAILimits(ctx context.Context, who httpapi.Actor, slug string, requests *int, tokens *int64) (httpapi.AIUseView, error) {
	return w.service.SetAILimits(ctx, who, slug, requests, tokens)
}

var _ httpapi.AIUseService = (*Web)(nil)
