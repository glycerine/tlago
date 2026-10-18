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

func TestDistributedWorkerCacheRatioLocaleProcess(t *testing.T) {
	if expected := os.Getenv("TLAGO_WORKER_RATIO_EXPECTED"); expected != "" {
		if os.Getenv("TLAGO_WORKER_RATIO_MP_FIRST") == "true" {
			MessageNumberFormat(1234)
		}
		if got := distributedWorkerCacheRatio(1234.125); got != expected {
			t.Fatalf("process worker ratio = %q, want %q", got, expected)
		}
		if got, want := MessageNumberFormat(1234), os.Getenv("TLAGO_WORKER_RATIO_MP_EXPECTED"); got != want {
			t.Fatalf("message grouping = %q, want %q", got, want)
		}
		return
	}
	for _, test := range []struct{ language, country, variant, extensions, ratio, integer string }{
		{"de", "DE", "", "", "1.234,13", "1.234"},
		{"fr", "FR", "", "", "1\u202f234,13", "1\u202f234"},
		{"ar", "EG", "", "", "١٬٢٣٤٫١٣", "١٬٢٣٤"},
		{"en", "US", "POSIX", "", "1234.13", "1,234"},
		{"wae", "CH", "", "", "1’234,13", "1’234"},
		{"th", "TH", "", "u-nu-thai", "๑,๒๓๔.๑๓", "๑,๒๓๔"},
	} {
		for _, first := range []string{"false", "true"} {
			command := exec.Command(os.Args[0], "-test.run=^TestDistributedWorkerCacheRatioLocaleProcess$")
			command.Env = append(os.Environ(),
				"TLAGO_WORKER_RATIO_EXPECTED="+test.ratio, "TLAGO_WORKER_RATIO_MP_EXPECTED="+test.integer,
				"TLAGO_WORKER_RATIO_MP_FIRST="+first,
				"user.language.format="+test.language, "user.country.format="+test.country,
				"user.variant.format="+test.variant, "user.extensions.format="+test.extensions,
				"user.script.format=")
			if output, err := command.CombinedOutput(); err != nil {
				t.Fatalf("locale %s-%s MP first %s: %v\n%s", test.language, test.country, first, err, output)
			}
		}
	}
}

func TestDistributedWorkerCacheRatioLocaleReference(t *testing.T) {
	input, err := os.Open("test_vectors/distributed/worker_cache_ratio_locales.tsv")
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	scanner := bufio.NewScanner(input)
	rows := 0
	for scanner.Scan() {
		row := strings.Split(scanner.Text(), "\t")
		if len(row) != 6 {
			t.Fatal("invalid locale reference row")
		}
		tag, extensions, _ := strings.Cut(row[0], "-u-")
		if extensions != "" {
			extensions = "u-" + extensions
		}
		symbols, key := selectMessageNumberSymbols(strings.Split(tag, "-"), extensions)
		decimal := distributedCacheDecimalSeparator(symbols, key)
		for i, char := range []rune{symbols.zero, []rune(symbols.group)[0], []rune(decimal)[0]} {
			if strconv.FormatInt(int64(char), 10) != row[i+1] {
				t.Errorf("locale %s symbol %d = %d, want %s", row[0], i, char, row[i+1])
			}
		}
		for i, value := range []float64{1234.125, math.Copysign(0, -1)} {
			if got := distributedWorkerCacheRatioForLocale(value, symbols, key); got != row[i+4] {
				t.Errorf("locale %s ratio %v = %q, want %q", row[0], value, got, row[i+4])
			}
		}
		for _, test := range []struct {
			value float64
			want  string
		}{{math.NaN(), "NaN"}, {math.Inf(1), "Infinity"}, {-1, "n/a"}} {
			if got := distributedWorkerCacheRatioForLocale(test.value, symbols, key); got != test.want {
				t.Errorf("locale %s special ratio = %q, want %q", row[0], got, test.want)
			}
		}
		rows++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if rows != 1860 {
		t.Fatalf("locale reference rows = %d, want 1860", rows)
	}
}
