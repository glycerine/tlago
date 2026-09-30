package tlc

type CostModel struct{}

func (CostModel) Get(SemanticNode) CostModel {
	return CostModel{}
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
