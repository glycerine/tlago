// Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"strconv"
	"strings"
)

func (n *sanySemAssumeNode) levelCheck(iter int32, errors *Diagnostics) bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	if n.assumeLevelChecked >= iter {
		return true
	}
	n.assumeLevelChecked = iter
	body := func() sanyCanonicalLevelNode { return sanyRequireCanonicalLevelNode(n.assumeExpr) }
	result := body().levelCheck(iter, errors)
	if n.def != nil {
		correct := n.def.levelCheck(iter, errors)
		result = correct && result
	}
	if body().getLevel() != constantLevel {
		syntax, ok := n.TreeNode.(*SanySyntaxNode)
		if n.TreeNode == nil || ok && syntax == nil {
			panic(tlc.NewNullPointerException())
		}
		if !ok {
			panic(tlc.NewClassCastException("assumption syntax"))
		}
		location := syntax.Range
		level := n.getLevel()
		if errors == nil {
			panic(tlc.NewNullPointerException())
		}
		message := "Level error: assumptions must be level 0 (Constant), \nbut this one has level " + strconv.Itoa(int(level)) + "."
		diagnostic := errorAt(location.Begin, "E4206", "%s", message)
		diagnostic.SANYRange = location
		diagnostic.SANYMessage = message
		diagnostic.SANYParameters = []any{int32(level)}
		present := false
		for _, existing := range *errors {
			if existing.Code == diagnostic.Code && existing.SANYMessage == message && sanyLevelMessageParametersEqual(existing.SANYParameters, diagnostic.SANYParameters) &&
				existing.SANYRange.Begin.File == location.Begin.File && existing.SANYRange.Begin.Line == location.Begin.Line && existing.SANYRange.Begin.Column == location.Begin.Column && existing.SANYRange.End.Line == location.End.Line && existing.SANYRange.End.Column == location.End.Column {
				present = true
				break
			}
		}
		if !present {
			*errors = append(*errors, diagnostic)
		}
	}
	if result {
		sanyAddTemporalLevelConstraintToConstants(n.levelParams, n.levelConstraints)
	}
	return result
}

func (n *sanySemAssumeNode) getLevel() tlaLevel {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	return sanyRequireCanonicalLevelNode(n.assumeExpr).getLevel()
}

func (n *sanySemAssumeNode) getLevelParams() *sanyLevelSet[sanySemSymbol] {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	return sanyRequireCanonicalLevelNode(n.assumeExpr).getLevelParams()
}

func (n *sanySemAssumeNode) getAllParams() *sanyLevelSet[sanySemSymbol] {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	return sanyRequireCanonicalLevelNode(n.assumeExpr).getAllParams()
}

func (n *sanySemAssumeNode) getLevelConstraints() *sanySetOfLevelConstraints {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	return sanyRequireCanonicalLevelNode(n.assumeExpr).getLevelConstraints()
}

func (n *sanySemAssumeNode) getArgLevelConstraints() *sanySetOfArgLevelConstraints {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	return sanyRequireCanonicalLevelNode(n.assumeExpr).getArgLevelConstraints()
}

func (n *sanySemAssumeNode) getArgLevelParams() *sanyLevelSet[*sanyArgLevelParam] {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	return sanyRequireCanonicalLevelNode(n.assumeExpr).getArgLevelParams()
}

// AssumeNode overrides the common formatter without an inherited iteration guard.
func (n *sanySemAssumeNode) levelDataToString() string {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	out := "Level: " + strconv.Itoa(int(n.getLevel())) + "\n"
	params := n.getLevelParams()
	paramString := "null"
	if params != nil {
		paramString = sanyLevelSymbolSetString(params)
	}
	out += "LevelParameters: " + paramString + "\n"
	lc := n.getLevelConstraints()
	lcString := "null"
	if lc != nil {
		lcString = lc.String()
	}
	out += "LevelConstraints: " + lcString + "\n"
	ac := n.getArgLevelConstraints()
	acString := "null"
	if ac != nil {
		acString = ac.String()
	}
	out += "ArgLevelConstraints: " + acString + "\n"
	alp := n.getArgLevelParams()
	alpString := "null"
	if alp != nil {
		var elements []string
		for p := range alp.all() {
			if p == nil {
				elements = append(elements, "null")
			} else {
				elements = append(elements, p.String())
			}
		}
		alpString = "[" + strings.Join(elements, ", ") + "]"
	}
	return out + "ArgLevelParams: " + alpString + "\n"
}
