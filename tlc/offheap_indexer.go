package tlc

import (
	"math/big"
	"math/bits"
)

type OffHeapIndexerKind int

const (
	OffHeapIndexerBitshifting OffHeapIndexerKind = iota
	OffHeapIndexerMult1024
	OffHeapIndexerInfinitePrecision
)

type OffHeapIndexer struct {
	Kind       OffHeapIndexerKind
	Positions  uint64
	FPBits     int
	PrefixMask uint64
	RShift     uint
	Multiplier uint64
	Shift      uint
}

func NewOffHeapIndexer(positions int64, fpBits int) *OffHeapIndexer {
	if positions <= 0 || fpBits <= 0 || fpBits >= 64 {
		panic("invalid OffHeapIndexer configuration")
	}
	pos := uint64(positions)
	if bits.OnesCount64(pos) == 1 {
		return NewBitshiftingOffHeapIndexer(positions, fpBits)
	}
	if OffHeapMult1024IndexerIsSupported(positions) {
		return NewMult1024OffHeapIndexer(positions, fpBits)
	}
	return NewInfinitePrecisionOffHeapIndexer(positions, fpBits)
}

func NewBitshiftingOffHeapIndexer(positions int64, fpBits int) *OffHeapIndexer {
	if positions <= 0 || fpBits <= 0 || fpBits >= 64 {
		panic("invalid BitshiftingOffHeapIndexer configuration")
	}
	pos := uint64(positions)
	prefixMask := ^uint64(0) >> uint(fpBits)
	n := prefixMask - (pos - 1)
	var moveBy uint
	for n >= pos {
		moveBy++
		n >>= 1
	}
	return &OffHeapIndexer{
		Kind:       OffHeapIndexerBitshifting,
		Positions:  pos,
		FPBits:     fpBits,
		PrefixMask: prefixMask,
		RShift:     moveBy,
	}
}

func OffHeapMult1024IndexerIsSupported(positions int64) bool {
	return positions > 0 && ((uint64(positions)<<3)%(1<<30)) == 0
}

func NewMult1024OffHeapIndexer(positions int64, fpBits int) *OffHeapIndexer {
	if positions <= 0 || fpBits <= 0 || fpBits >= 64 || !OffHeapMult1024IndexerIsSupported(positions) {
		panic("invalid Mult1024OffHeapIndexer configuration")
	}
	pos := uint64(positions)
	if bits.TrailingZeros64(pos) <= fpBits {
		panic("fingerprint space is smaller than number of positions")
	}
	max := new(big.Int).Lsh(big.NewInt(1), uint(64-fpBits))
	bPos := new(big.Int).SetUint64(pos)
	gcd := new(big.Int).GCD(nil, nil, max, bPos)
	multiplier := new(big.Int).Div(new(big.Int).Set(bPos), gcd).Uint64()
	rMax := new(big.Int).Div(new(big.Int).Set(max), gcd)
	return &OffHeapIndexer{
		Kind:       OffHeapIndexerMult1024,
		Positions:  pos,
		FPBits:     fpBits,
		Multiplier: multiplier,
		Shift:      uint(rMax.TrailingZeroBits()),
	}
}

func NewInfinitePrecisionOffHeapIndexer(positions int64, fpBits int) *OffHeapIndexer {
	if positions <= 0 || fpBits <= 0 || fpBits >= 64 {
		panic("invalid InfinitePrecisionOffHeapIndexer configuration")
	}
	return &OffHeapIndexer{
		Kind:      OffHeapIndexerInfinitePrecision,
		Positions: uint64(positions),
		FPBits:    fpBits,
	}
}

func (i *OffHeapIndexer) GetIdx(fp uint64) int64 {
	return i.GetIdxProbe(fp, 0)
}

func (i *OffHeapIndexer) GetIdxProbe(fp uint64, probe int) int64 {
	if i == nil || i.Positions == 0 {
		panic("uninitialized OffHeapIndexer")
	}
	fp &= diskFPSetFlushedMask
	var idx uint64
	switch i.Kind {
	case OffHeapIndexerBitshifting:
		idx = (fp & i.PrefixMask) >> i.RShift
	case OffHeapIndexerMult1024:
		hi, lo := bits.Mul64(fp, i.Multiplier)
		leftShift := (64 - i.Shift) & 63
		rightShift := i.Shift & 63
		idx = (hi << leftShift) | (lo >> rightShift)
		idx %= i.Positions
	default:
		idx = i.infinitePrecisionIdx(fp)
	}
	if probe != 0 {
		idx = (idx + uint64(probe)) % i.Positions
	}
	return int64(idx)
}

func (i *OffHeapIndexer) infinitePrecisionIdx(fp uint64) uint64 {
	numerator := new(big.Int).SetUint64(fp)
	numerator.Mul(numerator, new(big.Int).SetUint64(i.Positions))
	denominator := new(big.Int).Lsh(big.NewInt(1), uint(64-i.FPBits))
	numerator.Div(numerator, denominator)
	numerator.Mod(numerator, new(big.Int).SetUint64(i.Positions))
	return numerator.Uint64()
}
