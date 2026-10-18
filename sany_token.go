package tlago

import (
	"fmt"

	"github.com/glycerine/tlago/tlc"
)

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

const SanyTokenInvalid SanyTokenKind = -1

type SanyToken struct {
	Kind        SanyTokenKind
	Image       string
	BeginLine   int
	BeginColumn int
	EndLine     int
	EndColumn   int
	SourceFile  string
	LexState    SanyLexState
	Next        *SanyToken
	// SpecialToken points to the last preceding special token. On a special token,
	// it points backward; Next links the chronological special sequence.
	SpecialToken *SanyToken
	imageSet     bool
}

func NewSanyToken(ofKind SanyTokenKind) *SanyToken {
	_ = ofKind
	return &SanyToken{}
}

func (t *SanyToken) String() string {
	if t == nil {
		panic(tlc.NewNullPointerException())
	}
	if !t.imageSet && t.Image == "" {
		return "null"
	}
	return t.Image
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
	return SanyRange{Begin: t.BeginPosition(), End: t.EndPosition()}
}

func (t SanyToken) BeginPosition() Position {
	return Position{File: t.SourceFile, Line: t.BeginLine, Column: t.BeginColumn}
}

func (t SanyToken) EndPosition() Position {
	return Position{File: t.SourceFile, Line: t.EndLine, Column: t.EndColumn}
}
