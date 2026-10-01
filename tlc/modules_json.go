package tlc

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf16"
)

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
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		dec := json.NewDecoder(strings.NewReader(line))
		dec.UseNumber()
		value, err := jsonReadValue(dec)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return NewTupleValue(values), nil
}

func JsonSerialize(path *StringValue, value Value) (*BoolValue, error) {
	if path == nil {
		return nil, newTLCError(ECGeneral, "JsonSerialize expected a string path")
	}
	if asTupleValue(value) == nil && asRecordValue(value) == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "JsonSerialize", "sequence or record", ValuesPPR(value))
	}
	if err := ensureJSONParent(path); err != nil {
		return nil, err
	}
	var b bytes.Buffer
	if err := jsonWriteValue(&b, value); err != nil {
		return nil, err
	}
	if err := os.WriteFile(path.RawString(), b.Bytes(), 0o644); err != nil {
		return nil, err
	}
	return BoolTrue, nil
}

func NDJsonSerialize(path *StringValue, value Value) (*BoolValue, error) {
	if path == nil {
		return nil, newTLCError(ECGeneral, "ndJsonSerialize expected a string path")
	}
	tuple := asTupleValue(value)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "ndJsonSerialize", "sequence", ValuesPPR(value))
	}
	if err := ensureJSONParent(path); err != nil {
		return nil, err
	}
	file, err := os.Create(path.RawString())
	if err != nil {
		return nil, err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	if err := jsonWriteNDJSON(writer, tuple); err != nil {
		return nil, err
	}
	if err := writer.Flush(); err != nil {
		return nil, err
	}
	return BoolTrue, nil
}

func JsonTextSerialize(path *StringValue, payload Value, options Value) (*BoolValue, error) {
	opts := asRecordValue(options)
	if opts == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "third", "ndJsonSerialize", "sequence", ValuesPPR(options))
	}
	format, err := opts.Apply(NewStringValue("format"))
	if err != nil {
		return nil, err
	}
	formatString, ok := format.(*StringValue)
	if !ok || formatString.RawString() != "NDJSON" {
		return nil, nil
	}
	tuple := asTupleValue(payload)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "Serialize", "sequence", ValuesPPR(payload))
	}
	file, err := os.OpenFile(path.RawString(), ioUtilsOpenFileFlag(opts), 0o644)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	charset := ioUtilsRecordString(opts, "charset")
	if err := jsonWriteNDJSONWithCharset(file, tuple, charset); err != nil {
		return nil, err
	}
	return BoolTrue, nil
}

func ensureJSONParent(path *StringValue) error {
	parent := filepath.Dir(path.RawString())
	if parent == "." || parent == "" {
		return nil
	}
	return os.MkdirAll(parent, 0o755)
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
		if enumerable, ok := asEnumerable(value); ok {
			set, err := enumerableValueToSet(enumerable)
			if err != nil {
				return err
			}
			return jsonWriteSet(b, set)
		}
		return newTLCError(ECGeneral, "Cannot convert value: unsupported value type %T", value)
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
		return newTLCError(ECGeneral, "Cannot convert value: unsupported value type %T", value)
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
		if enumerable, ok := asEnumerable(value); ok {
			set, err := enumerableValueToSet(enumerable)
			if err != nil {
				return err
			}
			return jsonWriteSet(b, set)
		}
		return newTLCError(ECGeneral, "Cannot convert value: unsupported value type %T", value)
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

func jsonWriteNDJSON(writer *bufio.Writer, tuple *TupleValue) error {
	for _, elem := range tuple.Elems {
		var b bytes.Buffer
		if err := jsonWriteValue(&b, elem); err != nil {
			return err
		}
		if _, err := writer.Write(b.Bytes()); err != nil {
			return err
		}
		if err := writer.WriteByte('\n'); err != nil {
			return err
		}
	}
	return nil
}

func jsonWriteNDJSONWithCharset(writer io.Writer, tuple *TupleValue, charset string) error {
	var utf8 bytes.Buffer
	buffered := bufio.NewWriter(&utf8)
	if err := jsonWriteNDJSON(buffered, tuple); err != nil {
		return err
	}
	if err := buffered.Flush(); err != nil {
		return err
	}
	encoded, err := encodeJavaCharset(charset, utf8.String())
	if err != nil {
		return err
	}
	_, err = writer.Write(encoded)
	return err
}

func encodeJavaCharset(name string, text string) ([]byte, error) {
	normalized := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(name), "_", "-"))
	switch normalized {
	case "", "UTF-8", "UTF8":
		return []byte(text), nil
	case "US-ASCII", "ASCII":
		out := make([]byte, 0, len(text))
		for _, r := range text {
			if r <= 0x7f {
				out = append(out, byte(r))
			} else {
				out = append(out, '?')
			}
		}
		return out, nil
	case "ISO-8859-1", "ISO8859-1", "LATIN1", "LATIN-1":
		out := make([]byte, 0, len(text))
		for _, r := range text {
			if r <= 0xff {
				out = append(out, byte(r))
			} else {
				out = append(out, '?')
			}
		}
		return out, nil
	case "UTF-16":
		encoded := encodeUTF16(text, true)
		return append([]byte{0xfe, 0xff}, encoded...), nil
	case "UTF-16BE":
		return encodeUTF16(text, true), nil
	case "UTF-16LE":
		return encodeUTF16(text, false), nil
	default:
		return nil, newTLCError(ECGeneral, "Unsupported charset: %s", name)
	}
}

func encodeUTF16(text string, bigEndian bool) []byte {
	units := utf16.Encode([]rune(text))
	out := make([]byte, 0, len(units)*2)
	for _, unit := range units {
		if bigEndian {
			out = append(out, byte(unit>>8), byte(unit))
		} else {
			out = append(out, byte(unit), byte(unit>>8))
		}
	}
	return out
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
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}
	b.Write(encoded)
	return nil
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

func enumerableValueToSet(enumerable Enumerable) (*SetEnumValue, error) {
	values := NewValueVec(0)
	enum := enumerable.Elements()
	for {
		elem := enum.NextElement()
		if elem == nil {
			if err := enum.Err(); err != nil {
				return nil, err
			}
			break
		}
		values.Add(elem)
	}
	return NewSetEnumValueVec(values, false), nil
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
