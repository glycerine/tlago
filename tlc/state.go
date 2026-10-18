package tlc

import (
	"iter"
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
	mutableStateRuntimes      [2]mutableStateRuntime
	EmptyState                *TLCStateMut
)

// Each concrete mutable state class owns its tool, captured VIEW and symmetry.
type mutableStateRuntime struct {
	tool         *Tool
	view         SemanticNode
	permutations []*MVPerm
}

func selectedMutableStateRuntime() *mutableStateRuntime {
	if statePreserveMetadata {
		return &mutableStateRuntimes[1]
	}
	return &mutableStateRuntimes[0]
}

func (s *TLCStateMut) mutableRuntime() *mutableStateRuntime {
	if s.extended {
		return &mutableStateRuntimes[1]
	}
	return &mutableStateRuntimes[0]
}

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
		out[i].CountDistinct = node.GetCountDistinct()
	}
	return out
}

func stateVariablesForTool(tool *Tool) []StateVariable {
	if processor := tool.GetSpecProcessor(); processor != nil {
		nodes := processor.GetVariablesNodes()
		out := make([]StateVariable, len(nodes))
		for i, node := range nodes {
			out[i] = StateVariable{Name: node.Name, CountDistinct: node.GetCountDistinct(), declaration: node}
		}
		return out
	}
	return StateVariables()
}

func InitializeStateVariableCoverageCounters() {
	if stateVariableDeclarations != nil {
		for _, node := range stateVariableDeclarations {
			node.SetCountDistinct(NewCountDistinctSyncedHyperLogLog(10))
		}
		return
	}
	for i := range stateVariables {
		stateVariables[i].CountDistinct = NewCountDistinctSyncedHyperLogLog(10)
	}
}

func CountStateVariableCoverage(state *TLCStateMut) {
	if state == nil || !CoverageVariableEnabled() {
		return
	}
	countStateVariableCoverage(state)
}

func countStateVariableCoverage(state *TLCStateMut) {
	if stateVariableDeclarations != nil {
		for _, node := range stateVariableDeclarations {
			node.Count(state.Lookup(node.Name))
		}
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
	stateSymmetryPermutations = append([]*MVPerm(nil), perms...)
	selectedMutableStateRuntime().permutations = stateSymmetryPermutations
}

func StateSymmetryPermutations() []*MVPerm {
	return append([]*MVPerm(nil), selectedMutableStateRuntime().permutations...)
}

func SetTLCStateTool(tool *Tool) {
	stateTool = tool
	if tool == nil {
		// Native isolated runtimes explicitly release both class tool contexts.
		mutableStateRuntimes = [2]mutableStateRuntime{}
		stateSymmetryPermutations = nil
		statePreserveMetadata = false
		return
	}
	prepareTLCStateRuntimeMode(tool.GetMode())
	// Installing a tool selects its runtime context in addition to the concrete
	// state kind selected before specification processing.
	runtime := selectedMutableStateRuntime()
	runtime.view = tool.GetViewSpec()
	SetStateSymmetryPermutations(tool.GetSymmetryPerms())
	// DebugTool's copy constructor retains FastTool in the class context.
	if tool.ownedDebugFastTool() != nil {
		stateTool = tool.NoDebug()
	}
	runtime.tool = stateTool
}

// Tool's constructor selects TLCStateMut or TLCStateMutExt before SpecProcessor
// evaluates definitions. The runner calls this before its deferred tool loader.
func prepareTLCStateRuntimeMode(mode ToolMode) {
	statePreserveMetadata = usesExtendedStateMetadataForMode(mode)
	// Installing a tool selects the factory for subsequent states. Existing
	// states keep their construction kind, as the source's concrete classes do.
	if EmptyState != nil && !EmptyState.functional && EmptyState.printRecord == nil && EmptyState.extended != statePreserveMetadata {
		EmptyState = newEmptyStateOfKind(statePreserveMetadata)
	}
}

func ResetExtendedTLCStateTool(tool *Tool) *Tool {
	runtime := &mutableStateRuntimes[1]
	old := runtime.tool
	runtime.tool = tool
	return old
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

func IsEmptyTLCState(state TLCState) bool {
	if EmptyState == nil {
		return isNil(state)
	}
	return !isNil(state) && state == EmptyState
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
	extended    bool // Concrete mutable-state kind, captured at construction.
	// ENABLED uses TLCStateFun's persistent bindings, including declarations
	// outside the root module that have no slot in the mutable state vector.
	functionalBindings *TLCStateFun
}

func NewEmptyState() *TLCStateMut {
	return newEmptyStateOfKind(statePreserveMetadata)
}

// Queue and pool readers construct through the retained empty-state factory.
func newStateFromEmptyFactory() *TLCStateMut {
	if EmptyState == nil {
		panic(NewNullPointerException())
	}
	return EmptyState.CreateEmpty()
}

func newEmptyStateOfKind(extended bool) *TLCStateMut {
	return &TLCStateMut{
		extended: extended,
		WorkerID: TLCStateInitWorkerID,
		UID:      TLCStateInitUID,
		level:    TLCStateInitLevel,
		values:   make([]Value, len(stateVariables)),
	}
}

// TLCStateFun.Empty is shared across factories and carries no mutable bindings.
var emptyFunctionalState = &TLCStateMut{
	WorkerID:           TLCStateInitWorkerID,
	UID:                TLCStateInitUID,
	level:              TLCStateInitLevel,
	functional:         true,
	functionalBindings: TLCStateFunEmpty,
}

func NewFunctionalState() *TLCStateMut {
	return emptyFunctionalState
}

func (s *TLCStateMut) CreateEmpty() *TLCStateMut {
	if s == nil {
		panic(NewNullPointerException())
	}
	if s.printState != nil {
		return s.printState.CreateEmpty()
	}
	if s.functional {
		return NewFunctionalState()
	}
	return newEmptyStateOfKind(s.extended)
}

func (s *TLCStateMut) Bind(name *UniqueString, value Value) *TLCStateMut {
	if s == nil {
		panic(NewNullPointerException())
	}
	if s.printState != nil {
		return s.printState.Bind(name, value)
	}
	if s.functional {
		panic(NewWrongInvocationException("TLCStateFun.bind: This is a TLC bug."))
	}
	loc := s.mutableVariableIndex(name)
	if loc >= 0 && loc < len(s.values) {
		s.values[loc] = value
	}
	return s
}

// BindSymbol preserves the state-specific overload used by ENABLED. Mutable
// states only support binding by name; functional states retain the symbol in
// a new persistent binding without changing the input state.
func (s *TLCStateMut) BindSymbol(id *SymbolNode, value Value) *TLCStateMut {
	if s == nil {
		panic(NewNullPointerException())
	}
	if s.printState != nil {
		return s.printState.BindSymbol(id, value)
	}
	if !s.functional {
		panic(NewWrongInvocationException("TLCStateMut.bind: This is a TLC bug."))
	}
	s = s.functionalCopy()
	s.functionalBindings = NewTLCStateFun(id, value, s.functionalBindings)
	if id != nil && id.Name != nil {
		loc := id.Name.VarLoc()
		if loc >= 0 && loc < len(s.values) {
			s.values[loc] = value
		}
	}
	return s
}

func (s *TLCStateMut) BindWithSource(name *UniqueString, value Value, source SemanticNode) *TLCStateMut {
	if s == nil {
		panic(NewNullPointerException())
	}
	if s.printState != nil {
		return s.printState.BindWithSource(name, value, source)
	}
	if s.functional {
		panic(NewWrongInvocationException("TLCStateFun.bind: This is a TLC bug."))
	}
	loc := s.mutableVariableIndex(name)
	if loc >= 0 && loc < len(s.values) {
		s.values[loc] = value
		s.ensureSources()
		s.sources[loc] = source
	}
	return s
}

func (s *TLCStateMut) Unbind(name *UniqueString) *TLCStateMut {
	if s == nil {
		panic(NewNullPointerException())
	}
	if s.printState != nil {
		return s.printState.Unbind(name)
	}
	if s.functional {
		panic(NewWrongInvocationException("TLCStateFun.unbind: This is a TLC bug."))
	}
	loc := s.mutableVariableIndex(name)
	if loc >= 0 && loc < len(s.values) {
		s.values[loc] = nil
		if s.sources != nil {
			s.sources[loc] = nil
		}
	}
	return s
}

// Mutable-state writes index the value array even for a non-variable name.
// Lookup alone returns null for a negative variable location.
func (s *TLCStateMut) mutableVariableIndex(name *UniqueString) int {
	if name == nil {
		panic(NewNullPointerException())
	}
	loc := name.VarLoc()
	if loc < 0 || loc >= len(s.values) {
		panic(NewArrayIndexOutOfBoundsException(loc, len(s.values)))
	}
	return loc
}

func (s *TLCStateMut) functionalCopy() *TLCStateMut {
	values := make([]Value, len(stateVariables))
	copy(values, s.values)
	// A fresh TLCStateFun binding starts with base metadata defaults.
	return &TLCStateMut{
		WorkerID:           TLCStateInitWorkerID,
		UID:                TLCStateInitUID,
		level:              TLCStateInitLevel,
		values:             values,
		functional:         true,
		functionalBindings: s.functionalBindings,
	}
}

// lookupSymbol retains the source declaration name in functional states.
// Mutable states use the bridge's qualified key to index instantiated variables.
func (s *TLCStateMut) lookupSymbol(symbol *SymbolNode) Value {
	if s == nil {
		panic(NewNullPointerException())
	}
	if s.functional {
		return s.Lookup(symbol.GetName())
	}
	return s.Lookup(symbol.Name)
}

func (s *TLCStateMut) Lookup(name *UniqueString) Value {
	if s == nil {
		panic(NewNullPointerException())
	}
	if s.printState != nil {
		if s.printState.ContainsKey(name) {
			return s.printState.Lookup(name)
		}
		value, err := s.printRecord.Select(NewStringValueFromUnique(name))
		if err != nil {
			panic(err)
		}
		return value
	}
	if s.functional {
		return s.functionalBindings.Lookup(name)
	}
	if name == nil {
		panic(NewNullPointerException())
	}
	loc := name.VarLoc()
	if loc >= len(s.values) {
		panic(NewArrayIndexOutOfBoundsException(loc, len(s.values)))
	}
	if loc >= 0 && loc < len(s.values) && s.values[loc] != nil {
		return s.values[loc]
	}
	if s.printRecord != nil {
		value, err := s.printRecord.Select(NewStringValueFromUnique(name))
		if err != nil {
			panic(err)
		}
		return value
	}
	return nil
}

func (s *TLCStateMut) LookupName(name string) Value {
	if s == nil {
		panic(NewNullPointerException())
	}
	return s.Lookup(UniqueStringOf(name))
}

func (s *TLCStateMut) ContainsKey(name *UniqueString) bool {
	if s != nil && s.printState != nil {
		if s.printState.ContainsKey(name) {
			return true
		}
		if s.printRecord.ownedNames() == nil {
			panic(NewNullPointerException())
		}
		for _, field := range s.printRecord.ownedNames() {
			if field == name {
				return true
			}
		}
		return false
	}
	if !isNil(s.Lookup(name)) {
		return true
	}
	if s != nil && s.printRecord != nil {
		value, err := s.printRecord.Select(NewStringValueFromUnique(name))
		return err == nil && !isNil(value)
	}
	return false
}

func (s *TLCStateMut) Copy() *TLCStateMut {
	if s == nil {
		panic(NewNullPointerException())
	}
	if s.printState != nil {
		return s.printState.Copy()
	}
	if s.functional {
		return s
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
		extended:           s.extended,
		functionalBindings: s.functionalBindings,
	}
	if s.retainsExtendedMetadata() {
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
	if s == nil {
		panic(NewNullPointerException())
	}
	if s.printState != nil {
		return s.printState.DeepCopy()
	}
	if s.functional {
		panic(NewWrongInvocationException("TLCStateFun.deepCopy: This is a TLC bug."))
	}
	values := make([]Value, len(s.values))
	var sources []SemanticNode
	if s.sources != nil {
		sources = make([]SemanticNode, len(s.sources))
		copy(sources, s.sources)
	}
	for i, value := range s.values {
		if !isNil(value) {
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
		extended:           s.extended,
		functionalBindings: s.functionalBindings,
	}
	if s.retainsExtendedMetadata() {
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
	return s.extended && !s.functional && s.printRecord == nil
}

func (s *TLCStateMut) GetCached(key int) Value {
	if s == nil {
		panic(NewNullPointerException())
	}
	key = int(int32(key))
	if !s.retainsExtendedMetadata() || s.cached == nil {
		return nil
	}
	return s.cached[key]
}

func (s *TLCStateMut) SetCached(key int, value Value) Value {
	if s == nil {
		panic(NewNullPointerException())
	}
	key = int(int32(key))
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
	if s == nil {
		panic(NewNullPointerException())
	}
	// Ordinary TLCStateMut inherits the source no-op. Only extended states
	// retain deferred execution, just as they retain action metadata.
	if s.retainsExtendedMetadata() {
		s.callable = callable
	}
}

func (s *TLCStateMut) ExecCallable() (any, error) {
	if s == nil {
		panic(NewNullPointerException())
	}
	if s.callable == nil {
		return nil, nil
	}
	return s.callable()
}

func (s *TLCStateMut) AddToVec(states *StateVec) *StateVec {
	if s == nil {
		panic(NewNullPointerException())
	}
	if s.printState != nil {
		return s.printState.AddToVec(states)
	}
	if states == nil {
		states = NewStateVec(1)
	}
	return states.Add(s.Copy())
}

func (s *TLCStateMut) DeepNormalize() {
	if s == nil {
		panic(NewNullPointerException())
	}
	if s.printState != nil {
		s.printState.DeepNormalize()
		return
	}
	if s.functional {
		panic(NewWrongInvocationException("TLCStateFun.normalizeFcns: This is a TLC bug."))
	}
	for _, value := range s.values {
		if !isNil(value) {
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
	if s == nil {
		panic(NewNullPointerException())
	}
	if s.printState != nil {
		return s.printState.FingerPrint()
	}
	if s.functional {
		panic(NewWrongInvocationException("TLCStateFun.fingerPrint: This is a TLC bug."))
	}
	return s.fingerPrint(s.mutableRuntime().tool)
}

func (s *TLCStateMut) FingerPrintWithTool(tool *Tool) uint64 {
	if s == nil {
		panic(NewNullPointerException())
	}
	if s.printState != nil {
		return s.printState.FingerPrint()
	}
	if s.functional {
		panic(NewWrongInvocationException("TLCStateFun.fingerPrint: This is a TLC bug."))
	}
	if !s.extended {
		return s.FingerPrint()
	}
	return s.fingerPrint(tool)
}

func (s *TLCStateMut) fingerPrint(tool *Tool) uint64 {
	view := s.mutableRuntime().view
	values := s.symmetryRepresentativeValues()
	fp := FP64New()
	if view != nil {
		for _, value := range s.values {
			if isNil(value) {
				panic(NewNullPointerException())
			}
			value.DeepNormalize()
		}
		state := s
		if len(values) != 0 && len(s.values) != 0 && &values[0] != &s.values[0] {
			state = newEmptyStateOfKind(s.extended)
			state.values = values
		}
		if tool == nil {
			panic(NewNullPointerException())
		}
		value, err := tool.Eval(view, EmptyContext, state)
		if err != nil {
			panic(err)
		}
		if isNil(value) {
			panic(NewNullPointerException())
		}
		return value.FingerPrint(fp)
	}
	for _, value := range values {
		if isNil(value) {
			panic(NewNullPointerException())
		}
		fp = value.FingerPrint(fp)
	}
	if len(values) > 0 && &values[0] != &s.values[0] {
		for _, value := range s.values {
			if isNil(value) {
				panic(NewNullPointerException())
			}
			value.DeepNormalize()
		}
	}
	return fp
}

func (s *TLCStateMut) symmetryRepresentativeValues() []Value {
	permutations := s.mutableRuntime().permutations
	if len(permutations) == 0 {
		return s.values
	}
	size := len(s.values)
	minVals := s.values
	vals := make([]Value, size)
	usingOriginal := true

nextPerm:
	for _, perm := range permutations {
		cmp := 0
		for j := 0; j < size; j++ {
			if !isNil(s.values[j]) {
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
	if s == nil {
		panic(NewNullPointerException())
	}
	if s.printState != nil {
		return s.printState.AllAssigned()
	}
	if s.functional {
		return true
	}
	for _, value := range s.values {
		if isNil(value) {
			return false
		}
	}
	return true
}

func (s *TLCStateMut) NoneAssigned() bool {
	if s == nil {
		panic(NewNullPointerException())
	}
	if s.printState != nil {
		return s.printState.NoneAssigned()
	}
	if s.functional {
		return len(stateVariables) == 0
	}
	for _, value := range s.values {
		if !isNil(value) {
			return false
		}
	}
	return true
}

func (s *TLCStateMut) Unassigned() []StateVariable {
	if s == nil {
		panic(NewNullPointerException())
	}
	if s.printState != nil {
		return s.printState.Unassigned()
	}
	if s.functional {
		return nil
	}
	var out []StateVariable
	for i, value := range s.values {
		if !isNil(value) {
			continue
		}
		var variable StateVariable
		if stateVariableDeclarations != nil {
			if i >= len(stateVariableDeclarations) {
				panic(NewArrayIndexOutOfBoundsException(i, len(stateVariableDeclarations)))
			}
			declaration := stateVariableDeclarations[i]
			if declaration == nil {
				panic(NewNullPointerException())
			}
			variable = StateVariable{
				Name:          declaration.GetName(),
				CountDistinct: declaration.GetCountDistinct(),
				declaration:   declaration,
			}
		} else {
			if i >= len(stateVariables) {
				panic(NewArrayIndexOutOfBoundsException(i, len(stateVariables)))
			}
			variable = stateVariables[i]
		}
		// TreeSet validates even its first key and keeps the first declaration
		// when later unassigned slots have the same name.
		if variable.Name == nil {
			panic(NewNullPointerException())
		}
		name := variable.Name.String()
		index := sort.Search(len(out), func(j int) bool {
			return dotCompareString(out[j].Name.String(), name) >= 0
		})
		if index < len(out) && dotCompareString(out[index].Name.String(), name) == 0 {
			continue
		}
		out = append(out, StateVariable{})
		copy(out[index+1:], out[index:])
		out[index] = variable
	}
	return out
}

func (s *TLCStateMut) GetUnassigned() []StateVariable {
	return s.Unassigned()
}

func incompleteNextStateParams(tool *Tool, action *Action, state *TLCStateMut) []string {
	unassigned := state.Unassigned()
	singleAction := len(tool.GetActions()) == 1
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
	if singleAction {
		return []string{verb, strings.Join(names, ", ")}
	}
	if action == nil {
		panic(NewNullPointerException())
	}
	return []string{action.requireNameString(), verb, strings.Join(names, ", ")}
}

// GetVals is Java TLCState.getVals: put every declared variable, including an
// unassigned (null) value, in a fresh HashMap. MCState uses its key-set order.
// The indexed traversal reads the current retained declarations.
func (s *TLCStateMut) GetVals() *javaHashMap[*UniqueString, Value] {
	if s == nil {
		panic(NewNullPointerException())
	}
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
	for i := 0; i < stateVariableCount(stateVariables, stateVariableDeclarations); i++ {
		key := stateVariableNameAt(stateVariables, stateVariableDeclarations, i)
		values.Set(key, s.Lookup(key))
	}
	return values
}

func (s *TLCStateMut) GetVarsAsStrings() []string {
	if s == nil {
		panic(NewNullPointerException())
	}
	out := make([]string, stateVariableCount(stateVariables, stateVariableDeclarations))
	for i := range out {
		name := stateVariableNameAt(stateVariables, stateVariableDeclarations, i)
		if name == nil {
			panic(NewNullPointerException())
		}
		out[i] = name.String()
	}
	return out
}

func (s *TLCStateMut) GetVars() []*SymbolNode {
	if s == nil {
		panic(NewNullPointerException())
	}
	return stateVariableDeclarations
}

func (s *TLCStateMut) Values() *InsMap[*UniqueString, Value] {
	if s != nil && s.printRecord != nil {
		out := NewInsMap[*UniqueString, Value]()
		for i, name := range s.printRecord.ownedNames() {
			if i < len(s.printRecord.ownedValues()) {
				out.Set(name, s.printRecord.ownedValues()[i])
			}
		}
		return out
	}
	out := NewInsMap[*UniqueString, Value]()
	if s.functional {
		for _, variable := range stateVariables {
			out.Set(variable.Name, s.Lookup(variable.Name))
		}
		return out
	}
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

func (s *TLCStateMut) SetPredecessorInfo(info *TLCStateInfo) *TLCStateMut {
	if s == nil {
		panic(NewNullPointerException())
	}
	return s.SetPredecessor(info.GetOriginalState())
}

func (s *TLCStateMut) UnsetPredecessor() *TLCStateMut {
	if s == nil {
		panic(NewNullPointerException())
	}
	s.pred = nil
	return s
}

func (s *TLCStateMut) Predecessor() *TLCStateMut {
	if s == nil {
		panic(NewNullPointerException())
	}
	if s.pred == nil {
		return nil
	}
	return s.pred.(*TLCStateMut)
}

func (s *TLCStateMut) GetPredecessor() *TLCStateMut {
	return s.Predecessor()
}

func (s *TLCStateMut) TracePredecessor() TLCPredecessorState {
	if s == nil {
		panic(NewNullPointerException())
	}
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
		panic(NewTLCRuntimeException(ECTLCTraceTooLong, s.String()))
	}
	s.level = pred.Level() + 1
	if !s.retainsExtendedMetadata() {
		s.pred = nil
	}
}

func (s *TLCStateMut) Level() int {
	if s == nil {
		panic(NewNullPointerException())
	}
	return s.level
}

func (s *TLCStateMut) GetLevel() int {
	return s.Level()
}

func (s *TLCStateMut) IsInitial() bool {
	if s == nil {
		panic(NewNullPointerException())
	}
	return s.level == TLCStateInitLevel
}

func (s *TLCStateMut) EvalStateLevelAlias() *TLCStateMut {
	if s == nil {
		panic(NewNullPointerException())
	}
	if s.functional || s.printRecord != nil {
		return s
	}
	tool := s.mutableRuntime().tool
	if tool == nil {
		panic(NewNullPointerException())
	}
	return tool.EvalAlias(s, EmptyState)
}

func (s *TLCStateMut) SetAction(action *Action) *TLCStateMut {
	if s == nil {
		panic(NewNullPointerException())
	}
	if s.retainsExtendedMetadata() {
		s.action = action
	}
	return s
}

func (s *TLCStateMut) HasAction() bool {
	if s == nil {
		panic(NewNullPointerException())
	}
	return s.action != nil
}

func (s *TLCStateMut) GetAction() *Action {
	if s == nil {
		panic(NewNullPointerException())
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
			panic(NewTLCRuntimeException(ECTLCTraceTooLong, s.String()))
		}
		s.level = pred.level + 1
	}
	return s
}

// State helpers read the current declaration owners, with the native names-only
// setup retained for callers that do not have checked semantic declarations.
func stateVariableNameAt(variables []StateVariable, declarations []*SymbolNode, index int) *UniqueString {
	if declarations != nil {
		if index < 0 || index >= len(declarations) {
			panic(NewArrayIndexOutOfBoundsException(index, len(declarations)))
		}
		if declarations[index] == nil {
			panic(NewNullPointerException())
		}
		return declarations[index].GetName()
	}
	if index < 0 || index >= len(variables) {
		panic(NewArrayIndexOutOfBoundsException(index, len(variables)))
	}
	return variables[index].Name
}

func stateVariableCount(variables []StateVariable, declarations []*SymbolNode) int {
	if declarations != nil {
		return len(declarations)
	}
	return len(variables)
}

func (s *TLCStateMut) CopyWith(prototype *TLCStateMut) *TLCStateMut {
	if s == nil {
		panic(NewNullPointerException())
	}
	out := s.CreateEmpty()
	for i := 0; i < stateVariableCount(stateVariables, stateVariableDeclarations); i++ {
		key := stateVariableNameAt(stateVariables, stateVariableDeclarations, i)
		if !isNil(prototype.Lookup(key)) {
			out.Bind(key, s.Lookup(key))
		}
	}
	out.level = s.level
	return out
}

func (s *TLCStateMut) Equal(obj TLCState) bool {
	if s == nil {
		panic(NewNullPointerException())
	}
	if s != nil && s.printState != nil {
		return s.printState.Equal(obj)
	}
	other, ok := obj.(*TLCStateMut)
	if !ok {
		return false
	}
	if other == nil {
		return false
	}
	if s.functional || other.functional {
		return s == other
	}
	if other.printState != nil || s.extended != other.extended {
		return false // Source equality requires the same concrete mutable class.
	}
	for i := range s.values {
		if i >= len(other.values) {
			panic(NewArrayIndexOutOfBoundsException(i, len(other.values)))
		}
		if isNil(s.values[i]) {
			if !isNil(other.values[i]) {
				return false
			}
			continue
		}
		if isNil(other.values[i]) {
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
	if s == nil {
		panic(NewNullPointerException())
	}
	if s != nil && s.printState != nil {
		return s.printState.HashCode()
	}
	return int32(reflect.ValueOf(s).Pointer())
}

func (s *TLCStateMut) String() string {
	if s == nil {
		panic(NewNullPointerException())
	}
	if s.functional {
		return s.functionalBindings.String()
	}
	if s.printRecord != nil {
		return s.printRecord.StateString()
	}
	runtime := s.mutableRuntime()
	if UseView() && runtime.view != nil {
		if runtime.tool == nil {
			panic(NewNullPointerException())
		}
		value, err := runtime.tool.Eval(runtime.view, EmptyContext, s)
		if err != nil {
			panic(err)
		}
		return SemanticValueString(runtime.view, value)
	}
	count := stateVariableCount(stateVariables, stateVariableDeclarations)
	var out strings.Builder
	for i := 0; i < count; i++ {
		key := stateVariableNameAt(stateVariables, stateVariableDeclarations, i)
		value := s.Lookup(key)
		appendStateAssignment(&out, key, value, count != 1)
	}
	return out.String()
}

// StringForVariables selects toString(last) when variables are omitted, and
// toString(variables, last) for an explicit non-nil slice, including an empty one.
func (s *TLCStateMut) StringForVariables(last *TLCStateMut, vars ...*UniqueString) string {
	if s == nil {
		panic(NewNullPointerException())
	}
	if vars != nil {
		return s.StringForVariableArray(last, vars)
	}
	if s.functional {
		panic(NewWrongInvocationException("TLCStateFun.toString: This is a TLC bug."))
	}
	if s.printState != nil {
		return s.printState.StringForVariableArray(last, s.printRecord.ownedNames())
	}
	variables, declarations := stateVariables, stateVariableDeclarations
	vars = make([]*UniqueString, stateVariableCount(variables, declarations))
	for i := range vars {
		vars[i] = stateVariableNameAt(variables, declarations, i)
	}
	return s.StringForVariableArray(last, vars)
}

// StringForVariableArray exposes the nullable array argument of Java's explicit
// variable overload separately from Go's omitted variadic argument.
func (s *TLCStateMut) StringForVariableArray(last *TLCStateMut, vars []*UniqueString) string {
	if s == nil {
		panic(NewNullPointerException())
	}
	if s.functional {
		panic(NewWrongInvocationException("TLCStateFun.toString: This is a TLC bug."))
	}
	if s.printState != nil {
		return s.printState.StringForVariableArray(last, vars)
	}
	// Extended states cast lastState before inspecting the variable array.
	// Ordinary states use the base state lookup and accept other state kinds.
	if s.extended && last != nil && (last.functional || last.printRecord != nil || !last.extended) {
		panic(valueStreamClassCast(last, javaValueClassName(s)))
	}
	if vars == nil {
		panic(NewNullPointerException())
	}
	var out strings.Builder
	for _, key := range vars {
		value := s.Lookup(key)
		if last == nil {
			panic(NewNullPointerException())
		}
		previous := last.Lookup(key)
		if isNil(previous) {
			panic(NewNullPointerException())
		}
		equal, err := previous.Equal(value)
		if err != nil {
			panic(err)
		}
		if !equal {
			appendStateAssignment(&out, key, value, len(vars) != 1)
		}
	}
	return out.String()
}

func appendStateAssignment(out *strings.Builder, key *UniqueString, value Value, conjunction bool) {
	if key == nil {
		panic(NewNullPointerException())
	}
	if conjunction {
		out.WriteString("/\\ ")
	}
	out.WriteString(key.String())
	out.WriteString(" = ")
	out.WriteString(ValuesPPR(value))
	out.WriteString("\n")
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

func TLCStateIsSubset(s1 *TLCStateMut, s2 *TLCStateMut) (result PartialBoolean) {
	if s2 == nil || s2 == EmptyState {
		return PartialYes
	}
	if s1 == nil {
		s1 = EmptyState
	}
	if s1 == s2 {
		return PartialYes
	}
	defer func() {
		if failure := recover(); failure != nil {
			if err, ok := failure.(error); ok {
				_, fingerprint := err.(*FingerprintException)
				if fingerprint || javaRuntimeException(err) != nil {
					result = PartialMaybe
					return
				}
			}
			panic(failure)
		}
	}()
	// Java's enhanced for loop captures the declaration array on entry.
	variables, declarations := stateVariables, stateVariableDeclarations
	for i := 0; i < stateVariableCount(variables, declarations); i++ {
		key := stateVariableNameAt(variables, declarations, i)
		val2 := s2.Lookup(key)
		if isNil(val2) {
			continue
		}
		val1 := s1.Lookup(key)
		if isNil(val1) {
			return PartialNo
		}
		// Objects.equals accepts identical values without invoking their methods.
		if val1 == val2 {
			continue
		}
		equal, err := val1.Equal(val2)
		if err != nil {
			panic(err)
		}
		if !equal {
			return PartialNo
		}
	}
	return PartialYes
}

func IsStateSubset(s1 *TLCStateMut, s2 *TLCStateMut) PartialBoolean {
	return TLCStateIsSubset(s1, s2)
}

// CopyState and DeepCopyState expose Java's covariant state-copy methods
// through the polymorphic collection surface.
func (s *TLCStateMut) CopyState() TLCState     { return s.Copy() }
func (s *TLCStateMut) DeepCopyState() TLCState { return s.DeepCopy() }

type StateVec struct {
	states      []TLCState
	distributed bool
}

// Worker partitions retain TLCStateVec's growth without a SetBound limit.
// Both collection kinds index backing capacity rather than logical size.
// Native result decoding restores the worker policy with active capacity.
func newDistributedStateVec(capacity int) *StateVec {
	capacity = int(int32(capacity))
	if capacity < 0 {
		panic(NewNegativeArraySizeException(fmtInt(capacity)))
	}
	return &StateVec{states: make([]TLCState, 0, capacity), distributed: true}
}

func NewDefaultDistributedStateVec() *StateVec {
	return newDistributedStateVec(10)
}

func NewStateVec(capacity int) *StateVec {
	capacity = int(int32(capacity))
	if capacity < 0 {
		panic(NewNegativeArraySizeException(fmtInt(capacity)))
	}
	return &StateVec{states: make([]TLCState, 0, capacity)}
}

func NewStateVecWithState(state TLCState) *StateVec {
	return &StateVec{states: []TLCState{state}}
}

func NewStateVecCopy(other *StateVec) *StateVec {
	other.requireStateVec()
	states := make([]TLCState, len(other.states))
	copy(states, other.states)
	return &StateVec{states: states}
}

func NewStateVecFrom(states []*TLCStateMut) *StateVec {
	v := NewStateVec(len(states))
	for _, state := range states {
		v.Add(state)
	}
	return v
}

// NewStateVecFromStates retains Java StateVec(TLCState[])'s array ownership.
func NewStateVecFromStates(states []TLCState) *StateVec {
	if states == nil {
		panic(NewNullPointerException())
	}
	return &StateVec{states: states[:len(states):len(states)]}
}

func (v *StateVec) ElementAt(i int) TLCState {
	v.requireStateVec()
	i = int(int32(i))
	if i < 0 || i >= cap(v.states) {
		panic(NewArrayIndexOutOfBoundsException(i, cap(v.states)))
	}
	return v.states[:cap(v.states)][i]
}

func (v *StateVec) requireStateVec() {
	if v == nil {
		panic(NewNullPointerException())
	}
}

func (v *StateVec) Empty() bool   { return v.Size() == 0 }
func (v *StateVec) IsEmpty() bool { return v.Size() == 0 }
func (v *StateVec) Size() int {
	v.requireStateVec()
	return len(v.states)
}
func (v *StateVec) At(i int) *TLCStateMut {
	state := v.ElementAt(i)
	if state == nil {
		return nil
	}
	return state.(*TLCStateMut)
}
func (v *StateVec) First() *TLCStateMut { return v.At(0) }
func (v *StateVec) Last() *TLCStateMut  { return v.At(v.Size() - 1) }
func (v *StateVec) FirstState() TLCState {
	return v.ElementAt(0)
}
func (v *StateVec) LastState() TLCState {
	return v.ElementAt(v.Size() - 1)
}
func (v *StateVec) Clear() {
	v.requireStateVec()
	v.states = v.states[:0]
}
func (v *StateVec) Reset() { v.Clear() }

func (v *StateVec) Grow(add int) {
	v.requireStateVec()
	Globals.Lock()
	bound := int32(Globals.SetBound)
	Globals.Unlock()
	oldLen := int32(cap(v.states))
	if oldLen >= bound {
		panic(NewTLCRuntimeException(ECTLCTooManyPossibleStates))
	}
	newLen := oldLen + int32(add)
	if doubled := int32(2) * oldLen; doubled > newLen {
		newLen = doubled
	}
	if newLen > bound {
		newLen = bound
	}
	if newLen < 0 {
		panic(NewNegativeArraySizeException(fmtInt(int(newLen))))
	}
	next := make([]TLCState, len(v.states), int(newLen))
	copy(next, v.states)
	v.states = next
}

func (v *StateVec) EnsureCapacity(minCapacity int) {
	v.requireStateVec()
	minCapacity = int(int32(minCapacity))
	if cap(v.states) >= minCapacity {
		return
	}
	newCapacity := cap(v.states) + cap(v.states)
	if newCapacity < minCapacity {
		newCapacity = minCapacity
	}
	if newCapacity < 0 {
		panic(NewNegativeArraySizeException(fmtInt(newCapacity)))
	}
	next := make([]TLCState, len(v.states), newCapacity)
	copy(next, v.states)
	v.states = next
}

func (v *StateVec) Add(state TLCState) *StateVec {
	v.requireStateVec()
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
	v.requireStateVec()
	return v.Add(state.SetPredecessor(pred))
}

func (v *StateVec) AddNextElement(pred *TLCStateMut, action *Action, state *TLCStateMut) (any, error) {
	v.requireStateVec()
	if state == nil {
		panic(NewNullPointerException())
	}
	return v.Add(state.SetPredecessor(pred).SetAction(action)), nil
}

func (v *StateVec) AddElements(other *StateVec) *StateVec {
	v.requireStateVec()
	other.requireStateVec()
	target, source := v, other
	if source.Size() > target.Size() {
		target, source = source, target
	}
	size0, size1 := target.Size(), source.Size()
	// Capture the source array before growth, including self-appends. Publish
	// size only after every indexed write; bounded growth can fail mid-copy.
	sourceArray := source.states[:cap(source.states)]
	target.ensureCanAdd(size1)
	for i := 0; i < size1; i++ {
		target.ReplaceAt(size0+i, sourceArray[i])
	}
	target.states = target.states[:size0+size1]
	return target
}

func (v *StateVec) AddAll(other *StateVec) *StateVec {
	return v.AddElements(other)
}

func (v *StateVec) Remove(index int) {
	v.RemoveElement(index)
}

func (v *StateVec) RemoveElement(index int) {
	v.requireStateVec()
	last := v.ElementAt(len(v.states) - 1)
	v.ReplaceAt(index, last)
	v.states = v.states[:len(v.states)-1]
}

func (v *StateVec) RemoveAt(index int) {
	v.ReplaceAt(index, nil)
}

func (v *StateVec) Replace(index int, state TLCState) {
	v.ReplaceAt(index, state)
}

func (v *StateVec) ReplaceAt(index int, state TLCState) {
	v.requireStateVec()
	index = int(int32(index))
	if index < 0 || index >= cap(v.states) {
		panic(NewArrayIndexOutOfBoundsException(index, cap(v.states)))
	}
	v.states[:cap(v.states)][index] = state
}

func (v *StateVec) Copy() *StateVec {
	v.requireStateVec()
	out := NewStateVec(len(v.states))
	for _, state := range v.states {
		if isNil(state) {
			panic(NewNullPointerException())
		}
		out.Add(state.(interface{ CopyState() TLCState }).CopyState())
	}
	return out
}

func (v *StateVec) DeepCopy() *StateVec {
	v.requireStateVec()
	out := NewStateVec(len(v.states))
	for _, state := range v.states {
		if isNil(state) {
			panic(NewNullPointerException())
		}
		out.Add(state.(interface{ DeepCopyState() TLCState }).DeepCopyState())
	}
	return out
}

func (v *StateVec) DeepNormalize() {
	v.requireStateVec()
	for _, state := range v.states {
		if isNil(state) {
			panic(NewNullPointerException())
		}
		state.(interface{ DeepNormalize() }).DeepNormalize()
	}
}

func (v *StateVec) Contains(state TLCState) bool {
	v.requireStateVec()
	for _, candidate := range v.states {
		if isNil(candidate) {
			panic(NewNullPointerException())
		}
		fp := candidate.FingerPrint()
		if isNil(state) {
			panic(NewNullPointerException())
		}
		if fp == state.FingerPrint() {
			return true
		}
	}
	return false
}

func (v *StateVec) ToSlice() []*TLCStateMut {
	v.requireStateVec()
	out := make([]*TLCStateMut, len(v.states))
	for i := range out {
		out[i] = v.At(i)
	}
	return out
}

func (v *StateVec) Stream() iter.Seq[TLCState] {
	v.requireStateVec()
	states := v.states
	return func(yield func(TLCState) bool) {
		for _, state := range states {
			if !yield(state) {
				return
			}
		}
	}
}

func (v *StateVec) ToList() []TLCState {
	out := make([]TLCState, 0, v.Size())
	for state := range v.Stream() {
		out = append(out, state)
	}
	return out
}

func (v *StateVec) ToStateList() []TLCState {
	return v.ToList()
}

func (v *StateVec) ToRecords(appendState *TLCStateMut) []Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	values := make([]Value, 0, len(v.states)+1)
	for _, state := range v.states {
		if isNil(state) {
			panic(NewNullPointerException())
		}
		values = append(values, newRecordValueFromStateLookup(state.(interface{ Lookup(*UniqueString) Value })))
	}
	values = append(values, NewRecordValueFromState(appendState))
	return values
}

func (v *StateVec) ToRecordsFrom(from *TLCStateMut, appendState *TLCStateMut) []Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	reversed := make([]Value, 0, len(v.states)+1)
	reversed = append(reversed, NewRecordValueFromState(appendState))
	for i := len(v.states) - 1; i >= 0; i-- {
		state := v.states[i]
		if isNil(state) {
			panic(NewNullPointerException())
		}
		reversed = append(reversed, newRecordValueFromStateLookup(state.(interface{ Lookup(*UniqueString) Value })))
		if from == nil {
			panic(NewNullPointerException())
		}
		if from.FingerPrint() == state.FingerPrint() {
			break
		}
	}
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	return reversed
}

func (v *StateVec) String() string {
	v.requireStateVec()
	parts := make([]string, len(v.states))
	for i, state := range v.states {
		if isNil(state) {
			panic(NewNullPointerException())
		}
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
		v.EnsureCapacity(needed)
		return
	}
	v.Grow(add)
}
