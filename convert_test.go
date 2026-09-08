package main

import "testing"

func TestMajorToMinor(t *testing.T) {
	cases := []struct {
		amount   string
		exponent int
		want     int64
		wantErr  bool
	}{
		{"1234.56", 2, 123456, false},
		{"1234.5", 2, 123450, false},
		{"1234", 2, 123400, false},
		{"-1234.56", 2, -123456, false},
		{"0.07", 2, 7, false},
		{"1234", 0, 1234, false},
		{"1.234", 3, 1234, false},
		{"1.2345", 3, 0, true},  // too many fractional digits
		{"12a4.56", 2, 0, true}, // not a number
		{"", 2, 0, true},
	}

	for _, c := range cases {
		got, _, err := majorToMinor(c.amount, c.exponent, RoundError)
		if c.wantErr {
			if err == nil {
				t.Errorf("majorToMinor(%q, %d) = %d, want error", c.amount, c.exponent, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("majorToMinor(%q, %d) unexpected error: %v", c.amount, c.exponent, err)
			continue
		}
		if got != c.want {
			t.Errorf("majorToMinor(%q, %d) = %d, want %d", c.amount, c.exponent, got, c.want)
		}
	}
}

func TestMajorToMinorRounding(t *testing.T) {
	cases := []struct {
		amount      string
		exponent    int
		mode        RoundMode
		want        int64
		wantRounded bool
	}{
		{"1.234", 2, RoundDown, 123, true},
		{"-1.234", 2, RoundDown, -123, true},
		{"1.230", 2, RoundDown, 123, false}, // extra digit is zero, but still present
		{"1.235", 2, RoundUp, 124, true},
		{"1.230", 2, RoundUp, 123, true},    // extra digits are all zero: no bump needed
		{"1.20", 2, RoundUp, 120, false},    // no extra digits at all
		{"-1.231", 2, RoundUp, -124, true},
		{"1.005", 2, RoundHalfUp, 101, true},
		{"1.004", 2, RoundHalfUp, 100, true},
		{"-1.005", 2, RoundHalfUp, -101, true},
		{"1.005", 2, RoundError, 0, false}, // no digits dropped; caller should have gotten an error
	}

	for _, c := range cases {
		got, rounded, err := majorToMinor(c.amount, c.exponent, c.mode)
		if c.mode == RoundError {
			if err == nil {
				t.Errorf("majorToMinor(%q, %d, %q) = %d, want error", c.amount, c.exponent, c.mode, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("majorToMinor(%q, %d, %q) unexpected error: %v", c.amount, c.exponent, c.mode, err)
			continue
		}
		if got != c.want {
			t.Errorf("majorToMinor(%q, %d, %q) = %d, want %d", c.amount, c.exponent, c.mode, got, c.want)
		}
		if rounded != c.wantRounded {
			t.Errorf("majorToMinor(%q, %d, %q) rounded = %v, want %v", c.amount, c.exponent, c.mode, rounded, c.wantRounded)
		}
	}
}

func TestParseRoundMode(t *testing.T) {
	for _, m := range []string{"error", "down", "up", "half-up"} {
		if _, err := parseRoundMode(m); err != nil {
			t.Errorf("parseRoundMode(%q) unexpected error: %v", m, err)
		}
	}
	if _, err := parseRoundMode("nearest"); err == nil {
		t.Error("parseRoundMode(\"nearest\") expected error, got nil")
	}
}

func TestMinorToMajor(t *testing.T) {
	cases := []struct {
		minor    int64
		exponent int
		want     string
	}{
		{123456, 2, "1234.56"},
		{7, 2, "0.07"},
		{-123456, 2, "-1234.56"},
		{1234, 0, "1234"},
		{1234, 3, "1.234"},
		{5, 3, "0.005"},
	}

	for _, c := range cases {
		got := minorToMajor(c.minor, c.exponent)
		if got != c.want {
			t.Errorf("minorToMajor(%d, %d) = %q, want %q", c.minor, c.exponent, got, c.want)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	amounts := []string{"1234.56", "0.01", "999999.99", "-42.00"}
	for _, a := range amounts {
		minor, err := majorToMinor(a, 2)
		if err != nil {
			t.Fatalf("majorToMinor(%q) error: %v", a, err)
		}
		back := minorToMajor(minor, 2)
		if back != a {
			t.Errorf("round trip for %q produced %q", a, back)
		}
	}
}

func TestExponentForUnknownCurrency(t *testing.T) {
	if _, err := exponentFor("XYZ"); err == nil {
		t.Error("expected error for unknown currency code, got nil")
	}
}

func TestExponentForKnownCurrencies(t *testing.T) {
	cases := []struct {
		code string
		want int
	}{
		{"usd", 2}, // lowercase input should still resolve
		{"CLP", 0}, // zero-decimal
		{"IQD", 3}, // three-decimal, added alongside the Gulf dinars
		{"CLF", 4}, // four-decimal
	}
	for _, c := range cases {
		got, err := exponentFor(c.code)
		if err != nil {
			t.Errorf("exponentFor(%q) unexpected error: %v", c.code, err)
			continue
		}
		if got != c.want {
			t.Errorf("exponentFor(%q) = %d, want %d", c.code, got, c.want)
		}
	}
}
