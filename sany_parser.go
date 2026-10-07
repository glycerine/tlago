package tlago

import (
	"strconv"
	"strings"
)

type SanyParser struct {
	output               SanyOutput
	tokens               []*SanyToken
	tokenManager         *SanyTokenManager
	at                   int
	diags                Diagnostics
	moduleName           string
	proofLevelStack      []int
	junctionListContext  sanyJunctionListContext
	dependencyList       []string
	internalModules      []string
	messageStack         []sanyParseFrame
	expecting            string
	failedLookaheadSizes map[*SanyToken]int
	fairnessHook         *SanySyntaxNode
	numberFlag           bool
	decimalFlag          bool
}

func ParseSanySyntax(file, source string) (*SanySyntaxNode, Diagnostics) {
	node, _, diags := parseSanySyntaxWithDependencies(file, source)
	return node, diags
}

func parseSanySyntaxWithDependencies(file, source string) (node *SanySyntaxNode, dependencies []string, diags Diagnostics) {
	return parseSanySyntaxUsingOutput(file, source, nil)
}

func parseSanySyntaxUsingOutput(file, source string, out SanyOutput) (node *SanySyntaxNode, dependencies []string, diags Diagnostics) {
	parser := &SanyParser{tokenManager: NewSanyTokenManager(file, source), output: out}
	defer func() {
		if failure := recover(); failure != nil {
			switch failure := failure.(type) {
			case *sanyTokenMgrError:
				diags = append(parser.diags, failure.diagnostic)
				dependencies = parser.Dependencies()
			case *sanyParseException:
				diags = append(parser.diags, failure.diagnostic)
				dependencies = parser.Dependencies()
			default:
				panic(failure)
			}
		}
		for _, diagnostic := range diags {
			if diagnostic.SANYParseMessage != "" {
				parser.log(SanyLogError, diagnostic.SANYParseMessage)
			}
		}
	}()
	node = parser.CompilationUnit()
	node.SetLevel(0)
	node.SetParent()
	return node, parser.Dependencies(), parser.diags
}

func (p *SanyParser) log(level SanyLogLevel, format string, args ...string) {
	if p.output != nil {
		p.output.Log(level, format, args...)
	}
}

func (p *SanyParser) Dependencies() []string { return append([]string(nil), p.dependencyList...) }

func (p *SanyParser) addDependency(name string) {
	for _, internal := range p.internalModules {
		if internal == name {
			return
		}
	}
	p.dependencyList = append(p.dependencyList, name)
}

func ParseSanySyntaxModules(file, source string) (modules []*SanySyntaxNode, diags Diagnostics) {
	tokens, lexDiags := SanyTokenize(file, source)
	parser := NewSanyParser(tokens, nil)
	defer func() {
		if failure := recover(); failure != nil {
			if parse, ok := failure.(*sanyParseException); ok {
				diags = append(filterSanyDiagnosticsToModuleSpans(lexDiags, modules), parser.diags...)
				diags = append(diags, parse.diagnostic)
			} else {
				panic(failure)
			}
		}
	}()
	if parser.match(SanyTokenBeginPragma) {
		for !parser.check(SanyTokenEOF) && !parser.atModuleStart() {
			parser.advance()
		}
	}
	for !parser.check(SanyTokenEOF) {
		for !parser.check(SanyTokenEOF) && !parser.atModuleStart() {
			parser.advance()
		}
		if parser.check(SanyTokenEOF) {
			break
		}
		parser.moduleName = ""
		parser.belchDEF()
		module := parser.Module()
		module.SetLevel(0)
		module.SetParent()
		modules = append(modules, module)
	}
	diags = filterSanyDiagnosticsToModuleSpans(lexDiags, modules)
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

func (p *SanyParser) Diagnostics() Diagnostics {
	return append(Diagnostics(nil), p.diags...)
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
	p.belchDEF()
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
	p.beginProduction("Module definition")
	defer p.endProduction()
	stackLevel := len(p.internalModules)
	p.expecting = "---- MODULE"
	begin := p.BeginModule()
	p.expecting = "EXTENDS clause or module body"
	extends := p.Extends()
	p.expecting = "Module body"
	body := p.Body()
	p.expecting = "==== or more Module body"
	end := p.EndModule()
	p.internalModules = p.internalModules[:stackLevel]
	p.internalModules = append(p.internalModules, begin.GetHeirs()[1].Image)
	return NewSanyNode(SanySyntaxNodeKindByName["N_Module"], begin, extends, body, end)
}

func (p *SanyParser) BeginModule() *SanySyntaxNode {
	p.beginProduction("Begin module")
	defer p.endProduction()
	p.expecting = "---- MODULE (beginning of module)"
	if !p.atModuleStart() {
		p.throwParseException([][]SanyTokenKind{{SanyTokenBm0}, {SanyTokenBm1}, {SanyTokenBm2}}, "expected ---- MODULE")
	}
	begin := NewSanyTokenNode(p.advance())
	p.expecting = "Identifier"
	p.reclassifyFieldName()
	name := p.Identifier()
	if name != nil && p.moduleName == "" {
		p.moduleName = name.Image
	}
	p.expecting = "----"
	separator := p.consumeParseToken(SanyTokenSeparator, "expected ---- after module name")
	return NewSanyNode(SanySyntaxNodeKindByName["N_BeginModule"], begin, name, separator)
}

func (p *SanyParser) EndModule() *SanySyntaxNode {
	if !p.check(SanyTokenEndModule) {
		p.throwParseException([][]SanyTokenKind{{SanyTokenEndModule}}, p.parseErrorMessage("==== or more Module body", p.peek()))
	}
	end := p.consume(SanyTokenEndModule, p.parseErrorMessage("==== or more Module body", p.peek()))
	return NewSanyNode(SanySyntaxNodeKindByName["N_EndModule"], end)
}

func (p *SanyParser) Extends() *SanySyntaxNode {
	p.beginProduction("Extends")
	defer p.endProduction()
	var heirs []*SanySyntaxNode
	if p.match(SanyTokenExtends) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
		p.expecting = "Identifier"
		p.reclassifyFieldName()
		name := p.Identifier()
		p.addDependency(name.Image)
		heirs = append(heirs, name)
		p.expecting = "comma or module body"
		for p.match(SanyTokenComma) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()))
			p.expecting = "Identifier"
			p.reclassifyFieldName()
			name := p.Identifier()
			p.addDependency(name.Image)
			heirs = append(heirs, name)
		}
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_Extends"], heirs...)
}

func (p *SanyParser) Body() *SanySyntaxNode {
	p.beginProduction("Module body")
	defer p.endProduction()
	p.expecting = "LOCAL, INSTANCE, PROOF, ASSUMPTION, THEOREM, RECURSIVE, declaration, or definition"
	var heirs []*SanySyntaxNode
	for p.startsBodyItemAt(0) {
		switch {
		case p.check(SanyTokenBm0) || p.check(SanyTokenBm1) || p.check(SanyTokenBm2):
			heirs = append(heirs, p.Module())
			p.belchDEF()
		case p.check(SanyTokenSeparator):
			heirs = append(heirs, NewSanyTokenNode(p.advance()))
		case p.check(SanyTokenVariable):
			heirs = append(heirs, p.VariableDeclaration())
		case p.check(SanyTokenConstant):
			heirs = append(heirs, p.ParamDeclaration())
		case p.check(SanyTokenRecursive):
			heirs = append(heirs, p.Recursive())
		case p.check(SanyTokenInstance) || (p.check(SanyTokenLocal) && p.peekNext().Kind == SanyTokenInstance):
			// Body's two-token Instance lookahead has not run field-name
			// reclassification. With LOCAL, its budget ends at INSTANCE.
			if p.check(SanyTokenInstance) && p.peekNext().Kind != SanyTokenIdentifier {
				p.rememberFailedLookahead(2)
				p.throwParseException([][]SanyTokenKind{{SanyTokenInstance, SanyTokenIdentifier}}, "expected INSTANCE module")
			}
			heirs = append(heirs, p.Instance())
		case (p.check(SanyTokenAssume) || p.check(SanyTokenAssumption)) && (p.tokenAt(1).Kind == SanyTokenDefbreak || p.startsExpressionFirstAt(1)):
			heirs = append(heirs, p.Assumption())
		case (p.check(SanyTokenTheorem) || p.check(SanyTokenProposition)) && (p.startsAssumeProveAt(1) || p.startsExpressionFirstAt(1)):
			heirs = append(heirs, p.Theorem())
		case p.check(SanyTokenUse) || p.check(SanyTokenHide):
			heirs = append(heirs, p.UseOrHide())
		case p.startsOperatorOrFunctionDefinition():
			heirs = append(heirs, p.OperatorOrFunctionDefinition())
		default:
			p.rememberFailedLookahead(2)
			p.throwParseException([][]SanyTokenKind{{SanyTokenDefbreak}}, "expected module body unit")
		}
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_Body"], heirs...)
}

func (p *SanyParser) VariableDeclaration() *SanySyntaxNode {
	p.beginProduction("variable declaration")
	defer p.endProduction()
	keyword := p.consumeParseToken(SanyTokenVariable, "expected VARIABLE declaration")
	p.expecting = "Identifier"
	var one []*SanySyntaxNode
	one = append(one, p.Identifier())
	p.expecting = "comma or module body"
	for p.match(SanyTokenComma) {
		one = append(one, NewSanyTokenNode(p.previous()))
		p.expecting = "Identifier"
		one = append(one, p.Identifier())
	}
	return NewSanySplitNode(SanySyntaxNodeKindByName["N_VariableDeclaration"], []*SanySyntaxNode{keyword}, one)
}

func (p *SanyParser) ParamDeclaration() *SanySyntaxNode {
	p.beginProduction("Parameter declaration")
	defer p.endProduction()
	p.expecting = "CONSTANT"
	heirs := []*SanySyntaxNode{p.ParamSubDecl()}
	p.expecting = "Identifier, operator or _"
	heirs = append(heirs, p.ConstantDeclarationItem())
	p.expecting = ","
	for p.match(SanyTokenComma) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
		p.expecting = "Identifier, operator or _"
		heirs = append(heirs, p.ConstantDeclarationItem())
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_ParamDeclaration"], heirs...)
}

func (p *SanyParser) ParamSubDecl() *SanySyntaxNode {
	p.beginProduction("Parameter declaration item")
	defer p.endProduction()
	keyword := p.consumeParseToken(SanyTokenConstant, "expected CONSTANT declaration")
	return NewSanyNode(SanySyntaxNodeKindByName["N_ConsDecl"], keyword)
}

func (p *SanyParser) ConstantDeclarationItem() *SanySyntaxNode {
	p.beginProduction("Constant declaration items")
	defer p.endProduction()
	p.expecting = "Identifier, _ or prefix op"
	if !p.check(SanyTokenIdentifier) {
		return p.fixDeclaration(true)
	}
	heirs := []*SanySyntaxNode{p.consumeParseToken(SanyTokenIdentifier, "expected identifier")}
	p.expecting = "(, comma, or Module Body"
	// Java jj_2_6(2) accepts only '(' followed by '_'.
	if p.check(SanyTokenLbr) && p.peekNext().Kind != SanyTokenUs {
		p.rememberFailedLookahead(2)
	}
	if p.check(SanyTokenLbr) && p.peekNext().Kind == SanyTokenUs {
		heirs = append(heirs, NewSanyTokenNode(p.advance()))
		p.expecting = "_"
		heirs = append(heirs, p.consumeParseToken(SanyTokenUs, "expected _ in constant declaration"))
		p.expecting = "comma or )"
		for p.match(SanyTokenComma) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()))
			p.expecting = "_"
			heirs = append(heirs, p.consumeParseToken(SanyTokenUs, "expected _ in constant declaration"))
			p.expecting = "comma or )"
		}
		heirs = append(heirs, p.consumeParseToken(SanyTokenRbr, "expected ) in constant declaration"))
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_IdentDecl"], heirs...)
}

func (p *SanyParser) Recursive() *SanySyntaxNode {
	p.beginProduction("Recursive")
	defer p.endProduction()
	p.expecting = "RECURSIVE"
	heirs := []*SanySyntaxNode{p.consumeParseToken(SanyTokenRecursive, "expected RECURSIVE")}
	p.expecting = "Identifier, operator or _"
	heirs = append(heirs, p.ConstantDeclarationItem())
	p.expecting = ","
	for p.match(SanyTokenComma) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
		p.expecting = "Identifier, operator or _"
		heirs = append(heirs, p.ConstantDeclarationItem())
		p.expecting = "`,' or `)'"
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_Recursive"], heirs...)
}

func (p *SanyParser) Instance() *SanySyntaxNode {
	p.beginProduction("Instance")
	defer p.endProduction()
	p.expecting = "LOCAL or instance"
	var zero []*SanySyntaxNode
	if p.match(SanyTokenLocal) {
		zero = append(zero, NewSanyTokenNode(p.previous()))
	}
	inst := p.Instantiation()
	p.expecting = "COMMA or Module Body"
	return NewSanySplitNode(SanySyntaxNodeKindByName["N_Instance"], zero, []*SanySyntaxNode{inst})
}

func (p *SanyParser) Instantiation() *SanySyntaxNode {
	p.beginProduction("NonLocalInstance")
	defer p.endProduction()
	heirs := []*SanySyntaxNode{p.consumeParseToken(SanyTokenInstance, "expected INSTANCE")}
	p.expecting = "Module identifier"
	p.reclassifyFieldName()
	name := p.consumeParseToken(SanyTokenIdentifier, "expected module identifier")
	p.addDependency(name.Image)
	heirs = append(heirs, name)
	p.expecting = "WITH or another definition."
	if p.match(SanyTokenWith) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
		p.expecting = ""
		heirs = append(heirs, p.Substitution())
		p.expecting = ""
		for p.startsFollowingSubstitution() {
			heirs = append(heirs, NewSanyTokenNode(p.advance()))
			p.expecting = ""
			heirs = append(heirs, p.Substitution())
			p.expecting = ""
		}
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_NonLocalInstance"], heirs...)
}

func (p *SanyParser) startsSubstitutionTarget(offset int) bool {
	kind := p.tokenAt(offset).Kind
	// Source postfix, nonexpression-prefix, infix (including Unicode), and
	// identifier alternatives cover this contiguous token interval.
	return kind >= SanyTokenOp57 && kind <= SanyTokenIdentifier
}

// Java jj_2_12(3) scans only comma, target and '<-'. A failed scan leaves
// the comma untouched and participates in the later expected-input length.
func (p *SanyParser) startsFollowingSubstitution() bool {
	if !p.check(SanyTokenComma) {
		return false
	}
	if !p.startsSubstitutionTarget(1) {
		p.rememberFailedLookahead(2)
		return false
	}
	if p.tokenAt(2).Kind != SanyTokenSubstitute {
		p.rememberFailedLookahead(3)
		return false
	}
	return true
}

func (p *SanyParser) Substitution() *SanySyntaxNode {
	p.beginProduction("Substitution")
	defer p.endProduction()
	var target *SanySyntaxNode
	switch {
	case p.check(SanyTokenIdentifier):
		target = p.consumeParseToken(SanyTokenIdentifier, "expected substitution target")
	case p.peek().Kind >= SanyTokenOp76 && p.peek().Kind <= SanyTokenOp116:
		target = sanyOperatorTokenNode("N_NonExpPrefixOp", p.advance())
	case p.peek().Kind >= SanyTokenOp1 && p.peek().Kind < SanyTokenIdentifier:
		target = sanyOperatorTokenNode("N_InfixOp", p.infixOpToken())
	case p.isGrammarPostfixOperator(p.peek()):
		target = sanyOperatorTokenNode("N_PostfixOp", p.advance())
	default:
		p.throwParseException([][]SanyTokenKind{{SanyTokenIdentifier}}, "expected substitution target")
	}
	p.expecting = "<-"
	arrow := p.consumeParseToken(SanyTokenSubstitute, "expected <- in substitution")
	p.expecting = "Expression or Op. Symbol"
	value := p.ExpressionUntilCommaOrBodyBoundary()
	return NewSanyNode(SanySyntaxNodeKindByName["N_Substitution"], target, arrow, value)
}

func (p *SanyParser) Assumption() *SanySyntaxNode {
	p.beginProduction("Assumption")
	defer p.endProduction()
	p.expecting = "ASSUM..."
	if !p.check(SanyTokenAssume) && !p.check(SanyTokenAssumption) {
		p.throwParseException([][]SanyTokenKind{{SanyTokenAssume}, {SanyTokenAssumption}}, "expected ASSUME or ASSUMPTION")
	}
	heirs := []*SanySyntaxNode{NewSanyTokenNode(p.advance())}
	if (p.check(SanyTokenDefbreak) && p.tokenAt(1).Kind == SanyTokenIdentifier) ||
		(p.check(SanyTokenIdentifier) && p.tokenAt(1).Kind == SanyTokenDef) {
		p.match(SanyTokenDefbreak)
		heirs = append(heirs, p.Identifier())
		p.expecting = "=="
		heirs = append(heirs, p.consumeParseToken(SanyTokenDef, "expected == in assumption"))
	} else {
		if p.check(SanyTokenDefbreak) || p.check(SanyTokenIdentifier) {
			p.rememberFailedLookahead(2)
		} else {
			p.rememberFailedLookahead(1)
		}
	}
	p.belchDEF()
	p.expecting = "Expression"
	heirs = append(heirs, p.ExpressionUntilBodyBoundary())
	return NewSanySplitNode(SanySyntaxNodeKindByName["N_Assumption"], nil, heirs)
}

func (p *SanyParser) Theorem() *SanySyntaxNode {
	p.beginProduction("Theorem")
	defer p.endProduction()
	p.expecting = "THEOREM, PROPOSITION"
	if !p.check(SanyTokenTheorem) && !p.check(SanyTokenProposition) {
		p.throwParseException([][]SanyTokenKind{{SanyTokenTheorem}, {SanyTokenProposition}}, "expected THEOREM or PROPOSITION")
	}
	heirs := []*SanySyntaxNode{NewSanyTokenNode(p.advance())}
	p.expecting = "Identifier, Assume-Prove or Expression"
	if p.check(SanyTokenIdentifier) && p.peekNext().Kind == SanyTokenDef {
		heirs = append(heirs, p.Identifier())
		p.expecting = "=="
		heirs = append(heirs, p.consumeParseToken(SanyTokenDef, "expected == in theorem"))
	} else if p.check(SanyTokenIdentifier) {
		p.rememberFailedLookahead(2)
	}
	p.belchDEF()
	if p.startsAssumeProveAt(0) {
		heirs = append(heirs, p.AssumeProve())
	} else if p.startsExpressionLookahead() {
		heirs = append(heirs, p.ExpressionUntilDefinitionBoundary(func(tok *SanyToken) bool {
			return beginsSanyProof(tok)
		}))
	} else {
		p.throwParseException([][]SanyTokenKind{{SanyTokenIdentifier}}, "expected theorem statement")
	}
	if beginsSanyProof(p.peek()) {
		heirs = append(heirs, p.Proof())
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_Theorem"], heirs...)
}

func (p *SanyParser) Proof() *SanySyntaxNode {
	p.pushProofLevel()
	defer p.popProofLevel()
	p.beginProduction("Proof")
	defer p.endProduction()
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
		node := NewSanyNode(SanySyntaxNodeKindByName["N_Proof"], heirs...)
		node.ProofLevel = p.currentProofLevel()
		return node
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
	if p.startsProofStepAt(0) && !p.startsNoOpExtension() {
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
	return p.definition(true, false)
}

func (p *SanyParser) ProofFunctionDefinition() *SanySyntaxNode {
	return p.functionDefinition(p.ExpressionUntilProofBoundary)
}

func (p *SanyParser) ProofOperatorDefinition(lhs *SanySyntaxNode) *SanySyntaxNode {
	heirs := []*SanySyntaxNode{lhs}
	p.expecting = "=="
	heirs = append(heirs, p.consumeParseToken(SanyTokenDef, "expected == in operator definition"))
	p.belchDEF()
	p.expecting = "Expression"
	if lhs.Kind.JavaName() == "N_IdentLHS" {
		p.expecting = "Expression or Instance"
	}
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
	return p.ExpressionUntil(func(tok *SanyToken) bool {
		return p.isProofBoundary(tok)
	})
}

func (p *SanyParser) ExpressionUntilCommaOrProofBoundary() *SanySyntaxNode {
	return p.ExpressionUntil(func(tok *SanyToken) bool {
		return tok.Kind == SanyTokenComma || p.isProofBoundary(tok)
	})
}

func (p *SanyParser) ExpressionUntilAssumeProveBoundary() *SanySyntaxNode {
	return p.ExpressionUntil(func(tok *SanyToken) bool {
		return tok.Kind == SanyTokenComma ||
			tok.Kind == SanyTokenProve ||
			tok.Kind == SanyTokenBoxprove ||
			p.isProofBoundary(tok)
	})
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
	return p.definition(false, false)
}

// definition retains one source Definition production even when LOCAL is
// present. The native caller contexts select their existing expression stops.
func (p *SanyParser) definition(proof, let bool) (node *SanySyntaxNode) {
	p.beginProduction("Definition")
	defer p.endProduction()
	p.expecting = "LOCAL, Identifier or Operator Symbol"
	var local *SanySyntaxNode
	if p.match(SanyTokenLocal) {
		local = NewSanyTokenNode(p.previous())
	}
	defer func() {
		if node != nil && local != nil {
			node.Zero = []*SanySyntaxNode{local}
			node.refreshHeirsAndRange()
		}
	}()
	p.consumeParseToken(SanyTokenDefbreak, "expected beginning of definition")
	p.expecting = "LOCAL, Identifier or Operator Symbol"
	operator := p.OperatorDefinition
	function := p.FunctionDefinition
	if proof {
		operator = p.ProofOperatorDefinition
		function = p.ProofFunctionDefinition
	}
	if let {
		operator = p.LetOperatorDefinition
		function = p.LetFunctionDefinition
	}
	switch {
	case p.check(SanyTokenIdentifier) && p.peekNext().Kind == SanyTokenLsb:
		return function()
	case p.check(SanyTokenIdentifier) && p.isPostfixOperator(p.peekNext()):
		return operator(p.PostfixLHS())
	case p.check(SanyTokenIdentifier) && p.isInfixOperator(p.peekNext()):
		return operator(p.InfixLHS())
	case p.startsModuleDefinitionHeadAt(0):
		return p.ModuleDefinition()
	case p.check(SanyTokenIdentifier) && (p.peekNext().Kind == SanyTokenLbr || p.peekNext().Kind == SanyTokenDef):
		return operator(p.IdentLHS())
	default:
		// Failed definition-head alternatives scan a second token only
		// after their common Identifier. Preserve that error-input span.
		if p.check(SanyTokenIdentifier) {
			p.rememberFailedLookahead(2)
		}
		if !p.startsDefinitionPrefix() {
			p.throwParseException([][]SanyTokenKind{{SanyTokenOp76}}, "expected definition identifier or prefix operator")
		}
		return operator(p.PrefixLHS())
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
	p.expecting = "=="
	heirs = append(heirs, p.consumeParseToken(SanyTokenDef, "expected == in module definition"))
	p.belchDEF()
	p.expecting = "Expression or Instance"
	heirs = append(heirs, p.Instantiation())
	return NewSanySplitNode(SanySyntaxNodeKindByName["N_ModuleDefinition"], nil, heirs)
}

func (p *SanyParser) FunctionDefinition() *SanySyntaxNode {
	return p.functionDefinition(p.ExpressionUntilBodyBoundary)
}

func (p *SanyParser) OperatorDefinition(lhs *SanySyntaxNode) *SanySyntaxNode {
	heirs := []*SanySyntaxNode{lhs}
	p.expecting = "=="
	heirs = append(heirs, p.consumeParseToken(SanyTokenDef, "expected == in operator definition"))
	p.belchDEF()
	p.expecting = "Expression"
	if lhs.Kind.JavaName() == "N_IdentLHS" {
		p.expecting = "Expression or Instance"
	}
	heirs = append(heirs, p.ExpressionUntilDefinitionBoundary(nil))
	return NewSanySplitNode(SanySyntaxNodeKindByName["N_OperatorDefinition"], nil, heirs)
}

func (p *SanyParser) LetOperatorDefinition(lhs *SanySyntaxNode) *SanySyntaxNode {
	heirs := []*SanySyntaxNode{lhs}
	p.expecting = "=="
	heirs = append(heirs, p.consumeParseToken(SanyTokenDef, "expected == in LET definition"))
	p.belchDEF()
	p.expecting = "Expression"
	if lhs.Kind.JavaName() == "N_IdentLHS" {
		p.expecting = "Expression or Instance"
	}
	heirs = append(heirs, p.ExpressionUntilDefinitionBoundary(func(tok *SanyToken) bool {
		return tok.Kind == SanyTokenLetin
	}))
	return NewSanySplitNode(SanySyntaxNodeKindByName["N_OperatorDefinition"], nil, heirs)
}

func (p *SanyParser) IdentLHS() *SanySyntaxNode {
	p.beginProduction("Identifier LHS")
	defer p.endProduction()
	var heirs []*SanySyntaxNode
	heirs = append(heirs, p.Identifier())
	p.expecting = "( or =="
	if p.match(SanyTokenLbr) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
		p.expecting = "Identifier Declaration, prefix op, _ or )"
		heirs = append(heirs, p.IdentDeclOrSomeFixDecl())
		p.expecting = "COMMA or )"
		for p.match(SanyTokenComma) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()))
			p.expecting = "Identifier Declaration, prefix op or _"
			heirs = append(heirs, p.IdentDeclOrSomeFixDecl())
			p.expecting = "COMMA or )"
		}
		heirs = append(heirs, p.consumeParseToken(SanyTokenRbr, "expected ) in operator definition parameters"))
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_IdentLHS"], heirs...)
}

func (p *SanyParser) PrefixLHS() *SanySyntaxNode {
	p.beginProduction("Prefix LHS")
	defer p.endProduction()
	op := p.consumeOperator("expected prefix operator in definition")
	p.expecting = "Identifier"
	id := p.Identifier()
	return NewSanyNode(SanySyntaxNodeKindByName["N_PrefixLHS"], op, id)
}

func (p *SanyParser) InfixLHS() *SanySyntaxNode {
	p.beginProduction("Infix LHS")
	defer p.endProduction()
	left := p.Identifier()
	op := p.consumeOperator("expected infix operator in definition")
	right := p.Identifier()
	return NewSanyNode(SanySyntaxNodeKindByName["N_InfixLHS"], left, op, right)
}

func (p *SanyParser) PostfixLHS() *SanySyntaxNode {
	p.beginProduction("Postfix LHS")
	defer p.endProduction()
	left := p.Identifier()
	op := p.consumeOperator("expected postfix operator in definition")
	return NewSanyNode(SanySyntaxNodeKindByName["N_PostfixLHS"], left, op)
}

func (p *SanyParser) startsDefinitionPrefix() bool {
	return p.startsDefinitionPrefixAt(0)
}

func (p *SanyParser) startsDefinitionPrefixAt(offset int) bool {
	switch p.tokenAt(offset).Kind {
	case SanyTokenOp76, SanyTokenOp26, SanyTokenOp29, SanyTokenOp58,
		SanyTokenCasesep, SanyTokenOp61, SanyTokenOp112, SanyTokenOp113,
		SanyTokenOp114, SanyTokenOp115, SanyTokenOp116:
		return true
	default:
		return false
	}
}

func (p *SanyParser) IdentDeclOrSomeFixDecl() *SanySyntaxNode {
	if p.check(SanyTokenIdentifier) {
		return p.IdentDecl()
	}
	if p.check(SanyTokenUs) || p.startsDefinitionPrefix() {
		return p.SomeFixDecl()
	}
	p.throwParseException([][]SanyTokenKind{{SanyTokenIdentifier}, {SanyTokenUs}, {SanyTokenOp76}}, "expected formal declaration")
	return nil
}

func (p *SanyParser) IdentDecl() *SanySyntaxNode {
	p.beginProduction("Identifier Declation")
	defer p.endProduction()
	heirs := []*SanySyntaxNode{p.Identifier()}
	p.expecting = "( or ..."
	if p.match(SanyTokenLbr) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
		p.expecting = "_"
		heirs = append(heirs, p.consumeParseToken(SanyTokenUs, "expected _ in operator parameter declaration"))
		p.expecting = "COMMA or )"
		for p.match(SanyTokenComma) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()))
			p.expecting = "_"
			heirs = append(heirs, p.consumeParseToken(SanyTokenUs, "expected _ in operator parameter declaration"))
			p.expecting = "COMMA or )"
		}
		heirs = append(heirs, p.consumeParseToken(SanyTokenRbr, "expected ) in operator parameter declaration"))
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_IdentDecl"], heirs...)
}

func (p *SanyParser) SomeFixDecl() *SanySyntaxNode {
	p.beginProduction("Op. Symbol Declaration")
	defer p.endProduction()
	return p.fixDeclaration(false)
}

// ConstantDeclarationItems and SomeFixDecl share token choices and wrappers,
// but each source production owns its distinct message frame.
func (p *SanyParser) fixDeclaration(constant bool) *SanySyntaxNode {
	if p.isNEPrefixOperator(p.peek()) {
		op := sanyOperatorTokenNode("N_NonExpPrefixOp", p.advance())
		p.expecting = "_"
		us := p.consumeParseToken(SanyTokenUs, "expected _ after prefix operator declaration")
		return NewSanyNode(SanySyntaxNodeKindByName["N_PrefixDecl"], op, us)
	}
	if !p.check(SanyTokenUs) {
		p.throwParseException([][]SanyTokenKind{{SanyTokenIdentifier}}, "expected operator declaration")
	}
	left := NewSanyTokenNode(p.advance())
	p.expecting = "infix or postfix operator"
	if constant {
		p.expecting = "prefix or postfix operator"
	}
	if p.isInfixOperator(p.peek()) {
		op := sanyOperatorTokenNode("N_InfixOp", p.infixOpToken())
		p.expecting = "_"
		right := p.consumeParseToken(SanyTokenUs, "expected _ after infix operator declaration")
		return NewSanyNode(SanySyntaxNodeKindByName["N_InfixDecl"], left, op, right)
	}
	if !p.isGrammarPostfixOperator(p.peek()) {
		p.throwParseException([][]SanyTokenKind{{SanyTokenOp57}}, "expected infix or postfix operator declaration")
	}
	op := sanyOperatorTokenNode("N_PostfixOp", p.advance())
	return NewSanyNode(SanySyntaxNodeKindByName["N_PostfixDecl"], left, op)
}

func (p *SanyParser) QuantBound() *SanySyntaxNode {
	return p.QuantBoundUntil(SanyTokenComma, SanyTokenRsb, SanyTokenEOF)
}

func (p *SanyParser) QuantBoundUntil(stopKinds ...SanyTokenKind) *SanySyntaxNode {
	p.beginProduction("Quant Bound")
	defer p.endProduction()
	var heirs []*SanySyntaxNode
	switch p.peek().Kind {
	case SanyTokenLab:
		heirs = append(heirs, p.IdentifierTuple())
	case SanyTokenIdentifier:
		heirs = append(heirs, p.consumeParseToken(SanyTokenIdentifier, "expected bound identifier"))
		for p.match(SanyTokenComma) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()))
			heirs = append(heirs, p.consumeParseToken(SanyTokenIdentifier, "expected bound identifier"))
			p.expecting = ", or \\in"
		}
	default:
		p.throwParseException([][]SanyTokenKind{{SanyTokenLab}, {SanyTokenIdentifier}}, "expected bound identifier or tuple")
	}
	in := p.consumeParseToken(SanyTokenIN, "expected \\in in quantifier bound")
	in.Kind = SanySyntaxNodeKindByName["T_IN"]
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
	p.beginProduction("Identifier tuple")
	defer p.endProduction()
	heirs := []*SanySyntaxNode{p.consumeParseToken(SanyTokenLab, "expected << in identifier tuple")}
	p.expecting = "Identifier or >>"
	if p.check(SanyTokenIdentifier) {
		heirs = append(heirs, p.consumeParseToken(SanyTokenIdentifier, "expected identifier"))
		p.expecting = "COMMA or >>"
		for p.match(SanyTokenComma) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()))
			p.expecting = "COMMA or >>"
			heirs = append(heirs, p.consumeParseToken(SanyTokenIdentifier, "expected identifier"))
			p.expecting = "COMMA or >>"
		}
	}
	heirs = append(heirs, p.consumeParseToken(SanyTokenRab, "expected >> in identifier tuple"))
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
	return p.OpOrExpr(func(tok *SanyToken) bool {
		// A named instance in LET ends its final substitution at the enclosing
		// IN. Java's OpOrExpr leaves that token for LetIn to consume.
		if tok.Kind == SanyTokenComma || tok.Kind == SanyTokenLetin || tok.Kind == SanyTokenEOF || tok.Kind == SanyTokenEndModule {
			return true
		}
		return tok.Begin.Line >= startLine && p.startsBodyItemAt(0)
	})
}

func (p *SanyParser) ExpressionUntil(stop func(*SanyToken) bool) *SanySyntaxNode {
	p.beginProduction("Expression")
	active := true
	defer func() {
		if failure := recover(); failure != nil {
			panic(failure)
		}
		if active {
			p.endProduction()
		}
	}()
	stack := NewSanyOperatorStack()
	stack.moduleName = p.moduleName
	stack.NewStack()
	p.expressionOperand(stack, stop)
	// Source Expression calls epa before finalReduce, so a reduction failure
	// does not retain this production in the residual stack.
	p.endProduction()
	active = false
	expr, err := stack.FinalReduce()
	if err != nil {
		p.throwOperatorStackFailure(err, p.peek().Begin)
	}
	for _, message := range stack.reportedErrors {
		diagnostic := errorAt(p.peek().Begin, "E1301", "could not reduce expression stack")
		diagnostic.SANYParseMessage = message
		p.diags = append(p.diags, diagnostic)
	}
	if expr == nil {
		p.throwReportedParseException(" Couldn't reduce expression stack.", p.peek().Begin, "E1301", "could not reduce expression stack")
	}
	if err := stack.PopStack(); err != nil {
		panic(err)
	}
	return expr
}

// expressionOperand is the prefix sequence followed by OpenExpression or
// ExtendableExpr in the source grammar. The latter pushes onto this same stack.
func (p *SanyParser) expressionOperand(stack *SanyOperatorStack, stop func(*SanyToken) bool) {
	for p.startsExpressionPrefix() && p.aboveCurrentJunction() {
		tok := p.advance()
		op, _ := GetSanyOperator(tok.Image)
		stack.Push(p.genericOperatorNode(tok, op), &op)
		if err := stack.ReduceStack(); err != nil {
			p.throwOperatorStackFailure(err, tok.Begin)
		}
	}
	if p.startsOpenExpression() && p.aboveCurrentJunction() {
		stack.Push(p.OpenExpression(stop), nil)
		return
	}
	if !p.aboveCurrentJunction() {
		p.throwParseException([][]SanyTokenKind{{SanyTokenIdentifier}}, "expected properly indented expression")
	}
	p.ExtendableExpr(stack, stop)
}

func (p *SanyParser) startsExpressionPrefix() bool {
	switch p.peek().Kind {
	case SanyTokenOp26, SanyTokenOp29, SanyTokenOp58, SanyTokenCasesep,
		SanyTokenOp61, SanyTokenOp112, SanyTokenOp113, SanyTokenOp114,
		SanyTokenOp115, SanyTokenOp116, SanyTokenOp77:
		return true
	default:
		return false
	}
}

func (p *SanyParser) aboveCurrentJunction() bool {
	return p.junctionListContext.isAboveCurrent(p.peek().Begin.Column)
}

// ExtendableExpr ports the source operand, postfix-extension loop, and optional
// infix or label continuation. It never consumes an unrelated following token.
func (p *SanyParser) ExtendableExpr(stack *SanyOperatorStack, stop func(*SanyToken) bool) {
	p.beginProduction("ExtendableExpr")
	defer p.endProduction()
	if IsSanyJunctionBullet(p.peek().Kind) && stack.PreInEmptyTop() {
		stack.Push(p.JunctionList(stop), nil)
	} else if p.startsLabelAt(0) {
		stack.Push(p.LabelExpression(stop, stack.TopOperator()), nil)
	} else {
		switch p.peek().Kind {
		case SanyTokenWF, SanyTokenSF:
			stack.Push(p.FairnessExpr(), nil)
		case SanyTokenLbr, SanyTokenLbc, SanyTokenLab, SanyTokenLsb:
			stack.Push(p.PrimitiveExpression(), nil)
		default:
			if !p.startsPrimitiveExp() {
				p.throwParseException([][]SanyTokenKind{{SanyTokenIdentifier}}, "expected expression")
			}
			stack.Push(p.PrimitiveExp(), nil)
		}
	}
	for p.aboveCurrentJunction() && !stop(p.peek()) {
		switch {
		case p.check(SanyTokenOp57) || p.check(SanyTokenOp68) || p.check(SanyTokenOp69) || p.check(SanyTokenOp70):
			tok := p.advance()
			op, _ := GetSanyOperator(tok.Image)
			stack.Push(p.genericOperatorNode(tok, op), &op)
		case p.check(SanyTokenDot):
			middle := NewSanyTokenNode(p.advance())
			p.reclassifyFieldName()
			if !p.aboveCurrentJunction() {
				p.throwParseException([][]SanyTokenKind{{SanyTokenIdentifier}}, "expected properly indented record field")
			}
			right := p.consumeParseToken(SanyTokenIdentifier, "expected record field identifier")
			if err := stack.ReduceRecord(middle, right); err != nil {
				p.throwOperatorStackFailure(err, middle.Range.Begin)
			}
			continue
		case p.check(SanyTokenLsb):
			p.expecting = "function argument"
			node := p.SBracketCases()
			if node.Kind.JavaName() == "N_FcnAppl" {
				op, _ := GetSanyOperator("[")
				stack.Push(node, &op)
			} else {
				stack.Push(node, nil)
			}
		default:
			goto continuation
		}
		if err := stack.ReduceStack(); err != nil {
			p.throwOperatorStackFailure(err, p.previous().End)
		}
	}
continuation:
	if p.aboveCurrentJunction() && !stop(p.peek()) && p.isGrammarInfixOperator(p.peek()) {
		tok := p.infixOpToken()
		op, _ := GetSanyOperator(tok.Image)
		stack.Push(p.genericOperatorNode(tok, op), &op)
		if err := stack.ReduceStack(); err != nil {
			p.throwOperatorStackFailure(err, tok.Begin)
		}
		p.expressionOperand(stack, stop)
	}
}

// Source ExtendableExpr performs one-token PrimitiveExp lookahead before
// entering that production, so an invalid start retains only ExtendableExpr.
func (p *SanyParser) startsPrimitiveExp() bool {
	switch p.peek().Kind {
	case SanyTokenNumberLiteral, SanyTokenStringLiteral, SanyTokenIdentifier, SanyTokenProofsteplexeme, SanyTokenProofimplicitsteplexeme:
		return true
	default:
		return p.isGrammarInfixOperator(p.peek()) || p.isGrammarPostfixOperator(p.peek()) || p.check(SanyTokenOp76)
	}
}

// PrimitiveExp covers the primitive (non-parenthesized) operand alternatives.
func (p *SanyParser) PrimitiveExp() *SanySyntaxNode {
	p.beginProduction("Primitive expression")
	defer p.endProduction()
	switch p.peek().Kind {
	case SanyTokenNumberLiteral:
		return p.Number()
	case SanyTokenStringLiteral:
		return p.String()
	case SanyTokenIdentifier, SanyTokenProofsteplexeme, SanyTokenProofimplicitsteplexeme:
		return p.primitiveSelectorExpr(p.NoOpExtensionBase())
	default:
		if _, ok := GetSanyOperator(p.peek().Image); ok && (p.isGrammarInfixOperator(p.peek()) || p.isGrammarPostfixOperator(p.peek()) || p.check(SanyTokenOp76)) {
			return p.primitiveSelectorExpr(p.BangOperatorSelector())
		}
		p.throwParseException([][]SanyTokenKind{{SanyTokenIdentifier}, {SanyTokenNumberLiteral}, {SanyTokenStringLiteral}}, "expected expression")
		return nil
	}
}

// OpOrExpr is used only for operator arguments and substitution values.
func (p *SanyParser) OpOrExpr(stop func(*SanyToken) bool) *SanySyntaxNode {
	if !p.startsOpOrExprAt(0) {
		p.throwParseException([][]SanyTokenKind{{SanyTokenIdentifier}}, "expected operator or expression")
	}
	if p.check(SanyTokenLambda) {
		return p.Lambda(stop)
	}
	if p.startsOperatorReference(stop) && p.peekNext().Kind != SanyTokenLbr {
		return p.OperatorReference()
	}
	return p.ExpressionUntil(stop)
}

func (p *SanyParser) JunctionList(stop func(*SanyToken) bool) *SanySyntaxNode {
	firstBullet := p.peek()
	p.junctionListContext.startNewJunctionList(firstBullet.Begin.Column, firstBullet.Kind)
	p.beginProduction("AND-OR Junction")
	defer p.endProduction()
	listKind, itemKind := "N_ConjList", "N_ConjItem"
	if firstBullet.Kind == SanyTokenOR {
		listKind, itemKind = "N_DisjList", "N_DisjItem"
	}
	items := []*SanySyntaxNode{p.junctionItem(stop, itemKind)}
	for p.junctionListContext.isNewBullet(p.peek().Begin.Column, p.peek().Kind) {
		items = append(items, p.junctionItem(stop, itemKind))
	}
	// Java does not terminate the context when a production throws.
	p.junctionListContext.terminateCurrentJunctionList()
	list := NewSanyNode(SanySyntaxNodeKindByName[listKind], items...)
	list.JunctionList = true
	return list
}

func (p *SanyParser) junctionItem(stop func(*SanyToken) bool, itemKind string) *SanySyntaxNode {
	p.beginProduction("Junction Item")
	active := true
	defer func() {
		if failure := recover(); failure != nil {
			panic(failure)
		}
		if active {
			p.endProduction()
		}
	}()
	if !p.check(SanyTokenAND) && !p.check(SanyTokenOR) {
		p.throwParseException([][]SanyTokenKind{{SanyTokenOR}, {SanyTokenAND}}, "expected junction bullet")
	}
	bullet := NewSanyTokenNode(p.advance())
	expression := p.ExpressionUntil(stop)
	// JuncItem calls epa before constructing the item and checking descendants.
	p.endProduction()
	active = false
	item := NewSanyNode(SanySyntaxNodeKindByName[itemKind], bullet, expression)
	p.checkJunctionIndentation(expression, item)
	return item
}

func (p *SanyParser) currentJunctionColumn() (int, bool) {
	current, exists := p.junctionListContext.current()
	return current.column, exists
}

// checkIndentation visits descendants, but stops at nested junction lists.
func (p *SanyParser) checkJunctionIndentation(node, item *SanySyntaxNode) {
	for _, child := range node.GetHeirs() {
		kind := child.Kind.JavaName()
		if kind == "N_ConjList" || kind == "N_DisjList" {
			continue
		}
		if !p.junctionListContext.isAboveCurrent(child.Range.Begin.Column) {
			message := "Item at " + p.junctionLocation(child.Range) +
				" is not properly indented inside conjunction or " +
				" disjunction list item at " + p.junctionLocation(item.Range)
			p.throwReportedParseException(message, child.Range.Begin, "E1300", message)
		}
		p.checkJunctionIndentation(child, item)
	}
}

func (p *SanyParser) junctionLocation(location SanyRange) string {
	return "line " + strconv.Itoa(location.Begin.Line) + ", col " + strconv.Itoa(location.Begin.Column) +
		" to line " + strconv.Itoa(location.End.Line) + ", col " + strconv.Itoa(location.End.Column) + " of module " + p.moduleName
}

func (p *SanyParser) startsOperatorReference(stop func(*SanyToken) bool) bool {
	if !p.aboveCurrentJunction() {
		return false
	}
	if !p.startsExpressionPrefix() && !p.check(SanyTokenOp76) && !p.isGrammarInfixOperator(p.peek()) && !p.isGrammarPostfixOperator(p.peek()) {
		return false
	}
	switch p.peekNext().Kind {
	case SanyTokenComma, SanyTokenRbr, SanyTokenDefbreak, SanyTokenLocal,
		SanyTokenInstance, SanyTokenTheorem, SanyTokenAssume, SanyTokenEndModule,
		SanyTokenSeparator, SanyTokenBm0, SanyTokenAssumption, SanyTokenConstant,
		SanyTokenVariable, SanyTokenRecursive:
		return true
	default:
		return false
	}
}

func (p *SanyParser) isGrammarInfixOperator(token *SanyToken) bool {
	return token.Kind >= SanyTokenOp1 && token.Kind <= SanyTokenOp119 && p.isInfixOperator(token)
}

func (p *SanyParser) isGrammarPostfixOperator(token *SanyToken) bool {
	return token.Kind == SanyTokenOp57 || token.Kind == SanyTokenOp68 || token.Kind == SanyTokenOp69 || token.Kind == SanyTokenOp70
}

func (p *SanyParser) OperatorReference() *SanySyntaxNode {
	tok := p.advance()
	op, ok := GetSanyOperator(tok.Image)
	if !ok {
		return NewSanyTokenNode(tok)
	}
	if tok.Image == "\\X" || tok.Image == "\\times" {
		p.add(tok.Begin, "E1300", tok.Image+" may not be used as an infix operator")
	}
	node := p.genericOperatorReferenceNode(tok, op)
	if (op.IsInfix() || op.IsPostfix()) && p.check(SanyTokenLbr) {
		node.AddHeir(p.OpArgs())
	}
	return node
}

func (p *SanyParser) LabelExpression(stop func(*SanyToken) bool, stackOp *SanyOperatorInfo) *SanySyntaxNode {
	label := p.LabelName()
	colon := p.consume(SanyTokenColoncolon, "expected :: after label")
	expr := p.ExpressionUntil(stop)
	if !p.labelDoesNotChangeParse(expr, stackOp) {
		p.add(label.Range.Begin, "E1300", "removing label would change expression parsing")
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_Label"], label, colon, expr)
}

func (p *SanyParser) LabelName() *SanySyntaxNode {
	if p.startsNoOpExtension() {
		return p.primitiveSelectorExpr(p.NoOpExtensionBase())
	}
	heirs := []*SanySyntaxNode{
		NewSanyNode(SanySyntaxNodeKindByName["N_IdPrefix"]),
		p.Identifier(),
	}
	if p.check(SanyTokenLbr) {
		heirs = append(heirs, p.OpArgs())
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_GeneralId"], heirs...)
}

func (p *SanyParser) labelDoesNotChangeParse(expr *SanySyntaxNode, stackOp *SanyOperatorInfo) bool {
	if expr == nil || stackOp == nil {
		return true
	}
	labelOp, ok := sanyLabelExpressionOperator(expr)
	if !ok {
		return true
	}
	return SanyOperatorPrec(*stackOp, labelOp)
}

func sanyLabelExpressionOperator(expr *SanySyntaxNode) (SanyOperatorInfo, bool) {
	if expr == nil {
		return SanyOperatorInfo{}, false
	}
	switch expr.Kind.JavaName() {
	case "N_InfixExpr":
		heirs := expr.GetHeirs()
		if len(heirs) >= 2 {
			return GetSanyOperator(sanyOperatorImage(heirs[1]))
		}
	case "N_PostfixExpr":
		heirs := expr.GetHeirs()
		if len(heirs) >= 2 {
			return GetSanyOperator(sanyOperatorImage(heirs[1]))
		}
	}
	return SanyOperatorInfo{}, false
}

func (p *SanyParser) startsOpenExpression() bool {
	switch p.peek().Kind {
	case SanyTokenIf, SanyTokenForall, SanyTokenExists, SanyTokenTExists, SanyTokenTForall, SanyTokenLet, SanyTokenCase, SanyTokenChoose:
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
	p.fairnessHook = nil
	p.beginProduction("Fairness Expression")
	active := true
	defer func() {
		if failure := recover(); failure != nil {
			panic(failure)
		}
		if active {
			p.endProduction()
		}
	}()
	heirs := make([]*SanySyntaxNode, 5)
	if p.check(SanyTokenWF) || p.check(SanyTokenSF) {
		heirs[0] = NewSanyTokenNode(p.advance())
	} else {
		p.throwParseException([][]SanyTokenKind{{SanyTokenWF}, {SanyTokenSF}}, "expected WF_ or SF_")
	}
	expr := p.ReducedExpression()
	if p.startsFairnessAction() {
		heirs[1], expr = expr, nil
		heirs[2] = p.consumeParseToken(SanyTokenLbr, "expected ( in fairness expression")
		heirs[3] = p.ExpressionUntil(func(tok *SanyToken) bool {
			return tok.Kind == SanyTokenRbr || tok.Kind == SanyTokenEOF
		})
		heirs[4] = p.consumeParseToken(SanyTokenRbr, "expected ) in fairness expression")
	}
	// Java calls epa before linking arguments and reporting structure errors.
	p.endProduction()
	active = false
	if expr != nil {
		if p.fairnessHook == nil {
			pos := heirs[0].Range.Begin
			p.throwReportedParseException("Ill-structured fairness expression at line "+strconv.Itoa(pos.Line)+", column "+strconv.Itoa(pos.Column), pos, "E1300", "ill-structured fairness expression")
		}
		parameters := p.fairnessHook.GetHeirs()
		if len(parameters) != 3 {
			p.reportFairnessParseError(heirs[0], "")
			return expr
		}
		heirs[1], heirs[2], heirs[3], heirs[4] = expr, parameters[0], parameters[1], parameters[2]
	} else if heirs[1].Kind.JavaName() == "N_GeneralId" && p.fairnessHook != nil {
		heirs[1] = NewSanyNode(SanySyntaxNodeKindByName["N_OpApplication"], heirs[1], p.fairnessHook)
	} else {
		switch heirs[1].Kind.JavaName() {
		case "N_Tuple", "N_ParenExpr", "N_SetEnumerate", "N_SubsetOf", "N_SetOfAll",
			"N_SetOfFcns", "N_RcdConstructor", "N_SetOfRcds", "N_Except", "N_FcnConst", "N_ActionExpr":
		default:
			p.reportFairnessParseError(heirs[0], ": could not link arguments")
			return heirs[1]
		}
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_FairnessExpr"], heirs...)
}

func (p *SanyParser) startsFairnessAction() bool {
	if !p.check(SanyTokenLbr) {
		return false
	}
	if p.tokenAt(1).Kind != SanyTokenLambda && p.startsOpOrExprAt(1) {
		return true
	}
	p.rememberFailedLookahead(2)
	return false
}

func (p *SanyParser) reportFairnessParseError(token *SanySyntaxNode, suffix string) {
	pos := token.Range.Begin
	message := "Error in fairness expression at " + strconv.Itoa(pos.Line) + ": " + strconv.Itoa(pos.Column) + suffix + "\n"
	diagnostic := errorAt(pos, "E1300", "%s", strings.TrimSpace(message))
	diagnostic.SANYParseMessage = message
	p.diags = append(p.diags, diagnostic)
}

func (p *SanyParser) FairnessSubscript() *SanySyntaxNode {
	return p.ReducedExpression()
}

func (p *SanyParser) LetIn(stop func(*SanyToken) bool) *SanySyntaxNode {
	// Preserve Java LetIn's original production name.
	p.beginProduction("Case Other Arm")
	defer p.endProduction()
	let := p.consumeParseToken(SanyTokenLet, "expected LET")
	defs := p.LetDefinitions()
	in := p.consumeParseToken(SanyTokenLetin, "expected IN in LET expression")
	body := p.ExpressionUntil(stop)
	return NewSanyNode(SanySyntaxNodeKindByName["N_LetIn"], let, defs, in, body)
}

func (p *SanyParser) LetDefinitions() *SanySyntaxNode {
	p.beginProduction("Let Definitions")
	defer p.endProduction()
	var heirs []*SanySyntaxNode
	for {
		switch p.peek().Kind {
		case SanyTokenLocal, SanyTokenDefbreak:
			heirs = append(heirs, p.LetOperatorOrFunctionDefinition())
		case SanyTokenRecursive:
			heirs = append(heirs, p.Recursive())
		default:
			p.throwParseException([][]SanyTokenKind{{SanyTokenLocal}, {SanyTokenDefbreak}, {SanyTokenRecursive}}, "expected LET definition")
		}
		kind := p.peek().Kind
		if kind != SanyTokenLocal && kind != SanyTokenDefbreak && kind != SanyTokenRecursive {
			break
		}
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_LetDefinitions"], heirs...)
}

func (p *SanyParser) LetOperatorOrFunctionDefinition() *SanySyntaxNode {
	return p.definition(false, true)
}

func (p *SanyParser) LetFunctionDefinition() *SanySyntaxNode {
	return p.functionDefinition(func() *SanySyntaxNode {
		return p.ExpressionUntilDefinitionBoundary(func(tok *SanyToken) bool { return tok.Kind == SanyTokenLetin })
	})
}

// All caller contexts share OperatorOrFunctionDefinition's function branch.
func (p *SanyParser) functionDefinition(body func() *SanySyntaxNode) *SanySyntaxNode {
	heirs := []*SanySyntaxNode{p.Identifier()}
	p.expecting = "["
	heirs = append(heirs, p.consumeParseToken(SanyTokenLsb, "expected [ in function definition"))
	p.expecting = "Identifier"
	heirs = append(heirs, p.QuantBound())
	p.expecting = "COMMA or ]"
	for p.match(SanyTokenComma) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
		p.expecting = "Identifier"
		heirs = append(heirs, p.QuantBound())
	}
	heirs = append(heirs, p.consumeParseToken(SanyTokenRsb, "expected ] in function definition"))
	p.expecting = "=="
	heirs = append(heirs, p.consumeParseToken(SanyTokenDef, "expected == in function definition"))
	p.belchDEF()
	p.expecting = "Expression"
	heirs = append(heirs, body())
	return NewSanySplitNode(SanySyntaxNodeKindByName["N_FunctionDefinition"], nil, heirs)
}

func (p *SanyParser) Case(stop func(*SanyToken) bool) *SanySyntaxNode {
	p.beginProduction("CASE Expression")
	defer p.endProduction()
	heirs := []*SanySyntaxNode{p.consumeParseToken(SanyTokenCase, "expected CASE"), p.CaseArm(stop)}
	for p.caseSeparatorIsAboveCurrentJunction() && p.tokenAt(1).Kind != SanyTokenOther {
		heirs = append(heirs, p.consumeParseToken(SanyTokenCasesep, "expected []"), p.CaseArm(stop))
	}
	if p.caseSeparatorIsAboveCurrentJunction() {
		heirs = append(heirs, p.consumeParseToken(SanyTokenCasesep, "expected []"), p.OtherArm(stop))
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_Case"], heirs...)
}

func (p *SanyParser) caseSeparatorIsAboveCurrentJunction() bool {
	return p.check(SanyTokenCasesep) && p.junctionListContext.isAboveCurrent(p.peek().Begin.Column)
}

func (p *SanyParser) CaseArm(stop func(*SanyToken) bool) *SanySyntaxNode {
	p.beginProduction("Case Arm")
	defer p.endProduction()
	test := p.ExpressionUntil(func(tok *SanyToken) bool {
		return tok.Kind == SanyTokenArrow || tok.Kind == SanyTokenEOF
	})
	arrow := p.consumeParseToken(SanyTokenArrow, "expected -> in CASE arm")
	value := p.ExpressionUntil(func(tok *SanyToken) bool {
		if tok.Kind == SanyTokenCasesep || tok.Kind == SanyTokenEOF || tok.Kind == SanyTokenEndModule {
			return true
		}
		return stop != nil && stop(tok)
	})
	return NewSanyNode(SanySyntaxNodeKindByName["N_CaseArm"], test, arrow, value)
}

func (p *SanyParser) OtherArm(stop func(*SanyToken) bool) *SanySyntaxNode {
	p.beginProduction("Case Other Arm")
	defer p.endProduction()
	other := p.consumeParseToken(SanyTokenOther, "expected OTHER")
	arrow := p.consumeParseToken(SanyTokenArrow, "expected -> in OTHER arm")
	value := p.ExpressionUntil(func(tok *SanyToken) bool {
		if tok.Kind == SanyTokenEOF || tok.Kind == SanyTokenEndModule {
			return true
		}
		return stop != nil && stop(tok)
	})
	return NewSanyNode(SanySyntaxNodeKindByName["N_OtherArm"], other, arrow, value)
}

func (p *SanyParser) UnboundOrBoundChoose(stop func(*SanyToken) bool) *SanySyntaxNode {
	p.beginProduction("(Un)Bounded Choose")
	defer p.endProduction()
	choose := p.consumeParseToken(SanyTokenChoose, "expected CHOOSE")
	var intro *SanySyntaxNode
	switch p.peek().Kind {
	case SanyTokenIdentifier:
		intro = p.consumeParseToken(SanyTokenIdentifier, "expected CHOOSE identifier")
	case SanyTokenLab:
		intro = p.IdentifierTuple()
	default:
		p.throwParseException([][]SanyTokenKind{{SanyTokenIdentifier}, {SanyTokenLab}}, "expected CHOOSE identifier or tuple")
	}
	maybe := p.MaybeBound()
	colon := p.consumeParseToken(SanyTokenColon, "expected : in CHOOSE expression")
	body := p.ExpressionUntil(stop)
	return NewSanyNode(SanySyntaxNodeKindByName["N_UnboundOrBoundChoose"], choose, intro, maybe, colon, body)
}

func (p *SanyParser) MaybeBound() *SanySyntaxNode {
	p.beginProduction("Domain binding")
	defer p.endProduction()
	if !p.match(SanyTokenIN) {
		return NewSanyNode(SanySyntaxNodeKindByName["N_MaybeBound"])
	}
	in := NewSanyTokenNode(p.previous())
	in.Kind = SanySyntaxNodeKindByName["T_IN"]
	p.expecting = "Expression"
	expr := p.ExpressionUntil(func(tok *SanyToken) bool { return tok.Kind == SanyTokenColon || tok.Kind == SanyTokenEOF })
	return NewSanyNode(SanySyntaxNodeKindByName["N_MaybeBound"], in, expr)
}

func (p *SanyParser) Lambda(stop func(*SanyToken) bool) *SanySyntaxNode {
	p.beginProduction("Lambda")
	defer p.endProduction()
	heirs := []*SanySyntaxNode{p.consumeParseToken(SanyTokenLambda, "expected LAMBDA")}
	p.expecting = "Identifier"
	heirs = append(heirs, p.consumeParseToken(SanyTokenIdentifier, "expected LAMBDA identifier"))
	p.expecting = "`,' or `:'"
	for p.match(SanyTokenComma) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
		p.expecting = "Identifier"
		heirs = append(heirs, p.consumeParseToken(SanyTokenIdentifier, "expected LAMBDA identifier"))
		p.expecting = "`,' or `:'"
	}
	heirs = append(heirs, p.consumeParseToken(SanyTokenColon, "expected : in LAMBDA expression"))
	p.expecting = "Expression"
	heirs = append(heirs, p.ExpressionUntil(stop))
	return NewSanyNode(SanySyntaxNodeKindByName["N_Lambda"], heirs...)
}

func (p *SanyParser) IfThenElse(stop func(*SanyToken) bool) *SanySyntaxNode {
	p.beginProduction("IF THEN ELSE")
	defer p.endProduction()
	var heirs []*SanySyntaxNode
	heirs = append(heirs, p.consumeParseToken(SanyTokenIf, "expected IF"))
	heirs = append(heirs, p.ExpressionUntil(func(tok *SanyToken) bool {
		return tok.Kind == SanyTokenThen || tok.Kind == SanyTokenEOF
	}))
	heirs = append(heirs, p.consumeParseToken(SanyTokenThen, "expected THEN"))
	heirs = append(heirs, p.ExpressionUntil(func(tok *SanyToken) bool {
		return tok.Kind == SanyTokenElse || tok.Kind == SanyTokenEOF
	}))
	heirs = append(heirs, p.consumeParseToken(SanyTokenElse, "expected ELSE"))
	heirs = append(heirs, p.ExpressionUntil(stop))
	return NewSanyNode(SanySyntaxNodeKindByName["N_IfThenElse"], heirs...)
}

// Java jj_2_41(MAX_VALUE) accepts Identifier (',' Identifier)* ':'.
func (p *SanyParser) startsUnboundQuantifier() bool {
	if p.tokenAt(0).Kind != SanyTokenIdentifier {
		p.rememberFailedLookahead(1)
		return false
	}
	at := 1
	for p.tokenAt(at).Kind == SanyTokenComma {
		if p.tokenAt(at+1).Kind != SanyTokenIdentifier {
			p.rememberFailedLookahead(at + 2)
			return false
		}
		at += 2
	}
	if p.tokenAt(at).Kind != SanyTokenColon {
		p.rememberFailedLookahead(at + 1)
		return false
	}
	return true
}

func (p *SanyParser) SomeQuant(stop func(*SanyToken) bool) *SanySyntaxNode {
	p.beginProduction("Quantified form")
	defer p.endProduction()
	var heirs []*SanySyntaxNode
	if p.match(SanyTokenExists) || p.match(SanyTokenForall) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
	} else {
		p.throwParseException([][]SanyTokenKind{{SanyTokenExists}, {SanyTokenForall}}, "expected quantified expression")
	}
	kind := SanySyntaxNodeKindByName["N_UnboundQuant"]
	if p.startsUnboundQuantifier() {
		heirs = append(heirs, p.consumeParseToken(SanyTokenIdentifier, "expected quantified identifier"))
		for p.match(SanyTokenComma) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()))
			heirs = append(heirs, p.consumeParseToken(SanyTokenIdentifier, "expected quantified identifier"))
		}
	} else {
		kind = SanySyntaxNodeKindByName["N_BoundQuant"]
		if !p.startsQuantBoundIntro() {
			p.throwParseException([][]SanyTokenKind{{SanyTokenLab}, {SanyTokenIdentifier}}, "expected quantified bound")
		}
		heirs = append(heirs, p.QuantBoundUntil(SanyTokenComma, SanyTokenColon, SanyTokenEOF))
		for p.match(SanyTokenComma) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()))
			heirs = append(heirs, p.QuantBoundUntil(SanyTokenComma, SanyTokenColon, SanyTokenEOF))
		}
	}
	heirs = append(heirs, p.consumeParseToken(SanyTokenColon, "expected : in quantified expression"))
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
	p.beginProduction("Bound Quantified Expression")
	defer p.endProduction()
	var heirs []*SanySyntaxNode
	if p.match(SanyTokenTExists) || p.match(SanyTokenTForall) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
	} else {
		p.throwParseException([][]SanyTokenKind{{SanyTokenTExists}, {SanyTokenTForall}}, "expected temporal quantified expression")
	}
	heirs = append(heirs, p.consumeParseToken(SanyTokenIdentifier, "expected temporal quantified identifier"))
	for p.match(SanyTokenComma) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
		heirs = append(heirs, p.consumeParseToken(SanyTokenIdentifier, "expected temporal quantified identifier"))
	}
	heirs = append(heirs, p.consumeParseToken(SanyTokenColon, "expected : in temporal quantified expression"))
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
	p.beginProduction("String")
	defer p.endProduction()
	node := p.consumeParseToken(SanyTokenStringLiteral, "expected string literal")
	node.Kind = SanySyntaxNodeKindByName["N_String"]
	node.Image = sanyReduceString(node.Image)
	return node
}

func (p *SanyParser) Number() *SanySyntaxNode {
	first := p.consumeParseToken(SanyTokenNumberLiteral, "expected number literal")
	if p.check(SanyTokenDot) {
		if p.tokenAt(1).Kind == SanyTokenNumberLiteral {
			dot := p.consumeParseToken(SanyTokenDot, "expected decimal point")
			second := p.consumeParseToken(SanyTokenNumberLiteral, "expected number literal after decimal point")
			p.decimalFlag = true
			return NewSanyNode(SanySyntaxNodeKindByName["N_Real"], first, dot, second)
		}
		p.rememberFailedLookahead(2)
	} else {
		p.rememberFailedLookahead(1)
	}
	p.numberFlag = true
	return NewSanyNode(SanySyntaxNodeKindByName["N_Number"], first)
}

func (p *SanyParser) ParenExpr() *SanySyntaxNode {
	left := p.consumeParseToken(SanyTokenLbr, "expected (")
	expr := p.ExpressionUntil(func(tok *SanyToken) bool {
		return tok.Kind == SanyTokenRbr || tok.Kind == SanyTokenEOF
	})
	right := p.consumeParseToken(SanyTokenRbr, "expected )")
	return NewSanyNode(SanySyntaxNodeKindByName["N_ParenExpr"], left, expr, right)
}

func (p *SanyParser) BraceCases() *SanySyntaxNode {
	p.beginProduction("Some { } form")
	defer p.endProduction()
	heirs := []*SanySyntaxNode{p.consumeParseToken(SanyTokenLbc, "expected {")}
	kind := "N_SetEnumerate"
	var held *SanySyntaxNode
	readExpression := func() *SanySyntaxNode {
		return p.ExpressionUntil(func(tok *SanyToken) bool {
			return tok.Kind == SanyTokenComma || tok.Kind == SanyTokenColon || tok.Kind == SanyTokenRbc || tok.Kind == SanyTokenEOF
		})
	}
	readElements := func() {
		for p.match(SanyTokenComma) {
			if held != nil {
				heirs = append(heirs, held)
				held = nil
			}
			heirs = append(heirs, NewSanyTokenNode(p.previous()), readExpression())
		}
	}
	readBounds := func() {
		heirs = append(heirs, p.QuantBound())
		for p.match(SanyTokenComma) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()), p.QuantBound())
		}
	}
	matchedFunctionHead := p.matchFcnConst()
	if matchedFunctionHead || p.startsExpressionLookahead() {
		if matchedFunctionHead {
			var intro *SanySyntaxNode
			if p.check(SanyTokenLab) {
				intro = p.IdentifierTuple()
			} else {
				intro = p.Identifier()
			}
			p.expecting = "\\in"
			in := p.consumeParseToken(SanyTokenIN, "expected \\in in set form")
			in.Kind = SanySyntaxNodeKindByName["T_IN"]
			domain := readExpression()
			p.expecting = "':', ',' or '}'"
			identifier := NewSanyNode(SanySyntaxNodeKindByName["N_GeneralId"], NewSanyNode(SanySyntaxNodeKindByName["N_IdPrefix"]), intro)
			operator := NewSanyNode(SanySyntaxNodeKindByName["N_GenInfixOp"], NewSanyNode(SanySyntaxNodeKindByName["N_IdPrefix"]), in)
			held = NewSanyNode(SanySyntaxNodeKindByName["N_InfixExpr"], identifier, operator, domain)
			if p.match(SanyTokenColon) {
				held = nil
				kind = "N_SubsetOf"
				heirs = append(heirs, intro, in, domain, NewSanyTokenNode(p.previous()), readExpression())
			} else {
				readElements()
			}
		} else if p.braceSimpleHead(SanyTokenComma) {
			heirs = append(heirs, readExpression())
			readElements()
		} else if p.braceSimpleHead(SanyTokenColon) {
			kind = "N_SetOfAll"
			heirs = append(heirs, readExpression(), p.consumeParseToken(SanyTokenColon, "expected : in set comprehension"))
			readBounds()
		} else if p.startsExpressionLookahead() {
			expression := readExpression()
			heirs = append(heirs, expression)
			if p.match(SanyTokenColon) {
				colon := p.previous()
				kind = "N_SetOfAll"
				children := expression.GetHeirs()
				if expression.Kind.JavaName() == "N_InfixExpr" && len(children) > 1 {
					operator := children[1].GetHeirs()
					if len(operator) > 1 && operator[1].Image == "\\in" {
						message := "Form {a \\in b : c \\in d }, at line " + strconv.Itoa(colon.Begin.Line) + ", is not allowed"
						p.throwReportedParseException(message, colon.Begin, "E1300", message)
					}
				}
				heirs = append(heirs, NewSanyTokenNode(colon))
				readBounds()
			} else {
				readElements()
			}
		} else {
			p.throwParseException([][]SanyTokenKind{{SanyTokenIdentifier}}, "expected set expression")
		}
	}
	close := p.consumeParseToken(SanyTokenRbc, "expected }")
	if held != nil {
		heirs = append(heirs, held)
	}
	heirs = append(heirs, close)
	return NewSanyNode(SanySyntaxNodeKindByName[kind], heirs...)
}

// Java jj_2_42/43 scan IdentifierTuple or Identifier followed immediately by
// COMMA/COLON. Failed previews retain their furthest following-input position.
func (p *SanyParser) braceSimpleHead(separator SanyTokenKind) bool {
	offset := 0
	if p.tokenAt(offset).Kind == SanyTokenLab {
		offset++
		if p.tokenAt(offset).Kind == SanyTokenIdentifier {
			offset++
			for p.tokenAt(offset).Kind == SanyTokenComma {
				offset++
				if p.tokenAt(offset).Kind != SanyTokenIdentifier {
					p.rememberFailedLookahead(offset + 1)
					return false
				}
				offset++
			}
		}
		if p.tokenAt(offset).Kind != SanyTokenRab {
			p.rememberFailedLookahead(offset + 1)
			return false
		}
		offset++
	} else if p.tokenAt(offset).Kind == SanyTokenIdentifier {
		offset++
	} else {
		p.rememberFailedLookahead(1)
		return false
	}
	if p.tokenAt(offset).Kind == separator {
		return true
	}
	p.rememberFailedLookahead(offset + 1)
	return false
}

// Java jj_2_49(1) checks Expression's first token, including its junction
// indentation predicate. All operator tokens and proof-step lexemes can start it.
func (p *SanyParser) startsExpressionFirstAt(offset int) bool {
	kind := p.tokenAt(offset).Kind
	starts := kind >= SanyTokenOp57 && kind <= SanyTokenProofimplicitsteplexeme
	if !starts {
		switch kind {
		case SanyTokenCase, SanyTokenChoose, SanyTokenExists, SanyTokenForall,
			SanyTokenIf, SanyTokenLet, SanyTokenSF, SanyTokenTExists,
			SanyTokenTForall, SanyTokenWF, SanyTokenLbr, SanyTokenLsb,
			SanyTokenLbc, SanyTokenLab, SanyTokenNumberLiteral, SanyTokenStringLiteral:
			starts = true
		}
	}
	return starts && p.junctionListContext.isAboveCurrent(p.tokenAt(offset).Begin.Column)
}

func (p *SanyParser) startsExpressionLookahead() bool {
	if p.startsExpressionFirstAt(0) {
		return true
	}
	p.rememberFailedLookahead(1)
	return false
}

func (p *SanyParser) TupleOrAction() *SanySyntaxNode {
	p.beginProduction("Some << -- >> or >>_ Form")
	defer p.endProduction()
	heirs := []*SanySyntaxNode{p.consumeParseToken(SanyTokenLab, "expected <<")}
	if p.startsExpressionLookahead() {
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
	switch p.peek().Kind {
	case SanyTokenRab:
		heirs = append(heirs, p.consumeParseToken(SanyTokenRab, "expected >>"))
		return NewSanyNode(SanySyntaxNodeKindByName["N_Tuple"], heirs...)
	case SanyTokenArab:
		heirs = append(heirs, p.consumeParseToken(SanyTokenArab, "expected >>_"))
		heirs = append(heirs, p.ReducedExpression())
		return NewSanyNode(SanySyntaxNodeKindByName["N_ActionExpr"], heirs...)
	default:
		p.throwParseException([][]SanyTokenKind{{SanyTokenRab}, {SanyTokenArab}}, "expected >> or >>_")
		return nil
	}
}

func (p *SanyParser) SBracketCases() *SanySyntaxNode {
	p.beginProduction("Some [] Form")
	defer p.endProduction()
	heirs := []*SanySyntaxNode{p.consumeParseToken(SanyTokenLsb, "expected [")}
	kind := "N_FcnAppl"
	if p.matchFcnConst() {
		kind = "N_FcnConst"
		heirs = append(heirs, p.QuantBound())
		for p.match(SanyTokenComma) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()), p.QuantBound())
		}
		heirs = append(heirs, p.consumeParseToken(SanyTokenMapto, "expected |-> in function constructor"))
		heirs = append(heirs, p.ExpressionUntil(func(tok *SanyToken) bool {
			return tok.Kind == SanyTokenRsb || tok.Kind == SanyTokenEOF
		}))
	} else if (p.check(SanyTokenIdentifier) || isSanyFieldNameToken(p.peek().Kind)) && p.peekNext().Kind == SanyTokenMapto {
		// Java first tries Identifier MAPTO, then reclassifies keyword fields.
		p.reclassifyFieldName()
		kind = "N_RcdConstructor"
		heirs = append(heirs, p.FieldVal())
		for p.match(SanyTokenComma) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()), p.FieldVal())
		}
	} else if p.check(SanyTokenIdentifier) && p.peekNext().Kind == SanyTokenColon {
		kind = "N_SetOfRcds"
		heirs = append(heirs, p.FieldSet())
		for p.match(SanyTokenComma) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()), p.FieldSet())
		}
	} else {
		// Both failed field lookaheads scan a second token after Identifier.
		if p.check(SanyTokenIdentifier) {
			p.rememberFailedLookahead(2)
		}
		if !p.startsExpressionLookahead() {
			p.throwParseException([][]SanyTokenKind{{SanyTokenIdentifier}}, "expected bracket expression")
		}
		heirs = append(heirs, p.ExpressionUntil(func(tok *SanyToken) bool {
			return tok.Kind == SanyTokenComma || tok.Kind == SanyTokenRsb || tok.Kind == SanyTokenArsb || tok.Kind == SanyTokenArrow || tok.Kind == SanyTokenExcept || tok.Kind == SanyTokenEOF
		}))
		switch p.peek().Kind {
		case SanyTokenComma, SanyTokenRsb:
			for p.match(SanyTokenComma) {
				heirs = append(heirs, NewSanyTokenNode(p.previous()))
				heirs = append(heirs, p.ExpressionUntil(func(tok *SanyToken) bool {
					return tok.Kind == SanyTokenComma || tok.Kind == SanyTokenRsb || tok.Kind == SanyTokenEOF
				}))
			}
		case SanyTokenArrow:
			kind = "N_SetOfFcns"
			heirs = append(heirs, p.consumeParseToken(SanyTokenArrow, "expected -> in function set"))
			heirs = append(heirs, p.ExpressionUntil(func(tok *SanyToken) bool {
				return tok.Kind == SanyTokenRsb || tok.Kind == SanyTokenEOF
			}))
		case SanyTokenExcept:
			kind = "N_Except"
			heirs = append(heirs, p.consumeParseToken(SanyTokenExcept, "expected EXCEPT"), p.ExceptSpec())
			for p.match(SanyTokenComma) {
				heirs = append(heirs, NewSanyTokenNode(p.previous()), p.ExceptSpec())
			}
		case SanyTokenArsb:
			heirs = append(heirs, p.consumeParseToken(SanyTokenArsb, "expected ]_"), p.ReducedExpression())
			return NewSanyNode(SanySyntaxNodeKindByName["N_ActionExpr"], heirs...)
		default:
			p.throwParseException([][]SanyTokenKind{{SanyTokenComma}, {SanyTokenRsb}, {SanyTokenArrow}, {SanyTokenExcept}, {SanyTokenArsb}}, "expected bracket expression continuation")
		}
	}
	heirs = append(heirs, p.consumeParseToken(SanyTokenRsb, "expected ]"))
	return NewSanyNode(SanySyntaxNodeKindByName[kind], heirs...)
}

// matchFcnConst preserves Java's preview scan, including its acceptance of any
// balanced <<...>> before IN; IdentifierTuple validates the contents later.
func (p *SanyParser) matchFcnConst() bool {
	offset := 0
	switch p.tokenAt(offset).Kind {
	case SanyTokenLab:
		depth := 1
		for depth != 0 {
			offset++
			switch p.tokenAt(offset).Kind {
			case SanyTokenLab:
				depth++
			case SanyTokenRab:
				depth--
			case SanyTokenEOF:
				return false
			}
		}
		return p.tokenAt(offset+1).Kind == SanyTokenIN
	case SanyTokenIdentifier:
		offset++
		for p.tokenAt(offset).Kind == SanyTokenComma {
			offset++
			if p.tokenAt(offset).Kind != SanyTokenIdentifier {
				return false
			}
			offset++
		}
		return p.tokenAt(offset).Kind == SanyTokenIN
	default:
		return false
	}
}

func (p *SanyParser) FieldVal() *SanySyntaxNode {
	p.beginProduction("Field Value")
	defer p.endProduction()
	id := p.Identifier()
	mapto := p.consumeParseToken(SanyTokenMapto, "expected |-> in record field")
	expr := p.ExpressionUntil(func(tok *SanyToken) bool {
		return tok.Kind == SanyTokenComma || tok.Kind == SanyTokenRsb || tok.Kind == SanyTokenEOF
	})
	return NewSanyNode(SanySyntaxNodeKindByName["N_FieldVal"], id, mapto, expr)
}

func (p *SanyParser) FieldSet() *SanySyntaxNode {
	p.beginProduction("Field Set")
	defer p.endProduction()
	id := p.Identifier()
	colon := p.consumeParseToken(SanyTokenColon, "expected : in record field set")
	expr := p.ExpressionUntil(func(tok *SanyToken) bool {
		return tok.Kind == SanyTokenComma || tok.Kind == SanyTokenRsb || tok.Kind == SanyTokenEOF
	})
	return NewSanyNode(SanySyntaxNodeKindByName["N_FieldSet"], id, colon, expr)
}

func (p *SanyParser) ExceptSpec() *SanySyntaxNode {
	p.beginProduction("Except Spec")
	defer p.endProduction()
	heirs := []*SanySyntaxNode{p.consumeParseToken(SanyTokenBang, "expected ! in EXCEPT spec")}
	for {
		heirs = append(heirs, p.ExceptComponent())
		p.expecting = "= or ,"
		if !p.check(SanyTokenDot) && !p.check(SanyTokenLsb) {
			break
		}
	}
	equals := p.consumeParseToken(SanyTokenEquals, "expected = in EXCEPT spec")
	equals.Kind = SanySyntaxNodeKindByName["T_EQUAL"]
	heirs = append(heirs, equals)
	heirs = append(heirs, p.ExpressionUntil(func(tok *SanyToken) bool {
		return tok.Kind == SanyTokenComma || tok.Kind == SanyTokenRsb || tok.Kind == SanyTokenEOF
	}))
	return NewSanyNode(SanySyntaxNodeKindByName["N_ExceptSpec"], heirs...)
}

func (p *SanyParser) ExceptComponent() *SanySyntaxNode {
	p.beginProduction("Except Component")
	defer p.endProduction()
	var heirs []*SanySyntaxNode
	switch p.peek().Kind {
	case SanyTokenDot:
		heirs = append(heirs, p.consumeParseToken(SanyTokenDot, "expected ."))
		p.reclassifyFieldName()
		identifier := p.Identifier()
		if identifier.Image == "@" {
			diagnostic := errorAt(identifier.Range.Begin, "E1300", "@ used in !.@")
			diagnostic.SANYParseMessage = diagnostic.Message
			p.diags = append(p.diags, diagnostic)
		}
		heirs = append(heirs, identifier)
	case SanyTokenLsb:
		heirs = append(heirs, p.consumeParseToken(SanyTokenLsb, "expected ["))
		heirs = append(heirs, p.ExpressionUntil(func(tok *SanyToken) bool {
			return tok.Kind == SanyTokenComma || tok.Kind == SanyTokenRsb || tok.Kind == SanyTokenEOF
		}))
		for p.match(SanyTokenComma) {
			heirs = append(heirs, NewSanyTokenNode(p.previous()))
			heirs = append(heirs, p.ExpressionUntil(func(tok *SanyToken) bool {
				return tok.Kind == SanyTokenComma || tok.Kind == SanyTokenRsb || tok.Kind == SanyTokenEOF
			}))
		}
		heirs = append(heirs, p.consumeParseToken(SanyTokenRsb, "expected ] in EXCEPT component"))
	default:
		p.throwParseException([][]SanyTokenKind{{SanyTokenDot}, {SanyTokenLsb}}, "expected EXCEPT component")
	}
	return NewSanyNode(SanySyntaxNodeKindByName["N_ExceptComponent"], heirs...)
}

func (p *SanyParser) ReducedExpression() *SanySyntaxNode {
	p.beginProduction("restricted form of expression")
	defer p.endProduction()
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
		p.throwParseException([][]SanyTokenKind{{SanyTokenIdentifier}, {SanyTokenLbr}, {SanyTokenLbc}, {SanyTokenLsb}, {SanyTokenLab}}, "expected restricted expression after action subscript")
		return nil
	}
}

func (p *SanyParser) startsNoOpExtension() bool {
	if !p.startsNoOpExtensionBaseAt(0) {
		return false
	}
	offset := 1
	if p.check(SanyTokenIdentifier) && p.tokenAt(offset).Kind == SanyTokenLbr {
		end := p.findMatchingBracketOffset(offset)
		if end < 0 {
			return false
		}
		offset = end + 1
	}
	return p.tokenAt(offset).Kind == SanyTokenBang
}

func (p *SanyParser) startsNoOpExtensionBaseAt(offset int) bool {
	switch p.tokenAt(offset).Kind {
	case SanyTokenIdentifier, SanyTokenProofsteplexeme, SanyTokenProofimplicitsteplexeme:
		return true
	default:
		return false
	}
}

func (p *SanyParser) startsStructOp() bool {
	return p.startsStructOpAt(0)
}

func (p *SanyParser) startsStructOpAt(offset int) bool {
	switch p.tokenAt(offset).Kind {
	case SanyTokenLab, SanyTokenRab, SanyTokenColon, SanyTokenNumberLiteral:
		return true
	case SanyTokenIdentifier:
		return p.tokenAt(offset).Image == "@"
	default:
		return false
	}
}

func (p *SanyParser) startsBangOperatorSelector() bool {
	return p.isOperatorTokenAt(0)
}

func (p *SanyParser) isOperatorTokenAt(offset int) bool {
	_, ok := GetSanyOperator(p.tokenAt(offset).Image)
	return ok
}

func (p *SanyParser) startsLabelAt(offset int) bool {
	if p.tokenAt(offset).Kind != SanyTokenIdentifier {
		return false
	}
	if p.tokenAt(offset+1).Kind == SanyTokenColoncolon {
		return true
	}
	if p.tokenAt(offset+1).Kind == SanyTokenLbr {
		end := p.findMatchingBracketOffset(offset + 1)
		if end >= 0 && p.tokenAt(end+1).Kind == SanyTokenColoncolon {
			return true
		}
	}
	return p.startsPrefixedLabelAt(offset)
}

func (p *SanyParser) startsPrefixedLabelAt(offset int) bool {
	at := offset
	if p.tokenAt(at).Kind != SanyTokenIdentifier {
		return false
	}
	at++
	if p.tokenAt(at).Kind == SanyTokenLbr {
		end := p.findMatchingBracketOffset(at)
		if end < 0 {
			return false
		}
		at = end + 1
	}
	sawBang := false
	for p.tokenAt(at).Kind == SanyTokenBang {
		sawBang = true
		at++
		switch {
		case p.tokenAt(at).Kind == SanyTokenLbr:
			end := p.findMatchingBracketOffset(at)
			if end < 0 {
				return false
			}
			at = end + 1
		case p.startsStructOpAt(at), p.isOperatorTokenAt(at), p.tokenAt(at).Kind == SanyTokenIdentifier:
			at++
			if p.tokenAt(at).Kind == SanyTokenLbr {
				end := p.findMatchingBracketOffset(at)
				if end < 0 {
					return false
				}
				at = end + 1
			}
		default:
			return false
		}
	}
	return sawBang && p.tokenAt(at).Kind == SanyTokenColoncolon
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
	p.beginProduction("Optional Arguments")
	defer p.endProduction()
	var heirs []*SanySyntaxNode
	heirs = append(heirs, p.consumeParseToken(SanyTokenLbr, "expected ( in operator arguments"))
	heirs = append(heirs, p.OpOrExpr(func(tok *SanyToken) bool {
		return tok.Kind == SanyTokenComma || tok.Kind == SanyTokenRbr || tok.Kind == SanyTokenEOF
	}))
	for p.match(SanyTokenComma) {
		heirs = append(heirs, NewSanyTokenNode(p.previous()))
		heirs = append(heirs, p.OpOrExpr(func(tok *SanyToken) bool {
			return tok.Kind == SanyTokenComma || tok.Kind == SanyTokenRbr || tok.Kind == SanyTokenEOF
		}))
	}
	heirs = append(heirs, p.consumeParseToken(SanyTokenRbr, "expected ) in operator arguments"))
	return NewSanyNode(SanySyntaxNodeKindByName["N_OpArgs"], heirs...)
}

// Source optional OpArgs uses two-token lookahead: an opening parenthesis
// followed by the first token of the mandatory OpOrExpr. A failed lookahead
// leaves the parenthesis for the surrounding production.
func (p *SanyParser) startsOpArgs() bool {
	if !p.check(SanyTokenLbr) {
		return false
	}
	if p.startsOpOrExprAt(1) {
		return true
	}
	p.rememberFailedLookahead(2)
	return false
}

func (p *SanyParser) startsOpOrExprAt(offset int) bool {
	kind := p.tokenAt(offset).Kind
	if kind >= SanyTokenOp57 && kind <= SanyTokenOp119 {
		return true
	}
	switch kind {
	case SanyTokenCase, SanyTokenChoose, SanyTokenExists, SanyTokenForall,
		SanyTokenIf, SanyTokenLet, SanyTokenSF, SanyTokenTExists, SanyTokenTForall,
		SanyTokenLambda, SanyTokenWF, SanyTokenLbr, SanyTokenLsb, SanyTokenLbc,
		SanyTokenLab, SanyTokenNumberLiteral, SanyTokenStringLiteral,
		SanyTokenIdentifier, SanyTokenProofsteplexeme, SanyTokenProofimplicitsteplexeme:
		return true
	default:
		return false
	}
}

// PrimitiveExp allows the full BangExt selector production; the restricted
// NoOpExtension used by action subscripts is a separate source production.
func (p *SanyParser) primitiveSelectorExpr(selector *SanySyntaxNode) *SanySyntaxNode {
	var prefix []*SanySyntaxNode
	args := p.OptionalSelectorOpArgs(selector)
	for p.check(SanyTokenBang) {
		bang, next, nextArgs := p.BangExtension()
		if args != nil {
			prefix = append(prefix, NewSanyNode(SanySyntaxNodeKindByName["N_IdPrefixElement"], selector, args, bang))
		} else {
			prefix = append(prefix, NewSanyNode(SanySyntaxNodeKindByName["N_IdPrefixElement"], selector, bang))
		}
		selector, args = next, nextArgs
	}
	genID := NewSanyNode(SanySyntaxNodeKindByName["N_GeneralId"],
		NewSanyNode(SanySyntaxNodeKindByName["N_IdPrefix"], prefix...), selector)
	if args != nil {
		return NewSanyNode(SanySyntaxNodeKindByName["N_OpApplication"], genID, args)
	}
	return genID
}

func (p *SanyParser) BangExtension() (bang, selector, args *SanySyntaxNode) {
	p.beginProduction("Bang Extension")
	defer p.endProduction()
	bang = p.consumeParseToken(SanyTokenBang, "expected ! in selector")
	selector = p.BangSelector()
	args = p.OptionalSelectorOpArgs(selector)
	return bang, selector, args
}

func (p *SanyParser) NoOpExtension() *SanySyntaxNode {
	var prefix []*SanySyntaxNode
	selector := p.consumeParseToken(SanyTokenIdentifier, "expected identifier in restricted expression")
	var args *SanySyntaxNode
	if p.startsOpArgs() {
		args = p.OpArgs()
	}
	for p.match(SanyTokenBang) {
		bang := NewSanyTokenNode(p.previous())
		if args != nil {
			prefix = append(prefix, NewSanyNode(SanySyntaxNodeKindByName["N_IdPrefixElement"], selector, args, bang))
		} else {
			prefix = append(prefix, NewSanyNode(SanySyntaxNodeKindByName["N_IdPrefixElement"], selector, bang))
		}
		selector = p.consumeParseToken(SanyTokenIdentifier, "expected identifier in restricted expression")
		args = nil
		if p.startsOpArgs() {
			args = p.OpArgs()
		}
	}
	// The final argument list is detached for FairnessExpr. Earlier prefix
	// argument lists remain attached to their IdPrefixElement nodes.
	p.fairnessHook = args
	return NewSanyNode(SanySyntaxNodeKindByName["N_GeneralId"],
		NewSanyNode(SanySyntaxNodeKindByName["N_IdPrefix"], prefix...), selector)
}

func (p *SanyParser) NoOpExtensionBase() *SanySyntaxNode {
	switch p.peek().Kind {
	case SanyTokenIdentifier:
		return p.Identifier()
	case SanyTokenProofsteplexeme, SanyTokenProofimplicitsteplexeme:
		return NewSanyTokenNode(p.advance())
	default:
		return p.Identifier()
	}
}

func (p *SanyParser) OptionalSelectorOpArgs(selector *SanySyntaxNode) *SanySyntaxNode {
	if selector == nil || !p.selectorAllowsOpArgs(selector) || !p.startsOpArgs() {
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
	if kindName == "N_InfixOp" {
		return NewSanyNode(SanySyntaxNodeKindByName[kindName], NewSanyTokenNode(p.infixOpToken()))
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

// TLAplusParser.reduceString decodes escapes while retaining the quote marks.
// Generator's StringNode strips those marks later, when creating its value.
func sanyReduceString(image string) string {
	var out strings.Builder
	padding := 0
	for i := 0; i < len(image); i++ {
		if image[i] != '\\' {
			out.WriteByte(image[i])
			continue
		}
		i++
		switch image[i] {
		case '\\', '"':
			out.WriteByte(image[i])
		case 'n':
			out.WriteByte('\n')
		case 'r':
			out.WriteByte('\r')
		case 'f':
			out.WriteByte('\f')
		case 't':
			out.WriteByte('\t')
		default:
			// Source increments its output length even for an unknown escape.
			padding++
		}
	}
	for ; padding > 0; padding-- {
		out.WriteByte(0)
	}
	return out.String()
}

func reduceTLAString(image string) string {
	value := sanyReduceString(image)
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		return value[1 : len(value)-1]
	}
	return value
}

func (p *SanyParser) genericOperatorNode(tok *SanyToken, op SanyOperatorInfo) *SanySyntaxNode {
	kindName := "N_GenInfixOp"
	leafKindName := "N_InfixOp"
	if op.IsPrefix() {
		kindName = "N_GenPrefixOp"
		leafKindName = "N_PrefixOp"
	} else if op.IsPostfix() {
		kindName = "N_GenPostfixOp"
		leafKindName = "N_PostfixOp"
	}
	return NewSanyNode(
		SanySyntaxNodeKindByName[kindName],
		NewSanyNode(SanySyntaxNodeKindByName["N_IdPrefix"]),
		sanyOperatorTokenNode(leafKindName, tok),
	)
}

func (p *SanyParser) genericOperatorReferenceNode(tok *SanyToken, op SanyOperatorInfo) *SanySyntaxNode {
	if !op.IsPrefix() {
		return p.genericOperatorNode(tok, op)
	}
	return NewSanyNode(
		SanySyntaxNodeKindByName["N_GenNonExpPrefixOp"],
		NewSanyNode(SanySyntaxNodeKindByName["N_IdPrefix"]),
		sanyOperatorTokenNode("N_NonExpPrefixOp", tok),
	)
}

// isSanyFieldNameToken mirrors Java tla+.jj isFieldNameToken. The generated
// token ranges retain Java's numbering and include its keyword synonyms.
func isSanyFieldNameToken(kind SanyTokenKind) bool {
	return (kind >= SanyTokenAction && kind <= SanyTokenExcept) ||
		kind == SanyTokenExtends ||
		(kind >= SanyTokenIf && kind <= SanyTokenSF) ||
		kind == SanyTokenState ||
		(kind >= SanyTokenThen && kind <= SanyTokenWith) ||
		kind == SanyTokenUs ||
		(kind >= SanyTokenOp112 && kind <= SanyTokenOp116)
}

func (p *SanyParser) reclassifyFieldName() {
	if isSanyFieldNameToken(p.peek().Kind) {
		p.peek().Kind = SanyTokenIdentifier
	}
}

func (p *SanyParser) Identifier() *SanySyntaxNode {
	return p.consumeParseToken(SanyTokenIdentifier, "expected identifier")
}

func (p *SanyParser) startsBodyItem() bool {
	return p.startsBodyItemAt(0)
}

// Body's one-token lookahead admits the unit's first token, with the
// source semantic exclusion for USE ONLY. Its later decisions use two tokens.
func (p *SanyParser) startsBodyItemAt(offset int) bool {
	switch p.tokenAt(offset).Kind {
	case SanyTokenBm0, SanyTokenBm1, SanyTokenBm2, SanyTokenSeparator,
		SanyTokenVariable, SanyTokenConstant, SanyTokenRecursive, SanyTokenAssume,
		SanyTokenAssumption, SanyTokenTheorem, SanyTokenProposition,
		SanyTokenLocal, SanyTokenInstance, SanyTokenHide, SanyTokenDefbreak:
		return true
	case SanyTokenUse:
		return p.tokenAt(offset+1).Kind != SanyTokenOnly
	default:
		return false
	}
}

func (p *SanyParser) startsOperatorOrFunctionDefinition() bool {
	return p.startsOperatorOrFunctionDefinitionAt(0)
}

func (p *SanyParser) startsOperatorOrFunctionDefinitionAt(offset int) bool {
	if p.tokenAt(offset).Kind == SanyTokenLocal {
		// The two-token budget ends at DEFBREAK, before validating the head.
		return p.tokenAt(offset+1).Kind == SanyTokenDefbreak
	}
	return p.tokenAt(offset).Kind == SanyTokenDefbreak &&
		(p.tokenAt(offset+1).Kind == SanyTokenIdentifier || p.startsDefinitionPrefixAt(offset+1))
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

func (p *SanyParser) parseErrorMessage(expected string, tok *SanyToken) string {
	image := "<EOF>"
	pos := Position{}
	if tok != nil {
		pos = tok.Begin
		if tok.Image != "" {
			image = tok.Image
		} else if javaImage := tok.Kind.JavaImage(); javaImage != "" {
			image = strings.Trim(javaImage, "\"")
		}
	}
	return "Was expecting \"" + expected + "\"\nEncountered \"" + image + "\" at line " + strconv.Itoa(pos.Line) + ", column " + strconv.Itoa(pos.Column) + "."
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

func (p *SanyParser) peek() *SanyToken { return p.tokenAt(0) }

func (p *SanyParser) peekNext() *SanyToken {
	return p.tokenAt(1)
}

func (p *SanyParser) tokenAt(offset int) *SanyToken {
	idx := p.at + offset
	for p.tokenManager != nil && idx >= len(p.tokens) {
		if len(p.tokens) != 0 && p.tokens[len(p.tokens)-1].Kind == SanyTokenEOF {
			return p.tokens[len(p.tokens)-1]
		}
		token := p.tokenManager.NextToken()
		if len(p.tokens) != 0 {
			p.tokens[len(p.tokens)-1].Next = token
		}
		p.tokens = append(p.tokens, token)
	}
	if idx >= len(p.tokens) {
		return &SanyToken{Kind: SanyTokenEOF}
	}
	return p.tokens[idx]
}

func (p *SanyParser) add(pos Position, code, msg string) {
	p.diags = append(p.diags, errorAt(pos, code, "%s", msg))
}

// InfixOp alone, unlike the source prefix and postfix productions, owns a frame.
func (p *SanyParser) infixOpToken() *SanyToken {
	p.beginProduction("Infix Op")
	defer p.endProduction()
	return p.advance()
}

// Source SyntaxTreeNode(module, kind, token) is a leaf with the token's image.
func sanyOperatorTokenNode(kind string, token *SanyToken) *SanySyntaxNode {
	node := NewSanyTokenNode(token)
	node.Kind = SanySyntaxNodeKindByName[kind]
	return node
}
