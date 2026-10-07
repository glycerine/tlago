// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/glycerine/tlago/tlc"
)

// LevelNode's mutable collections belong to each canonical node. The native
// analyzer's integer parameter IDs are not substitutes for SymbolNode keys.
type sanyLevelData struct {
	levelCorrect        bool
	level               tlaLevel
	levelParams         *sanyLevelSet[sanySemSymbol]
	levelConstraints    *sanySetOfLevelConstraints
	argLevelConstraints *sanySetOfArgLevelConstraints
	argLevelParams      *sanyLevelSet[*sanyArgLevelParam]
	allParams           *sanyLevelSet[sanySemSymbol]
	nonLeibnizParams    *sanyLevelSet[sanySemSymbol]
	levelChecked        int32
}

func newSanyLevelData() sanyLevelData {
	return sanyLevelData{
		levelCorrect:        true,
		levelParams:         newSanyLevelSymbolSet(),
		levelConstraints:    newSanySetOfLevelConstraints(),
		argLevelConstraints: newSanySetOfArgLevelConstraints(),
		argLevelParams:      newSanyArgLevelSet(),
		allParams:           newSanyLevelSymbolSet(),
		nonLeibnizParams:    newSanyLevelSymbolSet(),
	}
}

// This interface covers the canonical SANY classes. TLC-owned literal nodes
// still require integration with this shared metadata before mixed graphs check.
type sanyLevelAccess interface {
	getLevel() tlaLevel
	getLevelParams() *sanyLevelSet[sanySemSymbol]
	getAllParams() *sanyLevelSet[sanySemSymbol]
	getNonLeibnizParams() *sanyLevelSet[sanySemSymbol]
	getLevelConstraints() *sanySetOfLevelConstraints
	getArgLevelConstraints() *sanySetOfArgLevelConstraints
	getArgLevelParams() *sanyLevelSet[*sanyArgLevelParam]
}
type sanyCanonicalLevelNode interface {
	sanyLevelAccess
	getKind() sanySemKind
	getLevelData() *sanyLevelData
	levelCheck(int32, *Diagnostics) bool
}

func (n *sanyLevelData) getLevelData() *sanyLevelData {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	return n
}
func sanyLevelCheckNext(node sanyCanonicalLevelNode, errors *Diagnostics) bool {
	if node == nil || reflect.ValueOf(node).IsNil() {
		panic(tlc.NewNullPointerException())
	}
	return node.levelCheck(node.getLevelData().levelChecked+1, errors)
}
func (n *sanySemanticNode) levelCheck(iter int32, errors *Diagnostics) bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	kind := int(n.getKind())
	if kind < 0 || kind >= len(sanySemanticKindNames) {
		panic(tlc.NewArrayIndexOutOfBoundsException(kind, len(sanySemanticKindNames)))
	}
	panic(tlc.NewWrongInvocationException("Level checking of " + sanySemanticKindNames[kind] + " node not implemented."))
}
func (n *sanySemanticNode) levelCheckSubnodes(iter int32, sub []sanyCanonicalLevelNode, errors *Diagnostics) bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	if n.levelChecked >= iter {
		return n.levelCorrect
	}
	n.levelChecked = iter
	if sub == nil {
		panic(tlc.NewNullPointerException())
	}
	for _, child := range sub {
		if child == nil || reflect.ValueOf(child).IsNil() {
			panic(tlc.NewNullPointerException())
		}
		if child.getKind() != sanyModuleKind && child.getKind() != sanyModuleInstanceKind {
			// Evaluate the child even after an earlier child has returned false.
			correct := child.levelCheck(iter, errors)
			n.levelCorrect = correct && n.levelCorrect
			n.level = max(n.level, child.getLevel())
			n.levelParams.addAll(child.getLevelParams())
			constraints := child.getLevelConstraints()
			var constraintMap *tlc.JavaSemanticMap[sanySemSymbol, *int32]
			if constraints != nil {
				constraintMap = constraints.entries
			}
			n.levelConstraints.putAll(constraintMap)
			argConstraints := child.getArgLevelConstraints()
			var argConstraintMap *tlc.JavaSemanticMap[*sanyParamAndPosition, *int32]
			if argConstraints != nil {
				argConstraintMap = argConstraints.entries
			}
			n.argLevelConstraints.putAll(argConstraintMap)
			n.argLevelParams.addAll(child.getArgLevelParams())
			n.allParams.addAll(child.getAllParams())
			n.nonLeibnizParams.addAll(child.getNonLeibnizParams())
		}
	}
	return n.levelCorrect
}

func (n *sanyLevelData) requireChecked(message string) {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	if n.levelChecked == 0 {
		panic(tlc.NewWrongInvocationException(message))
	}
}
func (n *sanyLevelData) getLevel() tlaLevel {
	n.requireChecked("getLevel called before levelCheck")
	return n.level
}
func (n *sanyLevelData) getLevelParams() *sanyLevelSet[sanySemSymbol] {
	n.requireChecked("getLevelParams called before levelCheck")
	return n.levelParams
}
func (n *sanyLevelData) getAllParams() *sanyLevelSet[sanySemSymbol] {
	n.requireChecked("getAllParams called before levelCheck")
	return n.allParams
}
func (n *sanyLevelData) getNonLeibnizParams() *sanyLevelSet[sanySemSymbol] {
	// The source intentionally has the same error message as getAllParams.
	n.requireChecked("getAllParams called before levelCheck")
	return n.nonLeibnizParams
}
func (n *sanyLevelData) getLevelConstraints() *sanySetOfLevelConstraints {
	n.requireChecked("getLevelConstraints called before levelCheck")
	return n.levelConstraints
}
func (n *sanyLevelData) getArgLevelConstraints() *sanySetOfArgLevelConstraints {
	n.requireChecked("getArgLevelConstraints called before levelCheck")
	return n.argLevelConstraints
}
func (n *sanyLevelData) getArgLevelParams() *sanyLevelSet[*sanyArgLevelParam] {
	n.requireChecked("getArgLevelParams called before levelCheck")
	return n.argLevelParams
}
func sanyDefaultLevelDataToString(n sanyLevelAccess) string {
	if n == nil || reflect.ValueOf(n).IsNil() {
		panic(tlc.NewNullPointerException())
	}
	level := n.getLevel()
	params := n.getLevelParams()
	constraints := n.getLevelConstraints()
	argConstraints := n.getArgLevelConstraints()
	lp, lc, ac := "null", "null", "null"
	if params != nil {
		lp = sanyLevelSymbolSetString(params)
	}
	if constraints != nil {
		lc = constraints.String()
	}
	if argConstraints != nil {
		ac = argConstraints.String()
	}
	return "Level:" + strconv.Itoa(int(level)) + "\n" +
		"LevelParams: " + lp + "\n" +
		"LevelConstraints: " + lc + "\n" +
		"ArgLevelConstraints: " + ac + "\n" +
		"ArgLevelParams: " + sanyALPHashSetToString(n.getArgLevelParams()) + "\n" +
		"AllParams: " + sanyHashSetToString(n.getAllParams()) + "\n" +
		"NonLeibnizParams: " + sanyHashSetToString(n.getNonLeibnizParams())
}
func sanyLevelDataToString(n sanyCanonicalLevelNode) string {
	if n == nil || reflect.ValueOf(n).IsNil() {
		panic(tlc.NewNullPointerException())
	}
	n.getLevelData().requireChecked("levelDataToString called before levelCheck")
	return sanyDefaultLevelDataToString(n)
}

func sanyAddTemporalLevelConstraintToConstants(params *sanyLevelSet[sanySemSymbol], constraints *sanySetOfLevelConstraints) {
	for node := range params.all() {
		node = sanyLevelSymbolReference(node)
		if node == nil {
			panic(tlc.NewNullPointerException())
		}
		if node.semKind() == sanyConstantDeclKind {
			action := int32(2)
			constraints.put(node, &action)
		}
	}
}
func sanyHashSetToString(set *sanyLevelSet[sanySemSymbol]) string {
	var out strings.Builder
	out.WriteString("{")
	first := true
	for symbol := range set.all() {
		if !first {
			out.WriteString(", ")
		}
		first = false
		out.WriteString(sanyLevelSymbolName(symbol))
	}
	out.WriteString("}")
	return out.String()
}
func sanyALPHashSetToString(set *sanyLevelSet[*sanyArgLevelParam]) string {
	var out strings.Builder
	out.WriteString("{")
	first := true
	for param := range set.all() {
		if param == nil {
			panic(tlc.NewNullPointerException())
		}
		if !first {
			out.WriteString(", ")
		}
		first = false
		out.WriteString("<" + sanyLevelSymbolName(param.op) + ", " + strconv.FormatInt(int64(param.i), 10) + ", " + sanyLevelSymbolName(param.param) + ">")
	}
	out.WriteString("}")
	return out.String()
}
func sanyLevelSymbolSetString(set *sanyLevelSet[sanySemSymbol]) string {
	var out strings.Builder
	out.WriteString("[")
	first := true
	for symbol := range set.all() {
		if !first {
			out.WriteString(", ")
		}
		first = false
		out.WriteString(sanyLevelSymbolString(symbol))
	}
	out.WriteString("]")
	return out.String()
}
func sanyLevelSymbolName(symbol sanySemSymbol) string {
	symbol = sanyLevelSymbolReference(symbol)
	if symbol == nil {
		panic(tlc.NewNullPointerException())
	}
	if formal, ok := symbol.(*sanyFormalParamNode); ok && formal.nullName {
		return "null"
	}
	return symbol.semName()
}

// Only the HashSet operations used by canonical level checking are exposed.
// The backing map retains Java equality, bucket order and fail-fast iteration.
type sanyLevelSet[K comparable] struct {
	entries *tlc.JavaSemanticMap[K, struct{}]
}

func newSanyLevelSymbolSet() *sanyLevelSet[sanySemSymbol] {
	return &sanyLevelSet[sanySemSymbol]{tlc.NewJavaSemanticMap[sanySemSymbol, struct{}](sanyLevelConstraintHash, sanyLevelConstraintEqual, sanyLevelConstraintTieBreak)}
}
func newSanyArgLevelSet() *sanyLevelSet[*sanyArgLevelParam] {
	return &sanyLevelSet[*sanyArgLevelParam]{tlc.NewJavaSemanticMap[*sanyArgLevelParam, struct{}](func(key *sanyArgLevelParam) int32 {
		if key == nil {
			return 0
		}
		return key.hashCode()
	}, func(a, b *sanyArgLevelParam) bool { return a != nil && a.equals(b) }, func(a, b *sanyArgLevelParam) int { return sanyLevelIdentityTieBreak(a, b) })}
}
func (s *sanyLevelSet[K]) add(key K) bool {
	if s == nil {
		panic(tlc.NewNullPointerException())
	}
	_, present := s.entries.Put(key, struct{}{})
	return !present
}
func (s *sanyLevelSet[K]) contains(key K) bool {
	if s == nil {
		panic(tlc.NewNullPointerException())
	}
	_, present := s.entries.Get(key)
	return present
}
func (s *sanyLevelSet[K]) remove(key K) bool {
	if s == nil {
		panic(tlc.NewNullPointerException())
	}
	return s.entries.Remove(key)
}
func (s *sanyLevelSet[K]) len() int {
	if s == nil {
		panic(tlc.NewNullPointerException())
	}
	return s.entries.Len()
}
func (s *sanyLevelSet[K]) clear() {
	if s == nil {
		panic(tlc.NewNullPointerException())
	}
	s.entries.Clear()
}
func (s *sanyLevelSet[K]) all() func(func(K) bool) {
	if s == nil {
		panic(tlc.NewNullPointerException())
	}
	return func(yield func(K) bool) {
		for key := range s.entries.All() {
			if !yield(key) {
				return
			}
		}
	}
}
func (s *sanyLevelSet[K]) addAll(source *sanyLevelSet[K]) {
	if s == nil {
		panic(tlc.NewNullPointerException())
	}
	for key := range source.all() {
		s.add(key)
	}
}

// ASTConstants.kinds, including the source null entry at kind zero.
var sanySemanticKindNames = [...]string{
	"null",
	"ModuleKind",
	"ConstantDeclKind",
	"VariableDeclKind",
	"BoundSymbolKind",
	"UserDefinedOpKind",
	"ModuleInstanceKind",
	"BuiltInKind",
	"OpArgKind",
	"OpApplKind",
	"LetInKind",
	"FormalParamKind",
	"TheoremKind",
	"SubstInKind",
	"AssumeProveKind",
	"ProofKind",
	"NumeralKind",
	"DecimalKind",
	"StringKind",
	"AtNodeKind",
	"AssumeKind",
	"InstanceKind",
	"NewSymbKind",
	"ThmOrAssumpDefKind",
	"NewConstantKind",
	"NewVariableKind",
	"NewStateKind",
	"NewActionKind",
	"NewTemporalKind",
	"LabelKind",
	"APSubstInKind",
	"UseKind",
	"HideKind",
	"LeafProofKind",
	"NonLeafProofKind",
	"QEDStepKind",
	"DefStepKind",
	"NumberedProofStepKind",
}

func (n *sanySemOpDeclNode) levelCheck(iter int32, errors *Diagnostics) bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	// OpDeclNode's constructor already supplies the level data. The source
	// method returns true without changing levelChecked, even for a new iter.
	return true
}
