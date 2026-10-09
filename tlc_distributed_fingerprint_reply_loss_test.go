package tlago

import (
	"fmt"
	"os"
	"sync"

	"github.com/glycerine/tlago/tlc"
)

type nativeFingerprintHeldPutReply struct {
	tlc.DistributedFingerprintEndpoint
	once sync.Once
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
		var held tlc.DistributedFingerprintEndpoint = &nativeFingerprintHeldPutReply{DistributedFingerprintEndpoint: endpoint}
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
