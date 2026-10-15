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
                rev = self.shortRev or self.dirtyShortRev or "unknown";
                lastReleased = (builtins.fromJSON (builtins.readFile ./.release-please-manifest.json)).".";
                version = "${lastReleased}+${rev}";
                buildCmd = ''go build -ldflags "-X main.version=${version}" -o apix ./cmd/apix'';
                vetCmd = "go vet ./...";
                testCmd = "go test -v ./...";
            in {
                packages = {
                    apix = pkgs.buildGoModule {
                        pname = "apix";
                        inherit version;
                        src = ./.;
                        vendorHash = "sha256-8UYvuu60pwaZm/O1f8/atlCFsJdExZy/MUmOobgGXe8=";
                        ldflags = [ "-X main.version=${version}" ];
                        doCheck = false;
                    };
                    default = self.packages.${system}.apix;
                };

                checks = {
                    test = self.packages.${system}.apix.overrideAttrs (_: {
                        pname = "apix-checks";
                        doCheck = true;
                        checkPhase = ''
                            runHook preCheck
                            ${vetCmd}
                            ${testCmd}
                            runHook postCheck
                        '';
                    });
                };

                apps = {
                    default = self.apps.${system}.apix;

                    apix = {
                        type = "app";
                        program = "${self.packages.${system}.apix}/bin/apix";
                        meta = {
                          description = "APIX (API eXecuter) is a lightweight CLI tool to manage and interact with multiple API domains.";
                          homepage = "github.com/rhydianjenkins/apix";
                        };
                    };

                    build = {
                        type = "app";
                        program = "${pkgs.writeShellApplication {
                            name = "build";
                            runtimeInputs = [ pkgs.go ];
                            text = buildCmd;
                        }}/bin/build";
                        meta.description = "Build apix into a binary";
                    };
                };

                devShells.default = pkgs.mkShell {
                    buildInputs = with pkgs; [
                        gcc
                        go
                        gopls
                        semgrep
                        semgrep-core
                    ];
                };
            });
}
