package tlago

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

// CanonicalJSON parses JSON into a tree and emits compact JSON with object keys
// sorted lexicographically. Array order is preserved.
func CanonicalJSON(data []byte) ([]byte, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("multiple JSON values")
		}
		return nil, err
	}

	return canonicalJSONBytes(value), nil
}

func marshalCanonicalJSON(value any) ([]byte, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return CanonicalJSON(raw)
}

func canonicalJSONBytes(value any) []byte {
	var out bytes.Buffer
	writeCanonicalJSON(&out, value)
	return out.Bytes()
}

func writeCanonicalJSON(out *bytes.Buffer, value any) {
	switch v := value.(type) {
	case nil:
		out.WriteString("null")
	case bool:
		if v {
			out.WriteString("true")
		} else {
			out.WriteString("false")
		}
	case string:
		encoded, _ := json.Marshal(v)
		out.Write(encoded)
	case json.Number:
		out.WriteString(v.String())
	case float64:
		encoded, _ := json.Marshal(v)
		out.Write(encoded)
	case []any:
		out.WriteByte('[')
		for i, child := range v {
			if i > 0 {
				out.WriteByte(',')
			}
			writeCanonicalJSON(out, child)
		}
		out.WriteByte(']')
	case map[string]any:
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		out.WriteByte('{')
		for i, key := range keys {
			if i > 0 {
				out.WriteByte(',')
			}
			writeCanonicalJSON(out, key)
			out.WriteByte(':')
			writeCanonicalJSON(out, v[key])
		}
		out.WriteByte('}')
	default:
		encoded, _ := json.Marshal(v)
		out.Write(encoded)
	}
}
