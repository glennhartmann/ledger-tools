package main

import (
	"fmt"
	"os"

	"github.com/glennhartmann/ledger-tools/src/pricedb"
	"github.com/glennhartmann/ledger-tools/src/priceutils"

	flag "github.com/spf13/pflag"
	enumflag "github.com/thediveo/enumflag/v2"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

type outputType enumflag.Flag

const (
	json outputType = iota
	protoText
	protoWire
)

var outputTypeIDs = map[outputType][]string{
	json:      {"json", "j"},
	protoText: {"proto-text", "text-proto", "textpb", "pbascii", "tpb", "pba"},
	protoWire: {"proto-wire", "proto-binary", "binary-proto", "wire-proto", "proto", "pb"},
}

var (
	pricedbPath = flag.StringP("price-db-path", "p", pricedb.DefaultFile, "Path to the price.db file.")
	closeTime   = flag.StringP("close-time", "c", pricedb.DefaultCloseTime, "Close time in '15:04:05' format.")

	outputTypeFlag outputType
)

func main() {
	flag.VarP(enumflag.New(&outputTypeFlag, "outputType", outputTypeIDs, enumflag.EnumCaseInsensitive), "output-type", "o", fmt.Sprintf("Format of output. Valid values are %q (aliases %q), %q (aliases %q), or %q (aliases %q)", outputTypeIDs[json][0], outputTypeIDs[json][1:], outputTypeIDs[protoText][0], outputTypeIDs[protoText][1:], outputTypeIDs[protoWire][0], outputTypeIDs[protoWire][1:]))

	flag.Parse()

	lines, err := pricedb.ReadPriceDB(*pricedbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pricedb.ReadPriceDB(): %+v\n", err)
		os.Exit(1)
	}

	symbolMap := make(map[string]string)
	tsiws, err := pricedb.GetSortedTimeSeriesItemWithSymbol(lines, *closeTime, symbolMap)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pricebd.GetSortedTimeSeriesItemWithSymbol(): %+v\n", err)
		os.Exit(1)
	}

	p := priceutils.TimeSeriesItemWithSymbolSlice(tsiws).ToProto()

	var m interface {
		Marshal(m proto.Message) ([]byte, error)
	}

	switch outputTypeFlag {
	case json:
		m = protojson.MarshalOptions{
			Multiline: true,
			Indent:    "    ",
		}
	case protoText:
		m = prototext.MarshalOptions{
			Multiline: true,
			Indent:    "  ",
		}
	case protoWire:
		m = proto.MarshalOptions{}
	default:
		fmt.Fprintf(os.Stderr, "invalid --output-type: %v\n", outputTypeFlag)
		os.Exit(1)
	}

	b, err := m.Marshal(p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "m.Marshal(): %+v\n", err)
		os.Exit(1)
	}

	if _, err := os.Stdout.Write(b); err != nil {
		fmt.Fprintf(os.Stderr, "os.Stdout.Write(): %+v\n", err)
		os.Exit(1)
	}
}
