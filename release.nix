{ nixpkgs ? <nixpkgs>
, src ? { outPath = ./.; gitTag = "dirty"; }
, system ? builtins.currentSystem }:

let
  pkgs = import nixpkgs { inherit system; };

in with pkgs; rec {
  tarball = releaseTools.sourceTarball {
    src = src;
    name = "halfshell";
    version = src.gitTag;
    versionSuffix = "";
    doBuild = true;
    distPhase = ''
      mkdir -p $out/tarballs/
      tar czf $out/tarballs/halfshell-${src.gitTag}.tar.gz .
    '';
  };

  build = import ./default.nix {
    inherit nixpkgs;
    name = "halfshell-${src.gitTag}";
    src = src;
  };
}
