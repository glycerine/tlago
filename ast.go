package tlago

type Spec struct {
	Root    *Module
	Modules map[string]*Module
	Diags   Diagnostics
}

type Module struct {
	Name         string
	SourcePath   string
	Source       string
	Pos          Position
	Extends      []string
	Instances    []Instance
	Declarations []Declaration
	Recursives   []Declaration
	Definitions  []Definition
	Assumptions  []NamedExpr
	Theorems     []NamedExpr
	ProofRefs    []ProofRef
	Proofs       []ProofSummary
	Nested       []*Module
}

type Instance struct {
	Name             string
	Module           string
	Substitutions    map[string]Expr
	SubstitutionList []Substitution
	Local            bool
	PreComments      []string
	Pos              Position
	Source           Position
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
)

type Declaration struct {
	Kind          DeclarationKind
	Names         []string
	Arities       map[string]int
	NamePositions map[string]Position
	Pos           Position
}

type Definition struct {
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
	Pos             Position
	Source          Position
}

func (d Definition) SourcePosition() Position {
	if d.Source.Line > 0 || d.Source.Column > 0 || d.Source.File != "" {
		return d.Source
	}
	return d.Pos
}

type NamedExpr struct {
	Name   string
	Expr   Expr
	Pos    Position
	Source Position
	Syntax *SanySyntaxNode
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
	Name string
	Mode string
	Pos  Position
}

type ProofSummary struct {
	Goal  Expr
	Steps []ProofStep
	Pos   Position
}

type ProofStep struct {
	Name     string
	Implicit bool
	Depth    int
	Kind     string
	Expr     Expr
	Exprs    []Expr
	Bounds   []BoundVar
	Refs     []string
	Pos      Position
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

type Expr interface {
	Position() Position
	exprNode()
}

type IdentExpr struct {
	Name string
	Pos  Position
}

func (e *IdentExpr) Position() Position { return e.Pos }
func (*IdentExpr) exprNode()            {}

type LiteralExpr struct {
	Kind  string
	Value string
	Pos   Position
}

func (e *LiteralExpr) Position() Position { return e.Pos }
func (*LiteralExpr) exprNode()            {}

type UnaryExpr struct {
	Op   string
	Expr Expr
	Pos  Position
}

func (e *UnaryExpr) Position() Position { return e.Pos }
func (*UnaryExpr) exprNode()            {}

type BinaryExpr struct {
	Op           string
	Left         Expr
	Right        Expr
	Pos          Position
	JunctionList bool
}

func (e *BinaryExpr) Position() Position { return e.Pos }
func (*BinaryExpr) exprNode()            {}

type CallExpr struct {
	Callee Expr
	Args   []Expr
	Pos    Position
}

func (e *CallExpr) Position() Position { return e.Pos }
func (*CallExpr) exprNode()            {}

type IfExpr struct {
	Cond Expr
	Then Expr
	Else Expr
	Pos  Position
}

func (e *IfExpr) Position() Position { return e.Pos }
func (*IfExpr) exprNode()            {}

type LetExpr struct {
	Recursives  []Declaration
	Definitions []Definition
	Instances   []Instance
	Body        Expr
	Pos         Position
}

func (e *LetExpr) Position() Position { return e.Pos }
func (*LetExpr) exprNode()            {}

type QuantifierExpr struct {
	Kind   string
	Var    string
	VarPos Position
	Set    Expr
	Body   Expr
	Pos    Position
}

func (e *QuantifierExpr) Position() Position { return e.Pos }
func (*QuantifierExpr) exprNode()            {}

type CaseArm struct {
	Test  Expr
	Value Expr
	Pos   Position
}

type CaseExpr struct {
	Arms     []CaseArm
	Other    Expr
	OtherPos Position
	Pos      Position
}

func (e *CaseExpr) Position() Position { return e.Pos }
func (*CaseExpr) exprNode()            {}

type ChooseExpr struct {
	Var    string
	VarPos Position
	Set    Expr
	Body   Expr
	Pos    Position
}

func (e *ChooseExpr) Position() Position { return e.Pos }
func (*ChooseExpr) exprNode()            {}

type TupleExpr struct {
	Elems []Expr
	Pos   Position
}

func (e *TupleExpr) Position() Position { return e.Pos }
func (*TupleExpr) exprNode()            {}

type SetExpr struct {
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
	Fields []RecordField
	Pos    Position
}

func (e *RecordExpr) Position() Position { return e.Pos }
func (*RecordExpr) exprNode()            {}

type RecordComponentExpr struct {
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
	Fields []RecordSetField
	Pos    Position
}

func (e *RecordSetExpr) Position() Position { return e.Pos }
func (*RecordSetExpr) exprNode()            {}

type BoundVar struct {
	Name string
	Set  Expr
	Pos  Position
}

type FunctionExpr struct {
	Bounds   []BoundVar
	Body     Expr
	IsLambda bool
	Pos      Position
}

func (e *FunctionExpr) Position() Position { return e.Pos }
func (*FunctionExpr) exprNode()            {}

type FunctionAppExpr struct {
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
	Base  Expr
	Specs []ExceptSpec
	Pos   Position
}

func (e *ExceptExpr) Position() Position { return e.Pos }
func (*ExceptExpr) exprNode()            {}

type LabelExpr struct {
	Name   string
	Params []string
	Body   Expr
	Pos    Position
}

func (e *LabelExpr) Position() Position { return e.Pos }
func (*LabelExpr) exprNode()            {}

type ActionExpr struct {
	Kind      string
	Action    Expr
	Subscript Expr
	Pos       Position
}

func (e *ActionExpr) Position() Position { return e.Pos }
func (*ActionExpr) exprNode()            {}

type FairnessExpr struct {
	Kind      string
	Subscript Expr
	Action    Expr
	Pos       Position
}

func (e *FairnessExpr) Position() Position { return e.Pos }
func (*FairnessExpr) exprNode()            {}

type FunctionSetExpr struct {
	Domain Expr
	Range  Expr
	Pos    Position
}

func (e *FunctionSetExpr) Position() Position { return e.Pos }
func (*FunctionSetExpr) exprNode()            {}

type SetComprehensionExpr struct {
	Element   Expr
	Bounds    []BoundVar
	Predicate Expr
	Pos       Position
}

func (e *SetComprehensionExpr) Position() Position { return e.Pos }
func (*SetComprehensionExpr) exprNode()            {}
