package tlc

import "testing"

func TestJavaEnumerableValue(t *testing.T) {
	oldSeed := RandomEnumerableSeed()
	oldRandom := ResetRandomEnumerableValues()
	SetRandomEnumerableSeed(15041980)
	t.Cleanup(func() {
		SetRandomEnumerableSeed(oldSeed)
		SetRandomEnumerableGenerator(oldRandom)
	})

	indices := make(map[int]struct{})
	for n := 1; n < 10657; n++ {
		// Java DummyValue(n).elements(n) constructs SubsetEnumerator(n,
		// size() == n). Its other abstract-method stubs are never called.
		enumerator := newRandomSubsetIndices(n, n)
		for enumerator.hasNext() {
			index := enumerator.nextIndex()
			if !(0 <= index && index < n) {
				t.Fatalf("Index %d out of bounds for n=%d", index, n)
			}
			indices[index] = struct{}{}
		}
		if len(indices) != n {
			t.Fatalf("Missing indices for n=%d: got %d, want %d", n, len(indices), n)
		}
		clear(indices)
	}
}
