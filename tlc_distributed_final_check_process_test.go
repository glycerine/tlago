package tlago

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/glycerine/tlago/tlc"
)

const nativeFinalFingerprintCheckFailure = "native final fingerprint check I/O failure"

// This is a native transport fault case, not an adaptation of an original
// passing model assertion: its one GENERAL diagnostic is required by Java.
func TestNativeDistributedFinalFingerprintCheckReplyLoss(t *testing.T) {
	output := runNativeDistributedModelWithCheckFailure(t, "EWD840", true, false, "reply-loss")
	stats := nativeDistributedMessages(output, tlc.ECTLCStats)
	if len(stats) != 1 || !regexp.MustCompile(`^\d+ states generated, 114942 distinct states found, 0 states left on queue\.$`).MatchString(stats[0]) {
		t.Fatalf("completed-model counts changed after final-check loss: %q", stats)
	}
	success := nativeDistributedMessages(output, tlc.ECTLCSuccess)
	if len(success) != 1 || !strings.Contains(success[0], "based on the actual fingerprints:  val = 1.1E-19") || !strings.Contains(success[0], "calculated (optimistic):  val = -") {
		t.Fatalf("lost check/statistics did not preserve source fallback reporting: %q", success)
	}
	if strings.Count(output, "Warning: Failed to connect from ") != 1 || strings.Count(output, "Warning: there is no fp server available.") != 1 {
		t.Fatal("final states-seen failure did not report exhaustion of the sole fingerprint host")
	}
	previous := strings.Index(output, fmt.Sprintf("@!@!@STARTMSG %d:", tlc.ECGeneral))
	for _, marker := range []string{"Warning: Failed to connect from ", "Warning: there is no fp server available.", fmt.Sprintf("@!@!@STARTMSG %d:", tlc.ECTLCSuccess), fmt.Sprintf("@!@!@STARTMSG %d:", tlc.ECTLCStats), fmt.Sprintf("@!@!@STARTMSG %d:", tlc.ECTLCFinished)} {
		index := strings.Index(output, marker)
		if index <= previous {
			t.Fatalf("final failure/reporting order changed at %q", marker)
		}
		previous = index
	}
	if len(nativeDistributedMessages(output, tlc.ECTLCFinished)) != 1 {
		t.Fatal("final reply loss prevented normal model completion")
	}
}

func waitForNativeFinalFingerprintCheck(t *testing.T, ctx context.Context, server, fingerprint *nativeDistributedTestProcess) {
	t.Helper()
	waitForNativeFinalFingerprintCheckHeld(t, ctx, server, fingerprint, false)
}

func waitForNativeFinalFingerprintCheckPartition(t *testing.T, ctx context.Context, server, fingerprint *nativeDistributedTestProcess) {
	t.Helper()
	waitForNativeFinalFingerprintCheckHeld(t, ctx, server, fingerprint, true)
}

func waitForNativeFinalFingerprintCheckHeld(t *testing.T, ctx context.Context, server, fingerprint *nativeDistributedTestProcess, partitioned bool) {
	t.Helper()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	marker := "NATIVE_FINAL_FP_CHECK_COMPLETED_COUNT=114942 DISTANCE="
	if partitioned {
		marker = "NATIVE_FINAL_FP_CHECK_COMPLETED_COUNT="
	}
	for !strings.Contains(fingerprint.output.String(), marker) {
		select {
		case err := <-server.done:
			server.joined = true
			t.Fatalf("coordinator exited before accepted final check: %v", err)
		case err := <-fingerprint.done:
			fingerprint.joined = true
			t.Fatalf("fingerprint host exited before accepted final check: %v", err)
		case <-ctx.Done():
			t.Fatal("accepted final check watchdog expired")
		case <-ticker.C:
		}
	}
	output := server.output.String()
	if len(nativeDistributedMessages(output, tlc.ECTLCDistributedWorkerStats)) != 1 || len(nativeDistributedMessages(output, tlc.ECTLCSuccess)) != 0 || len(nativeDistributedMessages(output, tlc.ECTLCStats)) != 0 || len(nativeDistributedMessages(output, tlc.ECTLCFinished)) != 0 || len(nativeDistributedMessages(output, tlc.ECGeneral)) != 0 || strings.Contains(output, "Warning: Failed to connect from ") {
		t.Fatal("final check was not held between orderly worker shutdown and final reporting")
	}
	if err := fingerprint.command.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	err := <-fingerprint.done
	fingerprint.joined = true
	if err == nil {
		t.Fatal("killed fingerprint host exited successfully")
	}
	lost := fingerprint.output.String()
	if len(nativeDistributedMessages(lost, tlc.ECGeneral)) != 0 || strings.Contains(lost, "unexpected EOF") {
		t.Fatal("fingerprint host reported an earlier failure before its deliberate death")
	}
	pattern := `NATIVE_FINAL_FP_CHECK_COMPLETED_COUNT=114942 DISTANCE=[1-9][0-9]*\n`
	if partitioned {
		pattern = `NATIVE_FINAL_FP_CHECK_COMPLETED_COUNT=([1-9][0-9]*) DISTANCE=[1-9][0-9]*\n`
		match := regexp.MustCompile(pattern).FindStringSubmatch(lost)
		if len(match) != 2 {
			t.Fatal("lost host did not finish its real partition check")
		}
		count, err := strconv.ParseUint(match[1], 10, 64)
		if err != nil || count >= 114942 {
			t.Fatalf("lost host was not partitioned: %q", match)
		}
	}
	if strings.Count(lost, marker) != 1 || !regexp.MustCompile(pattern).MatchString(lost) || strings.Contains(lost, "NATIVE_FINAL_FP_STATES_SEEN=") || strings.Contains(lost, "NATIVE_FINAL_FP_EXIT_CLEANUP=") {
		t.Fatal("accepted final check was repeated or followed by storage calls before reply loss")
	}
	t.Log("killed fingerprint host after its real final distance check and before the reply")
}

// No upstream method exercises CheckFPsCallable's IOException branch at final
// model reporting. These native process checks retain the complete N=7 model;
// they do not add original-method completion credit.
func TestNativeDistributedFinalFingerprintCheckIO(t *testing.T) {
	for _, failure := range []string{"returned", "panicked"} {
		t.Run(failure, func(t *testing.T) {
			output := runNativeDistributedModelWithCheckFailure(t, "EWD840", true, false, failure)
			stats := nativeDistributedMessages(output, tlc.ECTLCStats)
			if len(stats) != 1 || !regexp.MustCompile(`^\d+ states generated, 114942 distinct states found, 0 states left on queue\.$`).MatchString(stats[0]) {
				t.Fatalf("full model statistics = %q", stats)
			}
			// CheckFPsCallable returns Long.MAX_VALUE on checked I/O failure.
			// AbstractChecker.reportSuccess rounds its reciprocal with MathContext(2).
			success := nativeDistributedMessages(output, tlc.ECTLCSuccess)
			if len(success) != 1 || !strings.Contains(success[0], "based on the actual fingerprints:  val = 1.1E-19") {
				t.Fatalf("final check fallback not reported: %q", success)
			}
			general := strings.Index(output, fmt.Sprintf("@!@!@STARTMSG %d:", tlc.ECGeneral))
			previous := general
			for _, code := range []int{tlc.ECTLCSuccess, tlc.ECTLCStats, tlc.ECTLCFinished} {
				index := strings.Index(output, fmt.Sprintf("@!@!@STARTMSG %d:", code))
				if len(nativeDistributedMessages(output, code)) != 1 || index <= previous {
					t.Fatalf("final reporting missing, repeated or reordered: event %d", code)
				}
				previous = index
			}
			if strings.Contains(output, "Warning: Failed to connect from ") {
				t.Fatal("final check I/O failure reassigned a live fingerprint host")
			}
		})
	}
}

type nativeFinalFingerprintCheckEndpoint struct {
	tlc.DistributedFingerprintEndpoint
	failure string
	checked atomic.Bool
	seen    atomic.Bool
}

func (e *nativeFinalFingerprintCheckEndpoint) CheckFPs() (uint64, error) {
	count, err := e.DistributedFingerprintEndpoint.Size()
	if err != nil {
		return 0, err
	}
	if e.failure == "survivor" {
		distance, err := e.DistributedFingerprintEndpoint.CheckFPs()
		if err == nil {
			if err := e.reportDiskCheck(count); err != nil {
				return 0, err
			}
			e.checked.Store(true)
			fmt.Printf("NATIVE_FINAL_FP_SURVIVOR_CHECK_COUNT=%d DISTANCE=%d\n", count, distance)
		}
		return distance, err
	}
	if e.failure == "reply-loss" {
		distance, err := e.DistributedFingerprintEndpoint.CheckFPs()
		if err != nil {
			return 0, err
		}
		if err := e.reportDiskCheck(count); err != nil {
			return 0, err
		}
		fmt.Printf("NATIVE_FINAL_FP_CHECK_COMPLETED_COUNT=%d DISTANCE=%d\n", count, distance)
		<-make(chan struct{}) // Parent kills the host before the accepted reply.
	}
	fmt.Printf("NATIVE_FINAL_FP_CHECK_COUNT=%d\n", count)
	e.checked.Store(true)
	failure := tlc.NewIOException(nativeFinalFingerprintCheckFailure)
	if e.failure == "panicked" {
		panic(failure)
	}
	return 0, failure
}

func (e *nativeFinalFingerprintCheckEndpoint) GetStatesSeen() (uint64, error) {
	count, err := e.DistributedFingerprintEndpoint.GetStatesSeen()
	if e.checked.Load() {
		alreadySeen := e.seen.Swap(true)
		if e.failure == "survivor" {
			fmt.Printf("NATIVE_FINAL_FP_SURVIVOR_STATES_SEEN=%d\n", count)
		} else if !alreadySeen {
			fmt.Printf("NATIVE_FINAL_FP_STATES_SEEN=%d\n", count)
		}
	}
	return count, err
}

func (e *nativeFinalFingerprintCheckEndpoint) Exit(cleanup bool) error {
	if !e.checked.Load() || !e.seen.Load() {
		return fmt.Errorf("fingerprint exit preceded final check/statistics")
	}
	count, err := e.DistributedFingerprintEndpoint.Size()
	if err != nil {
		return err
	}
	if e.failure == "survivor" {
		fmt.Printf("NATIVE_FINAL_FP_SURVIVOR_EXIT_CLEANUP=%t COUNT=%d\n", cleanup, count)
	} else {
		fmt.Printf("NATIVE_FINAL_FP_EXIT_CLEANUP=%t COUNT=%d\n", cleanup, count)
	}
	return e.DistributedFingerprintEndpoint.Exit(cleanup)
}

func nativeDistributedFinalFingerprintCheckHost(args []string, failure string) (result error) {
	args, err := tlc.ExtractDistributedStartupProperties(args)
	if err != nil {
		return err
	}
	network, err := tlc.NewDistributedFPServerNetwork("127.0.0.1:0", "127.0.0.1")
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, network.Close()) }()
	env := network.FPEnvironment(tlc.DistributedFPServerEnvironment{ToolOut: os.Stdout, SystemOut: os.Stdout, SystemErr: os.Stderr})
	const name = "final-check-failure-fingerprints"
	env.RegisterFPSet = func(server tlc.DistributedServerEndpoint, endpoint tlc.DistributedFingerprintEndpoint, hostname string) error {
		coordinator, ok := server.(*tlc.NetworkServerEndpoint)
		if !ok {
			return fmt.Errorf("final check fixture requires a native TCP coordinator")
		}
		fault := &nativeFinalFingerprintCheckEndpoint{DistributedFingerprintEndpoint: endpoint, failure: failure}
		if err := network.Host.RegisterFingerprint(name, fault); err != nil {
			return err
		}
		return coordinator.RegisterFPSetReference(tlc.DistributedEndpointReference{Address: network.Address, Object: name}, hostname)
	}
	env.UnpublishFPSet = func(tlc.FPSet, bool) { network.Host.UnregisterFingerprint(name) }
	tlc.RunDistributedFPServer(args, env)
	return nil
}

// The first host is lost after its real final check; the second completes the
// same full model's partition check and remains available for final statistics.
func TestNativeDistributedFinalFingerprintCheckReplyLossSurvivor(t *testing.T) {
	checkNativeDistributedFinalFingerprintSurvivor(t, false)
}

func TestNativeDistributedFinalFingerprintCheckDiskReplyLossSurvivor(t *testing.T) {
	checkNativeDistributedFinalFingerprintSurvivor(t, true)
}

func checkNativeDistributedFinalFingerprintSurvivor(t *testing.T, disk bool) {
	t.Helper()
	mode := "reply-loss-survivor"
	if disk {
		mode += "-lsb"
	}
	output := runNativeDistributedModelWithCheckFailure(t, "EWD840", true, false, mode)
	stats := nativeDistributedMessages(output, tlc.ECTLCStats)
	if len(stats) != 1 || !regexp.MustCompile(`^\d+ states generated, 114942 distinct states found, 0 states left on queue\.$`).MatchString(stats[0]) {
		t.Fatalf("captured completed-model counts changed: %q", stats)
	}
	check := regexp.MustCompile(`NATIVE_FINAL_FP_SURVIVOR_CHECK_COUNT=([1-9][0-9]*) DISTANCE=([1-9][0-9]*)\n`).FindStringSubmatch(output)
	if len(check) != 3 || strings.Count(output, "NATIVE_FINAL_FP_SURVIVOR_CHECK_COUNT=") != 1 {
		t.Fatalf("surviving check missing or replayed: %q", check)
	}
	count, err := strconv.ParseUint(check[1], 10, 64)
	if err != nil || count >= 114942 {
		t.Fatalf("surviving host was not partitioned: %q", check)
	}
	lost := regexp.MustCompile(`NATIVE_FINAL_FP_CHECK_COMPLETED_COUNT=([1-9][0-9]*) DISTANCE=[1-9][0-9]*\n`).FindStringSubmatch(output)
	if len(lost) != 2 {
		t.Fatal("completed lost-host partition was not retained")
	}
	lostCount, err := strconv.ParseUint(lost[1], 10, 64)
	if err != nil || lostCount+count != 114942 {
		t.Fatalf("real final partitions do not cover the captured distinct count: %q/%q", lost, check)
	}
	if disk {
		for _, size := range []uint64{lostCount, count} {
			marker := fmt.Sprintf("NATIVE_FINAL_FP_DISK_CHECK_CHILDREN=2 COUNT=%d\n", size)
			if strings.Count(output, marker) != 1 {
				t.Fatalf("actual nested disk flush missing or repeated: %q", marker)
			}
		}
		if strings.Count(output, "...with nested instance type: tlc2.tool.fp.LSBDiskFPSet") != 4 {
			t.Fatal("final check did not use two physical LSB children on each host")
		}
	}
	distance, err := strconv.ParseUint(check[2], 10, 64)
	if err != nil || distance >= math.MaxInt64 {
		t.Fatalf("healthy distance replaced with failed-call fallback: %q", check)
	}
	success := nativeDistributedMessages(output, tlc.ECTLCSuccess)
	if len(success) != 1 || strings.Contains(success[0], "based on the actual fingerprints:  val = 1.1E-19") || strings.Contains(success[0], "calculated (optimistic):  val = -") {
		t.Fatalf("surviving host did not supply final collision reporting: %q", success)
	}
	seenMatch := regexp.MustCompile(`NATIVE_FINAL_FP_SURVIVOR_STATES_SEEN=([0-9]+)\n`).FindStringSubmatch(output)
	if len(seenMatch) != 2 {
		t.Fatal("surviving host's real final states-seen count missing")
	}
	seen, err := strconv.ParseInt(seenMatch[1], 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	// Check the observed probabilities independently of the production TLC
	// reporter. This fixture's reciprocal is away from a half-rounding tie;
	// source decimal HALF_UP edge cases have their own focused checks.
	probabilities := regexp.MustCompile(`val = ([^\s]+)`).FindAllStringSubmatch(success[0], -1)
	if len(probabilities) != 2 {
		t.Fatalf("final probabilities missing: %q", success)
	}
	want := []float64{114942 * (float64(seen+1-114942) / math.Exp2(64)), 1 / float64(distance)}
	for i, probability := range probabilities {
		actual, err := strconv.ParseFloat(probability[1], 64)
		expected, parseErr := strconv.ParseFloat(fmt.Sprintf("%.1E", want[i]), 64)
		if err != nil || parseErr != nil || actual != expected {
			t.Fatalf("final probability %d = %q, want %g from actual surviving storage", i, probability[1], expected)
		}
	}
	if strings.Count(output, "Warning: Failed to connect from ") != 1 || strings.Contains(output, "Warning: there is no fp server available.") || strings.Count(output, "NATIVE_FINAL_FP_SURVIVOR_STATES_SEEN=") != 1 {
		t.Fatal("final statistics did not reassign the failed first slot and visit the surviving second slot exactly once")
	}
	if strings.Count(output, fmt.Sprintf("NATIVE_FINAL_FP_SURVIVOR_EXIT_CLEANUP=true COUNT=%d\n", count)) != 1 || len(nativeDistributedMessages(output, tlc.ECTLCFinished)) != 1 {
		t.Fatal("surviving host did not complete ordinary storage cleanup")
	}
	previous := strings.Index(output, fmt.Sprintf("@!@!@STARTMSG %d:", tlc.ECGeneral))
	for _, marker := range []string{"Warning: Failed to connect from ", fmt.Sprintf("@!@!@STARTMSG %d:", tlc.ECTLCSuccess), fmt.Sprintf("@!@!@STARTMSG %d:", tlc.ECTLCStats), fmt.Sprintf("@!@!@STARTMSG %d:", tlc.ECTLCFinished)} {
		index := strings.Index(output, marker)
		if index <= previous {
			t.Fatalf("failure/reporting order changed at %q", marker)
		}
		previous = index
	}
}

func waitForNativeDistributedMarker(t *testing.T, ctx context.Context, process *nativeDistributedTestProcess, description string, ready func() bool) {
	t.Helper()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for !ready() {
		select {
		case err := <-process.done:
			process.joined = true
			t.Fatalf("%s exited before %s: %v", process.output.role, description, err)
		case <-ctx.Done():
			t.Fatalf("watchdog waiting for %s", description)
		case <-ticker.C:
		}
	}
}

// A Mem host has nothing to report here. Disk cases must show actual complete
// flushes in both physical children after storage's own final check returns.
func (e *nativeFinalFingerprintCheckEndpoint) reportDiskCheck(count uint64) error {
	local, ok := e.DistributedFingerprintEndpoint.(*tlc.LocalFingerprintEndpoint)
	if !ok {
		return fmt.Errorf("final check requires owned local storage")
	}
	multi, ok := local.Set.(*tlc.MultiFPSet)
	if !ok {
		return nil
	}
	if len(multi.Sets) != 2 {
		return fmt.Errorf("final disk check requires two children")
	}
	var total uint64
	for _, child := range multi.Sets {
		disk, ok := child.(*tlc.LSBDiskFPSet)
		if !ok {
			return fmt.Errorf("unexpected final-check child %T", child)
		}
		if disk.GetFileCnt() <= 0 || uint64(disk.GetFileCnt()) != disk.Size() || disk.GetTblCnt() != 0 {
			return fmt.Errorf("disk check did not flush complete child membership: file=%d table=%d size=%d", disk.GetFileCnt(), disk.GetTblCnt(), disk.Size())
		}
		total += uint64(disk.GetFileCnt())
	}
	if total != count {
		return fmt.Errorf("physical final disk membership %d != host count %d", total, count)
	}
	fmt.Printf("NATIVE_FINAL_FP_DISK_CHECK_CHILDREN=2 COUNT=%d\n", total)
	return nil
}
