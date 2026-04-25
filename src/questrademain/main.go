package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/glennhartmann/ledger-tools/src/questrade"

	flag "github.com/spf13/pflag"
)

var (
	tokenFile          = flag.StringP("token-file", "t", questrade.DefaultTokenFile, "File where questrade token is stored.")
	accountNumbersFile = flag.StringP("account-numbers-file", "a", questrade.DefaultAccountNumbersFile, "File where questrade account numbers (comma-separated) are stored.")
	syms               = flag.StringP("symbols", "s", "BND,ZAG.TO", "Symbols (comma-separated) to search for.")
	oauthURLFmt        = flag.StringP("oauth-url-fmt", "o", questrade.DefaultOAuthURLFmt, "Format-string for questrade oauth API URL.")
)

func main() {
	flag.Parse()
	b, err := ioutil.ReadFile(*tokenFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ioutil.ReadFile(%s): %+v\n", *tokenFile, err)
		os.Exit(1)
	}
	token := strings.TrimSpace(string(b))

	b, err = ioutil.ReadFile(*accountNumbersFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ioutil.ReadFile(%s): %+v\n", *accountNumbersFile, err)
		os.Exit(1)
	}
	accountNumbers := strings.Split(strings.TrimSpace(string(b)), ",")

	oauthResponse, err := questrade.Authenticate(token, *tokenFile, *oauthURLFmt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Authenticate(%s): %+v\n", token, err)
		os.Exit(1)
	}

	symbols := strings.Split(*syms, ",")
	for _, symbolToSearch := range symbols {
		symbolToSearch = strings.TrimSpace(symbolToSearch)
		symbolResponse, err := questrade.FetchRawSymbol(oauthResponse, symbolToSearch, time.Time{} /* TODO */, time.Now())
		if err != nil {
			fmt.Fprintf(os.Stderr, "FetchRawSymbol(%s): %+v\n", symbolToSearch, err)
			os.Exit(1)
		}

		var out bytes.Buffer
		json.Indent(&out, []byte(symbolResponse), "", "    ")
		fmt.Printf("%s\n", out.String())
	}

	for _, accountNumber := range accountNumbers {
		positionsResponse, err := questrade.FetchRawPositions(oauthResponse, accountNumber)
		if err != nil {
			fmt.Fprintf(os.Stderr, "FetchRawPositions(%s): %+v\n", accountNumber, err)
			os.Exit(1)
		}

		var out bytes.Buffer
		json.Indent(&out, []byte(positionsResponse), "", "    ")
		fmt.Printf("%s\n", out.String())
	}
}
