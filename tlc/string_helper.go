package tlc

import (
	"regexp"
	"runtime"
	"strings"
	"sync"
	"unicode"
)

var stringHelperStatics struct {
	once sync.Once
}

var StringHelperPlatformNewline = stringHelperNativeNewline()

func stringHelperNativeNewline() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}

// InitializeStringHelperStatics represents a fresh Java classloader for
// StringHelper's mutable platform-newline field. It requires idle callers.
func InitializeStringHelperStatics() {
	stringHelperStatics.once = sync.Once{}
	StringHelperPlatformNewline = stringHelperNativeNewline()
}

func initializeStringHelperStatics() {
	stringHelperStatics.once.Do(func() {
		StringHelperPlatformNewline = stringHelperNativeNewline()
		if value, ok := tlcLookupSystemProperty("line.separator"); ok {
			StringHelperPlatformNewline = value
		}
	})
}

func stringHelperPlatformNewline() string {
	initializeStringHelperStatics()
	return StringHelperPlatformNewline
}

func stringHelperText(text *string) string {
	initializeStringHelperStatics()
	if text == nil {
		panic(NewNullPointerException())
	}
	return *text
}

// Character.isWhitespace(char), including Java's C0 separators and excluding
// the nonbreaking spaces and NEL accepted by Unicode's broader White_Space.
func stringHelperWhitespace(char uint16) bool {
	r := rune(char)
	if r >= 0x1c && r <= 0x1f {
		return true
	}
	if r == 0x85 || r == 0xa0 || r == 0x2007 || r == 0x202f {
		return false
	}
	return unicode.IsSpace(r)
}

func StringHelperCopyString(text *string, copies int32) string {
	initializeStringHelperStatics()
	result := ""
	power := text
	remaining := copies
	value := func(text *string) string {
		if text == nil {
			return "null"
		}
		return *text
	}
	for remaining > 0 {
		if remaining%2 != 0 {
			result += value(power)
		}
		remaining /= 2
		if remaining != 0 {
			next := value(power) + value(power)
			power = &next
		}
	}
	return result
}
func StringHelperOnlySpaces(text *string) bool {
	return strings.TrimFunc(stringHelperText(text), func(r rune) bool { return r <= 0x20 }) == ""
}
func StringHelperTrimFront(text *string) string {
	units := javaStringUTF16(stringHelperText(text))
	position := 0
	for position < len(units) && stringHelperWhitespace(units[position]) {
		position++
	}
	return javaStringFromUTF16(units[position:])
}
func StringHelperTrimEnd(text *string) string {
	units := javaStringUTF16(stringHelperText(text))
	position := len(units)
	for position > 0 && stringHelperWhitespace(units[position-1]) {
		position--
	}
	return javaStringFromUTF16(units[:position])
}
func StringHelperLeadingSpaces(text *string) int {
	return len(javaStringUTF16(stringHelperText(text))) - len(javaStringUTF16(StringHelperTrimFront(text)))
}

var stringHelperWordSeparator = regexp.MustCompile(`[ \t\n\x0B\f\r]+`)

func StringHelperGetWords(text *string) []string {
	trimmed := StringHelperTrimFront(text)
	words := stringHelperWordSeparator.Split(trimmed, -1)
	if trimmed != "" {
		for len(words) > 0 && words[len(words)-1] == "" {
			words = words[:len(words)-1]
		}
	}
	return words
}
func StringHelperIsIdentifier(text *string) bool {
	units := javaStringUTF16(stringHelperText(text))
	result, allDigits := true, true
	i := 0
	for result && i < len(units) {
		char := rune(units[i])
		digit := unicode.IsDigit(char)
		result = unicode.IsLetter(char) || digit || char == '_'
		allDigits = allDigits && digit
		i++
	}
	return result && !allDigits
}
