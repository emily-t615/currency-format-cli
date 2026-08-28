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
		got, err := majorToMinor(c.amount, c.exponent)
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
