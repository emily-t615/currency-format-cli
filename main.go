package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

type result struct {
	Currency     string `json:"currency"`
	Input        string `json:"input"`
	InputFormat  string `json:"input_format"`
	Output       string `json:"output"`
	OutputFormat string `json:"output_format"`
	Exponent     int    `json:"exponent"`
}

func main() {
	amount := flag.String("amount", "", "amount to convert, e.g. 1234.56 or 123456")
	currency := flag.String("currency", "", "ISO 4217 currency code, e.g. USD")
	from := flag.String("from", "", "format of -amount: \"major\" or \"minor\"")
	jsonOut := flag.Bool("json", false, "emit JSON instead of a human-readable line")
	flag.Parse()

	if *amount == "" || *currency == "" || *from == "" {
		fmt.Fprintln(os.Stderr, "usage: currencyfmt -amount <value> -currency <code> -from <major|minor> [--json]")
		os.Exit(2)
	}

	exp, err := exponentFor(*currency)
	if err != nil {
		fail(err)
	}

	res := result{
		Currency: strings.ToUpper(*currency),
		Input:    *amount,
		Exponent: exp,
	}

	switch *from {
	case "major":
		minor, err := majorToMinor(*amount, exp)
		if err != nil {
			fail(err)
		}
		res.InputFormat = "major"
		res.OutputFormat = "minor"
		res.Output = fmt.Sprintf("%d", minor)
	case "minor":
		minor, err := parseMinor(*amount)
		if err != nil {
			fail(err)
		}
		res.InputFormat = "minor"
		res.OutputFormat = "major"
		res.Output = minorToMajor(minor, exp)
	default:
		fail(fmt.Errorf("-from must be \"major\" or \"minor\", got %q", *from))
	}

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(res); err != nil {
			fail(err)
		}
		return
	}

	fmt.Printf("%s %s (%s) -> %s (%s)\n", res.Input, res.Currency, res.InputFormat, res.Output, res.OutputFormat)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
