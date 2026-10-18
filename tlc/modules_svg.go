package tlc

import (
	"fmt"
	"math"
	"strings"
)

func SVGElemToString(elem Value) (Value, error) {
	record, ok := asConcreteRecordValue(elem)
	if !ok {
		return nil, newTLCError(ECGeneral, "An SVG element must be a record. Value given is of type: %s", javaValueClassName(elem))
	}

	nameValue, err := record.Apply(NewStringValue("name"))
	if err != nil {
		return nil, err
	}
	var name *StringValue
	if !isNil(nameValue) {
		name, ok = asStringValue(nameValue)
		if !ok {
			return nil, valueStreamClassCast(nameValue, "tlc2.value.impl.StringValue")
		}
	}

	attrsValue, err := record.Apply(NewStringValue("attrs"))
	if err != nil {
		return nil, err
	}
	attrs := asRecordValue(attrsValue)
	if attrs == nil {
		return nil, newTLCError(ECGeneral, "Was unable to convert element to a record: %s", attrsValue)
	}

	var attrBuilder strings.Builder
	for _, attrName := range attrs.ownedNames() {
		attrBuilder.WriteByte(' ')
		attrBuilder.WriteString(strings.ReplaceAll(attrName.String(), "_", "-"))
		attrBuilder.WriteByte('=')
		attrValue, err := attrs.Apply(NewStringValueFromUnique(attrName))
		if err != nil {
			return nil, err
		}
		var attrString *StringValue
		if !isNil(attrValue) {
			attrString, ok = asStringValue(attrValue)
			if !ok {
				return nil, valueStreamClassCast(attrValue, "tlc2.value.impl.StringValue")
			}
		}
		attrBuilder.WriteByte('\'')
		attrBuilder.WriteString(attrString.RawString())
		attrBuilder.WriteByte('\'')
	}

	childrenValue, err := record.Apply(NewStringValue("children"))
	if err != nil {
		return nil, err
	}
	children, err := svgChildren(childrenValue)
	if err != nil {
		return nil, err
	}
	var childBuilder strings.Builder
	for _, child := range children {
		childValue, err := SVGElemToString(child)
		if err != nil {
			return nil, err
		}
		childString, ok := childValue.(*StringValue)
		if !ok {
			return nil, newTLCError(ECGeneral, "SVGElemToString returned non-string %s", childValue)
		}
		childBuilder.WriteString(childString.RawString())
	}

	innerTextValue, err := record.Apply(NewStringValue("innerText"))
	if err != nil {
		return nil, err
	}
	var innerText *StringValue
	if !isNil(innerTextValue) {
		innerText, ok = asStringValue(innerTextValue)
		if !ok {
			return nil, valueStreamClassCast(innerTextValue, "tlc2.value.impl.StringValue")
		}
	}
	text := strings.ReplaceAll(innerText.RawString(), "<<", "&lt;&lt;")
	text = strings.ReplaceAll(text, ">>", "&gt;&gt;")

	svg := fmt.Sprintf("<%s%s>", name.RawString(), attrBuilder.String())
	svg += text
	svg += childBuilder.String()
	svg += fmt.Sprintf("</%s>", name.RawString())
	return NewStringValue(svg), nil
}

func svgChildren(value Value) ([]Value, error) {
	switch v := value.(type) {
	case *TupleValue:
		return v.GetElems(), nil
	case *RecordValue:
		return v.ownedValues(), nil
	case *CounterExample:
		return asRecordValue(v).ownedValues(), nil
	case *FcnRcdValue:
		return v.ownedValues(), nil
	case *FcnLambdaValue:
		fcn := v.ToFcnRcd()
		if fcn != nil {
			return fcn.ownedValues(), nil
		}
	}
	return nil, newTLCError(ECGeneral, "Was unable to convert element to a tuple or (function) record: %s", value)
}

func SVGNodeOfRingNetwork(cx Value, cy Value, r Value, n Value, m Value) (Value, error) {
	cxInt, err := svgIntArg("first", "NodeOfRingNetwork", cx)
	if err != nil {
		return nil, err
	}
	cyInt, err := svgIntArg("second", "NodeOfRingNetwork", cy)
	if err != nil {
		return nil, err
	}
	rInt, err := svgIntArg("third", "NodeOfRingNetwork", r)
	if err != nil {
		return nil, err
	}
	nInt, err := svgIntArg("fourth", "NodeOfRingNetwork", n)
	if err != nil {
		return nil, err
	}
	mInt, err := svgIntArg("fifth", "NodeOfRingNetwork", m)
	if err != nil {
		return nil, err
	}

	angle := (2.0 * math.Pi / float64(mInt.ownedVal())) * float64(nInt.ownedVal())
	x := javaDoubleToInt(float64(cxInt.ownedVal()) + float64(rInt.ownedVal())*math.Cos(angle))
	y := javaDoubleToInt(float64(cyInt.ownedVal()) + float64(rInt.ownedVal())*math.Sin(angle))
	return NewRecordValue(
		[]*UniqueString{UniqueStringOf("x"), UniqueStringOf("y")},
		[]Value{NewIntValue(x), NewIntValue(y)},
		false,
	), nil
}

func SVGNodesOfDirectedMultiGraph(n Value, e Value, o Value) (Value, error) {
	nodes, err := tryToSetEnumValue(n)
	if err != nil {
		return nil, err
	}
	if nodes == nil {
		return nil, newTLCError(ECGeneral, "Nodes must be a set. Value given is of type: %s", javaValueClassName(n))
	}
	edges, err := tryToSetEnumValue(e)
	if err != nil {
		return nil, err
	}
	if edges == nil {
		return nil, newTLCError(ECGeneral, "Edges must be a set. Value given is of type: %s", javaValueClassName(e))
	}
	opts := asRecordValue(o)
	if opts == nil {
		return nil, newTLCError(ECGeneral, "Opts must be a record. Value given is of type: %s", javaValueClassName(o))
	}

	viewHeightValue, err := svgRecordInt(opts, "view_width")
	if err != nil {
		return nil, err
	}
	viewWidthValue, err := svgRecordInt(opts, "view_height")
	if err != nil {
		return nil, err
	}
	nodeHeightValue, err := svgRecordInt(opts, "node_width")
	if err != nil {
		return nil, err
	}
	nodeWidthValue, err := svgRecordInt(opts, "node_height")
	if err != nil {
		return nil, err
	}
	algoValue, err := svgRecordString(opts, "algo")
	if err != nil {
		return nil, err
	}
	if algoValue.RawString() == "Sugiyama" || algoValue.RawString() == "Eiglsperger" {
		if _, err := svgRecordString(opts, "layering"); err != nil {
			return nil, err
		}
	}

	domain := nodes.Elems.ToArray()
	edgePairs, err := svgDirectedEdgePairs(domain, edges)
	if err != nil {
		return nil, err
	}
	coords := svgLayeredDirectedGraphLayout(
		len(domain),
		edgePairs,
		int(viewWidthValue.ownedVal()),
		int(viewHeightValue.ownedVal()),
		int(nodeWidthValue.ownedVal()),
		int(nodeHeightValue.ownedVal()),
	)
	values := make([]Value, len(domain))
	for i, point := range coords {
		values[i] = svgPointRecord(point.x, point.y)
	}
	return NewFcnRcdValue(domain, values, false), nil
}

func SVGPointOnLine(from Value, to Value, idx Value) (Value, error) {
	fromRecord := asRecordValue(from)
	if fromRecord == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "PointOnLine", "record", ValuesPPR(from))
	}
	toRecord := asRecordValue(to)
	if toRecord == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "PointOnLine", "record", ValuesPPR(to))
	}
	segment, err := svgIntArg("third", "PointOnLine", idx)
	if err != nil {
		return nil, err
	}
	fx, err := svgRecordInt(fromRecord, "x")
	if err != nil {
		return nil, err
	}
	fy, err := svgRecordInt(fromRecord, "y")
	if err != nil {
		return nil, err
	}
	tx, err := svgRecordInt(toRecord, "x")
	if err != nil {
		return nil, err
	}
	ty, err := svgRecordInt(toRecord, "y")
	if err != nil {
		return nil, err
	}

	x := javaDoubleToInt(float64(fx.ownedVal()) + (float64(tx.ownedVal()-fx.ownedVal()) / float64(segment.ownedVal())))
	y := javaDoubleToInt(float64(fy.ownedVal()) + (float64(ty.ownedVal()-fy.ownedVal()) / float64(segment.ownedVal())))
	return NewRecordValue(
		[]*UniqueString{UniqueStringOf("x"), UniqueStringOf("y")},
		[]Value{NewIntValue(x), NewIntValue(y)},
		false,
	), nil
}

type svgDirectedEdge struct {
	from int
	to   int
}

type svgPoint struct {
	x int32
	y int32
}

func svgDirectedEdgePairs(nodes []Value, edges *SetEnumValue) ([]svgDirectedEdge, error) {
	out := make([]svgDirectedEdge, 0, edges.Elems.Len())
	for i := 0; i < edges.Elems.Len(); i++ {
		edge := edges.Elems.At(i)
		var tuple *TupleValue
		if !isNil(edge) {
			var ok bool
			tuple, ok = edge.(*TupleValue)
			if !ok {
				return nil, valueStreamClassCast(edge, "tlc2.value.impl.TupleValue")
			}
		}
		tupleElems := tuple.GetElems()
		from, err := svgNodeIndex(nodes, valueStreamArrayAt(tupleElems, 0))
		if err != nil {
			return nil, err
		}
		to, err := svgNodeIndex(nodes, valueStreamArrayAt(tupleElems, 1))
		if err != nil {
			return nil, err
		}
		if from != to {
			out = append(out, svgDirectedEdge{from: from, to: to})
		}
	}
	return out, nil
}

func svgNodeIndex(nodes []Value, target Value) (int, error) {
	for i, node := range nodes {
		eq, err := node.Equal(target)
		if err != nil {
			return 0, err
		}
		if eq {
			return i, nil
		}
	}
	return 0, newTLCError(ECGeneral, "Edge references a node that is not in the node set: %s", target)
}

func svgLayeredDirectedGraphLayout(nodeCount int, edges []svgDirectedEdge, width int, height int, nodeWidth int, nodeHeight int) []svgPoint {
	points := make([]svgPoint, nodeCount)
	if nodeCount == 0 {
		return points
	}
	layers := make([]int, nodeCount)
	for pass := 0; pass < nodeCount; pass++ {
		changed := false
		for _, edge := range edges {
			nextLayer := layers[edge.from] + 1
			if nextLayer > layers[edge.to] {
				layers[edge.to] = nextLayer
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	maxLayer := 0
	for _, layer := range layers {
		if layer > maxLayer {
			maxLayer = layer
		}
	}
	layerCounts := make([]int, maxLayer+1)
	for _, layer := range layers {
		layerCounts[layer]++
	}
	layerSeen := make([]int, maxLayer+1)
	xMargin := max(nodeWidth/2, 0)
	yMargin := max(nodeHeight/2, 0)
	usableWidth := max(width-2*xMargin, 0)
	usableHeight := max(height-2*yMargin, 0)
	for i, layer := range layers {
		posInLayer := layerSeen[layer]
		layerSeen[layer]++
		x := xMargin
		if layerCounts[layer] > 1 {
			x += (usableWidth * posInLayer) / (layerCounts[layer] - 1)
		} else {
			x += usableWidth / 2
		}
		y := yMargin
		if maxLayer > 0 {
			y += (usableHeight * layer) / maxLayer
		} else {
			y += usableHeight / 2
		}
		points[i] = svgPoint{x: int32(x), y: int32(y)}
	}
	return points
}

func svgPointRecord(x int32, y int32) Value {
	return NewRecordValue(
		[]*UniqueString{UniqueStringOf("x"), UniqueStringOf("y")},
		[]Value{NewIntValue(x), NewIntValue(y)},
		false,
	)
}

func svgRecordInt(record *RecordValue, field string) (*IntValue, error) {
	value, err := record.Select(NewStringValue(field))
	if err != nil {
		return nil, err
	}
	if isNil(value) {
		return nil, nil
	}
	intValue, ok := value.(*IntValue)
	if !ok {
		return nil, valueStreamClassCast(value, "tlc2.value.impl.IntValue")
	}
	return intValue, nil
}

func svgRecordString(record *RecordValue, field string) (*StringValue, error) {
	value, err := record.Select(NewStringValue(field))
	if err != nil {
		return nil, err
	}
	if isNil(value) {
		return nil, nil
	}
	stringValue, ok := asStringValue(value)
	if !ok {
		return nil, valueStreamClassCast(value, "tlc2.value.impl.StringValue")
	}
	return stringValue, nil
}

func svgIntArg(position string, operator string, value Value) (*IntValue, error) {
	intValue, ok := value.(*IntValue)
	if !ok {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, position, operator, "integer", ValuesPPR(value))
	}
	return intValue, nil
}
