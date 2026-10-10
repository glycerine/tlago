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
