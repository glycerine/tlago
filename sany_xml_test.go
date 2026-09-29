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

func xmlEntryUIDByKindAndName(root *canonicalXMLNode, kind, name string) string {
	for _, entry := range canonicalSanyXMLEntries(root) {
		payload := canonicalSanyXMLEntryPayload(entry)
		if payload == nil || payload.Name != kind || firstChildText(payload, "uniquename") != name {
			continue
		}
		return firstChildText(entry, "UID")
	}
	return ""
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
