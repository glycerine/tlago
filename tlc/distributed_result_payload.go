package tlc

import "fmt"

// DistributedResultPayload transfers a worker result using native Go graph
// references. Vector IDs are one-based; zero means a null vector. All state
// partitions share one state/value graph so aliases survive across partitions.
// Only active vector entries are sent, as in TLCStateVec and LongVec's source
// serialization contracts. Spare capacity is not part of a worker result.
// Vector IDs address States.StateVectors and States.LongVectors, shared with
// attachments, including recursive state-vector graphs.
type DistributedResultPayload struct {
	Nil                bool
	ComputationTime    int64
	StatesComputed     int64
	StatesNil          bool
	FingerprintsNil    bool
	StateVectors       []int
	FingerprintVectors []int
	States             *DistributedStatePayload
}

func EncodeDistributedResult(result *NextStateResult) (payload *DistributedResultPayload, err error) {
	defer func() {
		if failure := recover(); failure != nil {
			payload, err = nil, panicValueAsError(failure)
		}
	}()
	if result == nil {
		return &DistributedResultPayload{Nil: true}, nil
	}
	payload = &DistributedResultPayload{
		ComputationTime: result.ComputationTime, StatesComputed: result.StatesComputed,
		StatesNil: result.NextStates == nil, FingerprintsNil: result.NextFingerprints == nil,
		StateVectors:       make([]int, len(result.NextStates)),
		FingerprintVectors: make([]int, len(result.NextFingerprints)),
	}
	encoder := &distributedPayloadEncoder{}
	payload.States, err = encodeDistributedStates(nil, encoder)
	if err != nil {
		return nil, err
	}
	for i, vector := range result.NextStates {
		payload.StateVectors[i], err = encoder.stateVector(vector)
		if err != nil {
			return nil, fmt.Errorf("state vector %d: %w", i, err)
		}
	}
	for i, vector := range result.NextFingerprints {
		payload.FingerprintVectors[i] = encoder.longVector(vector)
	}
	return payload, nil
}

func DecodeDistributedResult(payload *DistributedResultPayload) (*NextStateResult, error) {
	if payload == nil {
		return nil, fmt.Errorf("missing distributed result payload")
	}
	if payload.Nil {
		if len(payload.StateVectors) != 0 || len(payload.FingerprintVectors) != 0 || payload.States != nil {
			return nil, fmt.Errorf("null result contains graph data")
		}
		return nil, nil
	}
	if payload.StatesNil && len(payload.StateVectors) != 0 || payload.FingerprintsNil && len(payload.FingerprintVectors) != 0 {
		return nil, fmt.Errorf("null result array contains vector references")
	}
	decoder := &distributedPayloadDecoder{}
	states, err := decodeDistributedStates(payload.States, decoder)
	if err != nil {
		return nil, err
	}
	if len(states) != 0 {
		return nil, fmt.Errorf("state graph has unused roots")
	}
	vectors := decoder.stateVectors
	fingerprints := decoder.longVectors
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
