package tlago

import "github.com/glycerine/tlago/tlc"

// SanySymbolMatcher adapts tla2sany.semantic.SymbolMatcher.
type SanySymbolMatcher = tlc.SemanticSymbolMatcher

// SanyNameAndTypeMatcher adapts SymbolMatcher.NameAndTypeMatcher.
type SanyNameAndTypeMatcher = tlc.SemanticNameAndTypeMatcher

func NewSanyNameAndTypeMatcher() *SanyNameAndTypeMatcher {
	return tlc.NewSemanticNameAndTypeMatcher()
}
