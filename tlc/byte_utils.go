package tlc

import (
	"encoding/binary"
	"fmt"
	"io"
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
	return int32(binary.BigEndian.Uint32(b[:4]))
}

func ByteArrayToLong(b []byte) int64 {
	return int64(binary.BigEndian.Uint64(b[:8]))
}

func ByteArrayToByteArray(src []byte, length int) ([]byte, error) {
	if len(src) > length {
		return nil, fmt.Errorf("byteArrayToByteArray: b needs more than length bytes.")
	}
	out := make([]byte, length)
	fill := byte(0)
	if len(src) > 0 && src[0]&0x80 != 0 {
		fill = 0xff
	}
	for i := range out {
		out[i] = fill
	}
	copy(out[length-len(src):], src)
	return out, nil
}

func WriteInt(out io.Writer, i int32) error {
	_, err := out.Write(IntToByteArray(i))
	return err
}

func WriteLong(out io.Writer, l int64) error {
	_, err := out.Write(LongToByteArray(l))
	return err
}

func WriteSizeByteArray(out io.Writer, bytes []byte) error {
	if err := WriteInt(out, int32(len(bytes))); err != nil {
		return err
	}
	_, err := out.Write(bytes)
	return err
}

func WriteByteArray(out io.Writer, bytes []byte, length int) error {
	if len(bytes) > length {
		return fmt.Errorf("writeByteArray: the byte array too large")
	}
	padded, err := ByteArrayToByteArray(bytes, length)
	if err != nil {
		return err
	}
	_, err = out.Write(padded)
	return err
}

func ReadInto(in io.Reader, b []byte, off int, length int) (int, error) {
	if off < 0 || length < 0 || off+length > len(b) {
		return 0, fmt.Errorf("ReadInto: invalid offset/length")
	}
	count := 0
	for count < length {
		n, err := in.Read(b[off+count : off+length])
		if n > 0 {
			count += n
		}
		if err != nil {
			if err == io.EOF {
				return count, nil
			}
			return count, err
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
			return 0, fmt.Errorf("readInt: the input stream is empty.")
		}
		return 0, fmt.Errorf("readInt: not enought bytes.")
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
			return 0, fmt.Errorf("readLong: the imput stream is empty.")
		}
		return 0, fmt.Errorf("readLong: not enought bytes.")
	}
	return ByteArrayToLong(b), nil
}

func ReadSizeByteArray(in io.Reader) ([]byte, error) {
	length, err := ReadInt(in)
	if err != nil {
		return nil, err
	}
	if length < 0 {
		return nil, fmt.Errorf("readSizeByteArray: negative length.")
	}
	out := make([]byte, int(length))
	count, err := ReadBytes(in, out)
	if err != nil {
		return nil, err
	}
	if count != int(length) {
		return nil, fmt.Errorf("readSizeByteArray: not enough bytes.")
	}
	return out, nil
}

func AppendByteArray(in io.Reader, out io.Writer) error {
	count, err := ReadInt(in)
	if err != nil {
		return fmt.Errorf("Can't append in to out; in is empty.")
	}
	for i := int32(0); i < count; i++ {
		bytes, err := ReadSizeByteArray(in)
		if err != nil {
			return fmt.Errorf("Can't append in to out; not enough bytes, but not empty.")
		}
		if err := WriteSizeByteArray(out, bytes); err != nil {
			return err
		}
	}
	return nil
}

func AppendSizeByteArray(in io.Reader, out io.Writer) error {
	count, err := ReadInt(in)
	if err != nil {
		return fmt.Errorf("Can't append in to out; in is empty.")
	}
	if err := WriteInt(out, count); err != nil {
		return err
	}
	for i := int32(0); i < count; i++ {
		bytes, err := ReadSizeByteArray(in)
		if err != nil {
			return fmt.Errorf("Can't append in to out; not enough bytes, but not empty.")
		}
		if err := WriteSizeByteArray(out, bytes); err != nil {
			return err
		}
	}
	return nil
}
