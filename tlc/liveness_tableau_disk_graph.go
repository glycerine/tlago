package tlc

import (
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
}

func NewTableauDiskGraph(metadir string, soln int, outDegreeStats ...any) (*TableauDiskGraph, error) {
	base, err := NewDiskGraph(metadir, soln, outDegreeStats...)
	if err != nil {
		return nil, err
	}
	return &TableauDiskGraph{DiskGraph: base, TableauNodePtrTbl: NewTableauNodePtrTable(255)}, nil
}

func NewDebugTableauDiskGraph(metadir string, soln int, outDegreeStats any, oos *OrderOfSolution) (*TableauDiskGraph, error) {
	graph, err := NewTableauDiskGraph(metadir, soln, outDegreeStats)
	if err != nil {
		return nil, err
	}
	graph.DebugEnabled = true
	graph.DebugOOS = oos
	return graph, nil
}

func (g *TableauDiskGraph) GetPtr(fp uint64, tidx int) int64 {
	return g.TableauNodePtrTbl.Get(fp, tidx)
}

func (g *TableauDiskGraph) GetElemLength() int {
	return g.TableauNodePtrTbl.GetElemLength()
}

func (g *TableauDiskGraph) IsDone(fp uint64) bool {
	return g.TableauNodePtrTbl.IsDone(fp)
}

func (g *TableauDiskGraph) SetDone(fp uint64) int {
	defer g.debugWriteDotForFP(fp)
	return g.TableauNodePtrTbl.SetDone(fp)
}

func (g *TableauDiskGraph) RecordNode(fp uint64, tidx int) {
	defer g.debugWriteDotForFP(fp)
	g.TableauNodePtrTbl.Put(fp, tidx)
}

func (g *TableauDiskGraph) GetNodesByLoc(loc int) []int32 {
	return g.TableauNodePtrTbl.GetNodesByLoc(loc)
}

func (g *TableauDiskGraph) AddNode(node *GraphNode) (int64, error) {
	if node == nil {
		return -1, fmt.Errorf("cannot add nil graph node")
	}
	defer g.debugWriteDotForFP(node.StateFP)
	if g.OutDegreeStats != nil {
		bucketStatisticsAddSample(g.OutDegreeStats, node.SuccSize())
	}
	ptr, err := g.nodeFile.Seek(0, io.SeekCurrent)
	if err != nil {
		return -1, err
	}
	g.putTableauNode(node, ptr)
	ptrOut := NewValueOutputStream(g.ptrFile)
	if err := ptrOut.WriteLong(int64(node.StateFP)); err != nil {
		return -1, err
	}
	if err := ptrOut.WriteInt(int32(node.TIndex)); err != nil {
		return -1, err
	}
	if err := ptrOut.WriteLongNat(ptr); err != nil {
		return -1, err
	}
	nodeOut := NewValueOutputStream(g.nodeFile)
	if err := node.Write(nodeOut); err != nil {
		return -1, err
	}
	return ptr, nil
}

func (g *TableauDiskGraph) debugWriteDotForFP(fp uint64) {
	if g == nil || !g.DebugEnabled || g.DebugOOS == nil {
		return
	}
	fragment := strconv.FormatInt(int64(fp), 10)
	if len(fragment) > 6 {
		fragment = fragment[:6]
	}
	path := filepath.Join(g.MetaDir, fmt.Sprintf("dgraph_%03d_%s.dot", g.DebugCounter, fragment))
	g.DebugCounter++

	hadCache := g.Cache != nil
	if !hadCache {
		g.CreateCache()
		defer g.DestroyCache()
	}
	if err := os.WriteFile(path, []byte(g.ToDotViz(g.DebugOOS, map[uint64]string{})), 0o644); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
}

func (g *TableauDiskGraph) putTableauNode(node *GraphNode, ptr int64) {
	g.TableauNodePtrTbl.PutElem(node.StateFP, node.TIndex, ptr)
}

func (g *TableauDiskGraph) CheckDuplicate(node *GraphNode) bool {
	if node == nil {
		return false
	}
	return g.TableauNodePtrTbl.Get(node.StateFP, node.TIndex) != -1
}

func (g *TableauDiskGraph) GetNode(fp uint64, tidx int) (*GraphNode, error) {
	ptr := g.TableauNodePtrTbl.Get(fp, tidx)
	if ptr < 0 {
		return NewGraphNode(fp, tidx), nil
	}
	if g.Cache == nil {
		return g.getNodeFromDisk(fp, tidx, ptr)
	}
	return g.getNode(fp, tidx, ptr)
}

func (g *TableauDiskGraph) GetLink(state uint64, tidx int) int64 {
	return g.TableauNodePtrTbl.Get(state, tidx)
}

func (g *TableauDiskGraph) PutLink(state uint64, tidx int, link int64) int64 {
	node := g.TableauNodePtrTbl.GetNodes(state)
	cloc := g.TableauNodePtrTbl.GetIdx(node, tidx)
	oldLink := TableauGetElem(node, cloc)
	if !IsDiskGraphFilePointer(oldLink) {
		return oldLink
	}
	TableauPutElem(node, link, cloc)
	return -1
}

func (g *TableauDiskGraph) SetMaxLink(state uint64, tidx int) {
	g.TableauNodePtrTbl.PutElem(state, tidx, DiskGraphMaxLink)
}

func (g *TableauDiskGraph) Reset() error {
	if err := g.DiskGraph.Reset(); err != nil {
		return err
	}
	g.TableauNodePtrTbl = NewTableauNodePtrTable(255)
	return nil
}

func (g *TableauDiskGraph) Recover() error {
	if g == nil {
		return nil
	}
	file, err := os.Open(g.chkptName + ".chkpt")
	if err != nil {
		return err
	}
	in := NewValueInputStream(file)
	nodePos, err := in.ReadLong()
	if err != nil {
		_ = file.Close()
		return err
	}
	ptrPos, err := in.ReadLong()
	if err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	g.TableauNodePtrTbl = NewTableauNodePtrTable(255)
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
	return g.TableauNodePtrTbl.Size()
}

func (g *TableauDiskGraph) CheckInvariants(slen int, alen int) (bool, error) {
	ok := true
	err := g.eachTableauGraphNode(func(node *GraphNode) error {
		if !node.CheckInvariants(slen, alen) {
			ok = false
		}
		return nil
	})
	return ok, err
}

func (g *TableauDiskGraph) MakeNodePtrTbl() error {
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
	return g.MakeNodePtrTblToTable(ptr, g.TableauNodePtrTbl)
}

func (g *TableauDiskGraph) MakeNodePtrTblToTable(ptr int64, table *TableauNodePtrTable) error {
	if _, err := g.ptrFile.Seek(0, io.SeekStart); err != nil {
		return err
	}
	in := NewValueInputStream(g.ptrFile)
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
	if g == nil || g.Cache == nil {
		return ""
	}
	var b strings.Builder
	_ = g.eachTableauGraphNode(func(node *GraphNode) error {
		b.WriteString(fmt.Sprintf("<%s,%d> -> ", livenessFPString(node.StateFP), node.TIndex))
		for i := 0; i < node.SuccSize(); i++ {
			b.WriteString(fmt.Sprintf("<%s,%d> ", livenessFPString(node.GetStateFP(i)), node.GetTIndex(i)))
		}
		b.WriteByte('\n')
		return nil
	})
	return b.String()
}

func (g *TableauDiskGraph) ToDotViz(oos *OrderOfSolution, labels map[uint64]string) string {
	if g == nil || g.Cache == nil || oos == nil {
		return ""
	}
	slen := len(oos.CheckState)
	alen := len(oos.CheckAction)
	var b strings.Builder
	b.WriteString("digraph DiskGraph {\n")
	b.WriteString("nodesep = 0.7\n")
	b.WriteString("rankdir=LR;\n")
	b.WriteString(diskGraphDotVizLegend(oos))
	b.WriteString(g.TableauNodePtrTbl.ToDotViz())
	b.WriteString("subgraph cluster_graph {\n")
	b.WriteString("color=\"white\";\n")
	_ = g.eachTableauGraphNode(func(node *GraphNode) error {
		b.WriteString(node.ToDotViz(g.isInitState(node), true, slen, alen, oos, labels))
		return nil
	})
	b.WriteString("}}")
	return b.String()
}

func (g *TableauDiskGraph) GetPath(state uint64, tidx int) (*LongVec, error) {
	if g == nil {
		return nil, fmt.Errorf("couldn't re-create liveness trace (path) starting at: %d and tidx: %d", state, tidx)
	}
	numOfInits := g.InitNodes.Size()
	for i := 0; i < numOfInits; i += 2 {
		state0 := uint64(g.InitNodes.ElementAt(i))
		tidx0 := int(g.InitNodes.ElementAt(i + 1))
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

	queue := NewIntQueueWithDisk(g.MetaDir, "", intQueueInitialSize)
	for i := 0; i < numOfInits; i += 2 {
		state0 := uint64(g.InitNodes.ElementAt(i))
		tidx0 := int(g.InitNodes.ElementAt(i + 1))
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
			if cloc == -1 {
				return nil, fmt.Errorf("liveness path successor missing tableau index %d for state %d", nextTidx, nextState)
			}
			nextPtr := TableauGetElem(nextNodes, cloc)
			if IsDiskGraphFilePointer(nextPtr) {
				queue.EnqueueLong(int64(nextState))
				queue.EnqueueInt(int32(nextTidx))
				queue.EnqueueLong(nextPtr)
				curLoc := reverseTable.GetNodesLoc(curState)
				if curLoc == -1 {
					return nil, fmt.Errorf("liveness path predecessor missing for state %d", curState)
				}
				reverseTable.PutRecordElem(nextNodes, tableauDiskGraphInitState+int64(curLoc), curTidx, cloc)
			}
		}
	}
	return nil, fmt.Errorf("couldn't re-create liveness trace (path) starting at: %d and tidx: %d", state, tidx)
}

func (g *TableauDiskGraph) reconstructReversePath(reverseTable *TableauNodePtrTable, startState uint64, startTidx int, finalState uint64, finalTidx int) (*LongVec, error) {
	res := NewLongVecWithCapacity(2)
	res.AddElement(int64(finalState))

	lastTidx := finalTidx
	currentState := startState
	currentTidx := startTidx
	currentLoc := reverseTable.GetNodesLoc(currentState)
	if currentLoc == -1 {
		return nil, fmt.Errorf("liveness path predecessor missing for state %d", currentState)
	}
	nodes := reverseTable.GetNodesByLoc(currentLoc)
	for {
		if uint64(res.LastElement()) == currentState && lastTidx == currentTidx {
			return nil, fmt.Errorf("self loop in trace path reconstruction")
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
		if currentLoc < 0 || currentLoc >= reverseTable.GetSize() {
			return nil, fmt.Errorf("liveness path predecessor location out of range: %d", currentLoc)
		}
		nodes = reverseTable.GetNodesByLoc(currentLoc)
		if nodes == nil {
			return nil, fmt.Errorf("liveness path predecessor node missing at location %d", currentLoc)
		}
		currentState = TableauGetKey(nodes)
		currentTidx = predecessorTidx
	}
	return res, nil
}

func (g *TableauDiskGraph) eachTableauGraphNode(fn func(*GraphNode) error) error {
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
	in := NewValueInputStream(g.ptrFile)
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
		if g.Cache == nil {
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
	if _, err := g.nodeFile.Seek(nodePtr, io.SeekStart); err != nil {
		return err
	}
	_, err = g.ptrFile.Seek(ptrPtr, io.SeekStart)
	return err
}
