package tlc

import "fmt"

func LivenessFindViolatedProperties(tool *Tool, states []*TLCStateMut, cyclePos int) []string {
	if states == nil {
		panic(NewNullPointerException())
	}
	cyclePos = int(int32(cyclePos))
	if len(states) == 0 {
		panic(NewAssertionError("Counter-example trace must contain at least one state"))
	}
	if cyclePos < 0 || cyclePos >= len(states) {
		panic(NewAssertionError(fmt.Sprintf("cyclePos=%d out of bounds for states.size()=%d", cyclePos, len(states))))
	}
	checks := tool.GetImpliedTemporals()
	violated := make([]string, 0)
	seen := make(map[string]struct{})
	for _, check := range checks {
		name, violates, err := livenessPropertyViolation(tool, check, states, cyclePos)
		if err != nil {
			printLivenessAttributionFailure(err)
			continue
		}
		if violates {
			if _, duplicate := seen[name]; !duplicate {
				seen[name] = struct{}{}
				violated = append(violated, name)
			}
		}
	}
	if len(violated) == 0 {
		panic(NewAssertionError("Failed to determine property"))
	}
	return violated
}

func livenessPropertyViolation(tool *Tool, check *Action, states []*TLCStateMut, cyclePos int) (name string, violates bool, err error) {
	defer func() {
		if failure := recover(); failure != nil {
			exception, translated := failure.(error)
			if !translated || isJavaError(exception) {
				panic(failure)
			}
			name, violates, err = "", false, exception
		}
	}()
	if check == nil {
		panic(NewNullPointerException())
	}
	prop, err := ASTToLive(tool, livenessExprNode(check.ownedPred()), check.ownedContext())
	if err != nil {
		return "", false, err
	}
	ok, err := prop.EvalOnLasso(tool, states, cyclePos, 0)
	if err != nil || ok {
		return "", false, err
	}
	return check.GetNameOfDefault(), true, nil
}

func LivenessFindViolatedPropertiesStuttering(tool *Tool, states []*TLCStateMut) []string {
	return LivenessFindViolatedProperties(tool, states, len(states)-1)
}

func printLivenessAttributionFailure(err error) {
	if failure, ok := err.(*TLCError); ok && failure.Runtime {
		PrintTLCRuntimeException(failure)
		return
	}
	PrintErrorThrowable(ECGeneral, err)
}

func LivenessFindViolatedPropertiesFromTrace(tool *Tool, trace []*TLCStateInfo, cyclePos int) []string {
	states := make([]*TLCStateMut, len(trace))
	for i, info := range trace {
		if info != nil && info.ownedState() != nil {
			states[i] = info.ownedState()
		}
	}
	return LivenessFindViolatedProperties(tool, states, cyclePos)
}

func LivenessPrintStutteringCounterExampleWarning(tool *Tool) {
	specName := tool.GetModelConfig().GetSpec()
	if specName != "" && len(tool.GetSpecProcessor().GetTemporal()) == 0 {
		definition := tool.GetSpecProcessor().GetDefns().Get(specName)
		op, ok := definition.(*OpDefNode)
		if !ok {
			if definition == nil {
				panic(NewNullPointerException())
			}
			panic(valueStreamClassCast(definition, "tla2sany.semantic.OpDefNode"))
		}
		if op == nil {
			panic(NewNullPointerException())
		}
		PrintWarning(ECTLCConfigNoFairnessButLiveProperty, specName, op.SourceLocation().String())
		return
	}
	if specName == "" {
		PrintWarning(ECTLCConfigNoSpecButProperty, "")
	}
}
