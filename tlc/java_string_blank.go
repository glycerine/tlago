package tlc

import "unicode"

// String.isBlank uses Character.isWhitespace. Unicode's nonbreaking spaces
// and NEL are not Java whitespace, while the ASCII information separators are.
func JavaStringIsBlank(expression string) bool {
	for _, char := range expression {
		if char >= '\t' && char <= '\r' || char >= '\u001c' && char <= '\u001f' {
			continue
		}
		if char != '\u00a0' && char != '\u2007' && char != '\u202f' && (unicode.Is(unicode.Zs, char) || unicode.Is(unicode.Zl, char) || unicode.Is(unicode.Zp, char)) {
			continue
		}
		return false
	}
	return true
}
