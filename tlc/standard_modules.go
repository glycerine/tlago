package tlc

import "reflect"

var sanyStandardModuleNames = map[string]struct{}{
	"FiniteSets":    {},
	"Sequences":     {},
	"Bags":          {},
	"Naturals":      {},
	"Integers":      {},
	"Reals":         {},
	"RealTime":      {},
	"Randomization": {},
	"TLC":           {},
}

// IsDefinedInStandardModuleName reports whether name belongs to SANY's fixed
// StandardModules set.
func IsDefinedInStandardModuleName(name string) bool {
	_, ok := sanyStandardModuleNames[name]
	return ok
}

// IsDefinedInStandardModule applies StandardModules' semantic-node overload.
// Java returns false for a null node and classifies a node by its source name.
func IsDefinedInStandardModule(node interface{ GetSourceLocation() SourceLocation }) bool {
	if node == nil {
		return false
	}
	value := reflect.ValueOf(node)
	if value.Kind() == reflect.Pointer && value.IsNil() {
		return false
	}
	return IsDefinedInStandardModuleName(node.GetSourceLocation().Source)
}

// FilterNonStandardModulesFromSet mutates names to retain only SANY standard
// modules, matching StandardModules.filterNonStandardModulesFromSet.
func FilterNonStandardModulesFromSet(names map[string]struct{}) {
	if names == nil {
		panic(NewNullPointerException())
	}
	for name := range names {
		if !IsDefinedInStandardModuleName(name) {
			delete(names, name)
		}
	}
}
