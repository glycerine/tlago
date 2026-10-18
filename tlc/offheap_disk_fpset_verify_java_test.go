//go:build tlc_verify

package tlc

import (
	"fmt"
	"math/bits"
	"strings"
	"testing"
)

// Original OffHeapDiskFPSetJPFTest.test and its Dummy model. Explore all
// interleavings at shared reads, synchronized trySet, and atomic increments.
// This verifies that source model, not production OffHeapDiskFPSet storage.
func TestJavaOffHeapDiskFPSetVerificationModel(t *testing.T) {
	const size, maximum, probeLimit = 3, 3, 2
	const read, compare, increment = 0, 1, 2
	type writer struct {
		fingerprint, probe, phase, position, expected uint8
	}
	type state struct {
		array                         [size]uint8
		writers                       [2]writer
		tableCount                    uint8
		scanIndex, scanPhase          uint8
		scanCount, scanDistinctValues uint8
	}
	type predecessor struct {
		previous state
		actor    int
	}
	indexer := NewInfinitePrecisionOffHeapIndexerWithMaxFingerprint(size, 1, maximum)
	initial := state{writers: [2]writer{{fingerprint: 1}, {fingerprint: 1}}}
	seen := map[state]predecessor{initial: {actor: -1}}
	queue := []state{initial}
	var transitions, terminal int
	witness := func(current state) string {
		var path []string
		for entry := seen[current]; entry.actor >= 0; entry = seen[current] {
			path = append(path, fmt.Sprintf("actor %d -> %+v", entry.actor, current))
			current = entry.previous
		}
		for left, right := 0, len(path)-1; left < right; left, right = left+1, right-1 {
			path[left], path[right] = path[right], path[left]
		}
		return strings.Join(path, "\n")
	}
	for head := 0; head < len(queue); head++ {
		current := queue[head]
		if current.writers[0].fingerprint > maximum && current.writers[1].fingerprint > maximum && current.scanIndex > size {
			terminal++
		}
		// Both workers are started before the source invariant call. Histories
		// where worker zero runs before worker one starts remain represented by
		// taking worker-zero transitions before any worker-one/scan transition.
		for actor := 0; actor < 3; actor++ {
			next := current
			if actor < 2 {
				worker := &next.writers[actor]
				if worker.fingerprint > maximum {
					continue
				}
				advance := func() { *worker = writer{fingerprint: worker.fingerprint + 1} }
				switch worker.phase {
				case read:
					worker.position = uint8(indexer.GetIdxProbe(uint64(worker.fingerprint), int(worker.probe)))
					worker.expected = next.array[worker.position]
					if worker.expected == 0 {
						worker.phase = compare
					} else if worker.expected == worker.fingerprint {
						advance() // memInsert returns true; caller ignores it.
					} else {
						worker.probe++
						if worker.probe == probeLimit {
							advance() // Exhausted probes return false, not failure.
						}
					}
				case compare:
					// trySet holds the array monitor through comparison/write.
					// It releases that monitor before incrementAndGet.
					if next.array[worker.position] == worker.expected {
						next.array[worker.position] = worker.fingerprint
						worker.phase = increment
					} else {
						worker.phase = read // i-- followed by loop increment retries.
					}
				case increment:
					next.tableCount++
					advance()
				}
			} else {
				// checkInvariant locks fpSet, a different monitor from the
				// array's trySet monitor. Writers can run during its scan.
				if next.scanIndex > size {
					continue
				}
				if next.scanIndex == size {
					if int(next.scanCount) != bits.OnesCount8(next.scanDistinctValues) {
						t.Fatalf("FPSet violates its invariant: %v\n%s", next.array, witness(current))
					}
					next.scanIndex++
				} else if next.scanPhase == read {
					if next.array[next.scanIndex] > 0 {
						next.scanPhase = compare
					} else {
						next.scanIndex++
					}
				} else {
					// Preserve the second source array read for Set.add.
					next.scanDistinctValues |= 1 << next.array[next.scanIndex]
					next.scanCount++
					next.scanIndex++
					next.scanPhase = read
				}
			}
			transitions++
			if _, present := seen[next]; !present {
				seen[next] = predecessor{previous: current, actor: actor}
				queue = append(queue, next)
			}
		}
	}
	t.Logf("Original bounded model: %d reachable states, %d transitions, %d terminal states", len(seen), transitions, terminal)
}
