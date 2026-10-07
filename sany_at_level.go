// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import "github.com/glycerine/tlago/tlc"

// AtNode.levelCheck deliberately ignores child correctness. Metadata includes
// the base and earlier EXCEPT components, excluding the containing component.
func (n *sanySemAtNode) levelCheck(iter int32, errors *Diagnostics) bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	if *n.levelChecked >= iter {
		return true
	}
	*n.levelChecked = iter
	if n.exceptRef == nil {
		panic(tlc.NewNullPointerException())
	}
	args := n.exceptRef.operands
	arg := func(i int) sanyCanonicalLevelNode {
		return sanyRequireCanonicalLevelNode(sanyLevelArrayAt(args, i))
	}
	arg(0).levelCheck(iter, errors)
	*n.level = arg(0).getLevel()
	for i := 1; i < len(args); i++ {
		arg(i).levelCheck(iter, errors)
		if args[i] == n.exceptComponentRef {
			break
		}
		*n.level = max(*n.level, arg(i).getLevel())
	}
	n.levelParams.addAll(arg(0).getLevelParams())
	n.allParams.addAll(arg(0).getAllParams())
	for i := 1; i < len(args); i++ {
		if args[i] == n.exceptComponentRef {
			break
		}
		n.levelParams.addAll(arg(i).getLevelParams())
		n.allParams.addAll(arg(i).getAllParams())
	}
	n.levelConstraints.putAll(sanyLevelConstraintMap(arg(0).getLevelConstraints()))
	for i := 1; i < len(args); i++ {
		if args[i] == n.exceptComponentRef {
			break
		}
		n.levelConstraints.putAll(sanyLevelConstraintMap(arg(i).getLevelConstraints()))
	}
	n.argLevelConstraints.putAll(sanyArgLevelConstraintMap(arg(0).getArgLevelConstraints()))
	for i := 1; i < len(args); i++ {
		if args[i] == n.exceptComponentRef {
			break
		}
		n.argLevelConstraints.putAll(sanyArgLevelConstraintMap(arg(i).getArgLevelConstraints()))
	}
	n.argLevelParams.addAll(arg(0).getArgLevelParams())
	for i := 1; i < len(args); i++ {
		if args[i] == n.exceptComponentRef {
			break
		}
		n.argLevelParams.addAll(arg(i).getArgLevelParams())
	}
	return true
}
