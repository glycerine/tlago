package tlago

import (
	"bytes"
	"strings"
	"testing"
)

func TestDistributedCLIHelp(t *testing.T) {
	for _, role := range []string{"server", "worker", "fpserver", "worker-fpserver"} {
		for _, args := range [][]string{{role, "-help"}, {"help", role}} {
			var out, errors bytes.Buffer
			if status := RunCLI(args, &out, &errors); status != ExitOK || errors.Len() != 0 {
				t.Fatalf("%v: status %d, errors %s", args, status, errors.String())
			}
			for _, required := range []string{"Usage: tlago " + role, "Java counterpart:", "native Go TCP RPC", "TLCServer.port", "advertiseHost"} {
				if !strings.Contains(out.String(), required) {
					t.Fatalf("%v: missing %q in help", args, required)
				}
			}
			if role == "server" && (!strings.Contains(out.String(), "-recover DIR") || strings.Contains(out.String(), "-simulate")) {
				t.Fatal("coordinator help must describe its actual source option subset")
			}
		}
	}
}

func TestDistributedCLIRejectsInvalidStartup(t *testing.T) {
	for _, args := range [][]string{{"worker", "-D=oops", "host"}, {"fpserver", "host", "extra"}, {"worker-fpserver", "-D"}} {
		var out, errors bytes.Buffer
		if status := RunCLI(args, &out, &errors); status != ExitToolFailure || errors.Len() == 0 || out.Len() != 0 {
			t.Fatalf("%v: status %d, stdout %q, stderr %q", args, status, out.String(), errors.String())
		}
	}
}
