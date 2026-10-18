package tlago

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/glycerine/tlago/tlc"
)

// LevelNode's mutable collections belong to each canonical node. The native
// analyzer's integer parameter IDs are not substitutes for SymbolNode keys.
type sanyLevelData struct {
	levelCorrect              bool
	level                     *tlaLevel
	levelParams               *sanyLevelSet[sanySemSymbol]
	levelConstraints          *sanySetOfLevelConstraints
	argLevelConstraints       *sanySetOfArgLevelConstraints
	argLevelParams            *sanyLevelSet[*sanyArgLevelParam]
	allParams                 *sanyLevelSet[sanySemSymbol]
	nonLeibnizParams          *sanyLevelSet[sanySemSymbol]
	levelChecked              *int32
	runtimeSymbol             func(sanySemSymbol) *tlc.SymbolNode
	runtimeGraph              func(sanySemanticGraphNode) tlc.SemanticNode
	runtimeParameters         func([]*sanyFormalParamNode) []*tlc.SymbolNode
	runtimeLabels             func(*sanyLabelTable) *tlc.SemanticLabelTable
	runtimeInstallLabels      func(*sanyLabelTable, *tlc.SemanticLabelTable)
	runtimeCheck              func(int32, *Diagnostics) bool
	runtimeCheckSubnodes      func(int32, []tlc.SemanticNode, *Diagnostics) bool
	runtimeSource             sanyCanonicalLevelNode
	runtimeString             func() string
	publicLevelParams         *SanyLevelSymbolSet
	publicAllParams           *SanyLevelSymbolSet
	publicNonLeibniz          *SanyLevelSymbolSet
	publicArgLevelParams      *SanyArgLevelParamSet
	publicLevelConstraints    *SanySetOfLevelConstraints
	publicArgLevelConstraints *SanySetOfArgLevelConstraints
}

// RuntimeAllParams exposes the live canonical set to TLC's OpApplNode value
// renderer.  OpApplNode.toString(IValue) reads this field directly in Java, so
// structural recomputation would miss public mutations made through getAllParams.
func (n *sanyLevelData) RuntimeAllParams() []*tlc.SymbolNode {
	if n == nil || n.allParams == nil {
		panic(tlc.NewNullPointerException())
	}
	return n.runtimeSymbols(n.allParams)
}

func (n *sanyLevelData) RuntimeCheckedLevel() int {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	if n.runtimeSource == nil || reflect.ValueOf(n.runtimeSource).IsNil() {
		return int(n.getLevel())
	}
	return int(n.runtimeSource.getLevel())
}

func (n *sanyLevelData) RuntimeSymbolArity() (int, bool) {
	if n == nil || n.runtimeSource == nil || reflect.ValueOf(n.runtimeSource).IsNil() {
		return 0, false
	}
	source, ok := n.runtimeSource.(sanySemSymbol)
	if !ok {
		return 0, false
	}
	return source.semArity(), true
}

func (n *sanyLevelData) RuntimeSetSymbolArity(arity int) bool {
	if n == nil || n.runtimeSource == nil || reflect.ValueOf(n.runtimeSource).IsNil() {
		return false
	}
	source, ok := n.runtimeSource.(sanySemSymbol)
	if !ok {
		return false
	}
	source.semBase().arity = int(int32(arity))
	return true
}

func (n *sanyLevelData) RuntimeSymbolLocal() (bool, bool) {
	if n == nil || n.runtimeSource == nil || reflect.ValueOf(n.runtimeSource).IsNil() {
		return false, false
	}
	source, ok := n.runtimeSource.(sanySemSymbol)
	if !ok {
		return false, false
	}
	return source.semLocal(), true
}

func (n *sanyLevelData) RuntimeSetSymbolLocal(local bool) bool {
	if n == nil || n.runtimeSource == nil || reflect.ValueOf(n.runtimeSource).IsNil() {
		return false
	}
	source, ok := n.runtimeSource.(sanySemSymbol)
	if !ok {
		return false
	}
	source.semBase().local = local
	return true
}

func (n *sanyLevelData) RuntimeTheoremDefinitionFlags() (bool, bool, bool) {
	if n == nil || n.runtimeSource == nil || reflect.ValueOf(n.runtimeSource).IsNil() {
		return false, false, false
	}
	source, ok := n.runtimeSource.(*sanySemThmOrAssumpDefNode)
	if !ok {
		return false, false, false
	}
	return source.theorem, source.suffices, true
}

func (n *sanyLevelData) RuntimeModuleInstantiated() (bool, bool) {
	if n == nil || n.runtimeSource == nil || reflect.ValueOf(n.runtimeSource).IsNil() {
		return false, false
	}
	source, ok := n.runtimeSource.(*sanySemModuleNode)
	if !ok {
		return false, false
	}
	return source.isInstantiated, true
}

func (n *sanyLevelData) RuntimeSetModuleInstantiated(instantiated bool) bool {
	if n == nil || n.runtimeSource == nil || reflect.ValueOf(n.runtimeSource).IsNil() {
		return false
	}
	source, ok := n.runtimeSource.(*sanySemModuleNode)
	if !ok {
		return false
	}
	source.isInstantiated = instantiated
	return true
}

func (n *sanyLevelData) RuntimeInstanceState() (bool, *tlc.UniqueString, bool) {
	if n == nil || n.runtimeSource == nil || reflect.ValueOf(n.runtimeSource).IsNil() {
		return false, nil, false
	}
	source, ok := n.runtimeSource.(*sanySemInstanceNode)
	if !ok {
		return false, nil, false
	}
	return source.local, source.stepName, true
}

func (n *sanyLevelData) RuntimeSetInstanceStepName(name *tlc.UniqueString) bool {
	if n == nil || n.runtimeSource == nil || reflect.ValueOf(n.runtimeSource).IsNil() {
		return false
	}
	source, ok := n.runtimeSource.(*sanySemInstanceNode)
	if !ok {
		return false
	}
	source.stepName = name
	return true
}

func (n *sanyLevelData) RuntimeUseOrHideStepName() (*tlc.UniqueString, bool) {
	if n == nil || n.runtimeSource == nil || reflect.ValueOf(n.runtimeSource).IsNil() {
		return nil, false
	}
	source, ok := n.runtimeSource.(*sanySemUseOrHideNode)
	if !ok {
		return nil, false
	}
	return source.stepName, true
}

func (n *sanyLevelData) RuntimeSetUseOrHideStepName(name *tlc.UniqueString) bool {
	if n == nil || n.runtimeSource == nil || reflect.ValueOf(n.runtimeSource).IsNil() {
		return false
	}
	source, ok := n.runtimeSource.(*sanySemUseOrHideNode)
	if !ok {
		return false
	}
	source.stepName = name
	return true
}

func (n *sanyLevelData) RuntimeAssumeProveFlags() (bool, bool, bool) {
	if n == nil || n.runtimeSource == nil || reflect.ValueOf(n.runtimeSource).IsNil() {
		return false, false, false
	}
	source, ok := n.runtimeSource.(*sanySemAssumeProveNode)
	if !ok {
		return false, false, false
	}
	return source.suffices, source.isBoxAssumeProve, true
}

func (n *sanyLevelData) RuntimeOpDefBody() (tlc.SemanticNode, bool) {
	if n == nil || n.runtimeSource == nil || reflect.ValueOf(n.runtimeSource).IsNil() || n.runtimeGraph == nil {
		return nil, false
	}
	source, ok := n.runtimeSource.(*sanySemOpDefNode)
	if !ok {
		return nil, false
	}
	return n.runtimeGraph(source.body), true
}

func (n *sanyLevelData) RuntimeSetOpDefBody(body tlc.SemanticNode) bool {
	if n == nil || n.runtimeSource == nil || reflect.ValueOf(n.runtimeSource).IsNil() {
		return false
	}
	source, ok := n.runtimeSource.(*sanySemOpDefNode)
	if !ok {
		return false
	}
	if sanyExploreNull(body) {
		source.body = nil
		return true
	}
	data, ok := body.(sanyRuntimeLevelDataCarrier)
	if !ok {
		return false
	}
	canonical, ok := data.GetCanonicalLevelData().(*sanyLevelData)
	if !ok || canonical == nil {
		return false
	}
	graph, ok := canonical.runtimeSource.(sanySemanticGraphNode)
	if !ok || sanyExploreNull(graph) {
		return false
	}
	source.body = graph
	return true
}

func (n *sanyLevelData) RuntimeOpDefParams() ([]*tlc.SymbolNode, bool) {
	if n == nil || n.runtimeSource == nil || reflect.ValueOf(n.runtimeSource).IsNil() || n.runtimeParameters == nil {
		return nil, false
	}
	source, ok := n.runtimeSource.(*sanySemOpDefNode)
	if !ok {
		return nil, false
	}
	return n.runtimeParameters(source.formalNodes), true
}

func (n *sanyLevelData) RuntimeSetOpDefParams(params []*tlc.SymbolNode) bool {
	if n == nil || n.runtimeSource == nil || reflect.ValueOf(n.runtimeSource).IsNil() {
		return false
	}
	source, ok := n.runtimeSource.(*sanySemOpDefNode)
	if !ok {
		return false
	}
	if params == nil {
		source.formalNodes = nil
		return true
	}
	formals := make([]*sanyFormalParamNode, len(params))
	for i, param := range params {
		if param == nil || param.SemanticBase == nil {
			return false
		}
		data, ok := param.SemanticBase.CanonicalLevelData.(*sanyLevelData)
		if !ok || data == nil {
			return false
		}
		formal, ok := data.runtimeSource.(*sanyFormalParamNode)
		if !ok || formal == nil {
			return false
		}
		formals[i] = formal
	}
	source.formalNodes = formals
	return true
}

func (n *sanyLevelData) RuntimeTheoremDefinitionBody() (tlc.SemanticNode, bool) {
	if n == nil || n.runtimeSource == nil || reflect.ValueOf(n.runtimeSource).IsNil() || n.runtimeGraph == nil {
		return nil, false
	}
	source, ok := n.runtimeSource.(*sanySemThmOrAssumpDefNode)
	if !ok {
		return nil, false
	}
	return n.runtimeGraph(source.body), true
}

func (n *sanyLevelData) RuntimeTheoremDefinitionParams() ([]*tlc.SymbolNode, bool) {
	if n == nil || n.runtimeSource == nil || reflect.ValueOf(n.runtimeSource).IsNil() || n.runtimeParameters == nil {
		return nil, false
	}
	source, ok := n.runtimeSource.(*sanySemThmOrAssumpDefNode)
	if !ok {
		return nil, false
	}
	return n.runtimeParameters(source.formalNodes), true
}

func (n *sanyLevelData) runtimeSubstitutionNode() (*sanySemSubstitutionNode, bool) {
	if n == nil || n.runtimeSource == nil || reflect.ValueOf(n.runtimeSource).IsNil() {
		return nil, false
	}
	switch source := n.runtimeSource.(type) {
	case *sanySemSubstInNode:
		return &source.sanySemSubstitutionNode, true
	case *sanySemAPSubstInNode:
		return &source.sanySemSubstitutionNode, true
	default:
		return nil, false
	}
}

func (n *sanyLevelData) RuntimeSubstitutionBody() (tlc.SemanticNode, bool) {
	source, ok := n.runtimeSubstitutionNode()
	if !ok || n.runtimeGraph == nil {
		return nil, false
	}
	return n.runtimeGraph(source.body), true
}

func (n *sanyLevelData) RuntimeSetSubstitutionBody(body tlc.SemanticNode) bool {
	source, ok := n.runtimeSubstitutionNode()
	if !ok {
		return false
	}
	if sanyExploreNull(body) {
		source.body = nil
		return true
	}
	carrier, ok := body.(sanyRuntimeLevelDataCarrier)
	if !ok {
		return false
	}
	canonical, ok := carrier.GetCanonicalLevelData().(*sanyLevelData)
	if !ok || canonical == nil {
		return false
	}
	graph, ok := canonical.runtimeSource.(sanySemanticGraphNode)
	if !ok || sanyExploreNull(graph) {
		return false
	}
	source.body = graph
	return true
}

func (n *sanyLevelData) runtimeLabelTable() (*sanyLabelTable, bool) {
	if n == nil || n.runtimeSource == nil || reflect.ValueOf(n.runtimeSource).IsNil() {
		return nil, false
	}
	switch source := n.runtimeSource.(type) {
	case *sanySemOpDefNode:
		return source.labels, true
	case *sanySemThmOrAssumpDefNode:
		return source.labels, true
	case *sanySemLabelNode:
		return source.labels, true
	default:
		return nil, false
	}
}

func (n *sanyLevelData) RuntimeLabels() (*tlc.SemanticLabelTable, bool) {
	table, ok := n.runtimeLabelTable()
	if !ok || n.runtimeLabels == nil {
		return nil, false
	}
	return n.runtimeLabels(table), true
}

func (n *sanyLevelData) RuntimeSetLabels(labels *tlc.SemanticLabelTable) bool {
	if _, ok := n.runtimeLabelTable(); !ok {
		return false
	}
	table, ok := sanyLabelTableFromPublic(labels)
	if !ok {
		return false
	}
	switch source := n.runtimeSource.(type) {
	case *sanySemOpDefNode:
		source.labels = table
	case *sanySemThmOrAssumpDefNode:
		source.labels = table
	case *sanySemLabelNode:
		source.labels = table
	default:
		return false
	}
	if table != nil && labels != nil && n.runtimeInstallLabels != nil {
		n.runtimeInstallLabels(table, labels)
	}
	return true
}

func (n *sanyLevelData) RuntimeOpDefRecursion() (bool, bool, int, int, bool) {
	if n == nil || n.runtimeSource == nil || reflect.ValueOf(n.runtimeSource).IsNil() {
		return false, false, 0, 0, false
	}
	source, ok := n.runtimeSource.(*sanySemOpDefNode)
	if !ok {
		return false, false, 0, 0, false
	}
	return source.inRecursive, source.inRecursiveSection, source.recursiveSection, source.letInLevel, true
}

func (n *sanyLevelData) RuntimeSetOpDefInRecursive(value bool) bool {
	if n == nil || n.runtimeSource == nil || reflect.ValueOf(n.runtimeSource).IsNil() {
		return false
	}
	source, ok := n.runtimeSource.(*sanySemOpDefNode)
	if !ok {
		return false
	}
	source.inRecursive = value
	return true
}

// RuntimeCheckedLevelParams and RuntimeCheckedAllParams retain LevelNode's
// virtual getter dispatch and its pre-levelCheck WrongInvocationException.
func (n *sanyLevelData) RuntimeCheckedLevelParams() []*tlc.SymbolNode {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	if n.runtimeSource == nil || reflect.ValueOf(n.runtimeSource).IsNil() {
		return n.runtimeSymbols(n.getLevelParams())
	}
	return n.runtimeSymbols(n.runtimeSource.getLevelParams())
}

func (n *sanyLevelData) RuntimeCheckedAllParams() []*tlc.SymbolNode {
	if n == nil || n.runtimeSource == nil || reflect.ValueOf(n.runtimeSource).IsNil() {
		panic(tlc.NewNullPointerException())
	}
	return n.runtimeSymbols(n.runtimeSource.getAllParams())
}

// RuntimeCheckedMaxLevel and the related accessors keep public semantic views
// attached to the arrays installed by the latest source level-check iteration.
// OpDefNode and ThmOrAssumpDefNode replace these arrays instead of mutating the
// previous ones in place.
func (n *sanyLevelData) RuntimeCheckedMaxLevel(index int) int {
	source, ok := n.runtimeSource.(interface{ getMaxLevel(int) tlaLevel })
	if !ok {
		panic(tlc.NewClassCastException("level node has no maximum-level data"))
	}
	return int(source.getMaxLevel(index))
}

func (n *sanyLevelData) RuntimeCheckedWeight(index int) int {
	source, ok := n.runtimeSource.(interface{ getWeight(int) int })
	if !ok {
		panic(tlc.NewClassCastException("level node has no argument-weight data"))
	}
	return source.getWeight(index)
}

func (n *sanyLevelData) RuntimeCheckedMinMaxLevel(i, j int) int {
	source, ok := n.runtimeSource.(interface{ getMinMaxLevel(int, int) tlaLevel })
	if !ok {
		panic(tlc.NewClassCastException("level node has no minimum-maximum-level data"))
	}
	return int(source.getMinMaxLevel(i, j))
}

func (n *sanyLevelData) RuntimeCheckedOpLevelCond(i, j, k int) bool {
	source, ok := n.runtimeSource.(interface{ getOpLevelCond(int, int, int) bool })
	if !ok {
		panic(tlc.NewClassCastException("level node has no operator-level condition data"))
	}
	return source.getOpLevelCond(i, j, k)
}

func (n *sanyLevelData) RuntimeArgMaxLevels() []int {
	source, ok := n.runtimeSource.(*sanySemOpDefNode)
	if !ok {
		panic(tlc.NewClassCastException("level node has no argument maximum-level array"))
	}
	if source.argMaxLevels == nil {
		return nil
	}
	result := make([]int, len(source.argMaxLevels))
	for i, level := range source.argMaxLevels {
		result[i] = int(level)
	}
	return result
}

func (n *sanyLevelData) RuntimeArgWeights() []int {
	source, ok := n.runtimeSource.(*sanySemOpDefNode)
	if !ok {
		panic(tlc.NewClassCastException("level node has no argument weight array"))
	}
	if source.argWeights == nil {
		return nil
	}
	result := make([]int, len(source.argWeights))
	copy(result, source.argWeights)
	return result
}

func (n *sanyLevelData) RuntimeIsLeibnizArg() []bool {
	source, ok := n.runtimeSource.(interface{ getIsLeibnizArg() []bool })
	if !ok {
		panic(tlc.NewClassCastException("level node has no Leibniz argument data"))
	}
	return source.getIsLeibnizArg()
}

func (n *sanyLevelData) RuntimeIsLeibniz() bool {
	source, ok := n.runtimeSource.(interface{ getIsLeibniz() bool })
	if !ok {
		panic(tlc.NewClassCastException("level node has no Leibniz data"))
	}
	return source.getIsLeibniz()
}

func (n *sanyLevelData) runtimeSymbols(set *sanyLevelSet[sanySemSymbol]) []*tlc.SymbolNode {
	if n == nil || set == nil {
		panic(tlc.NewNullPointerException())
	}
	result := make([]*tlc.SymbolNode, 0, set.len())
	for symbol := range set.all() {
		result = append(result, n.publicRuntimeSymbol(symbol))
	}
	return result
}

func newSanyLevelData(base *tlc.SemanticNodeBase) *sanyLevelData {
	return &sanyLevelData{
		level:               (*tlaLevel)(&base.LevelValue),
		levelChecked:        &base.LevelChecked,
		levelCorrect:        true,
		levelParams:         newSanyLevelSymbolSet(),
		levelConstraints:    newSanySetOfLevelConstraints(),
		argLevelConstraints: newSanySetOfArgLevelConstraints(),
		argLevelParams:      newSanyArgLevelSet(),
		allParams:           newSanyLevelSymbolSet(),
		nonLeibnizParams:    newSanyLevelSymbolSet(),
	}
}

// Canonical SANY classes and TLC literal views expose the same owned metadata.
// Literal views preserve the underlying concrete body node and semantic identity.
type sanyLevelAccess interface {
	getLevel() tlaLevel
	getLevelParams() *sanyLevelSet[sanySemSymbol]
	getAllParams() *sanyLevelSet[sanySemSymbol]
	getNonLeibnizParams() *sanyLevelSet[sanySemSymbol]
	getLevelConstraints() *sanySetOfLevelConstraints
	getArgLevelConstraints() *sanySetOfArgLevelConstraints
	getArgLevelParams() *sanyLevelSet[*sanyArgLevelParam]
}
type sanyCanonicalLevelNode interface {
	sanyLevelAccess
	getKind() sanySemKind
	getLevelData() *sanyLevelData
	levelCheck(int32, *Diagnostics) bool
}

func (n *sanyLevelData) getLevelData() *sanyLevelData {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	return n
}
func sanyLevelCheckNext(node sanyCanonicalLevelNode, errors *Diagnostics) bool {
	if node == nil || reflect.ValueOf(node).IsNil() {
		panic(tlc.NewNullPointerException())
	}
	return node.levelCheck(*node.getLevelData().levelChecked+1, errors)
}
func (n *sanySemanticNode) levelCheck(iter int32, errors *Diagnostics) bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	kind := int(n.getKind())
	if kind < 0 || kind >= len(SanySemanticKindNames) {
		panic(tlc.NewArrayIndexOutOfBoundsException(kind, len(SanySemanticKindNames)))
	}
	panic(tlc.NewWrongInvocationException("Level checking of " + SanySemanticKindNames[kind] + " node not implemented."))
}
func (n *sanySemanticNode) levelCheckSubnodes(iter int32, sub []sanyCanonicalLevelNode, errors *Diagnostics) bool {
	return n.levelCheckSubnodesAccess(iter, sub == nil, len(sub), func(i int) sanyCanonicalLevelNode { return sub[i] }, errors)
}

// Resolve graph children only when reached, preserving prior writes on failure.
func (n *sanySemanticNode) levelCheckGraphSubnodes(iter int32, sub []sanySemanticGraphNode, errors *Diagnostics) bool {
	return n.levelCheckSubnodesAccess(iter, sub == nil, len(sub), func(i int) sanyCanonicalLevelNode { return sanyRequireCanonicalLevelNode(sub[i]) }, errors)
}
func (n *sanySemanticNode) levelCheckSubnodesAccess(iter int32, nilArray bool, count int, childAt func(int) sanyCanonicalLevelNode, errors *Diagnostics) bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	if *n.levelChecked >= iter {
		return n.levelCorrect
	}
	*n.levelChecked = iter
	if nilArray {
		panic(tlc.NewNullPointerException())
	}
	for i := 0; i < count; i++ {
		child := childAt(i)
		if child == nil || reflect.ValueOf(child).IsNil() {
			panic(tlc.NewNullPointerException())
		}
		if child.getKind() != sanyModuleKind && child.getKind() != sanyModuleInstanceKind {
			// Evaluate the child even after an earlier child has returned false.
			correct := child.levelCheck(iter, errors)
			n.levelCorrect = correct && n.levelCorrect
			*n.level = max(*n.level, child.getLevel())
			n.levelParams.addAll(child.getLevelParams())
			constraints := child.getLevelConstraints()
			var constraintMap *tlc.JavaSemanticMap[sanySemSymbol, *int32]
			if constraints != nil {
				constraintMap = constraints.entries
			}
			n.levelConstraints.putAll(constraintMap)
			argConstraints := child.getArgLevelConstraints()
			var argConstraintMap *tlc.JavaSemanticMap[*sanyParamAndPosition, *int32]
			if argConstraints != nil {
				argConstraintMap = argConstraints.entries
			}
			n.argLevelConstraints.putAll(argConstraintMap)
			n.argLevelParams.addAll(child.getArgLevelParams())
			n.allParams.addAll(child.getAllParams())
			n.nonLeibnizParams.addAll(child.getNonLeibnizParams())
		}
	}
	return n.levelCorrect
}

func (n *sanyLevelData) requireChecked(message string) {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	if *n.levelChecked == 0 {
		panic(tlc.NewWrongInvocationException(message))
	}
}
func (n *sanyLevelData) getLevel() tlaLevel {
	n.requireChecked("getLevel called before levelCheck")
	return *n.level
}
func (n *sanyLevelData) getLevelParams() *sanyLevelSet[sanySemSymbol] {
	n.requireChecked("getLevelParams called before levelCheck")
	return n.levelParams
}
func (n *sanyLevelData) getAllParams() *sanyLevelSet[sanySemSymbol] {
	n.requireChecked("getAllParams called before levelCheck")
	return n.allParams
}
func (n *sanyLevelData) getNonLeibnizParams() *sanyLevelSet[sanySemSymbol] {
	// The source intentionally has the same error message as getAllParams.
	n.requireChecked("getAllParams called before levelCheck")
	return n.nonLeibnizParams
}
func (n *sanyLevelData) getLevelConstraints() *sanySetOfLevelConstraints {
	n.requireChecked("getLevelConstraints called before levelCheck")
	return n.levelConstraints
}
func (n *sanyLevelData) getArgLevelConstraints() *sanySetOfArgLevelConstraints {
	n.requireChecked("getArgLevelConstraints called before levelCheck")
	return n.argLevelConstraints
}
func (n *sanyLevelData) getArgLevelParams() *sanyLevelSet[*sanyArgLevelParam] {
	n.requireChecked("getArgLevelParams called before levelCheck")
	return n.argLevelParams
}
func sanyDefaultLevelDataToString(n sanyLevelAccess) string {
	if n == nil || reflect.ValueOf(n).IsNil() {
		panic(tlc.NewNullPointerException())
	}
	level := n.getLevel()
	params := n.getLevelParams()
	constraints := n.getLevelConstraints()
	argConstraints := n.getArgLevelConstraints()
	lp, lc, ac := "null", "null", "null"
	if params != nil {
		lp = sanyLevelSymbolSetString(params)
	}
	if constraints != nil {
		lc = constraints.String()
	}
	if argConstraints != nil {
		ac = argConstraints.String()
	}
	return "Level:" + strconv.Itoa(int(level)) + "\n" +
		"LevelParams: " + lp + "\n" +
		"LevelConstraints: " + lc + "\n" +
		"ArgLevelConstraints: " + ac + "\n" +
		"ArgLevelParams: " + sanyALPHashSetToString(n.getArgLevelParams()) + "\n" +
		"AllParams: " + sanyHashSetToString(n.getAllParams()) + "\n" +
		"NonLeibnizParams: " + sanyHashSetToString(n.getNonLeibnizParams())
}
func sanyLevelDataToString(n sanyCanonicalLevelNode) string {
	if n == nil || reflect.ValueOf(n).IsNil() {
		panic(tlc.NewNullPointerException())
	}
	if printer, ok := n.(interface{ levelDataToString() string }); ok {
		return printer.levelDataToString()
	}
	n.getLevelData().requireChecked("levelDataToString called before levelCheck")
	return sanyDefaultLevelDataToString(n)
}

type sanyRuntimeLevelDataCarrier interface {
	GetCanonicalLevelData() any
}

func sanyRuntimeLevelData(node tlc.SemanticNode) *sanyLevelData {
	if sanyExploreNull(node) {
		panic(tlc.NewNullPointerException())
	}
	carrier, ok := node.(sanyRuntimeLevelDataCarrier)
	if !ok {
		panic(tlc.NewClassCastException("semantic node does not carry SANY level data"))
	}
	data, ok := carrier.GetCanonicalLevelData().(*sanyLevelData)
	if !ok || data == nil {
		panic(tlc.NewClassCastException("semantic node does not carry SANY level data"))
	}
	return data
}

func sanyRuntimeLevelCheck(node tlc.SemanticNode, iteration int32, errors *SanyErrors) bool {
	data := sanyRuntimeLevelData(node)
	if data.runtimeCheck == nil {
		panic(tlc.NewUnsupportedOperationException("canonical SANY level checker has not been installed"))
	}
	var diagnostics *Diagnostics
	if errors != nil {
		diagnostics = &errors.diagnostics
	}
	return data.runtimeCheck(iteration, diagnostics)
}

type sanyCanonicalLevelSubnodeChecker interface {
	levelCheckSubnodesAccess(int32, bool, int, func(int) sanyCanonicalLevelNode, *Diagnostics) bool
}

func bindSanyRuntimeLevelNode(data *sanyLevelData, source sanyCanonicalLevelNode, runtimeSymbol func(sanySemSymbol) *tlc.SymbolNode) {
	data.runtimeSource = source
	data.runtimeSymbol = runtimeSymbol
	data.runtimeCheck = source.levelCheck
	if checker, ok := source.(sanyCanonicalLevelSubnodeChecker); ok {
		data.runtimeCheckSubnodes = func(iteration int32, children []tlc.SemanticNode, errors *Diagnostics) bool {
			return checker.levelCheckSubnodesAccess(iteration, children == nil, len(children), func(index int) sanyCanonicalLevelNode {
				child := sanyRuntimeLevelData(children[index]).runtimeSource
				if child == nil || reflect.ValueOf(child).IsNil() {
					panic(tlc.NewUnsupportedOperationException("canonical SANY level node has not been installed"))
				}
				return child
			}, errors)
		}
	}
}

// SanyLevelNodeLevelCheck adapts LevelNode.levelCheck(int, Errors).
func SanyLevelNodeLevelCheck(node tlc.SemanticNode, iteration int, errors *SanyErrors) bool {
	return sanyRuntimeLevelCheck(node, int32(iteration), errors)
}

// SanyLevelNodeLevelCheckNext adapts LevelNode.levelCheck(Errors).
func SanyLevelNodeLevelCheckNext(node tlc.SemanticNode, errors *SanyErrors) bool {
	data := sanyRuntimeLevelData(node)
	return sanyRuntimeLevelCheck(node, *data.levelChecked+1, errors)
}

// SanyLevelNodeLevelCheckSubnodes adapts
// LevelNode.levelCheckSubnodes(int, LevelNode[], Errors).
func SanyLevelNodeLevelCheckSubnodes(node tlc.SemanticNode, iteration int, children []tlc.SemanticNode, errors *SanyErrors) bool {
	data := sanyRuntimeLevelData(node)
	if data.runtimeCheckSubnodes == nil {
		panic(tlc.NewUnsupportedOperationException("canonical SANY subnode level checker has not been installed"))
	}
	var diagnostics *Diagnostics
	if errors != nil {
		diagnostics = &errors.diagnostics
	}
	return data.runtimeCheckSubnodes(int32(iteration), children, diagnostics)
}

// SanyLevelNodeGetLevel adapts LevelNode.getLevel().
func SanyLevelNodeGetLevel(node tlc.SemanticNode) int {
	return int(sanyRuntimeLevelData(node).getLevel())
}

// SanyLevelSymbolSet is a live adapter over LevelNode's Java HashSet fields.
type SanyLevelSymbolSet struct {
	owner *sanyLevelData
	set   *sanyLevelSet[sanySemSymbol]
}

func (s *SanyLevelSymbolSet) require() *SanyLevelSymbolSet {
	if s == nil || s.owner == nil || s.set == nil {
		panic(tlc.NewNullPointerException())
	}
	return s
}

func sanyRuntimeCanonicalSymbol(symbol tlc.SemanticNode) sanySemSymbol {
	source := sanyRuntimeLevelData(symbol).runtimeSource
	result, ok := source.(sanySemSymbol)
	if !ok || sanyLevelSymbolReference(result) == nil {
		panic(tlc.NewClassCastException("level parameter is not a canonical SymbolNode"))
	}
	return result
}

func (s *SanyLevelSymbolSet) Add(symbol tlc.SemanticNode) bool {
	return s.require().set.add(sanyRuntimeCanonicalSymbol(symbol))
}

func (s *SanyLevelSymbolSet) Contains(symbol tlc.SemanticNode) bool {
	return s.require().set.contains(sanyRuntimeCanonicalSymbol(symbol))
}

func (s *SanyLevelSymbolSet) Remove(symbol tlc.SemanticNode) bool {
	return s.require().set.remove(sanyRuntimeCanonicalSymbol(symbol))
}

func (s *SanyLevelSymbolSet) Len() int { return s.require().set.len() }
func (s *SanyLevelSymbolSet) Clear()   { s.require().set.clear() }

func (s *SanyLevelSymbolSet) Values() []tlc.SemanticNode {
	s.require()
	result := make([]tlc.SemanticNode, 0, s.set.len())
	for symbol := range s.set.all() {
		result = append(result, s.owner.publicRuntimeSymbol(symbol))
	}
	return result
}

func (s *SanyLevelSymbolSet) String() string {
	return sanyLevelSymbolSetString(s.require().set)
}

func sanyPublicLevelSymbolSet(data *sanyLevelData, set *sanyLevelSet[sanySemSymbol], cached **SanyLevelSymbolSet) *SanyLevelSymbolSet {
	if *cached == nil || (*cached).set != set {
		*cached = &SanyLevelSymbolSet{owner: data, set: set}
	}
	return *cached
}

func SanyLevelNodeGetLevelParams(node tlc.SemanticNode) *SanyLevelSymbolSet {
	data := sanyRuntimeLevelData(node)
	return sanyPublicLevelSymbolSet(data, data.getLevelParams(), &data.publicLevelParams)
}

func SanyLevelNodeGetAllParams(node tlc.SemanticNode) *SanyLevelSymbolSet {
	data := sanyRuntimeLevelData(node)
	return sanyPublicLevelSymbolSet(data, data.getAllParams(), &data.publicAllParams)
}

func SanyLevelNodeGetNonLeibnizParams(node tlc.SemanticNode) *SanyLevelSymbolSet {
	data := sanyRuntimeLevelData(node)
	return sanyPublicLevelSymbolSet(data, data.getNonLeibnizParams(), &data.publicNonLeibniz)
}

// SanyLevelNodeGetLevelConstraints adapts LevelNode.getLevelConstraints().
func SanyLevelNodeGetLevelConstraints(node tlc.SemanticNode) *SanySetOfLevelConstraints {
	data := sanyRuntimeLevelData(node)
	constraints := data.getLevelConstraints()
	if constraints == nil {
		return nil
	}
	if data.publicLevelConstraints == nil || data.publicLevelConstraints.canonical != constraints {
		data.publicLevelConstraints = &SanySetOfLevelConstraints{owner: data, canonical: constraints}
	}
	return data.publicLevelConstraints
}

// SanyLevelNodeGetArgLevelConstraints adapts LevelNode.getArgLevelConstraints().
func SanyLevelNodeGetArgLevelConstraints(node tlc.SemanticNode) *SanySetOfArgLevelConstraints {
	data := sanyRuntimeLevelData(node)
	constraints := data.getArgLevelConstraints()
	if constraints == nil {
		return nil
	}
	if data.publicArgLevelConstraints == nil || data.publicArgLevelConstraints.canonical != constraints {
		data.publicArgLevelConstraints = &SanySetOfArgLevelConstraints{owner: data, canonical: constraints}
	}
	return data.publicArgLevelConstraints
}

// SanyArgLevelParamSet is a live adapter over LevelNode.argLevelParams.
type SanyArgLevelParamSet struct {
	owner *sanyLevelData
	set   *sanyLevelSet[*sanyArgLevelParam]
}

func (s *SanyArgLevelParamSet) require() *SanyArgLevelParamSet {
	if s == nil || s.owner == nil || s.set == nil {
		panic(tlc.NewNullPointerException())
	}
	return s
}

func sanyCanonicalArgLevelParam(param *SanyArgLevelParam) *sanyArgLevelParam {
	if param == nil {
		panic(tlc.NewNullPointerException())
	}
	return newSanyArgLevelParam(sanyRuntimeCanonicalSymbol(param.op), param.index, sanyRuntimeCanonicalSymbol(param.param))
}

func (s *SanyArgLevelParamSet) Add(param *SanyArgLevelParam) bool {
	return s.require().set.add(sanyCanonicalArgLevelParam(param))
}

func (s *SanyArgLevelParamSet) Contains(param *SanyArgLevelParam) bool {
	return s.require().set.contains(sanyCanonicalArgLevelParam(param))
}

func (s *SanyArgLevelParamSet) Remove(param *SanyArgLevelParam) bool {
	return s.require().set.remove(sanyCanonicalArgLevelParam(param))
}

func (s *SanyArgLevelParamSet) Len() int { return s.require().set.len() }
func (s *SanyArgLevelParamSet) Clear()   { s.require().set.clear() }

func (s *SanyArgLevelParamSet) Values() []*SanyArgLevelParam {
	s.require()
	result := make([]*SanyArgLevelParam, 0, s.set.len())
	for param := range s.set.all() {
		if param == nil {
			panic(tlc.NewNullPointerException())
		}
		result = append(result, NewSanyArgLevelParam(s.owner.publicRuntimeSymbol(param.op), int(param.i), s.owner.publicRuntimeSymbol(param.param)))
	}
	return result
}

func (s *SanyArgLevelParamSet) String() string {
	s.require()
	var out strings.Builder
	out.WriteByte('[')
	first := true
	for _, param := range s.Values() {
		if !first {
			out.WriteString(", ")
		}
		first = false
		out.WriteString(param.String())
	}
	out.WriteByte(']')
	return out.String()
}

func SanyLevelNodeGetArgLevelParams(node tlc.SemanticNode) *SanyArgLevelParamSet {
	data := sanyRuntimeLevelData(node)
	set := data.getArgLevelParams()
	if data.publicArgLevelParams == nil || data.publicArgLevelParams.set != set {
		data.publicArgLevelParams = &SanyArgLevelParamSet{owner: data, set: set}
	}
	return data.publicArgLevelParams
}

// SanyLevelNodeHashSetToString adapts LevelNode.HashSetToString.
func SanyLevelNodeHashSetToString(set *SanyLevelSymbolSet) string {
	return sanyHashSetToString(set.require().set)
}

// SanyLevelNodeALPHashSetToString adapts LevelNode.ALPHashSetToString.
func SanyLevelNodeALPHashSetToString(set *SanyArgLevelParamSet) string {
	return sanyALPHashSetToString(set.require().set)
}

// SanyLevelNodeDefaultLevelDataToString adapts
// LevelNode.defaultLevelDataToString() for a canonical runtime node.
func SanyLevelNodeDefaultLevelDataToString(node tlc.SemanticNode) string {
	return sanyDefaultLevelDataToString(sanyRuntimeLevelData(node))
}

// SanyLevelNodeLevelDataToString dispatches LevelNode.levelDataToString() to
// the canonical source node, preserving concrete subclass overrides.
func SanyLevelNodeLevelDataToString(node tlc.SemanticNode) string {
	data := sanyRuntimeLevelData(node)
	if data.runtimeString != nil {
		return data.runtimeString()
	}
	data.requireChecked("levelDataToString called before levelCheck")
	return sanyDefaultLevelDataToString(data)
}

func sanyAddTemporalLevelConstraintToConstants(params *sanyLevelSet[sanySemSymbol], constraints *sanySetOfLevelConstraints) {
	for node := range params.all() {
		node = sanyLevelSymbolReference(node)
		if node == nil {
			panic(tlc.NewNullPointerException())
		}
		if node.semKind() == sanyConstantDeclKind {
			action := int32(2)
			constraints.put(node, &action)
		}
	}
}
func sanyHashSetToString(set *sanyLevelSet[sanySemSymbol]) string {
	var out strings.Builder
	out.WriteString("{")
	first := true
	for symbol := range set.all() {
		if !first {
			out.WriteString(", ")
		}
		first = false
		out.WriteString(sanyLevelSymbolName(symbol))
	}
	out.WriteString("}")
	return out.String()
}
func sanyALPHashSetToString(set *sanyLevelSet[*sanyArgLevelParam]) string {
	var out strings.Builder
	out.WriteString("{")
	first := true
	for param := range set.all() {
		if param == nil {
			panic(tlc.NewNullPointerException())
		}
		if !first {
			out.WriteString(", ")
		}
		first = false
		out.WriteString("<" + sanyLevelSymbolName(param.op) + ", " + strconv.FormatInt(int64(param.i), 10) + ", " + sanyLevelSymbolName(param.param) + ">")
	}
	out.WriteString("}")
	return out.String()
}
func sanyLevelSymbolSetString(set *sanyLevelSet[sanySemSymbol]) string {
	var out strings.Builder
	out.WriteString("[")
	first := true
	for symbol := range set.all() {
		if !first {
			out.WriteString(", ")
		}
		first = false
		out.WriteString(sanyLevelSymbolString(symbol))
	}
	out.WriteString("]")
	return out.String()
}
func sanyLevelSymbolName(symbol sanySemSymbol) string {
	symbol = sanyLevelSymbolReference(symbol)
	if symbol == nil {
		panic(tlc.NewNullPointerException())
	}
	if formal, ok := symbol.(*sanyFormalParamNode); ok && formal.nullName {
		return "null"
	}
	return symbol.semName()
}

// Only the HashSet operations used by canonical level checking are exposed.
// The backing map retains Java equality, bucket order and fail-fast iteration.
type sanyLevelSet[K comparable] struct {
	entries *tlc.JavaSemanticMap[K, struct{}]
}

func newSanyLevelSymbolSet() *sanyLevelSet[sanySemSymbol] {
	return &sanyLevelSet[sanySemSymbol]{tlc.NewJavaSemanticMap[sanySemSymbol, struct{}](sanyLevelConstraintHash, sanyLevelConstraintEqual, sanyLevelConstraintTieBreak)}
}
func newSanyArgLevelSet() *sanyLevelSet[*sanyArgLevelParam] {
	return &sanyLevelSet[*sanyArgLevelParam]{tlc.NewJavaSemanticMap[*sanyArgLevelParam, struct{}](func(key *sanyArgLevelParam) int32 {
		if key == nil {
			return 0
		}
		return key.hashCode()
	}, func(a, b *sanyArgLevelParam) bool { return a != nil && a.equals(b) }, func(a, b *sanyArgLevelParam) int { return sanyLevelIdentityTieBreak(a, b) })}
}
func (s *sanyLevelSet[K]) add(key K) bool {
	if s == nil {
		panic(tlc.NewNullPointerException())
	}
	_, present := s.entries.Put(key, struct{}{})
	return !present
}
func (s *sanyLevelSet[K]) contains(key K) bool {
	if s == nil {
		panic(tlc.NewNullPointerException())
	}
	_, present := s.entries.Get(key)
	return present
}
func (s *sanyLevelSet[K]) remove(key K) bool {
	if s == nil {
		panic(tlc.NewNullPointerException())
	}
	return s.entries.Remove(key)
}
func (s *sanyLevelSet[K]) len() int {
	if s == nil {
		panic(tlc.NewNullPointerException())
	}
	return s.entries.Len()
}
func (s *sanyLevelSet[K]) clear() {
	if s == nil {
		panic(tlc.NewNullPointerException())
	}
	s.entries.Clear()
}
func (s *sanyLevelSet[K]) all() func(func(K) bool) {
	if s == nil {
		panic(tlc.NewNullPointerException())
	}
	return func(yield func(K) bool) {
		for key := range s.entries.All() {
			if !yield(key) {
				return
			}
		}
	}
}
func (s *sanyLevelSet[K]) addAll(source *sanyLevelSet[K]) {
	if s == nil {
		panic(tlc.NewNullPointerException())
	}
	for key := range source.all() {
		s.add(key)
	}
}

// SanySemanticKindNames exposes ASTConstants.kinds, including its source null
// entry at kind zero. Callers can mutate the live array as they can in Java.
var SanySemanticKindNames = [...]string{
	"null",
	"ModuleKind",
	"ConstantDeclKind",
	"VariableDeclKind",
	"BoundSymbolKind",
	"UserDefinedOpKind",
	"ModuleInstanceKind",
	"BuiltInKind",
	"OpArgKind",
	"OpApplKind",
	"LetInKind",
	"FormalParamKind",
	"TheoremKind",
	"SubstInKind",
	"AssumeProveKind",
	"ProofKind",
	"NumeralKind",
	"DecimalKind",
	"StringKind",
	"AtNodeKind",
	"AssumeKind",
	"InstanceKind",
	"NewSymbKind",
	"ThmOrAssumpDefKind",
	"NewConstantKind",
	"NewVariableKind",
	"NewStateKind",
	"NewActionKind",
	"NewTemporalKind",
	"LabelKind",
	"APSubstInKind",
	"UseKind",
	"HideKind",
	"LeafProofKind",
	"NonLeafProofKind",
	"QEDStepKind",
	"DefStepKind",
	"NumberedProofStepKind",
}

func (n *sanySemOpDeclNode) levelCheck(iter int32, errors *Diagnostics) bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	// OpDeclNode's constructor already supplies the level data. The source
	// method returns true without changing levelChecked, even for a new iter.
	return true
}

func newSanyLevelSymbolSetFrom(source *sanyLevelSet[sanySemSymbol]) *sanyLevelSet[sanySemSymbol] {
	if source == nil {
		panic(tlc.NewNullPointerException())
	}
	return &sanyLevelSet[sanySemSymbol]{tlc.NewJavaSemanticSetCopy(source.entries, sanyLevelConstraintHash, sanyLevelConstraintEqual, sanyLevelConstraintTieBreak)}
}
