package tlc

import (
	"reflect"
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// SemanticSymbolMatcher adapts tla2sany.semantic.SymbolMatcher.
type SemanticSymbolMatcher interface {
	Matches(SemanticNode) bool
}

// SemanticNameAndTypeMatcher adapts SymbolMatcher.NameAndTypeMatcher. Types
// match exact concrete classes, as Class-set membership does in the source.
type SemanticNameAndTypeMatcher struct {
	prefix        *string
	Types         []reflect.Type
	CaseSensitive bool
}

func NewSemanticNameAndTypeMatcher() *SemanticNameAndTypeMatcher {
	return &SemanticNameAndTypeMatcher{}
}

func (m *SemanticNameAndTypeMatcher) SetPrefix(prefix string) *SemanticNameAndTypeMatcher {
	return m.SetPrefixNullable(&prefix)
}

// SetPrefixNullable preserves NameAndTypeMatcher.setPrefix's nullable String
// argument and fluent receiver identity.
func (m *SemanticNameAndTypeMatcher) SetPrefixNullable(prefix *string) *SemanticNameAndTypeMatcher {
	if m == nil {
		panic(NewNullPointerException())
	}
	m.prefix = prefix
	return m
}

func (m *SemanticNameAndTypeMatcher) GetPrefix() *string {
	if m == nil {
		panic(NewNullPointerException())
	}
	return m.prefix
}

func semanticContextNodeName(node SemanticNode) *UniqueString {
	switch node := node.(type) {
	case *SymbolNode:
		return node.GetName()
	case *OpDefNode:
		node.require()
		return node.Name
	case *ThmOrAssumpDefNode:
		node.require()
		return node.Name
	case *ModuleNode:
		return node.GetName()
	default:
		panic(NewClassCastException("context entry is not a SymbolNode"))
	}
}

func (m *SemanticNameAndTypeMatcher) Matches(node SemanticNode) bool {
	if m == nil || node == nil {
		panic(NewNullPointerException())
	}
	if len(m.Types) != 0 {
		actual, found := reflect.TypeOf(node), false
		for _, candidate := range m.Types {
			if candidate == actual {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	name := semanticContextNodeName(node)
	if name == nil {
		panic(NewNullPointerException())
	}
	text := name.String()
	if SemanticKindOf(node) == SemanticBuiltInKind && strings.HasPrefix(text, "$") {
		return false
	}
	if m.prefix == nil {
		panic(NewNullPointerException())
	}
	if m.CaseSensitive && !strings.HasPrefix(text, *m.prefix) {
		return false
	}
	return strings.HasPrefix(javaMatcherLower(text), javaMatcherLower(*m.prefix))
}

func javaMatcherLower(text string) string {
	languageName := strings.ToLower(tlcGetSystemProperty("user.language", "en"))
	switch languageName {
	case "iw":
		languageName = "he"
	case "ji":
		languageName = "yi"
	case "in":
		languageName = "id"
	}
	return cases.Lower(language.Make(languageName)).String(text)
}

// GetSymbols filters the context's current Hashtable bindings and sorts them
// through SemanticNode.compareTo, including its direct syntax dereference.
func (m *ModuleNode) GetSymbols(matcher SemanticSymbolMatcher) []SemanticNode {
	m.require()
	result := make([]SemanticNode, 0)
	entries := m.Context.GetContextSymbolEnumeration()
	for entries.HasMoreElements() {
		node := entries.NextElement()
		if matcher == nil {
			panic(NewNullPointerException())
		}
		if matcher.Matches(node) {
			result = append(result, node)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		return SemanticCompareTo(result[i], result[j]) < 0
	})
	return result
}
