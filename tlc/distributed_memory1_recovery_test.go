package tlc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// No upstream SetOfLong test exists. These bounded checks retain the source
// recover method's field assignments and ordinary put/grow behavior, including
// the serialized count being incremented rather than normalized.
func memory1RecoveryBytes(count, length, threshold int32, zero bool, keys ...int64) []byte {
	data := make([]byte, 13+8*len(keys))
	binary.BigEndian.PutUint32(data[0:4], uint32(count))
	binary.BigEndian.PutUint32(data[4:8], uint32(length))
	binary.BigEndian.PutUint32(data[8:12], uint32(threshold))
	if zero {
		data[12] = 1
	}
	for i, key := range keys {
		binary.BigEndian.PutUint64(data[13+i*8:], uint64(key))
	}
	return data
}

func TestDistributedMemory1RecoveryRetainsSourceCountAndGrowth(t *testing.T) {
	for _, test := range []struct {
		name                     string
		data                     []byte
		count, length, threshold int
		zero                     bool
		keys                     []int64
	}{
		{"count_increment", memory1RecoveryBytes(2, 9, 4, false, 41, 97), 4, 9, 4, false, []int64{41, 97}},
		{"zero_count_increment", memory1RecoveryBytes(2, 9, 4, true, 41), 3, 9, 4, true, []int64{0, 41}},
		{"growth_with_zero", memory1RecoveryBytes(2, 9, 2, true, 41), 1, 19, 9, true, []int64{0, 41}},
		{"empty_array", memory1RecoveryBytes(0, 0, 0, false), 0, 0, 0, false, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			set := NewSetOfLong(5)
			err := set.Recover(NewValueInputStreamWithoutHandles(bytes.NewReader(test.data)))
			if err != nil {
				t.Fatal(err)
			}
			if set.count != test.count || set.length != test.length || set.thresh != test.threshold || set.hasZero != test.zero || len(set.table) != test.length {
				t.Fatalf("recovered state = %#v", set)
			}
			for _, key := range test.keys {
				if !set.Contains(key) {
					t.Fatalf("lost fingerprint %d", key)
				}
			}
		})
	}
}

func TestDistributedMemory1RecoveryRetainsPartialFieldMutation(t *testing.T) {
	data := memory1RecoveryBytes(2, 9, 4, false, 41, 97)
	for length := 0; length < len(data); length++ {
		t.Run(fmtInt(length), func(t *testing.T) {
			old := []int64{17, 0, 0, 0, 0}
			set := &SetOfLong{count: 7, length: 5, thresh: 2, hasZero: true, table: old}
			err := set.Recover(NewValueInputStreamWithoutHandles(bytes.NewReader(data[:length])))
			if _, ok := err.(*EOFException); !ok {
				t.Fatalf("partial checkpoint = %v", err)
			}
			count, tableLength, threshold, zero := 7, 5, 2, true
			if length >= 4 {
				count = 2
			}
			if length >= 8 {
				tableLength = 9
			}
			if length >= 12 {
				threshold = 4
			}
			if length >= 13 {
				zero = false
			}
			if length >= 21 {
				count = 3
			}
			if set.count != count || set.length != tableLength || set.thresh != threshold || set.hasZero != zero {
				t.Fatalf("partial fields = %#v", set)
			}
			if length < 13 {
				if &set.table[0] != &old[0] || !reflect.DeepEqual(set.table, old) {
					t.Fatal("failed header replaced existing table")
				}
			} else {
				if len(set.table) != 9 || &set.table[0] == &old[0] || set.Contains(41) != (length >= 21) || set.Contains(97) {
					t.Fatal("failed record changed partial reconstructed table")
				}
			}
		})
	}
}

func TestDistributedMemory1RecoveryNegativeLengthRetainsHeader(t *testing.T) {
	old := []int64{17, 0, 0, 0, 0}
	set := &SetOfLong{count: 7, length: 5, thresh: 2, hasZero: false, table: old}
	defer func() {
		failure := recover()
		if _, ok := failure.(*NegativeArraySizeException); !ok {
			t.Fatalf("negative array failure = %v", failure)
		}
		if set.count != 2 || set.length != -1 || set.thresh != 4 || !set.hasZero || &set.table[0] != &old[0] {
			t.Fatalf("negative length changed header/table mutation: %#v", set)
		}
	}()
	if err := set.Recover(NewValueInputStreamWithoutHandles(bytes.NewReader(memory1RecoveryBytes(2, -1, 4, true)))); err != nil {
		t.Fatal(err)
	}
}

func TestDistributedMemory1FileRecoveryThroughNativeRPC(t *testing.T) {
	for _, complete := range []bool{false, true} {
		t.Run(fmt.Sprint(complete), func(t *testing.T) {
			captureFailoverToolIO(t, ToolIOTool)
			directory := t.TempDir()
			data := memory1RecoveryBytes(2, 9, 4, false, 41, 97)
			if !complete {
				data = data[:22]
			}
			if err := os.WriteFile(filepath.Join(directory, "job.fp.chkpt"), data, 0600); err != nil {
				t.Fatal(err)
			}
			storage := NewMemFPSet1(NewFPSetConfiguration())
			storage.Init(1, directory, "store")
			_, client := startFingerprintRPC(t, NewLocalFingerprintEndpoint(storage))
			manager := NewDistributedFPSetManager(client)
			// Source manager catches the incomplete file's I/O failure and
			// continues, retaining the memory set's already completed mutation.
			if err := manager.Recover("job"); err != nil {
				t.Fatal(err)
			}
			wantSize := uint64(3)
			if complete {
				wantSize = 4
			}
			if storage.Size() != wantSize || !storage.Contains(41) || storage.Contains(97) != complete || storage.set.length != 9 || storage.set.thresh != 4 {
				t.Fatal("native file recovery normalized source count or lost partial mutation")
			}
			messages := ToolIOGetAllMessages()
			if complete {
				if len(messages) != 0 {
					t.Fatalf("successful recovery output = %v", messages)
				}
			} else if len(messages) != 1 || !strings.Contains(messages[0], "Failed to checkpoint the fingerprint server") {
				t.Fatalf("source manager I/O diagnostic = %v", messages)
			}
		})
	}
}
