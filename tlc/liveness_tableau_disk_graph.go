package tlc

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const tableauDiskGraphInitState = DiskGraphMaxPtr + 1

type TableauDiskGraph struct {
	*DiskGraph
	TableauNodePtrTbl *TableauNodePtrTable
	DebugEnabled      bool
	DebugOOS          *OrderOfSolution
	DebugCounter      int

	ownersCaptured           bool
	tableauNodePtrTableOwner *TableauNodePtrTable
	debugEnabledOwner        bool
	debugOOSOwner            *OrderOfSolution
	debugCounterOwner        int
}

func (g *TableauDiskGraph) requireTableauDiskGraph() {
	if g == nil || g.DiskGraph == nil {
		panic(NewNullPointerException())
	}
}

func NewTableauDiskGraph(metadir string, soln int, outDegreeStats ...any) (*TableauDiskGraph, error) {
	if len(outDegreeStats) > 1 {
		panic(NewIllegalArgumentException("NewTableauDiskGraph accepts at most one out-degree statistics object"))
	}
	base, err := NewDiskGraph(metadir, soln, outDegreeStats...)
	if err != nil {
		return nil, err
	}
	table := NewTableauNodePtrTable(255)
	return &TableauDiskGraph{
		DiskGraph: base, TableauNodePtrTbl: table,
		ownersCaptured: true, tableauNodePtrTableOwner: table,
	}, nil
}

func NewDebugTableauDiskGraph(metadir string, soln int, outDegreeStats any, oos *OrderOfSolution) (*TableauDiskGraph, error) {
	graph, err := NewTableauDiskGraph(metadir, soln, outDegreeStats)
	if err != nil {
		return nil, err
	}
	graph.DebugEnabled = true
	graph.DebugOOS = oos
	graph.debugEnabledOwner = true
	graph.debugOOSOwner = oos
	return graph, nil
}

func (g *TableauDiskGraph) ownedTableauNodePtrTable() *TableauNodePtrTable {
	if g.ownersCaptured {
		return g.tableauNodePtrTableOwner
	}
	return g.TableauNodePtrTbl
}

func (g *TableauDiskGraph) setTableauNodePtrTable(table *TableauNodePtrTable) {
	g.TableauNodePtrTbl = table
	if g.ownersCaptured {
		g.tableauNodePtrTableOwner = table
	}
}

func (g *TableauDiskGraph) ownedDebugEnabled() bool {
	if g.ownersCaptured {
		return g.debugEnabledOwner
	}
	return g.DebugEnabled
}

func (g *TableauDiskGraph) ownedDebugOOS() *OrderOfSolution {
	if g.ownersCaptured {
		return g.debugOOSOwner
	}
	return g.DebugOOS
}

func (g *TableauDiskGraph) takeDebugCounter() int {
	if !g.ownersCaptured {
		counter := g.DebugCounter
		g.DebugCounter++
		return counter
	}
	counter := g.debugCounterOwner
	g.debugCounterOwner++
	g.DebugCounter = g.debugCounterOwner
	return counter
}

func (g *TableauDiskGraph) GetPtr(fp uint64, tidx int) int64 {
	g.requireTableauDiskGraph()
	tidx = graphInt(tidx)
	if tidx < 0 {
		panic(NewAssertionError())
	}
	return g.ownedTableauNodePtrTable().Get(fp, tidx)
}

func (g *TableauDiskGraph) GetElemLength() int {
	g.requireTableauDiskGraph()
	return g.ownedTableauNodePtrTable().GetElemLength()
}

func (g *TableauDiskGraph) IsDone(fp uint64) bool {
	g.requireTableauDiskGraph()
	return g.ownedTableauNodePtrTable().IsDone(fp)
}

func (g *TableauDiskGraph) SetDone(fp uint64) int {
	g.requireTableauDiskGraph()
	defer g.debugWriteDotForFP(fp)
	return g.ownedTableauNodePtrTable().SetDone(fp)
}

func (g *TableauDiskGraph) RecordNode(fp uint64, tidx int) {
	g.requireTableauDiskGraph()
	tidx = graphInt(tidx)
	if tidx < 0 {
		panic(NewAssertionError())
	}
	defer g.debugWriteDotForFP(fp)
	g.ownedTableauNodePtrTable().Put(fp, tidx)
}

func (g *TableauDiskGraph) GetNodesByLoc(loc int) []int32 {
	g.requireTableauDiskGraph()
	return g.ownedTableauNodePtrTable().GetNodesByLoc(loc)
}

func (g *TableauDiskGraph) AddNode(node *GraphNode) (int64, error) {
	g.requireTableauDiskGraph()
	defer func() {
		g.debugWriteDotForNode(node)
	}()
	if node == nil {
		panic(NewNullPointerException())
	}
	outDegree := node.SuccSize()
	if isNil(g.ownedOutDegreeStats()) {
		panic(NewNullPointerException())
	}
	bucketStatisticsAddSample(g.ownedOutDegreeStats(), outDegree)
	ptr, err := g.nodeFile.Seek(0, io.SeekCurrent)
	if err != nil {
		return -1, err
	}
	g.putTableauNode(node, ptr)
	ptrOut := g.ptrFile
	if err := ptrOut.WriteLong(int64(node.ownedStateFP())); err != nil {
		return -1, err
	}
	if err := ptrOut.WriteInt(int32(node.ownedTIndex())); err != nil {
		return -1, err
	}
	if err := ptrOut.WriteLongNat(ptr); err != nil {
		return -1, err
	}
	nodeOut := g.nodeFile
	if err := node.Write(nodeOut); err != nil {
		return -1, err
	}
	return ptr, nil
}

func (g *TableauDiskGraph) debugWriteDotForFP(fp uint64) {
	g.requireTableauDiskGraph()
	if !g.ownedDebugEnabled() {
		return
	}
	counter := g.takeDebugCounter()
	g.debugWriteDotForFPAndCounter(fp, counter)
}

func (g *TableauDiskGraph) debugWriteDotForNode(node *GraphNode) {
	g.requireTableauDiskGraph()
	if !g.ownedDebugEnabled() {
		return
	}
	counter := g.takeDebugCounter()
	if node == nil {
		panic(NewNullPointerException())
	}
	g.debugWriteDotForFPAndCounter(node.ownedStateFP(), counter)
}

func (g *TableauDiskGraph) debugWriteDotForFPAndCounter(fp uint64, counter int) {
	fragment := strconv.FormatInt(int64(fp), 10)
	if len(fragment) < 6 {
		panic(NewStringIndexOutOfBoundsException(6, len(fragment)))
	}
	fragment = fragment[:6]
	path := filepath.Join(g.ownedMetaDir(), fmt.Sprintf("dgraph_%03d_%s.dot", counter, fragment))

	g.WriteDotViz(g.ownedDebugOOS(), path)
}

func (g *TableauDiskGraph) putTableauNode(node *GraphNode, ptr int64) {
	g.ownedTableauNodePtrTable().PutElem(node.ownedStateFP(), node.ownedTIndex(), ptr)
}

func (g *TableauDiskGraph) CheckDuplicate(node *GraphNode) bool {
	g.requireTableauDiskGraph()
	if node == nil {
		panic(NewNullPointerException())
	}
	return g.ownedTableauNodePtrTable().Get(node.ownedStateFP(), node.ownedTIndex()) != -1
}

func (g *TableauDiskGraph) GetNode(fp uint64, tidx int) (*GraphNode, error) {
	g.requireTableauDiskGraph()
	tidx = graphInt(tidx)
	if tidx < 0 {
		panic(NewAssertionError())
	}
	ptr := g.ownedTableauNodePtrTable().Get(fp, tidx)
	if ptr < 0 {
		return NewGraphNode(fp, tidx), nil
	}
	if g.ownedCache() == nil {
		return g.getNodeFromDisk(fp, tidx, ptr)
	}
	return g.getNode(fp, tidx, ptr)
}

func (g *TableauDiskGraph) GetLink(state uint64, tidx int) int64 {
	g.requireTableauDiskGraph()
	tidx = graphInt(tidx)
	if tidx < 0 {
		panic(NewAssertionError())
	}
	return g.ownedTableauNodePtrTable().Get(state, tidx)
}

func (g *TableauDiskGraph) PutLink(state uint64, tidx int, link int64) int64 {
	g.requireTableauDiskGraph()
	tidx = graphInt(tidx)
	if tidx < 0 || link < DiskGraphMaxPtr || link >= DiskGraphMaxLink {
		panic(NewAssertionError())
	}
	node := g.ownedTableauNodePtrTable().GetNodes(state)
	cloc := g.ownedTableauNodePtrTable().GetIdx(node, tidx)
	oldLink := TableauGetElem(node, cloc)
	if !IsDiskGraphFilePointer(oldLink) {
		return oldLink
	}
	TableauPutElem(node, link, cloc)
	return -1
}

func (g *TableauDiskGraph) SetMaxLink(state uint64, tidx int) {
	g.requireTableauDiskGraph()
	tidx = graphInt(tidx)
	if tidx < 0 {
		panic(NewAssertionError())
	}
	g.ownedTableauNodePtrTable().PutElem(state, tidx, DiskGraphMaxLink)
}

func (g *TableauDiskGraph) Reset() error {
	g.requireTableauDiskGraph()
	if err := g.ptrFile.SetLength(0); err != nil {
		return err
	}
	if err := g.nodeFile.SetLength(0); err != nil {
		return err
	}
	g.setTableauNodePtrTable(NewTableauNodePtrTable(255))
	return nil
}

func (g *TableauDiskGraph) Recover() error {
	g.requireTableauDiskGraph()
	file, err := openFileInputStream(g.chkptName + ".chkpt")
	if err != nil {
		return err
	}
	// Source checkpoints use DataInputStream directly on FileInputStream.
	var nodePos, ptrPos int64
	if err := binary.Read(file, binary.BigEndian, &nodePos); err != nil {
		return bufferedRandomAccessFileIOError(valueStreamReadError(err))
	}
	if err := binary.Read(file, binary.BigEndian, &ptrPos); err != nil {
		return bufferedRandomAccessFileIOError(valueStreamReadError(err))
	}
	if err := file.Close(); err != nil {
		return bufferedRandomAccessFileIOError(err)
	}
	if err := g.MakeNodePtrTblTo(ptrPos); err != nil {
		return err
	}
	if _, err := g.nodeFile.Seek(nodePos, io.SeekStart); err != nil {
		return err
	}
	_, err = g.ptrFile.Seek(ptrPos, io.SeekStart)
	return err
}

func (g *TableauDiskGraph) Size() int {
	g.requireTableauDiskGraph()
	return g.ownedTableauNodePtrTable().Size()
}

func (g *TableauDiskGraph) CheckInvariants(slen int, alen int) (bool, error) {
	g.requireTableauDiskGraph()
	stop := &stopDiskGraphIteration{}
	err := g.eachTableauGraphNodeWithRestore(func(node *GraphNode) error {
		if !node.CheckInvariants(slen, alen) {
			return stop
		}
		return nil
	}, false)
	if err == stop {
		return false, nil
	}
	return err == nil, err
}

func (g *TableauDiskGraph) MakeNodePtrTbl() error {
	g.requireTableauDiskGraph()
	ptr, err := g.ptrFile.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	end, err := g.ptrFile.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}
	if err := g.MakeNodePtrTblTo(end); err != nil {
		return err
	}
	_, err = g.ptrFile.Seek(ptr, io.SeekStart)
	return err
}

func (g *TableauDiskGraph) MakeNodePtrTblTo(ptr int64) error {
	g.requireTableauDiskGraph()
	return g.MakeNodePtrTblToTable(ptr, g.ownedTableauNodePtrTable())
}

func (g *TableauDiskGraph) MakeNodePtrTblToTable(ptr int64, table *TableauNodePtrTable) error {
	g.requireTableauDiskGraph()
	if _, err := g.ptrFile.Seek(0, io.SeekStart); err != nil {
		return err
	}
	in := g.ptrFile
	for {
		cur, err := g.ptrFile.Seek(0, io.SeekCurrent)
		if err != nil {
			return err
		}
		if cur >= ptr {
			return nil
		}
		fp, err := in.ReadLong()
		if err != nil {
			return err
		}
		tidx, err := in.ReadInt()
		if err != nil {
			return err
		}
		loc, err := in.ReadLongNat()
		if err != nil {
			return err
		}
		table.PutElem(uint64(fp), int(tidx), loc)
	}
}

func (g *TableauDiskGraph) String() string {
	g.requireTableauDiskGraph()
	if g.ownedCache() == nil {
		return ""
	}
	var b strings.Builder
	err := g.eachTableauGraphNode(func(node *GraphNode) error {
		b.WriteString(fmt.Sprintf("<%s,%d> -> ", livenessFPString(node.ownedStateFP()), node.ownedTIndex()))
		for i := 0; i < node.SuccSize(); i++ {
			b.WriteString(fmt.Sprintf("<%s,%d> ", livenessFPString(node.GetStateFP(i)), node.GetTIndex(i)))
		}
		b.WriteByte('\n')
		return nil
	})
	if err != nil {
		exitLivenessDiskGraphIO(err)
	}
	return b.String()
}

func (g *TableauDiskGraph) ToDotViz(oos *OrderOfSolution, labels map[uint64]string) string {
	g.requireTableauDiskGraph()
	checkState := oos.GetCheckState()
	if checkState == nil {
		panic(NewNullPointerException())
	}
	checkAction := oos.GetCheckAction()
	if checkAction == nil {
		panic(NewNullPointerException())
	}
	slen := len(checkState)
	alen := len(checkAction)
	if g.ownedCache() == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString("digraph DiskGraph {\n")
	b.WriteString("nodesep = 0.7\n")
	b.WriteString("rankdir=LR;\n")
	b.WriteString(diskGraphDotVizLegend(oos))
	b.WriteString(g.ownedTableauNodePtrTable().ToDotViz())
	b.WriteString("subgraph cluster_graph {\n")
	b.WriteString("color=\"white\";\n")
	err := g.eachTableauGraphNode(func(node *GraphNode) error {
		b.WriteString(node.ToDotViz(g.isInitState(node), true, slen, alen, oos, labels))
		return nil
	})
	if err != nil {
		exitLivenessDiskGraphIO(err)
	}
	b.WriteString("}}")
	return b.String()
}

func (g *TableauDiskGraph) WriteDotViz(oos *OrderOfSolution, path string, labelSets ...map[uint64]string) {
	g.requireTableauDiskGraph()
	if len(labelSets) > 1 {
		panic(NewIllegalArgumentException("WriteDotViz accepts at most one label map"))
	}
	labels := map[uint64]string{}
	if len(labelSets) > 0 {
		labels = labelSets[0]
	}
	g.CreateCache()
	if err := writeLivenessDotFile(path, func() string { return g.ToDotViz(oos, labels) }); err != nil {
		_, _ = fmt.Fprint(os.Stderr, javaThrowableStackTrace(err))
	}
	g.DestroyCache()
}

func (g *TableauDiskGraph) GetPath(state uint64, tidx int) (*LongVec, error) {
	g.requireTableauDiskGraph()
	tidx = graphInt(tidx)
	numOfInits := g.ownedInitNodes().Size()
	for i := 0; i < numOfInits; i += 2 {
		state0 := uint64(g.ownedInitNodes().ElementAt(i))
		tidx0 := int(g.ownedInitNodes().ElementAt(i + 1))
		if state0 == state && tidx0 == tidx {
			res := NewLongVecWithCapacity(1)
			res.AddElement(int64(state0))
			return res, nil
		}
	}

	ptrEnd, err := g.ptrFile.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}
	reverseTable := NewReverseTableauNodePtrTable(255)
	if err := g.MakeNodePtrTblToTable(ptrEnd, reverseTable); err != nil {
		return nil, err
	}

	queue := NewIntQueueWithDisk(g.ownedMetaDir(), "", intQueueInitialSize)
	for i := 0; i < numOfInits; i += 2 {
		state0 := uint64(g.ownedInitNodes().ElementAt(i))
		tidx0 := int(g.ownedInitNodes().ElementAt(i + 1))
		ptr := reverseTable.Get(state0, tidx0)
		if ptr != -1 {
			queue.EnqueueLong(int64(state0))
			queue.EnqueueInt(int32(tidx0))
			queue.EnqueueLong(ptr)
			reverseTable.PutElem(state0, tidx0, DiskGraphMaxPtr)
		}
	}

	for queue.Size() > 0 {
		curState := uint64(queue.DequeueLong())
		curTidx := int(queue.DequeueInt())
		curPtr := queue.DequeueLong()
		curNode, err := g.diskGraphNodeAt(curState, curTidx, curPtr)
		if err != nil {
			return nil, err
		}
		for i := 0; i < curNode.SuccSize(); i++ {
			nextState := curNode.GetStateFP(i)
			nextTidx := curNode.GetTIndex(i)
			if nextState == curState && nextTidx == curTidx {
				continue
			}
			if nextState == state && nextTidx == tidx {
				return g.reconstructReversePath(reverseTable, curState, curTidx, nextState, nextTidx)
			}

			nextLoc := reverseTable.GetNodesLoc(nextState)
			if nextLoc == -1 {
				continue
			}
			nextNodes := reverseTable.GetNodesByLoc(nextLoc)
			cloc := reverseTable.GetIdx(nextNodes, nextTidx)
			nextPtr := TableauGetElem(nextNodes, cloc)
			if IsDiskGraphFilePointer(nextPtr) {
				queue.EnqueueLong(int64(nextState))
				queue.EnqueueInt(int32(nextTidx))
				queue.EnqueueLong(nextPtr)
				curLoc := reverseTable.GetNodesLoc(curState)
				reverseTable.PutRecordElem(nextNodes, tableauDiskGraphInitState+int64(curLoc), curTidx, cloc)
			}
		}
	}
	return nil, NewRuntimeException(fmt.Sprintf("Couldn't re-create liveness trace (path) starting at: %d and tidx: %d", int64(state), tidx))
}

func (g *TableauDiskGraph) reconstructReversePath(reverseTable *TableauNodePtrTable, startState uint64, startTidx int, finalState uint64, finalTidx int) (*LongVec, error) {
	res := NewLongVecWithCapacity(2)
	res.AddElement(int64(finalState))

	lastTidx := finalTidx
	currentState := startState
	currentTidx := startTidx
	currentLoc := reverseTable.GetNodesLoc(currentState)
	nodes := reverseTable.GetNodesByLoc(currentLoc)
	for {
		if uint64(res.LastElement()) == currentState && lastTidx == currentTidx {
			return nil, NewRuntimeException("Self loop in trace path reconstruction")
		}
		res.AddElement(int64(currentState))
		lastTidx = currentTidx
		predecessorLocation := int64(-1)
		predecessorTidx := -1
		for j := 2; j < len(nodes); j += reverseTable.GetElemLength() {
			candidateLocation := TableauGetElem(nodes, j)
			candidateTidx := TableauGetTidx(nodes, j)
			if currentTidx == candidateTidx && !IsDiskGraphFilePointer(candidateLocation) {
				predecessorLocation = candidateLocation
				predecessorTidx = reverseTable.GetElemTidx(nodes, j)
				if candidateLocation == DiskGraphMaxPtr {
					break
				}
			}
		}
		if predecessorLocation == DiskGraphMaxPtr {
			break
		}
		currentLoc = int(predecessorLocation - tableauDiskGraphInitState)
		nodes = reverseTable.GetNodesByLoc(currentLoc)
		currentState = TableauGetKey(nodes)
		currentTidx = predecessorTidx
	}
	return res, nil
}

func (g *TableauDiskGraph) eachTableauGraphNode(fn func(*GraphNode) error) error {
	return g.eachTableauGraphNodeWithRestore(fn, true)
}

func (g *TableauDiskGraph) eachTableauGraphNodeWithRestore(fn func(*GraphNode) error, restorePositions bool) error {
	if g == nil || g.ptrFile == nil || fn == nil {
		return nil
	}
	nodePtr, err := g.nodeFile.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	ptrPtr, err := g.ptrFile.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	end, err := g.ptrFile.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}
	if _, err := g.ptrFile.Seek(0, io.SeekStart); err != nil {
		return err
	}
	in := g.ptrFile
	for {
		cur, err := g.ptrFile.Seek(0, io.SeekCurrent)
		if err != nil {
			return err
		}
		if cur >= end {
			break
		}
		fp, err := in.ReadLong()
		if err != nil {
			return err
		}
		tidx, err := in.ReadInt()
		if err != nil {
			return err
		}
		loc, err := in.ReadLongNat()
		if err != nil {
			return err
		}
		var node *GraphNode
		if g.ownedCache() == nil {
			node, err = g.getNodeFromDisk(uint64(fp), int(tidx), loc)
		} else {
			node, err = g.getNode(uint64(fp), int(tidx), loc)
		}
		if err != nil {
			return err
		}
		if err := fn(node); err != nil {
			return err
		}
	}
	if restorePositions {
		if _, err := g.nodeFile.Seek(nodePtr, io.SeekStart); err != nil {
			return err
		}
		_, err = g.ptrFile.Seek(ptrPtr, io.SeekStart)
		return err
	}
	return nil
}
