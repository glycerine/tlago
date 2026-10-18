package tlc

import (
	"strings"
)

const (
	LiveLevelConstant = iota
	LiveLevelState
	LiveLevelAction
	LiveLevelTemporal
)

type LiveExprKind int

const (
	LiveExprBool LiveExprKind = iota
	LiveExprNeg
	LiveExprConj
	LiveExprDisj
	LiveExprAll
	LiveExprEven
	LiveExprNext
	LiveExprState
	LiveExprAction
)

type LiveEvalFunc func(tool *Tool, s1 *TLCStateMut, s2 *TLCStateMut) (bool, error)

type LiveExprNode struct {
	Kind       LiveExprKind
	Bool       bool
	Body       *LiveExprNode
	Bodies     []*LiveExprNode
	Pred       SemanticNode
	Con        *Context
	Tag        int
	Label      string
	stringFunc func() string
	dotVizFunc func() string
	EvalFunc   LiveEvalFunc
	Positive   bool
	HasAct     bool
	LevelValue int

	tagCaptured    bool
	tagOwner       int
	bodyCaptured   bool
	bodyOwner      *LiveExprNode
	atomCaptured   bool
	predOwner      SemanticNode
	conOwner       *Context
	evalOwner      LiveEvalFunc
	boolCaptured   bool
	boolOwner      bool
	bodiesCaptured bool
	bodiesOwner    []*LiveExprNode
	infoCaptured   bool
	levelOwner     int
	hasActOwner    bool
	positiveOwner  bool
	kindCaptured   bool
	kindOwner      LiveExprKind
}

var (
	LNTrue  = NewLNBool(true)
	LNFalse = NewLNBool(false)
)

func NewLNBool(value bool) *LiveExprNode {
	label := "FALSE"
	if value {
		label = "TRUE"
	}
	return &LiveExprNode{Kind: LiveExprBool, kindCaptured: true, kindOwner: LiveExprBool, Bool: value, Label: label, Positive: true, LevelValue: LiveLevelConstant, boolCaptured: true, boolOwner: value, infoCaptured: true, levelOwner: LiveLevelConstant, positiveOwner: true}
}

func NewLNNeg(body *LiveExprNode) *LiveExprNode {
	return &LiveExprNode{Kind: LiveExprNeg, kindCaptured: true, kindOwner: LiveExprNeg, Body: body, bodyCaptured: true, bodyOwner: body}
}

func NewLNConj(nodes ...*LiveExprNode) *LiveExprNode {
	bodies := append([]*LiveExprNode(nil), nodes...)
	out := &LiveExprNode{Kind: LiveExprConj, kindCaptured: true, kindOwner: LiveExprConj, Bodies: bodies, bodiesCaptured: true, bodiesOwner: bodies, infoCaptured: true}
	out.recomputeJunctionInfo()
	return out
}

func NewLNDisj(nodes ...*LiveExprNode) *LiveExprNode {
	bodies := append([]*LiveExprNode(nil), nodes...)
	out := &LiveExprNode{Kind: LiveExprDisj, kindCaptured: true, kindOwner: LiveExprDisj, Bodies: bodies, bodiesCaptured: true, bodiesOwner: bodies, infoCaptured: true}
	out.recomputeJunctionInfo()
	return out
}

func newLNConjRaw(nodes ...*LiveExprNode) *LiveExprNode {
	bodies := append([]*LiveExprNode(nil), nodes...)
	out := &LiveExprNode{Kind: LiveExprConj, kindCaptured: true, kindOwner: LiveExprConj, Bodies: bodies, bodiesCaptured: true, bodiesOwner: bodies, infoCaptured: true}
	out.recomputeJunctionInfo()
	return out
}

func newLNDisjRaw(nodes ...*LiveExprNode) *LiveExprNode {
	bodies := append([]*LiveExprNode(nil), nodes...)
	out := &LiveExprNode{Kind: LiveExprDisj, kindCaptured: true, kindOwner: LiveExprDisj, Bodies: bodies, bodiesCaptured: true, bodiesOwner: bodies, infoCaptured: true}
	out.recomputeJunctionInfo()
	return out
}

func NewLNAll(body *LiveExprNode) *LiveExprNode {
	return &LiveExprNode{Kind: LiveExprAll, kindCaptured: true, kindOwner: LiveExprAll, Body: body, LevelValue: LiveLevelTemporal, bodyCaptured: true, bodyOwner: body, infoCaptured: true, levelOwner: LiveLevelTemporal}
}

func NewLNEven(body *LiveExprNode) *LiveExprNode {
	return &LiveExprNode{Kind: LiveExprEven, kindCaptured: true, kindOwner: LiveExprEven, Body: body, LevelValue: LiveLevelTemporal, bodyCaptured: true, bodyOwner: body, infoCaptured: true, levelOwner: LiveLevelTemporal}
}

func NewLNNext(body *LiveExprNode) *LiveExprNode {
	return &LiveExprNode{Kind: LiveExprNext, kindCaptured: true, kindOwner: LiveExprNext, Body: body, LevelValue: LiveLevelAction, bodyCaptured: true, bodyOwner: body, infoCaptured: true, levelOwner: LiveLevelAction}
}

func NewLNState(label string, pred SemanticNode, con *Context, eval LiveEvalFunc) *LiveExprNode {
	if label == "" {
		label = "<state>"
	}
	return &LiveExprNode{Kind: LiveExprState, kindCaptured: true, kindOwner: LiveExprState, Pred: pred, Con: con, Label: label, EvalFunc: eval, Positive: true, LevelValue: LiveLevelState, tagCaptured: true, atomCaptured: true, predOwner: pred, conOwner: con, evalOwner: eval, infoCaptured: true, levelOwner: LiveLevelState, positiveOwner: true}
}

func NewLNAction(label string, pred SemanticNode, con *Context, eval LiveEvalFunc) *LiveExprNode {
	if label == "" {
		label = "<action>"
	}
	return &LiveExprNode{Kind: LiveExprAction, kindCaptured: true, kindOwner: LiveExprAction, Pred: pred, Con: con, Label: label, EvalFunc: eval, Positive: true, LevelValue: LiveLevelAction, HasAct: true, tagCaptured: true, atomCaptured: true, predOwner: pred, conOwner: con, evalOwner: eval, infoCaptured: true, levelOwner: LiveLevelAction, hasActOwner: true, positiveOwner: true}
}

func (n *LiveExprNode) AddConj(elem *LiveExprNode) {
	if n == nil {
		panic(NewNullPointerException())
	}
	if elem == nil {
		n.setBodies(append(n.ownedBodies(), nil))
		panic(NewNullPointerException())
	}
	if elem.ownedKind() == LiveExprConj {
		for _, body := range elem.ownedBodies() {
			n.AddConj(body)
		}
		return
	}
	n.setBodies(append(n.ownedBodies(), elem))
	n.recomputeJunctionInfo()
}

func (n *LiveExprNode) AddDisj(elem *LiveExprNode) {
	if n == nil {
		panic(NewNullPointerException())
	}
	if elem == nil {
		n.setBodies(append(n.ownedBodies(), nil))
		panic(NewNullPointerException())
	}
	if elem.ownedKind() == LiveExprDisj {
		for _, body := range elem.ownedBodies() {
			n.AddDisj(body)
		}
		return
	}
	n.setBodies(append(n.ownedBodies(), elem))
	n.recomputeJunctionInfo()
}

func (n *LiveExprNode) Count() int {
	if n == nil {
		panic(NewNullPointerException())
	}
	return len(n.ownedBodies())
}

func (n *LiveExprNode) GetBody(i ...int) *LiveExprNode {
	if n == nil {
		panic(NewNullPointerException())
	}
	if len(i) > 1 {
		panic(NewIllegalArgumentException("GetBody accepts at most one index"))
	}
	if len(i) == 0 {
		return n.ownedBody()
	}
	index := int(int32(i[0]))
	if index < 0 || index >= cap(n.ownedBodies()) {
		panic(NewArrayIndexOutOfBoundsException(index, cap(n.ownedBodies())))
	}
	return n.ownedBodies()[:cap(n.ownedBodies())][index]
}

func (n *LiveExprNode) ownedBody() *LiveExprNode {
	if n.bodyCaptured {
		return n.bodyOwner
	}
	return n.Body
}

func (n *LiveExprNode) ownedBool() bool {
	if n.boolCaptured {
		return n.boolOwner
	}
	return n.Bool
}

func (n *LiveExprNode) ownedBodies() []*LiveExprNode {
	if n.bodiesCaptured {
		return n.bodiesOwner
	}
	return n.Bodies
}

func (n *LiveExprNode) ownedKind() LiveExprKind {
	if n.kindCaptured {
		return n.kindOwner
	}
	return n.Kind
}

func (n *LiveExprNode) setBodies(bodies []*LiveExprNode) {
	n.Bodies = bodies
	if n.bodiesCaptured {
		n.bodiesOwner = bodies
	}
}

func (n *LiveExprNode) GetSemanticBody() SemanticNode {
	if n == nil {
		panic(NewNullPointerException())
	}
	if n.atomCaptured {
		return n.predOwner
	}
	return n.Pred
}

func (n *LiveExprNode) GetLevel() int {
	if n == nil {
		panic(NewNullPointerException())
	}
	if n.ownedKind() == LiveExprNeg {
		return n.GetBody().GetLevel()
	}
	if n.infoCaptured {
		return n.levelOwner
	}
	return n.LevelValue
}

func (n *LiveExprNode) GetContext() *Context {
	if n == nil {
		panic(NewNullPointerException())
	}
	if n.atomCaptured {
		return n.conOwner
	}
	return n.Con
}

func (n *LiveExprNode) ownedEvalFunc() LiveEvalFunc {
	if n.atomCaptured {
		return n.evalOwner
	}
	return n.EvalFunc
}

func (n *LiveExprNode) GetTag() int {
	if n == nil {
		panic(NewNullPointerException())
	}
	if n.tagCaptured {
		return n.tagOwner
	}
	return n.Tag
}

func (n *LiveExprNode) SetTag(tag int) {
	if n == nil {
		panic(NewNullPointerException())
	}
	tag = int(int32(tag))
	n.Tag = tag
	if n.tagCaptured {
		n.tagOwner = tag
	}
}

func (n *LiveExprNode) ContainAction() bool {
	if n == nil {
		panic(NewNullPointerException())
	}
	switch n.ownedKind() {
	case LiveExprNeg, LiveExprAll, LiveExprEven, LiveExprNext:
		return n.GetBody().ContainAction()
	default:
		if n.infoCaptured {
			return n.hasActOwner
		}
		return n.HasAct
	}
}

func (n *LiveExprNode) IsPositiveForm() bool {
	if n == nil {
		panic(NewNullPointerException())
	}
	switch n.ownedKind() {
	case LiveExprNeg:
		return n.GetBody() != nil && (n.GetBody().ownedKind() == LiveExprBool || n.GetBody().ownedKind() == LiveExprState)
	case LiveExprConj, LiveExprDisj:
		for _, body := range n.ownedBodies() {
			if !body.IsPositiveForm() {
				return false
			}
		}
		return true
	case LiveExprAll, LiveExprEven, LiveExprNext:
		return n.GetBody().IsPositiveForm()
	default:
		if n.infoCaptured {
			return n.positiveOwner
		}
		return n.Positive
	}
}

func (n *LiveExprNode) Eval(tool *Tool, s1 *TLCStateMut, s2 *TLCStateMut) (bool, error) {
	if n == nil {
		panic(NewNullPointerException())
	}
	switch n.ownedKind() {
	case LiveExprBool:
		return n.ownedBool(), nil
	case LiveExprNeg:
		value, err := n.GetBody().Eval(tool, s1, s2)
		return !value, err
	case LiveExprConj:
		for _, body := range n.ownedBodies() {
			value, err := body.Eval(tool, s1, s2)
			if err != nil || !value {
				return value, err
			}
		}
		return true, nil
	case LiveExprDisj:
		for _, body := range n.ownedBodies() {
			value, err := body.Eval(tool, s1, s2)
			if err != nil || value {
				return value, err
			}
		}
		return false, nil
	case LiveExprState, LiveExprAction:
		if eval := n.ownedEvalFunc(); eval != nil {
			return eval(tool, s1, s2)
		}
		if tool == nil {
			panic(NewNullPointerException())
		}
		var value Value
		var err error
		if n.ownedKind() == LiveExprState {
			value, err = tool.Eval(n.GetSemanticBody(), n.GetContext(), s1)
		} else {
			value, err = tool.Eval(n.GetSemanticBody(), n.GetContext(), s1, s2, EvalClear)
		}
		if err != nil {
			return false, err
		}
		boolValue, ok := value.(*BoolValue)
		if !ok {
			if n.ownedKind() == LiveExprState {
				return false, NewTLCRuntimeException(ECTLCLiveStatePredicateNonBool)
			}
			return false, NewTLCRuntimeException(ECTLCLiveEncounteredNonboolPredicate)
		}
		return boolValue.Val, nil
	case LiveExprNext:
		return n.GetBody().Eval(tool, s2, EmptyState)
	case LiveExprAll:
		return false, NewTLCRuntimeException(ECTLCLiveCannotEvalFormula, "[]")
	case LiveExprEven:
		return false, NewTLCRuntimeException(ECTLCLiveCannotEvalFormula, "<>")
	default:
		return false, newTLCError(ECGeneral, "cannot directly evaluate temporal liveness formula %s", n)
	}
}

func (n *LiveExprNode) EvalOnLasso(tool *Tool, states []*TLCStateMut, cyclePos int, pos int) (bool, error) {
	if n == nil {
		panic(NewNullPointerException())
	}
	cyclePos = int(int32(cyclePos))
	pos = int(int32(pos))
	switch n.ownedKind() {
	case LiveExprBool:
		return n.ownedBool(), nil
	case LiveExprNeg:
		value, err := n.GetBody().EvalOnLasso(tool, states, cyclePos, pos)
		return !value, err
	case LiveExprConj:
		for _, body := range n.ownedBodies() {
			value, err := body.EvalOnLasso(tool, states, cyclePos, pos)
			if err != nil || !value {
				return value, err
			}
		}
		return true, nil
	case LiveExprDisj:
		for _, body := range n.ownedBodies() {
			value, err := body.EvalOnLasso(tool, states, cyclePos, pos)
			if err != nil || value {
				return value, err
			}
		}
		return false, nil
	case LiveExprAll:
		if pos < cyclePos {
			for i := pos; i < cyclePos; i++ {
				ok, err := n.GetBody().EvalOnLasso(tool, states, cyclePos, i)
				if err != nil || !ok {
					return ok, err
				}
			}
		}
		for i, size := cyclePos, lassoStatesSize(states); i < size; i++ {
			ok, err := n.GetBody().EvalOnLasso(tool, states, cyclePos, i)
			if err != nil || !ok {
				return ok, err
			}
		}
		return true, nil
	case LiveExprEven:
		if pos < cyclePos {
			for i := pos; i < cyclePos; i++ {
				ok, err := n.GetBody().EvalOnLasso(tool, states, cyclePos, i)
				if err != nil || ok {
					return ok, err
				}
			}
		}
		for i, size := cyclePos, lassoStatesSize(states); i < size; i++ {
			ok, err := n.GetBody().EvalOnLasso(tool, states, cyclePos, i)
			if err != nil || ok {
				return ok, err
			}
		}
		return false, nil
	case LiveExprNext:
		panic(NewAssertionError("LNNext should have been rewritten by astToLive"))
	case LiveExprState:
		return n.Eval(tool, lassoStateAt(states, pos), EmptyState)
	case LiveExprAction:
		length := lassoStatesSize(states)
		s1 := lassoStateAt(states, pos)
		next := int(int32(pos) + 1)
		if next < length {
			return n.Eval(tool, s1, lassoStateAt(states, next))
		}
		return n.Eval(tool, s1, lassoStateAt(states, cyclePos))
	default:
		return false, newTLCError(ECGeneral, "cannot evaluate liveness formula %s on lasso", n)
	}
}

func lassoStateAt(states []*TLCStateMut, index int) *TLCStateMut {
	if states == nil {
		panic(NewNullPointerException())
	}
	if index < 0 || index >= len(states) {
		panic(NewIndexOutOfBoundsException(index, len(states)))
	}
	return states[index]
}

func lassoStatesSize(states []*TLCStateMut) int {
	if states == nil {
		panic(NewNullPointerException())
	}
	return len(states)
}

func (n *LiveExprNode) PushNeg() *LiveExprNode {
	if n == nil {
		panic(NewNullPointerException())
	}
	switch n.ownedKind() {
	case LiveExprBool:
		return NewLNBool(!n.ownedBool())
	case LiveExprNeg:
		return n.GetBody()
	case LiveExprConj:
		out := NewLNDisj()
		for _, body := range n.ownedBodies() {
			out.AddDisj(body.PushNeg())
		}
		return out
	case LiveExprDisj:
		out := NewLNConj()
		for _, body := range n.ownedBodies() {
			out.AddConj(body.PushNeg())
		}
		return out
	case LiveExprAll:
		return NewLNEven(n.GetBody().PushNeg())
	case LiveExprEven:
		return NewLNAll(n.GetBody().PushNeg())
	default:
		return NewLNNeg(n)
	}
}

func (n *LiveExprNode) PushNegWith(hasNeg bool) *LiveExprNode {
	if n == nil {
		panic(NewNullPointerException())
	}
	if !hasNeg {
		switch n.ownedKind() {
		case LiveExprBool:
			return n
		case LiveExprNeg:
			return n.GetBody().PushNegWith(true)
		case LiveExprConj:
			out := NewLNConj()
			for _, body := range n.ownedBodies() {
				out.AddConj(body.PushNegWith(false))
			}
			return out
		case LiveExprDisj:
			out := NewLNDisj()
			for _, body := range n.ownedBodies() {
				out.AddDisj(body.PushNegWith(false))
			}
			return out
		case LiveExprAll:
			return NewLNAll(n.GetBody().PushNegWith(false))
		case LiveExprEven:
			return NewLNEven(n.GetBody().PushNegWith(false))
		default:
			return n
		}
	}
	switch n.ownedKind() {
	case LiveExprBool:
		return NewLNBool(!n.ownedBool())
	case LiveExprNeg:
		return n.GetBody().PushNegWith(false)
	case LiveExprConj:
		out := NewLNDisj()
		for _, body := range n.ownedBodies() {
			out.AddDisj(body.PushNegWith(true))
		}
		return out
	case LiveExprDisj:
		out := NewLNConj()
		for _, body := range n.ownedBodies() {
			out.AddConj(body.PushNegWith(true))
		}
		return out
	case LiveExprAll:
		return NewLNEven(n.GetBody().PushNegWith(true))
	case LiveExprEven:
		return NewLNAll(n.GetBody().PushNegWith(true))
	default:
		return NewLNNeg(n)
	}
}

func (n *LiveExprNode) Simplify() *LiveExprNode {
	if n == nil {
		panic(NewNullPointerException())
	}
	switch n.ownedKind() {
	case LiveExprNeg:
		body := n.GetBody().Simplify()
		if body != nil && body.ownedKind() == LiveExprBool {
			return NewLNBool(!body.ownedBool())
		}
		return NewLNNeg(body)
	case LiveExprConj:
		out := NewLNConj()
		for _, body := range n.ownedBodies() {
			elem := body.Simplify()
			if elem.ownedKind() == LiveExprBool {
				if !elem.ownedBool() {
					return LNFalse
				}
				continue
			}
			out.AddConj(elem)
		}
		switch out.Count() {
		case 0:
			return LNTrue
		case 1:
			return out.GetBody(0)
		default:
			return out
		}
	case LiveExprDisj:
		out := NewLNDisj()
		for _, body := range n.ownedBodies() {
			elem := body.Simplify()
			if elem.ownedKind() == LiveExprBool {
				if elem.ownedBool() {
					return LNTrue
				}
				continue
			}
			out.AddDisj(elem)
		}
		switch out.Count() {
		case 0:
			return LNFalse
		case 1:
			return out.GetBody(0)
		default:
			return out
		}
	case LiveExprAll:
		body := n.GetBody().Simplify()
		if body.ownedKind() == LiveExprAll {
			body = body.GetBody()
		}
		return NewLNAll(body)
	case LiveExprEven:
		body := n.GetBody().Simplify()
		if body.ownedKind() == LiveExprEven {
			body = body.GetBody()
		}
		return NewLNEven(body)
	default:
		return n
	}
}

func (n *LiveExprNode) ToDNF() *LiveExprNode {
	if n == nil {
		panic(NewNullPointerException())
	}
	switch n.ownedKind() {
	case LiveExprNeg:
		if n.GetBody() != nil && (n.GetBody().ownedKind() == LiveExprState || n.GetBody().ownedKind() == LiveExprAction) {
			return n
		}
		if n.GetBody() != nil && n.GetBody().ownedKind() == LiveExprNext {
			// LNNext inherits LiveExprNode.pushNeg, so Java repeatedly rebuilds
			// this negation until it raises StackOverflowError.
			panic(NewStackOverflowError())
		}
		return n.GetBody().PushNeg().ToDNF()
	case LiveExprConj:
		return n.conjToDNF()
	case LiveExprDisj:
		out := NewLNDisj()
		aeBodies := NewLNDisj()
		for _, body := range n.ownedBodies() {
			elem := body.ToDNF()
			if elem.ownedKind() == LiveExprDisj {
				for _, disj := range elem.ownedBodies() {
					if ae := disj.GetAEBody(); ae != nil {
						aeBodies.AddDisj(ae)
					} else {
						out.AddDisj(disj)
					}
				}
				continue
			}
			if ae := elem.GetAEBody(); ae != nil {
				aeBodies.AddDisj(ae)
			} else {
				out.AddDisj(elem)
			}
		}
		switch aeBodies.Count() {
		case 0:
		case 1:
			out.AddDisj(NewLNAll(NewLNEven(aeBodies.GetBody(0))))
		default:
			out.AddDisj(NewLNAll(NewLNEven(aeBodies)))
		}
		return out
	default:
		return n
	}
}

func (n *LiveExprNode) FlattenSingleJunctions() *LiveExprNode {
	if n == nil {
		panic(NewNullPointerException())
	}
	switch n.ownedKind() {
	case LiveExprNeg:
		body := n.GetBody()
		if body != nil && body.ownedKind() == LiveExprNeg {
			return body.GetBody().FlattenSingleJunctions()
		}
		return NewLNNeg(body.FlattenSingleJunctions())
	case LiveExprConj:
		if n.Count() == 1 {
			return n.GetBody(0).FlattenSingleJunctions()
		}
		out := NewLNConj()
		for _, body := range n.ownedBodies() {
			out.AddConj(body.FlattenSingleJunctions())
		}
		return out
	case LiveExprDisj:
		if n.Count() == 1 {
			return n.GetBody(0).FlattenSingleJunctions()
		}
		out := NewLNDisj()
		for _, body := range n.ownedBodies() {
			out.AddDisj(body.FlattenSingleJunctions())
		}
		return out
	case LiveExprAll:
		return NewLNAll(n.GetBody().FlattenSingleJunctions())
	case LiveExprEven:
		return NewLNEven(n.GetBody().FlattenSingleJunctions())
	case LiveExprNext:
		return NewLNNext(n.GetBody().FlattenSingleJunctions())
	default:
		return n
	}
}

func (n *LiveExprNode) MakeBinary() *LiveExprNode {
	if n == nil {
		panic(NewNullPointerException())
	}
	switch n.ownedKind() {
	case LiveExprConj:
		if n.Count() == 0 {
			// Java recursively binaryizes two empty junctions until the JVM
			// raises StackOverflowError. Keep that failure catchable in Go.
			panic(NewStackOverflowError())
		}
		if n.Count() == 1 {
			return n.GetBody(0).MakeBinary()
		}
		mid := n.Count() / 2
		left := NewLNConj()
		right := NewLNConj()
		for _, body := range n.ownedBodies()[:mid] {
			left.AddConj(body)
		}
		for _, body := range n.ownedBodies()[mid:] {
			right.AddConj(body)
		}
		return newLNConjRaw(left.MakeBinary(), right.MakeBinary())
	case LiveExprDisj:
		if n.Count() == 0 {
			panic(NewStackOverflowError())
		}
		if n.Count() == 1 {
			return n.GetBody(0).MakeBinary()
		}
		mid := n.Count() / 2
		left := NewLNDisj()
		right := NewLNDisj()
		for _, body := range n.ownedBodies()[:mid] {
			left.AddDisj(body)
		}
		for _, body := range n.ownedBodies()[mid:] {
			right.AddDisj(body)
		}
		return newLNDisjRaw(left.MakeBinary(), right.MakeBinary())
	case LiveExprNeg:
		return NewLNNeg(n.GetBody().MakeBinary())
	case LiveExprAll:
		return NewLNAll(n.GetBody().MakeBinary())
	case LiveExprEven:
		return NewLNEven(n.GetBody().MakeBinary())
	case LiveExprNext:
		return NewLNNext(n.GetBody().MakeBinary())
	case LiveExprAction:
		panic(NewWrongInvocationException("LiveExprNode.makeBinary: TLC encounters actions."))
	default:
		return n
	}
}

func (n *LiveExprNode) TagExpr(tag int) int {
	if n == nil {
		panic(NewNullPointerException())
	}
	tag = int(int32(tag))
	switch n.ownedKind() {
	case LiveExprState, LiveExprAction:
		n.SetTag(tag)
		return int(int32(tag) + 1)
	case LiveExprConj, LiveExprDisj:
		for _, body := range n.ownedBodies() {
			tag = body.TagExpr(tag)
		}
		return tag
	case LiveExprNeg, LiveExprAll, LiveExprEven:
		return n.GetBody().TagExpr(tag)
	default:
		return tag
	}
}

func (n *LiveExprNode) ExtractPromises() []*LiveExprNode {
	promises := NewTBPar(10)
	n.ExtractPromisesInto(promises)
	exprs := promises.ownedExprs()
	result := make([]*LiveExprNode, len(exprs))
	copy(result, exprs)
	return result
}

func (n *LiveExprNode) ExtractPromisesInto(promises *TBPar) {
	if n == nil {
		panic(NewNullPointerException())
	}
	switch n.ownedKind() {
	case LiveExprEven:
		if !promises.Member(n) {
			promises.AddElement(n)
		}
		n.GetBody().ExtractPromisesInto(promises)
	case LiveExprConj, LiveExprDisj:
		n.GetBody(0).ExtractPromisesInto(promises)
		n.GetBody(1).ExtractPromisesInto(promises)
	case LiveExprNeg, LiveExprAll, LiveExprNext:
		n.GetBody().ExtractPromisesInto(promises)
	}
}

func (n *LiveExprNode) GetAEBody() *LiveExprNode {
	if n == nil {
		panic(NewNullPointerException())
	}
	if n.ownedKind() == LiveExprAll && n.GetBody() != nil && n.GetBody().ownedKind() == LiveExprEven {
		return n.GetBody().GetBody()
	}
	return nil
}

func (n *LiveExprNode) GetEABody() *LiveExprNode {
	if n == nil {
		panic(NewNullPointerException())
	}
	if n.ownedKind() == LiveExprEven && n.GetBody() != nil && n.GetBody().ownedKind() == LiveExprAll {
		return n.GetBody().GetBody()
	}
	return nil
}

func (n *LiveExprNode) IsGeneralTF() bool {
	if n == nil {
		panic(NewNullPointerException())
	}
	switch n.ownedKind() {
	case LiveExprAll:
		return !(n.GetBody() != nil && n.GetBody().ownedKind() == LiveExprEven)
	case LiveExprEven:
		return !(n.GetBody() != nil && n.GetBody().ownedKind() == LiveExprAll)
	case LiveExprConj, LiveExprDisj:
		for _, body := range n.ownedBodies() {
			if !body.IsGeneralTF() {
				return false
			}
		}
		return true
	case LiveExprNeg:
		return n.GetBody().IsGeneralTF()
	default:
		return true
	}
}

func (n *LiveExprNode) Equal(other *LiveExprNode) bool {
	if n == nil {
		panic(NewNullPointerException())
	}
	if other == nil {
		return false
	}
	if n.ownedKind() != other.ownedKind() {
		return false
	}
	switch n.ownedKind() {
	case LiveExprBool:
		return n.ownedBool() == other.ownedBool()
	case LiveExprState, LiveExprAction:
		return n.GetTag() == other.GetTag()
	case LiveExprConj, LiveExprDisj:
		if len(n.ownedBodies()) != len(other.ownedBodies()) {
			return false
		}
		for i := range n.ownedBodies() {
			if !n.ownedBodies()[i].Equal(other.ownedBodies()[i]) {
				return false
			}
		}
		return true
	default:
		return n.GetBody().Equal(other.GetBody())
	}
}

func (n *LiveExprNode) String() string {
	if n == nil {
		panic(NewNullPointerException())
	}
	var b strings.Builder
	n.writeString(&b, "")
	return b.String()
}

func (n *LiveExprNode) ToDotViz() string {
	if n == nil {
		panic(NewNullPointerException())
	}
	switch n.ownedKind() {
	case LiveExprNeg:
		return "-" + n.GetBody().ToDotViz()
	case LiveExprAll:
		return "[]" + n.GetBody().ToDotViz()
	case LiveExprEven:
		return "<>" + n.GetBody().ToDotViz()
	case LiveExprNext:
		return "()" + n.GetBody().ToDotViz()
	case LiveExprConj, LiveExprDisj:
		op := "/\\"
		if n.ownedKind() == LiveExprDisj {
			op = "\\/"
		}
		parts := make([]string, len(n.ownedBodies()))
		for i, body := range n.ownedBodies() {
			parts[i] = op + " (" + body.ToDotViz() + ")"
		}
		return strings.Join(parts, "\n")
	default:
		if n.dotVizFunc != nil {
			return n.dotVizFunc()
		}
		return n.String()
	}
}

func (n *LiveExprNode) conjToDNF() *LiveExprNode {
	temp := make([]*LiveExprNode, len(n.ownedBodies()))
	for i, body := range n.ownedBodies() {
		temp[i] = body.ToDNF()
	}
	parts := make([]*LiveExprNode, 0, len(temp))
	total := 1
	for _, elem := range temp {
		switch elem.ownedKind() {
		case LiveExprDisj:
			parts = append(parts, elem)
			// Java LNConj uses Math.multiplyExact on a signed 32-bit int,
			// rejecting overflow before allocating the cross product.
			product := int64(total) * int64(elem.Count())
			if product > 1<<31-1 || product < -1<<31 {
				panic(NewTLCRuntimeException(ECTLCLiveCannotHandleFormula,
					"because it exceeds the maximum supported size in disjunctive normal form."))
			}
			total = int(product)
		case LiveExprConj:
			parts = append(parts, elem.ownedBodies()...)
		default:
			parts = append(parts, elem)
		}
	}
	if total == 1 {
		return NewLNConj(parts...)
	}
	res := make([]*LiveExprNode, total)
	for i := range res {
		res[i] = NewLNConj()
	}
	num := 1
	rCount := total
	for _, part := range parts {
		if part.ownedKind() == LiveExprDisj {
			count := part.Count()
			if count == 0 {
				panic(NewArithmeticException("/ by zero"))
			}
			rCount /= count
			idx := 0
			for j := 0; j < num; j++ {
				for _, elem := range part.ownedBodies() {
					for k := 0; k < rCount; k++ {
						res[idx].AddConj(elem)
						idx++
					}
				}
			}
			num *= count
		} else {
			for _, conj := range res {
				conj.AddConj(part)
			}
		}
	}
	return NewLNDisj(res...)
}

func (n *LiveExprNode) writeString(b *strings.Builder, padding string) {
	if n == nil {
		panic(NewNullPointerException())
	}
	switch n.ownedKind() {
	case LiveExprBool:
		if n.ownedBool() {
			b.WriteString("TRUE")
		} else {
			b.WriteString("FALSE")
		}
	case LiveExprNeg:
		b.WriteByte('-')
		n.GetBody().writeString(b, padding+" ")
	case LiveExprConj, LiveExprDisj:
		op := "/\\"
		if n.ownedKind() == LiveExprDisj {
			op = "\\/"
		}
		padding1 := padding + "    "
		for i, body := range n.ownedBodies() {
			if i != 0 {
				b.WriteByte('\n')
				b.WriteString(padding)
			}
			b.WriteString(op)
			b.WriteString(" (")
			body.writeString(b, padding1)
			b.WriteByte(')')
		}
	case LiveExprAll:
		b.WriteString("[]")
		n.GetBody().writeString(b, padding+"  ")
	case LiveExprEven:
		b.WriteString("<>")
		n.GetBody().writeString(b, padding+"  ")
	case LiveExprNext:
		b.WriteString("()")
		n.GetBody().writeString(b, padding+"  ")
	case LiveExprState, LiveExprAction:
		if n.stringFunc != nil {
			b.WriteString(n.stringFunc())
		} else {
			b.WriteString(n.Label)
		}
	default:
		b.WriteString("<live>")
	}
}

func (n *LiveExprNode) recomputeJunctionInfo() {
	level := LiveLevelConstant
	hasAct := false
	positive := true
	for _, body := range n.ownedBodies() {
		if body.GetLevel() > level {
			level = body.GetLevel()
		}
		hasAct = hasAct || body.ContainAction()
		positive = positive && body.IsPositiveForm()
	}
	n.LevelValue = level
	n.HasAct = hasAct
	n.Positive = positive
	if n.infoCaptured {
		n.levelOwner = level
		n.hasActOwner = hasAct
		n.positiveOwner = positive
	}
}
