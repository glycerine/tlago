package tlago

import (
	"fmt"
	"strings"
	"unicode/utf16"
)

// TokenMgrError is caught by TLAplusParser.parse rather than recorded in the
// semantic Errors log. Keep its source message and native diagnostic together.
type sanyTokenMgrError struct {
	diagnostic Diagnostic
	message    string
	errorCode  int32
}

func (failure *sanyTokenMgrError) Error() string { return failure.message }

func sanyLexicalEscapes(value string) string {
	return sanyLexicalEscapeUnits(utf16.Encode([]rune(value)))
}

func sanyLexicalEscapeUnits(units []uint16) string {
	var result strings.Builder
	for _, ch := range units {
		switch ch {
		case 0:
		case '\b':
			result.WriteString(`\b`)
		case '\t':
			result.WriteString(`\t`)
		case '\n':
			result.WriteString(`\n`)
		case '\f':
			result.WriteString(`\f`)
		case '\r':
			result.WriteString(`\r`)
		case '"':
			result.WriteString(`\"`)
		case '\'':
			result.WriteString(`\'`)
		case '\\':
			result.WriteString(`\\`)
		default:
			if ch < 0x20 || ch > 0x7e {
				fmt.Fprintf(&result, `\u%04x`, ch)
			} else {
				result.WriteByte(byte(ch))
			}
		}
	}
	return result.String()
}

func sanyLexicalError(eof bool, position Position, after string, character rune) string {
	encountered := "<EOF> "
	if !eof {
		ch := utf16.Encode([]rune{character})[0]
		encountered = fmt.Sprintf("\"%s\" (%d), ", sanyLexicalEscapeUnits([]uint16{ch}), ch)
	}
	return fmt.Sprintf("Lexical error at line %d, column %d.  Encountered: %safter : \"%s\"", position.Line, position.Column, encountered, sanyLexicalEscapes(after))
}

func (tm *SanyTokenManager) lexicalFailure(begin Position, code, nativeMessage string, position Position, after string, character rune, eof bool) {
	tm.lexicalBegin = begin
	diagnostic := errorAt(begin, code, "%s", nativeMessage)
	diagnostic.SANYParseMessage = sanyLexicalError(eof, position, after, character)
	panic(&sanyTokenMgrError{diagnostic: diagnostic, message: diagnostic.SANYParseMessage})
}

func (tm *SanyTokenManager) lexicalEOFPosition() Position {
	position := tm.pos()
	if tm.column == 1 && tm.lastEnd.Line != 0 {
		position.Column = 0
	}
	return position
}
