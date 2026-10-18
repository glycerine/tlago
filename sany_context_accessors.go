package tlago

import (
	"github.com/glycerine/tlago/tlc"
)

// SanyExplorerVisitor adapts tla2sany.explorer.ExplorerVisitor for public
// semantic graph traversal. Contexts and substitutions are included in its
// callbacks even though they do not have semantic UIDs.
type SanyExplorerVisitor = tlc.SemanticExplorerVisitor

// SanyContextGetByClass returns current Hashtable bindings assignable to T in
// the same order as Context.getByClass(Class).
func SanyContextGetByClass[T tlc.SemanticNode](context *SanyContext) []T {
	if context == nil {
		panic(tlc.NewNullPointerException())
	}
	result := make([]T, 0)
	entries := context.GetContextSymbolEnumeration()
	for entries.HasMoreElements() {
		if node, ok := entries.NextElement().(T); ok {
			result = append(result, node)
		}
	}
	return result
}

func SanyContextLevelDataToString(context *SanyContext) string {
	if context == nil {
		panic(tlc.NewNullPointerException())
	}
	return "Dummy level string"
}

// SanyContextToString adapts Context.toString(int, Errors), whose arguments
// are deliberately unused by the source implementation.
func SanyContextToString(context *SanyContext, depth int, errors *SanyErrors) string {
	if context == nil {
		panic(tlc.NewNullPointerException())
	}
	return "Please use Context.getContextEntryStringVector() instead of Context.toString()"
}

// SanyContextGetContextEntryStringVector adapts
// Context.getContextEntryStringVector(int, boolean, Errors). The returned
// slice is a fresh vector in the source's oldest-to-newest Pair-history order.
func SanyContextGetContextEntryStringVector(context *SanyContext, depth int, includeInitial bool, errors *SanyErrors) []string {
	return sanyContextEntryStrings(context, depth, includeInitial, errors)
}

// SanyContextWalkGraph adapts Context.walkGraph(Hashtable, ExplorerVisitor).
func SanyContextWalkGraph(context *SanyContext, table map[int32]any, visitor *SanyExplorerVisitor) {
	tlc.WalkSemanticGraph(context, table, visitor)
}

func sanyRuntimeContextNodeKind(node tlc.SemanticNode) string {
	if _, ok := node.(*tlc.OpDefNode); ok {
		return "definition"
	}
	if symbol, ok := node.(*tlc.SymbolNode); ok && symbol.Kind == tlc.SymbolFormalParam {
		return "definition"
	}
	return "declaration"
}

func sanyRuntimeContextNodeLocation(node tlc.SemanticNode) *tlc.SourceLocation {
	located, ok := node.(interface{ GetSourceLocation() tlc.SourceLocation })
	if !ok {
		panic(tlc.NewClassCastException("context entry is not a SemanticNode"))
	}
	location := located.GetSourceLocation()
	return &location
}

// SanyContextMergeExtendContext adapts Context.mergeExtendContext and records
// its warning/error diagnostics in errors.
func SanyContextMergeExtendContext(context, imported *SanyContext, errors *SanyErrors) bool {
	success, conflicts := context.MergeExtendContext(imported)
	for _, conflict := range conflicts {
		codeValue := 4224
		format := "The %s of '%s' conflicts with \nits %s at %s."
		incomingKind := sanyRuntimeContextNodeKind(conflict.Incoming)
		existingKind := sanyRuntimeContextNodeKind(conflict.Existing)
		if conflict.SameClass {
			codeValue = 4800
			format = "Warning: the %s of '%s' conflicts with \nits %s at %s."
			incomingKind = existingKind
		}
		code := SanyErrorCodeFromStandardValue(codeValue)
		errors.AddMessage(code, sanyRuntimeContextNodeLocation(conflict.Incoming), format,
			incomingKind, conflict.Name.String(), existingKind, sanyRuntimeContextNodeLocation(conflict.Existing).String())
	}
	return success
}
