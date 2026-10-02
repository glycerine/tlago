package tlc

import (
	"errors"
	"math"
	"regexp"
	"strconv"
	"strings"
)

var javaDoubleLiteral = regexp.MustCompile(`^[+-]?(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)(?:[eE][+-]?[0-9]+)?[fFdD]?$|^[+-]?0[xX](?:[0-9a-fA-F]+(?:\.[0-9a-fA-F]*)?|\.[0-9a-fA-F]+)[pP][+-]?[0-9]+[fFdD]?$`)

// Double.parseDouble trims ASCII control/space characters, accepts Java
// decimal/hex literals and suffixes, and returns infinities on overflow.
func parseJavaDoubleProperty(value string) (float64, error) {
	value = strings.TrimFunc(value, func(char rune) bool { return char <= ' ' })
	switch value {
	case "NaN", "+NaN", "-NaN":
		return math.NaN(), nil
	case "Infinity", "+Infinity":
		return math.Inf(1), nil
	case "-Infinity":
		return math.Inf(-1), nil
	}
	if !javaDoubleLiteral.MatchString(value) {
		return 0, &strconv.NumError{Func: "ParseFloat", Num: value, Err: strconv.ErrSyntax}
	}
	if strings.ContainsAny(value[len(value)-1:], "dDfF") {
		value = value[:len(value)-1]
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if errors.Is(err, strconv.ErrRange) {
		err = nil
	}
	return parsed, err
}
