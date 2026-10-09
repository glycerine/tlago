package tlc

import "testing"

// No original method directly tests PrintTLCState's delegated ownership.
func TestPrintStateUnderlyingOwnership(t *testing.T) {
	states, _ := printStatePayloadStates(t)
	wrapper := states[0]
	owner := wrapper.printState
	if owner == nil || owner == wrapper {
		t.Fatal("print wrapper did not allocate a separate owner")
	}
	owner.UID, owner.WorkerID, owner.level = 99, 7, 8
	if wrapper.UID != 31 || wrapper.Level() != 4 || wrapper.HashCode() != owner.HashCode() {
		t.Fatal("wrapper metadata or delegated hash changed with owner metadata")
	}
	if !wrapper.Equal(owner) || owner.Equal(wrapper) || wrapper.Equal(wrapper) {
		t.Fatal("print equality lost source delegation and mutable-state type check")
	}
	shallow, deep := wrapper.Copy(), wrapper.DeepCopy()
	if shallow.printRecord != nil || shallow.printState != nil || shallow.Level() != TLCStateInitLevel || shallow.UID != TLCStateInitUID || deep.printRecord != nil || deep.Level() != 8 || deep.UID != 99 || deep.WorkerID != 7 {
		t.Fatal("delegated copies retained wrapper or copied wrapper metadata")
	}
	field := wrapper.printRecord.Names[0]
	if wrapper.Bind(field, NewIntValue(43)) != owner || wrapper.Lookup(field).(*IntValue).Val != 43 || wrapper.printRecord.Values[0].(*IntValue).Val != 42 {
		t.Fatal("bind did not return and mutate only the underlying state")
	}
	if wrapper.FingerPrint() != owner.FingerPrint() || !wrapper.AllAssigned() {
		t.Fatal("fingerprint or assignment did not delegate")
	}
	if wrapper.StringForVariables(nil) != owner.StringForVariables(nil, wrapper.printRecord.Names...) {
		t.Fatal("variable rendering did not delegate with record names")
	}
	if wrapper.Unbind(field) != owner || wrapper.AllAssigned() || len(wrapper.Unassigned()) != 1 || wrapper.Lookup(field).(*IntValue).Val != 42 {
		t.Fatal("unbind/lookup fallback did not preserve owner and record distinction")
	}
	vec := wrapper.AddToVec(NewStateVec(1))
	if vec.Size() != 1 || vec.At(0).printRecord != nil || vec.At(0).Level() != TLCStateInitLevel || wrapper.CreateEmpty().printRecord != nil {
		t.Fatal("collection or empty-state creation retained the wrapper")
	}
}

func TestDistributedPrintStateOwnerGraph(t *testing.T) {
	states, _ := printStatePayloadStates(t)
	wrapper, owner := states[0], states[0].printState
	owner.UID, owner.level = 99, 8
	owner.SetCached(9, &ModelValue{Data: wrapper})
	copied := distributedPayloadRoundTrip(t, []*TLCStateMut{wrapper, owner})
	if copied[0].printState != copied[1] || copied[1] == owner || copied[0].UID != 31 || copied[0].Level() != 4 || copied[1].UID != 99 || copied[1].Level() != 8 || copied[1].GetCached(9).(*ModelValue).Data != copied[0] {
		t.Fatal("owner transfer lost roots, cycle, independent metadata or receiver ownership")
	}
	field := copied[0].printRecord.Names[0]
	if copied[0].Bind(field, NewIntValue(44)) != copied[1] || copied[1].Lookup(field).(*IntValue).Val != 44 || owner.Lookup(field).(*IntValue).Val != 42 {
		t.Fatal("received delegated mutation lost owner sharing or touched sender")
	}
}

func TestDistributedPrintStateOwnerValidation(t *testing.T) {
	states, _ := printStatePayloadStates(t)
	for _, mutate := range []func(*DistributedStatePayload){
		func(p *DistributedStatePayload) { p.States[0].PrintState = -1 },
		func(p *DistributedStatePayload) { p.States[0].PrintState = len(p.States) + 1 },
		func(p *DistributedStatePayload) { p.States[0].PrintState = 0 },
		func(p *DistributedStatePayload) { p.States[0].PrintState = 1 },
		func(p *DistributedStatePayload) { p.States[0].PrintRecord = 0 },
		func(p *DistributedStatePayload) {
			p.States[p.States[0].PrintState-1].PrintRecord = p.States[0].PrintRecord
		},
		func(p *DistributedStatePayload) {
			p.StateCaches = append(p.StateCaches, []DistributedStateCacheEntry{})
			p.States[0].Cache = len(p.StateCaches)
		},
		func(p *DistributedStatePayload) {
			array := p.ValueArrays[p.States[0].ValuesArray-1]
			p.ValueArrays = append(p.ValueArrays, append([]int{}, array...))
			p.States[0].ValuesArray = len(p.ValueArrays)
		},
	} {
		payload, err := EncodeDistributedStates(states[:1])
		if err != nil {
			t.Fatal(err)
		}
		mutate(payload)
		if _, err := DecodeDistributedStates(payload); err == nil {
			t.Fatal("invalid wrapper/owner graph accepted")
		}
	}
}
