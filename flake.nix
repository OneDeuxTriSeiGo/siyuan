{
  description = "Siyuan: A privacy-first, self-hosted, FOSS PKMS.";

  inputs = {
    systems.url = "github:nix-systems/default";
    nixpkgs.url = "github:NixOS/nixpkgs";
    flake-parts.url = "github:hercules-ci/flake-parts";

    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    devshell = {
      url = "github:numtide/devshell";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = inputs @ {flake-parts, ...}:
    flake-parts.lib.mkFlake {inherit inputs;} {
      imports = [
        inputs.treefmt-nix.flakeModule
        #inputs.devshell.flakeModule
      ];
      systems = import inputs.systems;
      perSystem = {
        config,
        self',
        inputs',
        pkgs,
        system,
        ...
      }: let
        siyuan-drv = pkgs.callPackage ./package.nix {};
      in {
        packages.default = siyuan-drv;

        treefmt = {
          programs.alejandra.enable = true;
        };

        apps.default = {
          type = "app";
          program = siyuan-drv + "/bin/siyuan";
        };
      };
    };
}
