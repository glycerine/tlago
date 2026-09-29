package tlago

import "fmt"

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
	Fixity         SanyOperatorFixity
}

var SanyVoidOperator = SanyOperatorInfo{
	Symbol:         "$$_void",
	LowPrecedence:  0,
	HighPrecedence: 0,
	Associativity:  SanyAssociativityNone,
	Fixity:         SanyOperatorInfix,
}

func GetSanyOperator(symbol string) (SanyOperatorInfo, bool) {
	op, ok := SanyOperatorsBySymbol[symbol]
	return op, ok
}

func GetSanyMixfixOperator(op SanyOperatorInfo) (SanyOperatorInfo, bool) {
	if op.IsPrefix() {
		return op, true
	}
	return GetSanyOperator(op.Symbol + ".")
}

func ResolveSanyOperatorSynonym(symbol string) string {
	if canonical, ok := SanyOperatorSynonymCanonical[symbol]; ok {
		return canonical
	}
	return symbol
}

func (op SanyOperatorInfo) IsPrefix() bool {
	return op.Fixity == SanyOperatorPrefix
}

func (op SanyOperatorInfo) IsInfix() bool {
	return op.Fixity == SanyOperatorInfix || op.Fixity == SanyOperatorNfix
}

func (op SanyOperatorInfo) IsPostfix() bool {
	return op.Fixity == SanyOperatorPostfix
}

func (op SanyOperatorInfo) IsNfix() bool {
	return op.Fixity == SanyOperatorNfix
}

func (op SanyOperatorInfo) AssocLeft() bool {
	return op.Associativity == SanyAssociativityLeft
}

func (op SanyOperatorInfo) AssocRight() bool {
	return op.Associativity == SanyAssociativityRight
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

type SanyOperatorStack struct {
	stackOfStacks [][]sanyOperatorStackElement
}

func NewSanyOperatorStack() *SanyOperatorStack {
	return &SanyOperatorStack{}
}

func (s *SanyOperatorStack) NewStack() {
	s.stackOfStacks = append(s.stackOfStacks, nil)
}

func (s *SanyOperatorStack) PopStack() error {
	if len(s.stackOfStacks) == 0 {
		return fmt.Errorf("operator stack is empty")
	}
	s.stackOfStacks = s.stackOfStacks[:len(s.stackOfStacks)-1]
	return nil
}

func (s *SanyOperatorStack) Push(node *SanySyntaxNode, op *SanyOperatorInfo) {
	if len(s.stackOfStacks) == 0 {
		s.NewStack()
	}
	top := len(s.stackOfStacks) - 1
	s.stackOfStacks[top] = append(s.stackOfStacks[top], sanyOperatorStackElement{Node: node, Operator: op})
}

func (s *SanyOperatorStack) CurrentSize() int {
	if len(s.stackOfStacks) == 0 {
		return 0
	}
	return len(s.stackOfStacks[len(s.stackOfStacks)-1])
}

func (s *SanyOperatorStack) TopNode() *SanySyntaxNode {
	if s.CurrentSize() == 0 {
		return nil
	}
	top := s.stackOfStacks[len(s.stackOfStacks)-1]
	return top[len(top)-1].Node
}

func (s *SanyOperatorStack) TopOperator() *SanyOperatorInfo {
	if s.CurrentSize() == 0 {
		return nil
	}
	top := s.stackOfStacks[len(s.stackOfStacks)-1]
	return top[len(top)-1].Operator
}

func (s *SanyOperatorStack) FinalReduce() (*SanySyntaxNode, error) {
	s.Push(nil, &SanyVoidOperator)
	if err := s.ReduceStack(); err != nil {
		return nil, err
	}
	if !s.IsWellReduced() {
		return nil, fmt.Errorf("could not reduce expression stack")
	}
	return s.stackOfStacks[len(s.stackOfStacks)-1][0].Node, nil
}

func (s *SanyOperatorStack) IsWellReduced() bool {
	return s.CurrentSize() == 2
}

func (s *SanyOperatorStack) ReduceStack() error {
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

func (s *SanyOperatorStack) reduceWithRightPostfix(n int, oR SanyOperatorInfo) error {
	if n == 0 {
		return fmt.Errorf("postfix operator %s on empty stack", oR.Symbol)
	}
	top := s.currentTop()
	tm1 := top[n-1]
	if tm1.isOperator() {
		oL := *tm1.Operator
		if oL.IsInfix() || oL.IsPrefix() {
			return fmt.Errorf("postfix operator %s follows %s operator %s", oR.Symbol, oL.fixityName(), oL.Symbol)
		}
		s.reducePostfix()
		return nil
	}
	if n <= 1 {
		return nil
	}
	tm2 := top[n-2]
	if !tm2.isOperator() {
		return fmt.Errorf("adjacent expressions without intervening operator")
	}
	oL := *tm2.Operator
	if SanyOperatorSucc(oL, oR) {
		if oL.IsInfix() {
			s.reduceInfix(oL)
		} else {
			s.reducePrefix(oL)
		}
		return nil
	}
	if !SanyOperatorPrec(oL, oR) {
		return fmt.Errorf("precedence conflict between operators %s and %s", oL.Symbol, oR.Symbol)
	}
	return nil
}

func (s *SanyOperatorStack) reduceWithRightPrefix(n int, oR SanyOperatorInfo) error {
	if n == 0 {
		return nil
	}
	tm1 := s.currentTop()[n-1]
	if tm1.isOperator() {
		oL := *tm1.Operator
		if oL.IsPostfix() {
			return fmt.Errorf("prefix operator %s follows postfix operator %s", oR.Symbol, oL.Symbol)
		}
		return nil
	}
	return fmt.Errorf("prefix operator %s follows an expression", oR.Symbol)
}

func (s *SanyOperatorStack) reduceWithRightInfix(n int, oR SanyOperatorInfo) error {
	if n == 0 {
		if _, ok := GetSanyMixfixOperator(oR); !ok {
			return fmt.Errorf("infix operator %s on empty stack", oR.Symbol)
		}
		return nil
	}
	top := s.currentTop()
	tm1 := top[n-1]
	if tm1.isOperator() {
		oL := *tm1.Operator
		if oL.IsInfix() || oL.IsPrefix() {
			mixR, ok := GetSanyMixfixOperator(oR)
			if !ok {
				return fmt.Errorf("infix operator %s follows %s operator %s", oR.Symbol, oL.fixityName(), oL.Symbol)
			}
			if SanyOperatorSucc(oL, mixR) || (oL.Symbol == mixR.Symbol && oL.AssocLeft()) {
				return fmt.Errorf("precedence conflict between operators %s and %s", oL.Symbol, mixR.Symbol)
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
		return fmt.Errorf("adjacent expressions without intervening operator")
	}
	oL := *tm2.Operator
	if mixL, ok := GetSanyMixfixOperator(oL); ok && (n == 2 || top[n-3].isOperator()) {
		oL = mixL
	}
	if SanyOperatorSucc(oL, oR) || (oL.Symbol == oR.Symbol && oL.AssocLeft()) {
		if oL.IsInfix() {
			s.reduceInfix(oL)
		} else if oL.IsPrefix() {
			s.reducePrefix(oL)
		} else {
			return fmt.Errorf("illegal combination of operators %s and %s", oL.Symbol, oR.Symbol)
		}
		return nil
	}
	if !(SanyOperatorPrec(oL, oR) || (oL.Symbol == oR.Symbol && oL.AssocRight())) {
		return fmt.Errorf("precedence conflict between operators %s and %s", oL.Symbol, oR.Symbol)
	}
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

func (s *SanyOperatorStack) reducePrefix(op SanyOperatorInfo) {
	topIndex := len(s.stackOfStacks) - 1
	top := s.stackOfStacks[topIndex]
	n := len(top) - 1
	if n < 2 {
		return
	}
	reduced := NewSanyNode(SanySyntaxNodeKindByName["N_PrefixExpr"], sanyMixfixOperatorNode(top[n-2].Node, op), top[n-1].Node)
	top = append(top[:n-1], top[n:]...)
	top[n-2] = sanyOperatorStackElement{Node: reduced}
	s.stackOfStacks[topIndex] = top
}

func sanyMixfixOperatorNode(node *SanySyntaxNode, op SanyOperatorInfo) *SanySyntaxNode {
	if node == nil {
		return nil
	}
	kindName := "N_GenInfixOp"
	if op.IsPrefix() {
		kindName = "N_GenPrefixOp"
	} else if op.IsPostfix() {
		kindName = "N_GenPostfixOp"
	}
	kind := SanySyntaxNodeKindByName[kindName]
	if node.Kind == kind && sanyOperatorImage(node) == op.Symbol {
		return node
	}
	prefix := NewSanyNode(SanySyntaxNodeKindByName["N_IdPrefix"])
	if heirs := node.GetHeirs(); len(heirs) > 0 && heirs[0].Kind.JavaName() == "N_IdPrefix" {
		prefix = heirs[0]
	}
	return NewSanyNode(kind, prefix, sanyRetargetOperatorTokenNode(node, op.Symbol))
}

func sanyRetargetOperatorTokenNode(node *SanySyntaxNode, symbol string) *SanySyntaxNode {
	tok := sanyFirstToken(node)
	if tok == nil {
		return nil
	}
	copied := *tok
	copied.Image = symbol
	if kind, ok := sanyLiteralTokenKind(symbol); ok {
		copied.Kind = kind
	}
	return NewSanyTokenNode(&copied)
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
	kind := SanySyntaxNodeKindByName["N_PostfixExpr"]
	if op != nil && op.Symbol == "[" {
		kind = SanySyntaxNodeKindByName["N_FcnAppl"]
	}
	reduced := NewSanyNode(kind, expr, opNode)
	top = append(top[:n-1], top[n:]...)
	top[n-2] = sanyOperatorStackElement{Node: reduced}
	s.stackOfStacks[topIndex] = top
}

func (s *SanyOperatorStack) PreInEmptyTop() bool {
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

func (e sanyOperatorStackElement) isOperator() bool {
	return e.Operator != nil
}

func (op SanyOperatorInfo) fixityName() string {
	switch op.Fixity {
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
