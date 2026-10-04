package tlc

import (
	"fmt"
	"math"
	"math/big"
	"sort"
)

// SubsetUnrank is SubsetValue.Unrank. Its sorted cutoffs implement TreeMap's
// strict lowerEntry lookup, including replacement of duplicate binomial keys.
type SubsetUnrank struct {
	owner              Value
	cutoffs            []subsetBinomialCutoff
	partialPascalTable []int64
	maxK               int
	elems              *ValueVec
	k                  int
}

type subsetBinomialCutoff struct {
	binomial int64
	choice   int
}

func NewSubsetUnrank(owner Value, k int, n int64, partialPascalTable []int64, elems *ValueVec, maxK int) *SubsetUnrank {
	u := &SubsetUnrank{owner: owner, k: k, partialPascalTable: partialPascalTable, elems: elems, maxK: maxK - 1}
	choice := k - 1
	if choice < 0 {
		choice = 0
	}
	u.putCutoff(-1, choice)
	for binomial := u.memoizedBinomial(choice, k); binomial < n; binomial = u.memoizedBinomial(choice, k) {
		choice++
		u.putCutoff(binomial, choice)
	}
	return u
}

func (u *SubsetUnrank) putCutoff(binomial int64, choice int) {
	i := sort.Search(len(u.cutoffs), func(i int) bool { return u.cutoffs[i].binomial >= binomial })
	if i < len(u.cutoffs) && u.cutoffs[i].binomial == binomial {
		u.cutoffs[i].choice = choice
		return
	}
	u.cutoffs = append(u.cutoffs, subsetBinomialCutoff{})
	copy(u.cutoffs[i+1:], u.cutoffs[i:])
	u.cutoffs[i] = subsetBinomialCutoff{binomial: binomial, choice: choice}
}

func (u *SubsetUnrank) SubsetAt(index int64) Value {
	values := NewValueVec(u.k)
	i := sort.Search(len(u.cutoffs), func(i int) bool { return u.cutoffs[i].binomial >= index })
	var entry *subsetBinomialCutoff
	if i > 0 {
		entry = &u.cutoffs[i-1]
	}
	choice := entry.choice
	y := u.k
	for ; choice >= 0 && u.k > 0; choice-- {
		binomial := u.memoizedBinomial(choice, y)
		if binomial <= index {
			index -= binomial
			y--
			values.Add(u.elems.At(choice))
		}
	}
	return NewSetEnumValueVec(values, false, u.owner.GetCostModel())
}

func (u *SubsetUnrank) memoizedBinomial(n, k int) int64 {
	if k == 0 || k == n {
		return 1
	}
	if k == 1 || k == n-1 {
		return int64(n)
	}
	if n == 0 || k > n {
		return 0
	}
	if ChoosePairToInt(n, k) < len(ChooseTable) {
		return Choose(n, k)
	}
	return u.partialPascalTable[(n-MaxChooseNum-1)*u.maxK+k-2]
}

func subsetUnrankForValue(owner Value, base Value, k int) (*SubsetUnrank, error) {
	set, err := tryToSetEnumValue(base)
	if err != nil {
		return nil, err
	}
	set.Normalize()
	n := BigSumChoose(set.Elems.Len(), k)
	if !n.IsInt64() {
		return nil, fmt.Errorf("BigInteger out of long range")
	}
	return NewSubsetUnrank(owner, k, n.Int64(), PascalTableUpTo(set.Elems.Len(), k), set.Elems, k), nil
}

func (v *SubsetValue) GetUnrank(k int) (*SubsetUnrank, error) {
	return subsetUnrankForValue(v, v.Set, k)
}

func (v *KSubsetValue) GetUnrank(k int) (*SubsetUnrank, error) {
	return subsetUnrankForValue(v, v.Set, k)
}

func numberOfKElements(base Value, k int) (int64, error) {
	size, err := base.Size()
	if err != nil {
		return 0, err
	}
	if k < 0 || size < k {
		return 0, nil
	}
	choice := k
	if size-k < choice {
		choice = size - k
	}
	count := BigChoose(size, choice)
	if count.BitLen() > 63 {
		return 0, NewIllegalArgumentException(fmt.Sprintf("k=%d and n=%d", k, size))
	}
	return count.Int64(), nil
}

func (v *SubsetValue) NumberOfKElements(k int) (int64, error) {
	return numberOfKElements(v.Set, k)
}

func (v *KSubsetValue) NumberOfKElements(k int) (int64, error) {
	return numberOfKElements(v.Set, k)
}

type emptySubsetEnumeration struct {
	owner Value
	done  bool
}

func (e *emptySubsetEnumeration) Reset()     { e.done = false }
func (e *emptySubsetEnumeration) Err() error { return nil }
func (e *emptySubsetEnumeration) NextElement() Value {
	if e.done {
		return nil
	}
	e.done = true
	return NewSetEnumValue(nil, true, e.owner.GetCostModel())
}

// KElementEnumeration mirrors the bounded KElementEnumerator, distinct from
// the unbounded internal combinations used by normalized powerset traversal.
type KElementEnumeration struct {
	owner       Value
	elems       *ValueVec
	numElements int
	k           int
	indices     []int
	count       int
}

func newKElementEnumeration(owner Value, base Value, k int) (*KElementEnumeration, error) {
	count, err := numberOfKElements(base, k)
	if err != nil {
		return nil, err
	}
	if count > math.MaxInt32 {
		return nil, NewIllegalArgumentException("Subset too large.")
	}
	set, err := tryToSetEnumValue(base)
	if err != nil {
		return nil, err
	}
	set.Normalize()
	e := &KElementEnumeration{owner: owner, elems: set.Elems, numElements: int(count), k: k, indices: make([]int, k)}
	e.Reset()
	return e, nil
}

func (v *SubsetValue) NewKElementEnumerator(k int) (*KElementEnumeration, error) {
	return newKElementEnumeration(v, v.Set, k)
}

func (v *KSubsetValue) NewKElementEnumerator(k int) (*KElementEnumeration, error) {
	return newKElementEnumeration(v, v.Set, k)
}

func (e *KElementEnumeration) Reset() {
	for i := range e.indices {
		e.indices[i] = i
	}
	e.count = 0
}

func (e *KElementEnumeration) NextElement() Value {
	if e.count >= e.numElements {
		return nil
	}
	e.count++
	values := NewValueVec(e.k)
	for _, index := range e.indices {
		values.Add(e.elems.At(index))
	}
	i := e.k - 1
	for i >= 0 && e.indices[i] == e.elems.Len()-e.k+i {
		i--
	}
	if i >= 0 {
		e.indices[i]++
		for j := i + 1; j < e.k; j++ {
			e.indices[j] = e.indices[j-1] + 1
		}
	}
	return NewSetEnumValueVec(values, true, e.owner.GetCostModel())
}

func (e *KElementEnumeration) Err() error { return nil }

func (e *KElementEnumeration) Sort() (*KElementEnumeration, error) {
	if err := e.elems.Sort(true); err != nil {
		return nil, err
	}
	return e, nil
}

func (e *KElementEnumeration) AsSet() *SetEnumValue {
	values := NewValueVec(e.numElements)
	for value := e.NextElement(); value != nil; value = e.NextElement() {
		values.Add(value)
	}
	return NewSetEnumValueVec(values, true)
}

func kElementsForValue(owner Value, base Value, k int) ValueEnumeration {
	if k < 0 {
		return EmptySet.Elements()
	}
	if k == 0 {
		return &emptySubsetEnumeration{owner: owner}
	}
	size, err := base.Size()
	if err != nil {
		return newErrorEnumeration(err)
	}
	if size < k {
		return EmptySet.Elements()
	}
	e, err := newKElementEnumeration(owner, base, k)
	if err != nil {
		return newErrorEnumeration(err)
	}
	return e
}

func (v *SubsetValue) KElements(k int) ValueEnumeration {
	return kElementsForValue(v, v.Set, k)
}

func (v *KSubsetValue) KElements(k int) ValueEnumeration {
	return kElementsForValue(v, v.Set, k)
}

func kSubsetForValue(owner Value, base Value, k int) (*SetEnumValue, error) {
	e := kElementsForValue(owner, base, k)
	if err := e.Err(); err != nil {
		return nil, err
	}
	if bounded, ok := e.(*KElementEnumeration); ok {
		return bounded.AsSet(), nil
	}
	return setEnumFromEnumeration(e, false)
}

func (v *SubsetValue) KSubset(k int) (*SetEnumValue, error) {
	return kSubsetForValue(v, v.Set, k)
}

func (v *KSubsetValue) KSubset(k int) (*SetEnumValue, error) {
	return kSubsetForValue(v, v.Set, k)
}

// Java RandomUnrank intentionally takes its modulus from the requested picks,
// keeps signed long overflow, and consumes an offset even for ranks not used.
type randomSubsetUnrank struct {
	*SubsetUnrank
	n, a, i int64
}

func newRandomSubsetUnrank(owner Value, k int, n int64, partialPascalTable []int64, elems *ValueVec, maxK int) *randomSubsetUnrank {
	rng := RandomEnumerableGenerator()
	u := NewSubsetUnrank(owner, k, n, partialPascalTable, elems, maxK)
	a := rng.NextLong()
	if a < 0 {
		a = -a // Java Math.abs leaves Long.MIN_VALUE negative.
	}
	return &randomSubsetUnrank{SubsetUnrank: u, n: n, a: a % n}
}

func (u *randomSubsetUnrank) randomSubset() Value {
	if u.i >= u.n {
		return nil
	}
	index := (34359738337*u.i + u.a) % u.n
	u.i++
	return u.SubsetAt(index)
}

func randomSetOfSubsetsUpTo(owner Value, base Value, requested, maxLength int) (*SetEnumValue, error) {
	set, err := tryToSetEnumValue(base)
	if err != nil {
		return nil, err
	}
	set.Normalize()
	counts := make([]int64, maxLength+1)
	counts[0] = 1
	sum := big.NewInt(1)
	for i := 1; i <= maxLength; i++ {
		count := BigChoose(set.Elems.Len(), i)
		if !count.IsInt64() {
			return nil, fmt.Errorf("BigInteger out of long range")
		}
		counts[i] = count.Int64()
		sum.Add(sum, count)
	}
	table := PascalTableUpTo(set.Elems.Len(), maxLength)
	values := NewValueVec(requested)
	for rank, count := range counts {
		n := subsetRankPickCount(count, sum, requested)
		if rank == len(counts)-1 {
			n = int64(requested - values.Len())
		}
		u := newRandomSubsetUnrank(owner, rank, n, table, set.Elems, maxLength)
		// Java evaluates randomSubset before testing the vector-size bound.
		for subset := u.randomSubset(); subset != nil && values.Len() < requested; subset = u.randomSubset() {
			values.Add(subset)
		}
	}
	return NewSetEnumValueVec(values, false, owner.GetCostModel()), nil
}

func subsetRankPickCount(count int64, sum *big.Int, requested int) int64 {
	// BigDecimal.divide(scale=32, HALF_DOWN), followed by multiplication,
	// max(ONE), truncation, and longValueExact. Keep the decimal rounding.
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(32), nil)
	numerator := new(big.Int).Mul(big.NewInt(count), scale)
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(numerator, sum, remainder)
	if new(big.Int).Lsh(remainder, 1).Cmp(sum) > 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	quotient.Mul(quotient, big.NewInt(int64(requested)))
	if quotient.Cmp(scale) < 0 {
		return 1
	}
	quotient.Quo(quotient, scale)
	if !quotient.IsInt64() {
		panic(fmt.Errorf("BigInteger out of long range"))
	}
	return quotient.Int64()
}

func (v *SubsetValue) GetRandomSetOfSubsetsUpTo(requested, maxLength int) (*SetEnumValue, error) {
	return randomSetOfSubsetsUpTo(v, v.Set, requested, maxLength)
}

func (v *KSubsetValue) GetRandomSetOfSubsetsUpTo(requested, maxLength int) (*SetEnumValue, error) {
	return randomSetOfSubsetsUpTo(v, v.Set, requested, maxLength)
}

func randomSetOfSubsetsWithProbability(owner Value, base Value, picks int, probability float64) (*SetEnumValue, error) {
	e, err := newCoinTossingSubsetEnumeration(base, owner.GetCostModel(), picks, probability)
	if err != nil {
		return nil, err
	}
	sets := newValueHashSet(int(javaDoubleToInt(float64(picks) * probability)))
	for value := e.NextElement(); value != nil; value = e.NextElement() {
		sets.add(value)
	}
	return NewSetEnumValueVec(sets.values, false, owner.GetCostModel()), nil
}

func (v *SubsetValue) GetRandomSetOfSubsets(picks int, probability float64) (*SetEnumValue, error) {
	return randomSetOfSubsetsWithProbability(v, v.Set, picks, probability)
}

func (v *KSubsetValue) GetRandomSetOfSubsets(picks int, probability float64) (*SetEnumValue, error) {
	return randomSetOfSubsetsWithProbability(v, v.Set, picks, probability)
}

// Lexicographic enumeration is the package-visible Java bitset path. Its
// constructor deliberately replaces the outer base before normalizing it.
type lexicographicSubsetEnumeration struct {
	owner      Value
	elems      *ValueVec
	descriptor *BitVector
}

func lexicographicSubsetElements(owner Value, base *Value) ValueEnumeration {
	set, err := tryToSetEnumValue(*base)
	if err != nil {
		return newErrorEnumeration(err)
	}
	if set == nil {
		*base = nil
	} else {
		*base = set
	}
	set.Normalize()
	e := &lexicographicSubsetEnumeration{owner: owner, elems: set.Elems}
	e.Reset()
	return e
}

func (v *SubsetValue) ElementsLexicographic() ValueEnumeration {
	return lexicographicSubsetElements(v, &v.Set)
}

func (v *KSubsetValue) ElementsLexicographic() ValueEnumeration {
	return lexicographicSubsetElements(v, &v.Set)
}

func (e *lexicographicSubsetEnumeration) Reset() {
	e.descriptor = NewBitVector(e.elems.Len())
}

func (e *lexicographicSubsetEnumeration) Err() error { return nil }

func (e *lexicographicSubsetEnumeration) NextElement() Value {
	if e.descriptor == nil {
		return nil
	}
	size := e.elems.Len()
	values := NewValueVec(e.descriptor.TrueCount())
	if size == 0 {
		e.descriptor = nil
	} else {
		for i := 0; i < size; i++ {
			if e.descriptor.Get(i) {
				values.Add(e.elems.At(i))
			}
		}
		for i := 0; i < size; i++ {
			if e.descriptor.Get(i) {
				e.descriptor.Reset(i)
				if i >= size-1 {
					e.descriptor = nil
					break
				}
			} else {
				e.descriptor.Set(i)
				break
			}
		}
	}
	e.owner.GetCostModel().incValueSecondary(int64(values.Len()))
	return NewSetEnumValueVec(values, true, e.owner.GetCostModel())
}
