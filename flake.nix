{
    description = "APIX (API eXecuter) is a lightweight CLI tool to manage and interact with multiple API domains.";

    inputs = {
        nixpkgs.url = "github:nixos/nixpkgs/nixos-26.05";
        flake-utils.url = "github:numtide/flake-utils";
    };

    outputs = { self, nixpkgs, flake-utils, ... }:
        flake-utils.lib.eachDefaultSystem (system:
            let
                pkgs = import nixpkgs { inherit system; };
                version = pkgs.lib.strings.trim (builtins.readFile ./VERSION);
            in {
                packages = {
                    apix = pkgs.buildGoModule {
                        pname = "apix";
                        inherit version;
                        src = ./.;
                        vendorHash = "sha256-QFHmy/lYqPzhLxV3Cvi7p4AHtj+aiO0zggHCBNa3A28=";
                        ldflags = [ "-X main.version=${version}" ];
                    };
                    default = self.packages.${system}.apix;
                };

                apps = {
                    apix = {
                        type = "app";
                        program = "${self.packages.${system}.apix}/bin/apix";
                        meta = {
                          description = "APIX (API eXecuter) is a lightweight CLI tool to manage and interact with multiple API domains.";
                          homepage = "github.com/rhydianjenkins/apix";
                        };
                    };
                    default = self.apps.${system}.apix;
                };

                devShells.default = pkgs.mkShell {
                    buildInputs = [
                        pkgs.go
                        pkgs.gopls
                    ];
                };
            });
}
