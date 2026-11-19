package gitsync

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func (r *repo) Status(ctx context.Context) (Status, error) {
	status := Status{ExchangeInProgress: r.exchangeInProgress()}

	dirty, err := r.git.run(ctx, "status", "--porcelain", "--untracked-files=no")
	if err != nil {
		return Status{}, err
	}
	status.UnsavedWork = len(splitLines(dirty)) > 0

	if r.unborn(ctx) {
		return status, nil
	}

	remote := r.remoteRef()
	if !r.git.succeeds(ctx, "rev-parse", "--verify", "--quiet", remote) {
		// Nothing has been fetched yet, so every version is unpublished.
		count, err := r.countIn(ctx, "HEAD")
		if err != nil {
			return Status{}, err
		}
		status.Unpublished = count
		return status, nil
	}

	ahead, err := r.countIn(ctx, remote+"..HEAD")
	if err != nil {
		return Status{}, err
	}
	behind, err := r.countIn(ctx, "HEAD.."+remote)
	if err != nil {
		return Status{}, err
	}

	status.Unpublished = ahead
	status.RemoteAhead = behind
	if r.ownFirst {
		status.Unpublished, err = r.notYetPublished(ctx, remote)
		if err != nil {
			return Status{}, err
		}
	}
	return status, nil
}

func (r *repo) notYetPublished(ctx context.Context, shared string) (int, error) {
	args := []string{"rev-list", "--count", "HEAD", shared}
	if published := r.publishedRef(); r.git.succeeds(ctx, "rev-parse", "--verify", "--quiet", published) {
		args = append(args, "^"+published)
	}
	out, err := r.git.run(ctx, args...)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(out))
}

func (r *repo) unshared(ctx context.Context) (int, error) {
	if r.unborn(ctx) {
		return 0, nil
	}
	remote := r.remoteRef()
	if !r.git.succeeds(ctx, "rev-parse", "--verify", "--quiet", remote) {
		return r.countIn(ctx, "HEAD")
	}
	return r.countIn(ctx, remote+"..HEAD")
}

func (r *repo) remoteRef() string {
	return "refs/remotes/origin/" + r.branch
}

func (r *repo) publishedRef() string {
	if r.ownFirst {
		return "refs/remotes/published/" + r.branch
	}
	return r.remoteRef()
}

// exchangeInProgress asks the repository itself rather than any state of our
// own, so there is no second source of truth to get out of step.
func (r *repo) exchangeInProgress() bool {
	for _, marker := range []string{"rebase-merge", "rebase-apply"} {
		if _, err := os.Stat(filepath.Join(r.dir, ".git", marker)); err == nil {
			return true
		}
	}
	return false
}
