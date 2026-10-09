package tlc

import (
	"fmt"
	"math/big"
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
	SemanticUnknownKind        SemanticKind = 0
	SemanticModuleKind         SemanticKind = 1
	SemanticUserDefinedOpKind  SemanticKind = 5
	SemanticModuleInstanceKind SemanticKind = 6
	SemanticBuiltInKind        SemanticKind = 7
	SemanticOpArgKind          SemanticKind = 8
	SemanticOpApplKind         SemanticKind = 9
	SemanticLetInKind          SemanticKind = 10
	SemanticFormalParamKind    SemanticKind = 11
	SemanticTheoremKind        SemanticKind = 12
	SemanticAssumeKind         SemanticKind = 20
	SemanticSubstInKind        SemanticKind = 13
	SemanticNumeralKind        SemanticKind = 16
	SemanticDecimalKind        SemanticKind = 17
	SemanticStringKind         SemanticKind = 18
	SemanticAtNodeKind         SemanticKind = 19
	SemanticThmOrAssumpKind    SemanticKind = 23
	SemanticLabelKind          SemanticKind = 29
	SemanticAPSubstInKind      SemanticKind = 30
	SemanticValueKind          SemanticKind = 1001
	SemanticPossibleTrackKind  SemanticKind = 1002
	SemanticPossibleCheckKind  SemanticKind = 1003
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

func (nullSemanticSyntax) GetHumanReadableImage() string { return "***I do not exist***" }

func (n *NullSemanticNode) LevelDataToString() string { return "-2147483648" }

var NullSemanticNodeInstance = func() *NullSemanticNode {
	base := NewSemanticNodeBase(SemanticKind(-2147483648), "***I do not exist***")
	base.Location = NewSourceLocation("--TLA+ BUILTINS--", 0, 0, 0, 0)
	base.TreeNode = nullSemanticSyntax{}
	return &NullSemanticNode{base}
}()

func (n *SemanticNodeBase) GetTreeNode() any     { return n.TreeNode }
func (n *SemanticNodeBase) SetTreeNode(node any) { n.TreeNode = node }

func (n *SemanticNodeBase) GetHumanReadableImage() string { return n.Location.String() }

// SemanticNode.toString normally prints the location. Numeral and Decimal
// override it, while showPlainFormulae selects the actual SANY syntax image.
// SemanticNodeJavaString exposes the source zero-argument formatting to SANY.
func SemanticNodeJavaString(node SemanticNode) string { return semanticNodeJavaString(node) }

func semanticNodeJavaString(node SemanticNode) string {
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
	if n == nil {
		return SemanticUnknownKind
	}
	return n.KindValue
}

func (n *SemanticNodeBase) GetUID() int32 {
	if n == nil {
		return -1
	}
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

func (n *SemanticNodeBase) JavaHashCode() int32 {
	if n == nil {
		return 0
	}
	result := int32(1)
	result = 31*result + int32(n.Kind())
	result = 31*result + n.GetUID()
	return result
}

func (n *SemanticNodeBase) GetToolObject() any {
	if n == nil {
		return nil
	}
	return n.ToolObject
}

func (n *SemanticNodeBase) SetToolObject(value any) {
	if n != nil {
		n.ToolObject = value
	}
}

func (n *SemanticNodeBase) toolSlots() *semanticNodeToolSlots {
	if atomic.LoadUint32(&n.uidInitialized) == 0 {
		n.GetUID()
	}
	return n.indexedTools
}

// GetToolObjectAt ports SemanticNode.getToolObject(toolId). Missing nonnegative
// slots return null; negative indices throw even when the array is empty.
func (n *SemanticNodeBase) GetToolObjectAt(toolID int32) any {
	if n == nil {
		return nil
	}
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
	if n == nil {
		return
	}
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
	if n != nil && !n.Location.IsNull() {
		return n.Location.String()
	}
	if n == nil || n.Image == "" {
		return "<semantic node>"
	}
	return n.Image
}

func (n *SemanticNodeBase) SourceLocation() SourceLocation {
	if n == nil {
		return NullSourceLocation
	}
	return n.Location
}

func (n *SemanticNodeBase) GetSourceLocation() SourceLocation {
	return n.SourceLocation()
}

func (n *SemanticNodeBase) SetSourceLocation(location SourceLocation) {
	if n != nil {
		n.Location = location
	}
}

// SemanticNode.isStandardModule delegates to Java's StandardModules name set.
// It deliberately classifies by source module name, including replaced modules.
func (n *SemanticNodeBase) IsStandardModule() bool {
	if n == nil {
		return false
	}
	switch n.Location.Source {
	case "FiniteSets", "Sequences", "Bags", "Naturals", "Integers", "Reals", "RealTime", "Randomization", "TLC":
		return true
	default:
		return false
	}
}

func (n *SemanticNodeBase) GetLevel() int {
	if n == nil {
		return TLCLevelConstant
	}
	return n.LevelValue
}

func (n *SemanticNodeBase) SetLevel(level int) {
	if n != nil {
		n.LevelValue = level
	}
}

func (n *SemanticNodeBase) GetLevelParams() []*SymbolNode {
	if n == nil || len(n.LevelParamSet) == 0 {
		return nil
	}
	out := make([]*SymbolNode, len(n.LevelParamSet))
	copy(out, n.LevelParamSet)
	return out
}

func (n *SemanticNodeBase) SetLevelParams(params ...*SymbolNode) {
	if n == nil {
		return
	}
	n.LevelParamSet = nil
	n.AddLevelParams(params...)
}

func (n *SemanticNodeBase) AddLevelParams(params ...*SymbolNode) {
	if n == nil {
		return
	}
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

func SemanticLevel(node SemanticNode) int {
	switch n := node.(type) {
	case nil:
		return TLCLevelConstant
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
	Body   SemanticNode
	Params []*SymbolNode
}

func NewLabelNode(body SemanticNode) *LabelNode {
	return &LabelNode{SemanticNodeBase: newSemanticNodeBasePointer(SemanticLabelKind, "label"), Body: body}
}

type OpApplNode struct {
	*SemanticNodeBase
	Operator             *SymbolNode
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
		Args:             append([]SemanticNode(nil), args...),
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

func (n *OpApplNode) String() string {
	if n != nil && !n.Location.IsNull() {
		return n.Location.String()
	}
	if n == nil || n.Operator == nil {
		return "<operator application>"
	}
	if n.Image != "" && n.Image != "<operator application>" {
		return n.Image
	}
	return n.Operator.String()
}

type LetInNode struct {
	*SemanticNodeBase
	Lets     []*OpDefNode
	Bindings []LetBinding
	Body     SemanticNode
	Context  *SemanticContext
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
	return &LetInNode{
		SemanticNodeBase: base,
		Lets:             append([]*OpDefNode(nil), lets...),
		Body:             body,
	}
}

type Subst struct {
	*SubstFields
	id uint64
}

// Copies of a substitution reference one record, as Java Subst[] entries do.
type SubstFields struct {
	Op   *SymbolNode
	Expr SemanticNode
}

func NewSubst(op *SymbolNode, expr SemanticNode) Subst {
	return Subst{SubstFields: &SubstFields{Op: op, Expr: expr}, id: nextSubstID.Add(1)}
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
	Substs []Subst
	Body   SemanticNode
}

func NewSubstInNode(body SemanticNode, substs ...Subst) *SubstInNode {
	return &SubstInNode{
		SemanticNodeBase: newSemanticNodeBasePointer(SemanticSubstInKind, "subst"),
		Substs:           copySubstsWithIdentity(substs),
		Body:             body,
	}
}

// SANY's SubstInNode(SubstInNode, ExprNode, Errors) constructor reuses the
// source syntax and substitution array when replacing the expression body.
func NewSubstInNodeFromSource(source *SubstInNode, body SemanticNode) *SubstInNode {
	node := NewSubstInNode(body)
	node.TreeNode = source.TreeNode
	node.Location = source.Location
	node.Substs = source.Substs
	return node
}

type APSubstInNode struct {
	*SemanticNodeBase
	Substs []Subst
	Body   SemanticNode
}

func NewAPSubstInNode(body SemanticNode, substs ...Subst) *APSubstInNode {
	return &APSubstInNode{
		SemanticNodeBase: newSemanticNodeBasePointer(SemanticAPSubstInKind, "ap-subst"),
		Substs:           copySubstsWithIdentity(substs),
		Body:             body,
	}
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

func (n *NumeralNode) String() string { return n.Image }

func (n *NumeralNode) UseVal() bool     { return n.BigValue == nil }
func (n *NumeralNode) Val() int32       { return n.Value.Val }
func (n *NumeralNode) BigVal() *big.Int { return n.BigValue }

// NumeralNode.levelCheck records the supplied iteration, even if it is lower
// than a previous one. It has no descendants or level errors to compute.
func (n *NumeralNode) LevelCheck(iter int32) bool {
	if n == nil {
		panic(NewNullPointerException())
	}
	n.LevelChecked = iter
	return true
}

// LevelNode's no-iteration overload requests the next iteration.
func (n *NumeralNode) LevelCheckNext() bool {
	if n == nil {
		panic(NewNullPointerException())
	}
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

func (n *DecimalNode) String() string               { return n.IntegralPart + "." + n.FractionalPart }
func (n *DecimalNode) Mantissa() int64              { return n.mantissa }
func (n *DecimalNode) Exponent() int32              { return n.exponent }
func (n *DecimalNode) BigVal() *DecimalLiteralValue { return n.bigValue }
func (n *DecimalNode) LevelCheck(iter int32) bool {
	if n == nil {
		panic(NewNullPointerException())
	}
	n.LevelChecked = iter
	return true
}
func (n *DecimalNode) LevelCheckNext() bool {
	if n == nil {
		panic(NewNullPointerException())
	}
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

func (n *StringNode) GetRep() *UniqueString { return n.Value.Val }
func (n *StringNode) LevelCheck(iter int32) bool {
	if n == nil {
		panic(NewNullPointerException())
	}
	n.LevelChecked = iter
	return true
}
func (n *StringNode) LevelCheckNext() bool {
	if n == nil {
		panic(NewNullPointerException())
	}
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
}

func NewAtNode() *AtNode {
	return &AtNode{SemanticNodeBase: newSemanticNodeBasePointer(SemanticAtNodeKind, "@")}
}

type OpArgNode struct {
	*SemanticNodeBase
	Op *SymbolNode
}

func NewOpArgNode(op *SymbolNode) *OpArgNode {
	return &OpArgNode{SemanticNodeBase: newSemanticNodeBasePointer(SemanticOpArgKind, op.String()), Op: op}
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
	Local                     bool
	OriginallyDefinedInModule *ModuleNode
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

func (n *AssumeNode) GetAssume() SemanticNode     { return n.Assume }
func (n *AssumeNode) GetDef() *ThmOrAssumpDefNode { return n.Def }

func NewThmOrAssumpDefNode(name string, body SemanticNode, params ...*SymbolNode) *ThmOrAssumpDefNode {
	node := &ThmOrAssumpDefNode{SemanticNodeBase: newSemanticNodeBasePointer(SemanticThmOrAssumpKind, name), Name: UniqueStringOf(name), Body: body, Params: append([]*SymbolNode(nil), params...), Symbol: NewSymbolNode(name)}
	node.Symbol.Arity = len(params)
	node.Symbol.Data = node
	return node
}

func (n *ThmOrAssumpDefNode) GetSource() *ThmOrAssumpDefNode {
	if n.SourceDefinition != nil {
		return n.SourceDefinition
	}
	return n
}

func (n *ThmOrAssumpDefNode) String() string {
	if n == nil || n.Name == nil {
		return "<theorem>"
	}
	return n.Name.String()
}

func SemanticKindOf(node SemanticNode) SemanticKind {
	switch n := node.(type) {
	case nil:
		return SemanticUnknownKind
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

func SemanticJavaHashCode(node SemanticNode) int32 {
	if node == nil {
		return 0
	}
	if h, ok := node.(interface{ JavaHashCode() int32 }); ok {
		return h.JavaHashCode()
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
		toolID = tool.ID
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
		toolID = tool.ID
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
	if n == nil {
		panic(NewNullPointerException())
	}
	if n.LevelChecked == 0 {
		panic(NewWrongInvocationException("getLevel called before levelCheck"))
	}
	return n.LevelValue
}

func (n *DecimalNode) GetLevel() int {
	if n == nil {
		panic(NewNullPointerException())
	}
	if n.LevelChecked == 0 {
		panic(NewWrongInvocationException("getLevel called before levelCheck"))
	}
	return n.LevelValue
}

func (n *StringNode) GetLevel() int {
	if n == nil {
		panic(NewNullPointerException())
	}
	if n.LevelChecked == 0 {
		panic(NewWrongInvocationException("getLevel called before levelCheck"))
	}
	return n.LevelValue
}

func (n *NumeralNode) GetLevelParams() []*SymbolNode {
	if n == nil {
		panic(NewNullPointerException())
	}
	if n.LevelChecked == 0 {
		panic(NewWrongInvocationException("getLevelParams called before levelCheck"))
	}
	return n.SemanticNodeBase.GetLevelParams()
}

func (n *DecimalNode) GetLevelParams() []*SymbolNode {
	if n == nil {
		panic(NewNullPointerException())
	}
	if n.LevelChecked == 0 {
		panic(NewWrongInvocationException("getLevelParams called before levelCheck"))
	}
	return n.SemanticNodeBase.GetLevelParams()
}

func (n *StringNode) GetLevelParams() []*SymbolNode {
	if n == nil {
		panic(NewNullPointerException())
	}
	if n.LevelChecked == 0 {
		panic(NewWrongInvocationException("getLevelParams called before levelCheck"))
	}
	return n.SemanticNodeBase.GetLevelParams()
}
