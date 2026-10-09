package tlc

import (
	"fmt"
	"reflect"
)

func (e *distributedPayloadEncoder) stateVectorArray(vectors []*StateVec) (int, error) {
	if vectors == nil {
		return 0, nil
	}
	key := distributedByteArrayKey{reflect.ValueOf(vectors).Pointer(), len(vectors)}
	if id := e.stateVectorArrays[key]; id != 0 && len(vectors) != 0 {
		return id, nil
	}
	if e.stateVectorArrays == nil {
		e.stateVectorArrays = make(map[distributedByteArrayKey]int)
	}
	id := len(e.payload.StateVectorArrays) + 1
	e.stateVectorArrays[key] = id
	e.stateVectorArrayRoots = append(e.stateVectorArrayRoots, vectors)
	e.payload.StateVectorArrays = append(e.payload.StateVectorArrays, nil)
	refs := make([]int, len(vectors))
	for i, vector := range vectors {
		ref, err := e.stateVector(vector)
		if err != nil {
			return 0, fmt.Errorf("state vector %d: %w", i, err)
		}
		refs[i] = ref
	}
	e.payload.StateVectorArrays[id-1] = refs
	return id, nil
}

func (e *distributedPayloadEncoder) longVectorArray(vectors []*LongVec) int {
	if vectors == nil {
		return 0
	}
	key := distributedByteArrayKey{reflect.ValueOf(vectors).Pointer(), len(vectors)}
	if id := e.longVectorArrays[key]; id != 0 && len(vectors) != 0 {
		return id
	}
	if e.longVectorArrays == nil {
		e.longVectorArrays = make(map[distributedByteArrayKey]int)
	}
	id := len(e.payload.LongVectorArrays) + 1
	e.longVectorArrays[key] = id
	e.longVectorArrayRoots = append(e.longVectorArrayRoots, vectors)
	refs := make([]int, len(vectors))
	for i, vector := range vectors {
		refs[i] = e.longVector(vector)
	}
	e.payload.LongVectorArrays = append(e.payload.LongVectorArrays, refs)
	return id
}

func (d *distributedPayloadDecoder) stateVectorArray(id int) ([]*StateVec, error) {
	if id < 0 || id > len(d.stateVectorArrays) {
		return nil, fmt.Errorf("invalid state vector array reference %d", id)
	}
	if id == 0 {
		return nil, nil
	}
	return d.stateVectorArrays[id-1], nil
}

func (d *distributedPayloadDecoder) longVectorArray(id int) ([]*LongVec, error) {
	if id < 0 || id > len(d.longVectorArrays) {
		return nil, fmt.Errorf("invalid long vector array reference %d", id)
	}
	if id == 0 {
		return nil, nil
	}
	return d.longVectorArrays[id-1], nil
}
