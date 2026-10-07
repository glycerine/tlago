package tlago

import "testing"

func TestSanyParserBehaviors(t *testing.T) {
	t.Run("builds the SANY module spine with extends and declarations", func(t *testing.T) {
		source := `---- MODULE M ----
EXTENDS Naturals, TLC
CONSTANT Max
\* attached to variable declaration
VARIABLE x, y
Init == x = 0
====`
		root, diags := ParseSanySyntax("M.tla", source)
		requireNoErrors(t, diags)
		if root.Kind.JavaName() != "N_Module" {
			t.Fatalf("root kind = %s, want N_Module", root.Kind.JavaName())
		}
		requireSanyNodeKinds(t, root.GetHeirs(), "N_BeginModule", "N_Extends", "N_Body", "N_EndModule")
		begin := root.GetHeirs()[0]
		if begin.GetHeirs()[1].Image != "M" {
			t.Fatalf("module name image = %q, want M", begin.GetHeirs()[1].Image)
		}
		body := root.GetHeirs()[2]
		if countSanyChildren(body, "N_ParamDeclaration") != 1 || countSanyChildren(body, "N_VariableDeclaration") != 1 || countSanyChildren(body, "N_OperatorDefinition") != 1 {
			t.Fatalf("body heirs = %v", sanyNodeKindNames(body.GetHeirs()))
		}
		varDecl := findSanyChild(t, body, "N_VariableDeclaration")
		if len(varDecl.Zero) != 1 || varDecl.Zero[0].Kind.JavaName() != "VARIABLE" {
			t.Fatalf("variable zero heirs = %v", sanyNodeKindNames(varDecl.Zero))
		}
		if got := len(varDecl.Zero[0].PreComments); got != 1 {
			t.Fatalf("variable pre-comments = %d, want 1", got)
		}
		def := findSanyChild(t, body, "N_OperatorDefinition")
		expr := def.GetHeirs()[2]
		if expr.Kind.JavaName() != "N_InfixExpr" {
			t.Fatalf("definition expression kind = %s, want N_InfixExpr", expr.Kind.JavaName())
		}
	})

	t.Run("ignores tool commands after the root module footer", func(t *testing.T) {
		_, diags := ParseSanySyntax("TrailingCommand.tla", `---- MODULE TrailingCommand ----
A == TRUE
====

$ rm *.csv ; tlc TrailingCommand -note -generate -depth -1`)
		requireNoErrors(t, diags)
	})

	t.Run("parses adjacent same-line operator definitions", func(t *testing.T) {
		root, diags := ParseSanySyntax("SameLineDefs.tla", `---- MODULE SameLineDefs ----
W == 4 H == 5
Pos == W + H
====`)
		requireNoErrors(t, diags)
		body := root.GetHeirs()[2]
		defs := collectSanyChildrenByKind(body, "N_OperatorDefinition")
		if got := len(defs); got != 3 {
			t.Fatalf("operator definitions = %d, want 3; body heirs = %v", got, sanyNodeKindNames(body.GetHeirs()))
		}
		for i, want := range []string{"W", "H", "Pos"} {
			lhs := defs[i].GetHeirs()[0]
			if len(lhs.GetHeirs()) == 0 || lhs.GetHeirs()[0].Image != want {
				t.Fatalf("definition %d lhs = %v, want %s", i, sanyNodeKindNames(lhs.GetHeirs()), want)
			}
		}
	})

	t.Run("outdented junction bullet ends nested quantified junction body", func(t *testing.T) {
		root, diags := ParseSanySyntax("OutdentedQuantJunction.tla", `---- MODULE OutdentedQuantJunction ----
VARIABLE x
A(p) == /\ x = x
   /\ \E req \in S :
         /\ Send(p, req)
         /\ x' = x
   /\ UNCHANGED x
====`)
		requireNoErrors(t, diags)
		def := onlySanyDefinition(t, root)
		expr := def.GetHeirs()[2]
		if expr.Kind.JavaName() != "N_InfixExpr" {
			t.Fatalf("definition expression = %s, want N_InfixExpr", expr.Kind.JavaName())
		}
		if got := countSanyDescendants(expr, "N_BoundQuant"); got != 1 {
			t.Fatalf("bound quantifier count = %d, want 1", got)
		}
	})

	t.Run("reports SANY operator precedence conflicts inside definitions", func(t *testing.T) {
		_, diags := ParseSanySyntax("Bad.tla", `---- MODULE Bad ----
Bad == A = B = C
====`)
		requireHasErrorContaining(t, diags, "precedence conflict")
	})

	t.Run("parses operator constant declarations like SANY", func(t *testing.T) {
		root, diags := ParseSanySyntax("Consts.tla", `---- MODULE Consts ----
CONSTANT C, F(_), G(_, _)
====`)
		requireNoErrors(t, diags)
		body := root.GetHeirs()[2]
		decl := findSanyChild(t, body, "N_ParamDeclaration")
		requireSanyNodeKinds(t, decl.GetHeirs(), "N_ConsDecl", "N_IdentDecl", "COMMA", "N_IdentDecl", "COMMA", "N_IdentDecl")
		f := decl.GetHeirs()[3]
		requireSanyNodeKinds(t, f.GetHeirs(), "IDENTIFIER", "LBR", "US", "RBR")
		g := decl.GetHeirs()[5]
		requireSanyNodeKinds(t, g.GetHeirs(), "IDENTIFIER", "LBR", "US", "COMMA", "US", "RBR")
	})

	t.Run("parses identifier LHS parameters like SANY", func(t *testing.T) {
		root, diags := ParseSanySyntax("Defs.tla", `---- MODULE Defs ----
Add(a, b) == a + b
====`)
		requireNoErrors(t, diags)
		def := onlySanyDefinition(t, root)
		if def.Kind.JavaName() != "N_OperatorDefinition" {
			t.Fatalf("definition kind = %s", def.Kind.JavaName())
		}
		lhs := def.GetHeirs()[0]
		if lhs.Kind.JavaName() != "N_IdentLHS" {
			t.Fatalf("lhs kind = %s, want N_IdentLHS", lhs.Kind.JavaName())
		}
		requireSanyNodeKinds(t, lhs.GetHeirs(), "IDENTIFIER", "LBR", "N_IdentDecl", "COMMA", "N_IdentDecl", "RBR")
	})

	t.Run("parses function definitions with quantifier bounds", func(t *testing.T) {
		root, diags := ParseSanySyntax("Defs.tla", `---- MODULE Defs ----
F[x \in S, y \in T] == x + y
====`)
		requireNoErrors(t, diags)
		def := onlySanyDefinition(t, root)
		if def.Kind.JavaName() != "N_FunctionDefinition" {
			t.Fatalf("definition kind = %s, want N_FunctionDefinition", def.Kind.JavaName())
		}
		requireSanyNodeKinds(t, def.GetHeirs(), "IDENTIFIER", "LSB", "N_QuantBound", "COMMA", "N_QuantBound", "RSB", "DEF", "N_InfixExpr")
		firstBound := def.GetHeirs()[2]
		requireSanyNodeKinds(t, firstBound.GetHeirs(), "IDENTIFIER", "T_IN", "N_GeneralId")
		requireSanyNodeKinds(t, firstBound.GetHeirs()[2].GetHeirs(), "N_IdPrefix", "IDENTIFIER")
	})

	t.Run("parses infix LHS definitions", func(t *testing.T) {
		root, diags := ParseSanySyntax("Defs.tla", `---- MODULE Defs ----
a + b == a
====`)
		requireNoErrors(t, diags)
		def := onlySanyDefinition(t, root)
		lhs := def.GetHeirs()[0]
		if lhs.Kind.JavaName() != "N_InfixLHS" {
			t.Fatalf("lhs kind = %s, want N_InfixLHS", lhs.Kind.JavaName())
		}
		requireSanyNodeKinds(t, lhs.GetHeirs(), "IDENTIFIER", "op_81", "IDENTIFIER")
	})

	t.Run("parses recursive instance assumption and theorem body items", func(t *testing.T) {
		root, diags := ParseSanySyntax("Body.tla", `---- MODULE Body ----
RECURSIVE Step
INSTANCE Helper WITH C <- D, X <- Y
ASSUME A1 == X = Y
THEOREM T1 == X = X
====`)
		requireNoErrors(t, diags)
		body := root.GetHeirs()[2]
		requireSanyNodeKinds(t, body.GetHeirs(), "N_Recursive", "N_Instance", "N_Assumption", "N_Theorem")
		rec := body.GetHeirs()[0]
		requireSanyNodeKinds(t, rec.GetHeirs(), "RECURSIVE", "N_IdentDecl")
		inst := body.GetHeirs()[1]
		if inst.Kind.JavaName() != "N_Instance" || len(inst.GetHeirs()) != 1 {
			t.Fatalf("instance node = %s heirs %v", inst.Kind.JavaName(), sanyNodeKindNames(inst.GetHeirs()))
		}
		nonLocal := inst.GetHeirs()[0]
		if countSanyChildren(nonLocal, "N_Substitution") != 2 {
			t.Fatalf("instance heirs = %v", sanyNodeKindNames(nonLocal.GetHeirs()))
		}
		assume := body.GetHeirs()[2]
		requireSanyNodeKinds(t, assume.GetHeirs(), "ASSUME", "IDENTIFIER", "DEF", "N_InfixExpr")
		theorem := body.GetHeirs()[3]
		requireSanyNodeKinds(t, theorem.GetHeirs(), "THEOREM", "IDENTIFIER", "DEF", "N_InfixExpr")
	})

	t.Run("parses nested module body items", func(t *testing.T) {
		root, diags := ParseSanySyntax("Nested.tla", `---- MODULE Outer ----
---- MODULE Inner ----
A == 1
====
B == 2
====`)
		requireNoErrors(t, diags)
		body := root.GetHeirs()[2]
		requireSanyNodeKinds(t, body.GetHeirs(), "N_Module", "N_OperatorDefinition")
		inner := body.GetHeirs()[0]
		if got := SanyModuleName(inner); got != "Inner" {
			t.Fatalf("nested module name = %q, want Inner", got)
		}
		innerBody := inner.GetHeirs()[2]
		requireSanyNodeKinds(t, innerBody.GetHeirs(), "N_OperatorDefinition")
	})

	t.Run("parses theorem terminal proofs", func(t *testing.T) {
		root, diags := ParseSanySyntax("Proofs.tla", `---- MODULE Proofs ----
THEOREM T == TRUE PROOF OMITTED
THEOREM U == TRUE PROOF OBVIOUS
====`)
		requireNoErrors(t, diags)
		body := root.GetHeirs()[2]
		theorems := countSanyChildren(body, "N_Theorem")
		if theorems != 2 {
			t.Fatalf("theorem count = %d, want 2", theorems)
		}
		for _, theorem := range body.GetHeirs() {
			if theorem.Kind.JavaName() != "N_Theorem" {
				continue
			}
			if countSanyChildren(theorem, "N_TerminalProof") != 1 {
				t.Fatalf("theorem heirs = %v, want terminal proof", sanyNodeKindNames(theorem.GetHeirs()))
			}
		}
	})

	t.Run("parses terminal proofs with BY DEF references", func(t *testing.T) {
		root, diags := ParseSanySyntax("Proofs.tla", `---- MODULE Proofs ----
THEOREM Known == TRUE PROOF OBVIOUS
LEMMA UsesDef == TRUE
PROOF BY RefersTo(Known, "Known") DEF Known
====`)
		requireNoErrors(t, diags)
		body := root.GetHeirs()[2]
		if got := countSanyDescendants(body, "N_TerminalProof"); got != 2 {
			t.Fatalf("terminal proofs = %d, want 2", got)
		}
	})

	t.Run("parses simple QED proof blocks", func(t *testing.T) {
		root, diags := ParseSanySyntax("Proofs.tla", `---- MODULE Proofs ----
THEOREM T == TRUE PROOF QED
====`)
		requireNoErrors(t, diags)
		theorem := findSanyChild(t, root.GetHeirs()[2], "N_Theorem")
		proof := findSanyChild(t, theorem, "N_Proof")
		if countSanyChildren(proof, "N_ProofStep") != 1 {
			t.Fatalf("proof heirs = %v, want QED proof step", sanyNodeKindNames(proof.GetHeirs()))
		}
		qedStep := findSanyChild(t, proof.GetHeirs()[1], "N_QEDStep")
		if qedStep == nil {
			t.Fatalf("QED step not found")
		}
	})

	t.Run("parses numbered assertion proof steps", func(t *testing.T) {
		root, diags := ParseSanySyntax("Proofs.tla", `---- MODULE Proofs ----
THEOREM T == TRUE
PROOF
<1>1. TRUE
<1>2. SUFFICES TRUE
<1> QED
====`)
		requireNoErrors(t, diags)
		theorem := findSanyChild(t, root.GetHeirs()[2], "N_Theorem")
		proof := findSanyChild(t, theorem, "N_Proof")
		steps := collectSanyChildrenByKind(proof, "N_ProofStep")
		if len(steps) != 3 {
			t.Fatalf("proof steps = %d heirs %v, want 3", len(steps), sanyNodeKindNames(proof.GetHeirs()))
		}
		findSanyChild(t, steps[0], "N_AssertStep")
		findSanyChild(t, steps[1], "N_AssertStep")
		findSanyChild(t, steps[2], "N_QEDStep")
	})

	t.Run("parses named proof command steps", func(t *testing.T) {
		root, diags := ParseSanySyntax("Proofs.tla", `---- MODULE Proofs ----
CONSTANT S
Known == TRUE
THEOREM T == TRUE
PROOF
<1>1. HAVE TRUE
<1>2. TAKE x
<1>3. WITNESS x, y
<1>4. PICK z \in S : TRUE
<1>5. CASE TRUE
<1>6. USE Known
<1>7. HIDE Known
<1>8. DEFINE Local == TRUE
<1> QED
====`)
		requireNoErrors(t, diags)
		theorem := findSanyChild(t, root.GetHeirs()[2], "N_Theorem")
		proof := findSanyChild(t, theorem, "N_Proof")
		steps := collectSanyChildrenByKind(proof, "N_ProofStep")
		if len(steps) != 9 {
			t.Fatalf("proof steps = %d heirs %v, want 9", len(steps), sanyNodeKindNames(proof.GetHeirs()))
		}
		requireSanyProofStepBodies(t, steps,
			"N_HaveStep",
			"N_TakeStep",
			"N_WitnessStep",
			"N_PickStep",
			"N_CaseStep",
			"N_UseOrHide",
			"N_UseOrHide",
			"N_DefStep",
			"N_QEDStep",
		)
	})

	t.Run("parses proof binder variants", func(t *testing.T) {
		root, diags := ParseSanySyntax("Proofs.tla", `---- MODULE Proofs ----
CONSTANT S
THEOREM T == TRUE
PROOF
<1>1. TAKE x \in S
<1>2. PICK y : TRUE
<1> QED
====`)
		requireNoErrors(t, diags)
		theorem := findSanyChild(t, root.GetHeirs()[2], "N_Theorem")
		proof := findSanyChild(t, theorem, "N_Proof")
		steps := collectSanyChildrenByKind(proof, "N_ProofStep")
		if len(steps) != 3 {
			t.Fatalf("proof steps = %d heirs %v, want 3", len(steps), sanyNodeKindNames(proof.GetHeirs()))
		}
		take := findSanyChild(t, steps[0], "N_TakeStep")
		pick := findSanyChild(t, steps[1], "N_PickStep")
		if countSanyChildren(take, "N_QuantBound") != 1 {
			t.Fatalf("TAKE heirs = %v, want N_QuantBound", sanyNodeKindNames(take.GetHeirs()))
		}
		if countSanyChildren(pick, "N_IdentDecl") != 1 {
			t.Fatalf("PICK heirs = %v, want N_IdentDecl", sanyNodeKindNames(pick.GetHeirs()))
		}
	})

	t.Run("parses ASSUME PROVE assertion proof steps", func(t *testing.T) {
		root, diags := ParseSanySyntax("Proofs.tla", `---- MODULE Proofs ----
CONSTANT S
THEOREM T == TRUE
PROOF
<1>1. ASSUME TRUE, x \in S PROVE TRUE
<1>2. SUFFICES ASSUME TRUE PROVE TRUE
<1> QED
====`)
		requireNoErrors(t, diags)
		theorem := findSanyChild(t, root.GetHeirs()[2], "N_Theorem")
		proof := findSanyChild(t, theorem, "N_Proof")
		steps := collectSanyChildrenByKind(proof, "N_ProofStep")
		if len(steps) != 3 {
			t.Fatalf("proof steps = %d heirs %v, want 3", len(steps), sanyNodeKindNames(proof.GetHeirs()))
		}
		firstAssert := findSanyChild(t, steps[0], "N_AssertStep")
		secondAssert := findSanyChild(t, steps[1], "N_AssertStep")
		if countSanyChildren(firstAssert, "N_AssumeProve") != 1 {
			t.Fatalf("first assert heirs = %v, want N_AssumeProve", sanyNodeKindNames(firstAssert.GetHeirs()))
		}
		if countSanyChildren(secondAssert, "N_AssumeProve") != 1 {
			t.Fatalf("second assert heirs = %v, want N_AssumeProve", sanyNodeKindNames(secondAssert.GetHeirs()))
		}
	})

	t.Run("parses ASSUME PROVE theorem bodies", func(t *testing.T) {
		root, diags := ParseSanySyntax("Proofs.tla", `---- MODULE Proofs ----
THEOREM T == ASSUME TRUE PROVE TRUE
====`)
		requireNoErrors(t, diags)
		theorem := findSanyChild(t, root.GetHeirs()[2], "N_Theorem")
		findSanyChild(t, theorem, "N_AssumeProve")
	})

	t.Run("parses NEW declarations in ASSUME PROVE bodies", func(t *testing.T) {
		root, diags := ParseSanySyntax("Proofs.tla", `---- MODULE Proofs ----
CONSTANT S
THEOREM T == ASSUME NEW x \in S, NEW VARIABLE v, NEW CONSTANT C PROVE TRUE
====`)
		requireNoErrors(t, diags)
		theorem := findSanyChild(t, root.GetHeirs()[2], "N_Theorem")
		assumeProve := findSanyChild(t, theorem, "N_AssumeProve")
		if countSanyChildren(assumeProve, "N_NewSymb") != 3 {
			t.Fatalf("ASSUME/PROVE heirs = %v, want three N_NewSymb declarations", sanyNodeKindNames(assumeProve.GetHeirs()))
		}
	})

	t.Run("parses labeled nested ASSUME PROVE items", func(t *testing.T) {
		root, diags := ParseSanySyntax("Proofs.tla", `---- MODULE Proofs ----
CONSTANT S
THEOREM T ==
  ASSUME
    inner ::
      ASSUME NEW c \in S
      PROVE c \in S,
    exprLbl :: TRUE
  PROVE TRUE
PROOF BY T!inner, T!exprLbl
====`)
		requireNoErrors(t, diags)
		theorem := findSanyChild(t, root.GetHeirs()[2], "N_Theorem")
		if got := countSanyDescendants(theorem, "N_AssumeProve"); got != 2 {
			t.Fatalf("ASSUME/PROVE count = %d, want nested outer and inner nodes", got)
		}
		if got := countSanyDescendants(theorem, "N_Label"); got != 2 {
			t.Fatalf("label count = %d, want labels for nested proof item and expression item", got)
		}
	})

	t.Run("parses tuple-bound TAKE and PICK proof steps", func(t *testing.T) {
		root, diags := ParseSanySyntax("Proofs.tla", `---- MODULE Proofs ----
CONSTANT S
THEOREM T == TRUE
PROOF
<1>a TAKE <<x, y>> \in S, u \in S
<1>b PICK <<z, w>> \in S, v \in S : TRUE
<1> QED
====`)
		requireNoErrors(t, diags)
		theorem := findSanyChild(t, root.GetHeirs()[2], "N_Theorem")
		proof := findSanyChild(t, theorem, "N_Proof")
		steps := collectSanyChildrenByKind(proof, "N_ProofStep")
		requireSanyProofStepBodies(t, steps, "N_TakeStep", "N_PickStep", "N_QEDStep")
		take := findSanyChild(t, steps[0], "N_TakeStep")
		pick := findSanyChild(t, steps[1], "N_PickStep")
		if got := countSanyChildren(take, "N_QuantBound"); got != 2 {
			t.Fatalf("TAKE quant bounds = %d, want 2", got)
		}
		if got := countSanyChildren(pick, "N_QuantBound"); got != 2 {
			t.Fatalf("PICK quant bounds = %d, want 2", got)
		}
	})

	t.Run("parses implicit proof step blocks", func(t *testing.T) {
		root, diags := ParseSanySyntax("Proofs.tla", `---- MODULE Proofs ----
THEOREM T == TRUE
<*> TRUE
<0>a TRUE
<0> QED BY <0>a
====`)
		requireNoErrors(t, diags)
		theorem := findSanyChild(t, root.GetHeirs()[2], "N_Theorem")
		proof := findSanyChild(t, theorem, "N_Proof")
		steps := collectSanyChildrenByKind(proof, "N_ProofStep")
		requireSanyProofStepBodies(t, steps, "N_AssertStep", "N_AssertStep", "N_QEDStep")
	})

	t.Run("does not let nested QED BY proofs swallow enclosing QED steps", func(t *testing.T) {
		root, diags := ParseSanySyntax("Proofs.tla", `---- MODULE Proofs ----
THEOREM T == TRUE
<*> TRUE
<0>b TRUE
  PROOF
  <*> TRUE
  <*> QED BY TRUE
<0>d QED BY <0>b
====`)
		requireNoErrors(t, diags)
		theorem := findSanyChild(t, root.GetHeirs()[2], "N_Theorem")
		proof := findSanyChild(t, theorem, "N_Proof")
		steps := collectSanyChildrenByKind(proof, "N_ProofStep")
		requireSanyProofStepBodies(t, steps, "N_AssertStep", "N_AssertStep", "N_QEDStep")
	})

	t.Run("parses higher-level child proof steps after assertions", func(t *testing.T) {
		root, diags := ParseSanySyntax("Proofs.tla", `---- MODULE Proofs ----
CONSTANT S
THEOREM T == TRUE
<1>1. TRUE
<1>2. TRUE
  <2>. SUFFICES ASSUME TRUE, TRUE
                PROVE TRUE
    OBVIOUS
  <2>1. ASSUME NEW self \in S, TRUE
        PROVE TRUE
    BY <2>1 DEF T
  <2>2. CASE /\ TRUE
             /\ TRUE
    <3>. SUFFICES ASSUME NEW q \in S
                  PROVE TRUE
      BY DEF T
    <3>1. CASE q = q
      BY <3>1
    <3>. QED BY <3>1
  <2>. QED BY <2>1, <2>2
<1>. QED BY <1>1, <1>2
====`)
		requireNoErrors(t, diags)
		theorem := findSanyChild(t, root.GetHeirs()[2], "N_Theorem")
		if got := countSanyDescendants(theorem, "N_Proof"); got < 3 {
			t.Fatalf("nested proof count = %d, want at least 3", got)
		}
		if got := countSanyDescendants(theorem, "N_TerminalProof"); got < 3 {
			t.Fatalf("terminal proof count = %d, want at least 3", got)
		}
	})

	t.Run("parses proof step references in USE proof steps", func(t *testing.T) {
		root, diags := ParseSanySyntax("Proofs.tla", `---- MODULE Proofs ----
THEOREM T == TRUE
<1>1. TRUE
  <2>. USE <1>1 DEF T
  <2>1. TRUE
  <2>. QED BY <2>1
<1>. QED BY <1>1
====`)
		requireNoErrors(t, diags)
		theorem := findSanyChild(t, root.GetHeirs()[2], "N_Theorem")
		if got := countSanyDescendants(theorem, "N_UseOrHide"); got != 1 {
			t.Fatalf("USE/HIDE steps = %d, want one", got)
		}
	})

	t.Run("parses proof DEFINE module instances", func(t *testing.T) {
		root, diags := ParseSanySyntax("Proofs.tla", `---- MODULE Proofs ----
---- MODULE Inner ----
A == TRUE
====
THEOREM T == TRUE
<1> DEFINE h == INSTANCE Inner
<1> QED BY h!A
====`)
		requireNoErrors(t, diags)
		theorem := findSanyChild(t, root.GetHeirs()[2], "N_Theorem")
		proof := findSanyChild(t, theorem, "N_Proof")
		steps := collectSanyChildrenByKind(proof, "N_ProofStep")
		requireSanyProofStepBodies(t, steps, "N_DefStep", "N_QEDStep")
		defStep := findSanyChild(t, steps[0], "N_DefStep")
		if countSanyChildren(defStep, "N_ModuleDefinition") != 1 {
			t.Fatalf("DEFINE heirs = %v, want N_ModuleDefinition", sanyNodeKindNames(defStep.GetHeirs()))
		}
	})

	t.Run("parses dotted proof step definitions without DEFINE", func(t *testing.T) {
		root, diags := ParseSanySyntax("Proofs.tla", `---- MODULE Proofs ----
THEOREM TRUE
<1>a..... op == 1
<1>b..... QED
====`)
		requireNoErrors(t, diags)
		theorem := findSanyChild(t, root.GetHeirs()[2], "N_Theorem")
		proof := findSanyChild(t, theorem, "N_Proof")
		steps := collectSanyChildrenByKind(proof, "N_ProofStep")
		requireSanyProofStepBodies(t, steps, "N_DefStep", "N_QEDStep")
	})

	t.Run("wraps literal and parenthesized primitive expressions in SANY nodes", func(t *testing.T) {
		root, diags := ParseSanySyntax("Exprs.tla", `---- MODULE Exprs ----
N == 42
R == 1.5
S == "hi"
P == (A = B)
====`)
		requireNoErrors(t, diags)
		body := root.GetHeirs()[2]
		defs := collectSanyDefinitions(body)
		if len(defs) != 4 {
			t.Fatalf("definition count = %d, want 4", len(defs))
		}
		if expr := defs[0].GetHeirs()[2]; expr.Kind.JavaName() != "N_Number" {
			t.Fatalf("N expression = %s, want N_Number", expr.Kind.JavaName())
		}
		if expr := defs[1].GetHeirs()[2]; expr.Kind.JavaName() != "N_Real" {
			t.Fatalf("R expression = %s, want N_Real", expr.Kind.JavaName())
		}
		if expr := defs[2].GetHeirs()[2]; expr.Kind.JavaName() != "N_String" || expr.Image != `"hi"` {
			t.Fatalf("S expression = %s/%q, want N_String/quoted hi", expr.Kind.JavaName(), expr.Image)
		}
		if expr := defs[3].GetHeirs()[2]; expr.Kind.JavaName() != "N_ParenExpr" {
			t.Fatalf("P expression = %s, want N_ParenExpr", expr.Kind.JavaName())
		}
	})

	t.Run("parses conditional and quantified open expressions", func(t *testing.T) {
		root, diags := ParseSanySyntax("Exprs.tla", `---- MODULE Exprs ----
I == IF A THEN B ELSE C
BQ == \A x \in S : P
UQ == \E x, y : P
====`)
		requireNoErrors(t, diags)
		defs := collectSanyDefinitions(root.GetHeirs()[2])
		if len(defs) != 3 {
			t.Fatalf("definition count = %d, want 3", len(defs))
		}
		if expr := defs[0].GetHeirs()[2]; expr.Kind.JavaName() != "N_IfThenElse" {
			t.Fatalf("I expression = %s, want N_IfThenElse", expr.Kind.JavaName())
		}
		if expr := defs[1].GetHeirs()[2]; expr.Kind.JavaName() != "N_BoundQuant" {
			t.Fatalf("BQ expression = %s, want N_BoundQuant", expr.Kind.JavaName())
		}
		if expr := defs[2].GetHeirs()[2]; expr.Kind.JavaName() != "N_UnboundQuant" {
			t.Fatalf("UQ expression = %s, want N_UnboundQuant", expr.Kind.JavaName())
		}
	})

	t.Run("parses temporal quantifiers", func(t *testing.T) {
		root, diags := ParseSanySyntax("Exprs.tla", `---- MODULE Exprs ----
TE == \EE x : TRUE
TA == \AA x, y : TRUE
====`)
		requireNoErrors(t, diags)
		defs := collectSanyDefinitions(root.GetHeirs()[2])
		if len(defs) != 2 {
			t.Fatalf("definition count = %d, want 2", len(defs))
		}
		for _, def := range defs {
			if expr := def.GetHeirs()[2]; expr.Kind.JavaName() != "N_UnboundQuant" {
				t.Fatalf("temporal quantifier expression = %s heirs %v, want N_UnboundQuant", expr.Kind.JavaName(), sanyNodeKindNames(expr.GetHeirs()))
			}
		}
	})

	t.Run("parses simple set enumeration and tuple expressions", func(t *testing.T) {
		root, diags := ParseSanySyntax("Exprs.tla", `---- MODULE Exprs ----
Set == {1, 2}
Tuple == <<A, B>>
====`)
		requireNoErrors(t, diags)
		defs := collectSanyDefinitions(root.GetHeirs()[2])
		if len(defs) != 2 {
			t.Fatalf("definition count = %d, want 2", len(defs))
		}
		if expr := defs[0].GetHeirs()[2]; expr.Kind.JavaName() != "N_SetEnumerate" {
			t.Fatalf("Set expression = %s, want N_SetEnumerate", expr.Kind.JavaName())
		}
		if expr := defs[1].GetHeirs()[2]; expr.Kind.JavaName() != "N_Tuple" {
			t.Fatalf("Tuple expression = %s, want N_Tuple", expr.Kind.JavaName())
		}
	})

	t.Run("parses subset and set comprehension brace forms", func(t *testing.T) {
		root, diags := ParseSanySyntax("Exprs.tla", `---- MODULE Exprs ----
Subset == {x \in S : P}
All == {x + 1 : x \in S}
====`)
		requireNoErrors(t, diags)
		defs := collectSanyDefinitions(root.GetHeirs()[2])
		if len(defs) != 2 {
			t.Fatalf("definition count = %d, want 2", len(defs))
		}
		subset := defs[0].GetHeirs()[2]
		if subset.Kind.JavaName() != "N_SubsetOf" {
			t.Fatalf("Subset expression = %s heirs %v, want N_SubsetOf", subset.Kind.JavaName(), sanyNodeKindNames(subset.GetHeirs()))
		}
		all := defs[1].GetHeirs()[2]
		if all.Kind.JavaName() != "N_SetOfAll" || countSanyChildren(all, "N_QuantBound") != 1 {
			t.Fatalf("All expression = %s heirs %v", all.Kind.JavaName(), sanyNodeKindNames(all.GetHeirs()))
		}
	})

	t.Run("parses Java semantic corpus tuple quantifier bounds", func(t *testing.T) {
		root, diags := ParseSanySyntax("TupleBounds.tla", `---- MODULE TupleBounds ----
TupleFilter == {<<x, y>> \in S : {x, y}}
TupleFcn == [<<x, y>> \in S, z \in T |-> <<x, y, z>>]
Mixed == \A x, y \in Nat, <<z>> \in Int, a, b, c \in Real : FALSE
====`)
		requireNoErrors(t, diags)
		defs := collectSanyDefinitions(root.GetHeirs()[2])
		if len(defs) != 3 {
			t.Fatalf("definition count = %d, want 3", len(defs))
		}
		filter := defs[0].GetHeirs()[2]
		if filter.Kind.JavaName() != "N_SubsetOf" || countSanyChildren(filter, "N_IdentifierTuple") != 1 {
			t.Fatalf("tuple filter expression = %s heirs %v", filter.Kind.JavaName(), sanyNodeKindNames(filter.GetHeirs()))
		}
		fcn := defs[1].GetHeirs()[2]
		if fcn.Kind.JavaName() != "N_FcnConst" || countSanyChildren(fcn, "N_QuantBound") != 2 || countSanyDescendants(fcn, "N_IdentifierTuple") != 1 {
			t.Fatalf("tuple function expression = %s heirs %v", fcn.Kind.JavaName(), sanyNodeKindNames(fcn.GetHeirs()))
		}
		mixed := defs[2].GetHeirs()[2]
		if mixed.Kind.JavaName() != "N_BoundQuant" || countSanyChildren(mixed, "N_QuantBound") != 3 || countSanyDescendants(mixed, "N_IdentifierTuple") != 1 {
			t.Fatalf("mixed quantifier expression = %s heirs %v", mixed.Kind.JavaName(), sanyNodeKindNames(mixed.GetHeirs()))
		}
	})

	t.Run("parses let case choose and lambda expressions", func(t *testing.T) {
		root, diags := ParseSanySyntax("Exprs.tla", `---- MODULE Exprs ----
L == LET A == B IN A
C == CASE A -> B [] OTHER -> C
Ch == CHOOSE x \in S : P
TupleCh == CHOOSE <<x, y>> \in S : P
UnboundTupleCh == CHOOSE <<x, y>> : P
Lam == F(LAMBDA x, y : x)
====`)
		requireNoErrors(t, diags)
		defs := collectSanyDefinitions(root.GetHeirs()[2])
		if len(defs) != 6 {
			t.Fatalf("definition count = %d, want 6", len(defs))
		}
		if expr := defs[0].GetHeirs()[2]; expr.Kind.JavaName() != "N_LetIn" {
			t.Fatalf("L expression = %s, want N_LetIn", expr.Kind.JavaName())
		}
		if expr := defs[1].GetHeirs()[2]; expr.Kind.JavaName() != "N_Case" {
			t.Fatalf("C expression = %s, want N_Case", expr.Kind.JavaName())
		}
		if expr := defs[2].GetHeirs()[2]; expr.Kind.JavaName() != "N_UnboundOrBoundChoose" {
			t.Fatalf("Ch expression = %s, want N_UnboundOrBoundChoose", expr.Kind.JavaName())
		}
		if expr := defs[3].GetHeirs()[2]; expr.Kind.JavaName() != "N_UnboundOrBoundChoose" || countSanyChildren(expr, "N_IdentifierTuple") != 1 {
			t.Fatalf("TupleCh expression = %s heirs %v, want tuple N_UnboundOrBoundChoose", expr.Kind.JavaName(), sanyNodeKindNames(expr.GetHeirs()))
		}
		if expr := defs[4].GetHeirs()[2]; expr.Kind.JavaName() != "N_UnboundOrBoundChoose" || countSanyChildren(expr, "N_IdentifierTuple") != 1 {
			t.Fatalf("UnboundTupleCh expression = %s heirs %v, want tuple N_UnboundOrBoundChoose", expr.Kind.JavaName(), sanyNodeKindNames(expr.GetHeirs()))
		}
		if expr := findSanyChild(t, findSanyChild(t, defs[5].GetHeirs()[2], "N_OpArgs"), "N_Lambda"); expr.Kind.JavaName() != "N_Lambda" {
			t.Fatalf("Lam argument = %s, want N_Lambda", expr.Kind.JavaName())
		}
	})

	t.Run("parses LET module instances inside junction lists", func(t *testing.T) {
		root, diags := ParseSanySyntax("Exprs.tla", `---- MODULE Exprs ----
Conj ==
  /\ LET M1 == INSTANCE M2 IN P
  /\ Q
Disj ==
  \/ LET M1 == INSTANCE M2 IN P
  \/ Q
====`)
		requireNoErrors(t, diags)
		defs := collectSanyDefinitions(root.GetHeirs()[2])
		if len(defs) != 2 {
			t.Fatalf("definitions = %v, want conjunction and disjunction definitions", sanyNodeKindNames(defs))
		}
		for _, def := range defs {
			if countSanyDescendants(def, "N_ModuleDefinition") != 1 {
				t.Fatalf("definition heirs = %v, want LET-local N_ModuleDefinition", sanyNodeKindNames(def.GetHeirs()))
			}
		}
	})

	t.Run("parses simple square bracket constructors", func(t *testing.T) {
		root, diags := ParseSanySyntax("Exprs.tla", `---- MODULE Exprs ----
R == [a |-> b, c |-> d]
F == [x \in S |-> x]
====`)
		requireNoErrors(t, diags)
		defs := collectSanyDefinitions(root.GetHeirs()[2])
		if len(defs) != 2 {
			t.Fatalf("definition count = %d, want 2", len(defs))
		}
		rcd := defs[0].GetHeirs()[2]
		if rcd.Kind.JavaName() != "N_RcdConstructor" || countSanyChildren(rcd, "N_FieldVal") != 2 {
			t.Fatalf("record expression = %s heirs %v", rcd.Kind.JavaName(), sanyNodeKindNames(rcd.GetHeirs()))
		}
		fcn := defs[1].GetHeirs()[2]
		if fcn.Kind.JavaName() != "N_FcnConst" || countSanyChildren(fcn, "N_QuantBound") != 1 {
			t.Fatalf("function expression = %s heirs %v", fcn.Kind.JavaName(), sanyNodeKindNames(fcn.GetHeirs()))
		}
	})

	t.Run("parses square bracket set forms", func(t *testing.T) {
		root, diags := ParseSanySyntax("Exprs.tla", `---- MODULE Exprs ----
RecSet == [a: S, b: T]
FcnSet == [S -> T]
====`)
		requireNoErrors(t, diags)
		defs := collectSanyDefinitions(root.GetHeirs()[2])
		if len(defs) != 2 {
			t.Fatalf("definition count = %d, want 2", len(defs))
		}
		recSet := defs[0].GetHeirs()[2]
		if recSet.Kind.JavaName() != "N_SetOfRcds" || countSanyChildren(recSet, "N_FieldSet") != 2 {
			t.Fatalf("record-set expression = %s heirs %v", recSet.Kind.JavaName(), sanyNodeKindNames(recSet.GetHeirs()))
		}
		fcnSet := defs[1].GetHeirs()[2]
		if fcnSet.Kind.JavaName() != "N_SetOfFcns" {
			t.Fatalf("function-set expression = %s heirs %v", fcnSet.Kind.JavaName(), sanyNodeKindNames(fcnSet.GetHeirs()))
		}
	})

	t.Run("parses function application suffixes", func(t *testing.T) {
		root, diags := ParseSanySyntax("Exprs.tla", `---- MODULE Exprs ----
App == F[x, y]
====`)
		requireNoErrors(t, diags)
		expr := onlySanyDefinition(t, root).GetHeirs()[2]
		if expr.Kind.JavaName() != "N_FcnAppl" {
			t.Fatalf("application expression = %s heirs %v, want N_FcnAppl", expr.Kind.JavaName(), sanyNodeKindNames(expr.GetHeirs()))
		}
	})

	t.Run("parses raw negative sign as Java SANY prefix minus", func(t *testing.T) {
		root, diags := ParseSanySyntax("Exprs.tla", `---- MODULE Exprs ----
A == -(N - 1)
====`)
		requireNoErrors(t, diags)
		expr := onlySanyDefinition(t, root).GetHeirs()[2]
		if expr.Kind.JavaName() != "N_PrefixExpr" {
			t.Fatalf("negative expression = %s heirs %v, want N_PrefixExpr", expr.Kind.JavaName(), sanyNodeKindNames(expr.GetHeirs()))
		}
		if got := sanyOperatorImage(expr.GetHeirs()[0]); got != "-" {
			t.Fatalf("negative operator = %q, want original source token %q", got, "-")
		}
		if got := expr.GetHeirs()[0].GetHeirs()[1].Kind.JavaName(); got != "N_InfixOp" {
			t.Fatalf("negative operator leaf = %s, want source N_InfixOp", got)
		}
		if got := expr.GetHeirs()[0].Kind.JavaName(); got != "N_GenInfixOp" {
			t.Fatalf("negative operator node = %s, want source N_GenInfixOp", got)
		}
	})

	t.Run("parses EXCEPT square bracket forms", func(t *testing.T) {
		root, diags := ParseSanySyntax("Exprs.tla", `---- MODULE Exprs ----
Except == [r EXCEPT !.a = v, ![i] = w]
ExceptCase == [r EXCEPT !.a = CASE x -> 0 [] OTHER -> @ - 1, !.b = @ + 1]
====`)
		requireNoErrors(t, diags)
		defs := collectSanyDefinitions(root.GetHeirs()[2])
		if len(defs) != 2 {
			t.Fatalf("definition count = %d, want 2", len(defs))
		}
		expr := defs[0].GetHeirs()[2]
		if expr.Kind.JavaName() != "N_Except" || countSanyChildren(expr, "N_ExceptSpec") != 2 {
			t.Fatalf("EXCEPT expression = %s heirs %v", expr.Kind.JavaName(), sanyNodeKindNames(expr.GetHeirs()))
		}
		caseExpr := defs[1].GetHeirs()[2]
		if caseExpr.Kind.JavaName() != "N_Except" || countSanyChildren(caseExpr, "N_ExceptSpec") != 2 || countSanyDescendants(caseExpr, "N_Case") != 1 {
			t.Fatalf("EXCEPT CASE expression = %s heirs %v", caseExpr.Kind.JavaName(), sanyNodeKindNames(caseExpr.GetHeirs()))
		}
	})

	t.Run("parses bracket action expressions", func(t *testing.T) {
		root, diags := ParseSanySyntax("Exprs.tla", `---- MODULE Exprs ----
Square == [A]_x
Tuple == <<A>>_x
====`)
		requireNoErrors(t, diags)
		defs := collectSanyDefinitions(root.GetHeirs()[2])
		if len(defs) != 2 {
			t.Fatalf("definition count = %d, want 2", len(defs))
		}
		if expr := defs[0].GetHeirs()[2]; expr.Kind.JavaName() != "N_ActionExpr" {
			t.Fatalf("Square expression = %s heirs %v, want N_ActionExpr", expr.Kind.JavaName(), sanyNodeKindNames(expr.GetHeirs()))
		}
		if expr := defs[1].GetHeirs()[2]; expr.Kind.JavaName() != "N_ActionExpr" {
			t.Fatalf("Tuple expression = %s heirs %v, want N_ActionExpr", expr.Kind.JavaName(), sanyNodeKindNames(expr.GetHeirs()))
		}
	})

	t.Run("parses temporal boxes over square actions before function constructors", func(t *testing.T) {
		_, diags := ParseSanySyntax("Exprs.tla", `---- MODULE Exprs ----
Spec == Init /\ [][Next]_vars
Fcn == [i \in S |-> i]
====`)
		requireNoErrors(t, diags)
	})

	t.Run("parses record component expressions", func(t *testing.T) {
		root, diags := ParseSanySyntax("Exprs.tla", `---- MODULE Exprs ----
Field == r.a
====`)
		requireNoErrors(t, diags)
		expr := onlySanyDefinition(t, root).GetHeirs()[2]
		if expr.Kind.JavaName() != "N_RecordComponent" {
			t.Fatalf("record component expression = %s heirs %v, want N_RecordComponent", expr.Kind.JavaName(), sanyNodeKindNames(expr.GetHeirs()))
		}
	})

	t.Run("parses bang qualified general identifiers", func(t *testing.T) {
		root, diags := ParseSanySyntax("Exprs.tla", `---- MODULE Exprs ----
Qual == M!Op
====`)
		requireNoErrors(t, diags)
		expr := onlySanyDefinition(t, root).GetHeirs()[2]
		if expr.Kind.JavaName() != "N_GeneralId" || countSanyChildren(expr.GetHeirs()[0], "N_IdPrefixElement") != 1 {
			t.Fatalf("qualified expression = %s heirs %v, want N_GeneralId with prefix", expr.Kind.JavaName(), sanyNodeKindNames(expr.GetHeirs()))
		}
	})

	t.Run("parses bang structural selectors", func(t *testing.T) {
		root, diags := ParseSanySyntax("Exprs.tla", `---- MODULE Exprs ----
TupleOpen == M!<<
TupleClose == M!>>
TupleIndex == M!2
ExceptAt == M!@
ColonThenField == M!:!Field
====`)
		requireNoErrors(t, diags)
		defs := collectSanyDefinitions(root.GetHeirs()[2])
		if len(defs) != 5 {
			t.Fatalf("definitions = %v, want five structural selector definitions", sanyNodeKindNames(defs))
		}

		for i, wantImage := range []string{"<<", ">>", "2", "@"} {
			expr := defs[i].GetHeirs()[2]
			if expr.Kind.JavaName() != "N_GeneralId" {
				t.Fatalf("definition %d expression = %s, want N_GeneralId", i, expr.Kind.JavaName())
			}
			selector := expr.GetHeirs()[1]
			if selector.Kind.JavaName() != "N_StructOp" || sanyFirstTokenImage(selector) != wantImage {
				t.Fatalf("definition %d selector = %s %q, want N_StructOp %q", i, selector.Kind.JavaName(), sanyFirstTokenImage(selector), wantImage)
			}
		}

		expr := defs[4].GetHeirs()[2]
		prefix := expr.GetHeirs()[0]
		prefixElements := collectSanyChildrenByKind(prefix, "N_IdPrefixElement")
		if len(prefixElements) != 2 {
			t.Fatalf("ColonThenField prefix elements = %d, want module selector and structural selector", len(prefixElements))
		}
		if selector := prefixElements[1].GetHeirs()[0]; selector.Kind.JavaName() != "N_StructOp" || sanyFirstTokenImage(selector) != ":" {
			t.Fatalf("ColonThenField structural prefix = %s %q, want N_StructOp ':'", selector.Kind.JavaName(), sanyFirstTokenImage(selector))
		}
		if final := expr.GetHeirs()[1]; final.Kind.JavaName() != "IDENTIFIER" || final.Image != "Field" {
			t.Fatalf("ColonThenField final selector = %s %q, want identifier Field", final.Kind.JavaName(), final.Image)
		}
	})

	t.Run("parses bang operator selectors", func(t *testing.T) {
		root, diags := ParseSanySyntax("Exprs.tla", `---- MODULE Exprs ----
InfixSelector == M!\oplus
InfixCall == M!\oplus(1, 2)
PrefixThenField == M!DOMAIN!Field
PostfixSelector == M!^+
====`)
		requireNoErrors(t, diags)
		defs := collectSanyDefinitions(root.GetHeirs()[2])
		if len(defs) != 4 {
			t.Fatalf("definitions = %v, want four operator selector definitions", sanyNodeKindNames(defs))
		}

		infix := defs[0].GetHeirs()[2].GetHeirs()[1]
		if infix.Kind.JavaName() != "N_InfixOp" || sanyOperatorImage(infix) != `\oplus` {
			t.Fatalf("infix selector = %s %q, want N_InfixOp \\oplus", infix.Kind.JavaName(), sanyOperatorImage(infix))
		}

		call := defs[1].GetHeirs()[2]
		if call.Kind.JavaName() != "N_OpApplication" || countSanyChildren(call, "N_OpArgs") != 1 {
			t.Fatalf("infix selector call = %s heirs %v, want N_OpApplication with N_OpArgs", call.Kind.JavaName(), sanyNodeKindNames(call.GetHeirs()))
		}
		callSelector := call.GetHeirs()[0].GetHeirs()[1]
		if callSelector.Kind.JavaName() != "N_InfixOp" || sanyOperatorImage(callSelector) != `\oplus` {
			t.Fatalf("call selector = %s %q, want N_InfixOp \\oplus", callSelector.Kind.JavaName(), sanyOperatorImage(callSelector))
		}

		prefixExpr := defs[2].GetHeirs()[2]
		prefixElements := collectSanyChildrenByKind(prefixExpr.GetHeirs()[0], "N_IdPrefixElement")
		if len(prefixElements) != 2 {
			t.Fatalf("PrefixThenField prefix elements = %d, want module and prefix-operator selectors", len(prefixElements))
		}
		if selector := prefixElements[1].GetHeirs()[0]; selector.Kind.JavaName() != "N_NonExpPrefixOp" || sanyOperatorImage(selector) != "DOMAIN" {
			t.Fatalf("prefix selector = %s %q, want N_NonExpPrefixOp DOMAIN", selector.Kind.JavaName(), sanyOperatorImage(selector))
		}

		postfix := defs[3].GetHeirs()[2].GetHeirs()[1]
		if postfix.Kind.JavaName() != "N_PostfixOp" || sanyOperatorImage(postfix) != "^+" {
			t.Fatalf("postfix selector = %s %q, want N_PostfixOp ^+", postfix.Kind.JavaName(), sanyOperatorImage(postfix))
		}
	})

	t.Run("parses bang selector chains with intermediate operator arguments", func(t *testing.T) {
		root, diags := ParseSanySyntax("Exprs.tla", `---- MODULE Exprs ----
Nested == Foo(1)!Bar!Baz(2, 3)
MiddleArgs == M!F(1)!G
====`)
		requireNoErrors(t, diags)
		defs := collectSanyDefinitions(root.GetHeirs()[2])
		if len(defs) != 2 {
			t.Fatalf("definitions = %v, want nested selector definitions", sanyNodeKindNames(defs))
		}

		nested := defs[0].GetHeirs()[2]
		if nested.Kind.JavaName() != "N_OpApplication" {
			t.Fatalf("Nested expression = %s, want final N_OpApplication", nested.Kind.JavaName())
		}
		nestedPrefix := nested.GetHeirs()[0].GetHeirs()[0]
		nestedElements := collectSanyChildrenByKind(nestedPrefix, "N_IdPrefixElement")
		if len(nestedElements) != 2 {
			t.Fatalf("Nested prefix elements = %d, want Foo(1)! and Bar!", len(nestedElements))
		}
		if first := nestedElements[0]; countSanyChildren(first, "N_OpArgs") != 1 {
			t.Fatalf("Nested first prefix heirs = %v, want Foo selector with N_OpArgs before bang", sanyNodeKindNames(first.GetHeirs()))
		}
		if final := nested.GetHeirs()[0].GetHeirs()[1]; final.Kind.JavaName() != "IDENTIFIER" || final.Image != "Baz" {
			t.Fatalf("Nested final selector = %s %q, want Baz", final.Kind.JavaName(), final.Image)
		}

		middle := defs[1].GetHeirs()[2]
		if middle.Kind.JavaName() != "N_GeneralId" {
			t.Fatalf("MiddleArgs expression = %s, want N_GeneralId", middle.Kind.JavaName())
		}
		middleElements := collectSanyChildrenByKind(middle.GetHeirs()[0], "N_IdPrefixElement")
		if len(middleElements) != 2 || countSanyChildren(middleElements[1], "N_OpArgs") != 1 {
			t.Fatalf("MiddleArgs prefix elements = %v, want M! and F(1)!", sanyNodeKindNames(middleElements))
		}
		if final := middle.GetHeirs()[1]; final.Kind.JavaName() != "IDENTIFIER" || final.Image != "G" {
			t.Fatalf("MiddleArgs final selector = %s %q, want G", final.Kind.JavaName(), final.Image)
		}
	})

	t.Run("parses parameterized module refs in bounded quantifier bodies", func(t *testing.T) {
		root, diags := ParseSanySyntax("Exprs.tla", `---- MODULE Exprs ----
CONSTANT SuccSet
P(Succ) == INSTANCE Proofs
Test == <<\A Succ \in SuccSet : P(Succ)!Reachable0>>
====`)
		requireNoErrors(t, diags)
		body := root.GetHeirs()[2]
		requireSanyNodeKinds(t, body.GetHeirs(), "N_ParamDeclaration", "N_ModuleDefinition", "N_OperatorDefinition")
		moduleDef := body.GetHeirs()[1]
		lhs := moduleDef.GetHeirs()[0]
		if lhs.Kind.JavaName() != "N_IdentLHS" || countSanyChildren(lhs, "N_IdentDecl") != 1 {
			t.Fatalf("module definition lhs = %s heirs %v, want parameterized N_IdentLHS", lhs.Kind.JavaName(), sanyNodeKindNames(lhs.GetHeirs()))
		}
		tuple := body.GetHeirs()[2].GetHeirs()[2]
		genIDs := collectSanyDescendantsByKind(tuple, "N_GeneralId")
		var found bool
		for _, genID := range genIDs {
			if sanyGeneralIDName(genID) == "P!Reachable0" {
				found = true
			}
		}
		if !found {
			t.Fatalf("Test expression did not contain P(Succ)!Reachable0 general id; descendants=%v", sanyNodeKindNames(genIDs))
		}
	})

	t.Run("parses bang argument-only selectors", func(t *testing.T) {
		root, diags := ParseSanySyntax("Exprs.tla", `---- MODULE Exprs ----
ArgsOnly == M!(1, 2)
ArgsThenField == M!(1)!Field
====`)
		requireNoErrors(t, diags)
		defs := collectSanyDefinitions(root.GetHeirs()[2])
		if len(defs) != 2 {
			t.Fatalf("definitions = %v, want argument-only selector definitions", sanyNodeKindNames(defs))
		}

		argsOnly := defs[0].GetHeirs()[2]
		if argsOnly.Kind.JavaName() != "N_GeneralId" || argsOnly.GetHeirs()[1].Kind.JavaName() != "N_OpArgs" {
			t.Fatalf("ArgsOnly expression = %s heirs %v, want N_GeneralId ending in N_OpArgs", argsOnly.Kind.JavaName(), sanyNodeKindNames(argsOnly.GetHeirs()))
		}

		argsThenField := defs[1].GetHeirs()[2]
		prefixElements := collectSanyChildrenByKind(argsThenField.GetHeirs()[0], "N_IdPrefixElement")
		if len(prefixElements) != 2 {
			t.Fatalf("ArgsThenField prefix elements = %d, want M! and !(1)!", len(prefixElements))
		}
		if selector := prefixElements[1].GetHeirs()[0]; selector.Kind.JavaName() != "N_OpArgs" {
			t.Fatalf("ArgsThenField second selector = %s, want N_OpArgs", selector.Kind.JavaName())
		}
		if final := argsThenField.GetHeirs()[1]; final.Kind.JavaName() != "IDENTIFIER" || final.Image != "Field" {
			t.Fatalf("ArgsThenField final selector = %s %q, want Field", final.Kind.JavaName(), final.Image)
		}
	})

	t.Run("parses operators as higher-order expression arguments", func(t *testing.T) {
		root, diags := ParseSanySyntax("Exprs.tla", `---- MODULE Exprs ----
Refs == f(+, ^+, ~, \lnot, UNION, /\, \/)
ParenRefs == f((-), (.))
CallRef == \o(1, 2)
====`)
		requireNoErrors(t, diags)
		defs := collectSanyDefinitions(root.GetHeirs()[2])
		if len(defs) != 3 {
			t.Fatalf("definitions = %v, want operator-reference definitions", sanyNodeKindNames(defs))
		}
		refs := defs[0].GetHeirs()[2]
		if refs.Kind.JavaName() != "N_OpApplication" {
			t.Fatalf("Refs expression = %s, want N_OpApplication", refs.Kind.JavaName())
		}
		if countSanyDescendants(refs, "N_GenInfixOp") < 3 || countSanyDescendants(refs, "N_GenNonExpPrefixOp") < 3 || countSanyDescendants(refs, "N_GenPostfixOp") < 1 {
			t.Fatalf("Refs operator nodes = %v, want infix, non-expression prefix, and postfix operator references", sanyNodeKindNames(refs.GetHeirs()))
		}
		callRef := defs[2].GetHeirs()[2]
		if callRef.Kind.JavaName() != "N_OpApplication" || countSanyChildren(callRef, "N_OpArgs") != 1 {
			t.Fatalf("CallRef expression = %s heirs %v, want N_OpApplication with N_OpArgs", callRef.Kind.JavaName(), sanyNodeKindNames(callRef.GetHeirs()))
		}
		genID := findSanyChild(t, callRef, "N_GeneralId")
		requireSanyNodeKinds(t, genID.GetHeirs(), "N_IdPrefix", "N_InfixOp")
	})

	t.Run("parses labeled expressions", func(t *testing.T) {
		root, diags := ParseSanySyntax("Exprs.tla", `---- MODULE Exprs ----
Labeled == Lbl:: TRUE
Parameterized == Lbl(a, b):: P(a, b)
Precedence == a + Lab:: b * c
====`)
		requireNoErrors(t, diags)
		defs := collectSanyDefinitions(root.GetHeirs()[2])
		if len(defs) != 3 {
			t.Fatalf("definitions = %v, want three labeled definitions", sanyNodeKindNames(defs))
		}
		if expr := defs[0].GetHeirs()[2]; expr.Kind.JavaName() != "N_Label" {
			t.Fatalf("labeled expression = %s heirs %v, want N_Label", expr.Kind.JavaName(), sanyNodeKindNames(expr.GetHeirs()))
		}
		if expr := defs[1].GetHeirs()[2]; expr.Kind.JavaName() != "N_Label" || countSanyChildren(findSanyChild(t, expr, "N_GeneralId"), "N_OpArgs") != 1 {
			t.Fatalf("parameterized label = %s heirs %v, want N_Label with N_OpArgs", expr.Kind.JavaName(), sanyNodeKindNames(expr.GetHeirs()))
		}
		if countSanyDescendants(defs[2], "N_Label") != 1 {
			t.Fatalf("precedence definition heirs = %v, want embedded N_Label", sanyNodeKindNames(defs[2].GetHeirs()))
		}
	})

	t.Run("parses simple operator applications", func(t *testing.T) {
		root, diags := ParseSanySyntax("Exprs.tla", `---- MODULE Exprs ----
Call == F(a, b)
====`)
		requireNoErrors(t, diags)
		expr := onlySanyDefinition(t, root).GetHeirs()[2]
		if expr.Kind.JavaName() != "N_OpApplication" || countSanyChildren(expr, "N_OpArgs") != 1 {
			t.Fatalf("operator application = %s heirs %v, want N_OpApplication with N_OpArgs", expr.Kind.JavaName(), sanyNodeKindNames(expr.GetHeirs()))
		}
	})

	t.Run("parses qualified operator applications", func(t *testing.T) {
		root, diags := ParseSanySyntax("Exprs.tla", `---- MODULE Exprs ----
QualCall == M!F(a)
====`)
		requireNoErrors(t, diags)
		expr := onlySanyDefinition(t, root).GetHeirs()[2]
		if expr.Kind.JavaName() != "N_OpApplication" || countSanyChildren(expr, "N_OpArgs") != 1 {
			t.Fatalf("qualified operator application = %s heirs %v, want N_OpApplication with N_OpArgs", expr.Kind.JavaName(), sanyNodeKindNames(expr.GetHeirs()))
		}
	})

	t.Run("parses fairness expressions", func(t *testing.T) {
		root, diags := ParseSanySyntax("Exprs.tla", `---- MODULE Exprs ----
Weak == WF_<<x>>(A)
Strong == SF_<<x>>(A)
====`)
		requireNoErrors(t, diags)
		defs := collectSanyDefinitions(root.GetHeirs()[2])
		if len(defs) != 2 {
			t.Fatalf("definition count = %d, want 2", len(defs))
		}
		if expr := defs[0].GetHeirs()[2]; expr.Kind.JavaName() != "N_FairnessExpr" {
			t.Fatalf("weak fairness expression = %s heirs %v, want N_FairnessExpr", expr.Kind.JavaName(), sanyNodeKindNames(expr.GetHeirs()))
		}
		if expr := defs[1].GetHeirs()[2]; expr.Kind.JavaName() != "N_FairnessExpr" {
			t.Fatalf("strong fairness expression = %s heirs %v, want N_FairnessExpr", expr.Kind.JavaName(), sanyNodeKindNames(expr.GetHeirs()))
		}
	})

	t.Run("parses temporal prefix operators", func(t *testing.T) {
		root, diags := ParseSanySyntax("Exprs.tla", `---- MODULE Exprs ----
Enabled == ENABLED A
Eventually == <>A
====`)
		requireNoErrors(t, diags)
		defs := collectSanyDefinitions(root.GetHeirs()[2])
		if len(defs) != 2 {
			t.Fatalf("definition count = %d, want 2", len(defs))
		}
		for _, def := range defs {
			if expr := def.GetHeirs()[2]; expr.Kind.JavaName() != "N_PrefixExpr" {
				t.Fatalf("temporal prefix expression = %s heirs %v, want N_PrefixExpr", expr.Kind.JavaName(), sanyNodeKindNames(expr.GetHeirs()))
			}
		}
	})
}

func requireSanyNodeKinds(t *testing.T, nodes []*SanySyntaxNode, want ...string) {
	t.Helper()
	got := sanyNodeKindNames(nodes)
	if len(got) != len(want) {
		t.Fatalf("node kinds = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("node kinds = %v, want %v", got, want)
		}
	}
}

func countSanyChildren(node *SanySyntaxNode, kind string) int {
	count := 0
	for _, child := range node.GetHeirs() {
		if child.Kind.JavaName() == kind {
			count++
		}
	}
	return count
}

func countSanyDescendants(node *SanySyntaxNode, kind string) int {
	count := 0
	var visit func(*SanySyntaxNode)
	visit = func(cur *SanySyntaxNode) {
		if cur == nil {
			return
		}
		if cur.Kind.JavaName() == kind {
			count++
		}
		for _, child := range cur.GetHeirs() {
			visit(child)
		}
	}
	visit(node)
	return count
}

func collectSanyChildrenByKind(node *SanySyntaxNode, kind string) []*SanySyntaxNode {
	var found []*SanySyntaxNode
	for _, child := range node.GetHeirs() {
		if child.Kind.JavaName() == kind {
			found = append(found, child)
		}
	}
	return found
}

func collectSanyDescendantsByKind(node *SanySyntaxNode, kind string) []*SanySyntaxNode {
	var found []*SanySyntaxNode
	var visit func(*SanySyntaxNode)
	visit = func(cur *SanySyntaxNode) {
		if cur == nil {
			return
		}
		if cur.Kind.JavaName() == kind {
			found = append(found, cur)
		}
		for _, child := range cur.GetHeirs() {
			visit(child)
		}
	}
	visit(node)
	return found
}

func findSanyChild(t *testing.T, node *SanySyntaxNode, kind string) *SanySyntaxNode {
	t.Helper()
	for _, child := range node.GetHeirs() {
		if child.Kind.JavaName() == kind {
			return child
		}
	}
	t.Fatalf("child %s not found in %v", kind, sanyNodeKindNames(node.GetHeirs()))
	return nil
}

func requireSanyProofStepBodies(t *testing.T, steps []*SanySyntaxNode, want ...string) {
	t.Helper()
	if len(steps) != len(want) {
		t.Fatalf("proof steps = %d, want %d", len(steps), len(want))
	}
	for i, step := range steps {
		if len(step.GetHeirs()) < 2 {
			t.Fatalf("proof step %d heirs = %v, want body %s", i, sanyNodeKindNames(step.GetHeirs()), want[i])
		}
		if got := step.GetHeirs()[1].Kind.JavaName(); got != want[i] {
			t.Fatalf("proof step %d body = %s heirs %v, want %s", i, got, sanyNodeKindNames(step.GetHeirs()), want[i])
		}
	}
}

func onlySanyDefinition(t *testing.T, root *SanySyntaxNode) *SanySyntaxNode {
	t.Helper()
	body := root.GetHeirs()[2]
	defs := collectSanyDefinitions(body)
	if len(defs) != 1 {
		t.Fatalf("definitions = %v, want exactly one", sanyNodeKindNames(defs))
	}
	return defs[0]
}

func collectSanyDefinitions(body *SanySyntaxNode) []*SanySyntaxNode {
	var defs []*SanySyntaxNode
	for _, child := range body.GetHeirs() {
		switch child.Kind.JavaName() {
		case "N_OperatorDefinition", "N_FunctionDefinition":
			defs = append(defs, child)
		}
	}
	return defs
}

func sanyNodeKindNames(nodes []*SanySyntaxNode) []string {
	names := make([]string, 0, len(nodes))
	for _, node := range nodes {
		if node == nil {
			names = append(names, "<nil>")
			continue
		}
		names = append(names, node.Kind.JavaName())
	}
	return names
}
