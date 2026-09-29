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
			"\n<modules>\n  <RootModule>Simple</RootModule>\n  <context>\n    <entry>\n",
			`<RootModule>Simple</RootModule>`,
			`<BuiltInKind>`,
			`<uniquename>TRUE</uniquename>`,
			`<OpDeclNode>`,
			`<uniquename>C</uniquename>`,
			`<kind>2</kind>`,
			`<uniquename>x</uniquename>`,
			`<kind>3</kind>`,
			`<UserDefinedOpKind>`,
			`<uniquename>A</uniquename>`,
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

	t.Run("definition references do not raise proof command level", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("ProofDefLevelXML.tla", `---- MODULE ProofDefLevelXML ----
VARIABLE x
UsesVariable == x = x
THEOREM T == TRUE
PROOF
<1>. TRUE
  BY DEF UsesVariable
<1>. QED
  OBVIOUS
====`)
		requireNoErrors(t, diags)

		got := string(xmlText)
		start := strings.Index(got, "<by>")
		if start < 0 {
			t.Fatalf("SANY XML missing BY proof\n%s", got)
		}
		end := strings.Index(got[start:], "</by>")
		if end < 0 {
			t.Fatalf("SANY XML has unterminated BY proof\n%s", got)
		}
		by := got[start : start+end]
		for _, want := range []string{
			`<level>0</level>`,
			`<facts/>`,
			`<defs>`,
			`<UserDefinedOpKindRef>`,
		} {
			if !strings.Contains(by, want) {
				t.Fatalf("BY proof XML missing %q\n%s", want, by)
			}
		}
	})

	t.Run("USE proof commands preserve expression facts before DEF", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("ProofUseFactXML.tla", `---- MODULE ProofUseFactXML ----
CONSTANT C
S == {C}
D == TRUE
THEOREM T == TRUE
PROOF
<1>. USE C \in S DEF D
<1>. QED
  OBVIOUS
====`)
		requireNoErrors(t, diags)

		got := string(xmlText)
		start := strings.Index(got, "<UseOrHideNode>")
		if start < 0 {
			t.Fatalf("SANY XML missing USE proof command\n%s", got)
		}
		end := strings.Index(got[start:], "</UseOrHideNode>")
		if end < 0 {
			t.Fatalf("SANY XML has unterminated USE proof command\n%s", got)
		}
		use := got[start : start+end]
		for _, want := range []string{
			`<begin>1</begin>`,
			`<facts>`,
			`<OpApplNode>`,
			`<BuiltInKindRef>`,
			`<defs>`,
			`<UserDefinedOpKindRef>`,
		} {
			if !strings.Contains(use, want) {
				t.Fatalf("USE proof XML missing %q\n%s", want, use)
			}
		}
		if !strings.Contains(got, `<proofLevel>1</proofLevel>`) {
			t.Fatalf("structured proof XML missing proofLevel\n%s", got)
		}
	})

	t.Run("proof step references use the current step definition level", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("ProofSelfRefXML.tla", `---- MODULE ProofSelfRefXML ----
VARIABLE x
THEOREM T == TRUE
PROOF
<1>1. CASE x' = x
  BY <1>1
<1>. QED
  BY <1>1
====`)
		requireNoErrors(t, diags)

		got := string(xmlText)
		start := strings.Index(got, "<by>")
		if start < 0 {
			t.Fatalf("SANY XML missing BY proof\n%s", got)
		}
		end := strings.Index(got[start:], "</by>")
		if end < 0 {
			t.Fatalf("SANY XML has unterminated BY proof\n%s", got)
		}
		by := got[start : start+end]
		if !strings.Contains(by, `<level>2</level>`) || !strings.Contains(by, `<TheoremDefRef>`) {
			t.Fatalf("self-referential BY proof did not use current action-level theorem definition\n%s", by)
		}
		last := strings.LastIndex(got, "<by>")
		if last == start {
			t.Fatalf("SANY XML missing QED BY proof\n%s", got)
		}
		end = strings.Index(got[last:], "</by>")
		if end < 0 {
			t.Fatalf("SANY XML has unterminated QED BY proof\n%s", got)
		}
		by = got[last : last+end]
		if !strings.Contains(by, `<level>2</level>`) || !strings.Contains(by, `<TheoremDefRef>`) {
			t.Fatalf("sibling-referential QED proof did not use local action-level theorem definition\n%s", by)
		}
	})

	t.Run("BY proof commands preserve call-style facts", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("ProofCallFactXML.tla", `---- MODULE ProofCallFactXML ----
P(n) == TRUE
THEOREM T == TRUE
PROOF
<1>. QED
  BY P(30)
====`)
		requireNoErrors(t, diags)

		got := string(xmlText)
		start := strings.Index(got, "<by>")
		if start < 0 {
			t.Fatalf("SANY XML missing BY proof\n%s", got)
		}
		end := strings.Index(got[start:], "</by>")
		if end < 0 {
			t.Fatalf("SANY XML has unterminated BY proof\n%s", got)
		}
		by := got[start : start+end]
		for _, want := range []string{
			`<begin>6</begin>`,
			`<end>10</end>`,
			`<UserDefinedOpKindRef>`,
			`<NumeralNode>`,
			`<IntValue>30</IntValue>`,
		} {
			if !strings.Contains(by, want) {
				t.Fatalf("call-style BY proof missing %q\n%s", want, by)
			}
		}
	})

	t.Run("inline body comments are not definition pre-comments", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("InlineCommentXML.tla", `---- MODULE InlineCommentXML ----
A == (* inline body comment *) TRUE
====`)
		requireNoErrors(t, diags)

		got := string(xmlText)
		if strings.Contains(got, `<pre-comments>`) || strings.Contains(got, `inline body comment`) {
			t.Fatalf("inline body comment was exported as a definition pre-comment\n%s", got)
		}
	})

	t.Run("serializes single bullet branches as SANY junction lists", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("SingleBulletXML.tla", `---- MODULE SingleBulletXML ----
VARIABLE x
A == IF x = x THEN /\ x' = x ELSE \/ TRUE
====`)
		requireNoErrors(t, diags)

		got := string(xmlText)
		for _, want := range []string{
			`<uniquename>$ConjList</uniquename>`,
			`<uniquename>$DisjList</uniquename>`,
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("single bullet XML missing %q\n%s", want, got)
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
			`<UID>105</UID>`,
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
			`<UID>83</UID>`,
			`<UID>84</UID>`,
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
		start := strings.Index(got, "<AssumeProveNode>")
		if start < 0 {
			t.Fatalf("quantified bang SANY XML missing AssumeProveNode\n%s", got)
		}
		end := strings.Index(got[start:], "</AssumeProveNode>")
		if end < 0 {
			t.Fatalf("quantified bang SANY XML has unterminated AssumeProveNode\n%s", got)
		}
		if ap := got[start : start+end]; !strings.Contains(ap, `<suffices/>`) {
			t.Fatalf("SUFFICES AssumeProveNode missing suffices marker\n%s", ap)
		}
	})
}
