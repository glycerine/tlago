package tlc

import (
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
)

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
	KindValue     SemanticKind
	uidPlusOne    int32
	ToolObject    any
	Image         string
	LevelValue    int
	LevelParamSet []*SymbolNode
	Location      SourceLocation
}

var nextSemanticNodeUID atomic.Int32
var semanticToolObjects = struct {
	sync.RWMutex
	values map[semanticToolObjectKey]any
}{values: make(map[semanticToolObjectKey]any)}

type semanticToolObjectKey struct {
	toolID int64
	nodeID int32
}

func newSemanticNodeBase(kind SemanticKind, image string) SemanticNodeBase {
	return SemanticNodeBase{KindValue: kind, uidPlusOne: nextSemanticNodeUID.Add(1), Image: image}
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
	uidPlusOne := atomic.LoadInt32(&n.uidPlusOne)
	if uidPlusOne == 0 {
		next := nextSemanticNodeUID.Add(1)
		if atomic.CompareAndSwapInt32(&n.uidPlusOne, 0, next) {
			uidPlusOne = next
		} else {
			uidPlusOne = atomic.LoadInt32(&n.uidPlusOne)
		}
	}
	return uidPlusOne - 1
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
	SemanticNodeBase
	Body SemanticNode
}

func NewLabelNode(body SemanticNode) *LabelNode {
	return &LabelNode{SemanticNodeBase: newSemanticNodeBase(SemanticLabelKind, "label"), Body: body}
}

type OpApplNode struct {
	SemanticNodeBase
	Operator             *SymbolNode
	Args                 []SemanticNode
	BdedQuantSymbolLists [][]*SymbolNode
	BdedQuantBounds      []SemanticNode
	BdedQuantATuple      []bool
	UnbdedQuantSymbols   []*SymbolNode
}

func NewOpApplNode(operator *SymbolNode, args ...SemanticNode) *OpApplNode {
	out := &OpApplNode{
		SemanticNodeBase: newSemanticNodeBase(SemanticOpApplKind, ""),
		Operator:         operator,
		Args:             append([]SemanticNode(nil), args...),
	}
	out.Image = out.String()
	return out
}

func NewBuiltinOpApplNode(op *UniqueString, args ...SemanticNode) *OpApplNode {
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
	SemanticNodeBase
	Lets     []*OpDefNode
	Bindings []LetBinding
	Body     SemanticNode
}

type LetBinding struct {
	Symbol *SymbolNode
	Value  any
}

func NewLetInNode(body SemanticNode, lets ...*OpDefNode) *LetInNode {
	return &LetInNode{
		SemanticNodeBase: newSemanticNodeBase(SemanticLetInKind, "LET"),
		Lets:             append([]*OpDefNode(nil), lets...),
		Body:             body,
	}
}

type Subst struct {
	Op   *SymbolNode
	Expr SemanticNode
	id   uint64
}

var nextSubstID atomic.Uint64

func ensureSubstIdentity(subst Subst) Subst {
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
	SemanticNodeBase
	Substs []Subst
	Body   SemanticNode
}

func NewSubstInNode(body SemanticNode, substs ...Subst) *SubstInNode {
	return &SubstInNode{
		SemanticNodeBase: newSemanticNodeBase(SemanticSubstInKind, "subst"),
		Substs:           copySubstsWithIdentity(substs),
		Body:             body,
	}
}

type APSubstInNode struct {
	SemanticNodeBase
	Substs []Subst
	Body   SemanticNode
}

func NewAPSubstInNode(body SemanticNode, substs ...Subst) *APSubstInNode {
	return &APSubstInNode{
		SemanticNodeBase: newSemanticNodeBase(SemanticAPSubstInKind, "ap-subst"),
		Substs:           copySubstsWithIdentity(substs),
		Body:             body,
	}
}

type ValueNode struct {
	SemanticNodeBase
	Value Value
}

func NewValueNode(value Value) *ValueNode {
	base := newSemanticNodeBase(SemanticValueKind, semanticValueString(value))
	return &ValueNode{
		SemanticNodeBase: base,
		Value:            value,
	}
}

type NumeralNode struct {
	SemanticNodeBase
	Value *IntValue
}

func NewNumeralNode(value int32) *NumeralNode {
	intValue := NewIntValue(value)
	base := newSemanticNodeBase(SemanticNumeralKind, strconv.FormatInt(int64(value), 10))
	base.ToolObject = intValue
	return &NumeralNode{
		SemanticNodeBase: base,
		Value:            intValue,
	}
}

type DecimalNode struct {
	SemanticNodeBase
	Value Value
}

func NewDecimalNode(value Value, image string) *DecimalNode {
	base := newSemanticNodeBase(SemanticDecimalKind, image)
	base.ToolObject = value
	return &DecimalNode{
		SemanticNodeBase: base,
		Value:            value,
	}
}

type StringNode struct {
	SemanticNodeBase
	Value *StringValue
}

func NewStringNode(value string) *StringNode {
	stringValue := NewStringValue(value)
	base := newSemanticNodeBase(SemanticStringKind, strconv.Quote(value))
	base.ToolObject = stringValue
	return &StringNode{
		SemanticNodeBase: base,
		Value:            stringValue,
	}
}

type AtNode struct {
	SemanticNodeBase
}

func NewAtNode() *AtNode {
	return &AtNode{SemanticNodeBase: newSemanticNodeBase(SemanticAtNodeKind, "@")}
}

type OpArgNode struct {
	SemanticNodeBase
	Op *SymbolNode
}

func NewOpArgNode(op *SymbolNode) *OpArgNode {
	return &OpArgNode{SemanticNodeBase: newSemanticNodeBase(SemanticOpArgKind, op.String()), Op: op}
}

type PossibleTrackNode struct {
	SemanticNodeBase
	Pred SemanticNode
	Name string
}

func NewPossibleTrackNode(pred SemanticNode, name string) *PossibleTrackNode {
	base := newSemanticNodeBase(SemanticPossibleTrackKind, "_Possible!_Track("+name+")")
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
	base := newSemanticNodeBase(SemanticPossibleCheckKind, "_Possible!_CheckName("+name+")")
	base.LevelValue = TLCLevelConstant
	return &PossibleCheckNode{
		SemanticNodeBase: base,
		Name:             name,
	}
}

type ThmOrAssumpDefNode struct {
	Name   *UniqueString
	Body   SemanticNode
	Params []*SymbolNode
}

func NewThmOrAssumpDefNode(name string, body SemanticNode, params ...*SymbolNode) *ThmOrAssumpDefNode {
	return &ThmOrAssumpDefNode{Name: UniqueStringOf(name), Body: body, Params: append([]*SymbolNode(nil), params...)}
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
	toolID := int64(0)
	if tool != nil {
		toolID = tool.ID
	}
	return SemanticToolObjectForToolID(toolID, node)
}

func SemanticToolObjectForToolID(toolID int64, node SemanticNode) any {
	key, ok := semanticToolObjectKeyForNode(toolID, node)
	if !ok {
		return nil
	}
	semanticToolObjects.RLock()
	value := semanticToolObjects.values[key]
	semanticToolObjects.RUnlock()
	return value
}

func SetSemanticToolObjectForTool(tool *Tool, node SemanticNode, value any) {
	toolID := int64(0)
	if tool != nil {
		toolID = tool.ID
	}
	SetSemanticToolObjectForToolID(toolID, node, value)
}

func SetSemanticToolObjectForToolID(toolID int64, node SemanticNode, value any) {
	key, ok := semanticToolObjectKeyForNode(toolID, node)
	if !ok {
		return
	}
	semanticToolObjects.Lock()
	if value == nil {
		delete(semanticToolObjects.values, key)
	} else {
		semanticToolObjects.values[key] = value
	}
	semanticToolObjects.Unlock()
}

func semanticToolObjectKeyForNode(toolID int64, node SemanticNode) (semanticToolObjectKey, bool) {
	if node == nil {
		return semanticToolObjectKey{}, false
	}
	nodeID := SemanticJavaHashCode(node)
	if withUID, ok := node.(interface{ GetUID() int32 }); ok {
		nodeID = withUID.GetUID()
	}
	return semanticToolObjectKey{toolID: toolID, nodeID: nodeID}, true
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
