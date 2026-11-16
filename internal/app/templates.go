package app

import (
	"context"
	"log/slog"
	"os/exec"
	"time"

	"github.com/quiqr/chatcms-app/internal/jobs"
	"github.com/quiqr/chatcms-app/internal/onboarding"
)

func registerTemplateJob(worker *jobs.Worker, queue *jobs.Queue, catalogue *onboarding.Catalogue, logger *slog.Logger) {
	refresh := catalogue.RefreshJob(queue, func(err error) {
		logger.Warn("the template list could not be refreshed", "error", err)
	})
	worker.Handle(onboarding.TemplateRefreshJobKind, func(ctx context.Context, _ jobs.Job) error {
		return refresh(ctx)
	})
}

func installedHugo(ctx context.Context, hugoBin string, logger *slog.Logger) onboarding.Hugo {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, hugoBin, "version").Output()
	if err != nil {
		logger.WarnContext(ctx, "the installed Hugo could not say its version", "error", err)
		return onboarding.Hugo{}
	}
	return onboarding.ParseHugoVersion(string(out))
}
