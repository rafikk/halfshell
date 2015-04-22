with import <nixpkgs> {};

with goPackages; rec {
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
    name = "go-s3";
    goPackagePath = "github.com/rafikk/imagick";
    buildInputs = [ pkgconfig imagemagick ];
    src = fetchFromGitHub {
      rev = "master";
      owner = "rafikk";
      repo = "imagick";
      sha256 = "1paarlszxn63cwawgb5m0b1p8k35n6r34raps3383w5wnrqf6w2a";
    };
  };

  halfshell = buildGoPackage {
    goPackagePath = "github.com/oysterbooks/halfshell";
    name = "halfshell";
    src = ./.;
    buildInputs = [ go-s3 go-imagick ];
  };
}
