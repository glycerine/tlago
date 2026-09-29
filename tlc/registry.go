package tlc

import "strings"

var tlaRegistry = NewInsMap[string, string]()

func init() {
	RegisterStandardTLAOperators()
}

func RegisterStandardTLAOperators() {
	TLARegistryPut("Plus", "+")
	TLARegistryPut("Minus", "-")
	TLARegistryPut("Times", "*")
	TLARegistryPut("LT", "<")
	TLARegistryPut("LE", "\\leq")
	TLARegistryPut("GT", ">")
	TLARegistryPut("GEQ", "\\geq")
	TLARegistryPut("DotDot", "..")
	TLARegistryPut("Neg", "-.")
	TLARegistryPut("Divide", "\\div")
	TLARegistryPut("Mod", "%")
	TLARegistryPut("Expt", "^")
	TLARegistryPut("Concat", "\\o")
	TLARegistryPut("MakeFcn", ":>")
	TLARegistryPut("CombineFcn", "@@")
}

func TLARegistryGet(name string) (string, bool) {
	return tlaRegistry.Get2(name)
}

func TLARegistryPut(tlaName string, goName string) (previous string, replaced bool) {
	previous, replaced = tlaRegistry.Get2(tlaName)
	tlaRegistry.Set(tlaName, goName)
	return previous, replaced
}

func TLARegistryMapName(name string) string {
	if mapped, ok := TLARegistryGet(name); ok {
		return mapped
	}
	return name
}

func TLARegistryAllNames() string {
	var names []string
	for name := range tlaRegistry.All() {
		names = append(names, name)
	}
	return "{" + strings.Join(names, ", ") + "}"
}
