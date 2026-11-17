package app

import (
	"context"
	"runtime"
	"time"

	chatcms "github.com/quiqr/chatcms-app"
	"github.com/quiqr/chatcms-app/internal/httpapi"
	"github.com/quiqr/chatcms-app/internal/version"
)

const askFor = 2 * time.Second

type question func(ctx context.Context) (string, error)

type about struct {
	hugo     string
	git      question
	postgres question
}

func ask(ctx context.Context, q question) string {
	if q == nil {
		return ""
	}
	ctx, cancel := context.WithTimeout(ctx, askFor)
	defer cancel()
	answer, err := q(ctx)
	if err != nil {
		return ""
	}
	return answer
}

func (a *about) About(ctx context.Context) httpapi.AboutView {
	return httpapi.AboutView{
		Version:  version.Version,
		Source:   version.Short(),
		Changed:  version.Changed(),
		Describe: version.Describe(),
		RunsOn: httpapi.RunsOnView{
			Hugo:     a.hugo,
			Git:      ask(ctx, a.git),
			Postgres: ask(ctx, a.postgres),
			Go:       runtime.Version(),
		},
		License: httpapi.LicenseView{Name: "Apache 2.0", Text: chatcms.License()},
		Links: httpapi.LinksView{
			Source:       "https://github.com/quiqr/chatcms-app",
			Changelog:    "https://github.com/quiqr/chatcms-app/blob/main/CHANGELOG.md",
			Issues:       "https://github.com/quiqr/chatcms-app/issues",
			Contributors: "https://github.com/quiqr/chatcms-app/graphs/contributors",
		},
	}
}
