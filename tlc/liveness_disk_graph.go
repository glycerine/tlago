package tlc

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type DiskGraph struct {
	MetaDir        string
	Solution       int
	chkptName      string
	nodeFile       *livenessDataFile
	ptrFile        *livenessDataFile
	InitNodes      *LongVec
	Cache          []*GraphNode
	NodePtrTbl     *NodePtrTable
	OutDegreeStats any
	sizeAtCheck    int64
	nodeReadMu     sync.Mutex

	ownersCaptured      bool
	metaDirOwner        string
	initNodesOwner      *LongVec
	cacheOwner          []*GraphNode
	nodePtrTableOwner   *NodePtrTable
	outDegreeStatsOwner any
}

type stopDiskGraphIteration struct{}

func (*stopDiskGraphIteration) Error() string { return "stop disk graph iteration" }

func exitLivenessDiskGraphIO(failure error) {
	if failure == nil {
		return
	}
	if !isJavaIOException(failure) {
		panic(failure)
	}
	PrintErrorThrowable(ECSystemDiskgraphAccess, failure)
	ExitTLCProcess(1)
}

func (g *DiskGraph) requireDiskGraph() {
	if g == nil {
		panic(NewNullPointerException())
	}
}

func NewDiskGraph(metadir string, soln int, outDegreeStats ...any) (*DiskGraph, error) {
	if len(outDegreeStats) > 1 {
		panic(NewIllegalArgumentException("NewDiskGraph accepts at most one out-degree statistics object"))
	}
	soln = int(int32(soln))
	nodeFile, err := newLivenessDataFile(filepath.Join(metadir, fmt.Sprintf("nodes_%d", soln)))
	if err != nil {
		return nil, err
	}
	ptrFile, err := newLivenessDataFile(filepath.Join(metadir, fmt.Sprintf("ptrs_%d", soln)))
	if err != nil {
		return nil, err
	}
	var stats any
	if len(outDegreeStats) > 0 {
		stats = outDegreeStats[0]
	}
	initNodes := NewLongVecWithCapacity(1)
	nodePtrTable := NewNodePtrTable(255)
	return &DiskGraph{
		MetaDir:        metadir,
		Solution:       soln,
		chkptName:      filepath.Join(metadir, fmt.Sprintf("dgraph_%d", soln)),
		nodeFile:       nodeFile,
		ptrFile:        ptrFile,
		InitNodes:      initNodes,
		NodePtrTbl:     nodePtrTable,
		OutDegreeStats: stats,
		sizeAtCheck:    1,
		ownersCaptured: true, metaDirOwner: metadir, initNodesOwner: initNodes,
		nodePtrTableOwner: nodePtrTable, outDegreeStatsOwner: stats,
	}, nil
}

func (g *DiskGraph) ownedMetaDir() string {
	if g.ownersCaptured {
		return g.metaDirOwner
	}
	return g.MetaDir
}

func (g *DiskGraph) ownedInitNodes() *LongVec {
	if g.ownersCaptured {
		return g.initNodesOwner
	}
	return g.InitNodes
}

func (g *DiskGraph) ownedCache() []*GraphNode {
	if g.ownersCaptured {
		return g.cacheOwner
	}
	return g.Cache
}

func (g *DiskGraph) setCache(cache []*GraphNode) {
	g.Cache = cache
	if g.ownersCaptured {
		g.cacheOwner = cache
	}
}

func (g *DiskGraph) ownedNodePtrTable() *NodePtrTable {
	if g.ownersCaptured {
		return g.nodePtrTableOwner
	}
	return g.NodePtrTbl
}

func (g *DiskGraph) setNodePtrTable(table *NodePtrTable) {
	g.NodePtrTbl = table
	if g.ownersCaptured {
		g.nodePtrTableOwner = table
	}
}

func (g *DiskGraph) ownedOutDegreeStats() any {
	if g.ownersCaptured {
		return g.outDegreeStatsOwner
	}
	return g.OutDegreeStats
}

func (g *DiskGraph) AddInitNode(node uint64, tidx int) {
	g.requireDiskGraph()
	tidx = graphInt(tidx)
	g.ownedInitNodes().AddElement(int64(node))
	g.ownedInitNodes().AddElement(int64(tidx))
}

func (g *DiskGraph) GetInitNodes() *LongVec {
	g.requireDiskGraph()
	return g.ownedInitNodes()
}

func (g *DiskGraph) CreateCache() {
	g.requireDiskGraph()
	g.setCache(make([]*GraphNode, 65536))
}

func (g *DiskGraph) DestroyCache() {
	g.requireDiskGraph()
	g.setCache(nil)
}

func (g *DiskGraph) Close() error {
	g.requireDiskGraph()
	if g.nodeFile != nil {
		if err := g.nodeFile.Close(); err != nil {
			return err
		}
	}
	if g.ptrFile != nil {
		return g.ptrFile.Close()
	}
	return nil
}

func (g *DiskGraph) AddNode(node *GraphNode) (int64, error) {
	g.requireDiskGraph()
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
	g.putNode(node, ptr)
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

func (g *DiskGraph) GetNode(fp uint64, tidx int) (*GraphNode, error) {
	g.requireDiskGraph()
	return g.GetNodeNoTableau(fp)
}

func (g *DiskGraph) GetNodeNoTableau(fp uint64) (*GraphNode, error) {
	g.requireDiskGraph()
	ptr := g.ownedNodePtrTable().Get(fp)
	if ptr < 0 {
		return NewGraphNode(fp, -1), nil
	}
	if g.ownedCache() == nil {
		return g.getNodeFromDisk(fp, -1, ptr)
	}
	return g.getNode(fp, -1, ptr)
}

func (g *DiskGraph) GetNodeAt(fp uint64, tidx int, ptr int64) (*GraphNode, error) {
	g.requireDiskGraph()
	tidx = graphInt(tidx)
	if g.ownedCache() == nil {
		panic(NewNullPointerException())
	}
	return g.getNode(fp, tidx, ptr)
}

func (g *DiskGraph) GetPtr(fp uint64, tidx int) int64 {
	g.requireDiskGraph()
	return g.GetPtrNoTableau(fp)
}

func (g *DiskGraph) GetPtrNoTableau(fp uint64) int64 {
	g.requireDiskGraph()
	return g.ownedNodePtrTable().Get(fp)
}

func (g *DiskGraph) Reset() error {
	g.requireDiskGraph()
	if err := g.ptrFile.Reset(); err != nil {
		return err
	}
	if err := g.nodeFile.Reset(); err != nil {
		return err
	}
	g.setNodePtrTable(NewNodePtrTable(255))
	return nil
}

func (g *DiskGraph) FlushWritesToDiskFiles() error {
	g.requireDiskGraph()
	if g.nodeFile != nil {
		if err := g.nodeFile.Sync(); err != nil {
			return err
		}
	}
	if g.ptrFile != nil {
		if err := g.ptrFile.Sync(); err != nil {
			return err
		}
	}
	return nil
}

func (g *DiskGraph) BeginChkpt() error {
	g.requireDiskGraph()
	g.nodeReadMu.Lock()
	defer g.nodeReadMu.Unlock()
	if err := g.FlushWritesToDiskFiles(); err != nil {
		return err
	}
	file, err := openFileOutputStream(g.chkptName + ".chkpt.tmp")
	if err != nil {
		return err
	}
	nodePos, err := g.nodeFile.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	if err := binary.Write(file, binary.BigEndian, nodePos); err != nil {
		return bufferedRandomAccessFileIOError(err)
	}
	ptrPos, err := g.ptrFile.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	if err := binary.Write(file, binary.BigEndian, ptrPos); err != nil {
		return bufferedRandomAccessFileIOError(err)
	}
	return bufferedRandomAccessFileIOError(file.Close())
}

func (g *DiskGraph) CommitChkpt() error {
	g.requireDiskGraph()
	oldName := g.chkptName + ".chkpt"
	newName := g.chkptName + ".chkpt.tmp"
	if _, err := os.Stat(oldName); err == nil {
		if err := os.Remove(oldName); err != nil {
			return NewIOException(fmt.Sprintf("DiskGraph.commitChkpt: cannot delete %s", oldName))
		}
	}
	if err := os.Rename(newName, oldName); err != nil {
		return NewIOException(fmt.Sprintf("DiskGraph.commitChkpt: cannot delete %s", oldName))
	}
	return nil
}

func (g *DiskGraph) Recover() error {
	g.requireDiskGraph()
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

func (g *DiskGraph) GetSizeOnDisk() (int64, error) {
	g.requireDiskGraph()
	var size int64
	if g.ptrFile != nil {
		info, err := g.ptrFile.Stat()
		if err != nil {
			return 0, err
		}
		size += info.Size()
	}
	if g.nodeFile != nil {
		info, err := g.nodeFile.Stat()
		if err != nil {
			return 0, err
		}
		size += info.Size()
	}
	return size, nil
}

func (g *DiskGraph) GetSizeAtLastCheck() int64 {
	g.requireDiskGraph()
	return g.sizeAtCheck
}

func (g *DiskGraph) RecordSize() {
	g.requireDiskGraph()
	g.sizeAtCheck = int64(g.Size())
}

func (g *DiskGraph) putNode(node *GraphNode, ptr int64) {
	g.ownedNodePtrTable().Put(node.ownedStateFP(), ptr)
}

func (g *DiskGraph) CheckDuplicate(node *GraphNode) bool {
	g.requireDiskGraph()
	if node == nil {
		panic(NewNullPointerException())
	}
	return g.ownedNodePtrTable().Get(node.ownedStateFP()) != -1
}

func (g *DiskGraph) GetLink(state uint64, tidx int) int64 {
	g.requireDiskGraph()
	return g.ownedNodePtrTable().Get(state)
}

func (g *DiskGraph) PutLink(state uint64, tidx int, link int64) int64 {
	g.requireDiskGraph()
	if link < DiskGraphMaxPtr || link >= DiskGraphMaxLink {
		panic(NewAssertionError())
	}
	loc := g.ownedNodePtrTable().GetLoc(state)
	oldLink := g.ownedNodePtrTable().GetByLoc(loc)
	if !IsDiskGraphFilePointer(oldLink) {
		return oldLink
	}
	g.ownedNodePtrTable().PutByLoc(state, link, loc)
	return -1
}

func (g *DiskGraph) SetMaxLink(state uint64, tidx int) {
	g.requireDiskGraph()
	g.ownedNodePtrTable().Put(state, DiskGraphMaxLink)
}

func (g *DiskGraph) MakeNodePtrTbl() error {
	g.requireDiskGraph()
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

func (g *DiskGraph) MakeNodePtrTblTo(ptr int64) error {
	g.requireDiskGraph()
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
		if _, err := in.ReadInt(); err != nil {
			return err
		}
		loc, err := in.ReadLongNat()
		if err != nil {
			return err
		}
		g.ownedNodePtrTable().Put(uint64(fp), loc)
	}
}

func (g *DiskGraph) Size() int {
	g.requireDiskGraph()
	return g.ownedNodePtrTable().Size()
}

func (g *DiskGraph) CheckInvariants(slen int, alen int) (bool, error) {
	g.requireDiskGraph()
	stop := &stopDiskGraphIteration{}
	err := g.eachGraphNodeWithRestore(func(node *GraphNode) error {
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

func (g *DiskGraph) CalculateOutDegreeDiskGraph(stats any) (result any, err error) {
	defer func() {
		if err != nil {
			exitLivenessDiskGraphIO(err)
		}
	}()
	g.requireDiskGraph()
	if err := g.ptrFile.Sync(); err != nil {
		return stats, err
	}
	if err := g.nodeFile.Sync(); err != nil {
		return stats, err
	}
	if _, err := g.ptrFile.Seek(0, io.SeekStart); err != nil {
		return stats, err
	}
	end, err := g.ptrFile.Seek(0, io.SeekEnd)
	if err != nil {
		return stats, err
	}
	if _, err := g.ptrFile.Seek(0, io.SeekStart); err != nil {
		return stats, err
	}
	for {
		cur, err := g.ptrFile.Seek(0, io.SeekCurrent)
		if err != nil {
			return stats, err
		}
		if cur >= end {
			return stats, nil
		}
		if _, err := g.ptrFile.Seek(cur+12, io.SeekStart); err != nil {
			return stats, err
		}
		ptr, err := g.ptrFile.ReadLongNat()
		if err != nil {
			return stats, err
		}
		if _, err := g.nodeFile.Seek(ptr, io.SeekStart); err != nil {
			return stats, err
		}
		count, err := g.nodeFile.ReadNat()
		if err != nil {
			return stats, err
		}
		if isNil(stats) {
			panic(NewNullPointerException())
		}
		bucketStatisticsAddSample(stats, int(count)/graphNodeRecordSize)
	}
}

func (g *DiskGraph) CalculateInDegreeDiskGraph(stats any) (result any, err error) {
	defer func() {
		if err != nil {
			exitLivenessDiskGraphIO(err)
		}
	}()
	g.requireDiskGraph()
	counts := NewInsMap[string, int]()
	if err := g.nodeFile.Sync(); err != nil {
		return stats, err
	}
	if _, err := g.nodeFile.Seek(0, io.SeekStart); err != nil {
		return stats, err
	}
	end, err := g.nodeFile.Seek(0, io.SeekEnd)
	if err != nil {
		return stats, err
	}
	if _, err := g.nodeFile.Seek(0, io.SeekStart); err != nil {
		return stats, err
	}
	for {
		cur, err := g.nodeFile.Seek(0, io.SeekCurrent)
		if err != nil {
			return stats, err
		}
		if cur >= end {
			break
		}
		words, err := g.nodeFile.ReadNat()
		if err != nil {
			return stats, err
		}
		for i := 0; i < int(words)/graphNodeRecordSize; i++ {
			high, err := g.nodeFile.ReadInt()
			if err != nil {
				return stats, err
			}
			low, err := g.nodeFile.ReadInt()
			if err != nil {
				return stats, err
			}
			tidx, err := g.nodeFile.ReadInt()
			if err != nil {
				return stats, err
			}
			fp := uint64(uint32(high))<<32 | uint64(uint32(low))
			key := graphNodeKey(fp, int(tidx))
			counts.Set(key, counts.Get(key)+1)
		}
		checksLen, err := g.nodeFile.ReadNat()
		if err != nil {
			return stats, err
		}
		position, err := g.nodeFile.Seek(0, io.SeekCurrent)
		if err != nil {
			return stats, err
		}
		skip := int64(int32(checksLen * 8))
		if _, err := g.nodeFile.Seek(position+skip, io.SeekStart); err != nil {
			return stats, err
		}
	}
	for _, count := range counts.All() {
		if isNil(stats) {
			panic(NewNullPointerException())
		}
		bucketStatisticsAddSample(stats, count)
	}
	return stats, nil
}

func (g *DiskGraph) String() string {
	if g == nil {
		panic(NewNullPointerException())
	}
	if g.ownedCache() == nil {
		return ""
	}
	var b strings.Builder
	err := g.eachGraphNode(func(node *GraphNode) error {
		b.WriteString(fmt.Sprintf("%s -> ", livenessFPString(node.ownedStateFP())))
		for i := 0; i < node.SuccSize(); i++ {
			b.WriteString(fmt.Sprintf("%s ", livenessFPString(node.GetStateFP(i))))
		}
		b.WriteByte('\n')
		return nil
	})
	if err != nil {
		exitLivenessDiskGraphIO(err)
	}
	return b.String()
}

func (g *DiskGraph) ToDotViz(oos *OrderOfSolution, labels map[uint64]string) string {
	if g == nil {
		panic(NewNullPointerException())
	}
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
	b.WriteString("subgraph cluster_graph {\n")
	b.WriteString("color=\"white\";\n")
	err := g.eachGraphNode(func(node *GraphNode) error {
		b.WriteString(node.ToDotViz(g.isInitState(node), false, slen, alen, oos, labels))
		return nil
	})
	if err != nil {
		exitLivenessDiskGraphIO(err)
	}
	b.WriteString("}}")
	return b.String()
}

func (g *DiskGraph) WriteDotViz(oos *OrderOfSolution, path string, labelSets ...map[uint64]string) {
	g.requireDiskGraph()
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

func writeLivenessDotFile(path string, render func() string) error {
	file, err := openFileOutputStream(path)
	if err != nil {
		return err
	}
	writer := NewJavaBufferedWriter(file)
	if _, err := writer.WriteString(render()); err != nil {
		return bufferedRandomAccessFileIOError(err)
	}
	if err := writer.Flush(); err != nil {
		return bufferedRandomAccessFileIOError(err)
	}
	return bufferedRandomAccessFileIOError(file.Close())
}

func (g *DiskGraph) GetPath(state uint64, tidxIgnored int) (*LongVec, error) {
	g.requireDiskGraph()
	tidxIgnored = graphInt(tidxIgnored)
	numOfInits := g.ownedInitNodes().Size()
	for i := 0; i < numOfInits; i += 2 {
		state0 := uint64(g.ownedInitNodes().ElementAt(i))
		if state0 == state {
			res := NewLongVecWithCapacity(1)
			res.AddElement(int64(state0))
			return res, nil
		}
	}

	if err := g.MakeNodePtrTbl(); err != nil {
		return nil, err
	}

	offset := DiskGraphMaxPtr + 1
	queue := NewIntQueueWithDisk(g.ownedMetaDir(), "", intQueueInitialSize)
	for i := 0; i < numOfInits; i += 2 {
		state0 := uint64(g.ownedInitNodes().ElementAt(i))
		ptr := g.ownedNodePtrTable().Get(state0)
		if ptr != -1 {
			queue.EnqueueLong(int64(state0))
			queue.EnqueueLong(ptr)
			g.ownedNodePtrTable().Put(state0, DiskGraphMaxPtr)
		}
	}

	for queue.Size() > 0 {
		curState := uint64(queue.DequeueLong())
		curPtr := queue.DequeueLong()
		curNode, err := g.diskGraphNodeAt(curState, -1, curPtr)
		if err != nil {
			return nil, err
		}
		for i := 0; i < curNode.SuccSize(); i++ {
			nextState := curNode.GetStateFP(i)
			if nextState == state {
				res := NewLongVecWithCapacity(2)
				res.AddElement(int64(nextState))
				curLoc := g.ownedNodePtrTable().GetLoc(curState)
				for {
					res.AddElement(int64(curState))
					ploc := g.ownedNodePtrTable().GetByLoc(curLoc)
					if ploc == DiskGraphMaxPtr {
						break
					}
					curLoc = int(ploc - offset)
					curState = g.ownedNodePtrTable().GetKeyByLoc(curLoc)
				}
				return res, nil
			}

			nextLoc := g.ownedNodePtrTable().GetLoc(nextState)
			if nextLoc == -1 {
				continue
			}
			nextPtr := g.ownedNodePtrTable().GetByLoc(nextLoc)
			if IsDiskGraphFilePointer(nextPtr) {
				queue.EnqueueLong(int64(nextState))
				queue.EnqueueLong(nextPtr)
				curLoc := g.ownedNodePtrTable().GetLoc(curState)
				g.ownedNodePtrTable().PutByLoc(nextState, offset+int64(curLoc), nextLoc)
			}
		}
	}
	return nil, NewRuntimeException(fmt.Sprintf("Couldn't re-create liveness trace (path) starting at: %d and tidx: %d", int64(state), tidxIgnored))
}

func (g *DiskGraph) eachGraphNode(fn func(*GraphNode) error) error {
	return g.eachGraphNodeWithRestore(fn, true)
}

func (g *DiskGraph) eachGraphNodeWithRestore(fn func(*GraphNode) error, restorePositions bool) error {
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

func (g *DiskGraph) diskGraphNodeAt(fp uint64, tidx int, ptr int64) (*GraphNode, error) {
	g.requireDiskGraph()
	if g.ownedCache() == nil {
		return g.getNodeFromDisk(fp, tidx, ptr)
	}
	return g.getNode(fp, tidx, ptr)
}

func (g *DiskGraph) isInitState(node *GraphNode) bool {
	g.requireDiskGraph()
	numOfInits := g.ownedInitNodes().Size()
	for i := 0; i < numOfInits; i += 2 {
		state := uint64(g.ownedInitNodes().ElementAt(i))
		tidx := int(g.ownedInitNodes().ElementAt(i + 1))
		if node == nil {
			panic(NewNullPointerException())
		}
		if node.ownedStateFP() == state && node.ownedTIndex() == tidx {
			return true
		}
	}
	return false
}

func diskGraphDotVizLegend(oos *OrderOfSolution) string {
	checkState := oos.GetCheckState()
	if checkState == nil {
		panic(NewNullPointerException())
	}
	var b strings.Builder
	b.WriteString("subgraph cluster_legend {")
	b.WriteString("graph[style=bold];")
	b.WriteString("label = \"PossibleErrorModel\" style=\"solid\"\n")
	b.WriteString("node [ labeljust=\"l\",shape=record ]\n")
	for i, node := range checkState {
		b.WriteString(fmt.Sprintf("S%d [label=\"S%d: %s\"]\n", i+1, i+1, liveNodeToDot(node)))
	}
	checkAction := oos.GetCheckAction()
	if checkAction == nil {
		panic(NewNullPointerException())
	}
	for i, node := range checkAction {
		b.WriteString(fmt.Sprintf("A%d [label=\"A%d: %s\"]\n", i+1, i+1, liveNodeToDot(node)))
	}
	b.WriteString("}")
	return b.String()
}

func liveNodeToDot(node *LiveExprNode) string {
	if node == nil {
		panic(NewNullPointerException())
	}
	text := node.String()
	text = strings.ReplaceAll(text, "\\", "\\\\")
	text = strings.ReplaceAll(text, "\"", "\\\"")
	text = strings.ReplaceAll(text, "<", "\\<")
	text = strings.ReplaceAll(text, ">", "\\>")
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, "\n", "\\l")
	return text
}

func (g *DiskGraph) getNode(fp uint64, tidx int, ptr int64) (*GraphNode, error) {
	g.requireDiskGraph()
	g.nodeReadMu.Lock()
	defer g.nodeReadMu.Unlock()
	return g.getNodeLocked(fp, tidx, ptr)
}

func (g *DiskGraph) getNodeLocked(fp uint64, tidx int, ptr int64) (*GraphNode, error) {
	if g.ownedCache() == nil {
		panic(NewNullPointerException())
	}
	idx := int(fp+uint64(tidx)) & 0xffff
	if idx < 0 || idx >= len(g.ownedCache()) {
		panic(NewArrayIndexOutOfBoundsException(idx, len(g.ownedCache())))
	}
	gnode := g.ownedCache()[idx]
	if gnode != nil && gnode.ownedStateFP() == fp && gnode.ownedTIndex() == tidx {
		return gnode, nil
	}
	gnode1, err := g.getNodeFromDiskLocked(fp, tidx, ptr)
	if err != nil {
		return nil, err
	}
	if gnode == nil {
		g.ownedCache()[idx] = gnode1
	}
	return gnode1, nil
}

func (g *DiskGraph) getNodeFromDisk(fp uint64, tidx int, ptr int64) (*GraphNode, error) {
	g.requireDiskGraph()
	g.nodeReadMu.Lock()
	defer g.nodeReadMu.Unlock()
	return g.getNodeFromDiskLocked(fp, tidx, ptr)
}

func (g *DiskGraph) getNodeFromDiskLocked(fp uint64, tidx int, ptr int64) (*GraphNode, error) {
	if ptr < 0 {
		panic(NewIllegalArgumentException(fmt.Sprintf("Invalid negative file pointer: %d", ptr)))
	}
	cur, err := g.nodeFile.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}
	if _, err := g.nodeFile.Seek(ptr, io.SeekStart); err != nil {
		return nil, err
	}
	gnode := NewGraphNode(fp, tidx)
	if err := gnode.Read(g.nodeFile); err != nil {
		return nil, err
	}
	_, err = g.nodeFile.Seek(cur, io.SeekStart)
	return gnode, err
}
