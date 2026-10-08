package tlc

import "sync"

// TLCServerStatusLookup is one Naming.lookup followed by server.isDone. Checked
// failures from either call reach TLCTimerTask's same catch region.
type TLCServerStatusLookup func(url string) (bool, error)

var distributedWorkerKeepAliveProperties struct {
	sync.Once
	timeout int32
}

func DistributedWorkerKeepAliveTimeoutMillis() int32 {
	distributedWorkerKeepAliveProperties.Do(func() {
		seconds := 60
		if configured, ok := distributedIntProperty("tlc2.tool.distributed.TLCTimerTask.timeout"); ok {
			seconds = configured
		}
		distributedWorkerKeepAliveProperties.timeout = int32(seconds) * 1000
	})
	return distributedWorkerKeepAliveProperties.timeout
}

// ConfigureKeepAliveLookup supplies the source task's final URL and lookup
// boundary before StartKeepAlive creates it. Later changes leave that task's
// captured fields intact. A nil finest logger preserves JUL's default level.
func (r *DistributedWorkerRuntime) ConfigureKeepAliveLookup(url string, lookup TLCServerStatusLookup, finest func(string, error)) {
	r.keepAliveMu.Lock()
	defer r.keepAliveMu.Unlock()
	r.keepAliveURL, r.statusLookup, r.keepAliveLog = url, lookup, finest
}

// RunKeepAliveOnce is TLCTimerTask's public run invocation at the Go library
// boundary. The scheduler treats returned failures as uncaught timer errors.
func (r *DistributedWorkerRuntime) RunKeepAliveOnce() error {
	if r == nil {
		return NewNullPointerException()
	}
	r.keepAliveMu.Lock()
	task := r.keepAlive
	r.keepAliveMu.Unlock()
	if task == nil {
		return NewNullPointerException()
	}
	return task.run()
}

func invokeTLCServerStatusLookup(lookup TLCServerStatusLookup, url string) (done bool, err error) {
	defer func() {
		if failure := recover(); failure != nil {
			done = false
			err = panicValueAsError(failure)
		}
	}()
	if lookup == nil {
		return false, NewNullPointerException()
	}
	return lookup(url)
}

// NewTLCServerStatusLookup keeps registry discovery and the done call together
// while preserving a null lookup result's NPE instead of treating it as done.
func NewTLCServerStatusLookup(lookup TLCServerLookup) TLCServerStatusLookup {
	return func(url string) (bool, error) {
		server, err := invokeTLCServerLookup(lookup, url)
		if err != nil {
			return false, err
		}
		if server == nil {
			return false, NewNullPointerException()
		}
		return server.IsDone()
	}
}
