package tlc

import "fmt"

// Reserve the vector identity before walking states: an attached vector may
// contain a state whose model data points back to this same vector.
func (e *distributedPayloadEncoder) stateVector(vector *StateVec) (int, error) {
	if vector == nil {
		return 0, nil
	}
	if id := e.stateVectors[vector]; id != 0 {
		return id, nil
	}
	if e.stateVectors == nil {
		e.stateVectors = make(map[*StateVec]int)
	}
	id := len(e.payload.StateVectors) + 1
	e.stateVectors[vector] = id
	e.payload.StateVectors = append(e.payload.StateVectors, nil)
	refs := make([]int, vector.Size())
	for i, state := range vector.states {
		if state == nil {
			continue
		}
		mutable, ok := state.(*TLCStateMut)
		if !ok {
			return 0, fmt.Errorf("state vector entry %d: unsupported distributed state %T", i, state)
		}
		ref, err := e.state(mutable)
		if err != nil {
			return 0, err
		}
		refs[i] = ref
	}
	e.payload.StateVectors[id-1] = refs
	return id, nil
}
