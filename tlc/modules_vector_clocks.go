package tlc

import "sort"

type vectorClockGraphNode struct {
	parents  []*vectorClockGraphNode
	children []*vectorClockGraphNode
	value    Value
	clock    Value
	domain   Enumerable
	time     Value
}

type vectorClockHostLog struct {
	host  Value
	nodes []*vectorClockGraphNode
}

func VectorClocksCausalOrder(log Value, opClock Value, opNode Value, opDomain Value) (Value, error) {
	tuple := asTupleValue(log)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "CausalOrder", "sequence", ValuesPPR(log))
	}

	hostLogs := []vectorClockHostLog{}
	for _, entry := range tuple.Elems {
		nodeID, err := EvalOperatorValue(opNode, []Value{entry}, EvalClear)
		if err != nil {
			return nil, err
		}
		clock, err := EvalOperatorValue(opClock, []Value{entry}, EvalClear)
		if err != nil {
			return nil, err
		}
		nodeTime, err := vectorClockSelect(clock, nodeID)
		if err != nil {
			return nil, err
		}
		domainValue, err := EvalOperatorValue(opDomain, []Value{clock}, EvalClear)
		if err != nil {
			return nil, err
		}
		domainSet, err := toSetEnumValue(domainValue)
		if err != nil {
			return nil, err
		}
		domainSet.Normalize()

		idx, err := vectorClockHostIndex(hostLogs, nodeID)
		if err != nil {
			return nil, err
		}
		if idx < 0 {
			hostLogs = append(hostLogs, vectorClockHostLog{host: nodeID})
			idx = len(hostLogs) - 1
		}
		hostLogs[idx].nodes = append(hostLogs[idx].nodes, &vectorClockGraphNode{
			value:  entry,
			clock:  clock,
			time:   nodeTime,
			domain: domainSet,
		})
	}

	for i := range hostLogs {
		var sortErr error
		sort.SliceStable(hostLogs[i].nodes, func(left int, right int) bool {
			if sortErr != nil {
				return false
			}
			cmp, err := hostLogs[i].nodes[left].time.Compare(hostLogs[i].nodes[right].time)
			if err != nil {
				sortErr = err
				return false
			}
			return cmp < 0
		})
		if sortErr != nil {
			return nil, sortErr
		}
	}

	for hostIdx := range hostLogs {
		list := hostLogs[hostIdx].nodes
		globalClock := make([]Value, len(hostLogs))
		for i := range globalClock {
			globalClock[i] = IntZero
		}
		for _, node := range list {
			globalClock[hostIdx] = node.time
			hosts := node.domain.Elements()
			for {
				otherHost := hosts.NextElement()
				if otherHost == nil {
					if err := hosts.Err(); err != nil {
						return nil, err
					}
					break
				}
				time, err := vectorClockSelect(node.clock, otherHost)
				if err != nil {
					return nil, err
				}
				otherIdx, err := vectorClockHostIndex(hostLogs, otherHost)
				if err != nil {
					return nil, err
				}
				if otherIdx < 0 {
					return nil, newTLCError(ECGeneral, "CausalOrder vector clock domain contains unknown host %s", otherHost)
				}
				cmp, err := globalClock[otherIdx].Compare(time)
				if err != nil {
					return nil, err
				}
				if cmp < 0 {
					globalClock[otherIdx] = time
					timeValue, ok := time.(*IntValue)
					if !ok {
						return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "CausalOrder", "integer vector-clock time", ValuesPPR(time))
					}
					parentIdx := int(timeValue.Val) - 1
					if parentIdx < 0 || parentIdx >= len(hostLogs[otherIdx].nodes) {
						return nil, newTLCError(ECGeneral, "CausalOrder vector clock time %s is outside host log %s", time, otherHost)
					}
					node.addParent(hostLogs[otherIdx].nodes[parentIdx])
				}
			}
		}
	}

	sorted := make([]Value, 0, len(tuple.Elems))
	for i := 0; i < len(tuple.Elems); i++ {
		for hostIdx := range hostLogs {
			list := hostLogs[hostIdx].nodes
			if len(list) == 0 {
				continue
			}
			if !list[0].hasParents() {
				node := list[0]
				hostLogs[hostIdx].nodes = list[1:]
				sorted = append(sorted, node.delete())
			}
		}
	}
	return NewTupleValue(sorted), nil
}

func (n *vectorClockGraphNode) hasParents() bool {
	return len(n.parents) != 0
}

func (n *vectorClockGraphNode) addParent(parent *vectorClockGraphNode) {
	if n == nil || parent == nil {
		return
	}
	if !vectorClockNodeSliceContains(n.parents, parent) {
		n.parents = append(n.parents, parent)
	}
	if !vectorClockNodeSliceContains(parent.children, n) {
		parent.children = append(parent.children, n)
	}
}

func (n *vectorClockGraphNode) delete() Value {
	for _, child := range n.children {
		child.parents = vectorClockRemoveNode(child.parents, n)
	}
	return n.value
}

func vectorClockNodeSliceContains(nodes []*vectorClockGraphNode, node *vectorClockGraphNode) bool {
	for _, existing := range nodes {
		if existing == node {
			return true
		}
	}
	return false
}

func vectorClockRemoveNode(nodes []*vectorClockGraphNode, node *vectorClockGraphNode) []*vectorClockGraphNode {
	out := nodes[:0]
	for _, existing := range nodes {
		if existing != node {
			out = append(out, existing)
		}
	}
	return out
}

func vectorClockHostIndex(logs []vectorClockHostLog, host Value) (int, error) {
	for i, log := range logs {
		eq, err := log.host.Equal(host)
		if err != nil {
			return -1, err
		}
		if eq {
			return i, nil
		}
	}
	return -1, nil
}

func vectorClockSelect(container Value, arg Value) (Value, error) {
	switch v := container.(type) {
	case *FcnRcdValue:
		value, err := v.Select(arg)
		if err != nil {
			return nil, err
		}
		if value != nil {
			return value, nil
		}
	case *FcnLambdaValue:
		value, err := v.Select(arg)
		if err != nil {
			return nil, err
		}
		if value != nil {
			return value, nil
		}
	case *TupleValue:
		if value := v.Select(arg); value != nil {
			return value, nil
		}
	case *RecordValue:
		if value := v.Select(arg); value != nil {
			return value, nil
		}
	}
	return nil, newTLCError(ECGeneral, "cannot select %s from vector clock %s", arg, container)
}
