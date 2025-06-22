{
    description = "APIX (API eXecuter) is a lightweight CLI tool to manage and interact with multiple API domains.";

    inputs = {
        nixpkgs.url = "github:nixos/nixpkgs/nixos-25.05";
        flake-utils.url = "github:numtide/flake-utils";
    };

    outputs = { self, nixpkgs, flake-utils, ... }:
        flake-utils.lib.eachDefaultSystem (system:
            let
                pkgs = import nixpkgs {
                    inherit system;
                };
            in {
                packages.apix = pkgs.buildGoModule {
                    pname = "apix";
                    version = "1.0.0";
                    src = ./.;
                    vendorHash = "sha256-4/w2pzqPgy+vsVaq4gDhRLsVlrm1WAj2LEgNiUcp1vk=";
                };

                defaultPackage = self.packages.${system}.apix;

                devShell = pkgs.mkShell {
                    buildInputs = [
                        pkgs.go
                        pkgs.gopls
                    ];
                };
            });
}
