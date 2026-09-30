package tlc

import "sync/atomic"

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

func (w *TLCServerMXWrapper) GetStatesGenerated() int64 {
	if w == nil || w.Server == nil || !w.Server.IsRunning() {
		return -1
	}
	return w.Server.GetStatesGenerated()
}

func (w *TLCServerMXWrapper) GetDistinctStatesGenerated() int64 {
	if w == nil || w.Server == nil || !w.Server.IsRunning() {
		return -1
	}
	if manager := w.Server.GetFPSetManager(); manager != nil {
		return int64(manager.Size())
	}
	return -1
}

func (w *TLCServerMXWrapper) GetStateQueueSize() int64 {
	if w == nil || w.Server == nil {
		return 0
	}
	return w.Server.GetNewStates()
}

func (w *TLCServerMXWrapper) GetStatesGeneratedPerMinute() int64 {
	if w == nil || w.Server == nil {
		return 0
	}
	return w.Server.GetStatesGeneratedPerMinute()
}

func (w *TLCServerMXWrapper) GetDistinctStatesGeneratedPerMinute() int64 {
	if w == nil || w.Server == nil {
		return 0
	}
	return w.Server.GetDistinctStatesGeneratedPerMinute()
}

func (w *TLCServerMXWrapper) GetProgress() int {
	if w == nil || w.Server == nil || !w.Server.IsRunning() || w.Server.Trace == nil {
		return -1
	}
	return w.Server.Trace.GetLevelForReporting()
}

func (w *TLCServerMXWrapper) GetWorkerCount() int {
	if w == nil || w.Server == nil {
		return 0
	}
	return w.Server.GetWorkerCount()
}

func (w *TLCServerMXWrapper) Checkpoint() {
	ForceCheckpoint()
}

func (w *TLCServerMXWrapper) GetAverageBlockCnt() int64 {
	if w == nil || w.Server == nil {
		return 0
	}
	return w.Server.GetAverageBlockCnt()
}

func (w *TLCServerMXWrapper) GetRuntimeRatio() float64 {
	return 0
}

func (w *TLCServerMXWrapper) LiveCheck() {}

func (w *TLCServerMXWrapper) GetCurrentState() string {
	if w == nil || w.Server == nil || w.Server.StateQueue == nil {
		return "N/A"
	}
	state := w.Server.StateQueue.SPeek()
	if state == nil {
		return "N/A"
	}
	return state.String()
}

func (w *TLCServerMXWrapper) GetSpecName() string {
	if w == nil || w.Server == nil || !w.Server.IsRunning() {
		return "N/A"
	}
	return w.Server.GetSpecFileName()
}

func (w *TLCServerMXWrapper) GetModelName() string {
	if w == nil || w.Server == nil || !w.Server.IsRunning() {
		return "N/A"
	}
	return w.Server.GetConfigFileName()
}

func (w *TLCServerMXWrapper) Stop() {
	if w == nil || w.Server == nil {
		return
	}
	w.Server.SetDone()
	if w.Server.StateQueue != nil {
		w.Server.StateQueue.FinishAll()
	}
}

func (w *TLCServerMXWrapper) Suspend() {
	if w != nil && w.Server != nil && w.Server.StateQueue != nil {
		w.Server.StateQueue.SuspendAll()
	}
}

func (w *TLCServerMXWrapper) Resume() {
	if w != nil && w.Server != nil && w.Server.StateQueue != nil {
		w.Server.StateQueue.ResumeAll()
	}
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
