package tlc

import (
	"fmt"
	"strings"
)

type OpDefNode struct {
	*SemanticNodeBase
	Symbol                    *SymbolNode
	Name                      *UniqueString
	Params                    []*SymbolNode
	Body                      SemanticNode
	StepNode                  SemanticNode
	DeclarationLocation       SourceLocation
	sourceSyntax              bool
	InRecursive               bool
	InRecursiveSection        bool
	RecursiveSection          int
	LetInLevel                int
	SymbolTablePresent        bool
	ParamsNull                bool
	Local                     bool
	OriginallyDefinedInModule *ModuleNode
	SourceDefinition          *OpDefNode
	CompoundID                []*UniqueString
	Labels                    *SemanticLabelTable
	ArgMaxLevels              []int
	ArgWeights                []int
	MinMaxLevel               [][]int
	OpLevelCond               [][][]bool
	IsLeibnizArg              []bool
	IsLeibnizValue            bool
	SANYSymbolTable           any
	arityValue                int
	aritySet                  bool
}

func (n *OpDefNode) GetLabelsHT() *SemanticLabelTable {
	n.require()
	return semanticNodeLabels(n.SemanticNodeBase, n.Labels)
}

func (n *OpDefNode) SetLabels(labels *SemanticLabelTable) {
	n.require()
	setSemanticNodeLabels(n.SemanticNodeBase, labels)
	n.Labels = labels
}

func (n *OpDefNode) GetLabel(name *UniqueString) *LabelNode {
	n.require()
	labels := n.GetLabelsHT()
	if labels == nil {
		return nil
	}
	return labels.Get(name)
}

func (n *OpDefNode) AddLabel(label *LabelNode) bool {
	n.require()
	labels := n.GetLabelsHT()
	if labels == nil {
		n.SetLabels(NewSemanticLabelTable())
		labels = n.GetLabelsHT()
	}
	return labels.Add(label)
}

func (n *OpDefNode) GetLabels() []*LabelNode {
	n.require()
	labels := n.GetLabelsHT()
	if labels == nil {
		return make([]*LabelNode, 0)
	}
	return labels.Elements()
}

func (n *OpDefNode) HasOpcode(opcode int) bool {
	n.require()
	opcode = int(int32(opcode))
	if n.Name == nil {
		panic(NewNullPointerException())
	}
	return opcode == GetOpCode(n.Name)
}

func (n *OpDefNode) GetChildren() []SemanticNode {
	n.require()
	return []SemanticNode{n.GetBody()}
}

func (n *OpDefNode) GetParams() []*SymbolNode {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface {
		RuntimeOpDefParams() ([]*SymbolNode, bool)
	}); ok {
		if params, present := provider.RuntimeOpDefParams(); present {
			return params
		}
	}
	return n.Params
}

func (n *OpDefNode) SetParams(params []*SymbolNode) {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface {
		RuntimeSetOpDefParams([]*SymbolNode) bool
	}); ok {
		provider.RuntimeSetOpDefParams(params)
	}
	n.Params = params
}

func (n *OpDefNode) GetBody() SemanticNode {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface {
		RuntimeOpDefBody() (SemanticNode, bool)
	}); ok {
		if body, present := provider.RuntimeOpDefBody(); present {
			return body
		}
	}
	return n.Body
}

func (n *OpDefNode) SetBody(body SemanticNode) {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface {
		RuntimeSetOpDefBody(SemanticNode) bool
	}); ok {
		provider.RuntimeSetOpDefBody(body)
	}
	n.Body = body
}

func (n *OpDefNode) GetArity() int {
	return n.Arity()
}

func (n *OpDefNode) GetNumberOfArgs() int {
	return n.Arity()
}

func (n *OpDefNode) SetNumberOfArgs(arity int) {
	n.require()
	arity = int(int32(arity))
	if provider, ok := n.CanonicalLevelData.(interface{ RuntimeSetSymbolArity(int) bool }); ok {
		provider.RuntimeSetSymbolArity(arity)
	}
	n.arityValue = arity
	n.aritySet = true
	if n.Symbol != nil {
		n.Symbol.Arity = arity
	}
}

func (n *OpDefNode) GetStepNode() SemanticNode {
	n.require()
	return n.StepNode
}

func (n *OpDefNode) GetMaxLevel(index int) int {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface{ RuntimeCheckedMaxLevel(int) int }); ok {
		return provider.RuntimeCheckedMaxLevel(index)
	}
	n.requireLevelChecked("getMaxLevel called before levelCheck")
	if n.Arity() == -1 {
		index = 0
	}
	return opDefArrayAt(n.ArgMaxLevels, index)
}

func (n *OpDefNode) GetWeight(index int) int {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface{ RuntimeCheckedWeight(int) int }); ok {
		return provider.RuntimeCheckedWeight(index)
	}
	n.requireLevelChecked("getWeight called before levelCheck")
	if n.Arity() == -1 {
		index = 0
	}
	return opDefArrayAt(n.ArgWeights, index)
}

func (n *OpDefNode) GetMinMaxLevel(i, j int) int {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface{ RuntimeCheckedMinMaxLevel(int, int) int }); ok {
		return provider.RuntimeCheckedMinMaxLevel(i, j)
	}
	n.requireLevelChecked("getMinMaxLevel called before levelCheck")
	if n.MinMaxLevel == nil {
		return TLCLevelConstant
	}
	return opDefArrayAt(opDefArrayAt(n.MinMaxLevel, i), j)
}

func (n *OpDefNode) GetOpLevelCond(i, j, k int) bool {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface{ RuntimeCheckedOpLevelCond(int, int, int) bool }); ok {
		return provider.RuntimeCheckedOpLevelCond(i, j, k)
	}
	n.requireLevelChecked("getOpLevelCond called before levelCheck")
	if n.OpLevelCond == nil {
		return false
	}
	return opDefArrayAt(opDefArrayAt(opDefArrayAt(n.OpLevelCond, i), j), k)
}

func (n *OpDefNode) GetArgMaxLevels() []int {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface{ RuntimeArgMaxLevels() []int }); ok {
		return provider.RuntimeArgMaxLevels()
	}
	if n.ArgMaxLevels == nil {
		return nil
	}
	result := make([]int, len(n.ArgMaxLevels))
	copy(result, n.ArgMaxLevels)
	return result
}

func (n *OpDefNode) GetArgWeights() []int {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface{ RuntimeArgWeights() []int }); ok {
		return provider.RuntimeArgWeights()
	}
	if n.ArgWeights == nil {
		return nil
	}
	result := make([]int, len(n.ArgWeights))
	copy(result, n.ArgWeights)
	return result
}

func (n *OpDefNode) GetIsLeibnizArg() []bool {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface{ RuntimeIsLeibnizArg() []bool }); ok {
		return provider.RuntimeIsLeibnizArg()
	}
	return n.IsLeibnizArg
}

func (n *OpDefNode) GetIsLeibniz() bool {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface{ RuntimeIsLeibniz() bool }); ok {
		return provider.RuntimeIsLeibniz()
	}
	return n.IsLeibnizValue
}

func (n *OpDefNode) IsLocal() bool {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface{ RuntimeSymbolLocal() (bool, bool) }); ok {
		if local, present := provider.RuntimeSymbolLocal(); present {
			return local
		}
	}
	return n.Local
}
func (n *OpDefNode) HasSource() bool {
	n.require()
	return n.SourceDefinition != nil
}
func (n *OpDefNode) GetOriginallyDefinedInModuleNode() *ModuleNode {
	n.require()
	return n.OriginallyDefinedInModule
}
func (n *OpDefNode) GetSource() *OpDefNode {
	n.require()
	if n.SourceDefinition != nil {
		return n.SourceDefinition
	}
	return n
}

func (n *OpDefNode) GetCompoundID() []*UniqueString {
	n.require()
	if n.CompoundID != nil {
		return n.CompoundID
	}
	return []*UniqueString{n.Name}
}

func (n *OpDefNode) GetLocalName() *UniqueString {
	n.require()
	if n.CompoundID != nil {
		return opDefArrayAt(n.CompoundID, len(n.CompoundID)-1)
	}
	return n.Name
}

func (n *OpDefNode) HasPath() bool {
	n.require()
	return len(n.CompoundID) > 1
}

func (n *OpDefNode) GetPathName() *UniqueString {
	n.require()
	if n.CompoundID == nil {
		return UniqueStringOf("")
	}
	var result *UniqueString
	for i := 0; i < len(n.CompoundID)-1; i++ {
		if result == nil {
			result = n.CompoundID[i]
		} else {
			result = result.ConcatString("!")
			result = result.Concat(n.CompoundID[i])
		}
	}
	return result
}

func NewOpDefNode(name string, params []*SymbolNode, body SemanticNode) *OpDefNode {
	return NewOpDefNodeForSymbol(NewSymbolNode(name), params, body)
}

func NewOpDefNodeForSymbol(symbol *SymbolNode, params []*SymbolNode, body SemanticNode) *OpDefNode {
	return NewOpDefNodeForSymbolWithBase(symbol, params, body, nil)
}

// A parser adapter borrows the already constructed SANY identity. Standalone
// runtime definitions allocate their own semantic base through the same path.
func NewOpDefNodeForSymbolWithBase(symbol *SymbolNode, params []*SymbolNode, body SemanticNode, base *SemanticNodeBase) *OpDefNode {
	sourceSyntax := base != nil
	paramsNull := params == nil
	if params == nil {
		params = make([]*SymbolNode, 0)
	}
	if symbol == nil {
		symbol = NewSymbolNode("")
	}
	symbol.MarkUserDefinedOp()
	symbol.Arity = len(params)
	name := symbol.Name
	image := ""
	if name != nil {
		image = name.String()
	}
	if base == nil {
		owned := NewSemanticNodeBase(SemanticUserDefinedOpKind, image)
		base = &owned
	}
	definition := &OpDefNode{
		SemanticNodeBase: base,
		Symbol:           symbol,
		Name:             name,
		Params:           params,
		Body:             body,
		sourceSyntax:     sourceSyntax,
		LetInLevel:       -1,
		RecursiveSection: -1,
		ParamsNull:       paramsNull,
		arityValue:       len(params),
		aritySet:         true,
		ArgMaxLevels:     make([]int, len(params)),
		ArgWeights:       make([]int, len(params)),
		IsLeibnizArg:     make([]bool, len(params)),
		IsLeibnizValue:   true,
	}
	for i := range definition.ArgMaxLevels {
		definition.ArgMaxLevels[i] = TLCLevelTemporal
		definition.IsLeibnizArg[i] = true
	}
	symbol.Definition = definition
	symbol.SemanticBase = definition.SemanticNodeBase
	return definition
}

func (n *OpDefNode) Arity() int {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface{ RuntimeSymbolArity() (int, bool) }); ok {
		if arity, present := provider.RuntimeSymbolArity(); present {
			return arity
		}
	}
	if n.Kind() == SemanticBuiltInKind {
		return n.Symbol.Arity
	}
	if n.aritySet {
		return n.arityValue
	}
	return len(n.Params)
}

func (n *OpDefNode) SetInRecursive(value bool) {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface{ RuntimeSetOpDefInRecursive(bool) bool }); ok {
		provider.RuntimeSetOpDefInRecursive(value)
	}
	n.InRecursive = value
}

func (n *OpDefNode) GetInRecursive() bool {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface {
		RuntimeOpDefRecursion() (bool, bool, int, int, bool)
	}); ok {
		if recursive, _, _, _, present := provider.RuntimeOpDefRecursion(); present {
			return recursive
		}
	}
	return n.InRecursive
}

func (n *OpDefNode) GetInRecursiveSection() bool {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface {
		RuntimeOpDefRecursion() (bool, bool, int, int, bool)
	}); ok {
		if _, inSection, _, _, present := provider.RuntimeOpDefRecursion(); present {
			return inSection
		}
	}
	return n.InRecursiveSection
}

func (n *OpDefNode) GetRecursiveSection() int {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface {
		RuntimeOpDefRecursion() (bool, bool, int, int, bool)
	}); ok {
		if _, _, section, _, present := provider.RuntimeOpDefRecursion(); present {
			return section
		}
	}
	return n.RecursiveSection
}

func (n *OpDefNode) GetLetInLevel() int {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface {
		RuntimeOpDefRecursion() (bool, bool, int, int, bool)
	}); ok {
		if _, _, _, level, present := provider.RuntimeOpDefRecursion(); present {
			return level
		}
	}
	return n.LetInLevel
}

func (n *OpDefNode) SetDeclarationLocation(location SourceLocation) {
	n.require()
	n.DeclarationLocation = location
}

func (n *OpDefNode) GetDeclarationLocation() SourceLocation {
	n.require()
	if n.sourceSyntax {
		// Parser-backed definitions share their current syntax with SANY.
		// Action.getDeclaration reads one[0] on every call, including after
		// the syntax or its children have been replaced or removed.
		if syntax, ok := n.GetTreeNode().(interface{ GetDeclarationLocation() SourceLocation }); ok {
			return syntax.GetDeclarationLocation()
		}
		return NullSourceLocation
	}
	return n.DeclarationLocation
}

func (n *OpDefNode) String() string {
	n.require()
	if n.hasSourceSyntax() {
		return semanticNodeJavaString(n)
	}
	if n.Name == nil {
		return "<anonymous>"
	}
	return n.Name.String()
}

func (n *OpDefNode) GetSignature() string {
	n.require()
	var out strings.Builder
	out.WriteString(n.Name.String())
	if n.Arity() > 0 && n.Kind() != SemanticBuiltInKind {
		out.WriteByte('(')
		params := n.GetParams()
		if params == nil {
			panic(NewNullPointerException())
		}
		paramAt := func(index int) *SymbolNode {
			param := opDefArrayAt(params, index)
			if param == nil {
				panic(NewNullPointerException())
			}
			return param
		}
		for i := 0; i < len(params)-1; i++ {
			param := paramAt(i)
			if param.GetTreeNode() != nil {
				out.WriteString(syntaxNodeHumanReadableImage(param))
				out.WriteString(", ")
			}
		}
		last := paramAt(len(params) - 1)
		if last.GetTreeNode() != nil {
			out.WriteString(syntaxNodeHumanReadableImage(last))
		}
		out.WriteByte(')')
	}
	return out.String()
}

func (n *OpDefNode) IsBuiltIn() bool {
	n.require()
	return semanticNodeIsBuiltIn(n, n.Symbol != nil && n.Symbol.Kind == SymbolBuiltIn)
}

func (n *OpDefNode) GetComment() string {
	n.require()
	return semanticOpDefOrDeclComment(n.hasSourceSyntax(), n.GetTreeNode())
}

func semanticOpDefOrDeclComment(hasSourceSyntax bool, syntax any) string {
	var out strings.Builder
	if hasSourceSyntax && semanticExploreNull(syntax) {
		panic(NewNullPointerException())
	}
	if tree, ok := syntax.(interface{ GetAttachedComments() []string }); ok {
		for _, comment := range tree.GetAttachedComments() {
			out.WriteString(comment)
			out.WriteByte('\n')
		}
	} else if hasSourceSyntax {
		panic(valueStreamClassCast(syntax, "tla2sany.parser.SyntaxTreeNode"))
	}
	return strings.TrimFunc(strings.ReplaceAll(out.String(), "\n$", ""), func(r rune) bool { return r <= 0x20 })
}

func (n *OpDefNode) GetHumanReadableImage() string {
	n.require()
	var out strings.Builder
	out.WriteString(n.GetComment())
	out.WriteByte('\n')
	tree, ok := n.GetTreeNode().(interface{ GetOneHumanReadableImages() []string })
	var images []string
	if ok {
		images = tree.GetOneHumanReadableImages()
	}
	if images != nil {
		for _, image := range images {
			out.WriteString(image)
			out.WriteByte(' ')
		}
	} else {
		if n.hasSourceSyntax() {
			out.WriteString(n.String())
		} else {
			out.WriteString(n.GetSourceLocation().String())
		}
	}
	return strings.TrimFunc(out.String(), func(r rune) bool { return r <= 0x20 })
}

func (n *OpDefNode) require() {
	if n == nil {
		panic(NewNullPointerException())
	}
}

func (n *OpDefNode) requireLevelChecked(message string) {
	n.require()
	if n.LevelChecked == 0 {
		panic(NewWrongInvocationException(message))
	}
}

func opDefArrayAt[T any](values []T, index int) T {
	if values == nil {
		panic(NewNullPointerException())
	}
	index = int(int32(index))
	if index < 0 || index >= len(values) {
		panic(NewArrayIndexOutOfBoundsException(index, len(values)))
	}
	return values[index]
}

type OperatorEvalFunc func(args []Value, control int) (Value, error)
type EvaluatingEvalFunc func(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error)
type CallableEvalFunc func(args []Value) (func() (any, error), error)

// OperatorValue represents Java's OpValue evaluation contract.
type OperatorValue interface {
	Value
	Eval([]Value, int) (Value, error)
	EvalWithTool(*Tool, []SemanticNode, *Context, *TLCStateMut, *TLCStateMut, int, CostModel) (Value, error)
}

var (
	_ OperatorValue = (*OpLambdaValue)(nil)
	_ OperatorValue = (*OpRcdValue)(nil)
	_ OperatorValue = (*MethodValue)(nil)
	_ OperatorValue = (*EvaluatingValue)(nil)
	_ OperatorValue = (*PriorityEvaluatingValue)(nil)
	_ OperatorValue = (*CallableValue)(nil)
)

func evalOperatorValueWithToolDefault(op interface {
	Eval([]Value, int) (Value, error)
}, tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
	if isNil(op) {
		panic(NewNullPointerException())
	}
	control = int(int32(control))
	if args == nil {
		panic(NewNullPointerException())
	}
	if len(args) == 0 {
		return op.Eval([]Value{}, control)
	}
	argVals := make([]Value, len(args))
	for i, arg := range args {
		value, err := tool.Eval(arg, con, state, pstate, control, cm)
		if err != nil {
			return nil, err
		}
		argVals[i] = value
	}
	return op.Eval(argVals, control)
}

type operatorValueBase struct {
	BaseValue
	owner            Value
	KindValue        ValueKind
	Label            string
	NormalizeMessage string
}

// Shared Go methods represent Java's concrete methods and inherited defaults.
// Keep their receiver identity, source, and dynamically dispatched printing.
func (b *operatorValueBase) receiver() Value {
	if b == nil {
		panic(NewNullPointerException())
	}
	if b.owner != nil {
		return b.owner
	}
	return b
}

func (b *operatorValueBase) Kind() ValueKind {
	if b == nil {
		panic(NewNullPointerException())
	}
	return b.KindValue
}
func (b *operatorValueBase) KindString() string {
	return ValueKindString(b.receiver())
}
func (b *operatorValueBase) GetCostModel() CostModel {
	if b == nil {
		panic(NewNullPointerException())
	}
	return b.BaseValue.GetCostModel()
}

func (b *operatorValueBase) Compare(other Value) (resultInt int, err error) {
	defer catchValueFailure(b.receiver(), &err)
	if b.KindValue == MethodValueKind {
		return 0, b.runtimeFailure(b.methodComparisonMessage(other))
	}
	operator := ValuesPPR(b.receiver())
	if other == nil {
		panic(NewNullPointerException())
	}
	return 0, b.runtimeFailure(fmt.Sprintf("Attempted to compare operator %s with value:\n%s", operator, ValuesPPRString(other.String())))
}

func (b *operatorValueBase) Equal(other Value) (resultBool bool, err error) {
	defer catchValueFailure(b.receiver(), &err)
	if b.KindValue == MethodValueKind {
		return false, b.runtimeFailure(b.methodComparisonMessage(other))
	}
	operator := ValuesPPR(b.receiver())
	if other == nil {
		panic(NewNullPointerException())
	}
	return false, b.runtimeFailure(fmt.Sprintf("Attempted to check equality of operator %s with value:\n%s", operator, ValuesPPRString(other.String())))
}

func (b *operatorValueBase) Member(elem Value) (resultBool bool, err error) {
	defer catchValueFailure(b.receiver(), &err)
	if elem == nil {
		panic(NewNullPointerException())
	}
	if b.KindValue == MethodValueKind {
		if elem != nil {
			_ = elem.String()
		}
		return false, b.runtimeFailure(fmt.Sprintf("%s\nis an element of operator %s", ValuesPPRString(elem.String()), b.receiver().String()))
	}
	return false, b.runtimeFailure(fmt.Sprintf("Attempted to check if the value:\n%s\nis an element of operator %s", ValuesPPRString(elem.String()), ValuesPPR(b.receiver())))
}

func (b *operatorValueBase) IsFinite() (resultBool bool, err error) {
	defer catchValueFailure(b.receiver(), &err)
	return false, b.runtimeFailure(fmt.Sprintf("Attempted to check if the operator %s is a finite set.", b.diagnosticString()))
}

func (b *operatorValueBase) Size() (resultInt int, err error) {
	defer catchValueFailure(b.receiver(), &err)
	return 0, b.runtimeFailure(fmt.Sprintf("Attempted to compute the number of elements in the operator %s.", b.diagnosticString()))
}

func (b *operatorValueBase) Normalize() Value {
	defer catchValueFailure(b.receiver(), nil)
	panic(NewWrongInvocationException(b.normalizeMessage()))
}

func (b *operatorValueBase) DeepNormalize() { _ = b.receiver() }

func (b *operatorValueBase) IsNormalized() bool {
	defer catchValueFailure(b.receiver(), nil)
	panic(NewWrongInvocationException(b.normalizeMessage()))
}

func (b *operatorValueBase) IsDefined() bool { return true }
func (b *operatorValueBase) DeepCopy() Value { return b.receiver() }

func (b *operatorValueBase) FingerPrint(fp uint64) uint64 {
	return unsupportedValueFingerprint(b.receiver())
}

func (b *operatorValueBase) Permute(*MVPerm) Value {
	return unsupportedValuePermutation(b.receiver())
}

func (b *operatorValueBase) TakeExcept(ex *ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(b.receiver(), &err)
	return nil, b.runtimeFailure(fmt.Sprintf("Attempted to appy EXCEPT construct to the operator %s.", b.diagnosticString()))
}

func (b *operatorValueBase) TakeExcepts(exs []*ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(b.receiver(), &err)
	return nil, b.runtimeFailure(fmt.Sprintf("Attempted to apply EXCEPT construct to the operator %s.", b.diagnosticString()))
}

func (b *operatorValueBase) String() string {
	return ValueToString(b.receiver(), "", true)
}

func (b *operatorValueBase) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	defer catchValueFailure(b.receiver(), nil)
	sb.WriteString(b.Label)
	return sb
}

func (b *operatorValueBase) diagnosticString() string {
	if b.KindValue == MethodValueKind {
		return b.receiver().String()
	}
	return ValuesPPR(b.receiver())
}

func (b *operatorValueBase) methodComparisonMessage(other Value) string {
	// Preserve MethodValue/EvaluatingValue's Java precedence quirk: the entire
	// concatenation is tested for null, so only obj.toString() becomes the error.
	_ = b.receiver().String()
	if other != nil {
		_ = other.String()
	}
	if other == nil {
		panic(NewNullPointerException())
	}
	return ValuesPPRString(other.String())
}

func (b *operatorValueBase) normalizeMessage() string {
	if b.NormalizeMessage != "" {
		return b.NormalizeMessage
	}
	return "It is a TLC bug: Attempted to normalize an operator."
}

type OpLambdaValue struct {
	operatorValueBase
	OpDef   *OpDefNode
	Tool    *Tool
	Con     *Context
	State   *TLCStateMut
	PState  *TLCStateMut
	Control int

	ownersCaptured bool
	opDefOwner     *OpDefNode
	toolOwner      *Tool
	conOwner       *Context
	stateOwner     *TLCStateMut
	pstateOwner    *TLCStateMut
	controlOwner   int
}

func NewOpLambdaValue(opDef *OpDefNode, tool *Tool, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cms ...CostModel) *OpLambdaValue {
	if len(cms) > 1 {
		panic(NewIllegalArgumentException("NewOpLambdaValue accepts at most one cost model"))
	}
	cm := DoNotRecordCostModel
	if len(cms) > 0 {
		cm = cms[0]
	}
	out := &OpLambdaValue{
		operatorValueBase: operatorValueBase{BaseValue: newBaseValue(cm), KindValue: OpLambdaValueKind, NormalizeMessage: "Should not normalize an operator."},
		OpDef:             opDef,
		Tool:              tool,
		Con:               con,
		State:             state,
		PState:            pstate,
		Control:           int(int32(control)),
		ownersCaptured:    true,
		opDefOwner:        opDef,
		toolOwner:         tool,
		conOwner:          con,
		stateOwner:        state,
		pstateOwner:       pstate,
		controlOwner:      int(int32(control)),
	}
	out.owner = out
	return out
}

func (v *OpLambdaValue) ownedOpDef() *OpDefNode {
	if v.ownersCaptured {
		return v.opDefOwner
	}
	return v.OpDef
}

func (v *OpLambdaValue) ownedTool() *Tool {
	if v.ownersCaptured {
		return v.toolOwner
	}
	return v.Tool
}

func (v *OpLambdaValue) ownedContext() *Context {
	if v.ownersCaptured {
		return v.conOwner
	}
	return v.Con
}

func (v *OpLambdaValue) ownedState() *TLCStateMut {
	if v.ownersCaptured {
		return v.stateOwner
	}
	return v.State
}

func (v *OpLambdaValue) ownedPState() *TLCStateMut {
	if v.ownersCaptured {
		return v.pstateOwner
	}
	return v.PState
}

func (v *OpLambdaValue) ownedControl() int {
	if v.ownersCaptured {
		return v.controlOwner
	}
	return v.Control
}

func NewOpLambdaValueFrom(other *OpLambdaValue, tool *Tool) *OpLambdaValue {
	if other == nil {
		panic(NewNullPointerException())
	}
	return NewOpLambdaValue(other.ownedOpDef(), tool, other.ownedContext(), other.ownedState(), other.ownedPState(), other.ownedControl())
}

func (v *OpLambdaValue) Kind() ValueKind {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.operatorValueBase.Kind()
}
func (v *OpLambdaValue) KindString() string {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.operatorValueBase.KindString()
}
func (v *OpLambdaValue) GetCostModel() CostModel {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.operatorValueBase.GetCostModel()
}
func (v *OpLambdaValue) Compare(other Value) (int, error) { return v.operator().Compare(other) }
func (v *OpLambdaValue) Equal(other Value) (bool, error)  { return v.operator().Equal(other) }
func (v *OpLambdaValue) Member(elem Value) (bool, error)  { return v.operator().Member(elem) }
func (v *OpLambdaValue) IsFinite() (bool, error)          { return v.operator().IsFinite() }
func (v *OpLambdaValue) Size() (int, error)               { return v.operator().Size() }
func (v *OpLambdaValue) FingerPrint(fp uint64) uint64     { return v.operator().FingerPrint(fp) }
func (v *OpLambdaValue) TakeExcept(ex *ValueExcept) (Value, error) {
	return v.operator().TakeExcept(ex)
}
func (v *OpLambdaValue) TakeExcepts(exs []*ValueExcept) (Value, error) {
	return v.operator().TakeExcepts(exs)
}
func (v *OpLambdaValue) operator() *operatorValueBase {
	if v == nil {
		panic(NewNullPointerException())
	}
	return &v.operatorValueBase
}
func (v *OpLambdaValue) Normalize() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.operatorValueBase.Normalize()
}
func (v *OpLambdaValue) IsNormalized() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.operatorValueBase.IsNormalized()
}
func (v *OpLambdaValue) IsDefined() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.operatorValueBase.IsDefined()
}
func (v *OpLambdaValue) DeepCopy() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v
}
func (v *OpLambdaValue) DeepNormalize() {
	if v == nil {
		panic(NewNullPointerException())
	}
	v.operatorValueBase.DeepNormalize()
}

func (v *OpLambdaValue) String() string {
	if v == nil {
		panic(NewNullPointerException())
	}
	return ValueToString(v, "", true)
}

func (v *OpLambdaValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	opDef := v.ownedOpDef()
	if opDef == nil || opDef.Name == nil || sb == nil {
		panic(NewNullPointerException())
	}
	sb.WriteString("<Operator ")
	sb.WriteString(opDef.Name.String())
	sb.WriteString(">")
	return sb
}
func (v *OpLambdaValue) Permute(perm *MVPerm) Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.operatorValueBase.Permute(perm)
}

func (v *OpLambdaValue) Eval(args []Value, control int) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	control = int(int32(control))
	defer catchValueFailure(v, &err)
	opDef := v.ownedOpDef()
	if opDef == nil {
		return nil, NewNullPointerException(`Cannot invoke "tla2sany.semantic.OpDefNode.getArity()" because "this.opDef" is null`)
	}
	arity := opDef.Arity()
	if args == nil {
		panic(NewNullPointerException())
	}
	if arity != len(args) {
		return nil, v.runtimeFailure(fmt.Sprintf("Applying the operator %s with wrong number of arguments.", ValuesPPR(v)))
	}
	ctx := v.ownedContext()
	for i, param := range opDef.GetParams() {
		if ctx == nil {
			panic(NewNullPointerException())
		}
		ctx = ctx.Cons(param, args[i])
	}
	tool := v.ownedTool()
	if tool == nil {
		panic(NewNullPointerException())
	}
	if EvalIsEnabled(v.ownedControl()) {
		control = EvalSetEnabled(control)
	}
	return tool.Eval(opDef.GetBody(), ctx, v.ownedState(), v.ownedPState(), control, DoNotRecordCostModel)
}

func (v *OpLambdaValue) EvalWithTool(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
	return evalOperatorValueWithToolDefault(v, tool, args, con, state, pstate, control, cm)
}

type OpRcdValue struct {
	operatorValueBase
	Domain [][]Value
	Values []Value

	operandsCaptured bool
	domainOwner      [][]Value
	valuesOwner      []Value
}

func NewOpRcdValue() *OpRcdValue {
	out := &OpRcdValue{
		operatorValueBase: operatorValueBase{KindValue: OpRcdValueKind, Label: "<Operator record>", NormalizeMessage: "Should not normalize an operator."},
	}
	out.owner = out
	out.captureOperands(nil, nil)
	return out
}

func NewOpRcdValueFrom(domain [][]Value, values []Value) *OpRcdValue {
	out := &OpRcdValue{
		operatorValueBase: operatorValueBase{KindValue: OpRcdValueKind, Label: "<Operator record>", NormalizeMessage: "Should not normalize an operator."},
		Domain:            domain,
		Values:            values,
	}
	out.owner = out
	out.captureOperands(domain, values)
	return out
}

func (v *OpRcdValue) ownedDomain() [][]Value {
	if v.operandsCaptured {
		return v.domainOwner
	}
	return v.Domain
}

func (v *OpRcdValue) ownedValues() []Value {
	if v.operandsCaptured {
		return v.valuesOwner
	}
	return v.Values
}

func (v *OpRcdValue) captureOperands(domain [][]Value, values []Value) {
	v.Domain, v.Values = domain, values
	v.operandsCaptured, v.domainOwner, v.valuesOwner = true, domain, values
}

func (v *OpRcdValue) Kind() ValueKind {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.operatorValueBase.Kind()
}
func (v *OpRcdValue) KindString() string {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.operatorValueBase.KindString()
}
func (v *OpRcdValue) GetCostModel() CostModel {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.operatorValueBase.GetCostModel()
}
func (v *OpRcdValue) Compare(other Value) (int, error) { return v.operator().Compare(other) }
func (v *OpRcdValue) Equal(other Value) (bool, error)  { return v.operator().Equal(other) }
func (v *OpRcdValue) Member(elem Value) (bool, error)  { return v.operator().Member(elem) }
func (v *OpRcdValue) IsFinite() (bool, error)          { return v.operator().IsFinite() }
func (v *OpRcdValue) Size() (int, error)               { return v.operator().Size() }
func (v *OpRcdValue) FingerPrint(fp uint64) uint64     { return v.operator().FingerPrint(fp) }
func (v *OpRcdValue) TakeExcept(ex *ValueExcept) (Value, error) {
	return v.operator().TakeExcept(ex)
}
func (v *OpRcdValue) TakeExcepts(exs []*ValueExcept) (Value, error) {
	return v.operator().TakeExcepts(exs)
}
func (v *OpRcdValue) operator() *operatorValueBase {
	if v == nil {
		panic(NewNullPointerException())
	}
	return &v.operatorValueBase
}

func (v *OpRcdValue) AddLine(values []Value) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	args := make([]Value, len(values)-2)
	copy(args, values[1:len(values)-1])
	v.captureOperands(append(v.ownedDomain(), args), append(v.ownedValues(), values[len(values)-1]))
}

func (v *OpRcdValue) Eval(args []Value, control int) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	_ = control
	for i, vals := range v.ownedDomain() {
		if args == nil || vals == nil {
			panic(NewNullPointerException())
		}
		if len(args) != len(vals) {
			return nil, v.runtimeFailure(fmt.Sprintf("Attempted to apply the operator %s\nwith wrong number of arguments.", ValuesPPR(v)))
		}
		matched := true
		for j := range vals {
			eq, err := vals[j].Equal(args[j])
			if err != nil {
				return nil, err
			}
			if !eq {
				matched = false
				break
			}
		}
		if matched {
			return v.ownedValues()[i], nil
		}
	}
	operator := ValuesPPR(v)
	if args == nil {
		panic(NewNullPointerException())
	}
	return nil, v.runtimeFailure(fmt.Sprintf("Attempted to apply operator:\n%s\nto arguments (%s), which is undefined.", operator, joinValueStrings(args, ", ")))
}

func (v *OpRcdValue) EvalWithTool(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
	return evalOperatorValueWithToolDefault(v, tool, args, con, state, pstate, control, cm)
}

func (v *OpRcdValue) Initialize() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	for i, arguments := range v.ownedDomain() {
		if arguments == nil {
			panic(NewNullPointerException())
		}
		for j, argument := range arguments {
			if argument == nil {
				panic(NewNullPointerException())
			}
			v.ownedDomain()[i][j] = InitializeValue(argument)
		}
		if v.ownedValues()[i] == nil {
			panic(NewNullPointerException())
		}
		v.ownedValues()[i] = InitializeValue(v.ownedValues()[i])
	}
	return v
}

func (v *OpRcdValue) IsDefined() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	for _, value := range v.ownedValues() {
		if !value.IsDefined() {
			return false
		}
	}
	return true
}

func (v *OpRcdValue) Normalize() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.operatorValueBase.Normalize()
}
func (v *OpRcdValue) IsNormalized() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.operatorValueBase.IsNormalized()
}
func (v *OpRcdValue) DeepCopy() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v
}
func (v *OpRcdValue) DeepNormalize() {
	if v == nil {
		panic(NewNullPointerException())
	}
	v.operatorValueBase.DeepNormalize()
}
func (v *OpRcdValue) Permute(perm *MVPerm) Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.operatorValueBase.Permute(perm)
}

func (v *OpRcdValue) String() string {
	if v == nil {
		panic(NewNullPointerException())
	}
	return ValueToString(v, "", true)
}

func (v *OpRcdValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	sb.WriteString("{ ")
	for i, value := range v.ownedValues() {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("<")
		for _, arg := range v.ownedDomain()[i] {
			sb = appendValueString(arg, sb, offset, swallow)
			sb.WriteString(", ")
		}
		sb = appendValueString(value, sb, offset, swallow)
		sb.WriteString(">")
	}
	sb.WriteString("}")
	return sb
}

type MethodValue struct {
	operatorValueBase
	Name     string
	MinLevel int
	// ParameterCount retains the reflected Java method's arity for native
	// override registration. -1 denotes an unrepresented method signature.
	ParameterCount int
	// Instance and DeclaringClass retain the receiver slot present on a Java
	// MethodHandle for a public non-static annotated method.
	Instance       bool
	DeclaringClass string
	EvalFunc       OperatorEvalFunc

	metadataCaptured bool
	metadataOwner    methodValueMetadata
}

type methodValueMetadata struct {
	name           string
	minLevel       int
	parameterCount int
	instance       bool
	declaringClass string
	evalFunc       OperatorEvalFunc
}

func NewMethodValue(name string, minLevel int, eval OperatorEvalFunc) *MethodValue {
	out := &MethodValue{
		operatorValueBase: operatorValueBase{KindValue: MethodValueKind, Label: "<Java Method: " + name + ">", NormalizeMessage: "It is a TLC bug: Attempted to normalize an operator."},
		Name:              name,
		MinLevel:          int(int32(minLevel)),
		ParameterCount:    -1,
		EvalFunc:          eval,
	}
	out.owner = out
	out.captureMetadata()
	return out
}

func (v *MethodValue) captureMetadata() {
	v.metadataOwner = methodValueMetadata{
		name: v.Name, minLevel: v.MinLevel, parameterCount: v.ParameterCount,
		instance: v.Instance, declaringClass: v.DeclaringClass, evalFunc: v.EvalFunc,
	}
	v.metadataCaptured = true
}

func (v *MethodValue) ownedMetadata() methodValueMetadata {
	if v.metadataCaptured {
		return v.metadataOwner
	}
	return methodValueMetadata{
		name: v.Name, minLevel: v.MinLevel, parameterCount: v.ParameterCount,
		instance: v.Instance, declaringClass: v.DeclaringClass, evalFunc: v.EvalFunc,
	}
}

func (v *MethodValue) Kind() ValueKind {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.operatorValueBase.Kind()
}
func (v *MethodValue) KindString() string {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.operatorValueBase.KindString()
}
func (v *MethodValue) GetCostModel() CostModel {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.operatorValueBase.GetCostModel()
}
func (v *MethodValue) Compare(other Value) (int, error) { return v.operator().Compare(other) }
func (v *MethodValue) Equal(other Value) (bool, error)  { return v.operator().Equal(other) }
func (v *MethodValue) Member(elem Value) (bool, error)  { return v.operator().Member(elem) }
func (v *MethodValue) IsFinite() (bool, error)          { return v.operator().IsFinite() }
func (v *MethodValue) Size() (int, error)               { return v.operator().Size() }
func (v *MethodValue) FingerPrint(fp uint64) uint64     { return v.operator().FingerPrint(fp) }
func (v *MethodValue) TakeExcept(ex *ValueExcept) (Value, error) {
	return v.operator().TakeExcept(ex)
}
func (v *MethodValue) TakeExcepts(exs []*ValueExcept) (Value, error) {
	return v.operator().TakeExcepts(exs)
}
func (v *MethodValue) String() string { return v.operator().String() }
func (v *MethodValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	if v == nil {
		panic(NewNullPointerException())
	}
	sb.WriteString("<Java Method: " + v.ownedMetadata().name + ">")
	return sb
}
func (v *MethodValue) operator() *operatorValueBase {
	if v == nil {
		panic(NewNullPointerException())
	}
	return &v.operatorValueBase
}

func (v *MethodValue) Eval(args []Value, control int) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	control = int(int32(control))
	defer catchValueFailure(v, &err)
	defer catchJavaMethodFailure(v.ownedMetadata().name, &resultValue, &err, true)
	if args == nil {
		panic(NewNullPointerException())
	}
	if v.ownedMetadata().instance {
		resultType := methodResultType(v.ownedMetadata().name)
		if v.ownedMetadata().parameterCount == 0 {
			if len(args) == 0 {
				return nil, fmt.Errorf("handle's method type (%s)%s but found ()Value", v.ownedMetadata().declaringClass, resultType)
			}
			return nil, fmt.Errorf("Cannot cast [Ltlc2.value.impl.Value; to %s", v.ownedMetadata().declaringClass)
		}
		if len(args) == 0 {
			return nil, fmt.Errorf("handle's method type (%s,IValue[])%s but found ()Value", v.ownedMetadata().declaringClass, resultType)
		}
		return nil, fmt.Errorf("cannot convert MethodHandle(%s,IValue[])%s to (Value[])Value", v.ownedMetadata().declaringClass, resultType)
	}
	if v.ownedMetadata().parameterCount == 0 && len(args) != 0 {
		return nil, fmt.Errorf("cannot convert MethodHandle()%s to (Value[])Value", methodResultType(v.ownedMetadata().name))
	}
	if v.ownedMetadata().parameterCount > 0 && len(args) != v.ownedMetadata().parameterCount {
		if len(args) == 0 {
			// Source MethodValue uses a distinct invocation path for an empty
			// array. Retain its diagnostic at the native override boundary.
			return nil, fmt.Errorf("handle's method type (IValue[])%s but found ()Value", methodResultType(v.ownedMetadata().name))
		}
		return nil, fmt.Errorf("array is not of length %d", v.ownedMetadata().parameterCount)
	}
	result, failure := v.ownedMetadata().evalFunc(args, control)
	if isNil(result) {
		result = nil
	}
	return result, failure
}

func (v *MethodValue) EvalWithTool(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
	return evalOperatorValueWithToolDefault(v, tool, args, con, state, pstate, control, cm)
}

func (v *MethodValue) Initialize() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	v.DeepNormalize()
	return v
}

// Read the declared result from the registered source signature. Restrict the
// read to the declaration prefix so a throws clause cannot become the result.
func methodResultType(signature string) string {
	if args := strings.IndexByte(signature, '('); args >= 0 {
		signature = signature[:args]
	}
	fields := strings.Fields(signature)
	if len(fields) < 2 {
		return "Value"
	}
	result := fields[len(fields)-2]
	if dot := strings.LastIndexByte(result, '.'); dot >= 0 {
		result = result[dot+1:]
	}
	return result
}

// MethodValue preserves direct EvalExceptions; unevaluated and callable
// overrides use Assert.fail for every failure in their Java catch(Throwable).
// The defer belongs inside CallableValue's argument evaluation boundary.
func catchJavaMethodFailure(signature string, result *Value, err *error, preserveEval bool) {
	if failure := recover(); failure != nil {
		cause, ok := failure.(error)
		if !ok {
			cause = fmt.Errorf("%v", failure)
		}
		*result = nil
		*err = adaptJavaMethodFailure(signature, cause, preserveEval)
		return
	}
	if *err != nil {
		*result = nil
		*err = adaptJavaMethodFailure(signature, *err, preserveEval)
	}
}

func adaptJavaMethodFailure(signature string, cause error, preserveEval bool) error {
	if preserveEval {
		if isValueEvalException(cause) {
			return cause
		}
		if isJavaNullPointerException(cause) {
			// A native nil dereference has Go's runtime text, while the source
			// implicit null failure has no detail message. Explicit null
			// exceptions retain their supplied nullable message.
			var message *string
			if throwable, ok := cause.(interface{ GetMessage() *string }); ok {
				message = throwable.GetMessage()
			}
			return NewEvalExceptionNullable(ECTLCModuleValueJavaMethodOverride, javaString(signature), message)
		}
	}
	message := javaThrowableDetailMessage(cause)
	if preserveEval && message == nil {
		message = javaString(javaThrowableStackTrace(cause))
	}
	failure := newTLCErrorCodeNullable(ECTLCModuleValueJavaMethodOverride, javaString(signature), message)
	failure.Runtime = true
	return failure
}

func (v *MethodValue) Normalize() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.operatorValueBase.Normalize()
}
func (v *MethodValue) IsNormalized() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.operatorValueBase.IsNormalized()
}
func (v *MethodValue) IsDefined() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.operatorValueBase.IsDefined()
}
func (v *MethodValue) DeepCopy() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v
}
func (v *MethodValue) DeepNormalize() {
	if v == nil {
		panic(NewNullPointerException())
	}
	v.operatorValueBase.DeepNormalize()
}
func (v *MethodValue) Permute(perm *MVPerm) Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.operatorValueBase.Permute(perm)
}

func (v *MethodValue) GetMinLevel() int {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.ownedMetadata().minLevel
}

type EvaluatingValue struct {
	operatorValueBase
	Name           string
	MinLevel       int
	Priority       int
	OpDef          *OpDefNode
	Instance       bool
	DeclaringClass string
	EvalFunc       EvaluatingEvalFunc

	metadataCaptured bool
	metadataOwner    evaluatingValueMetadata
}

type evaluatingValueMetadata struct {
	name           string
	minLevel       int
	priority       int
	opDef          *OpDefNode
	instance       bool
	declaringClass string
	evalFunc       EvaluatingEvalFunc
}

func NewEvaluatingValue(name string, minLevel int, priority int, opDef *OpDefNode, eval EvaluatingEvalFunc) *EvaluatingValue {
	if GetOpCode(opDef.Name) != 0 {
		panic(javaMethodOverrideRuntimeError(name, "@Evaluation fallback to pure TLA+ definition only works for user-defined operators."))
	}
	return newEvaluatingValue(name, minLevel, priority, opDef, eval)
}

// CallableValue uses Java's protected constructor, which skips the pure-fallback guard.
func newEvaluatingValue(name string, minLevel int, priority int, opDef *OpDefNode, eval EvaluatingEvalFunc) *EvaluatingValue {
	out := &EvaluatingValue{
		operatorValueBase: operatorValueBase{KindValue: MethodValueKind, Label: "<Java Method: " + name + ">", NormalizeMessage: "It is a TLC bug: Attempted to normalize an operator."},
		Name:              name,
		MinLevel:          int(int32(minLevel)),
		Priority:          int(int32(priority)),
		OpDef:             opDef,
		EvalFunc:          eval,
	}
	out.owner = out
	out.captureMetadata()
	return out
}

func (v *EvaluatingValue) captureMetadata() {
	v.metadataOwner = evaluatingValueMetadata{
		name: v.Name, minLevel: v.MinLevel, priority: v.Priority, opDef: v.OpDef,
		instance: v.Instance, declaringClass: v.DeclaringClass, evalFunc: v.EvalFunc,
	}
	v.metadataCaptured = true
}

func (v *EvaluatingValue) ownedMetadata() evaluatingValueMetadata {
	if v.metadataCaptured {
		return v.metadataOwner
	}
	return evaluatingValueMetadata{
		name: v.Name, minLevel: v.MinLevel, priority: v.Priority, opDef: v.OpDef,
		instance: v.Instance, declaringClass: v.DeclaringClass, evalFunc: v.EvalFunc,
	}
}

func (v *EvaluatingValue) Kind() ValueKind {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.operatorValueBase.Kind()
}
func (v *EvaluatingValue) KindString() string {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.operatorValueBase.KindString()
}
func (v *EvaluatingValue) GetCostModel() CostModel {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.operatorValueBase.GetCostModel()
}
func (v *EvaluatingValue) Compare(other Value) (int, error) { return v.operator().Compare(other) }
func (v *EvaluatingValue) Equal(other Value) (bool, error)  { return v.operator().Equal(other) }
func (v *EvaluatingValue) Member(elem Value) (bool, error)  { return v.operator().Member(elem) }
func (v *EvaluatingValue) IsFinite() (bool, error)          { return v.operator().IsFinite() }
func (v *EvaluatingValue) Size() (int, error)               { return v.operator().Size() }
func (v *EvaluatingValue) FingerPrint(fp uint64) uint64     { return v.operator().FingerPrint(fp) }
func (v *EvaluatingValue) TakeExcept(ex *ValueExcept) (Value, error) {
	return v.operator().TakeExcept(ex)
}
func (v *EvaluatingValue) TakeExcepts(exs []*ValueExcept) (Value, error) {
	return v.operator().TakeExcepts(exs)
}
func (v *EvaluatingValue) String() string { return v.operator().String() }
func (v *EvaluatingValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	if v == nil {
		panic(NewNullPointerException())
	}
	sb.WriteString("<Java Method: " + v.ownedMetadata().name + ">")
	return sb
}
func (v *EvaluatingValue) operator() *operatorValueBase {
	if v == nil {
		panic(NewNullPointerException())
	}
	return &v.operatorValueBase
}

func (v *EvaluatingValue) Eval(args []Value, control int) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v.receiver(), &err)
	return nil, NewWrongInvocationException("It is a TLC bug: Should use the other eval method.")
}

func (v *EvaluatingValue) EvalWithTool(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	control = int(int32(control))
	defer catchJavaMethodFailure(v.ownedMetadata().name, &resultValue, &err, false)
	if v.ownedMetadata().instance {
		return nil, evaluatingInstanceHandleError(v.ownedMetadata().declaringClass)
	}
	value, err := v.ownedMetadata().evalFunc(tool, args, con, state, pstate, control, cm)
	if err != nil || !isNil(value) {
		return value, err
	}
	return tool.EvalPure(v.ownedMetadata().opDef, args, con, state, pstate, control, cm)
}

func (v *EvaluatingValue) Initialize() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	v.DeepNormalize()
	return v.receiver()
}

func (v *EvaluatingValue) Normalize() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.operatorValueBase.Normalize()
}
func (v *EvaluatingValue) IsNormalized() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.operatorValueBase.IsNormalized()
}
func (v *EvaluatingValue) IsDefined() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.operatorValueBase.IsDefined()
}
func (v *EvaluatingValue) DeepCopy() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.receiver()
}
func (v *EvaluatingValue) DeepNormalize() {
	if v == nil {
		panic(NewNullPointerException())
	}
	v.operatorValueBase.DeepNormalize()
}
func (v *EvaluatingValue) Permute(perm *MVPerm) Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.operatorValueBase.Permute(perm)
}

func (v *EvaluatingValue) GetMinLevel() int {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.ownedMetadata().minLevel
}

func (v *EvaluatingValue) GetOpDef() *OpDefNode {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.ownedMetadata().opDef
}

// asEvaluatingValue matches Java's instanceof EvaluatingValue check for the
// concrete subclasses represented by embedding in Go.
func asEvaluatingValue(value any) (*EvaluatingValue, bool) {
	switch v := value.(type) {
	case *EvaluatingValue:
		return v, v != nil
	case *PriorityEvaluatingValue:
		return v.EvaluatingValue, v != nil && v.EvaluatingValue != nil
	case *CallableValue:
		return v.EvaluatingValue, v != nil && v.EvaluatingValue != nil
	default:
		return nil, false
	}
}

type PriorityEvaluatingValue struct {
	*EvaluatingValue
	Handles []*EvaluatingValue

	handlesCaptured bool
	handlesOwner    []*EvaluatingValue
}

func (v *PriorityEvaluatingValue) ownedHandles() []*EvaluatingValue {
	if v.handlesCaptured {
		return v.handlesOwner
	}
	return v.Handles
}

func (v *PriorityEvaluatingValue) captureHandles(handles []*EvaluatingValue) {
	v.Handles = handles
	v.handlesOwner = handles
	v.handlesCaptured = true
}

func (v *PriorityEvaluatingValue) require() *EvaluatingValue {
	if v == nil || v.EvaluatingValue == nil {
		panic(NewNullPointerException())
	}
	return v.EvaluatingValue
}

func (v *PriorityEvaluatingValue) Kind() ValueKind { return v.require().Kind() }
func (v *PriorityEvaluatingValue) KindString() string {
	return v.require().KindString()
}
func (v *PriorityEvaluatingValue) GetCostModel() CostModel {
	return v.require().GetCostModel()
}
func (v *PriorityEvaluatingValue) Normalize() Value { return v.require().Normalize() }
func (v *PriorityEvaluatingValue) DeepNormalize()   { v.require().DeepNormalize() }
func (v *PriorityEvaluatingValue) IsNormalized() bool {
	return v.require().IsNormalized()
}
func (v *PriorityEvaluatingValue) IsDefined() bool { return v.require().IsDefined() }
func (v *PriorityEvaluatingValue) DeepCopy() Value { return v.require().DeepCopy() }
func (v *PriorityEvaluatingValue) Permute(perm *MVPerm) Value {
	return v.require().Permute(perm)
}
func (v *PriorityEvaluatingValue) Compare(other Value) (int, error) {
	return v.require().Compare(other)
}
func (v *PriorityEvaluatingValue) Equal(other Value) (bool, error) {
	return v.require().Equal(other)
}
func (v *PriorityEvaluatingValue) Member(elem Value) (bool, error) {
	return v.require().Member(elem)
}
func (v *PriorityEvaluatingValue) IsFinite() (bool, error) {
	return v.require().IsFinite()
}
func (v *PriorityEvaluatingValue) Size() (int, error) { return v.require().Size() }
func (v *PriorityEvaluatingValue) FingerPrint(fp uint64) uint64 {
	return v.require().FingerPrint(fp)
}
func (v *PriorityEvaluatingValue) TakeExcept(ex *ValueExcept) (Value, error) {
	return v.require().TakeExcept(ex)
}
func (v *PriorityEvaluatingValue) TakeExcepts(exs []*ValueExcept) (Value, error) {
	return v.require().TakeExcepts(exs)
}
func (v *PriorityEvaluatingValue) String() string { return v.require().String() }
func (v *PriorityEvaluatingValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	return v.require().ToString(sb, offset, swallow)
}

func NewPriorityEvaluatingValue(primary *EvaluatingValue, secondary *EvaluatingValue) *PriorityEvaluatingValue {
	out := &PriorityEvaluatingValue{}
	out.captureHandles(nil)
	out.Add(primary)
	out.Add(secondary)
	return out
}

func (v *PriorityEvaluatingValue) Add(ev *EvaluatingValue) {
	if v == nil {
		panic(NewNullPointerException())
	}
	if ev == nil {
		panic(NewNullPointerException())
	}
	if v.EvaluatingValue != nil && (v.ownedMetadata().opDef != ev.ownedMetadata().opDef || v.ownedMetadata().minLevel != ev.ownedMetadata().minLevel) {
		panic(NewTLCRuntimeException(ECGeneral))
	}
	if v.EvaluatingValue == nil {
		// Java constructs a new wrapper from the primary method and adds this
		// wrapper itself as a handle. Sorting never changes its method metadata.
		primary := *ev
		primary.BaseValue = BaseValue{}
		v.EvaluatingValue = &primary
		v.owner = v
		ev = v.EvaluatingValue
	}
	v.captureHandles(append(v.ownedHandles(), ev))
	for i := 1; i < len(v.ownedHandles()); i++ {
		cur := v.ownedHandles()[i]
		j := i
		for j > 0 && v.ownedHandles()[j-1].ownedMetadata().priority > cur.ownedMetadata().priority {
			v.ownedHandles()[j] = v.ownedHandles()[j-1]
			j--
		}
		v.ownedHandles()[j] = cur
	}
}

func (v *PriorityEvaluatingValue) EvalWithTool(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (resultValue Value, err error) {
	v.require()
	control = int(int32(control))
	defer catchJavaMethodFailure(v.ownedMetadata().name, &resultValue, &err, false)
	for _, ev := range v.ownedHandles() {
		if ev.ownedMetadata().instance {
			return nil, evaluatingInstanceHandleError(ev.ownedMetadata().declaringClass)
		}
		value, err := ev.ownedMetadata().evalFunc(tool, args, con, state, pstate, control, cm)
		if err != nil || !isNil(value) {
			return value, err
		}
	}
	return tool.Eval(v.ownedMetadata().opDef.GetBody(), con, state, pstate, control, cm)
}

func evaluatingInstanceHandleError(declaringClass string) error {
	return fmt.Errorf("cannot convert MethodHandle(%s,Tool,ExprOrOpArgNode[],Context,TLCState,TLCState,int,CostModel)Value to (Tool,ExprOrOpArgNode[],Context,TLCState,TLCState,int,CostModel)Object", declaringClass)
}

func (v *PriorityEvaluatingValue) Eval(args []Value, control int) (Value, error) {
	return v.require().Eval(args, control)
}

type CallableValue struct {
	*EvaluatingValue
	CallableFunc CallableEvalFunc

	callableCaptured bool
	callableOwner    CallableEvalFunc
}

func (v *CallableValue) ownedCallable() CallableEvalFunc {
	if v.callableCaptured {
		return v.callableOwner
	}
	return v.CallableFunc
}

func (v *CallableValue) require() *EvaluatingValue {
	if v == nil || v.EvaluatingValue == nil {
		panic(NewNullPointerException())
	}
	return v.EvaluatingValue
}

func (v *CallableValue) Kind() ValueKind { return v.require().Kind() }
func (v *CallableValue) KindString() string {
	return v.require().KindString()
}
func (v *CallableValue) GetCostModel() CostModel {
	return v.require().GetCostModel()
}
func (v *CallableValue) Normalize() Value { return v.require().Normalize() }
func (v *CallableValue) DeepNormalize()   { v.require().DeepNormalize() }
func (v *CallableValue) IsNormalized() bool {
	return v.require().IsNormalized()
}
func (v *CallableValue) IsDefined() bool { return v.require().IsDefined() }
func (v *CallableValue) DeepCopy() Value { return v.require().DeepCopy() }
func (v *CallableValue) Permute(perm *MVPerm) Value {
	return v.require().Permute(perm)
}
func (v *CallableValue) Compare(other Value) (int, error) {
	return v.require().Compare(other)
}
func (v *CallableValue) Equal(other Value) (bool, error) {
	return v.require().Equal(other)
}
func (v *CallableValue) Member(elem Value) (bool, error) {
	return v.require().Member(elem)
}
func (v *CallableValue) IsFinite() (bool, error) { return v.require().IsFinite() }
func (v *CallableValue) Size() (int, error)      { return v.require().Size() }
func (v *CallableValue) FingerPrint(fp uint64) uint64 {
	return v.require().FingerPrint(fp)
}
func (v *CallableValue) TakeExcept(ex *ValueExcept) (Value, error) {
	return v.require().TakeExcept(ex)
}
func (v *CallableValue) TakeExcepts(exs []*ValueExcept) (Value, error) {
	return v.require().TakeExcepts(exs)
}
func (v *CallableValue) String() string { return v.require().String() }
func (v *CallableValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	return v.require().ToString(sb, offset, swallow)
}

func NewCallableValue(name string, minLevel int, opDef *OpDefNode, callable CallableEvalFunc) *CallableValue {
	out := &CallableValue{
		EvaluatingValue: newEvaluatingValue(name, minLevel, 100, opDef, nil),
		CallableFunc:    callable,
	}
	out.owner = out
	out.callableCaptured, out.callableOwner = true, callable
	return out
}

func (v *CallableValue) EvalWithTool(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (resultValue Value, err error) {
	v.require()
	control = int(int32(control))
	if args == nil {
		panic(NewNullPointerException())
	}
	argVals := make([]Value, len(args))
	for i, arg := range args {
		value, err := tool.Eval(arg, con, state, pstate, control, cm)
		if err != nil {
			return nil, err
		}
		argVals[i] = value
	}
	defer catchJavaMethodFailure(v.ownedMetadata().name, &resultValue, &err, false)
	if v.ownedMetadata().instance {
		return nil, fmt.Errorf("cannot convert MethodHandle(%s,IValue[])Callable to (Value[])Callable", v.ownedMetadata().declaringClass)
	}
	callable, err := v.ownedCallable()(argVals)
	if err != nil {
		return nil, err
	}
	if pstate == nil {
		panic(NewNullPointerException())
	}
	pstate.SetCallable(callable)
	return BoolTrue, nil
}

func (v *CallableValue) Eval(args []Value, control int) (Value, error) {
	return v.require().Eval(args, control)
}

func WithEvaluatingOpDef(value any, opDef *OpDefNode) any {
	if opDef == nil {
		return value
	}
	switch v := value.(type) {
	case *EvaluatingValue:
		return v.withOpDef(opDef)
	case *PriorityEvaluatingValue:
		out := &PriorityEvaluatingValue{EvaluatingValue: v.EvaluatingValue.withOpDef(opDef)}
		out.owner = out
		for _, handle := range v.ownedHandles() {
			if handle == v.EvaluatingValue {
				out.Add(out.EvaluatingValue)
			} else {
				out.Add(handle.withOpDef(opDef))
			}
		}
		return out
	case *CallableValue:
		return v.withOpDef(opDef)
	default:
		return value
	}
}

func (v *EvaluatingValue) withOpDef(opDef *OpDefNode) *EvaluatingValue {
	if v == nil || opDef == nil {
		return v
	}
	out := *v
	out.OpDef = opDef
	out.owner = &out
	out.captureMetadata()
	return &out
}

func (v *CallableValue) withOpDef(opDef *OpDefNode) *CallableValue {
	if v == nil || opDef == nil {
		return v
	}
	out := *v
	out.EvaluatingValue = v.EvaluatingValue.withOpDef(opDef)
	out.owner = &out
	return &out
}

func EvalOperatorValue(op Value, args []Value, control int) (Value, error) {
	control = int(int32(control))
	if operator, ok := op.(OperatorValue); ok {
		return operator.Eval(args, control)
	}
	return nil, newTLCError(ECGeneral, "Attempted to apply non-operator value %s.", op)
}

func EvalOperatorValueWithTool(op Value, tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
	control = int(int32(control))
	if operator, ok := op.(OperatorValue); ok {
		return operator.EvalWithTool(tool, args, con, state, pstate, control, cm)
	}
	if args == nil {
		panic(NewNullPointerException())
	}
	argVals := make([]Value, len(args))
	for i, arg := range args {
		value, err := tool.Eval(arg, con, state, pstate, control, cm)
		if err != nil {
			return nil, err
		}
		argVals[i] = value
	}
	return EvalOperatorValue(op, argVals, control)
}
