package tlc

import (
	"io"
	"math/big"
	"unicode"
)

var (
	BigZero = NewBigIntString("0")
	BigOne  = NewBigIntString("1")
	BigTwo  = NewBigIntString("2")
)

type BigInt struct {
	value *big.Int
}

func NewBigIntString(s string) *BigInt {
	return NewBigIntStringNullable(javaString(s))
}

func NewBigIntStringNullable(s *string) *BigInt {
	if s == nil {
		panic(NewNullPointerException())
	}
	units := javaStringUTF16(*s)
	if len(units) == 0 {
		panic(NewNumberFormatExceptionNullable(javaString("Zero length BigInteger")))
	}

	cursor := 0
	negative := false
	if units[0] == '+' || units[0] == '-' {
		negative = units[0] == '-'
		cursor = 1
	}
	for _, unit := range units[cursor:] {
		if unit == '+' || unit == '-' {
			panic(NewNumberFormatExceptionNullable(javaString("Illegal embedded sign character")))
		}
	}
	if cursor == len(units) {
		panic(NewNumberFormatExceptionNullable(javaString("Zero length BigInteger")))
	}

	for cursor < len(units) {
		digit, ok := javaDecimalDigit(units[cursor])
		if !ok || digit != 0 {
			break
		}
		cursor++
	}
	if cursor == len(units) {
		return &BigInt{value: new(big.Int)}
	}

	numDigits := len(units) - cursor
	firstGroupLength := numDigits % 9
	if firstGroupLength == 0 {
		firstGroupLength = 9
	}
	digits := make([]byte, 0, numDigits+1)
	if negative {
		digits = append(digits, '-')
	}
	for groupLength := firstGroupLength; cursor < len(units); groupLength = 9 {
		end := cursor + groupLength
		group := units[cursor:end]
		for _, unit := range group {
			digit, ok := javaDecimalDigit(unit)
			if !ok {
				panic(NewNumberFormatExceptionForInputString(javaStringFromUTF16(group), 10))
			}
			digits = append(digits, '0'+digit)
		}
		cursor = end
	}
	value, ok := new(big.Int).SetString(string(digits), 10)
	if !ok {
		panic(NewNumberFormatExceptionForInputString(*s, 10))
	}
	return &BigInt{value: value}
}

func javaDecimalDigit(unit uint16) (byte, bool) {
	r := rune(unit)
	if r >= '0' && r <= '9' {
		return byte(r - '0'), true
	}
	for _, value := range unicode.Digit.R16 {
		if r >= rune(value.Lo) && r <= rune(value.Hi) && (r-rune(value.Lo))%rune(value.Stride) == 0 {
			return byte(((r - rune(value.Lo)) / rune(value.Stride)) % 10), true
		}
	}
	for _, value := range unicode.Digit.R32 {
		if r >= rune(value.Lo) && r <= rune(value.Hi) && (r-rune(value.Lo))%rune(value.Stride) == 0 {
			return byte(((r - rune(value.Lo)) / rune(value.Stride)) % 10), true
		}
	}
	return 0, false
}

func NewBigIntBytes(bytes []byte) *BigInt {
	value, err := JavaBytesToBigInt(bytes)
	if err != nil {
		panic(err)
	}
	return &BigInt{value: value}
}

// BigInt(int, Random) delegates to BigInteger's unsigned randomBits path.
func NewBigIntRandom(numBits int, rnd *JavaRandom) *BigInt {
	numBits = int(int32(numBits))
	if numBits < 0 {
		panic(NewIllegalArgumentException("numBits must be non-negative"))
	}
	numBytes := int(int32(numBits)+7) / 8
	if numBytes < 0 {
		panic(NewNegativeArraySizeException(fmtInt(numBytes)))
	}
	bytes := make([]byte, numBytes)
	if numBytes > 0 {
		if rnd == nil {
			panic(NewNullPointerException())
		}
		rnd.NextBytes(bytes)
		excessBits := 8*numBytes - numBits
		bytes[0] &= byte((uint16(1) << uint(8-excessBits)) - 1)
	}
	return &BigInt{value: new(big.Int).SetBytes(bytes)}
}

func NewBigInt(value *big.Int) *BigInt {
	if value == nil {
		value = new(big.Int)
	}
	return &BigInt{value: new(big.Int).Set(value)}
}

func (b *BigInt) Value() *big.Int {
	b.require()
	return new(big.Int).Set(b.value)
}

func (b *BigInt) FingerPrint() uint64 {
	return FP64NewBytes(BigIntToJavaBytes(b.Value()))
}

func (b *BigInt) Equal(other *BigInt) bool {
	b.require()
	if other == nil {
		return false
	}
	other.require()
	return b.value.Cmp(other.value) == 0
}

func (b *BigInt) Cmp(other *BigInt) int {
	b.require()
	other.require()
	return b.value.Cmp(other.value)
}

func (b *BigInt) Write(out io.Writer) error {
	b.require()
	return WriteSizeBigInt(out, b.Value())
}

func (b *BigInt) Read(in io.Reader) (ExternalSortable, error) {
	b.require()
	return ReadBigInt(in)
}

func ReadBigInt(in io.Reader) (*BigInt, error) {
	value, err := ReadSizeBigInt(in)
	if err != nil {
		return nil, err
	}
	return NewBigInt(value), nil
}

func (b *BigInt) String() string {
	b.require()
	return b.value.String()
}

func (b *BigInt) require() {
	if b == nil || b.value == nil {
		panic(NewNullPointerException())
	}
}
