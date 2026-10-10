package tlc

import (
	"math"
	"reflect"
	"sort"
	"strings"
)

const (
	TLCStateInitWorkerID int16 = int16(^uint16(0) >> 1)
	TLCStateInitUID      int64 = -1
	TLCStateInitLevel    int   = 1
)

type StateVariable struct {
	Name          *UniqueString
	Location      SourceLocation
	CountDistinct *CountDistinct
	declaration   *SymbolNode
}

func (v StateVariable) GetSourceLocation() SourceLocation {
	if v.declaration != nil {
		return v.declaration.GetSourceLocation()
	}
	return v.Location
}

var (
	stateVariables            []StateVariable
	stateVariableDeclarations []*SymbolNode
	stateSymmetryPermutations []*MVPerm
	stateTool                 *Tool
	statePreserveMetadata     bool
	EmptyState                *TLCStateMut
)

func SetStateVariables(names []string) {
	SetStateVariablesWithLocations(names, nil)
}

func SetStateVariablesWithLocations(names []string, locations map[string]SourceLocation) {
	stateVariableDeclarations = nil
	stateVariables = make([]StateVariable, len(names))
	SetUniqueStringVariableCount(len(names))
	for i, name := range names {
		us := UniqueStringOf(name)
		us.SetLoc(i)
		stateVariables[i] = StateVariable{Name: us, Location: locations[name]}
	}
	EmptyState = NewEmptyState()
}

// SetStateVariableDeclarations retains the declaration array, as TLCState does.
// Locations are read from the current syntax when metadata is requested.
func SetStateVariableDeclarations(nodes []*SymbolNode) {
	names := make([]string, len(nodes))
	for i, node := range nodes {
		names[i] = node.Name.String()
	}
	SetStateVariables(names)
	stateVariableDeclarations = nodes
}

func StateVariables() []StateVariable {
	out := make([]StateVariable, len(stateVariables))
	copy(out, stateVariables)
	for i, node := range stateVariableDeclarations {
		out[i].Name = node.Name
		out[i].declaration = node
	}
	return out
}

func InitializeStateVariableCoverageCounters() {
	for i := range stateVariables {
		stateVariables[i].CountDistinct = NewCountDistinctSyncedHyperLogLog(10)
	}
}

func CountStateVariableCoverage(state *TLCStateMut) {
	if state == nil || !CoverageVariableEnabled() {
		return
	}
	for i := range stateVariables {
		counter := stateVariables[i].CountDistinct
		if counter == nil {
			continue
		}
		var value Value
		if i < len(state.values) {
			value = state.values[i]
		}
		counter.AddValue(value)
	}
}

func SetStateSymmetryPermutations(perms []*MVPerm) {
	if len(perms) == 0 {
		stateSymmetryPermutations = nil
		return
	}
	stateSymmetryPermutations = make([]*MVPerm, len(perms))
	copy(stateSymmetryPermutations, perms)
}

func StateSymmetryPermutations() []*MVPerm {
	out := make([]*MVPerm, len(stateSymmetryPermutations))
	copy(out, stateSymmetryPermutations)
	return out
}

func SetTLCStateTool(tool *Tool) {
	stateTool = tool
	if tool == nil {
		SetStateSymmetryPermutations(nil)
		statePreserveMetadata = false
		return
	}
	statePreserveMetadata = tool.usesExtendedStateMetadata()
	SetStateSymmetryPermutations(tool.GetSymmetryPerms())
	// Java constructs FastTool first; DebugTool's copy constructor does not
	// replace TLCStateMutExt.mytool. Workers explicitly fingerprint with their
	// debug tool, while parameterless state fingerprints retain FastTool.
	if tool.DebugFastTool != nil {
		stateTool = tool.NoDebug()
	}
}

// TLCState is the polymorphic state surface used by state collections. The
// evaluator continues to operate on mutable states; collection implementations
// must retain each state's own fingerprint, equality and hash semantics.
type TLCState interface {
	FingerPrint() uint64
	FingerPrintWithTool(*Tool) uint64
	Equal(TLCState) bool
	HashCode() int32
	GetAction() *Action
	String() string
}

// TLCPredecessorState retains the predecessor operations used by SimulationWorker.
// Java dispatches these operations through TLCState, including custom states.
type TLCPredecessorState interface {
	TLCState
	TracePredecessor() TLCPredecessorState
	SetTracePredecessor(TLCPredecessorState)
	Level() int
	IsInitial() bool
}

type TLCStateMut struct {
	WorkerID    int16
	UID         int64
	level       int
	values      []Value
	sources     []SemanticNode
	pred        TLCPredecessorState
	action      *Action
	callable    func() (any, error)
	cached      map[int]Value
	printRecord *RecordValue
	printState  *TLCStateMut // Separate owner delegated to by RecordValue.PrintTLCState.
	functional  bool
	// ENABLED uses TLCStateFun's persistent bindings, including declarations
	// outside the root module that have no slot in the mutable state vector.
	functionalBindings *TLCStateFun
}

func NewEmptyState() *TLCStateMut {
	return &TLCStateMut{
		WorkerID: TLCStateInitWorkerID,
		UID:      TLCStateInitUID,
		level:    TLCStateInitLevel,
		values:   make([]Value, len(stateVariables)),
	}
}

func NewFunctionalState() *TLCStateMut {
	state := NewEmptyState()
	state.functional = true
	return state
}

func (s *TLCStateMut) CreateEmpty() *TLCStateMut {
	if s.printState != nil {
		return s.printState.CreateEmpty()
	}
	return NewEmptyState()
}

func (s *TLCStateMut) Bind(name *UniqueString, value Value) *TLCStateMut {
	if s != nil && s.printState != nil {
		return s.printState.Bind(name, value)
	}
	if s != nil && s.functional {
		s = s.functionalCopy()
		s.functionalBindings = NewTLCStateFun(&SymbolNode{Name: name}, value, s.functionalBindings)
	}
	loc := name.VarLoc()
	if loc >= 0 && loc < len(s.values) {
		s.values[loc] = value
	}
	return s
}

func (s *TLCStateMut) BindWithSource(name *UniqueString, value Value, source SemanticNode) *TLCStateMut {
	if s != nil && s.printState != nil {
		return s.printState.BindWithSource(name, value, source)
	}
	if s != nil && s.functional {
		s = s.functionalCopy()
		s.functionalBindings = NewTLCStateFun(&SymbolNode{Name: name}, value, s.functionalBindings)
	}
	loc := name.VarLoc()
	if loc >= 0 && loc < len(s.values) {
		s.values[loc] = value
		s.ensureSources()
		s.sources[loc] = source
	}
	return s
}

func (s *TLCStateMut) Unbind(name *UniqueString) *TLCStateMut {
	if s != nil && s.printState != nil {
		return s.printState.Unbind(name)
	}
	if s != nil && s.functional {
		s = s.functionalCopy()
	}
	loc := name.VarLoc()
	if loc >= 0 && loc < len(s.values) {
		s.values[loc] = nil
		if s.sources != nil {
			s.sources[loc] = nil
		}
	}
	return s
}

func (s *TLCStateMut) functionalCopy() *TLCStateMut {
	if s == nil {
		return NewFunctionalState()
	}
	values := make([]Value, len(s.values))
	copy(values, s.values)
	var sources []SemanticNode
	if s.sources != nil {
		sources = make([]SemanticNode, len(s.sources))
		copy(sources, s.sources)
	}
	return &TLCStateMut{
		WorkerID:           s.WorkerID,
		UID:                s.UID,
		level:              s.level,
		values:             values,
		sources:            sources,
		pred:               s.pred,
		action:             s.action,
		callable:           s.callable,
		cached:             s.cached,
		printRecord:        s.printRecord,
		functional:         true,
		functionalBindings: s.functionalBindings,
	}
}

func (s *TLCStateMut) Lookup(name *UniqueString) Value {
	if s == nil || name == nil {
		return nil
	}
	if s.printState != nil {
		if s.printState.ContainsKey(name) {
			return s.printState.Lookup(name)
		}
		value, err := s.printRecord.Select(NewStringValueFromUnique(name))
		if err != nil {
			return nil
		}
		return value
	}
	if s.functional {
		return s.functionalBindings.Lookup(name)
	}
	loc := name.VarLoc()
	if loc >= 0 && loc < len(s.values) && s.values[loc] != nil {
		return s.values[loc]
	}
	if s.printRecord != nil {
		value, err := s.printRecord.Select(NewStringValueFromUnique(name))
		if err != nil {
			return nil
		}
		return value
	}
	return nil
}

func (s *TLCStateMut) ContainsKey(name *UniqueString) bool {
	if s != nil && s.printState != nil {
		if s.printState.ContainsKey(name) {
			return true
		}
		for _, field := range s.printRecord.Names {
			if field == name {
				return true
			}
		}
		return false
	}
	if s.Lookup(name) != nil {
		return true
	}
	if s != nil && s.printRecord != nil {
		value, err := s.printRecord.Select(NewStringValueFromUnique(name))
		return err == nil && value != nil
	}
	return false
}

func (s *TLCStateMut) Copy() *TLCStateMut {
	if s.printState != nil {
		return s.printState.Copy()
	}
	values := make([]Value, len(s.values))
	copy(values, s.values)
	var sources []SemanticNode
	if s.sources != nil {
		sources = make([]SemanticNode, len(s.sources))
		copy(sources, s.sources)
	}
	out := &TLCStateMut{
		WorkerID:           TLCStateInitWorkerID,
		UID:                TLCStateInitUID,
		level:              s.level,
		values:             values,
		sources:            sources,
		printRecord:        s.printRecord,
		functional:         s.functional,
		functionalBindings: s.functionalBindings,
	}
	if statePreserveMetadata {
		// TLCStateMutExt.copy goes straight to copyExt, bypassing the
		// base level-copy hook. Its new state starts at the initial level.
		out.level = TLCStateInitLevel
		if pred := s.TracePredecessor(); pred != nil {
			out.SetTracePredecessor(pred)
		}
		out.SetAction(s.GetAction())
	}
	return out
}

func (s *TLCStateMut) DeepCopy() *TLCStateMut {
	if s.printState != nil {
		return s.printState.DeepCopy()
	}
	values := make([]Value, len(s.values))
	var sources []SemanticNode
	if s.sources != nil {
		sources = make([]SemanticNode, len(s.sources))
		copy(sources, s.sources)
	}
	for i, value := range s.values {
		if value != nil {
			values[i] = value.DeepCopy()
		}
	}
	out := &TLCStateMut{
		WorkerID:           s.WorkerID,
		UID:                s.UID,
		level:              s.level,
		values:             values,
		sources:            sources,
		printRecord:        s.printRecord,
		functional:         s.functional,
		functionalBindings: s.functionalBindings,
	}
	if statePreserveMetadata {
		// Extended deepCopy first applies the base metadata hook, then
		// copyExt recalculates the level through the predecessor setter.
		if pred := s.TracePredecessor(); pred != nil {
			out.SetTracePredecessor(pred)
		}
		out.SetAction(s.GetAction())
	}
	return out
}

// Functional states and record-backed print wrappers inherit TLCState's base
// methods independently of which mutable state implementation the tool selects.
func (s *TLCStateMut) retainsExtendedMetadata() bool {
	return statePreserveMetadata && !s.functional && s.printRecord == nil
}

func (s *TLCStateMut) GetCached(key int) Value {
	if s == nil || !s.retainsExtendedMetadata() || s.cached == nil {
		return nil
	}
	return s.cached[key]
}

func (s *TLCStateMut) SetCached(key int, value Value) Value {
	if s == nil {
		return value
	}
	// Base TLCState methods return null without allocating. Only the
	// extended mutable implementation overrides them with a cache.
	if !s.retainsExtendedMetadata() {
		return nil
	}
	if s.cached == nil {
		s.cached = make(map[int]Value)
	}
	s.cached[key] = value
	return value
}

func (s *TLCStateMut) SetCallable(callable func() (any, error)) {
	// Ordinary TLCStateMut inherits the source no-op. Only extended states
	// retain deferred execution, just as they retain action metadata.
	if s != nil && s.retainsExtendedMetadata() {
		s.callable = callable
	}
}

func (s *TLCStateMut) ExecCallable() (any, error) {
	if s == nil || s.callable == nil {
		return nil, nil
	}
	return s.callable()
}

func (s *TLCStateMut) AddToVec(states *StateVec) *StateVec {
	if s.printState != nil {
		return s.printState.AddToVec(states)
	}
	if states == nil {
		states = NewStateVec(1)
	}
	return states.Add(s.Copy())
}

func (s *TLCStateMut) DeepNormalize() {
	if s.printState != nil {
		s.printState.DeepNormalize()
		return
	}
	for _, value := range s.values {
		if value != nil {
			value.DeepNormalize()
		}
	}
}

func (s *TLCStateMut) AddCounts(counts *SemanticNodeLongTable) {
	if s == nil || counts == nil {
		return
	}
	if s.printState != nil {
		s.printState.AddCounts(counts)
		return
	}
	for _, source := range s.sources {
		if source != nil {
			counts.Add(source, 1)
		}
	}
}

func (s *TLCStateMut) Sources() []SemanticNode {
	if s != nil && s.printState != nil {
		return s.printState.Sources()
	}
	if s == nil || s.sources == nil {
		return nil
	}
	out := make([]SemanticNode, len(s.sources))
	copy(out, s.sources)
	return out
}

func (s *TLCStateMut) ensureSources() {
	if s.sources == nil {
		s.sources = make([]SemanticNode, len(s.values))
	}
}

func (s *TLCStateMut) FingerPrint() uint64 {
	return s.FingerPrintWithTool(stateTool)
}

func (s *TLCStateMut) FingerPrintWithTool(tool *Tool) uint64 {
	if s.printState != nil {
		return s.printState.FingerPrintWithTool(tool)
	}
	values := s.symmetryRepresentativeValues()
	fp := FP64New()
	if tool != nil && tool.ViewSpec != nil {
		for _, value := range s.values {
			if value == nil {
				panic(NewNullPointerException())
			}
			value.DeepNormalize()
		}
		state := s
		if len(values) != 0 && len(s.values) != 0 && &values[0] != &s.values[0] {
			state = NewEmptyState()
			state.values = values
		}
		value, err := tool.Eval(tool.ViewSpec, EmptyContext, state)
		if err != nil {
			panic(err)
		}
		if value == nil {
			panic(NewNullPointerException())
		}
		return value.FingerPrint(fp)
	}
	for _, value := range values {
		if value == nil {
			panic(NewNullPointerException())
		}
		fp = value.FingerPrint(fp)
	}
	if len(values) > 0 && &values[0] != &s.values[0] {
		for _, value := range s.values {
			if value == nil {
				panic(NewNullPointerException())
			}
			value.DeepNormalize()
		}
	}
	return fp
}

func (s *TLCStateMut) symmetryRepresentativeValues() []Value {
	if len(stateSymmetryPermutations) == 0 {
		return s.values
	}
	size := len(s.values)
	minVals := s.values
	vals := make([]Value, size)
	usingOriginal := true

nextPerm:
	for _, perm := range stateSymmetryPermutations {
		cmp := 0
		for j := 0; j < size; j++ {
			if s.values[j] != nil {
				vals[j] = s.values[j].Permute(perm)
			} else {
				panic(NewNullPointerException())
			}
			if cmp == 0 {
				var err error
				cmp, err = compareStateValues(vals[j], minVals[j])
				if err != nil {
					panic(err)
				}
				if cmp > 0 {
					continue nextPerm
				}
			}
		}
		if cmp < 0 {
			if usingOriginal {
				minVals = vals
				vals = make([]Value, size)
				usingOriginal = false
			} else {
				minVals, vals = vals, minVals
			}
		}
	}
	if !usingOriginal {
		for _, value := range s.values {
			if value != nil {
				value.DeepNormalize()
			}
		}
	}
	return minVals
}

func (s *TLCStateMut) AllAssigned() bool {
	if s.printState != nil {
		return s.printState.AllAssigned()
	}
	for _, value := range s.values {
		if value == nil {
			return false
		}
	}
	return true
}

func (s *TLCStateMut) NoneAssigned() bool {
	if s.printState != nil {
		return s.printState.NoneAssigned()
	}
	for _, value := range s.values {
		if value != nil {
			return false
		}
	}
	return true
}

func (s *TLCStateMut) Unassigned() []StateVariable {
	if s.printState != nil {
		return s.printState.Unassigned()
	}
	var out []StateVariable
	for i, value := range s.values {
		if value == nil && i < len(stateVariables) {
			out = append(out, stateVariables[i])
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name.String() < out[j].Name.String()
	})
	return out
}

func incompleteNextStateParams(tool *Tool, action *Action, state *TLCStateMut) []string {
	unassigned := state.Unassigned()
	names := make([]string, 0, len(unassigned))
	for _, variable := range unassigned {
		if variable.Name != nil {
			names = append(names, variable.Name.String())
		}
	}
	verb := " is"
	if len(unassigned) > 1 {
		verb = "s are"
	}
	if tool != nil && len(tool.GetActions()) == 1 {
		return []string{verb, strings.Join(names, ", ")}
	}
	return []string{actionName(action), verb, strings.Join(names, ", ")}
}

// GetVals is Java TLCState.getVals: put every declared variable, including an
// unassigned (null) value, in a fresh HashMap. MCState uses its key-set order;
// Values remains the declaration-ordered adapter used by RecordValue(TLCState).
func (s *TLCStateMut) GetVals() *javaHashMap[*UniqueString, Value] {
	values := newJavaHashMap[*UniqueString, Value](func(key *UniqueString) int32 {
		if key == nil {
			return 0
		}
		return javaFormatStringHash(key.String())
	}, nil)
	// UniqueString is not Comparable in Java. All keys have the same class, so
	// tieBreakOrder uses identity hashes. A native Go pointer supplies a stable
	// object identity here; equality lookup still searches both collision subtrees.
	values.tieBreak = func(a, b *UniqueString) int {
		aHash := uint32(reflect.ValueOf(a).Pointer()) & 0x7fffffff
		bHash := uint32(reflect.ValueOf(b).Pointer()) & 0x7fffffff
		if aHash <= bHash {
			return -1
		}
		return 1
	}
	for _, variable := range stateVariables {
		values.Set(variable.Name, s.Lookup(variable.Name))
	}
	return values
}

func (s *TLCStateMut) Values() *InsMap[*UniqueString, Value] {
	if s != nil && s.printRecord != nil {
		out := NewInsMap[*UniqueString, Value]()
		for i, name := range s.printRecord.Names {
			if i < len(s.printRecord.Values) {
				out.Set(name, s.printRecord.Values[i])
			}
		}
		return out
	}
	out := NewInsMap[*UniqueString, Value]()
	for i, variable := range stateVariables {
		if i < len(s.values) {
			out.Set(variable.Name, s.values[i])
		}
	}
	return out
}

func (s *TLCStateMut) SetPredecessor(pred *TLCStateMut) *TLCStateMut {
	s.SetTracePredecessor(pred)
	return s
}

func (s *TLCStateMut) UnsetPredecessor() *TLCStateMut {
	s.pred = nil
	return s
}

func (s *TLCStateMut) Predecessor() *TLCStateMut {
	if s.pred == nil {
		return nil
	}
	return s.pred.(*TLCStateMut)
}

func (s *TLCStateMut) TracePredecessor() TLCPredecessorState {
	// Metadata restoration can assign a typed nil mutable pointer. Java's null
	// predecessor must remain an empty interface at the polymorphic boundary.
	if pred, ok := s.pred.(*TLCStateMut); ok && pred == nil {
		return nil
	}
	return s.pred
}

func (s *TLCStateMut) SetTracePredecessor(pred TLCPredecessorState) {
	if s == nil {
		panic(NewNullPointerException())
	}
	// A typed nil mutable pointer also represents an absent predecessor.
	if mutable, ok := pred.(*TLCStateMut); ok && mutable == nil {
		pred = nil
	}
	// Extended source states store the predecessor before the base level
	// access, including when it fails for a missing predecessor or depth limit.
	if s.retainsExtendedMetadata() {
		s.pred = pred
	}
	if pred == nil {
		panic(NewNullPointerException())
	}
	if pred.Level() >= math.MaxInt32 {
		panic(newTLCError(ECTLCTraceTooLong, "%s", s.String()))
	}
	s.level = pred.Level() + 1
	if !s.retainsExtendedMetadata() {
		s.pred = nil
	}
}

func (s *TLCStateMut) Level() int {
	return s.level
}

func (s *TLCStateMut) IsInitial() bool {
	return s.level == TLCStateInitLevel
}

func (s *TLCStateMut) EvalStateLevelAlias() *TLCStateMut {
	if s == nil || stateTool == nil {
		return s
	}
	if alias := stateTool.EvalAlias(s, EmptyState); alias != nil {
		return alias
	}
	return s
}

func (s *TLCStateMut) SetAction(action *Action) *TLCStateMut {
	if s.retainsExtendedMetadata() {
		s.action = action
	}
	return s
}

func (s *TLCStateMut) HasAction() bool {
	return s != nil && s.action != nil
}

func (s *TLCStateMut) GetAction() *Action {
	if s == nil {
		return nil
	}
	return s.action
}

func (s *TLCStateMut) attachTraceMetadata(pred *TLCStateMut, action *Action) *TLCStateMut {
	if s == nil {
		return nil
	}
	s.pred = pred
	s.action = action
	if pred != nil {
		if pred.level >= math.MaxInt32 {
			panic(newTLCError(ECTLCTraceTooLong, "%s", s.String()))
		}
		s.level = pred.level + 1
	}
	return s
}

func (s *TLCStateMut) CopyWith(prototype *TLCStateMut) *TLCStateMut {
	out := NewEmptyState()
	for i, variable := range stateVariables {
		if prototype != nil && prototype.Lookup(variable.Name) != nil {
			out.values[i] = s.Lookup(variable.Name)
		}
	}
	out.level = s.level
	return out
}

func (s *TLCStateMut) Equal(obj TLCState) bool {
	if s != nil && s.printState != nil {
		return s.printState.Equal(obj)
	}
	other, ok := obj.(*TLCStateMut)
	if !ok {
		return false
	}
	if s == nil || other == nil {
		return s == other
	}
	if other.printState != nil {
		return false // Source mutable-state equals rejects a PrintTLCState object.
	}
	for i := range s.values {
		if i >= len(other.values) {
			panic(NewArrayIndexOutOfBoundsException(i, len(other.values)))
		}
		if s.values[i] == nil {
			if other.values[i] != nil {
				return false
			}
			continue
		}
		if other.values[i] == nil {
			return false
		}
		equal, err := s.values[i].Equal(other.values[i])
		if err != nil {
			panic(err)
		}
		if !equal {
			return false
		}
	}
	return true
}

// Java TLCStateMut deliberately inherits Object.hashCode despite overriding
// equals. Retain object identity here rather than substituting a fingerprint.
func (s *TLCStateMut) HashCode() int32 {
	if s != nil && s.printState != nil {
		return s.printState.HashCode()
	}
	return int32(reflect.ValueOf(s).Pointer())
}

func (s *TLCStateMut) String() string {
	if s != nil && s.printRecord != nil {
		return s.printRecord.StateString()
	}
	if UseView() && stateTool != nil && stateTool.ViewSpec != nil {
		value, err := stateTool.Eval(stateTool.ViewSpec, EmptyContext, s)
		if err != nil {
			panic(err)
		}
		return SemanticValueString(stateTool.ViewSpec, value)
	}
	return s.StringForVariables(nil)
}

func (s *TLCStateMut) StringForVariables(last *TLCStateMut, vars ...*UniqueString) string {
	if s != nil && s.printState != nil {
		if len(vars) == 0 {
			vars = s.printRecord.Names
		}
		return s.printState.StringForVariables(last, vars...)
	}
	if s != nil && s.printRecord != nil && len(vars) == 0 {
		return s.printRecord.StateString()
	}
	if len(vars) == 0 {
		vars = make([]*UniqueString, len(stateVariables))
		for i, variable := range stateVariables {
			vars[i] = variable.Name
		}
	}

	var b strings.Builder
	if len(vars) == 1 {
		key := vars[0]
		value := s.Lookup(key)
		if last == nil || !stateValuesEqual(last.Lookup(key), value) {
			b.WriteString(key.String())
			b.WriteString(" = ")
			b.WriteString(ValuesPPR(value))
			b.WriteString("\n")
		}
		return b.String()
	}

	for _, key := range vars {
		value := s.Lookup(key)
		if last != nil && stateValuesEqual(last.Lookup(key), value) {
			continue
		}
		b.WriteString("/\\ ")
		b.WriteString(key.String())
		b.WriteString(" = ")
		b.WriteString(ValuesPPR(value))
		b.WriteString("\n")
	}
	return b.String()
}

func stateValuesEqual(a, b Value) bool {
	if a == nil || b == nil {
		return a == b
	}
	eq, err := a.Equal(b)
	return err == nil && eq
}

func compareStateValues(a, b Value) (int, error) {
	if a == nil || b == nil {
		switch {
		case a == b:
			return 0, nil
		case a == nil:
			return -1, nil
		default:
			return 1, nil
		}
	}
	return a.Compare(b)
}

func valueString(v Value) string {
	if v == nil {
		return "<nil>"
	}
	return v.String()
}

func IsStateSubset(s1 *TLCStateMut, s2 *TLCStateMut) PartialBoolean {
	if s2 == nil || s2 == EmptyState {
		return PartialYes
	}
	if s1 == nil {
		s1 = EmptyState
	}
	if s1 == s2 {
		return PartialYes
	}
	for _, variable := range stateVariables {
		key := variable.Name
		val2 := s2.Lookup(key)
		if val2 == nil {
			continue
		}
		val1 := s1.Lookup(key)
		if val1 == nil {
			return PartialNo
		}
		equal, err := val1.Equal(val2)
		if err != nil {
			return PartialMaybe
		}
		if !equal {
			return PartialNo
		}
	}
	return PartialYes
}

// CopyState and DeepCopyState expose Java's covariant state-copy methods
// through the polymorphic collection surface.
func (s *TLCStateMut) CopyState() TLCState     { return s.Copy() }
func (s *TLCStateMut) DeepCopyState() TLCState { return s.DeepCopy() }

type StateVec struct {
	states      []TLCState
	distributed bool
}

// Worker partitions follow TLCStateVec rather than the tool's StateVec: their
// growth has no SetBound limit and element access addresses backing capacity.
// Native result decoding restores this collection policy with active capacity.
func newDistributedStateVec(capacity int) *StateVec {
	if capacity < 0 {
		panic(NewNegativeArraySizeException(fmtInt(capacity)))
	}
	return &StateVec{states: make([]TLCState, 0, capacity), distributed: true}
}

func NewStateVec(capacity int) *StateVec {
	if capacity < 0 {
		capacity = 0
	}
	return &StateVec{states: make([]TLCState, 0, capacity)}
}

func NewStateVecFrom(states []*TLCStateMut) *StateVec {
	v := NewStateVec(len(states))
	for _, state := range states {
		v.Add(state)
	}
	return v
}

// NewStateVecFromStates retains Java StateVec(TLCState[])'s array ownership.
func NewStateVecFromStates(states []TLCState) *StateVec { return &StateVec{states: states} }

func (v *StateVec) ElementAt(i int) TLCState {
	if v.distributed {
		if i < 0 || i >= cap(v.states) {
			panic(NewArrayIndexOutOfBoundsException(i, cap(v.states)))
		}
		return v.states[:cap(v.states)][i]
	}
	return v.states[i]
}

func (v *StateVec) Empty() bool   { return len(v.states) == 0 }
func (v *StateVec) IsEmpty() bool { return len(v.states) == 0 }
func (v *StateVec) Size() int     { return len(v.states) }
func (v *StateVec) At(i int) *TLCStateMut {
	state := v.ElementAt(i)
	if state == nil {
		return nil
	}
	return state.(*TLCStateMut)
}
func (v *StateVec) First() *TLCStateMut { return v.At(0) }
func (v *StateVec) Last() *TLCStateMut  { return v.At(v.Size() - 1) }
func (v *StateVec) Clear()              { v.states = v.states[:0] }
func (v *StateVec) Reset()              { v.states = v.states[:0] }

func (v *StateVec) Add(state TLCState) *StateVec {
	v.ensureCanAdd(1)
	v.states = append(v.states, state)
	return v
}

func (v *StateVec) AddElement(state *TLCStateMut) (any, error) {
	v.Add(state)
	return v, nil
}

func (v *StateVec) SetElement(state TLCState) (any, error) {
	v.Clear()
	v.Add(state)
	return v, nil
}

func (v *StateVec) HasStates() bool {
	return !v.IsEmpty()
}

func (v *StateVec) AddWithPredecessor(pred *TLCStateMut, state *TLCStateMut) *StateVec {
	return v.Add(state.SetPredecessor(pred))
}

func (v *StateVec) AddNextElement(pred *TLCStateMut, action *Action, state *TLCStateMut) (any, error) {
	if state != nil {
		state.SetPredecessor(pred).SetAction(action)
	}
	return v.Add(state), nil
}

func (v *StateVec) AddElements(other *StateVec) *StateVec {
	if v == nil {
		return other
	}
	if other == nil {
		return v
	}
	target := v
	source := other
	if source.Size() > target.Size() {
		target, source = source, target
	}
	target.ensureCanAdd(source.Size())
	target.states = append(target.states, source.states...)
	return target
}

func (v *StateVec) AddAll(other *StateVec) *StateVec {
	return v.AddElements(other)
}

func (v *StateVec) Remove(index int) {
	v.states[index] = v.states[len(v.states)-1]
	v.states = v.states[:len(v.states)-1]
}

func (v *StateVec) RemoveAt(index int) {
	v.Replace(index, nil)
}

func (v *StateVec) Replace(index int, state TLCState) {
	v.states[index] = state
}

func (v *StateVec) Copy() *StateVec {
	out := NewStateVec(len(v.states))
	for _, state := range v.states {
		out.Add(state.(interface{ CopyState() TLCState }).CopyState())
	}
	return out
}

func (v *StateVec) DeepCopy() *StateVec {
	out := NewStateVec(len(v.states))
	for _, state := range v.states {
		out.Add(state.(interface{ DeepCopyState() TLCState }).DeepCopyState())
	}
	return out
}

func (v *StateVec) DeepNormalize() {
	for _, state := range v.states {
		state.(interface{ DeepNormalize() }).DeepNormalize()
	}
}

func (v *StateVec) Contains(state TLCState) bool {
	fp := state.FingerPrint()
	for _, candidate := range v.states {
		if candidate.FingerPrint() == fp {
			return true
		}
	}
	return false
}

func (v *StateVec) ToSlice() []*TLCStateMut {
	out := make([]*TLCStateMut, len(v.states))
	for i := range out {
		out[i] = v.At(i)
	}
	return out
}

func (v *StateVec) ToList() []*TLCStateMut {
	return v.ToSlice()
}

func (v *StateVec) ToRecords(appendState *TLCStateMut) []Value {
	values := make([]Value, 0, len(v.states)+1)
	for _, state := range v.states {
		values = append(values, NewRecordValueFromInsMap(state.(interface {
			Values() *InsMap[*UniqueString, Value]
		}).Values()))
	}
	values = append(values, NewRecordValueFromInsMap(appendState.Values()))
	return values
}

func (v *StateVec) ToRecordsFrom(from *TLCStateMut, appendState *TLCStateMut) []Value {
	reversed := make([]Value, 0, len(v.states)+1)
	reversed = append(reversed, NewRecordValueFromInsMap(appendState.Values()))
	fromFP := from.FingerPrint()
	for i := len(v.states) - 1; i >= 0; i-- {
		state := v.states[i]
		reversed = append(reversed, NewRecordValueFromInsMap(state.(interface {
			Values() *InsMap[*UniqueString, Value]
		}).Values()))
		if state.FingerPrint() == fromFP {
			break
		}
	}
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	return reversed
}

func (v *StateVec) String() string {
	parts := make([]string, len(v.states))
	for i, state := range v.states {
		parts[i] = state.String()
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

func (v *StateVec) ensureCanAdd(add int) {
	if add <= 0 {
		return
	}
	needed := len(v.states) + add
	if needed <= cap(v.states) {
		return
	}
	if v.distributed {
		newCapacity := int(int32(cap(v.states)) * 2)
		if newCapacity < needed {
			newCapacity = needed
		}
		next := make([]TLCState, len(v.states), newCapacity)
		copy(next, v.states)
		v.states = next
		return
	}
	Globals.Lock()
	bound := Globals.SetBound
	Globals.Unlock()
	if cap(v.states) >= bound || needed > bound {
		panic(newTLCError(ECTLCTooManyPossibleStates, "too many possible states"))
	}
	newCap := cap(v.states) + add
	if doubled := 2 * cap(v.states); doubled > newCap {
		newCap = doubled
	}
	if newCap > bound {
		newCap = bound
	}
	next := make([]TLCState, len(v.states), newCap)
	copy(next, v.states)
	v.states = next
}
