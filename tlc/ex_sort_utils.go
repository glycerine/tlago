package tlc

import (
	"fmt"
	"io"
)

func WriteSizeArrayOfExternalSortableBigInts(out io.Writer, values []*BigInt, start int, finish int) error {
	if err := WriteInt(out, int32(finish-start+1)); err != nil {
		return err
	}
	return WriteArrayOfExternalSortableBigInts(out, values, start, finish)
}

func WriteArrayOfExternalSortableBigInts(out io.Writer, values []*BigInt, start int, finish int) error {
	for i := start; i <= finish; i++ {
		if err := values[i].Write(out); err != nil {
			return err
		}
	}
	return nil
}

func ReadSizeArrayOfExternalSortableBigInts(in io.Reader) ([]*BigInt, error) {
	length, err := ReadInt(in)
	if err != nil {
		return nil, fmt.Errorf("Can't read an array of ExternalSortables from the input stream; it's empty.")
	}
	values := make([]*BigInt, int(length))
	for i := range values {
		value, err := ReadBigInt(in)
		if err != nil {
			return nil, fmt.Errorf("Can't read an array of ExternalSortables from the input stream; not enough bytes, but not empty.")
		}
		values[i] = value
	}
	return values, nil
}

func ReadArrayOfExternalSortableBigInts(in io.Reader) []*BigInt {
	values := make([]*BigInt, 0)
	for {
		value, err := ReadBigInt(in)
		if err != nil {
			return values
		}
		values = append(values, value)
	}
}

func AppendSizeExternalSortableBigIntArraySizeArray(in io.Reader, out io.Writer) error {
	length, err := ReadInt(in)
	if err != nil {
		return fmt.Errorf("Can't append in to out; in is empty.")
	}
	if err := WriteInt(out, length); err != nil {
		return err
	}
	for i := int32(0); i < length; i++ {
		value, err := ReadBigInt(in)
		if err != nil {
			return fmt.Errorf("Can't append in to out; not enough bytes, but not empty.")
		}
		if err := value.Write(out); err != nil {
			return err
		}
	}
	return nil
}

func AppendSizeExternalSortableBigIntArrayArray(in io.Reader, out io.Writer) error {
	length, err := ReadInt(in)
	if err != nil {
		return fmt.Errorf("Can't append in to out; in is empty.")
	}
	for i := int32(0); i < length; i++ {
		value, err := ReadBigInt(in)
		if err != nil {
			return fmt.Errorf("Can't append in to out; not enough bytes, but not empty.")
		}
		if err := value.Write(out); err != nil {
			return err
		}
	}
	return nil
}
