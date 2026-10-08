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
	return groupDecimalIntegerPart(text)
}
