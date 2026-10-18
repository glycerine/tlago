package tlc

import (
	"fmt"
	"strings"
	"sync/atomic"
)

const (
	nodePtrEmpty              = int64(-1)
	DiskGraphMaxPtr           = int64(0x4000000000000000)
	DiskGraphMaxLink          = int64(0x7fffffffffffffff)
	TableauNodePtrTableUndone = int64(-2) << 32
	TableauNodePtrTableDone   = int64(-3) << 32
	TableauEndMarker          = -1
	TableauNoParent           = -1
)

func IsDiskGraphFilePointer(loc int64) bool {
	return loc < DiskGraphMaxPtr
}

// Periodic liveness work reads size before suspending workers. Publish Java's
// primitive int count atomically while graph mutations retain their solution lock.
type NodePtrTable struct {
	count  int32
	length int
	thresh int
	keys   []uint64
	elems  []int64
}

func NewNodePtrTable(size int) *NodePtrTable {
	size = int(int32(size))
	if size < 0 {
		panic(NewNegativeArraySizeException(fmtInt(size)))
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

func (t *NodePtrTable) requireNodePtrTable() {
	if t == nil {
		panic(NewNullPointerException())
	}
}

func (t *NodePtrTable) Put(k uint64, elem int64) {
	t.requireNodePtrTable()
	if atomic.LoadInt32(&t.count) >= int32(t.thresh) {
		t.grow()
	}
	loc := livenessHashLoc(k, t.length)
	for {
		if t.elems[loc] == nodePtrEmpty {
			t.keys[loc] = k
			t.elems[loc] = elem
			atomic.AddInt32(&t.count, 1)
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
	t.requireNodePtrTable()
	if atomic.LoadInt32(&t.count) >= int32(t.thresh) {
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
	t.requireNodePtrTable()
	loc := t.GetLoc(k)
	if loc == -1 {
		return nodePtrEmpty
	}
	return t.elems[loc]
}

func (t *NodePtrTable) GetByLoc(loc int) int64 {
	t.requireNodePtrTable()
	loc = int(int32(loc))
	if loc < 0 || loc >= len(t.elems) {
		panic(NewArrayIndexOutOfBoundsException(loc, len(t.elems)))
	}
	return t.elems[loc]
}

func (t *NodePtrTable) GetKeyByLoc(loc int) uint64 {
	t.requireNodePtrTable()
	loc = int(int32(loc))
	if loc < 0 || loc >= len(t.keys) {
		panic(NewArrayIndexOutOfBoundsException(loc, len(t.keys)))
	}
	return t.keys[loc]
}

func (t *NodePtrTable) PutByLoc(k uint64, elem int64, loc int) {
	t.requireNodePtrTable()
	loc = int(int32(loc))
	if loc < 0 || loc >= len(t.keys) {
		panic(NewArrayIndexOutOfBoundsException(loc, len(t.keys)))
	}
	t.keys[loc] = k
	t.elems[loc] = elem
}

func (t *NodePtrTable) ResetElems() {
	t.requireNodePtrTable()
	for i := range t.elems {
		t.elems[i] &= 0x7fffffffffffffff
	}
}

func (t *NodePtrTable) Size() int {
	t.requireNodePtrTable()
	return int(atomic.LoadInt32(&t.count))
}

func (t *NodePtrTable) GetSize() int {
	t.requireNodePtrTable()
	return t.length
}

func (t *NodePtrTable) grow() {
	newLength := int(int32(2*int32(t.length) + 1))
	if newLength < 0 {
		panic(NewNegativeArraySizeException(fmtInt(newLength)))
	}
	oldKeys := t.keys
	oldElems := t.elems
	t.keys = make([]uint64, newLength)
	t.elems = make([]int64, newLength)
	for i := range t.elems {
		t.elems[i] = nodePtrEmpty
	}
	atomic.StoreInt32(&t.count, 0)
	for i, elem := range oldElems {
		if elem != nodePtrEmpty {
			loc := livenessHashLoc(oldKeys[i], newLength)
			for {
				if t.elems[loc] == nodePtrEmpty {
					t.keys[loc] = oldKeys[i]
					t.elems[loc] = elem
					atomic.AddInt32(&t.count, 1)
					break
				}
				if t.keys[loc] == oldKeys[i] {
					t.elems[loc] = elem
					break
				}
				loc = (loc + 1) % newLength
			}
		}
	}
	t.length = newLength
	t.thresh = int(float64(newLength) * 0.75)
}

type NodeTable struct {
	count  int
	length int
	thresh int
	elems  []any
	isBT   bool
}

func NewNodeTable(size int, isBT bool) *NodeTable {
	size = int(int32(size))
	if size < 0 {
		panic(NewNegativeArraySizeException(fmtInt(size)))
	}
	return &NodeTable{
		length: size,
		thresh: size / 2,
		elems:  make([]any, size),
		isBT:   isBT,
	}
}

func (t *NodeTable) requireNodeTable() {
	if t == nil {
		panic(NewNullPointerException())
	}
}

func (t *NodeTable) Size() int {
	t.requireNodeTable()
	return t.count
}

func (t *NodeTable) PutBENode(node *BEGraphNode) int {
	t.requireNodeTable()
	if t.count >= t.thresh {
		t.grow()
	}
	if node == nil {
		panic(NewNullPointerException())
	}
	k := node.StateFP
	loc := livenessHashLoc(k, t.length)
	for {
		elem := nodeTableBEGraphNode(t.elems[loc])
		if elem == nil {
			t.elems[loc] = nodeTableStoredBENode(node)
			t.count++
			return loc
		}
		if elem.StateFP == k {
			t.elems[loc] = nodeTableStoredBENode(node)
			return loc
		}
		loc = (loc + 1) % t.length
	}
}

func (t *NodeTable) GetBENode(k uint64) *BEGraphNode {
	t.requireNodeTable()
	loc := livenessHashLoc(k, t.length)
	for {
		node := nodeTableBEGraphNode(t.elems[loc])
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
	t.requireNodeTable()
	if t.count >= t.thresh {
		t.grow()
	}
	if node == nil {
		panic(NewNullPointerException())
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
		default:
			panic(valueStreamClassCast(elem, "[Ltlc2.tool.liveness.BTGraphNode;"))
		}
		loc = (loc + 1) % t.length
	}
}

func (t *NodeTable) GetBTNodes(k uint64) []*BTGraphNode {
	t.requireNodeTable()
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
		default:
			panic(valueStreamClassCast(elem, "[Ltlc2.tool.liveness.BTGraphNode;"))
		}
		loc = (loc + 1) % t.length
	}
}

func (t *NodeTable) GetBTNodeWithHint(k uint64, hintLoc int) []*BTGraphNode {
	t.requireNodeTable()
	hintLoc = int(int32(hintLoc))
	if hintLoc < 0 || hintLoc >= len(t.elems) {
		panic(NewArrayIndexOutOfBoundsException(hintLoc, len(t.elems)))
	}
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
	case nil:
	default:
		panic(valueStreamClassCast(t.elems[hintLoc], "[Ltlc2.tool.liveness.BTGraphNode;"))
	}
	return t.GetBTNodes(k)
}

func (t *NodeTable) GetBTNodesWithHint(k uint64, hintLoc int) []*BTGraphNode {
	return t.GetBTNodeWithHint(k, hintLoc)
}

func (t *NodeTable) GetBTNode(k1 uint64, k2 int) *BTGraphNode {
	t.requireNodeTable()
	k2 = int(int32(k2))
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
		default:
			panic(valueStreamClassCast(elem, "[Ltlc2.tool.liveness.BTGraphNode;"))
		}
		loc = (loc + 1) % t.length
	}
}

func (t *NodeTable) IsDone(loc int) bool {
	t.requireNodeTable()
	loc = int(int32(loc))
	if loc < 0 || loc >= len(t.elems) {
		panic(NewArrayIndexOutOfBoundsException(loc, len(t.elems)))
	}
	switch e := t.elems[loc].(type) {
	case nil:
		return false
	case *BTGraphNode:
		return e.IsDone()
	case []*BTGraphNode:
		return len(e) > 0 && e[0].IsDone()
	default:
		panic(valueStreamClassCast(t.elems[loc], "[Ltlc2.tool.liveness.BTGraphNode;"))
	}
}

func (t *NodeTable) SetDone(k uint64) {
	t.requireNodeTable()
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
		default:
			panic(valueStreamClassCast(elem, "[Ltlc2.tool.liveness.BTGraphNode;"))
		}
		loc = (loc + 1) % t.length
	}
}

func (t *NodeTable) grow() {
	oldElems := t.elems
	t.count = 0
	t.length = int(int32(2*int32(t.length) + 1))
	t.thresh = t.length / 2
	if t.length < 0 {
		panic(NewNegativeArraySizeException(fmtInt(t.length)))
	}
	t.elems = make([]any, t.length)
	for _, elem := range oldElems {
		if elem == nil {
			continue
		}
		if t.isBT {
			t.putBTNodes(elem)
		} else {
			t.PutBENode(nodeTableBEGraphNode(elem))
		}
	}
}

func nodeTableBEGraphNode(elem any) *BEGraphNode {
	switch node := elem.(type) {
	case nil:
		return nil
	case *BEGraphNode:
		return node
	case *BTGraphNode:
		if node == nil {
			return nil
		}
		return node.BEGraphNode
	default:
		panic(valueStreamClassCast(elem, "tlc2.tool.liveness.BEGraphNode"))
	}
}

func nodeTableStoredBENode(node *BEGraphNode) any {
	if node != nil && node.ownedTableau() {
		return newBTGraphNodeView(node)
	}
	return node
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
		panic(valueStreamClassCast(nodes, "[Ltlc2.tool.liveness.BTGraphNode;"))
	}
}

func livenessHashLoc(k uint64, length int) int {
	if length == 0 {
		panic(NewArithmeticException("/ by zero"))
	}
	return int(uint32(k)&0x7fffffff) % length
}

type TableauNodePtrTable struct {
	count   int32
	length  int
	thresh  int
	nodes   [][]int32
	reverse bool
}

func NewTableauNodePtrTable(size int) *TableauNodePtrTable {
	return newTableauNodePtrTable(size, false)
}

func NewReverseTableauNodePtrTable(size int) *TableauNodePtrTable {
	return newTableauNodePtrTable(size, true)
}

func newTableauNodePtrTable(size int, reverse bool) *TableauNodePtrTable {
	size = int(int32(size))
	if size < 0 {
		panic(NewNegativeArraySizeException(fmtInt(size)))
	}
	return &TableauNodePtrTable{
		length:  size,
		thresh:  int(float64(size) * 0.75),
		nodes:   make([][]int32, size),
		reverse: reverse,
	}
}

func (t *TableauNodePtrTable) requireTableauNodePtrTable() {
	if t == nil {
		panic(NewNullPointerException())
	}
}

func requireTableauIndex(tidx int) int {
	tidx = int(int32(tidx))
	if tidx < -1 {
		panic(NewAssertionError())
	}
	return tidx
}

func (t *TableauNodePtrTable) Size() int {
	t.requireTableauNodePtrTable()
	return int(atomic.LoadInt32(&t.count))
}

func (t *TableauNodePtrTable) GetSize() int {
	t.requireTableauNodePtrTable()
	return t.length
}

func (t *TableauNodePtrTable) Get(k uint64, tidx int) int64 {
	t.requireTableauNodePtrTable()
	tidx = requireTableauIndex(tidx)
	if atomic.LoadInt32(&t.count) >= int32(t.thresh) {
		t.grow()
	}
	loc := livenessHashLoc(k, t.length)
	for {
		node := t.nodes[loc]
		if node == nil {
			return -1
		}
		if TableauGetKey(node) == k {
			idx := t.GetIdx(node, tidx)
			if idx == -1 {
				return -1
			}
			return TableauGetElem(node, idx)
		}
		loc = (loc + 1) % t.length
	}
}

func (t *TableauNodePtrTable) Put(k uint64, tidx int) {
	t.put(k, tidx, TableauNodePtrTableUndone, TableauNodePtrTableDone)
}

func (t *TableauNodePtrTable) PutElem(k uint64, tidx int, elem int64) {
	t.put(k, tidx, elem, elem)
}

func (t *TableauNodePtrTable) put(k uint64, tidx int, addElem int64, newElem int64) {
	t.requireTableauNodePtrTable()
	tidx = requireTableauIndex(tidx)
	if atomic.LoadInt32(&t.count) >= int32(t.thresh) {
		t.grow()
	}
	loc := livenessHashLoc(k, t.length)
	for {
		node := t.nodes[loc]
		if node == nil {
			t.nodes[loc] = t.addElem(k, tidx, addElem)
			atomic.AddInt32(&t.count, 1)
			return
		}
		if TableauGetKey(node) == k {
			cloc := t.GetIdx(node, tidx)
			if cloc == -1 {
				t.nodes[loc] = t.appendElem(node, tidx, newElem)
			} else {
				TableauPutElem(t.nodes[loc], addElem, cloc)
			}
			return
		}
		loc = (loc + 1) % t.length
	}
}

func (t *TableauNodePtrTable) GetLoc(k uint64, tidx int) int {
	t.requireTableauNodePtrTable()
	tidx = requireTableauIndex(tidx)
	if atomic.LoadInt32(&t.count) >= int32(t.thresh) {
		t.grow()
	}
	loc := livenessHashLoc(k, t.length)
	for {
		node := t.nodes[loc]
		if node == nil {
			return -1
		}
		if TableauGetKey(node) == k {
			if t.GetIdx(node, tidx) == -1 {
				return -1
			}
			return loc
		}
		loc = (loc + 1) % t.length
	}
}

func (t *TableauNodePtrTable) GetNodes(k uint64) []int32 {
	t.requireTableauNodePtrTable()
	if atomic.LoadInt32(&t.count) >= int32(t.thresh) {
		t.grow()
	}
	loc := livenessHashLoc(k, t.length)
	for {
		node := t.nodes[loc]
		if node == nil {
			return nil
		}
		if TableauGetKey(node) == k {
			return node
		}
		loc = (loc + 1) % t.length
	}
}

func (t *TableauNodePtrTable) GetNodesLoc(k uint64) int {
	t.requireTableauNodePtrTable()
	if atomic.LoadInt32(&t.count) >= int32(t.thresh) {
		t.grow()
	}
	loc := livenessHashLoc(k, t.length)
	for {
		node := t.nodes[loc]
		if node == nil {
			return -1
		}
		if TableauGetKey(node) == k {
			return loc
		}
		loc = (loc + 1) % t.length
	}
}

func (t *TableauNodePtrTable) GetNodesByLoc(loc int) []int32 {
	t.requireTableauNodePtrTable()
	loc = int(int32(loc))
	if loc < 0 || loc >= len(t.nodes) {
		panic(NewArrayIndexOutOfBoundsException(loc, len(t.nodes)))
	}
	return t.nodes[loc]
}

func (t *TableauNodePtrTable) IsDone(k uint64) bool {
	t.requireTableauNodePtrTable()
	node := t.GetNodes(k)
	if node == nil {
		return false
	}
	if len(node) == 2 {
		return true
	}
	return node[3] != -2
}

func (t *TableauNodePtrTable) SetDone(k uint64) int {
	t.requireTableauNodePtrTable()
	if atomic.LoadInt32(&t.count) >= int32(t.thresh) {
		t.grow()
	}
	loc := livenessHashLoc(k, t.length)
	for {
		node := t.nodes[loc]
		if node == nil {
			t.nodes[loc] = tableauAddKey(k)
			atomic.AddInt32(&t.count, 1)
			return loc
		}
		if TableauGetKey(node) == k {
			if len(node) > 2 && node[3] == -2 {
				node[3] = -3
			}
			return loc
		}
		loc = (loc + 1) % t.length
	}
}

func (t *TableauNodePtrTable) ResetElems() {
	t.requireTableauNodePtrTable()
	for _, node := range t.nodes {
		if node == nil {
			continue
		}
		for j := 3; j < len(node); j += t.GetElemLength() {
			node[j] &= 0x7fffffff
		}
	}
}

func (t *TableauNodePtrTable) GetElemLength() int {
	t.requireTableauNodePtrTable()
	if t.reverse {
		return 4
	}
	return 3
}

func (t *TableauNodePtrTable) GetIdx(node []int32, tidx int) int {
	t.requireTableauNodePtrTable()
	tidx = requireTableauIndex(tidx)
	if node == nil {
		panic(NewNullPointerException())
	}
	for i := 2; i < len(node); i += t.GetElemLength() {
		if int(node[i]) == tidx {
			return i
		}
	}
	return -1
}

func (t *TableauNodePtrTable) GetElemTidx(node []int32, loc int) int {
	t.requireTableauNodePtrTable()
	if t.reverse {
		return int(tableauArrayGet(node, tableauOffset(loc, 3)))
	}
	return -1
}

func (t *TableauNodePtrTable) PutRecordElem(node []int32, elem int64, tableauIdx int, loc int) {
	t.requireTableauNodePtrTable()
	tableauIdx = requireTableauIndex(tableauIdx)
	TableauPutElem(node, elem, loc)
	if t.reverse {
		tableauArraySet(node, tableauOffset(loc, 3), int32(tableauIdx))
	}
}

func TableauNodePtrTableDetailedFormatter(table *TableauNodePtrTable) string {
	if table == nil {
		panic(NewNullPointerException())
	}
	var out strings.Builder
	out.Grow(table.Size())
	for index, node := range table.nodes {
		if node == nil {
			continue
		}
		out.WriteString(fmt.Sprintf("fp (key): %d (idx: %d) isDone: %t\n", int64(TableauGetKey(node)), index, len(node) == 2 || len(node) > 2 && node[3] != -2))
		for loc := 2; loc < len(node)-1; loc += table.GetElemLength() {
			out.WriteString(fmt.Sprintf("\t tidx: %d", TableauGetTidx(node, loc)))
			elem := TableauGetElem(node, loc)
			switch {
			case IsDiskGraphFilePointer(elem):
				switch elem {
				case TableauNodePtrTableUndone:
					out.WriteString("  ptr: undone")
				case TableauNodePtrTableDone:
					out.WriteString("  ptr: done")
				default:
					out.WriteString(fmt.Sprintf("  ptr: %d", elem))
				}
			case elem == DiskGraphMaxPtr:
				out.WriteString(" elem: Init State")
			default:
				out.WriteString(fmt.Sprintf(" pred: %d", elem-tableauDiskGraphInitState))
			}
			if predTidx := table.GetElemTidx(node, loc); predTidx != -1 {
				out.WriteString(fmt.Sprintf(" predtidx: %d", predTidx))
			}
			out.WriteString(fmt.Sprintf(", isSeen: %t\n", !TableauIsSeenAt(node, loc)))
		}
	}
	return out.String()
}

func (t *TableauNodePtrTable) ToDotViz() string {
	t.requireTableauNodePtrTable()
	out := "subgraph cluster_table {graph[style=bold];label = \"NodePtrTable\" style=\"solid\"\n"
	out += "node [ labeljust=\"l\",shape=record ]\n"
	out += "key [label=<<table border=\"1\" cellpadding=\"2\" cellspacing=\"0\" cellborder=\"1\">\n"
	out += "<tr> <td BGCOLOR=\"lightblue\">fp</td> <td BGCOLOR=\"lightblue\">tid</td> <td BGCOLOR=\"lightblue\">idx</td> <td BGCOLOR=\"lightblue\">isDone</td> <td BGCOLOR=\"lightblue\">isSeen</td> <td BGCOLOR=\"lightblue\">ptr</td> <td BGCOLOR=\"lightblue\">pred tid</td> </tr>\n"
	for i, node := range t.nodes {
		if node == nil {
			continue
		}
		fp := TableauGetKey(node)
		if 2 == len(node)-1 {
			fpText := tableauDotFingerprint(fp)
			out += fmt.Sprintf("<tr> <td>%s</td> <td>%s</td> <td>%d</td> <td>%v</td> <td>%s</td> <td>%s</td> <td>%s</td> </tr>\n", fpText, "NA", i, t.IsDone(fp), "NA", "NA", "NA")
			continue
		}
		for j := 2; j < len(node)-1; j += t.GetElemLength() {
			fpText := tableauDotFingerprint(fp)
			tidx := TableauGetTidx(node, j)
			elem := TableauGetElem(node, j)
			ptr := fmt.Sprintf("%d", elem)
			switch {
			case IsDiskGraphFilePointer(elem):
				if elem == TableauNodePtrTableUndone {
					ptr = "undone"
				} else if elem == TableauNodePtrTableDone {
					ptr = "done"
				}
			case elem == DiskGraphMaxPtr:
				ptr = "Initial"
			default:
				ptr = fmt.Sprintf("%d", elem-tableauDiskGraphInitState)
			}
			predTidx := "-"
			if tidx := t.GetElemTidx(node, j); tidx != -1 {
				predTidx = fmt.Sprintf("%d", tidx)
			}
			out += fmt.Sprintf("<tr> <td>%s</td> <td>%d</td> <td>%d</td> <td>%v</td> <td>%v</td> <td>%s</td> <td>%s</td> </tr>\n", fpText, tidx, i, t.IsDone(fp), !TableauIsSeenAt(node, j), ptr, predTidx)
		}
	}
	out += "</table>>]\n}"
	return out
}

func (t *TableauNodePtrTable) grow() {
	t.length = int(int32(2*int32(t.length) + 1))
	t.thresh = int(float64(t.length) * 0.75)
	oldNodes := t.nodes
	if t.length < 0 {
		panic(NewNegativeArraySizeException(fmtInt(t.length)))
	}
	t.nodes = make([][]int32, t.length)
	for _, node := range oldNodes {
		if node != nil {
			t.putNode(node)
		}
	}
}

func (t *TableauNodePtrTable) putNode(node []int32) {
	k := TableauGetKey(node)
	loc := livenessHashLoc(k, t.length)
	for {
		if t.nodes[loc] == nil {
			t.nodes[loc] = node
			return
		}
		loc = (loc + 1) % t.length
	}
}

func (t *TableauNodePtrTable) addElem(key uint64, tidx int, elem int64) []int32 {
	t.requireTableauNodePtrTable()
	tidx = requireTableauIndex(tidx)
	node := make([]int32, 3+t.GetElemLength()-1)
	node[0] = int32(key >> 32)
	node[1] = int32(uint32(key))
	node[2] = int32(tidx)
	node[3] = int32(elem >> 32)
	node[4] = int32(uint32(elem))
	if t.reverse {
		node[5] = -1
	}
	return node
}

func (t *TableauNodePtrTable) appendElem(node []int32, tidx int, elem int64) []int32 {
	t.requireTableauNodePtrTable()
	tidx = requireTableauIndex(tidx)
	if node == nil {
		panic(NewNullPointerException())
	}
	oldLen := len(node)
	next := make([]int32, oldLen+t.GetElemLength())
	copy(next, node)
	next[oldLen] = int32(tidx)
	next[oldLen+1] = int32(elem >> 32)
	next[oldLen+2] = int32(uint32(elem))
	if t.reverse {
		next[oldLen+3] = -1
	}
	return next
}

func tableauAddKey(key uint64) []int32 {
	return []int32{int32(key >> 32), int32(uint32(key))}
}

func TableauGetKey(node []int32) uint64 {
	high := uint64(uint32(tableauArrayGet(node, 0)))
	low := uint64(uint32(tableauArrayGet(node, 1)))
	return (high << 32) | low
}

func TableauGetElem(node []int32, loc int) int64 {
	high := int64(tableauArrayGet(node, tableauOffset(loc, 1)))
	low := int64(uint32(tableauArrayGet(node, tableauOffset(loc, 2))))
	return (high << 32) | low
}

func TableauPutElem(node []int32, elem int64, loc int) {
	tableauArraySet(node, tableauOffset(loc, 1), int32(elem>>32))
	tableauArraySet(node, tableauOffset(loc, 2), int32(uint32(elem)))
}

func TableauGetTidx(node []int32, loc int) int {
	loc = int(int32(loc))
	return int(tableauArrayGet(node, loc))
}

func TableauStartLoc(node []int32) int {
	if node == nil {
		panic(NewNullPointerException())
	}
	if len(node) > 2 {
		return 2
	}
	return TableauEndMarker
}

func TableauNextLoc(node []int32, curLoc int) int {
	if node == nil {
		panic(NewNullPointerException())
	}
	loc := tableauOffset(curLoc, 3)
	if loc < len(node) {
		return loc
	}
	return TableauEndMarker
}

func TableauIsSeenAt(node []int32, tloc int) bool {
	return TableauGetElem(node, tloc) < 0
}

func TableauSetSeenAt(node []int32, tloc int) {
	ptr := TableauGetElem(node, tloc)
	TableauPutElem(node, ptr|int64(-1<<63), tloc)
}

func TableauGetPtr(ptr int64) int64 {
	return ptr & 0x7fffffffffffffff
}

func TableauIsSeen(node []int32) bool {
	return tableauArrayGet(node, 3) < 0
}

func TableauSetSeen(node []int32) {
	tableauArraySet(node, 3, int32(uint32(tableauArrayGet(node, 3))|0x80000000))
}

func TableauGetParent(node []int32) int {
	return int(tableauArrayGet(node, 4))
}

func TableauSetParent(node []int32, loc int) {
	loc = int(int32(loc))
	if loc < TableauNoParent || int64(loc) > DiskGraphMaxPtr {
		panic(NewAssertionError())
	}
	tableauArraySet(node, 4, int32(loc))
}

func tableauDotFingerprint(fp uint64) string {
	text := fmt.Sprintf("%d", int64(fp))
	if len(text) < 6 {
		panic(NewStringIndexOutOfBoundsException(6, len(text)))
	}
	return text[:6]
}

func tableauOffset(index, delta int) int {
	return int(int32(index) + int32(delta))
}

func tableauArrayGet(node []int32, index int) int32 {
	if node == nil {
		panic(NewNullPointerException())
	}
	if index < 0 || index >= len(node) {
		panic(NewArrayIndexOutOfBoundsException(index, len(node)))
	}
	return node[index]
}

func tableauArraySet(node []int32, index int, value int32) {
	if node == nil {
		panic(NewNullPointerException())
	}
	if index < 0 || index >= len(node) {
		panic(NewArrayIndexOutOfBoundsException(index, len(node)))
	}
	node[index] = value
}
