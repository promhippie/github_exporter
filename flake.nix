{
  description = "Nix flake for development";

  inputs = {
    nixpkgs = {
      url = "github:NixOS/nixpkgs/nixos-unstable";
    };

    devenv = {
      url = "github:cachix/devenv";
    };

    pre-commit-hooks-nix = {
      url = "github:cachix/pre-commit-hooks.nix";
    };

    flake-parts = {
      url = "github:hercules-ci/flake-parts";
    };
  };

  outputs = inputs@{ flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [
        inputs.devenv.flakeModule
        inputs.pre-commit-hooks-nix.flakeModule
      ];

      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];

      perSystem = { config, self', inputs', pkgs, system, ... }: {
        imports = [
          {
            _module.args.pkgs = import inputs.nixpkgs {
              inherit system;
              config.allowUnfree = true;
            };
          }
        ];

        pre-commit = {
          settings = {
            hooks = {
              nixpkgs-fmt = {
                enable = true;
              };
              golangci-lint = {
                enable = true;
              };
            };
          };
        };

        devenv = {
          shells = {
            default = {
              languages = {
                go = {
                  enable = true;
                  package = pkgs.go_1_22;
                };
              };

              packages = with pkgs; [
                bingo
                gnumake
                nixpkgs-fmt
              ];

              env = {
                CGO_ENABLED = "0";
              };
            };
          };
        };
      };
    };
}
