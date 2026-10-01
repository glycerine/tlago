package tlago

import (
	"fmt"
	"os"

	"github.com/glycerine/tlago/tlc"
)

// NewCheckImplFileTraceLoader returns the root-package bridge needed by
// tlc.CheckImplFile without letting the tlc package import the SANY parser.
func NewCheckImplFileTraceLoader(tool *tlc.Tool, loadOpts LoadOptions) func(string) ([]*tlc.TLCStateMut, error) {
	return func(filename string) ([]*tlc.TLCStateMut, error) {
		return LoadCheckImplFileTraceStates(tool, filename, loadOpts)
	}
}

func LoadCheckImplFileTraceStates(tool *tlc.Tool, filename string, loadOpts LoadOptions) ([]*tlc.TLCStateMut, error) {
	if tool == nil {
		return nil, fmt.Errorf("CheckImplFile trace loader has no TLC tool")
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	spec, diags := CheckSanySourceWithOptions(filename, string(data), loadOpts)
	if diags.HasErrors() {
		return nil, diags
	}
	return CheckImplFileTraceStates(tool, spec)
}

func CheckImplFileTraceStates(tool *tlc.Tool, spec *Spec) ([]*tlc.TLCStateMut, error) {
	if tool == nil {
		return nil, fmt.Errorf("CheckImplFile trace loader has no TLC tool")
	}
	if spec == nil || spec.Root == nil {
		return nil, fmt.Errorf("CheckImplFile trace spec has no root module")
	}
	bridge := &tlcBridge{
		tool:                  tool,
		spec:                  spec,
		cfg:                   tool.ModelConfig,
		defs:                  tlcBridgeDefinitionsByName(spec),
		symbols:               map[string]*tlc.SymbolNode{},
		rootModuleName:        spec.Root.Name,
		moduleDefinitionNames: moduleDefinitionNameIndex(spec),
		convertBoundNames:     map[string]int{},
	}
	opDefs := make([]*tlc.OpDefNode, 0, len(spec.Root.Definitions))
	for i := range spec.Root.Definitions {
		def := &spec.Root.Definitions[i]
		opDef := bridge.convertDefinitionAs(def.Name, def)
		if opDef != nil {
			opDefs = append(opDefs, opDef)
		}
	}
	if bridge.diags.HasErrors() {
		return nil, bridge.diags
	}
	restore := installTemporaryTraceDefinitions(tool, opDefs)
	defer restore()

	states := make([]*tlc.TLCStateMut, len(opDefs))
	for i, opDef := range opDefs {
		state, err := tool.MakeState(opDef.Body)
		if err != nil {
			return nil, err
		}
		states[i] = state
	}
	return states, nil
}

func installTemporaryTraceDefinitions(tool *tlc.Tool, defs []*tlc.OpDefNode) func() {
	if tool == nil {
		return func() {}
	}
	if tool.Definitions == nil {
		tool.Definitions = make(map[*tlc.SymbolNode]any)
	}
	if tool.DefnsByName == nil {
		tool.DefnsByName = make(map[*tlc.UniqueString]any)
	}
	type savedDefinition struct {
		sym       *tlc.SymbolNode
		name      *tlc.UniqueString
		bySym     any
		hasBySym  bool
		byName    any
		hasByName bool
	}
	saved := make([]savedDefinition, 0, len(defs))
	for _, def := range defs {
		if def == nil || def.Symbol == nil {
			continue
		}
		entry := savedDefinition{sym: def.Symbol, name: def.Symbol.Name}
		entry.bySym, entry.hasBySym = tool.Definitions[def.Symbol]
		if def.Symbol.Name != nil {
			entry.byName, entry.hasByName = tool.DefnsByName[def.Symbol.Name]
		}
		saved = append(saved, entry)
		tool.Define(def.Symbol, def)
	}
	return func() {
		for i := len(saved) - 1; i >= 0; i-- {
			entry := saved[i]
			if entry.hasBySym {
				tool.Definitions[entry.sym] = entry.bySym
			} else {
				delete(tool.Definitions, entry.sym)
			}
			if entry.name != nil {
				if entry.hasByName {
					tool.DefnsByName[entry.name] = entry.byName
				} else {
					delete(tool.DefnsByName, entry.name)
				}
			}
		}
	}
}
