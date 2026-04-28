{
  inputs = {
    nixpkgs.url = github:NixOS/nixpkgs;
    flake-compat.url = "https://flakehub.com/f/edolstra/flake-compat/1.tar.gz";
    flake-utils.url = "github:numtide/flake-utils";
  };
  outputs = { self, nixpkgs, flake-compat, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
        ledger-tools = pkgs.buildGoModule {
          pname = "ledger-tools";
          version = "v0.5.0";
          src = builtins.path { path = ./.; name = "ledger-tools"; };
          vendorHash = "sha256-HPcteUQizKNgBjxfSlMCVPjGlDOjX1KUhKllF0VQlUY=";
        };

        ledger-tools-shell = pkgs.mkShell {
          inputsFrom = [ ledger-tools ];
          packages = with pkgs; [
            fd
            gotools
            protobuf
            protoc-gen-go
          ];
        };
      in
      {
        packages = {
          inherit ledger-tools;
          default = ledger-tools;
        };
        devShells = {
          inherit ledger-tools-shell;
          default = ledger-tools-shell;
        };
      }
    );
}
