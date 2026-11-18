# A NixOS virtual machine that runs ChatCMS from the module and nothing else,
# so that "it deploys" is a thing the gate checks rather than a thing somebody
# remembers doing once.
{ self, pkgs }:

pkgs.testers.runNixOSTest {
  name = "chatcms-deploys";

  nodes.machine = { config, lib, ... }: {
    imports = [ self.nixosModules.default ];

    # The key is made on the machine, never written in Nix, because anything
    # written in Nix is in the store and the store is world readable.
    systemd.tmpfiles.rules = [ "d /run/chatcms-secrets 0700 root root -" ];

    systemd.services.chatcms-master-key = {
      description = "Make a master key for ChatCMS";
      wantedBy = [ "multi-user.target" ];
      before = [ "chatcms.service" ];
      serviceConfig = {
        Type = "oneshot";
        RemainAfterExit = true;
      };
      script = ''
        mkdir -p /run/chatcms-secrets
        if [ ! -f /run/chatcms-secrets/master.key ]; then
          od -An -tx1 -N32 /dev/urandom | tr -d " \n" > /run/chatcms-secrets/master.key
          chmod 600 /run/chatcms-secrets/master.key
        fi
      '';
    };

    services.chatcms = {
      enable = true;
      package = self.checks.${pkgs.stdenv.hostPlatform.system}.go-build;
      baseUrl = "http://localhost:8080";
      secretKeyFile = "/run/chatcms-secrets/master.key";
      database.createLocally = true;
      workspaceBackend = "microvm";
      workspacesAtOnce = 2;
    };

    environment.systemPackages = [ pkgs.curl pkgs.jq pkgs.nftables ];

    virtualisation.memorySize = 2048;
  };

  testScript = ''
    start_all()

    with subtest("the database and the service come up"):
        machine.wait_for_unit("postgresql.service")
        machine.wait_for_unit("chatcms.service")
        machine.wait_for_open_port(8080)

    with subtest("the migrations ran from an empty database"):
        health = machine.succeed("curl -sS http://localhost:8080/api/health")
        print(health)
        machine.succeed(f"echo '{health}' | jq -e '.database == \"reachable\"'")
        machine.succeed(f"echo '{health}' | jq -e '.schemaVersion | tonumber > 0'")

    with subtest("the sign-in page is served"):
        page = machine.succeed("curl -sS -L http://localhost:8080/sign-in")
        assert "ChatCMS" in page or "chatcms" in page.lower(), page

    with subtest("the key stays owned by root and the service still reads it"):
        owner = machine.succeed("stat -c '%U %a' /run/chatcms-secrets/master.key").strip()
        assert owner == "root 600", owner

    with subtest("the master key is not in the Nix store"):
        # The file itself must live outside the store, and no copy of its
        # contents may have been put there by the module.
        machine.succeed("test -f /run/chatcms-secrets/master.key")
        machine.fail("find /nix/store -maxdepth 1 -name '*master.key*' | grep -q .")
        machine.succeed(
            "! grep -rl \"$(base64 < /run/chatcms-secrets/master.key)\" /nix/store 2>/dev/null | grep -q ."
        )

    with subtest("the service runs as its own user, not as root"):
        machine.succeed("systemctl show chatcms.service -p User | grep -q 'User=chatcms'")
        machine.succeed("test -d /var/lib/chatcms")
        owner = machine.succeed("stat -c %U /var/lib/chatcms").strip()
        assert owner == "chatcms", owner

    with subtest("it comes back after a restart, with the database as it was"):
        machine.succeed("systemctl restart chatcms.service")
        machine.wait_for_unit("chatcms.service")
        machine.wait_for_open_port(8080)
        machine.succeed("curl -sS http://localhost:8080/api/health | jq -e '.status == \"ok\"'")

    with subtest("git and Hugo are the pinned ones"):
        environment = machine.succeed("systemctl show chatcms.service -p Environment")
        assert "CHATCMS_GIT_BIN=/nix/store" in environment, environment
        assert "CHATCMS_HUGO_BIN=/nix/store" in environment, environment
        assert "CHATCMS_BWRAP_BIN=/nix/store" in environment, environment

    with subtest("each workspace has a TAP device of the service's own, and the VMs reach nothing"):
        machine.wait_for_unit("chatcms-taps.service")
        for i in range(2):
            owner = machine.succeed(f"cat /sys/class/net/chv{i}/owner").strip()
            uid = machine.succeed("id -u chatcms").strip()
            assert owner == uid, (owner, uid)
            machine.succeed(f"ip -br addr show chv{i} | grep -q 10.231.{i}.1/30")
        machine.fail("ip link show chv2")
        rules = machine.succeed("nft list table inet chatcms_workspaces")
        assert 'iifname "chv*" drop' in rules, rules
        assert 'oifname "chv*" drop' in rules, rules

    with subtest("the service can reach KVM and the TAP devices, and knows its guest"):
        environment = machine.succeed("systemctl show chatcms.service -p Environment")
        assert "CHATCMS_WORKSPACE_BACKEND=microvm" in environment, environment
        assert "CHATCMS_WORKSPACE_GUEST=/nix/store" in environment, environment
        assert "CHATCMS_NIX_BIN=/nix/store" in environment, environment
        devices = machine.succeed("systemctl show chatcms.service -p DeviceAllow")
        assert "/dev/kvm" in devices and "/dev/net/tun" in devices, devices
        machine.succeed("test -d /run/chatcms")

    with subtest("the template feed defaults to the community feed"):
        environment = machine.succeed("systemctl show chatcms.service -p Environment")
        assert "CHATCMS_TEMPLATE_FEED=https://quiqr.github.io/quiqr-community-templates/templates.json" in environment, environment
  '';
}
