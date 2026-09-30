package tlc

func (t *Tool) InstallStandardDefinitions() *Tool {
	if t == nil {
		return nil
	}

	t.defineStandardValue("Nat", Nat())
	t.defineStandardValue("Int", Int())
	t.defineStandardValue("STRING", STRING())
	t.defineStandardValue("Any", Any(), "ANY")

	t.defineStandardMethod("Plus", 2, standardBinaryInt("Plus", NatPlus), "+")
	t.defineStandardMethod("Minus", 2, standardBinaryInt("Minus", NatMinus), "-")
	t.defineStandardMethod("Times", 2, standardBinaryInt("Times", NatTimes), "*")
	t.defineStandardMethod("LT", 2, func(args []Value) (Value, error) { return NatLT(args[0], args[1]) }, "<")
	t.defineStandardMethod("LE", 2, func(args []Value) (Value, error) { return NatLE(args[0], args[1]) }, "\\leq")
	t.defineStandardMethod("GT", 2, func(args []Value) (Value, error) { return NatGT(args[0], args[1]) }, ">")
	t.defineStandardMethod("GEQ", 2, func(args []Value) (Value, error) { return NatGEQ(args[0], args[1]) }, "\\geq")
	t.defineStandardMethod("DotDot", 2, func(args []Value) (Value, error) {
		x, err := standardIntArg("DotDot", args, 0)
		if err != nil {
			return nil, err
		}
		y, err := standardIntArg("DotDot", args, 1)
		if err != nil {
			return nil, err
		}
		return DotDot(x, y), nil
	}, "..")
	t.defineStandardMethod("Neg", 1, func(args []Value) (Value, error) {
		x, err := standardIntArg("Neg", args, 0)
		if err != nil {
			return nil, err
		}
		return IntNeg(x)
	}, "-.")
	t.defineStandardMethod("Divide", 2, standardBinaryInt("Divide", IntDivide), "\\div")
	t.defineStandardMethod("Mod", 2, standardBinaryInt("Mod", NatMod), "%")
	t.defineStandardMethod("Expt", 2, standardBinaryInt("Expt", NatExpt), "^")

	t.defineStandardMethod("IsFiniteSet", 1, func(args []Value) (Value, error) { return IsFiniteSet(args[0]) })
	t.defineStandardMethod("Cardinality", 1, func(args []Value) (Value, error) { return Cardinality(args[0]) })

	t.defineStandardMethod("Seq", 1, func(args []Value) (Value, error) { return Seq(args[0]), nil })
	t.defineStandardMethod("BSeq", 2, func(args []Value) (Value, error) {
		bound, err := standardIntArg("BSeq", args, 1)
		if err != nil {
			return nil, err
		}
		return BSeq(args[0], int(bound.Val)), nil
	})
	t.defineStandardMethod("Len", 1, func(args []Value) (Value, error) { return Len(args[0]) })
	t.defineStandardMethod("Head", 1, func(args []Value) (Value, error) { return Head(args[0]) })
	t.defineStandardMethod("Tail", 1, func(args []Value) (Value, error) { return Tail(args[0]) })
	t.defineStandardMethod("Cons", 2, func(args []Value) (Value, error) { return Cons(args[0], args[1]) })
	t.defineStandardMethod("Append", 2, func(args []Value) (Value, error) { return Append(args[0], args[1]) })
	t.defineStandardMethod("Concat", 2, func(args []Value) (Value, error) { return Concat(args[0], args[1]) }, "\\o")
	t.defineStandardMethod("SubSeq", 3, func(args []Value) (Value, error) { return SubSeq(args[0], args[1], args[2]) })
	t.defineStandardMethod("SelectInSeq", 2, func(args []Value) (Value, error) { return SelectInSeq(args[0], args[1]) })
	t.defineStandardMethod("SelectSeq", 2, func(args []Value) (Value, error) { return SelectSeq(args[0], args[1]) })

	t.defineStandardMethod("EmptyBag", 0, func(args []Value) (Value, error) { return EmptyBag(), nil })
	t.defineStandardMethod("IsABag", 1, func(args []Value) (Value, error) { return IsABag(args[0]) })
	t.defineStandardMethod("BagCardinality", 1, func(args []Value) (Value, error) { return BagCardinality(args[0]) })
	t.defineStandardMethod("BagIn", 2, func(args []Value) (Value, error) { return BagIn(args[0], args[1]) })
	t.defineStandardMethod("CopiesIn", 2, func(args []Value) (Value, error) { return CopiesIn(args[0], args[1]) })
	t.defineStandardMethod("BagCup", 2, func(args []Value) (Value, error) { return BagCup(args[0], args[1]) }, "\\oplus")
	t.defineStandardMethod("BagDiff", 2, func(args []Value) (Value, error) { return BagDiff(args[0], args[1]) }, "\\ominus")
	t.defineStandardMethod("BagUnion", 1, func(args []Value) (Value, error) { return BagUnion(args[0]) })
	t.defineStandardMethod("SqSubseteq", 2, func(args []Value) (Value, error) { return SqSubseteq(args[0], args[1]) }, "\\sqsubseteq")
	t.defineStandardMethod("BagOfAll", 2, func(args []Value) (Value, error) { return BagOfAll(args[0], args[1]) })
	t.defineStandardMethod("BagToSet", 1, func(args []Value) (Value, error) { return BagToSet(args[0]) })
	t.defineStandardMethod("SetToBag", 1, func(args []Value) (Value, error) { return SetToBag(args[0]) })

	t.defineStandardMethod("Print", 2, func(args []Value) (Value, error) { return TLCPrint(args[0], args[1]), nil })
	t.defineStandardMethod("PrintT", 1, func(args []Value) (Value, error) { return TLCPrintT(args[0]), nil })
	t.defineStandardMethod("Assert", 2, func(args []Value) (Value, error) { return TLCAssert(args[0], args[1]) })
	t.defineStandardMethod("JavaTime", 0, func(args []Value) (Value, error) { return JavaTime(), nil })
	t.defineStandardMethod("MakeFcn", 2, func(args []Value) (Value, error) { return MakeFcn(args[0], args[1]), nil }, ":>")
	t.defineStandardMethod("CombineFcn", 2, func(args []Value) (Value, error) { return CombineFcn(args[0], args[1]) }, "@@")
	t.defineStandardMethod("Permutations", 1, func(args []Value) (Value, error) { return Permutations(args[0]) })
	t.defineStandardMethod("SortSeq", 2, func(args []Value) (Value, error) { return SortSeq(args[0], args[1]) })
	t.defineStandardMethod("RandomElement", 1, func(args []Value) (Value, error) { return RandomElement(args[0]) })
	t.defineStandardMethod("ToString", 1, func(args []Value) (Value, error) { return TLCToString(args[0]), nil })
	t.defineStandardMethod("TLCEval", 1, func(args []Value) (Value, error) { return TLCEval(args[0]), nil })
	t.defineStandardEvaluating("TLCGet", 1, standardTLCGet)
	t.defineStandardEvaluating("TLCSet", 2, standardTLCSet)

	t.defineStandardMethod("RandomSubset", 2, func(args []Value) (Value, error) { return RandomSubset(args[0], args[1]) })
	t.defineStandardMethod("RandomSetOfSubsets", 3, func(args []Value) (Value, error) {
		return RandomSetOfSubsets(args[0], args[1], args[2])
	})
	t.defineStandardMethod("RandomSubsetSet", 3, func(args []Value) (Value, error) {
		return RandomSubsetSet(args[0], args[1], args[2])
	})
	t.defineStandardMethod("Warshall", 1, func(args []Value) (Value, error) { return Warshall(args[0]) })

	t.defineStandardMethod("ToJson", 1, func(args []Value) (Value, error) { return JsonToJson(args[0]) })
	t.defineStandardMethod("ToJsonArray", 1, func(args []Value) (Value, error) { return JsonToJsonArray(args[0]) })
	t.defineStandardMethod("ToJsonObject", 1, func(args []Value) (Value, error) { return JsonToJsonObject(args[0]) })
	t.defineStandardMethod("JsonDeserialize", 1, func(args []Value) (Value, error) {
		path, err := standardStringArg("JsonDeserialize", args, 0)
		if err != nil {
			return nil, err
		}
		return JsonDeserialize(path)
	})
	t.defineStandardMethod("ndJsonDeserialize", 1, func(args []Value) (Value, error) {
		path, err := standardStringArg("ndJsonDeserialize", args, 0)
		if err != nil {
			return nil, err
		}
		return NDJsonDeserialize(path)
	})
	t.defineStandardMethod("JsonSerialize", 2, func(args []Value) (Value, error) {
		path, err := standardStringArg("JsonSerialize", args, 0)
		if err != nil {
			return nil, err
		}
		return JsonSerialize(path, args[1])
	})
	t.defineStandardMethod("ndJsonSerialize", 2, func(args []Value) (Value, error) {
		path, err := standardStringArg("ndJsonSerialize", args, 0)
		if err != nil {
			return nil, err
		}
		return NDJsonSerialize(path, args[1])
	})

	t.defineStandardEvaluating("AssertError", 2, standardAssertError)
	t.defineStandardEvaluating("PickSuccessor", 1, standardPickSuccessor)
	t.defineStandardMethod("ToTrace", 1, func(args []Value) (Value, error) { return TLCExtToTrace(args[0]) })
	t.defineStandardMethod("CounterExample", 0, func(args []Value) (Value, error) { return TLCExtCounterExample(), nil })
	t.defineStandardEvaluating("Trace", 0, standardTrace)
	t.defineStandardEvaluating("TLCDefer", 1, standardTLCDefer)
	t.defineStandardMethod("TLCNoOp", 1, func(args []Value) (Value, error) { return TLCExtTLCNoOp(args[0]), nil })
	t.defineStandardMethod("TLCModelValue", 1, func(args []Value) (Value, error) { return TLCExtTLCModelValue(args[0]) })
	t.defineStandardEvaluating("TLCCache", 2, standardTLCCache)
	t.defineStandardMethod("TLCFP", 1, func(args []Value) (Value, error) { return TLCExtTLCFP(args[0]), nil })
	t.defineStandardEvaluating("TLCEvalDefinition", 1, standardTLCEvalDefinition)
	t.defineStandardEvaluating("TLCGetOrDefault", 2, standardTLCGetOrDefault)
	t.defineStandardEvaluating("TLCGetAndSet", 4, standardTLCGetAndSet)

	t.defineStandardMethod("_TLCTraceDeserialize", 1, func(args []Value) (Value, error) {
		path, err := standardStringArg("_TLCTraceDeserialize", args, 0)
		if err != nil {
			return nil, err
		}
		return TLCTraceDeserialize(path)
	})
	t.defineStandardMethod("_TLCTraceSerialize", 2, func(args []Value) (Value, error) {
		path, err := standardStringArg("_TLCTraceSerialize", args, 1)
		if err != nil {
			return nil, err
		}
		return TLCTraceSerialize(args[0], path)
	})
	t.defineStandardEvaluating("_TLCState", 1, standardTLCState, "_TLCTrace!_TLCState", "_JsonTrace!_TLCState")
	t.defineStandardMethod("_Counts", 0, func(args []Value) (Value, error) { return PossibleCounts(), nil })

	return t
}

func (t *Tool) defineStandardValue(name string, value any, aliases ...string) {
	t.DefineName(name, value)
	for _, alias := range aliases {
		t.DefineName(alias, value)
	}
}

func (t *Tool) defineStandardMethod(name string, arity int, eval func([]Value) (Value, error), aliases ...string) {
	value := NewMethodValue(name, 0, func(args []Value, control int) (Value, error) {
		_ = control
		if len(args) != arity {
			return nil, newTLCError(ECGeneral, "%s expected %d arguments, got %d", name, arity, len(args))
		}
		return eval(args)
	})
	t.defineStandardValue(name, value, aliases...)
}

func (t *Tool) defineStandardEvaluating(name string, arity int, eval EvaluatingEvalFunc, aliases ...string) {
	value := NewEvaluatingValue(name, 0, 100, nil, func(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
		if len(args) != arity {
			return nil, newTLCError(ECGeneral, "%s expected %d arguments, got %d", name, arity, len(args))
		}
		return eval(tool, args, con, state, pstate, control, cm)
	})
	t.defineStandardValue(name, value, aliases...)
}

func standardBinaryInt(name string, eval func(*IntValue, *IntValue) (*IntValue, error)) func([]Value) (Value, error) {
	return func(args []Value) (Value, error) {
		x, err := standardIntArg(name, args, 0)
		if err != nil {
			return nil, err
		}
		y, err := standardIntArg(name, args, 1)
		if err != nil {
			return nil, err
		}
		return eval(x, y)
	}
}

func standardIntArg(name string, args []Value, index int) (*IntValue, error) {
	value, ok := args[index].(*IntValue)
	if !ok {
		return nil, newTLCError(ECGeneral, "%s argument %d must be an integer, got %s", name, index+1, args[index])
	}
	return value, nil
}

func standardStringArg(name string, args []Value, index int) (*StringValue, error) {
	value, ok := args[index].(*StringValue)
	if !ok {
		return nil, newTLCError(ECGeneral, "%s argument %d must be a string, got %s", name, index+1, args[index])
	}
	return value, nil
}

func standardTLCGet(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
	index, err := tool.Eval(args[0], con, state, pstate, control, cm)
	if err != nil {
		return nil, err
	}
	return TLCGetValue(tool, index, state, pstate, control)
}

func standardTLCSet(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
	index, err := tool.Eval(args[0], con, state, pstate, control, cm)
	if err != nil {
		return nil, err
	}
	value, err := tool.Eval(args[1], con, state, pstate, control, cm)
	if err != nil {
		return nil, err
	}
	return TLCSet(index, value)
}

func standardAssertError(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
	expected, err := tool.Eval(args[0], con, state, pstate, control, cm)
	if err != nil {
		return nil, err
	}
	expectedString, ok := expected.(*StringValue)
	if !ok {
		return nil, newTLCError(ECGeneral, "AssertError expected a string error, got %s", expected)
	}
	return TLCExtAssertError(expectedString, func() (Value, error) {
		return tool.Eval(args[1], con, state, pstate, control, cm)
	})
}

func standardPickSuccessor(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
	guard, err := tool.Eval(args[0], con, state, pstate, control, cm)
	if err != nil {
		return nil, err
	}
	return TLCExtPickSuccessor(guard, state, pstate)
}

func standardTrace(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
	_ = tool
	_ = args
	_ = con
	_ = pstate
	_ = control
	_ = cm
	return TLCExtTrace(state)
}

func standardTLCState(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
	_ = tool
	_ = args
	_ = con
	_ = pstate
	_ = control
	_ = cm
	return TLCTraceState(state), nil
}

func standardTLCDefer(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
	callable := func() (any, error) {
		for _, arg := range args {
			if _, err := tool.Eval(arg, con, state, pstate, control, cm); err != nil {
				return nil, err
			}
		}
		return nil, nil
	}
	return TLCExtTLCDefer([]*TLCStateMut{state, pstate}, callable), nil
}

func standardTLCCache(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
	expr := args[0]
	closure := args[1]
	key, err := tool.Eval(closure, con, state, pstate, control, cm)
	if err != nil {
		return nil, err
	}

	level := SemanticLevel(expr)
	if tool != nil && level == TLCLevelConstant {
		level = tool.GetLevelBound(expr, con)
	}
	if level == TLCLevelConstant {
		cache, _ := SemanticToolObject(expr).(*TLCExtCache)
		if cache == nil {
			cache = NewTLCExtCache()
			if SemanticToolObject(expr) == nil {
				setSemanticToolObject(expr, cache)
			}
		}
		return cache.Eval(key, func() (Value, error) {
			return tool.Eval(expr, con, state, pstate, control, cm)
		})
	}
	if level == TLCLevelState && state != nil {
		cacheKey := standardTLCCacheKey(expr, closure, key)
		if value := state.GetCached(cacheKey); value != nil {
			return value, nil
		}
		value, err := tool.Eval(expr, con, state, pstate, control, cm)
		if err != nil {
			return nil, err
		}
		return state.SetCached(cacheKey, value), nil
	}
	return tool.Eval(expr, con, state, pstate, control, cm)
}

func standardTLCCacheKey(expr SemanticNode, closure SemanticNode, key Value) int {
	fp := FP64NewString(SemanticString(expr))
	fp = FP64ExtendString(fp, SemanticString(closure))
	if key != nil {
		fp = key.FingerPrint(fp)
	}
	return int(FP64Hash(fp))
}

func standardTLCEvalDefinition(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
	name, err := tool.Eval(args[0], con, state, pstate, control, cm)
	if err != nil {
		return nil, err
	}
	return TLCExtTLCEvalDefinition(tool, name, con, state, pstate, control, cm)
}

func standardTLCGetOrDefault(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
	index, err := tool.Eval(args[0], con, state, pstate, control, cm)
	if err != nil {
		return nil, err
	}
	value, err := TLCGetValue(tool, index, state, pstate, control)
	if err == nil && value != nil {
		return value, nil
	}
	return tool.Eval(args[1], con, state, pstate, control, cm)
}

func standardTLCGetAndSet(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
	index, err := tool.Eval(args[0], con, state, pstate, control, cm)
	if err != nil {
		return nil, err
	}
	oldValue, err := TLCGetValue(tool, index, state, pstate, control)
	if err != nil || oldValue == nil {
		oldValue, err = tool.Eval(args[3], con, state, pstate, control, cm)
		if err != nil {
			return nil, err
		}
	}
	op, err := tool.Eval(args[1], con, state, pstate, control, cm)
	if err != nil {
		return nil, err
	}
	value, err := tool.Eval(args[2], con, state, pstate, control, cm)
	if err != nil {
		return nil, err
	}
	newValue, err := EvalOperatorValue(op, []Value{oldValue, value}, control)
	if err != nil {
		return nil, err
	}
	if _, err := TLCSet(index, newValue); err != nil {
		return nil, err
	}
	return oldValue, nil
}
