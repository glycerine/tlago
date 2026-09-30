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
		b.tool.Define(&tlc.SymbolNode{Name: opDef.Name}, opDef)
	}
}

func (b *tlcBridge) installConfigConstants() {
	if b.cfg == nil || b.cfg.GetConstants() == nil {
		return
	}
	for _, constant := range b.cfg.GetConstants().All() {
		if len(constant.Args) != 0 {
			b.diags = append(b.diags, errorAt(Position{}, "E7001", "operator-valued CONSTANT assignment %s is not yet supported by the TLC bridge", constant.Name))
			continue
		}
		if constant.Value == nil {
			continue
		}
		b.tool.DefineName(constant.Name, constant.Value)
	}
	for specName, configName := range b.cfg.GetOverrides().All() {
		def := b.defs[configName]
		if def == nil {
			b.diags = append(b.diags, errorAt(Position{}, "E7002", "CONSTANT override %s <- %s references an unknown operator", specName, configName))
			continue
		}
		opDef := b.convertDefinitionAs(specName, def)
		if opDef != nil {
			b.tool.Define(&tlc.SymbolNode{Name: opDef.Name}, opDef)
		}
	}
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
			b.tool.NextStateSpec = b.actionFromExpr(b.cfg.GetSpec()+"!Next", nextExpr, nil, false)
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
		b.tool.NextStateSpec = b.actionFromDefinition(nextName, false)
	}
	if b.tool.NextStateSpec != nil {
		b.tool.Actions = []*tlc.Action{b.tool.NextStateSpec}
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
	if name := b.cfg.GetView(); name != "" {
		b.tool.ViewSpec = b.nodeForDefinition(name)
	}
	for _, name := range b.cfg.GetPostConditions() {
		if action := b.actionFromDefinition(name, false); action != nil {
			b.tool.PostConditionSpecs = append(b.tool.PostConditionSpecs, action)
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
	params := make([]*tlc.SymbolNode, len(def.Params))
	for i, param := range def.Params {
		params[i] = b.symbol(param)
	}
	body := b.convertExpr(def.Expr)
	if body == nil {
		return nil
	}
	return tlc.NewOpDefNode(name, params, body)
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
	return tlc.NewLetInNode(b.convertExpr(e.Body), lets...)
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
