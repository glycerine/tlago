package tlc

import (
	"fmt"
	"os"
	"sync"
)

type DotActionWriter struct {
	mu     sync.Mutex
	FName  string
	file   *os.File
	writer *statePrintWriter
	closed bool
}

func NewDotActionWriter(fname string, strict string) (*DotActionWriter, error) {
	file, err := openPoolFile(fname, true)
	if err != nil {
		ToolIOPrintln("Error: Unable to write to file " + fname)
		return nil, err
	}
	w := &DotActionWriter{FName: fname, file: file, writer: &statePrintWriter{buffer: NewJavaBufferedWriter(file)}}
	_, _ = w.writer.WriteString(strict + "digraph ActionGraph {\n")
	_, _ = w.writer.WriteString("nodesep=0.35;\n")
	_, _ = w.writer.WriteString("subgraph cluster_legend {\n")
	_, _ = w.writer.WriteString("label = \"Coverage\";\n")
	_, _ = w.writer.WriteString("node [shape=point] {\n")
	_, _ = w.writer.WriteString("d0 [style = invis];\n")
	_, _ = w.writer.WriteString("d1 [style = invis];\n")
	_, _ = w.writer.WriteString("p0 [style = invis];\n")
	_, _ = w.writer.WriteString("p0 [style = invis];\n")
	_, _ = w.writer.WriteString("}\n")
	_, _ = w.writer.WriteString("d0 -> d1 [label=unseen, color=\"green\", style=dotted]\n")
	_, _ = w.writer.WriteString("p0 -> p1 [label=seen]\n")
	_, _ = w.writer.WriteString("}\n")
	w.writer.Flush()
	return w, nil
}

func (w *DotActionWriter) WriteAction(action *Action, id int) error {
	if w == nil || w.writer == nil {
		panic(NewNullPointerException())
	}
	id = int(int32(id))
	w.mu.Lock()
	defer w.mu.Unlock()
	if _, err := fmt.Fprintf(w.writer, "%d [shape=box,label=\"%s", id, action.GetNameOfDefault()); err != nil {
		return err
	}
	if action.IsInitPredicate() {
		_, err := w.writer.WriteString("\",style = filled]\n")
		return err
	}
	_, err := w.writer.WriteString("\"]\n")
	return err
}

func (w *DotActionWriter) WriteEdge(fromID int, toID int, weight ...float64) error {
	if w == nil || w.writer == nil {
		panic(NewNullPointerException())
	}
	if len(weight) > 1 {
		panic(NewIllegalArgumentException("WriteEdge accepts at most one weight"))
	}
	fromID = int(int32(fromID))
	toID = int(int32(toID))
	w.mu.Lock()
	defer w.mu.Unlock()
	edgeWeight := 0.0
	if len(weight) > 0 {
		edgeWeight = weight[0]
	}
	if _, err := fmt.Fprintf(w.writer, "%d -> %d", fromID, toID); err != nil {
		return err
	}
	if edgeWeight == 0 {
		_, _ = w.writer.WriteString("[color=\"green\",style=dotted]")
	} else {
		_, _ = fmt.Fprintf(w.writer, "[penwidth=%s]", javaDoubleString(edgeWeight))
	}
	_, err := w.writer.WriteString(";\n")
	return err
}

func (w *DotActionWriter) WriteSubGraphStart(key string, label string) error {
	if w == nil || w.writer == nil {
		panic(NewNullPointerException())
	}
	_, err := fmt.Fprintf(w.writer, "subgraph cluster_%s {\ncolor=\"white\"\nlabel=\"%s\"\n", key, label)
	return err
}

func (w *DotActionWriter) WriteSubGraphEnd() error {
	if w == nil || w.writer == nil {
		panic(NewNullPointerException())
	}
	_, err := w.writer.WriteString("}\n")
	return err
}

func (w *DotActionWriter) Close() error {
	if w == nil || w.writer == nil {
		panic(NewNullPointerException())
	}
	if w.closed {
		return nil
	}
	_, _ = w.writer.WriteString("}")
	w.writer.Flush()
	w.writer.closed = true
	w.closed = true
	_ = w.file.Close()
	return nil
}
