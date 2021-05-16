package pricedb

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/glennhartmann/ledger-tools/src/priceutils"
)

const (
	DefaultFile      = "~/.price.db"
	DefaultCloseTime = "22:45:00"
	DateTimeFormat   = "2006/01/02 15:04:05"

	unquotedCommodityRxStr = `[^\s"]+`
	quotedCommodityRxStr   = `"[^"]+"`
	lineRxFmt              = `^P\s+(\d\d\d\d\/\d\d\/\d\d \d\d:\d\d:\d\d)\s+((%s)|(%s))[^\d]+((\d,?)+(\.\d+)?)$`
)

var (
	lineRx = regexp.MustCompile(fmt.Sprintf(lineRxFmt, unquotedCommodityRxStr, quotedCommodityRxStr))
)

type PriceData struct {
	LastPrice string
}

func (pd *PriceData) GetLastPrice() string {
	return pd.LastPrice
}

func ReadPriceDB(path string) ([]string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "GetData(%s)", path)
	}

	sp := strings.Split(string(data), "\n")
	ret := make([]string, 0, len(sp))
	for _, line := range sp {
		if !IsWhitespaceOrComment(line) {
			ret = append(ret, line)
		}
	}

	return ret, nil
}

func IsWhitespaceOrComment(s string) bool {
	s2 := strings.TrimSpace(s)
	return s2 == "" || strings.HasPrefix(s2, ";")
}

func GetSortedTimeSeriesItemWithSymbol(lines []string, closeTime string, symbolMap map[string]string) ([]*priceutils.TimeSeriesItemWithSymbol, error) {
	return GetDedupedSortedTimeSeriesItemWithSymbol(lines, closeTime, symbolMap, nil)
}

func GetDedupedSortedTimeSeriesItemWithSymbol(lines []string, closeTime string, symbolMap map[string]string, others []*priceutils.TimeSeriesItemWithSymbol) ([]*priceutils.TimeSeriesItemWithSymbol, error) {
	ret := make([]*priceutils.TimeSeriesItemWithSymbol, 0, len(lines))
	ds := makeDateSymbolSet(others)
	for _, line := range lines {
		tl := strings.TrimSpace(line)
		r := lineRx.FindStringSubmatch(tl)
		if len(r) != 8 {
			return nil, errors.Errorf("expected 6 Rx submatches, got %d (line: %s)", len(r), line)
		}
		dateTimeStr := r[1]
		symbol := r[2]
		if s, ok := symbolMap[symbol]; ok {
			symbol = s
		}
		price := r[5]
		dts := strings.Split(dateTimeStr, " ")
		if len(dts) != 2 {
			return nil, errors.Errorf("expected 2 string split pieces, got %d", len(dts))
		}
		timeOnlyStr := dts[1]
		d, err := time.Parse(DateTimeFormat, dateTimeStr)
		if err != nil {
			return nil, errors.Wrapf(err, "time.Parse(%s)", dateTimeStr)
		}
		fds := formatDateSymbol(d, symbol)
		if _, ok := ds[fds]; !ok || timeOnlyStr < closeTime {
			ret = append(ret, &priceutils.TimeSeriesItemWithSymbol{d, symbol, &PriceData{price}})
		}
	}
	sort.Sort(priceutils.TimeSeriesItemWithSymbolSorter{ret})
	return ret, nil
}

type dateSymbolSet map[string]struct{}

func makeDateSymbolSet(others []*priceutils.TimeSeriesItemWithSymbol) dateSymbolSet {
	ds := make(dateSymbolSet, len(others))
	for _, item := range others {
		ds[formatDateSymbol(item.Date, item.Symbol)] = struct{}{}
	}
	return ds
}

func formatDateSymbol(date time.Time, symbol string) string {
	return fmt.Sprintf("%s_%s", date.Format("2006/01/02"), symbol)
}
