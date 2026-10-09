package tlc

import (
	"math"
	"strconv"
	"strings"
	"testing"
)

func TestCalculateOptimisticProbabilityUsesJavaLongSubtractionBeforeDoubleConversion(t *testing.T) {
	distinct := int64(1<<54 + 1)
	generated := distinct + 1

	got := CalculateOptimisticProbability(uint64(distinct), generated)
	want := float64(distinct) / math.Pow(2, 64)
	if got != want {
		t.Fatalf("optimistic probability = %.17g, want %.17g", got, want)
	}
	if got == 0 {
		t.Fatalf("optimistic probability collapsed to zero; Java subtracts long counts before double conversion")
	}
}

// No original test directly covers AbstractChecker.reportSuccess's signed
// distance/MathContext(2) boundary. Expected values use the source decimal
// division, including its default HALF_UP rounding, before double conversion.
// ProbabilityToString omits the leading zero for positive fractional values.
func TestReportSuccessSignedFingerprintDistance(t *testing.T) {
	for _, row := range []struct {
		distance int64
		want     string
	}{
		{1, "1.0"}, {2, ".5"}, {3, ".33"}, {6, ".17"},
		{8, ".13"}, {16, ".063"}, {64, ".016"},
		{80, ".013"}, {800, ".0013"},
		{799999999999999999, "1.3E-18"},
		{800000000000000000, "1.3E-18"},
		{800000000000000001, "1.2E-18"},
		{-800000000000000001, "-1.2E-18"},
		{math.MaxInt64, "1.1E-19"},
		{-1, "-1.0"}, {-8, "-0.13"}, {math.MinInt64, "-1.1E-19"},
	} {
		t.Run(strconv.FormatInt(row.distance, 10), func(t *testing.T) {
			captureFailoverToolIO(t, ToolIOTool)
			ReportSuccessCountsDistance(2, uint64(row.distance), 3)
			messages := ToolIOGetAllMessages()
			want := "based on the actual fingerprints:  val = " + row.want
			if len(messages) != 1 || !strings.HasSuffix(messages[0], want) {
				t.Fatalf("signed distance %d reporting = %q, want suffix %q", row.distance, messages, want)
			}
		})
	}
}

func TestReportSuccessZeroFingerprintDistance(t *testing.T) {
	t.Run("nonempty", func(t *testing.T) {
		captureFailoverToolIO(t, ToolIOTool)
		defer func() {
			failure, ok := recover().(*ArithmeticException)
			if !ok || javaNullableString(javaThrowableDetailMessage(failure)) != "Division by zero" {
				t.Fatalf("zero distance failure = %v, want ArithmeticException: Division by zero", failure)
			}
			if messages := ToolIOGetAllMessages(); len(messages) != 0 {
				t.Fatalf("failed distance calculation published success: %q", messages)
			}
		}()
		ReportSuccessCountsDistance(2, 0, 3)
	})
	t.Run("empty", func(t *testing.T) {
		captureFailoverToolIO(t, ToolIOTool)
		ReportSuccessCountsDistance(0, 0, 0)
		messages := ToolIOGetAllMessages()
		if len(messages) != 1 || !strings.HasSuffix(messages[0], "based on the actual fingerprints:  val = 0.0") {
			t.Fatalf("empty model success = %q", messages)
		}
	})
}

func TestLocalFingerprintCheckFailureSuccessReporting(t *testing.T) {
	captureFailoverToolIO(t, ToolIOTool)
	endpoint := &distributedCheckFailureEndpoint{LocalFingerprintEndpoint: NewLocalFingerprintEndpoint(NewMemFPSet()), failure: NewIOException("local final check failure")}
	manager := NewNonDistributedFPSetManager(endpoint.Set, "local", nil)
	manager.entry(0).set = endpoint
	distance := manager.CheckFPs()
	if distance != math.MaxUint64 {
		t.Fatalf("local checked-I/O fallback = %d, want signed -1 bits", distance)
	}
	ReportSuccessCountsDistance(2, distance, 3)
	messages := ToolIOGetAllMessages()
	if len(messages) != 2 || !strings.Contains(messages[0], "local final check failure") || !strings.HasSuffix(messages[1], "based on the actual fingerprints:  val = -1.0") {
		t.Fatalf("local final-check reporting = %q", messages)
	}
}
