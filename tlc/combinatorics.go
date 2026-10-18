package tlc

import (
	"math/big"
	"strings"
)

const (
	MaxChooseNum    = 62
	ChooseTableSize = (MaxChooseNum-3)*(MaxChooseNum-4)/2 + MaxChooseNum - 3
)

var (
	ChooseTable    [ChooseTableSize]int64
	SumChooseTable [ChooseTableSize]int64
)

func init() {
	n := 4
	m := 2
	sum := int64(5)
	for i := 0; i < ChooseTableSize; i++ {
		ChooseTable[i] = Choose(n-1, m) + Choose(n-1, m-1)
		sum += ChooseTable[i]
		SumChooseTable[i] = sum
		if n == m+2 {
			n++
			m = 2
			sum = int64(1 + n)
		} else {
			m++
		}
	}
}

func Choose(n int, m int) int64 {
	n = int(int32(n))
	m = int(int32(m))
	if n < 0 || m < 0 {
		panic(NewTLCRuntimeException(ECTLCChooseArgumentsWrong, "choose"))
	}
	if m == 0 || m == n {
		return 1
	}
	if m == 1 || m == n-1 {
		return int64(n)
	}
	if n == 0 || m > n {
		return 0
	}
	j := ChoosePairToInt(n, m)
	if j < ChooseTableSize {
		return ChooseTable[j]
	}
	return Binomial(n, m)
}

func SumChoose(n int, m int) int64 {
	n = int(int32(n))
	m = int(int32(m))
	if m < 0 || n < 0 || n < m {
		panic(NewTLCRuntimeException(ECTLCChooseArgumentsWrong, "sumChoose"))
	}
	if m == 0 {
		return 1
	}
	if m == n {
		return int64(1) << uint(n&63)
	}
	if m == 1 {
		return int64(n)
	}
	if m == n-1 {
		return (int64(2) << uint(n&63)) - int64(n)
	}
	j := ChoosePairToInt(n, m)
	if j < ChooseTableSize {
		return SumChooseTable[j]
	}
	panic(NewTLCRuntimeException(ECTLCChooseUpperBound, fmtInt(MaxChooseNum)))
}

func ChoosePairToInt(n int, m int) int {
	n32, m32 := int32(n), int32(m)
	return int(((n32-3)*(n32-4))/2 + m32 - 2)
}

func ToNum(bases []*big.Int, digits []*big.Int, length int) *big.Int {
	length = int(int32(length))
	if bases == nil {
		panic(NewNullPointerException())
	}
	if len(bases) < length || length <= 0 {
		panic(NewTLCRuntimeException(ECSystemIndexError))
	}
	num := new(big.Int).Set(combinatoricsBigIntAt(digits, length-1))
	for i := length - 2; i >= 0; i-- {
		num.Mul(num, combinatoricsBigIntAt(bases, i))
		num.Add(num, combinatoricsBigIntAt(digits, i))
	}
	return num
}

func ToNumAll(bases []*big.Int, digits []*big.Int) *big.Int {
	return ToNum(bases, digits, len(bases))
}

func ToSeq(bases []*big.Int, n *big.Int, length int) []*big.Int {
	length = int(int32(length))
	if bases == nil {
		panic(NewNullPointerException())
	}
	if len(bases) < length || length == 0 {
		panic(NewTLCRuntimeException(ECSystemIndexError))
	}
	if length < 0 {
		panic(NewNegativeArraySizeException(fmtInt(length)))
	}
	nlist := make([]*big.Int, length)
	if n == nil {
		panic(NewNullPointerException())
	}
	num := new(big.Int).Set(n)
	base := combinatoricsBigIntAt(bases, 0)
	if base.Sign() == 0 {
		panic(NewArithmeticException("BigInteger divide by zero"))
	}
	nlist[0] = new(big.Int).Rem(num, base)
	for i := 1; i < length; i++ {
		num.Quo(num, base)
		base = combinatoricsBigIntAt(bases, i)
		if base.Sign() == 0 {
			panic(NewArithmeticException("BigInteger divide by zero"))
		}
		nlist[i] = new(big.Int).Rem(num, base)
	}
	return nlist
}

func combinatoricsBigIntAt(values []*big.Int, index int) *big.Int {
	if values == nil {
		panic(NewNullPointerException())
	}
	if index < 0 || index >= len(values) {
		panic(NewArrayIndexOutOfBoundsException(index, len(values)))
	}
	if values[index] == nil {
		panic(NewNullPointerException())
	}
	return values[index]
}

func ToSeqAll(bases []*big.Int, n *big.Int) []*big.Int {
	return ToSeq(bases, n, len(bases))
}

func Fact(n int) *big.Int {
	n = int(int32(n))
	result := big.NewInt(1)
	for i := int32(n); i > 1; i-- {
		result.Mul(result, big.NewInt(int64(i)))
	}
	return result
}

func BigChoose(n int, m int) *big.Int {
	n = int(int32(n))
	m = int(int32(m))
	if n < MaxChooseNum && m < MaxChooseNum {
		return big.NewInt(Choose(n, m))
	}
	binomial := big.NewInt(1)
	for i, j := int32(1), int32(n); i <= int32(m); i, j = i+1, j-1 {
		binomial.Mul(binomial, big.NewInt(int64(j)))
		binomial.Div(binomial, big.NewInt(int64(i)))
	}
	return binomial
}

func CombinatoricsFactorial(n Value) (Value, error) {
	iv, ok := n.(*IntValue)
	if !ok || iv.ownedVal() < 0 {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "factorial", "natural number", ValuesPPR(n))
	}
	result := Fact(int(iv.ownedVal()))
	out, ok := bigIntToExactInt32(result)
	if !ok {
		return nil, newTLCErrorCode(ECTLCModuleOverflow, n.String())
	}
	return NewIntValue(out), nil
}

func CombinatoricsChoose(n Value, k Value) (Value, error) {
	nv, ok := n.(*IntValue)
	if !ok || nv.ownedVal() < 0 {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "choose", "natural number", ValuesPPR(n))
	}
	kv, ok := k.(*IntValue)
	if !ok || kv.ownedVal() < 0 {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "choose", "natural number", ValuesPPR(k))
	}
	result := Choose(int(nv.ownedVal()), int(kv.ownedVal()))
	if int64(int32(result)) != result {
		return nil, newTLCErrorCode(ECTLCModuleOverflow, n.String()+"choose"+k.String())
	}
	return NewIntValue(int32(result)), nil
}

func bigIntToExactInt32(value *big.Int) (int32, bool) {
	if value == nil || !value.IsInt64() {
		return 0, false
	}
	out := value.Int64()
	if int64(int32(out)) != out {
		return 0, false
	}
	return int32(out), true
}

func SlowBigChoose(n int, m int) *big.Int {
	n = int(int32(n))
	m = int(int32(m))
	num := Fact(n)
	denom := Fact(int(int32(n - m)))
	denom.Mul(denom, Fact(m))
	return num.Div(num, denom)
}

func BigSumChoose(n int, m int) *big.Int {
	n = int(int32(n))
	m = int(int32(m))
	result := new(big.Int)
	if n/2 >= m {
		for i := int32(0); i <= int32(m); i++ {
			result.Add(result, BigChoose(n, int(i)))
		}
		return result
	}
	result.SetInt64(1)
	if n >= 0 {
		result.Lsh(result, uint(n))
	} else {
		result.Rsh(result, uint(-int64(n)))
	}
	for i := int32(m) + 1; i <= int32(n); i++ {
		result.Sub(result, BigChoose(n, int(i)))
	}
	return result
}

func PrintBigInts(values []*big.Int) string {
	var b strings.Builder
	for _, value := range values {
		b.WriteString(value.String())
		b.WriteString(", ")
	}
	return b.String()
}

func Binomial(n int, k int) int64 {
	n32, k32 := int32(n), int32(k)
	if k32 > n32 {
		return 0
	}
	if k32 > n32-k32 {
		k32 = n32 - k32
	}
	binomial := int64(1)
	for i, m := int32(1), n32; i <= k32; i, m = i+1, m-1 {
		binomial = binomial * int64(m) / int64(i)
	}
	return binomial
}

func PascalTableUpTo(maxN int, maxK int) []int64 {
	maxN = int(int32(maxN))
	maxK = int(int32(maxK))
	if maxN <= MaxChooseNum {
		return []int64{}
	}
	length := int(int32(maxN-MaxChooseNum) * int32(maxK-1))
	if length < 0 {
		panic(NewNegativeArraySizeException(fmtInt(length)))
	}
	ppt := make([]int64, length)
	idx := 0
	i := MaxChooseNum + 1
	for j := 2; j <= maxK; j++ {
		ppt[idx] = Choose(i, j)
		idx++
	}
	k := maxK - 1
	for j := 1; j < maxN-MaxChooseNum; j++ {
		for l := 0; l < k; l++ {
			if l == 0 {
				ppt[idx] = int64(i) + ppt[idx-k]
				i++
			} else {
				ppt[idx] = ppt[idx-k] + ppt[idx-k-1]
			}
			idx++
		}
	}
	return ppt
}
