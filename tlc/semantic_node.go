package tlc

import (
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

// FrontEnd.getToolId returns the prior value of its process-wide Java int.
var nextSemanticToolID atomic.Int32

func GetSemanticToolID() int32 {
	return nextSemanticToolID.Add(1) - 1
}

type SemanticNode any

type SemanticKind int

const (
	TLCLevelConstant = iota
	TLCLevelState
	TLCLevelAction
	TLCLevelTemporal
)

const (
	SemanticConstantLevel = TLCLevelConstant
	SemanticVariableLevel = TLCLevelState
	SemanticActionLevel   = TLCLevelAction
	SemanticTemporalLevel = TLCLevelTemporal
	SemanticMinLevel      = SemanticConstantLevel
	SemanticMaxLevel      = SemanticTemporalLevel
)

func semanticLevelPointer(level int) *int { return &level }

// SemanticLevels exposes LevelConstants.Levels as a live nullable boxed array.
var SemanticLevels = []*int{
	semanticLevelPointer(SemanticConstantLevel),
	semanticLevelPointer(SemanticVariableLevel),
	semanticLevelPointer(SemanticActionLevel),
	semanticLevelPointer(SemanticTemporalLevel),
}

func SemanticLevelToString(level int) string {
	level = int(int32(level))
	name := "Illegal"
	switch level {
	case TLCLevelConstant:
		name = "Constant"
	case TLCLevelState:
		name = "Variable"
	case TLCLevelAction:
		name = "Action"
	case TLCLevelTemporal:
		name = "Temporal"
	}
	return strconv.Itoa(level) + " (" + name + ")"
}

const (
	SemanticUnknownKind           SemanticKind = 0
	SemanticModuleKind            SemanticKind = 1
	SemanticConstantDeclKind      SemanticKind = 2
	SemanticVariableDeclKind      SemanticKind = 3
	SemanticBoundSymbolKind       SemanticKind = 4
	SemanticUserDefinedOpKind     SemanticKind = 5
	SemanticModuleInstanceKind    SemanticKind = 6
	SemanticBuiltInKind           SemanticKind = 7
	SemanticOpArgKind             SemanticKind = 8
	SemanticOpApplKind            SemanticKind = 9
	SemanticLetInKind             SemanticKind = 10
	SemanticFormalParamKind       SemanticKind = 11
	SemanticTheoremKind           SemanticKind = 12
	SemanticSubstInKind           SemanticKind = 13
	SemanticAssumeProveKind       SemanticKind = 14
	SemanticNumeralKind           SemanticKind = 16
	SemanticDecimalKind           SemanticKind = 17
	SemanticStringKind            SemanticKind = 18
	SemanticAtNodeKind            SemanticKind = 19
	SemanticAssumeKind            SemanticKind = 20
	SemanticInstanceKind          SemanticKind = 21
	SemanticNewSymbKind           SemanticKind = 22
	SemanticThmOrAssumpKind       SemanticKind = 23
	SemanticNewConstantKind       SemanticKind = 24
	SemanticNewVariableKind       SemanticKind = 25
	SemanticNewStateKind          SemanticKind = 26
	SemanticNewActionKind         SemanticKind = 27
	SemanticNewTemporalKind       SemanticKind = 28
	SemanticLabelKind             SemanticKind = 29
	SemanticAPSubstInKind         SemanticKind = 30
	SemanticUseKind               SemanticKind = 31
	SemanticHideKind              SemanticKind = 32
	SemanticLeafProofKind         SemanticKind = 33
	SemanticNonLeafProofKind      SemanticKind = 34
	SemanticQEDStepKind           SemanticKind = 35
	SemanticDefStepKind           SemanticKind = 36
	SemanticNumberedProofStepKind SemanticKind = 37
	SemanticValueKind             SemanticKind = 1001
	SemanticPossibleTrackKind     SemanticKind = 1002
	SemanticPossibleCheckKind     SemanticKind = 1003
)

type SemanticNodeBase struct {
	KindValue          SemanticKind
	uidPlusOne         int32
	uidInitialized     uint32
	indexedTools       *semanticNodeToolSlots
	ToolObject         any
	Image              string
	LevelValue         int
	LevelChecked       int32
	CanonicalLevelData any // Parser-owned canonical collections on this same node.
	LevelParamSet      []*SymbolNode
	Location           SourceLocation
	TreeNode           any // The production SANY syntax node, owned by the parser package.
	syntaxLocation     func(any) SourceLocation
}

var nextSemanticNodeUID atomic.Int32
var semanticNodeUIDInitialization sync.Mutex

// Each semantic node owns its array. The lock preserves Go's concurrent cache
// access while array growth, retained slots and bounds follow SemanticNode.
type semanticNodeToolSlots struct {
	sync.RWMutex
	values []any
}

func newSemanticNodeBasePointer(kind SemanticKind, image string) *SemanticNodeBase {
	base := NewSemanticNodeBase(kind, image)
	return &base
}

// NewSemanticNodeBase allocates SemanticNode's process-wide Java int UID.
// SANY context symbols and TLC evaluator nodes share this constructor.
func NewSemanticNodeBase(kind SemanticKind, image string) SemanticNodeBase {
	return SemanticNodeBase{KindValue: kind, uidPlusOne: nextSemanticNodeUID.Add(1), uidInitialized: 1, indexedTools: &semanticNodeToolSlots{}, Image: image}
}

// SemanticNode.nullSN retains the builtin syntax/location used by Action.UNKNOWN.
type NullSemanticNode struct{ SemanticNodeBase }

type nullSemanticSyntax struct{}

func (nullSemanticSyntax) GetHumanReadableImage() string       { return "***I do not exist***" }
func (nullSemanticSyntax) GetAttachedComments() []string       { return []string{} }
func (nullSemanticSyntax) GetOneHumanReadableImages() []string { return []string{} }
func (nullSemanticSyntax) GetLocation() SourceLocation {
	return NewSourceLocation("--TLA+ BUILTINS--", 0, 0, 0, 0)
}

func (n *NullSemanticNode) LevelDataToString() string { return "-2147483648" }

var NullSemanticNodeInstance = func() *NullSemanticNode {
	base := NewSemanticNodeBase(SemanticKind(-2147483648), "***I do not exist***")
	base.Location = NewSourceLocation("--TLA+ BUILTINS--", 0, 0, 0, 0)
	base.TreeNode = nullSemanticSyntax{}
	return &NullSemanticNode{base}
}()

func (n *SemanticNodeBase) GetTreeNode() any {
	n.require()
	return n.TreeNode
}
func (n *SemanticNodeBase) SetTreeNode(node any) {
	n.require()
	n.TreeNode = node
}

func (n *SemanticNodeBase) GetHumanReadableImage() string {
	n.require()
	return n.SourceLocation().String()
}

// Parser-owned bases read the current syntax using the parser's TreeNode
// contract. Native evaluator nodes retain their explicit Location field.
func (n *SemanticNodeBase) SetSyntaxLocationResolver(resolve func(any) SourceLocation) {
	n.require()
	n.syntaxLocation = resolve
}

func (n *SemanticNodeBase) hasSourceSyntax() bool {
	return n != nil && n.syntaxLocation != nil
}

func (n *SemanticNodeBase) hasLocationSyntax() bool {
	if n == nil {
		return false
	}
	if n.syntaxLocation != nil {
		return true
	}
	if semanticExploreNull(n.TreeNode) {
		return false
	}
	_, ok := n.TreeNode.(interface{ GetLocation() SourceLocation })
	return ok
}

// SemanticNode.toString normally prints the location. Numeral and Decimal
// override it, while showPlainFormulae selects the actual SANY syntax image.
// SemanticNodeJavaString exposes the source zero-argument formatting to SANY.
func SemanticNodeJavaString(node SemanticNode) string { return semanticNodeJavaString(node) }

// SemanticNodeAppendString ports SemanticNode.toString(StringBuffer, String).
// The second argument is intentionally unused by the source. Unlike the
// zero-argument virtual method, this final overload does not use literal-node
// rendering overrides.
func SemanticNodeAppendString(node SemanticNode, buffer *strings.Builder, ignored string) {
	_ = ignored
	if semanticExploreNull(node) {
		panic(NewNullPointerException())
	}
	rendered, renderedSyntax := "", false
	if _, present := tlcLookupSystemProperty("tla2sany.semantic.SemanticNode.showPlainFormulae"); present {
		if treeOwner, ok := node.(interface{ GetTreeNode() any }); ok {
			if tree, ok := treeOwner.GetTreeNode().(interface{ GetHumanReadableImage() string }); ok {
				rendered = tree.GetHumanReadableImage()
				renderedSyntax = true
			}
		}
	}
	if !renderedSyntax {
		location, ok := semanticNodeSourceLocation(node)
		if !ok {
			panic(NewClassCastException("object is not a SemanticNode"))
		}
		rendered = location.String()
	}
	if buffer == nil {
		panic(NewNullPointerException())
	}
	buffer.WriteString(rendered)
}

func semanticNodeJavaString(node SemanticNode) string {
	if semanticExploreNull(node) {
		panic(NewNullPointerException())
	}
	switch n := node.(type) {
	case *NumeralNode:
		return n.String()
	case *DecimalNode:
		return n.String()
	}
	if _, present := tlcLookupSystemProperty("tla2sany.semantic.SemanticNode.showPlainFormulae"); present {
		if n, ok := node.(interface{ GetTreeNode() any }); ok {
			if tree, ok := n.GetTreeNode().(interface{ GetHumanReadableImage() string }); ok {
				return tree.GetHumanReadableImage()
			}
		}
	}
	if location, ok := semanticNodeSourceLocation(node); ok {
		return location.String()
	}
	return toContextString(node)
}

func (n *SemanticNodeBase) Kind() SemanticKind {
	n.require()
	return n.KindValue
}

func (n *SemanticNodeBase) SetKind(kind SemanticKind) {
	n.require()
	n.KindValue = kind
}

// GetPreComments returns the syntax node's attached comment array.  Java's
// SemanticNode casts stn to SyntaxTreeNode on every call, so a null syntax node
// and a syntax node of the wrong class retain their distinct runtime failures.
func (n *SemanticNodeBase) GetPreComments() []string {
	n.require()
	tree := n.GetTreeNode()
	if semanticExploreNull(tree) {
		panic(NewNullPointerException())
	}
	comments, ok := tree.(interface{ GetAttachedComments() []string })
	if !ok {
		panic(valueStreamClassCast(tree, "tla2sany.parser.SyntaxTreeNode"))
	}
	return comments.GetAttachedComments()
}

func (n *SemanticNodeBase) GetPreCommentsAsString() string {
	n.require()
	comments := n.GetPreComments()
	if len(comments) == 0 {
		return ""
	}
	var out strings.Builder
	out.WriteString("\n preComment: ")
	for i, comment := range comments {
		if i != 0 {
			out.WriteString("\n             ")
		}
		out.WriteString(strconv.Itoa(i))
		out.WriteByte(' ')
		out.WriteString(comment)
	}
	return out.String()
}

func (n *SemanticNodeBase) GetUid() int32 {
	n.require()
	// Actual semantic constructors eagerly assign identity. Retain Go's existing
	// zero-value base support independently of the UID bits: uidPlusOne == 0
	// represents the valid Java UID -1 after signed wraparound.
	if atomic.LoadUint32(&n.uidInitialized) == 0 {
		semanticNodeUIDInitialization.Lock()
		if atomic.LoadUint32(&n.uidInitialized) == 0 {
			atomic.StoreInt32(&n.uidPlusOne, nextSemanticNodeUID.Add(1))
			n.indexedTools = &semanticNodeToolSlots{}
			atomic.StoreUint32(&n.uidInitialized, 1)
		}
		semanticNodeUIDInitialization.Unlock()
	}
	return atomic.LoadInt32(&n.uidPlusOne) - 1
}

func (n *SemanticNodeBase) GetUID() int32 {
	return n.GetUid()
}

func (n *SemanticNodeBase) JavaHashCode() int32 {
	n.require()
	result := int32(1)
	result = 31*result + int32(n.Kind())
	result = 31*result + n.GetUid()
	return result
}

func (n *SemanticNodeBase) GetToolObject() any {
	n.require()
	return n.ToolObject
}

func (n *SemanticNodeBase) SetToolObject(value any) {
	n.require()
	n.ToolObject = value
}

func (n *SemanticNodeBase) toolSlots() *semanticNodeToolSlots {
	if atomic.LoadUint32(&n.uidInitialized) == 0 {
		n.GetUid()
	}
	return n.indexedTools
}

// GetToolObjectAt ports SemanticNode.getToolObject(toolId). Missing nonnegative
// slots return null; negative indices throw even when the array is empty.
func (n *SemanticNodeBase) GetToolObjectAt(toolID int32) any {
	n.require()
	slots := n.toolSlots()
	slots.RLock()
	defer slots.RUnlock()
	if int64(len(slots.values)) <= int64(toolID) {
		return nil
	}
	if toolID < 0 {
		panic(NewArrayIndexOutOfBoundsException(int(toolID), len(slots.values)))
	}
	return slots.values[int(toolID)]
}

// SetToolObjectAt grows through the requested index and preserves all earlier
// slots. Storing null still grows the source array and does not shrink it.
func (n *SemanticNodeBase) SetToolObjectAt(toolID int32, object any) {
	n.require()
	slots := n.toolSlots()
	slots.Lock()
	defer slots.Unlock()
	if int64(len(slots.values)) <= int64(toolID) {
		length := toolID + 1 // Java int addition wraps before array allocation.
		if length < 0 {
			panic(NewNegativeArraySizeException(strconv.FormatInt(int64(length), 10)))
		}
		values := make([]any, int(length))
		copy(values, slots.values)
		slots.values = values
	}
	if toolID < 0 {
		panic(NewArrayIndexOutOfBoundsException(int(toolID), len(slots.values)))
	}
	slots.values[int(toolID)] = object
}

func (n *SemanticNodeBase) String() string {
	n.require()
	if n.hasLocationSyntax() {
		return semanticNodeJavaString(n)
	}
	if !n.Location.IsNull() {
		return n.Location.String()
	}
	if n.Image == "" {
		return "<semantic node>"
	}
	return n.Image
}

func (n *SemanticNodeBase) SourceLocation() SourceLocation {
	n.require()
	if n.syntaxLocation != nil {
		return n.syntaxLocation(n.TreeNode)
	}
	if !semanticExploreNull(n.TreeNode) {
		if tree, ok := n.TreeNode.(interface{ GetLocation() SourceLocation }); ok {
			return tree.GetLocation()
		}
	}
	return n.Location
}

func (n *SemanticNodeBase) GetSourceLocation() SourceLocation {
	return n.SourceLocation()
}

func (n *SemanticNodeBase) SetSourceLocation(location SourceLocation) {
	n.require()
	n.Location = location
}

// SemanticNode.isStandardModule delegates to Java's StandardModules name set.
// It deliberately classifies by source module name, including replaced modules.
func (n *SemanticNodeBase) IsStandardModule() bool {
	n.require()
	return IsDefinedInStandardModule(n)
}

func (n *SemanticNodeBase) GetLevel() int {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface{ RuntimeCheckedLevel() int }); ok {
		return provider.RuntimeCheckedLevel()
	}
	return n.LevelValue
}

func (n *SemanticNodeBase) SetLevel(level int) {
	n.require()
	n.LevelValue = level
}

// GetCanonicalLevelData exposes parser-owned level metadata to the root SANY
// adapter without making package tlc depend on its concrete collection types.
func (n *SemanticNodeBase) GetCanonicalLevelData() any {
	n.require()
	return n.CanonicalLevelData
}

func (n *SemanticNodeBase) GetLevelParams() []*SymbolNode {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface{ RuntimeCheckedLevelParams() []*SymbolNode }); ok {
		return provider.RuntimeCheckedLevelParams()
	}
	if len(n.LevelParamSet) == 0 {
		return nil
	}
	out := make([]*SymbolNode, len(n.LevelParamSet))
	copy(out, n.LevelParamSet)
	return out
}

func (n *SemanticNodeBase) SetLevelParams(params ...*SymbolNode) {
	n.require()
	n.LevelParamSet = nil
	n.AddLevelParams(params...)
}

func (n *SemanticNodeBase) AddLevelParams(params ...*SymbolNode) {
	n.require()
	for _, param := range params {
		if param == nil || n.hasLevelParam(param) {
			continue
		}
		n.LevelParamSet = append(n.LevelParamSet, param)
	}
}

func (n *SemanticNodeBase) hasLevelParam(param *SymbolNode) bool {
	for _, existing := range n.LevelParamSet {
		if existing == param {
			return true
		}
	}
	return false
}

func (n *SemanticNodeBase) require() {
	if n == nil {
		panic(NewNullPointerException())
	}
}

func SemanticLevel(node SemanticNode) int {
	switch n := node.(type) {
	case nil:
		return TLCLevelConstant
	case *AssumeNode:
		return n.GetLevel()
	case *ModuleNode:
		return n.GetLevel()
	case *OpDefNode:
		return n.GetLevel()
	case *LabelNode:
		return n.GetLevel()
	case *OpApplNode:
		return n.GetLevel()
	case *LetInNode:
		return n.GetLevel()
	case *SubstInNode:
		return n.GetLevel()
	case *APSubstInNode:
		return n.GetLevel()
	case *ValueNode:
		return n.GetLevel()
	case *NumeralNode:
		return n.GetLevel()
	case *DecimalNode:
		return n.GetLevel()
	case *StringNode:
		return n.GetLevel()
	case *AtNode:
		return n.GetLevel()
	case *OpArgNode:
		return n.GetLevel()
	case *LiveExprNode:
		return n.GetLevel()
	case *PossibleTrackNode:
		return n.GetLevel()
	case *PossibleCheckNode:
		return n.GetLevel()
	default:
		return TLCLevelConstant
	}
}

func SemanticLevelParams(node SemanticNode) []*SymbolNode {
	switch n := node.(type) {
	case nil:
		return nil
	case *AssumeNode:
		return n.GetLevelParams()
	case *OpDefNode:
		return n.GetLevelParams()
	case *LabelNode:
		return n.GetLevelParams()
	case *OpApplNode:
		return n.GetLevelParams()
	case *LetInNode:
		return n.GetLevelParams()
	case *SubstInNode:
		return n.GetLevelParams()
	case *APSubstInNode:
		return n.GetLevelParams()
	case *ValueNode:
		return n.GetLevelParams()
	case *NumeralNode:
		return n.GetLevelParams()
	case *DecimalNode:
		return n.GetLevelParams()
	case *StringNode:
		return n.GetLevelParams()
	case *AtNode:
		return n.GetLevelParams()
	case *OpArgNode:
		return n.GetLevelParams()
	case *PossibleTrackNode:
		return n.GetLevelParams()
	case *PossibleCheckNode:
		return n.GetLevelParams()
	default:
		return nil
	}
}

type LabelNode struct {
	*SemanticNodeBase
	Name            *UniqueString
	Arity           int
	Body            SemanticNode
	Goal            SemanticNode
	Params          []*SymbolNode
	Labels          *SemanticLabelTable
	IsAssumeProve   bool
	SubExpressionOf *SymbolNode
	GoalClause      int
}

func NewLabelNode(body SemanticNode) *LabelNode {
	return &LabelNode{SemanticNodeBase: newSemanticNodeBasePointer(SemanticLabelKind, "label"), Body: body}
}

func (n *LabelNode) require() {
	if n == nil {
		panic(NewNullPointerException())
	}
}

func (n *LabelNode) GetName() *UniqueString {
	n.require()
	return n.Name
}

func (n *LabelNode) GetArity() int {
	n.require()
	return n.Arity
}

func (n *LabelNode) GetBody() SemanticNode {
	n.require()
	return n.Body
}

func (n *LabelNode) GetGoal() SemanticNode {
	n.require()
	return n.Goal
}

func (n *LabelNode) SetLabels(labels *SemanticLabelTable) {
	n.require()
	setSemanticNodeLabels(n.SemanticNodeBase, labels)
	n.Labels = labels
}

func (n *LabelNode) GetLabelsHT() *SemanticLabelTable {
	n.require()
	return semanticNodeLabels(n.SemanticNodeBase, n.Labels)
}

func (n *LabelNode) GetLabel(name *UniqueString) *LabelNode {
	n.require()
	labels := n.GetLabelsHT()
	if labels == nil {
		return nil
	}
	return labels.Get(name)
}

func (n *LabelNode) AddLabel(label *LabelNode) bool {
	n.require()
	labels := n.GetLabelsHT()
	if labels == nil {
		n.SetLabels(NewSemanticLabelTable())
		labels = n.GetLabelsHT()
	}
	return labels.Add(label)
}

func (n *LabelNode) GetLabels() []*LabelNode {
	n.require()
	labels := n.GetLabelsHT()
	if labels == nil {
		return make([]*LabelNode, 0)
	}
	return labels.Elements()
}

func (n *LabelNode) GetChildren() []SemanticNode {
	n.require()
	return []SemanticNode{n.Body}
}

type OpApplNode struct {
	*SemanticNodeBase
	Operator             *SymbolNode
	SubExpressionOf      *SymbolNode
	Args                 []SemanticNode
	BdedQuantSymbolLists [][]*SymbolNode
	BdedQuantBounds      []SemanticNode
	BdedQuantATuple      []bool
	UnbdedQuantSymbols   []*SymbolNode
}

func NewOpApplNode(operator *SymbolNode, args ...SemanticNode) *OpApplNode {
	out := &OpApplNode{
		SemanticNodeBase: newSemanticNodeBasePointer(SemanticOpApplKind, ""),
		Operator:         operator,
		Args:             append([]SemanticNode{}, args...),
		BdedQuantBounds:  []SemanticNode{},
	}
	out.Image = out.String()
	return out
}

func NewBuiltinOpApplNode(op *UniqueString, args ...SemanticNode) *OpApplNode {
	ensureBuiltInOPs()
	if op != nil {
		op = UniqueStringOf(op.String())
	}
	return NewOpApplNode(&SymbolNode{Name: op}, args...)
}

func (n *OpApplNode) require() {
	if n == nil {
		panic(NewNullPointerException())
	}
}

func (n *OpApplNode) GetOperator() *SymbolNode {
	n.require()
	return n.Operator
}

func (n *OpApplNode) ResetOperator(definition *OpDefNode) {
	n.require()
	if definition == nil {
		n.Operator = nil
		return
	}
	n.Operator = definition.Symbol
}

func (n *OpApplNode) GetArgs() []SemanticNode {
	n.require()
	return n.Args
}

func (n *OpApplNode) SetArgs(args []SemanticNode) {
	n.require()
	n.Args = args
}

func (n *OpApplNode) ArgsContainOpArgNodes() bool {
	n.require()
	if n.Args == nil {
		panic(NewNullPointerException())
	}
	for _, arg := range n.Args {
		if node, ok := arg.(*OpArgNode); ok && node != nil {
			return true
		}
	}
	return false
}

func (n *OpApplNode) GetUnbdedQuantSymbols() []*SymbolNode {
	n.require()
	return n.UnbdedQuantSymbols
}

func (n *OpApplNode) GetBdedQuantSymbolLists() [][]*SymbolNode {
	n.require()
	return n.BdedQuantSymbolLists
}

func (n *OpApplNode) IsBdedQuantATuple() []bool {
	n.require()
	return n.BdedQuantATuple
}

func (n *OpApplNode) GetBdedQuantBounds() []SemanticNode {
	n.require()
	return n.BdedQuantBounds
}

func (n *OpApplNode) HasOpcode(opcode int) bool {
	n.require()
	opcode = int(int32(opcode))
	return opcode == GetOpCode(n.Operator.GetName())
}

func (n *OpApplNode) String() string {
	n.require()
	if n.hasLocationSyntax() {
		return semanticNodeJavaString(n)
	}
	if !n.Location.IsNull() {
		return n.Location.String()
	}
	if n.Operator == nil {
		return "<operator application>"
	}
	if n.Image != "" && n.Image != "<operator application>" {
		return n.Image
	}
	return n.Operator.String()
}

type LetInNode struct {
	*SemanticNodeBase
	Lets      []*OpDefNode
	OpDefs    []SemanticNode
	Instances []SemanticNode
	Bindings  []LetBinding
	Body      SemanticNode
	Context   *SemanticContext
}

type LetBinding struct {
	Symbol *SymbolNode
	Value  any
}

func NewLetInNode(body SemanticNode, lets ...*OpDefNode) *LetInNode {
	return NewLetInNodeWithBase(nil, body, lets...)
}

// Parser adapters retain the existing SANY node's UID, syntax and indexed tool
// slots. Standalone runtime LET nodes allocate their own semantic identity.
func NewLetInNodeWithBase(base *SemanticNodeBase, body SemanticNode, lets ...*OpDefNode) *LetInNode {
	if base == nil {
		owned := NewSemanticNodeBase(SemanticLetInKind, "LET")
		base = &owned
	}
	node := &LetInNode{
		SemanticNodeBase: base,
		Lets:             append([]*OpDefNode{}, lets...),
		Instances:        []SemanticNode{},
		Body:             body,
	}
	node.OpDefs = make([]SemanticNode, len(lets))
	for i, definition := range lets {
		node.OpDefs[i] = definition
	}
	return node
}

func (n *LetInNode) require() {
	if n == nil {
		panic(NewNullPointerException())
	}
}

func (n *LetInNode) GetLets() []*OpDefNode {
	n.require()
	return n.Lets
}

func (n *LetInNode) GetBody() SemanticNode {
	n.require()
	return n.Body
}

func (n *LetInNode) GetChildren() []SemanticNode {
	n.require()
	if n.OpDefs == nil || n.Instances == nil {
		panic(NewNullPointerException())
	}
	children := make([]SemanticNode, 0, len(n.OpDefs)+len(n.Instances)+1)
	children = append(children, n.OpDefs...)
	children = append(children, n.Instances...)
	return append(children, n.Body)
}

type Subst struct {
	*SubstFields
	id uint64
}

// Copies of a substitution reference one record, as Java Subst[] entries do.
type SubstFields struct {
	Op         *SymbolNode
	Expr       SemanticNode
	ExprSyntax any
	Implicit   bool
	runtime    SubstRuntime
}

type SubstRuntime interface {
	RuntimeSubstOp() (*SymbolNode, bool)
	RuntimeSetSubstOp(*SymbolNode) bool
	RuntimeSubstExpr() (SemanticNode, bool)
	RuntimeSetSubstExpr(SemanticNode, bool) bool
	RuntimeSubstExprSyntax() (any, bool)
	RuntimeSetSubstExprSyntax(any) bool
	RuntimeSubstImplicit() (bool, bool)
}

func NewSubst(op *SymbolNode, expr SemanticNode) Subst {
	return Subst{SubstFields: &SubstFields{Op: op, Expr: expr}, id: nextSubstID.Add(1)}
}

func (s *Subst) require() *SubstFields {
	if s == nil {
		panic(NewNullPointerException())
	}
	if s.SubstFields == nil {
		s.SubstFields = &SubstFields{}
	}
	return s.SubstFields
}

func (s *Subst) GetOp() *SymbolNode {
	fields := s.require()
	if fields.runtime != nil {
		if op, present := fields.runtime.RuntimeSubstOp(); present {
			return op
		}
	}
	return fields.Op
}

func (s *Subst) SetOp(op *SymbolNode) {
	fields := s.require()
	if fields.runtime != nil {
		fields.runtime.RuntimeSetSubstOp(op)
	}
	fields.Op = op
}

func (s *Subst) GetExpr() SemanticNode {
	fields := s.require()
	if fields.runtime != nil {
		if expr, present := fields.runtime.RuntimeSubstExpr(); present {
			return expr
		}
	}
	return fields.Expr
}

func (s *Subst) SetExpr(expr SemanticNode, implicit bool) {
	fields := s.require()
	if fields.runtime != nil {
		fields.runtime.RuntimeSetSubstExpr(expr, implicit)
	}
	fields.Expr = expr
	fields.Implicit = implicit
}

func (s *Subst) GetExprSTN() any {
	fields := s.require()
	if fields.runtime != nil {
		if syntax, present := fields.runtime.RuntimeSubstExprSyntax(); present {
			return syntax
		}
	}
	return fields.ExprSyntax
}

func (s *Subst) SetExprSTN(syntax any) {
	fields := s.require()
	if fields.runtime != nil {
		fields.runtime.RuntimeSetSubstExprSyntax(syntax)
	}
	fields.ExprSyntax = syntax
}

func (s *Subst) IsImplicit() bool {
	fields := s.require()
	if fields.runtime != nil {
		if implicit, present := fields.runtime.RuntimeSubstImplicit(); present {
			return implicit
		}
	}
	return fields.Implicit
}

func (s *Subst) BindRuntime(runtime SubstRuntime) {
	s.require().runtime = runtime
}

func (s *Subst) LevelDataToString() string {
	s.require()
	return "Dummy level string"
}

var nextSubstID atomic.Uint64

func ensureSubstIdentity(subst Subst) Subst {
	if subst.SubstFields == nil {
		subst.SubstFields = &SubstFields{}
	}
	if subst.id == 0 {
		subst.id = nextSubstID.Add(1)
	}
	return subst
}

func substIdentity(subst Subst) uint64 {
	return ensureSubstIdentity(subst).id
}

func copySubstsWithIdentity(substs []Subst) []Subst {
	out := make([]Subst, len(substs))
	for i, subst := range substs {
		out[i] = ensureSubstIdentity(subst)
	}
	return out
}

type SubstInNode struct {
	*SemanticNodeBase
	Substs              []Subst
	Body                SemanticNode
	InstantiatingModule *ModuleNode
	InstantiatedModule  *ModuleNode
}

func NewSubstInNode(body SemanticNode, substs ...Subst) *SubstInNode {
	return &SubstInNode{
		SemanticNodeBase: newSemanticNodeBasePointer(SemanticSubstInKind, "subst"),
		Substs:           copySubstsWithIdentity(substs),
		Body:             body,
	}
}

func (n *SubstInNode) require() {
	if n == nil {
		panic(NewNullPointerException())
	}
}

func (n *SubstInNode) GetSubsts() []Subst {
	n.require()
	return n.Substs
}

func (n *SubstInNode) GetBody() SemanticNode {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface {
		RuntimeSubstitutionBody() (SemanticNode, bool)
	}); ok {
		if body, present := provider.RuntimeSubstitutionBody(); present {
			return body
		}
	}
	return n.Body
}

func (n *SubstInNode) SetBody(body SemanticNode) {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface {
		RuntimeSetSubstitutionBody(SemanticNode) bool
	}); ok {
		provider.RuntimeSetSubstitutionBody(body)
	}
	n.Body = body
}

func (n *SubstInNode) GetInstantiatingModule() *ModuleNode {
	n.require()
	return n.InstantiatingModule
}

func (n *SubstInNode) GetInstantiatedModule() *ModuleNode {
	n.require()
	return n.InstantiatedModule
}

func substitutionAt(substs []Subst, index int) *Subst {
	if substs == nil {
		panic(NewNullPointerException())
	}
	index = int(int32(index))
	if index < 0 || index >= len(substs) {
		panic(NewArrayIndexOutOfBoundsException(index, len(substs)))
	}
	return &substs[index]
}

func (n *SubstInNode) GetSubFor(index int) *SymbolNode {
	n.require()
	return substitutionAt(n.Substs, index).GetOp()
}

func (n *SubstInNode) GetSubWith(index int) SemanticNode {
	n.require()
	return substitutionAt(n.Substs, index).GetExpr()
}

func (n *SubstInNode) GetChildren() []SemanticNode {
	n.require()
	if n.Substs == nil {
		panic(NewNullPointerException())
	}
	children := make([]SemanticNode, len(n.Substs)+1)
	children[0] = n.GetBody()
	for i := range n.Substs {
		children[i+1] = n.Substs[i].GetExpr()
	}
	return children
}

// SANY's SubstInNode(SubstInNode, ExprNode, Errors) constructor reuses the
// source syntax and substitution array when replacing the expression body.
func NewSubstInNodeFromSource(source *SubstInNode, body SemanticNode) *SubstInNode {
	if source == nil {
		panic(NewNullPointerException())
	}
	node := NewSubstInNode(body)
	node.TreeNode = source.TreeNode
	node.Location = source.Location
	node.Substs = source.Substs
	node.InstantiatingModule = source.InstantiatingModule
	node.InstantiatedModule = source.InstantiatedModule
	return node
}

type APSubstInNode struct {
	*SemanticNodeBase
	Substs              []Subst
	Body                SemanticNode
	InstantiatingModule *ModuleNode
	InstantiatedModule  *ModuleNode
}

func NewAPSubstInNode(body SemanticNode, substs ...Subst) *APSubstInNode {
	return &APSubstInNode{
		SemanticNodeBase: newSemanticNodeBasePointer(SemanticAPSubstInKind, "ap-subst"),
		Substs:           copySubstsWithIdentity(substs),
		Body:             body,
	}
}

func (n *APSubstInNode) require() {
	if n == nil {
		panic(NewNullPointerException())
	}
}

func (n *APSubstInNode) GetSubsts() []Subst {
	n.require()
	return n.Substs
}

func (n *APSubstInNode) GetBody() SemanticNode {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface {
		RuntimeSubstitutionBody() (SemanticNode, bool)
	}); ok {
		if body, present := provider.RuntimeSubstitutionBody(); present {
			return body
		}
	}
	return n.Body
}

func (n *APSubstInNode) SetBody(body SemanticNode) {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface {
		RuntimeSetSubstitutionBody(SemanticNode) bool
	}); ok {
		provider.RuntimeSetSubstitutionBody(body)
	}
	n.Body = body
}

func (n *APSubstInNode) GetInstantiatingModule() *ModuleNode {
	n.require()
	return n.InstantiatingModule
}

func (n *APSubstInNode) GetInstantiatedModule() *ModuleNode {
	n.require()
	return n.InstantiatedModule
}

func (n *APSubstInNode) GetSubFor(index int) *SymbolNode {
	n.require()
	return substitutionAt(n.Substs, index).GetOp()
}

func (n *APSubstInNode) GetSubWith(index int) SemanticNode {
	n.require()
	return substitutionAt(n.Substs, index).GetExpr()
}

func (n *APSubstInNode) GetChildren() []SemanticNode {
	n.require()
	if n.Substs == nil {
		panic(NewNullPointerException())
	}
	children := make([]SemanticNode, len(n.Substs)+1)
	children[0] = n.GetBody()
	for i := range n.Substs {
		children[i+1] = n.Substs[i].GetExpr()
	}
	return children
}

type ValueNode struct {
	SemanticNodeBase
	Value Value
}

func NewValueNode(value Value) *ValueNode {
	base := NewSemanticNodeBase(SemanticValueKind, semanticValueString(value))
	return &ValueNode{
		SemanticNodeBase: base,
		Value:            value,
	}
}

type NumeralNode struct {
	SemanticNodeBase
	Value    *IntValue
	BigValue *big.Int
}

func (n *NumeralNode) require() {
	if n == nil {
		panic(NewNullPointerException())
	}
}

func (n *NumeralNode) String() string {
	n.require()
	return n.Image
}

func (n *NumeralNode) UseVal() bool {
	n.require()
	return n.BigValue == nil
}
func (n *NumeralNode) Val() int32 {
	n.require()
	return n.Value.ownedVal()
}
func (n *NumeralNode) BigVal() *big.Int {
	n.require()
	return n.BigValue
}

// NumeralNode.levelCheck records the supplied iteration, even if it is lower
// than a previous one. It has no descendants or level errors to compute.
func (n *NumeralNode) LevelCheck(iter int32) bool {
	n.require()
	n.LevelChecked = iter
	return true
}

// LevelNode's no-iteration overload requests the next iteration.
func (n *NumeralNode) LevelCheckNext() bool {
	n.require()
	return n.LevelCheck(n.LevelChecked + 1)
}

func NewNumeralNode(value int32) *NumeralNode {
	intValue := NewIntValue(value)
	base := NewSemanticNodeBase(SemanticNumeralKind, strconv.FormatInt(int64(value), 10))
	base.ToolObject = intValue
	return &NumeralNode{
		SemanticNodeBase: base,
		Value:            intValue,
	}
}

// SANY retains out-of-range numerals and their original radix/image. TLC
// rejects the big value later, when SpecProcessor processes the constants.
func NewNumeralNodeFromString(image string) (*NumeralNode, error) {
	number, radix := strings.ToLower(image), 10
	if strings.HasPrefix(number, "\\") && len(number) > 2 {
		switch number[1] {
		case 'b':
			radix = 2
		case 'o':
			radix = 8
		case 'h':
			radix = 16
		default:
			return nil, fmt.Errorf("Unknown numeral format: %s", number)
		}
		number = number[2:]
	}
	value, err := strconv.ParseInt(number, radix, 32)
	node := NewNumeralNode(int32(value))
	node.Image = image
	if err != nil {
		large, valid := new(big.Int).SetString(number, radix)
		if !valid {
			return nil, err
		}
		node.Value = IntZero
		node.BigValue = large
		node.SetToolObject(nil)
	}
	return node, nil
}

// DecimalLiteralValue retains BigDecimal's unscaled value and scale for SANY
// literal metadata. TLC still rejects real numbers during constant processing.
type DecimalLiteralValue struct {
	unscaled *big.Int
	scale    int32
}

func (v *DecimalLiteralValue) UnscaledValue() *big.Int { return v.unscaled }
func (v *DecimalLiteralValue) Scale() int32            { return v.scale }
func (v *DecimalLiteralValue) String() string {
	image := new(big.Int).Abs(v.unscaled).String()
	sign := ""
	if v.unscaled.Sign() < 0 {
		sign = "-"
	}
	adjusted := int64(len(image)) - 1 - int64(v.scale)
	if v.scale >= 0 && adjusted >= -6 {
		point := len(image) - int(v.scale)
		if v.scale == 0 {
			return sign + image
		}
		if point > 0 {
			return sign + image[:point] + "." + image[point:]
		}
		return sign + "0." + strings.Repeat("0", -point) + image
	}
	if len(image) > 1 {
		image = image[:1] + "." + image[1:]
	}
	exponent := strconv.FormatInt(adjusted, 10)
	if adjusted >= 0 {
		exponent = "+" + exponent
	}
	return sign + image + "E" + exponent
}

type DecimalNode struct {
	SemanticNodeBase
	Value          Value
	IntegralPart   string
	FractionalPart string
	mantissa       int64
	exponent       int32
	bigValue       *DecimalLiteralValue
}

func (n *DecimalNode) require() {
	if n == nil {
		panic(NewNullPointerException())
	}
}

func (n *DecimalNode) String() string {
	n.require()
	return n.IntegralPart + "." + n.FractionalPart
}
func (n *DecimalNode) Mantissa() int64 {
	n.require()
	return n.mantissa
}
func (n *DecimalNode) Exponent() int32 {
	n.require()
	return n.exponent
}
func (n *DecimalNode) BigVal() *DecimalLiteralValue {
	n.require()
	return n.bigValue
}
func (n *DecimalNode) LevelCheck(iter int32) bool {
	n.require()
	n.LevelChecked = iter
	return true
}
func (n *DecimalNode) LevelCheckNext() bool {
	n.require()
	return n.LevelCheck(n.LevelChecked + 1)
}

// DecimalNode retains trailing zeros; Java's constructor does not normalize
// the mantissa or exponent, despite its historical class comment.
func NewDecimalNodeFromParts(integral, fractional string) *DecimalNode {
	n := &DecimalNode{
		SemanticNodeBase: NewSemanticNodeBase(SemanticDecimalKind, integral+"."+fractional),
		IntegralPart:     integral, FractionalPart: fractional,
	}
	mantissa, err := strconv.ParseInt(integral+fractional, 10, 64)
	if err == nil {
		n.mantissa = mantissa
		n.exponent = -int32(len(fractional))
	} else {
		unscaled, valid := new(big.Int).SetString(integral+fractional, 10)
		if !valid {
			panic(NewNumberFormatException(n.String()))
		}
		n.bigValue = &DecimalLiteralValue{unscaled: unscaled, scale: int32(len(fractional))}
	}
	return n
}

func NewDecimalNode(value Value, image string) *DecimalNode {
	parts := strings.SplitN(image, ".", 2)
	if len(parts) != 2 {
		panic(NewNumberFormatException(image))
	}
	n := NewDecimalNodeFromParts(parts[0], parts[1])
	n.Value = value
	n.SetToolObject(value)
	return n
}

type StringNode struct {
	SemanticNodeBase
	Value *StringValue
}

func (n *StringNode) require() {
	if n == nil {
		panic(NewNullPointerException())
	}
}

func (n *StringNode) GetRep() *UniqueString {
	n.require()
	return n.Value.ownedVal()
}
func (n *StringNode) LevelCheck(iter int32) bool {
	n.require()
	n.LevelChecked = iter
	return true
}
func (n *StringNode) LevelCheckNext() bool {
	n.require()
	return n.LevelCheck(n.LevelChecked + 1)
}

func NewStringNode(value string) *StringNode {
	stringValue := NewStringValue(value)
	base := NewSemanticNodeBase(SemanticStringKind, strconv.Quote(value))
	base.ToolObject = stringValue
	return &StringNode{
		SemanticNodeBase: base,
		Value:            stringValue,
	}
}

type AtNode struct {
	*SemanticNodeBase
	ExceptRef          *OpApplNode
	ExceptComponentRef *OpApplNode
}

func NewAtNode(refs ...*OpApplNode) *AtNode {
	if len(refs) != 0 && len(refs) != 2 {
		panic(NewIllegalArgumentException("NewAtNode accepts either zero or two EXCEPT references"))
	}
	node := &AtNode{SemanticNodeBase: newSemanticNodeBasePointer(SemanticAtNodeKind, "@")}
	if len(refs) == 2 {
		node.ExceptRef = refs[0]
		node.ExceptComponentRef = refs[1]
		if refs[1] != nil {
			node.TreeNode = refs[1].TreeNode
			node.Location = refs[1].Location
		}
	}
	return node
}

func (n *AtNode) require() {
	if n == nil {
		panic(NewNullPointerException())
	}
}

func (n *AtNode) GetExceptRef() *OpApplNode {
	n.require()
	return n.ExceptRef
}

func (n *AtNode) GetExceptComponentRef() *OpApplNode {
	n.require()
	return n.ExceptComponentRef
}

func semanticNodeArrayFirst(nodes []SemanticNode) SemanticNode {
	if nodes == nil {
		panic(NewNullPointerException())
	}
	if len(nodes) == 0 {
		panic(NewArrayIndexOutOfBoundsException(0, 0))
	}
	return nodes[0]
}

func (n *AtNode) GetAtBase() SemanticNode {
	n.require()
	base := semanticNodeArrayFirst(n.ExceptRef.GetArgs())
	if isNil(base) {
		return nil
	}
	switch base.(type) {
	case *OpApplNode, *LetInNode, *SubstInNode, *NumeralNode, *DecimalNode, *StringNode, *AtNode, *LabelNode, *ValueNode, *LiveExprNode:
		return base
	default:
		panic(valueStreamClassCast(base, "tla2sany.semantic.ExprNode"))
	}
}

func (n *AtNode) GetAtModifier() *OpApplNode {
	n.require()
	modifier := semanticNodeArrayFirst(n.ExceptComponentRef.GetArgs())
	if modifier == nil {
		return nil
	}
	result, ok := modifier.(*OpApplNode)
	if !ok {
		panic(valueStreamClassCast(modifier, "tla2sany.semantic.OpApplNode"))
	}
	return result
}

type OpArgNode struct {
	*SemanticNodeBase
	Op     *SymbolNode
	Name   *UniqueString
	Arity  int
	Module *ModuleNode
}

func NewOpArgNode(op *SymbolNode) *OpArgNode {
	op.require()
	return &OpArgNode{
		SemanticNodeBase: newSemanticNodeBasePointer(SemanticOpArgKind, op.String()),
		Op:               op,
		Name:             op.Name,
		Arity:            op.GetArity(),
	}
}

func (n *OpArgNode) require() {
	if n == nil {
		panic(NewNullPointerException())
	}
}

func (n *OpArgNode) GetOp() *SymbolNode {
	n.require()
	return n.Op
}

func (n *OpArgNode) GetArity() int {
	n.require()
	return n.Arity
}

func (n *OpArgNode) GetName() *UniqueString {
	n.require()
	return n.Name
}

func (n *OpArgNode) GetModule() *ModuleNode {
	n.require()
	return n.Module
}

type PossibleTrackNode struct {
	SemanticNodeBase
	Pred SemanticNode
	Name string
}

func NewPossibleTrackNode(pred SemanticNode, name string) *PossibleTrackNode {
	base := NewSemanticNodeBase(SemanticPossibleTrackKind, "_Possible!_Track("+name+")")
	base.LevelValue = SemanticLevel(pred)
	out := &PossibleTrackNode{
		SemanticNodeBase: base,
		Pred:             pred,
		Name:             name,
	}
	out.SetLevelParams(SemanticLevelParams(pred)...)
	return out
}

type PossibleCheckNode struct {
	SemanticNodeBase
	Name string
}

func NewPossibleCheckNode(name string) *PossibleCheckNode {
	base := NewSemanticNodeBase(SemanticPossibleCheckKind, "_Possible!_CheckName("+name+")")
	base.LevelValue = TLCLevelConstant
	return &PossibleCheckNode{
		SemanticNodeBase: base,
		Name:             name,
	}
}

type ThmOrAssumpDefNode struct {
	*SemanticNodeBase
	Name                      *UniqueString
	Body                      SemanticNode
	Params                    []*SymbolNode
	Symbol                    *SymbolNode
	Proof                     SemanticNode
	Local                     bool
	Theorem                   bool
	Suffices                  bool
	Arity                     int
	ArgMaxLevels              []int
	ArgWeights                []int
	MinMaxLevel               [][]int
	OpLevelCond               [][][]bool
	IsLeibnizArg              []bool
	IsLeibnizValue            bool
	Labels                    *SemanticLabelTable
	OriginallyDefinedInModule *ModuleNode
	InstantiatedFrom          *ModuleNode
	SourceDefinition          *ThmOrAssumpDefNode
}

type AssumeNode struct {
	*SemanticNodeBase
	Assume  SemanticNode
	Module  *ModuleNode
	Def     *ThmOrAssumpDefNode
	IsAxiom bool
}

func NewAssumeNode(expr SemanticNode, module *ModuleNode, definition *ThmOrAssumpDefNode) *AssumeNode {
	return &AssumeNode{SemanticNodeBase: newSemanticNodeBasePointer(SemanticAssumeKind, "ASSUME"), Assume: expr, Module: module, Def: definition}
}

func (n *AssumeNode) require() {
	if n == nil {
		panic(NewNullPointerException())
	}
}

func (n *AssumeNode) GetIsAxiom() bool {
	n.require()
	return n.IsAxiom
}

func (n *AssumeNode) GetAssume() SemanticNode {
	n.require()
	return n.Assume
}
func (n *AssumeNode) GetDef() *ThmOrAssumpDefNode {
	n.require()
	return n.Def
}

func (n *AssumeNode) GetLevel() int {
	n.require()
	if semanticExploreNull(n.Assume) {
		panic(NewNullPointerException())
	}
	return semanticNodeLevel(n.Assume)
}

func (n *AssumeNode) GetLevelParams() []*SymbolNode {
	n.require()
	if semanticExploreNull(n.Assume) {
		panic(NewNullPointerException())
	}
	if carrier, ok := n.Assume.(interface{ GetCanonicalLevelData() any }); ok {
		if provider, ok := carrier.GetCanonicalLevelData().(interface{ RuntimeCheckedLevelParams() []*SymbolNode }); ok {
			return provider.RuntimeCheckedLevelParams()
		}
	}
	return SemanticLevelParams(n.Assume)
}

func (n *AssumeNode) GetAllParams() []*SymbolNode {
	n.require()
	if semanticExploreNull(n.Assume) {
		panic(NewNullPointerException())
	}
	if carrier, ok := n.Assume.(interface{ GetCanonicalLevelData() any }); ok {
		if provider, ok := carrier.GetCanonicalLevelData().(interface{ RuntimeCheckedAllParams() []*SymbolNode }); ok {
			return provider.RuntimeCheckedAllParams()
		}
	}
	return SemanticAllParams(n.Assume)
}

func (n *AssumeNode) GetChildren() []SemanticNode {
	n.require()
	return []SemanticNode{n.Assume}
}

func NewThmOrAssumpDefNode(name string, body SemanticNode, params ...*SymbolNode) *ThmOrAssumpDefNode {
	node := &ThmOrAssumpDefNode{SemanticNodeBase: newSemanticNodeBasePointer(SemanticThmOrAssumpKind, name), Name: UniqueStringOf(name), Body: body, Params: append([]*SymbolNode{}, params...), Symbol: NewSymbolNode(name), Theorem: true, Arity: len(params)}
	node.Symbol.SemanticBase = node.SemanticNodeBase
	node.Symbol.Arity = len(params)
	node.Symbol.Data = node
	return node
}

func (n *ThmOrAssumpDefNode) require() {
	if n == nil {
		panic(NewNullPointerException())
	}
}

func (n *ThmOrAssumpDefNode) GetSource() *ThmOrAssumpDefNode {
	n.require()
	if n.SourceDefinition != nil {
		return n.SourceDefinition
	}
	return n
}

func (n *ThmOrAssumpDefNode) GetBody() SemanticNode {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface {
		RuntimeTheoremDefinitionBody() (SemanticNode, bool)
	}); ok {
		if body, present := provider.RuntimeTheoremDefinitionBody(); present {
			return body
		}
	}
	return n.Body
}

func (n *ThmOrAssumpDefNode) GetOriginallyDefinedInModuleNode() *ModuleNode {
	n.require()
	return n.OriginallyDefinedInModule
}

func (n *ThmOrAssumpDefNode) GetInstantiatedFrom() *ModuleNode {
	n.require()
	return n.InstantiatedFrom
}

func (n *ThmOrAssumpDefNode) IsTheorem() bool {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface{ RuntimeTheoremDefinitionFlags() (bool, bool, bool) }); ok {
		if theorem, _, present := provider.RuntimeTheoremDefinitionFlags(); present {
			return theorem
		}
	}
	return n.Theorem
}

func (n *ThmOrAssumpDefNode) IsSuffices() bool {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface{ RuntimeTheoremDefinitionFlags() (bool, bool, bool) }); ok {
		if _, suffices, present := provider.RuntimeTheoremDefinitionFlags(); present {
			return suffices
		}
	}
	return n.Suffices
}

func (n *ThmOrAssumpDefNode) GetProof() SemanticNode {
	n.require()
	return n.Proof
}

func (n *ThmOrAssumpDefNode) GetParams() []*SymbolNode {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface {
		RuntimeTheoremDefinitionParams() ([]*SymbolNode, bool)
	}); ok {
		if params, present := provider.RuntimeTheoremDefinitionParams(); present {
			return params
		}
	}
	return n.Params
}

func (n *ThmOrAssumpDefNode) IsExpr() bool {
	n.require()
	switch n.GetBody().(type) {
	case *OpApplNode, *LetInNode, *LabelNode, *AtNode, *NumeralNode, *DecimalNode, *StringNode, *SubstInNode:
		return true
	default:
		return false
	}
}

func (n *ThmOrAssumpDefNode) GetArity() int {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface{ RuntimeSymbolArity() (int, bool) }); ok {
		if arity, present := provider.RuntimeSymbolArity(); present {
			return arity
		}
	}
	return n.Arity
}

func (n *ThmOrAssumpDefNode) IsLocal() bool {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface{ RuntimeSymbolLocal() (bool, bool) }); ok {
		if local, present := provider.RuntimeSymbolLocal(); present {
			return local
		}
	}
	return n.Local
}

func (n *ThmOrAssumpDefNode) SetLocal(local bool) {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface{ RuntimeSetSymbolLocal(bool) bool }); ok {
		provider.RuntimeSetSymbolLocal(local)
	}
	n.Local = local
}

func (n *ThmOrAssumpDefNode) GetIsLeibnizArg() []bool {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface{ RuntimeIsLeibnizArg() []bool }); ok {
		return provider.RuntimeIsLeibnizArg()
	}
	return n.IsLeibnizArg
}

func (n *ThmOrAssumpDefNode) GetIsLeibniz() bool {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface{ RuntimeIsLeibniz() bool }); ok {
		return provider.RuntimeIsLeibniz()
	}
	return n.IsLeibnizValue
}

func (n *ThmOrAssumpDefNode) GetMaxLevel(index int) int {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface{ RuntimeCheckedMaxLevel(int) int }); ok {
		return provider.RuntimeCheckedMaxLevel(index)
	}
	n.requireLevelChecked("getMaxLevel called before levelCheck")
	if n.Arity == -1 {
		index = 0
	}
	return opDefArrayAt(n.ArgMaxLevels, index)
}

func (n *ThmOrAssumpDefNode) GetWeight(index int) int {
	n.require()
	if provider, ok := n.CanonicalLevelData.(interface{ RuntimeCheckedWeight(int) int }); ok {
		return provider.RuntimeCheckedWeight(index)
	}
	n.requireLevelChecked("getWeight called before levelCheck")
	if n.Arity == -1 {
		index = 0
	}
	return opDefArrayAt(n.ArgWeights, index)
}

func (n *ThmOrAssumpDefNode) GetMinMaxLevel(i, j int) int {
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

func (n *ThmOrAssumpDefNode) GetOpLevelCond(i, j, k int) bool {
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

func (n *ThmOrAssumpDefNode) requireLevelChecked(message string) {
	n.require()
	if n.LevelChecked == 0 {
		panic(NewWrongInvocationException(message))
	}
}

func (n *ThmOrAssumpDefNode) GetChildren() []SemanticNode {
	n.require()
	return []SemanticNode{n.GetBody()}
}

func (n *ThmOrAssumpDefNode) GetLabelsHT() *SemanticLabelTable {
	n.require()
	return semanticNodeLabels(n.SemanticNodeBase, n.Labels)
}

func (n *ThmOrAssumpDefNode) SetLabels(labels *SemanticLabelTable) {
	n.require()
	setSemanticNodeLabels(n.SemanticNodeBase, labels)
	n.Labels = labels
}

func (n *ThmOrAssumpDefNode) GetLabel(name *UniqueString) *LabelNode {
	n.require()
	labels := n.GetLabelsHT()
	if labels == nil {
		return nil
	}
	return labels.Get(name)
}

func (n *ThmOrAssumpDefNode) AddLabel(label *LabelNode) bool {
	n.require()
	labels := n.GetLabelsHT()
	if labels == nil {
		n.SetLabels(NewSemanticLabelTable())
		labels = n.GetLabelsHT()
	}
	return labels.Add(label)
}

func (n *ThmOrAssumpDefNode) GetLabels() []*LabelNode {
	n.require()
	labels := n.GetLabelsHT()
	if labels == nil {
		return make([]*LabelNode, 0)
	}
	return labels.Elements()
}

func (n *ThmOrAssumpDefNode) String() string {
	n.require()
	if n.hasLocationSyntax() {
		return semanticNodeJavaString(n)
	}
	if n.Name == nil {
		return "<theorem>"
	}
	return n.Name.String()
}

func SemanticKindOf(node SemanticNode) SemanticKind {
	switch n := node.(type) {
	case nil:
		return SemanticUnknownKind
	case *SymbolNode:
		return semanticSymbolKind(n)
	case *OpDefNode:
		return n.Kind()
	case *LabelNode:
		return SemanticLabelKind
	case *OpApplNode:
		return SemanticOpApplKind
	case *LetInNode:
		return SemanticLetInKind
	case *SubstInNode:
		return SemanticSubstInKind
	case *APSubstInNode:
		return SemanticAPSubstInKind
	case *NumeralNode:
		return SemanticNumeralKind
	case *DecimalNode:
		return SemanticDecimalKind
	case *StringNode:
		return SemanticStringKind
	case *AtNode:
		return SemanticAtNodeKind
	case *OpArgNode:
		return SemanticOpArgKind
	case *ValueNode:
		return SemanticValueKind
	case *PossibleTrackNode:
		return SemanticPossibleTrackKind
	case *PossibleCheckNode:
		return SemanticPossibleCheckKind
	case interface{ Kind() SemanticKind }:
		return n.Kind()
	default:
		return SemanticUnknownKind
	}
}

func semanticSymbolKind(symbol *SymbolNode) SemanticKind {
	symbol.require()
	if symbol.SemanticBase != nil {
		return symbol.SemanticBase.Kind()
	}
	switch symbol.Kind {
	case SymbolVariableDecl:
		return SemanticVariableDeclKind
	case SymbolConstantDecl:
		return SemanticConstantDeclKind
	case SymbolUserDefinedOp:
		return SemanticUserDefinedOpKind
	case SymbolBuiltIn:
		return SemanticBuiltInKind
	case SymbolFormalParam:
		return SemanticFormalParamKind
	default:
		return SemanticUnknownKind
	}
}

func SemanticJavaHashCode(node SemanticNode) int32 {
	if node == nil {
		return 0
	}
	if symbol, ok := node.(*SymbolNode); ok {
		result := int32(1)
		result = 31*result + int32(semanticSymbolKind(symbol))
		if symbol.SemanticBase != nil {
			result = 31*result + symbol.SemanticBase.GetUid()
		} else {
			result *= 31
		}
		return result
	}
	if h, ok := node.(interface{ JavaHashCode() int32 }); ok {
		return h.JavaHashCode()
	}
	if n, ok := node.(interface {
		Kind() SemanticKind
		GetUid() int32
	}); ok {
		result := int32(1)
		result = 31*result + int32(n.Kind())
		result = 31*result + n.GetUid()
		return result
	}
	if n, ok := node.(interface {
		Kind() SemanticKind
		GetUID() int32
	}); ok {
		result := int32(1)
		result = 31*result + int32(n.Kind())
		result = 31*result + n.GetUID()
		return result
	}
	result := int32(1)
	result = 31*result + int32(SemanticKindOf(node))
	result = 31 * result
	return result
}

// SemanticCompareTo ports SemanticNode.compareTo(SemanticNode). The source
// ordering uses only the current syntax location's begin line and column.
func SemanticCompareTo(left, right SemanticNode) int {
	if semanticExploreNull(left) || semanticExploreNull(right) {
		panic(NewNullPointerException())
	}
	leftSyntax, leftSyntaxOK := left.(interface{ GetTreeNode() any })
	rightSyntax, rightSyntaxOK := right.(interface{ GetTreeNode() any })
	if !leftSyntaxOK || !rightSyntaxOK {
		panic(NewClassCastException("object is not a SemanticNode"))
	}
	// Java compareTo dereferences each node's stn directly rather than using
	// SemanticNode.getLocation(), whose null-syntax fallback is Location.nullLoc.
	if semanticExploreNull(leftSyntax.GetTreeNode()) || semanticExploreNull(rightSyntax.GetTreeNode()) {
		panic(NewNullPointerException())
	}
	leftLocation, leftOK := semanticNodeSourceLocation(left)
	rightLocation, rightOK := semanticNodeSourceLocation(right)
	if !leftOK || !rightOK {
		panic(NewClassCastException("object is not a SemanticNode"))
	}
	if leftLocation.BeginLine < rightLocation.BeginLine {
		return -1
	}
	if leftLocation.BeginLine > rightLocation.BeginLine {
		return 1
	}
	if leftLocation.BeginColumn < rightLocation.BeginColumn {
		return -1
	}
	if leftLocation.BeginColumn > rightLocation.BeginColumn {
		return 1
	}
	return 0
}

// SemanticEquals ports SemanticNode.equals(Object): identity succeeds first,
// followed by exact Java class, semantic kind and UID comparisons.
func SemanticEquals(left SemanticNode, other any) bool {
	if semanticExploreNull(left) {
		panic(NewNullPointerException())
	}
	leftValue := reflect.ValueOf(left)
	rightValue := reflect.ValueOf(other)
	if rightValue.IsValid() && leftValue.Type() == rightValue.Type() && leftValue.Kind() == reflect.Pointer && leftValue.Pointer() == rightValue.Pointer() {
		return true
	}
	if !rightValue.IsValid() || rightValue.Kind() == reflect.Pointer && rightValue.IsNil() {
		return false
	}
	leftKind, leftUID, leftClass, leftOK := semanticEqualityIdentity(left)
	rightKind, rightUID, rightClass, rightOK := semanticEqualityIdentity(other)
	return leftOK && rightOK && leftClass == rightClass && leftKind == rightKind && leftUID == rightUID
}

func semanticEqualityIdentity(node any) (SemanticKind, int32, string, bool) {
	if semanticExploreNull(node) {
		return SemanticUnknownKind, 0, "", false
	}
	if symbol, ok := node.(*SymbolNode); ok {
		if symbol.SemanticBase == nil {
			return SemanticUnknownKind, 0, "", false
		}
		return symbol.SemanticBase.Kind(), symbol.SemanticBase.GetUid(), semanticJavaClass(symbol), true
	}
	semantic, ok := node.(interface {
		Kind() SemanticKind
		GetUid() int32
	})
	if !ok {
		return SemanticUnknownKind, 0, "", false
	}
	return semantic.Kind(), semantic.GetUid(), semanticJavaClass(node), true
}

func SemanticToolObject(node SemanticNode) any {
	switch n := node.(type) {
	case nil:
		return nil
	case Value:
		return n
	case *ValueNode:
		return n.Value
	case *NumeralNode:
		return n.Value
	case *DecimalNode:
		return n.Value
	case *StringNode:
		return n.Value
	case interface{ GetToolObject() any }:
		return n.GetToolObject()
	default:
		return nil
	}
}

func SemanticToolObjectForTool(tool *Tool, node SemanticNode) any {
	toolID := int32(0)
	if tool != nil {
		toolID = tool.GetID()
	}
	return SemanticToolObjectForToolID(toolID, node)
}

func SemanticToolObjectForToolID(toolID int32, node SemanticNode) any {
	if node == nil {
		return nil
	}
	if slots, ok := node.(interface{ GetToolObjectAt(int32) any }); ok {
		return slots.GetToolObjectAt(toolID)
	}
	panic(NewClassCastException("semantic node does not implement indexed tool-object storage"))
}

func SetSemanticToolObjectForTool(tool *Tool, node SemanticNode, value any) {
	toolID := int32(0)
	if tool != nil {
		toolID = tool.GetID()
	}
	SetSemanticToolObjectForToolID(toolID, node, value)
}

func SetSemanticToolObjectForToolID(toolID int32, node SemanticNode, value any) {
	if node == nil {
		return
	}
	if slots, ok := node.(interface{ SetToolObjectAt(int32, any) }); ok {
		slots.SetToolObjectAt(toolID, value)
		return
	}
	panic(NewClassCastException("semantic node does not implement indexed tool-object storage"))
}

func SemanticString(node SemanticNode) string {
	if node == nil {
		return "<nil>"
	}
	if source, ok := node.(interface{ hasSourceSyntax() bool }); ok && source.hasSourceSyntax() {
		return semanticNodeJavaString(node)
	}
	if s, ok := node.(fmt.Stringer); ok {
		return s.String()
	}
	return fmt.Sprint(node)
}

func semanticValueString(value Value) string {
	if value == nil {
		return "<nil>"
	}
	return value.String()
}

func (n *NumeralNode) GetLevel() int {
	n.require()
	if n.LevelChecked == 0 {
		panic(NewWrongInvocationException("getLevel called before levelCheck"))
	}
	return n.LevelValue
}

func (n *DecimalNode) GetLevel() int {
	n.require()
	if n.LevelChecked == 0 {
		panic(NewWrongInvocationException("getLevel called before levelCheck"))
	}
	return n.LevelValue
}

func (n *StringNode) GetLevel() int {
	n.require()
	if n.LevelChecked == 0 {
		panic(NewWrongInvocationException("getLevel called before levelCheck"))
	}
	return n.LevelValue
}

func (n *NumeralNode) GetLevelParams() []*SymbolNode {
	n.require()
	if n.LevelChecked == 0 {
		panic(NewWrongInvocationException("getLevelParams called before levelCheck"))
	}
	return n.SemanticNodeBase.GetLevelParams()
}

func (n *DecimalNode) GetLevelParams() []*SymbolNode {
	n.require()
	if n.LevelChecked == 0 {
		panic(NewWrongInvocationException("getLevelParams called before levelCheck"))
	}
	return n.SemanticNodeBase.GetLevelParams()
}

func (n *StringNode) GetLevelParams() []*SymbolNode {
	n.require()
	if n.LevelChecked == 0 {
		panic(NewWrongInvocationException("getLevelParams called before levelCheck"))
	}
	return n.SemanticNodeBase.GetLevelParams()
}
