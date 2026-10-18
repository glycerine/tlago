package tlc

import (
	"fmt"
	"strings"
)

func GraphVizDotDiGraph(graph Value, vertexLabel Value, edgeLabel Value) (Value, error) {
	record, ok := asConcreteRecordValue(graph)
	if !ok {
		return nil, newTLCError(ECGeneral, "An DiGraph must be a record. Value given is of type: %s", javaValueClassName(graph))
	}

	var b strings.Builder
	b.WriteString("digraph MyGraph {")

	nodesValue, err := record.Apply(NewStringValue("node"))
	if err != nil {
		return nil, err
	}
	nodes, err := tryToSetEnumValue(nodesValue)
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
	edges, err := tryToSetEnumValue(edgesValue)
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
		tupleElems := tuple.GetElems()
		fromFingerprint := int64(valueStreamArrayAt(tupleElems, 0).FingerPrint(FP64New()))
		toFingerprint := int64(valueStreamArrayAt(tupleElems, 1).FingerPrint(FP64New()))
		label, err := EvalOperatorValue(edgeLabel, []Value{tuple}, EvalClear)
		if err != nil {
			return nil, err
		}
		b.WriteString(fmt.Sprintf("%d->%d[label=%s];",
			fromFingerprint,
			toFingerprint,
			label))
	}

	b.WriteString("}")
	return NewStringValue(b.String()), nil
}
