package tlc

import (
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
)

type ValueOutputStream struct {
	out     io.Writer
	closer  io.Closer
	handles map[uintptr]int
}

type orderedCloser []io.Closer

func closeInOrder(closers ...io.Closer) io.Closer {
	out := make(orderedCloser, 0, len(closers))
	for _, closer := range closers {
		if closer != nil {
			out = append(out, closer)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func (c orderedCloser) Close() error {
	var first error
	for _, closer := range c {
		if err := closer.Close(); err != nil && first == nil {
			first = err
		}
	}
	return first
}

func NewValueOutputStream(out io.Writer) *ValueOutputStream {
	return NewValueOutputStreamWithCompression(out, false)
}

func NewValueOutputStreamWithCompression(out io.Writer, compress bool) *ValueOutputStream {
	writer := out
	var closer io.Closer
	if c, ok := out.(io.Closer); ok {
		closer = c
	}
	if compress {
		gzipWriter := gzip.NewWriter(out)
		writer = gzipWriter
		closer = closeInOrder(gzipWriter, closer)
	}
	return &ValueOutputStream{out: writer, closer: closer, handles: make(map[uintptr]int)}
}

func NewValueOutputStreamWithGlobalCompression(out io.Writer) *ValueOutputStream {
	return NewValueOutputStreamWithCompression(out, UseGZIP())
}

func (s *ValueOutputStream) WriteShort(value int16) error {
	return binary.Write(s.out, binary.BigEndian, value)
}

func (s *ValueOutputStream) WriteInt(value int32) error {
	return binary.Write(s.out, binary.BigEndian, value)
}

func (s *ValueOutputStream) WriteLong(value int64) error {
	return binary.Write(s.out, binary.BigEndian, value)
}

func (s *ValueOutputStream) WriteByte(value byte) error {
	_, err := s.out.Write([]byte{value})
	return err
}

func (s *ValueOutputStream) WriteBool(value bool) error {
	if value {
		return s.WriteByte(1)
	}
	return s.WriteByte(0)
}

func (s *ValueOutputStream) WriteShortNat(value int16) error {
	if value < 0 {
		return fmt.Errorf("short nat cannot be negative: %d", value)
	}
	if value > 0x7f {
		return s.WriteShort(-value)
	}
	return s.WriteByte(byte(value))
}

func (s *ValueOutputStream) WriteNat(value int32) error {
	if value < 0 {
		return fmt.Errorf("nat cannot be negative: %d", value)
	}
	if value > 0x7fff {
		return s.WriteInt(-value)
	}
	return s.WriteShort(int16(value))
}

func (s *ValueOutputStream) WriteLongNat(value int64) error {
	if value < 0 {
		return fmt.Errorf("long nat cannot be negative: %d", value)
	}
	if value <= 0x7fffffff {
		return s.WriteInt(int32(value))
	}
	return s.WriteLong(-value)
}

func (s *ValueOutputStream) WriteRaw(bytes []byte) (int, error) {
	return s.out.Write(bytes)
}

func (s *ValueOutputStream) Close() error {
	if s.closer != nil {
		return s.closer.Close()
	}
	return nil
}

func (s *ValueOutputStream) Put(value any) int {
	key := pointerKey(value)
	if key == 0 {
		return -1
	}
	if idx, ok := s.handles[key]; ok {
		return idx
	}
	s.handles[key] = len(s.handles)
	return -1
}

func (s *ValueOutputStream) WriteExternal(value Value) error {
	if value == nil {
		return fmt.Errorf("cannot pickle nil TLC value")
	}
	switch v := value.(type) {
	case *BoolValue:
		if err := s.WriteByte(byte(BoolValueKind)); err != nil {
			return err
		}
		return s.WriteBool(v.Val)
	case *IntValue:
		if err := s.WriteByte(byte(IntValueKind)); err != nil {
			return err
		}
		return s.WriteInt(v.Val)
	case *StringValue:
		if idx := s.Put(v); idx >= 0 {
			return s.writeDummy(idx)
		}
		if err := s.WriteByte(byte(StringValueKind)); err != nil {
			return err
		}
		return s.WriteUniqueString(v.Val)
	case *ModelValue:
		if err := s.WriteByte(byte(ModelValueKind)); err != nil {
			return err
		}
		return s.WriteShort(int16(v.Index))
	case *TupleValue:
		if idx := s.Put(v); idx >= 0 {
			return s.writeDummy(idx)
		}
		if err := s.WriteByte(byte(TupleValueKind)); err != nil {
			return err
		}
		if err := s.WriteNat(int32(len(v.Elems))); err != nil {
			return err
		}
		for _, elem := range v.Elems {
			if err := s.WriteExternal(elem); err != nil {
				return err
			}
		}
		return nil
	case *SetEnumValue:
		if idx := s.Put(v); idx >= 0 {
			return s.writeDummy(idx)
		}
		if err := s.WriteByte(byte(SetEnumValueKind)); err != nil {
			return err
		}
		length := int32(v.Elems.Len())
		if !v.IsNorm {
			length = -length
		}
		if err := s.WriteInt(length); err != nil {
			return err
		}
		for i := 0; i < v.Elems.Len(); i++ {
			if err := s.WriteExternal(v.Elems.At(i)); err != nil {
				return err
			}
		}
		return nil
	case *IntervalValue:
		if err := s.WriteByte(byte(IntervalValueKind)); err != nil {
			return err
		}
		if err := s.WriteInt(v.Low); err != nil {
			return err
		}
		return s.WriteInt(v.High)
	case *FcnRcdValue:
		if idx := s.Put(v); idx >= 0 {
			return s.writeDummy(idx)
		}
		if err := s.WriteByte(byte(FcnRcdValueKind)); err != nil {
			return err
		}
		if err := s.WriteNat(int32(len(v.Values))); err != nil {
			return err
		}
		if v.Intv != nil {
			if err := s.WriteByte(0); err != nil {
				return err
			}
			if err := s.WriteInt(v.Intv.Low); err != nil {
				return err
			}
			if err := s.WriteInt(v.Intv.High); err != nil {
				return err
			}
			for _, value := range v.Values {
				if err := s.WriteExternal(value); err != nil {
					return err
				}
			}
			return nil
		}
		info := byte(2)
		if v.IsNorm {
			info = 1
		}
		if err := s.WriteByte(info); err != nil {
			return err
		}
		for i, dval := range v.Domain {
			if err := s.WriteExternal(dval); err != nil {
				return err
			}
			if err := s.WriteExternal(v.Values[i]); err != nil {
				return err
			}
		}
		return nil
	case *RecordValue:
		if idx := s.Put(v); idx >= 0 {
			return s.writeDummy(idx)
		}
		if err := s.WriteByte(byte(RecordValueKind)); err != nil {
			return err
		}
		length := int32(len(v.Names))
		if !v.IsNorm {
			length = -length
		}
		if err := s.WriteInt(length); err != nil {
			return err
		}
		for i, name := range v.Names {
			if idx := s.Put(name); idx >= 0 {
				if err := s.writeDummy(idx); err != nil {
					return err
				}
			} else {
				if err := s.WriteByte(byte(StringValueKind)); err != nil {
					return err
				}
				if err := s.WriteUniqueString(name); err != nil {
					return err
				}
			}
			if err := s.WriteExternal(v.Values[i]); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("cannot pickle value of kind %s", value.KindString())
	}
}

func (s *ValueOutputStream) writeDummy(index int) error {
	if err := s.WriteByte(byte(DummyValueKind)); err != nil {
		return err
	}
	return s.WriteNat(int32(index))
}

func (s *ValueOutputStream) WriteUniqueString(value *UniqueString) error {
	if value == nil {
		value = UniqueStringOf("")
	}
	raw := []byte(value.String())
	if err := s.WriteInt(-1); err != nil {
		return err
	}
	if err := s.WriteInt(-1); err != nil {
		return err
	}
	if err := s.WriteInt(int32(len(raw))); err != nil {
		return err
	}
	_, err := s.WriteRaw(raw)
	return err
}

func pointerKey(value any) uintptr {
	if value == nil {
		return 0
	}
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return 0
	}
	return rv.Pointer()
}

func valuePointerKey(value Value) uintptr {
	return pointerKey(value)
}

type ValueInputStream struct {
	in      io.Reader
	closer  io.Closer
	handles []any
}

func NewValueInputStream(in io.Reader) *ValueInputStream {
	stream := &ValueInputStream{in: in}
	if closer, ok := in.(io.Closer); ok {
		stream.closer = closer
	}
	return stream
}

func NewValueInputStreamWithCompression(in io.Reader, compressed bool) (*ValueInputStream, error) {
	reader := in
	var closer io.Closer
	if c, ok := in.(io.Closer); ok {
		closer = c
	}
	if compressed {
		gzipReader, err := gzip.NewReader(in)
		if err != nil {
			return nil, err
		}
		reader = gzipReader
		closer = closeInOrder(gzipReader, closer)
	}
	return &ValueInputStream{in: reader, closer: closer}, nil
}

func NewValueInputStreamWithGlobalCompression(in io.Reader) (*ValueInputStream, error) {
	return NewValueInputStreamWithCompression(in, UseGZIP())
}

func (s *ValueInputStream) ReadShort() (int16, error) {
	var value int16
	err := binary.Read(s.in, binary.BigEndian, &value)
	return value, err
}

func (s *ValueInputStream) ReadInt() (int32, error) {
	var value int32
	err := binary.Read(s.in, binary.BigEndian, &value)
	return value, err
}

func (s *ValueInputStream) ReadLong() (int64, error) {
	var value int64
	err := binary.Read(s.in, binary.BigEndian, &value)
	return value, err
}

func (s *ValueInputStream) ReadByte() (byte, error) {
	var buf [1]byte
	_, err := io.ReadFull(s.in, buf[:])
	return buf[0], err
}

func (s *ValueInputStream) ReadBool() (bool, error) {
	value, err := s.ReadByte()
	return value != 0, err
}

func (s *ValueInputStream) ReadShortNat() (int16, error) {
	first, err := s.ReadByte()
	if err != nil {
		return 0, err
	}
	signedFirst := int8(first)
	if signedFirst >= 0 {
		return int16(signedFirst), nil
	}
	second, err := s.ReadByte()
	if err != nil {
		return 0, err
	}
	encoded := (int16(signedFirst) << 8) | int16(second)
	return -encoded, nil
}

func (s *ValueInputStream) ReadNat() (int32, error) {
	first, err := s.ReadShort()
	if err != nil {
		return 0, err
	}
	if first >= 0 {
		return int32(first), nil
	}
	second, err := s.ReadShort()
	if err != nil {
		return 0, err
	}
	encoded := (int32(first) << 16) | int32(uint16(second))
	return -encoded, nil
}

func (s *ValueInputStream) ReadLongNat() (int64, error) {
	first, err := s.ReadInt()
	if err != nil {
		return 0, err
	}
	if first >= 0 {
		return int64(first), nil
	}
	second, err := s.ReadInt()
	if err != nil {
		return 0, err
	}
	encoded := (int64(first) << 32) | int64(uint32(second))
	return -encoded, nil
}

func (s *ValueInputStream) Close() error {
	if s.closer != nil {
		return s.closer.Close()
	}
	return nil
}

func (s *ValueInputStream) ReadExternal() (Value, error) {
	kind, err := s.ReadByte()
	if err != nil {
		return nil, err
	}
	return s.readExternalKind(ValueKind(kind))
}

func (s *ValueInputStream) readExternalKind(kind ValueKind) (Value, error) {
	switch kind {
	case BoolValueKind:
		value, err := s.ReadBool()
		if err != nil {
			return nil, err
		}
		return NewBoolValue(value), nil
	case IntValueKind:
		value, err := s.ReadInt()
		if err != nil {
			return nil, err
		}
		return NewIntValue(value), nil
	case StringValueKind:
		return s.readExternalStringValue()
	case ModelValueKind:
		index, err := s.ReadShort()
		if err != nil {
			return nil, err
		}
		value := ModelValueAtIndex(int(index))
		if value == nil {
			return nil, fmt.Errorf("model value index %d is not defined", index)
		}
		return value, nil
	case TupleValueKind:
		return s.readExternalTupleValue()
	case SetEnumValueKind:
		return s.readExternalSetEnumValue()
	case IntervalValueKind:
		low, err := s.ReadInt()
		if err != nil {
			return nil, err
		}
		high, err := s.ReadInt()
		if err != nil {
			return nil, err
		}
		return NewIntervalValue(low, high), nil
	case FcnRcdValueKind:
		return s.readExternalFcnRcdValue()
	case RecordValueKind:
		return s.readExternalRecordValue()
	case DummyValueKind:
		idx, err := s.ReadNat()
		if err != nil {
			return nil, err
		}
		value := s.valueAt(int(idx))
		if typed, ok := value.(Value); ok {
			return typed, nil
		}
		return nil, fmt.Errorf("dummy value index %d does not reference a Value", idx)
	default:
		return nil, fmt.Errorf("cannot unpickle value of kind %d", kind)
	}
}

func (s *ValueInputStream) readExternalTupleValue() (Value, error) {
	index := s.GetIndex()
	length, err := s.ReadNat()
	if err != nil {
		return nil, err
	}
	elems := make([]Value, int(length))
	for i := range elems {
		value, err := s.ReadExternal()
		if err != nil {
			return nil, err
		}
		elems[i] = value
	}
	value := NewTupleValue(elems)
	s.Assign(value, index)
	return value, nil
}

func (s *ValueInputStream) readExternalSetEnumValue() (Value, error) {
	index := s.GetIndex()
	length, err := s.ReadInt()
	if err != nil {
		return nil, err
	}
	isNorm := true
	if length < 0 {
		length = -length
		isNorm = false
	}
	elems := make([]Value, int(length))
	for i := range elems {
		value, err := s.ReadExternal()
		if err != nil {
			return nil, err
		}
		elems[i] = value
	}
	value := NewSetEnumValue(elems, isNorm)
	s.Assign(value, index)
	return value, nil
}

func (s *ValueInputStream) readExternalFcnRcdValue() (Value, error) {
	index := s.GetIndex()
	length, err := s.ReadNat()
	if err != nil {
		return nil, err
	}
	info, err := s.ReadByte()
	if err != nil {
		return nil, err
	}
	values := make([]Value, int(length))
	var value Value
	if info == 0 {
		low, err := s.ReadInt()
		if err != nil {
			return nil, err
		}
		high, err := s.ReadInt()
		if err != nil {
			return nil, err
		}
		for i := range values {
			values[i], err = s.ReadExternal()
			if err != nil {
				return nil, err
			}
		}
		value = NewFcnRcdIntervalValue(NewIntervalValue(low, high), values)
	} else {
		domain := make([]Value, int(length))
		for i := range domain {
			domain[i], err = s.ReadExternal()
			if err != nil {
				return nil, err
			}
			values[i], err = s.ReadExternal()
			if err != nil {
				return nil, err
			}
		}
		value = NewFcnRcdValue(domain, values, info == 1)
	}
	s.Assign(value, index)
	return value, nil
}

func (s *ValueInputStream) readExternalStringValue() (Value, error) {
	str, err := s.readExternalUniqueString()
	if err != nil {
		return nil, err
	}
	value := NewStringValueFromUnique(str)
	s.Assign(value, s.GetIndex())
	return value, nil
}

func (s *ValueInputStream) readExternalRecordValue() (Value, error) {
	index := s.GetIndex()
	length, err := s.ReadInt()
	if err != nil {
		return nil, err
	}
	isNorm := false
	if length < 0 {
		length = -length
	}
	names := make([]*UniqueString, int(length))
	values := make([]Value, int(length))
	for i := range names {
		kind, err := s.ReadByte()
		if err != nil {
			return nil, err
		}
		if ValueKind(kind) == DummyValueKind {
			idx, err := s.ReadNat()
			if err != nil {
				return nil, err
			}
			name, ok := s.valueAt(int(idx)).(*UniqueString)
			if !ok {
				return nil, fmt.Errorf("dummy string index %d does not reference a UniqueString", idx)
			}
			names[i] = name
		} else {
			stringIndex := s.GetIndex()
			name, err := s.readExternalUniqueString()
			if err != nil {
				return nil, err
			}
			s.Assign(name, stringIndex)
			names[i] = name
		}
		value, err := s.ReadExternal()
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	value := NewRecordValue(names, values, isNorm)
	s.Assign(value, index)
	return value, nil
}

func (s *ValueInputStream) readExternalUniqueString() (*UniqueString, error) {
	if _, err := s.ReadInt(); err != nil {
		return nil, err
	}
	if _, err := s.ReadInt(); err != nil {
		return nil, err
	}
	length, err := s.ReadInt()
	if err != nil {
		return nil, err
	}
	if length < 0 {
		return nil, fmt.Errorf("negative unique string byte length: %d", length)
	}
	buf := make([]byte, int(length))
	if _, err := io.ReadFull(s.in, buf); err != nil {
		return nil, err
	}
	return UniqueStringOf(string(buf)), nil
}

func (s *ValueInputStream) GetIndex() int {
	index := len(s.handles)
	s.handles = append(s.handles, nil)
	return index
}

func (s *ValueInputStream) Assign(value any, index int) {
	for index >= len(s.handles) {
		s.handles = append(s.handles, nil)
	}
	s.handles[index] = value
}

func (s *ValueInputStream) valueAt(index int) any {
	if index < 0 || index >= len(s.handles) {
		return nil
	}
	return s.handles[index]
}
