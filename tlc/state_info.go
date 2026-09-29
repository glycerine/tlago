package tlc

const (
	InitialPredicateNoAngle = "Initial predicate"
	InitialPredicate        = "<" + InitialPredicateNoAngle + ">"
)

type TLCStateInfo struct {
	StateNumber int64
	Info        any
	State       *TLCStateMut
	FP          *uint64
	action      *Action
}

func NewTLCStateInfo(state *TLCStateMut, action ...*Action) *TLCStateInfo {
	info := &TLCStateInfo{State: state}
	if state != nil {
		info.StateNumber = int64(state.Level())
	}
	if len(action) > 0 {
		info.action = action[0]
		info.Info = actionInfo(state, action[0])
	} else if state != nil && state.HasAction() {
		info.Info = actionInfo(state, state.GetAction())
	} else {
		info.Info = InitialPredicate
	}
	return info
}

func NewTLCStateInfoWithOrdinal(state *TLCStateMut, ordinal int) *TLCStateInfo {
	return &TLCStateInfo{State: state, StateNumber: int64(ordinal), Info: ""}
}

func AliasTLCStateInfo(state *TLCStateMut, info *TLCStateInfo) *TLCStateInfo {
	out := &TLCStateInfo{State: state}
	if info != nil {
		out.Info = info.Info
		out.StateNumber = info.StateNumber
		out.FP = info.FP
		out.action = info.action
	}
	return out
}

func actionInfo(state *TLCStateMut, action *Action) any {
	if state != nil && state.IsInitial() && (action == nil || !action.IsNamed()) {
		return InitialPredicate
	}
	if action == nil {
		return UnknownAction.GetName()
	}
	return action.GetName()
}

func (i *TLCStateInfo) FingerPrint() uint64 {
	if i == nil || i.State == nil {
		return 0
	}
	if i.FP == nil {
		fp := i.State.FingerPrint()
		i.FP = &fp
	}
	return *i.FP
}

func (i *TLCStateInfo) String() string {
	if i == nil || i.State == nil {
		return ""
	}
	return i.State.String()
}

func (i *TLCStateInfo) Equal(other any) bool {
	if i == nil {
		return other == nil
	}
	switch o := other.(type) {
	case *TLCStateInfo:
		if o == nil {
			return false
		}
		return i.State.Equal(o.State)
	case *TLCStateMut:
		return i.State.Equal(o)
	default:
		return false
	}
}

func (i *TLCStateInfo) OriginalState() *TLCStateMut {
	if i == nil {
		return nil
	}
	return i.State
}

func (i *TLCStateInfo) Action() *Action {
	if i == nil {
		return UnknownAction
	}
	if i.State != nil && i.State.HasAction() {
		return i.State.GetAction()
	}
	if i.action != nil {
		return i.action
	}
	return UnknownAction
}

func (i *TLCStateInfo) GetStateNumber() int {
	if i == nil {
		return 0
	}
	return int(i.StateNumber)
}

func (i *TLCStateInfo) ToRecordValue() Value {
	if i == nil || i.State == nil {
		return EmptyRecord
	}
	return NewRecordValueFromInsMap(i.State.Values())
}
