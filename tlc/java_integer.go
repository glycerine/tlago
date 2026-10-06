package tlc

import "unicode"

// Integer.parseInt with the default decimal radix differs from decode: a
// leading zero is decimal, and hexadecimal/octal prefixes are not recognized.
func javaParseDecimalInt(value string) (int32, bool) {
	if value == "" {
		return 0, false
	}
	negative := value[0] == '-'
	if negative || value[0] == '+' {
		value = value[1:]
	}
	if value == "" {
		return 0, false
	}
	limit := int64(1<<31 - 1)
	if negative {
		limit++
	}
	var result int64
	for _, char := range value {
		digit := javaIntegerDigit(char)
		if digit < 0 || digit > 9 || result > (limit-int64(digit))/10 {
			return 0, false
		}
		result = result*10 + int64(digit)
	}
	if negative {
		result = -result
	}
	return int32(result), true
}

// Long.parseLong uses the same UTF-16 decimal digits as Integer.parseInt.
// Accumulate negatively so the asymmetric signed minimum never overflows.
func javaParseDecimalLong(value string) (int64, bool) {
	if value == "" {
		return 0, false
	}
	negative := value[0] == '-'
	if negative || value[0] == '+' {
		value = value[1:]
	}
	if value == "" {
		return 0, false
	}
	limit := int64(-1<<63 + 1)
	if negative {
		limit = -1 << 63
	}
	var result int64
	for _, char := range value {
		digit := javaIntegerDigit(char)
		if digit < 0 || digit > 9 || result < limit/10 {
			return 0, false
		}
		result *= 10
		if result < limit+int64(digit) {
			return 0, false
		}
		result -= int64(digit)
	}
	if !negative {
		result = -result
	}
	return result, true
}

// Integer.getInteger uses Integer.decode: signs precede a hex/octal prefix,
// whitespace is invalid, and the result must fit a signed Java int.
func javaDecodeIntProperty(value string) (int, bool) {
	if value == "" {
		return 0, false
	}
	negative := value[0] == '-'
	if negative || value[0] == '+' {
		value = value[1:]
	}
	radix := int64(10)
	if len(value) > 1 && value[0] == '0' && (value[1] == 'x' || value[1] == 'X') {
		radix, value = 16, value[2:]
	} else if len(value) > 0 && value[0] == '#' {
		radix, value = 16, value[1:]
	} else if len(value) > 1 && value[0] == '0' {
		radix, value = 8, value[1:]
	}
	if value == "" {
		return 0, false
	}
	limit := int64(1<<31 - 1)
	if negative {
		limit++
	}
	var result int64
	for _, char := range value {
		digit := javaIntegerDigit(char)
		if digit < 0 || int64(digit) >= radix || result > (limit-int64(digit))/radix {
			return 0, false
		}
		result = result*radix + int64(digit)
	}
	if negative {
		result = -result
	}
	return int(result), true
}

// Integer.parseInt examines UTF-16 chars, so supplementary digits are rejected.
func javaIntegerDigit(char rune) int {
	switch {
	case char >= '0' && char <= '9':
		return int(char - '0')
	case char >= 'a' && char <= 'z':
		return int(char-'a') + 10
	case char >= 'A' && char <= 'Z':
		return int(char-'A') + 10
	case char >= 0xff41 && char <= 0xff5a:
		return int(char-0xff41) + 10
	case char >= 0xff21 && char <= 0xff3a:
		return int(char-0xff21) + 10
	}
	if char <= 0xffff {
		for _, digits := range unicode.Digit.R16 {
			if uint16(char) >= digits.Lo && uint16(char) <= digits.Hi && (uint16(char)-digits.Lo)%digits.Stride == 0 {
				return int((uint16(char)-digits.Lo)/digits.Stride) % 10
			}
		}
	}
	return -1
}
