package main

import "testing"

func TestConvertOneMajorToMinor(t *testing.T) {
	res, err := convertOne("1234.56", "usd", "major", 2, RoundError)
	if err != nil {
		t.Fatalf("convertOne unexpected error: %v", err)
	}
	if res.Currency != "USD" {
		t.Errorf("Currency = %q, want %q", res.Currency, "USD")
	}
	if res.InputFormat != "major" || res.OutputFormat != "minor" {
		t.Errorf("formats = %q/%q, want major/minor", res.InputFormat, res.OutputFormat)
	}
	if res.Output != "123456" {
		t.Errorf("Output = %q, want %q", res.Output, "123456")
	}
}

func TestConvertOneMinorToMajor(t *testing.T) {
	res, err := convertOne("123456", "USD", "minor", 2, RoundError)
	if err != nil {
		t.Fatalf("convertOne unexpected error: %v", err)
	}
	if res.InputFormat != "minor" || res.OutputFormat != "major" {
		t.Errorf("formats = %q/%q, want minor/major", res.InputFormat, res.OutputFormat)
	}
	if res.Output != "1234.56" {
		t.Errorf("Output = %q, want %q", res.Output, "1234.56")
	}
}

func TestConvertOneInvalidFrom(t *testing.T) {
	if _, err := convertOne("100", "USD", "sideways", 2, RoundError); err == nil {
		t.Error("expected error for invalid -from value, got nil")
	}
}

func TestConvertOnePropagatesAmountError(t *testing.T) {
	if _, err := convertOne("1.2345", "USD", "major", 2, RoundError); err == nil {
		t.Error("expected error for amount with too many fractional digits, got nil")
	}
}

func TestConvertOneRoundingFlag(t *testing.T) {
	res, err := convertOne("1.2345", "USD", "major", 2, RoundHalfUp)
	if err != nil {
		t.Fatalf("convertOne unexpected error: %v", err)
	}
	if res.Output != "123" {
		t.Errorf("Output = %q, want %q", res.Output, "123")
	}
	if !res.Rounded {
		t.Error("Rounded = false, want true")
	}
}

func TestConvertOneNotRoundedWhenExact(t *testing.T) {
	res, err := convertOne("1.23", "USD", "major", 2, RoundHalfUp)
	if err != nil {
		t.Fatalf("convertOne unexpected error: %v", err)
	}
	if res.Rounded {
		t.Error("Rounded = true, want false for an amount with no extra digits")
	}
}
