package tlc

import (
	"net"
	"strconv"
	"strings"
	"testing"
)

func TestNativeWorkerAdvertisedHostUsesBoundPort(t *testing.T) {
	for _, advertised := range []string{"127.0.0.1", "worker.example", "::1", "[::1]", "fe80::1%eth0", "[fe80::1%eth0]", "worker.example:23456"} {
		network, err := NewDistributedWorkerNetwork("127.0.0.1:0", advertised)
		if err != nil {
			t.Fatal(err)
		}
		host, portText, err := net.SplitHostPort(network.Address)
		port, parseErr := strconv.Atoi(portText)
		if err != nil || parseErr != nil || port <= 0 {
			t.Fatalf("invalid advertised address %q", network.Address)
		}
		if advertised == "worker.example:23456" {
			if host != "worker.example" || port != 23456 {
				t.Fatalf("explicit address changed: %q", network.Address)
			}
		} else {
			if host != strings.TrimSuffix(strings.TrimPrefix(advertised, "["), "]") || network.workerAddress.Port != port {
				t.Fatalf("host/port mismatch: %q", network.Address)
			}
			// Reach the actual listener on loopback even when a different
			// host is advertised for remote peers.
			conn, err := net.Dial("tcp", net.JoinHostPort("127.0.0.1", portText))
			if err != nil {
				t.Fatalf("advertised callback port is not bound: %v", err)
			}
			_ = conn.Close()
		}
		if err := network.Close(); err != nil {
			t.Fatal(err)
		}
	}
}
