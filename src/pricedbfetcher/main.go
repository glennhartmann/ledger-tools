package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/glennhartmann/ledger-tools/src/alphavantage"
	"github.com/glennhartmann/ledger-tools/src/coinbase"
	"github.com/glennhartmann/ledger-tools/src/pricedb"
	"github.com/glennhartmann/ledger-tools/src/questrade"

	"github.com/glennhartmann/ledger-tools/src/pricedbfetcher/lib"

	flag "github.com/spf13/pflag"
)

var (
	alphavantageBaseURL         = flag.String("alphavantage-base-url", alphavantage.DefaultBaseURL, "Alpha Vantage base URL (not including query string) to fetch from.")
	configFile                  = flag.StringP("config-file", "c", lib.DefaultConfigFile, "Config file location.")
	alphavantageAPIKeyFile      = flag.StringP("alphavantage-api-key-file", "a", alphavantage.DefaultAPIKeyFile, "Alpha Vantage API Key file location.")
	priceDBFile                 = flag.StringP("price-db-file", "p", pricedb.DefaultFile, "price.db file location.")
	outFile                     = flag.StringP("out-path", "o", pricedb.DefaultFile, "Where to write output. Empty means stdout. It's safe to make this the same as -price-db-file.")
	closeTime                   = flag.StringP("close-time", "e", pricedb.DefaultCloseTime, "The time to use for close prices.")
	alphavantageBackoffDuration = flag.DurationP("alphavantage-backoff-duration", "b", alphavantage.DefaultBackoffDuration, "How long to back off for after hitting the rate limit. Must be parseable by https://golang.org/pkg/time/#ParseDuration.")
	alphavantageBackoffRetry    = flag.IntP("alphavantage-backoff-retry", "r", alphavantage.DefaultBackoffRetry, "Number of times to retry after hitting rate limit before giving up.")
	questradeOAuthURLFmt        = flag.String("questrade-oauth-url-fmt", questrade.DefaultOAuthURLFmt, "Format-string for questrade OAuth URL.")
	questradeTokenFile          = flag.StringP("questrade-token-file", "t", questrade.DefaultTokenFile, "File to find questrade OAuth token.")
	questradeAccountNumbersFile = flag.StringP("questrade-account-numbers-file", "q", questrade.DefaultAccountNumbersFile, "File to find questrade account numbers.")
	now                         = flag.StringP("now", "n", "", fmt.Sprintf("Override 'time.Now()' value if not blank. Must be RFC3339 ('%s') format.", time.RFC3339))
	coinbaseBaseURL             = flag.String("coinbase-base-url", coinbase.DefaultBaseURL, "Coinbase base API URL.")
)

func main() {
	flag.Parse()
	c := &lib.Conn{
		AlphavantageBaseURL:         *alphavantageBaseURL,
		ConfigFile:                  *configFile,
		AlphavantageAPIKeyFile:      *alphavantageAPIKeyFile,
		PriceDBFile:                 *priceDBFile,
		OutFile:                     *outFile,
		CloseTime:                   *closeTime,
		AlphavantageBackoffDuration: *alphavantageBackoffDuration,
		AlphavantageBackoffRetry:    *alphavantageBackoffRetry,
		QuestradeOAuthURLFmt:        *questradeOAuthURLFmt,
		QuestradeTokenFile:          *questradeTokenFile,
		QuestradeAccountNumbersFile: *questradeAccountNumbersFile,
		Now:                         setupNow(strings.TrimSpace(*now)),
		CoinbaseBaseURL:             *coinbaseBaseURL,
	}
	if err := c.Fetch(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %+v\n", err)
		os.Exit(1)
	}
}

func setupNow(n string) time.Time {
	if n == "" {
		return time.Now()
	}
	d, err := time.Parse(time.RFC3339, n)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't parse -now=%s (%v)\n", n, err)
		os.Exit(1)
	}
	return d
}
