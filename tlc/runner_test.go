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
