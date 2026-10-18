package tlc

import "testing"

// Java FPSetManager(FPSetRMI) retains one endpoint even when the supplied storage
// owns nested tables. There is no direct original test of this Go adapter.
func TestDistributedLocalMultiFPSetRetainsStorageRouting(t *testing.T) {
	for _, implementation := range []string{"tlc2.tool.fp.MemFPSet", "tlc2.tool.fp.MSBDiskFPSet", "tlc2.tool.fp.LSBDiskFPSet"} {
		t.Run(implementation, func(t *testing.T) {
			config := NewFPSetConfigurationWithRatioAndImplementation(1, implementation)
			config.SetMemory(1 << 20)
			config.SetFPBits(1)
			storage := NewMultiFPSet(config)
			storage.Init(1, t.TempDir(), "fingerprints")
			t.Cleanup(storage.Close)
			manager := NewDistributedFPSetManagerFromFPSet(storage)
			if manager.NumOfServers() != 1 || manager.NumOfAliveServers() != 1 {
				t.Errorf("local nested storage became %d servers/%d alive, want one endpoint", manager.NumOfServers(), manager.NumOfAliveServers())
			}
			// Low-bit manager routing and high-bit storage routing deliberately
			// disagree for both values. Only the supplied storage may choose its
			// internal table; the distributed manager sees one server.
			fingerprints := []uint64{3, uint64(1)<<63 | 2}
			for _, fp := range fingerprints {
				if manager.Put(fp) || !manager.Contains(fp) {
					t.Errorf("fresh manager insertion/membership failed for %d", fp)
				}
				if !storage.Contains(fp) {
					t.Errorf("manager insertion bypassed storage routing for %d", fp)
				}
				if !storage.Put(fp) {
					t.Errorf("storage did not see manager insertion for %d", fp)
				}
				if !manager.Put(fp) {
					t.Errorf("repeated manager insertion was new for %d", fp)
				}
			}
			if storage.Size() != 2 || manager.Size() != 2 || storage.Sets[0].Size() != 1 || storage.Sets[1].Size() != 1 {
				t.Errorf("nested membership counts: storage %d, manager %d, children %d/%d", storage.Size(), manager.Size(), storage.Sets[0].Size(), storage.Sets[1].Size())
			}
			if manager.NumOfServers() == 1 {
				endpoint, ok := manager.entries()[0].set.(*LocalFingerprintEndpoint)
				if !ok || endpoint.Set != storage {
					t.Fatal("manager replaced its supplied storage owner")
				}
				block := NewLongVec()
				block.AddElement(5)
				block.AddElement(int64(-9223372036854775804)) // high bit plus 4
				requireJavaBitCounts(t, manager.ContainsBlock([]*LongVec{block}), 2)
				requireJavaBitCounts(t, manager.PutBlock([]*LongVec{block}), 2)
				requireJavaBitCounts(t, manager.ContainsBlock([]*LongVec{block}), 0)
				requireJavaBitCounts(t, manager.PutBlock([]*LongVec{block}), 0)
				if !storage.Contains(5) || !storage.Contains(uint64(1)<<63|4) || storage.Size() != 4 || manager.Size() != 4 || storage.Sets[0].Size() != 2 || storage.Sets[1].Size() != 2 {
					t.Fatal("block operations bypassed storage routing or duplicated membership")
				}
			}
		})
	}
}
