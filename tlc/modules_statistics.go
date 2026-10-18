package tlc

import (
	"math"
	"strings"
)

const (
	statisticsGammaEpsilon       = 1e-14
	statisticsGammaMaxIterations = 100000
	statisticsGammaTiny          = 1e-300
	statisticsChiSquareRescale   = 1e-5
)

func StatisticsChiSquare(expected Value, actual Value, alpha Value) (Value, error) {
	if isNil(expected) {
		return nil, NewNullPointerException(`Cannot invoke "tlc2.value.impl.Value.normalize()" because "<parameter1>" is null`)
	}
	expFcn := asFcnRcdValue(expected.Normalize())
	if expFcn == nil {
		return nil, NewNullPointerException(`Cannot invoke "tlc2.value.impl.Value.toString()" because "<local0>" is null`)
	}
	if isNil(actual) {
		return nil, NewNullPointerException(`Cannot invoke "tlc2.value.impl.Value.normalize()" because "<parameter2>" is null`)
	}
	actFcn := asFcnRcdValue(actual.Normalize())
	if actFcn == nil {
		return nil, NewNullPointerException(`Cannot invoke "tlc2.value.impl.Value.toString()" because "<local1>" is null`)
	}
	if isNil(alpha) {
		return nil, NewNullPointerException(`Cannot invoke "tlc2.value.impl.Value.toString()" because "<parameter3>" is null`)
	}
	alphaString, ok := asStringValue(alpha)
	if !ok {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "third", "ChiSquare", "StringValue representing a double in (0, .5]", ValuesPPR(alpha))
	}
	exp, err := statisticsIntValuesAsExpected(expFcn.ownedValues())
	if err != nil {
		return nil, err
	}
	act, err := statisticsIntValuesAsObserved(actFcn.ownedValues())
	if err != nil {
		return nil, err
	}
	if alphaString.ownedVal() == nil {
		return nil, NewNullPointerException(`Cannot invoke "util.UniqueString.toString()" because "<parameter3>.val" is null`)
	}
	a, err := statisticsParseJavaDouble(alphaString.RawString())
	if err != nil {
		return nil, err
	}
	if !math.IsNaN(a) && (a <= 0 || a > 0.5) {
		return nil, newStatisticsMathException("org.apache.commons.math3.exception.OutOfRangeException", "out of bounds significance level "+statisticsSignificanceNumber(a)+", must be between 0 and 0.5")
	}
	p, err := statisticsChiSquarePValue(exp, act)
	if err != nil {
		return nil, err
	}
	return BoolValueOf(math.IsNaN(p) || math.IsNaN(a) || p >= a), nil
}

func statisticsParseJavaDouble(value string) (float64, error) {
	trimmed := strings.TrimFunc(value, func(char rune) bool { return char <= ' ' })
	parsed, err := parseJavaDoubleProperty(value)
	if err == nil {
		return parsed, nil
	}
	if trimmed == "" {
		return 0, NewNumberFormatExceptionNullable(javaString("empty String"))
	}
	return 0, NewNumberFormatExceptionForInputString(trimmed, 10)
}

func statisticsIntValuesAsExpected(values []Value) ([]float64, error) {
	if values == nil {
		return nil, NewNullPointerException()
	}
	out := make([]float64, len(values))
	for i, value := range values {
		if isNil(value) {
			return nil, NewNullPointerException(`Cannot read field "val" because "<parameter1>" is null`)
		}
		intValue, ok := value.(*IntValue)
		if !ok {
			return nil, valueStreamClassCast(value, "tlc2.value.impl.IntValue")
		}
		out[i] = float64(intValue.ownedVal())
	}
	return out, nil
}

func statisticsIntValuesAsObserved(values []Value) ([]int64, error) {
	if values == nil {
		return nil, NewNullPointerException()
	}
	out := make([]int64, len(values))
	for i, value := range values {
		if isNil(value) {
			return nil, NewNullPointerException(`Cannot read field "val" because "<parameter1>" is null`)
		}
		intValue, ok := value.(*IntValue)
		if !ok {
			return nil, valueStreamClassCast(value, "tlc2.value.impl.IntValue")
		}
		out[i] = int64(intValue.ownedVal())
	}
	return out, nil
}

func statisticsChiSquarePValue(expected []float64, observed []int64) (float64, error) {
	// ChiSquareTest constructs the distribution before chiSquare validates the
	// samples. Its constructor rejects zero and negative degrees of freedom.
	if len(expected) < 2 {
		return 0, newStatisticsMathException("org.apache.commons.math3.exception.NotStrictlyPositiveException", "shape ("+statisticsMathNumber(float64(len(expected)-1))+")")
	}
	statistic, err := statisticsChiSquareStatistic(expected, observed)
	if err != nil {
		return 0, err
	}
	degreesOfFreedom := float64(len(expected) - 1)
	return statisticsRegularizedGammaQ(degreesOfFreedom/2, statistic/2), nil
}

func statisticsChiSquareStatistic(expected []float64, observed []int64) (float64, error) {
	if len(expected) != len(observed) {
		return 0, newStatisticsMathException("org.apache.commons.math3.exception.DimensionMismatchException", statisticsMathNumber(float64(len(expected)))+" != "+statisticsMathNumber(float64(len(observed))))
	}
	var sumExpected float64
	var sumObserved int64
	for _, value := range expected {
		if value <= 0 {
			return 0, newStatisticsMathException("org.apache.commons.math3.exception.NotStrictlyPositiveException", statisticsMathNumber(value)+" is smaller than, or equal to, the minimum (0)")
		}
		sumExpected += value
	}
	for _, value := range observed {
		if value < 0 {
			return 0, newStatisticsMathException("org.apache.commons.math3.exception.NotPositiveException", statisticsMathNumber(float64(value))+" is smaller than the minimum (0)")
		}
		sumObserved += value
	}

	ratio := 1.0
	rescale := math.Abs(sumExpected-float64(sumObserved)) > statisticsChiSquareRescale
	if rescale {
		ratio = float64(sumObserved) / sumExpected
	}

	var sum float64
	for i, exp := range expected {
		if rescale {
			exp *= ratio
		}
		dev := float64(observed[i]) - exp
		sum += dev * dev / exp
	}
	return sum, nil
}

type statisticsMathException struct {
	*IllegalArgumentException
	class string
}

func newStatisticsMathException(class, message string) *statisticsMathException {
	return &statisticsMathException{IllegalArgumentException: NewIllegalArgumentException(message), class: class}
}

func (e *statisticsMathException) diagnosticClassName() string { return e.class }

// Commons Math's getMessage uses MessageFormat with Locale.US, whose default
// number format groups thousands and uses at most three fractional digits.
func statisticsMathNumber(value float64) string {
	return compactCacheRatio(value,
		mpDecimalSymbols{zero: '0', group: ",", negativePrefix: "-"},
		cacheDecimalSymbols{decimal: ".", infinity: "∞", nan: "NaN"})
}

func statisticsSignificanceNumber(value float64) string {
	return statisticsMathNumber(value)
}

func statisticsRegularizedGammaQ(a float64, x float64) float64 {
	if math.IsNaN(a) || math.IsNaN(x) || a <= 0 || x < 0 {
		return math.NaN()
	}
	if x == 0 {
		return 1
	}
	if x < a+1 {
		return 1 - statisticsRegularizedGammaPSeries(a, x)
	}
	return statisticsRegularizedGammaQContinuedFraction(a, x)
}

func statisticsRegularizedGammaPSeries(a float64, x float64) float64 {
	lgamma, _ := math.Lgamma(a)
	ap := a
	sum := 1.0 / a
	del := sum
	for n := 1; n <= statisticsGammaMaxIterations; n++ {
		ap++
		del *= x / ap
		sum += del
		if math.Abs(del) < math.Abs(sum)*statisticsGammaEpsilon {
			break
		}
	}
	return sum * math.Exp(-x+a*math.Log(x)-lgamma)
}

func statisticsRegularizedGammaQContinuedFraction(a float64, x float64) float64 {
	lgamma, _ := math.Lgamma(a)
	b := x + 1 - a
	c := 1.0 / statisticsGammaTiny
	d := b
	if math.Abs(d) < statisticsGammaTiny {
		d = statisticsGammaTiny
	}
	d = 1.0 / d
	h := d
	for i := 1; i <= statisticsGammaMaxIterations; i++ {
		fi := float64(i)
		an := -fi * (fi - a)
		b += 2
		d = an*d + b
		if math.Abs(d) < statisticsGammaTiny {
			d = statisticsGammaTiny
		}
		c = b + an/c
		if math.Abs(c) < statisticsGammaTiny {
			c = statisticsGammaTiny
		}
		d = 1.0 / d
		del := d * c
		h *= del
		if math.Abs(del-1) < statisticsGammaEpsilon {
			break
		}
	}
	return math.Exp(-x+a*math.Log(x)-lgamma) * h
}
