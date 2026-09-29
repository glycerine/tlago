package tlago

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanyXMLBehaviors(t *testing.T) {
	t.Run("serializes a checked module as Java-shaped SANY XML", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("Simple.tla", `---- MODULE Simple ----
CONSTANT C
VARIABLE x
A == TRUE
====`)
		requireNoErrors(t, diags)

		got := string(xmlText)
		for _, want := range []string{
			`<?xml version="1.0" encoding="UTF-8" standalone="no"?>`,
			`<RootModule>Simple</RootModule>`,
			`<BuiltInKind>`,
			`<uniquename>TRUE</uniquename>`,
			`<OpDeclNode>`,
			`<uniquename>C</uniquename><arity>0</arity><kind>2</kind>`,
			`<uniquename>x</uniquename><arity>0</arity><kind>3</kind>`,
			`<UserDefinedOpKind>`,
			`<uniquename>A</uniquename><arity>0</arity>`,
			`<ModuleNodeRef>`,
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("SANY XML missing %q\n%s", want, got)
			}
		}
	})

	t.Run("CLI writes SANY XML", func(t *testing.T) {
		dir := t.TempDir()
		spec := filepath.Join(dir, "CliXML.tla")
		writeFile(t, spec, `---- MODULE CliXML ----
VARIABLE x
Init == x = 0
====`)

		var stdout, stderr bytes.Buffer
		if code := RunCLI([]string{"sany-xml", spec}, &stdout, &stderr); code != ExitOK {
			t.Fatalf("sany-xml exit = %d, want %d; stderr=%s", code, ExitOK, stderr.String())
		}
		if !strings.Contains(stdout.String(), `<RootModule>CliXML</RootModule>`) ||
			!strings.Contains(stdout.String(), `<ModuleNode`) ||
			!strings.Contains(stdout.String(), `<UserDefinedOpKind`) {
			t.Fatalf("sany-xml stdout = %q, want SANY XML", stdout.String())
		}
	})

	t.Run("serializes proof steps as theorem XML", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("ProofXML.tla", `---- MODULE ProofXML ----
THEOREM T == TRUE
PROOF
<1>1. TRUE
  OBVIOUS
<1>2. CASE TRUE
  OBVIOUS
<1>. QED
  BY <1>1
====`)
		requireNoErrors(t, diags)

		got := string(xmlText)
		for _, want := range []string{
			`<steps>`,
			`<TheoremDefNode>`,
			`<uniquename>&lt;1&gt;1</uniquename>`,
			`<TheoremNodeRef>`,
			`<obvious>`,
			`<by>`,
			`<uniquename>$Pfcase</uniquename>`,
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("proof SANY XML missing %q\n%s", want, got)
			}
		}
	})

	t.Run("serializes EXCEPT @ as Java-shaped AtNode XML", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("ExceptAtXML.tla", `---- MODULE ExceptAtXML ----
VARIABLE x
Next == x' = [x EXCEPT ![1] = @ + 1]
====`)
		requireNoErrors(t, diags)

		got := string(xmlText)
		for _, want := range []string{
			`<AtNode>`,
			`<BuiltInKindRef><UID>105</UID></BuiltInKindRef>`,
			`<uniquename>$Seq</uniquename>`,
			`<uniquename>$Except</uniquename>`,
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("EXCEPT @ SANY XML missing %q\n%s", want, got)
			}
		}
	})

	t.Run("serializes aligned junction bullets as SANY list operators", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("JunctionXML.tla", `---- MODULE JunctionXML ----
A == /\ TRUE
     /\ \/ TRUE
        \/ FALSE
     /\ FALSE
====`)
		requireNoErrors(t, diags)

		got := string(xmlText)
		for _, want := range []string{
			`<uniquename>$ConjList</uniquename>`,
			`<uniquename>$DisjList</uniquename>`,
			`<BuiltInKindRef><UID>83</UID></BuiltInKindRef>`,
			`<BuiltInKindRef><UID>84</UID></BuiltInKindRef>`,
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("junction SANY XML missing %q\n%s", want, got)
			}
		}
	})

	t.Run("serializes bang application of quantified definitions as local LAMBDA", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("QuantifiedBangXML.tla", `---- MODULE QuantifiedBangXML ----
CONSTANT S
Pred(i) == i = i
Inv == \A i \in S: Pred(i)
THEOREM T == TRUE
PROOF
<1>. SUFFICES ASSUME NEW q \in S PROVE Inv!(q)
  OBVIOUS
<1>. QED
  OBVIOUS
====`)
		requireNoErrors(t, diags)

		got := string(xmlText)
		for _, want := range []string{
			`<uniquename>LAMBDA</uniquename>`,
			`<UserDefinedOpKindRef>`,
			`<uniquename>i</uniquename>`,
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("quantified bang SANY XML missing %q\n%s", want, got)
			}
		}
	})
}
