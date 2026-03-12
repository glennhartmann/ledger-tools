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
          vendorHash = "sha256-HUgD+Sv3gIU6kHJyZiUVWYQfAlRfqG8yRxe1d8dmRi8=";
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
