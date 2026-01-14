{
    description = "APIX (API eXecuter) is a lightweight CLI tool to manage and interact with multiple API domains.";

    inputs = {
        nixpkgs.url = "github:nixos/nixpkgs/nixos-25.05";
        flake-utils.url = "github:numtide/flake-utils";
    };

    outputs = { self, nixpkgs, flake-utils, ... }:
        flake-utils.lib.eachDefaultSystem (system:
            let
                pkgs = import nixpkgs { inherit system; };
            in {
                packages = {
                    apix = pkgs.buildGoModule {
                        pname = "apix";
                        version = builtins.readFile ./VERSION;
                        src = ./.;
                        vendorHash = "sha256-QFHmy/lYqPzhLxV3Cvi7p4AHtj+aiO0zggHCBNa3A28=";
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
