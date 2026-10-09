package tlc

import "fmt"

// DistributedResultPayload transfers a worker result using native Go graph
// references. Vector IDs are one-based; zero means a null vector. All state
// partitions share one state/value graph so aliases survive across partitions.
// Only active vector entries are sent, as in TLCStateVec and LongVec's source
// serialization contracts. Spare capacity is not part of a worker result.
// Array IDs address States.StateVectorArrays and States.LongVectorArrays,
// shared with attachments, including recursive state-vector graphs.
type DistributedResultPayload struct {
	Nil              bool
	ComputationTime  int64
	StatesComputed   int64
	StateArray       int
	FingerprintArray int
	States           *DistributedStatePayload
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
	}
	encoder := &distributedPayloadEncoder{}
	payload.States, err = encodeDistributedStates(nil, encoder)
	if err != nil {
		return nil, err
	}
	payload.StateArray, err = encoder.stateVectorArray(result.NextStates)
	if err != nil {
		return nil, err
	}
	payload.FingerprintArray = encoder.longVectorArray(result.NextFingerprints)
	return payload, nil
}

func DecodeDistributedResult(payload *DistributedResultPayload) (*NextStateResult, error) {
	if payload == nil {
		return nil, fmt.Errorf("missing distributed result payload")
	}
	if payload.Nil {
		if payload.StateArray != 0 || payload.FingerprintArray != 0 || payload.States != nil {
			return nil, fmt.Errorf("null result contains graph data")
		}
		return nil, nil
	}
	decoder := &distributedPayloadDecoder{}
	states, err := decodeDistributedStates(payload.States, decoder)
	if err != nil {
		return nil, err
	}
	if len(states) != 0 {
		return nil, fmt.Errorf("state graph has unused roots")
	}
	result := &NextStateResult{ComputationTime: payload.ComputationTime, StatesComputed: payload.StatesComputed}
	result.NextStates, err = decoder.stateVectorArray(payload.StateArray)
	if err != nil {
		return nil, err
	}
	result.NextFingerprints, err = decoder.longVectorArray(payload.FingerprintArray)
	if err != nil {
		return nil, err
	}
	return result, nil
}
