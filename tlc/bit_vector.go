package tlc

import (
	"math/bits"
	"strings"
)

type BitVector struct {
	word []uint64
}

func NewBitVector(initCapacity int) *BitVector {
	if initCapacity < 0 {
		initCapacity = 0
	}
	length := 0
	if initCapacity != 0 {
		length = ((initCapacity - 1) / 64) + 1
	}
	return &BitVector{word: make([]uint64, length)}
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
		return NewBitVector(0)
	}
	words := make([]uint64, len(other.word))
	copy(words, other.word)
	return &BitVector{word: words}
}

func (bv *BitVector) Equal(other *BitVector) bool {
	if bv == nil || other == nil {
		return bv == other
	}
	minLen := min(len(bv.word), len(other.word))
	for i := 0; i < minLen; i++ {
		if bv.word[i] != other.word[i] {
			return false
		}
	}
	if len(bv.word) == len(other.word) {
		return true
	}
	tail := bv.word
	if len(other.word) > len(bv.word) {
		tail = other.word
	}
	for i := minLen; i < len(tail); i++ {
		if tail[i] != 0 {
			return false
		}
	}
	return true
}

func (bv *BitVector) Hash() int {
	var res int32
	for _, word := range bv.word {
		if word != 0 {
			res ^= int32(word & 0xffff)
			res ^= int32(word >> 32)
		}
	}
	return int(res)
}

func (bv *BitVector) Clear() {
	for i := range bv.word {
		bv.word[i] = 0
	}
}

func (bv *BitVector) Get(i int) bool {
	if i < 0 {
		return false
	}
	wd := i / 64
	if wd >= len(bv.word) {
		return false
	}
	bit := uint(i % 64)
	return (bv.word[wd] & (uint64(1) << bit)) != 0
}

func (bv *BitVector) Set(i int) {
	if i < 0 {
		return
	}
	wd := i / 64
	if wd >= len(bv.word) {
		bv.grow(wd)
	}
	bit := uint(i % 64)
	bv.word[wd] |= uint64(1) << bit
}

func (bv *BitVector) SetRange(lo, hi int) {
	if lo < 0 || hi < 0 || lo > hi {
		return
	}
	lwd := lo / 64
	hwd := hi / 64
	if hwd >= len(bv.word) {
		bv.grow(hwd)
	}
	lbit := uint(lo % 64)
	hbit := uint(hi % 64)
	if lwd < hwd {
		for i := lwd + 1; i < hwd; i++ {
			bv.word[i] = ^uint64(0)
		}
		bv.word[lwd] = ^uint64(0) << lbit
		bv.word[hwd] = ^uint64(0) >> (63 - hbit)
		return
	}
	bv.word[lwd] = (^uint64(0) << lbit) & (^uint64(0) >> (63 - hbit))
}

func (bv *BitVector) Reset(i int) {
	if i < 0 {
		return
	}
	wd := i / 64
	if wd >= len(bv.word) {
		bv.grow(wd)
	}
	bit := uint(i % 64)
	bv.word[wd] &^= uint64(1) << bit
}

func (bv *BitVector) SetBool(i int, val bool) {
	if val {
		bv.Set(i)
	} else {
		bv.Reset(i)
	}
}

func (bv *BitVector) TrueCount() int {
	count := 0
	for _, word := range bv.word {
		count += bits.OnesCount64(word)
	}
	return count
}

func (bv *BitVector) String() string {
	if len(bv.word) == 0 {
		return "[]"
	}
	var b strings.Builder
	started := false
	for i := len(bv.word) * 64; i >= 0; i-- {
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

func (bv *BitVector) StringRangeChars(start, length int, one, zero byte) string {
	if length < 0 {
		length = 0
	}
	buf := make([]byte, length)
	for i := 0; i < length; i++ {
		if bv.Get(start + i) {
			buf[length-1-i] = one
		} else {
			buf[length-1-i] = zero
		}
	}
	return "[" + string(buf) + "]"
}

func (bv *BitVector) Write(out dataOutput) error {
	if bv == nil {
		return out.WriteNat(0)
	}
	if err := out.WriteNat(int32(len(bv.word))); err != nil {
		return err
	}
	for _, word := range bv.word {
		if err := out.WriteLong(int64(word)); err != nil {
			return err
		}
	}
	return nil
}

func (bv *BitVector) Read(in dataInput) error {
	length, err := in.ReadNat()
	if err != nil {
		return err
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
	if bv == nil {
		return raf.WriteNat(0)
	}
	if err := raf.WriteNat(len(bv.word)); err != nil {
		return err
	}
	for _, word := range bv.word {
		if err := raf.WriteLong(int64(word)); err != nil {
			return err
		}
	}
	return nil
}

func (bv *BitVector) ReadFromBufferedRandomAccessFile(raf *BufferedRandomAccessFile) error {
	length, err := raf.ReadNat()
	if err != nil {
		return err
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
	if wd < len(bv.word) {
		return
	}
	tmp := make([]uint64, wd+1)
	copy(tmp, bv.word)
	bv.word = tmp
}

type BitVectorIter struct {
	word []uint64
	wd   int
	bit  int
	mask uint64
}

func NewBitVectorIter(bv *BitVector) *BitVectorIter {
	iter := &BitVectorIter{}
	iter.Init(bv)
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
