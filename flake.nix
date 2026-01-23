{
  description = "diffstory - TUI for reviewing AI-generated code changes";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "diffstory";
          version = "0.1.0";
          src = ./.;
          vendorHash = "sha256-wsIOW+q5/wNCjIDhOMDba6hGzts0t6DyDoEgpfvTKUA=";
          # Tests have environment-specific dependencies (hardcoded /tmp paths)
          # that don't work in the Nix sandbox. Run tests locally instead.
          doCheck = false;
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            golangci-lint
            just
            vhs      # For generating demo GIFs
            ffmpeg   # Required by vhs
            ttyd     # Required by vhs
          ];

          shellHook = ''
            echo "diffstory Development Environment"
            echo "Go version: $(go version)"
          '';
        };
      }
    );
}
