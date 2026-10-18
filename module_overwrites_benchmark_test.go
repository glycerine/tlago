// Java sources:
//
//	test-benchmark/tlc2/tool/ModuleOverwritesBenchmark.java
//	test-benchmark/ModuleOverwrites.java
//
// (Microsoft Research, MIT license).
package tlago

import (
	"math/rand"
	"path/filepath"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

var javaModuleOverwritesBenchmarkResult bool

func javaModuleOverwritesNativeClass() tlc.NativeClass {
	return tlc.NativeClass{
		Name: "ModuleOverwrites",
		Methods: []tlc.NativeMethod{{
			Name:           "noDupesOverwrite",
			Public:         true,
			Static:         true,
			ParameterCount: 2,
			Signature:      "public static tlc2.value.IValue ModuleOverwrites.noDupesOverwrite(tlc2.value.impl.FcnRcdValue,tlc2.value.impl.IntValue)",
			Eval: func(args []tlc.Value, _ int) (tlc.Value, error) {
				function := args[0].(*tlc.FcnRcdValue)
				exclude := args[1].(*tlc.IntValue)
				filtered := make([]tlc.Value, 0, len(function.Values))
				for _, value := range function.Values {
					if value != exclude {
						filtered = append(filtered, value)
					}
				}
				if len(filtered) < 2 {
					return tlc.BoolTrue, nil
				}
				for i := 0; i < len(filtered)-1; i++ {
					for j := i + 1; j < len(filtered); j++ {
						if filtered[i] == filtered[j] {
							return tlc.BoolFalse, nil
						}
					}
				}
				return tlc.BoolTrue, nil
			},
		}},
	}
}

func BenchmarkJavaModuleOverwrites(b *testing.B) {
	oldUserDir := tlc.GetFilenameUserDirectory()
	oldMain, oldSimulator, oldTool := tlc.Globals.MainChecker, tlc.Globals.Simulator, tlc.Globals.Tool
	oldCoverage, oldWarn := tlc.Globals.CoverageInterval, tlc.Globals.Warn
	oldContinuation, oldView, oldPoly := tlc.Globals.Continuation, tlc.UseView(), tlc.FP64IrredPoly()
	tlc.UniqueStringInitialize()
	tlc.InitializeSpecStatics()
	tlc.InitializeSpecProcessorStatics()
	tlc.InitializeToolStatics()
	tlc.InitializeIntValueStatics()
	tlc.InitializeLazyValueStatics()
	tlc.InitializeMPStatics()
	tlc.InitializeStringHelperStatics()
	tlc.InitializeSimulatorStatics()
	tlc.InitializeRLSimulationWorkerStatics()
	tlc.InitializeFcnRcdValueStatics()
	tlc.InitializeValuesStatics()
	tlc.InitializeStatePrinterStatics()
	tlc.InitializePrintTLCStateStatics()
	tlc.InitializeActionItemListStatics()
	tlc.InitializeAbstractCheckerStatics()
	tlc.InitializeModelCheckerStatics()
	tlc.InitializeWorkerStatics()
	tlc.InitializeSimulationWorkerStatics()
	tlc.InitializeFPIntSetStatics()
	tlc.InitializeHeapBasedDiskFPSetStatics()
	tlc.InitializeOffHeapDiskFPSetStatics()
	tlc.SetMainChecker(nil)
	tlc.SetSimulator(nil)
	tlc.Globals.Continuation = false
	tlc.Globals.CoverageInterval = -1
	tlc.Globals.Warn = true
	tlc.SetUseView(false)
	tlc.FP64Init()
	directory, err := filepath.Abs(filepath.Join("tlc", "test_vectors", "models", "ModuleOverwrites"))
	if err != nil {
		b.Fatal(err)
	}
	tlc.SetFilenameUserDirectory(&directory)
	restoreNativeClass := tlc.RegisterNativeClass(javaModuleOverwritesNativeClass())
	b.Cleanup(func() {
		restoreNativeClass()
		tlc.SetMainChecker(oldMain)
		tlc.SetSimulator(oldSimulator)
		tlc.Globals.Tool = oldTool
		tlc.Globals.CoverageInterval, tlc.Globals.Warn = oldCoverage, oldWarn
		tlc.Globals.Continuation = oldContinuation
		tlc.SetUseView(oldView)
		tlc.FP64InitPoly(oldPoly)
		tlc.SetFilenameUserDirectory(oldUserDir)
		tlc.SetTLCStateTool(nil)
	})

	tool, diags, err := loadTLCAppTool("ModuleOverwrites", "ModuleOverwrites", nil, tlc.RuntimeParameters{})
	if diags.HasErrors() {
		b.Fatal(diags.Error())
	}
	if err != nil {
		b.Fatal(err)
	}
	var state *tlc.TLCStateMut
	if err := tool.GetInitStates(tlc.NewStateFunctor(func(initial *tlc.TLCStateMut) (any, error) {
		if state != nil {
			b.Fatal("more than one initial state")
		}
		state = initial
		return nil, nil
	})); err != nil {
		b.Fatal(err)
	}
	if state == nil {
		b.Fatal("missing initial state")
	}
	invariants := tool.GetInvariants()
	if len(invariants) != 3 {
		b.Fatalf("invariants=%d, want 3", len(invariants))
	}
	function, ok := state.Lookup(tlc.UniqueStringOf("t")).(*tlc.FcnRcdValue)
	if !ok {
		b.Fatalf("t=%T, want *tlc.FcnRcdValue", state.Lookup(tlc.UniqueStringOf("t")))
	}

	for index, name := range []string{"aNoModuleOverwrite", "bModuleOverwrite", "cModuleOverwriteLinear"} {
		index, name := index, name
		b.Run(name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rand.Shuffle(len(function.Values), func(i, j int) {
					function.Values[i], function.Values[j] = function.Values[j], function.Values[i]
				})
				valid, err := tool.IsValidState(invariants[index], state)
				if err != nil {
					b.Fatal(err)
				}
				javaModuleOverwritesBenchmarkResult = valid
			}
		})
	}
}
