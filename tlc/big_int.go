// Copyright (c) 2003 Compaq Corporation.  All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation.  All rights reserved.
package tlc

import (
	"io"
	"math/big"
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
	value, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic("invalid BigInt string: " + s)
	}
	return &BigInt{value: value}
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
	if numBits < 0 {
		panic(NewIllegalArgumentException("numBits must be non-negative"))
	}
	numBytes := int((int64(numBits) + 7) / 8)
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
	if b == nil || b.value == nil {
		return new(big.Int)
	}
	return new(big.Int).Set(b.value)
}

func (b *BigInt) FingerPrint() uint64 {
	return FP64NewBytes(BigIntToJavaBytes(b.Value()))
}

func (b *BigInt) Equal(other *BigInt) bool {
	if b == nil || other == nil {
		return b == other
	}
	return b.value.Cmp(other.value) == 0
}

func (b *BigInt) Cmp(other *BigInt) int {
	if b == nil && other == nil {
		return 0
	}
	if b == nil {
		return -1
	}
	if other == nil {
		return 1
	}
	return b.value.Cmp(other.value)
}

func (b *BigInt) Write(out io.Writer) error {
	return WriteSizeBigInt(out, b.Value())
}

func ReadBigInt(in io.Reader) (*BigInt, error) {
	value, err := ReadSizeBigInt(in)
	if err != nil {
		return nil, err
	}
	return NewBigInt(value), nil
}

func (b *BigInt) String() string {
	if b == nil || b.value == nil {
		return "0"
	}
	return b.value.String()
}
