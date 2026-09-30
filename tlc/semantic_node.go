package tlc

import (
	"fmt"
	"strconv"
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
	SemanticUnknownKind SemanticKind = iota
	SemanticLabelKind
	SemanticOpApplKind
	SemanticLetInKind
	SemanticSubstInKind
	SemanticAPSubstInKind
	SemanticNumeralKind
	SemanticDecimalKind
	SemanticStringKind
	SemanticAtNodeKind
	SemanticOpArgKind
	SemanticValueKind
	SemanticPossibleTrackKind
	SemanticPossibleCheckKind
)

type SemanticNodeBase struct {
	KindValue     SemanticKind
	ToolObject    any
	Image         string
	LevelValue    int
	LevelParamSet []*SymbolNode
}

func (n *SemanticNodeBase) Kind() SemanticKind {
	if n == nil {
		return SemanticUnknownKind
	}
	return n.KindValue
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
	if n == nil || n.Image == "" {
		return "<semantic node>"
	}
	return n.Image
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
	return &LabelNode{SemanticNodeBase: SemanticNodeBase{KindValue: SemanticLabelKind, Image: "label"}, Body: body}
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
		SemanticNodeBase: SemanticNodeBase{KindValue: SemanticOpApplKind},
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
		SemanticNodeBase: SemanticNodeBase{KindValue: SemanticLetInKind, Image: "LET"},
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
		SemanticNodeBase: SemanticNodeBase{KindValue: SemanticSubstInKind, Image: "subst"},
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
		SemanticNodeBase: SemanticNodeBase{KindValue: SemanticAPSubstInKind, Image: "ap-subst"},
		Substs:           copySubstsWithIdentity(substs),
		Body:             body,
	}
}

type ValueNode struct {
	SemanticNodeBase
	Value Value
}

func NewValueNode(value Value) *ValueNode {
	return &ValueNode{
		SemanticNodeBase: SemanticNodeBase{KindValue: SemanticValueKind, ToolObject: value, Image: semanticValueString(value)},
		Value:            value,
	}
}

type NumeralNode struct {
	SemanticNodeBase
	Value *IntValue
}

func NewNumeralNode(value int32) *NumeralNode {
	intValue := NewIntValue(value)
	return &NumeralNode{
		SemanticNodeBase: SemanticNodeBase{KindValue: SemanticNumeralKind, ToolObject: intValue, Image: strconv.FormatInt(int64(value), 10)},
		Value:            intValue,
	}
}

type DecimalNode struct {
	SemanticNodeBase
	Value Value
}

func NewDecimalNode(value Value, image string) *DecimalNode {
	return &DecimalNode{
		SemanticNodeBase: SemanticNodeBase{KindValue: SemanticDecimalKind, ToolObject: value, Image: image},
		Value:            value,
	}
}

type StringNode struct {
	SemanticNodeBase
	Value *StringValue
}

func NewStringNode(value string) *StringNode {
	stringValue := NewStringValue(value)
	return &StringNode{
		SemanticNodeBase: SemanticNodeBase{KindValue: SemanticStringKind, ToolObject: stringValue, Image: strconv.Quote(value)},
		Value:            stringValue,
	}
}

type AtNode struct {
	SemanticNodeBase
}

func NewAtNode() *AtNode {
	return &AtNode{SemanticNodeBase: SemanticNodeBase{KindValue: SemanticAtNodeKind, Image: "@"}}
}

type OpArgNode struct {
	SemanticNodeBase
	Op *SymbolNode
}

func NewOpArgNode(op *SymbolNode) *OpArgNode {
	return &OpArgNode{SemanticNodeBase: SemanticNodeBase{KindValue: SemanticOpArgKind, Image: op.String()}, Op: op}
}

type PossibleTrackNode struct {
	SemanticNodeBase
	Pred SemanticNode
	Name string
}

func NewPossibleTrackNode(pred SemanticNode, name string) *PossibleTrackNode {
	out := &PossibleTrackNode{
		SemanticNodeBase: SemanticNodeBase{
			KindValue:  SemanticPossibleTrackKind,
			Image:      "_Possible!_Track(" + name + ")",
			LevelValue: SemanticLevel(pred),
		},
		Pred: pred,
		Name: name,
	}
	out.SetLevelParams(SemanticLevelParams(pred)...)
	return out
}

type PossibleCheckNode struct {
	SemanticNodeBase
	Name string
}

func NewPossibleCheckNode(name string) *PossibleCheckNode {
	return &PossibleCheckNode{
		SemanticNodeBase: SemanticNodeBase{
			KindValue:  SemanticPossibleCheckKind,
			Image:      "_Possible!_CheckName(" + name + ")",
			LevelValue: TLCLevelConstant,
		},
		Name: name,
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
