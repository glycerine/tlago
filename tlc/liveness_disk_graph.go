package tlc

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type DiskGraph struct {
	MetaDir     string
	Solution    int
	chkptName   string
	nodeFile    *os.File
	ptrFile     *os.File
	InitNodes   *LongVec
	Cache       []*GraphNode
	NodePtrTbl  *NodePtrTable
	sizeAtCheck int64
}

func NewDiskGraph(metadir string, soln int) (*DiskGraph, error) {
	if err := os.MkdirAll(metadir, 0o755); err != nil {
		return nil, err
	}
	nodeFile, err := os.OpenFile(filepath.Join(metadir, fmt.Sprintf("nodes_%d", soln)), os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, err
	}
	ptrFile, err := os.OpenFile(filepath.Join(metadir, fmt.Sprintf("ptrs_%d", soln)), os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		_ = nodeFile.Close()
		return nil, err
	}
	return &DiskGraph{
		MetaDir:     metadir,
		Solution:    soln,
		chkptName:   filepath.Join(metadir, fmt.Sprintf("dgraph_%d", soln)),
		nodeFile:    nodeFile,
		ptrFile:     ptrFile,
		InitNodes:   NewLongVecWithCapacity(1),
		NodePtrTbl:  NewNodePtrTable(255),
		sizeAtCheck: 1,
	}, nil
}

func (g *DiskGraph) AddInitNode(node uint64, tidx int) {
	g.InitNodes.AddElement(int64(node))
	g.InitNodes.AddElement(int64(tidx))
}

func (g *DiskGraph) GetInitNodes() *LongVec {
	return g.InitNodes
}

func (g *DiskGraph) CreateCache() {
	g.Cache = make([]*GraphNode, 65536)
}

func (g *DiskGraph) DestroyCache() {
	g.Cache = nil
}

func (g *DiskGraph) Close() error {
	var err error
	if g.nodeFile != nil {
		err = g.nodeFile.Close()
	}
	if g.ptrFile != nil {
		if e := g.ptrFile.Close(); err == nil {
			err = e
		}
	}
	return err
}

func (g *DiskGraph) AddNode(node *GraphNode) (int64, error) {
	if node == nil {
		return -1, fmt.Errorf("cannot add nil graph node")
	}
	ptr, err := g.nodeFile.Seek(0, io.SeekEnd)
	if err != nil {
		return -1, err
	}
	g.putNode(node, ptr)
	if _, err := g.ptrFile.Seek(0, io.SeekEnd); err != nil {
		return -1, err
	}
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

func (g *DiskGraph) GetNode(fp uint64, tidx int) (*GraphNode, error) {
	return g.GetNodeNoTableau(fp)
}

func (g *DiskGraph) GetNodeNoTableau(fp uint64) (*GraphNode, error) {
	ptr := g.NodePtrTbl.Get(fp)
	if ptr < 0 {
		return NewGraphNode(fp, -1), nil
	}
	if g.Cache == nil {
		return g.getNodeFromDisk(fp, -1, ptr)
	}
	return g.getNode(fp, -1, ptr)
}

func (g *DiskGraph) GetPtr(fp uint64, tidx int) int64 {
	return g.NodePtrTbl.Get(fp)
}

func (g *DiskGraph) Reset() error {
	if err := g.nodeFile.Truncate(0); err != nil {
		return err
	}
	if _, err := g.nodeFile.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if err := g.ptrFile.Truncate(0); err != nil {
		return err
	}
	if _, err := g.ptrFile.Seek(0, io.SeekStart); err != nil {
		return err
	}
	g.NodePtrTbl = NewNodePtrTable(255)
	return nil
}

func (g *DiskGraph) FlushWritesToDiskFiles() error {
	if g == nil {
		return nil
	}
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
	if g == nil {
		return nil
	}
	if err := g.FlushWritesToDiskFiles(); err != nil {
		return err
	}
	nodePos, err := g.nodeFile.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	ptrPos, err := g.ptrFile.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	file, err := os.Create(g.chkptName + ".chkpt.tmp")
	if err != nil {
		return err
	}
	out := NewValueOutputStream(file)
	if err := out.WriteLong(nodePos); err != nil {
		_ = file.Close()
		return err
	}
	if err := out.WriteLong(ptrPos); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func (g *DiskGraph) CommitChkpt() error {
	if g == nil {
		return nil
	}
	oldName := g.chkptName + ".chkpt"
	newName := g.chkptName + ".chkpt.tmp"
	if _, err := os.Stat(oldName); err == nil {
		if err := os.Remove(oldName); err != nil {
			return err
		}
	}
	return os.Rename(newName, oldName)
}

func (g *DiskGraph) Recover() error {
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
	g.NodePtrTbl = NewNodePtrTable(255)
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
	if g == nil {
		return 0, nil
	}
	var size int64
	if g.nodeFile != nil {
		info, err := g.nodeFile.Stat()
		if err != nil {
			return 0, err
		}
		size += info.Size()
	}
	if g.ptrFile != nil {
		info, err := g.ptrFile.Stat()
		if err != nil {
			return 0, err
		}
		size += info.Size()
	}
	return size, nil
}

func (g *DiskGraph) GetSizeAtLastCheck() int64 {
	if g == nil {
		return 0
	}
	return g.sizeAtCheck
}

func (g *DiskGraph) RecordSize() {
	if g != nil {
		g.sizeAtCheck = int64(g.Size())
	}
}

func (g *DiskGraph) putNode(node *GraphNode, ptr int64) {
	g.NodePtrTbl.Put(node.StateFP, ptr)
}

func (g *DiskGraph) CheckDuplicate(node *GraphNode) bool {
	if node == nil {
		return false
	}
	return g.NodePtrTbl.Get(node.StateFP) != -1
}

func (g *DiskGraph) GetLink(state uint64, tidx int) int64 {
	return g.NodePtrTbl.Get(state)
}

func (g *DiskGraph) PutLink(state uint64, tidx int, link int64) int64 {
	loc := g.NodePtrTbl.GetLoc(state)
	oldLink := g.NodePtrTbl.GetByLoc(loc)
	if !IsDiskGraphFilePointer(oldLink) {
		return oldLink
	}
	g.NodePtrTbl.PutByLoc(state, link, loc)
	return -1
}

func (g *DiskGraph) SetMaxLink(state uint64, tidx int) {
	g.NodePtrTbl.Put(state, DiskGraphMaxLink)
}

func (g *DiskGraph) MakeNodePtrTbl() error {
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
		if _, err := in.ReadInt(); err != nil {
			return err
		}
		loc, err := in.ReadLongNat()
		if err != nil {
			return err
		}
		g.NodePtrTbl.Put(uint64(fp), loc)
	}
}

func (g *DiskGraph) Size() int {
	return g.NodePtrTbl.Size()
}

func (g *DiskGraph) CheckInvariants(slen int, alen int) (bool, error) {
	ok := true
	err := g.eachGraphNode(func(node *GraphNode) error {
		if !node.CheckInvariants(slen, alen) {
			ok = false
		}
		return nil
	})
	return ok, err
}

func (g *DiskGraph) String() string {
	if g == nil || g.Cache == nil {
		return ""
	}
	var b strings.Builder
	_ = g.eachGraphNode(func(node *GraphNode) error {
		b.WriteString(fmt.Sprintf("%d -> ", node.StateFP))
		for i := 0; i < node.SuccSize(); i++ {
			b.WriteString(fmt.Sprintf("%d ", node.GetStateFP(i)))
		}
		b.WriteByte('\n')
		return nil
	})
	return b.String()
}

func (g *DiskGraph) ToDotViz(oos *OrderOfSolution, labels map[uint64]string) string {
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
	b.WriteString("subgraph cluster_graph {\n")
	b.WriteString("color=\"white\";\n")
	_ = g.eachGraphNode(func(node *GraphNode) error {
		b.WriteString(node.ToDotViz(g.isInitState(node), false, slen, alen, oos, labels))
		return nil
	})
	b.WriteString("}}")
	return b.String()
}

func (g *DiskGraph) eachGraphNode(fn func(*GraphNode) error) error {
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

func (g *DiskGraph) isInitState(node *GraphNode) bool {
	if g == nil || node == nil || g.InitNodes == nil {
		return false
	}
	for i := 0; i < g.InitNodes.Size(); i += 2 {
		if uint64(g.InitNodes.ElementAt(i)) == node.StateFP && int(g.InitNodes.ElementAt(i+1)) == node.TIndex {
			return true
		}
	}
	return false
}

func diskGraphDotVizLegend(oos *OrderOfSolution) string {
	if oos == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString("subgraph cluster_legend {")
	b.WriteString("graph[style=bold];")
	b.WriteString("label = \"PossibleErrorModel\" style=\"solid\"\n")
	b.WriteString("node [ labeljust=\"l\",shape=record ]\n")
	for i, node := range oos.CheckState {
		b.WriteString(fmt.Sprintf("S%d [label=\"S%d: %s\"]\n", i+1, i+1, liveNodeToDot(node)))
	}
	for i, node := range oos.CheckAction {
		b.WriteString(fmt.Sprintf("A%d [label=\"A%d: %s\"]\n", i+1, i+1, liveNodeToDot(node)))
	}
	b.WriteString("}")
	return b.String()
}

func liveNodeToDot(node *LiveExprNode) string {
	if node == nil {
		return ""
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
	idx := int(fp+uint64(tidx)) & 0xffff
	gnode := g.Cache[idx]
	if gnode != nil && gnode.StateFP == fp && gnode.TIndex == tidx {
		return gnode, nil
	}
	gnode, err := g.getNodeFromDisk(fp, tidx, ptr)
	if err != nil {
		return nil, err
	}
	if g.Cache[idx] == nil {
		g.Cache[idx] = gnode
	}
	return gnode, nil
}

func (g *DiskGraph) getNodeFromDisk(fp uint64, tidx int, ptr int64) (*GraphNode, error) {
	if ptr < 0 {
		return nil, fmt.Errorf("invalid negative file pointer: %d", ptr)
	}
	cur, err := g.nodeFile.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}
	if _, err := g.nodeFile.Seek(ptr, io.SeekStart); err != nil {
		return nil, err
	}
	gnode := NewGraphNode(fp, tidx)
	if err := gnode.Read(NewValueInputStream(g.nodeFile)); err != nil {
		return nil, err
	}
	_, err = g.nodeFile.Seek(cur, io.SeekStart)
	return gnode, err
}
