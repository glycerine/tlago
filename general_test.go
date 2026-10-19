package tlago

import (
	"testing"
)

func skipSlowTest(t *testing.T) {
	if testing.Short() {
		return // ironically we get to run the long tests with -short
	}
	t.Skip("test is too slow")
}
