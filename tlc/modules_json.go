package tlc

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strconv"
	"strings"
)

// Only the three source synchronized static writers acquire Json.class.
// Evaluation may recursively invoke another writer on the same Java thread.
var jsonClassMonitor distributedServerMonitor

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
		return nil, newTLCError(ECGeneral, "JsonDeserialize expected a string path")
	}
	file, err := os.Open(path.RawString())
	if err != nil {
		return nil, err
	}
	defer file.Close()

	dec := json.NewDecoder(file)
	dec.UseNumber()
	return jsonReadValue(dec)
}

func NDJsonDeserialize(path *StringValue) (Value, error) {
	if path == nil {
		return nil, newTLCError(ECGeneral, "ndJsonDeserialize expected a string path")
	}
	file, err := os.Open(path.RawString())
	if err != nil {
		return nil, err
	}
	defer file.Close()

	values := make([]Value, 0)
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, err
		}
		if len(line) != 0 {
			line = strings.TrimSpace(line)
			if line != "" {
				dec := json.NewDecoder(strings.NewReader(line))
				dec.UseNumber()
				value, err := jsonReadValue(dec)
				if err != nil {
					return nil, err
				}
				values = append(values, value)
			}
		}
		if err == io.EOF {
			break
		}
	}
	return NewTupleValue(values), nil
}

func JsonSerialize(path *StringValue, value Value) (*BoolValue, error) {
	jsonClassMonitor.Lock()
	defer jsonClassMonitor.Unlock()
	if value == nil {
		return nil, NewNullPointerException()
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
		return nil, NewNullPointerException()
	}
	tuple := asTupleValue(value)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "ndJsonSerialize", "sequence", ValuesPPR(value))
	}
	if err := jsonWriteOrdinaryFile(path, tuple.Elems, true); err != nil {
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
		b.WriteString(strconv.FormatInt(int64(v.Val), 10))
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
		for i, elem := range v.Elems {
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
	for i, name := range value.Names {
		if i > 0 {
			b.WriteByte(',')
		}
		if err := jsonWriteString(b, name.String()); err != nil {
			return err
		}
		b.WriteByte(':')
		if err := jsonWriteValue(b, value.Values[i]); err != nil {
			return err
		}
	}
	b.WriteByte('}')
	return nil
}

func jsonWriteTuple(b *bytes.Buffer, value *TupleValue) error {
	b.WriteByte('[')
	for i, elem := range value.Elems {
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
	domain := value.DomainAsValues()
	b.WriteByte('{')
	for i, dval := range domain {
		if i > 0 {
			b.WriteByte(',')
		}
		key := dval.String()
		if sv, ok := dval.(*StringValue); ok {
			key = sv.RawString()
		}
		if err := jsonWriteString(b, key); err != nil {
			return err
		}
		b.WriteByte(':')
		if err := jsonWriteValue(b, value.Values[i]); err != nil {
			return err
		}
	}
	b.WriteByte('}')
	return nil
}

func jsonWriteFcnArray(b *bytes.Buffer, value *FcnRcdValue) error {
	value.Normalize()
	b.WriteByte('[')
	for i, elem := range value.Values {
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
		return NewNullPointerException()
	}
	return NewIOException("Cannot convert value: unsupported value type " + javaValueClassName(value))
}

func jsonFcnIsSequence(value *FcnRcdValue) bool {
	domain := value.DomainAsValues()
	for _, dval := range domain {
		if _, ok := dval.(*IntValue); !ok {
			return false
		}
	}
	value.Normalize()
	for i, dval := range domain {
		if dval.(*IntValue).Val != int32(i+1) {
			return false
		}
	}
	return true
}

func jsonReadValue(dec *json.Decoder) (Value, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	return jsonTokenToTLC(dec, tok)
}

func jsonTokenToTLC(dec *json.Decoder, tok json.Token) (Value, error) {
	switch v := tok.(type) {
	case json.Delim:
		switch v {
		case '[':
			values := make([]Value, 0)
			for dec.More() {
				value, err := jsonReadValue(dec)
				if err != nil {
					return nil, err
				}
				values = append(values, value)
			}
			if err := expectJSONDelim(dec, ']'); err != nil {
				return nil, err
			}
			return NewTupleValue(values), nil
		case '{':
			names := make([]*UniqueString, 0)
			values := make([]Value, 0)
			indices := make(map[string]int)
			for dec.More() {
				keyTok, err := dec.Token()
				if err != nil {
					return nil, err
				}
				key, ok := keyTok.(string)
				if !ok {
					return nil, newTLCError(ECGeneral, "Cannot convert value: unsupported JSON object key %v", keyTok)
				}
				value, err := jsonReadValue(dec)
				if err != nil {
					return nil, err
				}
				if idx, ok := indices[key]; ok {
					values[idx] = value
				} else {
					indices[key] = len(names)
					names = append(names, UniqueStringOf(key))
					values = append(values, value)
				}
			}
			if err := expectJSONDelim(dec, '}'); err != nil {
				return nil, err
			}
			return NewRecordValue(names, values, false), nil
		}
		return nil, newTLCError(ECGeneral, "Cannot convert value: unexpected JSON delimiter %q", v)
	case json.Number:
		i, err := strconv.ParseInt(v.String(), 10, 32)
		if err != nil {
			return nil, err
		}
		return NewIntValue(int32(i)), nil
	case string:
		return NewStringValue(v), nil
	case bool:
		return NewBoolValue(v), nil
	case nil:
		return nil, newTLCError(ECGeneral, "Cannot convert value: unsupported JSON value null")
	default:
		return nil, newTLCError(ECGeneral, "Cannot convert value: unsupported JSON value %v", tok)
	}
}

func expectJSONDelim(dec *json.Decoder, want json.Delim) error {
	tok, err := dec.Token()
	if err != nil {
		if err == io.EOF {
			return newTLCError(ECGeneral, "Cannot convert value: missing JSON delimiter %q", want)
		}
		return err
	}
	got, ok := tok.(json.Delim)
	if !ok || got != want {
		return newTLCError(ECGeneral, "Cannot convert value: expected JSON delimiter %q, got %v", want, tok)
	}
	return nil
}
