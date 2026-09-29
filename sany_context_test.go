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

func TestSanyBuiltInContextBehaviors(t *testing.T) {
	t.Run("initial context contains Java SANY built-in properties", func(t *testing.T) {
		ctx := newSanyInitialContext()
		for _, tc := range []struct {
			name  string
			arity int
			level tlaLevel
		}{
			{name: "TRUE", arity: 0, level: constantLevel},
			{name: "ENABLED", arity: 1, level: variableLevel},
			{name: "$Nop", arity: 1, level: constantLevel},
			{name: "$Witness", arity: -1, level: constantLevel},
			{name: "$WF", arity: 2, level: temporalLevel},
		} {
			symbol, ok := ctx.getSymbol(tc.name).(*sanySemBuiltInSymbol)
			if !ok {
				t.Fatalf("builtin %s = %#v, want semantic built-in", tc.name, ctx.getSymbol(tc.name))
			}
			if symbol.semArity() != tc.arity || symbol.level != tc.level {
				t.Fatalf("builtin %s arity/level = %d/%d, want %d/%d", tc.name, symbol.semArity(), symbol.level, tc.arity, tc.level)
			}
		}
	})

	t.Run("initial contexts are fresh per spec", func(t *testing.T) {
		first := newSanyInitialContext()
		second := newSanyInitialContext()
		first.addSymbol(newSanySemSymbol("Extra", sanyUserDefinedOpKind, 0, "First", Position{File: "First.tla", Line: 1, Column: 1}))
		if second.getSymbol("Extra") != nil {
			t.Fatalf("fresh initial context observed a symbol added to another context")
		}
		if first.getSymbol("TRUE") == second.getSymbol("TRUE") {
			t.Fatalf("built-in symbols are shared across fresh initial contexts")
		}
	})

	t.Run("XML built-in lookup uses Java SANY properties", func(t *testing.T) {
		for _, tc := range []struct {
			name  string
			arity int
			level tlaLevel
		}{
			{name: "$Nop", arity: 1, level: constantLevel},
			{name: "$Take", arity: 1, level: constantLevel},
			{name: "$Witness", arity: -1, level: constantLevel},
			{name: "\\X", arity: -1, level: constantLevel},
		} {
			info := sanyXMLBuiltin(tc.name)
			if info.arity != tc.arity || info.level != tc.level {
				t.Fatalf("XML builtin %s arity/level = %d/%d, want %d/%d", tc.name, info.arity, info.level, tc.arity, tc.level)
			}
		}
	})
}
