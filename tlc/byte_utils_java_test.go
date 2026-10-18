package tlc

import (
	"fmt"
	"math/big"
	"os"
	"reflect"
	"testing"
	"time"
)

const javaByteUtilsArraySize = 10000
const javaByteUtilsBits = 1000

type javaByteUtilsFixture struct {
	files  [2]string
	arrays [5][]*big.Int
}

func javaByteUtilsSetUp(t *testing.T) *javaByteUtilsFixture {
	t.Helper()
	f := &javaByteUtilsFixture{}
	for i := range f.arrays {
		f.arrays[i] = make([]*big.Int, javaByteUtilsArraySize)
	}
	for i, prefix := range []string{"ByteUtilsTestA", "ByteUtilsTestB"} {
		file, err := os.CreateTemp(t.TempDir(), prefix+"*.tmp")
		if err != nil {
			t.Fatal(err)
		}
		f.files[i] = file.Name()
		if err = file.Close(); err != nil {
			t.Fatal(err)
		}
	}
	return f
}
func javaByteUtilsOpen(t *testing.T, path string, write bool) *os.File {
	t.Helper()
	var file *os.File
	var err error
	if write {
		file, err = os.Create(path)
	} else {
		file, err = os.Open(path)
	}
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = file.Close() })
	return file
}
func javaByteUtilsIO(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
func javaByteUtilsInitialize(a, b []*big.Int) {
	r := NewJavaRandomDefault()
	for i := 0; i < javaByteUtilsArraySize; i++ {
		a[i] = NewBigIntRandom(javaByteUtilsBits, r).Value()
		b[i] = a[i]
	}
}
func javaByteUtilsCompareArrays(a, b, c []*big.Int, half int) {
	// The original methods print mismatches; they contain no assertEquals.
	for j := 0; j < half; j++ {
		if a[j].Cmp(b[j]) != 0 {
			fmt.Printf("A[%d] :%s   B[%d]: %s\n", j, a[j], j, b[j])
		}
	}
	for j := half; j < javaByteUtilsArraySize; j++ {
		if a[j].Cmp(c[j-half]) != 0 {
			fmt.Printf("A[%d] :%s   C[%d]: %s\n", j, a[j], j-half, c[j-half])
		}
	}
}

// All six original ByteUtilsTest methods, including their complete 10,000-value
// exercises, unseeded Java Random instances, console diagnostics, timing reports,
// source file operations, 1000-bit BigInt arrays and append IOException catch.
func TestJavaByteUtils(t *testing.T) {
	labels := []string{"IntToByteArray", "WriteInt, ReadInt", "longToByteArray", "WriteLong, ReadLong", "Write, Read", "Append"}
	for index, label := range labels {
		t.Run(fmt.Sprintf("test%d", index+1), func(t *testing.T) {
			fixture := javaByteUtilsSetUp(t)
			start := time.Now().UnixMilli()
			switch index {
			case 0:
				r := NewJavaRandomDefault()
				for j := 0; j < 10000; j++ {
					i := r.NextInt()
					b := IntToByteArray(i)
					if i != ByteArrayToInt(b) || len(b) != 4 {
						fmt.Printf("i :%d    byte :[B@%x    i: %d    size: %d\n", i, uint32(reflect.ValueOf(b).Pointer()), ByteArrayToInt(b), len(b))
					}
				}
			case 1:
				out := javaByteUtilsOpen(t, fixture.files[0], true)
				a := make([]int32, 10000)
				r := NewJavaRandomDefault()
				for j := 0; j < 10000; j++ {
					a[j] = r.NextInt()
					javaByteUtilsIO(t, WriteInt(out, a[j]))
				}
				// FileOutputStream.flush is a no-op; os.File writes are unbuffered.
				javaByteUtilsIO(t, out.Close())
				in := javaByteUtilsOpen(t, fixture.files[0], false)
				for j := 0; j < 10000; j++ {
					i, err := ReadInt(in)
					javaByteUtilsIO(t, err)
					if i != a[j] {
						fmt.Printf("i :%d   A[j]: %d\n", i, a[j])
					}
				}
			case 2:
				r := NewJavaRandomDefault()
				for j := 0; j < 10000; j++ {
					i := r.NextLong()
					b := LongToByteArray(i)
					if i != ByteArrayToLong(b) || len(b) != 8 {
						fmt.Printf("i :%d    byte :[B@%x    i: %d    size: %d\n", i, uint32(reflect.ValueOf(b).Pointer()), ByteArrayToLong(b), len(b))
					}
				}
			case 3:
				out := javaByteUtilsOpen(t, fixture.files[0], true)
				a := make([]int64, 10000)
				r := NewJavaRandomDefault()
				for j := 0; j < 10000; j++ {
					a[j] = r.NextLong()
					javaByteUtilsIO(t, WriteLong(out, a[j]))
				}
				javaByteUtilsIO(t, out.Close())
				in := javaByteUtilsOpen(t, fixture.files[0], false)
				for j := 0; j < 10000; j++ {
					i, err := ReadLong(in)
					javaByteUtilsIO(t, err)
					if i != a[j] {
						fmt.Printf("i :%d   A[j]: %d\n", i, a[j])
					}
				}
			case 4, 5:
				out := javaByteUtilsOpen(t, fixture.files[0], true)
				a, b := make([]*big.Int, javaByteUtilsArraySize), make([]*big.Int, javaByteUtilsArraySize)
				javaByteUtilsInitialize(a, b)
				half := (javaByteUtilsArraySize - 1) / 2
				javaByteUtilsIO(t, WriteSizeArrayOfSizeBigInts(out, a, 0, half-1))
				javaByteUtilsIO(t, WriteSizeArrayOfSizeBigInts(out, a, half, javaByteUtilsArraySize-1))
				javaByteUtilsIO(t, out.Close())
				in := javaByteUtilsOpen(t, fixture.files[0], false)
				if index == 5 {
					out = javaByteUtilsOpen(t, fixture.files[1], true)
					// A single original try block contains all three append calls. Only an
					// actual IOException is caught; no requirement for an exception is added.
					for i := 0; i < 3; i++ {
						if err := AppendSizeByteArray(in, out); err != nil {
							if !isJavaIOException(err) {
								panic(err)
							}
							break
						}
					}
					javaByteUtilsIO(t, in.Close())
					javaByteUtilsIO(t, out.Close())
					in = javaByteUtilsOpen(t, fixture.files[1], false)
				}
				var err error
				b, err = ReadSizeArrayOfSizeBigInts(in)
				javaByteUtilsIO(t, err)
				_, err = ReadInt(in)
				javaByteUtilsIO(t, err)
				c := ReadArrayOfSizeBigInts(in)
				javaByteUtilsCompareArrays(a, b, c, half)
			}
			end := time.Now().UnixMilli()
			fmt.Printf("Testing %s took %dms\n", label, end-start)
		})
	}
}
