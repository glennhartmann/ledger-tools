//go:generate protoc --proto_path=. --go_out=. --go_opt=paths=source_relative proto/priceutils.proto
package priceutils

import (
	"bytes"
	"fmt"
	"time"

	pb "github.com/glennhartmann/ledger-tools/src/priceutils/proto"

	"google.golang.org/protobuf/proto"
)

type PriceData interface {
	GetLastPrice() string
}

type TimeSeriesItemWithSymbol struct {
	Date   time.Time
	Symbol string
	Data   PriceData
}

func (tsiws *TimeSeriesItemWithSymbol) String() string {
	pds := ""
	if pdc, ok := tsiws.Data.(fmt.Stringer); ok {
		pds = pdc.String()
	} else {
		pds = fmt.Sprintf("%q", tsiws.Data.GetLastPrice())
	}

	return fmt.Sprintf("{%q, %q, %s}", tsiws.Date.String(), tsiws.Symbol, pds)
}

func (tsiws *TimeSeriesItemWithSymbol) ToProto() *pb.TimeSeriesItemWithSymbol {
	return pb.TimeSeriesItemWithSymbol_builder{
		TimeInUnixMicros: proto.Int64(tsiws.Date.UnixMicro()),
		Symbol:           proto.String(tsiws.Symbol),
		Data: pb.PriceData_builder{
			LastPrice: proto.String(tsiws.Data.GetLastPrice()),
		}.Build(),
	}.Build()
}

type TimeSeriesItemWithSymbolSlice []*TimeSeriesItemWithSymbol

func (tsiwss TimeSeriesItemWithSymbolSlice) String() string {
	var b bytes.Buffer

	fmt.Fprintf(&b, "{\n")
	for _, tsiws := range tsiwss {
		fmt.Fprintf(&b, "	%s\n", tsiws.String())
	}
	fmt.Fprintf(&b, "}")

	return b.String()
}

func (tsiwss TimeSeriesItemWithSymbolSlice) ToProto() *pb.TimeSeriesWithSymbol {
	items := make([]*pb.TimeSeriesItemWithSymbol, 0, len(tsiwss))
	for _, item := range tsiwss {
		items = append(items, item.ToProto())
	}
	return pb.TimeSeriesWithSymbol_builder{
		Items: items,
	}.Build()
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
