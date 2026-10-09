package tlago

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

const nativeFinalFingerprintCheckFailure = "native final fingerprint check I/O failure"

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
	panics  bool
	checked atomic.Bool
	seen    atomic.Bool
}

func (e *nativeFinalFingerprintCheckEndpoint) CheckFPs() (uint64, error) {
	count, err := e.DistributedFingerprintEndpoint.Size()
	if err != nil {
		return 0, err
	}
	fmt.Printf("NATIVE_FINAL_FP_CHECK_COUNT=%d\n", count)
	e.checked.Store(true)
	failure := tlc.NewIOException(nativeFinalFingerprintCheckFailure)
	if e.panics {
		panic(failure)
	}
	return 0, failure
}

func (e *nativeFinalFingerprintCheckEndpoint) GetStatesSeen() (uint64, error) {
	count, err := e.DistributedFingerprintEndpoint.GetStatesSeen()
	if e.checked.Load() && !e.seen.Swap(true) {
		fmt.Printf("NATIVE_FINAL_FP_STATES_SEEN=%d\n", count)
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
	fmt.Printf("NATIVE_FINAL_FP_EXIT_CLEANUP=%t COUNT=%d\n", cleanup, count)
	return e.DistributedFingerprintEndpoint.Exit(cleanup)
}

func nativeDistributedFinalFingerprintCheckHost(args []string, panics bool) (result error) {
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
		fault := &nativeFinalFingerprintCheckEndpoint{DistributedFingerprintEndpoint: endpoint, panics: panics}
		if err := network.Host.RegisterFingerprint(name, fault); err != nil {
			return err
		}
		return coordinator.RegisterFPSetReference(tlc.DistributedEndpointReference{Address: network.Address, Object: name}, hostname)
	}
	env.UnpublishFPSet = func(tlc.FPSet, bool) { network.Host.UnregisterFingerprint(name) }
	tlc.RunDistributedFPServer(args, env)
	return nil
}
