package alphavantage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/pkg/errors"

	"github.com/glennhartmann/ledger-tools/src/priceutils"
)

const (
	DefaultBaseURL         = "https://www.alphavantage.co/query"
	DefaultAPIKeyFile      = "~/.alphavantage_api_key"
	DefaultBackoffDuration = 1 * time.Minute
	DefaultBackoffRetry    = 3
)

var (
	queryTemplate = template.Must(template.New("query").Parse("?function={{.Function}}&symbol={{.Symbol}}&from_symbol={{.FromSymbol}}&to_symbol=CAD&market=CAD&outputsize=full&apikey={{.APIKey}}"))
)

const (
	stockFunction          = "TIME_SERIES_DAILY"
	forexFunction          = "FX_DAILY"
	cryptocurrencyFunction = "DIGITAL_CURRENCY_DAILY"
)

type Config struct {
	StockSymbols          []string `json:"stock_symbols"`
	ForexSymbols          []string `json:"forex_symbols"`
	CryptocurrencySymbols []string `json:"cryptocurrency_symbols"`
}

type Conn struct {
	Conf            *Config
	BaseURL         string
	APIKey          string
	BackoffDuration time.Duration
	BackoffRetry    int
}

func (c *Conn) Fetch() ([]Response, error) {
	stock, err := c.fetchType(c.Conf.StockSymbols, stockFunction, &StockResponse{})
	if err != nil {
		return nil, errors.Wrap(err, "c.fetchType(stock)")
	}
	forex, err := c.fetchType(c.Conf.ForexSymbols, forexFunction, &ForexResponse{})
	if err != nil {
		return nil, errors.Wrap(err, "c.fetchType(forex)")
	}
	crypto, err := c.fetchType(c.Conf.CryptocurrencySymbols, cryptocurrencyFunction, &CryptocurrencyResponse{})
	if err != nil {
		return nil, errors.Wrap(err, "c.fetchType(cryptocurrency)")
	}
	return append(append(stock, forex...), crypto...), nil
}

func SortResponsesByDateThenBySymbol(rs []Response, closeTime string) ([]*priceutils.TimeSeriesItemWithSymbol, error) {
	tsiws := make([]*priceutils.TimeSeriesItemWithSymbol, 0, 50)
	for _, r := range rs {
		s := r.GetMetaData().GetSymbol()
		for date, data := range r.GetTimeSeries() {
			dateTime := fmt.Sprintf("%s %s", date, closeTime)
			d, err := time.Parse("2006-01-02 15:04:05", dateTime)
			if err != nil {
				return nil, errors.Wrapf(err, "time.Parse(%s)", dateTime)
			}
			tsiws = append(tsiws, &priceutils.TimeSeriesItemWithSymbol{
				Date:   d,
				Symbol: s,
				Data:   data,
			})
		}
	}
	sort.Sort(priceutils.TimeSeriesItemWithSymbolSorter{tsiws})
	return tsiws, nil
}

type queryParams struct {
	Function   string
	Symbol     string
	APIKey     string
	FromSymbol string
}

func (c *Conn) fetchType(symbols []string, function string, responsePrototype Response) ([]Response, error) {
	r := make([]Response, 0, len(symbols))
	for _, symbol := range symbols {
		parsedResponse, err := c.fetchSymbolWithBackoff(symbol, function, responsePrototype)
		if err != nil {
			return nil, errors.Wrap(err, "c.fetchSymbolWithBackoff()")
		}
		//fmt.Printf("%s\n", ResponseDebugString(parsedResponse))
		r = append(r, parsedResponse)
	}
	return r, nil
}

type rateLimitResponse struct {
	Note string `json:"Note"`
}

func (c *Conn) fetchSymbolWithBackoff(symbol string, function string, responsePrototype Response) (Response, error) {
	for i := 0; i < c.BackoffRetry; i++ {
		parsedResponse, backoff, err := c.fetchSymbol(symbol, function, responsePrototype)
		if err != nil {
			return nil, errors.Wrap(err, "c.fetchSymbol()")
		}
		if !backoff {
			return parsedResponse, nil
		}
		log.Printf("rate-limited, backing off for %s (attempt %d of %d)", c.BackoffDuration.String(), i+1, c.BackoffRetry)
		time.Sleep(c.BackoffDuration)
	}
	return nil, errors.New("exhausted backoff retry limit")
}

func (c *Conn) fetchSymbol(symbol string, function string, responsePrototype Response) (parsedResponse Response, backoff bool, err error) {
	var queryBuf bytes.Buffer
	if err := queryTemplate.Execute(&queryBuf, &queryParams{
		Function:   function,
		Symbol:     symbol,
		FromSymbol: symbol,
		APIKey:     c.APIKey,
	}); err != nil {
		return nil, false, errors.Wrap(err, "queryTemplate.Execute()")
	}

	url := c.BaseURL + queryBuf.String()
	var responseBody []byte
	if err := func() error {
		log.Printf("starting fetch: %s", url)
		resp, err := http.Get(url)
		if err != nil {
			return errors.Wrapf(err, "http.Get(%s)", url)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return errors.Errorf("http.Get() returned status %s for %s", resp.Status, url)
		}
		responseBody, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "ioutil.ReadAll(resp.Body)")
		}
		return nil
	}(); err != nil {
		log.Printf("fetch failed: %s (%v)", url, err)
		return nil, false, err
	}

	var rls rateLimitResponse
	if err := json.Unmarshal(responseBody, &rls); err != nil {
		return nil, false, errors.Wrapf(err, "json.Unmarshal(ratelimit %s %s response)", symbol, function)
	}
	if strings.Contains(rls.Note, "API call frequency") {
		return nil, true, nil
	}

	log.Printf("fetch succeeded: %s", url)
	parsedResponse = responsePrototype.New()
	if err := json.Unmarshal(responseBody, parsedResponse); err != nil {
		return nil, false, errors.Wrapf(err, "json.Unmarshal(%s %s response)", symbol, function)
	}
	return parsedResponse, false, nil
}
