package app

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestAboutAnswersWhatItCan(t *testing.T) {
	source := &about{
		hugo:     "0.152.2",
		git:      func(context.Context) (string, error) { return "", errors.New("no git") },
		postgres: func(context.Context) (string, error) { return "16.10", nil },
	}

	got := source.About(t.Context())

	if got.RunsOn.Git != "" {
		t.Errorf("git = %q, want empty when it cannot answer", got.RunsOn.Git)
	}
	if got.RunsOn.Hugo != "0.152.2" || got.RunsOn.Postgres != "16.10" || !strings.HasPrefix(got.RunsOn.Go, "go") {
		t.Errorf("runs on = %+v, want the others answered", got.RunsOn)
	}
	if got.Version == "" || got.Describe == "" {
		t.Errorf("version = %q, describe = %q", got.Version, got.Describe)
	}
	if !strings.HasPrefix(strings.TrimSpace(got.License.Text), "Apache License") || got.License.Name != "Apache 2.0" {
		t.Errorf("license = %q", got.License.Name)
	}
	if !strings.HasPrefix(got.Links.Contributors, "https://github.com/quiqr/chatcms-app/") {
		t.Errorf("links = %+v", got.Links)
	}
}

func TestASlowQuestionGivesUp(t *testing.T) {
	slow := func(ctx context.Context) (string, error) {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(10 * time.Second):
			return "too late", nil
		}
	}

	started := time.Now()
	got := ask(t.Context(), slow)
	if got != "" {
		t.Errorf("ask() = %q, want nothing from a question that took too long", got)
	}
	if time.Since(started) > askFor+time.Second {
		t.Errorf("ask() took %v, want it to give up after %v", time.Since(started), askFor)
	}
	if ask(t.Context(), nil) != "" {
		t.Error("ask(nil) answered something")
	}
}
