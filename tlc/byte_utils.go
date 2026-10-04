// Copyright (c) 2003 Compaq Corporation.  All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation.  All rights reserved.
/*  Notes: If imporved efficiency is needed, one place to look is at
    int to byte arrays and BigInts to byte arrays and back,
    because I use the built in Java routines, and it may be possible
    to optimize them.
*/
package tlc

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/big"
	"strconv"
	"strings"
)

func IntToByteArray(x int32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(x))
	return b
}

func LongToByteArray(x int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(x))
	return b
}

func ByteArrayToInt(b []byte) int32 {
	if len(b) < 4 {
		panic(NewArrayIndexOutOfBoundsException(len(b), len(b)))
	}
	return int32(binary.BigEndian.Uint32(b[:4]))
}

func ByteArrayToLong(b []byte) int64 {
	if len(b) < 8 {
		panic(NewArrayIndexOutOfBoundsException(len(b), len(b)))
	}
	return int64(binary.BigEndian.Uint64(b[:8]))
}

func ByteArrayToByteArray(src []byte, length int) ([]byte, error) {
	if len(src) > length {
		return nil, NewIOException("byteArrayToByteArray: b needs more than length bytes.")
	}
	out := make([]byte, length)
	fill := byte(0)
	if len(src) == 0 {
		panic(NewArrayIndexOutOfBoundsException(0, 0))
	}
	if src[0]&0x80 != 0 {
		fill = 0xff
	}
	for i := range out {
		out[i] = fill
	}
	copy(out[length-len(src):], src)
	return out, nil
}

func BigIntToJavaBytes(value *big.Int) []byte {
	if value == nil || value.Sign() == 0 {
		return []byte{0}
	}
	if value.Sign() > 0 {
		out := value.Bytes()
		if out[0]&0x80 != 0 {
			padded := make([]byte, len(out)+1)
			copy(padded[1:], out)
			out = padded
		}
		return out
	}
	for length := 1; ; length++ {
		min := new(big.Int).Lsh(big.NewInt(1), uint(8*length-1))
		min.Neg(min)
		max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(8*length-1)), big.NewInt(1))
		if value.Cmp(min) < 0 || value.Cmp(max) > 0 {
			continue
		}
		mod := new(big.Int).Lsh(big.NewInt(1), uint(8*length))
		encoded := new(big.Int).Add(value, mod).Bytes()
		out := make([]byte, length)
		copy(out[length-len(encoded):], encoded)
		return out
	}
}

func JavaBytesToBigInt(bytes []byte) (*big.Int, error) {
	if len(bytes) == 0 {
		return nil, &NumberFormatException{NewIllegalArgumentException("Zero length BigInteger")}
	}
	value := new(big.Int).SetBytes(bytes)
	if bytes[0]&0x80 == 0 {
		return value, nil
	}
	mod := new(big.Int).Lsh(big.NewInt(1), uint(8*len(bytes)))
	return value.Sub(value, mod), nil
}

func BigIntToByteArray(value *big.Int, length int) ([]byte, error) {
	return ByteArrayToByteArray(BigIntToJavaBytes(value), length)
}

// io.Reader/io.Writer failures correspond to InputStream/OutputStream
// IOExceptions; represented Java unchecked exceptions retain their identity.
func byteUtilsStreamError(err error) error {
	if err == nil || isJavaIOException(err) {
		return err
	}
	if javaThrowableClassName(err) == fmt.Sprintf("%T", err) {
		return NewIOException(err.Error())
	}
	return err
}

func WriteInt(out io.Writer, i int32) error {
	_, err := out.Write(IntToByteArray(i))
	return byteUtilsStreamError(err)
}

func WriteLong(out io.Writer, l int64) error {
	_, err := out.Write(LongToByteArray(l))
	return byteUtilsStreamError(err)
}

func WriteSizeByteArray(out io.Writer, bytes []byte) error {
	if err := WriteInt(out, int32(len(bytes))); err != nil {
		return byteUtilsStreamError(err)
	}
	_, err := out.Write(bytes)
	return byteUtilsStreamError(err)
}

func WriteSizeBigInt(out io.Writer, value *big.Int) error {
	if value == nil {
		panic(NewNullPointerException())
	}
	return WriteSizeByteArray(out, BigIntToJavaBytes(value))
}

func WriteByteArray(out io.Writer, bytes []byte, length int) error {
	if len(bytes) > length {
		return NewIOException("writeByteArray: the byte array too large")
	}
	padded, err := ByteArrayToByteArray(bytes, length)
	if err != nil {
		return byteUtilsStreamError(err)
	}
	_, err = out.Write(padded)
	return byteUtilsStreamError(err)
}

func WriteBigInt(out io.Writer, value *big.Int, length int) error {
	if value == nil {
		panic(NewNullPointerException())
	}
	return WriteByteArray(out, BigIntToJavaBytes(value), length)
}

func WriteSizeArrayOfSizeBigInts(out io.Writer, values []*big.Int, start int, finish int) error {
	if err := WriteInt(out, int32(finish-start+1)); err != nil {
		return err
	}
	return WriteArrayOfSizeBigInts(out, values, start, finish)
}

func WriteArrayOfSizeBigInts(out io.Writer, values []*big.Int, start int, finish int) error {
	for i := start; i <= finish; i++ {
		if i < 0 || i >= len(values) {
			panic(NewArrayIndexOutOfBoundsException(i, len(values)))
		}
		if values[i] == nil {
			panic(NewNullPointerException())
		}
		if err := WriteSizeBigInt(out, values[i]); err != nil {
			return err
		}
	}
	return nil
}

func ReadInto(in io.Reader, b []byte, off int, length int) (int, error) {
	count := 0
	for count < length {
		if off+count < 0 || off+length > len(b) {
			panic(&IndexOutOfBoundsException{javaExceptionBase: newJavaExceptionBase(nil, nil)})
		}
		n, err := in.Read(b[off+count : off+length])
		if n > 0 {
			count += n
		}
		if err != nil {
			if err == io.EOF {
				return count, nil
			}
			return count, byteUtilsStreamError(err)
		}
		if n <= 0 {
			return count, nil
		}
	}
	return count, nil
}

func ReadBytes(in io.Reader, b []byte) (int, error) {
	return ReadInto(in, b, 0, len(b))
}

func ReadInt(in io.Reader) (int32, error) {
	b := make([]byte, 4)
	count, err := ReadBytes(in, b)
	if err != nil {
		return 0, err
	}
	if count < 4 {
		if count <= 0 {
			return 0, NewIOException("readInt: the input stream is empty.")
		}
		return 0, NewIOException("readInt: not enought bytes.")
	}
	return ByteArrayToInt(b), nil
}

func ReadLong(in io.Reader) (int64, error) {
	b := make([]byte, 8)
	count, err := ReadBytes(in, b)
	if err != nil {
		return 0, err
	}
	if count < 8 {
		if count <= 0 {
			return 0, NewIOException("readLong: the imput stream is empty.")
		}
		return 0, NewIOException("readLong: not enought bytes.")
	}
	return ByteArrayToLong(b), nil
}

func ReadSizeByteArray(in io.Reader) ([]byte, error) {
	length, err := ReadInt(in)
	if err != nil {
		return nil, err
	}
	if length < 0 {
		panic(NewNegativeArraySizeException(fmt.Sprint(length)))
	}
	out := make([]byte, int(length))
	count, err := ReadBytes(in, out)
	if err != nil {
		return nil, err
	}
	if count != int(length) {
		return nil, NewIOException("readSizeByteArray: not enough bytes.")
	}
	return out, nil
}

func ReadSizeBigInt(in io.Reader) (*big.Int, error) {
	length, err := ReadInt(in)
	if err != nil {
		return nil, err
	}
	if length < 0 {
		panic(NewNegativeArraySizeException(fmt.Sprint(length)))
	}
	bytes := make([]byte, int(length))
	count, err := ReadBytes(in, bytes)
	if err != nil {
		return nil, err
	}
	if count != int(length) {
		return nil, NewIOException("readSizeBigInt: not enough bytes.")
	}
	value, err := JavaBytesToBigInt(bytes)
	if err != nil {
		panic(err) // BigInt(byte[]) throws unchecked NumberFormatException.
	}
	return value, nil
}

func ReadSizeArrayOfSizeBigInts(in io.Reader) ([]*big.Int, error) {
	length, err := ReadInt(in)
	if err != nil {
		if !isJavaIOException(err) {
			return nil, err
		}
		return nil, NewIOException("Can't read an array of BigInts from the input stream; it's empty.")
	}
	if length < 0 {
		panic(NewNegativeArraySizeException(fmt.Sprint(length)))
	}
	out := make([]*big.Int, int(length))
	for i := range out {
		value, err := ReadSizeBigInt(in)
		if err != nil {
			if !isJavaIOException(err) {
				return nil, err
			}
			return nil, NewIOException("Can't read an array of BigInts from the input stream; not enough bytes, but not empty.")
		}
		out[i] = value
	}
	return out, nil
}

func ReadArrayOfSizeBigInts(in io.Reader) []*big.Int {
	var out []*big.Int
	for {
		value, err := ReadSizeBigInt(in)
		if err != nil {
			if !isJavaIOException(err) {
				panic(err)
			}
			return out
		}
		out = append(out, value)
	}
}

func AppendByteArray(in io.Reader, out io.Writer) error {
	count, err := ReadInt(in)
	if err != nil {
		if !isJavaIOException(err) {
			return err
		}
		return NewIOException("Can't append in to out; in is empty.")
	}
	for i := int32(0); i < count; i++ {
		bytes, err := ReadSizeByteArray(in)
		if err != nil {
			if !isJavaIOException(err) {
				return err
			}
			return NewIOException("Can't append in to out; not enough bytes, but not empty.")
		}
		if err := WriteSizeByteArray(out, bytes); err != nil {
			if !isJavaIOException(err) {
				return err
			}
			return NewIOException("Can't append in to out; not enough bytes, but not empty.")
		}
	}
	return nil
}

func AppendSizeByteArray(in io.Reader, out io.Writer) error {
	count, err := ReadInt(in)
	if err != nil {
		if !isJavaIOException(err) {
			return err
		}
		return NewIOException("Can't append in to out; in is empty.")
	}
	if err := WriteInt(out, count); err != nil {
		return err
	}
	for i := int32(0); i < count; i++ {
		bytes, err := ReadSizeByteArray(in)
		if err != nil {
			if !isJavaIOException(err) {
				return err
			}
			return NewIOException("Can't append in to out; not enough bytes, but not empty.")
		}
		if err := WriteSizeByteArray(out, bytes); err != nil {
			if !isJavaIOException(err) {
				return err
			}
			return NewIOException("Can't append in to out; not enough bytes, but not empty.")
		}
	}
	return nil
}

func PrintHex(bytes []byte) string {
	var out strings.Builder
	for _, value := range bytes {
		text := strconv.FormatUint(uint64(value), 16)
		if len(text) > 2 {
			text = text[len(text)-2:]
		}
		out.WriteString(text)
	}
	return out.String()
}
