package sany_tests

// Mechanically ported from test/util/AstNode.java and SyntaxCorpusFileParser.java.
import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"unicode"
)

type syntaxCorpusAST struct {
	kind       string
	children   []*syntaxCorpusAST
	fields     map[string][]*syntaxCorpusAST
	fieldNames map[*syntaxCorpusAST]string
}

var syntaxCorpusKinds = []struct{ enum, name string }{
	{"ADDRESS", "address"},
	{"ALL_MAP_TO", "all_map_to"},
	{"ALWAYS", "always"},
	{"APPROX", "approx"},
	{"ASSIGN", "assign"},
	{"ASSUME_PROVE", "assume_prove"},
	{"ASSUMPTION", "assumption"},
	{"ASYMP", "asymp"},
	{"BIGCIRC", "bigcirc"},
	{"BINARY_NUMBER", "binary_number"},
	{"BNF_RULE", "bnf_rule"},
	{"BOOLEAN", "boolean"},
	{"BOUND_INFIX_OP", "bound_infix_op"},
	{"BOUND_NONFIX_OP", "bound_nonfix_op"},
	{"BOUND_OP", "bound_op"},
	{"BOUND_POSTFIX_OP", "bound_postfix_op"},
	{"BOUND_PREFIX_OP", "bound_prefix_op"},
	{"BOUNDED_QUANTIFICATION", "bounded_quantification"},
	{"BULLET", "bullet"},
	{"BULLET_CONJ", "bullet_conj"},
	{"BULLET_DISJ", "bullet_disj"},
	{"CAP", "cap"},
	{"CASE", "case"},
	{"CASE_ARM", "case_arm"},
	{"CASE_ARROW", "case_arrow"},
	{"CASE_BOX", "case_box"},
	{"CASE_PROOF_STEP", "case_proof_step"},
	{"CDOT", "cdot"},
	{"CHILD_ID", "child_id"},
	{"CHOOSE", "choose"},
	{"CIRC", "circ"},
	{"COLON", "colon"},
	{"CONG", "cong"},
	{"CONJ_ITEM", "conj_item"},
	{"CONJ_LIST", "conj_list"},
	{"CONSTANT_DECLARATION", "constant_declaration"},
	{"CUP", "cup"},
	{"DEF_EQ", "def_eq"},
	{"DEFINITION_PROOF_STEP", "definition_proof_step"},
	{"DISJ_ITEM", "disj_item"},
	{"DISJ_LIST", "disj_list"},
	{"DIV", "div"},
	{"DOMAIN", "domain"},
	{"DOTEQ", "doteq"},
	{"DOTS_2", "dots_2"},
	{"DOTS_3", "dots_3"},
	{"DOUBLE_LINE", "double_line"},
	{"ENABLED", "enabled"},
	{"EQ", "eq"},
	{"EQUIV", "equiv"},
	{"EVENTUALLY", "eventually"},
	{"EXCEPT", "except"},
	{"EXCEPT_UPDATE", "except_update"},
	{"EXCEPT_UPDATE_FN_APPL", "except_update_fn_appl"},
	{"EXCEPT_UPDATE_RECORD_FIELD", "except_update_record_field"},
	{"EXCEPT_UPDATE_SPECIFIER", "except_update_specifier"},
	{"EXCL", "excl"},
	{"EXISTS", "exists"},
	{"EXTENDS", "extends"},
	{"FAIR", "fair"},
	{"FAIRNESS", "fairness"},
	{"FINITE_SET_LITERAL", "finite_set_literal"},
	{"FORALL", "forall"},
	{"FUNCTION_DEFINITION", "function_definition"},
	{"FUNCTION_EVALUATION", "function_evaluation"},
	{"FUNCTION_LITERAL", "function_literal"},
	{"GEQ", "geq"},
	{"GETS", "gets"},
	{"GG", "gg"},
	{"HAVE_PROOF_STEP", "have_proof_step"},
	{"HEADER_LINE", "header_line"},
	{"HEX_NUMBER", "hex_number"},
	{"IF_THEN_ELSE", "if_then_else"},
	{"IFF", "iff"},
	{"IMPLIES", "implies"},
	{"IN", "in"},
	{"INFIX_OP_SYMBOL", "infix_op_symbol"},
	{"INNER_ASSUME_PROVE", "inner_assume_prove"},
	{"INSTANCE", "instance"},
	{"INT_NUMBER_SET", "int_number_set"},
	{"LABEL", "label"},
	{"LABEL_AS", "label_as"},
	{"LAMBDA", "lambda"},
	{"LAND", "land"},
	{"LANGLE_BRACKET", "langle_bracket"},
	{"LD_TTILE", "ld_ttile"},
	{"LEADS_TO", "leads_to"},
	{"LEQ", "leq"},
	{"LET_IN", "let_in"},
	{"LL", "ll"},
	{"LNOT", "lnot"},
	{"LOCAL_DEFINITION", "local_definition"},
	{"LOR", "lor"},
	{"LS_TTILE", "ls_ttile"},
	{"LT", "lt"},
	{"MAPS_TO", "maps_to"},
	{"MINUS", "minus"},
	{"MODULE", "module"},
	{"MODULE_DEFINITION", "module_definition"},
	{"MODULE_REF", "module_ref"},
	{"NAT_NUMBER", "nat_number"},
	{"NAT_NUMBER_SET", "nat_number_set"},
	{"NEGATIVE", "negative"},
	{"NEQ", "neq"},
	{"NEW", "new"},
	{"NON_TERMINAL_PROOF", "non_terminal_proof"},
	{"NOTIN", "notin"},
	{"OCTAL_NUMBER", "octal_number"},
	{"ODOT", "odot"},
	{"OMINUS", "ominus"},
	{"OPERATOR_ARGS", "operator_args"},
	{"OPERATOR_DECLARATION", "operator_declaration"},
	{"OPERATOR_DEFINITION", "operator_definition"},
	{"OPLUS", "oplus"},
	{"OSLASH", "oslash"},
	{"OTHER_ARM", "other_arm"},
	{"OTIMES", "otimes"},
	{"PARENTHESES", "parentheses"},
	{"PCAL_ALGORITHM", "pcal_algorithm"},
	{"PCAL_ALGORITHM_BODY", "pcal_algorithm_body"},
	{"PCAL_ALGORITHM_START", "pcal_algorithm_start"},
	{"PCAL_ASSERT", "pcal_assert"},
	{"PCAL_ASSIGN", "pcal_assign"},
	{"PCAL_AWAIT", "pcal_await"},
	{"PCAL_DEFINITIONS", "pcal_definitions"},
	{"PCAL_EITHER", "pcal_either"},
	{"PCAL_GOTO", "pcal_goto"},
	{"PCAL_IF", "pcal_if"},
	{"PCAL_LHS", "pcal_lhs"},
	{"PCAL_MACRO", "pcal_macro"},
	{"PCAL_MACRO_CALL", "pcal_macro_call"},
	{"PCAL_MACRO_DECL", "pcal_macro_decl"},
	{"PCAL_PRINT", "pcal_print"},
	{"PCAL_PROC_CALL", "pcal_proc_call"},
	{"PCAL_PROC_DECL", "pcal_proc_decl"},
	{"PCAL_PROC_VAR_DECL", "pcal_proc_var_decl"},
	{"PCAL_PROC_VAR_DECLS", "pcal_proc_var_decls"},
	{"PCAL_PROCEDURE", "pcal_procedure"},
	{"PCAL_PROCESS", "pcal_process"},
	{"PCAL_RETURN", "pcal_return"},
	{"PCAL_SKIP", "pcal_skip"},
	{"PCAL_VAR_DECL", "pcal_var_decl"},
	{"PCAL_VAR_DECLS", "pcal_var_decls"},
	{"PCAL_WHILE", "pcal_while"},
	{"PCAL_WITH", "pcal_with"},
	{"PCAL_END_EITHER", "pcal_end_either"},
	{"PCAL_END_IF", "pcal_end_if"},
	{"PCAL_END_WHILE", "pcal_end_while"},
	{"PCAL_END_WITH", "pcal_end_with"},
	{"PICK_PROOF_STEP", "pick_proof_step"},
	{"PLUS", "plus"},
	{"PLUS_ARROW", "plus_arrow"},
	{"POSTFIX_OP_SYMBOL", "postfix_op_symbol"},
	{"POWERSET", "powerset"},
	{"PREC", "prec"},
	{"PRECEQ", "preceq"},
	{"PREFIX_OP_SYMBOL", "prefix_op_symbol"},
	{"PREFIXED_OP", "prefixed_op"},
	{"PREV_FUNC_VAL", "prev_func_val"},
	{"PROOF_STEP", "proof_step"},
	{"PROOF_STEP_ID", "proof_step_id"},
	{"PROOF_STEP_REF", "proof_step_ref"},
	{"PROPTO", "propto"},
	{"QED_STEP", "qed_step"},
	{"QQ", "qq"},
	{"QUANTIFIER_BOUND", "quantifier_bound"},
	{"RANGLE_BRACKET", "rangle_bracket"},
	{"RANGLE_BRACKET_SUB", "rangle_bracket_sub"},
	{"RD_TTILE", "rd_ttile"},
	{"REAL_NUMBER_SET", "real_number_set"},
	{"RECORD_LITERAL", "record_literal"},
	{"RECORD_VALUE", "record_value"},
	{"RECURSIVE_DECLARATION", "recursive_declaration"},
	{"RS_TTILE", "rs_ttile"},
	{"SET_FILTER", "set_filter"},
	{"SET_IN", "set_in"},
	{"SET_MAP", "set_map"},
	{"SET_OF_FUNCTIONS", "set_of_functions"},
	{"SET_OF_RECORDS", "set_of_records"},
	{"SIM", "sim"},
	{"SIMEQ", "simeq"},
	{"SINGLE_LINE", "single_line"},
	{"SOURCE_FILE", "source_file"},
	{"SQCAP", "sqcap"},
	{"SQCUP", "sqcup"},
	{"SQSUBSET", "sqsubset"},
	{"SQSUBSETEQ", "sqsubseteq"},
	{"SQSUPSET", "sqsupset"},
	{"SQSUPSETEQ", "sqsupseteq"},
	{"STAR", "star"},
	{"STATEMENT_LEVEL", "statement_level"},
	{"STEP_EXPR_NO_STUTTER", "step_expr_no_stutter"},
	{"STEP_EXPR_OR_STUTTER", "step_expr_or_stutter"},
	{"STRING", "string"},
	{"SUBEXPR_COMPONENT", "subexpr_component"},
	{"SUBEXPR_PREFIX", "subexpr_prefix"},
	{"SUBEXPR_TREE_NAV", "subexpr_tree_nav"},
	{"SUBEXPRESSION", "subexpression"},
	{"SUBSET", "subset"},
	{"SUBSETEQ", "subseteq"},
	{"SUBSTITUTION", "substitution"},
	{"SUCC", "succ"},
	{"SUCCEQ", "succeq"},
	{"SUFFICES_PROOF_STEP", "suffices_proof_step"},
	{"SUP_PLUS", "sup_plus"},
	{"SUPSET", "supset"},
	{"SUPSETEQ", "supseteq"},
	{"TAKE_PROOF_STEP", "take_proof_step"},
	{"TEMPORAL_EXISTS", "temporal_exists"},
	{"TEMPORAL_FORALL", "temporal_forall"},
	{"TERMINAL_PROOF", "terminal_proof"},
	{"THEOREM", "theorem"},
	{"TIMES", "times"},
	{"TUPLE_LITERAL", "tuple_literal"},
	{"TUPLE_OF_IDENTIFIERS", "tuple_of_identifiers"},
	{"UNBOUNDED_QUANTIFICATION", "unbounded_quantification"},
	{"UNCHANGED", "unchanged"},
	{"UNION", "union"},
	{"UPLUS", "uplus"},
	{"USE_BODY", "use_body"},
	{"USE_BODY_DEF", "use_body_def"},
	{"USE_BODY_EXPR", "use_body_expr"},
	{"USE_OR_HIDE", "use_or_hide"},
	{"VARIABLE_DECLARATION", "variable_declaration"},
	{"VERTVERT", "vertvert"},
	{"WITNESS_PROOF_STEP", "witness_proof_step"},
	{"WR", "wr"},
	{"AMP", "amp"},
	{"AMPAMP", "ampamp"},
	{"ASTERISK", "asterisk"},
	{"BOOLEAN_SET", "boolean_set"},
	{"COMPOSE", "compose"},
	{"DOL", "dol"},
	{"DOLDOL", "doldol"},
	{"ESCAPE_CHAR", "escape_char"},
	{"FORMAT", "format"},
	{"GT", "gt"},
	{"HASHHASH", "hashhash"},
	{"IDENTIFIER", "identifier"},
	{"IDENTIFIER_REF", "identifier_ref"},
	{"LEVEL", "level"},
	{"MAP_FROM", "map_from"},
	{"MAP_TO", "map_to"},
	{"MINUSMINUS", "minusminus"},
	{"MOD", "mod"},
	{"MODMOD", "modmod"},
	{"MUL", "mul"},
	{"MULMUL", "mulmul"},
	{"NAME", "name"},
	{"PLACEHOLDER", "placeholder"},
	{"PLUSPLUS", "plusplus"},
	{"POW", "pow"},
	{"POWPOW", "powpow"},
	{"PRIME", "prime"},
	{"REAL_NUMBER", "real_number"},
	{"SETMINUS", "setminus"},
	{"SLASH", "slash"},
	{"SLASHSLASH", "slashslash"},
	{"STRING_SET", "string_set"},
	{"SUP_HASH", "sup_hash"},
	{"VALUE", "value"},
	{"VERT", "vert"},
}
var syntaxCorpusUnused = func() map[string]bool {
	result := make(map[string]bool)
	for _, kind := range syntaxCorpusKinds {
		result[kind.name] = true
	}
	return result
}()

func newSyntaxCorpusAST(kind string) *syntaxCorpusAST {
	return &syntaxCorpusAST{kind: kind, fields: make(map[string][]*syntaxCorpusAST), fieldNames: make(map[*syntaxCorpusAST]string)}
}
func (n *syntaxCorpusAST) addChild(child *syntaxCorpusAST) *syntaxCorpusAST {
	n.children = append(n.children, child)
	return n
}
func (n *syntaxCorpusAST) addChildren(children []*syntaxCorpusAST) *syntaxCorpusAST {
	n.children = append(n.children, children...)
	return n
}
func (n *syntaxCorpusAST) addField(name string, child *syntaxCorpusAST) *syntaxCorpusAST {
	n.children = append(n.children, child)
	// Preserve source addField: it appends to an existing field list only.
	if _, ok := n.fields[name]; ok {
		n.fields[name] = append(n.fields[name], child)
	}
	n.fieldNames[child] = name
	return n
}
func (n *syntaxCorpusAST) alias(from, to string) *syntaxCorpusAST {
	if n.kind == from {
		n.kind = to
	}
	for _, child := range n.children {
		child.alias(from, to)
	}
	return n
}
func (n *syntaxCorpusAST) String() string {
	var out strings.Builder
	out.WriteByte('(')
	out.WriteString(n.kind)
	for _, child := range n.children {
		out.WriteByte(' ')
		if name, ok := n.fieldNames[child]; ok {
			out.WriteString(name)
			out.WriteString(": ")
		}
		out.WriteString(child.String())
	}
	out.WriteByte(')')
	return out.String()
}
func (n *syntaxCorpusAST) testEquality(t *testing.T, other *syntaxCorpusAST) {
	t.Helper()
	if other == nil || n.kind != other.kind {
		t.Fatalf("AST kind: expected %s; actual %v", n.kind, other)
	}
	if len(n.children) != len(other.children) {
		t.Fatalf("%s child count: expected %d; actual %d", n.kind, len(n.children), len(other.children))
	}
	for i, child := range n.children {
		actual := other.children[i]
		if name, ok := n.fieldNames[child]; ok {
			actualName, present := other.fieldNames[actual]
			if !present || name != actualName {
				t.Fatalf("%s field: expected %q; actual %q (present %t)", n.kind, name, actualName, present)
			}
		}
		child.testEquality(t, actual)
	}
}

type syntaxCorpusDSLError struct {
	message string
	offset  int
}

func (e *syntaxCorpusDSLError) Error() string {
	return fmt.Sprintf("%s (offset %d)", e.message, e.offset)
}

type syntaxCorpusDSLToken struct{ kind, lexeme string }

func tokenizeSyntaxCorpusDSL(input string) ([]syntaxCorpusDSLToken, error) {
	var tokens []syntaxCorpusDSLToken
	var word strings.Builder
	for i, ch := range []rune(input) {
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' {
			word.WriteRune(ch)
			continue
		}
		if word.Len() != 0 {
			tokens = append(tokens, syntaxCorpusDSLToken{"IDENTIFIER", word.String()})
			word.Reset()
		}
		// Character.isWhitespace excludes the three nonbreaking spaces.
		if (unicode.IsSpace(ch) && ch != '\u00a0' && ch != '\u2007' && ch != '\u202f' && ch != '\u0085') || (ch >= 0x1c && ch <= 0x1f) {
			continue
		}
		kind := ""
		switch ch {
		case ':':
			kind = "COLON"
		case '(':
			kind = "LPAREN"
		case ')':
			kind = "RPAREN"
		default:
			return nil, &syntaxCorpusDSLError{fmt.Sprintf("Unhandled char %s", string(ch)), i}
		}
		tokens = append(tokens, syntaxCorpusDSLToken{kind, string(ch)})
	}
	// Source does not flush a trailing identifier at end of input.
	return tokens, nil
}

type syntaxCorpusDSLParser struct {
	tokens  []syntaxCorpusDSLToken
	current int
}

func (p *syntaxCorpusDSLParser) check(kind string) bool {
	return p.current < len(p.tokens) && p.tokens[p.current].kind == kind
}
func (p *syntaxCorpusDSLParser) consume(kind string) (syntaxCorpusDSLToken, error) {
	if !p.check(kind) {
		return syntaxCorpusDSLToken{}, &syntaxCorpusDSLError{"Expected " + kind, p.current}
	}
	token := p.tokens[p.current]
	p.current++
	return token, nil
}
func (p *syntaxCorpusDSLParser) parseAST() (*syntaxCorpusAST, error) {
	if _, err := p.consume("LPAREN"); err != nil {
		return nil, err
	}
	token, err := p.consume("IDENTIFIER")
	if err != nil {
		return nil, err
	}
	found := false
	for _, kind := range syntaxCorpusKinds {
		if kind.name == token.lexeme {
			found = true
			break
		}
	}
	if !found {
		return nil, &syntaxCorpusDSLError{"Could not find node " + token.lexeme, 0}
	}
	delete(syntaxCorpusUnused, token.lexeme)
	node := newSyntaxCorpusAST(token.lexeme)
	for p.current < len(p.tokens) && !p.check("RPAREN") {
		if p.check("LPAREN") {
			child, err := p.parseAST()
			if err != nil {
				return nil, err
			}
			node.addChild(child)
		} else {
			field, err := p.consume("IDENTIFIER")
			if err != nil {
				return nil, err
			}
			if _, err := p.consume("COLON"); err != nil {
				return nil, err
			}
			child, err := p.parseAST()
			if err != nil {
				return nil, err
			}
			node.addField(field.lexeme, child)
		}
	}
	if _, err := p.consume("RPAREN"); err != nil {
		return nil, err
	}
	return node, nil
}
func parseSyntaxCorpusAST(input string) (*syntaxCorpusAST, error) {
	tokens, err := tokenizeSyntaxCorpusDSL(input)
	if err != nil {
		return nil, err
	}
	parser := &syntaxCorpusDSLParser{tokens: tokens}
	root, err := parser.parseAST()
	if err != nil {
		return nil, err
	}
	if parser.current != len(tokens) {
		return nil, &syntaxCorpusDSLError{"Unparsed tokens remain", parser.current}
	}
	return root, nil
}

var syntaxCorpusHeaderRE = regexp.MustCompile(`(?ms)^===+\|\|\|\r?\n(.*?)\r?\n===+\|\|\|\r?\n`)
var syntaxCorpusSeparatorRE = regexp.MustCompile(`(?m)^---+\|\|\|\r?\n`)

func javaCorpusTrim(s string) string {
	return strings.TrimFunc(s, func(ch rune) bool { return ch <= 0x20 })
}
func parseSyntaxCorpusFile(file, content string) ([]sanySyntaxCorpusCase, error) {
	headers := syntaxCorpusHeaderRE.FindAllStringSubmatchIndex(content, -1)
	separators := syntaxCorpusSeparatorRE.FindAllStringIndex(content, -1)
	var cases []sanySyntaxCorpusCase
	for i, header := range headers {
		text := content[header[2]:header[3]]
		if i >= len(separators) {
			return nil, &syntaxCorpusDSLError{fmt.Sprintf("%s: Test %s does not have separator", file, text), 0}
		}
		lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
		var name []string
		readingName := true
		tc := sanySyntaxCorpusCase{}
		for _, line := range lines {
			if strings.HasPrefix(line, ":") {
				readingName = false
				switch strings.ToLower(strings.TrimSpace(line)[1:]) {
				case "error":
					tc.ExpectError = true
				case "skip":
					tc.Skip = true
				default:
					return nil, fmt.Errorf("Invalid test attribute %s", line)
				}
			} else if readingName {
				name = append(name, line)
			}
		}
		tc.Title = javaCorpusTrim(strings.Join(name, "\n"))
		separator := separators[i]
		if separator[0] < header[1] {
			return nil, fmt.Errorf("%s: separator precedes input", file)
		}
		tc.Source = content[header[1]:separator[0]]
		end := len(content)
		if i+1 < len(headers) {
			end = headers[i+1][0]
		}
		if !tc.ExpectError {
			root, err := parseSyntaxCorpusAST(content[separator[1]:end])
			if err != nil {
				return nil, err
			}
			tc.ExpectedAST = root
		}
		cases = append(cases, tc)
	}
	if len(cases) == 0 {
		return nil, &syntaxCorpusDSLError{fmt.Sprintf("%s: Failed to find any tests", file), 0}
	}
	return cases, nil
}
