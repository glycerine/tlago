package tlago

import (
	"fmt"

	"github.com/glycerine/tlago/tlc"
)

// NewCheckImplFileTraceLoader returns the root-package bridge needed by
// tlc.CheckImplFile without letting the tlc package import the SANY parser.
func NewCheckImplFileTraceLoader(tool *tlc.Tool, loadOpts LoadOptions) func(string) ([]*tlc.TLCStateMut, error) {
	return func(filename string) ([]*tlc.TLCStateMut, error) {
		return LoadCheckImplFileTraceStates(tool, filename, loadOpts)
	}
}

func NewCheckImplFileTraceIntoLoader(tool *tlc.Tool, loadOpts LoadOptions) func(string, func([]*tlc.TLCStateMut)) error {
	return func(filename string, publish func([]*tlc.TLCStateMut)) error {
		return LoadCheckImplFileTraceStatesInto(tool, filename, loadOpts, publish)
	}
}

func LoadCheckImplFileTraceStates(tool *tlc.Tool, filename string, loadOpts LoadOptions) ([]*tlc.TLCStateMut, error) {
	var states []*tlc.TLCStateMut
	err := LoadCheckImplFileTraceStatesInto(tool, filename, loadOpts, func(allocated []*tlc.TLCStateMut) {
		states = allocated
	})
	return states, err
}

func LoadCheckImplFileTraceStatesInto(tool *tlc.Tool, filename string, loadOpts LoadOptions, publish func([]*tlc.TLCStateMut)) error {
	if tool == nil {
		return fmt.Errorf("CheckImplFile trace loader has no TLC tool")
	}
	// Java checks the extensionless file as the ready marker, then gives that
	// name to SANY, whose resolver reads the corresponding .tla module.
	spec, diags, err := loadCheckImplFileTraceSpec(filename, loadOpts)
	if err != nil {
		return err
	}
	if diags.HasErrors() {
		return tlc.NewTLCRuntimeException(tlc.ECTLCParsingFailed)
	}
	return CheckImplFileTraceStatesInto(tool, spec, publish)
}

func loadCheckImplFileTraceSpec(filename string, loadOpts LoadOptions) (spec *Spec, diags Diagnostics, err error) {
	defer func() {
		if failure := recover(); failure != nil {
			message := fmt.Sprint(failure)
			if failureErr, ok := failure.(error); ok {
				message = failureErr.Error()
			}
			spec = nil
			diags = nil
			err = tlc.NewTLCRuntimeException(tlc.ECCheckCouldNotReadTrace, message)
		}
	}()
	spec, diags = LoadSanySpec(filename, loadOpts)
	if !diags.HasErrors() {
		diags = append(diags, CheckSpec(spec)...)
	}
	return spec, diags, nil
}

func CheckImplFileTraceStates(tool *tlc.Tool, spec *Spec) ([]*tlc.TLCStateMut, error) {
	var states []*tlc.TLCStateMut
	err := CheckImplFileTraceStatesInto(tool, spec, func(allocated []*tlc.TLCStateMut) {
		states = allocated
	})
	return states, err
}

func CheckImplFileTraceStatesInto(tool *tlc.Tool, spec *Spec, publish func([]*tlc.TLCStateMut)) error {
	if tool == nil {
		return fmt.Errorf("CheckImplFile trace loader has no TLC tool")
	}
	if tool.GetSpecProcessor() == nil {
		return fmt.Errorf("CheckImplFile trace loader has no TLC spec processor")
	}
	if spec == nil || spec.Root == nil {
		return fmt.Errorf("CheckImplFile trace spec has no root module")
	}
	symbols := map[string]*tlc.SymbolNode{}
	for _, variable := range tool.GetSpecProcessor().GetVariablesNodes() {
		if variable != nil && variable.Name != nil {
			symbols[variable.Name.String()] = variable
		}
	}
	bridge := &tlcBridge{
		tool:                  tool,
		processor:             tool.GetSpecProcessor(),
		spec:                  spec,
		cfg:                   tool.GetModelConfig(),
		defs:                  tlcBridgeDefinitionsByName(spec),
		symbols:               symbols,
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
		return bridge.diags
	}
	restore := installTemporaryTraceDefinitions(tool, opDefs)
	defer restore()

	states := make([]*tlc.TLCStateMut, len(opDefs))
	if publish == nil {
		panic(tlc.NewNullPointerException())
	}
	publish(states)
	for i, opDef := range opDefs {
		body := opDef.GetBody()
		tool.GetSpecProcessor().ProcessConstants(body, tool.GetSpecProcessor().GetDefns())
		state, err := tool.MakeState(body)
		if err != nil {
			return err
		}
		states[i] = state
	}
	return nil
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
