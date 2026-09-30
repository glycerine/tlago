package tlc

import (
	"math"
	"strconv"
	"strings"
)

func CalculateOptimisticProbability(numDistinct uint64, numGenerated int64) float64 {
	return float64(numDistinct) * ((float64(numGenerated) - float64(numDistinct)) / math.Pow(2, 64))
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
	actual := "val = " + probabilityToString(1/float64(actualDistance), 2)
	PrintMessage(ECTLCSuccess, optimistic, actual)
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
