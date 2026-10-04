package tlc

import (
	"math"
	"strconv"
	"testing"
)

// Complete original tlc2.TLCTest; runtime memory assertions use the same native
// process-budget boundary as production, just as Runtime.maxMemory does in Java.
func TestJavaTLC(t *testing.T) {
	t.Chdir(t.TempDir())
	Globals.Lock()
	oldBound, oldMeta, oldWarn, oldTool := Globals.SetBound, Globals.MetaDir, Globals.Warn, Globals.Tool
	Globals.Warn, Globals.Tool = true, false
	Globals.Unlock()
	t.Cleanup(func() {
		Globals.Lock()
		Globals.SetBound, Globals.MetaDir, Globals.Warn, Globals.Tool = oldBound, oldMeta, oldWarn, oldTool
		Globals.Unlock()
	})
	t.Run("testHandleParametersAbsoluteInvalid", func(t *testing.T) {
		checker := NewTLC(Options{})
		javaMCEquals(t, false, checker.HandleParameters([]string{"-fpmem", "-1", "MC"}) == nil)
	})
	t.Run("testHandleParametersAbsoluteValid", func(t *testing.T) {
		checker := NewTLC(Options{})
		javaMCEquals(t, true, checker.HandleParameters([]string{"-fpmem", "101", "MC"}) == nil)
	})
	t.Run("testHandleParametersFractionInvalid", func(t *testing.T) {
		checker := NewTLC(Options{})
		javaMCEquals(t, false, checker.HandleParameters([]string{"-fpmem", "-0.5", "MC"}) == nil)
	})
	t.Run("testHandleParametersAllocateLowerBound", func(t *testing.T) {
		checker := NewTLC(Options{})
		javaMCEquals(t, true, checker.HandleParameters([]string{"-fpmem", "0", "MC"}) == nil)
		cfg := checker.FPSetConfiguration
		if !FPSetAllocatesOnHeap(cfg.GetImplementation()) {
			t.Skip("original assumeTrue: allocatesOnHeap")
		}
		javaMCEquals(t, tlcRuntimeMinFPMemSize, cfg.GetMemoryInBytes())
	})
	t.Run("testHandleParametersAllocateUpperBound", func(t *testing.T) {
		checker := NewTLC(Options{})
		javaMCEquals(t, true, checker.HandleParameters([]string{"-fpmem", strconv.FormatInt(math.MaxInt64, 10), "MC"}) == nil)
		maxMemory := javaDoubleToLong(float64(tlcRuntimeMaxHeapMemoryBytes()) * 0.75)
		cfg := checker.FPSetConfiguration
		if !FPSetAllocatesOnHeap(cfg.GetImplementation()) {
			t.Skip("original assumeTrue: allocatesOnHeap")
		}
		javaMCEquals(t, maxMemory, cfg.GetMemoryInBytes())
	})
	t.Run("testHandleParametersAllocateHalf", func(t *testing.T) {
		checker := NewTLC(Options{})
		javaMCEquals(t, true, checker.HandleParameters([]string{"-fpmem", ".5", "MC"}) == nil)
		maxMemory := javaDoubleToLong(float64(tlcRuntimeMaxHeapMemoryBytes()) * 0.50)
		cfg := checker.FPSetConfiguration
		if !FPSetAllocatesOnHeap(cfg.GetImplementation()) {
			t.Skip("original assumeTrue: allocatesOnHeap")
		}
		javaMCEquals(t, maxMemory, cfg.GetMemoryInBytes())
	})
	t.Run("testHandleParametersAllocate90", func(t *testing.T) {
		checker := NewTLC(Options{})
		javaMCEquals(t, true, checker.HandleParameters([]string{"-fpmem", ".99", "MC"}) == nil)
		maxMemory := javaDoubleToLong(float64(tlcRuntimeMaxHeapMemoryBytes()) * 0.99)
		cfg := checker.FPSetConfiguration
		if !FPSetAllocatesOnHeap(cfg.GetImplementation()) {
			t.Skip("original assumeTrue: allocatesOnHeap")
		}
		javaMCEquals(t, maxMemory, cfg.GetMemoryInBytes())
	})
	t.Run("testHandleParametersMaxSetSize", func(t *testing.T) {
		progDefault := Globals.SetBound
		checker := NewTLC(Options{})
		javaMCEquals(t, false, checker.HandleParameters([]string{"-maxSetSize", "NaN", "MC"}) == nil)
		checker = NewTLC(Options{})
		javaMCEquals(t, false, checker.HandleParameters([]string{"-maxSetSize", "0", "MC"}) == nil)
		checker = NewTLC(Options{})
		javaMCEquals(t, false, checker.HandleParameters([]string{"-maxSetSize", "-1", "MC"}) == nil)
		checker = NewTLC(Options{})
		javaMCEquals(t, false, checker.HandleParameters([]string{"-maxSetSize", strconv.FormatInt(math.MinInt32, 10), "MC"}) == nil)
		checker = NewTLC(Options{})
		javaMCEquals(t, true, checker.HandleParameters([]string{"-maxSetSize", "1", "MC"}) == nil)
		javaMCEquals(t, 1, Globals.SetBound)
		checker = NewTLC(Options{})
		javaMCEquals(t, true, checker.HandleParameters([]string{"-maxSetSize", strconv.Itoa(progDefault), "MC"}) == nil)
		javaMCEquals(t, progDefault, Globals.SetBound)
		checker = NewTLC(Options{})
		javaMCEquals(t, true, checker.HandleParameters([]string{"-maxSetSize", strconv.FormatInt(math.MaxInt32, 10), "MC"}) == nil)
		javaMCEquals(t, int(math.MaxInt32), Globals.SetBound)
	})
	t.Run("testHandleParametersSimulateFileNum", func(t *testing.T) {
		checker := NewTLC(Options{})
		javaMCEquals(t, true, checker.HandleParameters([]string{"MC", "-simulate", "num=5,file=test.txt"}) == nil)
		checker = NewTLC(Options{})
		javaMCEquals(t, false, checker.HandleParameters([]string{"MC", "-simulate", "file=test.txt"}) == nil)
		checker = NewTLC(Options{})
		javaMCEquals(t, true, checker.HandleParameters([]string{"MC", "-simulate", "num=10"}) == nil)
	})
	t.Run("testRuntimeConversion", func(t *testing.T) {
		javaMCEquals(t, "59s", ConvertRuntimeToHumanReadable(59000))
		javaMCEquals(t, "59min 59s", ConvertRuntimeToHumanReadable(3599000))
		javaMCEquals(t, "23h 59min", ConvertRuntimeToHumanReadable(86340000))
		javaMCEquals(t, "1d 23h", ConvertRuntimeToHumanReadable(169200000))
		javaMCEquals(t, "2d 23h", ConvertRuntimeToHumanReadable(255600000))
		javaMCEquals(t, "99d 23h", ConvertRuntimeToHumanReadable(8636400000))
	})
}
