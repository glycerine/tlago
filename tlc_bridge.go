package tlago

import (
	"sort"
	"strconv"
	"strings"

	"github.com/glycerine/tlago/tlc"
)

type tlcBridge struct {
	tool    *tlc.Tool
	spec    *Spec
	cfg     *tlc.ModelConfig
	defs    map[string]*Definition
	diags   Diagnostics
	symbols map[string]*tlc.SymbolNode
}

var bridgeStandardModuleMembers = map[string][]string{
	"Naturals": {
		"Nat", "+", "-", "*", "^", "<", ">", "\\leq", "\\geq", "%", "\\div", "..",
	},
	"Integers": {
		"Int", "-.",
	},
	"Sequences": {
		"Seq", "BSeq", "Len", "Head", "Tail", "Cons", "Append", "Concat", "\\o",
		"SubSeq", "SelectInSeq", "SelectSeq",
	},
	"FiniteSets": {
		"IsFiniteSet", "Cardinality",
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
		"RandomSubset", "RandomSetOfSubsets", "RandomSubsetSet",
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

// BuildTLCTool converts the production Go SANY semantic tree into the TLC
// runtime tree used by the mechanical TLC port.
func BuildTLCTool(spec *Spec, cfg *tlc.ModelConfig) (*tlc.Tool, Diagnostics) {
	if spec == nil || spec.Root == nil {
		return nil, Diagnostics{errorAt(Position{}, "E7000", "missing root module for TLC tool")}
	}
	if cfg == nil {
		cfg = tlc.NewModelConfig(spec.Root.Name)
	}
	bridge := &tlcBridge{
		tool:    tlc.NewToolWithModelConfig(cfg),
		spec:    spec,
		cfg:     cfg,
		defs:    definitionsByName(spec),
		symbols: map[string]*tlc.SymbolNode{},
	}
	bridge.tool.RootName = spec.Root.Name
	bridge.tool.RootFile = spec.Root.SourcePath
	bridge.tool.SpecDir = ""
	bridge.tool.ConfigFile = spec.Root.Name
	bridge.installVariables()
	bridge.installDefinitions()
	bridge.installConfigConstants()
	bridge.installInstanceAliases()
	bridge.installAssumptions()
	bridge.installModelTargets()
	bridge.tool.AssignActionIDs()
	return bridge.tool, bridge.diags
}

func (b *tlcBridge) installVariables() {
	vars := moduleVariables(b.spec.Root)
	tlc.SetStateVariables(vars)
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
		opDef := b.convertDefinitionAs(name, def)
		if opDef == nil {
			continue
		}
		b.tool.Define(opDef.Symbol, opDef)
	}
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
				b.tool.DefineName(name, opVal)
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
		b.tool.DefineName(name, constant.Value)
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
			b.tool.Define(opDef.Symbol, opDef)
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
				b.tool.Define(binding.Symbol, binding.Value)
			}
		}
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
		value := b.tool.DefnsByName[tlc.UniqueStringOf(member)]
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
	if b.cfg == nil {
		return
	}
	initName := b.cfg.GetInit()
	nextName := b.cfg.GetNext()
	if (initName == "" || nextName == "") && b.cfg.GetSpec() != "" {
		def := b.defs[b.cfg.GetSpec()]
		if def == nil {
			b.diags = append(b.diags, errorAt(Position{}, "E7003", "SPECIFICATION operator %s not found", b.cfg.GetSpec()))
		} else if initExpr, nextExpr, ok := decomposeTemporalSpecification(def.Expr); ok {
			b.tool.InitStateSpec = append(b.tool.InitStateSpec, b.actionFromExpr(b.cfg.GetSpec()+"!Init", initExpr, nil, true))
			b.tool.SetNextStateSpec(b.actionFromExpr(b.cfg.GetSpec()+"!Next", nextExpr, nil, false))
		} else {
			b.diags = append(b.diags, errorAt(def.Pos, "E7004", "SPECIFICATION %s must have form Init /\\ [][Next]_vars for TLC bridge decomposition", b.cfg.GetSpec()))
		}
	}
	if len(b.tool.InitStateSpec) == 0 && initName != "" {
		if action := b.actionFromDefinition(initName, true); action != nil {
			b.tool.InitStateSpec = append(b.tool.InitStateSpec, action)
		}
	}
	if b.tool.NextStateSpec == nil && nextName != "" {
		b.tool.SetNextStateSpec(b.actionFromDefinition(nextName, false))
	}
	for _, name := range b.cfg.GetInvariants() {
		if action := b.actionFromDefinition(name, false); action != nil {
			b.tool.Invariants = append(b.tool.Invariants, action)
			b.tool.InvariantNames = append(b.tool.InvariantNames, name)
		}
	}
	for _, name := range b.cfg.GetProperties() {
		if action := b.actionFromDefinition(name, false); action != nil {
			b.tool.Temporals = append(b.tool.Temporals, action)
			b.tool.TemporalNames = append(b.tool.TemporalNames, name)
		}
	}
	for _, name := range b.cfg.GetConstraints() {
		if node := b.nodeForDefinition(name); node != nil {
			b.tool.ModelConstraints = append(b.tool.ModelConstraints, node)
		}
	}
	for _, name := range b.cfg.GetActionConstraints() {
		if node := b.nodeForDefinition(name); node != nil {
			b.tool.ActionConstraints = append(b.tool.ActionConstraints, node)
		}
	}
	b.installPossibleTargets()
	if name := b.cfg.GetView(); name != "" {
		b.tool.ViewSpec = b.nodeForDefinition(name)
	}
	if name := b.cfg.GetAlias(); name != "" {
		b.installAliasTarget(name)
	}
	if name := b.cfg.GetPeriodic(); name != "" {
		b.tool.Periodic = b.nodeForDefinition(name)
	}
	if name := b.cfg.GetRLReward(); name != "" {
		b.tool.RLReward = b.nodeForDefinition(name)
	}
	for _, name := range b.cfg.GetPostConditions() {
		if action := b.actionFromDefinition(name, false); action != nil {
			b.tool.PostConditionSpecs = append(b.tool.PostConditionSpecs, action)
		}
	}
}

func (b *tlcBridge) installPossibleTargets() {
	for _, name := range b.cfg.GetPossible() {
		def := b.defs[name]
		if def == nil {
			b.diags = append(b.diags, errorAt(Position{}, "E7014", "_POSSIBLE operator %s not found", name))
			continue
		}
		if len(def.Params) != 0 {
			b.diags = append(b.diags, errorAt(def.Pos, "E7015", "_POSSIBLE operator %s must be zero-arity", name))
			continue
		}
		opDef := b.convertDefinitionAs(name, def)
		if opDef == nil || opDef.Symbol == nil {
			continue
		}
		track := tlc.NewPossibleTrackNode(tlc.NewOpApplNode(opDef.Symbol), name)
		if b.possibleExprIsActionLevel(def.Expr, map[string]bool{}, nil) {
			b.tool.ActionConstraints = append(b.tool.ActionConstraints, track)
		} else {
			b.tool.ModelConstraints = append(b.tool.ModelConstraints, track)
		}
		check := tlc.NewPossibleCheckNode(name)
		b.tool.PostConditionSpecs = append(b.tool.PostConditionSpecs, tlc.NewPossibleAction(check, tlc.EmptyContext, opDef))
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
	return b.convertExpr(def.Expr)
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

func (b *tlcBridge) possibleExprIsActionLevel(expr Expr, seen map[string]bool, locals map[string]*Definition) bool {
	return b.possibleExprIsActionLevelBound(expr, seen, locals, nil)
}

func (b *tlcBridge) possibleExprIsActionLevelBound(expr Expr, seen map[string]bool, locals map[string]*Definition, bound map[string]bool) bool {
	switch e := expr.(type) {
	case nil, *LiteralExpr:
		return false
	case *IdentExpr:
		if bound != nil && bound[e.Name] {
			return false
		}
		return b.possibleNamedExprIsActionLevel(e.Name, seen, locals, bound)
	case *UnaryExpr:
		switch e.Op {
		case "'", "UNCHANGED":
			return true
		case "ENABLED":
			return false
		case "[]", "<>":
			return true
		default:
			return b.possibleExprIsActionLevelBound(e.Expr, seen, locals, bound)
		}
	case *BinaryExpr:
		return b.possibleExprIsActionLevelBound(e.Left, seen, locals, bound) ||
			b.possibleExprIsActionLevelBound(e.Right, seen, locals, bound)
	case *CallExpr:
		if ident, ok := e.Callee.(*IdentExpr); ok && (bound == nil || !bound[ident.Name]) {
			if b.possibleNamedExprIsActionLevel(ident.Name, seen, locals, bound) {
				return true
			}
		} else if b.possibleExprIsActionLevelBound(e.Callee, seen, locals, bound) {
			return true
		}
		for _, arg := range e.Args {
			if b.possibleExprIsActionLevelBound(arg, seen, locals, bound) {
				return true
			}
		}
	case *IfExpr:
		return b.possibleExprIsActionLevelBound(e.Cond, seen, locals, bound) ||
			b.possibleExprIsActionLevelBound(e.Then, seen, locals, bound) ||
			b.possibleExprIsActionLevelBound(e.Else, seen, locals, bound)
	case *LetExpr:
		letLocals := b.possibleLetLocals(e, locals)
		return b.possibleExprIsActionLevelBound(e.Body, seen, letLocals, bound)
	case *QuantifierExpr:
		nextBound := copyBoolMap(bound)
		nextBound[e.Var] = true
		return b.possibleExprIsActionLevelBound(e.Set, seen, locals, bound) ||
			b.possibleExprIsActionLevelBound(e.Body, seen, locals, nextBound)
	case *CaseExpr:
		for _, arm := range e.Arms {
			if b.possibleExprIsActionLevelBound(arm.Test, seen, locals, bound) ||
				b.possibleExprIsActionLevelBound(arm.Value, seen, locals, bound) {
				return true
			}
		}
		return b.possibleExprIsActionLevelBound(e.Other, seen, locals, bound)
	case *ChooseExpr:
		nextBound := copyBoolMap(bound)
		nextBound[e.Var] = true
		return b.possibleExprIsActionLevelBound(e.Set, seen, locals, bound) ||
			b.possibleExprIsActionLevelBound(e.Body, seen, locals, nextBound)
	case *TupleExpr:
		for _, elem := range e.Elems {
			if b.possibleExprIsActionLevelBound(elem, seen, locals, bound) {
				return true
			}
		}
	case *SetExpr:
		for _, elem := range e.Elems {
			if b.possibleExprIsActionLevelBound(elem, seen, locals, bound) {
				return true
			}
		}
	case *RecordExpr:
		for _, field := range e.Fields {
			if b.possibleExprIsActionLevelBound(field.Value, seen, locals, bound) {
				return true
			}
		}
	case *RecordComponentExpr:
		return b.possibleExprIsActionLevelBound(e.Record, seen, locals, bound)
	case *RecordSetExpr:
		for _, field := range e.Fields {
			if b.possibleExprIsActionLevelBound(field.Set, seen, locals, bound) {
				return true
			}
		}
	case *FunctionExpr:
		nextBound := copyBoolMap(bound)
		for _, next := range e.Bounds {
			if b.possibleExprIsActionLevelBound(next.Set, seen, locals, bound) {
				return true
			}
			nextBound[next.Name] = true
		}
		return b.possibleExprIsActionLevelBound(e.Body, seen, locals, nextBound)
	case *FunctionAppExpr:
		if b.possibleExprIsActionLevelBound(e.Function, seen, locals, bound) {
			return true
		}
		for _, arg := range e.Args {
			if b.possibleExprIsActionLevelBound(arg, seen, locals, bound) {
				return true
			}
		}
	case *ExceptExpr:
		if b.possibleExprIsActionLevelBound(e.Base, seen, locals, bound) {
			return true
		}
		for _, spec := range e.Specs {
			for _, component := range spec.Components {
				for _, index := range component.Indices {
					if b.possibleExprIsActionLevelBound(index, seen, locals, bound) {
						return true
					}
				}
			}
			if b.possibleExprIsActionLevelBound(spec.Value, seen, locals, bound) {
				return true
			}
		}
	case *LabelExpr:
		return b.possibleExprIsActionLevelBound(e.Body, seen, locals, bound)
	case *ActionExpr, *FairnessExpr:
		return true
	case *FunctionSetExpr:
		return b.possibleExprIsActionLevelBound(e.Domain, seen, locals, bound) ||
			b.possibleExprIsActionLevelBound(e.Range, seen, locals, bound)
	case *SetComprehensionExpr:
		nextBound := copyBoolMap(bound)
		for _, next := range e.Bounds {
			if b.possibleExprIsActionLevelBound(next.Set, seen, locals, bound) {
				return true
			}
			nextBound[next.Name] = true
		}
		return b.possibleExprIsActionLevelBound(e.Element, seen, locals, nextBound) ||
			b.possibleExprIsActionLevelBound(e.Predicate, seen, locals, nextBound)
	}
	return false
}

func (b *tlcBridge) possibleNamedExprIsActionLevel(name string, seen map[string]bool, locals map[string]*Definition, bound map[string]bool) bool {
	if seen[name] {
		return false
	}
	var def *Definition
	if locals != nil {
		def = locals[name]
	}
	if def == nil {
		def = b.defs[name]
	}
	if def == nil {
		return false
	}
	seen[name] = true
	defer delete(seen, name)
	nextBound := copyBoolMap(bound)
	for _, param := range def.Params {
		nextBound[param] = true
	}
	return b.possibleExprIsActionLevelBound(def.Expr, seen, locals, nextBound)
}

func (b *tlcBridge) possibleLetLocals(expr *LetExpr, parent map[string]*Definition) map[string]*Definition {
	locals := make(map[string]*Definition, len(parent)+len(expr.Definitions))
	for name, def := range parent {
		locals[name] = def
	}
	for i := range expr.Definitions {
		def := &expr.Definitions[i]
		locals[def.Name] = def
	}
	return locals
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
