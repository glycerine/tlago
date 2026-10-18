package tlago

import (
	"fmt"
	"os"
	"sync"

	"github.com/glycerine/tlago/tlc"
)

type nativeFingerprintHeldPutReply struct {
	tlc.DistributedFingerprintEndpoint
	once    sync.Once
	flush   func() error
	partial bool
}

type nativeFingerprintHeldLookupReply struct {
	tlc.DistributedFingerprintEndpoint
	once  sync.Once
	flush func() error
	disk  *tlc.MultiFPSet
}

func (e *nativeFingerprintHeldLookupReply) ContainsBlock(fps *tlc.LongVec) (*tlc.BitVector, error) {
	if e.flush != nil && fps != nil && fps.Size() != 0 {
		if err := e.flush(); err != nil {
			return nil, err
		}
	}
	before := nativeFingerprintDiskReadCount(e.disk)
	result, err := e.DistributedFingerprintEndpoint.ContainsBlock(fps)
	if err != nil || result == nil || fps.Size() == 0 {
		return result, err
	}
	if e.disk != nil {
		reads := nativeFingerprintDiskReadCount(e.disk) - before
		if reads == 0 {
			return nil, fmt.Errorf("completed fingerprint lookup did not read disk fingerprints")
		}
		fmt.Printf("NATIVE_FP_DISK_LOOKUP_READS=%d\n", reads)
	}
	e.once.Do(func() {
		fmt.Printf("NATIVE_FP_LOOKUP_COMPLETED_COUNT=%d MISSING=%d\n", fps.Size(), result.TrueCount())
		// Hold the actual lookup answer until the parent kills this host.
		// The worker must recover through the source manager's failover path.
		<-make(chan struct{})
	})
	return result, nil
}

func nativeFingerprintDiskReadCount(store *tlc.MultiFPSet) uint64 {
	var count uint64
	if store != nil {
		for _, child := range store.Sets {
			stats := child.(interface {
				GetDiskSeekCnt() uint64
				GetDiskSeekCache() uint64
			})
			count += stats.GetDiskSeekCnt() + stats.GetDiskSeekCache()
		}
	}
	return count
}

func (e *nativeFingerprintHeldPutReply) PutBlock(fps *tlc.LongVec) (*tlc.BitVector, error) {
	if e.partial && fps != nil && fps.Size() > 1 {
		missing, err := e.DistributedFingerprintEndpoint.ContainsBlock(fps)
		if err != nil {
			return nil, err
		}
		if missing != nil && missing.TrueCount() == fps.Size() {
			// FPSet.putBlock calls put for each fingerprint in input order.
			// Stop at a real prefix of that loop, before processing the suffix.
			prefix := fps.Size() / 2
			for i := 0; i < prefix; i++ {
				seen, err := e.DistributedFingerprintEndpoint.Put(uint64(fps.ElementAt(i)))
				if err != nil || seen {
					return nil, fmt.Errorf("partial insertion failed at %d: seen=%v error=%v", i, seen, err)
				}
			}
			if e.flush != nil {
				if err := e.flush(); err != nil {
					return nil, err
				}
			}
			remaining, err := e.DistributedFingerprintEndpoint.ContainsBlock(fps)
			if err != nil || remaining == nil || remaining.TrueCount() != fps.Size()-prefix {
				return nil, fmt.Errorf("partial insertion membership count changed: %v", err)
			}
			for i := 0; i < fps.Size(); i++ {
				if remaining.Get(i) != (i >= prefix) {
					return nil, fmt.Errorf("partial insertion membership differs at %d", i)
				}
			}
			e.once.Do(func() {
				fmt.Printf("NATIVE_FP_PUT_PARTIAL_PREFIX=%d TOTAL=%d\n", prefix, fps.Size())
				<-make(chan struct{}) // Parent kills this host before any reply.
			})
		}
	}
	result, err := e.DistributedFingerprintEndpoint.PutBlock(fps)
	if e.partial || err != nil || result == nil || result.TrueCount() == 0 {
		return result, err
	}
	if e.flush != nil {
		if err := e.flush(); err != nil {
			return nil, err
		}
	}
	// Verify actual storage mutation before withholding the completed answer.
	missing, err := e.DistributedFingerprintEndpoint.ContainsBlock(fps)
	if err != nil || missing == nil || missing.TrueCount() != 0 {
		return nil, fmt.Errorf("completed insertion did not retain fingerprints: %v", err)
	}
	e.once.Do(func() {
		fmt.Printf("NATIVE_FP_PUT_COMPLETED_NEW=%d\n", result.TrueCount())
		// The parent kills this host after observing the committed membership.
		// No response is fabricated and no accepted mutation is replayed here.
		<-make(chan struct{})
	})
	return result, nil
}

func nativeDistributedFingerprintLostReply(args []string, lookup, partial bool) error {
	args, err := tlc.ExtractDistributedStartupProperties(args)
	if err != nil {
		return err
	}
	network, err := tlc.NewDistributedFPServerNetwork("127.0.0.1:0", "127.0.0.1")
	if err != nil {
		return err
	}
	defer network.Close()
	env := network.FPEnvironment(tlc.DistributedFPServerEnvironment{ToolOut: os.Stdout, SystemOut: os.Stdout, SystemErr: os.Stderr})
	const name = "held-fingerprint-reply"
	env.RegisterFPSet = func(server tlc.DistributedServerEndpoint, endpoint tlc.DistributedFingerprintEndpoint, hostname string) error {
		coordinator, ok := server.(*tlc.NetworkServerEndpoint)
		if !ok {
			return fmt.Errorf("fingerprint reply-loss fixture requires a native TCP coordinator")
		}
		put := &nativeFingerprintHeldPutReply{DistributedFingerprintEndpoint: endpoint, partial: partial}
		if local, ok := endpoint.(*tlc.LocalFingerprintEndpoint); ok {
			if multi, ok := local.Set.(*tlc.MultiFPSet); ok {
				if _, disk := multi.Sets[0].(interface{ GetDiskWriteCnt() uint64 }); disk {
					put.flush = func() error {
						if !multi.CheckInvariant() {
							return fmt.Errorf("disk fingerprint invariant failed before withholding reply")
						}
						var count int64
						for _, child := range multi.Sets {
							stats, ok := child.(interface {
								GetDiskWriteCnt() uint64
								GetFileCnt() int64
								GetTblCnt() int64
							})
							if !ok || stats.GetDiskWriteCnt() == 0 || stats.GetFileCnt() == 0 || stats.GetTblCnt() != 0 {
								return fmt.Errorf("fingerprint child did not flush actual disk membership")
							}
							count += stats.GetFileCnt()
						}
						fmt.Printf("NATIVE_FP_DISK_FLUSHED_CHILDREN=%d COUNT=%d\n", len(multi.Sets), count)
						return nil
					}
				}
			}
		}
		var held tlc.DistributedFingerprintEndpoint = put
		if lookup {
			get := &nativeFingerprintHeldLookupReply{DistributedFingerprintEndpoint: endpoint, flush: put.flush}
			if put.flush != nil {
				get.disk = endpoint.(*tlc.LocalFingerprintEndpoint).Set.(*tlc.MultiFPSet)
			}
			held = get
		}
		if err := network.Host.RegisterFingerprint(name, held); err != nil {
			return err
		}
		return coordinator.RegisterFPSetReference(tlc.DistributedEndpointReference{Address: network.Address, Object: name}, hostname)
	}
	env.UnpublishFPSet = func(tlc.FPSet, bool) { network.Host.UnregisterFingerprint(name) }
	tlc.RunDistributedFPServer(args, env)
	return nil
}
