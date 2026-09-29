package tlago

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParserBehaviors(t *testing.T) {
	t.Run("accepts a minimal module", func(t *testing.T) {
		mod, diags := ParseSanyModuleSource("Test.tla", "---- MODULE Test ----\n====")
		requireNoErrors(t, diags)
		if mod.Name != "Test" {
			t.Fatalf("module name = %q, want Test", mod.Name)
		}
	})

	t.Run("ignores pre-module text and nested comments", func(t *testing.T) {
		source := `this is toolbox prose before the module
---- MODULE Test ----
(* outer (* inner *) still comment *)
VARIABLE x
Init == x = 0
====`
		mod, diags := ParseSanyModuleSource("Test.tla", source)
		requireNoErrors(t, diags)
		if got := len(mod.Declarations); got != 1 {
			t.Fatalf("declarations = %d, want 1", got)
		}
		if got := len(mod.Definitions); got != 1 {
			t.Fatalf("definitions = %d, want 1", got)
		}
	})

	t.Run("records extends declarations assumptions and definitions", func(t *testing.T) {
		source := `---- MODULE Counter ----
EXTENDS Naturals, TLC
CONSTANT Max
VARIABLE x
ASSUME Max > 0
Init == x = 0
Next == x' = x + 1
====`
		mod, diags := ParseSanyModuleSource("Counter.tla", source)
		requireNoErrors(t, diags)
		if strings.Join(mod.Extends, ",") != "Naturals,TLC" {
			t.Fatalf("extends = %v", mod.Extends)
		}
		if len(mod.Declarations) != 2 || mod.Declarations[0].Names[0] != "Max" || mod.Declarations[1].Names[0] != "x" {
			t.Fatalf("declarations = %#v", mod.Declarations)
		}
		if len(mod.Assumptions) != 1 {
			t.Fatalf("assumptions = %d, want 1", len(mod.Assumptions))
		}
		if len(mod.Definitions) != 2 {
			t.Fatalf("definitions = %d, want 2", len(mod.Definitions))
		}
	})

	t.Run("uses SANY-style operator precedence", func(t *testing.T) {
		source := `---- MODULE Test ----
ASSUME A /\ B => C \/ D
====`
		mod, diags := ParseSanyModuleSource("Test.tla", source)
		requireNoErrors(t, diags)
		root := asBinary(t, mod.Assumptions[0].Expr, "=>")
		asBinary(t, root.Left, "/\\")
		asBinary(t, root.Right, "\\/")
	})

	t.Run("rejects non-associative operator chains", func(t *testing.T) {
		_, diags := ParseSanyModuleSource("Test.tla", `---- MODULE Test ----
ASSUME A = B = C
====`)
		requireHasErrorContaining(t, diags, "precedence conflict")
	})
}

func TestResolverAndCheckerBehaviors(t *testing.T) {
	t.Run("loads sibling modules and accepts standard modules", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Root.tla"), `---- MODULE Root ----
EXTENDS Naturals, Helper
VARIABLE x
Init == x = Helper!Zero
====`)
		writeFile(t, filepath.Join(dir, "Helper.tla"), `---- MODULE Helper ----
Zero == 0
====`)

		spec, diags := LoadSanySpec(filepath.Join(dir, "Root.tla"), LoadOptions{})
		requireNoErrors(t, diags)
		if spec.Modules["Root"] == nil || spec.Modules["Helper"] == nil {
			t.Fatalf("loaded modules = %v", moduleNames(spec.Modules))
		}
	})

	t.Run("loads the TLAPS proof helper standard module", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Root.tla"), `---- MODULE Root ----
EXTENDS TLAPS
UseBackendFacts == SMT /\ SMTT(30) /\ PTL
====`)

		spec, diags := LoadSanySpec(filepath.Join(dir, "Root.tla"), LoadOptions{})
		requireNoErrors(t, diags)
		if spec.Modules["TLAPS"] == nil {
			t.Fatalf("loaded modules = %v, want TLAPS", moduleNames(spec.Modules))
		}
	})

	t.Run("can prefer library modules over embedded standard modules", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "Root.tla")
		writeFile(t, root, `---- MODULE Root ----
EXTENDS TLAPS
UseLocalProofLib == OnlyHere
====`)
		writeFile(t, filepath.Join(dir, "TLAPS.tla"), `---- MODULE TLAPS ----
OnlyHere == TRUE
====`)

		spec, diags := LoadSanySpec(root, LoadOptions{LibraryPaths: []string{dir}, PreferLibraryModules: true})
		requireNoErrors(t, diags)
		if spec.Modules["TLAPS"] == nil {
			t.Fatalf("loaded modules = %v, want TLAPS", moduleNames(spec.Modules))
		}
		diags = CheckSpec(spec)
		requireNoErrors(t, diags)
	})

	t.Run("accepts theorem proof rules with temporal and action assume prove obligations", func(t *testing.T) {
		_, diags := CheckSanySource("ProofRule.tla", `---- MODULE ProofRule ----
THEOREM RuleTLA1 == ASSUME STATE P, STATE f,
                           P /\ (f' = f) => P'
                    PROVE  []P <=> P /\ [][P => P']_f
====`)
		requireNoErrors(t, diags)
	})

	t.Run("loads TLC runtime helper standard operators", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Root.tla"), `---- MODULE Root ----
EXTENDS TLC
Choice == RandomElement({1, 2})
Symmetry == Permutations({"a", "b"})
Register == TLCSet("level", TLCGet("level") + 1)
Text == ToString(Choice)
Eval == TLCEval(Choice)
Time == JavaTime
====`)

		spec, diags := LoadSanySpec(filepath.Join(dir, "Root.tla"), LoadOptions{})
		requireNoErrors(t, diags)
		if spec.Modules["TLC"] == nil {
			t.Fatalf("loaded modules = %v, want TLC", moduleNames(spec.Modules))
		}
		diags = CheckSpec(spec)
		requireNoErrors(t, diags)
	})

	t.Run("loads common function and proof helper standard modules", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Root.tla"), `---- MODULE Root ----
EXTENDS Naturals, Functions, FunctionTheorems, FiniteSetTheorems, SequenceTheorems, NaturalsInduction,
        SequencesExt, SequencesExtTheorems, FiniteSetsExt, FiniteSetsExtTheorems,
        WellFoundedInduction, GraphTheorems, DyadicRationals, Apalache, TLAPS
CONSTANT f, S
Folded == FoldFunction(+, Zero, f)
FoldedOnSet == FoldFunctionOnSet(+, Zero, f, S)
Summed == SumFunction(f) + SumFunctionOnSet(f, S)
SeqHelpers == SetToSeq(S) \in BoundedSeq(S, One) /\ IsPrefix(<<>>, SetToSeq(S))
FiniteHelpers == RandomSubset(One, S)
DyadicHelper == PrettyPrint(Add(One, Half(One))) \o PrettyPrint(IsDyadicRational(Half(One)))
ApalacheHelper == FunAsSeq(f, One, One)
Facts == Folded \in Nat /\ IsFiniteSet(S)
THEOREM UsesProofLibs == Facts
  BY SumFunctionNat, FoldFunctionOnSetType, Induction, SimpleArithmetic DEF Facts
====`)

		spec, diags := LoadSanySpec(filepath.Join(dir, "Root.tla"), LoadOptions{})
		requireNoErrors(t, diags)
		for _, name := range []string{
			"Functions", "FunctionTheorems", "FiniteSetTheorems", "SequenceTheorems", "NaturalsInduction",
			"SequencesExt", "SequencesExtTheorems", "FiniteSetsExt", "FiniteSetsExtTheorems",
			"WellFoundedInduction", "GraphTheorems", "DyadicRationals", "Apalache",
		} {
			if spec.Modules[name] == nil {
				t.Fatalf("loaded modules = %v, want %s", moduleNames(spec.Modules), name)
			}
		}
		diags = CheckSpec(spec)
		requireNoErrors(t, diags)
	})

	t.Run("imports standard symbols through extended modules", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Root.tla"), `---- MODULE Root ----
EXTENDS Wrapper
AssumeNat == Nat \in Nat
====`)
		writeFile(t, filepath.Join(dir, "Wrapper.tla"), `---- MODULE Wrapper ----
EXTENDS Naturals
====`)

		spec, diags := LoadSanySpec(filepath.Join(dir, "Root.tla"), LoadOptions{})
		requireNoErrors(t, diags)
		diags = CheckSpec(spec)
		requireNoErrors(t, diags)
	})

	t.Run("reports file and module name mismatches", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "WrongFile.tla")
		writeFile(t, path, "---- MODULE Actual ----\n====")
		_, diags := LoadSanySpec(path, LoadOptions{})
		requireHasErrorContaining(t, diags, "does not match")
	})

	t.Run("semantic check catches duplicates and undefined identifiers", func(t *testing.T) {
		_, diags := CheckSanySource("Bad.tla", `---- MODULE Bad ----
VARIABLE x, x
Init == y = 0
====`)
		requireHasErrorContaining(t, diags, "duplicate")
		requireHasErrorContaining(t, diags, "undefined")
	})

	t.Run("semantic check rejects primed constants", func(t *testing.T) {
		_, diags := CheckSanySource("BadPrime.tla", `---- MODULE BadPrime ----
CONSTANT C
VARIABLE x
Bad == C' = x
====`)
		requireHasErrorContaining(t, diags, "cannot prime constant")
	})

	t.Run("INSTANCE operator substitutions accept bare local operator replacements", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Base.tla"), `---- MODULE Base ----
CONSTANT Op(_)
Use == Op(1)
====`)
		writeFile(t, filepath.Join(dir, "Root.tla"), `---- MODULE Root ----
LocalOp(x) == x = x
INSTANCE Base WITH Op <- LocalOp
Check == Use
====`)

		spec, diags := LoadSanySpec(filepath.Join(dir, "Root.tla"), LoadOptions{})
		requireNoErrors(t, diags)
		diags = CheckSpec(spec)
		requireNoErrors(t, diags)
	})
}

func TestCLIBehaviors(t *testing.T) {
	t.Run("prints supported commands in usage", func(t *testing.T) {
		var stderr bytes.Buffer
		if code := RunCLI(nil, nil, &stderr); code != ExitToolFailure {
			t.Fatalf("usage exit = %d, want %d", code, ExitToolFailure)
		}
		text := stderr.String()
		for _, command := range []string{"parse", "check", "modelcheck", "apalache-json", "sany-xml"} {
			if !strings.Contains(text, command) {
				t.Fatalf("usage %q does not mention %s", text, command)
			}
		}
		for _, removed := range []string{"sany ", "sany-parse", "bootstrap"} {
			if strings.Contains(text, removed) {
				t.Fatalf("usage %q mentions removed parser alias %q", text, removed)
			}
		}
	})

	dir := t.TempDir()
	good := filepath.Join(dir, "Good.tla")
	bad := filepath.Join(dir, "Bad.tla")
	warn := filepath.Join(dir, "Warn.tla")
	counter := filepath.Join(dir, "Counter.tla")
	cfg := filepath.Join(dir, "MC.cfg")
	unbounded := filepath.Join(dir, "Unbounded.tla")
	unboundedCfg := filepath.Join(dir, "Unbounded.cfg")
	traceFail := filepath.Join(dir, "TraceFail.tla")
	traceFailCfg := filepath.Join(dir, "TraceFail.cfg")
	writeFile(t, good, "---- MODULE Good ----\nVARIABLE x\nInit == x = 0\n====")
	writeFile(t, bad, "---- MODULE Bad ----\nInit == missing\n====")
	writeFile(t, warn, "---- MODULE Warn ----\nFoo == TRUE\nSomeRecord == [Foo |-> 42]\n====")
	writeFile(t, counter, counterSpec("x <= 3"))
	writeFile(t, cfg, counterCfg())
	writeFile(t, unbounded, `---- MODULE Unbounded ----
VARIABLE x
Init == x = 0
Next == x' = x + 1
Inv == TRUE
====`)
	writeFile(t, unboundedCfg, "INIT Init\nNEXT Next\nINVARIANT Inv\n")
	writeFile(t, traceFail, `---- MODULE TraceFail ----
VARIABLE y, x
Init == x = 1 /\ y = 2
Next == x' = x /\ y' = y
Inv == x = 0
====`)
	writeFile(t, traceFailCfg, "INIT Init\nNEXT Next\nINVARIANT Inv\n")

	if code := RunCLI([]string{"parse", good}, nil, nil); code != ExitOK {
		t.Fatalf("parse exit = %d, want %d", code, ExitOK)
	}
	sanyOnly := filepath.Join(dir, "SanyOnly.tla")
	writeFile(t, sanyOnly, "---- MODULE SanyOnly ----\nR == [a |-> b]\n====")
	if code := RunCLI([]string{"parse", sanyOnly}, nil, nil); code != ExitOK {
		t.Fatalf("parse SANY-only exit = %d, want %d", code, ExitOK)
	}
	if code := RunCLI([]string{"sany", good}, nil, nil); code != ExitToolFailure {
		t.Fatalf("sany alias exit = %d, want %d", code, ExitToolFailure)
	}
	if code := RunCLI([]string{"check", bad}, nil, nil); code != ExitSemanticFailure {
		t.Fatalf("check exit = %d, want %d", code, ExitSemanticFailure)
	}
	stderr := bytes.Buffer{}
	if code := RunCLI([]string{"check", warn}, nil, &stderr); code != ExitOK {
		t.Fatalf("warning check exit = %d, want %d; stderr=%s", code, ExitOK, stderr.String())
	}
	if !strings.Contains(stderr.String(), "W4802") {
		t.Fatalf("warning check stderr = %q, want W4802", stderr.String())
	}
	stderr.Reset()
	if code := RunCLI([]string{"check", "-suppressMessages", "4802", warn}, nil, &stderr); code != ExitOK {
		t.Fatalf("suppressed warning exit = %d, want %d; stderr=%s", code, ExitOK, stderr.String())
	}
	if strings.Contains(stderr.String(), "W4802") {
		t.Fatalf("suppressed warning stderr = %q, want no W4802", stderr.String())
	}
	stderr.Reset()
	if code := RunCLI([]string{"check", "-messagesAsErrors", "4802", warn}, nil, &stderr); code != ExitSemanticFailure {
		t.Fatalf("elevated warning exit = %d, want %d; stderr=%s", code, ExitSemanticFailure, stderr.String())
	}
	if !strings.Contains(stderr.String(), "Warning treated as error") {
		t.Fatalf("elevated warning stderr = %q, want elevation message", stderr.String())
	}
	if code := RunCLI([]string{"modelcheck", "-config", cfg, counter}, nil, nil); code != ExitOK {
		t.Fatalf("modelcheck exit = %d, want %d", code, ExitOK)
	}
	stderr.Reset()
	if code := RunCLI([]string{"modelcheck", "-maxStates", "3", "-config", unboundedCfg, unbounded}, nil, &stderr); code != ExitSemanticFailure {
		t.Fatalf("modelcheck maxStates exit = %d, want %d; stderr=%s", code, ExitSemanticFailure, stderr.String())
	}
	if !strings.Contains(stderr.String(), "state limit 3 reached") {
		t.Fatalf("modelcheck maxStates stderr = %q, want state limit message", stderr.String())
	}
	stderr.Reset()
	if code := RunCLI([]string{"modelcheck", "-config", traceFailCfg, traceFail}, nil, &stderr); code != ExitSemanticFailure {
		t.Fatalf("modelcheck trace failure exit = %d, want %d; stderr=%s", code, ExitSemanticFailure, stderr.String())
	}
	traceText := stderr.String()
	if !strings.Contains(traceText, "1: x = 1, y = 2") || strings.Contains(traceText, "map[") {
		t.Fatalf("modelcheck trace stderr = %q, want sorted assignment trace", traceText)
	}
}

func requireNoErrors(t *testing.T, diags Diagnostics) {
	t.Helper()
	if diags.HasErrors() {
		t.Fatalf("unexpected diagnostics:\n%s", diags.Error())
	}
}

func requireHasErrorContaining(t *testing.T, diags Diagnostics, text string) {
	t.Helper()
	for _, d := range diags {
		if d.Severity == SeverityError && strings.Contains(d.Message, text) {
			return
		}
	}
	t.Fatalf("expected error containing %q; diagnostics:\n%s", text, diags.Error())
}

func requireHasWarningContaining(t *testing.T, diags Diagnostics, text string) {
	t.Helper()
	for _, d := range diags {
		if d.Severity == SeverityWarning && strings.Contains(d.Message, text) {
			return
		}
	}
	t.Fatalf("expected warning containing %q; diagnostics:\n%s", text, diags.Error())
}

func asBinary(t *testing.T, expr Expr, op string) *BinaryExpr {
	t.Helper()
	bin, ok := expr.(*BinaryExpr)
	if !ok {
		t.Fatalf("expr %#v is %T, want *BinaryExpr", expr, expr)
	}
	if bin.Op != op {
		t.Fatalf("binary op = %q, want %q", bin.Op, op)
	}
	return bin
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func moduleNames(mods map[string]*Module) []string {
	names := make([]string, 0, len(mods))
	for name := range mods {
		names = append(names, name)
	}
	return names
}
