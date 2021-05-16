package pricedb

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/glennhartmann/ledger-tools/src/priceutils"
)

func TestGetDedupedSortedTimeSeriesItemWithSymbol(t *testing.T) {
	sm := make(map[string]string)
	got, err := GetDedupedSortedTimeSeriesItemWithSymbol(lines, DefaultCloseTime, sm, nil)
	if err != nil {
		t.Errorf("GetDedupedSortedTimeSeriesItemWithSymbol() = err(%+v)", err)
	}
	if !reflect.DeepEqual(got, wantTSIWSs) {
		t.Errorf("GetDedupedSortedTimeSeriesItemWithSymbol() = %s, wanted %s", priceutils.TimeSeriesItemWithSymbolSlice(got).String(), priceutils.TimeSeriesItemWithSymbolSlice(wantTSIWSs).String())
	}
}

func mustParseTime(s string) time.Time {
	parsed, err := time.Parse(DateTimeFormat, s)
	if err != nil {
		panic(fmt.Sprintf("unable to parse %q as time: %+v", s, err))
	}
	return parsed
}

var (
	lines = []string{
		"P 2021/01/18 19:23:00 £       $6.23635",
		"P 2021/01/18 19:23:00 GOOG    £2362.428722",
		"P 2021/02/19 12:42:40 BTC     $25135.3262473",
		"P 2021/02/19 12:42:40 DOGE    $99.2384627935711",
		"P 2021/02/19 12:51:44 BTC     $34826.23897923",
		"P 2021/02/19 12:51:44 DOGE    $0.112382582858",
		"P 2021/02/19 18:30:01 BTC     $22384.1824282",
		"P 2021/02/19 18:30:01 DOGE    $0.0000000342354",
		"P 2021/02/26 18:30:02 BTC     $22932.24982324",
		"P 2021/02/26 18:30:02 DOGE    $8.35983489234236",
		"P 2021/02/27 22:45:00 £       $2.38532",
		"P 2021/02/27 22:45:00 GOOG    USD$4382.385283",
	}
	wantTSIWSs = []*priceutils.TimeSeriesItemWithSymbol{
		&priceutils.TimeSeriesItemWithSymbol{
			Date:   mustParseTime("2021/01/18 19:23:00"),
			Symbol: "GOOG",
			Data:   &PriceData{"2362.428722", "£"},
		},
		&priceutils.TimeSeriesItemWithSymbol{
			Date:   mustParseTime("2021/01/18 19:23:00"),
			Symbol: "£",
			Data:   &PriceData{"6.23635", "$"},
		},
		&priceutils.TimeSeriesItemWithSymbol{
			Date:   mustParseTime("2021/02/19 12:42:40"),
			Symbol: "BTC",
			Data:   &PriceData{"25135.3262473", "$"},
		},
		&priceutils.TimeSeriesItemWithSymbol{
			Date:   mustParseTime("2021/02/19 12:42:40"),
			Symbol: "DOGE",
			Data:   &PriceData{"99.2384627935711", "$"},
		},
		&priceutils.TimeSeriesItemWithSymbol{
			Date:   mustParseTime("2021/02/19 12:51:44"),
			Symbol: "BTC",
			Data:   &PriceData{"34826.23897923", "$"},
		},
		&priceutils.TimeSeriesItemWithSymbol{
			Date:   mustParseTime("2021/02/19 12:51:44"),
			Symbol: "DOGE",
			Data:   &PriceData{"0.112382582858", "$"},
		},
		&priceutils.TimeSeriesItemWithSymbol{
			Date:   mustParseTime("2021/02/19 18:30:01"),
			Symbol: "BTC",
			Data:   &PriceData{"22384.1824282", "$"},
		},
		&priceutils.TimeSeriesItemWithSymbol{
			Date:   mustParseTime("2021/02/19 18:30:01"),
			Symbol: "DOGE",
			Data:   &PriceData{"0.0000000342354", "$"},
		},
		&priceutils.TimeSeriesItemWithSymbol{
			Date:   mustParseTime("2021/02/26 18:30:02"),
			Symbol: "BTC",
			Data:   &PriceData{"22932.24982324", "$"},
		},
		&priceutils.TimeSeriesItemWithSymbol{
			Date:   mustParseTime("2021/02/26 18:30:02"),
			Symbol: "DOGE",
			Data:   &PriceData{"8.35983489234236", "$"},
		},
		&priceutils.TimeSeriesItemWithSymbol{
			Date:   mustParseTime("2021/02/27 22:45:00"),
			Symbol: "GOOG",
			Data:   &PriceData{"4382.385283", "USD$"},
		},
		&priceutils.TimeSeriesItemWithSymbol{
			Date:   mustParseTime("2021/02/27 22:45:00"),
			Symbol: "£",
			Data:   &PriceData{"2.38532", "$"},
		},
	}
)
