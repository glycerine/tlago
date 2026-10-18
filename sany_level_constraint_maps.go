package tlago

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/glycerine/tlago/tlc"
)

func sanyRuntimeSymbolUID(symbol tlc.SemanticNode) int32 {
	if sanyLevelParameterSymbolIsNil(symbol) {
		panic(tlc.NewNullPointerException())
	}
	if symbol, ok := symbol.(*tlc.SymbolNode); ok {
		if symbol.SemanticBase == nil {
			panic(tlc.NewNullPointerException())
		}
		return symbol.SemanticBase.GetUID()
	}
	if symbol, ok := symbol.(interface{ GetUID() int32 }); ok {
		return symbol.GetUID()
	}
	panic(tlc.NewClassCastException("level constraint key is not a SymbolNode"))
}

func sanyRuntimeSymbolHash(symbol tlc.SemanticNode) int32 {
	if sanyLevelParameterSymbolIsNil(symbol) {
		return 0
	}
	return sanyLevelParameterSymbolHash(symbol)
}

func sanyRuntimeSymbolKind(symbol tlc.SemanticNode) tlc.SemanticKind {
	if declaration, ok := symbol.(*tlc.SymbolNode); ok {
		if declaration == nil || declaration.SemanticBase == nil {
			panic(tlc.NewNullPointerException())
		}
		return declaration.SemanticBase.Kind()
	}
	return tlc.SemanticKindOf(symbol)
}

func sanyRuntimeSymbolEqual(left, right tlc.SemanticNode) bool {
	leftNil, rightNil := sanyLevelParameterSymbolIsNil(left), sanyLevelParameterSymbolIsNil(right)
	if leftNil || rightNil {
		return leftNil && rightNil
	}
	return reflect.TypeOf(left) == reflect.TypeOf(right) && sanyRuntimeSymbolKind(left) == sanyRuntimeSymbolKind(right) &&
		sanyRuntimeSymbolUID(left) == sanyRuntimeSymbolUID(right)
}

func sanyRuntimeSymbolClassName(symbol tlc.SemanticNode) string {
	switch symbol.(type) {
	case *tlc.SymbolNode:
		if sanyRuntimeSymbolKind(symbol) == tlc.SemanticFormalParamKind {
			return "tla2sany.semantic.FormalParamNode"
		}
		return "tla2sany.semantic.OpDeclNode"
	case *tlc.OpDefNode:
		return "tla2sany.semantic.OpDefNode"
	case *tlc.ModuleNode:
		return "tla2sany.semantic.ModuleNode"
	case *tlc.ThmOrAssumpDefNode:
		return "tla2sany.semantic.ThmOrAssumpDefNode"
	default:
		panic(tlc.NewClassCastException("level constraint key is not a SymbolNode"))
	}
}

func sanyRuntimeIdentityTieBreak(left, right any) int {
	identity := func(value any) uint32 {
		if value == nil {
			return 0
		}
		return uint32(reflect.ValueOf(value).Pointer()) & 0x7fffffff
	}
	if identity(left) <= identity(right) {
		return -1
	}
	return 1
}

func sanyRuntimeSymbolTieBreak(left, right tlc.SemanticNode) int {
	if !sanyLevelParameterSymbolIsNil(left) && !sanyLevelParameterSymbolIsNil(right) {
		leftClass, rightClass := sanyRuntimeSymbolClassName(left), sanyRuntimeSymbolClassName(right)
		if leftClass < rightClass {
			return -1
		}
		if leftClass > rightClass {
			return 1
		}
	}
	return sanyRuntimeIdentityTieBreak(left, right)
}

type SanySetOfLevelConstraints struct {
	entries   *tlc.JavaSemanticMap[tlc.SemanticNode, *int32]
	owner     *sanyLevelData
	canonical *sanySetOfLevelConstraints
}

func sanyRuntimeCanonicalConstraintSymbol(symbol tlc.SemanticNode) sanySemSymbol {
	if sanyLevelParameterSymbolIsNil(symbol) {
		return nil
	}
	return sanyRuntimeCanonicalSymbol(symbol)
}

func NewSanySetOfLevelConstraints(sources ...*SanySetOfLevelConstraints) *SanySetOfLevelConstraints {
	if len(sources) > 1 {
		panic(tlc.NewIllegalArgumentException("NewSanySetOfLevelConstraints accepts at most one source"))
	}
	if len(sources) == 1 {
		if sources[0] == nil {
			panic(tlc.NewNullPointerException())
		}
		source := sources[0].require()
		if source.canonical != nil {
			entries := tlc.NewJavaSemanticMapCopyFrom(source.canonical.entries.Len(), func(yield func(tlc.SemanticNode, *int32) bool) {
				for key, value := range source.canonical.entries.All() {
					if !yield(source.owner.publicRuntimeSymbol(key), value) {
						return
					}
				}
			}, sanyRuntimeSymbolHash, sanyRuntimeSymbolEqual, sanyRuntimeSymbolTieBreak)
			return &SanySetOfLevelConstraints{entries: entries}
		}
		return &SanySetOfLevelConstraints{entries: tlc.NewJavaSemanticMapCopy(source.entries, sanyRuntimeSymbolHash, sanyRuntimeSymbolEqual, sanyRuntimeSymbolTieBreak)}
	}
	return &SanySetOfLevelConstraints{entries: tlc.NewJavaSemanticMap[tlc.SemanticNode, *int32](sanyRuntimeSymbolHash, sanyRuntimeSymbolEqual, sanyRuntimeSymbolTieBreak)}
}

func (s *SanySetOfLevelConstraints) require() *SanySetOfLevelConstraints {
	if s == nil || s.entries == nil && (s.owner == nil || s.canonical == nil) {
		panic(tlc.NewNullPointerException())
	}
	return s
}
func (s *SanySetOfLevelConstraints) Get(param tlc.SemanticNode) *int32 {
	s.require()
	if s.canonical != nil {
		return s.canonical.get(sanyRuntimeCanonicalConstraintSymbol(param))
	}
	value, _ := s.entries.Get(param)
	return value
}
func (s *SanySetOfLevelConstraints) Len() int {
	s.require()
	if s.canonical != nil {
		return s.canonical.entries.Len()
	}
	return s.entries.Len()
}
func (s *SanySetOfLevelConstraints) Put(param tlc.SemanticNode, level *int32) *int32 {
	s.require()
	if level == nil {
		panic(tlc.NewNullPointerException())
	}
	if s.canonical != nil {
		return s.canonical.put(sanyRuntimeCanonicalConstraintSymbol(param), level)
	}
	old := s.Get(param)
	oldLevel := int32(tlc.SemanticTemporalLevel)
	if old != nil {
		oldLevel = *old
	}
	value := min(*level, oldLevel)
	s.entries.Put(param, &value)
	return old
}
func (s *SanySetOfLevelConstraints) PutAll(source *SanySetOfLevelConstraints) {
	s.require()
	if source == nil {
		panic(tlc.NewNullPointerException())
	}
	if source.canonical != nil {
		for key, value := range source.canonical.entries.All() {
			s.Put(source.owner.publicRuntimeSymbol(key), value)
		}
		return
	}
	for key, value := range source.entries.All() {
		s.Put(key, value)
	}
}
func (s *SanySetOfLevelConstraints) String() string {
	s.require()
	if s.canonical != nil {
		return s.canonical.String()
	}
	var out strings.Builder
	out.WriteString("{ ")
	first := true
	for key, value := range s.entries.All() {
		if !first {
			out.WriteString(", ")
		}
		first = false
		if sanyLevelParameterSymbolIsNil(key) {
			panic(tlc.NewNullPointerException())
		}
		name := sanySymbolTableNodeName(key)
		if name == nil {
			out.WriteString("null")
		} else {
			out.WriteString(name.String())
		}
		out.WriteString(" -> ")
		if value == nil {
			out.WriteString("null")
		} else {
			out.WriteString(strconv.FormatInt(int64(*value), 10))
		}
	}
	out.WriteByte('}')
	return out.String()
}

type SanySetOfArgLevelConstraints struct {
	entries   *tlc.JavaSemanticMap[*SanyParamAndPosition, *int32]
	owner     *sanyLevelData
	canonical *sanySetOfArgLevelConstraints
}

func sanyPublicArgConstraintHash(key *SanyParamAndPosition) int32 {
	if key == nil {
		return 0
	}
	return key.HashCode()
}
func sanyPublicArgConstraintEqual(left, right *SanyParamAndPosition) bool {
	return left != nil && left.Equals(right)
}
func sanyPublicArgConstraintTieBreak(left, right *SanyParamAndPosition) int {
	return sanyRuntimeIdentityTieBreak(left, right)
}

func NewSanySetOfArgLevelConstraints(sources ...*SanySetOfArgLevelConstraints) *SanySetOfArgLevelConstraints {
	if len(sources) > 1 {
		panic(tlc.NewIllegalArgumentException("NewSanySetOfArgLevelConstraints accepts at most one source"))
	}
	if len(sources) == 1 {
		if sources[0] == nil {
			panic(tlc.NewNullPointerException())
		}
		source := sources[0].require()
		if source.canonical != nil {
			entries := tlc.NewJavaSemanticMapCopyFrom(source.canonical.entries.Len(), func(yield func(*SanyParamAndPosition, *int32) bool) {
				for key, value := range source.canonical.entries.All() {
					if key == nil {
						if !yield(nil, value) {
							return
						}
						continue
					}
					public := NewSanyParamAndPosition(source.owner.publicRuntimeSymbol(key.param), int(key.position))
					if !yield(public, value) {
						return
					}
				}
			}, sanyPublicArgConstraintHash, sanyPublicArgConstraintEqual, sanyPublicArgConstraintTieBreak)
			return &SanySetOfArgLevelConstraints{entries: entries}
		}
		return &SanySetOfArgLevelConstraints{entries: tlc.NewJavaSemanticMapCopy(source.entries, sanyPublicArgConstraintHash, sanyPublicArgConstraintEqual, sanyPublicArgConstraintTieBreak)}
	}
	return &SanySetOfArgLevelConstraints{entries: tlc.NewJavaSemanticMap[*SanyParamAndPosition, *int32](sanyPublicArgConstraintHash, sanyPublicArgConstraintEqual, sanyPublicArgConstraintTieBreak)}
}

func (s *SanySetOfArgLevelConstraints) require() *SanySetOfArgLevelConstraints {
	if s == nil || s.entries == nil && (s.owner == nil || s.canonical == nil) {
		panic(tlc.NewNullPointerException())
	}
	return s
}
func (s *SanySetOfArgLevelConstraints) Get(param *SanyParamAndPosition) *int32 {
	s.require()
	if s.canonical != nil {
		if param == nil {
			return s.canonical.get(nil)
		}
		return s.canonical.get(newSanyParamAndPosition(sanyRuntimeCanonicalSymbol(param.param), param.position))
	}
	value, _ := s.entries.Get(param)
	return value
}
func (s *SanySetOfArgLevelConstraints) Len() int {
	s.require()
	if s.canonical != nil {
		return s.canonical.entries.Len()
	}
	return s.entries.Len()
}
func (s *SanySetOfArgLevelConstraints) Put(param *SanyParamAndPosition, level *int32) *int32 {
	s.require()
	if level == nil {
		panic(tlc.NewNullPointerException())
	}
	if s.canonical != nil {
		var canonical *sanyParamAndPosition
		if param != nil {
			canonical = newSanyParamAndPosition(sanyRuntimeCanonicalSymbol(param.param), param.position)
		}
		return s.canonical.put(canonical, level)
	}
	old := s.Get(param)
	oldLevel := int32(tlc.SemanticConstantLevel)
	if old != nil {
		oldLevel = *old
	}
	value := max(*level, oldLevel)
	s.entries.Put(param, &value)
	return old
}
func (s *SanySetOfArgLevelConstraints) PutArgument(param tlc.SemanticNode, position, level int) *int32 {
	value := int32(level)
	return s.Put(NewSanyParamAndPosition(param, position), &value)
}
func (s *SanySetOfArgLevelConstraints) PutAll(source *SanySetOfArgLevelConstraints) {
	s.require()
	if source == nil {
		panic(tlc.NewNullPointerException())
	}
	if source.canonical != nil {
		for key, value := range source.canonical.entries.All() {
			var public *SanyParamAndPosition
			if key != nil {
				public = NewSanyParamAndPosition(source.owner.publicRuntimeSymbol(key.param), int(key.position))
			}
			s.Put(public, value)
		}
		return
	}
	for key, value := range source.entries.All() {
		s.Put(key, value)
	}
}
func (s *SanySetOfArgLevelConstraints) String() string {
	s.require()
	if s.canonical != nil {
		return s.canonical.String()
	}
	var out strings.Builder
	out.WriteString("{ ")
	first := true
	for key, value := range s.entries.All() {
		if !first {
			out.WriteString(", ")
		}
		first = false
		if key == nil {
			panic(tlc.NewNullPointerException())
		}
		out.WriteString(key.String())
		out.WriteString(" -> ")
		if value == nil {
			out.WriteString("null")
		} else {
			out.WriteString(strconv.FormatInt(int64(*value), 10))
		}
	}
	out.WriteByte('}')
	return out.String()
}
