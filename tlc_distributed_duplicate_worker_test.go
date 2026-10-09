package tlago

import (
	"fmt"

	"github.com/glycerine/tlago/tlc"
)

// The parent releases the native host after joining the coordinator. Its
// removed endpoint must remain reachable to report the second exit failure.
func nativeDistributedRegisterWorkerTwice(args []string, releasePath string) error {
	if releasePath == "" {
		return fmt.Errorf("coordinator completion path missing")
	}
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
	register := env.RegisterWorker
	env.RegisterWorker = func(server tlc.DistributedServerEndpoint, worker *tlc.DistributedWorker) error {
		if err := register(server, worker); err != nil {
			return err
		}
		return register(server, worker)
	}
	if _, err := RunDistributedWorker(process, args, env, tlc.RuntimeParameters{}); err != nil {
		return err
	}
	if process.Group == nil {
		return fmt.Errorf("worker startup failed")
	}
	if err := process.Runtime.AwaitTermination(); err != nil {
		return err
	}
	if err := nativeDistributedWaitForFile(releasePath); err != nil {
		return err
	}
	fmt.Println("NATIVE_DUPLICATE_WORKER_COORDINATOR_COMPLETED")
	return nil
}
