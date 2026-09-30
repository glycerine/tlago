package tlago

import (
	"embed"
	"strings"
)

//go:embed test_vectors/java-sany/StandardModules/*.tla
var embeddedJavaStandardModules embed.FS

type LoadOptions struct {
	LibraryPaths         []string
	PreferLibraryModules bool
	ExtraModules         []string
}

var standardModules = loadStandardModules(map[string]string{
	"Naturals": `---- MODULE Naturals ----
CONSTANT Nat
====`,
	"Integers": `---- MODULE Integers ----
EXTENDS Naturals
CONSTANT Int
====`,
	"Reals": `---- MODULE Reals ----
EXTENDS Integers
CONSTANT Real
====`,
	"Sequences": `---- MODULE Sequences ----
Seq(S) == S
Len(s) == Len(s)
Head(s) == Head(s)
Tail(s) == Tail(s)
Append(s, e) == Append(s, e)
SubSeq(s, m, n) == SubSeq(s, m, n)
SelectSeq(s, test) == SelectSeq(s, test)
====`,
	"FiniteSets": `---- MODULE FiniteSets ----
IsFiniteSet(S) == IsFiniteSet(S)
Cardinality(S) == Cardinality(S)
====`,
	"Functions": `---- MODULE Functions ----
Range(f) == { f[x] : x \in DOMAIN f }
FoldFunction(op(_, _), base, f) == base
FoldFunctionOnSet(op(_, _), base, f, S) == base
SumFunction(f) == 0
SumFunctionOnSet(f, S) == 0
====`,
	"FunctionTheorems": `---- MODULE FunctionTheorems ----
EXTENDS Functions
FoldFunctionOnSetEmpty == TRUE
FoldFunctionOnSetEqual == TRUE
FoldFunctionOnSetType == TRUE
SumFunctionNat == TRUE
SumFunctionZero == TRUE
SumFunctionMonotonic == TRUE
SumFunctionStrictlyMonotonic == TRUE
====`,
	"FiniteSetTheorems": `---- MODULE FiniteSetTheorems ----
EXTENDS FiniteSets
FiniteSetCardinality == TRUE
FiniteSetExtensionality == TRUE
FiniteSetInduction == TRUE
SubsetIsFiniteSet == TRUE
UnionIsFiniteSet == TRUE
====`,
	"SequencesExt": `---- MODULE SequencesExt ----
EXTENDS Sequences, FiniteSetsExt
BoundedSeq(S, n) == Seq(S)
SetToSeq(S) == <<>>
SeqToSet(s) == {}
RemoveAt(s, i) == SubSeq(s, 1, i-1) \o SubSeq(s, i+1, Len(s))
Front(s) == SubSeq(s, 1, Len(s)-1)
IsPrefix(s, t) == Len(s) <= Len(t) /\ SubSeq(s, 1, Len(s)) = SubSeq(t, 1, Len(s))
IsStrictPrefix(s, t) == IsPrefix(s, t) /\ s # t
Prefixes(s) == {SubSeq(s, 1, l) : l \in 0..Len(s)}
CommonPrefixes(S) ==
  LET P == UNION {Prefixes(seq) : seq \in S}
  IN {prefix \in P : \A t \in S : IsPrefix(prefix, t)}
LongestCommonPrefix(S) ==
  CHOOSE longest \in CommonPrefixes(S) :
    \A other \in CommonPrefixes(S) : Len(other) <= Len(longest)
FoldSequence(op(_, _), base, seq) == base
====`,
	"SequenceTheorems": `---- MODULE SequenceTheorems ----
EXTENDS Sequences
AppendSequence == TRUE
HeadTailSequence == TRUE
TailInductiveDef == TRUE
SubSeqSequence == TRUE
====`,
	"SequencesExtTheorems": `---- MODULE SequencesExtTheorems ----
EXTENDS SequencesExt, SequenceTheorems
BoundedSeqIsFiniteSet == TRUE
SetToSeqIsSequence == TRUE
FoldSequenceType == TRUE
====`,
	"FiniteSetsExt": `---- MODULE FiniteSetsExt ----
EXTENDS FiniteSets, Functions
RandomSubset(k, S) == S
IsInjective(f) == TRUE
Max(S) == CHOOSE x \in S : \A y \in S : x >= y
Min(S) == CHOOSE x \in S : \A y \in S : x =< y
====`,
	"FiniteSetsExtTheorems": `---- MODULE FiniteSetsExtTheorems ----
EXTENDS FiniteSetsExt, FiniteSetTheorems
PermutationsFinite == TRUE
RandomSubsetFinite == TRUE
====`,
	"NaturalsInduction": `---- MODULE NaturalsInduction ----
EXTENDS Naturals
Induction == TRUE
NatInduction == TRUE
StrongInduction == TRUE
====`,
	"WellFoundedInduction": `---- MODULE WellFoundedInduction ----
WellFoundedInduction == TRUE
WellFoundedness == TRUE
====`,
	"GraphTheorems": `---- MODULE GraphTheorems ----
GraphInduction == TRUE
TransitiveClosure == TRUE
====`,
	"DyadicRationals": `---- MODULE DyadicRationals ----
IsDyadicRational(r) == \E i \in 0..r.den : 2^i = r.den
Zero == 0
One == 1
Add(x, y) == x + y
Half(x) == x
PrettyPrint(x) == x
====`,
	"Apalache": `---- MODULE Apalache ----
Gen(n) == n
FunAsSeq(f, n, m) == f
====`,
	"TLC": `---- MODULE TLC ----
LOCAL INSTANCE Naturals
LOCAL INSTANCE Sequences
LOCAL INSTANCE FiniteSets
Print(out, val) == Print(out, val)
PrintT(x) == TRUE
Assert(val, out) == Assert(val, out)
TLCGet(k) == 0
TLCSet(k, v) == TRUE
Permutations(S) == {S}
RandomElement(S) == CHOOSE x \in S : TRUE
ToString(x) == ""
TLCEval(x) == x
JavaTime == 0
====`,
	"TLAPS": `---- MODULE TLAPS ----
SimpleArithmetic == TRUE
SMT == TRUE
SMTT(X) == TRUE
CVC3 == TRUE
CVC3T(X) == TRUE
Yices == TRUE
YicesT(X) == TRUE
veriT == TRUE
veriTT(X) == TRUE
Z3 == TRUE
Z3T(X) == TRUE
Spass == TRUE
SpassT(X) == TRUE
LS4 == TRUE
PTL == TRUE
Zenon == TRUE
ZenonT(X) == TRUE
Isa == TRUE
IsaT(X) == TRUE
IsaM(X) == TRUE
IsaMT(X, Y) == TRUE
IsaWithSetExtensionality == TRUE
SetExtensionality == TRUE
NoSetContainsEverything == TRUE
SlowZenon == TRUE
SlowerZenon == TRUE
VerySlowZenon == TRUE
SlowestZenon == TRUE
Auto == TRUE
SlowAuto == TRUE
SlowerAuto == TRUE
SlowestAuto == TRUE
Force == TRUE
SlowForce == TRUE
SlowerForce == TRUE
SlowestForce == TRUE
SimplifyAndSolve == TRUE
SlowSimplifyAndSolve == TRUE
SlowerSimplifyAndSolve == TRUE
SlowestSimplifyAndSolve == TRUE
Simplification == TRUE
SlowSimplification == TRUE
SlowerSimplification == TRUE
SlowestSimplification == TRUE
Blast == TRUE
SlowBlast == TRUE
SlowerBlast == TRUE
SlowestBlast == TRUE
AutoBlast == TRUE
AllProvers == TRUE
AllProversT(X) == TRUE
AllSMT == TRUE
AllSMTT(X) == TRUE
AllIsa == TRUE
AllIsaT(X) == TRUE
RuleTLA1 == TRUE
RuleTLA2 == TRUE
RuleINV1 == TRUE
RuleINV2 == TRUE
RuleWF1 == TRUE
RuleSF1 == TRUE
RuleWF2 == TRUE
RuleSF2 == TRUE
RuleInvImplication == TRUE
RuleStepSimulation == TRUE
PropositionalTemporalLogic == TRUE
====`,
})

func loadStandardModules(fallbacks map[string]string) map[string]string {
	out := make(map[string]string, len(fallbacks)+32)
	for name, source := range fallbacks {
		out[name] = source
	}
	entries, err := embeddedJavaStandardModules.ReadDir("test_vectors/java-sany/StandardModules")
	if err != nil {
		return out
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tla") {
			continue
		}
		data, err := embeddedJavaStandardModules.ReadFile("test_vectors/java-sany/StandardModules/" + entry.Name())
		if err != nil {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".tla")
		out[name] = string(data)
	}
	return out
}
