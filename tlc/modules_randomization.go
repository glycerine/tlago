package tlc

import (
	"fmt"
	"math/big"
	"math/bits"
	"strconv"
)

func RandomSubset(k Value, set Value) (Value, error) {
	count, ok := k.(*IntValue)
	if !ok {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "RandomSubset", "nonnegative integer", ValuesPPR(k))
	}
	return randomSubsetOfEnumerable(int(count.Val), set)
}

func RandomSetOfSubsets(numberOfPicks Value, subsetSize Value, set Value) (Value, error) {
	picks, ok := numberOfPicks.(*IntValue)
	if !ok || picks.Val < 0 {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "RandomSetOfSubsets", "nonnegative integer", ValuesPPR(numberOfPicks))
	}
	n, ok := subsetSize.(*IntValue)
	if !ok || n.Val < 0 {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "RandomSetOfSubsets", "nonnegative integer", ValuesPPR(subsetSize))
	}
	size, err := randomizationFiniteSetSize("third", "RandomSetOfSubsets", "finite set", set)
	if err != nil {
		return nil, err
	}
	if err := checkRandomSubsetPickCount("RandomSetOfSubsets", int(picks.Val), size); err != nil {
		return nil, err
	}
	if int(n.Val) > size {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "RandomSetOfSubsets", "nonnegative integer in range 0..Cardinality(S)", ValuesPPR(subsetSize))
	}
	probability := float64(n.Val) / float64(size)
	if probability < 0 || probability > 1 {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "RandomSetOfSubsets", "nonnegative integer in range 0..Cardinality(S)", ValuesPPR(subsetSize))
	}
	return randomSetOfSubsets(int(picks.Val), probability, set)
}

func RandomSubsetSet(numberOfPicks Value, probabilityString Value, set Value) (Value, error) {
	const operator = "RandomSubsetSetProbability"
	picks, ok := numberOfPicks.(*IntValue)
	if !ok || picks.Val < 0 {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", operator, "nonnegative integer", ValuesPPR(numberOfPicks))
	}
	probText, ok := probabilityString.(*StringValue)
	if !ok {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", operator, "string literal representing a probability", ValuesPPR(probabilityString))
	}
	// Java Double.valueOf accepts its own literal grammar, suffixes and ASCII
	// whitespace; Go ParseFloat alone does not preserve that grammar.
	probability, err := parseJavaDoubleProperty(probText.Val.String())
	if err != nil || probability < 0 || probability > 1 {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", operator, "string literal does not represent a parsable probability", ValuesPPR(probabilityString))
	}
	size, err := randomizationFiniteSetSize("third", operator, "finite set", set)
	if err != nil {
		return nil, err
	}
	if err := checkRandomSubsetPickCount(operator, int(picks.Val), size); err != nil {
		return nil, err
	}
	return randomSetOfSubsets(int(picks.Val), probability, set)
}

func randomSubsetOfEnumerable(k int, value Value) (Value, error) {
	_, ok := asEnumerable(value)
	if !ok {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "RandomSubset", "a finite set", ValuesPPR(value))
	}
	finite, err := value.IsFinite()
	if err != nil {
		return nil, err
	}
	if !finite {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "RandomSubset", "a finite set", ValuesPPR(value))
	}
	if subset, handled, err := randomSubsetOfProductValue(k, value); handled {
		return subset, err
	}
	// Java's default getRandomSubset materializes first, then delegates to
	// SetEnumValue. Only intervals and product sets override that default.
	if _, ok := value.(*IntervalValue); !ok {
		value, err = toSetEnumValue(value)
		if err != nil {
			return nil, err
		}
	}
	if k < 0 {
		return nil, randomSubsetNegativeSizeError(k)
	}
	elements, err := randomValueEnumeration(value, k)
	if err != nil {
		return nil, err
	}
	return collectRandomSubset(elements, k, value.GetCostModel())
}

func randomSubsetNegativeSizeError(k int) error {
	const signature = "public static tlc2.value.impl.Value tlc2.module.Randomization.RandomSubset(tlc2.value.impl.Value,tlc2.value.impl.Value)"
	return javaMethodOverrideRuntimeError(signature, strconv.Itoa(k))
}

func collectRandomSubset(elements ValueEnumeration, k int, cm CostModel) (*SetEnumValue, error) {
	out := NewValueVec(k)
	for elem := elements.NextElement(); elem != nil; elem = elements.NextElement() {
		out.Add(elem)
	}
	if err := elements.Err(); err != nil {
		return nil, err
	}
	return NewSetEnumValueVec(out, false, cm), nil
}

type randomProductValue struct {
	constituents []*SetEnumValue
	makeValue    func([]Value) Value
	rescaleBy    []int
	bigRescaleBy []*big.Int
}

func randomSubsetOfProductValue(k int, value Value) (Value, bool, error) {
	switch value.(type) {
	case *SetOfFcnsValue, *SetOfRcdsValue, *SetOfTuplesValue:
	default:
		return nil, false, nil
	}
	if k < 0 {
		return nil, true, randomSubsetNegativeSizeError(k)
	}
	elements, err := newRandomProductEnumeration(value, k)
	if err != nil {
		return nil, true, err
	}
	set, err := collectRandomSubset(elements, k, value.GetCostModel())
	if err != nil {
		return nil, true, err
	}
	value.GetCostModel().incValueSecondary(int64(set.Elems.Len()))
	return set, true, nil
}

func randomProductForValue(value Value) (*randomProductValue, bool, error) {
	switch v := value.(type) {
	case *SetOfFcnsValue:
		domSet, err := toSetEnumValue(v.Domain)
		if err != nil {
			return nil, true, err
		}
		if _, err := domSet.normalizeSet(); err != nil {
			return nil, true, err
		}
		rangeSet, err := toSetEnumValue(v.Range)
		if err != nil {
			return nil, true, err
		}
		constituents := make([]*SetEnumValue, domSet.Elems.Len())
		for i := range constituents {
			constituents[i] = rangeSet
		}
		domain := domSet.Elems.ToArray()
		return &randomProductValue{
			constituents: constituents,
			makeValue: func(values []Value) Value {
				return NewFcnRcdValue(domain, values, true)
			},
		}, true, nil
	case *SetOfRcdsValue:
		constituents, err := randomProductConstituents(v.Values)
		if err != nil {
			return nil, true, err
		}
		return &randomProductValue{
			constituents: constituents,
			makeValue: func(values []Value) Value {
				return NewRecordValue(v.Names, values, false, v.CM)
			},
		}, true, nil
	case *SetOfTuplesValue:
		constituents, err := randomProductConstituents(v.Sets)
		if err != nil {
			return nil, true, err
		}
		return &randomProductValue{
			constituents: constituents,
			makeValue: func(values []Value) Value {
				return NewTupleValue(values, v.CM)
			},
		}, true, nil
	default:
		return nil, false, nil
	}
}

func randomProductConstituents(values []Value) ([]*SetEnumValue, error) {
	constituents := make([]*SetEnumValue, len(values))
	for i, value := range values {
		set, err := toSetEnumValue(value)
		if err != nil {
			return nil, err
		}
		constituents[i] = set
	}
	return constituents, nil
}

func (p *randomProductValue) initializeRadices(useBig bool) *big.Int {
	if useBig {
		p.bigRescaleBy = make([]*big.Int, len(p.constituents))
		numElems := big.NewInt(1)
		for i := len(p.constituents) - 1; i >= 0; i-- {
			p.bigRescaleBy[i] = new(big.Int).Set(numElems)
			numElems.Mul(numElems, big.NewInt(int64(p.constituents[i].Elems.Len())))
		}
		return numElems
	}
	p.rescaleBy = make([]int, len(p.constituents))
	numElems := int32(1)
	for i := len(p.constituents) - 1; i >= 0; i-- {
		p.rescaleBy[i] = int(numElems)
		numElems *= int32(p.constituents[i].Elems.Len())
	}
	return nil
}

func (p *randomProductValue) elementAtInt(index int) Value {
	values := make([]Value, len(p.constituents))
	for i, set := range p.constituents {
		values[i] = set.Elems.At((index / p.rescaleBy[i]) % set.Elems.Len())
	}
	return p.makeValue(values)
}

func (p *randomProductValue) elementAtBig(index *big.Int) Value {
	values := make([]Value, len(p.constituents))
	for i, set := range p.constituents {
		scaled := new(big.Int).Div(new(big.Int).Set(index), p.bigRescaleBy[i])
		scaled.Mod(scaled, big.NewInt(int64(set.Elems.Len())))
		values[i] = set.Elems.At(int(scaled.Int64()))
	}
	return p.makeValue(values)
}

func randomSubsetIndices(k int, n int) []int {
	e := newRandomSubsetIndices(k, n)
	indices := make([]int, 0)
	for e.hasNext() {
		indices = append(indices, e.nextIndex())
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
	set, err := NewSubsetValue(value).GetRandomSetOfSubsets(k, probability)
	if err != nil {
		return nil, err
	}
	return set, nil
}

type valueHashSet struct {
	buckets map[int32][]Value
	values  *ValueVec
}

func newValueHashSet(estimated int) *valueHashSet {
	if estimated < 0 {
		panic(fmt.Errorf("Illegal initial capacity: %d", estimated))
	}
	return &valueHashSet{
		buckets: make(map[int32][]Value, estimated),
		values:  NewValueVec(estimated),
	}
}

func (s *valueHashSet) add(value Value) bool {
	hash := ValueJavaHashCode(value)
	for _, existing := range s.buckets[hash] {
		eq, err := value.Equal(existing)
		if err != nil {
			panic(err)
		}
		if eq {
			return false
		}
	}
	s.buckets[hash] = append(s.buckets[hash], value)
	s.values.Add(value)
	return true
}

func randomizationFiniteSetSize(position string, operator string, expected string, value Value) (int, error) {
	_, ok := asEnumerable(value)
	if !ok {
		return 0, newTLCErrorCode(ECTLCModuleArgumentError, position, operator, expected, ValuesPPR(value))
	}
	finite, err := value.IsFinite()
	if err != nil {
		return 0, err
	}
	if !finite {
		return 0, newTLCErrorCode(ECTLCModuleArgumentError, position, operator, expected, ValuesPPR(value))
	}
	size, err := value.Size()
	if err != nil {
		return 0, err
	}
	return size, nil
}

func checkRandomSubsetPickCount(operator string, picks int, size int) error {
	if 31-bits.LeadingZeros32(uint32(picks))+1 > size && picks > int(javaIntOneLeftShift(size)) {
		expected := "nonnegative integer that is smaller than the subset's size of 2^" + strconv.Itoa(size)
		return newTLCErrorCode(ECTLCModuleArgumentError, "first", operator, expected, strconv.Itoa(picks))
	}
	return nil
}
