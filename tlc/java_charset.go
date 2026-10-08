// Core text I/O charset behavior, following OpenJDK String's no-replacement
// file helpers and sun.nio.cs UTF-8/Unicode codecs. OpenJDK's GPLv2 with
// Classpath Exception notices are retained in tlc/licenses/.
package tlc

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

type CharacterCodingException struct{ *IOException }
type MalformedInputException struct {
	*CharacterCodingException
	InputLength int
}
type UnmappableCharacterException struct {
	*CharacterCodingException
	InputLength int
}

func NewCharacterCodingException() *CharacterCodingException {
	return &CharacterCodingException{NewIOException()}
}
func NewMalformedInputException(length int) *MalformedInputException {
	return &MalformedInputException{NewCharacterCodingException(), length}
}
func NewUnmappableCharacterException(length int) *UnmappableCharacterException {
	return &UnmappableCharacterException{NewCharacterCodingException(), length}
}
func (e *MalformedInputException) GetInputLength() int      { return e.InputLength }
func (e *UnmappableCharacterException) GetInputLength() int { return e.InputLength }
func (e *MalformedInputException) GetMessage() *string {
	return javaString("Input length = " + strconv.Itoa(e.InputLength))
}
func (e *UnmappableCharacterException) GetMessage() *string {
	return javaString("Input length = " + strconv.Itoa(e.InputLength))
}
func (e *MalformedInputException) Error() string      { return javaThrowableMessage(e) }
func (e *UnmappableCharacterException) Error() string { return javaThrowableMessage(e) }

type UnsupportedCharsetException struct {
	*IllegalArgumentException
	CharsetName string
}
type IllegalCharsetNameException struct {
	*IllegalArgumentException
	CharsetName string
}

func NewUnsupportedCharsetException(name string) *UnsupportedCharsetException {
	return &UnsupportedCharsetException{NewIllegalArgumentException(name), name}
}
func NewIllegalCharsetNameException(name string) *IllegalCharsetNameException {
	return &IllegalCharsetNameException{NewIllegalArgumentException(name), name}
}
func (e *UnsupportedCharsetException) GetCharsetName() string { return e.CharsetName }
func (e *IllegalCharsetNameException) GetCharsetName() string { return e.CharsetName }

// These are the guaranteed standard charsets and their actual OpenJDK aliases.
// Extended/provider charsets remain a separate porting requirement.
func javaCharsetName(name string) (string, error) {
	if name == "" {
		return "", NewIllegalCharsetNameException(name)
	}
	for i, ch := range []byte(name) {
		if ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' {
			continue
		}
		if i > 0 && strings.ContainsRune("-+.:_", rune(ch)) {
			continue
		}
		return "", NewIllegalCharsetNameException(name)
	}
	switch strings.ToUpper(name) {
	case "UTF-8", "UTF8", "UNICODE-1-1-UTF-8":
		return "UTF-8", nil
	case "US-ASCII", "646", "ANSI_X3.4-1968", "ANSI_X3.4-1986", "ASCII", "IBM367", "ISO646-US", "ISO_646.IRV:1991", "ASCII7", "CP367", "CSASCII", "ISO-IR-6", "ISO_646.IRV:1983", "US":
		return "US-ASCII", nil
	case "ISO-8859-1", "819", "8859_1", "IBM-819", "IBM819", "ISO8859-1", "ISO8859_1", "ISO_8859-1", "ISO_8859-1:1987", "ISO_8859_1", "CP819", "CSISOLATIN1", "ISO-IR-100", "L1", "LATIN1":
		return "ISO-8859-1", nil
	case "UTF-16", "UTF_16", "UNICODEBIG", "UNICODE", "UTF16":
		return "UTF-16", nil
	case "UTF-16BE", "ISO-10646-UCS-2", "UTF_16BE", "UNICODEBIGUNMARKED", "X-UTF-16BE":
		return "UTF-16BE", nil
	case "UTF-16LE", "UTF_16LE", "UNICODELITTLEUNMARKED", "X-UTF-16LE":
		return "UTF-16LE", nil
	default:
		return "", NewUnsupportedCharsetException(name)
	}
}

func javaCharsetEncode(text, charset string, strict bool) ([]byte, error) {
	name, err := javaCharsetName(charset)
	if err != nil {
		return nil, err
	}
	units := javaStringUTF16(text)
	switch name {
	case "US-ASCII", "ISO-8859-1":
		limit := uint16(127)
		if name == "ISO-8859-1" {
			limit = 255
		}
		result := make([]byte, 0, len(units))
		for i := 0; i < len(units); i++ {
			unit := units[i]
			if unit <= limit {
				result = append(result, byte(unit))
				continue
			}
			paired := unit >= 0xd800 && unit <= 0xdbff && i+1 < len(units) && units[i+1] >= 0xdc00 && units[i+1] <= 0xdfff
			if strict {
				// String.getBytesNoRepl's Latin-1 fast path reports the first
				// unencodable unit, including any surrogate, as unmappable(1).
				if name == "ISO-8859-1" {
					return nil, NewUnmappableCharacterException(1)
				}
				if paired {
					return nil, NewUnmappableCharacterException(2)
				}
				if unit >= 0xd800 && unit <= 0xdfff {
					return nil, NewMalformedInputException(1)
				}
				return nil, NewUnmappableCharacterException(1)
			}
			result = append(result, '?')
			if paired {
				i++
			}
		}
		return result, nil
	default:
		for i := 0; i < len(units); i++ {
			if units[i] >= 0xd800 && units[i] <= 0xdbff && i+1 < len(units) && units[i+1] >= 0xdc00 && units[i+1] <= 0xdfff {
				i++
			} else if units[i] >= 0xd800 && units[i] <= 0xdfff {
				if strict {
					// String's optimized UTF-8 no-replacement path also uses
					// UnmappableCharacterException, unlike CharsetEncoder.
					if name == "UTF-8" {
						return nil, NewUnmappableCharacterException(1)
					}
					return nil, NewMalformedInputException(1)
				}
				units[i] = 0xfffd
				if name == "UTF-8" {
					units[i] = '?'
				}
			}
		}
		if name == "UTF-8" {
			return []byte(javaStringFromUTF16(units)), nil
		}
		return javaCharsetUTF16Bytes(units, name == "UTF-16", name == "UTF-16LE"), nil
	}
}

func javaCharsetUTF16Bytes(units []uint16, bom, littleEndian bool) []byte {
	result := make([]byte, 0, len(units)*2+2)
	if bom && len(units) != 0 {
		result = append(result, 0xfe, 0xff)
	}
	for _, unit := range units {
		if littleEndian {
			result = append(result, byte(unit), byte(unit>>8))
		} else {
			result = append(result, byte(unit>>8), byte(unit))
		}
	}
	return result
}

func javaCharsetDecode(data []byte, charset string, replace bool) (string, error) {
	name, err := javaCharsetName(charset)
	if err != nil {
		return "", err
	}
	if name == "UTF-8" {
		return javaCharsetDecodeUTF8(data, replace)
	}
	if name == "US-ASCII" || name == "ISO-8859-1" {
		units := make([]uint16, len(data))
		for i, value := range data {
			units[i] = uint16(value)
			if name == "US-ASCII" && value >= 128 {
				if !replace {
					return "", NewMalformedInputException(1)
				}
				units[i] = 0xfffd
			}
		}
		return javaStringFromUTF16(units), nil
	}
	littleEndian := name == "UTF-16LE"
	if name == "UTF-16" && len(data) >= 2 {
		if data[0] == 0xfe && data[1] == 0xff {
			data = data[2:]
		} else if data[0] == 0xff && data[1] == 0xfe {
			littleEndian = true
			data = data[2:]
		}
	}
	unitAt := func(i int) uint16 {
		if littleEndian {
			return uint16(data[i]) | uint16(data[i+1])<<8
		}
		return uint16(data[i])<<8 | uint16(data[i+1])
	}
	var units []uint16
	for i := 0; i < len(data); {
		length := 0
		var unit uint16
		if len(data)-i < 2 {
			length = 1
		} else {
			unit = unitAt(i)
			if unit >= 0xdc00 && unit <= 0xdfff {
				length = 2
			} else if unit >= 0xd800 && unit <= 0xdbff {
				if len(data)-i < 4 {
					length = len(data) - i
				} else if low := unitAt(i + 2); low < 0xdc00 || low > 0xdfff {
					length = 4
				} else {
					units = append(units, unit, low)
					i += 4
					continue
				}
			}
		}
		if length != 0 {
			if !replace {
				return "", NewMalformedInputException(length)
			}
			units = append(units, 0xfffd)
			i += length
		} else {
			units = append(units, unit)
			i += 2
		}
	}
	return javaStringFromUTF16(units), nil
}

// DecodeUTF8Replacing applies the UTF-8 reader's malformed-input replacement
// rules. SANY's InputStream constructor uses this codec explicitly, independently
// of the default charset used by other Java text I/O.
func DecodeUTF8Replacing(data []byte) string {
	text, _ := javaCharsetDecodeUTF8(data, true)
	return text
}

func javaCharsetDecodeUTF8(data []byte, replace bool) (string, error) {
	var result strings.Builder
	for i := 0; i < len(data); {
		first := data[i]
		if first < 128 {
			result.WriteByte(first)
			i++
			continue
		}
		width := 0
		switch {
		case first >= 0xc2 && first <= 0xdf:
			width = 2
		case first >= 0xe0 && first <= 0xef:
			width = 3
		case first >= 0xf0 && first <= 0xf7:
			width = 4
		}
		length := 0
		if width == 0 || first > 0xf4 {
			length = 1
		} else {
			for j := 1; j < width; j++ {
				if i+j >= len(data) {
					length = j
					break
				}
				value := data[i+j]
				if value < 0x80 || value > 0xbf || j == 1 && (first == 0xe0 && value < 0xa0 || first == 0xf0 && value < 0x90 || first == 0xf4 && value > 0x8f) {
					length = j
					break
				}
			}
			if length == 0 && width == 3 && first == 0xed && data[i+1] >= 0xa0 {
				length = 3
			}
		}
		if length != 0 {
			if !replace {
				// Files.readString uses String.newStringNoRepl's optimized
				// UTF-8 path, whose reported lengths differ from the generic
				// CharsetDecoder. Full malformed sequences report their width;
				// truncated sequences report 1, except a bad second byte in
				// a two-byte prefix of a three-byte sequence, which reports 2.
				reported := 1
				if width >= 3 && len(data)-i >= width {
					reported = width
				} else if width == 3 && len(data)-i == 2 && length == 1 {
					reported = 2
				}
				return "", NewMalformedInputException(reported)
			}
			result.WriteRune(utf8.RuneError)
			i += length
		} else {
			result.Write(data[i : i+width])
			i += width
		}
	}
	return result.String(), nil
}
