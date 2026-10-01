package tlago

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/glycerine/tlago/tlc"
)

type tlcBridge struct {
	tool      *tlc.Tool
	processor *tlc.SpecProcessor
	spec      *Spec
	cfg       *tlc.ModelConfig
	runtime   tlc.RuntimeParameters
	defs      map[string]*Definition
	defns     *tlc.Defns
	diags     Diagnostics
	symbols   map[string]*tlc.SymbolNode
}

var bridgeStandardModuleMembers = map[string][]string{
	"Naturals": {
		"Nat", "+", "-", "*", "^", "<", ">", "\\leq", "\\geq", "%", "\\div", "..",
	},
	"Integers": {
		"Int", "Nat", "+", "-", "*", "^", "<", ">", "\\leq", "\\geq", "%", "\\div", "..", "-.",
	},
	"Sequences": {
		"Seq", "Len", "Head", "Tail", "Cons", "Append", "Concat", "\\o",
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
		"_TLCState",
	},
	"_TLCTrace": {
		"_TLCTraceDeserialize", "_TLCTraceSerialize", "_TLCState",
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
		"SVGElemToString", "NodeOfRingNetwork", "PointOnLine",
	},
	"SequencesExt": {
		"SetToSeq", "SetToSeqs", "Contains", "LongestCommonPrefix",
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
		"SVGElemToString", "NodeOfRingNetwork", "PointOnLine",
	),
	"SequencesExt": setOf(
		"SetToSeq", "SetToSeqs", "Contains", "LongestCommonPrefix",
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

// BuildTLCTool converts the production Go SANY semantic tree into the TLC
// runtime tree used by the mechanical TLC port.
func BuildTLCTool(spec *Spec, cfg *tlc.ModelConfig, runtime tlc.RuntimeParameters) (*tlc.Tool, Diagnostics) {
	if spec == nil || spec.Root == nil {
		return nil, Diagnostics{errorAt(Position{}, "E7000", "missing root module for TLC tool")}
	}
	if cfg == nil {
		cfg = tlc.NewModelConfig(spec.Root.Name)
	}
	defns := tlc.NewDefns()
	bridge := &tlcBridge{
		tool:      tlc.NewToolWithModelConfig(cfg),
		processor: tlc.NewSpecProcessor(spec.Root.Name, defns, cfg),
		spec:      spec,
		cfg:       cfg,
		runtime:   runtime,
		defs:      definitionsByName(spec),
		defns:     defns,
		symbols:   map[string]*tlc.SymbolNode{},
	}
	bridge.tool.RootName = spec.Root.Name
	bridge.tool.RootFile = spec.Root.SourcePath
	bridge.tool.SpecDir = ""
	bridge.tool.ConfigFile = spec.Root.Name
	bridge.tool.ParseDebuggerExpressionFunc = bridge.parseDebuggerExpression
	bridge.installVariables()
	bridge.installDefinitions()
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
	tlc.SetStateVariables(vars)
	if b.processor != nil {
		b.processor.SetVariables(vars)
	}
	for _, name := range vars {
		b.symbol(name)
	}
}

func (b *tlcBridge) installDefinitions() {
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
		b.define(opDef.Symbol, opDef)
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
	b.define(b.symbol(name), value)
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
			b.define(opDef.Symbol, opDef)
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
	b.define(sym, value)
	return sym
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
	for i := range mod.Definitions {
		def := &mod.Definitions[i]
		if def.Local {
			continue
		}
		instDef := instantiatedDefinition(def, inst.Substitutions)
		if instDef == nil {
			continue
		}
		opDef := b.convertDefinitionAs(inst.qualifier()+"!"+instDef.Name, instDef)
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
	b.installRuntimeConstants()
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
	for _, post := range b.runtime.PostConditions {
		if action := b.actionFromModuleDefinition(post.Module, post.Operator, false); action != nil {
			b.tool.PostConditionSpecs = append(b.tool.PostConditionSpecs, action)
		}
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
	return b.actionFromDefinition(name, init)
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
		tool:      b.tool,
		processor: b.processor,
		spec:      wrapped,
		cfg:       b.cfg,
		runtime:   b.runtime,
		defs:      definitionsByName(wrapped),
		defns:     b.defns,
		symbols:   b.symbols,
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
	return b.actionFromExpr(name, def.Expr, opDef, init)
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
	sym := b.symbol(name)
	params := make([]*tlc.SymbolNode, len(def.Params))
	for i, param := range def.Params {
		params[i] = b.symbol(param)
	}
	body := b.convertExpr(def.Expr)
	if body == nil {
		return nil
	}
	return tlc.NewOpDefNodeForSymbol(sym, params, body)
}

func (b *tlcBridge) convertExpr(expr Expr) tlc.SemanticNode {
	switch e := expr.(type) {
	case nil:
		return nil
	case *IdentExpr:
		return tlc.NewOpApplNode(b.symbol(e.Name))
	case *LiteralExpr:
		return b.convertLiteral(e)
	case *UnaryExpr:
		return b.unaryNode(e)
	case *BinaryExpr:
		return b.binaryNode(e)
	case *CallExpr:
		return b.callNode(e)
	case *IfExpr:
		return tlc.NewBuiltinOpApplNode(tlc.OpITE, b.convertExpr(e.Cond), b.convertExpr(e.Then), b.convertExpr(e.Else))
	case *LetExpr:
		return b.letNode(e)
	case *QuantifierExpr:
		return b.quantifierNode(e)
	case *CaseExpr:
		return b.caseNode(e)
	case *ChooseExpr:
		return b.chooseNode(e)
	case *TupleExpr:
		args := make([]tlc.SemanticNode, 0, len(e.Elems))
		for _, elem := range e.Elems {
			args = append(args, b.convertExpr(elem))
		}
		return tlc.NewBuiltinOpApplNode(tlc.OpTup, args...)
	case *SetExpr:
		args := make([]tlc.SemanticNode, 0, len(e.Elems))
		for _, elem := range e.Elems {
			args = append(args, b.convertExpr(elem))
		}
		return tlc.NewBuiltinOpApplNode(tlc.OpSE, args...)
	case *RecordExpr:
		args := make([]tlc.SemanticNode, 0, len(e.Fields))
		for _, field := range e.Fields {
			args = append(args, tlc.NewBuiltinOpApplNode(tlc.OpPair, tlc.NewStringNode(field.Name), b.convertExpr(field.Value)))
		}
		return tlc.NewBuiltinOpApplNode(tlc.OpRC, args...)
	case *RecordComponentExpr:
		return tlc.NewBuiltinOpApplNode(tlc.OpRS, b.convertExpr(e.Record), tlc.NewStringNode(e.Field))
	case *RecordSetExpr:
		args := make([]tlc.SemanticNode, 0, len(e.Fields))
		for _, field := range e.Fields {
			args = append(args, tlc.NewBuiltinOpApplNode(tlc.OpPair, tlc.NewStringNode(field.Name), b.convertExpr(field.Set)))
		}
		return tlc.NewBuiltinOpApplNode(tlc.OpSOR, args...)
	case *FunctionExpr:
		return b.functionNode(e)
	case *FunctionAppExpr:
		arg := b.convertFunctionArgs(e.Args)
		return tlc.NewBuiltinOpApplNode(tlc.OpFA, b.convertExpr(e.Function), arg)
	case *ExceptExpr:
		return b.exceptNode(e)
	case *LabelExpr:
		return tlc.NewLabelNode(b.convertExpr(e.Body))
	case *ActionExpr:
		op := tlc.OpSA
		if e.Kind == "angle" {
			op = tlc.OpAA
		}
		return tlc.NewBuiltinOpApplNode(op, b.convertExpr(e.Action), b.convertExpr(e.Subscript))
	case *FairnessExpr:
		op := tlc.OpWF
		if e.Kind == "SF" {
			op = tlc.OpSF
		}
		return tlc.NewBuiltinOpApplNode(op, b.convertExpr(e.Action), b.convertExpr(e.Subscript))
	case *FunctionSetExpr:
		return tlc.NewBuiltinOpApplNode(tlc.OpSOF, b.convertExpr(e.Domain), b.convertExpr(e.Range))
	case *SetComprehensionExpr:
		return b.setComprehensionNode(e)
	default:
		b.diags = append(b.diags, errorAt(expr.Position(), "E7009", "unsupported expression %T in TLC bridge", expr))
		return tlc.NewValueNode(tlc.ValUndef)
	}
}

func (b *tlcBridge) convertLiteral(e *LiteralExpr) tlc.SemanticNode {
	switch e.Kind {
	case "bool":
		return tlc.NewValueNode(tlc.NewBoolValue(strings.EqualFold(e.Value, "TRUE")))
	case "number":
		if strings.Contains(e.Value, ".") {
			b.diags = append(b.diags, errorAt(e.Pos, "E7010", "decimal literal %s is not yet supported by the TLC bridge", e.Value))
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
	case "~", "\\neg":
		op = "\\lnot"
	case "-.":
		op = "-"
	}
	return tlc.NewOpApplNode(b.symbol(op), b.convertExpr(e.Expr))
}

func (b *tlcBridge) binaryNode(e *BinaryExpr) tlc.SemanticNode {
	op := tlcBinaryOperator(e.Op)
	return tlc.NewOpApplNode(b.symbol(op), b.convertExpr(e.Left), b.convertExpr(e.Right))
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
	return tlc.NewOpApplNode(b.symbol(callee.Name), args...)
}

func (b *tlcBridge) letNode(e *LetExpr) tlc.SemanticNode {
	lets := make([]*tlc.OpDefNode, 0, len(e.Definitions))
	for _, def := range e.Definitions {
		next := def
		lets = append(lets, b.convertDefinitionAs(next.Name, &next))
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
	node := tlc.NewBuiltinOpApplNode(op, b.convertExpr(e.Body))
	if e.Set != nil {
		node.BdedQuantSymbolLists = [][]*tlc.SymbolNode{{b.symbol(e.Var)}}
		node.BdedQuantBounds = []tlc.SemanticNode{b.convertExpr(e.Set)}
		node.BdedQuantATuple = []bool{e.TupleBound}
	} else {
		node.UnbdedQuantSymbols = []*tlc.SymbolNode{b.symbol(e.Var)}
	}
	return node
}

func (b *tlcBridge) caseNode(e *CaseExpr) tlc.SemanticNode {
	args := make([]tlc.SemanticNode, 0, len(e.Arms)+1)
	for _, arm := range e.Arms {
		args = append(args, tlc.NewBuiltinOpApplNode(tlc.OpPair, b.convertExpr(arm.Test), b.convertExpr(arm.Value)))
	}
	if e.Other != nil {
		args = append(args, tlc.NewBuiltinOpApplNode(tlc.OpPair, nil, b.convertExpr(e.Other)))
	}
	return tlc.NewBuiltinOpApplNode(tlc.OpCase, args...)
}

func (b *tlcBridge) chooseNode(e *ChooseExpr) tlc.SemanticNode {
	op := tlc.OpUC
	if e.Set != nil {
		op = tlc.OpBC
	}
	node := tlc.NewBuiltinOpApplNode(op, b.convertExpr(e.Body))
	if e.Set != nil {
		node.BdedQuantSymbolLists = [][]*tlc.SymbolNode{{b.symbol(e.Var)}}
		node.BdedQuantBounds = []tlc.SemanticNode{b.convertExpr(e.Set)}
		node.BdedQuantATuple = []bool{false}
	} else {
		node.UnbdedQuantSymbols = []*tlc.SymbolNode{b.symbol(e.Var)}
	}
	return node
}

func (b *tlcBridge) functionNode(e *FunctionExpr) tlc.SemanticNode {
	node := tlc.NewBuiltinOpApplNode(tlc.OpFC, b.convertExpr(e.Body))
	for _, bound := range e.Bounds {
		node.BdedQuantSymbolLists = append(node.BdedQuantSymbolLists, []*tlc.SymbolNode{b.symbol(bound.Name)})
		node.BdedQuantBounds = append(node.BdedQuantBounds, b.convertExpr(bound.Set))
		node.BdedQuantATuple = append(node.BdedQuantATuple, bound.TupleBound)
	}
	return node
}

func (b *tlcBridge) exceptNode(e *ExceptExpr) tlc.SemanticNode {
	args := []tlc.SemanticNode{b.convertExpr(e.Base)}
	for _, spec := range e.Specs {
		pathElems := make([]tlc.SemanticNode, 0)
		for _, component := range spec.Components {
			if component.Field != "" {
				pathElems = append(pathElems, tlc.NewStringNode(component.Field))
			}
			for _, index := range component.Indices {
				pathElems = append(pathElems, b.convertExpr(index))
			}
		}
		path := tlc.NewBuiltinOpApplNode(tlc.OpTup, pathElems...)
		args = append(args, tlc.NewBuiltinOpApplNode(tlc.OpPair, path, b.convertExpr(spec.Value)))
	}
	return tlc.NewBuiltinOpApplNode(tlc.OpExc, args...)
}

func (b *tlcBridge) setComprehensionNode(e *SetComprehensionExpr) tlc.SemanticNode {
	if e.Predicate != nil {
		if !setComprehensionElementIsBound(e) {
			b.diags = append(b.diags, errorAt(e.Pos, "E7013", "set comprehension with mapped element and predicate is not yet supported by the TLC bridge"))
		}
		node := tlc.NewBuiltinOpApplNode(tlc.OpSSO, b.convertExpr(e.Predicate))
		for _, bound := range e.Bounds {
			node.BdedQuantSymbolLists = append(node.BdedQuantSymbolLists, []*tlc.SymbolNode{b.symbol(bound.Name)})
			node.BdedQuantBounds = append(node.BdedQuantBounds, b.convertExpr(bound.Set))
			node.BdedQuantATuple = append(node.BdedQuantATuple, bound.TupleBound)
		}
		return node
	}
	body := b.convertExpr(e.Element)
	node := tlc.NewBuiltinOpApplNode(tlc.OpSOA, body)
	for _, bound := range e.Bounds {
		node.BdedQuantSymbolLists = append(node.BdedQuantSymbolLists, []*tlc.SymbolNode{b.symbol(bound.Name)})
		node.BdedQuantBounds = append(node.BdedQuantBounds, b.convertExpr(bound.Set))
		node.BdedQuantATuple = append(node.BdedQuantATuple, bound.TupleBound)
	}
	return node
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
