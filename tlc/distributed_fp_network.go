package tlc

import (
	"errors"
	"fmt"
)

// A fingerprint server uses the same native host as workers, so a combined
// process can publish both roles without another transport or evaluator copy.
type DistributedFPServerNetwork struct{ *DistributedWorkerNetwork }

func NewDistributedFPServerNetwork(listenAddress, advertisedAddress string) (*DistributedFPServerNetwork, error) {
	network, err := NewDistributedWorkerNetwork(listenAddress, advertisedAddress)
	if err != nil {
		return nil, err
	}
	return &DistributedFPServerNetwork{network}, nil
}
func (s *DistributedRPCServer) UnregisterFingerprint(name string) {
	s.mu.Lock()
	delete(s.fingerprints, name)
	s.mu.Unlock()
}
func (n *DistributedWorkerNetwork) FPEnvironment(base DistributedFPServerEnvironment) DistributedFPServerEnvironment {
	base.Lookup = n.Discovery.Lookup
	base.LocalHostName = func() (string, error) { return n.workerAddress.Hostname, nil }
	base.RegisterFPSet = func(server DistributedServerEndpoint, endpoint DistributedFingerprintEndpoint, hostname string) error {
		coordinator, ok := server.(*NetworkServerEndpoint)
		if !ok {
			return errors.New("native fingerprint registration requires a discovered TCP coordinator")
		}
		local, ok := endpoint.(*LocalFingerprintEndpoint)
		if !ok {
			return fmt.Errorf("native fingerprint publication requires owned local storage, got %T", endpoint)
		}
		name := fmt.Sprintf("fingerprint-%d", n.sequence.Add(1))
		n.fingerprintsMu.Lock()
		if err := n.Host.RegisterFingerprint(name, endpoint); err != nil {
			n.fingerprintsMu.Unlock()
			return err
		}
		if n.fingerprints == nil {
			n.fingerprints = make(map[FPSet][]string)
		}
		n.fingerprints[local.Set] = append(n.fingerprints[local.Set], name)
		if n.ownedFPSets == nil {
			n.ownedFPSets = make(map[FPSet]bool)
		}
		if !n.ownedFPSets[local.Set] {
			n.ownedFPSets[local.Set] = true
			n.fpOwners = append(n.fpOwners, local.Set)
		}
		n.fingerprintsMu.Unlock()
		return coordinator.RegisterFPSetReference(DistributedEndpointReference{Address: n.Address, Object: name}, hostname)
	}
	base.UnpublishFPSet = func(set FPSet, force bool) {
		n.fingerprintsMu.Lock()
		names := n.fingerprints[set]
		delete(n.fingerprints, set)
		n.fingerprintsMu.Unlock()
		for _, name := range names {
			n.Host.UnregisterFingerprint(name)
		}
	}
	return base
}
