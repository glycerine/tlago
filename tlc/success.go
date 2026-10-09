package tlc

import (
	"math"
	"math/big"
	"strconv"
	"strings"
)

func CalculateOptimisticProbability(numDistinct uint64, numGenerated int64) float64 {
	distinct := int64(numDistinct)
	return float64(distinct) * (float64(numGenerated-distinct) / math.Pow(2, 64))
}

func ReportSuccess(fpSet FPSet, numGenerated int64) {
	if fpSet == nil {
		ReportSuccessCounts(0, numGenerated)
		return
	}
	numDistinct := fpSet.Size()
	if CalculateOptimisticProbability(numDistinct, numGenerated) < 1e-10 {
		ReportSuccessCounts(numDistinct, numGenerated)
		return
	}
	ReportSuccessCountsDistance(numDistinct, fpSet.CheckFPs(), numGenerated)
}

func ReportSuccessCounts(numDistinct uint64, numGenerated int64) {
	optimistic := CalculateOptimisticProbability(numDistinct, numGenerated)
	PrintMessage(ECTLCSuccess, "val = "+probabilityToString(optimistic, 2))
}

func ReportSuccessCountsDistance(numDistinct uint64, actualDistance uint64, numGenerated int64) {
	if numDistinct == uint64(numGenerated) && numGenerated == 0 {
		PrintMessage(ECTLCSuccess, "val = 0.0", "val = 0.0")
		return
	}
	optimistic := "val = " + probabilityToString(CalculateOptimisticProbability(numDistinct, numGenerated), 2)
	actual := "val = " + probabilityToString(fingerprintCollisionProbability(actualDistance), 2)
	PrintMessage(ECTLCSuccess, optimistic, actual)
}

// AbstractChecker divides 1 by the signed long distance with two significant
// decimal digits and half-up rounding before converting the result to double.
// Round the exact fraction here; binary division followed by display rounding
// loses signed failure sentinels and skips the source decimal rounding stage.
func fingerprintCollisionProbability(distanceBits uint64) float64 {
	distance := int64(distanceBits)
	if distance == 0 {
		panic(NewArithmeticException("Division by zero"))
	}
	magnitude := uint64(distance)
	if distance < 0 {
		magnitude = -magnitude // Also retains the magnitude of MinInt64.
	}
	// A signed-long magnitude is at most 2^63, so the next power of ten
	// fits uint64. Its multiple of ten needs the wider integer below.
	power := uint64(1)
	for power < magnitude {
		power *= 10
	}
	numerator := new(big.Int).SetUint64(power)
	numerator.Mul(numerator, big.NewInt(10))
	denominator := new(big.Int).SetUint64(magnitude)
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(numerator, denominator, remainder)
	comparison := remainder.Lsh(remainder, 1).Cmp(denominator)
	if comparison >= 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	if distance < 0 {
		quotient.Neg(quotient)
	}
	probability, _ := new(big.Rat).SetFrac(quotient, numerator).Float64()
	return probability
}

func probabilityToString(val float64, significantDigits int) string {
	if val == 0 {
		return "0.0"
	}
	valString := javaDoubleString(val)
	valStringLen := len(valString)
	result := make([]byte, 0, valStringLen)
	next := 0
	significantDigitsFound := 0
	for next < valStringLen && valString[next] == '0' {
		next++
	}
	for next < valStringLen && isASCIIDigit(valString[next]) {
		result = append(result, valString[next])
		significantDigitsFound++
		next++
	}
	if next == valStringLen {
		return string(result)
	}
	if valString[next] != '.' {
		return valString
	}
	if significantDigitsFound >= significantDigits {
		next++
		for next < valStringLen && isASCIIDigit(valString[next]) {
			next++
		}
	} else {
		next++
		result = append(result, '.')
		if significantDigitsFound == 0 {
			for next < valStringLen && valString[next] == '0' {
				next++
				result = append(result, '0')
			}
		}
		for next < valStringLen && isASCIIDigit(valString[next]) && significantDigitsFound < significantDigits {
			result = append(result, valString[next])
			next++
			significantDigitsFound++
		}
		if next < valStringLen && isASCIIDigit(valString[next]) && valString[next] >= '5' {
			result = incrementDecimalBytes(result)
		}
		for next < valStringLen && isASCIIDigit(valString[next]) {
			next++
		}
	}
	if next >= valStringLen {
		return string(result)
	}
	if valString[next] == 'E' {
		return string(result) + valString[next:]
	}
	return valString
}

func javaDoubleString(val float64) string {
	switch {
	case math.IsNaN(val):
		return "NaN"
	case math.IsInf(val, 1):
		return "Infinity"
	case math.IsInf(val, -1):
		return "-Infinity"
	}
	s := strconv.FormatFloat(val, 'g', -1, 64)
	if idx := strings.IndexAny(s, "eE"); idx != -1 {
		mantissa := s[:idx]
		exponent := s[idx+1:]
		if !strings.Contains(mantissa, ".") {
			mantissa += ".0"
		}
		sign := ""
		if strings.HasPrefix(exponent, "+") || strings.HasPrefix(exponent, "-") {
			sign = exponent[:1]
			exponent = exponent[1:]
		}
		exponent = strings.TrimLeft(exponent, "0")
		if exponent == "" {
			exponent = "0"
		}
		return mantissa + "E" + sign + exponent
	}
	if !strings.Contains(s, ".") {
		s += ".0"
	}
	return s
}

func incrementDecimalBytes(result []byte) []byte {
	for prev := len(result) - 1; ; prev-- {
		if prev < 0 {
			return append([]byte{'1'}, result...)
		}
		prevChar := result[prev]
		if !isASCIIDigit(prevChar) {
			continue
		}
		if prevChar == '9' {
			result[prev] = '0'
			continue
		}
		result[prev] = prevChar + 1
		return result
	}
}

func isASCIIDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}
