package tlc

import (
	"math"
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
	probability, err := strconv.ParseFloat(probText.Val.String(), 64)
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
	enum, ok := asEnumerable(value)
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
	if subset, ok := value.(*SubsetValue); ok {
		return randomSubsetOfSubsetValue(k, subset)
	}
	if subset, handled, err := randomSubsetOfProductValue(k, value); handled {
		return subset, err
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

func randomSubsetOfSubsetValue(k int, value *SubsetValue) (Value, error) {
	set, err := toSetEnumValue(value.Set)
	if err != nil {
		return nil, err
	}
	if _, err := set.normalizeSet(); err != nil {
		return nil, err
	}
	base := set.Elems
	out := NewValueVec(k)
	if base.Len() >= 31 || k > (1<<16) {
		rng := RandomEnumerableGenerator()
		for i := 0; i < k; i++ {
			subset := NewValueVec(base.Len())
			for j := 0; j < base.Len(); j++ {
				if rng.NextDouble() < 0.5 {
					subset.Add(base.At(j))
				}
			}
			out.Add(NewSetEnumValueVec(subset, false))
		}
		return NewSetEnumValueVec(out, false), nil
	}
	for _, bits := range randomSubsetIndices(k, 1<<base.Len()) {
		subset := NewValueVec(bitsOnesCount(bits))
		for i := 0; bits > 0 && i < base.Len(); i++ {
			if bits&0x1 > 0 {
				subset.Add(base.At(i))
			}
			bits = int(uint(bits) >> 1)
		}
		out.Add(NewSetEnumValueVec(subset, false))
	}
	return NewSetEnumValueVec(out, false), nil
}

func bitsOnesCount(bits int) int {
	count := 0
	for bits > 0 {
		count += bits & 1
		bits = int(uint(bits) >> 1)
	}
	return count
}

type randomProductValue struct {
	constituents []*SetEnumValue
	makeValue    func([]Value) Value
}

func randomSubsetOfProductValue(k int, value Value) (Value, bool, error) {
	product, handled, err := randomProductForValue(value)
	if !handled || err != nil {
		return nil, handled, err
	}
	return product.randomSubset(k), true, nil
}

func randomProductForValue(value Value) (*randomProductValue, bool, error) {
	switch v := value.(type) {
	case *SetOfFcnsValue:
		empty, err := IsEmptyValue(v)
		if err != nil || empty {
			return randomEmptyProduct(empty), true, err
		}
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
		empty, err := IsEmptyValue(v)
		if err != nil || empty {
			return randomEmptyProduct(empty), true, err
		}
		constituents, err := randomProductConstituents(v.Values)
		if err != nil {
			return nil, true, err
		}
		names := make([]*UniqueString, len(v.Names))
		copy(names, v.Names)
		return &randomProductValue{
			constituents: constituents,
			makeValue: func(values []Value) Value {
				return NewRecordValue(names, values, true)
			},
		}, true, nil
	case *SetOfTuplesValue:
		empty, err := IsEmptyValue(v)
		if err != nil || empty {
			return randomEmptyProduct(empty), true, err
		}
		constituents, err := randomProductConstituents(v.Sets)
		if err != nil {
			return nil, true, err
		}
		return &randomProductValue{
			constituents: constituents,
			makeValue: func(values []Value) Value {
				return NewTupleValue(values)
			},
		}, true, nil
	default:
		return nil, false, nil
	}
}

func randomEmptyProduct(empty bool) *randomProductValue {
	if !empty {
		return nil
	}
	return &randomProductValue{}
}

func randomProductConstituents(values []Value) ([]*SetEnumValue, error) {
	constituents := make([]*SetEnumValue, len(values))
	for i, value := range values {
		set, err := toSetEnumValue(value)
		if err != nil {
			return nil, err
		}
		if _, err := set.normalizeSet(); err != nil {
			return nil, err
		}
		constituents[i] = set
	}
	return constituents, nil
}

func (p *randomProductValue) randomSubset(k int) Value {
	if p == nil || p.makeValue == nil || k <= 0 {
		return NewSetEnumValueVec(NewValueVec(0), false)
	}
	size := p.cardinality()
	if size.Sign() <= 0 {
		return NewSetEnumValueVec(NewValueVec(0), false)
	}
	out := NewValueVec(k)
	if size.Cmp(big.NewInt(math.MaxInt32)) <= 0 {
		n := int(size.Int64())
		for _, index := range randomSubsetIndices(k, n) {
			out.Add(p.elementAtInt(index))
		}
	} else {
		rng := RandomEnumerableGenerator()
		a := rng.NextLong()
		if a < 0 {
			a = -a
		}
		offset := big.NewInt(a)
		x := big.NewInt(math.MaxInt64 - 24)
		for i := 0; i < k; i++ {
			index := new(big.Int).Mul(x, big.NewInt(int64(i)))
			index.Add(index, offset)
			index.Mod(index, size)
			out.Add(p.elementAtBig(index))
		}
	}
	return NewSetEnumValueVec(out, false)
}

func (p *randomProductValue) cardinality() *big.Int {
	size := big.NewInt(1)
	for _, set := range p.constituents {
		size.Mul(size, big.NewInt(int64(set.Elems.Len())))
	}
	return size
}

func (p *randomProductValue) elementAtInt(index int) Value {
	values := make([]Value, len(p.constituents))
	rescaleBy := make([]int, len(p.constituents))
	numElems := 1
	for i := len(p.constituents) - 1; i >= 0; i-- {
		rescaleBy[i] = numElems
		numElems *= p.constituents[i].Elems.Len()
	}
	for i, set := range p.constituents {
		values[i] = set.Elems.At((index / rescaleBy[i]) % set.Elems.Len())
	}
	return p.makeValue(values)
}

func (p *randomProductValue) elementAtBig(index *big.Int) Value {
	values := make([]Value, len(p.constituents))
	rescaleBy := make([]*big.Int, len(p.constituents))
	numElems := big.NewInt(1)
	for i := len(p.constituents) - 1; i >= 0; i-- {
		rescaleBy[i] = new(big.Int).Set(numElems)
		numElems.Mul(numElems, big.NewInt(int64(p.constituents[i].Elems.Len())))
	}
	for i, set := range p.constituents {
		scaled := new(big.Int).Div(new(big.Int).Set(index), rescaleBy[i])
		scaled.Mod(scaled, big.NewInt(int64(set.Elems.Len())))
		values[i] = set.Elems.At(int(scaled.Int64()))
	}
	return p.makeValue(values)
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
	if _, err := set.normalizeSet(); err != nil {
		return nil, err
	}
	base := set.Elems.ToArray()
	sets := newValueHashSet(int(float64(k) * probability))
	rng := RandomEnumerableGenerator()
	for i := 0; i < k; i++ {
		subset := NewValueVec(0)
		for _, elem := range base {
			if rng.NextDouble() < probability {
				subset.Add(elem)
			}
		}
		sets.add(NewSetEnumValueVec(subset, false))
	}
	return NewSetEnumValueVec(sets.values, false), nil
}

type valueHashSet struct {
	buckets map[int32][]Value
	values  *ValueVec
}

func newValueHashSet(estimated int) *valueHashSet {
	if estimated < 0 {
		estimated = 0
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
		if err == nil && eq {
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
