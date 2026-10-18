package tlc

import (
	"bytes"
	"io"
	"math/big"
	"strconv"
	"strings"
)

// Only the three source synchronized static writers acquire Json.class.
// Evaluation may recursively invoke another writer on the same Java thread.
var jsonClassMonitor distributedServerMonitor

const jsonUnsupportedNullMessage = "Cannot convert value: unsupported JSON value null"

func JsonToJson(value Value) (*StringValue, error) {
	var b bytes.Buffer
	if err := jsonWriteValue(&b, value); err != nil {
		return nil, err
	}
	return NewStringValue(b.String()), nil
}

func JsonToJsonArray(value Value) (*StringValue, error) {
	var b bytes.Buffer
	if err := jsonWriteArray(&b, value); err != nil {
		return nil, err
	}
	return NewStringValue(b.String()), nil
}

func JsonToJsonObject(value Value) (*StringValue, error) {
	var b bytes.Buffer
	if err := jsonWriteObject(&b, value); err != nil {
		return nil, err
	}
	return NewStringValue(b.String()), nil
}

func JsonDeserialize(path *StringValue) (Value, error) {
	if path == nil {
		return nil, NewNullPointerException(`Cannot read field "val" because "path" is null`)
	}
	file, err := openFileInputStream(path.RawString())
	if err != nil {
		return nil, err
	}
	raw, err := io.ReadAll(file)
	if err != nil {
		return nil, bufferedRandomAccessFileIOError(err)
	}
	text, err := javaCharsetDecode(raw, javaDefaultCharset(), true)
	if err != nil {
		return nil, err
	}
	value, err := parseGsonJSON(text)
	if err == nil && value == nil {
		return nil, NewIOException(jsonUnsupportedNullMessage)
	}
	return value, err
}

func NDJsonDeserialize(path *StringValue) (result Value, err error) {
	if path == nil {
		return nil, NewNullPointerException(`Cannot read field "val" because "path" is null`)
	}
	file, err := openFileInputStream(path.RawString())
	if err != nil {
		return nil, err
	}
	defer func() {
		failure := recover()
		closeErr := bufferedRandomAccessFileIOError(file.Close())
		if failure != nil {
			if primary, ok := failure.(error); ok {
				javaSuppressCloseError(primary, closeErr)
			}
			panic(failure)
		}
		if closeErr != nil {
			if err == nil {
				result = nil
				err = closeErr
			} else {
				javaSuppressCloseError(err, closeErr)
			}
		}
	}()
	reader, err := newJavaReplacingLineReader(file, javaDefaultCharset())
	if err != nil {
		return nil, err
	}

	values := make([]Value, 0)
	for {
		line, readErr := reader.ReadLine()
		if readErr != nil {
			return nil, bufferedRandomAccessFileIOError(readErr)
		}
		if line == nil {
			break
		}
		trimmed := strings.TrimFunc(*line, func(char rune) bool { return char <= ' ' })
		if trimmed != "" {
			value, err := parseGsonJSON(trimmed)
			if err != nil {
				return nil, err
			}
			values = append(values, value)
		}
	}
	return NewTupleValue(values), nil
}

func javaBufferedReaderLines(text string) []string {
	lines := make([]string, 0)
	start := 0
	for index := 0; index < len(text); index++ {
		if text[index] != '\n' && text[index] != '\r' {
			continue
		}
		lines = append(lines, text[start:index])
		if text[index] == '\r' && index+1 < len(text) && text[index+1] == '\n' {
			index++
		}
		start = index + 1
	}
	if start < len(text) {
		lines = append(lines, text[start:])
	}
	return lines
}

func JsonSerialize(path *StringValue, value Value) (*BoolValue, error) {
	jsonClassMonitor.Lock()
	defer jsonClassMonitor.Unlock()
	if value == nil {
		return nil, NewNullPointerException(`Cannot invoke "tlc2.value.impl.Value.toTuple()" because "v" is null`)
	}
	if asTupleValue(value) == nil && asRecordValue(value) == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "JsonSerialize", "sequence or record", ValuesPPR(value))
	}
	// Java checks the converted shape, then serializes the original value.
	if err := jsonWriteOrdinaryFile(path, []Value{value}, false); err != nil {
		return nil, err
	}
	return BoolTrue, nil
}

func NDJsonSerialize(path *StringValue, value Value) (*BoolValue, error) {
	jsonClassMonitor.Lock()
	defer jsonClassMonitor.Unlock()
	if value == nil {
		return nil, NewNullPointerException(`Cannot invoke "tlc2.value.impl.Value.toTuple()" because "v" is null`)
	}
	tuple := asTupleValue(value)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "ndJsonSerialize", "sequence", ValuesPPR(value))
	}
	if err := jsonWriteOrdinaryFile(path, tuple.GetElems(), true); err != nil {
		return nil, err
	}
	return BoolTrue, nil
}

func JsonTextSerialize(path *StringValue, payload Value, options Value) (*BoolValue, error) {
	jsonClassMonitor.Lock()
	defer jsonClassMonitor.Unlock()
	if options == nil {
		return nil, NewNullPointerException()
	}
	opts := asRecordValue(options)
	if opts == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "third", "ndJsonSerialize", "sequence", ValuesPPR(options))
	}
	format, err := ioUtilsTXTFormat(opts)
	if err != nil {
		return nil, err
	}
	if format != "NDJSON" {
		return nil, nil
	}
	if payload == nil {
		return nil, NewNullPointerException()
	}
	tuple := asTupleValue(payload)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "Serialize", "sequence", ValuesPPR(payload))
	}
	return jsonTextSerializeTuple(path, tuple, opts)
}

func jsonTextSerializeTuple(path *StringValue, tuple *TupleValue, opts *RecordValue) (*BoolValue, error) {
	openOptions, charset, err := ioUtilsTXTOptions(opts)
	if err != nil {
		return nil, err
	}
	if err := jsonWriteNDJSONFile(path, tuple, openOptions, charset); err != nil {
		return nil, err
	}
	return BoolTrue, nil
}

func jsonWriteValue(b *bytes.Buffer, value Value) error {
	switch v := value.(type) {
	case *RecordValue:
		return jsonWriteRecord(b, v)
	case *CounterExample:
		return jsonWriteRecord(b, asRecordValue(v))
	case *TupleValue:
		return jsonWriteTuple(b, v)
	case *StringValue:
		return jsonWriteString(b, v.RawString())
	case *DebuggerValue:
		return jsonWriteString(b, v.RawString())
	case *ModelValue:
		return jsonWriteString(b, v.String())
	case *IntValue:
		b.WriteString(strconv.FormatInt(int64(v.ownedVal()), 10))
		return nil
	case *BoolValue:
		if v.Val {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
		return nil
	case *FcnRcdValue:
		return jsonWriteFcn(b, v)
	case *FcnLambdaValue:
		fcn := v.ToFcnRcd()
		return jsonWriteFcn(b, fcn)
	case *SetEnumValue:
		return jsonWriteSet(b, v)
	default:
		if _, ok := asEnumerable(value); ok {
			// Java invokes the value's toSetEnum, preserving its cache,
			// allocation size, ordering and coverage semantics.
			set, err := toSetEnumValue(value)
			if err != nil {
				return err
			}
			return jsonWriteSet(b, set)
		}
		return jsonUnsupportedValue(value)
	}
}

func jsonWriteObject(b *bytes.Buffer, value Value) error {
	switch v := value.(type) {
	case *RecordValue:
		return jsonWriteRecord(b, v)
	case *CounterExample:
		return jsonWriteRecord(b, asRecordValue(v))
	case *TupleValue:
		b.WriteByte('{')
		for i, elem := range v.GetElems() {
			if i > 0 {
				b.WriteByte(',')
			}
			if err := jsonWriteString(b, strconv.Itoa(i)); err != nil {
				return err
			}
			b.WriteByte(':')
			if err := jsonWriteValue(b, elem); err != nil {
				return err
			}
		}
		b.WriteByte('}')
		return nil
	case *FcnRcdValue:
		return jsonWriteFcn(b, v)
	case *FcnLambdaValue:
		fcn := v.ToFcnRcd()
		return jsonWriteFcn(b, fcn)
	default:
		return jsonUnsupportedValue(value)
	}
}

func jsonWriteArray(b *bytes.Buffer, value Value) error {
	switch v := value.(type) {
	case *TupleValue:
		return jsonWriteTuple(b, v)
	case *FcnRcdValue:
		if jsonFcnIsSequence(v) {
			return jsonWriteFcnArray(b, v)
		}
		return jsonWriteFcnObject(b, v)
	case *FcnLambdaValue:
		fcn := v.ToFcnRcd()
		return jsonWriteArray(b, fcn)
	case *SetEnumValue:
		return jsonWriteSet(b, v)
	default:
		if _, ok := asEnumerable(value); ok {
			// Java invokes the value's toSetEnum, preserving its cache,
			// allocation size, ordering and coverage semantics.
			set, err := toSetEnumValue(value)
			if err != nil {
				return err
			}
			return jsonWriteSet(b, set)
		}
		return jsonUnsupportedValue(value)
	}
}

func jsonWriteRecord(b *bytes.Buffer, value *RecordValue) error {
	b.WriteByte('{')
	for i, name := range value.ownedNames() {
		if i > 0 {
			b.WriteByte(',')
		}
		if err := jsonWriteString(b, name.String()); err != nil {
			return err
		}
		b.WriteByte(':')
		if err := jsonWriteValue(b, value.ownedValues()[i]); err != nil {
			return err
		}
	}
	b.WriteByte('}')
	return nil
}

func jsonWriteTuple(b *bytes.Buffer, value *TupleValue) error {
	b.WriteByte('[')
	for i, elem := range value.GetElems() {
		if i > 0 {
			b.WriteByte(',')
		}
		if err := jsonWriteValue(b, elem); err != nil {
			return err
		}
	}
	b.WriteByte(']')
	return nil
}

func jsonWriteFcn(b *bytes.Buffer, value *FcnRcdValue) error {
	if jsonFcnIsSequence(value) {
		return jsonWriteFcnArray(b, value)
	}
	return jsonWriteFcnObject(b, value)
}

func jsonWriteFcnObject(b *bytes.Buffer, value *FcnRcdValue) error {
	domain := value.GetDomainAsValues()
	b.WriteByte('{')
	for i, dval := range domain {
		if i > 0 {
			b.WriteByte(',')
		}
		key := dval.String()
		if sv, ok := asStringValue(dval); ok {
			key = sv.RawString()
		}
		if err := jsonWriteString(b, key); err != nil {
			return err
		}
		b.WriteByte(':')
		if err := jsonWriteValue(b, value.ownedValues()[i]); err != nil {
			return err
		}
	}
	b.WriteByte('}')
	return nil
}

func jsonWriteFcnArray(b *bytes.Buffer, value *FcnRcdValue) error {
	value.Normalize()
	b.WriteByte('[')
	for i, elem := range value.ownedValues() {
		if i > 0 {
			b.WriteByte(',')
		}
		if err := jsonWriteValue(b, elem); err != nil {
			return err
		}
	}
	b.WriteByte(']')
	return nil
}

func jsonWriteSet(b *bytes.Buffer, value *SetEnumValue) error {
	value.Normalize()
	b.WriteByte('[')
	for i := 0; i < value.Elems.Len(); i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		if err := jsonWriteValue(b, value.Elems.At(i)); err != nil {
			return err
		}
	}
	b.WriteByte(']')
	return nil
}

func jsonWriteString(b *bytes.Buffer, value string) error {
	// JsonElement.toString uses Gson JsonWriter without HTML-safe escaping.
	// Keep Java UTF-16 units, including unmatched surrogates, for the writer.
	units := []uint16{'"'}
	const hex = "0123456789abcdef"
	for _, unit := range javaStringUTF16(value) {
		switch unit {
		case '"', '\\':
			units = append(units, '\\', unit)
		case '\b':
			units = append(units, '\\', 'b')
		case '\f':
			units = append(units, '\\', 'f')
		case '\n':
			units = append(units, '\\', 'n')
		case '\r':
			units = append(units, '\\', 'r')
		case '\t':
			units = append(units, '\\', 't')
		default:
			if unit < 0x20 || unit == 0x2028 || unit == 0x2029 {
				units = append(units, '\\', 'u', uint16(hex[unit>>12]), uint16(hex[unit>>8&15]), uint16(hex[unit>>4&15]), uint16(hex[unit&15]))
			} else {
				units = append(units, unit)
			}
		}
	}
	units = append(units, '"')
	b.WriteString(javaStringFromUTF16(units))
	return nil
}

func jsonUnsupportedValue(value Value) error {
	if value == nil {
		return NewNullPointerException(`Cannot invoke "tlc2.value.IValue.getClass()" because "value" is null`)
	}
	return NewIOException("Cannot convert value: unsupported value type " + javaValueClassName(value))
}

func jsonFcnIsSequence(value *FcnRcdValue) bool {
	domain := value.GetDomainAsValues()
	for _, dval := range domain {
		if _, ok := dval.(*IntValue); !ok {
			return false
		}
	}
	value.Normalize()
	for i, dval := range domain {
		if dval.(*IntValue).ownedVal() != int32(i+1) {
			return false
		}
	}
	return true
}

// Gson's LazilyParsedNumber.intValue first tries Integer.parseInt, then
// Long.parseLong, and finally BigDecimal.intValue. The final conversion drops
// the fractional part toward zero and retains the low 32 bits.
func jsonNumberToJavaInt(text string) (int32, error) {
	if value, err := strconv.ParseInt(text, 10, 32); err == nil {
		return int32(value), nil
	}
	if value, err := strconv.ParseInt(text, 10, 64); err == nil {
		return int32(value), nil
	}
	// Gson's NumberLimits applies only after both primitive integer parses fail.
	if len(text) > 10000 {
		return 0, NewNumberFormatExceptionNullable(javaString("Number string too large: " + text[:30] + "..."))
	}

	negative := strings.HasPrefix(text, "-")
	unsigned := strings.TrimPrefix(text, "-")
	exponent := int64(0)
	if index := strings.IndexAny(unsigned, "eE"); index >= 0 {
		parsed, err := strconv.ParseInt(unsigned[index+1:], 10, 64)
		if err != nil {
			return 0, NewNumberFormatExceptionNullable(javaString("Exponent overflow."))
		}
		exponent = parsed
		unsigned = unsigned[:index]
	}
	fractionalDigits := int64(0)
	if index := strings.IndexByte(unsigned, '.'); index >= 0 {
		fractionalDigits = int64(len(unsigned) - index - 1)
		unsigned = unsigned[:index] + unsigned[index+1:]
	}
	scale := fractionalDigits - exponent
	if scale < -1<<31 || scale > 1<<31-1 {
		return 0, NewNumberFormatExceptionNullable(javaString("Exponent overflow."))
	}
	if scale <= -10000 || scale >= 10000 {
		return 0, NewNumberFormatExceptionNullable(javaString("Number has unsupported scale: " + text))
	}
	coefficient, ok := new(big.Int).SetString(unsigned, 10)
	if !ok {
		return 0, NewNumberFormatException(text)
	}
	power := -scale
	if power >= 0 {
		coefficient.Mul(coefficient, new(big.Int).Exp(big.NewInt(10), big.NewInt(power), nil))
	} else {
		coefficient.Quo(coefficient, new(big.Int).Exp(big.NewInt(10), big.NewInt(-power), nil))
	}
	if negative {
		coefficient.Neg(coefficient)
	}
	modulus := new(big.Int).Lsh(big.NewInt(1), 32)
	coefficient.Mod(coefficient, modulus)
	return int32(uint32(coefficient.Uint64())), nil
}
