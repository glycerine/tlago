package tlago

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/glycerine/tlago/tlc"
)

const (
	replSpecName  = "tlarepl"
	replValueName = "replvalue"
)

type REPLEvalOptions struct {
	TempDir              string
	SpecFile             string
	LibraryPaths         []string
	PreferLibraryModules bool
	PrintOutput          io.Writer
	KeepTempFiles        bool
}

func EvaluateREPLExpression(expr string, opts REPLEvalOptions) (string, Diagnostics, error) {
	tempDir := opts.TempDir
	removeTempDir := false
	if tempDir == "" {
		dir, err := os.MkdirTemp("", replSpecName)
		if err != nil {
			return "", nil, err
		}
		tempDir = dir
		removeTempDir = !opts.KeepTempFiles
	}
	if removeTempDir {
		defer os.RemoveAll(tempDir)
	}
	if err := os.MkdirAll(tempDir, 0o755); err != nil {
		return "", nil, err
	}

	extends := []string{"Naturals", "Reals", "Sequences", "Bags", "FiniteSets", "TLC"}
	libraryPaths := append([]string(nil), opts.LibraryPaths...)
	if opts.SpecFile != "" {
		specModule := strings.TrimSuffix(filepath.Base(opts.SpecFile), filepath.Ext(opts.SpecFile))
		if specModule != "" {
			extends = append(extends, specModule)
		}
		if specDir := filepath.Dir(opts.SpecFile); specDir != "." && specDir != "" {
			libraryPaths = append([]string{specDir}, libraryPaths...)
		}
	}

	specPath := filepath.Join(tempDir, replSpecName+".tla")
	cfgPath := filepath.Join(tempDir, replSpecName+".cfg")
	specSource := strings.Join([]string{
		"---- MODULE " + replSpecName + " ----",
		"EXTENDS " + strings.Join(extends, ", "),
		"VARIABLE replvar",
		"replinit == replvar = 0",
		"replnext == replvar' = 0",
		replValueName + " == " + expr,
		"====",
		"",
	}, "\n")
	cfgSource := "INIT replinit\nNEXT replnext\n"
	if err := os.WriteFile(specPath, []byte(specSource), 0o644); err != nil {
		return "", nil, err
	}
	if err := os.WriteFile(cfgPath, []byte(cfgSource), 0o644); err != nil {
		return "", nil, err
	}

	spec, diags := LoadSanySpec(specPath, LoadOptions{
		LibraryPaths:         libraryPaths,
		PreferLibraryModules: opts.PreferLibraryModules,
	})
	if diags.HasErrors() {
		return "", diags, nil
	}
	sem := CheckSpec(spec)
	diags = append(diags, sem...)
	if sem.HasErrors() {
		return "", diags, nil
	}
	cfg, err := tlc.ParseModelConfigSource(cfgPath, cfgSource)
	if err != nil {
		return "", diags, err
	}
	oldOutput := tlc.TLCOutput
	oldOutputToUserFile := tlc.TLCOutputToUserFile
	tlc.TLCOutput = io.Discard
	tlc.TLCOutputToUserFile = false
	defer func() {
		tlc.TLCOutput = oldOutput
		tlc.TLCOutputToUserFile = oldOutputToUserFile
	}()

	tool, toolDiags := BuildTLCTool(spec, cfg, tlc.RuntimeParameters{})
	diags = append(diags, toolDiags...)
	if toolDiags.HasErrors() {
		return "", diags, nil
	}
	if fallbackDiags := installREPLRootDefinitions(tool, spec, cfg); fallbackDiags.HasErrors() {
		diags = append(diags, fallbackDiags...)
		return "", diags, nil
	}
	valueDef, ok := lookupREPLValueDefinition(tool)
	if !ok || valueDef == nil {
		return "", diags, fmt.Errorf("REPL value definition %s not found; root definitions: %s; available definitions: %s", replValueName, strings.Join(moduleDefinitionNamesForDebug(spec.Root), ", "), strings.Join(replDefinitionNames(tool), ", "))
	}

	if opts.PrintOutput != nil {
		tlc.TLCOutput = opts.PrintOutput
	} else {
		tlc.TLCOutput = oldOutput
	}
	tlc.TLCOutputToUserFile = false
	value, err := tool.Eval(valueDef.Body)
	if err != nil {
		return "", diags, err
	}
	return value.String(), diags, nil
}

func installREPLRootDefinitions(tool *tlc.Tool, spec *Spec, cfg *tlc.ModelConfig) Diagnostics {
	if tool == nil || spec == nil || spec.Root == nil {
		return nil
	}
	var defnStore *tlc.Defns
	if tool.SpecProcessor != nil {
		defnStore = tool.SpecProcessor.Defns
	}
	bridge := &tlcBridge{
		tool:                  tool,
		processor:             tool.SpecProcessor,
		spec:                  spec,
		cfg:                   cfg,
		defs:                  tlcBridgeDefinitionsByName(spec),
		defns:                 defnStore,
		symbols:               map[string]*tlc.SymbolNode{},
		rootModuleName:        spec.Root.Name,
		moduleDefinitionNames: moduleDefinitionNameIndex(spec),
		convertBoundNames:     map[string]int{},
	}
	for i := range spec.Root.Definitions {
		def := &spec.Root.Definitions[i]
		opDef := bridge.convertDefinitionAs(def.Name, def)
		if opDef != nil {
			bridge.define(opDef.Symbol, opDef)
			if spec.Root.Name != "" {
				bridge.defineName(spec.Root.Name+"!"+def.Name, opDef)
			}
		}
	}
	return bridge.diags
}

func moduleDefinitionNamesForDebug(mod *Module) []string {
	if mod == nil || len(mod.Definitions) == 0 {
		return nil
	}
	names := make([]string, 0, len(mod.Definitions))
	for i := range mod.Definitions {
		names = append(names, mod.Definitions[i].Name)
	}
	sort.Strings(names)
	return names
}

func replDefinitionNames(tool *tlc.Tool) []string {
	if tool == nil || len(tool.DefnsByName) == 0 {
		return nil
	}
	names := make([]string, 0, len(tool.DefnsByName))
	for name := range tool.DefnsByName {
		if name != nil {
			names = append(names, name.String())
		}
	}
	sort.Strings(names)
	if len(names) > 24 {
		names = append(names[:24], "...")
	}
	return names
}

func lookupREPLValueDefinition(tool *tlc.Tool) (*tlc.OpDefNode, bool) {
	if tool == nil {
		return nil, false
	}
	for _, name := range []string{replValueName, replSpecName + "!" + replValueName} {
		if valueDef, ok := tool.DefnsByName[tlc.UniqueStringOf(name)].(*tlc.OpDefNode); ok {
			return valueDef, true
		}
	}
	for name, value := range tool.DefnsByName {
		if name == nil {
			continue
		}
		text := name.String()
		if text == replValueName || strings.HasSuffix(text, "!"+replValueName) {
			valueDef, ok := value.(*tlc.OpDefNode)
			return valueDef, ok
		}
	}
	return nil, false
}
