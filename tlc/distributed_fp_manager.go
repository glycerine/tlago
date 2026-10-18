package tlc

import (
	"fmt"
	"math"
	"os"
	"sync"
	"time"
)

// A Java FPSets wrapper is a registration identity, even when two wrappers
// point to the same remote FPSet. Reassignment shares the wrapper, its cached
// hostname and availability across all replaced partitions.
type DistributedFPSets struct {
	set       DistributedFingerprintEndpoint
	hostname  string
	available bool
}

// distributedFPSets retains the internal spelling used by the manager and its
// package-level fault-injection tests while exposing Java's public nested type.
type distributedFPSets = DistributedFPSets

func NewDistributedFPSets(set DistributedFingerprintEndpoint, hostname string) *DistributedFPSets {
	if isNil(set) {
		set = nil
	}
	return &DistributedFPSets{set: set, hostname: hostname, available: true}
}

// Lifecycle operations fail at endpoint access; selection may still skip an
// empty registration slot, and checkpoint/close retain their distinct catches.
func (entry *DistributedFPSets) endpoint() DistributedFingerprintEndpoint {
	if entry == nil || isNil(entry.set) {
		panic(NewNullPointerException())
	}
	return entry.set
}

func (entry *DistributedFPSets) SetUnavailable() {
	if entry == nil {
		panic(NewNullPointerException())
	}
	entry.available = false
}

func (entry *DistributedFPSets) IsAvailable() bool {
	if entry == nil {
		panic(NewNullPointerException())
	}
	return entry.available
}

func (entry *DistributedFPSets) Exit(cleanup bool) error {
	return entry.endpoint().Exit(cleanup)
}

func (entry *DistributedFPSets) Recover(filename string) error {
	return entry.endpoint().RecoverFile(filename)
}

func (entry *DistributedFPSets) CommitChkpt(filename string) error {
	return entry.endpoint().CommitChkptFile(filename)
}

func (entry *DistributedFPSets) BeginChkpt(filename string) error {
	return entry.endpoint().BeginChkptFile(filename)
}

func (entry *DistributedFPSets) GetStatesSeen() (uint64, error) {
	return entry.endpoint().GetStatesSeen()
}

func (entry *DistributedFPSets) Size() (uint64, error) {
	return entry.endpoint().Size()
}

func (entry *DistributedFPSets) ContainsBlock(fingerprints *LongVec) (*BitVector, error) {
	return entry.endpoint().ContainsBlock(fingerprints)
}

func (entry *DistributedFPSets) PutBlock(fingerprints *LongVec) (*BitVector, error) {
	return entry.endpoint().PutBlock(fingerprints)
}

func (entry *DistributedFPSets) Put(fingerprint uint64) (bool, error) {
	return entry.endpoint().Put(fingerprint)
}

func (entry *DistributedFPSets) Contains(fingerprint uint64) (bool, error) {
	return entry.endpoint().Contains(fingerprint)
}

func (entry *DistributedFPSets) GetHostname() string {
	if entry == nil {
		panic(NewNullPointerException())
	}
	return entry.hostname
}

func (entry *DistributedFPSets) GetFPSet() DistributedFingerprintEndpoint {
	if entry == nil {
		panic(NewNullPointerException())
	}
	if isNil(entry.set) {
		return nil
	}
	return entry.set
}

type DistributedFPSetManager struct {
	mu                  sync.RWMutex
	fpSets              []*distributedFPSets
	managerIsBroken     bool
	Mask                uint64
	Description         string
	ExpectedNumServers  int
	NonDistributed      bool
	Trace               *TLCTrace
	modeCaptured        bool
	nonDistributedOwner bool
	descriptionOwner    string
	traceOwner          *TLCTrace
}

func (m *DistributedFPSetManager) require() {
	if m == nil {
		panic(NewNullPointerException())
	}
}

func NewDistributedFPSetManager(sets ...DistributedFingerprintEndpoint) *DistributedFPSetManager {
	m := &DistributedFPSetManager{Mask: math.MaxInt64, modeCaptured: true}
	for _, set := range sets {
		if isNil(set) {
			panic(NewNullPointerException())
		}
		m.fpSets = append(m.fpSets, &distributedFPSets{set: set, hostname: fmt.Sprint(set), available: true})
	}
	return m
}

func NewDynamicDistributedFPSetManager(expectedNumOfServers int) *DistributedFPSetManager {
	expectedNumOfServers = int(int32(expectedNumOfServers))
	if expectedNumOfServers <= 0 {
		panic(NewIllegalArgumentException())
	}
	m := NewDistributedFPSetManager()
	m.ExpectedNumServers = expectedNumOfServers
	log := 0
	for expectedNumOfServers > 0 {
		expectedNumOfServers /= 2
		log++
	}
	m.Mask = (uint64(1) << log) - 1
	return m
}

func NewDistributedFPSetManagerFromFPSet(set FPSet) *DistributedFPSetManager {
	// FPSetManager(FPSetRMI) retains one server reference. A MultiFPSet owns
	// its high-bit child routing; exposing those children as servers would
	// replace it with the manager's low-bit routing and bypass the owner.
	return NewDistributedFPSetManager(NewLocalFingerprintEndpoint(set))
}

func NewNonDistributedFPSetManager(set FPSet, hostname string, trace *TLCTrace) *DistributedFPSetManager {
	// Unlike FPSetManager(FPSetRMI), the source local constructor stores its
	// reference without dereferencing it. Preserve a missing endpoint until
	// an operation actually uses it; counts and empty batches remain valid.
	var endpoint DistributedFingerprintEndpoint
	if !isNil(set) {
		endpoint = NewLocalFingerprintEndpoint(set)
	}
	m := &DistributedFPSetManager{Mask: math.MaxInt64, fpSets: []*distributedFPSets{{set: endpoint, hostname: hostname, available: true}}}
	m.NonDistributed, m.Description, m.Trace = true, hostname, trace
	m.modeCaptured, m.nonDistributedOwner, m.descriptionOwner, m.traceOwner = true, true, hostname, trace
	return m
}

func (m *DistributedFPSetManager) isNonDistributed() bool {
	m.require()
	if m.modeCaptured {
		return m.nonDistributedOwner
	}
	return m.NonDistributed
}

func (m *DistributedFPSetManager) ownedDescription() string {
	m.require()
	if m.modeCaptured && m.nonDistributedOwner {
		return m.descriptionOwner
	}
	return m.Description
}

func (m *DistributedFPSetManager) ownedTrace() *TLCTrace {
	m.require()
	if m.modeCaptured && m.nonDistributedOwner {
		return m.traceOwner
	}
	return m.Trace
}

func (m *DistributedFPSetManager) entries() []*distributedFPSets {
	if m == nil {
		panic(NewNullPointerException())
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]*distributedFPSets(nil), m.fpSets...)
}

// A worker receives its own manager state, with references to the same
// fingerprint servers. Preserve shared registration wrappers within that
// snapshot, while isolating failover/availability changes between processes.
// The coordinator's recovery trace is deliberately omitted.
func (m *DistributedFPSetManager) snapshotForWorker() *DistributedFPSetManager {
	if m == nil {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	snapshot := &DistributedFPSetManager{
		managerIsBroken:     m.managerIsBroken,
		Mask:                m.Mask,
		Description:         m.ownedDescription(),
		ExpectedNumServers:  m.ExpectedNumServers,
		NonDistributed:      m.isNonDistributed(),
		fpSets:              make([]*distributedFPSets, len(m.fpSets)),
		modeCaptured:        true,
		nonDistributedOwner: m.isNonDistributed(),
		descriptionOwner:    m.ownedDescription(),
	}
	wrappers := make(map[*distributedFPSets]*distributedFPSets, len(m.fpSets))
	for i, entry := range m.fpSets {
		if entry == nil {
			continue
		}
		copy := wrappers[entry]
		if copy == nil {
			copy = &distributedFPSets{set: entry.set, hostname: entry.hostname, available: entry.available}
			wrappers[entry] = copy
		}
		snapshot.fpSets[i] = copy
	}
	return snapshot
}

// registrationAt permits an empty slot during checkpoint and shutdown traversal.
func (m *DistributedFPSetManager) registrationAt(index int) *distributedFPSets {
	if m == nil {
		panic(NewNullPointerException())
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if index < 0 || index >= len(m.fpSets) {
		panic(NewIndexOutOfBoundsException(index, len(m.fpSets)))
	}
	return m.fpSets[index]
}

func (m *DistributedFPSetManager) entry(index int) *distributedFPSets {
	entry := m.registrationAt(index)
	if entry == nil {
		panic(NewNullPointerException())
	}
	return entry
}

func (m *DistributedFPSetManager) NumOfServers() int {
	m.require()
	if m.isNonDistributed() {
		return 1
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.fpSets)
}

func (m *DistributedFPSetManager) NumOfAliveServers() int {
	m.require()
	if m.isNonDistributed() {
		return 1
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	seen := make(map[*distributedFPSets]bool, len(m.fpSets))
	count := 0
	for _, entry := range m.fpSets {
		if entry == nil {
			panic(NewNullPointerException())
		}
		if !seen[entry] && entry.available {
			count++
		}
		seen[entry] = true
	}
	return count
}

func (m *DistributedFPSetManager) Register(set DistributedFingerprintEndpoint, hostname ...string) error {
	m.require()
	if len(hostname) > 1 {
		panic(NewIllegalArgumentException("Register accepts at most one hostname"))
	}
	if m.isNonDistributed() {
		panic(NewUnsupportedOperationException("Not applicable for non-distributed FPSetManager"))
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ExpectedNumServers > 0 && len(m.fpSets) >= m.ExpectedNumServers {
		return NewFPSetManagerException(fmt.Sprintf("Limit for FPset servers reached (%d). Cannot handle additional servers", m.ExpectedNumServers))
	}
	if isNil(set) {
		set = nil
	}
	var name string
	if len(hostname) > 0 {
		name = hostname[0]
	} else {
		name = fmt.Sprint(set)
	}
	m.fpSets = append(m.fpSets, &distributedFPSets{set: set, hostname: name, available: true})
	return nil
}

func (m *DistributedFPSetManager) RegisterFPSet(set DistributedFingerprintEndpoint, hostname ...string) error {
	m.require()
	if len(hostname) > 1 {
		panic(NewIllegalArgumentException("RegisterFPSet accepts at most one hostname"))
	}
	return m.Register(set, hostname...)
}

func (m *DistributedFPSetManager) GetMask() uint64 {
	m.require()
	return m.Mask
}

func (m *DistributedFPSetManager) GetHostName() string {
	m.require()
	if m.isNonDistributed() {
		return m.ownedDescription()
	}
	if host, err := distributedLocalHostName(); err == nil {
		return host
	} else {
		fmt.Fprint(os.Stderr, javaThrowableStackTrace(err))
	}
	return "Unknown"
}

func (m *DistributedFPSetManager) GetFPSetIndex(fp uint64) int {
	m.require()
	if m.isNonDistributed() {
		return 0
	}
	count := m.NumOfServers()
	if count == 0 {
		panic(NewArithmeticException("/ by zero"))
	}
	return int((fp & m.Mask) % uint64(count))
}

func (m *DistributedFPSetManager) Reassign(index int) int {
	m.require()
	index = int(int32(index))
	m.mu.Lock()
	defer m.mu.Unlock()
	count := len(m.fpSets)
	if index < 0 || index >= count {
		panic(NewIllegalArgumentException("index not within bounds"))
	}
	if m.managerIsBroken {
		return -1
	}
	broken := m.fpSets[index]
	if broken == nil {
		panic(NewNullPointerException())
	}
	broken.SetUnavailable()
	for next := (index + 1) % count; next != index; next = (next + 1) % count {
		replacement := m.fpSets[next]
		if replacement == nil {
			panic(NewNullPointerException())
		}
		if replacement.IsAvailable() {
			// Preserve Java's non-wrapping assignment loop, even when the
			// selected successor's index wrapped around to the list's start.
			for j := index; j < next; j++ {
				m.fpSets[j] = replacement
			}
			return next
		}
	}
	m.managerIsBroken = true
	return -1
}

// Go FPSet methods use panic for Java's unchecked/remote failure paths. A
// catch(Exception) catches runtime exceptions but does not catch Java Error.
func tryFPSetCall[T any](call func() T) (value T, err error) {
	defer func() {
		if failure := recover(); failure != nil {
			err = panicValueAsError(failure)
			if isJavaError(err) {
				panic(failure)
			}
		}
	}()
	return call(), nil
}

func (m *DistributedFPSetManager) failover(index int, failure error) bool {
	message := "null"
	if detail := javaThrowableDetailMessage(failure); detail != nil {
		message = *detail
	}
	ToolIOPrintln(fmt.Sprintf("Warning: Failed to connect from %s to the fp server at %s.\n%s", m.GetHostName(), m.entry(index).GetHostname(), message))
	if m.Reassign(index) == -1 {
		ToolIOPrintln("Warning: there is no fp server available.")
		return false
	}
	return true
}

func (m *DistributedFPSetManager) Put(fp uint64) bool {
	m.require()
	return m.scalarCall(fp, true)
}

func (m *DistributedFPSetManager) Contains(fp uint64) bool {
	m.require()
	return m.scalarCall(fp, false)
}

func (m *DistributedFPSetManager) scalarCall(fp uint64, put bool) bool {
	index := m.GetFPSetIndex(fp)
	for {
		result, err := invokeFingerprintEndpoint(func() (bool, error) {
			if put {
				return m.entry(index).Put(fp)
			}
			return m.entry(index).Contains(fp)
		})
		if err == nil {
			return result
		}
		if m.isNonDistributed() {
			if !isJavaIOException(err) {
				panic(err)
			}
			PrintError(ECGeneral, generalErrorParams("", err)...)
			return false
		}
		if !m.failover(index, err) {
			return false
		}
	}
}

func fpBlockAt(fingerprints []*LongVec, index int) *LongVec {
	if fingerprints == nil {
		panic(NewNullPointerException())
	}
	if index < 0 || index >= len(fingerprints) {
		panic(NewArrayIndexOutOfBoundsException(index, len(fingerprints)))
	}
	return fingerprints[index]
}

func (m *DistributedFPSetManager) blockCall(fingerprints []*LongVec, index int, put bool) *BitVector {
	for {
		result, err := invokeFingerprintEndpoint(func() (*BitVector, error) {
			entryIndex := index
			if m.isNonDistributed() {
				entryIndex = 0
			}
			fpv := fpBlockAt(fingerprints, index)
			if put {
				return m.entry(entryIndex).PutBlock(fpv)
			}
			return m.entry(entryIndex).ContainsBlock(fpv)
		})
		if err == nil {
			return result
		}
		if m.isNonDistributed() {
			if !isJavaIOException(err) {
				panic(err)
			}
			PrintError(ECGeneral, generalErrorParams("", err)...)
			return NewBitVector(0)
		}
		if !m.failover(index, err) {
			fpv := fpBlockAt(fingerprints, index)
			if fpv == nil {
				panic(NewNullPointerException())
			}
			return NewBitVectorWithValue(fpv.Size(), true)
		}
	}
}

func (m *DistributedFPSetManager) ContainsBlock(fingerprints []*LongVec, executor ...*DistributedExecutor) []*BitVector {
	m.require()
	if len(executor) > 1 {
		panic(NewIllegalArgumentException("ContainsBlock accepts at most one executor"))
	}
	return m.blocks(fingerprints, false, executor)
}

func (m *DistributedFPSetManager) PutBlock(fingerprints []*LongVec, executor ...*DistributedExecutor) []*BitVector {
	m.require()
	if len(executor) > 1 {
		panic(NewIllegalArgumentException("PutBlock accepts at most one executor"))
	}
	return m.blocks(fingerprints, true, executor)
}

type BitVectorWrapper struct {
	index     int
	bitVector *BitVector
}

func NewBitVectorWrapper(index int, bitVector *BitVector) *BitVectorWrapper {
	return &BitVectorWrapper{index: int(int32(index)), bitVector: bitVector}
}

func (w *BitVectorWrapper) GetBitVector() *BitVector {
	if w == nil {
		panic(NewNullPointerException())
	}
	return w.bitVector
}

func (w *BitVectorWrapper) GetIndex() int {
	if w == nil {
		panic(NewNullPointerException())
	}
	return w.index
}

type distributedBitVectorResult struct {
	wrapper *BitVectorWrapper
	err     error
}

var distributedFPSetRetryRandom = NewJavaRandomDefault()

func (m *DistributedFPSetManager) blocks(fingerprints []*LongVec, put bool, executor []*DistributedExecutor) []*BitVector {
	count := m.NumOfServers()
	if m.isNonDistributed() {
		if fingerprints == nil {
			panic(NewNullPointerException())
		}
		count = len(fingerprints)
	}
	if m.isNonDistributed() || len(executor) == 0 {
		out := make([]*BitVector, count)
		for i := range count {
			out[i] = m.blockCall(fingerprints, i, put)
		}
		return out
	}
	es := executor[0]
	if es == nil {
		panic(NewNullPointerException())
	}
	completed := make(chan distributedBitVectorResult, count)
	retry := 0
	for index := 0; index < count; index++ {
		_, err := tryFPSetCall(func() bool {
			if es == nil {
				panic(NewNullPointerException())
			}
			es.submit(func() {
				result := distributedBitVectorResult{}
				defer func() {
					if failure := recover(); failure != nil {
						result.err = NewExecutionException(panicValueAsError(failure))
					}
					completed <- result
				}()
				result.wrapper = NewBitVectorWrapper(index, m.blockCall(fingerprints, index, put))
			})
			return true
		})
		if err == nil {
			retry = 0
			continue
		}
		if _, rejected := err.(*RejectedExecutionException); !rejected {
			panic(err)
		}
		oldRetry := retry
		retry++
		if oldRetry < 3 && !es.IsShutdown() {
			time.Sleep(time.Duration(1+distributedFPSetRetryRandom.NextIntN(5)) * time.Second)
			index--
			continue
		}
		panic(err)
	}
	out := make([]*BitVector, count)
	for range count {
		result := <-completed
		if result.err != nil {
			fmt.Fprint(os.Stderr, javaThrowableStackTrace(result.err))
			continue
		}
		index := result.wrapper.GetIndex()
		if out[index] != nil {
			panic(NewTLCRuntimeException(ECGeneral))
		}
		out[index] = result.wrapper.GetBitVector()
	}
	return out
}

func (m *DistributedFPSetManager) Size() uint64 {
	m.require()
	return m.sumStatistics(false)
}

func (m *DistributedFPSetManager) GetStatesSeen() uint64 {
	m.require()
	return m.sumStatistics(true)
}

func (m *DistributedFPSetManager) sumStatistics(statesSeen bool) uint64 {
	var total uint64
	if statesSeen && !m.isNonDistributed() {
		total = 1
	}
	count := m.NumOfServers()
	for index := range count {
		value, err := invokeFingerprintEndpoint(func() (uint64, error) {
			if statesSeen && !m.isNonDistributed() {
				return m.entry(index).GetStatesSeen()
			}
			return m.entry(index).Size()
		})
		if err == nil {
			total += value
			continue
		}
		if m.isNonDistributed() {
			if !isJavaIOException(err) {
				panic(err)
			}
			PrintError(ECGeneral, generalErrorParams("", err)...)
			return math.MaxUint64
		}
		// Java reassigns on a statistics failure, but does not retry this slot.
		m.failover(index, err)
	}
	return total
}

type distributedCheckResult[T any] struct {
	value T
	err   error
}

func submitDistributedCheck[T any](executor *DistributedExecutor, results chan<- distributedCheckResult[T], call func() (T, error)) {
	executor.submit(func() {
		result := distributedCheckResult[T]{}
		defer func() {
			if failure := recover(); failure != nil {
				result.err = NewExecutionException(panicValueAsError(failure))
			}
			results <- result
		}()
		value, err := call()
		result.value = value
		if err != nil {
			result.err = NewExecutionException(err)
		}
	})
}

// The source check callables catch IOException before the executor can wrap
// it. Other failures still reach the manager as failed task completions.
func invokeDistributedFingerprintCheck[T any](failureValue T, call func() (T, error)) (T, error) {
	value, err := invokeFingerprintEndpoint(call)
	if err != nil && isJavaIOException(err) {
		PrintError(ECGeneral, generalErrorParams("", err)...)
		return failureValue, nil
	}
	return value, err
}

func (m *DistributedFPSetManager) CheckFPs() uint64 {
	m.require()
	if m.isNonDistributed() {
		value, err := invokeFingerprintEndpoint(func() (uint64, error) {
			set := m.entry(0).set
			if set == nil {
				panic(NewNullPointerException())
			}
			return set.CheckFPs()
		})
		if err != nil {
			if !isJavaIOException(err) {
				panic(err)
			}
			PrintError(ECGeneral, generalErrorParams("", err)...)
			return math.MaxUint64
		}
		return value
	}
	count := m.NumOfServers()
	if count == 0 {
		panic(NewIllegalArgumentException())
	}
	executor := NewDistributedExecutor()
	defer executor.Shutdown()
	results := make(chan distributedCheckResult[uint64], count)
	for index := range count {
		set := m.entry(index).set
		submitDistributedCheck(executor, results, func() (uint64, error) {
			return invokeDistributedFingerprintCheck(uint64(math.MaxInt64), func() (uint64, error) {
				if set == nil {
					panic(NewNullPointerException())
				}
				return set.CheckFPs()
			})
		})
	}
	value := uint64(math.MaxInt64)
	for range count {
		result := <-results
		if result.err != nil {
			fmt.Fprint(os.Stderr, javaThrowableStackTrace(result.err))
			continue
		}
		value = javaLongMinBits(value, result.value)
	}
	return value
}

func (m *DistributedFPSetManager) CheckInvariant(expectFPs ...uint64) bool {
	m.require()
	requireInvariantExpectedCount(expectFPs)
	if m.isNonDistributed() {
		value, err := invokeFingerprintEndpoint(func() (bool, error) {
			set := m.entry(0).set
			if set == nil {
				panic(NewNullPointerException())
			}
			return set.CheckInvariant(expectFPs...)
		})
		if err != nil {
			if !isJavaIOException(err) {
				panic(err)
			}
			PrintError(ECGeneral, generalErrorParams("", err)...)
			return false
		}
		return value
	}
	count := m.NumOfServers()
	if count == 0 {
		panic(NewIllegalArgumentException())
	}
	executor := NewDistributedExecutor()
	defer executor.Shutdown()
	results := make(chan distributedCheckResult[bool], count)
	for index := range count {
		set := m.entry(index).set
		submitDistributedCheck(executor, results, func() (bool, error) {
			return invokeDistributedFingerprintCheck(false, func() (bool, error) {
				if set == nil {
					panic(NewNullPointerException())
				}
				return set.CheckInvariant(expectFPs...)
			})
		})
	}
	for range count {
		result := <-results
		if result.err != nil {
			fmt.Fprint(os.Stderr, javaThrowableStackTrace(result.err))
			continue
		}
		if !result.value {
			return false
		}
	}
	return true
}

// firstRegistration and lastRegistration retain wrapper identity and read live
// registrations. Callers fix the traversal length before invoking endpoints.
func (m *DistributedFPSetManager) firstRegistration(length int) (int, *distributedFPSets) {
	for index := 0; index < length; index++ {
		if entry := m.registrationAt(index); entry != nil {
			return index, entry
		}
	}
	return length, nil
}

func (m *DistributedFPSetManager) lastRegistration(length, first int, current *distributedFPSets) int {
	last := length - 1
	for ; last > first; last-- {
		if entry := m.registrationAt(last); entry != nil && entry != current {
			break
		}
	}
	return last
}

func exitFingerprintRegistration(entry *distributedFPSets, cleanup bool) {
	_, failure := invokeFingerprintEndpoint(func() (struct{}, error) {
		return struct{}{}, entry.Exit(cleanup)
	})
	if failure != nil {
		if operation, ok := failure.(*DistributedOperationError); !ok || !operation.ExitIgnorable {
			fmt.Fprint(os.Stderr, javaThrowableStackTrace(failure))
		}
	}
}

func (m *DistributedFPSetManager) Close(cleanup bool) error {
	m.require()
	if m.isNonDistributed() {
		if err := m.entry(0).endpoint().Close(); err != nil {
			return err
		}
		return m.entry(0).endpoint().Exit(cleanup)
	}
	length := m.NumOfServers()
	first, current := m.firstRegistration(length)
	if current == nil {
		return nil
	}
	last := m.lastRegistration(length, first, current)
	for index := first + 1; index <= last; index++ {
		next := m.registrationAt(index)
		if next != nil && next != current {
			// Capture next before exit: exit can change later registrations.
			exitFingerprintRegistration(current, cleanup)
			current = next
		}
	}
	exitFingerprintRegistration(current, cleanup)
	return nil
}

func (m *DistributedFPSetManager) Checkpoint(filename string) error {
	m.require()
	if m.isNonDistributed() {
		return m.entry(0).endpoint().BeginChkpt()
	}
	return m.checkpointInner(filename, true)
}

func (m *DistributedFPSetManager) CommitCheckpoint() error {
	m.require()
	if m.isNonDistributed() {
		return m.entry(0).endpoint().CommitChkpt()
	}
	return nil
}

func (m *DistributedFPSetManager) Recover(filename string) error {
	m.require()
	if m.isNonDistributed() {
		return m.entry(0).endpoint().RecoverTrace(m.ownedTrace())
	}
	return m.checkpointInner(filename, false)
}

func (m *DistributedFPSetManager) checkpointInner(filename string, checkpoint bool) error {
	length := m.NumOfServers()
	first, current := m.firstRegistration(length)
	if current == nil {
		return nil
	}
	// The source runs the first checkpoint before selecting the trailing boundary.
	if err := m.checkpointRegistration(first, filename, checkpoint); err != nil {
		return err
	}
	last := m.lastRegistration(length, first, current)
	for index := first + 1; index <= last; index++ {
		next := m.registrationAt(index)
		if next != nil && next != current {
			current = next
			if err := m.checkpointRegistration(index, filename, checkpoint); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *DistributedFPSetManager) checkpointRegistration(index int, filename string, checkpoint bool) error {
	// Each phase resolves its registration separately, including error reporting.
	// Source checkpoint work is synchronous; no extra goroutine is needed.
	_, err := invokeFingerprintEndpoint(func() (struct{}, error) {
		if checkpoint {
			if err := m.entry(index).BeginChkpt(filename); err != nil {
				return struct{}{}, err
			}
			return struct{}{}, m.entry(index).CommitChkpt(filename)
		}
		return struct{}{}, m.entry(index).Recover(filename)
	})
	if err != nil {
		if !isJavaIOException(err) {
			return err
		}
		ToolIOPrintln(fmt.Sprintf("Error: Failed to checkpoint the fingerprint server at %s. This server might be down.", m.entry(index).GetHostname()))
	}
	return nil
}
