// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import (
	"fmt"
	"github.com/glycerine/tlago/tlc"
)

// InstanceNode checks the resolved substitution array, including defaults. Its
// expressions retain their instancer's lexical context; a completed-module scan
// cannot reconstruct either the bindings or the substitution order.
func (c *sanyLevelCompositionChecker) checkInstanceSubstitutionLevels(instance Instance, declKinds map[string]DeclarationKind) Diagnostics {
	spec := c.dependencies.resolver.spec
	target := spec.Modules[instance.Module]
	if target == nil {
		return nil
	}
	parameterNames := map[string]bool{}
	declKinds = copyDeclKindMap(declKinds)
	for _, name := range instance.Params {
		parameterNames[name] = true
		delete(declKinds, name)
	}
	context := c.contextWithLocals(parameterNames)
	checker := &sanyLevelCompositionChecker{dependencies: c.dependencies, context: context}
	var diags Diagnostics
	valid := map[string]bool{}
	for _, substitution := range instance.generatedSubstitutions {
		componentDiags := checker.check(substitution.expr, parameterNames)
		valid[substitution.name] = !componentDiags.HasErrors()
		diags = append(diags, componentDiags...)
	}
	levelDiagnostic := func(name string, maximum tlaLevel) Diagnostic {
		message := fmt.Sprintf("Level error in instantiating module '%s':\nThe level of the expression or operator substituted for '%s' \nmust be at most %d.", instance.Module, name, maximum)
		return sanyDiagnosticParameters(sanyRegistrationDiagnostic(instance.SourcePosition(), "E4245", "%s", message), instance.Module, name, maximum)
	}
	substitutions := map[string]sanyGeneratedSubstitution{}
	matchLevels := moduleRequiresSubstitutionLevelMatch(target, spec)
	for _, substitution := range instance.generatedSubstitutions {
		name, expr := substitution.name, substitution.expr
		substitutions[name] = substitution
		if matchLevels && valid[name] {
			maximum := constantLevel
			if substitution.target.Kind == VariableDecl {
				maximum = variableLevel
			}
			if checker.level(expr, parameterNames) > maximum {
				diags = append(diags, levelDiagnostic(name, maximum))
			}
		}
		if valid[name] && substitution.target.Arity > 0 && c.dependencies.operatorNonLeibnizInContext(expr, substitution.target.Arity, context) {
			message := fmt.Sprintf("Error in instantiating module '%s':\n A non-Leibniz operator substituted for '%s'.", instance.Module, name)
			diags = append(diags, sanyDiagnosticParameters(sanyRegistrationDiagnostic(instance.SourcePosition(), "E4244", "%s", message), instance.Module, name))
		}
	}
	names, moduleUse := moduleSubstitutionConstraints(target, spec)
	ids := map[string]int{}
	constraints := map[string]tlaLevel{}
	for id, name := range names {
		ids[name] = id
		if maximum, exists := moduleUse.constraints[id]; exists {
			constraints[name] = maximum
		}
	}
	for _, substitution := range instance.generatedSubstitutions {
		name, expr := substitution.name, substitution.expr
		if maximum, constrained := constraints[name]; valid[name] && constrained && checker.level(expr, parameterNames) > maximum {
			diags = append(diags, levelDiagnostic(name, maximum))
		}
		if substitution.target.Arity > 0 && valid[name] {
			maxima := c.dependencies.applicationMaximums(expr, nil, substitution.target.Arity, context)
			operatorName := name
			if ident, ok := expr.(*IdentExpr); ok {
				operatorName = ident.Name
			}
			for i, maximum := range maxima {
				if minimum, exists := moduleUse.argConstraints[sanyArgumentPosition{ids[name], i}]; exists && maximum < minimum {
					message := fmt.Sprintf("Level error in instantiating module '%s':\nThe level of the argument %d of the operator %s \nmust be at least %d.", instance.Module, i+1, operatorName, minimum)
					diags = append(diags, sanyDiagnosticParameters(sanyRegistrationDiagnostic(instance.SourcePosition(), "E4246", "%s", message), instance.Module, i+1, operatorName, int(minimum)))
				}
			}
		}
		diags = append(diags, checkPrimedConstants(expr, declKinds, parameterNames)...)
	}
	// ArgLevelParam.hashCode adds both declaration hashes and the zero-based
	// argument position. HashSet deduplicates relationships before InstanceNode
	// iterates them; repeated applications do not report repeated constraints.
	hashes := map[int]int32{}
	for _, substitution := range instance.generatedSubstitutions {
		if substitution.declaration != nil {
			hashes[ids[substitution.name]] = substitution.declaration.hashCode()
		}
	}
	relationships := tlc.NewJavaSemanticSet[sanyArgumentParameter](func(key sanyArgumentParameter) int32 {
		return hashes[key.operator] + int32(key.position) + hashes[key.parameter]
	})
	for _, key := range moduleUse.argParamOrder {
		if key.operator < len(names) && key.parameter < len(names) {
			relationships.Add(key)
		}
	}
	for key := range relationships.All() {
		operatorName, parameterName := names[key.operator], names[key.parameter]
		operator, hasOperator := substitutions[operatorName]
		parameter, hasParameter := substitutions[parameterName]
		if !hasOperator || !hasParameter || !valid[operatorName] || !valid[parameterName] || operator.target.Arity <= key.position {
			continue
		}
		maxima := c.dependencies.applicationMaximums(operator.expr, nil, operator.target.Arity, context)
		maximum := maxima[key.position]
		if checker.level(parameter.expr, parameterNames) > maximum {
			message := fmt.Sprintf("Level error when instantiating module '%s':\nThe level of the argument %d of the operator %s' \nmust be at most %d.", instance.Module, key.position, operatorName, maximum)
			diags = append(diags, sanyDiagnosticParameters(sanyRegistrationDiagnostic(instance.SourcePosition(), "E4247", "%s", message), instance.Module, key.position, operatorName, int(maximum)))
		}
	}
	return diags
}
