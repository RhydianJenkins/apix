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
                buildCmd = ''go build -ldflags "-X main.version=$(cat VERSION)" -o apix ./cmd/apix'';
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
                    };
                    default = self.packages.${system}.apix;
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

                    test = {
                        type = "app";
                        program = "${pkgs.writeShellApplication {
                            name = "test";
                            runtimeInputs = [ pkgs.go ];
                            text = testCmd;
                        }}/bin/test";
                        meta.description = "Run all tests";
                    };

                    vet = {
                        type = "app";
                        program = "${pkgs.writeShellApplication {
                            name = "vet";
                            runtimeInputs = [ pkgs.go ];
                            text = vetCmd;
                        }}/bin/vet";
                        meta.description = "Run go vet";
                    };

                    release = {
                        type = "app";
                        program = "${pkgs.writeShellApplication {
                            name = "release";
                            runtimeInputs = [ pkgs.go pkgs.git ];
                            text = ''
                                ${buildCmd}
                                ${vetCmd}
                                ${testCmd}

                                version="v$(cat VERSION)"

                                if git rev-parse "$version" >/dev/null 2>&1; then
                                    echo "Tag $version already exists"
                                    exit 1
                                fi

                                git tag -a "$version" -m "Release $version"
                                git push origin "$version"
                                echo "Released $version"
                            '';
                        }}/bin/release";
                        meta.description = "Build, vet, and test, then tag and push a release using the version in VERSION";
                    };
                };

                devShells.default = pkgs.mkShell {
                    buildInputs = [
                        pkgs.go
                        pkgs.gopls
                    ];
                };
            });
}
