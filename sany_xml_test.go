package tlago

import (
	"bytes"
	"path/filepath"
	"strconv"
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

	t.Run("resolves LOCAL INSTANCE symbols while exporting the module body", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Helper.tla"), `---- MODULE Helper ----
Zero == 0
====`)
		root := filepath.Join(dir, "LocalInstanceXML.tla")
		writeFile(t, root, `---- MODULE LocalInstanceXML ----
\* Helper import
LOCAL INSTANCE Helper
Use == Zero
====`)

		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
		xmlText, xmlDiags := SanyXML(spec)
		requireNoErrors(t, xmlDiags)
		got := string(xmlText)
		for _, want := range []string{
			`<InstanceNode>`,
			`<module>Helper</module>`,
			`<uniquename>Zero</uniquename>`,
			`<pre-comments><![CDATA[\* Helper import]]></pre-comments>`,
			`<local/>`,
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("SANY XML did not include LOCAL INSTANCE detail %q\n%s", want, got)
			}
		}
	})

	t.Run("instance-cloned definitions reuse original LET-local operators", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Helper.tla"), `---- MODULE Helper ----
A(S) ==
  LET F[i \in S] == i
  IN F[1]
====`)
		root := filepath.Join(dir, "LetInstanceXML.tla")
		writeFile(t, root, `---- MODULE LetInstanceXML ----
LOCAL INSTANCE Helper
B == A({1})
====`)

		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
		xmlText, xmlDiags := SanyXML(spec)
		requireNoErrors(t, xmlDiags)
		got := string(xmlText)
		if count := strings.Count(got, `<uniquename>F</uniquename>`); count != 1 {
			t.Fatalf("LET-local operator F appears %d times, want Java SANY-style single original definition\n%s", count, got)
		}
	})

	t.Run("extended modules contribute their instance nodes", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Helper.tla"), `---- MODULE Helper ----
Zero == 0
====`)
		writeFile(t, filepath.Join(dir, "Base.tla"), `---- MODULE Base ----
LOCAL INSTANCE Helper
Use == Zero
====`)
		root := filepath.Join(dir, "ExtendsInstanceXML.tla")
		writeFile(t, root, `---- MODULE ExtendsInstanceXML ----
EXTENDS Base
RootUse == Use
====`)

		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
		xmlText, xmlDiags := SanyXML(spec)
		requireNoErrors(t, xmlDiags)
		if count := strings.Count(string(xmlText), `<module>Helper</module>`); count != 2 {
			t.Fatalf("Helper instance nodes = %d, want one on Base and one imported into ExtendsInstanceXML\n%s", count, xmlText)
		}
	})

	t.Run("sorts extended module names like Java SANY XML", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "A.tla"), `---- MODULE A ----
A == TRUE
====`)
		writeFile(t, filepath.Join(dir, "B.tla"), `---- MODULE B ----
B == TRUE
====`)
		root := filepath.Join(dir, "ExtendsOrderXML.tla")
		writeFile(t, root, `---- MODULE ExtendsOrderXML ----
EXTENDS B, A
Root == A /\ B
====`)

		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
		xmlText, xmlDiags := SanyXML(spec)
		requireNoErrors(t, xmlDiags)
		got := string(xmlText)
		moduleStart := strings.Index(got, `<uniquename>ExtendsOrderXML</uniquename>`)
		if moduleStart < 0 {
			t.Fatalf("SANY XML missing ExtendsOrderXML module\n%s", got)
		}
		aIndex := strings.Index(got[moduleStart:], `<uniquename>A</uniquename>`)
		bIndex := strings.Index(got[moduleStart:], `<uniquename>B</uniquename>`)
		if aIndex < 0 || bIndex < 0 || aIndex > bIndex {
			t.Fatalf("extends names were not sorted as A then B\n%s", got[moduleStart:])
		}
	})

	t.Run("serializes implicit same-name instance substitutions", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Base.tla"), `---- MODULE Base ----
CONSTANT C
VARIABLE x
Use == x = C
====`)
		root := filepath.Join(dir, "InstanceSubstXML.tla")
		writeFile(t, root, `---- MODULE InstanceSubstXML ----
CONSTANT C
VARIABLE x
INSTANCE Base
RootUse == Use
====`)

		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
		xmlText, xmlDiags := SanyXML(spec)
		requireNoErrors(t, xmlDiags)
		got := string(xmlText)
		if count := strings.Count(got, `<Subst>`); count != 4 {
			t.Fatalf("instance substitutions = %d, want C and x substitutions on the instance and cloned body\n%s", count, got)
		}
		if !strings.Contains(got, `<SubstInNode>`) {
			t.Fatalf("instance-cloned definition body did not include SubstInNode\n%s", got)
		}
	})

	t.Run("serializes multi-index function application through a tuple operand", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("MultiIndexFunctionAppXML.tla", `---- MODULE MultiIndexFunctionAppXML ----
VARIABLE f
A(i, j) == f[i, j]
====`)
		requireNoErrors(t, diags)

		got := string(xmlText)
		if !strings.Contains(got, `<uniquename>A</uniquename>`) {
			t.Fatalf("SANY XML missing A definition\n%s", got)
		}
		for _, want := range []string{
			`<uniquename>$FcnApply</uniquename>`,
			`<uniquename>$Tuple</uniquename>`,
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("multi-index function application XML missing %q\n%s", want, got)
			}
		}
	})

	t.Run("serializes multi-index EXCEPT selectors through a tuple operand", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("MultiIndexExceptXML.tla", `---- MODULE MultiIndexExceptXML ----
VARIABLE f
A(i, j) == [f EXCEPT ![i, j] = 0]
====`)
		requireNoErrors(t, diags)

		got := string(xmlText)
		for _, want := range []string{
			`<uniquename>$Except</uniquename>`,
			`<uniquename>$Tuple</uniquename>`,
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("multi-index EXCEPT XML missing %q\n%s", want, got)
			}
		}
	})

	t.Run("bounded operators include bound set level", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("BoundedLevelXML.tla", `---- MODULE BoundedLevelXML ----
VARIABLE S
A == \A x \in S : TRUE
====`)
		requireNoErrors(t, diags)

		got := string(xmlText)
		a := strings.Index(got, `<uniquename>A</uniquename>`)
		if a < 0 {
			t.Fatalf("SANY XML missing A definition\n%s", got)
		}
		body := strings.Index(got[a:], `<body>`)
		if body < 0 {
			t.Fatalf("SANY XML missing A body\n%s", got[a:])
		}
		levelStart := strings.Index(got[a+body:], `<level>`)
		if levelStart < 0 {
			t.Fatalf("SANY XML missing bounded quantifier level\n%s", got[a+body:])
		}
		levelEnd := strings.Index(got[a+body+levelStart:], `</level>`)
		level := got[a+body+levelStart : a+body+levelStart+levelEnd+len(`</level>`)]
		if level != `<level>1</level>` {
			t.Fatalf("bounded quantifier level = %s, want <level>1</level>\n%s", level, got[a+body:])
		}
	})

	t.Run("LET node level includes local definition levels", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("LetLevelXML.tla", `---- MODULE LetLevelXML ----
VARIABLE v
A == LET F == v IN F
====`)
		requireNoErrors(t, diags)

		got := string(xmlText)
		let := strings.Index(got, `<LetInNode>`)
		if let < 0 {
			t.Fatalf("SANY XML missing LetInNode\n%s", got)
		}
		levelStart := strings.Index(got[let:], `<level>`)
		if levelStart < 0 {
			t.Fatalf("SANY XML missing LetInNode level\n%s", got[let:])
		}
		levelEnd := strings.Index(got[let+levelStart:], `</level>`)
		level := got[let+levelStart : let+levelStart+levelEnd+len(`</level>`)]
		if level != `<level>1</level>` {
			t.Fatalf("LetInNode level = %s, want <level>1</level>\n%s", level, got[let:])
		}
	})

	t.Run("serializes named assumptions as AssumeDef", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("NamedAssumeXML.tla", `---- MODULE NamedAssumeXML ----
CONSTANT C
ASSUME CAssumption == C = C
====`)
		requireNoErrors(t, diags)

		got := string(xmlText)
		for _, want := range []string{
			`<AssumeDef>`,
			`<uniquename>CAssumption</uniquename>`,
			`<AssumeDefRef>`,
			`<AssumeNode>`,
			`<definition>`,
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("named assumption XML missing %q\n%s", want, got)
			}
		}
	})

	t.Run("serializes LAMBDA expressions as local operator definitions", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("LambdaExprXML.tla", `---- MODULE LambdaExprXML ----
VARIABLE x
ChooseOne(S, P(_)) == CHOOSE y \in S : P(y)
A == ChooseOne({1}, LAMBDA y : x = x)
====`)
		requireNoErrors(t, diags)

		got := string(xmlText)
		for _, want := range []string{
			`<uniquename>LAMBDA</uniquename>`,
			`<OpArgNode>`,
			`<uniquename>y</uniquename>`,
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("lambda expression XML missing %q\n%s", want, got)
			}
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

	t.Run("serializes prefix minus as the SANY unary minus operator", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("PrefixMinusXML.tla", `---- MODULE PrefixMinusXML ----
-. x == x
A == -1
====`)
		requireNoErrors(t, diags)

		got := string(xmlText)
		def := strings.Index(got, `<uniquename>A</uniquename>`)
		if def < 0 {
			t.Fatalf("SANY XML missing definition A\n%s", got)
		}
		body := got[def:]
		for _, want := range []string{
			`<UserDefinedOpKindRef>`,
			`<IntValue>1</IntValue>`,
		} {
			if !strings.Contains(body, want) {
				t.Fatalf("prefix minus XML missing %q\n%s", want, body)
			}
		}
		if !strings.Contains(got, `<uniquename>-.</uniquename>`) {
			t.Fatalf("prefix minus XML missing unary minus definition\n%s", got)
		}
	})

	t.Run("serializes Cartesian product as SANY CartesianProd", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("CartesianProductXML.tla", `---- MODULE CartesianProductXML ----
A == {1} \X {2}
====`)
		requireNoErrors(t, diags)

		got := string(xmlText)
		if !strings.Contains(got, `<uniquename>$CartesianProd</uniquename>`) {
			t.Fatalf("Cartesian product XML missing $CartesianProd\n%s", got)
		}
	})

	t.Run("does not reuse reserved builtin UIDs for generated symbols", func(t *testing.T) {
		var source strings.Builder
		source.WriteString("---- MODULE ReservedBuiltinUIDXML ----\n")
		source.WriteString("CONSTANTS ")
		for i := 0; i < 90; i++ {
			if i > 0 {
				source.WriteString(", ")
			}
			source.WriteString("C")
			source.WriteString(strconv.Itoa(i))
		}
		source.WriteString("\nA == {1} \\X {2}\n====")
		xmlText, diags := SanyXMLSource("ReservedBuiltinUIDXML.tla", source.String())
		requireNoErrors(t, diags)

		got := string(xmlText)
		if strings.Count(got, "<entry>\n      <UID>231</UID>") != 1 {
			t.Fatalf("reserved Cartesian product UID was reused by a generated entry\n%s", got)
		}
		if !strings.Contains(got, "<UID>231</UID>\n      <BuiltInKind>") {
			t.Fatalf("reserved Cartesian product UID did not belong to the builtin entry\n%s", got)
		}
	})

	t.Run("wraps self-referential function definitions as SANY recursive function specs", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("RecursiveFunctionSpecXML.tla", `---- MODULE RecursiveFunctionSpecXML ----
Cardinality(S) ==
  LET CS[T \in SUBSET S] == IF T = {} THEN 0 ELSE CS[T]
  IN CS[S]
====`)
		requireNoErrors(t, diags)

		got := string(xmlText)
		for _, want := range []string{
			`<uniquename>$RecursiveFcnSpec</uniquename>`,
			`<UID>253</UID>`,
			`<uniquename>$IfThenElse</uniquename>`,
			`<unbound>`,
			`<uniquename>CS</uniquename>`,
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("recursive function spec XML missing %q\n%s", want, got)
			}
		}
		if strings.Contains(got, `<uniquename>$FcnConstructor</uniquename>`) {
			t.Fatalf("recursive function spec XML should use the function body directly, not a $FcnConstructor wrapper\n%s", got)
		}
	})

	t.Run("distinguishes bounded function definitions from function expressions", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("NonRecursiveFunctionSpecXML.tla", `---- MODULE NonRecursiveFunctionSpecXML ----
CONSTANT S
F[x \in S] == x
A == [x \in S |-> x]
====`)
		requireNoErrors(t, diags)

		got := string(xmlText)
		for _, want := range []string{
			`<uniquename>$NonRecursiveFcnSpec</uniquename>`,
			`<uniquename>$FcnConstructor</uniquename>`,
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("function XML missing %q\n%s", want, got)
			}
		}
		f := strings.Index(got, `<uniquename>F</uniquename>`)
		if f < 0 || !strings.Contains(got, `<uniquename>A</uniquename>`) {
			t.Fatalf("function XML missing F or A\n%s", got)
		}
		fEnd := strings.Index(got[f:], `</UserDefinedOpKind>`)
		if fEnd < 0 {
			t.Fatalf("function XML missing end of F definition\n%s", got[f:])
		}
		fBody := got[f : f+fEnd]
		if strings.Contains(fBody, `<uniquename>$FcnConstructor</uniquename>`) {
			t.Fatalf("bounded function definition F should use $NonRecursiveFcnSpec, not $FcnConstructor\n%s", fBody)
		}
	})

	t.Run("serializes recursive definitions with a SANY recursive section", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("RecursiveSectionXML.tla", `---- MODULE RecursiveSectionXML ----
RECURSIVE F(_)
F(n) == IF n = 0 THEN TRUE ELSE LET y == F(n - 1) IN y
RECURSIVE G(_)
G(n) == IF n = 0 THEN TRUE ELSE F(n - 1)
====`)
		requireNoErrors(t, diags)

		got := string(xmlText)
		for _, want := range []string{
			`<uniquename>F</uniquename>`,
			`<recursive/>`,
			`<recursiveSection>1</recursiveSection>`,
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("recursive definition XML missing %q\n%s", want, got)
			}
		}
		g := strings.Index(got, `<uniquename>G</uniquename>`)
		if g < 0 {
			t.Fatalf("recursive definition XML missing definition G\n%s", got)
		}
		if !strings.Contains(got[g:], `<recursiveSection>2</recursiveSection>`) {
			t.Fatalf("second recursive definition did not use recursive section 2\n%s", got[g:])
		}
		y := strings.Index(got, `<uniquename>y</uniquename>`)
		if y < 0 {
			t.Fatalf("recursive definition XML missing local definition y\n%s", got)
		}
		yEntryEnd := strings.Index(got[y:], `</UserDefinedOpKind>`)
		if yEntryEnd < 0 {
			t.Fatalf("recursive local definition XML has unterminated entry\n%s", got[y:])
		}
		yEntry := got[y : y+yEntryEnd]
		if strings.Contains(yEntry, `<recursive/>`) || !strings.Contains(yEntry, `<recursiveSection>1</recursiveSection>`) {
			t.Fatalf("recursive local definition did not inherit only recursive section 1\n%s", yEntry)
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
