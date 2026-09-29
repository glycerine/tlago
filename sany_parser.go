package tlago

import (
	"strconv"
	"strings"
)

type SanyParser struct {
	tokens          []*SanyToken
	at              int
	diags           Diagnostics
	moduleName      string
	proofLevelStack []int
}

func ParseSanySyntax(file, source string) (*SanySyntaxNode, Diagnostics) {
	tokens, lexDiags := SanyTokenize(file, source)
	parser := NewSanyParser(tokens, nil)
	node := parser.CompilationUnit()
	diags := filterSanyDiagnosticsThroughRootEnd(lexDiags, node)
	diags = append(diags, parser.diags...)
	return node, diags
}

func ParseSanySyntaxModules(file, source string) ([]*SanySyntaxNode, Diagnostics) {
	tokens, lexDiags := SanyTokenize(file, source)
	parser := NewSanyParser(tokens, nil)
	if parser.match(SanyTokenBeginPragma) {
		for !parser.check(SanyTokenEOF) && !parser.atModuleStart() {
			parser.advance()
		}
	}
	var modules []*SanySyntaxNode
	for !parser.check(SanyTokenEOF) {
		for !parser.check(SanyTokenEOF) && !parser.atModuleStart() {
			parser.advance()
		}
		if parser.check(SanyTokenEOF) {
			break
		}
		modules = append(modules, parser.Module())
	}
	diags := filterSanyDiagnosticsToModuleSpans(lexDiags, modules)
	diags = append(diags, parser.diags...)
	return modules, diags
}

func filterSanyDiagnosticsToModuleSpans(diags Diagnostics, modules []*SanySyntaxNode) Diagnostics {
	if len(diags) == 0 || len(modules) == 0 {
		return diags
	}
	out := make(Diagnostics, 0, len(diags))
	for _, diag := range diags {
		if diag.Pos.Line == 0 {
			out = append(out, diag)
			continue
		}
		for _, mod := range modules {
			if mod != nil && positionInSanyRange(diag.Pos, mod.Range) {
				out = append(out, diag)
				break
			}
		}
	}
	return out
}

func positionInSanyRange(pos Position, rng SanyRange) bool {
	if rng.Begin.Line == 0 || rng.End.Line == 0 {
		return false
	}
	return !positionBefore(pos, rng.Begin) && !positionBefore(rng.End, pos)
}

func filterSanyDiagnosticsThroughRootEnd(diags Diagnostics, root *SanySyntaxNode) Diagnostics {
	if len(diags) == 0 {
		return diags
	}
	end := sanyRootEndModulePosition(root)
	if end.Line == 0 {
		return diags
	}
	out := make(Diagnostics, 0, len(diags))
	for _, diag := range diags {
		if diag.Pos.Line == 0 || !positionBefore(end, diag.Pos) {
			out = append(out, diag)
		}
	}
	return out
}

func sanyRootEndModulePosition(root *SanySyntaxNode) Position {
	if root == nil || root.Kind.JavaName() != "N_Module" {
		return Position{}
	}
	heirs := root.GetHeirs()
	if len(heirs) == 0 {
		return Position{}
	}
	last := heirs[len(heirs)-1]
	if last == nil || last.Kind.JavaName() != "N_EndModule" {
		return Position{}
	}
	return last.Range.End
}

func NewSanyParser(tokens []*SanyToken, diags Diagnostics) *SanyParser {
	return &SanyParser{tokens: tokens, diags: diags}
}

func SanyModuleName(root *SanySyntaxNode) string {
	if root == nil || root.Kind.JavaName() != "N_Module" {
		return ""
	}
	heirs := root.GetHeirs()
	if len(heirs) == 0 {
		return ""
	}
	begin := heirs[0].GetHeirs()
	if len(begin) < 2 {
		return ""
	}
	return begin[1].Image
}

func (p *SanyParser) CompilationUnit() *SanySyntaxNode {
	if p.match(SanyTokenBeginPragma) {
		for !p.check(SanyTokenEOF) && !p.check(SanyTokenBm2) {
			p.advance()
		}
	}
	return p.Module()
}

func (p *SanyParser) atModuleStart() bool {
	return p.check(SanyTokenBm0) || p.check(SanyTokenBm1) || p.check(SanyTokenBm2)
}

func (p *SanyParser) Module() *SanySyntaxNode {
	begin := p.BeginModule()
	extends := p.Extends()
	body := p.Body()
	end := p.EndModule()
	return NewSanyNode(SanySyntaxNodeKindByName["N_Module"], begin, extends, body, end)
}

func (p *SanyParser) BeginModule() *SanySyntaxNode {
	begin := p.consumeAny([]SanyTokenKind{SanyTokenBm0, SanyTokenBm1, SanyTokenBm2}, "expected ---- MODULE")
	name := p.Identifier()
	if name != nil {
		p.moduleName = name.Image
	}
	separator := p.consume(SanyTokenSeparator, "expected ---- after module name")
	return NewSanyNode(SanySyntaxNodeKindByName["N_BeginModule"], begin, name, separator)
}

func (p *SanyParser) EndModule() *SanySyntaxNode {
	end := p.consume(SanyTokenEndModule, "expected ==== at end of module")
	return NewSanyNode(SanySyntaxNodeKindByName["N_EndModule"], end)
}

func (p *SanyParser) Extends() *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	if p.match(SanyTokenExtends) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
		heirs = append(heirs, p.Identifier())
		for p.match(SanyTokenComma) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()))
			heirs = append(heirs, p.Identifier())
		}
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_Extends"], heirs...)
}

func (p *SanyParser) Body() *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	for !p.check(SanyTokenEOF) && !p.check(SanyTokenEndModule) {
		switch {
		case p.check(SanyTokenBm0) || p.check(SanyTokenBm1) || p.check(SanyTokenBm2):
			heirs = append(heirs, p.Module())
		case p.check(SanyTokenSeparator):
			heirs = append(heirs, NewSanyTokenNode(p.advance()))
		case p.check(SanyTokenVariable):
			heirs = append(heirs, p.VariableDeclaration())
		case p.check(SanyTokenConstant):
			heirs = append(heirs, p.ParamDeclaration())
		case p.check(SanyTokenRecursive):
			heirs = append(heirs, p.Recursive())
		case p.check(SanyTokenInstance) || (p.check(SanyTokenLocal) && p.peekNext().Kind == SanyTokenInstance):
			heirs = append(heirs, p.Instance())
		case p.check(SanyTokenAssume) || p.check(SanyTokenAssumption):
			heirs = append(heirs, p.Assumption())
		case p.check(SanyTokenTheorem) || p.check(SanyTokenProposition):
			heirs = append(heirs, p.Theorem())
		case p.check(SanyTokenUse) || p.check(SanyTokenHide):
			heirs = append(heirs, p.UseOrHide())
		case p.startsOperatorOrFunctionDefinition():
			heirs = append(heirs, p.OperatorOrFunctionDefinition())
		default:
			heirs = append(heirs, NewSanyTokenNode(p.advance()))
		}
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_Body"], heirs...)
}

func (p *SanyParser) VariableDeclaration() *SanySyntaxNode {
	keyword := p.consume(SanyTokenVariable, "expected VARIABLE declaration")
	var one []*SanySyntaxNode
	one = append(one, p.Identifier())
	for p.match(SanyTokenComma) {
		one = append(one, NewSanyTokenNode(p.previous()))
		one = append(one, p.Identifier())
	}
	return NewSanySplitNode(SanySyntaxNodeKindByName["N_VariableDeclaration"], []*SanySyntaxNode{keyword}, one)
}

func (p *SanyParser) ParamDeclaration() *SanySyntaxNode {
	keyword := p.consume(SanyTokenConstant, "expected CONSTANT declaration")
	heirs := []*SanySyntaxNode{NewSanyNode(SanySyntaxNodeKindByName["N_ConsDecl"], keyword)}
	heirs = append(heirs, p.ConstantDeclarationItem())
	for p.match(SanyTokenComma) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
		heirs = append(heirs, p.ConstantDeclarationItem())
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_ParamDeclaration"], heirs...)
}

func (p *SanyParser) ConstantDeclarationItem() *SanySyntaxNode {
	if p.check(SanyTokenIdentifier) {
		return p.IdentDecl()
	}
	return p.SomeFixDecl()
}

func (p *SanyParser) Recursive() *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	heirs = append(heirs, p.consume(SanyTokenRecursive, "expected RECURSIVE"))
	heirs = append(heirs, p.ConstantDeclarationItem())
	for p.match(SanyTokenComma) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
		heirs = append(heirs, p.ConstantDeclarationItem())
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_Recursive"], heirs...)
}

func (p *SanyParser) Instance() *SanySyntaxNode {
	var zero []*SanySyntaxNode
	if p.match(SanyTokenLocal) {
		zero = append(zero, NewSanyTokenNode(p.previous()))
	}
	inst := p.Instantiation()
	return NewSanySplitNode(SanySyntaxNodeKindByName["N_Instance"], zero, []*SanySyntaxNode{inst})
}

func (p *SanyParser) Instantiation() *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	heirs = append(heirs, p.consume(SanyTokenInstance, "expected INSTANCE"))
	heirs = append(heirs, p.Identifier())
	if p.match(SanyTokenWith) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
		heirs = append(heirs, p.Substitution())
		for p.match(SanyTokenComma) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()))
			heirs = append(heirs, p.Substitution())
		}
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_NonLocalInstance"], heirs...)
}

func (p *SanyParser) Substitution() *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	if p.check(SanyTokenIdentifier) {
		heirs = append(heirs, p.Identifier())
	} else {
		heirs = append(heirs, p.consumeOperator("expected substitution target"))
	}
	heirs = append(heirs, p.consume(SanyTokenSubstitute, "expected <- in substitution"))
	heirs = append(heirs, p.ExpressionUntilCommaOrBodyBoundary())
	return NewSanyNode(SanySyntaxNodeKindByName["N_Substitution"], heirs...)
}

func (p *SanyParser) Assumption() *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	if p.match(SanyTokenAssume) || p.match(SanyTokenAssumption) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
	} else {
		heirs = append(heirs, p.consume(SanyTokenAssume, "expected ASSUME or ASSUMPTION"))
	}
	if p.match(SanyTokenDefbreak) {
	}
	if p.check(SanyTokenIdentifier) && p.peekNext().Kind == SanyTokenDef {
		heirs = append(heirs, p.Identifier())
		heirs = append(heirs, p.consume(SanyTokenDef, "expected == in assumption"))
	}
	heirs = append(heirs, p.ExpressionUntilBodyBoundary())
	return NewSanySplitNode(SanySyntaxNodeKindByName["N_Assumption"], nil, heirs)
}

func (p *SanyParser) Theorem() *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	if p.match(SanyTokenTheorem) || p.match(SanyTokenProposition) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
	} else {
		heirs = append(heirs, p.consume(SanyTokenTheorem, "expected THEOREM or PROPOSITION"))
	}
	if p.check(SanyTokenIdentifier) && p.peekNext().Kind == SanyTokenDef {
		heirs = append(heirs, p.Identifier())
		heirs = append(heirs, p.consume(SanyTokenDef, "expected == in theorem"))
	}
	if p.startsAssumeProveAt(0) {
		heirs = append(heirs, p.AssumeProve())
	} else {
		heirs = append(heirs, p.ExpressionUntilDefinitionBoundary(func(tok *SanyToken) bool {
			return beginsSanyProof(tok)
		}))
	}
	if beginsSanyProof(p.peek()) {
		heirs = append(heirs, p.Proof())
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_Theorem"], heirs...)
}

func (p *SanyParser) Proof() *SanySyntaxNode {
	p.pushProofLevel()
	defer p.popProofLevel()
	var heirs []*SanySyntaxNode
	if p.match(SanyTokenProof) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
	}
	if p.match(SanyTokenObvious) || p.match(SanyTokenOmitted) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
		return NewSanyNode(SanySyntaxNodeKindByName["N_TerminalProof"], heirs...)
	}
	if p.match(SanyTokenBy) {
		by := p.previous()
		heirs = append(heirs, NewSanyTokenNode(by))
		p.proofCommandTail(&heirs, by, true)
		return NewSanyNode(SanySyntaxNodeKindByName["N_TerminalProof"], heirs...)
	}
	if p.check(SanyTokenQed) || p.startsProofStepAt(0) {
		for !p.check(SanyTokenEOF) && !p.check(SanyTokenEndModule) && !p.startsQEDStep() {
			heirs = append(heirs, p.Step())
		}
		heirs = append(heirs, p.QEDStep())
		return NewSanyNode(SanySyntaxNodeKindByName["N_Proof"], heirs...)
	}
	p.add(p.peek().Begin, "E1300", "expected terminal proof")
	return NewSanyNode(SanySyntaxNodeKindByName["N_TerminalProof"], heirs...)
}

func (p *SanyParser) QEDStep() *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	if p.startsProofStepAt(0) {
		heirs = append(heirs, p.StepStartToken())
	}
	qed := p.consume(SanyTokenQed, "expected QED")
	qedStep := NewSanyNode(SanySyntaxNodeKindByName["N_QEDStep"], qed)
	heirs = append(heirs, qedStep)
	if p.beginsProofAt(0) {
		heirs = append(heirs, p.Proof())
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_ProofStep"], heirs...)
}

func (p *SanyParser) Step() *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	heirs = append(heirs, p.StepStartToken())
	var body *SanySyntaxNode
	mayHaveProof := false
	switch {
	case p.check(SanyTokenUse) || p.check(SanyTokenHide):
		body = p.UseOrHide()
	case p.check(SanyTokenInstance):
		body = p.Instantiation()
	case p.check(SanyTokenDefbreak) || p.check(SanyTokenDefine) || p.startsOperatorOrFunctionDefinition():
		body = p.DefStep()
	case p.check(SanyTokenHave):
		body = p.HaveStep()
	case p.check(SanyTokenTake):
		body = p.TakeStep()
	case p.check(SanyTokenWitness):
		body = p.WitnessStep()
	case p.check(SanyTokenPick):
		body = p.PickStep()
		mayHaveProof = true
	case p.check(SanyTokenCase):
		body = p.CaseStep()
		mayHaveProof = true
	default:
		body = p.AssertStep()
		mayHaveProof = true
	}
	heirs = append(heirs, body)
	if p.beginsProofAt(0) {
		if mayHaveProof {
			heirs = append(heirs, p.Proof())
		} else {
			p.add(p.peek().Begin, "E1300", "proof of step that does not take a proof")
		}
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_ProofStep"], heirs...)
}

func (p *SanyParser) UseOrHide() *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	if p.match(SanyTokenUse) || p.match(SanyTokenHide) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
	} else {
		heirs = append(heirs, p.consume(SanyTokenUse, "expected USE or HIDE"))
	}
	command := p.previous()
	if p.match(SanyTokenOnly) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
	}
	if !p.atUseOrHideBoundary(command) && !p.check(SanyTokenDF) {
		heirs = append(heirs, p.UseOrHideItem())
		for p.match(SanyTokenComma) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()))
			heirs = append(heirs, p.UseOrHideItem())
		}
	}
	if p.match(SanyTokenDF) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
		heirs = append(heirs, p.UseOrHideItem())
		for p.match(SanyTokenComma) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()))
			heirs = append(heirs, p.UseOrHideItem())
		}
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_UseOrHide"], heirs...)
}

func (p *SanyParser) proofCommandTail(heirs *[]*SanySyntaxNode, command *SanyToken, terminal bool) {
	if p.match(SanyTokenOnly) {
		*heirs = append(*heirs, NewSanyTokenNode(p.previous()))
	}
	if !p.atProofCommandBoundary(command, terminal) && !p.check(SanyTokenDF) {
		*heirs = append(*heirs, p.ProofCommandItem(command, terminal))
		for p.match(SanyTokenComma) {
			*heirs = append(*heirs, NewSanyTokenNode(p.previous()))
			if p.atProofCommandBoundary(command, terminal) || p.check(SanyTokenDF) {
				break
			}
			*heirs = append(*heirs, p.ProofCommandItem(command, terminal))
		}
	}
	if p.match(SanyTokenDF) {
		*heirs = append(*heirs, NewSanyTokenNode(p.previous()))
		if !p.atProofCommandBoundary(command, terminal) {
			*heirs = append(*heirs, p.ProofCommandItem(command, terminal))
			for p.match(SanyTokenComma) {
				*heirs = append(*heirs, NewSanyTokenNode(p.previous()))
				if p.atProofCommandBoundary(command, terminal) {
					break
				}
				*heirs = append(*heirs, p.ProofCommandItem(command, terminal))
			}
		}
	}
}

func (p *SanyParser) ProofCommandItem(command *SanyToken, terminal bool) *SanySyntaxNode {
	if p.match(SanyTokenModule) {
		module := NewSanyTokenNode(p.previous())
		return NewSanyNode(SanySyntaxNodeKindByName["N_ModuleDefinition"], module, p.Identifier())
	}
	if p.startsProofStepAt(0) {
		return NewSanyTokenNode(p.advance())
	}
	if p.startsBareProofCommandOperatorReference(command, terminal) {
		return p.OperatorReference()
	}
	return p.ExpressionUntilProofCommandItemBoundary(command, terminal)
}

func (p *SanyParser) ExpressionUntilProofCommandItemBoundary(command *SanyToken, terminal bool) *SanySyntaxNode {
	return p.ExpressionUntil(func(tok *SanyToken) bool {
		return tok.Kind == SanyTokenComma ||
			tok.Kind == SanyTokenDF ||
			tok.Kind == SanyTokenEOF ||
			tok.Kind == SanyTokenEndModule ||
			p.atProofCommandBoundary(command, terminal)
	})
}

func (p *SanyParser) atProofCommandBoundary(command *SanyToken, terminal bool) bool {
	if terminal {
		if command == nil {
			return p.check(SanyTokenEOF) || p.check(SanyTokenEndModule)
		}
		return p.atTerminalProofBoundary(command.Begin.Column)
	}
	return p.atUseOrHideBoundary(command)
}

func (p *SanyParser) UseOrHideItem() *SanySyntaxNode {
	if p.match(SanyTokenModule) {
		module := NewSanyTokenNode(p.previous())
		return NewSanyNode(SanySyntaxNodeKindByName["N_ModuleDefinition"], module, p.Identifier())
	}
	if p.startsBareProofCommandOperatorReference(nil, false) {
		return p.OperatorReference()
	}
	return p.ExpressionUntilUseOrHideItemBoundary()
}

func (p *SanyParser) startsBareProofCommandOperatorReference(command *SanyToken, terminal bool) bool {
	if _, ok := GetSanyOperator(p.peek().Image); !ok {
		return false
	}
	next := p.tokenAt(1)
	switch next.Kind {
	case SanyTokenComma, SanyTokenDF, SanyTokenEOF, SanyTokenEndModule:
		return true
	case SanyTokenQed:
		return terminal && command != nil && next.Begin.Column <= command.Begin.Column
	}
	if p.startsBodyItemAt(1) {
		return true
	}
	if terminal {
		return p.startsProofStepAt(1) && command != nil && next.Begin.Column <= command.Begin.Column
	}
	return beginsExplicitSanyProof(next) || p.startsProofStepAt(1)
}

func (p *SanyParser) ExpressionUntilUseOrHideItemBoundary() *SanySyntaxNode {
	return p.ExpressionUntil(func(tok *SanyToken) bool {
		return tok.Kind == SanyTokenComma || tok.Kind == SanyTokenDF || p.isProofBoundary(tok)
	})
}

func (p *SanyParser) DefStep() *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	if p.match(SanyTokenDefbreak) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
	}
	if p.match(SanyTokenDefine) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
	}
	for p.startsOperatorOrFunctionDefinition() {
		heirs = append(heirs, p.ProofOperatorOrFunctionDefinition())
	}
	if len(heirs) == 0 || !p.atProofBoundary() {
		p.add(p.peek().Begin, "E1300", "expected proof definition")
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_DefStep"], heirs...)
}

func (p *SanyParser) ProofOperatorOrFunctionDefinition() *SanySyntaxNode {
	if p.match(SanyTokenLocal) {
		local := NewSanyTokenNode(p.previous())
		node := p.ProofOperatorOrFunctionDefinition()
		if node != nil {
			node.Zero = []*SanySyntaxNode{local}
			node.refreshHeirsAndRange()
		}
		return node
	}
	if p.match(SanyTokenDefbreak) {
	}
	switch {
	case p.check(SanyTokenIdentifier) && p.peekNext().Kind == SanyTokenLsb:
		return p.ProofFunctionDefinition()
	case p.check(SanyTokenIdentifier) && p.isPostfixOperator(p.peekNext()):
		return p.ProofOperatorDefinition(p.PostfixLHS())
	case p.check(SanyTokenIdentifier) && p.isInfixOperator(p.peekNext()) && p.tokenAt(2).Kind == SanyTokenIdentifier:
		return p.ProofOperatorDefinition(p.InfixLHS())
	case p.check(SanyTokenIdentifier) && p.peekNext().Kind == SanyTokenDef && p.tokenAt(2).Kind == SanyTokenInstance:
		return p.ModuleDefinition()
	case p.check(SanyTokenIdentifier) && (p.peekNext().Kind == SanyTokenLbr || p.peekNext().Kind == SanyTokenDef):
		return p.ProofOperatorDefinition(p.IdentLHS())
	default:
		return p.ProofOperatorDefinition(p.PrefixLHS())
	}
}

func (p *SanyParser) ProofFunctionDefinition() *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	heirs = append(heirs, p.Identifier())
	heirs = append(heirs, p.consume(SanyTokenLsb, "expected [ in function definition"))
	heirs = append(heirs, p.QuantBound())
	for p.match(SanyTokenComma) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
		heirs = append(heirs, p.QuantBound())
	}
	heirs = append(heirs, p.consume(SanyTokenRsb, "expected ] in function definition"))
	heirs = append(heirs, p.consume(SanyTokenDef, "expected == in function definition"))
	heirs = append(heirs, p.ExpressionUntilProofBoundary())
	return NewSanySplitNode(SanySyntaxNodeKindByName["N_FunctionDefinition"], nil, heirs)
}

func (p *SanyParser) ProofOperatorDefinition(lhs *SanySyntaxNode) *SanySyntaxNode {
	heirs := []*SanySyntaxNode{lhs}
	heirs = append(heirs, p.consume(SanyTokenDef, "expected == in operator definition"))
	heirs = append(heirs, p.ExpressionUntilProofBoundary())
	return NewSanySplitNode(SanySyntaxNodeKindByName["N_OperatorDefinition"], nil, heirs)
}

func (p *SanyParser) HaveStep() *SanySyntaxNode {
	have := p.consume(SanyTokenHave, "expected HAVE")
	expr := p.ExpressionUntilProofBoundary()
	return NewSanyNode(SanySyntaxNodeKindByName["N_HaveStep"], have, expr)
}

func (p *SanyParser) TakeStep() *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	heirs = append(heirs, p.consume(SanyTokenTake, "expected TAKE"))
	if p.takeUsesQuantBounds() {
		heirs = append(heirs, p.ProofQuantBoundUntil(SanyTokenComma, SanyTokenEOF))
		for p.match(SanyTokenComma) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()))
			heirs = append(heirs, p.ProofQuantBoundUntil(SanyTokenComma, SanyTokenEOF))
		}
	} else {
		heirs = append(heirs, p.Identifier())
		for p.match(SanyTokenComma) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()))
			heirs = append(heirs, p.Identifier())
		}
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_TakeStep"], heirs...)
}

func (p *SanyParser) WitnessStep() *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	heirs = append(heirs, p.consume(SanyTokenWitness, "expected WITNESS"))
	heirs = append(heirs, p.ExpressionUntilCommaOrProofBoundary())
	for p.match(SanyTokenComma) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
		heirs = append(heirs, p.ExpressionUntilCommaOrProofBoundary())
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_WitnessStep"], heirs...)
}

func (p *SanyParser) PickStep() *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	heirs = append(heirs, p.consume(SanyTokenPick, "expected PICK"))
	if p.pickUsesIdentifierList() {
		heirs = append(heirs, p.IdentDecl())
		for p.match(SanyTokenComma) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()))
			heirs = append(heirs, p.IdentDecl())
		}
	} else {
		heirs = append(heirs, p.ProofQuantBoundUntil(SanyTokenComma, SanyTokenColon, SanyTokenEOF))
		for p.match(SanyTokenComma) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()))
			heirs = append(heirs, p.ProofQuantBoundUntil(SanyTokenComma, SanyTokenColon, SanyTokenEOF))
		}
	}
	heirs = append(heirs, p.consume(SanyTokenColon, "expected : in PICK step"))
	heirs = append(heirs, p.ExpressionUntilProofBoundary())
	return NewSanyNode(SanySyntaxNodeKindByName["N_PickStep"], heirs...)
}

func (p *SanyParser) CaseStep() *SanySyntaxNode {
	caseTok := p.consume(SanyTokenCase, "expected CASE")
	expr := p.ExpressionUntilProofBoundary()
	return NewSanyNode(SanySyntaxNodeKindByName["N_CaseStep"], caseTok, expr)
}

func (p *SanyParser) AssertStep() *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	if p.match(SanyTokenSuffices) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
	}
	if p.startsAssumeProveAt(0) {
		heirs = append(heirs, p.AssumeProve())
	} else {
		heirs = append(heirs, p.ExpressionUntilProofBoundary())
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_AssertStep"], heirs...)
}

func (p *SanyParser) AssumeProve() *SanySyntaxNode {
	return p.assumeProveUntil(func(tok *SanyToken) bool {
		return p.isProofBoundary(tok)
	})
}

func (p *SanyParser) assumeProveItem() *SanySyntaxNode {
	return p.assumeProveUntil(func(tok *SanyToken) bool {
		return tok.Kind == SanyTokenComma ||
			tok.Kind == SanyTokenProve ||
			tok.Kind == SanyTokenBoxprove ||
			p.isProofBoundary(tok)
	})
}

func (p *SanyParser) assumeProveUntil(proveStop func(*SanyToken) bool) *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	if p.match(SanyTokenAssume) || p.match(SanyTokenBoxassume) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
	} else {
		heirs = append(heirs, p.consume(SanyTokenAssume, "expected ASSUME"))
	}
	heirs = append(heirs, p.AssumeProveItem())
	for p.match(SanyTokenComma) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
		heirs = append(heirs, p.AssumeProveItem())
	}
	if p.match(SanyTokenProve) || p.match(SanyTokenBoxprove) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
	} else {
		heirs = append(heirs, p.consume(SanyTokenProve, "expected PROVE"))
	}
	heirs = append(heirs, p.ExpressionUntil(proveStop))
	return NewSanyNode(SanySyntaxNodeKindByName["N_AssumeProve"], heirs...)
}

func (p *SanyParser) AssumeProveItem() *SanySyntaxNode {
	if p.startsLabeledAssumeProveAt(0) {
		label := p.LabelName()
		colon := p.consume(SanyTokenColoncolon, "expected :: after label")
		return NewSanyNode(SanySyntaxNodeKindByName["N_Label"], label, colon, p.assumeProveItem())
	}
	if p.startsAssumeProveAt(0) {
		return p.assumeProveItem()
	}
	if p.startsNewSymbAt(0) {
		return p.NewSymb()
	}
	return p.ExpressionUntilAssumeProveBoundary()
}

func (p *SanyParser) ProofQuantBoundUntil(stopKinds ...SanyTokenKind) *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	heirs = append(heirs, p.QuantBoundIntro())
	for p.match(SanyTokenComma) {
		if p.findBeforeStopIgnoring(SanyTokenIN, SanyTokenComma, stopKinds...) < 0 {
			p.at--
			break
		}
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
		heirs = append(heirs, p.QuantBoundIntro())
	}
	in := p.consume(SanyTokenIN, "expected \\in in quantifier bound")
	if in != nil {
		in.Kind = SanySyntaxNodeKindByName["T_IN"]
	}
	heirs = append(heirs, in)
	heirs = append(heirs, p.ExpressionUntil(func(tok *SanyToken) bool {
		for _, stop := range stopKinds {
			if tok.Kind == stop {
				return true
			}
		}
		return p.isProofBoundary(tok)
	}))
	return NewSanyNode(SanySyntaxNodeKindByName["N_QuantBound"], heirs...)
}

func (p *SanyParser) NewSymb() *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	if (p.check(SanyTokenNew) && p.tokenAt(1).Kind == SanyTokenVariable) || p.check(SanyTokenVariable) {
		if p.match(SanyTokenNew) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()))
		}
		heirs = append(heirs, p.consume(SanyTokenVariable, "expected VARIABLE"))
		heirs = append(heirs, NewSanyNode(SanySyntaxNodeKindByName["N_IdentDecl"], p.Identifier()))
		return NewSanyNode(SanySyntaxNodeKindByName["N_NewSymb"], heirs...)
	}
	if (p.check(SanyTokenNew) && isSanyStateActionTemporal(p.tokenAt(1).Kind)) || isSanyStateActionTemporal(p.peek().Kind) {
		if p.match(SanyTokenNew) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()))
		}
		heirs = append(heirs, NewSanyTokenNode(p.advance()))
		heirs = append(heirs, p.IdentDeclOrSomeFixDecl())
		return NewSanyNode(SanySyntaxNodeKindByName["N_NewSymb"], heirs...)
	}
	if p.match(SanyTokenNew) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
	}
	if p.match(SanyTokenConstant) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
	}
	heirs = append(heirs, p.IdentDeclOrSomeFixDecl())
	if p.match(SanyTokenIN) {
		in := NewSanyTokenNode(p.previous())
		in.Kind = SanySyntaxNodeKindByName["T_IN"]
		heirs = append(heirs, in)
		heirs = append(heirs, p.ExpressionUntilAssumeProveBoundary())
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_NewSymb"], heirs...)
}

func (p *SanyParser) StepStartToken() *SanySyntaxNode {
	if p.startsProofStepAt(0) {
		tok := p.advance()
		node := NewSanyTokenNode(tok)
		if !p.correctProofLevel(tok) {
			p.add(tok.Begin, "E1300", "proof step has bad level")
		}
		return node
	}
	p.add(p.peek().Begin, "E1300", "expected proof step")
	return nil
}

func (p *SanyParser) ExpressionUntilProofBoundary() *SanySyntaxNode {
	startLine := p.peek().Begin.Line
	return p.ExpressionUntil(func(tok *SanyToken) bool {
		return p.isProofBoundaryAfterExpressionStart(tok, startLine)
	})
}

func (p *SanyParser) ExpressionUntilCommaOrProofBoundary() *SanySyntaxNode {
	startLine := p.peek().Begin.Line
	return p.ExpressionUntil(func(tok *SanyToken) bool {
		return tok.Kind == SanyTokenComma || p.isProofBoundaryAfterExpressionStart(tok, startLine)
	})
}

func (p *SanyParser) ExpressionUntilAssumeProveBoundary() *SanySyntaxNode {
	startLine := p.peek().Begin.Line
	return p.ExpressionUntil(func(tok *SanyToken) bool {
		return tok.Kind == SanyTokenComma ||
			tok.Kind == SanyTokenProve ||
			tok.Kind == SanyTokenBoxprove ||
			p.isProofBoundaryAfterExpressionStart(tok, startLine)
	})
}

func (p *SanyParser) isProofBoundaryAfterExpressionStart(tok *SanyToken, startLine int) bool {
	if tok != nil && tok.Begin.Line == startLine && isSanyProofStepStartKind(tok.Kind) {
		return false
	}
	return p.isProofBoundary(tok)
}

func (p *SanyParser) atProofBoundary() bool {
	return p.isProofBoundary(p.peek())
}

func (p *SanyParser) atUseOrHideBoundary(command *SanyToken) bool {
	if command != nil && p.startsProofStepAt(0) {
		tok := p.peek()
		if tok.Begin.Line == command.Begin.Line || tok.Begin.Column > command.Begin.Column {
			return false
		}
	}
	return p.atProofBoundary()
}

func (p *SanyParser) isProofBoundary(tok *SanyToken) bool {
	return tok.Kind == SanyTokenEOF ||
		tok.Kind == SanyTokenEndModule ||
		beginsExplicitSanyProof(tok) ||
		p.startsProofStepAt(0) ||
		tok.Kind == SanyTokenQed ||
		p.startsBodyItemAt(0)
}

func (p *SanyParser) startsQEDStep() bool {
	if p.check(SanyTokenQed) {
		return true
	}
	return p.startsProofStepAt(0) && p.tokenAt(1).Kind == SanyTokenQed
}

func (p *SanyParser) atTerminalProofBoundary(startColumn int) bool {
	if p.startsBodyItem() {
		return true
	}
	if p.check(SanyTokenQed) && p.peek().Begin.Column <= startColumn {
		return true
	}
	return p.startsProofStepAt(0) && p.peek().Begin.Column <= startColumn
}

func (p *SanyParser) startsProofStepAt(offset int) bool {
	return isSanyProofStepStartKind(p.tokenAt(offset).Kind)
}

func (p *SanyParser) takeUsesQuantBounds() bool {
	return p.check(SanyTokenLab) || (p.check(SanyTokenIdentifier) && p.peekNext().Kind == SanyTokenIN)
}

func (p *SanyParser) pickUsesIdentifierList() bool {
	offset := 0
	for {
		if p.tokenAt(offset).Kind != SanyTokenIdentifier {
			return false
		}
		offset++
		if p.tokenAt(offset).Kind == SanyTokenComma {
			offset++
			continue
		}
		return p.tokenAt(offset).Kind == SanyTokenColon
	}
}

func (p *SanyParser) startsAssumeProveAt(offset int) bool {
	switch p.tokenAt(offset).Kind {
	case SanyTokenAssume, SanyTokenBoxassume:
		return true
	default:
		return false
	}
}

func (p *SanyParser) startsLabeledAssumeProveAt(offset int) bool {
	return p.startsLabelAt(offset) && p.startsAssumeProveAt(offset+2)
}

func (p *SanyParser) startsNewSymbAt(offset int) bool {
	switch p.tokenAt(offset).Kind {
	case SanyTokenNew, SanyTokenConstant, SanyTokenVariable, SanyTokenState, SanyTokenAction, SanyTokenTemporal:
		return true
	default:
		return false
	}
}

func isSanyStateActionTemporal(kind SanyTokenKind) bool {
	switch kind {
	case SanyTokenState, SanyTokenAction, SanyTokenTemporal:
		return true
	default:
		return false
	}
}

func isSanyProofStepStartKind(kind SanyTokenKind) bool {
	switch kind {
	case SanyTokenProofsteplexeme, SanyTokenProofimplicitsteplexeme, SanyTokenProofstepdotlexeme, SanyTokenBarelevellexeme, SanyTokenUnnumberedsteplexeme:
		return true
	default:
		return false
	}
}

func beginsExplicitSanyProof(tok *SanyToken) bool {
	if tok == nil {
		return false
	}
	switch tok.Kind {
	case SanyTokenProof, SanyTokenBy, SanyTokenObvious, SanyTokenOmitted:
		return true
	default:
		return false
	}
}

func beginsSanyProof(tok *SanyToken) bool {
	if tok == nil {
		return false
	}
	switch tok.Kind {
	case SanyTokenProof, SanyTokenBy, SanyTokenObvious, SanyTokenOmitted, SanyTokenProofsteplexeme, SanyTokenProofimplicitsteplexeme, SanyTokenProofstepdotlexeme, SanyTokenBarelevellexeme, SanyTokenUnnumberedsteplexeme, SanyTokenQed:
		return true
	default:
		return false
	}
}

func (p *SanyParser) pushProofLevel() {
	p.proofLevelStack = append(p.proofLevelStack, -1)
}

func (p *SanyParser) popProofLevel() {
	if len(p.proofLevelStack) == 0 {
		return
	}
	p.proofLevelStack = p.proofLevelStack[:len(p.proofLevelStack)-1]
}

func (p *SanyParser) currentProofLevel() int {
	if len(p.proofLevelStack) == 0 {
		return -1
	}
	return p.proofLevelStack[len(p.proofLevelStack)-1]
}

func (p *SanyParser) beginsProofAt(offset int) bool {
	tok := p.tokenAt(offset)
	if tok == nil {
		return false
	}
	switch tok.Kind {
	case SanyTokenBy, SanyTokenProof, SanyTokenObvious, SanyTokenOmitted:
		return true
	}
	if !isSanyProofStepStartKind(tok.Kind) {
		return false
	}
	if len(tok.Image) >= 2 {
		switch tok.Image[1] {
		case '*':
			return len(p.proofLevelStack) == 0
		case '+':
			return true
		}
	}
	if len(p.proofLevelStack) == 0 {
		return true
	}
	level, ok := sanyProofStepLevel(tok)
	return ok && p.currentProofLevel() >= 0 && level > p.currentProofLevel()
}

func (p *SanyParser) correctProofLevel(tok *SanyToken) bool {
	if len(p.proofLevelStack) == 0 {
		return true
	}
	level, ok := sanyProofStepLevel(tok)
	if !ok {
		return true
	}
	idx := len(p.proofLevelStack) - 1
	lastLevel := -1
	if idx > 0 {
		lastLevel = p.proofLevelStack[idx-1]
	}
	switch level {
	case -1:
		if p.proofLevelStack[idx] < 0 {
			p.proofLevelStack[idx] = lastLevel + 1
		}
		return true
	case -2:
		if p.proofLevelStack[idx] < 0 {
			p.proofLevelStack[idx] = lastLevel + 1
			return true
		}
		return false
	default:
		if p.proofLevelStack[idx] < 0 {
			p.proofLevelStack[idx] = level
			return level > lastLevel
		}
		return p.proofLevelStack[idx] == level
	}
}

func sanyProofStepLevel(tok *SanyToken) (int, bool) {
	if tok == nil || !isSanyProofStepStartKind(tok.Kind) || len(tok.Image) < 3 || tok.Image[0] != '<' {
		return 0, false
	}
	switch tok.Image[1] {
	case '*':
		return -1, true
	case '+':
		return -2, true
	}
	end := strings.IndexByte(tok.Image, '>')
	if end <= 1 {
		return 0, false
	}
	level, err := strconv.Atoi(tok.Image[1:end])
	if err != nil {
		return 0, false
	}
	return level, true
}

func (p *SanyParser) OperatorOrFunctionDefinition() *SanySyntaxNode {
	if p.match(SanyTokenLocal) {
		local := NewSanyTokenNode(p.previous())
		node := p.OperatorOrFunctionDefinition()
		if node != nil {
			node.Zero = []*SanySyntaxNode{local}
			node.refreshHeirsAndRange()
		}
		return node
	}
	if p.match(SanyTokenDefbreak) {
	}
	switch {
	case p.check(SanyTokenIdentifier) && p.peekNext().Kind == SanyTokenLsb:
		return p.FunctionDefinition()
	case p.check(SanyTokenIdentifier) && p.isPostfixOperator(p.peekNext()):
		return p.OperatorDefinition(p.PostfixLHS())
	case p.check(SanyTokenIdentifier) && p.isInfixOperator(p.peekNext()) && p.tokenAt(2).Kind == SanyTokenIdentifier:
		return p.OperatorDefinition(p.InfixLHS())
	case p.startsModuleDefinitionHeadAt(0):
		return p.ModuleDefinition()
	case p.check(SanyTokenIdentifier) && (p.peekNext().Kind == SanyTokenLbr || p.peekNext().Kind == SanyTokenDef):
		return p.OperatorDefinition(p.IdentLHS())
	default:
		return p.OperatorDefinition(p.PrefixLHS())
	}
}

func (p *SanyParser) startsModuleDefinitionHeadAt(offset int) bool {
	if p.tokenAt(offset).Kind != SanyTokenIdentifier {
		return false
	}
	next := p.tokenAt(offset + 1)
	if next.Kind == SanyTokenDef {
		return p.tokenAt(offset+2).Kind == SanyTokenInstance
	}
	if next.Kind != SanyTokenLbr {
		return false
	}
	end := p.findMatchingBracketOffset(offset + 1)
	return end >= 0 && p.tokenAt(end+1).Kind == SanyTokenDef && p.tokenAt(end+2).Kind == SanyTokenInstance
}

func (p *SanyParser) ModuleDefinition() *SanySyntaxNode {
	lhs := p.IdentLHS()
	heirs := []*SanySyntaxNode{lhs}
	heirs = append(heirs, p.consume(SanyTokenDef, "expected == in module definition"))
	heirs = append(heirs, p.Instantiation())
	return NewSanySplitNode(SanySyntaxNodeKindByName["N_ModuleDefinition"], nil, heirs)
}

func (p *SanyParser) FunctionDefinition() *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	heirs = append(heirs, p.Identifier())
	heirs = append(heirs, p.consume(SanyTokenLsb, "expected [ in function definition"))
	heirs = append(heirs, p.QuantBound())
	for p.match(SanyTokenComma) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
		heirs = append(heirs, p.QuantBound())
	}
	heirs = append(heirs, p.consume(SanyTokenRsb, "expected ] in function definition"))
	heirs = append(heirs, p.consume(SanyTokenDef, "expected == in function definition"))
	heirs = append(heirs, p.ExpressionUntilBodyBoundary())
	return NewSanySplitNode(SanySyntaxNodeKindByName["N_FunctionDefinition"], nil, heirs)
}

func (p *SanyParser) OperatorDefinition(lhs *SanySyntaxNode) *SanySyntaxNode {
	heirs := []*SanySyntaxNode{lhs}
	heirs = append(heirs, p.consume(SanyTokenDef, "expected == in operator definition"))
	heirs = append(heirs, p.ExpressionUntilDefinitionBoundary(nil))
	return NewSanySplitNode(SanySyntaxNodeKindByName["N_OperatorDefinition"], nil, heirs)
}

func (p *SanyParser) LetOperatorDefinition(lhs *SanySyntaxNode) *SanySyntaxNode {
	heirs := []*SanySyntaxNode{lhs}
	heirs = append(heirs, p.consume(SanyTokenDef, "expected == in LET definition"))
	heirs = append(heirs, p.ExpressionUntilDefinitionBoundary(func(tok *SanyToken) bool {
		return tok.Kind == SanyTokenLetin
	}))
	return NewSanySplitNode(SanySyntaxNodeKindByName["N_OperatorDefinition"], nil, heirs)
}

func (p *SanyParser) IdentLHS() *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	heirs = append(heirs, p.Identifier())
	if p.match(SanyTokenLbr) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
		if !p.check(SanyTokenRbr) {
			heirs = append(heirs, p.IdentDeclOrSomeFixDecl())
			for p.match(SanyTokenComma) {
				heirs = append(heirs, NewSanyTokenNode(p.previous()))
				heirs = append(heirs, p.IdentDeclOrSomeFixDecl())
			}
		}
		heirs = append(heirs, p.consume(SanyTokenRbr, "expected ) in operator definition parameters"))
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_IdentLHS"], heirs...)
}

func (p *SanyParser) PrefixLHS() *SanySyntaxNode {
	op := p.consumeOperator("expected prefix operator in definition")
	id := p.Identifier()
	return NewSanyNode(SanySyntaxNodeKindByName["N_PrefixLHS"], op, id)
}

func (p *SanyParser) InfixLHS() *SanySyntaxNode {
	left := p.Identifier()
	op := p.consumeOperator("expected infix operator in definition")
	right := p.Identifier()
	return NewSanyNode(SanySyntaxNodeKindByName["N_InfixLHS"], left, op, right)
}

func (p *SanyParser) PostfixLHS() *SanySyntaxNode {
	left := p.Identifier()
	op := p.consumeOperator("expected postfix operator in definition")
	return NewSanyNode(SanySyntaxNodeKindByName["N_PostfixLHS"], left, op)
}

func (p *SanyParser) IdentDeclOrSomeFixDecl() *SanySyntaxNode {
	if p.check(SanyTokenIdentifier) {
		return p.IdentDecl()
	}
	return p.SomeFixDecl()
}

func (p *SanyParser) IdentDecl() *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	heirs = append(heirs, p.Identifier())
	if p.match(SanyTokenLbr) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
		heirs = append(heirs, p.consume(SanyTokenUs, "expected _ in operator parameter declaration"))
		for p.match(SanyTokenComma) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()))
			heirs = append(heirs, p.consume(SanyTokenUs, "expected _ in operator parameter declaration"))
		}
		heirs = append(heirs, p.consume(SanyTokenRbr, "expected ) in operator parameter declaration"))
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_IdentDecl"], heirs...)
}

func (p *SanyParser) SomeFixDecl() *SanySyntaxNode {
	if p.isNEPrefixOperator(p.peek()) {
		op := p.consumeOperator("expected prefix operator declaration")
		us := p.consume(SanyTokenUs, "expected _ after prefix operator declaration")
		return NewSanyNode(SanySyntaxNodeKindByName["N_PrefixDecl"], op, us)
	}
	left := p.consume(SanyTokenUs, "expected _ in operator declaration")
	if p.isInfixOperator(p.peek()) {
		op := p.consumeOperator("expected infix operator declaration")
		right := p.consume(SanyTokenUs, "expected _ after infix operator declaration")
		return NewSanyNode(SanySyntaxNodeKindByName["N_InfixDecl"], left, op, right)
	}
	op := p.consumeOperator("expected postfix operator declaration")
	return NewSanyNode(SanySyntaxNodeKindByName["N_PostfixDecl"], left, op)
}

func (p *SanyParser) QuantBound() *SanySyntaxNode {
	return p.QuantBoundUntil(SanyTokenComma, SanyTokenRsb, SanyTokenEOF)
}

func (p *SanyParser) QuantBoundUntil(stopKinds ...SanyTokenKind) *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	heirs = append(heirs, p.QuantBoundIntro())
	for p.match(SanyTokenComma) {
		if p.findBeforeStopIgnoring(SanyTokenIN, SanyTokenComma, stopKinds...) < 0 {
			p.at--
			break
		}
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
		heirs = append(heirs, p.QuantBoundIntro())
	}
	in := p.consume(SanyTokenIN, "expected \\in in quantifier bound")
	if in != nil {
		in.Kind = SanySyntaxNodeKindByName["T_IN"]
	}
	heirs = append(heirs, in)
	heirs = append(heirs, p.ExpressionUntil(func(tok *SanyToken) bool {
		for _, stop := range stopKinds {
			if tok.Kind == stop {
				return true
			}
		}
		return false
	}))
	return NewSanyNode(SanySyntaxNodeKindByName["N_QuantBound"], heirs...)
}

func (p *SanyParser) startsQuantBoundIntro() bool {
	return p.check(SanyTokenIdentifier) || p.check(SanyTokenLab)
}

func (p *SanyParser) QuantBoundIntro() *SanySyntaxNode {
	if p.check(SanyTokenLab) {
		return p.IdentifierTuple()
	}
	return p.Identifier()
}

func (p *SanyParser) IdentifierTuple() *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	heirs = append(heirs, p.consume(SanyTokenLab, "expected << in identifier tuple"))
	if !p.check(SanyTokenRab) {
		heirs = append(heirs, p.Identifier())
		for p.match(SanyTokenComma) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()))
			heirs = append(heirs, p.Identifier())
		}
	}
	heirs = append(heirs, p.consume(SanyTokenRab, "expected >> in identifier tuple"))
	return NewSanyNode(SanySyntaxNodeKindByName["N_IdentifierTuple"], heirs...)
}

func (p *SanyParser) ExpressionUntilBodyBoundary() *SanySyntaxNode {
	return p.ExpressionUntilDefinitionBoundary(nil)
}

func (p *SanyParser) ExpressionUntilDefinitionBoundary(extraStop func(*SanyToken) bool) *SanySyntaxNode {
	startLine := p.peek().Begin.Line
	return p.ExpressionUntil(func(tok *SanyToken) bool {
		if tok.Kind == SanyTokenEOF || tok.Kind == SanyTokenEndModule {
			return true
		}
		if extraStop != nil && extraStop(tok) {
			return true
		}
		return tok.Begin.Line >= startLine && p.startsBodyItemAt(0)
	})
}

func (p *SanyParser) ExpressionUntilCommaOrBodyBoundary() *SanySyntaxNode {
	startLine := p.peek().Begin.Line
	return p.ExpressionUntil(func(tok *SanyToken) bool {
		if tok.Kind == SanyTokenComma || tok.Kind == SanyTokenEOF || tok.Kind == SanyTokenEndModule {
			return true
		}
		return tok.Begin.Line >= startLine && p.startsBodyItemAt(0)
	})
}

func (p *SanyParser) ExpressionUntil(stop func(*SanyToken) bool) *SanySyntaxNode {
	stack := NewSanyOperatorStack()
	stack.NewStack()
	sawExpressionToken := false
	for {
		if sawExpressionToken && stop(p.peek()) && !(stack.PreInEmptyTop() && p.startsJunctionList(stop)) {
			break
		}
		if p.splitLeadingFairnessIdentifier() {
			continue
		}
		if stack.PreInEmptyTop() && p.startsJunctionList(stop) {
			stack.Push(p.JunctionList(stop), nil)
			sawExpressionToken = true
			continue
		}
		if !sawExpressionToken && p.startsOperatorReference(stop) {
			stack.Push(p.OperatorReference(), nil)
			sawExpressionToken = true
			continue
		}
		if p.startsLabelAt(0) {
			stack.Push(p.LabelExpression(stop), nil)
			sawExpressionToken = true
			continue
		}
		if p.startsNoOpExtension() {
			stack.Push(p.NoOpExtension(), nil)
			sawExpressionToken = true
			continue
		}
		if p.startsOpApplication() {
			stack.Push(p.OpApplication(), nil)
			sawExpressionToken = true
			continue
		}
		if sawExpressionToken && !stack.PreInEmptyTop() && p.check(SanyTokenLsb) {
			node := p.SBracketCases()
			op, _ := GetSanyOperator("[")
			stack.Push(node, &op)
			if err := stack.ReduceStack(); err != nil {
				p.add(p.previous().End, "E1301", err.Error())
				return stack.TopNode()
			}
			continue
		}
		if p.startsOpenExpression() {
			stack.Push(p.OpenExpression(stop), nil)
			sawExpressionToken = true
			continue
		}
		if p.startsPrimitiveExpression() {
			stack.Push(p.PrimitiveExpression(), nil)
			sawExpressionToken = true
			continue
		}
		if p.startsPrefixJunctionOperand() {
			stack.Push(p.PrefixJunctionExpression(stop), nil)
			sawExpressionToken = true
			continue
		}
		tok := p.advance()
		if op, ok := GetSanyOperator(tok.Image); ok {
			stack.Push(p.genericOperatorNode(tok, op), &op)
			if err := stack.ReduceStack(); err != nil {
				p.add(tok.Begin, "E1301", err.Error())
				return stack.TopNode()
			}
			continue
		}
		stack.Push(NewSanyTokenNode(tok), nil)
		sawExpressionToken = true
	}
	if !sawExpressionToken {
		p.add(p.peek().Begin, "E1302", "expected expression")
		return nil
	}
	expr, err := stack.FinalReduce()
	if err != nil {
		p.add(p.peek().Begin, "E1301", err.Error())
		return stack.TopNode()
	}
	return expr
}

func (p *SanyParser) startsJunctionList(stop func(*SanyToken) bool) bool {
	if !IsSanyJunctionBullet(p.peek().Kind) {
		return false
	}
	next := p.tokenAt(1)
	if next == nil {
		return false
	}
	switch next.Kind {
	case SanyTokenEOF, SanyTokenRbr, SanyTokenRbc, SanyTokenRsb, SanyTokenComma:
		return false
	default:
		if stop(next) && !(IsSanyJunctionBullet(next.Kind) && next.Kind != p.peek().Kind) {
			return false
		}
		return true
	}
}

func (p *SanyParser) startsPrefixJunctionOperand() bool {
	op, ok := GetSanyOperator(p.peek().Image)
	return ok && op.IsPrefix() && IsSanyJunctionBullet(p.peekNext().Kind)
}

func (p *SanyParser) PrefixJunctionExpression(stop func(*SanyToken) bool) *SanySyntaxNode {
	tok := p.advance()
	op, _ := GetSanyOperator(tok.Image)
	return NewSanyNode(SanySyntaxNodeKindByName["N_PrefixExpr"], p.genericOperatorNode(tok, op), p.JunctionList(stop))
}

func (p *SanyParser) JunctionList(stop func(*SanyToken) bool) *SanySyntaxNode {
	kind := p.peek().Kind
	firstBullet := p.advance()
	minColumn := firstBullet.Begin.Column
	itemStop := func(tok *SanyToken) bool {
		if tok.Kind == kind && tok.Begin.Line > firstBullet.Begin.Line && tok.Begin.Column == minColumn {
			return true
		}
		if IsSanyJunctionBullet(tok.Kind) && tok.Begin.Line > firstBullet.Begin.Line && tok.Begin.Column < minColumn {
			return true
		}
		if stop(tok) {
			return true
		}
		return tok.Begin.Line > firstBullet.Begin.Line && tok.Begin.Column <= minColumn && !IsSanyJunctionBullet(tok.Kind)
	}
	left := p.ExpressionUntil(func(tok *SanyToken) bool {
		return itemStop(tok)
	})
	sawMore := false
	for p.check(kind) && p.peek().Begin.Line > firstBullet.Begin.Line && p.peek().Begin.Column == minColumn {
		sawMore = true
		bullet := p.advance()
		right := p.ExpressionUntil(func(tok *SanyToken) bool {
			return itemStop(tok)
		})
		infix := NewSanyNode(SanySyntaxNodeKindByName["N_InfixExpr"], left, p.junctionOperatorNode(bullet), right)
		infix.JunctionList = true
		infix.Range.Begin = firstBullet.Begin
		left = infix
	}
	if !sawMore {
		return NewSanyNode(SanySyntaxNodeKindByName["N_PrefixExpr"], p.junctionOperatorNode(firstBullet), left)
	}
	return left
}

func (p *SanyParser) startsOperatorReference(stop func(*SanyToken) bool) bool {
	op, ok := GetSanyOperator(p.peek().Image)
	if !ok {
		return false
	}
	if op.Symbol == "[" {
		return false
	}
	if op.Symbol == "-" && p.peekNext().Kind == SanyTokenLbr {
		return false
	}
	return stop(p.peekNext()) || ((op.IsInfix() || op.IsPostfix()) && p.peekNext().Kind == SanyTokenLbr)
}

func (p *SanyParser) OperatorReference() *SanySyntaxNode {
	tok := p.advance()
	op, ok := GetSanyOperator(tok.Image)
	if !ok {
		return NewSanyTokenNode(tok)
	}
	node := p.genericOperatorNode(tok, op)
	if (op.IsInfix() || op.IsPostfix()) && p.check(SanyTokenLbr) {
		node.AddHeir(p.OpArgs())
	}
	return node
}

func (p *SanyParser) LabelExpression(stop func(*SanyToken) bool) *SanySyntaxNode {
	label := p.LabelName()
	colon := p.consume(SanyTokenColoncolon, "expected :: after label")
	expr := p.ExpressionUntil(stop)
	return NewSanyNode(SanySyntaxNodeKindByName["N_Label"], label, colon, expr)
}

func (p *SanyParser) LabelName() *SanySyntaxNode {
	heirs := []*SanySyntaxNode{
		NewSanyNode(SanySyntaxNodeKindByName["N_IdPrefix"]),
		p.Identifier(),
	}
	if p.check(SanyTokenLbr) {
		heirs = append(heirs, p.OpArgs())
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_GeneralId"], heirs...)
}

func (p *SanyParser) junctionOperatorNode(tok *SanyToken) *SanySyntaxNode {
	if op, ok := GetSanyOperator(tok.Image); ok {
		return p.genericOperatorNode(tok, op)
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_GenInfixOp"], NewSanyNode(SanySyntaxNodeKindByName["N_IdPrefix"]), NewSanyTokenNode(tok))
}

func (p *SanyParser) startsOpenExpression() bool {
	switch p.peek().Kind {
	case SanyTokenIf, SanyTokenForall, SanyTokenExists, SanyTokenTExists, SanyTokenTForall, SanyTokenLet, SanyTokenCase, SanyTokenChoose, SanyTokenLambda, SanyTokenWF, SanyTokenSF:
		return true
	default:
		return false
	}
}

func (p *SanyParser) OpenExpression(stop func(*SanyToken) bool) *SanySyntaxNode {
	switch p.peek().Kind {
	case SanyTokenIf:
		return p.IfThenElse(stop)
	case SanyTokenForall, SanyTokenExists:
		return p.SomeQuant(stop)
	case SanyTokenTExists, SanyTokenTForall:
		return p.SomeTQuant(stop)
	case SanyTokenLet:
		return p.LetIn(stop)
	case SanyTokenCase:
		return p.Case(stop)
	case SanyTokenChoose:
		return p.UnboundOrBoundChoose(stop)
	case SanyTokenLambda:
		return p.Lambda(stop)
	case SanyTokenWF, SanyTokenSF:
		return p.FairnessExpr()
	default:
		return p.PrimitiveExpression()
	}
}

func (p *SanyParser) FairnessExpr() *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	if p.match(SanyTokenWF) || p.match(SanyTokenSF) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
	} else {
		heirs = append(heirs, p.consume(SanyTokenWF, "expected WF_ or SF_"))
	}
	heirs = append(heirs, p.ReducedExpression())
	if p.match(SanyTokenLbr) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
		heirs = append(heirs, p.ExpressionUntil(func(tok *SanyToken) bool {
			return tok.Kind == SanyTokenRbr || tok.Kind == SanyTokenEOF
		}))
		heirs = append(heirs, p.consume(SanyTokenRbr, "expected ) in fairness expression"))
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_FairnessExpr"], heirs...)
}

func (p *SanyParser) FairnessSubscript() *SanySyntaxNode {
	return p.ReducedExpression()
}

func (p *SanyParser) LetIn(stop func(*SanyToken) bool) *SanySyntaxNode {
	let := p.consume(SanyTokenLet, "expected LET")
	defs := p.LetDefinitions()
	in := p.consume(SanyTokenLetin, "expected IN in LET expression")
	body := p.ExpressionUntil(stop)
	return NewSanyNode(SanySyntaxNodeKindByName["N_LetIn"], let, defs, in, body)
}

func (p *SanyParser) LetDefinitions() *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	for !p.check(SanyTokenLetin) && !p.check(SanyTokenEOF) {
		if p.check(SanyTokenRecursive) {
			heirs = append(heirs, p.Recursive())
			continue
		}
		if p.startsOperatorOrFunctionDefinition() {
			heirs = append(heirs, p.LetOperatorOrFunctionDefinition())
			continue
		}
		p.add(p.peek().Begin, "E1300", "expected LET definition")
		heirs = append(heirs, NewSanyTokenNode(p.advance()))
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_LetDefinitions"], heirs...)
}

func (p *SanyParser) LetOperatorOrFunctionDefinition() *SanySyntaxNode {
	if p.match(SanyTokenDefbreak) {
	}
	switch {
	case p.check(SanyTokenIdentifier) && p.peekNext().Kind == SanyTokenLsb:
		return p.LetFunctionDefinition()
	case p.check(SanyTokenIdentifier) && p.isPostfixOperator(p.peekNext()):
		return p.LetOperatorDefinition(p.PostfixLHS())
	case p.check(SanyTokenIdentifier) && p.isInfixOperator(p.peekNext()) && p.tokenAt(2).Kind == SanyTokenIdentifier:
		return p.LetOperatorDefinition(p.InfixLHS())
	case p.check(SanyTokenIdentifier) && p.peekNext().Kind == SanyTokenDef && p.tokenAt(2).Kind == SanyTokenInstance:
		return p.ModuleDefinition()
	case p.check(SanyTokenIdentifier) && (p.peekNext().Kind == SanyTokenLbr || p.peekNext().Kind == SanyTokenDef):
		return p.LetOperatorDefinition(p.IdentLHS())
	default:
		return p.LetOperatorDefinition(p.PrefixLHS())
	}
}

func (p *SanyParser) LetFunctionDefinition() *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	heirs = append(heirs, p.Identifier())
	heirs = append(heirs, p.consume(SanyTokenLsb, "expected [ in LET function definition"))
	heirs = append(heirs, p.QuantBound())
	for p.match(SanyTokenComma) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
		heirs = append(heirs, p.QuantBound())
	}
	heirs = append(heirs, p.consume(SanyTokenRsb, "expected ] in LET function definition"))
	heirs = append(heirs, p.consume(SanyTokenDef, "expected == in LET function definition"))
	heirs = append(heirs, p.ExpressionUntilDefinitionBoundary(func(tok *SanyToken) bool {
		return tok.Kind == SanyTokenLetin
	}))
	return NewSanySplitNode(SanySyntaxNodeKindByName["N_FunctionDefinition"], nil, heirs)
}

func (p *SanyParser) Case(stop func(*SanyToken) bool) *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	heirs = append(heirs, p.consume(SanyTokenCase, "expected CASE"))
	heirs = append(heirs, p.CaseArm(stop))
	for p.match(SanyTokenCasesep) {
		sep := NewSanyTokenNode(p.previous())
		if p.check(SanyTokenOther) {
			heirs = append(heirs, sep, p.OtherArm(stop))
			return NewSanyNode(SanySyntaxNodeKindByName["N_Case"], heirs...)
		}
		heirs = append(heirs, sep, p.CaseArm(stop))
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_Case"], heirs...)
}

func (p *SanyParser) CaseArm(stop func(*SanyToken) bool) *SanySyntaxNode {
	test := p.ExpressionUntil(func(tok *SanyToken) bool {
		return tok.Kind == SanyTokenArrow || tok.Kind == SanyTokenEOF
	})
	arrow := p.consume(SanyTokenArrow, "expected -> in CASE arm")
	value := p.ExpressionUntil(func(tok *SanyToken) bool {
		if tok.Kind == SanyTokenCasesep || tok.Kind == SanyTokenEOF || tok.Kind == SanyTokenEndModule {
			return true
		}
		return stop != nil && stop(tok)
	})
	return NewSanyNode(SanySyntaxNodeKindByName["N_CaseArm"], test, arrow, value)
}

func (p *SanyParser) OtherArm(stop func(*SanyToken) bool) *SanySyntaxNode {
	other := p.consume(SanyTokenOther, "expected OTHER")
	arrow := p.consume(SanyTokenArrow, "expected -> in OTHER arm")
	value := p.ExpressionUntil(func(tok *SanyToken) bool {
		if tok.Kind == SanyTokenEOF || tok.Kind == SanyTokenEndModule {
			return true
		}
		return stop != nil && stop(tok)
	})
	return NewSanyNode(SanySyntaxNodeKindByName["N_OtherArm"], other, arrow, value)
}

func (p *SanyParser) UnboundOrBoundChoose(stop func(*SanyToken) bool) *SanySyntaxNode {
	choose := p.consume(SanyTokenChoose, "expected CHOOSE")
	intro := p.QuantBoundIntro()
	maybe := p.MaybeBound()
	colon := p.consume(SanyTokenColon, "expected : in CHOOSE expression")
	body := p.ExpressionUntil(stop)
	return NewSanyNode(SanySyntaxNodeKindByName["N_UnboundOrBoundChoose"], choose, intro, maybe, colon, body)
}

func (p *SanyParser) MaybeBound() *SanySyntaxNode {
	if !p.match(SanyTokenIN) {
		return NewSanyNode(SanySyntaxNodeKindByName["N_MaybeBound"])
	}
	in := NewSanyTokenNode(p.previous())
	in.Kind = SanySyntaxNodeKindByName["T_IN"]
	expr := p.ExpressionUntil(func(tok *SanyToken) bool {
		return tok.Kind == SanyTokenColon || tok.Kind == SanyTokenEOF
	})
	return NewSanyNode(SanySyntaxNodeKindByName["N_MaybeBound"], in, expr)
}

func (p *SanyParser) Lambda(stop func(*SanyToken) bool) *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	heirs = append(heirs, p.consume(SanyTokenLambda, "expected LAMBDA"))
	heirs = append(heirs, p.IdentDeclOrSomeFixDecl())
	for p.match(SanyTokenComma) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
		heirs = append(heirs, p.IdentDeclOrSomeFixDecl())
	}
	heirs = append(heirs, p.consume(SanyTokenColon, "expected : in LAMBDA expression"))
	heirs = append(heirs, p.ExpressionUntil(stop))
	return NewSanyNode(SanySyntaxNodeKindByName["N_Lambda"], heirs...)
}

func (p *SanyParser) IfThenElse(stop func(*SanyToken) bool) *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	heirs = append(heirs, p.consume(SanyTokenIf, "expected IF"))
	heirs = append(heirs, p.ExpressionUntil(func(tok *SanyToken) bool {
		return tok.Kind == SanyTokenThen || tok.Kind == SanyTokenEOF
	}))
	heirs = append(heirs, p.consume(SanyTokenThen, "expected THEN"))
	heirs = append(heirs, p.ExpressionUntil(func(tok *SanyToken) bool {
		return tok.Kind == SanyTokenElse || tok.Kind == SanyTokenEOF
	}))
	heirs = append(heirs, p.consume(SanyTokenElse, "expected ELSE"))
	heirs = append(heirs, p.ExpressionUntil(stop))
	return NewSanyNode(SanySyntaxNodeKindByName["N_IfThenElse"], heirs...)
}

func (p *SanyParser) SomeQuant(stop func(*SanyToken) bool) *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	if p.match(SanyTokenExists) || p.match(SanyTokenForall) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
	} else {
		heirs = append(heirs, p.consume(SanyTokenForall, "expected quantified expression"))
	}
	kind := SanySyntaxNodeKindByName["N_UnboundQuant"]
	if p.findBeforeStop(SanyTokenIN, SanyTokenColon, SanyTokenEOF) >= 0 {
		kind = SanySyntaxNodeKindByName["N_BoundQuant"]
		heirs = append(heirs, p.QuantBoundUntil(SanyTokenComma, SanyTokenColon, SanyTokenEOF))
		for p.match(SanyTokenComma) {
			if p.findBeforeStop(SanyTokenIN, SanyTokenColon, SanyTokenEOF) < 0 {
				p.at--
				break
			}
			heirs = append(heirs, NewSanyTokenNode(p.previous()))
			heirs = append(heirs, p.QuantBoundUntil(SanyTokenComma, SanyTokenColon, SanyTokenEOF))
		}
	} else {
		heirs = append(heirs, p.Identifier())
		for p.match(SanyTokenComma) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()))
			heirs = append(heirs, p.Identifier())
		}
	}
	heirs = append(heirs, p.consume(SanyTokenColon, "expected : in quantified expression"))
	heirs = append(heirs, p.ExpressionUntil(p.stopAfterOpenExpressionBody(stop)))
	return NewSanyNode(kind, heirs...)
}

func (p *SanyParser) stopAfterOpenExpressionBody(stop func(*SanyToken) bool) func(*SanyToken) bool {
	startLine := p.peek().Begin.Line
	return func(tok *SanyToken) bool {
		if IsSanyJunctionBullet(tok.Kind) && tok.Begin.Line == startLine {
			return false
		}
		if stop == nil {
			return false
		}
		return stop(tok)
	}
}

func (p *SanyParser) SomeTQuant(stop func(*SanyToken) bool) *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	if p.match(SanyTokenTExists) || p.match(SanyTokenTForall) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
	} else {
		heirs = append(heirs, p.consume(SanyTokenTForall, "expected temporal quantified expression"))
	}
	heirs = append(heirs, p.Identifier())
	for p.match(SanyTokenComma) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
		heirs = append(heirs, p.Identifier())
	}
	heirs = append(heirs, p.consume(SanyTokenColon, "expected : in temporal quantified expression"))
	heirs = append(heirs, p.ExpressionUntil(stop))
	return NewSanyNode(SanySyntaxNodeKindByName["N_UnboundQuant"], heirs...)
}

func (p *SanyParser) startsPrimitiveExpression() bool {
	switch p.peek().Kind {
	case SanyTokenNumberLiteral, SanyTokenStringLiteral, SanyTokenLbr, SanyTokenLbc, SanyTokenLab, SanyTokenLsb:
		return true
	default:
		return false
	}
}

func (p *SanyParser) PrimitiveExpression() *SanySyntaxNode {
	switch p.peek().Kind {
	case SanyTokenNumberLiteral:
		return p.Number()
	case SanyTokenStringLiteral:
		return p.String()
	case SanyTokenLbr:
		return p.ParenExpr()
	case SanyTokenLbc:
		return p.BraceCases()
	case SanyTokenLab:
		return p.TupleOrAction()
	case SanyTokenLsb:
		return p.SBracketCases()
	default:
		return NewSanyTokenNode(p.advance())
	}
}

func (p *SanyParser) String() *SanySyntaxNode {
	tok := p.consume(SanyTokenStringLiteral, "expected string literal")
	node := NewSanyNode(SanySyntaxNodeKindByName["N_String"], tok)
	if tok != nil {
		node.Image = reduceTLAString(tok.Image)
	}
	return node
}

func (p *SanyParser) Number() *SanySyntaxNode {
	first := p.consume(SanyTokenNumberLiteral, "expected number literal")
	if p.check(SanyTokenDot) && p.peekNext().Kind == SanyTokenNumberLiteral {
		dot := NewSanyTokenNode(p.advance())
		second := p.consume(SanyTokenNumberLiteral, "expected number literal after decimal point")
		return NewSanyNode(SanySyntaxNodeKindByName["N_Real"], first, dot, second)
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_Number"], first)
}

func (p *SanyParser) ParenExpr() *SanySyntaxNode {
	left := p.consume(SanyTokenLbr, "expected (")
	expr := p.ExpressionUntil(func(tok *SanyToken) bool {
		return tok.Kind == SanyTokenRbr || tok.Kind == SanyTokenEOF
	})
	right := p.consume(SanyTokenRbr, "expected )")
	return NewSanyNode(SanySyntaxNodeKindByName["N_ParenExpr"], left, expr, right)
}

func (p *SanyParser) BraceCases() *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	heirs = append(heirs, p.consume(SanyTokenLbc, "expected {"))
	if p.startsQuantBoundIntro() &&
		p.findTopLevelBeforeStop(SanyTokenIN, SanyTokenColon, SanyTokenRbc, SanyTokenEOF) >= 0 &&
		p.findTopLevelSetComprehensionColonBeforeStop(SanyTokenComma, SanyTokenRbc, SanyTokenEOF) >= 0 {
		heirs = append(heirs, p.QuantBoundIntro())
		in := p.consume(SanyTokenIN, "expected \\in in subset expression")
		if in != nil {
			in.Kind = SanySyntaxNodeKindByName["T_IN"]
		}
		heirs = append(heirs, in)
		heirs = append(heirs, p.ExpressionUntil(func(tok *SanyToken) bool {
			return tok.Kind == SanyTokenColon || tok.Kind == SanyTokenEOF
		}))
		heirs = append(heirs, p.consume(SanyTokenColon, "expected : in subset expression"))
		heirs = append(heirs, p.ExpressionUntil(func(tok *SanyToken) bool {
			return tok.Kind == SanyTokenRbc || tok.Kind == SanyTokenEOF
		}))
		heirs = append(heirs, p.consume(SanyTokenRbc, "expected }"))
		return NewSanyNode(SanySyntaxNodeKindByName["N_SubsetOf"], heirs...)
	}
	if p.findTopLevelSetComprehensionColonBeforeStop(SanyTokenComma, SanyTokenRbc, SanyTokenEOF) >= 0 {
		heirs = append(heirs, p.ExpressionUntil(func(tok *SanyToken) bool {
			return tok.Kind == SanyTokenColon || tok.Kind == SanyTokenEOF
		}))
		heirs = append(heirs, p.consume(SanyTokenColon, "expected : in set comprehension"))
		heirs = append(heirs, p.QuantBoundUntil(SanyTokenComma, SanyTokenRbc, SanyTokenEOF))
		for p.match(SanyTokenComma) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()))
			heirs = append(heirs, p.QuantBoundUntil(SanyTokenComma, SanyTokenRbc, SanyTokenEOF))
		}
		heirs = append(heirs, p.consume(SanyTokenRbc, "expected }"))
		return NewSanyNode(SanySyntaxNodeKindByName["N_SetOfAll"], heirs...)
	}
	if !p.check(SanyTokenRbc) {
		heirs = append(heirs, p.ExpressionUntil(func(tok *SanyToken) bool {
			return tok.Kind == SanyTokenComma || tok.Kind == SanyTokenRbc || tok.Kind == SanyTokenEOF
		}))
		for p.match(SanyTokenComma) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()))
			heirs = append(heirs, p.ExpressionUntil(func(tok *SanyToken) bool {
				return tok.Kind == SanyTokenComma || tok.Kind == SanyTokenRbc || tok.Kind == SanyTokenEOF
			}))
		}
	}
	heirs = append(heirs, p.consume(SanyTokenRbc, "expected }"))
	return NewSanyNode(SanySyntaxNodeKindByName["N_SetEnumerate"], heirs...)
}

func (p *SanyParser) TupleOrAction() *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	heirs = append(heirs, p.consume(SanyTokenLab, "expected <<"))
	if !p.check(SanyTokenRab) && !p.check(SanyTokenArab) {
		heirs = append(heirs, p.ExpressionUntil(func(tok *SanyToken) bool {
			return tok.Kind == SanyTokenComma || tok.Kind == SanyTokenRab || tok.Kind == SanyTokenArab || tok.Kind == SanyTokenEOF
		}))
		for p.match(SanyTokenComma) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()))
			heirs = append(heirs, p.ExpressionUntil(func(tok *SanyToken) bool {
				return tok.Kind == SanyTokenComma || tok.Kind == SanyTokenRab || tok.Kind == SanyTokenArab || tok.Kind == SanyTokenEOF
			}))
		}
	}
	if p.match(SanyTokenArab) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
		heirs = append(heirs, p.ReducedExpression())
		return NewSanyNode(SanySyntaxNodeKindByName["N_ActionExpr"], heirs...)
	}
	heirs = append(heirs, p.consume(SanyTokenRab, "expected >>"))
	return NewSanyNode(SanySyntaxNodeKindByName["N_Tuple"], heirs...)
}

func (p *SanyParser) SBracketCases() *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	heirs = append(heirs, p.consume(SanyTokenLsb, "expected ["))
	if p.startsQuantBoundIntro() && p.findTopLevelBeforeStop(SanyTokenMapto, SanyTokenRsb, SanyTokenArsb, SanyTokenEOF) >= 0 &&
		p.findTopLevelBeforeStop(SanyTokenIN, SanyTokenMapto, SanyTokenRsb, SanyTokenArsb, SanyTokenEOF) >= 0 {
		heirs = append(heirs, p.QuantBoundUntil(SanyTokenComma, SanyTokenMapto, SanyTokenRsb, SanyTokenArsb, SanyTokenEOF))
		for p.match(SanyTokenComma) {
			if p.findTopLevelBeforeStop(SanyTokenIN, SanyTokenMapto, SanyTokenRsb, SanyTokenArsb, SanyTokenEOF) < 0 {
				p.at--
				break
			}
			heirs = append(heirs, NewSanyTokenNode(p.previous()))
			heirs = append(heirs, p.QuantBoundUntil(SanyTokenComma, SanyTokenMapto, SanyTokenRsb, SanyTokenArsb, SanyTokenEOF))
		}
		heirs = append(heirs, p.consume(SanyTokenMapto, "expected |-> in function constructor"))
		heirs = append(heirs, p.ExpressionUntil(func(tok *SanyToken) bool {
			return tok.Kind == SanyTokenRsb || tok.Kind == SanyTokenEOF
		}))
		heirs = append(heirs, p.consume(SanyTokenRsb, "expected ]"))
		return NewSanyNode(SanySyntaxNodeKindByName["N_FcnConst"], heirs...)
	}
	if p.check(SanyTokenIdentifier) && p.peekNext().Kind == SanyTokenColon {
		heirs = append(heirs, p.FieldSet())
		for p.match(SanyTokenComma) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()))
			heirs = append(heirs, p.FieldSet())
		}
		heirs = append(heirs, p.consume(SanyTokenRsb, "expected ]"))
		return NewSanyNode(SanySyntaxNodeKindByName["N_SetOfRcds"], heirs...)
	}
	if p.check(SanyTokenIdentifier) && p.peekNext().Kind == SanyTokenMapto {
		heirs = append(heirs, p.FieldVal())
		for p.match(SanyTokenComma) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()))
			heirs = append(heirs, p.FieldVal())
		}
		heirs = append(heirs, p.consume(SanyTokenRsb, "expected ]"))
		return NewSanyNode(SanySyntaxNodeKindByName["N_RcdConstructor"], heirs...)
	}
	if p.findTopLevelBeforeStop(SanyTokenExcept, SanyTokenRsb, SanyTokenArsb, SanyTokenEOF) >= 0 {
		heirs = append(heirs, p.ExpressionUntil(func(tok *SanyToken) bool {
			return tok.Kind == SanyTokenExcept || tok.Kind == SanyTokenEOF
		}))
		heirs = append(heirs, p.consume(SanyTokenExcept, "expected EXCEPT"))
		heirs = append(heirs, p.ExceptSpec())
		for p.match(SanyTokenComma) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()))
			heirs = append(heirs, p.ExceptSpec())
		}
		heirs = append(heirs, p.consume(SanyTokenRsb, "expected ]"))
		return NewSanyNode(SanySyntaxNodeKindByName["N_Except"], heirs...)
	}
	if p.findTopLevelBeforeStop(SanyTokenArrow, SanyTokenRsb, SanyTokenArsb, SanyTokenEOF) >= 0 {
		heirs = append(heirs, p.ExpressionUntil(func(tok *SanyToken) bool {
			return tok.Kind == SanyTokenArrow || tok.Kind == SanyTokenEOF
		}))
		heirs = append(heirs, p.consume(SanyTokenArrow, "expected -> in function set"))
		heirs = append(heirs, p.ExpressionUntil(func(tok *SanyToken) bool {
			return tok.Kind == SanyTokenRsb || tok.Kind == SanyTokenEOF
		}))
		heirs = append(heirs, p.consume(SanyTokenRsb, "expected ]"))
		return NewSanyNode(SanySyntaxNodeKindByName["N_SetOfFcns"], heirs...)
	}
	if p.findTopLevelBeforeStop(SanyTokenArsb, SanyTokenRsb, SanyTokenEOF) >= 0 {
		heirs = append(heirs, p.ExpressionUntil(func(tok *SanyToken) bool {
			return tok.Kind == SanyTokenArsb || tok.Kind == SanyTokenEOF
		}))
		heirs = append(heirs, p.consume(SanyTokenArsb, "expected ]_"))
		heirs = append(heirs, p.ReducedExpression())
		return NewSanyNode(SanySyntaxNodeKindByName["N_ActionExpr"], heirs...)
	}
	if !p.check(SanyTokenRsb) {
		heirs = append(heirs, p.ExpressionUntil(func(tok *SanyToken) bool {
			return tok.Kind == SanyTokenComma || tok.Kind == SanyTokenRsb || tok.Kind == SanyTokenEOF
		}))
		for p.match(SanyTokenComma) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()))
			heirs = append(heirs, p.ExpressionUntil(func(tok *SanyToken) bool {
				return tok.Kind == SanyTokenComma || tok.Kind == SanyTokenRsb || tok.Kind == SanyTokenEOF
			}))
		}
	}
	heirs = append(heirs, p.consume(SanyTokenRsb, "expected ]"))
	return NewSanyNode(SanySyntaxNodeKindByName["N_FcnAppl"], heirs...)
}

func (p *SanyParser) FieldVal() *SanySyntaxNode {
	id := p.Identifier()
	mapto := p.consume(SanyTokenMapto, "expected |-> in record field")
	expr := p.ExpressionUntil(func(tok *SanyToken) bool {
		return tok.Kind == SanyTokenComma || tok.Kind == SanyTokenRsb || tok.Kind == SanyTokenEOF
	})
	return NewSanyNode(SanySyntaxNodeKindByName["N_FieldVal"], id, mapto, expr)
}

func (p *SanyParser) FieldSet() *SanySyntaxNode {
	id := p.Identifier()
	colon := p.consume(SanyTokenColon, "expected : in record field set")
	expr := p.ExpressionUntil(func(tok *SanyToken) bool {
		return tok.Kind == SanyTokenComma || tok.Kind == SanyTokenRsb || tok.Kind == SanyTokenEOF
	})
	return NewSanyNode(SanySyntaxNodeKindByName["N_FieldSet"], id, colon, expr)
}

func (p *SanyParser) ExceptSpec() *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	heirs = append(heirs, p.consume(SanyTokenBang, "expected ! in EXCEPT spec"))
	if !p.check(SanyTokenDot) && !p.check(SanyTokenLsb) {
		p.add(p.peek().Begin, "E1300", "expected EXCEPT component")
	}
	for p.check(SanyTokenDot) || p.check(SanyTokenLsb) {
		heirs = append(heirs, p.ExceptComponent())
	}
	equals := p.consume(SanyTokenEquals, "expected = in EXCEPT spec")
	if equals != nil {
		equals.Kind = SanySyntaxNodeKindByName["T_EQUAL"]
	}
	heirs = append(heirs, equals)
	heirs = append(heirs, p.ExpressionUntil(func(tok *SanyToken) bool {
		return tok.Kind == SanyTokenComma || tok.Kind == SanyTokenRsb || tok.Kind == SanyTokenEOF
	}))
	return NewSanyNode(SanySyntaxNodeKindByName["N_ExceptSpec"], heirs...)
}

func (p *SanyParser) ExceptComponent() *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	if p.match(SanyTokenDot) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
		heirs = append(heirs, p.Identifier())
		return NewSanyNode(SanySyntaxNodeKindByName["N_ExceptComponent"], heirs...)
	}
	heirs = append(heirs, p.consume(SanyTokenLsb, "expected [ in EXCEPT component"))
	if !p.check(SanyTokenRsb) {
		heirs = append(heirs, p.ExpressionUntil(func(tok *SanyToken) bool {
			return tok.Kind == SanyTokenComma || tok.Kind == SanyTokenRsb || tok.Kind == SanyTokenEOF
		}))
		for p.match(SanyTokenComma) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()))
			heirs = append(heirs, p.ExpressionUntil(func(tok *SanyToken) bool {
				return tok.Kind == SanyTokenComma || tok.Kind == SanyTokenRsb || tok.Kind == SanyTokenEOF
			}))
		}
	}
	heirs = append(heirs, p.consume(SanyTokenRsb, "expected ] in EXCEPT component"))
	return NewSanyNode(SanySyntaxNodeKindByName["N_ExceptComponent"], heirs...)
}

func (p *SanyParser) ReducedExpression() *SanySyntaxNode {
	switch {
	case p.check(SanyTokenIdentifier):
		return p.NoOpExtension()
	case p.check(SanyTokenLbr):
		return p.ParenExpr()
	case p.check(SanyTokenLbc):
		return p.BraceCases()
	case p.check(SanyTokenLsb):
		return p.SBracketCases()
	case p.check(SanyTokenLab):
		return p.TupleOrAction()
	default:
		p.add(p.peek().Begin, "E1302", "expected restricted expression after action subscript")
		return nil
	}
}

func (p *SanyParser) startsNoOpExtension() bool {
	if !p.check(SanyTokenIdentifier) {
		return false
	}
	offset := 1
	if p.tokenAt(offset).Kind == SanyTokenLbr {
		end := p.findMatchingBracketOffset(offset)
		if end < 0 {
			return false
		}
		offset = end + 1
	}
	return p.tokenAt(offset).Kind == SanyTokenBang
}

func (p *SanyParser) splitLeadingFairnessIdentifier() bool {
	tok := p.peek()
	if tok == nil {
		return false
	}
	var kind SanyTokenKind
	var prefix string
	switch {
	case (tok.Kind == SanyTokenIdentifier || tok.Kind == SanyTokenWF) && strings.HasPrefix(tok.Image, "WF_") && len(tok.Image) > len("WF_"):
		kind = SanyTokenWF
		prefix = "WF_"
	case (tok.Kind == SanyTokenIdentifier || tok.Kind == SanyTokenSF) && strings.HasPrefix(tok.Image, "SF_") && len(tok.Image) > len("SF_"):
		kind = SanyTokenSF
		prefix = "SF_"
	default:
		return false
	}

	suffixImage := tok.Image[len(prefix):]
	prefixEnd := tok.Begin
	prefixEnd.Column += len(prefix) - 1
	suffixBegin := tok.Begin
	suffixBegin.Column += len(prefix)
	prefixTok := &SanyToken{
		Kind:     kind,
		Image:    prefix,
		Begin:    tok.Begin,
		End:      prefixEnd,
		LexState: tok.LexState,
		Special:  tok.Special,
	}
	suffixTok := &SanyToken{
		Kind:     SanyTokenIdentifier,
		Image:    suffixImage,
		Begin:    suffixBegin,
		End:      tok.End,
		LexState: tok.LexState,
		Next:     tok.Next,
	}
	prefixTok.Next = suffixTok
	if p.at > 0 {
		p.tokens[p.at-1].Next = prefixTok
	}
	p.tokens = append(p.tokens, nil)
	copy(p.tokens[p.at+2:], p.tokens[p.at+1:])
	p.tokens[p.at] = prefixTok
	p.tokens[p.at+1] = suffixTok
	return true
}

func (p *SanyParser) startsStructOp() bool {
	switch p.peek().Kind {
	case SanyTokenLab, SanyTokenRab, SanyTokenColon, SanyTokenNumberLiteral:
		return true
	case SanyTokenIdentifier:
		return p.peek().Image == "@"
	default:
		return false
	}
}

func (p *SanyParser) startsBangOperatorSelector() bool {
	if _, ok := GetSanyOperator(p.peek().Image); ok {
		return true
	}
	return false
}

func (p *SanyParser) startsLabelAt(offset int) bool {
	if p.tokenAt(offset).Kind != SanyTokenIdentifier {
		return false
	}
	if p.tokenAt(offset+1).Kind == SanyTokenColoncolon {
		return true
	}
	if p.tokenAt(offset+1).Kind != SanyTokenLbr {
		return false
	}
	end := p.findMatchingBracketOffset(offset + 1)
	return end >= 0 && p.tokenAt(end+1).Kind == SanyTokenColoncolon
}

func (p *SanyParser) startsOpApplication() bool {
	return p.check(SanyTokenIdentifier) && p.peekNext().Kind == SanyTokenLbr
}

func (p *SanyParser) OpApplication() *SanySyntaxNode {
	id := p.Identifier()
	genID := NewSanyNode(
		SanySyntaxNodeKindByName["N_GeneralId"],
		NewSanyNode(SanySyntaxNodeKindByName["N_IdPrefix"]),
		id,
	)
	return NewSanyNode(SanySyntaxNodeKindByName["N_OpApplication"], genID, p.OpArgs())
}

func (p *SanyParser) OpArgs() *SanySyntaxNode {
	var heirs []*SanySyntaxNode
	heirs = append(heirs, p.consume(SanyTokenLbr, "expected ( in operator arguments"))
	if !p.check(SanyTokenRbr) {
		heirs = append(heirs, p.ExpressionUntil(func(tok *SanyToken) bool {
			return tok.Kind == SanyTokenComma || tok.Kind == SanyTokenRbr || tok.Kind == SanyTokenEOF
		}))
		for p.match(SanyTokenComma) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()))
			heirs = append(heirs, p.ExpressionUntil(func(tok *SanyToken) bool {
				return tok.Kind == SanyTokenComma || tok.Kind == SanyTokenRbr || tok.Kind == SanyTokenEOF
			}))
		}
	}
	heirs = append(heirs, p.consume(SanyTokenRbr, "expected ) in operator arguments"))
	return NewSanyNode(SanySyntaxNodeKindByName["N_OpArgs"], heirs...)
}

func (p *SanyParser) NoOpExtension() *SanySyntaxNode {
	var prefix []*SanySyntaxNode
	selector := p.Identifier()
	args := p.OptionalSelectorOpArgs(selector)
	for p.match(SanyTokenBang) {
		bang := NewSanyTokenNode(p.previous())
		if args != nil {
			prefix = append(prefix, NewSanyNode(SanySyntaxNodeKindByName["N_IdPrefixElement"], selector, args, bang))
		} else {
			prefix = append(prefix, NewSanyNode(SanySyntaxNodeKindByName["N_IdPrefixElement"], selector, bang))
		}
		selector = p.BangSelector()
		args = p.OptionalSelectorOpArgs(selector)
	}
	genID := NewSanyNode(
		SanySyntaxNodeKindByName["N_GeneralId"],
		NewSanyNode(SanySyntaxNodeKindByName["N_IdPrefix"], prefix...),
		selector,
	)
	if args != nil {
		return NewSanyNode(SanySyntaxNodeKindByName["N_OpApplication"], genID, args)
	}
	return genID
}

func (p *SanyParser) OptionalSelectorOpArgs(selector *SanySyntaxNode) *SanySyntaxNode {
	if selector == nil || !p.selectorAllowsOpArgs(selector) || !p.check(SanyTokenLbr) {
		return nil
	}
	return p.OpArgs()
}

func (p *SanyParser) selectorAllowsOpArgs(selector *SanySyntaxNode) bool {
	switch selector.Kind.JavaName() {
	case "IDENTIFIER", "N_InfixOp", "N_NonExpPrefixOp", "N_PostfixOp", "N_PrefixOp":
		return true
	default:
		return false
	}
}

func (p *SanyParser) BangSelector() *SanySyntaxNode {
	if p.check(SanyTokenLbr) {
		return p.OpArgs()
	}
	if p.startsStructOp() {
		return p.StructOp()
	}
	if p.startsBangOperatorSelector() {
		return p.BangOperatorSelector()
	}
	return p.Identifier()
}

func (p *SanyParser) BangOperatorSelector() *SanySyntaxNode {
	tok := p.peek()
	op, ok := GetSanyOperator(tok.Image)
	if !ok {
		p.add(tok.Begin, "E1300", "expected operator selector")
		return NewSanyTokenNode(p.advance())
	}
	kindName := "N_InfixOp"
	if op.IsPrefix() {
		kindName = "N_NonExpPrefixOp"
	} else if op.IsPostfix() {
		kindName = "N_PostfixOp"
	}
	return NewSanyNode(SanySyntaxNodeKindByName[kindName], NewSanyTokenNode(p.advance()))
}

func (p *SanyParser) StructOp() *SanySyntaxNode {
	if p.check(SanyTokenNumberLiteral) {
		number := p.Number()
		if number != nil && number.Kind.JavaName() == "N_Real" {
			p.add(number.Range.Begin, "E1300", "illegal structural term")
		}
		return NewSanyNode(SanySyntaxNodeKindByName["N_StructOp"], number)
	}
	if p.check(SanyTokenIdentifier) && p.peek().Image == "@" {
		return NewSanyNode(SanySyntaxNodeKindByName["N_StructOp"], NewSanyTokenNode(p.advance()))
	}
	return NewSanyNode(
		SanySyntaxNodeKindByName["N_StructOp"],
		p.consumeAny([]SanyTokenKind{SanyTokenLab, SanyTokenRab, SanyTokenColon}, "expected structural operator"),
	)
}

func reduceTLAString(image string) string {
	value, err := strconv.Unquote(image)
	if err == nil {
		return value
	}
	if len(image) >= 2 && image[0] == '"' && image[len(image)-1] == '"' {
		return image[1 : len(image)-1]
	}
	return image
}

func (p *SanyParser) genericOperatorNode(tok *SanyToken, op SanyOperatorInfo) *SanySyntaxNode {
	kindName := "N_GenInfixOp"
	if op.IsPrefix() {
		kindName = "N_GenPrefixOp"
	} else if op.IsPostfix() {
		kindName = "N_GenPostfixOp"
	}
	return NewSanyNode(
		SanySyntaxNodeKindByName[kindName],
		NewSanyNode(SanySyntaxNodeKindByName["N_IdPrefix"]),
		NewSanyTokenNode(tok),
	)
}

func (p *SanyParser) Identifier() *SanySyntaxNode {
	return p.consume(SanyTokenIdentifier, "expected identifier")
}

func (p *SanyParser) startsBodyItem() bool {
	return p.startsBodyItemAt(0)
}

func (p *SanyParser) startsBodyItemAt(offset int) bool {
	switch p.tokenAt(offset).Kind {
	case SanyTokenBm0, SanyTokenBm1, SanyTokenBm2, SanyTokenSeparator, SanyTokenVariable, SanyTokenConstant, SanyTokenRecursive, SanyTokenAssume, SanyTokenAssumption, SanyTokenTheorem, SanyTokenProposition, SanyTokenLocal, SanyTokenInstance, SanyTokenUse, SanyTokenHide:
		return true
	default:
		return p.startsOperatorOrFunctionDefinitionAt(offset)
	}
}

func (p *SanyParser) startsOperatorOrFunctionDefinition() bool {
	return p.startsOperatorOrFunctionDefinitionAt(0)
}

func (p *SanyParser) startsOperatorOrFunctionDefinitionAt(offset int) bool {
	if p.tokenAt(offset).Kind == SanyTokenLocal {
		offset++
	}
	if p.tokenAt(offset).Kind == SanyTokenDefbreak {
		offset++
	}
	first := p.tokenAt(offset)
	second := p.tokenAt(offset + 1)
	switch {
	case first.Kind == SanyTokenIdentifier && second.Kind == SanyTokenLsb:
		end := p.findMatchingBracketOffset(offset + 1)
		return end >= 0 && p.tokenAt(end+1).Kind == SanyTokenDef
	case first.Kind == SanyTokenIdentifier && p.isPostfixOperator(second) && p.tokenAt(offset+2).Kind == SanyTokenDef:
		return true
	case first.Kind == SanyTokenIdentifier && p.isInfixOperator(second) && p.tokenAt(offset+2).Kind == SanyTokenIdentifier && p.tokenAt(offset+3).Kind == SanyTokenDef:
		return true
	case first.Kind == SanyTokenIdentifier && second.Kind == SanyTokenLbr:
		end := p.findMatchingBracketOffset(offset + 1)
		return end >= 0 && p.tokenAt(end+1).Kind == SanyTokenDef
	case first.Kind == SanyTokenIdentifier && second.Kind == SanyTokenDef:
		return true
	case p.isNEPrefixOperator(first) && second.Kind == SanyTokenIdentifier && p.tokenAt(offset+2).Kind == SanyTokenDef:
		return true
	default:
		return false
	}
}

func (p *SanyParser) isInfixOperator(tok *SanyToken) bool {
	if tok == nil {
		return false
	}
	op, ok := GetSanyOperator(tok.Image)
	return ok && op.IsInfix()
}

func (p *SanyParser) isPostfixOperator(tok *SanyToken) bool {
	if tok == nil {
		return false
	}
	op, ok := GetSanyOperator(tok.Image)
	return ok && op.IsPostfix()
}

func (p *SanyParser) isNEPrefixOperator(tok *SanyToken) bool {
	if tok == nil {
		return false
	}
	op, ok := GetSanyOperator(tok.Image)
	return ok && op.IsPrefix()
}

func (p *SanyParser) consumeOperator(msg string) *SanySyntaxNode {
	if _, ok := GetSanyOperator(p.peek().Image); ok {
		return NewSanyTokenNode(p.advance())
	}
	p.add(p.peek().Begin, "E1300", msg)
	return nil
}

func (p *SanyParser) findMatchingBracketOffset(start int) int {
	close, ok := matchingSanyCloseBracket(p.tokenAt(start).Kind)
	if !ok {
		return -1
	}
	stack := []SanyTokenKind{close}
	for offset := start + 1; ; offset++ {
		tok := p.tokenAt(offset)
		if tok.Kind == SanyTokenEOF {
			return -1
		}
		if close, ok := matchingSanyCloseBracket(tok.Kind); ok {
			stack = append(stack, close)
			continue
		}
		if len(stack) > 0 && tok.Kind == stack[len(stack)-1] {
			stack = stack[:len(stack)-1]
			if len(stack) == 0 {
				return offset
			}
		}
	}
}

func matchingSanyCloseBracket(kind SanyTokenKind) (SanyTokenKind, bool) {
	switch kind {
	case SanyTokenLbr:
		return SanyTokenRbr, true
	case SanyTokenLsb:
		return SanyTokenRsb, true
	case SanyTokenLbc:
		return SanyTokenRbc, true
	case SanyTokenLab:
		return SanyTokenRab, true
	default:
		return SanyTokenEOF, false
	}
}

func (p *SanyParser) findBeforeStop(target SanyTokenKind, stops ...SanyTokenKind) int {
	for offset := 0; ; offset++ {
		tok := p.tokenAt(offset)
		if tok.Kind == target {
			return offset
		}
		for _, stop := range stops {
			if tok.Kind == stop {
				return -1
			}
		}
		if tok.Kind == SanyTokenEOF {
			return -1
		}
	}
}

func (p *SanyParser) findBeforeStopIgnoring(target, ignoreStop SanyTokenKind, stops ...SanyTokenKind) int {
	for offset := 0; ; offset++ {
		tok := p.tokenAt(offset)
		if tok.Kind == target {
			return offset
		}
		for _, stop := range stops {
			if stop == ignoreStop {
				continue
			}
			if tok.Kind == stop {
				return -1
			}
		}
		if tok.Kind == SanyTokenEOF {
			return -1
		}
	}
}

func (p *SanyParser) findTopLevelBeforeStop(target SanyTokenKind, stops ...SanyTokenKind) int {
	depth := 0
	for offset := 0; ; offset++ {
		tok := p.tokenAt(offset)
		if depth == 0 {
			if tok.Kind == target {
				return offset
			}
			for _, stop := range stops {
				if tok.Kind == stop {
					return -1
				}
			}
		}
		switch tok.Kind {
		case SanyTokenLbr, SanyTokenLsb, SanyTokenLbc, SanyTokenLab:
			depth++
		case SanyTokenRbr, SanyTokenRsb, SanyTokenRbc, SanyTokenRab:
			if depth > 0 {
				depth--
			}
		case SanyTokenEOF:
			return -1
		}
	}
}

func (p *SanyParser) findTopLevelSetComprehensionColonBeforeStop(stops ...SanyTokenKind) int {
	depth := 0
	binderColons := 0
	for offset := 0; ; offset++ {
		tok := p.tokenAt(offset)
		if depth == 0 {
			if tok.Kind == SanyTokenColon {
				if binderColons > 0 {
					binderColons--
				} else {
					return offset
				}
			}
			for _, stop := range stops {
				if tok.Kind == stop {
					return -1
				}
			}
			if sanyTokenOwnsFollowingColon(tok.Kind) {
				binderColons++
			}
		}
		switch tok.Kind {
		case SanyTokenLbr, SanyTokenLsb, SanyTokenLbc, SanyTokenLab:
			depth++
		case SanyTokenRbr, SanyTokenRsb, SanyTokenRbc, SanyTokenRab:
			if depth > 0 {
				depth--
			}
		case SanyTokenEOF:
			return -1
		}
	}
}

func sanyTokenOwnsFollowingColon(kind SanyTokenKind) bool {
	switch kind {
	case SanyTokenChoose, SanyTokenForall, SanyTokenExists, SanyTokenTExists, SanyTokenTForall, SanyTokenLambda:
		return true
	default:
		return false
	}
}

func (p *SanyParser) match(kind SanyTokenKind) bool {
	if p.check(kind) {
		p.advance()
		return true
	}
	return false
}

func (p *SanyParser) consume(kind SanyTokenKind, msg string) *SanySyntaxNode {
	if p.check(kind) {
		return NewSanyTokenNode(p.advance())
	}
	p.add(p.peek().Begin, "E1300", msg)
	return nil
}

func (p *SanyParser) consumeAny(kinds []SanyTokenKind, msg string) *SanySyntaxNode {
	for _, kind := range kinds {
		if p.check(kind) {
			return NewSanyTokenNode(p.advance())
		}
	}
	p.add(p.peek().Begin, "E1300", msg)
	return nil
}

func (p *SanyParser) check(kind SanyTokenKind) bool {
	return p.peek().Kind == kind
}

func (p *SanyParser) advance() *SanyToken {
	if !p.check(SanyTokenEOF) {
		p.at++
	}
	return p.previous()
}

func (p *SanyParser) previous() *SanyToken {
	if p.at == 0 {
		return nil
	}
	return p.tokens[p.at-1]
}

func (p *SanyParser) peek() *SanyToken {
	if p.at >= len(p.tokens) {
		return &SanyToken{Kind: SanyTokenEOF}
	}
	return p.tokens[p.at]
}

func (p *SanyParser) peekNext() *SanyToken {
	return p.tokenAt(1)
}

func (p *SanyParser) tokenAt(offset int) *SanyToken {
	idx := p.at + offset
	if idx >= len(p.tokens) {
		return &SanyToken{Kind: SanyTokenEOF}
	}
	return p.tokens[idx]
}

func (p *SanyParser) add(pos Position, code, msg string) {
	p.diags = append(p.diags, errorAt(pos, code, "%s", msg))
}
