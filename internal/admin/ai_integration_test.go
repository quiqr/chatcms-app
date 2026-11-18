//go:build integration

package admin_test

import (
	"errors"
	"testing"

	"github.com/quiqr/chatcms-app/internal/admin"
	"github.com/quiqr/chatcms-app/internal/aigateway"
	"github.com/quiqr/chatcms-app/internal/httpapi"
)

func TestOnlyAPlatformAdminSetsAndSeesAIUse(t *testing.T) {
	h := newHarness(t)
	service := admin.NewService(h.store, h.tenancy, h.identity)
	service.UseAIBooks(aigateway.NewBooks(aigateway.NewPostgres(h.db.Queries), nil))
	web := admin.NewWeb(service)

	boss := h.administrator(t, "boss@example.org")
	owner := h.person(t, "owner@example.org")
	if _, err := web.MakeOrganisation(t.Context(), boss, "bakery", "Bakery", owner.Email); err != nil {
		t.Fatal(err)
	}

	use, err := web.AIUse(t.Context(), boss, "bakery")
	if err != nil || use.RequestsPerMinute != nil || use.TokensPerMonth != nil || use.TokensThisMonth != 0 {
		t.Fatalf("AIUse() of a new organisation = %+v, %v, want no limits and nothing used", use, err)
	}

	rate, budget := 30, int64(2_000_000)
	if _, err := web.SetAILimits(t.Context(), owner, "bakery", &rate, &budget); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("SetAILimits() as the organisation's admin error = %v, want ErrNotAllowed", err)
	}
	if _, err := web.AIUse(t.Context(), owner, "bakery"); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("AIUse() as the organisation's admin error = %v, want ErrNotAllowed", err)
	}

	use, err = web.SetAILimits(t.Context(), boss, "bakery", &rate, &budget)
	if err != nil || *use.RequestsPerMinute != 30 || *use.TokensPerMonth != 2_000_000 {
		t.Fatalf("SetAILimits() = %+v, %v", use, err)
	}

	zero := 0
	_, err = web.SetAILimits(t.Context(), boss, "bakery", &zero, nil)
	var refused *httpapi.RefusedError
	if !errors.As(err, &refused) || refused.Key != "refused.limit-below-one" {
		t.Errorf("a limit of zero error = %v", err)
	}

	if _, err := web.AIUse(t.Context(), boss, "nobody"); err == nil {
		t.Errorf("AIUse() of an organisation that does not exist succeeded")
	}

	records, err := h.store.Records(t.Context(), 20)
	if err == nil {
		found := false
		for _, record := range records {
			if record.Did == admin.SetTheAILimits {
				found = true
			}
		}
		if !found {
			t.Errorf("setting the AI limits was not recorded")
		}
	}

	bare := admin.NewWeb(admin.NewService(h.store, h.tenancy, h.identity))
	if _, err := bare.AIUse(t.Context(), boss, "bakery"); !errors.Is(err, httpapi.ErrMissing) {
		t.Errorf("AIUse() without books error = %v, want ErrMissing", err)
	}
}
