package tlago

import "github.com/glycerine/tlago/tlc"

type sanySemUseOrHideNode struct {
	sanySemanticNode
	facts    []sanySemanticGraphNode
	defs     []sanySemSymbol
	isOnly   bool
	stepName *tlc.UniqueString
}

func newSanySemUseOrHideNode(kind sanySemKind, syntax *SanySyntaxNode, facts []sanySemanticGraphNode, definitions []sanySemSymbol, only bool) *sanySemUseOrHideNode {
	node := &sanySemUseOrHideNode{sanySemanticNode: newSanySemanticNode(kind), facts: facts, defs: definitions, isOnly: only}
	sanyAssertionSyntax(&node.sanySemanticNode, syntax)
	return node
}
func (node *sanySemUseOrHideNode) setStepName(name *tlc.UniqueString) { node.stepName = name }
func (node *sanySemUseOrHideNode) getStepName() *tlc.UniqueString     { return node.stepName }
func (node *sanySemUseOrHideNode) getChildren() []sanySemanticGraphNode {
	if len(node.facts) == 0 {
		return nil
	}
	children := make([]sanySemanticGraphNode, len(node.facts))
	copy(children, node.facts)
	return children
}
func (node *sanySemUseOrHideNode) factCheck() Diagnostics {
	if node.facts == nil || node.getKind() == sanyUseKind {
		return nil
	}
	var diagnostics Diagnostics
	for _, fact := range node.facts {
		if fact == nil {
			panic(tlc.NewNullPointerException(""))
		}
		if fact.Kind() != tlc.SemanticOpApplKind {
			continue
		}
		application, ok := fact.(*sanySemOpApplNode)
		if !ok {
			panic(tlc.NewClassCastException(""))
		}
		if application.operator == nil {
			panic(tlc.NewNullPointerException(""))
		}
		location := application.Location
		if application.operator.semKind() == sanyThmOrAssumpDefKind {
			continue
		}

		begin := Position{File: location.Source, Line: location.BeginLine, Column: location.BeginColumn}
		end := Position{File: location.Source, Line: location.EndLine, Column: location.EndColumn}
		message := "The only expression allowed as a fact in a HIDE is \nthe name of a theorem, assumption, or step."
		diagnostic := errorAt(begin, "E4357", "%s", message)
		diagnostic.SANYMessage = message
		diagnostic.SANYRange = SanyRange{Begin: begin, End: end}
		diagnostics = append(diagnostics, diagnostic)
	}
	return diagnostics
}

// SanyUseOrHideNodeFactCheck adapts UseOrHideNode.factCheck(Errors). It reads
// the public node's current fact array, preserving Java's mutable public field
// semantics rather than consulting the parser's original canonical slice.
func SanyUseOrHideNodeFactCheck(node *tlc.UseOrHideNode, errors *SanyErrors) {
	if node == nil {
		panic(tlc.NewNullPointerException())
	}
	if node.Facts == nil || node.Kind() == tlc.SemanticUseKind {
		return
	}
	message := "The only expression allowed as a fact in a HIDE is \nthe name of a theorem, assumption, or step."
	code := SanyErrorCodeFromStandardValue(4357)
	for _, fact := range node.Facts {
		if fact == nil {
			panic(tlc.NewNullPointerException())
		}
		if tlc.SemanticKindOf(fact) != tlc.SemanticOpApplKind {
			continue
		}
		application, ok := fact.(*tlc.OpApplNode)
		if !ok {
			panic(tlc.NewClassCastException())
		}
		if application.Operator == nil {
			panic(tlc.NewNullPointerException())
		}
		operatorKind := tlc.SemanticUnknownKind
		if application.Operator.SemanticBase != nil {
			operatorKind = application.Operator.SemanticBase.Kind()
		} else if application.Operator.Definition != nil {
			operatorKind = application.Operator.Definition.Kind()
		}
		if operatorKind == tlc.SemanticThmOrAssumpKind {
			continue
		}
		location := application.GetSourceLocation()
		errors.AddMessage(code, &location, message)
	}
}
