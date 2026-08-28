package main

import (
	"fmt"
	"strings"
)

// currencyExponents maps an ISO 4217 code to the number of digits after
// the decimal point in its "major" representation. Most currencies use 2,
// a handful use 0 or 3. This is not the full ISO 4217 list yet, just the
// currencies I've actually needed so far.
var currencyExponents = map[string]int{
	"USD": 2, "EUR": 2, "GBP": 2, "CHF": 2, "CAD": 2, "AUD": 2, "NZD": 2,
	"CNY": 2, "INR": 2, "MXN": 2, "BRL": 2, "ZAR": 2, "SGD": 2, "HKD": 2,
	"SEK": 2, "NOK": 2, "DKK": 2, "PLN": 2, "UYU": 2,
	"JPY": 0, "KRW": 0, "VND": 0, "ISK": 0,
	"BHD": 3, "KWD": 3, "OMR": 3, "JOD": 3, "TND": 3,
}

func exponentFor(currency string) (int, error) {
	exp, ok := currencyExponents[strings.ToUpper(currency)]
	if !ok {
		return 0, fmt.Errorf("unknown currency code %q (not in currencyExponents yet)", currency)
	}
	return exp, nil
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func splitSign(s string) (negative bool, rest string) {
	switch {
	case strings.HasPrefix(s, "-"):
		return true, s[1:]
	case strings.HasPrefix(s, "+"):
		return false, s[1:]
	default:
		return false, s
	}
}

// majorToMinor converts a decimal amount like "1234.56" into the integer
// number of minor units, e.g. 123456 for a currency with exponent 2.
// It refuses to round: an amount with more fractional digits than the
// currency allows is an error rather than a silently truncated value.
func majorToMinor(amount string, exponent int) (int64, error) {
	neg, s := splitSign(amount)

	parts := strings.SplitN(s, ".", 2)
	whole := parts[0]
	frac := ""
	if len(parts) == 2 {
		frac = parts[1]
	}
	if whole == "" || !isDigits(whole) {
		return 0, fmt.Errorf("invalid amount %q", amount)
	}
	if frac != "" && !isDigits(frac) {
		return 0, fmt.Errorf("invalid amount %q", amount)
	}
	if len(frac) > exponent {
		return 0, fmt.Errorf("amount %q has more fractional digits than the currency allows (%d)", amount, exponent)
	}
	frac += strings.Repeat("0", exponent-len(frac))

	var value int64
	for _, r := range whole + frac {
		value = value*10 + int64(r-'0')
	}
	if neg {
		value = -value
	}
	return value, nil
}

// minorToMajor converts an integer number of minor units back into a
// decimal string, e.g. 123456 with exponent 2 becomes "1234.56".
func minorToMajor(minor int64, exponent int) string {
	neg := minor < 0
	if neg {
		minor = -minor
	}
	digits := fmt.Sprintf("%d", minor)
	if len(digits) <= exponent {
		digits = strings.Repeat("0", exponent-len(digits)+1) + digits
	}

	out := digits
	if exponent > 0 {
		cut := len(digits) - exponent
		out = digits[:cut] + "." + digits[cut:]
	}
	if neg {
		out = "-" + out
	}
	return out
}

// parseMinor validates that a string is a plain signed integer, since
// minor-unit amounts should never carry a decimal point.
func parseMinor(s string) (int64, error) {
	neg, rest := splitSign(s)
	if !isDigits(rest) {
		return 0, fmt.Errorf("invalid minor amount %q", s)
	}
	var value int64
	for _, r := range rest {
		value = value*10 + int64(r-'0')
	}
	if neg {
		value = -value
	}
	return value, nil
}
