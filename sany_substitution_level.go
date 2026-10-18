// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import (
	"reflect"

	"github.com/glycerine/tlago/tlc"
)

func sanySubstParamSet(param sanySemSymbol, subs []*sanySemSubst) *sanyLevelSet[sanySemSymbol] {
	return sanySubstParameterSet(param, subs, false)
}
func sanySubstAllParamSet(param sanySemSymbol, subs []*sanySemSubst) *sanyLevelSet[sanySemSymbol] {
	return sanySubstParameterSet(param, subs, true)
}
func sanySubstParameterSet(param sanySemSymbol, subs []*sanySemSubst, all bool) *sanyLevelSet[sanySemSymbol] {
	if subs == nil {
		panic(tlc.NewNullPointerException())
	}
	param = sanyLevelSymbolReference(param)
	for _, sub := range subs {
		op := sub.getOp()
		if (op == nil && param == nil) || (op != nil && op == param) {
			replacement := sanyRequireCanonicalLevelNode(sub.getExpr())
			if all {
				return replacement.getAllParams()
			}
			return replacement.getLevelParams()
		}
	}
	result := newSanyLevelSymbolSet()
	result.add(param)
	return result
}
func sanySubstAsOpArg(node sanySemanticGraphNode) *sanySemOpArgNode {
	if !sanyGraphNodePresent(node) {
		return nil
	}
	result, ok := node.(*sanySemOpArgNode)
	if !ok {
		panic(tlc.NewClassCastException("not an OpArgNode"))
	}
	return result
}
func sanySubstOperator(arg *sanySemOpArgNode) sanySemSymbol {
	if arg == nil {
		panic(tlc.NewNullPointerException())
	}
	return sanyLevelSymbolReference(arg.operator)
}
func sanySubstOperatorIsParam(op sanySemSymbol) bool {
	op = sanyLevelSymbolReference(op)
	if op == nil {
		panic(tlc.NewNullPointerException())
	}
	return sanySymbolIsParam(op)
}
func sanySubstGetSubLCSet(body sanyCanonicalLevelNode, subs []*sanySemSubst, constant bool, iter int32, errors *Diagnostics) *sanySetOfLevelConstraints {
	result := newSanySetOfLevelConstraints()
	if body == nil || reflect.ValueOf(body).IsNil() {
		panic(tlc.NewNullPointerException())
	}
	lc := body.getLevelConstraints()
	if lc == nil {
		panic(tlc.NewNullPointerException())
	}
	for param := range lc.entries.All() {
		level := lc.get(param)
		if !constant {
			symbol := sanyLevelSymbolReference(param)
			if symbol == nil {
				panic(tlc.NewNullPointerException())
			}
			if symbol.semKind() == sanyConstantDeclKind {
				v := int32(0)
				level = &v
			} else if symbol.semKind() == sanyVariableDeclKind {
				v := int32(1)
				level = &v
			}
		}
		for p := range sanySubstParamSet(param, subs).all() {
			result.put(p, level)
		}
	}
	for alp := range body.getArgLevelParams().all() {
		if alp == nil {
			panic(tlc.NewNullPointerException())
		}
		sub := sanySubstAsOpArg(sanySubstGetSub(sanyLevelSymbolReference(alp.op), subs))
		if sub != nil {
			if definition, ok := sanySubstOperator(sub).(*sanySemOpDefNode); ok && definition != nil {
				definition.levelCheck(iter, errors)
				level := int32(definition.getMaxLevel(int(alp.i)))
				for p := range sanySubstParamSet(alp.param, subs).all() {
					result.put(p, &level)
				}
			}
		}
	}
	return result
}
func sanySubstGetSubALCSet(body sanyCanonicalLevelNode, subs []*sanySemSubst, iter int32, errors *Diagnostics) *sanySetOfArgLevelConstraints {
	result := newSanySetOfArgLevelConstraints()
	if body == nil || reflect.ValueOf(body).IsNil() {
		panic(tlc.NewNullPointerException())
	}
	alc := body.getArgLevelConstraints()
	if alc == nil {
		panic(tlc.NewNullPointerException())
	}
	for pap := range alc.entries.All() {
		level := alc.get(pap)
		if pap == nil {
			panic(tlc.NewNullPointerException())
		}
		sub := sanySubstGetSub(sanyLevelSymbolReference(pap.param), subs)
		if !sanyGraphNodePresent(sub) {
			result.put(pap, level)
		} else {
			op := sanySubstOperator(sanySubstAsOpArg(sub))
			if sanySubstOperatorIsParam(op) {
				result.put(newSanyParamAndPosition(op, pap.position), level)
			}
		}
	}
	for alp := range body.getArgLevelParams().all() {
		if alp == nil {
			panic(tlc.NewNullPointerException())
		}
		param := sanySubstGetSub(sanyLevelSymbolReference(alp.param), subs)
		if sanyGraphNodePresent(param) {
			sub := sanySubstGetSub(sanyLevelSymbolReference(alp.op), subs)
			op := sanyLevelSymbolReference(alp.op)
			if sanyGraphNodePresent(sub) {
				op = sanySubstOperator(sanySubstAsOpArg(sub))
			}
			if sanySubstOperatorIsParam(op) {
				pap := newSanyParamAndPosition(op, alp.i)
				replacement := sanyRequireCanonicalLevelNode(param)
				replacement.levelCheck(iter, errors)
				level := int32(replacement.getLevel())
				result.put(pap, &level)
			}
		}
	}
	return result
}
func sanySubstGetSubALPSet(body sanyCanonicalLevelNode, subs []*sanySemSubst) *sanyLevelSet[*sanyArgLevelParam] {
	result := newSanyArgLevelSet()
	if body == nil || reflect.ValueOf(body).IsNil() {
		panic(tlc.NewNullPointerException())
	}
	for alp := range body.getArgLevelParams().all() {
		if alp == nil {
			panic(tlc.NewNullPointerException())
		}
		sub := sanySubstGetSub(sanyLevelSymbolReference(alp.op), subs)
		if !sanyGraphNodePresent(sub) {
			result.add(alp)
		} else {
			op := sanySubstOperator(sanySubstAsOpArg(sub))
			if sanySubstOperatorIsParam(op) {
				for param := range sanySubstParamSet(alp.param, subs).all() {
					result.add(newSanyArgLevelParam(op, alp.i, param))
				}
			}
		}
	}
	return result
}
