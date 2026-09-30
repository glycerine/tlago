package tlc

const actionLiveExprAuxKey = "tlc.liveExpr"

func AttachLiveExprToAction(action *Action, expr *LiveExprNode) *Action {
	if action == nil || expr == nil {
		return action
	}
	action.GetAuxiliary()[actionLiveExprAuxKey] = expr
	return action
}

func LiveExprFromAction(action *Action) (*LiveExprNode, bool) {
	if action == nil {
		return nil, false
	}
	if expr, ok := action.Pred.(*LiveExprNode); ok {
		return expr, true
	}
	if action.Auxiliary != nil {
		if expr, ok := action.Auxiliary[actionLiveExprAuxKey].(*LiveExprNode); ok {
			return expr, true
		}
	}
	return nil, false
}

func LivenessFindViolatedProperties(tool *Tool, states []*TLCStateMut, cyclePos int) []string {
	if tool == nil || len(states) == 0 {
		return nil
	}
	if cyclePos < 0 || cyclePos >= len(states) {
		cyclePos = len(states) - 1
	}
	checks := tool.GetImpliedTemporals()
	violated := make([]string, 0)
	for _, check := range checks {
		prop, ok := LiveExprFromAction(check)
		if !ok {
			live, err := ASTToLive(tool, check.Pred, check.Con)
			if err != nil {
				PrintError(ECGeneral, err.Error())
				prop = liveExprFallbackForAction(check)
			} else {
				prop = live
				AttachLiveExprToAction(check, prop)
			}
		}
		if prop == nil {
			continue
		}
		ok, err := prop.EvalOnLasso(tool, states, cyclePos, 0)
		if err != nil {
			PrintError(ECGeneral, err.Error())
			continue
		}
		if !ok {
			violated = append(violated, check.GetNameOfDefault())
		}
	}
	return violated
}

func LivenessFindViolatedPropertiesFromTrace(tool *Tool, trace []*TLCStateInfo, cyclePos int) []string {
	states := make([]*TLCStateMut, 0, len(trace))
	for _, info := range trace {
		if info != nil && info.State != nil {
			states = append(states, info.State)
		}
	}
	return LivenessFindViolatedProperties(tool, states, cyclePos)
}

func LivenessPrintStutteringCounterExampleWarning(tool *Tool) {
	if tool == nil {
		return
	}
	config := tool.GetModelConfig()
	specName := ""
	if config != nil {
		specName = config.GetSpec()
	}
	if specName != "" && len(tool.GetTemporals()) == 0 {
		PrintWarning(ECTLCConfigNoFairnessButLiveProperty, specName, livenessSpecLocation(tool, specName))
		return
	}
	if specName == "" {
		PrintWarning(ECTLCConfigNoSpecButProperty, "")
	}
}

func livenessSpecLocation(tool *Tool, specName string) string {
	if tool == nil || tool.SpecProcessor == nil {
		return specName
	}
	if op, ok := tool.SpecProcessor.defn(specName).(*OpDefNode); ok && op != nil {
		if loc, ok := semanticNodeSourceLocation(op.Body); ok && !loc.IsNull() {
			return loc.String()
		}
		if op.Body != nil {
			return SemanticString(op.Body)
		}
	}
	return specName
}

func liveExprFallbackForAction(action *Action) *LiveExprNode {
	if action == nil || action.Pred == nil {
		return nil
	}
	level := SemanticLevel(action.Pred)
	label := action.GetNameOfDefault()
	if level <= TLCLevelState {
		return NewLNState(label, action.Pred, action.Con, nil)
	}
	return NewLNAction(label, action.Pred, action.Con, nil)
}
