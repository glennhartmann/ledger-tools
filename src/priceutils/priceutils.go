package priceutils

import (
	"time"
)

type PriceData interface {
	GetLastPrice() string
}

type TimeSeriesItemWithSymbol struct {
	Date   time.Time
	Symbol string
	Data   PriceData
}

type TimeSeriesItemWithSymbolSorter struct {
	TSIWS []*TimeSeriesItemWithSymbol
}

func (tsiwss TimeSeriesItemWithSymbolSorter) Len() int {
	return len(tsiwss.TSIWS)
}

func (tsiwss TimeSeriesItemWithSymbolSorter) Less(i, j int) bool {
	if tsiwss.TSIWS[i].Date.Equal(tsiwss.TSIWS[j].Date) {
		return tsiwss.TSIWS[i].Symbol < tsiwss.TSIWS[j].Symbol
	}
	return tsiwss.TSIWS[i].Date.Before(tsiwss.TSIWS[j].Date)
}

func (tsiwss TimeSeriesItemWithSymbolSorter) Swap(i, j int) {
	tmp := tsiwss.TSIWS[i]
	tsiwss.TSIWS[i] = tsiwss.TSIWS[j]
	tsiwss.TSIWS[j] = tmp
}
