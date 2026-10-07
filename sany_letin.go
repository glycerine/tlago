// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import "github.com/glycerine/tlago/tlc"

// LetInNode retains the supplied arrays, body and definition context. getLets
// lazily filters user-defined operators once, returning that cached array.
type sanySemLetInNode struct {
	sanySemanticNode
	opDefs     []sanySemSymbol
	instances  []sanySemanticGraphNode
	body       sanySemanticGraphNode
	context    *sanyContext
	gottenLets []*sanySemOpDefNode
}

func newSanySemLetInNode(syntax *SanySyntaxNode, definitions []sanySemSymbol, instances []sanySemanticGraphNode, body sanySemanticGraphNode, context *sanyContext) *sanySemLetInNode {
	node := &sanySemLetInNode{sanySemanticNode: newSanySemanticNode(sanyLetInKind), opDefs: definitions, instances: instances, body: body, context: context}
	if syntax == nil {
		node.TreeNode = nil
		node.Location = tlc.NullSourceLocation
	} else {
		node.TreeNode = syntax
		bridge := tlcBridge{}
		node.Location = bridge.sourceLocationForPosition(sanyNodePosition(syntax))
	}
	return node
}

func (n *sanySemLetInNode) getLets() []*sanySemOpDefNode {
	if n.gottenLets == nil {
		if n.opDefs == nil {
			panic(tlc.NewNullPointerException())
		}
		count := 0
		for _, definition := range n.opDefs {
			if definition == nil {
				panic(tlc.NewNullPointerException())
			}
			if definition.semKind() == sanyUserDefinedOpKind {
				count++
			}
		}
		n.gottenLets = make([]*sanySemOpDefNode, count)
		index := 0
		for _, definition := range n.opDefs {
			if definition.semKind() == sanyUserDefinedOpKind {
				n.gottenLets[index] = definition.(*sanySemOpDefNode)
				index++
			}
		}
	}
	return n.gottenLets
}

func (n *sanySemLetInNode) getChildren() []sanySemanticGraphNode {
	if n.opDefs == nil || n.instances == nil {
		panic(tlc.NewNullPointerException())
	}
	children := make([]sanySemanticGraphNode, len(n.opDefs)+len(n.instances)+1)
	for i, definition := range n.opDefs {
		if definition != nil {
			children[i] = definition.(sanySemanticGraphNode)
		}
	}
	copy(children[len(n.opDefs):], n.instances)
	children[len(children)-1] = n.body
	return children
}

// Ident LHS parameters own their entire declaration syntax, including the
// placeholders of operator-valued parameters. Fixity LHS parameters are tokens.
func sanyDefinitionFormalSyntax(definition *Definition, index int) *SanySyntaxNode {
	if definition.Syntax == nil {
		return nil
	}
	children := definition.Syntax.One
	if len(children) == 0 {
		return nil
	}
	lhs := children[0]
	parts := lhs.GetHeirs()
	position := -1
	switch lhs.Kind.JavaName() {
	case "N_IdentLHS":
		position = 2 + 2*index
	case "N_PrefixLHS":
		position = 1
	case "N_InfixLHS":
		position = 2 * index
	case "N_PostfixLHS":
		position = 0
	}
	if position >= 0 && position < len(parts) {
		return parts[position]
	}
	return nil
}

func (n *sanySemLetInNode) getBody() sanySemanticGraphNode {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	return n.body
}

// LetInNode.levelCheck retains the source traversal, copy and merge order.
func (n *sanySemLetInNode) levelCheck(iter int32, errors *Diagnostics) bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	if *n.levelChecked >= iter {
		return n.levelCorrect
	}
	*n.levelChecked = iter
	n.levelCorrect = true
	definition := func(i int) sanySemSymbol {
		value := sanyLevelSymbolReference(n.opDefs[i])
		if value == nil {
			panic(tlc.NewNullPointerException())
		}
		return value
	}
	checkDefinition := func(i int) sanyCanonicalLevelNode {
		value, ok := definition(i).(sanySemanticGraphNode)
		if !ok {
			panic(tlc.NewClassCastException("definition is not a level node"))
		}
		return sanyRequireCanonicalLevelNode(value)
	}
	body := func() sanyCanonicalLevelNode { return sanyRequireCanonicalLevelNode(n.body) }
	if n.opDefs == nil {
		panic(tlc.NewNullPointerException())
	}
	for i := 0; i < len(n.opDefs); i++ {
		if definition(i).semKind() != sanyModuleInstanceKind && !checkDefinition(i).levelCheck(iter, errors) {
			n.levelCorrect = false
		}
	}
	if !body().levelCheck(iter, errors) {
		n.levelCorrect = false
	}
	if n.instances == nil {
		panic(tlc.NewNullPointerException())
	}
	for i := 0; i < len(n.instances); i++ {
		if !sanyRequireCanonicalLevelNode(n.instances[i]).levelCheck(iter, errors) {
			n.levelCorrect = false
		}
	}
	*n.level = body().getLevel()
	n.levelParams = newSanyLevelSymbolSetFrom(body().getLevelParams())
	n.allParams = newSanyLevelSymbolSetFrom(body().getAllParams())
	n.levelConstraints.putAll(sanyLevelConstraintMap(body().getLevelConstraints()))
	for i := 0; i < len(n.opDefs); i++ {
		if definition(i).semKind() != sanyModuleInstanceKind {
			n.levelConstraints.putAll(sanyLevelConstraintMap(checkDefinition(i).getLevelConstraints()))
		}
	}
	n.argLevelConstraints.putAll(sanyArgLevelConstraintMap(body().getArgLevelConstraints()))
	for i := 0; i < len(n.opDefs); i++ {
		if definition(i).semKind() != sanyModuleInstanceKind {
			n.argLevelConstraints.putAll(sanyArgLevelConstraintMap(checkDefinition(i).getArgLevelConstraints()))
		}
	}
	n.argLevelParams.addAll(body().getArgLevelParams())
	for i := 0; i < len(n.opDefs); i++ {
		if definition(i).semKind() != sanyModuleInstanceKind {
			params := []sanySemSymbol{}
			if definition(i).semKind() != sanyThmOrAssumpDefKind {
				op, ok := definition(i).(*sanySemOpDefNode)
				if !ok {
					panic(tlc.NewClassCastException("definition is not an OpDefNode"))
				}
				formals := op.getParams()
				if formals == nil {
					params = nil
				} else {
					params = make([]sanySemSymbol, len(formals))
					for j, p := range formals {
						params[j] = p
					}
				}
			}
			for alp := range checkDefinition(i).getArgLevelParams().all() {
				if !alp.occur(params) {
					n.argLevelParams.add(alp)
				}
			}
		}
	}
	for i := 0; i < len(n.instances); i++ {
		n.argLevelParams.addAll(sanyRequireCanonicalLevelNode(n.instances[i]).getArgLevelParams())
	}
	return n.levelCorrect
}
