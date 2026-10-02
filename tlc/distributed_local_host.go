package tlc

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type distributedLocalHost struct {
	address   net.IP
	expires   time.Time
	canonical atomic.Pointer[string]
}

var distributedCachedLocalHost atomic.Pointer[distributedLocalHost]
var distributedHostPolicy struct {
	sync.Once
	ipv4Only, ipv6First, systemOrder bool
}

// The native lookups use Go's OS resolver. InetAddress's local-host cache,
// address preference and canonical reverse/forward verification are retained;
// general JVM resolver providers and security-manager hooks are not represented.
func distributedCanonicalLocalHost() (string, error) {
	distributedHostPolicy.Do(func() {
		ipv4, _ := tlcLookupSystemProperty("java.net.preferIPv4Stack")
		distributedHostPolicy.ipv4Only = ipv4 == "true"
		preference, _ := tlcLookupSystemProperty("java.net.preferIPv6Addresses")
		distributedHostPolicy.ipv6First = strings.EqualFold(preference, "true")
		distributedHostPolicy.systemOrder = strings.EqualFold(preference, "system")
	})
	local := distributedCachedLocalHost.Load()
	if local == nil || time.Now().After(local.expires) {
		host, err := os.Hostname()
		if err != nil {
			return "", NewUnknownHostException(err.Error())
		}
		var address net.IP
		if host == "localhost" {
			// Inet6AddressImpl prefers IPv6 loopback for IPv6-first/system
			// policy, but uses the other family if only it is bound locally.
			address = net.IPv4(127, 0, 0, 1)
			if !distributedHostPolicy.ipv4Only && (distributedHostPolicy.ipv6First || distributedHostPolicy.systemOrder) {
				address = net.IPv6loopback
			}
			other := net.IPv6loopback
			if address.To4() == nil {
				other = net.IPv4(127, 0, 0, 1)
			}
			if interfaces, err := net.InterfaceAddrs(); err == nil {
				preferredBound, otherBound := false, false
				for _, entry := range interfaces {
					if ip, _, err := net.ParseCIDR(entry.String()); err == nil {
						preferredBound = preferredBound || ip.Equal(address)
						otherBound = otherBound || ip.Equal(other)
					}
				}
				if !preferredBound && otherBound && !distributedHostPolicy.ipv4Only {
					address = other
				}
			}
		} else {
			addresses, err := net.LookupIP(host)
			if err == nil {
				address = distributedPreferredAddress(addresses)
			}
			if address == nil {
				message := host
				if err != nil {
					message = err.Error()
				}
				cause := NewUnknownHostException(message)
				failure := NewUnknownHostException(host + ": " + message)
				failure.Cause = cause
				return "", failure
			}
		}
		local = &distributedLocalHost{address: address, expires: time.Now().Add(5 * time.Second)}
		distributedCachedLocalHost.Store(local)
	}
	if canonical := local.canonical.Load(); canonical != nil {
		return *canonical, nil
	}
	canonical := distributedNumericAddress(local.address)
	if names, err := net.LookupAddr(local.address.String()); err == nil && len(names) > 0 {
		name := strings.TrimSuffix(names[0], ".")
		if forward, err := net.LookupIP(name); err == nil {
			for _, candidate := range forward {
				if (!distributedHostPolicy.ipv4Only || candidate.To4() != nil) && local.address.Equal(candidate) {
					canonical = name
					break
				}
			}
		}
	}
	local.canonical.Store(&canonical)
	return canonical, nil
}

func distributedPreferredAddress(addresses []net.IP) net.IP {
	var first net.IP
	for _, address := range addresses {
		ipv4 := address.To4() != nil
		if distributedHostPolicy.ipv4Only && !ipv4 {
			continue
		}
		if first == nil {
			first = address
		}
		if distributedHostPolicy.systemOrder || ipv4 != distributedHostPolicy.ipv6First {
			return address
		}
	}
	return first
}

func distributedNumericAddress(address net.IP) string {
	if ipv4 := address.To4(); ipv4 != nil {
		return ipv4.String()
	}
	// Inet6Address prints all eight groups without :: abbreviation.
	address = address.To16()
	return fmt.Sprintf("%x:%x:%x:%x:%x:%x:%x:%x", uint16(address[0])<<8|uint16(address[1]), uint16(address[2])<<8|uint16(address[3]), uint16(address[4])<<8|uint16(address[5]), uint16(address[6])<<8|uint16(address[7]), uint16(address[8])<<8|uint16(address[9]), uint16(address[10])<<8|uint16(address[11]), uint16(address[12])<<8|uint16(address[13]), uint16(address[14])<<8|uint16(address[15]))
}
