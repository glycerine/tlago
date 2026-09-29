# tlago SANY Architecture

This document records the architecture of Java SANY and the target architecture
for the Go port. The source of truth is the Java implementation under
`../tlaplus/tlatools/org.lamport.tlatools/src/tla2sany` plus the frozen copies
under `test_vectors/java-sany`. The near-term conformance oracle is Java SANY's
native XML export, not ApalacheIR.

The important conclusion is that SANY is not "parser -> AST -> XML". It is:

1. driver initialization,
2. external module discovery and dependency ordering,
3. JavaCC concrete syntax tree construction,
4. semantic graph generation with scoped symbol contexts,
5. level checking over the semantic graph,
6. XML export from that semantic graph while collecting referenced symbols into
   a context table.

The Go port should mirror those layers while keeping them in one central
`tlago` package.

## Java Package Map

`tla2sany.drivers`

- `SANY` is the public front-end driver.
- `SANY.parse` calls `frontEndInitialize`, `frontEndParse`,
  `frontEndSemanticAnalysis`, and optionally linting.
- It owns exit-code policy and routing of parse vs semantic vs level errors.

`tla2sany.modanalyzer`

- `SpecObj` owns one complete root specification load.
- `ParseUnit` owns one external `.tla` file and its concrete parse tree.
- `ModulePointer`, `ModuleRelationships`, `ModuleRelatives`,
  `ParseUnitRelatives`, and `ModuleContext` model the file/module dependency
  graph before semantic analysis.
- This layer finds every module imported by `EXTENDS` or `INSTANCE`, enforces
  file/module naming, resolves inner-module visibility, detects dependency
  cycles, and computes the order for semantic analysis.

`tla2sany.parser`

- `TLAplusParser` and `TLAplusParserTokenManager` are generated from
  `javacc/tla+.jj`.
- `Operators` is the authoritative table of symbolic operators, precedence,
  associativity, fixity, synonyms, and built-in canonical spellings.
- `OperatorStack` and `JunctionListContext` are parser machinery, not optional
  conveniences. They define expression grouping and indentation-sensitive
  conjunction/disjunction lists.

`tla2sany.st`

- `TreeNode` is the CST interface.
- `SyntaxTreeNode` is the concrete CST node used by the parser and semantic
  generator.
- `Location` is the file/line/column interval format used throughout SANY XML
  and diagnostics.

`tla2sany.semantic`

- `Generator` translates CST nodes into the semantic graph.
- `Context` is a mapping from names to `SymbolNode`s.
- `SymbolTable` is a stack of `Context`s used for lookup while generating a
  module or expression.
- `ExternalModuleTable` stores all semantic `ModuleNode`s in dependency order
  and remembers the root module.
- `SemanticNode`, `LevelNode`, `SymbolNode`, `OpDefNode`, `OpDeclNode`,
  `FormalParamNode`, `OpApplNode`, `LetInNode`, `SubstInNode`,
  `APSubstInNode`, `InstanceNode`, `AssumeNode`, `TheoremNode`,
  `ThmOrAssumpDefNode`, proof nodes, and value nodes are the semantic graph.
- `BuiltInOperators` and `ASTConstants` encode built-in operator names and
  semantic node kinds.
- `LevelNode`, `SetOfLevelConstraints`, `SetOfArgLevelConstraints`,
  `ArgLevelParam`, and `ParamAndPosition` implement SANY level checking.

`tla2sany.xml`

- `XMLExporter` parses a spec through the normal SANY front end and exports the
  semantic graph.
- `SymbolContext` collects symbol definitions referenced during export into
  the XML `<context>` section.
- XML ordering is partly semantic and partly an artifact of Java collections;
  conformance tests should canonicalize unstable entry order and UIDs while
  preserving expression/declaration child order.

## Java Front-End Pipeline

### Driver

`SANY.parse(spec, fileName, out, settings)` performs the complete front-end run:

1. `frontEndInitialize` resets the global built-in context by calling
   `Context.reInit`.
2. `frontEndParse` calls `spec.loadSpec`, reports parse warnings/errors, and
   aborts on syntax or module-resolution failure.
3. `frontEndSemanticAnalysis` walks `spec.semanticAnalysisVector` in dependency
   order. For each external module, it creates a fresh `Generator`, translates
   the CST root into a `ModuleNode`, stores the module's external context in the
   `ExternalModuleTable`, and runs `moduleNode.levelCheck`.
4. Optional linting runs only after semantic analysis and level checking
   succeed.

The Go equivalent should keep this as the top-level API shape:

- `LoadSanySpec` should do dependency discovery and CST parsing.
- A SANY-shaped semantic generation pass should run after loading.
- `CheckSanySource`/`CheckSanySpec` should combine semantic analysis and level
  checking.
- XML export should consume the semantic graph, not reconstruct one from the
  simplified AST.

### Module Discovery

`SpecObj.loadSpec` starts from the root module name and repeatedly finds the
next unresolved `EXTENDS` or `INSTANCE` dependency.

The essential Java sequence is:

1. `findOrCreateParsedUnit(root, firstCall=true)` resolves the root file,
   creates a `ParseUnit`, parses it, and records its root `ModulePointer`.
2. While there is an unresolved extension or instantiation:
   - resolve or reuse the target `ParseUnit`,
   - record extension/instantiation edges between parse units,
   - reject cycles with `nonCircularityTest`,
   - resolve names between the importing module and the newly loaded external
     module.
3. `calculateDependencies(rootParseUnit)` appends parse units only after their
   extendees and instancees. This produces `semanticAnalysisVector`, the order
   consumed by `SANY.frontEndSemanticAnalysis`.

`ParseUnit.parseFile` does the file-level work:

- create `TLAplusParser`,
- call `parseTree.parse()`,
- set CST parent/level metadata,
- set the spec root name on the first parse,
- create the root `ModulePointer`,
- call `determineModuleRelationships`,
- enforce exact filename/module-name equality,
- optionally print the CST.

`ParseUnit.determineModuleRelationships` walks the CST of a file:

- allocates a `ModuleRelatives` record for each module in the file,
- records direct `EXTENDS` names,
- calculates the visible inner-module context at the module declaration point,
- scans the module body for inner modules and all forms of `INSTANCE`,
- recurses into inner modules.

The Go resolver should preserve this separation. Source loading and dependency
ordering belong before semantic generation. The semantic generator should
receive modules in dependency order and should not discover external files on
the fly.

### Concrete Syntax Tree

The CST is a JavaCC syntax tree, not the public semantic graph.

Important CST properties to preserve:

- Each node has a Java syntax kind from `SyntaxTreeConstants`.
- Token leaves preserve source image and source range.
- Comments are stored as special tokens and attached as pre-comments to the
  following syntactic node.
- Line endings are normalized the way Java's reader/token manager observes
  them: `\r\n` is one source newline for all token, CST, semantic-node, and XML
  locations.
- Tab handling uses Java SANY columns: tab stops are every 8 columns.
- Parser-generated nodes sometimes have wider locations than their expression
  children; SANY XML uses these exact node locations.
- Multiple modules in one file are represented as nested module CST nodes, but
  a file still has a single top-level root module whose name must match the
  file.

The Go parser should continue using the production SANY parser port only. There
must be no bootstrap parser, fallback parser, or compatibility shim.

## Semantic Graph Model

### Base Classes

Java uses a small inheritance lattice:

- `SemanticNode`: every semantic node has a UID, kind, CST node, location, and
  XML export hook.
- `LevelNode`: nodes with TLA+ level data. It owns `level`, `levelParams`,
  `levelConstraints`, `argLevelConstraints`, `argLevelParams`, `allParams`,
  `nonLeibnizParams`, `levelCorrect`, and `levelChecked`.
- `SymbolNode`: a node that defines a name. It has a `UniqueString` name, arity,
  symbol-table behavior, and XML reference/export behavior.
- `ExprOrOpArgNode`: base for ordinary expressions and operator arguments.
- `ExprNode`: ordinary expressions.

The Go port does not need Java inheritance, but it should represent the same
roles explicitly. A single package can still have Java-shaped types such as:

- `sanySemNode`
- `sanyLevelNode`
- `sanySymbol`
- `sanyModuleNode`
- `sanyOpDefNode`
- `sanyOpDeclNode`
- `sanyFormalParamNode`
- `sanyOpApplNode`
- `sanyLetInNode`
- `sanySubstInNode`
- `sanyInstanceNode`
- `sanyAssumeNode`
- `sanyTheoremNode`
- `sanyProofNode`
- `sanyStringNode`
- `sanyNumeralNode`
- `sanyDecimalNode`
- `sanyAtNode`

Those can be private Go structs in `package tlago`. Public APIs can keep the
existing `Spec`, `Module`, and expression model as a downstream/lowered view.

### Symbol Kinds

The Java semantic node kind constants that matter most for XML parity are:

- `ModuleKind = 1`
- `ConstantDeclKind = 2`
- `VariableDeclKind = 3`
- `BoundSymbolKind = 4`
- `UserDefinedOpKind = 5`
- `ModuleInstanceKind = 6`
- `BuiltInKind = 7`
- `OpArgKind = 8`
- `OpApplKind = 9`
- `LetInKind = 10`
- `FormalParamKind = 11`
- `TheoremKind = 12`
- `SubstInKind = 13`
- `AssumeProveKind = 14`
- `NumeralKind = 16`
- `DecimalKind = 17`
- `StringKind = 18`
- `AtNodeKind = 19`
- `AssumeKind = 20`
- `InstanceKind = 21`
- `NewSymbKind = 22`
- `ThmOrAssumpDefKind = 23`
- `NewConstantKind = 24`
- `NewVariableKind = 25`
- `NewStateKind = 26`
- `NewActionKind = 27`
- `NewTemporalKind = 28`
- `LabelKind = 29`
- `APSubstInKind = 30`
- `UseKind = 31`
- `HideKind = 32`
- proof kinds `33` through `37`.

Do not invent new kind numbers in the SANY semantic graph. If Go needs helper
nodes, keep them outside the XML-producing semantic graph or mark them as
internal-only.

### Context And Symbol Table

Java has two related structures:

- `Context`: a flat map from names to `SymbolNode`s. It is also a linked-list
  insertion history for retrieving declarations/definitions in SANY order.
- `SymbolTable`: a stack of `Context`s. Lookup searches from the top context
  down. New contexts are pushed for operator formal parameters, LETs, bound
  variables, inner modules, and nested ASSUME/PROVE scopes.

The base context of each external module starts as a duplicate of the global
built-in context. Module-level declarations and definitions are added to that
base context. Temporary local scopes are pushed above it and popped after the
local construct is generated.

Go should mirror this:

- A `sanyContext` maps both operator names and module names to semantic symbols.
- It must also preserve insertion history so `ModuleNode.getOpDefs`,
  `getConstantDecls`, and `getVariableDecls` can match Java ordering.
- A `sanySymbolTable` should hold a stack of contexts and expose
  `resolveSymbol`, `resolveModule`, `addSymbol`, `addModule`, `pushContext`,
  and `popContext`.
- Duplicate-name behavior belongs here, because Java diagnostics are emitted by
  `SymbolTable.addSymbol` and `addModule`.

### Built-Ins

`Context.reInit` creates a new global context and calls `Context.initialize`.
That function adds one `OpDefNode(BuiltInKind)` per `BuiltInOperators.Property`.

Each built-in operator records:

- name,
- arity (`-1` means variable arity),
- operator level,
- maximum allowed argument levels,
- argument weights.

`OpDefNode.setBuiltInLevel` also derives Leibniz metadata:

- an argument is Leibniz iff its weight is `1`;
- the operator is Leibniz iff every argument is Leibniz;
- built-ins are treated as already level-checked.

The Go port already has built-in metadata for XML. That metadata should become
the single built-in semantic symbol table, shared by semantic generation, level
checking, and XML export.

## Semantic Generation

Java `Generator` is the heart of SANY. Its public entry is:

- `generate(TreeNode) -> ModuleNode`

The important private decomposition is:

- `generateModule`
- `processExtendsList`
- `processVariables`
- `processParameters`
- `processOperator`
- `processFunction`
- `processLetIn`
- `generateExpression` / `generateExpressionOrLAP`
- `processChoose`
- `processBoundQuant`
- `processUnboundQuant`
- `processCase`
- `processSubsetOf`
- `processSetOfAll`
- `processFcnConst`
- `processRcdForms`
- `processAction`
- `processExcept`
- `generateExprOrOpArg`
- `processModuleDefinition`
- `processSubst` / `generateSubst`
- proof and ASSUME/PROVE helpers

The Go port should follow that decomposition. This is more important than the
exact Java class layout, because it encodes where contexts are pushed, where
symbols are registered, and which CST node locations become semantic locations.

### Module Generation

`generateModule` performs the module body pass:

1. Create the `ModuleNode` with the current context.
2. If this is an inner module, append it to the parent's definition list.
3. Set the generator's current module in the symbol table.
4. Process `EXTENDS`, merging extendee contexts into the current base context
   and copying extendee assumptions/theorems/top-level nodes.
5. Walk each module-body CST node and dispatch by syntax kind:
   - variable declaration,
   - constant/operator parameter declaration,
   - operator definition,
   - function definition,
   - module definition,
   - inner module,
   - unnamed instance,
   - theorem,
   - assumption,
   - top-level USE/HIDE,
   - recursive declaration.
6. Check for undefined recursive operators.
7. Return the `ModuleNode`.

For inner modules, Java creates a new `SymbolTable` that copies the outer stack,
pushes a fresh context for the inner module, generates the inner module, pops,
then adds the inner module's name to the outer module's symbol table.

### Operator Definitions

`processOperator` handles every syntactic operator-definition LHS form:

- identifier/function style `Foo(a, b(_)) == expr`,
- prefix operator definition,
- infix operator definition,
- postfix operator definition.

It:

1. pushes a new formal-parameter context,
2. creates `FormalParamNode`s with arities,
3. resolves operator synonyms for symbolic names,
4. checks whether the name was declared by a matching `RECURSIVE`,
5. generates the body expression while formals are in scope,
6. pops the formal context,
7. creates or finalizes an `OpDefNode`,
8. records labels,
9. appends LET-local definitions to the LET list and module-level definitions
   to the module's definition list.

This matters for XML levels: formal parameters are semantic symbols. Their
presence in `levelParams`, `allParams`, `argLevelParams`, and Leibniz metadata
drives operator application levels.

### Function Definitions

`processFunction` implements `f[x \in S] == body` differently from ordinary
operator definitions.

It creates an `OpDefNode` whose body is an `OpApplNode` for either
`$NonRecursiveFcnSpec` or `$RecursiveFcnSpec`. The bounded symbols of that
`OpApplNode` describe the function domain variables and their ranges. A
generator stack tracks whether `f[...]` appears in its own body; if so, the
operator of the function-spec node is changed to `$RecursiveFcnSpec`.

Go XML should not approximate function definitions as ordinary definitions.
They need the semantic function-spec shape because Java XML and level checking
see that built-in operator.

### LET/IN

`processLetIn`:

1. creates a new LET context and pushes it,
2. increments `curLevel`,
3. processes LET definitions and LET module definitions into local vectors,
4. handles LET-local recursive declarations,
5. decrements `curLevel`,
6. generates the IN expression in the LET context,
7. creates `LetInNode(defs, insts, body, context)`,
8. pops the LET context.

Java appends LET definitions to the containing module's `definitions` list in
some cases, but `LetInNode` is the expression node used by XML. Go should make
LET definitions semantic symbols and emit them through the LET node, rather
than synthesizing XML from expression AST alone.

`LetInNode.levelCheck` has a source-of-truth quirk that matters for XML parity:
it level-checks LET definitions, the IN body, and LET instances, but then sets
the LET expression's `level`, `levelParams`, and `allParams` from the IN body
only. Local definitions affect those sets only when the IN body references the
local definition's semantic symbol. `LetInNode` does not copy the IN body's
`nonLeibnizParams` into the LET node. Consequently:

- `F(x) == ENABLED G(x)` can make `x` non-Leibniz.
- `F(x) == LET D == TRUE IN ENABLED G(x)` keeps `x` Leibniz at the outer
  definition, because the LET node does not propagate the body's non-Leibniz
  coloring.

The current compact-AST XML exporter has compatibility code for this rule. In
the final semantic graph port, this should fall out of a Java-shaped
`LetInNode.levelCheck` implementation rather than a special XML heuristic.

### Expressions

`generateExpressionOrLAP` dispatches by CST kind.

Important mappings:

- numbers -> `NumeralNode` or `DecimalNode`
- strings -> `StringNode`
- infix/prefix/postfix -> `OpApplNode` with resolved operator symbol
- zero-argument general identifier -> `OpApplNode` with zero operands
- operator application -> `selectorToNode`, preserving generalized identifiers
  such as `A(1)!B(x)!C`
- tuple -> `$Tuple`
- function application -> `$FcnApply`, with multi-argument applications packed
  into `$Tuple`
- set enumeration -> `$SetEnumerate`
- Cartesian product -> `$CartesianProd`
- choose/quantifiers/comprehensions/function constructors -> `OpApplNode`s with
  bound-symbol arrays and range arrays
- record selection -> `$RcdSelect`
- record constructors and record sets -> `$RcdConstructor` / `$SetOfRcds`
  applied to `$Pair` nodes
- case -> `$Case` applied to `$Pair` nodes, with `OTHER` represented by a null
  first operand that XML later emits as `$Other`
- EXCEPT -> `$Except`, `$Pair`, `$Seq`, and `AtNode` for `@`
- action/fairness -> `$SquareAct`, `$AngleAct`, `$WF`, `$SF`

Operator arguments require `generateExprOrOpArg`. If an argument position
expects an operator (formal parameter arity > 0), SANY can create an `OpArgNode`
instead of an ordinary expression. The Go semantic graph must support this for
higher-order operators and for exact XML.

Function application has two easy-to-miss XML details:

- The first `$FcnApply` operand is the function expression serialized with its
  own level, not with the whole function application level. If `F` is a
  level-0 function and `arg` is variable level, Java serializes the `F` operand
  at level 0 even though `F[arg]` as a whole is level 1.
- `@` inside EXCEPT is always an `AtNode`, including when it is itself the
  function part of an application such as `@[2]`. It must not be emitted as a
  built-in operator named `@`.

### Module Instantiation

`processModuleDefinition` handles `M(args) == INSTANCE N WITH ...`.

It:

1. creates formals for the module instance operator,
2. resolves the instancee module/context,
3. marks the instancee as instantiated,
4. builds a `SubstInNode` for explicit substitutions,
5. creates one qualified definition per non-local imported definition:
   `M!Op`,
6. prepends the module-instance formals to the imported definition's params,
7. wraps imported bodies in `SubstInNode` or `APSubstInNode` when substitutions
   exist,
8. imports theorem/assumption definition nodes similarly,
9. creates an `InstanceNode`,
10. creates an `OpDefNode(ModuleInstanceKind)` for the module instance name.

Unnamed top-level `INSTANCE N WITH ...` follows related substitution/import
rules. The Go port should not shortcut this by copying textual definitions into
scope. It should create semantic symbols and substitution nodes.

## Level Checking

Java levels are:

- `0`: constant,
- `1`: state/variable,
- `2`: action,
- `3`: temporal.

SANY level checking is not a recursive `max(child.level)` calculation.

Every `LevelNode` carries:

- `level`: the node's base level after checking;
- `levelParams`: parameters whose substituted level contributes to this node's
  level;
- `levelConstraints`: constraints of the form `param <= level`;
- `argLevelConstraints`: constraints of the form
  `operator-param argument-position accepts at least level`;
- `argLevelParams`: triples `<operator, index, param>` saying that a param
  appears inside an operator argument position;
- `allParams`: all parameters occurring in the node;
- `nonLeibnizParams`: parameters occurring under non-Leibniz positions;
- `levelChecked`: an iteration marker used for recursion.

### OpDefNode Level Check

For a user-defined operator:

1. level-check its body;
2. set the operator's level to at least the body level;
3. derive `maxLevels[i]` from body constraints on formal param `i`;
4. set `weights[i] = 1` iff formal param `i` occurs in the body's
   `levelParams`;
5. derive higher-order `minMaxLevel` and `opLevelCond` from argument-level
   constraints;
6. copy body level params/all params/non-Leibniz params, then remove the
   operator's own formal params;
7. copy constraints and remove constraints on the operator's own formals;
8. copy remaining arg-level params that mention symbols outside the operator's
   own formal parameter list.

For built-ins, the above data comes from `BuiltInOperators.Properties` and the
node is marked already level-checked.

### OpApplNode Level Check

For an application of an `AnyDefNode` (built-in, user definition, theorem/
assumption definition):

1. level-check operator definition and all operands/ranges;
2. reject operands whose level exceeds `opDef.getMaxLevel(i)`;
3. enforce higher-order operator-argument constraints;
4. require bounded quantifier ranges to be at most action level;
5. set application `level = opDef.level`;
6. for each operand with `opDef.weight(i) == 1`, raise application level to
   the operand level;
7. always raise application level to bounded range levels;
8. compose `levelParams`, `allParams`, `nonLeibnizParams`, constraints, and
   arg-level constraints;
9. remove bound symbols from param sets and constraints;
10. run additional temporal/action misuse checks for `[]`, `<>`, `~>`,
    `-+->`, logical operators with mixed action/temporal arguments, and
    quantified temporal formulas with action-level bounds.

For an application of a declared operator parameter rather than a definition:

- the operator itself is added to `levelParams` and `allParams`;
- application level is max(operator level, operand levels);
- arg-level constraints record that this operator parameter must accept the
  levels of the actual operands.

This distinction explains why the current Go `exprLevel` heuristics are
insufficient. User-defined operator application levels depend on formal
weights, constraints, and bound-symbol removal, not simply on operand levels.

### Recursive Operators

`ModuleNode.levelCheck` gives recursive sections a special two-pass treatment:

1. initialize recursive operator max levels and weights conservatively;
2. level-check every operator in the recursive section;
3. accumulate max recursive level and recursive level params;
4. reset recursive operators with accumulated data;
5. level-check the section again.

Go must port this before expecting XML level parity for recursive specs.

## XML Export

Java XML export starts from a fully level-checked `ExternalModuleTable`.

`XMLExporter.specToXMLStream`:

1. creates a DOM document with root `<modules>`;
2. creates a shared `SymbolContext`;
3. exports the root module or all external modules;
4. prepends `context.getContextElement(doc)`;
5. prepends `<RootModule>`;
6. optionally validates the DOM against `sany.xsd`;
7. pretty-prints by Java DOM transformer.

### SymbolContext

When a symbol reference is exported, `SymbolNode.export`:

1. puts the full symbol definition into `SymbolContext` if this UID has not
   already been seen;
2. emits only a reference element such as `<OpDeclNodeRef><UID>...</UID>`.

When a symbol definition is exported into the context, `SymbolNode` emits:

- location,
- level when available,
- symbol-specific payload.

This means XML context entries are discovered lazily from references. Go should
implement the same semantic mechanism. Sorting context entries for deterministic
Go output is fine, but entry contents should be semantic-node exports.

### ModuleNode XML

`ModuleNode.getSymbolElement` emits:

1. `<uniquename>`;
2. `<extends>` with extended module names sorted lexicographically;
3. constant declarations;
4. variable declarations;
5. operator definitions;
6. top-level assumptions/theorems/instances/use-hide nodes;
7. syntactically nested modules from `definitions`, not inherited modules from
   the context.

Module export emits references in the main `<modules>` body. The full module
payload is emitted in `<context>` when the module reference is encountered.

### OpDefNode XML

For `UserDefinedOpKind`, Java emits:

- `uniquename`,
- `arity`,
- `originalOperator`,
- `originallyDefinedInModule`,
- `body`,
- `params` with each `leibnizparam`,
- optional pre-comments,
- optional `recursive`,
- optional `recursiveSection`,
- optional `local`.

For `BuiltInKind`, it emits name, arity, and fake params when present.

For `ModuleInstanceKind`, it emits name and optional local marker.

### OpApplNode XML

`OpApplNode.getLevelElement` emits:

- `location`,
- `level`,
- `<operator>` containing a symbol ref,
- `<operands>` in semantic order,
- optional `<boundSymbols>` with `unbound` or `bound` entries and tuple markers.

The `$Case`/`OTHER` null operand is a special case: `SymbolContext.OTHER_BUG`
causes a null first operand of a `$Pair` to export as a string node `$Other`.

Terminal proof XML uses the terminal proof node's range for `OBVIOUS` and
`OMITTED`, not necessarily the keyword token's range. Thus `PROOF OMITTED`
exports an `<omitted>` location beginning at the `PROOF` token and ending at
the `OMITTED` token. Bare `OMITTED` has the same start/end as the omitted
keyword because the terminal proof node spans only that token.

### Determinism

Java XML context order comes from `HashMap`/`Hashtable` behavior and is not a
semantic contract. Go output should be pretty-printed and deterministic. Tests
should canonicalize:

- context entry ordering,
- unstable UID numbering,
- final module reference ordering where Java order is collection-derived.

Tests must not canonicalize expression operand order, declaration order within
a module, bound-symbol order, or proof step order.

## Current Go State And Mismatch

Current Go code has three useful but conflated representations:

- `SanySyntaxNode`: CST-ish parse result from the production SANY parser port.
- `Spec`, `Module`, `Definition`, and `Expr`: a compact central AST used by
  checking, model checking, XML generation, and ApalacheIR.
- `sanyXMLExporter`: a direct XML emitter that reconstructs Java semantic
  nodes, UIDs, scopes, and levels from the compact AST.

The compact AST is useful for user-facing APIs and downstream tools, but it is
too lossy to be the source of truth for Java SANY XML:

- symbol nodes are not first-class semantic objects;
- contexts are not the same as Java `Context`/`SymbolTable`;
- formal params, operator args, and bound symbols are not represented with
  Java's node identities;
- module instantiation and substitution are flattened too early;
- level checking is approximated by expression recursion instead of Java
  `LevelNode` data;
- XML context entries are allocated by exporter heuristics rather than by
  `SymbolNode.export` references;
- proof and ASSUME/PROVE scoping is reconstructed from syntax fragments.

The new Go architecture should keep the compact AST as a lowered view, but the
SANY conformance path must be:

`source -> SanySyntaxNode -> sanySemModuleGraph -> levelCheck -> XML`.

## Target Go Architecture

Keep all code in `package tlago`, but split files by Java phase:

- `sany_source.go`: source loader, root-relative/library resolution, file/module
  name checks.
- `sany_module_graph.go`: `SpecObj`/`ParseUnit`-like dependency discovery and
  semantic-analysis ordering.
- `sany_semantic_nodes.go`: Java-shaped semantic node structs and node-kind
  constants.
- `sany_context.go`: `Context` and `SymbolTable` equivalents.
- `sany_generator.go`: CST-to-semantic generator, following Java method names
  where useful.
- `sany_builtin.go`: built-in operator semantic symbols and levels.
- `sany_level.go`: `LevelNode`-equivalent checks and constraints.
- `sany_semantic_xml.go`: XML export from semantic nodes and symbol context.
- `sany_lower.go`: optional lowering from semantic graph to the existing public
  `Spec`/`Module`/`Expr` model.

These are file boundaries only, not packages.

The public functions should converge to:

- `ParseSanySyntax(file, source)` -> CST only.
- `LoadSanySpec(root, opts)` -> loaded CST/module graph plus semantic graph.
- `CheckSanySpec` -> semantic graph after level checking.
- `SanyXML` -> XML from semantic graph.
- existing model-checker/Apalache paths lower from the semantic graph or use the
  compact AST as a downstream product, not as the SANY source of truth.

## Migration Plan

1. Add Java-shaped semantic node types and constants in the central package.
   Keep them private until the design settles.
2. Add `sanyContext` and `sanySymbolTable` with Java-compatible lookup,
   insertion order, duplicate checks, and module-name lookup.
3. Move built-in operator allocation into the semantic context. XML, level
   checking, and expression generation should all reference the same built-in
   symbols.
4. Implement a semantic generator that consumes `SanySyntaxNode` and mirrors
   `Generator.generateModule`, `processOperator`, `processFunction`,
   `processLetIn`, `generateExpression`, and the expression helper methods.
5. Initially keep a bridge that lowers the new semantic graph into the existing
   compact AST for commands not under XML conformance. Do not add backwards
   compatibility shims for removed parser APIs.
6. Port level checking from `LevelNode`, `OpDefNode`, `OpApplNode`,
   `LetInNode`, `SubstInNode`, `ModuleNode`, proof nodes, and value nodes.
7. Replace `sanyXMLExporter` internals with semantic-node XML export methods.
   Reuse the current pretty-printer and canonical test helper only for output
   formatting/comparison.
8. Retire heuristic level/context code from the current XML exporter once the
   semantic graph path covers the same behaviors.
9. Expand corpus conformance in windows until all committed `.xml.gold` specs
   match canonical Go output.

## Behavior Suites To Drive The Port

Use behavior-level tests, not tiny helper tests for every Java method.

Core architecture behaviors:

- A root module and all dependencies are parsed before semantic generation, and
  dependencies are semantically generated before dependents.
- File/module name mismatches fail during module loading.
- Inner modules are visible only according to Java `ParseUnit` context rules.
- `EXTENDS` merges non-local declarations, definitions, assumptions, theorems,
  and top-level nodes in Java order.
- Named and unnamed `INSTANCE` create the same semantic symbols and substitution
  wrappers that Java creates.

Semantic graph behaviors:

- Constants, variables, formal params, bound symbols, and new-symbol
  declarations are distinct semantic nodes with Java kind numbers.
- Identifier, infix, prefix, postfix, and zero-arity operator uses resolve to
  the same symbol kinds as Java.
- Higher-order operator parameters generate `OpArgNode` where Java does.
- LET definitions are scoped only in the LET body and exported through
  `LetInNode`.
- Function definitions generate `$NonRecursiveFcnSpec` or
  `$RecursiveFcnSpec`.
- Record, tuple, function, set, CASE, EXCEPT, action, fairness, CHOOSE, and
  quantifier syntax lowers to the Java built-in operator shapes.

Level behaviors:

- Built-in operator levels, argument max levels, and weights match
  `BuiltInOperators.Properties`.
- A user-defined operator application level uses operator base level and
  argument weights, not unconditional max operand level.
- Unused formal parameters do not lift operator application level.
- Higher-order operator arguments create arg-level constraints.
- Bound symbols are removed from level params and constraints after checking
  bound operators.
- Recursive operators use Java's recursive-section two-pass algorithm.
- Temporal/action misuse diagnostics match Java for `[]`, `<>`, `~>`, `-+->`,
  logical operators, and quantified temporal formulas.

XML behaviors:

- XML is pretty-printed and parsable.
- Context entries are full symbol definitions discovered through semantic
  references.
- Module references and context entries are deterministic in Go but compare
  canonically to Java.
- Locations use the CST nodes Java uses, including proof nodes and multi-token
  constructs.
- Pre-comments match Java attachment and stripping behavior.
- The `.xml.gold` corpus under `test_vectors` passes with no Java execution
  when goldens are present.

## Porting Rules

- Prefer a Java-shaped semantic implementation over a clever Go shortcut.
- Keep the implementation in `package tlago`.
- Do not reintroduce a second parser.
- Do not add compatibility shims around old parser names; rename use sites.
- Do not patch Java SANY or Apalache to make Go pass.
- Freeze source-of-truth fixtures under `test_vectors` when tests need them.
- Commit after each coherent chunk or before risky rewrites.

## Source-Level Architecture Map

This section is the detailed map to use while porting. It names the Java
classes and methods that should be mirrored in Go. The Go names do not need to
match byte-for-byte, but the same responsibilities should sit behind the same
phase boundaries.

### Driver Phase

Source files:

- `tla2sany/SANY.java`
- `tla2sany/drivers/SANY.java`
- `tla2sany/drivers/SanySettings.java`
- `tla2sany/drivers/SanyExitCode.java`

The public shell `tla2sany.SANY.main` delegates immediately to
`tla2sany.drivers.SANY.SANYmain`. The real driver is
`tla2sany.drivers.SANY.parse`.

The method decomposition is:

- `parse(spec, fileName, out, settings)`
- `frontEndInitialize()`
- `frontEndParse(spec, out, settings)`
- `frontEndSemanticAnalysis(spec, out, settings)`
- `frontEndLinting(spec, out)`

`parse` does not interleave these phases. It initializes global built-ins,
loads all syntax parse units, semantically generates modules in dependency
order, then level-checks/lints. The Go port should preserve that phase
separation even if all implementation remains in one package.

`frontEndInitialize` is small but important: it calls `Context.reInit()`.
Because the Java built-in context is mutable global state reset for each spec,
the Go equivalent should create a fresh built-in semantic context for each
`LoadSanySpec`/`CheckSanySpec` run rather than share mutable symbol nodes across
specs.

`frontEndParse` delegates all real loading work to `SpecObj.loadSpec`. It
reports syntax/module-resolution errors and aborts before semantic generation.
Go should keep missing-module/file-name/module-name diagnostics in loading, not
in XML export.

`frontEndSemanticAnalysis` owns the semantic-analysis loop:

1. Read `spec.semanticAnalysisVector`.
2. For each external module name, skip it if `ExternalModuleTable` already has
   a context.
3. Fetch the `ParseUnit` from `spec.parseUnitContext`.
4. Create a new `Generator(externalModuleTable, semanticErrors)`.
5. Call `gen.generate(parseUnit.getParseTree())`.
6. Store `gen.getSymbolTable().getExternalContext()` and the `ModuleNode` in
   the `ExternalModuleTable`.
7. Call `moduleNode.levelCheck` when semantic generation succeeded.
8. Mark the last module as the root.

The important Go rule is: semantic generation should consume an explicit module
order produced by loading, not discover modules recursively while generating
XML.

### Module Discovery Phase

Source files:

- `tla2sany/modanalyzer/SpecObj.java`
- `tla2sany/modanalyzer/ParseUnit.java`
- `tla2sany/modanalyzer/ModulePointer.java`
- `tla2sany/modanalyzer/ModuleRelatives.java`
- `tla2sany/modanalyzer/ParseUnitRelatives.java`
- `tla2sany/modanalyzer/ModuleContext.java`
- `tla2sany/modanalyzer/ModuleRelationships.java`

Java separates file-level parse units from module pointers. A `ParseUnit`
represents one external `.tla` file. A `ModulePointer` can point at the
top-level module in that file or at an inner module. A `ModuleRelatives` record
hangs off every `ModulePointer` and stores:

- the immediate outer module,
- direct inner modules,
- string names from direct `EXTENDS`,
- string names from direct `INSTANCE`,
- a `ModuleContext` mapping module names visible at that syntactic point to
  `ModulePointer`s.

`SpecObj.loadSpec` is the root algorithm:

1. `findOrCreateParsedUnit(root, firstCall=true)` loads and parses the root
   file.
2. Repeatedly call `findNextUnresolvedExtention(rootModule)` and
   `findNextUnresolvedInstantiation(rootModule)`.
3. For each unresolved module name, resolve or reuse a `ParseUnit`.
4. Record the parse-unit dependency edge as extendee or instancee.
5. Run `nonCircularityTest`.
6. Resolve newly available module names back into the importing module context.
7. After no unresolved references remain, call
   `calculateDependencies(rootParseUnit)`.

`calculateDependencies` is Java's semantic-order rule. It recursively visits
all extendees first, then all instancees, then the current parse unit. It appends
a parse unit only if not already present. The Go `Spec.SemanticOrder` field is
the direct analogue of `semanticAnalysisVector` and must preserve this rule.

`ParseUnit.parseFile` does four things that should remain separate in Go:

- run the JavaCC parser and get a CST root,
- set CST parent/level metadata,
- call `determineModuleRelationships`,
- enforce exact top-level file/module-name equality.

`determineModuleRelationships(currentModule, parent)` is not semantic
generation. It reads only module structure from the CST:

- allocate `ModuleRelatives`,
- record `outerModule`,
- read direct `EXTENDS` from the `N_Extends` node,
- call `calculateContextWithinParseUnit`,
- walk the module body for direct inner modules,
- walk all remaining CST descendants with `getInstances` to find
  `N_NonLocalInstance` nodes, including those hidden inside LETs or
  substitutions.

The Go loader currently has the first piece of this shape: it now computes a
dependency postorder in `Spec.SemanticOrder`. The remaining work is to preserve
the Java distinction between top-level parse units and inner-module pointers,
because visibility of inner modules depends on source order and enclosing
module context, not on a global map from module name to module.

### Parser/CST Phase

Source files:

- `javacc/tla+.jj`
- `tla2sany/parser/TLAplusParser.java`
- `tla2sany/parser/TLAplusParserTokenManager.java`
- `tla2sany/parser/SyntaxTreeNode.java`
- `tla2sany/parser/OperatorStack.java`
- `tla2sany/parser/JunctionListContext.java`
- `tla2sany/parser/Operators.java`
- `tla2sany/st/TreeNode.java`
- `tla2sany/st/Location.java`

The generated parser creates a concrete syntax tree with Java syntax kinds from
`SyntaxTreeConstants`. It is not the semantic graph. The semantic generator
depends on CST details including:

- exact syntax kind names such as `N_Module`, `N_Extends`, `N_Body`,
  `N_OperatorDefinition`, `N_GeneralId`, and `N_NonLocalInstance`,
- exact child positions within syntax nodes,
- token images and `UniqueString` canonical names,
- comments attached as special tokens,
- locations covering parser-created nodes, not merely leaf tokens,
- indentation-sensitive conjunction/disjunction list grouping,
- operator synonym resolution by `Operators.resolveSynonym`.

`OperatorStack` and `JunctionListContext` are part of the grammar semantics.
The Go parser port must keep using the production SANY parser tables and should
treat these components as source-of-truth behavior, not as parser
implementation detail to be approximated later.

### Context And Symbol Phase

Source files:

- `tla2sany/semantic/Context.java`
- `tla2sany/semantic/SymbolTable.java`
- `tla2sany/semantic/SymbolNode.java`
- `tla2sany/semantic/OpDefOrDeclNode.java`

Java has two nested lookup structures:

- `Context`: one flat table plus a linked insertion list.
- `SymbolTable`: a stack of contexts.

`Context.addSymbolToContext` writes both the hash table and the linked
insertion list. That insertion list is later used by `getOpDefs`,
`getThmOrAssDefs`, and context printing. A Go `map` is not enough for semantic
conformance; the context must also keep insertion history.

`Context.mergeExtendContext` is the source-of-truth for `EXTENDS` imports. It:

- walks the imported context in original insertion order by reversing the Java
  linked-list representation,
- ignores local symbols,
- inserts symbols not already present,
- permits same-symbol duplicates,
- suppresses warnings for symbols with the same original module,
- warns for ambiguous same-class symbols from different source modules,
- errors for different-class conflicts.

`SymbolTable.resolveSymbol` searches stack top to bottom. `resolveModule`
searches for module-name entries in the context stack and then falls back to the
external module table. `addSymbol` and `addModule` own duplicate-name
diagnostics. Go should move duplicate/conflict diagnostics toward this layer as
the semantic graph is ported.

### Built-In Symbol Phase

Source files:

- `tla2sany/semantic/Context.java`
- `tla2sany/semantic/BuiltInOperators.java`
- `tla2sany/semantic/OpDefNode.java`
- `tla2sany/parser/Operators.java`

`Context.reInit` creates a fresh initial context and calls
`Context.initialize`, which creates one `OpDefNode(BuiltInKind)` per
`BuiltInOperators.Properties` entry. Each built-in has:

- canonical name,
- arity (`-1` for variable arity),
- operator level,
- max accepted level per operand,
- argument weights.

`OpDefNode.setBuiltInLevel` derives Leibniz metadata from weights: weight `1`
means the argument is Leibniz and contributes to application level; weight `0`
means it is non-Leibniz. The built-in symbol object should be shared by parsing
of expressions, semantic checking, level checking, and XML export.

Operator spelling/precedence/synonyms are separate from built-in semantic
operator metadata. `Operators.resolveSynonym` canonicalizes source spellings
such as symbolic synonyms before symbol-table lookup. The Go parser/operator
table tests already compare Java operator tables; the semantic graph should
reuse that data rather than maintain a second synonym table.

### Semantic Generator Phase

Source file:

- `tla2sany/semantic/Generator.java`

`Generator.generate(TreeNode)` is the public entry and dispatches only
`N_Module` to `generateModule`. A fresh `Generator` is created per external
module by the driver, but it shares the spec's `ExternalModuleTable`.

The generator's major state includes:

- `context`: current module context,
- `symbolTable`: stack of contexts,
- `moduleTable`: already semantically generated external modules,
- `errors`,
- `excStack` and `excSpecStack` for EXCEPT/@ handling,
- a nested `Function` stack for recursive function definitions,
- `curLevel` and recursive-declaration bookkeeping,
- label/proof context stacks,
- current ASSUME/PROVE goal state.

The method boundaries to mirror are:

- `generateModule`
- `processExtendsList`
- `processVariables`
- `buildParameter`
- `processParameters`
- `processOperator`
- `processFunction`
- `processLetIn`
- `generateExpression`
- `generateExpressionOrLAP`
- `processQuantBoundArgs`
- `processChoose`
- `processBoundQuant`
- `processUnboundQuant`
- `processCase`
- `processSubsetOf`
- `processSetOfAll`
- `processFcnConst`
- `processRcdForms`
- `processAction`
- `processExcept`
- `generateExprOrOpArg`
- `generateGenID`
- `generateLambda`
- `selectorToNode` and selector helpers
- `processModuleDefinition`
- `generateSubst`
- `generateOpArg`
- `processSubst`
- `generateInstance`
- `generateAssumeProve`
- `generateNewSymb`
- `processTheorem`
- `processAssumption`
- `generateProof`
- `generateLeafProof`
- `generateUseOrHide`
- `processRecursive`
- label helpers.

`generateModule` creates a `ModuleNode`, installs it as the current module in
the symbol table, processes `EXTENDS`, then walks module-body syntax in source
order. Inner modules get a copied symbol table plus a fresh context, are
generated recursively, and then their module name is added to the outer module's
symbol table. This is why Go should not use one global `map[string]*Module` as
the semantic source of truth for inner modules.

`processExtendsList` resolves each extendee module, merges its context, copies
assumptions/theorems/top-level nodes, and records the extendee array. The
context merge is the semantic import; XML should see the merged semantic
symbols, not independently gather inherited definitions.

`processOperator` must own formal parameter scope. It pushes a parameter
context, creates `FormalParamNode`s, resolves symbolic operator synonyms,
generates the body while formals are visible, pops the context, then creates or
finalizes the `OpDefNode`. This method is where recursive declarations are
matched to later definitions.

`processFunction` does not create an ordinary operator body. It creates an
`OpApplNode` headed by `$NonRecursiveFcnSpec`, with the function's domain
binders in bounded-bound-symbol arrays. While generating the function body, the
`Function` stack detects recursive calls and switches the operator to
`$RecursiveFcnSpec`. Go XML parity for function definitions depends on
preserving this shape.

`processLetIn` creates one LET context, processes local operator/function/module
definitions into vectors, generates the IN body while the LET context remains
visible, and returns a `LetInNode` containing the local definitions,
instantiations, body, and context. LET definitions are semantic symbols, not
textual substitutions.

`generateExpressionOrLAP` maps CST forms to semantic node forms. Important
rewrites include:

- `N_Number` -> `NumeralNode`
- `N_Real` -> `DecimalNode`
- `N_String` -> `StringNode`
- `N_GeneralId` and `N_OpApplication` -> selector resolution and `OpApplNode`
- tuple -> `$Tuple`
- set enumeration -> `$SetEnumerate`
- Cartesian product -> `$CartesianProd`
- function application -> `$FcnApply`, with multi-argument applications packed
  into `$Tuple`
- bounded/unbounded quantifiers and CHOOSE -> `OpApplNode` with bound-symbol
  arrays and range expressions
- record selection -> `$RcdSelect`
- record constructor/set -> `$RcdConstructor`/`$SetOfRcds` over `$Pair`
- CASE -> `$Case` over `$Pair`, with OTHER represented by a null first operand
- EXCEPT -> `$Except` over `$Pair` and `$Seq`, with `AtNode` for `@`
- action/fairness forms -> `$SquareAct`, `$AngleAct`, `$WF`, `$SF`
- LET -> `LetInNode`.

`generateExprOrOpArg` is the gate for higher-order operators. It checks the
formal parameter arity expected by the main operator. If the expected arity is
zero, it generates an expression. If nonzero, it generates an `OpArgNode` or a
lambda-backed `OpArgNode`. This distinction drives both level checking and XML.

### Instantiation And Substitution Phase

Source files:

- `tla2sany/semantic/Generator.java`
- `tla2sany/semantic/SubstInNode.java`
- `tla2sany/semantic/APSubstInNode.java`
- `tla2sany/semantic/Subst.java`
- `tla2sany/semantic/InstanceNode.java`
- `tla2sany/semantic/ThmOrAssumpDefNode.java`

Java has two related but different paths:

- `processModuleDefinition`: named module definition
  `Name(params) == INSTANCE Module WITH ...`
- `generateInstance`: unnamed `INSTANCE Module WITH ...`

Both paths call `processSubst`, which creates a `SubstInNode` template. The
template first builds implicit substitutions for all constant and variable
declarations visible in the instancee context where the same name exists in the
instancer symbol table. Explicit substitutions then replace or add entries.
`matchAll` verifies all required declarations have substitutions.

`generateSubst` resolves the substitution target in the instancee context. Only
`OpDeclNode` constants and variables are legal targets. If the target arity is
zero, the right-hand side is generated as an expression. If the target arity is
positive, the right-hand side is generated as an operator argument and arity is
checked.

Named module definitions create qualified clones:

- every non-local user definition becomes `Name!Op`,
- module-instance definitions become `Name!Instance`,
- theorem/assumption definitions become `Name!Thm`,
- module-definition formals are prepended to imported definition formals,
- substituted bodies are wrapped in `SubstInNode` or `APSubstInNode`,
- an `InstanceNode` is appended,
- an `OpDefNode(ModuleInstanceKind)` reserves the module-definition name.

Unnamed top-level instances import unqualified definitions. When the instancee
is not parameter-free, Java may either reuse original definitions from
parameter-free origin modules or clone definitions with substitution wrappers.
The exact local/top-level/original-module checks in `generateInstance` are a
major source of XML mismatches and should be ported, not reinterpreted.

### Level Checking Phase

Source files:

- `tla2sany/semantic/LevelNode.java`
- `tla2sany/semantic/ModuleNode.java`
- `tla2sany/semantic/OpDefNode.java`
- `tla2sany/semantic/OpApplNode.java`
- `tla2sany/semantic/SubstInNode.java`
- `tla2sany/semantic/InstanceNode.java`
- `tla2sany/semantic/AssumeNode.java`
- `tla2sany/semantic/TheoremNode.java`
- `tla2sany/semantic/LetInNode.java`
- `tla2sany/semantic/SetOfLevelConstraints.java`
- `tla2sany/semantic/SetOfArgLevelConstraints.java`
- `tla2sany/semantic/ArgLevelParam.java`
- `tla2sany/semantic/ParamAndPosition.java`

Every Java `LevelNode` carries the same data:

- `levelCorrect`,
- `level`,
- `levelParams`,
- `levelConstraints`,
- `argLevelConstraints`,
- `argLevelParams`,
- `allParams`,
- `nonLeibnizParams`,
- `levelChecked`.

The recursive algorithm is not a simple expression tree walk. The data above is
composed, filtered, and constrained differently by each node kind.

`ModuleNode.levelCheck` first handles recursive sections. For each recursive
section it initializes recursive operator max levels/weights conservatively,
level-checks the section, accumulates max recursive level and params, resets the
operators, then level-checks the section again. Only after this special pass
does it check inner modules, definitions, theorem/assumption definitions, and
top-level nodes.

`OpDefNode.levelCheck` checks the body, derives formal `maxLevels`, argument
weights, higher-order constraints, all params, non-Leibniz params, and removes
the operator's own formal params from outward-facing constraints.

`OpApplNode.levelCheck` behaves differently depending on whether the operator
is an `AnyDefNode` definition or a declared operator parameter. Definition
applications use the operator's base level, formal weights, max-level
constraints, bound range levels, and parameter filtering. Applications of
declared operator parameters add the operator itself to level params and record
arg-level constraints.

Go's current `tlaLevel` computation is useful scaffolding but not an
architecture match. XML level parity requires porting `LevelNode` data and
node-specific `levelCheck` methods.

### XML Export Phase

Source files:

- `tla2sany/xml/XMLExporter.java`
- `tla2sany/xml/SymbolContext.java`
- `tla2sany/xml/XMLExportable.java`
- semantic node `export` / `getSymbolElement` methods.

`XMLExporter.parseSpec` runs the ordinary driver and returns an
`ExternalModuleTable`. `specToXMLStream` creates a DOM document, creates a
shared `SymbolContext`, exports module references, inserts the context element,
inserts `RootModule`, optionally validates, then pretty-prints with Java's DOM
transformer.

`SymbolContext` is intentionally lazy. Exporting a symbol reference calls
`context.put(symbol)`. The first `put` for a UID exports the full symbol
definition into the context; later references emit only the ref. Some context
entry order is Java collection behavior, so Go may sort context entries for
determinism and tests should canonicalize entry order.

The XML exporter should be thin:

- semantic node refs decide when context entries exist,
- semantic node definitions decide payload shape,
- XML formatting decides only whitespace and deterministic ordering where Java
  order is unstable.

The current Go XML exporter still reconstructs semantic facts from the compact
AST. It now at least uses `Spec.SemanticOrder` for the external module list,
but the target is to replace its internal maps with semantic node export
methods.

## Current Go Architecture Alignment

The Go port now has the following source-of-truth layers:

- `ParseSanySyntax` produces the SANY CST using the production parser port.
- `LoadSanySpec` loads root-relative and library modules and computes
  `Spec.SemanticOrder`, matching Java's extendees-first, instancees-second
  dependency postorder.
- `CheckSpec` performs a compact-AST semantic bridge check.
- `SanyXML` emits Java-shaped XML from the compact AST plus XML-specific symbol
  reconstruction.

The next alignment milestone is to insert an internal SANY semantic graph
between `LoadSanySpec` and `CheckSpec`/`SanyXML`:

`SanySyntaxNode -> sanyParseUnit graph -> sanySemanticGraph -> levelCheck -> XML/lowered Spec`.

`Spec`, `Module`, `Definition`, and `Expr` remain useful as a public, lowered
view for model checking and future transformations. They should not remain the
source of truth for SANY conformance.
