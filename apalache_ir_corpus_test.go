package tlago

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestApalacheIRTargetCorpusAgainstApalache(t *testing.T) {
	if os.Getenv("TLAGO_APALACHE_CORPUS") == "" {
		t.Skip("set TLAGO_APALACHE_CORPUS=1 to compare the target fixture trees against Apalache")
	}
	apalache := apalacheCommandForTest(t)

	targets := apalacheCorpusTargets(t)
	if len(targets) == 0 {
		t.Fatalf("no .tla corpus files found under target directories")
	}
	targets = apalacheCorpusWindow(targets)
	limit := envPositiveInt("TLAGO_APALACHE_CORPUS_LIMIT")
	if limit > 0 && limit < len(targets) {
		targets = targets[:limit]
	}
	maxFailures := envPositiveInt("TLAGO_APALACHE_CORPUS_MAX_FAILURES")
	if maxFailures == 0 {
		maxFailures = 1
	}

	var failures []string
	for i, path := range targets {
		result := compareApalacheIRCorpusFile(t, apalache, path)
		if result != "" {
			failures = append(failures, fmt.Sprintf("%d/%d %s: %s", i+1, len(targets), path, result))
			if len(failures) >= maxFailures {
				break
			}
		}
	}
	if len(failures) > 0 {
		t.Fatalf("ApalacheIR corpus mismatch:\n%s", strings.Join(failures, "\n\n"))
	}
}

func TestApalacheIRCorpusCanonicalJSONBehaviors(t *testing.T) {
	t.Run("ignores object key order and declaration order", func(t *testing.T) {
		left := []byte(`{"modules":[{"declarations":[{"name":"B","kind":"TlaOperDecl"},{"name":"A","kind":"TlaVarDecl","body":{"decls":[{"name":"LocalB","kind":"TlaOperDecl"},{"name":"LocalA","kind":"TlaOperDecl"}],"kind":"LetInEx"}}],"kind":"TlaModule","name":"M"}],"version":"1.0","name":"ApalacheIR"}`)
		right := []byte(`{"name":"ApalacheIR","version":"1.0","modules":[{"name":"M","kind":"TlaModule","declarations":[{"kind":"TlaVarDecl","name":"A","body":{"kind":"LetInEx","decls":[{"kind":"TlaOperDecl","name":"LocalA"},{"kind":"TlaOperDecl","name":"LocalB"}]}},{"kind":"TlaOperDecl","name":"B"}]}]}`)
		leftCanon, err := canonicalApalacheIRJSON(left)
		if err != nil {
			t.Fatalf("canonicalize left: %v", err)
		}
		rightCanon, err := canonicalApalacheIRJSON(right)
		if err != nil {
			t.Fatalf("canonicalize right: %v", err)
		}
		if !bytes.Equal(leftCanon, rightCanon) {
			t.Fatalf("canonical JSON differs\nleft:  %s\nright: %s", leftCanon, rightCanon)
		}
	})

	t.Run("preserves expression argument order", func(t *testing.T) {
		left := []byte(`{"kind":"OperEx","oper":"MINUS","args":[{"kind":"NameEx","name":"x"},{"kind":"NameEx","name":"y"}]}`)
		right := []byte(`{"oper":"MINUS","kind":"OperEx","args":[{"name":"y","kind":"NameEx"},{"name":"x","kind":"NameEx"}]}`)
		leftCanon, err := canonicalApalacheIRJSON(left)
		if err != nil {
			t.Fatalf("canonicalize left: %v", err)
		}
		rightCanon, err := canonicalApalacheIRJSON(right)
		if err != nil {
			t.Fatalf("canonicalize right: %v", err)
		}
		if bytes.Equal(leftCanon, rightCanon) {
			t.Fatalf("canonical JSON unexpectedly ignored expression argument order: %s", leftCanon)
		}
	})

	t.Run("corpus canonicalization ignores Apalache output-file module names", func(t *testing.T) {
		left := []byte(`{"name":"ApalacheIR","version":"1.0","modules":[{"kind":"TlaModule","name":"Bakery","declarations":[]}]}`)
		right := []byte(`{"name":"ApalacheIR","version":"1.0","modules":[{"kind":"TlaModule","name":"out","declarations":[]}]}`)
		plainLeft, err := canonicalApalacheIRJSON(left)
		if err != nil {
			t.Fatalf("canonicalize plain left: %v", err)
		}
		plainRight, err := canonicalApalacheIRJSON(right)
		if err != nil {
			t.Fatalf("canonicalize plain right: %v", err)
		}
		if bytes.Equal(plainLeft, plainRight) {
			t.Fatalf("plain canonical JSON unexpectedly ignored module names: %s", plainLeft)
		}
		corpusLeft, err := canonicalApalacheIRCorpusJSON(left)
		if err != nil {
			t.Fatalf("canonicalize corpus left: %v", err)
		}
		corpusRight, err := canonicalApalacheIRCorpusJSON(right)
		if err != nil {
			t.Fatalf("canonicalize corpus right: %v", err)
		}
		if !bytes.Equal(corpusLeft, corpusRight) {
			t.Fatalf("corpus canonical JSON differs\nleft:  %s\nright: %s", corpusLeft, corpusRight)
		}
	})
}

func apalacheCorpusTargets(t *testing.T) []string {
	t.Helper()
	var targets []string
	for _, root := range []string{
		filepath.Join("test_vectors", "tla-plus-bench", "specs"),
		filepath.Join("test_vectors", "Examples", "specifications"),
	} {
		if _, err := os.Stat(root); os.IsNotExist(err) {
			continue
		} else if err != nil {
			t.Fatalf("stat %s: %v", root, err)
		}
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || filepath.Ext(path) != ".tla" {
				return nil
			}
			targets = append(targets, path)
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", root, err)
		}
	}
	sort.Strings(targets)
	return targets
}

func apalacheCorpusWindow(targets []string) []string {
	start := os.Getenv("TLAGO_APALACHE_CORPUS_START")
	if start == "" {
		return targets
	}
	for i, path := range targets {
		if path == start || strings.Contains(path, start) {
			return targets[i:]
		}
	}
	return nil
}

func compareApalacheIRCorpusFile(t *testing.T, apalache, sourcePath string) string {
	t.Helper()
	if red, err := redFileExists(sourcePath, airRedSuffix); err != nil {
		return err.Error()
	} else if red {
		return ""
	}
	want, libraryPaths, err := goldenOrApalacheIRCorpusJSON(t, apalache, sourcePath)
	if err != nil {
		return err.Error()
	}

	spec, diags := LoadSanySpec(sourcePath, LoadOptions{LibraryPaths: libraryPaths, PreferLibraryModules: true})
	if diags.HasErrors() {
		return "tlago load rejected staged spec:\n" + diags.Error()
	}
	semDiags := CheckSpec(spec)
	if semDiags.HasErrors() {
		return "tlago check rejected staged spec:\n" + semDiags.Error()
	}
	got, irDiags := apalacheIRCorpusJSON(spec)
	if irDiags.HasErrors() {
		return "tlago ApalacheIR rejected staged spec:\n" + irDiags.Error()
	}
	gotCanon, err := canonicalApalacheIRCorpusJSON(got)
	if err != nil {
		return fmt.Sprintf("canonicalize tlago JSON: %v", err)
	}
	wantCanon, err := canonicalApalacheIRCorpusJSON(want)
	if err != nil {
		return fmt.Sprintf("canonicalize Apalache JSON: %v", err)
	}
	if !bytes.Equal(gotCanon, wantCanon) {
		if err := writeApalacheCorpusArtifacts(sourcePath, got, want, gotCanon, wantCanon); err != nil {
			return fmt.Sprintf("write corpus artifacts: %v", err)
		}
		return apalacheCorpusJSONMismatch(gotCanon, wantCanon)
	}
	return ""
}

func goldenOrApalacheIRCorpusJSON(t *testing.T, apalache, sourcePath string) ([]byte, []string, error) {
	t.Helper()
	libraryPaths := apalacheCorpusDirectLibraryPaths(sourcePath)
	if data, ok, err := readGoldenFile(sourcePath, airGoldenSuffix); ok || err != nil {
		return data, libraryPaths, err
	}
	if apalache == "" {
		return nil, nil, fmt.Errorf("missing %s and APALACHE_MC/apalache-mc is unavailable; run `make -C test_vectors expected` to generate it", goldenFilePath(sourcePath, airGoldenSuffix))
	}
	data, err := apalacheIRCorpusJSONFromApalache(t, apalache, sourcePath)
	if err != nil {
		return nil, nil, err
	}
	if saved, err := saveGoldenFileIfMissing(sourcePath, airGoldenSuffix, data); err != nil {
		return nil, nil, err
	} else if saved {
		t.Logf("wrote missing ApalacheIR golden file %s", goldenFilePath(sourcePath, airGoldenSuffix))
	}
	return data, libraryPaths, nil
}

func apalacheIRCorpusJSONFromApalache(t *testing.T, apalache, sourcePath string) ([]byte, error) {
	t.Helper()
	dir := t.TempDir()
	wantPath := filepath.Join(dir, strings.TrimSuffix(filepath.Base(sourcePath), filepath.Ext(sourcePath))+".json")
	sourceDir, err := filepath.Abs(filepath.Dir(sourcePath))
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(apalache, "--out-dir="+filepath.Join(dir, "apalache-out"), "parse", "--output="+wantPath, filepath.Base(sourcePath))
	cmd.Dir = sourceDir
	cmd.Env = apalacheCorpusEnvForTest(sourcePath)
	var combined bytes.Buffer
	cmd.Stdout = &combined
	cmd.Stderr = &combined
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("Apalache rejected corpus spec: %v\n%s", err, combined.String())
	}
	want, err := os.ReadFile(wantPath)
	if err != nil {
		return nil, fmt.Errorf("read Apalache JSON: %v", err)
	}
	return want, nil
}

func apalacheCorpusEnvForTest(sourcePath string) []string {
	env := apalacheEnvForTest()
	var dirs []string
	for _, dir := range apalacheCorpusDirectLibraryPaths(sourcePath) {
		abs, err := filepath.Abs(dir)
		if err == nil {
			dirs = append(dirs, abs)
		}
	}
	joined := strings.Join(dirs, string(os.PathListSeparator))
	env = prependPathListEnv(env, "TLA_LIBRARY_PATH", joined)
	env = prependPathListEnv(env, "TLA_PATH", joined)
	return env
}

func prependPathListEnv(env []string, name, value string) []string {
	if value == "" {
		return env
	}
	prefix := name + "="
	for i, item := range env {
		if strings.HasPrefix(item, prefix) {
			if item == prefix {
				env[i] = prefix + value
			} else {
				env[i] = prefix + value + string(os.PathListSeparator) + strings.TrimPrefix(item, prefix)
			}
			return env
		}
	}
	return append(env, prefix+value)
}

func apalacheCorpusDirectLibraryPaths(sourcePath string) []string {
	paths := []string{
		filepath.Dir(sourcePath),
		filepath.Join("test_vectors", "java-sany", "StandardModules"),
		filepath.Join("test_vectors", "tlaplus-standard-modules"),
		filepath.Join("test_vectors", "CommunityModules", "modules"),
		filepath.Join("test_vectors", "CommunityModules", "tests"),
		filepath.Join("test_vectors", "tlaps-stdlib"),
		filepath.Join("test_vectors", "apalache-stdlib"),
	}
	paths = append(paths, tlaPlusBenchSpecLibraryPaths(sourcePath)...)
	sourceDir := filepath.Dir(sourcePath)
	parent := filepath.Dir(sourceDir)
	switch filepath.Base(parent) {
	case "gold", "silver":
		paths = append(paths, parent)
	}
	return paths
}

func apalacheIRCorpusJSON(spec *Spec) ([]byte, Diagnostics) {
	return ApalacheIRJSON(spec, ApalacheIROptions{IncludeSource: true})
}

func canonicalApalacheIRJSON(data []byte) ([]byte, error) {
	value, err := decodeApalacheIRJSON(data)
	if err != nil {
		return nil, err
	}
	canon := canonicalizeApalacheIRValue(value)
	var out bytes.Buffer
	writeCanonicalJSON(&out, canon)
	return out.Bytes(), nil
}

func canonicalApalacheIRCorpusJSON(data []byte) ([]byte, error) {
	value, err := decodeApalacheIRJSON(data)
	if err != nil {
		return nil, err
	}
	normalizeApalacheCorpusModuleNames(value)
	canon := canonicalizeApalacheIRValue(value)
	var out bytes.Buffer
	writeCanonicalJSON(&out, canon)
	return out.Bytes(), nil
}

func decodeApalacheIRJSON(data []byte) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	return value, nil
}

func normalizeApalacheCorpusModuleNames(value any) {
	root, ok := value.(map[string]any)
	if !ok {
		return
	}
	modules, ok := root["modules"].([]any)
	if !ok {
		return
	}
	for _, item := range modules {
		module, ok := item.(map[string]any)
		if !ok || module["kind"] != "TlaModule" {
			continue
		}
		module["name"] = "__apalache_corpus_module__"
	}
}

func canonicalizeApalacheIRValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, child := range v {
			out[key] = canonicalizeApalacheIRValue(child)
		}
		if kind, _ := out["kind"].(string); kind == "TlaModule" {
			if decls, ok := out["declarations"].([]any); ok {
				out["declarations"] = sortedCanonicalApalacheDecls(decls)
			}
		}
		if kind, _ := out["kind"].(string); kind == "LetInEx" {
			if decls, ok := out["decls"].([]any); ok {
				out["decls"] = sortedCanonicalApalacheDecls(decls)
			}
		}
		if modules, ok := out["modules"].([]any); ok {
			out["modules"] = sortedCanonicalApalacheModules(modules)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, child := range v {
			out[i] = canonicalizeApalacheIRValue(child)
		}
		return out
	default:
		return value
	}
}

func sortedCanonicalApalacheDecls(decls []any) []any {
	out := append([]any(nil), decls...)
	sort.SliceStable(out, func(i, j int) bool {
		return canonicalApalacheDeclSortKey(out[i]) < canonicalApalacheDeclSortKey(out[j])
	})
	return out
}

func sortedCanonicalApalacheModules(modules []any) []any {
	out := append([]any(nil), modules...)
	sort.SliceStable(out, func(i, j int) bool {
		return canonicalApalacheModuleSortKey(out[i]) < canonicalApalacheModuleSortKey(out[j])
	})
	return out
}

func canonicalApalacheDeclSortKey(value any) string {
	if object, ok := value.(map[string]any); ok {
		kind, _ := object["kind"].(string)
		name, _ := object["name"].(string)
		if kind != "" || name != "" {
			return kind + "\x00" + name + "\x00" + canonicalJSONSortKey(value)
		}
	}
	return canonicalJSONSortKey(value)
}

func canonicalApalacheModuleSortKey(value any) string {
	if object, ok := value.(map[string]any); ok {
		name, _ := object["name"].(string)
		if name != "" {
			return name + "\x00" + canonicalJSONSortKey(value)
		}
	}
	return canonicalJSONSortKey(value)
}

func canonicalJSONSortKey(value any) string {
	return string(canonicalJSONBytes(value))
}

func apalacheCorpusJSONMismatch(got, want []byte) string {
	offset := firstByteDiff(got, want)
	gotSnippet := corpusSnippet(got, offset)
	wantSnippet := corpusSnippet(want, offset)
	return fmt.Sprintf("canonical JSON differs at offset %d; got %d bytes, want %d bytes\n got: %s\nwant: %s", offset, len(got), len(want), gotSnippet, wantSnippet)
}

func writeApalacheCorpusArtifacts(sourcePath string, got, want, gotCanon, wantCanon []byte) error {
	root := os.Getenv("TLAGO_APALACHE_CORPUS_ARTIFACT_DIR")
	if root == "" {
		return nil
	}
	name := strings.NewReplacer("/", "__", "\\", "__", ":", "_").Replace(sourcePath)
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for _, artifact := range []struct {
		name string
		data []byte
	}{
		{name: "got.raw.json", data: got},
		{name: "want.raw.json", data: want},
		{name: "got.canonical.json", data: gotCanon},
		{name: "want.canonical.json", data: wantCanon},
	} {
		if err := os.WriteFile(filepath.Join(dir, artifact.name), artifact.data, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func firstByteDiff(left, right []byte) int {
	n := len(left)
	if len(right) < n {
		n = len(right)
	}
	for i := 0; i < n; i++ {
		if left[i] != right[i] {
			return i
		}
	}
	return n
}

func corpusSnippet(data []byte, offset int) string {
	start := offset - 120
	if start < 0 {
		start = 0
	}
	end := offset + 240
	if end > len(data) {
		end = len(data)
	}
	return strconv.Quote(string(data[start:end]))
}

func envPositiveInt(name string) int {
	raw := os.Getenv(name)
	if raw == "" {
		return 0
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0
	}
	return value
}
