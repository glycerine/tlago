package tlc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGsonJSONParserFailuresMatchJava(t *testing.T) {
	const help = "\nSee https://github.com/google/gson/blob/main/Troubleshooting.md#malformed-json"
	cases := []struct {
		name, input, causeClass, causeMessage string
	}{
		{"trailing number", "1 2", "com.google.gson.stream.MalformedJsonException", "Use JsonReader.setStrictness(Strictness.LENIENT) to accept malformed JSON at line 1 column 4 path $" + help},
		{"trailing array", "{} []", "com.google.gson.stream.MalformedJsonException", "Use JsonReader.setStrictness(Strictness.LENIENT) to accept malformed JSON at line 1 column 5 path $" + help},
		{"trailing comment", "1 //x", "com.google.gson.stream.MalformedJsonException", "Use JsonReader.setStrictness(Strictness.LENIENT) to accept malformed JSON at line 1 column 4 path $" + help},
		{"array eof", "[1", "java.io.EOFException", "End of input at line 1 column 3 path $[1]"},
		{"object eof", `{"a":1`, "java.io.EOFException", "End of input at line 1 column 7 path $.a"},
		{"missing value", `{"a":}`, "com.google.gson.stream.MalformedJsonException", "Expected value at line 1 column 6 path $.a" + help},
		{"unterminated string", `"x`, "com.google.gson.stream.MalformedJsonException", "Unterminated string at line 1 column 3 path $" + help},
		{"trailing object comma", `{"a":1,}`, "com.google.gson.stream.MalformedJsonException", "Expected name at line 1 column 9 path $.a" + help},
		{"missing object comma", "{\n\"a\": 1\n\"b\": 2\n}", "com.google.gson.stream.MalformedJsonException", "Unterminated object at line 3 column 2 path $.a" + help},
		{"invalid escape", `"a\q"`, "com.google.gson.stream.MalformedJsonException", "Invalid escape sequence at line 1 column 5 path $" + help},
		{"unicode escape", `"\uZZZZ"`, "com.google.gson.stream.MalformedJsonException", "Malformed Unicode escape \\uZZZZ at line 1 column 4 path $" + help},
		{"utf16 column", "[\"😀\" 2]", "com.google.gson.stream.MalformedJsonException", "Unterminated array at line 1 column 8 path $[1]" + help},
		{"missing colon", "{a;1}", "com.google.gson.stream.MalformedJsonException", "Expected ':' at line 1 column 4 path $.a" + help},
		{"unterminated comment", "/*x", "com.google.gson.stream.MalformedJsonException", "Unterminated comment at line 1 column 3 path $" + help},
		{"initial bom column", "\ufeffa/b", "com.google.gson.stream.MalformedJsonException", "Use JsonReader.setStrictness(Strictness.LENIENT) to accept malformed JSON at line 1 column 3 path $" + help},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseGsonJSON(tc.input)
			syntax, ok := err.(*JsonSyntaxException)
			if !ok {
				t.Fatalf("error = %T/%v, want JsonSyntaxException", err, err)
			}
			cause := syntax.GetCause()
			if got := javaThrowableClassName(cause); got != tc.causeClass {
				t.Fatalf("cause class = %q, want %q", got, tc.causeClass)
			}
			if cause.Error() != tc.causeMessage {
				t.Fatalf("cause message = %q, want %q", cause.Error(), tc.causeMessage)
			}
			wantOuter := tc.causeClass + ": " + tc.causeMessage
			if syntax.Error() != wantOuter || javaThrowableClassName(syntax) != "com.google.gson.JsonSyntaxException" {
				t.Fatalf("outer = %q/%q, want JsonSyntaxException/%q", javaThrowableClassName(syntax), syntax.Error(), wantOuter)
			}
		})
	}
}

func TestGsonJSONParserLenientFormsMatchJava(t *testing.T) {
	cases := []struct {
		input, want string
	}{
		{"{a:1}", "[a |-> 1]"},
		{"01", `"01"`},
		{"/*c*/[1]", "<<1>>"},
		{"[1;2]", "<<1, 2>>"},
		{"{'a'=1}", "[a |-> 1]"},
		{`{"a"=>1}`, "[a |-> 1]"},
		{"#c\ntrue", "TRUE"},
		{")]}'\n{\"a\":1}", "[a |-> 1]"},
		{"  )]}'\n{\"a\":1}", "[a |-> 1]"},
		{"\ufeff{\"a\":1}", "[a |-> 1]"},
		{"TRUE", "TRUE"},
		{"NaN", `"NaN"`},
		{"Infinity", `"Infinity"`},
		{"-.5", `"-.5"`},
		{"1.", `"1."`},
		{"// c\nfalse", "FALSE"},
		{"[1/*x*/,2]", "<<1, 2>>"},
		{"{\"a\":1,\"a\":2}", "[a |-> 2]"},
		{"{a:1;b:2}", "[a |-> 1, b |-> 2]"},
		{"'x'", `"x"`},
		{"\"a\\\nb\"", `"a\nb"`},
		{" \ufeff1", "\u0022\uFEFF1\u0022"},
	}
	for _, tc := range cases {
		value, err := parseGsonJSON(tc.input)
		if err != nil {
			t.Fatalf("parse %q: %v", tc.input, err)
		}
		if got := ValueToString(value, "", true); got != tc.want {
			t.Fatalf("parse %q = %q, want %q", tc.input, got, tc.want)
		}
	}

	for _, input := range []string{"[1,]", "[1,,2]", "[;]", "[true,False,NULL]", "null 1", "null/*x*/2"} {
		_, err := parseGsonJSON(input)
		if _, ok := err.(*IOException); !ok || err.Error() != jsonUnsupportedNullMessage {
			t.Fatalf("parse %q error = %T/%v, want IOException %q", input, err, err, jsonUnsupportedNullMessage)
		}
	}
}

func TestJSONDeserializersUseGsonDocumentBoundary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "trailing.json")
	if err := os.WriteFile(path, []byte("1 2\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, call := range []struct {
		name string
		fn   func(*StringValue) (Value, error)
	}{{"ordinary", JsonDeserialize}, {"ndjson", NDJsonDeserialize}} {
		t.Run(call.name, func(t *testing.T) {
			_, err := call.fn(NewStringValue(path))
			if javaThrowableClassName(err) != "com.google.gson.JsonSyntaxException" {
				t.Fatalf("error = %T/%v", err, err)
			}
		})
	}
}

func TestNDJSONUsesJavaReadLineAndTrimBoundaries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lines.ndjson")
	if err := os.WriteFile(path, []byte("1\r2\r\n3\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	value, err := NDJsonDeserialize(NewStringValue(path))
	if err != nil {
		t.Fatal(err)
	}
	if got := ValueToString(value, "", true); got != "<<1, 2, 3>>" {
		t.Fatalf("value = %q", got)
	}
}
