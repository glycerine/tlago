package tlc

const ModelCheckerMXObjectName = "tlc2.tool:type=ModelChecker"

type TLCStandardMBean struct {
	ObjectName string
	Registered bool
}

func NewTLCStandardMBean() *TLCStandardMBean {
	return &TLCStandardMBean{}
}

func (m *TLCStandardMBean) GetVersion() string {
	return "dev"
}

func (m *TLCStandardMBean) GetRevision() string {
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
	return 0
}

func (w *ModelCheckerMXWrapper) LiveCheck() {
	if w == nil || w.ModelChecker == nil || w.ModelChecker.LiveCheck == nil {
		return
	}
	w.ModelChecker.LiveCheck.ForceCheck()
}

func (w *ModelCheckerMXWrapper) GetCurrentState() string {
	if w == nil || w.ModelChecker == nil || w.ModelChecker.StateQueue == nil {
		return "N/A"
	}
	state := w.ModelChecker.StateQueue.SPeek()
	if state == nil {
		return "N/A"
	}
	return state.String()
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
	if w != nil && w.ModelChecker != nil && w.ModelChecker.StateQueue != nil {
		w.ModelChecker.StateQueue.SuspendAll()
	}
}

func (w *ModelCheckerMXWrapper) Resume() {
	if w != nil && w.ModelChecker != nil && w.ModelChecker.StateQueue != nil {
		w.ModelChecker.StateQueue.ResumeAll()
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
