package tlago

import "github.com/glycerine/tlago/tlc"

type Spec struct {
	levelChecks      map[*Module]*sanyModuleLevelChecks
	initialContext   *sanyContext
	FilenameResolver tlc.FilenameToStream
	LibraryPaths     []string
	Root             *Module
	Modules          map[string]*Module
	SemanticOrder    []string
	ModuleFiles      []string // Logical filenames supplied to the filename resolver.
	Diags            Diagnostics
	SemanticDiags    Diagnostics // SANY semantic Errors, retained even if generation throws.
}

type Module struct {
	symbolTable      *sanySymbolTable
	declarationNodes []*sanySemOpDeclNode
	semanticNode     *sanySemModuleNode
	generatorNodes   *sanyGeneratorNodes
	Syntax           *SanySyntaxNode
	Name             string
	SourcePath       string
	Source           string
	Library          bool
	Pos              Position
	Extends          []string
	Instances        []Instance
	Declarations     []Declaration
	Recursives       []Declaration
	Definitions      []Definition
	Assumptions      []NamedExpr
	Theorems         []NamedExpr
	ProofRefs        []ProofRef
	ProofRefNodes    []*SanySyntaxNode
	Proofs           []ProofSummary
	Nested           []*Module

	ImplicitExtends []string // Runtime additions, absent from the source module context.
}

type Instance struct {
	generatedSubstitutions []sanyGeneratedSubstitution
	Syntax                 *SanySyntaxNode
	Name                   string
	Params                 []string
	ParamArities           map[string]int
	ParamPositions         map[string]Position
	Module                 string
	Substitutions          map[string]Expr
	SubstitutionList       []Substitution
	Local                  bool
	PreComments            []string
	Pos                    Position
	Source                 Position
	LHSPos                 Position
}

func (i Instance) SourcePosition() Position {
	if i.Source.Line > 0 || i.Source.Column > 0 || i.Source.File != "" {
		return i.Source
	}
	return i.Pos
}

type Substitution struct {
	Name string
	Expr Expr
	Pos  Position
}

func (i Instance) qualifier() string {
	if i.Name != "" {
		return i.Name
	}
	return i.Module
}

func (i Instance) exportsUnqualified() bool {
	return i.Name == ""
}

type DeclarationKind string

const (
	ConstantDecl  DeclarationKind = "CONSTANT"
	VariableDecl  DeclarationKind = "VARIABLE"
	RecursiveDecl DeclarationKind = "RECURSIVE"
	OperatorDecl  DeclarationKind = "OPERATOR"
	InstanceDecl  DeclarationKind = "INSTANCE"
)

type Declaration struct {
	Syntax          *SanySyntaxNode
	Kind            DeclarationKind
	Names           []string
	Arities         map[string]int
	NamePositions   map[string]Position
	NamePreComments map[string][]string
	Pos             Position
}

type Definition struct {
	semanticNode    *sanySemOpDefNode
	Syntax          *SanySyntaxNode
	Name            string
	Params          []string
	ParamArities    map[string]int
	ParamPositions  map[string]Position
	PreComments     []string
	Expr            Expr
	Local           bool
	FunctionDef     bool
	AssumeProve     bool
	AssumeProveBody *AssumeProve
	TheoremLike     bool
	FactKind        string
	FactKeyword     string
	Pos             Position
	Source          Position
	LHSPos          Position
}

func (d Definition) SourcePosition() Position {
	if d.Source.Line > 0 || d.Source.Column > 0 || d.Source.File != "" {
		return d.Source
	}
	return d.Pos
}

func (d Definition) DeclarationPosition() Position {
	if d.LHSPos.Line > 0 || d.LHSPos.Column > 0 || d.LHSPos.File != "" {
		return d.LHSPos
	}
	return d.SourcePosition()
}

type NamedExpr struct {
	Name            string
	Expr            Expr
	AssumeProve     bool
	AssumeProveBody *AssumeProve
	PreComments     []string
	Pos             Position
	Source          Position
	Syntax          *SanySyntaxNode
}

func (e NamedExpr) Position() Position {
	if e.Pos.Line > 0 || e.Pos.Column > 0 || e.Pos.File != "" {
		return e.Pos
	}
	if e.Expr != nil {
		return e.Expr.Position()
	}
	return Position{}
}

func (e NamedExpr) SourcePosition() Position {
	if e.Source.Line > 0 || e.Source.Column > 0 || e.Source.File != "" {
		return e.Source
	}
	return e.Position()
}

type ProofRef struct {
	Module string
	Syntax *SanySyntaxNode
	Expr   Expr
	Name   string
	Mode   string
	Defs   bool
	Pos    Position
}

type ProofFact struct {
	Expr   Expr
	Direct bool
}

type ProofSummary struct {
	LeafRefs []ProofRef
	Syntax   *SanySyntaxNode
	Facts    []ProofFact
	Goal     Expr
	Steps    []ProofStep
	Pos      Position
}

type ProofStep struct {
	formalNodes     []*sanyFormalParamNode
	Syntax          *SanySyntaxNode
	Definitions     []Definition
	Instances       []Instance
	LeafRefs        []ProofRef
	UseHideRefs     []ProofRef
	Statement       Position
	Facts           []ProofFact
	QualifiedName   string
	AssumeProveBody *AssumeProve
	Suffices        bool
	Name            string
	Implicit        bool
	Depth           int
	Kind            string
	Expr            Expr
	Exprs           []Expr
	Bounds          []BoundVar
	Refs            []string
	Pos             Position
}

type AssumeProve struct {
	Assumptions []AssumeProveItem
	Prove       Expr
	Pos         Position
}

type AssumeProveItem struct {
	Expr      Expr
	Nested    *AssumeProve
	NewSymbol *NewSymbol
}

type NewSymbol struct {
	Name   string
	Arity  int
	Kind   int
	Level  tlaLevel
	Domain Expr
	Pos    Position
	Source Position
}

// SanyExprSource retains the parser node selected by semantic generation.
// Parenthesized expressions retain their inner expression node.
type SanyExprSource struct {
	semanticGraph              sanySemanticGraphNode
	definitionFormals          []*sanyFormalParamNode
	operatorArgumentsGenerated bool
	Syntax                     *SanySyntaxNode
	// Selector keeps the generator's per-step argument syntax. A flattened
	// CallExpr argument list cannot distinguish Op(a)!lab(b) from Op(a,b)!lab.
	Selector           *SanySelector
	selection          *sanySelectorSelection
	generationArity    *int
	selectorDiagnostic *Diagnostic
	selectorFailure    bool
	generationFailure  sanyGenerationFailure
}

func (s *SanyExprSource) GetSyntaxNode() *SanySyntaxNode    { return s.Syntax }
func (s *SanyExprSource) generationSource() *SanyExprSource { return s }
func (s *SanyExprSource) SetSyntaxNode(node *SanySyntaxNode) {
	s.Syntax = node
	s.Selector = sanySelectorFromSyntax(node)
}

type Expr interface {
	Position() Position
	exprNode()
}

type IdentExpr struct {
	SanyExprSource
	formalNode      *sanyFormalParamNode
	declarationNode *sanySemOpDeclNode
	proofAtTarget   Expr // Generated $Nop operand reuses the previous infix RHS.
	Name            string
	Pos             Position
}

func (e *IdentExpr) Position() Position { return e.Pos }
func (*IdentExpr) exprNode()            {}

type LiteralExpr struct {
	SanyExprSource
	numeralNode *tlc.NumeralNode
	decimalNode *tlc.DecimalNode
	stringNode  *tlc.StringNode
	Kind        string
	Value       string
	Pos         Position
}

func (e *LiteralExpr) Position() Position { return e.Pos }
func (*LiteralExpr) exprNode()            {}

type UnaryExpr struct {
	SanyExprSource
	Op   string
	Expr Expr
	Pos  Position
}

func (e *UnaryExpr) Position() Position { return e.Pos }
func (*UnaryExpr) exprNode()            {}

type BinaryExpr struct {
	SanyExprSource
	Op           string
	Left         Expr
	Right        Expr
	Pos          Position
	JunctionList bool
	SanyNary     bool
}

func (e *BinaryExpr) Position() Position { return e.Pos }
func (*BinaryExpr) exprNode()            {}

type CallExpr struct {
	SanyExprSource
	Callee Expr
	Args   []Expr
	Pos    Position
}

func (e *CallExpr) Position() Position { return e.Pos }
func (*CallExpr) exprNode()            {}

type IfExpr struct {
	SanyExprSource
	Cond Expr
	Then Expr
	Else Expr
	Pos  Position
}

func (e *IfExpr) Position() Position { return e.Pos }
func (*IfExpr) exprNode()            {}

type LetExpr struct {
	SanyExprSource
	instanceDefinitions []sanySelectorDefinition
	Recursives          []Declaration
	Definitions         []Definition
	Instances           []Instance
	Body                Expr
	Pos                 Position
}

func (e *LetExpr) Position() Position { return e.Pos }
func (*LetExpr) exprNode()            {}

type QuantifierExpr struct {
	SanyExprSource
	formalNode        *sanyFormalParamNode
	quantifierFormals []*sanyFormalParamNode
	Kind              string
	Var               string
	VarPos            Position
	Set               Expr
	Body              Expr
	OperatorArity     int
	HasOperatorArity  bool
	TupleBound        bool
	LevelKnown        bool
	Level             int
	Pos               Position
}

func (e *QuantifierExpr) Position() Position { return e.Pos }
func (*QuantifierExpr) exprNode()            {}

type CaseArm struct {
	Test  Expr
	Value Expr
	Pos   Position
}

type CaseExpr struct {
	SanyExprSource
	Arms     []CaseArm
	Other    Expr
	OtherPos Position
	Pos      Position
}

func (e *CaseExpr) Position() Position { return e.Pos }
func (*CaseExpr) exprNode()            {}

type ChooseExpr struct {
	SanyExprSource
	formalNodes []*sanyFormalParamNode
	// TupleVars retains each formal parameter of CHOOSE <<x, ...>>.
	TupleVars []BoundVar
	Var       string
	VarPos    Position
	Set       Expr
	Body      Expr
	Pos       Position
}

func (e *ChooseExpr) Position() Position { return e.Pos }
func (*ChooseExpr) exprNode()            {}

func (e *ChooseExpr) boundVars() []BoundVar {
	if e.TupleVars != nil {
		bounds := append([]BoundVar(nil), e.TupleVars...)
		for i := range bounds {
			bounds[i].Set = e.Set
		}
		return bounds
	}
	pos := e.VarPos
	if pos.Line == 0 && pos.Column == 0 && pos.File == "" {
		pos = e.Pos
	}
	return []BoundVar{{Name: e.Var, Set: e.Set, Pos: pos}}
}

func (e *ChooseExpr) boundNames() []string {
	bounds := e.boundVars()
	names := make([]string, len(bounds))
	for i, bound := range bounds {
		names[i] = bound.Name
	}
	return names
}

func (e *ChooseExpr) bindsName(name string) bool {
	for _, bound := range e.boundVars() {
		if bound.Name == name {
			return true
		}
	}
	return false
}

type TupleExpr struct {
	SanyExprSource
	Elems []Expr
	Pos   Position
}

func (e *TupleExpr) Position() Position { return e.Pos }
func (*TupleExpr) exprNode()            {}

type SetExpr struct {
	SanyExprSource
	Elems []Expr
	Pos   Position
}

func (e *SetExpr) Position() Position { return e.Pos }
func (*SetExpr) exprNode()            {}

type RecordField struct {
	Name   string
	Value  Expr
	Pos    Position
	Source Position
}

type RecordExpr struct {
	SanyExprSource
	Fields []RecordField
	Pos    Position
}

func (e *RecordExpr) Position() Position { return e.Pos }
func (*RecordExpr) exprNode()            {}

type RecordComponentExpr struct {
	SanyExprSource
	Record   Expr
	Field    string
	FieldPos Position
	Pos      Position
}

func (e *RecordComponentExpr) Position() Position { return e.Pos }
func (*RecordComponentExpr) exprNode()            {}

type RecordSetField struct {
	Name   string
	Set    Expr
	Pos    Position
	Source Position
}

type RecordSetExpr struct {
	SanyExprSource
	Fields []RecordSetField
	Pos    Position
}

func (e *RecordSetExpr) Position() Position { return e.Pos }
func (*RecordSetExpr) exprNode()            {}

type BoundVar struct {
	Name             string
	Set              Expr
	Pos              Position
	OperatorArity    int
	HasOperatorArity bool
	TupleBound       bool
	LevelKnown       bool
	Level            int
}

type FunctionExpr struct {
	lambdaNode *sanySemOpDefNode
	SanyExprSource
	definitionFormalContext map[string]localSymbol
	// Preparation retains the temporary self formal; Java removes it from
	// a nonrecursive function's final OpApplNode array.
	functionSymbol          *sanyFormalParamNode
	constructorSymbol       localSymbol
	constructorSymbolExists bool
	formalNodes             []*sanyFormalParamNode
	Bounds                  []BoundVar
	Body                    Expr
	IsLambda                bool
	Pos                     Position
	PreComments             []string
}

func (e *FunctionExpr) Position() Position { return e.Pos }
func (*FunctionExpr) exprNode()            {}

type FunctionAppExpr struct {
	SanyExprSource
	Function Expr
	Args     []Expr
	Pos      Position
}

func (e *FunctionAppExpr) Position() Position { return e.Pos }
func (*FunctionAppExpr) exprNode()            {}

type ExceptComponent struct {
	Field    string
	FieldPos Position
	Indices  []Expr
	Pos      Position
}

type ExceptSpec struct {
	Components []ExceptComponent
	Value      Expr
	Pos        Position
}

type ExceptExpr struct {
	SanyExprSource
	Base  Expr
	Specs []ExceptSpec
	Pos   Position
}

func (e *ExceptExpr) Position() Position { return e.Pos }
func (*ExceptExpr) exprNode()            {}

type LabelExpr struct {
	formalNodes            []*sanyFormalParamNode
	illegalParameterSyntax []*SanySyntaxNode
	SanyExprSource
	Name   string
	Params []string
	Body   Expr
	Pos    Position
}

func (e *LabelExpr) Position() Position { return e.Pos }
func (*LabelExpr) exprNode()            {}

type ActionExpr struct {
	SanyExprSource
	Kind      string
	Action    Expr
	Subscript Expr
	Pos       Position
}

func (e *ActionExpr) Position() Position { return e.Pos }
func (*ActionExpr) exprNode()            {}

type FairnessExpr struct {
	SanyExprSource
	Kind      string
	Subscript Expr
	Action    Expr
	Pos       Position
}

func (e *FairnessExpr) Position() Position { return e.Pos }
func (*FairnessExpr) exprNode()            {}

type FunctionSetExpr struct {
	SanyExprSource
	Domain Expr
	Range  Expr
	Pos    Position
}

func (e *FunctionSetExpr) Position() Position { return e.Pos }
func (*FunctionSetExpr) exprNode()            {}

type SetComprehensionExpr struct {
	SanyExprSource
	formalNodes []*sanyFormalParamNode
	Element     Expr
	Bounds      []BoundVar
	Predicate   Expr
	Pos         Position
}

func (e *SetComprehensionExpr) Position() Position { return e.Pos }
func (*SetComprehensionExpr) exprNode()            {}
