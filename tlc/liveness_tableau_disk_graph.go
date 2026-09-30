package tlc

import (
	"fmt"
	"io"
	"os"
	"strings"
)

const tableauDiskGraphInitState = DiskGraphMaxPtr + 1

type TableauDiskGraph struct {
	*DiskGraph
	TableauNodePtrTbl *TableauNodePtrTable
}

func NewTableauDiskGraph(metadir string, soln int) (*TableauDiskGraph, error) {
	base, err := NewDiskGraph(metadir, soln)
	if err != nil {
		return nil, err
	}
	return &TableauDiskGraph{DiskGraph: base, TableauNodePtrTbl: NewTableauNodePtrTable(255)}, nil
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
	return g.TableauNodePtrTbl.SetDone(fp)
}

func (g *TableauDiskGraph) RecordNode(fp uint64, tidx int) {
	g.TableauNodePtrTbl.Put(fp, tidx)
}

func (g *TableauDiskGraph) GetNodesByLoc(loc int) []int32 {
	return g.TableauNodePtrTbl.GetNodesByLoc(loc)
}

func (g *TableauDiskGraph) AddNode(node *GraphNode) (int64, error) {
	if node == nil {
		return -1, fmt.Errorf("cannot add nil graph node")
	}
	ptr, err := g.nodeFile.Seek(0, io.SeekEnd)
	if err != nil {
		return -1, err
	}
	g.putTableauNode(node, ptr)
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
		b.WriteString(fmt.Sprintf("<%d,%d> -> ", node.StateFP, node.TIndex))
		for i := 0; i < node.SuccSize(); i++ {
			b.WriteString(fmt.Sprintf("<%d,%d> ", node.GetStateFP(i), node.GetTIndex(i)))
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
