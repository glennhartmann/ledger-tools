# ledger-tools

This is a small collection of utility scripts to help managing [ledger-cli](https://www.ledger-cli.org/) data.

## Building

`transactionsorter`, `pricedbfetcher`, and `questrademain` are written in [Go](https://golang.org/). Download a copy of the Go compiler, and run `./build.sh`.

Or, better yet, install [Nix](https://en.wikipedia.org/wiki/Nix_(package_manager)) and build with `nix build`.

## transactionsorter

Usage: `./transactionsorter <file>`.

This sorts a file full of Ledger transactions by date, in-place. A compelling use-case is for importing multiple CSV files (using [icsv2ledger](https://github.com/quentinsf/icsv2ledger), for example) into the same transactions file.

Although ledger does somewhat support having per-account transaction files, which would somewhat lessen the value of this use-case, but this is [widely acknowledged](https://ledger-cli.narkive.com/nMgbSE28/balance-assertions-should-not-be-based-on-position-in-file) [to break](https://github.com/ledger/ledger/issues/554) [balance assertions](https://github.com/ledger/ledger/issues/2015).

Note that this tool is currently quite limited - it only understands basic transaction syntax and comments, and could fail if other ledger directives are in the file.

## pricedbfetcher

See [pricedbfetcher README](src/pricedbfetcher/README.md).

## pricedbtocsv

Usage: `./pricedbtocsv [-close-time=<time in '22:45:00' format>] [-price-db-file=<path>]`

As the name suggests, this tool converts a ledger-cli price-db file (see [here](https://github.com/glennhartmann/ledger-tools/tree/master/src/pricedbfetcher#pricedb) for more details) into CSV data. The CSV data is printed to stdout, so you may want to redirect it to a file.

## questrademain

Mostly just for testing the Questrade API.

## pricedbmain

Usage: `./pricedbmain [--close-time=<time in '22:45:00' format>] [--price-db-path=<path>] [--output-type=<"json"|"proto-text"|"proto-wire">]`

This utility parses the price-db file, converts it into a slice of [TimeSeriesItemWithSymbol](https://github.com/glennhartmann/ledger-tools/blob/4da12d9f8197ae0b0a3ad38c1c418d34b2a3a403/src/priceutils/priceutils.go#L13), and then outputs it in a [protocol buffer](https://en.wikipedia.org/wiki/Protocol_Buffers) [format](https://github.com/glennhartmann/ledger-tools/blob/master/src/priceutils/proto/priceutils.proto) for storage or consumption by other programs.

## networthbyday

[networthbyday.py](https://github.com/glennhartmann/ledger-tools/blob/master/misc/networthbyday.py) computes a one-row-per-day CSV file of total Assets minus total Liabilities.

Usage: `misc/networthbyday.py --start_date=YYYY-MM-DD --end_date=YYYY-MM-DD`. The output is printed to stdout, so you may want to redirect it to a file.

## regcsv

[This one](https://github.com/glennhartmann/ledger-tools/blob/master/misc/regcsv.sh) probably should just be an alias.
