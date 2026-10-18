package tlc

import (
	"fmt"
	"net/url"
	"reflect"
	"runtime"
	"strings"
	"unicode/utf16"
)

type TLCEvaluateArguments struct {
	Context    string
	Expression *string
	FrameID    *int
}

type TLCEvaluateResponse struct {
	Result             *string
	Type               *string
	VariablesReference int
}

func (d *TLCDebugger) Evaluate(arguments TLCEvaluateArguments) TLCEvaluateResponse {
	d.require()
	d.mu.Lock()
	defer d.mu.Unlock()
	hover := false
	switch arguments.Context {
	case "hover":
		if arguments.Expression == nil {
			panic(NewNullPointerException())
		}
		hover = strings.HasPrefix(*arguments.Expression, "tlaplus://")
		if !hover {
			return TLCEvaluateResponse{}
		}
	case "variables":
		var result *string
		if arguments.Expression != nil {
			result = javaString(*arguments.Expression)
		}
		return TLCEvaluateResponse{Result: result}
	case "repl", "clipboard", "watch":
	default:
		return TLCEvaluateResponse{}
	}
	// Java unboxes frameId while comparing it to each frame's primitive id.
	stack := d.ownedStack()
	for i := len(stack) - 1; i >= 0; i-- {
		if arguments.FrameID == nil {
			panic(NewNullPointerException())
		}
		frame := stack[i]
		if frame.Base.ID == *arguments.FrameID {
			if hover {
				return frame.GetHover(arguments.Expression)
			}
			return frame.Evaluate(arguments.Expression)
		}
	}
	return TLCEvaluateResponse{}
}

func (f *TLCStackFrame) GetHover(expression *string) TLCEvaluateResponse {
	f.require()
	return f.getHover(expression, EmptyState, nil, false, false)
}
func (f *TLCStateStackFrame) GetHover(expression *string) TLCEvaluateResponse {
	f.require()
	return f.TLCStackFrame.getHover(expression, f.GetS(), nil, true, false)
}
func (f *TLCActionStackFrame) GetHover(expression *string) TLCEvaluateResponse {
	f.require()
	return f.TLCStackFrame.getHover(expression, f.GetS(), f.GetT(), true, true)
}
func (f *TLCDebuggerFrame) GetHover(expression *string) TLCEvaluateResponse {
	if f.Action != nil {
		return f.Action.GetHover(expression)
	}
	if f.State != nil {
		return f.State.GetHover(expression)
	}
	return f.Base.GetHover(expression)
}

func (f *TLCStackFrame) getHover(expression *string, state, successor *TLCStateMut, hasState, action bool) (response TLCEvaluateResponse) {
	if expression == nil {
		panic(NewNullPointerException())
	}
	defer func() {
		if failure := recover(); failure != nil {
			if _, ok := failure.(*IllegalArgumentException); ok {
				return
			}
			if _, ok := failure.(*NumberFormatException); ok {
				return
			}
			panic(failure)
		}
	}()
	u, err := url.Parse(*expression)
	if err != nil {
		return response
	}
	if !debugURICharactersValid(*expression) {
		return response
	}
	if u.Opaque != "" {
		panic(NewNullPointerException())
	}
	name := u.Path[strings.LastIndex(u.Path, "/")+1:]
	name = strings.TrimSuffix(name, ".tla")
	if !strings.Contains(*expression, "#") {
		panic(NewNullPointerException())
	}
	coordinates := strings.Split(u.Fragment, " ")
	for u.Fragment != "" && len(coordinates) > 0 && coordinates[len(coordinates)-1] == "" {
		coordinates = coordinates[:len(coordinates)-1]
	}
	// Java evaluates all four constructor arguments before parsing any int.
	if len(coordinates) < 4 {
		panic(NewArrayIndexOutOfBoundsException(len(coordinates), len(coordinates)))
	}
	var values [4]int
	for i := range values {
		value, valid := javaParseDecimalInt(coordinates[i])
		if !valid {
			return response
		}
		values[i] = int(value)
	}
	location := NewSourceLocation(name, values[0], values[1], values[2], values[3])
	module := f.GetTool().GetModule(location.Source)
	if module == nil {
		panic(NewNullPointerException())
	}
	path := module.PathTo(location)
	if len(path) == 0 {
		return TLCEvaluateResponse{Result: javaString(location.String())}
	}
	variable := f.getHoverVariable(path, state, successor, hasState, action)
	if variable == nil {
		return response
	}
	response.Result = javaString(variable.Value)
	if variable.Type != "" {
		response.Type = javaString(variable.Type)
	}
	response.VariablesReference = variable.VariablesReference
	return response
}

func debugPrimeScope(path []SemanticNode) bool {
	for _, node := range path {
		if application, ok := node.(*OpApplNode); ok && application.Operator != nil && application.Operator.Name == OpPrime {
			return true
		}
	}
	return false
}

func (f *TLCStackFrame) getHoverVariable(path []SemanticNode, state, successor *TLCStateMut, hasState, action bool) *DebugTLCVariable {
	first := path[0]
	prime := debugPrimeScope(path)
	if action && prime {
		if variable := f.GetTool().GetPrimedVar(first, f.GetContext(), false); variable != nil {
			if value := successor.Lookup(variable.GetName()); !isNil(value) {
				return f.debugVariableForValue(value, variable.GetName().String()+"'", nil)
			}
			return &DebugTLCVariable{Name: variable.GetName().String() + "'", Value: "?"}
		}
	}
	if hasState {
		if !prime {
			if variable := f.GetTool().GetVar(first, f.GetContext(), false); variable != nil {
				if value := state.Lookup(variable.GetName()); !isNil(value) {
					return f.debugVariableForValue(value, variable.GetName().String(), nil)
				}
				return &DebugTLCVariable{Name: variable.GetName().String(), Value: "?"}
			}
		} else {
			location, _ := semanticNodeSourceLocation(first)
			return &DebugTLCVariable{Name: semanticNodeHoverImage(first), Value: location.String()}
		}
	}
	var object any
	if value, symbol := debugLookupSymbolValue(f.GetTool(), first, f.GetContext()); symbol {
		object = value
		if object == nil {
			object = first
		}
	} else if debugExprOrOpArg(first) {
		expression := first
		for _, node := range path {
			if _, ok := node.(*OpApplNode); ok {
				expression = node
				break
			}
		}
		object = f.GetTool().GetVal(expression, f.GetContext(), false, DoNotRecordCostModel)
	} else {
		object = first
	}
	if lazy := asLazyValue(object); lazy != nil {
		object = first
		tool := f.GetTool()
		value, err := tool.withDebugEvalModeAny(DebugEvalDebugger, func() (any, error) { return debugUnlazy(lazy, tool, state, successor, !hasState) })
		if err == nil {
			object = value
		}
	}
	if value, ok := object.(Value); ok {
		return f.debugVariableForValue(value, semanticNodeHoverImage(first), nil)
	}
	return &DebugTLCVariable{Value: debugSemanticObjectString(object), Type: debugSemanticClassName(object)}
}

// Java's OpDefNode, ThmOrAssumpDefNode and ModuleNode inherit SymbolNode. Go
// keeps their concrete semantic data in sibling structs, so debugger lookup
// has to recover the shared symbol view before applying the source lookup.
func debugLookupSymbolValue(tool *Tool, node SemanticNode, context *Context) (any, bool) {
	switch node := node.(type) {
	case *SymbolNode:
		return tool.LookupSymbolValue(node, context, false), true
	case *OpDefNode:
		if node == nil || node.Symbol == nil {
			panic(NewNullPointerException())
		}
		return tool.LookupSymbolValue(node.Symbol, context, false), true
	case *ThmOrAssumpDefNode:
		if node == nil || node.Symbol == nil {
			panic(NewNullPointerException())
		}
		return tool.LookupSymbolValue(node.Symbol, context, false), true
	case *ModuleNode:
		if tool == nil || node == nil || context == nil {
			panic(NewNullPointerException())
		}
		if value := SemanticToolObjectForTool(tool, node); !isNil(value) {
			return muxToolObject(value), true
		}
		return node, true
	default:
		return nil, false
	}
}

// TLCStackFrame.getVariable renders built-in symbols by location even when
// SemanticNode.showPlainFormulae is set. Other semantic objects use their
// ordinary toString representation, which does honor that property.
func debugSemanticObjectString(object any) string {
	if builtIn, ok := object.(interface{ IsBuiltIn() bool }); ok && builtIn.IsBuiltIn() {
		location, _ := semanticNodeSourceLocation(object)
		return location.String()
	}
	if _, ok := semanticNodeSourceLocation(object); ok {
		return semanticNodeJavaString(object)
	}
	return fmt.Sprint(object)
}

func debugExprOrOpArg(node SemanticNode) bool {
	switch node.(type) {
	case *OpApplNode, *LetInNode, *SubstInNode, *NumeralNode, *DecimalNode, *StringNode, *AtNode, *OpArgNode, *LabelNode:
		return true
	}
	return false
}

func debugSemanticClassName(node SemanticNode) string {
	if symbol, ok := node.(*SymbolNode); ok {
		switch symbol.Kind {
		case SymbolFormalParam:
			return "FormalParamNode"
		case SymbolVariableDecl, SymbolConstantDecl:
			return "OpDeclNode"
		case SymbolUserDefinedOp, SymbolBuiltIn:
			return "OpDefNode"
		}
	}
	return reflect.TypeOf(node).Elem().Name()
}

func syntaxNodeHumanReadableImage(node SemanticNode) string {
	if node, ok := node.(interface{ GetTreeNode() any }); ok {
		if tree, ok := node.GetTreeNode().(interface{ GetHumanReadableImage() string }); ok {
			return tree.GetHumanReadableImage()
		}
	}
	return ""
}

// Expression is nullable in the source protocol. Keeping the pointer preserves
// Java's distinction between a null request and an empty expression.
func (f *TLCStackFrame) Evaluate(expression *string) TLCEvaluateResponse {
	f.require()
	return f.evaluateExpression(expression, func() *Context { return f.GetContext() }, EmptyState, EmptyState, true)
}

func (f *TLCStateStackFrame) Evaluate(expression *string) TLCEvaluateResponse {
	f.require()
	return f.TLCStackFrame.evaluateExpression(expression, func() *Context { return f.getEvaluateContext(f.GetS()) }, f.GetS(), f.GetS(), false)
}

func (f *TLCActionStackFrame) Evaluate(expression *string) TLCEvaluateResponse {
	f.require()
	return f.TLCStackFrame.evaluateExpression(expression, func() *Context { return f.getEvaluateContext(f.GetS()) }, f.GetS(), f.GetT(), false)
}

func (f *TLCSyntheticStateStackFrame) Evaluate(expression *string) TLCEvaluateResponse {
	f.require()
	return f.TLCStackFrame.evaluateExpression(expression, func() *Context { return f.getEvaluateContext(f.GetS()) }, f.GetT(), f.ownedSuccessor(), false)
}

func (f *TLCDebuggerFrame) Evaluate(expression *string) TLCEvaluateResponse {
	switch {
	case f.Synthetic != nil:
		return f.Synthetic.Evaluate(expression)
	case f.Action != nil:
		return f.Action.Evaluate(expression)
	case f.State != nil:
		return f.State.Evaluate(expression)
	default:
		return f.Base.Evaluate(expression)
	}
}

func (f *TLCStackFrame) evaluateExpression(expression *string, getContext func() *Context, s, t *TLCStateMut, base bool) (response TLCEvaluateResponse) {
	if expression == nil {
		return response
	}
	defer func() {
		if failure := recover(); failure != nil {
			if err, ok := failure.(error); ok && debugEvaluationException(err) {
				response = f.evaluationError(err, *expression)
				return
			}
			panic(failure)
		}
	}()
	location, _ := semanticNodeSourceLocation(f.GetNode())
	op, err := f.GetTool().ParseDebuggerExpression(location, *expression)
	if err != nil {
		return f.evaluationError(err, *expression)
	}
	if op == nil {
		return TLCEvaluateResponse{Result: javaString(*expression)}
	}
	ctxt := getContext()
	for _, parameter := range op.GetParams() {
		value := f.GetContext().LookupFunc(func(symbol *SymbolNode) bool {
			return symbol != nil && symbol.GetName() == parameter.GetName()
		})
		ctxt = ctxt.Cons(parameter, value)
	}
	variable, err := f.evaluateVariable(op.Name.String(), func() (Value, error) {
		if base {
			return f.GetTool().NoDebug().Eval(op.GetBody(), ctxt)
		}
		return f.GetTool().NoDebug().Eval(op.GetBody(), ctxt, s, t, EvalClear)
	})
	if err != nil {
		if !debugEvaluationException(err) {
			panic(err)
		}
		return f.evaluationError(err, *expression)
	}
	return debugVariableEvaluationResponse(variable)
}

func (f *TLCStackFrame) evaluationError(err error, name string) TLCEvaluateResponse {
	message := ""
	if detail := javaThrowableDetailMessage(err); detail != nil {
		message = *detail
	}
	return debugVariableEvaluationResponse(f.debugVariableForValue(NewStringValue(message), name, nil))
}

func debugVariableEvaluationResponse(variable *DebugTLCVariable) TLCEvaluateResponse {
	// Java evaluate/getWatch do not populate the response's type field.
	return TLCEvaluateResponse{Result: javaString(variable.Value), VariablesReference: variable.VariablesReference}
}

func debugEvaluationException(err error) bool {
	_, runtimeFailure := err.(runtime.Error)
	return err != nil && !runtimeFailure && !isJavaError(err)
}

func (f *TLCStackFrame) evaluateVariable(name string, evaluate func() (Value, error)) (*DebugTLCVariable, error) {
	variable, err := f.GetTool().withDebugEvalModeAny(DebugEvalDebugger, func() (any, error) {
		value, err := evaluate()
		if err != nil {
			return nil, err
		}
		return f.debugVariableForValue(value, name, nil), nil
	})
	if err != nil {
		return nil, err
	}
	return variable.(*DebugTLCVariable), nil
}

func (f *TLCStackFrame) GetWatch(op *OpDefNode) TLCEvaluateResponse {
	f.require()
	return f.getWatch(op, EmptyState, EmptyState)
}

func (f *TLCStateStackFrame) GetWatch(op *OpDefNode) TLCEvaluateResponse {
	f.require()
	return f.getWatch(op, f.GetS(), f.GetT())
}

func (f *TLCActionStackFrame) GetWatch(op *OpDefNode) TLCEvaluateResponse {
	f.require()
	return f.getWatch(op, f.GetS(), f.GetT())
}

func (f *TLCDebuggerFrame) GetWatch(op *OpDefNode) TLCEvaluateResponse {
	if f.Action != nil {
		return f.Action.GetWatch(op)
	}
	if f.State != nil {
		return f.State.GetWatch(op)
	}
	return f.Base.GetWatch(op)
}

func (f *TLCStackFrame) getWatch(op *OpDefNode, s, t *TLCStateMut) (response TLCEvaluateResponse) {
	if op == nil {
		return response
	}
	defer func() {
		if failure := recover(); failure != nil {
			if err, ok := failure.(error); ok && debugValueFailure(err) {
				response = f.evaluationError(err, op.Name.String())
				return
			}
			panic(failure)
		}
	}()
	variable, err := f.evaluateVariable(op.Name.String(), func() (Value, error) {
		return f.GetTool().Eval(op.GetBody(), f.GetContext(), s, t, EvalClear)
	})
	if err != nil {
		if !debugValueFailure(err) {
			panic(err)
		}
		return f.evaluationError(err, op.Name.String())
	}
	return debugVariableEvaluationResponse(variable)
}

func (f *TLCStateStackFrame) getEvaluateContext(state *TLCStateMut) *Context {
	module := f.GetTool().GetModule("TLCExt")
	if module == nil {
		return f.GetContext()
	}
	counterexample := NewEmptyCounterExample()
	if simulator := CurrentSimulator(); simulator != nil {
		counterexample = NewCounterExampleFromStateVec(debuggerVariableSimulationTrace(simulator.GetUncompressedTrace(state)))
	}
	return f.GetContext().Cons(module.GetOpDef(UniqueStringOf("CounterExample")).Symbol, counterexample)
}

// URI.create checks distinct path, query, and fragment character classes.
// Reuse the java.net.URI scanner also used by TLCWorker, preserving UTF-16
// indexes and escape validation rather than Go URL's permissive raw spaces.
func debugURICharactersValid(raw string) bool {
	parser := workerURIParser{raw: raw, chars: utf16.Encode([]rune(raw))}
	end := len(parser.chars)
	pos := 0
	colon := parser.until(0, end, ":/?#")
	if parser.at(colon, end, ':') {
		pos = colon + 1
		if !parser.at(pos, end, '/') {
			fragment := parser.until(pos, end, "#")
			if fragment == pos || parser.check(pos, fragment, workerURIURIC, "opaque part") != nil {
				return false
			}
			return fragment == end || parser.check(fragment+1, end, workerURIURIC, "fragment") == nil
		}
	}
	if parser.at(pos, end, '/') && parser.at(pos+1, end, '/') {
		pos += 2
		authorityEnd := parser.until(pos, end, "/?#")
		if authorityEnd > pos {
			if _, err := parser.authority(pos, authorityEnd); err != nil {
				return false
			}
		} else if authorityEnd == end {
			return false
		}
		pos = authorityEnd
	}
	pathEnd := parser.until(pos, end, "?#")
	if parser.check(pos, pathEnd, workerURIPath, "path") != nil {
		return false
	}
	pos = pathEnd
	if parser.at(pos, end, '?') {
		queryEnd := parser.until(pos+1, end, "#")
		if parser.check(pos+1, queryEnd, workerURIURIC, "query") != nil {
			return false
		}
		pos = queryEnd
	}
	return !parser.at(pos, end, '#') || parser.check(pos+1, end, workerURIURIC, "fragment") == nil
}

func semanticNodeHoverImage(node SemanticNode) string {
	if symbol, ok := node.(*SymbolNode); ok {
		switch symbol.Kind {
		case SymbolConstantDecl:
			return symbol.GetName().String() + " CONSTANT"
		case SymbolVariableDecl:
			return symbol.GetName().String() + " VARIABLE"
		}
	}
	location, _ := semanticNodeSourceLocation(node)
	return location.String()
}
