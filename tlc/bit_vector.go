package tlc

import (
	"math/bits"
	"strconv"
	"strings"
)

type BitVector struct {
	word []uint64
}

func NewBitVector(capacities ...int) *BitVector {
	if len(capacities) == 0 {
		return &BitVector{}
	}
	if len(capacities) > 1 {
		panic(NewIllegalArgumentException("NewBitVector accepts at most one capacity"))
	}
	// Java computes this with signed int arithmetic. Small negative capacities
	// therefore produce zero or one word before -127 first produces a negative
	// array length.
	capacity := int32(capacities[0])
	length := int32(0)
	if capacity != 0 {
		length = (capacity-1)/64 + 1
	}
	if length < 0 {
		panic(NewNegativeArraySizeException(strconv.FormatInt(int64(length), 10)))
	}
	return &BitVector{word: make([]uint64, int(length))}
}

func NewBitVectorWithValue(initCapacity int, initValue bool) *BitVector {
	bv := NewBitVector(initCapacity)
	if initValue {
		bv.SetRange(0, len(bv.word))
	}
	return bv
}

func NewBitVectorCopy(other *BitVector) *BitVector {
	if other == nil {
		panic(NewNullPointerException())
	}
	otherWords := other.requireWords()
	words := make([]uint64, len(otherWords))
	copy(words, otherWords)
	return &BitVector{word: words}
}

func (bv *BitVector) Equal(other *BitVector) bool {
	if bv == nil {
		panic(NewNullPointerException())
	}
	if other == nil {
		return false
	}
	words := bv.requireWords()
	otherWords := other.requireWords()
	minLen := min(len(words), len(otherWords))
	for i := 0; i < minLen; i++ {
		if words[i] != otherWords[i] {
			return false
		}
	}
	if len(words) == len(otherWords) {
		return true
	}
	tail := words
	if len(otherWords) > len(words) {
		tail = otherWords
	}
	for i := minLen; i < len(tail); i++ {
		if tail[i] != 0 {
			return false
		}
	}
	return true
}

func (bv *BitVector) HashCode() int {
	var res int32
	for _, word := range bv.requireWords() {
		if word != 0 {
			res ^= int32(word & 0xffff)
			res ^= int32(word >> 32)
		}
	}
	return int(res)
}

func (bv *BitVector) Hash() int {
	return bv.HashCode()
}

func (bv *BitVector) Clear() {
	words := bv.requireWords()
	for i := range words {
		words[i] = 0
	}
}

func (bv *BitVector) Get(i int) bool {
	i = int(int32(i))
	words := bv.requireWords()
	wd := i / 64
	if wd >= len(words) {
		return false
	}
	// Java masks long shift distances to six bits, including negative indices.
	bit := uint(i & 63)
	return (bitVectorWordAt(words, wd) & (uint64(1) << bit)) != 0
}

func (bv *BitVector) Set(i int) {
	i = int(int32(i))
	bv.requireWords()
	wd := i / 64
	if wd >= len(bv.word) {
		bv.grow(wd)
	}
	bit := uint(i & 63)
	bitVectorSetWord(bv.word, wd, bitVectorWordAt(bv.word, wd)|(uint64(1)<<bit))
}

func (bv *BitVector) SetRange(lo, hi int) {
	lo = int(int32(lo))
	hi = int(int32(hi))
	bv.requireWords()
	lwd := lo / 64
	hwd := hi / 64
	if hwd >= len(bv.word) {
		bv.grow(hwd)
	}
	lbit := uint(lo & 63)
	hbit := uint(hi & 63)
	if lwd < hwd {
		for i := lwd + 1; i < hwd; i++ {
			bitVectorSetWord(bv.word, i, ^uint64(0))
		}
		bitVectorSetWord(bv.word, lwd, ^uint64(0)<<lbit)
		bitVectorSetWord(bv.word, hwd, ^uint64(0)>>(63-hbit))
		return
	}
	if lo <= hi {
		bitVectorSetWord(bv.word, lwd, (^uint64(0)<<lbit)&(^uint64(0)>>(63-hbit)))
	}
}

func (bv *BitVector) Reset(i int) {
	i = int(int32(i))
	bv.requireWords()
	wd := i / 64
	if wd >= len(bv.word) {
		bv.grow(wd)
	}
	bit := uint(i & 63)
	bitVectorSetWord(bv.word, wd, bitVectorWordAt(bv.word, wd)&^(uint64(1)<<bit))
}

func (bv *BitVector) SetBool(i int, val bool) {
	if val {
		bv.Set(i)
	} else {
		bv.Reset(i)
	}
}

func (bv *BitVector) TrueCnt() int {
	count := 0
	for _, word := range bv.requireWords() {
		count += bits.OnesCount64(word)
	}
	return count
}

func (bv *BitVector) TrueCount() int {
	return bv.TrueCnt()
}

func (bv *BitVector) String() string {
	words := bv.requireWords()
	if len(words) == 0 {
		return "[]"
	}
	var b strings.Builder
	started := false
	for i := len(words) * 64; i >= 0; i-- {
		if bv.Get(i) {
			started = true
			b.WriteByte('1')
		} else if started {
			b.WriteByte('0')
		}
	}
	if !started {
		return "[]"
	}
	return "[" + b.String() + "]"
}

func (bv *BitVector) StringRange(start, length int) string {
	return bv.StringRangeChars(start, length, '1', '0')
}

func (bv *BitVector) StringRangeChars(start, length int, one, zero uint16) string {
	start = int(int32(start))
	length = int(int32(length))
	if length < 0 {
		if bv == nil {
			panic(NewNullPointerException())
		}
		panic(NewNegativeArraySizeException(strconv.Itoa(length)))
	}
	buf := make([]uint16, length)
	for i := 0; i < length; i++ {
		if bv.Get(start + i) {
			buf[length-1-i] = one
		} else {
			buf[length-1-i] = zero
		}
	}
	return javaStringFromUTF16(append(append([]uint16{'['}, buf...), ']'))
}

func (bv *BitVector) Write(out dataOutput) error {
	words := bv.requireWords()
	if isNil(out) {
		panic(NewNullPointerException())
	}
	if err := out.WriteNat(int32(len(words))); err != nil {
		return err
	}
	for _, word := range words {
		if err := out.WriteLong(int64(word)); err != nil {
			return err
		}
	}
	return nil
}

func (bv *BitVector) Read(in dataInput) error {
	if bv == nil {
		panic(NewNullPointerException())
	}
	if isNil(in) {
		panic(NewNullPointerException())
	}
	length, err := in.ReadNat()
	if err != nil {
		return err
	}
	if length < 0 {
		panic(NewNegativeArraySizeException(strconv.FormatInt(int64(length), 10)))
	}
	bv.word = make([]uint64, int(length))
	for i := range bv.word {
		word, err := in.ReadLong()
		if err != nil {
			return err
		}
		bv.word[i] = uint64(word)
	}
	return nil
}

func (bv *BitVector) WriteToBufferedRandomAccessFile(raf *BufferedRandomAccessFile) error {
	words := bv.requireWords()
	if raf == nil {
		panic(NewNullPointerException())
	}
	if err := raf.WriteNat(len(words)); err != nil {
		return err
	}
	for _, word := range words {
		if err := raf.WriteLong(int64(word)); err != nil {
			return err
		}
	}
	return nil
}

func (bv *BitVector) ReadFromBufferedRandomAccessFile(raf *BufferedRandomAccessFile) error {
	if bv == nil {
		panic(NewNullPointerException())
	}
	if raf == nil {
		panic(NewNullPointerException())
	}
	length, err := raf.ReadNat()
	if err != nil {
		return err
	}
	if length < 0 {
		panic(NewNegativeArraySizeException(strconv.Itoa(length)))
	}
	bv.word = make([]uint64, length)
	for i := range bv.word {
		word, err := raf.ReadLong()
		if err != nil {
			return err
		}
		bv.word[i] = uint64(word)
	}
	return nil
}

func (bv *BitVector) grow(wd int) {
	bv.requireWords()
	if wd < len(bv.word) {
		return
	}
	tmp := make([]uint64, wd+1)
	copy(tmp, bv.word)
	bv.word = tmp
}

func (bv *BitVector) requireWords() []uint64 {
	if bv == nil || bv.word == nil {
		panic(NewNullPointerException())
	}
	return bv.word
}

func bitVectorWordAt(words []uint64, index int) uint64 {
	if index < 0 || index >= len(words) {
		panic(NewArrayIndexOutOfBoundsException(index, len(words)))
	}
	return words[index]
}

func bitVectorSetWord(words []uint64, index int, value uint64) {
	if index < 0 || index >= len(words) {
		panic(NewArrayIndexOutOfBoundsException(index, len(words)))
	}
	words[index] = value
}

type BitVectorIter struct {
	word []uint64
	wd   int
	bit  int
	mask uint64
}

func NewBitVectorIter(vectors ...*BitVector) *BitVectorIter {
	if len(vectors) > 1 {
		panic(NewIllegalArgumentException("NewBitVectorIter accepts at most one bit vector"))
	}
	iter := &BitVectorIter{}
	if len(vectors) > 0 {
		iter.Init(vectors[0])
	}
	return iter
}

func (it *BitVectorIter) Init(bv *BitVector) {
	if it == nil || bv == nil {
		panic(NewNullPointerException())
	}
	it.word = bv.word
	it.wd = 0
	it.bit = 0
	it.mask = 1
}

func (it *BitVectorIter) Next() int {
	if it == nil || it.word == nil {
		panic(NewNullPointerException())
	}
	for ; it.wd < len(it.word); it.wd++ {
		word := it.word[it.wd]
		for ; it.bit < 64; it.bit, it.mask = it.bit+1, it.mask<<1 {
			if (word & it.mask) != 0 {
				res := it.wd*64 + it.bit
				it.bit++
				if it.bit < 64 {
					it.mask <<= 1
				} else {
					it.wd++
					it.bit = 0
					it.mask = 1
				}
				return res
			}
		}
		it.bit = 0
		it.mask = 1
	}
	return -1
}
