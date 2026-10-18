package tlc

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

type StateVisitStatus int

const (
	StateVisitUnseen StateVisitStatus = iota
	StateVisitSeen
	StateVisitNotInModel
)

func stateVisitStatusHas(status, flag StateVisitStatus) bool {
	return status&flag == flag
}

type StateVisualization int

const (
	StateVisualizationStuttering StateVisualization = iota
	StateVisualizationDefault
	StateVisualizationDotted
)

func (v StateVisualization) String() string {
	switch v {
	case StateVisualizationStuttering:
		return "STUTTERING"
	case StateVisualizationDefault:
		return "DEFAULT"
	case StateVisualizationDotted:
		return "DOTTED"
	default:
		return fmt.Sprintf("IStateWriter.Visualization(%d)", v)
	}
}

// IStateWriter is the state-output contract used by TLC's checker. Keeping the
// contract separate from the default writer preserves Java's virtual dispatch
// for custom writers, including constrained transitions.
type IStateWriter interface {
	WriteInitState(*TLCStateMut) error
	WriteTransition(*TLCStateMut, *TLCStateMut, StateVisitStatus, *Action, ...SemanticNode) error
	WriteTransitionVisual(*TLCStateMut, *TLCStateMut, StateVisitStatus, *Action, StateVisualization, ...SemanticNode) error
	IsConstrained() bool
	IsNoop() bool
	IsDot() bool
	GetDumpFileName() string
	Snapshot() error
	Close() error
}

var _ IStateWriter = (*StateWriter)(nil)

// State and action writers use PrintWriter in Java. Suppress I/O failures after
// opening while evaluating labels, advancing state numbers and maintaining ranks.
// Custom writer callbacks and explicit snapshot file operations retain errors.
type statePrintWriter struct {
	buffer *JavaBufferedWriter
	closed bool
}

func (w *statePrintWriter) Write(data []byte) (int, error) {
	if !w.closed {
		_, _ = w.buffer.Write(data)
	}
	return len(data), nil
}

func (w *statePrintWriter) WriteString(data string) (int, error) {
	if !w.closed {
		_, _ = w.buffer.WriteString(data)
	}
	return len(data), nil
}

func (w *statePrintWriter) Flush() {
	if !w.closed {
		_ = w.buffer.Flush()
	}
}

type StateWriter struct {
	mu                  sync.Mutex
	Noop                bool
	Dot                 bool
	Constrained         bool
	ownersCaptured      bool
	noOpOwner           bool
	dotOwner            bool
	constrainedOwner    bool
	WriteInitStateFunc  func(*TLCStateMut) error
	WriteTransitionFunc func(*TLCStateMut, *TLCStateMut, StateVisitStatus, *Action, ...SemanticNode) error
	CloseFunc           func() error
	fname               string
	file                *os.File
	writer              *statePrintWriter
	stateNum            int
	colorize            bool
	actionLabels        bool
	snapshot            bool
	stuttering          bool
	strict              map[uint64]struct{}
	actionToColors      *javaHashMap[string, int]
	rankToNodes         *javaHashMap[int, *dotLongSet]
	colorGen            int
	closed              bool
}

type StateWriterFactory func() (*StateWriter, error)

var (
	stateWriterFactoryMu sync.RWMutex
	stateWriterFactories = map[string]StateWriterFactory{}
)

func init() {
	RegisterStateWriterClass("NoopStateWriter", func() (*StateWriter, error) {
		return NewNoopStateWriter(), nil
	})
	RegisterStateWriterClass("tlc2.util.NoopStateWriter", func() (*StateWriter, error) {
		return NewNoopStateWriter(), nil
	})
	RegisterStateWriterClass("DotStateWriter", func() (*StateWriter, error) {
		return NewDotStateWriter("DotStateWriter.dot", DotStateWriterOptions{})
	})
	RegisterStateWriterClass("tlc2.util.DotStateWriter", func() (*StateWriter, error) {
		return NewDotStateWriter("DotStateWriter.dot", DotStateWriterOptions{})
	})
}

func RegisterStateWriterClass(name string, factory StateWriterFactory) {
	name = strings.TrimSpace(name)
	if name == "" || factory == nil {
		return
	}
	stateWriterFactoryMu.Lock()
	stateWriterFactories[name] = factory
	stateWriterFactoryMu.Unlock()
}

func NewCustomStateWriter(className string) (*StateWriter, error) {
	className = strings.TrimSpace(className)
	if className == "" {
		return nil, fmt.Errorf("custom state writer class name is empty")
	}
	stateWriterFactoryMu.RLock()
	factory := stateWriterFactories[className]
	stateWriterFactoryMu.RUnlock()
	if factory == nil {
		return nil, fmt.Errorf("custom state writer class %s is not registered", className)
	}
	return factory()
}

func NewNoopStateWriter() *StateWriter {
	return &StateWriter{Noop: true, ownersCaptured: true, noOpOwner: true}
}

func NewStateDumpWriter(fname string) (*StateWriter, error) {
	file, err := openPoolFile(fname, true)
	if err != nil {
		ToolIOPrintln("Error: Unable to write to file " + fname)
		return nil, err
	}
	return &StateWriter{
		fname:          fname,
		file:           file,
		writer:         &statePrintWriter{buffer: NewJavaBufferedWriter(file)},
		stateNum:       1,
		ownersCaptured: true,
	}, nil
}

type DotStateWriterOptions struct {
	Colorize     bool
	ActionLabels bool
	Snapshot     bool
	Constrained  bool
	Stuttering   bool
	Strict       bool
	StrictPrefix bool
}

func NewDotStateWriter(fname string, opts DotStateWriterOptions) (*StateWriter, error) {
	w, err := NewStateDumpWriter(fname)
	if err != nil {
		return nil, err
	}
	w.Dot = true
	w.dotOwner = true
	w.colorize = opts.Colorize
	w.actionLabels = opts.ActionLabels
	w.snapshot = opts.Snapshot
	w.Constrained = opts.Constrained
	w.constrainedOwner = opts.Constrained
	w.stuttering = opts.Stuttering
	w.actionToColors = newJavaHashMap[string, int](javaFormatStringHash, dotCompareString)
	w.rankToNodes = newJavaHashMap[int, *dotLongSet](func(level int) int32 { return int32(level) }, dotCompareInt)
	w.colorGen = 1
	if opts.Strict {
		w.strict = make(map[uint64]struct{})
	}
	if opts.Strict || opts.StrictPrefix {
		_, _ = w.writer.WriteString("strict ")
	}
	_, _ = w.writer.WriteString("digraph DiskGraph {\n")
	_, _ = w.writer.WriteString("node [shape=box,style=rounded]\n")
	if opts.Colorize {
		_, _ = w.writer.WriteString("edge [colorscheme=\"paired12\"]\n")
	}
	_, _ = w.writer.WriteString("nodesep=0.35;\n")
	_, _ = w.writer.WriteString("subgraph cluster_graph {\n")
	_, _ = w.writer.WriteString("color=\"white\";\n")
	// DotStateWriter's constructor flushes the graph header immediately.
	w.writer.Flush()
	return w, nil
}

func (w *StateWriter) WriteInitState(state *TLCStateMut) error {
	w.require()
	if w.ownedNoOp() {
		return nil
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.writeInitState(state)
}

// Java's synchronized writeState methods call each other with the monitor held.
// Internal helpers retain that boundary without reacquiring Go's mutex.
func (w *StateWriter) writeInitState(state *TLCStateMut) error {
	if w != nil && w.WriteInitStateFunc != nil {
		if err := w.WriteInitStateFunc(state); err != nil {
			return err
		}
	}
	if w == nil || w.writer == nil || w.ownedNoOp() {
		return nil
	}
	if w.ownedDot() {
		fp := state.FingerPrint()
		_, err := fmt.Fprintf(w.writer, "%d [label=\"%s\",style = filled]\n", int64(fp), stateToDot(state.EvalStateLevelAlias(), nil, false))
		w.maintainRank(state)
		if err == nil && w.snapshot {
			err = w.snapshotAfterWrite()
		}
		return err
	}
	if state == nil {
		_, _ = fmt.Fprintf(w.writer, "State %d:\n", w.stateNum)
		panic(NewNullPointerException())
	}
	_, err := fmt.Fprintf(w.writer, "State %d:\n%s\n", w.stateNum, state.String())
	w.stateNum++
	return err
}

func (w *StateWriter) WriteTransition(curState *TLCStateMut, succState *TLCStateMut, status StateVisitStatus, action *Action, reason ...SemanticNode) error {
	w.require()
	if len(reason) > 1 {
		panic(NewIllegalArgumentException("WriteTransition accepts at most one predicate"))
	}
	return w.WriteTransitionVisual(curState, succState, status, action, StateVisualizationDefault, reason...)
}

func (w *StateWriter) WriteTransitionVisual(curState *TLCStateMut, succState *TLCStateMut, status StateVisitStatus, action *Action, visualization StateVisualization, reason ...SemanticNode) error {
	w.require()
	if len(reason) > 1 {
		panic(NewIllegalArgumentException("WriteTransitionVisual accepts at most one predicate"))
	}
	if w.ownedNoOp() {
		return nil
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if w != nil && w.WriteTransitionFunc != nil {
		if err := w.WriteTransitionFunc(curState, succState, status, action, reason...); err != nil {
			return err
		}
	}
	if w == nil || w.writer == nil || w.ownedNoOp() {
		return nil
	}
	if !w.ownedDot() {
		if !stateVisitStatusHas(status, StateVisitSeen) {
			return w.writeInitState(succState)
		}
		return nil
	}
	return w.writeDotTransition(curState, succState, status, visualization, action, reason...)
}

func (w *StateWriter) writeDotTransition(curState *TLCStateMut, succState *TLCStateMut, status StateVisitStatus, visualization StateVisualization, action *Action, reason ...SemanticNode) error {
	if !w.stuttering && visualization == StateVisualizationStuttering {
		return nil
	}
	sfp := succState.FingerPrint()
	cfp := curState.FingerPrint()
	if w.strict != nil {
		key := cfp ^ sfp
		if _, ok := w.strict[key]; ok {
			return nil
		}
		w.strict[key] = struct{}{}
	}
	// Java's Long.toString renders the 64-bit fingerprint as signed decimal.
	_, err := fmt.Fprintf(w.writer, "%d -> %d", int64(cfp), int64(sfp))
	if err != nil {
		return err
	}
	if visualization == StateVisualizationStuttering {
		_, err = w.writer.WriteString(" [style=\"dashed\",color=\"lightgray\"];\n")
	} else {
		if action != nil {
			_, _ = w.writer.WriteString(w.dotTransitionLabel(action, reason...))
		}
		_, _ = w.writer.WriteString(";\n")
		if !stateVisitStatusHas(status, StateVisitSeen) {
			style := ""
			if stateVisitStatusHas(status, StateVisitNotInModel) {
				style = ",style = filled, fillcolor=lightyellow"
			}
			predLabelState := curState.EvalStateLevelAlias()
			succLabelState := succState.EvalStateLevelAlias()
			_, err = fmt.Fprintf(w.writer, "%d [label=\"%s\",tooltip=\"%s\"%s];\n",
				int64(sfp),
				stateToDot(succLabelState, predLabelState, printDiffsOnly()),
				stateToDot(succState, nil, false),
				style,
			)
		}
	}
	w.maintainRank(curState)
	if err == nil && w.snapshot {
		err = w.snapshotAfterWrite()
	}
	return err
}

func (w *StateWriter) dotTransitionLabel(action *Action, reason ...SemanticNode) string {
	color := "black"
	if w.colorize {
		color = strconv.Itoa(w.getActionColor(action))
	}
	actionName := ""
	if w.actionLabels && action != nil {
		actionName = dotJavaTrim(strings.ReplaceAll(strings.ReplaceAll(action.GetInvocationSignature(), "\\", "\\\\"), "\"", "\\\""))
	}
	predText := ""
	if len(reason) > 0 && reason[0] != nil {
		predText = "\n" + fmt.Sprint(reason[0])
	}
	return fmt.Sprintf(" [label=\"%s%s\",color=\"%s\",fontcolor=\"%s\"]", actionName, predText, color, color)
}

func (w *StateWriter) getActionColor(action *Action) int {
	if action == nil {
		return 1
	}
	if w.actionToColors == nil {
		w.actionToColors = newJavaHashMap[string, int](javaFormatStringHash, dotCompareString)
	}
	name := action.requireNameString()
	if color, ok := w.actionToColors.Get2(name); ok {
		return color
	}
	w.colorGen++
	w.actionToColors.Set(name, w.colorGen)
	return w.colorGen
}

func (w *StateWriter) maintainRank(state *TLCStateMut) {
	if w.rankToNodes == nil || state == nil {
		return
	}
	level := state.Level()
	nodes := w.rankToNodes.getOrCreate(level, newDotLongSet)
	nodes.Set(state.FingerPrint(), struct{}{})
}

func (w *StateWriter) IsConstrained() bool {
	w.require()
	if w.ownedNoOp() {
		return false
	}
	return w.ownedConstrained()
}

func (w *StateWriter) IsNoop() bool {
	w.require()
	return w.ownedNoOp()
}

func (w *StateWriter) IsDot() bool {
	w.require()
	if w.ownedNoOp() {
		return false
	}
	return w.ownedDot()
}

func (w *StateWriter) GetDumpFileName() string {
	w.require()
	if w.ownedNoOp() {
		return ""
	}
	return w.fname
}

func (w *StateWriter) Snapshot() error {
	w.require()
	if w.ownedNoOp() || !w.ownedDot() {
		return nil
	}
	return w.snapshotFile()
}

// Automatic snapshots wrap checked copy/append failures after printing them;
// the public Snapshot method retains its checked I/O error.
func (w *StateWriter) snapshotAfterWrite() error {
	err := w.snapshotFile()
	if err != nil && isJavaIOException(err) {
		_, _ = fmt.Fprint(os.Stderr, javaThrowableStackTrace(err))
		return NewRuntimeExceptionFromCause(err)
	}
	return err
}

func (w *StateWriter) snapshotFile() error {
	if w == nil || w.writer == nil || w.fname == "" {
		return nil
	}
	w.writer.Flush()
	snapshot := dotWriterJavaReplace(w.fname, "_snapshot.dot")
	// FileUtil.copyFile replaces destination links and empty directories, while
	// preserving a same-file destination. Append only after the copy succeeds.
	if err := copyFile(w.fname, snapshot); err != nil {
		return err
	}
	trailer := w.dotClosingTrailer()
	snapshot, err := ioUtilsTXTPath(snapshot)
	if err != nil {
		return err
	}
	data, err := javaCharsetEncode(trailer, javaDefaultCharset(), false)
	if err != nil {
		return err
	}
	options, err := ioUtilsFileOptionsFromEnums([]string{"APPEND"})
	if err != nil {
		return err
	}
	return ioUtilsTXTWriteFile(snapshot, data, options)
}

func dotWriterJavaReplace(fname string, replacement string) string {
	return strings.ReplaceAll(fname, ".dot", replacement)
}

func (w *StateWriter) Close() error {
	w.require()
	if w.ownedNoOp() {
		return nil
	}
	if w.closed {
		return nil
	}
	w.closed = true
	if w != nil && w.CloseFunc != nil {
		if err := w.CloseFunc(); err != nil {
			return err
		}
	}
	if w.writer != nil {
		if w.ownedDot() {
			_, _ = w.writer.WriteString(w.dotClosingTrailer())
		}
		w.writer.Flush()
		w.writer.closed = true
	}
	if w.file != nil {
		// PrintWriter closes its output even after a failed flush and suppresses
		// the close error as well.
		_ = w.file.Close()
	}
	return nil
}

func (w *StateWriter) ownedNoOp() bool {
	if w.ownersCaptured {
		return w.noOpOwner
	}
	return w.Noop
}

func (w *StateWriter) ownedDot() bool {
	if w.ownersCaptured {
		return w.dotOwner
	}
	return w.Dot
}

func (w *StateWriter) ownedConstrained() bool {
	if w.ownersCaptured {
		return w.constrainedOwner
	}
	return w.Constrained
}

func (w *StateWriter) require() {
	if w == nil {
		panic(NewNullPointerException())
	}
}

func (w *StateWriter) dotClosingTrailer() string {
	var b strings.Builder
	for _, nodes := range w.rankToNodes.All() {
		b.WriteString("{rank = same; ")
		for fp := range nodes.All() {
			b.WriteString(strconv.FormatInt(int64(fp), 10))
			b.WriteByte(';')
		}
		b.WriteString("}\n")
	}
	b.WriteString("}\n")
	if w.colorize && w.actionToColors != nil && w.actionToColors.Len() > 1 {
		b.WriteString("subgraph cluster_legend {graph[style=bold];label = \"Next State Actions\" style=\"solid\"\n")
		b.WriteString("node [ labeljust=\"l\",colorscheme=\"paired12\",style=filled,shape=record ]\n")
		for action, color := range w.actionToColors.All() {
			name := strings.ReplaceAll(action, "!", ":")
			b.WriteString(fmt.Sprintf("%s [label=\"%s\",fillcolor=%d]\n", name, action, color))
		}
		b.WriteString("}")
	}
	b.WriteString("}")
	return b.String()
}

func stateToDot(state *TLCStateMut, pred *TLCStateMut, diffsOnly bool) string {
	if state == nil {
		return ""
	}
	if pred != nil && diffsOnly {
		return dotEscape(dotJavaTrim(state.StringForVariables(pred)))
	}
	return dotEscape(dotJavaTrim(state.String()))
}

func dotEscape(text string) string {
	text = strings.ReplaceAll(text, "\\", "\\\\")
	text = strings.ReplaceAll(text, "\"", "\\\"")
	text = strings.ReplaceAll(text, "\n", "\\n")
	return text
}

func dotJavaTrim(text string) string {
	return strings.TrimFunc(text, func(r rune) bool { return r <= ' ' })
}
