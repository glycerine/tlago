package tlc

import "io"

type ExternalSortable interface {
	Write(io.Writer) error
	Read(io.Reader) (ExternalSortable, error)
}

func WriteSizeArrayOfExternalSortable(values []ExternalSortable, start int, finish int, out io.Writer) error {
	start = int(int32(start))
	finish = int(int32(finish))
	if err := WriteInt(out, int32(finish)-int32(start)+1); err != nil {
		return err
	}
	return WriteArrayOfExternalSortable(values, start, finish, out)
}

func WriteArrayOfExternalSortable(values []ExternalSortable, start int, finish int, out io.Writer) error {
	start = int(int32(start))
	finish = int(int32(finish))
	for i := start; i <= finish; i = int(int32(i) + 1) {
		if isNil(values[i]) {
			panic(NewNullPointerException())
		}
		if err := values[i].Write(out); err != nil {
			return err
		}
	}
	return nil
}

func ReadSizeArrayOfExternalSortable(in io.Reader, prototype ExternalSortable) ([]ExternalSortable, error) {
	length, err := ReadInt(in)
	if err != nil {
		return nil, NewIOException("Can't read an array of ExternalSortables from the input stream; it's empty.")
	}
	if length < 0 {
		return nil, NewNegativeArraySizeException(fmtInt(int(length)))
	}
	values := make([]ExternalSortable, int(length))
	for i := range values {
		if isNil(prototype) {
			panic(NewNullPointerException())
		}
		value, err := prototype.Read(in)
		if err != nil {
			return nil, NewIOException("Can't read an array of ExternalSortables from the input stream; not enough bytes, but not empty.")
		}
		values[i] = value
	}
	return values, nil
}

func ReadArrayOfExternalSortable(in io.Reader, prototype ExternalSortable) []ExternalSortable {
	values := make([]ExternalSortable, 0)
	for {
		if isNil(prototype) {
			panic(NewNullPointerException())
		}
		value, err := prototype.Read(in)
		if err != nil {
			return values
		}
		values = append(values, value)
	}
}

func AppendSizeExternalSortableArraySizeArray(in io.Reader, out io.Writer, prototype ExternalSortable) error {
	length, err := ReadInt(in)
	if err != nil {
		return NewIOException("Can't append in to out; in is empty.")
	}
	if err := WriteInt(out, length); err != nil {
		return err
	}
	return appendExternalSortables(in, out, prototype, length)
}

func AppendSizeExternalSortableArrayArray(in io.Reader, out io.Writer, prototype ExternalSortable) error {
	length, err := ReadInt(in)
	if err != nil {
		return NewIOException("Can't append in to out; in is empty.")
	}
	return appendExternalSortables(in, out, prototype, length)
}

func appendExternalSortables(in io.Reader, out io.Writer, prototype ExternalSortable, length int32) error {
	for i := int32(0); i < length; i++ {
		if isNil(prototype) {
			panic(NewNullPointerException())
		}
		value, err := prototype.Read(in)
		if err != nil {
			return NewIOException("Can't append in to out; not enough bytes, but not empty.")
		}
		if isNil(value) {
			panic(NewNullPointerException())
		}
		if err := value.Write(out); err != nil {
			return NewIOException("Can't append in to out; not enough bytes, but not empty.")
		}
	}
	return nil
}

func WriteSizeArrayOfExternalSortableBigInts(out io.Writer, values []*BigInt, start int, finish int) error {
	return WriteSizeArrayOfExternalSortable(bigIntsAsExternalSortables(values), start, finish, out)
}

func WriteArrayOfExternalSortableBigInts(out io.Writer, values []*BigInt, start int, finish int) error {
	return WriteArrayOfExternalSortable(bigIntsAsExternalSortables(values), start, finish, out)
}

func ReadSizeArrayOfExternalSortableBigInts(in io.Reader) ([]*BigInt, error) {
	values, err := ReadSizeArrayOfExternalSortable(in, BigZero)
	return externalSortablesAsBigInts(values), err
}

func ReadArrayOfExternalSortableBigInts(in io.Reader) []*BigInt {
	return externalSortablesAsBigInts(ReadArrayOfExternalSortable(in, BigZero))
}

func AppendSizeExternalSortableBigIntArraySizeArray(in io.Reader, out io.Writer) error {
	return AppendSizeExternalSortableArraySizeArray(in, out, BigZero)
}

func AppendSizeExternalSortableBigIntArrayArray(in io.Reader, out io.Writer) error {
	return AppendSizeExternalSortableArrayArray(in, out, BigZero)
}

func bigIntsAsExternalSortables(values []*BigInt) []ExternalSortable {
	out := make([]ExternalSortable, len(values))
	for i, value := range values {
		out[i] = value
	}
	return out
}

func externalSortablesAsBigInts(values []ExternalSortable) []*BigInt {
	if values == nil {
		return nil
	}
	out := make([]*BigInt, len(values))
	for i, value := range values {
		out[i] = value.(*BigInt)
	}
	return out
}
