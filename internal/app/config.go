package app

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/quiqr/chatcms-app/internal/mail"
	"github.com/quiqr/chatcms-app/internal/onboarding"
)

type Config struct {
	Listen          string
	BaseURL         string
	StateDir        string
	DatabaseURL     string
	ShutdownTimeout time.Duration
	LogFormat       string
	GitBin          string
	HugoBin         string
	SSHKeyPath      string
	SecretKeyFile   string
	// FirstAdministrator is the email address of the person the server makes
	// an administrator when it starts. A new installation has nobody who
	// could appoint anybody, so this is the way in.
	FirstAdministrator string
	// Mail is where messages are sent, and as whom. With no host there is
	// nowhere to send, and links go to the log instead.
	MailHost         string
	MailPort         string
	MailUsername     string
	MailPassword     string
	MailFrom         string
	MailFromName     string
	MailStartTLS     string
	SyncInterval     time.Duration
	BwrapBin         string
	PrlimitBin       string
	BuildTimeout     time.Duration
	BuildMemory      int64
	WorkflowTools    []string
	CheckTimeout     time.Duration
	CheckMemory      int64
	WorkspaceIdle    time.Duration
	WorkspacesAtOnce int
	WorkspaceBackend string
	WorkspaceGuest   string
	WorkspaceRunDir  string
	WorkspaceTaps    int
	CloudHypervisor  string
	VirtiofsdBin     string
	NixBin           string
	NixhubURL        string
	AIListen         string
	AIProviderURL    string
	AIModel          string
	AIKey            string
	OpenCodeBin      string
	PreviewURL       string
	TemplateFeed     string
	Deprecated       []string
}

// SitesDir is where every site's working copy lives.
func (c Config) AIConfigured() bool {
	return c.AIProviderURL != "" && c.AIKey != "" && c.AIModel != ""
}

func (c Config) SitesDir() string {
	return filepath.Join(c.StateDir, "sites")
}

type Lookup func(string) (string, bool)

type ConfigError struct {
	Variable string
	Reason   string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("%s %s", e.Variable, e.Reason)
}

const (
	reasonMissing    = "is not set"
	reasonConflict   = "is set both directly and as a _FILE variable"
	reasonUnreadable = "names a file that cannot be read"
	reasonInvalid    = "is not valid"
)

func LoadConfig(lookup Lookup) (Config, error) {
	var cfg Config

	fields := []struct {
		name     string
		target   *string
		fallback string
		required bool
	}{
		{"CHATCMS_LISTEN", &cfg.Listen, "127.0.0.1:8080", false},
		{"CHATCMS_BASE_URL", &cfg.BaseURL, "http://localhost:8080", false},
		{"CHATCMS_STATE_DIR", &cfg.StateDir, "./.chatcms-state", false},
		{"CHATCMS_DATABASE_URL", &cfg.DatabaseURL, "", true},
		{"CHATCMS_LOG_FORMAT", &cfg.LogFormat, "auto", false},
		{"CHATCMS_GIT_BIN", &cfg.GitBin, "git", false},
		{"CHATCMS_HUGO_BIN", &cfg.HugoBin, "hugo", false},
		{"CHATCMS_SSH_KEY_PATH", &cfg.SSHKeyPath, "", false},
		{"CHATCMS_SECRET_KEY_FILE", &cfg.SecretKeyFile, "", true},
		{"CHATCMS_BWRAP_BIN", &cfg.BwrapBin, "bwrap", false},
		{"CHATCMS_PRLIMIT_BIN", &cfg.PrlimitBin, "prlimit", false},
		{"CHATCMS_FIRST_ADMINISTRATOR", &cfg.FirstAdministrator, "", false},
		{"CHATCMS_MAIL_HOST", &cfg.MailHost, "", false},
		{"CHATCMS_MAIL_PORT", &cfg.MailPort, "587", false},
		{"CHATCMS_MAIL_USERNAME", &cfg.MailUsername, "", false},
		{"CHATCMS_MAIL_PASSWORD", &cfg.MailPassword, "", false},
		{"CHATCMS_MAIL_FROM", &cfg.MailFrom, "", false},
		{"CHATCMS_MAIL_FROM_NAME", &cfg.MailFromName, "Quiqr ChatCMS", false},
		{"CHATCMS_MAIL_STARTTLS", &cfg.MailStartTLS, "yes", false},
		{"CHATCMS_WORKSPACE_BACKEND", &cfg.WorkspaceBackend, "local", false},
		{"CHATCMS_WORKSPACE_GUEST", &cfg.WorkspaceGuest, "", false},
		{"CHATCMS_WORKSPACE_RUN_DIR", &cfg.WorkspaceRunDir, "/run/chatcms/w", false},
		{"CHATCMS_CLOUD_HYPERVISOR_BIN", &cfg.CloudHypervisor, "cloud-hypervisor", false},
		{"CHATCMS_VIRTIOFSD_BIN", &cfg.VirtiofsdBin, "virtiofsd", false},
		{"CHATCMS_NIX_BIN", &cfg.NixBin, "", false},
		{"CHATCMS_NIXHUB_URL", &cfg.NixhubURL, "https://search.devbox.sh", false},
		{"CHATCMS_AI_LISTEN", &cfg.AIListen, "127.0.0.1:8091", false},
		{"CHATCMS_AI_PROVIDER_URL", &cfg.AIProviderURL, "", false},
		{"CHATCMS_AI_MODEL", &cfg.AIModel, "", false},
		{"CHATCMS_AI_KEY", &cfg.AIKey, "", false},
		{"CHATCMS_OPENCODE_BIN", &cfg.OpenCodeBin, "", false},
		{"CHATCMS_PREVIEW_URL", &cfg.PreviewURL, "", false},
	}

	for _, field := range fields {
		value, err := resolve(lookup, field.name)
		if err != nil {
			return Config{}, err
		}
		if value == "" {
			if field.required {
				return Config{}, &ConfigError{Variable: field.name, Reason: reasonMissing}
			}
			value = field.fallback
		}
		*field.target = value
	}

	durations := []struct {
		name     string
		target   *time.Duration
		fallback string
	}{
		{"CHATCMS_SHUTDOWN_TIMEOUT", &cfg.ShutdownTimeout, "20s"},
		{"CHATCMS_SYNC_INTERVAL", &cfg.SyncInterval, "2m"},
		{"CHATCMS_BUILD_TIMEOUT", &cfg.BuildTimeout, "10m"},
		{"CHATCMS_CHECK_TIMEOUT", &cfg.CheckTimeout, "5m"},
	}

	for _, field := range durations {
		raw, err := resolve(lookup, field.name)
		if err != nil {
			return Config{}, err
		}
		if raw == "" {
			raw = field.fallback
		}
		value, err := time.ParseDuration(raw)
		if err != nil || value <= 0 {
			return Config{}, &ConfigError{Variable: field.name, Reason: reasonInvalid}
		}
		*field.target = value
	}

	sizes := []struct {
		name     string
		target   *int64
		fallback string
	}{
		{"CHATCMS_BUILD_MEMORY", &cfg.BuildMemory, "2GiB"},
		{"CHATCMS_CHECK_MEMORY", &cfg.CheckMemory, "1GiB"},
	}

	for _, field := range sizes {
		raw, err := resolve(lookup, field.name)
		if err != nil {
			return Config{}, err
		}
		if raw == "" {
			raw = field.fallback
		}
		value, err := ParseBytes(raw)
		if err != nil {
			return Config{}, &ConfigError{Variable: field.name, Reason: reasonInvalid}
		}
		*field.target = value
	}

	idle, idleName, err := renamed(lookup, "CHATCMS_WORKSPACE_IDLE", "CHATCMS_PREVIEW_IDLE", &cfg.Deprecated)
	if err != nil {
		return Config{}, err
	}
	if idle == "" {
		idle = "15m"
	}
	cfg.WorkspaceIdle, err = time.ParseDuration(idle)
	if err != nil || cfg.WorkspaceIdle <= 0 {
		return Config{}, &ConfigError{Variable: idleName, Reason: reasonInvalid}
	}

	atOnce, atOnceName, err := renamed(lookup, "CHATCMS_WORKSPACES_AT_ONCE", "CHATCMS_PREVIEWS_AT_ONCE", &cfg.Deprecated)
	if err != nil {
		return Config{}, err
	}
	if atOnce == "" {
		atOnce = "8"
	}
	cfg.WorkspacesAtOnce, err = strconv.Atoi(atOnce)
	if err != nil || cfg.WorkspacesAtOnce <= 0 {
		return Config{}, &ConfigError{Variable: atOnceName, Reason: reasonInvalid}
	}

	taps, err := resolve(lookup, "CHATCMS_WORKSPACE_TAPS")
	if err != nil {
		return Config{}, err
	}
	cfg.WorkspaceTaps = cfg.WorkspacesAtOnce
	if taps != "" {
		cfg.WorkspaceTaps, err = strconv.Atoi(taps)
		if err != nil || cfg.WorkspaceTaps <= 0 || cfg.WorkspaceTaps > 250 {
			return Config{}, &ConfigError{Variable: "CHATCMS_WORKSPACE_TAPS", Reason: reasonInvalid}
		}
	}

	if (cfg.AIProviderURL == "") != (cfg.AIKey == "") || (cfg.AIProviderURL != "" && cfg.AIModel == "") {
		missing := "CHATCMS_AI_KEY"
		switch {
		case cfg.AIProviderURL == "":
			missing = "CHATCMS_AI_PROVIDER_URL"
		case cfg.AIModel == "":
			missing = "CHATCMS_AI_MODEL"
		}
		return Config{}, &ConfigError{Variable: missing, Reason: reasonMissing}
	}

	if cfg.PreviewURL != "" {
		parsed, err := url.Parse(cfg.PreviewURL)
		app, _ := url.Parse(cfg.BaseURL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || (parsed.Path != "" && parsed.Path != "/") || (app != nil && app.Host == parsed.Host) {
			return Config{}, &ConfigError{Variable: "CHATCMS_PREVIEW_URL", Reason: reasonInvalid}
		}
	}

	switch cfg.WorkspaceBackend {
	case "local":
	case "microvm":
		if cfg.WorkspaceGuest == "" {
			return Config{}, &ConfigError{Variable: "CHATCMS_WORKSPACE_GUEST", Reason: reasonMissing}
		}
	default:
		return Config{}, &ConfigError{Variable: "CHATCMS_WORKSPACE_BACKEND", Reason: reasonInvalid}
	}

	tools, err := resolve(lookup, "CHATCMS_WORKFLOW_TOOLS")
	if err != nil {
		return Config{}, err
	}
	cfg.WorkflowTools = splitPaths(tools)

	cfg.TemplateFeed, err = templateFeed(lookup)
	if err != nil {
		return Config{}, err
	}

	switch cfg.LogFormat {
	case "auto", "json", "text":
	default:
		return Config{}, &ConfigError{Variable: "CHATCMS_LOG_FORMAT", Reason: reasonInvalid}
	}

	return cfg, nil
}

func templateFeed(lookup Lookup) (string, error) {
	raw, err := resolve(lookup, "CHATCMS_TEMPLATE_FEED")
	if err != nil {
		return "", err
	}
	raw = strings.TrimSpace(raw)
	switch raw {
	case "":
		return onboarding.DefaultTemplateFeed, nil
	case "off":
		return "", nil
	}
	parsed, err := url.Parse(raw)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" {
		return "", &ConfigError{Variable: "CHATCMS_TEMPLATE_FEED", Reason: reasonInvalid}
	}
	return raw, nil
}

// splitPaths reads a colon-separated list the way PATH is written, so the
// setting looks like every other list of program locations.
func splitPaths(raw string) []string {
	out := []string{}
	for _, part := range strings.Split(raw, string(os.PathListSeparator)) {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

var byteUnits = []struct {
	suffix string
	scale  int64
}{
	{"GiB", 1 << 30},
	{"MiB", 1 << 20},
	{"KiB", 1 << 10},
	{"G", 1 << 30},
	{"M", 1 << 20},
	{"K", 1 << 10},
	{"B", 1},
}

func ParseBytes(raw string) (int64, error) {
	text := strings.TrimSpace(raw)
	scale := int64(1)
	for _, unit := range byteUnits {
		if trimmed, found := strings.CutSuffix(text, unit.suffix); found {
			text, scale = strings.TrimSpace(trimmed), unit.scale
			break
		}
	}
	value, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return 0, err
	}
	if value <= 0 {
		return 0, fmt.Errorf("a size must be more than nothing")
	}
	return value * scale, nil
}

func renamed(lookup Lookup, name, old string, deprecated *[]string) (string, string, error) {
	value, err := resolve(lookup, name)
	if err != nil || value != "" {
		return value, name, err
	}
	value, err = resolve(lookup, old)
	if err != nil || value == "" {
		return value, name, err
	}
	*deprecated = append(*deprecated, old+" is now "+name)
	return value, old, nil
}

func resolve(lookup Lookup, name string) (string, error) {
	direct, hasDirect := lookup(name)
	path, hasFile := lookup(name + "_FILE")

	if hasDirect && direct != "" && hasFile && path != "" {
		return "", &ConfigError{Variable: name, Reason: reasonConflict}
	}

	if hasFile && path != "" {
		contents, err := os.ReadFile(path)
		if err != nil {
			return "", &ConfigError{Variable: name + "_FILE", Reason: reasonUnreadable}
		}
		return strings.TrimSuffix(string(contents), "\n"), nil
	}

	return direct, nil
}

func EnvLookup() Lookup {
	return os.LookupEnv
}

// MailSettings is where messages go, read from the configuration. Sending is
// off when no host is named, which is the default: an install has to be told
// where to send before it sends anywhere.
func (c Config) MailSettings() mail.Settings {
	from := strings.TrimSpace(c.MailFrom)
	if from == "" && c.MailHost != "" {
		// A sensible default rather than a refusal to start: the host of the
		// address people already visit is the one a receiving server is most
		// likely to accept.
		from = "chatcms@" + hostOf(c.BaseURL)
	}

	return mail.Settings{
		Host:     c.MailHost,
		Port:     c.MailPort,
		Username: c.MailUsername,
		Password: c.MailPassword,
		From:     from,
		FromName: c.MailFromName,
		StartTLS: !isNo(c.MailStartTLS),
	}
}

// isNo reads the words people write when they mean no.
func isNo(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "no", "false", "off", "0":
		return true
	}
	return false
}
