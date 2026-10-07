package tlc

import (
	"math"
	"sync"
	"sync/atomic"
	"time"
)

const (
	javaRandomMultiplier = int64(0x5DEECE66D)
	javaRandomAddend     = int64(0xB)
	javaRandomMask       = int64((1 << 48) - 1)
)

type JavaRandom struct {
	mu               sync.Mutex
	seed             int64
	aril             int64
	haveNextGaussian bool
	nextGaussian     float64
}

var javaRandomPrimes = generateJavaRandomPrimes()
var javaRandomSeedUniquifierValue int64 = 8682522807148012

func NewJavaRandom(seed int64) *JavaRandom {
	r := &JavaRandom{}
	r.SetSeed(seed)
	return r
}

func NewJavaRandomDefault() *JavaRandom {
	return NewJavaRandom(javaRandomSeedUniquifier() ^ time.Now().UnixNano())
}

func javaRandomSeedUniquifier() int64 {
	for {
		current := atomic.LoadInt64(&javaRandomSeedUniquifierValue)
		next := current * 1181783497276652981
		if atomic.CompareAndSwapInt64(&javaRandomSeedUniquifierValue, current, next) {
			return next
		}
	}
}

func (r *JavaRandom) SetSeed(seed int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seed = (seed ^ javaRandomMultiplier) & javaRandomMask
	r.aril = 0
	r.haveNextGaussian = false
}

func (r *JavaRandom) SetSeedWithAril(seed int64, cnt int64) {
	r.SetSeed(seed)
	for cnt > 0 {
		r.NextDouble()
		cnt--
	}
}

func (r *JavaRandom) Aril() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.aril
}

func (r *JavaRandom) Next(bits int) int32 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.nextLocked(bits)
}

func (r *JavaRandom) nextLocked(bits int) int32 {
	r.seed = (r.seed*javaRandomMultiplier + javaRandomAddend) & javaRandomMask
	return int32(uint64(r.seed) >> uint(48-bits))
}

func (r *JavaRandom) NextBytes(bytes []byte) {
	for i := 0; i < len(bytes); {
		rnd := r.NextInt()
		for n := min(len(bytes)-i, 4); n > 0; n-- {
			bytes[i] = byte(rnd)
			i++
			rnd >>= 8
		}
	}
}

func (r *JavaRandom) NextInt() int32 {
	return r.Next(32)
}

func (r *JavaRandom) NextIntN(bound int32) int32 {
	if bound <= 0 {
		panic("bound must be positive")
	}
	if bound&-bound == bound {
		return int32((int64(bound) * int64(r.Next(31))) >> 31)
	}
	for {
		bits := r.Next(31)
		val := bits % bound
		if bits-val+(bound-1) >= 0 {
			return val
		}
	}
}

func (r *JavaRandom) NextLong() int64 {
	hi := int64(r.Next(32))
	lo := int64(r.Next(32))
	return (hi << 32) + lo
}

func (r *JavaRandom) NextFloat() float32 {
	return float32(r.Next(24)) / float32(1<<24)
}

func (r *JavaRandom) NextDouble() float64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.aril++
	hi := int64(r.nextLocked(26))
	lo := int64(r.nextLocked(27))
	return float64((hi<<27)+lo) / float64(1<<53)
}

func (r *JavaRandom) NextGaussian() float64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.haveNextGaussian {
		r.haveNextGaussian = false
		return r.nextGaussian
	}
	var v1, v2, s float64
	for {
		r.aril += 2
		hi1 := int64(r.nextLocked(26))
		lo1 := int64(r.nextLocked(27))
		u1 := float64((hi1<<27)+lo1) / float64(1<<53)
		hi2 := int64(r.nextLocked(26))
		lo2 := int64(r.nextLocked(27))
		u2 := float64((hi2<<27)+lo2) / float64(1<<53)
		v1 = 2*u1 - 1
		v2 = 2*u2 - 1
		s = v1*v1 + v2*v2
		if s < 1 && s != 0 {
			break
		}
	}
	multiplier := math.Sqrt(-2 * math.Log(s) / s)
	r.nextGaussian = v2 * multiplier
	r.haveNextGaussian = true
	return v1 * multiplier
}

func (r *JavaRandom) NextPrime() int {
	index := len(javaRandomPrimes)
	for index == len(javaRandomPrimes) {
		index = int(math.Floor(r.NextDouble() * float64(index)))
	}
	return javaRandomPrimes[index]
}

func generateJavaRandomPrimes() []int {
	// Java RandomGenerator.primes is exactly the ascending list of all primes in
	// this interval. Generate it to avoid carrying a large static table in Go.
	primes := make([]int, 0, 1341)
	for n := 1277011; n <= 1295953; n++ {
		if isJavaRandomPrime(n) {
			primes = append(primes, n)
		}
	}
	return primes
}

func isJavaRandomPrime(n int) bool {
	if n < 2 {
		return false
	}
	if n%2 == 0 {
		return n == 2
	}
	for d := 3; d*d <= n; d += 2 {
		if n%d == 0 {
			return false
		}
	}
	return true
}

func (r *JavaRandom) Perm(n int) []int {
	perm := make([]int, n)
	for i := range perm {
		perm[i] = i
	}
	for i := n; i > 1; i-- {
		j := int(r.NextIntN(int32(i)))
		perm[i-1], perm[j] = perm[j], perm[i-1]
	}
	return perm
}

var randomEnumerableValues = struct {
	sync.Mutex
	seed    int64
	threads map[uint64]*randomEnumerableThreadState
}{
	threads: make(map[uint64]*randomEnumerableThreadState),
}

type randomEnumerableThreadState struct {
	rng      *JavaRandom
	rngState *TLCStateMut
}

func RandomEnumerableSeed() int64 {
	randomEnumerableValues.Lock()
	defer randomEnumerableValues.Unlock()
	return randomEnumerableValues.seed
}

func SetRandomEnumerableSeed(seed int64) {
	randomEnumerableValues.Lock()
	randomEnumerableValues.seed = seed
	randomEnumerableValues.Unlock()
	// Java setSeed calls reset, which removes only the current thread's RNG.
	ResetRandomEnumerableValues()
}

func ResetRandomEnumerableValues() *JavaRandom {
	// Source reset obtains/initializes the prior RNG before ThreadLocal.remove.
	old := RandomEnumerableGenerator()
	randomEnumerableValues.Lock()
	defer randomEnumerableValues.Unlock()
	state := randomEnumerableThreadStateForLocked(currentGoroutineID())
	state.rng = nil
	state.rngState = nil
	// IdThread's current predecessor lives in currentStateScope, independently
	// of this RNG entry. Removing the RNG must not discard that state scope.
	return old
}

func SetRandomEnumerableGenerator(rng *JavaRandom) *JavaRandom {
	randomEnumerableValues.Lock()
	defer randomEnumerableValues.Unlock()
	if rng == nil {
		rng = NewJavaRandom(randomEnumerableValues.seed)
	}
	gid := currentGoroutineID()
	state := randomEnumerableThreadStateForLocked(gid)
	old := state.rng
	state.rng = rng
	state.rngState = nil
	return old
}

func PushRandomEnumerableState(state *TLCStateMut) func() {
	// Both random enumeration and checker error handling use IdThread's one
	// current-state slot. Do not create a second predecessor scope for RNGs.
	return PushCurrentState(state)
}

func RandomEnumerableGenerator() *JavaRandom {
	modelChecking := MainChecker() != nil && CurrentSimulator() == nil
	currentState, _ := CurrentState()
	randomEnumerableValues.Lock()
	defer randomEnumerableValues.Unlock()
	threadState := randomEnumerableThreadStateForLocked(currentGoroutineID())
	if threadState.rng == nil {
		threadState.rng = NewJavaRandom(randomEnumerableValues.seed)
	}
	if modelChecking && currentState != nil && threadState.rngState != currentState {
		seed := int64(currentState.FingerPrint()) ^ randomEnumerableValues.seed
		threadState.rng.SetSeed(seed)
		threadState.rngState = currentState
	}
	return threadState.rng
}

func randomEnumerableThreadStateForLocked(gid uint64) *randomEnumerableThreadState {
	if randomEnumerableValues.threads == nil {
		randomEnumerableValues.threads = make(map[uint64]*randomEnumerableThreadState)
	}
	state := randomEnumerableValues.threads[gid]
	if state == nil {
		state = &randomEnumerableThreadState{rng: NewJavaRandom(randomEnumerableValues.seed)}
		randomEnumerableValues.threads[gid] = state
	}
	return state
}
