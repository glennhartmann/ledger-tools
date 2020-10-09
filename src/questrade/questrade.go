package questrade

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
	DefaultOAuthURLFmt        = "https://login.questrade.com/oauth2/token?grant_type=refresh_token&refresh_token=%s"
	DefaultTokenFile          = "~/.questrade_token"
	DefaultAccountNumbersFile = "~/.questrade_account_numbers"

	dateTimeFormat = "2006-01-02T15:04:05.999999-07:00"
)

var (
	symbolQueryTemplate = template.Must(template.New("query").Parse("v1/markets/candles/{{.SymbolID}}?startTime={{.StartTime}}&endTime={{.EndTime}}&interval=OneDay"))
)

type Config struct {
	MarketSymbols   []string `json:"market_symbols"`
	PositionSymbols []string `json:"position_symbols"`
}

type Conn struct {
	Conf           *Config
	OAuthURLFmt    string
	TokenFile      string
	AccountNumbers []string
	CloseTime      string
	Now            time.Time
	StartDate      time.Time
}

func (c *Conn) Fetch() ([]*priceutils.TimeSeriesItemWithSymbol, error) {
	b, err := ioutil.ReadFile(c.TokenFile)
	if err != nil {
		return nil, errors.Wrapf(err, "ioutil.ReadFile(%s)", c.TokenFile)
	}
	token := strings.TrimSpace(string(b))

	oauthResponse, err := Authenticate(token, c.TokenFile, c.OAuthURLFmt)
	if err != nil {
		return nil, errors.Wrapf(err, "Authenticate(%s)", token)
	}

	tsiws := make([]*priceutils.TimeSeriesItemWithSymbol, 0, len(c.Conf.MarketSymbols)+10*len(c.AccountNumbers))
	for _, symbol := range c.Conf.MarketSymbols {
		symbolResponse, err := FetchSymbol(oauthResponse, symbol, c.StartDate, c.Now)
		if err != nil {
			return nil, errors.Wrapf(err, "FetchSymbol(%s)", symbol)
		}
		for _, candle := range symbolResponse {
			d, err := time.Parse(dateTimeFormat, candle.Start)
			if err != nil {
				return nil, errors.Wrapf(err, "time.Parse(%s)", candle.Start)
			}
			d, err = c.dateAtCloseTime(d)
			if err != nil {
				return nil, errors.Wrap(err, "dateAtCloseTime()")
			}
			tsiws = append(tsiws, &priceutils.TimeSeriesItemWithSymbol{d, symbol, candle})
		}
	}

	positionSymbols := makePositionSymbolsMap(c.Conf.PositionSymbols)
	seenPositionSymbols := make(map[string]struct{}, len(positionSymbols))
	for _, accountNumber := range c.AccountNumbers {
		positions, err := FetchPositions(oauthResponse, accountNumber)
		if err != nil {
			return nil, errors.Wrapf(err, "FetchPositions(%s)", accountNumber)
		}
		for _, position := range positions {
			if _, ok := positionSymbols[position.Symbol]; ok {
				seenPositionSymbols[position.Symbol] = struct{}{}
				tsiws = append(tsiws, &priceutils.TimeSeriesItemWithSymbol{c.Now, position.Symbol, position})
			}
		}
	}
	if err := checkSeenPositionSymbols(positionSymbols, seenPositionSymbols); err != nil {
		return nil, errors.Wrap(err, "checkSeenPositionSymbols()")
	}

	sort.Sort(priceutils.TimeSeriesItemWithSymbolSorter{tsiws})
	return tsiws, nil
}

func (c *Conn) dateAtCloseTime(t time.Time) (time.Time, error) {
	date := t.Format("2006-01-02")
	dateTime := fmt.Sprintf("%s %s", date, c.CloseTime)
	d, err := time.Parse("2006-01-02 15:04:05", dateTime)
	if err != nil {
		return time.Time{}, errors.Wrapf(err, "time.Parse(%s)", dateTime)
	}
	return d, nil
}

// TODO: this whole file is badly in need of a refactor
func Authenticate(token, tokenFile, oauthURLFmt string) (*oauthResponse, error) {
	oauthURL := fmt.Sprintf(oauthURLFmt, token)
	var responseBody []byte
	if err := func() error {
		log.Printf("starting fetch: %s", oauthURL)
		resp, err := http.Get(oauthURL)
		if err != nil {
			return errors.Wrapf(err, "http.Get(%s)", oauthURL)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return errors.Errorf("http.Get() returned status %s for %s", resp.Status, oauthURL)
		}
		responseBody, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "ioutil.ReadAll(resp.Body)")
		}
		return nil
	}(); err != nil {
		return nil, errors.Wrapf(err, "oauth fetch: %s", oauthURL)
	}
	log.Printf("fetch succeeded: %s", oauthURL)

	var oauthResponse oauthResponse
	if err := json.Unmarshal(responseBody, &oauthResponse); err != nil {
		return nil, errors.Wrap(err, "json.Unmarshal(oauth response)")
	}

	if err := ioutil.WriteFile(tokenFile, []byte(oauthResponse.RefreshToken), 0640); err != nil {
		log.Printf("unable to write refresh token back to %s: %v", tokenFile, err)
	}

	return &oauthResponse, nil
}

func FetchSymbol(oauthResponse *oauthResponse, symbolToSearch string, startTime, endTime time.Time) ([]*Candle, error) {
	raw, err := FetchRawSymbol(oauthResponse, symbolToSearch, startTime, endTime)
	if err != nil {
		return nil, errors.Wrapf(err, "FetchRawSymbol(%s)", symbolToSearch)
	}

	var symbolResponse symbolResponse
	if err := json.Unmarshal([]byte(raw), &symbolResponse); err != nil {
		return nil, errors.Wrapf(err, "json.Unmarshal(symbol response for %s)", symbolToSearch)
	}

	return symbolResponse.Candles, nil
}

type symbolQueryParams struct {
	SymbolID  int
	StartTime string
	EndTime   string
}

func FetchRawSymbol(oauthResponse *oauthResponse, symbolToSearch string, startTime, endTime time.Time) (string, error) {
	var responseBody []byte
	fetchURL := fmt.Sprintf("%sv1/symbols/search?prefix=%s", oauthResponse.APIServer, symbolToSearch)
	if err := func() error {
		client := http.Client{}
		req, err := http.NewRequest("GET", fetchURL, nil)
		if err != nil {
			return errors.Wrapf(err, "http.NewRequest(%s)", fetchURL)
		}
		req.Header.Add("Authorization", fmt.Sprintf("%s %s", oauthResponse.TokenType, oauthResponse.AccessToken))
		log.Printf("starting fetch: %s", fetchURL)
		resp, err := client.Do(req)
		if err != nil {
			return errors.Wrapf(err, "http.Get(%s)", fetchURL)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return errors.Errorf("http.Get() returned status %s for %s", resp.Status, fetchURL)
		}
		responseBody, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "ioutil.ReadAll(resp.Body)")
		}
		return nil
	}(); err != nil {
		return "", errors.Wrapf(err, "symbol search fetch: %s (%s)", symbolToSearch, fetchURL)
	}
	log.Printf("fetch succeeded: %s", fetchURL)

	var searchResponse symbolSearchResponse
	if err := json.Unmarshal(responseBody, &searchResponse); err != nil {
		return "", errors.Wrapf(err, "json.Unmarshal(%s symbol search response)", symbolToSearch)
	}

	symbol := findSymbol(&searchResponse, symbolToSearch)
	if symbol == nil {
		return "", errors.Errorf("couldn't find %s", symbolToSearch)
	}

	var queryBuf bytes.Buffer
	if err := symbolQueryTemplate.Execute(&queryBuf, &symbolQueryParams{
		SymbolID:  symbol.SymbolID,
		StartTime: startTime.Format(dateTimeFormat),
		EndTime:   endTime.Format(dateTimeFormat),
	}); err != nil {
		return "", errors.Wrap(err, "symbolQueryTemplate.Execute()")
	}

	fetchURL = oauthResponse.APIServer + queryBuf.String()
	if err := func() error {
		client := http.Client{}
		req, err := http.NewRequest("GET", fetchURL, nil)
		if err != nil {
			return errors.Wrapf(err, "http.NewRequest(%s)", fetchURL)
		}
		req.Header.Add("Authorization", fmt.Sprintf("%s %s", oauthResponse.TokenType, oauthResponse.AccessToken))
		log.Printf("starting fetch: %s", fetchURL)
		resp, err := client.Do(req)
		if err != nil {
			return errors.Wrapf(err, "http.Get(%s)", fetchURL)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return errors.Errorf("http.Get() returned status %s for %s", resp.Status, fetchURL)
		}
		responseBody, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "ioutil.ReadAll(resp.Body)")
		}
		return nil
	}(); err != nil {
		return "", errors.Wrapf(err, "symbol fetch: %s (%s)", symbolToSearch, fetchURL)
	}
	log.Printf("fetch succeeded: %s", fetchURL)

	return string(responseBody), nil
}

func FetchPositions(oauthResponse *oauthResponse, accountNumber string) ([]*Position, error) {
	raw, err := FetchRawPositions(oauthResponse, accountNumber)
	if err != nil {
		return nil, errors.Wrapf(err, "FetchRawPositions(%s)", accountNumber)
	}

	var positionsResponse positionsResponse
	if err := json.Unmarshal([]byte(raw), &positionsResponse); err != nil {
		return nil, errors.Wrapf(err, "json.Unmarshal(position response for %s)", accountNumber)
	}

	return positionsResponse.Positions, nil
}

func FetchRawPositions(oauthResponse *oauthResponse, accountNumber string) (string, error) {
	var responseBody []byte
	fetchURL := fmt.Sprintf("%sv1/accounts/%s/positions", oauthResponse.APIServer, accountNumber)
	if err := func() error {
		client := http.Client{}
		req, err := http.NewRequest("GET", fetchURL, nil)
		if err != nil {
			return errors.Wrapf(err, "http.NewRequest(%s)", fetchURL)
		}
		req.Header.Add("Authorization", fmt.Sprintf("%s %s", oauthResponse.TokenType, oauthResponse.AccessToken))
		log.Printf("starting fetch: %s", fetchURL)
		resp, err := client.Do(req)
		if err != nil {
			return errors.Wrapf(err, "http.Get(%s)", fetchURL)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return errors.Errorf("http.Get() returned status %s for %s", resp.Status, fetchURL)
		}
		responseBody, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "ioutil.ReadAll(resp.Body)")
		}
		return nil
	}(); err != nil {
		return "", errors.Wrapf(err, "position fetch failed for account: %s (%s)", accountNumber, fetchURL)
	}
	log.Printf("fetch succeeded: %s", fetchURL)

	return string(responseBody), nil
}

func findSymbol(searchResponse *symbolSearchResponse, symbolToSearch string) *symbol {
	for _, symbol := range searchResponse.Symbols {
		if symbol.Symbol == symbolToSearch {
			return symbol
		}
	}
	return nil
}

func makePositionSymbolsMap(ps []string) map[string]struct{} {
	r := make(map[string]struct{}, len(ps))
	for _, item := range ps {
		r[item] = struct{}{}
	}
	return r
}

func checkSeenPositionSymbols(positionSymbols, seenPositionSymbols map[string]struct{}) error {
	for symbol := range positionSymbols {
		if _, ok := seenPositionSymbols[symbol]; !ok {
			return errors.Errorf("did not find %s in positions", symbol)
		}
	}
	return nil
}
