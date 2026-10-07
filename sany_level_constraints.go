// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/glycerine/tlago/tlc"
)

type sanySetOfLevelConstraints struct {
	entries *tlc.JavaSemanticMap[sanySemSymbol, *int32]
}

func newSanySetOfLevelConstraints() *sanySetOfLevelConstraints {
	return &sanySetOfLevelConstraints{tlc.NewJavaSemanticMap[sanySemSymbol, *int32](sanyLevelConstraintHash, sanyLevelConstraintEqual, sanyLevelConstraintTieBreak)}
}
func newSanySetOfLevelConstraintsFrom(source *tlc.JavaSemanticMap[sanySemSymbol, *int32]) *sanySetOfLevelConstraints {
	return &sanySetOfLevelConstraints{tlc.NewJavaSemanticMapCopy(source, sanyLevelConstraintHash, sanyLevelConstraintEqual, sanyLevelConstraintTieBreak)}
}
func (s *sanySetOfLevelConstraints) get(param sanySemSymbol) *int32 {
	if s == nil {
		panic(tlc.NewNullPointerException())
	}
	value, _ := s.entries.Get(sanyLevelSymbolReference(param))
	return value
}
func (s *sanySetOfLevelConstraints) put(param sanySemSymbol, level *int32) *int32 {
	if s == nil || level == nil {
		panic(tlc.NewNullPointerException())
	}
	newLevel := *level
	param = sanyLevelSymbolReference(param)
	old := s.get(param)
	oldLevel := int32(3)
	if old != nil {
		oldLevel = *old
	}
	value := min(newLevel, oldLevel)
	s.entries.Put(param, &value)
	return old
}
func (s *sanySetOfLevelConstraints) putAll(source *tlc.JavaSemanticMap[sanySemSymbol, *int32]) {
	if s == nil || source == nil {
		panic(tlc.NewNullPointerException())
	}
	for key := range source.All() {
		value, _ := source.Get(key)
		s.put(key, value)
	}
}
func (s *sanySetOfLevelConstraints) String() string {
	if s == nil {
		panic(tlc.NewNullPointerException())
	}
	var out strings.Builder
	out.WriteString("{ ")
	first := true
	for param := range s.entries.All() {
		param = sanyLevelSymbolReference(param)
		if param == nil {
			panic(tlc.NewNullPointerException())
		}
		if !first {
			out.WriteString(", ")
		}
		first = false
		if formal, ok := param.(*sanyFormalParamNode); ok && formal.nullName {
			out.WriteString("null")
		} else {
			out.WriteString(param.semName())
		}
		out.WriteString(" -> ")
		out.WriteString(sanyLevelIntegerString(s.get(param)))
	}
	out.WriteString("}")
	return out.String()
}

type sanySetOfArgLevelConstraints struct {
	entries *tlc.JavaSemanticMap[*sanyParamAndPosition, *int32]
}

func newSanySetOfArgLevelConstraints() *sanySetOfArgLevelConstraints {
	return &sanySetOfArgLevelConstraints{tlc.NewJavaSemanticMap[*sanyParamAndPosition, *int32](sanyArgConstraintHash, sanyArgConstraintEqual, sanyArgConstraintTieBreak)}
}
func newSanySetOfArgLevelConstraintsFrom(source *tlc.JavaSemanticMap[*sanyParamAndPosition, *int32]) *sanySetOfArgLevelConstraints {
	return &sanySetOfArgLevelConstraints{tlc.NewJavaSemanticMapCopy(source, sanyArgConstraintHash, sanyArgConstraintEqual, sanyArgConstraintTieBreak)}
}
func (s *sanySetOfArgLevelConstraints) get(param *sanyParamAndPosition) *int32 {
	if s == nil {
		panic(tlc.NewNullPointerException())
	}
	value, _ := s.entries.Get(param)
	return value
}
func (s *sanySetOfArgLevelConstraints) put(param *sanyParamAndPosition, level *int32) *int32 {
	if s == nil || level == nil {
		panic(tlc.NewNullPointerException())
	}
	newLevel := *level
	old := s.get(param)
	oldLevel := int32(0)
	if old != nil {
		oldLevel = *old
	}
	value := max(newLevel, oldLevel)
	s.entries.Put(param, &value)
	return old
}
func (s *sanySetOfArgLevelConstraints) putArgument(param sanySemSymbol, position, level int32) *int32 {
	return s.put(newSanyParamAndPosition(param, position), &level)
}
func (s *sanySetOfArgLevelConstraints) putAll(source *tlc.JavaSemanticMap[*sanyParamAndPosition, *int32]) {
	if s == nil || source == nil {
		panic(tlc.NewNullPointerException())
	}
	for key := range source.All() {
		value, _ := source.Get(key)
		s.put(key, value)
	}
}
func (s *sanySetOfArgLevelConstraints) String() string {
	if s == nil {
		panic(tlc.NewNullPointerException())
	}
	var out strings.Builder
	out.WriteString("{ ")
	first := true
	for param := range s.entries.All() {
		if !first {
			out.WriteString(", ")
		}
		first = false
		out.WriteString(param.String())
		out.WriteString(" -> ")
		out.WriteString(sanyLevelIntegerString(s.get(param)))
	}
	out.WriteString("}")
	return out.String()
}

func sanyLevelIntegerString(value *int32) string {
	if value == nil {
		return "null"
	}
	return strconv.FormatInt(int64(*value), 10)
}
func sanyLevelConstraintHash(symbol sanySemSymbol) int32 {
	if sanyLevelSymbolReference(symbol) == nil {
		return 0
	}
	return sanyLevelSymbolHash(symbol)
}
func sanyLevelConstraintEqual(a, b sanySemSymbol) bool {
	a, b = sanyLevelSymbolReference(a), sanyLevelSymbolReference(b)
	if a == nil || b == nil {
		return a == b
	}
	return reflect.TypeOf(a) == reflect.TypeOf(b) && a.semKind() == b.semKind() && a.semBase().getUID() == b.semBase().getUID()
}
func sanyArgConstraintHash(key *sanyParamAndPosition) int32 {
	if key == nil {
		return 0
	}
	return key.hashCode()
}
func sanyArgConstraintEqual(a, b *sanyParamAndPosition) bool {
	return a != nil && a.equals(b)
}
func sanyLevelConstraintTieBreak(a, b sanySemSymbol) int {
	// HashMap orders different non-Comparable classes by their Java class name.
	a, b = sanyLevelSymbolReference(a), sanyLevelSymbolReference(b)
	if a != nil && b != nil {
		an, bn := sanyLevelSymbolClassName(a), sanyLevelSymbolClassName(b)
		if an < bn {
			return -1
		}
		if an > bn {
			return 1
		}
	}
	return sanyLevelIdentityTieBreak(a, b)
}
func sanyArgConstraintTieBreak(a, b *sanyParamAndPosition) int {
	return sanyLevelIdentityTieBreak(a, b)
}
func sanyLevelIdentityTieBreak(a, b any) int {
	// Native object addresses supply the identity tie-break, as in other ports.
	identity := func(value any) uint32 {
		if value == nil {
			return 0
		}
		return uint32(reflect.ValueOf(value).Pointer()) & 0x7fffffff
	}
	ah, bh := identity(a), identity(b)
	if ah <= bh {
		return -1
	}
	return 1
}
func sanyLevelSymbolClassName(symbol sanySemSymbol) string {
	switch symbol.(type) {
	case *sanySemOpDeclNode:
		return "tla2sany.semantic.OpDeclNode"
	case *sanySemOpDefNode:
		return "tla2sany.semantic.OpDefNode"
	case *sanySemModuleNode:
		return "tla2sany.semantic.ModuleNode"
	case *sanySemThmOrAssumpDefNode:
		return "tla2sany.semantic.ThmOrAssumpDefNode"
	case *sanyFormalParamNode:
		return "tla2sany.semantic.FormalParamNode"
	default:
		panic(tlc.NewClassCastException("not a canonical SymbolNode"))
	}
}
