package tlago

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestSanySemanticBridgeBehaviors(t *testing.T) {
	t.Run("checks duplicate and undefined names through SANY syntax", func(t *testing.T) {
		_, diags := CheckSanySource("Bad.tla", `---- MODULE Bad ----
VARIABLE x, x
Init == y = 0
====`)
		// Java reports the same-kind/arity declaration as warning 4801,
		// followed by error 4200 for y. Keep the exact source fixture.
		if len(diags) != 2 || diags[0].Severity != SeverityWarning || diags[0].Code != "W4801" {
			t.Fatalf("want warning 4801 followed by the undefined-name error: %v", diags)
		}
		message := "Multiple declarations or definitions for symbol x.  \nThis duplicates the one at line 2, col 10 to line 2, col 10 of module Bad."
		if diags[0].SANYMessage != message {
			t.Fatalf("duplicate declaration message=%q, want %q", diags[0].SANYMessage, message)
		}
		location := diags[0].SANYRange
		if location.Begin.Line != 2 || location.Begin.Column != 13 || location.End.Line != 2 || location.End.Column != 13 {
			t.Fatalf("duplicate declaration range=%v, want 2:13 to 2:13", location)
		}
		requireHasErrorContaining(t, diags, "undefined")
	})

	t.Run("checks operator call arity through SANY syntax", func(t *testing.T) {
		_, diags := CheckSanySource("BadArity.tla", `---- MODULE BadArity ----
Inc(n) == n + 1
Bad == Inc(1, 2)
====`)
		requireHasErrorContaining(t, diags, "arity")
	})

	t.Run("checks higher-order operator argument arity through SANY syntax", func(t *testing.T) {
		_, diags := CheckSanySource("HigherOrderArgs.tla", `---- MODULE HigherOrderArgs ----
NeedsUnary(F(_)) == 0
NeedsExpr(x) == 0
Zero == 0
BadExpr == NeedsUnary({})
BadZero == NeedsUnary(Zero)
BadLambda == NeedsUnary(LAMBDA x, y : 0)
BadLambdaExpr == NeedsExpr(LAMBDA x : 0)
====`)
		requireHasErrorContaining(t, diags, "operator parameter F")
		requireHasErrorContaining(t, diags, "operator argument arity")
		requireHasErrorContaining(t, diags, "expression parameter x")
	})

	t.Run("allows function constructors as expression arguments", func(t *testing.T) {
		_, diags := CheckSanySource("FunctionConstructorArg.tla", `---- MODULE FunctionConstructorArg ----
EXTENDS Naturals
Use(f) == f[1]
OK == Use([i \in {1} |-> i])
====`)
		requireNoErrors(t, diags)
	})

	t.Run("checks duplicate operator parameters through SANY syntax", func(t *testing.T) {
		_, diags := CheckSanySource("BadParams.tla", `---- MODULE BadParams ----
Dup(n, n) == n
====`)
		requireHasErrorContaining(t, diags, "duplicate")
		requireHasErrorContaining(t, diags, "parameter")
	})

	t.Run("checks duplicate LET operator parameters through SANY syntax", func(t *testing.T) {
		_, diags := CheckSanySource("BadLetParams.tla", `---- MODULE BadLetParams ----
Outer == LET Dup(n, n) == n IN Dup(1)
====`)
		requireHasErrorContaining(t, diags, "duplicate")
		requireHasErrorContaining(t, diags, "parameter")
	})

	t.Run("captures LET-local module instances in the central AST", func(t *testing.T) {
		mod, diags := ParseSanyModuleSource("LetInstance.tla", `---- MODULE LetInstance ----
Use == LET I == INSTANCE Inner IN I!Op
====`)
		requireNoErrors(t, diags)
		if len(mod.Definitions) != 1 {
			t.Fatalf("definitions = %d, want 1", len(mod.Definitions))
		}
		let, ok := mod.Definitions[0].Expr.(*LetExpr)
		if !ok {
			t.Fatalf("definition expression = %T, want *LetExpr", mod.Definitions[0].Expr)
		}
		if len(let.Instances) != 1 {
			t.Fatalf("LET instances = %d, want 1", len(let.Instances))
		}
		if got := let.Instances[0]; got.Name != "I" || got.Module != "Inner" {
			t.Fatalf("LET instance = %#v, want I == INSTANCE Inner", got)
		}
	})

	t.Run("loads modules referenced only by LET-local instances", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "Root.tla")
		writeFile(t, filepath.Join(dir, "Inner.tla"), `---- MODULE Inner ----
Op == 1
====`)
		writeFile(t, root, `---- MODULE Root ----
Use == LET I == INSTANCE Inner IN I!Op
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		if spec.Modules["Inner"] == nil {
			t.Fatalf("LET-local instance target Inner was not loaded")
		}
	})

	t.Run("loads sibling modules declared later in the same TLA file", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "Root.tla")
		writeFile(t, root, `---- MODULE Root ----
EXTENDS Common
Use == Helper
====

---- MODULE Common ----
Helper == TRUE
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		common := spec.Modules["Common"]
		if common == nil {
			t.Fatalf("same-file sibling module Common was not registered")
		}
		if common.Pos.File != "Common" || common.Pos.Line != 1 {
			t.Fatalf("Common position = %s:%d, want Common:1", common.Pos.File, common.Pos.Line)
		}
		if len(common.Definitions) != 1 || common.Definitions[0].SourcePosition().File != "Common" || common.Definitions[0].SourcePosition().Line != 2 {
			t.Fatalf("Common helper position = %#v, want Common line 2", common.Definitions)
		}
		requireNoErrors(t, CheckSpec(spec))
	})

	t.Run("ignores trailing non-module text after the root module", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "Root.tla")
		writeFile(t, root, `---- MODULE Root ----
Op == TRUE
====
$ echo run this with a shell; false || true
`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		if spec.Root == nil || spec.Root.Name != "Root" {
			t.Fatalf("root module = %#v, want Root", spec.Root)
		}
		requireNoErrors(t, CheckSpec(spec))
	})

	t.Run("records SANY semantic analysis order", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "ExtBase.tla"), `---- MODULE ExtBase ----
BaseOp == TRUE
====`)
		writeFile(t, filepath.Join(dir, "Extender.tla"), `---- MODULE Extender ----
EXTENDS ExtBase
ExtOp == BaseOp
====`)
		writeFile(t, filepath.Join(dir, "InstBase.tla"), `---- MODULE InstBase ----
InstBaseOp == TRUE
====`)
		writeFile(t, filepath.Join(dir, "Instancer.tla"), `---- MODULE Instancer ----
EXTENDS InstBase
InstOp == InstBaseOp
====`)
		root := filepath.Join(dir, "Root.tla")
		writeFile(t, root, `---- MODULE Root ----
EXTENDS Extender
INSTANCE Instancer
Use == ExtOp /\ InstOp
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		got := strings.Join(spec.SemanticOrder, ",")
		want := "ExtBase,Extender,InstBase,Instancer,Root"
		if got != want {
			t.Fatalf("semantic order = %s, want %s", got, want)
		}
	})

	t.Run("checks local name collisions and bare operator arity", func(t *testing.T) {
		_, diags := CheckSanySource("BindingErrors.tla", `---- MODULE BindingErrors ----
VARIABLE x
BadParam(x) == TRUE
BadBound == \A x \in {} : TRUE
NeedsArg(y) == y
BadRef == NeedsArg
====`)
		requireHasErrorContaining(t, diags, "parameter x")
		requireHasErrorContaining(t, diags, "bound symbol x")
		requireHasErrorContaining(t, diags, "got 0 args")
	})

	t.Run("allows bound names before later top-level definitions", func(t *testing.T) {
		_, diags := CheckSanySource("BoundBeforeLaterDef.tla", `---- MODULE BoundBeforeLaterDef ----
CONSTANT S
Init == [t \in S |-> t]
t(self) == self
====`)
		requireNoErrors(t, diags)
	})

	t.Run("allows parameters before later top-level declarations", func(t *testing.T) {
		_, diags := CheckSanySource("ParamBeforeLaterDecl.tla", `---- MODULE ParamBeforeLaterDecl ----
Op(state) == state = state
VARIABLE state
====`)
		requireNoErrors(t, diags)
	})

	t.Run("allows LET definition parameters before later LET definitions", func(t *testing.T) {
		_, diags := CheckSanySource("LetParamBeforeLaterDef.tla", `---- MODULE LetParamBeforeLaterDef ----
Use == LET F(S) == S
           S == TRUE
       IN F(S)
====`)
		requireNoErrors(t, diags)
	})

	t.Run("allows constant-level substitutions for variable instance parameters", func(t *testing.T) {
		dir := t.TempDir()
		helper := filepath.Join(dir, "Helper.tla")
		root := filepath.Join(dir, "Root.tla")
		writeFile(t, helper, `---- MODULE Helper ----
VARIABLE V
Use == V
====`)
		writeFile(t, root, `---- MODULE Root ----
CONSTANT C
I == INSTANCE Helper WITH V <- C
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		requireNoErrors(t, CheckSpec(spec))
	})

	t.Run("checks undefined USE DEF references through SANY syntax", func(t *testing.T) {
		_, diags := CheckSanySource("BadUseDef.tla", `---- MODULE BadUseDef ----
USE DEF DoesNotExist
====`)
		requireHasErrorContaining(t, diags, "undefined identifier DoesNotExist")
	})

	t.Run("rejects built-in operator redefinitions", func(t *testing.T) {
		_, diags := CheckSanySource("BuiltinRedef.tla", `---- MODULE BuiltinRedef ----
TRUE == FALSE
====`)
		requireHasErrorContaining(t, diags, "built-in")
	})

	t.Run("checks LET-local infix and postfix operator definitions", func(t *testing.T) {
		_, diags := CheckSanySource("LetSymbolDefs.tla", `---- MODULE LetSymbolDefs ----
VARIABLE v
Use == LET x + y == v'
           x^* == x
       IN (1 + 2) /\ (1^* = 1)
====`)
		requireNoErrors(t, diags)
	})

	t.Run("allows LET definition bodies to bind their own definition name", func(t *testing.T) {
		_, diags := CheckSanySource("LetChooseScope.tla", `---- MODULE LetChooseScope ----
CONSTANT S
Use == LET x == CHOOSE x \in S : TRUE
       IN x
====`)
		requireNoErrors(t, diags)
	})

	t.Run("LET zero-arity definitions preserve distinct outer operator arities", func(t *testing.T) {
		_, diags := CheckSanySource("LetShadowArity.tla", `---- MODULE LetShadowArity ----
Outer(x, y) == x
Use == LET Inner == {1}
       IN {d \in Inner : Outer(d, 1) = 1}
====`)
		requireNoErrors(t, diags)
	})

	t.Run("checks declared operator constant arity through SANY syntax", func(t *testing.T) {
		_, diags := CheckSanySource("BadOperatorConstantArity.tla", `---- MODULE BadOperatorConstantArity ----
CONSTANT F(_)
Good == F(1)
Bad == F(1, 2)
====`)
		requireHasErrorContaining(t, diags, "arity")
	})

	t.Run("checks primed constants through SANY syntax", func(t *testing.T) {
		_, diags := CheckSanySource("BadPrime.tla", `---- MODULE BadPrime ----
CONSTANT C
VARIABLE x
Bad == C' = x
====`)
		requireHasErrorContaining(t, diags, "cannot prime constant")
	})

	t.Run("checks fundamental level errors through SANY syntax", func(t *testing.T) {
		_, diags := CheckSanySource("LevelErrors.tla", `---- MODULE LevelErrors ----
VARIABLE v
BadPrime == v''
ASSUME v
====`)
		requireHasErrorContaining(t, diags, "primed")
		requireHasErrorContaining(t, diags, "constant-level")
	})

	t.Run("checks temporal level composition through SANY syntax", func(t *testing.T) {
		_, diags := CheckSanySource("LevelComposition.tla", `---- MODULE LevelComposition ----
VARIABLE v
BadAlways == [](v')
BadEventually == <>(v')
BadLeadL == (v') ~> v
BadLeadR == v -+-> (v')
BadMix == v' /\ []v
BadTemporalBound == \A x \in []v : TRUE
BadActionBoundTemporalBody == \E x \in v' : []v
====`)
		requireHasErrorContaining(t, diags, "temporal operator")
		requireHasErrorContaining(t, diags, "leads-to")
		requireHasErrorContaining(t, diags, "mix action and temporal")
		requireHasErrorContaining(t, diags, "temporal-level bound")
		requireHasErrorContaining(t, diags, "action-level bound")
	})

	t.Run("checks record constructors and field access through SANY syntax", func(t *testing.T) {
		_, diags := CheckSanySource("Records.tla", `---- MODULE Records ----
VARIABLE x
Rec == [a |-> x, b |-> 1]
Field == Rec.a
====`)
		requireNoErrors(t, diags)
	})

	t.Run("checks duplicate record constructor fields through SANY syntax", func(t *testing.T) {
		_, diags := CheckSanySource("BadRecord.tla", `---- MODULE BadRecord ----
Rec == [
  def |-> 0,
  def |-> 1
]
====`)
		requireHasErrorContaining(t, diags, "duplicate record field def")
	})

	t.Run("checks function constructors and applications through SANY syntax", func(t *testing.T) {
		_, diags := CheckSanySource("Functions.tla", `---- MODULE Functions ---- EXTENDS Naturals
VARIABLE x
Fcn == [i \in {1, 2} |-> i + x]
Val == Fcn[1]
====`)
		requireNoErrors(t, diags)
	})

	t.Run("checks function definitions through SANY syntax", func(t *testing.T) {
		_, diags := CheckSanySource("FunctionDefs.tla", `---- MODULE FunctionDefs ----
CONSTANT S
F[x \in S] == x
Val == F[1]
====`)
		requireNoErrors(t, diags)
	})

	t.Run("checks LET-local bounded function definitions through SANY syntax", func(t *testing.T) {
		_, diags := CheckSanySource("LetBoundedFunctionDefs.tla", `---- MODULE LetBoundedFunctionDefs ---- EXTENDS Naturals
Cardinality(Base) ==
  LET CS[T \in SUBSET Base] == IF T = {} THEN 0
                            ELSE 1 + CS[T \ {CHOOSE x : x \in T}]
  IN  CS[Base]
====`)
		requireNoErrors(t, diags)
	})

	t.Run("checks function application arity through SANY syntax", func(t *testing.T) {
		_, diags := CheckSanySource("BadFunctionArity.tla", `---- MODULE BadFunctionArity ----
f[x \in {}] == 0
g[x \in {}, y \in {}] == 0
BadF == f[0, 0]
BadG == g[0, 0, 0]
TupleG == g[<<0, 0>>]
====`)
		requireHasErrorContaining(t, diags, "function f arity")
		requireHasErrorContaining(t, diags, "function g arity")
	})

	t.Run("checks infix operator definitions through SANY syntax", func(t *testing.T) {
		_, diags := CheckSanySource("InfixDefs.tla", `---- MODULE InfixDefs ----
a + b == a
====`)
		requireNoErrors(t, diags)
	})

	t.Run("model checks user-defined infix operators through SANY syntax", func(t *testing.T) {
		spec := `---- MODULE InfixOperatorModel ---- EXTENDS Naturals
VARIABLE x
a \oplus b == a + b
Init == x = 1 \oplus 2
Next == x' = x
Inv == x = 3
====`
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("InfixOperatorModel.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 1 {
			t.Fatalf("result = %#v, want user-defined infix operator success", result)
		}
	})

	t.Run("checks set comprehensions through SANY syntax", func(t *testing.T) {
		_, diags := CheckSanySource("Sets.tla", `---- MODULE Sets ---- EXTENDS Naturals
CONSTANT S
Subset == {x \in S : x = x}
Mapped == {x + 1 : x \in S}
====`)
		requireNoErrors(t, diags)
	})

	t.Run("checks tuple quantifier bounds through SANY syntax", func(t *testing.T) {
		_, diags := CheckSanySource("TupleBounds.tla", `---- MODULE TupleBounds ----
CONSTANT S, T
TupleFilter == {<<x, y>> \in S : x = y}
TupleFcn == [<<x, y>> \in S, z \in T |-> x]
====`)
		requireNoErrors(t, diags)
	})

	t.Run("checks record and function set expressions through SANY syntax", func(t *testing.T) {
		_, diags := CheckSanySource("SetTypes.tla", `---- MODULE SetTypes ----
CONSTANT S, T
RSet == [a: S, b: T]
FSet == [S -> T]
Product == S \X T
====`)
		requireNoErrors(t, diags)
	})

	t.Run("checks bounded CHOOSE over compound function sets", func(t *testing.T) {
		source := `---- MODULE ChooseFunctionSetBound ----
CONSTANT S, T
Use == CHOOSE f \in [(S \X T) \cup (T \X S) -> S] : f = f
====`
		_, diags := CheckSanySource("ChooseFunctionSetBound.tla", source)
		requireNoErrors(t, diags)
	})

	t.Run("allows bounded CHOOSE variable to reuse enclosing definition name", func(t *testing.T) {
		_, diags := CheckSanySource("ChooseSelfName.tla", `---- MODULE ChooseSelfName ----
RM == CHOOSE RM : RM = RM
====`)
		requireNoErrors(t, diags)
	})

	t.Run("checks EXCEPT updates through SANY syntax", func(t *testing.T) {
		_, diags := CheckSanySource("Except.tla", `---- MODULE Except ----
CONSTANT i, v, w
VARIABLE r
Updated == [r EXCEPT !.a = v, ![i] = w]
====`)
		requireNoErrors(t, diags)
	})

	t.Run("checks labeled expressions through SANY syntax", func(t *testing.T) {
		_, diags := CheckSanySource("Labels.tla", `---- MODULE Labels ----
VARIABLE x
Named == Step:: x = x
====`)
		requireNoErrors(t, diags)
	})

	t.Run("rejects invalid labels through SANY syntax", func(t *testing.T) {
		_, diags := CheckSanySource("LabelErrors.tla", `---- MODULE LabelErrors ----
CONSTANT x
f[y \in {}] == y
Missing == \A y \in {} : missing :: TRUE
Repeated == \A y \in {} : repeated(y, y) :: TRUE
Extra(p) == extra(p) :: TRUE
Except == [f EXCEPT ![0] = badExcept :: 0]
Duplicate == {dup :: 0, dup :: 1}
ASSUME topLabel :: TRUE
====`)
		requireHasErrorContaining(t, diags, "must contain bound parameter")
		requireHasErrorContaining(t, diags, "repeated label parameter")
		requireHasErrorContaining(t, diags, "unnecessary label parameter")
		requireHasErrorContaining(t, diags, "EXCEPT")
		requireHasErrorContaining(t, diags, "Duplicate label")
		requireHasErrorContaining(t, diags, "not in definition")
	})

	t.Run("checks action and fairness wrappers through SANY syntax", func(t *testing.T) {
		_, diags := CheckSanySource("Actions.tla", `---- MODULE Actions ----
CONSTANT Servers
VARIABLE x
A == x = x
Step(s) == x = s
OtherStep(s) == x # s
Square == [A]_x
Angle == <<A>>_x
Weak == WF_<<x>>(A)
Strong == SF_<<x>>(A)
Vars == <<x>>
NamedWeak == WF_Vars(A)
NamedStrong == SF_Vars(A)
AllWeak == \A s \in Servers : WF_Vars(Step(s)) /\ WF_Vars(OtherStep(s))
Spec == /\ A
        /\ \A s \in Servers : WF_Vars(Step(s)) /\ WF_Vars(OtherStep(s))
====`)
		requireNoErrors(t, diags)
	})

	t.Run("checks fairness subscripts that are operator applications", func(t *testing.T) {
		_, diags := CheckSanySource("FairnessCallSubscripts.tla", `---- MODULE FairnessCallSubscripts ----
VARIABLE x
Vars == <<x>>
RefersTo(value, name) == value
IsLevel(value, level) == TRUE
Weak == IsLevel(WF_RefersTo(Vars, "vars")(1), TRUE)
Strong == IsLevel(SF_RefersTo(Vars, "vars")(RefersTo(x, "x")'), TRUE)
====`)
		requireNoErrors(t, diags)
	})

	t.Run("checks step expressions with primed and temporal action wrappers", func(t *testing.T) {
		_, diags := CheckSanySource("StepExpressions.tla", `---- MODULE StepExpressions ----
CONSTANT F(_, _)
VARIABLE v
Square == [v']_<<v>>
Angle == <<v'>>_<<v>>
AlwaysSquare == [][v]_<<v>>
EventuallyAngle == <><<v>>_<<v>>
UseSquare == F([v']_<<v>>, TRUE)
UseAlwaysSquare == F([][v]_<<v>>, TRUE)
====`)
		requireNoErrors(t, diags)
	})

	t.Run("checks open expression binders through SANY syntax", func(t *testing.T) {
		_, diags := CheckSanySource("OpenExprs.tla", `---- MODULE OpenExprs ----
CONSTANT S
VARIABLE x, y
C == CASE x = x -> y [] OTHER -> x
Ch == CHOOSE z \in S : z = z
Q == \A z : z = z
F(op(_,_)) == op(1,1)
Lam == F(LAMBDA a, b : a = b)
====`)
		requireNoErrors(t, diags)
	})

	t.Run("checks CASE arms with action and temporal definitions", func(t *testing.T) {
		_, diags := CheckSanySource("CaseLevels.tla", `---- MODULE CaseLevels ----
CONSTANT c
VARIABLE v
action == v'
temporal == []c
ActionCase == CASE action -> 1
TemporalCase == CASE temporal -> 1
MixedCase == CASE c = c -> v = v [] action -> temporal
====`)
		requireNoErrors(t, diags)
	})

	t.Run("checks real literals through SANY syntax", func(t *testing.T) {
		_, diags := CheckSanySource("Reals.tla", `---- MODULE Reals ----
R == 1.25
====`)
		requireNoErrors(t, diags)
	})

	t.Run("checks named assumptions as referable definitions", func(t *testing.T) {
		_, diags := CheckSanySource("NamedAssumptions.tla", `---- MODULE NamedAssumptions ----
ASSUME A == TRUE
ASSUMPTION B == A
AXIOM C == B
Use == C
====`)
		requireNoErrors(t, diags)
	})

	t.Run("checks operator formals passed through higher-order calls", func(t *testing.T) {
		_, diags := CheckSanySource("HigherOrderOperatorFormals.tla", `---- MODULE HigherOrderOperatorFormals ----
Apply2(Op(_,_), x, y) == Op(x, y)
Forward(Op(_,_), x, y) == Apply2(Op, x, y)
====`)
		requireNoErrors(t, diags)
	})

	t.Run("checks mixfix operator formals contribute to definition arity", func(t *testing.T) {
		_, diags := CheckSanySource("MixfixOperatorFormalArity.tla", `---- MODULE MixfixOperatorFormalArity ----
OpToRel(_\prec_, S) == TRUE
Use == OpToRel(=, {})
====`)
		requireNoErrors(t, diags)
	})

	t.Run("checks prefix negative operator definitions at body boundaries", func(t *testing.T) {
		_, diags := CheckSanySource("NegativeOps.tla", `---- MODULE NegativeOps ----
EXTENDS Naturals
CONSTANT Node, N
RefersTo(value, name) == TRUE
op(-._) == RefersTo(-1, "op!-.")
CONSTANT -._
ASSUME RefersTo(-1, "constant_negative")
-. x == x
ASSUME RefersTo(-1, "-.")
Range == [Node -> -(N-1)..(N-1)]
====`)
		requireNoErrors(t, diags)
	})

	t.Run("checks RECURSIVE declarations through SANY syntax", func(t *testing.T) {
		_, diags := CheckSanySource("RecursiveDefs.tla", `---- MODULE RecursiveDefs ----
RECURSIVE Fact(_), Missing(_), Bad(_, _)
Fact(n) == IF n = 0 THEN 1 ELSE n * Fact(n - 1)
Bad(n) == n
====`)
		requireHasErrorContaining(t, diags, "recursive")
		requireHasErrorContaining(t, diags, "Missing")
		requireHasErrorContaining(t, diags, "arity")
	})

	t.Run("rejects primed recursive arguments through SANY syntax", func(t *testing.T) {
		_, diags := CheckSanySource("RecursivePrime.tla", `---- MODULE RecursivePrime ----
RECURSIVE op(_)
op(x) == x'
====`)
		requireHasErrorContaining(t, diags, "Argument 1 of recursive operator op is primed")
		requireHasErrorContaining(t, diags, "prime")
	})

	t.Run("rejects invalid recursive definition sections through SANY syntax", func(t *testing.T) {
		_, diags := CheckSanySource("RecursiveSections.tla", `---- MODULE RecursiveSections ----
RECURSIVE top
CONSTANT c
top == 0
Outer ==
  LET
    RECURSIVE recDef
    def ==
      LET recDef == 0
      IN 0
  IN 0
====`)
		requireHasErrorContaining(t, diags, "recursive definition section")
		requireHasErrorContaining(t, diags, "recursive declaration recDef is defined in the wrong LET/IN level")
	})

	t.Run("checks ASSUME PROVE theorem bodies through SANY syntax", func(t *testing.T) {
		_, diags := CheckSanySource("AssumeProve.tla", `---- MODULE AssumeProve ----
CONSTANT A, B
THEOREM Good == ASSUME A = A PROVE B = B
THEOREM Bad == ASSUME A = A PROVE Missing = Missing
====`)
		requireHasErrorContaining(t, diags, "undefined")
	})

	t.Run("ignores terminal proof nodes while checking theorem bodies", func(t *testing.T) {
		_, diags := CheckSanySource("ProofTheorem.tla", `---- MODULE ProofTheorem ----
CONSTANT A, B
THEOREM Good == ASSUME A = A PROVE B = B
PROOF BY Good
====`)
		requireNoErrors(t, diags)
	})

	t.Run("allows instantiated theorem statement references", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Base.tla"), `---- MODULE Base ----
THEOREM NextDef == TRUE
====`)
		root := filepath.Join(dir, "Root.tla")
		writeFile(t, root, `---- MODULE Root ----
P == INSTANCE Base
LEMMA PNextDef == P!NextDef!:
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		requireNoErrors(t, CheckSpec(spec))
	})

	t.Run("rejects invalid proof semantics through SANY syntax", func(t *testing.T) {
		_, diags := CheckSanySource("ProofErrors.tla", `---- MODULE ProofErrors ----
VARIABLE v
THEOREM []v
PROOF
<+>bad. QED

THEOREM []v
PROOF
<1> HAVE v
<1> TAKE x \in v
<1> WITNESS v
<1> CASE v
<1> PICK y \in v : []v
<1> QED

THEOREM thm == ASSUME TRUE PROVE TRUE
op == thm
def == TRUE
HIDE def
====`)
		requireHasErrorContaining(t, diags, "implicit proof step")
		// Java reports the three generation errors in this fixture and never
		// calls ModuleNode.levelCheck, so temporal proof errors are absent.
		if len(diags) != 3 {
			t.Fatalf("generation diagnostics=%d, want 3; diagnostics:\n%s", len(diags), diags)
		}
		for _, diagnostic := range diags {
			if diagnostic.Code == "E4352" || diagnostic.Code == "E4353" || diagnostic.Code == "E4354" {
				t.Fatalf("level checking ran after generation errors: %s", diagnostic)
			}
		}
		requireHasErrorContaining(t, diags, "ASSUME/PROVE")
		requireHasErrorContaining(t, diags, "HIDE")
	})

	t.Run("does not apply temporal theorem proof restrictions to nested proofs", func(t *testing.T) {
		_, diags := CheckSanySource("NestedProofLevels.tla", `---- MODULE NestedProofLevels ----
VARIABLE v
THEOREM []v
<1>1. TRUE
  <2>1. CASE v
    BY <2>1
  <2>. QED BY <2>1
<1>. QED BY <1>1
====`)
		requireNoErrors(t, diags)
	})

	t.Run("allows proof step names to collide with scoped bound identifiers", func(t *testing.T) {
		// PICK keeps I in scope after its proof. Java rejects another binder I;
		// keep the named <1>I step collision and use a fresh quantifier variable.
		_, diags := CheckSanySource("ProofStepBoundIdentifierCollision.tla", `---- MODULE ProofStepBoundIdentifierCollision ----
EXTENDS Naturals
U == {1}
THEOREM TRUE
PROOF
<1>1. TRUE
<1>2. PICK I \in U : I \in U
<1>3. CASE I \in U
  BY <1>2
<1>I. PICK j : j = j
<1>4. CASE \A k \in U : k = k
  BY <1>2
<1>. QED
====`)
		requireNoErrors(t, diags)
	})

	t.Run("checks higher-order operator references through SANY syntax", func(t *testing.T) {
		_, diags := CheckSanySource("OperatorRefs.tla", `---- MODULE OperatorRefs ----
EXTENDS Naturals, Sequences
Apply(op(_, _), x, y) == op(x, y)
Add(x, y) == x + y
Sum == Apply(+, 1, 2)
NamedSum == Apply(Add, 1, 2)
Meet == Apply(\intersect, {1}, {1})
Compose == \o(1, 2)
====`)
		requireNoErrors(t, diags)
	})

	t.Run("checks named theorem keyword synonyms as referable definitions", func(t *testing.T) {
		_, diags := CheckSanySource("NamedTheorems.tla", `---- MODULE NamedTheorems ----
THEOREM T == TRUE
PROPOSITION P == T
LEMMA L == P
COROLLARY C == L
Use == C
====`)
		requireNoErrors(t, diags)
	})

	t.Run("treats ASSUME PROVE NEW declarations as theorem locals", func(t *testing.T) {
		_, diags := CheckSanySource("NewAssumeProve.tla", `---- MODULE NewAssumeProve ----
CONSTANT S
THEOREM Good == ASSUME NEW x \in S, NEW CONSTANT C PROVE x = x /\ C = C
THEOREM Bad == ASSUME NEW y \in S PROVE Missing = y
====`)
		requireHasErrorContaining(t, diags, "Missing")
		for _, local := range []string{"x", "C", "y"} {
			if strings.Contains(diags.Error(), "undefined identifier "+local) {
				t.Fatalf("diagnostics treated NEW local %s as undefined:\n%s", local, diags.Error())
			}
		}
	})

	t.Run("keeps ASSUME PROVE NEW symbols scoped after earlier assumptions", func(t *testing.T) {
		_, diags := CheckSanySource("AssumeProveSequentialNew.tla", `---- MODULE AssumeProveSequentialNew ----
CONSTANT S
THEOREM Good ==
  ASSUME \A s \in S : s = s,
         NEW s \in S,
         s = s
  PROVE  s = s
====`)
		requireNoErrors(t, diags)
	})

	t.Run("model checks a finite counter through SANY syntax", func(t *testing.T) {
		result, diags := ModelCheckSanySource("Counter.tla", counterSpec("x <= 3"), counterCfg(), ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 4 {
			t.Fatalf("result = %#v, want success with 4 states", result)
		}
	})

	t.Run("model checks nondeterministic IF specs through SANY syntax", func(t *testing.T) {
		spec := `---- MODULE Nondet ---- EXTENDS Naturals
VARIABLE x
Init == x \in {0, 1}
Next == (x < 2 /\ x' \in {x, x + 1}) \/ (x = 2 /\ x' = x)
Inv == IF x <= 2 THEN TRUE ELSE FALSE
====`
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("Nondet.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 3 {
			t.Fatalf("result = %#v, want success with 3 states", result)
		}
	})

	t.Run("model checks UNCHANGED tuple actions through SANY syntax", func(t *testing.T) {
		spec := `---- MODULE TwoVars ---- EXTENDS Naturals
VARIABLE x, y
Init == x = 0 /\ y = 7
Next == (x < 2 /\ x' = x + 1 /\ UNCHANGED y) \/ (x = 2 /\ UNCHANGED <<x, y>>)
Inv == y = 7
====`
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("TwoVars.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 3 {
			t.Fatalf("result = %#v, want success with 3 states", result)
		}
	})

	t.Run("model checks LET IN specs through SANY syntax", func(t *testing.T) {
		spec := `---- MODULE LetCounter ---- EXTENDS Naturals
VARIABLE x
Init == x = 0
Next == LET Inc == x + 1 IN (x < 3 /\ x' = Inc) \/ (x = 3 /\ x' = x)
Inv == LET Bound == 3 IN x <= Bound
====`
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("LetCounter.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 4 {
			t.Fatalf("result = %#v, want success with 4 states", result)
		}
	})

	t.Run("model checks finite quantifiers through SANY syntax", func(t *testing.T) {
		spec := `---- MODULE Quantified ---- EXTENDS Naturals
VARIABLE x
Init == x = 0
Next == (x < 2 /\ x' = x + 1) \/ (x = 2 /\ x' = x)
Inv == \A n \in {0, 1, 2}: n <= 2 /\ \E m \in {x, x + 1}: m >= x
====`
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("Quantified.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 3 {
			t.Fatalf("result = %#v, want success with 3 states", result)
		}
	})

	t.Run("model checks multi-variable bounded quantifiers through SANY syntax", func(t *testing.T) {
		spec := `---- MODULE MultiQuantified ---- EXTENDS Naturals
VARIABLE x
Init == x = 0
Next == x' = x
Inv == \A a, b \in {0, 1}: a <= 1 /\ b <= 1
====`
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("MultiQuantified.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 1 {
			t.Fatalf("result = %#v, want success with multi-variable quantifier", result)
		}
	})

	t.Run("model checks multi-domain bounded quantifiers through SANY syntax", func(t *testing.T) {
		spec := `---- MODULE MultiDomainQuantified ---- EXTENDS Naturals
VARIABLE x
Init == x = 0
Next == x' = x
Inv == \A a \in {0, 1}, b \in {1, 2}: a < b + 1
====`
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("MultiDomainQuantified.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 1 {
			t.Fatalf("result = %#v, want success with multi-domain quantifier", result)
		}
	})

	t.Run("model checks existential action assignments through SANY syntax", func(t *testing.T) {
		spec := `---- MODULE ExistentialActionModel ---- EXTENDS Naturals
VARIABLE x
Init == x = 0
Next == (x < 2 /\ (\E n \in {x, x + 1}: x' = n)) \/ (x = 2 /\ x' = x)
Inv == x <= 2
====`
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("ExistentialActionModel.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 3 {
			t.Fatalf("result = %#v, want success with existential action assignments", result)
		}
	})

	t.Run("model checks CASE and labels through SANY syntax", func(t *testing.T) {
		spec := `---- MODULE CaseModel ---- EXTENDS Naturals
VARIABLE x
Init == x = 0
Next == x' = CASE x < 2 -> x + 1 [] OTHER -> x
Inv == Bound:: CASE x <= 2 -> TRUE [] OTHER -> FALSE
====`
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("CaseModel.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 3 {
			t.Fatalf("result = %#v, want success with 3 states", result)
		}
	})

	t.Run("model checks finite CHOOSE through SANY syntax", func(t *testing.T) {
		spec := `---- MODULE ChooseModel ----
VARIABLE x
Init == x = CHOOSE n \in {0, 1, 2}: n = 1
Next == x' = x
Inv == x = 1
====`
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("ChooseModel.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 1 {
			t.Fatalf("result = %#v, want success with one chosen state", result)
		}
	})

	t.Run("model checks finite set comprehensions through SANY syntax", func(t *testing.T) {
		spec := `---- MODULE ComprehensionModel ---- EXTENDS Naturals
VARIABLE x
Init == x \in {n + 1 : n \in 0..2}
Next == x' = x
Inv == x \in {y \in 0..3 : y > 0}
====`
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("ComprehensionModel.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 3 {
			t.Fatalf("result = %#v, want success with 3 comprehension states", result)
		}
	})

	t.Run("model checks finite function applications through SANY syntax", func(t *testing.T) {
		spec := `---- MODULE FunctionModel ---- EXTENDS Naturals
VARIABLE x
F[i \in {1, 2}] == i + 1
Init == x = F[1]
Next == x' = x
Inv == x = 2
====`
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("FunctionModel.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 1 {
			t.Fatalf("result = %#v, want success with one function state", result)
		}
	})

	t.Run("model checks record field access through SANY syntax", func(t *testing.T) {
		spec := `---- MODULE RecordModel ----
VARIABLE x
Rec == [a |-> 1, b |-> 2]
Init == x = Rec.a
Next == x' = x
Inv == x = 1
====`
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("RecordModel.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 1 {
			t.Fatalf("result = %#v, want success with one record-derived state", result)
		}
	})

	t.Run("model checks record EXCEPT updates through SANY syntax", func(t *testing.T) {
		spec := `---- MODULE RecordExceptModel ---- EXTENDS Naturals
VARIABLE x
Rec == [a |-> x, b |-> 2]
Updated == [Rec EXCEPT !.a = @ + 1]
Init == x = 0
Next == (x < 2 /\ x' = Updated.a) \/ (x = 2 /\ x' = x)
Inv == x <= 2
====`
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("RecordExceptModel.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 3 {
			t.Fatalf("result = %#v, want success with three record-updated states", result)
		}
	})

	t.Run("model checks function EXCEPT updates through SANY syntax", func(t *testing.T) {
		spec := `---- MODULE FunctionExceptModel ---- EXTENDS Naturals
VARIABLE x
F[i \in 0..2] == i
Updated == [F EXCEPT ![x] = @ + 1]
Init == x = 0
Next == (x < 2 /\ x' = Updated[x]) \/ (x = 2 /\ x' = x)
Inv == x <= 2
====`
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("FunctionExceptModel.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 3 {
			t.Fatalf("result = %#v, want success with three function-updated states", result)
		}
	})

	t.Run("model checks DOMAIN of finite functions through SANY syntax", func(t *testing.T) {
		spec := `---- MODULE FunctionDomainModel ---- EXTENDS Naturals
VARIABLE x
F[i \in 0..2] == i
Init == x \in DOMAIN F
Next == x' = x
Inv == x \in {0, 1, 2}
====`
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("FunctionDomainModel.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 3 {
			t.Fatalf("result = %#v, want success with three domain states", result)
		}
	})

	t.Run("model checks FiniteSets Cardinality and IsFiniteSet from standard modules", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "CardinalityModel.tla")
		writeFile(t, root, `---- MODULE CardinalityModel ----
EXTENDS FiniteSets
VARIABLE x
Init == /\ x = Cardinality({0, 1, 1})
        /\ IsFiniteSet({0, 1})
Next == x' = x
Inv == x = 2
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
		cfg, cfgDiags := ParseConfigSource(filepath.Join(dir, "MC.cfg"), "INIT Init\nNEXT Next\nINVARIANT Inv\n")
		requireNoErrors(t, cfgDiags)
		result, runDiags := ModelCheck(spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, runDiags)
		if !result.OK || result.StatesExplored != 1 {
			t.Fatalf("result = %#v, want cardinality-derived one-state success", result)
		}
	})

	t.Run("checks primitive standard sets from built-ins and standard modules", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "StandardSets.tla")
		writeFile(t, root, `---- MODULE StandardSets ----
EXTENDS Reals
R == [ok : BOOLEAN, s : STRING, i : Int, r : Real]
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
	})

	t.Run("model checks FiniteSets Cardinality over string sets", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "StringCardinalityModel.tla")
		writeFile(t, root, `---- MODULE StringCardinalityModel ----
EXTENDS FiniteSets
VARIABLE x
Init == x = Cardinality({"ready", "done", "ready"})
Next == x' = x
Inv == x = 2
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
		cfg, cfgDiags := ParseConfigSource(filepath.Join(dir, "MC.cfg"), "INIT Init\nNEXT Next\nINVARIANT Inv\n")
		requireNoErrors(t, cfgDiags)
		result, runDiags := ModelCheck(spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, runDiags)
		if !result.OK || result.StatesExplored != 1 {
			t.Fatalf("result = %#v, want string cardinality-derived one-state success", result)
		}
	})

	t.Run("model checks FiniteSets Cardinality over powersets", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "PowersetCardinalityModel.tla")
		writeFile(t, root, `---- MODULE PowersetCardinalityModel ----
EXTENDS FiniteSets
VARIABLE x
Init == x = Cardinality(SUBSET {0, 1})
Next == x' = Cardinality(SUBSET {"ready", "done"})
Inv == x = 4
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
		cfg, cfgDiags := ParseConfigSource(filepath.Join(dir, "MC.cfg"), "INIT Init\nNEXT Next\nINVARIANT Inv\n")
		requireNoErrors(t, cfgDiags)
		result, runDiags := ModelCheck(spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, runDiags)
		if !result.OK || result.StatesExplored != 1 {
			t.Fatalf("result = %#v, want powerset cardinality one-state success", result)
		}
	})

	t.Run("model checks FiniteSets Cardinality over set-of-set literals", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "SetOfSetsCardinalityModel.tla")
		writeFile(t, root, `---- MODULE SetOfSetsCardinalityModel ----
EXTENDS FiniteSets
VARIABLE x
Init == x = Cardinality({{0}, {1}, {0}})
Next == x' = Cardinality({{"ready"}, {"done"}, {"ready"}})
Inv == x = 2
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
		cfg, cfgDiags := ParseConfigSource(filepath.Join(dir, "MC.cfg"), "INIT Init\nNEXT Next\nINVARIANT Inv\n")
		requireNoErrors(t, cfgDiags)
		result, runDiags := ModelCheck(spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, runDiags)
		if !result.OK || result.StatesExplored != 1 {
			t.Fatalf("result = %#v, want set-of-set cardinality one-state success", result)
		}
	})

	t.Run("model checks Sequences Len from standard modules", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "LenModel.tla")
		writeFile(t, root, `---- MODULE LenModel ----
EXTENDS Sequences
VARIABLE x
Init == x = Len(<<1, 2, 3>>)
Next == x' = x
Inv == x = 3
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
		cfg, cfgDiags := ParseConfigSource(filepath.Join(dir, "MC.cfg"), "INIT Init\nNEXT Next\nINVARIANT Inv\n")
		requireNoErrors(t, cfgDiags)
		result, runDiags := ModelCheck(spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, runDiags)
		if !result.OK || result.StatesExplored != 1 {
			t.Fatalf("result = %#v, want Len-derived one-state success", result)
		}
	})

	t.Run("model checks finite sequence Head Tail and Append", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "SequenceOpsModel.tla")
		writeFile(t, root, `---- MODULE SequenceOpsModel ----
EXTENDS Naturals, Sequences
VARIABLE x
S == Append(<<1, 2>>, 3)
Init == x = Head(S) + Len(Tail(S))
Next == x' = x
Inv == x = 3
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
		cfg, cfgDiags := ParseConfigSource(filepath.Join(dir, "MC.cfg"), "INIT Init\nNEXT Next\nINVARIANT Inv\n")
		requireNoErrors(t, cfgDiags)
		result, runDiags := ModelCheck(spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, runDiags)
		if !result.OK || result.StatesExplored != 1 {
			t.Fatalf("result = %#v, want sequence operation success", result)
		}
	})

	t.Run("model checks finite sequence indexing", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "SequenceIndexModel.tla")
		writeFile(t, root, `---- MODULE SequenceIndexModel ----
EXTENDS Naturals, Sequences
VARIABLE x
S == Append(<<1, 2>>, 3)
Init == x = S[2] + S[3]
Next == x' = x
Inv == x = 5
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
		cfg, cfgDiags := ParseConfigSource(filepath.Join(dir, "MC.cfg"), "INIT Init\nNEXT Next\nINVARIANT Inv\n")
		requireNoErrors(t, cfgDiags)
		result, runDiags := ModelCheck(spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, runDiags)
		if !result.OK || result.StatesExplored != 1 {
			t.Fatalf("result = %#v, want sequence index success", result)
		}
	})

	t.Run("model checks finite sequence domain and subsequence", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "SequenceDomainModel.tla")
		writeFile(t, root, `---- MODULE SequenceDomainModel ----
EXTENDS Sequences
VARIABLE x
S == SubSeq(<<4, 5, 6>>, 2, 3)
Init == x \in DOMAIN S
Next == x' = x
Inv == x \in {1, 2} /\ S[1] = 5 /\ S[2] = 6
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
		cfg, cfgDiags := ParseConfigSource(filepath.Join(dir, "MC.cfg"), "INIT Init\nNEXT Next\nINVARIANT Inv\n")
		requireNoErrors(t, cfgDiags)
		result, runDiags := ModelCheck(spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, runDiags)
		if !result.OK || result.StatesExplored != 2 {
			t.Fatalf("result = %#v, want sequence domain success with two states", result)
		}
	})

	t.Run("model checks finite sequence selection", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "SequenceSelectModel.tla")
		writeFile(t, root, `---- MODULE SequenceSelectModel ----
EXTENDS Naturals, Sequences
VARIABLE x
S == SelectSeq(<<1, 2, 3>>, LAMBDA n : n > 1)
Init == x = Len(S) + Head(S)
Next == x' = x
Inv == x = 4
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
		cfg, cfgDiags := ParseConfigSource(filepath.Join(dir, "MC.cfg"), "INIT Init\nNEXT Next\nINVARIANT Inv\n")
		requireNoErrors(t, cfgDiags)
		result, runDiags := ModelCheck(spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, runDiags)
		if !result.OK || result.StatesExplored != 1 {
			t.Fatalf("result = %#v, want sequence selection success", result)
		}
	})

	t.Run("model checks finite sequence equality", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "SequenceEqualityModel.tla")
		writeFile(t, root, `---- MODULE SequenceEqualityModel ----
EXTENDS Sequences
VARIABLE x
S == SubSeq(<<0, 1, 2>>, 2, 3)
Init == x = 0
Next == x' = x
Inv == S = <<1, 2>> /\ S /= <<2, 1>>
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
		cfg, cfgDiags := ParseConfigSource(filepath.Join(dir, "MC.cfg"), "INIT Init\nNEXT Next\nINVARIANT Inv\n")
		requireNoErrors(t, cfgDiags)
		result, runDiags := ModelCheck(spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, runDiags)
		if !result.OK || result.StatesExplored != 1 {
			t.Fatalf("result = %#v, want sequence equality success", result)
		}
	})

	t.Run("model checks TLC Assert from standard modules", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "AssertModel.tla")
		writeFile(t, root, `---- MODULE AssertModel ----
EXTENDS TLC
VARIABLE x
Init == x = 0
Next == x' = x
Inv == Assert(x = 0, "x changed")
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
		cfg, cfgDiags := ParseConfigSource(filepath.Join(dir, "MC.cfg"), "INIT Init\nNEXT Next\nINVARIANT Inv\n")
		requireNoErrors(t, cfgDiags)
		result, runDiags := ModelCheck(spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, runDiags)
		if !result.OK || result.StatesExplored != 1 {
			t.Fatalf("result = %#v, want TLC Assert success", result)
		}
	})

	t.Run("reports TLC Assert string messages", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "AssertMessageModel.tla")
		writeFile(t, root, `---- MODULE AssertMessageModel ----
EXTENDS TLC
VARIABLE x
Init == x = 1
Next == x' = x
Inv == Assert(x = 0, "x changed")
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
		cfg, cfgDiags := ParseConfigSource(filepath.Join(dir, "MC.cfg"), "INIT Init\nNEXT Next\nINVARIANT Inv\n")
		requireNoErrors(t, cfgDiags)
		_, runDiags := ModelCheck(spec, cfg, ModelCheckOptions{})
		requireHasErrorContaining(t, runDiags, "x changed")
	})

	t.Run("model checks TLC Print returns its value argument", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "PrintModel.tla")
		writeFile(t, root, `---- MODULE PrintModel ----
EXTENDS TLC
VARIABLE x
Init == x = 1
Next == x' = x
Inv == Print("checking x", x = 0)
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
		cfg, cfgDiags := ParseConfigSource(filepath.Join(dir, "MC.cfg"), "INIT Init\nNEXT Next\nINVARIANT Inv\n")
		requireNoErrors(t, cfgDiags)
		result, runDiags := ModelCheck(spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, runDiags)
		if result.OK || !strings.Contains(result.Error, "invariant Inv") {
			t.Fatalf("result = %#v, want TLC Print value to drive invariant failure", result)
		}
	})

	t.Run("model checks finite integer set union and intersection", func(t *testing.T) {
		spec := `---- MODULE SetAlgebraModel ----
VARIABLE x
Init == x \in (({0, 1} \cup {1, 2}) \cap {1, 2})
Next == x' = x
Inv == x \in {1, 2}
====`
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("SetAlgebraModel.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 2 {
			t.Fatalf("result = %#v, want success with two set-algebra states", result)
		}
	})

	t.Run("model checks finite integer set predicates", func(t *testing.T) {
		spec := `---- MODULE SetPredicateModel ----
VARIABLE x
Init == x \in {0, 1}
Next == x' = x
Inv == {x} \subseteq {0, 1} /\ {0, 1} = ({0} \cup {1}) /\ x \notin {2}
====`
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("SetPredicateModel.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 2 {
			t.Fatalf("result = %#v, want success with two set-predicate states", result)
		}
	})

	t.Run("model checks TLA integer division and remainder operators", func(t *testing.T) {
		spec := `---- MODULE IntegerArithmeticModel ----
EXTENDS Integers
VARIABLE x
Init == x = (5 \div 2) + (5 % 2)
Next == x' = x
Inv == x = 3
====`
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("IntegerArithmeticModel.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 1 {
			t.Fatalf("result = %#v, want integer arithmetic success", result)
		}
	})

	t.Run("model checks boolean equivalence operators", func(t *testing.T) {
		spec := `---- MODULE EquivalenceModel ----
VARIABLE x
Init == x \in {0, 1}
Next == x' = x
Inv == ((x = 0) <=> (x \notin {1})) /\ ((x = 1) \equiv (x \in {1}))
====`
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("EquivalenceModel.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 2 {
			t.Fatalf("result = %#v, want equivalence success for two states", result)
		}
	})

	t.Run("model checks square action wrappers through SANY syntax", func(t *testing.T) {
		spec := `---- MODULE ActionWrapperModel ---- EXTENDS Naturals
VARIABLE x
Init == x = 0
Step == x < 2 /\ x' = x + 1
Next == [Step]_x
Inv == x <= 2
====`
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("ActionWrapperModel.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 3 {
			t.Fatalf("result = %#v, want success with 3 states", result)
		}
	})

	t.Run("model checks SPECIFICATION cfg entries through SANY syntax", func(t *testing.T) {
		spec := `---- MODULE SpecificationModel ---- EXTENDS Naturals
VARIABLE x
Init == x = 0
Next == (x < 2 /\ x' = x + 1) \/ (x = 2 /\ x' = x)
Spec == Init /\ [][Next]_x
Inv == x <= 2
====`
		cfg := `SPECIFICATION Spec
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("SpecificationModel.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 3 {
			t.Fatalf("result = %#v, want success with 3 states", result)
		}
	})

	t.Run("model checks integer intervals through SANY syntax", func(t *testing.T) {
		spec := `---- MODULE Intervals ---- EXTENDS Naturals
CONSTANT N
VARIABLE x
Init == x \in 0..N
Next == x' = x
Inv == \A n \in 0..N: n <= N
====`
		cfg := `CONSTANT N = 3
INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("Intervals.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 4 {
			t.Fatalf("result = %#v, want success with 4 interval states", result)
		}
	})

	t.Run("loads sibling modules for SANY model checking", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "Root.tla")
		writeFile(t, filepath.Join(dir, "Helper.tla"), `---- MODULE Helper ---- EXTENDS Naturals
Zero == 0
One == Zero + 1
====`)
		writeFile(t, root, `---- MODULE Root ----
EXTENDS Helper
Helper == INSTANCE Helper
VARIABLE x
Init == x = Helper!Zero
Next == (x < Helper!One /\ x' = x + Helper!One) \/ (x = Helper!One /\ x' = x)
Inv == x <= Helper!One
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		cfg, cfgDiags := ParseConfigSource(filepath.Join(dir, "MC.cfg"), "INIT Init\nNEXT Next\nINVARIANT Inv\n")
		requireNoErrors(t, cfgDiags)
		result, runDiags := ModelCheck(spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, runDiags)
		if !result.OK || result.StatesExplored != 2 {
			t.Fatalf("result = %#v, want success with 2 states", result)
		}
	})

	t.Run("loads plain INSTANCE modules for SANY model checking", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "Root.tla")
		writeFile(t, filepath.Join(dir, "Helper.tla"), `---- MODULE Helper ---- EXTENDS Naturals
Zero == 0
Inc(n) == n + 1
====`)
		writeFile(t, root, `---- MODULE Root ----
INSTANCE Helper
VARIABLE x
Init == x = Zero
Next == (x < Inc(Zero) /\ x' = Inc(x)) \/ (x = Inc(Zero) /\ x' = x)
Inv == x <= Inc(Zero)
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		cfg, cfgDiags := ParseConfigSource(filepath.Join(dir, "MC.cfg"), "INIT Init\nNEXT Next\nINVARIANT Inv\n")
		requireNoErrors(t, cfgDiags)
		result, runDiags := ModelCheck(spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, runDiags)
		if !result.OK || result.StatesExplored != 2 {
			t.Fatalf("result = %#v, want success with 2 instance states", result)
		}
	})

	t.Run("applies INSTANCE WITH substitutions for SANY model checking", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "Root.tla")
		writeFile(t, filepath.Join(dir, "Helper.tla"), `---- MODULE Helper ---- EXTENDS Naturals
CONSTANT C
Start == C
Inc(n) == n + C
====`)
		writeFile(t, root, `---- MODULE Root ----
INSTANCE Helper WITH C <- 1
VARIABLE x
Init == x = Start
Next == (x < 3 /\ x' = Inc(x)) \/ (x = 3 /\ x' = x)
Inv == x <= 3
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		cfg, cfgDiags := ParseConfigSource(filepath.Join(dir, "MC.cfg"), "INIT Init\nNEXT Next\nINVARIANT Inv\n")
		requireNoErrors(t, cfgDiags)
		result, runDiags := ModelCheck(spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, runDiags)
		if !result.OK || result.StatesExplored != 3 {
			t.Fatalf("result = %#v, want success with substituted instance", result)
		}
	})

	t.Run("checks named INSTANCE definitions with nested modules", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "Root.tla")
		writeFile(t, root, `---- MODULE Root ----
VARIABLE v
---- MODULE Inner ----
CONSTANT c
Op == c
====
constantLevelImport == INSTANCE Inner WITH c <- 0
variableLevelImport == INSTANCE Inner WITH c <- v
Use == variableLevelImport!Op
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		if spec.Modules["Inner"] == nil {
			t.Fatalf("nested module Inner was not registered: %#v", spec.Modules)
		}
		requireNoErrors(t, CheckSpec(spec))
	})

	t.Run("checks nested modules with enclosing module context", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "Root.tla")
		writeFile(t, filepath.Join(dir, "Helper.tla"), `---- MODULE Helper ----
VARIABLE h
Init == h = h
====`)
		writeFile(t, root, `---- MODULE Root ----
EXTENDS Naturals
VARIABLE x
ParentDef == x = x
H == INSTANCE Helper WITH h <- x
---- MODULE Inner ----
Use == H!Init /\ ParentDef /\ 1 \in Nat
====
I == INSTANCE Inner
RootUse == I!Use
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		requireNoErrors(t, CheckSpec(spec))
	})

	t.Run("checks INSTANCE WITH substitution arity", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "Root.tla")
		writeFile(t, filepath.Join(dir, "Helper.tla"), `---- MODULE Helper ---- EXTENDS Naturals
CONSTANT C, F(_)
Start == C
Inc(n) == n + 1
UseF == F(1)
====`)
		writeFile(t, root, `---- MODULE Root ----
INSTANCE Helper WITH C <- 1, F <- 1
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireHasErrorContaining(t, sem, "substitution")
		requireHasErrorContaining(t, sem, "arity")
	})

	t.Run("allows LAMBDA substitutions for operator parameters", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "Root.tla")
		writeFile(t, filepath.Join(dir, "Helper.tla"), `---- MODULE Helper ----
CONSTANT Constraint(_, _)
Use == Constraint(1, 2)
====`)
		writeFile(t, root, `---- MODULE Root ----
INSTANCE Helper WITH Constraint <- LAMBDA s, l : TRUE
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		requireNoErrors(t, CheckSpec(spec))
	})

	t.Run("checks INSTANCE substitution targets and required assignments", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "Root.tla")
		writeFile(t, root, `---- MODULE Root ----
---- MODULE Missing ----
CONSTANT c
====
INSTANCE Missing

---- MODULE Duplicate ----
CONSTANT c
====
INSTANCE Duplicate WITH c <- 0, c <- 1

---- MODULE Illegal ----
c == TRUE
====
INSTANCE Illegal WITH c <- 0
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		if len(sem) != 3 || sem[0].Code != "E4240" {
			t.Fatalf("want three INSTANCE errors starting with 4240: %v", sem)
		}
		message := "Substitution missing for symbol c declared at line 3, col 10 to line 3, col 10 of module Root \nand instantiated in module Root."
		if sem[0].SANYMessage != message {
			t.Fatalf("missing substitution message=%q, want %q", sem[0].SANYMessage, message)
		}
		location := sem[0].SANYRange
		if location.Begin.Line != 5 || location.Begin.Column != 1 || location.End.Line != 5 || location.End.Column != 16 {
			t.Fatalf("missing substitution range=%v, want 5:1 to 5:16", location)
		}
		requireHasErrorContaining(t, sem, "duplicate INSTANCE substitution")
		requireHasErrorContaining(t, sem, "not a CONSTANT or VARIABLE")
	})

	t.Run("allows INSTANCE substitutions for inherited constants", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Parent.tla"), `---- MODULE Parent ----
CONSTANT C
====`)
		writeFile(t, filepath.Join(dir, "Child.tla"), `---- MODULE Child ----
EXTENDS Parent
====`)
		root := filepath.Join(dir, "Root.tla")
		writeFile(t, root, `---- MODULE Root ----
CONSTANT C
INSTANCE Child WITH C <- C
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		requireNoErrors(t, CheckSpec(spec))
	})

	t.Run("allows implicit INSTANCE substitutions from transitive EXTENDS", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Interface.tla"), `---- MODULE Interface ----
VARIABLE v
CONSTANT C, Op(_)
====`)
		writeFile(t, filepath.Join(dir, "Provider.tla"), `---- MODULE Provider ----
EXTENDS Interface
====`)
		writeFile(t, filepath.Join(dir, "Target.tla"), `---- MODULE Target ----
EXTENDS Interface
CONSTANT R
Use == Op(C) /\ v = v /\ R = R
====`)
		root := filepath.Join(dir, "Root.tla")
		writeFile(t, root, `---- MODULE Root ----
EXTENDS Provider
CONSTANT R
I == INSTANCE Target
Use == I!Use
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		requireNoErrors(t, CheckSpec(spec))
	})

	t.Run("checks INSTANCE substitution levels for non-constant modules", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "Root.tla")
		writeFile(t, root, `---- MODULE Root ----
---- MODULE Inner ----
CONSTANT c
VARIABLE v1
====
VARIABLE v2
BadConstant == INSTANCE Inner WITH c <- v2, v1 <- v2
BadVariable == INSTANCE Inner WITH c <- 0, v1 <- v2'
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireHasErrorContaining(t, sem, "constant-level")
		requireHasErrorContaining(t, sem, "variable-level")
	})

	t.Run("checks imported module and symbol conflicts", func(t *testing.T) {
		t.Run("rejects distinct modules with the same imported name", func(t *testing.T) {
			dir := t.TempDir()
			root := filepath.Join(dir, "Root.tla")
			writeFile(t, root, `---- MODULE Root ----
EXTENDS Naturals
---- MODULE Naturals ----
====
INSTANCE Naturals
====`)
			spec, diags := LoadSanySpec(root, LoadOptions{})
			requireNoErrors(t, diags)
			sem := CheckSpec(spec)
			requireHasErrorContaining(t, sem, "distinct modules")
		})

		t.Run("rejects imported symbols with incompatible kinds", func(t *testing.T) {
			dir := t.TempDir()
			root := filepath.Join(dir, "Root.tla")
			writeFile(t, filepath.Join(dir, "Extended.tla"), `---- MODULE Extended ----
CONSTANT _+_
====`)
			writeFile(t, root, `---- MODULE Root ----
EXTENDS Naturals, Extended
====`)
			spec, diags := LoadSanySpec(root, LoadOptions{})
			requireNoErrors(t, diags)
			extended := spec.Modules["Extended"]
			if extended == nil {
				t.Fatalf("Extended was not loaded")
			}
			foundConstantPlus := false
			for _, decl := range extended.Declarations {
				for _, name := range decl.Names {
					if name == "+" && decl.Kind == ConstantDecl {
						foundConstantPlus = true
					}
				}
			}
			if !foundConstantPlus {
				t.Fatalf("Extended declarations = %#v, want CONSTANT +", extended.Declarations)
			}
			sem := CheckSpec(spec)
			requireHasErrorContaining(t, sem, "conflicting imported symbol +")
		})

		t.Run("warns when EXTENDS imports same-kind symbols from distinct modules", func(t *testing.T) {
			dir := t.TempDir()
			root := filepath.Join(dir, "Root.tla")
			writeFile(t, filepath.Join(dir, "Extended.tla"), `---- MODULE Extended ----
x + y == x
====`)
			writeFile(t, root, `---- MODULE Root ----
EXTENDS Extended, Naturals
====`)
			spec, diags := LoadSanySpec(root, LoadOptions{})
			requireNoErrors(t, diags)
			sem := CheckSpec(spec)
			requireNoErrors(t, sem)
			requireHasWarningContaining(t, sem, "+")
		})

		t.Run("warns when plain INSTANCE imports shadow earlier instance symbols", func(t *testing.T) {
			dir := t.TempDir()
			root := filepath.Join(dir, "Root.tla")
			writeFile(t, filepath.Join(dir, "First.tla"), `---- MODULE First ----
Shared == 1
====`)
			writeFile(t, filepath.Join(dir, "Second.tla"), `---- MODULE Second ----
Shared == 2
====`)
			writeFile(t, root, `---- MODULE Root ----
INSTANCE First
INSTANCE Second
====`)
			spec, diags := LoadSanySpec(root, LoadOptions{})
			requireNoErrors(t, diags)
			sem := CheckSpec(spec)
			requireNoErrors(t, sem)
			requireHasWarningContaining(t, sem, "Shared")
		})

		t.Run("warns when plain INSTANCE imports shadow local symbols", func(t *testing.T) {
			dir := t.TempDir()
			root := filepath.Join(dir, "InstanceLocalWarnings.tla")
			writeFile(t, root, `---- MODULE InstanceLocalWarnings ----
Nat == TRUE
INSTANCE Naturals
====`)
			spec, diags := LoadSanySpec(root, LoadOptions{})
			requireNoErrors(t, diags)
			sem := CheckSpec(spec)
			requireNoErrors(t, sem)
			requireHasWarningContaining(t, sem, "Nat")
		})
	})

	t.Run("warns when record constructor field names clash with symbols", func(t *testing.T) {
		_, diags := CheckSanySource("RecordFieldWarnings.tla", `---- MODULE RecordFieldWarnings ----
Foo == TRUE
SomeRecord == [Foo |-> 42]
====`)
		requireNoErrors(t, diags)
		requireHasWarningContaining(t, diags, `The field name "Foo"`)
	})

	t.Run("warns for stale PlusCal translation checksums", func(t *testing.T) {
		cases := []struct {
			name string
			pcal string
			tla  string
			want string
		}{
			{name: "BothChanged", pcal: "00000000", tla: "00000000", want: "Both the PlusCal algorithm and its TLA+ translation"},
			{name: "AlgorithmChanged", pcal: "00000000", tla: "bff29291", want: "The PlusCal algorithm"},
			{name: "TranslationChanged", pcal: "c0cb232", tla: "00000000", want: "The TLA+ translation"},
		}
		for _, tc := range cases {
			_, diags := CheckSanySource(tc.name+".tla", plusCalChecksumSpec(tc.name, tc.pcal, tc.tla))
			requireNoErrors(t, diags)
			requireHasWarningContaining(t, diags, tc.want)
		}
	})

	t.Run("keeps LOCAL definitions private across EXTENDS", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Middle.tla"), `---- MODULE Middle ----
LOCAL Hidden == 1
Public == Hidden
====`)
		writeFile(t, filepath.Join(dir, "Root.tla"), `---- MODULE Root ----
EXTENDS Middle
Good == Public
Bad == Hidden
====`)
		spec, diags := LoadSanySpec(filepath.Join(dir, "Root.tla"), LoadOptions{})
		requireNoErrors(t, diags)
		middle := spec.Modules["Middle"]
		if middle == nil {
			t.Fatalf("Middle module was not loaded")
		}
		requireNoErrors(t, checkModule(middle, spec))
		root := spec.Modules["Root"]
		if root == nil {
			t.Fatalf("Root module was not loaded")
		}
		rootDiags := checkModule(root, spec)
		requireHasErrorContaining(t, rootDiags, "Hidden")
		if strings.Contains(rootDiags.Error(), "undefined identifier Public") {
			t.Fatalf("LOCAL visibility check hid exported Public definition:\n%s", rootDiags.Error())
		}
	})

	t.Run("exports non-local INSTANCE symbols across EXTENDS", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Helper.tla"), `---- MODULE Helper ----
EXTENDS Naturals
Zero == 0
Inc(n) == n + 1
====`)
		writeFile(t, filepath.Join(dir, "Middle.tla"), `---- MODULE Middle ----
INSTANCE Helper
Helper == INSTANCE Helper
Public == Inc(Zero)
====`)
		writeFile(t, filepath.Join(dir, "Root.tla"), `---- MODULE Root ----
EXTENDS Middle
Good == Inc(Zero) /\ Helper!Inc(Zero) = 1
====`)
		spec, diags := LoadSanySpec(filepath.Join(dir, "Root.tla"), LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
	})

	t.Run("exports nested named INSTANCE symbols across EXTENDS", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Leaf.tla"), `---- MODULE Leaf ----
Safe == TRUE
====`)
		writeFile(t, filepath.Join(dir, "Middle.tla"), `---- MODULE Middle ----
Leaf == INSTANCE Leaf
====`)
		writeFile(t, filepath.Join(dir, "Base.tla"), `---- MODULE Base ----
Outer == INSTANCE Middle
====`)
		writeFile(t, filepath.Join(dir, "Root.tla"), `---- MODULE Root ----
EXTENDS Base
UseNested == Outer!Leaf!Safe
====`)
		spec, diags := LoadSanySpec(filepath.Join(dir, "Root.tla"), LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
	})

	t.Run("model checks exported INSTANCE symbols across EXTENDS", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "Root.tla")
		writeFile(t, filepath.Join(dir, "Helper.tla"), `---- MODULE Helper ----
EXTENDS Naturals
Zero == 0
Inc(n) == n + 1
====`)
		writeFile(t, filepath.Join(dir, "Middle.tla"), `---- MODULE Middle ----
INSTANCE Helper
Helper == INSTANCE Helper
====`)
		writeFile(t, root, `---- MODULE Root ----
EXTENDS Middle
VARIABLE x
Init == x = Zero
Next == (x < Helper!Inc(Zero) /\ x' = Inc(x)) \/ (x = Helper!Inc(Zero) /\ x' = x)
Inv == x <= Helper!Inc(Zero)
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
		cfg, cfgDiags := ParseConfigSource(filepath.Join(dir, "MC.cfg"), "INIT Init\nNEXT Next\nINVARIANT Inv\n")
		requireNoErrors(t, cfgDiags)
		result, runDiags := ModelCheck(spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, runDiags)
		if !result.OK || result.StatesExplored != 2 {
			t.Fatalf("result = %#v, want exported instance success with 2 states", result)
		}
	})

	t.Run("keeps LOCAL INSTANCE symbols private across EXTENDS", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Helper.tla"), `---- MODULE Helper ----
Zero == 0
====`)
		writeFile(t, filepath.Join(dir, "Middle.tla"), `---- MODULE Middle ----
LOCAL INSTANCE Helper
Public == Zero
====`)
		writeFile(t, filepath.Join(dir, "Root.tla"), `---- MODULE Root ----
EXTENDS Middle
Bad == Zero
BadQualified == Helper!Zero
====`)
		spec, diags := LoadSanySpec(filepath.Join(dir, "Root.tla"), LoadOptions{})
		requireNoErrors(t, diags)
		middle := spec.Modules["Middle"]
		if middle == nil {
			t.Fatalf("Middle module was not loaded")
		}
		requireNoErrors(t, checkModule(middle, spec))
		root := spec.Modules["Root"]
		if root == nil {
			t.Fatalf("Root module was not loaded")
		}
		rootDiags := checkModule(root, spec)
		requireHasErrorContaining(t, rootDiags, "Zero")
		requireHasErrorContaining(t, rootDiags, "Helper!Zero")
	})

	t.Run("applies INSTANCE WITH substitutions to qualified calls", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "Root.tla")
		writeFile(t, filepath.Join(dir, "Helper.tla"), `---- MODULE Helper ---- EXTENDS Naturals
CONSTANT C
Start == C
Inc(n) == n + C
====`)
		writeFile(t, root, `---- MODULE Root ----
INSTANCE Helper WITH C <- 1
VARIABLE x
Init == x = Helper!Start
Next == (x < 3 /\ x' = Helper!Inc(x)) \/ (x = 3 /\ x' = x)
Inv == x <= 3
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		cfg, cfgDiags := ParseConfigSource(filepath.Join(dir, "MC.cfg"), "INIT Init\nNEXT Next\nINVARIANT Inv\n")
		requireNoErrors(t, cfgDiags)
		result, runDiags := ModelCheck(spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, runDiags)
		if !result.OK || result.StatesExplored != 3 {
			t.Fatalf("result = %#v, want success with qualified substituted instance", result)
		}
	})

	t.Run("inlines sibling module operator calls for SANY model checking", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "Root.tla")
		writeFile(t, filepath.Join(dir, "Helper.tla"), `---- MODULE Helper ---- EXTENDS Naturals
Inc(n) == n + 1
====`)
		writeFile(t, root, `---- MODULE Root ----
EXTENDS Helper
Helper == INSTANCE Helper
VARIABLE x
Init == x = 0
Next == (x < 3 /\ x' = Helper!Inc(x)) \/ (x = 3 /\ x' = x)
Inv == Helper!Inc(x) <= 4
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		cfg, cfgDiags := ParseConfigSource(filepath.Join(dir, "MC.cfg"), "INIT Init\nNEXT Next\nINVARIANT Inv\n")
		requireNoErrors(t, cfgDiags)
		result, runDiags := ModelCheck(spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, runDiags)
		if !result.OK || result.StatesExplored != 4 {
			t.Fatalf("result = %#v, want success with 4 states", result)
		}
	})

	t.Run("checks named INSTANCE sibling references", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "Root.tla")
		writeFile(t, filepath.Join(dir, "Helper.tla"), `---- MODULE Helper ---- EXTENDS Naturals
Zero == 0
Inc(n) == n + 1
====`)
		writeFile(t, root, `---- MODULE Root ----
EXTENDS Helper
Helper == INSTANCE Helper
Good == Helper!Inc(Helper!Zero)
BadName == Helper!Missing
BadArity == Helper!Inc(1, 2)
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireHasErrorContaining(t, sem, "undefined")
		requireHasErrorContaining(t, sem, "arity")
	})

	t.Run("inlines named INSTANCE infix operator calls", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "Root.tla")
		writeFile(t, filepath.Join(dir, "Helper.tla"), `---- MODULE Helper ---- EXTENDS Naturals
a \oplus b == a + b
====`)
		writeFile(t, root, `---- MODULE Root ----
EXTENDS Helper
Helper == INSTANCE Helper
VARIABLE x
Init == x = Helper!\oplus(1, 2)
Next == x' = x
Inv == x = Helper!\oplus(1, 2)
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		cfg, cfgDiags := ParseConfigSource(filepath.Join(dir, "MC.cfg"), "INIT Init\nNEXT Next\nINVARIANT Inv\n")
		requireNoErrors(t, cfgDiags)
		result, runDiags := ModelCheck(spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, runDiags)
		if !result.OK || result.StatesExplored != 1 {
			t.Fatalf("result = %#v, want success with qualified infix operator call", result)
		}
	})

	t.Run("model checks prefix junction lists through SANY syntax", func(t *testing.T) {
		spec := `---- MODULE JunctionLists ---- EXTENDS Naturals
VARIABLE x, y
Init == /\ x = 0
        /\ y = 0
Next == \/ /\ x < 1
           /\ x' = x + 1
           /\ UNCHANGED y
        \/ /\ x = 1
           /\ UNCHANGED <<x, y>>
Inv == /\ x <= 1
       /\ y = 0
====`
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("JunctionLists.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 2 {
			t.Fatalf("result = %#v, want success with 2 states", result)
		}
	})

	t.Run("checks nested action-style junction indentation through SANY syntax", func(t *testing.T) {
		_, diags := CheckSanySource("NestedJunctions.tla", `---- MODULE NestedJunctions ----
VARIABLE x, y, z
Step == /\ x = 0
        /\ \/ /\ x' = 1
              /\ y' = y
           \/ /\ x' = 2
              /\ y' = y
        /\ z' = z
NotBoth == ~ /\ x = 1
             /\ y = 1
Guarded == \A i, j \in {1, 2} : (i # j) => ~ /\ x = i
                                                /\ y = j
CONSTANT A, B, C
ImplicationOperand == A =>
  /\ B
  /\ C
NestedImplicationOperand ==
  /\ A =>
    /\ B
    /\ C
  /\ B
====`)
		requireNoErrors(t, diags)
	})

	t.Run("checks function constructors over set comprehensions in records", func(t *testing.T) {
		_, diags := CheckSanySource("FunctionConstructorRecord.tla", `---- MODULE FunctionConstructorRecord ---- EXTENDS Naturals
CONSTANT S, f, Zero
PrettyPrint(x) == x
View == [
  weights |-> [p \in {q \in S : f[q] # Zero} |-> PrettyPrint(f[p])],
  nested |-> [p \in {q \in S : f[q] # Zero} |-> [i \in 1..1 |-> PrettyPrint(f[p][i])]]
]
====`)
		requireNoErrors(t, diags)
	})

	t.Run("checks community helper exports for range and extrema", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "CommunityHelpers.tla")
		writeFile(t, root, `---- MODULE CommunityHelpers ----
EXTENDS FiniteSetsExt
CONSTANT f
Values == Range(f)
Lo == Min(Values)
Hi == Max(Values)
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
	})

	t.Run("checks sequence extension helper exports", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "SequenceHelpers.tla")
		writeFile(t, root, `---- MODULE SequenceHelpers ----
EXTENDS SequencesExt
Removed == RemoveAt(<<1, 2, 3>>, 2)
FrontPart == Front(<<1, 2, 3>>)
Strict == IsStrictPrefix(<<1>>, <<1, 2>>)
Common == LongestCommonPrefix({<<1, 2>>, <<1, 3>>})
Prepended == Cons(0, <<1, 2, 3>>)
====`)
		spec, diags := LoadSanySpec(root, LoadOptions{})
		requireNoErrors(t, diags)
		sem := CheckSpec(spec)
		requireNoErrors(t, sem)
	})

	t.Run("keeps SequencesExt helpers out of base Sequences", func(t *testing.T) {
		_, diags := CheckSanySource("BaseSequencesNoCons.tla", `---- MODULE BaseSequencesNoCons ----
EXTENDS Sequences
Prepended == Cons(0, <<1, 2, 3>>)
====`)
		if !diags.HasErrors() || !strings.Contains(diags.Error(), "undefined identifier Cons") {
			t.Fatalf("expected Cons to stay undefined under base Sequences, got:\n%s", diags.Error())
		}
	})
}

func plusCalChecksumSpec(name, pcalChecksum, tlaChecksum string) string {
	template := `---- MODULE {{NAME}} ----
(*
--algorithm Test {
  {
    lbl: skip;
  }
}
*)
\* BEGIN TRANSLATION (chksum(pcal) = "{{PCAL}}" /\ chksum(tla) = "{{TLA}}")
VARIABLE pc

vars == << pc >>

Init == /\ pc = "lbl"

lbl == /\ pc = "lbl"
       /\ TRUE
       /\ pc' = "Done"

(* Allow infinite stuttering to prevent deadlock on termination. *)
Terminating == pc = "Done" /\ UNCHANGED vars

Next == lbl
           \/ Terminating

Spec == Init /\ [][Next]_vars

Termination == <>(pc = "Done")

\* END TRANSLATION 
====`
	template = strings.ReplaceAll(template, "{{NAME}}", name)
	template = strings.ReplaceAll(template, "{{PCAL}}", pcalChecksum)
	return strings.ReplaceAll(template, "{{TLA}}", tlaChecksum)
}
