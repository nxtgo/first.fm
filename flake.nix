{
  description = "minimal flake for go dev";

  inputs.nixpkgs.url = "https://flakehub.com/f/NixOS/nixpkgs/0.1";

  outputs =
    inputs:
    let
      goVersion = 25;

      supportedSystems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forEachSupportedSystem =
        f:
        inputs.nixpkgs.lib.genAttrs supportedSystems (
          system:
          f {
            pkgs = import inputs.nixpkgs {
              inherit system;
              overlays = [ inputs.self.overlays.default ];
            };
          }
        );
    in
    {
      overlays.default = final: prev: {
        go = final."go_1_${toString goVersion}";
        golangci-lint = prev.golangci-lint.overrideAttrs (old: rec {
          version = "2.4.0";

          src = prev.fetchFromGitHub {
            owner = "golangci";
            repo = "golangci-lint";
            rev = "v${version}";
            hash = "sha256-L0TsVOUSU+nfxXyWsFLe+eU4ZxWbW3bHByQVatsTpXA=";
          };

          vendorHash = "sha256-tYoAUumnHgA8Al3jKjS8P/ZkUlfbmmmBcJYUR7+5u9w=";

          ldflags = [
            "-s"
            "-X main.version=${version}"
            "-X main.commit=v${version}"
            "-X main.date=19700101-00:00:00"
          ];

          meta = old.meta // {
            changelog = "https://github.com/golangci/golangci-lint/blob/v${version}/CHANGELOG.md";
          };
        });
      };

      devShells = forEachSupportedSystem (
        { pkgs }:
        {
          default = pkgs.mkShell {
            packages = with pkgs; [
              go
              gotools
              golangci-lint

              # non-go
              gnumake
              sqlc
            ];
          };
        }
      );
    };
}
