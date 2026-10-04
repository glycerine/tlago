package tlago

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Complete original SpecTraceExpressionWriterTest, including the source
// legacy SANY return-code boundary and every generated file/fixture.
func TestJavaSpecTraceExpressionWriter(t *testing.T) {
	preamble := "VARIABLE x, y\n" +
		"XIncr == (x' = x * 2)\n" +
		"            /\\ (x < 8)\n" +
		"            /\\ UNCHANGED y\n" +
		"YIncr == (y' = x + y)\n" +
		"            /\\ (y < 15)\n" +
		"            /\\ UNCHANGED x\n"
	initDefinition := []string{"TestInit", "TestInit == x \\in 1 .. 10 /\\ y \\in 1 .. 10\n"}
	nextDefinition := []string{"TestNext", "TestNext == YIncr \\/ XIncr\n"}
	generateStates := func() []*tlc.MCState {
		return []*tlc.MCState{
			tlc.ParseMCState("1: <Initial predicate>\n/\\ x = 8\n/\\ y = 7\n"),
			tlc.ParseMCState("2: <YIncr line 8, col 10 to line 10, col 26 of module Bla>\n/\\ x = 8\n/\\ y = 15\n"),
		}
	}
	run := func(name string, body func(*testing.T, *tlc.SpecTraceExpressionWriter)) {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			tlaFile, err := os.CreateTemp(dir, "sptewt_*.tla")
			if err != nil {
				t.Fatal(err)
			}
			tlaPath := tlaFile.Name()
			if err := tlaFile.Close(); err != nil {
				t.Fatal(err)
			}
			cfgFile, err := os.CreateTemp(dir, "sptewt_*.cfg")
			if err != nil {
				t.Fatal(err)
			}
			cfgPath := cfgFile.Name()
			if err := cfgFile.Close(); err != nil {
				t.Fatal(err)
			}
			specName := strings.TrimSuffix(filepath.Base(tlaPath), ".tla")
			writer := tlc.NewSpecTraceExpressionWriter()
			writer.AddPrimer(specName, "Naturals")
			writer.AppendContentToBuffers(&preamble, nil)
			body(t, writer)
			if err := writer.WriteFiles(tlaPath, cfgPath); err != nil {
				t.Fatal(err)
			}
			// Original concludeTest calls the non-strict legacy front end,
			// not a diagnostic-count assertion. Semantic analysis still runs.
			_, result := SanyFrontEndMain(tlaPath, LoadOptions{})
			if result != 0 {
				t.Fatalf("Parsing returned a non-zero success code (%d)", result)
			}

		})
	}
	run("testInitNextWithNoError", func(t *testing.T, writer *tlc.SpecTraceExpressionWriter) {
		writer.AddInitNextDefinitions(initDefinition, nextDefinition, "writerTestInit", "writerTextNext")
	})
	run("testInitNextWithError", func(t *testing.T, writer *tlc.SpecTraceExpressionWriter) {
		trace := generateStates()
		var cfg strings.Builder
		buffers := tlc.SpecTraceExpressionBuildInitNextBuffers(&cfg, trace, nil, "STEWInit", "STEWNext", "STEWAC", nextDefinition[0], true)
		first, config := buffers[0].String(), cfg.String()
		writer.AppendContentToBuffers(&first, &config)
		writer.AddTraceFunction(trace)
		second := buffers[1].String()
		writer.AppendContentToBuffers(&second, nil)
	})
	run("testInitNextWithErrorAndTraceExpression", func(t *testing.T, writer *tlc.SpecTraceExpressionWriter) {
		trace := generateStates()
		writer.AddTraceFunction(trace)
		expressions := []*tlc.Formula{tlc.NewFormula("ENABLED XIncr"), tlc.NewFormula("y # 7")}
		data := writer.CreateAndAddVariablesAndDefinitions(expressions, "writerTestTraceExpressions")
		writer.AddInitNext(trace, data, "STEWInit", "STEWNext", "STEWAC", nextDefinition[0])
	})
	run("testMultilineTraceExpression", func(t *testing.T, writer *tlc.SpecTraceExpressionWriter) {
		trace := generateStates()
		writer.AddTraceFunction(trace)
		expressions := []*tlc.Formula{tlc.NewFormula("\n/\\ y # 7\n/\\ \\/ TRUE\n \\* Comment   \\/ FALSE")}
		e := tlc.NewFormula("namedExpression == \n  (* A commend \n over two lines*)  /\\ \\/ TRUE\n \\* Comment     \\/ FALSE")
		if !e.IsNamed() {
			t.Fatal("assertTrue(e.isNamed())")
		}
		expressions = append(expressions, e)
		data := writer.CreateAndAddVariablesAndDefinitions(expressions, "writerTestTraceExpressions")
		writer.AddInitNext(trace, data, "STEWInit", "STEWNext", "STEWAC", nextDefinition[0])
	})
}
