package app

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func lookupFrom(pairs map[string]string) Lookup {
	return func(key string) (string, bool) {
		value, ok := pairs[key]
		return value, ok
	}
}

func TestLoadConfigAppliesDefaults(t *testing.T) {
	cfg, err := LoadConfig(lookupFrom(map[string]string{
		"CHATCMS_DATABASE_URL":    "postgres:///chatcms",
		"CHATCMS_SECRET_KEY_FILE": "/run/secrets/master-key",
	}))
	if err != nil {
		t.Fatalf("LoadConfig() error = %v, want nil", err)
	}

	want := map[string]string{
		"Listen":    "127.0.0.1:8080",
		"BaseURL":   "http://localhost:8080",
		"StateDir":  "./.chatcms-state",
		"LogFormat": "auto",
		"GitBin":    "git",
		"HugoBin":   "hugo",
	}
	got := map[string]string{
		"Listen":    cfg.Listen,
		"BaseURL":   cfg.BaseURL,
		"StateDir":  cfg.StateDir,
		"LogFormat": cfg.LogFormat,
		"GitBin":    cfg.GitBin,
		"HugoBin":   cfg.HugoBin,
	}
	for field, value := range want {
		if got[field] != value {
			t.Errorf("%s = %q, want %q", field, got[field], value)
		}
	}

	if cfg.ShutdownTimeout != 20*time.Second {
		t.Errorf("ShutdownTimeout = %v, want 20s", cfg.ShutdownTimeout)
	}
}

func TestLoadConfigPrefersEnvironment(t *testing.T) {
	cfg, err := LoadConfig(lookupFrom(map[string]string{
		"CHATCMS_DATABASE_URL":     "postgres:///other",
		"CHATCMS_SECRET_KEY_FILE":  "/run/secrets/master-key",
		"CHATCMS_LISTEN":           "0.0.0.0:9000",
		"CHATCMS_BASE_URL":         "https://chatcms.eu",
		"CHATCMS_STATE_DIR":        "/var/lib/chatcms",
		"CHATCMS_GIT_BIN":          "/nix/store/git",
		"CHATCMS_HUGO_BIN":         "/nix/store/hugo",
		"CHATCMS_LOG_FORMAT":       "json",
		"CHATCMS_SHUTDOWN_TIMEOUT": "45s",
	}))
	if err != nil {
		t.Fatalf("LoadConfig() error = %v, want nil", err)
	}

	if cfg.Listen != "0.0.0.0:9000" || cfg.BaseURL != "https://chatcms.eu" {
		t.Errorf("environment not applied: %+v", cfg)
	}
	if cfg.StateDir != "/var/lib/chatcms" || cfg.GitBin != "/nix/store/git" {
		t.Errorf("environment not applied: %+v", cfg)
	}
	if cfg.HugoBin != "/nix/store/hugo" || cfg.LogFormat != "json" {
		t.Errorf("environment not applied: %+v", cfg)
	}
	if cfg.ShutdownTimeout != 45*time.Second {
		t.Errorf("ShutdownTimeout = %v, want 45s", cfg.ShutdownTimeout)
	}
}

func TestLoadConfigRequiresDatabaseURL(t *testing.T) {
	_, err := LoadConfig(lookupFrom(nil))

	var configErr *ConfigError
	if !errors.As(err, &configErr) {
		t.Fatalf("LoadConfig() error = %v, want a *ConfigError", err)
	}
	if configErr.Variable != "CHATCMS_DATABASE_URL" {
		t.Errorf("Variable = %q, want CHATCMS_DATABASE_URL", configErr.Variable)
	}
}

func TestLoadConfigErrorNamesVariableNotValue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "url")
	if err := os.WriteFile(path, []byte("postgres://secret@host/db"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := LoadConfig(lookupFrom(map[string]string{
		"CHATCMS_DATABASE_URL_FILE": path,
		"CHATCMS_SECRET_KEY_FILE":   "/run/secrets/master-key",
		"CHATCMS_LOG_FORMAT":        "yelling",
	}))
	if err == nil {
		t.Fatal("LoadConfig() error = nil, want an error for the log format")
	}
	if !strings.Contains(err.Error(), "CHATCMS_LOG_FORMAT") {
		t.Errorf("error = %q, want it to name the variable", err)
	}
	if strings.Contains(err.Error(), "secret") {
		t.Errorf("error = %q, want it to carry no value", err)
	}
}

func TestLoadConfigReadsSecretFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "database-url")
	if err := os.WriteFile(path, []byte("postgres:///from-file\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(lookupFrom(map[string]string{
		"CHATCMS_DATABASE_URL_FILE": path,
		"CHATCMS_SECRET_KEY_FILE":   "/run/secrets/master-key",
	}))
	if err != nil {
		t.Fatalf("LoadConfig() error = %v, want nil", err)
	}
	if cfg.DatabaseURL != "postgres:///from-file" {
		t.Errorf("DatabaseURL = %q, want the file contents without the trailing newline", cfg.DatabaseURL)
	}
}

func TestLoadConfigRejectsBothValueAndFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "database-url")
	if err := os.WriteFile(path, []byte("postgres:///from-file"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := LoadConfig(lookupFrom(map[string]string{
		"CHATCMS_DATABASE_URL":      "postgres:///direct",
		"CHATCMS_DATABASE_URL_FILE": path,
		"CHATCMS_SECRET_KEY_FILE":   "/run/secrets/master-key",
	}))

	var configErr *ConfigError
	if !errors.As(err, &configErr) {
		t.Fatalf("LoadConfig() error = %v, want a *ConfigError", err)
	}
	if !strings.Contains(configErr.Error(), "both") {
		t.Errorf("error = %q, want it to say both forms are set", configErr)
	}
}

func TestLoadConfigRejectsUnreadableSecretFile(t *testing.T) {
	_, err := LoadConfig(lookupFrom(map[string]string{
		"CHATCMS_DATABASE_URL_FILE": filepath.Join(t.TempDir(), "absent"),
	}))

	var configErr *ConfigError
	if !errors.As(err, &configErr) {
		t.Fatalf("LoadConfig() error = %v, want a *ConfigError", err)
	}
	if configErr.Variable != "CHATCMS_DATABASE_URL_FILE" {
		t.Errorf("Variable = %q, want the _FILE variable", configErr.Variable)
	}
}

func TestLoadConfigRejectsInvalidShutdownTimeout(t *testing.T) {
	for name, value := range map[string]string{
		"unparseable": "soon",
		"zero":        "0s",
		"negative":    "-5s",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := LoadConfig(lookupFrom(map[string]string{
				"CHATCMS_DATABASE_URL":     "postgres:///chatcms",
				"CHATCMS_SECRET_KEY_FILE":  "/run/secrets/master-key",
				"CHATCMS_SHUTDOWN_TIMEOUT": value,
			}))
			if err == nil {
				t.Fatalf("LoadConfig() with %q error = nil, want an error", value)
			}
		})
	}
}

func TestLoadConfigReadsSyncSettings(t *testing.T) {
	cfg, err := LoadConfig(lookupFrom(map[string]string{
		"CHATCMS_DATABASE_URL":    "postgres:///chatcms",
		"CHATCMS_SECRET_KEY_FILE": "/run/secrets/master-key",
		"CHATCMS_SYNC_INTERVAL":   "45s",
		"CHATCMS_SSH_KEY_PATH":    "/run/secrets/site-key",
	}))
	if err != nil {
		t.Fatalf("LoadConfig() error = %v, want nil", err)
	}
	if cfg.SyncInterval != 45*time.Second {
		t.Errorf("SyncInterval = %v, want 45s", cfg.SyncInterval)
	}
	if cfg.SSHKeyPath != "/run/secrets/site-key" {
		t.Errorf("SSHKeyPath = %q", cfg.SSHKeyPath)
	}
}

func TestSyncIntervalHasADefault(t *testing.T) {
	cfg, err := LoadConfig(lookupFrom(map[string]string{
		"CHATCMS_DATABASE_URL":    "postgres:///chatcms",
		"CHATCMS_SECRET_KEY_FILE": "/run/secrets/master-key",
	}))
	if err != nil {
		t.Fatalf("LoadConfig() error = %v, want nil", err)
	}
	if cfg.SyncInterval <= 0 {
		t.Errorf("SyncInterval = %v, want a usable default", cfg.SyncInterval)
	}
}

func TestLoadConfigRejectsAnInvalidSyncInterval(t *testing.T) {
	for _, value := range []string{"never", "0s", "-1m"} {
		_, err := LoadConfig(lookupFrom(map[string]string{
			"CHATCMS_DATABASE_URL":    "postgres:///chatcms",
			"CHATCMS_SECRET_KEY_FILE": "/run/secrets/master-key",
			"CHATCMS_SYNC_INTERVAL":   value,
		}))
		if err == nil {
			t.Errorf("LoadConfig() with %q error = nil, want a rejection", value)
		}
	}
}

func TestSitesDirIsUnderTheStateDirectory(t *testing.T) {
	cfg := Config{StateDir: "/var/lib/chatcms"}

	if got := cfg.SitesDir(); got != "/var/lib/chatcms/sites" {
		t.Errorf("SitesDir() = %q, want it under the state directory", got)
	}
}

func TestHostOf(t *testing.T) {
	cases := map[string]string{
		"https://chatcms.eu":      "chatcms.eu",
		"http://localhost:8080":   "localhost",
		"https://CHATCMS.EU/path": "chatcms.eu",
		"not a url":               "chatcms.local",
		"":                        "chatcms.local",
	}

	for in, want := range cases {
		if got := hostOf(in); got != want {
			t.Errorf("hostOf(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLoadConfigRequiresTheSecretKeyFile(t *testing.T) {
	_, err := LoadConfig(lookupFrom(map[string]string{
		"CHATCMS_DATABASE_URL": "postgres:///chatcms",
	}))

	var configErr *ConfigError
	if !errors.As(err, &configErr) {
		t.Fatalf("LoadConfig() error = %v, want a *ConfigError", err)
	}
	if configErr.Variable != "CHATCMS_SECRET_KEY_FILE" {
		t.Errorf("Variable = %q, want CHATCMS_SECRET_KEY_FILE", configErr.Variable)
	}
}

func TestBuildSettingsHaveDefaults(t *testing.T) {
	cfg, err := LoadConfig(lookupFrom(map[string]string{
		"CHATCMS_DATABASE_URL":    "postgres:///chatcms",
		"CHATCMS_SECRET_KEY_FILE": "/run/secrets/master-key",
	}))
	if err != nil {
		t.Fatalf("LoadConfig() error = %v, want nil", err)
	}

	if cfg.BwrapBin != "bwrap" || cfg.PrlimitBin != "prlimit" {
		t.Errorf("sandbox tools = %q and %q, want the plain names", cfg.BwrapBin, cfg.PrlimitBin)
	}
	if cfg.BuildTimeout != 10*time.Minute {
		t.Errorf("BuildTimeout = %v, want 10m", cfg.BuildTimeout)
	}
	if cfg.BuildMemory != 2<<30 {
		t.Errorf("BuildMemory = %d, want two gibibytes", cfg.BuildMemory)
	}
}

func TestBuildSettingsComeFromTheEnvironment(t *testing.T) {
	cfg, err := LoadConfig(lookupFrom(map[string]string{
		"CHATCMS_DATABASE_URL":    "postgres:///chatcms",
		"CHATCMS_SECRET_KEY_FILE": "/run/secrets/master-key",
		"CHATCMS_BWRAP_BIN":       "/nix/store/bwrap",
		"CHATCMS_PRLIMIT_BIN":     "/nix/store/prlimit",
		"CHATCMS_BUILD_TIMEOUT":   "90s",
		"CHATCMS_BUILD_MEMORY":    "512MiB",
	}))
	if err != nil {
		t.Fatalf("LoadConfig() error = %v, want nil", err)
	}

	if cfg.BwrapBin != "/nix/store/bwrap" || cfg.PrlimitBin != "/nix/store/prlimit" {
		t.Errorf("sandbox tools = %q and %q", cfg.BwrapBin, cfg.PrlimitBin)
	}
	if cfg.BuildTimeout != 90*time.Second {
		t.Errorf("BuildTimeout = %v, want 90s", cfg.BuildTimeout)
	}
	if cfg.BuildMemory != 512<<20 {
		t.Errorf("BuildMemory = %d, want 512 mebibytes", cfg.BuildMemory)
	}
}

func TestParseBytes(t *testing.T) {
	good := map[string]int64{
		"1024":    1024,
		"1024B":   1024,
		"2K":      2 << 10,
		"2KiB":    2 << 10,
		"512MiB":  512 << 20,
		"512 MiB": 512 << 20,
		"3G":      3 << 30,
		"4GiB":    4 << 30,
	}
	for raw, want := range good {
		got, err := ParseBytes(raw)
		if err != nil {
			t.Errorf("ParseBytes(%q) error = %v", raw, err)
			continue
		}
		if got != want {
			t.Errorf("ParseBytes(%q) = %d, want %d", raw, got, want)
		}
	}

	for _, raw := range []string{"", "lots", "0", "-5", "-1GiB", "1.5GiB", "GiB"} {
		if _, err := ParseBytes(raw); err == nil {
			t.Errorf("ParseBytes(%q) error = nil, want a rejection", raw)
		}
	}
}

func TestAnInvalidBuildMemoryNamesTheVariable(t *testing.T) {
	_, err := LoadConfig(lookupFrom(map[string]string{
		"CHATCMS_DATABASE_URL":    "postgres:///chatcms",
		"CHATCMS_SECRET_KEY_FILE": "/run/secrets/master-key",
		"CHATCMS_BUILD_MEMORY":    "as much as it likes",
	}))

	var configErr *ConfigError
	if !errors.As(err, &configErr) {
		t.Fatalf("LoadConfig() error = %v, want a *ConfigError", err)
	}
	if configErr.Variable != "CHATCMS_BUILD_MEMORY" {
		t.Errorf("Variable = %q, want CHATCMS_BUILD_MEMORY", configErr.Variable)
	}
}

func TestWorkflowSettingsHaveDefaults(t *testing.T) {
	cfg, err := LoadConfig(lookupFrom(map[string]string{
		"CHATCMS_DATABASE_URL":    "postgres:///chatcms",
		"CHATCMS_SECRET_KEY_FILE": "/run/secrets/master-key",
	}))
	if err != nil {
		t.Fatalf("LoadConfig() error = %v, want nil", err)
	}

	if cfg.CheckTimeout != 5*time.Minute {
		t.Errorf("CheckTimeout = %v, want 5m", cfg.CheckTimeout)
	}
	if cfg.CheckMemory != 1<<30 {
		t.Errorf("CheckMemory = %d, want one gibibyte", cfg.CheckMemory)
	}
	if len(cfg.WorkflowTools) != 0 {
		t.Errorf("WorkflowTools = %v, want none until Nix supplies them", cfg.WorkflowTools)
	}
}

func TestTheWorkflowToolboxIsReadLikeAPath(t *testing.T) {
	cfg, err := LoadConfig(lookupFrom(map[string]string{
		"CHATCMS_DATABASE_URL":    "postgres:///chatcms",
		"CHATCMS_SECRET_KEY_FILE": "/run/secrets/master-key",
		"CHATCMS_WORKFLOW_TOOLS":  "/nix/store/bash/bin/sh:/nix/store/coreutils/bin/cat: :/nix/store/gnugrep/bin/grep",
		"CHATCMS_CHECK_TIMEOUT":   "45s",
		"CHATCMS_CHECK_MEMORY":    "256MiB",
	}))
	if err != nil {
		t.Fatalf("LoadConfig() error = %v, want nil", err)
	}

	want := []string{"/nix/store/bash/bin/sh", "/nix/store/coreutils/bin/cat", "/nix/store/gnugrep/bin/grep"}
	if len(cfg.WorkflowTools) != len(want) {
		t.Fatalf("WorkflowTools = %v, want %v", cfg.WorkflowTools, want)
	}
	for index, tool := range want {
		if cfg.WorkflowTools[index] != tool {
			t.Errorf("tool %d = %q, want %q", index, cfg.WorkflowTools[index], tool)
		}
	}
	if cfg.CheckTimeout != 45*time.Second || cfg.CheckMemory != 256<<20 {
		t.Errorf("check limits = %v and %d", cfg.CheckTimeout, cfg.CheckMemory)
	}
}

func TestAnInvalidCheckMemoryNamesItsOwnVariable(t *testing.T) {
	_, err := LoadConfig(lookupFrom(map[string]string{
		"CHATCMS_DATABASE_URL":    "postgres:///chatcms",
		"CHATCMS_SECRET_KEY_FILE": "/run/secrets/master-key",
		"CHATCMS_CHECK_MEMORY":    "plenty",
	}))

	var configErr *ConfigError
	if !errors.As(err, &configErr) {
		t.Fatalf("LoadConfig() error = %v, want a *ConfigError", err)
	}
	if configErr.Variable != "CHATCMS_CHECK_MEMORY" {
		t.Errorf("Variable = %q, want CHATCMS_CHECK_MEMORY", configErr.Variable)
	}
}

func TestTheTemplateFeed(t *testing.T) {
	base := map[string]string{
		"CHATCMS_DATABASE_URL":    "postgres:///chatcms",
		"CHATCMS_SECRET_KEY_FILE": "/run/secrets/master-key",
	}
	with := func(value string) (Config, error) {
		pairs := map[string]string{}
		for k, v := range base {
			pairs[k] = v
		}
		if value != "<unset>" {
			pairs["CHATCMS_TEMPLATE_FEED"] = value
		}
		return LoadConfig(lookupFrom(pairs))
	}

	for value, want := range map[string]string{
		"<unset>":                             "https://quiqr.github.io/quiqr-community-templates/templates.json",
		"off":                                 "",
		"https://templates.example/feed.json": "https://templates.example/feed.json",
	} {
		cfg, err := with(value)
		if err != nil || cfg.TemplateFeed != want {
			t.Errorf("CHATCMS_TEMPLATE_FEED=%q gave %q, %v; want %q", value, cfg.TemplateFeed, err, want)
		}
	}

	for _, bad := range []string{"file:///etc/feed.json", "templates.json"} {
		var configErr *ConfigError
		if _, err := with(bad); !errors.As(err, &configErr) || configErr.Variable != "CHATCMS_TEMPLATE_FEED" {
			t.Errorf("CHATCMS_TEMPLATE_FEED=%q gave %v, want a configuration error", bad, err)
		}
	}
}

func workspaceConfig(t *testing.T, extra map[string]string) (Config, error) {
	t.Helper()

	pairs := map[string]string{
		"CHATCMS_DATABASE_URL":    "postgres:///chatcms",
		"CHATCMS_SECRET_KEY_FILE": "/run/secrets/master-key",
	}
	for key, value := range extra {
		pairs[key] = value
	}
	return LoadConfig(lookupFrom(pairs))
}

func TestWorkspaceSettingsHaveDefaults(t *testing.T) {
	cfg, err := workspaceConfig(t, nil)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.WorkspaceIdle != 15*time.Minute || cfg.WorkspacesAtOnce != 8 || len(cfg.Deprecated) != 0 {
		t.Errorf("got %v, %d, %v, want 15m, 8 and nothing renamed", cfg.WorkspaceIdle, cfg.WorkspacesAtOnce, cfg.Deprecated)
	}
}

func TestWorkspaceSettingsComeFromTheEnvironment(t *testing.T) {
	cfg, err := workspaceConfig(t, map[string]string{
		"CHATCMS_WORKSPACE_IDLE":     "5m",
		"CHATCMS_WORKSPACES_AT_ONCE": "3",
		"CHATCMS_PREVIEW_IDLE":       "1h",
		"CHATCMS_PREVIEWS_AT_ONCE":   "30",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.WorkspaceIdle != 5*time.Minute || cfg.WorkspacesAtOnce != 3 || len(cfg.Deprecated) != 0 {
		t.Errorf("got %v, %d, %v, want the new names to win", cfg.WorkspaceIdle, cfg.WorkspacesAtOnce, cfg.Deprecated)
	}
}

func TestTheOldPreviewNamesAreStillRead(t *testing.T) {
	cfg, err := workspaceConfig(t, map[string]string{
		"CHATCMS_PREVIEW_IDLE":     "1h",
		"CHATCMS_PREVIEWS_AT_ONCE": "30",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.WorkspaceIdle != time.Hour || cfg.WorkspacesAtOnce != 30 {
		t.Errorf("got %v, %d, want the old names read", cfg.WorkspaceIdle, cfg.WorkspacesAtOnce)
	}
	if len(cfg.Deprecated) != 2 || !strings.Contains(cfg.Deprecated[0], "CHATCMS_WORKSPACE_IDLE") {
		t.Errorf("Deprecated = %v, want both renames named", cfg.Deprecated)
	}
}

func TestAnInvalidWorkspaceSettingNamesTheVariableUsed(t *testing.T) {
	for variable, value := range map[string]string{
		"CHATCMS_WORKSPACE_IDLE":     "soon",
		"CHATCMS_WORKSPACES_AT_ONCE": "0",
		"CHATCMS_PREVIEWS_AT_ONCE":   "many",
	} {
		_, err := workspaceConfig(t, map[string]string{variable: value})
		var configErr *ConfigError
		if !errors.As(err, &configErr) || configErr.Variable != variable {
			t.Errorf("%s=%s error = %v, want it named", variable, value, err)
		}
	}
}

func TestTheWorkspaceBackendIsLocalUnlessChosen(t *testing.T) {
	cfg, err := workspaceConfig(t, nil)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.WorkspaceBackend != "local" || cfg.WorkspaceTaps != cfg.WorkspacesAtOnce || cfg.NixBin != "" || cfg.WorkspaceRunDir != "/run/chatcms/w" {
		t.Errorf("got %+v", cfg)
	}
}

func TestTheMicroVMBackendNeedsItsGuest(t *testing.T) {
	_, err := workspaceConfig(t, map[string]string{"CHATCMS_WORKSPACE_BACKEND": "microvm"})
	var configErr *ConfigError
	if !errors.As(err, &configErr) || configErr.Variable != "CHATCMS_WORKSPACE_GUEST" {
		t.Errorf("error = %v, want the guest named", err)
	}

	cfg, err := workspaceConfig(t, map[string]string{
		"CHATCMS_WORKSPACE_BACKEND": "microvm",
		"CHATCMS_WORKSPACE_GUEST":   "/nix/store/x-guest.json",
		"CHATCMS_WORKSPACE_TAPS":    "4",
		"CHATCMS_NIX_BIN":           "/run/current-system/sw/bin/nix",
	})
	if err != nil || cfg.WorkspaceTaps != 4 || cfg.WorkspaceGuest != "/nix/store/x-guest.json" || cfg.NixBin == "" {
		t.Errorf("got %+v, %v", cfg, err)
	}
}

func TestAnUnusableWorkspaceSettingIsNamed(t *testing.T) {
	for variable, value := range map[string]string{
		"CHATCMS_WORKSPACE_BACKEND": "firecracker",
		"CHATCMS_WORKSPACE_TAPS":    "0",
	} {
		_, err := workspaceConfig(t, map[string]string{variable: value})
		var configErr *ConfigError
		if !errors.As(err, &configErr) || configErr.Variable != variable {
			t.Errorf("%s=%s error = %v, want it named", variable, value, err)
		}
	}
}

func TestTheAIGatewayIsOffUntilItHasAProviderAKeyAndAModel(t *testing.T) {
	cfg, err := workspaceConfig(t, nil)
	if err != nil || cfg.AIConfigured() || cfg.AIListen != "127.0.0.1:8091" {
		t.Errorf("got %+v, %v, want the gateway off", cfg, err)
	}

	cfg, err = workspaceConfig(t, map[string]string{
		"CHATCMS_AI_PROVIDER_URL": "https://api.deepseek.com/v1",
		"CHATCMS_AI_KEY":          "sk-test",
		"CHATCMS_AI_MODEL":        "deepseek-chat",
	})
	if err != nil || !cfg.AIConfigured() {
		t.Errorf("got %+v, %v, want the gateway on", cfg, err)
	}

	for missing, given := range map[string]map[string]string{
		"CHATCMS_AI_KEY":          {"CHATCMS_AI_PROVIDER_URL": "https://x", "CHATCMS_AI_MODEL": "m"},
		"CHATCMS_AI_PROVIDER_URL": {"CHATCMS_AI_KEY": "k", "CHATCMS_AI_MODEL": "m"},
		"CHATCMS_AI_MODEL":        {"CHATCMS_AI_PROVIDER_URL": "https://x", "CHATCMS_AI_KEY": "k"},
	} {
		_, err := workspaceConfig(t, given)
		var configErr *ConfigError
		if !errors.As(err, &configErr) || configErr.Variable != missing {
			t.Errorf("without %s error = %v", missing, err)
		}
	}
}

func TestThePreviewOriginMustBeAnotherAddress(t *testing.T) {
	cfg, err := workspaceConfig(t, map[string]string{"CHATCMS_PREVIEW_URL": "https://preview.chatcms.example"})
	if err != nil || cfg.PreviewURL != "https://preview.chatcms.example" {
		t.Errorf("got %q, %v", cfg.PreviewURL, err)
	}
	for _, bad := range []string{"preview.example", "ftp://preview.example", "https://preview.example/sub", "http://localhost:8080"} {
		_, err := workspaceConfig(t, map[string]string{"CHATCMS_PREVIEW_URL": bad})
		var configErr *ConfigError
		if !errors.As(err, &configErr) || configErr.Variable != "CHATCMS_PREVIEW_URL" {
			t.Errorf("%s error = %v", bad, err)
		}
	}
}

func TestProvidedPackagesAreNamesAndStorePaths(t *testing.T) {
	cfg, err := workspaceConfig(t, map[string]string{"CHATCMS_PACKAGES": "quarto=/nix/store/a-quarto-1.10.18, typst=/nix/store/b-typst-0.15.1"})
	if err != nil || cfg.Packages["quarto"] != "/nix/store/a-quarto-1.10.18" || cfg.Packages["typst"] != "/nix/store/b-typst-0.15.1" {
		t.Errorf("Packages = %v, %v", cfg.Packages, err)
	}
	for _, bad := range []string{"quarto", "quarto=relative/path", "=/nix/store/x", "quarto=/nix/store/../etc"} {
		_, err := workspaceConfig(t, map[string]string{"CHATCMS_PACKAGES": bad})
		var configErr *ConfigError
		if !errors.As(err, &configErr) || configErr.Variable != "CHATCMS_PACKAGES" {
			t.Errorf("%q error = %v", bad, err)
		}
	}
}
