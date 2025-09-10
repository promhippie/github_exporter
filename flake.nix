{
  description = "Nix flake for development";

  inputs = {
    nixpkgs = {
      url = "github:nixos/nixpkgs/nixpkgs-unstable";
    };

    devenv = {
      url = "github:cachix/devenv";
    };

    flake-parts = {
      url = "github:hercules-ci/flake-parts";
    };

    git-hooks = {
      url = "github:cachix/git-hooks.nix";
    };
  };

  outputs =
    inputs@{ flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [
        inputs.devenv.flakeModule
        inputs.git-hooks.flakeModule
      ];

      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];

      perSystem =
        {
          config,
          self',
          inputs',
          pkgs,
          system,
          ...
        }:
        {
          imports = [
            {
              _module.args.pkgs = import inputs.nixpkgs {
                inherit system;
                config.allowUnfree = true;
              };
            }
          ];

          devenv = {
            shells = {
              default = {
                git-hooks = {
                  hooks = {
                    nixfmt-rfc-style = {
                      enable = true;
                    };
                    gofmt = {
                      enable = true;
                    };

                    golangci-lint = {
                      enable = true;
                      entry = "go tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint run ./...";
                      pass_filenames = false;
                    };
                  };
                };

                languages = {
                  go = {
                    enable = true;
                    package = pkgs.go_1_25;
                  };
                };

                packages = with pkgs; [
                  go-task
                  goreleaser
                  hugo
                  nixfmt-rfc-style
                ];

                env = {
                  CGO_ENABLED = "0";
                };

                processes = {
                  current-server = {
                    exec = "task watch";

                    process-compose = {
                      readiness_probe = {
                        exec.command = "${pkgs.curl}/bin/curl -sSf http://localhost:9504/readyz";
                        initial_delay_seconds = 2;
                        period_seconds = 10;
                        timeout_seconds = 4;
                        success_threshold = 1;
                        failure_threshold = 5;
                      };

                      availability = {
                        restart = "on_failure";
                      };
                    };
                  };
                };
              };
            };
          };
        };
    };
}
