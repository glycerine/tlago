package tlago

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestApalacheIRBehaviors(t *testing.T) {
	t.Run("serializes simple modules with declarations and operators", func(t *testing.T) {
		jsonText, diags := ApalacheIRJSONSource("Simple.tla", `---- MODULE Simple ----
EXTENDS Naturals
CONSTANT C
VARIABLE x
A == 1 + 2
B(y) == IF y \in {1, 2} THEN TRUE ELSE FALSE
ASSUME C = 3
====`, ApalacheIROptions{})
		requireNoErrors(t, diags)

		got := compactJSON(t, jsonText)
		want := compactJSON(t, []byte(`{
		  "name": "ApalacheIR",
		  "version": "1.0",
		  "description": "https://apalache-mc.org/docs/adr/005adr-json.html",
		  "modules": [
		    {
		      "kind": "TlaModule",
		      "name": "Simple",
		      "declarations": [
		        {"type": "Untyped", "kind": "TlaConstDecl", "name": "C"},
		        {"type": "Untyped", "kind": "TlaVarDecl", "name": "x"},
		        {
		          "type": "Untyped",
		          "kind": "TlaOperDecl",
		          "name": "A",
		          "formalParams": [],
		          "isRecursive": false,
		          "body": {
		            "type": "Untyped",
		            "kind": "OperEx",
		            "oper": "PLUS",
		            "args": [
		              {"type": "Untyped", "kind": "ValEx", "value": {"kind": "TlaInt", "value": 1}},
		              {"type": "Untyped", "kind": "ValEx", "value": {"kind": "TlaInt", "value": 2}}
		            ]
		          }
		        },
		        {
		          "type": "Untyped",
		          "kind": "TlaOperDecl",
		          "name": "B",
		          "formalParams": [{"kind": "OperParam", "name": "y", "arity": 0}],
		          "isRecursive": false,
		          "body": {
		            "type": "Untyped",
		            "kind": "OperEx",
		            "oper": "IF_THEN_ELSE",
		            "args": [
		              {
		                "type": "Untyped",
		                "kind": "OperEx",
		                "oper": "SET_IN",
		                "args": [
		                  {"type": "Untyped", "kind": "NameEx", "name": "y"},
		                  {
		                    "type": "Untyped",
		                    "kind": "OperEx",
		                    "oper": "SET_ENUM",
		                    "args": [
		                      {"type": "Untyped", "kind": "ValEx", "value": {"kind": "TlaInt", "value": 1}},
		                      {"type": "Untyped", "kind": "ValEx", "value": {"kind": "TlaInt", "value": 2}}
		                    ]
		                  }
		                ]
		              },
		              {"type": "Untyped", "kind": "ValEx", "value": {"kind": "TlaBool", "value": true}},
		              {"type": "Untyped", "kind": "ValEx", "value": {"kind": "TlaBool", "value": false}}
		            ]
		          }
		        },
		        {
		          "type": "Untyped",
		          "kind": "TlaAssumeDecl",
		          "body": {
		            "type": "Untyped",
		            "kind": "OperEx",
		            "oper": "EQ",
		            "args": [
		              {"type": "Untyped", "kind": "NameEx", "name": "C"},
		              {"type": "Untyped", "kind": "ValEx", "value": {"kind": "TlaInt", "value": 3}}
		            ]
		          }
		        }
		      ]
		    }
		  ]
		}`))
		if got != want {
			t.Fatalf("ApalacheIR JSON mismatch\n got: %s\nwant: %s", got, want)
		}
	})

	t.Run("CLI writes ApalacheIR JSON", func(t *testing.T) {
		dir := t.TempDir()
		spec := filepath.Join(dir, "CliJson.tla")
		writeFile(t, spec, `---- MODULE CliJson ----
VARIABLE x
Init == x = 0
====`)

		var stdout, stderr bytes.Buffer
		if code := RunCLI([]string{"apalache-json", spec}, &stdout, &stderr); code != ExitOK {
			t.Fatalf("apalache-json exit = %d, want %d; stderr=%s", code, ExitOK, stderr.String())
		}
		if !strings.Contains(stdout.String(), `"name":"ApalacheIR"`) ||
			!strings.Contains(stdout.String(), `"kind":"TlaModule"`) ||
			!strings.Contains(stdout.String(), `"name":"CliJson"`) {
			t.Fatalf("apalache-json stdout = %q, want ApalacheIR module JSON", stdout.String())
		}
		if got := strings.TrimSpace(stdout.String()); !strings.HasPrefix(got, `{"description":`) {
			t.Fatalf("apalache-json stdout is not canonical key-sorted JSON: %q", stdout.String())
		}
	})

	t.Run("serializes SANY prefix minus as Apalache unary minus", func(t *testing.T) {
		jsonText, diags := ApalacheIRJSONSource("UnaryMinus.tla", `---- MODULE UnaryMinus ---- EXTENDS Integers
A == -1
B == -(1 + 2)
====`, ApalacheIROptions{})
		requireNoErrors(t, diags)

		got := string(jsonText)
		if count := strings.Count(got, `"oper":"UNARY_MINUS"`); count != 2 {
			t.Fatalf("UNARY_MINUS count = %d, want 2\n%s", count, got)
		}
	})

	t.Run("serializes source ranges when requested", func(t *testing.T) {
		jsonText, diags := ApalacheIRJSONSource("SourceLocal.tla", `---- MODULE SourceLocal ----
EXTENDS Naturals
CONSTANT C
A == 1 + 2
====`, ApalacheIROptions{IncludeSource: true})
		requireNoErrors(t, diags)

		var root struct {
			Modules []struct {
				Declarations []struct {
					Name   string `json:"name"`
					Source struct {
						Filename string `json:"filename"`
						From     struct {
							Line   int `json:"line"`
							Column int `json:"column"`
						} `json:"from"`
						To struct {
							Line   int `json:"line"`
							Column int `json:"column"`
						} `json:"to"`
					} `json:"source"`
					Body struct {
						Source struct {
							From struct {
								Line   int `json:"line"`
								Column int `json:"column"`
							} `json:"from"`
							To struct {
								Line   int `json:"line"`
								Column int `json:"column"`
							} `json:"to"`
						} `json:"source"`
					} `json:"body"`
				} `json:"declarations"`
			} `json:"modules"`
		}
		if err := json.Unmarshal(jsonText, &root); err != nil {
			t.Fatalf("unmarshal ApalacheIR JSON: %v\n%s", err, jsonText)
		}
		if got := root.Modules[0].Declarations[0].Source.Filename; got != "SourceLocal" {
			t.Fatalf("source filename = %q, want module-style SourceLocal", got)
		}
		bodySource := root.Modules[0].Declarations[1].Body.Source
		if bodySource.From.Line != 4 || bodySource.From.Column != 6 || bodySource.To.Line != 4 || bodySource.To.Column != 10 {
			t.Fatalf("operator body source = %#v, want line 4 columns 6..10", bodySource)
		}
	})

	t.Run("materializes instance definitions for ApalacheIR", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Inner.tla"), `---- MODULE Inner ----
CONSTANT C
Op == C
====`)
		rootPath := filepath.Join(dir, "InstanceLocal.tla")
		writeFile(t, rootPath, `---- MODULE InstanceLocal ----
CONSTANT X
INSTANCE Inner WITH C <- X
Use == Op
====`)

		spec, diags := LoadSanySpec(rootPath, LoadOptions{})
		requireNoErrors(t, diags)
		jsonText, irDiags := ApalacheIRJSON(spec, ApalacheIROptions{})
		requireNoErrors(t, irDiags)

		var root struct {
			Modules []struct {
				Declarations []struct {
					Name string `json:"name"`
					Body struct {
						Kind string `json:"kind"`
						Name string `json:"name"`
						Oper string `json:"oper"`
					} `json:"body"`
				} `json:"declarations"`
			} `json:"modules"`
		}
		if err := json.Unmarshal(jsonText, &root); err != nil {
			t.Fatalf("unmarshal ApalacheIR JSON: %v\n%s", err, jsonText)
		}
		if got, want := root.Modules[0].Declarations[1].Name, "Op"; got != want {
			t.Fatalf("instance declaration name = %q, want %q", got, want)
		}
		if got, want := root.Modules[0].Declarations[1].Body.Name, "X"; got != want {
			t.Fatalf("instance declaration body name = %q, want substitution %q", got, want)
		}
		if got, want := root.Modules[0].Declarations[2].Body.Oper, "OPER_APP"; got != want {
			t.Fatalf("Use body operator = %q, want %q", got, want)
		}
	})

	t.Run("sorts declarations by define-before-use dependencies", func(t *testing.T) {
		spec, diags := CheckSanySource("Sorted.tla", `---- MODULE Sorted ----
EXTENDS Naturals
B == 1
A == B + 1
====`)
		requireNoErrors(t, diags)
		// Exercise exporter sorting with an unordered IR input after checking
		// valid SANY source. Java rejects forward references in TLA+ source.
		spec.Root.Definitions[0], spec.Root.Definitions[1] = spec.Root.Definitions[1], spec.Root.Definitions[0]
		jsonText, diags := ApalacheIRJSON(spec, ApalacheIROptions{})
		requireNoErrors(t, diags)

		var root struct {
			Modules []struct {
				Declarations []struct {
					Kind string `json:"kind"`
					Name string `json:"name"`
				} `json:"declarations"`
			} `json:"modules"`
		}
		if err := json.Unmarshal(jsonText, &root); err != nil {
			t.Fatalf("unmarshal ApalacheIR JSON: %v\n%s", err, jsonText)
		}
		var names []string
		for _, decl := range root.Modules[0].Declarations {
			names = append(names, decl.Name)
		}
		if got, want := strings.Join(names, ","), "B,A"; got != want {
			t.Fatalf("declaration order = %s, want %s", got, want)
		}
	})

	t.Run("marks recursive operator declarations", func(t *testing.T) {
		jsonText, diags := ApalacheIRJSONSource("Recursive.tla", `---- MODULE Recursive ----
EXTENDS Naturals
RECURSIVE Fact(_)
Fact(n) == IF n = 0 THEN 1 ELSE n * Fact(n - 1)
====`, ApalacheIROptions{})
		requireNoErrors(t, diags)

		var root struct {
			Modules []struct {
				Declarations []struct {
					Name        string `json:"name"`
					IsRecursive bool   `json:"isRecursive"`
				} `json:"declarations"`
			} `json:"modules"`
		}
		if err := json.Unmarshal(jsonText, &root); err != nil {
			t.Fatalf("unmarshal ApalacheIR JSON: %v\n%s", err, jsonText)
		}
		if got := root.Modules[0].Declarations[0].IsRecursive; !got {
			t.Fatalf("Fact isRecursive = %v, want true", got)
		}
	})

	t.Run("uses recursive function constructor for recursive function definitions", func(t *testing.T) {
		jsonText, diags := ApalacheIRJSONSource("RecursiveFunction.tla", `---- MODULE RecursiveFunction ----
EXTENDS Naturals
RECURSIVE Fact
Fact[n \in 0..2] == IF n = 0 THEN 1 ELSE n * Fact[n - 1]
====`, ApalacheIROptions{})
		requireNoErrors(t, diags)

		var root struct {
			Modules []struct {
				Declarations []struct {
					Name string `json:"name"`
					Body struct {
						Oper string `json:"oper"`
					} `json:"body"`
				} `json:"declarations"`
			} `json:"modules"`
		}
		if err := json.Unmarshal(jsonText, &root); err != nil {
			t.Fatalf("unmarshal ApalacheIR JSON: %v\n%s", err, jsonText)
		}
		if got, want := root.Modules[0].Declarations[0].Body.Oper, "FUN_REC_CTOR"; got != want {
			t.Fatalf("recursive function constructor = %s, want %s", got, want)
		}
	})

	t.Run("preserves named assumptions", func(t *testing.T) {
		jsonText, diags := ApalacheIRJSONSource("NamedAssume.tla", `---- MODULE NamedAssume ----
CONSTANT C
ASSUME NonZero == C /= 0
====`, ApalacheIROptions{})
		requireNoErrors(t, diags)

		var root struct {
			Modules []struct {
				Declarations []struct {
					Kind string `json:"kind"`
					Name string `json:"name"`
				} `json:"declarations"`
			} `json:"modules"`
		}
		if err := json.Unmarshal(jsonText, &root); err != nil {
			t.Fatalf("unmarshal ApalacheIR JSON: %v\n%s", err, jsonText)
		}
		if got, want := root.Modules[0].Declarations[1].Name, "NonZero"; got != want {
			t.Fatalf("assumption name = %q, want %q", got, want)
		}
	})

	t.Run("flattens Cartesian products for ApalacheIR", func(t *testing.T) {
		jsonText, diags := ApalacheIRJSONSource("Times.tla", `---- MODULE Times ----
VARIABLE x
CartesianProd == x \X x \X x
====`, ApalacheIROptions{})
		requireNoErrors(t, diags)

		var root struct {
			Modules []struct {
				Declarations []struct {
					Name string `json:"name"`
					Body struct {
						Oper string            `json:"oper"`
						Args []json.RawMessage `json:"args"`
					} `json:"body"`
				} `json:"declarations"`
			} `json:"modules"`
		}
		if err := json.Unmarshal(jsonText, &root); err != nil {
			t.Fatalf("unmarshal ApalacheIR JSON: %v\n%s", err, jsonText)
		}
		body := root.Modules[0].Declarations[1].Body
		if body.Oper != "SET_TIMES" || len(body.Args) != 3 {
			t.Fatalf("CartesianProd body = %s/%d args, want SET_TIMES/3", body.Oper, len(body.Args))
		}
	})

	t.Run("flattens associative boolean operators for ApalacheIR", func(t *testing.T) {
		jsonText, diags := ApalacheIRJSONSource("BooleanChains.tla", `---- MODULE BooleanChains ----
And ==
  /\ TRUE
  /\ FALSE
  /\ TRUE
Or ==
  \/ TRUE
  \/ FALSE
  \/ TRUE
====`, ApalacheIROptions{})
		requireNoErrors(t, diags)

		var root struct {
			Modules []struct {
				Declarations []struct {
					Name string `json:"name"`
					Body struct {
						Oper string            `json:"oper"`
						Args []json.RawMessage `json:"args"`
					} `json:"body"`
				} `json:"declarations"`
			} `json:"modules"`
		}
		if err := json.Unmarshal(jsonText, &root); err != nil {
			t.Fatalf("unmarshal ApalacheIR JSON: %v\n%s", err, jsonText)
		}
		for _, decl := range root.Modules[0].Declarations {
			switch decl.Name {
			case "And":
				if decl.Body.Oper != "AND" || len(decl.Body.Args) != 3 {
					t.Fatalf("And body = %s/%d args, want AND/3", decl.Body.Oper, len(decl.Body.Args))
				}
			case "Or":
				if decl.Body.Oper != "OR" || len(decl.Body.Args) != 3 {
					t.Fatalf("Or body = %s/%d args, want OR/3", decl.Body.Oper, len(decl.Body.Args))
				}
			}
		}
	})

	t.Run("serializes structured expression operators", func(t *testing.T) {
		jsonText, diags := ApalacheIRJSONSource("Structured.tla", `---- MODULE Structured ----
EXTENDS Naturals
VARIABLE x
Set == {1, 2}
Tuple == <<1, 2>>
Record == [a |-> 1, b |-> 2]
RecordSet == [a: {1}, b: {2}]
Fcn == [i \in {1, 2} |-> i + 1]
App == Fcn[1]
Filtered == {i \in {1, 2}: i > 1}
Mapped == {i + 1 : i \in {1, 2}}
Local == LET Inc(i) == i + 1 IN Inc(1)
All == \A i \in {1, 2}: i \in {1, 2}
Choice == CHOOSE i \in {1, 2}: i > 1
All2 == \A j: TRUE
Exists2 == \E j: TRUE
Choice2 == CHOOSE j: TRUE
TemporalExists == \EE j: TRUE
TemporalForall == \AA j: TRUE
SquareAct == [TRUE]_x
AngleAct == <<TRUE>>_x
Fair == WF_x(TRUE)
====`, ApalacheIROptions{})
		requireNoErrors(t, diags)

		var root struct {
			Modules []struct {
				Declarations []struct {
					Body json.RawMessage `json:"body"`
				} `json:"declarations"`
			} `json:"modules"`
		}
		if err := json.Unmarshal(jsonText, &root); err != nil {
			t.Fatalf("unmarshal ApalacheIR JSON: %v\n%s", err, jsonText)
		}
		operators := map[string]bool{}
		for _, decl := range root.Modules[0].Declarations {
			collectJSONOperators(t, decl.Body, operators)
		}
		for _, oper := range []string{
			"SET_ENUM", "TUPLE", "RECORD", "RECORD_SET", "FUN_CTOR", "FUN_APP",
			"SET_FILTER", "SET_MAP", "LET", "OPER_APP", "FORALL3", "CHOOSE3",
			"FORALL2", "EXISTS2", "CHOOSE2", "TEMPORAL_EXISTS", "TEMPORAL_FORALL",
			"STUTTER", "NO_STUTTER", "WEAK_FAIRNESS",
		} {
			if !operators[oper] {
				t.Fatalf("operators %v missing %s in %s", operators, oper, jsonText)
			}
		}
	})
}

func compactJSON(t *testing.T, data []byte) string {
	t.Helper()
	canon, err := CanonicalJSON(data)
	if err != nil {
		t.Fatalf("canonical JSON: %v\n%s", err, data)
	}
	return string(canon)
}

func collectJSONOperators(t *testing.T, data []byte, operators map[string]bool) {
	t.Helper()
	if len(data) == 0 {
		return
	}
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatalf("unmarshal expression JSON: %v\n%s", err, data)
	}
	collectJSONOperatorsValue(value, operators)
}

func collectJSONOperatorsValue(value any, operators map[string]bool) {
	switch v := value.(type) {
	case map[string]any:
		if oper, ok := v["oper"].(string); ok {
			operators[oper] = true
		}
		if v["kind"] == "LetInEx" {
			operators["LET"] = true
		}
		for _, child := range v {
			collectJSONOperatorsValue(child, operators)
		}
	case []any:
		for _, child := range v {
			collectJSONOperatorsValue(child, operators)
		}
	}
}
