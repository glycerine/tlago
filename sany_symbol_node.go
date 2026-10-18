package tlago

import (
	"reflect"

	"github.com/glycerine/tlago/tlc"
)

func sanyRuntimeCanonicalLevelNode(node tlc.SemanticNode) sanyCanonicalLevelNode {
	source := sanyRuntimeLevelData(node).runtimeSource
	if source == nil || reflect.ValueOf(source).IsNil() {
		panic(tlc.NewUnsupportedOperationException("canonical SANY level node has not been installed"))
	}
	return source
}

func sanyAppendPublicMatchDiagnostics(errors *SanyErrors, diagnostics Diagnostics) {
	if len(diagnostics) == 0 {
		return
	}
	if errors == nil {
		panic(tlc.NewNullPointerException())
	}
	for _, diagnostic := range diagnostics {
		diagnostic.sanyDetails = sanyErrorDetailsForDiagnostic(diagnostic)
		errors.diagnostics = appendSanyDiagnostics(errors.diagnostics, diagnostic)
	}
}

// SanySymbolNodeMatch adapts SymbolNode.match(OpApplNode, ModuleNode, Errors)
// over parser-owned canonical symbols. The module argument is retained for the
// Java signature; each source implementation ignores it.
func SanySymbolNodeMatch(symbol tlc.SemanticNode, application *tlc.OpApplNode, module *tlc.ModuleNode, errors *SanyErrors) bool {
	_ = module
	source := sanyRuntimeCanonicalLevelNode(symbol)
	if _, ok := source.(*sanySemModuleNode); ok {
		return false
	}
	applicationSource, ok := sanyRuntimeCanonicalLevelNode(application).(*sanySemOpApplNode)
	if !ok || applicationSource == nil {
		panic(tlc.NewClassCastException("application is not a canonical OpApplNode"))
	}
	switch symbol := source.(type) {
	case *sanyFormalParamNode:
		return symbol.match(applicationSource)
	case *sanySemThmOrAssumpDefNode:
		return symbol.match(applicationSource)
	case *sanySemOpDeclNode:
		matched, diagnostics, _ := symbol.match(applicationSource)
		sanyAppendPublicMatchDiagnostics(errors, diagnostics)
		return matched
	case *sanySemOpDefNode:
		matched, diagnostics, failure := symbol.match(applicationSource)
		sanyAppendPublicMatchDiagnostics(errors, diagnostics)
		if failure != nil {
			last := diagnostics[len(diagnostics)-1]
			abort := newSanySemanticAbort(last, errors.diagnostics, &errors.diagnostics)
			abort.sourceErrors = errors
			panic(abort)
		}
		return matched
	default:
		panic(tlc.NewClassCastException("semantic node is not a SymbolNode"))
	}
}

// SanyFormalParamNodeMatchSemantic adapts FormalParamNode.match(SemanticNode).
// The source overload intentionally ignores its argument.
func SanyFormalParamNodeMatchSemantic(symbol *tlc.SymbolNode, test tlc.SemanticNode) bool {
	_ = test
	formal, ok := sanyRuntimeCanonicalLevelNode(symbol).(*sanyFormalParamNode)
	if !ok || formal == nil {
		panic(tlc.NewClassCastException("semantic node is not a FormalParamNode"))
	}
	return formal.semArity() == 0
}

// SanyOpDefOrDeclNodeGetSymbolTable adapts
// OpDefOrDeclNode.getSymbolTable(). Definitions and declarations generated
// under the same source table share the same public adapter.
func SanyOpDefOrDeclNodeGetSymbolTable(node tlc.SemanticNode) *SanySymbolTable {
	if sanyExploreNull(node) {
		panic(tlc.NewNullPointerException())
	}
	var value any
	switch node := node.(type) {
	case *tlc.SymbolNode:
		value = node.SANYSymbolTable
	case *tlc.OpDefNode:
		value = node.SANYSymbolTable
	default:
		panic(tlc.NewClassCastException("semantic node is not an OpDefOrDeclNode"))
	}
	if value == nil {
		return nil
	}
	table, ok := value.(*SanySymbolTable)
	if !ok {
		panic(tlc.NewClassCastException("invalid SANY SymbolTable adapter"))
	}
	return table
}

// SanyModuleNodeGetSymbolTable adapts ModuleNode.getSymbolTable().
func SanyModuleNodeGetSymbolTable(node *tlc.ModuleNode) *SanySymbolTable {
	if node == nil {
		panic(tlc.NewNullPointerException())
	}
	return nil
}

// SanyModuleNodeMatch adapts ModuleNode.match(OpApplNode, ModuleNode, Errors).
// The source method is a literal false stub and does not inspect its arguments.
func SanyModuleNodeMatch(node *tlc.ModuleNode, _ *tlc.OpApplNode, _ *tlc.ModuleNode, _ *SanyErrors) bool {
	if node == nil {
		panic(tlc.NewNullPointerException())
	}
	return false
}

// SanyModuleNodeIsConstant adapts ModuleNode.isConstant(Errors) through the
// parser-owned level checker.
func SanyModuleNodeIsConstant(node *tlc.ModuleNode, errors *SanyErrors) bool {
	source, ok := sanyRuntimeCanonicalLevelNode(node).(*sanySemModuleNode)
	if !ok || source == nil {
		panic(tlc.NewClassCastException("semantic node is not a canonical ModuleNode"))
	}
	var diagnostics *Diagnostics
	if errors != nil {
		diagnostics = &errors.diagnostics
	}
	return source.isConstant(diagnostics)
}
