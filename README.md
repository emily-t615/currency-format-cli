# currencyfmt

A small command-line tool for converting currency amounts between the two
formats that show up constantly and don't mix well: "minor units" (the
integer cents/pence/fils that payment APIs and databases store) and "major
units" (the decimal amount a human reads, like `1234.56`).

The annoying part isn't the arithmetic, it's that the conversion factor
depends on the currency. Most currencies use two decimal places, but yen and
won use zero, and a few Gulf currencies use three. Doing this by hand with
floating point is also a good way to introduce off-by-one-cent bugs. This
tool does the conversion with integer/string arithmetic and a currency-aware
table of decimal places, so `1234.56` USD and `123456` minor units always
round-trip exactly.

## Usage

Convert a human-readable amount into minor units:

```
$ go run . -amount 1234.56 -currency USD -from major
1234.56 USD (major) -> 123456 (minor)
```

Convert minor units back into a decimal amount:

```
$ go run . -amount 123456 -currency USD -from minor
123456 USD (minor) -> 1234.56 (major)
```

Yen has no minor units, so the conversion factor is 1:

```
$ go run . -amount 5000 -currency JPY -from major
5000 JPY (major) -> 5000 (minor)
```

Bahraini dinar uses three decimal places:

```
$ go run . -amount 1.500 -currency BHD -from major
1.500 BHD (major) -> 1500 (minor)
```

### JSON output

Pass `--json` to get machine-readable output instead, for piping into other
tools or scripts:

```
$ go run . -amount 1234.56 -currency USD -from major --json
{
  "currency": "USD",
  "input": "1234.56",
  "input_format": "major",
  "output": "123456",
  "output_format": "minor",
  "exponent": 2
}
```

## Flags

- `-amount` the value to convert
- `-currency` an ISO 4217 currency code (see `currencyExponents` in
  `convert.go` for the currently supported list)
- `-from` which format `-amount` is in: `major` or `minor` (the tool always
  converts to the other one)
- `--json` emit the JSON object shown above instead of a one-line summary

An amount with more fractional digits than the currency allows (e.g.
`1.234` for USD) is rejected rather than silently rounded.

## Building

```
go build -o currencyfmt .
```

Standard library only, no external dependencies.

## Status

Early skeleton. The currency table covers the active ISO 4217 circulating
currencies (see `currencyExponents` in `convert.go`); unknown codes return
an error rather than guessing. It does not include the non-country "funds
code" entries (BOV, USN, and similar) or the precious-metal/SDR codes.
