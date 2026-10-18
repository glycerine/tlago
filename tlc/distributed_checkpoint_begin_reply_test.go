package tlc

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
)

type lostCheckpointBeginReply struct {
	*LocalFingerprintEndpoint
	host    *DistributedRPCServer
	begins  atomic.Int32
	commits atomic.Int32
}

func (e *lostCheckpointBeginReply) BeginChkptFile(name string) error {
	e.begins.Add(1)
	if err := e.LocalFingerprintEndpoint.BeginChkptFile(name); err != nil {
		return err
	}
	return e.host.Close() // Real pending files exist; the reply cannot reach the caller.
}

func (e *lostCheckpointBeginReply) CommitChkptFile(name string) error {
	e.commits.Add(1)
	return e.LocalFingerprintEndpoint.CommitChkptFile(name)
}

// Java's manager reports checkpoint I/O failure without replay or reassignment.
// Recovering an incomplete snapshot follows the storage-specific failure path:
// direct MemFPSet I/O is reported and ignored; MultiFPSet wraps child I/O and
// propagates it. No enabled upstream method exercises this transport boundary.
func TestFingerprintRPCCompletedBeginReplyLoss(t *testing.T) {
	for _, backend := range []string{"mem", "lsb", "msb"} {
		t.Run(backend, func(t *testing.T) {
			captureFailoverToolIO(t, ToolIOTool)
			directories := []string{t.TempDir(), t.TempDir()}
			newStore := func(directory string) FPSet {
				var set FPSet
				if backend == "mem" {
					set = NewMemFPSet()
				} else {
					implementation := "tlc2.tool.fp.LSBDiskFPSet"
					if backend == "msb" {
						implementation = "tlc2.tool.fp.MSBDiskFPSet"
					}
					config := NewFPSetConfigurationWithRatioAndImplementation(1, implementation)
					config.SetFPBits(1)
					config.SetMemory(1 << 20)
					set = NewFPSet(config)
				}
				set.Init(1, directory, "Spec")
				t.Cleanup(set.Close)
				return set
			}
			primary, healthy := newStore(directories[0]), newStore(directories[1])
			for _, set := range []FPSet{primary, healthy} {
				set.Put(41)
				set.Put(uint64(1)<<63 | 43)
			}
			endpoint := &lostCheckpointBeginReply{LocalFingerprintEndpoint: NewLocalFingerprintEndpoint(primary)}
			host, failed := startFingerprintRPC(t, endpoint)
			endpoint.host = host
			_, available := startFingerprintRPC(t, NewLocalFingerprintEndpoint(healthy))
			manager := NewDistributedFPSetManager(failed, available)
			manager.fpSets[0].hostname = "lost-begin"
			first, second := manager.entry(0), manager.entry(1)
			if err := manager.Checkpoint("job"); err != nil {
				t.Fatal(err)
			}
			if endpoint.begins.Load() != 1 || endpoint.commits.Load() != 0 || manager.NumOfAliveServers() != 2 || manager.entry(0) != first || manager.entry(1) != second {
				t.Fatal("lost begin reply was replayed, committed or reassigned")
			}
			const warning = "Error: Failed to checkpoint the fingerprint server at lost-begin. This server might be down."
			if strings.Count(strings.Join(ToolIOGetAllMessages(), "\n"), warning) != 1 {
				t.Fatal("missing exact source checkpoint warning")
			}
			children := 1
			if backend != "mem" {
				children = 2
			}
			pending := make(map[string][]byte)
			for child := 0; child < children; child++ {
				name := "job"
				if children == 2 {
					name += "_" + strconv.Itoa(child)
				}
				filename := filepath.Join(directories[0], name+".fp.tmp")
				data, err := os.ReadFile(filename)
				if err != nil || len(data)%8 != 0 || len(data) == 0 {
					t.Fatalf("completed begin lacks real pending snapshot: %x/%v", data, err)
				}
				if children == 2 && (len(data) != 8 || binary.BigEndian.Uint64(data) != []uint64{41, 43}[child]) {
					t.Fatal("nested pending snapshot lost partition identity")
				}
				if children == 1 {
					if len(data) != 16 {
						t.Fatal("memory pending snapshot lost fingerprint count")
					}
					first, second := binary.BigEndian.Uint64(data[:8]), binary.BigEndian.Uint64(data[8:])
					if !((first == 41 && second == uint64(1)<<63|43) || (second == 41 && first == uint64(1)<<63|43)) {
						t.Fatal("memory pending snapshot lost fingerprint membership")
					}
				}
				pending[filename] = data
				if _, err := os.Stat(filepath.Join(directories[0], name+".fp.chkpt")); !os.IsNotExist(err) {
					t.Fatalf("lost begin reply promoted uncommitted snapshot: %v", err)
				}
			}
			primary.Close()
			healthy.Close()
			primary, healthy = newStore(directories[0]), newStore(directories[1])
			_, failed = startFingerprintRPC(t, NewLocalFingerprintEndpoint(primary))
			_, available = startFingerprintRPC(t, NewLocalFingerprintEndpoint(healthy))
			manager = NewDistributedFPSetManager(failed, available)
			manager.fpSets[0].hostname = "lost-begin"
			err := manager.Recover("job")
			if backend == "mem" {
				if err != nil || primary.Size() != 0 || healthy.Size() != 2 || strings.Count(strings.Join(ToolIOGetAllMessages(), "\n"), warning) != 2 {
					t.Fatalf("direct recovery I/O did not retain source continuation: %v", err)
				}
			} else if err == nil || isJavaIOException(err) || !strings.Contains(err.Error(), "recover checkpoint job_0") || primary.Size() != 0 || healthy.Size() != 0 {
				t.Fatalf("nested recovery failure did not stop before the healthy registration: %v", err)
			}
			if manager.NumOfAliveServers() != 2 {
				t.Fatal("recovery changed registration availability")
			}
			for filename, data := range pending {
				after, err := os.ReadFile(filename)
				if err != nil || !bytes.Equal(after, data) {
					t.Fatalf("recovery changed pending snapshot: %v", err)
				}
			}
		})
	}
}
