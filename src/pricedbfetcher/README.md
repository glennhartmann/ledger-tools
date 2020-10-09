# pricedbfetcher

For basic usage, run `./pricedbfetcher -help`.

## Supported Data Sources

Like most APIs, these have query limits, so your queries may be rate-limited if you try to download too much at once. Check each one's official documentation for specifics.

### [Questrade](https://www.questrade.com/home)

You need an account to use this API, but you can use it to query full price history for many stocks, ETFs, etc - even ones you don't own.

1. Follow [Questrade's instructions](https://www.questrade.com/api/documentation/getting-started) to "create an app" and generate an API key.
2. Create a file on your computer with a comma-separated list of your account numbers (all on a signle line, no whitespace between them).
3. Create another file with your API key in it (and nothing else).
4. Use the `-questrade-account-numbers-file` and `-questrade-token-file` to let `pricedbfetcher` know where to find your files. Note that the token file will be overwritten.

We use the [markets/candles](https://www.questrade.com/api/documentation/rest-operations/market-calls/markets-candles-id) and [accounts/positions](https://www.questrade.com/api/documentation/rest-operations/account-calls/accounts-id-positions) API calls.

### [Coinbase](https://www.coinbase.com/)

This is a completely public, unauthenticated API (at least [the parts we need](https://developers.coinbase.com/api/v2#exchange-rates)), good for convenient access to current crypto prices (but not full history). No setup, registration, or local credentials files are needed.

### [Alpha Vantage](https://www.alphavantage.co/)

You'll need to sign up for Alpha Vantage, but it actually supports stocks, foreign-exchange, and crypto. I prefer using Questrade for stocks and Coinbase for crypto (don't remember why, but I think it may have been due to small query limit and long waits for rate-limiting). But it's certainly handy for FX, if nothing else.

1. [Register for an API key](https://www.alphavantage.co/support/#api-key)
2. Put your API key in a file (alone)
3. Use the `-alphavantage-api-key-file` flag to tell `pricedbfetcher` about your file.

API calls we may use (depending on which categories you want to use it for):
* [TIME_SERIES_DAILY](https://www.alphavantage.co/documentation/#daily)
* [FX_DAILY](https://www.alphavantage.co/documentation/#fx-daily)
* [DIGITAL_CURRENCY_DAILY](https://www.alphavantage.co/documentation/#currency-daily)


## Other Required Files

### config

This is a JSON file described by the [pricedbfetcher/lib/lib.go](https://github.com/glennhartmann/ledger-tools/blob/master/src/pricedbfetcher/lib/lib.go) Config struct. The `start_date` field should be a date in YYYY-MM-DD format. Each section is described in more detail below.

This file should be pointed to by the `-price-db-file` flag.

Example config:

```json
{
  "start_date": "2019-05-10",
  "alphavantage": {
    "forex_symbols": [
      "GBP",
      "USD"
    ],
    "cryptocurrency_symbols": [
      "BTC",
      "ETH"
    ],
    "stock_synbols": [
      "GOOG"
    ]
  },
  "questrade": {
    "market_symbols": [
      "XBAL.TO",
      "BNDX"
    ],
    "position_symbols": [
      "FAKE.SYMBOL"
    ]
  },
  "coinbase": {
    "currencies": [
      "LTC"
    ]
  },
  "commodity": {
    "GBP": {
      "display": "£"
    },
    "XBAL.TO": {
      "display": "\"XBAL.TO\"",
      "currency": "CAD"
    }
  }
}
```

#### alphavantage

Defined in [alphavantage/alphavantage.go](https://github.com/glennhartmann/ledger-tools/blob/master/src/alphavantage/alphavantage.go) Config struct.

* `forex_symbols`: array of foreign currencies to track
* `cryptocurrency_symbols`: array of crypto symbols to track
* `stock_synbols`: array of stocks to track

#### questrade

Defined in [questrade/questrade.go](https://github.com/glennhartmann/ledger-tools/blob/master/src/questrade/questrade.go) Config struct.

* `market_symbols`: stock symbols that Questrade can natively handle
* `position_symbols`: stock symbols that aren't in Questrade's normal database, but you have in your account (possibly due to a transfer from another account)

#### coinbase

Defined in [coinbase/coinbase.go](https://github.com/glennhartmann/ledger-tools/blob/master/src/coinbase/coinbase.go) Config struct.

* `currencies`: list of cryptocurrencies to track

#### commodity

Defined in [pricedbfetcher/lib/lib.go](https://github.com/glennhartmann/ledger-tools/blob/master/src/pricedbfetcher/lib/lib.go) CommodityConfig struct.

* object where each attribute name should be a symbol specified in one of the previous sections, and each attribute value should be an instance of an object with the following (optional) properties:
  * `display`: string to record in `price.db`. If unspecified, we'll use the symbol as written elsewhere in the file.
  * `currency`: currency to use for transactions of this commodity. If unspecified, we'll use '$'.

### price.db

A ledger price-db file (see [documentation](https://www.ledger-cli.org/3.0/doc/ledger3.html#Commodity-price-histories)). The file is expected to exist already, but may be empty. Anything in the file apart from comments or `P` statements may not be supported.
