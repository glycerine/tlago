/*******************************************************************************
 * Copyright (c) 2016 Microsoft Research. All rights reserved.
 * Copyright (c) 2024, Oracle and/or its affiliates.
 *
 * The MIT License (MIT)
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
 * of the Software, and to permit persons to whom the Software is furnished to do
 * so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN
 * AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
 * WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 *
 * Contributors:
 *   Markus Alexander Kuppe - initial API and implementation
 ******************************************************************************/
package tlc

import (
	"io"
	"os"
	"strings"
	"testing"
)

// These adapters represent RandomAccessFile's throw and -1-return boundaries;
// all file operations still execute the production BufferedRandomAccessFile.
type javaBRAF struct{ *BufferedRandomAccessFile }

func javaBRAFThrow(err error) {
	if err != nil {
		panic(err)
	}
}
func javaBRAFTemp(t *testing.T, prefix string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), prefix+"*.bin")
	javaBRAFThrow(err)
	javaBRAFThrow(f.Close())
	return f.Name()
}
func javaBRAFOpen(name, mode string) *javaBRAF {
	f, err := NewBufferedRandomAccessFile(name, mode)
	javaBRAFThrow(err)
	return &javaBRAF{f}
}
func (f *javaBRAF) close()            { javaBRAFThrow(f.Close()) }
func (f *javaBRAF) seek(n int64)      { javaBRAFThrow(f.Seek(n)) }
func (f *javaBRAF) setLength(n int64) { javaBRAFThrow(f.SetLength(n)) }
func (f *javaBRAF) length() int64     { n, e := f.Length(); javaBRAFThrow(e); return n }
func (f *javaBRAF) pointer() int64    { n, e := f.GetFilePointer(); javaBRAFThrow(e); return n }
func (f *javaBRAF) read() int64       { n, e := f.ReadByteValue(); javaBRAFThrow(e); return int64(n) }
func (f *javaBRAF) readArray(b []byte, offset, length int) int64 {
	n, e := f.Read(b[offset : offset+length])
	if e == io.EOF {
		return -1
	}
	javaBRAFThrow(e)
	return int64(n)
}
func (f *javaBRAF) readInt() int32  { n, e := f.ReadInt(); javaBRAFThrow(e); return n }
func (f *javaBRAF) readLong() int64 { n, e := f.ReadLong(); javaBRAFThrow(e); return n }
func (f *javaBRAF) write(n int)     { javaBRAFThrow(f.WriteByteValue(n)) }
func (f *javaBRAF) writeArray(b []byte, offset, length int) {
	javaBRAFThrow(f.WriteFull(b[offset : offset+length]))
}
func (f *javaBRAF) writeInt(n int32)  { javaBRAFThrow(f.WriteInt(n)) }
func (f *javaBRAF) writeLong(n int64) { javaBRAFThrow(f.WriteLong(n)) }
func (f *javaBRAF) flush()            { javaBRAFThrow(f.Flush()) }
func (f *javaBRAF) invalidate()       { javaBRAFThrow(f.InvalidateBufferedData()) }
func (f *javaBRAF) reset()            { javaBRAFThrow(f.Reset()) }
func javaBRAFBytes(b []int8) []byte {
	r := make([]byte, len(b))
	for i, v := range b {
		r[i] = byte(v)
	}
	return r
}
func javaBRAFEquals(t *testing.T, want, got int64) {
	t.Helper()
	if want != got {
		t.Fatalf("got %d, want %d", got, want)
	}
}

// The two seek/no-length tests accept IOException, but do not require a throw.
func javaBRAFIsIOException(e error) bool {
	_, represented := e.(interface{ GetMessage() *string })
	return represented && isJavaIOException(e)
}
func javaBRAFAcceptIOException(body func()) {
	defer func() {
		if v := recover(); v != nil {
			if e, ok := v.(error); !ok || !javaBRAFIsIOException(e) {
				panic(v)
			}
		}
	}()
	body()
}
func javaBRAFClosedIOException(t *testing.T, op func()) {
	t.Helper()
	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("The operation did not throw anything")
		}
		e, ok := v.(error)
		if !ok || !javaBRAFIsIOException(e) {
			t.Fatalf("The operation threw a non-IOException: %T: %v", v, v)
		}
		message := e.(interface{ GetMessage() *string }).GetMessage()
		if !strings.Contains(*message, "File handle closed") {
			t.Fatalf("Message must contain 'File handle closed', but was '%v'", *message)
		}
	}()
	op()
}
func TestJavaBufferedRandomAccessFile(t *testing.T) {
	t.Run("testWrite", func(t *testing.T) {
		f := javaBRAFOpen(javaBRAFTemp(t, "BufferedRandomAccessFileTest_testWrite"), "rw")
		for i := int64(0); i < BufferedRandomAccessFileBuffSz/8; i++ {
			f.writeLong(i)
		}
		f.close()
	})
	t.Run("testWriteSeek", func(t *testing.T) {
		f := javaBRAFOpen(javaBRAFTemp(t, "BufferedRandomAccessFileTest_testWriteSeek"), "rw")
		f.setLength(BufferedRandomAccessFileBuffSz + 1)
		f.seek(1)
		for i := int64(0); i < BufferedRandomAccessFileBuffSz/8; i++ {
			f.writeLong(i)
		}
		f.close()
	})
	t.Run("testWriteSeekNoLength", func(t *testing.T) {
		f := javaBRAFOpen(javaBRAFTemp(t, "BufferedRandomAccessFileTest_testWriteSeekNoLength"), "rw")
		f.seek(1)
		defer f.close()
		javaBRAFAcceptIOException(func() {
			for i := int64(0); i < BufferedRandomAccessFileBuffSz/8; i++ {
				f.writeLong(i)
			}
		})
	})
	t.Run("testRead", func(t *testing.T) {
		f := javaBRAFOpen(javaBRAFTemp(t, "BufferedRandomAccessFileTest_testRead"), "rw")
		for i := int64(0); i < BufferedRandomAccessFileBuffSz/8; i++ {
			f.writeLong(i)
		}
		f.seek(0)
		for i := int64(0); i < BufferedRandomAccessFileBuffSz/8; i++ {
			javaBRAFEquals(t, i, f.readLong())
		}
		f.close()
	})
	t.Run("testReadSeekNoLength", func(t *testing.T) {
		f := javaBRAFOpen(javaBRAFTemp(t, "BufferedRandomAccessFileTest_testReadSeekNoLength"), "rw")
		f.seek(1)
		defer f.close()
		javaBRAFAcceptIOException(func() {
			for i := 0; i < BufferedRandomAccessFileBuffSz/8; i++ {
				f.readLong()
			}
		})
	})
	t.Run("testInvalidateBufferedData", func(t *testing.T) {
		name := javaBRAFTemp(t, "BufferedRandomAccessFileTest_testReadSeekNoLength")
		f := javaBRAFOpen(name, "rw")
		defer f.close()
		f.writeArray([]byte{1, 2, 3}, 0, 3)
		f.flush()
		func() {
			other := javaBRAFOpen(name, "rw")
			defer other.close()
			other.writeArray([]byte{10, 20, 30}, 0, 3)
		}()
		f.seek(0)
		javaBRAFEquals(t, 1, f.read())
		javaBRAFEquals(t, 2, f.read())
		f.invalidate()
		javaBRAFEquals(t, 30, f.read())
		f.seek(0)
		javaBRAFEquals(t, 10, f.read())
		javaBRAFEquals(t, 20, f.read())
		javaBRAFEquals(t, 30, f.read())
	})
	t.Run("testReadAfterSeekPastEndOfFile", func(t *testing.T) {
		f := javaBRAFOpen(javaBRAFTemp(t, "BufferedRandomAccessFileTest_testReadAfterSeekPastEndOfFile"), "r")
		defer f.close()
		f.seek(1)
		javaBRAFEquals(t, 1, f.pointer())
		javaBRAFEquals(t, -1, f.read())
		b := make([]byte, 100)
		javaBRAFEquals(t, -1, f.readArray(b, 0, len(b)))
	})
	t.Run("testWriteAfterSeekPastEndOfFile", func(t *testing.T) {
		f := javaBRAFOpen(javaBRAFTemp(t, "BufferedRandomAccessFileTest_testWriteAfterSeekPastEndOfFile"), "rw")
		defer f.close()
		f.seek(10000)
		javaBRAFEquals(t, 10000, f.pointer())
		b := make([]byte, 100)
		f.writeArray(b, 0, len(b))
		javaBRAFEquals(t, int64(10000+len(b)), f.length())
		javaBRAFEquals(t, int64(10000+len(b)), f.pointer())
	})
	t.Run("testObscureSetLengthBehavior", func(t *testing.T) {
		f := javaBRAFOpen(javaBRAFTemp(t, "tmp"), "rw")
		defer f.close()
		f.seek(200)
		javaBRAFEquals(t, 200, f.pointer())
		f.setLength(100)
		javaBRAFEquals(t, 100, f.pointer())
		f.setLength(300)
		javaBRAFEquals(t, 100, f.pointer())
	})
	t.Run("testIdempotentClose", func(t *testing.T) {
		f1 := javaBRAFTemp(t, "tmp")
		f2 := javaBRAFTemp(t, "tmp")
		func() { f := javaBRAFOpen(f1, "rw"); defer f.close(); f.writeLong(100); f.close() }()
		a := javaBRAFOpen(f1, "rw")
		defer a.close()
		b := javaBRAFOpen(f2, "rw")
		defer b.close()
		b.writeLong(200)
		javaBRAFEquals(t, 100, a.readLong())
	})
	t.Run("testIOExceptionOnUseAfterClose", func(t *testing.T) {
		b := make([]byte, 100)
		f := javaBRAFOpen(javaBRAFTemp(t, "tmp"), "rw")
		defer f.close()
		f.close()
		for _, op := range []func(){func() { f.pointer() }, func() { f.seek(0) }, func() { f.length() }, func() { f.setLength(100) }, func() { f.read() }, func() { f.readArray(b, 0, len(b)) }, func() { f.readInt() }, func() { f.readLong() }, func() { f.write(1) }, func() { f.writeArray(b, 0, len(b)) }, func() { f.writeInt(100) }, func() { f.writeLong(100) }, func() { f.flush() }, func() { f.invalidate() }, func() { f.reset() }} {
			javaBRAFClosedIOException(t, op)
		}
	})
	t.Run("regressionTest01", func(t *testing.T) {
		f := javaBRAFOpen(javaBRAFTemp(t, "tmp"), "rw")
		defer f.close()

		f.seek(98)
		f.write(20)
		f.write(43)
		f.seek(98)
		javaBRAFEquals(t, 20, f.read())

	})
	t.Run("regressionTest02", func(t *testing.T) {
		f := javaBRAFOpen(javaBRAFTemp(t, "tmp"), "rw")
		defer f.close()

		f.seek(98)
		f.write(20)
		f.write(43)
		f.write(126)
		f.seek(98)
		f.read()
		f.read()
		f.write(123)
		javaBRAFEquals(t, -1, f.read())

	})
	t.Run("regressionTest03", func(t *testing.T) {
		f := javaBRAFOpen(javaBRAFTemp(t, "tmp"), "rw")
		defer f.close()

		f.seek(11624)
		javaBRAFEquals(t, 0, f.length())

	})
	t.Run("regressionTest04", func(t *testing.T) {
		f := javaBRAFOpen(javaBRAFTemp(t, "tmp"), "rw")
		defer f.close()

		f.write(83)
		f.seek(0)
		buffer := make([]byte, 2)
		n := f.readArray(buffer, 0, 2)
		javaBRAFEquals(t, 1, n)
		javaBRAFEquals(t, 83, int64(buffer[0]))

	})
	t.Run("regressionTest05", func(t *testing.T) {
		f := javaBRAFOpen(javaBRAFTemp(t, "tmp"), "rw")
		defer f.close()

		b1 := javaBRAFBytes([]int8{13, 41, -113, 61, -7})
		f.writeArray(b1, 3, 1)
		f.write(117)
		b2 := javaBRAFBytes([]int8{63, -59, 84, -4, 7})
		f.writeArray(b2, 1, 1)
		f.seek(2)

		javaBRAFEquals(t, 197, f.read())

	})
	t.Run("regressionTest06", func(t *testing.T) {
		f := javaBRAFOpen(javaBRAFTemp(t, "tmp"), "rw")
		defer f.close()

		f.seek(9037)
		f.setLength(8311)
		b1 := javaBRAFBytes([]int8{114, -128, 118, -15, -71, -43, -70, -27, 122, -33, -8, 80, -75, -7, -123, -88, -109, 4, 124, -69, 73, -86, 21, -63, 84, -13, 106, 81, 64, 69, -109, -56, -87, -48, 66, -98, -83, 44, 111, -77, 78, 17, -123, 115, -18, -55, -42, -75, 73, -73, -23, -125, -113, -39, 85, -39, -119, -16, 76, -106, -47, -72, 113, 25, 78, -90, 35, -2, -109, 1, 63, 43, 14, 44, -112, 8, -15, 15, 18, -82, -20, 13, 63, 86, -35, -92, -71, 85, 1, 82, 116, 29, 112, -53, -31, 10, 5, 37, 84, 96, -43, 2, -11, 10, 23, 50, -100, 72, -17, -124, -43, 14, -69, -98, 34, 119, 46, -30, -71, 45, 69, 31, -81, 60, 48, 37, 59, 15, 69, -91, -100, 125, 35, -43, 97, 106, -71, 24, -90, 26, -4, -117, 79, -89, 6, 46, -90, 52, 3, -2, 71, 121, 13, 14, 14, -82, 71, 82, -84, -31, -37, -6, 108, -84, -16, 23, -29, -125, -70, -78, -83, -23, -111, 10, -51, -51, -76, -36, 46, -70, -30, -57, 112, 34, -54, -108, -16, -5, -114, -93, -3, 109, 101, 11, -119, -102, -14, 121, 22, -112, -47, 106, 96, -6, 8, -47, -7, 42, -66, -90, -60, -50, -43, 78, -36, -10, 29, -51, -112, 26, 4, -97, -97, -84, -40, -76, 65, -82, -91, -35, 42, -56, 44, -50, 112, -12, 36, -109, -107, -120, -79, 49, 99, -123, -25, -51, -68, -79, 94, 38, 57, 81, 75, 0, -25, -70, -5, 3, 116, -74, -55, -115, -127, -34, -21, -26, 100, 122, -122, 112, -97, 71, 48, -21, 37, -128, -85, -88, -2, 86, 22, -24, -115, -16, 41, 73, -104, -47, -22, -70, -56, 125, -32, -127, -111, -63, -94, 102, -27, 121, 56, 13, 58, -32, 112, 119, 105, -24, -75, 119, -53, -64, -123, -124, -108, -42, 9, -30, 36, 30, -109, 39, -18, 42, -79, -78, 62, 73, -37, -96, -41, 85, 39, 58, -54, 19, -69, 111, 60, -89, 30, 37, -78, -112, 29, -69, 98, 108, 120, -54, 59, -19, 60, 74, -58, 20, -56, 126, 23, -125, -110, 85, -52, 42})
		f.writeArray(b1, 196, 47)

		javaBRAFEquals(t, 8358, f.length())

	})
	t.Run("regressionTest07", func(t *testing.T) {
		f := javaBRAFOpen(javaBRAFTemp(t, "tmp"), "rw")
		defer f.close()

		f.seek(9247)
		f.setLength(1175)
		b1 := javaBRAFBytes([]int8{28, -61, -90, 105, -37, 115, 76, 15, -2, -82, -61, 66, -77, 73, -96, 119, 102, 8, 126, 40, -6, -80, 35, 112, 6, -9, -128, 72, 37, -20, 29, 119, -116, 110, 29, -83, -26, -63, 28, 79, -55, 75, 68, -44, -77, 13, 22, -58, 13, 5, 76, 88, 124, 104, -19, -87, -9, 121, -43, -83, 44, 84, 89, 39, 61, 60, -113, -46, -26, 56, 113, -121, 60, 90, -46, -70, 37, -95, 10, 18, 89, 21, 103, -77, 30, 10, -20, -23, 109, 81, -65, 7, 29, 96, -60, -88, -38, -65, -108, -127, -79, 106, -28, 63, -3, 31, -93, -74, -31, 18, 66, 37, 13, 56, -99, 2, -59, 29, -86, 33, -75, -76, 75, 61, 50, -64, -21, -122, -29, -70, -66, 8, 93, -28, 88, 40, -92, -10, 0, -11, -92, 40, 87, 40, 5, 35, -87, -86, -38, 22, 104, 113, 83, -70, 69, 32, -56, -27, -84, -78, 43, -125, -74, -84, -60, -46, -37, 99, 61, 86, -106, 56, -115, -41, 29, -28, -82, -111, 98, 17, -54, -64, 52, 2, 125, -118, 23, 65, -30, -125, 119, -102, 91, -71, -108, 75, 7, -73, -121, 58, -55, -81, 31, -85, -96, 76, -84, 66, -119, -29, 57, -126, -122, -69, -10, 124, -98, 97, -81, -23, -89, -97, 14, 12, 109, 79, -24, 38, 8, 43, 89, 108, -68, 23, -114, 112, 70, -123, 76, 72, -84, 10, 25, -46, 4, 8, 121, -59, -43, 98, 18, -114, -32, 48, 43, 26, 49, -33, 39, -68, -18, -75, 39, -21, -127, -64, -28, 74, -53, 119, -40, -62, 64, -54, 3, 44, -68, -92, -52, -80, 116, 58, -121, 108, 123, 82, -94, 106, -53, 80, -121, 16, -91, -80, 52, -106, -15, 68, 91, 125, -33, -53, 66, 124, 4, 69, 17, -127, 43, 58, -111, 70, -127, -112, -126, -86, 10, 3, -15, -125, 51, -47, 107, -70, 91, 58, 2, -82, -32, -62, -114, -92, 22, 35, 78, -45, -56, -37, -104, 15, 83, -41, -10, -68, 56, 75, -65, -10, 69, 117, 79, -21, -76, 24, -8, 63, 5, 61, 100, 27, -111, -59, -53, 104, 104, -15, -76, 59, -13, 91, -84, 86, 57, -66, -118, -57, -122, -95, 44, -100, -124, -16, -83, 38, 109, 1, 107, 23, -40, -127, -3, 80, 60, 0, -69, -58, -79, -27, 67, 77, -2, 125, 57, -108, 102, 64, 33, 1, -112, -3, -101, -4, 69, 65, 61, 0, -45, 109, 74, -77, -48, -43, -102, -126, -113, -40, -6, 41, 50, -30, -88, -73, -9, -77, 108, -99, 103, -85, 110, 49, 42, -81, 88, -99, -84, -71, 100, -69, 45, 127, -108, -119, 44, 62, 30, 42, 42, -63, 111, -68, 14, -123, -8, -90, -1, 71, 47, 122, 77, 69, 33, -17, 97, 110, 19, 75, -2, -63, 71, 69, 23, 28, 113, -14, -17, 60, -20, 85, 34, 100, 85, 28, -50, 64, 44, -29, -52, -54, -38, -83, 46, 89, -114, -89, -14, -83, -53, 93, 37, -7, 60, -51, -1, -58, 29, 16, 97, 37, 13, 9, 126, -99, 90, 34, -34, 4, 27, 59, 40, -46, 123, -128, -94, 83, 68, 101, -48, -67, -96, -48, 47, -27, -40, -101, -85, -60, 13, -81, -115, -128, 116, 9, -67, 69, 82, -128, 13, 118, -82, -103, 112, 7, 62, -19, 70, -17, 19, -57, -94, -67, -76, -76, -49, 51, 110, 113, -78, -110, 120, 50, -64, -127, 32, -43, 79, -52, 66, -56, 17, -73, 86, 122, 22, -122, -37, 19, 61, 113, 98, 121, 66, 95, -117, -127, -37, -118, 37, 47, -2, -83, -48, 46, 16, -117, 3, 27, 23, -94, -34, -38, 1, -63, -2, 79, 31, -118, 41, 44, -86, -11, -42, 43, 63, 121, -114, -24, -52, -24, 14, 16, 80, -87, -36, 72, -45, -63, -23, 0, -16, 115, -81, -45, -39, 75, 3, -40, -91, 110, 28, 4, 55, -52, -76, 109, 92, 105, -94, 28, 65, -107, -50, 65, -68, -109, 108, 77, 35, -77, -52, -77, 36, 53, 126, 72, -6, 70, -18, 64, 86, -4, -36, 1, -111, -28, 116, 19, 15, 71, -12, -15, 124})
		f.writeArray(b1, 55, 500)
		f.write(2)
		f.setLength(71)
		f.seek(1675)

		f.write(97)

	})
}
