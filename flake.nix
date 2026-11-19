{
  description = "Quiqr ChatCMS: hosted, Nix-powered headless CMS for Markdown static sites";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  inputs.microvm.url = "github:astro/microvm.nix";
  inputs.microvm.inputs.nixpkgs.follows = "nixpkgs";
  inputs.nixpkgs-opencode.url = "github:NixOS/nixpkgs/f9d8b65950353691ab56561e7c73d2e1063d810b";

  outputs = { self, nixpkgs, microvm, nixpkgs-opencode }:
    let
      version = nixpkgs.lib.fileContents ./VERSION;

      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];

      forAllSystems = f:
        nixpkgs.lib.genAttrs systems (system: f {
          inherit system;
          pkgs = nixpkgs.legacyPackages.${system};
          lib = nixpkgs.lib;
        });

      goSource = nixpkgs.lib.fileset.toSource {
        root = ./.;
        fileset = nixpkgs.lib.fileset.unions [
          ./go.mod
          ./go.sum
          ./VERSION
          ./LICENSE
          ./version.go
          ./version_test.go
          ./cmd
          ./db
          ./internal
          ./test/fixtures
          ./scripts/goldens
        ];
      };

      webSource = nixpkgs.lib.fileset.toSource {
        root = ./web;
        fileset = nixpkgs.lib.fileset.unions [
          ./web/package.json
          ./web/package-lock.json
          ./web/tsconfig.json
          ./web/vite.config.ts
          ./web/vite.preview.config.ts
          ./web/vitest.setup.ts
          ./web/index.html
          ./web/src
          ./web/public
          ./web/scripts
        ];
      };

      vendorHash = "sha256-iPidfJoWbvcLbSqecx8hCrA60T25InEnNuhAGrfAe84=";

      mkWeb = { pkgs, checked ? false, ... }: pkgs.buildNpmPackage {
        pname = "chatcms-web";
        inherit version;
        src = webSource;
        npmDepsHash = "sha256-sx0r3y4shTHJ3AG1n0SfR0HxtyzmEyhktTbZIbDCHWI=";

        nativeBuildInputs = [ pkgs.nodejs_22 ];

        env.ESBUILD_BINARY_PATH = pkgs.lib.getExe pkgs.esbuild;
        env.ROLLUP_SKIP_NODEJS_NATIVE = "1";

        doCheck = checked;
        checkPhase = ''
          runHook preCheck
          npm run typecheck
          npm run check:interface
          npm run check:strings
          npm run check:bundle
          npm run test -- --maxWorkers=4 --testTimeout=30000
          runHook postCheck
        '';

        installPhase = ''
          runHook preInstall
          mkdir -p $out
          cp -r dist/* $out/
          runHook postInstall
        '';

        meta.description = "Quiqr ChatCMS browser application";
      };

      mkServer = { pkgs, source ? "", ... }@args: pkgs.buildGoModule {
        pname = "chatcmsd";
        inherit version vendorHash;
        src = goSource;
        subPackages = [ "cmd/chatcmsd" ];
        doCheck = false;
        env.CGO_ENABLED = 0;

        preBuild = ''
          cp -r ${mkWeb args}/. internal/webui/dist/
        '';

        ldflags = [
          "-s"
          "-w"
          "-X=github.com/quiqr/chatcms-app/internal/version.Version=${version}"
        ] ++ nixpkgs.lib.optional (source != "")
          "-X=github.com/quiqr/chatcms-app/internal/version.Source=${source}";

        meta = {
          description = "Quiqr ChatCMS server";
          mainProgram = "chatcmsd";
        };
      };

      # workflowTools is what a site's own check may use. It is fixed and
      # pinned: a workflow cannot ask for another tool, because doing that
      # properly means shareable actions as Nix flakes, which milestone 1
      # puts out of scope.
      workflowToolPackages = pkgs: with pkgs; [ bash coreutils gnugrep gnused findutils diffutils ];

      workflowTools = pkgs: nixpkgs.lib.concatStringsSep ":"
        (map (tool: "${tool}/bin") (workflowToolPackages pkgs));

      mkGoTests = { pkgs, ... }@args: (mkServer args).overrideAttrs (old: {
        pname = "chatcmsd-tests";
        subPackages = null;
        preBuild = "";
        doCheck = true;
        env = old.env // { CGO_ENABLED = 1; };
        nativeBuildInputs = (old.nativeBuildInputs or [ ]) ++ [
          pkgs.git
          pkgs.openssh
          pkgs.bubblewrap
          pkgs.util-linux
          pkgs.hugo
          pkgs.procps
        ] ++ (workflowToolPackages pkgs);
        buildPhase = "runHook preBuild; runHook postBuild";
        checkPhase = ''
          runHook preCheck
          export HOME=$TMPDIR
          export CHATCMS_GIT_BIN=${pkgs.lib.getExe pkgs.git}
          export CHATCMS_HUGO_BIN=${pkgs.lib.getExe pkgs.hugo}
          go vet ./...
          go test -race -covermode=atomic -coverprofile=coverage.out ./...
          runHook postCheck
        '';
        installPhase = ''
          runHook preInstall
          mkdir -p $out
          cp coverage.out $out/coverage.out
          runHook postInstall
        '';
      });

      mkIntegrationTests = { pkgs, ... }@args: (mkGoTests args).overrideAttrs (old: {
        pname = "chatcmsd-integration-tests";
        nativeBuildInputs = (old.nativeBuildInputs or [ ]) ++ [ pkgs.postgresql_16 ];
        checkPhase = ''
          runHook preCheck
          export HOME=$TMPDIR
          export CHATCMS_GIT_BIN=${pkgs.lib.getExe pkgs.git}
          export CHATCMS_HUGO_BIN=${pkgs.lib.getExe pkgs.hugo}
          go vet -tags=integration ./...
          go test -tags=integration -covermode=atomic -coverprofile=coverage.out ./...
          ${pkgs.bash}/bin/bash ${./scripts/coverage-gate.sh} coverage.out
          runHook postCheck
        '';
      });

      mkSqlcCheck = { pkgs, ... }: pkgs.runCommand "chatcms-sqlc-diff"
        {
          nativeBuildInputs = [ pkgs.sqlc ];
          src = nixpkgs.lib.fileset.toSource {
            root = ./.;
            fileset = nixpkgs.lib.fileset.unions [
              ./sqlc.yaml
              ./db
              ./internal/store
            ];
          };
        } ''
        cp -r $src/. .
        chmod -R u+w .
        sqlc diff
        touch $out
      '';

      mkDevScriptsCheck = { pkgs, ... }: pkgs.runCommand "chatcms-dev-scripts"
        {
          nativeBuildInputs = [ pkgs.bash pkgs.shellcheck ];
          src = nixpkgs.lib.fileset.toSource {
            root = ./.;
            fileset = nixpkgs.lib.fileset.unions [
              ./scripts/dev-instance.sh
              ./scripts/dev-be.sh
              ./scripts/dev-fe.sh
              ./scripts/dev-taps.sh
              ./scripts/dev-scripts.test.sh
              ./scripts/gate.sh
              ./scripts/gate.test.sh
            ];
          };
        } ''
        cd $src
        shellcheck -x scripts/dev-instance.sh scripts/dev-be.sh scripts/dev-fe.sh scripts/dev-taps.sh scripts/dev-scripts.test.sh scripts/gate.sh scripts/gate.test.sh
        bash scripts/dev-scripts.test.sh
        bash scripts/gate.test.sh
        touch $out
      '';

      mkReleaseCheck = { pkgs, ... }: pkgs.runCommand "chatcms-release-script"
        {
          nativeBuildInputs = [ pkgs.bash pkgs.shellcheck pkgs.git ];
          src = nixpkgs.lib.fileset.toSource {
            root = ./.;
            fileset = nixpkgs.lib.fileset.unions [
              ./scripts/release.sh
              ./scripts/release.test.sh
            ];
          };
        } ''
        cd $src
        shellcheck scripts/release.sh scripts/release.test.sh
        bash scripts/release.test.sh
        touch $out
      '';

      mkChangelogCheck = { pkgs, ... }: pkgs.runCommand "chatcms-changelog"
        {
          nativeBuildInputs = [ pkgs.bash pkgs.gawk ];
          src = nixpkgs.lib.fileset.toSource {
            root = ./.;
            fileset = nixpkgs.lib.fileset.unions [
              ./CHANGELOG.md
              ./scripts/check-changelog.sh
            ];
          };
        } ''
        bash $src/scripts/check-changelog.sh $src/CHANGELOG.md
        touch $out
      '';

      mkDevShell = { pkgs, system, ... }: pkgs.mkShell {
        name = "chatcms";
        packages = with pkgs; [
          go
          gopls
          gotools
          go-tools
          golangci-lint
          delve
          air

          nodejs_22
          typescript-language-server

          postgresql_16
          sqlc
          goose

          hugo
          git
          jujutsu
          openssh
          bubblewrap
          util-linux

          # Mailpit rather than MailHog, which has been unmaintained since
          # 2020: the same idea, still receiving fixes.
          mailpit

          playwright-driver.browsers
          prefetch-npm-deps
          jq
          yq-go
        ];
        env = {
          PLAYWRIGHT_BROWSERS_PATH = "${pkgs.playwright-driver.browsers}";
          PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD = "1";
          CHATCMS_HUGO_BIN = pkgs.lib.getExe pkgs.hugo;
          CHATCMS_GIT_BIN = pkgs.lib.getExe pkgs.git;
          CHATCMS_BWRAP_BIN = "${pkgs.bubblewrap}/bin/bwrap";
          CHATCMS_PRLIMIT_BIN = "${pkgs.util-linux}/bin/prlimit";
          CHATCMS_WORKFLOW_TOOLS = workflowTools pkgs;
          CHATCMS_OPENCODE_BIN = "${opencodeFor system}/bin/opencode";
        };
        shellHook = ''
          echo "Quiqr ChatCMS dev shell (${system})"
          echo "  scripts/dev-db.sh up     start a local Postgres in ./.pgdata"
          echo "  scripts/gate.sh fast     build, unit tests, frontend: while working"
          echo "  scripts/gate.sh ship     fast plus integration and e2e: what shipping runs"
          echo "  scripts/gate.sh release  nix flake check, every check: what a release runs"
          echo "  bash scripts/ship-change.sh <change>  ship one OpenSpec change"
        '';
      };
      opencodeFor = system: nixpkgs-opencode.legacyPackages.${system}.opencode;

      mkWorkspaceGuest = { pkgs, system, ... }:
        let
          guest = nixpkgs.lib.nixosSystem {
            inherit system;
            specialArgs = { opencode = opencodeFor system; };
            modules = [ microvm.nixosModules.microvm ./nix/workspace-guest.nix ];
          };
          cfg = guest.config.microvm;
        in
        pkgs.writeText "chatcms-workspace-guest.json" (builtins.toJSON {
          kernel = "${cfg.kernel.dev}/vmlinux";
          initrd = cfg.initrdPath;
          cmdline = "console=ttyS0 reboot=t panic=-1 ${toString cfg.kernelParams}";
          memory = cfg.mem;
          vcpu = cfg.vcpu;
        });

    in
    {
      packages = forAllSystems (args: rec {
        chatcmsd = mkServer (args // { source = self.shortRev or self.dirtyShortRev or ""; });
        web = mkWeb args;
        default = chatcmsd;
      } // nixpkgs.lib.optionalAttrs (args.system == "x86_64-linux" || args.system == "aarch64-linux") {
        workspace-guest = mkWorkspaceGuest args;
        opencode = opencodeFor args.system;
      });

      devShells = forAllSystems (args: {
        default = mkDevShell args;
      });

      checks = forAllSystems ({ pkgs, system, ... }@args: {
        go-build = mkServer args;
        go-test = mkGoTests args;
        integration = mkIntegrationTests args;
        frontend = mkWeb (args // { checked = true; });
        sqlc = mkSqlcCheck args;
        dev-scripts = mkDevScriptsCheck args;
        release = mkReleaseCheck args;
        changelog = mkChangelogCheck args;
      } // nixpkgs.lib.optionalAttrs pkgs.stdenv.hostPlatform.isLinux {
        # A virtual machine can only be booted on Linux, so this check exists
        # there and not on a Mac, where the rest still runs.
        nixos-vm = import ./nix/vm-test.nix { inherit self pkgs; };
        e2e = import ./nix/e2e.nix { inherit self pkgs system; };
        workspace-guest = mkWorkspaceGuest args;
      });

      formatter = forAllSystems ({ pkgs, ... }: pkgs.nixpkgs-fmt);

      nixosModules.default = import ./nix/nixos-module.nix { inherit self; };
      nixosModules.chatcms = self.nixosModules.default;
    };
}
