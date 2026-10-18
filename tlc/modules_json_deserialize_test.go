package tlc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJsonDeserializeOpenFailuresMatchJava(t *testing.T) {
	deserializers := []struct {
		name string
		call func(*StringValue) (Value, error)
	}{
		{name: "JsonDeserialize", call: JsonDeserialize},
		{name: "ndJsonDeserialize", call: NDJsonDeserialize},
	}

	for _, deserializer := range deserializers {
		t.Run(deserializer.name+"/null", func(t *testing.T) {
			_, err := deserializer.call(nil)
			failure, ok := err.(*NullPointerException)
			want := `Cannot read field "val" because "path" is null`
			if !ok || failure.Error() != want {
				t.Fatalf("failure = %T/%v, want NullPointerException %q", err, err, want)
			}
		})

		t.Run(deserializer.name+"/missing", func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "missing.json")
			_, err := deserializer.call(NewStringValue(path))
			failure, ok := err.(*FileNotFoundException)
			want := path + " (No such file or directory)"
			if !ok || failure.Error() != want {
				t.Fatalf("failure = %T/%v, want FileNotFoundException %q", err, err, want)
			}
		})

		t.Run(deserializer.name+"/invalid", func(t *testing.T) {
			_, err := deserializer.call(NewStringValue("bad\x00path"))
			failure, ok := err.(*FileNotFoundException)
			if !ok || failure.Error() != "Invalid file path" {
				t.Fatalf("failure = %T/%v, want FileNotFoundException %q", err, err, "Invalid file path")
			}
		})

		t.Run(deserializer.name+"/directory", func(t *testing.T) {
			path := t.TempDir()
			_, err := deserializer.call(NewStringValue(path))
			failure, ok := err.(*FileNotFoundException)
			want := path + " (Is a directory)"
			if !ok || failure.Error() != want {
				t.Fatalf("failure = %T/%v, want FileNotFoundException %q", err, err, want)
			}
		})
	}
}

func TestJsonNullConversionFailuresMatchJava(t *testing.T) {
	conversions := []struct {
		name string
		call func(Value) (*StringValue, error)
	}{
		{name: "toJson", call: JsonToJson},
		{name: "toJsonArray", call: JsonToJsonArray},
		{name: "toJsonObject", call: JsonToJsonObject},
	}
	want := `Cannot invoke "tlc2.value.IValue.getClass()" because "value" is null`
	for _, conversion := range conversions {
		t.Run(conversion.name, func(t *testing.T) {
			_, err := conversion.call(nil)
			failure, ok := err.(*NullPointerException)
			if !ok || failure.Error() != want {
				t.Fatalf("failure = %T/%v, want NullPointerException %q", err, err, want)
			}
		})
	}
}

func TestJsonSerializeNullFailuresMatchJava(t *testing.T) {
	serializers := []struct {
		name string
		call func(*StringValue, Value) (*BoolValue, error)
	}{
		{name: "serialize", call: JsonSerialize},
		{name: "ndSerialize", call: NDJsonSerialize},
	}

	for _, serializer := range serializers {
		t.Run(serializer.name+"/value precedes path", func(t *testing.T) {
			_, err := serializer.call(nil, nil)
			failure, ok := err.(*NullPointerException)
			want := `Cannot invoke "tlc2.value.impl.Value.toTuple()" because "v" is null`
			if !ok || failure.Error() != want {
				t.Fatalf("failure = %T/%v, want NullPointerException %q", err, err, want)
			}
		})

		t.Run(serializer.name+"/path", func(t *testing.T) {
			_, err := serializer.call(nil, NewTupleValue(nil))
			failure, ok := err.(*NullPointerException)
			want := `Cannot read field "val" because "path" is null`
			if !ok || failure.Error() != want {
				t.Fatalf("failure = %T/%v, want NullPointerException %q", err, err, want)
			}
		})
	}
}

func TestJsonDeserializeNullLiteralMatchesJava(t *testing.T) {
	path := filepath.Join(t.TempDir(), "null.json")
	if err := os.WriteFile(path, []byte("null"), 0o600); err != nil {
		t.Fatal(err)
	}
	deserializers := []struct {
		name string
		call func(*StringValue) (Value, error)
	}{
		{name: "deserialize", call: JsonDeserialize},
		{name: "ndDeserialize", call: NDJsonDeserialize},
	}
	for _, deserializer := range deserializers {
		t.Run(deserializer.name, func(t *testing.T) {
			_, err := deserializer.call(NewStringValue(path))
			failure, ok := err.(*IOException)
			want := "Cannot convert value: unsupported JSON value null"
			if !ok || failure.Error() != want {
				t.Fatalf("failure = %T/%v, want IOException %q", err, err, want)
			}
		})
	}
}

func TestJsonDeserializeEmptyDocumentMatchesGson(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.json")
	for name, content := range map[string]string{"empty": "", "whitespace": " \t\r\n"} {
		t.Run(name, func(t *testing.T) {
			if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}
			_, err := JsonDeserialize(NewStringValue(path))
			failure, ok := err.(*IOException)
			if !ok || failure.Error() != jsonUnsupportedNullMessage {
				t.Fatalf("failure = %T/%v, want IOException %q", err, err, jsonUnsupportedNullMessage)
			}
		})
	}

	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	value, err := NDJsonDeserialize(NewStringValue(path))
	if err != nil {
		t.Fatal(err)
	}
	tuple, ok := value.(*TupleValue)
	if !ok || len(tuple.Elems) != 0 {
		t.Fatalf("empty NDJSON = %T/%v, want empty TupleValue", value, value)
	}
}
