package tlc

import (
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

func javaStringUTF16(text string) []uint16 {
	var result []uint16
	for len(text) > 0 {
		// Preserve the surrogate units produced by an earlier precision truncation.
		if len(text) >= 3 && text[0] == 0xed && text[1] >= 0xa0 && text[1] <= 0xbf && text[2] >= 0x80 && text[2] <= 0xbf {
			result = append(result, uint16(text[0]&15)<<12|uint16(text[1]&63)<<6|uint16(text[2]&63))
			text = text[3:]
			continue
		}
		char, size := utf8.DecodeRuneInString(text)
		text = text[size:]
		if char > 0xffff {
			high, low := utf16.EncodeRune(char)
			result = append(result, uint16(high), uint16(low))
		} else {
			result = append(result, uint16(char))
		}
	}
	return result
}

// javaStringCompareTo preserves String.compareTo's unsigned UTF-16 code-unit
// ordering, including strings containing supplementary or unpaired surrogates.
func javaStringCompareTo(left, right string) int {
	lunits := javaStringUTF16(left)
	runits := javaStringUTF16(right)
	limit := min(len(lunits), len(runits))
	for i := 0; i < limit; i++ {
		if lunits[i] < runits[i] {
			return -1
		}
		if lunits[i] > runits[i] {
			return 1
		}
	}
	if len(lunits) < len(runits) {
		return -1
	}
	if len(lunits) > len(runits) {
		return 1
	}
	return 0
}

func javaStringFromUTF16(units []uint16) string {
	var result strings.Builder
	for i := 0; i < len(units); i++ {
		unit := units[i]
		if unit >= 0xd800 && unit <= 0xdbff && i+1 < len(units) && units[i+1] >= 0xdc00 && units[i+1] <= 0xdfff {
			result.WriteRune(utf16.DecodeRune(rune(unit), rune(units[i+1])))
			i++
		} else if unit >= 0xd800 && unit <= 0xdfff {
			result.WriteByte(0xe0 | byte(unit>>12))
			result.WriteByte(0x80 | byte((unit>>6)&63))
			result.WriteByte(0x80 | byte(unit&63))
		} else {
			result.WriteRune(rune(unit))
		}
	}
	return result.String()
}

func javaStringConcat(left, right string) string {
	units := append(javaStringUTF16(left), javaStringUTF16(right)...)
	return javaStringFromUTF16(units)
}

func javaStringEqual(left, right string) bool {
	leftUnits, rightUnits := javaStringUTF16(left), javaStringUTF16(right)
	if len(leftUnits) != len(rightUnits) {
		return false
	}
	for i := range leftUnits {
		if leftUnits[i] != rightUnits[i] {
			return false
		}
	}
	return true
}

func javaFormatStringHash(text string) int32 {
	var hash int32
	for _, unit := range javaStringUTF16(text) {
		hash = 31*hash + int32(unit)
	}
	return hash
}

// Java's UTF-8 encoder replaces unpaired UTF-16 surrogates with '?' when an output API
// writes a Java string. Valid pairs remain supplementary code points.
func javaStringUTF8(text string) string {
	units := javaStringUTF16(text)
	for i := 0; i < len(units); i++ {
		if units[i] >= 0xd800 && units[i] <= 0xdbff && i+1 < len(units) && units[i+1] >= 0xdc00 && units[i+1] <= 0xdfff {
			i++
		} else if units[i] >= 0xd800 && units[i] <= 0xdfff {
			units[i] = '?'
		}
	}
	return javaStringFromUTF16(units)
}
