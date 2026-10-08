# TLC Architecture Notes for the Go Port

## SANY XML driver and schema boundary

XMLExporter has its own exit enum and structured exception; a semantic spec
failure is SPEC_PARSING_FAILURE (2), not the SANY semantic exit (4). The XML
parse driver uses validAstSettings: generation and level checking with linting
disabled, reporting only ERROR output. The library moduleToXML entry and command
runner share that driver and preserve code, message and nested failure. XML 1.0
control-character rejection is code 7 and is not an exporter bug.

The schema is embedded from the exact pinned Java sany.xsd. Non-offline export
uses libxml2's complete XSD engine through xmllint, requiring it on PATH and
failing explicitly if unavailable. Validation disables network access and
does not use test-specific element checks. The original module tests also
validate captured stdout independently. Terse output omits formatting while
retaining character data. The source four-option decimal library path and
complete XML help printer are exercised directly by their translated tests.
This establishes the repaired source contracts, not unrestricted semantic graph
or XML equality beyond the original assertions.

## Missing EXTENDS module aborts generation and uses checked boundaries

When no ModuleNode resolves, Generator records INTERNAL_ERROR 4003 at that
EXTENDS occurrence, with `Could not find module %s` and the actual UniqueString
parameter, then throws AbortException. Do not finish the extendee array or the
body, or resume an enclosing body's later units. Prior inherited vectors remain
copied, while the extendee array keeps its constructor value on the failed module.
The exception retains the triggering details and the shared semantic diagnostic
log; unwinding preserves earlier parent and external-module diagnostics.

The semantic driver catches this internal family, prints the fatal header,
captured port stack and retained Errors, then chains SemanticException. Its
message comes from AbortException.toString, while the abort's ordinary Exception
message and cause are null. Wrapper accessors preserve the same detail pointer,
source-log pointer and cause. The legacy SANY entry point catches this checked
family and returns ERROR (-1). Unexpected runtime failures still become
FrontEndException and propagate. Ordinary semantic diagnostics retain their
existing legacy return behavior. The exceptions use the generic Exception base
with real Go frames; no JVM frames are fabricated. Throwable printing honors
these source toString overrides.

All 150 direct comparison rows match across 18 root/one-level/two-level abort
scenarios, including earlier recoverable errors, repeated/mixed EXTENDS, partial
vectors, stopped symbol scopes, nullable messages, log/detail values and wrapper
identities. Three semantic-driver scenarios match all 12 output rows with native
stack-frame lines excluded from cross-language text comparison; both stacks are
observed as present. Three legacy boundary observations match checked failure,
runtime cause propagation and normal success. The retained 24,529 whole-module
rows also match. Earlier redefinition diagnostics exposed by retained logs now
keep their source name parameter. No permanent tests or original-method credit
added. General Errors ownership, inherited level checks, selectors, visitors and
evaluator sharing still require their broader audits.

## Missing EXTENDS context logs an error and continues

`Generator.processExtendsList` distinguishes an absent ModuleNode from an absent
context on a resolved module. For an absent context, log INTERNAL_ERROR 4003 with
``Couldn't find context for module `%s'.`` and the actual UniqueString parameter.
Continue copying assumption, theorem and top-level vectors, then finish the
extendee array and generate the body. Java returns an AbortException object from
Errors.addMessage here but does not throw it. Go now preserves that control flow
rather than invoking Context.mergeExtendContext with a null argument.

Select the diagnostic position by the EXTENDS occurrence index. Repeated module
names retain distinct token locations and therefore distinct ErrorDetails.
Searching by name alone wrongly assigns every occurrence the first location and
causes error deduplication to drop later diagnostics. Native implicit EXTENDS has
no source token and retains the module position fallback.

All 118 observed Java/Go rows match across 20 root/nested-module scenarios with
missing or present contexts, forward/reverse/repeated EXTENDS, copied vectors,
body node UIDs, locations, messages, parameters and actual UniqueString parameter
types. No permanent tests or original-method completion credit added. Missing module resolution now follows the thrown-abort boundary described below.

## Module generation retains canonical EXTENDS diagnostics

Each direct EXTENDS context merge contributes its canonical ErrorDetails to the
module's diagnostic sequence before inherited vectors are copied. This preserves
source traversal order and source-pointer suppression through diamonds and
instances. Native import scans retain expression metadata only; they no longer
reconstruct conflicts from syntax positions or originating module-name strings.
The native API keeps its existing conflict prefix in `Diagnostic.Message`, while
`SANYMessage`, ranges and structured parameters retain the exact source details.
Unused kind-to-class inference and its diagnostic wrapper have been removed.

All 24,529 manual comparison rows match Java across 68 whole modules, retaining
the earlier 47 and adding all constant/variable/operator/assertion conflict pairs,
multiple conflicts in forward/reverse/repeated EXTENDS order, local definitions
and shared original definitions. The 25 explicit diagnostic detail rows compare
incoming locations and complete structured parameters. Existing assertions remain
unchanged; no permanent tests or original-method credit added. Inherited level checks and evaluator graph sharing remain pending. Missing contexts log an error and continue, as
described below.

## EXTENDS Context merge uses actual node classes and names

`Context.mergeExtendContext` snapshots the source Pair history before traversal,
then processes it oldest-first. Derive each lookup key from the symbol's actual
name and class: ModuleNodes use ModuleName keys; every other node uses a plain
name. An original insertion key is not retained by the merge. This matters after
Context duplication, and for self-merges that add a previously absent key.

Skip local nodes, retain the first accepted binding, and suppress identical
pointers before comparing concrete classes. Definitions from the same immediate
source in a parameter-free module avoid a warning. Otherwise, equal concrete
classes warn and different classes error. Diagnostic descriptions follow
kindOfNode's class checks; warning descriptions both use the existing node.
Read the current syntax trees for diagnostic locations, including null failures.
Null contexts and null history or conflicting lookup symbols throw typed
NullPointerException rather than returning success or silently skipping entries.

All 1,115 direct Java/Go rows match across six actual node classes, constant and
variable declarations, three supplied-key forms, separate/self/identity merges,
lookup retention, diagnostic messages/parameters, local formals and null cases.
The retained 16,872 whole-module rows also match. Existing scaffold helpers now
supply actual operator/declaration classes with their original requested kinds;
assertions and cases remain unchanged. This adds no original-method credit.
Module generation now retains the canonical merge diagnostics in direct extendee
order; native expression metadata no longer reports competing conflicts. See the
caller integration section below.

## Context duplication preserves source history and lookup quirks

`Context.duplicate` copies the complete Pair history, sharing its symbol nodes
while allocating independent entries. It rebuilds the Hashtable newest-first,
using each node's plain name rather than the original insertion key. Consequently,
the oldest history entry wins a repeated name, and module-name keys become plain
symbol keys. Repeated duplication preserves these results. Keep this behavior
even though ordinary insertion uses the newest binding and distinct module keys.
Hashtable collision and rehash order follows that reverse insertion order.
Null contexts or null symbols in history throw NullPointerException.

All 1,578 manual Java/Go rows match across empty contexts, 39 insertions,
collisions/rehashes, replacements, module nodes and keys, a key differing from
the symbol name, repeat copies and copy-only mutations. The retained 16,872
whole-module rows still match. No permanent tests or original-method credit added.
The remaining general Context/graph audits remain incomplete.

## Definition paths and counted UniqueString joins

`OpDefNode.getLocalName`, `hasPath` and `getPathName` now read the actual
compound identifier array. `getCompoundId` returns that same array when supplied,
or a fresh one-element array containing the full name otherwise. A full name
containing `!` does not itself establish a path. Preserve null receivers, null
array elements and empty-array failures rather than reconstructing name segments.

`UniqueStringJoinN` ports the counted Java overload. With source assertions
enabled, a nonpositive count or count beyond the array fails. The source ignores
its delimiter argument and uses literal `!`. Leading nulls leave the accumulator
null; a null following text fails after interning the separator-bearing prefix.
Successful joins also intern each intermediate prefix, preserving token allocation
order. The variadic wrapper preserves nullable-array failure; an explicitly
non-null empty slice reaches the counted overload assertion. Singleton compound
arrays therefore fail `getPathName` under source assertions. Do not replace these
source behaviors with string splitting or a conventional joining helper.

Manual Java/Go observation matches all 16,872 rows across the retained 47 whole
modules, ten compound-array shapes, three delimiters, counted-prefix bounds,
array mutation/identity, null receivers and intermediate token allocation. This
adds no permanent tests or original-method completion credit.

## Original-definition comparison and cached parameter freedom

`SymbolNode.sameOriginallyDefinedInModule` now follows the source concrete-class
rule: only two actual OpDefNodes or two actual ThmOrAssumpDefNodes can match.
Their immediate `getSource` pointers must match. Read the originating module of
that shared source, treating a null module as parameter-free, then query the
module's cached constant/variable declaration arrays in source short-circuit
order. Do not scan live context history or recursively follow source chains.
Declarations and modules do not match through this method, even when the supplied
pointers are identical; callers retain their earlier identity shortcuts.
Null symbols preserve Java NullPointerException failures, including Go typed nulls.

All 8,318 observed rows match across 47 whole-module scenarios and 17 comparison
pairs per scenario. Cover live-context mutation after declaration caches freeze,
fresh modules sharing that context, same/different source pointers, one-hop
source chains, different definition names/kinds, theorem versus operator classes,
declarations, modules and nulls. Existing Context and SymbolTable callers use the
same helper. No permanent tests or original-method credit added. Inherited levels,
visitors and evaluator sharing remain pending.

## EXTENDS vectors and cached extension sets

Production EXTENDS resolves actual ModuleNodes through the SymbolTable, merges
their contexts in direct source order, then copies their assumption, theorem and
top-level vectors. These copies retain the original node references and preserve
repeated entries, including diamonds. They read source vectors rather than lazy
getter arrays. `copyTopLevel` does not copy `instanceVec`: inherited instances
appear in the top-level vector while `getInstances` reports the local vector.
Preserve Java's code even though the instance accessor's comment says otherwise.

`createExtendeeArray` copies its supplied vector. Direct and recursive extension
sets have separate lazy caches; each returned mutable set keeps its identity.
Recursive sets union the source modules' cached recursive sets. Replacing the
extendee array does not invalidate either cache. Null arguments and receivers
retain Java NullPointerException failures. Constructor extendees start as an
empty nonnull array. Native imported expression metadata now retains the actual
accepted operator/theorem definition nodes from the merged context.

All 7,519 observed rows match Java across 47 whole modules and accompanying cache
and null observations. The comparison includes actual vectors, shared node UIDs,
repeated EXTENDS, diamond inheritance, private definitions, inherited named facts
and instances, nested contexts, direct/recursive set members, returned-set aliasing,
array-copy ownership and cache lifetime. No persistent tests or original-method
completion credit are added. Inherited levels, visitors and evaluator graph
sharing remain pending; this is bounded evidence.

## Module collection snapshots and inner-module registration

Context declaration vectors filter actual OpDeclNode classes by constant or
variable kind in Hashtable enumeration order. Inner-module vectors use the same
enumeration and concrete class filter. Operator and theorem definition vectors
follow the separate Pair history chain newest-first, retaining replaced entries;
operator vectors exclude builtin and module-name kinds. ModuleNode lazily caches
a reversed copy of declaration/operator/theorem vectors, and a forward copy of
inner-module vectors. Subsequent mutations do not invalidate those arrays, and
callers receive the same mutable array. `isParameterFree` queries those cached
constant/variable arrays with Java's short-circuit order.

Completed inner ModuleNodes now register in the enclosing SymbolTable, after
body generation. A shared external-module table is populated in loader order
after standard provenance assignment and before level checks. Inner symbol tables
share that table. Actual registration now owns E4223 and rejected bindings;
preserve the native message separately from the exact source SANY diagnostic.
The private single-module checker prepares preceding dependency graphs in loader
order rather than attempting INSTANCE against native metadata alone.

A forward inner-module INSTANCE reports missing context after named formals and
before substitution-template allocation, preserving relative UIDs and leaving
the target uninstantiated. Missing-context source locations use the actual module
name token. All 5,697 observed rows match across 43 whole-module scenarios,
including hash rehashing, duplicate history entries, shared declaration/definition
identities, cache mutation, returned-array aliasing, nested modules and forward
references. Seventeen loader/front-end observations also match Java, including
external/inner module name conflicts. These comparisons add no original-method
completion credit and do not establish complete module inheritance, level,
visitor or evaluator parity.

## Qualified instance references and rejection order

Qualified GeneralId operator arguments now validate the actual imported symbol's
arity before allocating `OpArgNode`, matching `selectorToNode`. Preserve terminal
module-name rejection before that final arity check: an instance name remains an
incomplete operator even when used in a higher-order argument position. Ordinary
prefix-application checks use the actual `ModuleInstanceKind` node and arity, so
an invalid applied prefix is diagnosed before resolving an unknown final name.
Native signature metadata alone does not establish the semantic node's class.

Qualified instance facts resolve the actual module-name node. DEF references with
unapplied instance prefixes retain the actual qualified definition in the
USE/HIDE/BY definition array. Each applied DEF selector is rejected with the
source diagnostic before resolving that item. A failed earlier lookup stops
further processing; preserve the caller's follow-up diagnostic. Label and operand selectors still require their distinct source paths.

All 1,750 observed rows match Java across 41 whole-module scenarios, including
actual application operands/OpArg symbols, relative allocation IDs, source
parameters, wrappers, proof fact/definition arrays, valid qualified arguments,
wrong arities, missing names and invalid applied prefixes. Scratch comparisons
add no original-method completion credit. General subexpression selections,
qualified fixity operators, inherited levels and evaluator graph sharing remain
pending; this comparison is not a full semantic-parity claim.

## Named INSTANCE definitions and caller ownership

`processModuleDefinition` now constructs actual qualified operator, theorem and
module-name definitions in context enumeration order. Prepend the instance's
actual formals to a fresh parameter array, retain source/label pointers, and wrap
bodies in `SubstIn` or `APSubstIn` only when substitutions exist. The substituted
operator overload retains the source compound identifier array. Module-name
constructors preserve null versus empty parameters, arity, originating module,
source pointers and default flags; their `defined` flag remains false, matching
Java's code despite its constructor comment.

Allocate `InstanceNode` after imported definitions, then allocate the instance
name symbol. Top-level imports enter module definitions and instance vectors.
LET owns imported definitions plus the module-name symbol and its instance array.
Proof DEFINE owns only the module-name symbol in `DefStepNode`; its instance goes
into the enclosing proof's separate instance array, and its imported definitions
stay in the proof context. `LetInNode.getChildren` includes definitions, instances
and body in that order. Accepted native metadata retains the actual symbol;
canonical constructors own registration diagnostics.

Named formal parameters retain their complete source declaration syntax and
individual locations, including duplicate names and operator placeholders.
Qualified applications retain the actual imported operator. A bare module-name
expression is rejected before application construction; a fact or DEF reference
retains the module-name node itself, as Java's selector does.

All 605 observed rows match Java across 18 whole modules. The bounded comparison
covers LOCAL/parameterized imports, higher-order formals, empty targets, nested
named instances, LET and proof ownership, parameter UID/location identity,
compound identifiers, duplicate registrations and bare expression/fact use.
It adds no persistent tests or original-method completion credit. General
qualified selectors and operator arguments, EXTENDS inheritance, inherited level
checks, visitors and evaluator graph sharing still require faithful port work.

## Unnamed INSTANCE imports and ownership

Production unnamed INSTANCE generation now uses actual context enumeration,
filtering local, builtin and module-name definitions. Reuse parameter-free
operators where Java does; LOCAL copies and parameterized bodies retain the
source constructor/module rules. SubstIn and APSubstIn wrappers share the actual
substitutions, original parameter arrays, labels and source pointers. Preserve
the source's asymmetric theorem-registration branch for a parameter-free original
inside a parameterized target. Operator imports always enter the module definition
vector; theorem imports enter it only for top-level INSTANCE generation.

Allocate the InstanceNode after importing definitions. Top-level instances enter
both the module instance vector and top-level vector; proof instances retain their
body syntax and nullable step number, with numbered symbols pointing back to the
same instance. Proof contexts retain actual accepted imported symbols. Native
metadata now shares those pointers and canonical registration owns diagnostics.
Targets retain the source instantiated flag. getInstances copies and caches the
vector once, preserving the source cache lifetime and supplied instance identities.

Use the syntax node's raw Zero-array presence for Java local(), including proof
N_NonLocalInstance nodes whose nonnull Zero array makes localness true. Do not
replace this with a LOCAL-token check. Proof INSTANCE keeps its body syntax;
source replaces full-step syntax on theorem and USE/HIDE nodes separately.

All 347 Java/Go observation rows agree across 12 valid whole modules and one
retained parser-rejection input. Cover top-level/shared/LOCAL/parameterized
imports, assertion wrappers, repeated imports, conflicts, chained instances,
default/explicit substitutions, numbered/unnamed proof instances, exact UID order,
source/body/module identity, labels, module vectors and proof-context backlinks.
Anonymous LET INSTANCE syntax is rejected by both parsers and gets no semantic
completion credit. The source LET generation branch is connected; named module definitions now
retain actual caller-owned arrays as described above. EXTENDS inheritance,
complete instance vectors, inherited level checks, visitors and evaluator graph
sharing remain pending. No permanent tests or original-method completion credit added.

## Production INSTANCE substitution templates

The production INSTANCE generator now retains an actual SubstIn template and
formal array on the source Instance. Prepare parameter contexts before the
template; allocate the template before default expressions. For each resolved
default retain the actual application or operator argument and declaration, with
implicitness and null expression syntax. The native default-expression view
shares that same generated node rather than regenerating its identity.

Generate explicit RHS nodes before duplicate checking and mutate/append actual
Subst objects using the source wrapper operations. Canonical constructors and
mutation own their diagnostics; legacy reporting is retained only when earlier
imports have no actual graph. Missing declaration/expression identities keep the
template incomplete rather than installing fabricated substitutions. Contexts
and label-generation settings restore their callers' state after completion.
INSTANCE RHS labels now invoke canonical guards even outside a definition,
retaining error 4333 and the shared nullLabelNode without generating its body.

Direct comparison with Java Generator.processSubst matches all 51 rows across
17 scenarios: actual template/module references, allocation deltas, ordered slots,
implicit/explicit flags, expression/operator kinds and references, full diagnostic
codes/messages, duplicate/missing/illegal substitutions, discarded defaults,
arity checks, lambdas, functions and label rejection. This verifies template
integration, not all INSTANCE import generation. Named-instance formal syntax,
imported definition/source/module ownership, instantiated flags, module/proof/LET
vectors, inherited levels/visitors and evaluator sharing still require completion.
No new permanent tests or original-method completion credit is added.

## SubstIn and APSubstIn wrapper construction

The two wrappers now retain actual Subst arrays, body and instantiating/
instantiated module identities. Copy constructors share the template's array;
SubstIn also accepts an APSubstIn template. Null arrays stay null. Access the
template before allocating the new UID; report null-body diagnostics afterward,
preserving the distinct source kinds, error codes and message line breaks.
A null syntax with a null body raises the source typed NullPointerException.
Children are fresh arrays containing the body first, then substitution
expressions, with null expressions retained and null Subst entries rejected.
Indexed access preserves null-array and bounds failures.

Default constructors allocate the wrapper before generating defaults in the
supplied declaration order. Resolve actual symbols in the instancer table;
missing names create no slot. Variables and zero-arity constants get actual
zero-argument applications; operator constants get actual OpArgNodes with the
instantiating module. Subst objects remain implicit and have null expression
syntax. Preserve matcher diagnostics and aborts rather than synthesizing nodes.

Explicit replacement mutates the existing implicit Subst, so copies observe the
same expression/syntax/implicitness changes. Duplicate explicit substitutions
leave it unchanged and report the source diagnostic. Appending copies the array
and installs it only on that wrapper. Unknown/non-declaration targets add no slot.
Completeness checks preserve declaration order, locations and messages. APSubstIn
uses the source generic 4004 codes here; SubstIn uses specialized 4240/4241 codes.

All 518 Java/Go rows match across 128 copy-constructor cases, 48 default cases
and four mutation/completeness sequences. The observer supplies Java's actual
F, c, v declaration order to both constructors; module-vector enumeration is a
separate caller responsibility. No permanent tests or method completion credit
are added. Actual INSTANCE generation, complete inherited levels/visitors,
module/proof/LET vectors and evaluator graph integration remain pending.

## InstanceNode and substitution storage

InstanceNode now retains its nullable interned name, localness, target module,
syntax and nullable step name. Supplied parameter and Subst arrays remain shared;
null arrays become nonnull empty arrays. Construction allocates exactly one
semantic UID. Its level accessor returns zero, as an INSTANCE proof step does
not contribute to the enclosing proof's level. Children are a fresh array of
substitution expressions, including null expressions; a null Subst entry raises
the source typed NullPointerException. Module and parameters are not children.

Subst is separate from SemanticNode and allocates no UID. Preserve declaration,
expression, expression syntax and implicitness independently. setExpr updates
both expression and implicitness. getSub searches in array order by declaration
identity, including a null declaration, and returns the first match. Null arrays
and null entries retain their source failures.

All 582 rows across 192 InstanceNode constructor combinations and Subst mutation/
lookup observations match Java, covering supplied-array mutation, null entries,
step-name reset and UID allocation. These are temporary comparisons, not new
permanent tests or original-method completion credit. The existing translated
TestInstanceNode level-diagnostic method remains unchanged. Actual generation,
SubstIn/APSubstIn wrappers, complete level inheritance, visitor dispatch and
module/proof/LET instance vectors remain pending; no fabricated instance graphs
are installed to stand in for those features.

## AP proof-step goals and declaration lifetimes

Create the provisional named step goal before generating AP. The outer AP and
its labels retain that exact goal; nested AP goals stay null. The caller owns
the outer declaration context during generation, captures it afterward and
removes it before registering the theorem definition in the enclosing proof.
Restore the caller's symbol metadata after body generation. Preserve ordinary
ASSUME's shared marker and boxed-AP scope diagnostics.

For ASSERT AP, push the same declaration context only for the step's proof,
including terminal BY and nested proofs, then pop it before the theorem owner.
For SUFFICES AP, generate its proof without that context and push it afterward
for subsequent enclosing-proof steps. Before leaving the proof, merge retained
SUFFICES contexts into the proof context and pop them in reverse stack order.
Set the outer AP's inProof false before allocating its owner; nested AP flags
retain their source state. AP, named-definition and theorem suffices flags remain
consistent. Missing children still prevent canonical owner completion.

Context.content uses Hashtable enumeration, independently of Context's linked
Pair insertion history. Preserve Java string hashing, separate symbol/module
keys, default bucket capacity/threshold, head insertion, replacement position,
rehash traversal and descending bucket enumeration. SUFFICES context merging
uses this actual order and actual symbol references. A null receiver raises the
source typed NullPointerException. Other Context iteration callers and complete
live/concurrent Hashtable enumeration still require their own port audits.

All 2,281 complete Java/Go rows agree across 29 AP proof modules, including
unsorted context order, goal/reference identities, scope failures, labels,
boxed AP and contexts crossing collision/resize thresholds. Java's observer
reinitializes Context to match the Go frontend API; retaining the shared static
ASSUME marker otherwise exposes different observer setup. All 37 direct context
observations also agree. No original-test or full-workspace credit is added.
InstanceNode identities, general selectors, inherited level checks and visitors
remain pending. TAKE/PICK graphs and binding lifetimes are covered below.

## Structured-proof generation and contexts

Traverse actual proof syntax in source order while consuming each projected
statement once. Enter a fresh proof context before its steps. Exit nested proofs
before allocating their containing theorem owners, then pop the proof context
before constructing its NonLeafProofNode. Preserve the context with the parent
node. A missing canonical step or subproof keeps the parent graph incomplete.

Every named step pushes a label scope and allocates a provisional theorem symbol
before its body, including non-theorem steps whose provisional symbol is discarded.
DEFINE produces a DefStepNode retaining actual definitions; ordinary operator
proof definitions also enter the module definition vector. USE/HIDE retain actual
vectors, step names and validation order. Numbered non-theorem symbols register
with their real step backlinks and popped labels. The actual definition-step
symbol is accepted in DEF; other numbered non-theorem steps retain both source
rejection diagnostics. Keep the numbered symbol itself in the DEF array.

ASSERT, HAVE, CASE, WITNESS, QED and ordinary-expression SUFFICES construct their
source bodies and named definitions before proof processing. Their separate
TheoremNodes follow the complete subproof. Replace statement syntax with the full
step syntax where Java does, including the cached Go location. Expression lookup
retains the actual accepted symbol rather than rebuilding a proof-step operator.

For @ infix shorthand, generate the current RHS first, then create $Nop around
the already-generated previous RHS, then the infix application. Reuse the exact
previous node. A parameter-free named label selection similarly wraps the actual
LabelNode, preserving its identity and scope. Parameterized labels, general
operand selectors and INSTANCE selections still need their real binding graphs.

InstanceNode/module-definition instances, inherited level checks and visitors
remain unfinished. Native diagnostic coverage does not complete these graphs.
The 21-module, 993-row source comparison covers graph kinds, syntax, locations,
UID order, nested owners, context and reference pointers, plus full diagnostics;
it adds no original-method or full-workspace completion credit.

## Structured-proof constructor ownership

NonLeafProofNode retains the supplied step array, module-definition instance
array and context identity. Its children contain only steps: null/empty steps
return null; populated steps return a fresh copy, including null slots.
DefStepNode retains its definition array and nullable UniqueString step number.
Its children always copy definitions to a fresh array, including a nonnull empty
array; a null definition array raises NullPointerException. Preserve null slots
as null graph references rather than typed nil interfaces.

These constructors use the shared semantic UID counter and source syntax/location
handling. Numbered non-theorem steps now have their own OpDefNode constructor: kind 37,
zero arity, actual step backlink and module/table identities. Ordinary body,
formal parameters and argument-level arrays remain null; defined/local/Leibniz
flags remain false. A null symbol table fails after allocating the node. Register
only after installing the backlink. getStepNode returns that exact node;
getChildren still returns a fresh one-slot array containing the ordinary body,
which is null for these symbols. The step is a walkGraph link, not a child-array
entry. The 32 constructor cases and three binding collisions match all 72 Java
rows, including exact diagnostic messages and retained earlier bindings.

InstanceNode identities, inherited level checking and visitors remain unfinished.
Do not create a parent proof graph before its complete children and scope are
available.
The 32-case, 80-row Java comparison verifies only these constructors and accessors;
it adds no original-test or whole-feature completion credit.

## USE/HIDE vectors and BY leaf proofs

UseOrHideNode retains supplied fact/definition arrays and ONLY flags. Its nullable
step name uses the shared UniqueString identity. getChildren returns null for
null/empty facts and a fresh copy otherwise. Source factCheck returns immediately
for USE or null facts. HIDE checks only OpAppl facts, requiring an actual theorem/
assumption definition operator; other fact kinds retain source behavior. Preserve
source kind/cast checks and null failures.

Build vectors by consuming each reference immediately after generation. Actual
failed expression nodes keep their fact slot; invalid DEF entries and unavailable
modules do not add slots. DEF entries use actual accepted symbol identities. Missing
canonical imports or qualified selected definitions leave the vector incomplete.
Top-level USE/HIDE nodes enter the module's ordered vector after fact generation;
HIDE validation runs after all generation diagnostics, then empty-command checking.

BY creates a temporary USE node, then a LeafProofNode sharing its arrays and ONLY
flag. The temporary allocation is part of source UID order even though that node
is discarded. Retain the actual leaf by its terminal syntax, and use it when the
outer theorem completes rather than allocating another proof. Complete theorem
owners follow their actual BY leaf and precede the next module unit.

Binder/INSTANCE proof graphs, general qualified selections and native-only
import identities remain incomplete. Existing diagnostic fallbacks are separate from
canonical graph coverage. Inherited level checks, visitors, EXTENDS vectors and
shared evaluator ownership still require porting; bounded leaf observations do
not establish complete module or original Java test-suite completion.

## Theorem statements, AP goals and proof scope

Create a named theorem's provisional ThmOrAssumpDefNode before its statement.
For AP statements, the outer AP and ordinary/labeled AP labels retain this actual
goal; nested AP nodes retain null goals. Ordinary expression theorems do not set
currentGoal. Complete/register the definition after the body, pop its label set,
and append the definition before proof generation. For AP, temporarily remove the
assumption context for registration, then push the same context back for the proof.
A proof-local NEW declaration may shadow the theorem name; the enclosing module
binding still determines accepted registration and exported scope metadata.

The top-level caller owns the AP context until proof processing ends. Nested AP
contexts close at body completion. Preserve actual NEW declaration bindings and
ordinary ASSUME marker visibility through the proof; restore symbols/context and
reset only the outer AP's inProof afterward. Clause counters and labels retain
source timing. Labeled AP syntax retains its own AP body rather than generating
the native quantifier view. Generate its body before resolving label formals and
constructing the LabelNode. AP theorem references in expression contexts produce
E4355 and the actual null operator application before allocating a normal call.

Allocate the theorem owner after a complete statement and proof. OBVIOUS and
OMITTED produce actual LeafProofNodes with nonnil empty fact/definition arrays,
source flags and syntax. Complete canonical BY vectors now produce their real leaf proofs and owners.
Missing canonical BY vectors and incomplete structured proof graphs keep owners incomplete
rather than pretending the proof is absent. Missing
statement children or earlier native-only imported identities also keep owners
incomplete. Complete inherited level checking, visitors, imported/INSTANCE
identities, EXTENDS vectors and other top-level owners remain separate work.

The bounded 47-module comparison establishes the observed constructor/generation
paths only. It is not original Java method completion or complete module parity.

## Assumption generation

For source expression assumptions, generate the body before allocating either a
named ThmOrAssumpDefNode or its AssumeNode owner. Named bodies use a fresh label
scope; completion registers the definition, attaches the popped labels and appends
it to module definitions. The separate owner is then allocated and appended to
both assumption and ordered top-level vectors, retaining the definition backlink.
Unsuccessful duplicate registration retains the earlier binding while the new
owner and definition remain in their source collections. Actual accepted
assumption definitions are carried through symbol lookup and scalar applications.

A failed expression can supply the generator's actual failure node. Missing
canonical child graphs or earlier native-only imported identities keep the owner
incomplete. Legacy scope metadata changes only for an accepted canonical binding;
the fallback remains for paths without those identities. Legacy label checking
runs after generation so canonical labels retain their own source diagnostics.

Structured proof generation, proof-step goals, inherited assertion vectors,
other top-level owners and complete inherited level checking/visitors remain
separate unfinished work. A partially populated module vector is not evidence of complete module
construction. Bounded assumption graph observations do not add original Java
method completion credit.

## Assertion owners and leaf-proof constructors

Concrete AssumeNode and TheoremNode constructors retain actual module, statement,
definition and proof pointers. Named definitions point back to their owner. AXIOM
is detected from the assumption syntax's first heir. The theorem constructor
checks statement identity against its definition after setting the backlink,
preserving source assertion failure order. Child arrays contain only the statement
for assumptions, and statement followed by an optional proof for theorems. Each
call creates a new array; absent statements remain null entries. Theorem name
access distinguishes an unnamed theorem from an empty definition name.

LeafProofNode construction aliases facts and definition arrays and retains omitted
and ONLY flags. Its child getter returns null for null/empty facts and a fresh copy
of nonempty facts. Module assertion additions append the same owner to the typed
assertion vector and ordered top-level vector. Each getter caches its first array;
mutating that array persists and later vector appends do not invalidate it.

Assumption and complete theorem statement generation populate these owners and
collections. Structured proof generation, EXTENDS collection inheritance, other
top-level owners, inherited level checking and visitors remain unfinished. Direct constructor observations are separate from
whole generated-module and original Java test-suite coverage.

## Theorem/assumption definition constructors

ThmOrAssumpDefNode has a provisional constructor for goals created before their
body, a full constructor and construct completion. Defaults retain theorem=true,
arity zero, a nonnil empty formal array, nil body/module/source/instantiation/
proof/backlink/labels, and false local/Suffices flags. Completion updates theorem,
body and original module, registers the same node, then installs supplied formals.
A nil array preserves existing formals/arity on repeated completion. Full
construction sets source and instantiatedFrom before registration; parameter
installation still follows it. Duplicate detection therefore observes initial
arity zero, even when the later supplied array has parameters.

Source/instantiation/module/body pointers and arrays retain identity. getSource
returns the supplied immediate source or the node itself. Same-origin checks
require the same concrete class and immediate source; that source's module must
have no constant/variable declarations. Label tables retain the same operations
and aliasing as other label owners. Matching tests the application's actual
operator arity, ignoring the receiver's arity and operands, as Java does.

Assumption/theorem statement generation uses the constructor class and matching
dispatch. Structured proofs, complete module vectors and inherited level data/
visitors still require implementation. Bounded constructor comparisons alone do
not establish complete named AP or theorem graph parity.

## Unnamed ASSUME/PROVE bodies and scope metadata

Allocate AssumeProveNode before assumptions and PROVE. Retain source syntax,
mutable assumption arrays and the actual prove child. inScopeOfDecl has one more
entry than assumptions: carry the previous entry forward, setting the next entry
true only for a direct NEW clause. Nested NEW declarations do not change their
parent's direct declaration scope. Nested AP nodes retain null goals and their
constructor inProof=true; the outer theorem's AP changes to false after its proof.
getAssumes aliases its array; getChildren constructs an assumptions-then-prove
array and preserves each child pointer.

Match boxed ASSUME/PROVE delimiters during generation. Ordinary AP registers the
Generator class-wide actual $$InAssume declaration in its temporary SymbolTable
context; boxed AP checks that binding and reports source E4005 when inside ordinary
assumptions. Restore nested declaration contexts after generation. The current
clause resets at outer AP entry and advances after each outer assumption only;
ordinary labels following these AP bodies retain that clause rather than zero.

Named theorem statements now provide their actual goals and top-level AP context
lifetime. Proof-step goals remain unported and retain the native fallback. Missing
children leave owners incomplete. Complete proof graph construction, Suffices,
level checking and visitors remain pending.

## NEW declaration construction and binding identity

NEW retains its wrapper syntax and its declaration's own syntax. Generate the
optional domain before buildParameter constructs and registers the OpDeclNode.
Use the declared NEW kind/level, actual arity, operator synonym and current module.
Only ordinary ConstantDeclKind populates the declaration's own level/all-parameter
sets; NEW constant/state/action/temporal declarations do not do so.

Allocate NewSymbNode after registration, retaining the actual declaration and
domain nodes. Its getChildren returns null without a domain, otherwise only the
domain, matching source rather than treating the declaration as this getter's
child. Registration retains an earlier duplicate/rejected binding; the wrapper
still owns its newly constructed declaration. The native proof scope carries the
selected actual declaration rather than rebuilding name/arity metadata. Missing
earlier imported/native identities remain an explicit registration gap.

Bare higher-arity symbols reject during generation and retain the owning
Generator's nullOAN before enclosing construction. Reporting the same error only
in a later native arity pass leaves enclosing graphs incomplete and loses source
allocation order. Operator operands and call-name resolution keep separate paths.

The bounded comparisons invoke Java's actual generateNewSymb in a prepared
module context. They establish this method's construction order and references,
not complete named AP/goal/proof-marker allocation or NEW level checking. Named
owners, proof clause lifetime, NEW inherited level data/visitors, imported identity
integration and evaluator sharing still require implementation.

## EXCEPT construction and AtNode reference ownership

Generate the base before allocating the EXCEPT application. Allocate its mutable
operand array and actual node before specs. Generate each component in order:
record selectors use actual StringNodes, single indices use their generated node,
and multi-index components construct tuples. Preserve Java's exact tuple syntax
index selection, including its unusual selection for later components.

Construct the spec's sequence and mutable pair before its RHS. Push both actual
nodes during RHS generation, then pop them and fill the same operand arrays.
The outer EXCEPT context remains active while generating a nested EXCEPT's base
and indices. Every AtNode retains the innermost active EXCEPT/pair references,
uses its pair's syntax/location, and exposes the original base and modifier by
reference. Multiple AtNodes share those target identities. Proof previous-RHS
references retain their separate generation path.

Label rejection now checks these actual stacks, as Java does, without a parallel
depth approximation. A missing child graph leaves the enclosing EXCEPT incomplete,
while source construction order and real failure sentinels are preserved. AtNode
inherited level checking, visitors and evaluator graph reuse remain pending.
The existing semantic corpus translation still uses native reference/level helpers;
constructor comparisons alone do not establish complete canonical corpus parity.

## Ordinary expression labels and LS scopes

Generator label scopes retain a nullable label table and a sequence of actual
formal arrays. Each ordinary definition, named function and label body pushes a
fresh frame. Quantifiers, CHOOSE, anonymous functions, comprehensions and LAMBDA
push their actual bound formals on the current frame. LAMBDA shares its enclosing
label table; ordinary definition parameters are resolvable symbols but do not
enter the LS formal sequence. A nested label starts a fresh sequence even while
outer formal symbols remain visible in the SymbolTable.

Generate the body and pop its table before resolving label parameters. Reuse
actual FormalParams; an illegal parameter gets the source diagnostic and a dummy
FormalParam with its argument syntax. Allocate LabelNode afterward, retaining
nonnil empty parameters, actual syntax/body identity and the popped table. Compare
actual formal identities against the parent frame, then register the label;
duplicates report source E4336 and preserve the earlier entry. Completed ordinary
and recursive OpDefs retain their body's same label table. Function conflicts
retain source behavior, including attaching the generated table to a reused OpDef.

Implemented nested-NEW and EXCEPT guards return the owning Generator's actual
nullLabelNode before generating the body. A missing canonical body remains an
explicit graph gap; no substitute completes its enclosing definition. The bounded
ordinary-label comparisons cover initial nil goal/zero clause fields. AP/proof
contexts continue using the native path: their actual goal nodes, clause lifetime,
NEW/marker ownership and labeled AP bodies require further implementation.
Label tables preserve default Java Hashtable capacity 11, threshold 8, head
insertion, descending bucket enumeration and source rehash chain reversal at
capacity growth 2*n+1. Hash names as Java UTF-16 strings, then mask the sign bit
for bucket selection. Replacing a value preserves its chain slot; rejected
addLabel duplicates change neither entry nor capacity. A nullable table remains
distinct from an allocated empty table. setLabels/getLabelsHT preserve the shared
mutable table pointer; getLabels returns a nonnil slice and a separate mutable
array for nonempty results. LabelNode accessors retain name/arity/body/goal and
getChildren's single body entry, including a null body.

Bounded comparisons now include actual unsorted enumeration, collisions, Unicode,
rehash boundaries, duplicates, table alias mutation and returned-array mutation.
Complete Hashtable APIs/concurrent enumeration, nullable UniqueString-key APIs,
LabelNode level checking/visitors, complete EXCEPT/AtNode level data and evaluator
sharing remain pending. The table here serves the Generator's sequential default
label-table operations; it does not claim every Java Hashtable API.

## Explicit recursive declarations and completion

RECURSIVE generation increments the shared unresolved counters and advances the
section number only when the prior unresolved sum is zero. Each syntax entry
allocates an unregistered constant OpDecl, unnamed dummy formals and a registered
undefined OpDef, in that order. Dummy formals preserve null names and syntax.
Every new OpDef enters recursiveDecls, recursiveOpDefNodes and
opDefsInRecursiveSection, including rejected duplicate declarations. Earlier
bindings remain in the actual SymbolTable. Missing prior imported/native graph
identities do not receive a fabricated registration target.

An operator definition replaces the declaration's formal array before generating
its body. Source setParams changes neither declared arity nor constructor-sized
level/weight/Leibniz arrays. Completion changes body, syntax and defined on that
same OpDef; LOCAL on the later definition does not change the declaration's
original local flag. Zero-arity recursive function definitions complete that
same node with the provisional function specification before generating its body.
A function declaration with operator arguments remains undefined. Wrong-level
operator definitions construct rejected new nodes, preserving the earlier binding.
Ordinary definitions inside a recursive section retain the actual section fields
and module vector membership, even when their own LET level has no unresolved
entries. Existing native completion still owns the shared unresolved counters.

Preserve Java's paramsMatch loop overwrite: after matching array lengths, only
the last formal determines its arity check. Calls use actual OpDef formal arrays
rather than stale native declaration signatures. Operands beyond that array after
a mismatched definition return the owning Generator's nullOAN without generation.
The real matcher replaces repeated native checks once a call graph exists.

A completed native body whose canonical graph is unported still updates the
actual declaration's defined/syntax state. Its body remains visibly incomplete;
no substitute body node completes the surrounding LET. This prevents incorrect
undefined and duplicate-definition diagnostics while child graphs are pending.
Full module recursive level checking, inherited LevelNode data, graph visitors,
imported/qualified identities and evaluator sharing remain unfinished.

## Named function definition graphs

Named function domains and bound formals retain their actual generated nodes.
The temporary self FormalParam uses the entire function-definition syntax.
The same actual context is popped after preparation and pushed again for an
accepted body; restoring only the native formal-name map is insufficient.

After domain/formal preparation, processFunction allocates a provisional
$NonRecursiveFcnSpec with empty operands, the self formal, grouped bounded
symbols, tuple flags and domains. It constructs/registers its zero-arity OpDef
and appends the module definition before generating the body. Nested body LET
definitions therefore follow their enclosing function in ModuleNode's list.
A rejected name that already resolves to an OpDef reuses that existing node,
as Java does, while still generating the unused new function specification.
A declaration/formal conflict generates the body outside the function context.

The active function stack checks names when a GeneralId generates an OpAppl,
including a bare function reference. It changes the first matching active
specification to $RecursiveFcnSpec. After body generation, the specification
retains the actual body; a nonrecursive specification clears its temporary
self-formal array. Ordinary later references share the registered definition.
GeneralId selector failures retain the owning Generator's actual nullOAN,
including when nested inside domain expressions.

Unimplemented bodies leave the provisional specification incomplete and do
not mark its enclosing LET graph complete. Missing imported/qualified identities,
AP labels, full LevelNode checking and evaluator graph sharing remain pending.
Explicit recursive completion follows the declaration path described above.

## Higher-order operands, LAMBDA and ASSUME/PROVE scope

Higher-order calls now pass actual generated operands to the concrete OpAppl
constructor and symbol matcher when their graphs are available. A failed
operator operand retains its owning Generator's nullOAN; an arity-mismatched
lambda retains that Generator's nullOpArg. GeneralId expected-arity rejection
precedes OpArg allocation. Missing unported graphs remain incomplete rather
than acquiring a replacement node. Unqualified prefix/infix/postfix GenID
operands resolve the raw name and synonym before allocating OpArg. Wrong arity
returns nullOpArg and reports E4004 at the enclosing call; an undefined symbol
returns nullOAN with GenID's E4004 declaration-lookup diagnostic. Error parameters
retain the actual enclosing operator object; its default Java string is its
location. Qualified GenID prefix arguments still need canonical integration.

LAMBDA generation creates actual formal nodes and its body in a temporary
context, restores that context, then constructs an unregistered user-defined
OpDef named LAMBDA. It preserves the constructor's recursion and LET defaults.
Only a matching arity allocates the OpArg, with the actual module and syntax.
A lambda in an expression slot returns the source null-expression result and
its diagnostic without generating its body. Constructors distinguish that
actual null result from a child graph whose implementation is still missing.

Selected calls use each receiving formal's operator arity to generate operands.
Their final qualified/compound selector graphs remain pending. ASSUME/PROVE
uses the current expression generator and a restored lexical scope for NEW
symbol metadata, including operator arity. Domains are generated before the
new symbol becomes visible. Named and unnamed bodies dispatch through the
actual ASSUME/PROVE structure; the legacy quantifier-shaped view must not
allocate false scalar FormalParams or quantifier applications for NEW.
Canonical NEW/AssumeProve and INSTANCE/fact graphs, missing function bodies,
inherited LevelNode fields and iteration, visitors and shared evaluator
construction remain pending. The later shared literal level integration closes
the original basicOpDefTest; it does not complete those other features.

## Ordinary definitions and LET ownership

For ordinary nonrecursive local/top-level definitions with generated bodies, OpDef
construction occurs after the temporary parameter context is popped. The actual
SymbolTable registers the new node and keeps the earlier binding if registration
is rejected. Native symbol metadata, identifier references and applications
retain that canonical definition identity. Definition formals keep their exact
LHS syntax, including the entire declaration of operator-valued parameters.
The operator's own name is resolved while those parameters remain visible.

Top-level generation retains the constructed OpDef on the actual Definition and
appends it to ModuleNode.definitions. Native module symbol metadata publishes
that node only if the actual SymbolTable binds it; rejected definitions retain
the earlier declaration/definition binding. Actual constructor registration
diagnostics replace the native finishDefinition surrogate when the graph exists.
The constructor helper is shared with local definition generation. Retained LET
definitions do not thereby become bindings in the root module context.

LET owns a real semantic context through its IN body. LetIn retains that context,
the supplied arrays and generated body before the context is popped. getLets
lazily filters user-defined definitions once and returns its cached mutable
array. Completed ordinary LET definitions also append to ModuleNode's definition
list in source generation order, including nested-body definitions before their
enclosing definition.

Missing function/label bodies and INSTANCE nodes leave LET graphs incomplete.
Recursive declarations whose earlier imported bindings lack canonical identity
also remain incomplete; no native placeholder completes them.
Imported/theorem/ASSUME-PROVE graphs, labels, full inherited LevelNode
fields and canonical level checking, visitors and evaluator sharing are pending.

## CASE, records and syntax-time token identities

CASE retains one actual $Pair per arm and a final $Case. Each pair is allocated
before generating the next arm, and OTHER has a null condition. Record forms
allocate each field StringNode, generate its value, then allocate its pair.
The final record or record-set application is appended to ModuleNode's record
vector. The getter returns a fresh array retaining application identities and
generation order. Record selection creates its field StringNode after generating
the record expression. All nodes retain their actual arm/field syntax.

Record duplicate checks compare all earlier labels. Java Errors deduplicates
reports with equal code, location, format and parameters. The fixed-format,
parameterless field diagnostic preserves that behavior by its actual location;
full shared error-log ownership is still pending.

Token syntax constructors intern their raw images while parsing, as Java does,
so record-field StringNodes reuse identities established by syntax construction.
Late bridge-time interning alone changes UniqueString token order when semantic
nodes are generated earlier. Retain the raw token identity even when a parser
production normalizes its display image. Other syntax constructor identities
and the bridge's remaining interning reconstruction still need integration.
Unported child graphs remain incomplete; ordinary OpDef, LET/INSTANCE, full
LevelNode fields and evaluator sharing remain pending.

## Operator applications and flattened syntax forms

Unqualified ordinary calls now resolve an available concrete symbol before
operands and reject incorrect supplied arity at the argument-list location.
A callee name does not allocate its own application. Actual generated operands
feed the general matched constructor at the call's syntax node. Symbols and
children lacking canonical graphs remain incomplete. Compound selectors and operands lacking canonical symbols still require
complete graph integration. Ordinary OpDefs with generated bodies now use the
shared constructor and actual registration described above.

Prefix/infix/postfix generation resolves the raw GenID and source synonym
before generating operands. Prefix '-' becomes '-.' as in GenID.finalAppend.
The AST's normalized spelling is insufficient for declaration lookup. Actual
resolved symbols and generated children feed the general matched constructor.
Unknown canonical symbols/children remain incomplete rather than synthesized.

Junction lists and Cartesian products construct one source builtin application.
Their AST wrappers lacking syntax are flattened; explicitly nested source nodes
keep their own applications. Single-item junctions still use $ConjList/$DisjList,
and N_Times uses $CartesianProd even for two operands. Avoid allocating unused
intermediate applications: semantic UID order is part of source graph identity.
Bounded source comparisons verify shapes, syntax kinds, operand order and
relative allocation. Full call/selector checks and identities, ordinary OpDef
integration, other bodies, LevelNode data and evaluator sharing remain pending.

## Concrete symbol matching and primary OpArg construction

The general OpAppl constructor initializes its graph, calls the actual formal,
declaration or OpDef matcher, and ignores the returned boolean as Java does.
Recoverable failures keep their application and diagnostics; thrown internal
errors retain diagnostic details and yield no completed application. OpDecl
checks argument count/nullness. OpDef handles module kinds, variadic/fixed
arguments, builtin expression operands and scalar/higher-order formals. An
operator operand is invalid if its own OpDef expects an operator argument.
Negative formal arity logs an internal error without changing the result or
throwing. Keep these source distinctions rather than imposing uniform rejection.
Shared Errors ownership and frontend exception conversion remain pending.

Primary OpArg nodes retain the actual resolved symbol, name/arity, syntax and
canonical semantic ModuleNode. Available zero-arity references and boolean
literals now retain matched application graphs; operator references retain
OpArgs. Full call and selector wiring remain pending,
alongside imported identities, LevelNode data and evaluator sharing. Direct
source comparisons verify matcher/constructor branches; they do not establish
complete ordinary operator generation or original-method completion.

## Bound application construction and formal references

OpAppl constructors 4/5 retain their supplied unbounded formals or bounded
formal groups, tuple flags and ranges. Bounded ranges preserve caller nullness;
unbounded constructors create a non-null empty ranges array. Zero-arity formal
references construct actual OpAppls pointing at accepted FormalParamNodes.
FormalParamNode.match compares operator arity with its own and emits no errors;
constructor 2 ignores its boolean result, matching source behavior.

Quantifier, CHOOSE, anonymous bounded function and comprehension graphs reuse
actual generated domains/body/formals. Parsed formals sharing a domain form one
group; distinct equal-text domain syntax remains separate. Empty tuple binders
preserve source's unusual formal counts/names. Unbounded CHOOSE has no tuple
flags, even for a tuple binder. Quantifiers pop formal scope before application
construction; CHOOSE constructs before the pop. Rejected declarations remain in
arrays, while body references keep the earlier accepted symbol. These graphs
remain incomplete for unported children. General application/selector wiring, named functions,
lambdas, other expression forms, LevelNode data, LET/INSTANCE and evaluator
sharing remain pending. Bounded evidence verifies identities/shape/allocation;
it does not credit the ordinary OpDef test or establish whole graph parity.

## Retained builtin application graphs

Semantic generation retains actual graphs on expression generation sources.
Literal graph references point to the already-generated TLC literal nodes.
Tuple/set enumeration, IF, function-set/application, action and fairness
branches construct actual SANY OpApplNodes after generating their children.
The builtin constructor preserves operands and syntax, resolves its operator
from the global context, leaves bound/tuple arrays null and creates a non-null
empty ranges array. It skips matching exactly as Java constructor 3 does.
Function application uses one argument directly; all other argument counts
construct $Tuple at the same application syntax before $FcnApply. Fresh
generation replaces retained graphs. No placeholder represents an unported
child: its containing graph remains incomplete. Remaining symbol matching and bound-body cases involving unported children,
LET/INSTANCE graphs, LevelNode data and evaluator sharing remain pending.
Bounded source comparisons establish the implemented tree shapes, syntax kinds,
operand order and relative allocation order, not complete expression parity.

## Ordinary OpDef construction and actual graph links

Graph edges use the shared Kind/GetUID semantic identity methods so they can
retain actual TLC literal nodes alongside root SANY nodes, without wrappers or
second allocations. The ordinary OpDef constructor preserves the supplied
parameter array, body, module, table, syntax and source. Null parameters become
a non-null empty array; argument maxima start at MaxLevel, weights at zero and
Leibniz flags true. Registration occurs after field initialization. Recursion
fields begin with source defaults, including -1 LET/section values for all
OpDef constructor variants. Generator's ordinary-body path is not wired to this
constructor yet: generate all actual body graphs first, then construct/register
after popping formal scope. Recursive definitions complete an earlier node.
Constructor evidence does not establish ordinary graph or LevelNode parity.

## Temporary formal SymbolTable contexts

Expression generation pushes a real semantic context alongside its native
formal map. Accepted retained FormalParamNodes register in that context;
rejected declarations preserve the earlier binding. Nested scope exit restores
both representations. Ordinary definition bodies use this same scope boundary,
matching Generator.processOperator's temporary formal context. Full module
generation uses the module's SymbolTable; standalone generation starts with the
global builtin context. Existing native conflict checks remain necessary while
ordinary operator and theorem registration is incomplete. This establishes
accepted formal registration and scope restoration, not complete constructor
registration timing or canonical ordinary OpDef/body graphs.

## Shared string construction

The syntax parser decodes TLA string escapes while retaining its outer quotes.
Semantic generation strips those quotes once, interns the resulting value in
StringNode, retains syntax/location, and records source level-check iterations.
The TLC bridge reuses the generated node, whose GetRep exposes the interned
UniqueString. Fresh generation replaces any previously retained literal nodes.
XML reads the same value and treats every stringXML argument as decoded data;
its callers include literal values and record/selector names. Calling Go Unquote
again corrupted values containing their own quotes or backslash escape text.
The original source value must survive unchanged. Bounded actual Java comparisons
cover UTF-16 contents, controls, supplementary Unicode, iteration and XML data.
This does not complete common LevelNode guards, visitors or ordinary OpDef graphs.

## Decimal and integer literal metadata

SANY creates and retains DecimalNode alongside NumeralNode during expression
generation. Fresh generation replaces prior literal nodes. Decimal construction
parses the concatenated integral/fractional parts as a signed long and keeps the
negative fractional length as its exponent. It preserves zeros instead of
normalizing. On overflow, the literal's metadata retains an arbitrary-precision
unscaled value and signed 32-bit scale; numeric fields stay zero, matching Java.
Decimal image parts retain their exact source spelling. Level checking records
the supplied iteration and always succeeds. The TLC bridge reuses generated
nodes, and constant processing still rejects real-number expressions.

XML reads numeric representation fields. For a small decimal it emits its
mantissa and negative exponent; for an overflow decimal source DecimalNode
emits its unscaled value and positive scale. Preserve this source branch rather
than imposing a different sign convention. Integral/fractional XML fields retain
the source image. Integer XML uses NumeralNode's integer or big-integer value,
including decimal leading zeros and TLA radix prefixes. Go's base-zero parser
is not a compatible substitute. Existing ungenerated AST callers use the same
literal constructors. This verifies literal metadata, not complete graph export
or common LevelNode fields/guards.

## Builtin OpDef initialization

The initial context stores actual `sanySemOpDefNode` objects for builtins and
Generator's null operator node. The former separate builtin symbol class is
removed. Builtin construction allocates syntax and phony formals first, marks
the operator defined, and initializes level data only when both property arrays
exist. Missing metadata remains nil; zero-argument `mk()` arrays remain empty
and non-nil. Variadic operators keep null formal arrays and store the first
maximum/weight entry, while Leibniz flags follow the complete weight array.
Leibniz uses positive weights; initialized builtins record level iteration 99.
Builtin classification compares current initial-context node identities, and
parameter classification recognizes declaration/formal node classes. The whole
original initialization test checks all 72 property rows before/after reInit.
Ordinary OpDef body graphs, Generator registration timing and recursive completion
remain pending; builtin or null-node success does not establish their parity.

## Shared numeral construction

SANY expression generation constructs the existing TLC `NumeralNode` for integer
literals and retains it on the source expression. The node holds the original
syntax/image, source location, signed 32-bit representation or big-integer
fallback. The TLC bridge reuses that same node; synthetic expressions that have
not undergone SANY generation still use its existing constructor path. TLC's
later constant processing retains responsibility for rejecting out-of-range
values. The original incremental radix test now inspects generated nodes rather
than independently reparsing a native-AST string. This establishes numeral
construction and sharing. Numeral level checking records its supplied signed
32-bit iteration without a cache guard and always succeeds; its no-iteration
overload requests `LevelChecked + 1`, including Java's signed wraparound. The
normal composite checker invokes this on retained numerals at iteration 1. The
original basic expression test checks syntax identity and ConstantLevel after
actual level checking. Common LevelNode getter guards, other metadata and full
recursive graph iteration propagation remain pending. These methods establish
neither full LevelNode nor other semantic graph parity.

## Expression namespaces and INSTANCE substitutions

A source module name is not an expression namespace. EXTENDS imports its names
without a module prefix; only an actual named INSTANCE registers qualified
exports. Keep private evaluator lookup indexes separate from names accepted by
semantic generation. A named instance may have the same name as its module.

SubstInNode.constructSubst resolves defaults against the current SymbolTable,
not a completed-module scan. Earlier LOCAL definitions and instance parameters
are eligible; later declarations are not. Scalar defaults construct zero-argument
applications before explicit WITH substitutions are generated. OpDeclNode and
OpDefNode have different wrong-arity messages; FormalParamNode's match does not
report that application error. Replacing a scalar default with WITH does not
undo its construction diagnostics. After explicit substitutions, check remaining
operator defaults, then report missing declarations in Context enumeration order.

Duplicate substitutions report the second RHS location after RHS generation.
Illegal targets report the target location. For a wrong-arity GeneralId operator
argument, selectorToNode reports 4271; generateOpArg returns nullOpArg and the
subsequent substitution check reports 4243 at its null location. Preserve both
errors, exact messages and ranges. These paths match 30 complete Java comparisons;
complete canonical failure nodes and shared evaluator graphs remain pending.

Module and proof INSTANCE units call the same generateInstanceSubstitutions
implementation. It uses accepted target declaration metadata, current scoped
bindings, expression generation for scalar RHS values and generateOperatorOperand
for operator RHS values. Invalid expression RHS syntax is rejected before its
children are generated; unknown operator names retain their original error before
the nullOpArg arity error. Deferred substitution level checks run separately.

Operator-argument generation carries its mode into qualified prefix validation.
An unapplied parameterized INSTANCE prefix contributes parameters to the operator
argument; it is not a scalar call with missing arguments. Clear the mode before
generating child expressions. Nested ASSUME/PROVE label rejection retains the
whole label syntax range. Unsupported ASSUME/PROVE selector paths retain code
4005 instead of the former placeholder 4340; labeled access also retains its
source label-name parameter. Complete original Test210/Test212
diagnostics match Java in addition to 50 namespace/RHS cases.

LET INSTANCE units now share substitution generation and export registration
with proof scopes. They register only actual exported names and signatures;
an instance prefix does not authorize every possible qualified name. Keep the
first accepted binding and restore the outer symbol and recursive contexts when
leaving LET. A named export's location is its instantiation syntax; its retained
original source syntax and module identify the shared definition independently.
Check kind and arity before suppressing duplicate registration for a shared
parameter-free origin. Different arities still report both export and instance
name conflicts. Missing labels retain Generator's 4004 code, exact selector
range, message and label-name parameter.

All 28 LET cases match complete Java diagnostics, including RHS failures,
missing/default substitutions, higher-order instance parameters, duplicate
exports, local-only definitions and nested scope boundaries. This establishes
generation and registration behavior. INSTANCE generation also retains the
resolved substitution array: defaults follow Context enumeration; WITH replaces
a matching slot or appends an explicit substitution. Retain target declaration
identity and default expression bindings instead of rebuilding from a completed
module. LetInNode checks definitions, its body, then instances, including unused
instances. LetInNode imports only InstanceNode's ArgLevelParams into its own
level data, in addition to body/OpDef constraints and relationships. It does not
copy whole instance scalar or argument constraint sets. Collection now has
explicit result sets and a shared resolved-instance path; LET imports only the
resulting ordered co-parameter relationships. Generation now retains accepted
LET INSTANCE exported definition references in registration order. Collect their
constraints alongside the ordinary definitions. Evaluate imported bodies with
symbolic declaration identities before translating LC/ALC/ALP through the actual
resolved substitutions. Subst.getSubLCSet applies nonconstant-module declaration
maxima before parameter-set translation; getSubALCSet keeps substituted parameter
operators and getSubALPSet translates their relationships. A substituted OpDef
consumes a relationship via its calculated argument maximum rather than exposing
an inlined body relationship. Retained InstanceNode collection shares these
transformations and separately merges each RHS's own constraints. All 22 expanded
LET cases match, including the two previously missing bounds and an alias case.
Do not compensate by merging an entire target module's fields. Full canonical
OpDef/SubstIn/LetIn graph construction remains pending. Ordinary expression/
signature LET summaries now merge constraint fields from ordinary and retained
imported definitions and only co-parameter fields from instances. Clone the body
summary before adding those sets; level/levelParams/allParams remain body-only,
while nonLeibnizParams remains empty. Unused local definitions can constrain an
enclosing formal even when they do not affect the LET body's level or weight.
Own local formals do not become enclosing argument maxima.

OpDefNode.opLevelCond comes from symbolic formal-operator relationships before
actual specialization. OpApplNode uses each actual operator's per-argument Leibniz
result to add the affected operand's allParams to its nonLeibnizParams. Preserve
this rule independently of a containing LET's own empty nonLeibnizParams. Local
signature keys include sorted captured formal identities and actual operator
identities; do not share another enclosing specialization's summary. Nested Subst
argument-maximum queries use an independent analyzer rather than clearing the
caller's active signature state. Twenty-two additional LET signature comparisons
match complete Java diagnostics, including higher-order non-Leibniz results.
Complete canonical fields and iteration remain pending.
Check substitution expressions before constraints, preserving diagnostic
order and the instancer's lexical LET context. User-defined substituted operators
use the symbolic argument maximum calculation rather than builtin-only maxima.

AssumeNode writes its own temporal constraints but getLevelConstraints returns
the expression's constraints. ASSUME C alone therefore does not export an
ActionLevel bound, while ASSUME C = {} exports equality's argument maximum.
ASSUME/PROVE retains its additional temporal bounds. Module constraint collection
combines actual expression bounds and these temporal bounds by minimum.
All 24 module/LET comparisons match complete source diagnostics. Complete
InstanceNode/LetInNode graphs remain pending. Argument level requirements now
travel with symbolic expression/definition summaries: ParamAndPosition requirements
combine by maximum, while ordinary level constraints combine by minimum. Declared
operators contribute each operand's intrinsic level and ArgLevelParam relationships.
Definition application replaces a formal scalar dependency with the actual
argument's level/dependencies, retaining separate operator identities for
higher-order specialization. INSTANCE compares each required minimum with the
actual operator's calculated maximum, using resolved substitution-array order and
ascending argument positions. It reports the substituted operator name, one-based
position and required level as Java's 4246 diagnostic does.

The builtin-only argument syntax scan is removed. All 24 additional paired cases
match Java, including indirect definitions, higher-order forwarding and theorem
expressions. The co-parameter diagnostic path now uses propagated ArgLevelParam relationships;
its direct-expression scanner is removed. Preserve relationship insertion when
combining summaries and translating scalar formal dependencies. The final module
set uses ArgLevelParam.hashCode: operator declaration hash + zero-based position
+ parameter declaration hash. The shared Java HashMap retains buckets, resizing,
collision trees and same-class object-identity tie-breaks. Runtime identities differ
across Java and Go, so collision-tree order is not a portable byte-for-byte promise.
Repeated relationships retain one membership. Check both substituted components'
level correctness, then compare the substituted parameter's level with the
substituted operator's maximum. Error 4247 keeps Java's zero-based position,
target operator name, whole INSTANCE range and original message punctuation.

ModuleNode initializes constant declaration constraints to ConstantLevel when
isConstant is false before merging expression bounds by minimum. Leaving an
expression's weaker ActionLevel bound as a separate module bound fabricates
another diagnostic and loses the source's constraint-set semantics. All 26
co-parameter cases and 48 preceding level/argument cases match complete source
diagnostics. A separate 12-relationship collision-tree probe matches membership,
counts, codes, ranges and messages; identity-dependent iteration orders differ.
When collecting ModuleNode constraints, distinguish its own Context from
EXTENDS imports. Own LOCAL operator definitions contribute to the module;
unexported LOCAL definitions in an extendee do not independently contribute to
the importing Context. Dependencies from an exported definition into its owner's
LOCAL body still propagate. Preserve top-level INSTANCE constraints, including
LOCAL INSTANCE. Initialize nonconstant ConstantLevel declaration bounds for each
retained instantiated module and translate their parameter sets through its
substitution context; initializing only the outermost module loses stronger
nested bounds. EXTENDS imports nodes rather than the extendee's whole module
bound set. Sort translated constant names to retain deterministic collection.
Thirty expanded source comparisons match, including 29 parseable modules and
one grammar rejection. InstanceNode also merges each substitution expression's
own level constraints, argument constraints and co-parameter relationships,
independently of whether the substituted declaration occurs in target bodies.
Collect those expressions in the instancer's lexical context from the retained
resolved substitution array, including implicit defaults and WITH replacements.
Do not reconstruct defaults from a completed-module name scan. Twenty-seven
additional parseable RHS comparisons match complete source diagnostics, including
unused LET definitions, composite expressions and scalar/operator defaults.
Full formal-operator metadata, imported graph identities,
original frontend phase APIs and canonical evaluator sharing remain separate
pending work.

## Separate semantic-generation and level-check phases

`ParseSanySpecSource` resolves the native syntax/dependency graph without
semantic generation. `GenerateSanySpec` builds the retained module checking
plans without running level checks, matching SANYFrontend.processSemantics's
Context reset and Generator phase. `CheckSanySpecLevels` checks the generated
root and returns its level result separately from the diagnostics. The legacy
runner still generates and checks each external module in its original reporting
sequence, then runs linting separately.

The programmatic root shares imported definition/fact plans and copied top-level
nodes. LOCAL definitions do not become imported checks. Repeated root checks
start a module iteration while child definitions and top-level nodes retain
their distinct caching rules. AssumeNode returns its expression's check result
independently of its own constant-level error; InstanceNode also retains its
levelCorrect result independently of a non-Leibniz diagnostic. Keep those results
separate from aggregate Errors.isSuccess. Full canonical node fields and iteration
algorithms are still pending; retained checking plans do not establish graph parity.

Expression checks also carry child return values separately from Errors.
LetInNode combines each definition/body/INSTANCE result, and parent expressions
propagate the child's result rather than interpreting its diagnostics as a false
result. InstanceNode's non-Leibniz report leaves its own levelCorrect true unless
another constraint fails; a containing LET or OpDef therefore retains that true
result. Substitution expression validity and action/fairness argument gates use
these child results. A non-Leibniz report must not suppress a separate enclosing
argument-max error. Definition, assumption and theorem plans retain the actual
expression result alongside the accumulated diagnostics.

Unary OpApplNode operands check before their own builtin argument maxima.
Propagate the operand's levelCorrect return separately from diagnostics and gate
an enclosing maximum error on that result. Prime allows ConstantLevel and
VariableLevel operands; its intrinsic ActionLevel and zero argument weight are
already retained in the builtin metadata. Constants, literals and booleans may
therefore be primed. Do not add a separate constant-prime ban or duplicate
syntax-based double-prime check. Ordinary maxima produce Java's exact 4205
error. Twelve bounded prime comparisons match complete legacy diagnostics.
ASSUME-PROVE checks its assumptions before PROVE. NEW domains check before
the declared symbol enters the lexical context; nested blocks copy that context.
NewSymbNode's aggregate level is the maximum of declaration and domain levels,
while references to its OpDeclNode retain the declared level. Temporal domain
errors use the whole NEW syntax range and source 4356 message. Indirect domain
levels come from the shared dependency analyzer rather than a declaration-kind
syntax scan. Combine assumption validity for the AP return value; call PROVE's
check and preserve its diagnostics without combining its boolean, as Java does.
Twenty-nine parseable AP phase cases match exact diagnostics, phase results and
repeated checks; two additional grammar rejections match the legacy frontend.
Complete ASSUME-PROVE child graphs and LevelNode metadata remain pending.

All 27 phase comparisons match generation errors, level errors, codes, ranges,
messages, return booleans and repeated checks. The original TestInstanceNode
method preserves every phase/count/code/parameter assertion; unchanged Java and
Go pass. These retained plans do not establish complete canonical LevelNode
fields/iteration, full frontend-helper API or full-workspace completion.

## TLC bridge declaration metadata

For generated source modules, use the retained owned declaration object when
constructing runtime context entries. A repeated declaration keeps its first
location and signature. A declaration rejected against an enclosing or EXTENDS
binding contributes no new local runtime context entry, constant registration
or INSTANCE substitution target. This avoids phantom inner declarations and
later writes that replace the original declaration's location. Native AST APIs
without generated source graphs continue through their existing metadata path.

All 34 observations match lower Java FastTool across five vectors: complete
variable/constant arrays, counts, names, arities, ranges, and initial-state counts
and values. This verifies declaration metadata and rejection handling, not complete
canonical SANY/evaluator node sharing or general semantic allocation order.

## Source symbol registration and concrete context classes

SymbolTable registration first accepts the identical node, then a vacant name.
For an existing different node, reject builtin source locations before checking
kind and arity. Formal and bound symbols always require a fresh name. Only after
those checks may a shared operator source suppress a warning; its actual defining
module must have no constants or variables. Other same-kind/arity duplicates
return true with warning 4801 and keep the earlier binding. Module conflicts
return false with error 4223. Retain complete messages, locations and parameters.
All 149 Java observations across 77 registration cases match.

Context merge classification uses the concrete node class. In particular,
OpDefNode(UniqueString) has kind zero and arity -2 yet remains a definition.
The faithful original context test now constructs that node and an OpDeclNode
with the shared nullSTN. Its unchanged assertions caught the old kind-based
classification, which is corrected in production. Both original Java JUnit and
Go pass. Module generation now retains an external SymbolTable and copies the
enclosing stack before pushing each nested context. Its module pointer tracks
the current semantic owner. Actual repeated declarations use that table's
registration primitive. Other symbol classes and source phases still use native
scope adapters until their complete graphs are constructed.

## Retained ordinary declaration nodes

Generate OpDeclNode for every constant and variable occurrence before attempting
registration, including rejected duplicates. Retain the constructed nodes in
source order on the native module. Constants use kind 2 and level 0; variables
use kind 3 and level 1. Arity and whole-item syntax follow Generator.buildParameter
and processVariables. Each node owns its UID, semantic module link and levelChecked
value 1. Constant levelParams and allParams contain the node itself; variable
sets are empty. Local accepted nodes enter the retained module context, and
identifier expressions retain their actual declaration identity. The class-wide
ASSUME marker uses the same constructor with null module and syntax.

Thirty-eight original constructor observations still match. Each declaration also
retains its original SymbolTable. Arity is read per syntax occurrence, including
rejected duplicates with a different signature. Warnings retain the first binding,
location and arity. Nine complete diagnostic comparisons match Java.

Merge direct EXTENDS contexts in source order without cloning declaration nodes.
Carry the retained declaration through imported and enclosing bindings into
identifier expressions. Sixty-one Java observations match transitive/diamond
identity, original module/table ownership, table contexts and enclosing references.
Nested scopes resolve enclosing declarations without placing them in their own
context. Remaining operator/theorem entries, qualified declaration selectors,
INSTANCE wrappers, LevelNode constraints and visitors require further porting.
The native diagnostic phase remains authoritative until all exported graphs exist.


## Retained ModuleNode and Generator ASSUME marker

Generate each ModuleNode before its body. The native semantic node retains source
arity -2, whole-module syntax/location, context and zero-based nesting. External
modules start with a duplicate of the retained builtin context; internal modules
start with their own empty context and resolve enclosing names through the symbol
stack. Append an internal module to its parent's semantic definition list before
generating the child body. Formal nodes retain the actual constructed module node
alongside their existing native module adapter. Fourteen actual Java module
observations match for these constructor and ownership properties.

Generator also initializes a class-wide OpDeclNode named $$InAssume before its
four per-instance sentinels. Retain that node once, with kind/arity/level zero,
levelChecked one, no module and null syntax/location. Twenty-eight sentinel/marker
observations match, including sharing across Generator instances and construction
before nullODN. Source uses the marker for ordinary ASSUME scope; native marker
scope installation/resolution remains pending.

All five expanded aggregate label cases now match complete Java diagnostics. The
70-formal range and first hash are exactly 305..374 and 1607 in both runtimes.
This proves the inspected construction prefix, not general allocation order:
complete declaration registration and ordinary OpDefNode/OpApplNode construction,
all context entries and the
complete ordered module-definition graph remain incomplete. Canonical failure
results, LevelNode data, visitors and evaluator sharing still require porting.

## Generator-owned sentinel nodes

Each external-module Generator constructs nullODN, nullOAN, nullOpArg and
nullLabelNode in that order. Go now retains concrete semantic nodes for these
constructors, using the shared UID allocator/indexed slots and the non-null
nullSTN builtin syntax/location. nullODN has kind zero, arity -2 and nil body/params.
nullOAN has OpApplKind, points to nullODN, and retains empty operands/ranges with
nil bound arrays. nullOpArg has OpArgKind, name nullOpArg, arity -2 and no operator
or module. nullLabelNode has LabelKind, name nullLabelNode, arity zero, an empty
formal array and nullOAN as its body.

Native modules retain their Generator-owned set. External analyses receive fresh
sets; nested module generation and expression generation use the enclosing set.
Twenty-three actual Java constructor observations match, including syntax,
locations, shapes, links and consecutive relative UID construction. Existing
parser/SANY checks and four unchanged original TLC models pass. The module/static-marker construction described above now also matches the
inspected first-formal allocation prefix.

This is constructor and ownership scaffolding, not complete ordinary semantic node
classes or canonical failure graphs. The existing generationFailure adapter still
controls error propagation; returning the actual nullOAN/nullOpArg/nullLabelNode
from all source failure paths remains pending. ModuleNode and other graph
constructors, LevelNode data, visitors and evaluator sharing remain incomplete.

## Builtin operator formal construction

OpDefNode(BuiltInOperator) allocates its own node before its phony formals.
The Go builtin constructor now retains an array for every fixed arity, including
an empty array at arity zero; variadic operators retain nil. Entries are distinct
FormalParamNodes named Formal_0, Formal_1, etc., each zero-arity, local, with no
module and no SymbolTable installation. Each builtin retains its separate kind-zero
syntax carrying the builtin image, filename and zero coordinates.

A null FormalParamNode syntax remains null and its location is unknown. It does
not inherit SemanticNode.nullSN's non-null builtin syntax/location. Correct this
in the shared constructor. All 668 actual Java constructor observations match
across 72 builtin operators and 77 phony formals, including arrays, node metadata,
syntax/location and consecutive relative construction order.

The native frontend now mirrors Context's lazy class initialization and reInit:
first use constructs the global table; every full frontend entry replaces it with
a fresh table before parsing, including failed parses. Each Spec retains its own
analysis table; lower semantic checking uses the current global table if no
frontend snapshot exists. The initialization/reset mutex protects the global
pointer, while native generators read the retained spec table. Named-function
builtin resolution retains the actual context node rather than arity-only metadata.
Eleven source lifecycle observations match, including reset construction spans and
old-analysis stability. This does not establish general selector/evaluator sharing,
complete LevelNode data or absolute allocation order. Retain real constructed
nodes; do not simulate allocations with offsets.

## Retained label parameter arrays

After generating a label body, resolve each label argument in the current symbol
context. Retain the actual FormalParamNode when resolution supplies one. For
non-formal symbols, construct a fresh zero-arity dummy FormalParamNode for every
occurrence using the complete argument syntax node and location, with current
module ownership and no binding installation. A parameterless label retains an
empty array. Same-formal repetitions share identity; repeated non-formal names
have distinct dummy identities.

Twenty-two direct Java SANY observations match for definition formals, quantified
formals, repeated references, empty arrays and non-formal dummies, including their
relative construction order and syntax/location. This is retained native metadata,
not a complete LabelNode graph. The diagnostic traversal now emits an illegal
parameter message for every non-formal occurrence at its complete argument range,
then checks repetition by retained UID identity. All entries have the same concrete
FormalParamNode class and kind, so UID membership implements their equals test.
Distinct dummy nodes do not trigger repetition despite equal names.

Twenty-two exact argument/repetition diagnostic observations across twelve scratch
cases match Java. Required parameter traversal now carries an ordered sequence of
formal arrays for the current label frame. Quantifiers push one flattened source
group after all domains; CHOOSE, functions and set comprehensions carry their
retained arrays. Labels and LET definitions reset required groups. Matching removes
each UID in sequence order, so rejected same-named nodes remain distinct missing
requirements. Native expressions without resolved metadata retain their old scope
representation. Filtered sets traverse their source predicate once.

Fourteen required-parameter observations across sixteen scratch cases match Java;
all sixteen cases now match complete diagnostics. Extra validation uses a Java
HashSet adapter keyed by UID for one concrete semantic class/kind. Its hash uses
the exact SemanticNode formula, `31*(31+kind)+UID`, with int32 overflow. Hashes are
injective within this domain, so distinct keys cannot need an identity tie-break.
Required nodes are removed in LS sequence order; iteration of remaining buckets
produces the source's single aggregate error.

The adapter uses the existing OpenJDK HashMap port. Removal specializes source
removeNode's `matchValue=false, movable=true` path, including removeTreeNode,
balanceDeletion, conversion back to lists and moving the tree root to the front.
3,470 direct Java/Go operation and iteration observations match, with tree bins
confirmed before removal across collision/high-bit patterns. Nine earlier complete
label diagnostic cases match. Native/unresolved metadata retains its old adapter.

This does not establish general graph allocation order. After retained module
and static-marker construction, all five expanded aggregate cases match. The
70-formal first UID/hash are 305/1607 in both runtimes. Complete remaining graph
constructors rather than manufacturing UID offsets for other allocation gaps. Complete unified LS/generator integration and stop allocation beneath
forbidden label bodies. No full LabelNode graph parity claim.

## TAKE/PICK formal construction and proof scope

The proof generator retains an ordered formal-node array on each TAKE/PICK step.
Generate distinct bound domains in the enclosing context before constructing any
formals. Allocate every source parameter, even when binding is rejected. Accepted
bindings retain their actual nodes; rejected declarations preserve the earlier
binding and its diagnostic location.

TAKE introduces accepted bindings immediately for its proof and later steps.
PICK temporarily introduces them for its predicate, hides them from its own
proof, then installs the same accepted nodes when that proof finishes. The pending
scope carries resolved symbols rather than reconstructing names and positions.
Twenty-three observations against actual Java FastTool proof nodes and seventeen
complete SANY diagnostic cases match. Existing focused parser/semantic checks,
the whole SANY package and package compilation pass.

Actual TAKE/PICK generation now uses bounded/unbounded OpApplNode constructors.
Generate all domains before any formals; retain one group per syntactic bound,
including adjacent tuple groups. TAKE uses the current proof context directly.
PICK pushes its temporary context before domain generation, captures it before
the predicate, and supplies the complete formal array to label generation. Pop
that context before constructing the PICK application. After the step proof and
theorem owner, insert its actual symbols into the enclosing context using Java
Hashtable enumeration; preserve rejected bindings and registration diagnostics.
Each proof frame restores its inherited formal map when its context closes.

All 2,424 rows across 35 whole modules agree with Java: complete graph structure,
syntax/locations, relative allocation order, formal/reference identities, label
parameters, tuple groups, diagnostics and context order. Nested proofs, duplicate
names, simultaneous domains and contexts crossing resize thresholds are covered.
The earlier 2,281 AP rows remain unchanged. These observations do not complete
LevelNode inheritance, visitors, evaluator sharing or allocation order across
all graph features; no original-method completion credit is added.

## Named-function formal preparation and body context

processFunction generates domains and constructs bound formals under a fresh
context before resolving the function name. It always allocates a temporary
zero-arity self formal using the complete function-definition syntax, binding it
only if the name was previously unresolved. It then pops the context while
constructing/resolving the OpDefNode, and pushes the same context for an accepted
body. Go now retains and restores that preparation context, including parent
formals for LET definitions. RECURSIVE declarations retain their original operator
binding; body references do not resolve to the unbound temporary self formal.

Resolving the name after constructing bounds detects a function name equal to
its bound parameter. A rejected non-operator name generates its body outside that
prepared context. Initial-context builtin names resolve as existing operators,
producing source's function-name diagnostic rather than addSymbol's builtin
redefinition message. Builtin resolution here is a metadata adapter; complete
builtin SymbolNode identity and tool-slot wiring are still pending.

Twenty-five valid-node observations and fifteen complete diagnostics match Java.
The initial source probe incorrectly assumed every final function application
retains its self array; makeNonRecursive removes that array. The corrected probe
compares retained self nodes on recursive functions only. Go's private preparation
record retains construction information and is not a complete final OpApplNode
representation. Final nonrecursive array shape, LevelNode data, evaluator sharing
and absolute source allocation order remain separate graph-port requirements.

## Function and set expression formal construction

processFcnConst, processSetOfAll and processSubsetOf generate all domains in the
enclosing scope before allocating any formal nodes. generateLambda uses the same
fresh parameter context without domains. The native expression generator now
retains these formals and actual body references, preserving prior bindings when
new declarations conflict. Shared native domains are generated once per source
bound group; flat multi-name metadata does not regenerate their domain for each
name. Nested domains resolve enclosing formals, and sibling scopes have distinct
identities. Domain errors precede formal declaration errors, matching Java.

Filtered sets have a predicate operand in the source graph. Their native derived
Element is not separately generated. Thirty-three node observations cover function
constructors, tuple bounds, set-of-all, filtered scalar/tuple sets and LAMBDA;
eighteen complete diagnostics cover scope, conflicts and error order. This does
not establish full OpApplNode bound groups, LevelNode data, evaluator sharing or
absolute allocation order. Named-function preparation is covered separately above.

## CHOOSE formal construction

processChoose generates its domain in the enclosing scope before allocating
formals, then generates the predicate in their context. The Go generator now
retains those formal nodes and resolved predicate references. Shared formal-scope
and binding helpers preserve old bindings on rejected declarations and restore
the outer context afterward. Equal-named sibling CHOOSE binders remain distinct.

Preserve the source syntax distinction: bounded scalar and all tuple formals use
the identifier nodes; an unbounded scalar formal receives the CHOOSE keyword node
(children[0]). Its location and duplicate-binding diagnostics therefore point at
the keyword. The existing empty-tuple translation retains the closing token as
one formal, matching the source constructor loop. Twenty-five node observations
and thirteen complete diagnostics match Java; the earlier nineteen quantifier
observations also pass after binding helper reuse.

This does not establish full OpApplNode shape, LevelNode data, evaluator sharing,
label parameter arrays or absolute allocation order. Remaining proof/declaration
graph construction must preserve actual node identities too.

## Quantified formal construction and scope

Bounded quantifiers generate all domains in the enclosing symbol scope before
allocating any FormalParamNode. Bounded, unbounded and temporal quantifiers create
all formals in declaration order under a fresh symbol context. Newly allocated
nodes remain in the quantifier's array even when binding conflicts occur; body
references resolve to the previous binding after such a conflict. Scope restoration
preserves enclosing formals and keeps sibling quantifier identities distinct.
Builtin redefinitions and duplicate bindings use the source messages and locations,
including conflicts with earlier formals on the same source line.

Go retains one formal on each native quantifier wrapper and a flattened source
parameter array on the root wrapper. Nineteen source node/reference observations
and eleven complete diagnostic cases match. This is not complete OpApplNode bound
grouping, LevelNode data, evaluator-node sharing or absolute source UID allocation
order. Proof/native locals without retained nodes still use their existing binding
representation; no artificial formal identity is assigned to those bindings.
Label checks still need actual resolved formal arrays and source LS stacks, and
generation below rejected labels must eventually stop in the unified generator.

## Retained definition formal identity

Native expression generation now creates a concrete sanyFormalParamNode for each
ordinary definition parameter and proof-INSTANCE parameter, before deciding
whether SymbolTable can accept the binding. Conflicts retain the previous binding.
Definition bodies retain the full ordered parameter array, including conflicting
new nodes, and identifier expressions retain the formal resolved in their actual
context. Repeated references share a node; same-named formals from different
definitions have distinct identities.

The node extends the existing SANY semantic symbol base with kind FormalParamKind,
localness true, native module ownership and retained declaration syntax/location.
It uses the shared SemanticNode UID allocator and per-node indexed slots.
Its equality checks concrete class, kind and UID, matching SemanticNode.equals.
Sixteen source observations establish these constructor/reference properties and
relative UID/hash behavior. Absolute source allocation order is not established.
LevelNode data, visitors, actual label formal arrays and evaluator sharing still
require porting. TAKE/PICK construction is covered separately above. The native
Module ownership link is not a claim of a complete Java ModuleNode graph.

## Label generation guards and body order

Generator.generateLabel first rejects labels outside definitions/proof steps,
then labels under nested ASSUME/PROVE NEW declarations, then labels within an
EXCEPT clause. Each guard returns nullLabelNode immediately. Go's label traversal
now uses that order and stops rather than recursing into forbidden bodies or
checking their parameters. Allowed labels generate their body under a new label
scope before resolving/checking their own parameters. LET definitions reset the
required bound-parameter stack while retaining ambient EXCEPT and NEW restrictions;
this also applies to ASSUME/PROVE definitions inside LET.

Repetition and missing-parameter ErrorDetails retain the source format strings,
parameters and locations. Fifteen complete source diagnostic observations match.
The earlier nine-case audit still differs in five cases; full FormalParamNode
identity, dummy-node creation, UID-dependent HashSet extra-parameter ordering and
selector handling remain pending. The name-based parameter checker is still a
port gap; these traversal corrections do not establish full graph parity.

## WorkerValue demultiplexing

WorkerValue.demux evaluates through OpDefEvaluator's state-expression overload.
Java Tool delegates that overload with Empty successor state and EvalControl.Clear;
Go constant preprocessing now uses the same control, replacing EvalConst.
Demux deeply normalizes the first value before checking its mutation category
and current global worker count. With mutable values and multiple workers,
allocate using the current count, retain the first value, snapshot the seed,
then reset that seed and reevaluate/deep-normalize each remaining copy. Remove
the extra mutable switch and caller-captured worker-count argument. Source nil
values fail during normalization; do not return a successful null value.

WorkerValue's constructor retains its supplied array, matching source private
constructor ownership. Fifteen exact direct-source observations cover all six
immutable overrides, mutable tuple/set copies, equal random contents, independent
objects, normalization, count changes during first evaluation, one worker,
initial/copied nulls, evaluation failure and array aliasing. The source evaluator
proxy verifies Empty context/state and the exact cost model; the Clear control
comes from inspecting Tool's actual overload. Null exception-message text is
not established by the probe. Eleven unchanged original models pass.

Ordinary lookup now selects the current worker ID, or zero outside an IdThread
scope, matching WorkerValue.mux. State worker metadata does not select a constant
copy. Eleven observations match actual Java FastTool lookup and WorkerValue.mux,
including invalid indices and the unmuxed context/body paths. The Go comparison
uses its legacy definition cache; indexed SymbolNode storage and the complete
lookup-provider graph remain pending.

## Random-enumerable instance initialization and restoration

Java's DefaultRandom and TLCStateRandom retain their behavior on each Random
object. The Go JavaRandom now carries an enumerable-kind marker and its initialized
predecessor. The factory captures the BFS/default choice only when a thread-local
instance is absent. Existing default instances remain default when the checker
changes; existing BFS instances keep predecessor reseeding after that change.
The initialization hook consults the shared IdThread current state and reseeds
only for a different non-null predecessor. Saved/restored instances therefore
retain both their stream position and the predecessor marker.

Thread-local assignment has a separate flag: remove permits fresh factory
initialization, while set(null) makes the next get throw NullPointerException.
A plain java.util.Random lacks the source EnumerableValueRandom interface;
the corresponding ordinary NewJavaRandom instance raises ClassCastException
when retrieved through RandomEnumerableGenerator. Setter first calls get and
returns that initialized old instance, then installs the replacement without
clearing its predecessor. It does not fabricate an RNG for a null argument.

Seventeen temporary actual-source observations match exact fingerprints and
integer draws, captured mode, instance identity, saved/restored stream continuity,
setter initialization and exception types. Source allocates an actual ModelChecker
class marker without running its constructor solely for the factory's getClass
predicate; no model-checking or JVM behavior is claimed from that marker.
Thirteen earlier thread-local observations still match under the short race
probe. Nine unchanged original model methods pass at their original bounds.
Exception-message text and thread-lifetime cleanup remain separate work.

## Random-enumerable seed and reset locality

RandomEnumerableValues.setSeed updates the shared seed, then calls reset on
only the current thread. The Go port now removes only the caller's generator;
other goroutines retain their RNG objects and continue their streams. Reset
first obtains/initializes the prior generator, as source get does, and returns
that object before removing it. Preserve the current predecessor scope: Java's
IdThread state is a separate thread-local variable, so deleting the RNG cannot
delete that state. Clear the thread's RNG assignment while retaining the saved
instance's own initialization state. Eliminate the duplicate predecessor field in the RNG registry; random
enumeration reads the same CurrentState slot used by checker error handling.
PushRandomEnumerableState delegates to the common scope instead of maintaining
a second copy. AbstractChecker's existing error reset now clears the state
before trace regeneration, as source does.

Thirteen exact Java/Go observations match caller replacement, peer identity and
continued draws, seed changes, returned object identity and state-scope
preservation and explicit shared-state clearing. The isolated two-goroutine
scratch probe passes -race in 1.028 seconds. Six unchanged original models pass,
including RandomElement's full eleven-state trace. Existing original random-value
tests remain unchanged. RNG implementation choice and restored-instance state
are now ported below; thread-lifetime cleanup remains pending. This is not
complete ThreadLocal parity.
WorkerValue's mutation categories and demux decisions are ported above;
remaining lookup and lifecycle gaps still require reconciliation.

## TLCCache constant-expression map

TLCCache stores its HashMap on the actual expression node's indexed tool slot.
Remove the global UID-keyed store and per-cache RWMutex. The override uses one
class-wide reentrant read/write lock. It reads the slot and value under the
read lock, releases that lock, takes the write lock, re-fetches a missing map,
then repeats the lookup before evaluating, initializing and storing the value.
Its finally boundary releases either lock on lookup, cast and evaluation
failures. Nested calls can reacquire the lock while computing another entry.

Reuse the existing OpenJDK HashMap port, extending it with optional key equality.
Native key identity remains the fast path and existing users retain their
behavior. TLC keys use ValueJavaHashCode and the lookup key's Equal method only
when hashes match. This avoids comparisons between unrelated mixed-type keys.
Lists, resizing and tree bins share that equality path. Non-Comparable Value
keys use source class-name ordering and native pointer identity for ties,
following the established Go identity convention used by state-value maps.
Map iteration is not exposed by this override; source identity hashes differ
between processes. The state-level cache branch is unchanged by this port.

Temporary observations compare actual Java and Go overrides: mixed-type keys,
equal fresh objects, node slots, equal-UID isolation, nested caching and class-cast
cleanup. Thirty-two source-generated StringValue keys have one exact shared
hash and force actual tree bins in both implementations; all fresh-equal lookups
return the same cached objects. No permanent test was invented or original
model weakened. Broader value/cache fidelity remains an independent requirement.

## TLCEval expression-node cache

TLCEval's constant path reads and writes the actual expression-node tool slot.
Remove the separate process-wide UID-keyed cache and its literal-node exclusions.
Preprocessed values and WorkerValue instances are visible to the override;
separate nodes with colliding UIDs remain independent. Read under the existing
reentrant lock, then release/read again under its write lock before evaluating,
converting and storing the resulting Value. Mux uses the current worker index,
or zero outside a worker, as Java does. Invalid worker indices raise the source
array bounds exception; they must not silently select worker zero.

Preserve source lock boundaries: value-cast failure releases the read lock,
while WorkerValue.mux failure occurs before its finally block. Invalid worker
observations must come after normal writes in a scratch process, since source
read-lock leakage would otherwise prevent those writes. The source and Go
probes both demonstrated that behavior. Thirteen exact observations cover
pre-existing values, UID collisions, conversion/cache reuse, worker selection,
bounds and cast identity. Cast-message text is not established by that probe.
TLCCache's constant-path storage, map and reentrant lock are ported below.
Broader worker/cache behavior and native unindexed cache APIs remain pending;
WorkerValue demux decisions are ported above.
No permanent test was invented; the existing original TLCEval and TLCCache
model methods remain unchanged and pass.

## Numeric proof-step conversion

The parser's proof-step helper uses signed 32-bit parsing, as Java's
Integer.parseInt does. Overflow raises the shared NumberFormatException with
`For input string: "..."`; it must not return the ordinary unrecognized-token
result. Leading zeros remain valid, even when their count exceeds an integer's
usual decimal width. Implicit `*` and `+` levels keep their source sentinel
values. Existing parser/front-end exception boundaries retain propagation.
Thirteen temporary exact helper observations match Java; this does not establish
whole proof-generation parity or add original-test inventory credit.

## Semantic node constructor foundation

SANY symbols embed `tlc.SemanticNodeBase` and allocate identity through
`tlc.NewSemanticNodeBase`. Module, ordinary-symbol, built-in, evaluator and
wrapper factories therefore share one process-wide AtomicInteger equivalent,
as Java's SemanticNode superclass does. The former SANY-only counter introduced
in `1470b93` was incorrect and is removed. Kind/UID/hash access delegate to the
common base; remaining graph constructors and allocation order are unported.

The base retains signed 32-bit UID/hash overflow. Its constructor assigns UID
eagerly and marks assignment independently of UID bits. `uidPlusOne == 0` is a
valid representation of Java UID -1 after wraparound; it cannot mean unassigned.
The existing Go zero-value base path uses separate assignment state and a lock,
so initialization does not change an already assigned wrapped UID. Normal source
constructors do not use that zero-value initialization path. Do not reset the
counter during front-end initialization.

Indexed tool slots now reside in SemanticNodeBase. Both SANY and evaluator
nodes use that same storage. Replace the former process-wide map keyed by UID
or fallback hash: source stores objects on each node, independently of hash or
UID collisions, and unreferenced graphs can release their cached objects.
GetToolObjectAt/SetToolObjectAt retain sparse growth, preserved earlier slots,
null writes that still grow the array, unchanged length after clearing and
negative-index exception type/message. Go synchronizes each node's array for
concurrent cache access. Constructor and zero-value initialization publish the
storage together with assignment state. The helper routes indexed calls to
actual semantic nodes; non-node inputs are rejected instead of hashed into a
shared namespace. Tool-ID storage and indexed APIs now use int32, including
lazy values and cache keys. GetSemanticToolID ports FrontEnd's zero-based
process-wide counter, including signed wraparound. Spec's static ID is allocated
once through synchronized lazy initialization; every Tool and default spec
processor reuses it. Explicit native Tool IDs remain available to callers.
At MAX_VALUE, the source toolId+1 wraps before array allocation; raise
NegativeArraySizeException without changing the existing slots. Thirteen exact
source observations match shared IDs, explicit allocator calls, wraparound and
array overflow; the sixteen previous slot observations still match.
Existing unindexed literal/cache APIs remain pending; this is not whole
superclass parity.

Sixteen exact source observations compare tool slots on distinct nodes with
colliding UIDs, separate tool numbers, sparse holes, clearing, growth on null
writes and bounds before/after growth. Nineteen earlier SANY observations still
match through the shared API. These are temporary manual probes, not invented
persistent tests or translation credit.

The current expression-generation formals still need actual FormalParamNode
identity. Java equality also checks the concrete runtime class, kind and UID.
Do not substitute symbol names or coarse wrapper classes for that hierarchy.
Keep label-parameter HashSet ordering pending until graph identities and actual
allocations are faithful; a common counter alone does not resolve it.

Twenty-six bounded source observations match: 19 retained kind/hash/tool-slot
observations, one alternating SANY/evaluator constructor sequence and six
stable-getter/hash observations spanning MAX_VALUE/MIN_VALUE and -2/-1/0.
These are temporary manual evidence, not permanent tests or translation credit.

## NEW declarations and source JavaCC scanners

`NewSymb` retains the source NEW symbol declaration frame and expectations.
Its constant, variable and state/action/temporal alternatives use the actual
JavaCC two-token previews. Constant declarations select IdentDecl with its
separate preview. An operator with arguments followed by `\in` throws the
source custom ParseException after consuming that token and before appending
it or entering its expression. A valid constant domain retains an ordinary
IN token leaf; the prior T_IN substitution belonged to a different production.

Theorem uses source `jj_2_22(3)`, including its semantic ASSUME/BOXASSUME gate
and bounded scan of AssumeProve. Invalid facts within that budget can fail in
Theorem before an Assume-Prove frame is entered. A first-token predicate alone
cannot preserve that failure boundary.

`sany_generate_lookahead.go` mechanically translates the 159-method source
scanner closure for entry points 16–20 and 22 into
`sany_lookahead_generated.go`. Regenerate with `go run sany_generate_lookahead.go`;
the generator accepts optional parser-source and output paths. Alternative
backtracking restores scan position while retaining the furthest position and
remaining budget. A matching token at the budget boundary raises the source
LookaheadSuccess control flow. Semantic predicates read from scan position,
including function-head recognition, field-token checks and junction indentation.
The operator-stack predicate uses the active production stack, restored after
normal expression completion.

`sany_lookahead.go` saves both successful and failed calls with their source
expiration generation. Error construction rescans live calls in entry-point
order, catches LookaheadSuccess around each linked call list and accumulates
actual error sequences with JavaCC's 100-token bound. This supplies following
input length from actual rescans, including the extra token in the invalid
nested numeric label observation. Other handwritten previews still need their
source entry points and rescan integration; this is not full JavaCC parity.

All 599 complete parser TRACE/results and 137 selected raw module trees match
pinned Java. A temporary matrix compares all 295-by-295 token pairs for the five
NEW previews, and both ASSUME variants over all 295-by-295 following token pairs
for theorem lookahead: 609,175 matching verdicts in 261,075 rows, with the default
junction/operator-stack context. These are bounded scratch observations, not
permanent tests, canonical corpus AST equality or test-port inventory credit.

## Parser output routing and original formatting assertions

`sany_output.go` ports the parser's level ordinals and Simple/Silent/OutErr stream
routing. Messages without arguments remain verbatim; supplied string arguments
use TLC's source-derived Java formatter, including its exception types. Ordinary
filtered messages are still formatted before being discarded; silent output
bypasses formatting. Platform line separators and constructor null-stream errors
follow Java. ParseSanySyntaxWithOutput reports actual syntax diagnostics without
file resolution or semantic analysis, matching the original formatting test's
in-memory parser use. The three unchanged original formatting assertions now
exercise this production implementation. Numeric/general-Object driver format
arguments, PrintStream's queryable error state and remaining production-frame
coverage remain reconciliation work, so this does not establish full output parity.

The syntax parser now owns its output and reports errors after catching only the
source parse/token-manager exceptions. Its bpa equivalent logs TRACE before
looking up the first token; epa pops the message frame before logging TRACE.
Deferred production exits propagate exceptions without popping frames or logging
spurious endings. Expression and fairness retain their explicit early epa before
reduction/structure checks. Restored frames include variable declarations,
identifier LHS/declarations, theorem/proof, LET definitions/expressions, quantifier
bounds and applicable infix operations. Preserve source spellings, including
Identifier Declation and LetIn's Case Other Arm label. Eighteen bounded actual
Java/Go TRACE and parse-result observations match; this is not a complete grammar
frame audit.

Constant and recursive declarations use Java's separate ConstantDeclarationItems
production rather than formal IdentDecl. Its two-token lookahead consumes an
argument list only when '(' is followed by '_'; failed lookahead retains the
source following-input length for later parse errors. ParamSubDecl retains its
own production frame and N_ConsDecl node. Source operator declaration wrappers
are typed leaves whose image/range/comments come directly from their token;
they have no fabricated token child. Formal operator declarations own their
separate Op. Symbol Declaration frame and source error wording. Definition
prefix/infix/postfix LHS frames are also retained. Thirty-one bounded complete
TRACE/parse-result comparisons and selected declaration/LHS heir-kind/image
comparisons match pinned Java. These observations do not establish general
JavaCC rescan, source-range or whole-module syntax-tree parity.

INSTANCE and Substitution retain their source frames and expectation boundaries.
Source jj_2_12(3) accepts a following substitution only after comma, target and
'<-'; failed lookahead leaves the comma untouched and retains scanned following
input length for later error reporting. Source alternatives admit the contiguous
177-token postfix/prefix/infix/Unicode/identifier interval; using general operator
metadata would incorrectly admit '.' as a target. Operator targets are typed
source leaves with token images, without fabricated children. Body's earlier
two-token INSTANCE lookahead precedes field-name reclassification: bare INSTANCE
requires an identifier as its second token, while LOCAL exhausts that budget at
INSTANCE. Instantiation performs field-name reclassification later, as Java does.
Fifty-four bounded complete TRACE/parse-result observations and selected target/
arrow subtree kinds/images match pinned Java. General JavaCC rescan, remaining
productions and full syntax-tree parity remain unproven.

LAMBDA's parameter list follows Java Identifier (, Identifier)*, rather than
formal operator declarations. Its syntax heirs contain direct identifier tokens,
and semantic bound extraction consumes those actual heirs. Source Lambda frame,
expectations and mandatory-token failure boundaries are retained. Operator-arity
mismatch E4274 uses the operator-name location, independently of the subsequent
OpDef.match E4271 on the complete application. Sixty-four bounded complete parser
observations and 12 LAMBDA semantic observations match Java. These observations
cover nested bindings, substitution, arity and unknown identifiers without claiming
complete semantic graph or application-range parity.

CHOOSE, MaybeBound and IdentifierTuple preserve source production frames,
expectations and mandatory-token parse failures. Empty identifier tuples remain
syntactically permitted. Source processChoose derives formal count from all tuple
heirs (including delimiters); for <<>>, it therefore creates one formal named >>
and retains the tuple flag. The Go semantic bridge preserves this arithmetic
rather than losing tuple identity through an empty/nil slice. Three direct
FastTool/Go-tool probes match actual bound metadata and results: the empty-tuple
form over a singleton 1-tuple returns <<1>>, an empty-tuple domain fails the tuple
shape check (matching error prefix), and ordinary two-formal CHOOSE returns
<<1, 2>>. These probes use semantic definitions rather than eager-evaluation cache
values, and do not invoke Java TLC reporting. Seventy-seven bounded complete parser
observations and 13 CHOOSE semantic observations match Java. General quantifier
construction and whole semantic-graph parity remain separate reconciliation work.

SomeQuant follows Java's Identifier (, Identifier)* colon lookahead, otherwise
consuming mandatory QuantBound groups. QuantBound accepts one tuple or a list of
identifiers before IN; it does not speculate across commas to repair malformed
syntax. SomeTQuant accepts unbounded identifier lists and retains Java's original
Bound Quantified Expression frame name. The bridge maps temporal universal and
existential forms to OP_tf and OP_te, independently of OP_uf/OP_ue.

Source processQuantBoundArgs generates all domains before adding any formal.
Native quantifier wrappers sharing a source node therefore form one scope for
expression generation, call arity and operator argument checks. Distinct nested
source nodes retain distinct scopes. Domain diagnostics precede formal conflicts;
previous-symbol conflict ranges come from the actual semantic symbol. Source tuple
heir arithmetic also applies to quantified and subset bounds, including the one
closing-token formal for an empty tuple. ContextEnumerator's tuple mismatch uses
Assert.fail's TLC_ARGUMENT_MISMATCH runtime carrier and source message, rather than
a generic native error. Ninety-five complete parser observations, 22 quantified
semantic diagnostics and ten lower-level quantified metadata/runtime probes match
Java. The runtime scope includes bound group/name/tuple metadata, five successful
bounded evaluations, one invalid tuple error prefix, and four unbounded/temporal
metadata checks. It does not establish temporal evaluation or whole-graph parity.

String syntax follows TLAplusParser.String and reduceString: a typed token leaf
retains quotes while decoding escapes. Generator's StringNode(treeNode, true)
strips only the surrounding marks during semantic translation. TLC preinterning
uses the same decoder. Native literal values retain embedded quote characters as
data. The original EWD998 debugger checks observe the quoted syntax image, while
TLC evaluates the unquoted string value. One old native parser assertion expected
an unquoted syntax image; it now matches the observed Java image.

Number preserves Java's two-token decimal lookahead and failed-scan lengths, plus
its number/decimal flags. ParenExpr uses mandatory delimiter exceptions. TupleOrAction
uses Expression's actual one-token FIRST set and junction indentation predicate,
then requires RAB or ARAB and ReducedExpression for an action subscript. Reflection
of source jj_2_49(1) accepts 195 token kinds in the default indentation context.
Junctions/JuncItem's actual production frames are retained before bullet consumption
and after expression generation. JunctionListContext stores list type and alignment;
new bullets match type and column without an additional line predicate. Invalid
start and empty termination preserve IllegalArgumentException/NoSuchElementException.
Junctions terminates its context only on normal completion. checkIndentation visits
descendants, skips nested conjunction/disjunction list subtrees, and throws on the
first descendant not strictly right of the current alignment. JuncItem calls epa
before this check, preserving the source residual stack and exact message (including
its doubled space before disjunction). General JavaCC lookahead remains separate. Generic expression operator nodes use
typed token leaves. SyntaxTreeNode.updateLocation's aggregation recomputes from
both child arrays and preserves Java int extrema for an empty node. All 150 bounded
complete parser-output observations and 31 selected primitive/junction expression trees
(kinds, images and ranges) match Java. These checks do not establish whole-module
canonical AST or complete SyntaxTreeNode parity.

FieldVal and FieldSet retain their actual frames and mandatory MAPTO/COLON tokens.
ExceptSpec requires BANG followed by at least one ExceptComponent, retains the
source = or , expectation after each component, and constructs a T_EQUAL leaf
before the replacement expression. ExceptComponent requires either DOT Identifier
(with keyword field-name reclassification) or LSB Expression (, Expression)* RSB;
it does not accept an empty index. A field named @ pushes the nonthrowing source
parse error @ used in !.@, preserving subsequent parsing and ordered messages.
Shared Identifier now uses the actual throwing token-consumption path, eliminating
nil dereferences in malformed components. BraceCases has its actual Some { } form
frame and mandatory delimiters. The later source decision-flow port of brace and
square-bracket forms is described below. All 176 complete bounded parser observations
and 43 selected expression trees (including retained trees after nonthrowing !.@
errors) match Java. Existing original model assertions and fixtures are unchanged.

IfThenElse, Case, CaseArm and OtherArm retain source frames IF THEN ELSE,
CASE Expression, Case Arm and Case Other Arm, respectively, with mandatory token
exceptions. Case consumes ordinary arms while the next aligned CASESEP is not
followed by OTHER, then optionally consumes its final OtherArm. The actual junction
context governs separator indentation. LetIn retains Java's misleading Case Other
Arm frame, but its LET/IN tokens are mandatory. LetDefinitions requires one or more
LOCAL/DEFBREAK/RECURSIVE units and stops at the actual source token boundary instead
of skipping malformed tokens until IN.

Source OperatorStack.reducePostfix constructs N_FcnAppl with the function in zero
and the bracket node's heirs in one. The port no longer nests a fabricated N_FcnAppl
wrapper, and semantic translation no longer unwraps a lone application argument.
That old compensation would discard a real nested argument under the corrected
source tree shape. All 208 bounded complete parser observations and 56 selected
expression trees match Java. Four lower-level FastTool/Go-tool probes match actual
application opcode/metadata and values for chained, nested, multiple and tuple-valued
applications. No Java TLC reporting entry point is used. Full constructor/parser/
semantic-graph parity and whole-module canonical AST assertions remain unproven.

BraceCases and SBracketCases now follow their source decisions rather than
searching ahead for a separator. matchFcnConst previews an identifier list or
balanced LAB/RAB followed by IN, leaving tuple validation to IdentifierTuple.
Braces separately preview an identifier/identifier tuple followed immediately by
COMMA or COLON; other forms parse one expression before choosing its continuation.
Preserve membership reconstruction through N_GeneralId/N_GenInfixOp, the typed
T_IN leaf and the source complex-membership comprehension error. Square-bracket
forms require their source bounds, expressions and delimiters. Function and
set-comprehension bridge groups share one conversion of each syntactic domain,
before formals enter context, and preserve separate adjacent tuple domains.

SyntaxTreeNode constructors use the full source SyntaxNodeImage array, distinct
from node-kind constant names. All 446 source entries match both native constructor
images. Selector.finish's unknown syntax retains a zero-valued slot and the node's
own image, so generation reports its constructor diagnostic before FindingOpName
reports a missing name. Reconstructed tuple membership reports E4003 and E4005
instead of inventing an operator named <<. Current bounded observations match:
308 complete parser traces/results, 93 selected expression trees, 16 brace semantic
cases and nine actual brace opcode/group/value cases. These are scratch evidence,
not original test-inventory credit; whole-module canonical AST, general JavaCC
rescan and complete selector/semantic-node fidelity remain unproven.

## Original heap fingerprint stress target

The original long-test LSB and MSB random methods are available under the
`tlc_fp_stress` build tag. Java's `test-dist-long` target excludes these concrete
heap classes because of runtime. This explicit Go target preserves their full
2,147,483,648 iterations, default concrete factories and configuration, Java RNG
seed, per-insertion assertions, checkpoint commit and invariant checks. Run with
`-timeout=0` and no race instrumentation. Original abstract setup and progress
logic are retained; verbose progress also includes iteration counts. Inventory
translation status and full-run verification receipts are tracked separately in
TODO_TEST_PORT.md and HANDOFF.md. This target adds original tests only and changes
no production fingerprint behavior.

## CLI loading and fatal configuration I/O

Ordinary and packaged CLI models construct their tool through the same deferred
loader inside `TLC.process`, after intern recovery and the mode banner. Loading
parses configuration before SANY. `HandleParameters` prints coded command-line
errors through MP, and the TLC parser validates TLC and SANY message codes.

`ProcessExit` carries a source `System.exit` outcome through the native library
boundary. The CLI returns its mapped status without killing a library caller's
host process. This path omits source finally actions: OUTPUT cleanup, FINISHED
and trace generation. Configuration I/O code `5001`, like the other CFG lexer
codes, maps to source fallback `255`; only the explicit TLC configuration cases
in EC.ExitStatus map to `151`.

SANY tool markers bracket front-end processing, and STARTING follows completion
before checked errors are raised. Suppressed SANY_START retains buffered output
on checked failures. Parsing Exceptions become checked failure even with an empty
error list; unexpected semantic Exceptions chain through FrontEndException. Only
that checked exception releases buffered SANY output and becomes the cause of
TLC_PARSING_FAILED2, without SANY_END or STARTING. Java Error propagates unchanged.
SANY record linting follows successful semantic checking, with nonstandard
external modules in semantic insertion order. The linter retains the generating
lexical contexts for values and the module context for field-name lookup. Any
same-domain record depending on declarations or parameters suppresses its
warnings, including records in extended modules. Definitions, INSTANCE
substitution wrappers and proof scopes retain source traversal semantics;
operator arguments remain leaves. Lint output contains only the newly generated
warnings, ordinary warnings before elevated warnings, with exact source ranges
and messages. Earlier semantic errors or elevated warnings prevent linting.
Per-module semantic reporting uses accumulated diagnostics. Full ordinary
parser diagnostic rendering remains audit work.

Missing-module resolution and filename/module-name mismatch raise the internal
`sanyParseAbort`, retaining source E4220/E4221 and structured parameters. The
frontEndParse exception boundary reports the abort details and accumulated
Errors, stopping further loading. The native LoadSanySpec API returns those
diagnostics and a previously parsed root. Missing dependencies retain their
actual importing module, including nested modules. Source null and module-only
locations have distinct rendering. Resolver failures use stderr; the root-only
missing-file notice follows FileUtil's monolith-fallback branch.

The direct SANY Context and production EXTENDS conflict path share
`sanyExtendConflict`. It preserves W4800/E4224, exact messages, incoming ranges
and four structured parameters, including the prior location. Context merging
returns false for incompatible node classes and retains the first binding.
Class comparisons distinguish operator definitions, named facts, formal
parameters, declarations and modules. Reused definitions require identical
source-node identity and a parameter-free originating module; equal module
names alone do not establish identity. The original Java TestContext method
exercises this operation directly from the root package. Further semantic-node
construction and traversal parity remains audit work.

DelayedPrintStream releases a copied byte snapshot, flushes, then resets the
buffer. Its byte-buffer operations synchronize separately; destination callbacks
can write reentrantly. The original PrintStream swallows I/O errors; release
catches runtime Exception, prints its stack to System.err and retains the buffer.
Error propagates. Successful reset discards bytes appended during release.
ToolIO's string capture does not intercept source inherited raw-byte writes.

## Parser lexical failure boundary

The parser requests tokens lazily from its token manager. Token production
stops on a lexical failure instead of continuing with invalid tokens; the native
LexAll API catches that failure and returns diagnostics. The parser catches the
same actual failure at its requested-token boundary, preserving the source
TokenMgrError message separately from native diagnostic text.

TLAplusParser.parse reports the lexical message; ParseUnit then raises the
single E4003 failure in SpecObj.parseErrors, with module-only location and module
and logical filename parameters. The existing frontEndParse abort boundary
reports those details and accumulated Errors. Lexical escaping and character
columns follow Java UTF-16 code units, including invalid supplementary escapes
and EOF prefixes. Successful root parsing does not tokenize trailing input.
Definition lookahead now ports TLAplusParser.belchDEF over the same lazy token
stream. It inserts the actual DEFBREAK marker, retains the source token-kind
operator ranges and stopping conditions, and runs at the source production
boundaries. ASSUME and ASSUMPTION remain distinct in this lookahead. Definition
recognition requires the marker; DefStep leaves it for the definition parser.
Go uses buffered token indices for the source reverse traversal, preserving
forward Next links and the parser's consumed-token position.

Module parsing now records the actual bpa/epa message frames and Module's
expecting states. Failed EndModule consumption throws a typed internal parse
exception; its source message renders the actual prior/current tokens and last
five active frames. The native syntax API catches it and preserves source text
in diagnostic metadata. The file loader reports that message before ParseUnit's
E4003 abort. This mechanism currently covers the footer failure, not all grammar
productions or JavaCC's full lookahead-derived expected-token alternatives.
EOF token locations follow SimpleCharStream's last-read character, including
zero positions for empty input, CRLF's separate code units and tab expansion.

runSanyFrontEndParse contains the shared actual loading/parsing phase and its
checked failure boundary; runSanyFrontEnd calls it before semantic work. The
original ParseErrorTests.testAll uses this phase directly and checks recorded
parser output.

Expression now follows the source prefix sequence and OpenExpression versus
ExtendableExpr split. ExtendableExpr parses an actual operand, its postfix,
record and function-application extensions, then the optional recursive infix
continuation. It leaves unrelated tokens unconsumed for the enclosing grammar.
A bracket-close failure throws the typed parser exception within the real
Definition, Expression, ExtendableExpr and SBracketCases message frames.
Definition uses one frame even with LOCAL. Operator arguments and substitutions
use OpOrExpr; its reference lookahead retains source token alternatives.
Lambdas are argument forms, and fairness remains a ParenthesesExpression form.
Primitive identifiers use source GeneralId wrappers; the semantic bridge keeps
unqualified built-in Boolean literals through them. General JavaCC expected
alternatives, remaining production states and label failure continuations still
require reconciliation. Operator-stack failures now retain source messages and
locations, dedicated record reduction and raw prefix syntax; final reduction
reports accumulated errors before the parser exception. Matching bounded failure
observations does not establish full parser parity.

## Semantic operator lookup and packaged sources

Parser operator metadata supplies fixity and precedence, not declarations.
Expression generation resolves an operator in the semantic context before
visiting its operands. Unresolved operators retain the source GenID message and
following infix/prefix/postfix resolution message at their actual syntax ranges.
Both retain source error code 4004 (`SUSPECTED_UNREACHABLE_CHECK`).
GenID assembles raw identifiers and operator spelling; only lookup resolves
aliases. Unary minus changes its final raw `-` to `-.` during generation.
Initial-context arity uses the original built-in table: `\times` has arity two,
independently of XML's variadic Cartesian-product representation.

Native packaged loading embeds existing original module bytes. Java standard
modules take precedence over CommunityModules, then TLAPS modules and the single
Apalache module source. The previous abbreviated bodies and synthesized arithmetic
exports are removed; imports preserve actual declarations and their identities.
Runtime overrides and remaining semantic constructs still need separate parity
proof. No XML or ApalacheIR corpus sweep is introduced by this source loading.

Expression generation uses per-unit contexts built from actual module-body
syntax order. Declarations and INSTANCE exports enter when their units are
visited; ordinary operators and named facts enter after their bodies. `RECURSIVE`
introduces a name before its body exists. Function domains are checked before the
temporary recursion symbol; that symbol and the bounds enter for the body.
LET uses its actual body-node order for both expression checks and selector
preparation. Completed contexts remain available for subsequent level checking.

Selector preparation cannot inspect later definitions. Recursive declaration
visibility is separate from body completion: selecting an unfinished body reports
source UNSUPPORTED_LANGUAGE_FEATURE (4005). Selector failures return the semantic
null-operator placeholder rather than triggering another unresolved-name error.
Generation also distinguishes that placeholder from a null symbolic expression.
The latter stops function-application generation before arguments; an unresolved
named function still generates its arguments. Unresolved operator applications
stop before argument generation. Arity and operator-argument checks honor those
failure boundaries rather than using later definitions' metadata.

Module generation dispatches actual module-body heirs, completing each unit
before moving to the next. Named theorem definitions, statements and proofs
share their original syntax unit. Selector preparation runs at that unit rather
than as a whole-module prepass. Recursive-section checks occur before the unit's
body; unfinished recursive operators are reported in declaration order. Invalid
operand selection mirrors Generator.reportSelectorError, including source error
code 4005, selector description and individual selector-node range.

Nested modules are generated within their original module units and share the
same module-level recursive count and sum. As in Java, checking undefined
recursive operators subtracts the count from the sum without clearing the count.
Invalid nested recursive input can therefore raise WrongInvocationException
when a later definition decrements the sum below zero. Preserve that exception,
rather than converting it to a completed semantic graph. Spec.SemanticDiags
retains the semantic Errors accumulated before interrupted generation. Parent
unit errors precede child errors on both ordinary and interrupted paths.

The original semantic-error corpus helper catches WrongInvocationException for
upstream issue 1149 and returns the retained semantic Errors. Its Go translation
now retains all four original assertion families, including parameter counts and
absence of suspected-unreachable checks. SanyErrorCode carries all 66 original
metadata entries. Diagnostic arguments are retained separately from native
summary wording. INSTANCE constraint checks carry the actual instancee name and
higher-order checks carry the enclosing operator and argument position.

Label arity failures use the original code 4337, label-node location and message.
Failed selection retains nullOAN and suppresses later duplicate label checks.
Mismatched INSTANCE operator arguments remain OpArgNodes rather than being
regenerated as zero-argument expressions. Ordinary application level violations
use source code 4205, not the lambda arity code. Obsolete nested-standard-module
conflict approximation is removed; conflicts arise at actual nested module units.

The fixed-parameter comparison matches all 121 primary error-corpus fixtures. It
concerns fixed-code order/count and displayed parameter values; it does not
prove all diagnostic messages, ranges or semantic object types match.

Module definition validation preserves the original binding and arity. For
ordinary operators, duplicate validation occurs before body generation and the
new OpDefNode constructor's symbol conflict afterward. Rejected function
construction preserves whether its formal context is pushed. Function domains
are generated once per syntactic group, then formal names, before symbol
validation. Bound-symbol and numeric/string leaf-selector errors retain source
messages and locations.

Module-instance names validate argument count before operand generation and
incomplete-name failure. Receiving formal arities choose expression versus
operator-argument generation. Expression arity failures and operator, expression
and lambda mismatches precede incomplete-name errors. GeneralId operator
arguments resolve the operator without generating attached expression arguments,
matching Java's behavior. Fifty bounded full-message observations match errors,
warnings, ranges and ordering. General operand and compound-selector generation
still needs broader reconciliation; this is not complete Generator coverage.

Module recursive functions complete only zero-arity declarations. Rejected
functions preserve the original operator signature and unfinished binding.
Generation follows source domains, declaration validation, then function body.
INSTANCE declaration-level matching follows `ModuleNode.isConstant`, including
local operator bodies and EXTENDS theorems. Instantiated theorem context
bindings do not enter the module's theorem vector. Declared operators retain
operand levels and parameters, as in `OpApplNode`. Declaration-level errors
precede non-Leibniz checks; INSTANCE errors use the full instance range and
original messages and argument indices. Twenty-eight additional full diagnostic
observations match Java. Empty operator applications such as `F()` now follow
Java's grammar. Optional argument lists use two-token lookahead and leave the
opening parenthesis untouched when the mandatory first argument cannot start.
The source `OpOrExpr` failure boundary precedes expression production entry.
`OpArgs` and `BangExt` record their actual production frames. Saved failed
argument lookahead lengths contribute to ParseException's following-input
rendering, as JavaCC rescanning does. Twenty-two bounded observations match
complete source parse messages; the 196 accepted lookahead token kinds match
pinned generated Java. This is argument-production evidence, not complete
JavaCC lookahead or parser parity.

Restricted action subscripts use the source identifier-only `NoOpExtension`
production. Its final argument list remains detached in the parser's shared
fairness hook; prefix arguments remain attached. `FairnessExpr` uses that hook
to construct the original five heirs, or reattach the subscript call when a
separate action follows. Nested hook changes and malformed-list errors follow
Java. The translator now consumes those heirs directly; the call-style fairness
reconstruction fallback is removed. `WF_` and `SF_` are tokenized at the source
identifier boundary, with their suffix lexed normally.

Fairness and action applications use the existing builtin argument maxima:
subscripts at most state level, actions at most action level. Their operands
are checked first, and an invalid child suppresses a redundant parent diagnostic.
Forty-two bounded observations match parse messages, syntax heir kinds/images
and complete semantic diagnostics. Empty-node location sentinels remain a
separate unverified boundary. No complete parser or semantic-graph parity is
claimed by this matrix.

LET expression generation carries recursive declaration identities, levels and
shared module counters, replacing the nested-name scan. Declarations remain in
the module vector after lexical scope exit; each LET resets its level count,
checks unfinished declarations, subtracts without clearing that count and lowers
the level before IN generation. Wrong-level definitions preserve the original
binding. Function completion precedes its body; operator completion follows it.
Selector preparation retains unfinished signatures; LET failures stay attached
to their expression until actual generation reaches it, without diagnostic
sorting. Application checks use the retained binding's arity. Twenty-nine bounded
lower-phase observations match source diagnostics and recursive exceptions.
Broader constructor, formal/bound context and operand-generation parity remains
work; these observations do not establish complete LET/Generator coverage.

Proof references retain the original fact/DEF/MODULE distinction and source
syntax, including leaf-only proofs. Direct GeneralId facts allow complete module
instance names; their operands and embedded references use expression mode.
Rejected DEF and MODULE entries are omitted from Java's generated vectors;
failed expression facts occupy a slot. An empty resulting leaf reports Empty BY.
Selector errors stay attached until reference generation, preserving source order.

Reference contexts retain qualified proof steps, DEFINE operator/function and
instance bindings, TAKE/PICK names and NEW declaration arities. Assertion NEW
names are visible in their own subproof; SUFFICES NEW names become visible only
after the entire subproof. Definition-step names have a distinct semantic kind
for fact and DEF validation. Forty-nine bounded semantic-phase comparisons match
Java diagnostic messages, ranges and ordering.

Statement and DEFINE-body generation now precede each step's proof. Mixed
DEFINE steps retain the original proof-step syntax, and both semantic generation
and selector preparation traverse its definition heirs in order. Module definitions
therefore become visible to later operators in the same step, without admitting
forward references. LET and
DEFINE share operator/function generation, preserving duplicate/body/constructor
order, first signatures and recursive identity. A failed DEFINE function leaves
Java's definition vector empty and raises its original no-message array-bounds
exception; earlier diagnostics survive. PICK introduces names in its formula,
removes them during the entire subproof and installs them afterward. Pseudo-step
and ASSUME/PROVE expression errors arise at actual reference generation.

Formal operator bindings retain arity and source location. Higher-order calls
with matched argument counts generate operands from receiving formal arities,
then perform operator-constructor matching. Existing higher-order level constraints
remain checked. Unknown context metadata does not manufacture arity zero: this
fixes imported Len in an INSTANCE substitution's LET without changing the original
EWD998 debugger test. A further 57 full diagnostic observations and one lower-phase
exception observation match Java. These are bounded comparisons, not complete
coverage. Non-local proof INSTANCE syntax is now retained; proof-local substitutions
follow default construction, explicit RHS generation, duplicate detection,
remaining implicit arity checks and completeness. Illegal targets skip the RHS.
Operator-generation failures preserve the nullOpArg sentinel and its follow-up
arity error, while lambda body errors preserve the lambda operator. Module-definition
formals remain scoped to substitutions, conflicting imports keep first bindings,
and named-instance prefix arguments are validated separately from the imported
operator's arguments. Context.getByClass enumerates Hashtable buckets and chains,
including the effects of builtin and unrelated entries on rehashing. Reuse the
existing Java-compatible context traversal before filtering declarations or
importing operator and theorem classes. Four context-order and 31 INSTANCE
observations match Java. Theorem conflicts preserve their original declaration
location. Full inherited instance contexts, constructor identity, symbolic
operand failure branches and instance levels still require reconciliation.
Operator-operand selection now rejects named prefix arguments before generating
those arguments or resolving later components, preserving the null result and
subsequent constructor/substitution arity diagnostics. Nine bounded symbolic
comparisons cover this shared higher-order and substitution boundary. Proof `@`
shorthand exposed by the complete root gate now retains the actual previous
infix RHS as a `$Nop` reference, without regenerating that expression. Per-depth
proof histories preserve nested scopes and the source reset rules. Undefined
`@` emits Java's diagnostic and nullOAN result. Eleven bounded diagnostic
comparisons and the unchanged native proof-assertion test pass. Complete proof
level-node/graph reconciliation remains work. Module namespace
resolution, typed NEW/bound contexts, general failure boundaries and hierarchical
proof level checks remain work. Existing corpus passes do not establish complete proof-graph or Generator
parity. The fixed-parameter comparison selects 121 primary fixtures; earlier prose
stating 126 overstated that selection. Supporting dependency modules are not
additional primary comparisons.

## Module loading and parse-unit relationships

The loader follows SpecObj's unresolved-name loop, restarting at the root after
each binding. It exhausts EXTENDS before searching INSTANCE and checks a
module's direct unresolved names before traversing resolved extendees,
instancees and inner modules. INSTANCE traversal retains source syntax order,
including LET definitions. Inner names obey declaration visibility; only
EXTENDS exports inner modules.

A parse unit represents one source file. Its ordered extendee and instancee
relationships drive cycle checking and semantic ordering; inner modules do not
become independent external processing or lint iterations. Preserve the source
extendee-vector mutation during cycle traversal and its short-circuit search
flag behavior. E4222 aborts retain the complete filename cycle.

Nested definitions see preceding inherited INSTANCE bindings with their actual
source positions, node kinds and arities. After generating an inner module's
body, SymbolTable.addModule also checks the already generated external module
table; its conflict retains source E4223. Further semantic graph construction,
context iteration and generation traversal remain audit work.

## Checker construction

Parent construction creates coverage cost models, initializes liveness, caches
TLCGet("config") and schedules source termination before concrete checker storage.
Liveness disk-opening errors propagate immediately. DFID applies its worker and
liveness assertions only after that parent constructor completes, before its
fingerprint set and workers. BFS creates queue and trace, then initializes the
selected fingerprint set and workers. Configuration-taking native construction
defers factory selection to that phase; explicit prebuilt objects remain native
injection options.

Disk graphs require their metadata directory to exist, as Java does. They do
not create missing parents. LiveCheck propagates each subchecker's opening error
before publishing the check. RandomAccessFile open failures retain
FileNotFoundException and the native operating-system reason. Its later I/O
failures remain IOException. Further trace/storage constructor failure boundaries
and RandomAccessFile rws/rwd modes remain audit work.

## Message formatting

`MP.getMessage0` first constructs a template, then substitutes parameters in
order. The port follows those two steps exactly once for normal and nullable calls. Missing
arguments keep their markers; replacement stops at the first null. A parameter
can introduce a later marker, which is then substituted. Preserve malformed
source markers literally, including the missing final percent in the argument
mismatch diagnostic.

All 245 source cases are covered: `output_mp_generated.go` contains the 225
literal append sequences, and `output_mp.go` ports the 20 conditional cases.
`go generate` in package `tlc` reads pinned sibling Java MP.java/EC.java. The
strict generator rejects unfamiliar expressions, changed case counts and any
change in the manually ported conditional source digest. It needs no Java VM.

Argument counts select the source's exact branches, including empty bodies for
unsupported counts. Message class distinguishes pool cleanup's error/warning
text. Tool mode selects simulation/DFID progress and back-to-state text; debug
mode selects state fingerprint text. Banner PID presence is checked against
original parameters before substitution: null counts as present, an empty string
does not, and a missing element raises the source array-bounds exception. Unknown
codes retain Java's wrong-invocation diagnostic. Recorder and console boundaries
remain separate.

MP printer recorders receive raw parameters before formatting. Ordinary messages
and bug reports check suppression after notification and skip formatting when
suppressed. Warnings check elevation first, then notify, then check the global
warning switch. Enabled warnings format and enter history even when individually
suppressed. Errors always format; states format and return their message even
when suppressed. Recorder callbacks can change controls, change tool mode, or
throw before formatting. Do not attach derived console text to raw events.

## Exception printers

`PrintTLCRuntimeException` preserves MP's object-valued recorder event when the
exception carries parameters. It reads those parameters after notification and
skips formatting if suppressed. Nullable entries and empty arrays retain their
distinct source meanings. Exceptions without parameters use the legacy throwable
printer, which records the substituted detail and prints errors regardless of
suppression. The runner and simulator call the same production entry point.

Throwable error overloads retain their separate source rules: a String cause
with GENERAL builds ECGeneralMsg; a String[] cause formats its ordinary template.
Non-GENERAL errors and the array overload print stacks only under TLCGlobals.debug.
GENERAL's stack policy follows MP.noDebug and nullable messages independently.
Warnings with a throwable record both arguments, check elevation first, and print
a stack whenever warnings are enabled, including suppressed or duplicate text.
Null throwable dereferences preserve the source exception boundary. Stack frames
describe actual Go execution; do not fabricate JVM frames. `DebugPrintMessage` and `DebugPrintThrowable` preserve util.DebugPrinter's
separate process stdout/stderr, global debug guard, exact source prefixes and
actual native goroutine IDs. MP entry/parameter/exit diagnostics retain source
ordering and omissions on exceptional exits. GetTLCBug formats without recording.
Successful parameter parsing reports the source argument string; BFS construction,
init, run, liveness and cleanup report their source phases. DFID parent constructor
failure ordering and full native CLI load/welcome ordering still need source audit.

## CLI result wrapper

The native CLI returns Result.ExitStatus directly after TLC.process. Source TLC
main maps the checker code through EC.ExitStatus. Do not replace distinct TLC
violations with the SANY semantic-error status, duplicate already handled
exceptions on stderr, or append a second completion summary. Non-TLC parser and
semantic commands retain their own exit contracts.

## Runner completion and memory reporting

TLC.process preserves the checker result through its final I/O cleanup. It
ignores user-output flush/close IOException, prints the finished message, then
generates a trace spec. Generation reports its own I/O diagnostic without
promoting that failure to the checker result. The generator ignores mkdirs'
boolean result; file creation supplies failure details. Its resource close
shares the I/O catch and preserves a primary write failure. Completion is
reported before leaving the resource body, as in the source.

Model-checking and simulation startup banners use the configured maximum heap
budget and direct-memory budget in MiB. Native process heap reservation is not
the source Runtime.maxMemory equivalent; use the same native budget functions
that fingerprint configuration uses. Runtime vendor/version fields describe Go.

## State-generation functor defaults

Go callback-backed `StateFunctor` and `NextStateFunctor` preserve Java interface
default-method contracts. Unimplemented `IStateFunctor.setElement` and unary
`INextStateFunctor.addElement` return a message-less `UnsupportedOperationException`
through their native error return. Unimplemented `hasStates` raises that runtime
exception through its boolean-only signature. Overrides still dispatch to their
provided callbacks. The default `getStates` returns an empty `SetOfStates`.
These contracts matter for debugger state selection and probabilistic enumeration.

## Fingerprint memory and original model-test settings

Java `TLCRuntime.getNonHeapPhysicalMemory` reads the VM's
`-XX:MaxDirectMemorySize` argument, falling back to 64 MiB. The native equivalent
is `TLAGO_MAX_DIRECT_MEMORY`: bytes or a case-insensitive `k`, `m`, or `g` suffix.
Its parsing and signed shifts follow Java; off-heap configuration still ignores
`-fpmem` and divides this budget among nested fingerprint sets.

The shared original model-test runner applies `customBuild.xml`'s `test-dist`
settings before parsing TLC arguments: `tlc2.tool.fp.FPSet.impl` selects
`OffHeapDiskFPSet`, and the direct-memory limit is `512k`. It retains the parsed
configuration, including explicit test flags. Replacing that configuration with
an arbitrary small MSB set would change the feature being exercised.

## Fingerprint checkpoint file operations

`FileUtil.replaceFile` uses `Files.move(REPLACE_EXISTING)`; the native helper
renames directly, preserving the live file when a missing source causes failure.
Do not delete the destination before renaming. Named checkpoint commit failures
remain typed `IOException` with the original `DiskFPSet.commitChkpt` message.

Checkpoint copying follows `Files.copy(REPLACE_EXISTING)` for fingerprint files:
no parent-directory creation, same-file no-op, destination-link replacement,
source file permissions, and typed I/O propagation. A destination symlink is
replaced by the copied file; its referent remains untouched. Failed data copying
removes the partial destination. The source remains open throughout the copy.

## Off-heap shared eviction barrier

The source singleton uses Phaser with a monotonically increasing registered
worker count. A successful advance flushes all sets, clears the pending flag,
then advances the phase and releases waiters. If an eviction throws, it stops
at that set: the flag remains pending and the phase has all parties arrived.
Waiters remain in that phase; another arrival raises IllegalStateException.
Registration waits behind an incomplete advance and retains Phaser's 65,535
party limit. Phase numbering wraps at 31 bits. Model-test class isolation creates
a fresh singleton only after the previous runtime's workers have joined.

Duplicate primitive merging and full-flush validation are distinct source
operations. The original testMergeDuplicate invokes the primitive merge and
checks its deduplicated output and warnings. A duplicate submitted to the whole
preallocated flush can instead fail index validation before file-count
publication. Do not relax full-flush assertions to mimic the primitive test.

## Off-heap recovery and flusher dispatch

`DiskFPSet.recoverFP` has exclusive access during recovery. Its off-heap virtual
insertion starts with the source CAS/probe routine; only probe exhaustion signals
the shared barrier and falls back to ordinary `put`. A table reaching capacity
flushes the currently selected flusher directly, retaining its executor state
and checked-I/O boundary. It does not select a new flusher or count an eviction.
Duplicate recovery uses the coded runtime exception; warning mode continues to
the same capacity check. The initial insertion deliberately follows Java's
memory-only routine rather than adding a disk lookup.

Public invariant checks share that current-flusher path. Normal eviction selects
the flusher and wraps checked I/O, increments its growth counter even for an
empty table, and records elapsed flush time only after success. Its input/sorted
assertion details and repeated sorted-check diagnostic evaluation follow Java.

## Disk fingerprint streaming and recovery

`DiskFPSet.Flusher` merges buffered disk input with new sorted entries into a
preallocated buffered temporary file. MSB traverses its table iterator directly;
LSB uses its prescribed sorted buffer. Index entries and disk-write counts are
built with each write. The Go port must not materialize the complete disk file
or copy the full MSB table. Collision and invariant scans likewise read
sequentially through `BufferedRandomAccessFile`.

Named-file recovery follows Java `DiskFPSet.recover(String)`: it preserves the
table and existing file-tail semantics, streams the checkpoint into the current
file, rebuilds the index while counting writes, and checks signed ordering and
index length with `SYSTEM_INDEX_ERROR`. It is distinct from the normal no-argument
checkpoint no-ops and `recover(TLCTrace)` replay. The public invariant method
checks order; its expected-size overload compares `Size()`. Collision/invariant scans propagate IOException through the same typed-panic
boundary as native FPSet put/contains; a false ordering invariant is distinct
from an I/O failure. The invariant scan also propagates source finally-close
failures. Flusher IOException wrapping and rename runtime causes preserve
Java's exception boundaries.
Flusher validation has two separate source boundaries: off-heap ordering and
index alignment are checked against the temporary output before file-count
publication and replacement; both heap and off-heap flushers then scan the
reopened file for its count, signed order and index endpoints. The latter
preserves the reader cursor on successful completion and is distinct from the
public order-only invariant scan. Dedicated readers close in source order before
the output file, with the first I/O failure propagated through the flusher catch.
Off-heap sequential and concurrent merges share the source outer flusher
lifecycle, including reader seek/pool close, full preallocation, temporary-file
replacement and early file-count publication. Its eviction entry point converts
checked merge I/O into `OffHeapRuntimeException`; public invariant validation
keeps the separate checked-I/O boundary. Neither path silently recreates removed
metadata directories or reopens readers after a failed rename.
See `HANDOFF.md` for current verification snapshots and live jobs.

## Checkpoint and random initialization

TLC.Process restores UniqueString identities before constructing the tool, as
Java TLC.process does. ModelChecker and DFIDModelChecker recovery must not reload
the table after parsing: parsed strings would otherwise retain different tokens
from recovered values, causing extra fingerprints and search states. The original
three-worker CodePlexBug08 archive recovery test now matches Java recovery/final
counts, graph sizes and complete trace. See PORT_PROGRESS.md for verification.
Recovery also precedes fingerprint-polynomial initialization and enumerable-seed
assignment. Failure must not change those globals or choose a new implicit seed.
Simulation chooses its implicit seed without applying aril, but preserves the
parsed aril field. An explicit seed applies aril through the simulator RNG.
Enumerable seeding occurs once before tool construction, preserving random
consumption during configuration/constant evaluation.

2026-10-05 independent heap table counters:
HeapBasedDiskFPSet increments its LongAdder table counts without a global monitor.
Native atomic tblLoad/tblCnt increments now likewise run outside the metadata
mutex. Bucket-capacity metadata retains its Go lock, and source stripe locking
still protects table entries. Counter reads remain the source independent reads.
No test inputs/assertions or memory budgets changed; see HANDOFF.md for gates.

2026-10-05 full original binary dump/load assertions enabled:
DumpLoadTraceTest now includes both EWD840 binary methods and is complete:
32 enabled methods and three original Ignore translations. The two methods use
the existing full source dump/load helper after original graph/value buffering
was restored. Binary format, polynomial4, worker settings, partial liveness
checks and complete state/ordinal equality or prefix assertions remain unchanged.
The entire unchanged Java class and complete Go class pass; Go also passes race
checks. Historical paced Java replay failures remain source timing evidence,
not grounds to weaken the assertions or declare universal scheduling stability.

2026-10-05 disk fingerprint reader ownership:
A disk lookup retains one BufferedRandomAccessFile for its entire interpolation
search. Indexed worker readers require no pool mutex; fallback readers acquire
and return through the source pool. Normal completion returns the reader;
I/O failures propagate before that step. Adding/reopening readers atomically
publishes an immutable snapshot, matching the source array-reference read while
avoiding races on a Go slice header. Pool close errors propagate as in Java.
See HANDOFF.md for original/native assertion comparisons and running gates.

2026-10-05 long disk-queue growth verification:
The original growth test now performs all2,147,483,648 enqueue operations and
checks the64-bit queue size. One Dummy-equivalent state has workerShort.MAX_VALUE,
uid0,level1 and no value payload, matching its inherited Java base serialization.
The source class has no spec variables; its native fixture keeps that context
until background pool threads stop, then restores outer statics. Pool storage
is approximately14GiB; no compact replacement or reduced loop is used. Both
complete Java and Go ten-method classes pass. See HANDOFF.md for exact checks.
No production changes were required for this original test.

2026-10-05 source buffering restored:
Sequential ValueOutputStream uses the original 8,192-byte BufferedDataOutputStream
above optional gzip and flushes/closes that chain. ValueInputStream eagerly fills
BufferedDataInputStream; its File overload retains FileUtil's additional buffer.
The original byte-array queue adapters remain direct and disable handle tables.
Graph node/pointer files use BufferedRandomAccessFile instead of sequential
ValueStreams at shared seek offsets. GraphNode/BitVector share only the primitive
data protocol used by those two source types. File length and cursor are logical
buffered values; flush and both graph reset variants follow the source operations.
Graph checkpoint longs remain direct DataInputStream/DataOutputStream primitives.
InternTable uses original buffered data streams and atEOF. Peeking at the backing
reader after an eager prefetch loses records. UniqueString reads/writes use the
source primitive string protocol; no extra length field or encoding changes.
See HANDOFF.md for exact verification snapshots and TODO_TEST_PORT.md for inventory.
No new persistent tests or test credits; all original assertions retained.

2026-10-05 actual OffHeap CAS insertion restored after3c274be:
Source put/contains use atomic array words and CAS without a set-wide monitor.
Removed native s.mu serialization from both paths; existing per-slot CAS retry,
duplicate detection, atomic counts and shared eviction quiescence remain intact.
Shared eviction flag now atomic: CAS selection, no barrier-mutex acquisition on
ordinary nonpending operations, source compareAndSet(true,false) assertion before
releasing waiters. IncWorkers now asserts num==initialized numThreads as Java.
No new persistent tests or weakened source assertions; inventory credit unchanged.
Unchanged Java9710 terminal exit1/retired, source-supported two workers with all
8,388,608 scalar insertions: completes eviction,8,416,189 total puts (quota
overshoot), zero collisions; fails original upper bound atline182. Native17875
race terminal exit1/131.49s,8,439,146 puts, same source upper-bound failure and
no race reports. After final reset CAS assertion audit, native2746 normal terminal
exit1/54.50s,8,429,817 puts, same upper-bound failure. Both retired. Scheduling
changes exact overshoot; size==overallPuts and eviction assertions hold. Logs
concurrent-fpset-cas-eviction-{original-java,native-race}.log and concurrent-fpset-
cas-final-eviction-native-go.log. Configured audits do not credit default2.1B cases.
Other configured nine-method24685 race terminal exit1/same source upper bounds,
no races; default partitioned57170 race terminal exit1/5.51s/source lower bound,
all48 producer counts match original, no races; both retired. One-worker source-
supported OffHeap three whole methods20480 entries ALL pass17.04s/race94947;
retired. Original98082 terminal exit0/retired: all23 OffHeap (161.78s, includes
99,999,999-entry index), all3 OffHeapLong(.55s) and three related native checks
pass race, offheap-cas-original-related-race.log. Before final reset guard audit;
final guard exercised by full configured eviction above. No failure credited pass.
Whole52175 still LIVE at3c274be counter snapshot BEFORE CAS; standalone TLC88405
LIVE at CAS snapshot before final reset guard. Logs fpset-atomic-counters-final-
workspace-go.log/offheap-cas-final-tlc-go.log. No duplicate whole run started.
Next source audit found ValueOutputStream/InputStream constructors omit source
8192-byte BufferedDataOutputStream/InputStream. Must reconcile serializers and
random-access primitive adapters before applying buffering; no stream change yet.
Prepared ignored long_disk_state_queue_growth_draft_test.go with exact2,147,483,648
enqueues of one Dummy-equivalent object: workerShort.MAX_VALUE,uid0,level1,no value
payload, Empty-state identity restored at cleanup. Source method never dequeues,
fingerprints or invokes other dummy stubs. Full growth remains unexecuted and
uncredited; raw state bytes14GiB before filesystem metadata. Do not substitute a
smaller loop or variable-valued state. Goal active; deferred topics unchanged.

2026-10-05 fingerprint counter races fixed while auditing concurrent generators:
Staged complete original four producer bodies and twelve concrete method contexts
in ignored concurrent_fpset_full_draft_test.go/overlay. Preserve source default
2,147,483,649 insertions, runtime worker count, seed15041980+id, short-circuit
Size frequency, signed1024-item sorting, private TestLongVec raw-array/count
formula, partition arithmetic, start barrier/timer and normal-only latch countdown.
Source Long.getLong decoder and original excludes property retained; no invented
persistent tests, no reduced default bounds and no new translation credit.
Unchanged Java nine random methods at source-supported properties workers2 and
insertions20000 all fail size<=INSERTIONS+NUM_THREADS, line182. Generator quotas
allow a fast producer to continue until slower producers finish their individual
quotas; batch1024 further overshoots. Counts can vary with scheduling. Java79784
wrapper terminal exit0 with nine individual exit1 results, property-audit-original
logs retained, retired. This verifies configured source failures, NOT default
2.1B-entry workload failures or full completion of those methods.
Native69190 same nine configured methods initially fails source assertions AND
reports real port counter races (terminal exit1,2.021s; retired). GetLoadFactor
read bare table count during writes; OffHeap Size inherited heap table locking
which did not protect OffHeap writers. Java table count/load use LongAdder;
source size independently reads table and file counts without taking table locks.
All native table-count/table-load/disk-count accesses in heap/offheap insertion,
flush, recovery, index and concurrent merge paths now consistently use atomic
loads/adds/stores. Size now follows independent source reads; GetLoadFactor and
management count getters read safely. Existing per-bucket/table/file coordination
and all source assertions remain unchanged; no table-lock workaround introduced.
Corrected93790 nine-method configured race terminal exit1/source upper-bound
failures only, no race reports; retired. Corrected40845 full default OffHeap
partitioned race terminal exit1/22.62s, source lower-bound failure at8,388,575,
all48 original producer counts unchanged and no race reports; retired. These
expected source-failure receipts are not successful test-gate claims.
Original Java nine methods at source-supported workers1/insertions20480 ALL pass
including full invariants (37625 terminal exit0/retired). Matching nine native
whole-method configured race checks ALL pass20.94s (88137 terminal exit0/retired),
concurrent-fpset-one-worker-native-race.log. These audit generator/factory helpers,
not the default huge contexts; inventory counts unchanged. Source configurable
properties are preserved, never substituted as default test bounds.
Production final68125 terminal exit0/retired: ALL52 persistent original contexts
race pass across five top-level methods: longLSB2.07s,longMSB4.87s,OffHeap23
164.09s (includes full99,999,999 index),OffHeapLong3 .58s,ShortDisk24 54.81s.
Log fpset-atomic-counters-original-race.log, no race reports or weakened inputs.
Old full12436 terminal GREEN atdd8f136: root1808.606s,SANY1.027s,TLC421.843s,
offheap-concurrent-flusher-final-workspace-go.log; retired. Predates queue/new
simple-fill/counter changes. Fresh full52175 LIVE at current atomic-counter
production snapshot, fpset-atomic-counters-final-workspace-go.log; no duplicate.
Removed ONLY938 obsolete private Go compiler-cache blobs older12h (15.18GiB),
regenerable build outputs. Source/fixtures/logs untouched. Free disk16GiB after
checks; full stress file merge can require two16GiB files, so resource question
remains relevant. All current source-failure/stress decisions remain unanswered.
Main1251/1269,shared55/56,long14/22,concurrent2/17 unchanged. TODO/mapping updated
with configured failure evidence and exact credit limits. Goal ACTIVE.

2026-10-05 original concurrent OffHeap partitioned default failure verified:
Compiled unchanged MultiThreadedFPSetTest, OffHeap subclass and original generator
sources. Original JUnit method runs with source default48 workers,64MiB offheap,
INSERTIONS=2,147,483,649 and all per-thread bucket traversals intact. Java94369
terminal exit1/retired: all8,388,575 puts complete with zero collisions; fails base
size>=INSERTIONS at MultiThreadedFPSetTest.java:181. Source generator explicitly
ignores INSERTIONS and instead fills floor8,388,608/48 buckets per thread, with
worker0 skipping fingerprint0. Thus even the original source cannot satisfy its
minimum-size assertion at defaults. No Go production shortcut implicated.
Ignored default-property full native draft44125 terminal exit1/retired,1.016s;
size8,388,575 and same lower-bound assertion. Verified ALL48 producer counts equal
source: worker0=174761, other47=174762; all collisions0. Original invariant and
local bucketCapacity==0 assertions remain intact but are not reached in either
run. Native scratch reporter/lifecycle scaffolding is uncredited; no persistent
new test or translation count assigned. Logs concurrent-offheap-partitioned-
original-{javac,java}.log and concurrent-offheap-partitioned-draft-go.log; ignored
concurrent_offheap_partitioned_draft_test.go and overlay retained for next audit.
TODO marks this method Reconcile with verified pinned-source failure instead of
an unexplained missing case. Counts unchanged main1251/1269,shared55/56,long14/22,
concurrent2/17. Whole12436 still LIVE at dd8f136 before later queue/simple-fill
commits; poll exact handle. Local free disk now2.6GiB, full random stress workloads
still require >=16GiB per file. Pending source behavior/resource questions remain
unanswered; no new skip/reduced bound or source behavior changes. Goal ACTIVE.

2026-10-05 long LSB/MSB inherited simple-fill contexts complete after 9f8e452:
Ported FPSetTest.testSimpleFill in both concrete long classes, preserving direct
LSBDiskFPSet/MSBDiskFPSet construction, supplied default configuration, init(1),
source filename and all four unrolled put/contains pairs. OffHeap inherited
simple-fill now shares the same exact body. Cleanup closes native resources;
no reduced memory budget, source assertion changes, invented tests or new skips.
Unchanged pinned Java methods both pass runs=1/failures=0/ignored=0 with -ea,
long-heap-simple-fill-original-{DiskFPSetTest,MSBDiskFPSetTest}.log. Both source
class files compiled unchanged. Java oracle uses -Xmx256m as process memory limit;
Go constructors retain their native default memory calculation and configuration.
Go96720 terminal exit0/full two contexts race9.142s,
long-heap-simple-fill-initial-race.log; Go8259 terminal exit0/shared OffHeap
simple-fill race1.284s, long-heap-shared-offheap-final-race.log. Both retired.
Long appendix14/22 contexts,8 pending: LSB/MSB two huge methods each, OffHeap
three (including verified source MultipleFlushes failure), DiskStateQueue growth
one. Main1251/1269,609/626; shared55/56; concurrent2/17 unchanged. TODO and ignored
mapping updated. Whole12436 still LIVE at dd8f136 snapshot before queue and these
new contexts, offheap-concurrent-flusher-final-workspace-go.log; no duplicate run
or current full-suite success claim. Full stress workloads retain original bounds;
local space question remains unanswered. Goal active; deferred topics unchanged.

2026-10-05 shared util StringHelperTest complete after0643052:
All20 original JUnit3 methods in string_helper_java_test.go: exact two word-array
inputs/leading-space input, all five copy counts, six onlySpaces assertions,
complete seven-row front/end trims and unchanged-text rows, five leading-space
counts, eight identifier assertions and six strict NullPointerException catches.
Source commented TODO methods remain inactive/uncredited; no invented tests.
Production helpers preserve doubling algorithm/negative-copy empty result/null
concatenation, Java trim<=0x20, Character whitespace exclusions/C0 separators,
ASCII regex word separators/trailing-empty split behavior and UTF16-char letter/
digit scans. StringHelper platform newline comes from startup property/native OS.
SpecWriterModuleClosingTag previously substituted77 for nonpositive width; source
copy semantics now used directly, with source platform newline in this method.
Ordinary source callers still explicitly pass77. No default-size replacement.
Unchanged source Java20 methods pass.011s (string-helper-original-java.log).
Go77423 terminal exit0/full20 plus native SpecWriter checks race1.031s,
string-helper-final-race.log. Root25033 terminal exit0/all four original
SpecTraceExpressionWriter methods race1.088s, string-helper-spec-writer-race.log.
Both retired. Scratch Java/Go Unicode-table probe compares whitespace/letter-or-
digit/digit flags for EVERY65536 UTF16 char; byte-identical SHA256
646c1a146a3ac62262537fe76ec25501b78226b01dd5b19952421cb73afbd7dc.
Probe files StringHelperCharacterFlags.java/string-helper-character-flags.go and
string-helper-{java,go}-character-flags.bin are ignored scratch, not invented
persistent tests or credited original contexts. Actual source predicates match.
Shared util inventory43/56 contexts (previous23/56); remaining ExecutionStatistics
Collector12 and TLCRuntime1. Main1251/1269,609/626 unchanged. TODO/mapping updated.
Whole88583 remains LIVE at Monolith snapshot before StringHelper/writer changes,
monolith-util-final-workspace-go.log. Poll exact handle. No current full green
claim. Last green whole65591 SHORT snapshot remains older as documented below.
Space inspection1564 terminal: no private Go-cache blob>=300MiB. Removed ONLY
own generated scratch two2GiB-apparent CoreTXT huge.bin probes and three retired
Java CodePlex08 FL1/FL2 DOT graphs (2.7GiB actually allocated). Raw result logs
retained; no committed fixtures/source vectors deleted. First apparent-size
cleanup did not reclaim expected space; graph cleanup yields9.7GiB free observed.
Goal active; next portable shared utility batch ExecutionStatisticsCollector12:
read full source/test subclass hooks, no real submissions during test runs.
TLCRuntime original one method requires Ant UseParallelGC; retain JVM-specific
case visible rather than substituting a constant or claiming Go GC is ParallelGC.
Pending captured AliasSub2 source-behavior question unchanged; no replay changes.

2026-10-05 shared util MonolithSpecExtractorTest complete afterdfe33b7:
All five original methods in monolith_spec_extractor_java_test.go retain exact
Windows-prefixed MONOLITH_SPEC fixture, config result, non-null module/name/full
module text, literal Windows-name empty config/null module and both getConfig
assertions. ByteArrayInputStream fixture lowers to identical source characters;
module tests use actual source file and file-backed NamedInputStream metadata.
Ported MonolithGetConfig/NamedInputStream dependency and MonolithModule temp-file
creation/default charset output/delete-on-exit registration. Cleanup does not
repeat the source explicit close; NamedInputStream counter and idempotent file
close are retained. No invented tests or weakened assertions.
Actual config parser previously allowed loose markers/unicode trimming. Shared
source helpers now use exact quoted names, Java ASCII whitespace, four required
config trailing hyphens/three module trailing hyphens, source CR/LF line rules,
Java trim<=0x20 and Java dot exclusions for NEL/Unicode separators in end marker.
SANY loadMonolithModule uses same module extractor while preserving its existing
real temporary-file/provenance parsing workflow. Test module extraction and
production loader therefore share their actual marker/text implementation.
Unchanged Java classes/tests pass5/.039s (monolith-util-original-java.log).
Final54968 terminal exit0/full5 plus native monolith config/path race1.042s,
monolith-util-final-race.log. Earlier17150/84161 terminal and retired.
Final root original MonolithSpec2844 terminal exit0/race5.776s,
monolith-util-model-final-race.log; initial53365 also6.151s before end-marker
refinement. Final SANY76789 terminal exit0/race8.735s,
monolith-util-sany-final-race.log; initial23777 also8.875s before refinement.
All these handles retired. Shared util inventory23/56 methods (previous18/56);
main1251/1269 and609/626 unchanged. TODO/method map mark class Port complete.
Whole SHORT snapshot65591 now terminal exit0/root1853.440s/SANY1.041s/TLC364.572s,
short-disk-final-workspace-go.log. This is green at634db62 before OffHeap, buffered
streams and monolith changes; do not claim current whole-workspace green.
Current whole88583 is LIVE, -count=1 -failfast -timeout=60m ./..., redirected to
monolith-util-final-workspace-go.log. Poll exact handle, do not restart because
of silent output. Free space8.5GiB last observed; monitor actual failures.
Known captured AliasSub2 replay issue and pinned-Java evidence remain unchanged;
async semantics choice still pending. Goal active. Next independent utility
batch: whole StringHelper20 methods, inspect source implementation before tests.

2026-10-05 shared util BufferedDataInputStreamTest complete after04c6f15:
Ported full source BufferedDataInputStream/BufferedDataOutputStream classes in
buffered_data_stream.go before their original seven tests. Preserve8192-byte
buffers, eager refills and EOF state, zero-read runtime assertion, signed-byte
Java char conversion, UTF-16 low-byte output, binary numeric encodings, canonical
NaNs, line decoding via original default charset and open/close/skip semantics.
Native Reader bytes-plus-EOF is deferred to the next refill to match Java; reopen
clears that native pending error while retaining the source curr behavior.
All seven original methods in buffered_data_stream_java_test.go: exact10/15 and
2/10 short cases, strict EOFException catches, full8292 payload/8190 advance,
zero-return stream/strict runtime catch and SYSTEM_STREAM_EMPTY code, initial
empty EOF and two read failures, original string length/payload/int42 sentinel.
No invented tests; repeated Close cleanup applies only while still open.
Fresh original Java source classes/tests compiled unchanged ahead of frozen jar;
all7 pass0.031s (buffered-data-original-java.log). Go93612 initial full7 race
1.028s terminal; final65555 full7 plus original ByteUtils/value-stream methods
and existing UniqueString check race1.897s terminal after length0 assertion-state
refinement (buffered-data-final-related-race.log). Earlier compile17274 terminal
exit0/.011s; preparatory test compile failed only nonexistent ErrorCode getter,
corrected to Code; no verification credit for that failed compilation.
Supplementary shared util inventory18/56 confirmed contexts, previously11/56.
Main counts unchanged1251/1269,609/626; supplementary methods are not part of that
denominator. TODO/method mapping mark BufferedDataInputStreamTest Port complete.
OffHeap full TLC85410 now terminal exit0/414.941s, log offheap-disk-final-tlc-go.log.
Snapshot includes04c6f15 production/tests before these new stream declarations;
new stream batch verified separately above. Retire85410/65555/93612/17274.
Whole65591 remains LIVE at SHORT snapshot before OffHeap/new streams, log
short-disk-final-workspace-go.log. Poll exact handle, no full current-workspace
green claim. Existing captured AliasSub2 replay issue/source evidence and pending
async choice unchanged. Goal active; next shared-util candidate whole Monolith
extractor five methods, inspect complete source and native implementation first.

2026-10-05 whole original OffHeapDiskFPSetTest complete after634db62:
All23 original methods translated in offheap_disk_fpset_java_test.go: all16 exact
seeds/lengths, complete insertion/special-position/contains/invariant/memory-only
checks, all four original page sizes and shrinking-TreeSet loop conditions,
full99999999-entry index with signed-int overflow assertion/actual native seeks,
original zero-filling reader and dummy iterator virtual overrides, duplicate and
distinct merges/all10 output values/six warnings/exact warning fingerprint set.
No smaller index, substitute seed, weakened assertion or invented tests.
OffHeap constructor previously went through heap-only allocation and checks;
NonCheckpointable constructor now matches its Java DiskFPSet parent with exact
positive fingerprint budget/IllegalArgumentException boundary and no heap table.
Ported source getDiskOffset/isHigher, writeIndex and full streaming merge with
original independent branches, read counts, duplicates, monotonic assertions and
virtual markNext/hasNext. Actual eviction now sorts the array using original
comparator and consumes/marks its source iterator rather than collecting and
sorting a separate slice. Ported source input/sorted/index assertion helpers and
OffHeap invariant-flush dispatch. Existing native slice merge adapter shares the
lower stream merge; it is no longer the eviction path. Original concurrent
partitioning implementation/stress suites remain outstanding in supplementary
scope; this serial source path completion does not credit those suites.
Java49155 terminal exit0/all23 methods34.089s, offheap-disk-original-java.log.
InitialGo1117 terminal exit0/all23 normal52.027s before final assertion helpers.
FinalGo1154 terminal exit0/all23 normal53.440s, offheap-disk-final-go.log; full
index traversal53.23s. Race49218 terminal exit0/22 methods1.278s excluding ONLY
writeIndex, offheap-disk-declared-race.log; do not claim all23-method race.
Related96417 terminal exit0/race21.292s, offheap-disk-related-race.log: native
OffHeap and disk checks, original factory methods and original OffHeap iterators.
Initial native26998 also terminal exit0/.043s. All these handles retired.
Inventory1251/1269 contexts(98.6%),609/626 classes(97.3%);18pending/17classes.
Main fingerprint group165/165 methods/all20 classes complete. Supplementary util,
long-running/concurrent suites remain in scope and uncredited where unchecked.
LIVE whole65591 is still the SHORT snapshot before this OffHeap implementation,
short-disk-final-workspace-go.log; poll exact handle, no current green claim.
LIVE current TLC85410 is -count=1 -failfast -timeout=20m ./tlc redirected to
 offheap-disk-final-tlc-go.log; poll exact handle. Neither is terminal yet.
Existing AliasSub2 captured replay mismatch reproduces in pinned Java; pending
async behavior choice remains unanswered and no replay semantics/test changed.
Goal active. Next independent batch: shared util original test translations;
read full source constructors/methods and implementation first. User-deferred
models/distributed/JPF/benchmarks and forbidden email scopes remain unchanged.

2026-10-05 whole original ShortDiskFPSetTest complete after297210b:
All24 original methods translated in short_disk_fpset_java_test.go with complete
AbstractFPSetTest setup/helpers. Preserve all six source runKnown conditional
early returns and their full bodies; defaultfalse means those six bodies are
not executed, matching Java, rather than asserting known zero/min bugs fixed.
Preserve all1125 interpolation tuples/invalid-input filter/four signed bounds,
3072 low-page insertions and lookups, exact dummy/plain configuration distinction,
manual zero/min/max flushes, full2047 duplicate page loops and block comparisons.
Initial source testValues failed on adjacent MaxInt64 fingerprints rounding to
the same double: Go castNaN becameMinInt64 instead of Java zero. Actual disk
implementation now uses separate source double conversions, javaDoubleToLong
and unconditional equal-high decrement. LSB prepareTable now exactly narrows
count toint32/assertspositive and allocates fixed-length sorted buffer, retaining
unfilled zeros as Java does rather than dropping them with append. Enabled disk
zero/min assertions pass without changing their expectations. No invented tests.
Original unchanged Java all24 pass0.373s (short-disk-original-java.log).
Go83186 terminal exit0/full24 normal18.151s (short-disk-final-go.log).
Go89953 terminal exit0/full24 plus factory/three TLCIterator variants/native disk
race76.002s (short-disk-final-race.log). These handles are retired.
Preparatory17409 failed on malformed timeout option before running tests; it
has no verification credit. Actual final command uses -timeout=20m.
Inventory1228/1269 contexts(96.8%),608/626 classes(97.1%);41pending/18classes.
Fingerprint142/165 methods,19/20 classes complete;23pending in OffHeapDiskFPSetTest.
Whole65591 remains LIVE, -count=1 -failfast -timeout=60m ./..., redirected to
short-disk-final-workspace-go.log. Poll that exact handle; do not mistake silent
output for termination. No full current-workspace green claim yet. Existing
AliasSub2 captured replay mismatch reproduces in pinned Java; async behavior
question remains pending and no replay semantics/assertions changed. Goal active.
CommunityModules progress change840d494 is confirmed in HEAD ancestry and already
committed; user can remotely rerun TestJavaCommunityModulesAnt with -v/-count=1.
Next independent correctness batch is whole OffHeapDiskFPSetTest23 methods.

2026-10-05 whole original MSB disk-fingerprint class complete afterd835600:
All19 MSBDiskFPSetTest2 contexts translated in tlc/msb_disk_fpset_java_test.go.
Reuse complete inherited12-constructor/three-recovery source helpers with
subclass lower256/upper2147483648 and MSB constructor. Preserve all99998 trace
fingerprints/contains checks, all1024 forced-flush calls and duplicate warning;
no smaller default memory budget or loop. getMSBDiskFPSet uses original dummy
raw100-fingerprint budget, exact100 assertion, init1 and class-name/timestamp
filename. Four declared methods retain highFP1<<62, discarded new iterator
then OLD iterator getLast after manual flush, catch-only-NoSuchElementException,
lowFP1/new iterator last check, high fingerprints9223368718049406096 and
9223335424116589377 with false/true/true puts across two flushes, and no-bucket
exception. Every original assertion retained; no invented tests or production
changes. Existing native iterator/configuration implementation passes directly.
Original Java19 methods pass7.373s, log msb-disk-original-java.log; fresh source
was inspected again and sibling source worktree remains unchanged.
Go97266 terminal exit0/all19 normal206.495s, log msb-disk-initial-go.log.
Four DECLARED methods only44212 terminal exit0/race1.038s, log
 msb-disk-declared-methods-race.log. Do not describe this as all19-method race.
Related68491 terminal exit0/race21.109s, log msb-disk-related-race.log: original
three TLCIterator subclasses, FPSetFactory methods and native disk checks.
Inherited helper implementation already passed full LSB15 race963.741s; current
MSB full1024 forced-flush traversal is verified by its complete normal run.
All appropriate gates for this test-only batch are green; ready to commit.
Inventory1204/1269 contexts(94.9%),607/626 classes(97.0%);65pending/19classes.
Fingerprint118/165 methods,18/20 classes complete;47pending in OffHeap23/Short24.
No full current-source workspace green claim: existing AliasSub2 auto-worker
replay mismatch remains documented and reproduces identically in Java when
loading the captured trace. Async behavior question remains pending; preserve
pinned Java semantics until answered. No replay assertion/semantics changed.
No tests remain live;97266/44212/68491 terminal and retired. User-deferred topics
unchanged; goal remains active. Next eligible independent batch: whole Short24.
Re-read all six original runKnown conditional early returns (testZeroFP,
testMinFP,testMemLookupWithZeros,testMemLookupWithMin,testDiskLookupWithZerosOnPage,
testDiskLookupWithLongMinValueOnPage), not merely the first two. Preserve every
branch/body even when default runKnown=false; do not introduce extra skips.
All24 unchanged Java methods already pass0.373s, log short-disk-original-java.log.
Source LSB clone buffer is fixed-length cnt including unfilled zeros; Go current
append-only buffer drops those slots. Port full tests and fix actual production
if they expose this discrepancy, without weakening disk-zero assertions.

2026-10-05 validated LSB batch ready for isolated commit:
All15 original contexts and current source-derived constructor/recovery fixes
pass final full-class race963.741s, original Java1.127s and related race56.789s.
Current whole TLC package passes169.325s; whole root suite remains red on the
separate existing AliasSub2 auto-worker replay prefix case as documented below.
Captured-file replay through unchanged Java confirms identical divergent state10
and raw10-state counterexample, so no Go-source shortcut was identified for
that captured mismatch. No replay semantics/assertions were changed.
Commit the verified LSB feature and inventory independently, with the known
whole-suite failure visible; this does not claim a fully green current workspace
or completion of the goal. Main inventory1185/1269 and606/626 remains verified.
Async source-behavior clarification still pending; dependent replay changes await
an answer. No permission is needed for this already-authorized isolated commit.
No new test-port batch started; next MSB19 and Short24 source preflights remain
ready. All background test/reference handles are now terminal and retired.

2026-10-05 LSB15 final race complete; current whole-suite replay failure under investigation:
Final34156 terminal exit0/all15 original methods race963.741s; recovery1 11.10s,
full1024 forced-flush recovery2 947.92s, duplicate warning3.12s. Log
 lsb-disk-final-race.log. This verifies final configuration/constructor/recovery
changes without smaller budgets/loops. Related61730 race56.789s remains green.
LSBDiskFPsetTest15 now Port complete in TODO/method mapping. Inventory1185/1269
contexts(93.4%),606/626 classes(96.8%);84pending/20classes. Fingerprint99/165,
17/20 classes complete,66pending in MSB19/OffHeap23/Short24. Source Java unchanged.
Batch still uncommitted because current whole14785 terminal exit1/root774.106s:
TestJavaSafetyDieHardAliasSub2DumpLoadTraceJSONAutoWorkers action mismatch at
zero-based state8, dump="empty big"/load="pour big to small". SANY1.148s/TLC169.325s
pass. Log lsb-disk-final-workspace-go.log. This is not a full green receipt.
Older full33379 now terminal exit0 root1892.570s/SANY1.163s/TLC82.424s, log
 offheap-bitshifting-final-workspace-go.log; snapshot before equivalence/current
LSB code, includes final nil-hasNext refinement/standalone tests. Retire33379.
All earlier whole handles are now retired; no current whole suite live.

Replay investigation retains every original assertion/fixture/worker setting:
57522 terminal exit0/30 unchanged Go repetitions20.581s, log
 lsb-batch-dump-load-alias-reproduce.log. Higher300 repetition12419 terminal exit1
147.573s, action mismatch at zero-based9: empty big vs pour big to small; log
 lsb-batch-dump-load-alias-reproduce-300.log. No success credit for failed repetitions.
Source Java Worker.addElement explicitly checks invariants even when isInModel
rejects a successor, matching Go processSuccessorForWorker. The source JSON
constraint checks only alias variables (action here); a rejected alternate
successor can therefore terminate replay with an invariant violation before
reaching the recorded next action. Do not weaken prefix assertions or reorder
normal source invariant checks to conceal this case.
Scratch source overlay only copies the real dump before replay; all assertions
unchanged, no persistent test edits. Capture1340 terminal exit1/69.819s with
11-state alias dump. Files lsb-alias-replay-capture-overlay_test.go/.json,
 lsb-alias-replay-capture-go.log and lsb-alias-replay-captured.json. Exact captured
state10 action is pour big to small; Go replay action is pour small to big.
Original Java CLI (unchanged actual sibling production classes) loads that same
captured file and produces EXACT same10-state replay/actions/raw values, safety
exit12, state10 pour small to big,55generated/10distinct/1queue. Log
 lsb-alias-replay-captured-java.log. Source checkout clean and compiled _JsonTrace
resource cmp byte-identical to pinned source. This is direct evidence of same
replay behavior, not an assumption from ordinary Java success.
Unchanged Java test method fresh compilation/reflection/cleanup:44873 terminal
exit0/all100 repetitions;51816 terminal exit0/all1000 repetitions. Logs
 lsb-alias-replay-method-java.log and lsb-alias-replay-method-java-1000.log. These
ordinary runs did not trigger this captured longer path; do not claim they prove
replay expectation for every valid multiworker trace. First attempted JUnit
Request.method ran the whole class due non-filterable IsolatedTestCaseRunner and
exited255 on missing unrelatedMC.cfg; log lsb-alias-replay-original-java.log.
Retire as preparation failure, no Java behavior/test-pass credit. Reflection
invokes the exact source method with all assertions and explicit cleanup; no
original @Before hooks exist in this class/base. Helper ignored/uncredited.
User clarification is pending via async question: preserve pinned Java behavior
and document source test issue, or fix replay to exclude violating successors
outside the recorded trace (explicit difference from Java). Do not choose a
source-semantic change from elapsed wait/no answer. Do not weaken/remove/skip
the enabled test or claim whole green. All process handles listed above terminal.
Next port remains MSB19 after current batch is settled; no new MSB/Short persistent
implementation yet. Their unchanged Java preflights remain green19/24 methods.
Goal active, user-deferred topics unchanged.

2026-10-05 additional authoritative verification during LSB race wait:
Read-only /proc2559634/fd metadata confirms testFPSetRecovery2 backing file
contains734 fingerprints (size/8), so the unchanged1024-call loop is advancing.
Final34156 still live/uncredited; full14785 and33379 still live.
Older full83491 now terminal exit0: root1895.739s/SANY1.001s/TLC82.629s, log
 offheap-iterator-final-workspace-go.log. Covers iterator core before final
nil-hasNext exception refinement, standalone/equivalence and current LSB changes.
It is the latest completed whole green, with that exact scope; retire83491.
Unchanged MSBDiskFPSetTest2 preflight29152 passes19 methods7.373s as recorded.
Independent unchanged ShortDiskFPSetTest Java preflight is terminal exit0:
all24 methods0.373s, log short-disk-original-java.log; compile
 short-disk-original-javac.log. Original runKnown property false returns early
in testZeroFP/testMinFP; do not invent additional early exits in other methods.
Source remaining tests still run zero/min disk lookup assertions. Original LSB
prepareTable creates buff=new long[cnt], leaving unfilled zero slots present;
Go prepareLSBTable currently creates a zero-length slice and appends positives,
which removes those slots. Inspect this discrepancy when porting the whole Short
class; preserve zero "magically" appearing on disk after source explicit flush,
complete1125-value Cartesian calculateMidEntry input matrix/invalid-input filter,
all bounds and signed midEntry*8 assertion, block/contains consistency and full
helper checks. Do not weaken zero assertions or claim source conditional
returns give coverage for skipped zero/min insertions. No Short tests ported or
credited yet; this read-only audit does not change production pending LSB gates.

2026-10-05 verification continuation of active LSB batch:
Final15-method race34156 confirmed live in testFPSetRecovery2, process2559634
at7m48s/100% CPU; do not interpret silent redirect as hang or restart it.
Final full14785 remains live. Older full33379/83491 likewise confirmed live,
latest re-poll yielded no terminal status; prior observation elapsed is not exit.
No implementation edits during this continuation; all existing gate scopes and
uncredited LSB inventory status retained.
Independent unchanged Java MSBDiskFPSetTest2 preflight29152 now terminal exit0:
all19 methods pass7.373s, log msb-disk-original-java.log; compile
 msb-disk-original-javac.log. Preserve inherited12 constructor/three recovery
methods, lower256/upper2147483648 and helper dummy100-fingerprint budget/assertion;
original getLast keeps discarded new iterator and catches NoSuchElementException
from the OLD iterator before inserting1, then exact new-iterator getLast1.
Retain exact high fingerprints9223368718049406096 and9223335424116589377,
false/true/true duplicate assertions across two manual flushes, and no-bucket
exception. This class still pending/uncredited; do not start its implementation
before final LSB gate is complete and batch committed. No persistent MSB edits.

2026-10-05 active LSB heap-disk batch afterb8db764:
All15 original LSBDiskFPsetTest inherited methods translated in
 tlc/heap_disk_fpset_java_test.go, with complete AbstractHeapBasedDiskFPSetTest
constructor/recovery helpers and test-side DummyFPSetConfiguration raw-memory
override. Preserve lower512/upper2147483648, all12 exact byte budgets and all
three doTest inequalities. Preserve full99999 trace limit (99998 entries),
original predecessor.uid=1 then fp updates, trace begin/commit checkpoint,
init1/default configuration, exact size99998 and all99998 contains assertions.
Retain source ForceFlush, all1024 recoverFP calls, duplicate fingerprint1 twice
and error2warning=true. Temporary dirs and resource cleanup isolate original
filename/class-name files without altering inputs/assertions.
Go adds optional GetMemoryInBytesOverride dispatch to preserve Java virtual
configuration calls from GetMemoryInFingerprintCnt; test dummy returns its raw
MemoryInBytes, without ratio substitution/runtime clamping. Expose source LSB/
MSB auxiliary-storage requirements and use them in constructor calculations.
Replace silent underflow repair with exact source TLCRuntimeException message,
retain signed/masked Java int capacity shift and negative-cap overflow fallback,
add omitted upper-memory and negative-maxTblCnt assertions, source lock creation
before config calls. RecoverFP duplicates now use source TLCRuntimeException;
a warning continues to the original flush decision instead of returning early.
No new persistent tests, smaller tables, early flush reset or weaker assertions.
Unchanged Java38436 terminal exit0/all15 tests1.127s, log lsb-disk-original-java.log;
compile lsb-disk-original-javac.log. Initial Go compile undefined NewTLCStateMut
was corrected to native NewEmptyState UID carrier, not a test pass.
First normal76156 terminal exit0/all15 methods66.635s, log lsb-disk-final-go.log,
before final constructor lock-order/duplicate warning-flow corrections.
Related61730 terminal exit0/race56.789s, log lsb-disk-related-race.log, on final
code: original factory/MultiFPSet and native configuration/disk checks.
Final full15 race34156 remains live in testFPSetRecovery2, exact1024 forced-flush
calls/default runtime-sized table. Last read-only process inspection confirms
active CPU usage; silence is not terminal and not authorization to reduce loop.
Full current workspace14785 live, log lsb-disk-final-workspace-go.log.
Inventory still1170/1269,605/626; LSB15 remain uncredited until final gate completes.
Do not commit this batch or start another implementation before required gate.

Previous zero-reader full9378 now terminal exit0: root1875.820s/SANY1.160s/
TLC80.061s, log fpset-bug246-zero-readers-final-workspace-go.log. Includes6f3a1de,
before iterator/standalone/equivalence/current constructor changes; this latest
completed whole green does not verify those later changes.
Older33379/83491 remain live on their documented snapshots. Resume14785/34156/
33379/83491 exact handles; completed61730/76156/38436/9378 are retired.
Goal active, user-deferred topics unchanged.

2026-10-05 whole original indexer equivalence context complete afterb98941d:
OffHeapIndexerEquivalenceTest.testInfiniteInfMult translated in
 tlc/offheap_indexer_equivalence_java_test.go, all7254 original parameter rows,
including duplicates/order. Retain original isSupported assumption and explicit
Mult1024/InfinitePrecision constructors; no automatic selection or alternate
expected implementation. Complete doTest: min(1024,range) unseeded uniformly
random [lower,upper) samples and equality/monotonicity, full uniform loop with
both fp/fp+1 equalities, both1023 upper-bound-neighborhood loops with exact
termination/monotonicity, final upperBound equality. math/rand/v2 preserves
unseeded bounded random semantics; source uses ThreadLocalRandom without seed.
Two original @Ignore methods remain uncredited; no extra skips or sampling.
No production change, weakened assertions or invented persistent tests.
Unchanged Java76902 terminal exit0/all7254 tests9.235s, log
 offheap-equivalence-original-java.log; compile offheap-equivalence-original-javac.log.
Scratch Java original data() and extracted actual Go row generator produce
byte-identical7254-line CSVs, including order/duplicates; files
 offheap-equivalence-{java,go}-rows.csv. Probe ignored/uncredited.
Go74992 terminal exit0/normal20.270s, log offheap-equivalence-initial-go.log.
Combined79310 terminal exit0/race87.870s, log
 offheap-equivalence-final-related-race.log; covers full equivalence matrix,
all three original complete1104-row/five-method indexer matrices, standalone
bitshifting class, original iterator methods and native indexer test.
Inventory1170/1269 contexts(92.2%),605/626 classes(96.6%),99pending/21classes.
Fingerprint84/165 methods,16/20 classes complete;81pending/4classes:
LSBDiskFPset15,MSBDiskFPSetTest2 19,OffHeapDiskFPSet23,ShortDiskFPSet24.
Next eligible batch: whole LSBDiskFPsetTest inherited12 constructor checks and
three low-level fingerprint recovery methods. These are not deferred recovery
models; preserve99999 full trace range,1024 recoverFP calls,duplicate warning
property and original DummyFPSetConfiguration raw-memory override.
Read original AbstractHeapBasedDiskFPSetTest/DummyFPSetConfiguration first;
Go currently has no test dummy config virtual override, so inspect required
accurate constructor dispatch rather than substitute ratio or memory clamping.

Whole90170 now terminal exit0: root1889.327s/SANY1.206s/TLC74.773s, log
 fpset-bugs210-242-final-workspace-go.log. Snapshot includes page-arithmetic fix
53bb262, before zero-reader/iterator/standalone/equivalence tests. It is the
latest completed whole-suite green, not a receipt for those later changes.
Full33379/83491/9378 remain live on documented snapshots; resume exact handles.
33379 includes final nil-hasNext refinement/standalone tests, before equivalence.
No restart merely from observation timeout; no full current-source green claim.
All focused/source/related checks for this test-only batch green, ready to commit.
Goal active; five user-deferred topics remain unchanged.

2026-10-05 whole standalone bitshifting indexer class complete after8c70c38:
All five original OffHeapBitshiftingIndexerTest methods translated in
 tlc/offheap_bitshifting_indexer_java_test.go. Preserve both complete128-position
sweeps (fpBits1/logPos8 and fpBits2/logPos9), exact power/leading-zero checks,
both boundary fingerprints at every position, wraparound at positions, all four
268435456-position checks, overshoot9223371952792813846/probe5 and constructor
Integer.MAX_VALUE+1 with catch-only-TLCRuntimeException. uint64 preserves Java
long shift/wrap bit patterns; direct bitshifting constructors retained.
No production changes, reduced loops, weakened assertions or invented tests.
The three parameterized indexer classes were already complete (all1104 rows,
five inherited methods each); the previous next-step matrix note was stale.
Unchanged Java -ea passes all5 in0.007s; logs offheap-bitshifting-original-java.log
and offheap-bitshifting-original-javac.log. Go focused36588 terminal exit0/race
1.036s, log offheap-bitshifting-final-race.log. Related18301 terminal exit0/race
16.120s, log offheap-bitshifting-related-race.log: all three original complete
parameter matrices, standalone methods, iterator methods and native indexer test.
Inventory1169/1269 contexts (92.1%),604/626 classes (96.5%);100pending/22classes.
Fingerprint83/165 methods,15/20 classes complete;82pending/5classes.
Next eligible indexer batch: complete OffHeapIndexerEquivalenceTest.testInfiniteInfMult,
all7254 original rows, full1024 random/uniform/boundary loops and assumptions;
retain two original @Ignore methods uncredited, no sampling or extra skips.

Current whole suite33379 is live, log offheap-bitshifting-final-workspace-go.log,
including final nil-hasNext refinement and these five tests. Full83491/9378/90170
remain live on earlier documented snapshots; re-polled exact handles.
Old84016 is terminal exit1, timeout1h in TestJavaTLCSetSim54m16s/root3600.060s;
SANY1.052s/TLC68.802s pass. It predates exploration fix318811c; retire as failure,
not current regression or green receipt. Log simulation-models-recursive-final-workspace-go.log.
Latest completed whole green remains49471 at exploration-fix318811c:
root1859.599s/SANY1.142s/TLC69.948s, before subsequent fingerprint work.
Do not claim current whole-suite green until terminal evidence. Goal remains active.

2026-10-05 whole original off-heap iterator contexts complete after6f3a1de:
Both OffHeapIteratorTest methods are translated in tlc/offheap_iterator_java_test.go.
Preserve source LongArray.isSupported assumption, elements32, complete64-element
arrays initialized1..64, InfinitePrecisionIndexer64/1, all32 returned values,
exact count32, all64 preservation checks for next, all32 high-bit marking checks
for markNext. Preserve the upper-half loop's repeated array.get(elements) check;
do not substitute get(i) or invent stronger/different test assertions.
Production offheap_fpset.go ports missing markNext, shared mark-aware next0,
Java do/while ordering, skip-empty/evicted traversal, source wrap assertion,
bounded EMPTY return and NoSuchElementException exhaustion. Preserve null/zero
array exception families and nil-receiver exception instead of fabricated empty
results. Native CheckFPs partition loop catches only the source
NoSuchElementException; do not swallow other failures or fabricate elements.

Unchanged Java -ea passes both methods0.013s, log offheap-iterator-original-java.log;
compile offheap-iterator-original-javac.log. Go focused18633 passes race1.029s,
log offheap-iterator-initial-race.log. Final nil-hasNext refinement31282 passes
race1.027s, log offheap-iterator-final-race.log. Related22353 passes race54.183s,
log offheap-iterator-related-race.log: original LongArray/LongArrays/FPSetFactory/
MultiFPSet and related native offheap/disk checks retain assertions. That gate
precedes the final nil-hasNext exception refinement, which affects only nil
receivers; final focused gate covers current complete iterator code.
No invented persistent tests, smaller arrays or weakened assertions.
Inventory1164/1269 contexts (91.7%),603/626 classes (96.3%);105pending across23classes.
Fingerprint topic78/165 methods,14/20 classes complete;87pending across6classes:
LSBDiskFPset15,MSBDiskFPSetTest2 19,OffHeapBitshiftingIndexer5,
OffHeapDiskFPSet23,OffHeapIndexerEquivalence1,ShortDiskFPSet24.
Next complete bitshifting parameter matrix (all1104 original rows/assumptions),
then remaining heap/offheap cases and full equivalence matrix; no sampling.

Full workspace49471 is now terminal exit0: root1859.599s,SANY1.142s,TLC69.948s;
log simulation-ttrace-exploration-final-workspace-go.log. This is the latest
completed whole suite, snapshot318811c with the exploration stopping fix, before
five later liveness rechecks/page arithmetic/zero readers/offheap iterator changes.
Do not credit those later changes from that older receipt.
New full83491 is live, log offheap-iterator-final-workspace-go.log; iterator core
before only the final nil-hasNext refinement. Full9378 is live at zero-reader
work before this iterator port, log fpset-bug246-zero-readers-final-workspace-go.log.
Full90170 is live before zero readers/iterator, log
fpset-bugs210-242-final-workspace-go.log. Old84016 remains live without the
exploration fix. Resume83491/9378/90170/84016 exact handles. Do not restart from
observation timeout or claim current whole-suite green before terminal evidence.

2026-10-05 whole Bug246 disk-fingerprint contexts complete after53bb262:
Both original Bug246DiskFPSetTest methods are translated in
 tlc/fpset_bug246_java_test.go. testLinearFillup preserves runtime.maxMemory /
TLCRuntime.getFPMemSize(0.5), the original heap-bound assertion, ratio1, LSB dummy,
init0, timestamp-based filename, full getTblCapacity()-1 descending MaxInt64
insertions and every duplicate rejection. Retain last bucketCapacity/tblCapacity/
tblLoad/tblCnt/growDiskMark observations, catch-only-OutOfMemoryError, nulling/
GC, growDiskMark==0 assertion and unconditional failure with source statistics.
No fixed memory override, smaller loop or sampled inputs. testFlushDiskFPSet has
no active statements in original Java (all body commented); the Go method remains
empty. Count it as a whole translation, not evidence of actual flush coverage.

Source DiskFPSet.init allocates exactly numThreads dedicated readers plus five
pool readers. Go substituted one when numThreads==0, including opening/reopening.
Remove that shortcut; retain zero throughout initialization and reopening.
Negative counts throw the source array-allocation NegativeArraySizeException.
Remove the fabricated default-reader constant; preserve native readers/pool
lifecycle, existing pooled reads and source-derived table sizing.

Unchanged original Java -ea passes both methods0.216s, log
fpset-Bug246DiskFPSetTest-original-java.log; compile fpset-bug246-original-javac.log.
Native Go focused48195 passes race162.804s, log fpset-bug246-initial-race.log.
Actual runtime budget67528591360, memory33764295680, capacity67108864; all67108863
insertions retained. Process inspection confirmed running CPU/allocation before
terminal pass; observation silence was not treated as a hang or restart trigger.
Final related87167 passes race57.332s, log
fpset-bug246-zero-readers-related-race.log; original FPSetFactory/MultiFPSet plus
related disk/configuration/duplicate/recovery checks retain their assertions.
No invented persistent tests or weakened original assertions.
Inventory1162/1269 contexts (91.6%),602/626 classes (96.2%);107pending across24classes.
Fingerprint topic76/165 methods,13/20 classes complete;89pending across7classes.
Next original OffHeapIterator methods and full heap/indexer matrices, preserving
all rows/resources/catch rules; no new checkpoint/recovery MODEL work.

Current full9378 is live, log fpset-bug246-zero-readers-final-workspace-go.log,
-count=1 -failfast -timeout60m ./.... It includes the zero-reader fix and the full
runtime-derived Bug246 loop. Previous full90170 is live before zero-reader work,
log fpset-bugs210-242-final-workspace-go.log; full49471 is live at318811c before
later test ports/page arithmetic. Old84016 remains live and lacks the exploration
stopping fix. Do not credit current whole-suite success before terminal exit0.
Resume9378/90170/49471/84016 exact handles; no restart from observation timeout.

2026-10-05 Bug210/Bug242 whole disk-fingerprint methods complete after42719d2:
All six original methods are translated in tlc/fpset_disk_bugs_java_test.go.
Preserve DummyDiskFPSet's LSB constructor/index visibility, Bug210's exact
MaxInt32/1024+8 index and MaxInt64-3/-1/-2 values, one-thread initialized file,
false diskLookup result and IOException failure. Preserve Bug242 literal memory
2097153638/MaxInt32/MinInt32/0/1, ratio1 and uninitialized constructors. Original
large-memory cases accept only OutOfMemoryError or successful construction;
negative memory must throw Exception; other cases reject Exceptions. Do not
swallow Java Error families in catch(Exception). Source lifecycle progress logs
remain verbose. No sampled/smaller index or weakened assertion.
Source DiskFPSet.diskLookupBinarySearch widens page to long before multiplication.
Go cast the product afterward, which overflows int on386. Change both lo/hi
entry computations to int64(page)*NumEntriesPerPage. No other implementation
changes. Native amd64 focused24613 passes race3.535s; log
fpset-bugs210-242-final-race.log. Initial attempt only failed compile because Go
Close has no return, corrected before test gates; initial log is not a pass.
Unchanged originals pass Java -ea:Bug210 0.132s (1method),Bug2420.143s (5methods);
logs fpset-<Class>-original-java.log; compile fpset-bugs210-242-original-javac.log.
Original full-sized Bug210 passes Linux3860.750s,9471 terminal exit0; log
fpset-bug210-386-unsandboxed.log. Initial sandboxed41678 terminates bad-system-call
before testing (fpset-bug210-386.log); approved unsandboxed run is actual evidence.
A source overlay with only the prior multiplication order reproduces the original
Bug210 failure on386:50376 terminal exit1, negative seek invalid argument,
0.762s; log fpset-bug210-386-before-widening.log. Overlay leaves worktree untouched.
This negative control is not a green gate. Final related24149 passes race58.220s;
log fpset-bugs210-242-related-race.log, original FPSetFactory/MultiFPSet plus
related disk/configuration/duplicate/recovery tests retain all assertions.
Inventory1160/1269 contexts (91.4%),601/626 classes (96.0%);109pending across25classes.
Fingerprint topic74/165 methods,12/20 classes complete;91pending across8classes.
Next Bug246 whole insertion/flush methods and remaining heap/offheap/iterator
cases with full parameter matrices; no new checkpoint/recovery MODEL work.
New full90170 is live after polls, log fpset-bugs210-242-final-workspace-go.log,
-count=1 -failfast -timeout60m ./.... Previous final49471 remains live, snapshot
318811c before five liveness recheck and six fingerprint contexts; it includes
the exploration stopping fix. Old84016 remains live and lacks that fix.
Old59582 is terminal timeout failure at60m/root3600.043s, TLCSetSim54m3s;
SANY0.972s/TLC68.388s pass in that old snapshot, which lacks the stopping fix.
Log simulator-correctness-final-workspace-go.log. Old79605/39038 are already
terminal TLCSetSim timeout failures. No current whole-suite success claimed.
Resume90170/49471/84016 exact handles; do not restart from observation timeout.

2026-10-05 five remaining in-scope generated liveness contexts complete after318811c:
Whole LoopTest_TTraceTest, LoopTestForcedPartial_TTraceTest,
OneBitMutexNoSymmetryTest_TTraceTest, Test3_TTraceTest and
UnsymmetricModelCheckerTestA_TTraceTest methods are translated in
 tlc_liveness_remaining_ttrace_java_test.go. Preserve constructors, inherited
VIOLATION_LIVENESS/debugger/dot/JSON/no-generation/no-coverage/FP0/seed1/workers1,
resolver paths and actual generated artifacts. Full original generation assertions
are retained; five original helpers accept only destination overrides, with no
change to existing test inputs/assertions. Forced partial retains the original
AddAndCheckLiveCheck static switch across generation and recheck. Graphs are
retained/flushed/closed via existing source-parity helpers.
Loop exact stats3/3/0/init1, graph80/48, x0/1/2, stuttering4 and zero uncovered;
forced partial exact1/1/0/init1,x0,stuttering2; Test3 exact5/4/0,x0/1/0/2,loop1,
zero uncovered; Unsymmetric exactlive1/init1+empty/stats3/2/0,x=a/1,loop1 and
exact generated _init23:5-23:24/_next27:5-33:29 labels; OneBitMutex exact14/13/0,
graph380/208, all13states, loop4/_next42:5-54:37. The latter's complete Java
string literals are mechanically concatenated and byte-verified against Go.
All12 existing model input files match original Java bytes. No production changes
or weakened assertions were needed. Unchanged original generation/recheck pairs
pass Java -ea: Loop0.492/0.525s,forced0.508/0.482s,Test3 0.513/0.563s,
OneBitMutex0.966/0.829s,Unsymmetric0.511/0.532s. Logs
liveness-remaining-ttrace-<Class>-original-java.log; compilation
liveness-remaining-ttrace-original-javac.log. Go focused65048 is terminal exit0,
race55.020s, log liveness-remaining-ttrace-initial-race.log. Final related80759
is terminal exit0; ok  	github.com/glycerine/tlago	96.755s; log liveness-remaining-ttrace-related-race.log.
Original and generated models plus ChooseTableauSymmetry/CodePlex rechecks
retain all original assertions.
Inventory1154/1269 contexts (90.9%),599/626 classes (95.7%);115pending across27classes.
Generated rechecks44/45 complete; the only remaining checkpoint context is deferred.
Next in-scope correctness work: fingerprint/indexer/iterator original cases and
full parameter matrices (97pending across10classes), plus existing unresolved
safety/liveness/dump-load contexts. Do not substitute smaller parameter samples.
Full workspace49471 is last polled live, snapshot318811c before these five
test-only contexts; log simulation-ttrace-exploration-final-workspace-go.log.
Old84016/59582 were polled live and lack the exploration fix;79605/39038 are
terminal TLCSetSim timeout failures as documented below.
No current full-suite pass claimed. Resume exact live handles; do not restart
from observation timeout. Focused/related gates cover the additional tests.

2026-10-05 seven generated simulation recheck contexts complete after9f5cc2c:
Whole AbstractExample_TTrace.testSpec is translated for Example1/Example2 and
LiveCheckExample1/LiveCheckExample2 generated subclasses in
 tlc_simulation_ttrace_java_test.go: finished/no-GENERAL, temporal/counterexample,
all ten x0..9 state strings/ordinals/action-shape assertions and back-loop1.
Whole SimulationTest2a/LiveCheckSimulationTest2a generated methods preserve
first/last states0/4, ordinals, stuttering absence and repeated STATE_PRINT2 check.
Whole Stuttering generated method preserves stuttering presence/back-loop absence.
Every original generation phase runs its full original simulation assertions;
LiveCheck examples retain their assumeTrue(false) only after full trace checking.
Original helpers now accept a destination override to generate actual artifacts.
Every recheck loads that artifact with source resolver path, -noGenerateSpecTE,
no coverage, inherited debugger/dot/JSON/seed1/workers1/FP0/exit settings.
No fixture changes or weakened assertions. Seven unchanged originals pass -ea
against actual Java-generated files:0.577/0.571/0.576/0.560/0.535/0.519/0.548s;
logs simulation-ttrace-<Class>-original-java.log, compilation
simulation-ttrace-original-javac.log. Go focused21433 passes race37.045s;
log simulation-ttrace-initial-race.log. Related90927 passes race80.048s before
the exploration fix; log simulation-ttrace-related-race.log, not a gate of that fix.
Inventory1149/1269 contexts (90.5%),594/626 classes (94.9%);120pending across32classes.
Generated rechecks39/45 complete; six remain, including one deferred checkpoint
context and five in-scope liveness model rechecks.

Old workspace39038 is terminal failure timeout60m/root3600.055s while running
unchanged TestJavaTLCSetSim for54m5s. SANY1.048s/TLC72.061s pass in that older
snapshot. Log simulation-worker-correctness-final-workspace-go.log. Stack/coverage
show worker continuously regenerating initial states after TLCSet(exit,true).
The ordinary SimulationWorker interruption port missed ExplorationWorker's two
checkForInterrupt calls: the Go exploration path returned nil for a stopped
worker, so the source consumer loop interpreted it as a completed trace and
started another. Replace both checks in simulation_worker_modes.go with source
checkForInterrupt, which throws InterruptedException and terminates/report OK
through SimulateAndReport's existing catch. Preserve original TLCSetSim/MultiSim
flags, depth4224, unbounded traces, coverage/exit assertions; no test weakening.
Focused48871 is terminal exit0, race122.551s; both unchanged TLCSetSim
and TLCSetMultiSim pass. Log simulation-exploration-interrupt-final-race.log.
Final related91538 is terminal exit0; original generation/rechecks, TLCSet
stopping and related liveness assertions pass with the production fix. Log
simulation-ttrace-exploration-final-related-race.log. ok  	github.com/glycerine/tlago	211.320s; ok  	github.com/glycerine/tlago/tlc	1.040s [no tests to run] Final current full49471
is live, log simulation-ttrace-exploration-final-workspace-go.log,
-count=1 -failfast -timeout60m ./.... Resume exact handles.
Old workspace79605 is also terminal failure timeout60m/root3600.065s while
running unchanged TLCSetSim54m12s; SANY1.110s/TLC71.313s pass in that snapshot.
Log simulation-worker-correctness-bounds-final-workspace-go.log. It lacks this
exploration fix and reproduces the same failure. Older84016/59582 were last
polled live and likewise lack this fix; no current pass credited.
Latest completed whole suite remains46404, before queue/simulator/recursion work.

2026-10-05 remaining seven simulation counterexample contexts complete after548e359:
Whole AbstractExampleTestCase.testSpec is translated for Example1Test,
Example2Test, LiveCheckExample1Test and LiveCheckExample2Test in
 tlc_simulation_counterexamples_java_test.go. Preserve constructor depth11,
unbounded trace count, property names Liveness1/Liveness2, stats12, all ten exact
state strings/ordinals/actions, back-loop1/action and source postcondition checks.
Experimental static initializer is preserved. LiveCheck examples execute every
counterexample/trace assertion before source assumeTrue(false), translated as
 t.Skip; no new early skip and no claimed postcondition execution in those cases.
Whole SimulationTest2a/LiveCheckSimulationTest2a methods preserve original
unbounded/depth6 and num100/depth10 constructors, Prop1, first/last states0/4,
ordinals and repeated source STATE_PRINT2 check. Whole StutteringTest preserves
CodePlexBug08/MC, default-depth/unbounded simulation, stuttering presence and
back-loop absence. Retain inherited debugger/seed1/workers1/coverage/dot/JSON/
trace-generation flags and exit assertions in every context.
All nine model input files match original Java bytes; four new Example2/Test2a
vectors, five existing Example1/CodePlex inputs. No production changes or weakened
assertions were needed. Unchanged original Java preflights pass -ea: Example1
0.441s (prepared log),Example2 0.442s,LiveCheck examples0.552s/0.540s,
SimulationTest2a0.402s,LiveCheckSimulationTest2a0.480s,Stuttering0.414s;
logs simulation-remaining-<Class>-original-java.log (Example1 prepared variant).
Initial Example1 scratch directory failure is not a behavioral parity failure:
original build target/GeneratedTESpecs must exist; unchanged rerun passes.
Focused Go race76633 terminal exit0,7.495s; log
simulation-counterexamples-initial-race.log, with five passes/two original
assumption skips after full trace assertions. Related original simulator,
successful simulation, Example1 dump/load and alias-liveness checks60295 pass
root race50.127s; TLC1.042s explicitly has no matching tests. Log
simulation-counterexamples-related-race.log. Fixtures/source assertions reviewed.
Inventory1142/1269 contexts (90.0%),587/626 classes (93.8%);127pending across39classes.
Simulation55/55 methods,20/20 classes complete. Generated TTrace variants are
separate pending contexts; next seven simulation-generated rechecks, retaining
full original generation settings and recheck assertions.
Full workspace84016 (73419a7),59582 (c0c789b),79605 (3d715c3) and39038
(pre-final queue bounds) remain live after this turn's polls. These precede eleven
new test-only contexts; current focused/related gates cover the additional tests.
No terminal current full-suite success is claimed. Resume exact handles.

2026-10-05 four successful/assumption simulation contexts complete after73419a7:
Whole original SimulationTest2, LiveCheckSimulationTest2,
SimulationTest2PostCondition and SimulationTestAssumption testSpec methods are
translated in tlc_simulation_success_java_test.go. Preserve source depth6/10,
trace limits50/1, seed1, worker1, inherited debugger/coverage/dot/JSON/trace-spec
flags and exit assertions. Experimental mode follows the original static
initializer. SuccessfulSimulationTestCase's whole method, including repeated
STATE_PRINT2 absence check, is retained in all three inheriting contexts;
postcondition adds its own assertion. Assumption overrides without calling super.
All five input files match Java bytes; three new vectors, two existing Test2 files.
No production changes or weakened assertions were needed. Four unchanged Java
originals pass -ea:0.419s/0.543s/0.357s/0.332s, logs
simulation-success-<Class>-original-java.log; compilation log
simulation-success-original-javac.log. Final focused Go race passes5.582s,
89078 terminal exit0, log simulation-success-models-initial-race.log.
Inventory1135/1269 contexts (89.4%),580/626 classes (92.7%);134pending across46classes.
Simulation48/55 methods,13/20 classes complete; remaining seven original models
and their generated TTrace variants. Broader original Simulation/LiveCheckSimulation/Simulator selection
65602 is terminal exit0; log simulation-success-models-related-race.log.
Full workspace84016 (73419a7),59582 (c0c789b),79605 (3d715c3) and39038
(pre-final queue bounds) were polled and remain live; no current whole-suite pass
claimed. New test-only changes are covered by focused/broader gates.
Next seven unchanged original Java preflights are compiled and run in
correctness-java/simulation-success-run; logs simulation-remaining-<Class>-
original-java.log. Example1's initial postcondition failure is scratch setup:
missing target/GeneratedTESpecs directory. After creating the original build
output directory, unchanged Example1 passes (prepared log). Six others pass.
LiveCheck examples retain original assumeTrue(false) after full trace assertions;
that conditional applies only to postcondition checks, not earlier assertions.
Do not weaken tests or credit these unported methods from Java preflight alone.

2026-10-05 four original simulation models complete afterc0c789b:
Whole testSpec methods for simulation.Github1191Test, Github1191aTest,
Github602Test and NQSpecTest are translated in tlc_simulation_models_java_test.go.
Preserve all constructors, inherited ModelCheckerTestCase flags/exit assertions,
finished/no-GENERAL checks, Github1191a's zero-uncovered check and Github602's
exact first progress parameters156/1/100/0/0. NQSpec retains its inherited existing
debugger flag and all100traces; no new debugger/scoped-identifier work is done.
All8 input files match Java bytes;7 are new vectors, Github602 already exists.

NQSpec initially fails E4290 because Go rejects any prime in a recursive body.
ModuleNode instead checks each recursive formal's propagated maximum against
ActionLevel. Primed free variables are allowed. Both module and LET checks now
use that rule and exact Java diagnostic/range. Recursive signatures initialize
original formals to ActionLevel and weight1; INSTANCE-prepended and selected bound
parameters retain their own rules. Preserve monotone dependency/constraint
propagation and per-iteration caches. This addresses formal argument priming and
initialization; it does not claim complete SANY higher-order/recursive-section
level constraint graphs. An existing Go negative case keeps its unchanged input
and both rejection assertions; its obsolete generic substring is strengthened
to Java's full Argument1-of-recursive-operator-op-is-primed diagnostic.

Unchanged originals pass Java with -ea:1191 0.378s,1191a0.364s,6020.366s,NQ3.141s;
logs simulation-models-<Class>Test-original-java.log. Initial Go22109 is terminal
failure on NQ's blanket rule, log simulation-models-initial-race.log. Final all4
methods pass race187.391s (39534 terminal), NQ183.72s, log
simulation-models-recursive-final-race.log. All original assertions/settings
retained. Full SANY Java passes race9.032s (89421 terminal). Related legacy201–220,
99/999/InvalidInvariant/ETest plus parser/semantic/XML/elevated-warning selections
pass race101.892s (64110 terminal), log
simulation-models-recursive-final-related-java-message-race.log. Earlier related
84635/44257 fail the obsolete native message assertion and are retired failures,
not credited passes. Source assertions and fixture bytes were reviewed.
Inventory1131/1269 contexts (89.1%),576/626 classes (92.0%);138pending across50classes.
Simulation44/55 methods,9/20 classes complete; next11 liveness/simulation models
and their generated TTrace variants, retaining every inherited assertion.

Trace workspace82090/46404 are terminal exit0: root1839.855s/1836.917s,
SANY1.169s/1.135s,TLC68.787s/68.899s. Latest completed full snapshot46404 precedes
result-queue/simulator/recursive changes. Queue39038/79605 and simulator59582
snapshots remain live. Current final workspace84016 is live, log
simulation-models-recursive-final-workspace-go.log, -count=1 -failfast -timeout60m
./.... Resume exact handles; retain snapshot scopes and do not credit current
full success before terminal exit0. This supersedes earlier live notes.

2026-10-05 original simulator correctness classes complete after3d715c3:
Whole tool.SimulatorTest.testPrintBehaviorShouldPrintErrorState is translated in
 tlc_simulator_error_state_java_test.go: Github726 FastTool inputs, seed/aril0,
traceDepth-1/traceNum0, explicit0workers, empty state, uncoded TestException and
original TLC_ERROR_STATE assertion. Existing two vectors match source bytes.
Port the uncoded TLCRuntimeException constructor, explicit simulator worker count
and exception-printing overload. That overload reports no summary; normal worker
runtime-exception reporting now uses it. Remove fabricated fallback workers and
empty-trace suppression, preserving the source consumer loop and trace dispatch.

All ten simulation.SimulatorTest methods and all ten inherited
SimulatorMultiThreadTest contexts are translated in
 tlc_simulator_correctness_java_test.go. Each context preserves all22 source
assertions, numWorkers1/4, seed0, FP64 setup, exact configuration names, depth100,
traceNum100 or MAX_INT64 and continuation settings. TestMPRecorder's string-value
check examines parameter0 of the first record, as in Java. No substitution with
later records, reduced worker counts, finite limits or weakened checks.

Initial Go fails original testInvariantBadEvalInitState by returning its exception
after reporting it. Production now preserves the full initial-state try/catch:
generation, initial validity/invariants and model constraints catch Exception,
print the correct current-state/general diagnostic and summary, and return its
code. Error subclasses escape. Preserve the zero generated-state assertion; the
no-init/constraint checks remain outside that catch. Assertions are unchanged.

Unchanged Java Simulator.java and tool.SimulatorTest pass1method with -ea,0.247s,
log simulator-error-state-original-java.log. Original simulation.SimulatorTest
and SimulatorMultiThreadTest each pass10methods,0.432s/0.506s, logs
simulator-correctness-SimulatorTest-original-java.log and
simulator-correctness-SimulatorMultiThreadTest-original-java.log. Initial Go
failure47825 is retired, log simulator-correctness-initial-race.log. First return
fix49926 passes20contexts race3.054s. Final all simulator/worker originals plus
existing simulator/vector checks pass root race5.389s/TLCrace1.145s (84521 terminal),
log simulator-correctness-final-related-race.log. All45 new source assertions and
complete method/context settings were checked. Inventory1127/1269 contexts
(88.8%),572/626 classes (91.4%);142pending across54classes. Simulation40/55 methods,
5/20 classes complete; next its15 original model contexts and their TTrace forms.

Current final whole workspace59582 is live, log
simulator-correctness-final-workspace-go.log (-count=1 -failfast -timeout60m ./...).
Older trace82090/46404 and result-queue39038/79605 full snapshots remain live;
resume those handles and retain their documented snapshot scope. Latest completed
full workspace is still Test21731625, before trace/result-queue/simulator changes.
Do not claim current full-suite success before a terminal exit0. This entry
supersedes earlier live notes. Deferred topics and pending upstream-failing
assertions remain on the inventory; this does not establish goal completion.

2026-10-05 original simulation worker correctness class complete afterb8dc445:
All twelve whole original tlc2.tool.simulation.SimulationWorkerTest methods are
translated in tlc_simulation_worker_correctness_java_test.go. Preserve126 source
assertions, constructors/seed0, exact branch/depth values, all trace levels/error
codes, queue emptiness, join/isAlive, interruption, depth1 restriction and the
70generated states/5traces/packed mean5 counters. All13 BasicMultiTrace .tla/.cfg
files match source bytes. Setup retains the isolated source statics and user
directory, FastTool simulation mode before state evaluation and NoOpLiveCheck.
Cleanup joins owned workers before restoring globals; it does not affect asserts.

Production Simulator and SimulationWorker no longer use a small bounded result
channel. Port LinkedBlockingQueue FIFO with blocking take, default MAX_INT32
capacity, interruptible producer waits, nonblocking offer and interruption wakeup.
This allows original tests to join workers while unread error results remain.
Source InterruptedException termination uses offer(OK) without counting an
interrupted trace as completed; exception reports use offer(error). The worker
loop now terminates through simulateAndReport, preserving source check points
and final result behavior. IsAlive observes the existing completion channel.

Unchanged original Java class passes12methods with -ea,0.444s, log
simulation-worker-correctness-original-java.log. Initial faithful Go methods pass
race2.940s (71558 terminal). Final all19 original worker/trace methods plus
existing simulator/vector checks pass root race3.116s/TLCrace1.115s (18824 terminal),
log simulation-worker-correctness-final-bounds-related-race.log. All126 assertions
were counted per method and reviewed against source, not replaced by subset
checks. Inventory1106/1269 contexts (87.2%),569/626 classes (90.9%);163pending across
57classes. Simulation19/55 methods,2/20 classes complete.

Test217 full workspace31625 is terminal exit0: root1815.949s,SANY1.049s,
TLC69.240s, log legacy-error-test217-final-workspace-go.log. Broad TLC race73112 is
terminal green638.479s, snapshot before final trace copy adapters/assertions/null
normalization. Trace full82090 (before null normalization) and46404 (final trace,
before result queue) remain live. Result queue workspace39038 (before final
capacity/wakeup details) is live, log simulation-worker-correctness-final-workspace-go.log.
Final current-code full79605 is live, log
simulation-worker-correctness-bounds-final-workspace-go.log. Each uses -count=1
-failfast -timeout60m ./.... Resume exact handles; record snapshot scope and wait
for terminal exit0 before crediting full success. This supersedes older live notes.
Next correctness work: original tool.SimulatorTest error-state printing, then
simulation.SimulatorTest/SimulatorMultiThreadTest and the simulation models.

2026-10-05 original SimulationWorker trace helper class complete after2c0c149:
All seven whole original tlc2.tool.SimulationWorkerTest methods and its equality
dummy are translated in tlc/simulation_worker_trace_java_test.go. Preserve all37
assertions, constructor defaults, exact fingerprints/IDs, original state chains
and predecessor identity after finite stuttering removal. The Go trace path
previously accepted only TLCStateMut, which could not represent the original
state's overridden equality/fingerprint/predecessor behavior. StateVec now stores
TLCState and provides polymorphic ElementAt alongside existing typed access;
SimulationWorker traverses TLCPredecessorState and corrects predecessors through
dynamic dispatch. Native state metadata retains generic predecessors. Preserve
source synchronization on compressed/uncompressed traces and compressed trace's
initial/level invariants. Copy adapters retain covariant mutable-state return
methods through the collection surface. This credits the trace helper class,
not the separate twelve-method simulation.SimulationWorkerTest.

Unchanged Java SimulationWorker.java and both original test/support sources were
compiled against frozen dependencies; JUnit passes7methods with -ea,0.021s,
log simulation-worker-trace-original-java.log. Final seven methods plus existing
state/vector/queue/simulator checks pass race1.127s (4767 terminal), log
simulation-worker-trace-final-related-race.log. Original assertion counts per
method match1/4/4/4/7/10/7, and their expressions/operands are retained.
Inventory1094/1269 contexts (86.2%),568/626 classes (90.7%);175pending across
58classes. Simulation7/55 methods,1/20 classes complete.

Test215 full workspace58038 is terminal exit0: root1780.992s,SANY0.982s,
TLC70.398s, log legacy-error-test215-final-workspace-go.log. Test217 full31625
remains live. Trace workspace82090 is live, compiled before typed-nil normalization, log
simulation-worker-trace-final-workspace-go.log, using -count=1 -failfast
-timeout60m ./.... Broad TLC race73112 is live, log
simulation-worker-trace-final-tlc-race.log; it compiled before the final copy
adapters and source assertions, so preserve that snapshot scope. Resume each
handle, do not infer terminal status or full-suite success from silent logs.
Typed-nil metadata predecessors are normalized to Java null at trace dispatch;
the final focused gate includes this fix. Final whole workspace46404 is live,
log simulation-worker-trace-null-final-workspace-go.log, same -count=1 gate.
This entry supersedes earlier live notes. Next correctness work: the separate
simulation.SimulationWorkerTest's twelve original methods and BasicMultiTrace
vectors; retain all seeds, generated trace values, worker completion and counts.

2026-10-05 original Test217 complete after9b0d44e:
Whole original Test217.testSpec preserves ERROR_SPEC_PARSE, no GENERAL, all three
exact TestPrintStream substrings, inherited SuiteETestCase flags and three
byte-identical vectors. Go initially reached the configuration error instead.
Production now propagates minimum argument maximum levels through operator
signatures and INSTANCE formal arguments. Preserve Java builtin bounds, bounded
expression domains, CHOOSE/set comprehension bounds, EXCEPT components and @'s
base/prior-component dependencies; retain per-iteration definition/expression
caches. Report the whole application range and exact Java diagnostic. This is
formal maximum-level propagation, not a claim that all higher-order SANY level
constraint graphs are complete.

Focused Test217 passes race1.269s (85988 terminal). Related original legacy and
ETest1–16 plus root parser/semantic/XML/elevated-warning checks pass race98.308s
(91961 terminal), log legacy-error-test217-final-related-race.log. Full SANY Java
passes race8.775s (55732 terminal), log legacy-error-test217-final-sany-race.log.
All three source assertion strings and three fixture files match Java exactly.
Inventory1087/1269 contexts (85.7%),567/626 classes (90.6%);182pending across
59classes. All104 numbered legacy contexts are translated.

Disk recovery on committed9b0d44e is terminal green: both original EWD840FL1 and
BufferedRandomAccessFile pass normal138.167s (87959), log
workspace-disk-recovery-final-code-go.log. Earlier recovery96031 and related
8809/35777 snapshots timed out before final per-iteration caches; they are retired
and superseded by the final gates, not counted as passes. Earlier full213/214
ENOSPC receipts remain failures. Current full21558038 and full21731625 are live,
logs legacy-error-test215-final-workspace-go.log and
legacy-error-test217-final-workspace-go.log. Resume these handles; do not claim
full-suite success until terminal exit0. Full runs use -count=1 to avoid storing
large successful output cache blobs. This entry supersedes earlier live notes.

2026-10-05 original Test215 complete after50e713f:
Whole original Test215.testSpec preserves ERROR_SPEC_PARSE, no GENERAL, all nine
exact TestPrintStream substrings, inherited SuiteETestCase flags and two
byte-identical vectors. Initial Go reported eight errors with point ranges and
generic messages. Report Java's whole temporal-expression range/message and
individual quantified-bound range/message. Level composition now evaluates
actual definitions, arguments, enclosing declarations and lexical LET bindings;
remove the call-argument-dropping logicalOperandLevel shortcut. Preserve
OpApplNode's node-kind conditions for []/<> action wrappers.

Dependency checking now retains Java OpDefNode/OpApplNode's levelChecked guards:
check definitions and expressions once per fixed-point iteration. Do not mutate
cached argument nonLeibnizParams when a parent restricts its allParams.
Focused Test212/213/214/215/99 pass race6.585s (61277 terminal); Test99 drops from
38.38s with definition-only caching to4.28s with both source caches, preserving
all original assertions. Final related original legacy201–215/216/219/220/99/999/
InvalidInvariant and ETest1–16 plus root parser/semantic/XML/elevated-warning
selections pass race101.616s (12610 terminal), log
.codex-gotmp/legacy-error-test215-expression-cache-related-race.log. Full SANY Java
race8.504s (6787 terminal), log legacy-error-test215-expression-cache-sany-race.log.
All nine assertion strings and fixture bytes were independently compared.
Inventory1086/1269 contexts (85.6%),566/626 classes (90.4%);183pending across
60classes; legacy103/104. Next original Test217 after finishing disk recovery.

Test212 full workspace82923 is terminal exit0: root1776.964s,
SANY0.935s/TLC70.342s, log legacy-error-test212-final-workspace-go.log. Later
full21343019 and full21490130 are terminal exit1 due to ENOSPC, not credited green:
root873.473s/365.299s; each reports explicit no-space-left write failures in
BufferedRandomAccessFile's nodes_0 or EWD840MC1_liveness.dot and Go's testlog.txt.
Removed only three stale private Go cache blobs, freeing21.2GiB; vectors intact.
The two affected original contexts are running race96031, log
.codex-gotmp/workspace-disk-full-recovery-race.log. EWD840FL1 passes490.04s;
BufferedRandomAccessFile remains live. This process compiled before per-iteration
cache additions. Resume the same handle; do not weaken tests or claim it passed.

Current full workspace58038 is live, log legacy-error-test215-final-workspace-go.log.
Use -count=1 to avoid retaining large successful test-output cache blobs, with
-failfast -timeout60m ./... otherwise unchanged. Older related race8809 (before
level caches) and35777 (definition-only cache) are still live; their logs are
legacy-error-test215-final-related-race.log and
legacy-error-test215-level-cache-related-race.log. Preserve their eventual
receipts' snapshot scopes. Current focused/related/SANY gates above are terminal.
This entry is authoritative; earlier live/full-suite notes are historical.

2026-10-05 original Test214 complete after6a37709:
Whole original Test214.testSpec preserves ERROR_SPEC_PARSE, no GENERAL, both
exact TestPrintStream substrings, inherited SuiteETestCase flags and two
byte-identical vectors. Initial Go erroneously completed model checking.
UseOrHideNode.factCheck restricts HIDE facts by semantic node kind; it does not
restrict USE expressions. The old Go check exempted builtins and omitted other
fact forms. ProofRef now retains fact expressions; numbered USE/HIDE steps retain
their facts and participate in proof scope. Resolve selectors in fact mode and
recognize imported theorem/assumption names and earlier scoped steps. Preserve
Java's OpAppl-only restriction, including its separate number/string, LET,
label and ASSUME/PROVE node behavior. Report the exact fact range/message.

Focused original Test214 passes race1.152s (90435 terminal). Related original
legacy201–214/216/219/220/99/999/InvalidInvariant and ETest1–16 plus root parser,
semantic/XML/elevated-warning selections pass race99.522s (45817 terminal), log
.codex-gotmp/legacy-error-test214-final-related-race.log. Full SANY Java race
passes9.731s (46442 terminal), log legacy-error-test214-final-sany-race.log.
Both source assertions and vector bytes were independently compared.
Inventory1085/1269 contexts (85.5%),565/626 classes (90.3%);184pending across
61classes; legacy102/104. Next original Test215, then217.

Current Test214 full workspace90130 is live, log
.codex-gotmp/legacy-error-test214-final-workspace-go.log. Older full21282923 and
full21343019 remain live, logs legacy-error-test212-final-workspace-go.log and
legacy-error-test213-final-workspace-go.log. Resume these exact handles and
record each snapshot's scope; no full-suite success claim before terminal exit0.
This is authoritative; preceding entries record historical checkpoints.

2026-10-05 original Test213 complete after31106ad:
Whole Test213.testSpec preserves ERROR_SPEC_PARSE, no GENERAL, all six exact
TestPrintStream substrings, full SuiteETestCase settings and three byte-identical
vectors. Go initially reached the missing-init configuration error, then reported
only three errors. The real semantic implementation now yields Java's five errors.

TheoremNode.LevelCheckTemporal's traversal preserves a temporal goal through
CASE/QED subproofs; ordinary assertions introduce their own goals. ProofStep
retains the statement range separately from its numbered-step position, producing
Java's exact CASE/HAVE/WITNESS ranges and messages. Proof levels now resolve actual
operator definitions and weighted arguments instead of assuming named operators
are constant-level. The dependency analyzer separately tracks levelParams and
intrinsic levels, with monotone recursive summaries and known NEW symbol levels.
LevelNode.addTemporalLevelConstraintToConstants constrains ConstantDecl symbols
contributing to ASSUME or ASSUME/PROVE levelParams to ActionLevel. Carry these
constraints through definitions, LAMBDA, extended modules and module INSTANCE
substitutions. Check explicit and implicit substitutions, including operator
bodies, even for constant modules. Preserve Java's whole-INSTANCE range/message.
This ports the temporal-constant constraint rule, not all SANY level-constraint
or native proof-graph machinery.

Test212/213 first pass race1.283s; intrinsic-level refinement passes race1.333s
(59205/11020 terminal). Final native proof-level changes pass the related original
legacy201–213/216/219/220/99/999/InvalidInvariant and ETest1–16 plus root parser,
semantic/XML/elevated-warning selections: race92.485s (94504 terminal), log
.codex-gotmp/legacy-error-test213-proof-levels-related-race.log. Complete SANY Java
race9.299s (39707 terminal), log legacy-error-test213-proof-levels-sany-race.log.
Earlier final constraints gate race98.769s (6337 terminal) and SANY8.221s
(24270 terminal) preceded the last native proof-level refinement.
Exact Java assertion strings and all fixture bytes were independently compared.
Inventory1084/1269 contexts (85.4%),564/626 classes (90.1%);185pending across
62classes; legacy101/104. Next original Test214, then215/217.

Full Test212 workspace82923 remains live and predates Test213 changes, log
.codex-gotmp/legacy-error-test212-final-workspace-go.log. Full current Test213
workspace43019 remains live, log legacy-error-test213-final-workspace-go.log.
Resume these exact handles; no full-suite success claim until terminal receipts.
Preserve the five deferred topics and test_vectors naming.

2026-10-05 Test213 implementation in progress after07b206a:
Whole original Test213.testSpec is translated locally with its constructor,
no-GENERAL assertion, six exact substrings and byte-identical source vectors.
It remains uncredited and uncommitted. Initial Go run wrongly reached TLC's
missing-init configuration error. Temporal proof validation now follows Java
TheoremNode.LevelCheckTemporal: CASE/QED subproofs inherit the enclosing goal,
other assertions start their own goal. ProofStep retains the statement's syntax
range separately from the numbered step. Go now produces all three exact Java
CASE/HAVE/WITNESS errors, but still misses the two INSTANCE level constraints.
The unchanged original test is red (three errors instead of five); do not
advance to Test214 or weaken it.

Next implementation: LevelNode.addTemporalLevelConstraintToConstants limits
ConstantDecl symbols contributing to ASSUME or ASSUME/PROVE levelParams to
ActionLevel. InstanceNode checks the resulting module levelConstraints even
for a constant module. Track levelParams through builtin argument weights,
definitions, shadowing and INSTANCE substitutions; do not approximate this by
scanning identifier occurrences or reuse XML export as a semantic roundtrip.
Java sources: semantic/{LevelNode,AssumeNode,AssumeProveNode,InstanceNode}.java.
Root related parser/semantic/XML and original208/210/212 pass race9.385s,
log .codex-gotmp/legacy-error-test213-temporal-related-race.log; the sany_tests
line in that log has no matching tests and gives no SANY suite credit.
Separate full SANY Java race38683 is terminal exit0 (8.718s), log
.codex-gotmp/legacy-error-test213-temporal-sany-race.log.

Test210 whole workspace15753 is terminal exit0: root1762.590s,
SANY0.909s/TLC69.724s, log legacy-error-test210-final-workspace-go.log.
This covers2303cdb production before Test212 and current Test213 changes.
Test212 whole workspace82923 remains live; resume its handle, not a new run.
Current inventory stays1083/1269 contexts and563/626 classes.

2026-10-05 original Test212 complete:
The original ERROR_SPEC_PARSE constructor, no-GENERAL assertion, seven exact
TestPrintStream substrings, inherited SuiteETestCase settings and four
byte-identical vectors are translated without weakened assertions. Replaced
INSTANCE's prime-anywhere shortcut with separate all/non-Leibniz formal
dependencies. Propagate builtin argument weights, imported definitions, LAMBDA,
INSTANCE substitutions and higher-order arguments; recursive definition summaries
grow to a fixed point. Preserve Java LetInNode's dependency propagation rule.
Diagnostics use the whole INSTANCE range and original Java message.

Focused Test212 race passes1.177s; related original legacy contexts plus existing
parser/semantic/XML/warning selections pass race110.981s (63865 terminal).
Explicit root parser/SANY semantic/XML/elevated-warning race passes4.239s
(43677 terminal), complete SANY Java race8.340s (52980 terminal). A temporary
recursive dependency probe passes race1.029s and was removed, not credited.
Logs: .codex-gotmp/legacy-error-test212-{fixedpoint,related,root,sany}-race.log.
Inventory1083/1269 contexts,563/626 classes;186pending across63classes;
legacy100/104. Next original Test213, then214/215/217.

Earlier ETest16 whole workspace75902 is terminal exit0: root1764.043s,
SANY0.930s/TLC69.463s (7f51f9c production, before Test210/Test212).
Test210 full workspace15753 remains live, log
.codex-gotmp/legacy-error-test210-final-workspace-go.log; resume that handle.
Current Test212 full workspace82923 is live, log
.codex-gotmp/legacy-error-test212-final-workspace-go.log; resume that handle.
Do not claim current full-suite success before its terminal receipt.

2026-10-05 verified original Test210 diagnostic checkpoint afteread7f9a:
Original Test210.testSpec is mechanically translated with ERROR_SPEC_PARSE,
no GENERAL, all ten exact TestPrintStream substrings (header/count9 plus nine
full location/message blocks), original SuiteETestCase flags and two byte-identical
vectors. Initial Go output had two errors instead of nine. Selector resolution
used ASSUME/PROVE's quantified expression projection and did not walk proof steps;
it lost NEW clause indices/domains, lexical declaration scope and SUFFICES handling.
Proof steps now retain their qualified names, ASSUME/PROVE bodies and SUFFICES
status. Resolve actual clause structure and nested NEW domains, track the enclosing
proof versus subsequent statements, follow Java's illegalLabelRef/illegalAPPosRef
and SUFFICES positional wrapper rules, and retain individual selector ranges.
Named assumption labels are legal. Nested proof ASSUME/PROVE labels are validated
with Java's full range/message. BY facts retain fact/expression mode, including
leaf-only proofs without numbered steps; NEW fact selections retain their source
NewSymbol rather than evaluating a projected domain. This is a selector/diagnostic
port, not a declaration of complete native proof-graph or all Generator parity.

Whole original Test210 first passes normal0.040s; focused race1.159s (38423
terminal/retired) before the final explicit domain/leaf traversal additions.
Related original legacy contexts (201–210/216/219/220/99/999/InvalidInvariant and
ETest1–16) pass race98.963s (73617 terminal/retired) before the final leaf-only
proof retention. Root parser/semantic/XML/elevated-warning race14.474s (43387
terminal/retired), SANY Java race8.226s (47389 terminal/retired).
Final leaf-retention code: complete SANY Java race7.901s (85424 terminal/retired),
log legacy-error-test210-final-sany-race.log. Final combined related root/legacy
race gate passes114.107s (42109 terminal/retired), log
legacy-error-test210-final-related-race.log. Keyword directives now use token
kinds (BY/DF/SUFFICES) rather than images, preserving keyword record fields.
That refinement passes root/Test208/Test210 race21.221s and full SANY Java
race8.109s (83989/68691 terminal/retired). Final
ASSUME/PROVE selections retain structured targets in fact mode; Java rejects a
whole AP as an expression rather than evaluating its quantified projection.
Final guard code passes related root/Test208/Test210 race20.627s (30971
terminal/retired) and full SANY Java race8.121s (90316 terminal/retired), logs
legacy-error-test210-ap-final-race.log and legacy-error-test210-ap-sany-race.log.
All logs under .codex-gotmp/. Every original assertion remains unchanged.
Inventory1082/1269 contexts (85.3%),562/626 complete classes (89.8%),187pending
across64classes; legacy99/104 complete. Next original Test212, then213/214/215/217.
Test212 requires six exact non-Leibniz substitution errors at whole INSTANCE
ranges and its two dependency modules; fix real Leibniz analysis before advancing.

Corrected INSTANCE full workspace57071 passes exit0/terminal, root1772.759s,
SANY0.932s/TLC72.721s, log legacy-error-context-final-workspace-go.log. This covers
caf0ad3 production code and ETest1–3, before ETest5/16 semantic report/validation
changes and before Test210. Full diagnostic snapshot23927 finished exit1/retired,
root1791.100s/SANY1.083s/TLC70.782s; it compiled BEFORE caf0ad3 and failed only the
three now-fixed INSTANCE-order cases. Do not credit that older snapshot as green.
Current ETest16 whole workspace75902 remains live, log
legacy-error-etest16-final-workspace-go.log. It includes7f51f9c and ETest1–16,
but not the current Test210 semantic changes. Resume this exact handle, never
restart on an observation timeout, and preserve its eventual receipt's scope.
Final current-production whole workspace launched after commit2303cdb: live15753,
log legacy-error-test210-final-workspace-go.log under .codex-gotmp/. It includes
all current Test210 selector/proof semantic changes and complete original ports.
Resume this exact handle. Do not claim its full-suite result before terminal success.
Preserve the five deferred topics and test_vectors fixture naming.
This entry is authoritative; older counts/gate/next-context notes are historical.

2026-10-05 ETest8–16 complete after6e5d315:
All16 original SuiteETestCase contexts are now complete translations. This batch
preserves the nine original constructors, all inherited runner flags, every source
assertion and18 byte-identical vectors. ETest8 uses -simulate, ERROR (the upstream
simulator TODO), checkDeadLock=false, FINISHED/STATS_SIMU, no DEADLOCK and its
exact zero-count uncovered location. ETest9–11 retain FAILURE_SPEC_EVAL, first
GENERAL record membership-error substrings and exact zero coverage. ETest12–15
retain VIOLATION_ASSUMPTION and exact first assumption-evaluation error substrings
for Cardinality overflow, unsuccessful CHOOSE and invalid integer comparison.
ETest16 initially reported one semantic error instead of Java's two: duplicate
fields in sets of records were unchecked. Generator.processRcdForms:4094 handles
both record forms. Go now validates both, reports the repeated field token range
and Java's Non-unique fields message via the existing SANY diagnostic metadata.
The source assertions and vectors are unchanged. No invented tests were added.

Final all ETest1–16 pass race40.497s (82330 terminal/retired), log
legacy-error-etest1-etest16-race.log. Individual8 race1.808s,9–11 race14.175s,
12–15 race18.217s,16 race1.132s, handles97738/43233/80532/92787 terminal/retired.
Related root parser/semantic/XML/elevated-warning race14.895s (53867 terminal/
retired), log legacy-error-record-constructors-related-race.log. Complete SANY
Java race8.013s (15921 terminal/retired), log
legacy-error-record-constructors-sany-race.log.
The three previously failing workspace contexts plus ETest1–3 also pass
race720.306s (15192 terminal/retired), log workspace-instance-context-fix-race.log.
This validates caf0ad3's actual INSTANCE filtering correction with unchanged
EvalExceptionLiveness, EWD998ChanDebugger and BufferedRandomAccessFile assertions.
Inventory1081/1269 contexts (85.2%),561/626 complete classes (89.6%),188pending
across65classes; legacy98/104 complete. Next original Test210, then212/213/214/
215/217. Test210 requires exactly nine semantic errors, exact ranges and full
ASSUME/PROVE scope, nested-label and SUFFICES diagnostics. Fix the semantic
implementation before proceeding after any failure; no generic parse-failure
substitute. The five deferred topics remain deferred.

Only two earlier whole-workspace handles remain confirmed live:23927 (diagnostic
snapshot BEFORE caf0ad3's INSTANCE context filter fix) and57071 (caf0ad3 production
code BEFORE f8bd7c9's unknown-operator metadata and the current duplicate-record-set
validation). Resume these handles; do not restart on observation timeouts. Logs
legacy-error-reporting-final-workspace-go.log and
legacy-error-context-final-workspace-go.log under .codex-gotmp/.
Final current-source full workspace launched after commit7f51f9c: live75902,
log legacy-error-etest16-final-workspace-go.log under .codex-gotmp/. It includes
the duplicate-record-set validation and all16 ETests. Resume this exact handle
and preserve earlier snapshot receipts with their accurate scope.
Do not claim full current-source green from focused evidence alone.
This entry is authoritative; earlier inventory and live-handle notes are historical.

2026-10-05 original ETest6 and ETest7 complete after f8bd7c9:
Original ETest6 passes FAILURE_SPEC_EVAL, the first GENERAL record's exact
undefined-identifier substring with line16:23 location, and both exact zero-count
uncovered locations. Its first-record substring search follows TestMPRecorder.
Original ETest7 passes ERROR_CONFIG_PARSE and exact
TLC_CONFIG_SUBSTITUTION_NON_CONSTANT parameters C/Foo. Preserve full inherited
SuiteETestCase runner flags and original constructors. All four vectors are
source-identical; neither test required an implementation change.
ETest6 race5.864s (80093 terminal/retired), ETest7 race1.764s (31611 terminal/
retired); logs legacy-error-etest6-race.log and legacy-error-etest7-race.log.
Inventory1072/1269 contexts (84.5%),552/626 complete classes (88.2%),197pending
across74classes; legacy89/104 complete. Next original context ETest8. Authoritative
source overrides earlier informal notes: its constructor uses -simulate and ERROR
(the upstream TODO notes the simulator's exit status), checkDeadLock=false;
assert FINISHED/STATS_SIMU, no DEADLOCK and exact uncovered18:15–18:22.
No substitute model, status correction or extra deadlock checking is allowed.

Last confirmed live handles:15192 (three corrected workspace failures + ETest1–3
race),23927 (older diagnostic full workspace BEFORE context-filter fix),57071
(final caf0ad3 production full workspace). Resume these, do not restart on an
observation timeout. Logs workspace-instance-context-fix-race.log,
legacy-error-reporting-final-workspace-go.log and
legacy-error-context-final-workspace-go.log under .codex-gotmp/.
The corrected three-failure normal run passes69.425s; related original legacy
race64.544s; ETest1–5 and root/SANY diagnostic race gates pass as recorded below.
Full workspace verification remains pending; no full-current green claim.
The earlier bc01728 workspace83892 failed three cases, fixed in caf0ad3, and is
terminal/retired. No changes to the five deferred topics' scope or fixture naming.
This is the authoritative next-context checkpoint; older entries are historical.

2026-10-05 original ETest5 complete after ce93212:
Original ETest5.testSpec first rejected M!Init correctly but printed a point range
and Go's generic undefined-identifier text. Generator.selectorToNode:810 reports
the complete unresolved compound-name range and Unknown operator message.
The semantic diagnostic now carries those SANY details separately, excluding
supplied argument ranges. Generic diagnostic positions and text remain available
to other frontends. No fixture-dependent reporting or assertion weakening.
ETest5 preserves ERROR_SPEC_PARSE, no GENERAL, one error, range13:15–13:20 and
exact Unknown operator text. Both fixtures match original source byte for byte.
ETest1–5 pass race11.578s (56875 terminal/retired), related root parser/semantic/
XML/elevated-warning race14.724s (89702 terminal/retired), complete SANY Java
race8.418s (75558 terminal/retired). Logs: legacy-error-etest1-etest5-race.log,
legacy-error-unknown-operator-related-race.log,
legacy-error-unknown-operator-sany-race.log under .codex-gotmp/.
Inventory1070/1269 contexts (84.3%),550/626 complete classes (87.9%),199pending
across76classes; legacy87/104 complete. Next original context ETest6, preserve its
undefined primed-variable GENERAL substring and both exact uncovered locations.
Workspace57071 is still the caf0ad3 production gate; the later ETest5 change adds
unknown-identifier SANY report metadata, covered by the focused root/SANY gates.
Older23927 compiled before the context filter correction. The three original
workspace failures have a normal PASS69.425s; their additional race gate15192
remains live. Resume existing handles; do not restart on observation timeouts.
Do not claim a full current-source green workspace while verification is pending.

2026-10-05 original ETest4 complete after caf0ad3:
Whole original ETest4.testSpec passes with FAILURE_SPEC_EVAL, FINISHED, stats0/0/0,
the exact four-frame TLC_NESTED_EXPRESSION string including its final blank line,
and the exact zero-count uncovered location. Both vectors are source-identical.
The initial port passed a literal ': 0' suffix to requireJavaTLCUncovered, whose
API already filters zero-count records and compares locations. Corrected that
mechanical translation without changing coverage semantics or implementation.
ETest1–4 pass race12.043s (82804 terminal/retired), log
legacy-error-etest1-etest4-race.log. Inventory1069/1269 contexts (84.2%),549/626
complete classes (87.7%),200pending/77classes; legacy86/104 complete.
Next context ETest5: preserve one semantic error, range13:15–13:20 and exact
Unknown operator M!Init diagnostic; fix actual semantic behavior if it fails.
Workspace57071 remains the final caf0ad3 production gate; it compiled before
ETest4 was added, so combine its eventual receipt with this focused race evidence.
Older diagnostic gate23927 compiled before caf0ad3's context filter correction.
Do not claim a full green workspace while its authoritative final gate is live.
The five deferred topics remain deferred. All process/log details below retain
snapshot scope; logs are under .codex-gotmp/.

2026-10-05 ETest1–3 and semantic diagnostic checkpoint (workspace verification pending):
Original ETest1, ETest2 and ETest3 are complete translations. All six fixture files
match the original Java vectors byte for byte. Preserve the original constructors,
SuiteETestCase flags and every assertion: exact semantic error count, argument-list
ranges and arity text for ETest1/2; deadlock exit, FINISHED, stats2/2/0, no GENERAL
and zero uncovered for ETest3. Inventory1068/1269 contexts (84.2%),548/626 complete
classes (87.5%),201pending/78classes; legacy85/104 complete.
Generator.java:900 reports incorrect arity at the final supplied argument-list
range, with the remaining signature. Diagnostics now carry optional SANY range
and text separately from their existing generic formatting. The shared TLC loader
prints actual semantic errors with Java's header/count and ErrorDetails shape;
warning elevation affects exit status without inflating the actual error count.
Existing warning-control, elevated-warning and full SANY Java race checks pass.
ETest1/2 race pass1.277s; ETest3 race pass5.849s. Related root parser/semantic/XML
and installed legacy contexts pass race60.790s before the context correction below.
No invented tests or weakened original assertions.

Full workspace83892 finished exit1, root1798.112s/SANY1.160s/TLC72.815s.
Its bc01728 snapshot failed original EvalExceptionLiveness trace variable order,
EWD998ChanDebugger variable selection, and BufferedRandomAccessFile postcondition.
This is authoritative red evidence; do not describe bc01728 as full-suite green.
The theorem-import filter admitted global builtins and ModuleNodes as theorem
entries because their declaration kind is empty. Filter those out and import
operator definitions before theorem/assumption definitions as Java Generator does.
The three unchanged failing contexts now pass69.425s (98866 terminal/retired),
log workspace-instance-context-fix-focused.log. Legacy contexts pass race64.544s
(71457 terminal/retired), log legacy-error-context-fix-race.log, including
ETest1–3 and Test201–209,216/219/220/99/999/InvalidInvariant. Current guard/final
code race recheck of the three failures plus ETest1–3 is live15192, log
workspace-instance-context-fix-race.log. Final corrected whole workspace is
live57071, log legacy-error-context-final-workspace-go.log.
Older workspace77547 completed exit0, root1806.145s/SANY1.001s/TLC70.894s,
log legacy-instance-exports-workspace-go.log; compiled before the theorem fix.
Current diagnostic workspace23927 remains live, log
legacy-error-reporting-final-workspace-go.log, but compiled BEFORE the context
filter correction above. Resume this handle; do not restart it on an observation
timeout and do not credit it as a gate for the later correction. Logs are under
.codex-gotmp/. After resolving these failures and recording current verification,
commit this batch and continue with original ETest4. Its exact nested-expression
stack and uncovered location must be preserved. The five deferred topics remain
deferred; existing tests in them must remain green. Never use testdata fixtures.
This entry supersedes older gate and next-context notes below.

2026-10-04 latest verified implementation checkpoint after63c8bc4:
Seven further original contexts are complete: Test209, Test219, Test99, Test999,
Test216, Test220 and TestInvalidInvariant. Preserve every original constructor,
flag, inherited/custom assertion, and all19 byte-identical fixture files. Test216
includes all three original dependency modules (216a/b/c); the earlier ignored
fixture draft omitted them and has been corrected before installation.
Test219 first failed semantic operator-argument checks: instance formal parameters
were counted in arity but missing from argument-position metadata. Prefix the
instance formal specifications as Java Generator.java:4880 does. Its next failure
was missing imported Inst!Foo: instantiate the whole source semantic context,
including nested imports, retaining SubstIn wrappers and source-node reuse.
Test216 then exposed omitted imported named theorems. Java imports those separately
as ThmOrAssumpDefNodes (Generator.java:4940 onward), using APSubstIn wrappers.
The bridge now preserves those context entries, actual theorem symbols, parameters,
locality, defining module and original source identity. Ordinary theorem references
resolve to their theorem symbols; no constant replacements or weakened tests.
Whole Test216 now passes28/7/0, INIT_GENERATED2 21/s/7, and the exact original
uncovered location. Test220 retains checkDeadLock=true so its config decides.
TestInvalidInvariant retains FAILURE_SAFETY_EVAL, the invariant-level diagnostic
and both GENERAL assertions. Whole99/999 normal pass1.097s/1.130s; race9.185s
(84526 terminal/retired). Whole216/220/InvalidInvariant normal pass1.374s/1.622s/
1.121s (71276/72356/71961 terminal/retired). Final focused root parser/semantic/XML,
all installed201–209/216/219/220/99/999/InvalidInvariant, cyclic INSTANCE,
native override and four bridge cases pass race82.897s (40154 terminal/retired).
Existing full SANY Java tests pass race7.978s (14888 terminal/retired), covering
the new instance-parameter semantic metadata; no later SANY semantic changes.
Inventory1065/1269 contexts (83.9%),545/626 complete classes (87.1%),204pending
across81 classes; legacy82/104 complete. Next original context: ETest1. Remaining
legacy contexts: ETest1–16, Test210/212/213/214/215/217. Preserve exact source
error counts, location ranges, diagnostic substrings, statuses and coverage.
Several ETests are runtime/config/simulation errors, not all SANY parse failures.
Do not replace their assertions with a generic failure or add invented tests.

Completed gates from the preceding snapshot: workspace44152 exits0/retired,
root1793.053s/SANY0.923s/TLC70.700s, log legacy-proof-fields-workspace-go.log.
It compiled the207/208 production code committed in63c8bc4, before the verbosity
harness and later INSTANCE/theorem changes. Whole CommunityModules verbose run
73656 exits0/retired, both unchanged phases pass306.247s, log
community-progress-verbose-go.log. That covers the progress-output change840d494
and preceding207/208 code, before the later INSTANCE/theorem implementation.
Live output, setup/loading messages and10s phase/elapsed/PID heartbeats confirmed.

Whole current TLC package race42143 passes646.992s (terminal/retired), log
legacy-theorem-imports-tlc-race.log.
Required workspace gates still active; resume existing handles, never restart:
-83892: final current-source whole workspace, legacy-theorem-imports-final-workspace-go.log.
-77547: older INSTANCE-export workspace, legacy-instance-exports-workspace-go.log;
 compiled before the theorem fix and later test files, so historical scope only.
All logs are under .codex-gotmp/. Do not claim full current-source green until
83892 reports terminal success. All focused and TLC package race checks are
green. Record
terminal gate receipts and fix any actual failure before proceeding to ETest1.
Preserve the five deferred topics and never create a testdata fixture directory.
The snapshots below are historical; this entry is authoritative.


2026-10-04 Test207/Test208 checkpoint (focused checks pass; full gates still running):
Original Test207 and Test208 are now complete translations, with all inherited
assertions, original constructor/settings, and four byte-identical source vectors.
Test207 exposed incorrect NEW declaration treatment in label validation: validate
preserved ASSUME/PROVE clauses, distinguish NEW OpDeclNodes from quantified formal
parameters, and track nested declaration scope in source order through PROVE.
Test208 exposed missing Java isFieldNameToken handling; reserved keyword record
components and EXCEPT paths now use Java's exact token ranges. Module identifiers
and the original initial keyword record-constructor branch use the same helper.
Neither original test nor fixture was weakened. Original207 passes2.032s;
original208 passes1.164s. Related parser/semantic/XML and original201–208 race
checks pass33.572s (65197 terminal/retired); earlier complete SANY Java race checks
pass8.386s. Inventory1058/1269 contexts,538/626 classes,211pending/88classes;
legacy75/104. Whole original Test209 passes1.205s in ignored overlay35080
(terminal/retired), not installed/credited. Whole original Test219 fails in
ignored overlay43516 (exit1/retired): semantic INSTANCE export metadata adds
parameterized instance arity but omits those parameter positions from operator
argument specifications. Fix the actual metadata before any later context.
Logs: .codex-gotmp/legacy-preflight209-go.log and legacy-preflight219-go.log.

CommunityModules progress-output change committed separately as840d494.
User's CommunityModules steering: keep the whole original test, but make go test
-v visibly report progress. community_modules_java_test.go now streams the child
stdout/stderr while preserving complete failure output, reports setup/dependency
and module-loading phases, and emits phase/elapsed-time/PID heartbeats every10s.
No original assertions, execution phases or production IOExec behavior changed.
Live verbose run73656 in .codex-gotmp/community-progress-verbose-go.log shows
module parsing and repeated10s heartbeats; this is evidence of progress output,
not a completed correctness gate. Run is still active: resume with write_stdin.
Whole workspace44152 in .codex-gotmp/legacy-proof-fields-workspace-go.log is
also still active; it compiled before the verbosity harness change, so its gate
covers the207/208 parser/semantic fixes. Do not restart or call either run green
until its authoritative handle/log reports terminal success. Finalize gate
receipts after these runs finish. This checkpoint records only focused green
checks; full-workspace success remains pending. Resume the Test219 metadata fix,
then install/verify/credit original209/219 once complete.
Preserve the five deferred topics and never create a testdata fixture directory.


Final gate correction: the first workspace run84712 failed after1460.845s
at original Test63's exact coverage assertion (terminal/retired); SANY1.100s
and TLC68.170s pass for that snapshot. Do not label the root snapshot green.
Reproduced unchanged Github680a/b/c then Test63:39115 exit1/retired, while12
isolated Test63 repetitions pass9.214s (28911 retired). Cause: the Go model-test
runner omitted reset/restore of TLCGlobals.warn, so earlier original -nowarning
flags changed Tool.getNextStates0's allAssigned path and coverage. Java Tool
line1006 uses the same flag; Ant's original runner forks perTest JVMs. Restore
Warn=true before each original class's flags, then restore prior Warn on cleanup.
No evaluator/reporting/assertion changes; all original -nowarning flags retained.
Three full Github680->Test63 race repetitions pass48.668s (85432 retired).
Final module/LET binding review keeps shared cached bindings for canonical
module INSTANCEs only; independently lowered LET contexts retain their own
bindings so captured outer formals are not merged by source position. Existing
parser/semantic/XML, original Test201–206, cyclic instance and four native-module
bridge checks pass race32.615s (51123 retired). Final current-source workspace
passes root1777.556s/SANY1.121s/TLC70.088s (2130 terminal/retired); log
`.codex-gotmp/selector-resolution-isolated-final-workspace-go.log`. All required
checks green. No production changes after that compilation; committed `b0ee5ae`.

2026-10-04 selector resolution after `fa672eb`: original Test206 now passes
with its entire inherited SuiteTestCase.testSpec, original constructor/settings,
all model assumptions and three byte-identical source vectors. Ordered source
selectors are resolved before ordinary semantic checks. Preserve each argument
group's arity, source definition/module, labels, LET contexts, lifted quantifier/
function/set/CHOOSE formals, and INSTANCE wrappers. Native TLC receives $Nop or
LAMBDA/OpArg/OpAppl nodes directly; no XML roundtrip. Final selected LabelNodes
are retained, `!>>` requires exactly two operands, and INSTANCE definitions and
selected SubstIn copies share cached module-instance bindings/substitution identities.
Selected operator substitutions are compiled rather than installed as raw
flattened names. XML's argument-bearing selector lambdas now consume the same
resolved selection and retain unused-formal argument weights/level omission.
Two old Go-only XML fixtures had selected a quantifier domain before binding
it, which Java rejects. Corrected only their source selector order; retained all
assertions. Both corrected fixtures pass Java SANY. No original test weakened,
new skips or invented persistent tests.
Focused existing parser/semantic/XML + original Test201–206 pass normal6.510s
(21537 retired), race30.383s (58546 retired). Earlier complete Test206 alone
passes1.973s (24502 retired); unchanged Java JUnit passes0.521s. Existing Java
SANY suite passes race8.373s (32940 retired). Existing INSTANCE/level checks pass race95.601s (12144 retired);
all four standard-module/native bridge checks pass race2.043s (5214 retired).
Workspace84712 later failed at Test63; see final gate correction above.
The superseded workspace84999 was interrupted after source-identity changes
(exit130/retired), not a successful gate. Inventory1056/1269 contexts (83.2%),
536/626 classes (85.6%),213 pending across90 classes; legacy73/104 complete.
Remaining selector-negative diagnostics and other legacy contexts are uncredited;
this test translation does not establish complete Generator selector parity.
Next: install original Test207 alone from the retained draft. Preserve the five deferred topics.


2026-10-04 selector work after `c1831ed`: added Generator.Selector's ordered
step representation to SanyExprSource (`sany_selector.go`, `ast.go`). Preserve
name/null/numeric/first/last/colon/@ kinds, selector syntax and each step's raw
argument node. N_OpApplication attaches its arguments to the final selector as
Java does; argument interpretation stays delayed until operator arity is known.
Added FindingSubExpr operand primitives, distinct from generic AST traversal:
call arguments exclude callee; record access retains string field operand;
function application keeps one tuple argument; CASE retains pair structure;
fairness operand order is subscript/action; quantified/function/set domains keep
source groups; labels unwrap; EXCEPT replacements retain Java's prohibition.
These primitives are not yet wired into semantic selector resolution/TLC lowering.
Test206 remains missing/uncredited; no test assertions changed or new tests added.
An ignored inspection probe of five original Test206 forms verifies exact ordered
steps/kinds/argument groups (64712 retired). Existing SANY parser checks pass
0.022s (67271 retired); existing parser + semantic bridge + all original Test201–205
pass race22.678s (37114 retired). No full-suite claim and no new inventory credit.
Next: connect the selector state machine to source definition/label/LET scopes,
lift bound formals with exact arity, retain INSTANCE/SubstIn identities and carry
selected expressions to native TLC. Fix original Test206 before another feature.


2026-10-04: Eighteen original legacy coverage/INSTANCE methods committed
as `dd022c6`; all normal/race checks green. Next original Test201–205 all pass:
complete lambda/recursive operator/recursive function/INSTANCE recursion and
full prefix/qualified operator assumptions retained, 13 byte-exact vectors.
Focused five normal 6.761s (89042 retired), race 22.646s (6503 retired).
No production changes. All five unchanged Java references pass.
Inventory 1055/1269 contexts (83.1%), 535/626 classes (85.5%),
214 pending across 91 classes; legacy 72/104 complete.
Discovery twelve-method run 99132 failed at Test206 after Test201–205 passed
(4.879s, retired). Java passes; Go rejects structural/operator subexpression
selectors as undefined names/incorrect arities, then graph-retention helper
panics because model construction failed. Complete Test206 and six unrun tests
remain ignored drafts and uncredited, with no weakened assertions/new skips.
Fix the real selector translation before installing them or advancing to a new
feature. No broad gates run against a known-failing draft. Committed production
and the five retained whole methods have green checks; deferred topics unchanged.


2026-10-04: Test34–49 committed as `b5dcd29`. Current batch ports 18 more
whole original legacy methods: Test50–60, Test62–65 and Test63a/64a/65a.
Two source test files preserve complete constructors, inherited settings and
assertions; Test57 retains its entire overriding method and safety exit without
adding inherited statistics/coverage assertions. All four additional coverage
hooks (Test52/55/56/63) retain the exact original strings, verified mechanically.
All 45 copied vectors match original bytes, including INSTANCE dependencies.
All 18 unchanged Java JUnit references pass. Go first seven normal 5.314s,
race 30.478s; next eleven normal 7.452s, race 42.566s. Handles 33873/33292/
80515/6050 all terminal/retired. No production changes or weakened assertions.
Inventory: 1050/1269 contexts (82.7%), 530/626 complete classes (84.7%),
219 pending across 96 classes, one partial; legacy suite 67/104 complete.
The five user-deferred topics remain deferred. All required checks green.


2026-10-04: Original legacy Test34–49 methods pass normal 10.110s after a
source failure exposed a bridge shortcut mapping user-definable `\subset` to
built-in `\subseteq`. Java distinguishes them; removing that false alias lets
normal symbol resolution call the user definition. Test44's full precedence/
operator model and Test48/49's complete definition/alias cases are preserved.
Related subset/module/config models and compact semantic checks pass 5.581s;
whole 16-method race passes 75.251s. All handles retired. TLC package production is unchanged from its
last full green gate. Inventory 1032/1269 contexts, 512/626 classes, 237 pending.


2026-10-04: Legacy Test27–33 and complete config literal recognition committed
as `956f530` after whole-method normal, direct config race, exact original
config/override model race and current full TLC checks. Additional complete
FPSetFactoryTest (15 original methods/helpers) passes normal 1.132s/race 2.348s;
unchanged Java passes all 15 (0.152s). No production change. Full source type,
nested memory/ratio/budget assertions and platform assumptions are retained.
Inventory: 1016/1269 contexts, 496/626 classes, 253 pending. All handles retired.


2026-10-04: Test19–26 committed as `c5d6b51`. Original Test27–33 now pass
normal (91.927s) after fixing ModelConfig's symbolic operator tokenization.
Java uses the SANY token manager for configuration files; Go now recognizes
its full generated literal vocabulary rather than splitting operators into
characters. `sany_generate.go` maintains both token tables from the same grammar;
all 274 literals and longest character-match order verified, with the root SANY
generated output byte unchanged. The original Test30 symbolic overrides pass
unchanged. Existing config tests pass race (1.032s). Current full TLC normal passes 67.357s; exact 10 original TLC config/override
models pass race 36.420s. All handles retired. Inventory: 1001/1269 contexts,
495/626 classes, 268 pending.


2026-10-04: Legacy Test1–18 and their four production fixes committed as
`051bb16` after full workspace normal and focused race checks. Additional
original Test19–26 methods pass normal (19.290s), preserving complete original
inputs, settings and assertions; no production changes. Test19 retains all
400000 initial/distinct states and 800000 generated states. All 8 also pass focused race (98.765s); all handles retired. Production remains identical to the verified workspace snapshot;
new source methods are checked separately. Inventory: 994/1269 contexts,
488/626 classes, 275 pending.


2026-10-04 current: original legacy Test1–18 methods exposed four corrected
SANY/TLC translation gaps. Each EXCEPT bracket is one function argument;
multiple indices form a tuple, preserving distinct path components. CHOOSE
with a tuple binder retains every formal and source position through semantic
scope and transformations, then emits one bounded symbol group marked as a
tuple. Existing TLC tuple CHOOSE evaluation binds its components and returns
the chosen tuple. Quantifier semantics use token kinds, including aliases.
Named function parameter counts match Java's bounded-symbol count: scalar
formals count separately, each tuple binder counts once. Java's static
function-application arity check remains intact.
All 18 original methods pass normal (9.238s) and race (66.557s). Full
workspace passes: root 1647.375s, SANY tests 0.935s, TLC 68.700s. Inventory: 986/1269 contexts, 480/626 classes, 283 pending.


2026-10-04 latest:41-context batch committedfcb3f4d after broad/race checks.
Current13original LongArray/LongArrays methods preserve full source inputs,
loops and assertions. Java language assertions carry AssertionError (an Error,
excluded from catch(Exception)), exact throwable class and nullable message;
LongArray bounds now use this core carrier. InfinitePrecisionIndexer source
constructor permits zero positions; empty sorting fixture remains unchanged.
All13normal PASS0.019s, relevant arrays/throwable/indexer race PASS31.603s,
fullTLC PASS66.714s. All handles retired.
Inventory968/1269contexts,462/626classes,301pending.

2026-10-04 latest: four original liveness cases and complete periodic coverage
boundary committed3822bf9. Current17 original liveness tests preserve complete
source methods/settings and expose two fixed shortcuts: check0 manufactured a
violation exception in addition to Java's result code, prematurely aborting
forced partial insertion; Tool.eval discarded the source coded undefined
fairness-variable diagnostic. The corrected paths retain source result/error
boundaries and detailed expression/context parameters. Eleven source methods
PASS normal6.943s/race41.715s; six PASS normal14.494s/race62.663s. Full
workspace PASS(root1645.594s,SANY1.010s,TLC65.352s); all handles retired. Additional original15offheap
indexer contexts pass complete1104-row matrices normal3.923s/race16.023s, with
no production changes. Additional original9iterator contexts pass normal0.010s after source reverse
positive-entry scan, read count and typed exhaustion/monotonic Assert boundaries.
Latest fullTLCnormal PASS69.871s/fingerprint-storage race PASS52.546s.
Inventory955/1269contexts,460/626classes,314pending.


2026-10-04 current correctness slice: twelve Examples methods/29pristine
fixtures PASS570.714s after preserving Java's substitution-constructor metadata.
YoYo plus19 original coverage race PASS100.888s; full TLC race PASS624.947s.
Eight original temporal-negation contexts/16fixtures PASS normal5.658s and
race26.526s. Four original success/initial-error contexts/8fixtures PASS4.967s
after preserving the original initialization exception code before replay.
Current initialization/temporal race PASS72.231s (23original methods), full
TLC normal PASS64.054s. Fresh full workspace PASS(root1621.178s,
SANY0.918s,TLC66.273s) with both fixes and all24new contexts. All gates green,
ready for authorized commit.
Earlier substitution-only workspace also PASS(root1600.051s), but the fresh
1621.178s gate is the current-source receipt.
TODO_TEST_PORT:910/1269 contexts (71.7%),433/626 complete classes,359pending.
The five deferred topics remain skipped; no original assertion weakened.



**Scope exception required by the user (2026-10-02): email reporting is
forbidden.** Email delivery and its JavaMail/SMTP/MIME/Activation/ImageIO/AWT/
codec dependencies, including mail-driven JVM emulation, are excluded from
this port's completion requirements. Stop that work immediately; no further
porting, audits, probes, test translation or downloads for that dependency chain.
The later user instruction authorizes surgical removal of the email-only
source, tests, fixtures, resources and stale plans. Preserve core distributed
behavior, packaged property loading, console output, generic exceptions,
OpenJDK notices and x/text. Continue core TLC and its existing Java tests.

This document records the architecture of the Java TLC model checker in
`../tlaplus/tlatools/org.lamport.tlatools/src/tlc2` and its tests in
`../tlaplus/tlatools/org.lamport.tlatools/test/tlc2`. It is meant to guide a
faithful Go port under `tlago/tlc`.

The porting rule for this effort is strict: implement TLC by following the Java
function decomposition mechanically before translating the Java tests. The Go
code may use Go idioms for memory, errors, goroutines, and concrete structs, but
semantic boundaries, evaluation order, fingerprinting, state normalization,
queue behavior, and error precedence must match Java TLC.

Current handoff and next-step status live in `HANDOFF.md` and
`PORT_PROGRESS.md`. This architecture document explains the source-of-truth
design and original porting order; use the handoff/progress files to decide
where to resume.

## Source Inventory

As of this survey, Java TLC contains:

- `371` Java files under `src/tlc2`.
- About `96,749` lines under `src/tlc2`.
- `699` Java test files under `test/tlc2`.
- `1,135` JUnit `@Test` sites under `test/tlc2`.
- `646` `.tla` fixtures and `524` `.cfg` fixtures under `test-model`.

Top-level source package sizes:

- `tlc2/tool`: model checking, simulation, state, trace, queues, fingerprint
  sets, liveness, coverage, distributed TLC, and most core algorithms.
- `tlc2/value`: TLA+ runtime value model, serialization, random enumerable
  support, model values, and value constants.
- `tlc2/util`: FP64, custom vectors/tables/stacks/queues, disk IO helpers,
  random generator, state writers, statistics.
- `tlc2/module`: Java implementations of TLA+ standard modules and TLC
  extension operators.
- `tlc2/output`: message codes, message formatting, recorders, state printing.
- `tlc2/model`: Toolbox-oriented model/error/trace presentation classes.
- `tlc2/debug`: debugger protocol and stack-frame support.
- `tlc2/overrides`: Java module override discovery and annotations.
- `tlc2/pprint`: pretty-printer support used by output paths.
  Go mirrors this as central concrete helpers in `pretty_print.go`.
  `ValuesPPR`/`ValuesPPRString` correspond to Java `tlc2.value.Values.ppr`
  and read the Java property `tlc2.value.Values.width` with default width 80.
- `tlc2.tool.impl.SpecProcessor.vetoed` controls Java's constant-operator
  pre-evaluation veto list. The Go constant-definition pass reads the same
  property through the central TLC property map.
- Java system properties that steer runner/checker behavior are read through
  the same central property map: `tlc2.TLC.nosuspend`,
  `tlc2.TLC.nohalt`, `tlc2.TLC.stopAfter`,
  `tlc2.tool.ModelChecker.vetoCleanup`, and
  `tlc2.tool.ModelChecker.BAQueue`.
- Disk state queues and disk byte-array queues both use Java's
  `tlc2.tool.queue.DiskStateQueue.BufSize` property, defaulting to 8192.
- FPSet construction follows Java properties through the central property map:
  `tlc2.tool.fp.FPSet.impl`, `tlc2.tool.fp.DiskFPSet.logLockCnt`, and
  `tlc2.tool.fp.OffHeapDiskFPSet.probeLimit`. Disk fingerprint sets also
  honor `tlc2.tool.fp.DiskFPSet.metadirPrefix` and
  `tlc2.tool.fp.DiskFPSet.error2warning`.
- Tableau construction mirrors Java's debug export hook:
  `tlc2.tool.liveness.Liveness.tableauExportPath` writes `TBGraph` DOT output.
- Coverage cost-model creation mirrors
  `tlc2.tool.coverage.CostModelCreator.implied`, defaulting to true for
  implied init/action coverage.

Top-level test package sizes:

- `tlc2/tool`: the dominant end-to-end model-checking suite.
- `tlc2/tool/liveness`: temporal property, tableau, SCC, symmetry, simulation,
  and trace/lasso tests.
- `tlc2/tool/fp`: fingerprint set implementations and disk/off-heap indexing.
- `tlc2/tool/queue`: state queue behavior and disk pool writer tests.
- `tlc2/value/impl`: value semantics and enumerator tests.
- `tlc2/module`: standard module and TLC extension behavior.
- `tlc2/debug`: debugger protocol integration tests.
- `tlc2/util`: FP64, vectors, queues, stacks, statistics, byte encoding.
- `tlc2/model` and `tlc2/output`: Toolbox presentation and message behavior.

## Package Strategy in Go

Keep TLC mostly in package `tlc` under `tlago/tlc`, with subdirectories only
when Go requires it for commands or frozen fixtures. This follows the user's
preference to avoid circular imports. Java packages should map to file groups
inside the same Go package:

- `cli_*.go`: command parsing and top-level run orchestration.
- `tool_*.go`: `Tool`, `FastTool`, spec processing, expression evaluation.
- `state_*.go`: `TLCState`, mutable state variants, state vectors.
- `value_*.go`: value hierarchy and enumerable values.
- `fp_*.go`: fingerprinting and fingerprint sets.
- `queue_*.go`: state queues and checkpointing.
- `liveness_*.go`: temporal formula nodes, tableau, behavior graph, SCC.
- `module_*.go`: TLC/Naturals/Integers/Sequences/etc. built-ins.
- `output_*.go`: message codes and recorder-compatible output.
- `config_*.go`: `.cfg` parser and model configuration.

The Go package can still preserve Java names in type and method names where
useful during the mechanical port. After tests are green, cosmetic refactors can
be considered separately.

Prefer concrete structs over Go interfaces throughout the port. Java uses
interfaces such as `ITool`, `IStateFunctor`, `INextStateFunctor`, `IWorker`,
`FPSet`, and `IStateQueue` heavily, but the Go port should introduce an
interface only when there are already multiple meaningful implementations or
when polymorphism is itself part of the runtime value model. For single
implementations, use concrete structs such as `Tool`, `StateFunctor`,
`NextStateFunctor`, `Worker`, `MemFPSet`, `MemStateQueue`, `ModelConfig`, and
`ConfigConstant`; this keeps stack traces, debugger watches, and crash dumps
straightforward during the mechanical port.

When Java has an interface with exactly one production implementation, port the
implementation as a concrete Go struct and let call sites name that struct
directly. `IMVPerm`/`MVPerm` is the model: Java exposes an interface, but Go uses
`*MVPerm` because there is only one real permutation representation and the
fixed indexed array is important to understand in a debugger.

Any Go map whose iteration can affect output, fingerprinting, exploration
order, diagnostics, or tests must use `InsMap` from `insmap.go`. Built-in Go
maps are acceptable for lookup-only sets/tables that are never ranged over in
observable code.

## Main Execution Architecture

### `tlc2.TLC`

`tlc2.TLC` is the top-level runner. Its responsibilities are:

- Parse command-line flags.
- Select run mode: exhaustive model checking or random simulation.
- Resolve spec/config/metadir paths.
- Configure global TLC options in `TLCGlobals`.
- Initialize FP64 polynomial selection.
- Create the `Tool`/`FastTool` spec handle.
- Create an `FPSet`.
- Create a `ModelChecker`, `DFIDModelChecker`, or `Simulator`.
- Wire optional state writers (`-dump`, `-dot`).
- Start debugger mode when requested.
- Print welcome, progress, error, and summary messages.

Important fields:

- `runMode`: `MODEL_CHECK` or `SIMULATE`.
- `cleanup`: remove metadata after successful run.
- `deadlock`: whether deadlock checking is enabled.
- `seed`, `aril`, `traceDepth`, `traceNum`: simulation control.
- `mainFile`, `configFile`, `metadir`, `fromChkpt`: model input and metadata.
- `fpIndex`, `fpSetConfiguration`: fingerprint polynomial and storage config.
- `stateWriter`: optional trace/graph dumping sink.
- `tool`: volatile handle used by debugger and runtime components.

Port guidance:

- Implement a `Runner` or `TLC` struct with the same lifecycle.
- Do not hide global options initially. Java code depends heavily on
  `TLCGlobals`; mirror it with a package-level `Globals` struct to preserve
  behavior, then later reduce global mutation if tests allow.
- Keep version metadata centralized in the TLC globals layer. Java's
  `TLCGlobals.Version` feeds management beans, `TLCGet("revision")`, and
  `TLCGet("config")` install fields; Go should route all of those through the
  same helper functions.
- Preserve Java's global property defaults and live property updates:
  `tlc2.TLC.progressInterval` is seconds coerced with `max(abs(x), 1)`,
  `tlc2.TLCGlobals.chkpt` is milliseconds, and
  `tlc2.TLCGlobals.coverage` is the coverage bitmask. Go initializes these
  from the process environment and updates them when the Java-style `-D`
  parser or `TLCSet("-D...")` writes the same property names.
- Preserve exit-status categories and message codes, because the Java tests
  assert message recorder events rather than only stdout text.
- Mirror `handleParameters` as a real library parser, not only as command
  wrapper glue. Java accepts a broad command vocabulary before it constructs
  `FastTool`, and later subsystems depend on those parsed fields being present:
  `-simulate`/`-generate` with `num=`, `file=`, `stats=basic|full`, and
  `sched=rl|rlaction`; model-checking controls such as `-dfid`, `-workers`,
  `-recover`, `-metadir`, `-checkpoint`, `-coverage`, `-cleanup`, `-deadlock`;
  evaluator/output globals such as `-difftrace`, `-nowarning`, `-gzip`,
  `-terse`, `-continue`, `-view`, `-debug`, `-tool`; fingerprint controls
  `-fp`, `-fpmem`, and `-fpbits`; debugger controls `-debugger` with optional
  `port=`, `nosuspend`, and `nohalt`; trace exploration controls
  `-generateSpecTE`, `nomonolith`, `-noTE`, `-noTEBin`, and `-teSpecOutDir`;
  state/trace dump controls `-dump`, `-dumpTrace`, and `-loadTrace`; and
  runtime spec additions `-inv`, `-invlevel`, and `-postCondition`.
- The Go port keeps these in concrete structs: `Options` for direct runner
  fields, `SimulationSchedule` for RL scheduling, and `RuntimeParameters` for
  dynamically generated invariant/constraint/postcondition/view additions.
  This mirrors Java's `params` map without introducing an untyped central map
  into the Go API. `InsMap` backs parsed message-control sets to keep any later
  iteration deterministic.
- Runtime parameters follow Java `ParameterizedSpecObj` ordering. Runtime
  extendees are postcondition modules, invariant dependency modules, model
  constraint modules, action constraint modules, then view. Runtime string
  constants are bound in the same order Java's `processConstantDefns` visits
  them: model constraints, action constraints, postconditions, invariant
  dependency modules if they declare constants, then view. The current Go
  runtime invariant template form has no constant definitions of its own, but
  view constants are still carried through `RuntimeView` like Java
  `ParameterizedSpecObj.View.constDefs`. The Go bridge installs these runtime
  constants before ordinary config constants because Java
  `ParameterizedSpecObj.processConstantDefns` runs before
  `SpecObj.processConstantDefns`; if the same constant name appears in both,
  the config constant wins. Runtime invariant actions are appended after normal
  config processing, so the missing `INIT`/`NEXT` checks only see the static
  model config just as Java's `SpecProcessor` does.
- Runtime postconditions follow Java `Spec.getPostConditionSpecs`: actions from
  `ParameterizedSpecObj.getPostConditionSpecs()` come before config-file
  `POSTCONDITION(S)`, and their user-visible action name is the unqualified
  operator name even though the Go bridge resolves the qualified module
  definition internally. This matters for diagnostic/postcondition output order.
- Java's `-dump class,...` loads an `IStateWriter` by reflection. Go keeps this
  concrete with `StateWriterFactory` functions registered by class name through
  `RegisterStateWriterClass`; built-in parity names cover Java's zero-argument
  `NoopStateWriter` and `DotStateWriter`, and unknown class names fail during
  option parsing instead of silently dropping state output. The requested class
  name is still recorded in `RuntimeParameters.CustomStateWriterClass` for
  downstream visibility.
- Java's `DotStateWriter` prints full successor labels unless
  `TLCGlobals.printDiffsOnly` is set. Stuttering is an explicit
  `Visualization.STUTTERING` hint; an ordinary self-loop is still an ordinary
  edge. Keep those decisions on the concrete `StateWriter` rather than adding
  writer interfaces.
- Java's CLI DOT writer constructor always writes a `strict ` graph header
  prefix, even when the parsed `strict` sub-option is false; the sub-option only
  controls the in-memory duplicate-edge suppression set. Go models this with a
  separate `StrictPrefix` option for the CLI path. The zero-argument custom
  `DotStateWriter` and liveness DOT writer use Java's prefix-free constructors.
- DOT transition colors are assigned by action name, but transition labels use
  `Action.getInvocationSignature()`, so parameterized action instances print
  their concrete values on edges.
- `DotStateWriter` fingerprints and ranks the original states but labels
  initial and successor nodes with `state.evalStateLevelAlias()`. Successor
  tooltips remain the original state text. The local `ModelCheckerMXWrapper`
  also renders `getCurrentState()` through the same state-level alias; the
  distributed server wrapper does not.
- Java defaults simulation `traceNum` to `Long.MAX_VALUE`, not one trace. Go
  runner and simulator defaults must preserve that, with explicit `num=` or API
  options narrowing the trace count.
- Per-message suppression/elevation belongs in the output recorder path because
  Java's tests observe `MP` recorder events. In Go, `PrintWarning` honors
  `Globals.Warn` and per-code suppression, and `MessagesAsErrors` upgrades the
  recorded severity.

### `Tool`, `FastTool`, and `ITool`

`tlc2.tool.impl.Tool` is the core runtime/evaluator API. `FastTool` subclasses
`Tool` only to mark hot methods final and inline-friendly in Java. In Go, make a
single concrete `Tool` type unless a debugger/call-stack wrapper requires
separate structs.

The `ITool` boundary includes:

- `GetActions`
- `GetInitStates`
- `MakeState`
- `GetNextStates`
- `Eval`
- `IsGoodState`
- `IsInModel`
- `IsInActions`
- `EvalReward`
- `Enabled`
- `IsValid`
- `CheckAssumptions`
- `CheckPostCondition`
- `GetState` trace reconstruction
- `GetSymmetryPerms`
- `Contexts`
- getters for invariants, implied actions, temporals, constraints, aliases,
  postconditions, config, module files, and root names.

The `Tool` constructor does important work:

1. It calls `Spec`/`SpecProcessor` to parse and process the spec/config.
2. It initializes static state variable metadata on `TLCStateMut` or
   `TLCStateMutExt`.
3. It obtains the next-state specification.
4. It splits the next-state predicate into actions.
5. It assigns stable numeric IDs to init predicates and actions.

Port guidance:

- Keep `Tool` as the central object that holds processed SANY semantic nodes,
  config-derived declarations, and evaluator helpers.
- Preserve the Java call order. Many tests depend on exactly when constants,
  model values, overrides, variable locations, and action names are available.
- Port `FastTool` as either an alias/wrapper or simply as constructor options
  on `Tool`. The important thing is behavior, not Java inheritance.
- Java's `SymbolNodeValueLookupProvider` is a default-method interface mixed
  into `Tool` and `TraceApp`. In Go it should stay as concrete `Tool` methods:
  `LookupWithCutoff`, `GetVal`, `GetOpContext`, `GetVar`, `GetLevelBound`, and
  `GetLevelBoundAppl`. This keeps lookup semantics centralized without adding a
  one-implementation interface.
- Mirror Java `ITool` accessors as concrete `Tool` fields and methods. In
  particular, expose `SpecProcessor`, module-file paths, assumptions,
  assumption-axiom flags, and `CounterExample`'s backing operator definition on
  `Tool` itself instead of adding an `ITool`-style interface.

## Spec and Config Processing

### `Spec`

`tlc2.tool.impl.Spec` wraps parsed SANY semantic artifacts and exposes processed
model components:

- root module, module table, resolver, root file/config file/spec dir.
- variables and primed variable locations.
- initial predicate actions.
- next-state action.
- model constraints and action constraints.
- view, alias, symmetry, temporals, invariants, implied init/actions.
- assumptions and postconditions.
- lookup and substitution helpers.

It also collects primed variable locations for action/state-level evaluation.

### `SpecProcessor`

`SpecProcessor` converts SANY output plus the model config into TLC runtime
objects. Key jobs:

- Process constant definitions and module constants.
- Process Java module overrides.
- Apply model overrides and substitutions.
- Resolve `INIT`, `NEXT`, `SPECIFICATION`, `INVARIANT`, `PROPERTY`,
  `CONSTRAINT`, `ACTION_CONSTRAINT`, `VIEW`, `SYMMETRY`, `ALIAS`,
  `POSTCONDITION`, `_PERIODIC`, `_RL_REWARD`, and `_POSSIBLE`.
- Split temporal formulas into implied init/action/temporal pieces.
- Build `Action` objects with contexts and names.
- Initialize `ModelValue` state.
- Set variable locations on `UniqueString`.
- Retain the variable declaration symbols and their semantic source locations.
  `SetVariableNodes` stores the parser bridge's declaration identities, and
  `ApplyToTool` preserves their module/line/column metadata in the state table.
  `GetVariablesNodes` exposes the backing declaration array, matching Java.
- Set the definition table count to the number of variables before storing
  definitions. Java does this explicitly with `defns.setDefnCount(varDecls.length)`;
  the Go port keeps the same visible step instead of hiding it in `Defns`.
- Expose the same processed-spec accessor surface as Java
  `SpecProcessor`: init/next predicates, temporal and implied temporal actions,
  invariants, implied init/action checks, model/action constraints, assumptions,
  `_RL_REWARD`, `_PERIODIC`, definition snapshots, constant-definition cache,
  and postcondition specs. Most Go accessors return slice copies;
  `GetVariablesNodes` retains Java's backing-array behavior.

Tricky details:

- The config table is intentionally heterogeneous in Java: values can be
  strings, vectors, raw values, model values, or override descriptions. A
  cleaned-up Go representation is fine only if it preserves every parse-time
  and process-time behavior.
- Module overrides are part of semantics, not a performance-only feature.
- Model values must be initialized and ordered exactly as Java does because
  comparison, printing, and fingerprinting depend on it.
- Constants can be static, dynamic, module-scoped, or override-driven.
- `tlc2.tool.impl.ModelConfig.nosymmetry=true` makes Java's
  `ModelConfig.getSymmetry()` return the empty string even when the cfg file
  contains `SYMMETRY`; the Go port mirrors this at `GetSymmetry` so downstream
  tool/state setup sees no symmetry set.
- Config processing must rebuild every derived vector from scratch when it is
  run. The Go bridge constructs `SpecProcessor` before all definitions and
  variables are installed, then runs it again while applying the processor to a
  `Tool`; accumulated init/property/temporal vectors would diverge from Java.
  Capture Java's `Snapshot` after config constants/overrides and before
  `ProcessConstantDefinitions`. `getUnprocessedDefns`, symmetry, `_RL_REWARD`
  and `_PERIODIC` use it after ordinary constants have been evaluated. The
  separate native-only `preConstantSnapshot` is captured after the predefined
  TRUE/FALSE/BOOLEAN and Strings.STRING MethodValue entries, before ordinary
  definitions/config bindings. Represented `processConstants` walks modules,
  operator and LET bodies, substitutions, assumptions, bounds, labels and
  operator arguments, using an identity set of processed OpDefs. Initial and
  dynamic extendees share that snapshot independently of pre-evaluation. SANY
  numerals retain radix/image and big-integer metadata; constant processing
  rejects large integers and decimal nodes with the original Java messages.
  Theorem/proof graph traversal and full native override registration remain
  broader source work.
- Config keyword validation is keyword-specific. `INIT`, `NEXT`, `VIEW`,
  `POSTCONDITION`, `_POSSIBLE`, `_PERIODIC`, and `_RL_REWARD` require
  zero-arity operator definitions; `INVARIANT` and `PROPERTY` ignore literal
  `TRUE` but reject literal `FALSE` or non-boolean values with
  `TLC_CONFIG_ID_HAS_VALUE`; `SPECIFICATION` has its own value error wording.
- Structural config decomposition has the same checks when it follows a
  zero-argument operator reference inside `SPECIFICATION` or `PROPERTY`: Java
  rejects references to parameterized operators, unknown names, literal
  `FALSE`, and non-boolean values immediately instead of falling through and
  classifying the expression by level.
- Model constraints and action constraints accept zero-arity operator
  definitions. Java stores the operator definition on the body node for later
  coverage reporting, appends the body to the constraint list, ignores literal
  `TRUE` values, and treats non-zero-arity definitions, literal `FALSE`,
  non-boolean values, and unknown names as configuration errors before model
  checking begins.
- After definitions and config substitutions are installed,
  `SpecProcessor.ProcessConstantDefinitions` visits the actual external module
  table in dependency order. It skips instantiated modules with parameters,
  also checking each operator's originally defining module, then recurses into
  inner modules without reapplying module eligibility. Constant declarations
  read their own tool object and initialize explicit values. Operator
  pre-evaluation requires lookup to return an operator definition, zero source
  arity and constant effective level; vetoes use the resolved operator name.
  Each result is immediately installed on that exact node. The global `Defns`
  entry changes only if it still denotes that node; module-map keys use its
  source origin when present. Evaluation failures for ordinary operators are
  swallowed as Java's `Throwable` catch requires, leaving the use-site error
  for checking. Declared operator substitutions retain their narrower exception
  boundary and nonconstant/arity diagnostics. Vetoes use
  `tlc2.tool.impl.SpecProcessor.vetoed` or `TLAGO_SPEC_PROCESSOR_VETOED`.
- Java routes each successfully evaluated zero-arity constant operator through
  `WorkerValue.demux`. The value is deep-normalized but not eagerly initialized
  for fingerprinting at spec-processing time. Mutable values with multiple
  workers are re-evaluated once per worker under the same
  `RandomEnumerableValues` seed and stored as a `WorkerValue`; immutable
  primitives stay as plain values. The Go port retains that storage in `Defns`,
  the operator tool object and `ConstantDefns`, whose shape is module identity
  -> declaration/operator identity -> raw value or `WorkerValue`. Debugger
  consumers mux for their current worker. One module's constants are flattened;
  multiple modules have nested groups with source module names, retained
  compound instance paths and local operator/declaration names.

### Go SANY to TLC Bridge

Application loading includes runtime extendee modules before building the tool,
so trace postconditions and runtime constraints can resolve their operators.
Actions created from definitions reuse the body converted in that definition's
module scope; reconverting at the root loses bindings to LOCAL operators.
The initial definition table includes TRUE, FALSE, the normalized BOOLEAN set
{FALSE, TRUE}, and STRING before processing user definitions. Root registration
uses the retained module's actual operator definitions, including transitive
EXTENDS exports omitted by the central runtime index. Source contexts preserve
original definitions even when a qualified runtime alias denotes a named
INSTANCE. Native overrides visit each module's actual operators in source
module order and attach to shared source bodies. The evaluator's native name
cache follows those bodies while `Defns` retains semantic INSTANCE entries.
TLCExt's TLCGetAndSet is evaluated from its source TLA+ definition. The REPL
reads its actual root operator, matching Java's ModuleNode lookup, without
reconverting definitions into a second graph. Root assumption collection copies
extendees' complete vectors in EXTENDS order before local assumptions, sharing
source expression identities across repeated paths and excluding INSTANCE.
The source AXIOM keyword is retained for runtime assumption checks. Runtime
string constants bind actual module declaration tool objects before per-module
constant processing, including the JSON trace-file constant shown by the debugger.

The root package owns the production Go SANY parser and semantic tree. The TLC
runtime package must not import the root package, so the adapter lives in the
root package as `BuildTLCTool`. Import direction is one-way: root `tlago`
imports concrete `tlc` structs and builds a `Tool`; `tlc` remains independent.

Current adapter responsibilities:

- collect the root module variables and install TLC state variable locations.
- convert SANY `Definition` and `Expr` nodes into TLC `OpDefNode`,
  `OpApplNode`, `LetInNode`, quantifier, value, and action nodes.
- install config constants, including arity-bearing operator-constant rows as
  `OpRcdValue`, module-qualified constants as `Module!Name`, and install
  operator overrides into `Tool`.
- attach stable `SymbolNode` identities to bridged `OpDefNode`s so LET binding
  and later operator application use the same symbols.
- install named-instance aliases for TLC standard-module overrides and local
  LET instance aliases such as `T!PrintT`.
- evaluate `ALIAS` config operators by converting record/function-record values
  into printable TLC states. Java's `RecordValue.toState()` permits aliases
  that are subsets or supersets of the spec variables; Go mirrors that with a
  concrete record-backed state, preserving state-variable fingerprint/equality
  behavior while printing the alias record and honoring `_format`.
  `SpecProcessor` installs `Tool.AliasSpec`; the bridge does not replace the
  tool's alias overloads. Prefix-aware evaluation binds imported `TLCExt.Trace`
  to `LazySupplierValue`, while pairwise evaluation uses an empty context.
  Only `EvalException`/`TLCRuntimeException` become `_ALIASEvalError` fields;
  other failures escape. Printable record states retain their fresh metadata,
  and `AliasTLCStateInfo` preserves the original ordinal/action separately.
  The native `TLCExt.getTrace` override does not look up this context itself.
  `_format` now distinguishes absent/default formatting from an explicitly
  empty string, retains source null/class-cast failures and accepts the concrete
  DebuggerValue StringValue subtype. Record-to-state conversion binds every
  matching field in source order, preserving last-binding behavior for raw
  duplicate fields. State formatting uses the shared Java string-argument
  formatter, with whole-template parsing/validation before argument rendering,
  independent ordinary/explicit/previous indexes, typed source exceptions,
  flags/width/UTF-16 precision, boolean and Java string-hash conversions,
  uppercase and platform line separators. Numeric, character and date/time
  templates reject the source string arguments as Java does. Locale/category
  startup properties freeze on first formatter use; OS locale discovery remains
  work. All 60,551 source-grammar observations, every Unicode scalar uppercase
  mapping, all 65,536 code-unit print observations and eighteen record cases
  match Java. Unpaired UTF-16 units are retained as WTF-8 for internal formatting;
  the CSV UTF-8 boundary replaces them with '?' as Java's encoder does. Other
  UTF-16 consumers remain source work; this is not full string-system parity.
  The existing RecordValueTest translation now retains both whole methods,
  all fifteen assertions and the original conditional catch behavior.
- resolve config-selected `INIT`, `NEXT`, `SPECIFICATION`, invariants,
  properties, constraints, view, and postconditions into `Action` or semantic
  nodes.
- preserve deterministic definition installation by sorting the collected names.

This bridge is a staging boundary, not a replacement for Java `SpecProcessor`.
The mechanical port still has to move exact Java visibility, full INSTANCE
processing for parameterized and non-standard modules, exact module-constant
scoping, action decomposition, implied init/action splitting, symmetry, aliases,
`_POSSIBLE`, `_PERIODIC`, and `_RL_REWARD` into the TLC-side `SpecProcessor`
shape. Until that is complete, the adapter should stay simple and explicit so
mismatches are easy to see.

### `Defns` and `Specs`

`Defns` is a compact definition table keyed indirectly by `UniqueString.loc`.
The same `loc` field is also used for state variable positions:

- variables occupy locations `[0, varCount)`.
- definitions start at `defnIdx`, normally initialized to `varCount`.
- `UniqueString.getDefnLoc` returns `-1` for variable locations.
- `put` assigns a new location only when the key's definition location is `-1`.
- `snapshot` copies the backing array and the current index.

`Specs.getLevel` is a static helper used by TLC to compute the effective level
of a level-checked expression under a context. It starts from the node's SANY
level, then follows each level parameter through the context. If the binding is
a `LazyValue`, it recurses on the lazy expression and context; if the binding is
an `OpDefNode`, it recurses on that operator definition. The Go semantic nodes
therefore carry concrete level metadata on `SemanticNodeBase`, and `OpDefNode`
embeds the same base rather than hiding level behind an interface.

`Specs.addSubsts` wraps an expression in the queued `SubstInNode` substitutions
in list order. This is a structural SANY helper and should remain mechanical.

`LET` context extension is shared by evaluation, enabledness, init generation,
next-state generation, and variable lookup. Zero-arity local definitions bind as
lazy values, while arity-bearing local operators bind as concrete `OpDefNode`
values in the same context. This mirrors Java's semantic-node environment
without introducing a one-implementation closure interface. Local module
instance aliases that resolve directly to built-in method values are stored as
explicit `LetBinding` entries on `LetInNode`; the same binding path is used by
evaluation, enabledness, init, next, level calculation, and variable discovery.
Parameterized theorem and assumption definitions use Java's separate
`Spec.getOpContext(ThmOrAssumpDefNode, ...)` path: bind their formal parameters
with `getVal(arg, context, cachable)` before evaluating the body. The Go
`ThmOrAssumpDefNode` therefore carries `Params`, and eval/init/next/enabled all
route through `GetThmOrAssumpContext`.

### Model Presentation

`tlc2/model` contains Toolbox-facing data holders, not checker algorithms:
`Formula`, `Assignment`, `TypedSet`, `MCVariable`, `MCState`, `MCError`, and
trace-expression metadata. Keep these as concrete structs in `model.go`.

`TypedSet` owns its own string, equality, hash, and type-validation helpers,
mirroring Java's `TypedSet` class. Spec-writing code should call those methods
rather than define model formatting locally. Its parser also preserves Java
`String.split(pattern, 0)` behavior: trailing empty pieces are discarded, so
inputs such as `", , , ,"` and `"{, , , ,}"` parse as the empty set.

`MCVariable.isTraceExplorerExpression()` is a nullness test in Java, not a
non-empty-string test. Go therefore tracks whether `SetTraceExpression` was
called separately from the stored expression text. `MCState` conjunctive
descriptions also preserve Java's optional ANSI bold/reset wrapping for trace
expressions.

`MCError` has two easy-to-clean-up Java quirks that affect trace exploration:
`toSequenceOfRecords` decides comma insertion from the original state index, not
from whether a previous non-marker record was emitted, and
`isLassoWithDuplicates` compares the full lasso trace length against the unique
normalized records from all states except the final back-to-state marker. The
method name sounds narrower than the algorithm; keep the algorithm because it
controls whether a liveness trace gets an explicit trace-view definition.

The trace-expression spec writer is also intentionally strict like Java.
Back-to-state targets and next-state variables are indexed directly, so malformed
traces fail loudly instead of producing a best-effort TE spec. The
`TLC_TRACE_EXPLORER_JSON_UNCOMMENTED` property removes comment prefixes from the
JSON `ASSUME` block just as Java does for tests, and `indentString` uses Java
split semantics by dropping trailing empty lines. The concrete
`TraceExpressionExplorerSpecWriter` stores expressions in Java `TreeMap` order,
so generated declarations and conjuncts use lexicographic variable-name order
rather than insertion order.

`StatePrinter` is the Java helper that gives state trace messages their
metadata. Preserve its small details: invariant traces use predecessor diffs
only when `TLCGlobals.printDiffsOnly` is set, incomplete states print
fingerprint `-1`, `TLC_BACK_TO_STATE` routes through `printState` only in
tool mode, and the opt-in `tlc2.output.StatePrinter.overwrite` property sleeps
for the configured milliseconds and emits terminal reset sequences unless the
state is final.

Generic spec-writer utilities also carry visible Java details. The generated
identifier counter starts like Java's `AtomicLong(1L)` and uses
`incrementAndGet()` before formatting, so a fresh process first emits suffix
`2000`; timestamps and modification-history dates use Java `Date.toString()`-
style strings rather than Go's `time.Time.String()` output.

`ErrorTraceMessagePrinterRecorder` observes printed state messages to reconstruct
the error trace. Its back-to-state path treats any parseable positive ordinal as
a terminal marker before it validates that the ordinal points at an existing
state, so Go's recorder must finish the trace before target validation too.

### `ModelConfig`

`ModelConfig` parses `.cfg` files with the TLA+ token manager rather than a
separate parser. It recognizes the keywords listed above, stores raw constants,
module constants, operator overrides, and the deadlock setting.

Port guidance:

- Keep the Go representation explicit and concrete rather than reproducing
  Java's heterogeneous `Hashtable`. Use `ConfigConstant`, `ConfigConstants`,
  typed string slices, and `InsMap` for module-scoped or override tables.
- Keep config grammar behavior separate and compatible with Java. The TLC
  package should not import the root `tlago` parser package just to tokenize
  config files, because the root package will later need to wire parsed specs
  into TLC.
- Preserve duplicate-key errors, missing identifier errors, and keyword
  pluralization behavior.
- Preserve support for configs embedded in monolithic `.tla` files.

## Action Decomposition

`Tool.getActions` splits a next-state predicate for performance. Instead of
treating `Next` as one large action, Java TLC decomposes the maximum prefix of:

- disjunctions,
- bounded existentials,
- substitutions,
- selected user-defined operator bodies.

This yields separate `Action` objects, each with:

- `pred`: semantic predicate.
- `con`: context bindings.
- `actionName`: name used in traces and coverage.
- `opDef`: optional declaration node.
- `id`: stable action/init id.
- flags for init/internal actions.
- auxiliary metadata map.
- coverage cost model.

Correctness notes:

- Decomposition must not change enabledness or generated successor states.
- Action names in error traces and coverage depend on this decomposition.
- Bounded existential action splitting is observable through trace action names.
- Some specs have unnamed actions directly inside `Spec == Init /\ [][...]_vars`.
- Java distinguishes an action declaration from its definition: declaration is
  the left-hand-side `IdentLHS` tree-node location of the operator definition,
  while definition is the predicate/body source location. Coverage roots use the
  declaration; debugger, trace, DOT, and counterexample display use the
  definition through `Action.getLocation`/`getDefinition`.

Performance notes:

- More precise splitting avoids repeated evaluation of irrelevant disjuncts.
- Context allocation during splitting and next-state generation is hot.
- Go should avoid reflection or map-heavy inner loops here.

## Expression Evaluation

`Tool.eval` and `Tool.evalAppl` are TLC's interpreter for SANY semantic nodes.
It evaluates:

- constants and variables in contexts/states.
- user-defined operators and recursive operators.
- LET/IN and substitutions.
- operator arguments and higher-order operator values.
- builtin TLA+ operators through `BuiltInOPs` opcodes.
- sets, tuples, records, functions, function application, EXCEPT.
- bounded/unbounded quantifiers where supported by TLC.
- temporal/action operators when used by enabledness/liveness machinery.
- Java module overrides.

Core control inputs:

- `Context`: linked-list binding chain with branch/cutoff behavior for
  `ENABLED`.
- `s0`: current state.
- `s1`: successor/prototype state.
- `EvalControl`: control flags for priming, enabledness, and evaluation mode.
  The helper `PartialBoolean` mirrors Java enum order `YES, NO, MAYBE`; avoid
  relying on Go's zero value as semantic "maybe", and let invalid values fail
  loudly like Java's unreachable enum default path.
- `CostModel`: coverage accounting.

Important implementation patterns:

- `evalImpl` dispatches on semantic node kind.
- `evalApplImpl` dispatches on builtin opcode or user-defined operators.
- Values returned by evaluation are `Value` implementations and may be lazy.
- When a primed variable lookup cannot be resolved, Java reports an incomplete
  state only outside `ENABLED`. Under `ENABLED`, the same primed lookup must fall
  through to the ordinary undefined-operator handling so enabledness exploration
  can continue to use its special context/control flow.
- `GetVar` must recurse through `SubstInNode`, `APSubstInNode`, `LetInNode`,
  labels, lazy values, and operator definitions before deciding that an
  operator application is a state variable.
- `GetPrimedVar` has the same decomposition shape and only succeeds when the
  expanded expression is a prime application whose argument resolves through
  `GetVar`. This is how Java recognizes assignments hidden behind labels,
  substitutions, LET aliases, lazy values, or zero-arity operator definitions.
- `GetLevelBound` is only a conservative bound. It returns temporal/action
  constants immediately for temporal/action opcodes, treats `ENABLED` as state
  level, scans bounded-quantifier ranges and arguments, and follows user
  operator definitions, lazy values, `EvaluatingValue`, and `MethodValue`.
  For `LET`, Java first takes the maximum level of every local definition body,
  then binds each local operator name to `1` before checking the `IN` body; use
  the same pattern to avoid recursive level walks and to classify hidden
  higher-level local definitions conservatively.
  Recursive functions use the same `name -> 1` sentinel for the recursive
  function body.
- `setSource` associates semantic nodes with values for fingerprint exception
  diagnostics only in `CallStackTool`; `FastTool` and `DebugTool` leave values
  untouched. Attachment overwrites earlier sources and occurs at Java's explicit
  constructor, set-operation, `DOMAIN`/`UNION`, operator-argument, and subset
  assignment sites. Lambda source attachment precedes materialization. Replay's
  later lambda/predicate copy constructors deliberately drop the source.
- `CallStackTool` freezes only for TLC runtime, evaluation, and fingerprint
  exceptions; unrelated errors unwind and pop normally. Value-side exception
  wrapping remains a separate porting concern from source attachment. Primitive,
  explicit-set, tuple, record, finite-function, lazy-value, set-constructor,
  set-operation, model/special, and operator boundaries are ported. Returned
  errors and panics both preserve the concrete receiver in the value chain.
  Detailed checked/unchecked nested string APIs still need audit.
- Set-predicate membership has a separate inner EvalException-only rewrite;
  domain, binding, and predicate evaluation are inside that catch, and ordinary
  runtime/fingerprint failures pass to its outer source wrapper. Membership uses
  the no-cost-model evaluation overload; enumeration retains its model.
  Checked predicate printing catches expansion failures before symbolic fallback.
  Lazy deep normalization is Java's inherited no-op, and direct lazy evaluation
  uses the stored cost model.
- Shared operator methods retain the concrete receiver for diagnostics, source
  wrapping, and copying. Priority evaluating wrappers own their source/model
  and retain primary method metadata while their handle list is stably sorted.
  Method/evaluating initialization skips unsupported fingerprinting.
- MethodValue invocation preserves direct evaluation exceptions and wraps other
  failures at the Java method boundary before its outer source-aware catch.
  Evaluating/priority wrappers rewrite invocation and pure-fallback failures
  through a broad catch with no outer source wrapper; callable wrappers put
  argument evaluation outside their broad invocation/state-assignment catch.
  TLCError.Runtime distinguishes those runtime override failures from legacy
  native EvalException carriers, so predicate membership does not rewrite them.
- Fingerprint exception traces read each value's current source at trace time;
  sources are not snapshots taken when an exception head is created. Null
  source entries remain present in `asTrace` like Java.
- `FingerprintException.getTrace` is intentionally recursive: Java assigns
  labels while walking the linked exception head, then prints the recursive tail
  before the current frame. Preserve that order because the formatted trace is
  user-facing diagnostic text.
- Errors are not generic exceptions; they carry TLC error codes and source
  context.

Port guidance:

- Implement the evaluator against the Go SANY semantic tree, not against raw
  parse nodes.
- Keep a mechanical switch structure close to Java's `evalApplImpl` until the
  test suite passes.
- Preserve short-circuiting for boolean operators.
- Preserve Java's eager/lazy choices. Lazy values are not optional; they avoid
  explosive enumeration and support recursive/function semantics.
- Preserve Java's special `OPCODE_fa` branch: function records and function
  lambdas evaluate `args[1]`, while tuples and records first reject
  `f[e1, ... , eN]` when `N > 1`. The tuple/record argument must not be
  evaluated before that arity check.
- The initial-state, next-state, and `ENABLED` interpreters have their own
  `OPCODE_fa` paths. When the function expression evaluates to a
  `FcnLambdaValue` without a materialized function record, Java calls
  `getFcnContext` and recurses into the lambda body under that argument-bound
  context. This preserves symbolic state generation for predicates hidden
  behind function application.
- `getFcnContext` has its own binding and diagnostics algorithm; it must not
  delegate to `FcnLambdaValue` selection/application helpers. Java checks each
  parameter domain before converting tuple arguments, reports the relevant
  argument number and function-expression source, ignores excess arguments,
  and directly indexes short argument lists. Its single tuple-mismatch error
  prints `this.toString()`: object identity for ordinary tools, and the
  overridden stack string for `CallStackTool`. Lazy-function formatting must swallow
  expansion failures so diagnostic printing can fall back to symbolic text.
- In `evalApplImpl`, if a looked-up `LazyValue` is forced with `s1 == null`,
  Java evaluates the lazy expression directly with the lazy value's saved
  context and cost model. The cached path is reserved for the `s1 != null`
  branch through `LazyValue.getValue`.
- `tlc2.value.impl.LazyValue.off=true` disables LazyValue caching in Java by
  constructing every lazy value with the `UndefValue` sentinel. Go mirrors this
  through the same Java-style property key so lazy expressions still exist as
  thunks but never cache evaluated values.
- Preserve exact undefined-value behavior. TLC distinguishes "not enumerable",
  "undefined", "not comparable", and ordinary false in user-visible ways.

## Init and Next State Generation

Initial state generation flows through:

- `Tool.getInitStates()`
- `Tool.getInitStates(IStateFunctor)`
- `Tool.makeState`
- `Tool.getInitStatesAppl`
- action item lists that delay assignments and constraints.

Next-state generation flows through:

- `Tool.getNextStates(Action, TLCState)`
- `Tool.getNextStates(INextStateFunctor, TLCState[, Action])`
- `getNextStatesImpl`
- `getNextStatesApplImpl`
- `processUnchanged`
- assignment/action item list evaluation.

Go represents Java's `IStateFunctor` and `INextStateFunctor` with concrete
callback structs, not interfaces. Their unsupported operations should remain
explicit errors: Java has no default for `IStateFunctor.addElement` or
`INextStateFunctor.addElement(predecessor, action, successor)`, while
`setElement`, plain next-state `addElement`, and `hasStates` are default
unsupported operations.

Key semantics:

- State variables are assigned by primed variable equalities.
- Partial successor states are legal during construction.
- `UNCHANGED` copies values from current to successor.
- `ENABLED` uses a special context branch/cutoff mechanism.
- Java represents delayed conjuncts with positive action-list `kind` values,
  where the integer is the original conjunct position. The special sentinels
  are `0` for the base conjunct marker, `-1` for plain predicates, `-2` for
  `UNCHANGED`, and `-3` for changed expressions. Keep these numeric values
  exact rather than replacing them with an ordinary Go enum.
- Java's `ActionItemListExt` subclasses `ActionItemList` only to carry an
  action and a previous-node pointer across `cdr()` traversal. Go folds this
  into the single concrete `ActionItemList` struct with `act` and `prev` fields
  to avoid another type while preserving `getAction()` behavior.
- `ActionItemList.cons` must mirror Java's `ActionItemList.coverage` guard: it
  descends from the parent cost model with `cm.get(pred)` only when action
  coverage is enabled. When coverage is disabled, or only variable coverage is
  enabled through the coverage bitmask, it keeps the incoming cost model.
- Preserve Java's next-state action-list decomposition inside the port:
  public `GetNextStatesFromActionList` is the coverage wrapper,
  `getNextStates0` is the ordinary recursive dispatcher, and
  `getNextStatesAllAssigned` is the `TLCGlobals.warn && s1.allAssigned()`
  branch that evaluates remaining predicates directly while routing
  `UNCHANGED` and unsatisfied predicates through the same functor hooks as Java.
- Java recognizes `OPCODE_cdot` for action composition but disables it by
  default behind the `tlc2.tool.impl.Tool.cdot` system property. The Go port
  keeps the same default via `Globals.Cdot == false` and the same property key
  through `-Dtlc2.tool.impl.Tool.cdot=true` or `TLCSet`. It returns Java's
  unsupported-action-composition message from evaluation, next-state
  generation, and `ENABLED`. When explicitly enabled, `A \cdot B` first
  collects intermediate `s -A-> t` states into a concrete `StateVec`, then
  evaluates or generates `t -B-> u`. The next-state path uses `s0.CopyWith(s1)`
  for Java's partial intermediate state, forwards composed successors through a
  concrete `NextStateFunctor`, and stamps the original predecessor/action on
  `u`. The predicate-evaluation path intentionally starts from a fresh empty
  intermediate state, matching Java's `TLCState.Empty.createEmpty()` comment.
- Java `-generate` sets `tlc2.tool.impl.Tool.probabilistic=true`. In that
  mode, next-state generation randomizes disjunction order with the simulator
  RNG's `nextDouble` start index and `nextPrime` stride, randomizes bounded
  existential and assignment enumeration, and returns as soon as the next-state
  functor reports that a successor exists. If no successor is found in the
  randomized pass for assignments, Java falls back to ordinary sequential
  enumeration. `CASE` in the next-state relation remains deliberately
  unsupported in probabilistic mode.
- A complete state must assign every declared variable.
- `isGoodState` detects incomplete or illegal states.
- Model constraints and action constraints filter states but do not replace
  invariant/action-property checks.

The model checker's init path uses a `DoInitFunctor` to avoid materializing a
large `StateVec` of all initial states. Each init state is checked and inserted
as it is generated. Fresh runs print `TLC_COMPUTING_INIT` before the functor is
invoked and then print `TLC_INIT_GENERATED1` or `TLC_INIT_GENERATED2` with
Java's generated-state pluralization and distinct-state-count rule.
Successful model-checking runs print `TLC_SUCCESS` from the checker rather than
from the top-level runner. The payload mirrors Java's collision-probability
reporting: use only the optimistic probability when it is below `1E-10`;
otherwise also compute the observed probability from `FPSet.checkFPs()`. DFID
uses the two-probability path directly, matching Java.

Correctness notes:

- Invariants on init states are checked before queue insertion.
- Implied initial conditions are checked separately.
- Successor states that fail model constraints are not queued, but implied
  action checks may still apply depending on where filtering happens.
- An unseen successor is written to the trace before invariant/implied-action
  checks so a violation can print a trace.

Performance notes:

- Avoid allocating `StateVec` for huge init sets where the Java functor path
  streams states.
- Mutable state copying and partial assignment order are hot.
- Context enumeration for quantifiers and set membership is hot.

## State Representation

### `TLCState`

`TLCState` is an abstract assignment of spec variables to explicit values.
It carries:

- `workerId`: trace fragment owner.
- `uid`: pointer into per-worker trace file.
- `level`: state graph depth, with initial states at level `1`.
- static `vars`: ordered variable declarations.
- static `Empty` and `Null` sentinels.

Core methods:

- `Bind`, `Unbind`, `Lookup`, `ContainsKey`.
- `Copy`, `DeepCopy`, `CreateEmpty`.
- `AllAssigned`, `GetUnassigned`.
- `DeepNormalize`.
- `FingerPrint`.
- `ToString` variants for traces.
- predecessor/action/cached-value hooks used by extended/debug modes.

### `TLCStateMut`

`TLCStateMut` stores values in an array indexed by `UniqueString.varLoc`.
It is optimized for model checking:

- binding is O(1).
- copying is array copy.
- fingerprinting iterates variables in declaration order.
- values are deep-normalized before sharing through queues.
- optional VIEW expression replaces raw state fingerprinting.
- optional symmetry permutations choose the lexicographically smallest
  representative before fingerprinting.
- Java installs the active `Tool` into `TLCStateMut`/`TLCStateMutExt` during
  tool construction so state fingerprints can see the VIEW expression and
  symmetry permutations. The Go port keeps the same static state context
  through `SetTLCStateTool(tool)` and must call it when constructing model
  checkers or simulators. The order is important: first select the symmetry
  representative, then fingerprint either the raw representative values or the
  value produced by evaluating VIEW on that representative.
- Configured `SYMMETRY` is not only a boolean flag. Java evaluates the named
  zero-arity operator from the unprocessed definition table, evaluates its body
  with the processed constant environment, and installs the concrete
  `MVPerms.permutationSubgroup` result before fingerprints are taken. Go mirrors
  this in `SpecProcessor.processConfigSymmetry`, leaving `HasSymmetry` config
  based while `GetSymmetryPerms` returns only the concrete installed subgroup.
  Java also warns with `TLC_SYMMETRY_SET_TOO_SMALL` when a configured symmetry
  set has fewer than two useful model values; the Go port mirrors this for
  direct `Permutations(S)` argument shapes whose constant set value is available
  through the processed tool definition table.
- `TLCStateMut.toString` also honors VIEW, but only when the global `useView`
  flag is enabled; fingerprinting uses VIEW whenever the active tool has one.
- `setPredecessor` is also the level increment path. Java fails with
  `TLC_TRACE_TOO_LONG` when the predecessor is already at `Integer.MAX_VALUE`;
  Go keeps the same 32-bit ceiling even on wider `int` platforms.

### `TLCStateMutExt`

`TLCStateMutExt` is mostly a copy of `TLCStateMut` with extra fields for:

- predecessor,
- action,
- callable,
- cached values.

It is used for simulation, executor, debugger, and modes needing richer state
metadata. The Go port should avoid duplicating the whole implementation if it
can do so without changing behavior, but should preserve the two operational
modes.

Tricky details:

- State equality in Java is value-wise but `hashCode` is intentionally absent
  for mutable states. Do not place mutable states in Go maps keyed by struct.
- Fingerprint uses the configured VIEW and symmetry. This is state identity for
  visited-state purposes.
- Symmetry reduction is expensive: all permutations are applied to find the
  lexicographically smallest representative.
- `TLCGet("level")` depends on exact level initialization.

### `StateVec`

`StateVec` is Java's mutable array-backed state vector and implements both the
plain state functor and next-state functor contracts. The Go port keeps it as a
single concrete slice-backed struct in `state.go`:

- `AddElement(state)` appends the state unchanged.
- `AddNextElement(predecessor, action, state)` stamps predecessor/action before
  appending, matching Java's `INextStateFunctor.addElement`.
- `AddElements` returns the larger vector as the receiver of the append, just as
  Java swaps `s0`/`s1` to reduce copying.
- `Remove` swaps in the last element; `RemoveAt`/`Replace` only overwrite the
  slot. These two removal styles are intentionally distinct.
- `Reset` and `Clear` both set logical size to zero while keeping capacity.
- `Copy` calls state `copy`; `DeepCopy` calls state `deepCopy`.
- `ToRecords(append)` returns all current states as record values followed by
  the appended state.
- `ToRecordsFrom(from, append)` walks backward until `from`'s fingerprint,
  then returns that suffix in forward order followed by `append`, matching
  Java's linked-list `push` behavior.

## Value System

`tlc2.value.impl.Value` is the base class for all TLA+ runtime values.

Value kind constants:

- `BOOLVALUE`
- `INTVALUE`
- `REALVALUE`
- `STRINGVALUE`
- `RECORDVALUE`
- `SETENUMVALUE`
- `SETPREDVALUE`
- `TUPLEVALUE`
- `FCNLAMBDAVALUE`
- `FCNRCDVALUE`
- `OPLAMBDAVALUE`
- `OPRCDVALUE`
- `METHODVALUE`
- `SETOFFCNSVALUE`
- `SETOFRCDSVALUE`
- `SETOFTUPLESVALUE`
- `SUBSETVALUE`
- `SETDIFFVALUE`
- `SETCAPVALUE`
- `SETCUPVALUE`
- `UNIONVALUE`
- `MODELVALUE`
- `USERVALUE`
- `INTERVALVALUE`
- `UNDEFVALUE`
- `LAZYVALUE`
- `DUMMYVALUE`

Common operations:

- `Kind`
- `CompareTo`
- `Equals`
- `Member`
- `IsFinite`
- `Size`
- `Normalize`
- `DeepNormalize`
- `IsNormalized`
- `IsDefined`
- `DeepCopy`
- `FingerPrint`
- `Permute`
- `TakeExcept`
- `Apply`
- `Select`
- `Elements`
- `RandomElement`
- `ToString`
- binary serialization/deserialization.

Java value printing has two layers. `Value.toStringImpl` creates a buffer,
invokes the concrete printer, appends a delimiter, and catches sourced runtime
failures. Nested values call the concrete buffer overload with the same offset
and checked/unchecked flag. Go represents this with `ValueToString` and concrete
`ToString(*strings.Builder, int, bool)` methods; `StringWithDelimiter` and
`StringUnchecked` expose the Java string overloads without expanding the required
`Value` interface for foreign Go implementations.

Lazy printers preserve Java's individual `catch(Throwable)` regions. Cup, cap,
difference, and predicate-set expansion catch both materialization and printing;
product printers catch materialization and print the resulting set outside that
catch. Subset printers catch the size decision only, while `UNION` propagates
expansion failures. Unchecked calls rethrow instead of taking symbolic fallback.
All these paths share the original buffer and retain partial output after a
caught failure. Function lambdas always swallow expansion/printing failures and
request checked printing of a materialized function, including from unchecked
callers. Predicate-set symbolic domains and function-parameter domains invoke
standalone checked public string conversion. `TLC!Print`, `PrintT`, and `ToString`
use unchecked conversion; file `PrintT` appends its newline before pretty printing.

Important concrete values:

- `BoolValue`: singleton true/false.
- `IntValue`: cached small-ish integer values via factory.
- `StringValue`: backed by interned `UniqueString`.
- `TupleValue`: function from `1..n` to values.
- `RecordValue`: sorted/normalized field names and values.
- `SetEnumValue`: explicit finite set with normalization and duplicate
  handling.
- `IntervalValue`: finite integer interval without materializing all elements.
- Java's `Reducible` set operations live on concrete `IntervalValue` and
  `SetEnumValue` methods in Go. Difference reduces only its left operand;
  intersection and union prefer the left reducible operand, then the right.
  These methods use membership checks instead of enumerating the other side.
  Explicit-set results preserve raw left order and duplicates; union normalizes
  its reducible right operand by enumerating it and leaves the result
  unnormalized. Empty reducible unions return the other operand directly.
- `BaseValue.CM` mirrors Java's shared `Value.cm`. Evaluating/lambda/predicate
  values use that same field. Primitive and lazy set iterators preserve Java's
  secondary-count boundaries: explicit sets count calls after exhaustion,
  intervals count returned elements, filters count examined candidates, and
  materialization additionally counts produced elements. `UNION` uses the
  concrete optimization helper and inherits its input set's cost model.
- Product enumeration counts generated fields before advancing and attaches
  the original model to each tuple, record, or function. A function set with
  an empty domain counts its sole empty function once. Normalized powerset
  and k-subset enumeration attaches models without counting generated fields;
  materialization validates the full cardinality first and counts the final
  set's elements. Empty powerset bases produce the empty subset before trying
  set conversion. A null conversion result must remain distinct from a
  materialization failure.
- Java conversion accounting has asymmetric branches: interval-domain
  `FcnRcdValue.toTuple` constructs a tuple without the model or an increment,
  while explicit-domain `FcnLambdaValue.toTuple` increments even with global
  coverage disabled. Lazy-function materialization uses raw tuple casts and
  fixed declared-size arrays, installs the cached function with its model,
  then increments the materialization count before applying EXCEPT updates.
- Pointer-returning function conversions preserve Java exceptions using typed
  TLC error panics when a Go error return is unavailable. Normalization, size,
  selection, and materialization failures must not become null conversion
  results. `Tool.Eval`/`EvalAppl` turn those typed panics into returned errors;
  other runtime panics propagate. Eager construction fails when materialization
  fails instead of falling back to a lazy function.
- `FcnRcdValue`: explicit finite function, optimized for interval domains.
  Java's `tlc2.value.impl.FcnRcdValue.threshold` controls when normalized
  finite function records switch from linear lookup to binary search; Go keeps
  the default `32` and reads the same property key.
- `FcnLambdaValue`: lazy function with params/body/tool/context.
- Function constructors stay lazy unless Java's `Tool.evalApplImpl` can prove
  every bounded domain is `Reducible` and `EvalControl.KeepLazy` is clear. In
  Go, mirror that check concretely with `IntervalValue` and `SetEnumValue`
  rather than adding a one-use `Reducible` interface. Recursive functions also
  stay lazy. `DOMAIN` on `FcnLambdaValue` reads the lambda's parameter-domain
  metadata through `GetDomain`, not by materializing the whole function into a
  `FcnRcdValue`.
- `SetOfFcnsValue`, `SetOfRcdsValue`, `SetOfTuplesValue`: lazy enumerable set
  spaces.
- `SubsetValue` and `KSubsetValue`: lazy subset enumeration and unranking.
  Go's `SubsetUnrank` keeps strict predecessor lookups in a sorted cutoff
  vector, including replacement of duplicate binomial keys, and retains Java's
  Pascal-table indices and exact long conversion. The integer random-subset
  overload is `GetRandomSetOfSubsetsUpTo`; it uses decimal scale 32/HALF_DOWN
  allocation and Java's signed-long stride/offset behavior, including zero-
  modulus failures and random draws after the output bound has been reached.
  Fixed-cardinality `KElementEnumeration` requires counts to fit an int, while
  `NumberOfKElements` requires a long. Normalized powerset traversal retains
  its separate combinations iterator. Lexicographic bitset traversal replaces
  the outer base with its enumerated set before normalization and counts the
  sizes of emitted subsets for coverage. Empty traversal returns a fresh set
  after each reset. These inherited helpers also operate on Go k-subsets.
  `KSubsetValue` keeps Java's special comparison/equality shortcuts against
  ordinary `SUBSET S`: finite cardinalities are compared without materializing
  powersets, and `kSubset(0, S)` equals `SUBSET {}` independent of `S`.
- Lazy set constructors and set operations preserve Java's
  `SetEnumValue.DummyEnum` state after `deepNormalize` with explicit Go dummy
  flags. A dummy cache means "deep-normalized but not materialized"; when later
  converted to `SetEnumValue`, the materialized set must be deep-normalized
  before the dummy flag is cleared.
- `SetPredValue`: predicate-filtered set. `Tool` eagerly evaluates
  `{x \\in S : P(x)}` only when `S` is Java `Reducible`, represented by
  `IntervalValue` and `SetEnumValue` in Go. Other values retain a lazy
  `SetPredValue` even when they support enumeration; this avoids premature
  predicate evaluation and materialization of powersets or product spaces.
- `LazyValue` and `EvaluatingValue`: deferred evaluation.
- `ModelValue`: named atoms with special comparison/permutation behavior.
- `UserValue`: module-defined values such as unbounded standard sets.
- `UndefValue`: explicit undefined marker.

Value stream details:

- Java's `ValueOutputStream` handle table is a format-level compression
  detail, not a generic "all values get IDs" rule. Only the concrete `write`
  methods that call `vos.put` can later emit `DUMMYVALUE`: `StringValue`,
  `TupleValue`, `SetEnumValue`, `FcnRcdValue`, `RecordValue`, and record field
  `UniqueString` names. `BoolValue`, `IntValue`, `IntervalValue`, and
  `ModelValue` always write themselves directly.
- `ModelValue` serializes as kind `MODELVALUE` followed by the short index into
  the already-initialized global `ModelValue.mvs` table. The stream reader does
  not build that table on demand; a null table or invalid index retains Java's
  runtime failure rather than becoming an I/O error.
- Compound lengths follow Java's encodings: tuples and function records use
  compact naturals, set enumerations and records use signed lengths to preserve
  normalizedness, and function records write an interval-domain marker byte
  before either interval bounds plus values or explicit domain/value pairs.

- Output handles retain the actual objects for the stream lifetime, as Java's
  `Object[]` does, and preserve reference identity including null. Go pointer
  addresses alone are insufficient because reclaimed addresses can be reused.
  Input handles reserve sequential indices in a table starting at capacity 16;
  assignment/lookup use the allocated bounds and null entries remain null.
- Java uses the same write format for checkpoints and external values.
  `WriteExternal` is a Go convenience alias; external reads discard saved
  UniqueString token/location metadata, re-intern the text, and leave records
  unnormalized. Compact natural writes preserve Java's comment-only
  non-negative precondition without adding validation branches.
- Cached lazy-value writes do not materialize values. Null caches throw NPE,
  while a dummy set cache writes its shared sentinel handle and kind byte before
  failing on its null ValueVec. Unsupported values/kinds use Java's
  WrongInvocationException, and unknown kind bytes print as signed Java bytes.
- File-stream truncation produces a null-message EOFException, with Go
  `errors.Is(err, io.EOF)` compatibility for recovery loops. Invalid array sizes,
  indices, and handle casts retain runtime exception types. Queue byte input
  uses no handles (`getIndex=-1`, assignment is a no-op), rejects reference
  records, and throws array-bounds failures on truncated state bytes.

Correctness notes:

- Value comparison order is semantic and must match Java for normalization,
  symmetry representatives, sorted set/record/function printing, and
  fingerprinting.
- `Normalize` is often mutating in Java. Go can choose immutable or mutable
  internals, but must preserve sharing assumptions and performance.
- Lazy enumerable values must not enumerate huge spaces prematurely.
- `isEmpty` has special mathematical cases for function sets, record sets,
  tuple sets, subsets, k-subsets, and user values.
- Fingerprint exception wrapping uses value source metadata. Preserve enough
  metadata to emit equivalent diagnostics.

Performance notes:

- The value system is the heart of TLC's runtime. Avoid interface dispatch in
  very small arithmetic/boolean paths if it becomes measurable, but first keep
  behavior obvious and close to Java.
- Enumeration should be streaming where Java streams.
- Preserve interval/function/set lazy representations to avoid catastrophic
  memory growth.

## Contexts and Enumerators

`Context` is a linked list of `(SymbolNode, value)` pairs:

- `Empty` is the base context.
- `branch(base)` inserts a null-name marker used by `ENABLED`.
- `lookup(var, cutoff)` stops at a branch marker when cutoff is true.
- `cons(name, value)` prepends a binding.

Context is used for:

- level boundedness during processing.
- runtime lexical/operator bindings during evaluation.
- quantified variables and function parameters.
- enabledness branch scoping.

`ContextEnumerator` enumerates bindings for bounded quantifiers and set-based
operator parameters. Enumeration order is user-visible in traces, random
simulation, and sometimes coverage, so preserve Java order.

## Utility Collections

`tlc2.util` contains deliberately small custom collections and IO helpers used
by values, queues, traces, liveness graphs, and fingerprint sets.

Porting guidance:

- Keep deterministic iteration order for any structure whose order reaches
  fingerprints, traces, XML/JSON output, or diagnostics.
- Java's small hash tables (`ObjLongTable`, `LongObjTable`, and the
  semantic-node variant used by primed-location coverage) expose probe-slot
  effects through returned indices and slot-order enumeration. Do not replace
  those paths with ordinary Go maps; mirror Java's `count/length/thresh`
  arrays, linear probing, `2*length+1` growth, and physical-slot scans.
- Prefer concrete Go structs. Java's `ExternalSortable` has `BigInt` as its
  practical TLC implementation, so the Go external-sort helpers operate on
  `[]*BigInt` directly instead of creating a one-implementation interface.
- Do not create public pseudo-abstract base structs whose methods rely on Go
  method dispatch that does not exist. Java abstract bases such as
  `ObjectStack` should become private shared fields plus concrete methods on
  `MemObjectStack` and `DiskObjectStack`.
- Port utility tests early. These are cheap, stable conformance checks and do
  not disturb the mechanical core model-checker port.

## Fingerprinting

`tlc2.util.FP64` implements 64-bit fingerprints over GF(2^64):

- `New()` starts from selected irreducible polynomial.
- `Extend` handles strings, chars, bytes, ints, and longs.
- `Polys` contains selectable polynomial constants.
- `TLC` selects `fpIndex` unless user specifies it.

Heap-backed disk fingerprint sets allocate Java-style striped read/write locks
at construction. The default stripe count is
`2^(floor(log2(NumWorkers)) + 8)`, with `tlc2.tool.fp.DiskFPSet.logLockCnt`
or `TLAGO_DISK_FPSET_LOG_LOCK_CNT` overriding the exponent. Go `Put` and
`Contains` use those stripes, while flushing acquires all stripes. Java's
flusher recursively acquires its already-held write lock; Go skips that stripe
during the all-stripe acquisition because its mutexes are not reentrant.
Disk fingerprint reads mirror Java `IdThread.GetId(braf.length)`: a goroutine
with a current worker id uses its corresponding fixed `BufferedRandomAccessFile`
reader; calls outside worker scope fall back to the reader pool.

Trace level is trace-authoritative when extending partial state spaces:
`CheckImpl.makeStateSpace` uses `TLCTrace.getLevel(state.uid) + depth` in Java,
so Go's `CheckImpl.MakeStateSpace` uses `TLCTrace.GetLevel(state.UID)` rather
than trusting the mutable state's cached level or mirrored object identity.

Visited-state identity is a 64-bit fingerprint of:

- normalized state values in declared variable order,
- or VIEW value when a view is configured,
- after symmetry representative reduction when symmetry is configured.

Correctness requirements:

- Go FP64 must match Java exactly for all primitive extension functions,
  including Java signed-byte and unsigned-shift behavior.
- Every `Value.FingerPrint` method must emit the same extension sequence as
  Java.
- State-level fingerprinting must normalize the same values at the same points.
- Collision handling is probabilistic just like Java TLC. Do not "improve" it
  by using full state keys unless it is explicitly a debugging mode.

## Fingerprint Sets

`FPSet` is the abstract concurrent set of 64-bit fingerprints. `put(fp)` returns
true if the fingerprint was already present and false if it was newly inserted.

Implementations:

- `MemFPSet`: synchronized in-memory hash table of buckets. Rehashes by
  doubling table capacity and splitting buckets by one bit.
- `MemFPSet1`: deprecated memory variant backed by `SetOfLong` with Java's
  open-addressing table and checkpoint format.
- `MemFPSet2`: deprecated memory variant with a `2^24` spine. The low 24 bits
  are encoded by the bucket index and only the five higher bytes are stored in
  each collision bucket.
- `DiskFPSet`: bounded memory plus sorted disk backing file with per-worker
  buffered random-access readers and reader/writer locking.
- `MSBDiskFPSet`, `LSBDiskFPSet`, `HeapBasedDiskFPSet`,
  `OffHeapDiskFPSet`: disk/off-heap variants with different indexing choices.
- `MultiFPSet`: partitions fingerprints by high-order bits over nested FPSets.
- `NoopFPSet`: testing/no-op behavior.
- distributed wrappers and managers for RMI-based distributed TLC.
- `OffHeapDiskFPSet` uses open addressing over `LongArray` and chooses one of
  three indexers: bit-shifting for power-of-two position counts, a multiply-high
  1024MiB-multiple indexer, or an exact/infinite-precision fallback. The Go port
  keeps these as one concrete `OffHeapIndexer` with a kind field.
- In the off-heap primary table, `0` means empty, a positive fingerprint means
  not yet evicted, and the same value with the high bit set means already
  evicted to the sorted disk file. Flushed slots still detect duplicates and can
  be reused for different fingerprints.

Factory/configuration behavior:

- Java property `tlc2.tool.fp.FPSet.impl` selects the implementation class;
  the Go port reads that environment key plus `TLAGO_FPSET_IMPL`.
- `getImplementations()` advertises `MSBDiskFPSet`, `LSBDiskFPSet`, and
  `OffHeapDiskFPSet`, with `MSBDiskFPSet` as the default.
- VM argument recommendations split on storage type: heap-based sets use
  `-Xmx`, while off-heap sets use `-XX:MaxDirectMemorySize`.
- `FPSetConfiguration.getMemoryInBytes()` applies the ratio to explicit memory
  too, matching Java's call to `TLCRuntime.getFPMemSize(memoryInBytes * ratio)`.
- `MultiFPSet` creates nested sets through a Go `NewMultiFPSetConfiguration`
  helper corresponding to Java `MultiFPSetConfiguration`: it copies the parent
  config, disables nesting, and divides memory across `2^fpBits` children.

Checkpointing:

- `beginChkpt`: write temporary snapshot.
- `commitChkpt`: atomically rename tmp to checkpoint.
- `recover`: load checkpoint or rebuild from trace.
- `recoverFP`: insert a recovered fingerprint and assert it was absent.
- `DiskFPSet` no-argument checkpoint methods are silent no-ops because Java
  rebuilds disk-backed fingerprints from the trace. `MultiFPSet`'s no-argument
  checkpoint methods forward to child no-argument methods, so they stay silent
  for disk-backed children. `NonCheckpointableDiskFPSet` only warns on the named
  overloads used by nesting/distributed surfaces; the warning says checkpointing
  is not implemented for the concrete Java class name.

Port guidance:

- Start with `MemFPSet` and `MultiFPSet`, then disk variants. This is an
  implementation staging order, not a semantic excuse.
- Match `put` return polarity exactly.
- Preserve thread-safety.
- Preserve checkpoint file semantics for later Java test parity.
- Disk/off-heap implementations matter for performance and large models; do not
  permanently replace them with maps.
- `DiskFPSet` keeps persistent `BufferedRandomAccessFile` readers over the
  backing `.fp` file. After a flush merges into a temporary file and replaces
  the old backing file, all readers must be closed and reopened, otherwise
  lookups can keep reading the old file handle.

## State Queues

`IStateQueue` abstracts the frontier of unexplored states.

Implementations:

- `StateQueue`: synchronization and worker coordination base class.
- `MemStateQueue`: in-memory queue.
- `StateDeque`: deque variant; Java adds every state to the front and polls
  from the front, so this is LIFO search order and it does not support
  checkpointing.
- `DiskStateQueue`: two in-memory buffers plus state-pool disk files.
- `DiskByteArrayQueue`: byte-array backed disk queue that serializes states
  before taking the queue lock and stores raw state bytes in its disk buffers.
  Java gives this queue the same reader/writer/cleaner choreography as
  `DiskStateQueue`; the Go port keeps a concrete byte-array cleaner rather
  than generalizing through an interface.
- `SynchronousDiskIntStack`: real disk-backed integer stack used by utility
  code. Java's `DiskIntStack` is documented as an unused asynchronous sketch.

`StateQueue` behavior:

- Queue selection is normally `DiskStateQueue`. Java property
  `tlc2.tool.queue.IStateQueue` selects `MemStateQueue`, `StateDeque`, or
  `DiskByteArrayQueue`; the Go port reads that environment key and the
  command-line-friendly alias `TLAGO_STATE_QUEUE`. The legacy Java boolean
  `tlc2.tool.ModelChecker.BAQueue` is mirrored by
  `TLAGO_MODEL_CHECKER_BAQUEUE`.
- `sEnqueue` adds states and wakes waiting workers.
- The array/slice enqueue overload enqueues every entry and does not filter
  `null`/`nil`; callers are expected to pass dense state arrays. The `StateVec`
  overload is different and skips nil elements. Keep this distinction even
  though it is easy to accidentally collapse the two in Go.
- `sDequeue` blocks when empty until work appears or all workers are waiting.
- `finishAll` terminates all workers and wakes main/checkpoint waiters.
- `suspendAll` stops workers at a barrier for checkpointing.
- `resumeAll` releases the checkpoint barrier.
- `resumeAllStuck` handles distributed worker death cases.

Important synchronization details:

- `numWaiting >= numWorkers` and empty queue means no work remains.
- `numWaiting >= numWorkers` while `stop` is true but the queue is not empty
  means checkpoint suspension has reached its barrier. Java notifies the
  separate `mu` monitor in this case; the Go queues broadcast on their concrete
  condition variable so `SuspendAll` can return and checkpoint the still-live
  queue.
- `finish` is volatile in Java to avoid checkpoint race deadlocks.
- `suspendAll` uses a second monitor `mu` to coordinate checkpoint waiting.
- Lock ordering is deliberate. Preserve it when translating to Go mutex/cond.

Disk queue behavior:

- `enqBuf` fills then spills to disk through `StatePoolWriter`.
- `deqBuf` drains then refills through `StatePoolReader`, which can return a
  prefetched full buffer, synchronously read a pending pool file, or fall back
  to the in-memory `enqBuf`.
- State-pool files are full-buffer writes. Java calls `TLCState.write` on every
  pool slot and therefore fails on a null entry; Go must not substitute empty
  states for nil slots because that hides queue corruption.
- A cleaner thread deletes old pool files once `loPool - lastLoPool > 100`.
  Checkpointing stops the cleaner before writing queue metadata, and after the
  first checkpoint the checkpoint commit path owns obsolete pool-file deletion.
- `DiskByteArrayQueue` uses the same cleaner threshold and checkpoint handoff
  for liveness/raw-state pool files.
- checkpoint includes queue metadata and buffered states.

Port guidance:

- Start with a faithful condition-variable based `StateQueue`.
- Recreate the Java blocking/termination semantics before optimizing.
- Add disk queue once value/state binary serialization exists.

## Model Checking Algorithm

`ModelChecker` extends `AbstractChecker`.

Main fields:

- `theFPSet`: reachable fingerprints.
- `theStateQueue`: frontier.
- `trace`: concurrent trace for error reconstruction.
- `workers`: exploration workers. Java allocates these in the constructor,
  before initial-state generation: worker 0 keeps the normal/debug-capable tool
  and later workers use `tool.noDebug()`.
- `liveCheck`: liveness subsystem. Java `AbstractChecker` creates
  `NoOpLiveCheck` only when `tool.livenessIsTrue()` is true; otherwise it
  warns for liveness plus symmetry and constructs `LiveCheck`, whose
  constructor immediately calls `Liveness.processLiveness(tool)`. The Go port
  mirrors this in `NewModelChecker` with a concrete `LiveCheck`; callers may
  still inject a prebuilt `LiveCheck` through options for tests.
- `errState`, `predErrState`, `errorCode`, `done`, `keepCallStack`.

High-level flow in `modelCheckImpl`:

1. Attempt recovery from checkpoint.
2. If starting fresh:
   - check assumptions,
   - compute initial states through `doInit`,
   - check init invariants and implied init conditions,
   - insert good in-model init states into FP set, queue, trace, liveness.
3. If no actions exist:
   - success when queue is empty,
   - otherwise error for states with no next action.
   - Java calls `cleanup(true)` in both no-action cases, even when the queue is
     non-empty and the result is `TLC_STATES_AND_NO_NEXT_ACTION`; Go preserves
     that single success-cleanup call instead of using the returned error code
     to keep artifacts.
4. Run worker exploration with `runTLC`, which starts all pre-created workers,
   periodically performs coordinator work while they run, and joins them after
   the shared queue reaches completion or an error.
5. During exploration:
   - workers dequeue states,
   - generate all action successors,
   - detect incomplete states,
   - apply model/action constraints,
   - insert unseen fingerprints,
   - write trace before invariant/implied action checks,
   - enqueue only unseen in-model non-violating states,
   - add behavior graph edges for liveness,
   - detect deadlock.
6. Run final liveness check if needed.
7. Check postconditions.
8. Print success/error summary, coverage, statistics.
9. Cleanup closes the FP set, trace, liveness checker, and state writer, then
   deletes metadata when configured and not vetoed by checkpoint-preservation
   rules.

Java wraps `modelCheckImpl` with `AbstractChecker.modelCheck`, which returns
the stored `errorCode` when the implementation result is `NO_ERROR`. This is
observable after kept-call-stack worker failures because `runTLC` deliberately
returns `NO_ERROR` so the outer code can replay the error. Go keeps the same
wrapper precedence after the cleanup decision, so cleanup still sees the
implementation result Java would have used.

`doNext` error precedence:

- incomplete successor state,
- unseen successor invariant violation,
- implied action violation,
- deadlock after all actions produce no successors,
- evaluation exceptions with call-stack replay where configured.

Continuation mode:

- Some invariant/action violations can be printed and exploration continues.
- Without continuation, first error terminates all workers.

Port guidance:

- Implement worker/error synchronization carefully. Only one worker should own
  the primary error trace unless continuation mode says otherwise.
- Java holds the checker monitor across next-state error acceptance, trace
  printing, and worker error-time postconditions, and across continuation-mode
  violation printing. Go serializes those reports with `ModelChecker.nextErrorMu`
  while using the existing state mutex for checker fields. The separate report
  mutex lets ALIAS and postconditions query/control the checker without requiring
  reentrant Go mutexes. The postcondition wrapper calls a lock-held error helper
  and reads `Done` under the state mutex before deciding the counterexample's
  `console` field.
- Preserve Java's `AbstractChecker.runTLC` coordinator role: workers do state
  generation, while the checker thread periodically suspends the queue for
  liveness/checkpoint work and then resumes or finishes the workers.
- Positive `runTLC(depth)` limits are enforced by that coordinator after
  progress reporting: if the reported level exceeds `depth`, it finishes the
  shared state queue. Workers must not skip next-state generation locally based
  on depth; that would change Java's scheduling, trace, and liveness handoff.
- Periodic coverage reporting is also driven by the coordinator, using
  `coverageInterval / progressInterval` as Java does. This is separate from the
  final coverage report emitted when model checking exits.
- `_PERIODIC` is part of that coordinator loop. Java treats a configured
  periodic expression as sufficient reason to suspend workers, evaluates it
  after any liveness work and before checkpointing, and returns
  `TLC_ASSUMPTION_FALSE` only when the evaluated value is exactly `FALSE`.
- `LiveCheck.doLiveCheck()` only decides whether liveness itself should trigger
  an expensive queue suspension. If checkpointing or `_PERIODIC` already caused
  the suspension, Java still runs `liveCheck.check` whenever liveness is enabled
  and the runtime ratio is below the configured liveness ratio.
- Java's `tlc2.TLC.stopAfter` property is a time-bound escape hatch for both
  model checking and simulation. When configured, a timer calls the active
  checker or simulator stop method. Cleanup keeps explicit checkpoint data if
  unexplored work remains and either an error was found or the run was
  time-bound.
- Simulation keeps `seed` and `aril` as separate user-visible values. Java
  seeds the simulator RNG with `seed`, advances it by `aril` `nextDouble()`
  calls, and seeds `RandomEnumerableValues` with the original `seed`, not
  `seed + aril`. `TLCGet("config")` must report the original pair.
- Java treats seed absence, not seed value zero, as the cue to draw a seed from
  a no-argument `RandomGenerator`. An explicit `-seed 0` must be replayable as
  zero. The no-argument generator follows `java.util.Random()` seeding via the
  `seedUniquifier() ^ nanoTime` pattern; DFID uses the same constructor before
  reseeding its worker RNG from `nextLong()`.
- Java's `tlc2.tool.ModelChecker.vetoCleanup` property forces metadata
  retention even when `-cleanup` was requested; Go mirrors it with the same key
  plus `TLAGO_MODEL_CHECKER_VETO_CLEANUP`.
- Cleanup has one deliberately load-bearing Java quirk: `FileUtil.deleteDir`
  receives the checker cleanup `success` argument as its `recurse` argument.
  When cleanup is called with `success=false`, Go must close checker resources
  but only attempt a non-recursive metadir delete, so a non-empty metadir with
  trace/checkpoint artifacts is preserved for debugging. When cleanup is called
  with `success=true`, Go may recursively remove the metadir. Do not copy
  Java's likely early-return cleanup leak, but do preserve this artifact
  retention rule. The unit test `TestModelCheckerCleanupPreservesFailureArtifactsLikeJava`
  is intentionally small and high-signal because this behavior is easy to
  "simplify" incorrectly.
- Preserve when traces are written relative to checks.
- Preserve generated-state counters versus distinct-state counters.
- Preserve final liveness check behavior even when no safety error occurs.
- Preserve Java's `doNextFailed` keep-call-stack polarity. Generic next-state
  failures call `setErrState(cur, succ, true, ec)` in the ordinary case; Java
  only flips that flag for stack overflow, out-of-memory, and assertion errors.
- Java `Worker.addElement(cur, action, succ)` wraps callback exceptions in a
  `WrappingRuntimeException` carrying `succ`. Go mirrors this with the concrete
  `workerNextStateError` struct so outer worker error handling can pass the
  partially generated successor to `doNextFailed` without adding an interface.
- Java exception catches and `instanceof` classify the thrown object itself;
  they do not inspect its cause chain. Go checker/call-stack/simulator/liveness
  consumers use direct concrete type checks and explicitly handle the embedded
  `LiveCounterExampleException` subclass. Fingerprint causes remain accessible
  for dedicated root-cause diagnostics. A standalone stateful exception does
  not enter the worker's EvalException/TLCRuntimeException liveness replay catch.
- Debugger control exceptions and distributed worker exception reuse also
  match only the directly thrown class. Debugger exception pushes set the
  inherited known flag once; evaluation frames catch only evaluation/runtime
  TLC exceptions. Distributed worker catch coverage includes fingerprinting
  and state routing, preserving Java's current predecessor/successor fields
  and resetting the computing flag even when a Go operation panics.
- Throwable detail messages are represented separately from Go `Error()` by
  `GetMessage() *string`: null messages and empty strings have different
  diagnostic behavior. Constructors capture Go runtime frames for Java-style
  stack rendering, including actual causes and common-frame elision. A
  fingerprint exception's `next` field is a semantic trace link, not the
  Throwable cause printed by `printStackTrace`.
- Nullable parameter arrays survive in exception carriers and
  `Message.NullableParams`. `PrintErrorNullable` preserves recorder metadata
  and Java MP's stop-at-first-null placeholder substitution. Native MethodValue
  alone substitutes a stack trace for a null detail message; evaluating and
  callable overrides keep the null array entry. GENERAL stack suffix control
  uses Java MP's historical `noDebug` startup property independently of `-debug`.
- Simulation invariant/action evaluation catches construct a dedicated worker
  error without setting its fatal exception field. Its property name and caught
  nullable detail message print with the evaluation-failed code, and the
  simulator's non-continuable-error policy stops the run.
- `TLC.process` and the simulator report EvalExceptions through GENERAL; only
  direct TLCRuntimeExceptions preserve their code and parameters there.
  Simulator exceptions overwrite an earlier worker code and stop simulation.
  `TLCError.Runtime` marks explicit Java runtime carriers; existing native
  evaluation carriers and direct system-error carriers remain distinguished.
- Call-stack replay uses a synthetic worker id `4223` like Java, but that
  worker is deliberately not registered with the shared trace. Java's comment
  says replay must not rewrite the trace file while reconstructing states for
  diagnostics. Go preserves this with a concrete `Worker.DisableTraceMirror`
  flag: replay may use its own worker trace context, but it must not append to
  `ModelChecker.Trace` through the normal worker mirror hook.
- Liveness `addNextState` call-stack replay has an additional Java bug path:
  if ordinary liveness evaluation fails but the `CallStackTool` rerun succeeds,
  Java calls `Assert.fail(EC.GENERAL, origExp)` because replay success "should
  never happen." Go preserves that as a general TLC error rather than silently
  reporting the original liveness error code.
- Preserve Java's error-time postcondition behavior: init failures with an
  `errState` call `checkPostConditionWithCounterExample(new CounterExample(errState))`,
  and worker `doNextSetErr` paths build a safety counterexample, evaluate aliases
  over that trace, and call the postcondition hook before returning the safety
  error. Java's direct `ModelChecker.doNext` helper is different: its
  `doNextSetErr` path prints and terminates but does not invoke the
  postcondition hook. Evaluation failures remain on Java's separate
  `doNextEvalFailed` path.

The original DieHard auto-worker binary dump/replay prefix test is timing
sensitive in Java too. Temporary worker scheduling delays reproduce the exact
index-6 empty-big vs pour-big-to-small comparison failure seen in Go. A worker
can explore a longer prefix before another worker claims the shortest violation;
trace constraints reject other successors but do not suppress their invariant
checks. Single-worker replay can therefore stop at an earlier off-trace violation.
Keep the source checks and complete assertions intact. No retries, skips or
production trace changes were added to hide this source limitation.

### CheckImpl and CheckImplFile

`CheckImpl` is Java TLC's implementation checker. It reuses the model-checking
engine to build a partial state space from a supplied starting state, tracks
covered states in `coverSet`, keeps a current-state enumerator, and exports new
traces to states that have not yet been covered by the external implementation.
It is not a normal exhaustive run: the checker repeatedly imports a concrete
trace, verifies that each adjacent state pair is reachable in the spec, marks
the visited states, and emits another trace when uncovered behavior remains.

`CheckImplFile` is the file-backed variant used to communicate with an external
simulation engine. It reads trace inputs named by a prefix plus a monotonically
increasing counter, and writes exported traces using the same prefix with
`_out_` plus a separate output counter. The output format is deliberately plain:
`STATE_n` headers followed by each TLC state's text form.

Java parses input trace files as TLA+ modules and turns operation bodies into
states through `Tool.makeState`. The Go port keeps `CheckImplFile` concrete and
uses a `LoadTraceFunc` hook at this layer so the `tlc` package does not import
the production SANY parser. The root `tlago` package wires that hook with
`NewCheckImplFileTraceLoader`: it parses the exact trace filename rather than
adding `.tla`, converts each root-module operator definition body through the
same SANY-to-TLC bridge used for normal model checking, temporarily installs the
trace definitions for helper-definition lookup, calls `Tool.MakeState`, then
restores the main tool definition table. The command option parser mirrors
Java's `CheckImplFile.main` surface: `-config`, `-deadlock`, `-recover`,
`-workers`, `-depth`, `-trace`, `-coverage`, and the root module; the root CLI
exposes it as `tlago checkimplfile`. On recovery runs, the CLI restores the
`UniqueString` checkpoint before parsing/building the tool and resets FP64 to
polynomial index 0 before construction, matching Java's `CheckImplFile.main`
ordering before `new FastTool(...)`.

## DFID Architecture

`DFIDModelChecker` is the depth-first iterative-deepening checker. Java keeps it
separate from the breadth-first `ModelChecker` and runs only with one worker and
without liveness checking.

Important structures:

- `MemFPIntSet`: fingerprint table that stores both membership and DFID level
  status bits.
- `DFIDWorker`: owns explicit stacks for states, fingerprints, successor
  vectors, and successor fingerprints up to `DFIDMax`.
- The worker randomly chooses an unfinished initial state and then randomly
  chooses unfinished successors from the current depth stack.

Important behavior:

- Each outer iteration increments `FPIntSet`'s global level and searches up to
  that level.
- `doNext` for DFID is a one-step generator. It fills successor vectors with
  states not completed at the current level and returns whether all successors
  were non-leaf.
- DFID initial-state processing records new in-model states before checking
  invariants and implied-init properties, matching Java's array/write/liveness
  ordering. Initial states excluded by model constraints are still checked
  against invariants and implied-init properties because Java leaves their
  `status` as `FPIntSet.NEW`.
- DFID `setErrState` stops the DFID workers after the abstract checker accepts
  the error state.
- Backtracking marks fingerprints leveled. States are marked done when all
  children are done or when a leaf has no new child.
- Keep DFID as worker-plus-stacks, not recursive traversal; recursion diverges
  from Java's scheduling, status updates, and randomization points.

## Worker and Trace Architecture

`Worker` is both a thread and an `INextStateFunctor`.

Worker maximum trace depth is Java's private volatile int. Go stores it in an
atomic Int32 and exposes GetMaxLevel/SetLevel; ConcurrentTLCTrace reporting
loads it through the getter while the worker writes successor trace records.
The generated-state count has one worker writer and concurrent progress readers.
Go stores it in a private atomic Int64, preserving increment and deadlock-check
order, and the checker reads it through GetStatesGenerated.

Worker loop:

1. `sDequeue` current state from the shared blocking queue. A `null`/`nil`
   dequeue means all workers are waiting and the frontier is empty; the worker
   sets checker completion and finishes the queue.
2. Set thread-local current state.
3. Allocate `SetOfStates` if liveness/debug mode needs successor collection.
4. Ask `Tool` to generate successors through the functor callback.
5. Detect deadlock if no successors were generated.
6. Add the stuttering self-loop to the liveness successor set, write the
   `Visualization.STUTTERING` transition to the all-state writer, and add the
   collected successor set to the liveness graph with one
   `LiveCheck.addNextState` call for the predecessor.
7. Record out-degree statistics.

Constrained dumps:

- When a generated successor is excluded by state or action constraints and
  the state writer is in constrained mode, Java rechecks each configured
  constraint and writes one transition per failed constraint with that
  constraint as the DOT reason. Generic not-in-model output is only a fallback
  when no individual reason can be identified.

Trace writing:

- Each worker owns a trace fragment file named by spec and worker id.
- Java increments the worker out-degree counter in `writeState(curState, fp,
  succState)`, immediately after recording an unseen successor in the trace.
  Keep that placement: the counter describes unseen successors discovered while
  evaluating `Next`, even if later invariant or implied-action checks stop the
  run before the state is enqueued.
- Java trace recovery for transition errors treats the stored trace as a prefix:
  `ConcurrentTLCTrace.getTrace(state)` excludes `state` itself except for the
  initial-state special case. Error printing and postcondition counterexamples
  then recover `curState` and `succState` explicitly through `Tool.getState`.
  Preserve the same prefix-plus-explicit-state construction; otherwise
  non-initial transition errors silently drop the predecessor state.
- The Go port now keeps a concrete `ConcurrentTLCTrace` on `ModelChecker`,
  writes Java-style worker fragment files as the authoritative model-checking
  path, and mirrors records into the existing in-memory `TLCTrace` for current
  Go callers.
- Error trace reconstruction follows Java's two-phase model: walk worker
  records backward to collect fingerprints and predecessor locations, then ask
  `Tool.GetState` to regenerate states forward from the initial fingerprint.
- Java backs worker trace fragments, `TLCTrace`, disk FP sets, and bit-vector
  persistence with `BufferedRandomAccessFile`. The port should preserve its
  concrete buffer state (`dirty`, `length`, `curr`, `lo`, `diskPos`, `mark`),
  8K page size, `seeek` page-read signal, and nat encodings.
- Initial state record contains previous pointer `1`, worker id, fingerprint.
- Successor record contains predecessor pointer, predecessor worker id,
  successor fingerprint.
- The successor state receives `(workerId, uid)` for later reconstruction; the
  `workerId` is the generating worker, not the predecessor's worker.
- Reads/writes are synchronized so trace printing sees consistent fragments.

Trace reconstruction and aliasing:

- Java exposes `TraceApp`/`ITool` methods to reconstruct an initial state from
  a fingerprint, reconstruct a successor from a predecessor plus fingerprint,
  and reconstruct transition metadata from a successor/predecessor pair.
- The Go port keeps those duties as concrete `Tool` methods:
  `GetInitState`, `GetStateAfter`, and `GetStateForTransition`; no Go
  `TraceApp` interface is needed.
- Java regenerates states from the init and next-state predicates. The Go port
  first checks the in-memory `KnownStates` registry when present to preserve
  identity for existing trace records, then falls back to Java-style
  regeneration.
- Unnamed actions carry Java's literal `UnnamedAction` sentinel while
  `Action.isNamed()` remains false for that sentinel. Printed
  `TLCStateInfo.info` values are action locations (`<Action ...>` or named
  variants), and unnamed initial actions use `<Initial predicate ...>` rather
  than dropping source context.
- Action locations render all non-nil bound formal parameters with ordinary
  string conversion, matching Java `Action.getLocation`. Counterexample action
  records still use `Action.getParameters`, which includes only parameter
  bindings that are TLC `Value`s. The location portion uses the action
  predicate/body source range, not the operator declaration range. The
  declaration range is retained separately on `OpDefNode` for coverage, matching
  Java `Action.getDeclaration`.
- Base `TLCStateInfo.getStateNumber()` returns the stored trace ordinal.
  Java's `AliasTLCStateInfo` overrides it to return `originalState.getLevel()`;
  the Go alias path keeps the retained original state and follows that override.
- Before replaying trace fingerprints into concrete states, Java snapshots and
  resets `RandomEnumerableValues`; restore the snapshot after replay so specs
  using randomized enumeration regenerate the same path without perturbing the
  active runtime RNG.
- Java's no-alias behavior returns the current state/info, not the successor.
  Go `EvalAlias`, `EvalAliasInfo`, and `EvalAliasInfoPair` preserve that.
- The Java default `evalAlias` overloads evaluate the resolved `ALIAS` operator
  body under `EvalControl.Clear`, convert record-like values to alias states,
  and on evaluation errors attach `_ALIASEvalError` to an alias record instead
  of throwing. Go keeps this concrete on `Tool`, not behind a separate alias
  interface. Prefix/suffix trace aliases bind `TLCExt!Trace` through the
  standard operator's `OpDefNode` symbol so aliases using `Trace()` see the
  supplied prefix records.
- Safety-error postcondition traces run Java's pairwise alias evaluation over
  each trace entry before wrapping the trace in `CounterExample`, matching
  `Worker.doPostCondition`'s `evalAlias(current, successor)` overload. Do not
  bind `TLCExt!Trace` on this path. Prefix/suffix trace aliases belong only to
  Java paths that explicitly call the prefix overload, such as checker
  `TLCTrace.printTrace` and liveness/error-trace reconstruction. The `console`
  field is omitted only when the checker was not already done at the moment the
  error was accepted.
- Java also calls `TLCTrace.printTrace(curState, succState)` for next-state
  safety errors and for continuation-mode invariant/action-property violations.
  Go mirrors this with `ModelChecker.printBehaviorTrace`, which prints
  `TLC_BEHAVIOR_UP_TO_THIS_POINT`, aliases each recovered state, and emits
  state-printer messages in ordinal order before continuing or finishing the
  queue.
- Simulator behavior printing is different: Java `Simulator.printBehavior`
  constructs `TLCStateInfo` directly from each stored simulated state and calls
  pairwise `evalAlias(current, successor)`, so Go simulator printing must not
  bind `TLCExt!Trace` either.

`ConcurrentTLCTrace` merges per-worker trace fragments to reconstruct:

- safety counterexamples,
- liveness lassos,
- trace expressions,
- alias-enhanced traces.

Port guidance:

- Go goroutines can replace worker threads, but trace fragment ownership and
  synchronization must stay equivalent.
- Normal `TLCTrace` writes the Java-compatible `MC.st`-style file: each record
  is `longNat(predecessorPointer)` followed by the state fingerprint, and
  checkpoints store only the current file pointer and `lastPtr`.
- Java's static `TLCTrace.writeBehavior(File, TLCState, StateVec)` writes a
  compressed `ValueOutputStream` containing one non-normalized `TupleValue`.
  Each tuple element is `RecordValue(TLCState)` for the corresponding trace
  state. The Go helper `TLCTraceWriteBehavior` mirrors that format so
  `IOUtils!IODeserialize` can read serialized behaviors.
- `TLCState` disk serialization follows Java field order: `shortNat(workerId)`,
  `longNat(uid)`, `shortNat(level)`, then every state variable value. It does
  not serialize variable-count headers, nil markers, predecessors, or action
  names; those are runtime/trace-reconstruction metadata.
- `TLCStateMutExt.copy` preserves level, predecessor, and action but resets
  worker id and uid; `deepCopy` preserves worker id and uid. Neither copy path
  carries deferred executor callables, matching Java's `TLCStateMutExt`.
- The Go `TLCTrace` still keeps an in-memory mirror of `TraceRecord` values as a
  transitional convenience for direct state-object trace access. This mirror is
  not the source-of-truth file format and should shrink as reconstruction moves
  fully to Java-style fingerprint replay.
- Worker-local concurrent traces keep Java's separate record format:
  `longNat(predecessorPointer)`, `shortNat(predecessorWorker)`, fingerprint.
- `ConcurrentTLCTrace` checkpointing deliberately skips the shared `TLCTrace`
  checkpoint path. It checkpoints each worker trace fragment and creates the
  shared `MC.st.chkpt` marker file for Toolbox/script compatibility, matching
  Java's override.
- `ConcurrentTLCTrace.elements()` merges per-worker trace enumerators. Java's
  `nextPos` returns sentinel `42` for "has another fingerprint" and `-1` for
  exhaustion; Go mirrors that odd API so trace-counting and management code can
  stay mechanically aligned.
- State numbering and action labels in printed traces are test-observed.

## Simulation Architecture

`Simulator` performs random behavior generation instead of exhaustive search.

Constructor responsibilities:

- Create a `FastTool` in `Simulation` mode.
- Determine deadlock/liveness settings.
- Compute trace depth and trace count.
- Create one `SimulationWorker` per configured worker.
- Use independent RNG seeds per worker.
- Use per-worker liveness checker when liveness is enabled.
- Eagerly create `TLCGet("config")` record value.
- Initialize coverage if requested.

Simulation flow:

1. Check assumptions.
2. Compute all initial states once.
3. Validate init states and invariants.
4. Filter by model constraints.
5. Normalize initial states.
6. Start progress report thread.
7. Start workers.
8. Workers randomly choose initial and next states up to depth/count.
9. Main thread collects worker results.
10. Stop on fatal error or report continuable violations depending on mode.

Important behavior:

- Simulation is not just model checking with random queue order. It has separate
  liveness checking, trace file behavior, random element/subset hooks, and
  worker result aggregation.
- Seed and aril are printed and must be reproducible. `SimulationWorker` uses
  TLC's Java-compatible `RandomGenerator` semantics, including `nextDouble`
  based element selection, worker seeds from `nextLong`, and `nextPrime` as the
  alternative-action stride.
- Simulation traces compress finite stuttering steps in `getTrace` and then
  repair predecessor links, while `getUncompressedTrace` preserves the raw
  predecessor chain.
- After worker completion, Java evaluates postconditions even on simulation
  errors. If the worker result has a trace, it passes
  `SimulationWorkerError.getCounterExample()` into
  `Tool.checkPostConditionWithCounterExample`; otherwise it calls the ordinary
  postcondition check. The returned error code replaces the worker error only
  when its mapped exit status is more severe. Initial-state invariant
  violations are different: Java evaluates postconditions with a one-state
  `CounterExample` for output side effects and still returns the invariant
  violation code.
- Java has `SimulationWorker`, `ExplorationWorker`, `RLSimulationWorker`, and
  `RLActionSimulationWorker`. The Go port keeps one concrete
  `SimulationWorker` struct. The `debug bool` selects Java
  `ExplorationWorker` behavior: generate all successors through the next-state
  functor, randomly select one successor, execute deferred callables, and
  perform liveness/post-trace checks. RL modes remain enum values because they
  select different scheduling and Q-table behavior, while still using the same
  worker struct. Java gives debugger/exploration worker construction precedence
  over RL worker construction; the Go constructor therefore keeps the configured
  scheduler on `Simulator` but forces the effective worker mode to standard when
  `debug bool` is set.
- Java can select RL simulation with `tlc2.tool.Simulator.rl` or
  `tlc2.tool.Simulator.rlaction` system properties and tune RL with
  `.rl.alpha`, `.rl.gamma`, `.rl.reward`, and `.rl.enabledOnly`. Go accepts the
  same property names through the Java-style `-Dname=value` parser and keeps
  the `TLAGO_*` environment aliases as non-Java conveniences.
- Java reports `TLCGet("config").mode` as `"generate"` when
  `tlc2.tool.impl.Tool.probabilistic` is true and `"simulate"` otherwise.
  This is separate from the scheduler name (`random`, `rl`, or `rlaction`).
- Java's simulation statistics size action-transition matrices and report
  `TLCGet("actions")` from `tool.getSpecActions()`, i.e. initial-state actions
  followed by next-state actions. Go must use `Tool.GetSpecActions()` here too;
  using only next actions shifts ids and drops init-action entries.
- Java writes `Root_actions.dot` when trace-action output is `BASIC` or `FULL`.
  `BASIC` reduces action instances by source definition; `FULL` keeps each
  instantiated action and clusters vertices by context string. Both modes
  aggregate per-worker matrices, draw weighted seen edges with the Java log-log
  formula rounded to two decimals, and draw dotted unseen edges only when the
  sink is not an init predicate.
- Java `SimulationWorker.simulateRandomTrace` calls
  `IdThread.setCurrentState` after selecting the initial state and again after
  each selected successor. Go mirrors that with concrete `SetCurrentState`
  replacement semantics; scoped `PushCurrentState` remains for nested evaluator
  calls such as next-state generation.
- Java `SimulationWorker.simulateAndReport` catches generic `Exception`,
  publishes a worker error with code `0`, and stops that worker. Go mirrors this
  with a recover boundary because many ported Java `Assert.fail` paths are
  represented as panics carrying `*TLCError`.
- In simulation mode, `TLCExt!Trace` uses Java's
  `RecordValue(TLCState, Action)` constructor: each tuple element is a record
  whose first field is `_action`, followed by the state's variables. The
  non-simulation trace constructors remain state-record only.
- Java `RecordValue(Action)` emits `name`, structured `location`, and, for
  parameterized action instances, `context` and `parameters` in parameter
  order. Go mirrors the parameter fields on the concrete `Action` record helper
  and keeps the current string location until the SANY semantic nodes carry the
  Java `Location` coordinates.
- Preserve the Java `INextStateFunctor` contract: plain `addElement(TLCState)`
  is unsupported for simulation workers; only `setElement` and the
  action-tagged successor `addElement(s, a, t)` paths are valid.
- `ActionItemListExt.cons(Action, kind)` has a subtle coverage lookup:
  Java calls `act.cm.get(this.pred)`, not `act.cm.get(act.pred)`. Keep the Go
  `ConsAction` lookup on the receiver list's current predicate so coverage
  counters line up with Java's decomposition.
- Successor-processing errors from model/action constraints, trace writes, or
  state-writer paths bubble out to the caller's `doNextFailed` path. Only
  invariant and implied-action evaluation errors use Java's dedicated
  `doNextEvalFailed` path with the property/action name parameter.
- Initial-state processing writes unseen states to the state writer and worker
  trace, enqueues them, and only then calls `liveCheck.addInitState` with
  `tool.noDebug()`, matching Java `DoInitFunctor`.
- Simulator result consumption follows Java's continuation policy. Worker
  exceptions and liveness exceptions stop the run, and
  `TLC_INVARIANT_EVALUATION_FAILED`,
  `TLC_ACTION_PROPERTY_EVALUATION_FAILED`, and
  `TLC_STATE_NOT_COMPLETELY_SPECIFIED_NEXT` are non-continuable regardless of
  `-continue`. Other behavior errors keep workers running only when
  `Globals.Continuation` is set.
- After the result loop exits, Java interrupts every simulation worker and
  joins each for up to ten seconds. The Go port mirrors this with a stop flag
  and concrete worker `Join` channel so simulation returns without leaving
  worker goroutines running.

Port guidance:

- Port model checking first, then simulation.
- Preserve `RandomGenerator` behavior before porting simulation tests.

## Liveness Architecture

The liveness subsystem implements temporal property checking by translating
properties into liveness expressions, orders of solution, optional tableau
graphs, behavior graphs, and SCC checks.

### Formula Translation

`Liveness.astToLive` converts SANY expression nodes into `LiveExprNode`:

- constant-level formulas are evaluated immediately to `LNBool`.
- variable-level formulas become `LNStateAST`.
- action-level formulas become `LNAction`.
- temporal operators become `LNAll`, `LNEven`, `LNNext`, conjunctions,
  disjunctions, negation, etc.
- bounded quantifiers may expand into conjunction/disjunction when enumerable.
- recursive temporal operators must be expanded when their level exceeds action
  level.

The translator then normalizes:

- negation pushing,
- disjunctive normal form,
- fairness decomposition,
- promises,
- state/action checks,
- possible error models.

Java decomposes startup liveness processing into two distinct phases:

- `ParseLiveness(tool)` builds `livespec /\ ~livecheck` from config-derived
  actions. It conjoins every fairness/temporal action from
  `tool.getTemporals()`. If there is one implied temporal property, it appends
  its negation directly; if there are several, it appends a disjunction of
  their negations. If both lists are empty, it returns nil.
- `astToLive` handles only the temporal constructs Java handles explicitly.
  Fallback expressions whose level is still temporal fail with
  `TLC_LIVE_CANNOT_HANDLE_FORMULA`; in particular Java deliberately rejects
  `CASE` in temporal properties instead of inheriting TLC's first-match runtime
  evaluator semantics.
- `SF_e(A)` expands to `<>[]-ENABLED <A>_e \/ []<><A>_e`, and `WF_e(A)`
  expands to `[]<>(-ENABLED <A>_e \/ <A>_e)`. Go keeps the enabled-action and
  subscripted-action pieces as ordinary concrete `LiveExprNode` state/action
  nodes with `EvalFunc`, rather than introducing Java's `LNStateEnabled` and
  subclass hierarchy.
- `<<A>>_e` (`OPCODE_aa`) becomes the same subscript-aware action node, with
  the action body and subscript preserved separately. Prime still becomes an
  action over the whole primed expression.
- Bounded temporal quantifiers are expanded only when their domains enumerate:
  `\E` becomes an `LNDisj`, `\A` becomes an `LNConj`, and empty domains collapse
  to `FALSE`/`TRUE` respectively. Enumeration or child-conversion failure falls
  back through the whole expression's level, matching Java's guarded
  try/catch around `tool.contexts`.
- Function application mirrors Java's `OPCODE_fa` handling: if the function
  expression evaluates to a non-record lambda, the application context is
  validated with `getFcnContext`, then the lambda body is translated with the
  original context just as Java does.
- `processLiveness(tool)` tags state/action predicates, pushes negation into
  positive form, simplifies, converts to DNF, classifies each DNF conjunct into
  `<>[]A`, `[]<>A`, `<>[]S`, and remaining action-free temporal formulae, bins
  equivalent temporal formulae, and creates one `OrderOfSolution` per temporal
  formula bin.

The Go port keeps that decomposition in `liveness_process.go`. `ProcessLiveness`
returns the concrete `[]*OrderOfSolution` used by `LiveCheck`; it does not hide
the work behind an interface. Fast unit tests cover the parser/normalizer shape
because this code is lightweight and easy to drift away from Java.

### `OrderOfSolution`

Each temporal formula or conjunct maps to an `OrderOfSolution`:

- optional tableau `TBGraph`.
- eventuality promises.
- `checkState`: state predicates to cache per node.
- `checkAction`: action predicates to cache per edge.
- `PossibleErrorModel` entries describing accepting cycle conditions.

### Tableau

`TBGraph` builds tableau nodes from particles:

- initial particles are closure of the temporal formula.
- successors are implied-successor particles.
- duplicate particles reuse existing nodes.
- nodes get stable integer indexes.

### Behavior Graphs

`LiveCheck` has one live checker per `OrderOfSolution`.

- non-tableau checker stores behavior graph nodes keyed by state fingerprint.
- tableau checker stores product graph nodes keyed by state fingerprint plus
  tableau node index.
- Java stores graphs on disk for scale.
- state/action check results are cached as booleans/bitvectors.
- During the Go transition, `LiveChecker` keeps the existing in-memory graph for
  the current SCC implementation and also populates the Java-style
  `DiskGraph`/`TableauDiskGraph`. Lifecycle calls (`close`, checkpoint,
  recover, reset, flush) must go to the concrete disk graph fields. Once
  `LiveWorker` is ported, disk graphs become the primary SCC input and the
  in-memory checker can be retired.
- Tableau safety-like liveness shortcuts mirror Java's `errorGraphNode`
  handoff with concrete `ErrorGraphNode` and `ErrorPrefix` fields; trace
  printing will consume those when the Java `printErrorTrace` path is ported.
- `LivenessStateWriter` embeds the concrete `StateWriter` for DOT output
  instead of introducing `ILivenessStateWriter`. `NewDotLivenessStateWriter`
  writes Java's product-graph node ids (`stateFP.tableauIndex`) and preserves
  the stuttering/dotted/default visualization hints.
- The older in-memory `BEGraph` owner keeps Java's `initNodes`, `metadir`, and
  `NodeTable` fields. Its reset and shortest-path routines are intentionally
  iterative: `ResetNumberField` uses `MemObjectStack`, and `BEGraphGetPath`
  uses `MemObjectQueue` while destructively reusing parent pointers just like
  Java.

`AbstractDiskGraph` responsibilities:

- keep two files per solution, `nodes_N` for serialized `GraphNode` successor
  records and `ptrs_N` for `(fingerprint, tableau-index, node-file-pointer)`.
- write duplicate nodes without rewriting older records; the in-memory pointer
  table determines the distinguishable graph size.
- write new records at the current file pointer, not unconditionally at EOF.
  After recovery Java seeks both graph files back to checkpoint positions and
  subsequent writes overwrite from there while stale tails are ignored.
- rebuild the pointer table from `ptrs_N` before SCC search or recovery.
- treat values below `MAX_PTR` as node-file pointers and values in
  `[MAX_PTR, MAX_LINK]` as SCC link numbers.
- support checkpoint/recover by saving and restoring the current file pointers.
- provide optional fixed-size node caching, invariant checks over all graph
  records, and DOT/string traversal helpers for debugging.
- `DebugTableauDiskGraph` is selected by Java property
  `tlc2.tool.liveness.LiveCheck.debug` only for tableau disk graphs. It is not a
  different graph algorithm: after each `addNode`, `setDone`, and `recordNode`
  it creates the normal disk-graph cache, writes `dgraph_NNN_prefix.dot`, and
  destroys the cache in a `finally` path. Snapshot write failures are diagnostic
  and print rather than changing liveness checking results.
- reconstruct counterexample prefixes with breadth-first `GetPath` searches.
  The search rebuilds the pointer table from `ptrs_N` and then destructively
  reuses element slots as predecessor links. `TableauDiskGraph` needs the same
  concrete table with a `reverse` flag that adds one packed predecessor tableau
  index per record; this mirrors Java's reverse traversable table subclass
  without adding a Go interface or duplicate table hierarchy.
- `TableauDiskGraph` must have its own recovery and traversal methods in Go.
  Java gets virtual dispatch through `AbstractDiskGraph`; Go embedding does not
  make `DiskGraph.Recover` rebuild a `TableauNodePtrTable`.

### Checking

`LiveCheck.check0`:

- decides when to run based on graph growth threshold unless final check.
- distributes live checkers to `LiveWorker` threads.
- finds accepting SCCs/cycles.
- prints liveness counterexample and lasso.
- handles worker failure precedence: liveness violation wins over checker
  failure until all live workers complete.

`LiveCheck.checkTrace` is the simulation/debug trace path. Java converts the
current trace into a temporary behavior graph by adding the first state as an
init state, then for every non-final state adding both the stuttering
self-successor and the next trace state as successors. It then adds the final
state with an empty successor set, runs a final liveness check, and resets the
graph for the next simulated behavior.

`AddAndCheckLiveCheck` is Java's testing-only subclass that synchronizes
`addInitState` and `addNextState`, then calls `check0(tool, false)` after every
addition. Go keeps this on the concrete `LiveCheck` as an `AddAndCheck` flag
and mutex instead of creating another checker type.

Tableau liveness has an important safety-like short-circuit. When
`TableauLiveChecker.addNextState` recursively finds an accepting sink tableau
node and the `OrderOfSolution` has an empty possible-error model, Java records
`errorGraphNode`, reconstructs the prefix after the surrounding graph updates
complete, prints a temporal-property counterexample, stops the main checker,
and throws the invariant-violation control exception to escape the worker
without printing a second generic error. The Go port mirrors this with
`LiveChecker.printSafetyLikeLivenessError` and a fresh source
`INextStateFunctor.InvariantViolatedException` carrying its own known flag.
The initial-state functor uses its separate plain-runtime invariant exception;
worker wrapping exceptions retain the failing state/exception without a cause.
Debugger catches inspect direct exception classes rather than cause chains.

`LiveWorker`:

- implements Java's iterative Tarjan SCC search over disk-backed behavior
  graphs. The DFS stack stores packed `(state, tableau-index, location,
  lowlink)` cells and uses `SCC_MARKER = -42` to revisit a node after its
  successors have been explored.
- starts from disk graph initial nodes instead of preloading every vertex;
  successors that do not satisfy the PEM's EA action are queued as fresh roots
  if they still point to disk.
- uses `TableauNodePtrTable` as the temporary SCC component set, matching Java
  even for non-tableau graphs so trace construction can share the same shape.
- dispatches to the concrete `DiskGraph` or `TableauDiskGraph` fields on
  `LiveChecker`. There is intentionally no Go `AbstractDiskGraph` interface;
  the helper methods are private concrete switches.
- ports Java's lasso construction in data form: `GetPath` reconstructs the
  prefix, `dfsPostFix` greedily finds the component segment satisfying the PEM,
  and `bfsPostFix` closes the cycle. The current Go fields store the fingerprint
  `ErrorPrefix` and `ErrorCycle`.
- reconstructs `TLCStateInfo` traces from the prefix/cycle fingerprints,
  creates a `CounterExample`, and invokes the tool post-condition hook with
  that value. The Go checker also emits the temporal-property violation,
  counterexample marker, state trace, and stuttering/back-to-state marker.
- post-hoc property attribution mirrors Java's
  `Liveness.findViolatedProperties`: evaluate each implied temporal property
  over the reconstructed lasso and print the violated property names in
  deterministic action order. `ASTToLive` converts each action predicate to
  the same `LiveExprNode` family used by the liveness checker, and the result
  is cached in the action's auxiliary map via `AttachLiveExprToAction`.
  The structural core is in place; recursive temporal-operator metadata and
  exact Java diagnostic wording still need the same breadth-first mechanical
  deepening as their Java counterparts.

`LiveCheck1` is an older in-memory implementation used by simulation and some
trace checks. It follows the Manna-Pnueli book algorithm with component
numbering ranges. The Go port keeps the supporting `BEGraph`, `BEGraphNode`,
`BTGraphNode`, and `NodeTable` concrete; Java's subclass polymorphism is mapped
onto embedded structs only where the existing code already uses it.

Port guidance:

- Treat liveness as a major milestone, not a small extension.
- First port live expression nodes and formula processing.
- Then port in-memory `LiveCheck1` for simpler tests and simulation.
- Then port disk-backed `LiveCheck` and tableau product graphs.
- Preserve warning behavior for no fairness/liveness constraints/symmetry.

## Standard Modules and Overrides

Java TLC implements standard modules in `tlc2/module`:

- `Naturals`
- `Integers`
- `Sequences`
- `FiniteSets`
- `Bags`
- `Strings`
- `TLC`
- `TLCExt`
- `TLCEval`
- `TLCGetSet`
- `Randomization`
- `Json`
- `_TLCTrace`
- `_JsonTrace`
- `_Possible`
- `TransitiveClosure`
- `AnySet`

The Go port keeps these as concrete functions in `modules_*.go`. Related Java
modules may be folded together when the behavior is still direct; for example
`Strings` and `FiniteSets` live in `modules_misc.go`, and `TLCEval`'s value
conversion helper lives beside other TLC module operators. `TLCExt` includes
the definition-by-name hook as `TLCExtTLCEvalDefinition`, which looks up a
zero-arity `OpDefNode` in the concrete `Tool` definition table and evaluates
its body in the existing context/state pair.

Imported operators retain their original symbols/arities while source bodies
are converted, before module contexts are installed. Builtin operator arguments
retain the original builtin OpDef fallback used by SymbolNodeValueLookupProvider.
For unnamed non-LOCAL instances, Generator.generateInstance reuses the original
OpDef when the instancee or the definition's original module is parameter-free;
config replacements depend on that identity. LOCAL and named instances keep
separate definitions. SequencesExt.Cons has no annotated override in the source,
so its TLA body remains even though a core Sequences.Cons helper exists.
Graph simple-path prefixes snapshot their element array before backtracking.

The complete original CommunityModules Ant test target is translated in
community_modules_java_test.go, using the frozen existing sources and jars:
AllTestsUnix on Unix (AllTests on Windows), startup environment variables, the
unchanged nested IOExec Java commands/classpath and the separate ShiViz liveness
phase. Ant's two forks are fresh Go processes; source success and exit-13 checks
are retained. This is whole-suite evidence for the represented feature surface,
not proof of full native discovery, semantic metadata or TLC completion.

Standard module failures should preserve Java's specific `EC.TLC_MODULE_*`
codes and `Values.ppr` value rendering. In particular, `FiniteSets.Cardinality`
throws `TLC_MODULE_COMPUTING_CARDINALITY` for non-enumerable values, `STRING`
membership failures throw `TLC_MODULE_CHECK_MEMBER_OF`, and `STRING`/`ANY`
comparison failures throw `TLC_MODULE_COMPARE_VALUE`. These are observable
through error reporting and debugger paths, so avoid replacing them with generic
Go errors while porting module code. The numeric modules follow the same rule:
arithmetic overflow uses `TLC_MODULE_OVERFLOW`, `\div` by zero uses
`TLC_MODULE_DIVISION_BY_ZERO`, `0^0` uses `TLC_MODULE_NULL_POWER_NULL`, invalid
`%`/`^` arguments use `TLC_MODULE_ARGUMENT_ERROR`, comparison type errors use
`TLC_MODULE_ARGUMENT_ERROR_AN`, and `Nat`/`Int` membership/compare failures use
the same module membership/compare codes as Java.

`Sequences` has a few source-order details to preserve. `SubSeq` first decides
whether its first argument is a string or sequence and reports a first-argument
`TLC_MODULE_ARGUMENT_ERROR` before checking `m` and `n`; only after that does it
check natural-number arguments and domain membership. `Len`, `Head`, `Tail`,
`Cons`, `Append`, `Concat`, and `SelectSeq` should use
Java's sequence-specific `ONE_ARGUMENT_ERROR`, `APPLY_EMPTY_SEQ`,
`EVALUATING`, `ARGUMENT_ERROR`, and `ARGUMENT_NOT_IN_DOMAIN` codes. The public
Java helpers `SelectInSeq`, `Insert`, and `Remove` are not exported by the
frozen standard `Sequences.tla` interface and are not registered by the Go
standard definitions. Java's internal sequence-set object can print as
`BSeq(...)`, but `BSeq` is likewise not a standard operator name.

`Bags` distinguishes wrong-shape values from malformed bags. Operators such as
`BagCardinality`, `BagUnion`, `SqSubseteq`, `BagToSet`, and Java's odd
`SetToBag` error path use `TLC_MODULE_APPLYING_TO_WRONG_VALUE` when the value is
not the expected finite function/set shape. `BagIn`, `CopiesIn`, `BagCup`,
`BagDiff`, and `BagOfAll` use `TLC_MODULE_ARGUMENT_ERROR` or
`TLC_MODULE_ARGUMENT_ERROR_AN` for bad arguments. When `BagUnion` sees a finite
set whose element is not a bag, it reports `TLC_MODULE_BAG_UNION1` against the
whole set, not the individual element.

`TransitiveClosure.Warshall` reports a non-enumerable relation with
`TLC_MODULE_APPLYING_TO_WRONG_VALUE` and a non-pair element with
`TLC_MODULE_TRANSITIVE_CLOSURE`. The algorithm itself is Warshall over the
distinct relation endpoints in first-seen order from relation enumeration.

Core `TLC` module operators also carry observable Java validation order.
`Assert` throws `TLC_VALUE_ASSERT_FAILED` with `Values.ppr` of the second
argument. `@@` reports first/second function-shape errors with
`TLC_MODULE_ARGUMENT_ERROR`. `SortSeq` checks that its first argument converts
to a tuple, then checks that the comparator is an operator before returning for
an empty sequence; Java's first-argument message says "natural number" and the
Go port intentionally preserves that wording. `Permutations` and the finite-set
paths of `RandomElement` report `TLC_MODULE_APPLYING_TO_WRONG_VALUE`.

`Randomization` validates public arguments in Java order and reports
`TLC_MODULE_ARGUMENT_ERROR`. `RandomSetOfSubsets` checks first-argument count,
second-argument subset size, third-argument finite set, requested number of
subsets against `2^Cardinality(S)`, then subset size against `0..Cardinality(S)`.
Java also has a `RandomSubsetSet` helper that reports under the internal
operator name `RandomSubsetSetProbability`, including for probability parsing
and the requested-subsets bound. The frozen standard `Randomization.tla` module
does not export this operator, so the Go standard registry must not install it.
The helper remains useful for direct module parity. The requested-subsets bound
uses Java's
`31 - Integer.numberOfLeadingZeros(numberOfPicks) + 1 > Cardinality(S)` guard and
Java `int` left-shift overflow semantics, not a floating-point `2^n` shortcut.

`TLCGetSet` uses `TLC_MODULE_TLCGET_UNDEFINED` for missing numeric and string
registers. Invalid `TLCGet` arguments report `TLC_MODULE_ONE_ARGUMENT_ERROR`;
invalid `TLCSet` arguments report `TLC_MODULE_ARGUMENT_ERROR`. String register
names are rendered without quotes in the undefined `TLCGet(name)` message, just
as Java uses `String.valueOf(sv.val)`.

`TLCExt` one-argument operators should use Java's `TLC_MODULE_ONE_ARGUMENT_ERROR`
surface: `ToTrace` expects a `CounterExample`, `TLCModelValue` reports as
`ModelValue` and expects a string, and `TLCEvalDefinition` distinguishes
non-string names, unreachable definitions, and non-zero arity definitions through
that same error code.

`TLCExt!AssertError` is an evaluating operator with an important ordering
constraint. Java first requires the expected-error argument to be syntactically
a `StringNode`; an expression that happens to evaluate to a string is rejected.
It then evaluates the expression that is expected to fail, and only compares the
string literal with the caught exception message when that expression throws.

`Json` standard operators use `TLC_MODULE_ARGUMENT_ERROR` for the public payload
shape checks Java performs inside the module: `JsonSerialize` requires its
second argument to be a sequence or record, and `ndJsonSerialize` requires a
sequence. The separate registered `IOUtils!Serialize` path is implemented in
`modules_ioutils.go`; do not silently merge its result-record convention with
the direct `Json` module exceptions.

`TLCExt!CounterExample` is context-sensitive in Java: postcondition checking
conses the current `CounterExample` value under the actual
`CounterExample` `OpDefNode` symbol and the module operator returns that value
when present, otherwise an empty counterexample. The Go standard operator is
therefore registered as an evaluating operator rather than a plain method so it
can read the current postcondition context; do not bind by name through a
synthetic symbol.

`TLCExt!Trace` is also an evaluating operator. In simulation mode Java asks the
active `Simulator` to build the current trace. In model-checking mode Java uses
the trace file for committed states, but has a special transient-state path when
the current state still has `TLCState.INIT_UID`: it reads
`IdThread.currentState`, reconstructs the committed prefix for that predecessor,
then appends the predecessor and transient state. The Go port mirrors this with
the goroutine-local current-state scope and the concrete `ModelChecker`
trace-reconstruction helpers.

`TLCExt!PickSuccessor` is synchronized and deliberately interactive. Before
prompting, Java accepts already-seen BFS successor fingerprints because TLC
checks action constraints before it filters old states. When the guard is
`FALSE` and the successor state is complete, Java identifies the action from the
extended successor state or regenerates next states to find the first matching
action, then reads `stdin` commands: yes/blank accepts, `n` rejects, `s` prints
both states, `d` prints the state diff, and `e` marks the successor explored in
BFS mode. All prompt/status text goes through `MP.printMessage` with
`TLC_MODULE_OVERRIDE_STDOUT`, so the Go port must use `PrintMessage` for this
interactive output rather than writing directly to stdout.

Preserve Java override annotations when registering standard operators.
`TLC!TLCEval` is an evaluating override because it receives the unevaluated
expression and caches converted constant-level results on the semantic node.
Its Java implementation branches directly on the argument semantic level:
state/action/temporal expressions evaluate in the incoming context and are not
cached; constant-level expressions with a non-empty context also evaluate
without caching; only constant-level expressions under an empty context use the
static read/write-lock protected semantic-node cache. Cache reads must mux
`WorkerValue` through the active worker id before returning it. Cache writes
must mirror Java's `WorkerValue.demux`: evaluate under
`EmptyContext`/`EmptyState`, deep-normalize, and, when a mutable value is shared
across multiple workers, reevaluate it once per worker with the same
random-enumerable seed before muxing the active worker's copy and converting it
through the legacy `toSetEnum`/`toFcnRcd` path.
Ordinary symbol lookup muxes a cached `WorkerValue` by the current worker ID.
Outside that scope it selects zero, matching Java's non-`IdThread` path.
Copied-state worker metadata does not override the per-worker constant choice.
`TLC!TLCGet`, `TLCExt!CounterExample`, `TLCExt!Trace`, `_TLCTrace!_TLCState`,
`_JsonTrace!_TLCState`, and `_Possible!_Counts` all carry non-constant
minimum levels in Java to prevent invalid constant folding. `TLCExt!PickSuccessor`
is action-level. Conversely, `TLCExt!TLCGetOrDefault` is a plain Java operator,
so both arguments are evaluated before it chooses between the register value and
the default. `TLCExt!TLCGetAndSet` is a TLA definition, but it calls the
Java-overridden `TLCGetOrDefault`; a direct Go implementation must therefore
evaluate `defaultVal` before reading the register rather than lazily only on a
missing register.

`Randomization!RandomSubset(k, S)` must follow Java's `EnumerableValue`
subset enumerator rather than a plain shuffled sample. Java chooses a seed
index with `Random.nextInt(|S|)`, chooses an increment with
`RandomGenerator.nextPrime`, computes the same `m/a` LCG parameters used by
`EnumerableValue.computeOptimalMandA`, and emits `k` random indices from the
LCG. The emitted `SetEnumValue` is initially unnormalized, so duplicate draws
are preserved until normal set comparison/fingerprinting forces
normalization. Randomized enumeration (`elements(Ordering.RANDOMIZED)`) uses
that same `SubsetEnumerator` path with `k = size`, so it must not be
implemented as a Fisher-Yates shuffle; the random-value consumption is
observable for seeded runs and aril replay. The LCG iterator is lazy, and reset
clears its call count while retaining its current seed. Explicit-set and interval
random iterators select directly without ordinary traversal counts. Nonempty
zero-count iterators still initialize their seed and increment. Java's module
argument checks accept negative integer counts, but `getRandomSubset` then fails
allocating `ValueVec(k)` at the Java-method override boundary.

`RandomSetOfSubsets` and `RandomSubsetSetProbability` construct a
`SubsetValue` over the input enumerable and use `CoinTossingSubsetEnumerator`.
That path normalizes the base set before tossing one coin per base element,
then accumulates generated subsets in a hash set so duplicates are dropped
during generation rather than by sorting a `ValueVec` afterward.

`TLC!RandomElement` also uses `RandomEnumerableValues.get().nextDouble()` for
intervals and finite enumerated sets. The Go port must use the shared
Java-compatible random enumerable generator here, not Go's process-global
random source, so seeded simulation and trace replay consume random values in
the same places Java does.

`RandomEnumerableValues` is mode-sensitive in Java. During BFS model checking,
worker evaluation and trace reconstruction set the current predecessor state on
the current worker thread; the random-enumerable generator is then seeded with
`enumFractionSeed XOR predecessor.fingerPrint()`, and repeated random choices
within the same predecessor continue the same stream. Initial-state generation,
DFID, and simulation use the default per-thread seed stream without predecessor
reseeding. The Go port mirrors Java `ThreadLocal<Random>` with per-goroutine
state; do not collapse it into one process-global RNG.

`Json!ToJsonObject` mirrors Java's `getObjectNode` dispatch. Records and
tuples become JSON objects, but a function record whose domain is a valid
sequence `1..n` is routed back to the array writer even under `ToJsonObject`.
This oddity is source-compatible with Java and must be preserved for
round-tripping existing specs.

Java's `Sequences` class contains helper methods such as `SelectInSeq` and
`Insert`, and its internal sequence-set object can print as `BSeq(...)`, but
the frozen `Sequences.tla` module exports only the standard sequence operators.
Do not register those helper names as Go standard definitions unless upstream
adds them to the standard module interface.

`TLCGet("diameter")` is mode-sensitive. In model checking it reports checker
progress. In simulation Java reads the current `SimulationWorker` trace count,
returning zero while initial states are generated outside a worker. The Go port
stamps simulation states with the concrete worker id and uses that id to read
the worker-local trace count, which also keeps worker-local `TLCGet`/`TLCSet`
registers aligned with Java.

`TLCGet("generated")`, `"distinct"`, and `"queue"` are direct main-checker
queries in Java. Simulation exposes generated trace/state counters through
`TLCGet("stats")`, not those direct string keys. Checker `"stats"` has fields
`queue`, `distinct`, `initial`, `generated`, `diameter`, `duration`, and
`worker`; checker `"config"` has fields `mode="bfs"`, `deadlock`, `worker`,
`seed`, `fingerprint`, and `install`. Simulator `"stats"` has fields `traces`,
`duration`, `generated`, `behavior`, `worker`, `distinct`, `distinctvalues`,
`retries`, `actions`, `levelmean`, and `levelvariance`; simulator `"config"`
has fields `mode="simulate"`, `depth`, `traces`, `deadlock`, `seed`, `aril`,
`worker`, `install`, and `sched`. Java eagerly caches checker and simulator
config records; simulator `aril` in that record is the construction-time field,
not a recomputation from the RNG after worker seeds have been drawn.
Java distinguishes direct `TLCGet` counters from stats-record counters:
`TLCGet("generated")`, `"distinct"`, `"queue"`, and `"duration"` use
`Math.toIntExact` and report overflow, while stats and coverage records use
`IntValue.narrowToIntValue`, which returns `-1` when a `long` does not fit in
TLC's 32-bit integer value.
Java enables expensive simulator extended statistics through
`tlc2.tool.Simulator.extendedStatistics` and switches exact counters on with
`.extendedStatistics.naive`. The Go port supports those property names and
`TLAGO_SIMULATOR_EXTENDED_STATISTICS(_NAIVE)` aliases, using HyperLogLog bits 8
for distinct states and 10 for distinct variable values as Java does.

Integer `TLCGet(i)`/`TLCSet(i, v)` and named registers (`"s:..."`) are
current-worker local in Java when evaluated on an `IdThread`; they do not use
the predecessor state's stored trace worker id. Outside a worker, Java
broadcasts `TLCSet` through the checker/simulator; simulator `TLCGet` fallback
reads worker 0 exactly, even when worker 0 has no value. Store these registers
on the concrete Go `Worker` and `SimulationWorker` structs, not centrally on
the checker/simulator, because Java's `AbstractChecker` and `Simulator` read
them back from worker objects. `TLCGet("all")` and `TLCGet("all:named")`
return functions whose values are per-worker tuples and, like Java, iterate
indices or named keys visible on worker 0. The Go port's ambient current-worker
helper must be goroutine-scoped, mirroring Java's thread-local worker identity.
A single process-wide worker slot is not
correct once workers run concurrently. Java `IdThread` also stores the
predecessor state while next states, alias state records, and reconstruction
states are evaluated; the Go port mirrors this with a goroutine-local
current-state stack and resets it on checker error paths. `TLCSet("exit",
TRUE)` stops the active checker and simulator. `TLCSet("pause", TRUE)` is a
blocking BFS model-checker control: Java prints
`Press enter to resume model checking.` and waits on
standard input before returning `TRUE`.

Incomplete next-state errors must carry Java's parameter vector: for a
single-action spec, the plurality fragment and comma-joined unassigned variable
names; for multi-action specs, the action name followed by those two fields.
Simulation, BFS, and DFID use this exact shape for next-state failures.

Random subsets of product-shaped values (`[S -> T]`, record sets, tuple
products) must use Java's product-index strategy from
`SetOfFcnsOrRcdsValue`. TLC converts each constituent to a `SetEnumValue`,
draws random indices over the mixed-radix product, and reconstructs the
corresponding function, record, or tuple. When the product cardinality exceeds
32-bit size, Java switches to a `BigInteger` path using `Long.MAX_VALUE - 24`
as the stride and `RandomEnumerableValues.nextLong()` as the offset.
For `SUBSET S`, Java's `SubsetValue.elements(k)` draws random bit-mask indices
when `|S| < 31` and `k <= 2^16`; otherwise it uses coin tossing over the
normalized base elements. This direct `elements(k)` API differs from both
`RandomSubset(k, SUBSET S)` and powerset randomized ordering, which first
materialize the powerset to an explicit set. `KSubsetValue` randomized ordering
uses an endless Algorithm S generator with a no-op reset and the default model.
Random-subset results inherit the receiver's cost model; product subsets also
count their result elements. Random tuple and record elements inherit the model,
while random function elements deliberately use the default model. Product
iterators initialize their random seed or offset before converting constituents
and cache their mixed-radix weights, preserving Java's ordering and laziness.

Override infrastructure:

- `TLARegistry` maps TLA+ names to Java names.
- `TLAClass` loads module classes.
- `TLCBuiltInOverrides`, `ITLCOverrides`, and annotations describe overrides.
- `MethodValue` and `EvaluatingValue` bridge operator calls to module methods.

Go mapping:

- Java standard-module static initializers are represented by
  `Tool.InstallStandardDefinitions`.
- `NewTool` installs concrete values such as `Nat`, `Int`, `STRING`, and
  `Any`, plus `MethodValue` entries for pure module operators.
- Java's `TLARegistry` aliases are installed beside the method names, e.g.
  `Plus` and `+`, `Concat` and `\o`, `MakeFcn` and `:>`.
- Context-sensitive operators such as `TLCGet` and `TLCSet` use
  `EvaluatingValue` so their implementations receive the concrete `Tool`,
  context, current state, successor state, eval control, and cost model.
- JSON, TLCExt, `_TLCTrace`, and `_Possible` overrides follow the same table.
  `TLCExt!TLCCache` uses a concrete `TLCExtCache` for constant-level
  expressions and the concrete `TLCStateMut` cache for state-level
  expressions, matching Java's split between expression tool objects and
  `TLCStateMutExt`. Java only evaluates the closure in those two cacheable
  cases; action- and temporal-level expressions bypass the cache and evaluate
  the expression directly. The state-level closure key is evaluated against the
  current state only, mirroring `tool.eval(closure, c, s0)`.
- `_TLCTrace!_TLCState` and `_JsonTrace!_TLCState` are evaluation overrides in
  Java. They ignore the syntactic level argument and return a record
  representation of the current state directly, avoiding reconstruction through
  `TLCExt!Trace`. The Go port registers `_TLCState` as an `EvaluatingValue`
  with both module-qualified aliases.
- CommunityModules `IOUtils` is also part of the practical TLC runtime surface.
  Java implements it in `tlc2.overrides.IOUtils`, not in the core
  `tlc2.module` package. The Go port keeps it in the same central `tlc`
  package with concrete functions for `IOSerialize`, `IODeserialize`,
  text/NDJSON `Serialize`, text `Deserialize`, environment lookup, process
  execution, template execution, and `atoi`. Explicit IOUtils compression uses
  gzip-wrapped Java value streams. Commands must be actual TupleValue objects;
  both template tuple checks precede string-element conversion. Environment
  record validation comes first, while environment-value conversion follows
  formatting. Templates use shared JavaFormatStrings, preserving ignored extra
  arguments, indexing and exceptions before process creation. Unquoted
  environment values preserve Java's StringValue-subclass behavior.
  TXT uses Java's strict Files.writeString/readString codec paths, including
  optimized UTF-8 malformed-length reporting, Latin-1/ASCII unmappable failures,
  UTF-16 surrogate/BOM behavior and empty-output BOM suppression. Ordinary
  OutputStreamWriter/FileWriter constructors use replacement; Files.newBufferedWriter
  in core Json.textSerialize uses a REPORT encoder. Core codecs operate on preserved
  UTF-16 units and recognize the six guaranteed charsets with actual JDK aliases;
  extended/provider charsets remain pending. Coding exceptions retain their
  IOException family and source class/message; illegal/unsupported charset names
  retain their separate IllegalArgumentException families. Read charset lookup
  precedes file access, and NUL path validation precedes charset lookup. Process
  default/native charset discovery and process byte decoding remain distinct
  source requirements; callback casts and option-validation order are represented
  below.
  IOUtils.atoi uses the shared Java decimal parser: UTF-16/BMP decimal digits,
  ASCII signs, signed int limits, decimal leading zeros and rejection of
  supplementary digits, whitespace and radix prefixes. Source StringValue
  subclasses are accepted; invalid values retain the original TLC error code
  and parameters. Null reaches the source NullPointerException boundary while
  constructing error parameters; VM-generated helpful-null detail remains pending.
  TXT callback casts retain concrete Java type errors and evaluation order;
  format lookup/cast lies outside the parameter catch. Only Exception enters
  error records, while Error reaches the evaluating override's outer catch.
  Charset lookup precedes enum mapping, strict encoding precedes incompatible
  output-stream flag checks, and explicit options do not gain empty-list defaults.
  The priority wrapper retains the IOUtils primary reflected signature while
  executing JSON 25 before TXT 50; its real source TLA fallback is preserved.
  Native DSYNC is distinct from SYNC where available; creation permissions honor
  the process umask. Linux TXT paths collapse repeated/trailing slashes without
  resolving dot/parent components; empty syscall paths address the current
  directory. The represented native UTF-8 encoding rejects malformed UTF-16 paths
  before charset lookup. Linux opens produce typed NIO exceptions with normalized
  path messages; channel failures produce IOException without a path. CREATE_NEW
  retains its source empty/final-dot quirks. DELETE_ON_CLOSE forces final-link
  nofollow except with CREATE_NEW, unlinks immediately, and ignores unlink errors.
  Files writes complete 8192-byte chunks; close failure is reported if no earlier
  error exists. Reads retain the int-array size Error before allocating content.
  Native filename charset discovery, other filesystem providers, memory-exhaustion
  allocation behavior and other public convenience I/O boundaries remain source work.
  Core Json.textSerialize evaluates options, converts its record and checks format,
  then evaluates/converts payload once, then destination. Open-option tuple and
  charset casts precede element casts outside the try; path/charset/enum/flag/open,
  node conversion, writing and close lie inside catch(Exception), retaining Error.
  Its writer mirrors platform-thread BufferedWriter's 8192 UTF-16-unit buffer and
  StreamEncoder's separate 8192-byte buffer with REPORT coding exceptions. Close
  flushes and closes even after failure; later close failures are suppressed on
  the primary failure, recursively, before the outer RuntimeException cause is
  constructed. Throwable rendering prints suppressed failures before the cause.
  Gson string escaping keeps raw UTF-16 units and does not apply HTML escaping;
  unsupported nodes retain the source IOException class/message. Record apply
  returns a present null component directly. Ordinary Json.serialize/ndSerialize
  use the same buffered stream carrier with replacement encoding, open before
  node conversion, ignore mkdirs' return and retain unwrapped failures/suppression.
  Linux legacy File paths replace malformed filename units for native access while
  retaining the logical UTF-16 path in FileNotFoundException; NUL/empty behavior
  remains distinct from NIO. Directory creation retains canonical fallback and
  hard parent failures and the native kernel symlink limit, rather than Go's
  more permissive EvalSymlinks limit. Adjacent high surrogates preserve the
  next pending unit after replacement. Initial file.encoding is frozen separately from runtime
  property mutation and default charset resolution is cached; the six guaranteed
  charsets/aliases are represented. Json's three synchronized static writers
  share the existing reentrant class monitor; NDJSON holds it across evaluation.
  Reader boundaries, extended provider/COMPAT/native default charset discovery,
  Gson parser leniency and other native File providers remain source work.
  Generic value-stream file paths that
  correspond to Java's `ValueOutputStream(File/String)` and
  `ValueInputStream(File/String)` honor the global gzip flag; raw
  `DataOutputStream`/random-access graph files remain uncompressed like Java.
- CommunityModules `Combinatorics` provides annotated `factorial` and `choose`
  candidates, distinct from the lower-level `tlc2.util.Combinatorics` helpers.
  Java rejects the one-argument factorial method for the zero-argument source
  function `factorial[n \in Nat]`; the bridge retains that recursive TLA
  function and installs only the eligible choose override for this source.
  MethodValue retains reflected parameter counts for annotated registration;
  do not skip a source body solely because an override name is registered.
- CommunityModules `Bitwise` contributes native overrides for LOCAL recursive
  helper definitions `And`, `Or`, and `Xor`, plus exported `Not` and `shiftR`.
  The exported infix operators `&`, `|`, and `^^` remain TLA definitions that
  call the native helpers. The bridge therefore skips all five annotated native
  definitions when converting `Bitwise.tla`, but named-instance export bindings
  include only `&`, `|`, `^^`, `Not`, and `shiftR`; exporting `B!And` would
  violate the module's LOCAL boundary.
- CommunityModules fold/function/set/bag helpers are practical native override
  surface too. Java overrides `DyadicRationals!Reduce` even though it is LOCAL,
  so the Go bridge skips the TLA definition but does not export `D!Reduce` for
  named instances. Java `Functions.IsInjective` uses a non-mutating O(n^2)
  duplicate check for already-user-visible tuples and interval functions, but
  sorts freshly converted tuple values in place. `Functions.AntiFunction`
  materializes a function record and normalizes the inverse record. The fold
  overrides intentionally differ in accumulator order: `Functions` and
  `FiniteSetsExt` call `op(value, acc)`, while `BagsExt.FoldBag` follows Java's
  `op(acc, bagElement)` loop for each multiplicity.
- CommunityModules `CSV` and `GraphViz` are side-effect/string rendering
  overrides. CSV's actual string arguments use the shared Java formatter,
  preserving indexing/reuse, flags, UTF-16 precision, uppercase/boolean/hash
  output and typed invalid-template/conversion failures before file creation.
  Ten direct source observations match return values, exact UTF-8 file bytes,
  exceptions and creation timing. It appends the source line separator,
  normalizes records before header/value emission, uses
  raw `StringValue` contents for paths and delimiters, treats the delimiter in
  `CSVRead` as a Java regex, returns an empty tuple/zero count for missing
  files, and overflows beyond signed 32-bit record counts. `GraphViz.DotDiGraph`
  renders a fixed `digraph MyGraph {...}` string; node ids are Java signed
  `long` fingerprints, not unsigned Go `uint64` decimal strings.
- CommunityModules `Graphs` and `UndirectedGraphs` share Java
  `AbstractGraphs`: validate graph records, normalize node/edge sets, build
  adjacency restricted to endpoints in `G.node`, skip malformed edge values,
  enumerate simple paths with depth-first backtracking, and answer connectivity
  with breadth-first reachability. Directed edges are ordered 2-tuples;
  undirected edges are sets of one or two nodes, with singleton sets treated as
  self-loops. `UndirectedGraphs.ConnectedComponents` uses union-find. Because
  `SimplePath` and `AreConnectedIn` are exported by both modules but differ in
  semantics, the Go port registers module-qualified native values such as
  `Graphs!SimplePath` and lets the root bridge alias skipped SANY definitions
  and named instances to the module-qualified value.
- CommunityModules `VectorClocks.CausalOrder` is a ShiViz-style topological
  sort over vector-clock log entries. Java first groups entries by the node
  returned from `node(entry)`, stable-sorts each node's log by its own clock
  value, constructs parent links whenever another host's clock component
  advances beyond the per-host global clock, then repeatedly emits root entries
  with no remaining parents. The Go port keeps this phase decomposition and the
  same parent-removal behavior, but stores host logs and parent/child sets in
  explicit slices so iteration remains deterministic.
- CommunityModules `SVG` has three self-contained native overrides and one
  third-party graph-layout override. Go ports `SVGElemToString`,
  `NodeOfRingNetwork`, and `PointOnLine` directly: attribute underscores become
  dashes, attribute values are single-quoted, child elements are recursively
  serialized, `<<`/`>>` in inner text are escaped, ring coordinates use Java's
  polar conversion and integer truncation, and `PointOnLine` uses the Java
  floating division/truncation formula. `NodesOfDirectedMultiGraph` remains on
  the TLA fallback until the JGraphT/JUNG layout algorithms can be mirrored
  faithfully.
- CommunityModules `SequencesExt` mixes ordinary Java
  `@TLAPlusOperator` overrides with two string-only `@Evaluation` shortcuts.
  Go ports the native set/sequence conversion, longest-common-prefix, fold,
  search, remove, suffix, and all-subsequence helpers directly. Java returns
  `null` from `ReplaceFirstSubSeq`/`ReplaceAllSubSeqs` for non-string values so
  the pure TLA definition handles tuples; until the Go evaluator grows that
  exact `@Evaluation` fallback path, the Go helper must preserve equivalent
  tuple behavior itself. Keep the Java quirks: `SelectInSubSeq` and
  `SelectLastInSubSeq` return indices in the original sequence range, and
  `SelectInSeq` reports the non-boolean predicate position as `"third"`.
- CommunityModules `Statistics.ChiSquare` delegates to Apache Commons Math
  `ChiSquareTest`. Java normalizes/converts both inputs to function records,
  ignores the function domains after normalization, converts values to expected
  `double[]` and observed `long[]`, parses alpha from a `StringValue`, and
  returns `FALSE` when Commons Math rejects the null hypothesis. Go mirrors the
  Commons Math statistic, including rescaling expected counts by
  `sumObserved/sumExpected` when totals differ, and computes the chi-square
  survival probability through the regularized gamma Q function locally.
- `_POSSIBLE` is also a config-driven model transformation in Java
  `SpecProcessor`. Each configured predicate is wrapped in `_Possible!_Track`
  and installed as a model constraint for state-level predicates or an action
  constraint for action-level predicates. A matching `_Possible!_CheckName`
  postcondition reports `TLC_POSSIBLE_UNWITNESSED` after model checking if the
  named predicate was never witnessed. The Go bridge mirrors this with concrete
  `PossibleTrackNode` and `PossibleCheckNode` structs and stores counts in the
  named register `s:_possible`, using worker-local checker values so `_Counts`
  can merge them like Java's `TLCGet("all:named")`. Java's `_Possible.java`
  only declares `_Counts` as a native override; `_Track`, `_CheckName`, and
  `_PrintCounts` remain ordinary TLA definitions in `_Possible.tla`, so the Go
  bridge must install them as `OpDefNode`s for explicit standard-module use and
  named instances. For postcondition errors, Java evaluates the generated
  `_CheckName` predicate but reports the user's original predicate body;
  ordinary postcondition evaluation errors also include the predicate body as
  the second message parameter.

Port guidance:

- Implement built-ins as Go functions registered in a central table.
- Preserve override names and arities.
- User-defined Java override jars do not have a direct Go equivalent. Record
  this as an interoperability feature to design after core TLC parity.
- Do not silently ignore unsupported overrides; emit equivalent errors.

## Output and Error Architecture

`tlc2.output.EC` defines message/error codes. `MP` formats and prints messages
and broadcasts them to recorders. Tests commonly assert:

- that a code was recorded,
- that a code was not recorded,
- exact string values attached to a code,
- error trace state/action messages,
- coverage messages.

Important codes include:

- `TLC_FINISHED`
- `TLC_STATS`
- `TLC_INIT_GENERATED1`
- `TLC_INIT_GENERATED2`
- `TLC_SUCCESS`
- `TLC_COUNTER_EXAMPLE`
- `TLC_STATE_PRINT1/2/3`
- `TLC_BACK_TO_STATE`
- `TLC_DEADLOCK_REACHED`
- invariant/action violation codes.
- liveness violation codes.
- config/parser/general error codes.

Port guidance:

- Port `EC` constants early.
- Implement message recording before porting end-to-end tests. Use a concrete
  recorder/broadcaster unless multiple external recorder implementations are
  truly needed.
- Java records messages before suppression/`-nowarning` decides whether the
  console sees them. Go `Message.Suppressed` preserves that split for callers
  that need recorder parity while still knowing whether a diagnostic was
  user-visible. Java `printError` does not consult suppression, while
  `printMessage`, `printWarning`, `printTLCBug`, and state printing do.
- `MP.getMessage` notifies recorders before formatting without printing. Go
  marks these events `Message.FormattingOnly`; coded exception constructors
  preserve them too.
- `-messagesAsErrors` and the dynamic `tlc2.output.MP.warning2error` property
  abort `printWarning` through `Assert.fail`. Its `TLCRuntimeException`
  construction records an `MP.getMessage` event before throwing; it never
  reaches the ordinary warning event.
- Keep human text close to Java but assert primarily through codes and
  structured parameters like Java tests do.
- State string formatting is semantic output. Treat it as part of compatibility.

## Trace Exploration and Spec Writers

Java TLC has a small but important output-writing subsystem that creates TLA+
modules and configs from model data:

- `AbstractSpecWriter`: owns TLA and CFG buffers, appends module closing tags,
  writes streams/files, emits constants, formulas, views, aliases, and model
  value declarations.
- `SpecWriterUtilities`: creates generated identifiers, module primers,
  closing tags, formula/source content arrays, and override definitions.
- `SpecTraceExpressionWriter`: builds trace-exploration Init/Next relations,
  trace functions, trace-expression stubs, properties/invariants that reproduce
  error traces, lasso views, and config wrappers.
- `TraceExpressionExplorerSpecWriter`: creates the `TEExpression` helper module
  for user trace expressions, preserving a deterministic variable-expression
  map in comments.
- `TraceExplorationSpec`: coordinates the full generated `_TTrace`/`TETrace`
  spec from an `MCError` trace and registers the binary trace postcondition.

Go keeps this as concrete structs in the central `tlc` package:

- `SpecWriter` is the non-abstract buffer owner corresponding to Java's
  `AbstractSpecWriter`.
- `SpecTraceExpressionWriter` embeds `SpecWriter` and ports the trace-specific
  emitters.
- `TraceExpressionExplorerSpecWriter` stores the variable-expression map in
  `InsMap` so generated module order is deterministic.
- `TraceExplorationSpec` captures the naming/postcondition shell and can now
  generate the monolithic TE `.tla` file from an `MCError`.
  Its variable list comes from the global TLC state variables, matching Java's
  `TLCState.Empty.getVarsAsStrings()` call, rather than being reconstructed
  from the recorded `MCError`.
- `ErrorTraceMessageRecorder` mirrors Java's
  `ErrorTraceMessagePrinterRecorder`: it observes state-print/back-to-state
  message codes and builds an `MCError` for TE generation. The `TLC` runner
  subscribes it only when `Options.GenerateTraceSpec` is set.
- `-generateSpecTE` also installs Java's implicit binary-trace postcondition
  before the tool is built when binary trace generation is enabled. Go stores
  the derived TE module name in `Options.TraceSpecModuleName` so the
  postcondition path and final generated `.tla` module use the same name.

Correctness notes:

- Generated identifiers intentionally mirror Java's `scheme + currentMillis +
  counter*1000` shape.
- Writer output order is semantic. Do not use unordered Go maps in emitters.
- The generated timestamp text is not a semantic input to TLC tests; compare
  structural output or normalize that line when doing byte-level writer tests.
- Trace expressions are represented as variables and definitions, with
  temporal-level expressions initialized to `"--"` and primed according to
  Java's level rules.

## Coverage Architecture

Coverage is represented by:

- `CostModel`
- `CostModelNode`
- action wrappers and op-application wrappers.
- coverage hash tables.
- reporting through `CostModelCreator` and output messages.

`CostModelCreator` does not merely walk the syntax tree and add children. Java
reconstructs one call tree per action because the SANY semantic graph shares
operator definitions globally, while coverage must be reported per action. Its
per-action state is load-bearing:

- a stack of current `CostModelNode`s, rooted at the current `ActionWrapper`;
- a substitution map from substituted expression nodes to `Subst` identity;
- a map from higher-order operator body nodes to wrappers that should later
  receive that body as a child;
- an active `OpDefNode` set used with `CoverageHashTable` to stop recursive
  operator expansion only when a recursive definition is already on the path;
- a LET-IN map from each LET body to the IN body wrapper that should also see
  the LET part through `OpApplNodeWrapper.addLets`;
- a global context approximation used only when an operator application has
  operator arguments, so `Op(s)` can later be connected to the passed
  operator/LAMBDA body.

The Go port keeps the same side-table shape in `coverageCreator`. It uses
`InsMap` for child order and ordinary Go maps only for key lookup where no
iteration order is observable. Because Go's semantic nodes store an operator
symbol rather than Java's `SymbolNode` subclass hierarchy, the coverage creator
resolves `OpDefNode`s through `Tool.Lookup` and `Context.Lookup`; the parser
front-end must preserve recursive flags and operator-argument nodes for full
Java-equivalent coverage trees.

The TLC bridge now transfers module and LET RECURSIVE declarations to
OpDefNode.InRecursive. LAMBDA arguments produce separate LAMBDA definition
symbols and OpArgNodes, with fresh parameter identities scoped during body
conversion. Coverage visits an OpArgNode's definition as SANY walkGraph does,
so the higher-order body links resolve to the actual lambda expression wrappers.

ENABLED evaluation carries the current cost model into the enabled interpreter;
the convenience overload without coverage remains separate. Anonymous actions
created while splitting a SPECIFICATION retain Java's unnamed marker, and their
string labels use the predicate's semantic location (with numeric literal-node
string overrides preserved).

Named function definitions use `$NonRecursiveFcnSpec` or `$RecursiveFcnSpec`
instead of the `$FcnConstructor` used by bracketed `|->` expressions. Recursive
definitions have a fresh self symbol in the body and the constructor's unbounded
symbol list; the evaluator binds it to the function lambda and keeps evaluation
lazy. Both specification constructors span the full function definition.

INSTANCE clones in the TLC bridge retain their original definition bodies under
SubstInNode wrappers. Each instance shares its Subst identities and replacement
expressions across the cloned definitions; each Action coverage root owns its
wrappers for those shared substitutions. Explicit and implicit substitutions bind
the instancee declaration symbols through lazy values, while operator parameters
use OpArgNode. Instance parameters have their own scoped symbols. EXTENDS retains
root declaration identity. Definition conversion tracks actual module membership,
including inner modules whose physical source file belongs to the outer module.
The parser retains the instance definition LHS range for action declarations.
Imported operator definitions (including native module aliases) share a canonical
symbol and original definition body. Alias names remain in the tool/processor
definition tables. Replacing a module definition therefore reaches references
through EXTENDS and instance imports without rebinding unrelated definitions.
LET definitions use fresh scoped symbols carrying the original OpDefNode for
default lookup; contextual lazy values still take precedence during evaluation.
Coverage walks LET definitions in SANY Context Hashtable order: Java UTF-16 name
hashes, descending buckets, head-first chains and source capacity/rehashing.
Evaluation retains declaration order.

Coverage expression labels come from semantic source locations. Standard-module
exclusion follows Java's fixed set of nine source-module names, including its
name-only behavior for replacement modules; a built-in opcode in a user's
module remains covered. Primed wrapper membership compares source locations,
matching OpApplNodeWrapper.equals. ActionWrapper.get resolves its predicate to
the child expression wrapper and unwraps SubstIn/LetIn bodies; only expression
wrappers return themselves for their own expression. Zero-count primed parents
retain the extra indentation before reporting their children.

Constraint evaluation resolves its per-tool Action metadata and passes that
Action's cost model through evaluation. Rejected constraints increment the
primary counter; accepted constraints increment the secondary counter. Action
validity checks likewise pass the Action's cost model, preserving invariant and
implied-property expression coverage.

The parser bridge preserves SANY's n-ary `$ConjList`/`$DisjList` applications
for bullet lists, flattening only binary AST folds within the same junction
frame. Nested junctions keep their own nodes. Single-item lists keep a unary
list application. Boolean literals retain predefined TRUE/FALSE operator
applications, allowing LET bodies and boolean expressions to be covered at
their actual source locations.

Coverage counts:

- expression/action hit counts,
- secondary action counts for newly discovered successor states,
- zero/non-zero coverage,
- cost coverage.

Java has two knobs. `TLCGlobals.isCoverageEnabled()` is driven by the
`-coverage` interval and enables full expression/action coverage.
`TLCGlobals.Coverage.coverage` is a bitmask read from the
`tlc2.TLCGlobals.coverage` system property: bit `1` enables action coverage and
bit `2` enables variable coverage. `Coverage.isEnabled` is true for either
interval coverage or any bit.

The guards are intentionally not interchangeable. `Spec.coverage` and
`Value.coverage` are plain `TLCGlobals.isCoverageEnabled()`; they guard
`cm.get(...)`/`cm.getAndIncrement(...)` in `eval`, `enabled`, `next`,
`processUnchanged`, substitution cost lookup, and `LazyValue` construction.
`ActionItemList.coverage` is `Coverage.isActionEnabled()`; it guards delayed
action-list cost traversal and action `incInvocations`/`incSecondary` counts.
Java's worker and simulation functor paths count action invocations one
successor at a time through `addElement`, but Java's direct bulk
`tool.getNextStates(action, state)` path increments `action.cm` by the generated
`StateVec` size when `Spec.coverage` is enabled. Keep those paths separate in
Go to avoid double counting simulation, which uses a bulk vector internally.
The Go port mirrors this with `CoverageInterval`, `CoverageFlags`,
`CoverageAnyEnabled`, `CoverageActionEnabled`, and `CoverageVariableEnabled`;
`TLAGO_COVERAGE` is accepted as the Go-friendly environment spelling of the
Java bitmask property.

Variable coverage is a separate count-distinct path. During cost-model
creation, Java installs a `CountDistinct.SyncedHyperLogLog(10)` on every state
variable declaration. `Worker.addElement` updates those counters only when an
unseen in-model successor is enqueued, and coverage reporting plus
`TLCGet("variables")` expose the resulting distinct-value estimate. The Go port
stores the same counter on each concrete `StateVariable` and updates it at the
same successor enqueue point in `processSuccessorForWorker`.

Port guidance:

- Do not wire coverage into the first evaluator pass unless the code shape
  makes it cheap, but reserve fields in actions/values to avoid later invasive
  edits.
- End-to-end test parity eventually requires coverage output.

## Debugger Architecture

`tlc2/debug` implements Debug Adapter Protocol support:

- stack frames for init, next, action, state, synthetic states.
- source breakpoints.
- debugger expression evaluation.
- scoped identifier discovery.
- goto-state events.
- attach/suspend/halt behavior.

Important Java classes:

- `IDebugTarget`: the central control surface used by `Tool` and stack frames.
  It defines the stepping enums `StepDirection`, `Granularity`, and `Step`, and
  a large family of `pushFrame`/`popFrame` overloads. In Go, avoid turning this
  into an interface while there is only one debugger path; keep the enum values
  and concrete debugger state on `TLCDebugger`.
- `TLCStackFrame`: the base debugger frame. It carries semantic-node identity,
  context, tool, optional exception, parent, and eventual value. Java uses it
  for stepping/breakpoint target checks and for `ResetEvalException`; Go mirrors
  that as concrete `TLCStackFrame`, `ResetEvalException`, and
  `AbortEvalException` structs. Java also keeps nested variable-reference
  caches on the frame; Go mirrors those as insertion-ordered maps and exposes
  concrete `TLCScope`/`DebugTLCVariable` slices for `Context`, `Constants`, and
  `Stack` scopes without pulling in DAP transport types.
- `TLCStateStackFrame` and `TLCActionStackFrame`: Java specializes frames for
  state and action evaluation. The state frame's `getS/getT` both return the
  captured state; the action frame's `getS` returns the predecessor and `getT`
  returns the successor. Go mirrors those as embedded concrete structs and keeps
  the pending debugger value as `"?"`. The state/action scope split is explicit
  because Go embedding is not Java virtual dispatch: action frames provide their
  own `Action` and `Trace` variable accessors rather than relying on inherited
  state-frame methods to override themselves.
- `TLCSyntheticStateStackFrame`: a manually inserted marker frame for trace
  display. It is still a state frame, but it also stores the successor used when
  evaluating expressions against a trace edge. Java creates these frames lazily
  in `TLCDebugger.stackTrace`, not when the worker halts, so BFS can still evict
  ordinary states if the front-end never asks for a call stack. Go mirrors this
  with `TLCStateStackFrame.GetTraceAsStackFrames` and `TLCDebugger.StackTrace`:
  stack traces are empty unless execution is halted, the first stack-trace
  request reconstructs synthetic frames from simulator traces, predecessor
  links, or `ModelChecker.GetTraceInfo`, and resume/step/goto/disconnect
  commands remove the synthetic frames again. Because Go stores the stack with
  the top at the end rather than Java's `LinkedList.push` head, synthetic trace
  frames are prepended internally in reverse order so the observable response
  remains Java top-to-bottom order.
- `TLCInitStatesStackFrame` and `TLCNextStatesStackFrame`: Java debugger frames
  that expose generated initial or successor states. Both keep a
  variable-reference-to-state map because DAP variable references are integers
  and partial states may not have fingerprints. Go mirrors this as concrete
  structs over `StateFunctor` and `NextStateFunctor`. Initial states sort
  lexicographically by state text; successor states cluster by action location
  before state text. Step-in chooses the successor with minimum string Hamming
  distance; step-over chooses maximum; step-out selects the predecessor or halts
  the functor. The Go frames expose `Initials`, `Successors`, and `Trace`
  scopes directly, maintain the variable-reference-to-state selection maps, and
  apply conditional breakpoint expressions against the generated states.
- `TLCDebugger`: owns breakpoints, exception-breakpoint filters, the active
  stack-frame list, stepping state, granularity, halt flags, and the connection
  to the debug adapter. The Go port keeps a concrete `TLCDebuggerFrame` union
  rather than an `IDebugTarget` interface; each pushed frame preserves its
  concrete kind while exposing the embedded base `TLCStackFrame` for common
  push/pop, source-frame, and breakpoint operations. The protocol transport is
  less important than preserving where model-checker/evaluator state is
  captured. Exception, unsatisfied-state, invariant-violation, and
  assumption-violation hooks duplicate the active frame just like Java; the Go
  stack layer accepts ordinary `error` values until the Java
  `StatefulRuntimeException` hierarchy has a full concrete Go mirror.
  Conditional breakpoint parse/semantic failures are reported back on the
  breakpoint object, but Java leaves the compiled `condition` field null; a null
  condition does not suppress a later location match. Go mirrors that by letting
  a breakpoint with no compiled condition op keep the incoming `fire` value.
  Continue/step-over/step-in/step-out/step-back/reverse-continue/goto-state
  commands update `Step`, `SourceFrame`, `Granularity`, and generated-state
  selection in the same place as Java's DAP handlers, while leaving the protocol
  transport itself for a later pass. Java initializes `step` to `In` so the
  first ordinary pushed frame can halt; `nosuspend` instead selects Continue
  and `nohalt` disables exception/invariant stops. The factory override preserves
  a debugger installed before TLC starts. `MaybeHaltExecution` decides ordinary
  stops; `HaltExecution` handles already-decided stops such as exception, spec,
  unsatisfied and violation breakpoints. Go now uses a condition variable on
  the debugger monitor: the worker publishes a stopped event, releases the
  monitor while waiting and resumes on continue/step/disconnect notifications.
  State-selection frames switch granularity around the halt. After resume,
  synthetic trace frames are removed and reset commands unwind to the exact
  target frame; frame cleanup covers push as well as evaluation. Stepping out
  retains Java's Exit frame pause and subtle/normal presentation. Frame IDs
  combine actual semantic UIDs with a random integer because repeated semantic
  nodes may occupy multiple frames. Parser syntax nodes retain depth and
  parent links, and semantic expression nodes retain those exact syntax objects;
  frame matching compares syntax depth and owning operator-definition identity.
  One-item junctions retain their list/item structure, and all variables in one
  source quantifier share one semantic application with the original bound
  groups. Generated Init state frames reuse the active evaluation metadata.
  Concrete frame dispatch preserves state/action hit counts, conditional
  expressions, ancestor suppression, and action/next trace getS/addT behavior.
  Complete live test assertions and attaching DAP transport/capability events
  still require source work.
- `TLCDebugger` breakpoint ownership is concrete state on the debugger:
  `breakpoints` keyed by module/source, boolean exception and invariant halt
  flags, and two conditional `TLCSourceBreakpoint` values for the Java "after
  Init/Next" and "unsatisfied" filters. The Go port mirrors Java's filter IDs
  and multi-worker warning as plain `TLCExceptionBreakpointFilter` values, not
  DAP protocol objects. Source breakpoint setting derives the module from the
  source name, replaces the module's breakpoint list, returns concrete
  `TLCBreakpoint` verification records, and keeps Java's parent-frame
  suppression rule: if an ancestor stack frame already matches the same source
  breakpoint, a nested frame does not fire it again. Verification is fuzzy like
  Java: the breakpoint is always stored, but the response is marked verified
  only when a same-line semantic node is found under a source range that
  `Location.includes` the breakpoint range. Go now walks the actual retained
  ModuleNode's SemanticChildren, matching Java walkChildren's unconditional root
  visitation and child preemption. INSTANCE modules retain their own source
  assumptions for location lookup, without adding them to TLC's checked root
  EXTENDS closure. Verification preserves the deliberately non-point range with begin line
  `line + 1` and end line `line`. Breakpoints in modules that are not part of
  the debugged spec remain verified, matching Java's `moduleNode == null` case.
  A hit condition on the configured `Next` predicate is reported as unverified
  with Java's "A Next breakpoint does not support a hit condition." message.
- `DebugTLCVariable`: adapts TLC `Value` objects into debugger variables.
  Scalars expose `type` and `value`; enumerable/function/record/tuple values
  receive a non-zero `variablesReference` and lazily produce children.
- `TLCSourceBreakpoint`: stores source line/column, optional hit count, optional
  log message, a parsed condition operator, and the source `Location`. Location
  matching succeeds for `nullLoc`, otherwise it checks equal line and breakpoint
  column less than or equal to the semantic node begin column. Conditional
  breakpoints evaluate through `tool.noDebug().eval` and swallow evaluation
  failures so a broken debugger expression does not crash TLC. Named operator
  conditions resolve by exact name in the selected semantic module. Other
  nonblank conditions go through `Tool.ParseDebuggerExpressionFunc`, installed
  by the root `tlcBridge`; it mirrors Java `TLCDebuggerExpression.process` by
  building a synthetic `__DebuggerModule__N` module that `EXTENDS` the selected
  module. Frame expression requests retain the processor root. The wrapper
  defines `__DebuggerExpr__N == <condition>`, parsing/checking it through the
  production SANY path, and converting the generated operator back into TLC's
  semantic node graph. The hook is source-location aware, matching Java's API
  shape. The bridge now collects parameters and LET operators from the retained
  semantic child path, including formal arities and quantifier/operator-argument
  parameters. Child order follows the represented SANY getChildren methods;
  OpAppl bounds precede operands, and LET definitions precede its body. The root
  of pathTo is visited without preemption, as Java walkChildren does. LOCAL
  wrapper stubs reconnect to the actual LET definitions and symbols so captured
  lazy bindings and operator parameters remain usable. Expression conversion
  reuses the existing source graph and native/config bindings. Source parse and
  represented semantic/level diagnostics keep Java messages and locations.
  Missing dependencies resolve through the default filename resolver and actual
  parser dependency list, recursively extending the live external module table.
  Successful dependencies persist after later expression failures; transient
  wrappers stay unpublished and leave the original root in place. Incremental
  contexts reuse existing source/config/native/lazy identities. Constant graph
  processing uses the native-only snapshot, followed by represented native
  module overrides and then LOCAL-stub substitution through ModuleNode's graph.
  Full level/theorem/proof/instance metadata and complete native override
  registration remain source work.
- `TLCCapabilities` and `GotoStateEvent`: small protocol data types. They are
  useful in Go as plain structs even before a debug-adapter server exists.
  `TLCDebugger.InitializeCapabilities` mirrors Java `initialize`: goto-state is
  enabled only while a simulator exists, hover evaluation, terminate requests,
  exception-filter options, hit-conditional breakpoints, conditional
  breakpoints, step-back, and clipboard context are enabled; exception-info,
  log points, value-formatting options, stepping granularity, goto-targets,
  data/function/instruction breakpoints, and disassembly remain disabled. The
  returned exception-breakpoint filters are the same Java filter IDs plus the
  multi-worker warning when `NumWorkers() > 1`.
  `TLCStackTraceArguments`/`TLCStackTraceResponse` likewise model the Java
  request surface without importing DAP transport types; `startFrame` outside
  the active frame range returns an empty response, and positive `levels`
  restricts the returned slice exactly like Java's `StackTraceArguments`.
  The small Java request handlers with no deeper TLC semantics are also kept
  concrete: `threads` returns one thread `{0, "worker"}`, `setVariable` returns
  an empty response, `configurationDone` is a no-op, and `terminate` stops the
  active checker/simulator before running the same disconnect cleanup path.

Debugger variable details:

- Base-frame variable rendering and synthetic trace reconstruction run in the
  source DebugEvalDebugger mode; displaying lazy function values and rebuilding
  disk-backed traces must not step recursively into the debugger. The supplier
  boundary temporarily installs FastTool for state fingerprint evaluation and
  restores both tool and mode. DebugTool's one-state Eval overload selects State
  mode (including VIEW), while its context-only overload selects Const mode.
  Workers compute a successor fingerprint once and reuse it for state collection,
  preserving VIEW evaluation count and debugger stops. Java creates FastTool
  first and the DebugTool copy constructor leaves the parameterless state
  fingerprint tool on FastTool; Go preserves that initialization and keeps the
  original tool mode on the fast copy. TraceVariables reconstructs the disk
  prefix and predecessor suffix, excludes the action successor and adds the
  current next-state state. Simulation filters incomplete states; next-state
  simulation traces use padded names and retain selection IDs. State renderers
  remember nested references and report the current getT fingerprint, including
  source exception fallback, even for older records in the Trace scope. Paired
  action records interleave variable names with a trailing space and their
  primed names. Pending values print ? without quotes and carry the source
  evaluation-pending type. Concrete context lazy evaluation uses getS/getT as
  appropriate and does not call the cache-mutating GetValue.
  The complete original Echo debugger test, including its inherited assertions,
  now passes in Go normally and under race instrumentation; unchanged Java
  passes too. The complete ExpressionBreakpointTest is translated too with
  its source context-map and lazy-cache assertions. Concrete expression/getWatch
  methods preserve nullable requests/results, supplier evaluation, base/state/action/
  synthetic state dispatch, source exception catches, parameter-name lookup,
  TLCExt CounterExample injection and nested references. Stack names use the
  actual syntax tree's human-readable image (including its zero/one quirk) and
  source Variable equality. The EWD998 live probe matches 211 expression and
  156 getWatch responses plus 21 stack-variable records across eight stops/52
  frames under race instrumentation. Hover and protocol Evaluate dispatch now
  follow Java's nullable response/request behavior, concrete frame state/prime
  lookup, lazy fallback and source exception classes. Hover URI parsing retains
  Java path/query/fragment masks, percent escapes, coordinate splitting and
  missing-fragment/module/index failures. Ordinary formal and bound parameters
  retain fresh symbols, source locations and actual parser syntax; all bounds
  convert before entering the quantified scope. Lookup preserves a source
  definition before a same-name INSTANCE alias. Base-only NullPointerException
  catches depend on frame type rather than the EmptyState sentinel.
  The corrected live hover comparison matches five stops/79 frames, 120 paths,
  16 responses, 171 breakpoint line-verification results and 21 URI cases under
  race instrumentation. Empty/syntax-invalid and missing DoesNotExist dependency
  conditions agree, including the exact located semantic error. Final comparisons
  match all 419 hover/breakpoint and 448 expression/watch/stack records, plus 60
  dynamic dependency records. Retained modules, Bags/Randomization overrides,
  original root identity, transient-wrapper exclusion and radix/large/decimal
  diagnostics agree. The entire EWD998ChanDebuggerTest is now translated with
  every original assertion and four byte-identical model/config vectors. The
  inherited Set<Variable> overload preserves displayed name/value/type checks
  and lazy-cache counts without invoking the separate Context overload's state
  assertions. Context maps and displayed names use declaration names while the
  evaluator retains qualified lookup keys. Semantic context bindings use source
  signatures and OpDefNode human-readable images: leftmost attached comments,
  one-child images separated by spaces, Java ASCII trim and no variable type.
  State constraints call the one-state eval overload, so debugger hit conditions
  use the current state's level; action constraints retain the two-state path.
  Unchanged Java and the complete Go method pass, as do full offline normal/race
  suites. The complete EWD998TraceDebuggerTest is also translated after verifying
  production parity: _TETrace hover preserves the exact two-state tuple, type,
  nested reference and expected liveness-violation exit. Its byte-identical
  embedded-config/multimodule source is frozen in test_vectors. The inherited
  test harness now retains each Java constructor's expected exit status; source
  success tests still require success. Unchanged Java, full normal Go and all four
  debugger model tests under race instrumentation pass. The complete original
  EWD840DebuggerTest is now translated after a production comparison matching
  624 frame/context/state records across 37 stops and Java's 127/38/3 statistics,
  depth 6 and safety exit. Initialization suppresses debugger evaluation until a
  checker/simulator exists. Source definition symbols retain declaration names,
  syntax and locations; lazy semantic images use SemanticNode.toString, including
  numeral/decimal overrides and showPlainFormulae. Unsupported evaluation retains
  the source detailed runtime exception's expression/context and located message.
  Invariant control exceptions are fresh objects with distinct initial/next-state
  classes. Worker wrappers expose their stored state/exception without setting a
  cause/message. Debugger catches use direct source exception types and actual
  base/action frame overloads; invariant handling sets each object's known flag.
  The whole test preserves initial-state/action/constraint/invariant/alias loops
  and assertions. Three original vectors remain byte-identical in test_vectors.
  Unchanged Java, the complete Go method and full normal/race suites pass.
  Both entire original EWD840 error debugger methods are also translated after
  source comparisons matching 31 initial-error and 61 action-error records under
  race instrumentation. Legacy exception variables use semantic human-readable
  locations, nullable detail messages, Java simple class names and no nested
  reference; a missing exception retains Java's NullPointerException boundary.
  Action.UNKNOWN keeps the source null semantic node with its builtin location,
  syntax image and minimum-integer kind. LET evaluation binds only zero-arity
  definitions as lazy values, leaving parameterized source operators on their
  semantic symbols. The harness checks the source pending exit sentinel -1 while
  still paused, then disconnects for cleanup. It refreshes the cached integer
  statics to match Java's isolated test classloader and prevent source metadata
  from a preceding CallStackTool run from changing the next exception stop.
  Source metadata reads/writes are synchronized to preserve Java reference
  atomicity when workers attach lazy sources to shared cached integer values.
  Two new original vectors remain byte-identical in test_vectors. Unchanged Java,
  both complete Go methods and full normal/race suites pass.
  The entire original EWD840DebuggerSimTest also passes after source comparison
  of all 911 frame/exception/context/state records across 39 stops under -race,
  35 generated states, one trace and safety exit 12. SimulationWorkerError extends
  the invariant exception and initializes its own inherited known flag. Direct
  catch dispatch recognizes the subclass and retains its formatted error-code/
  parameter message and separate stored exception, restoring the invariant stop
  before alias evaluation without duplicate handling. All construction paths use
  the initialized source class. The complete test preserves conditional spec
  breakpoints, synthetic trace levels, action/state/constraint/invariant/alias
  assertions, all three loops and original simulation/config/seed/fingerprint
  arguments. One more original vector is byte-identical in test_vectors.
  Unchanged Java, the whole Go method and full normal/race suites pass.
  The complete original Debug03Test and Debug03SimTest are also translated
  after matching 17 checker and 572 simulation frame/context/state/initial/
  successor/trace records. Existing production behavior already matches this
  slice: nine sorted successors, selection by variables reference, forward and
  backward trace construction, stepping out to the initial-state chooser and
  selecting its second state. All four simulation loops and all source assertions
  remain intact, including inherited state/next/synthetic helpers. The shared
  harness mirrors gotoState's stop synchronization. One original embedded-config
  vector is byte-identical in test_vectors/models/debug. The source checker has
  92/10/0/depth 2; simulation generates 1,055 states and two traces. Both unchanged
  Java methods and both whole Go translations pass under the represented runtime.
  The complete original Debug02Test also passes after comparison of all 24
  hover/frame/state records under -race. Source exact-location lookup distinguishes
  a point in a primed application from its full range; pending assignments retain
  null response types and "?", while assigned booleans retain source type strings
  and TRUE/FALSE. All 51 equality/8 true/4 false assertions, constant-module view,
  step commands and next-frame helper remain intact. The source hover harness
  preserves absolute module URI, symbol query, coordinate fragment and top-frame
  identity. The original embedded-config vector is byte-identical in test_vectors.
  Both source and production Go complete successfully with 3/2/0/depth 2.
  The whole original Debug04SimTest passes after checking its complete production
  sequence under -race. StepIn chooses a minimum-distance successor; stepOver
  chooses a maximum-distance successor. Java uses HashSet iteration to resolve
  equal distances; source tests require adjacent values to differ, without
  specifying which tied maximum wins. Of 1,410 records, 1,288 match exactly and
  122 contain choice-derived state/trace/expression/fingerprint differences.
  Each trace and expression remains consistent with its chosen state. Both
  runtimes generate 543 states and four traces and preserve all original
  assertion sites, five loops, 215 expression evaluations, state assignment,
  idempotence and action/ENABLED/level breakpoint behavior. The inherited stepOut
  count overload is translated; the embedded-config fixture is byte-identical
  in test_vectors. No production correction was needed for this slice.
  Debug05's complete production sequence matches all 950 frame/context/state/
  variable/expression records under -race, including 21 stops and 63 evaluations.
  Named LET instances keep WITH substitutions through the parser, standard-module
  declaration enumeration, cloned SubstIn bodies and original module-context
  symbol identities. Spec.getCounterExampleDef now reads shared live definitions
  after dynamic imports, so the no-debug tool observes the same CounterExample
  context as the debugger. All nineteen JSON payloads agree. Both Java and Go
  readers decode all 38 cross-runtime binary exports to the same original trace
  values; serialized string tokens and normalized field order are runtime-local.
  The complete original Debug05SimTest translation preserves its full export/
  readback loop, all assertions, no automatic dumpTrace and explicit resolver
  overrides. The original fixture is byte-identical under test_vectors. Both
  runtimes generate 51 states/one trace/depth 25 and finish successfully.
  Broader semantic metadata and full debugger completion remain source work. Raw displayed
  fingerprint numbers use each runtime's interned tokens; matching record
  values and fingerprint type presence do not establish numerical equality.
- `TLCDebuggerExpression.getScopedSymbols` now has a shared Go helper used by
  debugger expression construction and `GetScopedIdentifiers`. It collects
  identity-distinct LET definitions, local operator/formal parameters, quantified
  variables and LAMBDA parameters on the inner-to-root path. Top-level definitions,
  constants and variables are inherited from the root module. Generated parameter
  signatures retain arity, and LET names are excluded before formatting because
  their LOCAL stubs are substituted with source definitions. The complete
  GetScopedIdentifiersTests translation retains all eighteen original inputs,
  expected sets and four assertions per case, including the source's known
  infix-parameter quirk. All eighteen production records match Java exactly.
- The complete DebugTLCVariableTest translation retains all four original methods,
  eight equality sites and the nested tuple's two-child loop. Production Go and
  Java agree on all nine expansions and five child records (name, type, value,
  reference presence); numeric random references are not compared. Empty and
  infinite sets yield no children. Existing nested-value production behavior
  required no correction. Both whole test classes and existing debugger expression
  tests pass under -race. Full semantic/protocol parity remains work.
- Java's `Value.toTLCVariable` sets type to
  `<ValueClass>: <kind string>` and value to `toString()`.
- `StringValue` replaces quoted `toString()` output with the unquoted display
  string for debugger variables.
- `TupleValue` children are named with zero-padded 1-based indexes.
- `RecordValue` children are named by record field.
- `FcnRcdValue` children are named by domain element.
- `SetEnumValue` children are named by element string.
- Infinite or non-finite values must not be eagerly expanded.

Port guidance:

- Defer debugger protocol until core CLI TLC is functional.
- Preserve enough state/action metadata in core structs so debugger support does
  not require a second state model later.
- Keep debugger structures concrete. Do not port Java's `IDebugTarget` as a Go
  interface unless a second real implementation appears.
- Simulation debugging attaches only one worker to the debugger, but Java still
  uses `ExplorationWorker` for every simulator worker when `tool.isDebugger()`
  is true. In Go this means every worker is still the same concrete
  `SimulationWorker`; its `debug bool` selects exploration behavior without a
  second worker type.

## Distributed TLC Architecture

Current direction (2026-10-08): port the original distributed TLC algorithm
using Go coordinator/worker/fingerprint-server boundaries. Leave the rpc25519
alternative aside. Java uses RMI, but the Go implementation must not emulate
Java RMI or a JVM. Preserve batching, fingerprint answers, retries, checkpoint
boundaries and termination behavior independently of transport.

The user's October 8 clarification also rejects Java runtime impersonation.
The current wire transport is native Go TCP/net/rpc. Retry, remote-failure,
worker-availability and exit decisions use native operation traits exclusively.
Java remote-exception hierarchy and nested-cause inference do not drive these
decisions. Source diagnostic carriers remain generic reporting compatibility,
without implementing an RMI protocol or Java serialization.

`DistributedWorkerEndpoint` supplies the five worker operations used by the
coordinator. Server registration, server threads, smart-proxy measurements and
shutdown use this interface. `LocalWorkerEndpoint` supplies the existing
in-process behavior. `DistributedServerEndpoint` supplies the coordinator's
settings, file, interning, registration and status operations to discovery,
worker bootstrap and fingerprint-server startup. The local coordinator adapter
pins the interning table before worker context initialization and returns owned
copies of string metadata. `DistributedFingerprintEndpoint` supplies storage
operations without local allocation/configuration methods. The manager consumes
returned failures as well as existing local storage exceptions. Worker manager
snapshots copy failover state and retain shared registration-wrapper identity;
the coordinator recovery trace is excluded. The first network endpoint uses
Go net/rpc over TCP for named fingerprint objects. The transport performs no
insertion retry. It preserves scalar/batch answers, full fingerprint bits,
null/empty vectors and checkpoint-file recovery; trace objects remain local to
the coordinator. Server shutdown closes listeners and client connections
without implicitly exiting storage. DistributedStatePayload is a native object
graph with one-based references (zero is null), explicit nil-array flags and
full int32 state levels. It preserves object sharing/cycles and copies raw
string metadata independently of the receiver table. Symbolic set types/caches
are retained; source custom serialization materializes functions/predicates
and sends only cached lazy values. Extended evaluator state metadata and opaque
custom user/data/operator types are explicitly rejected pending their remaining
implementation. DistributedResultPayload references state and fingerprint vector
nodes separately, retaining repeated vector identities and null/empty arrays.
All state partitions use a single state/value graph. Signed computation time
and states-computed counters are preserved without recomputation. Vector nodes
send active entries only; decoded capacity equals size, matching the source
TLCStateVec/LongVec serialization contracts.

NetworkWorkerEndpoint supplies the five worker calls over the same Go TCP host.
Request states and NextStateResult payloads use the shared graph codecs. The
host removes a named worker after successful exit; removal does not close other
connections or cancel calls that already captured the endpoint. No worker call
is implicitly retried. Native DistributedOperationError traits preserve the
coordinator's worker-loss, null-failure, block-size retry and shutdown decisions.
Failure nodes preserve shared/cyclic causes and suppression, nullable messages,
sender diagnostic class/stack text, worker error states and KeepCallStack. Only
the existing concrete WorkerException is reconstructed for the coordinator's
error-state path; general remote causes remain native Go operation errors.
NetworkServerEndpoint supplies all coordinator operations through a named Go
TCP endpoint. Worker/fingerprint registrations pass address/object references
and establish coordinator-owned callback connections. Fingerprint manager nodes
retain wrapper aliases, hostname, availability, mask and mode settings, omitting
the transient trace. Coordinator-local storage is published on the explicitly
advertised host listener; existing network endpoints retain their references.
Snapshot decoding reuses endpoints within each manager and closes newly acquired
connections if decoding/dialing fails. Successful snapshot connections belong
to the coordinator client; callback connections belong to the RPC host. Both
owners close their connections independently of worker/storage exit. Native
coordinator failures enter the existing file resolver and keepalive catches.
DistributedNetworkDiscovery probes binding presence with a dedicated native
lookup call, independently of coordinator settings/status. Connection refusal
and reachable/unbound bindings enter the original backoff loop; unrelated
connection/protocol errors propagate. Owned connections are reused for repeated
binding probes and discarded after transport failure. UnregisterCoordinator
removes a binding independently of listener lifetime.

DistributedWorkerNetwork configures discovery, worker publication and
registration boundaries on the existing worker command. A publication hook
sets the actual TCP address/object identity before storing the runnable worker
pointer, preserving construction-before-visibility. Unique native object names
allow repeated worker thread IDs without overwriting earlier publications.
Registration transfers a published reference without requiring a self-dial.
Bootstrap keeps polynomial/interner/resolver/app/manager initialization order.
Networking shutdown and worker runtime shutdown remain distinct.

DistributedCoordinatorNetwork.Publication connects the existing modelCheck
callbacks to native TCP publication. Listener creation stays at CreateRegistry,
after recovery and hostname resolution. Rebind replaces a named endpoint;
Unbind removes only that binding, and the shutdown hook's local Lookup observes
the same host publications. Unexport removes the supplied coordinator's aliases
without closing the host or cancelling calls that already captured an endpoint.
Process-owned Close releases the listener and connections, including the early
initialization-error path where modelCheck deliberately skips normal unbinding.
Repeated Close calls wait for the same shutdown result.

DistributedFPServerNetwork uses the same Go host as a worker network, allowing
combined roles to share listener/discovery ownership. FPEnvironment publishes
owned local storage before registering its native address/object reference.
The command removes this publication at its original rejection and reporting
shutdown points through an optional native hook. The source storage methods
remain intact: upstream FPSet.unexportObject is itself a no-op, so this hook owns
Go endpoint removal rather than changing generic storage or emulating RMI.
FingerprintRejected is an explicit operation-error trait preserving the source
FPSetManagerException catch and early no-flush return. Other failures propagate.
Remote storage exit uses the existing running flag and monitor notification;
the reporting loop retains its original five-minute wait. Removing FP endpoints
leaves a worker on the same host available. Remaining payload types and
process/CLI wiring still need
implementation and separate-process model verification.

The Java reference implementation uses RMI:

- `TLCServer`
- `TLCWorker`
- distributed FP set managers.
- block selectors.
- smart proxies.
- server/worker management beans.

Core Java flow:

1. `TLCServer` owns the state queue, trace, fingerprint-set manager, worker
   threads, error state, progress counters, and checkpoint lifecycle.
2. Workers register with the server. Registration wakes stuck server threads
   through `stateQueue.resumeAllStuck()` and creates one `TLCServerThread`.
3. `TLCServerThread` dequeues blocks of states, calls a remote `TLCWorker`, then
   merges the returned new states/fingerprints into trace, queue, and progress
   counters.
4. `TLCWorker.getNextStates` computes all successors for a block of predecessor
   states via `DistApp.getNextStates`.
5. Each successor fingerprint is checked against a worker-local `SimpleCache`.
   Cache hits are not sent to the FP-set manager, but the server later adds the
   skipped count through `addStatesGeneratedDelta`.
6. By default, remaining `(fp, successor, predecessor)` triples are sorted by
   signed fingerprint in a `TreeSet<Holder>`. Its comparison is fingerprint-only;
   two states with the same fingerprint collapse, retaining the first holder.
   The static `TLCWorker.unsorted` property instead uses holder allocation
   identity. Go retains deterministic insertion order for that mode.
7. Holders must pass Java's strict `last < fp` assertion, starting with
   `Long.MIN_VALUE`, before partitioning by `fpSetManager.getFPSetIndex(fp)` into
   parallel vectors of predecessors, successors, and fingerprints. This rejects
   a `Long.MIN_VALUE` fingerprint even in sorted mode; the source unsorted mode
   can also fail on its traversal order or duplicate fingerprints.
8. `fpSetManager.containsBlock` returns bit vectors whose set bits identify
   fingerprints not yet present.
9. Only those unseen states are checked with `work.checkState`,
   `work.isInModel`, and, only if the model constraint passes, `work.isInActions`.
   Passing states inherit the predecessor UID and return in `NextStateResult`.

Coordinator block selection requires its queue; null queue/selector access fails
instead of returning a completion block. After the first recorded thread error,
a nonnull predecessor requires trace printing even if the trace is missing.
The printing attempt catches ordinary failures and reports that failure, while
fatal errors escape. Queue FinishAll follows reporting and is required; failure
there prevents completion notification and retains earlier model-error fields.
Run’s separately registered finally still executes final cache read, keepalive
cancellation and assigned-state clearing under their existing catch boundaries.

Worker construction retains the supplied fingerprint manager without a fallback.
A nil manager fails at the partition-count lookup after completed generation and
the overall-state counter update. WorkerException retains the current predecessor,
null successor and call-stack flag, and finally clears computing. Generation
failures occur first and do not update the overall counter. A real empty manager
is a separate source case; zero-successor work can return empty partition arrays.

The concrete `TLCApp` runtime captures implied-init, invariant, implied-action,
and next-action arrays at construction, retaining their identity even if the
tool later replaces its arrays. Worker-group members share one application.
Generation and property loops read the application’s current array length and
entry at every index. Replacing those public application arrays during evaluator
callbacks therefore changes subsequent traversal: shrink stops it, replacement
uses the current entry, and growth includes added entries. This is separate from
replacing the tool’s arrays, which leaves the captured application arrays intact.
Successor generation calls the vector overload separately for each captured
action. `StateVec.addElements` chooses the larger vector as receiver, so a
later action producing more states can move its successors before earlier
ones. The application allocates a separate fixed-size successor array before
validation and copies each captured state after its complete-assignment check.
Returned storage is isolated from tool-vector replacement/clear/growth, while
state objects remain shared. Validation still reevaluates the accumulator size:
shrink retains the result array’s null tail, and growth beyond its original
length fails at the array write after validating the added state.
Deadlock and complete-assignment checks follow that combined order,
before worker fingerprinting. Property checks use captured action arrays and
the current tool's name arrays; state/alias reconstruction and call-stack
replacement delegate to the application tool. Server initialization uses the
same snapshot runtime.

The root `LoadTLCApp` constructor performs config parsing, SANY and semantic
checking, then tool construction through the production bridge. Both local and
remote loading use `LoadOptions.FilenameResolver`. It retains the constructor's
raw root/config names and lexical spec directory, with the source post-SANY
module-table lookup (direct constructors retaining a `.tla` suffix fail).
The parser records logical module filenames separately from physical temporary
paths; `Tool.GetModuleFiles` resolves them with `isModule=false`. The application
creates a fresh InJar resolver for enumeration, using bundled standard/model
resources and the process classpath. Enumeration is deterministic parse order
in place of Java's Hashtable traversal. Constructor flags and FP configuration
retain their source identity; basic applications leave metadata/FP configuration
unset, while `LoadTLCAppWithMetadata` completes metadata setup after tool loading.
The full constructor marks recovery from any non-null checkpoint path, including
an empty one, but does not itself read the intern table: Java's `create` does
that before constructing the tool.

`FileUtil.makeMetaDir` now has the concrete `MakeMetaDir` port: verbatim
checkpoint return, global metadata-root override, spec-directory `states`
fallback, local-time timestamp formatting and the milliseconds property.
Exclusive creation creates parents first, then atomically creates the requested
directory; a collision returns an absolute temporary sibling. Lexical dot
segments are retained, including the NIO parent-creation fallback that can
create normalized parents yet fail the original final path. Checked creation
failures become source code 2163 without an I/O cause. The regular TLC command
uses the same directory-creation algorithm with its existing error-return API.
All three upstream FileUtilTest cases are translated after implementation.
The root `CreateTLCApp` now performs the source option loop and startup branches.
Options preserve nullable names, raw path separators, Java numeric syntax and
source diagnostics/continuation quirks, including checkpoint int overflow.
Normal startup recovers the interner before FP64 and config/tool creation.
Packaged startup uses `ModelInJar` resource presence, config copying and generated
properties through the same classpath adapter; it ignores explicit config,
skips interner recovery, forces tool mode and disables checkpoints. Its FP
configuration was captured before these late properties. Properties loading
uses Java Latin-1, logical continuations, Unicode/other escapes and partial
checked-I/O results. Late properties leave already-loaded globals intact.
The probabilistic Tool property freezes on first class initialization. Actions
retain their backing slice, so applications observe in-place mutation as Java
arrays do, while replacement leaves the captured array intact.

Value fingerprints extend each kind tag with a byte, matching Java overload
selection; lengths, numeric indices and model-value tokens extend with ints.
Actual-source comparisons cover all polynomials and the recovered initial state.
Registry/export/RPC and full distributed command/error/exit wiring remain
pending. Go library loading currently returns config failures to its caller;
Java ModelConfig's process exits need the source command lifecycle adapter.
Broader JVM class-loader/provider/default-resolver initialization is separate.

`NewTLCServerFromApp` now represents the source application constructor. It
retains the application, uses DiskStateQueue directly, opens its trace before
factory FPSet initialization, and wraps that local set with a canonical-host
non-distributed manager. Checkpoint names come from the metadata leaf substring,
while the trace and FPSet root names come from the application. Queue/trace
paths retain source separator concatenation and symlink/dot behavior. Missing
recovery directories fail at trace open; basic apps retain null-metadata failure.

`NewDistributedFPSetTLCServer` folds the subclass into the concrete server.
Its manager is created during base construction with the static expected count;
its readiness latch then takes the explicit constructor argument. Registration
and countdown are synchronized, rejected registrations do not release the latch,
and manager retrieval waits for it. The startup wait hook prints the expected
count and also runs for a zero latch. ModelCheck invokes it before init generation.
The port/report/expected-count/veto properties freeze at first server class
initialization, while explicit port assignment remains mutable. Close follows
trace, FP manager, then recursive metadata cleanup unless vetoed. Actual-source
constructor/registration/cleanup comparisons pass, including mismatched counts.
Recovery and initialization failure reporting/replay now follow modelCheck,
including recovery outside the init catch, deferred callback Exceptions, escaped
Errors, the already-done replay quirk and early-done cleanup. The upstream
DistributedDoInitFunctorEvalExceptionTest runs through the production Go parser,
application and server; its fixture lives in test_vectors/.

The normal completion path now waits for later worker registration instead of
rejecting an empty worker map or inventing a no-action exit. It uses the source
static REPORT_INTERVAL with monitor notifications, including the unconditional
first wait, zero/negative timeout behavior, checkpoint-before-done ordering and
progress baselines updated after waiting. Server synchronization is reentrant
for recorder callbacks; wait releases and restores the complete acquisition
depth. Worker joins/statistics/exits precede executor shutdown and result
capture. Unexpected exit failures escape without extra close calls; only the
source remote exception families are ignored. Final distinct-state results are
shared across instances, success uses that snapshot, and failure completion
switches the application to CallStackTool. Reported worker violations return
normally; close failures escape even after a violation. Registry publication and
removal now follow modelCheck through concrete naming/export/flush boundaries.
The primary binding is published before FP readiness and initialization, while
the worker binding appears only after successful init/recovery. Early-done init
retains the primary binding. Completion flushes after FINISHED, closes, removes
worker then primary bindings and unexports with force=false, ignoring its boolean
result. Exceptional paths stop at the source phase without extra cleanup; the
registry itself is never unexported by this method.

RunWorkerShutdownHook skips empty registrations, performs one registry guard at
the current mutable port and preserves source checked catches before direct
worker exits. It retains registrations and propagates unchecked failures.
TLCRegistryNamespace supplies local catalog/binding lifetime and lazy references.
Duplicate configured keys, including zero, fail without replacing the catalog.
It has no wire listener, serialization, socket occupancy detection or actual OS
ephemeral-port assignment. Separate namespaces isolate names only;
FP64/interner package globals still need process/runtime isolation. The translated
init test uses its own naming namespace like the source per-test JVM fork.
Native endpoint publication, active calls and command/exit checks are tracked
below. Java constructor autoexport and registry internals are not Go completion
requirements. The four original DistributedTLCTestCase remote harness methods
remain pending; upstream disables that harness for OffHeapDiskFPSet.

Distributed discovery now has source worker and FP-server lookup loops with
native discovery and delay adapters. Connection refusal and missing bindings
retry through explicit Go operation traits. Missing bindings carry a distinct
BindingMissing trait without pretending to be a connection/I/O failure; lazy
local references distinguish absent catalogs from absent bindings. The source signed
int backoff counter wraps, producing zero sleeps after its negative/zero phase.
The root worker startup adapter joins this discovery to production tool/group
loading. Local catalogs and TCP publication use native failure categories for
duplicate creation, missing lookup/unbind and repeated coordinator removal.
The shutdown guard skips worker traversal for missing coordinator bindings;
keepalive follows its coordinator-loss path. Native TCP discovery/publication
and worker model checks are available; the four assumption-disabled upstream
remote harness methods remain separately pending. No dedicated upstream lookup
tests exist. Distributed algorithm, checkpoint and trace ownership parity remain
separately incomplete; generic source diagnostic carriers do not drive transport
decisions.

DistributedFPSet.main now has a concrete native RunDistributedFPServer boundary.
It preserves ToolIO/System stream separation, argument early returns, two clock
reads, raw temporary metadata path, ratio-one configuration/two nested FPSets,
factory initialization and class diagnostics, hostname timing, registration
rejection/unexport and skipped flush. Other Throwables record GENERAL before
the failure line and final flush. Registration, lookup, hostname, clock and
Object.wait can supply native provider boundaries; normal registration uses
the actual local server. MP console rendering and remote wire calls remain
separate from this recorder/stream implementation.

The report loop owns the FPSet monitor, reports signed long size and waits five
minutes. Shutdown sets the static running flag without waking it; another main
invocation does not reset that flag. Base FPSet.exit sets it before synchronized
notify, which wakes one waiter. Concrete sets carry their inherited wait state;
memory sets reuse the same reentrant monitor as their synchronized operations.
MultiFPSet init initializes children concurrently, ignores returned replacement
objects and wraps checked I/O. Worker RuntimeExceptions retain the ForkJoin
cause-copy behavior; general JVM pool scheduling/provider details remain
separate. Disk init retains lexical paths, explicit prefix mkdirs and coded
file-open failures. File.mkdirs tries the original spelling before canonical
parent creation, so missing components before .. can still fail at file open.
All sixteen upstream MultiFPSetTest methods are translated after implementation,
including identity partitioning, sign-bit collisions and the OffHeap guard.
Twenty actual-source command observations and race checks pass; parallel error
ordering/selected child is allowed to vary. Multi/Noop statesSeen uses atomic
loads/stores for snapshots while retaining the source separate read/modify/write.

A two-FP-server native integration loads the real parser/tool, registers both
sets, checks five distinct states and wakes both report waits on close under
-race. This harness loads worker classes before evaluation and registers after
init. Arbitrary concurrent native worker loading/server evaluation still races
shared FP64/interner globals; Java isolates those globals by JVM process. The
source raw init enqueue also assumes its startup phase. Process/runtime class
isolation, transport, concrete FP process exit/completion messages and full
standalone distributed command/process integration remains unfinished.

Worker registration first requires and resumes the coordinator queue, before
contacting the worker or constructing a server thread. A missing queue returns
the source null failure through the Go endpoint contract. A wakeup panic escapes
with its original failure and releases the coordinator monitor; neither path
creates threads or keepalive timers. Native TCP preserves the failure category
and leaves the coordinator host usable.

Worker registration is keyed by server-thread identity, not URI or worker
identity. Java can register the same worker more than once; each registration
has its own assigned block, statistics, keepalive, and removal. Go keeps one
protected `InsMap[*TLCServerThread, *DistributedWorker]`, snapshots thread keys
for iteration, and removes only the chosen thread. The URI remains display
metadata. Worker counts and block selectors use the number of registrations;
new-state counts sum queue size and each registered thread's entire assigned
block, including remote calls waiting at a worker's synchronized computation.

Registration and new-state counting preserve Java's monitor serialization.
Per-worker successor computation is serialized as well. Thread worklist
references use atomic publication for timer/progress readers, preserving null
versus empty arrays. Keepalive begins during construction before `Start`.
Graceful completion joins each thread, prints statistics, exits the worker,
then removes that registration quietly; lost-worker removal emits the
idempotent deregistration diagnostic. Sent/received counters wrap as Java ints,
and result/statistics/timestamp/delta processing remains in the inner remote/NPE
catch. Null results/first partitions take the lost-worker path; an empty
partition array takes the outer model-error path.

`DistributedWorkerRuntime` represents the static executor, keepalive timer,
ordered worker group, and completion latch for one worker process. A new Go
worker has its own runtime; `DistributedWorkerGroup` instead assembles one
shared runtime and publishes each runnable's worker before registration. The
group starts all registration goroutines, then schedules the timer and prints
readiness without waiting for registration. Standalone local registration still
starts its convenience timer after starting the server thread. Exit prints
completion, shuts down the executor,
cancels the shared timer, marks that worker endpoint removed, then decrements the
latch. It does not acquire the computation lock or wait for accepted tasks.
Repeated direct exits print again and fail at removal without decrementing;
endpoint calls to a removed worker fail with a native DistributedOperationError.
The EndpointRemoved trait survives native failure payloads and is distinct from
the broader WorkerUnavailable category, which also includes connection closure.
Direct `isAlive` remains true. Shutdown resolves each runnable’s worker only
when that index is reached, observing startup publication during earlier exits.
It releases the lifecycle lock before calling Exit. A nil runnable fails after
prior exits, retaining array/latch mutations and releasing the lock; a nonnull
runnable with an unpublished worker is skipped. Shutdown tolerates only prior
endpoint removal; keepalive logs that same condition and continues to later workers.
Shutdown does not recreate the executor or latch. AwaitTermination waits for the
latch, then sleeps ten seconds before returning.

TLCWorkerAndFPSet.main now has native and root RunDistributedWorkerAndFPServer
entry points. It starts named FP-server and worker threads in source order,
passing the same argument array, without a readiness barrier or joins. Start
failures escape the launcher and leave any earlier started thread intact;
command-body failures reach an uncaught-handler boundary on their own threads.
A missing native worker process is created lazily there, and the root wrapper
uses the production parser. Go goroutines supply asynchronous execution;
standalone non-daemon process lifetime, OS hooks and full JVM default-handler
behavior require process providers. Ten source launcher comparisons and a
combined server/FP/worker/local-naming integration pass under -race. That
integration defers worker bootstrap until after initialization; arbitrary
simultaneous bootstrap/evaluation still lacks FP64/UniqueString runtime isolation.
No dedicated upstream launcher tests exist.

TLCServer.main runs through DistributedServerProcess and the root
RunDistributedServer application bridge. Packaged model properties load before
application creation through the independent LoadProperties operation. The
server retains constructor selection, management and worker-hook registration
before checking, and source GC/error/close handling. Its finally block retains
the nil-server dereference, forced shutdown and management unregistration.
Exception precedence follows Java, including Error escaping the close Exception
catch. Email construction, capture, attachments and delivery have been removed
by user instruction. Normal console streams remain installed.
The executor exposes a native interruption signal for shutdownNow task/RPC
adapters, with ordinary shutdown leaving accepted work uninterrupted. OS hooks
are retained callbacks unless the installer registers them. Process hooks,
JMX/JDK interruption providers, wire RPC and config-read process exit remain
pending. The full upstream transport harness remains explicitly skipped; no
dedicated main tests exist.

TLCWorker.main now runs through a concrete DistributedWorkerProcess and the
root RunDistributedWorker parser bridge. It preserves the argument/error/flush
contract and allocates the volatile completion latch before entering the lookup
catch. Resolver, executor and interned strings survive repeated main calls;
main replaces the current latch and, after starting registration threads,
replaces the timer without canceling its predecessor. Old timer tasks retain
their supplied runnable array, while worker exit uses the current timer/latch.
Partial runnable publication and thread failures preserve their source order.
Shutdown only clears resolver/runnable state after direct exits succeed, leaving
executor/latch lifetime intact. AwaitTermination captures one latch and exposes
interruptible wait/sleep boundaries with the ten-second disposal delay.
Application config syntax failures preserve ConfigFileException. Actual-source
command comparisons and production parser/local naming/two-worker integration
pass under the race detector. Wire invocation, MP/SANY console routing,
config-read process exit, standalone process wiring and simultaneous
FP64/interner class-global isolation remain separate pending work. No dedicated
upstream main/shutdown/await tests exist; transport-dependent tests remain with
the full transport feature.

Worker keepalive captures the registry URL from startup and performs a fresh
lookup followed by isDone after inactivity. Local groups supply a direct-server
adapter; discovery groups retain the original Naming.lookup boundary. Timeout
freezes at first task initialization, with Java int multiplication overflow.
Computing workers or recent/future invocation timestamps suppress lookup; zero
invocation forces it. The task retains its original runnable array and reads
workers dynamically, so later worker publication remains visible and a missing
worker causes the source NPE. RunKeepAliveOnce represents Java's public run.

RemoteException and NotBoundException print the source nullable error detail,
then exit workers in order. A finished server prints its completion diagnostic
before the same exit loop. Malformed URLs and NoSuchObjectException from worker
exit reach an explicit FINEST logging boundary. Unexpected failures escape;
task cancellation follows the exit loop, while each worker exit still cancels
the shared timer earlier. Timer scheduling preserves the ten-second initial
delay and sixty-second period measured from the preceding actual start;
uncaught failures terminate its goroutine. JVM wall-clock changes and complete
uncaught-thread/logging-provider integration remain separate. Twenty-seven
actual-source task/worker comparisons and native race probes match the activity,
lookup/query, property, error, cancellation and export observations. No dedicated
upstream timer tests exist; complete distributed tests await transport.

Server keepalive and final cache reads preserve their RemoteException catches;
final exit ignores only the three Java dead-worker exception families, warns,
and removes the registration in finally order. Actual registry/export/RPC
transport remains pending. Worker
fingerprint lookups and server block inserts now use
their respective shared executors; accepted tasks finish after shutdown,
while new submissions are rejected. The worker preserves the rejection's
RemoteException and the proxy's ServerException envelope.

`RMIFilenameToStreamResolver` now fetches through the `GetFile` contract into a
private temporary directory, caches paths by basename, and fetches again only
when the cached file disappears. Java ignores constructor library paths and
the `isModule` flag here; remote failures print a stack trace and still create
an empty file. Null returned bytes throw after creating the file and before
caching it. `GetFullPath` retains the source's comparison against each key's
UTF-16 length, with deterministic insertion order in place of HashMap order.
File deletion follows Java's process-exit lifetime, rather than worker exit.
The command wrapper calls `CleanupDistributedFiles`; library hosts call that
hook on process shutdown themselves.

`TLCServer.GetFile` strips the request to its basename and performs the
filename-only InJar/SimpleFilenameToStream search: packaged `/model/` assets,
user/spec directory, explicit library directories, then embedded standard-module
assets supplied by the parser bridge. The superclass's one-argument resolve
invokes the virtual two-argument method, retaining InJar's packaged-model
precedence. Failed InJar copies fall back to Simple; failed Simple copies
return the attempted file. Server reads keep
the directory/Integer.MAX_VALUE checks, single-read zero-filled buffer, and
nested RuntimeException wrapping with FileNotFoundException/IOException causes.
`LoadDistributedWorkerTool` parses configuration first and routes all module
loads through the worker resolver using the existing parser, semantic checker,
and TLC bridge. It now installs a fresh worker interning context before any
configuration or semantic values are created, with FP64 initialization from
the server polynomial before installing that source. Fresh-context bootstrap
refreshes builtin, counterexample, TLCGetSet and TLCExt interned class-static
names; keeping names from the discarded table breaks register-key identity.
`StartDistributedWorkerGroup` then creates the shared local group and starts
asynchronous registration. The server exposes the TLCApp command-line deadlock
flag (default true), constant preprocess flag and current FP64 polynomial;
ModelConfig CHECK_DEADLOCK does not supply the distributed application flag.
Java Integer.getInteger decoding now governs distributed integer properties,
including signs, hex/octal prefixes, signed-int bounds, BMP digits and invalid
value fallback. FP64 random initialization uses Java Random, indexed
initialization preserves typed bounds failures, and the unrolled integer
extension has a separate source-loop implementation. The missing hex digit in
polynomial 65 is corrected; all 131 constants and 655 source-Java integer/long
extension pairs match. The upstream FP64Test is ported after implementation.
Network invocation, registry discovery/retry, endpoint export and full command
startup/error/shutdown wiring remain pending.
`FilenameToStream` now carries the local/remote resolver contract. Both concrete
resolvers return `TLAFile`, retaining existence-at-construction URI provenance,
library flags and the resolver used for `.class` override lookup. Remote cached
files keep their object identity and null original URI when constructed before
writing, even after the file starts to exist. The production parser and config
loader use `GetPath` at their OS file-reading boundaries.

`SimpleFilenameToStream` snapshots the user/CWD and explicit or TLA-Library
paths at construction, then searches standard-module and bare classpath roots.
Explicit filesystem libraries remain non-standard; both classpath locations
are standard. Module loading normalizes the .tla suffix and warns/truncates at
newlines; non-module loading keeps the filename. Missing results retain the
historical parent/child TLAFile behavior instead of returning null. File paths
retain dot components and provenance URIs preserve Unicode and escape spaces.

Go supplies ordered classpath directory/archive entries or mounted bundled
fs.FS resources. The parser bridge supplies its frozen standard assets; default
native classpath entries come from java.class.path/CLASSPATH or the CWD. Archive
resources keep their jar:file: URI while copying to one per-resolver temporary
directory. Filesystem resources retain their path. The descriptive search path
keeps Java's nested separators. Generic JVM class-loader providers, manifest
classpath expansion and module-loader VM hooks are not implemented by this
native resource adapter.

InJar behavior is folded into the same concrete resolver: raw packaged names
are tried before extension/newline handling, copied modules remain non-library,
and copy failure falls back to Simple. Simple copy errors print and return the
attempted file; opening a discovered resource and outer close errors preserve
the RuntimeException/IOException boundary. Copying uses Java's 1024-byte loop
and registers files for process-exit deletion. Native stream closes are
idempotent; resources are also closed on failure rather than leaking them.
`DistributedServerFiles` now delegates to this resolver instead of duplicating
its search/copy algorithm. Its temporary directory is retained across requests.

All eight upstream SimpleFilenameToStreamTest cases are ported after the
implementation, using existing frozen standard/community fixtures under
`test_vectors/`. The Windows-specific case retains its source platform guard;
verification here ran on Linux. Direct probes compiled the actual Java
resolver/locator/TLAFile sources and matched user/library/missing/standard/
community provenance, original Unicode/space URI spelling, path descriptions,
uppercase suffix handling, override lookup and archive predicates. The real
worker tool-loading/cache/failure/bootstrap probe also passes under -race.

`InternTable` is a concrete linear-probing table with Java String hash codes,
half-capacity growth thresholds, `2 * length + 1` growth, and signed int token
increments. `Put` grows even before a cache hit. With an `InternSource`, absent
strings retain the returned token/location and leave the local token counter
unchanged; exceptions become the cause-free `Failed to intern ...` runtime
assertion, while Java Error carriers escape. A null response increments count
without occupying a slot, preserving the source quirk.

Interning is serialized per logical JVM/table. Individual array accesses have
a separate Go lock, so `Get`, `Find`, `ToMap`, and checkpoint scans retain Java's
live slot iteration and do not wait across an in-flight remote create. `ToMap`
uses InsMap to make the resulting map iteration deterministic. Checkpoints
write the token counter followed by occupied slots in their current table order,
using Java's legacy UniqueString byte encoding. Recovery updates the counter
and inserts into the existing table, including duplicate strings/tokens and
growth, rather than replacing or deduplicating it. Incomplete records preserve
the nullable checkpoint-corruption assertion; a truncated header remains EOF.

The local worker interning source captures the master's table and copies raw
UniqueString fields to model Java serialization between JVMs. The worker then
recreates Go's cached built-in and counterexample/action/location names after
source installation. This avoids using package-initialization tokens or
sharing mutable locations with the master. Construct one worker Tool per fresh
interning context and share it among that JVM's worker threads; context reset
belongs to bootstrap before concurrent checking.

Server checkpoint ordering now includes the interner: queue/trace/FP begin,
queue resume, interner begin, queue/trace/interner commit, then FP commit. Errors
before queue resume leave the queue suspended like Java; completion messages
retain the source's recorder arguments. TLCApp restores interning before tool
construction, separately from TLCServer's trace/queue/FP recovery method.

Worker construction retains immutable raw URI metadata in the native form
`tcp://hostname:port/threadId`. Native publication replaces the default thread
path with its unique published object name and supplies its actual listener
address. Local construction uses the machine hostname and port zero until a
transport supplies an endpoint. The raw URI is used for registration,
statistics and getURI; worker failures use toASCIIString. Host classification
follows Java's server/registry authority distinction rather than Go net/url,
including null hosts for Unicode or underscore registry names, bracketed IPv6
hosts, scope IDs, and Java IPv4/hostname rules. URI validation preserves the
component masks, escaped-pair checks, UTF-16 failure indices and the
IllegalArgumentException -> URISyntaxException message/cause boundary.

ASCII conversion normalizes to NFC, encodes non-ASCII UTF-8 bytes with uppercase
percent escapes, and preserves existing ASCII spelling/escapes. Unicode tables
come from the cached golang.org/x/text v0.28.0 dependency. Since its normalizer
inserts stream-safe CGJs into long combining sequences and Java does not, a
canonical decomposition/order/composition fallback removes that behavioral
difference while retaining genuine input CGJs. Java TLC has no dedicated URI
test; 232 fixtures captured from the installed OpenJDK 21.0.12.1 verify worker
URI construction, host/null classification, syntax failures and ASCII output.
An integration check covers the executor-rejection message and proxy envelope.
Worker-group construction now performs the local canonical-host lookup:
resolve the machine name, prefer IPv4 by default (or the configured IPv6/system
policy), then verify the reverse name against a forward lookup. Failed
canonicalization returns the numeric address, with Java's uncompressed IPv6
spelling. Successful local-host addresses have a five-second monotonic cache;
each cached address retains its canonical name. Initial lookup failure keeps
the nested UnknownHostException cause and the runnable's RuntimeException
wrapper. Go supplies native DNS/hosts resolution; JVM resolver providers,
jdk.net.hosts.file, general DNS TTL caches, scoped-address metadata and VM
security hooks are not represented. Endpoint allocation and reflection/VM
diagnostics remain with the pending startup/export transport feature.

FP-manager registrations retain Java FPSets wrapper identity, cached hostname,
and availability. Reassignment shares wrappers across partitions while keeping
the original partition count. It marks the selected wrapper unavailable,
chooses the next available successor, and retains Java's non-wrapping
`for (j=index; j<next; j++)` replacement loop even when the successor index
wrapped to the start. Exhaustion caches managerIsBroken. Scalar and block calls
catch Exception, reassign and retry; Java Errors escape. Exhausted block calls
use Java's BitVector(size,true), including its closed-range initialization
quirk. Statistics failures reassign without retrying that slot, and size/seen
totals count every partition, including shared wrappers. Close and checkpoint
coalesce adjacent wrappers and trailing copies of the first wrapper, rather
than globally deduplicating underlying FPSet objects.
Native manager payload nodes carry an explicit endpoint-null bit. Empty slots
still use registration ID zero; nonnull wrappers with nil endpoints retain their
own IDs, hostname, availability and sharing. Encoding skips reference publication
for nil endpoints; decoding neither resolves them nor acquires connections.
A null endpoint containing a nonempty reference is rejected before resolution.
Receiving the manager does not invoke failover; operations retain manager policy.
Peers need the current payload build.

Traversal fixes the initial registration count but reads slots live. Checkpoint
runs its first operation before trimming trailing first-wrapper copies; each
begin, commit and caught-I/O hostname lookup resolves the slot separately.
Lifecycle calls require the endpoint at each access, preserving explicit null
failure diagnostics. Checkpoint’s unchecked null failure stops later phases and
registrations without I/O reporting/reassignment. Distributed close catches that
ordinary failure, prints its stack and continues; the local manager propagates
it before later lifecycle calls. Fatal failures keep their existing boundary.
Shutdown trims before exit and captures the next wrapper before exiting the
current one. Neither path holds the manager lock during endpoint calls.

Concurrent block calls submit one callable per partition and collect results
by their saved index, regardless of completion order. Submission retries retain
the three-retry bound, Java random one-to-five-second delay and shutdown check.
Null fingerprint-check endpoints fail inside the task with the source null
failure category; missing registration wrappers fail during submission. Failed
check completions do not reassign registrations or enter callable I/O catches.
ExecutionException is logged and leaves a null result slot; checkFPs and
checkInvariant also use concurrent completion collection, with Java's signed
minimum and early false return. NonDistributedFPSetManager bypasses executors
and preserves its IOException fallback contracts. Empty distributed managers
now report zero servers and indexing throws ArithmeticException.

All nineteen DynamicFPSetManagerTest cases and five FPSetManagerTest cases have
been ported after the feature implementation, including FaultyFPSet's scalar
and block virtual dispatch. The nested factory tests retain the production
factory and two high-bit subpartitions with a bounded Go memory budget.

Worker failure contracts preserve TLC decisions through native Go errors:

- The worker preserves direct WorkerExceptions and translates direct
  memory exhaustion/executor rejection to DistributedOperationError.
  Other failures retain predecessor/successor metadata in WorkerException.
- Local endpoints return worker errors directly. No ServerException envelope
  is fabricated. Native operation traits carry remote/I/O, recoverability and
  exit handling across TCP, with the original cause and worker URI retained.
- Worker memory exhaustion permits smaller-batch retry; executor rejection does
  not. Recoverable multi-state blocks print the reduction message, requeue the
  block, and reduce the selection limit in source order. Native truncated-reply
  failures also retain their existing smaller-batch decision. Native operation
  traits replace all source remote-type classification for these decisions.
- Other remote failures and direct null-pointer failures deregister the worker;
  null-pointer diagnostics include the stack. Other exceptions reach the
  server's outer model-error catch and keep their error-state metadata.
- Concrete EOF, memory, rejection, and null-pointer classes preserve null
  versus empty messages and saved Go stack frames. Timing uses differences of
  epoch millisecond readings; empty-block network-overhead division retains
  Java's floating-point result.

Native operation errors capture Go program counters when created at worker,
coordinator and fingerprint boundaries. Payloads carry their formatted sender
stacks, and decoding does not invent receiver frames. Causes retain application
context and native errors.Is/errors.As behavior locally. A nested memory or EOF
cause does not itself grant retry: the endpoint adapter supplies that trait.

Important Java data structures:

- `NextStateResult`: carries `TLCStateVec[] nextStates`,
  `LongVec[] nextFingerprints`, computation time, and raw states-computed count.
  `getStatesComputedDelta()` returns `statesComputed - nextStates.length`.
- `IFPSetManager`: partitions fingerprint space, performs block
  contains/put/checkpoint operations, and reports states-seen/distinct counts.
- `IBlockSelector`: chooses how many states to hand to each worker. Java keeps
  this separate from server threads for performance tuning.
- `TLCWorkerSmartProxy`: hides dead/slow worker behavior from server threads.

Port guidance:

- Port the distributed TLC algorithm faithfully through Go endpoints.
- Do not implement Java RMI/JVM machinery or the rpc25519 alternative.
- Preserve existing batching, failure and retry semantics when introducing
  Go network endpoints.
- Port a feature's Java tests after implementing that feature in Go. Keep
  transport-dependent tests with the transport feature they exercise.
- The Go port should keep `TLCServer`, `DistributedWorker`,
  `DistributedFPSetManager`, and `NextStateResult` concrete. Transport can wrap
  these structs later.
- Java's `DynamicFPSetManager` is represented as dynamic fields on the same
  concrete Go manager. It rejects registrations beyond the expected server
  count and computes the low-bit mask with Java's loop
  `while expected > 0 { expected /= 2; log++ }`, yielding `(1<<log)-1`.
  Preserve that mask calculation even when it looks different from a direct
  power-of-two helper.
- Java's `NonDistributedFPSetManager` is also folded into the concrete Go
  manager. It sends every `LongVec` block to the single wrapped FP set, reports
  `getStatesSeen()` as the wrapped set's `size()`, performs checkpoint begin
  and commit in separate calls, recovers by replaying the local trace, and
  closes by calling both `Close` and `Exit(cleanup)` on the wrapped set.
- Distributed fingerprint checkpointing is filename-paired. Java
  `FPSetManager.Checkpoint.run` calls `beginChkpt(filename)` and
  `commitChkpt(filename)` on the same remote FP set; the later manager-level
  `commitChkpt()` is a no-op for the distributed manager. The Go concrete
  `DistributedFPSetManager` performs the begin/commit pair immediately within
  `Checkpoint(filename)`; its later `CommitCheckpoint()` is a no-op. Java calls
  checkpoint Thread.run directly, so that work remains sequential. Distributed
  checkpoint/recovery I/O failures print the cached hostname diagnostic and
  continue; local checkpoint/recovery returns errors to the caller.
- Any worker/server map that is iterated for progress output must use `InsMap`.
  Fingerprint holder de-duplication may use a Go map only if iteration is over a
  separately maintained sorted fingerprint slice.
- Distributed TLC does not support `TLCGet`/`TLCSet` in Java; preserve that
  eventual error behavior rather than silently sharing mutable variables across
  workers.

## Management Architecture

Java exposes TLC progress through JMX:

- `TLCStatisticsMXBean` defines status accessors and control actions.
- `ModelCheckerMXWrapper` wraps single-process model checking.
- `TLCServerMXWrapper` wraps distributed TLC.
- `TLCStandardMBean` provides version/revision metadata and registration.
- `StateMonitor` attaches to a Java VM and periodically prints
  `getCurrentState()`.

Statistics exposed by both wrappers:

- generated states,
- distinct generated states,
- state queue size,
- generated/distinct states per minute,
- progress/search depth,
- worker count,
- average distributed block count,
- runtime ratio for liveness,
- current state string,
- spec/model names,
- checkpoint, live-check, stop, suspend, and resume controls.

Port guidance:

- Do not port JMX. Keep the wrapper names and accessors as plain Go structs so
  CLI status, HTTP, or future RPC layers can call them.
- `ModelCheckerMXWrapper` should delegate directly to `ModelChecker` and
  `StateQueue`; `TLCServerMXWrapper` should delegate to the distributed server
  shell.
- `checkpoint()` maps to `ForceCheckpoint()`.
- `liveCheck()` maps to a concrete flag on `LiveCheck` that is consumed by the
  next `LiveCheck.Check` call.
- `suspend()` and `resume()` map to the state queue's existing synchronization
  methods.

## Checkpointing and Recovery

Checkpointing spans:

- state queue.
- fingerprint set.
- trace fragments.
- liveness disk graphs.
- checker metadata.

Pattern:

1. Main checker suspends workers.
2. Components write `.tmp` checkpoint files.
3. Components atomically commit to `.chkpt`.
4. Workers resume.
5. Recovery rebuilds in-memory structures from checkpoint or trace.

Correctness notes:

- Queue suspension races have historically had subtle bugs. Preserve the Java
  condition-variable protocol.
- Checkpoint commits must be atomic enough for crash recovery assumptions.
- Disk file formats matter if tests inspect sizes or recovery behavior.
- `vars.chkpt` follows Java `InternTable`: a single `tokenCnt` integer,
  followed by `UniqueString` records until EOF. Each record is token, variable
  location, Java string length, and the low byte of each UTF-16 code unit. Do
  not add record counts or value-stream handle encoding.

## Test Architecture

### Harnesses

`CommonTestCase`:

- installs a `TestMPRecorder`.
- provides trace assertions.
- provides liveness lasso assertions.
- provides coverage assertions.
- checks generated trace-expression specs.
- validates disk graph file sizes in some tests.

`ModelCheckerTestCase`:

- runs `tlc2.TLC` with a spec/config path and extra arguments.
- records actual exit status.
- provides setup/teardown isolation.
- supports liveness and normal model checker tests.

`SuiteTestCase`:

- common expectations for the historical `suite` models.
- asserts finished, stats, init count, no general error, and coverage.

`TTraceModelCheckerTestCase`:

- variants that check trace expressions or Toolbox trace behavior.

`TLCDebuggerTestCase`:

- runs TLC under debugger configuration and checks DAP-like events.

### Test Fixture Strategy in Go

The Java tests rely heavily on `test-model`. For Go:

- Freeze required `.tla` and `.cfg` fixtures under `tlago/tlc/test_vectors/`
  or another self-contained path before enabling translated tests.
- Do not use a directory named `testdata/` for persistent fixtures. The user
  reserves that name for ephemeral Go fuzzer storage and cleanup. Persist
  vectors in `test_vectors/` instead, including generated Java oracle data.
- Keep path conventions compatible with module names.
- Prefer generated golden recorder-event JSON only if direct translated tests
  would be too expensive; otherwise port assertions directly.

### Test-Port Ordering

Per user instruction, do not port tests until the function-level mechanical
implementation exists. When test porting starts, port in this order:

1. Utility tests: FP64, byte utils, vectors, queues/stacks, statistics.
2. Value tests: primitive values, tuple/record/set/function/subset semantics.
3. Config parser tests and model-config behavior.
4. Fingerprint set tests.
5. State queue tests.
6. Small `suite` model checker tests.
7. General model checker tests in `tlc2/tool`.
8. Simulation tests.
9. Liveness tests.
10. Coverage tests.
11. Checkpoint/recovery tests.
12. Standard module tests.
13. Debugger tests.
14. Distributed TLC tests.

### Behavioral Axes Covered by Java Tests

The Go test suite must eventually cover:

- exact FP64 output and value fingerprinting.
- value equality, comparison, normalization, and enumeration order.
- model value creation, ordering, and symmetry permutation.
- config parsing and constant/operator override handling.
- init/next action decomposition.
- bounded quantifier and CHOOSE behavior.
- incomplete state detection.
- invariant and implied-action violation reporting.
- assumption and postcondition checking.
- deadlock detection.
- model/action constraints.
- VIEW and symmetry reduction.
- alias rendering in traces.
- random element/subset behavior under seeds.
- simulation trace counts and depths.
- liveness tableau construction and accepting-cycle detection.
- liveness safety-property short-circuiting.
- fairness and stuttering behavior.
- checkpoint and recovery.
- coverage output.
- standard module operators.
- trace-expression generation.
- debugger source/stack/scoped identifier behavior.

## Correctness Hotspots

Pay extra care to these areas:

- FP64 exactness, including signed byte and unsigned shift behavior.
- `Value.compareTo` ordering across all value kinds.
- Set/record/function normalization and duplicate elimination.
- Lazy enumerable values and their enumeration order.
- Model value initialization and global ordering.
- Symmetry permutations and representative selection.
- VIEW fingerprinting.
- Context branch/cutoff behavior for `ENABLED`.
- Prime and unprime handling in action evaluation.
- Square actions return TRUE immediately when A is true; angle actions return
  FALSE immediately when A is false. Their subscript is evaluated only on the
  remaining path, preserving Java exception precedence and coverage.
- `UNCHANGED` and tuple-of-vars handling.
- Action splitting and action names.
- Exact distinction among constraints, invariants, implied actions, and
  assumptions.
- Error precedence when multiple workers find errors.
- Trace write timing before invariant/implied-action checks.
- Deadlock definition: no successors generated, not merely no unseen successors.
- Liveness formula level handling and recursive temporal operator expansion.
- Tableau product graph fingerprint composition.
- Stuttering steps in liveness.
- Checkpoint queue suspension races.
- Output message codes and parameter order.

## Performance Hotspots

Java TLC performance depends on:

- mutable array-backed states.
- interned `UniqueString` variable locations.
- lazy values and streaming enumerators.
- action decomposition.
- FP64 speed.
- concurrent workers with shared FPSet and state queue.
- disk-backed queues/FPSets for large models.
- per-worker trace fragments.
- avoiding full `StateVec` allocation for huge initial states.
- avoiding repeated config record construction for `TLCGet("config")`.
- liveness action checks outside coarse graph locks.

Go performance guidance:

- Use slices for state values and vectors.
- Use integer variable locations, not map lookups, in hot state binding.
- Keep `Context` as a linked list initially; optimize only with tests.
- Use goroutines with explicit `sync.Mutex`/`sync.Cond` for queue parity.
- Use `uint64` internally for FP64 but match Java signed rendering where needed.
- Avoid `interface{}` churn and interface dispatch in inner loops where
  possible, but do not abstract before behavior is stable.
- Implement disk-backed structures before declaring performance parity.

## Initial Go API Shape

A minimal public API should support:

```go
type Options struct {
    SpecFile   string
    ConfigFile string
    MetaDir    string
    Mode       RunMode
    Workers    int
    Deadlock   bool
    Seed       int64
    Aril       int64
    TraceDepth int
    TraceNum   int64
}

type Result struct {
    ExitStatus      int
    ErrorCode       int
    StatesGenerated uint64
    DistinctStates  uint64
    QueueSize       uint64
    Messages        []Message
}

func ModelCheck(ctx context.Context, opts Options) (*Result, error)
func Simulate(ctx context.Context, opts Options) (*Result, error)
```

Command-line tools can wrap this API, but the library must be directly usable
from Go tests and downstream translators.

## Mechanical Port Checklist

Port implementation in this dependency order:

1. Output message constants and recorder API.
2. Utility collections needed by values: vectors, bit vectors, long vectors,
   object tables, random generator.
3. FP64 and fingerprint helper tests against Java constants.
4. Value interfaces and primitive values.
5. Composite and lazy values.
6. Value serialization streams.
7. `Context`, action item lists, and context enumerators.
8. `ModelConfig` parser.
9. `Spec` and `SpecProcessor` over the existing Go SANY semantic tree.
10. `Action`, `Tool`, evaluator, enabledness.
11. `TLCState`, state vectors, state printers.
12. Standard module built-ins and override registry.
13. `MemFPSet`, then `MultiFPSet` and disk FP sets if/when multiple concrete
    implementations are needed.
14. `MemStateQueue`, then disk queues if/when multiple concrete
    implementations are needed.
15. `ConcurrentTLCTrace`.
16. `AbstractChecker`, `ModelChecker`, `Worker`.
17. `Simulator` and simulation workers.
18. Liveness expression nodes and formula processing.
19. `LiveCheck1`.
20. Disk-backed liveness graph and `LiveCheck`.
21. Coverage.
22. Checkpoint/recovery.
23. Debugger support.
24. Distributed support.

After each feature is implemented, translate its existing Java tests into Go,
following the user's updated feature-by-feature sequence.

## Compatibility Definition

The Go TLC port is ready when:

- It parses specs using the production Go SANY parser.
- For TLC-supported TLA+, it produces the same reachable-state counts as Java
  TLC for the frozen test corpus.
- It reports the same primary error code for every Java test-model case.
- It prints equivalent counterexample traces, including state order, action
  labels, stuttering/back-to-state markers, alias output, and lasso shape.
- It matches Java value semantics, including comparison, normalization,
  membership, enumeration order, fingerprinting, and serialization.
- It matches Java liveness decisions for safety-like and true liveness
  properties.
- It can checkpoint and recover models that Java tests cover.
- It performs within the same rough complexity class as Java TLC on large
  models, with disk-backed queues and fingerprint sets available.

## Native user-module override registration (2026-10-04)

`tlc/native_class.go` supplies the native class-linkage boundary for Go ports of
Java override classes. An application registers `NativeClass` descriptors with
`RegisterNativeClass`: declared method names, source reflection signatures,
visibility/static/final flags, parameter counts, annotation metadata, and actual
Go callbacks. The returned function restores the previous class binding. These
callbacks are compiled Go implementations; Java class bytes remain provenance
and resolver/classpath resources. Loading an arbitrary Java binary does not
provide its executable Go implementation. `RequireResource` retains the source
class-file discovery requirement for a linked adapter; ordinary linked Go
classes can supply their own resource URI. Test fixture classes are linked in
root test code, never in production and never by replacing TLA definitions.

`NativeClassLoader` follows TLAClass's resolver/unqualified-classpath/package
fallback order and retains directory/archive resource URIs. Its resolver
attempt catches Exception; outer class loading catches Throwable and produces
TLC_ERROR_REPLACING_MODULES. `LoadSanySpec` now retains its FilenameResolver in
`Spec`, so the bridge and native loader share SANY's actual resolver/classpath.
The previously ported standard native bindings remain installed by the bridge;
`SpecProcessor.ProcessModuleOverrides` adds external conventional module classes
and ordered user ITLCOverrides indexes using the same source registration
boundaries. The built-in index name precedes the system-property-selected
indexes; the source property separator/default are preserved.

Conventional registration enumerates every declared public static method,
uses TLARegistry name mapping, skips Operator/Evaluation annotations for the
later index pass, and checks non-built-in methods against root operators
originally defined in that module. Missing names and arity mismatches emit the
source warning; accepted methods emit loaded metadata and replace actual body
ToolObjects/Defns. Public static final zero-argument methods are evaluated at
registration, following MethodValue.get, even before mismatch checks.
Annotation registration checks module/identifier first, handles Evaluation,
Callable, then Operator in source order, preserves priority composition,
minLevel/warn/silent settings, method arity and constructor failure catches.
Index instantiation catches only the thrown InstantiationException or
IllegalAccessException type; a cause of that type does not widen the catch.
Production MP templates now include exact source loaded and all three mismatch
messages. Source path-dependent parts remain actual resource URIs.

Original UserModuleOverride, jar, annotation and EvaluatingValue model tests
exercise native callbacks through production SANY/spec processing/model checking.
The conventional Get returns TRUE despite its FALSE TLA body; the annotated
Get/Get2 return TRUE; the Evaluation callback synchronously binds successor x
to42 despite its FALSE assignment in TLA. This distinguishes native execution
from running the unmodified TLA fallback. Original fixture files/jar/class
resources are retained under test_vectors, with callbacks mechanically ported
from Java source and jar bytecode. No evaluator branch depends on test names.

### Polymorphic state sets

`TLCState` defines the virtual fingerprint/tool-fingerprint, equality, hash,
action and string surface used by `SetOfStates`. The evaluator still uses
`*TLCStateMut`; liveness conversions and `ToSlice` are concrete evaluator
boundaries, while `Next` and `ToSet` accept genuine alternate state types.
`StateVec` remains concrete. Each collection entry supplies its own behavior;
no fingerprint/equality hooks or synthetic model variables replace state types.

`SetOfStates` retains Java's open addressing, source tool/no-tool growth,
iterator position across growth/clear, reset semantics, signed fingerprint
printing and unconditional UTF16 terminal-code-unit removal. Only the source
TLCRuntimeException equality mismatch is suppressed at the collision boundary,
with its message assertion preserved. Mutable-state equality propagates value
failures elsewhere. `TLCStateSet` preserves HashSet hash/equals behavior, including
TLCStateMut's inherited object-identity hash despite structural equality. Action
subsets filter by actual action identity and use these same set semantics.

### POSSIBLE counters and executing workers

`_Possible._Update`/`_Track` uses `TLCGetOrDefault` and `TLCSet` named registers
on the executing worker, with the source empty-tuple default. The Go native
`PossibleTrackNode` now delegates to those same production adapters. A state's
WorkerID denotes its creator and cannot select the executing worker's register:
successors may be generated by another worker. Using the creator allowed logical
lost updates across workers even when each individual read/write was locked.
Main-thread register initialization uses the source broadcast behavior, while
simulation retains local-worker registers. `_Counts` aggregates per-worker
values after model checking. The original four-worker DiameterTest's actual
model postcondition checks these counts; its unchanged assertions exercise the
production correction without synthetic state IDs or test-only callbacks.

The checker and runner accept the `IStateWriter` contract rather than the concrete
`StateWriter`. This mirrors Java's virtual state-writer dispatch: a custom writer
can delegate to the default DOT writer and then observe constrained transition
flags. Default dump constructors and the concrete liveness DOT implementation
retain their existing file/format behavior. Liveness setup accepts the interface
and derives its output filename through `GetDumpFileName`. The original
DotConstrainedTest uses an embedded DOT writer with an overridden transition
method; it receives the actual checker flags, without instrumentation hooks or
inferences from generated DOT text.

Runtime invariant templates now use the existing shared expression compiler with
root-module scope and the null source location, as RuntimeInvariantTemplate does
in Java. This performs semantic/dependency/level checking and retains the compiled
OpDefNode on an internal Action. Invariant names come from that Action. The old
standalone syntax-only expression module and direct AST conversion were removed.
SpecProcessor's failure boundary raises TLC_PARSING_FAILED2 for compiler errors;
it never silently omits a malformed invariant. TLC.process records the coded
exception, emits FINISHED, and selects ERROR_SPEC_PARSE. This is reuse of the
existing compiler by the invariant feature, without new debugger test work.

Initialization callbacks catch fingerprinting and other panics at Java's
DoInitFunctor throwable boundary, retaining the current state and cause. Runtime
and invariant failures abort; other throwables are stored until init enumeration
finishes; actual OutOfMemoryError selects TOO_MANY_INIT. The enclosing init and
next-state replay calls convert panics to retained errors for diagnostic replay.
Worker.addElement wraps Exception with the successor state but lets Java Error
subclasses escape. Worker.run's inner catch around next-state generation is
separate from its outer loop catch; both preserve the original throwable.

LazyValue.getValue and LazySupplierValue.getValue do not attach source metadata
in Java and now preserve that distinction in Go. LazyValue.eval retains its
source-if-missing behavior. Adding a source in the getValue path could turn an
ordinary runtime failure into a FingerprintException, whose nullable detail
message suppresses a GENERAL diagnostic. The original fingerprint-exception
models cover both paths, their exact trace strings, coverage, and source exits.

VIEW presentation delegates to SemanticValueString, matching
SemanticNode.toString(IValue) and OpApplNode's tuple override. SemanticAllParams
uses SANY's union, formal/bound-symbol removal and substitution equations over
semantic identities, iterating recursive definitions to a fixed point. Tuple
values with matching parameter count print names in var-location order; other
values use ordinary Values.ppr. VIEW fingerprinting remains separate from this
presentation method. The original ViewMap test checks all eight states and action
labels, its model postcondition and coverage.

SANY file loading now resolves ordinary dependencies before falling back to the
root monolith, matching FileUtil.createNamedInputStream. Embedded modules are
extracted between source delimiters to real temporary files and parsed on demand;
unused sibling modules are not eagerly registered or parsed. Parsing progress
reports canonical file paths plus physical monolith/resource provenance. The
production TLC loaders route progress and missing-file diagnostics through actual
ToolIO streams. LoadOptions exposes callbacks and retains silent library defaults.
The original MonolithSpec test captures those streams and checks three provenance
patterns; only the source fixture-directory suffix is adapted to test_vectors.

Extracted monolith siblings retain logical module-relative syntax positions,
while SourcePath and parsing-progress provenance retain the actual temporary
file. Parameter substitution follows the two distinct SANY equations:
SubstInNode replaces sequentially; APSubstInNode uses the first matching
Subst.allParamSet replacement for each original body parameter.

The shared TLC tool loader applies MP's SANY suppression/elevation settings to
front-end diagnostics before testing success, and writes controlled diagnostics
to ToolIO.out. Ordinary parse/semantic errors retain diagnostic parameters in
TLC_PARSING_FAILED; elevated warnings also abort with that code, as the original
SpecProcessor.processSpec does when SANY exits unsuccessfully. The original
ElevatedSanyWarning model exercises actual warning promotion, stdout and the
inherited parsing-failure exit rather than a test-only diagnostic path.

Named module-instance parameters belong to a local formal context when checking
WITH expressions, matching Generator.generateModuleDefinition. They are distinct
from the instancee's substitution targets; matching names can supply implicit
substitutions. The central Instance AST retains formal operator arities, which
flow into semantic checking and actual FormalParam symbols during conversion.
Source definition symbols retain their declared names and identities; qualified
bridge keys remain lookup aliases. Exported instantiated definitions keep their
qualified names. This matters when property decomposition uses an operator's
name for the initial-state diagnostic, as Github1134a checks exactly.

Alias registration binds the exact semantic symbol and the explicit alias only;
a source operator's declared name does not automatically become a root export.
Module-specific config overrides retrieve actual original definitions from the
named ModuleTbl node, including definitions with native body overrides. Using
qualified bridge lookup strings as semantic symbol names would miss the shared
body update and can turn a bounded model into an unbounded search (Github314).

Global config CONSTANT assignments to root operator definitions attach the value
to the OpDefNode's ToolObject, matching Java SpecProcessor's rootOpDefs loop.
Defns entries and module lookup aliases may still reference the original node;
its configured replacement remains authoritative when evaluating shared symbols.
This preserves the distinction from module-specific constants/overrides, which
Java attaches to original definition bodies.

Instance export signatures are prepared before converting bodies: instantiated
operator arity is the sum of instance formal count and source operator formal
count, matching Generator.generateModuleDefinition's concatenated parameters.
Operator arguments are therefore emitted as OpArgNode even when the source
operator's body is converted later. This is required by parameterized liveness
properties passing an instance operator through nested wrappers.
DOT fingerprints retain uint64 identity internally but use signed int64 decimal
in output, matching Java Long.toString in initial nodes, transition endpoints,
successor nodes and rank entries. The Github1147 golden graph checks every line.

Fairness-property origin tagging compares actual source syntax text, as Java's
SyntaxTreeNode.toString does, rather than semantic-node location text. The
specification decomposition stack includes the current OpAppl when reaching the
level-based handler, and non-next boxed formulas use that handler's origin
tagging. Conjuncts retain separate copied stacks, and lazy formal references are
removed before decomposing their original argument. This preserves the exact
positive/negative tautology-warning distinctions in the Github1198 suite.

When recursive liveness expansion returns a constant/state/action-level body,
astToLiveAppl uses that expanded level immediately, as Java does, instead of
recomputing the operator's static temporal bound. This preserves state-level
base cases inside temporal recursive wrappers.
Liveness translation also preserves recoverable execution-stack exhaustion.
The public entry holds a non-escaping stack anchor; all recursive private calls
share its tracked pointer, which Go relocates when growing the goroutine stack.
The native stack-byte span is bounded by the reference JVM's observed default
1 MiB thread-stack budget. No operator-count or semantic recursion limit is used.
Exhaustion raises the existing StackOverflowError type as a panic, bypassing
Exception-only fallbacks and reaching TLC.process's specific resource-error
catch. Normal and race compiler escape analysis confirm the markers stay on the
stack; finite expansion and the two original overflow tests all pass.


Native MethodValue diagnostics retain Java reflection's full method signature,
including Print and PrintT's declared return and parameter types. Call-stack
replay uses the ordinary evaluator's hooks: Java CallStackTool constructs a Tool
from shared Spec fields, and cannot inherit DebugTool's virtual evaluation
behavior. Starting the Go copy from NoDebug preserves the shared semantic graph
while routing every nested predicate through the replay call-stack recorder.
The original Github179a/b/c tests preserve exact native errors and full source
location stacks with all source runner defaults enabled.


Plain StateWriter writes a newline after the state's own textual representation,
as Java PrintWriter.println does; this retains the blank separator between states
in the original Github407 golden dump. Symmetry cardinality warnings derive
argument display names from the original syntax image, strip the Permutations
application text, then consult the config override-name mapping. Semantic node
stringification normally prints a location and cannot substitute for that image.
The four original Github432 rows preserve both warning grammar and exact names.


TLCEval's global cache guard must support Java ReentrantReadWriteLock's writer
and reader reentrancy. Nested constant TLCEval calls hold the outer writer while
entering the inner read/write paths. The Go internal lock preserves the default
nonfair policy (writers may barge; new readers wait behind a queued writer),
concurrent readers, per-goroutine hold counts and read-to-write upgrade blocking.
WorkerValue.demux uses the state-only eval overload with an Empty successor,
not a null marker. Passing null bypasses LazyValue's cache and changes exact
coverage counts for nested argument expressions. Both original Github648 tests
retain every coverage row and the ten-worker duplication case.
ExitStatusForErrorCode follows EC's explicit cases: only liveness tautology maps
to77; unsupported and malformed formulas fall through to the generic255 status.

2026-10-04: Github680a/b/c and all eight Github687 variants translated.
All eleven unchanged original Java JUnit tests pass; all twelve model/config
fixtures match original bytes. Original assertions and runner settings retained,
including suppressed warnings and the 687 trace-generation overrides. Five
focused Go repetitions pass (15.534s); five focused race repetitions pass
(190.095s). Production now applies variable context
cutoff during primed lookup, shares the branched ENABLED context with the
changed-subscript action item, and keeps persistent TLCStateFun bindings for
instantiated variables without root-state vector slots. Constant-false config
specification reports TLC_CONFIG_SPEC_IS_TRIVIAL as Java does. No assertions
weakened and no invented tests. Inventory: 714/1269 logical contexts, 279/626
fully mapped concrete classes; 555 contexts/347 classes pending, one partial.
Full offline workspace passes (root 409.440s, SANY 1.067s, TLC 65.403s);
full TLC race verification passes (567.957s). All check handles terminal.
Next pending issue family after this verified batch: Github696/696b. User's
five deferred topics remain deferred. Goal remains active.

2026-10-04: Github696/696b and Github715/b/c/d original methods translated.
All six unchanged Java JUnit tests pass; nine original fixtures byte-identical.
Five focused Go repetitions pass (12.085s); five focused race repetitions
including680/687 pass (295.983s). Exact diagnostics, presence/absence,
depth/stats and inherited exits retained, with original coverage=false for715.
696 passes without production changes. Source inspection and715 failures fixed
config decomposition: semantic lookup retains unbound symbol identity; a bare
variable in SPECIFICATION reports TLC_CONFIG_OP_IS_EQUAL with its declaration
location and spec tag, while a property variable reaches level-based handling.
Constant-false property and nonboolean config diagnostics match source branches.
Source operator definitions now retain static SANY levels using the existing
SANY level analysis, independently of coverage and runtime overrides; previously
all nonbuiltin definition levels defaulted to zero, suppressing715 warnings.
Inventory now720/1269 contexts,285/626 complete classes;549 contexts/341 classes
pending, one partial. Full offline workspace passes (root423.974s, SANY0.999s,
TLC64.192s). Full TLC race passes (570.326s); all check handles terminal.
Next pending family: Github725 through725h;
all eight unchanged Java JUnit tests already pass in the scratch reference runner.
No invented tests or weakened assertions. Goal active; deferred topics unchanged.

2026-10-04: All eight Github725/b/c/d/e/f/g/h original testSpec methods
translated. All eight unchanged Java JUnit references pass, and15 model/config
files match original bytes. First focused Go run passes (5.947s) without further
production changes. Preserve all original runner overrides: noGenerateSpec=true,
coverage/debugger=false, default DOT, JSON enabled for725..725f and disabled
for725g/h. Original assertions retained, including soundness cases:725g has
Prop liveness violation,2/1/0 stats,depth1,full single-state trace/ordinal/action
and stuttering2;725h uses725g with725h.cfg,exact Inv initial-state text/newline
and safety exit. Five normal repetitions pass12.897s and five focused race
repetitions pass136.002s. Full offline workspace passes (root415.375s,
SANY0.967s, TLC61.806s); all check handles terminal. TLC production unchanged since
ad50e68, whose full TLC race passed570.326s. Inventory728/1269 contexts,293/626
complete classes;541 contexts/333 classes pending,one partial. Next missing issue
is Github726Test; unchanged Java references for726/742/743/746/757 already
pass in the scratch runner. Goal active and user-deferred topics unchanged.

2026-10-04: Github726/742/743/746/757 original testSpec methods translated.
All five unchanged Java JUnit references pass; nine fixtures match source bytes.
Five focused Go repetitions pass9.308s. Exact original assertions retained:
726 captures complete writes and uses list membership,742 rejects the full6×6
fairness formula,743 retains all four trace states/actions/ordinals,746 keeps
initial-state field order/newline,757 retains CounterExample postcondition and
explicit JSON/noGenerateSpec overrides. Production fixes: LNConj DNF product
uses Java Math.multiplyExact's signed32-bit overflow boundary and exact diagnostic
before allocating; Tool.getState(successor,predecessor) catches only direct
TLCRuntimeException from candidate equality and continues to other successors,
preserving other exception types and state-generation failures. Initial742 run
failed with Go out-of-memory from the missing overflow check; after correction,
743 exposed the missing catch/truncated trace, now fully passing. No weakened
assertions or invented tests. Five focused root race repetitions (new cases
plus725 and alias liveness) pass261.674s. Full offline workspace passes
(root420.629s, SANY1.039s, TLC63.961s). Full TLC race passes555.004s; all
check handles terminal.
Inventory733/1269 contexts,298/626 complete classes;536 contexts/328 classes
pending,one partial. Next pending issues: Github766 and Github766Simulate;
both unchanged Java references already pass in the scratch runner.
Goal active; user-deferred topics remain deferred.

2026-10-04: Github766/766Simulate/798I/798N/807 original testSpec methods
translated with original settings and assertions. All five unchanged Java JUnit
references pass; four fixtures match original bytes. Five focused Go repetitions
pass11.056s and five focused root race repetitions (new cases plus aliases)
pass124.266s. Github807 exposed a production shortcut: action decomposition used
the enclosing operator when state-level arguments prevented further splitting.
Java instead uses the applied operator declaration. Tool.collectActionsAppl now
preserves that declaration, restoring the exact Add action/location in the full
three-state trace and CounterExample/TLCGet("spec") postcondition. No assertion
weakened; no invented tests. Github766 retains its explicit JSON/simulation
arguments; both798 cases retain their exact counts/depth;807 retains debugger=false,
noGenerateSpec=true and no JSON.
Full offline workspace passes (root425.504s, SANY1.070s, TLC62.970s).
Full TLC race passes551.188s; all check handles terminal.
Inventory738/1269 contexts,303/626 complete classes;531 contexts/323 classes
pending,one partial. Next pending family Github817/b/c/d/e; all five unchanged
Java references already pass. Later819/849/858/866/971a/b/c/d/e references also
pass unchanged. Goal active; user-deferred topics remain deferred.

2026-10-04: Github817/b/c/d/e,819,849,858,866 original testSpec methods
translated with original fixture bytes, constructor overrides and assertions.
All nine unchanged Java JUnit references pass. Five combined focused Go
repetitions pass20.610s; five focused root race repetitions pass161.558s.
817 retains zero-uncovered coverage and complete action-property counterexample;
819 retains liveness-tautology/FAILURE_LIVENESS_EVAL;849 retains forced trace
specification/default JSON and absence of native-method override diagnostic.
858 preserves seed1 simulation, both JSON and tlcaction dumps, exact complete
state100/ordinal/PassToken(2) action and forced trace generation. Initial858
failure was missing original CommunityModules classpath; supplying the same
archive as Java's tool manifest fixed setup, with model/assertions unchanged.
866 keeps two workers and its full postcondition. It exposed a missing platform
property default: tlcLookupSystemProperty now supplies native file.separator
when absent, retaining explicit property/environment overrides and undefined
property fallback. Five focused866 repetitions pass5.764s. No invented tests or
weakened assertions. Nine new fixture files are byte-identical to Java.
Full offline workspace passes (root430.386s, SANY0.979s, TLC64.206s).
Full TLC race passes567.289s; all check handles terminal.
Inventory747/1269 contexts,312/626 complete classes;522 contexts/314 classes
pending,one partial. Next pending issue family Github971a/b/c/d/e; all unchanged
Java references pass. Scratch translation preserves original three/four-count
queue latches and all delegation through the existing runner queue injection;
not yet activated or credited. Goal active; user-deferred topics unchanged.

Next trace-topic preflight: unchanged DumpAsDotTest passes, including exact DOT
bytes, postcondition register42 and zero-uncovered coverage. Unchanged
DistributedTrace fails assertNoTESpec (trace spec generated) and inherited
SUCCESS0 versus actual safety12. Keep pending; do not weaken assertions or
invent a skip. Logs original-DumpAsDotTest-junit.log and
original-DistributedTrace-junit.log in ignored correctness-java scratch.

2026-10-04: Github971a/b/c/d/e original testSpec and beforeSetUp queue wrappers
translated. All five unchanged Java JUnit references pass. Six original model/
config fixtures match bytes. First focused Go run passes5.210s; five combined
normal repetitions pass8.589s and five root race repetitions pass80.632s.
Original two workers and three/four-dequeue CountDownLatch semantics retained:
enqueue into MemStateQueue before selected-state await; count down before each
single dequeue; delegate all other queue methods unchanged. Existing runner
StateQueue injection represents Java's testing-only Factory.sq; no production
changes needed. Original lncheck/config settings, no coverage/DOT/debugger/JSON/
trace generation, diagnostics, exact counts/depth where asserted, property names,
complete state/action/ordinal traces and loop-back2 assertions preserved. Go's
Value.Equal error is propagated as Java's throwing equals, not discarded.
No weakened assertions or invented tests. Full offline workspace passes
(root424.096s, SANY1.056s, TLC62.045s); all check handles terminal. TLC production
unchanged since9083c0d, whose full race passed567.289s. Inventory752/1269 contexts,317/626 complete classes;517 contexts/309
classes pending,one partial. Issue-regression topic now98/98 classes and101/101
contexts complete. Next eligible trace case DumpAsDotTest: unchanged Java passes,
including exact DOT bytes; DistributedTrace remains pending with two original
Java assertion failures, and two EWD840 DumpLoadTrace methods retain their
previously documented timing discrepancy. Goal active; user-deferred topics
unchanged.

2026-10-04: original DumpAsDotTest.testSpec translated, with unchanged Java
JUnit reference passing and three new fixtures plus reused CodeplexBug8
byte-identical. Original exact DOT master comparison,18/11/0 counts,
FINISHED/no GENERAL, postcondition absence checks, register42 firstInt18,
zero-uncovered coverage and liveness exit retained. All default coverage,
DOT/JSON/debugger/forced trace generation and explicit colorize/actionlabels/
stuttering options preserved. First Go failure exposed actual production
shortcuts: numeric rank-node sorting, insertion-order action legend, and an
early stuttering return that skipped rank maintenance/snapshot. StateWriter
now mirrors Java's HashMap/HashSet order, source label/legend formatting,
stuttering continuation and constructor header flush. dot_hash_map.go ports
only the operations DOT uses for Comparable Integer/Long/String keys, including
put vs computeIfAbsent insertion/resize timing, bucket splits, red-black
collision trees, tree root/list order and untreeification on resize. Source
OpenJDK21 HashMap and local JDK21 bytecode consulted; copyright/license retained.
Ignored CLI probes (not permanent tests or inventory credit) match Java byte
for byte for3374 iteration/lookup snapshots (10,433,564 bytes), including signed
Long/hash collisions, UTF16 String collisions, resizing and tree splits.
No master normalization, weakened assertions or invented permanent tests.
Final five normal repetitions pass5.679s; final five focused root race
repetitions (DumpAsDot, existing DOT cases and Github391) pass61.651s.
Constructor-flush review superseded earlier broad verification:41682/29419
terminated130;87639 passed61.667s before final flush and is retired. Final full workspace51679 passes (root416.794s, SANY0.983s, TLC62.560s);
retired. Final full TLC race63793 passes551.758s; retired. All final verification
handles terminal; ready for green commit. Inventory753/1269 contexts,318/626 complete classes;
516 contexts/308 classes pending,one partial. Next eligible topic: generated
TTrace correctness variants, preserving original generation/recheck pipeline;
first source Github461Test_TTraceTest and Github597Test_TTraceTest inspected.
DistributedTrace remains pending with original Java failures, two EWD840 binary
methods retain documented partial-check timing limitation. Goal active;
user-deferred topics unchanged.

First TTrace Java preflight: Github461Test then Github461Test_TTraceTest and
Github597Test then Github597Test_TTraceTest all pass unchanged. Generated
prerequisite files were explicitly checked present before recheck, so these
are executed references, not missing-artifact assumption skips. Logs
 ttrace-first-<Class>-junit.log; javac ttrace-first-javac.log;16957 terminal.
Original TTrace base uses generated class-name TTrace.tla as both model/config,
noGenerateSpec=true, coverage=false, default JSON/DOT/debugger, resolver with
original spec path.597 additionally omits fp/seed (random), disables DOT and
keeps debugger/JSON. Its original first model is dekker with embedded config;
source first-Go port already preserves random fp/seed. No Go TTrace port credit
yet; upcoming tests must generate Go artifacts then recheck, retaining exact
461 five-state trace/actions/ordinals and597 diagnostic/loop existence checks.

Coverage reporting preserves the checker's action array. Java CostModelCreator
iterates a separate sorted TreeSet; Go sorts a copied slice before reporting
shared cost models. Sorting Tool.Actions in place during a periodic report
changes successor insertion order, which can alter tableau expansion and the
serialized liveness graph. The original CodePlexBug08EWD840FL2 graph-size
assertion exposed this during a run long enough to report coverage mid-search.

LiveChecker.Close mirrors Java AbstractLiveChecker.close: the explicit
ModelChecker.vetoCleanup property retains disk graph handles for subsequent
inspection, while the liveness writer still closes. A local checkpoint cleanup
veto only preserves metadata and does not retain graph handles. The original
CodePlexBug08 tests set the explicit property, flush the retained graphs and
check actual nodes_0/ptrs_0 sizes. Their Go teardown closes all retained graphs
after those assertions, before restoring checker statics or removing metadata.

2026-10-04 initialization parity discovered by original example postconditions:
Java parses ModelConfig before first active use of the built-in operator and
native-module classes. Their cached UniqueStrings therefore initialize lazily
within the current worker/classloader context. Eager Go package initialization
had allocated126 tokens before ACP's config model values, changing fingerprints
and the selected counterexample. classInterning now guards each class's names
with a per-context atomic fast path and mutex. Exported built-in names remain
available as unregistered names until first use; runtime nodes canonicalize them.
Config/model fingerprints still use actual UniqueString tokens, as Java does.
SANY StringNode also interns its unquoted value during semantic generation.
The bridge preserves that semantic module/source order before visiting runtime
definitions, whose alphabetical conversion must not choose string tokens or
function-set enumeration order. Existing checker/simulator records initialize
cached keys before constructing their field arrays. All nine original example
methods pass and are credited; see progress for broad and race verification.

The original CommonTestCase isolates TLC statics per test class with
IsolatedTestCaseRunner. CLI parsing writes -lncheck into TLCGlobals; its
default must therefore be fresh for later source classes, including the
infinite-state trace-expression lasso test requiring partial liveness checks.
The Go source-test runner resets LNCheck to default before parsing the
original class's flags and restores its outer value afterward. Production CLI
behavior and each original explicit strategy remain unchanged.


Source and INSTANCE operator identities: Java SANY Generator creates a symbol
per source declaration and a distinct qualified declaration per instantiated
export. A bridge lookup alias such as EWD998Chan!EWD998!Spec can also name a
module-local export; aliases cannot own semantic declaration identities. Source
symbols and prepared instance export symbols have separate identity caches.
LET's predeclared operator symbols are reused so recursive/local references
continue to bind to the same definition. Module context conversion validates
the instance owner and source position before taking an alias-table entry;
a nested export otherwise clones the operator from the actual instancee
context, preserving the outer qualified name and substitution boundary.


Liveness path failure diagnostics preserve AbstractDiskGraph.getPath's exact
Java message: "Couldn't re-create liveness trace (path) starting at: ... and
tidx: ...". Go fingerprints remain uint64 for storage, but diagnostic values
use int64 to match the source Java long's signed decimal representation.
Both DiskGraph and TableauDiskGraph keep that same failure boundary.

The failure also retains Java's RuntimeException family via the existing
generic exception type. Original graph test catch blocks reject I/O errors;
returning an untyped error and accepting every error would lose that boundary.


Specs.addSubsts reconstructs each wrapper with Java's SubstInNode(source,body)
constructor, preserving the source syntax tree and shared substitution array.
Go NewSubstInNodeFromSource carries TreeNode/Location/Substs while allocating
a fresh semantic node and replacing only the body. Unnamed action locations
and CounterExample action records therefore retain the INSTANCE source range.

Periodic checker coverage calls ReportCoverage(tool,StartTime), matching Java
AbstractChecker.reportCoverage/CostModelCreator.report, including END and
runtime overhead classification. The recorder keeps every periodic report;
source zero-uncovered assertions can observe transient zeros when diagnostic
instrumentation slows a model. Their original all-report semantics are retained.

LiveCheck.check0 returns ECTLCTemporalPropertyViolated with no thrown exception,
as Java does; AddAndCheckLiveCheck deliberately ignores intermediate result
codes and continues insertion. CheckTrace retains its separate source throw
when its final result is nonzero. This distinction preserves complete model
statistics and the forced partial shortest counterexample.

The evaluator's undefined primed fairness-variable branch carries code2148,
variable name, exact expression location and failed expression/context through
Java's detailed runtime exception path. Generic text loses recorder parameters.

MSBDiskFPSet.TLCIterator.getLast scans buckets/slots backward for signed-positive
unflushed entries without moving its cursor. reads returns successful reads.
Exhaustion uses NoSuchElementException; the strict increasing assertion uses
Java's coded runtime carrier. Existing successful next/flush paths are unchanged.

## Canonical level-dependency keys

`sany_level_parameters.go` ports Java `ParamAndPosition` and `ArgLevelParam`
using actual canonical symbol references. Equality compares references rather
than names or semantic UID equality. Positions and hash arithmetic use signed
Java `int` behavior. `occur` rejects a null array and scans a non-null array in
source order, allowing null symbol slots to match null stored references.
Formatting reads each symbol's current syntax tree and honors
`SemanticNode.showPlainFormulae` through the existing TLC formatting bridge;
nullable references print `null`. Constructor identity and native checker
integer IDs are not substituted for these canonical references.

These classes are prerequisites, not a completed level checker. Inherited level
metadata and recursive checking still require porting before closing the
remaining incremental operator tests.

## Canonical level-constraint maps

`sany_level_constraints.go` ports `SetOfLevelConstraints` and
`SetOfArgLevelConstraints` over `tlc.JavaSemanticMap`. The first tightens upper
bounds with `min`, defaulting to Java `MaxLevel`; the second tightens required
argument levels with `max`, defaulting to `MinLevel`. Retain even vacuous entries.
Null input levels throw, while previous nullable values are returned. Raw map
copy construction preserves values, including nulls and values outside those
bounds, without invoking the derived tightening method. `putAll` instead invokes
that method in source key-set order and retains earlier writes on failure.

Symbol keys use SemanticNode's concrete-class/kind/UID equality. Argument keys
use ParamAndPosition's actual symbol-reference equality. Keys retain their first
stored object after equivalent-key replacement. Formatting follows Java key
iteration, nullable names and values, and the current syntax-tree location for
argument keys. HashMap copy construction uses its source initial capacity and
bucket layout; iterators save their next node before returning the current entry
and check structural modifications on the next call. Tree ties between different
symbol classes use Java class names; native object identity supplies ties within
a class, as in the existing collection ports. The 680 comparison rows cover
bucket growth, copies, merges, removals, nulls, equality and iterator boundaries;
they do not establish cross-runtime object-identity tree ordering.

## Inherited canonical LevelNode data

`sany_level_node.go` supplies Java's common mutable level fields to canonical
SANY constructors through their semantic base. Each constructor owns fresh
symbol sets, argument-dependency sets and constraint maps; OpDef and OpDecl no
longer keep duplicate level fields. Symbol sets use concrete-class/kind/UID
equality, while argument-dependency sets use source ArgLevelParam reference
comparisons. The sets retain their Java backing-map iteration and mutation
behavior, including capacity-preserving clearing.

Getter guards reject only a zero `levelChecked`, preserving negative Java
iterations and the source non-Leibniz getter's `getAllParams` error message.
The next-iteration entry point uses signed Java wraparound and concrete-node
check dispatch. Formatting and subnode aggregation call virtual getters through
the canonical interface. Aggregation records the iteration before traversal,
skips ModuleKind and ModuleInstanceKind, still checks InstanceKind, visits later
children after a false result, and retains all prior metadata writes on failure.
The temporal-constant helper tests source symbol kind rather than concrete class.
Formal parameters populate level/all-parameter sets only while their iteration
is zero; declarations retain constructor data and return true without updating
that iteration. These are the complete original checks for those two classes.

The interface covers canonical SANY classes and views of actual TLC literal
nodes. Class-specific application, instance, proof and module checking and
shared evaluator collection ownership remain incomplete. The 575 direct Java observations
and retained 24,529 graph rows verify this inherited-data integration; they do
not close the remaining original incremental semantic methods.

## Canonical operator level checking

`sany_operator_level.go` translates the complete original `OpArgNode.levelCheck`
and `OpDefNode.levelCheck` algorithms. Operator arguments record their iteration,
check the actual operator, and share its level/all-parameter sets, constraints
and argument dependencies. Preserve the source omission of non-Leibniz sharing.
The primary constructor raises a typed null failure after allocating its UID.

Operator definitions retain the source recursive/nonrecursive cache guard and
record their iteration before checking descendants. Levels and weights grow
monotonically, maximum argument levels tighten, and higher-order minimum levels
and argument-dependency conditions rebuild each computation. Bound formals are
removed from parameter sets and constraint copies. Leibniz flags only change
from true to false. Argument dependencies persist unless both symbols occur by
reference in the formal array. Constraint copies preserve raw values. A numbered
step checks its step node, then the inherited same-iteration guard skips its
aggregation; preserve this source behavior. Body reads remain fresh after child
checking, so `setBody` replacements affect subsequent metadata reads.

Metadata getters preserve source guards, variadic indexing and nullable-table
fallbacks. Maximum-level and weight getters copy their arrays; the Leibniz getter
returns the actual mutable array. Body and parameter accessors retain their
reference identities. The original builtin initialization test now calls the
same getters as Java, retaining its assertions and null array-length behavior.

The 792 external observations use canonical source symbols and controlled
ExprNode bodies to verify the complete metadata algorithm and failure order.
They do not establish checking of all generated TLA bodies. Application,
instance, proof/module checking and shared evaluator integration remain pending.

## Shared canonical literal level metadata

Canonical SANY bases now hold pointers to the actual `tlc.SemanticNodeBase` and
its canonical metadata. The level and signed iteration cells alias the base
fields, so copying a canonical node base does not split them. Generated numeral,
decimal and string bodies retain their concrete TLC identity and syntax node.
`sany_literal_level.go` exposes those bodies through views that allocate no UID
and reuse the canonical collections stored on the literal base. Those are real
mutable symbol sets, argument dependencies and constraint maps, not constant
or empty stand-ins. Root generation initializes them; TLC-only constructors
initialize them when first accessed by the canonical interface. Direct TLC
level checks and setters share the scalar cells. Getter guards reject only
iteration zero, and next-iteration checks preserve signed wraparound.

The legacy TLC parameter getter still returns its TLC symbol slice projection.
It does not expose canonical root symbol sets; integrating that evaluator API
remains work. `SpecsGetLevel` now reads parameters before level, as Java does,
and requires already checked literal nodes. Its old synthetic test now performs
that missing setup, retaining all assertions. Sixty direct Java observations
verify literal checks, mutation and mixed aggregation; all retained common,
operator and module observations remain equal. The complete original
`basicOpDefTest` now runs the actual incremental operator generator and checker.


## Canonical LetInNode level checking

`sany_letin.go` now translates the complete source `LetInNode.levelCheck` method.
It records the iteration before traversal, resets correctness for each fresh
iteration, skips ModuleInstanceKind definitions, and checks definitions, body,
then instances even after false results. The resulting level comes only from
the body. Body level/all-parameter sets are copied; constraints and argument
dependencies merge into retained collections in source order. Definition
argument dependencies bound by either local formal reference are excluded.
The theorem branch uses an empty formal array. Only argument dependencies are
imported from instances, and non-Leibniz parameters are intentionally untouched.
Null failures retain earlier writes. The body accessor returns its actual node.

`NewJavaSemanticSetCopy` supplies HashSet(Collection)'s actual OpenJDK 21 backing
map behavior: reserve max(size, 12) mappings, compute ceil(mappings / 0.75), then
round up to the map's power-of-two capacity before adding the source entries.
This differs from HashMap(Map) and retains distinct empty-copy allocation state.
The source for this constructor behavior is the installed JDK 21 bytecode.

All 68 external Java/Go observations match, including complete metadata text and
collection order, false/repeated/decreasing checks, skipped definitions, null
failure state, and copy capacities across nine sizes. These observations add no
original-test credit. Application, instance and module descendant
checks are translated below, enabling the complete original incremental LET tests.


## Canonical LabelNode level checking

`sany_label_level.go` translates the source label check and all six overridden
metadata getters. A cached check returns true regardless of the body's previous
result or the label's inherited correctness field. A fresh check records iteration
before traversing non-null formals and then the body. It returns the body result
without copying that result into its own correctness field. Null formal entries
are skipped; a null formal array fails after recording iteration.

Level, level/all-parameter sets, both constraint maps and argument dependencies
are read directly from the current body through virtual getters. Preserve the
source's TheoremNode/ThmNode zero-iteration error messages. The non-Leibniz getter
is inherited and reads the label's separate set; do not delegate it. Other
inherited scalar and collection fields are not populated as side effects.

Eighty-one external Java/Go rows agree exactly, including full metadata text,
set/map identity, signed iterations, false body results, null body/parameter
failures, body replacement with an unchecked literal, direct mutation and
operator definitions whose bodies are labels. These scratch observations add no
original-method credit. General application, substitution, instance,
proof/module checking and evaluator collection sharing remain incomplete.


## Canonical theorem and assumption definition level checking

`sany_theorem_level.go` translates `ThmOrAssumpDefNode.levelCheck` and its metadata
accessors. This differs from ordinary OpDef checking: every fresh iteration
recreates maximum-level and weight arrays, assigns the current body's level
directly, and rebuilds higher-order tables. Preserve the source weight assignment
to itself and the Leibniz array allocation inside the parameter loop. With no
parameters the Leibniz fields retain their prior values; with parameters only
the last slot is initially true before non-Leibniz removal. Do not substitute
the ordinary operator's monotonic algorithm or repair Java's apparent quirks.

Parameter and dependency sets retain accumulated members across fresh checks.
Constraints are raw copies with bound formals removed; argument dependencies
persist unless both symbols occur by reference in the formal array. All guarded
accessors retain source zero-only checks, variadic indexing and nullable-table
fallbacks. The Leibniz accessor returns the actual mutable array. Partial failure
keeps prior writes and table allocation. The production body remains the actual
semantic node and all reads use current virtual getters.

All 458 external Java/Go output rows match across zero to three formals,
theorems/assumptions, false results, repeated/decreasing/overflow iterations,
complete metadata and getter matrices, twelve failure cases, Leibniz aliasing
and a theorem definition in a LET. These scratch observations add no original
method credit and do not establish checks for every generated descendant class.


## Canonical proof-node subnode checking

`sany_proof_level.go` translates the complete source checks for UseOrHideNode,
LeafProofNode, DefStepNode and NonLeafProofNode. USE/HIDE and leaf proofs check
only facts; DEF references, ONLY and omitted flags do not change the check.
Definition steps check their actual definition array. Non-leaf proofs allocate
an array of steps followed by instances before invoking common aggregation.
Their nil-array failure occurs before recording the iteration; the other three
classes defer array validation to the common helper after recording iteration.
All source cache guards remain intact.

Common aggregation now accepts indexed child access so actual graph children
are resolved as traversal reaches them. Do not eagerly adapt a whole array:
a later unsupported or null child must not suppress earlier checks and writes.
Canonical-array callers use the same unchanged aggregation. A definition step
captures the array object passed at entry, so replacing the parent field does
not replace its current traversal, while mutation of a captured element does.

All 282 external Java/Go rows match exactly, including all five USE/HIDE/proof
variants, seven array/failure modes, eight signed/repeated iterations, complete
metadata, child-call counts, literal facts, and both array mutation behaviors.
All 575 retained common LevelNode observations also match. These observations
add no original-method credit and do not complete all descendant, theorem,
ASSUME-PROVE, instance or module checks.


## Canonical NEW-symbol level checking

`sany_new_symbol_level.go` translates the complete source NewSymbNode check.
A fresh iteration is recorded before checking the actual declaration. The level
starts at the declaration level, then combines with the optional set level.
Set correctness replaces the current correctness only when a set exists.
Preserve the source exact equality against TemporalLevel, reporting E4356 with
`Level error:\nTemporal formula used as set.` before sharing metadata. A null
syntax node or Errors receiver leaves the earlier level/correctness writes.
Diagnostic equality uses Java Location's source and four coordinates, ignoring
Go Position's extra end-coordinate bookkeeping. This fixed diagnostic has no
format parameters, and equal reports are deduplicated in the caller's log.

After checking, level/all-parameter sets, constraints and argument dependencies
alias the set's actual collections. The non-Leibniz set stays separate. Removing
the set on a later iteration does not clear those aliases or reset correctness.
A typed-null graph reference has the same behavior as Java null. Body metadata
reads remain virtual and use the current set field.

All 195 external Java/Go rows agree exactly across four declaration levels,
six construction/failure modes and eight signed/repeated iterations, complete
metadata and Errors output, identity checks, set removal/restoration and a
pre-seeded duplicate diagnostic. These observations do not add original-method
credit, establish general Errors parity or close other ASSUME-PROVE descendants.


## Canonical assumption level checking and shadow iteration

`sany_assumption_level.go` translates AssumeNode checking, six delegated getters
and its overridden level-data formatter. Unlike LabelNode, these delegated
getters have no guard on the assumption's own iteration. The non-Leibniz getter
stays inherited. Preserve Java's separate AssumeNode `levelChecked` field, which
shadows the LevelNode field: checking records `assumeLevelChecked`, while the
common next-iteration overload reads the inherited cell. A cached assumption
check returns true even after its prior expression check returned false.

The expression is checked first, then an optional definition is checked even
after false. A nonconstant expression reports E4206 but does not force the result
false. Preserve exact message, location, fixed-format/integer-parameter duplicate
equality and argument evaluation before a null Errors failure. If the result is
true, temporal-constant constraints are applied to the assumption's inherited
fields, not its delegated body collections. These inherited fields are not
populated by this check. The common formatting entry point now dispatches actual
class overrides before its inherited guard; the assumption formatter prints
Java collection/map forms and skips that guard.

All 160 external Java/Go rows match exactly across ten modes and eight iteration
phases, including the no-iteration overload, both counters, named definitions,
false results, null bodies/logs/own sets, constant-kind filtering, exact Errors
text, getter identity and overridden formatting. All 575 retained common rows
still agree. These observations add no original-method credit. TheoremNode has
its own analogous shadow counter, now retained by its source check below; other
canonical graph/evaluator dependencies remain incomplete.


## Canonical theorem-node and temporal-proof level checking

`sany_theorem_node_level.go` translates the complete TheoremNode check and its
private recursive LevelCheckTemporal helper. TheoremNode retains a shadow
iteration counter: fresh checking records it, then common aggregation records
the separate inherited counter. A cached theorem check returns true. Subnodes
contain the named definition when present, otherwise the assertion, followed
by its optional proof. Preserve partial writes and the common next-iteration
overload's use of the inherited counter.

PICK diagnostics inspect the assertion's raw inherited level, not a delegated
getter. For an exact temporal level, each nonconstant bound reports E4354 without
forcing the returned result false. The temporal-proof helper processes only
NonLeafProofNode, ignores other steps, reports E4352 for nonconstant TAKE,
WITNESS or HAVE, and reports E4353 for a nonconstant CASE before recursing into
its proof. QED recurses without a level restriction. Preserve source else-if
ordering, null failures, actual syntax ownership and fixed-message/location
deduplication. Do not reinstate Java's commented-out prohibition on temporal
facts in a non-temporal proof.

All 223 external Java/Go rows agree exactly. Normal checking uses actual source
Theorem/ThmOrAssumpDef/LeafProof/NonLeafProof nodes and controlled ExprNode bodies.
Temporal application cases use real source OpApplNode objects with explicitly
prechecked metadata, and invoke the actual source temporal helper; PICK cases
also exercise the source inherited-counter cache. These verify theorem logic,
not ordinary OpApplNode.levelCheck, which is translated below. No source test is
replaced with prechecked metadata and no original-method credit is added.


## Canonical ASSUME/PROVE level checking

`sany_assume_prove_level.go` translates the complete source AssumeProveNode
algorithm. Each fresh iteration records its counter and resets correctness.
The first assumption pass skips null entries and collects boolean failures;
PROVE is checked next but its boolean result is deliberately ignored. The
second assumption pass calls their checks again without a null exemption, then
compares levels. Preserve both virtual getLevel calls when an assumption exceeds
the current level; do not replace them with one max call.

Level/all-parameter sets, constraint maps and argument dependencies merge into
retained collections, with PROVE first in each phase and assumptions in order.
Non-Leibniz parameters remain untouched. Temporal-constant constraints are added
only if the assumption checks succeeded. Every null failure retains preceding
counter, correctness, level and collection writes. Current prove/assumption
fields are read afresh, matching source traversal rather than capturing the
whole graph before checking.

All 143 external Java/Go rows agree exactly across zero to three assumptions,
false PROVE/assumption results, correctness resets, signed/repeated iterations,
fourteen partial-failure scenarios, complete metadata and child-call counts,
and a virtual level getter that changes between reads. These observations add
no original-method credit and do not complete unported descendant classes or
shared evaluator integration.


## Canonical substitution level helpers

`sany_substitution_level.go` translates Subst.paramSet, allParamSet, getSubLCSet,
getSubALCSet and getSubALPSet. Lookup uses symbol reference identity and the first
matching substitution. Parameter helpers return the actual replacement set;
unmatched parameters get a fresh HashSet singleton, including null. A matched
null replacement fails in these helpers, whereas getSub returns null. The
existing getSub lookup now normalizes typed-null references to Java null.

LC translation tightens constant/variable declaration bounds only for a
nonconstant module, expands replacement parameters and checks actual defined
operators before reading their maximum argument levels. Preserve the OpArg cast
before testing whether its operator is defined. ALC translation retains original
unsubstituted key objects, renames substituted parameter operators and checks
parameter replacements before reading their levels. ALP translation retains
original unsubstituted dependency objects, expands parameter operators and drops
defined-operator dependencies. All maps/sets use the source hashing, iteration,
reference equality and tightening rules; null failures retain source phase order.

All 130 external Java/Go observation rows agree, covering constant-module modes,
null/duplicate substitutions, shared set identity, same-UID distinct symbols,
nullable values, null collection keys and a dependency mutated after insertion.
Runtime failure observations compare NPE/ClassCastException classes, not enhanced
JVM messages. These are manual observations, not original-method test credit.
SubstInNode/APSubstInNode checks are translated below. InstanceNode checking is translated below; shared evaluator integration remains
separate implementation work.


## Canonical module level checking and constant classification

`sany_module_level.go` translates ModuleNode.levelCheck and isConstant. A fresh
module iteration first processes each contiguous recursive section. Initialize
recursive definitions to iteration 1, action argument bounds and unit weights;
reset nonrecursive definitions to iteration 0. Check each at iteration 1, report
E4290 for primed recursive arguments and collect recursive levels/parameters.
Propagate the maximum level and both parameter unions to every definition in
the section, then perform the source second pass at iteration 2. Preserve all
counter writes and failures before the module correctness reset.

Check inner modules, operators, theorem/assumption definitions and top-level
statements in source order at iteration 1. Preserve captured local arrays and
live theorem-definition field traversal. Add constant declarations to retained
parameter sets; for a nonconstant module, add constant-level bounds. Merge
operator and top-level constraint maps and dependencies, filtering dependencies
bound by operator formals. Source omits module-level non-Leibniz propagation.

isConstant returns false immediately for variable declarations. Otherwise it
invokes the real module check at iteration 1, inspects actual operator body
levels and walks the current theorem vector. Module getLevel always raises the
source WrongInvocationException. Its formatter uses guarded inherited getters
and Java collection representations, never the forbidden getLevel call.

All 332 external Java/Go observations agree. Actual source Module/OpDef/Theorem
objects use controlled ExprNode metadata and explicit arrays, covering empty,
constant/variable, false-result, nested-module, recursive multi-section and
theorem cases over signed/repeated iterations. Observations retain child-call
counts, recursive tables, complete collections, diagnostics, formatting and
twelve partial failures. NPE observations compare exception class rather than
JVM enhanced messages. No original test method is credited. Ordinary application,
remaining generation/evaluator integration still prevents whole-module completion
beyond the original incremental LET methods completed below.


## Canonical substitution-wrapper level checks

`sany_substin_level.go` translates the complete SubstInNode and APSubstInNode
algorithms. Both retain source cached correctness and reset correctness on a
fresh iteration. Check the current body first, then each current substitution
replacement. Read the current body again for its level and parameter set;
substitutions whose declarations occur in that set contribute to the maximum
level. Expand level parameters into the retained destination set.

SubstIn makes real HashSet copy-constructor copies of body all/non-Leibniz
parameters, then processes substitutions sequentially against those current
copies. Remove a present declaration, add replacement all parameters and merge
replacement non-Leibniz parameters. If the declaration is now in the non-Leibniz
set, remove it and add replacement all parameters. Preserve this ordering:
later substitutions see symbols introduced by earlier ones. APSubstIn instead
expands the body's all parameters into its retained set and leaves its own
non-Leibniz set untouched. Do not merge these distinct algorithms.

Both ask the actual instantiated module isConstant method, then replace LC,
ALC and ALP collections in source order using the ported Subst helpers. Fresh
body/replacement reads, null-array failures, copied-set assignments and earlier
writes remain visible when later phases fail. Constructors, semantic identities
and the shared Subst entries remain unchanged.

All 768 external Java/Go rows agree across both classes, constant/nonconstant
modules, 24 modes and eight signed/repeated iteration phases. Observations use
actual source wrappers through their private constructors, actual OpArg/OpDef
nodes and controlled ExprNode metadata. They compare complete inherited level
data, child calls, module iteration and collection identity, including successful
chained/duplicate substitutions, false child results and malformed metadata.
Runtime NPE/CCE observations compare class, not enhanced JVM messages. These
are manual evidence, not original-method credit. The complete original incremental LET methods are translated below.


## Canonical instance level checking

`sany_instance_level.go` translates the complete InstanceNode algorithm. A fresh
iteration records its counter, replaces levelParams with a fresh HashSet and
resets correctness. Check the actual module and every replacement; preserve the
source repeated child checks and short-circuit boolean evaluation. Nonconstant
modules restrict replacement levels to declaration levels. Then check explicit
LC bounds, defined operators' maximum argument levels and co-parameter ALPs.
Report E4245/E4246/E4247 only when the corresponding child check succeeds, but
still make correctness false when a bound fails. E4244 reports a non-Leibniz
operator without independently making the returned result false.

Preserve captured replacement references within each constraint phase and fresh
field reads in the nested co-parameter loops. Compute translated module LC/ALC/
ALP sets and merge replacement metadata in source phase order, retaining only
constraints/dependencies not bound by the instance's actual formal references.
Existing LC/ALC/ALP sets are tightened/extended, not reset. All/non-Leibniz sets
and raw inherited level remain unchanged. Preserve partial writes on failure.

Instance getLevel returns constant level even before checking; its return type
now satisfies the canonical LevelNode interface. getLevelParams and
getLevelConstraints bypass the inherited iteration guard, as in Java. Other
inherited getters retain their guards. Its formatter reads the raw three
constraint/dependency fields, uses Java collection forms and has no iteration
guard. Fixed-format level diagnostics retain parameter lists and compare them
when deduplicating, along with the source code, rendered message and Location.
Recursive-module E4290 now retains its two parameters too. This is class-specific
fixed-format diagnostic parity, not a claim of complete general Errors parity.

All 545 external Java/Go rows agree, including diagnostic parameter lists, full
metadata, both formatters, child calls, signed iterations and set identity.
Most observations use actual source modules with prechecked controlled metadata
to isolate Instance logic; one mode checks a fresh actual module. Include failed
children, argument/co-parameter bounds, diagnostic suppression, non-Leibniz
warnings, null logs/fields, formal filtering and successful chained/duplicate
substitutions. A final actual-instance case retains two distinct parameter lists
that render the same message. NPE/CCE observations compare class rather than
enhanced JVM messages. Retained 332 module rows still agree. No original method
credit is added; standalone generation and shared evaluator integration remain.


## Canonical application level checking

`sany_application_level.go` translates the complete OpApplNode.levelCheck and
private getArg helper. AnyDef dispatch covers actual OpDefNode and
ThmOrAssumpDefNode objects. Check operands and ranges first, skipping null entries
only in the source phases that do so. Preserve cached correctness, repeated child
checks and diagnostic suppression; a false operator check does not independently
force application correctness false. Check argument bounds, higher-order minimum
bounds, co-parameter conditions and temporal ranges with the source diagnostics
and parameters.

Defined applications compute weighted levels and propagate operator/operand/range
parameter sets. Remove bound symbols from all three sets, filter operand LC/ALP
metadata as specified and retain the source later additions of argument bounds.
Translate higher-order LC/ALC/ALP dependencies through actual argument definitions
and formal references, including non-Leibniz argument propagation. Preserve the
source fresh ALP allocation after ALC computation; other destination collections
remain retained. getArg captures the formal array, scans the current operand
array and returns the first reference match, preserving null and bounds failures.

Declared applications check their operator and operands, replace LP/LC/ALC/ALP
sets/maps in source phase order and retain all/non-Leibniz sets. Preserve the
source omission of range metadata from this branch. Both branches finish with
stuttering checks for []/<>, action restrictions for ~>/-+->, mixed logical
argument checks and action-bound restrictions on temporal quantification. Name
checks call getName().toString semantics, so a null formal name fails after all
preceding metadata writes. Fixed-format diagnostic parameters now compare actual
symbol references and normalize typed-null symbol arguments to null.

All 2,897 external Java/Go rows agree: 1,152 normal checks over defined, theorem,
declared/formal and builtin operators; 1,536 temporal cases using fresh actual
child applications; 96 malformed cases retaining every collection separately;
112 direct private getArg observations; and one null-name case. Actual source
application constructors are followed by explicit controlled metadata/array
assignments for manual observation, not substituted into original tests. Compare
full metadata, collection identity, child calls, diagnostics and raw parameter
values. Runtime NPE/CCE observations compare class, not enhanced JVM messages.
Retained 545 instance and 332 module rows agree. These add no original-method
credit or full-workspace verification claim. The original standalone LET tests
now retain exact generation, syntax ownership and source-reference assertions below;
shared evaluator collection integration and broader graph/parser fidelity remain.


## Public SANY canonical level-check phase

`CheckSanySpecLevels` calls `sanyLevelCheckNext` on the actual generated root
`ModuleNode`, matching Java `SANYFrontend.checkLevel`. It returns the source
level-check result combined with diagnostic success. It no longer substitutes
the older AST level analysis for the canonical graph. The complete original
51-row level matrix and INSTANCE minimum-level diagnostic pass, as do the
unchanged 52 Java cases. Diagnostic integer assertions use Java-width `int32`.
Generation and level logs remain separate.

The main `CheckSpec`/TLC driver and single-module helper now also call the actual
generated ModuleNode. The driver retains Java SANY's external-module order,
table publication before checking, shared diagnostic log and raw Errors-success
gate. Canonical checks replace the older AST check plans at these entry points.
Generating an incomplete graph must expose its missing implementation rather
than silently falling back to AST summaries.

Canonical named and unnamed INSTANCE generation no longer compares its export
count with the older AST-derived export list. Java iterates its actual Context;
it has no such comparison. In particular, recursive declarations keep their
source locality flag after LOCAL completion, so the two lists can legitimately
differ. Child/body completeness guards still expose unported canonical graphs.
The TLC bridge reads static definition levels from these checked nodes,
including actual Context theorem/assumption definitions. It no longer uses the
XML-exporter's estimated levels. The bridge still constructs runtime definitions
from AST views; complete canonical graph/evaluator sharing remains pending.

`sany_at_level.go` translates the complete EXCEPT AtNode level algorithm. It
captures the enclosing EXCEPT operand array, ignores child correctness and
checks the containing component before breaking the level loop. Subsequent
metadata passes merge the base and preceding components only. Collections are
retained and non-Leibniz parameters are untouched. An external manual Java/Go
comparison agrees on all 56 rows, including component positions and repeated
references, iteration caching, child calls and complete level metadata. No
original-method credit is claimed for manual observations. AssumeNode now also
retains its source integer diagnostic parameter. Full SANY and 26 original TLC
model methods pass; this is not full-workspace completion.

## Complete original incremental semantic LET tests

Root `sany_incremental_semantic_java_test.go` now translates both original LET
methods, completing all five IncrementalSemanticParseTests methods. Parse the
original standalone snippets and compare the single parser dependency. The test
helper loads the actual embedded Java standard-module source through production
SANY loading, generates each dependency and invokes its actual canonical level
check before publishing its context/module in ExternalModuleTable. Naturals,
TLC, Sequences and FiniteSets embedded files match pinned Java source byte for
byte. There are no synthetic enclosing modules, prechecked metadata or shortened
assertions in these original tests.

Generate the expression using the actual external table and a dummy module,
then preserve the source assertions: generation log success, non-null result,
actual level checking and its log success, original syntax reference, constant
level, exact LetInNode/OpApplNode/OpDefNode classes and the imported operator's
actual source-definition reference. OpDefNode getSource/hasSource now expose
Java's immediate source/self fallback without recursively flattening the source
chain; typed-null symbol references match Java null. Supplementary native AST
facade checks remain separate and add no duplicate original-method credit.

The unchanged original Java JUnit class passes all five methods, and all five
Go translations pass. Only the two newly translated SANY methods gain completion
credit; main TLC inventory totals do not change. This does not establish general
incremental-generator, corpus AST or whole-workspace completion.

## Canonical construction for validated subexpression selectors

`sany_selector_graph.go` follows the actual semantic graph after native selector
validation. Reuse arguments already generated in the caller's scope; never
regenerate the selected AST body. Retain actual declaration formals, label nodes,
LET contexts and substitution prefixes. Follow name, label and operand modes,
including quantifier parameters, record pairs, CASE pairs, EXCEPT bases,
ASSUME/PROVE and its SUFFICES wrapper, and LAMBDA arguments. Finalization retains
source formal cloning, prefix wrapping, LAMBDA/application construction and
`subExpressionOf` ownership. Preserve Java's original argument-array choice in
the selected operator-argument branch rather than repairing its source behavior.

Generation flags must be captured before recursive argument generation clears
the mutable fields. Instance and general selection helpers receive those captured
flags, preserving operator arguments as OpArgNodes. Source argument arrays remain
non-null even when empty. Ordinary general-identifier applications also retain
the source operator ownership; fixity applications use their separate path.

Original Test206 and Test209 pass unchanged. External manual observers agree on
151 and 82 graph rows respectively, including concrete node kinds, checked levels,
operator/formal names and arities, observed reference sharing and selector owners.
These observations do not compare every syntax position, allocation UID or mutable
collection, and earn no original-method credit. This construction path does not
complete Java's invalid-selector diagnostics, DEF handling or general graph
allocation order. Those remain explicit implementation work.

## Canonical semantic graph traversal

`sany_graph.go` translates the actual SANY classes' walkGraph edges and visitor
callbacks. Use semantic UIDs for the visited table. Context and Subst callbacks
carry their actual objects without allocating semantic identities. AtNode neither
registers itself nor follows EXCEPT references; repeated occurrences receive
repeated callbacks. Module traversal follows Context and topLevelVec; LET follows
its Context rather than the incomplete definition array. Preserve proof edges,
numbered OpDef stepNode edges, and the source omission of declaration/parameter
edges from NewSymbNode, InstanceNode and ThmOrAssumpDefNode where applicable.

Context snapshots the Hashtable buckets and advances each enumerator entry before
callbacks. Keep the source inner-module warning and skip behavior, and its unusual
postVisit inside the enumeration loop. Indexed node-array loops read the current
field after callbacks, retaining replacement, shortening and null-failure behavior.
Null arrays guarded by Java remain optional; unguarded source reads still fail.

isDefinedWith compares actual references and retains Java's reachable-SubstInNode
restriction. substituteFor collects applications before changing operators and
excludes applications in the replacement graph according to that restriction.
It changes only the operator field, preserving selector ownership and cached data.

Manual production Java/Go observers agree on 6,238 traversal rows for original
Test206, Test207 and Test209; 4,440 reachability results; 1,836 substitution rows
across six fresh-graph definition pairs; and 243 callback-mutation rows across
shortened, reordered, extended and null operand arrays. These are external manual
observations, not additional original-method credit or complete malformed-node
coverage. The original SemanticCorpusTests method now supplies persistent source
assertions over actual canonical identities and levels. Its full 28-row matrix
and one source assumption are preserved. The driver also now publishes the last
external module as the ExternalModuleTable root after level checking, matching
Java even when prior diagnostic errors exist. Native callers without loader order
publish their declared root when it is generated.

## Unresolved call selectors retain source name syntax

Native CallExpr separates its flattened callee image from its actual source
selector. Java selectorToNode instead resolves against that selector directly.
Pass the source selector into callee lookup so SYMBOL_UNDEFINED accumulates
only name components, extends the range only at those names, and ignores argument
expressions when lookup fails. Navigation components do not become operator-name
text. Preserve the source UniqueString parameter alongside the exact formatted
message and location. Plain identifier diagnostics keep their own source selector.

The complete original TestSubexpressionSelectors class now checks the original
syntax/dependency and generation-only phases, retained errors on AbortException,
failure without internal diagnostics, and exact first code/message/location.
All three original methods pass in Java and Go. Sixteen external observations
also match complete diagnostic names, ranges and parameter types/values for
compound names, attached arguments, name components after navigation and symbolic
operators. These observations do not establish the entire selector state machine,
higher-order malformed selection behavior or source allocation order.

## Canonical child traversal and source locations

SemanticNode containment traversal is separate from walkGraph. Raw getChildren
retains null entries; getListOfChildren filters them into a fresh list, and
hasChildren uses that filtered list. ModuleNode caches its actual OpDefs followed
by its populated topLevel array. OpApplNode builds a fresh array of ranges then
operands. Preserve source null-array failure timing instead of lazily populating
fields during traversal.

walkChildren calls preVisit before snapshotting filtered children, defaults to
preempting children, ignores child postVisit return values and returns the root's
postVisit result. It has no UID suppression. pathTo checks current syntax locations
and formal declarations, stops searching once a path is found, and returns the
innermost node first. An unmatched path is a nonnull empty list. The runtime
SemanticPathTo helper likewise filters raw null entries before deciding whether
a node is a leaf.

Read current syntax filenames and all four coordinates directly for canonical
paths; do not substitute cached semantic locations or fill zero end coordinates.
Token constructors must retain their filename, as Java's SyntaxTreeNode does.
External production observers match 15,082 child, callback, path and mutation
rows over three original models. This covers observed containment behavior, not
all syntax allocations, malformed nodes or evaluator graph sharing.

## Runtime semantic context enumeration

SemanticContext uses linked Hashtable entries distinct from its newest-first
Pair history. Replacements change an entry's value without moving it; rehashing
relinks those same entries into the enlarged table. Symbol enumerators capture
the bucket array and retain their next-entry cursor. They are not snapshots of
all values and do not fail fast on mutation. HasMoreElements may set the cursor;
NextElement advances it before returning the current symbol. Exhaustion raises
NoSuchElementException with `Hashtable Enumerator` as its message.

Runtime module graph traversal enumerates keys using that cursor and resolves
the current context binding before traversing its node, matching Context.walkGraph.
A stored null symbol is still present according to occurSymbol; getSymbol returns
null for it. Null keys raise NullPointerException. Content remains a materialized
Go convenience view; mutation-sensitive callers use GetContextSymbolEnumeration.
External source observations match 170 enumeration/presence/exception rows and
25 traversal rows with colliding keys and callback replacements/rehashes. This
does not close runtime LET context reconstruction, missing graph node classes,
canonical metadata sharing or full runtime visitor parity.

## Complete scanner surface and actual argument previews

The JavaCC scanner generator includes all 74 jj_3 entry points and their complete
268-method dependency closure. Generate from the pinned Java parser rather than
reimplementing each preview's token choices. The runtime dispatcher retains all
entry points, their recursive scanners and source semantic predicates. Generated
source comments identify the parser independently of the local checkout path.

Optional OpArgs now invokes the actual scanner with budget 2. Restricted fairness
NoOpExtension uses call 50, ordinary GeneralId uses call 68, and BangExtension
uses call 72. Preserve their distinct saved-call indices and source order, even
though they preview the same argument production. Remove the hand-maintained
opening-parenthesis/first-token shortcut at these sites. Failed and successful
calls now enter the existing source-style saved-call/rescan machinery.

External observers agree on 4,440 rows across all 74 entry points, 12 inputs and
five budgets, comparing the boolean result, remaining budget, current cursor and
farthest cursor. These use default parser contexts, not exhaustive junction,
operator-stack or malformed-graph states. Existing original ParseErrorTests,
selector/incremental methods and model checks pass. Remaining production call
sites and general diagnostic rescan integration are still pending; generated
scanner coverage alone does not establish the whole JavaCC parser contract.

## Module body and definition decision order

Body uses actual source lookahead 1 with budget 1, then source decisions 2–5
with budget 2 in Java's order: definitions before RECURSIVE, INSTANCE, ASSUMPTION,
THEOREM, nested modules and USE/HIDE. Retain the USE ONLY semantic exclusion.
These calls save scanner state even when the branch fails or the production
ultimately succeeds; direct token eligibility checks omit that rescan history.

Definition heads use actual calls 8–11 with Integer.MAX_VALUE, distinguishing
function, postfix, infix and Identifier-LHS definitions. The Identifier branch
parses its LHS, consumes DEF, runs belchDEF and sets Expression-or-Instance
expectation before call 7 with budget 1. It selects Expression or Instantiation
at that point. Do not preview an INSTANCE body before parsing its LHS or retain
a separate manually estimated failed-lookahead span for these heads. LET and
proof callers preserve their actual enclosing expression boundaries through the
shared definition production.

External observations agree on 266 complete rows across 23 valid/malformed
modules, including concrete syntax kinds, images, four coordinates and complete
parse messages/residual stack text (outer log whitespace excluded). These cover
observed decisions, not every grammar/error-token expectation or all predicates.
Remaining call sites, full JavaCC bookkeeping and evaluator graph sharing are
still completion requirements.

## Declaration and fact previews retain actual saved calls

ConstantDeclarationItems uses source call 6 with budget 2 for optional operator
parameters. WITH substitutions use call 12 with budget 3 to preview comma, target
and arrow. Named ASSUMPTION and THEOREM use calls 13 and 21 with budget 2. After
Theorem's Assume-Prove preview, its Expression alternative uses call 23 with
budget 1. ASSUME/PROVE expression clauses use call 14 for the first clause and
15 after commas, both with budget 1; declarations and nested Assume-Prove keep
their direct source choices before those expression scans.

Remove manually estimated failed-input spans at these converted sites and let
the actual scanners save calls for rescanning. Distinct call indices matter even
where two sites preview the same production. Source entry points 1–23, 50, 68
and 72 now have connected production callers. Full remaining caller integration
and expected-token bookkeeping are still pending.

External production observations agree on all 580 rows across 47 valid/malformed
modules, retaining earlier body/definition cases plus declaration parameters,
WITH lists, named assumptions/theorems and multiple/labeled ASSUME/PROVE clauses.
Compare complete observed syntax kinds/images/coordinates and parse messages,
with only outer log whitespace excluded. This is bounded evidence, not full
JavaCC diagnostic, semantic-predicate or evaluator graph completion.

## Proof alternatives and commands use source previews

Proof first retains Java's direct BY choice, then uses terminal preview 25 with
budget 2, structured preview 26 with budget 1 and optional PROOF preview 24 with
budget 2. Do not eagerly consume PROOF before those decisions. USE/HIDE/BY uses
preview 29 with budget 1 for its optional fact list, and distinct expression
calls 27/28 for its first/subsequent facts and 30/31 after DEF. MODULE items retain
the source direct choice before expression lookahead.

Step assertions use call 32 with budget 1; repeated proof definitions use 33 with
budget 2. TAKE/PICK choices use 34/35 with Integer.MAX_VALUE, replacing separate
native identifier/bound scanners. AssertStep uses 36 with budget 1. Number uses
37 with budget 2 after consuming its first literal, replacing the manually
estimated decimal preview/error span. Preserve parser number/decimal flags.
Connected source entry points are now 1–37, 50, 68 and 72; later expression/selector
call sites and complete JavaCC diagnostic bookkeeping remain pending.

External observations match 1,301 complete rows across 83 cases, including earlier
module/definition/declaration/fact cases and valid/malformed proof and number
forms. Compare syntax kinds/images/four coordinates and complete parse messages,
with outer log whitespace excluded. This does not establish every proof state,
node allocation, expected-token set or evaluator graph-sharing requirement.


## Expression forms retain generated preview order

OpOrExpr uses source calls 38 (budget 2) and 39 (Integer.MAX_VALUE), each followed
by the source junction-column predicate, then 40 (budget 1). SomeQuant uses 41
(Integer.MAX_VALUE). BraceCases runs expression preview 45 (budget 1) before the
function-head predicate; its other alternatives use 42/43 (Integer.MAX_VALUE)
and 44 (budget 1). Eagerly running the function-head predicate changes source
scan order and is not equivalent to this decision sequence.

SBracketCases preserves function-head selection, then record preview 46,
keyword-field reclassification, record-set preview 47 and expression preview 48.
Calls 46/47 use Integer.MAX_VALUE and 48 uses budget 1. TupleOrAction's optional
expression uses 49 (budget 1). These sites save actual generated calls rather
than retaining separate native estimates of failed-input spans. Source production
entry points 1–50, 68 and 72 are now connected; remaining expression/selector
callers and full diagnostic bookkeeping remain pending.

External production observations agree on 1,828 rows across 114 valid/malformed
cases, comparing parse results, syntax kinds/images/four coordinates and complete
parse messages with outer log whitespace excluded. They do not establish every
junction/operator-stack predicate context, expected-token alternative, allocation
identity or evaluator graph-sharing requirement.


## Fairness, junction and operand previews preserve caller identity

FairnessExpr uses generated call 51 with budget 2 after ReducedExpression.
JunctionList starts its indentation context before calls 52/53, selecting
DisjList before ConjList with Integer.MAX_VALUE. Failed previews save their actual
source calls instead of separate native input-span estimates.

Initial Expression uses prefix/open-expression calls 54/55; ExtendableExpr's
infix right operand uses distinct calls 62/63. All four use Integer.MAX_VALUE
before the junction-column predicate. The shared Go operand helper receives
these caller indices explicitly. Reusing 54/55 for right operands would preserve
some branch results but lose Java's distinct saved-call/rescan identity.
ExtendableExpr's initial operand uses junction call 56 (Integer.MAX_VALUE)
before PreInEmptyTop, then primitive call 57 (budget 1) after the direct
parenthesized-form alternatives.

Connected source entry points are 1–57, 62, 63, 68 and 72. Remaining extension
and selector decisions and full JavaCC bookkeeping still require integration.
External observations match 2,661 rows across 157 cases, comparing actual parse
results, syntax kinds/images/four coordinates and complete parse messages with
outer log whitespace excluded. This bounded evidence does not prove every stack
or indentation context, expected-token alternative, allocation identity or
canonical/evaluator-sharing requirement.


## Extension and continuation previews own expression boundaries

ExtendableExpr uses source loop preview 58 (budget 1), then calls 59/60/61
(Integer.MAX_VALUE) before the source column predicates for postfix, record-field
and function-argument alternatives. The generated loop predicate owns indentation
and eligibility; a separate native stop guard would skip saved source calls.
Optional continuation uses preview 66 (budget 1), then infix 64 or label 65
(Integer.MAX_VALUE), followed by their source column predicates. Infix right
operands retain distinct calls 62/63.

The record-field indentation failure and extension/continuation alternative
failures use source jj_consume_token(-1), contributing no explicit expected-token
entry. Pass an empty expected list and retain actual saved-call rescans. An
invented Identifier expectation changes Java's empty following-input diagnostic
for an outdented record field, even though both parsers reject the input.

Source entry points 1–66, 68 and 72 are connected. Remaining selector callers
and complete expected-token/bookkeeping fidelity still require work. External
production observations match 3,125 rows across 181 cases, comparing actual syntax
kinds/images/four coordinates and complete messages with outer log whitespace
excluded. They do not establish exhaustive predicate contexts or whole parser,
canonical graph or evaluator-sharing completion.


## Selector decisions complete the generated caller integration

PrimitiveExp runs source String/Number calls 69/70 (Integer.MAX_VALUE) before
the source column predicates. Identifier/proof-step, infix and postfix alternatives
retain direct source selection; the nonexpressive-prefix alternative uses call
67 (Integer.MAX_VALUE). BangExtension uses outer preview 73 (budget 1), then
identifier preview 71 (Integer.MAX_VALUE) followed by the source @ exclusion.
Direct operator choices precede optional OpArgs call 72 (budget 2). The other
branch selects direct OpArgs before structural preview 74 (budget 1); it does
not run the identifier/operator optional-argument preview.

All 74 generated entry points now have connected production callers. Complete
expected-token/bookkeeping fidelity and corpus syntax AST comparison remain
separate requirements. External production observations match 3,663 rows across
211 cases, comparing actual results, syntax kinds/images/four coordinates and
complete messages with outer log whitespace excluded. These observations do not
establish exhaustive predicate states, allocation sharing or whole parser and
canonical/evaluator parity, and add no original test-method credit.


## Direct-choice expectations and saved-call lifetime

The existing generator now also translates JavaCC's ten direct-choice mask arrays
for all 130 sites. Runtime sites record their consumed-token generation; unrecorded
sites are inactive even at generation zero. Source sites 120 and 123–128 are
connected; other grammar sites and remaining native failure-span estimates are
still pending. Direct expected tokens are emitted in ascending token-kind order
before saved-call rescan entries. The rescan accumulator starts with those direct
entries so Java's duplicate checks apply across both sources of expectations.

Successful consumption advances the cleanup clock. Every 101 tokens clear each
saved call's token reference only when its generation is strictly less than the
current generation. Equal-generation calls retain their token reference. This
ports the generated parser's reference lifetime without changing scan budgets
or expiration behavior.

External source observers match 636 mask, stale-generation, cleanup-clock and
saved-call retention rows, including all 130 individual sites and five module
sizes. Existing production observations retain all 3,663 matching rows across
211 cases. These checks do not prove all grammar-site wiring, combined-site
expectations, malformed rescan contexts, allocation identity or whole parser and
evaluator fidelity. No original test-method credit follows from these probes.


## Module and declaration choice recording follows nested source order

CompilationUnit records site 4 when Prelude is absent. Prelude consumes only the
source identifier/number sequence and records repetition exit 5. After successful
Module, CompilationUnit resets the token manager to DEFAULT; failure paths retain
their source lexical state. BeginModule uses site 7. Extends records comma-loop
exit 8 or absent-EXTENDS site 9, at the actual consumed-token generation.

Body retains nested Java selection: record 10 before definition preview, then
11 before instance/assumption/theorem previews when RECURSIVE is absent, then 12
before the USE/HIDE predicate when no nested module is selected. Declaration
loops record sites 13–16; constant operator-declaration failures use sites 17/18.
The shared formal-declaration helper retains its separate pending sites.

Connected direct-choice sites are 4, 5, 7–18, 120 and 123–128. Source prelude's
unreachable inner switch site 6 and other grammar choices remain pending.
External production observations agree on 3,801 rows across 229 cases, including
raw prelude/module inputs, and 641 bookkeeping/lexical-reset rows. Evidence is
bounded and does not establish whole grammar expectation or parser/AST parity.


## Definition and formal declaration expectation sites remain distinct

Definition records source sites 19–22 at optional LOCAL, function-bound comma
exit, failed Expression/INSTANCE choice and prefix fallback failure. Identifier
Tuples record comma exit 23 or absent first identifier 24. IdentLHS uses first
formal-choice site 25 and subsequent-choice site 27, comma exit 26 and optional
argument-list absence 28. The shared helper receives the caller's actual site;
reusing one formal-choice site loses Java's distinct generation stamps.

IdentDecl records comma exit 29 and absent argument list 30. Formal operator
failures use 31/32 while constant operator failures retain 17/18. Instance,
Instantiation, Substitution and Assumption use sites 33–37 at their source
choices. Converted failures contribute actual generated mask expectations.
Other helper callers retain their pending source-site integration.

Connected direct-choice sites are 4, 5, 7–37, 120 and 123–128. External production
observations match 4,142 rows across 251 cases. Complete expected-token sets,
remaining grammar choices, native error-span estimates and whole parser/AST and
evaluator fidelity remain requirements beyond this bounded evidence.


## ASSUME/PROVE and NEW retain distinct expectation choices

AssumeProve records source label absence 38, ASSUME failure 39, first/subsequent
expression-choice sites 40/42, comma exit 41 and PROVE failure 43. Nested
AssumeProve and new-symbol declarations retain their direct choices before
expression preview and its direct-choice recording.

NewSymb records source sites 44–50 for constant keyword/formal failures, absent
constant domain, optional NEW before variable/state/action/temporal declarations,
type keyword failure and typed formal failure. Typed declarations use site 50
in the shared formal helper, separate from definition sites 25/27. MaybeBound's
absence uses 51, Theorem keyword failure 52 and terminal Proof choices 53/54.
Converted failures retain actual source mask and saved-rescan expectations.

Connected direct-choice sites are 4, 5, 7–54, 120 and 123–128. External production
observations match 4,559 rows across 273 cases; remaining grammar sites and full
expected-token, parser/AST and canonical/evaluator fidelity remain requirements.


## Fact commands and proof steps retain direct-choice identity

UseOrHideOrBy records source sites 55–65 for optional PROOF/ONLY, command keyword
failure, first/subsequent fact expressions and comma exits, first/subsequent DEF
expressions and comma exits, and absent DEF. The shared item helper receives
its direct-choice site and scanner index explicitly: pairs 59/27, 61/28, 62/30
and 64/31. Recording only scanner decisions loses the direct MODULE expectation.

StepStartToken failure uses 66, Step's assertion alternative 67 and absent DEFINE
68. TAKE bound/identifier exits use 69/70 and failure 71; WITNESS exit 72; PICK
identifier/bound exits 73/74 and failure 75; absent SUFFICES 76. Converted failure
branches retain source mask/rescan expectations without native input estimates.

Connected direct-choice sites are 4, 5, 7–76, 120 and 123–128. External production
observations match 5,001 rows across 295 cases. Remaining grammar sites, native
error-span estimates and complete parser/AST and canonical/evaluator fidelity
remain requirements beyond these bounded observations.


## Expression and set choices retain source expectation boundaries

Parenthesized/OpenExpression failures use sites 77/78 and reject token kinds
outside their source productions. OpArgs comma exit uses 79, operator-reference
failure 80, ordinary quantifier choices 81–84, temporal quantifier choices 85/86
and QuantBound choices 87/88. Converted failures retain generated expectations
and actual saved-call rescans instead of native token estimates.

BraceCases uses bound-intro failure 89, membership enumeration exit 90 or absent
continuation 92, simple enumeration exit 93, comprehension bound exits 94/95,
generic enumeration exit 96 or absent continuation 98. Shared loops receive
the caller's site explicitly. Absent continuation and a comma-loop exit are
separate source decisions, even when they occur at the same consumed generation.
Redundant source inner-switch failures 91/97 and remaining grammar sites remain
pending.

Connected direct-choice sites are 4, 5, 7–90, 92–96, 98, 120 and 123–128. External
production observations match 5,424 rows across 317 cases. Remaining expected-token
bookkeeping and full parser/AST and canonical/evaluator fidelity remain separate
requirements beyond this bounded evidence.


## Brackets and control forms preserve direct-choice sites

SBracketCases records function-bound exit 99, distinct identifier/keyword record
field exits 100/101, record-set exit 102, function-argument exit 103, EXCEPT exit
104 and continuation failure 105. Separate source record alternatives precede
field reclassification; the shared construction helper receives each actual site.
ExceptSpec/component choices use 106–108; tuples use 109/110; restricted
NoOpExtension and ReducedExpression use 111/112; fairness uses 113; LET definition
choice and repetition exit use 114/115; junction item failure 116; CHOOSE intro
failure 117; lambda comma exit 118.

Connected direct-choice sites are 4, 5, 7–90, 92–96, 98–118, 120 and 123–128.
Remaining sites are 0–3, 6, 91, 97, 119, 121, 122 and 129, including redundant
source switch failures. Full expected-token bookkeeping, native error-span
estimates and parser/AST fidelity remain separate requirements. External
production observations match 5,823 rows across 341 cases, providing bounded
syntax/message evidence rather than whole parser or evaluator completion.


## Operator-token choices complete the direct-choice site translation

PrefixOpToken, NEPrefixOpToken, InfixOpToken and PostfixOpToken accept the exact
token alternatives from generated source masks 0–3. Their callers validate the
source token production before constructing syntax nodes. Unary minus retains
the Infix Op production frame; InfixLHS calls the bare token production instead.
OpenStart uses site 129, including its source failure path. Prelude retains inner
failure 6; BraceCases keeps guarded switch failures 91/97; expression prefix
choices retain distinct initial/right-operand failures 119/121; proof lexemes
retain guarded failure 122. All 130 direct-choice sites now have translated paths.

External observations of the five token productions agree with Java for every
295 token kind: 1,475 rows compare acceptance, consumed-token count and expected
sequences. Existing 5,823 production rows over 341 cases still agree. These
observations do not establish exhaustive malformed rescan behavior, allocation
identity, native error-span fidelity or complete parser/AST and evaluator parity.


## Generated expected entries replace native span estimation

The obsolete startsExpressionLookahead branch existed only in operator-definition
helpers whose production callers select postfix, infix or prefix LHS. Identifier
LHS already follows Java's separate lookahead-7 Expression/INSTANCE branch.
Remove the duplicate native preview, its token-span map and the short-message
fallback. Short-message length now derives solely from the generated direct-mask
and saved-rescan entries. Junctions' failed choice calls the equivalent of
consume(-1), without adding invented AND/OR singleton expectations. Formal
declaration parsing requires the actual caller's site, removing its fallback.

Generated exceptions retain the consumed current token and ordered expected
sequences. Before first consumption, they retain a dummy token linked to the
input, matching JavaCC initialization. Ordinary message exceptions retain no
expected-token metadata. An external observer reads Java's actual jj_expentries
and Go's exception entries on the same 341 production cases. All entries agree,
including 14,741 sequences over 117 nonempty results; the 5,823 syntax/message
rows also remain identical. This is bounded evidence, not exhaustive parser,
syntax allocation, constructor API or evaluator parity.


## ParseException formatters retain their distinct source behavior

Generated parser failures construct an exception retaining currentToken and
expectedTokenSequences before formatting its diagnostic. ParseException's
getMessage is represented by Error: it escapes following token images, includes
all expected alternatives in order, retains the source ellipses and indentation,
and distinguishes one expected sequence from multiple or zero sequences.
Platform line separation follows the host OS. ParseExceptionExtended's short
formatter instead prints following images literally and escapes only the prior
token. Ordinary message exceptions return their supplied message directly.

External Java/Go observations cover all 295 singleton token alternatives,
empty expectations, EOF, multi-token/multiple alternatives and eight prior
images including quotes, slashes, controls, non-ASCII and a supplementary Unicode
character. All 2,400 full/short formatter observations agree on Linux. Existing
341 expected-entry and 5,823 syntax/message production rows remain identical.
Windows newline behavior, full constructor/token-reference identity and exhaustive
malformed rescan/AST contexts are not established by these observations.


## EOF is consumed and linked like any other JavaCC token

JavaCC jj_consume_token increments jj_gen and its cleanup clock on successful
EOF consumption. Go advance now does the same and returns the actual consumed
token. tokenAt requests a fresh token past EOF rather than returning the final
cached pointer; the token manager supplies distinct EOF objects with source
coordinates and links. The pretokenized adapter creates equivalent fresh EOF
objects, preserving EOF coordinates and linking the chain. EndModule uses
consumeParseToken; unused native consuming helpers are removed.

External observations of three input streams consume 110 tokens each and compare
kind, all four coordinates, generation, cleanup count, prior-token link and
distinct next-token identities. Java agrees with both Go paths on all 330 rows
per path. Existing 641 direct-choice/saved-call rows, 341 expected-entry rows and
5,823 syntax/message rows remain equal. These observations cover the reported
EOF state behavior, not full token constructor/reference, syntax AST or evaluator
parity.


## Saved lookaheads retain the actual consumed-token reference

Java jj_save stores token, the current consumed token, rather than token.next.
Go scanLookahead now saves that same reference. A parser owns one stable initial
dummy token, linked when the first input token is obtained; generated exceptions
and initial lookaheads share it. Rescans locate the saved consumed token and
begin scanning at its successor. This replaces the previous next-token reference
and compensating starting offset without changing generated scanner decisions.

External observations compare actual Java saved-call references for all 74 entry
points with three budgets, three consumed positions and three input streams:
1,998 rows agree on result, generation, saved generation, shared current-token
reference, unchanged cursor and successor presence. Another 18 observations
compare repeated generated exceptions' shared current token, cursor, successor,
full message and short message. Existing 641 saved-call/direct-choice rows, 341
expected-entry rows and 5,823 syntax/message rows also agree. This establishes
these reference paths within the observed contexts, not full ReInit/constructor,
syntax allocation, semantic graph or evaluator parity.


## Exception generation consumes failed-kind state once

Go generateParseException now follows Java's jj_kind path: union the recorded
failed kind with active direct-choice masks, clear a nonnegative failed kind,
order singleton alternatives by token kind, then rescan saved calls. Remove the
native API for injecting arbitrary singleton or longer expected sequences. The
generated constructor retains the current token and resulting sequences without
formatting or fetching the next token. Actual consume-failure paths obtain the
next token first and record their one expected kind (or -1 for a failed choice),
matching jj_consume_token. Short-message formatting remains the outer reporting
operation.

External observations exercise seven active-site groups, every failed kind
including -1, and two consecutive generations: all 4,144 rows match Java on
ordered entries, kind clearing, current-token identity and an unfetched successor.
Groups include no sites, one site, overlapping declaration/set/expression masks
and all 130 sites together. Existing 341 expected-entry, 5,823 syntax/message and
641 saved-call/direct-choice state rows remain equal. These observations do not
establish full ReInit/constructor, syntax AST, canonical/evaluator fidelity or
all malformed rescan contexts.


## Corpus expectations retain the source AST DSL

The test helper now translates AstNode's DSL kind table, ordered child/field-name
identity, aliasing, serialization and recursive equality behavior. Expected trees
are parsed by the source S-expression tokenization and recursive parser rules.
Retain source details: addField appends only to an already-present field list
while recording the child's field name; equality checks expected named edges;
tokenization does not flush an identifier at EOF. Header/separator matching
preserves complete TLA+ input bytes, multiline names and ERROR/SKIP attributes.
Kind usage is marked by fromString when expected DSL trees are loaded, matching
the Java static usage set. PlusCal and FAIR are excluded from the TLA+ usage check.

All 355 unchanged corpus metadata and expected-tree serialization rows agree
with Java, plus the zero-unused row. The original DSL-kind usage test now retains
all 355 parameter contexts. The full SANY-to-DSL translator, known-failure
runner and testAll AST equality assertions are now installed, with all 355
actual translated-output rows matching Java. Both original methods are port complete. These helper observations do not establish every malformed DSL,
constructor, Unicode-version edge or production AST fidelity.


## Syntax corpus translator helper paths

TlaPlusParserOutputTranslator's reparser now retains the source cursor, shared
lookahead node slice, merge, end check, previous/advance, kind-aware match and
strict consume behavior. Consume failures retain source expected/actual kind
names and offsets, distinct from structural assertion failures. Identifier
conversions retain the source Boolean/set/number-set/@ cases. Prefix/infix/postfix
conversion tables are mechanically copied from Java and use the production
operator synonym resolver. Identifier-list and tuple helper structure is ported.

External calls to the actual private Java helpers match 552 operator-literal
observations across all three fixities, including aliases and unsupported
conversion errors. Another 470 rows match all 446 kind-name mappings, twelve
identifier images, and twelve valid/malformed list/tuple conversions and offsets.
These observations do not establish all string-reference identity contexts or
exhaustive production parser parity. The recursive translate/flatTranslate
switches and expression-dependent bound/use helpers are now translated faithfully,
including source fallthrough and lookahead failure behavior. The corpus runner
preserves checked translation errors separately from structural assertion failures;
known failures invert acceptance and do not compare ASTs, as in Java. All 355
actual outputs agree, and the original Java class passes 710 test contexts.
No permissive fallback or placeholder AST translation is installed.


## Comment scanning preserves JavaCC segments and EOF boundaries

The token manager translates SPEC's comment-opening MORE rules, IN_COMMENT and
EMBEDDED's SPECIAL_TOKEN delimiters, and IN_EOL_COMMENT's newline special token.
Each special completes the accumulated image and starts a new segment. Actual
source kind values 29 through 33 replace synthetic grouped-comment kinds.
A following token points to the last special token; backward Special links and
forward Next links preserve the source chain. Syntax-node pre-comments reverse
the backward chain into chronological images. A fresh getNextToken call starts
with an empty local special chain, as Java does.

Source bracketCount is class-wide and survives construction and state switching;
the Go counter uses signed atomic updates to retain that shared state without a
Go data race. Embedded close transitions back only when the decremented count
is zero. At EOF, unfinished MORE input throws TokenMgrError and retains lexical
state. EOF immediately after a completed special token returns EOF, including
inside an unfinished outer comment. Do not repair this source behavior during
the faithful port. The parser's catch emits Java's multiline open-comment
summary using actual token-begin and last-character lines; direct tokenizer
errors retain TokenMgrError's original text.

All 200 bounded stream rows across DEFAULT, SPEC, IN_COMMENT, EMBEDDED and
IN_EOL_COMMENT agree on token/special kinds, images, coordinates, final state,
shared nesting count and failure messages. Another 28 rows agree on sequential
state and full parser output, including shared-counter effects. All 355 corpus
actual outputs remain equal after the change. These are bounded comparisons,
not exhaustive lexer, character-stream or concurrent-interleaving verification.


## PRAGMA longest match, junction prefixes and state validation

PRAGMA considers NUMBER, IDENTIFIER, BAND/BOR, module-header and anonymous
one-character SKIP candidates together. Greater matched length wins; ties use
source numeric kind priority. The SKIP kind 22 wins over a single-character
IDENTIFIER kind 289, while NUMBER kind 20 wins its single-character tie. This
retains digit-leading identifiers as whole tokens and recognizes junction tokens
in PRAGMA rather than discarding their suffix. The source header switches to SPEC.
SPEC whitespace remains exactly space, tab, CR and LF.

BAND/BOR use the grammar's CASE1b/c, CASE2b/c and CASE6b/c prefixes. They contain
ASCII letters/digits only; they exclude underscore, a lone W/S and WF/SF starts.
They must not use the broader general-identifier scanner. Invalid prefixes
therefore produce ordinary identifier/number/punctuation tokens according to
normal longest matching, including valid suffix junctions after an underscore.
All 1,080 bounded DEFAULT/PRAGMA/SPEC token-stream rows agree with Java on kinds,
images, coordinates, final state and lexical errors.

SwitchTo accepts exactly source states 0 through 5. Invalid states throw typed
TokenMgrError with original text and errorCode 2, preserving the old state.
The nil receiver retains NullPointerException. Lexical failures retain default
reason 0. All 66 valid/invalid transition observations agree, including signed
32-bit endpoint values. These observations do not establish full ReInit,
constructor, Unicode decoding or character-stream parity.


## Character-stream prerequisite for generated token scanning

`sany_char_stream.go` translates SimpleCharStream's actual circular buffering,
2048-unit expansion, refill/reuse choices, token beginning, saved line/column
arrays, backup, image/suffix extraction, Done, ReInit and adjustBeginLineColumn.
Characters are UTF-16 units; CR and LF remain separate reads, and rereading a
backed-up unit does not update position metadata again. Source integer fields
and stored positions use signed 32-bit values, retaining overflow. Reader EOF
closes the reader and follows the source buffer-position correction; other
reader errors follow that correction without treating them as EOF. String
reader adaptation supplies decoded UTF-16 units; byte decoding is separate.

External comparisons of 75 stream/size/chunk configurations produce 10,905
identical state/read/image/suffix/reinitialization/adjustment rows. Another 547
rows agree for multiple long tokens, buffer reuse, Done and signed endpoint
line/column values with three tab sizes. These compare actual Java fields and
methods, not a replacement stream model. They do not establish arbitrary reader
I/O failures, malformed internal states, allocation-error messages or every
encoding constructor.

The token manager now uses this stream directly. `sany_scanner_generated.go`
translates the source DFA/NFA methods, Unicode bit vectors, next-state tables,
nullable literal images and token/skip/special/more masks. Signed 32-bit scanner
fields and explicit pre/postincrement helpers retain Java evaluation order.
`sany_token_manager.go` follows the generated dispatch, backup, MORE accumulation,
special-token linking and lexical-error probe. Candidate selection is removed.
Source TokenMgrError text is retained; the Go diagnostic uses E1200 for lexical
failures rather than the previous heuristic categories.

All 10,392 character/operator rows across DEFAULT/PRAGMA/SPEC agree with Java,
including the 13 previously failing malformed strings and incomplete prefixes.
Another 16,247 observations compare scanner counters, raw UTF-16 image buffers
and hashes of both state arrays. All 200 comment-stream rows and 28 sequential
lexer/parser-output rows still agree. All 355 corpus translation results match
Java without changing fixtures or AST assertions. Another 300 rows agree for
ReInit across all six starting states, valid/invalid target states and round
values around the reset boundary, including full state arrays. All 5,823 parser
tree/error rows remain exact. General byte-reader/error constructors still need
source reconciliation. These bounded receipts do not establish full parser parity.


## SANY byte-source UTF-8 decoding

ParseUnit constructs TLAplusParser through its SanyOutput/InputStream overload,
which explicitly selects UTF-8. Direct conversion of file bytes to a Go string
previously let rune decoding replace each malformed byte separately, whereas
Java groups some malformed prefixes into one replacement. This shifted token
end columns, string values and following locations.

The loader and debugger dependency loader now call `tlc.DecodeUTF8Replacing`,
which exposes the existing Java-compatible replacement codec. Decoded strings
then feed the UTF-16 reader and generated scanner. The explicit UTF-8 source
path is independent of the configured default charset. Text-based parsing APIs
continue to accept already-decoded strings.

A bounded matrix of 6,166 byte inputs covers all single-byte values, selected
leading bytes with every second byte, truncated/surrogate/overlong/range cases,
BOM and a supplementary character in strings and expression positions. All
49,430 actual Java byte-stream token/decoded-text rows and 6,166 complete parser
outputs agree; seven direct file-loader cases pass. This does not establish
arbitrary stream chunking, I/O failures or every encoding constructor. No new
permanent tests or original test-method credit.


## SANY monolith fallback uses the ported extractor

`sanyLoader.loadMonolithModule` now calls `tlc.MonolithModule`, matching
FileUtil.createNamedInputStream's fallback. It reads and closes the actual
NamedInputStream and passes its bytes, source-file path and logical basename to
the common loader. The common `loadSource` path explicitly identifies extracted
siblings for syntax rebasing and retains root-file parsing provenance. The
separate loader extraction loop and shared temporary-directory field are removed.

Monolith extraction reads and writes with the default charset before SANY parses
the resulting file as UTF-8. The prior byte-to-string extraction bypassed those
steps: 27 of 36 bounded US-ASCII cases differed in retained source text. All 108
observations now agree across UTF-8, US-ASCII and UTF-16 defaults, covering malformed
prefixes, non-ASCII/supplementary values, BOM, missing extraction and CR/LF/CRLF.
Comparisons include source text, parser success/output and stream names. Original
MonolithSpecExtractorTest's five methods and the Go extractor/model checks pass.
This does not establish all filesystem failures or every default charset.


## Retained generation scopes after expression failures

Java Generator's LET, operator/function body, quantifier, CHOOSE and bound-
expression paths pop their contexts after normal body generation. They do not
use finally to erase the scope when a body or domain throws. The Go port now
uses explicit successful-path cleanup for these scopes, label sets/formal
parameter groups and named-function stacks. Domain generation starts after
pushing the new empty context and before introducing quantified parameters.

LET raises curLevel only below the source limit, retains its INTERNAL_ERROR
AbortException otherwise, and preserves the actual unresolved-count array
boundary. Consequently ordinary nesting at level 100 still throws the same
array-bounds family as Java; the port does not replace that source behavior.
The source clamp after decrement is also retained. Graph construction on normal
completion still uses the retained LET context before popping it.

All 4,256 external observations agree across 23 cases, including depths 98–101,
bound-variable bodies/domains, named function/operator bodies and supplied level
boundaries. Comparisons include context binding order, level, exception family,
label stacks/parameter groups, function-stack size and diagnostic messages.
The prior 247 normal LET graph rows remain exact. These are bounded body/domain
failure receipts, not constructor-abort or complete runtime canonical-sharing
verification. No invented permanent tests or extra original-method credit.


## ASSUME/PROVE declaration-scope array and failure state

The expression generator retains Java's 100-entry `inScopeOfAPDecl` array and
signed 32-bit assume/prove depth. Generation increments depth before constructing
the AP node, initializes that depth's array slot before pushing an inner context,
and marks it after generating each NEW declaration. A thrown body/domain retains
its active depth and symbol contexts. Normal completion pops the inner context
and decrements depth; array entries remain as source-owned state.

`noLabelsAllowed` now scans actual declaration flags from depth 2 through the
active depth. It excludes outermost declarations and preserves source array-
bounds behavior. The previous single forbidden-label boolean and deferred AP
scope cleanup are removed. Source endpoint arithmetic is retained, including
increment from Integer.MAX_VALUE to Integer.MIN_VALUE before array access.

All 1,077 external state/context/label-restriction observations agree across 21
cases: normal and nested APs, NEW declarations, boxed AP, thrown LET bodies and
NEW domains, depths 98–101 and signed endpoints. These compare all array flags,
depth, current goal clause, context binding order and exception families. The
225 prior AP graph rows still agree; original selector/Test212/Test213 and SANY
checks pass. These bounded observations do not establish all generation callers,
constructor-abort paths or canonical/evaluator graph sharing.

## Ordinary operator label scope through construction

Java `processOperator` pops its formal context after body generation, then
constructs/registers the OpDef or completes the actual recursive declaration.
It assigns recursion fields before popping the label set and attaching it to
the definition. Go now returns the pending label-scope completion from body
generation and invokes it at that same successful construction boundary. It
no longer stores an early label-table copy on `SanyExprSource`. Native paths
whose canonical graph remains incomplete explicitly finish their successful
scope without manufacturing a body or definition.

Recursion-field assignment retains the Java array-bounds exception family for
negative/out-of-range LET levels when unresolved declarations require the array
lookup. A failure here preserves the registered definition and active label
scope; the returned definition and attached labels remain absent. Twenty exact
external observations cover levels -1, 0, 99, 100 and 101 with zero/positive
unresolved sums, actual source registration, recursion fields and label tables.
The previous normal LET graph and body/domain failure observations remain exact.
These supplied-state probes do not establish all constructor failures or runtime
canonical graph sharing. No permanent tests or inventory credit added.

## Recursive definition completion owns its counters

`endRecursiveDefinition` now follows actual Java `endOpDefNode`: set defined,
body and syntax first, then decrement the current unresolved count and sum if
the actual node is recursive. Preserve typed array-bound failures and the source
WrongInvocationException after a negative sum. The node and native binding view
retain their already-completed state when counting fails. Ordinary operator
labels remain active until this call returns normally.

Canonical LET/module callers no longer count again after completion; canonical
named functions complete at specification construction, before body generation,
rather than during native preregistration. Existing native paths without actual
canonical declarations retain separate bookkeeping, without creating substitute
nodes. Thirty exact external processOperator observations cover levels -1, 0,
99, 100 and 101, unresolved sums 0/1/2 and counts 0/2. Compare actual registration,
returned identity, node/body/labels, contexts, active label sets and retained
counters. Earlier 247 LET, 500 named-function and 4,256 body/domain rows remain
exact. Other constructor boundaries and canonical evaluator sharing remain work.

## Recursive completion with null syntax

`endOpDefNode` assigns the supplied syntax even when it is null. Java
`SemanticNode.getLocation` then returns Location.nullLoc. The Go completion
path now clears its tree interface, native position and cached location at that
same mutation boundary. Assigning a typed nil syntax pointer to the interface
alone left GetTreeNode observably non-null; retaining the declaration's location
also contradicted source behavior. Null declaration arguments throw the source
NullPointerException family before changing any counters.

Thirty-two external observations agree with actual endOpDefNode over null/non-
null nodes, syntax and bodies, recursive/nonrecursive definitions and unresolved
sums zero/one. Counter underflow still retains the already-written body and
syntax. Comparison covers exception family, counters, body identity, defined
state, tree nullness and location text; it does not establish every exception
message or other syntax mutation path.

## Remaining runtime LET adapter requirements

The retained canonical `sanySemLetInNode.context` is the source of traversal
bindings. Runtime LetInNode currently has only Lets, Bindings and Body, and
coverageLetDefinitions reconstructs a Hashtable from filtered user definitions.
Java walkGraph visits Context instead, specifically because opDefs is incomplete.
Module-instance definitions and imported theorem/assumption bindings must survive
that traversal. Preserve context Pair history, current key-to-Pair selection,
bucket/chain order, original module namespace and shared symbol identities.
Replaying filtered getLets or Pair insertion order is insufficient for contexts
whose duplicate operation rebuilt lookup bindings in a different order.

Runtime conversion therefore needs a canonical node adapter with an identity
cache before it can transfer the actual context. The bridge retains original
local AST Definition pointers and caches actual LET adapters; it still constructs
runtime OpDefs after converting bodies and separately reconstructs INSTANCE
exports. These are not canonical-context identity mappings. Canonical formal
adapters retain their actual SemanticNode base; other symbol facades, imported
ThmOrAssumpDef bodies and module-instance kind/parameters still require complete
integration with their canonical counterparts. Preserve cycles by publishing
adapter shells before traversing their children, and use the retained context
rather than a name-matching substitute. Complete context integration remains
unimplemented. The source completion fixes above do not earn runtime LET parity
credit.

## First canonical LET and formal adapters

The bridge caches each runtime LET by its actual canonical LetInNode pointer,
with an adapter shell published before child conversion. It passes original
local Definition pointers to the existing definition cache, instead of copying
those declarations on every traversal. A separate canonical formal cache reuses
the runtime declaration for each actual FormalParamNode across source bodies,
bound expressions and lifted selector formals. Debugger bridge copies retain
these existing mappings while new parser nodes remain independent.

This identity change exposed the prior bridge's reliance on reconstructing
captured formals: unchanged Test206 and Test209 initially failed. Definitions,
quantifiers, CHOOSE, named/anonymous functions, comprehensions and INSTANCE
parameters now share the actual formal mappings with selectors. The native
selector path also constructs temporary lexical LET wrappers with replacement
bodies. Those wrappers must not retain the original canonical LET graph pointer;
otherwise caching returns the wrong body. Their native context lowering remains
pending replacement by the full canonical graph adapter.

Seven bounded Java/Go observations agree for ordinary, recursive, function and
nested LET bodies: repeated node/body identity, original local declaration
ownership, distinct adapter nodes for distinct source LETs, formal reuse and
constant evaluation. All 40 focused original model/debugger methods and 19
coverage methods pass unchanged. These receipts do not complete canonical
Context transfer, semantic-base/UID sharing, imported theorem adapters, complete
OpDef shell publication or source substitution-array sharing.

## Formal adapters retain the actual semantic base

`SymbolNode.SemanticBase` identifies the semantic owner of an evaluator symbol
view. A canonical formal adapter retains the parser's actual base pointer,
without allocating another UID or copying indexed tool slots. Its formal-name,
arity and declaration facade remain available to current evaluator callers.
Standalone runtime FormalParam symbols allocate their own kind-11 base. Runtime
OpDef symbols retain their actual runtime definition base; this does not yet
replace that definition with the parser's canonical node. Generic lookup aliases
can still lack an actual semantic owner and cannot store indexed tool objects.

Symbol indexed-slot operations delegate to this owner. Tool lookup checks the
context first, then the active tool's indexed symbol object. Definition body
lookup checks the active indexed body slot, including substitution wrappers.
Generic definition/body cache fallbacks are removed. Config constants, constant
pre-evaluation and native/module overrides now write indexed slots with the
processor's tool ID; config-processing tools retain that ID. Canonical declared
constants retain the actual parser base, so their slots belong to the source
node. Base-less native aliases still use Data and definition-table compatibility
paths. Literal evaluation now uses Java's active indexed slot and worker muxing;
full lookup/evaluator parity is not claimed, including the missing theorem
initialization path exposed by original Test216.
The current declaration facade's copied location/tree fields and other node
adapters also remain separate from the canonical semantic-base integration.

Canonical paths no longer allocate a temporary runtime formal before replacing
it with the actual adapter. Selector parameter conversion likewise resolves
actual formals before constructing a native fallback list. Two Java/Go rows
verify retained base/UID identity, writes visible in both directions and exactly
one UID allocation for the following declaration. The expanded 192-case lookup
comparison matches actual SymbolNodeValueLookupProvider across formal/definition
symbols, empty/bound contexts, tool IDs -1/0/1/4, indexed symbol values and tuple,
numeral and string bodies. Numeral/string cases exposed the removed generic body
fallback. These bounded cases do not verify all symbol classes or the remaining
native alias paths. Existing original methods and coverage checks remain intact.

## Remaining indexed literal evaluation boundary

Java Tool.evalImpl handles NumeralKind, DecimalKind and StringKind together:
cast WorkerValue.mux of the expression's active indexed slot to Value. It can
return null for an unprocessed node and throws ClassCastException for a cached
object of the wrong type. Real-number rejection occurs earlier in source
SpecProcessor.processConstants, rather than in this evaluation branch.
Go EvalImpl now follows this indexed branch. The external comparison retains
45 cache cases, 12 worker selections, seven repeated initializations and five
rejected literal inputs; all 69 rows match actual Java. Constant processing
creates fresh StringValues and IntValues outside the 0..9 pool, while preserving
unrelated slots and source rejection timing. Constructors do not fill indexed
slots to preserve the old direct-value shortcut.

The bridge reuses actual generated string operands for records, record sets,
record selection and EXCEPT paths. Native selector views retain the actual scalar
body through source LAMBDA/Nop, LET and substitution wrappers, so constant
processing and selectors share its slots. Original Test219 exposed both old
reconstructions and now passes. This does not complete wrapper/Context adapters.
Original Test216 exposed missing module theorem initialization. Runtime modules
now retain the canonical theorem statement vector through shared TLC views, and
processConstants visits those statements after assumptions. Views retain the
actual semantic base, proof and suffices flag. Repeated EXTENDS paths share
views; INSTANCE imports definitions without adding statements. Plain named
theorems share their runtime definition body. ASSUME/PROVE statements retain
the actual source graph, without initializing their goal as a substitute for
the whole statement. Complete runtime ASSUME/PROVE definition-body adaptation
and ordered top-level graph transfer remain pending. Test216 passes unchanged.

Record constructors and record sets now read their field names by casting the
active indexed tool slot directly to StringValue. They do not mux worker values,
matching the distinct source branches. Record selection does mux its field
slot to Value and never evaluates an absent slot. It preserves detailed failure
messages and expression/context identity, and treats CounterExample as Java's
RecordValue subclass. RecordValue.select retains the source null-argument
failure and runtime exception metadata. The 102 bounded source observations
cover these cache branches and worker choices, with independently built graph
locations normalized; they do not establish complete evaluator exception or
malformed-graph parity.

Runtime OpDef adapters now borrow the actual parser-owned SemanticNodeBase.
Their symbols share the same pointer, so UID, syntax/location and indexed
tool slots have one owner. Standalone runtime constructors still allocate a
base. The bridge caches a definition shell by both source AST and canonical
OpDef identity before visiting its body; debugger clones retain those caches.
This does not change source Generator construction order. The 190 inspected
explicit definitions from five original fixtures retain these identities.
Canonical LET Context and imported-wrapper graph transfer remain pending;
filtered runtime LET arrays and coverage's reconstructed Hashtable remain
incomplete representations of source Context membership/history.

Context state transfer now has a source descriptor for retained Pair history
and actual Hashtable bucket chains. Runtime import builds those structures
without replaying insertions and retains mapped semantic node/Pair identity,
the source threshold and the runtime module-table owner. The 1,381 bounded
Java Content/lookup/null observations agree, including subsequent duplication
and mutation. This infrastructure is not yet wired into LET: complete symbol
and graph adapters remain required. The SANY test fidelity audit is now the
user's priority before further bridge integration.

Java Spec's toolId is static final and obtained once from semantic.FrontEnd's
allocator. Current specToolID uses sync.Once for the same ownership. The tool
ID is shared by subclasses; allocating a new ID for each ordinary Tool would
change source behavior.

## Nested INSTANCE prefix arity

Generator.selectorToNode retains opDefArityFound across resolved name components.
An imported module-instance node's arity includes earlier prefix parameters, so
its current argument list must match nodeArity minus opDefArityFound. Go's early
selector check formerly compared against the whole imported signature, rejecting
valid I(55)!Inst and I(55)!Inst2(...) references in original Test219. The check now
retains Java's counter and reports the remaining signature. Native metadata that
lacks the actual cumulative source signature retains its local-arity accounting;
this compatibility path does not complete canonical selector generation.
The whole original Test219 passes with its unchanged fixture and assertions.

## Native distributed reply and process lifetime

Distributed Go endpoints use net/rpc over TCP with explicit Go payload graphs.
The server codec uses the standard gob request/response header and body format;
it also records accepted requests until the response is flushed. Request
admission and the closed flag share a mutex, so no request can join the drain
once orderly shutdown has started. The worker and FP command owners close their
hosts gracefully: stop listening, finish accepted replies, then release socket
and outbound-client ownership. This is necessary because storage Exit wakes the
FP reporting loop before the RPC handler returns its response.

Explicit Close remains an immediate transport abort, including during a drain.
It must not acknowledge an unfinished insertion or hide its ambiguous outcome.
No automatic retry is performed at this boundary. Fingerprint storage lifetime,
worker exit latch, reporting/checkpoint bounds and coordinator batching remain
owned by the original TLC port. These Go transport lifecycle rules do not imply
Java RMI compatibility or a JVM runtime. Process/model coverage still needs the
remaining failure, checkpoint/recovery and extended-payload cases.

## Native finite constant-operator payload

Configured constant operators are finite `OpRcdValue` argument-row/result maps,
not evaluator closures. Java's `Value` base and its `Vect` containers declare
Serializable; the rows and results contain ordinary values. The Go graph now
represents these rows explicitly and allocates the operator before filling its
references, preserving shared values and cycles. Nil/empty arrays and rows stay
distinct through gob. Decoding uses the existing Go operator constructor to
retain application, printing and unsupported-operation behavior; it does not
convert the operator into a function. Invalid row references are rejected.

Java serializability declarations alone do not define required native payloads.
`OpLambdaValue` holds a non-transient `Context` that does not implement
Serializable; `MethodValue` holds reflection objects and method handles.
`Action` declares Serializable, but its ordinary `SemanticNode` predicate and
`Context` do not. These evaluator objects are not a reason to introduce JVM
emulation or Java serialization. Extended metadata and custom data still require
a source-behavior audit; supported transferable data must not be discarded.

## Distributed successor publication failure boundary

TLCServerThread inserts result fingerprints before accessing selected successor
partitions. The original implementation dereferences each visited vector and,
for each selected bit, reads the state and fingerprint, writes its trace record,
then enqueues it. Missing/nil selected data fails through the outer Throwable
handler; it is not skipped or treated as an empty result. The native port now
retains that ordering with typed null/array failures. Arrays with no selected
bits are not eagerly validated, matching the source dereferences.

Trace-write errors propagate from publication to Run, which records the failure,
finishes the queue, notifies the coordinator and returns. Previously publication
handled the error internally and returned to the run loop, allowing another
dequeue after failure. Neither publication failure path undoes FP insertion;
adding rollback or retry would change the original distributed algorithm.

The shared BitVectorIter also enforces the source null dereferences: Init(null)
fails before replacing iterator state, and Next fails on uninitialized words.
Worker contains-block answers pass through this same iterator. A null answer
therefore becomes a WorkerException with its current predecessor, null successor
and KeepCallStack, rather than a successful empty result. Native worker errors
carry the null-failure trait instead of reconstructing JVM exception objects.
FP TCP replies include word-array nullness separately from vector presence, so
gob's empty-slice encoding cannot turn an initialized empty answer into an
uninitialized one. Initialized empty vectors remain valid and yield no bits.

Distributed FP failover diagnostics go through ToolIO rather than bypassing it
with direct stdout writes. The first println call includes the warning and
failure detail separated by an embedded newline; the second is issued only
when reassignment reports no available server. Tool mode records those two
messages, and system mode honors its assigned output stream. Reassignment,
retry and final scalar/block/statistics answers retain the original behavior.

Worker completion/readiness and TLCApp option diagnostics also use ToolIO's
println boundary, honoring native CLI stream assignments and tool-mode capture.
The completion message precedes executor shutdown, keepalive cancellation,
unpublication and latch countdown, as in the source. These routing corrections
retain existing diagnostic text and option acceptance. Source System.err stack
traces still use stderr; they are not redirected into ToolIO's message buffer.

Fingerprint RPC failures use the shared native Go failure payload rather than
flattening errors to strings. Nullable messages, cause/suppressed-error graphs,
sharing and sender Go stacks survive the boundary. Fatal returned or panicked
storage failures receive the remote I/O category with the original failure as
cause; ordinary failures retain their existing catch traits. This is TLC
behavioral parity over Go transport, with no RMI or Java serialization support.

The assigned-block checkpoint check covers local FP storage and a registered
native TCP fingerprint store. Recovery reopens empty storage behind a new
listener/client, then restores the remote checkpoint file together with the
committed queue and trace. This proves the existing checkpoint/recovery methods
with registered endpoints, not fresh-process distributed CLI recovery. In the
pinned Java source, TLCServer.modelCheck calls recover before coordinator
publication and waitForFPSetManager; DistributedFPSetTLCServer constructs an
empty dynamic manager. Keep that startup limitation explicit rather than
claiming this boundary test proves a remote recovery startup flow.

Native error-trace model coverage uses real coordinator/worker processes and
optional standalone fingerprint-server processes on owned loopback listeners.
DieHard keeps all seven exact source state strings; TSnapShot keeps the source
zero-queue, FINISHED and BEHAVIOR requirements. No GENERAL is allowed across
any role, retaining the scope of Java's shared recorder. Native subprocesses
need no JVM exit/security interception; all are joined on success or failure.
The original DistributedTLCTestCase remains unconditionally disabled upstream,
so these are supplemental model checks rather than completion of its harness.

The local fingerprint invocation boundary checks fatal errors returned from
Go endpoints as well as panics. Both escape the manager's ordinary exception
catches. Distributed Close uses this same boundary, so a fatal local Exit does
not become a printed ordinary error. Remote fatal failures are already wrapped
as native I/O operation errors by the RPC handler, retaining the source remote
failover behavior without reconstructing JVM exception objects.

All three native RPC services use encodeDistributedRPCFailure. It wraps fatal
endpoint failures as remote I/O operation errors, preserving the original graph
as cause regardless of whether the endpoint returned the error or panicked.
Previously worker/coordinator services wrapped only panics. Ordinary exceptions
and the worker's own evaluation/OutOfMemory wrappers retain their existing
traits. Failed worker Exit does not unpublish its endpoint. The shared encoder
uses native Go payloads and stacks, without RMI or JVM exception reconstruction.

Worker keepalive uses PrintErrorThrowable for a failed coordinator lookup/status
call, matching TLCTimerTask.exitWorker's source overload. Nullable detail text
still supplies the normal message; Globals.Debug additionally prints the sender
failure stack through ToolIO. The native lifecycle check invokes the public
RunKeepAliveOnce boundary against real TCP calls, preserving the production
schedule and timeout. Computing/recent activity suppresses lookup; idle completion
or loss shuts down the executor, worker callback, completion latch and timer.

In-flight fingerprint connection-loss checks hold an actual accepted PutBlock
or ContainsBlock handler, close its native host, and require the manager to
reassign to another TCP store before releasing the old handler. Sequential and
concurrent calls retain input partition order and alias both registrations to
the survivor wrapper. Contains does not copy lost fingerprints or insert new
ones; Put retries against surviving storage. No transport-level retry or new
batch protocol is introduced. Host.Close owns sockets only, so the paused local
storage handler remains alive and is explicitly released/joined by the test.
This boundary does not establish fingerprint-process crash/recovery behavior.

The coordinator main catch uses PrintErrorThrowable for stack-overflow and
out-of-memory categories, preserving the source debug stack policy. The shared
printThrowable writes each stack line using ToolIOPrintln, matching Java
Throwable.printStackTrace's println calls. A single raw ToolIOPrint previously
left the whole stack as an unfinished buffer prefix, which could also trigger
the retained source getAllMessages arraycopy bounds failure. Normal stream
bytes retain their line structure; buffered mode now receives completed lines.
GC, close(false), executor shutdown and management cleanup remain ordered as
in the source main catch/finally.

ModelValue byte data belongs to the native invocation graph. The payload carries
ByteArrays once and DataBytes as a one-based reference (zero is a nil byte slice).
The encoder identifies nonempty buffers by their starting address and length;
equal contents in different arrays do not merge. This represents shared source
byte-array objects, not arbitrary overlapping Go slice views. Empty buffers get
explicit nodes, and decoding allocates an initialized empty slice even if gob
represents its node contents as nil. The decoder copies each buffer once for
receiver ownership and reuses it for every model-data reference, validating
bounds. Sharing therefore survives across states and result partitions without
aliasing sender storage. This changes the native DTO layout; communicating roles
use the same build. No Java serialization or opaque custom-data codec is added.

Partitioned checkpoint/recovery coverage uses two independent native FP hosts
with private metadata directories and a common checkpoint name. The initial FP
61 occupies partition 1; successor FP 72 occupies partition 0. The checkpoint
waits for assigned successor publication and restores the queue/trace plus both
partition files into empty new tables without swapping or merging partitions.
Local and single-host cases retain their original boundary assertions.

The fresh-process remote-FP startup limitation is confirmed by the pinned source:
FPSetManager's default constructor starts with an empty ArrayList;
DistributedFPSetTLCServer.getFPSetManagerImpl constructs that dynamic manager;
TLCServer.modelCheck calls recover before publishing itself or awaiting FP
registration. FPSetManager.recover invokes chkptInner, which has no endpoint to
call in this empty manager. Go mirrors that ordering. The registered-endpoint
recovery checks do not claim the CLI restores those remote stores. Supporting
that startup flow would require an explicit algorithm enhancement beyond parity.

FPSetManager.Checkpoint.run catches I/O failures for both named checkpoint and
recovery calls, warns and continues synchronously to the next distinct store.
It does not mark the failed registration unavailable or reassign its partition.
The native TCP boundary now verifies disconnection after a begin/commit/recover
handler has been accepted but before storage executes. Manager continuation and
coordinator queue/trace commit complete before that handler is released. Closing
transport does not cancel the accepted storage operation. A failed begin has no
follow-up commit; an ambiguously completed commit may still commit at the old
store after the caller proceeds. These source semantics provide neither an
atomic cross-store checkpoint nor process-crash recovery.

ModelValue's non-transient data Object accepts ordinary boxed scalar objects in
the source. Native payload model data now additionally supports Go int8, int16
and float32, retaining their types. Decode rejects out-of-range narrow integers.
Floating-point data uses DataFloatBits instead of a float64 DTO field: gob omits
a zero-valued struct field, which includes negative zero when compared to zero.
Integer bit patterns retain both float widths exactly, including signed zero,
subnormals, infinities and NaN; float32 decode rejects bits beyond uint32. This
changes the native payload layout, so communicating roles use the same build.
It adds no Java serialization and no arbitrary opaque-data registration.

DistributedLocationError retains a malformed native coordinator location and
its parsing/validation cause. Discovery previously returned ordinary errors;
the keepalive task consequently missed its malformed-location catch and would
terminate the timer. The task now recognizes this Go category alongside the
existing source exception supplied by local adapters, logs at the finest-level
boundary and continues without worker exit or timer cancellation. Validation
rules are unchanged, and no Java transport type is used for native discovery.

The worker file resolver API is DistributedFilenameToStreamResolver, constructed
with NewDistributedFilenameToStreamResolver. It uses DistributedFileServer for
local or native TCP coordinator calls. The former RMI-named Go API is removed
from worker commands, root bootstrap, tests and process helpers. The upstream
RMIFilenameToStreamResolver is retained only as a behavior reference: basename
cache keys, actual file-existence checks before reuse, refetch after deletion,
ignored library paths/isModule flag, binary/empty writes and private temporary
directories remain intact. This is a Go API rename, not a wire-format change.

Native worker graph encode/decode validation failures belong to the remote I/O
boundary, retaining the codec cause with no retry. Request encoding fails before
dispatch; malformed request graphs fail before endpoint invocation; result
encoding fails after evaluation. The coordinator's inner remote catch requeues
the assigned block and removes that worker, rather than recording an application
GENERAL and terminating the model. Source exceptions raised during value
materialization retain their application category. In particular, LazyValue's
write contract raises the exact runtime diagnostic "Error(TLC): Attempted to
serialize lazy value." for null/undefined cached values, with source context
when present. Native payload rejection now retains that contract too.

Manager snapshots contain source FPSets registration wrappers and endpoint
references; receiving those references must not depend on remote aliveness.
GetFPSetManager now decodes native references without dialing stores and commits
their ownership only after complete graph validation. NetworkFingerprintEndpoint
connects on its first operation. Concurrent initial dials publish one client and
discard extras; closure disables unused references and discards successful dials
that finish after closure. Established failed clients are not redialed and no
operation is replayed. This moves unavailable-store detection into the existing
manager operation/failover path rather than incorrectly failing worker bootstrap.
Direct DialFingerprintEndpoint still explicitly connects at the caller's request.

Fingerprint registration follows the same endpoint-reference boundary:
DynamicFPSetManager.register only checks capacity and appends an FPSets wrapper;
DistributedFPSetTLCServer.registerFPSet then counts down and prints acceptance.
Native registerFP no longer dials before these calls. It validates complete
address/object references, registers ownership of their CloseConnection adapter
and passes the reference to the manager. A stopped store can register, and an
extra unreachable store reaches the source capacity rejection. The first FP
operation performs connection setup. Worker registration still executes its
source URI calls and therefore legitimately contacts the worker during registration.

The full native EWD840 MC06 process harness supports two standalone fingerprint
servers. It inspects the coordinator's manager reference graph after complete
initialization and before worker launch, without truncating model exploration.
Both distinct stores must be nonempty and total 16,384 initial fingerprints.
Each owned FP process has a private TMPDIR, representing separate-host storage
even when their source timestamp-based directories are created simultaneously.
One worker then completes the unchanged N=7 model: FINISHED, 114,942 distinct
states, zero queued states and no GENERAL/lost replies across all four roles.
These native topology choices do not complete the disabled upstream harness.

Native payloads now include a ValueArrays table. Composite fields and operator
argument rows hold one-based references to shared backing value slices, while
synthetic scalar edges remain inline. Encoding reserves array IDs before visiting
children; decoding allocates all Value objects before resolving arrays and then
populates owners, preserving recursive graphs. Exact nonempty backing pointer
and length identify source arrays; distinct equal-content arrays remain distinct.
Receiver arrays own their storage. Nil and initialized empty arrays remain
separate; empty-array object identity and arbitrary overlapping Go slice views
are not represented as Java array objects. Invalid/conflicting references fail
graph decoding. All communicating roles need the same native payload build.
ValueVec identity/full capacity and shared UniqueString name arrays remain pending.
This is a native Go graph representation, without Java serialization or RMI.

Native state payloads now include ValueVectors with a backing-array reference and
active count. Enum values reference vector IDs, preserving shared vector objects
as well as distinct vectors over shared storage. Encoding visits the entire
capacity, including inactive slots retained after sorting/deduplication. Vector
IDs are reserved before walking their arrays. Decoding resolves arrays against
allocated Value objects, validates each count against backing length, constructs
vectors, then populates Value owners. This supports cycles through inactive
storage. Appending within capacity changes shared storage; shared vector owners
also observe the updated active count. Receiver storage remains isolated from
the sender. Existing inline enum fixtures remain accepted. Communicating roles
need the same build. Shared name-array transfer remains pending; no Java runtime
or transport machinery is introduced.

Record and record-set names now reference a native NameArrays table independently
of the UniqueString object table. Exact nonempty backing pointer/length keys
retain one shared receiver array while distinct equal-content arrays stay
separate; name objects can still be shared across those arrays and StringValue.
Encoder roots keep pointer-keyed storage alive. Decoding resolves name arrays
after allocating strings and before populating Value owners. Null/initialized
empty arrays remain distinct; conflicting inline/null representations and
out-of-range array/string references are rejected. Existing inline fixtures
remain accepted. As with the value-array table, arbitrary overlapping Go slice
views and empty-array object identity are outside the source array representation.
All communicating native roles use the same build; this adds no Java machinery.

The native EWD840 process matrix now has mid_run_checkpoint_recovery with local
MemFPSet storage. An owned producer uses the production coordinator publication,
application loading and ModelCheck with a real TCP worker. After fingerprints
exceed the complete 16,384-state initial frontier, it calls production Checkpoint.
A forwarding StateQueue wrapper captures fingerprint/queue counts at BeginChkpt,
after all server threads suspend; counters after resume would describe a newer
frontier. The producer exits abruptly only after checkpoint commit, leaving the
saved files for a fresh real -recover CLI coordinator. Its old worker is retired
and joined before the fresh worker starts. Recovery must report those exact saved
counts once, skip initial-state regeneration and finish the unchanged MC06/N=7
model with 114,942 distinct states and no queued states. The recovered roles must
have no GENERAL/lost replies and exit normally. In both verified runs, the saved
frontier contains 20,480 fingerprints and 12,288 queued states. This is post-commit
process-loss coverage, not interruption during commit or distributed checkpoint
atomicity. No disabled original-harness completion credit is assigned.

The native checkpoint_interruption_before_commit process row first creates a
complete initial-frontier checkpoint, then recovers it in an owned producer and
advances the unchanged MC06 model with a real TCP worker. The production barrier
writes temporary queue/trace/FP/intern files. A forwarding queue exits the process
at CommitChkpt entry, before the first replacement-file commit. The old worker
is retired; fresh CLI roles recover and finish the original full model.

Do not infer coherent rollback of all components here. TLCTrace.recover seeks to
its saved cursor without truncating the trace. MultiFPSet.recover(TLCTrace) reads
an enumerator whose length is the complete file length; it does not use the old
child fingerprint checkpoints. Default nesting wraps the selected MemFPSet in
MultiFPSet. Thus queue recovery reads the old 16,384-state frontier while FP
reconstruction reads the persisted 20,480-record trace in the verified runs.
The test enumerates that persisted trace before restart to require its exact FP
count, retains the exact old queue count, forbids duplicate trace fingerprints
and lost flushed records, requires one started/uncompleted checkpoint with a
real producer worker and no premature FINISHED/GENERAL, then retains all final
EWD840 assertions. It does not claim atomic recovery or between-commit safety.
A separate audit found that Go TLCTraceEnumerator suppresses read/cursor errors;
source methods propagate IOException. This production gap remains pending.

TLCTrace.Elements and TLCTraceEnumerator.NextPos/NextFP/Reset now return native
Go errors instead of suppressing source IOException. Enumerator creation reads
owner RAF length and opens a read-only handle without the former extra flush;
source creation does not flush buffered writes. Missing files and closed owners
fail. NextFP preserves both predecessor/read-long failure paths, including a
valid zero fingerprint distinct from an error. Closing an enumerator retains
its closed handle state, so subsequent cursor/read calls fail rather than
falling back to an empty memory enumeration. Reset refreshes length from the
owner, preserves the current reader cursor for -1, opens before replacing the
reader and propagates length/cursor/open/seek failures. Native ownership releases
a successfully replaced read-only handle; a failed open keeps the old reader.

MultiFPSet, DiskFPSet and OffHeapDiskFPSet trace reconstruction now propagate
creation/cursor/read errors before inserting failed records, and return normal
reader-close failures. Deferred native cleanup on failure leaves its original
cause intact. CheckImpl creation/reset/trace generation and all known tests and
process oracles consume the error-returning API. Existing assertions are retained.
The full native interrupted-checkpoint recovery still finishes the unchanged
MC06 model after restoring old queue/full-trace fingerprints. This repairs the
previously documented error-propagation gap; it does not establish atomicity
across checkpoint file commits or broader distributed completion.

TLCTrace.Recover now retains a closed owner instead of calling the helper that
would reopen its RAF. The disk checkpoint fields are still read first; lastPtr
is assigned immediately after its field read, before input close and seek,
matching the source mutation order. A seek on the nil/closed native owner returns
its I/O error. This is a trace-recovery lifecycle correction; it adds no transport
compatibility behavior. Short TLCServer.ModelCheck checks require missing or
truncated metadata and a negative cursor to escape before queue recovery,
hostname resolution, registry creation, initial-state work or recovery-end output.
The method does not mark Done or emit GENERAL in that path; the outer command
catch owns reporting. Further failure phases and broader parity remain pending.

DiskStateQueue.Recover now follows source mutation order for partial state reads.
It retains untouched buffer entries, stores each NewEmptyState into its selected
slot before calling Read, and leaves completed state/header mutations visible if
a later read fails. TLCStateMut.Read publishes WorkerID and UID immediately after
each successful field read, rather than waiting for all three header fields.
Normal input close happens once before pool-reader restart and loFile update;
native failure cleanup does not double-close the stream. Four native corrupt
queue cases run through TLCServer.ModelCheck after valid trace metadata recovery.
They require retained inactive/unreached slots, completed enqueue records,
partial worker/UID fields, the original queue-header updates and failure before
fingerprint recovery, hostname lookup or publication. These checks add no
original-method credit. Later recovery phases and broader parity remain pending.

MultiFPSet.RecoverTrace now invokes fpSet(fp).RecoverFP(fp), matching source
getFPSet(fp).recoverFP(fp). Its own RecoverFP method remains the separate source
operation that uses parent Put and the FP-not-in-set assertion. The old trace
path incorrectly used that parent operation, bypassing child disk/off-heap
recovery insertion/flush semantics and replacing duplicate recovery-corruption
handling with FP-not-in-set. Four native cases retain the exact child runtime
code/parameters or one duplicate-warning event with continued reconstruction,
source option settings, selected partitions and prior insertions. Error-message
construction emits its source formatting-only recorder event without printing
a warning; the native oracle distinguishes that event from diagnostics. A fifth
case injects a selected child's checked I/O failure after real trace and disk
queue recovery, requiring original cause identity, partial insertions and no
later publication or recovery-end/initialization output. No disabled original
harness completion credit is assigned. Broader parity remains incomplete.

Memory/parent fingerprint RecoverFP duplicates now return the coded runtime
assertion failure used by source Assert.check, retaining its existing code and
message. MemFPSet, MemFPSet1, MemFPSet2 and MultiFPSet previously returned the
evaluator carrier. Existing short duplicate checks now require the runtime
category, preserved size/membership and native payload category. A real TCP
MemFPSet file recovery through DistributedFPSetManager requires the runtime
failure to escape its I/O catch, retain the first insertion and stop before
the next record. No original Java method completion credit is added.

DiskFPSet.recover(String) reopens worker readers followed by pool readers,
closing and replacing each individual slot before advancing. Go previously
closed/cleared both complete arrays before recreating them, losing source
partial mutation and resetting the pool cursor on failure. The native helper
now retains source slot replacement order and resets poolIndex only after
success. Recovery holds the table write lock; publication on helper exit makes
the native atomic reader snapshot reflect the retained slots on success or
failure. File/index reconstruction still precedes reader reopening. Supplemental
checks cover worker/pool close failures after earlier replacements, first reopen
failure and success, including exact retained pointers, closed states, completed
file/index/write counts and pool cursor. No original-method credit is added.

MemFPSet1.recover(String) delegates to SetOfLong.recover. Source assigns count,
length, threshold and zero flag immediately after each successful read, then
allocates exactly that length and uses ordinary put for its fixed record count.
Go previously delayed header mutation, clamped empty/negative lengths, rebuilt
the count and bypassed growth. Those shortcuts are removed. Source grow resets
count and rehashes nonzero entries while retaining hasZero without counting it
again. The Go growth method now does the same. This intentionally retains Java
quirks: two recovered records can yield count four, and growth can retain zero
membership without counting zero. Tests check these observations, all 29 partial
file lengths, negative allocation after header mutation and native TCP complete/
partial file recovery with the manager's original I/O catch. The subsequent constructor audit also restores exact zero allocation and
negative-size failure. This does not claim distributed parity complete.
No upstream SetOfLong test exists.

Memory fingerprint BeginChkptFile methods now open the supplied path directly,
matching FileOutputStream constructors used by source MemFPSet/MemFPSet1/
MemFPSet2. The former MkdirAll shortcut suppressed missing-parent I/O failures.
The native OS file error retains the I/O category through the native failure
payload and manager catch; no Java file-runtime compatibility is introduced.
Six local/TCP cases require failure without directory creation or storage
mutation, including the original manager storage diagnostic. NewSetOfLong now
allocates zero length exactly and rejects negative length rather than changing
both to one; first insertion grows the zero-length set as in source.

MemStateQueue, DiskStateQueue and TLCTrace commit failures now use the existing
I/O failure carrier rather than plain formatted errors. Source diagnostics and
file promotion order remain unchanged. Native payload encoding retains the I/O
trait. Checks cover failed old-file deletion, successful promotion and missing
temporary files after old-file deletion; disk pool deletion retains earlier
removed files without advancing lastLoPool or touching later files after failure.
Real coordinator checkpoint cases remove a queue/trace temporary at the source
commit boundary and require queue resume, earlier begins/commits, stopped later
commits and no checkpoint-end event. These are native supplemental checks.

MemStateQueue.recover now retains the source array and start cursor, mutates
len after its header read, and publishes each empty state before its read.
The previous clearing/resizing/cursor reset and deferred slot publication are
removed. Capacity overflow returns the represented bounds failure after the
earlier records and len mutation; storage is not grown. Native input ownership
is released once on success or failure without replacing the failed read cause.
Checks cover missing UID/level, completed reconstruction, the fixed 4096-entry
capacity boundary and real coordinator recovery before fingerprints/publication.
The subsequent DiskFPSet checkpoint ownership fix retains the source success-only
lock release and flag reset. No distributed completion claim follows this chunk.

DiskFPSet.beginChkpt(String) sets the flusher flag, acquires all table write
locks, flushes, copies and advances the checkpoint marker before releasing
locks and resetting the flag. The Go defer that released ownership on failure
is removed. Short native checks use TryLock to verify retained stripes without
starting blocked calls, covering flush I/O, runtime duplicate merge, copy I/O
and success. Owned fixture cleanup releases retained locks only after assertions.
A two-partition native TCP manager check requires completed failed-store flush,
unchanged marker/registrations, the source diagnostic and healthy-store commit.
Storage operations needing retained locks remain blocked after the source
failure; native transport does not retry or reset that storage ownership.

The native EWD840 process matrix now includes interruption after queue commit,
using the same production checkpoint and recovery paths. The producer starts
from an older complete initial checkpoint, advances with a real worker, suspends
through the source barrier, writes all temporary files, commits the queue and
exits before trace commit. Parent inspection requires the new exact queue count,
unchanged old trace metadata, absent promoted queue temporary and retained trace/
intern/FP temporary files. MultiFPSet recovery still enumerates the full persisted
trace independently of the old metadata cursor. Fresh CLI coordinator/worker
processes retain the full unchanged model and final assertions; no atomicity
protocol or disabled Java harness completion credit is introduced. Later trace/
intern/fingerprint commit boundaries remain pending.

The native EWD840 process matrix additionally checks interruption at entry to
fingerprint commit, after queue, trace and intern commits. Test creation wraps
the same configured production memory FP factory before recovery/initialization;
the discarded unused initial memory manager owns no open files. All storage
operations delegate to the resulting production set, and only CommitChkpt
triggers abrupt exit. No production checkpoint callback is added. Parent
inspection requires changed committed trace metadata, promoted queue/trace/intern
temporaries and retained FP temporaries. Fresh CLI coordinator/worker recovery
passes with the unchanged model and all final assertions. Isolated trace-commit
and nested fingerprint-commit interruption remain pending; source atomicity
or disabled Java harness completion is not claimed.

Native EWD840 process recovery now covers interruption between the two nested
MemFPSet commits. A test-only wrapper on child zero delegates its real commit
before exiting, leaving child one's old checkpoint and its new temporary file.
Parent inspection compares both FP checkpoints with the older complete baseline:
first-child contents changed and its temporary is promoted; second-child contents
remain byte-identical and its temporary remains. Earlier queue/trace/intern commits
also remain. MultiFPSet recovery reconstructs fingerprints from the full persisted
trace, preserving source behavior with mixed child checkpoint generations. Fresh
CLI coordinator/worker processes complete the unchanged model. No atomicity
protocol or production hook is added; isolated trace-commit interruption and
broader failures remain pending.


### Required distributed coordinator ownership

TLCServer checkpoint/recover/close use required queue, trace and fingerprint
manager ownership, matching source operation order. Null receiver/component
access raises the existing native null-failure trait rather than silently
skipping work. Guards occur at the actual access: checkpoint can retain the
queue temporary file before a missing trace, or queue/trace temporary files
before a missing fingerprint manager; it does not resume or begin intern-table
checkpointing on those failures. A false suspend result bypasses later owners.
Recovery publishes trace reads before accessing the queue and completes queue
recovery before accessing the manager. Close closes the trace before accessing
the manager and deletes metadata only after both close operations succeed.
Earlier I/O failures still return unchanged and stop subsequent operations.
No transport compatibility or JVM machinery is involved. Supplemental native
ownership tests and focused actual TCP checkpoint checks pass, without new
original-method completion credit or a broader completion claim.


### Distributed coordinator progress locale formatting

Periodic TLC_PROGRESS_STATS formats its five signed counters through the
existing MessageNumberFormat, corresponding to Java MP.format. Final tool
progress uses the same formatter for generated, distinct and queue counts.
Source conversion from unsigned Go fingerprint totals to signed long bits is
retained. Report rates and baseline timing are unchanged. Trace levels use
plain integer strings, final rate zeros are literal ASCII "0", and final
TLC_STATS parameters remain plain signed strings as in Java printSummary.
Nine fresh Go processes validate source DecimalFormat reference rows for
locale grouping/digits/negative affixes, positive/negative rates and signed
limits, including POSIX explicit grouping and Arabic/Persian bidi affixes.
Tool mode controls final progress, while success controls search depth.
No new formatter or Java runtime dependency is introduced. Supplemental tests
add no original-method completion credit; existing MP tests remain green.


### Distributed trace depth read failures

TLCTrace.GetLevelForReportingWithError supplies the native checked I/O boundary
for coordinator and management reporting. Disk predecessor traversal first
saves the file cursor, reads the chain without a synthetic cycle early exit,
and seeks back only after successful traversal. Read/seek failures propagate
and retain prior cursor mutations. The previously reported maximum is updated
only on a successful traversal; closed trace ownership produces I/O failure.
The integer-only convenience getter raises a returned error. The ordinary
in-memory trace path retains its separate representation.

Periodic coordinator reporting computes rates before depth access, then stops
before message output, waiting or baseline publication on a trace error. Final
reporting stops after worker/executor shutdown and final count collection but
before rate reset, postprocessing, summary/finished output and cleanup. Native
management getters catch I/O separately: plain ModelChecker progress returns
-1, while distributed management also prints the existing native stack. The
concurrent trace's prior embedded fallback behavior is unchanged and is not
claimed as fully audited here. Focused native failure tests, original
TLCGetLevel/TTrace methods and actual TCP checkpoint checks pass. Supplemental
failure checks add no original-method completion credit.


### Distributed server-thread catch/finally and goroutine ownership

Run uses separate deferred catch and finally operations. The catch is executed
first; an error raised inside it still executes finally and escapes to the
owned goroutine boundary. Compute/publication failures outside their narrow
remote/null catches raise into this one outer catch rather than invoking the
handler from within the body and recatching handler failures. Finally reads
remote cache statistics, cancels keepalive and clears assigned states in order.
Returned/panicked remote cache failures enter the same warning catch. Unchecked
cache failures escape and skip the remaining finally operations.

Start prints uncaught failures using existing native stack diagnostics before
closing its join channel. Failure terminates only the owned goroutine, matching
the source thread lifetime. Direct Run calls still propagate an uncaught
handler/finalizer failure to their caller. No extra model cleanup, worker count
decrement or result mutation occurs. Five joined native child-process tests
and focused local/TCP checks verify these boundaries; supplemental checks add
no original-method completion credit. No Java runtime machinery is introduced.


### Coordinator keepalive remote-failure forms

TLCTimerTask normalizes returned and panicked IsAlive endpoint failures using
the existing generic invocation boundary, then catches only the source remote
failure category. False liveness and either remote failure form invoke
HandleRemoteWorkerLost. Its one-time cleanup retains state identity/order and
exactly one deregistration/count decrement across repeated task invocations.
Unchecked local failures escape without ownership changes; recent or future
last-invocation timestamps suppress status calls. Native TCP fatal handler
failures retain the remote I/O category and invoke the same cleanup without
stopping the worker host. Focused synchronous task and owned TCP checks pass;
no source test methods exist for these boundaries and no method credit is added.


### Final worker exit and shutdown hook failure traits

Normal coordinator completion normalizes an exit call's returned/panicked
failure before applying its existing ExitIgnorable catch. A warning is printed
for that category and completion continues. Registration removal runs in finally
for success and every failure; other failures stop before later workers and
executor/final-result publication. The shutdown hook retains registrations,
silently ignores WorkerUnavailable, reports other I/O failures and continues,
and lets unchecked failures stop iteration.

WorkerUnavailable is a native operation/graph payload trait for missing worker
ownership or a dead connection. It preserves the source hook's narrower catch
without inspecting Java class strings or reconstructing remote exceptions.
ExitIgnorable also covers source server failures during normal completion,
which the hook must report. Native TCP unpublished endpoints and dead transport
connections set both traits; graph encode/decode retains them independently.
Communicating roles need the same payload build. The source GENERAL throwable
overload formats a string event, which the existing Go printer preserves.
Focused local/payload, fresh completion-process and actual TCP checks pass. No
direct original methods cover these exit-order boundaries; no method credit is
added and broader distributed completion remains unproven.


### Distributed cache raw ratio arithmetic

SimpleCache.GetHitRatio loads the signed hit count before the signed miss count
and performs native floating-point division, matching the source expression.
Zero denominators produce IEEE infinity/NaN rather than a substitute zero;
signed zero and signed-counter overflow conversion remain observable through
the worker endpoint. Missing worker/cache ownership raises the existing null
trait. Reference tests cover 271 original-expression counter pairs, with NaN
classified independently of architecture-specific arithmetic sign/payload.
Native TCP checks require exact sender/receiver bits. No original SimpleCache
methods exist and supplemental tests add no method credit.

GetHitRatioAsString uses a native compact formatter for worker exit messages,
separate from the coordinator's two-decimal statistics formatter. It groups
digits and emits at most three fractional digits with source rounding, localized
digits/decimal separators/negative affixes and locale-specific NaN/infinity
labels. Integral conversion follows FloatingDecimal.developLongDigits: discard
decimal places implied by the binary exponent, then round the retained integer.
Go shortest-decimal conversion alone disagrees on large integer values.
Special-symbol tables include numbering replacements and legacy Japanese/Thai
variants. The formatter shares the existing one-time locale initialization.
Saved Java reference vectors cover 271 counter ratios, 6,610 additional numeric
values (large integers, power-of-two divisors, rounding ties and counter edges),
1,860 locale/numbering rows and six fresh initialization processes. These native
checks require no Java runtime and add no original-method credit.

The native worker cache reply stores CacheRateBits as uint64 and decodes it
with math.Float64frombits. Gob's default float-field omission otherwise turns
negative zero into positive zero. Nonzero sign bits remain explicit with this
representation; finite/nonfinite values and NaN payloads retain exact sender
bits. Communicating roles need the current native reply build. Focused TCP
extreme/lifecycle/fatal-boundary checks pass.


### Distributed management control synchronization

TLCServerMXWrapper.Stop, Suspend and Resume now acquire the coordinator monitor
and release it with defer, matching the source synchronized blocks. Stop sets
the done flag before FinishAll, then notifies the existing reporting waiter.
Queue failure leaves done set but does not notify; missing queue ownership
fails at that same point. Suspend/Resume neither change done nor notify and
retain queue failures. Missing wrapper/server ownership fails before locking.
The existing native monitor permits reentrant control callbacks. No JMX or RMI
implementation is added. Six queue success/failure cases, nine missing-owner
cases and an actual joined zero-timeout report wait verify these boundaries.
Upstream has no direct management control tests; no original-method credit is
added. Broader management query parity remains separate.


### Distributed management query ownership

Management queries require their coordinator before testing IsRunning or
delegating. Missing ownership must not look like an inactive coordinator or a
zero-valued statistic. GetStatesGenerated reads the worker delta, then requires
the FP manager and adds GetStatesSeen with signed-long overflow. GetNewStates
requires queue size before adding each registered thread’s assigned block under
the coordinator monitor. GetAverageBlockCnt requires the selector; the earlier
extra AverageBlockCnt fallback field is removed. Running GetProgress requires
the trace; its existing I/O catch still prints the failure and returns -1.
Current-state queries require the queue and return N/A only for a null peek.
Inactive generated/distinct/progress and spec/model queries retain their source
sentinels. Distinct count additionally retains the explicit missing-FP-manager
-1 result. Queue size, rates, worker count and selector average remain
unconditional after completion. Supplemental native query checks add no
original-method credit. No Java management or remote runtime is introduced.


### Distributed worker vector collection policy

Java’s TLCWorker uses TLCStateVec for predecessor, successor and result
partitions, separately from its tool’s StateVec. The former starts at capacity
10, doubles without consulting TLCGlobals.setBound and indexes backing-array
capacity instead of active count. The Go StateVec storage now carries this
distributed collection policy, selected by the worker partition constructor.
Ordinary tool vectors retain their existing bound and active-slice behavior.
Growth follows signed-int doubling with the required minimum capacity. Unused
slots contain null; indexes beyond capacity raise the existing bounds category.

Native result decoding selects this policy with capacity equal to the active
count. Wire vectors still omit spare entries and preserve vector identity and
shared state roots; no wire format changes or Java serialization are introduced.
Coordinator publication honors backing-capacity indexing for in-process
distributed partitions before checking the selected state. FP insertion still
precedes null/bounds failures and trace/queue publication. Local/gob/TCP checks
require eleven returned successors with tool SetBound one, capacity 20 locally,
capacity 11 after transfer and capacity 22 after receiver growth. Malformed FP
selections at indexes nine and ten retain their distinct null-slot/state-check
and out-of-capacity failure contexts. Supplemental checks add no method credit;
there are no direct upstream TLCStateVec tests.


### Native predecessor graph transfer

TLCState implements Serializable in the Java reference; TLCStateMutExt’s
predecessor is an ordinary, non-transient TLCState field. Blanket rejection of
predecessors therefore excluded state graphs that have no action or callable.
The native DistributedStateNode now has a one-based Predecessor reference into
the existing state graph. Encoding reserves each state ID before following its
parent, preserving cycles and shared/non-root ancestors. The same value/array
tables apply to roots and ancestors, retaining shared value storage.

Decoding allocates all states before linking predecessor fields. It assigns the
fields directly rather than invoking SetTracePredecessor, which would alter
stored levels and consult process-local metadata configuration. Zero references
remain null; invalid IDs fail explicitly. Only supported native mutable-state
predecessors are accepted. Functional states, evaluator action/callable/cache
objects, record-backed states and custom predecessor types remain outside this
chunk’s supported representation. No evaluator objects are silently dropped.

Gob and TCP request/result/failure checks preserve shared parents, hidden
ancestors, typed nulls, self-cycles, stored levels, shared values and independent
receiver ownership. All communicating roles need the current payload build. No
Java serialization protocol, RMI or JVM runtime is introduced. No direct
upstream predecessor-transfer tests exist; native checks add no method credit.


### Native cached-state map transfer

TLCStateMutExt’s non-transient cache is Map<Integer, Value>. The native graph
now has a StateCaches table and a one-based cache reference per state node.
Zero denotes a nil cache; a referenced empty table denotes a non-nil empty map.
Encoding tracks map identity through native reflection, retaining the maps
throughout encoding. Sorting signed-int keys gives deterministic native entry
order. Keys outside the source int32 range are rejected instead of truncated.
Each entry references the existing value graph, preserving cached values shared
with state values and other maps, including recursive values and null entries.

Decoding creates a distinct map per table entry and reuses it for all referring
states, including predecessor states. Equal-content separate maps remain
separate; shared empty maps remain shared after later insertion. Negative or
out-of-range map/value references and duplicate keys fail explicitly. Cache
values retain the existing codec’s supported-value and failure contracts; no
arbitrary evaluator objects are reconstructed or silently dropped. All roles
need the updated native payload build. No Java serialization or RMI is added.
Gob/TCP requests, results and WorkerException contexts verify mutation-visible
sharing with isolated sender ownership. There are no direct upstream cache
transfer tests, so supplemental checks add no original-method credit.


### Native record-backed printable state transfer

RecordValue.PrintTLCState inherits TLCState’s serializability and holds its
record plus underlying state. The native Go state representation’s printRecord
now refers to the existing value graph through DistributedStateNode.PrintRecord.
Decoding requires a RecordValue and rejects invalid IDs or other value types.
Sharing with state values, caches and other printable states remains visible
without aliasing sender storage. State fields and ordinary-state distinctions
remain unchanged; all communicating roles need the updated payload build.

Transfer exposed a separate RecordValue.ToState shortcut: source
UniqueString.equals(UniqueString) compares tokens, while Go compared pointers.
ToState now uses token equality when binding spec variables, allowing received
record names to bind local variable objects. PrintTLCState’s _format selection
uses Arrays.asList.indexOf/Object.equals instead: UniqueString does not override
Object.equals. Its existing identity behavior is retained and explicitly
checked; received format-name objects do not match the static format object.
Gob/TCP request/result/WorkerException checks preserve default display, extra
record fields, shared records/values, stored UID/level and state fingerprints.
Existing original record methods and Alias safety checks remain green. No direct
upstream printable-state transfer tests exist; supplemental checks add no
method credit. No Java serialization/RMI/JVM implementation is introduced.


### Coordinator thread block-selector ownership

TLCServerThread retains the constructor's supplied selector; registration passes
the coordinator's chosen selector explicitly. The constructor and run loop do
not substitute policies or create selectors. A missing selector fails at block
selection, entering model-error reporting, queue completion and thread finally
cleanup. A recoverable batch failure requeues the assigned states before updating
the selector limit; a missing selector at that point retains the requeued work
and propagates its failure. Concrete static/proportional selector limit updates
remain no-ops. Focused native checks and existing original smart-proxy/short TCP
checks pass. Upstream has no direct selector-ownership test methods, so these
supplemental checks do not change original-method completion counts.


### Required requeueing before recoverable-batch limit updates

The coordinator's recoverable worker-failure catch requires the supplied queue
before reducing the transfer limit. It prints the exceed-blocksize diagnostic,
requeues the assigned block, updates the selector and then resumes selection.
A missing queue fails at the requeue operation. A failed enqueue retains its
preceding mutations and original failure; the limit update and continuation are
skipped. Failures raised inside this catch escape to the enclosing model-error
handler rather than being classified again as worker loss. Focused checks cover
missing queues, partial I/O/runtime/fatal failures and healthy retry; short TCP
checks and actual DieHard distributed trace execution remain green.


### Distributed block-selector startup settings

Selector mode is captured once on the first factory call, before constructing
a selector. Static takes precedence over unlimiting, which takes precedence over
limiting; otherwise the statistical selector is used. Static batch size has a
separate one-time capture on first static construction, before checking its
server argument. Using other selector modes does not initialize static size.
Later process-property changes cannot alter either captured setting. This is
implemented with native sync.Once fields, without Java class loading or runtime
machinery. Existing explicit Go constructor size arguments override the captured
default per instance. Fresh-process setting-lifetime checks and existing original
smart-proxy/distributed model and short TCP checks pass.


### Required owners and ordering of batch-success statistics

Coordinator threads first add the first returned partition's size to their
received-state count, then update their timer task's keepalive timestamp, then
add the result's generated-state delta to the coordinator. Neither owner is
optional. The result delta retains source subtraction of partition count,
rather than total successor count. Missing owners yield explicit null failures
within the existing worker-call catch boundary. An absent timer therefore
triggers worker-loss cleanup and requeues the assigned states; it does not return
a publishable result. Counts already updated remain updated. Production thread
construction supplies the task; direct-construction test fixtures must do so too.
Focused native checks cover mutation/failure ordering and assigned-state recovery.


### Worker-loss consumer notification

Worker loss cancels keepalive, claims the idempotent cleanup flag, deregisters,
requeues unfinished states, clears assigned work, wakes all queue consumers and
then decrements worker count. Consumer notification is unconditional even when
the queue is empty or suspended. StateQueue.WakeAllWaiters provides this native
operation with a locked consumer-condition broadcast in all four implementations;
it changes neither suspension nor completion. ResumeAllStuck is a different
operation, including checkpoint-barrier recovery, and remains the registration
wakeup. Joined suspended-consumer tests, cleanup-order checks and exact short
race checks cover the corrected boundary.


### Coordinator fingerprint registration mode

Base TLCServer rejects fingerprint-server registration at the coordinator
boundary before reading its manager, using the source server-level unsupported
operation. A dynamic manager supplied to a base server does not grant distributed
registration capability. Distributed coordinators carry startup registration
state; they register the supplied endpoint, then count down the latch and print
acceptance. Local and native TCP checks preserve base rejection without manager
mutation, probing or acceptance output. Existing distributed registration and
original fingerprint-manager tests remain green.


### Native invocation text in distributed usage errors

Invalid coordinator options and worker/fingerprint-server arguments retain their
source validation, diagnostic ordering and early-return behavior. Their usage
lines name the corresponding native tlago commands. They do not direct Go users
to execute Java classes. Explicit Java counterpart descriptions in CLI help are
reference documentation. Exact output checks cover worker/FP zero and extra
arguments, skipped discovery and both server ToolIO output destinations.


### Coordinator worker statistics labels

Coordinator thread labels retain TLCWorkerThread-, the source prefix consumed
by external statistics tooling, followed by a minimum three-digit worker
counter and the endpoint URI in brackets. URI rendering uses the existing
native ASCII formatter, preserving normalization and escaping without modifying
the URI metadata used for calls or deregistration output. These are diagnostic
labels for Go goroutines, without JVM thread emulation. Focused label, URI and
thread failure/finalizer checks pass.


### Distributed initial-state trace publication

DoInitFunctor inserts the fingerprint before publishing a new in-model initial
state. It requires the trace, writes a root record with predecessor marker 1,
assigns the resulting UID, requires the queue and enqueues the original object.
It then checks properties. The record-only trace writer preserves worker,
predecessor, action and level metadata; the single-process metadata attachment
path is not used. Failed trace/queue operations retain earlier fingerprint/write
mutations and record the original state/error through the initialization catch.
A recorded failure suppresses subsequent elements. Seen states skip trace/queue
access and property checks; excluded states skip publication owners but still
check properties. Focused native contracts, original initializer/model checks
and actual distributed DieHard trace execution pass.


### Trace last-record pointer on partial writes

Native trace writers assign the attempted record's file pointer to lastPtr before
writing predecessor and fingerprint bytes. The assignment survives partial write
failure. Buffered bytes/cursor mutations also remain, while state UID/metadata
updates and native record publication require successful writing. This order
matches TLCTrace's common source writer and applies to initial, distributed
successor and single-process successor paths. Closed native descriptor checks
force predecessor/fingerprint buffer-boundary failures without production hooks;
focused trace, checkpoint and native TCP checks pass.


### Error-trace evaluation failure propagation

State reconstruction and transition reconstruction errors returned by Tool, and
alias errors returned after Tool's own exception handling, propagate from trace
printing without replacing the result with a fallback state. Previous output
remains; subsequent states are not printed. The coordinator catches ordinary
trace-printing errors, reports them and finishes/notifies its queue. Fatal errors
escape that catch after preceding model-error mutation and skip queue completion
and notification. Existing tool-local alias EvalException/runtime handling is
preserved. Noninitial printing branches `3`, `4` and `5` now preserve nil
reconstruction rather than fabricate a state. They print recovery and bug
diagnostics, optionally the standalone unrecovered state (branches `4` and `5`),
and terminate the native process with `os.Exit(1)`. This bypasses error catches
and deferred cleanup as the source requires. Child-process checks exercise the
actual exit without a production test hook. Fingerprint-sequence recovery branch
`2` also exits after source diagnostics, formatting the fingerprint as a signed
64-bit value. It uses the predecessor-state lookup, restores randomness only
after normal completion and preserves returned info metadata. Fingerprint-only
and predecessor-state lookups return nil for no match; predecessor-info lookup
returns EvalException with FAILED_TO_RECOVER_NEXT. A sole missing initial result
remains a nil array element; with more fingerprints, dereferencing it fails before
another lookup. Printing now copies UID/worker metadata only in the source
nonempty-prefix current-state and noninitial successor branches, before alias
evaluation. Empty-prefix and initial-transition reconstructions retain their own
metadata. A missing initial transition reaches the state printer and raises a
null-pointer failure before another state event, allowing the ordinary coordinator
catch to report the printing failure and finish/notify the queue. The state printer
requires a nonnil info/state instead of rendering an empty substitute. Disk
fingerprint traversal restores its saved cursor only after successful reads and
propagates restoration failures instead of publishing a fingerprint sequence.
Failed reads retain the consumed cursor. Traversal follows predecessor links until
the source initial-state sentinel, without returning a partial self-linked chain.
Concurrent reconstruction uses a result length of record count minus one,
excluding the anchor record. It restores randomness only after normal completion,
retains source metadata updates before later failures, fails on missing initial
metadata dereference and exits on missing successors with branch `2` diagnostics.
Public methods propagate reconstruction errors before trace printing rather than
swallowing them and selecting another trace. Worker-slot lookup preserves
index/null failures and does not select an in-memory fallback. Initial/equal-state
short paths do not access workers. Source record collection rereads the end record
under the trace monitor, stops at the initial or requested-fingerprint boundary
and retains the monitor through state reconstruction. Predecessor access reads
the selected record even when its pointer is the initial-state sentinel; the
caller controls traversal termination. Checkpoint begin, commit and recovery
require every worker in source order; failure preserves earlier mutations and
stops before later workers or marker publication. Begin holds the trace monitor,
as does level reporting. Concurrent levels use worker maxima with a minimum of
one and do not consult the base trace level. Enumeration constructs required
worker readers under the trace monitor, preserving source creation order without
an added cleanup loop on construction failure. A worker reader snapshots the
existing writer cursor and opens a separate read-only file without flushing or
reopening the writer. Cursor failures propagate before selector advancement, and
exhausted fingerprint access returns an index failure instead of fabricated zero.
Close stops on its first failure and retains the reader owner, so subsequent
access still observes the closed resource. Position lookup checks only the next
neighbor; reset changes only the selector, preserving child-reader positions.
Reconstruction requires a tool only when performing a lookup; empty and supplied-
initial paths remain available without one. Missing-tool failures follow random
reset and do not restore the snapshot on failure. Printing requires its lookup/
alias tool, and a returned nil alias reaches the printer rather than producing
an unaliased substitute. Disk recovery no longer selects the in-memory path
simply because its tool is absent. Initial-only printing still omits tool access.
Worker record writers update lastPtr before writing predecessor, worker and
fingerprint bytes. Partial failure retains this attempted pointer and earlier
depth update while state metadata, unseen-successor count and mirror publication
wait for successful writes. Successor writes preserve the generated action and
source predecessor policy. The native mirror appends the published state record
and updates its reporting level without mutating the state again. Worker next-
level comparison uses signed 32-bit arithmetic before selecting the maximum.
Overflow preserves the prior maximum rather than storing a negative depth.
Record and state UID/worker updates precede the later trace-depth failure, while
count and mirror publication do not occur. Extended state predecessor assignment
also precedes the base depth check, retaining that assignment on failure. Worker
recovery publishes the last pointer immediately after reading it, before closing
the checkpoint reader and seeking the existing trace owner. Missing or closed
owners cannot be reopened by recovery; closing retains the closed owner handle.
Truncated reads do not publish the pointer. Worker checkpoint creation requires
its existing trace owner and flushes before opening temporary metadata, without
reopening the owner or consulting a saved creation error. Checkpoint commit
failures preserve the source I/O category and delete-before-promotion order.
Worker record reads require the existing owner rather than opening a replacement
or consulting a saved creation error. Mark publication precedes seek, including
closed-owner seek failure. Only a complete record restores the original cursor;
partial reads retain the consumed position and do not publish a record. Writes
require the predecessor before depth and file access; nil cannot select an
initial-state anchor. Missing targets fail after completed record bytes, retaining
earlier depth/last-pointer updates and suppressing metadata/count/mirror publication.
Closed-owner cursor errors precede target dereference, while missing predecessor
errors precede owner access. Worker construction, filename/context guards, writer
owner access and trace cleanup
boundaries still need audit. No RMI or JVM machinery is introduced.

Fingerprint transport failures use the shared DistributedOperationError with
the original Go cause retained. Dial, lazy connection, call and failure-payload
decoding errors preserve the manager's remote/I/O boundary; they do not acquire
worker smaller-batch retry or worker-exit traits. Established failed connections
are not redialed and interrupted calls are not replayed. A completed fingerprint
insertion may remain in storage after its reply is lost; its caller receives a
failure, not a fabricated insertion answer. The manager's existing source failover
policy remains unchanged. Application failure graphs retain their separate
diagnostic and catch categories.

Successor validation calls the state assignment check directly. A missing state
is an evaluator failure, rather than an incomplete-state result. The distributed
worker wraps that failure with its predecessor, cause and call-stack flag before
publishing generation totals. A present state with unassigned values retains the
ordinary incomplete-successor WorkerException without a cause or call-stack flag.
Native TCP preserves both outcomes without adopting Java transport machinery.

Fingerprint connection loss and process loss have separate coverage. The native
process check terminates independent MemFPSet, LSBDiskFPSet and MSBDiskFPSet hosts
after committing a checkpoint, adding a fingerprint and beginning another
snapshot. Disk rows verify the later membership was flushed to the live file;
the committed checkpoint remains byte-identical and pending snapshots are not
promoted. Manager failover preserves
the source partition reassignment and new-membership answers on surviving storage;
it does not copy the dead store. A fresh host starts empty and recovers committed
membership through its endpoint. Recovery neither reconnects the dead client nor
replaces the manager's registrations. Full-model server restart/recovery and network
partitions remain separate pending requirements. This short storage matrix does
not establish all disk-store process failure phases or full-model recovery.

Full native MC06 coverage now kills the first of two fingerprint hosts with a
worker block assigned. A test-owned pause precedes real successor evaluation;
resumption calls the production evaluator without fabricating states, fingerprints
or replies. Worker and coordinator retain their independent manager failover.
Source size() sums registration slots rather than unique endpoint identities,
so the surviving store is counted twice after both slots alias it. The separate
failure model row requires 229,884 reported states and queue zero; ordinary model
rows retain 114,942. This preserves source statistics rather than introducing
deduplication. Full-model server restart/recovery and other failure phases remain
pending; this native row adds no disabled Java harness completion credit.

Default Tool successor evaluation requires an action and dispatches its actual
predicate. Missing actions cannot return empty vectors. A missing predicate fails
at predicate dispatch, after successor predecessor/action setup, rather than
returning empty successors or a non-boolean diagnostic. DistributedWorker retains
these evaluator failures with predecessor context and a call-stack flag. A real
false predicate remains an empty successor set and follows ordinary deadlock
handling. Explicit evaluator overrides retain their existing dispatch boundary.

Literal TLCTrace filenames remain disk-backed even when their directory prefix
is empty. Checkpoint commit uses the captured literal filename, deletes the old
checkpoint before promotion and preserves the source IOException on failure.
Native Delete closes the owned handle and removes its trace/temporary/checkpoint
files for this prefix too. Memory-only traces retain their separate behavior.
The checks construct filenames entirely inside temporary storage, so no root
filesystem file is created. They do not establish the separate isolated full-model
trace-commit interruption boundary.

Native worker reply loss is covered separately from worker process death. The
short checks hold an actual computed result before endpoint serialization, close
the TCP host and let the coordinator finish source retry/loss cleanup before
releasing the reply. A multi-state EOF requeues the block and halves the limit;
the next closed-client call retires the worker and requeues the smaller assigned
block. A single-state EOF retires directly. Neither path publishes fingerprints,
trace records, received-state counts, timestamps or generation deltas. Repeated
loss/timer reports retain one deregistration, and late replies do not publish.
Closing transport does not terminate the owned worker runtime. This is short
native connection-loss coverage, not full-model network partition completion.

Call-stack evaluator construction requires the source Tool whose shared Spec is
copied. Nil cannot substitute a fresh default evaluator. Existing ordinary-tool
selection and fresh-stack behavior remain intact. Coordinator initialization
constraints precede fingerprinting: an in-model nil state already fails at its
state access before FP/trace/queue publication, while excluded states still reach
property checks. These initialization paths were verified without changing their
implementation; the replay constructor fallback was the actual repaired mismatch.
