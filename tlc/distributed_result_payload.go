package tlc

import "fmt"

// DistributedResultPayload transfers a worker result using native Go graph
// references. Vector IDs are one-based; zero means a null vector. All state
// partitions share one state/value graph so aliases survive across partitions.
// Only active vector entries are sent, as in TLCStateVec and LongVec's source
// serialization contracts. Spare capacity is not part of a worker result.
type DistributedResultPayload struct {
	Nil                bool
	ComputationTime    int64
	StatesComputed     int64
	StatesNil          bool
	FingerprintsNil    bool
	StateVectors       []int
	FingerprintVectors []int
	StateLengths       []int
	FingerprintNodes   [][]int64
	States             *DistributedStatePayload
}

func EncodeDistributedResult(result *NextStateResult) (*DistributedResultPayload, error) {
	if result == nil {
		return &DistributedResultPayload{Nil: true}, nil
	}
	payload := &DistributedResultPayload{
		ComputationTime: result.ComputationTime, StatesComputed: result.StatesComputed,
		StatesNil: result.NextStates == nil, FingerprintsNil: result.NextFingerprints == nil,
		StateVectors:       make([]int, len(result.NextStates)),
		FingerprintVectors: make([]int, len(result.NextFingerprints)),
	}
	var states []*TLCStateMut
	stateIDs := make(map[*StateVec]int)
	for i, vector := range result.NextStates {
		if vector == nil {
			continue
		}
		if id := stateIDs[vector]; id != 0 {
			payload.StateVectors[i] = id
			continue
		}
		id := len(payload.StateLengths) + 1
		stateIDs[vector] = id
		payload.StateVectors[i] = id
		payload.StateLengths = append(payload.StateLengths, vector.Size())
		for j, state := range vector.states {
			if state == nil {
				states = append(states, nil)
				continue
			}
			mutable, ok := state.(*TLCStateMut)
			if !ok {
				return nil, fmt.Errorf("state vector %d entry %d: unsupported distributed state %T", i, j, state)
			}
			states = append(states, mutable)
		}
	}
	fingerprintIDs := make(map[*LongVec]int)
	for i, vector := range result.NextFingerprints {
		if vector == nil {
			continue
		}
		id := fingerprintIDs[vector]
		if id == 0 {
			id = len(payload.FingerprintNodes) + 1
			fingerprintIDs[vector] = id
			payload.FingerprintNodes = append(payload.FingerprintNodes, append([]int64(nil), vector.data...))
		}
		payload.FingerprintVectors[i] = id
	}
	var err error
	payload.States, err = EncodeDistributedStates(states)
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func DecodeDistributedResult(payload *DistributedResultPayload) (*NextStateResult, error) {
	if payload == nil {
		return nil, fmt.Errorf("missing distributed result payload")
	}
	if payload.Nil {
		if len(payload.StateVectors) != 0 || len(payload.FingerprintVectors) != 0 || len(payload.StateLengths) != 0 || len(payload.FingerprintNodes) != 0 || payload.States != nil {
			return nil, fmt.Errorf("null result contains graph data")
		}
		return nil, nil
	}
	if payload.StatesNil && len(payload.StateVectors) != 0 || payload.FingerprintsNil && len(payload.FingerprintVectors) != 0 {
		return nil, fmt.Errorf("null result array contains vector references")
	}
	states, err := DecodeDistributedStates(payload.States)
	if err != nil {
		return nil, err
	}
	vectors := make([]*StateVec, len(payload.StateLengths))
	offset := 0
	for i, length := range payload.StateLengths {
		if length < 0 || length > len(states)-offset {
			return nil, fmt.Errorf("state vector %d has invalid length %d", i, length)
		}
		vectors[i] = NewStateVecFrom(states[offset : offset+length])
		offset += length
	}
	if offset != len(states) {
		return nil, fmt.Errorf("state graph has unused roots")
	}
	fingerprints := make([]*LongVec, len(payload.FingerprintNodes))
	for i, node := range payload.FingerprintNodes {
		fingerprints[i] = NewLongVecFrom(node)
	}
	result := &NextStateResult{ComputationTime: payload.ComputationTime, StatesComputed: payload.StatesComputed}
	if !payload.StatesNil {
		result.NextStates = make([]*StateVec, len(payload.StateVectors))
	}
	for i, id := range payload.StateVectors {
		if id < 0 || id > len(vectors) {
			return nil, fmt.Errorf("invalid state vector reference %d", id)
		}
		if id != 0 {
			result.NextStates[i] = vectors[id-1]
		}
	}
	if !payload.FingerprintsNil {
		result.NextFingerprints = make([]*LongVec, len(payload.FingerprintVectors))
	}
	for i, id := range payload.FingerprintVectors {
		if id < 0 || id > len(fingerprints) {
			return nil, fmt.Errorf("invalid fingerprint vector reference %d", id)
		}
		if id != 0 {
			result.NextFingerprints[i] = fingerprints[id-1]
		}
	}
	return result, nil
}
