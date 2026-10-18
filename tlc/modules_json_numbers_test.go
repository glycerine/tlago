package tlc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJsonDeserializeNumbersMatchGsonIntValue(t *testing.T) {
	cases := []struct {
		text string
		want int32
	}{
		{text: "2147483647", want: 2147483647},
		{text: "2147483648", want: -2147483648},
		{text: "4294967295", want: -1},
		{text: "4294967296", want: 0},
		{text: "9223372036854775807", want: -1},
		{text: "9223372036854775808", want: 0},
		{text: "1.9", want: 1},
		{text: "-1.9", want: -1},
		{text: "1e3", want: 1000},
		{text: "1e10", want: 1410065408},
		{text: "1e100", want: 0},
		{text: "-1e100", want: 0},
		{text: "1e9999", want: 0},
		{text: "1e-9999", want: 0},
		{text: "0.0001", want: 0},
		{text: "-0.0001", want: 0},
	}

	directory := t.TempDir()
	path := filepath.Join(directory, "number.json")
	for _, tc := range cases {
		t.Run(tc.text, func(t *testing.T) {
			if err := os.WriteFile(path, []byte(tc.text), 0o600); err != nil {
				t.Fatal(err)
			}
			value, err := JsonDeserialize(NewStringValue(path))
			if err != nil {
				t.Fatal(err)
			}
			integer, ok := value.(*IntValue)
			if !ok || integer.Val != tc.want {
				t.Fatalf("value = %T/%v, want IntValue(%d)", value, value, tc.want)
			}
		})
	}

	lines := make([]string, len(cases))
	for i, tc := range cases {
		lines[i] = tc.text
	}
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	value, err := NDJsonDeserialize(NewStringValue(path))
	if err != nil {
		t.Fatal(err)
	}
	tuple, ok := value.(*TupleValue)
	if !ok || len(tuple.Elems) != len(cases) {
		t.Fatalf("NDJSON value = %T/%v, want %d-element TupleValue", value, value, len(cases))
	}
	for i, tc := range cases {
		integer, ok := tuple.Elems[i].(*IntValue)
		if !ok || integer.Val != tc.want {
			t.Fatalf("NDJSON value %d = %T/%v, want IntValue(%d)", i, tuple.Elems[i], tuple.Elems[i], tc.want)
		}
	}
}

func TestJsonDeserializeNumberLimitsMatchGson(t *testing.T) {
	cases := []struct {
		text string
		want string
	}{
		{text: "1e10000", want: "Number has unsupported scale: 1e10000"},
		{text: "1e2147483647", want: "Number has unsupported scale: 1e2147483647"},
		{text: "1e2147483648", want: "Number has unsupported scale: 1e2147483648"},
		{text: "1e-2147483648", want: "Exponent overflow."},
		{text: "1.0e-2147483648", want: "Exponent overflow."},
		{text: "1e-2147483649", want: "Exponent overflow."},
	}
	tooLong := strings.Repeat("1", 10001)
	cases = append(cases, struct {
		text string
		want string
	}{text: tooLong, want: "Number string too large: " + tooLong[:30] + "..."})

	path := filepath.Join(t.TempDir(), "number.json")
	for _, tc := range cases {
		if err := os.WriteFile(path, []byte(tc.text), 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := JsonDeserialize(NewStringValue(path))
		failure, ok := err.(*NumberFormatException)
		if !ok || failure.Error() != tc.want {
			t.Fatalf("number %q failure = %T/%v, want NumberFormatException %q", tc.text, err, err, tc.want)
		}
	}
}
