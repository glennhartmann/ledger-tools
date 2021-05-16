package lib

import (
	"encoding/csv"
	"io"
	"log"
	"os"

	"github.com/pkg/errors"

	"github.com/glennhartmann/ledger-tools/src/pricedb"
)

var (
	// overridable for testing
	outWriter io.Writer = os.Stdout
)

func ToCSV(priceFile, closeTime string) error {
	lines, err := pricedb.ReadPriceDB(priceFile)
	if err != nil {
		return errors.Wrap(err, "pricedb.ReadPriceDB()")
	}

	symbolMap := make(map[string]string)
	tsiws, err := pricedb.GetSortedTimeSeriesItemWithSymbol(lines, closeTime, symbolMap)
	if err != nil {
		return errors.Wrap(err, "pricebd.GetSortedTimeSeriesItemWithSymbol()")
	}

	w := csv.NewWriter(outWriter)
	// wrap this part in a function so `defer w.Flush()` works properly
	cf := func() error {
		defer w.Flush()
		if err := w.Write([]string{"timestamp", "symbol", "currency", "price"}); err != nil {
			return errors.Wrap(err, "csv.Writer.Write(headers)")
		}
		for _, ts := range tsiws {
			lastCurrency := "UNK" // "unknown"
			lastPrice := ts.Data.GetLastPrice()
			if pc, ok := ts.Data.(*pricedb.PriceData); ok {
				lastCurrency = pc.LastCurrency
			} else {
				log.Printf("warning: unexpected type: %T (expected *pricedb.PriceData) on line (%v %q %q). Unable to determine currency.", ts.Data, ts.Date, ts.Symbol, lastPrice)
			}
			row := []string{ts.Date.Format(pricedb.DateTimeFormat), ts.Symbol, lastCurrency, lastPrice}
			if err := w.Write(row); err != nil {
				return errors.Wrapf(err, "csv.Writer.Write(%+v)", row)
			}
		}
		return nil
	}
	if err := cf(); err != nil {
		return errors.Wrap(err, "cf()")
	}

	return errors.Wrap(w.Error(), "csv.Writer.Error()")
}
