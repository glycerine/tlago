package tlc

import "fmt"

// WithStandardMethodMetadata retains the declaring module of numeric overrides.
// Naturals and Integers can install different Java methods on the same inherited
// definition body. Clone the wrapper so updating one module does not rewrite the
// previously captured wrapper or its aliases.
func WithStandardMethodMetadata(value any, module, name string) any {
	method, ok := value.(*MethodValue)
	if !ok {
		return value
	}
	signature := numericMethodSignature(module, name)
	if signature == "" || signature == method.Name {
		return value
	}
	out := *method
	out.Name = signature
	out.Label = "<Java Method: " + signature + ">"
	out.owner = &out
	return &out
}

func numericMethodSignature(module, name string) string {
	if module != "Naturals" && module != "Integers" {
		return ""
	}
	const values = "tlc2.value.impl.Value,tlc2.value.impl.Value"
	const ints = "tlc2.value.impl.IntValue,tlc2.value.impl.IntValue"
	var method, result, params string
	switch TLARegistryMapName(name) {
	case "+":
		method, result, params = "Plus", "tlc2.value.impl.IntValue", ints
	case "-":
		method, result, params = "Minus", "tlc2.value.impl.IntValue", ints
	case "*":
		method, result, params = "Times", "tlc2.value.impl.IntValue", ints
	case "<":
		method, result, params = "LT", "tlc2.value.IBoolValue", values
	case "\\leq":
		method, result, params = "LE", "tlc2.value.IBoolValue", values
	case ">":
		method, result, params = "GT", "tlc2.value.IBoolValue", values
		if module == "Integers" {
			result = "tlc2.value.impl.BoolValue"
		}
	case "\\geq":
		method, result, params = "GEQ", "tlc2.value.IBoolValue", values
	case "..":
		method, result, params = "DotDot", "tlc2.value.impl.IntervalValue", ints
	case "\\div":
		method, result, params = "Divide", "tlc2.value.impl.IntValue", ints
	case "%":
		method, result, params = "Mod", "tlc2.value.impl.IntValue", ints
	case "^":
		method, result, params = "Expt", "tlc2.value.impl.IntValue", ints
	case "-.":
		if module != "Integers" {
			return ""
		}
		method, result, params = "Neg", "tlc2.value.impl.IntValue", "tlc2.value.impl.IntValue"
	default:
		return ""
	}
	return fmt.Sprintf("public static %s tlc2.module.%s.%s(%s)", result, module, method, params)
}

// Source Method.toString metadata for the represented value-array overrides.
// Keep declared return types and arities; lookup aliases share these wrappers.
var standardValueMethodSignatures = map[string]string{
	"Warshall":           "public static tlc2.value.impl.Value tlc2.module.TransitiveClosure.Warshall(tlc2.value.impl.Value)",
	"IsFiniteSet":        "public static tlc2.value.IBoolValue tlc2.module.FiniteSets.IsFiniteSet(tlc2.value.impl.Value)",
	"Cardinality":        "public static tlc2.value.impl.IntValue tlc2.module.FiniteSets.Cardinality(tlc2.value.impl.Value)",
	"Seq":                "public static tlc2.value.impl.Value tlc2.module.Sequences.Seq(tlc2.value.impl.Value)",
	"Len":                "public static tlc2.value.impl.IntValue tlc2.module.Sequences.Len(tlc2.value.impl.Value)",
	"Head":               "public static tlc2.value.impl.Value tlc2.module.Sequences.Head(tlc2.value.impl.Value)",
	"Tail":               "public static tlc2.value.impl.Value tlc2.module.Sequences.Tail(tlc2.value.impl.Value)",
	"Cons":               "public static tlc2.value.impl.Value tlc2.module.Sequences.Cons(tlc2.value.impl.Value,tlc2.value.impl.Value)",
	"Append":             "public static tlc2.value.impl.Value tlc2.module.Sequences.Append(tlc2.value.impl.Value,tlc2.value.impl.Value)",
	"Concat":             "public static tlc2.value.impl.Value tlc2.module.Sequences.Concat(tlc2.value.impl.Value,tlc2.value.impl.Value)",
	"SubSeq":             "public static tlc2.value.impl.Value tlc2.module.Sequences.SubSeq(tlc2.value.impl.Value,tlc2.value.impl.Value,tlc2.value.impl.Value)",
	"SelectSeq":          "public static tlc2.value.impl.Value tlc2.module.Sequences.SelectSeq(tlc2.value.impl.Value,tlc2.value.impl.Value)",
	"EmptyBag":           "public static tlc2.value.impl.Value tlc2.module.Bags.EmptyBag()",
	"IsABag":             "public static tlc2.value.IBoolValue tlc2.module.Bags.IsABag(tlc2.value.impl.Value)",
	"BagCardinality":     "public static tlc2.value.impl.IntValue tlc2.module.Bags.BagCardinality(tlc2.value.impl.Value)",
	"BagIn":              "public static tlc2.value.IBoolValue tlc2.module.Bags.BagIn(tlc2.value.impl.Value,tlc2.value.impl.Value)",
	"CopiesIn":           "public static tlc2.value.impl.IntValue tlc2.module.Bags.CopiesIn(tlc2.value.impl.Value,tlc2.value.impl.Value)",
	"BagCup":             "public static tlc2.value.impl.Value tlc2.module.Bags.BagCup(tlc2.value.impl.Value,tlc2.value.impl.Value)",
	"BagDiff":            "public static tlc2.value.impl.Value tlc2.module.Bags.BagDiff(tlc2.value.impl.Value,tlc2.value.impl.Value)",
	"BagUnion":           "public static tlc2.value.impl.Value tlc2.module.Bags.BagUnion(tlc2.value.impl.Value)",
	"SqSubseteq":         "public static tlc2.value.IBoolValue tlc2.module.Bags.SqSubseteq(tlc2.value.impl.Value,tlc2.value.impl.Value)",
	"BagOfAll":           "public static tlc2.value.impl.Value tlc2.module.Bags.BagOfAll(tlc2.value.impl.Value,tlc2.value.impl.Value)",
	"BagToSet":           "public static tlc2.value.impl.Value tlc2.module.Bags.BagToSet(tlc2.value.impl.Value)",
	"SetToBag":           "public static tlc2.value.impl.Value tlc2.module.Bags.SetToBag(tlc2.value.impl.Value)",
	"Print":              "public static tlc2.value.impl.Value tlc2.module.TLC.Print(tlc2.value.impl.Value,tlc2.value.impl.Value)",
	"PrintT":             "public static tlc2.value.impl.Value tlc2.module.TLC.PrintT(tlc2.value.impl.Value)",
	"Assert":             "public static tlc2.value.impl.Value tlc2.module.TLC.Assert(tlc2.value.impl.Value,tlc2.value.impl.Value)",
	"JavaTime":           "public static tlc2.value.impl.Value tlc2.module.TLC.JavaTime()",
	"MakeFcn":            "public static tlc2.value.impl.Value tlc2.module.TLC.MakeFcn(tlc2.value.impl.Value,tlc2.value.impl.Value)",
	"CombineFcn":         "public static tlc2.value.impl.Value tlc2.module.TLC.CombineFcn(tlc2.value.impl.Value,tlc2.value.impl.Value)",
	"Permutations":       "public static tlc2.value.impl.Value tlc2.module.TLC.Permutations(tlc2.value.impl.Value)",
	"SortSeq":            "public static tlc2.value.impl.Value tlc2.module.TLC.SortSeq(tlc2.value.impl.Value,tlc2.value.impl.Value)",
	"RandomElement":      "public static tlc2.value.impl.Value tlc2.module.TLC.RandomElement(tlc2.value.impl.Value)",
	"ToString":           "public static tlc2.value.impl.Value tlc2.module.TLC.ToString(tlc2.value.impl.Value)",
	"RandomSubset":       "public static tlc2.value.impl.Value tlc2.module.Randomization.RandomSubset(tlc2.value.impl.Value,tlc2.value.impl.Value)",
	"RandomSetOfSubsets": "public static tlc2.value.impl.Value tlc2.module.Randomization.RandomSetOfSubsets(tlc2.value.impl.Value,tlc2.value.impl.Value,tlc2.value.impl.Value)",
}

type standardAnnotatedValueMethod struct {
	signature string
	minLevel  int
}

// Metadata from the represented TLAPlusOperator annotations. Qualified keys
// distinguish same-named graph operators without changing their lookup aliases.
var standardAnnotatedValueMethods = map[string]standardAnnotatedValueMethod{
	"ndJsonSerialize":                      {"public static synchronized tlc2.value.impl.BoolValue tlc2.overrides.Json.ndSerialize(tlc2.value.impl.StringValue,tlc2.value.impl.Value) throws java.io.IOException", 0},
	"JsonSerialize":                        {"public static synchronized tlc2.value.impl.BoolValue tlc2.overrides.Json.serialize(tlc2.value.impl.StringValue,tlc2.value.impl.Value) throws java.io.IOException", 0},
	"JsonDeserialize":                      {"public static tlc2.value.IValue tlc2.overrides.Json.deserialize(tlc2.value.impl.StringValue) throws java.io.IOException", 0},
	"ndJsonDeserialize":                    {"public static tlc2.value.IValue tlc2.overrides.Json.ndDeserialize(tlc2.value.impl.StringValue) throws java.io.IOException", 0},
	"ToJson":                               {"public static tlc2.value.impl.StringValue tlc2.overrides.Json.toJson(tlc2.value.IValue) throws java.io.IOException", 0},
	"ToJsonArray":                          {"public static tlc2.value.impl.StringValue tlc2.overrides.Json.toJsonArray(tlc2.value.IValue) throws java.io.IOException", 0},
	"ToJsonObject":                         {"public static tlc2.value.impl.StringValue tlc2.overrides.Json.toJsonObject(tlc2.value.IValue) throws java.io.IOException", 0},
	"IODeserialize":                        {"public static final tlc2.value.IValue tlc2.overrides.IOUtils.ioDeserialize(tlc2.value.impl.StringValue,tlc2.value.impl.BoolValue) throws java.io.IOException", 0},
	"IOSerialize":                          {"public static final tlc2.value.IValue tlc2.overrides.IOUtils.ioSerialize(tlc2.value.IValue,tlc2.value.impl.StringValue,tlc2.value.impl.BoolValue) throws java.io.IOException", 0},
	"atoi":                                 {"public static tlc2.value.impl.Value tlc2.overrides.IOUtils.atoi(tlc2.value.impl.Value)", 0},
	"IOEnv":                                {"public static tlc2.value.impl.Value tlc2.overrides.IOUtils.ioEnv() throws java.io.IOException,java.lang.InterruptedException", 0},
	"IOEnvExec":                            {"public static tlc2.value.impl.Value tlc2.overrides.IOUtils.ioEnvExec(tlc2.value.impl.Value,tlc2.value.impl.Value) throws java.io.IOException,java.lang.InterruptedException", 1},
	"IOEnvExecTemplate":                    {"public static tlc2.value.impl.Value tlc2.overrides.IOUtils.ioEnvExecTemplate(tlc2.value.impl.Value,tlc2.value.impl.Value,tlc2.value.impl.Value) throws java.io.IOException,java.lang.InterruptedException", 1},
	"IOExec":                               {"public static tlc2.value.impl.Value tlc2.overrides.IOUtils.ioExec(tlc2.value.impl.Value) throws java.io.IOException,java.lang.InterruptedException", 1},
	"IOExecTemplate":                       {"public static tlc2.value.impl.Value tlc2.overrides.IOUtils.ioExecTemplate(tlc2.value.impl.Value,tlc2.value.impl.Value) throws java.io.IOException,java.lang.InterruptedException", 1},
	"choose":                               {"public static tlc2.value.impl.Value tlc2.overrides.Combinatorics.chse(tlc2.value.impl.Value,tlc2.value.impl.Value)", 0},
	"factorial":                            {"public static tlc2.value.impl.Value tlc2.overrides.Combinatorics.fact(tlc2.value.impl.Value)", 0},
	"And":                                  {"public static tlc2.value.impl.IntValue tlc2.overrides.Bitwise.And(tlc2.value.impl.IntValue,tlc2.value.impl.IntValue,tlc2.value.impl.IntValue,tlc2.value.impl.IntValue)", 0},
	"Not":                                  {"public static tlc2.value.impl.IntValue tlc2.overrides.Bitwise.NotR(tlc2.value.impl.IntValue)", 0},
	"Or":                                   {"public static tlc2.value.impl.IntValue tlc2.overrides.Bitwise.Or(tlc2.value.impl.IntValue,tlc2.value.impl.IntValue,tlc2.value.impl.IntValue,tlc2.value.impl.IntValue)", 0},
	"Xor":                                  {"public static tlc2.value.impl.IntValue tlc2.overrides.Bitwise.Xor(tlc2.value.impl.IntValue,tlc2.value.impl.IntValue,tlc2.value.impl.IntValue,tlc2.value.impl.IntValue)", 0},
	"shiftR":                               {"public static tlc2.value.impl.IntValue tlc2.overrides.Bitwise.shiftR(tlc2.value.impl.IntValue,tlc2.value.impl.IntValue)", 0},
	"Reduce":                               {"public static tlc2.value.impl.Value tlc2.overrides.DyadicRationals.reduce(tlc2.value.impl.Value)", 0},
	"IsInjective":                          {"public static tlc2.value.impl.BoolValue tlc2.overrides.Functions.IsInjective(tlc2.value.impl.Value)", 0},
	"AntiFunction":                         {"public static tlc2.value.impl.Value tlc2.overrides.Functions.antiFunction(tlc2.value.impl.Value)", 0},
	"FoldFunction":                         {"public static tlc2.value.impl.Value tlc2.overrides.Functions.foldFunction(tlc2.value.impl.OpValue,tlc2.value.impl.Value,tlc2.value.impl.Value)", 0},
	"FoldFunctionOnSet":                    {"public static tlc2.value.impl.Value tlc2.overrides.Functions.foldFunctionOnSet(tlc2.value.impl.OpValue,tlc2.value.impl.Value,tlc2.value.impl.Value,tlc2.value.impl.Value)", 0},
	"FoldSet":                              {"public static tlc2.value.impl.Value tlc2.overrides.FiniteSetsExt.foldSet(tlc2.value.impl.OpValue,tlc2.value.impl.Value,tlc2.value.impl.Enumerable)", 0},
	"kSubset":                              {"public static tlc2.value.impl.Value tlc2.overrides.FiniteSetsExt.kSubset(tlc2.value.impl.Value,tlc2.value.impl.Value)", 0},
	"Quantify":                             {"public static tlc2.value.impl.Value tlc2.overrides.FiniteSetsExt.quantify(tlc2.value.impl.Value,tlc2.value.impl.OpValue)", 0},
	"FoldBag":                              {"public static tlc2.value.impl.Value tlc2.overrides.BagsExt.foldBag(tlc2.value.impl.OpValue,tlc2.value.impl.Value,tlc2.value.impl.Value)", 0},
	"CSVRead":                              {"public static tlc2.value.impl.Value tlc2.overrides.CSV.read(tlc2.value.impl.Value,tlc2.value.impl.StringValue,tlc2.value.impl.StringValue) throws java.io.IOException", 0},
	"CSVRecords":                           {"public static tlc2.value.impl.Value tlc2.overrides.CSV.records(tlc2.value.impl.StringValue) throws java.io.IOException", 0},
	"CSVWrite":                             {"public static tlc2.value.impl.Value tlc2.overrides.CSV.write(tlc2.value.impl.StringValue,tlc2.value.impl.Value,tlc2.value.impl.StringValue) throws java.io.IOException", 0},
	"CSVWriteRecord":                       {"public static tlc2.value.impl.Value tlc2.overrides.CSV.writeRecord(tlc2.value.impl.Value,tlc2.value.impl.StringValue,tlc2.value.impl.BoolValue,tlc2.value.impl.StringValue) throws java.io.IOException", 0},
	"DotDiGraph":                           {"public static tlc2.value.impl.Value tlc2.overrides.GraphViz.dotDiGraph(tlc2.value.impl.Value,tlc2.value.impl.OpValue,tlc2.value.impl.OpValue) throws java.lang.Exception", 0},
	"Graphs!AreConnectedIn":                {"public static tlc2.value.impl.Value tlc2.overrides.Graphs.areConnectedIn(tlc2.value.impl.Value,tlc2.value.impl.Value,tlc2.value.impl.Value)", 0},
	"Graphs!IsStronglyConnected":           {"public static tlc2.value.impl.Value tlc2.overrides.Graphs.isStronglyConnected(tlc2.value.impl.Value)", 0},
	"Graphs!SimplePath":                    {"public static tlc2.value.impl.Value tlc2.overrides.Graphs.simplePath(tlc2.value.impl.Value)", 0},
	"UndirectedGraphs!AreConnectedIn":      {"public static tlc2.value.impl.Value tlc2.overrides.UndirectedGraphs.areConnectedIn(tlc2.value.impl.Value,tlc2.value.impl.Value,tlc2.value.impl.Value)", 0},
	"UndirectedGraphs!ConnectedComponents": {"public static tlc2.value.impl.Value tlc2.overrides.UndirectedGraphs.connectedComponents(tlc2.value.impl.Value)", 0},
	"UndirectedGraphs!SimplePath":          {"public static tlc2.value.impl.Value tlc2.overrides.UndirectedGraphs.simplePath(tlc2.value.impl.Value)", 0},
	"CausalOrder":                          {"public static tlc2.value.impl.Value tlc2.overrides.VectorClocks.causalOrder(tlc2.value.impl.TupleValue,tlc2.value.impl.OpValue,tlc2.value.impl.OpValue,tlc2.value.impl.OpValue)", 0},
	"SVGElemToString":                      {"public static tlc2.value.impl.Value tlc2.overrides.SVG.SVGElemToString(tlc2.value.impl.Value) throws java.lang.Exception", 0},
	"NodesOfDirectedMultiGraph":            {"public static tlc2.value.impl.Value tlc2.overrides.SVG.directedMultiGraph(tlc2.value.impl.Value,tlc2.value.impl.Value,tlc2.value.impl.Value) throws java.lang.Exception", 0},
	"PointOnLine":                          {"public static tlc2.value.impl.Value tlc2.overrides.SVG.pointOnLine(tlc2.value.impl.RecordValue,tlc2.value.impl.RecordValue,tlc2.value.impl.IntValue) throws java.lang.Exception", 0},
	"NodeOfRingNetwork":                    {"public static tlc2.value.impl.Value tlc2.overrides.SVG.ringNetwork(tlc2.value.impl.IntValue,tlc2.value.impl.IntValue,tlc2.value.impl.IntValue,tlc2.value.impl.IntValue,tlc2.value.impl.IntValue) throws java.lang.Exception", 0},
	"AllSubSeqs":                           {"public static tlc2.value.impl.Value tlc2.overrides.SequencesExt.AllSubSeqs(tlc2.value.impl.Value)", 0},
	"Contains":                             {"public static tlc2.value.impl.Value tlc2.overrides.SequencesExt.Contains(tlc2.value.impl.Value,tlc2.value.impl.Value)", 0},
	"SelectInSubSeq":                       {"public static tlc2.value.impl.Value tlc2.overrides.SequencesExt.SelectInSubSeq(tlc2.value.impl.Value,tlc2.value.impl.Value,tlc2.value.impl.Value,tlc2.value.impl.OpValue)", 0},
	"SelectLastInSubSeq":                   {"public static tlc2.value.impl.Value tlc2.overrides.SequencesExt.SelectLastInSubSeq(tlc2.value.impl.Value,tlc2.value.impl.Value,tlc2.value.impl.Value,tlc2.value.impl.OpValue)", 0},
	"SetToSeq":                             {"public static tlc2.value.impl.Value tlc2.overrides.SequencesExt.SetToSeq(tlc2.value.impl.Value)", 0},
	"SetToSeqs":                            {"public static tlc2.value.impl.Value tlc2.overrides.SequencesExt.SetToSeqs(tlc2.value.impl.Value)", 0},
	"Suffixes":                             {"public static tlc2.value.impl.Value tlc2.overrides.SequencesExt.Suffixes(tlc2.value.impl.Value)", 0},
	"FoldLeft":                             {"public static tlc2.value.impl.Value tlc2.overrides.SequencesExt.foldLeft(tlc2.value.impl.OpValue,tlc2.value.impl.Value,tlc2.value.impl.Value)", 0},
	"FoldLeftDomain":                       {"public static tlc2.value.impl.Value tlc2.overrides.SequencesExt.foldLeftDomain(tlc2.value.impl.OpValue,tlc2.value.impl.Value,tlc2.value.impl.Value)", 0},
	"FoldRight":                            {"public static tlc2.value.impl.Value tlc2.overrides.SequencesExt.foldRight(tlc2.value.impl.OpValue,tlc2.value.impl.Value,tlc2.value.impl.Value)", 0},
	"FoldRightDomain":                      {"public static tlc2.value.impl.Value tlc2.overrides.SequencesExt.foldRightDomain(tlc2.value.impl.OpValue,tlc2.value.impl.Value,tlc2.value.impl.Value)", 0},
	"FoldSeq":                              {"public static tlc2.value.impl.Value tlc2.overrides.SequencesExt.foldSeq(tlc2.value.impl.OpValue,tlc2.value.impl.Value,tlc2.value.impl.Value)", 0},
	"IsPrefix":                             {"public static tlc2.value.impl.Value tlc2.overrides.SequencesExt.isPrefix(tlc2.value.impl.Value,tlc2.value.impl.Value)", 0},
	"LongestCommonPrefix":                  {"public static tlc2.value.impl.Value tlc2.overrides.SequencesExt.longestCommonPrefix(tlc2.value.impl.Value)", 0},
	"RemoveFirst":                          {"public static tlc2.value.impl.Value tlc2.overrides.SequencesExt.removeFirst(tlc2.value.impl.Value,tlc2.value.impl.Value)", 0},
	"RemoveFirstMatch":                     {"public static tlc2.value.impl.Value tlc2.overrides.SequencesExt.removeFirstMatch(tlc2.value.impl.Value,tlc2.value.impl.OpValue)", 0},
	"SelectInSeq":                          {"public static tlc2.value.impl.Value tlc2.overrides.SequencesExt.selectInSeq(tlc2.value.impl.Value,tlc2.value.impl.OpValue)", 0},
	"SelectLastInSeq":                      {"public static tlc2.value.impl.Value tlc2.overrides.SequencesExt.selectLastInSeq(tlc2.value.impl.Value,tlc2.value.impl.OpValue)", 0},
	"ChiSquare":                            {"public static tlc2.value.impl.Value tlc2.overrides.Statistics.chiSquare(tlc2.value.impl.Value,tlc2.value.impl.Value,tlc2.value.impl.Value)", 0},
	"TLCFP":                                {"public static synchronized tlc2.value.impl.IntValue tlc2.module.TLCExt.tlcFingerprint(tlc2.value.impl.Value)", 0},
	"TLCModelValue":                        {"public static synchronized tlc2.value.impl.Value tlc2.module.TLCExt.tlcModelValue(tlc2.value.impl.Value)", 0},
	"ToTrace":                              {"public static tlc2.value.impl.Value tlc2.module.TLCExt.lassoOrdinal(tlc2.value.impl.Value)", 0},
	"TLCGetOrDefault":                      {"public static tlc2.value.impl.Value tlc2.module.TLCExt.tlcGetOrDefault(tlc2.value.impl.Value,tlc2.value.impl.Value)", 0},
	"TLCNoOp":                              {"public static tlc2.value.impl.Value tlc2.module.TLCExt.tlcNoOp(tlc2.value.impl.Value)", 0},
	"_TLCTraceDeserialize":                 {"public static final tlc2.value.IValue tlc2.module._TLCTrace.ioDeserialize(tlc2.value.impl.StringValue) throws java.io.IOException", 0},
	"_TLCTraceSerialize":                   {"public static final tlc2.value.IValue tlc2.module._TLCTrace.ioSerialize(tlc2.value.IValue,tlc2.value.impl.StringValue) throws java.io.IOException", 0},
}
