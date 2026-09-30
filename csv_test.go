package main

import (
	"bytes"
	"strings"
	"testing"
)

func runCSV(t *testing.T, input string, opts csvOptions) (string, string, int, error) {
	t.Helper()
	var out, errOut bytes.Buffer
	n, err := convertCSV(strings.NewReader(input), &out, &errOut, opts)
	return out.String(), errOut.String(), n, err
}

func TestConvertCSVHeaderByName(t *testing.T) {
	in := "id,amount\na,1234.56\nb,0.07\n"
	out, _, n, err := runCSV(t, in, csvOptions{Column: "amount", Currency: "USD", From: "major", Mode: RoundError, Header: true})
	if err != nil || n != 0 {
		t.Fatalf("err = %v, failures = %d", err, n)
	}
	want := "id,amount,converted\na,1234.56,123456\nb,0.07,7\n"
	if out != want {
		t.Errorf("output = %q, want %q", out, want)
	}
}

func TestConvertCSVNoHeaderByIndex(t *testing.T) {
	out, _, _, err := runCSV(t, "x,5000\n", csvOptions{Column: "2", Currency: "JPY", From: "minor", Mode: RoundError})
	if err != nil {
		t.Fatal(err)
	}
	if out != "x,5000,5000\n" {
		t.Errorf("output = %q", out)
	}
}

func TestConvertCSVPerRowCurrency(t *testing.T) {
	in := "amount,cur\n1.500,BHD\n5000,jpy\n12.34,USD\n"
	out, _, n, err := runCSV(t, in, csvOptions{Column: "amount", CurrencyColumn: "cur", From: "major", Mode: RoundError, Header: true})
	if err != nil || n != 0 {
		t.Fatalf("err = %v, failures = %d", err, n)
	}
	want := "amount,cur,converted\n1.500,BHD,1500\n5000,jpy,5000\n12.34,USD,1234\n"
	if out != want {
		t.Errorf("output = %q, want %q", out, want)
	}
}

func TestConvertCSVQuotedFieldsSurvive(t *testing.T) {
	in := "\"Smith, J\",10.00\n"
	out, _, _, err := runCSV(t, in, csvOptions{Column: "2", Currency: "USD", From: "major", Mode: RoundError})
	if err != nil {
		t.Fatal(err)
	}
	if out != "\"Smith, J\",10.00,1000\n" {
		t.Errorf("output = %q", out)
	}
}

func TestConvertCSVBadRowKeepsAlignment(t *testing.T) {
	in := "1.00\nabc\n2.50\n"
	out, errOut, n, err := runCSV(t, in, csvOptions{Column: "1", Currency: "USD", From: "major", Mode: RoundError})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("failures = %d, want 1", n)
	}
	if out != "1.00,100\nabc,\n2.50,250\n" {
		t.Errorf("output = %q", out)
	}
	if !strings.Contains(errOut, "row 2:") {
		t.Errorf("stderr = %q, want a row 2 message", errOut)
	}
}

func TestConvertCSVShortRow(t *testing.T) {
	out, _, n, err := runCSV(t, "a,1.00\nb\n", csvOptions{Column: "2", Currency: "USD", From: "major", Mode: RoundError})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 || out != "a,1.00,100\nb,\n" {
		t.Errorf("failures = %d, output = %q", n, out)
	}
}

func TestConvertCSVUnknownHeaderColumn(t *testing.T) {
	_, _, _, err := runCSV(t, "a,b\n1,2\n", csvOptions{Column: "amount", Currency: "USD", From: "major", Mode: RoundError, Header: true})
	if err == nil {
		t.Error("expected error for a column name missing from the header")
	}
}

func TestConvertCSVColumnOutOfRange(t *testing.T) {
	_, _, _, err := runCSV(t, "1.00\n", csvOptions{Column: "3", Currency: "USD", From: "major", Mode: RoundError})
	if err == nil {
		t.Error("expected error for a column number past the end of the first row")
	}
}

func TestConvertCSVRounding(t *testing.T) {
	out, _, n, err := runCSV(t, "1.235\n", csvOptions{Column: "1", Currency: "USD", From: "major", Mode: RoundHalfUp})
	if err != nil || n != 0 {
		t.Fatalf("err = %v, failures = %d", err, n)
	}
	if out != "1.235,124\n" {
		t.Errorf("output = %q", out)
	}
}
