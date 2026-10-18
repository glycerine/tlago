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
	OffHeapIndexerInfinitePrecisionMult1024
)

type OffHeapIndexer struct {
	Kind       OffHeapIndexerKind
	Positions  uint64
	FPBits     int
	PrefixMask uint64
	RShift     uint
	Multiplier uint64
	Shift      uint
	// Explicit source constructor retains its exactly representable decimal factor.
	factor *big.Rat

	configurationCaptured bool
	configurationOwner    offHeapIndexerConfiguration
}

type offHeapIndexerConfiguration struct {
	kind       OffHeapIndexerKind
	positions  uint64
	fpBits     int
	prefixMask uint64
	rShift     uint
	multiplier uint64
	shift      uint
	factor     *big.Rat
}

func (i *OffHeapIndexer) captureConfiguration() *OffHeapIndexer {
	i.configurationOwner = offHeapIndexerConfiguration{
		kind:       i.Kind,
		positions:  i.Positions,
		fpBits:     i.FPBits,
		prefixMask: i.PrefixMask,
		rShift:     i.RShift,
		multiplier: i.Multiplier,
		shift:      i.Shift,
		factor:     i.factor,
	}
	i.configurationCaptured = true
	return i
}

func (i *OffHeapIndexer) configuration() offHeapIndexerConfiguration {
	if i.configurationCaptured {
		return i.configurationOwner
	}
	return offHeapIndexerConfiguration{
		kind:       i.Kind,
		positions:  i.Positions,
		fpBits:     i.FPBits,
		prefixMask: i.PrefixMask,
		rShift:     i.RShift,
		multiplier: i.Multiplier,
		shift:      i.Shift,
		factor:     i.factor,
	}
}

func NewOffHeapIndexer(positions int64, fpBits int) *OffHeapIndexer {
	fpBits = int(int32(fpBits))
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
	fpBits = int(int32(fpBits))
	if positions < 0 || fpBits <= 0 || fpBits >= 64 {
		panic(NewAssertionError())
	}
	// The source loop does not terminate for zero positions. Retain the native
	// defensive rejection instead of manufacturing a successful indexer.
	if positions == 0 {
		panic("invalid BitshiftingOffHeapIndexer configuration")
	}
	pos := uint64(positions)
	prefixMask := ^uint64(0) >> uint(fpBits)
	if prefixMask <= pos {
		panic(NewAssertionError("fingerprint equals index if positions exceeds fingerprint space."))
	}
	n := prefixMask - (pos - 1)
	var moveBy uint
	for n >= pos {
		moveBy++
		n >>= 1
	}
	return (&OffHeapIndexer{
		Kind:       OffHeapIndexerBitshifting,
		Positions:  pos,
		FPBits:     fpBits,
		PrefixMask: prefixMask,
		RShift:     moveBy,
	}).captureConfiguration()
}

func OffHeapMult1024IndexerIsSupported(positions int64) bool {
	return positions > 0 && Mult1024IndexerIsSupported(positions)
}

// Mult1024IndexerIsSupported ports Mult1024Indexer.isSupported, including
// Java long overflow in positions << 3. The constructor applies its separate
// nonnegative and fingerprint-space assertions.
func Mult1024IndexerIsSupported(positions int64) bool {
	return (positions<<3)%(1<<30) == 0
}

// InfinitePrecisionMult1024IndexerIsSupported ports the identically shaped
// static method on Java's distinct infinite-precision indexer class.
func InfinitePrecisionMult1024IndexerIsSupported(positions int64) bool {
	return (positions<<3)%(1<<30) == 0
}

func NewMult1024OffHeapIndexer(positions int64, fpBits int) *OffHeapIndexer {
	fpBits = int(int32(fpBits))
	if fpBits <= 0 || fpBits >= 64 {
		panic(NewAssertionError())
	}
	if !Mult1024IndexerIsSupported(positions) {
		panic(NewAssertionError("positions * 8 is not a multiple of 1024MB"))
	}
	if positions < 0 || bits.TrailingZeros64(uint64(positions)) <= fpBits {
		panic(NewAssertionError("fingerprint space is smaller than number of positions"))
	}
	pos := uint64(positions)
	max := new(big.Int).Lsh(big.NewInt(1), uint(64-fpBits))
	bPos := new(big.Int).SetUint64(pos)
	gcd := new(big.Int).GCD(nil, nil, max, bPos)
	multiplier := new(big.Int).Div(new(big.Int).Set(bPos), gcd).Uint64()
	rMax := new(big.Int).Div(new(big.Int).Set(max), gcd)
	return (&OffHeapIndexer{
		Kind:       OffHeapIndexerMult1024,
		Positions:  pos,
		FPBits:     fpBits,
		Multiplier: multiplier,
		Shift:      uint(rMax.TrailingZeroBits()),
	}).captureConfiguration()
}

// NewInfinitePrecisionMult1024OffHeapIndexer ports the public
// InfinitePrecisionMult1024Indexer constructor. It deliberately retains a
// distinct kind because Java uses signed BigInteger multiplication here,
// whereas Mult1024Indexer requires nonnegative fingerprints and multiplyHigh.
func NewInfinitePrecisionMult1024OffHeapIndexer(positions int64, fpBits int) *OffHeapIndexer {
	fpBits = int(int32(fpBits))
	if fpBits <= 0 || fpBits >= 64 {
		panic(NewAssertionError())
	}
	if !InfinitePrecisionMult1024IndexerIsSupported(positions) {
		panic(NewAssertionError("positions * 8 is not a multiple of 1024MB"))
	}
	if positions < 0 || bits.TrailingZeros64(uint64(positions)) <= fpBits {
		panic(NewAssertionError("fingerprint space is smaller than number of positions"))
	}
	max := new(big.Int).Lsh(big.NewInt(1), uint(64-fpBits))
	bPos := big.NewInt(positions)
	gcd := new(big.Int).GCD(nil, nil, max, bPos)
	multiplier := new(big.Int).Div(new(big.Int).Set(bPos), gcd).Uint64()
	rMax := new(big.Int).Div(new(big.Int).Set(max), gcd)
	return (&OffHeapIndexer{
		Kind:       OffHeapIndexerInfinitePrecisionMult1024,
		Positions:  uint64(positions),
		FPBits:     fpBits,
		Multiplier: multiplier,
		Shift:      uint(rMax.TrailingZeroBits()),
	}).captureConfiguration()
}

func NewInfinitePrecisionOffHeapIndexer(positions int64, fpBits int) *OffHeapIndexer {
	fpBits = int(int32(fpBits))
	if positions < 0 || fpBits <= 0 || fpBits >= 64 {
		panic(NewAssertionError())
	}
	return (&OffHeapIndexer{
		Kind:      OffHeapIndexerInfinitePrecision,
		Positions: uint64(positions),
		FPBits:    fpBits,
	}).captureConfiguration()
}

// NewInfinitePrecisionOffHeapIndexerWithMaxFingerprint is the source's
// three-argument constructor, also used by its bounded verification model.
func NewInfinitePrecisionOffHeapIndexerWithMaxFingerprint(positions int64, fpBits int, maxFingerprint int64) *OffHeapIndexer {
	fpBits = int(int32(fpBits))
	if positions < 0 || fpBits <= 0 || fpBits >= 64 {
		panic(NewAssertionError())
	}
	if maxFingerprint == 0 {
		if positions == 0 {
			panic(NewArithmeticException("Division undefined"))
		}
		panic(NewArithmeticException("Division by zero"))
	}
	factor := new(big.Rat).SetFrac(big.NewInt(positions), big.NewInt(maxFingerprint))
	denominator := new(big.Int).Set(factor.Denom())
	for _, prime := range []int64{2, 5} {
		divisor := big.NewInt(prime)
		for {
			quotient, remainder := new(big.Int), new(big.Int)
			quotient.QuoRem(denominator, divisor, remainder)
			if remainder.Sign() != 0 {
				break
			}
			denominator = quotient
		}
	}
	if denominator.Cmp(big.NewInt(1)) != 0 {
		panic(NewArithmeticException("Non-terminating decimal expansion; no exact representable decimal result."))
	}
	return (&OffHeapIndexer{Kind: OffHeapIndexerInfinitePrecision, Positions: uint64(positions), FPBits: fpBits, factor: factor}).captureConfiguration()
}

func (i *OffHeapIndexer) explicitMaxIdx(fp uint64, probe int) int64 {
	configuration := i.configuration()
	if configuration.positions == 0 {
		if int32(probe) == 0 {
			panic(NewArithmeticException("Division undefined"))
		}
		panic(NewArithmeticException("Division by zero"))
	}
	positions := new(big.Int).SetUint64(configuration.positions)
	scaled := new(big.Rat).Mul(new(big.Rat).SetInt64(int64(fp)), configuration.factor)
	scaled.Add(scaled, new(big.Rat).SetInt64(int64(int32(probe))))
	quotient := new(big.Rat).Quo(scaled, new(big.Rat).SetInt(positions))
	integral := new(big.Int).Quo(quotient.Num(), quotient.Denom())
	remainder := new(big.Rat).Sub(scaled, new(big.Rat).SetInt(new(big.Int).Mul(integral, positions)))
	index := new(big.Int).Quo(remainder.Num(), remainder.Denom()).Int64()
	if index < 0 || uint64(index) >= configuration.positions {
		panic(NewAssertionError())
	}
	return index
}

func (i *OffHeapIndexer) GetIdx(fp uint64) int64 {
	return i.GetIdxProbe(fp, 0)
}

func (i *OffHeapIndexer) GetIdxProbe(fp uint64, probe int) int64 {
	if i == nil {
		panic(NewNullPointerException())
	}
	configuration := i.configuration()
	if configuration.factor != nil {
		return i.explicitMaxIdx(fp, probe)
	}
	var idx uint64
	switch configuration.kind {
	case OffHeapIndexerBitshifting:
		idx = (fp & configuration.prefixMask) >> configuration.rShift
	case OffHeapIndexerMult1024:
		if int64(fp) < 0 {
			panic(NewAssertionError())
		}
		hi, lo := bits.Mul64(fp, configuration.multiplier)
		leftShift := (64 - configuration.shift) & 63
		rightShift := configuration.shift & 63
		idx = (hi << leftShift) | (lo >> rightShift)
		if idx >= configuration.positions {
			panic(NewAssertionError())
		}
		return (int64(idx) + int64(int32(probe))) % int64(configuration.positions)
	case OffHeapIndexerInfinitePrecisionMult1024:
		return i.infinitePrecisionMult1024Idx(fp, probe)
	default:
		return i.infinitePrecisionIdx(fp, probe)
	}
	if probe != 0 {
		idx = uint64((int64(idx) + int64(int32(probe))) % int64(configuration.positions))
	}
	return int64(idx)
}

func (i *OffHeapIndexer) infinitePrecisionIdx(fp uint64, probe int) int64 {
	configuration := i.configuration()
	if configuration.positions == 0 {
		panic(NewArithmeticException("Division undefined"))
	}
	positions := new(big.Int).SetUint64(configuration.positions)
	denominator := new(big.Int).Lsh(big.NewInt(1), uint(64-configuration.fpBits))
	// BigDecimal computes (fp*positions/divisor + probe) remainder positions
	// before truncating to long. Keep the shared denominator so negative
	// fractional values and negative probes retain that exact ordering.
	scaled := big.NewInt(int64(fp))
	scaled.Mul(scaled, positions)
	scaled.Add(scaled, new(big.Int).Mul(big.NewInt(int64(int32(probe))), denominator))
	modulus := new(big.Int).Mul(new(big.Int).Set(positions), denominator)
	scaled.Rem(scaled, modulus)
	scaled.Quo(scaled, denominator)
	idx := scaled.Int64()
	if idx < 0 || uint64(idx) >= configuration.positions {
		panic(NewAssertionError())
	}
	return idx
}

func (i *OffHeapIndexer) infinitePrecisionMult1024Idx(fp uint64, probe int) int64 {
	configuration := i.configuration()
	if configuration.positions == 0 {
		panic(NewArithmeticException("/ by zero"))
	}
	value := big.NewInt(int64(fp))
	value.Mul(value, new(big.Int).SetUint64(configuration.multiplier))
	value.Rsh(value, configuration.shift)
	positions := new(big.Int).SetUint64(configuration.positions)
	value.Rem(value, positions)
	idx := value.Int64() + int64(int32(probe))
	if idx < 0 || uint64(idx) >= configuration.positions {
		panic(NewAssertionError())
	}
	return idx % int64(configuration.positions)
}
