package tlago

import (
	"io"
	"os"

	"github.com/glycerine/tlago/tlc"
)

// SanyExternalModuleTable adapts tla2sany.semantic.ExternalModuleTable.
type SanyExternalModuleTable = tlc.ExternalModuleTable

// SanyExternalModuleTableEntry adapts the table's public nested entry class.
type SanyExternalModuleTableEntry = tlc.ExternalModuleTableEntry

func NewSanyExternalModuleTable() *SanyExternalModuleTable {
	return tlc.NewExternalModuleTable()
}

// SanyExternalModuleTableWalkGraph adapts both ExternalModuleTable.walkGraph
// overloads. Omitting visitor uses Java's no-op visitor.
func SanyExternalModuleTableWalkGraph(table *SanyExternalModuleTable, nodes map[int32]any, visitors ...*SanyExplorerVisitor) {
	if len(visitors) > 1 {
		panic(tlc.NewIllegalArgumentException("ExternalModuleTable.walkGraph accepts at most one visitor"))
	}
	visitor := &SanyExplorerVisitor{}
	if len(visitors) == 1 {
		visitor = visitors[0]
	}
	tlc.WalkSemanticGraph(table, nodes, visitor)
}

// SanyExternalModuleTableEntryWalkGraph adapts
// ExternalModuleTableEntry.walkGraph(Hashtable, ExplorerVisitor).
func SanyExternalModuleTableEntryWalkGraph(entry *SanyExternalModuleTableEntry, nodes map[int32]any, visitor *SanyExplorerVisitor) {
	tlc.WalkSemanticGraph(entry, nodes, visitor)
}

// SanyExternalModuleTableEntryToString adapts
// ExternalModuleTableEntry.toString(int, Errors).
func SanyExternalModuleTableEntryToString(entry *SanyExternalModuleTableEntry, depth int, errors *SanyErrors) string {
	if entry == nil {
		panic(tlc.NewNullPointerException())
	}
	if depth <= 0 {
		return ""
	}
	module := "\n***Null ExternalModuleTable entry; module contained error and was not created."
	if entry.ModuleNode != nil {
		module = SanyModuleNodeToString(entry.ModuleNode, depth, errors)
	}
	return sanyIndent(2, "\nModule: "+sanyIndent(2, module))
}

// SanyExternalModuleTableToString adapts
// ExternalModuleTable.toString(int, Errors), which uses dependency insertion
// order rather than Hashtable enumeration order.
func SanyExternalModuleTableToString(table *SanyExternalModuleTable, depth int, errors *SanyErrors) string {
	if table == nil {
		panic(tlc.NewNullPointerException())
	}
	if depth <= 0 {
		return ""
	}
	result := ""
	for _, module := range table.ModuleNodeVector {
		entry := &SanyExternalModuleTableEntry{ModuleNode: module}
		result += SanyExternalModuleTableEntryToString(entry, depth, errors)
	}
	return result
}

// SanyExternalModuleTableEntryPrintMTE adapts
// ExternalModuleTableEntry.printMTE(int, boolean, Errors).
func SanyExternalModuleTableEntryPrintMTE(entry *SanyExternalModuleTableEntry, depth int, includeInitial bool, errors *SanyErrors) {
	SanyExternalModuleTableEntryPrintMTETo(os.Stdout, entry, depth, includeInitial, errors)
}

// SanyExternalModuleTableEntryPrintMTETo retains printMTE's formatting while
// allowing an owned writer at the Go boundary.
func SanyExternalModuleTableEntryPrintMTETo(out io.Writer, entry *SanyExternalModuleTableEntry, depth int, includeInitial bool, errors *SanyErrors) {
	if entry == nil || out == nil {
		panic(tlc.NewNullPointerException())
	}
	_, _ = io.WriteString(out, sanyIndent(2, "\nModule: "))
	if entry.ModuleNode != nil {
		SanyModuleNodePrintTo(out, entry.ModuleNode, 2, depth, includeInitial, errors)
		return
	}
	_, _ = io.WriteString(out, sanyIndent(4, "\n***Null ExternalModuleTable entry; module contained error and was not created."))
}

// SanyExternalModuleTablePrintExternalModuleTable adapts
// ExternalModuleTable.printExternalModuleTable(int, boolean, Errors).
func SanyExternalModuleTablePrintExternalModuleTable(table *SanyExternalModuleTable, depth int, includeInitial bool, errors *SanyErrors) {
	SanyExternalModuleTablePrintExternalModuleTableTo(os.Stdout, table, depth, includeInitial, errors)
}

// SanyExternalModuleTablePrintExternalModuleTableTo retains
// printExternalModuleTable's formatting while allowing an owned writer at the
// Go boundary.
func SanyExternalModuleTablePrintExternalModuleTableTo(out io.Writer, table *SanyExternalModuleTable, depth int, includeInitial bool, errors *SanyErrors) {
	if table == nil || out == nil {
		panic(tlc.NewNullPointerException())
	}
	_, _ = io.WriteString(out, "\nExternal Module Table:")
	for _, module := range table.ModuleNodeVector {
		entry := &SanyExternalModuleTableEntry{ModuleNode: module}
		SanyExternalModuleTableEntryPrintMTETo(out, entry, depth, includeInitial, errors)
	}
}
