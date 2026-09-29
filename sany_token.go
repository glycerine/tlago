package tlago

import "fmt"

//go:generate go run ./sany_generate.go

type SanyLexState int

const (
	SanyLexDefault SanyLexState = iota
	SanyLexPragma
	SanyLexSpec
	SanyLexInComment
	SanyLexEmbedded
	SanyLexInEOLComment
)

var SanyLexStateNames = []string{
	"DEFAULT",
	"PRAGMA",
	"SPEC",
	"IN_COMMENT",
	"EMBEDDED",
	"IN_EOL_COMMENT",
}

func (s SanyLexState) String() string {
	if int(s) >= 0 && int(s) < len(SanyLexStateNames) {
		return SanyLexStateNames[s]
	}
	return fmt.Sprintf("SanyLexState(%d)", int(s))
}

type SanyTokenDefinition struct {
	Kind  SanyTokenKind
	Name  string
	Image string
}

const (
	SanyTokenInvalid      SanyTokenKind = -1
	SanyTokenBlockComment SanyTokenKind = -2
	SanyTokenEOLComment   SanyTokenKind = -3
)

type SanyToken struct {
	Kind     SanyTokenKind
	Image    string
	Begin    Position
	End      Position
	LexState SanyLexState
	Next     *SanyToken
	Special  *SanyToken
}

func (k SanyTokenKind) JavaName() string {
	for _, def := range SanyTokenKinds {
		if def.Kind == k {
			return def.Name
		}
	}
	return fmt.Sprintf("token(%d)", int(k))
}

func (k SanyTokenKind) JavaImage() string {
	if int(k) >= 0 && int(k) < len(SanyTokenImages) {
		return SanyTokenImages[k]
	}
	return ""
}

func (t SanyToken) Range() SanyRange {
	return SanyRange{Begin: t.Begin, End: t.End}
}
