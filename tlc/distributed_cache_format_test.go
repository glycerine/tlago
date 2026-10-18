package tlc

import (
	"bufio"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

// No original method tests SimpleCache's DecimalFormat output. Saved reference
// rows come from that exact source formatter; Go tests need no Java runtime.
func TestDistributedCacheCompactNumberReference(t *testing.T) {
	file, err := os.Open("test_vectors/distributed/simple_cache_ratios.tsv")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	symbols, _ := selectMessageNumberSymbols([]string{"und"}, "")
	special := selectCacheDecimalSymbols([]string{"und"}, "")
	scanner := bufio.NewScanner(file)
	rows := 0
	for scanner.Scan() {
		row := strings.Split(scanner.Text(), "\t")
		if len(row) != 4 {
			t.Fatal("invalid cache ratio row")
		}
		bits, err := strconv.ParseUint(row[2], 16, 64)
		if err != nil {
			t.Fatal(err)
		}
		if got := compactCacheRatio(math.Float64frombits(bits), symbols, special); got != row[3] {
			t.Errorf("ratio %s = %q, want %q", row[2], got, row[3])
		}
		rows++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if rows != 271 {
		t.Fatal("missing cache ratio reference rows")
	}
}

func TestDistributedCacheCompactLocaleReference(t *testing.T) {
	file, err := os.Open("test_vectors/distributed/simple_cache_ratio_locales.tsv")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	rows := 0
	values := []float64{float64(1234568) / 1000, -float64(1234568) / 1000, math.Copysign(0, -1), math.Inf(1), math.Inf(-1), math.NaN()}
	for scanner.Scan() {
		row := strings.Split(scanner.Text(), "\t")
		if len(row) != 7 {
			t.Fatal("invalid cache locale row")
		}
		tag, ext, _ := strings.Cut(row[0], "-u-")
		if ext != "" {
			ext = "u-" + ext
		}
		parts := strings.Split(tag, "-")
		symbols, _ := selectMessageNumberSymbols(parts, ext)
		special := selectCacheDecimalSymbols(parts, ext)
		for i, value := range values {
			if got := compactCacheRatio(value, symbols, special); got != row[i+1] {
				t.Errorf("locale %s value %v = %q, want %q", row[0], value, got, row[i+1])
			}
		}
		rows++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if rows != 1860 {
		t.Fatalf("cache locale rows = %d, want 1860", rows)
	}
}

func TestDistributedCacheCompactLargeNumberReference(t *testing.T) {
	file, err := os.Open("test_vectors/distributed/simple_cache_large_numbers.tsv")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	symbols, _ := selectMessageNumberSymbols([]string{"und"}, "")
	special := selectCacheDecimalSymbols([]string{"und"}, "")
	scanner := bufio.NewScanner(file)
	rows := 0
	for scanner.Scan() {
		row := strings.Split(scanner.Text(), "\t")
		if len(row) != 2 {
			t.Fatal("invalid large-number reference row")
		}
		bits, err := strconv.ParseUint(row[0], 16, 64)
		if err != nil {
			t.Fatal(err)
		}
		if got := compactCacheRatio(math.Float64frombits(bits), symbols, special); got != row[1] {
			t.Errorf("large ratio %s = %q, want %q", row[0], got, row[1])
		}
		rows++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if rows != 6610 {
		t.Fatal("missing large-number reference rows")
	}
}

func TestDistributedCacheCompactLocaleProcess(t *testing.T) {
	if want := os.Getenv("TLAGO_COMPACT_CACHE_EXPECTED"); want != "" {
		cache := NewSimpleCache()
		cache.cacheHit.Store(0)
		cache.cacheMiss.Store(0)
		if got := cache.GetHitRatioAsString(); got != want {
			t.Fatalf("process NaN = %q, want %q", got, want)
		}
		return
	}
	for _, test := range []struct{ language, country, variant, extensions, expected string }{
		{"en", "US", "", "", "NaN"}, {"ar", "EG", "", "", "ليس\u00a0رقم"},
		{"ar", "EG", "", "u-nu-latn", "ليس\u00a0رقمًا"}, {"fa", "IR", "", "", "ناعدد"},
		{"ja", "JP", "JP", "", "�"}, {"th", "TH", "TH", "", "�"},
	} {
		command := exec.Command(os.Args[0], "-test.run=^TestDistributedCacheCompactLocaleProcess$")
		command.Env = append(os.Environ(), "TLAGO_COMPACT_CACHE_EXPECTED="+test.expected,
			"user.language.format="+test.language, "user.country.format="+test.country,
			"user.variant.format="+test.variant, "user.extensions.format="+test.extensions, "user.script.format=")
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("native locale process failed: %v\n%s", err, output)
		}
	}
}
