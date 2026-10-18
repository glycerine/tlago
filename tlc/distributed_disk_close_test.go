package tlc

import (
	"encoding/binary"
	"os"
	"testing"
)

// DiskFPSet.close retains closed reader slots and ignores individual close I/O
// failures. No enabled original method covers later calls through an endpoint.
func TestDistributedDiskCloseRetainsReaderOwners(t *testing.T) {
	for _, msb := range []bool{false, true} {
		for _, remote := range []bool{false, true} {
			for _, phase := range []string{"success", "worker_close", "pool_close"} {
				name := map[bool]string{false: "lsb", true: "msb"}[msb] + "/" + map[bool]string{false: "local", true: "tcp"}[remote] + "/" + phase
				t.Run(name, func(t *testing.T) {
					config := NewFPSetConfiguration()
					config.SetMemory(1 << 20)
					var disk *DiskFPSet
					if msb {
						disk = NewMSBDiskFPSet(config).DiskFPSet
					} else {
						disk = NewLSBDiskFPSet(config).DiskFPSet
					}
					directory := t.TempDir()
					disk.Init(2, directory, "store")
					t.Cleanup(disk.Close)
					var data [24]byte
					for i, fp := range []uint64{41, 43, 97} {
						binary.BigEndian.PutUint64(data[i*8:], fp)
					}
					if err := os.WriteFile(disk.chkptName("job", "chkpt"), data[:], 0600); err != nil {
						t.Fatal(err)
					}
					if err := disk.RecoverFile("job"); err != nil {
						t.Fatal(err)
					}
					workers := append([]*BufferedRandomAccessFile(nil), disk.braf...)
					pool := append([]*BufferedRandomAccessFile(nil), disk.brafPool...)
					disk.poolIndex = 3
					if phase == "worker_close" {
						if err := workers[0].file.Close(); err != nil {
							t.Fatal(err)
						}
					} else if phase == "pool_close" {
						if err := pool[1].file.Close(); err != nil {
							t.Fatal(err)
						}
					}
					var endpoint DistributedFingerprintEndpoint = NewLocalFingerprintEndpoint(disk)
					if remote {
						_, endpoint = startFingerprintRPC(t, endpoint)
					}
					if err := endpoint.Close(); err != nil {
						t.Fatal("source close must ignore individual reader I/O failures", err)
					}
					if len(disk.braf) != 2 || len(disk.brafPool) != len(pool) || disk.poolIndex != 0 || disk.GetReaderWriterCnt() != 2+len(pool) {
						t.Fatal("close cleared reader slots or retained the old pool cursor")
					}
					for i, readers := range [][]*BufferedRandomAccessFile{disk.braf, disk.brafPool} {
						old := [][]*BufferedRandomAccessFile{workers, pool}[i]
						for j, reader := range readers {
							if reader != old[j] || !reader.closed {
								t.Fatal("close replaced or failed to release a source reader owner")
							}
						}
					}
					if snapshot := disk.readers.Load(); snapshot == nil || len(*snapshot) != 2 || (*snapshot)[0] != workers[0] || (*snapshot)[1] != workers[1] {
						t.Fatal("close lost the native snapshot of closed source readers")
					}
					if count, err := endpoint.Size(); err != nil || count != 3 {
						t.Fatalf("close changed fingerprint count: %d/%v", count, err)
					}
					// Named recovery leaves the memory table empty. The middle
					// fingerprint requires actual disk I/O, not an index endpoint hit.
					found, err := invokeFingerprintEndpoint(func() (bool, error) { return endpoint.Contains(43) })
					if found || !isJavaIOException(err) {
						t.Fatalf("lookup reopened closed storage: %v/%v", found, err)
					}
					requireNoDistributedConstructorDescriptors(t, directory)
					if err := endpoint.Close(); err != nil || len(disk.braf) != 2 || len(disk.brafPool) != len(pool) || disk.poolIndex != 0 {
						t.Fatal("repeated close changed source ownership or cursor", err)
					}
				})
			}
		}
	}
}
