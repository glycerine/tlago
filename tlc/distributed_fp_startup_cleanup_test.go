package tlc

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// No original method tests ownership before FP publication. Run the actual
// command's disk factory in isolated processes with a native configured heap.
func TestNativeFingerprintStartupCleanup(t *testing.T) {
	if phase := os.Getenv("TLAGO_FP_STARTUP_FAILURE"); phase != "" {
		checkNativeFingerprintStartupCleanup(t, phase)
		return
	}
	for _, backend := range []string{"LSBDiskFPSet", "MSBDiskFPSet"} {
		t.Run(backend, func(t *testing.T) {
			for _, phase := range []string{"partial_init", "hostname_return", "hostname_panic", "missing_coordinator", "registration_failure", "report_failure"} {
				t.Run(phase, func(t *testing.T) {
					command := exec.Command(os.Args[0], "-test.run=^TestNativeFingerprintStartupCleanup$")
					command.Env = append(os.Environ(), "TLAGO_FP_STARTUP_FAILURE="+phase, FPSetImplProperty+"=tlc2.tool.fp."+backend, "GOMEMLIMIT=64MiB")
					if output, err := command.CombinedOutput(); err != nil {
						t.Fatalf("fingerprint startup ownership failed: %v\n%s", err, output)
					}
				})
			}
		})
	}
}

func checkNativeFingerprintStartupCleanup(t *testing.T, phase string) {
	directory := t.TempDir()
	t.Setenv("java.io.tmpdir", directory)
	metadir := filepath.Join(directory, "FPSet12345")
	if phase == "partial_init" {
		if err := os.MkdirAll(filepath.Join(metadir, "FPSet12345_1.fp"), 0700); err != nil {
			t.Fatal(err)
		}
	}
	var output bytes.Buffer
	hostCalls, registrationCalls := 0, 0
	var registered FPSet
	var failure error = errors.New("native hostname unavailable")
	if phase == "hostname_panic" {
		failure = NewAssertionError("native hostname failure")
	}
	env := DistributedFPServerEnvironment{
		ToolOut: &output, SystemOut: &output, SystemErr: &output,
		CurrentTimeMillis: func() int64 { return 12345 },
		Lookup: func(string) (DistributedServerEndpoint, error) {
			if phase == "missing_coordinator" {
				return nil, nil
			}
			return NewLocalServerEndpoint(&TLCServer{}), nil
		},
		LocalHostName: func() (string, error) {
			hostCalls++
			for child := 0; child < 2; child++ {
				info, err := os.Stat(filepath.Join(metadir, fmt.Sprintf("FPSet12345_%d.fp", child)))
				if err != nil || !info.Mode().IsRegular() {
					t.Fatalf("hostname lookup preceded disk initialization: %v/%v", info, err)
				}
			}
			if runtime.GOOS == "linux" && nativeFingerprintStartupDescriptors(t, metadir) == 0 {
				t.Fatal("hostname lookup did not observe actual initialized disk handles")
			}
			if phase == "hostname_return" {
				return "", failure
			}
			if phase == "hostname_panic" {
				panic(failure)
			}
			return "owner", nil
		},
		RegisterFPSet: func(_ DistributedServerEndpoint, endpoint DistributedFingerprintEndpoint, _ string) error {
			registrationCalls++
			registered = endpoint.(*LocalFingerprintEndpoint).Set
			if phase == "registration_failure" {
				return failure
			}
			if phase == "report_failure" {
				return nil
			}
			return errors.New("unpublished failure reached registration")
		},
		Wait: func(set FPSet, duration time.Duration) error {
			if set != registered || duration != 300000*time.Millisecond {
				t.Fatal("reporting changed the registered store or source wait bound")
			}
			return failure
		},
	}
	flush, err := runDistributedFPServer("coordinator", env)
	attempted := phase == "registration_failure" || phase == "report_failure"
	expectedRegistrations := 0
	if attempted {
		expectedRegistrations = 1
	}
	if !flush || err == nil || registrationCalls != expectedRegistrations {
		t.Fatalf("startup failure = %v/%v, registration calls %d", flush, err, registrationCalls)
	}
	if phase == "partial_init" {
		if _, ok := err.(*RuntimeException); !ok || hostCalls != 0 {
			t.Fatalf("partial initialization failure changed precedence: %T/%v, host calls %d", err, err, hostCalls)
		}
	} else {
		if hostCalls != 1 {
			t.Fatalf("hostname calls = %d", hostCalls)
		}
		if phase == "missing_coordinator" {
			if _, ok := err.(*NullPointerException); !ok {
				t.Fatalf("missing coordinator failure = %T/%v", err, err)
			}
		} else if err != failure {
			t.Fatalf("cleanup replaced the original failure: %T/%v", err, err)
		}
	}
	if _, err := os.Stat(filepath.Join(metadir, "FPSet12345_0.fp")); err != nil {
		t.Fatal("rollback skipped or removed the initialized sibling", err)
	}
	if phase == "partial_init" && strings.Contains(output.String(), "FPSet instance type") {
		t.Fatal("failed initialization printed the later instance announcement")
	}
	if strings.Contains(output.String(), "is ready.") != (phase == "report_failure") || strings.Contains(output.String(), "Progress:") != (phase == "report_failure") {
		t.Fatal("failure changed source readiness/reporting order")
	}
	if attempted {
		defer registered.Close()
		if runtime.GOOS == "linux" && nativeFingerprintStartupDescriptors(t, metadir) == 0 {
			t.Fatal("registration/reporting failure prematurely released published storage")
		}
		if registered.Put(17) || !registered.Contains(17) {
			t.Fatal("registration/reporting failure stopped owned storage")
		}
		registered.Close()
	}
	requireNoDistributedConstructorDescriptors(t, metadir)
}

func nativeFingerprintStartupDescriptors(t *testing.T, directory string) int {
	t.Helper()
	entries, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, entry := range entries {
		path, err := os.Readlink(filepath.Join("/proc/self/fd", entry.Name()))
		if err == nil && strings.HasPrefix(path, directory+string(os.PathSeparator)) {
			count++
		}
	}
	return count
}
