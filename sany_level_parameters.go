// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import (
	"reflect"
	"strconv"

	"github.com/glycerine/tlago/tlc"
)

// ParamAndPosition and ArgLevelParam retain actual SymbolNode references.
// Equality is reference identity, rather than symbol names or SemanticNode.equals.
type sanyParamAndPosition struct {
	param    sanySemSymbol
	position int32
}

func newSanyParamAndPosition(param sanySemSymbol, position int32) *sanyParamAndPosition {
	return &sanyParamAndPosition{param, position}
}

func (p *sanyParamAndPosition) hashCode() int32 {
	if p == nil {
		panic(tlc.NewNullPointerException())
	}
	return sanyLevelSymbolHash(p.param) + p.position
}

func (p *sanyParamAndPosition) equals(other any) bool {
	if p == nil {
		panic(tlc.NewNullPointerException())
	}
	q, ok := other.(*sanyParamAndPosition)
	return ok && q != nil && sanyLevelSymbolReference(p.param) == sanyLevelSymbolReference(q.param) && p.position == q.position
}

func (p *sanyParamAndPosition) String() string {
	if p == nil {
		panic(tlc.NewNullPointerException())
	}
	return "<" + sanyLevelSymbolString(p.param) + ", " + strconv.FormatInt(int64(p.position), 10) + ">"
}

type sanyArgLevelParam struct {
	op    sanySemSymbol
	i     int32
	param sanySemSymbol
}

func newSanyArgLevelParam(op sanySemSymbol, i int32, param sanySemSymbol) *sanyArgLevelParam {
	return &sanyArgLevelParam{op, i, param}
}

func (p *sanyArgLevelParam) occur(symbols []sanySemSymbol) bool {
	if p == nil || symbols == nil {
		panic(tlc.NewNullPointerException())
	}
	for _, symbol := range symbols {
		symbol = sanyLevelSymbolReference(symbol)
		if sanyLevelSymbolReference(p.op) == symbol || sanyLevelSymbolReference(p.param) == symbol {
			return true
		}
	}
	return false
}

func (p *sanyArgLevelParam) equals(other any) bool {
	if p == nil {
		panic(tlc.NewNullPointerException())
	}
	q, ok := other.(*sanyArgLevelParam)
	return ok && q != nil && sanyLevelSymbolReference(p.op) == sanyLevelSymbolReference(q.op) && p.i == q.i && sanyLevelSymbolReference(p.param) == sanyLevelSymbolReference(q.param)
}

func (p *sanyArgLevelParam) hashCode() int32 {
	if p == nil {
		panic(tlc.NewNullPointerException())
	}
	return sanyLevelSymbolHash(p.op) + p.i + sanyLevelSymbolHash(p.param)
}

func (p *sanyArgLevelParam) String() string {
	if p == nil {
		panic(tlc.NewNullPointerException())
	}
	return "<" + sanyLevelSymbolString(p.op) + ", " + strconv.FormatInt(int64(p.i), 10) + ", " + sanyLevelSymbolString(p.param) + ">"
}

func sanyLevelSymbolReference(symbol sanySemSymbol) sanySemSymbol {
	if symbol == nil || reflect.ValueOf(symbol).IsNil() {
		return nil
	}
	return symbol
}

func sanyLevelSymbolHash(symbol sanySemSymbol) int32 {
	symbol = sanyLevelSymbolReference(symbol)
	if symbol == nil {
		panic(tlc.NewNullPointerException())
	}
	return symbol.semBase().hashCode()
}

func sanyLevelSymbolString(symbol sanySemSymbol) string {
	symbol = sanyLevelSymbolReference(symbol)
	if symbol == nil {
		return "null"
	}
	tree := symbol.semBase().GetTreeNode()
	location := tlc.NullSourceLocation
	if syntax, ok := tree.(*SanySyntaxNode); ok && syntax != nil {
		bridge := tlcBridge{}
		location = bridge.sourceLocationForPosition(sanyNodePosition(syntax))
	} else if tree == tlc.NullSemanticNodeInstance.GetTreeNode() {
		location = tlc.NullSemanticNodeInstance.Location
	}
	// Do not use cached declaration locations: Java reads the current tree.
	return tlc.SemanticNodeJavaString(&tlc.SemanticNodeBase{TreeNode: tree, Location: location})
}
