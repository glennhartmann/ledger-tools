#!/bin/bash
go test -mod=readonly github.com/glennhartmann/ledger-tools/src/transactionsorter/lib
go test -mod=readonly github.com/glennhartmann/ledger-tools/src/pricedbtocsv/lib
go test -mod=readonly github.com/glennhartmann/ledger-tools/src/pricedb
