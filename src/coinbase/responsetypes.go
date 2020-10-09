package coinbase

import (
	"github.com/glennhartmann/ledger-tools/src/priceutils"
)

type DayData interface {
	priceutils.PriceData
}

type Response struct {
	Data *Data `json:"data"`
}

type Data struct {
	Currency string `json:"currency"`
	Rates    Rates  `json:"rates"`
}

type Rates struct {
	CAD string `json:"CAD"`
	USD string `json:"USD"`
}

func (r *Rates) GetLastPrice() string {
	return r.CAD
}
