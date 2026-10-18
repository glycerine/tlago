package tlago

import (
	"reflect"
	"strconv"

	"github.com/glycerine/tlago/tlc"
)

// SanyParamAndPosition adapts semantic.ParamAndPosition over runtime symbols.
type SanyParamAndPosition struct {
	param    tlc.SemanticNode
	position int32
}

func NewSanyParamAndPosition(param tlc.SemanticNode, position int) *SanyParamAndPosition {
	return &SanyParamAndPosition{param: param, position: int32(position)}
}

func (p *SanyParamAndPosition) require() *SanyParamAndPosition {
	if p == nil {
		panic(tlc.NewNullPointerException())
	}
	return p
}

func sanyLevelParameterSymbolIsNil(symbol tlc.SemanticNode) bool {
	if symbol == nil {
		return true
	}
	value := reflect.ValueOf(symbol)
	return value.Kind() == reflect.Pointer && value.IsNil()
}

func sanyLevelParameterSymbolHash(symbol tlc.SemanticNode) int32 {
	if sanyLevelParameterSymbolIsNil(symbol) {
		panic(tlc.NewNullPointerException())
	}
	if symbol, ok := symbol.(*tlc.SymbolNode); ok {
		if symbol == nil || symbol.SemanticBase == nil {
			panic(tlc.NewNullPointerException())
		}
		return symbol.SemanticBase.JavaHashCode()
	}
	withHash, ok := symbol.(interface{ JavaHashCode() int32 })
	if !ok {
		panic(tlc.NewClassCastException("level parameter is not a SymbolNode"))
	}
	return withHash.JavaHashCode()
}

func sanyNullableSymbolString(symbol tlc.SemanticNode) string {
	if sanyLevelParameterSymbolIsNil(symbol) {
		return "null"
	}
	return tlc.SemanticNodeJavaString(symbol)
}

func (p *SanyParamAndPosition) HashCode() int32 {
	p.require()
	return sanyLevelParameterSymbolHash(p.param) + p.position
}

func (p *SanyParamAndPosition) Equals(other any) bool {
	p.require()
	that, ok := other.(*SanyParamAndPosition)
	return ok && that != nil && p.param == that.param && p.position == that.position
}

func (p *SanyParamAndPosition) String() string {
	p.require()
	return "<" + sanyNullableSymbolString(p.param) + ", " + strconv.FormatInt(int64(p.position), 10) + ">"
}

// SanyArgLevelParam adapts semantic.ArgLevelParam over runtime symbols.
type SanyArgLevelParam struct {
	op    tlc.SemanticNode
	index int32
	param tlc.SemanticNode
}

func NewSanyArgLevelParam(op tlc.SemanticNode, index int, param tlc.SemanticNode) *SanyArgLevelParam {
	return &SanyArgLevelParam{op: op, index: int32(index), param: param}
}

func (p *SanyArgLevelParam) require() *SanyArgLevelParam {
	if p == nil {
		panic(tlc.NewNullPointerException())
	}
	return p
}

func (p *SanyArgLevelParam) Occur(symbols []tlc.SemanticNode) bool {
	p.require()
	if symbols == nil {
		panic(tlc.NewNullPointerException())
	}
	for _, symbol := range symbols {
		if p.op == symbol || p.param == symbol {
			return true
		}
	}
	return false
}

func (p *SanyArgLevelParam) Equals(other any) bool {
	p.require()
	that, ok := other.(*SanyArgLevelParam)
	return ok && that != nil && p.op == that.op && p.index == that.index && p.param == that.param
}

func (p *SanyArgLevelParam) HashCode() int32 {
	p.require()
	return sanyLevelParameterSymbolHash(p.op) + p.index + sanyLevelParameterSymbolHash(p.param)
}

func (p *SanyArgLevelParam) String() string {
	p.require()
	return "<" + sanyNullableSymbolString(p.op) + ", " + strconv.FormatInt(int64(p.index), 10) + ", " + sanyNullableSymbolString(p.param) + ">"
}
