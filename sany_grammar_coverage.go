package tlago

type SanyGrammarProductionCoverageStatus string

const (
	SanyProductionFolded   SanyGrammarProductionCoverageStatus = "folded"
	SanyProductionToken    SanyGrammarProductionCoverageStatus = "token"
	SanyProductionDeferred SanyGrammarProductionCoverageStatus = "deferred"
)

type SanyGrammarProductionCoverage struct {
	Status         SanyGrammarProductionCoverageStatus
	Implementation string
	Note           string
}

var SanyGrammarProductionCoverageByName = map[string]SanyGrammarProductionCoverage{
	"BangExt": {
		Status:         SanyProductionFolded,
		Implementation: "NoOpExtension",
		Note:           "Bang-qualified selector chains are parsed by NoOpExtension and GeneralId instead of a separate JavaCC BangExt method.",
	},
	"ConjList": {
		Status:         SanyProductionFolded,
		Implementation: "JunctionList",
		Note:           "Conjunction and disjunction bullet lists share the indentation-aware JunctionList implementation.",
	},
	"ConstantDeclarationItems": {
		Status:         SanyProductionFolded,
		Implementation: "ParamDeclaration and ConstantDeclarationItem",
		Note:           "The Go parser represents JavaCC constant declaration items through ParamDeclaration plus the helper used by recursive declarations.",
	},
	"DisjList": {
		Status:         SanyProductionFolded,
		Implementation: "JunctionList",
		Note:           "Conjunction and disjunction bullet lists share the indentation-aware JunctionList implementation.",
	},
	"Expression": {
		Status:         SanyProductionFolded,
		Implementation: "ExpressionUntil",
		Note:           "Expression parsing is parameterized by a stop predicate while preserving the SANY operator-stack reduction path.",
	},
	"ExtendableExpr": {
		Status:         SanyProductionFolded,
		Implementation: "ExpressionUntil",
		Note:           "Extendable expression parsing is folded into the central expression loop, including postfix and infix extension handling.",
	},
	"InfixOp": {
		Status:         SanyProductionFolded,
		Implementation: "consumeOperator and genericOperatorNode",
		Note:           "Operator node construction is table-driven from generated SANY operator metadata.",
	},
	"InfixOpToken": {
		Status:         SanyProductionToken,
		Implementation: "SanyTokenize and SanyOperatorsBySymbol",
		Note:           "JavaCC operator-token productions are generated into token literals and operator metadata rather than parser methods.",
	},
	"Junctions": {
		Status:         SanyProductionFolded,
		Implementation: "JunctionList",
		Note:           "Both JavaCC Junctions alternatives are implemented by one indentation-aware junction parser.",
	},
	"NEPrefixOpToken": {
		Status:         SanyProductionToken,
		Implementation: "SanyTokenize and SanyOperatorsBySymbol",
		Note:           "The non-expression prefix token set is represented by generated token literals and operator fixity checks.",
	},
	"NonExpPrefixOp": {
		Status:         SanyProductionFolded,
		Implementation: "isNEPrefixOperator and consumeOperator",
		Note:           "Non-expression prefix operators are recognized by table-driven fixity checks instead of a separate parser production.",
	},
	"OpOrExpr": {
		Status:         SanyProductionFolded,
		Implementation: "Substitution and OpArgs",
		Note:           "The current parser accepts expression operands and operator identifiers through surrounding substitution and argument parsers.",
	},
	"OpSuite": {
		Status:         SanyProductionFolded,
		Implementation: "OpArgs",
		Note:           "OpSuite contributes heirs to JavaCC OpArgs; the Go parser builds the argument list directly in OpArgs.",
	},
	"OpenStart": {
		Status:         SanyProductionFolded,
		Implementation: "startsOpenExpression",
		Note:           "Open-expression lookahead is represented as a predicate over token kinds.",
	},
	"ParamSubDecl": {
		Status:         SanyProductionFolded,
		Implementation: "ParamDeclaration",
		Note:           "JavaCC ParamSubDecl only wraps CONSTANT(S) as N_ConsDecl; the Go ParamDeclaration method constructs that node directly.",
	},
	"ParenthesesExpression": {
		Status:         SanyProductionFolded,
		Implementation: "PrimitiveExpression",
		Note:           "Parenthesized, brace, square-bracket, tuple/action, and fairness primitives are dispatched from PrimitiveExpression.",
	},
	"PostfixOp": {
		Status:         SanyProductionFolded,
		Implementation: "consumeOperator and genericOperatorNode",
		Note:           "Operator node construction is table-driven from generated SANY operator metadata.",
	},
	"PostfixOpToken": {
		Status:         SanyProductionToken,
		Implementation: "SanyTokenize and SanyOperatorsBySymbol",
		Note:           "JavaCC operator-token productions are generated into token literals and operator metadata rather than parser methods.",
	},
	"PrefixOp": {
		Status:         SanyProductionFolded,
		Implementation: "consumeOperator and genericOperatorNode",
		Note:           "Operator node construction is table-driven from generated SANY operator metadata.",
	},
	"PrefixOpToken": {
		Status:         SanyProductionToken,
		Implementation: "SanyTokenize and SanyOperatorsBySymbol",
		Note:           "JavaCC operator-token productions are generated into token literals and operator metadata rather than parser methods.",
	},
	"Prelude": {
		Status:         SanyProductionFolded,
		Implementation: "CompilationUnit",
		Note:           "The tokenizer skips pre-module text until the first module header; CompilationUnit begins at that header.",
	},
	"PrimitiveExp": {
		Status:         SanyProductionFolded,
		Implementation: "PrimitiveExpression",
		Note:           "The Go method uses the Go naming form while covering JavaCC PrimitiveExp dispatch.",
	},
	"UseOrHideOrBy": {
		Status:         SanyProductionFolded,
		Implementation: "UseOrHide",
		Note:           "USE, HIDE, and BY proof references are parsed through UseOrHide and surrounding proof-step dispatch.",
	},
}
