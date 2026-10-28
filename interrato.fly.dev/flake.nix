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
            caddy
            docker-language-server
            go_1_27
            gopls
            shellcheck
            shfmt
            superhtml
            watchexec
          ];
        };
      });
    };
}
