package tlago

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/glycerine/tlago/tlc"
)

// Each endpoint holds its own real coordinator-assigned block. A shared tool
// evaluation gate would serialize the markers and could not prove that both
// workers have outstanding RPC calls when their process is killed.
type nativeDistributedAssignedBlockWorker struct {
	*tlc.LocalWorkerEndpoint
}

func (e *nativeDistributedAssignedBlockWorker) GetNextStates(states []*tlc.TLCStateMut) (*tlc.NextStateResult, error) {
	fmt.Printf("NATIVE_WORKER_BLOCK_ASSIGNED URI=%s COUNT=%d\n", e.Worker.GetURI(), len(states))
	<-make(chan struct{})
	return nil, fmt.Errorf("assigned-block gate unexpectedly released")
}

func nativeDistributedAssignedBlockWorkers(args []string) error {
	args, err := tlc.ExtractDistributedStartupProperties(args)
	if err != nil {
		return err
	}
	network, err := tlc.NewDistributedWorkerNetwork("127.0.0.1:0", "127.0.0.1")
	if err != nil {
		return err
	}
	defer network.Close()
	process := tlc.NewDistributedWorkerProcess()
	env := network.Environment(tlc.DistributedWorkerEnvironment{LoadApp: func(server tlc.DistributedServerEndpoint, resolver *tlc.DistributedFilenameToStreamResolver) (*tlc.TLCApp, error) {
		app, diagnostics, err := loadDistributedEndpointApp(server, resolver, tlc.RuntimeParameters{})
		if err != nil {
			return nil, err
		}
		if app == nil || diagnostics.HasErrors() {
			return nil, fmt.Errorf("worker model failed to load: %v", diagnostics)
		}
		return app, nil
	}})
	publish := env.PublishWorker
	env.PublishWorker = func(worker *tlc.DistributedWorker) error {
		if err := publish(worker); err != nil {
			return err
		}
		uri, err := url.Parse(worker.GetURI())
		if err != nil {
			return err
		}
		name := strings.TrimPrefix(uri.Path, "/")
		network.Host.UnregisterWorker(name)
		return network.Host.RegisterWorker(name, &nativeDistributedAssignedBlockWorker{LocalWorkerEndpoint: tlc.NewLocalWorkerEndpoint(worker)})
	}
	if _, err := RunDistributedWorker(process, args, env, tlc.RuntimeParameters{}); err != nil {
		return err
	}
	if process.Group == nil {
		return fmt.Errorf("worker startup failed")
	}
	return process.Runtime.AwaitTermination()
}
