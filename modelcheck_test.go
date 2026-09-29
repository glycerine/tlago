package tlago

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigAndModelCheckingBehaviors(t *testing.T) {
	t.Run("parses core TLC cfg entries", func(t *testing.T) {
		cfg, diags := ParseConfigSource("MC.cfg", `INIT Init
NEXT Next
CONSTANT N = 3
CONSTANT Red = Red
INVARIANT TypeOK
INVARIANTS Safe LiveEnough
VIEW ViewOp
SYMMETRY SymmetrySet
ALIAS AliasOp
POSTCONDITION Done
`)
		requireNoErrors(t, diags)
		if cfg.Init != "Init" || cfg.Next != "Next" {
			t.Fatalf("cfg init/next = %#v", cfg)
		}
		if strings.Join(cfg.Invariants, ",") != "TypeOK,Safe,LiveEnough" {
			t.Fatalf("invariants = %v", cfg.Invariants)
		}
		if cfg.Constants["N"] != 3 {
			t.Fatalf("constant N = %d, want 3", cfg.Constants["N"])
		}
		if cfg.ModelValues["Red"] != "Red" {
			t.Fatalf("model value Red = %q, want Red", cfg.ModelValues["Red"])
		}
		if strings.Join(cfg.Views, ",") != "ViewOp" || strings.Join(cfg.Symmetry, ",") != "SymmetrySet" || strings.Join(cfg.Aliases, ",") != "AliasOp" || strings.Join(cfg.Postconditions, ",") != "Done" {
			t.Fatalf("extra cfg entries not captured: %#v", cfg)
		}
	})

	t.Run("explores a finite counter and accepts a true invariant", func(t *testing.T) {
		result, diags := ModelCheckSanySource("Counter.tla", counterSpec("x <= 3"), counterCfg(), ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK {
			t.Fatalf("model check failed: %s trace=%v", result.Error, result.Trace)
		}
		if result.StatesExplored != 4 {
			t.Fatalf("states explored = %d, want 4", result.StatesExplored)
		}
	})

	t.Run("reports an invariant violation with a trace", func(t *testing.T) {
		result, diags := ModelCheckSanySource("Counter.tla", counterSpec("x < 3"), counterCfg(), ModelCheckOptions{})
		requireNoErrors(t, diags)
		if result.OK {
			t.Fatalf("model check unexpectedly succeeded")
		}
		if !strings.Contains(result.Error, "Inv") {
			t.Fatalf("error = %q, want invariant name", result.Error)
		}
		if len(result.Trace) == 0 || result.Trace[len(result.Trace)-1]["x"] != 3 {
			t.Fatalf("trace = %v, want final x=3", result.Trace)
		}
	})

	t.Run("uses integer constants from cfg", func(t *testing.T) {
		spec := `---- MODULE BoundedCounter ----
CONSTANT N
VARIABLE x
Init == x = 0
Next == (x < N /\ x' = x + 1) \/ (x = N /\ x' = x)
Inv == x <= N
====`
		cfg := `CONSTANT N = 3
INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("BoundedCounter.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 4 {
			t.Fatalf("result = %#v, want success with 4 states", result)
		}
	})

	t.Run("uses cfg model values in invariants", func(t *testing.T) {
		spec := `---- MODULE ModelValueInvariant ----
CONSTANT Red, Blue
VARIABLE x
ASSUME Red /= Blue
Init == x = 0
Next == x' = x
Inv == Red /= Blue /\ Red \in {Red, Blue} /\ Blue \notin {Red} /\ {Red} \subset {Red, Blue} /\ {Red, Blue} = {Blue, Red}
====`
		cfg := `CONSTANTS Red = Red Blue = Blue
INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("ModelValueInvariant.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 1 {
			t.Fatalf("result = %#v, want model-value invariant success with 1 state", result)
		}
	})

	t.Run("evaluates Java syntax corpus bitfield number formats", func(t *testing.T) {
		spec := `---- MODULE BitfieldNumbers ----
VARIABLE x
Init == x = \b1010
Next == x' = x
Inv == x = 10 /\ \B1010 = 10 /\ \o12 = 10 /\ \O12 = 10 /\ \h0a = 10 /\ \H0A = 10
====`
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("BitfieldNumbers.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 1 {
			t.Fatalf("result = %#v, want bitfield number invariant success with 1 state", result)
		}
	})

	t.Run("checks assumptions before initial states", func(t *testing.T) {
		spec := `---- MODULE AssumedCounter ----
CONSTANT N
ASSUME N > 0
VARIABLE x
Init == x = 0
Next == x' = x
Inv == TRUE
====`
		cfg := `CONSTANT N = 0
INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("AssumedCounter.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if result.OK || result.StatesExplored != 0 || !strings.Contains(result.Error, "assumption") {
			t.Fatalf("result = %#v, want assumption failure before exploration", result)
		}
	})

	t.Run("explores finite nondeterminism from set membership", func(t *testing.T) {
		spec := `---- MODULE Nondet ----
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

	t.Run("applies cfg state constraints", func(t *testing.T) {
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
CONSTRAINT InScope
CHECK_DEADLOCK FALSE
`
		spec := `---- MODULE Constrained ----
VARIABLE x
Init == x = 0
Next == x' = x + 1
Inv == x <= 3
InScope == x < 2
====`
		result, diags := ModelCheckSanySource("Constrained.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 2 {
			t.Fatalf("result = %#v, want constrained success with 2 states", result)
		}
	})

	t.Run("supports unchanged variables in actions", func(t *testing.T) {
		spec := `---- MODULE TwoVars ----
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

	t.Run("compares string literals in invariants", func(t *testing.T) {
		spec := `---- MODULE StringInvariant ----
VARIABLE x
Init == x = 0
Next == x' = x
Inv == "ready" = "ready" /\ "ready" /= "done"
====`
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("StringInvariant.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 1 {
			t.Fatalf("result = %#v, want string invariant success with 1 state", result)
		}
	})

	t.Run("checks finite string set predicates", func(t *testing.T) {
		spec := `---- MODULE StringSetInvariant ----
VARIABLE x
Init == x = 0
Next == x' = x
Inv == "ready" \in {"ready", "done"} /\ "other" \notin {"ready", "done"} /\ {"ready", "done"} = {"done", "ready"} /\ {"ready"} \subseteq {"ready", "done"}
====`
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("StringSetInvariant.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 1 {
			t.Fatalf("result = %#v, want string set invariant success with 1 state", result)
		}
	})

	t.Run("checks strict finite set subset predicates", func(t *testing.T) {
		spec := `---- MODULE StrictSubsetInvariant ----
VARIABLE x
Init == x = 0
Next == x' = x
Inv == {0} \subset {0, 1} /\ ~({0, 1} \subset {0, 1}) /\ {"ready"} \subset {"ready", "done"}
====`
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("StrictSubsetInvariant.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 1 {
			t.Fatalf("result = %#v, want strict subset invariant success with 1 state", result)
		}
	})

	t.Run("enumerates finite powersets with SUBSET", func(t *testing.T) {
		spec := `---- MODULE PowersetInvariant ----
VARIABLE x
Init == x = 0
Next == x' = x
Inv == {0} \in SUBSET {0, 1} /\ {"ready"} \in SUBSET {"ready", "done"}
====`
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("PowersetInvariant.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 1 {
			t.Fatalf("result = %#v, want powerset invariant success with 1 state", result)
		}
	})

	t.Run("compares finite set-of-set literals", func(t *testing.T) {
		spec := `---- MODULE SetOfSetsInvariant ----
VARIABLE x
Init == x = 0
Next == x' = x
Inv == SUBSET {0, 1} = {{}, {0}, {1}, {0, 1}} /\ {"ready"} \in {{"ready"}, {"done"}} /\ SUBSET {"ready"} = {{}, {"ready"}}
====`
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("SetOfSetsInvariant.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 1 {
			t.Fatalf("result = %#v, want set-of-set invariant success with 1 state", result)
		}
	})

	t.Run("applies cfg action constraints to primed successor values", func(t *testing.T) {
		spec := `---- MODULE ActionConstrained ----
VARIABLE x
Init == x = 0
Next == x' \in {x, x + 1}
Inv == x <= 1
StepOK == x' <= 1
====`
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
ACTION_CONSTRAINT StepOK
CHECK_DEADLOCK FALSE
`
		result, diags := ModelCheckSanySource("ActionConstrained.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 2 {
			t.Fatalf("result = %#v, want constrained success with 2 states", result)
		}
	})

	t.Run("evaluates ENABLED over simple actions", func(t *testing.T) {
		spec := `---- MODULE EnabledAction ----
VARIABLE x
Init == x = 0
Next == x < 1 /\ x' = x + 1
Inv == IF x = 0 THEN ENABLED Next ELSE ~ ENABLED Next
====`
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
CHECK_DEADLOCK FALSE
`
		result, diags := ModelCheckSanySource("EnabledAction.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 2 {
			t.Fatalf("result = %#v, want ENABLED invariant success with 2 states", result)
		}
	})

	t.Run("evaluates LET IN expressions in actions and invariants", func(t *testing.T) {
		spec := `---- MODULE LetCounter ----
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

	t.Run("inlines parameterized operator calls in model checking", func(t *testing.T) {
		spec := `---- MODULE OperatorCalls ----
VARIABLE x
Inc(n) == n + 1
Init == x = 0
Next == (x < 3 /\ x' = Inc(x)) \/ (x = 3 /\ x' = x)
Inv == Inc(x) <= 4
====`
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("OperatorCalls.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 4 {
			t.Fatalf("result = %#v, want success with 4 states", result)
		}
	})

	t.Run("inlines parameterless operator references in model checking", func(t *testing.T) {
		spec := `---- MODULE OperatorRefs ----
VARIABLE x
Start == x = 0
Step == (x < 2 /\ x' = x + 1) \/ (x = 2 /\ x' = x)
Init == Start
Next == Step
Inv == x <= 2
====`
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
`
		result, diags := ModelCheckSanySource("OperatorRefs.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 3 {
			t.Fatalf("result = %#v, want success with 3 states", result)
		}
	})

	t.Run("evaluates finite integer quantifiers", func(t *testing.T) {
		spec := `---- MODULE Quantified ----
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

	t.Run("checks cfg state properties", func(t *testing.T) {
		spec := `---- MODULE PropertyCounter ----
VARIABLE x
Init == x = 0
Next == (x < 2 /\ x' = x + 1) \/ (x = 2 /\ x' = x)
Safe == x <= 2
TooSmall == x < 2
====`
		okCfg := `INIT Init
NEXT Next
PROPERTY Safe
`
		result, diags := ModelCheckSanySource("PropertyCounter.tla", spec, okCfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 3 {
			t.Fatalf("result = %#v, want property success with 3 states", result)
		}

		badCfg := `INIT Init
NEXT Next
PROPERTY TooSmall
`
		result, diags = ModelCheckSanySource("PropertyCounter.tla", spec, badCfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if result.OK || !strings.Contains(result.Error, "property TooSmall") {
			t.Fatalf("result = %#v, want property violation", result)
		}
		if len(result.Trace) == 0 || result.Trace[len(result.Trace)-1]["x"] != 2 {
			t.Fatalf("trace = %v, want final x=2", result.Trace)
		}
	})

	t.Run("evaluates cfg postconditions after successful exploration", func(t *testing.T) {
		spec := `---- MODULE PostconditionCounter ----
CONSTANT N
VARIABLE x
Init == x = 0
Next == (x < N /\ x' = x + 1) \/ (x = N /\ x' = x)
Inv == x <= N
PostOK == N = 2
PostFail == N = 3
====`
		okCfg := `CONSTANT N = 2
INIT Init
NEXT Next
INVARIANT Inv
POSTCONDITION PostOK
`
		result, diags := ModelCheckSanySource("PostconditionCounter.tla", spec, okCfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 3 {
			t.Fatalf("result = %#v, want postcondition success after 3 states", result)
		}

		badCfg := `CONSTANT N = 2
INIT Init
NEXT Next
INVARIANT Inv
POSTCONDITION PostFail
`
		result, diags = ModelCheckSanySource("PostconditionCounter.tla", spec, badCfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if result.OK || result.StatesExplored != 3 || !strings.Contains(result.Error, "postcondition PostFail") {
			t.Fatalf("result = %#v, want postcondition violation after exploration", result)
		}
	})

	t.Run("uses cfg view expressions for state identity", func(t *testing.T) {
		spec := `---- MODULE ViewModel ----
VARIABLE x, y
Init == x = 0 /\ y \in {0, 1}
Next == UNCHANGED x /\ y' \in {0, 1}
Inv == x = 0
View == x
====`
		cfg := `INIT Init
NEXT Next
INVARIANT Inv
VIEW View
CHECK_DEADLOCK FALSE
`
		result, diags := ModelCheckSanySource("ViewModel.tla", spec, cfg, ModelCheckOptions{})
		requireNoErrors(t, diags)
		if !result.OK || result.StatesExplored != 1 {
			t.Fatalf("result = %#v, want view-collapsed success with 1 state", result)
		}
	})

	t.Run("evaluates integer intervals as finite sets", func(t *testing.T) {
		spec := `---- MODULE Intervals ----
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

	t.Run("uses module-qualified integer definitions from dependencies", func(t *testing.T) {
		dir := t.TempDir()
		root := filepath.Join(dir, "Root.tla")
		writeFile(t, filepath.Join(dir, "Helper.tla"), `---- MODULE Helper ----
Zero == 0
One == Zero + 1
====`)
		writeFile(t, root, `---- MODULE Root ----
EXTENDS Helper
VARIABLE x
Init == x = Helper!Zero
Next == (x < Helper!One /\ x' = x + Helper!One) \/ (x = Helper!One /\ x' = x)
Inv == x <= Helper!One
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
			t.Fatalf("result = %#v, want success with 2 states", result)
		}
	})

	t.Run("parses prefix conjunction and disjunction lists", func(t *testing.T) {
		spec := `---- MODULE JunctionLists ----
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
}

func counterSpec(invariant string) string {
	return `---- MODULE Counter ----
VARIABLE x
Init == x = 0
Next == (x < 3 /\ x' = x + 1) \/ (x = 3 /\ x' = x)
Inv == ` + invariant + `
====`
}

func counterCfg() string {
	return `INIT Init
NEXT Next
INVARIANT Inv
`
}
