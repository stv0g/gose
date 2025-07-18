# SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
{
  description = "GoSƐ: A terascale file-uploader";

  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    nix-update = {
      url = "github:Mic92/nix-update";
      inputs = {
        nixpkgs.follows = "nixpkgs";
      };
    };
  };

  nixConfig = {
    extra-substituters = "https://stv0g.cachix.org";
    extra-trusted-public-keys = "stv0g.cachix.org-1:Bliox3TtWqQhKr2w6HMSbpwn9E9M2vgKmA/N7VpYOmY=";
  };

  outputs =
    {
      self,
      flake-utils,
      nixpkgs,
      nix-update,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };
      in
      {
        devShells = with pkgs; {
          default = mkShell {
            inputsFrom = [
              self.packages.${system}.default
            ];

            packages = with pkgs; [
              nix-update.packages.${system}.nix-update
              golangci-lint
              reuse
              nodejs_22
              goreleaser
            ];
          };

          ci = mkShell {
            packages = [
              nix-update.packages.${system}.nix-update
              goreleaser
            ];
          };
        };

        packages = rec {
          gose = pkgs.callPackage ./default.nix { };
          default = gose;
        };

        formatter = pkgs.nixfmt-rfc-style;
      }
    );
}
