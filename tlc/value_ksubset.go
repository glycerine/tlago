package tlc

import (
	"math"
	"math/big"
	"strings"
)

type KSubsetValue struct {
	BaseValue
	K         int
	Set       Value
	PSet      *SetEnumValue
	PSetDummy bool
}

func NewKSubsetValue(k int, set Value) *KSubsetValue {
	return &KSubsetValue{K: k, Set: set}
}

func (v *KSubsetValue) Kind() ValueKind    { return SubsetValueKind }
func (v *KSubsetValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *KSubsetValue) Compare(other Value) (int, error) {
	if o, ok := other.(*KSubsetValue); ok {
		vEmpty, err := v.hasNoElements()
		if err != nil {
			return 0, err
		}
		oEmpty, err := o.hasNoElements()
		if err != nil {
			return 0, err
		}
		if vEmpty || oEmpty {
			if vEmpty == oEmpty {
				return 0, nil
			}
			if vEmpty {
				return -1, nil
			}
			return 1, nil
		}
		if v.K == o.K {
			if v.K == 0 {
				return 0, nil
			}
			return v.Set.Compare(o.Set)
		}
		vFinite, err := v.Set.IsFinite()
		if err != nil {
			return 0, err
		}
		oFinite, err := o.Set.IsFinite()
		if err != nil {
			return 0, err
		}
		if vFinite && oFinite {
			vCount, err := v.count()
			if err != nil {
				return 0, err
			}
			oCount, err := o.count()
			if err != nil {
				return 0, err
			}
			if cmp := vCount.Cmp(oCount); cmp != 0 {
				return cmp, nil
			}
		}
		setEqual, err := v.Set.Equal(o.Set)
		if err != nil {
			return 0, err
		}
		if setEqual {
			return v.K - o.K, nil
		}
	} else if o, ok := other.(*SubsetValue); ok {
		empty, err := v.hasNoElements()
		if err != nil {
			return 0, err
		}
		if empty {
			return -1, nil
		}
		cmp, err := v.compareSubsetCardinality(o)
		if err != nil {
			return 0, err
		}
		if cmp != nil && *cmp != 0 {
			return *cmp, nil
		}
		if v.K == 0 {
			otherEmpty, err := IsEmptyValue(o.Set)
			if err != nil {
				return 0, err
			}
			if otherEmpty {
				return 0, nil
			}
		}
	}
	set, err := v.convertAndCache()
	if err != nil {
		return 0, err
	}
	return set.Compare(other)
}

func (v *KSubsetValue) Equal(other Value) (bool, error) {
	if o, ok := other.(*KSubsetValue); ok {
		vEmpty, err := v.hasNoElements()
		if err != nil {
			return false, err
		}
		oEmpty, err := o.hasNoElements()
		if err != nil {
			return false, err
		}
		if vEmpty || oEmpty {
			return vEmpty && oEmpty, nil
		}
		if v.K == 0 && o.K == 0 {
			return true, nil
		}
		if v.K != o.K {
			return false, nil
		}
		return v.Set.Equal(o.Set)
	}
	if _, ok := other.(*SubsetValue); ok {
		empty, err := v.hasNoElements()
		if err != nil {
			return false, err
		}
		if empty {
			return false, nil
		}
		subset := other.(*SubsetValue)
		cmp, err := v.compareSubsetCardinality(subset)
		if err != nil {
			return false, err
		}
		if cmp != nil && *cmp != 0 {
			return false, nil
		}
		if v.K == 0 {
			otherEmpty, err := IsEmptyValue(subset.Set)
			if err != nil {
				return false, err
			}
			if otherEmpty {
				return true, nil
			}
		}
	}
	set, err := v.convertAndCache()
	if err != nil {
		return false, err
	}
	return set.Equal(other)
}

func (v *KSubsetValue) compareSubsetCardinality(other *SubsetValue) (*int, error) {
	vFinite, err := v.Set.IsFinite()
	if err != nil {
		return nil, err
	}
	otherFinite, err := other.Set.IsFinite()
	if err != nil {
		return nil, err
	}
	if !vFinite || !otherFinite {
		return nil, nil
	}
	count, err := v.count()
	if err != nil {
		return nil, err
	}
	otherSize, err := other.Set.Size()
	if err != nil {
		return nil, err
	}
	otherCount := new(big.Int).Lsh(big.NewInt(1), uint(otherSize))
	cmp := count.Cmp(otherCount)
	return &cmp, nil
}

func (v *KSubsetValue) Member(elem Value) (bool, error) {
	empty, err := v.hasNoElements()
	if err != nil || empty {
		return false, err
	}
	elemSize, err := elem.Size()
	if err != nil {
		return false, err
	}
	if elemSize != v.K {
		return false, nil
	}
	return NewSubsetValue(v.Set).Member(elem)
}

func (v *KSubsetValue) IsFinite() (bool, error) {
	if v.K <= 0 {
		return true, nil
	}
	return v.Set.IsFinite()
}

func (v *KSubsetValue) Size() (int, error) {
	count, err := v.count()
	if err != nil {
		return 0, err
	}
	if !count.IsInt64() || count.Int64() > math.MaxInt32 {
		baseSize, _ := v.Set.Size()
		return 0, v.unsupported("k=%d and n=%d", v.K, baseSize)
	}
	return int(count.Int64()), nil
}

func (v *KSubsetValue) IsNormalized() bool {
	return v.PSet != nil && !v.PSetDummy && v.PSet.IsNormalized()
}

func (v *KSubsetValue) Normalize() Value {
	if v.PSet != nil && !v.PSetDummy {
		v.PSet.Normalize()
	} else {
		v.Set.Normalize()
	}
	return v
}

func (v *KSubsetValue) DeepNormalize() {
	v.Set.DeepNormalize()
	if v.PSet == nil {
		v.PSetDummy = true
	} else if !v.PSetDummy {
		v.PSet.DeepNormalize()
	}
}

func (v *KSubsetValue) IsDefined() bool {
	return v.Set.IsDefined()
}

func (v *KSubsetValue) DeepCopy() Value { return v }

func (v *KSubsetValue) FingerPrint(fp uint64) uint64 {
	set, err := v.convertAndCache()
	if err != nil {
		return fp
	}
	return set.FingerPrint(fp)
}

func (v *KSubsetValue) Permute(perm *MVPerm) Value {
	set, err := v.convertAndCache()
	if err != nil {
		return v
	}
	return set.Permute(perm)
}

func (v *KSubsetValue) TakeExcept(ex ValueExcept) (Value, error) {
	return takeExceptOnSet(v, ex)
}

func (v *KSubsetValue) TakeExcepts(exs []ValueExcept) (Value, error) {
	return takeExceptsOnSet(v, exs)
}

func (v *KSubsetValue) ToSetEnum() (*SetEnumValue, error) {
	if v.PSet != nil && !v.PSetDummy {
		return v.PSet, nil
	}
	return setEnumFromEnumeration(v.Elements(), true)
}

func (v *KSubsetValue) convertAndCache() (*SetEnumValue, error) {
	if v.PSetDummy {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		set.DeepNormalize()
		v.PSet = set
		v.PSetDummy = false
	} else if v.PSet == nil {
		set, err := v.ToSetEnum()
		if err != nil {
			return nil, err
		}
		v.PSet = set
		v.PSetDummy = false
	}
	return v.PSet, nil
}

func (v *KSubsetValue) Elements() ValueEnumeration {
	empty, err := v.hasNoElements()
	if err != nil {
		return newErrorEnumeration(err)
	}
	if empty {
		return EmptySet.Elements()
	}
	if v.K == 0 {
		return &singleValueEnumeration{value: EmptySet}
	}
	set, err := toSetEnumValue(v.Set)
	if err != nil {
		return newErrorEnumeration(err)
	}
	if _, err := set.normalizeSet(); err != nil {
		return newErrorEnumeration(err)
	}
	return newKSubsetEnumeration(set.Elems, v.K)
}

func (v *KSubsetValue) String() string {
	empty, err := v.hasNoElements()
	if err == nil && empty {
		return "{}"
	}
	expand := Globals.Expand
	if expand {
		size, err := v.Size()
		expand = err == nil && size < 64
	}
	if expand {
		if set, err := v.ToSetEnum(); err == nil {
			return set.String()
		}
	}
	var b strings.Builder
	b.WriteString("{s \\in SUBSET (")
	b.WriteString(v.Set.String())
	b.WriteString(") : Cardinality(s) = ")
	b.WriteString(NewIntValue(int32(v.K)).String())
	b.WriteString("}")
	return b.String()
}

func (v *KSubsetValue) count() (*big.Int, error) {
	if v.K < 0 {
		return big.NewInt(0), nil
	}
	if v.K == 0 {
		return big.NewInt(1), nil
	}
	n, err := v.Set.Size()
	if err != nil {
		return nil, err
	}
	if v.K > n {
		return big.NewInt(0), nil
	}
	k := v.K
	if n-k < k {
		k = n - k
	}
	return BigChoose(n, k), nil
}

func (v *KSubsetValue) hasNoElements() (bool, error) {
	if v.K < 0 {
		return true, nil
	}
	finite, err := v.Set.IsFinite()
	if err != nil {
		return false, err
	}
	if finite {
		size, err := v.Set.Size()
		if err != nil {
			return false, err
		}
		return size < v.K, nil
	}
	return false, nil
}

type kSubsetEnumeration struct {
	elems   *ValueVec
	k       int
	indices []int
	done    bool
	err     error
}

func newKSubsetEnumeration(elems *ValueVec, k int) *kSubsetEnumeration {
	out := &kSubsetEnumeration{elems: elems, k: k}
	out.Reset()
	return out
}

func (e *kSubsetEnumeration) Reset() {
	if e.k < 0 || e.k > e.elems.Len() {
		e.indices = nil
		e.done = true
		return
	}
	e.indices = make([]int, e.k)
	for i := range e.indices {
		e.indices[i] = i
	}
	e.done = false
}

func (e *kSubsetEnumeration) NextElement() Value {
	if e.done || e.err != nil {
		return nil
	}
	vals := NewValueVec(e.k)
	for _, idx := range e.indices {
		vals.Add(e.elems.At(idx))
	}
	e.advance()
	return NewSetEnumValueVec(vals, true)
}

func (e *kSubsetEnumeration) Err() error { return e.err }

func (e *kSubsetEnumeration) advance() {
	n := e.elems.Len()
	for i := e.k - 1; i >= 0; i-- {
		if e.indices[i] < n-e.k+i {
			e.indices[i]++
			for j := i + 1; j < e.k; j++ {
				e.indices[j] = e.indices[j-1] + 1
			}
			return
		}
	}
	e.done = true
}
