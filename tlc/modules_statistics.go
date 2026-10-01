package tlc

import (
	"math"
	"strconv"
)

const (
	statisticsGammaEpsilon       = 1e-14
	statisticsGammaMaxIterations = 100000
	statisticsGammaTiny          = 1e-300
	statisticsChiSquareRescale   = 1e-5
)

func StatisticsChiSquare(expected Value, actual Value, alpha Value) (Value, error) {
	expFcn := asFcnRcdValue(expected.Normalize())
	if expFcn == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "ChiSquare", "function", ValuesPPR(expected))
	}
	actFcn := asFcnRcdValue(actual.Normalize())
	if actFcn == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "ChiSquare", "function", ValuesPPR(actual))
	}
	alphaString, ok := alpha.(*StringValue)
	if !ok {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "third", "ChiSquare", "StringValue representing a double in (0, .5]", ValuesPPR(alpha))
	}
	a, err := strconv.ParseFloat(alphaString.RawString(), 64)
	if err != nil || a <= 0 || a > 0.5 {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "third", "ChiSquare", "StringValue representing a double in (0, .5]", ValuesPPR(alpha))
	}

	exp, err := statisticsIntValuesAsExpected("first", expFcn.Values)
	if err != nil {
		return nil, err
	}
	act, err := statisticsIntValuesAsObserved("second", actFcn.Values)
	if err != nil {
		return nil, err
	}
	p, err := statisticsChiSquarePValue(exp, act)
	if err != nil {
		return nil, err
	}
	return NewBoolValue(p >= a), nil
}

func statisticsIntValuesAsExpected(position string, values []Value) ([]float64, error) {
	out := make([]float64, len(values))
	for i, value := range values {
		intValue, ok := value.(*IntValue)
		if !ok {
			return nil, newTLCErrorCode(ECTLCModuleArgumentError, position, "ChiSquare", "function with integer values", ValuesPPR(value))
		}
		out[i] = float64(intValue.Val)
	}
	return out, nil
}

func statisticsIntValuesAsObserved(position string, values []Value) ([]int64, error) {
	out := make([]int64, len(values))
	for i, value := range values {
		intValue, ok := value.(*IntValue)
		if !ok {
			return nil, newTLCErrorCode(ECTLCModuleArgumentError, position, "ChiSquare", "function with integer values", ValuesPPR(value))
		}
		out[i] = int64(intValue.Val)
	}
	return out, nil
}

func statisticsChiSquarePValue(expected []float64, observed []int64) (float64, error) {
	statistic, err := statisticsChiSquareStatistic(expected, observed)
	if err != nil {
		return 0, err
	}
	degreesOfFreedom := float64(len(expected) - 1)
	return statisticsRegularizedGammaQ(degreesOfFreedom/2, statistic/2), nil
}

func statisticsChiSquareStatistic(expected []float64, observed []int64) (float64, error) {
	if len(expected) < 2 {
		return 0, newTLCError(ECGeneral, "ChiSquare requires at least two bins")
	}
	if len(expected) != len(observed) {
		return 0, newTLCError(ECGeneral, "ChiSquare expected and actual functions have different sizes")
	}
	var sumExpected float64
	var sumObserved int64
	for _, value := range expected {
		if value <= 0 || math.IsNaN(value) {
			return 0, newTLCError(ECGeneral, "ChiSquare expected counts must be positive")
		}
		sumExpected += value
	}
	for _, value := range observed {
		if value < 0 {
			return 0, newTLCError(ECGeneral, "ChiSquare actual counts must be non-negative")
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
