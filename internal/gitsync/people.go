package gitsync

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const shareAttempts = 3

func (s *Sites) personRoot(site, person string) (string, error) {
	if err := checkSiteID(site); err != nil {
		return "", err
	}
	if err := checkSiteID(person); err != nil {
		return "", err
	}
	return filepath.Join(s.root, site, "people", person), nil
}

func (s *Sites) personDir(site, person string) (string, error) {
	root, err := s.personRoot(site, person)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "working-copy"), nil
}

func (s *Sites) DirFor(site, person string) (string, error) {
	return s.personDir(site, person)
}

func (s *Sites) localConfig() Config {
	cfg := s.cfg
	cfg.SSHKeyPath = ""
	return cfg
}

func (s *Sites) OpenFor(ctx context.Context, site, person string) (Repo, error) {
	r, err := s.openPerson(ctx, site, person)
	if err != nil {
		return nil, err
	}
	if err := r.catchUp(ctx); err != nil {
		return nil, err
	}
	return r, nil
}

func (s *Sites) openPerson(ctx context.Context, site, person string) (*repo, error) {
	shared, err := s.Dir(site)
	if err != nil {
		return nil, err
	}
	dir, err := s.personDir(site, person)
	if err != nil {
		return nil, err
	}
	if err := s.cfg.validate(); err != nil {
		return nil, err
	}
	if !isWorkingCopy(shared) {
		return nil, ErrNotAWorkingCopy
	}
	if err := s.acceptSends(ctx, shared); err != nil {
		return nil, err
	}

	cfg := s.localConfig()
	if !isWorkingCopy(dir) {
		if err := s.copyFor(ctx, cfg, shared, dir); err != nil {
			return nil, err
		}
	}

	opened, err := openWith(ctx, cfg, dir, true)
	if err != nil {
		return nil, err
	}
	return opened.(*repo), nil
}

func (s *Sites) copyFor(ctx context.Context, cfg Config, shared, dir string) error {
	lock := siteLock(dir)
	lock.Lock()
	defer lock.Unlock()

	if isWorkingCopy(dir) {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(dir), 0o700); err != nil {
		return fmt.Errorf("preparing the person's directory failed: %w", err)
	}
	home, err := gitHome(dir)
	if err != nil {
		return err
	}
	making, err := os.MkdirTemp(filepath.Dir(dir), ".making-")
	if err != nil {
		return fmt.Errorf("preparing the person's copy failed: %w", err)
	}
	defer os.RemoveAll(making)

	made := filepath.Join(making, "working-copy")
	if _, err := newRunner(cfg, making, home).run(ctx, "clone", "--local", "--", shared, made); err != nil {
		return err
	}
	git := newRunner(cfg, made, home)
	if _, err := git.run(ctx, "config", "--add", "remote.origin.fetch", "+refs/remotes/origin/*:refs/remotes/published/*"); err != nil {
		return err
	}
	if _, err := git.run(ctx, "fetch", "--prune", "origin"); err != nil {
		return err
	}
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("clearing the person's directory failed: %w", err)
	}
	if err := os.Rename(made, dir); err != nil {
		return fmt.Errorf("putting the person's copy in place failed: %w", err)
	}
	return nil
}

func (s *Sites) acceptSends(ctx context.Context, shared string) error {
	home, err := gitHome(shared)
	if err != nil {
		return err
	}
	git := newRunner(s.localConfig(), shared, home)

	current, _ := git.run(ctx, "config", "--local", "--get", "receive.denyCurrentBranch")
	if strings.TrimSpace(current) == "updateInstead" {
		return nil
	}

	lock := siteLock(shared)
	lock.Lock()
	defer lock.Unlock()

	_, err = git.run(ctx, "config", "--local", "receive.denyCurrentBranch", "updateInstead")
	return err
}

func (r *repo) catchUp(ctx context.Context) error {
	lock := siteLock(r.dir)
	lock.Lock()
	defer lock.Unlock()

	if r.exchangeInProgress() {
		return nil
	}
	if _, err := r.git.run(ctx, "fetch", "--prune", "origin"); err != nil {
		return err
	}
	remote := r.remoteRef()
	if !r.git.succeeds(ctx, "rev-parse", "--verify", "--quiet", remote) {
		return nil
	}
	if r.unborn(ctx) {
		_, err := r.git.run(ctx, "reset", "--hard", remote)
		return err
	}

	behind, err := r.countIn(ctx, "HEAD.."+remote)
	if err != nil || behind == 0 {
		return err
	}
	ahead, err := r.countIn(ctx, remote+"..HEAD")
	if err != nil || ahead > 0 {
		return err
	}
	dirty, err := r.git.run(ctx, "status", "--porcelain", "--untracked-files=no")
	if err != nil || strings.TrimSpace(dirty) != "" {
		return err
	}
	_, err = r.git.run(ctx, "merge", "--ff-only", remote)
	if err == nil {
		return nil
	}
	in, listErr := r.git.run(ctx, "diff", "--name-only", "--diff-filter=A", "HEAD", remote)
	if listErr != nil {
		return err
	}
	removed := false
	for _, name := range splitLines(in) {
		full := filepath.Join(r.dir, filepath.FromSlash(name))
		if !r.git.succeeds(ctx, "ls-files", "--error-unmatch", "--", name) {
			if _, statErr := os.Lstat(full); statErr == nil {
				if os.RemoveAll(full) == nil {
					removed = true
				}
			}
		}
	}
	if !removed {
		return err
	}
	_, err = r.git.run(ctx, "merge", "--ff-only", remote)
	return err
}

func (s *Sites) Share(ctx context.Context, site, person string) (SyncResult, error) {
	shared, err := s.Dir(site)
	if err != nil {
		return SyncResult{}, err
	}

	r, err := s.openPerson(ctx, site, person)
	if err != nil {
		return SyncResult{}, err
	}

	lock := siteLock(shared)
	lock.Lock()
	defer lock.Unlock()

	var last error
	for attempt := 0; attempt < shareAttempts; attempt++ {
		result, err := r.Sync(ctx)
		if err != nil {
			return result, err
		}
		if result.Conflict != nil {
			return result, nil
		}
		if _, err := r.Publish(ctx); err != nil {
			if errors.Is(err, ErrRemoteMoved) {
				last = err
				continue
			}
			return SyncResult{}, err
		}
		return result, nil
	}
	return SyncResult{}, last
}

func (s *Sites) Unshared(ctx context.Context, site, person string) (int, error) {
	dir, err := s.personDir(site, person)
	if err != nil {
		return 0, err
	}
	if !isWorkingCopy(dir) {
		return 0, nil
	}
	r, err := s.openPerson(ctx, site, person)
	if err != nil {
		return 0, err
	}
	if r.exchangeInProgress() {
		return 1, nil
	}
	if _, err := r.git.run(ctx, "fetch", "--prune", "origin"); err != nil {
		return 0, err
	}
	count, err := r.unshared(ctx)
	if err != nil {
		return 0, err
	}
	status, err := r.Status(ctx)
	if err != nil {
		return 0, err
	}
	if status.UnsavedWork {
		count++
	}
	return count, nil
}

func (s *Sites) RemoveFor(site, person string) error {
	root, err := s.personRoot(site, person)
	if err != nil {
		return err
	}
	dir := filepath.Join(root, "working-copy")
	lock := siteLock(dir)
	lock.Lock()
	defer lock.Unlock()

	if err := os.RemoveAll(root); err != nil {
		return fmt.Errorf("removing the person's copy failed: %w", err)
	}
	return nil
}
