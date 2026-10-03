// String-argument formatting used by RecordValue.PrintTLCState and CSV.write.
// Ports the applicable OpenJDK java.util.Formatter control flow; OpenJDK's
// GPLv2 with Classpath Exception notices are retained in tlc/licenses/.
package tlc

import (
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	formatLeft = 1 << iota
	formatUpper
	formatAlternate
	formatPlus
	formatSpace
	formatZero
	formatGroup
	formatParentheses
	formatPrevious
)

var javaFormatPattern = regexp.MustCompile(`^%([0-9]+\$)?([-#+ 0,(<]*)([0-9]+)?(\.[0-9]+)?([tT])?([a-zA-Z%])`)

type javaStringFormatSpecifier struct {
	literal                        string
	index, flags, width, precision int
	dateTime                       bool
	conversion                     byte
}

// JavaFormatStrings implements String.format for the string arguments supplied
// by TLC's record-state and CSV rendering. Numeric, character and date/time
// conversions retain Java's invalid-string-argument failures. UTF-16 truncation
// retains unpaired surrogates in WTF-8 rather than replacing their identity.
func JavaFormatStrings(format string, arguments ...string) (string, error) {
	locale := javaFormatDefaultLocale()
	lineSeparator := javaFormatLineSeparator()
	specifiers, err := parseJavaStringFormat(format)
	if err != nil {
		return "", err
	}
	output := []uint16{}
	last, ordinary := -1, -1
	for _, spec := range specifiers {
		if spec.conversion == 0 {
			output = append(output, javaStringUTF16(spec.literal)...)
			continue
		}
		if spec.index == -2 {
			text := "%"
			if spec.conversion == 'n' {
				text = lineSeparator
			}
			output = append(output, spec.justify(javaStringUTF16(text))...)
			continue
		}
		switch spec.index {
		case -1:
		case 0:
			ordinary++
			last = ordinary
		default:
			last = spec.index - 1
		}
		if last < 0 || last >= len(arguments) {
			return "", newMissingFormatArgumentException(spec.String())
		}
		value := arguments[last]
		if spec.dateTime {
			return "", newIllegalFormatConversionException(spec.conversion)
		}
		switch spec.conversion {
		case 's':
			if spec.flags&formatAlternate != 0 {
				return "", newFormatFlagsConversionMismatchException("#", 's')
			}
		case 'b':
			value = "true"
		case 'h':
			value = strconv.FormatUint(uint64(uint32(javaFormatStringHash(value))), 16)
		default:
			return "", newIllegalFormatConversionException(spec.conversion)
		}
		units := javaStringUTF16(value)
		if spec.precision >= 0 && spec.precision < len(units) {
			units = units[:spec.precision]
		}
		if spec.flags&formatUpper != 0 {
			units = javaFormatUpper(units, locale)
		}
		output = append(output, spec.justify(units)...)
	}
	return javaStringFromUTF16(output), nil
}

func parseJavaStringFormat(format string) ([]javaStringFormatSpecifier, error) {
	var specs []javaStringFormatSpecifier
	for position := 0; position < len(format); {
		offset := strings.IndexByte(format[position:], '%')
		if offset < 0 {
			specs = append(specs, javaStringFormatSpecifier{literal: format[position:]})
			break
		}
		offset += position
		if offset > position {
			specs = append(specs, javaStringFormatSpecifier{literal: format[position:offset]})
		}
		if offset+1 == len(format) {
			return nil, newUnknownFormatConversionException("%")
		}
		match := javaFormatPattern.FindStringSubmatch(format[offset:])
		if match == nil {
			char := javaStringUTF16(format[offset+1:])[0]
			return nil, newUnknownFormatConversionException(javaStringFromUTF16([]uint16{char}))
		}
		spec := javaStringFormatSpecifier{width: -1, precision: -1}
		if match[1] != "" {
			number, err := strconv.ParseInt(strings.TrimSuffix(match[1], "$"), 10, 32)
			if err != nil {
				return nil, newIllegalFormatArgumentIndexException(-1 << 31)
			}
			spec.index = int(number)
			if spec.index <= 0 {
				return nil, newIllegalFormatArgumentIndexException(spec.index)
			}
		}
		for _, flag := range match[2] {
			bit := javaFormatFlag(byte(flag))
			if spec.flags&bit != 0 {
				return nil, newDuplicateFormatFlagsException(javaFormatFlags(bit))
			}
			spec.flags |= bit
		}
		if spec.flags&formatPrevious != 0 {
			spec.index = -1
		}
		if match[3] != "" {
			number, err := strconv.ParseInt(match[3], 10, 32)
			if err != nil {
				return nil, newIllegalFormatWidthException(-1 << 31)
			}
			spec.width = int(number)
		}
		if match[4] != "" {
			number, err := strconv.ParseInt(match[4][1:], 10, 32)
			if err != nil {
				return nil, newIllegalFormatPrecisionException(-1 << 31)
			}
			spec.precision = int(number)
		}
		spec.dateTime = match[5] != ""
		spec.conversion = match[6][0]
		if spec.dateTime {
			if match[5] == "T" {
				spec.flags |= formatUpper
			}
		} else {
			if !strings.ContainsRune("bBhHsScCdoxXeEfgGaAn%", rune(spec.conversion)) {
				return nil, newUnknownFormatConversionException(string(spec.conversion))
			}
			if spec.conversion >= 'A' && spec.conversion <= 'Z' {
				spec.flags |= formatUpper
				spec.conversion += 'a' - 'A'
			}
			if spec.conversion == 'n' || spec.conversion == '%' {
				spec.index = -2
			}
		}
		if err := spec.check(); err != nil {
			return nil, err
		}
		specs = append(specs, spec)
		position = offset + len(match[0])
	}
	return specs, nil
}

func (s javaStringFormatSpecifier) check() error {
	missingWidth := func() error {
		if s.width == -1 && s.flags&formatLeft != 0 {
			return newMissingFormatWidthException(s.String())
		}
		return nil
	}
	precision := func() error {
		if s.precision != -1 {
			return newIllegalFormatPrecisionException(s.precision)
		}
		return nil
	}
	badFlags := func(mask int) error {
		if s.flags&mask != 0 {
			return newFormatFlagsConversionMismatchException(javaFormatFlags(s.flags&mask), s.conversion)
		}
		return nil
	}
	if s.dateTime {
		if err := precision(); err != nil {
			return err
		}
		if !strings.ContainsRune("HIklMSLNpzZsQBbhAaCYyjmdeRTrDFc", rune(s.conversion)) {
			return newUnknownFormatConversionException("t" + string(s.conversion))
		}
		if err := badFlags(formatAlternate | formatPlus | formatSpace | formatZero | formatGroup | formatParentheses); err != nil {
			return err
		}
		return missingWidth()
	}
	switch s.conversion {
	case 'b', 'h', 's':
		if s.conversion != 's' {
			if err := badFlags(formatAlternate); err != nil {
				return err
			}
		}
		if err := missingWidth(); err != nil {
			return err
		}
		return badFlags(formatPlus | formatSpace | formatZero | formatGroup | formatParentheses)
	case 'c':
		if err := precision(); err != nil {
			return err
		}
		if err := badFlags(formatAlternate | formatPlus | formatSpace | formatZero | formatGroup | formatParentheses); err != nil {
			return err
		}
		return missingWidth()
	case 'd', 'o', 'x', 'e', 'f', 'g', 'a':
		if s.width == -1 && s.flags&(formatLeft|formatZero) != 0 {
			return newMissingFormatWidthException(s.String())
		}
		if s.flags&(formatPlus|formatSpace) == formatPlus|formatSpace || s.flags&(formatLeft|formatZero) == formatLeft|formatZero {
			return newIllegalFormatFlagsException(javaFormatFlags(s.flags))
		}
		switch s.conversion {
		case 'd', 'o', 'x':
			if err := precision(); err != nil {
				return err
			}
			if s.conversion == 'd' {
				return badFlags(formatAlternate)
			}
			return badFlags(formatGroup)
		case 'a':
			return badFlags(formatGroup | formatParentheses)
		case 'e':
			return badFlags(formatGroup)
		case 'g':
			return badFlags(formatAlternate)
		}
	case 'n', '%':
		if err := precision(); err != nil {
			return err
		}
		if s.conversion == 'n' {
			if s.width != -1 {
				return newIllegalFormatWidthException(s.width)
			}
			if s.flags != 0 {
				return newIllegalFormatFlagsException(javaFormatFlags(s.flags))
			}
		} else {
			if s.flags != 0 && s.flags != formatLeft {
				return newIllegalFormatFlagsException(javaFormatFlags(s.flags))
			}
			return missingWidth()
		}
	}
	return nil
}

func (s javaStringFormatSpecifier) String() string {
	var result strings.Builder
	result.WriteByte('%')
	result.WriteString(javaFormatFlags(s.flags &^ formatUpper))
	if s.index > 0 {
		result.WriteString(strconv.Itoa(s.index))
		result.WriteByte('$')
	}
	if s.width != -1 {
		result.WriteString(strconv.Itoa(s.width))
	}
	if s.precision != -1 {
		result.WriteByte('.')
		result.WriteString(strconv.Itoa(s.precision))
	}
	if s.dateTime {
		if s.flags&formatUpper != 0 {
			result.WriteByte('T')
		} else {
			result.WriteByte('t')
		}
	}
	conversion := s.conversion
	if s.flags&formatUpper != 0 && conversion >= 'a' && conversion <= 'z' {
		conversion -= 'a' - 'A'
	}
	result.WriteByte(conversion)
	return result.String()
}

func (s javaStringFormatSpecifier) justify(value []uint16) []uint16 {
	if s.width <= len(value) {
		return value
	}
	output := make([]uint16, s.width)
	for i := range output {
		output[i] = ' '
	}
	if s.flags&formatLeft != 0 {
		copy(output, value)
	} else {
		copy(output[s.width-len(value):], value)
	}
	return output
}

func javaFormatFlag(flag byte) int {
	switch flag {
	case '-':
		return formatLeft
	case '#':
		return formatAlternate
	case '+':
		return formatPlus
	case ' ':
		return formatSpace
	case '0':
		return formatZero
	case ',':
		return formatGroup
	case '(':
		return formatParentheses
	case '<':
		return formatPrevious
	}
	return 0
}

func javaFormatFlags(flags int) string {
	var result strings.Builder
	for _, flag := range []struct {
		bit  int
		char byte
	}{{formatLeft, '-'}, {formatUpper, '^'}, {formatAlternate, '#'}, {formatPlus, '+'}, {formatSpace, ' '}, {formatZero, '0'}, {formatGroup, ','}, {formatParentheses, '('}, {formatPrevious, '<'}} {
		if flags&flag.bit != 0 {
			result.WriteByte(flag.char)
		}
	}
	return result.String()
}

var javaFormatLocaleOnce sync.Once
var javaFormatLocaleLanguage string

func javaFormatDefaultLocale() string {
	javaFormatLocaleOnce.Do(func() {
		javaFormatLocaleLanguage = strings.ToLower(tlcGetSystemProperty("user.language.format", tlcGetSystemProperty("user.language", "en")))
	})
	return javaFormatLocaleLanguage
}

var javaFormatLineSeparatorOnce sync.Once
var javaFormatLineSeparatorValue string

func javaFormatLineSeparator() string {
	javaFormatLineSeparatorOnce.Do(func() {
		separator := "\n"
		if runtime.GOOS == "windows" {
			separator = "\r\n"
		}
		javaFormatLineSeparatorValue = tlcGetSystemProperty("line.separator", separator)
	})
	return javaFormatLineSeparatorValue
}

func javaFormatUpper(units []uint16, locale string) []uint16 {
	tag := language.Und
	// Java's locale-specific uppercase conditions concern Turkish, Azeri and
	// Lithuanian; other languages use Unicode's locale-independent mappings.
	if locale == "tr" || locale == "az" || locale == "lt" {
		tag = language.Make(locale)
	}
	var result []uint16
	for position := 0; position < len(units); {
		end := position
		for end < len(units) {
			if units[end] >= 0xd800 && units[end] <= 0xdbff && end+1 < len(units) && units[end+1] >= 0xdc00 && units[end+1] <= 0xdfff {
				end += 2
				continue
			}
			if units[end] >= 0xd800 && units[end] <= 0xdfff {
				break
			}
			end++
		}
		result = append(result, javaStringUTF16(cases.Upper(tag).String(javaStringFromUTF16(units[position:end])))...)
		if end < len(units) {
			result = append(result, units[end])
			end++
		}
		position = end
	}
	return result
}
