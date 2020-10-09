package coinbase

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/glennhartmann/ledger-tools/src/priceutils"

	"github.com/pkg/errors"
)

const (
	DefaultBaseURL = "https://api.coinbase.com/v2/exchange-rates?currency="
)

type Config struct {
	Currencies []string `json:"currencies"`
}

type Conn struct {
	Conf    *Config
	BaseURL string
	Now     time.Time
}

func (c *Conn) Fetch() ([]*priceutils.TimeSeriesItemWithSymbol, error) {
	ret := make([]*priceutils.TimeSeriesItemWithSymbol, 0, len(c.Conf.Currencies))
	for _, currency := range c.Conf.Currencies {
		url := c.BaseURL + currency
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
			return nil, err
		}
		log.Printf("fetch succeeded: %s", url)

		var parsedResponse Response
		if err := json.Unmarshal(responseBody, &parsedResponse); err != nil {
			return nil, errors.Wrapf(err, "json.Unmarshal(%s response)", currency)
		}
		ret = append(ret, &priceutils.TimeSeriesItemWithSymbol{c.Now, currency, &parsedResponse.Data.Rates})
	}
	sort.Sort(priceutils.TimeSeriesItemWithSymbolSorter{ret})
	return ret, nil
}
