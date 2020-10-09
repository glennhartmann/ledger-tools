package lib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/pkg/errors"

	"github.com/glennhartmann/ledger-tools/src/alphavantage"
	"github.com/glennhartmann/ledger-tools/src/coinbase"
	"github.com/glennhartmann/ledger-tools/src/pricedb"
	"github.com/glennhartmann/ledger-tools/src/priceutils"
	"github.com/glennhartmann/ledger-tools/src/questrade"
)

const (
	DefaultCloseTime = "22:45:00"
)

type Conn struct {
	AlphavantageBaseURL         string
	ConfigFile                  string
	AlphavantageAPIKeyFile      string
	PriceDBFile                 string
	OutFile                     string
	CloseTime                   string
	AlphavantageBackoffDuration time.Duration
	AlphavantageBackoffRetry    int
	QuestradeOAuthURLFmt        string
	QuestradeTokenFile          string
	QuestradeAccountNumbersFile string
	Now                         time.Time
	CoinbaseBaseURL             string
}

func (c *Conn) Fetch() error {
	rc := &ResolvedConn{
		AlphavantageBaseURL:         c.AlphavantageBaseURL,
		AlphavantageBackoffDuration: c.AlphavantageBackoffDuration,
		AlphavantageBackoffRetry:    c.AlphavantageBackoffRetry,
		CloseTime:                   c.CloseTime,
		QuestradeOAuthURLFmt:        c.QuestradeOAuthURLFmt,
		QuestradeTokenFile:          c.QuestradeTokenFile,
		Now:                         c.Now,
		CoinbaseBaseURL:             c.CoinbaseBaseURL,
	}

	configBytes, err := ioutil.ReadFile(c.ConfigFile)
	if err != nil {
		return errors.Wrapf(err, "ioutil.ReadFile(%s)", c.ConfigFile)
	}
	rc.Conf = &Config{}
	if err := json.Unmarshal(configBytes, rc.Conf); err != nil {
		return errors.Wrap(err, "json.Unmarshal(config)")
	}
	if rc.Conf.StartDate != "" {
		rc.StartDate, err = time.Parse("2006-01-02", rc.Conf.StartDate)
		if err != nil {
			return errors.Wrapf(err, "time.Parse(%s)", rc.Conf.StartDate)
		}
	}

	alphavantageAPIKeyBytes, err := ioutil.ReadFile(c.AlphavantageAPIKeyFile)
	if err != nil {
		return errors.Wrapf(err, "ioutil.ReadFile(%s)", c.AlphavantageAPIKeyFile)
	}
	rc.AlphavantageAPIKey = strings.TrimSpace(string(alphavantageAPIKeyBytes))

	priceDBData, err := readPriceDB(c.PriceDBFile)
	if err != nil {
		return errors.Wrapf(err, "readPriceDB(%s)", c.PriceDBFile)
	}
	rc.PriceDBData = priceDBData

	outFileOpen := func() (*os.File, error) { return os.Stdout, nil }
	outFileClose := func(f *os.File) {}
	if c.OutFile != "" {
		outFileOpen = func() (*os.File, error) { return os.OpenFile(c.OutFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0640) }
		outFileClose = func(f *os.File) { f.Close() }
	}
	rc.OutFileOpen = outFileOpen
	rc.OutFileClose = outFileClose

	rc.QuestradeAccountNumbers, err = readQuestradeAccountNumbers(c.QuestradeAccountNumbersFile)
	if err != nil {
		return errors.Wrapf(err, "readQuestradeAccountNumbers(%s)", c.QuestradeAccountNumbersFile)
	}

	return rc.Fetch()
}

type Config struct {
	StartDate          string                      `json:"start_date"`
	AlphavantageConfig *alphavantage.Config        `json:"alphavantage"`
	QuestradeConfig    *questrade.Config           `json:"questrade"`
	CoinbaseConfig     *coinbase.Config            `json:"coinbase"`
	Commodity          map[string]*CommodityConfig `json:"commodity"`
}

type CommodityConfig struct {
	Currency string `json:"currency"`
	Display  string `json:"display"`
}

type ResolvedConn struct {
	Conf                        *Config
	StartDate                   time.Time
	AlphavantageBaseURL         string
	AlphavantageAPIKey          string
	AlphavantageBackoffDuration time.Duration
	AlphavantageBackoffRetry    int
	CloseTime                   string
	PriceDBData                 []string
	OutFileOpen                 func() (*os.File, error)
	OutFileClose                func(f *os.File)
	QuestradeOAuthURLFmt        string
	QuestradeTokenFile          string
	QuestradeAccountNumbers     []string
	Now                         time.Time
	CoinbaseBaseURL             string
}

func (c *ResolvedConn) Fetch() error {
	alphavantageConn := &alphavantage.Conn{
		Conf:            c.Conf.AlphavantageConfig,
		BaseURL:         c.AlphavantageBaseURL,
		APIKey:          c.AlphavantageAPIKey,
		BackoffDuration: c.AlphavantageBackoffDuration,
		BackoffRetry:    c.AlphavantageBackoffRetry,
	}
	r, err := alphavantageConn.Fetch()
	if err != nil {
		return errors.Wrap(err, "alphavantageConn.Fetch()")
	}
	sr, err := alphavantage.SortResponsesByDateThenBySymbol(r, c.CloseTime)
	if err != nil {
		return errors.Wrap(err, "alphavantage.SortResponsesByDateThenBySymbol()")
	}

	questradeConn := questrade.Conn{
		Conf:           c.Conf.QuestradeConfig,
		OAuthURLFmt:    c.QuestradeOAuthURLFmt,
		TokenFile:      c.QuestradeTokenFile,
		AccountNumbers: c.QuestradeAccountNumbers,
		CloseTime:      c.CloseTime,
		Now:            c.Now,
		StartDate:      c.StartDate,
	}
	sr2, err := questradeConn.Fetch()
	if err != nil {
		return errors.Wrap(err, "questradeConn.Fetch()")
	}

	coinbaseConn := &coinbase.Conn{
		Conf:    c.Conf.CoinbaseConfig,
		BaseURL: c.CoinbaseBaseURL,
		Now:     c.Now,
	}
	sr3, err := coinbaseConn.Fetch()
	if err != nil {
		return errors.Wrap(err, "coinbaseConn.Fetch()")
	}

	sr = append(append(sr, sr2...), sr3...)

	sr4, err := pricedb.GetDedupedSortedTimeSeriesItemWithSymbol(c.PriceDBData, c.CloseTime, makeReverseSymbolMap(c.Conf.Commodity), sr)
	if err != nil {
		return errors.Wrap(err, "pricedb.GetDedupedSortedTimeSeriesItemWithSymbol()")
	}

	sr = append(sr, sr4...)
	sort.Sort(priceutils.TimeSeriesItemWithSymbolSorter{sr})

	sr = c.filterOutPreStartDate(sr)
	if err := c.outputAsLedger(sr); err != nil {
		return errors.Wrap(err, "c.outputAsLedger()")
	}
	return nil
}

func (c *ResolvedConn) outputAsLedger(sr []*priceutils.TimeSeriesItemWithSymbol) error {
	f, err := c.OutFileOpen()
	if err != nil {
		return errors.Wrap(err, "c.OutFileOpen()")
	}
	defer c.OutFileClose(f)

	maxCommodityLength := c.getMaxCommodityLength(sr)

	blankDate := time.Time{}
	lastDate := blankDate
	for _, item := range sr {
		if item.Date != lastDate && lastDate != blankDate {
			fmt.Fprintf(f, "\n")
		}
		currency, display := c.getCurrencyAndDisplay(item)
		// TODO: do I need to convert time zone? likely not...
		fmt.Fprintf(f, "P %s %s%s%s%s\n", item.Date.Format(pricedb.DateTimeFormat), display, spaces(utf8.RuneCountInString(display), maxCommodityLength), currency, item.Data.GetLastPrice())
		lastDate = item.Date
	}
	return nil
}

func (c *ResolvedConn) getMaxCommodityLength(sr []*priceutils.TimeSeriesItemWithSymbol) int {
	if len(sr) == 0 {
		return 0
	}
	max := c.getCommodityLength(sr[0])
	for _, item := range sr {
		l := c.getCommodityLength(item)
		if l > max {
			max = l
		}
	}
	return max
}

func (c *ResolvedConn) getCommodityLength(item *priceutils.TimeSeriesItemWithSymbol) int {
	_, display := c.getCurrencyAndDisplay(item)
	return utf8.RuneCountInString(display)
}

func (c *ResolvedConn) getCurrencyAndDisplay(item *priceutils.TimeSeriesItemWithSymbol) (currency, display string) {
	currency = "$"
	display = item.Symbol
	if config, ok := c.Conf.Commodity[item.Symbol]; ok {
		if config.Currency != "" {
			currency = config.Currency
		}
		if config.Display != "" {
			display = config.Display
		}
	}
	return currency, display
}

func (c *ResolvedConn) filterOutPreStartDate(sr []*priceutils.TimeSeriesItemWithSymbol) []*priceutils.TimeSeriesItemWithSymbol {
	firstValid := -1
	for i, item := range sr {
		if item.Date.Equal(c.StartDate) || item.Date.After(c.StartDate) {
			firstValid = i
			break
		}
	}
	if firstValid < 0 {
		return nil
	}
	return sr[firstValid:]
}

func readPriceDB(path string) ([]string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "ioutil.ReadFile(%s)", path)
	}

	// TODO: is this necessary here? Maybe I can do that later.
	sp := strings.Split(string(data), "\n")
	ret := make([]string, 0, len(sp))
	for _, line := range sp {
		if !isWhitespaceOrComment(line) {
			ret = append(ret, line)
		}
	}

	return ret, nil
}

func isWhitespaceOrComment(s string) bool {
	s2 := strings.TrimSpace(s)
	return s2 == "" || strings.HasPrefix(s2, ";")
}

func spaces(commodityLength, maxCommodityLength int) string {
	desiredLength := maxCommodityLength + 2
	spacesNeeded := desiredLength - commodityLength
	return strings.Repeat(" ", spacesNeeded)
}

func makeReverseSymbolMap(c map[string]*CommodityConfig) map[string]string {
	m := make(map[string]string, len(c))
	for k, v := range c {
		if v.Display != "" {
			m[v.Display] = k
		}
	}
	return m
}

func readQuestradeAccountNumbers(path string) ([]string, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "ioutil.ReadFile(%s)", path)
	}
	return strings.Split(strings.TrimSpace(string(b)), ","), nil
}
