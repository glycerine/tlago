package sany_tests

import (
	"bytes"
	"encoding/xml"
	"io"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/xml/TestXMLExporterModule.java.
// Each test starts skipped until its Java assertions are ported and made green.
func TestTestXMLExporterModule_testExportDieHardModule(t *testing.T) {
	root, xmlText := checkedXMLExporterModule(t, "DieHard.tla")
	if root.Name != "modules" {
		t.Fatalf("root element = %s, want modules", root.Name)
	}
	rootModules := xmlDescendants(root, "RootModule")
	if got := len(rootModules); got != 1 {
		t.Fatalf("RootModule count = %d, want 1", got)
	}
	if got := strings.TrimSpace(rootModules[0].text()); got != "DieHard" {
		t.Fatalf("RootModule = %q, want DieHard", got)
	}
	if filename := rootModules[0].Attrs["filename"]; filename != "" && !strings.HasSuffix(filename, "DieHard.tla") {
		t.Fatalf("RootModule filename = %q, want suffix DieHard.tla", filename)
	}
	if got := len(xmlDescendants(root, "context")); got != 1 {
		t.Fatalf("context count = %d, want 1", got)
	}
	if got := len(xmlDescendants(root, "ModuleNode")); got == 0 {
		t.Fatal("missing ModuleNode elements")
	}
	if findXMLModuleByName(root, "DieHard") == nil {
		t.Fatal("DieHard ModuleNode not found")
	}
	requireXMLContainsAll(t, xmlText,
		">Min<", ">Init<", ">Next<", ">Spec<", ">Inv<", ">FixWater<",
		">NONDET<", ">FILL_BIG<", ">EMPTY_BIG<", ">FILL_SMALL<", ">EMPTY_SMALL<", ">BIG_TO_SMALL<", ">SMALL_TO_BIG<",
		">bigBucket<", ">smallBucket<", ">action<", ">water_to_pour<",
	)
	expected := map[string]bool{
		"DieHard": true, "Naturals": true, "Integers": true, "Sequences": true, "FiniteSets": true, "TLC": true,
	}
	for _, module := range xmlDescendants(root, "ModuleNode") {
		name := xmlUniqueName(module)
		if name != "" && !expected[name] {
			t.Fatalf("unexpected module %s", name)
		}
	}
}

func TestTestXMLExporterModule_testExportCaseOtherModule(t *testing.T) {
	root, _ := checkedXMLExporterModule(t, "CaseOtherXml.tla")
	for _, node := range xmlDescendants(root, "StringNode") {
		if child := xmlOnlyChildNamed(node, "StringValue"); child != nil && strings.TrimSpace(child.text()) == "$Other" {
			return
		}
	}
	t.Fatal(`XML missing StringNode/StringValue "$Other"`)
}

func TestTestXMLExporterModule_testExportWithOfflineMode(t *testing.T) {
	t.Skip("tla2sany wip")

	root, xmlText := runSANYXMLCommand(t, "-o", xmlExporterModulePath("DieHard.tla"))
	if root == nil || strings.TrimSpace(xmlText) == "" {
		t.Fatal("offline XML output is empty")
	}
}

func TestTestXMLExporterModule_testExportWithTerseMode(t *testing.T) {
	t.Skip("tla2sany wip")

	root, xmlText := runSANYXMLCommand(t, "-t", xmlExporterModulePath("DieHard.tla"))
	if root == nil || strings.TrimSpace(xmlText) == "" {
		t.Fatal("terse XML output is empty")
	}
}

func TestTestXMLExporterModule_testExportWithRestrictedMode(t *testing.T) {
	t.Skip("tla2sany wip")

	root, xmlText := runSANYXMLCommand(t, "-r", xmlExporterModulePath("DieHard.tla"))
	if root == nil || strings.TrimSpace(xmlText) == "" {
		t.Fatal("restricted XML output is empty")
	}
	if got := len(xmlDescendants(root, "RootModule")); got != 1 {
		t.Fatalf("RootModule count = %d, want 1", got)
	}
}

func TestTestXMLExporterModule_testRelationPreComments(t *testing.T) {
	root, _ := checkedXMLExporterModule(t, "Echo", "Relation.tla")
	preComments := xmlDescendants(root, "pre-comments")
	if len(preComments) == 0 {
		t.Fatal("Relation module has no pre-comments")
	}
	expected := map[string]string{
		"IsReflexive":                "Is the relation R reflexive over S?",
		"IsIrreflexive":              "Is the relation R irreflexive over set S?",
		"IsSymmetric":                "Is the relation R symmetric over set S?",
		"IsAsymmetric":               "Is the relation R asymmetric over set S?",
		"IsTransitive":               "Is the relation R transitive over set S?",
		"TransitiveClosure":          "Compute the transitive closure of relation R over set S",
		"ReflexiveTransitiveClosure": "Compute the reflexive transitive closure of relation R over set S",
		"IsConnected":                "Is the relation R connected over set S",
	}
	requireXMLPreComments(t, preComments, expected, false)
}

func TestTestXMLExporterModule_testTLACommentStylesPreComments(t *testing.T) {
	root, _ := checkedXMLExporterModule(t, "TLACommentStyles.tla")
	preComments := xmlDescendants(root, "pre-comments")
	if got := len(preComments); got != 4 {
		t.Fatalf("pre-comments count = %d, want 4", got)
	}
	expected := map[string]string{
		"CommentStyle1": `(*************************************************************************)
(* Calculate the sum of projections of the elements in a set.            *)
(*                                                                       *)
(* Example:                                                              *)
(*         MapThenSumSet(                                                *)
(*             LAMBDA e : e.n,                                           *)
(*             {[n |-> 0], [n |-> 1], [n |-> 2]}                         *)
(*         ) = 3                                                         *)
(*************************************************************************)`,
		"CommentStyle2": `(*************************************************************************)
(* COMMENT STYLE 2: Indented Boxed Comment                              *)
(* Used for nested or secondary explanations.                           *)
(* Note the indentation at the start.                                   *)
(*************************************************************************)`,
		"CommentStyle3": `(* COMMENT STYLE 3: Simple Multi-line Comment Without Box
   This style doesn't use asterisks on every line.
   It's more free-form and less structured.
   Often used for algorithm descriptions or citations.
 *)`,
		"CommentStyle4": "\\* Declaring instances local causes definition overrides to be hidden. In the\n" +
			"\\* case of Toolbox, this causes the definition override of `_TETrace` to be\n" +
			"\\* invisible.  In turn, TLC will then try to evaluate the TLA+ definition of\n" +
			"\\*\n" +
			"\\* `_TETrace` as defined in Tooblox.tla:\n" +
			"\\*   Attempted to enumerate S \\ T when S:\n" +
			"\\*   Nat\n" +
			"\\*   is not enumerable.\n" +
			"\\*\n" +
			"\\* See: https://github.com/tlaplus/CommunityModules/issues/37",
	}
	requireXMLPreComments(t, preComments, expected, true)
}

func TestTestXMLExporterModule_testNestedModuleIsChildOfEnclosingModule(t *testing.T) {
	root, _ := checkedXMLExporterModule(t, "NestedModuleXml.tla")
	nested := xmlNestedModules(root)
	requireXMLStringList(t, nested["NestedModuleXml"], []string{"Instantiated", "Standalone"}, "nested modules of NestedModuleXml")
	requireXMLStringList(t, nested["Standalone"], []string{"Innermost"}, "nested modules of Standalone")
	if _, ok := nested["Standalone"]; !ok {
		t.Fatal("unreferenced nested module Standalone should be exported")
	}
	requireXMLStringList(t, xmlTopLevelModules(root), []string{"Naturals", "NestedModuleXml"}, "top-level modules")
	assertEachXMLModuleReachableOnce(t, root)
}

func TestTestXMLExporterModule_testNestedModuleIsNotInheritedThroughExtends(t *testing.T) {
	root, _ := checkedXMLExporterModule(t, "NestedModuleXmlExtender.tla")
	nested := xmlNestedModules(root)
	requireXMLStringList(t, nested["NestedModuleXmlBase"], []string{"Sub"}, "nested modules of NestedModuleXmlBase")
	requireXMLStringList(t, nested["NestedModuleXmlExtender"], []string{"Borrower"}, "nested modules of NestedModuleXmlExtender")
	requireXMLStringList(t, nested["Borrower"], nil, "nested modules of Borrower")
	assertEachXMLModuleReachableOnce(t, root)
}

func TestTestXMLExporterModule_testLetInstanceOfEmptyModuleExportsModuleInstanceRef(t *testing.T) {
	root, _ := checkedXMLExporterModule(t, "Github1417.tla")
	requireXMLStringList(t, xmlLetInOpDefKinds(t, root, "op"), []string{"ModuleInstanceKindRef"}, "op LET opDefs")
	instance := xmlReferent(t, root, xmlLetInOpDefs(t, root, "op")[0])
	if instance.Name != "ModuleInstanceKind" {
		t.Fatalf("referent element = %s, want ModuleInstanceKind", instance.Name)
	}
	if got := xmlUniqueName(instance); got != "M" {
		t.Fatalf("instance uniquename = %s, want M", got)
	}
}

func TestTestXMLExporterModule_testLetInstanceWithNothingToInlineExportsModuleInstanceRef(t *testing.T) {
	root, _ := checkedXMLExporterModule(t, "Github1417Variants.tla")
	requireXMLStringList(t, xmlLetInOpDefKinds(t, root, "opLocal"), []string{"ModuleInstanceKindRef"}, "opLocal LET opDefs")
	if got := xmlUniqueName(xmlReferent(t, root, xmlLetInOpDefs(t, root, "opLocal")[0])); got != "Local" {
		t.Fatalf("opLocal instance = %s, want Local", got)
	}
	requireXMLStringList(t, xmlLetInOpDefKinds(t, root, "opParam"), []string{"ModuleInstanceKindRef"}, "opParam LET opDefs")
	if got := xmlUniqueName(xmlReferent(t, root, xmlLetInOpDefs(t, root, "opParam")[0])); got != "Param" {
		t.Fatalf("opParam instance = %s, want Param", got)
	}
}

func TestTestXMLExporterModule_testLetInstanceInlinesDefinitionsOfInstancee(t *testing.T) {
	root, _ := checkedXMLExporterModule(t, "Github1417Inlined.tla")
	if got := xmlUniqueName(xmlReferent(t, root, xmlLetInOpDef(t, root, "opOps", "UserDefinedOpKindRef"))); got != "Ops!foo" {
		t.Fatalf("opOps inlined operator = %s, want Ops!foo", got)
	}
	if got := xmlUniqueName(xmlReferent(t, root, xmlLetInOpDef(t, root, "opThm", "TheoremDefRef"))); got != "Thms!Thm" {
		t.Fatalf("opThm inlined theorem = %s, want Thms!Thm", got)
	}
	if got := xmlUniqueName(xmlReferent(t, root, xmlLetInOpDef(t, root, "opThm", "AssumeDefRef"))); got != "Thms!Asm" {
		t.Fatalf("opThm inlined assumption = %s, want Thms!Asm", got)
	}
}

func TestTestXMLExporterModule_testLetExportsEachModuleDefinitionIndependently(t *testing.T) {
	root, _ := checkedXMLExporterModule(t, "Github1417Variants.tla")
	requireXMLStringList(t, xmlLetInOpDefNames(t, root, "opMixed"), []string{"Both!foo", "Both", "Nothing"}, "opMixed LET opDef names")
}

func TestTestXMLExporterModule_testLetAlwaysExportsModuleInstanceRef(t *testing.T) {
	root, _ := checkedXMLExporterModule(t, "Github1417Inlined.tla")
	requireXMLStringList(t, xmlLetInOpDefNames(t, root, "opOps"), []string{"Ops!foo", "Ops"}, "opOps LET opDef names")
	inlined := xmlLetInOpDefNames(t, root, "opThm")
	if len(inlined) == 0 {
		t.Fatal("opThm LET opDefs are empty")
	}
	requireXMLStringSet(t, inlined[:len(inlined)-1], []string{"Thms!Thm", "Thms!Asm"}, "opThm inlined theorem/assumption")
	if got := inlined[len(inlined)-1]; got != "Thms" {
		t.Fatalf("opThm final LET opDef = %s, want Thms", got)
	}
}

func TestTestXMLExporterModule_testTopLevelInstanceOfEmptyModuleExportsInstanceNode(t *testing.T) {
	root, _ := checkedXMLExporterModule(t, "Github1417TopLevel.tla")
	instances := xmlDescendants(root, "InstanceNode")
	if got := len(instances); got != 1 {
		t.Fatalf("InstanceNode count = %d, want 1", got)
	}
	if got := xmlUniqueName(instances[0]); got != "M" {
		t.Fatalf("InstanceNode uniquename = %s, want M", got)
	}
	if got := strings.TrimSpace(xmlRequiredChild(t, instances[0], "module").text()); got != "Github1417Empty" {
		t.Fatalf("InstanceNode module = %s, want Github1417Empty", got)
	}
}

func TestTestXMLExporterModule_testUncommentFlagWithTLACommentStyles(t *testing.T) {
	t.Skip("tla2sany wip")

	root, _ := runSANYXMLCommand(t, "-u", xmlExporterModulePath("TLACommentStyles.tla"))
	preComments := xmlDescendants(root, "pre-comments")
	if got := len(preComments); got != 4 {
		t.Fatalf("pre-comments count = %d, want 4", got)
	}
	expected := map[string]string{
		"CommentStyle1": `Calculate the sum of projections of the elements in a set.

Example:
        MapThenSumSet(
            LAMBDA e : e.n,
            {[n |-> 0], [n |-> 1], [n |-> 2]}
        ) = 3`,
		"CommentStyle2": `COMMENT STYLE 2: Indented Boxed Comment
Used for nested or secondary explanations.
Note the indentation at the start.`,
		"CommentStyle4": "Declaring instances local causes definition overrides to be hidden. In the\n" +
			"case of Toolbox, this causes the definition override of `_TETrace` to be\n" +
			"invisible.  In turn, TLC will then try to evaluate the TLA+ definition of\n" +
			"\n" +
			"`_TETrace` as defined in Tooblox.tla:\n" +
			"  Attempted to enumerate S \\ T when S:\n" +
			"  Nat\n" +
			"  is not enumerable.\n" +
			"\n" +
			"See: https://github.com/tlaplus/CommunityModules/issues/37",
	}
	requireXMLPreComments(t, preComments, expected, false)
}

func TestTestXMLExporterModule_testUncommentFlagWithRelations(t *testing.T) {
	t.Skip("tla2sany wip")

	root, _ := runSANYXMLCommand(t, "-u", xmlExporterModulePath("Echo", "Relation.tla"))
	preComments := xmlDescendants(root, "pre-comments")
	if len(preComments) == 0 {
		t.Fatal("Relation module has no pre-comments")
	}
	expected := map[string]string{
		"IsReflexive":                "Is the relation R reflexive over S?",
		"IsIrreflexive":              "Is the relation R irreflexive over set S?",
		"IsSymmetric":                "Is the relation R symmetric over set S?",
		"IsAsymmetric":               "Is the relation R asymmetric over set S?",
		"IsTransitive":               "Is the relation R transitive over set S?",
		"TransitiveClosure":          "Compute the transitive closure of relation R over set S",
		"ReflexiveTransitiveClosure": "Compute the reflexive transitive closure of relation R over set S",
		"IsConnected":                "Is the relation R connected over set S",
	}
	requireXMLPreComments(t, preComments, expected, false)
}

func TestTestXMLExporterModule_testLetInstanceExportsInstantiatedTheoremAndAssumption(t *testing.T) {
	root, _ := checkedXMLExporterModule(t, "LetInstanceThmXml.tla")
	letIns := xmlDescendants(root, "LetInNode")
	if got := len(letIns); got != 1 {
		t.Fatalf("LetInNode count = %d, want 1", got)
	}
	if got := len(xmlDescendants(letIns[0], "TheoremDefRef")); got != 1 {
		t.Fatalf("TheoremDefRef count = %d, want 1", got)
	}
	if got := len(xmlDescendants(letIns[0], "AssumeDefRef")); got != 1 {
		t.Fatalf("AssumeDefRef count = %d, want 1", got)
	}
}

func TestTestXMLExporterModule_testRecursiveSectionGroupsJointDeclaration(t *testing.T) {
	root, _ := checkedXMLExporterModule(t, "RecursiveSectionXml.tla")
	f := xmlUserDefinedOpKind(t, root, "f")
	g := xmlUserDefinedOpKind(t, root, "g")
	h := xmlUserDefinedOpKind(t, root, "h")
	nonRecursive := xmlUserDefinedOpKind(t, root, "nonRecursive")
	fSection := xmlRecursiveSection(f)
	gSection := xmlRecursiveSection(g)
	hSection := xmlRecursiveSection(h)
	if fSection == "" || gSection == "" || hSection == "" {
		t.Fatalf("recursive sections: f=%q g=%q h=%q; want all non-empty", fSection, gSection, hSection)
	}
	if fSection != gSection {
		t.Fatalf("f and g recursiveSection differ: %q vs %q", fSection, gSection)
	}
	if hSection == fSection {
		t.Fatalf("h recursiveSection = %q, want distinct from f/g", hSection)
	}
	if got := xmlRecursiveSection(nonRecursive); got != "" {
		t.Fatalf("nonRecursive recursiveSection = %q, want empty", got)
	}
}

func TestTestXMLExporterModule_testUseHideDefsExportsModuleReference(t *testing.T) {
	root, _ := checkedXMLExporterModule(t, "UseHideModuleDefsXml.tla")
	useOrHide := xmlDescendants(root, "UseOrHideNode")
	if got := len(useOrHide); got != 1 {
		t.Fatalf("UseOrHideNode count = %d, want 1", got)
	}
	defs := xmlChildElements(xmlRequiredChild(t, useOrHide[0], "defs"))
	if got := len(defs); got != 2 {
		t.Fatalf("HIDE defs count = %d, want 2", got)
	}
	moduleRefs := 0
	for _, def := range defs {
		if def.Name == "ModuleNodeRef" {
			moduleRefs++
		}
	}
	if moduleRefs != 1 {
		t.Fatalf("ModuleNodeRef count in HIDE defs = %d, want 1", moduleRefs)
	}
}

type sanyXMLTestNode struct {
	Name     string
	Attrs    map[string]string
	Text     strings.Builder
	Parent   *sanyXMLTestNode
	Children []*sanyXMLTestNode
}

func xmlExporterModulePath(parts ...string) string {
	all := append([]string{"test-model"}, parts...)
	return sanyTestVectorPath(all...)
}

func checkedXMLExporterModule(t *testing.T, parts ...string) (*sanyXMLTestNode, string) {
	t.Helper()
	xmlText := checkedSANYXMLForPath(t, xmlExporterModulePath(parts...))
	return parseSANYXMLTestDocument(t, xmlText), xmlText
}

func runSANYXMLCommand(t *testing.T, args ...string) (*sanyXMLTestNode, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := tlago.RunCLI(append([]string{"sany-xml"}, args...), &stdout, &stderr)
	if code != tlago.ExitOK {
		t.Fatalf("sany-xml %v exit = %d, want %d; stderr=%s", args, code, tlago.ExitOK, stderr.String())
	}
	if got := strings.TrimSpace(stderr.String()); got != "" {
		t.Fatalf("sany-xml %v stderr = %q, want empty", args, got)
	}
	text := stdout.String()
	if strings.TrimSpace(text) == "" {
		t.Fatalf("sany-xml %v produced empty stdout", args)
	}
	return parseSANYXMLTestDocument(t, text), text
}

func parseSANYXMLTestDocument(t *testing.T, text string) *sanyXMLTestNode {
	t.Helper()
	decoder := xml.NewDecoder(strings.NewReader(text))
	var stack []*sanyXMLTestNode
	var root *sanyXMLTestNode
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("parse XML: %v\n%s", err, text)
		}
		switch token := token.(type) {
		case xml.StartElement:
			node := &sanyXMLTestNode{Name: token.Name.Local, Attrs: map[string]string{}}
			for _, attr := range token.Attr {
				node.Attrs[attr.Name.Local] = attr.Value
			}
			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				node.Parent = parent
				parent.Children = append(parent.Children, node)
			} else if root == nil {
				root = node
			} else {
				t.Fatalf("XML has multiple root elements: %s and %s", root.Name, node.Name)
			}
			stack = append(stack, node)
		case xml.EndElement:
			if len(stack) == 0 || stack[len(stack)-1].Name != token.Name.Local {
				t.Fatalf("unexpected XML end element %s", token.Name.Local)
			}
			stack = stack[:len(stack)-1]
		case xml.CharData:
			if len(stack) > 0 {
				stack[len(stack)-1].Text.Write([]byte(token))
			}
		}
	}
	if root == nil {
		t.Fatal("empty XML document")
	}
	if len(stack) != 0 {
		t.Fatalf("XML ended with %d unclosed elements", len(stack))
	}
	return root
}

func (node *sanyXMLTestNode) text() string {
	if node == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(node.Text.String())
	for _, child := range node.Children {
		b.WriteString(child.text())
	}
	return b.String()
}

func xmlDescendants(root *sanyXMLTestNode, name string) []*sanyXMLTestNode {
	var out []*sanyXMLTestNode
	var walk func(*sanyXMLTestNode)
	walk = func(node *sanyXMLTestNode) {
		if node.Name == name {
			out = append(out, node)
		}
		for _, child := range node.Children {
			walk(child)
		}
	}
	if root != nil {
		walk(root)
	}
	return out
}

func xmlChildElements(parent *sanyXMLTestNode) []*sanyXMLTestNode {
	if parent == nil {
		return nil
	}
	return parent.Children
}

func xmlOnlyChildNamed(parent *sanyXMLTestNode, name string) *sanyXMLTestNode {
	matches := xmlDirectChildren(parent, name)
	if len(matches) != 1 {
		return nil
	}
	return matches[0]
}

func xmlRequiredChild(t *testing.T, parent *sanyXMLTestNode, name string) *sanyXMLTestNode {
	t.Helper()
	matches := xmlDirectChildren(parent, name)
	if len(matches) != 1 {
		t.Fatalf("%s child %s count = %d, want 1", parent.Name, name, len(matches))
	}
	return matches[0]
}

func xmlDirectChildren(parent *sanyXMLTestNode, name string) []*sanyXMLTestNode {
	var out []*sanyXMLTestNode
	if parent == nil {
		return out
	}
	for _, child := range parent.Children {
		if child.Name == name {
			out = append(out, child)
		}
	}
	return out
}

func xmlUniqueName(node *sanyXMLTestNode) string {
	if child := xmlOnlyChildNamed(node, "uniquename"); child != nil {
		return strings.TrimSpace(child.text())
	}
	return ""
}

func findXMLModuleByName(root *sanyXMLTestNode, name string) *sanyXMLTestNode {
	for _, module := range xmlDescendants(root, "ModuleNode") {
		if xmlUniqueName(module) == name {
			return module
		}
	}
	return nil
}

func requireXMLContainsAll(t *testing.T, xmlText string, values ...string) {
	t.Helper()
	for _, value := range values {
		if !strings.Contains(xmlText, value) {
			t.Fatalf("XML missing %q\n%s", value, xmlText)
		}
	}
}

func requireXMLPreComments(t *testing.T, preComments []*sanyXMLTestNode, expected map[string]string, exact bool) {
	t.Helper()
	for _, preComment := range preComments {
		commentText := strings.TrimSpace(preComment.text())
		if commentText == "" {
			t.Fatal("pre-comment is empty")
		}
		opName := xmlUniqueName(preComment.Parent)
		want, ok := expected[opName]
		if !ok {
			continue
		}
		if exact {
			if commentText == want {
				delete(expected, opName)
			}
			continue
		}
		if strings.Contains(commentText, want) {
			delete(expected, opName)
		}
	}
	if len(expected) != 0 {
		t.Fatalf("missing expected pre-comments for %v", sortedXMLMapKeys(expected))
	}
}

func sortedXMLMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func xmlNestedModules(root *sanyXMLTestNode) map[string][]string {
	uidToName := map[string]string{}
	nestedUIDs := map[string][]string{}
	for _, entry := range xmlDescendants(root, "entry") {
		modules := xmlDescendants(entry, "ModuleNode")
		if len(modules) == 0 {
			continue
		}
		module := modules[0]
		uid := strings.TrimSpace(xmlDescendants(entry, "UID")[0].text())
		name := xmlUniqueName(module)
		uidToName[uid] = name
		for _, ref := range xmlDescendants(module, "ModuleNodeRef") {
			refUID := strings.TrimSpace(xmlRequiredDescendantText(ref, "UID"))
			nestedUIDs[name] = append(nestedUIDs[name], refUID)
		}
		if _, ok := nestedUIDs[name]; !ok {
			nestedUIDs[name] = nil
		}
	}
	result := map[string][]string{}
	for name, refs := range nestedUIDs {
		for _, uid := range refs {
			if childName, ok := uidToName[uid]; ok {
				result[name] = append(result[name], childName)
			} else {
				result[name] = append(result[name], "<unresolvable UID "+uid+">")
			}
		}
		if _, ok := result[name]; !ok {
			result[name] = nil
		}
	}
	return result
}

func xmlTopLevelModules(root *sanyXMLTestNode) []string {
	uidToName := map[string]string{}
	for _, entry := range xmlDescendants(root, "entry") {
		modules := xmlDescendants(entry, "ModuleNode")
		if len(modules) == 0 {
			continue
		}
		uid := strings.TrimSpace(xmlDescendants(entry, "UID")[0].text())
		uidToName[uid] = xmlUniqueName(modules[0])
	}
	var result []string
	for _, child := range root.Children {
		if child.Name != "ModuleNodeRef" {
			continue
		}
		uid := strings.TrimSpace(xmlRequiredDescendantText(child, "UID"))
		result = append(result, uidToName[uid])
	}
	return result
}

func assertEachXMLModuleReachableOnce(t *testing.T, root *sanyXMLTestNode) {
	t.Helper()
	reachable := append([]string{}, xmlTopLevelModules(root)...)
	nested := xmlNestedModules(root)
	for _, names := range nested {
		reachable = append(reachable, names...)
	}
	seen := map[string]bool{}
	for _, name := range reachable {
		if seen[name] {
			t.Fatalf("module %s is reachable more than once; reachable=%v", name, reachable)
		}
		seen[name] = true
	}
	var exported []string
	for name := range nested {
		exported = append(exported, name)
	}
	sort.Strings(exported)
	var reached []string
	for name := range seen {
		reached = append(reached, name)
	}
	sort.Strings(reached)
	if !reflect.DeepEqual(reached, exported) {
		t.Fatalf("reachable modules = %v, want exported modules %v", reached, exported)
	}
}

func requireXMLStringList(t *testing.T, got, want []string, label string) {
	t.Helper()
	if got == nil {
		got = []string{}
	}
	if want == nil {
		want = []string{}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s = %v, want %v", label, got, want)
	}
}

func requireXMLStringSet(t *testing.T, got, want []string, label string) {
	t.Helper()
	got = append([]string{}, got...)
	want = append([]string{}, want...)
	sort.Strings(got)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s = %v, want %v", label, got, want)
	}
}

func xmlLetInOpDefs(t *testing.T, root *sanyXMLTestNode, opName string) []*sanyXMLTestNode {
	t.Helper()
	definition := xmlUserDefinedOpKind(t, root, opName)
	letIns := xmlDescendants(definition, "LetInNode")
	if got := len(letIns); got != 1 {
		t.Fatalf("operator %s LetInNode count = %d, want 1", opName, got)
	}
	return xmlChildElements(xmlRequiredChild(t, letIns[0], "opDefs"))
}

func xmlLetInOpDef(t *testing.T, root *sanyXMLTestNode, opName, kind string) *sanyXMLTestNode {
	t.Helper()
	var matches []*sanyXMLTestNode
	for _, opDef := range xmlLetInOpDefs(t, root, opName) {
		if opDef.Name == kind {
			matches = append(matches, opDef)
		}
	}
	if len(matches) != 1 {
		t.Fatalf("operator %s LET opDef kind %s count = %d, want 1", opName, kind, len(matches))
	}
	return matches[0]
}

func xmlLetInOpDefKinds(t *testing.T, root *sanyXMLTestNode, opName string) []string {
	t.Helper()
	var kinds []string
	for _, opDef := range xmlLetInOpDefs(t, root, opName) {
		kinds = append(kinds, opDef.Name)
	}
	return kinds
}

func xmlLetInOpDefNames(t *testing.T, root *sanyXMLTestNode, opName string) []string {
	t.Helper()
	var names []string
	for _, opDef := range xmlLetInOpDefs(t, root, opName) {
		names = append(names, xmlUniqueName(xmlReferent(t, root, opDef)))
	}
	return names
}

func xmlReferent(t *testing.T, root, ref *sanyXMLTestNode) *sanyXMLTestNode {
	t.Helper()
	uid := strings.TrimSpace(xmlRequiredDescendantText(ref, "UID"))
	for _, entry := range xmlDescendants(root, "entry") {
		children := xmlChildElements(entry)
		if len(children) < 2 {
			continue
		}
		if strings.TrimSpace(children[0].text()) == uid {
			return children[1]
		}
	}
	t.Fatalf("%s reference UID %s does not resolve", ref.Name, uid)
	return nil
}

func xmlUserDefinedOpKind(t *testing.T, root *sanyXMLTestNode, opName string) *sanyXMLTestNode {
	t.Helper()
	for _, candidate := range xmlDescendants(root, "UserDefinedOpKind") {
		if xmlUniqueName(candidate) == opName {
			return candidate
		}
	}
	t.Fatalf("UserDefinedOpKind %s not found", opName)
	return nil
}

func xmlRecursiveSection(userDefinedOpKind *sanyXMLTestNode) string {
	sections := xmlDescendants(userDefinedOpKind, "recursiveSection")
	if len(sections) == 0 {
		return ""
	}
	return strings.TrimSpace(sections[0].text())
}

func xmlRequiredDescendantText(node *sanyXMLTestNode, name string) string {
	matches := xmlDescendants(node, name)
	if len(matches) == 0 {
		return ""
	}
	return matches[0].text()
}
