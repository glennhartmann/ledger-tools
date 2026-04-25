package main

import (
	"fmt"
	"os"

	"github.com/glennhartmann/ledger-tools/src/pricedb"

	"github.com/glennhartmann/ledger-tools/src/pricedbtocsv/lib"

	flag "github.com/spf13/pflag"
)

var (
	closeTime   = flag.StringP("close-time", "c", pricedb.DefaultCloseTime, "The time to use for close prices.")
	priceDBFile = flag.StringP("price-db-file", "p", pricedb.DefaultFile, "price.db file location.")
)

func main() {
	flag.Parse()
	if err := lib.ToCSV(*priceDBFile, *closeTime); err != nil {
		fmt.Fprintf(os.Stderr, "error: %+v\n", err)
		os.Exit(1)
	}
}
