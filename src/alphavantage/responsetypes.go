package alphavantage

import (
	"bytes"
	"fmt"
	"sort"
	"time"

	"github.com/pkg/errors"

	"github.com/glennhartmann/ledger-tools/src/priceutils"
)

type Response interface {
	New() Response
	GetMetaData() MetaData
	GetTimeSeries() TimeSeries
}

type MetaData interface {
	GetInformation() string
	GetSymbol() string
	GetLastRefreshed() (time.Time, error)
}

type TimeSeries map[string]DayData

type DayData interface {
	priceutils.PriceData
	GetOpen() string
	GetHigh() string
	GetLow() string
	GetClose() string
}

func ResponseDebugString(r Response) string {
	return fmt.Sprintf("metadata:\n%s\ntimeseries:\n%s", MetaDataDebugString(r.GetMetaData()), TimeSeriesDebugString(r.GetTimeSeries()))
}

func MetaDataDebugString(m MetaData) string {
	lrs := ""
	lr, err := m.GetLastRefreshed()
	if err != nil {
		lrs = "error: " + err.Error()
	} else {
		lrs = lr.String()
	}
	return fmt.Sprintf(`    information: %s
    symbol: %s
    lastrefreshed: %s`, m.GetInformation(), m.GetSymbol(), lrs)
}

type timeSeriesSorter struct {
	date    string
	dayData DayData
}

type timeSeriesSorterSorter struct {
	tss []*timeSeriesSorter
}

func (tsss timeSeriesSorterSorter) Len() int {
	return len(tsss.tss)
}

func (tsss timeSeriesSorterSorter) Less(i, j int) bool {
	return tsss.tss[i].date < tsss.tss[j].date
}

func (tsss timeSeriesSorterSorter) Swap(i, j int) {
	tmp := tsss.tss[i]
	tsss.tss[i] = tsss.tss[j]
	tsss.tss[j] = tmp
}

func TimeSeriesDebugString(ts TimeSeries) string {
	s := make([]*timeSeriesSorter, 0, len(ts))
	for k, v := range ts {
		s = append(s, &timeSeriesSorter{k, v})
	}
	sort.Sort(timeSeriesSorterSorter{s})

	var buf bytes.Buffer
	for _, tss := range s {
		fmt.Fprintf(&buf, "    %s:\n", tss.date)
		fmt.Fprintf(&buf, "%s\n", DayDataDebugString(tss.dayData))
	}
	return buf.String()
}

func DayDataDebugString(dd DayData) string {
	return fmt.Sprintf(`        open: %s
        high: %s
        low: %s
        close: %s`, dd.GetOpen(), dd.GetHigh(), dd.GetLow(), dd.GetClose())
}

// *** STOCKS ***
type StockResponse struct {
	MetaData *StockMetadata `json:"Meta Data"`

	// do not access directly:
	TimeSeries          StockTimeSeries `json:"Time Series (Daily)"`
	ConvertedTimeSeries TimeSeries
}

type StockTimeSeries map[string]*StockDayData

func (sr *StockResponse) New() Response {
	return &StockResponse{}
}

func (sr *StockResponse) GetMetaData() MetaData {
	return sr.MetaData
}

func (sr *StockResponse) GetTimeSeries() TimeSeries {
	if sr.ConvertedTimeSeries == nil {
		sr.ConvertedTimeSeries = make(TimeSeries, len(sr.TimeSeries))
		for k, v := range sr.TimeSeries {
			sr.ConvertedTimeSeries[k] = v
		}
		sr.TimeSeries = nil
	}
	return sr.ConvertedTimeSeries
}

type StockMetadata struct {
	Information   string `json:"1. Information"`
	Symbol        string `json:"2. Symbol"`
	LastRefreshed string `json:"3. Last Refreshed"`
	OutputSize    string `json:"4. Output Size"`
	TimeZone      string `json:"5. Time Zone"`
}

func (sm *StockMetadata) GetInformation() string {
	return sm.Information
}

func (sm *StockMetadata) GetSymbol() string {
	return sm.Symbol
}

func (sm *StockMetadata) GetLastRefreshed() (time.Time, error) {
	return parseLastRefreshed("2006-01-02", sm.LastRefreshed, sm.TimeZone)
}

type StockDayData struct {
	Open   string `json:"1. open"`
	High   string `json:"2. high"`
	Low    string `json:"3. low"`
	Close  string `json:"4. close"`
	Volume string `json:"5 volume"`
}

func (sdd *StockDayData) GetLastPrice() string {
	return sdd.GetClose()
}

func (sdd *StockDayData) GetOpen() string {
	return sdd.Open
}

func (sdd *StockDayData) GetHigh() string {
	return sdd.High
}

func (sdd *StockDayData) GetLow() string {
	return sdd.Low
}

func (sdd *StockDayData) GetClose() string {
	return sdd.Close
}

// *** FOREX ***
type ForexResponse struct {
	MetaData *ForexMetadata `json:"Meta Data"`

	// do not access directly:
	TimeSeries          ForexTimeSeries `json:"Time Series FX (Daily)"`
	ConvertedTimeSeries TimeSeries
}

type ForexTimeSeries map[string]*ForexDayData

func (sr *ForexResponse) New() Response {
	return &ForexResponse{}
}

func (sr *ForexResponse) GetMetaData() MetaData {
	return sr.MetaData
}

func (sr *ForexResponse) GetTimeSeries() TimeSeries {
	if sr.ConvertedTimeSeries == nil {
		sr.ConvertedTimeSeries = make(TimeSeries, len(sr.TimeSeries))
		for k, v := range sr.TimeSeries {
			sr.ConvertedTimeSeries[k] = v
		}
		sr.TimeSeries = nil
	}
	return sr.ConvertedTimeSeries
}

type ForexMetadata struct {
	Information   string `json:"1. Information"`
	FromSymbol    string `json:"2. From Symbol"`
	ToSymbol      string `json:"3. To Symbol"`
	OutputSize    string `json:"4. Output Size"`
	LastRefreshed string `json:"5. Last Refreshed"`
	TimeZone      string `json:"6. Time Zone"`
}

func (sm *ForexMetadata) GetInformation() string {
	return sm.Information
}

func (sm *ForexMetadata) GetSymbol() string {
	return sm.FromSymbol
}

func (sm *ForexMetadata) GetLastRefreshed() (time.Time, error) {
	return parseLastRefreshed("2006-01-02 15:04:05", sm.LastRefreshed, sm.TimeZone)
}

type ForexDayData struct {
	Open  string `json:"1. open"`
	High  string `json:"2. high"`
	Low   string `json:"3. low"`
	Close string `json:"4. close"`
}

func (sdd *ForexDayData) GetLastPrice() string {
	return sdd.GetClose()
}

func (sdd *ForexDayData) GetOpen() string {
	return sdd.Open
}

func (sdd *ForexDayData) GetHigh() string {
	return sdd.High
}

func (sdd *ForexDayData) GetLow() string {
	return sdd.Low
}

func (sdd *ForexDayData) GetClose() string {
	return sdd.Close
}

// *** CRYPTOCURRENCY ***
type CryptocurrencyResponse struct {
	MetaData *CryptocurrencyMetadata `json:"Meta Data"`

	// do not access directly:
	TimeSeries          CryptocurrencyTimeSeries `json:"Time Series (Digital Currency Daily)"`
	ConvertedTimeSeries TimeSeries
}

type CryptocurrencyTimeSeries map[string]*CryptocurrencyDayData

func (sr *CryptocurrencyResponse) New() Response {
	return &CryptocurrencyResponse{}
}

func (sr *CryptocurrencyResponse) GetMetaData() MetaData {
	return sr.MetaData
}

func (sr *CryptocurrencyResponse) GetTimeSeries() TimeSeries {
	if sr.ConvertedTimeSeries == nil {
		sr.ConvertedTimeSeries = make(TimeSeries, len(sr.TimeSeries))
		for k, v := range sr.TimeSeries {
			sr.ConvertedTimeSeries[k] = v
		}
		sr.TimeSeries = nil
	}
	return sr.ConvertedTimeSeries
}

type CryptocurrencyMetadata struct {
	Information   string `json:"1. Information"`
	Symbol        string `json:"2. Digital Currency Code"`
	Name          string `json:"3. Digital Currency Name"`
	MarketCode    string `json:"4. Market Code"`
	MarketName    string `json:"5. Market Name"`
	LastRefreshed string `json:"6. Last Refreshed"`
	TimeZone      string `json:"7. Time Zone"`
}

func (sm *CryptocurrencyMetadata) GetInformation() string {
	return sm.Information
}

func (sm *CryptocurrencyMetadata) GetSymbol() string {
	return sm.Symbol
}

func (sm *CryptocurrencyMetadata) GetLastRefreshed() (time.Time, error) {
	return parseLastRefreshed("2006-01-02 15:04:05", sm.LastRefreshed, sm.TimeZone)
}

type CryptocurrencyDayData struct {
	CADOpen      string `json:"1a. open (CAD)"`
	CADHigh      string `json:"2a. high (CAD)"`
	CADLow       string `json:"3a. low (CAD)"`
	CADClose     string `json:"4a. close (CAD)"`
	USDOpen      string `json:"1b. open (USD)"`
	USDHigh      string `json:"2b. high (USD)"`
	USDLow       string `json:"3b. low (USD)"`
	USDClose     string `json:"4b. close (USD)"`
	Volume       string `json:"5 volume"`
	USDMarketCap string `json:"6. market cap (USD)"`
}

func (sdd *CryptocurrencyDayData) GetLastPrice() string {
	return sdd.GetClose()
}

func (sdd *CryptocurrencyDayData) GetOpen() string {
	return sdd.CADOpen
}

func (sdd *CryptocurrencyDayData) GetHigh() string {
	return sdd.CADHigh
}

func (sdd *CryptocurrencyDayData) GetLow() string {
	return sdd.CADLow
}

func (sdd *CryptocurrencyDayData) GetClose() string {
	return sdd.CADClose
}

// *** UTIL ***
func parseLastRefreshed(format, lastRefreshed, timeZone string) (time.Time, error) {
	loc, err := time.LoadLocation(timeZone)
	if err != nil {
		return time.Time{}, errors.Wrapf(err, "time.LoadLocation(%s)", timeZone)
	}
	t, err := time.ParseInLocation(format, lastRefreshed, loc)
	if err != nil {
		return time.Time{}, errors.Wrapf(err, "time.ParseInLocation(%s)", lastRefreshed)
	}
	return t, nil
}
