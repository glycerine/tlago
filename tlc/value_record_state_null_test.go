package tlc

import "testing"

func TestRecordValueStateConstructorNullFailuresMatchJava(t *testing.T) {
	assertFailure := func(name, want string, body func()) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			defer func() {
				failure := recover()
				null, ok := failure.(*NullPointerException)
				if !ok || null.Error() != want {
					t.Fatalf("failure = %T/%v, want NullPointerException %q", failure, failure, want)
				}
			}()
			body()
		})
	}

	stateMessage := `Cannot invoke "tlc2.tool.TLCState.getVars()" because "state" is null`
	assertFailure("state", stateMessage, func() { NewRecordValueFromState(nil) })
	assertFailure("state with default", stateMessage, func() { NewRecordValueFromState(nil, IntZero) })
	assertFailure("state with action", stateMessage, func() { NewRecordValueFromStateAndAction(nil, nil) })
	assertFailure("action", `Cannot invoke "tlc2.tool.Action.getParameters()" because "action" is null`, func() {
		NewRecordValueFromStateAndAction(NewEmptyState(), nil)
	})
	var action *Action
	actionMessage := `Cannot invoke "tlc2.tool.Action.getParameters()" because "action" is null`
	assertFailure("action record", actionMessage, func() { action.ToRecordValue() })
	assertFailure("action record with extra field", actionMessage, func() {
		NewRecordValueFromActionField(action, UniqueStringOf("x"), IntOne)
	})
	assertFailure("map record", `Cannot invoke "java.util.Map.entrySet()" because "m" is null`, func() {
		NewRecordValueFromInsMap(nil)
	})
	assertFailure("operator definition record", `Cannot invoke "tla2sany.semantic.OpDefNode.getName()" because "odn" is null`, func() {
		NewRecordValueFromOpDef(nil)
	})
	assertFailure("paired state", stateMessage, func() {
		NewRecordValueFromStates(nil, NewEmptyState(), IntZero)
	})
	assertFailure("paired successor", `Cannot invoke "tlc2.tool.TLCState.getVars()" because "successor" is null`, func() {
		NewRecordValueFromStates(NewEmptyState(), nil, IntZero)
	})

	oldVariables, oldDeclarations, oldEmpty := stateVariables, stateVariableDeclarations, EmptyState
	oldCount := UniqueStringVariableCount()
	defer func() {
		stateVariables, stateVariableDeclarations, EmptyState = oldVariables, oldDeclarations, oldEmpty
		SetUniqueStringVariableCount(oldCount)
	}()
	SetStateVariables([]string{"x"})
	stateVariableDeclarations = []*SymbolNode{nil}
	variableMessage := `Cannot invoke "tla2sany.semantic.OpDeclNode.getName()" because "vars[i]" is null`
	assertFailure("null variable", variableMessage, func() { NewRecordValueFromState(NewEmptyState()) })
	assertFailure("paired null variable", variableMessage, func() {
		NewRecordValueFromStates(NewEmptyState(), NewEmptyState(), IntZero)
	})
}
