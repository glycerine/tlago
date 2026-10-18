package tlc

import (
	"fmt"
	"reflect"
	"strings"
	"sync/atomic"
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

	nodeIdentityCaptured   bool
	exprOwner              SemanticNode
	rootOwner              *CostModelNode
	actionIdentityCaptured bool
	actionOwner            *Action
	relationOwner          CoverageRelation
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
	node := &CostModelNode{
		Expr:     expr,
		Parent:   parent,
		Children: NewInsMap[semanticNodeKey, *CostModelNode](),
		Lets:     NewInsMap[semanticNodeKey, *CostModelNode](),
		Level:    level,
	}
	node.nodeIdentityCaptured = true
	node.exprOwner = expr
	node.rootOwner = node
	if parent != nil {
		node.rootOwner = parent.ownedRoot()
	}
	return node
}

func (n *CostModelNode) ownedExpr() SemanticNode {
	if n != nil && n.nodeIdentityCaptured {
		return n.exprOwner
	}
	if n == nil {
		return nil
	}
	return n.Expr
}

func (n *CostModelNode) ownedRoot() *CostModelNode {
	if n != nil && n.nodeIdentityCaptured {
		return n.rootOwner
	}
	if n == nil {
		return nil
	}
	root := n
	for root.Parent != nil {
		root = root.Parent
	}
	return root
}

func (n *CostModelNode) setActionIdentity(action *Action, relation CoverageRelation) {
	n.Action, n.Relation = action, relation
	n.actionOwner, n.relationOwner = action, relation
	n.actionIdentityCaptured = true
}

func (n *CostModelNode) ownedAction() *Action {
	if n != nil && n.actionIdentityCaptured {
		return n.actionOwner
	}
	if n == nil {
		return nil
	}
	return n.Action
}

func (n *CostModelNode) ownedRelation() CoverageRelation {
	if n != nil && n.actionIdentityCaptured {
		return n.relationOwner
	}
	if n == nil {
		return CoverageRelationUnknown
	}
	return n.Relation
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
	if semanticExploreNull(a) || semanticExploreNull(b) {
		return semanticExploreNull(a) && semanticExploreNull(b)
	}
	if newSemanticNodeKey(a) == newSemanticNodeKey(b) {
		return true
	}
	// A concrete definition and its split Go symbol view are one Java object.
	return SemanticEquals(a, b)
}

func (m CostModel) IncInvocations(values ...int64) CostModel {
	if len(values) > 1 {
		panic(NewIllegalArgumentException("IncInvocations accepts at most one value"))
	}
	if m.node == nil {
		return m
	}
	value := int64(1)
	if len(values) > 0 {
		value = values[0]
	}
	atomic.AddInt64(&m.node.Primary, value)
	return m
}

func (m CostModel) IncSecondary(values ...int64) CostModel {
	if len(values) > 1 {
		panic(NewIllegalArgumentException("IncSecondary accepts at most one value"))
	}
	if m.node == nil {
		return m
	}
	value := int64(1)
	if len(values) > 0 {
		value = values[0]
	}
	atomic.AddInt64(&m.node.Secondary, value)
	return m
}

func (m CostModel) incValueSecondary(values ...int64) {
	if m.node != nil && valueCoverageEnabled() {
		m.IncSecondary(values...)
	}
}

func (m CostModel) GetPrimary() int64 {
	if m.node == nil {
		return -1
	}
	return atomic.LoadInt64(&m.node.Primary)
}

func (m CostModel) GetSecondary() int64 {
	if m.node == nil {
		return -1
	}
	return atomic.LoadInt64(&m.node.Secondary)
}

func (m CostModel) HasValues() bool {
	return m.node != nil
}

// GetNode returns the semantic node wrapped by this cost model.
func (m CostModel) GetNode() SemanticNode {
	if m.node == nil {
		return nil
	}
	return m.node.ownedExpr()
}

// IsRoot reports whether this model is an action root or the node-less root
// used by OpApplNodeWrapper's test constructor.
func (m CostModel) IsRoot() bool {
	return m.node != nil && (m.node.ownedAction() != nil || m.node.ownedExpr() == nil)
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
	if m.node.ownedAction() != nil {
		// ActionWrapper.get resolves its predicate to a child wrapper. Only
		// OpApplNodeWrapper.get returns itself for its own expression.
		switch node := expr.(type) {
		case *SubstInNode:
			expr = node.GetBody()
		case *LetInNode:
			expr = node.Body
		}
		if child := m.node.Children.Get(newSemanticNodeKey(expr)); child != nil {
			return CostModel{node: child}
		}
		return m
	}
	if sameSemanticNode(expr, m.node.ownedExpr()) || SemanticKindOf(expr) != SemanticOpApplKind {
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
	key := newSemanticNodeKey(child.GetNode())
	if existing := m.node.Children.Get(key); existing != nil {
		return CostModel{node: existing}
	}
	child.node.Parent = m.node
	child.node.rootOwner = m.node.ownedRoot()
	child.node.nodeIdentityCaptured = true
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

func (m CostModel) AddLets(lets CostModel) CostModel {
	if m.node == nil {
		return m
	}
	if lets.node == nil {
		panic(NewNullPointerException())
	}
	if m.node.Lets == nil {
		m.node.Lets = NewInsMap[semanticNodeKey, *CostModelNode]()
	}
	m.node.Lets.Set(newSemanticNodeKey(lets.GetNode()), lets.node)
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
	level = int(int32(level))
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
	return CostModel{node: m.node.ownedRoot()}
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
	nameNull   bool
	OpDef      *OpDefNode
	ID         int
	IsInitPred bool
	Internal   bool
	Possible   *OpDefNode
	Auxiliary  map[any]any

	auxiliaryCaptured bool
	auxiliaryOwner    map[any]any
	classCaptured     bool
	initPredOwner     bool
	internalOwner     bool
	nameCaptured      bool
	nameOwner         string
	nameNullOwner     bool
	predConCaptured   bool
	predOwner         SemanticNode
	conOwner          *Context
	possibleCaptured  bool
	possibleOwner     *OpDefNode
}

const unnamedActionName = "UnnamedAction"

var UnknownAction = NewAction(NullSemanticNodeInstance, EmptyContext, unnamedActionName)

// NewUnnamedAction represents Java's constructor without an action name.
func NewUnnamedAction(pred SemanticNode, con *Context) *Action {
	return NewAction(pred, con, unnamedActionName)
}

// NewAction preserves the explicitly supplied action name, including empty names.
func NewAction(pred SemanticNode, con *Context, name string) *Action {
	auxiliary := make(map[any]any)
	return &Action{
		Pred: pred, Con: con, Name: name, CM: DoNotRecordCostModel,
		Auxiliary: auxiliary, auxiliaryCaptured: true, auxiliaryOwner: auxiliary,
		classCaptured: true, nameCaptured: true, nameOwner: name,
		predConCaptured: true, predOwner: pred, conOwner: con,
		possibleCaptured: true,
	}
}

func NewActionFromOpDef(pred SemanticNode, con *Context, opDef *OpDefNode, isInitPred bool, internal bool) *Action {
	name := unnamedActionName
	if opDef != nil && opDef.Name != nil {
		name = opDef.Name.String()
	}
	action := NewAction(pred, con, name)
	if opDef != nil && opDef.Name == nil {
		action.Name = ""
		action.nameNull = true
		action.nameOwner = ""
		action.nameNullOwner = true
	}
	action.OpDef = opDef
	action.IsInitPred = isInitPred
	action.Internal = internal
	action.initPredOwner = isInitPred
	action.internalOwner = internal
	return action
}

func (a *Action) require() *Action {
	if a == nil {
		panic(NewNullPointerException())
	}
	return a
}

func (a *Action) IsNamed() bool {
	a.require()
	name, null := a.ownedName()
	return !null && name != "" && name != unnamedActionName
}

func (a *Action) GetName() string {
	a.require()
	name, _ := a.ownedName()
	return name
}

// GetNameUnique preserves Java Action.getName's nullable UniqueString result.
// GetName remains the older string-returning Go compatibility surface.
func (a *Action) GetNameUnique() *UniqueString {
	a.require()
	name, null := a.ownedName()
	if null {
		return nil
	}
	return UniqueStringOf(name)
}

func (a *Action) nameForJavaFormat() string {
	if a.GetNameUnique() == nil {
		return "null"
	}
	name, _ := a.ownedName()
	return name
}

func (a *Action) ownedName() (string, bool) {
	if a.nameCaptured {
		return a.nameOwner, a.nameNullOwner
	}
	return a.Name, a.nameNull
}

func (a *Action) requireNameString() string {
	name := a.GetNameUnique()
	if name == nil {
		panic(NewNullPointerException())
	}
	return name.String()
}

func (a *Action) GetNameOfDefault() string {
	a.require()
	if a.IsNamed() {
		return a.GetName()
	}
	return a.String()
}

func (a *Action) String() string {
	a.require()
	pred := a.ownedPred()
	if pred == nil {
		panic(NewNullPointerException())
	}
	return "<Action " + semanticNodeJavaString(pred) + ">"
}

func (a *Action) GetPred() SemanticNode {
	a.require()
	if possible := a.ownedPossible(); possible != nil {
		return possible.Body
	}
	return a.ownedPred()
}

func (a *Action) ownedPossible() *OpDefNode {
	if a.possibleCaptured {
		return a.possibleOwner
	}
	return a.Possible
}

// ownedPred and ownedContext retain Java Action's public final pred and con
// fields. The fallback preserves package-literal compatibility for native tests.
func (a *Action) ownedPred() SemanticNode {
	if a.predConCaptured {
		return a.predOwner
	}
	return a.Pred
}

func (a *Action) ownedContext() *Context {
	if a.predConCaptured {
		return a.conOwner
	}
	return a.Con
}

func (a *Action) GetOpDef() *OpDefNode {
	a.require()
	return a.OpDef
}

func (a *Action) IsDeclared() bool {
	a.require()
	return a.OpDef != nil && !a.GetDeclaration().IsNull()
}

func (a *Action) GetDeclaration() SourceLocation {
	a.require()
	if a.OpDef == nil {
		return NullSourceLocation
	}
	// Java requires a declaration child in the operator's syntax. A builtin
	// or LAMBDA's definition location is not a declaration location.
	if loc := a.OpDef.GetDeclarationLocation(); !loc.IsNull() {
		return loc
	}
	return NullSourceLocation
}

// GetDeclarationLocation retains the older native spelling.
func (a *Action) GetDeclarationLocation() SourceLocation {
	return a.GetDeclaration()
}

func (a *Action) GetDefinition() SourceLocation {
	a.require()
	pred := a.ownedPred()
	if pred == nil {
		panic(NewNullPointerException())
	}
	if loc, ok := semanticNodeSourceLocation(pred); ok {
		return loc
	}
	return NullSourceLocation
}

// GetDefinitionLocation retains the older native spelling.
func (a *Action) GetDefinitionLocation() SourceLocation {
	return a.GetDefinition()
}

func (a *Action) GetParameters() *InsMap[*UniqueString, Value] {
	a.require()
	out := newInsMapAllowNilKeys[*UniqueString, Value]()
	if a.OpDef == nil {
		return out
	}
	for _, param := range a.OpDef.GetParams() {
		con := a.ownedContext()
		if con == nil {
			panic(NewNullPointerException())
		}
		value, ok := con.Lookup(param).(Value)
		if !ok || isNil(value) {
			continue
		}
		param.require()
		name := param.GetName()
		if _, exists := out.Get2(name); !exists {
			out.Set(name, value)
		}
	}
	return out
}

func (a *Action) GetInvocationSignature() string {
	a.require()
	params := a.GetParameters()
	name := a.nameForJavaFormat()
	if params.Len() == 0 {
		return name
	}
	parts := make([]string, 0, params.Len())
	for _, value := range params.All() {
		parts = append(parts, value.String())
	}
	return fmt.Sprintf("%s(%s)", name, strings.Join(parts, ","))
}

func (a *Action) GetAuxiliary() map[any]any {
	a.require()
	if a.auxiliaryCaptured {
		return a.auxiliaryOwner
	}
	if a.Auxiliary == nil {
		a.Auxiliary = make(map[any]any)
	}
	return a.Auxiliary
}

func (a *Action) SetID(id int) {
	a.require()
	a.ID = int(int32(id))
}

func (a *Action) GetID() int {
	a.require()
	return a.ID
}

func (a *Action) IsInitPredicate() bool {
	a.require()
	if a.classCaptured {
		return a.initPredOwner
	}
	return a.IsInitPred
}

func (a *Action) IsInternal() bool {
	a.require()
	if a.classCaptured {
		return a.internalOwner
	}
	return a.Internal
}

func (a *Action) IsPossible() bool {
	return a != nil && a.ownedPossible() != nil
}

type ActionItemList struct {
	Pred     SemanticNode
	Con      *Context
	Kind     int
	Next     *ActionItemList
	CM       CostModel
	act      *Action
	prev     *ActionItemList
	extended bool

	cellCaptured bool
	predOwner    SemanticNode
	conOwner     *Context
	kindOwner    int
	nextOwner    *ActionItemList
	cmOwner      CostModel
}

const (
	ActionItemConjunct  = 0
	ActionItemPred      = -1
	ActionItemUnchanged = -2
	ActionItemChanged   = -3
)

var EmptyActionItemList = newActionItemList(nil, nil, 0, nil, CostModel{}, false)

// Java uses ActionItemListExt only for getInitStates; ordinary next-state and
// enabled evaluation use ActionItemList with no previous link or action tracking.
var emptyActionItemListExt = newActionItemList(nil, nil, 0, nil, CostModel{}, true)

var actionItemListCoverage atomic.Int32

func init() {
	actionItemListCoverage.Store(-1)
}

func actionItemListCoverageEnabled() bool {
	if captured := actionItemListCoverage.Load(); captured >= 0 {
		return captured == 1
	}
	enabled := CoverageActionEnabled()
	captured := int32(0)
	if enabled {
		captured = 1
	}
	if actionItemListCoverage.CompareAndSwap(-1, captured) {
		return enabled
	}
	return actionItemListCoverage.Load() == 1
}

// InitializeActionItemListStatics represents a fresh Java classloader. The init
// singleton retains its previous/action links within one source runtime.
func InitializeActionItemListStatics() {
	EmptyActionItemList = newActionItemList(nil, nil, 0, nil, CostModel{}, false)
	emptyActionItemListExt = newActionItemList(nil, nil, 0, nil, CostModel{}, true)
	actionItemListCoverage.Store(-1)
}

func NewActionItemList(pred SemanticNode, con *Context, kind int, next *ActionItemList, cm CostModel) *ActionItemList {
	return newActionItemList(pred, con, kind, next, cm, false)
}

func newActionItemList(pred SemanticNode, con *Context, kind int, next *ActionItemList, cm CostModel, extended bool) *ActionItemList {
	kind = int(int32(kind))
	return &ActionItemList{
		Pred: pred, Con: con, Kind: kind, Next: next, CM: cm, extended: extended,
		cellCaptured: true, predOwner: pred, conOwner: con, kindOwner: kind, nextOwner: next, cmOwner: cm,
	}
}

func (l *ActionItemList) require() *ActionItemList {
	if l == nil {
		panic(NewNullPointerException())
	}
	return l
}

func (l *ActionItemList) CarPred() SemanticNode {
	l.require()
	if l.cellCaptured {
		return l.predOwner
	}
	return l.Pred
}

func (l *ActionItemList) CarContext() *Context {
	l.require()
	if l.cellCaptured {
		return l.conOwner
	}
	return l.Con
}

func (l *ActionItemList) CarKind() int {
	l.require()
	if l.cellCaptured {
		return l.kindOwner
	}
	return l.Kind
}

func (l *ActionItemList) ownedNext() *ActionItemList {
	if l.cellCaptured {
		return l.nextOwner
	}
	return l.Next
}

func (l *ActionItemList) ownedCostModel() CostModel {
	if l.cellCaptured {
		return l.cmOwner
	}
	return l.CM
}

func (l *ActionItemList) Cdr() *ActionItemList {
	l.require()
	next := l.ownedNext()
	if l.extended {
		if next == nil {
			panic(NewNullPointerException())
		}
		next.prev = l
	}
	return next
}

func (l *ActionItemList) Cons(pred SemanticNode, con *Context, cm CostModel, kind int) *ActionItemList {
	l.require()
	kind = int(int32(kind))
	itemCM := cm
	if actionItemListCoverageEnabled() {
		itemCM = cm.Get(pred)
	}
	item := newActionItemList(pred, con, kind, l, itemCM, l.extended)
	if l.extended {
		item.act = l.GetAction()
	}
	return item
}

func (l *ActionItemList) ConsAction(act *Action, kind int) *ActionItemList {
	l.require()
	kind = int(int32(kind))
	if act == nil {
		panic(NewNullPointerException())
	}
	itemCM := act.CM
	if actionItemListCoverageEnabled() {
		itemCM = act.CM.Get(l.CarPred())
	}
	item := newActionItemList(act.ownedPred(), act.ownedContext(), kind, l, itemCM, l.extended)
	item.act = act
	return item
}

func (l *ActionItemList) IsEmpty() bool {
	l.require()
	return l == EmptyActionItemList || l == emptyActionItemListExt
}

func (l *ActionItemList) SetAction(action *Action) {
	l.require()
	if !l.extended {
		return
	}
	l.act = action
}

func (l *ActionItemList) GetAction() *Action {
	l.require()
	if !l.extended {
		return nil
	}
	if l.prev != nil {
		return l.prev.act
	}
	return l.act
}
