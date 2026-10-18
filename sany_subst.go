package tlago

import "github.com/glycerine/tlago/tlc"

// SanySubst is the shared runtime representation of semantic.Subst.
type SanySubst = tlc.Subst

// NewSanySubst retains all four source constructor fields without allocating a
// semantic-node UID; Subst is an ordinary mutable graph carrier.
func NewSanySubst(op *tlc.SymbolNode, expr tlc.SemanticNode, syntax any, implicit bool) *SanySubst {
	substitution := tlc.NewSubst(op, expr)
	substitution.SetExprSTN(syntax)
	substitution.SetExpr(expr, implicit)
	return &substitution
}

// SanySubstGetSub returns the first replacement whose declaration is identical
// to param. A null declaration therefore matches a null param.
func SanySubstGetSub(param any, substitutions []*SanySubst) tlc.SemanticNode {
	if substitutions == nil {
		panic(tlc.NewNullPointerException())
	}
	parameter, parameterIsSymbol := param.(*tlc.SymbolNode)
	parameterIsNull := param == nil || parameterIsSymbol && parameter == nil
	for _, substitution := range substitutions {
		op := substitution.GetOp()
		if op == nil && parameterIsNull || parameterIsSymbol && op == parameter {
			return substitution.GetExpr()
		}
	}
	return nil
}

func sanyPublicSubstParameterSet(param *tlc.SymbolNode, substitutions []*SanySubst, all bool) []*tlc.SymbolNode {
	if substitutions == nil {
		panic(tlc.NewNullPointerException())
	}
	for _, substitution := range substitutions {
		if substitution.GetOp() != param {
			continue
		}
		expression := substitution.GetExpr()
		if sanyExploreNull(expression) {
			panic(tlc.NewNullPointerException())
		}
		if all {
			return tlc.SemanticAllParams(expression)
		}
		return tlc.SemanticLevelParams(expression)
	}
	return []*tlc.SymbolNode{param}
}

// SanySubstParamSet returns the replacement's level parameters or a fresh
// singleton containing the original declaration when it is not substituted.
func SanySubstParamSet(param *tlc.SymbolNode, substitutions []*SanySubst) []*tlc.SymbolNode {
	return sanyPublicSubstParameterSet(param, substitutions, false)
}

// SanySubstAllParamSet is the corresponding all-parameter transformation.
func SanySubstAllParamSet(param *tlc.SymbolNode, substitutions []*SanySubst) []*tlc.SymbolNode {
	return sanyPublicSubstParameterSet(param, substitutions, true)
}

func sanyPublicSubstLevelData(body tlc.SemanticNode) *sanyLevelData {
	if sanyExploreNull(body) {
		panic(tlc.NewNullPointerException())
	}
	carrier, ok := body.(interface{ GetCanonicalLevelData() any })
	if !ok {
		panic(tlc.NewClassCastException("substitution body is not a LevelNode"))
	}
	data, ok := carrier.GetCanonicalLevelData().(*sanyLevelData)
	if !ok || data == nil {
		panic(tlc.NewUnsupportedOperationException("substitution body has no canonical SANY level metadata"))
	}
	return data
}

func (data *sanyLevelData) publicRuntimeSymbol(symbol sanySemSymbol) *tlc.SymbolNode {
	symbol = sanyLevelSymbolReference(symbol)
	if symbol == nil {
		return nil
	}
	if data == nil || data.runtimeSymbol == nil {
		panic(tlc.NewUnsupportedOperationException("canonical SANY symbol adapter has not been installed"))
	}
	return data.runtimeSymbol(symbol)
}

func sanyPublicArgLevelParamHash(param *SanyArgLevelParam) int32 {
	if param == nil {
		return 0
	}
	return param.HashCode()
}

func sanyPublicArgLevelParamEqual(left, right *SanyArgLevelParam) bool {
	return left != nil && left.Equals(right)
}

func sanyPublicArgLevelParamTieBreak(left, right *SanyArgLevelParam) int {
	return sanyRuntimeIdentityTieBreak(left, right)
}

// SanySubstGetSubALPSet translates a checked body's argument-level parameters
// through substitutions and returns them in Java HashSet iteration order.
func SanySubstGetSubALPSet(body tlc.SemanticNode, substitutions []*SanySubst) []*SanyArgLevelParam {
	data := sanyPublicSubstLevelData(body)
	result := tlc.NewJavaSemanticMap[*SanyArgLevelParam, struct{}](sanyPublicArgLevelParamHash, sanyPublicArgLevelParamEqual, sanyPublicArgLevelParamTieBreak)
	for parameter := range data.getArgLevelParams().all() {
		if parameter == nil {
			panic(tlc.NewNullPointerException())
		}
		op := data.publicRuntimeSymbol(parameter.op)
		substitution := SanySubstGetSub(op, substitutions)
		if sanyExploreNull(substitution) {
			result.Put(NewSanyArgLevelParam(op, int(parameter.i), data.publicRuntimeSymbol(parameter.param)), struct{}{})
			continue
		}
		argument, ok := substitution.(*tlc.OpArgNode)
		if !ok {
			panic(tlc.NewClassCastException("substitution is not an OpArgNode"))
		}
		substitutedOperator := argument.GetOp()
		if substitutedOperator.IsParam() {
			for _, substitutedParameter := range SanySubstParamSet(data.publicRuntimeSymbol(parameter.param), substitutions) {
				result.Put(NewSanyArgLevelParam(substitutedOperator, int(parameter.i), substitutedParameter), struct{}{})
			}
		}
	}
	parameters := make([]*SanyArgLevelParam, 0, result.Len())
	for parameter := range result.All() {
		parameters = append(parameters, parameter)
	}
	return parameters
}

func sanyPublicSubstLevelCheck(node tlc.SemanticNode, iteration int32, errors *SanyErrors) bool {
	data := sanyPublicSubstLevelData(node)
	if data.runtimeCheck == nil {
		panic(tlc.NewUnsupportedOperationException("canonical SANY level checker has not been installed"))
	}
	var diagnostics *Diagnostics
	if errors != nil {
		diagnostics = &errors.diagnostics
	}
	return data.runtimeCheck(iteration, diagnostics)
}

// SanySubstGetSubALCSet translates a checked body's argument-level constraints
// through the source substitution rules.
func SanySubstGetSubALCSet(body tlc.SemanticNode, substitutions []*SanySubst, iteration int32, errors *SanyErrors) *SanySetOfArgLevelConstraints {
	data := sanyPublicSubstLevelData(body)
	result := NewSanySetOfArgLevelConstraints()
	constraints := data.getArgLevelConstraints()
	for parameter := range constraints.entries.All() {
		level := constraints.get(parameter)
		if parameter == nil {
			panic(tlc.NewNullPointerException())
		}
		publicParameter := NewSanyParamAndPosition(data.publicRuntimeSymbol(parameter.param), int(parameter.position))
		substitution := SanySubstGetSub(data.publicRuntimeSymbol(parameter.param), substitutions)
		if sanyExploreNull(substitution) {
			result.Put(publicParameter, level)
			continue
		}
		argument, ok := substitution.(*tlc.OpArgNode)
		if !ok {
			panic(tlc.NewClassCastException("substitution is not an OpArgNode"))
		}
		substitutedOperator := argument.GetOp()
		if substitutedOperator.IsParam() {
			result.Put(NewSanyParamAndPosition(substitutedOperator, int(parameter.position)), level)
		}
	}
	for parameter := range data.getArgLevelParams().all() {
		if parameter == nil {
			panic(tlc.NewNullPointerException())
		}
		substitutedParameter := SanySubstGetSub(data.publicRuntimeSymbol(parameter.param), substitutions)
		if sanyExploreNull(substitutedParameter) {
			continue
		}
		substitutedOperator := SanySubstGetSub(data.publicRuntimeSymbol(parameter.op), substitutions)
		op := data.publicRuntimeSymbol(parameter.op)
		if !sanyExploreNull(substitutedOperator) {
			argument, ok := substitutedOperator.(*tlc.OpArgNode)
			if !ok {
				panic(tlc.NewClassCastException("substitution is not an OpArgNode"))
			}
			op = argument.GetOp()
		}
		if op.IsParam() {
			sanyPublicSubstLevelCheck(substitutedParameter, iteration, errors)
			level := int32(tlc.SemanticLevel(substitutedParameter))
			result.Put(NewSanyParamAndPosition(op, int(parameter.i)), &level)
		}
	}
	return result
}

// SanySubstGetSubLCSet translates a checked body's scalar level constraints
// through the source substitution rules.
func SanySubstGetSubLCSet(body tlc.SemanticNode, substitutions []*SanySubst, isConstant bool, iteration int32, errors *SanyErrors) *SanySetOfLevelConstraints {
	data := sanyPublicSubstLevelData(body)
	result := NewSanySetOfLevelConstraints()
	constraints := data.getLevelConstraints()
	for parameter := range constraints.entries.All() {
		level := constraints.get(parameter)
		if !isConstant {
			symbol := sanyLevelSymbolReference(parameter)
			if symbol == nil {
				panic(tlc.NewNullPointerException())
			}
			switch symbol.semKind() {
			case sanyConstantDeclKind:
				value := int32(tlc.SemanticConstantLevel)
				level = &value
			case sanyVariableDeclKind:
				value := int32(tlc.SemanticVariableLevel)
				level = &value
			}
		}
		for _, substitutedParameter := range SanySubstParamSet(data.publicRuntimeSymbol(parameter), substitutions) {
			result.Put(substitutedParameter, level)
		}
	}
	for parameter := range data.getArgLevelParams().all() {
		if parameter == nil {
			panic(tlc.NewNullPointerException())
		}
		substitution := SanySubstGetSub(data.publicRuntimeSymbol(parameter.op), substitutions)
		if sanyExploreNull(substitution) {
			continue
		}
		argument, ok := substitution.(*tlc.OpArgNode)
		if !ok {
			panic(tlc.NewClassCastException("substitution is not an OpArgNode"))
		}
		op := argument.GetOp()
		if op == nil || op.Definition == nil {
			continue
		}
		sanyPublicSubstLevelCheck(op.Definition, iteration, errors)
		level := int32(op.Definition.GetMaxLevel(int(parameter.i)))
		for _, substitutedParameter := range SanySubstParamSet(data.publicRuntimeSymbol(parameter.param), substitutions) {
			result.Put(substitutedParameter, &level)
		}
	}
	return result
}
