package tlc

import (
	"fmt"
	"reflect"
	"strings"
)

type CostModel struct {
	node *CostModelNode
}

type CostModelNode struct {
	Expr              SemanticNode
	Action            *Action
	Relation          CoverageRelation
	Primary           int64
	Secondary         int64
	SnapshotPrimary   int64
	SnapshotSecondary int64
	Parent            *CostModelNode
	Children          *InsMap[semanticNodeKey, *CostModelNode]
	Lets              *InsMap[semanticNodeKey, *CostModelNode]
	Substs            *InsMap[uint64, *CostModelNode]
	ChildCounts       []coveragePair
	Recursive         *CostModelNode
	Primed            bool
	Unchanged         bool
	Level             int
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

func (m CostModel) incValueSecondary(values ...int64) {
	if m.node != nil && CoverageEnabled() {
		m.IncSecondary(values...)
	}
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
	return m.report()
}

func (m CostModel) Get(expr SemanticNode) CostModel {
	if m.node == nil {
		return m
	}
	if m.node.Children == nil {
		m.node.Children = NewInsMap[semanticNodeKey, *CostModelNode]()
	}
	if m.node.Action != nil {
		// ActionWrapper.get resolves its predicate to a child wrapper. Only
		// OpApplNodeWrapper.get returns itself for its own expression.
		switch node := expr.(type) {
		case *SubstInNode:
			expr = node.Body
		case *LetInNode:
			expr = node.Body
		}
		if child := m.node.Children.Get(newSemanticNodeKey(expr)); child != nil {
			return CostModel{node: child}
		}
		return m
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

func (m CostModel) AddChildModel(child CostModel) CostModel {
	if m.node == nil || child.node == nil {
		return m
	}
	if m.node.Children == nil {
		m.node.Children = NewInsMap[semanticNodeKey, *CostModelNode]()
	}
	key := newSemanticNodeKey(child.node.Expr)
	if existing := m.node.Children.Get(key); existing != nil {
		return CostModel{node: existing}
	}
	child.node.Parent = m.node
	m.node.Children.Set(key, child.node)
	return child
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
	root := m.GetRoot()
	if root.node == nil || root.node.Substs == nil {
		return DoNotRecordCostModel
	}
	if child := root.node.Substs.Get(substIdentity(subst)); child != nil {
		return CostModel{node: child}
	}
	return DoNotRecordCostModel
}

type Action struct {
	Pred       SemanticNode
	Con        *Context
	CM         CostModel
	Name       string
	OpDef      *OpDefNode
	ID         int
	IsInitPred bool
	Internal   bool
	Possible   *OpDefNode
	Auxiliary  map[any]any
}

const unnamedActionName = "UnnamedAction"

var UnknownAction = &Action{Name: unnamedActionName}

func NewAction(pred SemanticNode, con *Context, name string) *Action {
	if con == nil {
		con = EmptyContext
	}
	if name == "" {
		name = unnamedActionName
	}
	return &Action{Pred: pred, Con: con, Name: name, CM: DoNotRecordCostModel}
}

func NewActionFromOpDef(pred SemanticNode, con *Context, opDef *OpDefNode, isInitPred bool, internal bool) *Action {
	name := ""
	if opDef != nil && opDef.Name != nil {
		name = opDef.Name.String()
	}
	action := NewAction(pred, con, name)
	action.OpDef = opDef
	action.IsInitPred = isInitPred
	action.Internal = internal
	return action
}

func (a *Action) IsNamed() bool {
	return a != nil && a.Name != "" && a.Name != unnamedActionName
}

func (a *Action) GetName() string {
	if a == nil || a.Name == "" {
		return unnamedActionName
	}
	return a.Name
}

func (a *Action) GetNameOfDefault() string {
	if a == nil {
		return "Unknown"
	}
	if a.IsNamed() {
		return a.GetName()
	}
	return a.String()
}

func (a *Action) String() string {
	if a == nil {
		return "<Action nil>"
	}
	return "<Action " + SemanticString(a.Pred) + ">"
}

func (a *Action) GetPred() SemanticNode {
	if a == nil {
		return nil
	}
	if a.Possible != nil && a.Possible.Body != nil {
		return a.Possible.Body
	}
	return a.Pred
}

func (a *Action) GetOpDef() *OpDefNode {
	if a == nil {
		return nil
	}
	return a.OpDef
}

func (a *Action) IsDeclared() bool {
	return a != nil && a.OpDef != nil && !a.GetDeclarationLocation().IsNull()
}

func (a *Action) GetDeclaration() string {
	if !a.IsDeclared() {
		return ""
	}
	return a.GetDeclarationLocation().String()
}

func (a *Action) GetDeclarationLocation() SourceLocation {
	if a == nil || a.OpDef == nil {
		return NullSourceLocation
	}
	if loc := a.OpDef.GetDeclarationLocation(); !loc.IsNull() {
		return loc
	}
	return semanticNodeLocation(a.OpDef)
}

func (a *Action) GetDefinition() string {
	if a == nil {
		return ""
	}
	if loc := a.GetDefinitionLocation(); !loc.IsNull() {
		return loc.String()
	}
	return SemanticString(a.Pred)
}

func (a *Action) GetDefinitionLocation() SourceLocation {
	if a == nil {
		return NullSourceLocation
	}
	if loc, ok := semanticNodeSourceLocation(a.Pred); ok {
		return loc
	}
	return NullSourceLocation
}

func (a *Action) GetParameters() *InsMap[*UniqueString, Value] {
	out := NewInsMap[*UniqueString, Value]()
	if a == nil || a.OpDef == nil || a.Con == nil {
		return out
	}
	for _, param := range a.OpDef.Params {
		if param == nil || param.Name == nil {
			continue
		}
		if value, ok := a.Con.Lookup(param).(Value); ok {
			out.Set(param.Name, value)
		}
	}
	return out
}

func (a *Action) GetInvocationSignature() string {
	if a == nil {
		return "Unknown"
	}
	params := a.GetParameters()
	if params.Len() == 0 {
		return a.GetName()
	}
	parts := make([]string, 0, params.Len())
	for _, value := range params.All() {
		parts = append(parts, value.String())
	}
	return fmt.Sprintf("%s(%s)", a.GetName(), strings.Join(parts, ","))
}

func (a *Action) GetAuxiliary() map[any]any {
	if a == nil {
		return nil
	}
	if a.Auxiliary == nil {
		a.Auxiliary = make(map[any]any)
	}
	return a.Auxiliary
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

func (a *Action) IsPossible() bool {
	return a != nil && a.Possible != nil
}

type ActionItemList struct {
	Pred SemanticNode
	Con  *Context
	Kind int
	Next *ActionItemList
	CM   CostModel
	act  *Action
	prev *ActionItemList
}

const (
	ActionItemConjunct  = 0
	ActionItemPred      = -1
	ActionItemUnchanged = -2
	ActionItemChanged   = -3
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
	if l == nil || l.Next == nil {
		return EmptyActionItemList
	}
	l.Next.prev = l
	return l.Next
}

func (l *ActionItemList) Cons(pred SemanticNode, con *Context, cm CostModel, kind int) *ActionItemList {
	itemCM := cm
	if CoverageActionEnabled() {
		itemCM = cm.Get(pred)
	}
	item := NewActionItemList(pred, con, kind, l, itemCM)
	item.act = l.GetAction()
	return item
}

func (l *ActionItemList) ConsAction(act *Action, kind int) *ActionItemList {
	if act == nil {
		item := NewActionItemList(nil, nil, kind, l, CostModel{})
		item.act = l.GetAction()
		return item
	}
	itemCM := act.CM
	if CoverageActionEnabled() {
		itemCM = act.CM.Get(l.Pred)
	}
	return &ActionItemList{Pred: act.Pred, Con: act.Con, Kind: kind, Next: l, CM: itemCM, act: act}
}

func (l *ActionItemList) IsEmpty() bool {
	return l == nil || l == EmptyActionItemList || (l.Pred == nil && l.Next == EmptyActionItemList && l.Kind == 0)
}

func (l *ActionItemList) SetAction(action *Action) {
	if l == nil {
		return
	}
	l.act = action
}

func (l *ActionItemList) GetAction() *Action {
	if l == nil {
		return nil
	}
	if l.prev != nil {
		return l.prev.act
	}
	return l.act
}
