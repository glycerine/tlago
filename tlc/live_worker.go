package tlc

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
)

const liveWorkerSCCMarker = int64(-42)

var (
	liveWorkerStackMu sync.Mutex
	liveWorkerStatsMu sync.Mutex
	liveWorkerStats   = NewBucketStatisticsWithMX("Histogram SCC sizes", javaLivenessPackageName, "StronglyConnectedComponent sizes")
)

type LiveWorker struct {
	Tool       *Tool
	ID         int
	NumWorkers int
	Check      *LiveCheck
	Checker    *LiveChecker
	FinalCheck bool
	Solution   *OrderOfSolution
	PEM        *PossibleErrorModel

	ownersCaptured  bool
	toolOwner       *Tool
	idOwner         int
	numWorkersOwner int
	checkOwner      *LiveCheck
	finalCheckOwner bool
	checkerOwner    *LiveChecker
	solutionOwner   *OrderOfSolution
	pemOwner        *PossibleErrorModel
}

type liveIntStack struct {
	mem  *IntStack
	disk *SynchronousDiskIntStack
}

func (w *LiveWorker) require() {
	if w == nil {
		panic(NewNullPointerException())
	}
}

func NewLiveWorker(tool *Tool, id int, numWorkers int, liveCheck *LiveCheck, checker *LiveChecker, pem *PossibleErrorModel, finalCheck bool) *LiveWorker {
	id = int(int32(id))
	numWorkers = int(int32(numWorkers))
	worker := &LiveWorker{
		Tool:            tool,
		ID:              id,
		NumWorkers:      numWorkers,
		Check:           liveCheck,
		Checker:         checker,
		FinalCheck:      finalCheck,
		PEM:             pem,
		ownersCaptured:  true,
		toolOwner:       tool,
		idOwner:         id,
		numWorkersOwner: numWorkers,
		checkOwner:      liveCheck,
		finalCheckOwner: finalCheck,
		checkerOwner:    checker,
		pemOwner:        pem,
	}
	if checker != nil {
		worker.Solution = checker.ownedSolution()
		worker.solutionOwner = worker.Solution
	}
	return worker
}

func (w *LiveWorker) ownedTool() *Tool {
	if w.ownersCaptured {
		return w.toolOwner
	}
	return w.Tool
}

func (w *LiveWorker) ownedID() int {
	if w.ownersCaptured {
		return w.idOwner
	}
	return w.ID
}

func (w *LiveWorker) ownedNumWorkers() int {
	if w.ownersCaptured {
		return w.numWorkersOwner
	}
	return w.NumWorkers
}

func (w *LiveWorker) ownedLiveCheck() *LiveCheck {
	if w.ownersCaptured {
		return w.checkOwner
	}
	return w.Check
}

func (w *LiveWorker) ownedFinalCheck() bool {
	if w.ownersCaptured {
		return w.finalCheckOwner
	}
	return w.FinalCheck
}

func (w *LiveWorker) ownedChecker() *LiveChecker {
	if w.ownersCaptured {
		return w.checkerOwner
	}
	return w.Checker
}

func (w *LiveWorker) ownedSolution() *OrderOfSolution {
	if w.ownersCaptured {
		return w.solutionOwner
	}
	return w.Solution
}

func (w *LiveWorker) ownedPEM() *PossibleErrorModel {
	if w.ownersCaptured {
		return w.pemOwner
	}
	return w.PEM
}

func (w *LiveWorker) CheckSccs() (bool, error) {
	w.require()
	if w.ownedChecker() == nil {
		panic(NewNullPointerException())
	}
	if err := w.makeNodePtrTable(); err != nil {
		return false, err
	}

	initNodes := w.initNodes()
	numOfInits := initNodes.Size()
	nodeQueue := NewIntQueueWithDisk(w.liveCheckMetaDir(), "root", (numOfInits/2)*5)
	for j := 0; j < numOfInits; j += 2 {
		state := uint64(initNodes.ElementAt(j))
		tidx := int(initNodes.ElementAt(j + 1))
		ptr := w.getLink(state, tidx)
		if ptr >= 0 {
			if !IsDiskGraphFilePointer(ptr) {
				panic(NewAssertionError())
			}
			nodeQueue.EnqueueLong(int64(state))
			nodeQueue.EnqueueInt(int32(tidx))
			nodeQueue.EnqueueLong(ptr)
		}
	}

	if w.ownedPEM() == nil {
		panic(NewNullPointerException())
	}
	eaaction := w.ownedPEM().ownedEAAction()
	if w.ownedSolution() == nil {
		panic(NewNullPointerException())
	}
	slen := len(w.ownedSolution().GetCheckState())
	alen := len(w.ownedSolution().GetCheckAction())
	liveWorkerStackMu.Lock()
	dfsStack, err := w.getStack("dfs")
	if err != nil {
		liveWorkerStackMu.Unlock()
		return false, err
	}
	comStack, err := w.getStack("com")
	liveWorkerStackMu.Unlock()
	if err != nil {
		return false, err
	}

	for nodeQueue.Size() > 0 {
		state := uint64(nodeQueue.DequeueLong())
		tidx := int(nodeQueue.DequeueInt())
		loc := nodeQueue.DequeueLong()
		dfsStack.Reset()
		dfsStack.PushLong(int64(state))
		dfsStack.PushInt(int32(tidx))
		dfsStack.PushLong(loc)
		dfsStack.PushLong(DiskGraphMaxPtr)
		newLink := DiskGraphMaxPtr

		for dfsStack.Size() >= 7 {
			lowLink := dfsStack.PopLong()
			curLoc := dfsStack.PopLong()
			curTidx := int(dfsStack.PopInt())
			curState := uint64(dfsStack.PopLong())
			if !IsDiskGraphFilePointer(curLoc) {
				panic(NewAssertionError())
			}

			if curLoc == liveWorkerSCCMarker {
				curLink := w.getLink(curState, curTidx)
				if curLink >= DiskGraphMaxLink {
					panic(NewAssertionError())
				}
				if curLink == lowLink {
					ok, err := w.checkComponent(curState, curTidx, comStack)
					if err != nil || !ok {
						return !ok, err
					}
				}
				plowLink := dfsStack.PopLong()
				dfsStack.PushLong(minInt64(plowLink, lowLink))
				continue
			}

			link := w.putLink(curState, curTidx, newLink)
			if link == -1 {
				dfsStack.PushLong(lowLink)
				dfsStack.PushLong(int64(curState))
				dfsStack.PushInt(int32(curTidx))
				dfsStack.PushLong(liveWorkerSCCMarker)

				comStack.PushLong(curLoc)
				comStack.PushInt(int32(curTidx))
				comStack.PushLong(int64(curState))

				gnode, err := w.getNode(curState, curTidx, curLoc)
				if err != nil {
					return false, err
				}
				nextLowLink := newLink
				newLink++
				for i := 0; i < gnode.SuccSize(); i++ {
					nextState := gnode.GetStateFP(i)
					nextTidx := gnode.GetTIndex(i)
					nextLink := w.getLink(nextState, nextTidx)
					if nextLink >= 0 {
						if gnode.GetCheckActionAll(slen, alen, i, eaaction) {
							if IsDiskGraphFilePointer(nextLink) {
								dfsStack.PushLong(int64(nextState))
								dfsStack.PushInt(int32(nextTidx))
								dfsStack.PushLong(nextLink)
							} else {
								nextLowLink = minInt64(nextLowLink, nextLink)
							}
						} else if IsDiskGraphFilePointer(nextLink) {
							nodeQueue.EnqueueLong(int64(nextState))
							nodeQueue.EnqueueInt(int32(nextTidx))
							nodeQueue.EnqueueLong(nextLink)
						}
					}
				}
				dfsStack.PushLong(nextLowLink)
			} else {
				if link < DiskGraphMaxPtr || link > DiskGraphMaxLink {
					panic(NewAssertionError())
				}
				dfsStack.PushLong(minInt64(lowLink, link))
			}
		}
	}
	return false, nil
}

func (w *LiveWorker) ToDotViz(state uint64, tidx int, nodes *TableauNodePtrTable) (string, error) {
	w.require()
	if nodes == nil {
		panic(NewNullPointerException())
	}
	tidx = graphInt(tidx)
	var b strings.Builder
	b.Grow(nodes.Size() * 10)
	b.WriteString("digraph TableauNodePtrTable {\n")
	b.WriteString("nodesep = 0.7\n")
	b.WriteString("rankdir=LR;\n")
	for bucket := 0; bucket < nodes.GetSize(); bucket++ {
		record := nodes.GetNodesByLoc(bucket)
		if record == nil {
			continue
		}
		state1 := TableauGetKey(record)
		for loc := 2; loc < len(record); loc += nodes.GetElemLength() {
			tidx1 := TableauGetTidx(record, loc)
			nodeLoc := TableauGetElem(record, loc)
			node, err := w.getNode(state1, tidx1, nodeLoc)
			if err != nil {
				return "", err
			}
			slen := len(w.ownedSolution().GetCheckState())
			alen := len(w.ownedSolution().GetCheckAction())
			b.WriteString(node.ToDotVizWithFilter(state1 == state && tidx1 == tidx, true, slen, alen, nodes, w.ownedSolution()))
		}
	}
	b.WriteString("}")
	return b.String(), nil
}

func (w *LiveWorker) WriteDotViz(state uint64, tidx int, nodes *TableauNodePtrTable, path string) {
	w.require()
	if nodes == nil {
		panic(NewNullPointerException())
	}
	if nodes.Size() <= 1 {
		return
	}
	file, err := openFileOutputStream(path)
	if err != nil {
		_, _ = fmt.Fprint(os.Stderr, javaThrowableStackTrace(err))
		return
	}
	writer := NewJavaBufferedWriter(file)
	rendered, err := w.ToDotViz(state, tidx, nodes)
	if err == nil {
		_, err = writer.WriteString(rendered)
		err = bufferedRandomAccessFileIOError(err)
	}
	if err == nil {
		err = bufferedRandomAccessFileIOError(writer.Flush())
	}
	if err == nil {
		err = bufferedRandomAccessFileIOError(file.Close())
	}
	if err != nil {
		_, _ = fmt.Fprint(os.Stderr, javaThrowableStackTrace(err))
	}
}

func LiveWorkerDetailedFormatter(stack *IntStack) string {
	if stack == nil {
		panic(NewNullPointerException())
	}
	size := stack.Size()
	var b strings.Builder
	b.Grow(size / 5)
	for i := 0; i < size; i += 5 {
		loc := stack.PeakLong(size - i - 5)
		tidx := stack.PeakInt(size - i - 3)
		state := stack.PeakLong(size - i - 2)
		fmt.Fprintf(&b, "state: %d tidx: %d loc: %d\n", state, tidx, loc)
	}
	return b.String()
}

func LiveWorkerDFSStackDetailedFormatter(stack *IntStack) string {
	if stack == nil {
		panic(NewNullPointerException())
	}
	size := stack.Size()
	var b strings.Builder
	b.Grow(size / 7)
	i := 0
	for i < size {
		top := stack.PeakLong(size - i - 2)
		switch {
		case top == liveWorkerSCCMarker:
			state := stack.PeakLong(size - i - 5)
			tidx := stack.PeakInt(size - i - 3)
			lowLink := stack.PeakLong(size-i-7) - DiskGraphMaxPtr
			fmt.Fprintf(&b, "node [ fp: %d tidx: %d lowLink: %d]\n", state, tidx, lowLink)
			i += 7
		case IsDiskGraphFilePointer(top):
			state := stack.PeakLong(size - i - 5)
			tidx := stack.PeakInt(size - i - 3)
			fmt.Fprintf(&b, "succ [ fp: %d tidx: %d location: %d]\n", state, tidx, top)
			i += 5
		default:
			fmt.Fprintf(&b, "pLowLink: %d\n", top-DiskGraphMaxPtr)
			i += 2
		}
	}
	if i != size {
		panic(NewAssertionError())
	}
	return b.String()
}

func (w *LiveWorker) liveCheckMetaDir() string {
	w.require()
	check := w.ownedLiveCheck()
	if check == nil {
		panic(NewNullPointerException())
	}
	return check.GetMetaDir()
}

func (w *LiveWorker) getStack(name string) (*liveIntStack, error) {
	numWorkers := w.ownedNumWorkers()
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	freeMemoryInBytes := float64(uint64(1))
	if mem.HeapSys > mem.HeapAlloc {
		freeMemoryInBytes = liveWorkerFreeMemoryPerWorker(mem.HeapSys-mem.HeapAlloc, numWorkers)
	}
	graphSizeInBytes, err := w.graphSizeOnDisk()
	if err != nil {
		return nil, err
	}
	ratio := float64(graphSizeInBytes) / freeMemoryInBytes
	if ratio > Globals.LivenessGraphSizeThreshold {
		capacity := liveWorkerStackCapacity(ratio)
		if float64(capacity) < freeMemoryInBytes {
			return newDiskLiveIntStack(w.liveCheckMetaDir(), fmt.Sprintf("%s%d", name, w.ownedID()), capacity), nil
		}
		return newDiskLiveIntStack(w.liveCheckMetaDir(), fmt.Sprintf("%s%d", name, w.ownedID())), nil
	}
	return newMemLiveIntStack(), nil
}

func liveWorkerFreeMemoryPerWorker(freeMemory uint64, numWorkers int) float64 {
	return float64(freeMemory) / float64(numWorkers)
}

func liveWorkerStackCapacity(ratio float64) int {
	shift := javaDoubleToInt(ratio)
	if shift > 5 {
		shift = 5
	}
	return int(int32(uint32(SynchronousDiskIntStackBufSize) << (uint32(shift) & 31)))
}

func newMemLiveIntStack() *liveIntStack {
	return &liveIntStack{mem: NewIntStack()}
}

func newDiskLiveIntStack(metadir string, name string, capacity ...int) *liveIntStack {
	return &liveIntStack{disk: NewSynchronousDiskIntStack(metadir, name, capacity...)}
}

func (s *liveIntStack) require() {
	if s == nil {
		panic(NewNullPointerException())
	}
}

func (s *liveIntStack) Size() int64 {
	s.require()
	if s.disk != nil {
		return s.disk.Size()
	}
	return int64(s.mem.Size())
}

func (s *liveIntStack) PushInt(x int32) {
	s.require()
	if s.disk != nil {
		s.disk.PushInt(x)
		return
	}
	s.mem.PushInt(x)
}

func (s *liveIntStack) PushLong(x int64) {
	s.require()
	if s.disk != nil {
		s.disk.PushLong(x)
		return
	}
	s.mem.PushLong(x)
}

func (s *liveIntStack) PopInt() int32 {
	s.require()
	if s.disk != nil {
		return s.disk.PopInt()
	}
	return s.mem.PopInt()
}

func (s *liveIntStack) PopLong() int64 {
	s.require()
	if s.disk != nil {
		return s.disk.PopLong()
	}
	return s.mem.PopLong()
}

func (s *liveIntStack) Reset() {
	s.require()
	if s.disk != nil {
		s.disk.Reset()
		return
	}
	s.mem.Reset()
}

func (w *LiveWorker) graphSizeOnDisk() (int64, error) {
	w.require()
	if w.ownedChecker() == nil {
		panic(NewNullPointerException())
	}
	if w.ownedChecker().tableau {
		dgraph := w.ownedChecker().ownedTableauDiskGraph()
		if dgraph == nil {
			panic(NewNullPointerException())
		}
		return dgraph.GetSizeOnDisk()
	}
	dgraph := w.ownedChecker().ownedDiskGraph()
	if dgraph == nil {
		panic(NewNullPointerException())
	}
	return dgraph.GetSizeOnDisk()
}

func (w *LiveWorker) checkComponent(state uint64, tidx int, comStack *liveIntStack) (bool, error) {
	comStackSize := comStack.Size()
	if comStackSize < 5 || comStackSize%5 != 0 {
		panic(NewAssertionError())
	}
	state1 := uint64(comStack.PopLong())
	tidx1 := int(comStack.PopInt())
	loc1 := comStack.PopLong()

	if state1 == state && tidx1 == tidx {
		stuttering, err := w.isStuttering(state1, tidx1, loc1)
		if err != nil {
			return false, err
		}
		if !stuttering {
			w.setMaxLink(state, tidx)
			return true, nil
		}
	}

	com := NewTableauNodePtrTable(128)
	for {
		com.PutElem(state1, tidx1, loc1)
		if !IsDiskGraphFilePointer(loc1) {
			panic(NewAssertionError())
		}
		w.setMaxLink(state1, tidx1)
		if state == state1 && tidx == tidx1 {
			break
		}
		state1 = uint64(comStack.PopLong())
		tidx1 = int(comStack.PopInt())
		loc1 = comStack.PopLong()
	}
	if int64(com.Size()) > comStackSize/5 {
		panic(NewAssertionError())
	}
	liveWorkerAddSCCSample(com.Size())

	slen := len(w.ownedSolution().GetCheckState())
	alen := len(w.ownedSolution().GetCheckAction())
	aeslen := len(w.ownedPEM().ownedAEState())
	aealen := len(w.ownedPEM().ownedAEAction())
	plen := len(w.ownedSolution().GetPromises())
	aeStateRes := make([]bool, aeslen)
	aeActionRes := make([]bool, aealen)
	promiseRes := make([]bool, plen)
	eaaction := w.ownedPEM().ownedEAAction()

	for ci := 0; ci < com.GetSize(); ci++ {
		nodes := com.GetNodesByLoc(ci)
		if nodes == nil {
			continue
		}
		state1 = TableauGetKey(nodes)
		for nidx := 2; nidx < len(nodes); nidx += com.GetElemLength() {
			tidx1 = TableauGetTidx(nodes, nidx)
			loc1 = TableauGetElem(nodes, nidx)
			curNode, err := w.getNode(state1, tidx1, loc1)
			if err != nil {
				return false, err
			}
			for i, idx := range w.ownedPEM().ownedAEState() {
				if !aeStateRes[i] {
					aeStateRes[i] = curNode.GetCheckState(idx)
				}
			}
			if aealen > 0 {
				for i := 0; i < curNode.SuccSize(); i++ {
					nextState := curNode.GetStateFP(i)
					nextTidx := curNode.GetTIndex(i)
					if com.GetLoc(nextState, nextTidx) == -1 {
						continue
					}
					if !curNode.GetCheckActionAll(slen, alen, i, eaaction) {
						continue
					}
					for j, idx := range w.ownedPEM().ownedAEAction() {
						if !aeActionRes[j] {
							aeActionRes[j] = curNode.GetCheckAction(slen, alen, i, idx)
						}
					}
				}
			}
			for i, promise := range w.ownedSolution().GetPromises() {
				par := curNode.GetTNode(w.ownedSolution().GetTableau()).GetPar()
				if par.IsFulfilling(promise) {
					promiseRes[i] = true
				}
			}
		}
	}

	for _, ok := range aeStateRes {
		if !ok {
			return true, nil
		}
	}
	for _, ok := range aeActionRes {
		if !ok {
			return true, nil
		}
	}
	for _, ok := range promiseRes {
		if !ok {
			return true, nil
		}
	}
	if check := w.ownedLiveCheck(); check != nil && !check.claimLiveError(w.ownedID()) {
		return false, nil
	}
	w.ownedChecker().ErrorGraphNode = NewGraphNode(state, tidx)
	prefixFuture := w.startLivePrefixRecovery(state, tidx)
	cycle, err := w.traceFingerprintCycle(state, tidx, com)
	if err != nil {
		return false, err
	}
	prefix, err := awaitLivePrefixRecovery(prefixFuture)
	if err != nil {
		return false, err
	}
	w.ownedChecker().ErrorPrefix = prefix.fingerprints
	w.ownedChecker().ErrorCycle = cycle
	counterExample, printableTrace, rawTrace, closingInfo, loopOrdinal, cycleIndex, stuttering, err := w.buildCounterExample(prefix, cycle)
	if err != nil {
		return false, err
	}
	w.ownedChecker().ErrorTrace = printableTrace
	w.ownedChecker().ErrorRawTrace = rawTrace
	w.ownedChecker().ErrorCounterEx = counterExample
	w.ownedChecker().ErrorClosingInfo = closingInfo
	w.ownedChecker().ErrorLoopOrdinal = loopOrdinal
	w.ownedChecker().ErrorCycleIndex = cycleIndex
	w.ownedChecker().ErrorStuttering = stuttering
	return false, nil
}

func liveWorkerAddSCCSample(size int) {
	liveWorkerStatsMu.Lock()
	defer liveWorkerStatsMu.Unlock()
	liveWorkerStats.AddSample(size)
}

func liveWorkerStatsSnapshot() (string, int64) {
	liveWorkerStatsMu.Lock()
	defer liveWorkerStatsMu.Unlock()
	return liveWorkerStats.String(), liveWorkerStats.Observations()
}

func (w *LiveWorker) isStuttering(state uint64, tidx int, loc int64) (bool, error) {
	slen := len(w.ownedSolution().GetCheckState())
	alen := len(w.ownedSolution().GetCheckAction())
	gnode, err := w.getNode(state, tidx, loc)
	if err != nil {
		return false, err
	}
	for i := 0; i < gnode.SuccSize(); i++ {
		if gnode.GetStateFP(i) == state && gnode.GetTIndex(i) == tidx {
			return gnode.GetCheckActionAll(slen, alen, i, w.ownedPEM().ownedEAAction()), nil
		}
	}
	return false, nil
}

type livePrefixTrace struct {
	fingerprints *LongVec
	states       []*TLCStateInfo
	printable    []*TLCStateInfo
}

type livePrefixFutureResult struct {
	trace *livePrefixTrace
	err   error
}

func (w *LiveWorker) startLivePrefixRecovery(state uint64, tidx int) <-chan livePrefixFutureResult {
	future := make(chan livePrefixFutureResult, 1)
	go func() {
		result := livePrefixFutureResult{}
		defer func() {
			if recovered := recover(); recovered != nil {
				result.err = panicValueAsError(recovered)
			}
			future <- result
		}()
		result.trace, result.err = w.recoverLivePrefix(state, tidx)
	}()
	return future
}

func awaitLivePrefixRecovery(future <-chan livePrefixFutureResult) (*livePrefixTrace, error) {
	result := <-future
	if result.err == nil {
		return result.trace, nil
	}
	if eval, ok := result.err.(*EvalException); ok {
		return nil, eval
	}
	return nil, NewExecutionException(result.err)
}

func (w *LiveWorker) recoverLivePrefix(state uint64, tidx int) (*livePrefixTrace, error) {
	prefix, err := w.getPath(state, tidx)
	if err != nil {
		return nil, err
	}
	plen := prefix.Size()
	fp := uint64(prefix.ElementAt(plen - 1))
	tool := w.ownedTool()
	sinfo, err := tool.GetState(fp)
	if err != nil {
		return nil, err
	}
	if sinfo == nil {
		return nil, NewEvalException(ECTLCFailedToRecoverInit)
	}
	states := make([]*TLCStateInfo, 0, plen)
	states = append(states, sinfo)

	for i := plen - 2; i >= 0; i-- {
		curFP := uint64(prefix.ElementAt(i))
		if curFP == fp {
			continue
		}
		sinfo, err = tool.GetState(curFP, sinfo)
		if err != nil {
			return nil, err
		}
		states = append(states, sinfo)
		fp = curFP
	}

	printable := make([]*TLCStateInfo, 0, len(states)+16)
	for i := 0; i < len(states)-1; i++ {
		if states[i+1] == nil {
			panic(NewNullPointerException())
		}
		alias, err := tool.EvalAliasInfoPrefix(states[i], states[i+1].ownedState(), states[:i])
		if err != nil {
			return nil, err
		}
		printable = append(printable, alias)
	}
	return &livePrefixTrace{fingerprints: prefix, states: states, printable: printable}, nil
}

func (w *LiveWorker) traceFingerprintCycle(state uint64, tidx int, nodeTbl *TableauNodePtrTable) (*LongVec, error) {
	cycleStack := NewIntStack()
	curNode, err := w.dfsPostFix(state, tidx, nodeTbl, cycleStack)
	if err != nil {
		return nil, err
	}
	postfix, err := w.bfsPostFix(state, tidx, nodeTbl, curNode)
	if err != nil {
		return nil, err
	}
	for cycleStack.Size() > 0 {
		fp := cycleStack.PopLong()
		if postfix.IsEmpty() || postfix.LastElement() != fp {
			postfix.AddElement(fp)
		}
		_ = cycleStack.PopInt()
	}
	return postfix, nil
}

func (w *LiveWorker) buildCounterExample(prefix *livePrefixTrace, cycle *LongVec) (*CounterExample, []*TLCStateInfo, []*TLCStateInfo, *TLCStateInfo, int, int, bool, error) {
	tool := w.ownedTool()
	states := prefix.states
	printableStates := prefix.printable
	cycleState := states[len(states)-1]
	cycleIndex := len(states) - 1
	loopOrdinal := int(int32(cycleState.StateNumber))
	sinfo := cycleState
	closingInfo := sinfo
	stuttering := true
	if cycle != nil && !cycle.IsEmpty() {
		cycle.Pack().RemoveLastIf(int64(cycleState.FingerPrint()))
		for i := cycle.Size() - 1; i >= 0; i-- {
			curFP := uint64(cycle.ElementAt(i))
			sucinfo, err := tool.GetState(curFP, sinfo)
			if err != nil {
				return nil, nil, nil, nil, 0, 0, false, err
			}
			if sucinfo == nil {
				panic(NewNullPointerException())
			}
			alias, err := tool.EvalAliasInfoPrefix(sinfo, sucinfo.ownedState(), states)
			if err != nil {
				return nil, nil, nil, nil, 0, 0, false, err
			}
			printableStates = append(printableStates, alias)
			states = append(states, sucinfo)
			sinfo = sucinfo
		}
		alias, err := tool.EvalAliasInfoPrefix(sinfo, cycleState.ownedState(), states)
		if err != nil {
			return nil, nil, nil, nil, 0, 0, false, err
		}
		printableStates = append(printableStates, alias)
		if sinfo.FingerPrint() != cycleState.FingerPrint() {
			stuttering = false
			closing, err := tool.GetState(cycleState.FingerPrint(), sinfo)
			if err != nil {
				return nil, nil, nil, nil, 0, 0, false, err
			}
			closingInfo = closing
			if isNil(tool.GetViewSpec()) {
				if closingInfo == nil {
					panic(NewNullPointerException())
				}
				if !cycleState.ownedState().Equal(closingInfo.ownedState()) {
					panic(NewAssertionError())
				}
			}
		} else {
			closingInfo = sinfo
		}
	} else {
		alias, err := tool.EvalAliasInfoPrefix(cycleState, cycleState.ownedState(), states)
		if err != nil {
			return nil, nil, nil, nil, 0, 0, false, err
		}
		printableStates = append(printableStates, alias)
	}
	return NewCounterExampleFromTraceActionLoop(states, closingInfo.GetAction(), loopOrdinal), printableStates, states, closingInfo, loopOrdinal, cycleIndex, stuttering, nil
}

func (w *LiveWorker) dfsPostFix(state uint64, tidx int, nodeTbl *TableauNodePtrTable, cycleStack *IntStack) (*GraphNode, error) {
	slen := len(w.ownedSolution().GetCheckState())
	alen := len(w.ownedSolution().GetCheckAction())
	aeStateRes := make([]bool, len(w.ownedPEM().ownedAEState()))
	aeActionRes := make([]bool, len(w.ownedPEM().ownedAEAction()))
	promiseRes := make([]bool, len(w.ownedSolution().GetPromises()))
	eaaction := w.ownedPEM().ownedEAAction()
	cnt := len(aeStateRes) + len(aeActionRes) + len(promiseRes)

	nodes := nodeTbl.GetNodes(state)
	tloc := nodeTbl.GetIdx(nodes, tidx)
	ptr := TableauGetElem(nodes, tloc)
	TableauSetSeenAt(nodes, tloc)
	curNode, err := w.getNode(state, tidx, ptr)
	if err != nil {
		return nil, err
	}

	for cnt > 0 {
		cnt0 := cnt
	next:
		for {
			for i, idx := range w.ownedPEM().ownedAEState() {
				if !aeStateRes[i] && curNode.GetCheckState(idx) {
					aeStateRes[i] = true
					cnt--
				}
			}
			for i, promise := range w.ownedSolution().GetPromises() {
				par := curNode.GetTNode(w.ownedSolution().GetTableau()).GetPar()
				if !promiseRes[i] && par.IsFulfilling(promise) {
					promiseRes[i] = true
					cnt--
				}
			}
			if cnt <= 0 {
				break
			}

			var nextState1 uint64
			var nextState2 uint64
			nextTidx1 := 0
			nextTidx2 := 0
			tloc1 := -1
			tloc2 := -1
			var nodes1 []int32
			var nodes2 []int32
			hasUnvisitedSucc := false
			cnt1 := cnt
			for i := 0; i < curNode.SuccSize(); i++ {
				nextState := curNode.GetStateFP(i)
				nextTidx := curNode.GetTIndex(i)
				nodes = nodeTbl.GetNodes(nextState)
				tloc = -1
				if nodes != nil {
					tloc = nodeTbl.GetIdx(nodes, nextTidx)
					if tloc == -1 {
						continue
					}
					if !curNode.GetCheckActionAll(slen, alen, i, eaaction) {
						continue
					}
					nextState1 = nextState
					nextTidx1 = nextTidx
					tloc1 = tloc
					nodes1 = nodes
					for j, idx := range w.ownedPEM().ownedAEAction() {
						if !aeActionRes[j] && curNode.GetCheckAction(slen, alen, i, idx) {
							aeActionRes[j] = true
							cnt--
						}
					}
				}
				if cnt < cnt1 {
					cycleStack.PushInt(int32(curNode.ownedTIndex()))
					cycleStack.PushLong(int64(curNode.ownedStateFP()))
					nextPtr := TableauGetPtr(TableauGetElem(nodes, tloc))
					curNode, err = w.getNode(nextState, nextTidx, nextPtr)
					if err != nil {
						return nil, err
					}
					nodeTbl.ResetElems()
					break next
				}
				if nodes != nil && tloc != -1 && !TableauIsSeenAt(nodes, tloc) {
					hasUnvisitedSucc = true
					nextState2 = nextState
					nextTidx2 = nextTidx
					tloc2 = tloc
					nodes2 = nodes
				}
			}

			if cnt < cnt0 {
				cycleStack.PushInt(int32(curNode.ownedTIndex()))
				cycleStack.PushLong(int64(curNode.ownedStateFP()))
				nextPtr := TableauGetPtr(TableauGetElem(nodes1, tloc1))
				curNode, err = w.getNode(nextState1, nextTidx1, nextPtr)
				if err != nil {
					return nil, err
				}
				nodeTbl.ResetElems()
				break
			}

			for !hasUnvisitedSucc {
				curState := uint64(cycleStack.PopLong())
				curTidx := int(cycleStack.PopInt())
				curPtr := TableauGetPtr(nodeTbl.Get(curState, curTidx))
				curNode, err = w.getNode(curState, curTidx, curPtr)
				if err != nil {
					return nil, err
				}
				for i := 0; i < curNode.SuccSize(); i++ {
					nextState2 = curNode.GetStateFP(i)
					nextTidx2 = curNode.GetTIndex(i)
					nodes2 = nodeTbl.GetNodes(nextState2)
					if nodes2 != nil {
						tloc2 = nodeTbl.GetIdx(nodes2, nextTidx2)
						if tloc2 != -1 && !TableauIsSeenAt(nodes2, tloc2) {
							hasUnvisitedSucc = true
							break
						}
					}
				}
			}

			cycleStack.PushInt(int32(curNode.ownedTIndex()))
			cycleStack.PushLong(int64(curNode.ownedStateFP()))
			nextPtr := TableauGetPtr(TableauGetElem(nodes2, tloc2))
			curNode, err = w.getNode(nextState2, nextTidx2, nextPtr)
			if err != nil {
				return nil, err
			}
			TableauSetSeenAt(nodes2, tloc2)
		}
	}
	nodeTbl.ResetElems()
	return curNode, nil
}

func (w *LiveWorker) bfsPostFix(state uint64, tidx int, nodeTbl *TableauNodePtrTable, curNode *GraphNode) (*LongVec, error) {
	slen := len(w.ownedSolution().GetCheckState())
	alen := len(w.ownedSolution().GetCheckAction())
	eaaction := w.ownedPEM().ownedEAAction()
	postfix := NewLongVecWithCapacity(16)
	startState := curNode.ownedStateFP()
	startTidx := curNode.ownedTIndex()

	if startState == state && startTidx == tidx {
		return postfix, nil
	}

	queue := NewIntQueueWithDisk(w.liveCheckMetaDir(), "", intQueueInitialSize)
	curState := startState
	ploc := TableauNoParent
	curLoc := nodeTbl.GetNodesLoc(curState)
	nodes := nodeTbl.GetNodesByLoc(curLoc)
	TableauSetSeen(nodes)

	for {
		tloc := TableauStartLoc(nodes)
		for tloc != TableauEndMarker {
			curTidx := TableauGetTidx(nodes, tloc)
			curPtr := TableauGetPtr(TableauGetElem(nodes, tloc))
			curNode, err := w.getNode(curState, curTidx, curPtr)
			if err != nil {
				return nil, err
			}
			for j := 0; j < curNode.SuccSize(); j++ {
				nextState := curNode.GetStateFP(j)
				nextTidx := curNode.GetTIndex(j)
				if curState == nextState && curTidx == nextTidx {
					if !TableauIsSeen(nodes) {
						panic(NewAssertionError())
					}
					continue
				}
				if !curNode.GetCheckActionAll(slen, alen, j, eaaction) {
					continue
				}
				if nextState == state && nextTidx == tidx {
					for curState != startState {
						postfix.AddElement(int64(curState))
						nodes = nodeTbl.GetNodesByLoc(ploc)
						curState = TableauGetKey(nodes)
						ploc = TableauGetParent(nodes)
					}
					postfix.AddElement(int64(startState))
					return postfix, nil
				}

				nodes1 := nodeTbl.GetNodes(nextState)
				if nodes1 != nil && !TableauIsSeen(nodes1) {
					TableauSetSeen(nodes1)
					queue.EnqueueLong(int64(nextState))
					queue.EnqueueInt(int32(curLoc))
				}
			}
			tloc = TableauNextLoc(nodes, tloc)
		}
		TableauSetParent(nodes, ploc)
		curState = uint64(queue.DequeueLong())
		ploc = int(queue.DequeueInt())
		curLoc = nodeTbl.GetNodesLoc(curState)
		nodes = nodeTbl.GetNodesByLoc(curLoc)
	}
}

func (w *LiveWorker) makeNodePtrTable() error {
	w.require()
	if w.ownedChecker() == nil {
		panic(NewNullPointerException())
	}
	if w.ownedChecker().tableau {
		dgraph := w.ownedChecker().ownedTableauDiskGraph()
		if dgraph == nil {
			panic(NewNullPointerException())
		}
		return dgraph.MakeNodePtrTbl()
	}
	dgraph := w.ownedChecker().ownedDiskGraph()
	if dgraph == nil {
		panic(NewNullPointerException())
	}
	return dgraph.MakeNodePtrTbl()
}

func (w *LiveWorker) getPath(state uint64, tidx int) (*LongVec, error) {
	w.require()
	if w.ownedChecker() == nil {
		panic(NewNullPointerException())
	}
	if w.ownedChecker().tableau {
		dgraph := w.ownedChecker().ownedTableauDiskGraph()
		if dgraph == nil {
			panic(NewNullPointerException())
		}
		return dgraph.GetPath(state, tidx)
	}
	dgraph := w.ownedChecker().ownedDiskGraph()
	if dgraph == nil {
		panic(NewNullPointerException())
	}
	return dgraph.GetPath(state, tidx)
}

func (w *LiveWorker) initNodes() *LongVec {
	w.require()
	if w.ownedChecker() == nil {
		panic(NewNullPointerException())
	}
	if w.ownedChecker().tableau {
		dgraph := w.ownedChecker().ownedTableauDiskGraph()
		if dgraph == nil {
			panic(NewNullPointerException())
		}
		return dgraph.GetInitNodes()
	}
	dgraph := w.ownedChecker().ownedDiskGraph()
	if dgraph == nil {
		panic(NewNullPointerException())
	}
	return dgraph.GetInitNodes()
}

func (w *LiveWorker) getLink(state uint64, tidx int) int64 {
	w.require()
	if w.ownedChecker() == nil {
		panic(NewNullPointerException())
	}
	if w.ownedChecker().tableau {
		dgraph := w.ownedChecker().ownedTableauDiskGraph()
		if dgraph == nil {
			panic(NewNullPointerException())
		}
		return dgraph.GetLink(state, tidx)
	}
	dgraph := w.ownedChecker().ownedDiskGraph()
	if dgraph == nil {
		panic(NewNullPointerException())
	}
	return dgraph.GetLink(state, tidx)
}

func (w *LiveWorker) putLink(state uint64, tidx int, link int64) int64 {
	w.require()
	if w.ownedChecker() == nil {
		panic(NewNullPointerException())
	}
	if w.ownedChecker().tableau {
		dgraph := w.ownedChecker().ownedTableauDiskGraph()
		if dgraph == nil {
			panic(NewNullPointerException())
		}
		return dgraph.PutLink(state, tidx, link)
	}
	dgraph := w.ownedChecker().ownedDiskGraph()
	if dgraph == nil {
		panic(NewNullPointerException())
	}
	return dgraph.PutLink(state, tidx, link)
}

func (w *LiveWorker) setMaxLink(state uint64, tidx int) {
	w.require()
	if w.ownedChecker() == nil {
		panic(NewNullPointerException())
	}
	if w.ownedChecker().tableau {
		dgraph := w.ownedChecker().ownedTableauDiskGraph()
		if dgraph == nil {
			panic(NewNullPointerException())
		}
		dgraph.SetMaxLink(state, tidx)
		return
	}
	dgraph := w.ownedChecker().ownedDiskGraph()
	if dgraph == nil {
		panic(NewNullPointerException())
	}
	dgraph.SetMaxLink(state, tidx)
}

func (w *LiveWorker) getNode(state uint64, tidx int, loc int64) (*GraphNode, error) {
	w.require()
	if w.ownedChecker() == nil {
		panic(NewNullPointerException())
	}
	if w.ownedChecker().tableau {
		dgraph := w.ownedChecker().ownedTableauDiskGraph()
		if dgraph == nil {
			panic(NewNullPointerException())
		}
		return dgraph.diskGraphNodeAt(state, tidx, loc)
	}
	dgraph := w.ownedChecker().ownedDiskGraph()
	if dgraph == nil {
		panic(NewNullPointerException())
	}
	return dgraph.diskGraphNodeAt(state, tidx, loc)
}

func minInt64(a int64, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
