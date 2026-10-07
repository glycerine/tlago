// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import "fmt"

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
	substitutions := map[string]Expr{}
	matchLevels := moduleRequiresSubstitutionLevelMatch(target, spec)
	for _, substitution := range instance.generatedSubstitutions {
		name, expr := substitution.name, substitution.expr
		substitutions[name] = expr
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
	diags = append(diags, checker.checkInstanceSubstitutionCoparameterLevelConstraints(target, spec, substitutions, instance.SourcePosition(), declKinds)...)
	return diags
}
