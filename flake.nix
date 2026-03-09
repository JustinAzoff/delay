{
  description = "Delay HTTP server for testing";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.11";
  };

  outputs =
    { self, nixpkgs }:
    let
      supportedSystems = [
        "x86_64-linux"
        "aarch64-linux"
      ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
    in
    {
      packages = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.buildGoModule {
            pname = "delay";
            version = "0.1.0";
            src = ./.;
            vendorHash = null;
          };
        }
      );

      checks = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          integration = pkgs.testers.nixosTest {
            name = "delay";

            nodes.machine =
              { ... }:
              {
                imports = [ self.nixosModules.default ];
                services.delay.enable = true;
                services.delay.port = 8080;
              };

            testScript = ''
              machine.wait_for_unit("delay.service")
              machine.wait_for_open_port(8080)
              machine.succeed("curl -s 'http://localhost:8080/?delay=0.1&count=1' | grep -q 'Delayed'")
            '';
          };
        }
      );

      nixosModules.default =
        {
          config,
          lib,
          pkgs,
          ...
        }:
        let
          cfg = config.services.delay;
        in
        {
          options.services.delay = {
            enable = lib.mkEnableOption "delay HTTP server";

            port = lib.mkOption {
              type = lib.types.port;
              default = 8080;
              description = "Port to listen on";
            };
          };

          config = lib.mkIf cfg.enable {
            systemd.services.delay = {
              description = "Delay HTTP server";
              wantedBy = [ "multi-user.target" ];
              after = [ "network.target" ];

              serviceConfig = {
                ExecStart = "${self.packages.${pkgs.stdenv.hostPlatform.system}.default}/bin/delay";
                Restart = "always";
                RestartSec = 5;
                DynamicUser = true;
                Environment = "PORT=${toString cfg.port}";
              };
            };
          };
        };
    };
}
