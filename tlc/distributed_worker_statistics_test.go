package tlc

import (
	"bufio"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
)

// No original method directly checks the worker summary. Reference rows come
// from its exact ternary and String.format(Locale.ROOT, "%1$,.2f", ratio),
// using OpenJDK 21.0.12.1, edge values and 256 deterministic positive bit patterns.
func TestDistributedWorkerCacheRatioReference(t *testing.T) {
	input, err := os.Open("test_vectors/distributed/worker_cache_ratios.tsv")
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	rows := 0
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		row := strings.SplitN(scanner.Text(), "\t", 2)
		if len(row) != 2 {
			t.Fatal("invalid cache-ratio reference row")
		}
		bits, err := strconv.ParseUint(row[0], 16, 64)
		if err != nil {
			t.Fatal(err)
		}
		if got := distributedWorkerCacheRatio(math.Float64frombits(bits)); got != row[1] {
			t.Errorf("ratio bits %s = %q, want %q", row[0], got, row[1])
		}
		rows++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if rows != 271 {
		t.Fatalf("reference rows = %d, want 271", rows)
	}
}
