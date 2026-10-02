package tlc

import (
	"fmt"
	"io"
	"math"
	"os"
	"time"
)

// TLCServerLookup is the Naming.lookup boundary. The registry/RPC adapter must
// return the concrete Java exception families inspected by the source loop.
type TLCServerLookup func(url string) (*TLCServer, error)

// DistributedLookupSleep represents Thread.sleep, including interruption.
type DistributedLookupSleep func(time.Duration) error

// LookupTLCWorkerServer ports TLCWorker.main's pre-bootstrap lookup loop.
func LookupTLCWorkerServer(serverName string, lookup TLCServerLookup, sleep DistributedLookupSleep, output io.Writer) (*TLCServer, error) {
	server, _, err := DiscoverTLCWorkerServer(serverName, lookup, sleep, output)
	return server, err
}

// DiscoverTLCWorkerServer also retains the URL local used later by main's
// keepalive task, even if Port changes while discovery/bootstrap is running.
func DiscoverTLCWorkerServer(serverName string, lookup TLCServerLookup, sleep DistributedLookupSleep, output io.Writer) (*TLCServer, string, error) {
	url := "//" + serverName + ":" + fmtInt(TLCServerPort()) + "/" + TLCServerWorkerName
	server, err := lookupDistributedServerURL(serverName, url, lookup, sleep, output)
	return server, url, err
}

// LookupDistributedFPServer ports DistributedFPSet.lookupTLCServer.
func LookupDistributedFPServer(serverName string, lookup TLCServerLookup, sleep DistributedLookupSleep, output io.Writer) (*TLCServer, error) {
	return lookupDistributedServer(serverName, TLCServerName, lookup, sleep, output)
}

func lookupDistributedServer(serverName, binding string, lookup TLCServerLookup, sleep DistributedLookupSleep, output io.Writer) (*TLCServer, error) {
	url := "//" + serverName + ":" + fmtInt(TLCServerPort()) + "/" + binding
	return lookupDistributedServerURL(serverName, url, lookup, sleep, output)
}

func lookupDistributedServerURL(serverName, url string, lookup TLCServerLookup, sleep DistributedLookupSleep, output io.Writer) (*TLCServer, error) {
	if sleep == nil {
		sleep = func(duration time.Duration) error { time.Sleep(duration); return nil }
	}
	if output == nil {
		output = os.Stdout
	}
	var attempt int32 = 1
	for {
		server, err := invokeTLCServerLookup(lookup, url)
		if err == nil {
			return server, nil
		}
		reachable := false
		switch failure := err.(type) {
		case *ConnectException:
			// Java inspects the immediate cause only, and distinguishes the
			// java.net exception from java.rmi.ConnectException.
			if failure == nil || failure.RemoteException == nil {
				return nil, err
			}
			if cause, ok := failure.GetCause().(*NetConnectException); !ok || cause == nil {
				return nil, err
			}
		case *NotBoundException:
			if failure == nil {
				return nil, err
			}
			reachable = true
		default:
			return nil, err
		}
		seconds := javaDoubleToLong(math.Sqrt(float64(attempt)))
		if reachable {
			fmt.Fprintf(output, "Server %s reachable but not ready yet, sleeping %ds for server to come online...\n", serverName, seconds)
		} else {
			fmt.Fprintf(output, "Server %s unreachable, sleeping %ds for server to come online...\n", serverName, seconds)
		}
		if err := sleep(time.Duration(seconds) * time.Second); err != nil {
			return nil, err
		}
		// Java int multiplication wraps; sqrt of the negative value becomes
		// NaN, whose long conversion is zero, followed by permanent zero.
		attempt *= 2
	}
}

func invokeTLCServerLookup(lookup TLCServerLookup, url string) (server *TLCServer, err error) {
	defer func() {
		if failure := recover(); failure != nil {
			server = nil
			err = panicValueAsError(failure)
		}
	}()
	if lookup == nil {
		return nil, NewNullPointerException()
	}
	return lookup(url)
}
