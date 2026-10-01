package tlc

import (
	"fmt"
	"strings"
)

func GraphVizDotDiGraph(graph Value, vertexLabel Value, edgeLabel Value) (Value, error) {
	record := asRecordValue(graph)
	if record == nil {
		return nil, newTLCError(ECGeneral, "An DiGraph must be a record. Value given is of type: %T", graph)
	}

	var b strings.Builder
	b.WriteString("digraph MyGraph {")

	nodesValue, err := record.Apply(NewStringValue("node"))
	if err != nil {
		return nil, err
	}
	nodes, err := toSetEnumValue(nodesValue)
	if err != nil {
		return nil, err
	}
	nodes.Normalize()
	nodeEnum := nodes.Elements()
	for {
		elem := nodeEnum.NextElement()
		if elem == nil {
			if err := nodeEnum.Err(); err != nil {
				return nil, err
			}
			break
		}
		label, err := EvalOperatorValue(vertexLabel, []Value{elem}, EvalClear)
		if err != nil {
			return nil, err
		}
		b.WriteString(fmt.Sprintf("%d[label=%s];", int64(elem.FingerPrint(FP64New())), label))
	}

	edgesValue, err := record.Apply(NewStringValue("edge"))
	if err != nil {
		return nil, err
	}
	edges, err := toSetEnumValue(edgesValue)
	if err != nil {
		return nil, err
	}
	edges.Normalize()
	edgeEnum := edges.Elements()
	for {
		elem := edgeEnum.NextElement()
		if elem == nil {
			if err := edgeEnum.Err(); err != nil {
				return nil, err
			}
			break
		}
		tuple := asTupleValue(elem)
		if tuple == nil || len(tuple.Elems) < 2 {
			return nil, newTLCError(ECGeneral, "GraphViz edge must be a tuple with at least two elements: %s", elem)
		}
		label, err := EvalOperatorValue(edgeLabel, []Value{tuple}, EvalClear)
		if err != nil {
			return nil, err
		}
		b.WriteString(fmt.Sprintf("%d->%d[label=%s];",
			int64(tuple.Elems[0].FingerPrint(FP64New())),
			int64(tuple.Elems[1].FingerPrint(FP64New())),
			label))
	}

	b.WriteString("}")
	return NewStringValue(b.String()), nil
}
