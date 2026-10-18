// OpenJDK-derived numeric formatting rule; see licenses/openjdk-LICENSE and
// licenses/openjdk-ADDITIONAL_LICENSE_INFO for retained notices.
package tlc

import (
	"math"
	"strconv"
	"strings"
)

type cacheDecimalSymbols struct{ decimal, infinity, nan string }

var cacheNumberSpecialSymbols cacheDecimalSymbols

func selectCacheDecimalSymbols(parts []string, extensions string) cacheDecimalSymbols {
	if _, variant, ok := strings.Cut(extensions, "x-lvariant-"); ok && !(len(parts) >= 3 && parts[len(parts)-1] == variant) {
		parts = append(append([]string{}, parts...), variant)
	}
	key := "und"
	set := cacheDecimalSymbolSets[cacheDecimalSymbolLocales[key]]
	for len(parts) > 0 {
		if index, ok := cacheDecimalSymbolLocales[strings.Join(parts, "-")]; ok {
			key, set = strings.Join(parts, "-"), cacheDecimalSymbolSets[index]
			break
		}
		parts = parts[:len(parts)-1]
	}
	symbols := set["default"]
	numbering := mpNumberingSystem(extensions)
	// The legacy Thai locale's own compatibility keyword does not normalize
	// away its variant, unlike an explicit numbering-system replacement.
	if key == "th-TH-TH" && numbering == "thai" {
		return symbols
	}
	if selected, ok := set[numbering]; ok {
		symbols = selected
	}
	return symbols
}

func compactCacheRatio(ratio float64, symbols mpDecimalSymbols, special cacheDecimalSymbols) string {
	if math.IsNaN(ratio) {
		return special.nan
	}
	negative := math.Signbit(ratio)
	text := special.infinity
	if !math.IsInf(ratio, 0) {
		ratio = math.Abs(ratio)
		text = cacheRatioDecimalDigits(ratio)
		integer, fraction, _ := strings.Cut(text, ".")
		var out strings.Builder
		for i, char := range integer {
			if i > 0 && (len(integer)-i)%3 == 0 {
				out.WriteString(symbols.group)
			}
			out.WriteRune(symbols.zero + char - '0')
		}
		if fraction != "" {
			out.WriteString(special.decimal)
			for _, char := range fraction {
				out.WriteRune(symbols.zero + char - '0')
			}
		}
		text = out.String()
	}
	if negative {
		return symbols.negativePrefix + text + symbols.negativeSuffix
	}
	return text
}

// DecimalFormat's integral conversion discards decimal places implied by the
// binary exponent before generating digits (FloatingDecimal.developLongDigits).
// This is a numeric formatting rule, implemented with native integer arithmetic.
func cacheRatioDecimalDigits(ratio float64) string {
	if ratio == math.Trunc(ratio) && ratio < 9223372036854775808.0 {
		value := uint64(ratio)
		exponent := int((math.Float64bits(ratio)>>52)&2047) - 1023
		if exponent > 53 {
			uncertain := uint64(1) << uint(exponent-54)
			place := uint64(1)
			for uncertain >= 10 {
				uncertain /= 10
				place *= 10
			}
			residue := value % place
			value /= place
			if place > 1 && residue >= place/2 {
				value++
			}
			value *= place
		}
		return strconv.FormatUint(value, 10)
	}
	shortest := strconv.FormatFloat(ratio, 'f', -1, 64)
	_, fraction, _ := strings.Cut(shortest, ".")
	if len(fraction) <= 3 {
		return shortest
	}
	return strings.TrimRight(strings.TrimRight(strconv.FormatFloat(ratio, 'f', 3, 64), "0"), ".")
}
