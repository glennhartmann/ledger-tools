# ledger-tools

This is a small collection of utility scripts to help managing [ledger-cli](https://www.ledger-cli.org/) data.

## Building

`transactionsorter`, `pricedbfetcher`, and `questrademain` are written in [Go](https://golang.org/). Download a copy of the Go compiler, and run `./build.sh`.

## transactionsorter

Usage: `./transactionsorter <file>`.

This sorts a file full of Ledger transactions by date, in-place. A compelling use-case is for importing multiple CSV files (using [icsv2ledger](https://github.com/quentinsf/icsv2ledger), for example) into the same transactions file.

Although ledger does somewhat support having per-account transaction files, which would somewhat lessen the value of this use-case, but this is [widely acknowledged](https://ledger-cli.narkive.com/nMgbSE28/balance-assertions-should-not-be-based-on-position-in-file) [to break](https://github.com/ledger/ledger/issues/554) [balance assertions](https://github.com/ledger/ledger/issues/2015).

Note that this tool is currently quite limited - it only understands basic transaction syntax and comments, and could fail if other ledger directives are in the file.

## pricedbfetcher

See [pricedbfetcher README](src/pricedbfetcher/README.md).

## questrademain

Mostly just for testing the Questrade API.

## networthbyday

[networthbyday.py](https://github.com/glennhartmann/ledger-tools/blob/master/misc/networthbyday.py) computes a one-row-per-day CSV file of total Assets minus total Liabilities.

Usage: `misc/networthbyday.py --start_date=YYYY-MM-DD --end_date=YYYY-MM-DD`. The output is printed to stdout, so you may want to redirect it to a file.

## regcsv

[This one](https://github.com/glennhartmann/ledger-tools/blob/master/misc/regcsv.sh) probably should just be an alias.
