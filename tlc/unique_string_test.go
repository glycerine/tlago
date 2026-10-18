package tlc

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestUniqueStringJavaTextMethods(t *testing.T) {
	value := UniqueStringOf("A😀B")
	if got := value.Length(); got != 4 {
		t.Fatalf("length = %d, want 4 UTF-16 units", got)
	}
	if got := value.HashCode(); got != 56896350 {
		t.Fatalf("hashCode = %d, want 56896350", got)
	}
	alternatePair := javaStringFromUTF16([]uint16{'A', 0xd83d}) + javaStringFromUTF16([]uint16{0xde00, 'B'})
	if !value.EqualString("A😀B") || !value.EqualString(alternatePair) || value.EqualString("A😀") {
		t.Fatal("String equality does not match the represented Java String")
	}
	if UniqueStringOf(alternatePair) != value {
		t.Fatal("interner distinguished equivalent Java UTF-16 strings")
	}

	high := javaStringFromUTF16([]uint16{'A', 0xd83d})
	for _, tc := range []struct {
		prefix string
		want   bool
	}{
		{"", true},
		{"A", true},
		{high, true},
		{"A😀", true},
		{"A😀B", true},
		{"A😀BC", false},
	} {
		if got := value.StartsWith(tc.prefix); got != tc.want {
			t.Errorf("startsWith(%x) = %t, want %t", javaStringUTF16(tc.prefix), got, tc.want)
		}
	}

	for _, tc := range []struct {
		begin int
		want  []uint16
	}{
		{0, []uint16{'A', 0xd83d, 0xde00, 'B'}},
		{1, []uint16{0xd83d, 0xde00, 'B'}},
		{2, []uint16{0xde00, 'B'}},
		{3, []uint16{'B'}},
		{4, []uint16{}},
	} {
		if got := javaStringUTF16(value.Substring(tc.begin)); !equalUTF16Units(got, tc.want) {
			t.Errorf("substring(%d) units = %x, want %x", tc.begin, got, tc.want)
		}
	}

	loneHigh := UniqueStringOf(javaStringFromUTF16([]uint16{'X', 0xd83d}))
	loneLow := UniqueStringOf(javaStringFromUTF16([]uint16{0xde00, 'Y'}))
	if got := javaStringHashCode(loneHigh.String()); got != 58085 || got != loneHigh.HashCode() {
		t.Fatalf("unpaired-surrogate hashCode = %d, want 58085", got)
	}
	joined := loneHigh.Concat(loneLow)
	if got := javaStringUTF16(joined.String()); !equalUTF16Units(got, []uint16{'X', 0xd83d, 0xde00, 'Y'}) {
		t.Fatalf("concat(UniqueString) units = %x", got)
	}
	if loneHigh.ConcatString(loneLow.String()) != joined {
		t.Fatal("concat(String) did not intern the same Java String")
	}
	if got := value.ConcatNullableString(nil).String(); got != "A😀Bnull" {
		t.Fatalf("concat((String) null) = %q, want %q", got, "A😀Bnull")
	}
	if value.EqualNullableString(nil) {
		t.Fatal("equals((String) null) returned true")
	}
}

func TestUniqueStringJavaSubstringBounds(t *testing.T) {
	value := UniqueStringOf("A😀B")
	for _, begin := range []int{-1, 5} {
		func() {
			defer func() {
				failure, ok := recover().(*StringIndexOutOfBoundsException)
				want := "Range [" + fmt.Sprint(begin) + ", 4) out of bounds for length 4"
				if !ok || failure.Error() != want {
					t.Fatalf("substring(%d) panic = %T %v, want StringIndexOutOfBoundsException %q", begin, failure, failure, want)
				}
			}()
			_ = value.Substring(begin)
		}()
	}
}

func equalUTF16Units(left, right []uint16) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func TestUniqueStringCommitCheckpointErrorUsesJavaText(t *testing.T) {
	dir := t.TempDir()
	err := CommitChkptUniqueStrings(dir)
	if err == nil {
		t.Fatalf("CommitChkptUniqueStrings without tmp file returned nil")
	}
	want := "InternTable.commitChkpt: cannot delete " + filepath.Join(dir, "vars.chkpt")
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("CommitChkptUniqueStrings error = %q, want containing %q", err.Error(), want)
	}
}
