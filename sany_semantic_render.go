package tlago

import (
	"io"
	"os"
	"strconv"
	"unicode/utf16"

	"github.com/glycerine/tlago/tlc"
)

func sanySemanticNodeBaseToString(base *tlc.SemanticNodeBase, depth int) string {
	if base == nil {
		panic(tlc.NewNullPointerException())
	}
	if depth <= 0 {
		return ""
	}
	kind := int(base.Kind())
	kindName := "<none>"
	if kind != -1 {
		if kind < 0 || kind >= len(SanySemanticKindNames) {
			panic(tlc.NewArrayIndexOutOfBoundsException(kind, len(SanySemanticKindNames)))
		}
		kindName = SanySemanticKindNames[kind]
	}
	return "  uid: " + strconv.FormatInt(int64(base.GetUid()), 10) +
		"  kind: " + kindName + base.GetPreCommentsAsString()
}

// SanySemanticNodeBaseToString adapts SemanticNode.toString(int, Errors)
// without subclass dispatch, as used by source super calls.
func SanySemanticNodeBaseToString(base *tlc.SemanticNodeBase, depth int, errors *SanyErrors) string {
	return sanySemanticNodeBaseToString(base, depth)
}

func sanySymbolNodeBaseToString(node *tlc.SymbolNode, depth int) string {
	if node == nil || node.SemanticBase == nil {
		panic(tlc.NewNullPointerException())
	}
	return sanySemanticNodeBaseToString(node.SemanticBase, depth)
}

func sanyDefinitionModuleName(module *tlc.ModuleNode) string {
	if module == nil {
		return "<null>"
	}
	name := module.GetName()
	if name == nil {
		panic(tlc.NewNullPointerException())
	}
	return name.String()
}

// SanyFormalParamNodeToString adapts FormalParamNode.toString(int, Errors).
func SanyFormalParamNodeToString(node *tlc.SymbolNode, depth int, errors *SanyErrors) string {
	if node == nil {
		panic(tlc.NewNullPointerException())
	}
	if depth <= 0 {
		return ""
	}
	name := node.GetName()
	if name == nil {
		panic(tlc.NewNullPointerException())
	}
	return "\n*FormalParamNode: " + name.String() + "  " +
		sanySymbolNodeBaseToString(node, depth) + "  arity: " + strconv.Itoa(node.GetArity())
}

func sanyOpDefOrDeclNodeToString(node *tlc.SymbolNode, depth int) string {
	if depth <= 0 {
		return ""
	}
	return sanySymbolNodeBaseToString(node, depth) + "  arity: " + strconv.Itoa(node.GetArity()) +
		"  orgDefInModule: " + sanyDefinitionModuleName(node.GetOriginallyDefinedInModuleNode())
}

// SanyOpDefOrDeclNodeToString adapts OpDefOrDeclNode.toString(int, Errors)
// without subclass dispatch.
func SanyOpDefOrDeclNodeToString(node *tlc.SymbolNode, depth int, errors *SanyErrors) string {
	if node == nil {
		panic(tlc.NewNullPointerException())
	}
	return sanyOpDefOrDeclNodeToString(node, depth)
}

// SanyOpDeclNodeToString adapts OpDeclNode.toString(int, Errors).
func SanyOpDeclNodeToString(node *tlc.SymbolNode, depth int, errors *SanyErrors) string {
	if node == nil {
		panic(tlc.NewNullPointerException())
	}
	if depth <= 0 {
		return ""
	}
	name := "null"
	if node.GetName() != nil {
		name = node.GetName().String()
	}
	return "\n*OpDeclNode: " + name + "  " + sanyOpDefOrDeclNodeToString(node, depth) +
		"\n  originallyDefinedInModule: " + sanyDefinitionModuleName(node.GetOriginallyDefinedInModuleNode())
}

// SanyOpArgNodeToString adapts OpArgNode.toString(int, Errors).
func SanyOpArgNodeToString(node *tlc.OpArgNode, depth int, errors *SanyErrors) string {
	if node == nil {
		panic(tlc.NewNullPointerException())
	}
	if depth <= 0 {
		return ""
	}
	name := "null"
	if node.GetName() != nil {
		name = node.GetName().String()
	}
	op := "null"
	if node.GetOp() != nil {
		if node.GetOp().SemanticBase == nil {
			panic(tlc.NewNullPointerException())
		}
		op = strconv.FormatInt(int64(node.GetOp().SemanticBase.GetUid()), 10)
	}
	return "\n*OpArgNode: " + name + "  " +
		sanySemanticNodeBaseToString(node.SemanticNodeBase, depth) +
		"  arity: " + strconv.Itoa(node.GetArity()) + "  op: " + op
}

func sanySemanticDepthString(node tlc.SemanticNode, depth int, errors *SanyErrors) string {
	if sanyExploreNull(node) {
		panic(tlc.NewNullPointerException())
	}
	if depth <= 0 {
		return ""
	}
	switch node := node.(type) {
	case *tlc.NullSemanticNode:
		return sanySemanticNodeBaseToString(&node.SemanticNodeBase, depth)
	case *tlc.NumeralNode:
		return SanyNumeralNodeToString(node, depth, errors)
	case *tlc.DecimalNode:
		return SanyDecimalNodeToString(node, depth, errors)
	case *tlc.StringNode:
		return SanyStringNodeToString(node, depth, errors)
	case *tlc.OpArgNode:
		return SanyOpArgNodeToString(node, depth, errors)
	case *tlc.OpApplNode:
		return SanyOpApplNodeToString(node, depth, errors)
	case *tlc.AtNode:
		return SanyAtNodeToString(node, depth, errors)
	case *tlc.LetInNode:
		return SanyLetInNodeToString(node, depth, errors)
	case *tlc.LabelNode:
		return SanyLabelNodeToString(node, depth, errors)
	case *tlc.ThmOrAssumpDefNode:
		return SanyThmOrAssumpDefNodeToString(node, depth, errors)
	case *tlc.SubstInNode:
		return SanySubstInNodeToString(node, depth, errors)
	case *tlc.APSubstInNode:
		return SanyAPSubstInNodeToString(node, depth, errors)
	case *tlc.AssumeProveNode:
		return SanyAssumeProveNodeToString(node, depth, errors)
	case *tlc.NewSymbNode:
		return SanyNewSymbNodeToString(node, depth, errors)
	case *tlc.TheoremNode:
		return SanyTheoremNodeToString(node, depth, errors)
	case *tlc.LeafProofNode:
		return SanyLeafProofNodeToString(node, depth, errors)
	case *tlc.NonLeafProofNode:
		return SanyNonLeafProofNodeToString(node, depth, errors)
	case *tlc.DefStepNode:
		return SanyDefStepNodeToString(node, depth, errors)
	case *tlc.UseOrHideNode:
		return SanyUseOrHideNodeToString(node, depth, errors)
	case *tlc.InstanceNode:
		return SanyInstanceNodeToString(node, depth, errors)
	case *tlc.OpDefNode:
		return SanyOpDefNodeToString(node, depth, errors)
	case *tlc.AssumeNode:
		return SanyAssumeNodeToString(node, depth, errors)
	case *tlc.ModuleNode:
		return SanyModuleNodeToString(node, depth, errors)
	case *tlc.SymbolNode:
		if node.Definition != nil {
			return SanyOpDefNodeToString(node.Definition, depth, errors)
		}
		if node.Kind == tlc.SymbolFormalParam {
			return SanyFormalParamNodeToString(node, depth, errors)
		}
		return SanyOpDeclNodeToString(node, depth, errors)
	default:
		panic(tlc.NewUnsupportedOperationException("SANY depth rendering is not implemented for this semantic node"))
	}
}

// SanySemanticNodeToString dispatches SemanticNode.toString(int, Errors) to
// the runtime node's concrete source-class renderer.
func SanySemanticNodeToString(node tlc.SemanticNode, depth int, errors *SanyErrors) string {
	return sanySemanticDepthString(node, depth, errors)
}

// SanyAssumeNodeToString adapts AssumeNode.toString(int, Errors).
func SanyAssumeNodeToString(node *tlc.AssumeNode, depth int, errors *SanyErrors) string {
	if node == nil {
		panic(tlc.NewNullPointerException())
	}
	if depth <= 0 {
		return ""
	}
	inner := "\n*AssumeNode " + sanySemanticNodeBaseToString(node.SemanticNodeBase, depth)
	if !sanyExploreNull(node.Assume) {
		inner += sanyIndent(2, sanySemanticDepthString(node.Assume, depth-1, errors))
	}
	result := sanyIndent(2, inner)
	if node.Def != nil {
		result += sanyIndent(4, "\n def: "+sanyIndent(2, sanySemanticDepthString(node.Def, depth-1, errors)))
	}
	return result
}

func sanyModuleIsConstant(node *tlc.ModuleNode) bool {
	if len(node.GetVariableDecls()) > 0 {
		return false
	}
	for _, definition := range node.GetOpDefs() {
		if definition.Kind() == tlc.SemanticModuleInstanceKind {
			continue
		}
		body := definition.GetBody()
		if sanyExploreNull(body) {
			panic(tlc.NewNullPointerException())
		}
		leveled, ok := body.(interface{ GetLevel() int })
		if !ok {
			panic(tlc.NewClassCastException("operator body is not a LevelNode"))
		}
		if leveled.GetLevel() != tlc.TLCLevelConstant {
			return false
		}
	}
	for _, theorem := range node.GetTheorems() {
		if theorem.GetLevel() != tlc.TLCLevelConstant {
			return false
		}
	}
	return true
}

func sanyExtendedModuleSetString(node *tlc.ModuleNode) string {
	hash := func(module *tlc.ModuleNode) int32 {
		return 31*(31+int32(module.Kind())) + module.GetUid()
	}
	equal := func(left, right *tlc.ModuleNode) bool {
		return left == right || left != nil && right != nil && left.Kind() == right.Kind() && left.GetUid() == right.GetUid()
	}
	tieBreak := func(left, right *tlc.ModuleNode) int {
		if left.GetUid() < right.GetUid() {
			return -1
		}
		if left.GetUid() > right.GetUid() {
			return 1
		}
		return 0
	}
	cache := make(map[*tlc.ModuleNode]*tlc.JavaSemanticMap[*tlc.ModuleNode, struct{}])
	var build func(*tlc.ModuleNode) *tlc.JavaSemanticMap[*tlc.ModuleNode, struct{}]
	build = func(module *tlc.ModuleNode) *tlc.JavaSemanticMap[*tlc.ModuleNode, struct{}] {
		if set := cache[module]; set != nil {
			return set
		}
		set := tlc.NewJavaSemanticMap[*tlc.ModuleNode, struct{}](hash, equal, tieBreak)
		cache[module] = set
		for _, extendee := range module.Extendees {
			set.Put(extendee, struct{}{})
			for inherited := range build(extendee).All() {
				set.Put(inherited, struct{}{})
			}
		}
		return set
	}
	result := "{"
	first := true
	for module := range build(node).All() {
		if !first {
			result += ", "
		}
		first = false
		name := module.GetName()
		if name == nil {
			panic(tlc.NewNullPointerException())
		}
		result += name.String()
	}
	return result + "}"
}

// SanyModuleNodeToString adapts ModuleNode.toString(int, Errors).
func SanyModuleNodeToString(node *tlc.ModuleNode, depth int, errors *SanyErrors) string {
	if node == nil {
		panic(tlc.NewNullPointerException())
	}
	if depth <= 0 {
		return ""
	}
	name := node.GetName()
	if name == nil {
		panic(tlc.NewNullPointerException())
	}
	errorString := "null"
	if errors != nil {
		if count := errors.GetNumErrors(); count == 0 {
			errorString = "none"
		} else {
			errorString = strconv.Itoa(count)
		}
	}
	result := "\n*ModuleNode: " + name.String() + "  " +
		sanySemanticNodeBaseToString(node.SemanticNodeBase, depth) +
		"  constant module: " + strconv.FormatBool(sanyModuleIsConstant(node)) +
		"  errors: " + errorString
	for _, entry := range sanyContextEntryStrings(node.Context, depth-1, false, errors) {
		if entry == "" {
			result += "*** null ***"
		} else {
			result += sanyIndent(2, entry)
		}
	}
	result += sanyIndent(2, "\nAllExtended: "+sanyExtendedModuleSetString(node))
	if len(node.InstanceVec) > 0 {
		result += sanyIndent(2, "\nInstantiations:")
		for _, instance := range node.InstanceVec {
			result += sanyIndent(4, sanySemanticDepthString(instance, 1, errors))
		}
	}
	if assumptions := node.GetAssumptions(); len(assumptions) > 0 {
		result += sanyIndent(2, "\nAssumptions:")
		for _, assumption := range assumptions {
			result += sanyIndent(4, sanySemanticDepthString(assumption, 1, errors))
		}
	}
	if theorems := node.GetTheorems(); len(theorems) > 0 {
		result += sanyIndent(2, "\nTheorems:")
		for _, theorem := range theorems {
			result += sanyIndent(4, sanySemanticDepthString(theorem, 1, errors))
		}
	}
	if len(node.TopLevel) > 0 {
		result += sanyIndent(2, "\ntopLevelVec: ")
		for _, topLevel := range node.TopLevel {
			result += sanyIndent(4, sanySemanticDepthString(topLevel, 1, errors))
		}
	}
	return result
}

// SanyModuleNodePrint adapts ModuleNode.print(int, int, boolean, Errors).
func SanyModuleNodePrint(node *tlc.ModuleNode, indent, depth int, includeInitial bool, errors *SanyErrors) {
	SanyModuleNodePrintTo(os.Stdout, node, indent, depth, includeInitial, errors)
}

// SanyModuleNodePrintTo retains ModuleNode.print's formatting while allowing
// an owned writer at the Go boundary. As with PrintStream, write failures are
// not returned.
func SanyModuleNodePrintTo(out io.Writer, node *tlc.ModuleNode, indent, depth int, includeInitial bool, errors *SanyErrors) {
	if node == nil || out == nil {
		panic(tlc.NewNullPointerException())
	}
	if depth <= 0 {
		return
	}
	name := node.GetName()
	if name == nil {
		panic(tlc.NewNullPointerException())
	}
	errorString := "null"
	if errors != nil {
		if count := errors.GetNumErrors(); count == 0 {
			errorString = "none"
		} else {
			errorString = strconv.Itoa(count)
		}
	}
	_, _ = io.WriteString(out, "*ModuleNode: "+name.String()+"  "+
		sanySemanticNodeBaseToString(node.SemanticNodeBase, depth)+
		"  errors: "+errorString)
	for _, entry := range sanyContextEntryStrings(node.Context, depth-1, includeInitial, errors) {
		_, _ = io.WriteString(out, sanyIndent(2+indent, entry))
	}
}

// SanyNewSymbNodeToString adapts NewSymbNode.toString(int, Errors).
func SanyNewSymbNodeToString(node *tlc.NewSymbNode, depth int, errors *SanyErrors) string {
	if node == nil {
		panic(tlc.NewNullPointerException())
	}
	if depth <= 0 {
		return ""
	}
	set := ""
	if !sanyExploreNull(node.Set) {
		set = sanyIndent(2, "\nSet:"+sanyIndent(2, sanySemanticDepthString(node.Set, depth-1, errors)))
	}
	return "\n*NewSymbNode:   " + sanySemanticNodeBaseToString(node.SemanticNodeBase, depth) +
		sanyIndent(2, "\nKind: "+strconv.Itoa(int(node.Kind()))+
			"\nopDeclNode:"+sanyIndent(2, sanySemanticDepthString(node.OpDecl, depth-1, errors))+set)
}

// SanyTheoremNodeToString adapts TheoremNode.toString(int, Errors).
func SanyTheoremNodeToString(node *tlc.TheoremNode, depth int, errors *SanyErrors) string {
	if node == nil {
		panic(tlc.NewNullPointerException())
	}
	if depth <= 0 {
		return ""
	}
	result := "\n*TheoremNode " + sanySemanticNodeBaseToString(node.SemanticNodeBase, depth)
	if !sanyExploreNull(node.Theorem) {
		result += sanyIndent(2, sanySemanticDepthString(node.Theorem, depth-1, errors))
	}
	if node.Def != nil {
		result += sanyIndent(2, "\n def: "+sanyIndent(2, sanySemanticDepthString(node.Def, depth-1, errors)))
	}
	if node.Suffices {
		result += sanyIndent(2, "\n SUFFICES step")
	}
	if !sanyExploreNull(node.Proof) {
		result += sanyIndent(2, "\n proof: "+sanyIndent(2, sanySemanticDepthString(node.Proof, depth-1, errors)))
	}
	return result
}

// SanyLeafProofNodeToString adapts LeafProofNode.toString(int, Errors).
func SanyLeafProofNodeToString(node *tlc.LeafProofNode, depth int, errors *SanyErrors) string {
	if node == nil {
		panic(tlc.NewNullPointerException())
	}
	if depth <= 0 {
		return ""
	}
	if node.Facts == nil || node.Defs == nil {
		panic(tlc.NewNullPointerException())
	}
	result := "\n*LeafProofNode:\n" + sanySemanticNodeBaseToString(node.SemanticNodeBase, depth) +
		sanyIndent(2, "\nfacts:")
	for _, fact := range node.Facts {
		result += sanyIndent(4, sanySemanticDepthString(fact, depth-1, errors))
	}
	result += sanyIndent(2, "\ndefs:")
	for _, definition := range node.Defs {
		result += sanyIndent(4, sanySemanticDepthString(definition, depth-1, errors))
	}
	return result + sanyIndent(2, "\nomitted: "+strconv.FormatBool(node.Omitted)) +
		sanyIndent(2, "\nonlyFlag: "+strconv.FormatBool(node.Only))
}

// SanyNonLeafProofNodeToString adapts NonLeafProofNode.toString(int, Errors).
func SanyNonLeafProofNodeToString(node *tlc.NonLeafProofNode, depth int, errors *SanyErrors) string {
	if node == nil {
		panic(tlc.NewNullPointerException())
	}
	if depth <= 0 {
		return ""
	}
	if node.Steps == nil {
		panic(tlc.NewNullPointerException())
	}
	result := "\n*ProofNode:\n" + sanySemanticNodeBaseToString(node.SemanticNodeBase, depth) +
		sanyIndent(2, "\nsteps:")
	for _, step := range node.Steps {
		result += sanyIndent(4, sanySemanticDepthString(step, depth-1, errors))
	}
	for _, entry := range sanyContextEntryStrings(node.Context, depth-1, false, errors) {
		if entry == "" {
			result += "*** null ***"
		} else {
			result += sanyIndent(2, entry)
		}
	}
	return result
}

// SanyDefStepNodeToString adapts DefStepNode.toString(int, Errors).
func SanyDefStepNodeToString(node *tlc.DefStepNode, depth int, errors *SanyErrors) string {
	if node == nil {
		panic(tlc.NewNullPointerException())
	}
	if depth <= 0 {
		return ""
	}
	if node.Defs == nil {
		panic(tlc.NewNullPointerException())
	}
	result := "\n*DefStepNode:\n" + sanySemanticNodeBaseToString(node.SemanticNodeBase, depth) +
		sanyIndent(2, "\ndefs:")
	for _, definition := range node.Defs {
		result += sanyIndent(4, sanySemanticDepthString(definition, depth-1, errors))
	}
	return result
}

// SanyUseOrHideNodeToString adapts UseOrHideNode.toString(int, Errors).
func SanyUseOrHideNodeToString(node *tlc.UseOrHideNode, depth int, errors *SanyErrors) string {
	if node == nil {
		panic(tlc.NewNullPointerException())
	}
	if depth <= 0 {
		return ""
	}
	if node.Facts == nil || node.Defs == nil {
		panic(tlc.NewNullPointerException())
	}
	result := "\n*UseOrHideNode:\n" + sanySemanticNodeBaseToString(node.SemanticNodeBase, depth) +
		sanyIndent(2, "\nisOnly: "+strconv.FormatBool(node.Only)) + sanyIndent(2, "\nfacts:")
	for _, fact := range node.Facts {
		result += sanyIndent(4, sanySemanticDepthString(fact, 1, errors))
	}
	result += sanyIndent(2, "\ndefs:")
	for _, definition := range node.Defs {
		result += sanyIndent(4, sanySemanticDepthString(definition, 1, errors))
	}
	return result
}

// SanyInstanceNodeToString adapts InstanceNode.toString(int, Errors).
func SanyInstanceNodeToString(node *tlc.InstanceNode, depth int, errors *SanyErrors) string {
	if node == nil {
		panic(tlc.NewNullPointerException())
	}
	if depth <= 0 {
		return ""
	}
	if node.Module == nil || node.Params == nil || node.Substs == nil {
		panic(tlc.NewNullPointerException())
	}
	moduleName := node.Module.GetName()
	if moduleName == nil {
		panic(tlc.NewNullPointerException())
	}
	instanceName := "(none)"
	if node.Name != nil {
		instanceName = node.Name.String()
	}
	result := "\n*InstanceNode " + sanySemanticNodeBaseToString(node.SemanticNodeBase, depth) +
		"  InstanceName = " + instanceName + sanyIndent(2, "\nModule: "+moduleName.String()) +
		sanyIndent(2, "\nlocal: "+strconv.FormatBool(node.GetLocal()))
	if len(node.Params) > 0 {
		result += sanyIndent(2, "\nInstance parameters:")
		for _, parameter := range node.Params {
			result += sanyIndent(4, sanySemanticDepthString(parameter, depth-1, errors))
		}
	}
	if len(node.Substs) > 0 {
		result += sanyIndent(2, "\nSubstitutions:")
		for index := range node.Substs {
			result += sanyIndent(2, sanyIndent(2, "\nSubst:"+
				sanyIndent(2, SanySubstToString(&node.Substs[index], depth-1, errors))))
		}
	}
	return result
}

// SanyAssumeProveNodeToString adapts AssumeProveNode.toString(int, Errors).
func SanyAssumeProveNodeToString(node *tlc.AssumeProveNode, depth int, errors *SanyErrors) string {
	if node == nil {
		panic(tlc.NewNullPointerException())
	}
	if depth <= 0 {
		return ""
	}
	assumes := ""
	for _, assume := range node.Assumes {
		assumes += sanyIndent(4, sanySemanticDepthString(assume, depth-1, errors))
	}
	goal := "null"
	if node.Goal != nil {
		goal = sanyIndent(4, sanySemanticDepthString(node.Goal, 1, errors))
	}
	boxed := ""
	if node.GetIsBoxAssumeProve() {
		boxed = "[]"
	}
	result := "\n*AssumeProveNode: " + sanySemanticNodeBaseToString(node.SemanticNodeBase, depth) +
		"\n  " + boxed + "Assumes: " + assumes +
		"\n  " + boxed + "Prove: " + sanyIndent(4, sanySemanticDepthString(node.Prove, depth-1, errors)) +
		"\n  Goal: " + goal
	if node.GetSuffices() {
		result += "\n  SUFFICES"
	}
	return result
}

// SanySubstToString adapts Subst.toString(int, Errors).
func SanySubstToString(subst *tlc.Subst, depth int, errors *SanyErrors) string {
	if subst == nil {
		panic(tlc.NewNullPointerException())
	}
	op := "<null>"
	if subst.GetOp() != nil {
		op = sanySemanticDepthString(subst.GetOp(), depth-1, errors)
	}
	expr := "<null>"
	if !sanyExploreNull(subst.GetExpr()) {
		expr = sanySemanticDepthString(subst.GetExpr(), depth-1, errors)
	}
	return "\nOp: " + sanyIndent(2, op) + "\nExpr: " + sanyIndent(2, expr)
}

func sanySubstitutionNodeToString(kind string, base *tlc.SemanticNodeBase, substs []tlc.Subst,
	body tlc.SemanticNode, instantiating, instantiated *tlc.ModuleNode, depth int, errors *SanyErrors) string {
	if depth <= 0 {
		return ""
	}
	if instantiating == nil || instantiated == nil {
		panic(tlc.NewNullPointerException())
	}
	instantiatingName := instantiating.GetName()
	instantiatedName := instantiated.GetName()
	if instantiatingName == nil || instantiatedName == nil {
		panic(tlc.NewNullPointerException())
	}
	result := "\n*" + kind + ": " + sanySemanticNodeBaseToString(base, depth) +
		"\n  instantiating module: " + instantiatingName.String() +
		", instantiated module: " + instantiatedName.String() +
		sanyIndent(2, "\nSubstitutions:")
	if substs == nil {
		result += sanyIndent(2, "<null>")
	} else {
		for index := range substs {
			result += sanyIndent(2, sanyIndent(2, "\nSubst:"+
				sanyIndent(2, SanySubstToString(&substs[index], depth-1, errors))))
		}
	}
	bodyString := "<null>"
	if !sanyExploreNull(body) {
		bodyString = sanySemanticDepthString(body, depth-1, errors)
	}
	return result + sanyIndent(2, "\nBody:"+sanyIndent(2, bodyString))
}

// SanySubstInNodeToString adapts SubstInNode.toString(int, Errors).
func SanySubstInNodeToString(node *tlc.SubstInNode, depth int, errors *SanyErrors) string {
	if node == nil {
		panic(tlc.NewNullPointerException())
	}
	return sanySubstitutionNodeToString("SubstInNode", node.SemanticNodeBase, node.Substs, node.GetBody(),
		node.InstantiatingModule, node.InstantiatedModule, depth, errors)
}

// SanyAPSubstInNodeToString adapts APSubstInNode.toString(int, Errors).
func SanyAPSubstInNodeToString(node *tlc.APSubstInNode, depth int, errors *SanyErrors) string {
	if node == nil {
		panic(tlc.NewNullPointerException())
	}
	return sanySubstitutionNodeToString("APSubstInNode", node.SemanticNodeBase, node.Substs, node.GetBody(),
		node.InstantiatingModule, node.InstantiatedModule, depth, errors)
}

func sanyLabelNamesToString(labels *tlc.SemanticLabelTable) string {
	if labels == nil {
		return "\n  Labels: null"
	}
	result := "\n  Labels: "
	for _, label := range labels.Elements() {
		name := label.GetName()
		if name == nil {
			panic(tlc.NewNullPointerException())
		}
		result += name.String() + "  "
	}
	return result
}

// SanyThmOrAssumpDefNodeToString adapts
// ThmOrAssumpDefNode.toString(int, Errors).
func SanyThmOrAssumpDefNodeToString(node *tlc.ThmOrAssumpDefNode, depth int, errors *SanyErrors) string {
	if node == nil {
		panic(tlc.NewNullPointerException())
	}
	if depth <= 0 {
		return ""
	}
	if node.Name == nil {
		panic(tlc.NewNullPointerException())
	}
	module := "<null>"
	if node.OriginallyDefinedInModule != nil {
		name := node.OriginallyDefinedInModule.GetName()
		if name == nil {
			panic(tlc.NewNullPointerException())
		}
		module = name.String()
	}
	result := "\n*ThmOrAssumpDefNode: " + node.Name.String() + "  " +
		sanySemanticNodeBaseToString(node.SemanticNodeBase, depth) +
		" arity: " + strconv.Itoa(node.GetArity()) + " module: " + module
	if node.InstantiatedFrom != nil {
		name := node.InstantiatedFrom.GetName()
		if name == nil {
			panic(tlc.NewNullPointerException())
		}
		result += " instantiatedFrom: " + name.String()
	}
	parameters := node.GetParams()
	if parameters != nil {
		params := "\nFormal params: " + strconv.Itoa(len(parameters))
		for _, parameter := range parameters {
			params += sanyIndent(2, sanySemanticDepthString(parameter, depth-1, errors))
		}
		result += sanyIndent(2, params)
	}
	body := node.GetBody()
	if !sanyExploreNull(body) {
		result += sanyIndent(2, "\nisTheorem(): "+strconv.FormatBool(node.IsTheorem())+
			"\nBody:"+sanyIndent(2, sanySemanticDepthString(body, depth-1, errors))+
			"\nsuffices: "+strconv.FormatBool(node.IsSuffices()))
	}
	return result + sanyLabelNamesToString(node.GetLabelsHT())
}

// SanyLabelNodeToString adapts LabelNode.toString(int, Errors).
func SanyLabelNodeToString(node *tlc.LabelNode, depth int, errors *SanyErrors) string {
	if node == nil {
		panic(tlc.NewNullPointerException())
	}
	if depth <= 0 {
		return ""
	}
	if node.Name == nil {
		panic(tlc.NewNullPointerException())
	}
	result := "\n*LabelNode: " + sanySemanticNodeBaseToString(node.SemanticNodeBase, depth)
	result += sanyIndent(2, "\nname: "+node.Name.String())
	for index, parameter := range node.Params {
		result += sanyIndent(2, "\nparam["+strconv.Itoa(index)+"]:"+
			sanyIndent(2, sanySemanticDepthString(parameter, depth-1, errors)))
	}
	result += sanyIndent(2, "\nisAssumeProve: "+strconv.FormatBool(node.IsAssumeProve))
	result += sanyIndent(2, "\nBody:"+sanyIndent(2, sanySemanticDepthString(node.Body, depth-1, errors)))
	result += sanyLabelNamesToString(node.GetLabelsHT())
	if node.SubExpressionOf != nil {
		result += sanyIndent(2, "\nsubExpressionOf: "+
			sanyIndent(2, sanySemanticDepthString(node.SubExpressionOf, 1, errors)))
	}
	if !sanyExploreNull(node.Goal) {
		result += "\n goal: " + sanyIndent(4, sanySemanticDepthString(node.Goal, 1, errors)) +
			"\n goalClause: " + strconv.Itoa(node.GoalClause)
	}
	return result
}

func sanyContextEntryStrings(context *SanyContext, depth int, includeInitial bool, errors *SanyErrors) []string {
	if context == nil || context.ModuleTable == nil {
		panic(tlc.NewNullPointerException())
	}
	naturals := context.ModuleTable.GetContext(tlc.UniqueStringOf("Naturals"))
	entries := make([]string, 0)
	if depth <= 0 {
		return entries
	}
	initial := SanyContextGetGlobalContext()
	for _, pair := range context.SnapshotPairHistory() {
		name := sanySymbolTableNodeName(pair.Node)
		key := SanyContextKey{Name: name}
		if !includeInitial && (initial.OccurSymbol(key) || naturals != nil && naturals.OccurSymbol(key)) {
			continue
		}
		current := context.GetSymbol(key)
		if sanyExploreNull(current) {
			panic(tlc.NewNullPointerException())
		}
		var uid int32
		switch current := current.(type) {
		case *tlc.SymbolNode:
			if current.SemanticBase == nil {
				panic(tlc.NewNullPointerException())
			}
			uid = current.SemanticBase.GetUid()
		case *tlc.OpDefNode:
			uid = current.GetUid()
		case *tlc.ThmOrAssumpDefNode:
			uid = current.GetUid()
		case *tlc.ModuleNode:
			uid = current.GetUid()
		default:
			panic(tlc.NewClassCastException("context entry is not a SymbolNode"))
		}
		entry := "\nContext Entry: " + name.String() + "  " +
			strconv.FormatInt(int64(uid), 10) + " "
		if depth > 1 {
			entry += sanyIndent(2, sanySemanticDepthString(current, depth-1, errors))
		}
		entries = append(entries, entry)
	}
	for left, right := 0, len(entries)-1; left < right; left, right = left+1, right-1 {
		entries[left], entries[right] = entries[right], entries[left]
	}
	return entries
}

// SanyLetInNodeToString adapts LetInNode.toString(int, Errors).
func SanyLetInNodeToString(node *tlc.LetInNode, depth int, errors *SanyErrors) string {
	if node == nil {
		panic(tlc.NewNullPointerException())
	}
	if depth <= 0 {
		return ""
	}
	result := "\n*LetInNode: " + sanySemanticNodeBaseToString(node.SemanticNodeBase, depth)
	for _, entry := range sanyContextEntryStrings(node.Context, 1, false, errors) {
		if entry == "" {
			result += "*** null ***"
		} else {
			result += sanyIndent(2, entry)
		}
	}
	result += sanyIndent(2, "\nBody:"+sanyIndent(2, sanySemanticDepthString(node.Body, depth-1, errors)))
	return result
}

// SanyOpDefNodeToString adapts OpDefNode.toString(int, Errors).
func SanyOpDefNodeToString(node *tlc.OpDefNode, depth int, errors *SanyErrors) string {
	if node == nil {
		panic(tlc.NewNullPointerException())
	}
	if depth <= 0 {
		return ""
	}
	name := node.Name
	if name == nil {
		panic(tlc.NewNullPointerException())
	}
	result := "\n*OpDefNode: " + name.String() + "\n  " +
		sanyOpDefOrDeclNodeToString(node.Symbol, depth) +
		"\n  local: " + strconv.FormatBool(node.IsLocal()) +
		"\n  letInLevel: " + strconv.Itoa(node.GetLetInLevel()) +
		"\n  inRecursive: " + strconv.FormatBool(node.GetInRecursive()) +
		"\n  inRecursiveSection: " + strconv.FormatBool(node.GetInRecursiveSection()) +
		"\n  recursiveSection: " + strconv.Itoa(node.GetRecursiveSection()) +
		"\n  local: " + strconv.FormatBool(node.IsLocal()) + "\n  source: "
	if node.SourceDefinition == nil {
		result += "this"
	} else {
		sourceName := node.SourceDefinition.Name
		if sourceName == nil {
			panic(tlc.NewNullPointerException())
		}
		result += sourceName.String() + " (uid: " +
			strconv.FormatInt(int64(node.SourceDefinition.GetUid()), 10) + ")"
	}
	result += "\n  originallyDefinedInModule: "
	if node.OriginallyDefinedInModule == nil {
		result += "null"
	} else {
		moduleName := node.OriginallyDefinedInModule.GetName()
		if moduleName == nil {
			panic(tlc.NewNullPointerException())
		}
		result += moduleName.String() + " (uid: " +
			strconv.FormatInt(int64(node.OriginallyDefinedInModule.GetUid()), 10) + ")"
	}
	if !sanyExploreNull(node.StepNode) {
		result += "\n  stepNode: " + sanyIndent(4, sanySemanticDepthString(node.StepNode, depth-3, errors))
	}
	parameters := node.GetParams()
	if parameters != nil {
		result += "\n  Formal params: " + strconv.Itoa(len(parameters))
		for _, parameter := range parameters {
			if parameter == nil {
				result += sanyIndent(4, "\nnull")
			} else {
				result += sanyIndent(4, sanySemanticDepthString(parameter, depth-1, errors))
			}
		}
	} else {
		result += sanyIndent(2, "\nFormal params: null")
	}
	if depth > 1 {
		if node.SymbolTablePresent {
			result += "\n  SymbolTable: non-null"
		} else {
			result += "\n  SymbolTable: null"
		}
	}
	body := node.GetBody()
	if !sanyExploreNull(body) {
		result += sanyIndent(2, "\nBody:"+sanyIndent(2, sanySemanticDepthString(body, depth-1, errors)))
	} else {
		result += sanyIndent(2, "\nBody: null")
	}
	labels := node.GetLabelsHT()
	if labels == nil {
		result += "\n  Labels: null"
	} else {
		result += "\n  Labels: "
		for _, label := range labels.Elements() {
			labelName := label.GetName()
			if labelName == nil {
				panic(tlc.NewNullPointerException())
			}
			result += labelName.String() + "  "
		}
	}
	return result
}

func sanyOpApplBoundSymbolCount(node *tlc.OpApplNode) int {
	if node.BdedQuantSymbolLists == nil {
		return 0
	}
	count := 0
	for i, symbols := range node.BdedQuantSymbolLists {
		if node.BdedQuantATuple == nil {
			panic(tlc.NewNullPointerException())
		}
		if i >= len(node.BdedQuantATuple) {
			panic(tlc.NewArrayIndexOutOfBoundsException(i, len(node.BdedQuantATuple)))
		}
		if node.BdedQuantATuple[i] {
			count++
		} else {
			if symbols == nil {
				panic(tlc.NewNullPointerException())
			}
			count += len(symbols)
		}
	}
	return count
}

func sanyOpApplNodeBodyToString(node *tlc.OpApplNode, depth int, errors *SanyErrors) string {
	if depth <= 1 {
		return ""
	}
	result := ""
	if node.Operator == nil {
		result = "\nOperator: null"
	} else {
		name := node.Operator.GetName()
		if name == nil || node.Operator.SemanticBase == nil {
			panic(tlc.NewNullPointerException())
		}
		result = "\nOperator: " + name.String() + "  " +
			strconv.FormatInt(int64(node.Operator.SemanticBase.GetUid()), 10) + "  "
	}
	if len(node.UnbdedQuantSymbols) > 0 {
		result += "\nUnbounded bound symbols:  "
		for _, symbol := range node.UnbdedQuantSymbols {
			result += sanyIndent(2, sanySemanticDepthString(symbol, depth-1, errors))
		}
	}
	if len(node.BdedQuantSymbolLists) > 0 {
		result += "\nBounded bound symbols: " + strconv.Itoa(sanyOpApplBoundSymbolCount(node))
		for i, symbols := range node.BdedQuantSymbolLists {
			if len(symbols) == 0 {
				continue
			}
			for j, symbol := range symbols {
				result += sanyIndent(2, "\n["+strconv.Itoa(i)+","+strconv.Itoa(j)+"]"+
					sanyIndent(2, sanySemanticDepthString(symbol, depth-1, errors)))
			}
		}
	}
	if node.BdedQuantBounds == nil {
		panic(tlc.NewNullPointerException())
	}
	if len(node.BdedQuantBounds) > 0 {
		result += "\nRanges: "
		for _, bound := range node.BdedQuantBounds {
			if sanyExploreNull(bound) {
				result += sanyIndent(2, "null")
			} else {
				result += sanyIndent(2, sanySemanticDepthString(bound, depth-1, errors))
			}
		}
	}
	if len(node.BdedQuantATuple) > 0 {
		result += "\nTupleOrs:   "
		for _, tuple := range node.BdedQuantATuple {
			if tuple {
				result += sanyIndent(2, "\ntrue")
			} else {
				result += sanyIndent(2, "\nfalse")
			}
		}
	}
	if node.Args == nil {
		result += "\nOperands: null"
	} else if len(node.Args) > 0 {
		result += "\nOperands: " + strconv.Itoa(len(node.Args))
		for _, operand := range node.Args {
			if sanyExploreNull(operand) {
				result += sanyIndent(2, "\nnull")
			} else {
				result += sanyIndent(2, sanySemanticDepthString(operand, depth-1, errors))
			}
		}
	}
	return sanyIndent(2, result)
}

// SanyOpApplNodeToString adapts OpApplNode.toString(int, Errors).
func SanyOpApplNodeToString(node *tlc.OpApplNode, depth int, errors *SanyErrors) string {
	if node == nil {
		panic(tlc.NewNullPointerException())
	}
	if depth <= 0 {
		return ""
	}
	operator := node.GetOperator()
	if operator == nil {
		panic(tlc.NewNullPointerException())
	}
	name := "null"
	if operator.GetName() != nil {
		name = operator.GetName().String()
	}
	result := "\n*OpApplNode: " + name + "  " +
		sanySemanticNodeBaseToString(node.SemanticNodeBase, depth+1)
	if errors != nil {
		result += "  errors: non-null"
	} else {
		result += "  errors: null"
	}
	result += sanyOpApplNodeBodyToString(node, depth, errors)
	if node.SubExpressionOf != nil {
		result += sanyIndent(2, "\nsubExpressionOf: "+
			sanyIndent(2, sanySemanticDepthString(node.SubExpressionOf, 1, errors)))
	}
	return result
}

// SanyAtNodeToString adapts AtNode.toString(int, Errors).
func SanyAtNodeToString(node *tlc.AtNode, depth int, errors *SanyErrors) string {
	if node == nil {
		panic(tlc.NewNullPointerException())
	}
	if depth <= 0 {
		return ""
	}
	if node.ExceptRef == nil || node.ExceptComponentRef == nil {
		panic(tlc.NewNullPointerException())
	}
	return "\n*AtNode: " + sanySemanticNodeBaseToString(node.SemanticNodeBase, depth) +
		sanyIndent(2, "\nExceptRef: "+strconv.FormatInt(int64(node.ExceptRef.GetUid()), 10)+
			"\nExceptComponent: "+strconv.FormatInt(int64(node.ExceptComponentRef.GetUid()), 10))
}

// SanyNumeralNodeToString adapts NumeralNode.toString(int, Errors). The source
// accepts but does not inspect the error log.
func SanyNumeralNodeToString(node *tlc.NumeralNode, depth int, errors *SanyErrors) string {
	if node == nil {
		panic(tlc.NewNullPointerException())
	}
	if depth <= 0 {
		return ""
	}
	result := "\n*NumeralNode: " + sanySemanticNodeBaseToString(&node.SemanticNodeBase, depth) +
		" Value: " + strconv.FormatInt(int64(node.Val()), 10)
	if value := node.BigVal(); value != nil {
		result += "; big value: " + value.String()
	}
	return result + "; image: " + node.String()
}

// SanyDecimalNodeToString adapts DecimalNode.toString(int, Errors).
func SanyDecimalNodeToString(node *tlc.DecimalNode, depth int, errors *SanyErrors) string {
	if node == nil {
		panic(tlc.NewNullPointerException())
	}
	if depth <= 0 {
		return ""
	}
	bigValue := "<null>"
	if value := node.BigVal(); value != nil {
		bigValue = value.String()
	}
	return "\n*DecimalNode" + sanySemanticNodeBaseToString(&node.SemanticNodeBase, depth) +
		"Mantissa: " + strconv.FormatInt(node.Mantissa(), 10) +
		"; exponent: " + strconv.FormatInt(int64(node.Exponent()), 10) +
		"; big value: " + bigValue + "\n; image = " + node.String()
}

func sanyStringNodePrintVersion(value string) string {
	units := utf16.Encode([]rune(value))
	result := make([]uint16, 0, len(units))
	appendASCII := func(value string) {
		for i := 0; i < len(value); i++ {
			result = append(result, uint16(value[i]))
		}
	}
	for _, unit := range units {
		switch unit {
		case '"':
			appendASCII(`\"`)
		case '\\':
			appendASCII(`\\`)
		case '\t':
			appendASCII(`\t`)
		case '\n':
			appendASCII(`\n`)
		case '\f':
			appendASCII(`\f`)
		case '\r':
			appendASCII(`\r`)
		default:
			result = append(result, unit)
		}
	}
	return string(utf16.Decode(result))
}

// SanyStringNodeToString adapts StringNode.toString(int, Errors), including
// Java UTF-16 length and source escaping.
func SanyStringNodeToString(node *tlc.StringNode, depth int, errors *SanyErrors) string {
	if node == nil {
		panic(tlc.NewNullPointerException())
	}
	if depth <= 0 {
		return ""
	}
	if node.Value == nil || node.Value.Val == nil {
		panic(tlc.NewNullPointerException())
	}
	text := node.Value.Val.String()
	return "\n*StringNode: " + sanySemanticNodeBaseToString(&node.SemanticNodeBase, depth) +
		"Value: '" + sanyStringNodePrintVersion(text) + "' Length: " +
		strconv.Itoa(len(utf16.Encode([]rune(text))))
}
