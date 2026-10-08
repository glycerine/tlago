package tlc

// DistributedServerEndpoint supplies coordinator calls to workers and
// fingerprint servers. It describes TLC operations, independently of transport.
// Each call can fail; discovery succeeding does not make later calls infallible.
type DistributedServerEndpoint interface {
	DistributedFileServer
	InternSource
	RegisterWorker(DistributedWorkerEndpoint) error
	RegisterFPSet(FPSet, string) error
	GetCheckDeadlock() (bool, error)
	GetPreprocess() (bool, error)
	GetFPSetManager() (*DistributedFPSetManager, error)
	GetIrredPolyForFP() (uint64, error)
	IsDone() (bool, error)
	GetSpecFileName() (string, error)
	GetConfigFileName() (string, error)
}

// LocalServerEndpoint supplies the same calls for an in-process coordinator.
// Pin its interning table before worker bootstrap replaces the package table.
type LocalServerEndpoint struct {
	Server *TLCServer
}

func NewLocalServerEndpoint(server *TLCServer) *LocalServerEndpoint {
	if server == nil {
		panic(NewNullPointerException())
	}
	server.serverInternTable()
	return &LocalServerEndpoint{Server: server}
}

func (e *LocalServerEndpoint) RegisterWorker(worker DistributedWorkerEndpoint) error {
	return e.Server.RegisterWorker(worker)
}

func (e *LocalServerEndpoint) RegisterFPSet(set FPSet, hostname string) error {
	return e.Server.RegisterFPSet(set, hostname)
}

func (e *LocalServerEndpoint) GetCheckDeadlock() (bool, error) {
	return e.Server.GetCheckDeadlock(), nil
}

func (e *LocalServerEndpoint) GetPreprocess() (bool, error) {
	return e.Server.GetPreprocess(), nil
}

func (e *LocalServerEndpoint) GetFPSetManager() (*DistributedFPSetManager, error) {
	return e.Server.GetFPSetManager(), nil
}

func (e *LocalServerEndpoint) GetIrredPolyForFP() (uint64, error) {
	return e.Server.GetIrredPolyForFP(), nil
}

func (e *LocalServerEndpoint) IsDone() (bool, error) {
	return e.Server.IsDone(), nil
}

func (e *LocalServerEndpoint) GetSpecFileName() (string, error) {
	return e.Server.GetSpecFileName(), nil
}

func (e *LocalServerEndpoint) GetConfigFileName() (string, error) {
	return e.Server.GetConfigFileName(), nil
}

func (e *LocalServerEndpoint) GetFile(name string) ([]byte, error) {
	return e.Server.GetFile(name)
}

func (e *LocalServerEndpoint) Intern(str string) (*UniqueString, error) {
	value, err := e.Server.Intern(str)
	if value == nil || err != nil {
		return value, err
	}
	// Worker-owned metadata must not share mutable fields with the coordinator.
	// Preserve the assigned token and location instead of re-interning the reply.
	return &UniqueString{s: value.s, tok: value.tok, loc: value.loc}, nil
}

var _ DistributedServerEndpoint = (*LocalServerEndpoint)(nil)
