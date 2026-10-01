package tlc

type graphMode int

const (
	graphModeForward graphMode = iota
	graphModeTranspose
	graphModeSymmetric
)

type graphEndpoints func(Value) (Value, Value, bool, error)

type graphAdjacency struct {
	nodes []Value
	succ  [][]Value
}

func GraphsSimplePath(graph Value) (Value, error) {
	return graphSimplePath("graph", graphDirectedEndpoints, graphModeForward, graph)
}

func GraphsAreConnectedIn(m Value, n Value, graph Value) (Value, error) {
	return graphAreConnectedIn("graph", graphDirectedEndpoints, graphModeForward, m, n, graph)
}

func GraphsIsStronglyConnected(graph Value) (Value, error) {
	g, err := graphToGraph("IsStronglyConnected", "first", "graph", graph)
	if err != nil {
		return nil, err
	}
	nodes, err := graphNodes(g)
	if err != nil {
		return nil, err
	}
	order := nodes.Elems.Len()
	if order == 0 {
		return BoolTrue, nil
	}
	root := nodes.Elems.At(0)

	adj, err := graphBuildAdjacency(g, nodes, graphDirectedEndpoints, graphModeForward)
	if err != nil {
		return nil, err
	}
	reachable, err := graphReachableCount(root, adj)
	if err != nil {
		return nil, err
	}
	if reachable != order {
		return BoolFalse, nil
	}

	radj, err := graphBuildAdjacency(g, nodes, graphDirectedEndpoints, graphModeTranspose)
	if err != nil {
		return nil, err
	}
	reachable, err = graphReachableCount(root, radj)
	if err != nil {
		return nil, err
	}
	if reachable != order {
		return BoolFalse, nil
	}
	return BoolTrue, nil
}

func UndirectedGraphsSimplePath(graph Value) (Value, error) {
	return graphSimplePath("undirected graph", graphUndirectedEndpoints, graphModeSymmetric, graph)
}

func UndirectedGraphsAreConnectedIn(m Value, n Value, graph Value) (Value, error) {
	return graphAreConnectedIn("undirected graph", graphUndirectedEndpoints, graphModeSymmetric, m, n, graph)
}

func UndirectedGraphsConnectedComponents(graph Value) (Value, error) {
	g, err := graphToGraph("ConnectedComponents", "first", "undirected graph", graph)
	if err != nil {
		return nil, err
	}
	nodes, err := graphNodes(g)
	if err != nil {
		return nil, err
	}
	nodeValues := nodes.Elems.ToArray()
	uf := newGraphUnionFind(nodeValues)

	edges, err := graphEdges(g)
	if err != nil {
		return nil, err
	}
	enum := edges.Elements()
	for {
		edge := enum.NextElement()
		if edge == nil {
			if err := enum.Err(); err != nil {
				return nil, err
			}
			break
		}
		from, to, ok, err := graphUndirectedEndpoints(edge)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		fromIdx, err := graphIndexOf(nodeValues, from)
		if err != nil {
			return nil, err
		}
		toIdx, err := graphIndexOf(nodeValues, to)
		if err != nil {
			return nil, err
		}
		if fromIdx >= 0 && toIdx >= 0 {
			uf.union(fromIdx, toIdx)
		}
	}

	reps := []int{}
	components := [][]Value{}
	for i, node := range nodeValues {
		rep := uf.find(i)
		compIdx := -1
		for j, seen := range reps {
			if seen == rep {
				compIdx = j
				break
			}
		}
		if compIdx < 0 {
			reps = append(reps, rep)
			components = append(components, []Value{node})
		} else {
			components[compIdx] = append(components[compIdx], node)
		}
	}

	result := make([]Value, len(components))
	for i, component := range components {
		result[i] = NewSetEnumValue(component, false)
	}
	return NewSetEnumValue(result, false), nil
}

func graphToGraph(op string, argPos string, kind string, value Value) (*RecordValue, error) {
	record := asRecordValue(value)
	if record == nil {
		return nil, graphArgError(op, argPos, kind, "record with a node and an edge field", value)
	}
	node := record.Select(NewStringValue("node"))
	if node == nil {
		return nil, graphArgError(op, argPos, kind, "record whose node field is a set", value)
	}
	if _, err := toSetEnumValue(node); err != nil {
		return nil, graphArgError(op, argPos, kind, "record whose node field is a set", value)
	}
	edge := record.Select(NewStringValue("edge"))
	if edge == nil {
		return nil, graphArgError(op, argPos, kind, "record whose edge field is a set", value)
	}
	if _, err := toSetEnumValue(edge); err != nil {
		return nil, graphArgError(op, argPos, kind, "record whose edge field is a set", value)
	}
	return record, nil
}

func graphArgError(op string, argPos string, kind string, detail string, value Value) error {
	return newTLCErrorCode(ECTLCModuleArgumentError, argPos, op, kind+" "+detail, ValuesPPR(value))
}

func graphNodes(g *RecordValue) (*SetEnumValue, error) {
	value := g.Select(NewStringValue("node"))
	nodes, err := toSetEnumValue(value)
	if err != nil {
		return nil, err
	}
	nodes.Normalize()
	return nodes, nil
}

func graphEdges(g *RecordValue) (*SetEnumValue, error) {
	value := g.Select(NewStringValue("edge"))
	edges, err := toSetEnumValue(value)
	if err != nil {
		return nil, err
	}
	edges.Normalize()
	return edges, nil
}

func graphBuildAdjacency(g *RecordValue, nodes *SetEnumValue, parser graphEndpoints, mode graphMode) (*graphAdjacency, error) {
	nodeValues := nodes.Elems.ToArray()
	adj := &graphAdjacency{nodes: nodeValues, succ: make([][]Value, len(nodeValues))}
	edges, err := graphEdges(g)
	if err != nil {
		return nil, err
	}
	enum := edges.Elements()
	for {
		edge := enum.NextElement()
		if edge == nil {
			if err := enum.Err(); err != nil {
				return nil, err
			}
			return adj, nil
		}
		from, to, ok, err := parser(edge)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		fromIdx, err := graphIndexOf(nodeValues, from)
		if err != nil {
			return nil, err
		}
		toIdx, err := graphIndexOf(nodeValues, to)
		if err != nil {
			return nil, err
		}
		if fromIdx < 0 || toIdx < 0 {
			continue
		}
		if mode == graphModeTranspose {
			from, to = to, from
			fromIdx, toIdx = toIdx, fromIdx
		}
		adj.succ[fromIdx] = append(adj.succ[fromIdx], to)
		if mode == graphModeSymmetric {
			eq, err := from.Equal(to)
			if err != nil {
				return nil, err
			}
			if !eq {
				adj.succ[toIdx] = append(adj.succ[toIdx], from)
			}
		}
	}
}

func graphReachableCount(source Value, adj *graphAdjacency) (int, error) {
	sourceIdx, err := graphIndexOf(adj.nodes, source)
	if err != nil || sourceIdx < 0 {
		return 0, err
	}
	visited := make([]bool, len(adj.nodes))
	queue := []int{sourceIdx}
	visited[sourceIdx] = true
	count := 1
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, succ := range adj.succ[current] {
			idx, err := graphIndexOf(adj.nodes, succ)
			if err != nil {
				return 0, err
			}
			if idx >= 0 && !visited[idx] {
				visited[idx] = true
				count++
				queue = append(queue, idx)
			}
		}
	}
	return count, nil
}

func graphSimplePath(kind string, parser graphEndpoints, mode graphMode, graph Value) (Value, error) {
	g, err := graphToGraph("SimplePath", "first", kind, graph)
	if err != nil {
		return nil, err
	}
	nodes, err := graphNodes(g)
	if err != nil {
		return nil, err
	}
	adj, err := graphBuildAdjacency(g, nodes, parser, mode)
	if err != nil {
		return nil, err
	}

	paths := NewValueVec(0)
	path := []Value{}
	visited := make([]bool, len(adj.nodes))
	for i, start := range adj.nodes {
		path = append(path, start)
		visited[i] = true
		if err := graphExtendSimplePath(i, adj, &path, visited, paths); err != nil {
			return nil, err
		}
		visited[i] = false
		path = path[:len(path)-1]
	}
	return NewSetEnumValueVec(paths, false), nil
}

func graphExtendSimplePath(current int, adj *graphAdjacency, path *[]Value, visited []bool, paths *ValueVec) error {
	paths.Add(NewTupleValue(*path))
	for _, succ := range adj.succ[current] {
		idx, err := graphIndexOf(adj.nodes, succ)
		if err != nil {
			return err
		}
		if idx >= 0 && !visited[idx] {
			visited[idx] = true
			*path = append(*path, succ)
			if err := graphExtendSimplePath(idx, adj, path, visited, paths); err != nil {
				return err
			}
			*path = (*path)[:len(*path)-1]
			visited[idx] = false
		}
	}
	return nil
}

func graphAreConnectedIn(kind string, parser graphEndpoints, mode graphMode, m Value, n Value, graph Value) (Value, error) {
	g, err := graphToGraph("AreConnectedIn", "third", kind, graph)
	if err != nil {
		return nil, err
	}
	nodes, err := graphNodes(g)
	if err != nil {
		return nil, err
	}
	nodeValues := nodes.Elems.ToArray()
	mIdx, err := graphIndexOf(nodeValues, m)
	if err != nil {
		return nil, err
	}
	nIdx, err := graphIndexOf(nodeValues, n)
	if err != nil {
		return nil, err
	}
	if mIdx < 0 || nIdx < 0 {
		return BoolFalse, nil
	}
	eq, err := m.Equal(n)
	if err != nil {
		return nil, err
	}
	if eq {
		return BoolTrue, nil
	}
	adj, err := graphBuildAdjacency(g, nodes, parser, mode)
	if err != nil {
		return nil, err
	}
	reachable, err := graphReachableContains(m, n, adj)
	if err != nil {
		return nil, err
	}
	return NewBoolValue(reachable), nil
}

func graphReachableContains(source Value, target Value, adj *graphAdjacency) (bool, error) {
	sourceIdx, err := graphIndexOf(adj.nodes, source)
	if err != nil || sourceIdx < 0 {
		return false, err
	}
	visited := make([]bool, len(adj.nodes))
	queue := []int{sourceIdx}
	visited[sourceIdx] = true
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, succ := range adj.succ[current] {
			eq, err := succ.Equal(target)
			if err != nil {
				return false, err
			}
			if eq {
				return true, nil
			}
			idx, err := graphIndexOf(adj.nodes, succ)
			if err != nil {
				return false, err
			}
			if idx >= 0 && !visited[idx] {
				visited[idx] = true
				queue = append(queue, idx)
			}
		}
	}
	return false, nil
}

func graphDirectedEndpoints(edge Value) (Value, Value, bool, error) {
	tuple := asTupleValue(edge)
	if tuple == nil || len(tuple.Elems) != 2 {
		return nil, nil, false, nil
	}
	return tuple.Elems[0], tuple.Elems[1], true, nil
}

func graphUndirectedEndpoints(edge Value) (Value, Value, bool, error) {
	set, err := toSetEnumValue(edge)
	if err != nil {
		return nil, nil, false, nil
	}
	set.Normalize()
	size := set.Elems.Len()
	if size == 0 || size > 2 {
		return nil, nil, false, nil
	}
	from := set.Elems.At(0)
	to := from
	if size == 2 {
		to = set.Elems.At(1)
	}
	return from, to, true, nil
}

func graphIndexOf(values []Value, value Value) (int, error) {
	for i, elem := range values {
		eq, err := elem.Equal(value)
		if err != nil {
			return -1, err
		}
		if eq {
			return i, nil
		}
	}
	return -1, nil
}

type graphUnionFind struct {
	parent []int
}

func newGraphUnionFind(values []Value) *graphUnionFind {
	parent := make([]int, len(values))
	for i := range parent {
		parent[i] = i
	}
	return &graphUnionFind{parent: parent}
}

func (u *graphUnionFind) find(idx int) int {
	current := idx
	for u.parent[current] != current {
		current = u.parent[current]
	}
	root := current
	current = idx
	for u.parent[current] != root {
		next := u.parent[current]
		u.parent[current] = root
		current = next
	}
	return root
}

func (u *graphUnionFind) union(left int, right int) {
	leftRoot := u.find(left)
	rightRoot := u.find(right)
	if leftRoot != rightRoot {
		u.parent[rightRoot] = leftRoot
	}
}
