package tlago

import (
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/glycerine/tlago/tlc"
)

type tlcBridge struct {
	tool                   *tlc.Tool
	processor              *tlc.SpecProcessor
	spec                   *Spec
	cfg                    *tlc.ModelConfig
	runtime                tlc.RuntimeParameters
	defs                   map[string]*Definition
	defns                  *tlc.Defns
	diags                  Diagnostics
	symbols                map[string]*tlc.SymbolNode
	rootModuleName         string
	moduleDefinitionNames  map[string]map[string]bool
	convertingModule       string
	convertBoundNames      map[string]int
	definitionModules      map[*Definition]string
	sourceSymbols          map[*Definition]*tlc.SymbolNode
	sourceDefinitions      map[*Definition]*tlc.OpDefNode
	instanceDefinitions    map[string]*tlcBridgeInstance
	nativeDefinitions      map[*tlc.UniqueString]any
	builtinDefinitions     map[string]*tlc.OpDefNode
	moduleNodes            map[*Module]*tlc.ModuleNode
	localModuleDefinitions map[string]*tlc.OpDefNode
	indexedModules         map[*Module]bool
	assumptionModules      map[*Module]bool
}

type tlcBridgeInstance struct {
	owner  *Module
	inst   Instance
	substs []tlc.Subst
	params []*tlc.SymbolNode
	defs   map[*Definition]*tlc.OpDefNode
}

type tlcBridgeInstanceTarget struct {
	name  string
	arity int
	sym   *tlc.SymbolNode
}

var bridgeStandardModuleMembers = map[string][]string{
	"Naturals": {
		"Nat", "+", "-", "*", "^", "<", ">", "\\leq", "\\geq", "%", "\\div", "..",
	},
	"Integers": {
		"Int", "Nat", "+", "-", "*", "^", "<", ">", "\\leq", "\\geq", "%", "\\div", "..", "-.",
	},
	"Sequences": {
		"Seq", "Len", "Head", "Tail", "Append", "Concat", "\\o",
		"SubSeq", "SelectSeq",
	},
	"FiniteSets": {
		"IsFiniteSet", "Cardinality",
	},
	"Strings": {
		"STRING",
	},
	"AnySet": {
		"ANY", "Any",
	},
	"Bags": {
		"EmptyBag", "IsABag", "BagCardinality", "BagIn", "CopiesIn", "BagCup", "\\oplus",
		"BagDiff", "\\ominus", "BagUnion", "SqSubseteq", "\\sqsubseteq", "BagOfAll",
		"BagToSet", "SetToBag",
	},
	"TLC": {
		"Print", "PrintT", "Assert", "JavaTime",
		"TLCGet", "TLCSet", "MakeFcn", "CombineFcn", "Permutations",
		"SortSeq", "RandomElement", "Any", "ToString", "TLCEval",
	},
	"Randomization": {
		"RandomSubset", "RandomSetOfSubsets",
	},
	"Json": {
		"ToJson", "ToJsonArray", "ToJsonObject", "JsonSerialize", "JsonDeserialize",
		"ndJsonSerialize", "ndJsonDeserialize",
	},
	"_JsonTrace": {
		"_JsonTrace",
	},
	"_TLCTrace": {
		"_TLCTrace",
	},
	"IOUtils": {
		"IOSerialize", "IODeserialize", "Serialize", "Deserialize", "IOExec",
		"IOEnvExec", "IOExecTemplate", "IOEnvExecTemplate", "IOEnv", "atoi",
	},
	"Combinatorics": {
		"factorial", "choose",
	},
	"Bitwise": {
		"&", "|", "^^", "Not", "shiftR",
	},
	"Functions": {
		"IsInjective", "AntiFunction", "FoldFunction", "FoldFunctionOnSet",
	},
	"FiniteSetsExt": {
		"Quantify", "kSubset", "FoldSet",
	},
	"BagsExt": {
		"FoldBag",
	},
	"CSV": {
		"CSVWriteRecord", "CSVWrite", "CSVRead", "CSVRecords",
	},
	"GraphViz": {
		"DotDiGraph",
	},
	"Graphs": {
		"SimplePath", "AreConnectedIn", "IsStronglyConnected",
	},
	"UndirectedGraphs": {
		"SimplePath", "AreConnectedIn", "ConnectedComponents",
	},
	"VectorClocks": {
		"CausalOrder",
	},
	"SVG": {
		"SVGElemToString", "NodeOfRingNetwork", "NodesOfDirectedMultiGraph", "PointOnLine",
	},
	"SequencesExt": {
		"SetToSeq", "SetToSeqs", "Contains", "LongestCommonPrefix", "Cons",
		"FoldSeq", "FoldLeft", "FoldRight", "FoldLeftDomain",
		"FoldRightDomain", "ReplaceFirstSubSeq", "ReplaceAllSubSeqs",
		"IsPrefix", "SelectInSeq", "SelectInSubSeq", "SelectLastInSeq",
		"SelectLastInSubSeq", "RemoveFirst", "RemoveFirstMatch",
		"Suffixes", "AllSubSeqs",
	},
	"Statistics": {
		"ChiSquare",
	},
	"TLCExt": {
		"AssertError", "PickSuccessor", "ToTrace", "CounterExample", "Trace",
		"TLCDefer", "TLCNoOp", "TLCModelValue", "TLCCache", "TLCFP",
		"TLCEvalDefinition", "TLCGetOrDefault", "TLCGetAndSet",
	},
	"_Possible": {
		"_Track", "_Counts", "_CheckName", "_PrintCounts",
	},
	"TransitiveClosure": {
		"Warshall",
	},
}

var bridgeNativeOverrideModuleMembers = map[string]map[string]bool{
	"Naturals": setOf(
		"Nat", "+", "-", "*", "^", "<", ">", "\\leq", "\\geq", "%", "\\div", "..",
	),
	"Integers": setOf(
		"Int", "-.",
	),
	"Sequences": setOf(
		"Seq", "Len", "Head", "Tail", "Append", "Concat", "\\o", "SubSeq", "SelectSeq",
	),
	"FiniteSets": setOf(
		"IsFiniteSet", "Cardinality",
	),
	"Bags": setOf(
		"EmptyBag", "IsABag", "BagCardinality", "BagIn", "CopiesIn", "BagCup", "\\oplus",
		"BagDiff", "\\ominus", "BagUnion", "SqSubseteq", "\\sqsubseteq", "BagOfAll",
		"BagToSet", "SetToBag",
	),
	"TLC": setOf(
		"Print", "PrintT", "Assert", "JavaTime",
		"TLCGet", "TLCSet", "MakeFcn", "CombineFcn", ":>", "@@", "Permutations",
		"SortSeq", "RandomElement", "Any", "ToString", "TLCEval",
	),
	"Randomization": setOf(
		"RandomSubset", "RandomSetOfSubsets",
	),
	"Json": setOf(
		"ToJson", "ToJsonArray", "ToJsonObject", "JsonSerialize", "JsonDeserialize",
		"ndJsonSerialize", "ndJsonDeserialize",
	),
	"_JsonTrace": setOf(
		"_TLCState",
	),
	"_TLCTrace": setOf(
		"_TLCTraceDeserialize", "_TLCTraceSerialize", "_TLCState",
	),
	"IOUtils": setOf(
		"IOSerialize", "IODeserialize", "Serialize", "Deserialize", "IOExec",
		"IOEnvExec", "IOExecTemplate", "IOEnvExecTemplate", "IOEnv", "atoi",
	),
	"Combinatorics": setOf(
		"factorial", "choose",
	),
	"Bitwise": setOf(
		"And", "Or", "Xor", "Not", "shiftR",
	),
	"DyadicRationals": setOf(
		"Reduce",
	),
	"Functions": setOf(
		"IsInjective", "AntiFunction", "FoldFunction", "FoldFunctionOnSet",
	),
	"FiniteSetsExt": setOf(
		"Quantify", "kSubset", "FoldSet",
	),
	"BagsExt": setOf(
		"FoldBag",
	),
	"CSV": setOf(
		"CSVWriteRecord", "CSVWrite", "CSVRead", "CSVRecords",
	),
	"GraphViz": setOf(
		"DotDiGraph",
	),
	"Graphs": setOf(
		"SimplePath", "AreConnectedIn", "IsStronglyConnected",
	),
	"UndirectedGraphs": setOf(
		"SimplePath", "AreConnectedIn", "ConnectedComponents",
	),
	"VectorClocks": setOf(
		"CausalOrder",
	),
	"SVG": setOf(
		"SVGElemToString", "NodeOfRingNetwork", "NodesOfDirectedMultiGraph", "PointOnLine",
	),
	"SequencesExt": setOf(
		"SetToSeq", "SetToSeqs", "Contains", "LongestCommonPrefix", "Cons",
		"FoldSeq", "FoldLeft", "FoldRight", "FoldLeftDomain",
		"FoldRightDomain", "ReplaceFirstSubSeq", "ReplaceAllSubSeqs",
		"IsPrefix", "SelectInSeq", "SelectInSubSeq", "SelectLastInSeq",
		"SelectLastInSubSeq", "RemoveFirst", "RemoveFirstMatch",
		"Suffixes", "AllSubSeqs",
	),
	"Statistics": setOf(
		"ChiSquare",
	),
	"TLCExt": setOf(
		"AssertError", "PickSuccessor", "ToTrace", "CounterExample", "Trace",
		"TLCDefer", "TLCNoOp", "TLCModelValue", "TLCCache", "TLCFP",
		"TLCEvalDefinition", "TLCGetOrDefault",
	),
	"_Possible": setOf(
		"_Counts",
	),
}

func setOf(values ...string) map[string]bool {
	out := make(map[string]bool, len(values))
	for _, value := range values {
		out[value] = true
	}
	return out
}

func moduleDefinitionNameIndex(spec *Spec) map[string]map[string]bool {
	out := map[string]map[string]bool{}
	if spec == nil {
		return out
	}
	moduleNames := make([]string, 0, len(spec.Modules))
	for name := range spec.Modules {
		moduleNames = append(moduleNames, name)
	}
	sort.Strings(moduleNames)
	for _, moduleName := range moduleNames {
		mod := spec.Modules[moduleName]
		if mod == nil {
			continue
		}
		names := map[string]bool{}
		for i := range mod.Definitions {
			if mod.Definitions[i].Name != "" {
				names[mod.Definitions[i].Name] = true
			}
		}
		out[moduleName] = names
	}
	return out
}

// BuildTLCTool converts the production Go SANY semantic tree into the TLC
// runtime tree used by the mechanical TLC port.
func BuildTLCTool(spec *Spec, cfg *tlc.ModelConfig, runtime tlc.RuntimeParameters) (*tlc.Tool, Diagnostics) {
	if spec == nil || spec.Root == nil {
		return nil, Diagnostics{errorAt(Position{}, "E7000", "missing root module for TLC tool")}
	}
	if cfg == nil {
		cfg = tlc.NewModelConfig(spec.Root.Name)
	}
	internTLCSourceSyntaxNames(spec)
	tlc.ResetUniqueStringLocations()
	defns := tlc.NewDefns()
	bridge := &tlcBridge{
		tool:                  tlc.NewToolWithModelConfig(cfg),
		processor:             tlc.NewSpecProcessor(spec.Root.Name, defns, cfg),
		spec:                  spec,
		cfg:                   cfg,
		runtime:               runtime,
		defs:                  tlcBridgeDefinitionsByName(spec),
		defns:                 defns,
		symbols:               map[string]*tlc.SymbolNode{},
		rootModuleName:        spec.Root.Name,
		moduleDefinitionNames: moduleDefinitionNameIndex(spec),
		convertBoundNames:     map[string]int{},
	}
	bridge.tool.RootName = spec.Root.Name
	bridge.tool.RootFile = spec.Root.SourcePath
	bridge.tool.DistributedFiles = tlc.NewDistributedServerFiles(filepath.Dir(spec.Root.SourcePath), spec.LibraryPaths, embeddedJavaStandardModules, "test_vectors/java-sany/StandardModules")
	bridge.tool.SpecDir = ""
	bridge.tool.ConfigFile = spec.Root.Name
	bridge.tool.ModuleFiles = append([]string(nil), spec.ModuleFiles...)
	bridge.tool.ParseDebuggerExpressionFunc = bridge.parseDebuggerExpression
	bridge.installVariables()
	bridge.installConstantDeclarations()
	bridge.installDefinitions()
	bridge.installModuleTable()
	bridge.installAssumptions()
	for _, module := range bridge.processor.ModuleTbl.GetModuleNodes() {
		bridge.processor.ProcessConstantsDynamicExtendee(module)
	}
	bridge.installRootDefinitions()
	bridge.installModuleNativeOverrides()
	bridge.installRuntimeConstants()
	bridge.installConfigConstants()
	bridge.installInstanceAliases()
	bridge.installModelTargets()
	bridge.installRuntimeParameters()
	bridge.tool.AssignActionIDs()
	return bridge.tool, bridge.diags
}

// Java parses the root syntax before its dependencies and semantic conversion.
// Intern token images in that source order before converting definitions: record
// normalization compares UniqueString tokens, not alphabetical field names.
func internTLCSourceSyntaxNames(spec *Spec) {
	modules := []*Module{spec.Root}
	seen := map[*Module]bool{}
	for _, filename := range spec.ModuleFiles {
		name := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
		if mod := spec.Modules[name]; mod != nil {
			modules = append(modules, mod)
		}
	}
	for _, mod := range modules {
		if mod == nil || seen[mod] {
			continue
		}
		seen[mod] = true
		tokens, _ := NewSanyTokenManager(mod.SourcePath, mod.Source).LexAll()
		for _, token := range tokens {
			if token.Kind != SanyTokenEOF {
				tlc.UniqueStringOf(token.Image)
			}
		}
	}
}

func (b *tlcBridge) installVariables() {
	nodes := b.variableDeclarations()
	vars := make([]string, len(nodes))
	locations := make(map[string]tlc.SourceLocation, len(nodes))
	for i, node := range nodes {
		vars[i] = node.Name.String()
		locations[vars[i]] = node.Location
	}
	tlc.SetStateVariablesWithLocations(vars, locations)
	if b.processor != nil {
		b.processor.SetVariableNodes(nodes)
	}
}

func (b *tlcBridge) installConstantDeclarations() {
	seen := map[*tlc.SymbolNode]bool{}
	visited := map[*Module]bool{}
	var visit func(*Module)
	visit = func(mod *Module) {
		if mod == nil || visited[mod] {
			return
		}
		visited[mod] = true
		for _, ext := range mod.Extends {
			visit(b.spec.Modules[ext])
		}
		for _, declaration := range mod.Declarations {
			if declaration.Kind != ConstantDecl {
				continue
			}
			for _, name := range declaration.Names {
				symbol := b.declarationSymbol(mod, name)
				symbol.Arity = declaration.Arities[name]
				if !seen[symbol] {
					seen[symbol] = true
					b.processor.ConstantDeclarations = append(b.processor.ConstantDeclarations, symbol)
				}
			}
		}
		for _, inner := range mod.Nested {
			visit(inner)
		}
	}
	visit(b.spec.Root)
}

func (b *tlcBridge) installDefinitions() {
	if mod := b.spec.Modules["Integers"]; mod != nil && moduleNameForSourcePosition(mod.Pos) == mod.Name {
		b.tool.InstallIntegerDefinitions()
	}
	// Alias/source registration must not change which native implementation
	// processModuleOverrides associates with a module's original definition.
	b.nativeDefinitions = make(map[*tlc.UniqueString]any, len(b.tool.DefnsByName))
	for name, value := range b.tool.DefnsByName {
		b.nativeDefinitions[name] = value
	}
	b.retainIntegerNativeOverride()
	b.prepareInstanceDefinitions()
	names := make([]string, 0, len(b.defs))
	for name := range b.defs {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		def := b.defs[name]
		if def == nil {
			continue
		}
		if isNativeStandardDefinitionOverrideName(name, def) {
			b.installNativeStandardDefinitionOverrideAlias(name, def)
			continue
		}
		opDef := b.convertDefinitionAs(name, def)
		if opDef == nil {
			continue
		}
		b.defineAlias(name, opDef)
	}
}

// Keep the instance's substitutions in the runtime graph. The older central
// evaluator needs substituted AST copies; TLC instead binds the source symbols
// through SubstInNode, as SANY does, and shares the substitutions among clones.
func (b *tlcBridge) prepareInstanceDefinitions() {
	if b.spec == nil {
		return
	}
	if b.definitionModules == nil {
		b.definitionModules = map[*Definition]string{}
		b.sourceSymbols = map[*Definition]*tlc.SymbolNode{}
		b.sourceDefinitions = map[*Definition]*tlc.OpDefNode{}
		b.instanceDefinitions = map[string]*tlcBridgeInstance{}
	}
	if b.indexedModules == nil {
		b.indexedModules = map[*Module]bool{}
	}
	names := make([]string, 0, len(b.spec.Modules))
	for name := range b.spec.Modules {
		names = append(names, name)
	}
	sort.Strings(names)
	indexed := false
	for _, name := range names {
		mod := b.spec.Modules[name]
		if mod == nil || b.indexedModules[mod] {
			continue
		}
		b.indexedModules[mod] = true
		indexed = true
		for i := range mod.Definitions {
			b.definitionModules[&mod.Definitions[i]] = name
		}
		for _, inst := range mod.Instances {
			instancee := b.spec.Modules[inst.Module]
			if instancee == nil {
				continue
			}
			binding := &tlcBridgeInstance{owner: mod, inst: inst}
			for i := range instancee.Definitions {
				def := &instancee.Definitions[i]
				if def.Local {
					continue
				}
				key := inst.qualifier() + "!" + def.Name
				keys := []string{mod.Name + "!" + key}
				if mod == b.spec.Root || slices.Contains(b.spec.Root.Extends, mod.Name) && !inst.Local {
					keys = append(keys, key)
					if inst.exportsUnqualified() {
						if existing := b.defs[def.Name]; existing == nil || existing.SourcePosition() == def.SourcePosition() {
							keys = append(keys, def.Name)
						}
					}
				}
				exportName := key
				if inst.exportsUnqualified() {
					exportName = def.Name
				}
				exportSymbol := b.symbol(exportName)
				for _, key := range keys {
					b.defs[key] = def
					b.instanceDefinitions[key] = binding
					b.symbols[key] = exportSymbol
				}
			}
		}
	}
	if !indexed {
		return
	}
	// EXTENDS references the original definition, including native overrides.
	// Keep a single symbol for those aliases so a module-scoped replacement
	// reaches every expression bound to that definition.
	aliases := make([]string, 0, len(b.defs))
	for alias := range b.defs {
		aliases = append(aliases, alias)
	}
	sort.Strings(aliases)
	for _, alias := range aliases {
		def := b.defs[alias]
		module := b.definitionModules[def]
		if module == "" || b.instanceDefinitions[alias] != nil {
			continue
		}
		canonical := module + "!" + def.Name
		if module == b.rootModuleName {
			canonical = def.Name
		}
		b.symbols[alias] = b.sourceDefinitionSymbol(canonical, def)
	}
}

func isNativeStandardDefinitionOverrideName(name string, def *Definition) bool {
	if def == nil {
		return false
	}
	member := name
	if i := strings.LastIndex(member, "!"); i >= 0 {
		member = member[i+1:]
	}
	module := moduleNameForSourcePosition(def.SourcePosition())
	if module == "" {
		return false
	}
	return bridgeNativeOverrideModuleMembers[module][member]
}

func (b *tlcBridge) installNativeStandardDefinitionOverrideAlias(name string, def *Definition) {
	if b == nil || b.tool == nil || def == nil || name == "" {
		return
	}
	module := moduleNameForSourcePosition(def.SourcePosition())
	if module == "" {
		return
	}
	member := name
	if i := strings.LastIndex(member, "!"); i >= 0 {
		member = member[i+1:]
	}
	value := b.nativeDefinitions[tlc.UniqueStringOf(module+"!"+member)]
	if value == nil {
		value = b.nativeDefinitions[tlc.UniqueStringOf(member)]
	}
	if value == nil {
		return
	}
	// Java stores native overrides on the source OpDef body. INSTANCE clones
	// share that body, possibly underneath SubstIn wrappers.
	opDef := b.convertSourceDefinitionAs(module+"!"+member, def)
	if opDef != nil {
		value = tlc.WithEvaluatingOpDef(value, opDef)
		opDef.Body.(interface{ SetToolObject(any) }).SetToolObject(value)
	}
	b.rememberNativeStandardDefinition(module, member, name, opDef)
	b.defineAlias(name, value)
}

func (b *tlcBridge) rememberNativeStandardDefinition(module string, member string, name string, opDef *tlc.OpDefNode) {
	if b == nil || b.tool == nil || opDef == nil {
		return
	}
	if module != "TLCExt" {
		return
	}
	switch member {
	case "CounterExample":
		if name == "CounterExample" || b.tool.CounterExampleDef == nil {
			b.tool.CounterExampleDef = opDef
		}
	case "Trace":
		if name == "TLCExt!Trace" || b.tool.TraceDef == nil {
			b.tool.TraceDef = opDef
		}
	}
}

func moduleNameForSourcePosition(pos Position) string {
	if pos.File == "" {
		return ""
	}
	name := filepath.Base(pos.File)
	return strings.TrimSuffix(name, filepath.Ext(name))
}

func (b *tlcBridge) installConfigConstants() {
	if b.cfg == nil {
		return
	}
	b.installConfigConstantsUnder("", b.cfg.GetConstants())
	for moduleName, constants := range b.cfg.GetModConstants().All() {
		b.installConfigConstantsUnder(moduleName+"!", constants)
	}
}

func (b *tlcBridge) installConfigConstantsUnder(prefix string, constants *tlc.ConfigConstants) {
	if constants == nil {
		return
	}
	opConstants := map[string]*tlc.OpRcdValue{}
	for _, constant := range constants.All() {
		if constant.Value == nil {
			continue
		}
		name := prefix + constant.Name
		if len(constant.Args) != 0 {
			opVal := opConstants[name]
			if opVal == nil {
				opVal = tlc.NewOpRcdValue()
				opConstants[name] = opVal
				b.defineName(name, opVal)
			} else if len(opVal.Domain) != 0 && len(opVal.Domain[0]) != len(constant.Args) {
				b.diags = append(b.diags, errorAt(Position{}, "E7001", "operator-valued CONSTANT assignment %s has inconsistent arity", name))
				continue
			}
			values := make([]tlc.Value, 0, len(constant.Args)+2)
			values = append(values, nil)
			values = append(values, constant.Args...)
			values = append(values, constant.Value)
			opVal.AddLine(values)
			continue
		}
		b.defineName(name, constant.Value)
	}
}

func (b *tlcBridge) installInstanceAliases() {
	if b == nil || b.spec == nil || b.spec.Root == nil {
		return
	}
	for _, inst := range b.spec.Root.Instances {
		for _, binding := range b.standardInstanceBindings(inst) {
			if binding.Symbol != nil {
				if binding.Symbol.Definition != nil {
					// The semantic Defns entry remains the INSTANCE OpDef;
					// native lookup follows its shared source body override.
					b.tool.Define(binding.Symbol, binding.Value)
				} else {
					b.define(binding.Symbol, binding.Value)
				}
			}
		}
	}
}

func (b *tlcBridge) define(sym *tlc.SymbolNode, value any) {
	if b == nil || sym == nil {
		return
	}
	if b.tool != nil {
		b.tool.Define(sym, value)
	}
	if b.defns != nil {
		b.defns.Put(sym, value)
	}
}

func (b *tlcBridge) defineName(name string, value any) *tlc.SymbolNode {
	if b == nil || name == "" {
		return nil
	}
	sym := b.symbol(name)
	b.defineAlias(name, value)
	return sym
}

func (b *tlcBridge) defineAlias(name string, value any) {
	b.define(b.symbol(name), value)
	if b.tool != nil {
		b.tool.DefnsByName[tlc.UniqueStringOf(name)] = value
	}
	if b.defns != nil {
		b.defns.Put(name, value)
	}
}

func (b *tlcBridge) instanceOpDefinitions(inst Instance) []*tlc.OpDefNode {
	if b == nil || b.spec == nil || inst.Module == "" || inst.qualifier() == "" {
		return nil
	}
	mod := b.spec.Modules[inst.Module]
	if mod == nil {
		return nil
	}
	out := make([]*tlc.OpDefNode, 0, len(mod.Definitions))
	binding := &tlcBridgeInstance{owner: b.spec.Modules[b.convertingModule], inst: inst}
	if binding.owner == nil {
		binding.owner = b.spec.Root
	}
	for i := range mod.Definitions {
		def := &mod.Definitions[i]
		if def.Local {
			continue
		}
		opDef := b.convertInstanceDefinition(inst.qualifier()+"!"+def.Name, def, binding)
		if opDef != nil {
			out = append(out, opDef)
		}
	}
	return out
}

func (b *tlcBridge) standardInstanceBindings(inst Instance) []tlc.LetBinding {
	if b == nil || b.tool == nil || inst.Module == "" || inst.qualifier() == "" {
		return nil
	}
	members := bridgeStandardModuleMembers[inst.Module]
	if len(members) == 0 {
		return nil
	}
	out := make([]tlc.LetBinding, 0, len(members))
	for _, member := range members {
		value := b.tool.DefnsByName[tlc.UniqueStringOf(inst.Module+"!"+member)]
		if value == nil {
			value = b.tool.DefnsByName[tlc.UniqueStringOf(member)]
		}
		if value == nil {
			continue
		}
		if _, defined := value.(*tlc.OpDefNode); defined {
			// Keep the instance's cloned body and SubstIn bindings. Binding
			// the original TLA+ definition here would bypass WITH substitutions.
			continue
		}
		out = append(out, tlc.LetBinding{
			Symbol: b.symbol(inst.qualifier() + "!" + member),
			Value:  value,
		})
	}
	return out
}

func (b *tlcBridge) installAssumptions() {
	b.installModuleAssumptions(true)
}

func (b *tlcBridge) installModuleAssumptions(checkingRoot bool) {
	if b.spec == nil || b.spec.Root == nil {
		return
	}
	// ModuleNode.copyAssumes copies each direct extendee's complete vector
	// before this module's own assumptions. INSTANCE does not copy assumptions;
	// repeated EXTENDS paths retain the same source expression identity.
	converted := map[*NamedExpr]*tlc.AssumeNode{}
	if b.assumptionModules == nil {
		b.assumptionModules = map[*Module]bool{}
	}
	topLevels := b.assumptionModules
	var visit func(*Module, bool)
	visit = func(module *Module, checking bool) {
		if module == nil || !checking && topLevels[module] {
			return
		}
		for _, name := range module.Extends {
			visit(b.spec.Modules[name], checking)
			if !topLevels[module] {
				b.moduleNodes[module].TopLevel = append(b.moduleNodes[module].TopLevel, b.moduleNodes[b.spec.Modules[name]].TopLevel...)
			}
		}
		for i := range module.Assumptions {
			assumption := &module.Assumptions[i]
			if assumption.Expr == nil {
				continue
			}
			node := converted[assumption]
			if node == nil {
				var definition *tlc.ThmOrAssumpDefNode
				if assumption.Name != "" {
					definition, _ = b.moduleNodes[module].Context.GetSymbol(tlc.SemanticContextKey{Name: tlc.UniqueStringOf(assumption.Name)}).(*tlc.ThmOrAssumpDefNode)
				}
				previous := b.convertingModule
				b.convertingModule = module.Name
				var expr tlc.SemanticNode
				if definition != nil {
					expr = definition.Body
				} else {
					expr = b.convertExpr(assumption.Expr)
				}
				b.convertingModule = previous
				if expr == nil {
					continue
				}
				node = tlc.NewAssumeNode(expr, b.moduleNodes[module], definition)
				b.withSyntaxNode(assumption.Syntax, node)
				b.withPositionLocation(assumption.SourcePosition(), node)
				node.IsAxiom = assumption.Syntax != nil && len(assumption.Syntax.Heirs) > 0 && assumption.Syntax.Heirs[0].Image == "AXIOM"
				converted[assumption] = node
			}
			if !topLevels[module] {
				b.moduleNodes[module].TopLevel = append(b.moduleNodes[module].TopLevel, node)
			}
			if checking {
				b.processor.Assumptions = append(b.processor.Assumptions, node.Assume)
				b.processor.AssumptionIsAxiom = append(b.processor.AssumptionIsAxiom, node.IsAxiom)
			}
		}
		topLevels[module] = true
	}
	visit(b.spec.Root, checkingRoot)
	// INSTANCE modules keep their own source assumptions for location lookup;
	// TLC checks only the root module's EXTENDS closure above.
	names := make([]string, 0, len(b.spec.Modules))
	for name := range b.spec.Modules {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		module := b.spec.Modules[name]
		if !topLevels[module] {
			visit(module, false)
		}
	}
}

func (b *tlcBridge) installModelTargets() {
	if b == nil || b.cfg == nil {
		return
	}
	if b.processor != nil {
		b.processor.ApplyToTool(b.tool)
	}
}

func (b *tlcBridge) installRuntimeParameters() {
	if b == nil {
		return
	}
	for _, inv := range b.runtime.Invariants {
		expr, diags := parseRuntimeTLAExpression(inv.Expression, inv.Modules)
		b.diags = append(b.diags, diags...)
		if diags.HasErrors() || expr == nil {
			continue
		}
		action := b.actionFromExpr(inv.Expression, expr, nil, false)
		if action != nil {
			b.tool.Invariants = append(b.tool.Invariants, action)
			b.tool.InvariantNames = append(b.tool.InvariantNames, inv.Expression)
		}
	}
	for _, constraint := range b.runtime.Constraints {
		if node := b.nodeForModuleDefinition(constraint.Module, constraint.Operator, "runtime constraint"); node != nil {
			b.tool.ModelConstraints = append(b.tool.ModelConstraints, node)
		}
	}
	for _, constraint := range b.runtime.ActionConstraints {
		if node := b.nodeForModuleDefinition(constraint.Module, constraint.Operator, "runtime action constraint"); node != nil {
			b.tool.ActionConstraints = append(b.tool.ActionConstraints, node)
		}
	}
	if b.cfg == nil || b.cfg.GetView() == "" {
		if view := b.runtime.View; view != nil {
			b.tool.ViewSpec = b.nodeForModuleDefinition(view.Module, view.Operator, "runtime view")
		}
	}
	runtimePosts := make([]*tlc.Action, 0, len(b.runtime.PostConditions))
	for _, post := range b.runtime.PostConditions {
		if action := b.actionFromModuleDefinition(post.Module, post.Operator, false); action != nil {
			runtimePosts = append(runtimePosts, action)
		}
	}
	if len(runtimePosts) != 0 {
		b.tool.PostConditionSpecs = append(runtimePosts, b.tool.PostConditionSpecs...)
	}
}

func (b *tlcBridge) installRuntimeConstants() {
	for _, constant := range b.runtime.StringConstants() {
		b.defineRuntimeStringConstant(constant.Name, constant.Value)
	}
}

func (b *tlcBridge) defineRuntimeStringConstant(name string, value string) {
	if name == "" {
		return
	}
	constant := tlc.NewStringValue(value)
	b.defineName(name, constant)
	// processSpec installs runtime strings on their actual module declarations
	// before processConstantDefns reads declaration tool objects.
	for _, module := range b.processor.ModuleTbl.GetModuleNodes() {
		for _, declaration := range module.GetConstantDecls() {
			if declaration.GetName() == tlc.UniqueStringOf(name) {
				declaration.Data = constant
				b.define(declaration, constant)
			}
		}
	}
}

func (b *tlcBridge) nodeForModuleDefinition(module string, operator string, slot string) tlc.SemanticNode {
	name := moduleQualifiedName(module, operator)
	if name == "" {
		b.diags = append(b.diags, errorAt(Position{}, "E7018", "%s requires a module and operator", slot))
		return nil
	}
	return b.nodeForDefinition(name)
}

func (b *tlcBridge) actionFromModuleDefinition(module string, operator string, init bool) *tlc.Action {
	name := moduleQualifiedName(module, operator)
	if name == "" {
		b.diags = append(b.diags, errorAt(Position{}, "E7019", "runtime postcondition requires a module and operator"))
		return nil
	}
	action := b.actionFromDefinition(name, init)
	if action != nil {
		action.Name = operator
	}
	return action
}

func moduleQualifiedName(module string, operator string) string {
	if module == "" || operator == "" {
		return ""
	}
	return module + "!" + operator
}

func parseRuntimeTLAExpression(expr string, modules []string) (Expr, Diagnostics) {
	if strings.TrimSpace(expr) == "" {
		return nil, Diagnostics{errorAt(Position{}, "E7020", "runtime invariant expression is empty")}
	}
	var source strings.Builder
	source.WriteString("---- MODULE __TLCRuntimeExpression ----\n")
	if len(modules) > 0 {
		source.WriteString("EXTENDS ")
		source.WriteString(strings.Join(modules, ", "))
		source.WriteByte('\n')
	}
	source.WriteString("__RuntimeExpression == ")
	source.WriteString(expr)
	source.WriteString("\n====\n")
	mod, diags := ParseSanyModuleSource("__TLCRuntimeExpression.tla", source.String())
	if diags.HasErrors() || mod == nil {
		return nil, diags
	}
	for i := range mod.Definitions {
		if mod.Definitions[i].Name == "__RuntimeExpression" {
			return mod.Definitions[i].Expr, diags
		}
	}
	diags = append(diags, errorAt(Position{}, "E7021", "runtime invariant expression did not produce a definition"))
	return nil, diags
}

func (b *tlcBridge) parseDebuggerExpression(tool *tlc.Tool, root *tlc.ModuleNode, location tlc.SourceLocation, expression string) (*tlc.OpDefNode, error) {
	if tlc.JavaStringIsBlank(expression) {
		return nil, nil
	}
	if root == nil {
		panic(tlc.NewNullPointerException())
	}
	if op := root.GetOpDef(tlc.UniqueStringOf(expression)); op != nil {
		return op, nil
	}
	moduleName := b.unusedDebuggerModuleName()
	opName := b.unusedDebuggerOpName()
	path := root.PathTo(location, false)
	lets := map[string]*tlc.OpDefNode{}
	for _, node := range path {
		if node, ok := node.(*tlc.LetInNode); ok {
			for _, def := range node.Lets {
				lets[def.Name.String()] = def
			}
		}
	}
	parameters := tlc.GetScopedSymbols(root, path)
	signatureSet := tlc.NewInsMap[string, struct{}]()
	for symbol := range parameters.All() {
		name := symbol.GetName().String()
		if symbol.Definition != nil {
			name = symbol.Definition.Name.String()
		}
		if lets[name] == nil {
			signatureSet.Set(debuggerOperatorSignature(name, symbol.Arity), struct{}{})
		}
	}
	var signatures []string
	for signature := range signatureSet.All() {
		signatures = append(signatures, signature)
	}
	sort.Strings(signatures)
	letNames := make([]string, 0, len(lets))
	for name := range lets {
		letNames = append(letNames, name)
	}
	sort.Strings(letNames)
	var source strings.Builder
	fmt.Fprintf(&source, "---- MODULE %s ----\nEXTENDS %s\n", moduleName, root.Name)
	for _, name := range letNames {
		def := lets[name]
		params := make([]string, len(def.Params))
		for i, param := range def.Params {
			params[i] = debuggerOperatorSignature(param.GetName().String(), param.Arity)
		}
		signature := name
		if len(params) > 0 {
			signature += "(" + strings.Join(params, ", ") + ")"
		}
		fmt.Fprintf(&source, "LOCAL %s == TRUE\n", signature)
	}
	source.WriteString("\n" + opName)
	if len(signatures) > 0 {
		source.WriteString("(" + strings.Join(signatures, ", ") + ")")
	}
	fmt.Fprintf(&source, " == %s\n====\n", expression)
	file := moduleName + ".tla"
	if b.spec.Root.SourcePath != "" {
		file = filepath.Join(filepath.Dir(b.spec.Root.SourcePath), file)
	}
	mod, dependencies, diagnostics := parseSanyModuleSourceWithDependencies(file, source.String())
	if diagnostics.HasErrors() || mod == nil {
		return nil, fmt.Errorf("Syntax error while parsing breakpoint expression \"%s\"", expression)
	}
	failed, err := b.resolveDebuggerDependencies(dependencies)
	if err != nil {
		return nil, err
	}
	if failed {
		return nil, debuggerSemanticError(location, "Semantic error while parsing breakpoint expression \"%s\"", expression)
	}
	wrapped := b.debuggerSpec(mod)
	if diagnostics := CheckSpec(wrapped); diagnostics.HasErrors() {
		semanticFailure, _ := debuggerDiagnosticFailures(diagnostics)
		phase := "Level-checking"
		if semanticFailure {
			phase = "Semantic"
		}
		return nil, debuggerSemanticError(location, "%s error while parsing breakpoint expression \"%s\"", phase, expression)
	}
	convert := b.debuggerBridge(wrapped)
	convert.prepareInstanceDefinitions()
	convert.extendModuleTable(false)
	convert.installModuleAssumptions(false)
	module := convert.moduleNodes[mod]
	op := module.GetOpDef(tlc.UniqueStringOf(opName))
	if op == nil {
		return nil, debuggerSemanticError(location, "Unable to find debugger expression op %s", opName)
	}
	if convert.diags.HasErrors() {
		return nil, debuggerSemanticError(location, "Semantic error while parsing breakpoint expression \"%s\"", expression)
	}
	level := tool.GetLevelBound(op.Body, tlc.EmptyContext)
	if level > tlc.TLCLevelAction {
		return nil, debuggerSemanticError(location, "Debug expressions must be action-level or below; actual level: Temporal")
	}
	b.processor.ProcessConstantsDynamicExtendee(module)
	b.installModuleNativeOverrides()
	for _, name := range letNames {
		tlc.SemanticSubstituteFor([]tlc.SemanticNode{module}, lets[name], module.GetOpDef(tlc.UniqueStringOf(name)))
	}
	return op, nil
}

func debuggerSemanticError(location tlc.SourceLocation, format string, args ...any) error {
	return fmt.Errorf("%s\n\n%s", location.String(), fmt.Sprintf(format, args...))
}

func debuggerLevelDiagnostic(code string) bool {
	switch code {
	case "E4245", "E4310", "E4311", "E4312", "E4313", "E4314", "E4315", "E4352", "E4353", "E4354", "E4356":
		return true
	}
	return false
}

func debuggerOperatorSignature(name string, arity int) string {
	if arity <= 0 {
		return name
	}
	parts := make([]string, arity)
	for i := range parts {
		parts[i] = "_"
	}
	return name + "(" + strings.Join(parts, ",") + ")"
}

func (b *tlcBridge) unusedDebuggerModuleName() string {
	return b.unusedDebuggerName("__DebuggerModule__%d", func(name string) bool {
		return b != nil && b.spec != nil && b.spec.Modules != nil && b.spec.Modules[name] != nil
	})
}

func (b *tlcBridge) unusedDebuggerOpName() string {
	defs := map[string]bool{}
	if b != nil {
		for name := range b.defs {
			defs[name] = true
		}
	}
	return b.unusedDebuggerName("__DebuggerExpr__%d", func(name string) bool {
		return defs[name]
	})
}

func (b *tlcBridge) unusedDebuggerName(format string, used func(string) bool) string {
	for i := 0; ; i++ {
		name := fmt.Sprintf(format, i)
		if used == nil || !used(name) {
			return name
		}
	}
}

func (b *tlcBridge) actionFromDefinition(name string, init bool) *tlc.Action {
	def := b.defs[name]
	if def == nil {
		b.diags = append(b.diags, errorAt(Position{}, "E7005", "model operator %s not found", name))
		return nil
	}
	if len(def.Params) != 0 {
		b.diags = append(b.diags, errorAt(def.Pos, "E7006", "model operator %s must be zero-arity", name))
		return nil
	}
	opDef := b.convertDefinitionAs(name, def)
	if opDef == nil {
		return nil
	}
	// Reuse the body converted in the definition's module scope. Converting
	// it again here loses bindings to that module's LOCAL operators.
	action := tlc.NewActionFromOpDef(opDef.Body, tlc.EmptyContext, opDef, init, false)
	action.Name = name
	action.CM = tlc.NewCostModel(opDef.Body)
	return action
}

func (b *tlcBridge) actionFromExpr(name string, expr Expr, opDef *tlc.OpDefNode, init bool) *tlc.Action {
	pred := b.convertExpr(expr)
	if pred == nil {
		return nil
	}
	action := tlc.NewActionFromOpDef(pred, tlc.EmptyContext, opDef, init, false)
	action.Name = name
	action.CM = tlc.NewCostModel(pred)
	return action
}

func (b *tlcBridge) nodeForDefinition(name string) tlc.SemanticNode {
	def := b.defs[name]
	if def == nil {
		b.diags = append(b.diags, errorAt(Position{}, "E7007", "operator %s not found", name))
		return nil
	}
	if len(def.Params) != 0 {
		b.diags = append(b.diags, errorAt(def.Pos, "E7008", "operator %s must be zero-arity in this TLC model slot", name))
		return nil
	}
	opDef := b.convertDefinitionAs(name, def)
	if opDef == nil {
		return nil
	}
	if setter, ok := opDef.Body.(interface{ SetToolObject(any) }); ok {
		setter.SetToolObject(opDef)
	}
	return opDef.Body
}

func (b *tlcBridge) convertDefinitionAs(name string, def *Definition) *tlc.OpDefNode {
	if def == nil {
		return nil
	}
	b.prepareInstanceDefinitions()
	if binding := b.instanceDefinitions[name]; binding != nil && b.defs[name] == def {
		return b.convertInstanceDefinition(name, def, binding)
	}
	return b.convertSourceDefinitionAs(name, def)
}

// A source OpDefNode and its instantiated export can have the same name but
// retain distinct identities in Java SANY. Source conversion must not route
// back through the instancer's export table.
func (b *tlcBridge) sourceDefinitionSymbol(name string, def *Definition) *tlc.SymbolNode {
	if symbol := b.sourceSymbols[def]; symbol != nil {
		return symbol
	}
	if module := b.definitionModules[def]; module != "" {
		name = module + "!" + def.Name
		if module == b.rootModuleName {
			name = def.Name
		}
	}
	symbol := b.symbol(name)
	if b.instanceDefinitions[name] != nil {
		symbol = tlc.NewSymbolNode(name)
	}
	symbol.Arity = len(def.Params)
	symbol.DeclarationName = tlc.UniqueStringOf(def.Name)
	b.withSyntaxNode(def.Syntax, symbol)
	b.withPositionLocation(def.SourcePosition(), symbol)
	b.sourceSymbols[def] = symbol
	return symbol
}

func (b *tlcBridge) convertSourceDefinitionAs(name string, def *Definition) *tlc.OpDefNode {
	// SANY aliases share the original OpDefNode and body, rather than
	// reconstructing a definition for each imported name.
	if source := b.sourceDefinitions[def]; source != nil {
		return source
	}
	sym := b.sourceDefinitionSymbol(name, def)
	params := make([]*tlc.SymbolNode, len(def.Params))
	priorParams := make([]*tlc.SymbolNode, len(def.Params))
	for i, param := range def.Params {
		priorParams[i] = b.symbols[param]
		params[i] = b.formalParameter(param, def.ParamArities[param], def.ParamPositions[param], def.Syntax)
		b.symbols[param] = params[i]
	}
	prevModule := b.convertingModule
	if module := b.definitionModules[def]; module != "" {
		b.convertingModule = module
	} else if module := moduleNameForSourcePosition(def.SourcePosition()); module != "" && (prevModule == "" || b.spec.Modules[prevModule] == nil || filepath.Base(b.spec.Modules[prevModule].SourcePath) != filepath.Base(def.SourcePosition().File)) {
		b.convertingModule = module
	}
	restore := b.pushConvertBoundNames(def.Params...)
	defer func() {
		restore()
		for i, name := range def.Params {
			if priorParams[i] == nil {
				delete(b.symbols, name)
			} else {
				b.symbols[name] = priorParams[i]
			}
		}
		b.convertingModule = prevModule
	}()
	var body tlc.SemanticNode
	if function, ok := def.Expr.(*FunctionExpr); ok && def.FunctionDef {
		body = b.functionDefinitionNode(def, function)
	} else {
		body = b.convertExpr(def.Expr)
	}
	if body == nil {
		return nil
	}
	opDef := tlc.NewOpDefNodeForSymbol(sym, params, body)
	// Qualification belongs to the lookup alias. EXTENDS preserves the
	// instancee's original OpDef name for signatures and action labels.
	opDef.Name = tlc.UniqueStringOf(def.Name)
	if mod := b.spec.Modules[b.convertingModule]; mod != nil {
		opDef.SetInRecursive(recursiveDeclarationSections(mod.Recursives)[def.Name] != 0)
	}
	b.withSyntaxNode(def.Syntax, opDef)
	b.withPositionLocation(def.SourcePosition(), opDef)
	if declaration := b.sourceLocationForPosition(def.DeclarationPosition()); !declaration.IsNull() {
		opDef.SetDeclarationLocation(declaration)
	}
	b.sourceDefinitions[def] = opDef
	return opDef
}

func (b *tlcBridge) convertInstanceDefinition(name string, def *Definition, binding *tlcBridgeInstance) *tlc.OpDefNode {
	if clone := binding.defs[def]; clone != nil {
		return clone
	}
	inst := binding.inst
	source := b.convertSourceDefinitionAs(inst.Module+"!"+def.Name, def)
	if source == nil {
		return nil
	}
	// SANY instances reuse their source definition, whose native override may
	// already be installed. Register an otherwise hidden source without replacing
	// that override with the TLA+ placeholder body.
	if b.tool.Lookup(source.Symbol, nil, nil, false) == nil {
		b.define(source.Symbol, source)
	}
	b.prepareInstanceBinding(binding)
	params := append(append([]*tlc.SymbolNode(nil), binding.params...), source.Params...)
	body := source.Body
	if len(binding.substs) > 0 {
		body = b.withPositionLocation(inst.SourcePosition(), tlc.NewSubstInNode(body, binding.substs...))
	}
	clone := tlc.NewOpDefNodeForSymbol(b.symbol(name), params, body)
	clone.SourceDefinition = source.GetSource()
	clone.Local = inst.Local
	if inst.Name != "" && len(binding.substs) > 0 {
		clone.CompoundID = append([]*tlc.UniqueString{tlc.UniqueStringOf(inst.Name)}, source.GetCompoundID()...)
	}
	b.withPositionLocation(inst.SourcePosition(), clone)
	if !positionIsZero(inst.LHSPos) {
		clone.SetDeclarationLocation(b.sourceLocationForPosition(inst.LHSPos))
	}
	if binding.defs == nil {
		binding.defs = map[*Definition]*tlc.OpDefNode{}
	}
	binding.defs[def] = clone
	return clone
}

func (b *tlcBridge) prepareInstanceBinding(binding *tlcBridgeInstance) {
	inst := binding.inst
	if binding.params == nil {
		binding.params = make([]*tlc.SymbolNode, len(inst.Params))
		for i, param := range inst.Params {
			binding.params[i] = tlc.NewSymbolNode(param)
		}
	}
	if binding.substs == nil {
		previous := b.convertingModule
		b.convertingModule = binding.owner.Name
		priorParams := make([]*tlc.SymbolNode, len(inst.Params))
		for i, name := range inst.Params {
			priorParams[i] = b.symbols[name]
			b.symbols[name] = binding.params[i]
		}
		restore := b.pushConvertBoundNames(inst.Params...)
		var substs []tlc.Subst
		explicit := inst.Substitutions
		for _, target := range b.instanceTargets(b.spec.Modules[inst.Module], map[string]bool{}) {
			expr := explicit[target.name]
			if expr == nil {
				expr = &IdentExpr{Name: target.name, Pos: inst.SourcePosition()}
			}
			var replacement tlc.SemanticNode
			if target.arity > 0 {
				if ident, ok := expr.(*IdentExpr); ok {
					replacement = b.withExprLocation(expr, tlc.NewOpArgNode(b.exprSymbol(ident.Name)))
				}
			}
			if replacement == nil {
				replacement = b.convertExpr(expr)
			}
			substs = append(substs, tlc.Subst{Op: target.sym, Expr: replacement})
		}
		restore()
		for i, name := range inst.Params {
			if priorParams[i] == nil {
				delete(b.symbols, name)
			} else {
				b.symbols[name] = priorParams[i]
			}
		}
		b.convertingModule = previous
		// Assign identities once, then reuse the Subst values in every clone.
		binding.substs = tlc.NewSubstInNode(nil, substs...).Substs
	}
}

func (b *tlcBridge) instanceTargets(mod *Module, visiting map[string]bool) []tlcBridgeInstanceTarget {
	if mod == nil || visiting[mod.Name] {
		return nil
	}
	visiting[mod.Name] = true
	defer delete(visiting, mod.Name)
	var out []tlcBridgeInstanceTarget
	for _, ext := range mod.Extends {
		out = append(out, b.instanceTargets(b.spec.Modules[ext], visiting)...)
	}
	for _, decl := range mod.Declarations {
		if decl.Kind != ConstantDecl && decl.Kind != VariableDecl {
			continue
		}
		for _, name := range decl.Names {
			sym := b.declarationSymbol(mod, name)
			if decl.Kind == VariableDecl {
				sym.MarkVariableDecl()
			}
			target := tlcBridgeInstanceTarget{name: name, arity: decl.Arities[name], sym: sym}
			index := slices.IndexFunc(out, func(existing tlcBridgeInstanceTarget) bool { return existing.name == name })
			if index >= 0 {
				out[index] = target
			} else {
				out = append(out, target)
			}
		}
	}
	return out
}

func (b *tlcBridge) declarationSymbol(mod *Module, name string) *tlc.SymbolNode {
	if node := b.moduleNodes[mod]; node != nil {
		if declaration, ok := node.Context.GetSymbol(tlc.SemanticContextKey{Name: tlc.UniqueStringOf(name)}).(*tlc.SymbolNode); ok {
			// A dependency compiled as a temporary root keeps its original
			// declaration identity when later instantiated by debugger expressions.
			return declaration
		}
	}
	// EXTENDS shares declaration identity with the root. An INSTANCE retains
	// the instancee's declaration and substitutes a binding for that symbol.
	var extends func(*Module, map[string]bool) bool
	extends = func(current *Module, visiting map[string]bool) bool {
		if current == nil || visiting[current.Name] {
			return false
		}
		if current == mod {
			return true
		}
		visiting[current.Name] = true
		for _, ext := range current.Extends {
			if extends(b.spec.Modules[ext], visiting) {
				return true
			}
		}
		return false
	}
	if extends(b.spec.Root, map[string]bool{}) {
		symbol := b.symbol(name)
		symbol.DeclarationName = tlc.UniqueStringOf(name)
		return symbol
	}
	symbol := b.symbol(mod.Name + "!" + name)
	symbol.DeclarationName = tlc.UniqueStringOf(name)
	return symbol
}

func (b *tlcBridge) pushConvertBoundNames(names ...string) func() {
	if b.convertBoundNames == nil {
		b.convertBoundNames = map[string]int{}
	}
	for _, name := range names {
		if name == "" {
			continue
		}
		b.convertBoundNames[name]++
	}
	return func() {
		for _, name := range names {
			if name == "" {
				continue
			}
			next := b.convertBoundNames[name] - 1
			if next <= 0 {
				delete(b.convertBoundNames, name)
			} else {
				b.convertBoundNames[name] = next
			}
		}
	}
}

func (b *tlcBridge) convertBound(name string) bool {
	if b == nil || name == "" {
		return false
	}
	return b.convertBoundNames[name] > 0
}

func (b *tlcBridge) resolveExprName(name string) string {
	name = tlcSymbolName(name)
	if b == nil || name == "" || b.convertBound(name) {
		return name
	}
	module := b.convertingModule
	if module == "" || module == b.rootModuleName {
		return name
	}
	if b.instanceDefinitions[module+"!"+name] != nil {
		return module + "!" + name
	}
	if strings.Contains(name, "!") {
		return name
	}
	if b.moduleDefinitionNames[module][name] {
		return module + "!" + name
	}
	if mod := b.spec.Modules[module]; mod != nil {
		for _, target := range b.instanceTargets(mod, map[string]bool{}) {
			if target.name == name {
				return target.sym.Name.String()
			}
		}
	}
	return name
}

func (b *tlcBridge) exprSymbol(name string) *tlc.SymbolNode {
	if !b.convertBound(name) {
		if module := b.spec.Modules[b.convertingModule]; module != nil {
			if node := b.moduleNodes[module]; node != nil {
				switch symbol := node.Context.GetSymbol(tlc.SemanticContextKey{Name: tlc.UniqueStringOf(name)}).(type) {
				case *tlc.OpDefNode:
					return symbol.Symbol
				case *tlc.SymbolNode:
					return symbol
				}
			}
			for i := range module.Definitions {
				def := &module.Definitions[i]
				if def.Name == name {
					return b.sourceDefinitionSymbol(name, def)
				}
			}
		}
	}
	return b.symbol(b.resolveExprName(name))
}

func (b *tlcBridge) convertExpr(expr Expr) tlc.SemanticNode {
	if expr == nil {
		return nil
	}
	var node tlc.SemanticNode
	switch e := expr.(type) {
	case *IdentExpr:
		if e.Name == "@" {
			node = tlc.NewAtNode()
		} else {
			node = tlc.NewOpApplNode(b.exprSymbol(e.Name))
		}
	case *LiteralExpr:
		node = b.convertLiteral(e)
	case *UnaryExpr:
		node = b.unaryNode(e)
	case *BinaryExpr:
		node = b.binaryNode(e)
	case *CallExpr:
		node = b.callNode(e)
	case *IfExpr:
		node = b.builtinNode(tlc.OpITE, b.convertExpr(e.Cond), b.convertExpr(e.Then), b.convertExpr(e.Else))
	case *LetExpr:
		node = b.letNode(e)
	case *QuantifierExpr:
		node = b.quantifierNode(e)
	case *CaseExpr:
		node = b.caseNode(e)
	case *ChooseExpr:
		node = b.chooseNode(e)
	case *TupleExpr:
		args := make([]tlc.SemanticNode, 0, len(e.Elems))
		for _, elem := range e.Elems {
			args = append(args, b.convertExpr(elem))
		}
		node = b.builtinNode(tlc.OpTup, args...)
	case *SetExpr:
		args := make([]tlc.SemanticNode, 0, len(e.Elems))
		for _, elem := range e.Elems {
			args = append(args, b.convertExpr(elem))
		}
		node = b.builtinNode(tlc.OpSE, args...)
	case *RecordExpr:
		args := make([]tlc.SemanticNode, 0, len(e.Fields))
		for _, field := range e.Fields {
			fieldPos := field.Source
			if positionIsZero(fieldPos) {
				fieldPos = field.Pos
			}
			name := b.withPositionLocation(fieldPos, tlc.NewStringNode(field.Name))
			pair := b.builtinNode(tlc.OpPair, name, b.convertExpr(field.Value))
			args = append(args, b.withPositionLocation(fieldPos, pair))
		}
		node = b.builtinNode(tlc.OpRC, args...)
	case *RecordComponentExpr:
		field := b.withPositionLocation(e.FieldPos, tlc.NewStringNode(e.Field))
		node = b.builtinNode(tlc.OpRS, b.convertExpr(e.Record), field)
	case *RecordSetExpr:
		args := make([]tlc.SemanticNode, 0, len(e.Fields))
		for _, field := range e.Fields {
			fieldPos := field.Source
			if positionIsZero(fieldPos) {
				fieldPos = field.Pos
			}
			name := b.withPositionLocation(fieldPos, tlc.NewStringNode(field.Name))
			pair := b.builtinNode(tlc.OpPair, name, b.convertExpr(field.Set))
			args = append(args, b.withPositionLocation(fieldPos, pair))
		}
		node = b.builtinNode(tlc.OpSOR, args...)
	case *FunctionExpr:
		if e.IsLambda {
			node = b.lambdaNode(e)
		} else {
			node = b.functionNode(e)
		}
	case *FunctionAppExpr:
		arg := b.convertFunctionArgs(e.Args)
		node = b.builtinNode(tlc.OpFA, b.convertExpr(e.Function), arg)
	case *ExceptExpr:
		node = b.exceptNode(e)
	case *LabelExpr:
		node = tlc.NewLabelNode(b.convertExpr(e.Body))
	case *ActionExpr:
		op := tlc.OpSA
		if e.Kind == "angle" {
			op = tlc.OpAA
		}
		node = b.builtinNode(op, b.convertExpr(e.Action), b.convertExpr(e.Subscript))
	case *FairnessExpr:
		op := tlc.OpWF
		if e.Kind == "SF" {
			op = tlc.OpSF
		}
		node = b.builtinNode(op, b.convertExpr(e.Subscript), b.convertExpr(e.Action))
	case *FunctionSetExpr:
		node = b.builtinNode(tlc.OpSOF, b.convertExpr(e.Domain), b.convertExpr(e.Range))
	case *SetComprehensionExpr:
		node = b.setComprehensionNode(e)
	default:
		b.diags = append(b.diags, errorAt(expr.Position(), "E7009", "unsupported expression %T in TLC bridge", expr))
		node = tlc.NewValueNode(tlc.ValUndef)
	}
	return b.withExprLocation(expr, node)
}

func (b *tlcBridge) withExprLocation(expr Expr, node tlc.SemanticNode) tlc.SemanticNode {
	if expr == nil {
		return node
	}
	if source, ok := expr.(interface{ GetSyntaxNode() *SanySyntaxNode }); ok {
		b.withSyntaxNode(source.GetSyntaxNode(), node)
	}
	return b.withPositionLocation(expr.Position(), node)
}

func (b *tlcBridge) withSyntaxNode(syntax *SanySyntaxNode, node tlc.SemanticNode) tlc.SemanticNode {
	if syntax != nil {
		if setter, ok := node.(interface{ SetTreeNode(any) }); ok {
			setter.SetTreeNode(syntax)
		}
	}
	return node
}

func (b *tlcBridge) withPositionLocation(pos Position, node tlc.SemanticNode) tlc.SemanticNode {
	if node == nil || positionIsZero(pos) {
		return node
	}
	location := b.sourceLocationForPosition(pos)
	if location.IsNull() {
		return node
	}
	if setter, ok := node.(interface{ SetSourceLocation(tlc.SourceLocation) }); ok {
		setter.SetSourceLocation(location)
	}
	return node
}

func (b *tlcBridge) sourceLocationForPosition(pos Position) tlc.SourceLocation {
	if positionIsZero(pos) {
		return tlc.NullSourceLocation
	}
	source := pos.File
	if source != "" {
		// Parser positions retain physical filenames. Java semantic Location
		// uses the parse module's name in TLC diagnostics and call stacks.
		source = strings.TrimSuffix(filepath.Base(source), filepath.Ext(source))
	}
	if source == "" {
		source = b.convertingModule
	}
	if source == "" {
		source = b.rootModuleName
	}
	endLine, endColumn := pos.EndLine, pos.EndColumn
	if endLine == 0 {
		endLine = pos.Line
	}
	if endColumn == 0 {
		endColumn = pos.Column
	}
	return tlc.NewSourceLocation(source, pos.Line, pos.Column, endLine, endColumn)
}

func positionIsZero(pos Position) bool {
	return pos.Line == 0 && pos.Column == 0 && pos.File == "" && pos.EndLine == 0 && pos.EndColumn == 0
}

func (b *tlcBridge) convertLiteral(e *LiteralExpr) tlc.SemanticNode {
	switch e.Kind {
	case "bool":
		// TRUE and FALSE are predefined SANY operator applications. Keep the
		// application node so evaluation and coverage retain its source location.
		return tlc.NewOpApplNode(b.exprSymbol(strings.ToUpper(e.Value)))
	case "number":
		if strings.Contains(e.Value, ".") {
			return tlc.NewDecimalNode(nil, e.Value)
		}
		node, err := tlc.NewNumeralNodeFromString(e.Value)
		if err != nil {
			b.diags = append(b.diags, errorAt(e.Pos, "E7011", "%s", err))
			return tlc.NewValueNode(tlc.ValUndef)
		}
		return node
	case "string":
		value := e.Value
		if unquoted, err := strconv.Unquote(value); err == nil {
			value = unquoted
		}
		return tlc.NewStringNode(value)
	case "model":
		return tlc.NewValueNode(tlc.MakeModelValue(e.Value))
	default:
		b.diags = append(b.diags, errorAt(e.Pos, "E7012", "unsupported literal kind %q in TLC bridge", e.Kind))
		return tlc.NewValueNode(tlc.ValUndef)
	}
}

func (b *tlcBridge) unaryNode(e *UnaryExpr) tlc.SemanticNode {
	op := e.Op
	switch op {
	case "/\\":
		return b.builtinNode(tlc.OpCL, b.convertExpr(e.Expr))
	case "\\/":
		return b.builtinNode(tlc.OpDL, b.convertExpr(e.Expr))
	case "~", "\\neg":
		op = "\\lnot"
	}
	return tlc.NewOpApplNode(b.exprSymbol(op), b.convertExpr(e.Expr))
}

func (b *tlcBridge) binaryNode(e *BinaryExpr) tlc.SemanticNode {
	if e.Op == "\\X" || e.Op == "\\times" {
		// Java Generator.N_Times emits one $CartesianProd application. The
		// central AST's synthetic n-ary links share one source range;
		// parenthesized operands retain their own ranges and remain nested.
		var args []tlc.SemanticNode
		var collect func(Expr)
		collect = func(expr Expr) {
			if nested, ok := expr.(*BinaryExpr); ok && nested.SanyNary && nested.Pos == e.Pos && (nested.Op == "\\X" || nested.Op == "\\times") {
				collect(nested.Left)
				collect(nested.Right)
				return
			}
			args = append(args, b.convertExpr(expr))
		}
		collect(e.Left)
		collect(e.Right)
		return b.builtinNode(tlc.OpCP, args...)
	}
	if e.JunctionList {
		op := tlc.OpCL
		if e.Op == "\\/" {
			op = tlc.OpDL
		}
		var args []tlc.SemanticNode
		var collect func(Expr)
		collect = func(expr Expr) {
			if nested, ok := expr.(*BinaryExpr); ok && sanySameJunctionFrame(e, nested) {
				collect(nested.Left)
				collect(nested.Right)
				return
			}
			args = append(args, b.convertExpr(expr))
		}
		collect(e.Left)
		collect(e.Right)
		return b.builtinNode(op, args...)
	}
	op := tlcBinaryOperator(e.Op)
	return tlc.NewOpApplNode(b.exprSymbol(op), b.convertExpr(e.Left), b.convertExpr(e.Right))
}

func (b *tlcBridge) callNode(e *CallExpr) tlc.SemanticNode {
	callee, ok := e.Callee.(*IdentExpr)
	if !ok {
		args := make([]tlc.SemanticNode, 1, len(e.Args)+1)
		args[0] = b.convertExpr(e.Callee)
		for _, arg := range e.Args {
			args = append(args, b.convertExpr(arg))
		}
		if len(args) == 2 {
			return b.builtinNode(tlc.OpFA, args[0], args[1])
		}
		return b.builtinNode(tlc.OpFA, args[0], b.builtinNode(tlc.OpTup, args[1:]...))
	}
	args := make([]tlc.SemanticNode, 0, len(e.Args))
	for _, arg := range e.Args {
		if ident, ok := arg.(*IdentExpr); ok {
			if symbol := b.exprSymbol(ident.Name); symbol.Arity > 0 {
				args = append(args, b.withExprLocation(arg, tlc.NewOpArgNode(symbol)))
				continue
			}
		}
		args = append(args, b.convertExpr(arg))
	}
	return tlc.NewOpApplNode(b.exprSymbol(callee.Name), args...)
}

func (b *tlcBridge) letNode(e *LetExpr) tlc.SemanticNode {
	lets := make([]*tlc.OpDefNode, 0, len(e.Definitions))
	letNames := make([]string, 0, len(e.Definitions))
	for _, def := range e.Definitions {
		if def.Name != "" {
			letNames = append(letNames, def.Name)
		}
	}
	for _, inst := range e.Instances {
		if mod := b.spec.Modules[inst.Module]; mod != nil {
			for _, def := range mod.Definitions {
				if !def.Local {
					letNames = append(letNames, inst.qualifier()+"!"+def.Name)
				}
			}
		}
	}
	priorSymbols := make([]*tlc.SymbolNode, len(letNames))
	for i, name := range letNames {
		priorSymbols[i] = b.symbols[name]
		b.symbols[name] = tlc.NewSymbolNode(name)
	}
	defer func() {
		for i, name := range letNames {
			if priorSymbols[i] == nil {
				delete(b.symbols, name)
			} else {
				b.symbols[name] = priorSymbols[i]
			}
		}
	}()
	restore := b.pushConvertBoundNames(letNames...)
	defer restore()
	for _, def := range e.Definitions {
		next := def
		opDef := b.convertDefinitionAs(next.Name, &next)
		if opDef != nil {
			opDef.SetInRecursive(recursiveDeclarationSections(e.Recursives)[next.Name] != 0)
			// A Java OpApplNode references its LET OpDefNode directly. Keep
			// that definition available to graph walking and default lookup;
			// evaluation's contextual lazy binding still takes precedence.
			opDef.Symbol.Data = opDef
		}
		lets = append(lets, opDef)
	}
	for _, inst := range e.Instances {
		for _, opDef := range b.instanceOpDefinitions(inst) {
			// Java getActions unwraps LET without extending its context. The
			// exported OpDef is therefore also available directly at its symbol.
			opDef.Symbol.Data = opDef
			lets = append(lets, opDef)
		}
	}
	node := tlc.NewLetInNode(b.convertExpr(e.Body), lets...)
	for _, inst := range e.Instances {
		node.Bindings = append(node.Bindings, b.standardInstanceBindings(inst)...)
	}
	return node
}

func (b *tlcBridge) quantifierNode(e *QuantifierExpr) tlc.SemanticNode {
	op := tlc.OpUF
	if e.Set != nil {
		if e.Kind == "\\E" {
			op = tlc.OpBE
		} else {
			op = tlc.OpBF
		}
	} else if e.Kind == "\\E" {
		op = tlc.OpUE
	}
	// The central AST represents a source variable list as nested wrappers.
	// SANY generates one OpApplNode for the entire quantifier, not one per name.
	quantifiers := []*QuantifierExpr{e}
	body := e.Body
	for e.Syntax != nil {
		next, ok := body.(*QuantifierExpr)
		if !ok || next.Syntax != e.Syntax || next.Kind != e.Kind {
			break
		}
		quantifiers = append(quantifiers, next)
		body = next.Body
	}
	node := b.builtinNode(op)
	parameters := make([]*tlc.SymbolNode, len(quantifiers))
	for i, quantifier := range quantifiers {
		parameters[i] = b.formalParameter(quantifier.Var, quantifier.OperatorArity, quantifier.VarPos, e.Syntax)
	}
	// Generator.processQuantBoundArgs converts every domain before adding any
	// quantified variable to the new context.
	for i := 0; i < len(quantifiers); {
		first := quantifiers[i]
		symbols := []*tlc.SymbolNode{parameters[i]}
		i++
		for i < len(quantifiers) && quantifiers[i].Set == first.Set && quantifiers[i].TupleBound == first.TupleBound {
			symbols = append(symbols, parameters[i])
			i++
		}
		if first.Set != nil {
			node.BdedQuantSymbolLists = append(node.BdedQuantSymbolLists, symbols)
			node.BdedQuantBounds = append(node.BdedQuantBounds, b.convertExpr(first.Set))
			node.BdedQuantATuple = append(node.BdedQuantATuple, first.TupleBound)
		} else {
			node.UnbdedQuantSymbols = append(node.UnbdedQuantSymbols, symbols...)
		}
	}
	restore := b.pushFormalParameters(parameters)
	node.Args = []tlc.SemanticNode{b.convertExpr(body)}
	restore()
	return node
}

func (b *tlcBridge) caseNode(e *CaseExpr) tlc.SemanticNode {
	args := make([]tlc.SemanticNode, 0, len(e.Arms)+1)
	for _, arm := range e.Arms {
		pair := b.builtinNode(tlc.OpPair, b.convertExpr(arm.Test), b.convertExpr(arm.Value))
		args = append(args, b.withPositionLocation(arm.Pos, pair))
	}
	if e.Other != nil {
		pair := b.builtinNode(tlc.OpPair, nil, b.convertExpr(e.Other))
		args = append(args, b.withPositionLocation(e.OtherPos, pair))
	}
	return b.builtinNode(tlc.OpCase, args...)
}

func (b *tlcBridge) chooseNode(e *ChooseExpr) tlc.SemanticNode {
	op := tlc.OpUC
	if e.Set != nil {
		op = tlc.OpBC
	}
	var bound tlc.SemanticNode
	if e.Set != nil {
		bound = b.convertExpr(e.Set)
	}
	parameter := b.formalParameter(e.Var, 0, e.VarPos, e.Syntax)
	restore := b.pushFormalParameters([]*tlc.SymbolNode{parameter})
	body := b.convertExpr(e.Body)
	restore()
	node := b.builtinNode(op, body)
	if e.Set != nil {
		node.BdedQuantSymbolLists = [][]*tlc.SymbolNode{{parameter}}
		node.BdedQuantBounds = []tlc.SemanticNode{bound}
		node.BdedQuantATuple = []bool{false}
	} else {
		node.UnbdedQuantSymbols = []*tlc.SymbolNode{parameter}
	}
	return node
}

func (b *tlcBridge) lambdaNode(e *FunctionExpr) tlc.SemanticNode {
	params := b.boundParameters(e.Bounds, e.Syntax)
	restore := b.pushFormalParameters(params)
	body := b.convertExpr(e.Body)
	restore()
	// Each LAMBDA has its own definition identity; its interned name remains
	// LAMBDA, as in SANY's generateLambda. It is an operator argument.
	symbol := tlc.NewSymbolNode("LAMBDA")
	def := tlc.NewOpDefNodeForSymbol(symbol, params, body)
	b.withPositionLocation(e.Pos, def)
	symbol.Data = def
	return tlc.NewOpArgNode(symbol)
}

func (b *tlcBridge) functionDefinitionNode(def *Definition, e *FunctionExpr) tlc.SemanticNode {
	op := tlc.OpNRFS
	var self *tlc.SymbolNode
	if exprReferencesName(e, def.Name, nil) {
		op = tlc.OpRFS
		self = b.formalParameter(def.Name, 0, def.SourcePosition(), def.Syntax)
		previous := b.symbols[def.Name]
		b.symbols[def.Name] = self
		restore := b.pushConvertBoundNames(def.Name)
		defer func() {
			restore()
			if previous == nil {
				delete(b.symbols, def.Name)
			} else {
				b.symbols[def.Name] = previous
			}
		}()
	}
	node := b.functionNode(e).(*tlc.OpApplNode)
	node.Operator = tlc.NewSymbolNode(op.String())
	if self != nil {
		node.UnbdedQuantSymbols = []*tlc.SymbolNode{self}
	}
	// SANY function specifications cover the whole definition, unlike |->
	// constructor expressions whose location is the bracketed expression.
	return b.withPositionLocation(def.SourcePosition(), b.withSyntaxNode(def.Syntax, node))
}

func (b *tlcBridge) functionNode(e *FunctionExpr) tlc.SemanticNode {
	boundExprs := make([]tlc.SemanticNode, len(e.Bounds))
	for i, bound := range e.Bounds {
		boundExprs[i] = b.convertExpr(bound.Set)
	}
	parameters := b.boundParameters(e.Bounds, e.Syntax)
	restore := b.pushFormalParameters(parameters)
	defer restore()
	body := b.convertExpr(e.Body)
	node := b.builtinNode(tlc.OpFC, body)
	b.appendBoundGroups(node, b.boundGroups(e.Bounds, boundExprs))
	return node
}

func (b *tlcBridge) exceptNode(e *ExceptExpr) tlc.SemanticNode {
	args := []tlc.SemanticNode{b.convertExpr(e.Base)}
	for _, spec := range e.Specs {
		pathElems := make([]tlc.SemanticNode, 0)
		for _, component := range spec.Components {
			if component.Field != "" {
				pathElems = append(pathElems, b.withPositionLocation(component.FieldPos, tlc.NewStringNode(component.Field)))
			}
			for _, index := range component.Indices {
				pathElems = append(pathElems, b.convertExpr(index))
			}
		}
		path := b.withPositionLocation(spec.Pos, b.builtinNode(tlc.OpTup, pathElems...))
		pair := b.builtinNode(tlc.OpPair, path, b.convertExpr(spec.Value))
		args = append(args, b.withPositionLocation(spec.Pos, pair))
	}
	return b.builtinNode(tlc.OpExc, args...)
}

func (b *tlcBridge) setComprehensionNode(e *SetComprehensionExpr) tlc.SemanticNode {
	boundExprs := make([]tlc.SemanticNode, len(e.Bounds))
	for i, bound := range e.Bounds {
		boundExprs[i] = b.convertExpr(bound.Set)
	}
	parameters := b.boundParameters(e.Bounds, e.Syntax)
	restore := b.pushFormalParameters(parameters)
	defer restore()
	groups := b.boundGroups(e.Bounds, boundExprs)
	if e.Predicate != nil {
		predicate := b.convertExpr(e.Predicate)
		filtered := b.builtinNode(tlc.OpSSO, predicate)
		filterSymbols, filterTuple := b.appendFilteredComprehensionBounds(filtered, groups, e.Pos)
		if setComprehensionElementIsBound(e) {
			return filtered
		}
		body := b.convertExpr(e.Element)
		node := b.builtinNode(tlc.OpSOA, body)
		node.BdedQuantSymbolLists = append(node.BdedQuantSymbolLists, filterSymbols)
		node.BdedQuantBounds = append(node.BdedQuantBounds, filtered)
		node.BdedQuantATuple = append(node.BdedQuantATuple, filterTuple)
		return node
	}
	body := b.convertExpr(e.Element)
	node := b.builtinNode(tlc.OpSOA, body)
	b.appendBoundGroups(node, groups)
	return node
}

type tlcBoundGroup struct {
	symbols []*tlc.SymbolNode
	bound   tlc.SemanticNode
	tuple   bool
}

func (b *tlcBridge) boundGroups(bounds []BoundVar, boundExprs []tlc.SemanticNode) []tlcBoundGroup {
	groups := make([]tlcBoundGroup, 0, len(bounds))
	for i := 0; i < len(bounds); i++ {
		boundExpr := tlc.SemanticNode(nil)
		if i < len(boundExprs) {
			boundExpr = boundExprs[i]
		}
		if !bounds[i].TupleBound {
			groups = append(groups, tlcBoundGroup{
				symbols: []*tlc.SymbolNode{b.symbol(bounds[i].Name)},
				bound:   boundExpr,
			})
			continue
		}
		group := tlcBoundGroup{bound: boundExpr, tuple: true}
		for i < len(bounds) && bounds[i].TupleBound {
			group.symbols = append(group.symbols, b.symbol(bounds[i].Name))
			i++
		}
		i--
		groups = append(groups, group)
	}
	return groups
}

func (b *tlcBridge) appendBoundGroups(node *tlc.OpApplNode, groups []tlcBoundGroup) {
	for _, group := range groups {
		node.BdedQuantSymbolLists = append(node.BdedQuantSymbolLists, group.symbols)
		node.BdedQuantBounds = append(node.BdedQuantBounds, group.bound)
		node.BdedQuantATuple = append(node.BdedQuantATuple, group.tuple)
	}
}

func (b *tlcBridge) appendFilteredComprehensionBounds(node *tlc.OpApplNode, groups []tlcBoundGroup, pos Position) ([]*tlc.SymbolNode, bool) {
	if len(groups) == 1 {
		b.appendBoundGroups(node, groups)
		return groups[0].symbols, groups[0].tuple
	}
	allSymbols := make([]*tlc.SymbolNode, 0)
	productArgs := make([]tlc.SemanticNode, 0, len(groups))
	for _, group := range groups {
		if group.tuple {
			b.diags = append(b.diags, errorAt(pos, "E7013", "filtered set comprehension with multiple tuple-bound domains is not yet supported by the TLC bridge"))
		}
		allSymbols = append(allSymbols, group.symbols...)
		productArgs = append(productArgs, group.bound)
	}
	product := b.withPositionLocation(pos, b.builtinNode(tlc.OpCP, productArgs...))
	node.BdedQuantSymbolLists = append(node.BdedQuantSymbolLists, allSymbols)
	node.BdedQuantBounds = append(node.BdedQuantBounds, product)
	node.BdedQuantATuple = append(node.BdedQuantATuple, true)
	return allSymbols, true
}

func (b *tlcBridge) convertFunctionArgs(args []Expr) tlc.SemanticNode {
	if len(args) == 1 {
		return b.convertExpr(args[0])
	}
	elems := make([]tlc.SemanticNode, 0, len(args))
	for _, arg := range args {
		elems = append(elems, b.convertExpr(arg))
	}
	return b.builtinNode(tlc.OpTup, elems...)
}

func (b *tlcBridge) symbol(name string) *tlc.SymbolNode {
	name = tlcSymbolName(name)
	if sym := b.symbols[name]; sym != nil {
		return sym
	}
	sym := tlc.NewSymbolNode(name)
	b.symbols[name] = sym
	return sym
}

func tlcBinaryOperator(op string) string {
	switch op {
	case "/\\":
		return "\\land"
	case "\\/":
		return "\\lor"
	case "<=>":
		return "\\equiv"
	case "\\subseteq", "\\subset":
		return "\\subseteq"
	default:
		return tlcSymbolName(op)
	}
}

func tlcSymbolName(name string) string {
	switch name {
	case "~", "\\neg":
		return "\\lnot"
	case "/\\":
		return "\\land"
	case "\\/":
		return "\\lor"
	case "<=>":
		return "\\equiv"
	case "=<":
		return "\\leq"
	case ">=":
		return "\\geq"
	default:
		return name
	}
}

func setComprehensionElementIsBound(e *SetComprehensionExpr) bool {
	if e == nil {
		return false
	}
	if len(e.Bounds) == 1 {
		ident, ok := e.Element.(*IdentExpr)
		return ok && ident.Name == e.Bounds[0].Name
	}
	tuple, ok := e.Element.(*TupleExpr)
	if !ok || len(tuple.Elems) != len(e.Bounds) {
		return false
	}
	for i, elem := range tuple.Elems {
		ident, ok := elem.(*IdentExpr)
		if !ok || ident.Name != e.Bounds[i].Name {
			return false
		}
	}
	return true
}
