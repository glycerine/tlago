/*******************************************************************************
 * Copyright (c) 2025 Linux Foundation. All rights reserved.
 *
 * The MIT License (MIT)
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
 * of the Software, and to permit persons to whom the Software is furnished to do
 * so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN
 * AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
 * WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 ******************************************************************************/
package tlc

import "strings"

// Read-only views retain their backing set, as MP's Collections.unmodifiableSet
// does. Reset empties the sets without changing the views' identities.
type MessageCodeSetView struct{ codes *InsMap[int, bool] }

func (v MessageCodeSetView) Contains(code int) bool {
	Globals.Lock()
	defer Globals.Unlock()
	_, ok := v.codes.Get2(code)
	return ok
}
func (v MessageCodeSetView) IsEmpty() bool {
	Globals.Lock()
	defer Globals.Unlock()
	return v.codes.Len() == 0
}
func GetTLCSuppressedCodes() MessageCodeSetView {
	Globals.Lock()
	defer Globals.Unlock()
	return MessageCodeSetView{Globals.SuppressedMessages}
}
func GetTLCMessagesAsErrorCodes() MessageCodeSetView {
	Globals.Lock()
	defer Globals.Unlock()
	return MessageCodeSetView{Globals.MessagesAsErrors}
}
func GetSANYSuppressedCodes() MessageCodeSetView {
	Globals.Lock()
	defer Globals.Unlock()
	return MessageCodeSetView{Globals.SANYSuppressedMessages}
}
func GetSANYMessagesAsErrorCodes() MessageCodeSetView {
	Globals.Lock()
	defer Globals.Unlock()
	return MessageCodeSetView{Globals.SANYMessagesAsErrors}
}
func ResetMessageControl() {
	Globals.Lock()
	defer Globals.Unlock()
	Globals.SuppressedMessages.DeleteAll()
	Globals.MessagesAsErrors.DeleteAll()
	Globals.SANYSuppressedMessages.DeleteAll()
	Globals.SANYMessagesAsErrors.DeleteAll()
}
func SuppressSANYMessage(code int) {
	Globals.Lock()
	defer Globals.Unlock()
	Globals.SANYSuppressedMessages.Set(code, true)
}
func TreatSANYMessageAsError(code int) {
	Globals.Lock()
	defer Globals.Unlock()
	Globals.SANYMessagesAsErrors.Set(code, true)
}

// Snapshots let the parser bridge apply the very same controls as direct TLC
// parameter handling, without depending on root-package diagnostic types.
func SANYMessageControls() (suppressed, elevated []int) {
	Globals.Lock()
	defer Globals.Unlock()
	for c := range Globals.SANYSuppressedMessages.All() {
		suppressed = append(suppressed, c)
	}
	for c := range Globals.SANYMessagesAsErrors.All() {
		elevated = append(elevated, c)
	}
	return
}

var sanyMessageCodeNames = map[int]string{
	1000: "GENERAL",
	4003: "INTERNAL_ERROR",
	4004: "SUSPECTED_UNREACHABLE_CHECK",
	4005: "UNSUPPORTED_LANGUAGE_FEATURE",
	4200: "SYMBOL_UNDEFINED",
	4201: "SYMBOL_REDEFINED",
	4202: "BUILT_IN_SYMBOL_REDEFINED",
	4203: "OPERATOR_NAME_INCOMPLETE",
	4204: "OPERATOR_GIVEN_INCORRECT_NUMBER_OF_ARGUMENTS",
	4205: "OPERATOR_LEVEL_CONSTRAINTS_EXCEEDED",
	4206: "ASSUMPTION_IS_NOT_CONSTANT",
	4220: "MODULE_FILE_CANNOT_BE_FOUND",
	4221: "MODULE_NAME_DIFFERENT_FROM_FILE_NAME",
	4222: "MODULE_DEPENDENCIES_ARE_CIRCULAR",
	4223: "MODULE_REDEFINED",
	4224: "EXTENDED_MODULES_SYMBOL_UNIFICATION_CONFLICT",
	4240: "INSTANCE_SUBSTITUTION_MISSING_SYMBOL",
	4241: "INSTANCE_SUBSTITUTION_SYMBOL_REDEFINED_MULTIPLE_TIMES",
	4242: "INSTANCE_SUBSTITUTION_ILLEGAL_SYMBOL_REDEFINITION",
	4243: "INSTANCE_SUBSTITUTION_OPERATOR_CONSTANT_INCORRECT_ARITY",
	4244: "INSTANCE_SUBSTITUTION_NON_LEIBNIZ_OPERATOR",
	4245: "INSTANCE_SUBSTITUTION_LEVEL_CONSTRAINTS_EXCEEDED",
	4246: "INSTANCE_SUBSTITUTION_LEVEL_CONSTRAINT_NOT_MET",
	4247: "INSTANCE_SUBSTITUTION_COPARAMETER_LEVEL_CONSTRAINTS_EXCEEDED",
	4260: "FUNCTION_GIVEN_INCORRECT_NUMBER_OF_ARGUMENTS",
	4261: "FUNCTION_EXCEPT_AT_USED_WHERE_UNDEFINED",
	4262: "RECORD_CONSTRUCTOR_FIELD_REDEFINITION",
	4270: "HIGHER_ORDER_OPERATOR_REQUIRED_BUT_EXPRESSION_GIVEN",
	4271: "HIGHER_ORDER_OPERATOR_ARGUMENT_HAS_INCORRECT_ARITY",
	4272: "HIGHER_ORDER_OPERATOR_PARAMETER_LEVEL_CONSTRAINT_NOT_MET",
	4273: "HIGHER_ORDER_OPERATOR_COPARAMETER_LEVEL_CONSTRAINTS_EXCEEDED",
	4274: "LAMBDA_OPERATOR_ARGUMENT_HAS_INCORRECT_ARITY",
	4275: "LAMBDA_GIVEN_WHERE_EXPRESSION_REQUIRED",
	4290: "RECURSIVE_OPERATOR_PRIMES_PARAMETER",
	4291: "RECURSIVE_OPERATOR_DECLARED_BUT_NOT_DEFINED",
	4292: "RECURSIVE_OPERATOR_DECLARATION_DEFINITION_ARITY_MISMATCH",
	4293: "RECURSIVE_OPERATOR_DEFINED_IN_WRONG_LET_IN_LEVEL",
	4294: "RECURSIVE_SECTION_CONTAINS_ILLEGAL_DEFINITION",
	4310: "ALWAYS_PROPERTY_SENSITIVE_TO_STUTTERING",
	4311: "EVENTUALLY_PROPERTY_SENSITIVE_TO_STUTTERING",
	4312: "BINARY_TEMPORAL_OPERATOR_WITH_ACTION_LEVEL_PARAMETER",
	4313: "LOGICAL_OPERATOR_WITH_MIXED_ACTION_TEMPORAL_PARAMETERS",
	4314: "QUANTIFIED_TEMPORAL_FORMULA_WITH_ACTION_LEVEL_BOUND",
	4315: "QUANTIFICATION_WITH_TEMPORAL_LEVEL_BOUND",
	4330: "LABEL_PARAMETER_REPETITION",
	4331: "LABEL_PARAMETER_MISSING",
	4332: "LABEL_PARAMETER_UNNECESSARY",
	4333: "LABEL_NOT_IN_DEFINITION_OR_PROOF_STEP",
	4334: "LABEL_NOT_ALLOWED_IN_NESTED_ASSUME_PROVE_WITH_NEW",
	4335: "LABEL_NOT_ALLOWED_IN_FUNCTION_EXCEPT",
	4336: "LABEL_REDEFINITION",
	4337: "LABEL_GIVEN_INCORRECT_NUMBER_OF_ARGUMENTS",
	4350: "PROOF_STEP_WITH_IMPLICIT_LEVEL_CANNOT_HAVE_NAME",
	4351: "PROOF_STEP_NON_EXPRESSION_USED_AS_EXPRESSION",
	4352: "TEMPORAL_PROOF_GOAL_WITH_NON_CONSTANT_TAKE_WITNESS_HAVE",
	4353: "TEMPORAL_PROOF_GOAL_WITH_NON_CONSTANT_CASE",
	4354: "QUANTIFIED_TEMPORAL_PICK_FORMULA_WITH_NON_CONSTANT_BOUND",
	4355: "ASSUME_PROVE_USED_WHERE_EXPRESSION_REQUIRED",
	4356: "ASSUME_PROVE_NEW_CONSTANT_HAS_TEMPORAL_LEVEL_BOUND",
	4357: "USE_OR_HIDE_FACT_NOT_VALID",
	4800: "EXTENDED_MODULES_SYMBOL_UNIFICATION_AMBIGUITY",
	4801: "INSTANCED_MODULES_SYMBOL_UNIFICATION_AMBIGUITY",
	4802: "RECORD_CONSTRUCTOR_FIELD_NAME_CLASH",
	4803: "PLUSCAL_ALGORITHM_AND_TRANSLATION_CHANGED_SINCE_LAST_TRANSLATION",
	4804: "PLUSCAL_ALGORITHM_CHANGED_SINCE_LAST_TRANSLATION",
	4805: "PLUSCAL_TRANSLATION_CHANGED_SINCE_LAST_TRANSLATION",
}

func sanyMessageCodeOverlap(a, b *InsMap[int, bool]) string {
	var names []string
	for code := range a.All() {
		if _, ok := b.Get2(code); ok {
			names = append(names, sanyMessageCodeNames[code])
		}
	}
	if len(names) == 0 {
		return ""
	}
	return "[" + strings.Join(names, ", ") + "]"
}
func (t *TLC) printWelcome() {
	if t.WelcomePrinted {
		return
	}
	t.WelcomePrinted = true
	version := TLCVersion()
	if rev := TLCRevision(); rev != "" {
		version += " (rev: " + rev + ")"
	}
	PrintMessage(ECTLCVersion, version)
}
