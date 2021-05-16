package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/glennhartmann/ledger-tools/src/alphavantage"
	"github.com/glennhartmann/ledger-tools/src/coinbase"
	"github.com/glennhartmann/ledger-tools/src/homedir"
	"github.com/glennhartmann/ledger-tools/src/pricedb"
	"github.com/glennhartmann/ledger-tools/src/questrade"

	"github.com/glennhartmann/ledger-tools/src/pricedbfetcher/lib"
)

var (
	alphavantageBaseURL         = flag.String("alphavantage-base-url", alphavantage.DefaultBaseURL, "Alpha Vantage base URL (not including query string) to fetch from.")
	configFile                  = flag.String("config-file", "~/.pricedbfetcher_config", "Config file location.")
	alphavantageAPIKeyFile      = flag.String("alphavantage-api-key-file", alphavantage.DefaultAPIKeyFile, "Alpha Vantage API Key file location.")
	priceDBFile                 = flag.String("price-db-file", pricedb.DefaultFile, "price.db file location.")
	outFile                     = flag.String("out-path", pricedb.DefaultFile, "Where to write output. Empty means stdout. It's safe to make this the same as -price-db-file.")
	closeTime                   = flag.String("close-time", pricedb.DefaultCloseTime, "The time to use for close prices.")
	alphavantageBackoffDuration = flag.Duration("alphavantage-backoff-duration", alphavantage.DefaultBackoffDuration, "How long to back off for after hitting the rate limit. Must be parseable by https://golang.org/pkg/time/#ParseDuration.")
	alphavantageBackoffRetry    = flag.Int("alphavantage-backoff-retry", alphavantage.DefaultBackoffRetry, "Number of times to retry after hitting rate limit before giving up.")
	questradeOAuthURLFmt        = flag.String("questrade-oauth-url-fmt", questrade.DefaultOAuthURLFmt, "Format-string for questrade OAuth URL.")
	questradeTokenFile          = flag.String("questrade-token-file", questrade.DefaultTokenFile, "File to find questrade OAuth token.")
	questradeAccountNumbersFile = flag.String("questrade-account-numbers-file", questrade.DefaultAccountNumbersFile, "File to find questrade account numbers.")
	now                         = flag.String("now", "", fmt.Sprintf("Override 'time.Now()' value if not blank. Must be RFC3339 ('%s') format.", time.RFC3339))
	coinbaseBaseURL             = flag.String("coinbase-base-url", coinbase.DefaultBaseURL, "Coinbase base API URL.")
)

func main() {
	flag.Parse()
	if err := homedir.FillInHomeDir(configFile, alphavantageAPIKeyFile, priceDBFile, outFile, questradeTokenFile, questradeAccountNumbersFile); err != nil {
		fmt.Fprintf(os.Stderr, "homedir.FillInHomeDir(): %+v\n", err)
		os.Exit(1)
	}
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
