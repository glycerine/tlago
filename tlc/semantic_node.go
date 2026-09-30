package tlc

import (
	"fmt"
	"strconv"
)

type SemanticNode any

type SemanticKind int

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
)

type SemanticNodeBase struct {
	KindValue  SemanticKind
	ToolObject any
	Image      string
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
	Lets []*OpDefNode
	Body SemanticNode
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
}

type SubstInNode struct {
	SemanticNodeBase
	Substs []Subst
	Body   SemanticNode
}

func NewSubstInNode(body SemanticNode, substs ...Subst) *SubstInNode {
	return &SubstInNode{
		SemanticNodeBase: SemanticNodeBase{KindValue: SemanticSubstInKind, Image: "subst"},
		Substs:           append([]Subst(nil), substs...),
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
		Substs:           append([]Subst(nil), substs...),
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

type ThmOrAssumpDefNode struct {
	Name *UniqueString
	Body SemanticNode
}

func NewThmOrAssumpDefNode(name string, body SemanticNode) *ThmOrAssumpDefNode {
	return &ThmOrAssumpDefNode{Name: UniqueStringOf(name), Body: body}
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
