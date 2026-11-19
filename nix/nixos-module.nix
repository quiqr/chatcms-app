{ self }:
{ config, lib, pkgs, ... }:

let
  cfg = config.services.chatcms;
  system = pkgs.stdenv.hostPlatform.system;
  microvm = cfg.workspaceBackend == "microvm";
  aiOn = cfg.ai.providerUrl != null;
  taps = lib.range 0 (cfg.workspacesAtOnce - 1);
  tapsUp = pkgs.writeShellScript "chatcms-taps-up" ''
    set -eu
    ${lib.concatMapStrings (i: ''
      ${pkgs.iproute2}/bin/ip link show chv${toString i} >/dev/null 2>&1 \
        || ${pkgs.iproute2}/bin/ip tuntap add chv${toString i} mode tap user ${cfg.user} group ${cfg.group}
      ${pkgs.iproute2}/bin/ip addr replace 10.231.${toString i}.1/30 dev chv${toString i}
      ${pkgs.iproute2}/bin/ip link set chv${toString i} up
    '') taps}
    ${pkgs.nftables}/bin/nft -f ${pkgs.writeText "chatcms-workspaces.nft" ''
      table inet chatcms_workspaces
      delete table inet chatcms_workspaces
      table inet chatcms_workspaces {
        chain forward {
          type filter hook forward priority -10; policy accept;
          iifname "chv*" drop
          oifname "chv*" drop
        }
        chain input {
          type filter hook input priority -10; policy accept;
          iifname "chv*" ct state established,related accept
          ${lib.optionalString aiOn ''iifname "chv*" tcp dport ${toString cfg.ai.port} accept''}
          iifname "chv*" drop
        }
      }
    ''}
  '';
  tapsDown = pkgs.writeShellScript "chatcms-taps-down" ''
    ${pkgs.nftables}/bin/nft delete table inet chatcms_workspaces 2>/dev/null || true
    ${lib.concatMapStrings (i: ''
      ${pkgs.iproute2}/bin/ip link del chv${toString i} 2>/dev/null || true
    '') taps}
  '';
in
{
  options.services.chatcms = {
    enable = lib.mkEnableOption "Quiqr ChatCMS server";

    package = lib.mkOption {
      type = lib.types.package;
      default = self.packages.${pkgs.stdenv.hostPlatform.system}.chatcmsd;
      defaultText = lib.literalExpression "self.packages.\${system}.chatcmsd";
      description = "The chatcmsd package to run.";
    };

    listenAddress = lib.mkOption {
      type = lib.types.str;
      default = "127.0.0.1";
      description = "Address chatcmsd binds to.";
    };

    port = lib.mkOption {
      type = lib.types.port;
      default = 8080;
      description = "Port chatcmsd listens on.";
    };

    baseUrl = lib.mkOption {
      type = lib.types.str;
      example = "https://chatcms.eu";
      description = "Public base URL, used in invitation links and previews.";
    };

    packages = lib.mkOption {
      type = lib.types.attrsOf lib.types.package;
      default = { };
      example = lib.literalExpression "{ quarto = pkgs.quarto; }";
      description = "Packages this installation provides itself, by name. A site that asks for one at the same version gets it without a build.";
    };

    previewUrl = lib.mkOption {
      type = lib.types.nullOr lib.types.str;
      default = null;
      example = "https://preview.chatcms.eu";
      description = "An origin of its own for previews, on another host than baseUrl, so a site's scripts never run beside the app. Point its DNS and proxy at the same server.";
    };

    stateDir = lib.mkOption {
      type = lib.types.path;
      default = "/var/lib/chatcms";
      description = "Directory holding project working copies, media caches and build output.";
    };

    user = lib.mkOption {
      type = lib.types.str;
      default = "chatcms";
      description = "User the service runs as.";
    };

    group = lib.mkOption {
      type = lib.types.str;
      default = "chatcms";
      description = "Group the service runs as.";
    };

    database = {
      createLocally = lib.mkOption {
        type = lib.types.bool;
        default = true;
        description = "Provision a local PostgreSQL database and role for ChatCMS.";
      };

      name = lib.mkOption {
        type = lib.types.str;
        default = "chatcms";
        description = "PostgreSQL database name.";
      };

      urlFile = lib.mkOption {
        type = lib.types.nullOr lib.types.path;
        default = null;
        description = ''
          File containing the PostgreSQL connection string. Required when
          database.createLocally is false. Read at start-up, never put in the
          Nix store.
        '';
      };
    };

    secretKeyFile = lib.mkOption {
      type = lib.types.path;
      description = ''
        File containing the master key used to encrypt per-project SSH deploy
        keys and session secrets at rest. Never put in the Nix store.
      '';
    };

    firstAdministrator = lib.mkOption {
      type = lib.types.nullOr lib.types.str;
      default = null;
      example = "pim@example.org";
      description = ''
        Email address of the person who administers this installation. They
        are made an administrator every time the service starts, which is how
        a new installation gets its first one: nobody is there to appoint
        anybody. The account must already exist; until it does, the service
        starts and says it could not appoint anybody.
      '';
    };

    mail = {
      host = lib.mkOption {
        type = lib.types.nullOr lib.types.str;
        default = null;
        example = "smtp.example.org";
        description = ''
          Mail server invitations and password resets are sent through. With
          none, those links go to the service log instead, which suits an
          installation run by the person who owns it and nothing larger.
        '';
      };

      port = lib.mkOption {
        type = lib.types.port;
        default = 587;
        description = "Port on the mail server.";
      };

      from = lib.mkOption {
        type = lib.types.nullOr lib.types.str;
        default = null;
        example = "chatcms@example.org";
        description = ''
          Address messages come from. Defaults to chatcms@ the host of
          baseUrl, which is what a receiving server is most likely to accept.
        '';
      };

      fromName = lib.mkOption {
        type = lib.types.str;
        default = "Quiqr ChatCMS";
        description = "Name a reader sees instead of the address.";
      };

      username = lib.mkOption {
        type = lib.types.nullOr lib.types.str;
        default = null;
        description = "Username for the mail server, when it wants one.";
      };

      passwordFile = lib.mkOption {
        type = lib.types.nullOr lib.types.path;
        default = null;
        description = ''
          File holding the mail server password. Never put in the Nix store;
          systemd reads it and hands the service a copy only it can read.
        '';
      };

      startTLS = lib.mkOption {
        type = lib.types.bool;
        default = true;
        description = ''
          Ask for the connection to be encrypted. Leave it on for anything
          reachable over a network.
        '';
      };
    };

    hugoPackage = lib.mkOption {
      type = lib.types.package;
      default = pkgs.hugo;
      description = "Hugo used for builds and previews.";
    };

    gitPackage = lib.mkOption {
      type = lib.types.package;
      default = pkgs.git;
      description = "Git CLI used for all repository operations.";
    };

    bubblewrapPackage = lib.mkOption {
      type = lib.types.package;
      default = pkgs.bubblewrap;
      description = ''
        Bubblewrap, which isolates every site build. Without it, builds are
        refused rather than run unisolated.
      '';
    };

    utilLinuxPackage = lib.mkOption {
      type = lib.types.package;
      default = pkgs.util-linux;
      description = "Provides prlimit, which bounds the memory a build may use.";
    };

    templateFeed = lib.mkOption {
      type = lib.types.nullOr lib.types.str;
      default = "https://quiqr.github.io/quiqr-community-templates/templates.json";
      example = "https://templates.example.org/templates.json";
      description = ''
        The template feed a new project can start from, in the format of the
        Quiqr community feed. Null switches the template gallery off.
      '';
    };

    buildTimeout = lib.mkOption {
      type = lib.types.str;
      default = "10m";
      example = "2m";
      description = "How long a site build may run before it is ended.";
    };

    workflowTools = lib.mkOption {
      type = lib.types.listOf lib.types.package;
      default = with pkgs; [ bash coreutils gnugrep gnused findutils diffutils ];
      description = ''
        What a site's own check step may use. Fixed and pinned: a workflow
        cannot ask for another tool, because fetching software during a
        publish is exactly what the sandbox exists to prevent.
      '';
    };

    checkTimeout = lib.mkOption {
      type = lib.types.str;
      default = "5m";
      description = "How long a site's own check may run before it is ended.";
    };

    checkMemory = lib.mkOption {
      type = lib.types.str;
      default = "1GiB";
      description = "How much memory a site's own check may use.";
    };

    buildMemory = lib.mkOption {
      type = lib.types.str;
      default = "2GiB";
      example = "512MiB";
      description = ''
        How much memory a site build may use. A build that exceeds it is
        ended and its log is kept.
      '';
    };

    workspaceIdle = lib.mkOption {
      type = lib.types.str;
      default = "15m";
      description = "How long a workspace nobody uses is kept warm before it sleeps.";
    };

    workspacesAtOnce = lib.mkOption {
      type = lib.types.ints.positive;
      default = 8;
      description = ''
        How many workspaces may be starting or warm at once. Each one runs a
        preview of one person's copy of a site.
      '';
    };

    workspaceBackend = lib.mkOption {
      type = lib.types.enum [ "local" "microvm" ];
      default = "local";
      description = ''
        Where workspaces run. `local` runs each preview as a process of the
        server. `microvm` runs each in a Cloud Hypervisor micro VM of its own,
        which needs KVM; the module then makes a TAP device and an address per
        workspace and a firewall table that lets the VMs reach nothing but the
        host.
      '';
    };

    pinHugo = lib.mkOption {
      type = lib.types.bool;
      default = true;
      description = ''
        Run each project's own Hugo version, found through the Nixhub index
        and built with Nix. Off, every project uses `hugoPackage`.
      '';
    };

    nixhubUrl = lib.mkOption {
      type = lib.types.str;
      default = "https://search.devbox.sh";
      description = "The package index a project's Hugo version is looked up in.";
    };

    ai = {
      providerUrl = lib.mkOption {
        type = lib.types.nullOr lib.types.str;
        default = null;
        example = "https://api.deepseek.com/v1";
        description = ''
          An OpenAI-compatible endpoint the assistant's model is reached at.
          Null switches the AI gateway off.
        '';
      };

      model = lib.mkOption {
        type = lib.types.str;
        default = "";
        example = "deepseek-chat";
        description = "The model every workspace's requests go to.";
      };

      keyFile = lib.mkOption {
        type = lib.types.nullOr lib.types.path;
        default = null;
        description = ''
          A file holding the provider's API key. It is read by systemd and
          handed to the service; it never enters a workspace.
        '';
      };

      port = lib.mkOption {
        type = lib.types.port;
        default = 8091;
        description = "The port the AI gateway listens on.";
      };

      sessionHeader = lib.mkOption {
        type = lib.types.nullOr lib.types.str;
        default = null;
        example = "x-opencode-session";
        description = "A header the provider wants a stable session in; the gateway sends one per person and project. OpenCode Go asks for x-opencode-session.";
      };
    };

    extraEnvironment = lib.mkOption {
      type = lib.types.attrsOf lib.types.str;
      default = { };
      description = "Extra environment variables for the service.";
    };
  };

  config = lib.mkIf cfg.enable {
    networking.firewall.interfaces = lib.mkIf (microvm && aiOn) {
      "chv+".allowedTCPPorts = [ cfg.ai.port ];
    };

    assertions = [
      {
        assertion = !aiOn || (cfg.ai.keyFile != null && cfg.ai.model != "");
        message = "services.chatcms: set ai.keyFile and ai.model when ai.providerUrl is set.";
      }
      {
        assertion = cfg.database.createLocally || cfg.database.urlFile != null;
        message = "services.chatcms: set database.urlFile when database.createLocally is false.";
      }
    ];

    users.users = lib.mkIf (cfg.user == "chatcms") {
      chatcms = {
        isSystemUser = true;
        group = cfg.group;
        home = cfg.stateDir;
      };
    };

    users.groups = lib.mkIf (cfg.group == "chatcms") {
      chatcms = { };
    };

    services.postgresql = lib.mkIf cfg.database.createLocally {
      enable = true;
      ensureDatabases = [ cfg.database.name ];
      ensureUsers = [
        {
          name = cfg.user;
          ensureDBOwnership = true;
        }
      ];
    };

    systemd.services.chatcms = {
      description = "Quiqr ChatCMS server";
      wantedBy = [ "multi-user.target" ];
      after = [ "network-online.target" ] ++ lib.optional cfg.database.createLocally "postgresql.service";
      wants = [ "network-online.target" ];
      requires = lib.optional cfg.database.createLocally "postgresql.service";

      path = [
        cfg.gitPackage
        cfg.hugoPackage
        cfg.bubblewrapPackage
        cfg.utilLinuxPackage
        pkgs.openssh
      ];

      environment = {
        CHATCMS_LISTEN = "${cfg.listenAddress}:${toString cfg.port}";
        CHATCMS_BASE_URL = cfg.baseUrl;
      } // lib.optionalAttrs (cfg.packages != { }) {
        CHATCMS_PACKAGES = lib.concatStringsSep "," (lib.mapAttrsToList (name: pkg: "${name}=${pkg}") cfg.packages);
      } // lib.optionalAttrs (cfg.previewUrl != null) {
        CHATCMS_PREVIEW_URL = cfg.previewUrl;
      } // {
        CHATCMS_STATE_DIR = cfg.stateDir;
        CHATCMS_GIT_BIN = lib.getExe cfg.gitPackage;
        CHATCMS_HUGO_BIN = lib.getExe cfg.hugoPackage;
        # systemd reads the key as root and hands the service a copy only it
        # can read, so the key on disk stays owned by root and the service
        # does not run as root to read it. %d is where systemd puts it.
        CHATCMS_SECRET_KEY_FILE = "%d/master-key";
        CHATCMS_BWRAP_BIN = "${cfg.bubblewrapPackage}/bin/bwrap";
        CHATCMS_PRLIMIT_BIN = "${cfg.utilLinuxPackage}/bin/prlimit";
        CHATCMS_BUILD_TIMEOUT = cfg.buildTimeout;
        CHATCMS_BUILD_MEMORY = cfg.buildMemory;
        CHATCMS_WORKFLOW_TOOLS =
          lib.concatStringsSep ":" (map (tool: "${tool}/bin") cfg.workflowTools);
        CHATCMS_CHECK_TIMEOUT = cfg.checkTimeout;
        CHATCMS_CHECK_MEMORY = cfg.checkMemory;
        CHATCMS_WORKSPACE_IDLE = cfg.workspaceIdle;
        CHATCMS_WORKSPACES_AT_ONCE = toString cfg.workspacesAtOnce;
        CHATCMS_WORKSPACE_BACKEND = cfg.workspaceBackend;
        CHATCMS_TEMPLATE_FEED = if cfg.templateFeed == null then "off" else cfg.templateFeed;
      } // lib.optionalAttrs cfg.pinHugo {
        CHATCMS_NIX_BIN = "${config.nix.package}/bin/nix";
        CHATCMS_NIXHUB_URL = cfg.nixhubUrl;
        XDG_CACHE_HOME = "${cfg.stateDir}/cache";
      } // lib.optionalAttrs microvm {
        CHATCMS_WORKSPACE_GUEST = "${self.packages.${system}.workspace-guest}";
        CHATCMS_CLOUD_HYPERVISOR_BIN = "${pkgs.cloud-hypervisor}/bin/cloud-hypervisor";
        CHATCMS_VIRTIOFSD_BIN = "${pkgs.virtiofsd}/bin/virtiofsd";
        CHATCMS_WORKSPACE_TAPS = toString cfg.workspacesAtOnce;
        CHATCMS_WORKSPACE_RUN_DIR = "/run/chatcms/w";
      } // lib.optionalAttrs aiOn {
        CHATCMS_AI_PROVIDER_URL = cfg.ai.providerUrl;
        CHATCMS_AI_MODEL = cfg.ai.model;
        CHATCMS_AI_KEY_FILE = "%d/ai-key";
        CHATCMS_AI_LISTEN = "${if microvm then "0.0.0.0" else "127.0.0.1"}:${toString cfg.ai.port}";
        CHATCMS_OPENCODE_BIN = "${self.packages.${system}.opencode}/bin/opencode";
      } // lib.optionalAttrs (aiOn && cfg.ai.sessionHeader != null) {
        CHATCMS_AI_SESSION_HEADER = cfg.ai.sessionHeader;
      } // lib.optionalAttrs (cfg.firstAdministrator != null) {
        CHATCMS_FIRST_ADMINISTRATOR = cfg.firstAdministrator;
      } // lib.optionalAttrs (cfg.mail.host != null) {
        CHATCMS_MAIL_HOST = cfg.mail.host;
        CHATCMS_MAIL_PORT = toString cfg.mail.port;
        CHATCMS_MAIL_FROM_NAME = cfg.mail.fromName;
        CHATCMS_MAIL_STARTTLS = if cfg.mail.startTLS then "yes" else "no";
      } // lib.optionalAttrs (cfg.mail.from != null) {
        CHATCMS_MAIL_FROM = cfg.mail.from;
      } // lib.optionalAttrs (cfg.mail.username != null) {
        CHATCMS_MAIL_USERNAME = cfg.mail.username;
      } // lib.optionalAttrs (cfg.mail.passwordFile != null) {
        CHATCMS_MAIL_PASSWORD_FILE = "%d/mail-password";
      } // lib.optionalAttrs cfg.database.createLocally {
        CHATCMS_DATABASE_URL = "postgres:///${cfg.database.name}?host=/run/postgresql";
      } // lib.optionalAttrs (cfg.database.urlFile != null) {
        CHATCMS_DATABASE_URL_FILE = cfg.database.urlFile;
      } // cfg.extraEnvironment;

      serviceConfig = {
        ExecStart = "${lib.getExe cfg.package}";
        LoadCredential = [ "master-key:${cfg.secretKeyFile}" ]
          ++ lib.optional (aiOn && cfg.ai.keyFile != null) "ai-key:${cfg.ai.keyFile}"
          ++ lib.optional (cfg.mail.passwordFile != null) "mail-password:${cfg.mail.passwordFile}";
        User = cfg.user;
        Group = cfg.group;
        Restart = "on-failure";
        RestartSec = 5;
        StateDirectory = lib.mkIf (cfg.stateDir == "/var/lib/chatcms") "chatcms";
        WorkingDirectory = cfg.stateDir;
        ReadWritePaths = [ cfg.stateDir ];
        AmbientCapabilities = [ ];
        NoNewPrivileges = true;
        PrivateTmp = true;
        PrivateDevices = true;
        ProtectSystem = "strict";
        ProtectHome = true;
        ProtectKernelTunables = true;
        ProtectKernelModules = true;
        ProtectControlGroups = true;
        RestrictAddressFamilies = [ "AF_UNIX" "AF_INET" "AF_INET6" ];
        RestrictNamespaces = false;
        LockPersonality = true;
        SystemCallArchitectures = "native";
      } // lib.optionalAttrs microvm {
        PrivateDevices = false;
        DeviceAllow = [ "/dev/kvm rw" "/dev/net/tun rw" ];
        SupplementaryGroups = [ "kvm" ];
        RuntimeDirectory = "chatcms";
        RuntimeDirectoryMode = "0700";
        RestrictAddressFamilies = [ "AF_UNIX" "AF_INET" "AF_INET6" "AF_NETLINK" ];
      };
    };

    systemd.services.chatcms-taps = lib.mkIf microvm {
      description = "TAP devices and a firewall table for ChatCMS workspaces";
      wantedBy = [ "multi-user.target" ];
      before = [ "chatcms.service" ];
      requiredBy = [ "chatcms.service" ];
      after = [ "network-pre.target" ];
      serviceConfig = {
        Type = "oneshot";
        RemainAfterExit = true;
        ExecStart = tapsUp;
        ExecStop = tapsDown;
      };
    };
  };
}
