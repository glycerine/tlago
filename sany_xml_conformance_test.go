package tlago

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestSanyXMLConformanceAgainstJavaSANY(t *testing.T) {
	if os.Getenv("TLAGO_SANY_XML") == "" {
		t.Skip("set TLAGO_SANY_XML=1 to compare SANY XML against Java SANY")
	}

	for _, rel := range []string{
		filepath.Join("test_vectors", "sany-xml", "Simple.tla"),
		filepath.Join("test_vectors", "sany-xml", "Exprs.tla"),
	} {
		t.Run(rel, func(t *testing.T) {
			want, err := canonicalGoldenOrJavaSanyXMLForTest(t, rel)
			if err != nil {
				t.Fatal(err)
			}

			spec, diags := LoadSanySpec(rel, sanyXMLLoadOptions())
			requireNoErrors(t, diags)
			sem := CheckSpec(spec)
			requireNoErrors(t, sem)
			gotRaw, xmlDiags := SanyXML(spec)
			requireNoErrors(t, xmlDiags)
			got, err := canonicalSanyXML(gotRaw)
			if err != nil {
				t.Fatalf("canonicalize Go SANY XML: %v\n%s", err, gotRaw)
			}

			if !bytes.Equal(got, want) {
				t.Fatalf("SANY XML differs\n%s", sanyXMLMismatch(got, want))
			}
		})
	}
}

func TestSanyXMLTargetCorpusAgainstJavaSANY(t *testing.T) {
	if os.Getenv("TLAGO_SANY_XML_CORPUS") == "" {
		t.Skip("set TLAGO_SANY_XML_CORPUS=1 to compare the target fixture trees against Java SANY XML")
	}
	targets := apalacheCorpusTargets(t)
	if len(targets) == 0 {
		t.Fatalf("no .tla corpus files found under target directories")
	}
	targets = sanyXMLCorpusWindow(targets)
	limit := envPositiveInt("TLAGO_SANY_XML_CORPUS_LIMIT")
	if limit > 0 && limit < len(targets) {
		targets = targets[:limit]
	}
	maxFailures := envPositiveInt("TLAGO_SANY_XML_CORPUS_MAX_FAILURES")
	if maxFailures == 0 {
		maxFailures = 1
	}

	var failures []string
	for i, path := range targets {
		result := compareSanyXMLCorpusFile(t, path)
		if result != "" {
			failures = append(failures, fmt.Sprintf("%d/%d %s: %s", i+1, len(targets), path, result))
			if len(failures) >= maxFailures {
				break
			}
		}
	}
	if len(failures) > 0 {
		t.Fatalf("SANY XML corpus mismatch:\n%s", strings.Join(failures, "\n\n"))
	}
}

func compareSanyXMLCorpusFile(t *testing.T, sourcePath string) string {
	t.Helper()
	if red, err := redFileExists(sourcePath, xmlRedSuffix); err != nil {
		return err.Error()
	} else if red {
		return ""
	}
	includeDirs := sanyXMLCorpusIncludeDirs(sourcePath)
	wantRaw, err := goldenOrJavaSanyXMLForTest(t, sourcePath, includeDirs...)
	if err != nil {
		return err.Error()
	}
	want, err := canonicalSanyXML(wantRaw)
	if err != nil {
		return fmt.Sprintf("canonicalize Java XML: %v", err)
	}
	spec, diags := LoadSanySpec(sourcePath, sanyXMLLoadOptions(includeDirs...))
	if diags.HasErrors() {
		return "tlago load rejected corpus spec:\n" + diags.Error()
	}
	semDiags := CheckSpec(spec)
	if semDiags.HasErrors() {
		return "tlago check rejected staged spec:\n" + semDiags.Error()
	}
	gotRaw, xmlDiags := SanyXML(spec)
	if xmlDiags.HasErrors() {
		return "tlago SANY XML rejected staged spec:\n" + xmlDiags.Error()
	}
	got, err := canonicalSanyXML(gotRaw)
	if err != nil {
		return fmt.Sprintf("canonicalize tlago XML: %v", err)
	}

	if !bytes.Equal(got, want) {
		if err := writeSanyXMLCorpusArtifacts(sourcePath, gotRaw, wantRaw, got, want); err != nil {
			return fmt.Sprintf("write corpus artifacts: %v", err)
		}
		return sanyXMLMismatch(got, want)
	}
	return ""
}

func sanyXMLCorpusWindow(targets []string) []string {
	start := os.Getenv("TLAGO_SANY_XML_CORPUS_START")
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

func canonicalGoldenOrJavaSanyXMLForTest(t *testing.T, sourcePath string, includeDirs ...string) ([]byte, error) {
	t.Helper()
	raw, err := goldenOrJavaSanyXMLForTest(t, sourcePath, includeDirs...)
	if err != nil {
		return nil, err
	}
	return canonicalSanyXML(raw)
}

func goldenOrJavaSanyXMLForTest(t *testing.T, sourcePath string, includeDirs ...string) ([]byte, error) {
	t.Helper()
	if data, ok, err := readGoldenFile(sourcePath, xmlGoldenSuffix); ok || err != nil {
		return data, err
	}
	data, err := javaSanyXMLForTest(t, sourcePath, includeDirs...)
	if err != nil {
		return nil, err
	}
	if saved, err := saveGoldenFileIfMissing(sourcePath, xmlGoldenSuffix, data); err != nil {
		return nil, err
	} else if saved {
		t.Logf("wrote missing SANY XML golden file %s", goldenFilePath(sourcePath, xmlGoldenSuffix))
	}
	return data, nil
}

func javaSanyXMLForTest(t *testing.T, specPath string, includeDirs ...string) ([]byte, error) {
	t.Helper()
	classpath := javaSanyClasspathForTest(t)
	if classpath == "" {
		return nil, fmt.Errorf("missing %s and TLA2TOOLS_JAR/test_vectors/java-sany/tla2tools.jar is unavailable; run `make -C test_vectors expected-xml` to generate it", goldenFilePath(specPath, xmlGoldenSuffix))
	}

	args := []string{"-cp", classpath, "tla2sany.xml.XMLExporter", "-o"}
	for _, dir := range append([]string{sanyXMLJavaStandardModulesDir()}, includeDirs...) {
		abs, err := filepath.Abs(dir)
		if err != nil {
			return nil, err
		}
		args = append(args, "-I", abs)
	}
	args = append(args, filepath.Base(specPath))
	cmd := exec.Command("java", args...)
	cmd.Dir = filepath.Dir(specPath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("Java SANY XMLExporter failed: %v\nstdout:\n%s\nstderr:\n%s", err, stdout.String(), stderr.String())
	}
	xmlStart := bytes.Index(stdout.Bytes(), []byte("<?xml"))
	if xmlStart < 0 {
		return nil, fmt.Errorf("Java SANY XMLExporter did not print XML on stdout:\nstdout:\n%s\nstderr:\n%s", stdout.String(), stderr.String())
	}
	return append([]byte(nil), stdout.Bytes()[xmlStart:]...), nil
}

func javaSanyJarForTest(t *testing.T) string {
	t.Helper()
	if env := os.Getenv("TLA2TOOLS_JAR"); env != "" {
		return env
	}
	jar := filepath.Join("test_vectors", "java-sany", "tla2tools.jar")
	if _, err := os.Stat(jar); err == nil {
		abs, absErr := filepath.Abs(jar)
		if absErr != nil {
			t.Fatalf("abs %s: %v", jar, absErr)
		}
		return abs
	}
	return ""
}

func javaSanyClasspathForTest(t *testing.T) string {
	t.Helper()
	jar := javaSanyJarForTest(t)
	if jar == "" || os.Getenv("TLA2TOOLS_JAR") != "" {
		return jar
	}
	dependency := filepath.Join(filepath.Dir(jar), "CommunityModules.jar")
	if _, err := os.Stat(dependency); err != nil {
		t.Fatalf("missing frozen Java SANY dependency %s: %v", dependency, err)
	}
	classpath := []string{jar, dependency}
	tlapsDependency := filepath.Join(sanyXMLTLAPSStdlibDir(), "CommunityModules.jar")
	if _, err := os.Stat(tlapsDependency); err == nil {
		classpath = append(classpath, tlapsDependency)
	}
	return strings.Join(classpath, string(os.PathListSeparator))
}

func sanyXMLJavaStandardModulesDir() string {
	return filepath.Join("test_vectors", "java-sany", "StandardModules")
}

func sanyXMLTlaplusStandardModulesDir() string {
	return filepath.Join("test_vectors", "tlaplus-standard-modules")
}

func sanyXMLCommunityModulesDir() string {
	return filepath.Join("test_vectors", "CommunityModules", "modules")
}

func sanyXMLCommunityModulesTestsDir() string {
	return filepath.Join("test_vectors", "CommunityModules", "tests")
}

func sanyXMLTLAPSStdlibDir() string {
	return filepath.Join("test_vectors", "tlaps-stdlib")
}

func sanyXMLApalacheStdlibDir() string {
	return filepath.Join("test_vectors", "apalache-stdlib")
}

func sanyXMLCorpusIncludeDirs(sourcePath string) []string {
	dirs := []string{sanyXMLTlaplusStandardModulesDir(), sanyXMLCommunityModulesDir(), sanyXMLCommunityModulesTestsDir(), sanyXMLTLAPSStdlibDir(), sanyXMLApalacheStdlibDir()}
	dirs = append(dirs, tlaPlusBenchSpecLibraryPaths(sourcePath)...)
	sourceDir := filepath.Dir(sourcePath)
	parent := filepath.Dir(sourceDir)
	switch filepath.Base(parent) {
	case "gold", "silver":
		dirs = append(dirs, parent)
	}
	return dirs
}

func sanyXMLLoadOptions(includeDirs ...string) LoadOptions {
	libraryPaths := append([]string{sanyXMLJavaStandardModulesDir()}, includeDirs...)
	return LoadOptions{LibraryPaths: libraryPaths, PreferLibraryModules: true}
}

type canonicalXMLNode struct {
	Name     string
	Attrs    []xml.Attr
	Text     string
	Children []*canonicalXMLNode
}

func canonicalSanyXML(data []byte) ([]byte, error) {
	root, err := parseCanonicalXML(data)
	if err != nil {
		return nil, err
	}
	uidMap := canonicalSanyXMLUIDMap(root)
	rewriteCanonicalSanyXMLUIDs(root, uidMap)
	sortCanonicalSanyXML(root)
	var out bytes.Buffer
	writeCanonicalXMLNode(&out, root)
	return out.Bytes(), nil
}

func parseCanonicalXML(data []byte) (*canonicalXMLNode, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	var stack []*canonicalXMLNode
	var root *canonicalXMLNode
	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch token := tok.(type) {
		case xml.StartElement:
			node := &canonicalXMLNode{Name: token.Name.Local, Attrs: append([]xml.Attr(nil), token.Attr...)}
			if len(stack) == 0 {
				root = node
			} else {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, node)
			}
			stack = append(stack, node)
		case xml.CharData:
			if len(stack) == 0 {
				continue
			}
			text := strings.TrimSpace(string(token))
			if text != "" {
				stack[len(stack)-1].Text += text
			}
		case xml.EndElement:
			if len(stack) == 0 {
				return nil, fmt.Errorf("unexpected closing tag %s", token.Name.Local)
			}
			stack = stack[:len(stack)-1]
		}
	}
	if root == nil {
		return nil, fmt.Errorf("empty XML document")
	}
	return root, nil
}

func canonicalSanyXMLUIDMap(root *canonicalXMLNode) map[string]string {
	entries := canonicalSanyXMLEntries(root)
	out := map[string]string{}
	for _, entry := range entries {
		child := canonicalSanyXMLEntryPayload(entry)
		if child == nil {
			continue
		}
		registerOwnedFormalParamUIDs(out, child)
	}
	type uidEntry struct {
		uid string
		key string
	}
	byKey := map[string][]uidEntry{}
	for _, entry := range entries {
		uid := firstChildText(entry, "UID")
		if uid == "" || out[uid] != "" {
			continue
		}
		key := canonicalSanyXMLEntryKey(entry)
		byKey[key] = append(byKey[key], uidEntry{uid: uid, key: key})
	}
	keys := make([]string, 0, len(byKey))
	for key := range byKey {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		items := byKey[key]
		sort.SliceStable(items, func(i, j int) bool {
			return items[i].uid < items[j].uid
		})
		for i, item := range items {
			out[item.uid] = canonicalSanyXMLUID(key, i, len(items))
		}
	}
	return out
}

func canonicalSanyXMLUID(key string, index, count int) string {
	sum := sha256.Sum256([]byte(key))
	id := hex.EncodeToString(sum[:8])
	if count <= 1 {
		return "uid:" + id
	}
	return fmt.Sprintf("uid:%s:%03d", id, index)
}

func registerOwnedFormalParamUIDs(out map[string]string, node *canonicalXMLNode) {
	switch node.Name {
	case "BuiltInKind", "UserDefinedOpKind", "LabelNode":
	default:
		return
	}
	owner := node.Name + ":" + firstChildText(node, "uniquename") + ":" + firstDescendantText(node, "filename")
	for _, params := range directChildren(node, "params") {
		index := 0
		for _, item := range params.Children {
			uid := firstDescendantText(item, "UID")
			if uid == "" {
				continue
			}
			out[uid] = fmt.Sprintf("uid:param:%s:%03d", owner, index)
			index++
		}
	}
}

func canonicalSanyXMLEntries(root *canonicalXMLNode) []*canonicalXMLNode {
	var entries []*canonicalXMLNode
	var walk func(*canonicalXMLNode)
	walk = func(node *canonicalXMLNode) {
		if node == nil {
			return
		}
		if node.Name == "entry" {
			entries = append(entries, node)
		}
		for _, child := range node.Children {
			walk(child)
		}
	}
	walk(root)
	return entries
}

func canonicalSanyXMLEntryKey(entry *canonicalXMLNode) string {
	child := canonicalSanyXMLEntryPayload(entry)
	if child != nil {
		return child.Name + ":" + firstDescendantText(child, "uniquename") + ":" + firstDescendantText(child, "filename") + ":" + canonicalXMLSortKeyIgnoringUIDs(child)
	}
	return canonicalXMLSortKeyIgnoringUIDs(entry)
}

func canonicalSanyXMLEntryPayload(entry *canonicalXMLNode) *canonicalXMLNode {
	for _, child := range entry.Children {
		if child.Name != "UID" {
			return child
		}
	}
	return nil
}

func rewriteCanonicalSanyXMLUIDs(node *canonicalXMLNode, uidMap map[string]string) {
	if node == nil {
		return
	}
	if node.Name == "UID" {
		if mapped := uidMap[node.Text]; mapped != "" {
			node.Text = mapped
		}
	}
	for _, child := range node.Children {
		rewriteCanonicalSanyXMLUIDs(child, uidMap)
	}
}

func sortCanonicalSanyXML(node *canonicalXMLNode) {
	if node == nil {
		return
	}
	for _, child := range node.Children {
		sortCanonicalSanyXML(child)
	}
	sort.Slice(node.Attrs, func(i, j int) bool {
		if node.Attrs[i].Name.Local == node.Attrs[j].Name.Local {
			return node.Attrs[i].Value < node.Attrs[j].Value
		}
		return node.Attrs[i].Name.Local < node.Attrs[j].Name.Local
	})
	if node.Name == "context" {
		sort.SliceStable(node.Children, func(i, j int) bool {
			return canonicalXMLSortKey(node.Children[i]) < canonicalXMLSortKey(node.Children[j])
		})
	}
	if node.Name == "modules" {
		sortModuleNodeRefs(node)
	}
}

func sortModuleNodeRefs(node *canonicalXMLNode) {
	if len(node.Children) <= 3 {
		return
	}
	prefix := append([]*canonicalXMLNode(nil), node.Children[:2]...)
	refs := append([]*canonicalXMLNode(nil), node.Children[2:]...)
	sort.SliceStable(refs, func(i, j int) bool {
		return canonicalXMLSortKey(refs[i]) < canonicalXMLSortKey(refs[j])
	})
	node.Children = append(prefix, refs...)
}

func writeCanonicalXMLNode(b *bytes.Buffer, node *canonicalXMLNode) {
	b.WriteByte('<')
	b.WriteString(node.Name)
	for _, attr := range node.Attrs {
		b.WriteByte(' ')
		b.WriteString(attr.Name.Local)
		b.WriteString(`="`)
		xml.EscapeText(b, []byte(attr.Value))
		b.WriteByte('"')
	}
	b.WriteByte('>')
	if node.Text != "" {
		xml.EscapeText(b, []byte(node.Text))
	}
	for _, child := range node.Children {
		writeCanonicalXMLNode(b, child)
	}
	b.WriteString("</")
	b.WriteString(node.Name)
	b.WriteByte('>')
}

func canonicalXMLSortKey(node *canonicalXMLNode) string {
	var b bytes.Buffer
	writeCanonicalXMLNode(&b, node)
	return b.String()
}

func canonicalXMLSortKeyIgnoringUIDs(node *canonicalXMLNode) string {
	clone := cloneCanonicalXMLNode(node)
	replaceUIDText(clone, "#")
	var b bytes.Buffer
	writeCanonicalXMLNode(&b, clone)
	return b.String()
}

func cloneCanonicalXMLNode(node *canonicalXMLNode) *canonicalXMLNode {
	if node == nil {
		return nil
	}
	out := &canonicalXMLNode{Name: node.Name, Text: node.Text, Attrs: append([]xml.Attr(nil), node.Attrs...)}
	for _, child := range node.Children {
		out.Children = append(out.Children, cloneCanonicalXMLNode(child))
	}
	return out
}

func replaceUIDText(node *canonicalXMLNode, text string) {
	if node == nil {
		return
	}
	if node.Name == "UID" {
		node.Text = text
	}
	for _, child := range node.Children {
		replaceUIDText(child, text)
	}
}

func firstChildText(node *canonicalXMLNode, name string) string {
	for _, child := range node.Children {
		if child.Name == name {
			return child.Text
		}
	}
	return ""
}

func directChildren(node *canonicalXMLNode, name string) []*canonicalXMLNode {
	var out []*canonicalXMLNode
	for _, child := range node.Children {
		if child.Name == name {
			out = append(out, child)
		}
	}
	return out
}

func firstDescendantText(node *canonicalXMLNode, name string) string {
	if node == nil {
		return ""
	}
	if node.Name == name {
		return node.Text
	}
	for _, child := range node.Children {
		if text := firstDescendantText(child, name); text != "" {
			return text
		}
	}
	return ""
}

func sanyXMLMismatch(got, want []byte) string {
	offset := firstByteDiff(got, want)
	gotSnippet := corpusSnippet(got, offset)
	wantSnippet := corpusSnippet(want, offset)
	return fmt.Sprintf("canonical XML differs at offset %d; got %d bytes, want %d bytes\n got: %s\nwant: %s", offset, len(got), len(want), gotSnippet, wantSnippet)
}

func writeSanyXMLCorpusArtifacts(sourcePath string, got, want, gotCanon, wantCanon []byte) error {
	root := os.Getenv("TLAGO_SANY_XML_CORPUS_ARTIFACT_DIR")
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
		{name: "got.raw.xml", data: got},
		{name: "want.raw.xml", data: want},
		{name: "got.canonical.xml", data: gotCanon},
		{name: "want.canonical.xml", data: wantCanon},
	} {
		if err := os.WriteFile(filepath.Join(dir, artifact.name), artifact.data, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func TestCanonicalSanyXMLBehavior(t *testing.T) {
	left := []byte(`<modules><RootModule>M</RootModule><context><entry><UID>2</UID><BuiltInKind><uniquename>TRUE</uniquename></BuiltInKind></entry><entry><UID>9</UID><ModuleNode><uniquename>M</uniquename><BuiltInKindRef><UID>2</UID></BuiltInKindRef></ModuleNode></entry></context><ModuleNodeRef><UID>9</UID></ModuleNodeRef></modules>`)
	right := []byte(`<modules><RootModule>M</RootModule><context><entry><UID>100</UID><ModuleNode><uniquename>M</uniquename><BuiltInKindRef><UID>1</UID></BuiltInKindRef></ModuleNode></entry><entry><UID>1</UID><BuiltInKind><uniquename>TRUE</uniquename></BuiltInKind></entry></context><ModuleNodeRef><UID>100</UID></ModuleNodeRef></modules>`)
	leftCanon, err := canonicalSanyXML(left)
	if err != nil {
		t.Fatalf("canonicalize left: %v", err)
	}
	rightCanon, err := canonicalSanyXML(right)
	if err != nil {
		t.Fatalf("canonicalize right: %v", err)
	}
	if !bytes.Equal(leftCanon, rightCanon) {
		t.Fatalf("canonical XML differs\nleft:  %s\nright: %s", leftCanon, rightCanon)
	}
}
