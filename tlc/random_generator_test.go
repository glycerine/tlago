package tlc

import "testing"

func TestRandomEnumerableModelCheckingUsesSingleJavaStyleRNG(t *testing.T) {
	oldChecker := MainChecker()
	oldSimulator := CurrentSimulator()
	oldRandom := ResetRandomEnumerableValues()
	oldSeed := RandomEnumerableSeed()
	defer func() {
		SetMainChecker(oldChecker)
		SetSimulator(oldSimulator)
		SetRandomEnumerableSeed(oldSeed)
		SetRandomEnumerableGenerator(oldRandom)
	}()

	seed := int64(17)
	SetRandomEnumerableSeed(seed)
	SetMainChecker(&ModelChecker{})
	SetSimulator(nil)

	state := NewEmptyState()
	for _, variable := range StateVariables() {
		state.Bind(variable.Name, IntZero)
	}
	fpSeed := int64(state.FingerPrint()) ^ seed
	expected := NewJavaRandom(seed)
	_ = expected.NextDouble()
	expected.SetSeed(fpSeed)
	inStateWant := expected.NextDouble()
	afterStateWant := expected.NextDouble()

	first := RandomEnumerableGenerator()
	_ = first.NextDouble()
	restore := PushRandomEnumerableState(state)
	inStateGot := RandomEnumerableGenerator().NextDouble()
	restore()
	afterStateGot := RandomEnumerableGenerator().NextDouble()

	if inStateGot != inStateWant {
		t.Fatalf("state-seeded draw = %.17g, want %.17g", inStateGot, inStateWant)
	}
	if afterStateGot != afterStateWant {
		t.Fatalf("post-state draw = %.17g, want continued Java state RNG %.17g", afterStateGot, afterStateWant)
	}
}
