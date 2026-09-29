package tlc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestModelConfigDefaultsAndOrderedSections(t *testing.T) {
	ModelValueInit()
	cfg, err := ParseModelConfigSource("Spec.cfg", `
\* Lists collect tokens until the next config keyword, preserving order.
INIT Init
NEXT Next
SPECIFICATION Spec
VIEW ViewOp
SYMMETRY SymmetrySet
ALIAS TraceAlias
INVARIANTS InvA InvB
PROPERTY PropA
PROPERTIES PropB PropC
CONSTRAINTS ConstraintA ConstraintB
ACTION_CONSTRAINTS ActionConstraint
POSTCONDITIONS PostA PostB
_PERIODIC PeriodicOp
_RL_REWARD RewardOp
_POSSIBLE PossibleA PossibleB
`)
	if err != nil {
		t.Fatalf("ParseModelConfigSource returned error: %v", err)
	}

	if cfg.GetInit() != "Init" || cfg.GetNext() != "Next" || cfg.GetSpec() != "Spec" {
		t.Fatalf("singleton config values = INIT %q NEXT %q SPEC %q", cfg.GetInit(), cfg.GetNext(), cfg.GetSpec())
	}
	if cfg.GetView() != "ViewOp" || cfg.GetSymmetry() != "SymmetrySet" || cfg.GetAlias() != "TraceAlias" {
		t.Fatalf("view/symmetry/alias = %q/%q/%q", cfg.GetView(), cfg.GetSymmetry(), cfg.GetAlias())
	}
	if cfg.GetPeriodic() != "PeriodicOp" || cfg.GetRLReward() != "RewardOp" {
		t.Fatalf("periodic/reward = %q/%q", cfg.GetPeriodic(), cfg.GetRLReward())
	}
	if !cfg.GetCheckDeadlock() {
		t.Fatalf("check deadlock default = false, want true")
	}
	requireStrings(t, cfg.GetInvariants(), []string{"InvA", "InvB"})
	requireStrings(t, cfg.GetProperties(), []string{"PropA", "PropB", "PropC"})
	requireStrings(t, cfg.GetConstraints(), []string{"ConstraintA", "ConstraintB"})
	requireStrings(t, cfg.GetActionConstraints(), []string{"ActionConstraint"})
	requireStrings(t, cfg.GetPostConditions(), []string{"PostA", "PostB"})
	requireStrings(t, cfg.GetPossible(), []string{"PossibleA", "PossibleB"})
}

func TestModelConfigDuplicateSingletonKeywordsMatchJavaErrors(t *testing.T) {
	for _, tc := range []struct {
		name    string
		source  string
		keyword string
	}{
		{name: "init", source: "INIT Init\nINIT InitAgain\n", keyword: "INIT"},
		{name: "next", source: "NEXT Next\nNEXT NextAgain\n", keyword: "NEXT"},
		{name: "spec", source: "SPECIFICATION Spec\nSPECIFICATION SpecAgain\n", keyword: "SPECIFICATION"},
		{name: "deadlock", source: "CHECK_DEADLOCK TRUE\nCHECK_DEADLOCK FALSE\n", keyword: "CHECK_DEADLOCK"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseModelConfigSource("Spec.cfg", tc.source)
			if err == nil {
				t.Fatalf("ParseModelConfigSource succeeded, want duplicate keyword error")
			}
			cfgErr, ok := err.(*ConfigFileError)
			if !ok {
				t.Fatalf("error type = %T, want *ConfigFileError", err)
			}
			if cfgErr.Code != ECCFGTwiceKeyword {
				t.Fatalf("error code = %d, want %d", cfgErr.Code, ECCFGTwiceKeyword)
			}
			if cfgErr.Params[1] != tc.keyword {
				t.Fatalf("error keyword = %q, want %q", cfgErr.Params[1], tc.keyword)
			}
		})
	}
}

func TestModelConfigParsesConstantsOverridesAndModuleScopes(t *testing.T) {
	ModelValueInit()
	cfg, err := ParseModelConfigSource("Spec.cfg", `
CONSTANTS
  N = 3
  Color = Red
  Mod!K = {"blue", TRUE, 5}
  Abs <- Concrete
  Local <- [OtherMod] LocalImpl
  ModuleConst = [OtherMod] {One, Two}
`)
	if err != nil {
		t.Fatalf("ParseModelConfigSource returned error: %v", err)
	}

	if got := cfg.GetConstants().Len(); got != 3 {
		t.Fatalf("constants len = %d, want 3", got)
	}
	requireConfigConstant(t, cfg.GetConstants().At(0), "N", nil, "3")
	requireConfigConstant(t, cfg.GetConstants().At(1), "Color", nil, "Red")
	requireConfigConstant(t, cfg.GetConstants().At(2), "Mod!K", nil, `{"blue", TRUE, 5}`)

	if got := cfg.GetOverrides().Get("Abs"); got != "Concrete" {
		t.Fatalf("override Abs = %q, want Concrete", got)
	}
	if got := cfg.GetOverridenSpecNameForConfigName("Concrete"); got != "Abs" {
		t.Fatalf("reverse override Concrete = %q, want Abs", got)
	}
	modOverride, ok := cfg.GetModOverrides().Get2("OtherMod")
	if !ok {
		t.Fatalf("OtherMod overrides not found")
	}
	if got := modOverride.Get("Local"); got != "LocalImpl" {
		t.Fatalf("OtherMod Local override = %q, want LocalImpl", got)
	}
	modConstants, ok := cfg.GetModConstants().Get2("OtherMod")
	if !ok {
		t.Fatalf("OtherMod constants not found")
	}
	requireConfigConstant(t, modConstants.At(0), "ModuleConst", nil, "{One, Two}")
}

func TestModelConfigParsesParameterizedConstantsAndRawConstants(t *testing.T) {
	ModelValueInit()
	cfg, err := ParseModelConfigSource("Spec.cfg", `
CONSTANT
  Op(1, "x", {A, B}) = Result
CONSTANTS
  Other = TRUE
`)
	if err != nil {
		t.Fatalf("ParseModelConfigSource returned error: %v", err)
	}

	requireConfigConstant(t, cfg.GetConstants().At(0), "Op", []string{`1`, `"x"`, `{A, B}`}, "Result")
	requireConfigConstant(t, cfg.GetConstants().At(1), "Other", nil, "TRUE")

	raw := cfg.GetRawConstants()
	if len(raw) != 2 {
		t.Fatalf("raw constants len = %d, want 2", len(raw))
	}
	if raw[0] == "" || raw[1] == "" {
		t.Fatalf("raw constants must preserve the two CONSTANT sections: %#v", raw)
	}
	asList := cfg.GetConstantsAsList()
	requireNestedStrings(t, asList, [][]string{{"Op(1,\"x\",{A,B})", "Result"}, {"Other", "TRUE"}})
}

func TestModelConfigCheckDeadlockAndMalformedValues(t *testing.T) {
	cfg, err := ParseModelConfigSource("Spec.cfg", "CHECK_DEADLOCK FALSE\n")
	if err != nil {
		t.Fatalf("ParseModelConfigSource returned error: %v", err)
	}
	if cfg.GetCheckDeadlock() {
		t.Fatalf("check deadlock = true, want false")
	}

	_, err = ParseModelConfigSource("Spec.cfg", "CHECK_DEADLOCK MAYBE\n")
	if err == nil {
		t.Fatalf("ParseModelConfigSource succeeded, want malformed CHECK_DEADLOCK error")
	}
	if cfgErr := err.(*ConfigFileError); cfgErr.Code != ECCFGExpectedSymbol {
		t.Fatalf("error code = %d, want %d", cfgErr.Code, ECCFGExpectedSymbol)
	}
}

func TestModelConfigPathMirrorsJavaMonolithConfigPath(t *testing.T) {
	if got := ModelConfigPath("Scratch.tla"); got != "Scratch.tla" {
		t.Fatalf("ModelConfigPath Scratch.tla = %q, want Scratch.tla", got)
	}
	if got := ModelConfigPath("Scratch"); got != "Scratch.cfg" {
		t.Fatalf("ModelConfigPath Scratch = %q, want Scratch.cfg", got)
	}
	if got := ModelConfigPath("Scratch.cfg"); got != "Scratch.cfg" {
		t.Fatalf("ModelConfigPath Scratch.cfg = %q, want Scratch.cfg", got)
	}
}

func TestModelConfigFileParsingReadsCFG(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Spec.cfg")
	if err := os.WriteFile(path, []byte("SPECIFICATION Spec\nCHECK_DEADLOCK FALSE\n"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg, err := ParseModelConfigFile(path)
	if err != nil {
		t.Fatalf("ParseModelConfigFile returned error: %v", err)
	}
	if cfg.GetSpec() != "Spec" {
		t.Fatalf("SPECIFICATION = %q, want Spec", cfg.GetSpec())
	}
	if cfg.GetCheckDeadlock() {
		t.Fatalf("check deadlock = true, want false")
	}
}

func TestModelConfigExtractsMonolithConfig(t *testing.T) {
	source := "D:\\software\\TLA+\\Specs\\Scratch\\Scratch.tla\n\n\n" +
		"---- MODULE Scratch ----\n" +
		"EXTENDS TLC\n" +
		"Spec == TRUE /\\ [][TRUE]_TRUE\n" +
		"======\n\n" +
		"---- CONFIG Scratch ----\n" +
		"SPECIFICATION Spec\n" +
		"=====\n"

	config := ExtractMonolithConfigSource(source, "Scratch")
	if config != "SPECIFICATION Spec" {
		t.Fatalf("extracted config = %q, want SPECIFICATION Spec", config)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "Scratch.tla")
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	cfg, err := ParseModelConfigFile(path)
	if err != nil {
		t.Fatalf("ParseModelConfigFile returned error: %v", err)
	}
	if cfg.GetSpec() != "Spec" {
		t.Fatalf("full-path monolith config SPECIFICATION = %q, want Spec", cfg.GetSpec())
	}
	cfg, err = ParseModelConfigSource("Scratch.tla", ExtractMonolithConfigSource(source, "Scratch"))
	if err != nil {
		t.Fatalf("ParseModelConfigSource returned error: %v", err)
	}
	if cfg.GetSpec() != "Spec" {
		t.Fatalf("extracted monolith SPECIFICATION = %q, want Spec", cfg.GetSpec())
	}
}

func TestModelConfigMonolithWindowsPathNameDoesNotActLikeRegex(t *testing.T) {
	source := "D:\\software\\TLA+\\Specs\\Scratch\\Scratch.tla\n\n\n" +
		"---- MODULE Scratch ----\n" +
		"EXTENDS TLC\n" +
		"Spec == TRUE /\\ [][TRUE]_TRUE\n" +
		"======\n\n" +
		"---- CONFIG Scratch ----\n" +
		"SPECIFICATION Spec\n" +
		"=====\n"
	windowsPath := "d:\\software\\TLA+\\Specs\\Scratch\\Scratch"
	if got := ExtractMonolithConfigSource(source, windowsPath); got != "" {
		t.Fatalf("extracted config for windows path = %q, want empty", got)
	}
}

func requireStrings(t *testing.T, got []string, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("strings len = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("strings[%d] = %q, want %q in %#v", i, got[i], want[i], got)
		}
	}
}

func requireNestedStrings(t *testing.T, got [][]string, want [][]string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("nested strings len = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		requireStrings(t, got[i], want[i])
	}
}

func requireConfigConstant(t *testing.T, got ConfigConstant, name string, args []string, value string) {
	t.Helper()
	if got.Name != name {
		t.Fatalf("constant name = %q, want %q", got.Name, name)
	}
	if len(got.Args) != len(args) {
		t.Fatalf("constant %s arg len = %d, want %d", name, len(got.Args), len(args))
	}
	for i := range args {
		if got.Args[i].String() != args[i] {
			t.Fatalf("constant %s arg %d = %q, want %q", name, i, got.Args[i].String(), args[i])
		}
	}
	if got.Value == nil {
		t.Fatalf("constant %s value is nil", name)
	}
	if got.Value.String() != value {
		t.Fatalf("constant %s value = %q, want %q", name, got.Value.String(), value)
	}
}
