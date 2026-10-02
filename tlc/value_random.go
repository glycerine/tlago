package tlc

import (
	"math"
	"math/big"
)

// randomSubsetIndexEnumeration mirrors EnumerableValue.SubsetEnumerator.
// Reset resets the call count but deliberately retains the current LCG seed.
type randomSubsetIndexEnumeration struct {
	n, k, i int
	index   int
	m, a, c int
}

func newRandomSubsetIndices(k, n int) *randomSubsetIndexEnumeration {
	e := &randomSubsetIndexEnumeration{}
	if n <= 0 {
		return e
	}
	e.n, e.k = n, k
	e.m, e.a = randomSubsetLCGParameters(n)
	rng := RandomEnumerableGenerator()
	e.index = int(rng.NextIntN(int32(n)))
	e.c = rng.NextPrime()
	return e
}

func (e *randomSubsetIndexEnumeration) hasNext() bool { return e.i < e.k }
func (e *randomSubsetIndexEnumeration) reset()        { e.i = 0 }
func (e *randomSubsetIndexEnumeration) nextIndex() int {
	if e.n <= 0 {
		e.i++
		return 0
	}
	for {
		e.index = int((int64(e.a)*int64(e.index) + int64(e.c)) % int64(e.m))
		if e.index < e.n {
			break
		}
	}
	e.i++
	return e.index
}

type randomIndexedValueEnumeration struct {
	indices *randomSubsetIndexEnumeration
	element func(int) Value
}

func (e *randomIndexedValueEnumeration) Reset()     { e.indices.reset() }
func (e *randomIndexedValueEnumeration) Err() error { return nil }
func (e *randomIndexedValueEnumeration) NextElement() Value {
	if !e.indices.hasNext() {
		return nil
	}
	return e.element(e.indices.nextIndex())
}

// randomValueEnumeration is the concrete Go dispatch for elements(k), distinct
// from both getRandomSubset(k) and elements(Ordering.RANDOMIZED).
func randomValueEnumeration(value Value, k int) (ValueEnumeration, error) {
	switch v := value.(type) {
	case *SetEnumValue:
		if _, err := v.normalizeSet(); err != nil {
			return nil, err
		}
		n, err := v.Size()
		if err != nil {
			return nil, err
		}
		return &randomIndexedValueEnumeration{newRandomSubsetIndices(k, n), v.Elems.At}, nil
	case *IntervalValue:
		n, err := v.Size()
		if err != nil {
			return nil, err
		}
		return &randomIndexedValueEnumeration{
			indices: newRandomSubsetIndices(k, n),
			element: func(index int) Value { return NewIntValue(v.Low + int32(index)) },
		}, nil
	case *SubsetValue:
		return randomSubsetValueEnumeration(v.Set, v.CM, k)
	case *KSubsetValue:
		// Java inherits SubsetValue.elements(k); its ordering overload differs.
		return randomSubsetValueEnumeration(v.Set, v.CM, k)
	case *SetOfFcnsValue, *SetOfRcdsValue, *SetOfTuplesValue:
		return newRandomProductEnumeration(value, k)
	}
	enumerable, ok := asEnumerable(value)
	if !ok {
		return nil, newTLCError(ECGeneral, "value is not enumerable")
	}
	elements := enumerable.Elements()
	values := make([]Value, 0)
	for elem := elements.NextElement(); elem != nil; elem = elements.NextElement() {
		values = append(values, elem)
	}
	if err := elements.Err(); err != nil {
		return nil, err
	}
	n, err := value.Size()
	if err != nil {
		return nil, err
	}
	return &randomIndexedValueEnumeration{
		indices: newRandomSubsetIndices(k, n),
		element: func(index int) Value { return values[index] },
	}, nil
}

func randomSubsetValueEnumeration(base Value, cm CostModel, k int) (ValueEnumeration, error) {
	n, err := base.Size()
	if err != nil {
		return nil, err
	}
	if n >= 31 || k > 1<<16 {
		return newCoinTossingSubsetEnumeration(base, cm, k, 0.5)
	}
	// The Java superclass initializes the random seed before converting the base.
	n, err = base.Size()
	if err != nil {
		return nil, err
	}
	indices := newRandomSubsetIndices(k, int(javaIntOneLeftShift(n)))
	set, err := toSetEnumValue(base)
	if err != nil {
		return nil, err
	}
	if _, err := set.normalizeSet(); err != nil {
		return nil, err
	}
	return &randomIndexedValueEnumeration{
		indices: indices,
		element: func(bits int) Value {
			vals := NewValueVec(bitsOnesCount(bits))
			for i := 0; bits > 0 && i < set.Elems.Len(); i++ {
				if bits&1 != 0 {
					vals.Add(set.Elems.At(i))
				}
				bits = int(uint(bits) >> 1)
			}
			return NewSetEnumValueVec(vals, false, cm)
		},
	}, nil
}

func bitsOnesCount(bits int) int {
	count := 0
	for bits > 0 {
		count += bits & 1
		bits = int(uint(bits) >> 1)
	}
	return count
}

type coinTossingSubsetEnumeration struct {
	elems       *ValueVec
	cm          CostModel
	probability float64
	k, i        int
}

func newCoinTossingSubsetEnumeration(base Value, cm CostModel, k int, probability float64) (*coinTossingSubsetEnumeration, error) {
	set, err := tryToSetEnumValue(base)
	if err != nil {
		return nil, err
	}
	set.Normalize()
	return &coinTossingSubsetEnumeration{elems: set.Elems, cm: cm, probability: probability, k: k}, nil
}

func (e *coinTossingSubsetEnumeration) Reset()     { e.i = 0 }
func (e *coinTossingSubsetEnumeration) Err() error { return nil }
func (e *coinTossingSubsetEnumeration) NextElement() Value {
	if e.i >= e.k {
		return nil
	}
	vals := NewValueVec(e.elems.Len())
	for i := 0; i < e.elems.Len(); i++ {
		if RandomEnumerableGenerator().NextDouble() < e.probability {
			vals.Add(e.elems.At(i))
		}
	}
	e.i++
	return NewSetEnumValueVec(vals, false, e.cm)
}

func randomProductNeedsBigInteger(value Value) (bool, error) {
	var factors []Value
	switch v := value.(type) {
	case *SetOfFcnsValue:
		empty, err := IsEmptyValue(v.Domain)
		if err != nil || empty {
			return false, err
		}
		empty, err = IsEmptyValue(v.Range)
		if err != nil || empty {
			return false, err
		}
		rsz, err := v.Range.Size()
		if err != nil || rsz == 1 {
			return false, err
		}
		rsz, err = v.Range.Size()
		if err != nil {
			return false, err
		}
		dsz, err := v.Domain.Size()
		if err != nil {
			return false, err
		}
		sz := int64(1)
		for i := 0; i < dsz; i++ {
			sz *= int64(rsz)
			if sz < math.MinInt32 || sz > math.MaxInt32 {
				return true, nil
			}
		}
		return false, nil
	case *SetOfRcdsValue:
		factors = v.Values
	case *SetOfTuplesValue:
		factors = v.Sets
	}
	empty, err := IsEmptyValue(value)
	if err != nil || empty {
		return false, err
	}
	sz := int64(1)
	for _, factor := range factors {
		n, err := factor.Size()
		if err != nil {
			return false, err
		}
		sz *= int64(n)
		if sz < math.MinInt32 || sz > math.MaxInt32 {
			return true, nil
		}
	}
	return false, nil
}

type randomBigProductEnumeration struct {
	product *randomProductValue
	offset  *big.Int
	size    *big.Int
	k, i    int
}

func newRandomProductEnumeration(value Value, k int) (ValueEnumeration, error) {
	empty, err := IsEmptyValue(value)
	if err != nil {
		return nil, err
	}
	if empty {
		return &sliceValueEnumeration{}, nil
	}
	useBig, err := randomProductNeedsBigInteger(value)
	if err != nil {
		return nil, err
	}
	if useBig {
		a := RandomEnumerableGenerator().NextLong()
		if a < 0 {
			a = -a // Java Math.abs preserves Long.MIN_VALUE.
		}
		product, _, err := randomProductForValue(value)
		if err != nil {
			return nil, err
		}
		return &randomBigProductEnumeration{product: product, offset: big.NewInt(a), size: product.initializeRadices(true), k: k}, nil
	}
	n, err := value.Size()
	if err != nil {
		return nil, err
	}
	indices := newRandomSubsetIndices(k, n)
	product, _, err := randomProductForValue(value)
	if err != nil {
		return nil, err
	}
	product.initializeRadices(false)
	return &randomIndexedValueEnumeration{indices: indices, element: product.elementAtInt}, nil
}

func (e *randomBigProductEnumeration) Reset()     { e.i = 0 }
func (e *randomBigProductEnumeration) Err() error { return nil }
func (e *randomBigProductEnumeration) NextElement() Value {
	if e.i >= e.k {
		return nil
	}
	index := new(big.Int).Mul(big.NewInt(math.MaxInt64-24), big.NewInt(int64(e.i)))
	e.i++
	index.Add(index, e.offset)
	index.Mod(index, e.size)
	return e.product.elementAtBig(index)
}

// KSubsetValue's randomized ordering is an endless Algorithm S generator.
// Java intentionally drops the cost model and makes Reset a no-op here.
type randomKSubsetEnumeration struct {
	set *SetEnumValue
	k   int
	rng *JavaRandom
}

func newRandomKSubsetEnumeration(value *KSubsetValue) (ValueEnumeration, error) {
	empty, err := value.hasNoElements()
	if err != nil {
		return nil, err
	}
	if empty {
		return randomValueEnumeration(EmptySet, 0)
	}
	set, err := toSetEnumValue(value.Set)
	if err != nil {
		return nil, err
	}
	if _, err := set.normalizeSet(); err != nil {
		return nil, err
	}
	if value.K < 0 || value.K > set.Elems.Len() {
		return nil, newTLCError(ECGeneral, "k=%d and n=%d", value.K, set.Elems.Len())
	}
	return &randomKSubsetEnumeration{set: set, k: value.K, rng: RandomEnumerableGenerator()}, nil
}

func (e *randomKSubsetEnumeration) Reset()     {}
func (e *randomKSubsetEnumeration) Err() error { return nil }
func (e *randomKSubsetEnumeration) NextElement() Value {
	vals := NewValueVec(e.k)
	n := e.set.Elems.Len()
	for t := 0; t < n; t++ {
		p := float64(e.k-vals.Len()) / float64(n-t)
		if e.rng.NextDouble() <= p {
			vals.Add(e.set.Elems.At(t))
		}
		if vals.Len() == e.k {
			break
		}
	}
	return NewSetEnumValueVec(vals, false)
}
