// Copyright (c) 2024, Oracle and/or its affiliates.
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
	specFileSet          bool
	LibraryPaths         []string
	PreferLibraryModules bool
	PrintOutput          io.Writer
	KeepTempFiles        bool
}

func EvaluateREPLExpression(expr string, opts REPLEvalOptions) (string, Diagnostics, error) {
	return evaluateREPLExpression(expr, opts, false, nil)
}

func evaluateREPLExpression(expr string, opts REPLEvalOptions, processInput bool, innerStarted *bool) (string, Diagnostics, error) {
	tempDir := opts.TempDir
	removeTempDir := false
	if tempDir == "" && !processInput {
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
	if !processInput {
		if err := os.MkdirAll(tempDir, 0o755); err != nil {
			return "", nil, err
		}
	}

	extends := []string{"Reals", "Sequences", "Bags", "FiniteSets", "TLC", "Randomization"}
	if popular, available := replPopularModules(); available {
		extends = append(extends, strings.Split(popular, ",")...)
	}
	libraryPaths := append([]string(nil), opts.LibraryPaths...)
	if opts.SpecFile != "" || opts.specFileSet {
		specModule := ""
		if opts.SpecFile != "" {
			specModule = strings.TrimSuffix(filepath.Base(opts.SpecFile), ".tla")
		}
		if processInput || specModule != "" {
			extends = append(extends, specModule)
		}
		if specDir := filepath.Dir(opts.SpecFile); !processInput && specDir != "." && specDir != "" {
			libraryPaths = append([]string{specDir}, libraryPaths...)
		}
	}

	specPath := filepath.Join(tempDir, replSpecName+".tla")
	cfgPath := filepath.Join(tempDir, replSpecName+".cfg")
	specSource := strings.Join([]string{
		"---- MODULE " + replSpecName + " ----",
		"EXTENDS " + strings.Join(extends, ","),
		"VARIABLE replvar",
		"replinit == replvar = 0",
		"replnext == replvar' = 0",
		replValueName + " == " + expr,
		"====",
		"",
	}, "\n")
	cfgSource := "INIT replinit\nNEXT replnext\n"
	if err := os.WriteFile(cfgPath, []byte(cfgSource), 0o644); err != nil {
		return "", nil, err
	}
	if err := os.WriteFile(specPath, []byte(specSource), 0o644); err != nil {
		return "", nil, err
	}
	if processInput {
		tlc.ToolIOSetMode(tlc.ToolIOTool)
		tlc.ToolIOReset()
		*innerStarted = true
	}

	oldOutput := tlc.TLCOutput
	oldOutputToUserFile := tlc.TLCOutputToUserFile
	if !processInput {
		tlc.TLCOutput = io.Discard
		tlc.TLCOutputToUserFile = false
		defer func() {
			tlc.TLCOutput = oldOutput
			tlc.TLCOutputToUserFile = oldOutputToUserFile
		}()
	}

	spec, diags := LoadSanySpec(specPath, LoadOptions{
		LibraryPaths:         libraryPaths,
		PreferLibraryModules: opts.PreferLibraryModules,
	})
	if diags.HasErrors() {
		if processInput {
			return "", diags, replParsingFailure(diags)
		}
		return "", diags, nil
	}
	sem := CheckSpec(spec)
	diags = append(diags, sem...)
	if sem.HasErrors() {
		if processInput {
			return "", diags, replParsingFailure(diags)
		}
		return "", diags, nil
	}
	cfg, err := tlc.ParseModelConfigSource(cfgPath, cfgSource)
	if err != nil {
		return "", diags, err
	}

	tool, toolDiags := BuildTLCTool(spec, cfg, tlc.RuntimeParameters{})
	diags = append(diags, toolDiags...)
	if toolDiags.HasErrors() {
		if processInput {
			return "", diags, replParsingFailure(diags)
		}
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
	if processor := tool.GetSpecProcessor(); processor != nil && processor.GetRootModule() != nil {
		for _, def := range processor.GetRootModule().GetOpDefs() {
			if def.Name.String() == replValueName {
				return def, true
			}
		}
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
