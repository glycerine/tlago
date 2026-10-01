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
	EvalFunc   LiveEvalFunc
	Positive   bool
	HasAct     bool
	LevelValue int
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
	return &LiveExprNode{Kind: LiveExprBool, Bool: value, Label: label, Positive: true, LevelValue: LiveLevelConstant}
}

func NewLNNeg(body *LiveExprNode) *LiveExprNode {
	return &LiveExprNode{Kind: LiveExprNeg, Body: body, Positive: body == nil || body.IsPositiveForm(), LevelValue: liveLevel(body), HasAct: liveContainsAction(body)}
}

func NewLNConj(nodes ...*LiveExprNode) *LiveExprNode {
	out := &LiveExprNode{Kind: LiveExprConj, Positive: true}
	for _, node := range nodes {
		out.AddConj(node)
	}
	return out
}

func NewLNDisj(nodes ...*LiveExprNode) *LiveExprNode {
	out := &LiveExprNode{Kind: LiveExprDisj, Positive: true}
	for _, node := range nodes {
		out.AddDisj(node)
	}
	return out
}

func newLNConjRaw(nodes ...*LiveExprNode) *LiveExprNode {
	out := &LiveExprNode{Kind: LiveExprConj, Bodies: append([]*LiveExprNode(nil), nodes...)}
	out.recomputeJunctionInfo()
	return out
}

func newLNDisjRaw(nodes ...*LiveExprNode) *LiveExprNode {
	out := &LiveExprNode{Kind: LiveExprDisj, Bodies: append([]*LiveExprNode(nil), nodes...)}
	out.recomputeJunctionInfo()
	return out
}

func NewLNAll(body *LiveExprNode) *LiveExprNode {
	return &LiveExprNode{Kind: LiveExprAll, Body: body, Positive: livePositive(body), LevelValue: LiveLevelTemporal, HasAct: liveContainsAction(body)}
}

func NewLNEven(body *LiveExprNode) *LiveExprNode {
	return &LiveExprNode{Kind: LiveExprEven, Body: body, Positive: livePositive(body), LevelValue: LiveLevelTemporal, HasAct: liveContainsAction(body)}
}

func NewLNNext(body *LiveExprNode) *LiveExprNode {
	return &LiveExprNode{Kind: LiveExprNext, Body: body, Positive: livePositive(body), LevelValue: LiveLevelTemporal, HasAct: true}
}

func NewLNState(label string, pred SemanticNode, con *Context, eval LiveEvalFunc) *LiveExprNode {
	if label == "" {
		label = "<state>"
	}
	return &LiveExprNode{Kind: LiveExprState, Pred: pred, Con: con, Label: label, EvalFunc: eval, Positive: true, LevelValue: LiveLevelState}
}

func NewLNAction(label string, pred SemanticNode, con *Context, eval LiveEvalFunc) *LiveExprNode {
	if label == "" {
		label = "<action>"
	}
	return &LiveExprNode{Kind: LiveExprAction, Pred: pred, Con: con, Label: label, EvalFunc: eval, Positive: true, LevelValue: LiveLevelAction, HasAct: true}
}

func (n *LiveExprNode) AddConj(elem *LiveExprNode) {
	if n == nil || elem == nil {
		return
	}
	if elem.Kind == LiveExprConj {
		for _, body := range elem.Bodies {
			n.AddConj(body)
		}
		return
	}
	n.Bodies = append(n.Bodies, elem)
	n.recomputeJunctionInfo()
}

func (n *LiveExprNode) AddDisj(elem *LiveExprNode) {
	if n == nil || elem == nil {
		return
	}
	if elem.Kind == LiveExprDisj {
		for _, body := range elem.Bodies {
			n.AddDisj(body)
		}
		return
	}
	n.Bodies = append(n.Bodies, elem)
	n.recomputeJunctionInfo()
}

func (n *LiveExprNode) Count() int {
	if n == nil {
		return 0
	}
	return len(n.Bodies)
}

func (n *LiveExprNode) GetBody(i ...int) *LiveExprNode {
	if n == nil {
		return nil
	}
	if len(i) == 0 {
		return n.Body
	}
	if i[0] < 0 || i[0] >= len(n.Bodies) {
		return nil
	}
	return n.Bodies[i[0]]
}

func (n *LiveExprNode) GetLevel() int {
	if n == nil {
		return LiveLevelConstant
	}
	return n.LevelValue
}

func (n *LiveExprNode) ContainAction() bool {
	return n != nil && n.HasAct
}

func (n *LiveExprNode) IsPositiveForm() bool {
	if n == nil {
		return true
	}
	switch n.Kind {
	case LiveExprNeg:
		return n.Body != nil && (n.Body.Kind == LiveExprBool || n.Body.Kind == LiveExprState)
	case LiveExprConj, LiveExprDisj:
		for _, body := range n.Bodies {
			if !body.IsPositiveForm() {
				return false
			}
		}
		return true
	default:
		return n.Positive
	}
}

func (n *LiveExprNode) Eval(tool *Tool, s1 *TLCStateMut, s2 *TLCStateMut) (bool, error) {
	if n == nil {
		return true, nil
	}
	switch n.Kind {
	case LiveExprBool:
		return n.Bool, nil
	case LiveExprNeg:
		value, err := n.Body.Eval(tool, s1, s2)
		return !value, err
	case LiveExprConj:
		for _, body := range n.Bodies {
			value, err := body.Eval(tool, s1, s2)
			if err != nil || !value {
				return value, err
			}
		}
		return true, nil
	case LiveExprDisj:
		for _, body := range n.Bodies {
			value, err := body.Eval(tool, s1, s2)
			if err != nil || value {
				return value, err
			}
		}
		return false, nil
	case LiveExprState, LiveExprAction:
		if n.EvalFunc != nil {
			return n.EvalFunc(tool, s1, s2)
		}
		if tool == nil {
			return true, nil
		}
		value, err := tool.Eval(n.Pred, n.Con, s1, s2, EvalClear)
		if err != nil {
			return false, err
		}
		boolValue, ok := value.(*BoolValue)
		if !ok {
			return false, newTLCError(ECGeneral, "liveness predicate %s evaluated to non-boolean %s", n.Label, value)
		}
		return boolValue.Val, nil
	default:
		return false, newTLCError(ECGeneral, "cannot directly evaluate temporal liveness formula %s", n)
	}
}

func (n *LiveExprNode) EvalOnLasso(tool *Tool, states []*TLCStateMut, cyclePos int, pos int) (bool, error) {
	if n == nil {
		return true, nil
	}
	if len(states) == 0 {
		return n.Eval(tool, nil, nil)
	}
	if cyclePos < 0 || cyclePos >= len(states) {
		cyclePos = 0
	}
	if pos < 0 {
		pos = 0
	}
	if pos >= len(states) {
		pos = cyclePos
	}
	nextState := func(i int) *TLCStateMut {
		if i+1 < len(states) {
			return states[i+1]
		}
		if cyclePos >= 0 && cyclePos < len(states) {
			return states[cyclePos]
		}
		return nil
	}
	switch n.Kind {
	case LiveExprAll:
		if pos < cyclePos {
			for i := pos; i < cyclePos; i++ {
				ok, err := n.Body.EvalOnLasso(tool, states, cyclePos, i)
				if err != nil || !ok {
					return ok, err
				}
			}
		}
		for i := cyclePos; i < len(states); i++ {
			ok, err := n.Body.EvalOnLasso(tool, states, cyclePos, i)
			if err != nil || !ok {
				return ok, err
			}
		}
		return true, nil
	case LiveExprEven:
		if pos < cyclePos {
			for i := pos; i < cyclePos; i++ {
				ok, err := n.Body.EvalOnLasso(tool, states, cyclePos, i)
				if err != nil || ok {
					return ok, err
				}
			}
		}
		for i := cyclePos; i < len(states); i++ {
			ok, err := n.Body.EvalOnLasso(tool, states, cyclePos, i)
			if err != nil || ok {
				return ok, err
			}
		}
		return false, nil
	case LiveExprNext:
		nextPos := pos + 1
		if nextPos >= len(states) {
			nextPos = cyclePos
		}
		return n.Body.EvalOnLasso(tool, states, cyclePos, nextPos)
	case LiveExprAction:
		return n.Eval(tool, states[pos], nextState(pos))
	default:
		return n.Eval(tool, states[pos], nextState(pos))
	}
}

func (n *LiveExprNode) PushNeg() *LiveExprNode {
	if n == nil {
		return NewLNNeg(nil)
	}
	switch n.Kind {
	case LiveExprBool:
		return NewLNBool(!n.Bool)
	case LiveExprNeg:
		return n.Body
	case LiveExprConj:
		out := NewLNDisj()
		for _, body := range n.Bodies {
			out.AddDisj(body.PushNeg())
		}
		return out
	case LiveExprDisj:
		out := NewLNConj()
		for _, body := range n.Bodies {
			out.AddConj(body.PushNeg())
		}
		return out
	case LiveExprAll:
		return NewLNEven(n.Body.PushNeg())
	case LiveExprEven:
		return NewLNAll(n.Body.PushNeg())
	default:
		return NewLNNeg(n)
	}
}

func (n *LiveExprNode) PushNegWith(hasNeg bool) *LiveExprNode {
	if n == nil {
		return NewLNNeg(nil)
	}
	if !hasNeg {
		switch n.Kind {
		case LiveExprBool:
			return n
		case LiveExprNeg:
			return n.Body.PushNegWith(true)
		case LiveExprConj:
			out := NewLNConj()
			for _, body := range n.Bodies {
				out.AddConj(body.PushNegWith(false))
			}
			return out
		case LiveExprDisj:
			out := NewLNDisj()
			for _, body := range n.Bodies {
				out.AddDisj(body.PushNegWith(false))
			}
			return out
		case LiveExprAll:
			return NewLNAll(n.Body.PushNegWith(false))
		case LiveExprEven:
			return NewLNEven(n.Body.PushNegWith(false))
		default:
			return n
		}
	}
	switch n.Kind {
	case LiveExprBool:
		return NewLNBool(!n.Bool)
	case LiveExprNeg:
		return n.Body.PushNegWith(false)
	case LiveExprConj:
		out := NewLNDisj()
		for _, body := range n.Bodies {
			out.AddDisj(body.PushNegWith(true))
		}
		return out
	case LiveExprDisj:
		out := NewLNConj()
		for _, body := range n.Bodies {
			out.AddConj(body.PushNegWith(true))
		}
		return out
	case LiveExprAll:
		return NewLNEven(n.Body.PushNegWith(true))
	case LiveExprEven:
		return NewLNAll(n.Body.PushNegWith(true))
	default:
		return NewLNNeg(n)
	}
}

func (n *LiveExprNode) Simplify() *LiveExprNode {
	if n == nil {
		return n
	}
	switch n.Kind {
	case LiveExprNeg:
		body := n.Body.Simplify()
		if body != nil && body.Kind == LiveExprBool {
			return NewLNBool(!body.Bool)
		}
		return NewLNNeg(body)
	case LiveExprConj:
		out := NewLNConj()
		for _, body := range n.Bodies {
			elem := body.Simplify()
			if elem.Kind == LiveExprBool {
				if !elem.Bool {
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
		for _, body := range n.Bodies {
			elem := body.Simplify()
			if elem.Kind == LiveExprBool {
				if elem.Bool {
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
		body := n.Body.Simplify()
		if body.Kind == LiveExprAll {
			body = body.Body
		}
		return NewLNAll(body)
	case LiveExprEven:
		body := n.Body.Simplify()
		if body.Kind == LiveExprEven {
			body = body.Body
		}
		return NewLNEven(body)
	default:
		return n
	}
}

func (n *LiveExprNode) ToDNF() *LiveExprNode {
	if n == nil {
		return nil
	}
	switch n.Kind {
	case LiveExprNeg:
		if n.Body != nil && (n.Body.Kind == LiveExprState || n.Body.Kind == LiveExprAction) {
			return n
		}
		return n.Body.PushNeg().ToDNF()
	case LiveExprConj:
		return n.conjToDNF()
	case LiveExprDisj:
		out := NewLNDisj()
		aeBodies := NewLNDisj()
		for _, body := range n.Bodies {
			elem := body.ToDNF()
			if elem.Kind == LiveExprDisj {
				for _, disj := range elem.Bodies {
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
		return nil
	}
	switch n.Kind {
	case LiveExprNeg:
		body := n.Body
		if body != nil && body.Kind == LiveExprNeg {
			return body.Body.FlattenSingleJunctions()
		}
		return NewLNNeg(body.FlattenSingleJunctions())
	case LiveExprConj:
		if n.Count() == 1 {
			return n.GetBody(0).FlattenSingleJunctions()
		}
		out := NewLNConj()
		for _, body := range n.Bodies {
			out.AddConj(body.FlattenSingleJunctions())
		}
		return out
	case LiveExprDisj:
		if n.Count() == 1 {
			return n.GetBody(0).FlattenSingleJunctions()
		}
		out := NewLNDisj()
		for _, body := range n.Bodies {
			out.AddDisj(body.FlattenSingleJunctions())
		}
		return out
	case LiveExprAll:
		return NewLNAll(n.Body.FlattenSingleJunctions())
	case LiveExprEven:
		return NewLNEven(n.Body.FlattenSingleJunctions())
	default:
		return n
	}
}

func (n *LiveExprNode) MakeBinary() *LiveExprNode {
	if n == nil {
		return nil
	}
	switch n.Kind {
	case LiveExprConj:
		if n.Count() == 0 {
			return n
		}
		if n.Count() == 1 {
			return n.GetBody(0).MakeBinary()
		}
		mid := n.Count() / 2
		left := NewLNConj(n.Bodies[:mid]...)
		right := NewLNConj(n.Bodies[mid:]...)
		return newLNConjRaw(left.MakeBinary(), right.MakeBinary())
	case LiveExprDisj:
		if n.Count() == 0 {
			return n
		}
		if n.Count() == 1 {
			return n.GetBody(0).MakeBinary()
		}
		mid := n.Count() / 2
		left := NewLNDisj(n.Bodies[:mid]...)
		right := NewLNDisj(n.Bodies[mid:]...)
		return newLNDisjRaw(left.MakeBinary(), right.MakeBinary())
	case LiveExprNeg:
		return NewLNNeg(n.Body.MakeBinary())
	case LiveExprAll:
		return NewLNAll(n.Body.MakeBinary())
	case LiveExprEven:
		return NewLNEven(n.Body.MakeBinary())
	default:
		return n
	}
}

func (n *LiveExprNode) TagExpr(tag int) int {
	if n == nil {
		return tag
	}
	switch n.Kind {
	case LiveExprState, LiveExprAction:
		n.Tag = tag
		return tag + 1
	case LiveExprConj, LiveExprDisj:
		for _, body := range n.Bodies {
			tag = body.TagExpr(tag)
		}
		return tag
	case LiveExprNeg, LiveExprAll, LiveExprEven, LiveExprNext:
		return n.Body.TagExpr(tag)
	default:
		return tag
	}
}

func (n *LiveExprNode) ExtractPromises() []*LiveExprNode {
	promises := make([]*LiveExprNode, 0)
	n.extractPromises(&promises)
	return promises
}

func (n *LiveExprNode) extractPromises(promises *[]*LiveExprNode) {
	if n == nil {
		return
	}
	switch n.Kind {
	case LiveExprEven:
		if !liveExprListContains(*promises, n) {
			*promises = append(*promises, n)
		}
		n.Body.extractPromises(promises)
	case LiveExprConj, LiveExprDisj:
		for _, body := range n.Bodies {
			body.extractPromises(promises)
		}
	case LiveExprNeg, LiveExprAll, LiveExprNext:
		n.Body.extractPromises(promises)
	}
}

func (n *LiveExprNode) GetAEBody() *LiveExprNode {
	if n != nil && n.Kind == LiveExprAll && n.Body != nil && n.Body.Kind == LiveExprEven {
		return n.Body.Body
	}
	return nil
}

func (n *LiveExprNode) GetEABody() *LiveExprNode {
	if n != nil && n.Kind == LiveExprEven && n.Body != nil && n.Body.Kind == LiveExprAll {
		return n.Body.Body
	}
	return nil
}

func (n *LiveExprNode) IsGeneralTF() bool {
	if n == nil {
		return true
	}
	switch n.Kind {
	case LiveExprAll:
		return !(n.Body != nil && n.Body.Kind == LiveExprEven)
	case LiveExprEven:
		return !(n.Body != nil && n.Body.Kind == LiveExprAll)
	case LiveExprConj, LiveExprDisj:
		for _, body := range n.Bodies {
			if !body.IsGeneralTF() {
				return false
			}
		}
		return true
	case LiveExprNeg, LiveExprNext:
		return n.Body.IsGeneralTF()
	default:
		return true
	}
}

func (n *LiveExprNode) Equal(other *LiveExprNode) bool {
	if n == nil || other == nil {
		return n == other
	}
	if n.Kind != other.Kind {
		return false
	}
	switch n.Kind {
	case LiveExprBool:
		return n.Bool == other.Bool
	case LiveExprState, LiveExprAction:
		return n.Tag == other.Tag
	}
	if !n.Body.Equal(other.Body) {
		return false
	}
	if len(n.Bodies) != len(other.Bodies) {
		return false
	}
	for i := range n.Bodies {
		if !n.Bodies[i].Equal(other.Bodies[i]) {
			return false
		}
	}
	return true
}

func (n *LiveExprNode) String() string {
	var b strings.Builder
	n.writeString(&b, "")
	return b.String()
}

func (n *LiveExprNode) ToDotViz() string {
	if n == nil {
		return ""
	}
	switch n.Kind {
	case LiveExprNeg:
		return "-" + n.Body.ToDotViz()
	case LiveExprAll:
		return "[]" + n.Body.ToDotViz()
	case LiveExprEven:
		return "<>" + n.Body.ToDotViz()
	default:
		return n.String()
	}
}

func (n *LiveExprNode) conjToDNF() *LiveExprNode {
	temp := make([]*LiveExprNode, len(n.Bodies))
	for i, body := range n.Bodies {
		temp[i] = body.ToDNF()
	}
	parts := make([]*LiveExprNode, 0, len(temp))
	total := 1
	for _, elem := range temp {
		switch elem.Kind {
		case LiveExprDisj:
			parts = append(parts, elem)
			total *= elem.Count()
		case LiveExprConj:
			parts = append(parts, elem.Bodies...)
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
		if part.Kind == LiveExprDisj {
			rCount /= part.Count()
			idx := 0
			for j := 0; j < num; j++ {
				for _, elem := range part.Bodies {
					for k := 0; k < rCount; k++ {
						res[idx].AddConj(elem)
						idx++
					}
				}
			}
			num *= part.Count()
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
		b.WriteString("<nil>")
		return
	}
	switch n.Kind {
	case LiveExprBool:
		if n.Bool {
			b.WriteString("TRUE")
		} else {
			b.WriteString("FALSE")
		}
	case LiveExprNeg:
		b.WriteByte('-')
		n.Body.writeString(b, padding+" ")
	case LiveExprConj, LiveExprDisj:
		op := "/\\"
		if n.Kind == LiveExprDisj {
			op = "\\/"
		}
		padding1 := padding + "    "
		for i, body := range n.Bodies {
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
		n.Body.writeString(b, padding+"  ")
	case LiveExprEven:
		b.WriteString("<>")
		n.Body.writeString(b, padding+"  ")
	case LiveExprNext:
		b.WriteString("()")
		n.Body.writeString(b, padding+"  ")
	case LiveExprState, LiveExprAction:
		b.WriteString(n.Label)
	default:
		b.WriteString("<live>")
	}
}

func (n *LiveExprNode) recomputeJunctionInfo() {
	level := LiveLevelConstant
	hasAct := false
	positive := true
	for _, body := range n.Bodies {
		if body.GetLevel() > level {
			level = body.GetLevel()
		}
		hasAct = hasAct || body.ContainAction()
		positive = positive && body.IsPositiveForm()
	}
	n.LevelValue = level
	n.HasAct = hasAct
	n.Positive = positive
}

func liveExprListContains(values []*LiveExprNode, target *LiveExprNode) bool {
	for _, value := range values {
		if value.Equal(target) {
			return true
		}
	}
	return false
}

func liveLevel(node *LiveExprNode) int {
	if node == nil {
		return LiveLevelConstant
	}
	return node.GetLevel()
}

func liveContainsAction(node *LiveExprNode) bool {
	return node != nil && node.ContainAction()
}

func livePositive(node *LiveExprNode) bool {
	return node == nil || node.IsPositiveForm()
}
