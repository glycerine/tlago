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
	tool                  *tlc.Tool
	processor             *tlc.SpecProcessor
	spec                  *Spec
	cfg                   *tlc.ModelConfig
	runtime               tlc.RuntimeParameters
	defs                  map[string]*Definition
	defns                 *tlc.Defns
	diags                 Diagnostics
	symbols               map[string]*tlc.SymbolNode
	rootModuleName        string
	moduleDefinitionNames map[string]map[string]bool
	convertingModule      string
	convertBoundNames     map[string]int
	definitionModules     map[*Definition]string
	sourceDefinitions     map[*Definition]*tlc.OpDefNode
	instanceDefinitions   map[string]*tlcBridgeInstance
}

type tlcBridgeInstance struct {
	owner  *Module
	inst   Instance
	substs []tlc.Subst
	params []*tlc.SymbolNode
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
		"TLCEvalDefinition", "TLCGetOrDefault", "TLCGetAndSet",
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
	bridge.installDefinitions()
	bridge.installRuntimeConstants()
	bridge.installConfigConstants()
	bridge.installInstanceAliases()
	bridge.installModelTargets()
	bridge.installAssumptions()
	bridge.installRuntimeParameters()
	bridge.tool.AssignActionIDs()
	return bridge.tool, bridge.diags
}

func (b *tlcBridge) installVariables() {
	vars := moduleVariables(b.spec.Root)
	locations := moduleVariableLocations(b.spec.Root)
	tlc.SetStateVariablesWithLocations(vars, locations)
	nodes := make([]*tlc.SymbolNode, len(vars))
	for i, name := range vars {
		nodes[i] = b.symbol(name)
		nodes[i].MarkVariableDecl()
		nodes[i].Location = locations[name]
	}
	if b.processor != nil {
		b.processor.SetVariableNodes(nodes)
	}
}

func (b *tlcBridge) installDefinitions() {
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
	if b.definitionModules != nil || b.spec == nil {
		return
	}
	b.definitionModules = map[*Definition]string{}
	b.sourceDefinitions = map[*Definition]*tlc.OpDefNode{}
	b.instanceDefinitions = map[string]*tlcBridgeInstance{}
	names := make([]string, 0, len(b.spec.Modules))
	for name := range b.spec.Modules {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		mod := b.spec.Modules[name]
		if mod == nil {
			continue
		}
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
				for _, key := range keys {
					b.defs[key] = def
					b.instanceDefinitions[key] = binding
				}
			}
		}
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
		b.symbols[alias] = b.symbol(canonical)
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
	value := b.tool.DefnsByName[tlc.UniqueStringOf(module+"!"+member)]
	if value == nil {
		value = b.tool.DefnsByName[tlc.UniqueStringOf(member)]
	}
	if value == nil {
		return
	}
	var opDef *tlc.OpDefNode
	if _, ok := value.(*tlc.EvaluatingValue); ok {
		if opDef = b.convertDefinitionAs(name, def); opDef != nil {
			value = tlc.WithEvaluatingOpDef(value, opDef)
		}
	} else if _, ok := value.(*tlc.PriorityEvaluatingValue); ok {
		if opDef = b.convertDefinitionAs(name, def); opDef != nil {
			value = tlc.WithEvaluatingOpDef(value, opDef)
		}
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
	b.installConfigOverridesUnder("", b.cfg.GetOverrides())
	for moduleName, overrides := range b.cfg.GetModOverrides().All() {
		b.installConfigOverridesUnder(moduleName+"!", overrides)
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

func (b *tlcBridge) installConfigOverridesUnder(prefix string, overrides *tlc.InsMap[string, string]) {
	if overrides == nil {
		return
	}
	for specName, configName := range overrides.All() {
		def := b.defs[configName]
		if def == nil && prefix != "" {
			def = b.defs[prefix+configName]
		}
		qualifiedSpecName := prefix + specName
		if def == nil {
			b.diags = append(b.diags, errorAt(Position{}, "E7002", "CONSTANT override %s <- %s references an unknown operator", qualifiedSpecName, configName))
			continue
		}
		opDef := b.convertDefinitionAs(qualifiedSpecName, def)
		if opDef != nil {
			b.defineAlias(qualifiedSpecName, opDef)
		}
	}
}

func (b *tlcBridge) installInstanceAliases() {
	if b == nil || b.spec == nil || b.spec.Root == nil {
		return
	}
	for _, inst := range b.spec.Root.Instances {
		for _, binding := range b.standardInstanceBindings(inst) {
			if binding.Symbol != nil {
				b.define(binding.Symbol, binding.Value)
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
		out = append(out, tlc.LetBinding{
			Symbol: b.symbol(inst.qualifier() + "!" + member),
			Value:  value,
		})
	}
	return out
}

func (b *tlcBridge) installAssumptions() {
	if b.spec == nil || b.spec.Root == nil {
		return
	}
	for _, assumption := range b.spec.Root.Assumptions {
		if assumption.Expr == nil {
			continue
		}
		expr := b.convertExpr(assumption.Expr)
		if expr != nil {
			b.tool.Assumptions = append(b.tool.Assumptions, expr)
			b.tool.AssumptionIsAxiom = append(b.tool.AssumptionIsAxiom, false)
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
	if name := b.cfg.GetAlias(); name != "" {
		b.installAliasTarget(name)
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
	b.defineName(name, tlc.NewStringValue(value))
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

func (b *tlcBridge) parseDebuggerExpression(tool *tlc.Tool, location tlc.SourceLocation, condition string) (*tlc.OpDefNode, error) {
	_ = tool
	condition = strings.TrimSpace(condition)
	if condition == "" || strings.EqualFold(condition, "TRUE") {
		return nil, nil
	}
	if b == nil || b.spec == nil || b.spec.Root == nil {
		return nil, fmt.Errorf("cannot parse debugger expression %q without a root module", condition)
	}

	rootName := b.spec.Root.Name
	moduleName := b.unusedDebuggerModuleName()
	opName := b.unusedDebuggerOpName()
	scope := b.debuggerExpressionScope(location)
	var source strings.Builder
	source.WriteString("---- MODULE ")
	source.WriteString(moduleName)
	source.WriteString(" ----\nEXTENDS ")
	source.WriteString(rootName)
	source.WriteString("\n")
	source.WriteString(opName)
	if params := scope.paramSignatures(); len(params) > 0 {
		source.WriteString("(")
		source.WriteString(strings.Join(params, ", "))
		source.WriteString(")")
	}
	source.WriteString(" == ")
	if letStubs := scope.letStubSignatures(); len(letStubs) > 0 {
		source.WriteString("LET\n")
		for _, stub := range letStubs {
			source.WriteString(stub)
			source.WriteString(" == TRUE\n")
		}
		source.WriteString("IN ")
	}
	source.WriteString(condition)
	source.WriteString("\n====\n")

	file := moduleName + ".tla"
	if b.spec.Root.SourcePath != "" {
		file = filepath.Join(filepath.Dir(b.spec.Root.SourcePath), file)
	}
	mod, parseDiags := ParseSanyModuleSource(file, source.String())
	if parseDiags.HasErrors() || mod == nil {
		return nil, fmt.Errorf("syntax error while parsing breakpoint expression %q:\n%s", condition, parseDiags.Error())
	}

	modules := make(map[string]*Module, len(b.spec.Modules)+1)
	for name, module := range b.spec.Modules {
		modules[name] = module
	}
	modules[moduleName] = mod
	semanticOrder := append([]string(nil), b.spec.SemanticOrder...)
	semanticOrder = append(semanticOrder, moduleName)
	wrapped := &Spec{Root: mod, Modules: modules, SemanticOrder: semanticOrder}
	if checkDiags := CheckSpec(wrapped); checkDiags.HasErrors() {
		return nil, fmt.Errorf("semantic error while parsing breakpoint expression %q:\n%s", condition, checkDiags.Error())
	}

	var def *Definition
	for i := range mod.Definitions {
		if mod.Definitions[i].Name == opName {
			def = &mod.Definitions[i]
			break
		}
	}
	if def == nil {
		return nil, fmt.Errorf("unable to find debugger expression op %s", opName)
	}

	convert := &tlcBridge{
		tool:                  b.tool,
		processor:             b.processor,
		spec:                  wrapped,
		cfg:                   b.cfg,
		runtime:               b.runtime,
		defs:                  tlcBridgeDefinitionsByName(wrapped),
		defns:                 b.defns,
		symbols:               b.symbols,
		rootModuleName:        wrapped.Root.Name,
		moduleDefinitionNames: moduleDefinitionNameIndex(wrapped),
		convertBoundNames:     map[string]int{},
	}
	op := convert.convertDefinitionAs(opName, def)
	if op == nil {
		return nil, fmt.Errorf("unable to convert debugger expression %q", condition)
	}
	if len(scope.letDefs) > 0 {
		lets := make([]*tlc.OpDefNode, 0, len(scope.letDefs))
		for _, letDef := range scope.letDefs {
			defCopy := letDef
			letOp := convert.convertDefinitionAs(defCopy.Name, &defCopy)
			if letOp != nil {
				lets = append(lets, letOp)
			}
		}
		replaceDebuggerLetStubs(op.Body, lets)
	}
	if convert.diags.HasErrors() {
		return nil, fmt.Errorf("semantic error while converting breakpoint expression %q:\n%s", condition, convert.diags.Error())
	}
	return op, nil
}

type debuggerScopedOperator struct {
	name  string
	arity int
}

type debuggerExpressionScope struct {
	params   []debuggerScopedOperator
	paramSet map[string]bool
	letDefs  []Definition
	letSet   map[string]bool
	letNames map[string]bool
	letStubs []string
	stubSet  map[string]bool
}

func (s *debuggerExpressionScope) addParam(name string, arity int) {
	if s == nil || name == "" || s.letNames[name] {
		return
	}
	signature := debuggerOperatorSignature(name, arity)
	if s.paramSet == nil {
		s.paramSet = map[string]bool{}
	}
	if s.paramSet[signature] {
		return
	}
	s.paramSet[signature] = true
	s.params = append(s.params, debuggerScopedOperator{name: name, arity: arity})
}

func (s *debuggerExpressionScope) addDefinitionParams(def Definition) {
	for _, name := range def.Params {
		s.addParam(name, def.ParamArities[name])
	}
}

func (s *debuggerExpressionScope) addBound(bound BoundVar) {
	arity := 0
	if bound.HasOperatorArity {
		arity = bound.OperatorArity
	}
	s.addParam(bound.Name, arity)
}

func (s *debuggerExpressionScope) addLet(def Definition) {
	if s == nil || def.Name == "" {
		return
	}
	if s.letSet == nil {
		s.letSet = map[string]bool{}
	}
	if s.letNames == nil {
		s.letNames = map[string]bool{}
	}
	signature := debuggerDefinitionSignature(def)
	if s.letSet[signature] {
		return
	}
	s.letSet[signature] = true
	s.letNames[def.Name] = true
	s.letDefs = append(s.letDefs, def)
}

func (s *debuggerExpressionScope) paramSignatures() []string {
	if s == nil || len(s.params) == 0 {
		return nil
	}
	out := make([]string, 0, len(s.params))
	for _, param := range s.params {
		if s.letNames[param.name] {
			continue
		}
		out = append(out, debuggerOperatorSignature(param.name, param.arity))
	}
	sort.Strings(out)
	return out
}

func (s *debuggerExpressionScope) addLetStub(signature string) {
	if s == nil || signature == "" {
		return
	}
	if s.stubSet == nil {
		s.stubSet = map[string]bool{}
	}
	if s.stubSet[signature] {
		return
	}
	s.stubSet[signature] = true
	s.letStubs = append(s.letStubs, signature)
}

func (s *debuggerExpressionScope) letStubSignatures() []string {
	if s == nil || len(s.letDefs) == 0 {
		return nil
	}
	for _, def := range s.letDefs {
		s.addLetStub(debuggerDefinitionSignature(def))
	}
	out := append([]string(nil), s.letStubs...)
	sort.Strings(out)
	return out
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

func debuggerDefinitionSignature(def Definition) string {
	if len(def.Params) == 0 {
		return def.Name
	}
	parts := make([]string, len(def.Params))
	for i, param := range def.Params {
		parts[i] = debuggerOperatorSignature(param, def.ParamArities[param])
	}
	return def.Name + "(" + strings.Join(parts, ", ") + ")"
}

func replaceDebuggerLetStubs(node tlc.SemanticNode, lets []*tlc.OpDefNode) {
	if len(lets) == 0 {
		return
	}
	if letIn, ok := node.(*tlc.LetInNode); ok && letIn != nil {
		letIn.Lets = lets
	}
}

func (b *tlcBridge) debuggerExpressionScope(location tlc.SourceLocation) *debuggerExpressionScope {
	scope := &debuggerExpressionScope{letNames: map[string]bool{}}
	if b == nil || b.spec == nil || location.IsNull() {
		return scope
	}
	mod := b.debuggerModuleForLocation(location)
	if mod == nil {
		return scope
	}
	point := Position{Line: location.BeginLine, Column: location.BeginColumn, EndLine: location.BeginLine, EndColumn: location.BeginColumn}
	for i := range mod.Definitions {
		def := mod.Definitions[i]
		if !exprIncludesPoint(def.Expr, point) {
			continue
		}
		scope.addDefinitionParams(def)
		collectDebuggerExpressionScope(def.Expr, point, scope)
	}
	for _, assumption := range mod.Assumptions {
		if exprIncludesPoint(assumption.Expr, point) {
			collectDebuggerExpressionScope(assumption.Expr, point, scope)
		}
	}
	for _, theorem := range mod.Theorems {
		if exprIncludesPoint(theorem.Expr, point) {
			collectDebuggerExpressionScope(theorem.Expr, point, scope)
		}
	}
	return scope
}

func (b *tlcBridge) debuggerModuleForLocation(location tlc.SourceLocation) *Module {
	if b == nil || b.spec == nil {
		return nil
	}
	if location.Source != "" {
		if mod := b.spec.Modules[location.Source]; mod != nil {
			return mod
		}
		sourceBase := strings.TrimSuffix(filepath.Base(location.Source), filepath.Ext(location.Source))
		if mod := b.spec.Modules[sourceBase]; mod != nil {
			return mod
		}
		moduleNames := make([]string, 0, len(b.spec.Modules))
		for name := range b.spec.Modules {
			moduleNames = append(moduleNames, name)
		}
		sort.Strings(moduleNames)
		for _, name := range moduleNames {
			mod := b.spec.Modules[name]
			if mod == nil {
				continue
			}
			modBase := strings.TrimSuffix(filepath.Base(mod.SourcePath), filepath.Ext(mod.SourcePath))
			if mod.Name == location.Source || modBase == sourceBase {
				return mod
			}
		}
	}
	return b.spec.Root
}

func exprIncludesPoint(expr Expr, point Position) bool {
	if expr == nil || point.Line == 0 {
		return false
	}
	pos := expr.Position()
	if pos.Line == 0 {
		return false
	}
	return pos.Includes(point)
}

func collectDebuggerExpressionScope(expr Expr, point Position, scope *debuggerExpressionScope) {
	if !exprIncludesPoint(expr, point) || scope == nil {
		return
	}
	switch e := expr.(type) {
	case *UnaryExpr:
		collectDebuggerExpressionScope(e.Expr, point, scope)
	case *BinaryExpr:
		collectDebuggerExpressionScope(e.Left, point, scope)
		collectDebuggerExpressionScope(e.Right, point, scope)
	case *CallExpr:
		collectDebuggerExpressionScope(e.Callee, point, scope)
		for _, arg := range e.Args {
			collectDebuggerExpressionScope(arg, point, scope)
		}
	case *IfExpr:
		collectDebuggerExpressionScope(e.Cond, point, scope)
		collectDebuggerExpressionScope(e.Then, point, scope)
		collectDebuggerExpressionScope(e.Else, point, scope)
	case *LetExpr:
		for _, def := range e.Definitions {
			scope.addLet(def)
		}
		for _, def := range e.Definitions {
			if exprIncludesPoint(def.Expr, point) {
				scope.addDefinitionParams(def)
				collectDebuggerExpressionScope(def.Expr, point, scope)
			}
		}
		collectDebuggerExpressionScope(e.Body, point, scope)
	case *QuantifierExpr:
		if exprIncludesPoint(e.Body, point) {
			scope.addParam(e.Var, debuggerBoundArity(e.HasOperatorArity, e.OperatorArity))
		}
		collectDebuggerExpressionScope(e.Set, point, scope)
		collectDebuggerExpressionScope(e.Body, point, scope)
	case *CaseExpr:
		for _, arm := range e.Arms {
			collectDebuggerExpressionScope(arm.Test, point, scope)
			collectDebuggerExpressionScope(arm.Value, point, scope)
		}
		collectDebuggerExpressionScope(e.Other, point, scope)
	case *ChooseExpr:
		if exprIncludesPoint(e.Body, point) {
			scope.addParam(e.Var, 0)
		}
		collectDebuggerExpressionScope(e.Set, point, scope)
		collectDebuggerExpressionScope(e.Body, point, scope)
	case *TupleExpr:
		for _, elem := range e.Elems {
			collectDebuggerExpressionScope(elem, point, scope)
		}
	case *SetExpr:
		for _, elem := range e.Elems {
			collectDebuggerExpressionScope(elem, point, scope)
		}
	case *RecordExpr:
		for _, field := range e.Fields {
			collectDebuggerExpressionScope(field.Value, point, scope)
		}
	case *RecordComponentExpr:
		collectDebuggerExpressionScope(e.Record, point, scope)
	case *RecordSetExpr:
		for _, field := range e.Fields {
			collectDebuggerExpressionScope(field.Set, point, scope)
		}
	case *FunctionExpr:
		if exprIncludesPoint(e.Body, point) {
			for _, bound := range e.Bounds {
				scope.addBound(bound)
			}
		}
		for _, bound := range e.Bounds {
			collectDebuggerExpressionScope(bound.Set, point, scope)
		}
		collectDebuggerExpressionScope(e.Body, point, scope)
	case *FunctionAppExpr:
		collectDebuggerExpressionScope(e.Function, point, scope)
		for _, arg := range e.Args {
			collectDebuggerExpressionScope(arg, point, scope)
		}
	case *ExceptExpr:
		collectDebuggerExpressionScope(e.Base, point, scope)
		for _, spec := range e.Specs {
			for _, component := range spec.Components {
				for _, index := range component.Indices {
					collectDebuggerExpressionScope(index, point, scope)
				}
			}
			collectDebuggerExpressionScope(spec.Value, point, scope)
		}
	case *LabelExpr:
		collectDebuggerExpressionScope(e.Body, point, scope)
	case *ActionExpr:
		collectDebuggerExpressionScope(e.Action, point, scope)
		collectDebuggerExpressionScope(e.Subscript, point, scope)
	case *FairnessExpr:
		collectDebuggerExpressionScope(e.Action, point, scope)
		collectDebuggerExpressionScope(e.Subscript, point, scope)
	case *FunctionSetExpr:
		collectDebuggerExpressionScope(e.Domain, point, scope)
		collectDebuggerExpressionScope(e.Range, point, scope)
	case *SetComprehensionExpr:
		if exprIncludesPoint(e.Element, point) || exprIncludesPoint(e.Predicate, point) {
			for _, bound := range e.Bounds {
				scope.addBound(bound)
			}
		}
		for _, bound := range e.Bounds {
			collectDebuggerExpressionScope(bound.Set, point, scope)
		}
		collectDebuggerExpressionScope(e.Element, point, scope)
		collectDebuggerExpressionScope(e.Predicate, point, scope)
	}
}

func debuggerBoundArity(hasArity bool, arity int) int {
	if hasArity {
		return arity
	}
	return 0
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

func (b *tlcBridge) installAliasTarget(name string) {
	alias := b.nodeForDefinition(name)
	if alias == nil {
		return
	}
	evalAlias := func(tool *tlc.Tool, curState *tlc.TLCStateMut, sucState *tlc.TLCStateMut) *tlc.TLCStateMut {
		if curState == nil {
			return curState
		}
		value, err := tool.Eval(alias, tlc.EmptyContext, curState, sucState, tlc.EvalClear, tlc.CostModel{})
		if err != nil {
			return curState
		}
		record, ok := value.(*tlc.RecordValue)
		if !ok {
			return curState
		}
		if state := record.ToState(); state != nil {
			return state
		}
		return curState
	}
	b.tool.EvalAliasFunc = evalAlias
	b.tool.EvalAliasInfoFunc = func(tool *tlc.Tool, current *tlc.TLCStateInfo, successor *tlc.TLCStateMut, prefix func() []*tlc.TLCStateInfo) (*tlc.TLCStateInfo, error) {
		_ = prefix
		if current == nil || current.State == nil {
			return current, nil
		}
		aliasState := evalAlias(tool, current.State, successor)
		if aliasState == current.State {
			return current, nil
		}
		return tlc.AliasTLCStateInfo(aliasState, current), nil
	}
	b.tool.EvalAliasInfoPairFunc = func(tool *tlc.Tool, current *tlc.TLCStateInfo, successor *tlc.TLCStateMut) (*tlc.TLCStateInfo, error) {
		if current == nil || current.State == nil {
			return current, nil
		}
		aliasState := evalAlias(tool, current.State, successor)
		if aliasState == current.State {
			return current, nil
		}
		return tlc.AliasTLCStateInfo(aliasState, current), nil
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
	// SANY aliases share the original OpDefNode and body, rather than
	// reconstructing a definition for each imported name.
	if source := b.sourceDefinitions[def]; source != nil {
		return source
	}
	sym := b.symbol(name)
	params := make([]*tlc.SymbolNode, len(def.Params))
	for i, param := range def.Params {
		params[i] = b.symbol(param)
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
	if mod := b.spec.Modules[b.convertingModule]; mod != nil {
		opDef.SetInRecursive(recursiveDeclarationSections(mod.Recursives)[def.Name] != 0)
	}
	b.withPositionLocation(def.SourcePosition(), opDef)
	if declaration := b.sourceLocationForPosition(def.DeclarationPosition()); !declaration.IsNull() {
		opDef.SetDeclarationLocation(declaration)
	}
	b.sourceDefinitions[def] = opDef
	return opDef
}

func (b *tlcBridge) convertInstanceDefinition(name string, def *Definition, binding *tlcBridgeInstance) *tlc.OpDefNode {
	inst := binding.inst
	source := b.convertDefinitionAs(inst.Module+"!"+def.Name, def)
	if source == nil {
		return nil
	}
	b.define(source.Symbol, source)
	params := make([]*tlc.SymbolNode, 0, len(inst.Params)+len(source.Params))
	if binding.params == nil {
		binding.params = make([]*tlc.SymbolNode, len(inst.Params))
		for i, param := range inst.Params {
			binding.params[i] = tlc.NewSymbolNode(param)
		}
	}
	params = append(params, binding.params...)
	params = append(params, source.Params...)
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
	body := source.Body
	if len(binding.substs) > 0 {
		body = b.withPositionLocation(inst.SourcePosition(), tlc.NewSubstInNode(body, binding.substs...))
	}
	clone := tlc.NewOpDefNodeForSymbol(b.symbol(name), params, body)
	b.withPositionLocation(inst.SourcePosition(), clone)
	if !positionIsZero(inst.LHSPos) {
		clone.SetDeclarationLocation(b.sourceLocationForPosition(inst.LHSPos))
	}
	return clone
}

func (b *tlcBridge) instanceTargets(mod *Module, visiting map[string]bool) []tlcBridgeInstanceTarget {
	if mod == nil || visiting[mod.Name] || isEmbeddedStandardModule(mod) {
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
		return b.symbol(name)
	}
	return b.symbol(mod.Name + "!" + name)
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
	return b.symbol(b.resolveExprName(name))
}

func (b *tlcBridge) convertExpr(expr Expr) tlc.SemanticNode {
	if expr == nil {
		return nil
	}
	var node tlc.SemanticNode
	switch e := expr.(type) {
	case *IdentExpr:
		node = tlc.NewOpApplNode(b.exprSymbol(e.Name))
	case *LiteralExpr:
		node = b.convertLiteral(e)
	case *UnaryExpr:
		node = b.unaryNode(e)
	case *BinaryExpr:
		node = b.binaryNode(e)
	case *CallExpr:
		node = b.callNode(e)
	case *IfExpr:
		node = tlc.NewBuiltinOpApplNode(tlc.OpITE, b.convertExpr(e.Cond), b.convertExpr(e.Then), b.convertExpr(e.Else))
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
		node = tlc.NewBuiltinOpApplNode(tlc.OpTup, args...)
	case *SetExpr:
		args := make([]tlc.SemanticNode, 0, len(e.Elems))
		for _, elem := range e.Elems {
			args = append(args, b.convertExpr(elem))
		}
		node = tlc.NewBuiltinOpApplNode(tlc.OpSE, args...)
	case *RecordExpr:
		args := make([]tlc.SemanticNode, 0, len(e.Fields))
		for _, field := range e.Fields {
			fieldPos := field.Source
			if positionIsZero(fieldPos) {
				fieldPos = field.Pos
			}
			name := b.withPositionLocation(fieldPos, tlc.NewStringNode(field.Name))
			pair := tlc.NewBuiltinOpApplNode(tlc.OpPair, name, b.convertExpr(field.Value))
			args = append(args, b.withPositionLocation(fieldPos, pair))
		}
		node = tlc.NewBuiltinOpApplNode(tlc.OpRC, args...)
	case *RecordComponentExpr:
		field := b.withPositionLocation(e.FieldPos, tlc.NewStringNode(e.Field))
		node = tlc.NewBuiltinOpApplNode(tlc.OpRS, b.convertExpr(e.Record), field)
	case *RecordSetExpr:
		args := make([]tlc.SemanticNode, 0, len(e.Fields))
		for _, field := range e.Fields {
			fieldPos := field.Source
			if positionIsZero(fieldPos) {
				fieldPos = field.Pos
			}
			name := b.withPositionLocation(fieldPos, tlc.NewStringNode(field.Name))
			pair := tlc.NewBuiltinOpApplNode(tlc.OpPair, name, b.convertExpr(field.Set))
			args = append(args, b.withPositionLocation(fieldPos, pair))
		}
		node = tlc.NewBuiltinOpApplNode(tlc.OpSOR, args...)
	case *FunctionExpr:
		if e.IsLambda {
			node = b.lambdaNode(e)
		} else {
			node = b.functionNode(e)
		}
	case *FunctionAppExpr:
		arg := b.convertFunctionArgs(e.Args)
		node = tlc.NewBuiltinOpApplNode(tlc.OpFA, b.convertExpr(e.Function), arg)
	case *ExceptExpr:
		node = b.exceptNode(e)
	case *LabelExpr:
		node = tlc.NewLabelNode(b.convertExpr(e.Body))
	case *ActionExpr:
		op := tlc.OpSA
		if e.Kind == "angle" {
			op = tlc.OpAA
		}
		node = tlc.NewBuiltinOpApplNode(op, b.convertExpr(e.Action), b.convertExpr(e.Subscript))
	case *FairnessExpr:
		op := tlc.OpWF
		if e.Kind == "SF" {
			op = tlc.OpSF
		}
		node = tlc.NewBuiltinOpApplNode(op, b.convertExpr(e.Action), b.convertExpr(e.Subscript))
	case *FunctionSetExpr:
		node = tlc.NewBuiltinOpApplNode(tlc.OpSOF, b.convertExpr(e.Domain), b.convertExpr(e.Range))
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
	return b.withPositionLocation(expr.Position(), node)
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
			b.diags = append(b.diags, errorAt(e.Pos, "E7010", "TLC can't handle real numbers.\n%s", e.Value))
			return tlc.NewValueNode(tlc.ValUndef)
		}
		value, err := strconv.ParseInt(e.Value, 10, 32)
		if err != nil {
			b.diags = append(b.diags, errorAt(e.Pos, "E7011", "integer literal %s is outside TLC int32 range", e.Value))
			return tlc.NewValueNode(tlc.ValUndef)
		}
		return tlc.NewNumeralNode(int32(value))
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
		return tlc.NewBuiltinOpApplNode(tlc.OpCL, b.convertExpr(e.Expr))
	case "\\/":
		return tlc.NewBuiltinOpApplNode(tlc.OpDL, b.convertExpr(e.Expr))
	case "~", "\\neg":
		op = "\\lnot"
	case "-.":
		op = "-"
	}
	return tlc.NewOpApplNode(b.exprSymbol(op), b.convertExpr(e.Expr))
}

func (b *tlcBridge) binaryNode(e *BinaryExpr) tlc.SemanticNode {
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
		return tlc.NewBuiltinOpApplNode(op, args...)
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
			return tlc.NewBuiltinOpApplNode(tlc.OpFA, args[0], args[1])
		}
		return tlc.NewBuiltinOpApplNode(tlc.OpFA, args[0], tlc.NewBuiltinOpApplNode(tlc.OpTup, args[1:]...))
	}
	args := make([]tlc.SemanticNode, 0, len(e.Args))
	for _, arg := range e.Args {
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
		lets = append(lets, b.instanceOpDefinitions(inst)...)
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
	if e.Set != nil && e.TupleBound {
		if node, ok := b.tupleQuantifierNode(e, op); ok {
			return node
		}
	}
	var bound tlc.SemanticNode
	if e.Set != nil {
		bound = b.convertExpr(e.Set)
	}
	restore := b.pushConvertBoundNames(e.Var)
	body := b.convertExpr(e.Body)
	restore()
	node := tlc.NewBuiltinOpApplNode(op, body)
	if e.Set != nil {
		node.BdedQuantSymbolLists = [][]*tlc.SymbolNode{{b.symbol(e.Var)}}
		node.BdedQuantBounds = []tlc.SemanticNode{bound}
		node.BdedQuantATuple = []bool{e.TupleBound}
	} else {
		node.UnbdedQuantSymbols = []*tlc.SymbolNode{b.symbol(e.Var)}
	}
	return node
}

func (b *tlcBridge) tupleQuantifierNode(e *QuantifierExpr, op *tlc.UniqueString) (tlc.SemanticNode, bool) {
	if e == nil || e.Set == nil || !e.TupleBound {
		return nil, false
	}
	vars := []string{e.Var}
	body := e.Body
	for {
		next, ok := body.(*QuantifierExpr)
		if !ok || next.Set == nil || !next.TupleBound || next.Kind != e.Kind || next.Set != e.Set {
			break
		}
		vars = append(vars, next.Var)
		body = next.Body
	}
	if len(vars) == 1 {
		return nil, false
	}
	bound := b.convertExpr(e.Set)
	restore := b.pushConvertBoundNames(vars...)
	convertedBody := b.convertExpr(body)
	restore()
	node := tlc.NewBuiltinOpApplNode(op, convertedBody)
	symbols := make([]*tlc.SymbolNode, 0, len(vars))
	for _, name := range vars {
		symbols = append(symbols, b.symbol(name))
	}
	node.BdedQuantSymbolLists = [][]*tlc.SymbolNode{symbols}
	node.BdedQuantBounds = []tlc.SemanticNode{bound}
	node.BdedQuantATuple = []bool{true}
	return node, true
}

func (b *tlcBridge) caseNode(e *CaseExpr) tlc.SemanticNode {
	args := make([]tlc.SemanticNode, 0, len(e.Arms)+1)
	for _, arm := range e.Arms {
		pair := tlc.NewBuiltinOpApplNode(tlc.OpPair, b.convertExpr(arm.Test), b.convertExpr(arm.Value))
		args = append(args, b.withPositionLocation(arm.Pos, pair))
	}
	if e.Other != nil {
		pair := tlc.NewBuiltinOpApplNode(tlc.OpPair, nil, b.convertExpr(e.Other))
		args = append(args, b.withPositionLocation(e.OtherPos, pair))
	}
	return tlc.NewBuiltinOpApplNode(tlc.OpCase, args...)
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
	restore := b.pushConvertBoundNames(e.Var)
	body := b.convertExpr(e.Body)
	restore()
	node := tlc.NewBuiltinOpApplNode(op, body)
	if e.Set != nil {
		node.BdedQuantSymbolLists = [][]*tlc.SymbolNode{{b.symbol(e.Var)}}
		node.BdedQuantBounds = []tlc.SemanticNode{bound}
		node.BdedQuantATuple = []bool{false}
	} else {
		node.UnbdedQuantSymbols = []*tlc.SymbolNode{b.symbol(e.Var)}
	}
	return node
}

func (b *tlcBridge) lambdaNode(e *FunctionExpr) tlc.SemanticNode {
	params := make([]*tlc.SymbolNode, len(e.Bounds))
	names := make([]string, len(e.Bounds))
	previous := make([]*tlc.SymbolNode, len(e.Bounds))
	for i, bound := range e.Bounds {
		name := tlcSymbolName(bound.Name)
		names[i] = name
		previous[i] = b.symbols[name]
		params[i] = tlc.NewSymbolNode(name)
		b.symbols[name] = params[i]
	}
	restore := b.pushConvertBoundNames(names...)
	body := b.convertExpr(e.Body)
	restore()
	for i, name := range names {
		if previous[i] == nil {
			delete(b.symbols, name)
		} else {
			b.symbols[name] = previous[i]
		}
	}
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
		self = tlc.NewSymbolNode(def.Name)
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
	return b.withPositionLocation(def.SourcePosition(), node)
}

func (b *tlcBridge) functionNode(e *FunctionExpr) tlc.SemanticNode {
	priorBound := make([]string, 0, len(e.Bounds))
	boundExprs := make([]tlc.SemanticNode, 0, len(e.Bounds))
	for _, bound := range e.Bounds {
		restorePrior := b.pushConvertBoundNames(priorBound...)
		boundExprs = append(boundExprs, b.convertExpr(bound.Set))
		restorePrior()
		priorBound = append(priorBound, bound.Name)
	}
	restore := b.pushConvertBoundNames(priorBound...)
	body := b.convertExpr(e.Body)
	restore()
	node := tlc.NewBuiltinOpApplNode(tlc.OpFC, body)
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
		path := b.withPositionLocation(spec.Pos, tlc.NewBuiltinOpApplNode(tlc.OpTup, pathElems...))
		pair := tlc.NewBuiltinOpApplNode(tlc.OpPair, path, b.convertExpr(spec.Value))
		args = append(args, b.withPositionLocation(spec.Pos, pair))
	}
	return tlc.NewBuiltinOpApplNode(tlc.OpExc, args...)
}

func (b *tlcBridge) setComprehensionNode(e *SetComprehensionExpr) tlc.SemanticNode {
	priorBound := make([]string, 0, len(e.Bounds))
	boundExprs := make([]tlc.SemanticNode, 0, len(e.Bounds))
	for _, bound := range e.Bounds {
		restorePrior := b.pushConvertBoundNames(priorBound...)
		boundExprs = append(boundExprs, b.convertExpr(bound.Set))
		restorePrior()
		priorBound = append(priorBound, bound.Name)
	}
	groups := b.boundGroups(e.Bounds, boundExprs)
	if e.Predicate != nil {
		restore := b.pushConvertBoundNames(priorBound...)
		predicate := b.convertExpr(e.Predicate)
		restore()
		filtered := tlc.NewBuiltinOpApplNode(tlc.OpSSO, predicate)
		filterSymbols, filterTuple := b.appendFilteredComprehensionBounds(filtered, groups, e.Pos)
		if setComprehensionElementIsBound(e) {
			return filtered
		}
		bodyRestore := b.pushConvertBoundNames(priorBound...)
		body := b.convertExpr(e.Element)
		bodyRestore()
		node := tlc.NewBuiltinOpApplNode(tlc.OpSOA, body)
		node.BdedQuantSymbolLists = append(node.BdedQuantSymbolLists, filterSymbols)
		node.BdedQuantBounds = append(node.BdedQuantBounds, filtered)
		node.BdedQuantATuple = append(node.BdedQuantATuple, filterTuple)
		return node
	}
	restore := b.pushConvertBoundNames(priorBound...)
	body := b.convertExpr(e.Element)
	restore()
	node := tlc.NewBuiltinOpApplNode(tlc.OpSOA, body)
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
	product := b.withPositionLocation(pos, tlc.NewBuiltinOpApplNode(tlc.OpCP, productArgs...))
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
	return tlc.NewBuiltinOpApplNode(tlc.OpTup, elems...)
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
