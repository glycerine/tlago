package tlc

import "testing"

// MultiFPSet inherits FPSet.addThread's no-op, but overrides incWorkers to
// visit its children. No original test directly checks this distinction.
func TestDistributedNestedThreadRegistration(t *testing.T) {
	for _, remote := range []bool{false, true} {
		t.Run(map[bool]string{false: "local", true: "tcp"}[remote], func(t *testing.T) {
			first := NewLSBDiskFPSet(NewFPSetConfiguration())
			second := NewMSBDiskFPSet(NewFPSetConfiguration())
			first.Init(1, t.TempDir(), "first")
			t.Cleanup(first.Close)
			second.Init(1, t.TempDir(), "second")
			t.Cleanup(second.Close)
			set := &MultiFPSet{Sets: []FPSet{first, second}, Shift: 63}
			var endpoint DistributedFingerprintEndpoint = NewLocalFingerprintEndpoint(set)
			if remote {
				_, endpoint = startFingerprintRPC(t, endpoint)
			}
			if err := endpoint.AddThread(); err != nil {
				t.Fatal(err)
			}
			if len(first.braf) != 1 || len(second.braf) != 1 {
				t.Fatalf("nested addThread opened child readers: %d/%d, want 1/1", len(first.braf), len(second.braf))
			}
			set.IncWorkers(2)
			if len(first.braf) != 1 || len(second.braf) != 1 {
				t.Fatalf("nested incWorkers changed no-op child readers: %d/%d, want 1/1", len(first.braf), len(second.braf))
			}
		})
	}
}

// LSB/MSB inherit FPSet.incWorkers's no-op; DiskFPSet.addThread independently
// opens one reader. No original Java method directly tests this distinction.
func TestDistributedHeapFingerprintWorkerRegistration(t *testing.T) {
	for _, msb := range []bool{false, true} {
		for _, remote := range []bool{false, true} {
			name := map[bool]string{false: "lsb", true: "msb"}[msb] + "/" + map[bool]string{false: "local", true: "tcp"}[remote]
			t.Run(name, func(t *testing.T) {
				config := NewFPSetConfiguration()
				config.SetMemory(1 << 20)
				var disk *DiskFPSet
				if msb {
					disk = NewMSBDiskFPSet(config).DiskFPSet
				} else {
					disk = NewLSBDiskFPSet(config).DiskFPSet
				}
				disk.Init(1, t.TempDir(), "store")
				t.Cleanup(disk.Close)
				var endpoint DistributedFingerprintEndpoint = NewLocalFingerprintEndpoint(disk)
				if remote {
					_, endpoint = startFingerprintRPC(t, endpoint)
				}
				worker := disk.braf[0]
				pool := append([]*BufferedRandomAccessFile(nil), disk.brafPool...)
				disk.poolIndex = 2
				for _, count := range []int{-1, 0, 2} {
					disk.IncWorkers(count)
					if len(disk.braf) != 1 || disk.braf[0] != worker || worker.closed {
						t.Fatalf("incWorkers(%d) changed the source reader owner", count)
					}
				}
				filename := disk.fpFilename
				disk.fpFilename += ".missing"
				// The no-op must not access storage even when reader opening would fail.
				disk.IncWorkers(2)
				if err := endpoint.AddThread(); !isJavaIOException(err) {
					t.Fatalf("addThread missing-file failure = %v", err)
				}
				if len(disk.braf) != 1 || disk.braf[0] != worker || worker.closed {
					t.Fatal("failed addThread changed the existing reader array")
				}
				disk.fpFilename = filename
				if err := endpoint.AddThread(); err != nil {
					t.Fatal(err)
				}
				if len(disk.braf) != 2 || disk.braf[0] != worker || worker.closed || disk.braf[1] == worker || disk.braf[1].closed {
					t.Fatal("addThread did not retain old ownership and open exactly one reader")
				}
				if len(disk.brafPool) != len(pool) || disk.poolIndex != 2 {
					t.Fatal("thread registration changed pooled reader allocation/cursor")
				}
				for i, reader := range pool {
					if disk.brafPool[i] != reader || reader.closed {
						t.Fatalf("thread registration changed pooled reader %d", i)
					}
				}
				if snapshot := disk.readers.Load(); snapshot == nil || len(*snapshot) != 2 || (*snapshot)[0] != worker || (*snapshot)[1] != disk.braf[1] {
					t.Fatal("native reader snapshot lost successful addThread publication")
				}
			})
		}
	}
}

type workerRegistrationChild struct {
	*MemFPSet
	counts []int
}

func (s *workerRegistrationChild) IncWorkers(count int) { s.counts = append(s.counts, count) }

func TestDistributedNestedWorkerRegistrationDispatch(t *testing.T) {
	first := &workerRegistrationChild{MemFPSet: NewMemFPSet()}
	second := &workerRegistrationChild{MemFPSet: NewMemFPSet()}
	set := &MultiFPSet{Sets: []FPSet{first, second}}
	set.IncWorkers(2)
	for _, child := range []*workerRegistrationChild{first, second} {
		if len(child.counts) != 1 || child.counts[0] != 2 {
			t.Fatalf("nested incWorkers did not dispatch once with the unchanged count: %v", child.counts)
		}
	}
}
