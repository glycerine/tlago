// Copyright (c) 2007 Microsoft Corporation. All rights reserved.
package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"reflect"
)

func (n *sanySemNewSymbNode) levelCheck(iter int32, errors *Diagnostics) bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	if *n.levelChecked < iter {
		*n.levelChecked = iter
		declarationCorrect := n.opDeclNode.levelCheck(iter, errors)
		*n.level = n.opDeclNode.getLevel()
		hasSet := func() bool { return n.set != nil && !reflect.ValueOf(n.set).IsNil() }
		set := func() sanyCanonicalLevelNode { return sanyRequireCanonicalLevelNode(n.set) }
		if hasSet() {
			n.levelCorrect = set().levelCheck(iter, errors)
			*n.level = max(set().getLevel(), *n.level)
			if *n.level == temporalLevel {
				n.levelCorrect = false
				// Java evaluates stn.getLocation before calling the Errors receiver.
				syntax, ok := n.TreeNode.(*SanySyntaxNode)
				if n.TreeNode == nil || ok && syntax == nil {
					panic(tlc.NewNullPointerException())
				}
				if !ok {
					panic(tlc.NewClassCastException("NEW symbol syntax"))
				}
				location := syntax.Range
				if errors == nil {
					panic(tlc.NewNullPointerException())
				}
				message := "Level error:\nTemporal formula used as set."
				diagnostic := errorAt(location.Begin, "E4356", "%s", message)
				diagnostic.SANYMessage = message
				diagnostic.SANYRange = location
				present := false
				for _, existing := range *errors {
					if existing.Code == diagnostic.Code &&
						existing.SANYRange.Begin.File == location.Begin.File &&
						existing.SANYRange.Begin.Line == location.Begin.Line &&
						existing.SANYRange.Begin.Column == location.Begin.Column &&
						existing.SANYRange.End.Line == location.End.Line &&
						existing.SANYRange.End.Column == location.End.Column &&
						existing.SANYMessage == message {
						present = true
						break
					}
				}
				if !present {
					*errors = append(*errors, diagnostic)
				}
			}
		}
		n.levelCorrect = n.levelCorrect && declarationCorrect
		if hasSet() {
			n.levelParams = set().getLevelParams()
			n.allParams = set().getAllParams()
			n.levelConstraints = set().getLevelConstraints()
			n.argLevelConstraints = set().getArgLevelConstraints()
			n.argLevelParams = set().getArgLevelParams()
		}
	}
	return n.levelCorrect
}
