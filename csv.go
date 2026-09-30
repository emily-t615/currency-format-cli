package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// csvOptions describes how convertCSV finds the amount (and optionally the
// currency) in each row.
type csvOptions struct {
	// Column selects the amount column: a header name when Header is set,
	// otherwise (or if no header matches) a 1-based index.
	Column string
	// CurrencyColumn, if non-empty, names the column holding a per-row
	// currency code, using the same rules as Column. Currency is used when
	// it is empty.
	CurrencyColumn string
	Currency       string
	// Header says the first row is a header row. It is copied through with
	// a "converted" column appended.
	Header bool
	From   string
	Mode   RoundMode
}

// resolveColumn turns a column spec into a zero-based index. Header names
// win over numbers so a column literally called "2" is still reachable.
func resolveColumn(spec string, first []string, hasHeader bool) (int, error) {
	if hasHeader {
		for i, name := range first {
			if name == spec {
				return i, nil
			}
		}
	}
	n, err := strconv.Atoi(spec)
	if err != nil || n < 1 {
		if hasHeader {
			return 0, fmt.Errorf("no column named %q in header", spec)
		}
		return 0, fmt.Errorf("invalid column %q: want a 1-based column number", spec)
	}
	if n > len(first) {
		return 0, fmt.Errorf("column %d is out of range, the first row has %d columns", n, len(first))
	}
	return n - 1, nil
}

// convertCSV copies CSV rows from in to out with the converted amount
// appended as a new last column. A row that cannot be converted is reported
// on errw and still written, with an empty converted cell, so the output
// keeps one row per input row and stays aligned with the source file. The
// returned count is the number of such rows. The rows are read and written
// with encoding/csv, so quoted fields and embedded commas survive.
func convertCSV(in io.Reader, out, errw io.Writer, opts csvOptions) (int, error) {
	cr := csv.NewReader(in)
	cr.FieldsPerRecord = -1 // ragged rows are reported per row, not fatal
	cw := csv.NewWriter(out)

	amountIdx, currIdx := -1, -1
	failures := 0
	row := 0

	for {
		rec, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			cw.Flush()
			return failures, err
		}
		row++

		if row == 1 {
			if amountIdx, err = resolveColumn(opts.Column, rec, opts.Header); err != nil {
				return failures, err
			}
			if opts.CurrencyColumn != "" {
				if currIdx, err = resolveColumn(opts.CurrencyColumn, rec, opts.Header); err != nil {
					return failures, err
				}
			}
			if opts.Header {
				if err := cw.Write(append(append([]string(nil), rec...), "converted")); err != nil {
					return failures, err
				}
				continue
			}
		}

		converted, err := convertCSVRow(rec, amountIdx, currIdx, opts)
		if err != nil {
			fmt.Fprintf(errw, "row %d: error: %v\n", row, err)
			failures++
		}
		if err := cw.Write(append(append([]string(nil), rec...), converted)); err != nil {
			return failures, err
		}
	}

	cw.Flush()
	return failures, cw.Error()
}

func convertCSVRow(rec []string, amountIdx, currIdx int, opts csvOptions) (string, error) {
	if amountIdx >= len(rec) {
		return "", fmt.Errorf("row has %d columns, amount column is %d", len(rec), amountIdx+1)
	}
	currency := opts.Currency
	if currIdx >= 0 {
		if currIdx >= len(rec) {
			return "", fmt.Errorf("row has %d columns, currency column is %d", len(rec), currIdx+1)
		}
		currency = strings.TrimSpace(rec[currIdx])
	}
	exp, err := exponentFor(currency)
	if err != nil {
		return "", err
	}
	res, err := convertOne(strings.TrimSpace(rec[amountIdx]), currency, opts.From, exp, opts.Mode)
	if err != nil {
		return "", err
	}
	return res.Output, nil
}
