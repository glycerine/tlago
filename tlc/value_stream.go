package tlc

import (
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
)

type ValueOutputStream struct {
	out            io.Writer
	closer         io.Closer
	handles        map[any]int
	disableHandles bool
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

func NewValueOutputStreamWithoutHandles(out io.Writer) *ValueOutputStream {
	stream := NewValueOutputStream(out)
	stream.disableHandles = true
	return stream
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
	return &ValueOutputStream{out: writer, closer: closer, handles: make(map[any]int)}
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

// Precondition: value is non-negative, as in Java's compact natural encoders.
func (s *ValueOutputStream) WriteShortNat(value int16) error {
	if value > 0x7f {
		return s.WriteShort(-value)
	}
	return s.WriteByte(byte(value))
}

func (s *ValueOutputStream) WriteNat(value int32) error {
	if value > 0x7fff {
		return s.WriteInt(-value)
	}
	return s.WriteShort(int16(value))
}

func (s *ValueOutputStream) WriteLongNat(value int64) error {
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
	if s.disableHandles {
		return -1
	}
	if value != nil {
		rv := reflect.ValueOf(value)
		if rv.Kind() != reflect.Pointer {
			// Go objects passed to Java's reference-identity API are pointers.
			return -1
		}
		if rv.IsNil() {
			value = nil
		}
	}
	if idx, ok := s.handles[value]; ok {
		return idx
	}
	// Keep the object itself alive for the entire stream, like Java's Object[].
	// A uintptr key alone would let GC reclaim it and reuse its address.
	s.handles[value] = len(s.handles)
	return -1
}

func (s *ValueOutputStream) Write(value Value) error {
	return s.writeValue(value)
}

// External reads discard the saved intern-table metadata; Java has one write
// format for both local checkpoints and values sent to another process.
func (s *ValueOutputStream) WriteExternal(value Value) error {
	return s.Write(value)
}

func (s *ValueOutputStream) writeValue(value Value) error {
	if value == nil {
		panic(NewNullPointerException())
	}
	if rv := reflect.ValueOf(value); rv.Kind() == reflect.Pointer && rv.IsNil() {
		panic(NewNullPointerException())
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
			if err := s.writeValue(elem); err != nil {
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
		if v.Elems == nil {
			panic(NewNullPointerException())
		}
		count := v.Elems.Len()
		length := int32(count)
		if !v.IsNorm {
			length = -length
		}
		if err := s.WriteInt(length); err != nil {
			return err
		}
		for i := 0; i < count; i++ {
			if err := s.writeValue(v.Elems.At(i)); err != nil {
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
				if err := s.writeValue(value); err != nil {
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
		for i := range v.Values {
			if err := s.writeValue(v.Domain[i]); err != nil {
				return err
			}
			if err := s.writeValue(v.Values[i]); err != nil {
				return err
			}
		}
		return nil
	case *FcnLambdaValue:
		return s.writeValue(v.FcnRcd)
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
			if err := s.writeValue(v.Values[i]); err != nil {
				return err
			}
		}
		return nil
	case *CounterExample:
		return s.writeValue(v.RecordValue)
	case *SetOfTuplesValue:
		return s.writeValue(cachedSetForWrite(v.TupleSet, v.TupleSetDummy))
	case *SetOfRcdsValue:
		return s.writeValue(cachedSetForWrite(v.RcdSet, v.RcdSetDummy))
	case *SetOfFcnsValue:
		return s.writeValue(cachedSetForWrite(v.FcnSet, v.FcnSetDummy))
	case *SubsetValue:
		return s.writeValue(cachedSetForWrite(v.PSet, v.PSetDummy))
	case *KSubsetValue:
		return s.writeValue(cachedSetForWrite(v.PSet, v.PSetDummy))
	case *SetCupValue:
		return s.writeValue(cachedSetForWrite(v.CupSet, v.CupSetDummy))
	case *SetCapValue:
		return s.writeValue(cachedSetForWrite(v.CapSet, v.CapSetDummy))
	case *SetDiffValue:
		return s.writeValue(cachedSetForWrite(v.DiffSet, v.DiffSetDummy))
	case *UnionValue:
		return s.writeValue(cachedSetForWrite(v.RealSet, v.RealSetDummy))
	case *SetPredValue:
		return s.writeValue(v.InVal)
	default:
		return NewWrongInvocationException("ValueOutputStream: Can not pickle the value\n" + ValuesPPR(value))
	}
}

func (s *ValueOutputStream) writeDummy(index int) error {
	if err := s.WriteByte(byte(DummyValueKind)); err != nil {
		return err
	}
	return s.WriteNat(int32(index))
}

func (s *ValueOutputStream) WriteUniqueString(value *UniqueString) error {
	return writeJavaUniqueString(s, value)
}

func (s *ValueOutputStream) WriteExternalUniqueString(value *UniqueString) error {
	return s.WriteUniqueString(value)
}

// Java's DummyEnum has a null ValueVec. It is distinct from EmptySet and can
// enter the handle table before its first write fails while reading elems.size().
var dummySetEnumForStream = &SetEnumValue{IsNorm: true}

func cachedSetForWrite(set *SetEnumValue, dummy bool) *SetEnumValue {
	if dummy {
		return dummySetEnumForStream
	}
	return set
}

type ValueInputStream struct {
	in             io.Reader
	closer         io.Closer
	handles        []any
	handleIndex    int
	disableHandles bool
}

func NewValueInputStream(in io.Reader) *ValueInputStream {
	stream := &ValueInputStream{in: in, handles: make([]any, 16)}
	if closer, ok := in.(io.Closer); ok {
		stream.closer = closer
	}
	return stream
}

// NewValueInputStreamWithoutHandles mirrors the queue byte stream's no-op
// assign and getIndex=-1 behavior; reference records are unsupported.
func NewValueInputStreamWithoutHandles(in io.Reader) *ValueInputStream {
	stream := &ValueInputStream{in: in, disableHandles: true}
	if closer, ok := in.(io.Closer); ok {
		stream.closer = closer
	}
	return stream
}

// NewByteValueInputStream retains the byte queue's array-index failures rather
// than turning a truncated in-memory state into a checked EOFException.
func NewByteValueInputStream(raw []byte) *ValueInputStream {
	return NewValueInputStreamWithoutHandles(&byteValueInputReader{bytes: raw})
}

type byteValueInputReader struct {
	bytes []byte
	index int
}

func (r *byteValueInputReader) Read(dst []byte) (int, error) {
	for i := range dst {
		index := r.index
		r.index++
		if index >= len(r.bytes) {
			panic(NewArrayIndexOutOfBoundsException(index, len(r.bytes)))
		}
		dst[i] = r.bytes[index]
	}
	return len(dst), nil
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
	return &ValueInputStream{in: reader, closer: closer, handles: make([]any, 16)}, nil
}

func NewValueInputStreamWithGlobalCompression(in io.Reader) (*ValueInputStream, error) {
	return NewValueInputStreamWithCompression(in, UseGZIP())
}

func (s *ValueInputStream) ReadShort() (int16, error) {
	var value int16
	err := binary.Read(s.in, binary.BigEndian, &value)
	return value, valueStreamReadError(err)
}

func (s *ValueInputStream) ReadInt() (int32, error) {
	var value int32
	err := binary.Read(s.in, binary.BigEndian, &value)
	return value, valueStreamReadError(err)
}

func (s *ValueInputStream) ReadLong() (int64, error) {
	var value int64
	err := binary.Read(s.in, binary.BigEndian, &value)
	return value, valueStreamReadError(err)
}

func (s *ValueInputStream) ReadByte() (byte, error) {
	var buf [1]byte
	_, err := io.ReadFull(s.in, buf[:])
	return buf[0], valueStreamReadError(err)
}

func valueStreamReadError(err error) error {
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		return NewEOFException()
	}
	return err
}

func (s *ValueInputStream) ReadFully(dst []byte) error {
	_, err := io.ReadFull(s.in, dst)
	return valueStreamReadError(err)
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

func (s *ValueInputStream) Read() (Value, error) {
	return s.readValue(false)
}

func (s *ValueInputStream) ReadExternal() (Value, error) {
	return s.readValue(true)
}

func (s *ValueInputStream) readValue(external bool) (Value, error) {
	kind, err := s.ReadByte()
	if err != nil {
		return nil, err
	}
	return s.readKind(ValueKind(kind), external)
}

func (s *ValueInputStream) readKind(kind ValueKind, external bool) (Value, error) {
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
		return s.readStringValue(external)
	case ModelValueKind:
		index, err := s.ReadShort()
		if err != nil {
			return nil, err
		}
		value := modelValueFromStream(int(index))
		if value == nil {
			return nil, nil
		}
		return value, nil
	case TupleValueKind:
		return s.readTupleValue(external)
	case SetEnumValueKind:
		return s.readSetEnumValue(external)
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
		return s.readFcnRcdValue(external)
	case RecordValueKind:
		return s.readRecordValue(external)
	case DummyValueKind:
		if s.disableHandles {
			return nil, NewWrongInvocationException(fmt.Sprintf("ValueInputStream: Can not unpickle a value of kind %d", int8(kind)))
		}
		idx, err := s.ReadNat()
		if err != nil {
			return nil, err
		}
		value := s.valueAt(int(idx))
		if value == nil {
			return nil, nil
		}
		typed, ok := value.(Value)
		if !ok {
			panic(valueStreamClassCast(value, "tlc2.value.IValue"))
		}
		return typed, nil
	default:
		return nil, NewWrongInvocationException(fmt.Sprintf("ValueInputStream: Can not unpickle a value of kind %d", int8(kind)))
	}
}

func (s *ValueInputStream) readTupleValue(external bool) (Value, error) {
	index := s.GetIndex()
	length, err := s.ReadNat()
	if err != nil {
		return nil, err
	}
	elems := make([]Value, valueStreamArrayLength(length))
	for i := range elems {
		value, err := s.readValue(external)
		if err != nil {
			return nil, err
		}
		elems[i] = value
	}
	value := NewTupleValue(elems)
	s.Assign(value, index)
	return value, nil
}

func (s *ValueInputStream) readSetEnumValue(external bool) (Value, error) {
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
	elems := make([]Value, valueStreamArrayLength(length))
	for i := range elems {
		value, err := s.readValue(external)
		if err != nil {
			return nil, err
		}
		elems[i] = value
	}
	value := NewSetEnumValue(elems, isNorm)
	s.Assign(value, index)
	return value, nil
}

func (s *ValueInputStream) readFcnRcdValue(external bool) (Value, error) {
	index := s.GetIndex()
	length, err := s.ReadNat()
	if err != nil {
		return nil, err
	}
	info, err := s.ReadByte()
	if err != nil {
		return nil, err
	}
	values := make([]Value, valueStreamArrayLength(length))
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
			values[i], err = s.readValue(external)
			if err != nil {
				return nil, err
			}
		}
		value = NewFcnRcdIntervalValue(NewIntervalValue(low, high), values)
	} else {
		domain := make([]Value, valueStreamArrayLength(length))
		for i := range domain {
			domain[i], err = s.readValue(external)
			if err != nil {
				return nil, err
			}
			values[i], err = s.readValue(external)
			if err != nil {
				return nil, err
			}
		}
		value = NewFcnRcdValue(domain, values, info == 1)
	}
	s.Assign(value, index)
	return value, nil
}

func (s *ValueInputStream) readStringValue(external bool) (Value, error) {
	var str *UniqueString
	var err error
	if external {
		str, err = s.readExternalUniqueString()
	} else {
		str, err = readJavaUniqueString(s)
	}
	if err != nil {
		return nil, err
	}
	value := NewStringValueFromUnique(str)
	s.Assign(value, s.GetIndex())
	return value, nil
}

func (s *ValueInputStream) readRecordValue(external bool) (Value, error) {
	index := s.GetIndex()
	length, err := s.ReadInt()
	if err != nil {
		return nil, err
	}
	isNorm := !external
	if length < 0 {
		length = -length
		isNorm = false
	}
	names := make([]*UniqueString, valueStreamArrayLength(length))
	values := make([]Value, valueStreamArrayLength(length))
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
			names[i] = s.GetValue(int(idx))
		} else {
			stringIndex := s.GetIndex()
			var name *UniqueString
			if external {
				name, err = s.readExternalUniqueString()
			} else {
				name, err = readJavaUniqueString(s)
			}
			if err != nil {
				return nil, err
			}
			s.Assign(name, stringIndex)
			names[i] = name
		}
		value, err := s.readValue(external)
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
	return readExternalJavaUniqueString(s)
}

func (s *ValueInputStream) GetIndex() int {
	if s.disableHandles {
		return -1
	}
	index := s.handleIndex
	if index >= len(s.handles) {
		values := make([]any, index*2)
		copy(values, s.handles[:index])
		s.handles = values
	}
	s.handleIndex++
	return index
}

func (s *ValueInputStream) Assign(value any, index int) {
	if s.disableHandles {
		return
	}
	if index < 0 || index >= len(s.handles) {
		panic(NewArrayIndexOutOfBoundsException(index, len(s.handles)))
	}
	if value != nil {
		if rv := reflect.ValueOf(value); rv.Kind() == reflect.Pointer && rv.IsNil() {
			value = nil
		}
	}
	s.handles[index] = value
}

// GetValue is the UniqueString cast exposed by Java's IValueInputStream.
func (s *ValueInputStream) GetValue(index int) *UniqueString {
	value := s.valueAt(index)
	if value == nil {
		return nil
	}
	str, ok := value.(*UniqueString)
	if !ok {
		panic(valueStreamClassCast(value, "util.UniqueString"))
	}
	return str
}

func (s *ValueInputStream) valueAt(index int) any {
	if s.disableHandles {
		panic(NewWrongInvocationException("Not supported"))
	}
	if index < 0 || index >= len(s.handles) {
		panic(NewArrayIndexOutOfBoundsException(index, len(s.handles)))
	}
	return s.handles[index]
}

func valueStreamArrayLength(length int32) int {
	if length < 0 {
		panic(NewNegativeArraySizeException(fmt.Sprint(length)))
	}
	return int(length)
}

func valueStreamClassCast(value any, target string) *ClassCastException {
	var class string
	switch value.(type) {
	case *UniqueString:
		class = "util.UniqueString"
	case *DebuggerValue:
		class = "tlc2.debug.TLCStateStackFrame$DebuggerValue"
	case *BoolValue, *IntValue, *StringValue, *ModelValue, *TupleValue, *RecordValue,
		*SetEnumValue, *IntervalValue, *FcnRcdValue, *FcnLambdaValue, *CounterExample,
		*SetOfTuplesValue, *SetOfRcdsValue, *SetOfFcnsValue, *SubsetValue, *KSubsetValue,
		*SetCupValue, *SetCapValue, *SetDiffValue, *UnionValue, *SetPredValue, *LazyValue,
		*LazySupplierValue, *OpLambdaValue, *OpRcdValue, *MethodValue, *EvaluatingValue,
		*PriorityEvaluatingValue, *CallableValue, *UndefValue, *UserValue:
		class = "tlc2.value.impl." + reflect.TypeOf(value).Elem().Name()
	default:
		// Objects not represented by this port retain their Go type name.
		class = fmt.Sprintf("%T", value)
	}
	return NewClassCastException("class " + class + " cannot be cast to class " + target +
		" (" + class + " and " + target + " are in unnamed module of loader 'app')")
}
