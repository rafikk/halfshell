{ nixpkgs ? <nixpkgs>
, name ? "halfshell"
, src ? ./. }:

with import nixpkgs {};

with goPackages; let

  buildSrc = src;

  go-s3 = buildGoPackage rec {
    name = "go-s3";
    goPackagePath = "github.com/oysterbooks/s3";
    src = fetchFromGitHub {
      rev = "master";
      owner = "oysterbooks";
      repo = "s3";
      sha256 = "0ql1i7b8qjrvh6bbh43vka9va7q15s98s1x2h7b1c5q3nsgn77sy";
    };
  };

  go-imagick = buildGoPackage rec {
    name = "go-imagick";
    goPackagePath = "github.com/rafikk/imagick";
    propagatedBuildInputs = [ pkgconfig imagemagick ];
    src = fetchFromGitHub {
      rev = "master";
      owner = "rafikk";
      repo = "imagick";
      sha256 = "1paarlszxn63cwawgb5m0b1p8k35n6r34raps3383w5wnrqf6w2a";
    };
  };

  go-halfshell = buildGoPackage rec {
    name = "go-halfshell";
    goPackagePath = "github.com/oysterbooks/halfshell/halfshell";
    propagatedBuildInputs = [ go-s3 go-imagick ];
    src = builtins.toPath "${buildSrc}/halfshell";
  };

in buildGoPackage {
  goPackagePath = "github.com/oysterbooks/halfshell";
  name = name;
  src = buildSrc;
  propagatedBuildInputs = [ go-halfshell ];
} // {
  name = name;
  meta = {
    homepage = "https://github.com/oysterbooks/halfshell";
    maintainers = [
      "Rafik Salama <rafik@oysterbooks.com>"
    ];
  };
}
