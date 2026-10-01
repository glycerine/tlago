package tlc

import (
	"math/big"
	"testing"
)

func TestCombinatoricsChooseMatchesBinomialTable(t *testing.T) {
	for n := 0; n < MaxChooseNum; n++ {
		for k := 0; k < MaxChooseNum; k++ {
			if got, want := Choose(n, k), Binomial(n, k); got != want {
				t.Fatalf("Choose(%d,%d) = %d, want %d", n, k, got, want)
			}
		}
	}
}

func TestCombinatoricsBigChooseMatchesJavaKnownValues(t *testing.T) {
	cases := []struct {
		n      int
		k      int
		bitLen int
		dec    string
	}{
		{50, 1, 6, "50"},
		{50, 10, 34, "10272278170"},
		{50, 20, 46, "47129212243960"},
		{50, 30, 46, "47129212243960"},
		{400, 1, 9, "400"},
		{400, 50, 214, "17035900270730601418919867558071677342938596450600561760371485120"},
		{400, 100, 321, "2241854791554337561923210387201698554845411177476295990399942258896013007429693894018935107174320"},
		{400, 200, 396, "102952500135414432972975880320401986757210925381077648234849059575923332372651958598336595518976492951564048597506774120"},
	}
	for _, tc := range cases {
		got := BigChoose(tc.n, tc.k)
		if got.BitLen() != tc.bitLen {
			t.Fatalf("BigChoose(%d,%d).BitLen = %d, want %d", tc.n, tc.k, got.BitLen(), tc.bitLen)
		}
		if got.String() != tc.dec {
			t.Fatalf("BigChoose(%d,%d) = %s, want %s", tc.n, tc.k, got, tc.dec)
		}
	}
}

func TestCombinatoricsBigChooseMatchesSlowBeyondTable(t *testing.T) {
	for n := MaxChooseNum + 1; n < MaxChooseNum+12; n++ {
		for k := MaxChooseNum + 1; k < MaxChooseNum+12; k++ {
			got := BigChoose(n, k)
			want := SlowBigChoose(n, k)
			if got.Cmp(want) != 0 {
				t.Fatalf("BigChoose(%d,%d) = %s, want %s", n, k, got, want)
			}
		}
	}
}

func TestCombinatoricsMixedRadixRoundTrip(t *testing.T) {
	bases := []*big.Int{big.NewInt(5), big.NewInt(7), big.NewInt(11), big.NewInt(13)}
	digits := []*big.Int{big.NewInt(4), big.NewInt(6), big.NewInt(10), big.NewInt(12)}
	num := ToNum(bases, digits, len(digits))
	got := ToSeq(bases, num, len(digits))
	for i := range digits {
		if got[i].Cmp(digits[i]) != 0 {
			t.Fatalf("digit %d = %s, want %s", i, got[i], digits[i])
		}
	}
}

func TestCombinatoricsBigSumChooseAndPascalExtension(t *testing.T) {
	if got, want := BigSumChoose(6, 2), big.NewInt(22); got.Cmp(want) != 0 {
		t.Fatalf("BigSumChoose(6,2) = %s, want %s", got, want)
	}
	if got, want := BigSumChoose(6, 4), big.NewInt(57); got.Cmp(want) != 0 {
		t.Fatalf("BigSumChoose(6,4) = %s, want %s", got, want)
	}
	extension := PascalTableUpTo(MaxChooseNum+2, 4)
	if len(extension) != 6 {
		t.Fatalf("PascalTableUpTo length = %d, want 6", len(extension))
	}
	want := []int64{
		Choose(MaxChooseNum+1, 2),
		Choose(MaxChooseNum+1, 3),
		Choose(MaxChooseNum+1, 4),
		Choose(MaxChooseNum+2, 2),
		Choose(MaxChooseNum+2, 3),
		Choose(MaxChooseNum+2, 4),
	}
	for i := range want {
		if extension[i] != want[i] {
			t.Fatalf("PascalTableUpTo[%d] = %d, want %d", i, extension[i], want[i])
		}
	}
}

func TestCombinatoricsCommunityModuleOverrides(t *testing.T) {
	fact, err := CombinatoricsFactorial(NewIntValue(5))
	if err != nil || fact.(*IntValue).Val != 120 {
		t.Fatalf("factorial(5) = %v/%v, want 120/nil", fact, err)
	}
	choose, err := CombinatoricsChoose(NewIntValue(6), NewIntValue(2))
	if err != nil || choose.(*IntValue).Val != 15 {
		t.Fatalf("choose(6,2) = %v/%v, want 15/nil", choose, err)
	}
	if value, err := CombinatoricsFactorial(IntNegOne); err == nil {
		t.Fatalf("factorial(-1) = %v/nil, want argument error", value)
	}
	if value, err := CombinatoricsChoose(NewStringValue("n"), IntOne); err == nil {
		t.Fatalf("choose(\"n\",1) = %v/nil, want argument error", value)
	}
	if value, err := CombinatoricsFactorial(NewIntValue(13)); err == nil {
		t.Fatalf("factorial(13) = %v/nil, want int32 overflow error", value)
	}
	if value, err := CombinatoricsChoose(NewIntValue(50), NewIntValue(10)); err == nil {
		t.Fatalf("choose(50,10) = %v/nil, want int32 overflow error", value)
	}
}
