package tlago

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestApalacheIRConformanceAgainstApalache(t *testing.T) {
	apalache := apalacheCommandForTest(t)
	if apalache == "" {
		t.Skip("set APALACHE_MC or put apalache-mc on PATH to run ApalacheIR oracle conformance")
	}

	runApalacheIRConformanceCase(t, apalache, "OracleSmoke", `---- MODULE OracleSmoke ----
EXTENDS Naturals
CONSTANT C
VARIABLE x
A == 1 + 2
B(y) == IF y \in {1, 2} THEN TRUE ELSE FALSE
ASSUME C = 3
====`)

	runApalacheIRConformanceCase(t, apalache, "OracleStructured", `---- MODULE OracleStructured ----
EXTENDS Naturals
Set == {1, 2}
Tuple == <<1, 2>>
Record == [a |-> 1, b |-> 2]
RecordSet == [a: {1}, b: {2}]
Fcn == [i \in {1, 2} |-> i + 1]
App == Fcn[1]
Filtered == {i \in {1, 2}: i > 1}
Mapped == {i + 1 : i \in {1, 2}}
Local == LET Inc(i) == i + 1 IN Inc(1)
All == \A i \in {1, 2}: i \in {1, 2}
Choice == CHOOSE i \in {1, 2}: i > 1
====`)

	runApalacheIRConformanceCase(t, apalache, "OracleNestedJunctionImplication", `---- MODULE OracleNestedJunctionImplication ----
Nested == /\ TRUE
          /\ /\ TRUE
             /\ \/ TRUE /\ FALSE
                \/ TRUE
             => FALSE
          /\ TRUE
====`)

	runApalacheIRConformanceCase(t, apalache, "OracleFunctionDefinitions", `---- MODULE OracleFunctionDefinitions ----
EXTENDS Naturals
S == {1, 2}
F[i \in S] == i + 1
G[i \in S, j \in S] == i + j
App == F[1] + G[1, 2]
====`)

	runApalacheIRConformanceFiles(t, apalache, "OracleSequences", map[string]string{
		"__apalache_folds": mustReadFile(t, filepath.Join("test_vectors", "tlaplus-standard-modules", "__apalache_folds.tla")),
		"Sequences":        mustReadFile(t, filepath.Join("test_vectors", "tlaplus-standard-modules", "Sequences.tla")),
		"OracleSequences": `---- MODULE OracleSequences ----
EXTENDS Sequences
SeqOps(s) == <<Len(s), Head(s), Tail(s), Append(s, 1), SubSeq(s, 1, 1), s \o s>>
Use == SeqOps(<<1, 2>>)
====`,
	})

	runApalacheIRConformanceFiles(t, apalache, "OracleFiniteSets", map[string]string{
		"OracleFiniteSets": `---- MODULE OracleFiniteSets ----
EXTENDS FiniteSets
Use == IsFiniteSet({1}) /\ Cardinality({1}) = 1
====`,
	})

	runApalacheIRConformanceCase(t, apalache, "OracleRecursiveFunction", `---- MODULE OracleRecursiveFunction ----
EXTENDS Naturals
RECURSIVE Fact
Fact[n \in 0..2] == IF n = 0 THEN 1 ELSE n * Fact[n - 1]
Use == Fact[2]
====`)

	runApalacheIRConformanceCase(t, apalache, "OracleRecursive", `---- MODULE OracleRecursive ----
EXTENDS Naturals
RECURSIVE Fact(_)
Fact(n) == IF n = 0 THEN 1 ELSE n * Fact(n - 1)
====`)

	runApalacheIRConformanceCase(t, apalache, "OracleRecursiveLet", `---- MODULE OracleRecursiveLet ----
EXTENDS Naturals
RECURSIVE Sum(_)
Sum(S) == IF S = {} THEN 0 ELSE LET x == CHOOSE y \in S : TRUE IN Sum(S \ {x})
====`)

	runApalacheIRConformanceCase(t, apalache, "OracleRecursiveLetChooseShadow", `---- MODULE OracleRecursiveLetChooseShadow ----
EXTENDS Naturals
RECURSIVE Sum(_)
Sum(S) == IF S = {} THEN 0 ELSE LET x == CHOOSE x \in S : TRUE IN Sum(S \ {x})
====`)

	runApalacheIRConformanceCase(t, apalache, "OracleNamedFacts", `---- MODULE OracleNamedFacts ----
CONSTANT C
ASSUME NamedAssume == C = 1
THEOREM NamedTheorem == TRUE
====`)

	runApalacheIRConformanceCase(t, apalache, "OracleCaseExcept", `---- MODULE OracleCaseExcept ----
VARIABLE f
CaseNoOther == CASE TRUE -> 1 [] FALSE -> 2
CaseOther == CASE FALSE -> 1 [] OTHER -> 2
ExceptOne == [f EXCEPT ![1] = 2]
ExceptNested == [f EXCEPT ![1][2].a = 3]
ExceptTupleIndex == [f EXCEPT ![1, 2] = 3]
ExceptAt == [f EXCEPT ![1] = @]
ExceptAtIndex == [f EXCEPT ![1] = @[2]]
====`)

	runApalacheIRConformanceCase(t, apalache, "OracleTemporalAction", `---- MODULE OracleTemporalAction ----
VARIABLE x
True == TRUE
Prime == x'
Enabled == ENABLED x
Unchanged == UNCHANGED x
SquareAct == [True]_x
AngleAct == <<True>>_x
Temporal == []TRUE /\ <>TRUE /\ (TRUE ~> TRUE) /\ (True -+-> True)
Fairness == WF_x(True) /\ SF_x(True)
====`)

	runApalacheIRConformanceCase(t, apalache, "OracleUnboundedQuantifiers", `---- MODULE OracleUnboundedQuantifiers ----
ForAll == \A x : TRUE
Exists == \E x : TRUE
Choose == CHOOSE x : TRUE
TemporalExists == \EE x : TRUE
TemporalForall == \AA x : TRUE
====`)

	runApalacheIRConformanceCase(t, apalache, "OracleCoreBuiltins", `---- MODULE OracleCoreBuiltins ----
VARIABLE x
Eq == FALSE = TRUE
Ne == FALSE /= TRUE
Not == ~FALSE
Or == FALSE \/ TRUE
And == FALSE /\ TRUE
Equiv == FALSE <=> TRUE
Implies == FALSE => TRUE
Subset == SUBSET x
Union == UNION x
Domain == DOMAIN x
Subseteq == x \subseteq x
In == x \in x
Notin == x \notin x
Setminus == x \ x
Cap == x \cap x
Cup == x \cup x
Times == x \X x
CartesianProd == x \X x \X x
Cdot == (TRUE \cdot TRUE)
SetOfFcns == [x -> x]
Boolean == BOOLEAN
String == STRING
====`)

	runApalacheIRConformanceCase(t, apalache, "OracleUserInfix", `---- MODULE OracleUserInfix ----
EXTENDS Naturals
a \prec b == a[1] < b[1] \/ (a[1] = b[1] /\ a[2] < b[2])
Use == <<1, 1>> \prec <<2, 2>>
====`)

	runApalacheIRConformanceFiles(t, apalache, "OracleExtendsLocal", map[string]string{
		"Helper": `---- MODULE Helper ----
CONSTANT C
Shared == C
====`,
		"OracleExtendsLocal": `---- MODULE OracleExtendsLocal ----
EXTENDS Helper
Use == Shared
====`,
	})

	runApalacheIRConformanceFiles(t, apalache, "OraclePlainInstance", map[string]string{
		"Inner": `---- MODULE Inner ----
CONSTANT C
Op == C
====`,
		"OraclePlainInstance": `---- MODULE OraclePlainInstance ----
CONSTANT X
INSTANCE Inner WITH C <- X
Use == Op
====`,
	})

	runApalacheIRConformanceFiles(t, apalache, "OraclePlainInstanceDuplicateDefinition", map[string]string{
		"Inner": `---- MODULE Inner ----
Op == 1
Other == Op
====`,
		"OraclePlainInstanceDuplicateDefinition": `---- MODULE OraclePlainInstanceDuplicateDefinition ----
Op == 1
INSTANCE Inner
Use == Other
====`,
	})

	runApalacheIRConformanceFiles(t, apalache, "OracleNamedInstance", map[string]string{
		"Inner": `---- MODULE Inner ----
CONSTANT C
Op == C
====`,
		"OracleNamedInstance": `---- MODULE OracleNamedInstance ----
CONSTANT X
Inst == INSTANCE Inner WITH C <- X
Use == Inst!Op
====`,
	})

	runApalacheIRConformanceFiles(t, apalache, "OracleLetInstance", map[string]string{
		"Inner": `---- MODULE Inner ----
A == 1
B == A
====`,
		"OracleLetInstance": `---- MODULE OracleLetInstance ----
Use == LET I == INSTANCE Inner IN I!B
====`,
	})
}

func TestApalacheIRSourceConformanceAgainstApalache(t *testing.T) {
	apalache := apalacheCommandForTest(t)
	if apalache == "" {
		t.Skip("set APALACHE_MC or put apalache-mc on PATH to run ApalacheIR oracle source conformance")
	}

	runApalacheIRSourceConformanceCase(t, apalache, "OracleSourceSmoke", `---- MODULE OracleSourceSmoke ----
EXTENDS Naturals
CONSTANT C
VARIABLE x
A == 1 + 2
B(y) == IF y = 1 THEN x ELSE C
ASSUME C = 3
====`)

	runApalacheIRSourceConformanceCase(t, apalache, "OracleSourceStructured", `---- MODULE OracleSourceStructured ----
EXTENDS Naturals
VARIABLE f
Set == {1, 2}
Tuple == <<1, 2>>
Record == [a |-> 1]
Fcn == [i \in {1, 2} |-> i + 1]
App == Fcn[1]
Local == LET Inc(i) == i + 1 IN Inc(1)
CaseOther == CASE FALSE -> 1 [] OTHER -> 2
ExceptOne == [f EXCEPT ![1] = 2]
ExceptTupleIndex == [f EXCEPT ![1, 2] = 3]
ExceptAt == [f EXCEPT ![1] = @]
ExceptAtIndex == [f EXCEPT ![1] = @[2]]
====`)

	runApalacheIRSourceConformanceFiles(t, apalache, "OracleSourceInstance", map[string]string{
		"Inner": `---- MODULE Inner ----
CONSTANT C
Op == C
====`,
		"OracleSourceInstance": `---- MODULE OracleSourceInstance ----
CONSTANT X
INSTANCE Inner WITH C <- X
Use == Op
====`,
	})

	runApalacheIRSourceConformanceFiles(t, apalache, "OracleSourceRepeatedSubstitution", map[string]string{
		"Inner": `---- MODULE Inner ----
CONSTANT C
Op == C = 0 /\ C = 1
====`,
		"OracleSourceRepeatedSubstitution": `---- MODULE OracleSourceRepeatedSubstitution ----
CONSTANT X
INSTANCE Inner WITH C <- X
Use == Op
====`,
	})

	runApalacheIRSourceConformanceFiles(t, apalache, "OracleSourceInstanceLet", map[string]string{
		"Inner": `---- MODULE Inner ----
CONSTANT C
Op == LET Local == C IN Local
====`,
		"OracleSourceInstanceLet": `---- MODULE OracleSourceInstanceLet ----
CONSTANT X
INSTANCE Inner WITH C <- X
Use == Op
====`,
	})

	runApalacheIRSourceConformanceFiles(t, apalache, "OracleSourceImplicitInstance", map[string]string{
		"Inner": `---- MODULE Inner ----
EXTENDS Naturals
VARIABLE x
Op == x = 0 /\ x' = x + 1
====`,
		"OracleSourceImplicitInstance": `---- MODULE OracleSourceImplicitInstance ----
VARIABLE x
INSTANCE Inner
Use == Op
====`,
	})
}

func runApalacheIRConformanceCase(t *testing.T, apalache, moduleName, source string) {
	t.Helper()
	runApalacheIRCase(t, apalache, moduleName, source, ApalacheIROptions{}, normalizedApalacheJSON)
}

func runApalacheIRConformanceFiles(t *testing.T, apalache, rootName string, sources map[string]string) {
	t.Helper()
	runApalacheIRFiles(t, apalache, rootName, sources, ApalacheIROptions{}, normalizedApalacheJSON)
}

func runApalacheIRSourceConformanceCase(t *testing.T, apalache, moduleName, source string) {
	t.Helper()
	runApalacheIRCase(t, apalache, moduleName, source, ApalacheIROptions{IncludeSource: true}, canonicalApalacheJSON)
}

func runApalacheIRSourceConformanceFiles(t *testing.T, apalache, rootName string, sources map[string]string) {
	t.Helper()
	runApalacheIRFiles(t, apalache, rootName, sources, ApalacheIROptions{IncludeSource: true}, canonicalApalacheJSON)
}

func runApalacheIRCase(t *testing.T, apalache, moduleName, source string, opts ApalacheIROptions, normalize func(*testing.T, []byte) []byte) {
	t.Helper()
	dir := t.TempDir()
	specPath := filepath.Join(dir, moduleName+".tla")
	writeFile(t, specPath, source)

	wantPath := filepath.Join(dir, moduleName+".json")
	outDir := filepath.Join(dir, "apalache-out")
	cmd := exec.Command(apalache, "--out-dir="+outDir, "parse", "--output="+wantPath, specPath)
	cmd.Env = apalacheEnvForTest()
	var combined bytes.Buffer
	cmd.Stdout = &combined
	cmd.Stderr = &combined
	if err := cmd.Run(); err != nil {
		t.Fatalf("%s: run Apalache oracle: %v\n%s", moduleName, err, combined.String())
	}
	wantBytes, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("%s: read Apalache oracle JSON: %v", moduleName, err)
	}

	gotBytes, diags := ApalacheIRJSONSource(specPath, mustReadFile(t, specPath), opts)
	requireNoErrors(t, diags)

	want := normalize(t, wantBytes)
	got := normalize(t, gotBytes)
	if !bytes.Equal(got, want) {
		t.Fatalf("%s: canonical ApalacheIR mismatch\n got:\n%s\nwant:\n%s", moduleName, got, want)
	}
}

func runApalacheIRFiles(t *testing.T, apalache, rootName string, sources map[string]string, opts ApalacheIROptions, normalize func(*testing.T, []byte) []byte) {
	t.Helper()
	dir := t.TempDir()
	for moduleName, source := range sources {
		writeFile(t, filepath.Join(dir, moduleName+".tla"), source)
	}
	specPath := filepath.Join(dir, rootName+".tla")

	wantPath := filepath.Join(dir, rootName+".json")
	outDir := filepath.Join(dir, "apalache-out")
	cmd := exec.Command(apalache, "--out-dir="+outDir, "parse", "--output="+wantPath, specPath)
	cmd.Env = apalacheEnvForTest()
	var combined bytes.Buffer
	cmd.Stdout = &combined
	cmd.Stderr = &combined
	if err := cmd.Run(); err != nil {
		t.Fatalf("%s: run Apalache oracle: %v\n%s", rootName, err, combined.String())
	}
	wantBytes, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("%s: read Apalache oracle JSON: %v", rootName, err)
	}

	spec, diags := LoadSanySpec(specPath, LoadOptions{LibraryPaths: []string{dir}, PreferLibraryModules: true})
	requireNoErrors(t, diags)
	semDiags := CheckSpec(spec)
	requireNoErrors(t, semDiags)
	gotBytes, irDiags := ApalacheIRJSON(spec, opts)
	requireNoErrors(t, irDiags)

	want := normalize(t, wantBytes)
	got := normalize(t, gotBytes)
	if !bytes.Equal(got, want) {
		t.Fatalf("%s: canonical ApalacheIR mismatch\n got:\n%s\nwant:\n%s", rootName, got, want)
	}
}

func apalacheCommandForTest(t *testing.T) string {
	t.Helper()
	if path := os.Getenv("APALACHE_MC"); path != "" {
		return path
	}
	path, err := exec.LookPath("apalache-mc")
	if err != nil {
		return ""
	}
	return path
}

func apalacheEnvForTest() []string {
	env := os.Environ()
	if os.Getenv("APALACHE_HOME") == "" {
		const localApalache = "/mnt/oldrog/home/jaten/go/src/github.com/apalache-mc/apalache"
		if st, err := os.Stat(localApalache); err == nil && st.IsDir() {
			env = append(env, "APALACHE_HOME="+localApalache)
		}
	}
	return env
}

func normalizedApalacheJSON(t *testing.T, data []byte) []byte {
	t.Helper()
	return canonicalJSONValueForTest(t, canonicalizeApalacheIRValue(dropApalacheUnstableFields(apalacheJSONValue(t, data))))
}

func canonicalApalacheJSON(t *testing.T, data []byte) []byte {
	t.Helper()
	out, err := canonicalApalacheIRJSON(data)
	if err != nil {
		t.Fatalf("canonicalize ApalacheIR JSON: %v\n%s", err, data)
	}
	return out
}

func apalacheJSONValue(t *testing.T, data []byte) any {
	t.Helper()
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatalf("unmarshal ApalacheIR JSON: %v\n%s", err, data)
	}
	return value
}

func dropApalacheUnstableFields(value any) any {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, child := range v {
			if key == "source" {
				continue
			}
			out[key] = dropApalacheUnstableFields(child)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, child := range v {
			out[i] = dropApalacheUnstableFields(child)
		}
		return out
	default:
		return value
	}
}

func canonicalJSONValueForTest(t *testing.T, value any) []byte {
	t.Helper()
	return canonicalJSONBytes(value)
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
