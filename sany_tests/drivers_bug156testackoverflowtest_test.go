package sany_tests

import "testing"

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/drivers/Bug156TEStackOverflowTest.java.
// Each test starts skipped until its Java assertions are ported and made green.
func TestBug156TEStackOverflowTest_testFrontEndParse(t *testing.T) {
	// The Java test body is intentionally disabled:
	//
	//   // uncomment if bug 156 has been fixed
	//   // SANY.frontEndParse(moduleSpec, ToolIO.out)
	//
	// Keep this as a no-op until the Java source is changed or the porting
	// pass decides to turn the historical regression into an active Go check.
}
