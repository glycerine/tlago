package tlc

import "fmt"

type osExprPem struct {
	EAAction []*LiveExprNode
	AEState  []*LiveExprNode
	AEAction []*LiveExprNode
	TFs      []*LiveExprNode
}

func (p *osExprPem) toTFS() *LiveExprNode {
	if p == nil {
		return nil
	}
	switch len(p.TFs) {
	case 0:
		return nil
	case 1:
		return p.TFs[0]
	default:
		return NewLNConj(p.TFs...)
	}
}

func ParseLiveness(tool *Tool) (*LiveExprNode, error) {
	fairs := tool.GetTemporals()
	livespec := NewLNConj()
	for _, fair := range fairs {
		live, err := ASTToLive(tool, fair.Pred, fair.Con)
		if err != nil {
			return nil, err
		}
		livespec.AddConj(live)
	}

	checks := tool.GetImpliedTemporals()
	switch len(checks) {
	case 0:
		if len(fairs) == 0 {
			return nil, nil
		}
	case 1:
		live, err := ASTToLive(tool, checks[0].Pred, checks[0].Con)
		if err != nil {
			return nil, err
		}
		negated := NewLNNeg(live)
		if livespec.Count() == 0 {
			return negated, nil
		}
		livespec.AddConj(negated)
	default:
		disj := NewLNDisj()
		for _, check := range checks {
			live, err := ASTToLive(tool, check.Pred, check.Con)
			if err != nil {
				return nil, err
			}
			disj.AddDisj(NewLNNeg(live))
		}
		if livespec.Count() == 0 {
			return disj, nil
		}
		livespec.AddConj(disj)
	}
	return livespec, nil
}

func ProcessLiveness(tool *Tool) ([]*OrderOfSolution, error) {
	return ProcessLivenessSilent(tool, false)
}

func ProcessLivenessSilent(tool *Tool, silent bool) ([]*OrderOfSolution, error) {
	lexpr, err := ParseLiveness(tool)
	if err != nil {
		return nil, err
	}
	if lexpr == nil {
		return []*OrderOfSolution{}, nil
	}

	lexpr.TagExpr(1)
	lexpr = lexpr.PushNegWith(false).Simplify().ToDNF()
	if lexpr != nil && lexpr.Kind == LiveExprBool && !lexpr.Bool {
		return []*OrderOfSolution{}, nil
	}

	dnf := lexpr
	if dnf == nil || dnf.Kind != LiveExprDisj {
		dnf = NewLNDisj(lexpr)
	}

	pems := make([]*osExprPem, dnf.Count())
	tfs := make([]*LiveExprNode, dnf.Count())
	for i := 0; i < dnf.Count(); i++ {
		ln := dnf.GetBody(i).FlattenSingleJunctions()
		pem := &osExprPem{}
		pems[i] = pem
		if ln != nil && ln.Kind == LiveExprConj {
			for j := 0; j < ln.Count(); j++ {
				if err := classifyLivenessExpr(ln.GetBody(j), pem); err != nil {
					return nil, err
				}
			}
		} else if err := classifyLivenessExpr(ln, pem); err != nil {
			return nil, err
		}
		tfs[i] = pem.toTFS()
	}

	var tfbin []*LiveExprNode
	var pembin [][]*osExprPem
	for i, tf := range tfs {
		found := -1
		for j, tf0 := range tfbin {
			if (tf == nil && tf0 == nil) || (tf != nil && tf0 != nil && tf.Equal(tf0)) {
				found = j
				break
			}
		}
		if found == -1 {
			found = len(tfbin)
			tfbin = append(tfbin, tf)
			pembin = append(pembin, nil)
		}
		pembin[found] = append(pembin[found], pems[i])
	}

	solutions := make([]*OrderOfSolution, len(tfbin))
	for i, tf := range tfbin {
		if tf == nil {
			solutions[i] = NewOrderOfSolution(nil, nil)
		} else {
			tf1 := tf.MakeBinary()
			solutions[i] = NewOrderOfSolution(NewTBGraph(tf1), tf1.ExtractPromises())
		}

		stateBin := make([]*LiveExprNode, 0)
		actionBin := make([]*LiveExprNode, 0)
		models := make([]*PossibleErrorModel, len(pembin[i]))
		for j, pem := range pembin[i] {
			models[j] = NewPossibleErrorModel(
				addLivenessChecksToBin(pem.AEAction, &actionBin),
				addLivenessChecksToBin(pem.AEState, &stateBin),
				addLivenessChecksToBin(pem.EAAction, &actionBin),
			)
		}
		solutions[i].SetPEMs(models)
		solutions[i].SetCheckState(stateBin)
		solutions[i].SetCheckAction(actionBin)
	}

	if !silent {
		PrintMessage(ECTLCLiveImplied, fmt.Sprintf("%d", len(solutions)))
	}
	return solutions, nil
}

func classifyLivenessExpr(expr *LiveExprNode, pem *osExprPem) error {
	if expr == nil {
		return nil
	}
	if expr.Kind == LiveExprEven {
		if all := expr.Body; all != nil && all.Kind == LiveExprAll {
			body := all.Body
			if body != nil && body.GetLevel() < LiveLevelTemporal {
				pem.EAAction = append(pem.EAAction, body)
				return nil
			}
		}
	} else if expr.Kind == LiveExprAll {
		if even := expr.Body; even != nil && even.Kind == LiveExprEven {
			body := even.Body
			if body == nil {
				return newTLCError(ECTLCLiveWrongFormulaFormat, "liveness formula has the wrong format")
			}
			level := body.GetLevel()
			if level <= LiveLevelState {
				pem.AEState = append(pem.AEState, body)
				return nil
			}
			if level == LiveLevelAction {
				pem.AEAction = append(pem.AEAction, body)
				return nil
			}
		}
	}
	if expr.ContainAction() {
		return newTLCError(ECTLCLiveWrongFormulaFormat, "liveness formula has the wrong format")
	}
	pem.TFs = append(pem.TFs, expr)
	return nil
}

func addLivenessCheckToBin(check *LiveExprNode, bin *[]*LiveExprNode) int {
	if check == nil {
		return -1
	}
	for i, existing := range *bin {
		if check.Equal(existing) {
			return i
		}
	}
	*bin = append(*bin, check)
	return len(*bin) - 1
}

func addLivenessChecksToBin(checks []*LiveExprNode, bin *[]*LiveExprNode) []int {
	out := make([]int, len(checks))
	for i, check := range checks {
		out[i] = addLivenessCheckToBin(check, bin)
	}
	return out
}

func NewLiveCheckFromTool(tool *Tool, metadir string, stateWriter *StateWriter) (*LiveCheck, error) {
	solutions, err := ProcessLiveness(tool)
	if err != nil {
		return nil, err
	}
	return NewLiveCheckWithStateWriter(tool, solutions, metadir, stateWriter)
}
