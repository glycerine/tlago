package tlc

import (
	"fmt"
	"math"
	"reflect"
	"sort"
)

// DistributedStatePayload is a native Go graph for a single remote invocation.
// IDs are one-based, with zero reserved for null. Objects shared across states
// remain shared after decoding; receiver objects do not alias sender objects.
// State levels use the in-memory int32 range, not the short disk-queue format.
type DistributedStatePayload struct {
	Nil             bool
	Roots           []int
	RootsArray      int
	States          []DistributedStateNode
	StateArrays     [][]int
	StateVectors    [][]int
	Values          []DistributedValueNode
	Strings         []DistributedStringNode
	ByteArrays      [][]byte
	ValueArrays     [][]int
	ValueRows       [][]int
	ValueVectors    []DistributedValueVectorNode
	LongVectors     [][]int64
	BitVectors      []int // Primitive word-array references, retaining full storage.
	NameArrays      [][]int
	StateCaches     [][]DistributedStateCacheEntry
	ValueMaps       [][]DistributedValueMapEntry
	ObjectArrays    [][]DistributedObjectDataNode
	ObjectMaps      [][]DistributedObjectMapEntry
	ObjectKeyMaps   [][]DistributedObjectKeyMapEntry
	PrimitiveArrays []DistributedPrimitiveArrayNode
	StringArrays    [][]string

	StateVectorArrays [][]int
	LongVectorArrays  [][]int
}

// Mixed attachment entries carry the same tags as scalar model data without
// allocating unrelated value representation/cache fields for every entry.
type DistributedObjectDataNode struct {
	Kind      string
	String    string
	Name      int
	Integer   int64
	FloatBits uint64
	Bool      bool
	Bytes     int
	Value     int
	State     int
	Array     int
	Map       int
}

func distributedObjectDataNode(node DistributedValueNode) DistributedObjectDataNode {
	return DistributedObjectDataNode{Kind: node.DataKind, String: node.DataString, Name: node.DataName,
		Integer: node.DataInteger, FloatBits: node.DataFloatBits, Bool: node.DataBool,
		Bytes: node.DataBytes, Value: node.DataValue, State: node.DataState, Array: node.DataArray, Map: node.DataMap}
}

func (node DistributedObjectDataNode) valueNode() DistributedValueNode {
	return DistributedValueNode{DataKind: node.Kind, DataString: node.String, DataName: node.Name,
		DataInteger: node.Integer, DataFloatBits: node.FloatBits, DataBool: node.Bool,
		DataBytes: node.Bytes, DataValue: node.Value, DataState: node.State, DataArray: node.Array, DataMap: node.Map}
}

type DistributedObjectMapEntry struct {
	Key  string
	Data DistributedObjectDataNode
}

type DistributedObjectKeyMapEntry struct {
	Key  DistributedObjectDataNode
	Data DistributedObjectDataNode
}

type DistributedValueMapEntry struct {
	Key   string
	Value int
}

type DistributedStateCacheEntry struct {
	Key   int32
	Value int
}

type DistributedValueVectorNode struct {
	Array int
	Count int
}

type DistributedStateNode struct {
	WorkerID    int16
	UID         int64
	Level       int32
	Values      []int
	ValuesArray int
	ValuesNil   bool
	Predecessor int
	Cache       int
	PrintRecord int
}

type DistributedStringNode struct {
	Text         string
	Token        int
	Location     int
	Unregistered bool
}

type DistributedValueReferences struct {
	Nil        bool
	References []int
	Array      int
}

// Value nodes retain representation and caches for symbolic set constructors.
// Function/predicate/lazy wrappers use the same materialization contracts as
// their source network serialization, without serializing evaluator machinery.
type DistributedValueNode struct {
	Kind                string
	OperatorDomain      []DistributedValueReferences
	OperatorDomainNil   bool
	OperatorDomainArray int
	References          []int
	ReferencesArray     int
	ReferencesNil       bool
	Domain              []int
	DomainArray         int
	DomainNil           bool
	Names               []int
	NamesArray          int
	NamesNil            bool
	Flag                bool
	Cache               int
	Dummy               bool
	Integer             int64
	Low                 int32
	High                int32
	String              int
	ModelIndex          int
	ModelType           rune
	CollectionPresent   bool
	Vector              int
	DataKind            string
	DataString          string
	DataName            int
	DataInteger         int64
	DataFloatBits       uint64
	DataBool            bool
	DataBytes           int
	DataValue           int
	DataState           int
	DataArray           int
	DataMap             int
}

type distributedPayloadEncoder struct {
	payload             *DistributedStatePayload
	states              map[*TLCStateMut]int
	stateArrays         map[distributedByteArrayKey]int
	stateArrayRoots     [][]*TLCStateMut
	stateVectors        map[*StateVec]int
	values              map[Value]int
	strings             map[*UniqueString]int
	bytes               map[distributedByteArrayKey]int
	arrays              map[distributedByteArrayKey]int
	arrayRoots          [][]Value // Keep address-keyed backing storage alive while encoding.
	rows                map[distributedByteArrayKey]int
	rowRoots            [][][]Value
	vectors             map[*ValueVec]int
	longVectors         map[*LongVec]int
	bitVectors          map[*BitVector]int
	nameArrays          map[distributedByteArrayKey]int
	nameRoots           [][]*UniqueString
	caches              map[uintptr]int
	cacheRoots          []map[int]Value // Retain address-keyed maps throughout encoding.
	valueMaps           map[uintptr]int
	valueMapRoots       []map[string]Value
	objectArrays        map[distributedByteArrayKey]int
	objectRoots         [][]any
	objectMaps          map[uintptr]int
	objectMapRoots      []map[string]any
	objectKeyMaps       map[uintptr]int
	objectKeyMapRoots   []map[any]any
	primitiveArrays     map[distributedPrimitiveArrayKey]int
	primitiveArrayRoots []any
	stringArrays        map[distributedByteArrayKey]int
	stringArrayRoots    [][]string

	stateVectorArrays     map[distributedByteArrayKey]int
	stateVectorArrayRoots [][]*StateVec
	longVectorArrays      map[distributedByteArrayKey]int
	longVectorArrayRoots  [][]*LongVec
}

type distributedByteArrayKey struct {
	address uintptr
	length  int
}

func EncodeDistributedStates(states []*TLCStateMut) (payload *DistributedStatePayload, err error) {
	return encodeDistributedStates(states, &distributedPayloadEncoder{})
}

func encodeDistributedStates(states []*TLCStateMut, encoder *distributedPayloadEncoder) (payload *DistributedStatePayload, err error) {
	defer func() {
		if failure := recover(); failure != nil {
			payload = nil
			err = panicValueAsError(failure)
		}
	}()
	payload = &DistributedStatePayload{Nil: states == nil}
	*encoder = distributedPayloadEncoder{payload: payload, states: make(map[*TLCStateMut]int), values: make(map[Value]int), strings: make(map[*UniqueString]int), bytes: make(map[distributedByteArrayKey]int), arrays: make(map[distributedByteArrayKey]int), vectors: make(map[*ValueVec]int), longVectors: make(map[*LongVec]int), nameArrays: make(map[distributedByteArrayKey]int)}
	payload.Roots = make([]int, len(states))
	for i, state := range states {
		id, failure := encoder.state(state)
		if failure != nil {
			return nil, failure
		}
		payload.Roots[i] = id
	}
	// Only an attached alias needs an explicit root-container reference.
	// Keep the ordinary inline root format for all other invocations.
	if len(states) != 0 {
		key := distributedByteArrayKey{reflect.ValueOf(states).Pointer(), len(states)}
		if id := encoder.stateArrays[key]; id != 0 {
			payload.RootsArray, payload.Roots = id, nil
		}
	}
	return payload, nil
}

func (e *distributedPayloadEncoder) stateArray(states []*TLCStateMut) (int, error) {
	if states == nil {
		return 0, nil
	}
	key := distributedByteArrayKey{reflect.ValueOf(states).Pointer(), len(states)}
	if id := e.stateArrays[key]; id != 0 && len(states) != 0 {
		return id, nil
	}
	if e.stateArrays == nil {
		e.stateArrays = make(map[distributedByteArrayKey]int)
	}
	id := len(e.payload.StateArrays) + 1
	e.stateArrays[key] = id
	e.stateArrayRoots = append(e.stateArrayRoots, states)
	e.payload.StateArrays = append(e.payload.StateArrays, nil)
	refs := make([]int, len(states))
	for i, state := range states {
		ref, err := e.state(state)
		if err != nil {
			return 0, err
		}
		refs[i] = ref
	}
	e.payload.StateArrays[id-1] = refs
	return id, nil
}

func (e *distributedPayloadEncoder) state(state *TLCStateMut) (int, error) {
	if state == nil {
		return 0, nil
	}
	if id := e.states[state]; id != 0 {
		return id, nil
	}
	if state.level < 0 || state.level > math.MaxInt32 {
		return 0, fmt.Errorf("invalid network state level %d", state.level)
	}
	// Evaluator objects require their own representation and must not be dropped.
	// Predecessor links use the same native state graph as invocation roots.
	if state.functional || state.functionalBindings != nil || state.action != nil || state.callable != nil {
		return 0, fmt.Errorf("network state contains extended evaluator metadata")
	}
	id := len(e.payload.States) + 1
	e.states[state] = id
	e.payload.States = append(e.payload.States, DistributedStateNode{})
	array, err := e.array(state.values)
	if err != nil {
		return 0, err
	}
	var predecessor int
	if parent := state.TracePredecessor(); parent != nil {
		mutable, ok := parent.(*TLCStateMut)
		if !ok {
			return 0, fmt.Errorf("unsupported distributed predecessor state %T", parent)
		}
		predecessor, err = e.state(mutable)
		if err != nil {
			return 0, err
		}
	}
	cache, err := e.stateCache(state.cached)
	if err != nil {
		return 0, err
	}
	printRecord, err := e.value(state.printRecord)
	if err != nil {
		return 0, err
	}
	e.payload.States[id-1] = DistributedStateNode{WorkerID: state.WorkerID, UID: state.UID, Level: int32(state.level), ValuesArray: array, ValuesNil: state.values == nil, Predecessor: predecessor, Cache: cache, PrintRecord: printRecord}
	return id, nil
}

func (e *distributedPayloadEncoder) stateCache(cache map[int]Value) (int, error) {
	if cache == nil {
		return 0, nil
	}
	address := uintptr(reflect.ValueOf(cache).UnsafePointer())
	if id := e.caches[address]; id != 0 {
		return id, nil
	}
	if e.caches == nil {
		e.caches = make(map[uintptr]int)
	}
	id := len(e.payload.StateCaches) + 1
	e.caches[address] = id
	e.cacheRoots = append(e.cacheRoots, cache)
	e.payload.StateCaches = append(e.payload.StateCaches, nil)
	keys := make([]int, 0, len(cache))
	for key := range cache {
		if int64(key) < math.MinInt32 || int64(key) > math.MaxInt32 {
			return 0, fmt.Errorf("state cache key %d outside signed int32 range", key)
		}
		keys = append(keys, key)
	}
	sort.Ints(keys)
	entries := make([]DistributedStateCacheEntry, len(keys))
	for i, key := range keys {
		value, err := e.value(cache[key])
		if err != nil {
			return 0, err
		}
		entries[i] = DistributedStateCacheEntry{Key: int32(key), Value: value}
	}
	e.payload.StateCaches[id-1] = entries
	return id, nil
}

func (e *distributedPayloadEncoder) string(value *UniqueString) int {
	if value == nil {
		return 0
	}
	if id := e.strings[value]; id != 0 {
		return id
	}
	id := len(e.payload.Strings) + 1
	e.strings[value] = id
	e.payload.Strings = append(e.payload.Strings, DistributedStringNode{Text: value.s, Token: value.tok, Location: value.loc, Unregistered: value.unregistered})
	return id
}

func (e *distributedPayloadEncoder) refs(values []Value) ([]int, error) {
	if values == nil {
		return nil, nil
	}
	refs := make([]int, len(values))
	for i, value := range values {
		id, err := e.value(value)
		if err != nil {
			return nil, err
		}
		refs[i] = id
	}
	return refs, nil
}

func (e *distributedPayloadEncoder) array(values []Value) (int, error) {
	if values == nil {
		return 0, nil
	}
	key := distributedByteArrayKey{reflect.ValueOf(values).Pointer(), len(values)}
	if id := e.arrays[key]; id != 0 && len(values) != 0 {
		return id, nil
	}
	id := len(e.payload.ValueArrays) + 1
	e.arrays[key] = id
	e.arrayRoots = append(e.arrayRoots, values)
	// Reserve before visiting elements so an array may refer to its owner.
	e.payload.ValueArrays = append(e.payload.ValueArrays, nil)
	refs, err := e.refs(values)
	if err != nil {
		return 0, err
	}
	e.payload.ValueArrays[id-1] = refs
	return id, nil
}

func (e *distributedPayloadEncoder) names(names []*UniqueString) []int {
	if names == nil {
		return nil
	}
	refs := make([]int, len(names))
	for i, name := range names {
		refs[i] = e.string(name)
	}
	return refs
}

func (e *distributedPayloadEncoder) nameArray(names []*UniqueString) int {
	if names == nil {
		return 0
	}
	key := distributedByteArrayKey{reflect.ValueOf(names).Pointer(), len(names)}
	if id := e.nameArrays[key]; id != 0 && len(names) != 0 {
		return id
	}
	id := len(e.payload.NameArrays) + 1
	e.nameArrays[key] = id
	e.nameRoots = append(e.nameRoots, names)
	e.payload.NameArrays = append(e.payload.NameArrays, e.names(names))
	return id
}

func (e *distributedPayloadEncoder) vector(vector *ValueVec) (int, error) {
	if vector == nil {
		return 0, nil
	}
	if id := e.vectors[vector]; id != 0 {
		return id, nil
	}
	id := len(e.payload.ValueVectors) + 1
	e.vectors[vector] = id
	e.payload.ValueVectors = append(e.payload.ValueVectors, DistributedValueVectorNode{})
	array, err := e.array(vector.data[:cap(vector.data)])
	if err != nil {
		return 0, err
	}
	e.payload.ValueVectors[id-1] = DistributedValueVectorNode{Array: array, Count: len(vector.data)}
	return id, nil
}

func (e *distributedPayloadEncoder) value(value Value) (int, error) {
	if value == nil || reflect.ValueOf(value).Kind() == reflect.Pointer && reflect.ValueOf(value).IsNil() {
		return 0, nil
	}
	if id := e.values[value]; id != 0 {
		return id, nil
	}
	id := len(e.payload.Values) + 1
	e.values[value] = id
	e.payload.Values = append(e.payload.Values, DistributedValueNode{})
	node := DistributedValueNode{}
	var children, domain []Value
	var names []*UniqueString
	var cache Value
	sharedChildren, sharedDomain := false, false
	switch v := value.(type) {
	case *BoolValue:
		node.Kind, node.Flag = "bool", v.Val
	case *IntValue:
		node.Kind, node.Integer = "int", int64(v.Val)
	case *StringValue:
		node.Kind, node.String = "string", e.string(v.Val)
	case *ModelValue:
		node.Kind, node.String, node.ModelIndex, node.ModelType = "model", e.string(v.Val), v.Index, v.Type
		if err := e.modelData(&node, v.Data); err != nil {
			return 0, err
		}
	case *IntervalValue:
		node.Kind, node.Low, node.High = "interval", v.Low, v.High
	case *TupleValue:
		node.Kind, children = "tuple", v.Elems
		sharedChildren = true
	case *RecordValue:
		node.Kind, children, names, node.Flag = "record", v.Values, v.Names, v.IsNorm
		sharedChildren = true
	case *SetEnumValue:
		node.Kind, node.Flag, node.CollectionPresent = "enum", v.IsNorm, v.Elems != nil
		vector, err := e.vector(v.Elems)
		if err != nil {
			return 0, err
		}
		node.Vector = vector
	case *FcnRcdValue:
		node.Kind, children, domain, node.Flag = "function", v.Values, v.Domain, v.IsNorm
		cache = v.Intv
		sharedChildren, sharedDomain = true, true
	case *OpRcdValue:
		// Unlike evaluator-backed operators, a configured constant operator
		// contains only finite argument rows and result values. Keep it an
		// operator; converting it to a function would change application rules.
		node.Kind, children, node.OperatorDomainNil = "operatorRecord", v.Values, v.Domain == nil
		rows, err := e.valueRows(v.Domain)
		if err != nil {
			return 0, err
		}
		node.OperatorDomainArray = rows
		sharedChildren = true
	case *FcnLambdaValue:
		node.Kind, children = "lambda", []Value{v.ToFcnRcd()}
	case *LazySupplierValue:
		if v.Val == nil || v.Val == ValUndef {
			return 0, distributedLazySerializationFailure(v.GetSource())
		}
		if v.Supplier != nil {
			return 0, fmt.Errorf("unsupported network lazy supplier function")
		}
		node.Kind, children = "lazySupplier", []Value{v.Val}
	case *LazyValue:
		if v.Val == nil || v.Val == ValUndef {
			return 0, distributedLazySerializationFailure(v.GetSource())
		}
		node.Kind, children = "lazy", []Value{v.Val}
	case *SetPredValue:
		if !v.Converted {
			set, err := v.ToSetEnum()
			if err != nil {
				return 0, err
			}
			v.InVal, v.Converted = set, true
		}
		node.Kind, children = "predicate", []Value{v.InVal}
	case *SetOfTuplesValue:
		node.Kind, children, cache, node.Dummy = "product", v.Sets, v.TupleSet, v.TupleSetDummy
		sharedChildren = true
	case *SetOfRcdsValue:
		node.Kind, children, names, cache, node.Dummy = "recordSet", v.Values, v.Names, v.RcdSet, v.RcdSetDummy
		sharedChildren = true
	case *SetOfFcnsValue:
		node.Kind, children, cache, node.Dummy = "functionSet", []Value{v.Domain, v.Range}, v.FcnSet, v.FcnSetDummy
	case *KSubsetValue:
		node.Kind, children, cache, node.Dummy, node.Integer = "kSubset", []Value{v.Set}, v.PSet, v.PSetDummy, int64(v.K)
	case *SubsetValue:
		node.Kind, children, cache, node.Dummy = "subset", []Value{v.Set}, v.PSet, v.PSetDummy
	case *SetCupValue:
		node.Kind, children, cache, node.Dummy = "cup", []Value{v.Set1, v.Set2}, v.CupSet, v.CupSetDummy
	case *SetCapValue:
		node.Kind, children, cache, node.Dummy = "cap", []Value{v.Set1, v.Set2}, v.CapSet, v.CapSetDummy
	case *SetDiffValue:
		node.Kind, children, cache, node.Dummy = "difference", []Value{v.Set1, v.Set2}, v.DiffSet, v.DiffSetDummy
	case *UnionValue:
		node.Kind, children, cache, node.Dummy = "union", []Value{v.Set}, v.RealSet, v.RealSetDummy
	case *UndefValue:
		node.Kind = "undefined"
	case *CounterExample:
		node.Kind, children = "counterexample", []Value{v.RecordValue}
	case *UserValue:
		switch obj := v.UserObj.(type) {
		case naturalsObj:
			node.Kind = "naturals"
		case integersObj:
			node.Kind = "integers"
		case stringsObj:
			node.Kind = "strings"
		case AnySet:
			node.Kind = "any"
		case *sequencesObj:
			node.Kind, children, node.Integer = "sequences", []Value{obj.Range}, int64(obj.SizeBound)
		default:
			return 0, fmt.Errorf("unsupported network user value %T", v.UserObj)
		}
	default:
		return 0, fmt.Errorf("unsupported network value %T", value)
	}
	var err error
	node.ReferencesNil, node.DomainNil, node.NamesNil = children == nil, domain == nil, names == nil
	if sharedChildren {
		node.ReferencesArray, err = e.array(children)
	} else {
		node.References, err = e.refs(children)
	}
	if err != nil {
		return 0, err
	}
	if sharedDomain {
		node.DomainArray, err = e.array(domain)
	} else {
		node.Domain, err = e.refs(domain)
	}
	if err != nil {
		return 0, err
	}
	if node.Kind == "record" || node.Kind == "recordSet" {
		node.NamesArray = e.nameArray(names)
	} else {
		node.Names = e.names(names)
	}
	node.Cache, err = e.value(cache)
	if err != nil {
		return 0, err
	}
	e.payload.Values[id-1] = node
	return id, nil
}

func distributedLazySerializationFailure(source SemanticNode) error {
	const message = "Error(TLC): Attempted to serialize lazy value."
	if source != nil {
		return NewTLCDetailedRuntimeException(ECGeneral, message, source, EmptyContext)
	}
	failure := newTLCError(ECGeneral, "%s", message)
	failure.Runtime = true
	return failure
}

func (e *distributedPayloadEncoder) modelData(node *DistributedValueNode, data any) error {
	switch v := data.(type) {
	case nil:
	case string:
		node.DataKind, node.DataString = "string", v
	case *UniqueString:
		node.DataKind, node.DataName = "uniqueString", e.string(v)
	case []*UniqueString:
		node.DataKind, node.DataArray = "nameArray", e.nameArray(v)
	case *LongVec:
		node.DataKind, node.DataArray = "longVector", e.longVector(v)
	case *BitVector:
		node.DataKind, node.DataArray = "bitVector", e.bitVector(v)
	case *StateVec:
		id, err := e.stateVector(v)
		if err != nil {
			return err
		}
		node.DataKind, node.DataArray = "stateVector", id
	case []*StateVec:
		id, err := e.stateVectorArray(v)
		if err != nil {
			return err
		}
		node.DataKind, node.DataArray = "stateVectorArray", id
	case []*LongVec:
		node.DataKind, node.DataArray = "longVectorArray", e.longVectorArray(v)
	case *TLCStateMut:
		id, err := e.state(v)
		if err != nil {
			return err
		}
		node.DataKind, node.DataState = "state", id
	case []*TLCStateMut:
		id, err := e.stateArray(v)
		if err != nil {
			return err
		}
		node.DataKind, node.DataArray = "stateArray", id
	case bool:
		node.DataKind, node.DataBool = "bool", v
	case int:
		node.DataKind, node.DataInteger = "int", int64(v)
	case int8:
		node.DataKind, node.DataInteger = "int8", int64(v)
	case int16:
		node.DataKind, node.DataInteger = "int16", int64(v)
	case uint16:
		node.DataKind, node.DataInteger = "uint16", int64(v)
	case int32:
		node.DataKind, node.DataInteger = "int32", int64(v)
	case int64:
		node.DataKind, node.DataInteger = "int64", v
	case float64:
		node.DataKind, node.DataFloatBits = "float64", math.Float64bits(v)
	case float32:
		node.DataKind, node.DataFloatBits = "float32", uint64(math.Float32bits(v))
	case []bool:
		encodeDistributedPrimitiveArray(e, node, v, "boolArray", func(v bool) uint64 {
			if v {
				return 1
			}
			return 0
		})
	case []int:
		encodeDistributedPrimitiveArray(e, node, v, "intArray", func(v int) uint64 { return uint64(v) })
	case []int8:
		encodeDistributedPrimitiveArray(e, node, v, "int8Array", func(v int8) uint64 { return uint64(v) })
	case []int16:
		encodeDistributedPrimitiveArray(e, node, v, "int16Array", func(v int16) uint64 { return uint64(v) })
	case []int32:
		encodeDistributedPrimitiveArray(e, node, v, "int32Array", func(v int32) uint64 { return uint64(v) })
	case []int64:
		encodeDistributedPrimitiveArray(e, node, v, "int64Array", func(v int64) uint64 { return uint64(v) })
	case []uint64:
		encodeDistributedPrimitiveArray(e, node, v, "uint64Array", func(v uint64) uint64 { return v })
	case []uint16:
		encodeDistributedPrimitiveArray(e, node, v, "uint16Array", func(v uint16) uint64 { return uint64(v) })
	case []float32:
		encodeDistributedPrimitiveArray(e, node, v, "float32Array", func(v float32) uint64 { return uint64(math.Float32bits(v)) })
	case []float64:
		encodeDistributedPrimitiveArray(e, node, v, "float64Array", math.Float64bits)
	case []string:
		node.DataKind = "stringArray"
		if v != nil {
			if e.stringArrays == nil {
				e.stringArrays = make(map[distributedByteArrayKey]int)
			}
			key := distributedByteArrayKey{reflect.ValueOf(v).Pointer(), len(v)}
			id := e.stringArrays[key]
			if id == 0 || len(v) == 0 {
				array := make([]string, len(v))
				copy(array, v)
				id = len(e.payload.StringArrays) + 1
				e.stringArrays[key] = id
				e.stringArrayRoots = append(e.stringArrayRoots, v)
				e.payload.StringArrays = append(e.payload.StringArrays, array)
			}
			node.DataArray = id
		}
	case []byte:
		node.DataKind = "bytes"
		if v != nil {
			key := distributedByteArrayKey{reflect.ValueOf(v).Pointer(), len(v)}
			id := e.bytes[key]
			if id == 0 || len(v) == 0 {
				id = len(e.payload.ByteArrays) + 1
				e.bytes[key] = id
				e.payload.ByteArrays = append(e.payload.ByteArrays, append([]byte{}, v...))
			}
			node.DataBytes = id
		}
	case []Value:
		id, err := e.array(v)
		if err != nil {
			return err
		}
		node.DataKind, node.DataArray = "valueArray", id
	case [][]Value:
		id, err := e.valueRows(v)
		if err != nil {
			return err
		}
		node.DataKind, node.DataArray = "valueRows", id
	case map[string]Value:
		id, err := e.valueMap(v)
		if err != nil {
			return err
		}
		node.DataKind, node.DataMap = "valueMap", id
	case []any:
		id, err := e.objectArray(v)
		if err != nil {
			return err
		}
		node.DataKind, node.DataArray = "objectArray", id
	case map[string]any:
		id, err := e.objectMap(v)
		if err != nil {
			return err
		}
		node.DataKind, node.DataMap = "objectMap", id
	case map[any]any:
		id, err := e.objectKeyMap(v)
		if err != nil {
			return err
		}
		node.DataKind, node.DataMap = "objectKeyMap", id
	case Value:
		id, err := e.value(v)
		if err != nil {
			return err
		}
		node.DataKind, node.DataValue = "value", id
	default:
		return fmt.Errorf("unsupported network model-value data %T", data)
	}
	return nil
}

func (e *distributedPayloadEncoder) longVector(vector *LongVec) int {
	if vector == nil {
		return 0
	}
	if id := e.longVectors[vector]; id != 0 {
		return id
	}
	id := len(e.payload.LongVectors) + 1
	e.longVectors[vector] = id
	e.payload.LongVectors = append(e.payload.LongVectors, append([]int64(nil), vector.data...))
	return id
}

func (e *distributedPayloadEncoder) valueRows(rows [][]Value) (int, error) {
	if rows == nil {
		return 0, nil
	}
	if e.rows == nil {
		e.rows = make(map[distributedByteArrayKey]int)
	}
	key := distributedByteArrayKey{reflect.ValueOf(rows).Pointer(), len(rows)}
	if id := e.rows[key]; id != 0 && len(rows) != 0 {
		return id, nil
	}
	id := len(e.payload.ValueRows) + 1
	e.rows[key] = id
	e.rowRoots = append(e.rowRoots, rows)
	e.payload.ValueRows = append(e.payload.ValueRows, nil)
	entries := make([]int, len(rows))
	for i, row := range rows {
		array, err := e.array(row)
		if err != nil {
			return 0, err
		}
		entries[i] = array
	}
	e.payload.ValueRows[id-1] = entries
	return id, nil
}

func (e *distributedPayloadEncoder) objectArray(values []any) (int, error) {
	if values == nil {
		return 0, nil
	}
	if e.objectArrays == nil {
		e.objectArrays = make(map[distributedByteArrayKey]int)
	}
	key := distributedByteArrayKey{reflect.ValueOf(values).Pointer(), len(values)}
	if id := e.objectArrays[key]; id != 0 && len(values) != 0 {
		return id, nil
	}
	id := len(e.payload.ObjectArrays) + 1
	e.objectArrays[key] = id
	e.objectRoots = append(e.objectRoots, values)
	e.payload.ObjectArrays = append(e.payload.ObjectArrays, nil)
	entries := make([]DistributedObjectDataNode, len(values))
	for i, data := range values {
		var node DistributedValueNode
		if err := e.modelData(&node, data); err != nil {
			return 0, err
		}
		entries[i] = distributedObjectDataNode(node)
	}
	e.payload.ObjectArrays[id-1] = entries
	return id, nil
}

func (e *distributedPayloadEncoder) objectMap(values map[string]any) (int, error) {
	if values == nil {
		return 0, nil
	}
	if e.objectMaps == nil {
		e.objectMaps = make(map[uintptr]int)
	}
	key := uintptr(reflect.ValueOf(values).UnsafePointer())
	if id := e.objectMaps[key]; id != 0 {
		return id, nil
	}
	id := len(e.payload.ObjectMaps) + 1
	e.objectMaps[key] = id
	e.objectMapRoots = append(e.objectMapRoots, values)
	e.payload.ObjectMaps = append(e.payload.ObjectMaps, nil)
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	entries := make([]DistributedObjectMapEntry, len(keys))
	for i, key := range keys {
		entries[i].Key = key
		var node DistributedValueNode
		if err := e.modelData(&node, values[key]); err != nil {
			return 0, err
		}
		entries[i].Data = distributedObjectDataNode(node)
	}
	e.payload.ObjectMaps[id-1] = entries
	return id, nil
}

func (e *distributedPayloadEncoder) objectKeyMap(values map[any]any) (int, error) {
	if values == nil {
		return 0, nil
	}
	if e.objectKeyMaps == nil {
		e.objectKeyMaps = make(map[uintptr]int)
	}
	key := uintptr(reflect.ValueOf(values).UnsafePointer())
	if id := e.objectKeyMaps[key]; id != 0 {
		return id, nil
	}
	id := len(e.payload.ObjectKeyMaps) + 1
	e.objectKeyMaps[key] = id
	e.objectKeyMapRoots = append(e.objectKeyMapRoots, values)
	e.payload.ObjectKeyMaps = append(e.payload.ObjectKeyMaps, nil)
	entries := make([]DistributedObjectKeyMapEntry, 0, len(values))
	// Preserve native key types and identity. Unlike string keys, general keys
	// have no common ordering; do not fingerprint or evaluate value keys to sort.
	for key, data := range values {
		var keyNode, dataNode DistributedValueNode
		if err := e.modelData(&keyNode, key); err != nil {
			return 0, fmt.Errorf("object-key map key: %w", err)
		}
		if err := e.modelData(&dataNode, data); err != nil {
			return 0, fmt.Errorf("object-key map entry: %w", err)
		}
		entries = append(entries, DistributedObjectKeyMapEntry{Key: distributedObjectDataNode(keyNode), Data: distributedObjectDataNode(dataNode)})
	}
	e.payload.ObjectKeyMaps[id-1] = entries
	return id, nil
}

func (e *distributedPayloadEncoder) valueMap(values map[string]Value) (int, error) {
	if values == nil {
		return 0, nil
	}
	if e.valueMaps == nil {
		e.valueMaps = make(map[uintptr]int)
	}
	key := uintptr(reflect.ValueOf(values).UnsafePointer())
	if id := e.valueMaps[key]; id != 0 {
		return id, nil
	}
	id := len(e.payload.ValueMaps) + 1
	e.valueMaps[key] = id
	e.valueMapRoots = append(e.valueMapRoots, values)
	// Reserve identity before traversing values that may point back to this map.
	e.payload.ValueMaps = append(e.payload.ValueMaps, nil)
	keys := make([]string, 0, len(values))
	for name := range values {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	entries := make([]DistributedValueMapEntry, 0, len(keys))
	for _, name := range keys {
		value, err := e.value(values[name])
		if err != nil {
			return 0, err
		}
		entries = append(entries, DistributedValueMapEntry{Key: name, Value: value})
	}
	e.payload.ValueMaps[id-1] = entries
	return id, nil
}

type distributedPayloadDecoder struct {
	states              []*TLCStateMut
	stateArrays         [][]*TLCStateMut
	values              []Value
	strings             []*UniqueString
	bytes               [][]byte
	arrays              [][]Value
	rows                [][][]Value
	vectors             []*ValueVec
	longVectors         []*LongVec
	bitVectors          []*BitVector
	stateVectors        []*StateVec
	stateVectorArrays   [][]*StateVec
	longVectorArrays    [][]*LongVec
	nameArrays          [][]*UniqueString
	valueMaps           []map[string]Value
	objectArrays        [][]any
	objectMaps          []map[string]any
	objectKeyMaps       []map[any]any
	primitiveArrays     []any
	primitiveArrayKinds []string
	stringArrays        [][]string
}

func DecodeDistributedStates(payload *DistributedStatePayload) (states []*TLCStateMut, err error) {
	return decodeDistributedStates(payload, &distributedPayloadDecoder{})
}

func decodeDistributedStates(payload *DistributedStatePayload, decoder *distributedPayloadDecoder) (states []*TLCStateMut, err error) {
	defer func() {
		if failure := recover(); failure != nil {
			states = nil
			err = panicValueAsError(failure)
		}
	}()
	if payload == nil {
		return nil, fmt.Errorf("missing distributed state payload")
	}
	if payload.Nil && (len(payload.Roots) != 0 || payload.RootsArray != 0) {
		return nil, fmt.Errorf("null distributed state array contains roots")
	}
	if payload.RootsArray < 0 || payload.RootsArray > len(payload.StateArrays) {
		return nil, fmt.Errorf("invalid distributed root array reference %d", payload.RootsArray)
	}
	if payload.RootsArray != 0 && len(payload.Roots) != 0 {
		return nil, fmt.Errorf("root array reference conflicts with inline roots")
	}
	*decoder = distributedPayloadDecoder{values: make([]Value, len(payload.Values)), strings: make([]*UniqueString, len(payload.Strings)), bytes: make([][]byte, len(payload.ByteArrays))}
	decoder.longVectors = make([]*LongVec, len(payload.LongVectors))
	for i, data := range payload.LongVectors {
		decoder.longVectors[i] = NewLongVecFrom(data)
	}
	for i, data := range payload.ByteArrays {
		decoder.bytes[i] = make([]byte, len(data))
		copy(decoder.bytes[i], data)
	}
	decoder.primitiveArrays = make([]any, len(payload.PrimitiveArrays))
	decoder.stringArrays = make([][]string, len(payload.StringArrays))
	for i, array := range payload.StringArrays {
		decoder.stringArrays[i] = make([]string, len(array))
		copy(decoder.stringArrays[i], array)
	}
	decoder.primitiveArrayKinds = make([]string, len(payload.PrimitiveArrays))
	for i, node := range payload.PrimitiveArrays {
		array, err := decodeDistributedPrimitiveArray(node, false)
		if err != nil {
			return nil, fmt.Errorf("primitive array %d: %w", i+1, err)
		}
		decoder.primitiveArrays[i], decoder.primitiveArrayKinds[i] = array, node.Kind
	}
	decoder.bitVectors = make([]*BitVector, len(payload.BitVectors))
	for i, id := range payload.BitVectors {
		words, err := decoder.modelData(DistributedValueNode{DataKind: "uint64Array", DataArray: id})
		if err != nil {
			return nil, fmt.Errorf("bit vector %d words: %w", i+1, err)
		}
		decoder.bitVectors[i] = &BitVector{word: words.([]uint64)}
	}
	for i, name := range payload.Strings {
		decoder.strings[i] = &UniqueString{s: name.Text, tok: name.Token, loc: name.Location, unregistered: name.Unregistered}
	}
	decoder.nameArrays = make([][]*UniqueString, len(payload.NameArrays))
	for i, refs := range payload.NameArrays {
		names, err := decoder.names(refs, false)
		if err != nil {
			return nil, fmt.Errorf("name array %d: %w", i+1, err)
		}
		decoder.nameArrays[i] = names
	}
	for i, node := range payload.Values {
		value, err := allocateDistributedValue(node)
		if err != nil {
			return nil, fmt.Errorf("value %d: %w", i+1, err)
		}
		decoder.values[i] = value
	}
	// Allocate state identities before resolving attachments. A value or mixed
	// container can point back to a root or to another attached-only state.
	decoder.states = make([]*TLCStateMut, len(payload.States))
	for i := range decoder.states {
		decoder.states[i] = &TLCStateMut{}
	}
	decoder.stateVectors = make([]*StateVec, len(payload.StateVectors))
	for i, refs := range payload.StateVectors {
		vector := newDistributedStateVec(len(refs))
		for _, id := range refs {
			if id < 0 || id > len(decoder.states) {
				return nil, fmt.Errorf("state vector %d: invalid state reference %d", i+1, id)
			}
			var state *TLCStateMut
			if id != 0 {
				state = decoder.states[id-1]
			}
			vector.Add(state)
		}
		decoder.stateVectors[i] = vector
	}
	decoder.stateVectorArrays = make([][]*StateVec, len(payload.StateVectorArrays))
	for i, refs := range payload.StateVectorArrays {
		array := make([]*StateVec, len(refs))
		for j, id := range refs {
			if id < 0 || id > len(decoder.stateVectors) {
				return nil, fmt.Errorf("state vector array %d: invalid vector reference %d", i+1, id)
			}
			if id != 0 {
				array[j] = decoder.stateVectors[id-1]
			}
		}
		decoder.stateVectorArrays[i] = array
	}
	decoder.longVectorArrays = make([][]*LongVec, len(payload.LongVectorArrays))
	for i, refs := range payload.LongVectorArrays {
		array := make([]*LongVec, len(refs))
		for j, id := range refs {
			if id < 0 || id > len(decoder.longVectors) {
				return nil, fmt.Errorf("long vector array %d: invalid vector reference %d", i+1, id)
			}
			if id != 0 {
				array[j] = decoder.longVectors[id-1]
			}
		}
		decoder.longVectorArrays[i] = array
	}
	decoder.stateArrays = make([][]*TLCStateMut, len(payload.StateArrays))
	for i, refs := range payload.StateArrays {
		array := make([]*TLCStateMut, len(refs))
		for j, id := range refs {
			if id < 0 || id > len(decoder.states) {
				return nil, fmt.Errorf("state array %d: invalid state reference %d", i+1, id)
			}
			if id != 0 {
				array[j] = decoder.states[id-1]
			}
		}
		decoder.stateArrays[i] = array
	}
	decoder.valueMaps = make([]map[string]Value, len(payload.ValueMaps))
	for i, entries := range payload.ValueMaps {
		values := make(map[string]Value, len(entries))
		for _, entry := range entries {
			if _, duplicate := values[entry.Key]; duplicate {
				return nil, fmt.Errorf("value map %d has duplicate key %q", i+1, entry.Key)
			}
			value, err := decoder.value(entry.Value)
			if err != nil {
				return nil, fmt.Errorf("value map %d: %w", i+1, err)
			}
			values[entry.Key] = value
		}
		decoder.valueMaps[i] = values
	}
	decoder.arrays = make([][]Value, len(payload.ValueArrays))
	for i, refs := range payload.ValueArrays {
		values, err := decoder.refs(refs, false)
		if err != nil {
			return nil, fmt.Errorf("value array %d: %w", i+1, err)
		}
		decoder.arrays[i] = values
	}
	decoder.rows = make([][][]Value, len(payload.ValueRows))
	for i, entries := range payload.ValueRows {
		rows := make([][]Value, len(entries))
		for j, id := range entries {
			row, err := decoder.arrayRefs(nil, id == 0, id)
			if err != nil {
				return nil, fmt.Errorf("value rows %d row %d: %w", i+1, j, err)
			}
			rows[j] = row
		}
		decoder.rows[i] = rows
	}
	decoder.vectors = make([]*ValueVec, len(payload.ValueVectors))
	for i, node := range payload.ValueVectors {
		array, err := decoder.arrayRefs(nil, false, node.Array)
		if err != nil {
			return nil, fmt.Errorf("value vector %d: %w", i+1, err)
		}
		if node.Count < 0 || node.Count > len(array) {
			return nil, fmt.Errorf("value vector %d: count outside backing array", i+1)
		}
		decoder.vectors[i] = &ValueVec{data: array[:node.Count]}
	}
	// Allocate all mixed containers before resolving entries. Values and
	// attachments can point to one another or share recursive containers.
	decoder.objectArrays = make([][]any, len(payload.ObjectArrays))
	for i, entries := range payload.ObjectArrays {
		decoder.objectArrays[i] = make([]any, len(entries))
	}
	decoder.objectMaps = make([]map[string]any, len(payload.ObjectMaps))
	for i, entries := range payload.ObjectMaps {
		decoder.objectMaps[i] = make(map[string]any, len(entries))
	}
	decoder.objectKeyMaps = make([]map[any]any, len(payload.ObjectKeyMaps))
	for i, entries := range payload.ObjectKeyMaps {
		decoder.objectKeyMaps[i] = make(map[any]any, len(entries))
	}
	for i, entries := range payload.ObjectArrays {
		for j, node := range entries {
			data, err := decoder.modelData(node.valueNode())
			if err != nil {
				return nil, fmt.Errorf("object array %d entry %d: %w", i+1, j, err)
			}
			decoder.objectArrays[i][j] = data
		}
	}
	for i, entries := range payload.ObjectMaps {
		for _, entry := range entries {
			if _, duplicate := decoder.objectMaps[i][entry.Key]; duplicate {
				return nil, fmt.Errorf("object map %d has duplicate key %q", i+1, entry.Key)
			}
			data, err := decoder.modelData(entry.Data.valueNode())
			if err != nil {
				return nil, fmt.Errorf("object map %d key %q: %w", i+1, entry.Key, err)
			}
			decoder.objectMaps[i][entry.Key] = data
		}
	}
	for i, entries := range payload.ObjectKeyMaps {
		for j, entry := range entries {
			key, err := decoder.modelData(entry.Key.valueNode())
			if err != nil {
				return nil, fmt.Errorf("object-key map %d key %d: %w", i+1, j, err)
			}
			if key != nil && !reflect.TypeOf(key).Comparable() {
				return nil, fmt.Errorf("object-key map %d has non-comparable key %T", i+1, key)
			}
			if _, duplicate := decoder.objectKeyMaps[i][key]; duplicate {
				return nil, fmt.Errorf("object-key map %d has duplicate key", i+1)
			}
			data, err := decoder.modelData(entry.Data.valueNode())
			if err != nil {
				return nil, fmt.Errorf("object-key map %d entry %d: %w", i+1, j, err)
			}
			decoder.objectKeyMaps[i][key] = data
		}
	}
	for i, node := range payload.Values {
		if err := decoder.populate(decoder.values[i], node); err != nil {
			return nil, fmt.Errorf("value %d: %w", i+1, err)
		}
	}
	caches := make([]map[int]Value, len(payload.StateCaches))
	for i, entries := range payload.StateCaches {
		cache := make(map[int]Value, len(entries))
		for _, entry := range entries {
			key := int(entry.Key)
			if _, duplicate := cache[key]; duplicate {
				return nil, fmt.Errorf("state cache %d has duplicate key %d", i+1, key)
			}
			value, err := decoder.value(entry.Value)
			if err != nil {
				return nil, fmt.Errorf("state cache %d: %w", i+1, err)
			}
			cache[key] = value
		}
		caches[i] = cache
	}
	objects := decoder.states
	for i, node := range payload.States {
		if node.Level < 0 {
			return nil, fmt.Errorf("negative distributed state level")
		}
		values, err := decoder.arrayRefs(node.Values, node.ValuesNil, node.ValuesArray)
		if err != nil {
			return nil, err
		}
		objects[i].WorkerID, objects[i].UID, objects[i].level, objects[i].values = node.WorkerID, node.UID, int(node.Level), values
		record, err := decoder.value(node.PrintRecord)
		if err != nil {
			return nil, fmt.Errorf("state print record: %w", err)
		}
		objects[i].printRecord, err = distributedValueCast[*RecordValue](record)
		if err != nil {
			return nil, fmt.Errorf("state print record: %w", err)
		}
		if node.Cache < 0 || node.Cache > len(caches) {
			return nil, fmt.Errorf("invalid distributed state cache reference %d", node.Cache)
		}
		if node.Cache != 0 {
			objects[i].cached = caches[node.Cache-1]
		}
	}
	for i, node := range payload.States {
		if node.Predecessor < 0 || node.Predecessor > len(objects) {
			return nil, fmt.Errorf("invalid distributed predecessor reference %d", node.Predecessor)
		}
		if node.Predecessor != 0 {
			// Restore stored graph fields directly. SetTracePredecessor changes
			// level and consults process-local metadata settings.
			objects[i].pred = objects[node.Predecessor-1]
		}
	}
	if payload.Nil {
		return nil, nil
	}
	if payload.RootsArray != 0 {
		return decoder.stateArrays[payload.RootsArray-1], nil
	}
	states = make([]*TLCStateMut, len(payload.Roots))
	for i, id := range payload.Roots {
		if id < 0 || id > len(objects) {
			return nil, fmt.Errorf("invalid distributed state reference %d", id)
		}
		if id != 0 {
			states[i] = objects[id-1]
		}
	}
	return states, nil
}

func allocateDistributedValue(node DistributedValueNode) (Value, error) {
	base := newBaseValue()
	switch node.Kind {
	case "bool":
		return &BoolValue{BaseValue: base, Val: node.Flag}, nil
	case "int":
		if node.Integer < math.MinInt32 || node.Integer > math.MaxInt32 {
			return nil, fmt.Errorf("network integer is outside int32 range")
		}
		return &IntValue{BaseValue: base, Val: int32(node.Integer)}, nil
	case "string":
		return &StringValue{BaseValue: base}, nil
	case "model":
		return &ModelValue{BaseValue: base, Index: node.ModelIndex, Type: node.ModelType}, nil
	case "interval":
		return &IntervalValue{BaseValue: base, Low: node.Low, High: node.High}, nil
	case "tuple":
		return &TupleValue{BaseValue: base}, nil
	case "record":
		return &RecordValue{BaseValue: base, IsNorm: node.Flag}, nil
	case "enum":
		return &SetEnumValue{BaseValue: base, IsNorm: node.Flag}, nil
	case "function":
		return &FcnRcdValue{BaseValue: base, IsNorm: node.Flag}, nil
	case "operatorRecord":
		return NewOpRcdValue(), nil
	case "lambda":
		return &FcnLambdaValue{BaseValue: base}, nil
	case "lazy":
		return &LazyValue{BaseValue: base}, nil
	case "lazySupplier":
		return &LazySupplierValue{LazyValue: &LazyValue{BaseValue: base}}, nil
	case "predicate":
		return &SetPredValue{BaseValue: base, Converted: true}, nil
	case "product":
		return &SetOfTuplesValue{BaseValue: base, TupleSetDummy: node.Dummy}, nil
	case "recordSet":
		return &SetOfRcdsValue{BaseValue: base, RcdSetDummy: node.Dummy}, nil
	case "functionSet":
		return &SetOfFcnsValue{BaseValue: base, FcnSetDummy: node.Dummy}, nil
	case "kSubset":
		if node.Integer < math.MinInt32 || node.Integer > math.MaxInt32 {
			return nil, fmt.Errorf("network subset rank is outside int32 range")
		}
		return &KSubsetValue{BaseValue: base, K: int(node.Integer), PSetDummy: node.Dummy}, nil
	case "subset":
		return &SubsetValue{BaseValue: base, PSetDummy: node.Dummy}, nil
	case "cup":
		return &SetCupValue{BaseValue: base, CupSetDummy: node.Dummy}, nil
	case "cap":
		return &SetCapValue{BaseValue: base, CapSetDummy: node.Dummy}, nil
	case "difference":
		return &SetDiffValue{BaseValue: base, DiffSetDummy: node.Dummy}, nil
	case "union":
		return &UnionValue{BaseValue: base, RealSetDummy: node.Dummy}, nil
	case "undefined":
		return &UndefValue{BaseValue: base}, nil
	case "counterexample":
		return &CounterExample{}, nil
	case "naturals":
		return &UserValue{BaseValue: base, UserObj: naturalsObj{}}, nil
	case "integers":
		return &UserValue{BaseValue: base, UserObj: integersObj{}}, nil
	case "strings":
		return &UserValue{BaseValue: base, UserObj: stringsObj{}}, nil
	case "any":
		return &UserValue{BaseValue: base, UserObj: AnySet{}}, nil
	case "sequences":
		if node.Integer < math.MinInt32 || node.Integer > math.MaxInt32 {
			return nil, fmt.Errorf("network sequence bound is outside int32 range")
		}
		return &UserValue{BaseValue: base, UserObj: &sequencesObj{SizeBound: int(node.Integer)}}, nil
	default:
		return nil, fmt.Errorf("unknown network value kind %q", node.Kind)
	}
}

func (d *distributedPayloadDecoder) value(id int) (Value, error) {
	if id < 0 || id > len(d.values) {
		return nil, fmt.Errorf("invalid distributed value reference %d", id)
	}
	if id == 0 {
		return nil, nil
	}
	return d.values[id-1], nil
}

func (d *distributedPayloadDecoder) string(id int) (*UniqueString, error) {
	if id < 0 || id > len(d.strings) {
		return nil, fmt.Errorf("invalid distributed string reference %d", id)
	}
	if id == 0 {
		return nil, nil
	}
	return d.strings[id-1], nil
}

func (d *distributedPayloadDecoder) refs(ids []int, isNil bool) ([]Value, error) {
	if isNil {
		if len(ids) != 0 {
			return nil, fmt.Errorf("null value array contains references")
		}
		return nil, nil
	}
	values := make([]Value, len(ids))
	for i, id := range ids {
		value, err := d.value(id)
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	return values, nil
}

func (d *distributedPayloadDecoder) arrayRefs(ids []int, isNil bool, array int) ([]Value, error) {
	if array == 0 {
		return d.refs(ids, isNil)
	}
	if array < 0 || array > len(d.arrays) {
		return nil, fmt.Errorf("invalid value array reference %d", array)
	}
	if isNil || len(ids) != 0 {
		return nil, fmt.Errorf("value array reference conflicts with inline/null values")
	}
	return d.arrays[array-1], nil
}

func (d *distributedPayloadDecoder) names(ids []int, isNil bool) ([]*UniqueString, error) {
	if isNil {
		if len(ids) != 0 {
			return nil, fmt.Errorf("null name array contains references")
		}
		return nil, nil
	}
	names := make([]*UniqueString, len(ids))
	for i, id := range ids {
		name, err := d.string(id)
		if err != nil {
			return nil, err
		}
		names[i] = name
	}
	return names, nil
}

func (d *distributedPayloadDecoder) nameArrayRefs(ids []int, isNil bool, array int) ([]*UniqueString, error) {
	if array == 0 {
		return d.names(ids, isNil)
	}
	if array < 0 || array > len(d.nameArrays) {
		return nil, fmt.Errorf("invalid name array reference %d", array)
	}
	if isNil || len(ids) != 0 {
		return nil, fmt.Errorf("name array reference conflicts with inline/null names")
	}
	return d.nameArrays[array-1], nil
}

func distributedValueCast[T Value](value Value) (T, error) {
	var zero T
	if value == nil {
		return zero, nil
	}
	typed, ok := value.(T)
	if !ok {
		return zero, fmt.Errorf("network value %T cannot be used as %T", value, zero)
	}
	return typed, nil
}

func (d *distributedPayloadDecoder) populate(value Value, node DistributedValueNode) error {
	refs, err := d.arrayRefs(node.References, node.ReferencesNil, node.ReferencesArray)
	if err != nil {
		return err
	}
	domain, err := d.arrayRefs(node.Domain, node.DomainNil, node.DomainArray)
	if err != nil {
		return err
	}
	names, err := d.nameArrayRefs(node.Names, node.NamesNil, node.NamesArray)
	if err != nil {
		return err
	}
	cache, err := d.value(node.Cache)
	if err != nil {
		return err
	}
	name, err := d.string(node.String)
	if err != nil {
		return err
	}
	// Constructors with a fixed arity must not index malformed wire arrays.
	want := -1
	switch node.Kind {
	case "lambda", "lazy", "lazySupplier", "predicate", "subset", "kSubset", "union", "counterexample", "sequences":
		want = 1
	case "functionSet", "cup", "cap", "difference":
		want = 2
	}
	if want >= 0 && len(refs) != want {
		return fmt.Errorf("network %s requires %d references, got %d", node.Kind, want, len(refs))
	}
	var set *SetEnumValue
	switch node.Kind {
	case "product", "recordSet", "functionSet", "subset", "kSubset", "cup", "cap", "difference", "union":
		set, err = distributedValueCast[*SetEnumValue](cache)
		if err != nil {
			return err
		}
	}
	switch v := value.(type) {
	case *StringValue:
		v.Val = name
	case *ModelValue:
		v.Val = name
		v.Data, err = d.modelData(node)
	case *TupleValue:
		v.Elems = refs
	case *RecordValue:
		if len(names) != len(refs) {
			return fmt.Errorf("record has unequal names and values")
		}
		v.Names, v.Values = names, refs
	case *SetEnumValue:
		if node.Vector != 0 {
			if node.Vector < 0 || node.Vector > len(d.vectors) {
				return fmt.Errorf("invalid value vector reference %d", node.Vector)
			}
			if !node.CollectionPresent || len(refs) != 0 || node.ReferencesArray != 0 {
				return fmt.Errorf("value vector reference conflicts with inline/null collection")
			}
			v.Elems = d.vectors[node.Vector-1]
		} else if node.CollectionPresent {
			v.Elems = &ValueVec{data: refs}
		} else if len(refs) != 0 {
			return fmt.Errorf("null value vector contains elements")
		}
	case *FcnRcdValue:
		v.Intv, err = distributedValueCast[*IntervalValue](cache)
		if err != nil {
			return err
		}
		if v.Intv == nil && len(domain) != len(refs) {
			return fmt.Errorf("function has unequal domain and values")
		}
		v.Domain, v.Values = domain, refs
	case *OpRcdValue:
		if node.OperatorDomainArray != 0 {
			if node.OperatorDomainNil || len(node.OperatorDomain) != 0 {
				return fmt.Errorf("operator domain array conflicts with inline or null rows")
			}
			v.Domain, err = d.valueRows(node.OperatorDomainArray)
			if err != nil {
				return err
			}
		} else if node.OperatorDomainNil {
			if len(node.OperatorDomain) != 0 {
				return fmt.Errorf("null operator domain contains rows")
			}
		} else {
			v.Domain = make([][]Value, len(node.OperatorDomain))
			for i, row := range node.OperatorDomain {
				v.Domain[i], err = d.arrayRefs(row.References, row.Nil, row.Array)
				if err != nil {
					return err
				}
			}
		}
		v.Values = refs
	case *FcnLambdaValue:
		v.FcnRcd, err = distributedValueCast[*FcnRcdValue](refs[0])
	case *LazyValue:
		v.Val = refs[0]
	case *LazySupplierValue:
		v.Val = refs[0]
	case *SetPredValue:
		v.InVal = refs[0]
	case *SetOfTuplesValue:
		v.Sets, v.TupleSet = refs, set
	case *SetOfRcdsValue:
		if len(names) != len(refs) {
			return fmt.Errorf("record set has unequal names and values")
		}
		v.Names, v.Values, v.RcdSet = names, refs, set
	case *SetOfFcnsValue:
		v.Domain, v.Range, v.FcnSet = refs[0], refs[1], set
	case *KSubsetValue:
		v.Set, v.PSet = refs[0], set
	case *SubsetValue:
		v.Set, v.PSet = refs[0], set
	case *SetCupValue:
		v.Set1, v.Set2, v.CupSet = refs[0], refs[1], set
	case *SetCapValue:
		v.Set1, v.Set2, v.CapSet = refs[0], refs[1], set
	case *SetDiffValue:
		v.Set1, v.Set2, v.DiffSet = refs[0], refs[1], set
	case *UnionValue:
		v.Set, v.RealSet = refs[0], set
	case *CounterExample:
		v.RecordValue, err = distributedValueCast[*RecordValue](refs[0])
	case *UserValue:
		if sequence, ok := v.UserObj.(*sequencesObj); ok {
			sequence.Range = refs[0]
		}
	}
	return err
}

func (d *distributedPayloadDecoder) valueRows(id int) ([][]Value, error) {
	if id < 0 || id > len(d.rows) {
		return nil, fmt.Errorf("invalid value-row array reference %d", id)
	}
	if id == 0 {
		return nil, nil
	}
	return d.rows[id-1], nil
}

func (d *distributedPayloadDecoder) modelData(node DistributedValueNode) (any, error) {
	switch node.DataKind {
	case "":
		return nil, nil
	case "string":
		return node.DataString, nil
	case "uniqueString":
		return d.string(node.DataName)
	case "bitVector":
		if node.DataArray < 0 || node.DataArray > len(d.bitVectors) {
			return nil, fmt.Errorf("invalid attached bit vector reference %d", node.DataArray)
		}
		if node.DataArray == 0 {
			return (*BitVector)(nil), nil
		}
		return d.bitVectors[node.DataArray-1], nil
	case "longVector":
		if node.DataArray < 0 || node.DataArray > len(d.longVectors) {
			return nil, fmt.Errorf("invalid attached long vector reference %d", node.DataArray)
		}
		if node.DataArray == 0 {
			return (*LongVec)(nil), nil
		}
		return d.longVectors[node.DataArray-1], nil
	case "stateVector":
		if node.DataArray < 0 || node.DataArray > len(d.stateVectors) {
			return nil, fmt.Errorf("invalid attached state vector reference %d", node.DataArray)
		}
		if node.DataArray == 0 {
			return (*StateVec)(nil), nil
		}
		return d.stateVectors[node.DataArray-1], nil
	case "stateVectorArray":
		return d.stateVectorArray(node.DataArray)
	case "longVectorArray":
		return d.longVectorArray(node.DataArray)
	case "nameArray":
		return d.nameArrayRefs(nil, node.DataArray == 0, node.DataArray)
	case "state":
		if node.DataState < 0 || node.DataState > len(d.states) {
			return nil, fmt.Errorf("invalid model state reference %d", node.DataState)
		}
		if node.DataState == 0 {
			return (*TLCStateMut)(nil), nil
		}
		return d.states[node.DataState-1], nil
	case "stateArray":
		if node.DataArray < 0 || node.DataArray > len(d.stateArrays) {
			return nil, fmt.Errorf("invalid model state-array reference %d", node.DataArray)
		}
		if node.DataArray == 0 {
			return []*TLCStateMut(nil), nil
		}
		return d.stateArrays[node.DataArray-1], nil
	case "bool":
		return node.DataBool, nil
	case "int":
		value := int(node.DataInteger)
		if int64(value) != node.DataInteger {
			return nil, fmt.Errorf("model data outside native int range")
		}
		return value, nil
	case "int32":
		if node.DataInteger < math.MinInt32 || node.DataInteger > math.MaxInt32 {
			return nil, fmt.Errorf("model data outside int32 range")
		}
		return int32(node.DataInteger), nil
	case "int8":
		if node.DataInteger < math.MinInt8 || node.DataInteger > math.MaxInt8 {
			return nil, fmt.Errorf("model data outside int8 range")
		}
		return int8(node.DataInteger), nil
	case "int16":
		if node.DataInteger < math.MinInt16 || node.DataInteger > math.MaxInt16 {
			return nil, fmt.Errorf("model data outside int16 range")
		}
		return int16(node.DataInteger), nil
	case "uint16":
		if node.DataInteger < 0 || node.DataInteger > math.MaxUint16 {
			return nil, fmt.Errorf("model data outside uint16 range")
		}
		return uint16(node.DataInteger), nil
	case "int64":
		return node.DataInteger, nil
	case "float64":
		return math.Float64frombits(node.DataFloatBits), nil
	case "float32":
		if node.DataFloatBits > math.MaxUint32 {
			return nil, fmt.Errorf("model float32 data outside 32-bit representation")
		}
		return math.Float32frombits(uint32(node.DataFloatBits)), nil
	case "boolArray", "intArray", "int8Array", "int16Array", "int32Array", "int64Array", "uint64Array", "uint16Array", "float32Array", "float64Array":
		if node.DataArray < 0 || node.DataArray > len(d.primitiveArrays) {
			return nil, fmt.Errorf("invalid primitive array reference %d", node.DataArray)
		}
		if node.DataArray == 0 {
			return decodeDistributedPrimitiveArray(DistributedPrimitiveArrayNode{Kind: node.DataKind}, true)
		}
		if d.primitiveArrayKinds[node.DataArray-1] != node.DataKind {
			return nil, fmt.Errorf("primitive array kind mismatch: %s", node.DataKind)
		}
		return d.primitiveArrays[node.DataArray-1], nil
	case "stringArray":
		if node.DataArray < 0 || node.DataArray > len(d.stringArrays) {
			return nil, fmt.Errorf("invalid string array reference %d", node.DataArray)
		}
		if node.DataArray == 0 {
			return []string(nil), nil
		}
		return d.stringArrays[node.DataArray-1], nil
	case "bytes":
		if node.DataBytes < 0 || node.DataBytes > len(d.bytes) {
			return nil, fmt.Errorf("invalid model byte data reference %d", node.DataBytes)
		}
		if node.DataBytes == 0 {
			return []byte(nil), nil
		}
		return d.bytes[node.DataBytes-1], nil
	case "value":
		return d.value(node.DataValue)
	case "valueArray":
		return d.arrayRefs(nil, node.DataArray == 0, node.DataArray)
	case "valueRows":
		return d.valueRows(node.DataArray)
	case "valueMap":
		if node.DataMap < 0 || node.DataMap > len(d.valueMaps) {
			return nil, fmt.Errorf("invalid model value-map reference %d", node.DataMap)
		}
		if node.DataMap == 0 {
			return map[string]Value(nil), nil
		}
		return d.valueMaps[node.DataMap-1], nil
	case "objectArray":
		if node.DataArray < 0 || node.DataArray > len(d.objectArrays) {
			return nil, fmt.Errorf("invalid model object-array reference %d", node.DataArray)
		}
		if node.DataArray == 0 {
			return []any(nil), nil
		}
		return d.objectArrays[node.DataArray-1], nil
	case "objectMap":
		if node.DataMap < 0 || node.DataMap > len(d.objectMaps) {
			return nil, fmt.Errorf("invalid model object-map reference %d", node.DataMap)
		}
		if node.DataMap == 0 {
			return map[string]any(nil), nil
		}
		return d.objectMaps[node.DataMap-1], nil
	case "objectKeyMap":
		if node.DataMap < 0 || node.DataMap > len(d.objectKeyMaps) {
			return nil, fmt.Errorf("invalid model object-key-map reference %d", node.DataMap)
		}
		if node.DataMap == 0 {
			return map[any]any(nil), nil
		}
		return d.objectKeyMaps[node.DataMap-1], nil
	default:
		return nil, fmt.Errorf("unknown model data kind %q", node.DataKind)
	}
}
