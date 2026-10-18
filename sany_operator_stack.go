package tlago

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/glycerine/tlago/tlc"
)

type SanyOperatorFixity int

const (
	SanyOperatorNoFix SanyOperatorFixity = iota
	SanyOperatorPrefix
	SanyOperatorPostfix
	SanyOperatorInfix
	SanyOperatorNfix
)

type SanyOperatorAssociativity int

const (
	SanyAssociativityNone SanyOperatorAssociativity = iota
	SanyAssociativityLeft
	SanyAssociativityRight
	SanyAssociativityChain
)

type SanyOperatorInfo struct {
	Symbol         string
	LowPrecedence  int
	HighPrecedence int
	Associativity  SanyOperatorAssociativity
	Fix            SanyOperatorFixity
	identifier     *tlc.UniqueString
	identifierSet  bool
}

func NewSanyOperatorInfo(symbol string, low, high int, associativity SanyOperatorAssociativity, fix SanyOperatorFixity) *SanyOperatorInfo {
	return NewSanyOperatorInfoFromUniqueString(tlc.UniqueStringOf(symbol), low, high, associativity, fix)
}

func NewSanyOperatorInfoFromUniqueString(identifier *tlc.UniqueString, low, high int, associativity SanyOperatorAssociativity, fix SanyOperatorFixity) *SanyOperatorInfo {
	symbol := ""
	if identifier != nil {
		symbol = identifier.String()
	}
	return &SanyOperatorInfo{
		Symbol:         symbol,
		LowPrecedence:  low,
		HighPrecedence: high,
		Associativity:  associativity,
		Fix:            fix,
		identifier:     identifier,
		identifierSet:  true,
	}
}

var sanyVoidOperator = &SanyOperatorInfo{
	Symbol:         "$$_void",
	LowPrecedence:  0,
	HighPrecedence: 0,
	Associativity:  SanyAssociativityNone,
	Fix:            SanyOperatorInfix,
}

func SanyVoidOperator() *SanyOperatorInfo {
	return sanyVoidOperator
}

func GetSanyOperator(identifier *tlc.UniqueString) *SanyOperatorInfo {
	if identifier == nil {
		panic(tlc.NewNullPointerException())
	}
	return SanyOperatorsBySymbol[identifier.String()]
}

func LookupSanyOperator(symbol string) (SanyOperatorInfo, bool) {
	op, ok := SanyOperatorsBySymbol[symbol]
	if !ok {
		return SanyOperatorInfo{}, false
	}
	return *op, true

}

func SanyOperatorExists(identifier *tlc.UniqueString) bool {
	if identifier == nil {
		panic(tlc.NewNullPointerException())
	}
	_, ok := SanyOperatorsBySymbol[identifier.String()]
	return ok
}

func GetSanyMixfixOperator(op *SanyOperatorInfo) *SanyOperatorInfo {
	if op.IsPrefix() {
		return op
	}
	return GetSanyOperator(op.GetIdentifier().ConcatString("."))
}

func lookupSanyMixfixOperator(op SanyOperatorInfo) (SanyOperatorInfo, bool) {
	mixfix := GetSanyMixfixOperator(&op)
	if mixfix == nil {
		return SanyOperatorInfo{}, false
	}
	return *mixfix, true
}

func ResolveSanyOperatorSynonym(identifier *tlc.UniqueString) *tlc.UniqueString {
	operator := GetSanyOperator(identifier)
	if operator == nil {
		return identifier
	}
	return operator.GetIdentifier()
}

func ResolveSanyOperatorSynonymString(symbol string) string {
	return ResolveSanyOperatorSynonym(tlc.UniqueStringOf(symbol)).String()
}

func SanyPrintOperatorTable() {
	SanyPrintOperatorTableTo(os.Stdout)
}

func SanyPrintOperatorTableTo(out io.Writer) {
	if out == nil {
		panic(tlc.NewNullPointerException())
	}
	_, _ = fmt.Fprintln(out, "printing Operators table")
	for _, symbol := range sanyOperatorPrintOrder {
		_, _ = fmt.Fprintln(out, "-> "+symbol)
	}
}

func (op *SanyOperatorInfo) require() {
	if op == nil {
		panic(tlc.NewNullPointerException())
	}
}

func (op *SanyOperatorInfo) String() string {
	op.require()
	switch op.Fix {
	case SanyOperatorNoFix:
		return sanyOperatorIdentifierString(op.GetIdentifier()) + ", nofix"
	case SanyOperatorPrefix:
		return sanyOperatorIdentifierString(op.GetIdentifier()) + ", prefix"
	case SanyOperatorPostfix:
		return sanyOperatorIdentifierString(op.GetIdentifier()) + ", postfix"
	case SanyOperatorInfix:
		return sanyOperatorIdentifierString(op.GetIdentifier()) + ", infix"
	case SanyOperatorNfix:
		return sanyOperatorIdentifierString(op.GetIdentifier()) + ", nfix"
	default:
		return "null"
	}
}

func (op *SanyOperatorInfo) IsPrefix() bool {
	op.require()
	return op.Fix == SanyOperatorPrefix
}

func (op *SanyOperatorInfo) IsInfix() bool {
	op.require()
	return op.Fix == SanyOperatorInfix || op.Fix == SanyOperatorNfix
}

func (op *SanyOperatorInfo) IsPostfix() bool {
	op.require()
	return op.Fix == SanyOperatorPostfix
}

func (op *SanyOperatorInfo) IsNfix() bool {
	op.require()
	return op.Fix == SanyOperatorNfix
}

func (op *SanyOperatorInfo) IsPrefixDecl() bool {
	op.require()
	return op.Fix == SanyOperatorPrefix && strings.HasSuffix(sanyOperatorIdentifierString(op.GetIdentifier()), ".")
}

func (op *SanyOperatorInfo) AssocLeft() bool {
	op.require()
	return op.Associativity == SanyAssociativityLeft
}

func (op *SanyOperatorInfo) AssocRight() bool {
	op.require()
	return op.Associativity == SanyAssociativityRight
}

func (op *SanyOperatorInfo) GetIdentifier() *tlc.UniqueString {
	op.require()
	if op.identifierSet {
		return op.identifier
	}
	return tlc.UniqueStringOf(op.Symbol)
}

func sanyOperatorIdentifierString(identifier *tlc.UniqueString) string {
	if identifier == nil {
		panic(tlc.NewNullPointerException())
	}
	return identifier.String()
}

func SanyOperatorSucc(left, right SanyOperatorInfo) bool {
	return left.LowPrecedence > right.HighPrecedence
}

func SanyOperatorPrec(left, right SanyOperatorInfo) bool {
	return left.HighPrecedence < right.LowPrecedence
}

func SanyOperatorSamePrecedence(left, right SanyOperatorInfo) bool {
	return left.HighPrecedence == right.HighPrecedence && left.LowPrecedence == right.LowPrecedence
}

type sanyOperatorStackElement struct {
	Node     *SanySyntaxNode
	Operator *SanyOperatorInfo
}

func (e *sanyOperatorStackElement) require() {
	if e == nil {
		panic(tlc.NewNullPointerException())
	}
}

func (e *sanyOperatorStackElement) GetOperator() *SanyOperatorInfo {
	e.require()
	return e.Operator
}

func (e *sanyOperatorStackElement) GetNode() *SanySyntaxNode {
	e.require()
	return e.Node
}

func (e *sanyOperatorStackElement) IsOperator() bool {
	e.require()
	return e.Operator != nil
}

type SanyOperatorStack struct {
	stackOfStacks [][]sanyOperatorStackElement
	moduleName    string
	parseErrors   *SanyParseErrors
}

func NewSanyOperatorStack(parseErrors ...*SanyParseErrors) *SanyOperatorStack {
	errors := NewSanyParseErrors()
	if len(parseErrors) != 0 {
		errors = parseErrors[0]
	}
	return &SanyOperatorStack{parseErrors: errors}
}

func (s *SanyOperatorStack) require() {
	if s == nil {
		panic(tlc.NewNullPointerException())
	}
}

func (s *SanyOperatorStack) NewStack() {
	s.require()
	s.stackOfStacks = append(s.stackOfStacks, nil)
}

func (s *SanyOperatorStack) PopStack() {
	s.require()
	if len(s.stackOfStacks) == 0 {
		panic(tlc.NewArrayIndexOutOfBoundsException(-1, 0))
	}
	s.stackOfStacks = s.stackOfStacks[:len(s.stackOfStacks)-1]
}

func (s *SanyOperatorStack) Push(node *SanySyntaxNode, op *SanyOperatorInfo) {
	s.PushOnStack(node, op)
}

func (s *SanyOperatorStack) PushOnStack(node *SanySyntaxNode, op *SanyOperatorInfo) {
	s.require()
	if len(s.stackOfStacks) == 0 {
		panic(tlc.NewNullPointerException())
	}
	top := len(s.stackOfStacks) - 1
	s.stackOfStacks[top] = append(s.stackOfStacks[top], sanyOperatorStackElement{Node: node, Operator: op})
}

func (s *SanyOperatorStack) CurrentSize() int {
	return s.Size()
}

func (s *SanyOperatorStack) Empty() bool {
	return s.Size() == 0
}

func (s *SanyOperatorStack) Size() int {
	s.require()
	if len(s.stackOfStacks) == 0 {
		panic(tlc.NewNullPointerException())
	}
	return len(s.stackOfStacks[len(s.stackOfStacks)-1])
}

func (s *SanyOperatorStack) TopNode() *SanySyntaxNode {
	element := s.TopOfStack()
	if element == nil {
		return nil
	}
	return element.Node
}

// Java OperatorStack.popCurrentTop removes the label operand before parsing
// the labeled expression, leaving the enclosing operator available to its check.
func (s *SanyOperatorStack) PopCurrentTop() {
	s.require()
	if len(s.stackOfStacks) == 0 {
		panic(tlc.NewNullPointerException())
	}
	top := len(s.stackOfStacks) - 1
	if len(s.stackOfStacks[top]) == 0 {
		panic(tlc.NewArrayIndexOutOfBoundsException(-1, 0))
	}
	s.stackOfStacks[top] = s.stackOfStacks[top][:len(s.stackOfStacks[top])-1]
}

func (s *SanyOperatorStack) TopOperator() *SanyOperatorInfo {
	element := s.TopOfStack()
	if element == nil {
		return nil
	}
	return element.Operator
}

func (s *SanyOperatorStack) FinalReduce() (*SanySyntaxNode, error) {
	s.Push(nil, SanyVoidOperator())
	if err := s.ReduceStack(); err != nil {
		return nil, err
	}
	if !s.IsWellReduced() {
		var message strings.Builder
		message.WriteString("Couldn't properly parse expression")
		for _, element := range s.currentTop()[:s.CurrentSize()-1] {
			message.WriteString("-- incomplete expression at ")
			message.WriteString(s.location(element.Node))
			message.WriteString(".\n")
		}
		s.parseErrors.push(NewSanyParseError(message.String(), "-- --"))
		return nil, nil
	}
	return s.stackOfStacks[len(s.stackOfStacks)-1][0].Node, nil
}

func (s *SanyOperatorStack) IsWellReduced() bool {
	return s.Size() == 2
}

func (s *SanyOperatorStack) BottomOfStack() *SanySyntaxNode {
	s.require()
	if len(s.stackOfStacks) == 0 {
		panic(tlc.NewNullPointerException())
	}
	top := s.currentTop()
	if len(top) == 0 {
		panic(tlc.NewArrayIndexOutOfBoundsException(0, 0))
	}
	return top[0].Node
}

func (s *SanyOperatorStack) TopOfStack() *sanyOperatorStackElement {
	s.require()
	if len(s.stackOfStacks) == 0 || len(s.currentTop()) == 0 {
		return nil
	}
	top := s.currentTop()
	return &top[len(top)-1]
}

func (s *SanyOperatorStack) ReduceStack() error {
	s.require()
	if len(s.stackOfStacks) == 0 {
		panic(tlc.NewNullPointerException())
	}
	for {
		top := s.currentTop()
		if len(top) == 0 {
			return nil
		}
		n := len(top) - 1
		tm0 := top[n]
		if !tm0.isOperator() {
			return nil
		}
		oR := *tm0.Operator
		sizeBefore := len(top)
		var err error
		switch {
		case oR.IsPostfix():
			err = s.reduceWithRightPostfix(n, oR)
		case oR.IsPrefix():
			err = s.reduceWithRightPrefix(n, oR)
		default:
			err = s.reduceWithRightInfix(n, oR)
		}
		if err != nil {
			return err
		}
		if len(s.currentTop()) == sizeBefore {
			return nil
		}
	}
}

// sanyOperatorStackFailure retains the native API's text separately from the
// actual source ParseException message used by the parser.
type sanyOperatorStackFailure struct{ sourceMessage, nativeMessage string }

func (failure *sanyOperatorStackFailure) Error() string { return failure.nativeMessage }
func (s *SanyOperatorStack) failure(sourceMessage, nativeMessage string) error {
	return &sanyOperatorStackFailure{sourceMessage, nativeMessage}
}
func (s *SanyOperatorStack) location(node *SanySyntaxNode) string {
	module := s.moduleName
	if node.FileName != "" {
		module = node.FileName
	}
	if module == "" {
		module = "null"
	}
	return fmt.Sprintf("line %d, col %d to line %d, col %d of module %s", node.Range.Begin.Line, node.Range.Begin.Column, node.Range.End.Line, node.Range.End.Column, module)
}

func (s *SanyOperatorStack) reduceWithRightPostfix(n int, oR SanyOperatorInfo) error {
	top := s.currentTop()
	tm0 := top[n]
	if n == 0 {
		return s.failure(fmt.Sprintf("\n  Encountered postfix op %s in block %s on empty stack", oR.Symbol, s.location(tm0.Node)), fmt.Sprintf("postfix operator %s on empty stack", oR.Symbol))
	}
	tm1 := top[n-1]
	if tm1.isOperator() {
		oL := *tm1.Operator
		if oL.IsInfix() || oL.IsPrefix() {
			return s.failure(fmt.Sprintf("\n  Encountered postfix op %s in block %s following prefix or infix op %s.", oR.Symbol, s.location(tm0.Node), oL.Symbol), fmt.Sprintf("postfix operator %s follows %s operator %s", oR.Symbol, oL.fixityName(), oL.Symbol))
		}
		s.reducePostfix()
		return nil
	}
	if n <= 1 {
		return nil
	}
	tm2 := top[n-2]
	if !tm2.isOperator() {
		return s.failure(fmt.Sprintf("Expression at location %s and expression at location %s follow each other without any intervening operator.", s.location(tm2.Node), s.location(tm1.Node)), "adjacent expressions without intervening operator")
	}
	oL := *tm2.Operator
	if SanyOperatorSucc(oL, oR) {
		if oL.IsInfix() {
			s.reduceInfix(oL)
		} else {
			s.reducePrefix()
		}
		return nil
	}
	if !SanyOperatorPrec(oL, oR) {
		// Source intentionally reports tm0's location for this left operator.
		return s.failure(fmt.Sprintf("Precedence conflict between ops %s in block %s and %s.", oL.Symbol, s.location(tm0.Node), oR.Symbol), fmt.Sprintf("precedence conflict between operators %s and %s", oL.Symbol, oR.Symbol))
	}
	return nil
}

func (s *SanyOperatorStack) reduceWithRightPrefix(n int, oR SanyOperatorInfo) error {
	if n == 0 {
		return nil
	}
	top := s.currentTop()
	tm0, tm1 := top[n], top[n-1]
	if tm1.isOperator() {
		oL := *tm1.Operator
		if oL.IsPostfix() {
			return s.failure(fmt.Sprintf("\n  Encountered prefix op %s in block %s following postfix op %s.", oR.Symbol, s.location(tm0.Node), oL.Symbol), fmt.Sprintf("prefix operator %s follows postfix operator %s", oR.Symbol, oL.Symbol))
		}
		return nil
	}
	return s.failure(fmt.Sprintf("\n  Encountered prefix op %s in block %s following an expression.", oR.Symbol, s.location(tm0.Node)), fmt.Sprintf("prefix operator %s follows an expression", oR.Symbol))
}

func (s *SanyOperatorStack) reduceWithRightInfix(n int, oR SanyOperatorInfo) error {
	top := s.currentTop()
	tm0 := top[n]
	if n == 0 {
		if _, ok := lookupSanyMixfixOperator(oR); !ok {
			return s.failure(fmt.Sprintf("\n  Encountered infix op %s in block %s on empty stack.", oR.Symbol, s.location(tm0.Node)), fmt.Sprintf("infix operator %s on empty stack", oR.Symbol))
		}
		return nil
	}
	tm1 := top[n-1]
	if tm1.isOperator() {
		oL := *tm1.Operator
		if oL.IsInfix() || oL.IsPrefix() {
			mixR, ok := lookupSanyMixfixOperator(oR)
			if !ok {
				message := fmt.Sprintf("\n  Encountered infix op %s in block %s following prefix or infix op %s.", oR.Symbol, s.location(tm1.Node), oL.Symbol)
				if oR.Symbol == SanyVoidOperator().Symbol {
					message = fmt.Sprintf("\n  Missing expression in block %s following prefix or infix op %s.", s.location(tm1.Node), oL.Symbol)
				}
				return s.failure(message, fmt.Sprintf("infix operator %s follows %s operator %s", oR.Symbol, oL.fixityName(), oL.Symbol))
			}
			if SanyOperatorSucc(oL, mixR) || (oL.Symbol == mixR.Symbol && oL.AssocLeft()) {
				return s.failure(fmt.Sprintf("\n  Precedence conflict between ops %s in block %s and %s.", oL.Symbol, s.location(tm1.Node), mixR.Symbol), fmt.Sprintf("precedence conflict between operators %s and %s", oL.Symbol, mixR.Symbol))
			}
			return nil
		}
		s.reducePostfix()
		return nil
	}
	if n <= 1 {
		return nil
	}
	tm2 := top[n-2]
	if !tm2.isOperator() {
		return s.failure(fmt.Sprintf("Expression at location %s and expression at location %s follow each other without any intervening operator.", s.location(tm2.Node), s.location(tm1.Node)), "adjacent expressions without intervening operator")
	}
	oL := *tm2.Operator
	if mixL, ok := lookupSanyMixfixOperator(oL); ok && (n == 2 || top[n-3].isOperator()) {
		oL = mixL
	}
	if SanyOperatorSucc(oL, oR) || (oL.Symbol == oR.Symbol && oL.AssocLeft()) {
		if oL.IsInfix() {
			s.reduceInfix(oL)
		} else if oL.IsPrefix() {
			s.reducePrefix()
		} else {
			message := fmt.Sprintf("\n  Illegal combination of operators %s in block %s and %s.", oL.Symbol, s.location(tm2.Node), oR.Symbol)
			if tm2.Node.Range.Begin.Line < tm1.Node.Range.Begin.Line && oR.Symbol == "=" {
				message = "\n  *** Hint *** You may have mistyped ==" + message
			} else if oR.Symbol == SanyVoidOperator().Symbol {
				message = fmt.Sprintf("\n Error following expression at  %s, missing operator or separator.", s.location(tm2.Node))
			}
			return s.failure(message, fmt.Sprintf("illegal combination of operators %s and %s", oL.Symbol, oR.Symbol))
		}
		return nil
	}
	if !(SanyOperatorPrec(oL, oR) || (oL.Symbol == oR.Symbol && oL.AssocRight())) {
		return s.failure(fmt.Sprintf("\n  Precedence conflict between ops %s in block %s and %s.", oL.Symbol, s.location(tm2.Node), oR.Symbol), fmt.Sprintf("precedence conflict between operators %s and %s", oL.Symbol, oR.Symbol))
	}
	return nil
}

// ReduceRecord ports the dedicated source record reduction, which replaces the
// top expression immediately rather than treating a dot as an infix operator.
func (s *SanyOperatorStack) ReduceRecord(middle, right *SanySyntaxNode) error {
	index := s.CurrentSize() - 1
	if index < 0 {
		return s.failure("\n    ``.'' has no left hand side at "+s.location(middle)+".", "record dot has no left hand side")
	}
	element := s.currentTop()[index]
	if element.isOperator() {
		previous := s.currentTop()[index-1]
		if element.Operator.IsPostfix() && !previous.isOperator() {
			s.Push(nil, nil)
			s.reducePostfix()
			index = s.CurrentSize() - 1
			topIndex := len(s.stackOfStacks) - 1
			s.stackOfStacks[topIndex] = s.currentTop()[:index]
			index--
			element = s.currentTop()[index]
		} else {
			return s.failure("\n    ``.'' follows operator "+s.location(element.Node)+".", "record dot follows an operator")
		}
	}
	left := s.currentTop()[index].Node
	record := NewSanyNode(SanySyntaxNodeKindByName["N_RecordComponent"], left, middle, right)
	s.stackOfStacks[len(s.stackOfStacks)-1][index] = sanyOperatorStackElement{Node: record}
	return nil
}

func (s *SanyOperatorStack) reduceInfix(op SanyOperatorInfo) {
	topIndex := len(s.stackOfStacks) - 1
	top := s.stackOfStacks[topIndex]
	n := len(top) - 1
	if n < 3 {
		return
	}
	left := top[n-3].Node
	opNode := top[n-2].Node
	right := top[n-1].Node
	var reduced *SanySyntaxNode
	if op.IsNfix() && left != nil && left.IsKind(SanySyntaxNodeKindByName["N_Times"]) {
		heirs := append(left.GetHeirs(), opNode, right)
		reduced = NewSanyNode(SanySyntaxNodeKindByName["N_Times"], heirs...)
	} else if op.IsNfix() {
		reduced = NewSanyNode(SanySyntaxNodeKindByName["N_Times"], left, opNode, right)
	} else if op.Symbol == "." {
		reduced = NewSanyNode(SanySyntaxNodeKindByName["N_RecordComponent"], left, opNode, right)
	} else {
		reduced = NewSanyNode(SanySyntaxNodeKindByName["N_InfixExpr"], left, opNode, right)
	}
	top = append(top[:n-2], top[n:]...)
	top[n-3] = sanyOperatorStackElement{Node: reduced}
	s.stackOfStacks[topIndex] = top
}

func (s *SanyOperatorStack) reducePrefix() {
	topIndex := len(s.stackOfStacks) - 1
	top := s.stackOfStacks[topIndex]
	n := len(top) - 1
	if n < 2 {
		return
	}
	reduced := NewSanyNode(SanySyntaxNodeKindByName["N_PrefixExpr"], top[n-2].Node, top[n-1].Node)
	top = append(top[:n-1], top[n:]...)
	top[n-2] = sanyOperatorStackElement{Node: reduced}
	s.stackOfStacks[topIndex] = top
}

func sanyFirstToken(node *SanySyntaxNode) *SanyToken {
	if node == nil {
		return nil
	}
	if node.Token != nil {
		return node.Token
	}
	for _, child := range node.GetHeirs() {
		if tok := sanyFirstToken(child); tok != nil {
			return tok
		}
	}
	return nil
}

func sanyLiteralTokenKind(literal string) (SanyTokenKind, bool) {
	for _, def := range SanyLiteralTokens {
		if def.Literal == literal {
			return def.Kind, true
		}
	}
	return SanyTokenInvalid, false
}

func (s *SanyOperatorStack) reducePostfix() {
	topIndex := len(s.stackOfStacks) - 1
	top := s.stackOfStacks[topIndex]
	n := len(top) - 1
	if n < 2 {
		return
	}
	op := top[n-1].Operator
	opNode := top[n-1].Node
	expr := top[n-2].Node
	var reduced *SanySyntaxNode
	if op != nil && op.Symbol == "[" {
		// Source reducePostfix puts the function in zero and the bracket
		// node's heirs in one; it does not retain an application wrapper.
		reduced = NewSanySplitNode(SanySyntaxNodeKindByName["N_FcnAppl"], []*SanySyntaxNode{expr}, opNode.GetHeirs())
	} else {
		reduced = NewSanyNode(SanySyntaxNodeKindByName["N_PostfixExpr"], expr, opNode)
	}
	top = append(top[:n-1], top[n:]...)
	top[n-2] = sanyOperatorStackElement{Node: reduced}
	s.stackOfStacks[topIndex] = top
}

func (s *SanyOperatorStack) PreInEmptyTop() bool {
	s.require()
	if len(s.stackOfStacks) == 0 {
		return true
	}
	if s.CurrentSize() == 0 {
		return true
	}
	top := s.stackOfStacks[len(s.stackOfStacks)-1]
	op := top[len(top)-1].Operator
	return op != nil && (op.IsPrefix() || op.IsInfix())
}

func (s *SanyOperatorStack) currentTop() []sanyOperatorStackElement {
	if len(s.stackOfStacks) == 0 {
		return nil
	}
	return s.stackOfStacks[len(s.stackOfStacks)-1]
}

func (e *sanyOperatorStackElement) isOperator() bool {
	return e.IsOperator()
}

func (op *SanyOperatorInfo) fixityName() string {
	op.require()
	switch op.Fix {
	case SanyOperatorPrefix:
		return "prefix"
	case SanyOperatorPostfix:
		return "postfix"
	case SanyOperatorInfix, SanyOperatorNfix:
		return "infix"
	default:
		return "unknown"
	}
}
