package tlc

// DistributedWorkerEndpoint supplies the worker operations used by the coordinator.
// Server threads and smart proxies use the same calls for local and network
// workers. This Go interface does not prescribe a transport or Java RMI protocol.
// The smart proxy measures calls; endpoint adapters own transport failures.
type DistributedWorkerEndpoint interface {
	GetNextStates([]*TLCStateMut) (*NextStateResult, error)
	IsAlive() (bool, error)
	Exit() error
	GetURI() (string, error)
	GetCacheRateRatio() (float64, error)
}

// LocalWorkerEndpoint adapts an in-process worker to the coordinator calls.
// It retains the existing endpoint lifecycle checks and error behavior,
// independently of the smart proxy. It does not open a network connection.
type LocalWorkerEndpoint struct {
	Worker *DistributedWorker
}

func NewLocalWorkerEndpoint(worker *DistributedWorker) *LocalWorkerEndpoint {
	return &LocalWorkerEndpoint{Worker: worker}
}

func (r *LocalWorkerEndpoint) endpointError() error {
	if r == nil || r.Worker == nil {
		return NewNullPointerException()
	}
	return r.Worker.remoteEndpointError()
}

func (r *LocalWorkerEndpoint) GetNextStates(states []*TLCStateMut) (*NextStateResult, error) {
	if err := r.endpointError(); err != nil {
		return nil, err
	}
	return r.Worker.GetNextStates(states)
}

func (r *LocalWorkerEndpoint) IsAlive() (bool, error) {
	if err := r.endpointError(); err != nil {
		return false, err
	}
	return r.Worker.IsAlive(), nil
}

func (r *LocalWorkerEndpoint) Exit() error {
	if err := r.endpointError(); err != nil {
		return err
	}
	return r.Worker.Exit()
}

func (r *LocalWorkerEndpoint) GetURI() (string, error) {
	if err := r.endpointError(); err != nil {
		return "", err
	}
	return r.Worker.GetURI(), nil
}

func (r *LocalWorkerEndpoint) GetCacheRateRatio() (float64, error) {
	if err := r.endpointError(); err != nil {
		return 0, err
	}
	return r.Worker.GetCacheRateRatio(), nil
}

var _ DistributedWorkerEndpoint = (*LocalWorkerEndpoint)(nil)
var _ DistributedWorkerEndpoint = (*DistributedWorkerSmartProxy)(nil)
