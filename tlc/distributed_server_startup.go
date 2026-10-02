package tlc

import (
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

const tlcServerPropertyPrefix = "tlc2.tool.distributed.TLCServer"

var tlcServerProperties struct {
	sync.Once
	port                               atomic.Int32
	reportInterval, expectedFPSetCount int
	vetoCleanup                        bool
}

// InitializeTLCServerProperties captures TLCServer's static initializers. The
// port remains mutable in Java; SetTLCServerPort represents that assignment.
func InitializeTLCServerProperties() {
	tlcServerProperties.Do(func() {
		tlcServerProperties.port.Store(TLCServerDefaultPort)
		tlcServerProperties.reportInterval = 60 * 1000
		if value, ok := distributedIntProperty(tlcServerPropertyPrefix + ".port"); ok {
			tlcServerProperties.port.Store(int32(value))
		}
		if value, ok := distributedIntProperty(tlcServerPropertyPrefix + ".report"); ok {
			tlcServerProperties.reportInterval = value
		}
		if value, ok := distributedIntProperty(tlcServerPropertyPrefix + ".expectedFPSetCount"); ok {
			tlcServerProperties.expectedFPSetCount = value
		}
		if value, ok := tlcLookupSystemProperty(tlcServerVetoCleanup); ok {
			tlcServerProperties.vetoCleanup = javaBooleanProperty(value)
		}
	})
}

func TLCServerPort() int {
	InitializeTLCServerProperties()
	return int(tlcServerProperties.port.Load())
}
func SetTLCServerPort(port int) {
	InitializeTLCServerProperties()
	tlcServerProperties.port.Store(int32(port))
}
func TLCServerExpectedFPSetCount() int {
	InitializeTLCServerProperties()
	return tlcServerProperties.expectedFPSetCount
}
func TLCServerReportIntervalMillis() int {
	InitializeTLCServerProperties()
	return tlcServerProperties.reportInterval
}

// NewTLCServerFromApp is TLCServer(TLCApp). It uses DiskStateQueue directly,
// opens the trace before creating/initializing the local FPSet, and retains the
// application snapshot and its CLI flags. Endpoint export is transport work.
func NewTLCServerFromApp(app *TLCApp) (*TLCServer, error) {
	return newTLCServerFromApp(app, false)
}

// NewDistributedFPSetTLCServer folds the Java subclass into the concrete
// server. The virtual manager factory in super(app) receives the static
// expectedFPSetCount, before the subclass's explicit count is assigned.
func NewDistributedFPSetTLCServer(app *TLCApp, expectedFPSetCount int) (*TLCServer, error) {
	server, err := newTLCServerFromApp(app, true)
	if err != nil {
		return nil, err
	}
	expectedFPSetCount = int(int32(expectedFPSetCount))
	if expectedFPSetCount < 0 {
		panic(NewIllegalArgumentException("count < 0"))
	}
	registration := &distributedFPRegistration{expected: expectedFPSetCount, remaining: expectedFPSetCount, done: make(chan struct{})}
	if expectedFPSetCount == 0 {
		close(registration.done)
	}
	server.fpRegistration = registration
	return server, nil
}

func newTLCServerFromApp(app *TLCApp, distributed bool) (*TLCServer, error) {
	InitializeTLCServerProperties()
	if app == nil {
		failure := newTLCError(ECGeneral, "TLC server found null work.")
		failure.Runtime = true
		panic(failure)
	}
	if !app.metadataSet {
		panic(NewNullPointerException())
	}
	metadir := app.GetMetadir()
	end := len(metadir)
	separator := string(os.PathSeparator)
	if strings.HasSuffix(metadir, separator) {
		end--
	}
	start := strings.LastIndex(metadir[:end], separator)
	checkpointName := metadir[start+1 : end]
	queue := newDiskStateQueue(metadir, true)
	// Unlike the general host-side trace constructor, Java does not create a
	// missing metadata directory here. Failed trace opening precedes the FPSet.
	trace := &TLCTrace{lastPtr: 1, diskdir: metadir, rootName: app.GetFileName(), rawPaths: true, Tool: app.requireTool()}
	raf, err := NewBufferedRandomAccessFile(metadir+separator+app.GetFileName()+tlcTraceExt, "rw")
	if err != nil {
		return nil, distributedFileOpenException(metadir+separator+app.GetFileName()+tlcTraceExt, err)
	}
	trace.raf = raf
	var manager *DistributedFPSetManager
	if distributed {
		manager = NewDynamicDistributedFPSetManager(TLCServerExpectedFPSetCount())
	} else {
		config := app.GetFPSetConfiguration()
		if config == nil {
			panic(NewNullPointerException())
		}
		set := NewFPSet(config)
		if set == nil {
			panic(NewNullPointerException())
		}
		set.Init(1, metadir, app.GetFileName())
		hostname, err := distributedCanonicalLocalHost()
		if err != nil {
			return nil, err
		}
		manager = NewNonDistributedFPSetManager(set, hostname, trace)
	}
	server := NewTLCServer(app.GetFileName(), app.GetConfigName(), metadir, manager, queue, trace)
	server.Tool = app.Tool
	server.app = app
	server.checkDeadlock = &app.checkDeadlock
	server.checkpointName = &checkpointName
	return server, nil
}

func (s *TLCServer) checkpointFileName() string {
	if s.checkpointName != nil {
		return *s.checkpointName
	}
	return s.FileName
}

type distributedFPRegistration struct {
	mu                  sync.Mutex
	expected, remaining int
	done                chan struct{}
}

// WaitForFPSetManager is the protected Java startup hook. Base servers return
// immediately; the subclass prints its waiting message even for count zero.
func (s *TLCServer) WaitForFPSetManager() {
	if s == nil {
		panic(NewNullPointerException())
	}
	if registration := s.fpRegistration; registration != nil {
		PrintMessage(ECTLCDistributedServerFPSetWaiting, fmtInt(registration.expected))
		<-registration.done
	}
}

// RegisterFPSet keeps registration and countDown in one synchronized region.
// Failed/extra registrations do not release the latch or print acceptance.
func (s *TLCServer) RegisterFPSet(set FPSet, hostname string) error {
	if s == nil {
		panic(NewNullPointerException())
	}
	s.monitor.Lock()
	defer s.monitor.Unlock()
	registration := s.fpRegistration
	if registration == nil {
		return s.FPSetManager.RegisterFPSet(set, hostname)
	}
	registration.mu.Lock()
	defer registration.mu.Unlock()
	if err := s.FPSetManager.RegisterFPSet(set, hostname); err != nil {
		return err
	}
	if registration.remaining > 0 {
		registration.remaining--
		if registration.remaining == 0 {
			close(registration.done)
		}
	}
	PrintMessage(ECTLCDistributedServerFPSetRegistered, fmtInt(registration.expected-registration.remaining), fmtInt(registration.expected))
	return nil
}
