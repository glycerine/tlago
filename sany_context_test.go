package tlago

import (
	"strings"
	"testing"
)

func TestSanyContextBehaviors(t *testing.T) {
	sym := func(name string, kind sanySemKind, module string) *sanySemSymbolBase {
		return newSanySemSymbol(name, kind, 0, module, Position{File: module, Line: 1, Column: 1})
	}
	names := func(symbols []sanySemSymbol) string {
		parts := make([]string, 0, len(symbols))
		for _, symbol := range symbols {
			parts = append(parts, symbol.semName())
		}
		return strings.Join(parts, ",")
	}

	t.Run("preserves insertion order when merging extended contexts", func(t *testing.T) {
		imported := newSanyContext()
		local := sym("Hidden", sanyUserDefinedOpKind, "Base")
		local.local = true
		imported.addSymbol(local)
		imported.addSymbol(sym("A", sanyUserDefinedOpKind, "Base"))
		imported.addSymbol(sym("B", sanyVariableDeclKind, "Base"))

		dst := newSanyContext()
		dst.addSymbol(sym("Existing", sanyConstantDeclKind, "Root"))
		requireNoErrors(t, dst.mergeExtendContext(imported))

		if got, want := names(dst.orderedSymbols()), "Existing,A,B"; got != want {
			t.Fatalf("merged context order = %s, want %s", got, want)
		}
		if dst.getSymbol("Hidden") != nil {
			t.Fatalf("local symbol from extended context was imported")
		}
	})

	t.Run("keeps module names separate from operator names", func(t *testing.T) {
		ctx := newSanyContext()
		ctx.addSymbol(sym("M", sanyUserDefinedOpKind, "Root"))
		mod := newSanySemModuleNode("M", newSanyContext(), Position{File: "M.tla", Line: 1, Column: 1})
		ctx.addModule(mod)

		if got := ctx.getSymbol("M"); got == nil || got.semKind() != sanyUserDefinedOpKind {
			t.Fatalf("operator namespace lookup = %#v, want user op", got)
		}
		if got := ctx.getModule("M"); got != mod {
			t.Fatalf("module namespace lookup = %#v, want module", got)
		}
	})

	t.Run("resolves symbol stack top before base and modules from external table", func(t *testing.T) {
		base := newSanyContext()
		base.addSymbol(sym("X", sanyConstantDeclKind, "Base"))
		external := newSanyExternalModuleTable()
		extMod := newSanySemModuleNode("Ext", newSanyContext(), Position{File: "Ext.tla", Line: 1, Column: 1})
		external.put("Ext", extMod.context, extMod)

		st := newSanySymbolTable(base, external)
		top := newSanyContext()
		topSym := sym("X", sanyVariableDeclKind, "Top")
		top.addSymbol(topSym)
		st.pushContext(top)

		if got := st.resolveSymbol("X"); got != topSym {
			t.Fatalf("resolveSymbol found %#v, want top context symbol", got)
		}
		if got := st.resolveModule("Ext"); got != extMod {
			t.Fatalf("resolveModule found %#v, want external module", got)
		}
	})

	t.Run("reports Java-shaped merge conflicts", func(t *testing.T) {
		dst := newSanyContext()
		dst.addSymbol(sym("A", sanyUserDefinedOpKind, "Left"))
		dst.addSymbol(sym("B", sanyUserDefinedOpKind, "Left"))

		imported := newSanyContext()
		imported.addSymbol(sym("A", sanyUserDefinedOpKind, "Right"))
		imported.addSymbol(sym("B", sanyVariableDeclKind, "Right"))

		diags := dst.mergeExtendContext(imported)
		if len(diags) != 2 {
			t.Fatalf("merge diagnostics = %d, want 2: %v", len(diags), diags)
		}
		if diags[0].Severity != SeverityWarning || diags[1].Severity != SeverityError {
			t.Fatalf("merge severities = %s,%s; want warning,error", diags[0].Severity, diags[1].Severity)
		}
	})

	t.Run("suppresses duplicates from the same original module", func(t *testing.T) {
		dst := newSanyContext()
		dst.addSymbol(sym("A", sanyUserDefinedOpKind, "Shared"))
		imported := newSanyContext()
		imported.addSymbol(sym("A", sanyUserDefinedOpKind, "Shared"))

		requireNoErrors(t, dst.mergeExtendContext(imported))
		if got, want := names(dst.orderedSymbols()), "A"; got != want {
			t.Fatalf("same-origin merge order = %s, want %s", got, want)
		}
	})
}
