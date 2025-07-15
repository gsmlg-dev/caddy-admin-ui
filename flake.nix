{
  nixConfig = {
    extra-substituters = "https://mirrors.tuna.tsinghua.edu.cn/nix-channels/store";
  };

  inputs = {
    nixpkgs-next.url = "github:nixos/nixpkgs/release-24.11";
    nixpkgs.url = "github:nixos/nixpkgs/release-25.05";
    nixpkgs-unstable.url = "github:nixos/nixpkgs/nixos-unstable";
    systems.url = "github:nix-systems/default";
  };

  outputs = {
    systems,
    nixpkgs-next,
    nixpkgs,
    nixpkgs-unstable,
    ...
  } @ inputs: let
    # the project root in nix store
    PROJECT_ROOT = builtins.toString ./.;

    eachSystem = f:
      nixpkgs.lib.genAttrs (import systems) (
        system:
          f {
            pkgs-next = nixpkgs-next.legacyPackages.${system};
            pkgs = nixpkgs.legacyPackages.${system};
            pkgs-unstable = nixpkgs-unstable.legacyPackages.${system};
          }
      );
  in {
    formatter = eachSystem (opts: let
      pkgs = opts.pkgs;
    in {
      default = pkgs.alejandra;
    });

    devShells = eachSystem (opts: let
      next = opts.pkgs-next;
      pkgs = opts.pkgs;
      unstable = opts.pkgs-unstable;
    in {
      default = pkgs.mkShell {
        name = "Galaxy Dev Shell";

        buildInputs = [
          pkgs.caddy
          pkgs.xcaddy
          pkgs.alejandra
          pkgs.figlet
          pkgs.lolcat
          pkgs.go
          pkgs.npm-check-updates
          pkgs.nodePackages.nodejs
          pkgs.nodePackages.pnpm
        ];

        shellHook = ''
          figlet -w 120 -f starwars Caddy | lolcat
          figlet -w 120 -f starwars "Admin UI" | lolcat

          source .envrc

        '';
      };
    });
  };
}
