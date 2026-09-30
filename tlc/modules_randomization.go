package tlc

import (
	"math"
	"strconv"
)

func RandomSubset(k Value, set Value) (Value, error) {
	count, ok := k.(*IntValue)
	if !ok {
		return nil, newTLCError(ECGeneral, "first argument of RandomSubset must be a nonnegative integer, got %s", k)
	}
	return randomSubsetOfEnumerable(int(count.Val), set)
}

func RandomSetOfSubsets(numberOfPicks Value, subsetSize Value, set Value) (Value, error) {
	picks, ok := numberOfPicks.(*IntValue)
	if !ok || picks.Val < 0 {
		return nil, newTLCError(ECGeneral, "first argument of RandomSetOfSubsets must be a nonnegative integer, got %s", numberOfPicks)
	}
	n, ok := subsetSize.(*IntValue)
	if !ok || n.Val < 0 {
		return nil, newTLCError(ECGeneral, "second argument of RandomSetOfSubsets must be a nonnegative integer, got %s", subsetSize)
	}
	size, err := set.Size()
	if err != nil {
		return nil, err
	}
	if int(n.Val) > size {
		return nil, newTLCError(ECGeneral, "second argument of RandomSetOfSubsets must be in 0..Cardinality(S), got %s", subsetSize)
	}
	if err := checkRandomSubsetPickCount(int(picks.Val), size); err != nil {
		return nil, err
	}
	probability := float64(n.Val) / float64(size)
	return randomSetOfSubsets(int(picks.Val), probability, set)
}

func RandomSubsetSet(numberOfPicks Value, probabilityString Value, set Value) (Value, error) {
	picks, ok := numberOfPicks.(*IntValue)
	if !ok || picks.Val < 0 {
		return nil, newTLCError(ECGeneral, "first argument of RandomSubsetSet must be a nonnegative integer, got %s", numberOfPicks)
	}
	probText, ok := probabilityString.(*StringValue)
	if !ok {
		return nil, newTLCError(ECGeneral, "second argument of RandomSubsetSet must be a string probability, got %s", probabilityString)
	}
	probability, err := strconv.ParseFloat(probText.Val.String(), 64)
	if err != nil || probability < 0 || probability > 1 {
		return nil, newTLCError(ECGeneral, "second argument of RandomSubsetSet must parse to a probability in [0,1], got %s", probabilityString)
	}
	size, err := set.Size()
	if err != nil {
		return nil, err
	}
	if err := checkRandomSubsetPickCount(int(picks.Val), size); err != nil {
		return nil, err
	}
	return randomSetOfSubsets(int(picks.Val), probability, set)
}

func randomSubsetOfEnumerable(k int, value Value) (Value, error) {
	enum, ok := asEnumerable(value)
	if !ok {
		return nil, newTLCError(ECGeneral, "expected a finite enumerable set, got %s", value)
	}
	finite, err := value.IsFinite()
	if err != nil {
		return nil, err
	}
	if !finite {
		return nil, newTLCError(ECGeneral, "expected a finite enumerable set, got %s", value)
	}
	values := NewValueVec(0)
	elements := enum.Elements()
	for elem := elements.NextElement(); elem != nil; elem = elements.NextElement() {
		values.Add(elem)
	}
	if err := elements.Err(); err != nil {
		return nil, err
	}
	n := values.Len()
	out := NewValueVec(k)
	for _, index := range randomSubsetIndices(k, n) {
		out.Add(values.At(index))
	}
	return NewSetEnumValueVec(out, false), nil
}

func randomSubsetIndices(k int, n int) []int {
	if k <= 0 || n <= 0 {
		return nil
	}
	m, a := randomSubsetLCGParameters(n)
	rng := RandomEnumerableGenerator()
	index := int(rng.NextIntN(int32(n)))
	c := rng.NextPrime()
	indices := make([]int, 0, k)
	for len(indices) < k {
		for {
			index = int((int64(a)*int64(index) + int64(c)) % int64(m))
			if index < n {
				break
			}
		}
		indices = append(indices, index)
	}
	return indices
}

func randomSubsetLCGParameters(n int) (int, int) {
	if n < 9 {
		n = 9
	}
	factors := primeFactors(n)
	for n%4 == 0 || primeFactorsAreSquareFree(factors) {
		n++
		factors = primeFactors(n)
	}
	a := 1
	var used []int
	for _, factor := range factors {
		seen := false
		for _, existing := range used {
			if existing == factor {
				seen = true
				break
			}
		}
		if !seen {
			used = append(used, factor)
			a *= factor
		}
	}
	return n, a + 1
}

func primeFactorsAreSquareFree(factors []int) bool {
	for i := 0; i < len(factors); i++ {
		for j := i + 1; j < len(factors); j++ {
			if factors[i] == factors[j] {
				return false
			}
		}
	}
	return true
}

func primeFactors(n int) []int {
	var factors []int
	for n%2 == 0 {
		factors = append(factors, 2)
		n /= 2
	}
	for d := 3; d*d <= n; d += 2 {
		for n%d == 0 {
			factors = append(factors, d)
			n /= d
		}
	}
	if n > 1 {
		factors = append(factors, n)
	}
	return factors
}

func randomSetOfSubsets(k int, probability float64, value Value) (Value, error) {
	set, err := toSetEnumValue(value)
	if err != nil {
		return nil, err
	}
	base := set.Elems.ToArray()
	subsets := NewValueVec(k)
	rng := RandomEnumerableGenerator()
	for i := 0; i < k; i++ {
		subset := NewValueVec(0)
		for _, elem := range base {
			if rng.NextDouble() < probability {
				subset.Add(elem)
			}
		}
		subsets.Add(NewSetEnumValueVec(subset, false))
	}
	return NewSetEnumValueVec(subsets, false), nil
}

func checkRandomSubsetPickCount(picks int, size int) error {
	if size >= 31 {
		return nil
	}
	max := int(math.Pow(2, float64(size)))
	if picks > max {
		return newTLCError(ECGeneral, "number of requested random subsets %d exceeds subset space size 2^%d", picks, size)
	}
	return nil
}
