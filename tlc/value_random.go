package tlc

import (
	"fmt"
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
	k = int(int32(k))
	n = int(int32(n))
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

func (e *randomIndexedValueEnumeration) Reset() {
	requireValueEnumerationReceiver(e)
	e.indices.reset()
}
func (e *randomIndexedValueEnumeration) Err() error { return nil }
func (e *randomIndexedValueEnumeration) NextElement() Value {
	requireValueEnumerationReceiver(e)
	if !e.indices.hasNext() {
		return nil
	}
	return e.element(e.indices.nextIndex())
}

// randomValueEnumeration is the concrete Go dispatch for elements(k), distinct
// from both getRandomSubset(k) and elements(Ordering.RANDOMIZED).
func randomValueEnumeration(value Value, k int) (ValueEnumeration, error) {
	k = int(int32(k))
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

func enumerableElementsK(value Value, k int) ValueEnumeration {
	if isNil(value) {
		panic(NewNullPointerException())
	}
	elements, err := randomValueEnumeration(value, k)
	if err != nil {
		panic(err)
	}
	return elements
}

func enumerableElementsOrdered(value Value, ordering EnumerableOrdering) ValueEnumeration {
	if isNil(value) {
		panic(NewNullPointerException())
	}
	enumerable, ok := asEnumerable(value)
	if !ok {
		panic(valueStreamClassCast(value, "tlc2.value.impl.Enumerable"))
	}
	switch v := value.(type) {
	case *SubsetValue:
		if ordering != EnumerableOrderingRandomized {
			return v.Elements()
		}
		set, err := v.ToSetEnum()
		if err != nil {
			panic(err)
		}
		if set == nil {
			panic(NewNullPointerException())
		}
		return set.ElementsOrdered(ordering)
	case *KSubsetValue:
		empty, err := v.hasNoElements()
		if err != nil {
			panic(err)
		}
		if empty {
			return EmptySet.ElementsOrdered(ordering)
		}
		if ordering == EnumerableOrderingRandomized {
			elements, err := newRandomKSubsetEnumeration(v)
			if err != nil {
				panic(err)
			}
			return elements
		}
		return v.Elements()
	}
	if ordering == EnumerableOrderingNormalized {
		set, err := tryToSetEnumValue(value)
		if err != nil {
			panic(err)
		}
		if set != nil {
			set.Normalize()
			return set.Elements()
		}
	} else if ordering == EnumerableOrderingRandomized {
		size, err := value.Size()
		if err != nil {
			panic(err)
		}
		return enumerable.ElementsK(size)
	}
	return enumerable.Elements()
}

func (v *SetEnumValue) ElementsK(k int) ValueEnumeration {
	return enumerableElementsK(v, k)
}

func (v *SetEnumValue) ElementsOrdered(ordering EnumerableOrdering) ValueEnumeration {
	return enumerableElementsOrdered(v, ordering)
}

func (v *IntervalValue) ElementsK(k int) ValueEnumeration {
	return enumerableElementsK(v, k)
}

func (v *IntervalValue) ElementsOrdered(ordering EnumerableOrdering) ValueEnumeration {
	return enumerableElementsOrdered(v, ordering)
}

func (v *SetPredValue) ElementsK(k int) ValueEnumeration {
	return enumerableElementsK(v, k)
}

func (v *SetPredValue) ElementsOrdered(ordering EnumerableOrdering) ValueEnumeration {
	return enumerableElementsOrdered(v, ordering)
}

func (v *SetCupValue) ElementsK(k int) ValueEnumeration {
	return enumerableElementsK(v, k)
}

func (v *SetCupValue) ElementsOrdered(ordering EnumerableOrdering) ValueEnumeration {
	return enumerableElementsOrdered(v, ordering)
}

func (v *SetCapValue) ElementsK(k int) ValueEnumeration {
	return enumerableElementsK(v, k)
}

func (v *SetCapValue) ElementsOrdered(ordering EnumerableOrdering) ValueEnumeration {
	return enumerableElementsOrdered(v, ordering)
}

func (v *SetDiffValue) ElementsK(k int) ValueEnumeration {
	return enumerableElementsK(v, k)
}

func (v *SetDiffValue) ElementsOrdered(ordering EnumerableOrdering) ValueEnumeration {
	return enumerableElementsOrdered(v, ordering)
}

func (v *UnionValue) ElementsK(k int) ValueEnumeration {
	return enumerableElementsK(v, k)
}

func (v *UnionValue) ElementsOrdered(ordering EnumerableOrdering) ValueEnumeration {
	return enumerableElementsOrdered(v, ordering)
}

func (v *SetOfTuplesValue) ElementsK(k int) ValueEnumeration {
	return enumerableElementsK(v, k)
}

func (v *SetOfTuplesValue) ElementsOrdered(ordering EnumerableOrdering) ValueEnumeration {
	return enumerableElementsOrdered(v, ordering)
}

func (v *SetOfRcdsValue) ElementsK(k int) ValueEnumeration {
	return enumerableElementsK(v, k)
}

func (v *SetOfRcdsValue) ElementsOrdered(ordering EnumerableOrdering) ValueEnumeration {
	return enumerableElementsOrdered(v, ordering)
}

func (v *SetOfFcnsValue) ElementsK(k int) ValueEnumeration {
	return enumerableElementsK(v, k)
}

func (v *SetOfFcnsValue) ElementsOrdered(ordering EnumerableOrdering) ValueEnumeration {
	return enumerableElementsOrdered(v, ordering)
}

func (v *SubsetValue) ElementsK(k int) ValueEnumeration {
	return enumerableElementsK(v, k)
}

func (v *SubsetValue) ElementsOrdered(ordering EnumerableOrdering) ValueEnumeration {
	return enumerableElementsOrdered(v, ordering)
}

func (v *KSubsetValue) ElementsK(k int) ValueEnumeration {
	return enumerableElementsK(v, k)
}

func (v *KSubsetValue) ElementsOrdered(ordering EnumerableOrdering) ValueEnumeration {
	return enumerableElementsOrdered(v, ordering)
}

// GetRandomSubset ports IntervalValue.getRandomSubset.
func (v *IntervalValue) GetRandomSubset(k int) (*SetEnumValue, error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	values := NewValueVec(k)
	elements, err := randomValueEnumeration(v, k)
	if err != nil {
		return nil, err
	}
	for elem := elements.NextElement(); elem != nil; elem = elements.NextElement() {
		values.Add(elem)
	}
	if err := elements.Err(); err != nil {
		return nil, err
	}
	return NewSetEnumValueVec(values, false, v.CM), nil
}

// GetRandomSubset ports SetEnumValue.getRandomSubset.
func (v *SetEnumValue) GetRandomSubset(k int) (*SetEnumValue, error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	values := NewValueVec(k)
	elements, err := randomValueEnumeration(v, k)
	if err != nil {
		return nil, err
	}
	for elem := elements.NextElement(); elem != nil; elem = elements.NextElement() {
		values.Add(elem)
	}
	if err := elements.Err(); err != nil {
		return nil, err
	}
	return NewSetEnumValueVec(values, false, v.CM), nil
}

func defaultEnumerableRandomSubset(value Value, k int) (*SetEnumValue, error) {
	if isNil(value) {
		panic(NewNullPointerException())
	}
	set, err := tryToSetEnumValue(value)
	if err != nil {
		return nil, err
	}
	if set == nil {
		panic(NewNullPointerException())
	}
	return set.GetRandomSubset(k)
}

func (v *SetPredValue) GetRandomSubset(k int) (*SetEnumValue, error) {
	return defaultEnumerableRandomSubset(v, k)
}

func (v *SetCupValue) GetRandomSubset(k int) (*SetEnumValue, error) {
	return defaultEnumerableRandomSubset(v, k)
}

func (v *SetCapValue) GetRandomSubset(k int) (*SetEnumValue, error) {
	return defaultEnumerableRandomSubset(v, k)
}

func (v *SetDiffValue) GetRandomSubset(k int) (*SetEnumValue, error) {
	return defaultEnumerableRandomSubset(v, k)
}

func (v *UnionValue) GetRandomSubset(k int) (*SetEnumValue, error) {
	return defaultEnumerableRandomSubset(v, k)
}

func (v *SubsetValue) GetRandomSubset(k int) (*SetEnumValue, error) {
	return defaultEnumerableRandomSubset(v, k)
}

func (v *KSubsetValue) GetRandomSubset(k int) (*SetEnumValue, error) {
	return defaultEnumerableRandomSubset(v, k)
}

func getRandomProductSubset(value Value, k int) (*SetEnumValue, error) {
	if isNil(value) {
		panic(NewNullPointerException())
	}
	k = int(int32(k))
	values := NewValueVec(k)
	elements, err := newRandomProductEnumeration(value, k)
	if err != nil {
		return nil, err
	}
	for elem := elements.NextElement(); elem != nil; elem = elements.NextElement() {
		values.Add(elem)
	}
	if err := elements.Err(); err != nil {
		return nil, err
	}
	value.GetCostModel().incValueSecondary(int64(values.Len()))
	return NewSetEnumValueVec(values, false, value.GetCostModel()), nil
}

// GetRandomSubset ports SetOfFcnsOrRcdsValue.getRandomSubset.
func (v *SetOfFcnsValue) GetRandomSubset(k int) (*SetEnumValue, error) {
	return getRandomProductSubset(v, k)
}

// GetRandomSubset ports SetOfFcnsOrRcdsValue.getRandomSubset.
func (v *SetOfRcdsValue) GetRandomSubset(k int) (*SetEnumValue, error) {
	return getRandomProductSubset(v, k)
}

// GetRandomSubset ports SetOfFcnsOrRcdsValue.getRandomSubset.
func (v *SetOfTuplesValue) GetRandomSubset(k int) (*SetEnumValue, error) {
	return getRandomProductSubset(v, k)
}

func randomSubsetValueEnumeration(base Value, cm CostModel, k int) (ValueEnumeration, error) {
	k = int(int32(k))
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
	k = int(int32(k))
	set, err := tryToSetEnumValue(base)
	if err != nil {
		return nil, err
	}
	set.Normalize()
	return &coinTossingSubsetEnumeration{elems: set.Elems, cm: cm, probability: probability, k: k}, nil
}

func (e *coinTossingSubsetEnumeration) Reset() {
	requireValueEnumerationReceiver(e)
	e.i = 0
}
func (e *coinTossingSubsetEnumeration) Err() error { return nil }
func (e *coinTossingSubsetEnumeration) NextElement() Value {
	requireValueEnumerationReceiver(e)
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
		empty, err := IsEmptyValue(v.ownedDomain())
		if err != nil || empty {
			return false, err
		}
		empty, err = IsEmptyValue(v.ownedRange())
		if err != nil || empty {
			return false, err
		}
		rsz, err := v.ownedRange().Size()
		if err != nil || rsz == 1 {
			return false, err
		}
		rsz, err = v.ownedRange().Size()
		if err != nil {
			return false, err
		}
		dsz, err := v.ownedDomain().Size()
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
		factors = v.ownedValues()
	case *SetOfTuplesValue:
		factors = v.ownedSets()
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
	k = int(int32(k))
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

func (e *randomBigProductEnumeration) Reset() {
	requireValueEnumerationReceiver(e)
	e.i = 0
}
func (e *randomBigProductEnumeration) Err() error { return nil }
func (e *randomBigProductEnumeration) NextElement() Value {
	requireValueEnumerationReceiver(e)
	if e.i >= e.k {
		return nil
	}
	index := new(big.Int).Mul(big.NewInt(math.MaxInt64-24), big.NewInt(int64(e.i)))
	e.i++
	index.Add(index, e.offset)
	index.Mod(index, e.size)
	return e.product.elementAtBig(index)
}

// RandomSubsetGenerator is SubsetValue.RandomSubsetGenerator, the endless
// Algorithm S enumeration used by KSubsetValue's randomized ordering. Java
// intentionally drops the cost model and makes Reset a no-op here.
type RandomSubsetGenerator struct {
	set *SetEnumValue
	k   int
	rng *JavaRandom
}

func newRandomSubsetGenerator(base Value, k int) (*RandomSubsetGenerator, error) {
	set, err := toSetEnumValue(base)
	if err != nil {
		return nil, err
	}
	if _, err := set.normalizeSet(); err != nil {
		return nil, err
	}
	if k < 0 || k > set.Elems.Len() {
		return nil, NewIllegalArgumentException(fmt.Sprintf("k=%d and n=%d", k, set.Elems.Len()))
	}
	return &RandomSubsetGenerator{set: set, k: k, rng: RandomEnumerableGenerator()}, nil
}

func (v *SubsetValue) NewRandomSubsetGenerator(k int) (*RandomSubsetGenerator, error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	return newRandomSubsetGenerator(v.Set, int(int32(k)))
}

func (v *KSubsetValue) NewRandomSubsetGenerator(k int) (*RandomSubsetGenerator, error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	return newRandomSubsetGenerator(v.Set, int(int32(k)))
}

func newRandomKSubsetEnumeration(value *KSubsetValue) (ValueEnumeration, error) {
	empty, err := value.hasNoElements()
	if err != nil {
		return nil, err
	}
	if empty {
		return randomValueEnumeration(EmptySet, 0)
	}
	return value.NewRandomSubsetGenerator(value.ownedK())
}

func (e *RandomSubsetGenerator) Reset()     { requireValueEnumerationReceiver(e) }
func (e *RandomSubsetGenerator) Err() error { return nil }
func (e *RandomSubsetGenerator) NextElement() Value {
	requireValueEnumerationReceiver(e)
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
