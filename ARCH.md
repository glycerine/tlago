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
