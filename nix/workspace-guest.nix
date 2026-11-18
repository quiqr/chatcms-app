{ lib, pkgs, opencode, ... }:
let
  preview = pkgs.writeShellScript "chatcms-preview" ''
    export PATH=${lib.makeBinPath [ pkgs.iproute2 pkgs.coreutils pkgs.gnugrep pkgs.gnused pkgs.util-linux ]}
    param() { tr ' ' '\n' < /proc/cmdline | grep "^chatcms.$1=" | head -n1 | sed "s/^chatcms.$1=//"; }
    address=$(param ip)
    gateway=$(param gateway)
    hugo=$(param hugo)
    under=$(param under)
    reload=$(param reload)
    ai=$(param ai)
    token=$(param token)

    if [ -n "$token" ]; then
      mkdir -p /run/chatcms
      ( umask 077; printf 'CHATCMS_AI_URL=%s\nCHATCMS_AI_TOKEN=%s\n' "$ai" "$token" > /run/chatcms/ai.env )
    fi

    link=$(ip -o link show | grep -v ' lo:' | head -n1 | sed 's/^[0-9]*: \([^:@]*\).*/\1/')
    ip link set lo up
    ip link set "$link" up
    ip addr replace "$address" dev "$link"
    ip route replace default via "$gateway" dev "$link"

    home=/run/preview
    mkdir -p $home/cache $home/resources
    chown -R 65534:65534 $home

    set -- server --source /site --bind 0.0.0.0 --port 1313 --baseURL "$under" \
      --appendPort=false --renderToMemory --buildDrafts --noBuildLock --poll 500ms \
      --cacheDir $home/cache
    if [ -n "$reload" ]; then set -- "$@" --liveReloadPort "$reload"; fi
    exec setpriv --reuid=65534 --regid=65534 --clear-groups \
      env HOME=$home HUGO_RESOURCEDIR=$home/resources TZ=UTC "$hugo" "$@"
  '';
  agent = pkgs.writeShellScript "chatcms-agent" ''
    export PATH=${lib.makeBinPath [ pkgs.coreutils pkgs.curl pkgs.util-linux ]}
    while [ ! -f /run/chatcms/ai.env ]; do sleep 0.2; done
    . /run/chatcms/ai.env
    base=''${CHATCMS_AI_URL%/v1}

    home=/run/agent
    mkdir -p $home
    until curl -sf -H "Authorization: Bearer $CHATCMS_AI_TOKEN" "$base/agent/env" -o $home/env; do sleep 1; done
    chown -R 65534:65534 $home
    chmod 0600 $home/env

    cd /site
    exec setpriv --reuid=65534 --regid=65534 --clear-groups \
      env -i ${pkgs.bash}/bin/bash -c "set -a; . $home/env; set +a; rm -f $home/env; exec ${opencode}/bin/opencode serve --hostname 0.0.0.0 --port 4096"
  '';
in
{
  system.stateVersion = "25.11";
  networking.hostName = "workspace";
  networking.useDHCP = false;
  networking.firewall.enable = false;

  microvm = {
    hypervisor = "cloud-hypervisor";
    vcpu = 2;
    mem = 1024;
    shares = [
      { proto = "virtiofs"; tag = "ro-store"; source = "/nix/store"; mountPoint = "/nix/.ro-store"; }
      { proto = "virtiofs"; tag = "site"; source = "/var/empty"; mountPoint = "/site"; readOnly = true; }
    ];
    interfaces = [ { type = "tap"; id = "chv-guest"; mac = "02:00:00:00:00:02"; } ];
  };

  fileSystems."/site".options = [ "ro" ];

  systemd.services.chatcms-preview = {
    description = "The preview of this workspace's site";
    wantedBy = [ "sysinit.target" ];
    after = [ "local-fs.target" ];
    requires = [ "local-fs.target" ];
    unitConfig.DefaultDependencies = false;
    serviceConfig = {
      Type = "simple";
      Restart = "on-failure";
      RestartSec = 1;
      ExecStart = preview;
    };
  };

  systemd.services.chatcms-agent = {
    description = "The assistant of this workspace";
    wantedBy = [ "sysinit.target" ];
    after = [ "local-fs.target" "chatcms-preview.service" ];
    requires = [ "local-fs.target" ];
    unitConfig.DefaultDependencies = false;
    unitConfig.ConditionKernelCommandLine = "chatcms.token";
    serviceConfig = {
      Type = "simple";
      Restart = "on-failure";
      RestartSec = 1;
      ExecStart = agent;
    };
  };

  boot.kernelParams = [ "quiet" ];
  services.journald.settings.Journal = {
    ForwardToConsole = true;
    TTYPath = "/dev/hvc0";
  };
  documentation.enable = false;
  nix.enable = false;
  security.sudo.enable = false;
  services.nscd.enable = false;
  system.nssModules = lib.mkForce [ ];
  services.timesyncd.enable = false;
  services.logrotate.enable = false;
  systemd.oomd.enable = false;
  systemd.services.systemd-journal-catalog-update.enable = false;
  systemd.services.mount-pstore.enable = false;
  environment.defaultPackages = [ ];
}
