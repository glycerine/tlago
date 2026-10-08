package tlc

import (
	"bufio"
	"math"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"
)

// No original method directly tests the coordinator's formatting calls. The
// reference rows use MP's exact DecimalFormat pattern on OpenJDK 21.0.12.1.
// Each locale runs in a fresh native process, matching source static defaults.
func TestDistributedProgressStatisticsLocaleProcess(t *testing.T) {
	if row := os.Getenv("TLAGO_PROGRESS_NUMBER_REFERENCE"); row != "" {
		checkDistributedProgressStatistics(t, strings.Split(row, "\t"))
		return
	}
	file, err := os.Open("test_vectors/distributed/progress_numbers.tsv")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	rows := 0
	for scanner.Scan() {
		row := scanner.Text()
		fields := strings.Split(row, "\t")
		if len(fields) != 5 {
			t.Fatal("invalid progress number reference row")
		}
		t.Run(fields[0], func(t *testing.T) {
			tag, extensions, _ := strings.Cut(fields[0], "-u-")
			if extensions != "" {
				extensions = "u-" + extensions
			}
			parts := strings.Split(tag, "-")
			command := exec.Command(os.Args[0], "-test.run=^TestDistributedProgressStatisticsLocaleProcess$")
			command.Env = append(os.Environ(),
				"TLAGO_PROGRESS_NUMBER_REFERENCE="+row,
				"user.language.format="+parts[0], "user.country.format="+parts[1],
				"user.variant.format="+strings.Join(parts[2:], "-"),
				"user.script.format=", "user.extensions.format="+extensions,
				tlcServerPropertyPrefix+".report=60000")
			if output, err := command.CombinedOutput(); err != nil {
				t.Fatalf("native locale process failed: %v\n%s", err, output)
			}
		})
		rows++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if rows != 9 {
		t.Fatalf("progress locale rows = %d, want 9", rows)
	}
}

func checkDistributedProgressStatistics(t *testing.T, reference []string) {
	t.Helper()
	captureFailoverToolIO(t, ToolIOTool)
	recorder := &MemoryRecorder{}
	AddMessageRecorder(recorder)
	defer RemoveMessageRecorder(recorder)
	store := NewMemFPSet()
	queue := NewMemStateQueue()
	for fp := range uint64(1234) {
		store.Put(fp)
		queue.Enqueue(&TLCStateMut{UID: TLCStateInitUID, level: 1})
	}
	trace := NewTLCTrace()
	trace.level = 12
	server := &TLCServer{Trace: trace, StateQueue: queue,
		FPSetManager: NewNonDistributedFPSetManager(store, "local", trace)}
	for _, baseline := range []int64{0, 2468} {
		generated, distinct := baseline, uint64(baseline)
		server.PrintProgressStats(time.Time{}, &generated, &distinct)
		if generated != 1234 || distinct != 1234 || server.StatesPerMinute != 1234-baseline || server.DistinctStatesPerMinute != 1234-baseline {
			t.Fatal("formatting changed counters, rates or baseline publication")
		}
	}
	records := recorder.Records(ECTLCProgressStats)
	if len(records) != 2 {
		t.Fatalf("periodic progress records = %d, want 2", len(records))
	}
	for i, rate := range []string{reference[1], reference[2]} {
		want := []string{"12", reference[1], reference[1], reference[1], rate, rate}
		if !reflect.DeepEqual(records[i].Params, want) {
			t.Errorf("periodic progress = %q, want %q", records[i].Params, want)
		}
	}
	Globals.Lock()
	oldTool := Globals.Tool
	Globals.Unlock()
	defer func() { Globals.Lock(); Globals.Tool = oldTool; Globals.Unlock() }()
	for _, tool := range []bool{false, true} {
		for _, success := range []bool{false, true} {
			Globals.Lock()
			Globals.Tool = tool
			Globals.Unlock()
			beforeProgress := len(recorder.Records(ECTLCProgressStats))
			beforeStats := len(recorder.Records(ECTLCStats))
			beforeDepth := len(recorder.Records(ECTLCSearchDepth))
			server.PrintSummary(12, math.MinInt64, -1234, math.MaxInt64, success)
			progress := recorder.Records(ECTLCProgressStats)
			wantProgress := beforeProgress
			if tool {
				wantProgress++
				want := []string{"12", reference[3], reference[4], reference[2], "0", "0"}
				if !reflect.DeepEqual(progress[len(progress)-1].Params, want) {
					t.Errorf("final progress = %q, want %q", progress[len(progress)-1].Params, want)
				}
			}
			if len(progress) != wantProgress {
				t.Fatal("tool mode changed final progress record count")
			}
			stats := recorder.Records(ECTLCStats)
			want := []string{"-9223372036854775808", "9223372036854775807", "-1234"}
			if len(stats) != beforeStats+1 || !reflect.DeepEqual(stats[len(stats)-1].Params, want) {
				t.Fatal("summary changed the plain signed statistics parameters")
			}
			depth := recorder.Records(ECTLCSearchDepth)
			wantDepth := beforeDepth
			if success {
				wantDepth++
				if !reflect.DeepEqual(depth[len(depth)-1].Params, []string{"12"}) {
					t.Fatal("search depth was localized")
				}
			}
			if len(depth) != wantDepth {
				t.Fatal("success flag changed search depth reporting")
			}
		}
	}
}
