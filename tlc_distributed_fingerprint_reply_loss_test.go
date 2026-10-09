package tlago

import (
	"fmt"
	"os"
	"sync"

	"github.com/glycerine/tlago/tlc"
)

type nativeFingerprintHeldPutReply struct {
	tlc.DistributedFingerprintEndpoint
	once  sync.Once
	flush func() error
}

type nativeFingerprintHeldLookupReply struct {
	tlc.DistributedFingerprintEndpoint
	once sync.Once
}

func (e *nativeFingerprintHeldLookupReply) ContainsBlock(fps *tlc.LongVec) (*tlc.BitVector, error) {
	result, err := e.DistributedFingerprintEndpoint.ContainsBlock(fps)
	if err != nil || result == nil || fps.Size() == 0 {
		return result, err
	}
	e.once.Do(func() {
		fmt.Printf("NATIVE_FP_LOOKUP_COMPLETED_COUNT=%d MISSING=%d\n", fps.Size(), result.TrueCount())
		// Hold the actual lookup answer until the parent kills this host.
		// The worker must recover through the source manager's failover path.
		<-make(chan struct{})
	})
	return result, nil
}

func (e *nativeFingerprintHeldPutReply) PutBlock(fps *tlc.LongVec) (*tlc.BitVector, error) {
	result, err := e.DistributedFingerprintEndpoint.PutBlock(fps)
	if err != nil || result == nil || result.TrueCount() == 0 {
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

func nativeDistributedFingerprintLostReply(args []string, lookup bool) error {
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
		put := &nativeFingerprintHeldPutReply{DistributedFingerprintEndpoint: endpoint}
		if local, ok := endpoint.(*tlc.LocalFingerprintEndpoint); ok {
			if multi, ok := local.Set.(*tlc.MultiFPSet); ok {
				if _, disk := multi.Sets[0].(interface{ GetDiskWriteCnt() uint64 }); disk {
					put.flush = func() error {
						if !multi.CheckInvariant() {
							return fmt.Errorf("disk fingerprint invariant failed before withholding insertion reply")
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
			held = &nativeFingerprintHeldLookupReply{DistributedFingerprintEndpoint: endpoint}
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
