package lib

import (
	"testing"

	"bytes"
	"fmt"

	"github.com/prashantv/gostub"

	"github.com/glennhartmann/ledger-tools/src/pricedb"
)

func TestToCSV(t *testing.T) {
	stubs := gostub.New()
	defer stubs.Reset()

	var b bytes.Buffer
	stubs.Stub(&outWriter, &b)

	stubs.StubFunc(&pricedb.GetData, []byte(priceDB), nil)
	if err := ToCSV("", pricedb.DefaultCloseTime); err != nil {
		t.Errorf("ToCSV() = err(%+v)", err)
	}
	got := b.String()
	if got != wantCSV {
		t.Errorf("ToCSV() =\n%s\n\n\nwanted\n%s\n", got, wantCSV)
	}

	stubs.StubFunc(&pricedb.GetData, nil, fmt.Errorf("error"))
	if err := ToCSV("", pricedb.DefaultCloseTime); err == nil {
		t.Error("ToCSV() = err(nil), wanted an error")
	}
}

const priceDB = `
P 2021/01/18 19:23:00 £       $6.23635
P 2021/01/18 19:23:00 GOOG    £2362.428722

P 2021/02/19 12:42:40 BTC     $25135.3262473
P 2021/02/19 12:42:40 DOGE    $99.2384627935711

; this is a comment

P 2021/02/19 12:51:44 BTC     $34826.23897923
P 2021/02/19 12:51:44 DOGE    $0.112382582858

P 2021/02/19 18:30:01 BTC     $22384.1824282
P 2021/02/19 18:30:01 DOGE    $0.0000000342354

P 2021/02/26 18:30:02 BTC     $22932.24982324
P 2021/02/26 18:30:02 DOGE    $8.35983489234236

P 2021/02/27 22:45:00 £       $2.38532
P 2021/02/27 22:45:00 GOOG    USD$4382.385283
`

const wantCSV = `timestamp,symbol,currency,price
2021/01/18 19:23:00,GOOG,£,2362.428722
2021/01/18 19:23:00,£,$,6.23635
2021/02/19 12:42:40,BTC,$,25135.3262473
2021/02/19 12:42:40,DOGE,$,99.2384627935711
2021/02/19 12:51:44,BTC,$,34826.23897923
2021/02/19 12:51:44,DOGE,$,0.112382582858
2021/02/19 18:30:01,BTC,$,22384.1824282
2021/02/19 18:30:01,DOGE,$,0.0000000342354
2021/02/26 18:30:02,BTC,$,22932.24982324
2021/02/26 18:30:02,DOGE,$,8.35983489234236
2021/02/27 22:45:00,GOOG,USD$,4382.385283
2021/02/27 22:45:00,£,$,2.38532
`
