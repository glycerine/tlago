/*******************************************************************************
 * Copyright (c) 2015 Microsoft Research. All rights reserved.
 *
 * The MIT License (MIT)
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
 * of the Software, and to permit persons to whom the Software is furnished to do
 * so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN
 * AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
 * WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 *
 * Contributors:
 *   Markus Alexander Kuppe - initial API and implementation
 ******************************************************************************/
package tlc

import (
	"fmt"
	"testing"
)

// Concrete adapter for DiskGraphTest's AbstractDiskGraph and subclass override.
// Preserve dedicated directories, source solution number1 and bucket bound16.
type javaOriginalDiskGraph struct {
	tt      *testing.T
	disk    *DiskGraph
	tableau *TableauDiskGraph
}

var javaDiskGraphTestStats = NewFixedSizedBucketStatistics("Test Dummy", 16)
var javaTableauDiskGraphTestStats = NewFixedSizedBucketStatistics("Test Dummy", 16)

func javaOriginalGraph(tt *testing.T, tableau bool) *javaOriginalDiskGraph {
	tt.Helper()
	g := &javaOriginalDiskGraph{tt: tt}
	var err error
	if tableau {
		g.tableau, err = NewTableauDiskGraph(tt.TempDir(), 1, javaTableauDiskGraphTestStats)
		if err == nil {
			g.disk = g.tableau.DiskGraph
		}
	} else {
		g.disk, err = NewDiskGraph(tt.TempDir(), 1, javaDiskGraphTestStats)
	}
	if err != nil {
		tt.Fatal(err)
	}
	tt.Cleanup(func() {
		if err := g.disk.Close(); err != nil {
			tt.Error(err)
		}
	})
	return g
}
func (g *javaOriginalDiskGraph) addNode(n *GraphNode) int64 {
	g.tt.Helper()
	var ptr int64
	var err error
	if g.tableau != nil {
		ptr, err = g.tableau.AddNode(n)
	} else {
		ptr, err = g.disk.AddNode(n)
	}
	if err != nil {
		g.tt.Fatal(err)
	}
	return ptr
}
func (g *javaOriginalDiskGraph) getNode(fp uint64, tidx int, ptr ...int64) *GraphNode {
	g.tt.Helper()
	var n *GraphNode
	var err error
	if len(ptr) > 0 {
		n, err = g.disk.getNode(fp, tidx, ptr[0])
	} else if g.tableau != nil {
		n, err = g.tableau.GetNode(fp, tidx)
	} else {
		n, err = g.disk.GetNode(fp, tidx)
	}
	if err != nil {
		g.tt.Fatal(err)
	}
	return n
}
func (g *javaOriginalDiskGraph) getPathRaw(fp uint64, tidx int) (*LongVec, error) {
	if g.tableau != nil {
		return g.tableau.GetPath(fp, tidx)
	}
	return g.disk.GetPath(fp, tidx)
}
func (g *javaOriginalDiskGraph) getPath(fp uint64, tidx int) *LongVec {
	g.tt.Helper()
	p, err := g.getPathRaw(fp, tidx)
	if err != nil {
		g.tt.Fatal(err)
	}
	return p
}
func (g *javaOriginalDiskGraph) getLink(fp uint64, tidx int) int64 {
	if g.tableau != nil {
		return g.tableau.GetLink(fp, tidx)
	}
	return g.disk.GetLink(fp, tidx)
}
func (g *javaOriginalDiskGraph) makeNodePtrTbl() {
	g.tt.Helper()
	var err error
	if g.tableau != nil {
		err = g.tableau.MakeNodePtrTbl()
	} else {
		err = g.disk.MakeNodePtrTbl()
	}
	if err != nil {
		g.tt.Fatal(err)
	}
}
func (g *javaOriginalDiskGraph) beginChkpt() {
	g.tt.Helper()
	if err := g.disk.BeginChkpt(); err != nil {
		g.tt.Fatal(err)
	}
}
func (g *javaOriginalDiskGraph) commitChkpt() {
	g.tt.Helper()
	if err := g.disk.CommitChkpt(); err != nil {
		g.tt.Fatal(err)
	}
}
func (g *javaOriginalDiskGraph) recover() {
	g.tt.Helper()
	var err error
	if g.tableau != nil {
		err = g.tableau.Recover()
	} else {
		err = g.disk.Recover()
	}
	if err != nil {
		g.tt.Fatal(err)
	}
}
func (g *javaOriginalDiskGraph) size() int {
	if g.tableau != nil {
		return g.tableau.Size()
	}
	return g.disk.Size()
}
func (g *javaOriginalDiskGraph) addInitNode(fp uint64, tidx int) {
	if g.tableau != nil {
		g.tableau.AddInitNode(fp, tidx)
	} else {
		g.disk.AddInitNode(fp, tidx)
	}
}
func (g *javaOriginalDiskGraph) createCache()                   { g.disk.CreateCache() }
func (g *javaOriginalDiskGraph) destroyCache()                  { g.disk.DestroyCache() }
func (g *javaOriginalDiskGraph) isDone(fp uint64) bool          { return g.tableau.IsDone(fp) }
func (g *javaOriginalDiskGraph) setDone(fp uint64)              { g.tableau.SetDone(fp) }
func (g *javaOriginalDiskGraph) recordNode(fp uint64, tidx int) { g.tableau.RecordNode(fp, tidx) }

// Source catch(RuntimeException) must not absorb an unexpected IOException.
func javaOriginalGraphCatchRuntime(tt *testing.T, err error) {
	tt.Helper()
	if _, ok := err.(*RuntimeException); !ok {
		tt.Fatal(err)
	}
}

// Original DiskGraphTest.testGetPathWithoutInitNoTableau.
func TestJavaDiskGraphGetPathWithoutInitNoTableau(tt *testing.T) {
	dg := javaOriginalGraph(tt, false)
	tidx := -1
	dg.addNode(NewGraphNode(1, tidx))
	dg.createCache()
	if _, err := dg.getPathRaw(1, -1); err != nil {
		javaOriginalGraphCatchRuntime(tt, err)
		return
	}
	tt.Fatal("getPath() without init nodes has to throw a RuntimeException")
}

// Original DiskGraphTest.testGetMinimalPathWithoutTableau.
func TestJavaDiskGraphGetMinimalPathWithoutTableau(tt *testing.T) {
	dg := javaOriginalGraph(tt, false)
	tidx := -1
	initFP := uint64(1)
	successorFP := uint64(2)
	dg.addInitNode(1, tidx)
	dg.addNode(NewGraphNode(successorFP, tidx))
	node := NewGraphNode(initFP, tidx)
	node.AddTransition(successorFP, tidx, 1, 0, nil, 0, 0)
	dg.addNode(node)
	dg.createCache()
	path := dg.getPath(successorFP, tidx)
	dg.destroyCache()
	if path.Size() < 2 {
		tt.Fatal("original assertion 1 failed")
	}
	if path.Size() > 2 {
		tt.Fatal("original assertion 2 failed")
	}
}

// Original DiskGraphTest.testPathWithTwoInitNodes.
func TestJavaDiskGraphPathWithTwoInitNodes(tt *testing.T) {
	dg := javaOriginalGraph(tt, false)
	tidx := -1
	noSuccessorInitState := uint64(1)
	regularInitState := uint64(2)
	finalState := uint64(3)
	dg.addInitNode(noSuccessorInitState, tidx)
	dg.addInitNode(regularInitState, tidx)
	node := NewGraphNode(regularInitState, tidx)
	node.AddTransition(finalState, tidx, 1, 0, nil, 0, 0)
	dg.addNode(node)
	node = NewGraphNode(finalState, tidx)
	dg.addNode(node)
	dg.createCache()
	path := dg.getPath(finalState, tidx)
	dg.destroyCache()
	if int64(path.Size()) != int64(2) {
		tt.Fatal("original assertion 1 failed")
	}
	if int64(path.ElementAt(0)) != int64(finalState) {
		tt.Fatal("original assertion 2 failed")
	}
	if int64(path.ElementAt(1)) != int64(regularInitState) {
		tt.Fatal("original assertion 3 failed")
	}
	dg.createCache()
	path = dg.getPath(noSuccessorInitState, tidx)
	dg.destroyCache()
	if int64(path.Size()) != int64(1) {
		tt.Fatal("original assertion 4 failed")
	}
	if int64(path.ElementAt(0)) != int64(noSuccessorInitState) {
		tt.Fatal("original assertion 5 failed")
	}
}

// Original DiskGraphTest.testAddSameGraphNodeTwice.
func TestJavaDiskGraphAddSameGraphNodeTwice(tt *testing.T) {
	dg := javaOriginalGraph(tt, false)
	dg.addNode(NewGraphNode(1, 1))
	dg.addNode(NewGraphNode(1, 1))
	if int64(dg.size()) != int64(1) {
		tt.Fatal("original assertion 1 failed")
	}
}

// Original DiskGraphTest.testLookupExistingNode.
func TestJavaDiskGraphLookupExistingNode(tt *testing.T) {
	dg := javaOriginalGraph(tt, false)
	tidx := -1
	node := dg.getNode(1, tidx)
	if int64(node.SuccSize()) != int64(0) {
		tt.Fatal("original assertion 1 failed")
	}
	dg.addNode(node)
	dg.makeNodePtrTbl()
	node = dg.getNode(1, tidx)
	dg.addNode(node)
	if int64(node.SuccSize()) != int64(0) {
		tt.Fatal("original assertion 2 failed")
	}
	node.AddTransition(2, tidx, 1, 0, nil, 0, 0)
	dg.addNode(node)
	if int64(node.SuccSize()) != int64(1) {
		tt.Fatal("original assertion 3 failed")
	}
	if !node.TransExists(2, tidx) {
		tt.Fatal("original assertion 4 failed")
	}
	dg.makeNodePtrTbl()
	node = dg.getNode(1, tidx)
	if int64(node.SuccSize()) != int64(1) {
		tt.Fatal("original assertion 5 failed")
	}
	node.AddTransition(3, tidx, 1, 0, nil, 0, 0)
	dg.addNode(node)
	if int64(node.SuccSize()) != int64(2) {
		tt.Fatal("original assertion 6 failed")
	}
	if !node.TransExists(2, tidx) {
		tt.Fatal("original assertion 7 failed")
	}
	if !node.TransExists(3, tidx) {
		tt.Fatal("original assertion 8 failed")
	}
	dg.beginChkpt()
	dg.commitChkpt()
	dg.recover()
	node = dg.getNode(1, tidx)
	if int64(node.SuccSize()) != int64(2) {
		tt.Fatal("original assertion 9 failed")
	}
	if !node.TransExists(2, tidx) {
		tt.Fatal("original assertion 10 failed")
	}
	if !node.TransExists(3, tidx) {
		tt.Fatal("original assertion 11 failed")
	}
}

// Original DiskGraphTest.testAddSameGraphNodeTwiceCorrectSuccessors.
func TestJavaDiskGraphAddSameGraphNodeTwiceCorrectSuccessors(tt *testing.T) {
	dg := javaOriginalGraph(tt, false)
	tidx := -1
	graphNode := dg.getNode(1, tidx)
	graphNode.AddTransition(2, tidx, 1, 0, nil, 0, 0)
	firstPtr := dg.addNode(graphNode)
	graphNodeSecondInstance := dg.getNode(1, tidx)
	graphNodeSecondInstance.AddTransition(3, tidx, 1, 0, nil, 0, 0)
	secondPtr := dg.addNode(graphNodeSecondInstance)
	if int64(dg.size()) != int64(1) {
		tt.Fatal("original assertion 1 failed")
	}
	// Source assertNotSame boxes these small offsets through Java Long
	// caching, so object identity here is equivalent to numeric equality.
	if secondPtr == firstPtr {
		tt.Fatal("original pointer assertion failed")
	}
	if int64(dg.getLink(1, tidx)) != int64(secondPtr) {
		tt.Fatal("original assertion 3 failed")
	}
	node := dg.getNode(1, tidx)
	if int64(node.SuccSize()) != int64(2) {
		tt.Fatal("original assertion 4 failed")
	}
	if !node.TransExists(2, tidx) {
		tt.Fatal("original assertion 5 failed")
	}
	if !node.TransExists(3, tidx) {
		tt.Fatal("original assertion 6 failed")
	}
	dg.makeNodePtrTbl()
	dg.createCache()
	ptr := dg.getLink(1, tidx)
	n := dg.getNode(1, tidx, ptr)
	if int64(n.SuccSize()) != int64(2) {
		tt.Fatal("original assertion 7 failed")
	}
	if !n.TransExists(2, tidx) {
		tt.Fatal("original assertion 8 failed")
	}
	if !n.TransExists(3, tidx) {
		tt.Fatal("original assertion 9 failed")
	}
}

// Original DiskGraphTest.testGetPathPartialGraph.
func TestJavaDiskGraphGetPathPartialGraph(tt *testing.T) {
	dg := javaOriginalGraph(tt, false)
	tidx := -1
	initState := uint64(2)
	danglingState := uint64(3)
	dg.addInitNode(initState, tidx)
	node := NewGraphNode(initState, tidx)
	node.AddTransition(danglingState, tidx, 1, 0, nil, 0, 0)
	dg.addNode(node)
	dg.createCache()
	if _, err := dg.getPathRaw(5, tidx); err != nil {
		javaOriginalGraphCatchRuntime(tt, err)
		if err.Error() != fmt.Sprintf("Couldn't re-create liveness trace (path) starting at: 5 and tidx: %d", tidx) {
			tt.Fatal(err)
		}
	}
}

// Original DiskGraphTest.testGetPathWithoutInitNoTableau inherited by TableauDiskGraphTest.
func TestJavaTableauDiskGraphGetPathWithoutInitNoTableau(tt *testing.T) {
	dg := javaOriginalGraph(tt, true)
	tidx := 0
	dg.addNode(NewGraphNode(1, tidx))
	dg.createCache()
	if _, err := dg.getPathRaw(1, -1); err != nil {
		javaOriginalGraphCatchRuntime(tt, err)
		return
	}
	tt.Fatal("getPath() without init nodes has to throw a RuntimeException")
}

// Original DiskGraphTest.testGetMinimalPathWithoutTableau inherited by TableauDiskGraphTest.
func TestJavaTableauDiskGraphGetMinimalPathWithoutTableau(tt *testing.T) {
	dg := javaOriginalGraph(tt, true)
	tidx := 0
	initFP := uint64(1)
	successorFP := uint64(2)
	dg.addInitNode(1, tidx)
	dg.addNode(NewGraphNode(successorFP, tidx))
	node := NewGraphNode(initFP, tidx)
	node.AddTransition(successorFP, tidx, 1, 0, nil, 0, 0)
	dg.addNode(node)
	dg.createCache()
	path := dg.getPath(successorFP, tidx)
	dg.destroyCache()
	if path.Size() < 2 {
		tt.Fatal("original assertion 1 failed")
	}
	if path.Size() > 2 {
		tt.Fatal("original assertion 2 failed")
	}
}

// Original DiskGraphTest.testPathWithTwoInitNodes inherited by TableauDiskGraphTest.
func TestJavaTableauDiskGraphPathWithTwoInitNodes(tt *testing.T) {
	dg := javaOriginalGraph(tt, true)
	tidx := 0
	noSuccessorInitState := uint64(1)
	regularInitState := uint64(2)
	finalState := uint64(3)
	dg.addInitNode(noSuccessorInitState, tidx)
	dg.addInitNode(regularInitState, tidx)
	node := NewGraphNode(regularInitState, tidx)
	node.AddTransition(finalState, tidx, 1, 0, nil, 0, 0)
	dg.addNode(node)
	node = NewGraphNode(finalState, tidx)
	dg.addNode(node)
	dg.createCache()
	path := dg.getPath(finalState, tidx)
	dg.destroyCache()
	if int64(path.Size()) != int64(2) {
		tt.Fatal("original assertion 1 failed")
	}
	if int64(path.ElementAt(0)) != int64(finalState) {
		tt.Fatal("original assertion 2 failed")
	}
	if int64(path.ElementAt(1)) != int64(regularInitState) {
		tt.Fatal("original assertion 3 failed")
	}
	dg.createCache()
	path = dg.getPath(noSuccessorInitState, tidx)
	dg.destroyCache()
	if int64(path.Size()) != int64(1) {
		tt.Fatal("original assertion 4 failed")
	}
	if int64(path.ElementAt(0)) != int64(noSuccessorInitState) {
		tt.Fatal("original assertion 5 failed")
	}
}

// Original DiskGraphTest.testAddSameGraphNodeTwice inherited by TableauDiskGraphTest.
func TestJavaTableauDiskGraphAddSameGraphNodeTwice(tt *testing.T) {
	dg := javaOriginalGraph(tt, true)
	dg.addNode(NewGraphNode(1, 1))
	dg.addNode(NewGraphNode(1, 1))
	if int64(dg.size()) != int64(1) {
		tt.Fatal("original assertion 1 failed")
	}
}

// Original DiskGraphTest.testLookupExistingNode inherited by TableauDiskGraphTest.
func TestJavaTableauDiskGraphLookupExistingNode(tt *testing.T) {
	dg := javaOriginalGraph(tt, true)
	tidx := 0
	node := dg.getNode(1, tidx)
	if int64(node.SuccSize()) != int64(0) {
		tt.Fatal("original assertion 1 failed")
	}
	dg.addNode(node)
	dg.makeNodePtrTbl()
	node = dg.getNode(1, tidx)
	dg.addNode(node)
	if int64(node.SuccSize()) != int64(0) {
		tt.Fatal("original assertion 2 failed")
	}
	node.AddTransition(2, tidx, 1, 0, nil, 0, 0)
	dg.addNode(node)
	if int64(node.SuccSize()) != int64(1) {
		tt.Fatal("original assertion 3 failed")
	}
	if !node.TransExists(2, tidx) {
		tt.Fatal("original assertion 4 failed")
	}
	dg.makeNodePtrTbl()
	node = dg.getNode(1, tidx)
	if int64(node.SuccSize()) != int64(1) {
		tt.Fatal("original assertion 5 failed")
	}
	node.AddTransition(3, tidx, 1, 0, nil, 0, 0)
	dg.addNode(node)
	if int64(node.SuccSize()) != int64(2) {
		tt.Fatal("original assertion 6 failed")
	}
	if !node.TransExists(2, tidx) {
		tt.Fatal("original assertion 7 failed")
	}
	if !node.TransExists(3, tidx) {
		tt.Fatal("original assertion 8 failed")
	}
	dg.beginChkpt()
	dg.commitChkpt()
	dg.recover()
	node = dg.getNode(1, tidx)
	if int64(node.SuccSize()) != int64(2) {
		tt.Fatal("original assertion 9 failed")
	}
	if !node.TransExists(2, tidx) {
		tt.Fatal("original assertion 10 failed")
	}
	if !node.TransExists(3, tidx) {
		tt.Fatal("original assertion 11 failed")
	}
}

// Original DiskGraphTest.testAddSameGraphNodeTwiceCorrectSuccessors inherited by TableauDiskGraphTest.
func TestJavaTableauDiskGraphAddSameGraphNodeTwiceCorrectSuccessors(tt *testing.T) {
	dg := javaOriginalGraph(tt, true)
	tidx := 0
	graphNode := dg.getNode(1, tidx)
	graphNode.AddTransition(2, tidx, 1, 0, nil, 0, 0)
	firstPtr := dg.addNode(graphNode)
	graphNodeSecondInstance := dg.getNode(1, tidx)
	graphNodeSecondInstance.AddTransition(3, tidx, 1, 0, nil, 0, 0)
	secondPtr := dg.addNode(graphNodeSecondInstance)
	if int64(dg.size()) != int64(1) {
		tt.Fatal("original assertion 1 failed")
	}
	// Source assertNotSame boxes these small offsets through Java Long
	// caching, so object identity here is equivalent to numeric equality.
	if secondPtr == firstPtr {
		tt.Fatal("original pointer assertion failed")
	}
	if int64(dg.getLink(1, tidx)) != int64(secondPtr) {
		tt.Fatal("original assertion 3 failed")
	}
	node := dg.getNode(1, tidx)
	if int64(node.SuccSize()) != int64(2) {
		tt.Fatal("original assertion 4 failed")
	}
	if !node.TransExists(2, tidx) {
		tt.Fatal("original assertion 5 failed")
	}
	if !node.TransExists(3, tidx) {
		tt.Fatal("original assertion 6 failed")
	}
	dg.makeNodePtrTbl()
	dg.createCache()
	ptr := dg.getLink(1, tidx)
	n := dg.getNode(1, tidx, ptr)
	if int64(n.SuccSize()) != int64(2) {
		tt.Fatal("original assertion 7 failed")
	}
	if !n.TransExists(2, tidx) {
		tt.Fatal("original assertion 8 failed")
	}
	if !n.TransExists(3, tidx) {
		tt.Fatal("original assertion 9 failed")
	}
}

// Original DiskGraphTest.testGetPathPartialGraph inherited by TableauDiskGraphTest.
func TestJavaTableauDiskGraphGetPathPartialGraph(tt *testing.T) {
	dg := javaOriginalGraph(tt, true)
	tidx := 0
	initState := uint64(2)
	danglingState := uint64(3)
	dg.addInitNode(initState, tidx)
	node := NewGraphNode(initState, tidx)
	node.AddTransition(danglingState, tidx, 1, 0, nil, 0, 0)
	dg.addNode(node)
	dg.createCache()
	if _, err := dg.getPathRaw(5, tidx); err != nil {
		javaOriginalGraphCatchRuntime(tt, err)
		if err.Error() != fmt.Sprintf("Couldn't re-create liveness trace (path) starting at: 5 and tidx: %d", tidx) {
			tt.Fatal(err)
		}
	}
}

// Original TableauDiskGraphTest.testGetShortestPath.
func TestJavaTableauDiskGraphGetShortestPath(tt *testing.T) {
	dg := javaOriginalGraph(tt, true)
	initState := uint64(1)
	initTableauIdx := 0
	secondState := uint64(2)
	thirdState := uint64(3)
	thirdTableauIdx := 0
	finalState := uint64(4)
	finalTableauIdx := 0
	dg.addInitNode(initState, initTableauIdx)
	dg.addInitNode(secondState, 1)
	node := NewGraphNode(initState, initTableauIdx)
	node.AddTransition(secondState, 0, 1, 0, nil, 0, 1)
	node.AddTransition(secondState, 2, 1, 0, nil, 0, 1)
	dg.addNode(node)
	node = NewGraphNode(secondState, 0)
	node.AddTransition(thirdState, thirdTableauIdx, 1, 0, nil, 0, 1)
	dg.addNode(node)
	node = NewGraphNode(secondState, 1)
	node.AddTransition(thirdState, thirdTableauIdx, 1, 0, nil, 0, 1)
	dg.addNode(node)
	node = NewGraphNode(secondState, 2)
	node.AddTransition(thirdState, thirdTableauIdx, 1, 0, nil, 0, 1)
	dg.addNode(node)
	node = NewGraphNode(thirdState, thirdTableauIdx)
	node.AddTransition(finalState, finalTableauIdx, 1, 0, nil, 0, 1)
	dg.addNode(node)
	node = NewGraphNode(finalState, finalTableauIdx)
	dg.addNode(node)
	dg.createCache()
	path := dg.getPath(finalState, finalTableauIdx)
	dg.destroyCache()
	if int64(path.Size()) != int64(3) {
		tt.Fatal("original assertion 1 failed")
	}
	if int64(path.ElementAt(0)) != int64(finalState) {
		tt.Fatal("original assertion 2 failed")
	}
	if int64(path.ElementAt(1)) != int64(thirdState) {
		tt.Fatal("original assertion 3 failed")
	}
	if int64(path.ElementAt(2)) != int64(secondState) {
		tt.Fatal("original assertion 4 failed")
	}
}

// Original TableauDiskGraphTest.testUnifyingNodeInPath.
func TestJavaTableauDiskGraphUnifyingNodeInPath(tt *testing.T) {
	dg := javaOriginalGraph(tt, true)
	initState := uint64(1)
	initTableauIdx := 0
	secondState := uint64(2)
	secondTableauIdx := 0
	thirdState := uint64(3)
	thirdTableauIdx := 0
	finalState := uint64(4)
	finalTableauIdx := 0
	dg.addInitNode(initState, initTableauIdx)
	node := NewGraphNode(initState, initTableauIdx)
	node.AddTransition(secondState, secondTableauIdx, 1, 0, nil, 0, 1)
	node.AddTransition(thirdState, thirdTableauIdx, 1, 0, nil, 0, 1)
	dg.addNode(node)
	node = NewGraphNode(secondState, secondTableauIdx)
	node.AddTransition(finalState, finalTableauIdx, 1, 0, nil, 0, 1)
	dg.addNode(node)
	node = NewGraphNode(thirdState, thirdTableauIdx)
	node.AddTransition(finalState, finalTableauIdx, 1, 0, nil, 0, 1)
	dg.addNode(node)
	node = NewGraphNode(finalState, finalTableauIdx)
	dg.addNode(node)
	dg.createCache()
	path := dg.getPath(finalState, finalTableauIdx)
	dg.destroyCache()
	if int64(path.Size()) != int64(3) {
		tt.Fatal("original assertion 1 failed")
	}
	if int64(path.ElementAt(0)) != int64(finalState) {
		tt.Fatal("original assertion 2 failed")
	}
	if int64(path.ElementAt(1)) != int64(secondState) {
		tt.Fatal("original assertion 3 failed")
	}
	if int64(path.ElementAt(2)) != int64(initState) {
		tt.Fatal("original assertion 4 failed")
	}
}

// Original TableauDiskGraphTest.testUnifyingNodeShortestPath.
func TestJavaTableauDiskGraphUnifyingNodeShortestPath(tt *testing.T) {
	dg := javaOriginalGraph(tt, true)
	initState := uint64(1)
	initTableauIdx := 0
	frob := uint64(5)
	frobTableauIdx := 0
	secondState := uint64(2)
	secondTableauIdx := 0
	thirdState := uint64(3)
	thirdTableauIdx := 0
	finalState := uint64(4)
	finalTableauIdx := 0
	dg.addInitNode(initState, initTableauIdx)
	node := NewGraphNode(initState, initTableauIdx)
	node.AddTransition(frob, frobTableauIdx, 1, 0, nil, 0, 1)
	node.AddTransition(thirdState, thirdTableauIdx, 1, 0, nil, 0, 1)
	dg.addNode(node)
	node = NewGraphNode(frob, frobTableauIdx)
	node.AddTransition(secondState, secondTableauIdx, 1, 0, nil, 0, 1)
	dg.addNode(node)
	node = NewGraphNode(secondState, secondTableauIdx)
	node.AddTransition(finalState, finalTableauIdx, 1, 0, nil, 0, 1)
	dg.addNode(node)
	node = NewGraphNode(thirdState, thirdTableauIdx)
	node.AddTransition(finalState, finalTableauIdx, 1, 0, nil, 0, 1)
	dg.addNode(node)
	node = NewGraphNode(finalState, finalTableauIdx)
	dg.addNode(node)
	dg.createCache()
	path := dg.getPath(finalState, finalTableauIdx)
	dg.destroyCache()
	if int64(path.Size()) != int64(3) {
		tt.Fatal("original assertion 1 failed")
	}
	if int64(path.ElementAt(0)) != int64(finalState) {
		tt.Fatal("original assertion 2 failed")
	}
	if int64(path.ElementAt(1)) != int64(thirdState) {
		tt.Fatal("original assertion 3 failed")
	}
	if int64(path.ElementAt(2)) != int64(initState) {
		tt.Fatal("original assertion 4 failed")
	}
}

// Original TableauDiskGraphTest.testPathWithTwoInitNodesWithTableau.
func TestJavaTableauDiskGraphPathWithTwoInitNodesWithTableau(tt *testing.T) {
	dg := javaOriginalGraph(tt, true)
	noSuccessorInitState := uint64(1)
	regularInitState := uint64(2)
	finalState := uint64(3)
	dg.addInitNode(noSuccessorInitState, 0)
	dg.addInitNode(regularInitState, 0)
	node := NewGraphNode(regularInitState, 0)
	node.AddTransition(finalState, 0, 1, 0, nil, 0, 0)
	dg.addNode(node)
	node = NewGraphNode(finalState, 0)
	dg.addNode(node)
	dg.createCache()
	path := dg.getPath(finalState, 0)
	dg.destroyCache()
	if int64(path.Size()) != int64(2) {
		tt.Fatal("original assertion 1 failed")
	}
	if int64(path.ElementAt(0)) != int64(finalState) {
		tt.Fatal("original assertion 2 failed")
	}
	if int64(path.ElementAt(1)) != int64(regularInitState) {
		tt.Fatal("original assertion 3 failed")
	}
	dg.createCache()
	path = dg.getPath(noSuccessorInitState, 0)
	dg.destroyCache()
	if int64(path.Size()) != int64(1) {
		tt.Fatal("original assertion 4 failed")
	}
	if int64(path.ElementAt(0)) != int64(noSuccessorInitState) {
		tt.Fatal("original assertion 5 failed")
	}
}

// Original TableauDiskGraphTest.testGetPathWithTwoInits.
func TestJavaTableauDiskGraphGetPathWithTwoInits(tt *testing.T) {
	dg := javaOriginalGraph(tt, true)
	fingerprint := uint64(1)
	dg.addInitNode(fingerprint, 0)
	dg.addInitNode(fingerprint, 1)
	dg.createCache()
	if _, err := dg.getPathRaw(fingerprint, 2); err != nil {
		javaOriginalGraphCatchRuntime(tt, err)
		dg.addInitNode(fingerprint, 2)
		dg.createCache()
		path := dg.getPath(fingerprint, 2)
		dg.destroyCache()
		if path.Size() != 1 {
			tt.Fatal("original path size assertion failed")
		}
		if path.ElementAt(0) != int64(fingerprint) {
			tt.Fatal("original path fingerprint assertion failed")
		}
		return
	}
	dg.destroyCache()
	tt.Fatal("Returned path to non-existing node")
}

// Original TableauDiskGraphTest.testNodeSetDone.
func TestJavaTableauDiskGraphNodeSetDone(tt *testing.T) {
	dg := javaOriginalGraph(tt, true)
	fingerprint := uint64(1)
	dg.addInitNode(fingerprint, 0)
	if dg.isDone(fingerprint) {
		tt.Fatal("original assertion 1 failed")
	}
	node := NewGraphNode(fingerprint, 0)
	node.AddTransition(fingerprint, 1, 1, 0, nil, 0, 0)
	dg.addNode(node)
	if !dg.isDone(fingerprint) {
		tt.Fatal("original assertion 2 failed")
	}
}

// Original TableauDiskGraphTest.testGetPathWithTwoNodesWithSameFingerprint.
func TestJavaTableauDiskGraphGetPathWithTwoNodesWithSameFingerprint(tt *testing.T) {
	dg := javaOriginalGraph(tt, true)
	fingerprint := uint64(1)
	dg.addInitNode(fingerprint, 0)
	node := NewGraphNode(fingerprint, 0)
	node.AddTransition(fingerprint, 1, 1, 0, nil, 0, 0)
	dg.addNode(node)
	node = NewGraphNode(fingerprint, 1)
	dg.addNode(node)
	dg.createCache()
	path := dg.getPath(fingerprint, 1)
	dg.destroyCache()
	if int64(path.Size()) != int64(2) {
		tt.Fatal("original assertion 1 failed")
	}
	if int64(path.ElementAt(0)) != int64(fingerprint) {
		tt.Fatal("original assertion 2 failed")
	}
	if int64(path.ElementAt(1)) != int64(fingerprint) {
		tt.Fatal("original assertion 3 failed")
	}
}

// Original TableauDiskGraphTest.testLookupExistingNodeWithTidx.
func TestJavaTableauDiskGraphLookupExistingNodeWithTidx(tt *testing.T) {
	dg := javaOriginalGraph(tt, true)
	node := dg.getNode(1, 1)
	if int64(node.SuccSize()) != int64(0) {
		tt.Fatal("original assertion 1 failed")
	}
	dg.addNode(node)
	dg.makeNodePtrTbl()
	node = dg.getNode(1, 1)
	dg.addNode(node)
	if int64(node.SuccSize()) != int64(0) {
		tt.Fatal("original assertion 2 failed")
	}
	node.AddTransition(2, 2, 1, 0, nil, 0, 0)
	dg.addNode(node)
	if int64(node.SuccSize()) != int64(1) {
		tt.Fatal("original assertion 3 failed")
	}
	if !node.TransExists(2, 2) {
		tt.Fatal("original assertion 4 failed")
	}
	dg.makeNodePtrTbl()
	node = dg.getNode(1, 1)
	if int64(node.SuccSize()) != int64(1) {
		tt.Fatal("original assertion 5 failed")
	}
	node.AddTransition(3, 3, 1, 0, nil, 0, 0)
	dg.addNode(node)
	if int64(node.SuccSize()) != int64(2) {
		tt.Fatal("original assertion 6 failed")
	}
	if !node.TransExists(2, 2) {
		tt.Fatal("original assertion 7 failed")
	}
	if !node.TransExists(3, 3) {
		tt.Fatal("original assertion 8 failed")
	}
	dg.beginChkpt()
	dg.commitChkpt()
	dg.recover()
	node = dg.getNode(1, 1)
	if int64(node.SuccSize()) != int64(2) {
		tt.Fatal("original assertion 9 failed")
	}
	if !node.TransExists(2, 2) {
		tt.Fatal("original assertion 10 failed")
	}
	if !node.TransExists(3, 3) {
		tt.Fatal("original assertion 11 failed")
	}
}

// Original TableauDiskGraphTest.testWhatsDoneIsDoneRRS.
func TestJavaTableauDiskGraphWhatsDoneIsDoneRRS(tt *testing.T) {
	dg := javaOriginalGraph(tt, true)
	fp := uint64(42)
	t := 1
	u := 2
	dg.recordNode(fp, t)
	if dg.isDone(fp) {
		tt.Fatal("original assertion 1 failed")
	}
	dg.recordNode(fp, u)
	if dg.isDone(fp) {
		tt.Fatal("original assertion 2 failed")
	}
	dg.setDone(fp)
	if !dg.isDone(fp) {
		tt.Fatal("original assertion 3 failed")
	}
}

// Original TableauDiskGraphTest.testWhatsDoneIsDoneRSR.
func TestJavaTableauDiskGraphWhatsDoneIsDoneRSR(tt *testing.T) {
	dg := javaOriginalGraph(tt, true)
	fp := uint64(42)
	t := 1
	u := 2
	dg.recordNode(fp, t)
	if dg.isDone(fp) {
		tt.Fatal("original assertion 1 failed")
	}
	dg.setDone(fp)
	if !dg.isDone(fp) {
		tt.Fatal("original assertion 2 failed")
	}
	dg.recordNode(fp, u)
	if !dg.isDone(fp) {
		tt.Fatal("original assertion 3 failed")
	}
}

// Original TableauDiskGraphTest.testWhatsDoneIsDoneSRR.
func TestJavaTableauDiskGraphWhatsDoneIsDoneSRR(tt *testing.T) {
	dg := javaOriginalGraph(tt, true)
	fp := uint64(42)
	t := 1
	u := 2
	dg.setDone(fp)
	if !dg.isDone(fp) {
		tt.Fatal("original assertion 1 failed")
	}
	dg.recordNode(fp, t)
	if !dg.isDone(fp) {
		tt.Fatal("original assertion 2 failed")
	}
	dg.recordNode(fp, u)
	if !dg.isDone(fp) {
		tt.Fatal("original assertion 3 failed")
	}
}
