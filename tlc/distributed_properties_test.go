package tlc

import (
	"reflect"
	"testing"
)

func TestDistributedStartupProperties(t *testing.T) {
	const key = "tlago.unit.distributed.startup"
	// Restore the process store, including absence, rather than leaving a
	// startup property for unrelated tests in this package.
	tlcSystemProperties.Lock()
	old, present := tlcSystemProperties.values[key]
	tlcSystemProperties.Unlock()
	t.Cleanup(func() {
		tlcSystemProperties.Lock()
		defer tlcSystemProperties.Unlock()
		if present {
			tlcSystemProperties.values[key] = old
		} else {
			delete(tlcSystemProperties.values, key)
		}
	})
	args, err := ExtractDistributedStartupProperties([]string{"-D" + key + "=a=b", "-config", "Spec.cfg", "Spec"})
	if err != nil || !reflect.DeepEqual(args, []string{"-config", "Spec.cfg", "Spec"}) || DistributedSystemProperty(key, "absent") != "a=b" {
		t.Fatalf("args %v, error %v, property %q", args, err, DistributedSystemProperty(key, "absent"))
	}
	if _, err = ExtractDistributedStartupProperties([]string{"-D" + key + "=changed", "-D"}); err == nil || DistributedSystemProperty(key, "") != "a=b" {
		t.Fatal("malformed startup must not partially apply properties")
	}
}
