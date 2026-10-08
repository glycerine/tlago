package tlc

import (
	"math"
	"strconv"
	"strings"
)

// Worker statistics use two decimal places with decimal half-up rounding.
// Round the shortest decimal representation, not the binary value multiplied
// by 100; that multiplication can move decimal ties below the rounding point.
func distributedWorkerCacheRatio(ratio float64) string {
	mpNumberSymbolsOnce.Do(initializeMessageNumberSymbols)
	return distributedWorkerCacheRatioForLocale(ratio, mpNumberSymbols, mpNumberLocaleKey)
}

func distributedWorkerCacheRatioForLocale(ratio float64, symbols mpDecimalSymbols, key string) string {
	decimal := distributedCacheDecimalSeparator(symbols, key)
	// Formatter takes its grouping size from the locale's number pattern.
	// The POSIX pattern has zero grouping; MP's explicit pattern still groups.
	if key == "en-US-POSIX" {
		symbols.group = ""
	}
	return distributedWorkerCacheRatioWithSymbols(ratio, symbols, decimal)
}

func distributedWorkerCacheRatioWithSymbols(ratio float64, symbols mpDecimalSymbols, decimal string) string {
	if ratio < 0 {
		return "n/a"
	}
	if math.IsNaN(ratio) {
		return "NaN"
	}
	if math.IsInf(ratio, 1) {
		return "Infinity"
	}
	text := strconv.FormatFloat(ratio, 'f', -1, 64)
	parts := strings.SplitN(text, ".", 2)
	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
	}
	fraction += "00"
	digits := []byte(parts[0] + fraction[:2])
	if len(fraction) > 2 && fraction[2] >= '5' {
		digits = incrementDecimalBytes(digits)
	}
	text = string(digits[:len(digits)-2]) + "." + string(digits[len(digits)-2:])
	var output strings.Builder
	if strings.HasPrefix(text, "-") {
		output.WriteByte('-')
		text = text[1:]
	}
	integerLength := len(text) - 3
	for i, char := range text {
		if char == '.' {
			output.WriteString(decimal)
			continue
		}
		if i > 0 && i < integerLength && (integerLength-i)%3 == 0 {
			output.WriteString(symbols.group)
		}
		output.WriteRune(symbols.zero + char - '0')
	}
	return output.String()
}

// The existing console locale table supplies the zero digit and grouping.
// These separator pairs and the ambiguous-space/apostrophe exceptions are
// verified against the same OpenJDK locale data used by that table.
func distributedCacheDecimalSeparator(symbols mpDecimalSymbols, locale string) string {
	switch symbols.group {
	case ".", "\u202f":
		return ","
	case "\u066c":
		return "\u066b"
	case "\u00a0":
		language := strings.SplitN(locale, "-", 2)[0]
		switch language {
		case "dje", "khq", "mfe", "ses", "twq", "xh":
			return "."
		}
		return ","
	case "\u2019":
		if strings.SplitN(locale, "-", 2)[0] == "wae" {
			return ","
		}
	}
	return "."
}
