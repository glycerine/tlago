package tlago

import (
	"bytes"
	"os"
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

	t.Run("emits external module refs in SANY semantic order", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "ExtBase.tla"), `---- MODULE ExtBase ----
BaseOp == TRUE
====`)
		writeFile(t, filepath.Join(dir, "Extender.tla"), `---- MODULE Extender ----
EXTENDS ExtBase
ExtOp == BaseOp
====`)
		writeFile(t, filepath.Join(dir, "InstBase.tla"), `---- MODULE InstBase ----
InstBaseOp == TRUE
====`)
		writeFile(t, filepath.Join(dir, "Instancer.tla"), `---- MODULE Instancer ----
EXTENDS InstBase
InstOp == InstBaseOp
====`)
		rootPath := filepath.Join(dir, "Root.tla")
		writeFile(t, rootPath, `---- MODULE Root ----
EXTENDS Extender
INSTANCE Instancer
Use == ExtOp /\ InstOp
====`)

		spec, diags := LoadSanySpec(rootPath, LoadOptions{})
		requireNoErrors(t, diags)
		xmlText, xmlDiags := SanyXML(spec)
		requireNoErrors(t, xmlDiags)

		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v\n%s", err, xmlText)
		}
		moduleNamesByUID := map[string]string{}
		for _, entry := range canonicalSanyXMLEntries(root) {
			payload := canonicalSanyXMLEntryPayload(entry)
			if payload != nil && payload.Name == "ModuleNode" {
				moduleNamesByUID[firstChildText(entry, "UID")] = firstChildText(payload, "uniquename")
			}
		}
		var refs []string
		for _, child := range root.Children {
			if child.Name == "ModuleNodeRef" {
				refs = append(refs, moduleNamesByUID[firstChildText(child, "UID")])
			}
		}
		got := strings.Join(refs, ",")
		want := "ExtBase,Extender,InstBase,Instancer,Root"
		if got != want {
			t.Fatalf("module refs = %s, want %s\n%s", got, want, xmlText)
		}
	})

	t.Run("uses operator level for user-defined call nodes", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("UserCallLevelXML.tla", `---- MODULE UserCallLevelXML ----
VARIABLE x
Id(a) == a
Use == Id(x)
====`)
		requireNoErrors(t, diags)

		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v\n%s", err, xmlText)
		}
		call := opApplAtLocation(root, 4, 8, 12)
		if call == nil {
			t.Fatalf("Id(x) call node missing\n%s", xmlText)
		}
		if got := firstChildText(call, "level"); got != strconv.Itoa(int(variableLevel)) {
			t.Fatalf("Id(x) call level = %s, want 1\n%s", got, xmlText)
		}
		operand := opApplAtLocation(root, 4, 11, 11)
		if operand == nil {
			t.Fatalf("Id(x) operand node missing\n%s", xmlText)
		}
		if got := firstChildText(operand, "level"); got != strconv.Itoa(int(variableLevel)) {
			t.Fatalf("Id(x) operand level = %s, want 1\n%s", got, xmlText)
		}
	})

	t.Run("does not lift calls through unused formal parameters", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("UnusedFormalCallLevelXML.tla", `---- MODULE UnusedFormalCallLevelXML ----
VARIABLE x
Ignore(a) == TRUE
Use == Ignore(x)
====`)
		requireNoErrors(t, diags)

		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v\n%s", err, xmlText)
		}
		call := opApplAtLocation(root, 4, 8, 16)
		if call == nil {
			t.Fatalf("Ignore(x) call node missing\n%s", xmlText)
		}
		if got := firstChildText(call, "level"); got != strconv.Itoa(int(constantLevel)) {
			t.Fatalf("Ignore(x) call level = %s, want 0\n%s", got, xmlText)
		}
		operand := opApplAtLocation(root, 4, 15, 15)
		if operand == nil {
			t.Fatalf("Ignore(x) operand node missing\n%s", xmlText)
		}
		if got := firstChildText(operand, "level"); got != strconv.Itoa(int(variableLevel)) {
			t.Fatalf("Ignore(x) operand level = %s, want 1\n%s", got, xmlText)
		}
	})

	t.Run("used formals lift current-module symbolic infix call nodes", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("SymbolicInfixCallLevelXML.tla", `---- MODULE SymbolicInfixCallLevelXML ----
VARIABLE x
a \odot b == a
Use == x \odot x
====`)
		requireNoErrors(t, diags)

		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v\n%s", err, xmlText)
		}
		call := opApplAtLocation(root, 4, 8, 16)
		if call == nil {
			t.Fatalf("x \\odot x call node missing\n%s", xmlText)
		}
		if got := firstChildText(call, "level"); got != strconv.Itoa(int(variableLevel)) {
			t.Fatalf("x \\odot x call level = %s, want 1\n%s", got, xmlText)
		}
		operand := opApplAtLocation(root, 4, 8, 8)
		if operand == nil {
			t.Fatalf("x \\odot x operand node missing\n%s", xmlText)
		}
		if got := firstChildText(operand, "level"); got != strconv.Itoa(int(variableLevel)) {
			t.Fatalf("x \\odot x operand level = %s, want 1\n%s", got, xmlText)
		}
	})

	t.Run("unused formals do not lift current-module symbolic infix call nodes", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("UnusedSymbolicInfixCallLevelXML.tla", `---- MODULE UnusedSymbolicInfixCallLevelXML ----
VARIABLE x
a \odot b == TRUE
Use == x \odot x
====`)
		requireNoErrors(t, diags)

		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v\n%s", err, xmlText)
		}
		call := opApplAtLocation(root, 4, 8, 16)
		if call == nil {
			t.Fatalf("x \\odot x call node missing\n%s", xmlText)
		}
		if got := firstChildText(call, "level"); got != strconv.Itoa(int(constantLevel)) {
			t.Fatalf("x \\odot x call level = %s, want 0\n%s", got, xmlText)
		}
		operand := opApplAtLocation(root, 4, 8, 8)
		if operand == nil {
			t.Fatalf("x \\odot x operand node missing\n%s", xmlText)
		}
		if got := firstChildText(operand, "level"); got != strconv.Itoa(int(variableLevel)) {
			t.Fatalf("x \\odot x operand level = %s, want 1\n%s", got, xmlText)
		}
	})

	t.Run("zero-weight helper definitions do not lift caller arguments", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("ZeroWeightHelperLevelXML.tla", `---- MODULE ZeroWeightHelperLevelXML ----
VARIABLE x
Times(a, b) == TRUE
a \odot b == Times(a, b)
Use == x \odot x
====`)
		requireNoErrors(t, diags)

		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v\n%s", err, xmlText)
		}
		call := opApplAtLocation(root, 5, 8, 16)
		if call == nil {
			t.Fatalf("x \\odot x call node missing\n%s", xmlText)
		}
		if got := firstChildText(call, "level"); got != strconv.Itoa(int(constantLevel)) {
			t.Fatalf("x \\odot x call level = %s, want 0\n%s", got, xmlText)
		}
	})

	t.Run("transitive helper weights lift caller arguments", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("TransitiveHelperLevelXML.tla", `---- MODULE TransitiveHelperLevelXML ----
VARIABLE x
Same(a, b) == a = b
a \prec b == Same(a, b)
Use == x \prec x
====`)
		requireNoErrors(t, diags)

		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v\n%s", err, xmlText)
		}
		call := opApplAtLocation(root, 5, 8, 16)
		if call == nil {
			t.Fatalf("x \\prec x call node missing\n%s", xmlText)
		}
		if got := firstChildText(call, "level"); got != strconv.Itoa(int(variableLevel)) {
			t.Fatalf("x \\prec x call level = %s, want 1\n%s", got, xmlText)
		}
	})

	t.Run("operator formal calls lift higher-order actual arguments", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("HigherOrderFormalLevelXML.tla", `---- MODULE HigherOrderFormalLevelXML ----
VARIABLE x
Choose(S, P(_)) == CHOOSE y \in S : P(y)
Use == Choose({1}, LAMBDA y : x)
====`)
		requireNoErrors(t, diags)

		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v\n%s", err, xmlText)
		}
		call := opApplAtLocation(root, 4, 8, 32)
		if call == nil {
			t.Fatalf("Choose call node missing\n%s", xmlText)
		}
		if got := firstChildText(call, "level"); got != strconv.Itoa(int(variableLevel)) {
			t.Fatalf("Choose call level = %s, want 1\n%s", got, xmlText)
		}
	})

	t.Run("LET local definitions keep their own expression levels", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("LetLocalDefinitionLevelXML.tla", `---- MODULE LetLocalDefinitionLevelXML ----
VARIABLE x
Use == LET r == x IN r
====`)
		requireNoErrors(t, diags)

		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v\n%s", err, xmlText)
		}
		def := userDefinedOpAtLocation(root, 3, 12, 17)
		if def == nil {
			t.Fatalf("LET local definition r missing\n%s", xmlText)
		}
		if got := firstChildText(def, "level"); got != strconv.Itoa(int(variableLevel)) {
			t.Fatalf("LET local r level = %s, want 1\n%s", got, xmlText)
		}
	})

	t.Run("terminal BY serializes expression facts", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("TerminalByExpressionFactXML.tla", `---- MODULE TerminalByExpressionFactXML ----
THEOREM TRUE
  BY \A c \in {1} : c = c
====`)
		requireNoErrors(t, diags)
		got := normalizeXMLForContains(string(xmlText))
		for _, want := range []string{
			`<by>`,
			`<facts>`,
			`<OpApplNode>`,
			`<uniquename>$BoundedForall</uniquename>`,
			`<uniquename>c</uniquename>`,
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("terminal BY expression fact missing %q\n%s", want, got)
			}
		}
	})

	t.Run("terminal BY parses prefix operator expression facts", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("TerminalByPrefixFactXML.tla", `---- MODULE TerminalByPrefixFactXML ----
THEOREM TRUE
  BY DOMAIN <<1>> = 1..1
====`)
		requireNoErrors(t, diags)
		got := normalizeXMLForContains(string(xmlText))
		for _, want := range []string{
			`<by>`,
			`<facts>`,
			`<uniquename>DOMAIN</uniquename>`,
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("terminal BY prefix expression fact missing %q\n%s", want, got)
			}
		}
	})

	t.Run("generated bang subexpression calls emit lambdas", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("BangSubexpressionLambdaXML.tla", `---- MODULE BangSubexpressionLambdaXML ----
CONSTANT S
VARIABLE x
a == /\ x = x
     /\ IF TRUE
           THEN /\ \E I \in S : x' = x
                /\ x = x
           ELSE TRUE
     /\ x = x
THEOREM TRUE
<1>1. PICK I \in S : a!2!2!1!(I)
  BY DEF a
<1> QED OBVIOUS
====`)
		requireNoErrors(t, diags)
		got := normalizeXMLForContains(string(xmlText))
		if strings.Contains(got, `<uniquename>a!2!2!1</uniquename>`) {
			t.Fatalf("bang subexpression call emitted fake builtin\n%s", got)
		}
		if !strings.Contains(got, `<uniquename>LAMBDA</uniquename>`) {
			t.Fatalf("bang subexpression call did not emit lambda\n%s", got)
		}
	})

	t.Run("function application function operand carries function level", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("FunctionAppFunctionLevelXML.tla", `---- MODULE FunctionAppFunctionLevelXML ----
CONSTANT S
VARIABLE v
F[x \in S] == TRUE
A == LET arg == v IN F[arg]
====`)
		requireNoErrors(t, diags)
		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v\n%s", err, xmlText)
		}
		fUID := xmlEntryUIDByKindAndName(root, "UserDefinedOpKind", "F")
		if fUID == "" {
			t.Fatalf("SANY XML missing F definition\n%s", xmlText)
		}
		app := firstOpApplNodeForOperatorUID(root, "UserDefinedOpKindRef", fUID)
		if app == nil {
			t.Fatalf("SANY XML missing F application\n%s", xmlText)
		}
		if got := firstChildText(app, "level"); got != "0" {
			t.Fatalf("F application function operand level = %s, want 0\n%s", got, xmlText)
		}
	})

	t.Run("function definition level includes LET local definitions", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("FunctionDefLetLevelXML.tla", `---- MODULE FunctionDefLetLevelXML ----
CONSTANT S
VARIABLE v
F[x \in S] == LET local == v IN local = x
====`)
		requireNoErrors(t, diags)
		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v\n%s", err, xmlText)
		}
		def := xmlEntryPayloadByKindAndName(root, "UserDefinedOpKind", "F")
		if def == nil {
			t.Fatalf("SANY XML missing F definition\n%s", xmlText)
		}
		if got := firstChildText(def, "level"); got != "1" {
			t.Fatalf("F definition level = %s, want 1\n%s", got, xmlText)
		}
	})

	t.Run("LET expressions keep SANY non-Leibniz coloring local", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("LetLeibnizXML.tla", `---- MODULE LetLeibnizXML ----
VARIABLE v
G(y) == v' = y
Plain(x) == ENABLED G(x)
ViaLet(x) == LET D == TRUE IN ENABLED G(x)
====`)
		requireNoErrors(t, diags)
		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v\n%s", err, xmlText)
		}
		plain := xmlEntryPayloadByKindAndName(root, "UserDefinedOpKind", "Plain")
		if plain == nil {
			t.Fatalf("SANY XML missing Plain definition\n%s", xmlText)
		}
		plainParams := directChildren(directChildren(plain, "params")[0], "leibnizparam")
		if len(plainParams) != 1 || len(directChildren(plainParams[0], "leibniz")) != 0 {
			t.Fatalf("Plain param has Leibniz marker, want Java SANY non-Leibniz\n%s", xmlText)
		}
		viaLet := xmlEntryPayloadByKindAndName(root, "UserDefinedOpKind", "ViaLet")
		if viaLet == nil {
			t.Fatalf("SANY XML missing ViaLet definition\n%s", xmlText)
		}
		viaLetParams := directChildren(directChildren(viaLet, "params")[0], "leibnizparam")
		if len(viaLetParams) != 1 || len(directChildren(viaLetParams[0], "leibniz")) != 1 {
			t.Fatalf("ViaLet param missing Leibniz marker, want Java SANY LetInNode behavior\n%s", xmlText)
		}
	})

	t.Run("nested bullet lists remain nested SANY junction nodes", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("NestedJunctionXML.tla", `---- MODULE NestedJunctionXML ----
VARIABLE x
A == /\ /\ x = x
        /\ x' = x
     /\ x = x
====`)
		requireNoErrors(t, diags)
		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v\n%s", err, xmlText)
		}
		conjUID := xmlEntryUIDByKindAndName(root, "BuiltInKind", "$ConjList")
		if conjUID == "" {
			t.Fatalf("SANY XML missing $ConjList builtin\n%s", xmlText)
		}
		if !opApplContainsDirectOperandWithOperator(root, conjUID, conjUID) {
			t.Fatalf("nested bullet list was flattened instead of preserving a nested $ConjList\n%s", xmlText)
		}
	})

	t.Run("same-column bullet lists flatten into one SANY junction node", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("SameColumnJunctionXML.tla", `---- MODULE SameColumnJunctionXML ----
VARIABLE x
A == /\ x = x
     /\ x' = x
     /\ x = x
====`)
		requireNoErrors(t, diags)
		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v\n%s", err, xmlText)
		}
		conjUID := xmlEntryUIDByKindAndName(root, "BuiltInKind", "$ConjList")
		if conjUID == "" {
			t.Fatalf("SANY XML missing $ConjList builtin\n%s", xmlText)
		}
		if got := maxDirectOperandsForOperator(root, conjUID); got != 3 {
			t.Fatalf("same-column junction list has %d direct operands, want 3\n%s", got, xmlText)
		}
		if opApplContainsDirectOperandWithOperator(root, conjUID, conjUID) {
			t.Fatalf("same-column junction list kept a nested $ConjList wrapper\n%s", xmlText)
		}
	})

	t.Run("explicit PROOF steps location starts at PROOF token", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("ExplicitProofStepsLocationXML.tla", `---- MODULE ExplicitProofStepsLocationXML ----
THEOREM TRUE
PROOF
  <1>1. TRUE BY
  <1> QED BY <1>1
====`)
		requireNoErrors(t, diags)

		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v\n%s", err, xmlText)
		}
		steps := firstXMLDescendant(root, "steps")
		if steps == nil {
			t.Fatalf("steps node missing\n%s", xmlText)
		}
		if !xmlNodeLocationMatches(steps, 3, 1, 17) {
			t.Fatalf("steps location = %s, want line 3 column 1 through line 5 column 17\n%s", firstChildText(steps, "location"), xmlText)
		}
	})

	t.Run("LET recursive operator is in scope for its own body", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("LetRecursiveScopeXML.tla", `---- MODULE LetRecursiveScopeXML ----
Op(a) ==
  LET
    RECURSIVE R(_)
    R(i) == IF i = 0 THEN a ELSE R(i - 1)
  IN R(1)
====`)
		requireNoErrors(t, diags)
		got := normalizeXMLForContains(string(xmlText))
		for _, want := range []string{
			`<uniquename>R</uniquename>`,
			`<recursive/>`,
			`<recursiveSection>1</recursiveSection>`,
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("LET recursive XML missing %q\n%s", want, got)
			}
		}
	})

	t.Run("bound names shadow outer declarations for levels", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "BoundHelper.tla"), `---- MODULE BoundHelper ----
F(S) == {n : n \in S}
====`)
		rootPath := filepath.Join(dir, "BoundNameLevelShadowXML.tla")
		writeFile(t, rootPath, `---- MODULE BoundNameLevelShadowXML ----
EXTENDS BoundHelper
CONSTANT C
VARIABLE n
Use == F(C)
====`)
		spec, diags := LoadSanySpec(rootPath, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
		xmlText, xmlDiags := SanyXML(spec)
		requireNoErrors(t, xmlDiags)

		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v\n%s", err, xmlText)
		}
		def := userDefinedOpAtLocation(root, 5, 1, 11)
		if def == nil {
			t.Fatalf("Use definition missing\n%s", xmlText)
		}
		if got := firstChildText(def, "level"); got != strconv.Itoa(int(constantLevel)) {
			t.Fatalf("Use level = %s, want 0\n%s", got, xmlText)
		}
	})

	t.Run("symbolic infix formal parameter location covers the full declaration", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("InfixFormalSpanXML.tla", `---- MODULE InfixFormalSpanXML ----
Use(_\prec_, S) == TRUE
====`)
		requireNoErrors(t, diags)

		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v\n%s", err, xmlText)
		}
		var formal *canonicalXMLNode
		for _, entry := range canonicalSanyXMLEntries(root) {
			payload := canonicalSanyXMLEntryPayload(entry)
			if payload != nil && payload.Name == "FormalParamNode" && firstChildText(payload, "uniquename") == `\prec` {
				formal = payload
				break
			}
		}
		if formal == nil {
			t.Fatalf("infix formal parameter missing\n%s", xmlText)
		}
		locations := directChildren(formal, "location")
		if len(locations) != 1 {
			t.Fatalf("FormalParamNode locations = %d, want 1\n%s", len(locations), xmlText)
		}
		columns := directChildren(locations[0], "column")
		if len(columns) != 1 {
			t.Fatalf("FormalParamNode column locations = %d, want 1\n%s", len(columns), xmlText)
		}
		if begin, end := firstChildText(columns[0], "begin"), firstChildText(columns[0], "end"); begin != "5" || end != "11" {
			t.Fatalf("infix formal columns = %s..%s, want 5..11\n%s", begin, end, xmlText)
		}
	})

	t.Run("higher-order NEW symbol location covers the full arity declaration", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("NewOpDeclSpanXML.tla", `---- MODULE NewOpDeclSpanXML ----
THEOREM T ==
  ASSUME NEW Def(_,_)
  PROVE TRUE
PROOF OBVIOUS
====`)
		requireNoErrors(t, diags)

		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v\n%s", err, xmlText)
		}
		var decl *canonicalXMLNode
		for _, entry := range canonicalSanyXMLEntries(root) {
			payload := canonicalSanyXMLEntryPayload(entry)
			if payload != nil && payload.Name == "OpDeclNode" && firstChildText(payload, "uniquename") == "Def" {
				decl = payload
				break
			}
		}
		if decl == nil {
			t.Fatalf("NEW operator declaration missing\n%s", xmlText)
		}
		locations := directChildren(decl, "location")
		if len(locations) != 1 {
			t.Fatalf("OpDeclNode locations = %d, want 1\n%s", len(locations), xmlText)
		}
		columns := directChildren(locations[0], "column")
		if len(columns) != 1 {
			t.Fatalf("OpDeclNode column locations = %d, want 1\n%s", len(columns), xmlText)
		}
		if begin, end := firstChildText(columns[0], "begin"), firstChildText(columns[0], "end"); begin != "14" || end != "21" {
			t.Fatalf("NEW operator declaration columns = %s..%s, want 14..21\n%s", begin, end, xmlText)
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

	t.Run("clones extended definitions through LOCAL INSTANCE", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "LocalBase.tla"), `---- MODULE LocalBase ----
BaseOp == TRUE
====`)
		writeFile(t, filepath.Join(dir, "LocalTarget.tla"), `---- MODULE LocalTarget ----
EXTENDS LocalBase
TargetOp == BaseOp
====`)
		root := filepath.Join(dir, "LocalInstanceExtendsXML.tla")
		writeFile(t, root, `---- MODULE LocalInstanceExtendsXML ----
LOCAL INSTANCE LocalTarget
Use == BaseOp /\ TargetOp
====`)

		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
		xmlText, xmlDiags := SanyXML(spec)
		requireNoErrors(t, xmlDiags)

		rootXML, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v", err)
		}
		countRootUserDef := func(name string) int {
			count := 0
			for _, entry := range canonicalSanyXMLEntries(rootXML) {
				child := canonicalSanyXMLEntryPayload(entry)
				if child == nil || child.Name != "UserDefinedOpKind" {
					continue
				}
				if firstChildText(child, "uniquename") == name && firstDescendantText(child, "filename") == "LocalInstanceExtendsXML" {
					count++
				}
			}
			return count
		}
		if got := countRootUserDef("BaseOp"); got != 1 {
			t.Fatalf("LOCAL INSTANCE extended BaseOp clones = %d, want 1\n%s", got, string(xmlText))
		}
		if got := countRootUserDef("TargetOp"); got != 1 {
			t.Fatalf("LOCAL INSTANCE target TargetOp clones = %d, want 1\n%s", got, string(xmlText))
		}
	})

	t.Run("deduplicates overlapping unqualified LOCAL INSTANCE clones", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "SharedBase.tla"), `---- MODULE SharedBase ----
SharedOp == TRUE
====`)
		writeFile(t, filepath.Join(dir, "LocalA.tla"), `---- MODULE LocalA ----
EXTENDS SharedBase
AOp == SharedOp
====`)
		writeFile(t, filepath.Join(dir, "LocalB.tla"), `---- MODULE LocalB ----
EXTENDS SharedBase
BOp == SharedOp
====`)
		root := filepath.Join(dir, "LocalInstanceDedupXML.tla")
		writeFile(t, root, `---- MODULE LocalInstanceDedupXML ----
LOCAL INSTANCE LocalA
LOCAL INSTANCE LocalB
Use == SharedOp /\ AOp /\ BOp
====`)

		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
		xmlText, xmlDiags := SanyXML(spec)
		requireNoErrors(t, xmlDiags)

		rootXML, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v", err)
		}
		count := 0
		for _, entry := range canonicalSanyXMLEntries(rootXML) {
			child := canonicalSanyXMLEntryPayload(entry)
			if child == nil || child.Name != "UserDefinedOpKind" {
				continue
			}
			if firstChildText(child, "uniquename") == "SharedOp" && firstDescendantText(child, "filename") == "LocalInstanceDedupXML" {
				count++
			}
		}
		if count != 1 {
			t.Fatalf("LOCAL INSTANCE SharedOp clones = %d, want 1\n%s", count, string(xmlText))
		}
	})

	t.Run("emits implicit substitutions alongside explicit WITH substitutions", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Helper.tla"), `---- MODULE Helper ----
CONSTANTS A, B
VARIABLE x
====`)
		root := filepath.Join(dir, "ExplicitImplicitInstanceXML.tla")
		writeFile(t, root, `---- MODULE ExplicitImplicitInstanceXML ----
EXTENDS Naturals
CONSTANT A
VARIABLE x
Inst == INSTANCE Helper WITH B <- Nat
====`)

		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
		xmlText, xmlDiags := SanyXML(spec)
		requireNoErrors(t, xmlDiags)

		got := string(xmlText)
		if count := strings.Count(got, "<Subst>"); count != 3 {
			t.Fatalf("INSTANCE substitutions = %d, want explicit B plus implicit A and x\n%s", count, got)
		}
		for _, want := range []string{
			`<uniquename>Inst</uniquename>`,
			`<module>Helper</module>`,
			`<uniquename>A</uniquename>`,
			`<uniquename>B</uniquename>`,
			`<uniquename>x</uniquename>`,
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("SANY XML did not include substitution detail %q\n%s", want, got)
			}
		}
	})

	t.Run("clones extended definitions through unqualified INSTANCE", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Base.tla"), `---- MODULE Base ----
BaseOp == TRUE
====`)
		writeFile(t, filepath.Join(dir, "Target.tla"), `---- MODULE Target ----
EXTENDS Base
TargetOp == BaseOp
====`)
		rootPath := filepath.Join(dir, "UnqualifiedInstanceExtendsXML.tla")
		writeFile(t, rootPath, `---- MODULE UnqualifiedInstanceExtendsXML ----
INSTANCE Target
Use == TargetOp /\ BaseOp
====`)

		spec, diags := LoadSanySpec(rootPath, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
		xmlText, xmlDiags := SanyXML(spec)
		requireNoErrors(t, xmlDiags)

		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v", err)
		}
		countUserDef := func(filename, name string) int {
			count := 0
			for _, entry := range canonicalSanyXMLEntries(root) {
				child := canonicalSanyXMLEntryPayload(entry)
				if child == nil || child.Name != "UserDefinedOpKind" {
					continue
				}
				if firstChildText(child, "uniquename") == name && firstDescendantText(child, "filename") == filename {
					count++
				}
			}
			return count
		}
		if got := countUserDef("UnqualifiedInstanceExtendsXML", "TargetOp"); got != 1 {
			t.Fatalf("root TargetOp clones = %d, want 1\n%s", got, string(xmlText))
		}
		if got := countUserDef("UnqualifiedInstanceExtendsXML", "BaseOp"); got != 1 {
			t.Fatalf("root BaseOp clones = %d, want 1\n%s", got, string(xmlText))
		}
		if got := countUserDef("Base", "BaseOp"); got != 1 {
			t.Fatalf("original BaseOp entries = %d, want 1\n%s", got, string(xmlText))
		}

		uidInfo := map[string]string{}
		for _, entry := range canonicalSanyXMLEntries(root) {
			child := canonicalSanyXMLEntryPayload(entry)
			if child == nil || child.Name != "UserDefinedOpKind" {
				continue
			}
			uidInfo[firstChildText(entry, "UID")] = firstDescendantText(child, "filename") + ":" + firstChildText(child, "uniquename")
		}
		rootRefs := map[string]bool{}
		for _, entry := range canonicalSanyXMLEntries(root) {
			child := canonicalSanyXMLEntryPayload(entry)
			if child == nil || child.Name != "ModuleNode" || firstChildText(child, "uniquename") != "UnqualifiedInstanceExtendsXML" {
				continue
			}
			for _, ref := range child.Children {
				if ref.Name == "UserDefinedOpKindRef" {
					rootRefs[uidInfo[firstChildText(ref, "UID")]] = true
				}
			}
		}
		if !rootRefs["UnqualifiedInstanceExtendsXML:BaseOp"] {
			t.Fatalf("root module refs did not include cloned BaseOp\n%s", string(xmlText))
		}
	})

	t.Run("unqualified INSTANCE does not clone library EXTENDS definitions", func(t *testing.T) {
		dir := t.TempDir()
		libDir := filepath.Join(dir, "lib")
		if err := os.MkdirAll(libDir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", libDir, err)
		}
		writeFile(t, filepath.Join(libDir, "Lib.tla"), `---- MODULE Lib ----
LibOp == TRUE
====`)
		writeFile(t, filepath.Join(dir, "Target.tla"), `---- MODULE Target ----
EXTENDS Lib
TargetOp == LibOp
====`)
		rootPath := filepath.Join(dir, "UnqualifiedInstanceLibraryExtendsXML.tla")
		writeFile(t, rootPath, `---- MODULE UnqualifiedInstanceLibraryExtendsXML ----
INSTANCE Target
Use == TargetOp /\ LibOp
====`)

		spec, diags := LoadSanySpec(rootPath, LoadOptions{LibraryPaths: []string{libDir}, PreferLibraryModules: true})
		requireNoErrors(t, diags)
		requireNoErrors(t, CheckSpec(spec))
		xmlText, xmlDiags := SanyXML(spec)
		requireNoErrors(t, xmlDiags)

		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v", err)
		}
		countUserDef := func(filename, name string) int {
			count := 0
			for _, entry := range canonicalSanyXMLEntries(root) {
				child := canonicalSanyXMLEntryPayload(entry)
				if child == nil || child.Name != "UserDefinedOpKind" {
					continue
				}
				if firstChildText(child, "uniquename") == name && firstDescendantText(child, "filename") == filename {
					count++
				}
			}
			return count
		}
		if got := countUserDef("UnqualifiedInstanceLibraryExtendsXML", "TargetOp"); got != 1 {
			t.Fatalf("root TargetOp clones = %d, want 1\n%s", got, string(xmlText))
		}
		if got := countUserDef("UnqualifiedInstanceLibraryExtendsXML", "LibOp"); got != 0 {
			t.Fatalf("root LibOp clones = %d, want 0\n%s", got, string(xmlText))
		}
		if got := countUserDef("Lib", "LibOp"); got != 1 {
			t.Fatalf("original LibOp entries = %d, want 1\n%s", got, string(xmlText))
		}
	})

	t.Run("LOCAL unqualified INSTANCE clones library EXTENDS definitions", func(t *testing.T) {
		dir := t.TempDir()
		libDir := filepath.Join(dir, "lib")
		if err := os.MkdirAll(libDir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", libDir, err)
		}
		writeFile(t, filepath.Join(libDir, "Lib.tla"), `---- MODULE Lib ----
LibOp == TRUE
====`)
		writeFile(t, filepath.Join(dir, "Target.tla"), `---- MODULE Target ----
EXTENDS Lib
TargetOp == LibOp
====`)
		rootPath := filepath.Join(dir, "LocalInstanceLibraryExtendsXML.tla")
		writeFile(t, rootPath, `---- MODULE LocalInstanceLibraryExtendsXML ----
LOCAL INSTANCE Target
Use == TargetOp /\ LibOp
====`)

		spec, diags := LoadSanySpec(rootPath, LoadOptions{LibraryPaths: []string{libDir}, PreferLibraryModules: true})
		requireNoErrors(t, diags)
		requireNoErrors(t, CheckSpec(spec))
		xmlText, xmlDiags := SanyXML(spec)
		requireNoErrors(t, xmlDiags)

		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v", err)
		}
		var targetClones, libClones int
		for _, entry := range canonicalSanyXMLEntries(root) {
			child := canonicalSanyXMLEntryPayload(entry)
			if child == nil || child.Name != "UserDefinedOpKind" || firstDescendantText(child, "filename") != "LocalInstanceLibraryExtendsXML" {
				continue
			}
			switch firstChildText(child, "uniquename") {
			case "TargetOp":
				targetClones++
			case "LibOp":
				libClones++
			}
		}
		if targetClones != 1 || libClones != 1 {
			t.Fatalf("LOCAL INSTANCE clones TargetOp=%d LibOp=%d, want 1 each\n%s", targetClones, libClones, string(xmlText))
		}
	})

	t.Run("LOCAL INSTANCE omits recursive inherited originals from module refs", func(t *testing.T) {
		dir := t.TempDir()
		libDir := filepath.Join(dir, "lib")
		if err := os.MkdirAll(libDir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", libDir, err)
		}
		writeFile(t, filepath.Join(libDir, "Grand.tla"), `---- MODULE Grand ----
GrandOp == TRUE
====`)
		writeFile(t, filepath.Join(libDir, "Mid.tla"), `---- MODULE Mid ----
EXTENDS Grand
MidOp == GrandOp
====`)
		writeFile(t, filepath.Join(dir, "Target.tla"), `---- MODULE Target ----
EXTENDS Mid
TargetOp == MidOp
====`)
		rootPath := filepath.Join(dir, "LocalInstanceRecursiveLibraryExtendsXML.tla")
		writeFile(t, rootPath, `---- MODULE LocalInstanceRecursiveLibraryExtendsXML ----
LOCAL INSTANCE Target
Use == TargetOp /\ MidOp /\ GrandOp
====`)

		spec, diags := LoadSanySpec(rootPath, LoadOptions{LibraryPaths: []string{libDir}, PreferLibraryModules: true})
		requireNoErrors(t, diags)
		requireNoErrors(t, CheckSpec(spec))
		xmlText, xmlDiags := SanyXML(spec)
		requireNoErrors(t, xmlDiags)
		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v", err)
		}
		if got := moduleRefCountByPayloadLocation(root, "LocalInstanceRecursiveLibraryExtendsXML", "UserDefinedOpKind", "Grand", 2); got != 0 {
			t.Fatalf("root module has %d refs to recursive inherited GrandOp, want 0\n%s", got, xmlText)
		}
	})

	t.Run("diamond EXTENDS preserves duplicate inherited assumption refs", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "A.tla"), `---- MODULE A ----
ASSUME AAssump == TRUE
====`)
		writeFile(t, filepath.Join(dir, "B.tla"), `---- MODULE B ----
EXTENDS A
BOp == TRUE
====`)
		writeFile(t, filepath.Join(dir, "C.tla"), `---- MODULE C ----
EXTENDS A
COp == TRUE
====`)
		rootPath := filepath.Join(dir, "DiamondAssumeXML.tla")
		writeFile(t, rootPath, `---- MODULE DiamondAssumeXML ----
EXTENDS B, C
Root == BOp /\ COp
====`)

		spec, diags := LoadSanySpec(rootPath, LoadOptions{})
		requireNoErrors(t, diags)
		requireNoErrors(t, CheckSpec(spec))
		xmlText, xmlDiags := SanyXML(spec)
		requireNoErrors(t, xmlDiags)
		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v", err)
		}
		if got := moduleRefCountByPayloadLocation(root, "DiamondAssumeXML", "AssumeNode", "A", 2); got != 2 {
			t.Fatalf("root module has %d refs to inherited AAssump, want 2\n%s", got, xmlText)
		}
	})

	t.Run("does not unqualified-clone recursive EXTENDS definitions through INSTANCE", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "GrandBase.tla"), `---- MODULE GrandBase ----
GrandOp == TRUE
====`)
		writeFile(t, filepath.Join(dir, "MidBase.tla"), `---- MODULE MidBase ----
EXTENDS GrandBase
MidOp == GrandOp
====`)
		writeFile(t, filepath.Join(dir, "LeafTarget.tla"), `---- MODULE LeafTarget ----
EXTENDS MidBase
LeafOp == MidOp
====`)
		rootPath := filepath.Join(dir, "ShallowInstanceExtendsXML.tla")
		writeFile(t, rootPath, `---- MODULE ShallowInstanceExtendsXML ----
INSTANCE LeafTarget
Use == LeafOp /\ MidOp /\ GrandOp
====`)

		spec, diags := LoadSanySpec(rootPath, LoadOptions{})
		requireNoErrors(t, diags)
		requireNoErrors(t, CheckSpec(spec))
		xmlText, xmlDiags := SanyXML(spec)
		requireNoErrors(t, xmlDiags)

		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v", err)
		}
		countRootUserDef := func(name string) int {
			count := 0
			for _, entry := range canonicalSanyXMLEntries(root) {
				child := canonicalSanyXMLEntryPayload(entry)
				if child == nil || child.Name != "UserDefinedOpKind" {
					continue
				}
				if firstChildText(child, "uniquename") == name && firstDescendantText(child, "filename") == "ShallowInstanceExtendsXML" {
					count++
				}
			}
			return count
		}
		if got := countRootUserDef("LeafOp"); got != 1 {
			t.Fatalf("root LeafOp clones = %d, want 1\n%s", got, string(xmlText))
		}
		if got := countRootUserDef("MidOp"); got != 1 {
			t.Fatalf("root MidOp clones = %d, want 1\n%s", got, string(xmlText))
		}
		if got := countRootUserDef("GrandOp"); got != 0 {
			t.Fatalf("root GrandOp clones = %d, want 0\n%s", got, string(xmlText))
		}
		uidInfo := map[string]string{}
		for _, entry := range canonicalSanyXMLEntries(root) {
			child := canonicalSanyXMLEntryPayload(entry)
			if child == nil || child.Name != "UserDefinedOpKind" {
				continue
			}
			uidInfo[firstChildText(entry, "UID")] = firstDescendantText(child, "filename") + ":" + firstChildText(child, "uniquename")
		}
		rootRefs := map[string]bool{}
		for _, entry := range canonicalSanyXMLEntries(root) {
			child := canonicalSanyXMLEntryPayload(entry)
			if child == nil || child.Name != "ModuleNode" || firstChildText(child, "uniquename") != "ShallowInstanceExtendsXML" {
				continue
			}
			for _, ref := range child.Children {
				if ref.Name == "UserDefinedOpKindRef" {
					rootRefs[uidInfo[firstChildText(ref, "UID")]] = true
				}
			}
		}
		if !rootRefs["GrandBase:GrandOp"] {
			t.Fatalf("root module refs did not include original GrandOp\n%s", string(xmlText))
		}
	})

	t.Run("does not reference unqualified INSTANCE definitions hidden by owner definitions", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "HiddenHelper.tla"), `---- MODULE HiddenHelper ----
Op == TRUE
====`)
		rootPath := filepath.Join(dir, "HiddenInstanceOverrideXML.tla")
		writeFile(t, rootPath, `---- MODULE HiddenInstanceOverrideXML ----
Op == FALSE
INSTANCE HiddenHelper
Use == Op
====`)

		spec, diags := LoadSanySpec(rootPath, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
		xmlText, xmlDiags := SanyXML(spec)
		requireNoErrors(t, xmlDiags)

		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v", err)
		}
		uidInfo := map[string]string{}
		for _, entry := range canonicalSanyXMLEntries(root) {
			child := canonicalSanyXMLEntryPayload(entry)
			if child == nil || child.Name != "UserDefinedOpKind" {
				continue
			}
			uidInfo[firstChildText(entry, "UID")] = firstDescendantText(child, "filename") + ":" + firstChildText(child, "uniquename")
		}
		rootRefs := map[string]bool{}
		for _, entry := range canonicalSanyXMLEntries(root) {
			child := canonicalSanyXMLEntryPayload(entry)
			if child == nil || child.Name != "ModuleNode" || firstChildText(child, "uniquename") != "HiddenInstanceOverrideXML" {
				continue
			}
			for _, ref := range child.Children {
				if ref.Name == "UserDefinedOpKindRef" {
					rootRefs[uidInfo[firstChildText(ref, "UID")]] = true
				}
			}
		}
		if rootRefs["HiddenHelper:Op"] {
			t.Fatalf("root module refs included hidden helper Op\n%s", string(xmlText))
		}
		if !rootRefs["HiddenInstanceOverrideXML:Op"] {
			t.Fatalf("root module refs omitted owner Op\n%s", string(xmlText))
		}
	})

	t.Run("serializes named INSTANCE nodes with their uniquename", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Helper.tla"), `---- MODULE Helper ----
CONSTANT C
Op == C
====`)
		root := filepath.Join(dir, "NamedInstanceXML.tla")
		writeFile(t, root, `---- MODULE NamedInstanceXML ----
CONSTANT C
Inst == INSTANCE Helper WITH C <- C
Use == Inst!Op
====`)

		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
		xmlText, xmlDiags := SanyXML(spec)
		requireNoErrors(t, xmlDiags)
		got := string(xmlText)
		start := strings.Index(got, `<InstanceNode>`)
		if start < 0 {
			t.Fatalf("SANY XML missing InstanceNode\n%s", got)
		}
		end := strings.Index(got[start:], `</InstanceNode>`)
		if end < 0 {
			t.Fatalf("SANY XML has unterminated InstanceNode\n%s", got[start:])
		}
		node := got[start : start+end]
		if !strings.Contains(node, `<uniquename>Inst</uniquename>`) {
			t.Fatalf("named INSTANCE node missing uniquename\n%s", node)
		}
	})

	t.Run("serializes named INSTANCE clones for extended module definitions", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Base.tla"), `---- MODULE Base ----
BaseOp == TRUE
====`)
		writeFile(t, filepath.Join(dir, "Helper.tla"), `---- MODULE Helper ----
EXTENDS Base
Op == BaseOp
====`)
		root := filepath.Join(dir, "NamedInstanceExtendsXML.tla")
		writeFile(t, root, `---- MODULE NamedInstanceExtendsXML ----
Inst == INSTANCE Helper
Use == Inst!Op /\ Inst!BaseOp
====`)

		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
		xmlText, xmlDiags := SanyXML(spec)
		requireNoErrors(t, xmlDiags)
		got := string(xmlText)
		for _, want := range []string{
			`<uniquename>Inst!Op</uniquename>`,
			`<uniquename>Inst!BaseOp</uniquename>`,
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("named INSTANCE clone XML missing %q\n%s", want, got)
			}
		}
	})

	t.Run("EXTENDS imports named INSTANCE clone refs into the module node", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Helper.tla"), `---- MODULE Helper ----
Op == TRUE
====`)
		writeFile(t, filepath.Join(dir, "Mid.tla"), `---- MODULE Mid ----
Inst == INSTANCE Helper
====`)
		rootPath := filepath.Join(dir, "ExtendsInstanceCloneXML.tla")
		writeFile(t, rootPath, `---- MODULE ExtendsInstanceCloneXML ----
EXTENDS Mid
====`)

		spec, diags := LoadSanySpec(rootPath, LoadOptions{})
		requireNoErrors(t, diags)
		requireNoErrors(t, CheckSpec(spec))
		xmlText, xmlDiags := SanyXML(spec)
		requireNoErrors(t, xmlDiags)
		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v\n%s", err, xmlText)
		}
		cloneUID := xmlEntryUIDByKindAndName(root, "UserDefinedOpKind", "Inst!Op")
		if cloneUID == "" {
			t.Fatalf("Inst!Op clone missing\n%s", xmlText)
		}
		var moduleNode *canonicalXMLNode
		for _, entry := range canonicalSanyXMLEntries(root) {
			payload := canonicalSanyXMLEntryPayload(entry)
			if payload != nil && payload.Name == "ModuleNode" && firstChildText(payload, "uniquename") == "ExtendsInstanceCloneXML" {
				moduleNode = payload
				break
			}
		}
		if moduleNode == nil {
			t.Fatalf("root ModuleNode missing\n%s", xmlText)
		}
		found := false
		for _, child := range moduleNode.Children {
			if child.Name == "UserDefinedOpKindRef" && firstChildText(child, "UID") == cloneUID {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("root ModuleNode did not import Inst!Op ref\n%s", xmlText)
		}
	})

	t.Run("EXTENDS does not import LOCAL INSTANCE clone refs into the module node", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Helper.tla"), `---- MODULE Helper ----
Op == TRUE
====`)
		writeFile(t, filepath.Join(dir, "Mid.tla"), `---- MODULE Mid ----
LOCAL Inst == INSTANCE Helper
====`)
		rootPath := filepath.Join(dir, "ExtendsLocalInstanceCloneXML.tla")
		writeFile(t, rootPath, `---- MODULE ExtendsLocalInstanceCloneXML ----
EXTENDS Mid
====`)

		spec, diags := LoadSanySpec(rootPath, LoadOptions{})
		requireNoErrors(t, diags)
		requireNoErrors(t, CheckSpec(spec))
		xmlText, xmlDiags := SanyXML(spec)
		requireNoErrors(t, xmlDiags)
		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v\n%s", err, xmlText)
		}
		cloneUID := xmlEntryUIDByKindAndName(root, "UserDefinedOpKind", "Inst!Op")
		if cloneUID == "" {
			t.Fatalf("local Inst!Op clone missing from owner module XML\n%s", xmlText)
		}
		var moduleNode *canonicalXMLNode
		for _, entry := range canonicalSanyXMLEntries(root) {
			payload := canonicalSanyXMLEntryPayload(entry)
			if payload != nil && payload.Name == "ModuleNode" && firstChildText(payload, "uniquename") == "ExtendsLocalInstanceCloneXML" {
				moduleNode = payload
				break
			}
		}
		if moduleNode == nil {
			t.Fatalf("root ModuleNode missing\n%s", xmlText)
		}
		for _, child := range moduleNode.Children {
			if child.Name == "UserDefinedOpKindRef" && firstChildText(child, "UID") == cloneUID {
				t.Fatalf("root ModuleNode imported local Inst!Op ref\n%s", xmlText)
			}
		}
	})

	t.Run("computes named INSTANCE clone levels before owner definitions", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Helper.tla"), `---- MODULE Helper ----
VARIABLE v
Init == v = 0
====`)
		rootPath := filepath.Join(dir, "NamedInstanceLevelXML.tla")
		writeFile(t, rootPath, `---- MODULE NamedInstanceLevelXML ----
VARIABLE x
Inst == INSTANCE Helper WITH v <- x
RootInit == Inst!Init
====`)

		spec, diags := LoadSanySpec(rootPath, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
		xmlText, xmlDiags := SanyXML(spec)
		requireNoErrors(t, xmlDiags)
		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v", err)
		}
		instInitUID := xmlEntryUIDByKindAndName(root, "UserDefinedOpKind", "Inst!Init")
		if instInitUID == "" {
			t.Fatalf("Inst!Init clone missing\n%s", string(xmlText))
		}
		node := firstOpApplNodeForOperatorUID(root, "UserDefinedOpKindRef", instInitUID)
		if node == nil {
			t.Fatalf("RootInit reference to Inst!Init missing\n%s", string(xmlText))
		}
		if got := firstChildText(node, "level"); got != strconv.Itoa(int(variableLevel)) {
			t.Fatalf("Inst!Init reference level = %s, want %d\n%s", got, variableLevel, string(xmlText))
		}
	})

	t.Run("anonymous INSTANCE does not reclone inherited owner definitions", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Base.tla"), `---- MODULE Base ----
BaseOp == TRUE
====`)
		writeFile(t, filepath.Join(dir, "Helper.tla"), `---- MODULE Helper ----
EXTENDS Base
HelperOp == BaseOp
====`)
		rootPath := filepath.Join(dir, "AnonymousInstanceInheritedXML.tla")
		writeFile(t, rootPath, `---- MODULE AnonymousInstanceInheritedXML ----
EXTENDS Base
INSTANCE Helper
Use == HelperOp /\ BaseOp
====`)

		spec, diags := LoadSanySpec(rootPath, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
		xmlText, xmlDiags := SanyXML(spec)
		requireNoErrors(t, xmlDiags)
		if count := strings.Count(string(xmlText), `<uniquename>BaseOp</uniquename>`); count != 1 {
			t.Fatalf("BaseOp entries = %d, want only the inherited definition\n%s", count, string(xmlText))
		}
	})

	t.Run("serializes LET INSTANCE qualified definitions", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Helper.tla"), `---- MODULE Helper ----
Op == TRUE
====`)
		rootPath := filepath.Join(dir, "LetInstanceKindXML.tla")
		writeFile(t, rootPath, `---- MODULE LetInstanceKindXML ----
Use == LET Inst == INSTANCE Helper IN Inst!Op
====`)

		spec, diags := LoadSanySpec(rootPath, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
		xmlText, xmlDiags := SanyXML(spec)
		requireNoErrors(t, xmlDiags)
		got := string(xmlText)
		for _, want := range []string{
			`<ModuleInstanceKind>`,
			`<uniquename>Inst</uniquename>`,
			`<uniquename>Inst!Op</uniquename>`,
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("LET INSTANCE XML missing %q\n%s", want, got)
			}
		}
		if strings.Contains(got, `<BuiltInKind>`) && strings.Contains(got, `<uniquename>Inst!Op</uniquename><arity>-1</arity>`) {
			t.Fatalf("LET INSTANCE qualified operator was emitted as builtin\n%s", got)
		}
	})

	t.Run("LET INSTANCE substitutions can target inherited declarations", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Base.tla"), `---- MODULE Base ----
CONSTANT C
Op == C
====`)
		rootPath := filepath.Join(dir, "LetInstanceInheritedSubstXML.tla")
		writeFile(t, rootPath, `---- MODULE LetInstanceInheritedSubstXML ----
EXTENDS Base
Use == LET Inst == INSTANCE Base IN Inst!Op
====`)

		spec, diags := LoadSanySpec(rootPath, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
		xmlText, xmlDiags := SanyXML(spec)
		requireNoErrors(t, xmlDiags)
		got := string(xmlText)
		if !strings.Contains(got, `<uniquename>Inst!Op</uniquename>`) {
			t.Fatalf("LET INSTANCE clone missing\n%s", got)
		}
		if !strings.Contains(got, `<SubstInNode>`) {
			t.Fatalf("LET INSTANCE clone missing inherited substitution wrapper\n%s", got)
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

	t.Run("preserves duplicate imported instance nodes from distinct EXTENDS paths", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Leaf.tla"), `---- MODULE Leaf ----
Zero == 0
====`)
		writeFile(t, filepath.Join(dir, "Shared.tla"), `---- MODULE Shared ----
LOCAL INSTANCE Leaf
====`)
		writeFile(t, filepath.Join(dir, "Left.tla"), `---- MODULE Left ----
EXTENDS Shared
====`)
		writeFile(t, filepath.Join(dir, "Right.tla"), `---- MODULE Right ----
EXTENDS Shared
====`)
		root := filepath.Join(dir, "DuplicateImportedInstanceXML.tla")
		writeFile(t, root, `---- MODULE DuplicateImportedInstanceXML ----
EXTENDS Left, Right
====`)

		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
		xmlText, xmlDiags := SanyXML(spec)
		requireNoErrors(t, xmlDiags)
		got := string(xmlText)
		start := strings.Index(got, `<uniquename>DuplicateImportedInstanceXML</uniquename>`)
		if start < 0 {
			t.Fatalf("root ModuleNode missing\n%s", got)
		}
		end := strings.Index(got[start:], `</ModuleNode>`)
		if end < 0 {
			t.Fatalf("root ModuleNode unterminated\n%s", got[start:])
		}
		rootNode := got[start : start+end]
		if count := strings.Count(rootNode, `<module>Leaf</module>`); count != 2 {
			t.Fatalf("imported Leaf instance nodes = %d, want duplicate paths preserved\n%s", count, rootNode)
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

	t.Run("parameterized INSTANCE prepends parameters to clones and preserves theorem facts", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Base.tla"), `---- MODULE Base ----
CONSTANT C
Op(x) == C = x
THEOREM Lemma == C = C
====`)
		root := filepath.Join(dir, "ParamInstanceXML.tla")
		writeFile(t, root, `---- MODULE ParamInstanceXML ----
CONSTANT CSet
P(C) == INSTANCE Base
UseOp == \A c \in CSet : P(c)!Op(c)
UseLemma == \A c \in CSet : P(c)!Lemma
====`)

		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		requireNoErrors(t, CheckSpec(spec))
		xmlText, xmlDiags := SanyXML(spec)
		requireNoErrors(t, xmlDiags)
		rootXML, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v\n%s", err, xmlText)
		}
		op := xmlEntryPayloadByKindAndName(rootXML, "UserDefinedOpKind", "P!Op")
		if op == nil {
			t.Fatalf("parameterized instance clone P!Op missing\n%s", xmlText)
		}
		if got := firstChildText(op, "arity"); got != "2" {
			t.Fatalf("P!Op arity = %s, want 2\n%s", got, xmlText)
		}
		params := directChildren(op, "params")
		if len(params) != 1 || len(directChildren(params[0], "leibnizparam")) != 2 {
			t.Fatalf("P!Op params do not include instance param plus original param\n%s", xmlText)
		}
		theorem := xmlEntryPayloadByKindAndName(rootXML, "TheoremDefNode", "P!Lemma")
		if theorem == nil {
			t.Fatalf("parameterized instance theorem clone P!Lemma missing\n%s", xmlText)
		}
		if len(xmlNodesByName(theorem, "APSubstInNode")) != 1 {
			t.Fatalf("P!Lemma clone missing APSubstInNode\n%s", xmlText)
		}
		if refs := moduleRefCountByPayloadName(rootXML, "ParamInstanceXML", "TheoremDefNode", "P!Lemma"); refs != 0 {
			t.Fatalf("ParamInstanceXML module refs to P!Lemma theorem clone = %d, want 0\n%s", refs, xmlText)
		}
	})

	t.Run("plain unqualified INSTANCE does not clone theorem facts", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Base.tla"), `---- MODULE Base ----
CONSTANT C
THEOREM Lemma == C = C
====`)
		root := filepath.Join(dir, "PlainInstanceTheoremXML.tla")
		writeFile(t, root, `---- MODULE PlainInstanceTheoremXML ----
CONSTANT C
INSTANCE Base
Use == TRUE
====`)

		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		requireNoErrors(t, CheckSpec(spec))
		xmlText, xmlDiags := SanyXML(spec)
		requireNoErrors(t, xmlDiags)
		rootXML, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v\n%s", err, xmlText)
		}
		if count := xmlPayloadCountByKindNameFile(rootXML, "TheoremDefNode", "Lemma", "PlainInstanceTheoremXML"); count != 0 {
			t.Fatalf("plain INSTANCE cloned theorem facts into root module: %d\n%s", count, xmlText)
		}
		if count := xmlPayloadCountByKindNameFile(rootXML, "TheoremDefNode", "Lemma", "Base"); count != 1 {
			t.Fatalf("plain INSTANCE original theorem facts = %d, want 1\n%s", count, xmlText)
		}
	})

	t.Run("INSTANCE clones reuse source formal params after late standard module allocation", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("StdInstanceFormalXML.tla", `---- MODULE StdInstanceFormalXML ----
CC == INSTANCE Naturals
Use == TRUE
====`)
		requireNoErrors(t, diags)
		assertNoDanglingXMLRefs(t, xmlText)
	})

	t.Run("levels explicit INSTANCE substitution replacements", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Base.tla"), `---- MODULE Base ----
CONSTANT P
Use == P
====`)
		root := filepath.Join(dir, "ExplicitSubstLevelXML.tla")
		writeFile(t, root, `---- MODULE ExplicitSubstLevelXML ----
VARIABLE x
Repl == x
Inst == INSTANCE Base WITH P <- Repl
====`)

		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		requireNoErrors(t, CheckSpec(spec))
		xmlText, xmlDiags := SanyXML(spec)
		requireNoErrors(t, xmlDiags)
		rootXML, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v\n%s", err, xmlText)
		}
		replUID := xmlEntryUIDByKindAndName(rootXML, "UserDefinedOpKind", "Repl")
		if replUID == "" {
			t.Fatalf("Repl definition missing\n%s", xmlText)
		}
		var replacements int
		for _, subst := range xmlNodesByName(rootXML, "Subst") {
			for _, replacement := range directChildren(subst, "OpApplNode") {
				if !opApplNodeUsesOperatorUID(replacement, "UserDefinedOpKindRef", replUID) {
					continue
				}
				replacements++
				if got := firstChildText(replacement, "level"); got != strconv.Itoa(int(variableLevel)) {
					t.Fatalf("Repl substitution level = %s, want 1\n%s", got, xmlText)
				}
			}
		}
		if replacements == 0 {
			t.Fatalf("Repl substitution replacement missing\n%s", xmlText)
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

	t.Run("marks tuple destructuring bounds", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("TupleBoundXML.tla", `---- MODULE TupleBoundXML ----
Pairs == { <<1, 2>> }
A == { x : <<x, y>> \in Pairs }
====`)
		requireNoErrors(t, diags)

		got := string(xmlText)
		if !strings.Contains(got, `<tuple/>`) {
			t.Fatalf("tuple-bound XML missing tuple marker\n%s", got)
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

	t.Run("preserves unnamed theorem ASSUME PROVE bodies as AssumeProveNode", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("UnnamedAssumeProveTheoremXML.tla", `---- MODULE UnnamedAssumeProveTheoremXML ----
THEOREM ASSUME TRUE
        PROVE TRUE
====`)
		requireNoErrors(t, diags)

		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v\n%s", err, xmlText)
		}
		var found bool
		for _, theorem := range xmlNodesByName(root, "TheoremNode") {
			for _, body := range directChildren(theorem, "body") {
				if len(directChildren(body, "AssumeProveNode")) == 1 {
					found = true
				}
			}
		}
		if !found {
			t.Fatalf("unnamed theorem body was not exported as AssumeProveNode\n%s", xmlText)
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

	t.Run("serializes LAMBDA pre-comments", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("LambdaPreCommentXML.tla", `---- MODULE LambdaPreCommentXML ----
Apply(Op(_), x) == Op(x)
A == Apply(
  \* lambda explains the accumulator
  LAMBDA x : x,
  TRUE)
====`)
		requireNoErrors(t, diags)

		got := string(xmlText)
		if !strings.Contains(got, `<uniquename>LAMBDA</uniquename>`) {
			t.Fatalf("lambda XML missing\n%s", got)
		}
		if !strings.Contains(got, `<pre-comments><![CDATA[\* lambda explains the accumulator]]></pre-comments>`) {
			t.Fatalf("lambda XML missing pre-comment\n%s", got)
		}
	})

	t.Run("serializes operator formals passed as operator arguments", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("OperatorFormalArgXML.tla", `---- MODULE OperatorFormalArgXML ----
Apply(Op(_,_), x, y) == Op(x, y)
Forward(Op(_,_), x, y) == Apply(Op, x, y)
====`)
		requireNoErrors(t, diags)

		got := string(xmlText)
		if !strings.Contains(got, `<OpArgNode>`) {
			t.Fatalf("operator formal argument XML missing OpArgNode\n%s", got)
		}
		if !strings.Contains(got, `<argument>`) || !strings.Contains(got, `<FormalParamNodeRef>`) {
			t.Fatalf("operator formal argument XML missing FormalParamNodeRef argument\n%s", got)
		}
	})

	t.Run("canonicalizes parenthesized infix definition operators", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("ParenthesizedInfixDefXML.tla", `---- MODULE ParenthesizedInfixDefXML ----
A (+) B == A = B
====`)
		requireNoErrors(t, diags)

		got := string(xmlText)
		if !strings.Contains(got, `<uniquename>\oplus</uniquename>`) {
			t.Fatalf("parenthesized infix definition missing canonical \\oplus\n%s", got)
		}
		if strings.Contains(got, `<uniquename>(+)</uniquename>`) {
			t.Fatalf("parenthesized infix definition kept non-canonical (+)\n%s", got)
		}
	})

	t.Run("serializes ENABLED applications with variable level", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("EnabledLevelXML.tla", `---- MODULE EnabledLevelXML ----
VARIABLE x
Done == x' = x
Live == <>(ENABLED Done)
====`)
		requireNoErrors(t, diags)

		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v", err)
		}
		enabledUID := xmlEntryUIDByKindAndName(root, "BuiltInKind", "ENABLED")
		if enabledUID == "" {
			t.Fatalf("ENABLED builtin missing\n%s", string(xmlText))
		}
		node := firstOpApplNodeForOperatorUID(root, "BuiltInKindRef", enabledUID)
		if node == nil {
			t.Fatalf("ENABLED application missing\n%s", string(xmlText))
		}
		if got := firstChildText(node, "level"); got != strconv.Itoa(int(variableLevel)) {
			t.Fatalf("ENABLED level = %s, want %d\n%s", got, variableLevel, string(xmlText))
		}
	})

	t.Run("serializes action composition as non-Leibniz action builtin", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("ActionCompositionXML.tla", `---- MODULE ActionCompositionXML ----
VARIABLE x
A == (x' = x) \cdot (x' = x)
====`)
		requireNoErrors(t, diags)

		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v\n%s", err, xmlText)
		}
		var cdot *canonicalXMLNode
		for _, entry := range canonicalSanyXMLEntries(root) {
			payload := canonicalSanyXMLEntryPayload(entry)
			if payload != nil && payload.Name == "BuiltInKind" && firstChildText(payload, "uniquename") == `\cdot` {
				cdot = payload
				break
			}
		}
		if cdot == nil {
			t.Fatalf(`\cdot builtin missing`+"\n%s", xmlText)
		}
		if got := firstChildText(cdot, "level"); got != strconv.Itoa(int(actionLevel)) {
			t.Fatalf(`\cdot level = %s, want %d`+"\n%s", got, actionLevel, xmlText)
		}
		params := directChildren(cdot, "params")
		if len(params) != 1 {
			t.Fatalf(`\cdot params elements = %d, want 1`+"\n%s", len(params), xmlText)
		}
		for _, param := range directChildren(params[0], "leibnizparam") {
			if got := directChildren(param, "leibniz"); len(got) != 0 {
				t.Fatalf(`\cdot param has leibniz marker, want non-Leibniz`+"\n%s", xmlText)
			}
		}
	})

	t.Run("serializes inherited named assumptions as proof facts", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "BaseAssume.tla"), `---- MODULE BaseAssume ----
ASSUME NamedFact == TRUE
====`)
		rootPath := filepath.Join(dir, "ProofUsesAssume.tla")
		writeFile(t, rootPath, `---- MODULE ProofUsesAssume ----
EXTENDS BaseAssume
THEOREM T == TRUE
  BY NamedFact
====`)

		spec, diags := LoadSanySpec(rootPath, LoadOptions{})
		requireNoErrors(t, diags)
		requireNoErrors(t, CheckSpec(spec))
		xmlText, xmlDiags := SanyXML(spec)
		requireNoErrors(t, xmlDiags)

		if !strings.Contains(string(xmlText), `<AssumeDefRef>`) {
			t.Fatalf("proof fact did not reference inherited named assumption as AssumeDefRef\n%s", string(xmlText))
		}
	})

	t.Run("module context preserves duplicate imported theorem refs", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Base.tla"), `---- MODULE Base ----
THEOREM T == TRUE
====`)
		writeFile(t, filepath.Join(dir, "Mid.tla"), `---- MODULE Mid ----
EXTENDS Base
====`)
		rootPath := filepath.Join(dir, "Root.tla")
		writeFile(t, rootPath, `---- MODULE Root ----
EXTENDS Mid, Base
====`)

		spec, diags := LoadSanySpec(rootPath, LoadOptions{})
		requireNoErrors(t, diags)
		requireNoErrors(t, CheckSpec(spec))
		xmlText, xmlDiags := SanyXML(spec)
		requireNoErrors(t, xmlDiags)
		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v\n%s", err, xmlText)
		}
		if got := moduleRefCountByPayloadLocation(root, "Root", "TheoremNode", "Base", 2); got != 2 {
			t.Fatalf("Root module has %d refs to Base theorem line 2, want 2\n%s", got, xmlText)
		}
	})

	t.Run("implicit INSTANCE substitutions prefer owner definitions", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Base.tla"), `---- MODULE Base ----
CONSTANT C
Use == C
====`)
		rootPath := filepath.Join(dir, "Root.tla")
		writeFile(t, rootPath, `---- MODULE Root ----
C == TRUE
INSTANCE Base
====`)

		spec, diags := LoadSanySpec(rootPath, LoadOptions{})
		requireNoErrors(t, diags)
		requireNoErrors(t, CheckSpec(spec))
		xmlText, xmlDiags := SanyXML(spec)
		requireNoErrors(t, xmlDiags)
		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v\n%s", err, xmlText)
		}
		targetUID := xmlEntryUIDByKindAndName(root, "OpDeclNode", "C")
		replacementUID := xmlEntryUIDByKindAndName(root, "UserDefinedOpKind", "C")
		if targetUID == "" || replacementUID == "" {
			t.Fatalf("missing target or replacement symbol for implicit INSTANCE substitution\n%s", xmlText)
		}
		if !substUsesReplacementRef(root, targetUID, "UserDefinedOpKindRef", replacementUID) {
			t.Fatalf("implicit INSTANCE substitution did not use owner definition C\n%s", xmlText)
		}
	})

	t.Run("serializes proof DEFINE and PICK steps", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("ProofDefinePickXML.tla", `---- MODULE ProofDefinePickXML ----
THEOREM T == TRUE
<1>. DEFINE Local == TRUE
<1>1. Local
  BY DEF Local
<1>2. PICK x : x = x
  OBVIOUS
<1>3. x = x
  BY <1>2
<1>. QED
  BY <1>1
====`)
		requireNoErrors(t, diags)

		got := string(xmlText)
		for _, want := range []string{
			`<DefStepNode>`,
			`<uniquename>Local</uniquename>`,
			`<uniquename>$Pick</uniquename>`,
			`<uniquename>x</uniquename>`,
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("proof DEFINE/PICK XML missing %q\n%s", want, got)
			}
		}
	})

	t.Run("PICK proof step theorem level includes bounded domains", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("ProofPickBoundLevelXML.tla", `---- MODULE ProofPickBoundLevelXML ----
VARIABLE x
THEOREM T == TRUE
<1>1. PICK n \in 1..x : n = n
  OBVIOUS
<1> QED
  BY <1>1
====`)
		requireNoErrors(t, diags)
		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v\n%s", err, xmlText)
		}
		def := xmlEntryPayloadByKindAndName(root, "TheoremDefNode", "<1>1")
		if def == nil {
			t.Fatalf("PICK proof-step theorem definition missing\n%s", xmlText)
		}
		if got := firstChildText(def, "level"); got != "1" {
			t.Fatalf("PICK proof-step theorem definition level = %s, want 1\n%s", got, xmlText)
		}
	})

	t.Run("serializes multiple proof DEFINE definitions in one step", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("ProofMultiDefineXML.tla", `---- MODULE ProofMultiDefineXML ----
THEOREM T == TRUE
<1>. DEFINE First == TRUE
            Second == First
<1>1. Second
  BY DEF Second, First
<1>. QED
  BY <1>1
====`)
		requireNoErrors(t, diags)

		got := string(xmlText)
		start := strings.Index(got, "<DefStepNode>")
		if start < 0 {
			t.Fatalf("SANY XML missing DefStepNode\n%s", got)
		}
		end := strings.Index(got[start:], "</DefStepNode>")
		if end < 0 {
			t.Fatalf("SANY XML has unterminated DefStepNode\n%s", got)
		}
		defStep := got[start : start+end]
		if count := strings.Count(defStep, "<UserDefinedOpKindRef>"); count != 2 {
			t.Fatalf("DefStepNode refs = %d, want 2\n%s", count, defStep)
		}
		for _, want := range []string{
			`<uniquename>First</uniquename>`,
			`<uniquename>Second</uniquename>`,
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("proof DEFINE XML missing %q\n%s", want, got)
			}
		}
	})

	t.Run("omits Leibniz marker for proof DEFINE params under temporal operators", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("ProofTemporalLeibnizXML.tla", `---- MODULE ProofTemporalLeibnizXML ----
VARIABLE v
THEOREM T == TRUE
<1>. DEFINE P(m) == v = m
            L(m) == [](P(m) => <>TRUE)
<1>1. TRUE
  OBVIOUS
<1>. QED
  BY <1>1
====`)
		requireNoErrors(t, diags)

		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v\n%s", err, xmlText)
		}
		var def *canonicalXMLNode
		for _, entry := range canonicalSanyXMLEntries(root) {
			payload := canonicalSanyXMLEntryPayload(entry)
			if payload != nil && payload.Name == "UserDefinedOpKind" && firstChildText(payload, "uniquename") == "L" {
				def = payload
				break
			}
		}
		if def == nil {
			t.Fatalf("proof DEFINE L missing\n%s", xmlText)
		}
		params := directChildren(def, "params")
		if len(params) != 1 {
			t.Fatalf("L params elements = %d, want 1\n%s", len(params), xmlText)
		}
		leibnizParams := directChildren(params[0], "leibnizparam")
		if len(leibnizParams) != 1 {
			t.Fatalf("L leibnizparam elements = %d, want 1\n%s", len(leibnizParams), xmlText)
		}
		if got := directChildren(leibnizParams[0], "leibniz"); len(got) != 0 {
			t.Fatalf("L param has leibniz marker, want non-Leibniz\n%s", xmlText)
		}
	})

	t.Run("serializes ordinary SUFFICES proof steps", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("ProofSufficesXML.tla", `---- MODULE ProofSufficesXML ----
THEOREM T == TRUE
<1>1. SUFFICES TRUE
  OBVIOUS
<1>. QED
  BY <1>1
====`)
		requireNoErrors(t, diags)

		got := string(xmlText)
		if !strings.Contains(got, `<uniquename>$Suffices</uniquename>`) {
			t.Fatalf("ordinary SUFFICES XML missing $Suffices builtin\n%s", got)
		}
	})

	t.Run("PROOF OMITTED location starts at PROOF token", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("ProofOmittedLocationXML.tla", `---- MODULE ProofOmittedLocationXML ----
THEOREM T == TRUE
  PROOF OMITTED
====`)
		requireNoErrors(t, diags)
		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v\n%s", err, xmlText)
		}
		var omitted *canonicalXMLNode
		var walk func(*canonicalXMLNode)
		walk = func(node *canonicalXMLNode) {
			if node == nil || omitted != nil {
				return
			}
			if node.Name == "omitted" {
				omitted = node
				return
			}
			for _, child := range node.Children {
				walk(child)
			}
		}
		walk(root)
		if omitted == nil {
			t.Fatalf("SANY XML missing omitted proof node\n%s", xmlText)
		}
		locs := directChildren(omitted, "location")
		if len(locs) != 1 || firstChildText(directChildren(locs[0], "column")[0], "begin") != "3" {
			t.Fatalf("omitted proof location does not start at PROOF token\n%s", xmlText)
		}
	})

	t.Run("serializes WITNESS proof steps as witness builtin applications", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("ProofWitnessXML.tla", `---- MODULE ProofWitnessXML ----
CONSTANT x, y
THEOREM T == TRUE
PROOF
<1>1. WITNESS x, y
<1>. QED
  BY <1>1
====`)
		requireNoErrors(t, diags)
		root, err := parseCanonicalXML(xmlText)
		if err != nil {
			t.Fatalf("parse SANY XML: %v\n%s", err, xmlText)
		}
		witnessUID := xmlEntryUIDByKindAndName(root, "BuiltInKind", "$Witness")
		if witnessUID == "" {
			t.Fatalf("SANY XML missing $Witness builtin\n%s", xmlText)
		}
		node := firstOpApplNodeForOperatorUID(root, "BuiltInKindRef", witnessUID)
		if node == nil {
			t.Fatalf("proof step body missing $Witness application\n%s", xmlText)
		}
		operands := directChildren(node, "operands")
		if len(operands) != 1 || len(operands[0].Children) != 2 {
			t.Fatalf("$Witness operand count = %d containers/%d operands, want 1/2\n%s", len(operands), len(operands[0].Children), xmlText)
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

	t.Run("trims trailing whitespace from pre-comments", func(t *testing.T) {
		source := "---- MODULE PreCommentWhitespaceXML ----\n" +
			"\\* comment with trailing space \n" +
			"A == TRUE\n" +
			"===="
		xmlText, diags := SanyXMLSource("PreCommentWhitespaceXML.tla", source)
		requireNoErrors(t, diags)

		got := string(xmlText)
		if strings.Contains(got, `\* comment with trailing space ]]>`) {
			t.Fatalf("pre-comment retained trailing whitespace\n%s", got)
		}
		if !strings.Contains(got, `<![CDATA[\* comment with trailing space]]>`) {
			t.Fatalf("pre-comment missing trimmed text\n%s", got)
		}
	})

	t.Run("matches SANY nested block pre-comment tokenization", func(t *testing.T) {
		xmlText, diags := SanyXMLSource("NestedBlockPreCommentXML.tla", `---- MODULE NestedBlockPreCommentXML ----
(******************* disabled spec ********
  LET Helper ==
        (*********************)
        (* nested line       *)
        (*********************)
  IN Helper
  ***************************************)
A == TRUE
====`)
		requireNoErrors(t, diags)

		got := string(xmlText)
		for _, want := range []string{
			"        (*\n********************)\n\n",
			"        (*\n nested line       *)\n\n",
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("nested block pre-comment missing %q\n%s", want, got)
			}
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
B == {1} \times {2}
====`)
		requireNoErrors(t, diags)

		got := string(xmlText)
		if !strings.Contains(got, `<uniquename>$CartesianProd</uniquename>`) {
			t.Fatalf("Cartesian product XML missing $CartesianProd\n%s", got)
		}
		if strings.Contains(got, `<uniquename>\times</uniquename>`) {
			t.Fatalf("Cartesian product XML kept non-SANY \\times builtin\n%s", got)
		}
	})

	t.Run("does not reuse reserved builtin UIDs for generated symbols", func(t *testing.T) {
		var source strings.Builder
		source.WriteString("---- MODULE ReservedBuiltinUIDXML ----\n")
		source.WriteString("CONSTANTS ")
		for i := 0; i < 160; i++ {
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
		for _, uid := range []string{"294", "297"} {
			if strings.Contains(got, "<UID>"+uid+"</UID>\n      <OpDeclNode>") {
				t.Fatalf("reserved proof builtin UID %s was reused by a generated declaration\n%s", uid, got)
			}
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
Next == x' = [x EXCEPT ![1] = @ + 1, ![2] = @[2]]
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
		if strings.Contains(got, `<uniquename>@</uniquename>`) {
			t.Fatalf("EXCEPT @[i] emitted fake @ builtin\n%s", got)
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

func xmlEntryUIDByKindAndName(root *canonicalXMLNode, kind, name string) string {
	if payload := xmlEntryPayloadByKindAndName(root, kind, name); payload != nil {
		for _, entry := range canonicalSanyXMLEntries(root) {
			if canonicalSanyXMLEntryPayload(entry) == payload {
				return firstChildText(entry, "UID")
			}
		}
	}
	return ""
}

func xmlEntryPayloadByKindAndName(root *canonicalXMLNode, kind, name string) *canonicalXMLNode {
	for _, entry := range canonicalSanyXMLEntries(root) {
		payload := canonicalSanyXMLEntryPayload(entry)
		if payload == nil || payload.Name != kind || firstChildText(payload, "uniquename") != name {
			continue
		}
		return payload
	}
	return nil
}

func xmlPayloadCountByKindNameFile(root *canonicalXMLNode, kind, name, filename string) int {
	var count int
	for _, entry := range canonicalSanyXMLEntries(root) {
		payload := canonicalSanyXMLEntryPayload(entry)
		if payload == nil || payload.Name != kind || firstChildText(payload, "uniquename") != name {
			continue
		}
		if xmlNodeFilename(payload) == filename {
			count++
		}
	}
	return count
}

func substUsesReplacementRef(root *canonicalXMLNode, targetUID, refKind, replacementUID string) bool {
	for _, subst := range xmlNodesByName(root, "Subst") {
		if len(subst.Children) < 2 || firstChildText(subst.Children[0], "UID") != targetUID {
			continue
		}
		for _, operator := range directChildren(subst.Children[1], "operator") {
			for _, ref := range directChildren(operator, refKind) {
				if firstChildText(ref, "UID") == replacementUID {
					return true
				}
			}
		}
	}
	return false
}

func xmlNodesByName(root *canonicalXMLNode, name string) []*canonicalXMLNode {
	var out []*canonicalXMLNode
	var walk func(*canonicalXMLNode)
	walk = func(node *canonicalXMLNode) {
		if node == nil {
			return
		}
		if node.Name == name {
			out = append(out, node)
		}
		for _, child := range node.Children {
			walk(child)
		}
	}
	walk(root)
	return out
}

func firstOpApplNodeForOperatorUID(root *canonicalXMLNode, refKind, uid string) *canonicalXMLNode {
	var found *canonicalXMLNode
	var walk func(*canonicalXMLNode)
	walk = func(node *canonicalXMLNode) {
		if node == nil || found != nil {
			return
		}
		if node.Name == "OpApplNode" && opApplNodeUsesOperatorUID(node, refKind, uid) {
			found = node
			return
		}
		for _, child := range node.Children {
			walk(child)
		}
	}
	walk(root)
	return found
}

func opApplContainsDirectOperandWithOperator(root *canonicalXMLNode, parentUID, operandUID string) bool {
	var found bool
	var walk func(*canonicalXMLNode)
	walk = func(node *canonicalXMLNode) {
		if node == nil || found {
			return
		}
		if node.Name == "OpApplNode" && opApplNodeUsesOperatorUID(node, "BuiltInKindRef", parentUID) {
			for _, operands := range directChildren(node, "operands") {
				for _, operand := range directChildren(operands, "OpApplNode") {
					if opApplNodeUsesOperatorUID(operand, "BuiltInKindRef", operandUID) {
						found = true
						return
					}
				}
			}
		}
		for _, child := range node.Children {
			walk(child)
		}
	}
	walk(root)
	return found
}

func maxDirectOperandsForOperator(root *canonicalXMLNode, uid string) int {
	maxOperands := 0
	var walk func(*canonicalXMLNode)
	walk = func(node *canonicalXMLNode) {
		if node == nil {
			return
		}
		if node.Name == "OpApplNode" && opApplNodeUsesOperatorUID(node, "BuiltInKindRef", uid) {
			for _, operands := range directChildren(node, "operands") {
				if count := len(operands.Children); count > maxOperands {
					maxOperands = count
				}
			}
		}
		for _, child := range node.Children {
			walk(child)
		}
	}
	walk(root)
	return maxOperands
}

func moduleRefCountByPayloadLocation(root *canonicalXMLNode, moduleName, payloadKind, filename string, line int) int {
	uidPayloads := map[string]*canonicalXMLNode{}
	for _, entry := range canonicalSanyXMLEntries(root) {
		uid := firstChildText(entry, "UID")
		payload := canonicalSanyXMLEntryPayload(entry)
		if uid != "" && payload != nil {
			uidPayloads[uid] = payload
		}
	}
	var count int
	for _, entry := range canonicalSanyXMLEntries(root) {
		mod := canonicalSanyXMLEntryPayload(entry)
		if mod == nil || mod.Name != "ModuleNode" || firstChildText(mod, "uniquename") != moduleName {
			continue
		}
		for _, child := range mod.Children {
			if !strings.HasSuffix(child.Name, "Ref") {
				continue
			}
			payload := uidPayloads[firstChildText(child, "UID")]
			if payload != nil && payload.Name == payloadKind && xmlNodeFilename(payload) == filename && xmlNodeBeginLine(payload) == strconv.Itoa(line) {
				count++
			}
		}
	}
	return count
}

func moduleRefCountByPayloadName(root *canonicalXMLNode, moduleName, payloadKind, payloadName string) int {
	uidPayloads := map[string]*canonicalXMLNode{}
	for _, entry := range canonicalSanyXMLEntries(root) {
		uid := firstChildText(entry, "UID")
		payload := canonicalSanyXMLEntryPayload(entry)
		if uid != "" && payload != nil {
			uidPayloads[uid] = payload
		}
	}
	var count int
	for _, entry := range canonicalSanyXMLEntries(root) {
		mod := canonicalSanyXMLEntryPayload(entry)
		if mod == nil || mod.Name != "ModuleNode" || firstChildText(mod, "uniquename") != moduleName {
			continue
		}
		for _, child := range mod.Children {
			if !strings.HasSuffix(child.Name, "Ref") {
				continue
			}
			payload := uidPayloads[firstChildText(child, "UID")]
			if payload != nil && payload.Name == payloadKind && firstChildText(payload, "uniquename") == payloadName {
				count++
			}
		}
	}
	return count
}

func assertNoDanglingXMLRefs(t *testing.T, xmlText []byte) {
	t.Helper()
	root, err := parseCanonicalXML(xmlText)
	if err != nil {
		t.Fatalf("parse SANY XML: %v\n%s", err, xmlText)
	}
	entries := map[string]bool{}
	for _, entry := range canonicalSanyXMLEntries(root) {
		if uid := firstChildText(entry, "UID"); uid != "" {
			entries[uid] = true
		}
	}
	var missing []string
	var walk func(*canonicalXMLNode)
	walk = func(node *canonicalXMLNode) {
		if node == nil {
			return
		}
		if strings.HasSuffix(node.Name, "Ref") {
			uid := firstChildText(node, "UID")
			if uid != "" && !entries[uid] {
				missing = append(missing, node.Name+":"+uid)
			}
		}
		for _, child := range node.Children {
			walk(child)
		}
	}
	walk(root)
	if len(missing) > 0 {
		t.Fatalf("SANY XML has dangling refs %v\n%s", missing, xmlText)
	}
}

func xmlNodeFilename(node *canonicalXMLNode) string {
	for _, location := range directChildren(node, "location") {
		return firstChildText(location, "filename")
	}
	return ""
}

func xmlNodeBeginLine(node *canonicalXMLNode) string {
	for _, location := range directChildren(node, "location") {
		for _, line := range directChildren(location, "line") {
			return firstChildText(line, "begin")
		}
	}
	return ""
}

func opApplAtLocation(root *canonicalXMLNode, line, begin, end int) *canonicalXMLNode {
	var found *canonicalXMLNode
	var walk func(*canonicalXMLNode)
	walk = func(node *canonicalXMLNode) {
		if node == nil || found != nil {
			return
		}
		if node.Name == "OpApplNode" && xmlNodeLocationMatches(node, line, begin, end) {
			found = node
			return
		}
		for _, child := range node.Children {
			walk(child)
		}
	}
	walk(root)
	return found
}

func userDefinedOpAtLocation(root *canonicalXMLNode, line, begin, end int) *canonicalXMLNode {
	var found *canonicalXMLNode
	var walk func(*canonicalXMLNode)
	walk = func(node *canonicalXMLNode) {
		if node == nil || found != nil {
			return
		}
		if node.Name == "UserDefinedOpKind" && xmlNodeLocationMatches(node, line, begin, end) {
			found = node
			return
		}
		for _, child := range node.Children {
			walk(child)
		}
	}
	walk(root)
	return found
}

func firstXMLDescendant(root *canonicalXMLNode, name string) *canonicalXMLNode {
	if root == nil {
		return nil
	}
	if root.Name == name {
		return root
	}
	for _, child := range root.Children {
		if found := firstXMLDescendant(child, name); found != nil {
			return found
		}
	}
	return nil
}

func normalizeXMLForContains(text string) string {
	replacer := strings.NewReplacer(
		">\n", ">",
		"\n<", "<",
		">  <", "><",
		">    <", "><",
		">      <", "><",
		">        <", "><",
		">          <", "><",
	)
	prev := ""
	for text != prev {
		prev = text
		text = replacer.Replace(text)
	}
	return text
}

func xmlNodeLocationMatches(node *canonicalXMLNode, line, begin, end int) bool {
	locations := directChildren(node, "location")
	if len(locations) == 0 {
		return false
	}
	columns := directChildren(locations[0], "column")
	lines := directChildren(locations[0], "line")
	return len(columns) > 0 && len(lines) > 0 &&
		firstChildText(lines[0], "begin") == strconv.Itoa(line) &&
		firstChildText(columns[0], "begin") == strconv.Itoa(begin) &&
		firstChildText(columns[0], "end") == strconv.Itoa(end)
}

func opApplNodeUsesOperatorUID(node *canonicalXMLNode, refKind, uid string) bool {
	for _, child := range node.Children {
		if child.Name != "operator" {
			continue
		}
		for _, ref := range child.Children {
			if ref.Name == refKind && firstChildText(ref, "UID") == uid {
				return true
			}
		}
	}
	return false
}
