package tlc

import "fmt"

// Endpoint references identify named Go TCP objects, never Java remote stubs.
type DistributedEndpointReference struct{ Address, Object string }
type DistributedManagerPayload struct {
	Nil                bool
	Broken             bool
	Mask               uint64
	Description        string
	ExpectedNumServers int
	NonDistributed     bool
	Partitions         []int
	Nodes              []DistributedManagerNode
}
type DistributedManagerNode struct {
	Endpoint  DistributedEndpointReference
	Hostname  string
	Available bool
}

func encodeDistributedManager(manager *DistributedFPSetManager, reference func(DistributedFingerprintEndpoint) (DistributedEndpointReference, error)) (*DistributedManagerPayload, error) {
	if manager == nil {
		return &DistributedManagerPayload{Nil: true}, nil
	}
	manager = manager.snapshotForWorker()
	payload := &DistributedManagerPayload{Broken: manager.managerIsBroken, Mask: manager.Mask, Description: manager.Description, ExpectedNumServers: manager.ExpectedNumServers, NonDistributed: manager.NonDistributed, Partitions: make([]int, len(manager.fpSets))}
	ids := make(map[*distributedFPSets]int)
	for i, entry := range manager.fpSets {
		if entry == nil {
			continue
		}
		id := ids[entry]
		if id == 0 {
			ref, err := reference(entry.set)
			if err != nil {
				return nil, err
			}
			id = len(payload.Nodes) + 1
			ids[entry] = id
			payload.Nodes = append(payload.Nodes, DistributedManagerNode{Endpoint: ref, Hostname: entry.hostname, Available: entry.available})
		}
		payload.Partitions[i] = id
	}
	return payload, nil
}

func decodeDistributedManager(payload *DistributedManagerPayload, resolve func(DistributedEndpointReference) (DistributedFingerprintEndpoint, error)) (*DistributedFPSetManager, error) {
	if payload == nil {
		return nil, fmt.Errorf("missing fingerprint manager payload")
	}
	if payload.Nil {
		if len(payload.Nodes) != 0 || len(payload.Partitions) != 0 {
			return nil, fmt.Errorf("null manager contains partitions")
		}
		return nil, nil
	}
	for _, id := range payload.Partitions {
		if id < 0 || id > len(payload.Nodes) {
			return nil, fmt.Errorf("invalid fingerprint wrapper reference %d", id)
		}
	}
	for _, node := range payload.Nodes {
		if node.Endpoint.Address == "" || node.Endpoint.Object == "" {
			return nil, fmt.Errorf("incomplete fingerprint endpoint reference")
		}
	}
	manager := &DistributedFPSetManager{managerIsBroken: payload.Broken, Mask: payload.Mask, Description: payload.Description, ExpectedNumServers: payload.ExpectedNumServers, NonDistributed: payload.NonDistributed, fpSets: make([]*distributedFPSets, len(payload.Partitions))}
	nodes := make([]*distributedFPSets, len(payload.Nodes))
	endpoints := make(map[DistributedEndpointReference]DistributedFingerprintEndpoint)
	for i, node := range payload.Nodes {
		endpoint := endpoints[node.Endpoint]
		if endpoint == nil {
			var err error
			endpoint, err = resolve(node.Endpoint)
			if err != nil {
				return nil, err
			}
			endpoints[node.Endpoint] = endpoint
		}
		nodes[i] = &distributedFPSets{set: endpoint, hostname: node.Hostname, available: node.Available}
	}
	for i, id := range payload.Partitions {
		if id != 0 {
			manager.fpSets[i] = nodes[id-1]
		}
	}
	return manager, nil
}
