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

// Raw code units preserve a half-surrogate consumed by the generated scanner.
func sanyLexicalErrorUnits(eof bool, line, column int, after []uint16, character uint16) string {
	encountered := "<EOF> "
	if !eof {
		encountered = fmt.Sprintf("\"%s\" (%d), ", sanyLexicalEscapeUnits([]uint16{character}), character)
	}
	return fmt.Sprintf("Lexical error at line %d, column %d.  Encountered: %safter : \"%s\"", line, column, encountered, sanyLexicalEscapeUnits(after))
}
