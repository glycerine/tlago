package tlc

import (
	"os"
	"os/exec"
	"testing"
)

// The source has no direct selector-startup tests. Separate processes preserve
// the lifetime of startup settings without resetting production state.
func TestDistributedSelectorStartupCapture(t *testing.T) {
	if scenario := os.Getenv("TLAGO_SELECTOR_STARTUP_CASE"); scenario != "" {
		checkDistributedSelectorStartupCapture(t, scenario)
		return
	}
	for _, scenario := range []string{"default", "static", "unlimiting", "limiting", "invalid-booleans", "deferred-static", "failed-factory", "failed-static"} {
		t.Run(scenario, func(t *testing.T) {
			command := exec.Command(os.Args[0], "-test.run=^TestDistributedSelectorStartupCapture$", "-test.v")
			command.Env = append(os.Environ(), "TLAGO_SELECTOR_STARTUP_CASE="+scenario)
			if output, err := command.CombinedOutput(); err != nil {
				t.Fatalf("selector startup process failed: %v\n%s", err, output)
			}
		})
	}
}

func checkDistributedSelectorStartupCapture(t *testing.T, scenario string) {
	set := func(name, value string) { tlcSetStartupSystemProperty(name, value) }
	set(distributedSelectorStaticProperty, "false")
	set(distributedSelectorUnlimitingProperty, "false")
	set(distributedSelectorLimitingProperty, "false")
	set(distributedStaticBlockSizeProperty, "0x11")
	server := &TLCServer{}
	want := BlockSelectorStatistical
	switch scenario {
	case "static", "failed-factory":
		set(distributedSelectorStaticProperty, "TrUe")
		set(distributedSelectorUnlimitingProperty, "true")
		set(distributedSelectorLimitingProperty, "true")
		want = BlockSelectorStatic
	case "unlimiting":
		set(distributedSelectorUnlimitingProperty, "true")
		set(distributedSelectorLimitingProperty, "true")
		want = BlockSelectorProportional
	case "limiting":
		set(distributedSelectorLimitingProperty, "TRUE")
		want = BlockSelectorLimiting
	case "invalid-booleans":
		set(distributedSelectorStaticProperty, " true")
		set(distributedSelectorUnlimitingProperty, "1")
		set(distributedSelectorLimitingProperty, "yes")
	}
	if scenario == "failed-static" {
		err := invokeDistributedServerOperation(func() error { NewStaticBlockSelector(nil); return nil })
		if failure, ok := err.(*TLCError); !ok || !failure.Runtime || failure.Error() != "TLC found a null TLCServer" {
			t.Fatalf("missing server constructor failure: %T/%v", err, err)
		}
		set(distributedStaticBlockSizeProperty, "33")
		if got := NewStaticBlockSelector(server).StaticBlockSize; got != 17 {
			t.Fatalf("failed constructor did not capture static size: got %d, want 17", got)
		}
		return
	}
	if scenario == "failed-factory" {
		err := invokeDistributedServerOperation(func() error { NewBlockSelectorFromProperties(nil); return nil })
		if failure, ok := err.(*TLCError); !ok || !failure.Runtime || failure.Error() != "TLC found a null TLCServer" {
			t.Fatalf("missing server factory failure: %T/%v", err, err)
		}
	} else if got := NewBlockSelectorFromProperties(server).(*BlockSelector); got.Mode != want || (want == BlockSelectorStatic && (got.StaticBlockSize != 17 || got.GetAverageBlockCnt() != 17)) {
		t.Fatalf("initial selector = mode %v, size %d, average %d", got.Mode, got.StaticBlockSize, got.GetAverageBlockCnt())
	}
	if scenario == "deferred-static" {
		set(distributedStaticBlockSizeProperty, "33")
		if got := NewStaticBlockSelector(server).StaticBlockSize; got != 33 {
			t.Fatalf("factory initialized unused static size: got %d, want 33", got)
		}
		set(distributedStaticBlockSizeProperty, "44")
		if got := NewStaticBlockSelector(server); got.StaticBlockSize != 33 || got.GetAverageBlockCnt() != 33 {
			t.Fatalf("static selector reread settings: size %d, average %d", got.StaticBlockSize, got.GetAverageBlockCnt())
		}
		return
	}
	set(distributedSelectorStaticProperty, "true")
	set(distributedSelectorUnlimitingProperty, "false")
	set(distributedSelectorLimitingProperty, "false")
	if want == BlockSelectorStatic {
		set(distributedSelectorStaticProperty, "false")
	}
	set(distributedStaticBlockSizeProperty, "33")
	got := NewBlockSelectorFromProperties(server).(*BlockSelector)
	if got.Mode != want || (want == BlockSelectorStatic && (got.StaticBlockSize != 17 || got.GetAverageBlockCnt() != 17)) {
		t.Fatalf("later selector changed captured settings: mode %v (want %v), size %d, average %d", got.Mode, want, got.StaticBlockSize, got.GetAverageBlockCnt())
	}
}
