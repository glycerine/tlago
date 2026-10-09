package tlc

import (
	"fmt"
	"io"
	"math"
	"net"
	"net/netip"
	"net/url"
	"os"
	"strings"
	"time"
)

// TLCServerLookup discovers a coordinator endpoint. Adapters preserve the
// failure categories used by the original retry loop.
type TLCServerLookup func(url string) (DistributedServerEndpoint, error)

// DistributedLookupSleep represents Thread.sleep, including interruption.
type DistributedLookupSleep func(time.Duration) error

// LookupTLCWorkerServer ports TLCWorker.main's pre-bootstrap lookup loop.
func LookupTLCWorkerServer(serverName string, lookup TLCServerLookup, sleep DistributedLookupSleep, output io.Writer) (DistributedServerEndpoint, error) {
	server, _, err := DiscoverTLCWorkerServer(serverName, lookup, sleep, output)
	return server, err
}

// DiscoverTLCWorkerServer also retains the URL local used later by main's
// keepalive task, even if Port changes while discovery/bootstrap is running.
func DiscoverTLCWorkerServer(serverName string, lookup TLCServerLookup, sleep DistributedLookupSleep, output io.Writer) (DistributedServerEndpoint, string, error) {
	url := distributedCoordinatorLocation(serverName, TLCServerWorkerName)
	server, err := lookupDistributedServerURL(serverName, url, lookup, sleep, output)
	return server, url, err
}

// LookupDistributedFPServer ports DistributedFPSet.lookupTLCServer.
func LookupDistributedFPServer(serverName string, lookup TLCServerLookup, sleep DistributedLookupSleep, output io.Writer) (DistributedServerEndpoint, error) {
	return lookupDistributedServer(serverName, TLCServerName, lookup, sleep, output)
}

func lookupDistributedServer(serverName, binding string, lookup TLCServerLookup, sleep DistributedLookupSleep, output io.Writer) (DistributedServerEndpoint, error) {
	url := distributedCoordinatorLocation(serverName, binding)
	return lookupDistributedServerURL(serverName, url, lookup, sleep, output)
}

// Native hosts can be bare or bracketed IP literals, including IPv6 zones.
// Retain the supplied spelling; parsing determines the address structure only.
func distributedIPHost(host string) (string, bool) {
	candidate := host
	if strings.HasPrefix(candidate, "[") && strings.HasSuffix(candidate, "]") {
		candidate = candidate[1 : len(candidate)-1]
	}
	if _, err := netip.ParseAddr(candidate); err != nil {
		return host, false
	}
	return candidate, true
}

func distributedWildcardHost(host string) bool {
	address, err := netip.ParseAddr(host)
	return err == nil && address.WithZone("").Unmap().IsUnspecified()
}

func distributedCoordinatorLocation(serverName, binding string) string {
	port := fmtInt(TLCServerPort())
	if host, ip := distributedIPHost(serverName); ip && strings.Contains(host, ":") {
		// URL.String escapes the zone's percent sign for discovery; parsing
		// restores the raw zone before net.Dial receives the TCP address.
		return (&url.URL{Host: net.JoinHostPort(host, port), Path: "/" + binding}).String()
	}
	return "//" + serverName + ":" + port + "/" + binding
}

func lookupDistributedServerURL(serverName, url string, lookup TLCServerLookup, sleep DistributedLookupSleep, output io.Writer) (DistributedServerEndpoint, error) {
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
		case *DistributedOperationError:
			if !failure.DiscoveryRetry {
				return nil, err
			}
			reachable = failure.Reachable
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

func invokeTLCServerLookup(lookup TLCServerLookup, url string) (server DistributedServerEndpoint, err error) {
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
