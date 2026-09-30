package tlc

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type StateVisitStatus int

const (
	StateVisitUnseen StateVisitStatus = iota
	StateVisitSeen
	StateVisitNotInModel
)

type StateVisualization int

const (
	StateVisualizationStuttering StateVisualization = iota
	StateVisualizationDefault
	StateVisualizationDotted
)

type StateWriter struct {
	Noop                bool
	Dot                 bool
	Constrained         bool
	WriteInitStateFunc  func(*TLCStateMut) error
	WriteTransitionFunc func(*TLCStateMut, *TLCStateMut, StateVisitStatus, *Action, ...SemanticNode) error
	CloseFunc           func() error
	fname               string
	file                *os.File
	writer              *bufio.Writer
	stateNum            int
	colorize            bool
	actionLabels        bool
	snapshot            bool
	stuttering          bool
	strict              map[uint64]struct{}
	actionToColors      *InsMap[string, int]
	rankToNodes         map[int]map[uint64]struct{}
	colorGen            int
	closed              bool
}

func NewNoopStateWriter() *StateWriter {
	return &StateWriter{Noop: true}
}

func NewStateDumpWriter(fname string) (*StateWriter, error) {
	file, err := os.Create(fname)
	if err != nil {
		return nil, err
	}
	return &StateWriter{
		fname:    fname,
		file:     file,
		writer:   bufio.NewWriter(file),
		stateNum: 1,
	}, nil
}

type DotStateWriterOptions struct {
	Colorize     bool
	ActionLabels bool
	Snapshot     bool
	Constrained  bool
	Stuttering   bool
	Strict       bool
}

func NewDotStateWriter(fname string, opts DotStateWriterOptions) (*StateWriter, error) {
	w, err := NewStateDumpWriter(fname)
	if err != nil {
		return nil, err
	}
	w.Dot = true
	w.colorize = opts.Colorize
	w.actionLabels = opts.ActionLabels
	w.snapshot = opts.Snapshot
	w.Constrained = opts.Constrained
	w.stuttering = opts.Stuttering
	w.actionToColors = NewInsMap[string, int]()
	w.rankToNodes = make(map[int]map[uint64]struct{})
	w.colorGen = 1
	if opts.Strict {
		w.strict = make(map[uint64]struct{})
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
	return w, nil
}

func (w *StateWriter) WriteInitState(state *TLCStateMut) error {
	if w != nil && w.WriteInitStateFunc != nil {
		if err := w.WriteInitStateFunc(state); err != nil {
			return err
		}
	}
	if w == nil || w.writer == nil || w.Noop || state == nil {
		return nil
	}
	if w.Dot {
		fp := state.FingerPrint()
		_, err := fmt.Fprintf(w.writer, "%d [label=\"%s\",style = filled]\n", fp, stateToDot(state.EvalStateLevelAlias(), nil, false))
		w.maintainRank(state)
		if err == nil && w.snapshot {
			err = w.Snapshot()
		}
		return err
	}
	_, err := fmt.Fprintf(w.writer, "State %d:\n%s", w.stateNum, state.String())
	w.stateNum++
	return err
}

func (w *StateWriter) WriteTransition(curState *TLCStateMut, succState *TLCStateMut, status StateVisitStatus, action *Action, reason ...SemanticNode) error {
	return w.WriteTransitionVisual(curState, succState, status, action, StateVisualizationDefault, reason...)
}

func (w *StateWriter) WriteTransitionVisual(curState *TLCStateMut, succState *TLCStateMut, status StateVisitStatus, action *Action, visualization StateVisualization, reason ...SemanticNode) error {
	if w != nil && w.WriteTransitionFunc != nil {
		if err := w.WriteTransitionFunc(curState, succState, status, action, reason...); err != nil {
			return err
		}
	}
	if w == nil || w.writer == nil || w.Noop || succState == nil {
		return nil
	}
	if !w.Dot {
		if status != StateVisitSeen {
			return w.WriteInitState(succState)
		}
		return nil
	}
	return w.writeDotTransition(curState, succState, status, visualization, action, reason...)
}

func (w *StateWriter) writeDotTransition(curState *TLCStateMut, succState *TLCStateMut, status StateVisitStatus, visualization StateVisualization, action *Action, reason ...SemanticNode) error {
	if curState == nil {
		return w.WriteInitState(succState)
	}
	if !w.stuttering && visualization == StateVisualizationStuttering {
		return nil
	}
	cfp := curState.FingerPrint()
	sfp := succState.FingerPrint()
	if w.strict != nil {
		key := cfp ^ sfp
		if _, ok := w.strict[key]; ok {
			return nil
		}
		w.strict[key] = struct{}{}
	}
	_, err := fmt.Fprintf(w.writer, "%d -> %d", cfp, sfp)
	if err != nil {
		return err
	}
	if visualization == StateVisualizationStuttering {
		_, err = w.writer.WriteString(" [style=\"dashed\",color=\"lightgray\"];\n")
		return err
	}
	if label := w.dotTransitionLabel(action, reason...); label != "" {
		_, _ = w.writer.WriteString(label)
	}
	_, _ = w.writer.WriteString(";\n")
	if status != StateVisitSeen {
		style := ""
		if status == StateVisitNotInModel {
			style = ",style = filled, fillcolor=lightyellow"
		}
		predLabelState := curState.EvalStateLevelAlias()
		succLabelState := succState.EvalStateLevelAlias()
		_, err = fmt.Fprintf(w.writer, "%d [label=\"%s\",tooltip=\"%s\"%s];\n",
			sfp,
			stateToDot(succLabelState, predLabelState, printDiffsOnly()),
			stateToDot(succState, nil, false),
			style,
		)
	}
	w.maintainRank(curState)
	if err == nil && w.snapshot {
		err = w.Snapshot()
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
		actionName = dotEscape(action.GetName())
	}
	if len(reason) > 0 && reason[0] != nil {
		actionName += "\\n" + dotEscape(fmt.Sprint(reason[0]))
	}
	if actionName == "" && color == "black" {
		return ""
	}
	return fmt.Sprintf(" [label=\"%s\",color=\"%s\",fontcolor=\"%s\"]", actionName, color, color)
}

func (w *StateWriter) getActionColor(action *Action) int {
	if action == nil {
		return 1
	}
	if w.actionToColors == nil {
		w.actionToColors = NewInsMap[string, int]()
	}
	name := action.GetName()
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
	nodes := w.rankToNodes[level]
	if nodes == nil {
		nodes = make(map[uint64]struct{})
		w.rankToNodes[level] = nodes
	}
	nodes[state.FingerPrint()] = struct{}{}
}

func (w *StateWriter) IsConstrained() bool {
	return w != nil && w.Constrained
}

func (w *StateWriter) IsNoop() bool {
	return w == nil || w.Noop
}

func (w *StateWriter) IsDot() bool {
	return w != nil && w.Dot
}

func (w *StateWriter) GetDumpFileName() string {
	if w == nil {
		return ""
	}
	return w.fname
}

func (w *StateWriter) Snapshot() error {
	if w == nil || w.writer == nil || w.fname == "" {
		return nil
	}
	if err := w.writer.Flush(); err != nil {
		return err
	}
	data, err := os.ReadFile(w.fname)
	if err != nil {
		return err
	}
	snapshot := strings.TrimSuffix(w.fname, ".dot") + "_snapshot.dot"
	data = append(data, []byte(w.dotClosingTrailer())...)
	return os.WriteFile(snapshot, data, 0o644)
}

func (w *StateWriter) Close() error {
	if w == nil || w.closed {
		return nil
	}
	w.closed = true
	if w != nil && w.CloseFunc != nil {
		if err := w.CloseFunc(); err != nil {
			return err
		}
	}
	if w.writer != nil {
		if w.Dot {
			_, _ = w.writer.WriteString(w.dotClosingTrailer())
		}
		if err := w.writer.Flush(); err != nil {
			return err
		}
	}
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

func (w *StateWriter) dotClosingTrailer() string {
	var b strings.Builder
	levels := make([]int, 0, len(w.rankToNodes))
	for level := range w.rankToNodes {
		levels = append(levels, level)
	}
	sort.Ints(levels)
	for _, level := range levels {
		nodes := make([]uint64, 0, len(w.rankToNodes[level]))
		for fp := range w.rankToNodes[level] {
			nodes = append(nodes, fp)
		}
		sort.Slice(nodes, func(i, j int) bool { return nodes[i] < nodes[j] })
		b.WriteString("{rank = same; ")
		for _, fp := range nodes {
			b.WriteString(strconv.FormatUint(fp, 10))
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
			b.WriteString(fmt.Sprintf("%s [label=\"%s\",fillcolor=%d]\n", dotID(name), dotEscape(action), color))
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
		return dotEscape(strings.TrimSpace(state.StringForVariables(pred)))
	}
	return dotEscape(strings.TrimSpace(state.String()))
}

func dotEscape(text string) string {
	text = strings.ReplaceAll(text, "\\", "\\\\")
	text = strings.ReplaceAll(text, "\"", "\\\"")
	text = strings.ReplaceAll(text, "\n", "\\n")
	return text
}

func dotID(text string) string {
	var b strings.Builder
	for _, r := range text {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	if b.Len() == 0 {
		return "action"
	}
	return b.String()
}
