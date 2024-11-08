# SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
{
  buildGoModule,
  buildNpmPackage,
  versionCheckHook,
  lib,
}:
let
  version = "0.8.0";

  frontend = buildNpmPackage {
    pname = "gose-frontend";
    inherit version;
    src = ./frontend;

    npmDepsHash = "sha256-p24s2SgCL8E9vUoZEyWSrd15IdkprneAXS7dwb7UbyA=";

    installPhase = ''
      find ./dist
      mkdir $out
      cp -r dist/* $out/
    '';
  };
in
buildGoModule {
  pname = "gose";
  inherit version;
  src = ./.;

  vendorHash = "sha256-U/umJ6McCuD0HARVMj1JXHOpVxcph16z7Y7i47Nf3cg=";

  CGO_ENABLED = 0;

  postInstall = ''
    mv $out/bin/cmd $out/bin/gose
  '';

  tags = [ "embed" ];

  ldflags = [
    "-s"
    "-w"
    "-X"
    "main.version=${version}"
    "-X"
    "main.builtBy=Nix"
  ];

  checkFlags = "-skip TestShortener";

  nativeInstallCheckInputs = [
    versionCheckHook
  ];
  doInstallCheck = true;

  prePatch = ''
    cp -r ${frontend} frontend/dist
  '';

  meta = {
    description = "GoSƐ: A terascale file-uploader";
    homepage = "https://github.com/stv0g/gose";
    license = lib.licenses.asl20;
    maintainers = with lib.maintainers; [ stv0g ];
  };
}
