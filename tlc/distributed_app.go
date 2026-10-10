package tlc

import (
	"fmt"
	"time"
)

// TLCApp is the runtime portion of Java's distributed application. The root
// loader builds its Tool through the production parser before this snapshot.
type TLCApp struct {
	Tool           *Tool
	Actions        []*Action
	Invariants     []*Action
	ImpliedInits   []*Action
	ImpliedActions []*Action
	checkDeadlock  bool
	config         string
	fromCheckpoint *string
	metadir        string
	metadataSet    bool
	fpSetConfig    *FPSetConfiguration
	sourceArrays   bool // Preserve parser-backed null boundaries across Tool replacement.
}

func NewTLCApp(tool *Tool, deadlock bool) *TLCApp {
	if tool == nil {
		panic(NewNullPointerException())
	}
	app := &TLCApp{Tool: tool, checkDeadlock: deadlock, config: tool.GetConfigFile(), sourceArrays: tool.SpecProcessor != nil}
	app.ImpliedInits = tool.GetImpliedInits()
	app.Invariants = tool.GetInvariants()
	app.ImpliedActions = tool.GetImpliedActions()
	app.Actions = tool.GetActions()
	return app
}

// NewTLCAppWithMetadata captures the loaded tool's runtime arrays, then finishes
// Java's full constructor with metadata setup. Recovery is performed by create
// before loading, not by this constructor.
func NewTLCAppWithMetadata(tool *Tool, deadlock bool, fromCheckpoint *string, fpSetConfig *FPSetConfiguration) *TLCApp {
	app := NewTLCApp(tool, deadlock)
	app.fromCheckpoint = copyJavaMessage(fromCheckpoint)
	app.metadir = MakeMetaDir(time.Now(), tool.GetSpecDir(), fromCheckpoint)
	app.metadataSet = true
	app.fpSetConfig = fpSetConfig
	return app
}

func (a *TLCApp) GetFileName() string { return a.requireTool().GetRootFile() }
func (a *TLCApp) GetSpecDir() string  { return a.requireTool().GetSpecDir() }
func (a *TLCApp) GetConfigName() string {
	if a == nil {
		panic(NewNullPointerException())
	}
	return a.config
}
func (a *TLCApp) GetMetadir() string {
	if a == nil {
		panic(NewNullPointerException())
	}
	return a.metadir
}
func (a *TLCApp) CanRecover() bool {
	if a == nil {
		panic(NewNullPointerException())
	}
	return a.fromCheckpoint != nil
}
func (a *TLCApp) GetFPSetConfiguration() *FPSetConfiguration {
	if a == nil {
		panic(NewNullPointerException())
	}
	return a.fpSetConfig
}
func (a *TLCApp) GetModuleFiles() []*TLAFile {
	tool := a.requireTool()
	var options []FilenameResolverOptions
	if tool.DistributedFiles != nil {
		options = append(options, FilenameResolverOptions{Classpath: tool.DistributedFiles.filenameClasspath()})
	}
	return tool.GetModuleFiles(NewInJarFilenameToStream("/model/", options...))
}

func (a *TLCApp) GetCheckDeadlock() bool {
	if a == nil {
		panic(NewNullPointerException())
	}
	return a.checkDeadlock
}
func (a *TLCApp) GetPreprocess() bool {
	if a == nil {
		panic(NewNullPointerException())
	}
	return true
}
func (a *TLCApp) requireTool() *Tool {
	if a == nil || a.Tool == nil {
		panic(NewNullPointerException())
	}
	return a.Tool
}
func (a *TLCApp) GetInitStates(functor *StateFunctor) error {
	return a.requireTool().GetInitStates(functor)
}

func (a *TLCApp) GetNextStates(state *TLCStateMut) (*StateVec, error) {
	if a == nil {
		panic(NewNullPointerException())
	}
	out := NewStateVec(10)
	for i := 0; i < len(a.requireActionArray(a.Actions)); i++ {
		action := a.Actions[i]
		next, err := a.requireTool().GetNextStates(action, state)
		if err != nil {
			return nil, err
		}
		if next == nil {
			panic(NewNullPointerException())
		}
		// Java addElements chooses the larger vector as the receiver. This
		// changes order when a later action generates more states.
		out = out.AddElements(next)
	}
	if out.Size() == 0 && a.checkDeadlock {
		return nil, NewWorkerException("Error: deadlock reached.", state, nil, false)
	}
	// Source returns a separate array, allocated before state validation.
	// Keep the state objects shared, but not the mutable accumulator storage.
	states := make([]TLCState, out.Size())
	for i := 0; i < out.Size(); i++ {
		successor := out.At(i)
		if !a.requireTool().IsGoodState(successor) {
			return nil, NewWorkerException("Error: Successor state is not completely specified by the next-state action.", state, successor, false)
		}
		if i >= len(states) {
			panic(NewArrayIndexOutOfBoundsException(i, len(states)))
		}
		states[i] = successor
	}
	return NewStateVecFromStates(states), nil
}

func (a *TLCApp) CheckState(predecessor, successor *TLCStateMut) error {
	if a == nil {
		panic(NewNullPointerException())
	}
	for i := 0; i < len(a.requireActionArray(a.Invariants)); i++ {
		invariant := a.Invariants[i]
		valid, err := a.requireTool().IsValidState(invariant, successor)
		if err != nil {
			return err
		}
		if !valid {
			return NewWorkerException(fmt.Sprintf("Error: Invariant %s is violated.", a.propertyNameAt(a.Tool.GetInvNames(), i)), predecessor, successor, false)
		}
	}
	if predecessor == nil {
		for i := 0; i < len(a.requireActionArray(a.ImpliedInits)); i++ {
			implied := a.ImpliedInits[i]
			valid, err := a.requireTool().IsValidState(implied, successor)
			if err != nil {
				return err
			}
			if !valid {
				return NewWorkerException(fmt.Sprintf("Error: Implied-init %s is violated.", a.propertyNameAt(a.Tool.GetImpliedInitNames(), i)), predecessor, successor, false)
			}
		}
	} else {
		for i := 0; i < len(a.requireActionArray(a.ImpliedActions)); i++ {
			implied := a.ImpliedActions[i]
			valid, err := a.requireTool().IsValidTransition(implied, predecessor, successor)
			if err != nil {
				return err
			}
			if !valid {
				return NewWorkerException(fmt.Sprintf("Error: Implied-action %s is violated.", a.propertyNameAt(a.Tool.GetImpliedActNames(), i)), predecessor, successor, false)
			}
		}
	}
	return nil
}

func (a *TLCApp) requireActionArray(actions []*Action) []*Action {
	if a.sourceArrays && actions == nil {
		panic(NewNullPointerException())
	}
	return actions
}

func (a *TLCApp) propertyNameAt(names []string, index int) string {
	if a.sourceArrays && names == nil {
		panic(NewNullPointerException())
	}
	return distributedActionName(names, index)
}

func distributedActionName(names []string, index int) string {
	if index < 0 || index >= len(names) {
		panic(NewArrayIndexOutOfBoundsException(index, len(names)))
	}
	return names[index]
}

func (a *TLCApp) IsInModel(state *TLCStateMut) (bool, error) { return a.requireTool().IsInModel(state) }
func (a *TLCApp) IsInActions(predecessor, successor *TLCStateMut) (bool, error) {
	return a.requireTool().IsInActions(predecessor, successor)
}
func (a *TLCApp) GetState(fp uint64, previous ...any) (*TLCStateInfo, error) {
	return a.requireTool().GetState(fp, previous...)
}
func (a *TLCApp) GetStateForTransition(successor, predecessor *TLCStateMut) (*TLCStateInfo, error) {
	return a.requireTool().GetStateForTransition(successor, predecessor)
}
func (a *TLCApp) EvalAlias(current *TLCStateInfo, successor *TLCStateMut, prefix func() []*TLCStateInfo) (*TLCStateInfo, error) {
	return a.requireTool().EvalAliasInfo(current, successor, prefix)
}
func (a *TLCApp) SetCallStack() { a.Tool = NewCallStackTool(a.requireTool()) }
func (a *TLCApp) PrintCallStack() string {
	if a.requireTool().CallStack != nil {
		return a.Tool.CallStackString()
	}
	return fmt.Sprintf("%T@%p", a.Tool, a.Tool)
}
