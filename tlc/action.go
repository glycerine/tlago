package tlc

import "reflect"

type CostModel struct {
	node *CostModelNode
}

type CostModelNode struct {
	Expr      SemanticNode
	Primary   int64
	Secondary int64
	Parent    *CostModelNode
	Children  *InsMap[semanticNodeKey, *CostModelNode]
	Lets      *InsMap[semanticNodeKey, *CostModelNode]
	Recursive *CostModelNode
	Primed    bool
	Level     int
}

var DoNotRecordCostModel = CostModel{}

func NewCostModel(expr SemanticNode) CostModel {
	return CostModel{node: newCostModelNode(expr, nil)}
}

func newCostModelNode(expr SemanticNode, parent *CostModelNode) *CostModelNode {
	level := 0
	if parent != nil {
		level = parent.Level + 1
	}
	return &CostModelNode{
		Expr:     expr,
		Parent:   parent,
		Children: NewInsMap[semanticNodeKey, *CostModelNode](),
		Lets:     NewInsMap[semanticNodeKey, *CostModelNode](),
		Level:    level,
	}
}

type semanticNodeKey struct {
	typ   string
	ptr   uintptr
	image string
}

func newSemanticNodeKey(node SemanticNode) semanticNodeKey {
	if node == nil {
		return semanticNodeKey{}
	}
	rv := reflect.ValueOf(node)
	if rv.IsValid() && rv.Kind() == reflect.Pointer && !rv.IsNil() {
		return semanticNodeKey{typ: rv.Type().String(), ptr: rv.Pointer()}
	}
	return semanticNodeKey{typ: rv.Type().String(), image: SemanticString(node)}
}

func sameSemanticNode(a SemanticNode, b SemanticNode) bool {
	return newSemanticNodeKey(a) == newSemanticNodeKey(b)
}

func (m CostModel) IncInvocations(values ...int64) CostModel {
	if m.node == nil {
		return m
	}
	value := int64(1)
	if len(values) > 0 {
		value = values[0]
	}
	m.node.Primary += value
	return m
}

func (m CostModel) IncSecondary(values ...int64) CostModel {
	if m.node == nil {
		return m
	}
	value := int64(1)
	if len(values) > 0 {
		value = values[0]
	}
	m.node.Secondary += value
	return m
}

func (m CostModel) GetPrimary() int64 {
	if m.node == nil {
		return -1
	}
	return m.node.Primary
}

func (m CostModel) GetSecondary() int64 {
	if m.node == nil {
		return -1
	}
	return m.node.Secondary
}

func (m CostModel) HasValues() bool {
	return m.node != nil
}

func (m CostModel) Report() CostModel {
	return m
}

func (m CostModel) Get(expr SemanticNode) CostModel {
	if m.node == nil {
		return m
	}
	if m.node.Children == nil {
		m.node.Children = NewInsMap[semanticNodeKey, *CostModelNode]()
	}
	if sameSemanticNode(expr, m.node.Expr) || SemanticKindOf(expr) != SemanticOpApplKind {
		return m
	}
	key := newSemanticNodeKey(expr)
	if child := m.node.Children.Get(key); child != nil {
		return CostModel{node: child}
	}
	if m.node.Recursive != nil && m.node.Recursive.Children != nil {
		if child := m.node.Recursive.Children.Get(key); child != nil {
			return CostModel{node: child}
		}
	}
	if m.node.Lets != nil {
		if child := m.node.Lets.Get(key); child != nil {
			return CostModel{node: child}
		}
	}
	return m
}

func (m CostModel) AddChild(expr SemanticNode) CostModel {
	if m.node == nil {
		return m
	}
	if m.node.Children == nil {
		m.node.Children = NewInsMap[semanticNodeKey, *CostModelNode]()
	}
	key := newSemanticNodeKey(expr)
	if child := m.node.Children.Get(key); child != nil {
		return CostModel{node: child}
	}
	child := newCostModelNode(expr, m.node)
	m.node.Children.Set(key, child)
	return CostModel{node: child}
}

func (m CostModel) AddLet(expr SemanticNode, child CostModel) CostModel {
	if m.node == nil || child.node == nil {
		return m
	}
	if m.node.Lets == nil {
		m.node.Lets = NewInsMap[semanticNodeKey, *CostModelNode]()
	}
	m.node.Lets.Set(newSemanticNodeKey(expr), child.node)
	return m
}

func (m CostModel) SetRecursive(recursive CostModel) CostModel {
	if m.node != nil {
		m.node.Recursive = recursive.node
	}
	return m
}

func (m CostModel) SetPrimed() CostModel {
	if m.node != nil {
		m.node.Primed = true
	}
	return m
}

func (m CostModel) IsPrimed() bool {
	return m.node != nil && m.node.Primed
}

func (m CostModel) SetLevel(level int) CostModel {
	if m.node != nil {
		m.node.Level = level
	}
	return m
}

func (m CostModel) GetLevel() int {
	if m.node == nil {
		return 0
	}
	return m.node.Level
}

func (m CostModel) GetAndIncrement(expr SemanticNode) CostModel {
	return m.Get(expr).IncInvocations()
}

func (m CostModel) GetRoot() CostModel {
	if m.node == nil {
		return m
	}
	ptr := m.node
	for ptr.Parent != nil {
		ptr = ptr.Parent
	}
	return CostModel{node: ptr}
}

func (m CostModel) GetChild() CostModel {
	if m.node == nil {
		return m
	}
	if m.node.Children != nil {
		for _, child := range m.node.Children.All() {
			return CostModel{node: child}
		}
	}
	return m
}

func (m CostModel) GetSubst(subst Subst) CostModel {
	return m.GetRoot().Get(subst.Expr)
}

type Action struct {
	Pred       SemanticNode
	Con        *Context
	CM         CostModel
	Name       string
	ID         int
	IsInitPred bool
	Internal   bool
}

var UnknownAction = &Action{Name: "Unknown"}

func (a *Action) IsNamed() bool {
	return a != nil && a.Name != ""
}

func (a *Action) GetName() string {
	if a == nil || a.Name == "" {
		return "Unknown"
	}
	return a.Name
}

func (a *Action) SetID(id int) {
	if a != nil {
		a.ID = id
	}
}

func (a *Action) GetID() int {
	if a == nil {
		return 0
	}
	return a.ID
}

func (a *Action) IsInitPredicate() bool {
	return a != nil && a.IsInitPred
}

func (a *Action) IsInternal() bool {
	return a != nil && a.Internal
}

type ActionItemList struct {
	Pred SemanticNode
	Con  *Context
	Kind int
	Next *ActionItemList
	CM   CostModel
	act  *Action
}

const (
	ActionItemConjunct = iota
	ActionItemPred
	ActionItemUnchanged
	ActionItemChanged
)

var EmptyActionItemList = &ActionItemList{}

func NewActionItemList(pred SemanticNode, con *Context, kind int, next *ActionItemList, cm CostModel) *ActionItemList {
	if next == nil {
		next = EmptyActionItemList
	}
	return &ActionItemList{Pred: pred, Con: con, Kind: kind, Next: next, CM: cm}
}

func (l *ActionItemList) CarPred() SemanticNode {
	return l.Pred
}

func (l *ActionItemList) CarContext() *Context {
	return l.Con
}

func (l *ActionItemList) CarKind() int {
	return l.Kind
}

func (l *ActionItemList) Cdr() *ActionItemList {
	return l.Next
}

func (l *ActionItemList) Cons(pred SemanticNode, con *Context, cm CostModel, kind int) *ActionItemList {
	return NewActionItemList(pred, con, kind, l, cm.Get(pred))
}

func (l *ActionItemList) ConsAction(act *Action, kind int) *ActionItemList {
	if act == nil {
		return NewActionItemList(nil, nil, kind, l, CostModel{})
	}
	return &ActionItemList{Pred: act.Pred, Con: act.Con, Kind: kind, Next: l, CM: act.CM.Get(act.Pred), act: act}
}

func (l *ActionItemList) IsEmpty() bool {
	return l == nil || l == EmptyActionItemList || (l.Pred == nil && l.Next == EmptyActionItemList && l.Kind == 0)
}

func (l *ActionItemList) SetAction(action *Action) {
	l.act = action
}

func (l *ActionItemList) GetAction() *Action {
	return l.act
}
