package main

import (
	"bufio"
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
	batch := flag.Bool("batch", false, "read amounts from stdin, one per line, instead of -amount")
	flag.Parse()

	if *currency == "" || *from == "" || (*from != "major" && *from != "minor") {
		fmt.Fprintln(os.Stderr, "usage: currencyfmt -amount <value> -currency <code> -from <major|minor> [--json]")
		fmt.Fprintln(os.Stderr, "       currencyfmt -batch -currency <code> -from <major|minor> [--json] < amounts.txt")
		os.Exit(2)
	}

	exp, err := exponentFor(*currency)
	if err != nil {
		fail(err)
	}

	if *batch {
		if *amount != "" {
			fail(fmt.Errorf("-amount cannot be combined with -batch; amounts are read from stdin"))
		}
		runBatch(*currency, *from, exp, *jsonOut)
		return
	}

	if *amount == "" {
		fmt.Fprintln(os.Stderr, "usage: currencyfmt -amount <value> -currency <code> -from <major|minor> [--json]")
		os.Exit(2)
	}

	res, err := convertOne(*amount, *currency, *from, exp)
	if err != nil {
		fail(err)
	}
	printResult(res, *jsonOut, true)
}

// runBatch reads one amount per line from stdin and converts each using the
// same currency and direction, so a whole file of amounts can be piped
// through without spawning a process per line. Blank lines and lines
// starting with # are skipped. A bad line is reported and skipped rather
// than aborting the rest of the batch, and the process exits non-zero if
// any line failed.
func runBatch(currency, from string, exp int, jsonOut bool) {
	scanner := bufio.NewScanner(os.Stdin)
	lineNum := 0
	hadError := false

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		res, err := convertOne(line, currency, from, exp)
		if err != nil {
			fmt.Fprintf(os.Stderr, "line %d: error: %v\n", lineNum, err)
			hadError = true
			continue
		}
		printResult(res, jsonOut, false)
	}
	if err := scanner.Err(); err != nil {
		fail(err)
	}
	if hadError {
		os.Exit(1)
	}
}

// convertOne runs a single amount through the major/minor conversion and
// builds the result struct shared by both single-amount and batch modes.
func convertOne(amount, currency, from string, exp int) (result, error) {
	res := result{
		Currency: strings.ToUpper(currency),
		Input:    amount,
		Exponent: exp,
	}

	switch from {
	case "major":
		minor, err := majorToMinor(amount, exp)
		if err != nil {
			return result{}, err
		}
		res.InputFormat = "major"
		res.OutputFormat = "minor"
		res.Output = fmt.Sprintf("%d", minor)
	case "minor":
		minor, err := parseMinor(amount)
		if err != nil {
			return result{}, err
		}
		res.InputFormat = "minor"
		res.OutputFormat = "major"
		res.Output = minorToMajor(minor, exp)
	default:
		return result{}, fmt.Errorf("-from must be \"major\" or \"minor\", got %q", from)
	}
	return res, nil
}

// printResult writes one converted result. In batch mode JSON output is
// NDJSON (one compact object per line) so it can be streamed and re-parsed
// line by line; single-amount JSON output stays indented for readability.
func printResult(res result, jsonOut, indent bool) {
	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		if indent {
			enc.SetIndent("", "  ")
		}
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
