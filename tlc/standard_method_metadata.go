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
