// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import (
	"fmt"
	"github.com/glycerine/tlago/tlc"
)

// FormalParamNode owns its SemanticNode identity. The enclosing native module
// and retained parser node preserve declaration ownership for graph generation.
// Its LevelNode data and graph visitors are ported separately.
type sanyFormalParamNode struct {
	sanySemSymbolBase
	nullName       bool
	module         *Module
	semanticModule *sanySemModuleNode
}

func newSanyFormalParamNode(name string, arity int, position Position, syntax *SanySyntaxNode, module *Module) *sanyFormalParamNode {
	moduleName := ""
	if module != nil {
		moduleName = module.Name
	}
	node := &sanyFormalParamNode{
		sanySemSymbolBase: sanySemSymbolBase{
			sanySemanticNode: newSanySemanticNode(sanyFormalParamKind),
			name:             name, arity: arity, local: true,
			originalModuleName: moduleName, pos: position,
		},
		module: module,
	}
	if module != nil {
		node.semanticModule = module.semanticNode
	}
	if syntax == nil {
		node.TreeNode = nil
		node.Location = tlc.NullSourceLocation
		node.pos = Position{}
	} else {
		node.TreeNode = syntax
		bridge := tlcBridge{convertingModule: moduleName}
		node.Location = bridge.sourceLocationForPosition(position)
	}
	return node
}

// SemanticNode.equals checks the concrete class, kind and UID, rather than the
// formal's name or declaration location.
func (n *sanyFormalParamNode) equals(other any) bool {
	o, ok := other.(*sanyFormalParamNode)
	if !ok || o == nil {
		return false
	}
	return n == o || n.getKind() == o.getKind() && n.getUID() == o.getUID()
}

func (n *sanyFormalParamNode) match(application *sanySemOpApplNode) bool {
	return application.operator.semArity() == n.semArity()
}

func (g *sanyExpressionGeneration) newFormalParameter(name string, arity int, position Position, syntax *SanySyntaxNode) *sanyFormalParamNode {
	return newSanyFormalParamNode(name, arity, position, sanySyntaxAtPosition(syntax, position), g.currentModule)
}

// Full module generation owns its existing SymbolTable. Standalone expression
// generation starts with the same global builtin context as Java Generator.
func (g *sanyExpressionGeneration) formalSymbolTable() *sanySymbolTable {
	if g.currentModule != nil && g.currentModule.symbolTable != nil {
		return g.currentModule.symbolTable
	}
	if g.formalTable == nil {
		g.formalTable = newSanySymbolTable(sanyGlobalInitialContext(false).duplicate(), nil)
	}
	return g.formalTable
}

func (g *sanyExpressionGeneration) pushFormalContext(capacity int) func() {
	table := g.formalSymbolTable()
	table.pushContext(newSanyContext())
	previous := g.formals
	g.formals = make(map[string]localSymbol, len(previous)+capacity)
	for name, symbol := range previous {
		g.formals[name] = symbol
	}
	return func() {
		g.formals = previous
		table.popContext()
	}
}

// SymbolTable.addSymbol keeps the earlier binding after a rejected formal
// declaration. The caller retains the new node in its parameter array.
func (g *sanyExpressionGeneration) bindFormalParameter(node *sanyFormalParamNode, context map[string]Position, locals map[string]bool) Diagnostics {
	name, position := node.semName(), node.semPosition()
	if _, builtin := builtinOperatorArity(name); builtin || builtinIdentifiers[name] {
		diagnostic := sanyDiagnosticParameters(errorAt(position, "E4202", "cannot redefine built-in symbol %s", name), name)
		diagnostic.SANYMessage = fmt.Sprintf("Symbol %s is a built-in operator, and cannot be redefined.", name)
		return Diagnostics{diagnostic}
	}
	if symbol, exists := g.lookupSymbol(name, context); exists {
		diagnostic := sanyDiagnosticParameters(errorAt(position, "E4201", "bound symbol %s conflicts with existing symbol declared at %s", name, symbol.pos), name, sanySymbolLocation(symbol.pos))
		diagnostic.SANYMessage = fmt.Sprintf("Multiply-defined symbol '%s': this definition or declaration conflicts \nwith the one at %s.", name, sanySymbolLocation(symbol.pos))
		return Diagnostics{diagnostic}
	}
	if locals[name] {
		// Proof/native locals without nodes retain their representation.
		return checkBoundName(name, position, context, locals)
	}
	accepted, diags := g.formalSymbolTable().registerSymbol(node)
	if accepted {
		g.formals[name] = localSymbol{formalNode: node, kind: "FORMAL", arity: node.semArity(), pos: position}
	}
	return diags
}

// generateLabel resolves its parameter array after generating the body. A
// non-formal symbol receives a fresh dummy node for each argument occurrence.
// Diagnostic traversal and the LS parameter stack are integrated separately.
func (g *sanyExpressionGeneration) resolveLabelFormals(label *LabelExpr, context map[string]Position) {
	label.formalNodes = nil
	label.illegalParameterSyntax = nil
	if label.Syntax == nil {
		return
	}
	label.formalNodes = make([]*sanyFormalParamNode, 0, len(label.Params))
	label.illegalParameterSyntax = make([]*SanySyntaxNode, len(label.Params))
	heirs := label.Syntax.GetHeirs()
	if len(heirs) == 0 || heirs[0].Kind.JavaName() != "N_OpApplication" {
		return
	}
	args := heirs[0].GetHeirs()[1].GetHeirs()
	for i, name := range label.Params {
		symbol, _ := g.lookupSymbol(name, context)
		node := symbol.formalNode
		if node == nil {
			argument := args[2*i+1]
			label.illegalParameterSyntax[i] = argument
			node = newSanyFormalParamNode(name, 0, sanyNodePosition(argument), argument, g.currentModule)
		}
		label.formalNodes = append(label.formalNodes, node)
	}
}
