package tlc

import "sync"

// TLCServerRegistry supplies the three coordinator catalog operations used by
// TLCServer. Native failure categories preserve discovery and cleanup decisions.
type TLCServerRegistry struct {
	Rebind func(string, *TLCServer) error
	Unbind func(string) error
	Lookup func(string) (*TLCServer, error)
}

// TLCServerPublication supplies modelCheck's naming/export boundaries and the
// shutdown hook's independent LocateRegistry.getRegistry lookup. Configure it
// before model checking. The default represents local naming, without a wire
// listener or remote object serialization.
type TLCServerPublication struct {
	LocalHostName  func() (string, error)
	CreateRegistry func(int) (*TLCServerRegistry, error)
	GetRegistry    func(int) (*TLCServerRegistry, error)
	Unexport       func(*TLCServer, bool) (bool, error)
	Flush          func()
}

func (s *TLCServer) ConfigurePublication(publication TLCServerPublication) {
	s.publication = publication
}

var defaultTLCRegistryNamespace = NewTLCRegistryNamespace()

// TLCRegistryNamespace represents local registry object identity and binding
// lifetime. It is not a network transport. Separate namespaces separate names,
// but do not isolate the evaluator's package globals.
type TLCRegistryNamespace struct {
	mu         sync.Mutex
	registries *InsMap[int, *TLCServerRegistry]
}

func NewTLCRegistryNamespace() *TLCRegistryNamespace {
	return &TLCRegistryNamespace{registries: NewInsMap[int, *TLCServerRegistry]()}
}

func (n *TLCRegistryNamespace) CreateRegistry(port int) (*TLCServerRegistry, error) {
	if port < 0 || port > 65535 {
		return nil, NewIllegalArgumentException("Port value out of range: " + fmtInt(port))
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	// Local catalogs are keyed by the configured port, including zero. Creating
	// an existing key fails without replacing its catalog or inventing a listener.
	if _, found := n.registries.Get2(port); found {
		return nil, coordinatorPublicationFailure("coordinator catalog already exists at port " + fmtInt(port))
	}
	var bindingsMu sync.Mutex
	bindings := NewInsMap[string, *TLCServer]()
	registry := &TLCServerRegistry{
		Rebind: func(name string, server *TLCServer) error {
			if server == nil {
				return NewNullPointerException()
			}
			bindingsMu.Lock()
			defer bindingsMu.Unlock()
			bindings.Set(name, server)
			return nil
		},
		Unbind: func(name string) error {
			bindingsMu.Lock()
			defer bindingsMu.Unlock()
			if found, _ := bindings.Delkey(name); !found {
				return coordinatorBindingMissingFailure(name)
			}
			return nil
		},
		Lookup: func(name string) (*TLCServer, error) {
			bindingsMu.Lock()
			defer bindingsMu.Unlock()
			if server, found := bindings.Get2(name); found {
				return server, nil
			}
			return nil, coordinatorBindingMissingFailure(name)
		},
	}
	n.registries.Set(port, registry)
	return registry, nil
}

// GetRegistry returns a lazy local reference like LocateRegistry.getRegistry:
// connection failure happens on a registry operation, not reference creation.
func (n *TLCRegistryNamespace) GetRegistry(port int) (*TLCServerRegistry, error) {
	if port <= 0 {
		port = 1099
	}
	resolve := func() (*TLCServerRegistry, error) {
		if port > 65535 {
			return nil, NewIllegalArgumentException("port out of range:" + fmtInt(port))
		}
		n.mu.Lock()
		registry, found := n.registries.Get2(port)
		n.mu.Unlock()
		if !found {
			return nil, &DistributedOperationError{Message: javaString("coordinator catalog is unavailable at port " + fmtInt(port)), Class: "tlc.CoordinatorUnavailable", Remote: true, IO: true, DiscoveryRetry: true}
		}
		return registry, nil
	}
	return &TLCServerRegistry{
		Lookup: func(name string) (*TLCServer, error) {
			registry, err := resolve()
			if err != nil {
				return nil, err
			}
			return registry.Lookup(name)
		},
		Rebind: func(name string, server *TLCServer) error {
			registry, err := resolve()
			if err != nil {
				return err
			}
			return registry.Rebind(name, server)
		},
		Unbind: func(name string) error {
			registry, err := resolve()
			if err != nil {
				return err
			}
			return registry.Unbind(name)
		},
	}, nil
}

func (n *TLCRegistryNamespace) Publication() TLCServerPublication {
	return TLCServerPublication{CreateRegistry: n.CreateRegistry, GetRegistry: n.GetRegistry}
}

func (s *TLCServer) publicationBoundaries() TLCServerPublication {
	env := s.publication
	if env.LocalHostName == nil {
		env.LocalHostName = distributedLocalHostName
	}
	if env.CreateRegistry == nil {
		env.CreateRegistry = defaultTLCRegistryNamespace.CreateRegistry
	}
	if env.GetRegistry == nil {
		env.GetRegistry = defaultTLCRegistryNamespace.GetRegistry
	}
	if env.Unexport == nil {
		env.Unexport = func(server *TLCServer, force bool) (bool, error) {
			if !server.unexported.CompareAndSwap(false, true) {
				return false, coordinatorEndpointRemovedFailure()
			}
			return true, nil
		}
	}
	return env
}

func invokeRegistryBoundary[T any](call func() (T, error)) (value T, err error) {
	defer func() {
		if failure := recover(); failure != nil {
			err = panicValueAsError(failure)
		}
	}()
	return call()
}

func invokeRegistryOperation(call func() error) error {
	_, err := invokeRegistryBoundary(func() (struct{}, error) { return struct{}{}, call() })
	return err
}

// RunWorkerShutdownHook preserves the lookup guard, one registry check before
// iteration, and the source direct dead-worker catches. It neither clears
// registrations nor closes/unexports the server or registry.
func (s *TLCServer) RunWorkerShutdownHook() error {
	s.threadsMu.Lock()
	empty := s.threadsToWorkers.Len() == 0
	s.threadsMu.Unlock()
	if empty {
		return nil
	}
	env := s.publicationBoundaries()
	_, err := invokeRegistryBoundary(func() (*TLCServer, error) {
		registry, err := env.GetRegistry(TLCServerPort())
		if err != nil {
			return nil, err
		}
		if registry == nil || registry.Lookup == nil {
			return nil, NewNullPointerException()
		}
		return registry.Lookup(TLCServerName)
	})
	if err != nil {
		if isDistributedRemoteFailure(err) {
			return nil
		}
		if isDistributedCoordinatorBindingMissing(err) {
			return nil
		}
		return err
	}
	s.threadsMu.Lock()
	workers := make([]DistributedWorkerEndpoint, 0, s.threadsToWorkers.Len())
	for _, worker := range s.threadsToWorkers.All() {
		workers = append(workers, worker)
	}
	s.threadsMu.Unlock()
	for _, worker := range workers {
		err := invokeRegistryOperation(func() error {
			if worker == nil {
				return NewNullPointerException()
			}
			return worker.Exit()
		})
		if err == nil {
			continue
		}
		if isDistributedWorkerUnavailable(err) {
			continue
		}
		if isJavaIOException(err) {
			PrintErrorThrowable(ECGeneral, err)
			continue
		}
		return err
	}
	return nil
}
