package tlc

const (
	InitialPredicateNoAngle = "Initial predicate"
	InitialPredicate        = "<" + InitialPredicateNoAngle + ">"
)

type TLCStateInfo struct {
	StateNumber   int64
	Info          any
	State         *TLCStateMut
	FP            *uint64
	action        *Action
	original      *TLCStateMut
	aliased       bool // Retain the alias kind even when original is nil.
	stateCaptured bool
	stateOwner    *TLCStateMut
}

func NewTLCStateInfo(state *TLCStateMut, action ...*Action) *TLCStateInfo {
	if len(action) > 1 {
		panic(NewIllegalArgumentException("NewTLCStateInfo accepts at most one action"))
	}
	if state == nil {
		panic(NewNullPointerException())
	}
	info := &TLCStateInfo{State: state, stateCaptured: true, stateOwner: state}
	info.StateNumber = int64(state.Level())
	if len(action) > 0 {
		info.action = action[0]
		info.Info = actionInfo(state, action[0])
	} else if state.HasAction() {
		info.Info = actionInfo(state, state.GetAction())
	} else {
		info.Info = InitialPredicate
	}
	return info
}

func NewTLCStateInfoWithOrdinal(state *TLCStateMut, ordinal int) *TLCStateInfo {
	return &TLCStateInfo{State: state, StateNumber: int64(int32(ordinal)), Info: "", stateCaptured: true, stateOwner: state}
}

func AliasTLCStateInfo(state *TLCStateMut, info *TLCStateInfo) *TLCStateInfo {
	if info == nil {
		panic(NewNullPointerException())
	}
	value := info.Info
	if isNil(value) {
		value = nil
	}
	return &TLCStateInfo{
		State: state, Info: value, StateNumber: info.StateNumber,
		FP: info.FP, action: info.action, original: info.ownedState(), aliased: true,
		stateCaptured: true, stateOwner: state,
	}
}

func (i *TLCStateInfo) ownedState() *TLCStateMut {
	if i.stateCaptured {
		return i.stateOwner
	}
	return i.State
}

func actionInfo(state *TLCStateMut, action *Action) any {
	if action == nil {
		panic(NewNullPointerException())
	}
	if state.IsInitial() && !action.IsNamed() {
		return action.GetLocation(InitialPredicateNoAngle)
	}
	return action.GetLocation()
}

func (i *TLCStateInfo) FingerPrint() uint64 {
	if i == nil {
		panic(NewNullPointerException())
	}
	if i.FP == nil {
		fp := i.ownedState().FingerPrint()
		i.FP = &fp
	}
	return *i.FP
}

func (i *TLCStateInfo) String() string {
	if i == nil || i.ownedState() == nil {
		panic(NewNullPointerException())
	}
	return i.ownedState().String()
}

func (i *TLCStateInfo) Equal(other any) bool {
	if i == nil {
		panic(NewNullPointerException())
	}
	switch o := other.(type) {
	case *TLCStateInfo:
		if o == nil {
			return false
		}
		state := i.ownedState()
		if state == nil {
			panic(NewNullPointerException())
		}
		return state.Equal(o.ownedState())
	case TLCState:
		if isNil(o) {
			return false
		}
		state := i.ownedState()
		if state == nil {
			panic(NewNullPointerException())
		}
		return state.Equal(o)
	default:
		return false
	}
}

func (i *TLCStateInfo) GetOriginalState() *TLCStateMut {
	if i == nil {
		panic(NewNullPointerException())
	}
	if i.aliased {
		return i.original
	}
	return i.ownedState()
}

func (i *TLCStateInfo) OriginalState() *TLCStateMut {
	return i.GetOriginalState()
}

func (i *TLCStateInfo) GetAction() *Action {
	if i == nil {
		panic(NewNullPointerException())
	}
	if i.aliased {
		if i.original == nil {
			panic(NewNullPointerException())
		}
		if i.original.HasAction() {
			return i.original.GetAction()
		}
	}
	state := i.ownedState()
	if state == nil {
		panic(NewNullPointerException(`Cannot invoke "tlc2.tool.TLCState.hasAction()" because "this.state" is null`))
	}
	if state.HasAction() {
		return state.GetAction()
	}
	if i.action != nil {
		return i.action
	}
	return UnknownAction
}

func (i *TLCStateInfo) Action() *Action {
	return i.GetAction()
}

func (i *TLCStateInfo) GetStateNumber() int {
	if i == nil {
		panic(NewNullPointerException())
	}
	if i.aliased {
		if i.original == nil {
			panic(NewNullPointerException())
		}
		return i.original.Level()
	}
	return int(int32(i.StateNumber))
}

func (i *TLCStateInfo) HashCode() int32 {
	if i == nil || i.ownedState() == nil {
		panic(NewNullPointerException())
	}
	return i.ownedState().HashCode()
}

func (i *TLCStateInfo) ToRecordValue() Value {
	if i == nil {
		panic(NewNullPointerException())
	}
	state := i.ownedState()
	if state != nil && state.printRecord != nil {
		return state.printRecord
	}
	return NewRecordValueFromStateInfo(i)
}
