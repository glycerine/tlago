package tlc

import (
	"fmt"
	"os"
	"sync/atomic"
)

const ModelCheckerMXObjectName = "tlc2.tool:type=ModelChecker"

var diskFPSetMXWrapperCount atomic.Int64

type TLCStandardMBean struct {
	ObjectName string
	Registered bool
	noOp       bool
}

func NewTLCStandardMBean() *TLCStandardMBean {
	return &TLCStandardMBean{}
}

func (m *TLCStandardMBean) GetVersion() string {
	m.require()
	return TLCVersion()
}

func (m *TLCStandardMBean) GetRevision() string {
	m.require()
	if rev := TLCRevision(); rev != "" {
		return rev
	}
	return "N/A"
}

func (m *TLCStandardMBean) RegisterMBean(objectName string) bool {
	m.require()
	if m.noOp {
		return true
	}
	m.ObjectName = objectName
	m.Registered = true
	return true
}

func (m *TLCStandardMBean) Unregister() bool {
	m.require()
	if m.noOp {
		return true
	}
	m.Registered = false
	return true
}

func NewNullTLCStandardMBean() *TLCStandardMBean {
	bean := NewTLCStandardMBean()
	bean.noOp = true
	bean.Registered = true
	return bean
}

func (m *TLCStandardMBean) require() {
	if m == nil {
		panic(NewNullPointerException())
	}
}

type ModelCheckerMXWrapper struct {
	*TLCStandardMBean
	ModelChecker      *ModelChecker
	TLC               *TLC
	ownersCaptured    bool
	modelCheckerOwner *ModelChecker
	tlcOwner          *TLC
}

func NewModelCheckerMXWrapper(modelChecker *ModelChecker, tlc *TLC) *ModelCheckerMXWrapper {
	wrapper := &ModelCheckerMXWrapper{
		TLCStandardMBean:  NewTLCStandardMBean(),
		ModelChecker:      modelChecker,
		TLC:               tlc,
		ownersCaptured:    true,
		modelCheckerOwner: modelChecker,
		tlcOwner:          tlc,
	}
	wrapper.RegisterMBean(ModelCheckerMXObjectName)
	return wrapper
}

func (w *ModelCheckerMXWrapper) GetStatesGenerated() int64 {
	return w.requireModelChecker().GetStatesGenerated()
}

func (w *ModelCheckerMXWrapper) GetDistinctStatesGenerated() int64 {
	mc := w.requireModelChecker()
	if fpSet, ok := diskFPSetForMX(mc.FPSet); ok {
		return fpSet.GetFileCnt() + fpSet.GetTblCnt()
	}
	return int64(mc.GetDistinctStatesGenerated())
}

func (w *ModelCheckerMXWrapper) GetStateQueueSize() int64 {
	return w.requireModelChecker().GetStateQueueSize()
}

func (w *ModelCheckerMXWrapper) GetStatesGeneratedPerMinute() int64 {
	return atomic.LoadInt64(&w.requireModelChecker().StatesPerMinute)
}

func (w *ModelCheckerMXWrapper) GetDistinctStatesGeneratedPerMinute() int64 {
	return atomic.LoadInt64(&w.requireModelChecker().DistinctStatesPerMinute)
}

func (w *ModelCheckerMXWrapper) GetProgress() int {
	return int(w.requireModelChecker().GetProgress())
}

func (w *ModelCheckerMXWrapper) GetWorkerCount() int {
	w.require()
	return NumWorkers()
}

func (w *ModelCheckerMXWrapper) Checkpoint() {
	w.require()
	ForceCheckpoint()
}

func (w *ModelCheckerMXWrapper) GetAverageBlockCnt() int64 {
	w.require()
	return 1
}

func (w *ModelCheckerMXWrapper) GetRuntimeRatio() float64 {
	return w.requireModelChecker().GetRuntimeRatio()
}

func (w *ModelCheckerMXWrapper) LiveCheck() {
	w.requireModelChecker().ForceLivenessCheck()
}

func (w *ModelCheckerMXWrapper) GetCurrentState() string {
	mc := w.requireModelChecker()
	if isNil(mc.StateQueue) {
		panic(NewNullPointerException())
	}
	state := mc.StateQueue.SPeek()
	if state == nil {
		return "N/A"
	}
	return state.EvalStateLevelAlias().String()
}

func (w *ModelCheckerMXWrapper) GetSpecName() string {
	return w.requireTLC().GetSpecName()
}

func (w *ModelCheckerMXWrapper) GetModelName() string {
	return w.requireTLC().GetModelName()
}

func (w *ModelCheckerMXWrapper) Stop() {
	w.requireModelChecker().Stop()
}

func (w *ModelCheckerMXWrapper) Suspend() {
	w.requireModelChecker().Suspend()
}

func (w *ModelCheckerMXWrapper) Resume() {
	w.requireModelChecker().Resume()
}

func (w *ModelCheckerMXWrapper) require() {
	if w == nil {
		panic(NewNullPointerException())
	}
}

func (w *ModelCheckerMXWrapper) requireModelChecker() *ModelChecker {
	w.require()
	modelChecker := w.ModelChecker
	if w.ownersCaptured {
		modelChecker = w.modelCheckerOwner
	}
	if modelChecker == nil {
		panic(NewNullPointerException())
	}
	return modelChecker
}

func (w *ModelCheckerMXWrapper) requireTLC() *TLC {
	w.require()
	tlc := w.TLC
	if w.ownersCaptured {
		tlc = w.tlcOwner
	}
	if tlc == nil {
		panic(NewNullPointerException())
	}
	return tlc
}

func diskFPSetForMX(set FPSet) (*DiskFPSet, bool) {
	switch s := set.(type) {
	case *DiskFPSet:
		return s, s != nil
	case *HeapBasedDiskFPSet:
		if s == nil || s.DiskFPSet == nil {
			return nil, false
		}
		return s.DiskFPSet, true
	case *LSBDiskFPSet:
		if s == nil || s.HeapBasedDiskFPSet == nil || s.DiskFPSet == nil {
			return nil, false
		}
		return s.DiskFPSet, true
	case *MSBDiskFPSet:
		if s == nil || s.HeapBasedDiskFPSet == nil || s.DiskFPSet == nil {
			return nil, false
		}
		return s.DiskFPSet, true
	case *NonCheckpointableDiskFPSet:
		if s == nil || s.DiskFPSet == nil {
			return nil, false
		}
		return s.DiskFPSet, true
	case *OffHeapDiskFPSet:
		if s == nil || s.NonCheckpointableDiskFPSet == nil || s.DiskFPSet == nil {
			return nil, false
		}
		return s.DiskFPSet, true
	default:
		return nil, false
	}
}

type TLCServerMXWrapper struct {
	*TLCStandardMBean
	Server         *TLCServer
	serverCaptured bool
	serverOwner    *TLCServer
}

func NewTLCServerMXWrapper(server *TLCServer) *TLCServerMXWrapper {
	wrapper := &TLCServerMXWrapper{
		TLCStandardMBean: NewTLCStandardMBean(),
		Server:           server,
		serverCaptured:   true,
		serverOwner:      server,
	}
	wrapper.RegisterMBean(ModelCheckerMXObjectName)
	return wrapper
}

func (w *TLCServerMXWrapper) requireServer() *TLCServer {
	w.require()
	server := w.Server
	if w.serverCaptured {
		server = w.serverOwner
	}
	if server == nil {
		panic(NewNullPointerException())
	}
	return server
}

func (w *TLCServerMXWrapper) GetStatesGenerated() int64 {
	server := w.requireServer()
	if !server.IsRunning() {
		return -1
	}
	return server.GetStatesGenerated()
}

func (w *TLCServerMXWrapper) GetDistinctStatesGenerated() int64 {
	server := w.requireServer()
	if !server.IsRunning() {
		return -1
	}
	if manager := server.GetFPSetManager(); manager != nil {
		return int64(manager.Size())
	}
	return -1
}

func (w *TLCServerMXWrapper) GetStateQueueSize() int64 {
	return w.requireServer().GetNewStates()
}

func (w *TLCServerMXWrapper) GetStatesGeneratedPerMinute() int64 {
	return w.requireServer().GetStatesGeneratedPerMinute()
}

func (w *TLCServerMXWrapper) GetDistinctStatesGeneratedPerMinute() int64 {
	return w.requireServer().GetDistinctStatesGeneratedPerMinute()
}

func (w *TLCServerMXWrapper) GetProgress() int {
	server := w.requireServer()
	if !server.IsRunning() {
		return -1
	}
	if server.ownedTrace() == nil {
		panic(NewNullPointerException())
	}
	level, err := server.ownedTrace().GetLevelForReportingWithError()
	if err != nil {
		if !isJavaIOException(err) {
			panic(err)
		}
		fmt.Fprint(os.Stderr, javaThrowableStackTrace(err))
		return -1
	}
	return level
}

func (w *TLCServerMXWrapper) GetWorkerCount() int {
	return w.requireServer().GetWorkerCount()
}

func (w *TLCServerMXWrapper) Checkpoint() {
	w.require()
	ForceCheckpoint()
}

func (w *TLCServerMXWrapper) GetAverageBlockCnt() int64 {
	return w.requireServer().GetAverageBlockCnt()
}

func (w *TLCServerMXWrapper) GetRuntimeRatio() float64 {
	w.require()
	return 0
}

func (w *TLCServerMXWrapper) LiveCheck() { w.require() }

func (w *TLCServerMXWrapper) GetCurrentState() string {
	server := w.requireServer()
	if isNil(server.ownedStateQueue()) {
		panic(NewNullPointerException())
	}
	state := server.ownedStateQueue().SPeek()
	if state == nil {
		return "N/A"
	}
	return state.String()
}

func (w *TLCServerMXWrapper) GetSpecName() string {
	server := w.requireServer()
	if !server.IsRunning() {
		return "N/A"
	}
	return server.GetSpecFileName()
}

func (w *TLCServerMXWrapper) GetModelName() string {
	server := w.requireServer()
	if !server.IsRunning() {
		return "N/A"
	}
	return server.GetConfigFileName()
}

func (w *TLCServerMXWrapper) Stop() {
	server := w.requireServer()
	server.monitor.Lock()
	defer server.monitor.Unlock()
	server.SetDone()
	if isNil(server.ownedStateQueue()) {
		panic(NewNullPointerException())
	}
	server.ownedStateQueue().FinishAll()
	server.notifyCompletionLocked()
}

func (w *TLCServerMXWrapper) Suspend() {
	server := w.requireServer()
	server.monitor.Lock()
	defer server.monitor.Unlock()
	if isNil(server.ownedStateQueue()) {
		panic(NewNullPointerException())
	}
	server.ownedStateQueue().SuspendAll()
}

func (w *TLCServerMXWrapper) Resume() {
	server := w.requireServer()
	server.monitor.Lock()
	defer server.monitor.Unlock()
	if isNil(server.ownedStateQueue()) {
		panic(NewNullPointerException())
	}
	server.ownedStateQueue().ResumeAll()
}

func (w *TLCServerMXWrapper) require() {
	if w == nil {
		panic(NewNullPointerException())
	}
}

// FPSetStatistic preserves virtual dispatch for the source management contract.
type FPSetStatistic interface {
	GetTblCnt() int64
	GetFileCnt() int64
	GetIndexCapacity() int64
	GetDiskLookupCnt() uint64
	GetMemHitCnt() uint64
	GetDiskHitCnt() uint64
	GetDiskWriteCnt() uint64
	GetDiskSeekCnt() uint64
	GetDiskSeekCache() uint64
	GetGrowDiskMark() int
	GetCheckPointMark() int
	GetBucketCapacity() int64
	GetTblCapacity() int64
	GetOverallCapacity() int64
	GetTblLoad() int64
	GetMaxTblCnt() int64
	Sizeof() uint64
	GetFlushTime() int64
	GetReaderWriterCnt() int
	GetLoadFactor() float64
	ForceFlush()
	CheckInvariant(...uint64) bool
	GetLockCnt() int
}

type DiskFPSetMXWrapper struct {
	*TLCStandardMBean
	FPSet           FPSetStatistic
	ObjectName      string
	ownersCaptured  bool
	fpSetOwner      FPSetStatistic
	objectNameOwner string
}

func NewDiskFPSetMXWrapper(fpSet FPSetStatistic) *DiskFPSetMXWrapper {
	id := diskFPSetMXWrapperCount.Add(1) - 1
	wrapper := &DiskFPSetMXWrapper{
		TLCStandardMBean: NewTLCStandardMBean(),
		FPSet:            fpSet,
		ObjectName:       "DiskFPSet" + fmtInt64(id),
		ownersCaptured:   true,
		fpSetOwner:       fpSet,
	}
	wrapper.objectNameOwner = wrapper.ObjectName
	wrapper.RegisterMBean("tlc2.tool.fp:type=" + wrapper.ObjectName)
	return wrapper
}

func (w *DiskFPSetMXWrapper) GetObjectName() string {
	w.require()
	if w.ownersCaptured {
		return w.objectNameOwner
	}
	return w.ObjectName
}

func (w *DiskFPSetMXWrapper) GetTblCnt() int64 {
	return w.requireFPSet().GetTblCnt()
}

func (w *DiskFPSetMXWrapper) GetFileCnt() int64 {
	return w.requireFPSet().GetFileCnt()
}

func (w *DiskFPSetMXWrapper) GetIndexCnt() int64 {
	return w.requireFPSet().GetIndexCapacity()
}

func (w *DiskFPSetMXWrapper) GetIndexCapacity() int64 {
	w.require()
	return w.GetIndexCnt()
}

func (w *DiskFPSetMXWrapper) GetDiskLookupCnt() uint64 {
	return w.requireFPSet().GetDiskLookupCnt()
}

func (w *DiskFPSetMXWrapper) GetMemHitCnt() uint64 {
	return w.requireFPSet().GetMemHitCnt()
}

func (w *DiskFPSetMXWrapper) GetDiskHitCnt() uint64 {
	return w.requireFPSet().GetDiskHitCnt()
}

func (w *DiskFPSetMXWrapper) GetDiskWriteCnt() uint64 {
	return w.requireFPSet().GetDiskWriteCnt()
}

func (w *DiskFPSetMXWrapper) GetDiskSeekCnt() uint64 {
	return w.requireFPSet().GetDiskSeekCnt()
}

func (w *DiskFPSetMXWrapper) GetDiskSeekCache() uint64 {
	return w.requireFPSet().GetDiskSeekCache()
}

func (w *DiskFPSetMXWrapper) GetDiskSeekRate() float64 {
	w.require()
	diskSeekCnt := w.GetDiskSeekCnt()
	diskSeekCache := w.GetDiskSeekCache()
	return float64(diskSeekCache) / float64(diskSeekCache+diskSeekCnt)
}

func (w *DiskFPSetMXWrapper) GetGrowDiskMark() int {
	return w.requireFPSet().GetGrowDiskMark()
}

func (w *DiskFPSetMXWrapper) GetCheckPointMark() int {
	return w.requireFPSet().GetCheckPointMark()
}

func (w *DiskFPSetMXWrapper) GetBucketCapacity() int64 {
	return w.requireFPSet().GetBucketCapacity()
}

func (w *DiskFPSetMXWrapper) GetTblCapacity() int64 {
	return w.requireFPSet().GetTblCapacity()
}

func (w *DiskFPSetMXWrapper) GetOverallCapacity() int64 {
	return w.requireFPSet().GetOverallCapacity()
}

func (w *DiskFPSetMXWrapper) GetTblLoad() int64 {
	return w.requireFPSet().GetTblLoad()
}

func (w *DiskFPSetMXWrapper) GetMaxTblCnt() int64 {
	return w.requireFPSet().GetMaxTblCnt()
}

func (w *DiskFPSetMXWrapper) GetSizeOf() uint64 {
	return w.requireFPSet().Sizeof()
}

func (w *DiskFPSetMXWrapper) GetFlushTime() int64 {
	return w.requireFPSet().GetFlushTime()
}

func (w *DiskFPSetMXWrapper) GetReaderWriterCnt() int {
	return w.requireFPSet().GetReaderWriterCnt()
}

func (w *DiskFPSetMXWrapper) GetLoadFactor() float64 {
	return w.requireFPSet().GetLoadFactor()
}

func (w *DiskFPSetMXWrapper) ForceFlush() {
	w.requireFPSet().ForceFlush()
}

func (w *DiskFPSetMXWrapper) CheckInvariant() bool {
	return w.requireFPSet().CheckInvariant()
}

func (w *DiskFPSetMXWrapper) GetLockCnt() int {
	return w.requireFPSet().GetLockCnt()
}

func (w *DiskFPSetMXWrapper) require() {
	if w == nil {
		panic(NewNullPointerException())
	}
}

func (w *DiskFPSetMXWrapper) requireFPSet() FPSetStatistic {
	w.require()
	fpSet := w.FPSet
	if w.ownersCaptured {
		fpSet = w.fpSetOwner
	}
	if fpSet == nil {
		panic(NewNullPointerException())
	}
	return fpSet
}

type BucketStatisticsMXView interface {
	GetObservations() int64
	GetMedian() int
	GetMean() float64
	GetMin() int
	GetMax() int
	GetStdDev() float64
	GetPercentile(float64) float64
}

type BucketStatisticsMXWrapper struct {
	*TLCStandardMBean
	Stats              BucketStatisticsMXView
	ObjectName         string
	objectNameCaptured bool
	objectNameOwner    string
}

func NewBucketStatisticsMXWrapper(stats BucketStatisticsMXView, graphName string, pkg string) *BucketStatisticsMXWrapper {
	wrapper := &BucketStatisticsMXWrapper{
		TLCStandardMBean:   NewTLCStandardMBean(),
		Stats:              stats,
		ObjectName:         graphName,
		objectNameCaptured: true,
		objectNameOwner:    graphName,
	}
	wrapper.RegisterMBean(pkg + ":type=" + graphName)
	return wrapper
}

func (w *BucketStatisticsMXWrapper) GetObjectName() string {
	w.require()
	if w.objectNameCaptured {
		return w.objectNameOwner
	}
	return w.ObjectName
}

func (w *BucketStatisticsMXWrapper) GetObservations() int64 {
	return w.bucketStats().GetObservations()
}

func (w *BucketStatisticsMXWrapper) GetMedian() int {
	return w.bucketStats().GetMedian()
}

func (w *BucketStatisticsMXWrapper) GetMean() float64 {
	return w.bucketStats().GetMean()
}

func (w *BucketStatisticsMXWrapper) GetMin() int {
	return w.bucketStats().GetMin()
}

func (w *BucketStatisticsMXWrapper) GetMax() int {
	return w.bucketStats().GetMax()
}

func (w *BucketStatisticsMXWrapper) GetStdDev() float64 {
	return w.bucketStats().GetStdDev()
}

func (w *BucketStatisticsMXWrapper) Get75Percentile() float64 {
	return w.bucketStats().GetPercentile(0.75)
}

func (w *BucketStatisticsMXWrapper) Get95Percentile() float64 {
	return w.bucketStats().GetPercentile(0.95)
}

func (w *BucketStatisticsMXWrapper) Get98Percentile() float64 {
	return w.bucketStats().GetPercentile(0.98)
}

func (w *BucketStatisticsMXWrapper) Get99Percentile() float64 {
	return w.bucketStats().GetPercentile(0.99)
}

func (w *BucketStatisticsMXWrapper) Get999Percentile() float64 {
	return w.bucketStats().GetPercentile(0.999)
}

func (w *BucketStatisticsMXWrapper) bucketStats() BucketStatisticsMXView {
	w.require()
	if isNil(w.Stats) {
		panic(NewNullPointerException())
	}
	return w.Stats
}

func (w *BucketStatisticsMXWrapper) require() {
	if w == nil {
		panic(NewNullPointerException())
	}
}
