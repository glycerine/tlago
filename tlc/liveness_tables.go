package tlc

const nodePtrEmpty = int64(-1)

type NodePtrTable struct {
	count  int
	length int
	thresh int
	keys   []uint64
	elems  []int64
}

func NewNodePtrTable(size int) *NodePtrTable {
	if size <= 0 {
		size = 1
	}
	t := &NodePtrTable{
		length: size,
		thresh: int(float64(size) * 0.75),
		keys:   make([]uint64, size),
		elems:  make([]int64, size),
	}
	for i := range t.elems {
		t.elems[i] = nodePtrEmpty
	}
	return t
}

func (t *NodePtrTable) Put(k uint64, elem int64) {
	if t.count >= t.thresh {
		t.grow()
	}
	loc := livenessHashLoc(k, t.length)
	for {
		if t.elems[loc] == nodePtrEmpty {
			t.keys[loc] = k
			t.elems[loc] = elem
			t.count++
			return
		}
		if t.keys[loc] == k {
			t.elems[loc] = elem
			return
		}
		loc = (loc + 1) % t.length
	}
}

func (t *NodePtrTable) GetLoc(k uint64) int {
	if t.count >= t.thresh {
		t.grow()
	}
	loc := livenessHashLoc(k, t.length)
	for {
		if t.elems[loc] == nodePtrEmpty {
			return -1
		}
		if t.keys[loc] == k {
			return loc
		}
		loc = (loc + 1) % t.length
	}
}

func (t *NodePtrTable) Get(k uint64) int64 {
	loc := t.GetLoc(k)
	if loc == -1 {
		return nodePtrEmpty
	}
	return t.elems[loc]
}

func (t *NodePtrTable) GetByLoc(loc int) int64 {
	return t.elems[loc]
}

func (t *NodePtrTable) GetKeyByLoc(loc int) uint64 {
	return t.keys[loc]
}

func (t *NodePtrTable) PutByLoc(k uint64, elem int64, loc int) {
	t.keys[loc] = k
	if t.elems[loc] == nodePtrEmpty {
		t.count++
	}
	t.elems[loc] = elem
}

func (t *NodePtrTable) ResetElems() {
	for i := range t.elems {
		if t.elems[i] != nodePtrEmpty {
			t.elems[i] &= 0x7fffffffffffffff
		}
	}
}

func (t *NodePtrTable) Size() int {
	if t == nil {
		return 0
	}
	return t.count
}

func (t *NodePtrTable) GetSize() int {
	if t == nil {
		return 0
	}
	return t.length
}

func (t *NodePtrTable) grow() {
	oldKeys := t.keys
	oldElems := t.elems
	t.length = 2*t.length + 1
	t.thresh = int(float64(t.length) * 0.75)
	t.keys = make([]uint64, t.length)
	t.elems = make([]int64, t.length)
	for i := range t.elems {
		t.elems[i] = nodePtrEmpty
	}
	t.count = 0
	for i, elem := range oldElems {
		if elem != nodePtrEmpty {
			t.Put(oldKeys[i], elem)
		}
	}
}

type NodeTable struct {
	count  int
	length int
	thresh int
	elems  []any
	isBT   bool
}

func NewNodeTable(size int, isBT bool) *NodeTable {
	if size <= 0 {
		size = 1
	}
	return &NodeTable{
		length: size,
		thresh: size / 2,
		elems:  make([]any, size),
		isBT:   isBT,
	}
}

func (t *NodeTable) Size() int {
	if t == nil {
		return 0
	}
	return t.count
}

func (t *NodeTable) PutBENode(node *BEGraphNode) int {
	if t.count >= t.thresh {
		t.grow()
	}
	k := node.StateFP
	loc := livenessHashLoc(k, t.length)
	for {
		elem, _ := t.elems[loc].(*BEGraphNode)
		if elem == nil {
			t.elems[loc] = node
			t.count++
			return loc
		}
		if elem.StateFP == k {
			t.elems[loc] = node
			return loc
		}
		loc = (loc + 1) % t.length
	}
}

func (t *NodeTable) GetBENode(k uint64) *BEGraphNode {
	loc := livenessHashLoc(k, t.length)
	for {
		node, _ := t.elems[loc].(*BEGraphNode)
		if node == nil {
			return nil
		}
		if node.StateFP == k {
			return node
		}
		loc = (loc + 1) % t.length
	}
}

func (t *NodeTable) PutBTNode(node *BTGraphNode) int {
	if t.count >= t.thresh {
		t.grow()
	}
	k1 := node.StateFP
	k2 := node.GetIndex()
	loc := livenessHashLoc(k1, t.length)
	for {
		elem := t.elems[loc]
		switch e := elem.(type) {
		case nil:
			t.elems[loc] = node
			t.count++
			return loc
		case *BTGraphNode:
			if e.StateFP == k1 {
				if e.IsDummy() {
					t.elems[loc] = node
				} else if e.GetIndex() != k2 {
					t.elems[loc] = []*BTGraphNode{e, node}
				}
				return loc
			}
		case []*BTGraphNode:
			if len(e) > 0 && e[0].StateFP == k1 {
				for _, existing := range e {
					if existing.GetIndex() == k2 {
						return loc
					}
				}
				next := append(append([]*BTGraphNode(nil), e...), node)
				t.elems[loc] = next
				return loc
			}
		}
		loc = (loc + 1) % t.length
	}
}

func (t *NodeTable) GetBTNodes(k uint64) []*BTGraphNode {
	loc := livenessHashLoc(k, t.length)
	for {
		elem := t.elems[loc]
		switch e := elem.(type) {
		case nil:
			return nil
		case *BTGraphNode:
			if e.StateFP == k {
				if e.IsDummy() {
					return nil
				}
				return []*BTGraphNode{e}
			}
		case []*BTGraphNode:
			if len(e) > 0 && e[0].StateFP == k {
				return e
			}
		}
		loc = (loc + 1) % t.length
	}
}

func (t *NodeTable) GetBTNodesWithHint(k uint64, hintLoc int) []*BTGraphNode {
	if hintLoc >= 0 && hintLoc < len(t.elems) {
		switch e := t.elems[hintLoc].(type) {
		case *BTGraphNode:
			if e.StateFP == k {
				if e.IsDummy() {
					return nil
				}
				return []*BTGraphNode{e}
			}
		case []*BTGraphNode:
			if len(e) > 0 && e[0].StateFP == k {
				return e
			}
		}
	}
	return t.GetBTNodes(k)
}

func (t *NodeTable) GetBTNode(k1 uint64, k2 int) *BTGraphNode {
	loc := livenessHashLoc(k1, t.length)
	for {
		elem := t.elems[loc]
		switch e := elem.(type) {
		case nil:
			return nil
		case *BTGraphNode:
			if e.StateFP == k1 {
				if e.IsDummy() || e.GetIndex() != k2 {
					return nil
				}
				return e
			}
		case []*BTGraphNode:
			if len(e) > 0 && e[0].StateFP == k1 {
				for _, node := range e {
					if node.GetIndex() == k2 {
						return node
					}
				}
				return nil
			}
		}
		loc = (loc + 1) % t.length
	}
}

func (t *NodeTable) IsDone(loc int) bool {
	if loc < 0 || loc >= len(t.elems) {
		return false
	}
	switch e := t.elems[loc].(type) {
	case nil:
		return false
	case *BTGraphNode:
		return e.IsDone()
	case []*BTGraphNode:
		return len(e) > 0 && e[0].IsDone()
	default:
		return false
	}
}

func (t *NodeTable) SetDone(k uint64) {
	if t.count >= t.thresh {
		t.grow()
	}
	if !t.isBT {
		return
	}
	loc := livenessHashLoc(k, t.length)
	for {
		elem := t.elems[loc]
		switch e := elem.(type) {
		case nil:
			t.elems[loc] = NewDummyBTGraphNode(k)
			t.count++
			return
		case *BTGraphNode:
			if e.StateFP == k {
				e.SetDone()
				return
			}
		case []*BTGraphNode:
			if len(e) > 0 && e[0].StateFP == k {
				e[0].SetDone()
				return
			}
		}
		loc = (loc + 1) % t.length
	}
}

func (t *NodeTable) grow() {
	oldElems := t.elems
	t.count = 0
	t.length = 2*t.length + 1
	t.thresh = t.length / 2
	t.elems = make([]any, t.length)
	for _, elem := range oldElems {
		if elem == nil {
			continue
		}
		if t.isBT {
			t.putBTNodes(elem)
		} else if node, ok := elem.(*BEGraphNode); ok {
			t.PutBENode(node)
		}
	}
}

func (t *NodeTable) putBTNodes(nodes any) int {
	k := btNodesKey(nodes)
	loc := livenessHashLoc(k, t.length)
	for t.elems[loc] != nil {
		loc = (loc + 1) % t.length
	}
	t.elems[loc] = nodes
	t.count++
	return loc
}

func btNodesKey(nodes any) uint64 {
	switch n := nodes.(type) {
	case *BTGraphNode:
		return n.StateFP
	case []*BTGraphNode:
		if len(n) == 0 {
			return 0
		}
		return n[0].StateFP
	default:
		return 0
	}
}

func livenessHashLoc(k uint64, length int) int {
	return int(uint32(k)&0x7fffffff) % length
}
