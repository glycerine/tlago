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
package tlago

// SanyErrorCode is the metadata of tla2sany.semantic.ErrorCode. ParameterCount
// is -1 for VARIADIC_PARAMETERS; other values specify the exact argument count.
type SanyErrorCode struct {
	Name           string
	Value          int
	Severity       Severity
	ParameterCount int
}

// SanyErrorCodeFromStandardValue resolves the original standardized value.
func SanyErrorCodeFromStandardValue(value int) (SanyErrorCode, bool) {
	code, ok := sanyErrorCodes[value]
	return code, ok
}

var sanyErrorCodes = map[int]SanyErrorCode{
	1000: {Name: "GENERAL", Value: 1000, Severity: SeverityError, ParameterCount: -1},
	4003: {Name: "INTERNAL_ERROR", Value: 4003, Severity: SeverityError, ParameterCount: -1},
	4004: {Name: "SUSPECTED_UNREACHABLE_CHECK", Value: 4004, Severity: SeverityError, ParameterCount: -1},
	4005: {Name: "UNSUPPORTED_LANGUAGE_FEATURE", Value: 4005, Severity: SeverityError, ParameterCount: -1},
	4200: {Name: "SYMBOL_UNDEFINED", Value: 4200, Severity: SeverityError, ParameterCount: -1},
	4201: {Name: "SYMBOL_REDEFINED", Value: 4201, Severity: SeverityError, ParameterCount: -1},
	4202: {Name: "BUILT_IN_SYMBOL_REDEFINED", Value: 4202, Severity: SeverityError, ParameterCount: 1},
	4203: {Name: "OPERATOR_NAME_INCOMPLETE", Value: 4203, Severity: SeverityError, ParameterCount: 1},
	4204: {Name: "OPERATOR_GIVEN_INCORRECT_NUMBER_OF_ARGUMENTS", Value: 4204, Severity: SeverityError, ParameterCount: 2},
	4205: {Name: "OPERATOR_LEVEL_CONSTRAINTS_EXCEEDED", Value: 4205, Severity: SeverityError, ParameterCount: -1},
	4206: {Name: "ASSUMPTION_IS_NOT_CONSTANT", Value: 4206, Severity: SeverityError, ParameterCount: 1},
	4220: {Name: "MODULE_FILE_CANNOT_BE_FOUND", Value: 4220, Severity: SeverityError, ParameterCount: -1},
	4221: {Name: "MODULE_NAME_DIFFERENT_FROM_FILE_NAME", Value: 4221, Severity: SeverityError, ParameterCount: 2},
	4222: {Name: "MODULE_DEPENDENCIES_ARE_CIRCULAR", Value: 4222, Severity: SeverityError, ParameterCount: -1},
	4223: {Name: "MODULE_REDEFINED", Value: 4223, Severity: SeverityError, ParameterCount: 2},
	4224: {Name: "EXTENDED_MODULES_SYMBOL_UNIFICATION_CONFLICT", Value: 4224, Severity: SeverityError, ParameterCount: 4},
	4240: {Name: "INSTANCE_SUBSTITUTION_MISSING_SYMBOL", Value: 4240, Severity: SeverityError, ParameterCount: 3},
	4241: {Name: "INSTANCE_SUBSTITUTION_SYMBOL_REDEFINED_MULTIPLE_TIMES", Value: 4241, Severity: SeverityError, ParameterCount: 1},
	4242: {Name: "INSTANCE_SUBSTITUTION_ILLEGAL_SYMBOL_REDEFINITION", Value: 4242, Severity: SeverityError, ParameterCount: 1},
	4243: {Name: "INSTANCE_SUBSTITUTION_OPERATOR_CONSTANT_INCORRECT_ARITY", Value: 4243, Severity: SeverityError, ParameterCount: 2},
	4244: {Name: "INSTANCE_SUBSTITUTION_NON_LEIBNIZ_OPERATOR", Value: 4244, Severity: SeverityError, ParameterCount: 2},
	4245: {Name: "INSTANCE_SUBSTITUTION_LEVEL_CONSTRAINTS_EXCEEDED", Value: 4245, Severity: SeverityError, ParameterCount: 3},
	4246: {Name: "INSTANCE_SUBSTITUTION_LEVEL_CONSTRAINT_NOT_MET", Value: 4246, Severity: SeverityError, ParameterCount: 4},
	4247: {Name: "INSTANCE_SUBSTITUTION_COPARAMETER_LEVEL_CONSTRAINTS_EXCEEDED", Value: 4247, Severity: SeverityError, ParameterCount: 4},
	4260: {Name: "FUNCTION_GIVEN_INCORRECT_NUMBER_OF_ARGUMENTS", Value: 4260, Severity: SeverityError, ParameterCount: 3},
	4261: {Name: "FUNCTION_EXCEPT_AT_USED_WHERE_UNDEFINED", Value: 4261, Severity: SeverityError, ParameterCount: 0},
	4262: {Name: "RECORD_CONSTRUCTOR_FIELD_REDEFINITION", Value: 4262, Severity: SeverityError, ParameterCount: 0},
	4270: {Name: "HIGHER_ORDER_OPERATOR_REQUIRED_BUT_EXPRESSION_GIVEN", Value: 4270, Severity: SeverityError, ParameterCount: 2},
	4271: {Name: "HIGHER_ORDER_OPERATOR_ARGUMENT_HAS_INCORRECT_ARITY", Value: 4271, Severity: SeverityError, ParameterCount: -1},
	4272: {Name: "HIGHER_ORDER_OPERATOR_PARAMETER_LEVEL_CONSTRAINT_NOT_MET", Value: 4272, Severity: SeverityError, ParameterCount: 4},
	4273: {Name: "HIGHER_ORDER_OPERATOR_COPARAMETER_LEVEL_CONSTRAINTS_EXCEEDED", Value: 4273, Severity: SeverityError, ParameterCount: 2},
	4274: {Name: "LAMBDA_OPERATOR_ARGUMENT_HAS_INCORRECT_ARITY", Value: 4274, Severity: SeverityError, ParameterCount: 4},
	4275: {Name: "LAMBDA_GIVEN_WHERE_EXPRESSION_REQUIRED", Value: 4275, Severity: SeverityError, ParameterCount: 0},
	4290: {Name: "RECURSIVE_OPERATOR_PRIMES_PARAMETER", Value: 4290, Severity: SeverityError, ParameterCount: 2},
	4291: {Name: "RECURSIVE_OPERATOR_DECLARED_BUT_NOT_DEFINED", Value: 4291, Severity: SeverityError, ParameterCount: 1},
	4292: {Name: "RECURSIVE_OPERATOR_DECLARATION_DEFINITION_ARITY_MISMATCH", Value: 4292, Severity: SeverityError, ParameterCount: 1},
	4293: {Name: "RECURSIVE_OPERATOR_DEFINED_IN_WRONG_LET_IN_LEVEL", Value: 4293, Severity: SeverityError, ParameterCount: 1},
	4294: {Name: "RECURSIVE_SECTION_CONTAINS_ILLEGAL_DEFINITION", Value: 4294, Severity: SeverityError, ParameterCount: 1},
	4310: {Name: "ALWAYS_PROPERTY_SENSITIVE_TO_STUTTERING", Value: 4310, Severity: SeverityError, ParameterCount: 0},
	4311: {Name: "EVENTUALLY_PROPERTY_SENSITIVE_TO_STUTTERING", Value: 4311, Severity: SeverityError, ParameterCount: 0},
	4312: {Name: "BINARY_TEMPORAL_OPERATOR_WITH_ACTION_LEVEL_PARAMETER", Value: 4312, Severity: SeverityError, ParameterCount: 0},
	4313: {Name: "LOGICAL_OPERATOR_WITH_MIXED_ACTION_TEMPORAL_PARAMETERS", Value: 4313, Severity: SeverityError, ParameterCount: 1},
	4314: {Name: "QUANTIFIED_TEMPORAL_FORMULA_WITH_ACTION_LEVEL_BOUND", Value: 4314, Severity: SeverityError, ParameterCount: 0},
	4315: {Name: "QUANTIFICATION_WITH_TEMPORAL_LEVEL_BOUND", Value: 4315, Severity: SeverityError, ParameterCount: 2},
	4330: {Name: "LABEL_PARAMETER_REPETITION", Value: 4330, Severity: SeverityError, ParameterCount: 2},
	4331: {Name: "LABEL_PARAMETER_MISSING", Value: 4331, Severity: SeverityError, ParameterCount: 2},
	4332: {Name: "LABEL_PARAMETER_UNNECESSARY", Value: 4332, Severity: SeverityError, ParameterCount: -1},
	4333: {Name: "LABEL_NOT_IN_DEFINITION_OR_PROOF_STEP", Value: 4333, Severity: SeverityError, ParameterCount: 0},
	4334: {Name: "LABEL_NOT_ALLOWED_IN_NESTED_ASSUME_PROVE_WITH_NEW", Value: 4334, Severity: SeverityError, ParameterCount: 0},
	4335: {Name: "LABEL_NOT_ALLOWED_IN_FUNCTION_EXCEPT", Value: 4335, Severity: SeverityError, ParameterCount: 0},
	4336: {Name: "LABEL_REDEFINITION", Value: 4336, Severity: SeverityError, ParameterCount: 1},
	4337: {Name: "LABEL_GIVEN_INCORRECT_NUMBER_OF_ARGUMENTS", Value: 4337, Severity: SeverityError, ParameterCount: 1},
	4350: {Name: "PROOF_STEP_WITH_IMPLICIT_LEVEL_CANNOT_HAVE_NAME", Value: 4350, Severity: SeverityError, ParameterCount: 0},
	4351: {Name: "PROOF_STEP_NON_EXPRESSION_USED_AS_EXPRESSION", Value: 4351, Severity: SeverityError, ParameterCount: 1},
	4352: {Name: "TEMPORAL_PROOF_GOAL_WITH_NON_CONSTANT_TAKE_WITNESS_HAVE", Value: 4352, Severity: SeverityError, ParameterCount: 0},
	4353: {Name: "TEMPORAL_PROOF_GOAL_WITH_NON_CONSTANT_CASE", Value: 4353, Severity: SeverityError, ParameterCount: 0},
	4354: {Name: "QUANTIFIED_TEMPORAL_PICK_FORMULA_WITH_NON_CONSTANT_BOUND", Value: 4354, Severity: SeverityError, ParameterCount: 0},
	4355: {Name: "ASSUME_PROVE_USED_WHERE_EXPRESSION_REQUIRED", Value: 4355, Severity: SeverityError, ParameterCount: 0},
	4356: {Name: "ASSUME_PROVE_NEW_CONSTANT_HAS_TEMPORAL_LEVEL_BOUND", Value: 4356, Severity: SeverityError, ParameterCount: 0},
	4357: {Name: "USE_OR_HIDE_FACT_NOT_VALID", Value: 4357, Severity: SeverityError, ParameterCount: 0},
	4800: {Name: "EXTENDED_MODULES_SYMBOL_UNIFICATION_AMBIGUITY", Value: 4800, Severity: SeverityWarning, ParameterCount: 4},
	4801: {Name: "INSTANCED_MODULES_SYMBOL_UNIFICATION_AMBIGUITY", Value: 4801, Severity: SeverityWarning, ParameterCount: 2},
	4802: {Name: "RECORD_CONSTRUCTOR_FIELD_NAME_CLASH", Value: 4802, Severity: SeverityWarning, ParameterCount: 2},
	4803: {Name: "PLUSCAL_ALGORITHM_AND_TRANSLATION_CHANGED_SINCE_LAST_TRANSLATION", Value: 4803, Severity: SeverityWarning, ParameterCount: 1},
	4804: {Name: "PLUSCAL_ALGORITHM_CHANGED_SINCE_LAST_TRANSLATION", Value: 4804, Severity: SeverityWarning, ParameterCount: 1},
	4805: {Name: "PLUSCAL_TRANSLATION_CHANGED_SINCE_LAST_TRANSLATION", Value: 4805, Severity: SeverityWarning, ParameterCount: 1},
}

// Preserve the original ErrorDetails arguments separately from the native
// diagnostic summary, whose wording need not use the same format string.
func sanyDiagnosticParameters(diagnostic Diagnostic, parameters ...any) Diagnostic {
	diagnostic.SANYParameters = append([]any(nil), parameters...)
	return diagnostic
}
