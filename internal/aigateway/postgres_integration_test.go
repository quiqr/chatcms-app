//go:build integration

package aigateway_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/aigateway"
	"github.com/quiqr/chatcms-app/internal/store"
	"github.com/quiqr/chatcms-app/internal/testsupport"
	"github.com/quiqr/chatcms-app/internal/workspace"
)

func TestPostgresKeepsTokensUseAndLimits(t *testing.T) {
	db, err := store.Open(t.Context(), testsupport.Postgres(t))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Migrate(t.Context()); err != nil {
		t.Fatal(err)
	}

	organisation, project, person := uuid.New(), uuid.New(), uuid.New()
	for _, statement := range []struct {
		sql  string
		args []any
	}{
		{`INSERT INTO organisations (id, slug, name) VALUES ($1, $2, 'Studio')`, []any{organisation, "studio-" + organisation.String()[:8]}},
		{`INSERT INTO projects (id, organisation_id, slug, name) VALUES ($1, $2, 'site', 'Site')`, []any{project, organisation}},
		{`INSERT INTO people (id, email) VALUES ($1, $2)`, []any{person, person.String() + "@example.org"}},
	} {
		if _, err := db.Pool().Exec(t.Context(), statement.sql, statement.args...); err != nil {
			t.Fatal(err)
		}
	}

	books := aigateway.NewPostgres(db.Queries)
	tokens := aigateway.NewTokens(books)
	key := workspace.Key{Person: person, Project: project}

	first, err := tokens.Issue(t.Context(), key)
	if err != nil {
		t.Fatal(err)
	}
	second, err := tokens.Issue(t.Context(), key)
	if err != nil || first == second {
		t.Fatalf("Issue() again = %q, %v", second, err)
	}

	if _, found, _ := books.LookUp(t.Context(), aigateway.HashForTest(first)); found {
		t.Errorf("the first token still works after a second was issued")
	}
	caller, found, err := books.LookUp(t.Context(), aigateway.HashForTest(second))
	if err != nil || !found || caller.Organisation != organisation || caller.Person != person || caller.Project != project {
		t.Errorf("LookUp() = %+v, %v, %v", caller, found, err)
	}
	if err := tokens.Revoke(t.Context(), key); err != nil {
		t.Fatal(err)
	}
	if _, found, _ := books.LookUp(t.Context(), aigateway.HashForTest(second)); found {
		t.Errorf("a revoked token still works")
	}

	for _, used := range []aigateway.Usage{
		{Caller: caller, Model: "m", TokensIn: 100, TokensOut: 20, Status: 200},
		{Caller: caller, Model: "m", TokensIn: 5, TokensOut: 0, Status: 502},
	} {
		if err := books.Record(t.Context(), used); err != nil {
			t.Fatal(err)
		}
	}
	requests, total, err := books.UsedSince(t.Context(), organisation, time.Now().Add(-time.Hour))
	if err != nil || requests != 2 || total != 125 {
		t.Errorf("UsedSince() = %d, %d, %v", requests, total, err)
	}
	if requests, _, _ := books.UsedSince(t.Context(), organisation, time.Now().Add(time.Hour)); requests != 0 {
		t.Errorf("UsedSince() the future = %d", requests)
	}

	if limits, err := books.Limits(t.Context(), organisation); err != nil || limits.RequestsPerMinute != nil || limits.TokensPerMonth != nil {
		t.Errorf("a new organisation's limits = %+v, %v, want none", limits, err)
	}
	rate, budget := 30, int64(2_000_000)
	if err := books.SetLimits(t.Context(), organisation, aigateway.Limits{RequestsPerMinute: &rate, TokensPerMonth: &budget}); err != nil {
		t.Fatal(err)
	}
	limits, err := books.Limits(t.Context(), organisation)
	if err != nil || *limits.RequestsPerMinute != 30 || *limits.TokensPerMonth != 2_000_000 {
		t.Errorf("Limits() = %+v, %v", limits, err)
	}
	if err := books.SetLimits(t.Context(), uuid.New(), aigateway.Limits{}); !errors.Is(err, aigateway.ErrNotFound) {
		t.Errorf("SetLimits() of nobody = %v", err)
	}
	if _, err := books.Limits(t.Context(), uuid.New()); !errors.Is(err, aigateway.ErrNotFound) {
		t.Errorf("Limits() of nobody = %v", err)
	}
}
