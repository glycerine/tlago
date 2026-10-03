package tlc

import (
	"os"
	"runtime"
)

// DistributedServerEnvironment supplies TLCServer.main's process boundaries.
// Model properties load before application creation; normal console streams
// remain installed.
// The defaults use the native server/checker and local management bean.
type DistributedServerEnvironment struct {
	LoadProperties      func()
	CreateApp           func([]string) (*TLCApp, error)
	Property            func(string, string) string
	CreateServer        func(*TLCApp, int) (*TLCServer, error)
	CreateMBean         func(*TLCServer) (*TLCStandardMBean, error)
	InstallShutdownHook func(func() error) error
	ModelCheck          func(*TLCServer) error
	GC                  func()
	Close               func(*TLCServer, bool) error
	ShutdownNow         func(*TLCServer) error
	Unregister          func(*TLCStandardMBean) (bool, error)
	ModuleFiles         func(*TLCApp) ([]*TLAFile, error)
	Exit                func(int)
}

// DistributedServerProcess retains main's locals for native callers and the
// worker hook callbacks registered during its successful construction. A hook
// captures that invocation's server, even if Run is called again later.
type DistributedServerProcess struct {
	App           *TLCApp
	Server        *TLCServer
	MBean         *TLCStandardMBean
	ShutdownHooks []func() error
}

func NewDistributedServerProcess() *DistributedServerProcess {
	InitializeTLCServerProperties()
	return &DistributedServerProcess{}
}

// Run preserves main's try/catch/finally, including the source nil-server
// dereference in finally. A returned error escapes the Java main; reported
// startup/model errors themselves do not create a command exit status.
func (p *DistributedServerProcess) Run(args []string, env DistributedServerEnvironment) (err error) {
	defer func() {
		if failure := recover(); failure != nil {
			err = panicValueAsError(failure)
		}
	}()
	if p == nil {
		return NewNullPointerException()
	}
	InitializeTLCServerProperties()
	PrintMessage(ECTLCVersion, "TLC Server "+TLCVersion())
	p.MBean = NewNullTLCStandardMBean()
	p.App, p.Server = nil, nil
	env = distributedServerEnvironment(env)
	pending := invokeDistributedServerOperation(func() error {
		if failure := p.start(args, env); failure != nil {
			env.GC()
			printDistributedServerFailure(failure)
			if p.Server != nil {
				closeErr := invokeDistributedServerOperation(func() error { return env.Close(p.Server, false) })
				if closeErr != nil {
					// Source catches Exception here, but Error still escapes through finally.
					if isJavaError(closeErr) {
						return closeErr
					}
					PrintError(ECGeneral, javaGeneralErrorMessage("", closeErr))
				}
			}
		}
		return nil
	})
	if finalErr := invokeDistributedServerOperation(func() error { return p.finish(env) }); finalErr != nil {
		return finalErr
	}
	return pending
}
func distributedServerEnvironment(env DistributedServerEnvironment) DistributedServerEnvironment {
	if env.LoadProperties == nil {
		env.LoadProperties = func() { NewModelInJar().LoadProperties() }
	}
	if env.Property == nil {
		env.Property = func(key, fallback string) string {
			if value, ok := tlcLookupSystemProperty(key); ok {
				return value
			}
			return fallback
		}
	}
	if env.CreateServer == nil {
		env.CreateServer = func(app *TLCApp, count int) (*TLCServer, error) {
			if count > 0 {
				return NewDistributedFPSetTLCServer(app, count)
			}
			return NewTLCServerFromApp(app)
		}
	}
	if env.CreateMBean == nil {
		env.CreateMBean = func(server *TLCServer) (*TLCStandardMBean, error) {
			return NewTLCServerMXWrapper(server).TLCStandardMBean, nil
		}
	}
	if env.ModelCheck == nil {
		env.ModelCheck = func(server *TLCServer) error { _, err := server.ModelCheck(); return err }
	}
	if env.GC == nil {
		env.GC = runtime.GC
	}
	if env.Close == nil {
		env.Close = func(server *TLCServer, clean bool) error { return server.Close(clean) }
	}
	if env.ShutdownNow == nil {
		env.ShutdownNow = func(server *TLCServer) error { server.executor.ShutdownNow(); return nil }
	}
	if env.Unregister == nil {
		env.Unregister = func(bean *TLCStandardMBean) (bool, error) {
			if bean == nil {
				return false, NewNullPointerException()
			}
			return bean.Unregister(), nil
		}
	}
	if env.ModuleFiles == nil {
		env.ModuleFiles = func(app *TLCApp) ([]*TLAFile, error) { return app.GetModuleFiles(), nil }
	}
	if env.Exit == nil {
		env.Exit = os.Exit
	}
	return env
}
func (p *DistributedServerProcess) start(args []string, env DistributedServerEnvironment) (err error) {
	defer func() {
		if failure := recover(); failure != nil {
			err = panicValueAsError(failure)
		}
	}()
	SetNumWorkers(0)
	env.LoadProperties()
	if env.CreateApp == nil {
		return NewNullPointerException()
	}
	app, err := env.CreateApp(args)
	if err != nil {
		return err
	}
	p.App = app
	if app == nil {
		return NewNullPointerException()
	}
	count := TLCServerExpectedFPSetCount()
	if count <= 0 {
		count = 0 // Source selects the ordinary constructor with no count arg.
	}
	server, err := env.CreateServer(app, count)
	if err != nil {
		return err
	}
	p.Server = server
	bean, err := env.CreateMBean(server)
	if err != nil {
		return err
	}
	p.MBean = bean
	if server != nil {
		hook := func() error { return server.RunWorkerShutdownHook() }
		if env.InstallShutdownHook != nil {
			if err := env.InstallShutdownHook(hook); err != nil {
				return err
			}
		}
		p.ShutdownHooks = append(p.ShutdownHooks, hook)
		return env.ModelCheck(server)
	}
	return nil
}
func (p *DistributedServerProcess) finish(env DistributedServerEnvironment) error {
	// Deliberately preserve Java's server.es dereference, even after a startup
	// failure. This failure supersedes a Throwable pending in the catch region.
	if p.Server == nil {
		return NewNullPointerException()
	}
	if !p.Server.executor.IsShutdown() {
		if err := env.ShutdownNow(p.Server); err != nil {
			return err
		}
	}
	if _, err := env.Unregister(p.MBean); err != nil {
		return err
	}
	return nil
}
func printDistributedServerFailure(err error) {
	if javaSystemFailureCode(err) == ECSystemStackOverflow {
		PrintErrorNullable(ECSystemStackOverflow, javaThrowableDetailMessage(err))
		return
	}
	if isJavaOutOfMemoryError(err) {
		PrintErrorNullable(ECSystemOutOfMemory, javaThrowableDetailMessage(err))
		return
	}
	PrintError(ECGeneral, javaGeneralErrorMessage("", err))
}
func invokeDistributedServerOperation(call func() error) (err error) {
	defer func() {
		if failure := recover(); failure != nil {
			err = panicValueAsError(failure)
		}
	}()
	return call()
}
