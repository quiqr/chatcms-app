package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/admin"
	"github.com/quiqr/chatcms-app/internal/aigateway"
	"github.com/quiqr/chatcms-app/internal/assistant"
	"github.com/quiqr/chatcms-app/internal/auth"
	"github.com/quiqr/chatcms-app/internal/builds"
	"github.com/quiqr/chatcms-app/internal/content"
	"github.com/quiqr/chatcms-app/internal/gitsync"
	"github.com/quiqr/chatcms-app/internal/history"
	"github.com/quiqr/chatcms-app/internal/httpapi"
	"github.com/quiqr/chatcms-app/internal/jobs"
	"github.com/quiqr/chatcms-app/internal/mail"
	"github.com/quiqr/chatcms-app/internal/media"
	"github.com/quiqr/chatcms-app/internal/onboarding"
	"github.com/quiqr/chatcms-app/internal/preview"
	"github.com/quiqr/chatcms-app/internal/publish"
	"github.com/quiqr/chatcms-app/internal/runner"
	"github.com/quiqr/chatcms-app/internal/secrets"
	"github.com/quiqr/chatcms-app/internal/sitebuild"
	"github.com/quiqr/chatcms-app/internal/sitesync"
	"github.com/quiqr/chatcms-app/internal/store"
	"github.com/quiqr/chatcms-app/internal/tenancy"
	"github.com/quiqr/chatcms-app/internal/version"
	"github.com/quiqr/chatcms-app/internal/webui"
	"github.com/quiqr/chatcms-app/internal/wholesite"
	"github.com/quiqr/chatcms-app/internal/workflows"
	"github.com/quiqr/chatcms-app/internal/workspace"
)

var ErrShutdownForced = errors.New("in-flight requests did not finish within the shutdown timeout")

type Server struct {
	cfg      Config
	logger   *slog.Logger
	db       *store.DB
	listener net.Listener
	http     *http.Server
	known    int64

	queue         *jobs.Queue
	worker        *jobs.Worker
	sites         *gitsync.Sites
	keys          *onboarding.Keys
	auth          *auth.Service
	tenancy       *tenancy.Service
	admin         *admin.Web
	about         *about
	onboarding    *onboarding.Service
	templates     *onboarding.Catalogue
	builds        *builds.Service
	workflows     *workflows.Engine
	publish       *publish.Web
	history       *history.Web
	content       *content.Web
	media         *media.Web
	standing      *sitesync.Web
	preview       *preview.Web
	workspaces    *workspace.Workspaces
	hugo          *hugos
	aiListener    net.Listener
	aiHTTP        *http.Server
	aiTokens      *aigateway.Tokens
	aiGateway     *aigateway.Gateway
	aiMux         *http.ServeMux
	chat          *assistant.Chat
	chatWeb       *assistant.Web
	previewOrigin *preview.Origin
}

func NewServer(ctx context.Context, cfg Config, logger *slog.Logger) (*Server, error) {
	for _, warning := range cfg.Deprecated {
		logger.Warn("a configuration name was renamed", "renamed", warning)
	}
	db, err := store.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("the database is not reachable: %w", err)
	}

	if err := db.Migrate(ctx); err != nil {
		db.Close()
		return nil, err
	}

	known, err := db.KnownSchemaVersion()
	if err != nil {
		db.Close()
		return nil, err
	}

	server := &Server{cfg: cfg, logger: logger, db: db, known: known}

	keeper, err := secrets.NewKeeper(cfg.SecretKeyFile)
	if err != nil {
		db.Close()
		return nil, err
	}

	server.keys, err = onboarding.NewKeys(db, keeper, cfg.StateDir)
	if err != nil {
		db.Close()
		return nil, err
	}

	server.sites = gitsync.NewSites(gitsync.Config{
		GitBin:     cfg.GitBin,
		SSHKeyPath: cfg.SSHKeyPath,
		System:     gitsync.Identity{Name: "Quiqr ChatCMS", Email: "chatcms@" + hostOf(cfg.BaseURL)},
	}, cfg.SitesDir(), server.keys.PathForSite)

	server.auth = auth.NewService(db, logger, cfg.BaseURL)
	tenancyStore := tenancy.NewStore(db)
	server.tenancy = tenancy.NewService(tenancyStore, server.auth)
	server.tenancy.MakesCopies(wholesite.NewCopier(server.sites))
	server.tenancy.ForgetsKeptSites(func(_ context.Context, projectID uuid.UUID) error {
		return server.sites.Unhost(projectID.String())
	})
	server.tenancy.KnowsWhatTheyChose(server.auth)

	adminStore := admin.NewStore(db)
	administering := admin.NewService(adminStore, tenancyStore, server.auth)
	administering.UseAIBooks(aigateway.NewBooks(aigateway.NewPostgres(db.Queries), nil))
	server.admin = admin.NewWeb(administering)
	appointTheFirstAdministrator(ctx, adminStore, cfg.FirstAdministrator, logger)

	importer := onboarding.NewImporter(db, server.sites, server.keys,
		gitsync.Identity{Name: "Quiqr ChatCMS", Email: "chatcms@" + hostOf(cfg.BaseURL)})
	server.templates = onboarding.NewCatalogue(db, cfg.TemplateFeed, cfg.StateDir, nil)
	importer.UseTemplates(server.templates)
	server.onboarding = onboarding.NewService(db, tenancyStore, server.keys, importer)
	hugo := installedHugo(ctx, cfg.HugoBin, logger)
	server.onboarding.UseTemplates(server.templates, hugo)
	server.about = &about{
		hugo: hugo.Version,
		git:  server.sites.GitVersion,
		postgres: func(ctx context.Context) (string, error) {
			var answer string
			err := db.Pool().QueryRow(ctx, "SHOW server_version").Scan(&answer)
			return answer, err
		},
	}

	if err := server.startBuilds(ctx); err != nil {
		db.Close()
		return nil, err
	}

	if err := server.startPublishing(ctx, tenancyStore); err != nil {
		db.Close()
		return nil, err
	}

	server.queue = jobs.NewQueue(db)
	server.worker = jobs.NewWorker(server.queue, jobs.WorkerOptions{Logger: logger})
	registerSyncJob(server.worker, server.sites, logger)
	registerShareJob(server.worker, server.queue, server.sites, logger)
	registerMailJob(server.worker, cfg.MailSettings(), logger)
	registerTemplateJob(server.worker, server.queue, server.templates, logger)
	if err := server.templates.Start(ctx, server.queue); err != nil {
		logger.ErrorContext(ctx, "the template list could not be scheduled", "error", err)
	}

	// Sending goes through the queue, so that nobody waits on a mail server
	// and a slow one cannot lose an invitation.
	settings := cfg.MailSettings()
	server.auth.SendsWith(mail.NewPost(server.queue, settings.Configured(), logger, nil))
	if !settings.Configured() {
		logger.InfoContext(ctx, "no mail server is configured, so sign-in links go to this log")
	}

	if err := server.startAI(); err != nil {
		db.Close()
		return nil, err
	}
	if err := server.startContent(ctx, tenancyStore); err != nil {
		db.Close()
		return nil, err
	}

	handler := httpapi.New(httpapi.Options{
		Logger:         logger,
		Health:         server.health,
		Assets:         webui.Handler(),
		Auth:           server.auth,
		Tenancy:        server.tenancy,
		WhoAmI:         server.tenancy,
		Account:        server.auth,
		Administration: server.admin,
		AIUse:          server.admin,
		About:          server.about,
		Onboarding:     server.onboarding,
		Templates:      server.onboarding,
		Connect:        server.onboarding,
		Publish:        server.publish,
		History:        server.history,
		Content:        server.content,
		Browse:         server.content,
		Menu:           server.content,
		Media:          server.media,
		Sync:           server.standing,
		Preview:        server.appPreview(),
		Workspace:      server.preview,
		Chat:           server.chatWeb,
		Secure:         strings.HasPrefix(strings.ToLower(cfg.BaseURL), "https://"),
	})

	handler = server.byHost(handler)

	listener, err := net.Listen("tcp", cfg.Listen)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("listening on %s failed: %w", cfg.Listen, err)
	}

	server.listener = listener
	server.http = &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
	return server, nil
}

func (s *Server) Addr() string {
	return s.listener.Addr().String()
}

func (s *Server) health(ctx context.Context) httpapi.Health {
	health := httpapi.Health{
		Status:        httpapi.HealthStatusOK,
		ServerVersion: version.Version,
		Database:      httpapi.DatabaseReachable,
		SchemaVersion: fmt.Sprintf("%d", s.known),
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := s.db.Ping(ctx); err != nil {
		health.Status = httpapi.HealthStatusDegraded
		health.Database = httpapi.DatabaseUnreachable
		return health
	}

	applied, err := s.db.SchemaVersion(ctx)
	if err != nil {
		health.Status = httpapi.HealthStatusDegraded
		health.Database = httpapi.DatabaseUnreachable
		return health
	}
	health.SchemaVersion = fmt.Sprintf("%d", applied)
	return health
}

func (s *Server) Serve(ctx context.Context) error {
	background, stopBackground := context.WithCancel(context.Background())
	defer stopBackground()

	var workers sync.WaitGroup
	workers.Add(2)
	go func() {
		defer workers.Done()
		s.worker.Run(background)
	}()
	go func() {
		defer workers.Done()
		scheduleSyncs(background, s.queue, s.sites, s.cfg.SyncInterval, s.logger)
	}()

	errs := make(chan error, 1)
	go func() {
		if s.aiHTTP != nil {
			go s.aiHTTP.Serve(s.aiListener)
		}
		err := s.http.Serve(s.listener)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		errs <- err
	}()

	s.logger.Info("listening", "address", s.Addr(), "version", version.Version)

	select {
	case err := <-errs:
		stopBackground()
		workers.Wait()
		s.db.Close()
		return err
	case <-ctx.Done():
	}

	s.logger.Info("shutting down", "timeout", s.cfg.ShutdownTimeout)

	if s.aiHTTP != nil {
		gracefulShutdown(s.aiHTTP, s.cfg.ShutdownTimeout)
	}
	shutdownErr := gracefulShutdown(s.http, s.cfg.ShutdownTimeout)
	<-errs

	stopBackground()
	if err := s.worker.Drain(context.Background(), s.cfg.ShutdownTimeout); err != nil {
		s.logger.Error("returning unfinished work to the queue failed", "error", err)
	}
	workers.Wait()

	if s.workspaces != nil {
		closing, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
		s.workspaces.Close(closing)
		cancel()
	}
	if err := s.keys.Close(); err != nil {
		s.logger.Error("clearing the decrypted keys failed", "error", err)
	}
	s.db.Close()

	if errors.Is(shutdownErr, ErrShutdownForced) {
		s.logger.Error("shutdown was forced", "timeout", s.cfg.ShutdownTimeout)
	}
	return shutdownErr
}

func (s *Server) Close() error {
	if s.workspaces != nil {
		closing, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
		s.workspaces.Close(closing)
		cancel()
	}
	if s.aiHTTP != nil {
		s.aiHTTP.Close()
	}
	if s.db != nil {
		s.db.Close()
	}
	return s.http.Close()
}

func Serve(ctx context.Context, cfg Config, logger *slog.Logger) error {
	server, err := NewServer(ctx, cfg, logger)
	if err != nil {
		return err
	}
	return server.Serve(ctx)
}

// Auth exposes the identity service, which the integration tests use to
// create a person without an endpoint for it.
func (s *Server) Auth() *auth.Service { return s.auth }

// Administration exposes what somebody who administers the installation may
// do, which the acceptance test needs in order to be that person.
func (s *Server) Administration() *admin.Web { return s.admin }

// MakeAdministrator appoints somebody, which the integration tests use to
// set a scene that would otherwise need a configured address and a restart.
func (s *Server) MakeAdministrator(ctx context.Context, personID uuid.UUID) error {
	return admin.NewStore(s.db).MakeAdministrator(ctx, personID)
}

// Tenancy exposes the organisations and projects, which the integration
// tests use to set a scene without an endpoint for every step.
func (s *Server) Tenancy() *tenancy.Service { return s.tenancy }

// Onboarding exposes setting a project up, which the acceptance test drives.
func (s *Server) Onboarding() *onboarding.Service { return s.onboarding }

func (s *Server) Builds() *builds.Service { return s.builds }

func (s *Server) Workflows() *workflows.Engine { return s.workflows }

func (s *Server) Publish() *publish.Web { return s.publish }

func (s *Server) History() *history.Web { return s.history }

func (s *Server) Content() *content.Web { return s.content }

func (s *Server) Media() *media.Web { return s.media }

func (s *Server) Standing() *sitesync.Web { return s.standing }

func (s *Server) Preview() *preview.Web { return s.preview }

func (s *Server) startContent(ctx context.Context, tenancyStore *tenancy.Store) error {
	sources := contentSources{
		publishSources: publishSources{sites: s.sites, queries: s.db.Queries, baseURL: s.cfg.BaseURL},
		people:         s.auth,
	}

	service, err := content.NewService(sources, queueSync{queue: s.queue})
	if err != nil {
		return err
	}

	s.content, err = content.NewWeb(service, tenancyStore)
	if err != nil {
		return err
	}

	thumbnails, err := media.NewThumbnails(filepath.Join(s.cfg.StateDir, "thumbnails"))
	if err != nil {
		return err
	}

	library, err := media.NewLibrary(mediaSources{contentSources: sources}, queueSync{queue: s.queue}, thumbnails)
	if err != nil {
		return err
	}

	s.media, err = media.NewWeb(library, tenancyStore)
	if err != nil {
		return err
	}

	standing, err := sitesync.NewService(syncSources{publishSources: sources.publishSources})
	if err != nil {
		return err
	}

	s.standing, err = sitesync.NewWeb(standing, tenancyStore)
	if err != nil {
		return err
	}

	s.workspaces, err = s.startWorkspaces(ctx, tenancyStore)
	if err != nil {
		return err
	}
	s.startAssistant(service, tenancyStore)

	previewing, err := preview.NewService(sources.publishSources, s.workspaces, s.cfg.BaseURL)
	if err != nil {
		return err
	}
	if s.cfg.PreviewURL != "" {
		previewing.ServeFrom(s.cfg.PreviewURL)
		allowed := mayStillUse(sources.publishSources, tenancyStore)
		s.previewOrigin = preview.NewOrigin(s.workspaces, allowed)
	}

	s.chatWeb = assistant.NewWeb(s.chat, tenancyStore, func(organisation, project string) string {
		return previewing.Under(httpapi.PreviewPrefix + organisation + "/" + project + "/")
	})

	s.preview, err = preview.NewWeb(previewing, tenancyStore)
	return err
}

func (s *Server) startBuilds(ctx context.Context) error {
	records := builds.NewRecords(s.db.Queries)

	interrupted, err := records.InterruptLeftovers(ctx)
	if err != nil {
		return err
	}
	if interrupted > 0 {
		s.logger.Warn("builds were still running when the server last stopped", "count", interrupted)
	}

	sandbox, err := runner.NewSandbox(runner.Config{
		BwrapBin:   s.cfg.BwrapBin,
		PrlimitBin: s.cfg.PrlimitBin,
		WorkRoot:   filepath.Join(s.cfg.StateDir, "builds"),
	})
	if err != nil {
		if !errors.Is(err, runner.ErrNoIsolation) {
			return err
		}
		s.logger.Warn("builds are unavailable because they cannot be isolated on this machine", "reason", err)
		return nil
	}

	builder, err := sitebuild.New(sandbox, s.cfg.HugoBin, sitebuild.Limits{
		Timeout:     s.cfg.BuildTimeout,
		MemoryBytes: s.cfg.BuildMemory,
	})
	if err != nil {
		return err
	}
	builder.UseRuntimeOf(s.projectHugo().Runtime)

	s.builds, err = builds.NewService(records, builder, s.siteFor)
	if err != nil {
		return err
	}

	publisher, err := publish.NewPublisher(s.sites, publishTarget(s.db.Queries))
	if err != nil {
		return err
	}

	s.workflows, err = workflows.NewEngine(sandbox, builder, publisher, s.cfg.WorkflowTools, workflows.Limits{
		CheckTimeout: s.cfg.CheckTimeout,
		CheckMemory:  s.cfg.CheckMemory,
	})
	return err
}

func (s *Server) startPublishing(ctx context.Context, tenancyStore *tenancy.Store) error {
	records := publish.NewRecords(s.db.Queries)

	interrupted, err := records.InterruptLeftovers(ctx)
	if err != nil {
		return err
	}
	if interrupted > 0 {
		s.logger.Warn("publishes were still going when the server last stopped", "count", interrupted)
	}

	if s.workflows == nil {
		s.logger.Warn("publishing is unavailable because builds cannot be isolated on this machine")
		return nil
	}

	sources := publishSources{sites: s.sites, queries: s.db.Queries, baseURL: s.cfg.BaseURL}

	// One source for how much is waiting, so that the number on a button and
	// the number on the publish screen cannot disagree. They did.
	s.tenancy.KnowsWhatIsWaiting(waitingInSites{sources: sources})
	service, err := publish.NewService(sources, s.workflows, records)
	if err != nil {
		return err
	}

	s.publish, err = publish.NewWeb(service, tenancyStore)
	if err != nil {
		return err
	}

	timeline, err := history.NewService(
		historySources{publishSources: sources, records: records, people: s.auth, queue: s.queue},
		service,
	)
	if err != nil {
		return err
	}

	s.history, err = history.NewWeb(timeline, tenancyStore)
	return err
}

func (s *Server) siteFor(ctx context.Context, projectID uuid.UUID) (fs.FS, error) {
	repository, err := openSite(ctx, s.sites, projectID)
	if err != nil {
		return nil, err
	}
	return repository.FS(), nil
}

// hostOf is used for the address versions are written by, so history shows
// where an edit came from rather than a placeholder domain.
func hostOf(baseURL string) string {
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Hostname() == "" {
		return "chatcms.local"
	}
	return strings.ToLower(parsed.Hostname())
}

// appointTheFirstAdministrator makes the configured person an administrator
// when the server starts. A new installation has nobody who could appoint
// anybody, so this is the way in. It is safe to run on every start, because a
// NixOS module does exactly that, and it says what it did rather than leaving
// whoever is reading the log to guess.
func appointTheFirstAdministrator(ctx context.Context, store *admin.Store, email string, logger *slog.Logger) {
	if strings.TrimSpace(email) == "" {
		return
	}

	changed, err := store.MakeAdministratorByEmail(ctx, email)
	switch {
	case errors.Is(err, admin.ErrNoSuchPerson):
		// Not a failure to start on: the address is usually configured before
		// the person has accepted their invitation, and the next start will
		// find them.
		logger.WarnContext(ctx,
			"nobody with the configured address has an account yet, so no administrator was appointed",
			"email", email)
	case err != nil:
		logger.ErrorContext(ctx, "appointing the first administrator failed", "error", err, "email", email)
	case changed:
		logger.InfoContext(ctx, "the configured person now administers this installation", "email", email)
	}
}

func (s *Server) appPreview() httpapi.PreviewService {
	if s.previewOrigin != nil || s.preview == nil {
		return nil
	}
	return s.preview
}

func (s *Server) byHost(app http.Handler) http.Handler {
	if s.previewOrigin == nil {
		return app
	}
	parsed, err := url.Parse(s.cfg.PreviewURL)
	if err != nil {
		return app
	}
	previewHost := strings.ToLower(parsed.Host)
	origin := s.previewOrigin
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.ToLower(r.Host) == previewHost {
			origin.ServeHTTP(w, r)
			return
		}
		app.ServeHTTP(w, r)
	})
}
