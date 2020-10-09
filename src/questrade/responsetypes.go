package questrade

import (
	"fmt"

	"github.com/glennhartmann/ledger-tools/src/priceutils"
)

type DayData interface {
	priceutils.PriceData
}

type oauthResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	APIServer    string `json:"api_server"`
}

type symbolSearchResponse struct {
	Symbols []*symbol `json:"symbols"`
}

type symbol struct {
	Symbol          string `json:"symbol"`
	SymbolID        int    `json:"symbolId"`
	Description     string `json:"description"`
	SecurityType    string `json:"securityType"`
	ListingExchange string `json:"listingExchange"`
	IsTradable      bool   `json:"isTradable"`
	IsQuotable      bool   `json:"isQuotable"`
	Currency        string `json:"currency"`
}

type symbolResponse struct {
	Candles []*Candle `json:"candles"`
}

type Candle struct {
	Start  string  `json:"start"`
	End    string  `json:"end"`
	Low    float32 `json:"low"`
	High   float32 `json:"high"`
	Open   float32 `json:"open"`
	Close  float32 `json:"close"`
	Volume int     `json:"volume"`
	VWAP   float32 `json:"VWAP"`
}

func (c *Candle) GetLastPrice() string {
	return fmt.Sprintf("%f", c.Close)
}

type positionsResponse struct {
	Positions []*Position `json:"positions"`
}

type Position struct {
	Symbol             string   `json:"symbol"`
	SymbolID           int      `json:"symbolId"`
	OpenQuantity       float32  `json:"openQuantity"`
	ClosedQuantity     float32  `json:"closedQuantity"`
	CurrentMarketValue float32  `json:"currentMarketValue"`
	CurrentPrice       float32  `json:"currentPrice"`
	AverageEntryPrice  float32  `json:"averageEntryPrice"`
	DayPNL             *float32 `json:"dayPnl"`
	ClosedPNL          float32  `json:"closedPnl"`
	OpenPNL            float32  `json:"openPnl"`
	TotalCost          float32  `json:"totalCost"`
	IsRealTime         bool     `json:"isRealTime"`
	IsUnderReorg       bool     `json:"isUnderReorg"`
}

func (p *Position) GetLastPrice() string {
	return fmt.Sprintf("%f", p.CurrentPrice)
}
