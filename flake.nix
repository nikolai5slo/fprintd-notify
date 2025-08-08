{
  description = "A daemon to notify on fprintd events";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }:
    let
      supportedSystems = [ "x86_64-linux" "aarch64-linux" ]; # Add other systems if needed

      forAllSystems = f: nixpkgs.lib.genAttrs supportedSystems (system: f system);
    in
    {
      packages = forAllSystems (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          fprintd-notify = pkgs.buildGoModule {
            pname = "fprintd-notify";
            version = "0.1.0"; # You might want to manage this version
            src = ./.;
            vendorHash = "sha256-p6DqhvANhM5SpDEYLwlZzmkgiDnbyLOVs7ozaysEli0=";
            ldflags = [ "-s" "-w" ];
          };
        });

      nixosModules.fprintd-notify = { config, lib, pkgs, ... }: with lib; {
        options.services.fprintd-notify = {
          enable = mkEnableOption "fprintd-notify daemon";
          package = mkOption {
            type = types.package;
            default = self.packages.${pkgs.system}.fprintd-notify;
            description = "The fprintd-notify package to use.";
          };
        };

        config = mkIf config.services.fprintd-notify.enable {
          systemd.user.services.fprintd-notify = {
            description = "fprintd notification daemon";
            wantedBy = [ "graphical-session.target" ];
            after = [ "graphical-session.target" "dbus.service" ]; # Ensure graphical environment and D-Bus are ready
            serviceConfig = {
              ExecStart = "${getExe config.services.fprintd-notify.package}";
              Restart = "on-failure";
            };
          };
        };
      };
    };
}
