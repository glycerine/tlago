package tlc

import (
	"fmt"
	"math"
	"reflect"
)

// Bits keeps signed integers and floating-point representations intact through
// the native transport, including negative zero and NaN payloads. These are
// finite primitive types, not a general object serialization mechanism.
type DistributedPrimitiveArrayNode struct {
	Kind string
	Bits []uint64
}

type distributedPrimitiveArrayKey struct {
	distributedByteArrayKey
	kind string
}

func encodeDistributedPrimitiveArray[T any](e *distributedPayloadEncoder, node *DistributedValueNode, values []T, kind string, bits func(T) uint64) {
	node.DataKind = kind
	if values == nil {
		return
	}
	if e.primitiveArrays == nil {
		e.primitiveArrays = make(map[distributedPrimitiveArrayKey]int)
	}
	key := distributedPrimitiveArrayKey{distributedByteArrayKey{reflect.ValueOf(values).Pointer(), len(values)}, kind}
	id := e.primitiveArrays[key]
	if id == 0 || len(values) == 0 {
		data := make([]uint64, len(values))
		for i, value := range values {
			data[i] = bits(value)
		}
		id = len(e.payload.PrimitiveArrays) + 1
		e.primitiveArrays[key] = id
		e.primitiveArrayRoots = append(e.primitiveArrayRoots, values)
		e.payload.PrimitiveArrays = append(e.payload.PrimitiveArrays, DistributedPrimitiveArrayNode{Kind: kind, Bits: data})
	}
	node.DataArray = id
}

func decodeDistributedPrimitiveBits[T any](node DistributedPrimitiveArrayNode, isNil bool, convert func(uint64) (T, bool)) ([]T, error) {
	if isNil {
		return nil, nil
	}
	values := make([]T, len(node.Bits))
	for i, bits := range node.Bits {
		value, valid := convert(bits)
		if !valid {
			return nil, fmt.Errorf("%s element %d outside primitive representation", node.Kind, i)
		}
		values[i] = value
	}
	return values, nil
}

func decodeDistributedPrimitiveArray(node DistributedPrimitiveArrayNode, isNil bool) (any, error) {
	switch node.Kind {
	case "boolArray":
		return decodeDistributedPrimitiveBits(node, isNil, func(v uint64) (bool, bool) { return v == 1, v <= 1 })
	case "intArray":
		return decodeDistributedPrimitiveBits(node, isNil, func(v uint64) (int, bool) { return int(v), int64(int(v)) == int64(v) })
	case "int8Array":
		return decodeDistributedPrimitiveBits(node, isNil, func(v uint64) (int8, bool) { return int8(v), int64(int8(v)) == int64(v) })
	case "int16Array":
		return decodeDistributedPrimitiveBits(node, isNil, func(v uint64) (int16, bool) { return int16(v), int64(int16(v)) == int64(v) })
	case "int32Array":
		return decodeDistributedPrimitiveBits(node, isNil, func(v uint64) (int32, bool) { return int32(v), int64(int32(v)) == int64(v) })
	case "int64Array":
		return decodeDistributedPrimitiveBits(node, isNil, func(v uint64) (int64, bool) { return int64(v), true })
	case "uint64Array":
		return decodeDistributedPrimitiveBits(node, isNil, func(v uint64) (uint64, bool) { return v, true })
	case "uint16Array":
		return decodeDistributedPrimitiveBits(node, isNil, func(v uint64) (uint16, bool) { return uint16(v), v <= math.MaxUint16 })
	case "float32Array":
		return decodeDistributedPrimitiveBits(node, isNil, func(v uint64) (float32, bool) { return math.Float32frombits(uint32(v)), v <= math.MaxUint32 })
	case "float64Array":
		return decodeDistributedPrimitiveBits(node, isNil, func(v uint64) (float64, bool) { return math.Float64frombits(v), true })
	default:
		return nil, fmt.Errorf("unknown primitive array kind %q", node.Kind)
	}
}
