# SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
{
  description = "GoSƐ: A terascale file-uploader";

  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
  };

  outputs =
    {
      self,
      flake-utils,
      nixpkgs,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };
      in
      {
        devShell = pkgs.mkShell {
          inputsFrom = [
            self.packages.${system}.default
          ];

          packages = with pkgs; [
            golangci-lint
            reuse
            nodejs_22
            goreleaser
          ];
        };

        packages = rec {
          gose = pkgs.callPackage ./default.nix { };
          default = gose;
        };

        formatter = pkgs.nixfmt-rfc-style;
      }
    );
}
