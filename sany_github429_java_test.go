package tlago

import (
	"path/filepath"
	"testing"
)

// Github429Test.java initializes SANY, parses, then generates semantics with
// level checking disabled. Only exceptions fail its original contract.
func TestGithub429Test_testForFailedParse(t *testing.T) {
	file := filepath.Join("sany_tests", "test_vectors", "test-model", "Github429.tla")
	loader := newSanyLoader(LoadOptions{})
	loader.initialContext = sanyGlobalInitialContext(true)
	spec, _, failed := runSanyFrontEndParse(file, loader, nil)
	if failed {
		t.Fatal("SANY.frontEndParse threw its checked ParseException")
	}
	runSanyFrontEndSemanticsWithLevels(file, spec, nil, false, nil)
}
