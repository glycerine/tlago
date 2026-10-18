package tlago

import (
	"errors"
	"reflect"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// There is no original direct test for failures while reading the coordinator's
// constructor arguments. Preserve left-to-right calls and stop on the first
// failure, before attempting to parse a model or create its tool.
func TestDistributedWorkerAppEndpointFailureOrder(t *testing.T) {
	for _, failAt := range []string{"spec", "config", "deadlock"} {
		t.Run(failAt, func(t *testing.T) {
			failure := errors.New("coordinator call failed")
			endpoint := &appSettingsEndpoint{failAt: failAt, failure: failure}
			app, diags, err := loadDistributedEndpointApp(endpoint, nil, tlc.RuntimeParameters{})
			wantCalls := []string{"spec"}
			if failAt != "spec" {
				wantCalls = append(wantCalls, "config")
			}
			if failAt == "deadlock" {
				wantCalls = append(wantCalls, "deadlock")
			}
			if app != nil || diags != nil || err != failure || !reflect.DeepEqual(endpoint.calls, wantCalls) {
				t.Fatalf("app/diagnostics/error/calls = %v/%v/%v/%v, want nil/nil/original failure/%v", app, diags, err, endpoint.calls, wantCalls)
			}
		})
	}
}

type appSettingsEndpoint struct {
	tlc.DistributedServerEndpoint
	failAt  string
	failure error
	calls   []string
}

func (e *appSettingsEndpoint) call(name string) error {
	e.calls = append(e.calls, name)
	if name == e.failAt {
		return e.failure
	}
	return nil
}

func (e *appSettingsEndpoint) GetSpecFileName() (string, error) {
	return "Spec", e.call("spec")
}

func (e *appSettingsEndpoint) GetConfigFileName() (string, error) {
	return "Spec", e.call("config")
}

func (e *appSettingsEndpoint) GetCheckDeadlock() (bool, error) {
	return true, e.call("deadlock")
}
