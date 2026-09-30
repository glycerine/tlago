package tlc

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type DiskGraph struct {
	MetaDir    string
	Solution   int
	nodeFile   *os.File
	ptrFile    *os.File
	InitNodes  *LongVec
	Cache      []*GraphNode
	NodePtrTbl *NodePtrTable
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
		MetaDir:    metadir,
		Solution:   soln,
		nodeFile:   nodeFile,
		ptrFile:    ptrFile,
		InitNodes:  NewLongVecWithCapacity(1),
		NodePtrTbl: NewNodePtrTable(255),
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
