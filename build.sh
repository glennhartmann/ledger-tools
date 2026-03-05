#!/bin/bash
go build -mod=readonly github.com/glennhartmann/ledger-tools/src/transactionsorter
go build -mod=readonly github.com/glennhartmann/ledger-tools/src/pricedbfetcher
go build -mod=readonly github.com/glennhartmann/ledger-tools/src/questrademain
go build -mod=readonly github.com/glennhartmann/ledger-tools/src/pricedbtocsv
