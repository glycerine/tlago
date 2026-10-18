package tlc

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"testing"
)

// No original InternTable test covers incomplete records. This native check
// preserves recover's complete-prefix mutations and nullable Assert.fail
// parameter, independently of the source MP general-reporting hang.
func TestDistributedInternRecoveryTruncatedRecord(t *testing.T) {
	var payload []byte
	writeInt := func(n int32) { payload = binary.BigEndian.AppendUint32(payload, uint32(n)) }
	writeInt(777)
	writeInt(41)
	writeInt(2)
	writeInt(8)
	payload = append(payload, "complete"...)
	prefixEnd := len(payload)
	writeInt(52)
	writeInt(3)
	writeInt(4)
	payload = append(payload, "tail"...)
	for retained := 1; retained < len(payload)-prefixEnd; retained++ {
		t.Run(fmt.Sprintf("record_bytes_%d", retained), func(t *testing.T) {
			dir := t.TempDir()
			path := uniqueStringChkptName(dir, "chkpt")
			bytes := payload[:prefixEnd+retained]
			if err := os.WriteFile(path, bytes, 0600); err != nil {
				t.Fatal(err)
			}
			table := NewInternTable(16)
			existing := table.Put("existing")
			var thrown any
			func() {
				defer func() { thrown = recover() }()
				if err := table.Recover(dir); err != nil {
					t.Fatalf("record EOF returned instead of throwing corruption: %v", err)
				}
			}()
			failure, ok := thrown.(*TLCError)
			if !ok || failure.Code != ECSystemCheckpointRecoveryCorrupt || !failure.Runtime || len(failure.NullableParams) != 1 || failure.NullableParams[0] != nil {
				t.Fatalf("record corruption = %#v", thrown)
			}
			if !strings.Contains(failure.Error(), "checkpoint file is probably corrupted") || !strings.HasSuffix(failure.Error(), "%1%") {
				t.Fatalf("source nullable corruption message changed: %q", failure.Error())
			}
			complete := table.Find("complete")
			if complete == nil || complete.Token() != 41 || complete.loc != 2 || table.Find("tail") != nil || table.Find("existing") != existing || table.count != 2 {
				t.Fatal("failed record lost prior entries or published a partial string")
			}
			if next := table.Put("after-recovery").Token(); next != 778 {
				t.Fatalf("recovered count header not retained after failure: next token %d", next)
			}
			got, err := os.ReadFile(path)
			if err != nil || string(got) != string(bytes) {
				t.Fatalf("failed recovery changed committed bytes: %v", err)
			}
		})
	}
}
