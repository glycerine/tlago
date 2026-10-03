/*******************************************************************************
 * Copyright (c) 2020 Microsoft Research. All rights reserved.
 * Copyright (c) 2025 NVIDIA Corp. All rights reserved.
 *
 * The MIT License (MIT)
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
 * of the Software, and to permit persons to whom the Software is furnished to do
 * so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN
 * AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
 * WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 *
 * Contributors:
 *   Markus Alexander Kuppe - initial API and implementation
 ******************************************************************************/
// Port of tlc2.debug.TLCStackFrame and its concrete evaluation overrides.
package tlc

import "runtime"

type TLCEvaluateResponse struct {
	Result             *string
	Type               *string
	VariablesReference int
}

func semanticNodeHumanReadableImage(node SemanticNode) string {
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
	return f.evaluateExpression(expression, func() *Context { return f.Context }, EmptyState, EmptyState, true)
}

func (f *TLCStateStackFrame) Evaluate(expression *string) TLCEvaluateResponse {
	return f.TLCStackFrame.evaluateExpression(expression, func() *Context { return f.getEvaluateContext(f.GetS()) }, f.GetS(), f.GetS(), false)
}

func (f *TLCActionStackFrame) Evaluate(expression *string) TLCEvaluateResponse {
	return f.TLCStackFrame.evaluateExpression(expression, func() *Context { return f.getEvaluateContext(f.GetS()) }, f.GetS(), f.GetT(), false)
}

func (f *TLCSyntheticStateStackFrame) Evaluate(expression *string) TLCEvaluateResponse {
	return f.TLCStackFrame.evaluateExpression(expression, func() *Context { return f.getEvaluateContext(f.GetS()) }, f.State, f.Successor, false)
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
	location, _ := semanticNodeSourceLocation(f.Node)
	op, err := f.Tool.ParseDebuggerExpression(location, *expression)
	if err != nil {
		return f.evaluationError(err, *expression)
	}
	if op == nil {
		return TLCEvaluateResponse{Result: javaString(*expression)}
	}
	ctxt := getContext()
	for _, parameter := range op.Params {
		value := f.Context.LookupFunc(func(symbol *SymbolNode) bool {
			return symbol != nil && symbol.GetName() == parameter.GetName()
		})
		ctxt = ctxt.Cons(parameter, value)
	}
	variable, err := f.evaluateVariable(op.Name.String(), func() (Value, error) {
		if base {
			return f.Tool.NoDebug().Eval(op.Body, ctxt)
		}
		return f.Tool.NoDebug().Eval(op.Body, ctxt, s, t, EvalClear)
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
	variable, err := f.Tool.withDebugEvalModeAny(DebugEvalDebugger, func() (any, error) {
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
	return f.getWatch(op, EmptyState, EmptyState)
}

func (f *TLCStateStackFrame) GetWatch(op *OpDefNode) TLCEvaluateResponse {
	return f.getWatch(op, f.GetS(), f.GetT())
}

func (f *TLCActionStackFrame) GetWatch(op *OpDefNode) TLCEvaluateResponse {
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
		return f.Tool.Eval(op.Body, f.Context, s, t, EvalClear)
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
	module := f.Tool.GetModule("TLCExt")
	if module == nil {
		return f.Context
	}
	counterexample := NewEmptyCounterExample()
	if simulator := CurrentSimulator(); simulator != nil {
		counterexample = NewCounterExampleFromStateVec(simulator.GetUncompressedTrace(state))
	}
	return f.Context.Cons(module.GetOpDef(UniqueStringOf("CounterExample")).Symbol, counterexample)
}
