// Java source: test-benchmark/tlc2/tool/simulation/SimulatorBenchmark.java
package tlago

import (
	"fmt"
	"math"
	"path/filepath"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

var javaSimulatorBenchmarkWorkers = [...]int{1, 2, 3, 4, 5, 6}

func BenchmarkJavaSimulatorWorkers(b *testing.B) {
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
	tlc.InitializeFcnRcdValueStatics()
	tlc.InitializeValuesStatics()
	tlc.InitializeStatePrinterStatics()
	tlc.InitializePrintTLCStateStatics()
	tlc.InitializeActionItemListStatics()
	tlc.InitializeMPStatics()
	tlc.InitializeStringHelperStatics()
	tlc.InitializeSimulatorStatics()
	tlc.InitializeRLSimulationWorkerStatics()
	tlc.InitializeSimulationWorkerStatics()
	tlc.InitializeFPIntSetStatics()
	tlc.SetMainChecker(nil)
	tlc.SetSimulator(nil)
	tlc.Globals.Continuation = false
	tlc.Globals.CoverageInterval = -1
	tlc.Globals.Warn = true
	tlc.SetUseView(false)
	tlc.FP64Init()
	directory, err := filepath.Abs(filepath.Join("tlc", "test_vectors", "models", "BenchmarkSpec"))
	if err != nil {
		b.Fatal(err)
	}
	tlc.SetFilenameUserDirectory(&directory)
	b.Cleanup(func() {
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

	for _, workers := range javaSimulatorBenchmarkWorkers {
		b.Run(fmt.Sprintf("nWorkers=%d", workers), func(b *testing.B) {
			rng := tlc.NewJavaRandom(0)
			seed := int64(0)
			rng.SetSeed(0) // @Setup(Level.Trial)
			seed++
			rng.SetSeed(seed) // @Setup(Level.Iteration)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tool, diags, err := loadTLCAppTool("BenchmarkSpec", "MCInv", nil, tlc.RuntimeParameters{})
				if diags.HasErrors() {
					fmt.Println(diags.Error())
					continue
				}
				if err != nil {
					fmt.Println(err.Error())
					continue
				}
				tool.SetMode(tlc.ModeSimulation)
				tlc.SetTLCStateTool(tool)
				simulator := tlc.NewSimulator(tool, false, 20, math.MaxInt64, 0,
					tlc.WithSimulatorMetaDir(""), tlc.WithSimulatorWorkerCount(workers),
					tlc.WithSimulatorRandomGenerator(rng))
				if _, err := simulator.Simulate(); err != nil {
					fmt.Println(err.Error())
				}
			}
		})
	}
}
