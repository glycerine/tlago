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
