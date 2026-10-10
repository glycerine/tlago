package tlc

import "fmt"

var standardTLCEvalMu reentrantReadWriteLock

// SpecProcessor.processModuleOverrides visits inherited Naturals definitions
// when Integers is loaded too. Its GEQ method differs in its argument-error label.
func (t *Tool) InstallIntegerDefinitions() {
	if t != nil {
		t.defineStandardMethod("GEQ", 2, func(args []Value) (Value, error) { return IntGEQ(args[0], args[1]) }, "\\geq")
	}
}

func (t *Tool) InstallStandardDefinitions() *Tool {
	if t == nil {
		return nil
	}
	ensureTLCExtConsole()

	t.defineStandardValue("TRUE", BoolTrue)
	t.defineStandardValue("FALSE", BoolFalse)
	t.defineStandardValue("BOOLEAN", NewSetEnumValue([]Value{BoolFalse, BoolTrue}, true))
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
	t.defineStandardMethod("Len", 1, func(args []Value) (Value, error) { return Len(args[0]) })
	t.defineStandardMethod("Head", 1, func(args []Value) (Value, error) { return Head(args[0]) })
	t.defineStandardMethod("Tail", 1, func(args []Value) (Value, error) { return Tail(args[0]) })
	t.defineStandardMethod("Cons", 2, func(args []Value) (Value, error) { return Cons(args[0], args[1]) })
	t.defineStandardMethod("Append", 2, func(args []Value) (Value, error) { return Append(args[0], args[1]) })
	t.defineStandardMethod("Concat", 2, func(args []Value) (Value, error) { return Concat(args[0], args[1]) }, "\\o")
	t.defineStandardMethod("SubSeq", 3, func(args []Value) (Value, error) { return SubSeq(args[0], args[1], args[2]) })
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
	t.defineStandardEvaluating("TLCEval", 1, standardTLCEval)
	t.defineStandardEvaluatingWithMinLevel("TLCGet", 1, TLCLevelState, standardTLCGet)
	t.defineStandardEvaluating("TLCSet", 2, standardTLCSet)

	t.defineStandardMethod("RandomSubset", 2, func(args []Value) (Value, error) { return RandomSubset(args[0], args[1]) })
	t.defineStandardMethod("RandomSetOfSubsets", 3, func(args []Value) (Value, error) {
		return RandomSetOfSubsets(args[0], args[1], args[2])
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
	t.defineStandardMethod("IOSerialize", 3, func(args []Value) (Value, error) {
		path, err := standardStringArg("IOSerialize", args, 1)
		if err != nil {
			return nil, err
		}
		compress, err := standardBoolArg("IOSerialize", args, 2)
		if err != nil {
			return nil, err
		}
		return IOUtilsIOSerialize(args[0], path, compress)
	})
	t.defineStandardMethod("IODeserialize", 2, func(args []Value) (Value, error) {
		path, err := standardStringArg("IODeserialize", args, 0)
		if err != nil {
			return nil, err
		}
		compress, err := standardBoolArg("IODeserialize", args, 1)
		if err != nil {
			return nil, err
		}
		return IOUtilsIODeserialize(path, compress)
	})
	t.defineStandardPriorityEvaluating("Serialize", 3,
		standardEvaluatingHandler{priority: 50, eval: standardIOUtilsTextSerialize},
		standardEvaluatingHandler{priority: 25, eval: standardJsonTextSerialize},
	)
	t.defineStandardEvaluatingWithPriority("Deserialize", 2, 0, 50, standardIOUtilsTextDeserialize)
	t.defineStandardMethod("IOEnv", 0, func(args []Value) (Value, error) { return IOUtilsIOEnv(), nil })
	t.defineStandardMethod("IOExec", 1, func(args []Value) (Value, error) { return IOUtilsIOExec(args[0]) })
	t.defineStandardMethod("IOEnvExec", 2, func(args []Value) (Value, error) { return IOUtilsIOEnvExec(args[0], args[1]) })
	t.defineStandardMethod("IOExecTemplate", 2, func(args []Value) (Value, error) { return IOUtilsIOExecTemplate(args[0], args[1]) })
	t.defineStandardMethod("IOEnvExecTemplate", 3, func(args []Value) (Value, error) {
		return IOUtilsIOEnvExecTemplate(args[0], args[1], args[2])
	})
	t.defineStandardMethod("atoi", 1, func(args []Value) (Value, error) { return IOUtilsAtoi(args[0]) })
	t.defineStandardMethod("factorial", 1, func(args []Value) (Value, error) { return CombinatoricsFactorial(args[0]) })
	t.defineStandardMethod("choose", 2, func(args []Value) (Value, error) { return CombinatoricsChoose(args[0], args[1]) })
	t.defineStandardMethod("And", 4, func(args []Value) (Value, error) { return BitwiseAnd(args[0], args[1], args[2], args[3]) })
	t.defineStandardMethod("Or", 4, func(args []Value) (Value, error) { return BitwiseOr(args[0], args[1], args[2], args[3]) })
	t.defineStandardMethod("Xor", 4, func(args []Value) (Value, error) { return BitwiseXor(args[0], args[1], args[2], args[3]) })
	t.defineStandardMethod("Not", 1, func(args []Value) (Value, error) { return BitwiseNot(args[0]) })
	t.defineStandardMethod("shiftR", 2, func(args []Value) (Value, error) { return BitwiseShiftR(args[0], args[1]) })
	t.defineStandardMethod("Reduce", 1, func(args []Value) (Value, error) { return DyadicRationalsReduce(args[0]) })
	t.defineStandardMethod("IsInjective", 1, func(args []Value) (Value, error) { return FunctionsIsInjective(args[0]) })
	t.defineStandardMethod("AntiFunction", 1, func(args []Value) (Value, error) { return FunctionsAntiFunction(args[0]) })
	t.defineStandardMethod("FoldFunction", 3, func(args []Value) (Value, error) { return FunctionsFoldFunction(args[0], args[1], args[2]) })
	t.defineStandardMethod("FoldFunctionOnSet", 4, func(args []Value) (Value, error) {
		return FunctionsFoldFunctionOnSet(args[0], args[1], args[2], args[3])
	})
	t.defineStandardMethod("Quantify", 2, func(args []Value) (Value, error) { return FiniteSetsExtQuantify(args[0], args[1]) })
	t.defineStandardMethod("kSubset", 2, func(args []Value) (Value, error) { return FiniteSetsExtKSubset(args[0], args[1]) })
	t.defineStandardMethod("FoldSet", 3, func(args []Value) (Value, error) { return FiniteSetsExtFoldSet(args[0], args[1], args[2]) })
	t.defineStandardMethod("FoldBag", 3, func(args []Value) (Value, error) { return BagsExtFoldBag(args[0], args[1], args[2]) })
	t.defineStandardMethod("CSVWriteRecord", 4, func(args []Value) (Value, error) {
		return CSVWriteRecord(args[0], args[1], args[2], args[3])
	})
	t.defineStandardMethod("CSVWrite", 3, func(args []Value) (Value, error) { return CSVWrite(args[0], args[1], args[2]) })
	t.defineStandardMethod("CSVRead", 3, func(args []Value) (Value, error) { return CSVRead(args[0], args[1], args[2]) })
	t.defineStandardMethod("CSVRecords", 1, func(args []Value) (Value, error) { return CSVRecords(args[0]) })
	t.defineStandardMethod("DotDiGraph", 3, func(args []Value) (Value, error) {
		return GraphVizDotDiGraph(args[0], args[1], args[2])
	})
	t.defineStandardMethod("Graphs!SimplePath", 1, func(args []Value) (Value, error) { return GraphsSimplePath(args[0]) })
	t.defineStandardMethod("Graphs!AreConnectedIn", 3, func(args []Value) (Value, error) {
		return GraphsAreConnectedIn(args[0], args[1], args[2])
	})
	t.defineStandardMethod("Graphs!IsStronglyConnected", 1, func(args []Value) (Value, error) {
		return GraphsIsStronglyConnected(args[0])
	})
	t.defineStandardMethod("UndirectedGraphs!SimplePath", 1, func(args []Value) (Value, error) {
		return UndirectedGraphsSimplePath(args[0])
	})
	t.defineStandardMethod("UndirectedGraphs!AreConnectedIn", 3, func(args []Value) (Value, error) {
		return UndirectedGraphsAreConnectedIn(args[0], args[1], args[2])
	})
	t.defineStandardMethod("UndirectedGraphs!ConnectedComponents", 1, func(args []Value) (Value, error) {
		return UndirectedGraphsConnectedComponents(args[0])
	})
	t.defineStandardMethod("CausalOrder", 4, func(args []Value) (Value, error) {
		return VectorClocksCausalOrder(args[0], args[1], args[2], args[3])
	})
	t.defineStandardMethod("SVGElemToString", 1, func(args []Value) (Value, error) { return SVGElemToString(args[0]) })
	t.defineStandardMethod("NodeOfRingNetwork", 5, func(args []Value) (Value, error) {
		return SVGNodeOfRingNetwork(args[0], args[1], args[2], args[3], args[4])
	})
	t.defineStandardMethod("NodesOfDirectedMultiGraph", 3, func(args []Value) (Value, error) {
		return SVGNodesOfDirectedMultiGraph(args[0], args[1], args[2])
	})
	t.defineStandardMethod("PointOnLine", 3, func(args []Value) (Value, error) {
		return SVGPointOnLine(args[0], args[1], args[2])
	})
	t.defineStandardMethod("SetToSeq", 1, func(args []Value) (Value, error) { return SequencesExtSetToSeq(args[0]) })
	t.defineStandardMethod("SetToSeqs", 1, func(args []Value) (Value, error) { return SequencesExtSetToSeqs(args[0]) })
	t.defineStandardMethod("Contains", 2, func(args []Value) (Value, error) {
		return SequencesExtContains(args[0], args[1])
	})
	t.defineStandardMethod("LongestCommonPrefix", 1, func(args []Value) (Value, error) {
		return SequencesExtLongestCommonPrefix(args[0])
	})
	t.defineStandardMethod("FoldSeq", 3, func(args []Value) (Value, error) {
		return SequencesExtFoldSeq(args[0], args[1], args[2])
	})
	t.defineStandardMethod("FoldLeft", 3, func(args []Value) (Value, error) {
		return SequencesExtFoldLeft(args[0], args[1], args[2])
	})
	t.defineStandardMethod("FoldRight", 3, func(args []Value) (Value, error) {
		return SequencesExtFoldRight(args[0], args[1], args[2])
	})
	t.defineStandardMethod("FoldLeftDomain", 3, func(args []Value) (Value, error) {
		return SequencesExtFoldLeftDomain(args[0], args[1], args[2])
	})
	t.defineStandardMethod("FoldRightDomain", 3, func(args []Value) (Value, error) {
		return SequencesExtFoldRightDomain(args[0], args[1], args[2])
	})
	t.defineStandardEvaluating("ReplaceFirstSubSeq", 3, standardSequencesExtReplaceFirstSubSeq)
	t.defineStandardEvaluating("ReplaceAllSubSeqs", 3, standardSequencesExtReplaceAllSubSeqs)
	t.defineStandardMethod("IsPrefix", 2, func(args []Value) (Value, error) {
		return SequencesExtIsPrefix(args[0], args[1])
	})
	t.defineStandardMethod("SelectInSeq", 2, func(args []Value) (Value, error) {
		return SequencesExtSelectInSeq(args[0], args[1])
	})
	t.defineStandardMethod("SelectInSubSeq", 4, func(args []Value) (Value, error) {
		return SequencesExtSelectInSubSeq(args[0], args[1], args[2], args[3])
	})
	t.defineStandardMethod("SelectLastInSeq", 2, func(args []Value) (Value, error) {
		return SequencesExtSelectLastInSeq(args[0], args[1])
	})
	t.defineStandardMethod("SelectLastInSubSeq", 4, func(args []Value) (Value, error) {
		return SequencesExtSelectLastInSubSeq(args[0], args[1], args[2], args[3])
	})
	t.defineStandardMethod("RemoveFirst", 2, func(args []Value) (Value, error) {
		return SequencesExtRemoveFirst(args[0], args[1])
	})
	t.defineStandardMethod("RemoveFirstMatch", 2, func(args []Value) (Value, error) {
		return SequencesExtRemoveFirstMatch(args[0], args[1])
	})
	t.defineStandardMethod("Suffixes", 1, func(args []Value) (Value, error) { return SequencesExtSuffixes(args[0]) })
	t.defineStandardMethod("AllSubSeqs", 1, func(args []Value) (Value, error) {
		return SequencesExtAllSubSeqs(args[0])
	})
	t.defineStandardMethod("ChiSquare", 3, func(args []Value) (Value, error) {
		return StatisticsChiSquare(args[0], args[1], args[2])
	})

	t.defineStandardEvaluating("AssertError", 2, standardAssertError)
	t.defineStandardEvaluatingWithMinLevel("PickSuccessor", 1, TLCLevelAction, standardPickSuccessor)
	t.defineStandardMethod("ToTrace", 1, func(args []Value) (Value, error) { return TLCExtToTrace(args[0]) })
	t.defineStandardEvaluatingWithMinLevel("CounterExample", 0, 1, standardCounterExample)
	t.defineStandardEvaluatingWithMinLevel("Trace", 0, 1, standardTrace)
	t.defineStandardEvaluating("TLCDefer", 1, standardTLCDefer)
	t.defineStandardMethod("TLCNoOp", 1, func(args []Value) (Value, error) { return TLCExtTLCNoOp(args[0]), nil })
	t.defineStandardMethod("TLCModelValue", 1, func(args []Value) (Value, error) { return TLCExtTLCModelValue(args[0]) })
	t.defineStandardEvaluatingIdentity("TLCCache", 2, 0, standardTLCCache)
	t.defineStandardMethod("TLCFP", 1, func(args []Value) (Value, error) { return TLCExtTLCFP(args[0]), nil })
	t.defineStandardEvaluating("TLCEvalDefinition", 1, standardTLCEvalDefinition)
	t.defineStandardMethod("TLCGetOrDefault", 2, func(args []Value) (Value, error) { return TLCGetOrDefault(args[0], args[1]) })
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
	t.defineStandardEvaluatingWithMinLevel("_TLCState", 1, TLCLevelState, standardTLCState, "_TLCTrace!_TLCState")
	t.defineStandardEvaluatingWithMinLevel("_JsonTrace!_TLCState", 1, TLCLevelState, standardTLCState)
	t.defineStandardMethodWithMinLevel("_Counts", 0, TLCLevelState, func(args []Value) (Value, error) { return PossibleCounts(), nil })

	return t
}

func (t *Tool) defineStandardValue(name string, value any, aliases ...string) {
	t.DefineName(name, value)
	for _, alias := range aliases {
		t.DefineName(alias, value)
	}
}

func (t *Tool) defineStandardMethod(name string, arity int, eval func([]Value) (Value, error), aliases ...string) {
	t.defineStandardMethodWithMinLevel(name, arity, 0, eval, aliases...)
}

func (t *Tool) defineStandardMethodWithMinLevel(name string, arity int, minLevel int, eval func([]Value) (Value, error), aliases ...string) {
	method := name
	// MethodValue reports Method.toString() when a native value operation fails.
	// Preserve the source reflection metadata, including its declared return type.
	switch name {
	case "Print":
		method = "public static tlc2.value.impl.Value tlc2.module.TLC.Print(tlc2.value.impl.Value,tlc2.value.impl.Value)"
	case "PrintT":
		method = "public static tlc2.value.impl.Value tlc2.module.TLC.PrintT(tlc2.value.impl.Value)"
	case "ToString":
		method = "public static tlc2.value.impl.Value tlc2.module.TLC.ToString(tlc2.value.impl.Value)"
	case "IsFiniteSet":
		method = "public static tlc2.value.IBoolValue tlc2.module.FiniteSets.IsFiniteSet(tlc2.value.impl.Value)"
	case "Cardinality":
		method = "public static tlc2.value.impl.IntValue tlc2.module.FiniteSets.Cardinality(tlc2.value.impl.Value)"
	}
	value := NewMethodValue(method, minLevel, func(args []Value, control int) (Value, error) {
		_ = control
		if len(args) != arity {
			return nil, newTLCError(ECGeneral, "%s expected %d arguments, got %d", name, arity, len(args))
		}
		return eval(args)
	})
	value.ParameterCount = arity
	t.defineStandardValue(name, value, aliases...)
}

func (t *Tool) defineStandardEvaluating(name string, arity int, eval EvaluatingEvalFunc, aliases ...string) {
	t.defineStandardEvaluatingWithMinLevel(name, arity, 0, eval, aliases...)
}

// standardEvaluatingMethodSignature retains Method.toString() for override
// diagnostics and operator images. Lookup and fallback still use the TLA+ name.
func standardEvaluatingMethodSignature(name string) string {
	const params = "(tlc2.tool.impl.Tool,tla2sany.semantic.ExprOrOpArgNode[],tlc2.util.Context,tlc2.tool.TLCState,tlc2.tool.TLCState,int,tlc2.tool.coverage.CostModel)"
	switch name {
	case "TLCEval":
		return "public static tlc2.value.impl.Value tlc2.module.TLCEval.tlcEval" + params
	case "_TLCState":
		return "public static tlc2.value.impl.Value tlc2.module._TLCTrace.tlcState" + params
	case "_JsonTrace!_TLCState":
		return "public static tlc2.value.impl.Value tlc2.module._JsonTrace.tlcState" + params
	case "ReplaceFirstSubSeq":
		return "public static tlc2.value.impl.Value tlc2.overrides.SequencesExt.replaceFirstSubSeq" + params
	case "ReplaceAllSubSeqs":
		return "public static tlc2.value.impl.Value tlc2.overrides.SequencesExt.replaceAllSubSeq" + params
	case "TLCGet":
		return "public static tlc2.value.impl.Value tlc2.module.TLCGetSet.TLCGetEval" + params
	case "AssertError":
		return "public static synchronized tlc2.value.impl.Value tlc2.module.TLCExt.assertError" + params
	case "PickSuccessor":
		return "public static synchronized tlc2.value.impl.Value tlc2.module.TLCExt.pickSuccessor" + params
	case "CounterExample":
		return "public static tlc2.value.impl.Value tlc2.module.TLCExt.error" + params + " throws java.io.IOException"
	case "Trace":
		return "public static tlc2.value.impl.TupleValue tlc2.module.TLCExt.getTrace" + params + " throws java.io.IOException"
	case "TLCDefer":
		return "public static tlc2.value.impl.Value tlc2.module.TLCExt.tlcDefer" + params
	case "TLCCache":
		return "public static tlc2.value.impl.Value tlc2.module.TLCExt.tlcEval2" + params
	case "TLCEvalDefinition":
		return "public static tlc2.value.impl.Value tlc2.module.TLCExt.tlcDefByName" + params
	default:
		return name
	}
}

func (t *Tool) defineStandardEvaluatingWithMinLevel(name string, arity int, minLevel int, eval EvaluatingEvalFunc, aliases ...string) {
	opDef := &OpDefNode{SemanticNodeBase: &SemanticNodeBase{}, Name: UniqueStringOf(name), Symbol: NewSymbolNode(name)}
	method := standardEvaluatingMethodSignature(name)
	value := NewEvaluatingValue(method, minLevel, 100, opDef, func(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
		if len(args) != arity {
			return nil, newTLCError(ECGeneral, "%s expected %d arguments, got %d", name, arity, len(args))
		}
		return eval(tool, args, con, state, pstate, control, cm)
	})
	t.defineStandardValue(name, value, aliases...)
}

func (t *Tool) defineStandardEvaluatingIdentity(name string, arity int, identityArg int, eval EvaluatingEvalFunc, aliases ...string) {
	params := make([]*SymbolNode, arity)
	for i := range params {
		params[i] = NewSymbolNode(name + "$arg")
	}
	opDef := &OpDefNode{SemanticNodeBase: &SemanticNodeBase{}, Name: UniqueStringOf(name), Symbol: NewSymbolNode(name), Params: params}
	if identityArg >= 0 && identityArg < len(params) {
		opDef.Body = NewOpApplNode(params[identityArg])
	}
	value := NewEvaluatingValue(standardEvaluatingMethodSignature(name), 0, 100, opDef, func(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
		if len(args) != arity {
			return nil, newTLCError(ECGeneral, "%s expected %d arguments, got %d", name, arity, len(args))
		}
		return eval(tool, args, con, state, pstate, control, cm)
	})
	t.defineStandardValue(name, value, aliases...)
}

type standardEvaluatingHandler struct {
	priority int
	eval     EvaluatingEvalFunc
}

func (t *Tool) defineStandardEvaluatingWithPriority(name string, arity int, minLevel int, priority int, eval EvaluatingEvalFunc, aliases ...string) {
	t.defineStandardValue(name, newStandardEvaluatingValue(name, arity, minLevel, priority, eval), aliases...)
}

func (t *Tool) defineStandardPriorityEvaluating(name string, arity int, handlers ...standardEvaluatingHandler) {
	opDef := standardSyntheticOpDef(name, arity, NewValueNode(ValUndef))
	value := &PriorityEvaluatingValue{}
	for _, handler := range handlers {
		value.Add(newStandardEvaluatingValueForOpDef(name, arity, 0, handler.priority, opDef, handler.eval))
	}
	t.defineStandardValue(name, value)
}

func newStandardEvaluatingValue(name string, arity int, minLevel int, priority int, eval EvaluatingEvalFunc) *EvaluatingValue {
	return newStandardEvaluatingValueForOpDef(name, arity, minLevel, priority, standardSyntheticOpDef(name, arity, NewValueNode(ValUndef)), eval)
}

func newStandardEvaluatingValueForOpDef(name string, arity int, minLevel int, priority int, opDef *OpDefNode, eval EvaluatingEvalFunc) *EvaluatingValue {
	method := name
	switch name {
	case "Serialize":
		owner := "tlc2.overrides.IOUtils"
		if priority == 25 {
			owner = "tlc2.module.Json"
		}
		method = "public static synchronized tlc2.value.impl.Value " + owner + ".textSerialize(tlc2.tool.impl.Tool,tla2sany.semantic.ExprOrOpArgNode[],tlc2.util.Context,tlc2.tool.TLCState,tlc2.tool.TLCState,int,tlc2.tool.coverage.CostModel)"
	case "Deserialize":
		method = "public static synchronized tlc2.value.impl.Value tlc2.overrides.IOUtils.textDeserialize(tlc2.tool.impl.Tool,tla2sany.semantic.ExprOrOpArgNode[],tlc2.util.Context,tlc2.tool.TLCState,tlc2.tool.TLCState,int,tlc2.tool.coverage.CostModel)"
	}

	return NewEvaluatingValue(method, minLevel, priority, opDef, func(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
		if len(args) != arity {
			return nil, newTLCError(ECGeneral, "%s expected %d arguments, got %d", name, arity, len(args))
		}
		return eval(tool, args, con, state, pstate, control, cm)
	})
}

func standardSyntheticOpDef(name string, arity int, body SemanticNode) *OpDefNode {
	params := make([]*SymbolNode, arity)
	for i := range params {
		params[i] = NewSymbolNode(name + "$arg")
	}
	return &OpDefNode{SemanticNodeBase: &SemanticNodeBase{}, Name: UniqueStringOf(name), Symbol: NewSymbolNode(name), Params: params, Body: body}
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

func standardBoolArg(name string, args []Value, index int) (*BoolValue, error) {
	value, ok := args[index].(*BoolValue)
	if !ok {
		return nil, newTLCError(ECGeneral, "%s argument %d must be a boolean, got %s", name, index+1, args[index])
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

func standardJsonTextSerialize(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
	jsonClassMonitor.Lock()
	defer jsonClassMonitor.Unlock()
	options, err := tool.Eval(args[2], con, state, pstate, control, cm)
	if err != nil {
		return nil, err
	}
	if options == nil {
		return nil, NewNullPointerException()
	}
	opts := asRecordValue(options)
	if opts == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "third", "ndJsonSerialize", "sequence", ValuesPPR(options))
	}
	format, err := ioUtilsTXTFormat(opts)
	if err != nil {
		return nil, err
	}
	if format != "NDJSON" {
		return nil, nil
	}
	payload, err := tool.Eval(args[0], con, state, pstate, control, cm)
	if err != nil {
		return nil, err
	}
	if payload == nil {
		return nil, NewNullPointerException()
	}
	tuple := asTupleValue(payload)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "Serialize", "sequence", ValuesPPR(payload))
	}
	dest, err := tool.Eval(args[1], con, state, pstate, control, cm)
	if err != nil {
		return nil, err
	}
	if dest == nil {
		return nil, NewNullPointerException()
	}
	path, castError := ioUtilsTXTString(dest)
	if castError != nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "ndJsonSerialize", "sequence", ValuesPPR(dest))
	}
	return jsonTextSerializeTuple(path, tuple, opts)
}

func standardIOUtilsTextSerialize(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
	options, err := ioUtilsTXTEval(tool, args[2], con, state, pstate, control, cm)
	if err != nil {
		return ioUtilsTXTFailure("Serialize", "invalid parameters", err), nil
	}
	opts, err := ioUtilsTXTRecord(options)
	if err != nil {
		return ioUtilsTXTFailure("Serialize", "invalid parameters", err), nil
	}
	format, err := ioUtilsTXTFormat(opts)
	if err != nil {
		return nil, err
	}
	if format != "TXT" {
		return nil, nil
	}
	payload, err := ioUtilsTXTEval(tool, args[0], con, state, pstate, control, cm)
	if err != nil {
		return ioUtilsTXTFailure("Serialize", "invalid parameters", err), nil
	}
	if _, err := ioUtilsTXTString(payload); err != nil {
		return ioUtilsTXTFailure("Serialize", "invalid parameters", err), nil
	}
	dest, err := ioUtilsTXTEval(tool, args[1], con, state, pstate, control, cm)
	if err != nil {
		return ioUtilsTXTFailure("Serialize", "invalid parameters", err), nil
	}
	return ioUtilsSerializeTXT(payload, dest, opts), nil
}

func standardIOUtilsTextDeserialize(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
	options, err := ioUtilsTXTEval(tool, args[1], con, state, pstate, control, cm)
	if err != nil {
		return ioUtilsTXTFailure("Deserialize", "invalid parameters", err), nil
	}
	opts, err := ioUtilsTXTRecord(options)
	if err != nil {
		return ioUtilsTXTFailure("Deserialize", "invalid parameters", err), nil
	}
	format, err := ioUtilsTXTFormat(opts)
	if err != nil {
		return nil, err
	}
	if format != "TXT" {
		return nil, nil
	}
	src, err := ioUtilsTXTEval(tool, args[0], con, state, pstate, control, cm)
	if err != nil {
		return ioUtilsTXTFailure("Deserialize", "invalid parameters", err), nil
	}
	return IOUtilsDeserialize(src, options)
}

func standardSequencesExtReplaceFirstSubSeq(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
	replacement, subseq, target, ok, err := standardSequencesExtReplaceArgs(tool, args, con, state, pstate, control, cm)
	if err != nil || !ok {
		return nil, err
	}
	return SequencesExtReplaceFirstSubSeq(replacement, subseq, target)
}

func standardSequencesExtReplaceAllSubSeqs(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
	replacement, subseq, target, ok, err := standardSequencesExtReplaceArgs(tool, args, con, state, pstate, control, cm)
	if err != nil || !ok {
		return nil, err
	}
	return SequencesExtReplaceAllSubSeqs(replacement, subseq, target)
}

func standardSequencesExtReplaceArgs(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, Value, Value, bool, error) {
	replacement, err := tool.Eval(args[0], con, state, pstate, control, cm)
	if err != nil {
		return nil, nil, nil, false, err
	}
	subseq, err := tool.Eval(args[1], con, state, pstate, control, cm)
	if err != nil {
		return nil, nil, nil, false, err
	}
	target, err := tool.Eval(args[2], con, state, pstate, control, cm)
	if err != nil {
		return nil, nil, nil, false, err
	}
	if _, ok := replacement.(*StringValue); !ok {
		return replacement, subseq, target, false, nil
	}
	if _, ok := subseq.(*StringValue); !ok {
		return replacement, subseq, target, false, nil
	}
	if _, ok := target.(*StringValue); !ok {
		return replacement, subseq, target, false, nil
	}
	return replacement, subseq, target, true, nil
}

func standardTLCGet(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
	index, err := tool.Eval(args[0], con, state, pstate, control, cm)
	if err != nil {
		return nil, err
	}
	return TLCGetValue(tool, index, state, pstate, control)
}

func standardTLCEval(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
	expr := args[0]
	level := SemanticLevel(expr)
	if level > TLCLevelConstant || (con != nil && !con.IsDeepEmpty()) {
		value, err := tool.Eval(expr, con, state, pstate, control, cm)
		if err != nil {
			return nil, err
		}
		return TLCEvalChecked(value)
	}
	return standardTLCEvalConst(tool, expr, cm)
}

func standardTLCEvalConst(tool *Tool, expr SemanticNode, cm CostModel) (Value, error) {
	standardTLCEvalMu.RLock()
	if object := semanticCachedTLCEvalObject(tool, expr); object != nil {
		// Source releases the read lock around the value cast, but a failure
		// during WorkerValue.mux occurs before that finally block.
		defer standardTLCEvalMu.RUnlock()
		return castTLCEvalObject(object), nil
	}
	standardTLCEvalMu.RUnlock()

	standardTLCEvalMu.Lock()
	defer standardTLCEvalMu.Unlock()
	if object := semanticCachedTLCEvalObject(tool, expr); object != nil {
		return castTLCEvalObject(object), nil
	}
	demuxed, err := DemuxWorkerValue(func() (Value, error) {
		return tool.Eval(expr, EmptyContext, EmptyState, EmptyState, EvalClear, cm)
	})
	if err != nil {
		return nil, err
	}
	workerID, _ := CurrentWorkerID()
	value := MuxWorkerValue(demuxed, workerID)
	value, err = TLCEvalChecked(value)
	if err != nil {
		return nil, err
	}
	setSemanticTLCEvalValue(tool, expr, value)
	return value, nil
}

func semanticCachedTLCEvalObject(tool *Tool, node SemanticNode) any {
	object := SemanticToolObjectForTool(tool, node)
	if worker, ok := object.(*WorkerValue); ok {
		workerID, _ := CurrentWorkerID()
		object = worker.ValueForWorker(workerID)
	}
	return object
}

func castTLCEvalObject(object any) Value {
	value, ok := object.(Value)
	if !ok {
		panic(NewClassCastException("tool object is not a TLC value"))
	}
	return value
}

func setSemanticTLCEvalValue(tool *Tool, node SemanticNode, value Value) {
	SetSemanticToolObjectForTool(tool, node, value)
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
	tlcExtClassMonitor.Lock()
	defer tlcExtClassMonitor.Unlock()
	if _, ok := args[0].(*StringNode); !ok {
		panic(NewTLCRuntimeExceptionMessage(fmt.Sprintf("In computing AssertError, a non-string expression (%s) was used as the err of an AssertError(err, exp).", SemanticString(args[0]))))
	}
	_, failure := tool.Eval(args[1], con, state, pstate, control, cm)
	if failure == nil {
		return BoolFalse, nil
	}
	if !isJavaEvalOrRuntimeException(failure) {
		return nil, failure
	}
	// Java evaluates the expected message only in its catch block, using the
	// state-expression overload: empty successor, clear control, no coverage.
	value, err := tool.Eval(args[0], con, state)
	if err != nil {
		return nil, err
	}
	var expected *StringValue
	switch value := value.(type) {
	case nil:
	case *StringValue:
		expected = value
	case *DebuggerValue:
		expected = value.StringValue
	default:
		panic(valueStreamClassCast(value, "tlc2.value.impl.StringValue"))
	}
	if expected == nil || expected.Val == nil {
		panic(NewNullPointerException())
	}
	return TLCExtAssertError(expected, func() (Value, error) { return nil, failure })
}

func standardPickSuccessor(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
	tlcExtClassMonitor.Lock()
	defer tlcExtClassMonitor.Unlock()
	if tlcExtPickSuccessorSeen(pstate) {
		return BoolTrue, nil
	}
	guard, err := tool.Eval(args[0], con, state, pstate, control, cm)
	if err != nil {
		return nil, err
	}
	boolGuard, ok := guard.(*BoolValue)
	if !ok {
		if guard == nil {
			panic(NewNullPointerException())
		}
		panic(NewTLCRuntimeExceptionMessage(fmt.Sprintf("In evaluating TLCExt!PickSuccessor, a non-boolean expression (%s) was used as the condition of an IF.\n%s", guard.KindString(), SemanticString(args[0]))))
	}
	if boolGuard == nil {
		panic(NewNullPointerException())
	}
	return tlcExtPickSuccessorGuard(tool, guard, state, pstate)
}

func standardTrace(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
	_ = args
	_ = pstate
	_ = control
	_ = cm
	_ = con
	return TLCExtTraceWithTool(tool, state)
}

func standardCounterExample(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
	_ = args
	_ = state
	_ = pstate
	_ = control
	_ = cm
	return TLCExtCounterExampleWithTool(tool, con), nil
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
	return TLCExtTLCDefer([]*TLCStateMut{state, pstate}, callable)
}

func standardTLCCache(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
	expr := args[0]
	closure := args[1]

	level := SemanticLevel(expr)
	if level == TLCLevelConstant {
		key, err := tool.Eval(closure, con, state, pstate, control, cm)
		if err != nil {
			return nil, err
		}
		return evalTLCExtCache(nil, tool, expr, key, func() (Value, error) {
			return tool.Eval(expr, con, state, pstate, control, cm)
		})
	}
	if level == TLCLevelState {
		// Java uses the state-expression overload here, independently of the
		// successor, control and coverage model supplied to the cache override.
		key, err := tool.Eval(closure, con, state, EmptyState, EvalClear, DoNotRecordCostModel)
		if err != nil {
			return nil, err
		}
		cacheKey := standardTLCCacheKey(expr, closure, key)
		if state == nil {
			panic(NewNullPointerException())
		}
		if value := state.GetCached(cacheKey); value != nil {
			return value, nil
		}
		value, err := tool.Eval(expr, con, state, pstate, control, cm)
		if err != nil {
			return nil, err
		}
		return state.SetCached(cacheKey, value), nil
	}
	return nil, nil
}

func standardTLCCacheKey(expr SemanticNode, closure SemanticNode, key Value) int {
	if key == nil {
		panic(NewNullPointerException())
	}
	return int(SemanticJavaHashCode(expr) ^ SemanticJavaHashCode(closure) ^ ValueJavaHashCode(key))
}

func standardTLCEvalDefinition(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
	name, err := tool.Eval(args[0], con, state, pstate, control, cm)
	if err != nil {
		return nil, err
	}
	return TLCExtTLCEvalDefinition(tool, name, con, state, pstate, control, cm)
}

func standardTLCGetAndSet(tool *Tool, args []SemanticNode, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cm CostModel) (Value, error) {
	index, err := tool.Eval(args[0], con, state, pstate, control, cm)
	if err != nil {
		return nil, err
	}
	defaultValue, err := tool.Eval(args[3], con, state, pstate, control, cm)
	if err != nil {
		return nil, err
	}
	oldValue, err := TLCGetValue(tool, index, state, pstate, control)
	if err != nil || oldValue == nil {
		oldValue = defaultValue
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
