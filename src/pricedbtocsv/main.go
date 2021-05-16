package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/glennhartmann/ledger-tools/src/homedir"
	"github.com/glennhartmann/ledger-tools/src/pricedb"

	"github.com/glennhartmann/ledger-tools/src/pricedbtocsv/lib"
)

var (
	closeTime   = flag.String("close-time", pricedb.DefaultCloseTime, "The time to use for close prices.")
	priceDBFile = flag.String("price-db-file", pricedb.DefaultFile, "price.db file location.")
)

func main() {
	flag.Parse()
	if err := homedir.FillInHomeDir(priceDBFile); err != nil {
		fmt.Fprintf(os.Stderr, "homedir.FillInHomeDir(): %+v\n", err)
		os.Exit(1)
	}

	if err := lib.ToCSV(*priceDBFile, *closeTime); err != nil {
		fmt.Fprintf(os.Stderr, "error: %+v\n", err)
		os.Exit(1)
	}
}
