package main

import (
	"fmt"
	"os"

	"github.com/glennhartmann/ledger-tools/src/transactionsorter/lib"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("wrong args\n")
		os.Exit(1)
	}

	if err := lib.SortFile(os.Args[1]); err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}
