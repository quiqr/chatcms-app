{ self, pkgs }:

pkgs.testers.runNixOSTest {
  name = "chatcms-gvisor";

  nodes.machine = { config, lib, ... }: {
    imports = [ self.nixosModules.default ];

    systemd.services.chatcms-master-key = {
      wantedBy = [ "multi-user.target" ];
      before = [ "chatcms.service" ];
      serviceConfig = {
        Type = "oneshot";
        RemainAfterExit = true;
      };
      script = ''
        mkdir -p /run/chatcms-secrets
        od -An -tx1 -N32 /dev/urandom | tr -d " \n" > /run/chatcms-secrets/master.key
        chmod 600 /run/chatcms-secrets/master.key
      '';
    };

    services.chatcms = {
      enable = true;
      package = self.checks.${pkgs.stdenv.hostPlatform.system}.go-build;
      baseUrl = "http://localhost:8080";
      secretKeyFile = "/run/chatcms-secrets/master.key";
      database.createLocally = true;
      workspaceBackend = "gvisor";
    };

    environment.systemPackages = [ pkgs.curl pkgs.jq ];
    virtualisation.memorySize = 2048;
  };

  testScript = ''
    start_all()

    with subtest("the service comes up with gVisor as its backend"):
        machine.wait_for_unit("chatcms.service")
        machine.wait_for_open_port(8080)
        environment = machine.succeed("systemctl show chatcms.service -p Environment")
        assert "CHATCMS_WORKSPACE_BACKEND=gvisor" in environment, environment
        assert "CHATCMS_RUNSC=/nix/store" in environment, environment
        assert "CHATCMS_RELAY=/nix/store" in environment, environment
        machine.succeed("test -d /var/lib/chatcms/w/runsc")

    with subtest("runsc runs rootless, without network, under the service's own fences"):
        runsc = machine.succeed("systemctl show chatcms.service -p Environment | tr ' ' '\\n' | grep '^CHATCMS_RUNSC=' | cut -d= -f2").strip()
        properties = " ".join(
            "-p " + p
            for p in [
                "User=chatcms",
                "NoNewPrivileges=yes",
                "PrivateDevices=yes",
                "PrivateTmp=yes",
                "ProtectSystem=strict",
                "ProtectHome=yes",
                "ProtectKernelTunables=yes",
                "ProtectKernelModules=yes",
                "ProtectControlGroups=yes",
                "ReadWritePaths=/var/lib/chatcms",
                "'RestrictAddressFamilies=AF_UNIX AF_INET AF_INET6'",
                "LockPersonality=yes",
                "SystemCallArchitectures=native",
            ]
        )
        said = machine.succeed(
            f"systemd-run --wait --pipe --quiet {properties} {runsc} --rootless --root=/var/lib/chatcms/probe --network=none do /run/current-system/sw/bin/echo inside"
        )
        assert "inside" in said, said
  '';
}
