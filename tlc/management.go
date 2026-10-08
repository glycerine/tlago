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
}

func NewTLCStandardMBean() *TLCStandardMBean {
	return &TLCStandardMBean{}
}

func (m *TLCStandardMBean) GetVersion() string {
	return TLCVersion()
}

func (m *TLCStandardMBean) GetRevision() string {
	if rev := TLCRevision(); rev != "" {
		return rev
	}
	return "N/A"
}

func (m *TLCStandardMBean) RegisterMBean(objectName string) bool {
	if m == nil {
		return true
	}
	m.ObjectName = objectName
	m.Registered = true
	return true
}

func (m *TLCStandardMBean) Unregister() bool {
	if m == nil {
		return true
	}
	m.Registered = false
	return true
}

func NewNullTLCStandardMBean() *TLCStandardMBean {
	bean := NewTLCStandardMBean()
	bean.Registered = true
	return bean
}

type ModelCheckerMXWrapper struct {
	*TLCStandardMBean
	ModelChecker *ModelChecker
	TLC          *TLC
}

func NewModelCheckerMXWrapper(modelChecker *ModelChecker, tlc *TLC) *ModelCheckerMXWrapper {
	wrapper := &ModelCheckerMXWrapper{
		TLCStandardMBean: NewTLCStandardMBean(),
		ModelChecker:     modelChecker,
		TLC:              tlc,
	}
	wrapper.RegisterMBean(ModelCheckerMXObjectName)
	return wrapper
}

func (w *ModelCheckerMXWrapper) GetStatesGenerated() int64 {
	if w == nil || w.ModelChecker == nil {
		return 0
	}
	return w.ModelChecker.GetStatesGenerated()
}

func (w *ModelCheckerMXWrapper) GetDistinctStatesGenerated() int64 {
	if w == nil || w.ModelChecker == nil {
		return 0
	}
	if fpSet, ok := diskFPSetForMX(w.ModelChecker.FPSet); ok {
		return fpSet.GetFileCnt() + fpSet.GetTblCnt()
	}
	return int64(w.ModelChecker.GetDistinctStatesGenerated())
}

func (w *ModelCheckerMXWrapper) GetStateQueueSize() int64 {
	if w == nil || w.ModelChecker == nil {
		return 0
	}
	return w.ModelChecker.GetStateQueueSize()
}

func (w *ModelCheckerMXWrapper) GetStatesGeneratedPerMinute() int64 {
	if w == nil || w.ModelChecker == nil {
		return 0
	}
	return w.ModelChecker.StatesPerMinute
}

func (w *ModelCheckerMXWrapper) GetDistinctStatesGeneratedPerMinute() int64 {
	if w == nil || w.ModelChecker == nil {
		return 0
	}
	return w.ModelChecker.DistinctStatesPerMinute
}

func (w *ModelCheckerMXWrapper) GetProgress() int {
	if w == nil || w.ModelChecker == nil {
		return 0
	}
	return int(w.ModelChecker.GetProgress())
}

func (w *ModelCheckerMXWrapper) GetWorkerCount() int {
	return NumWorkers()
}

func (w *ModelCheckerMXWrapper) Checkpoint() {
	ForceCheckpoint()
}

func (w *ModelCheckerMXWrapper) GetAverageBlockCnt() int64 {
	return 1
}

func (w *ModelCheckerMXWrapper) GetRuntimeRatio() float64 {
	if w == nil || w.ModelChecker == nil {
		return 0
	}
	return w.ModelChecker.RuntimeRatio
}

func (w *ModelCheckerMXWrapper) LiveCheck() {
	if w == nil || w.ModelChecker == nil {
		return
	}
	w.ModelChecker.ForceLivenessCheck()
}

func (w *ModelCheckerMXWrapper) GetCurrentState() string {
	if w == nil || w.ModelChecker == nil || w.ModelChecker.StateQueue == nil {
		return "N/A"
	}
	state := w.ModelChecker.StateQueue.SPeek()
	if state == nil {
		return "N/A"
	}
	return state.EvalStateLevelAlias().String()
}

func (w *ModelCheckerMXWrapper) GetSpecName() string {
	if w == nil || w.TLC == nil {
		return "N/A"
	}
	return w.TLC.GetSpecName()
}

func (w *ModelCheckerMXWrapper) GetModelName() string {
	if w == nil || w.TLC == nil {
		return "N/A"
	}
	return w.TLC.GetModelName()
}

func (w *ModelCheckerMXWrapper) Stop() {
	if w != nil && w.ModelChecker != nil {
		w.ModelChecker.Stop()
	}
}

func (w *ModelCheckerMXWrapper) Suspend() {
	if w != nil && w.ModelChecker != nil {
		w.ModelChecker.Suspend()
	}
}

func (w *ModelCheckerMXWrapper) Resume() {
	if w != nil && w.ModelChecker != nil {
		w.ModelChecker.Resume()
	}
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
	Server *TLCServer
}

func NewTLCServerMXWrapper(server *TLCServer) *TLCServerMXWrapper {
	wrapper := &TLCServerMXWrapper{
		TLCStandardMBean: NewTLCStandardMBean(),
		Server:           server,
	}
	wrapper.RegisterMBean(ModelCheckerMXObjectName)
	return wrapper
}

func (w *TLCServerMXWrapper) requireServer() *TLCServer {
	if w == nil || w.Server == nil {
		panic(NewNullPointerException())
	}
	return w.Server
}

func (w *TLCServerMXWrapper) GetStatesGenerated() int64 {
	if !w.requireServer().IsRunning() {
		return -1
	}
	return w.Server.GetStatesGenerated()
}

func (w *TLCServerMXWrapper) GetDistinctStatesGenerated() int64 {
	if !w.requireServer().IsRunning() {
		return -1
	}
	if manager := w.Server.GetFPSetManager(); manager != nil {
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
	if !w.requireServer().IsRunning() {
		return -1
	}
	if w.Server.Trace == nil {
		panic(NewNullPointerException())
	}
	level, err := w.Server.Trace.GetLevelForReportingWithError()
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
	ForceCheckpoint()
}

func (w *TLCServerMXWrapper) GetAverageBlockCnt() int64 {
	return w.requireServer().GetAverageBlockCnt()
}

func (w *TLCServerMXWrapper) GetRuntimeRatio() float64 {
	return 0
}

func (w *TLCServerMXWrapper) LiveCheck() {}

func (w *TLCServerMXWrapper) GetCurrentState() string {
	if w.requireServer().StateQueue == nil {
		panic(NewNullPointerException())
	}
	state := w.Server.StateQueue.SPeek()
	if state == nil {
		return "N/A"
	}
	return state.String()
}

func (w *TLCServerMXWrapper) GetSpecName() string {
	if !w.requireServer().IsRunning() {
		return "N/A"
	}
	return w.Server.GetSpecFileName()
}

func (w *TLCServerMXWrapper) GetModelName() string {
	if !w.requireServer().IsRunning() {
		return "N/A"
	}
	return w.Server.GetConfigFileName()
}

func (w *TLCServerMXWrapper) Stop() {
	if w == nil || w.Server == nil {
		panic(NewNullPointerException())
	}
	w.Server.monitor.Lock()
	defer w.Server.monitor.Unlock()
	w.Server.SetDone()
	if w.Server.StateQueue == nil {
		panic(NewNullPointerException())
	}
	w.Server.StateQueue.FinishAll()
	w.Server.notifyCompletionLocked()
}

func (w *TLCServerMXWrapper) Suspend() {
	if w == nil || w.Server == nil {
		panic(NewNullPointerException())
	}
	w.Server.monitor.Lock()
	defer w.Server.monitor.Unlock()
	if w.Server.StateQueue == nil {
		panic(NewNullPointerException())
	}
	w.Server.StateQueue.SuspendAll()
}

func (w *TLCServerMXWrapper) Resume() {
	if w == nil || w.Server == nil {
		panic(NewNullPointerException())
	}
	w.Server.monitor.Lock()
	defer w.Server.monitor.Unlock()
	if w.Server.StateQueue == nil {
		panic(NewNullPointerException())
	}
	w.Server.StateQueue.ResumeAll()
}

type DiskFPSetMXWrapper struct {
	*TLCStandardMBean
	FPSet      *DiskFPSet
	ObjectName string
}

func NewDiskFPSetMXWrapper(fpSet *DiskFPSet) *DiskFPSetMXWrapper {
	id := diskFPSetMXWrapperCount.Add(1) - 1
	wrapper := &DiskFPSetMXWrapper{
		TLCStandardMBean: NewTLCStandardMBean(),
		FPSet:            fpSet,
		ObjectName:       "DiskFPSet" + fmtInt64(id),
	}
	wrapper.RegisterMBean("tlc2.tool.fp:type=" + wrapper.ObjectName)
	return wrapper
}

func (w *DiskFPSetMXWrapper) GetObjectName() string {
	if w == nil {
		return ""
	}
	return w.ObjectName
}

func (w *DiskFPSetMXWrapper) GetTblCnt() int64 {
	if w == nil || w.FPSet == nil {
		return 0
	}
	return w.FPSet.GetTblCnt()
}

func (w *DiskFPSetMXWrapper) GetFileCnt() int64 {
	if w == nil || w.FPSet == nil {
		return 0
	}
	return w.FPSet.GetFileCnt()
}

func (w *DiskFPSetMXWrapper) GetIndexCnt() int64 {
	if w == nil || w.FPSet == nil {
		return 0
	}
	return w.FPSet.GetIndexCapacity()
}

func (w *DiskFPSetMXWrapper) GetIndexCapacity() int64 {
	return w.GetIndexCnt()
}

func (w *DiskFPSetMXWrapper) GetDiskLookupCnt() uint64 {
	if w == nil || w.FPSet == nil {
		return 0
	}
	return w.FPSet.GetDiskLookupCnt()
}

func (w *DiskFPSetMXWrapper) GetMemHitCnt() uint64 {
	if w == nil || w.FPSet == nil {
		return 0
	}
	return w.FPSet.GetMemHitCnt()
}

func (w *DiskFPSetMXWrapper) GetDiskHitCnt() uint64 {
	if w == nil || w.FPSet == nil {
		return 0
	}
	return w.FPSet.GetDiskHitCnt()
}

func (w *DiskFPSetMXWrapper) GetDiskWriteCnt() uint64 {
	if w == nil || w.FPSet == nil {
		return 0
	}
	return w.FPSet.GetDiskWriteCnt()
}

func (w *DiskFPSetMXWrapper) GetDiskSeekCnt() uint64 {
	if w == nil || w.FPSet == nil {
		return 0
	}
	return w.FPSet.GetDiskSeekCnt()
}

func (w *DiskFPSetMXWrapper) GetDiskSeekCache() uint64 {
	if w == nil || w.FPSet == nil {
		return 0
	}
	return w.FPSet.GetDiskSeekCache()
}

func (w *DiskFPSetMXWrapper) GetDiskSeekRate() float64 {
	diskSeekCnt := w.GetDiskSeekCnt()
	diskSeekCache := w.GetDiskSeekCache()
	return float64(diskSeekCache) / float64(diskSeekCache+diskSeekCnt)
}

func (w *DiskFPSetMXWrapper) GetGrowDiskMark() int {
	if w == nil || w.FPSet == nil {
		return 0
	}
	return w.FPSet.GetGrowDiskMark()
}

func (w *DiskFPSetMXWrapper) GetCheckPointMark() int {
	if w == nil || w.FPSet == nil {
		return 0
	}
	return w.FPSet.GetCheckPointMark()
}

func (w *DiskFPSetMXWrapper) GetBucketCapacity() int64 {
	if w == nil || w.FPSet == nil {
		return 0
	}
	return w.FPSet.GetBucketCapacity()
}

func (w *DiskFPSetMXWrapper) GetTblCapacity() int64 {
	if w == nil || w.FPSet == nil {
		return 0
	}
	return w.FPSet.GetTblCapacity()
}

func (w *DiskFPSetMXWrapper) GetOverallCapacity() int64 {
	if w == nil || w.FPSet == nil {
		return 0
	}
	return w.FPSet.GetOverallCapacity()
}

func (w *DiskFPSetMXWrapper) GetTblLoad() int64 {
	if w == nil || w.FPSet == nil {
		return 0
	}
	return w.FPSet.GetTblLoad()
}

func (w *DiskFPSetMXWrapper) GetMaxTblCnt() int64 {
	if w == nil || w.FPSet == nil {
		return 0
	}
	return w.FPSet.GetMaxTblCnt()
}

func (w *DiskFPSetMXWrapper) GetSizeOf() uint64 {
	if w == nil || w.FPSet == nil {
		return 0
	}
	return w.FPSet.Sizeof()
}

func (w *DiskFPSetMXWrapper) GetFlushTime() int64 {
	if w == nil || w.FPSet == nil {
		return 0
	}
	return w.FPSet.GetFlushTime()
}

func (w *DiskFPSetMXWrapper) GetReaderWriterCnt() int {
	if w == nil || w.FPSet == nil {
		return 0
	}
	return w.FPSet.GetReaderWriterCnt()
}

func (w *DiskFPSetMXWrapper) GetLoadFactor() float64 {
	if w == nil || w.FPSet == nil {
		return 0
	}
	return w.FPSet.GetLoadFactor()
}

func (w *DiskFPSetMXWrapper) ForceFlush() {
	if w != nil && w.FPSet != nil {
		w.FPSet.ForceFlush()
	}
}

func (w *DiskFPSetMXWrapper) CheckInvariant() bool {
	if w == nil || w.FPSet == nil {
		return true
	}
	return w.FPSet.CheckInvariant()
}

func (w *DiskFPSetMXWrapper) GetLockCnt() int {
	if w == nil || w.FPSet == nil {
		return 0
	}
	return w.FPSet.GetLockCnt()
}

type BucketStatisticsMXWrapper struct {
	*TLCStandardMBean
	Stats      any
	ObjectName string
}

func NewBucketStatisticsMXWrapper(stats any, graphName string, pkg string) *BucketStatisticsMXWrapper {
	wrapper := &BucketStatisticsMXWrapper{
		TLCStandardMBean: NewTLCStandardMBean(),
		Stats:            stats,
		ObjectName:       graphName,
	}
	wrapper.RegisterMBean(pkg + ":type=" + graphName)
	return wrapper
}

func (w *BucketStatisticsMXWrapper) GetObjectName() string {
	if w == nil {
		return ""
	}
	return w.ObjectName
}

func (w *BucketStatisticsMXWrapper) GetObservations() int64 {
	return bucketMXObservations(w.bucketStats())
}

func (w *BucketStatisticsMXWrapper) GetMedian() int {
	return bucketMXMedian(w.bucketStats())
}

func (w *BucketStatisticsMXWrapper) GetMean() float64 {
	return bucketMXMean(w.bucketStats())
}

func (w *BucketStatisticsMXWrapper) GetMin() int {
	return bucketMXMin(w.bucketStats())
}

func (w *BucketStatisticsMXWrapper) GetMax() int {
	return bucketMXMax(w.bucketStats())
}

func (w *BucketStatisticsMXWrapper) GetStdDev() float64 {
	return bucketMXStdDev(w.bucketStats())
}

func (w *BucketStatisticsMXWrapper) Get75Percentile() float64 {
	return bucketMXPercentile(w.bucketStats(), 0.75)
}

func (w *BucketStatisticsMXWrapper) Get95Percentile() float64 {
	return bucketMXPercentile(w.bucketStats(), 0.95)
}

func (w *BucketStatisticsMXWrapper) Get98Percentile() float64 {
	return bucketMXPercentile(w.bucketStats(), 0.98)
}

func (w *BucketStatisticsMXWrapper) Get99Percentile() float64 {
	return bucketMXPercentile(w.bucketStats(), 0.99)
}

func (w *BucketStatisticsMXWrapper) Get999Percentile() float64 {
	return bucketMXPercentile(w.bucketStats(), 0.999)
}

func (w *BucketStatisticsMXWrapper) bucketStats() any {
	if w == nil {
		return nil
	}
	return w.Stats
}

func bucketMXObservations(stats any) int64 {
	switch s := stats.(type) {
	case *BucketStatistics:
		return s.GetObservations()
	case *FixedSizedBucketStatistics:
		return s.GetObservations()
	case *ConcurrentBucketStatistics:
		return s.GetObservations()
	case *FixedSizedConcurrentBucketStatistics:
		return s.GetObservations()
	case *DummyBucketStatistics:
		return s.GetObservations()
	default:
		return 0
	}
}

func bucketMXMedian(stats any) int {
	switch s := stats.(type) {
	case *BucketStatistics:
		return s.GetMedian()
	case *FixedSizedBucketStatistics:
		return s.GetMedian()
	case *ConcurrentBucketStatistics:
		return s.GetMedian()
	case *FixedSizedConcurrentBucketStatistics:
		return s.GetMedian()
	case *DummyBucketStatistics:
		return s.GetMedian()
	default:
		return 0
	}
}

func bucketMXMean(stats any) float64 {
	switch s := stats.(type) {
	case *BucketStatistics:
		return s.GetMean()
	case *FixedSizedBucketStatistics:
		return s.GetMean()
	case *ConcurrentBucketStatistics:
		return s.GetMean()
	case *FixedSizedConcurrentBucketStatistics:
		return s.GetMean()
	case *DummyBucketStatistics:
		return s.GetMean()
	default:
		return 0
	}
}

func bucketMXMin(stats any) int {
	switch s := stats.(type) {
	case *BucketStatistics:
		return s.GetMin()
	case *FixedSizedBucketStatistics:
		return s.GetMin()
	case *ConcurrentBucketStatistics:
		return s.GetMin()
	case *FixedSizedConcurrentBucketStatistics:
		return s.GetMin()
	case *DummyBucketStatistics:
		return s.GetMin()
	default:
		return 0
	}
}

func bucketMXMax(stats any) int {
	switch s := stats.(type) {
	case *BucketStatistics:
		return s.GetMax()
	case *FixedSizedBucketStatistics:
		return s.GetMax()
	case *ConcurrentBucketStatistics:
		return s.GetMax()
	case *FixedSizedConcurrentBucketStatistics:
		return s.GetMax()
	case *DummyBucketStatistics:
		return s.GetMax()
	default:
		return 0
	}
}

func bucketMXStdDev(stats any) float64 {
	switch s := stats.(type) {
	case *BucketStatistics:
		return s.GetStdDev()
	case *FixedSizedBucketStatistics:
		return s.GetStdDev()
	case *ConcurrentBucketStatistics:
		return s.GetStdDev()
	case *FixedSizedConcurrentBucketStatistics:
		return s.GetStdDev()
	case *DummyBucketStatistics:
		return s.GetStdDev()
	default:
		return 0
	}
}

func bucketMXPercentile(stats any, quantile float64) float64 {
	switch s := stats.(type) {
	case *BucketStatistics:
		return s.GetPercentile(quantile)
	case *FixedSizedBucketStatistics:
		return s.GetPercentile(quantile)
	case *ConcurrentBucketStatistics:
		return s.GetPercentile(quantile)
	case *FixedSizedConcurrentBucketStatistics:
		return s.GetPercentile(quantile)
	case *DummyBucketStatistics:
		return s.GetPercentile(quantile)
	default:
		return 0
	}
}
