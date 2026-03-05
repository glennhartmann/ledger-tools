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
          version = "v0.2.0";
          src = builtins.path { path = ./.; name = "ledger-tools"; };
          vendorHash = "sha256-5nx3O7h4N0AtAfZYR5v0qtS05Eect3MSCpTinbcHyYU=";
        };
      in
      {
        packages = {
          inherit ledger-tools;
          default = ledger-tools;
        };
      }
    );
}
