/*******************************************************************************
 * Copyright (c) 2017 Microsoft Research. All rights reserved.
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
	"os"
	"path/filepath"
	"testing"
)

// Complete translation of tlc2.value.ValueInputOutputStreamTest. In particular,
// the compact-natural cases retain the original compressed file-size checks,
// and the blind reads write raw string bytes before consulting the interner.
func TestJavaValueInputOutputStream(t *testing.T) {
	oldGzip := UseGZIP()
	oldIntern := internTable
	SetUseGZIP(true) // TLCGlobals' default in the original isolated JVM.
	internTable = NewInternTable(1024)
	t.Cleanup(func() { SetUseGZIP(oldGzip); internTable = oldIntern })
	for _, name := range []string{"testWriteShort", "testWriteInt", "testWriteShortNat", "testWriteNat", "testBlindReadStringValue", "testBlindReadRecordValue"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "ValueOutputStreamTest-"+name+".vos")
			if name == "testWriteShortNat" || name == "testWriteNat" {
				SetUseGZIP(true)
			}
			file, err := os.Create(path)
			if err != nil {
				t.Fatal(err)
			}
			out := NewValueOutputStreamWithGlobalCompression(file)
			check := func(err error) {
				t.Helper()
				if err != nil {
					t.Fatal(err)
				}
			}
			switch name {
			case "testWriteShort":
				check(out.WriteShort(32767))
				check(out.WriteShort(-32768))
				check(out.WriteShort(0))
			case "testWriteInt":
				check(out.WriteInt(2147483647))
				check(out.WriteInt(-2147483648))
				check(out.WriteInt(0))
			case "testWriteShortNat":
				check(out.WriteShortNat(32767))
				check(out.WriteShortNat(0))
			case "testWriteNat":
				check(out.WriteNat(2147483647))
				check(out.WriteNat(0))
			case "testBlindReadStringValue":
				str := "Hippopotomonstrosesquippedaliophobia"
				check(out.WriteByte(byte(StringValueKind)))
				check(out.WriteInt(-1))
				check(out.WriteInt(-1))
				check(out.WriteInt(int32(len([]byte(str)))))
				_, err = out.WriteRaw([]byte(str))
				check(err)
			case "testBlindReadRecordValue":
				str := "Well, let's see, we have on the bags, Who's on first, What's on second, I Don't Know is on third"
				check(out.WriteByte(byte(RecordValueKind)))
				check(out.WriteInt(1))
				check(out.WriteByte(byte(StringValueKind)))
				check(out.WriteInt(-1))
				check(out.WriteInt(-1))
				check(out.WriteInt(int32(len([]byte(str)))))
				_, err = out.WriteRaw([]byte(str))
				check(err)
				check(out.WriteByte(byte(IntValueKind)))
				check(out.WriteInt(42))
			}
			check(out.Close())
			if name == "testWriteShortNat" || name == "testWriteNat" {
				info, err := os.Stat(path)
				check(err)
				want := int64(20 + 2 + 1)
				if name == "testWriteNat" {
					want = 20 + 4 + 2
				}
				if info.Size() != want {
					t.Fatalf("compressed length: got %d, want %d", info.Size(), want)
				}
			}
			file, err = os.Open(path)
			check(err)
			in, err := NewValueInputStreamWithGlobalCompression(file)
			check(err)
			switch name {
			case "testWriteShort":
				for _, want := range []int16{32767, -32768, 0} {
					got, err := in.ReadShort()
					check(err)
					if got != want {
						t.Fatalf("readShort: %d != %d", got, want)
					}
				}
			case "testWriteInt":
				for _, want := range []int32{2147483647, -2147483648, 0} {
					got, err := in.ReadInt()
					check(err)
					if got != want {
						t.Fatalf("readInt: %d != %d", got, want)
					}
				}
			case "testWriteShortNat":
				for _, want := range []int16{32767, 0} {
					got, err := in.ReadShortNat()
					check(err)
					if got != want {
						t.Fatalf("readShortNat: %d != %d", got, want)
					}
				}
			case "testWriteNat":
				for _, want := range []int32{2147483647, 0} {
					got, err := in.ReadNat()
					check(err)
					if got != want {
						t.Fatalf("readNat: %d != %d", got, want)
					}
				}
			case "testBlindReadStringValue":
				value, err := in.ReadExternal()
				check(err)
				if value.Kind() != StringValueKind {
					t.Fatalf("kind: %d", value.Kind())
				}
				valueStr := value.(*StringValue)
				str := "Hippopotomonstrosesquippedaliophobia"
				if !UniqueStringOf(str).Equal(valueStr.Val) {
					t.Fatal("read string is not interned")
				}
				if valueStr.Val.String() != str {
					t.Fatalf("string: %q", valueStr.Val.String())
				}
			case "testBlindReadRecordValue":
				value, err := in.ReadExternal()
				check(err)
				if value.Kind() != RecordValueKind {
					t.Fatalf("kind: %d", value.Kind())
				}
				rec := value.(*RecordValue)
				size, err := rec.Size()
				check(err)
				if size != 1 {
					t.Fatalf("record size: %d", size)
				}
				str := "Well, let's see, we have on the bags, Who's on first, What's on second, I Don't Know is on third"
				got, err := rec.Select(NewStringValue(str))
				check(err)
				equal, err := NewIntValue(42).Equal(got)
				check(err)
				if !equal {
					t.Fatal("record selection did not return 42")
				}
			}
			check(in.Close())
		})
	}
}
