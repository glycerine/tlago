package tlc

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestTLCRunnerCleanupFlagPrecleanIsSeparateFromCheckerFinalCleanup(t *testing.T) {
	initTLCCheckerTest(t)
	tlcSetSystemProperty(modelCheckerVetoProperty, "false")
	t.Setenv("TLAGO_MODEL_CHECKER_VETO_CLEANUP", "false")
	t.Cleanup(func() { SetTLCStateTool(nil) })

	successMeta := filepath.Join(t.TempDir(), "success-states")
	result, err := NewTLC(Options{
		Tool:       runnerCleanupSuccessTool(),
		MetaDir:    successMeta,
		Cleanup:    false,
		NoDeadlock: true,
	}).Process(context.Background())
	if err != nil {
		t.Fatalf("successful Process returned error: %v", err)
	}
	if result.ErrorCode != NoError {
		t.Fatalf("successful Process code = %d, want %d", result.ErrorCode, NoError)
	}
	if _, err := os.Stat(successMeta); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("successful run left metadata with runner Cleanup=false: %v", err)
	}

	failureMeta := filepath.Join(t.TempDir(), "failure-states")
	stale := filepath.Join(failureMeta, "stale.chkpt")
	if err := os.MkdirAll(failureMeta, 0o755); err != nil {
		t.Fatalf("MkdirAll stale metadir: %v", err)
	}
	if err := os.WriteFile(stale, []byte("previous run"), 0o644); err != nil {
		t.Fatalf("WriteFile stale checkpoint: %v", err)
	}
	result, err = NewTLC(Options{
		Tool:       runnerCleanupFailingInitTool(),
		MetaDir:    failureMeta,
		Cleanup:    true,
		NoDeadlock: true,
	}).Process(context.Background())
	if err == nil {
		t.Fatalf("failing Process returned nil error")
	}
	if result == nil || result.ErrorCode == NoError {
		t.Fatalf("failing Process code = %v, want non-zero", result)
	}
	if _, err := os.Stat(stale); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("runner Cleanup=true did not pre-clean stale metadata: %v", err)
	}
}

func TestTLCRunnerAppliesFPIndexLikeJava(t *testing.T) {
	original := FP64IrredPoly()
	t.Cleanup(func() { FP64InitPoly(original) })

	NewTLC(Options{FPIndex: 3}).applyGlobals()
	if got := FP64IrredPoly(); got != FP64Polys[3] {
		t.Fatalf("FP64 polynomial = %#x, want FP64Polys[3] %#x", got, FP64Polys[3])
	}
}

func TestGenerateSpecTEInstallsBinaryTracePostConditionLikeJava(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	temp := t.TempDir()
	if err := os.Chdir(temp); err != nil {
		t.Fatalf("Chdir temp: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	outFile := filepath.Join("te-out", "CustomTE.tla")
	opts, err := ParseTLCOptions([]string{"-generateSpecTE", "-teSpecOutDir", outFile, "Spec.tla"})
	if err != nil {
		t.Fatalf("ParseTLCOptions returned error: %v", err)
	}
	if !opts.GenerateTraceSpec || !opts.GenerateTraceSpecBinary {
		t.Fatalf("GenerateTraceSpec/Binary = %v/%v, want true/true", opts.GenerateTraceSpec, opts.GenerateTraceSpecBinary)
	}
	if opts.TraceSpecOutputDir != "te-out" || opts.TraceSpecModuleName != "CustomTE" {
		t.Fatalf("trace spec output/module = %q/%q, want te-out/CustomTE", opts.TraceSpecOutputDir, opts.TraceSpecModuleName)
	}
	if len(opts.RuntimeParams.PostConditions) != 1 {
		t.Fatalf("runtime postconditions = %d, want 1 implicit binary trace postcondition", len(opts.RuntimeParams.PostConditions))
	}
	post := opts.RuntimeParams.PostConditions[0]
	if post.Module != "_TLCTrace" || post.Operator != "_TLCTraceSilent" || post.ConstantName != "_TLCTraceFile" {
		t.Fatalf("postcondition = %#v, want _TLCTrace!_TLCTraceSilent with _TLCTraceFile", post)
	}
	if want := filepath.Join("te-out", "CustomTE.bin"); post.FileName != want {
		t.Fatalf("postcondition file = %q, want %q", post.FileName, want)
	}

	opts, err = ParseTLCOptions([]string{"-generateSpecTE", "-noTEBin", "Spec.tla"})
	if err != nil {
		t.Fatalf("ParseTLCOptions -noTEBin returned error: %v", err)
	}
	if opts.GenerateTraceSpecBinary {
		t.Fatalf("GenerateTraceSpecBinary = true, want false")
	}
	if len(opts.RuntimeParams.PostConditions) != 0 {
		t.Fatalf("-noTEBin runtime postconditions = %d, want 0", len(opts.RuntimeParams.PostConditions))
	}
}

func runnerCleanupSuccessTool() *Tool {
	action := &Action{Name: "Next"}
	tool := NewTool()
	tool.InitStates = []*TLCStateMut{checkerTestState(0)}
	tool.Actions = []*Action{action}
	installTestNextStateGenerator(tool, func(tl *Tool, a *Action, state *TLCStateMut) (*StateVec, error) {
		return NewStateVec(0), nil
	})
	return tool
}

func runnerCleanupFailingInitTool() *Tool {
	tool := NewTool()
	tool.GetInitStatesFunc = func(tl *Tool, functor *StateFunctor) error {
		return errors.New("forced init failure")
	}
	return tool
}
