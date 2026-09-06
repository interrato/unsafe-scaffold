{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      forAllSystems =
        packagesFn:
        nixpkgs.lib.genAttrs nixpkgs.lib.systems.flakeExposed (
          system: packagesFn nixpkgs.legacyPackages.${system}
        );
    in
    {
      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShellNoCC {
          env = {
            CGO_ENABLED = 0;
          };

          packages = with pkgs; [
            bash-language-server
            docker
            docker-language-server
            go
            gopls
            just
            shellcheck
            shfmt
            superhtml
            watchexec
          ];
        };
      });
    };
}
