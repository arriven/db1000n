package spdnsclient

import (
	"context"
	"net"
	"sync"

	"github.com/Arriven/db1000n/src/core/spdnsclient/singleflight"
)

// dnsWaitGroup can be used by tests to wait for all DNS goroutines to
// complete. This avoids races on the test hooks.
var dnsWaitGroup sync.WaitGroup

type SPResolver struct {
	// StrictErrors controls the behavior of temporary errors
	// (including timeout, socket errors, and SERVFAIL) when using
	// Go's built-in resolver. For a query composed of multiple
	// sub-queries (such as an A+AAAA address lookup, or walking the
	// DNS search list), this option causes such errors to abort the
	// whole query instead of returning a partial result. This is
	// not enabled by default because it may affect compatibility
	// with resolvers that process AAAA queries incorrectly.
	StrictErrors bool

	// Dial optionally specifies an alternate dialer for use by
	// Go's built-in DNS resolver to make TCP and UDP connections
	// to DNS services. The host in the address parameter will
	// always be a literal IP address and not a host name, and the
	// port in the address parameter will be a literal port number
	// and not a service name.
	// If the Conn returned is also a PacketConn, sent and received DNS
	// messages must adhere to RFC 1035 section 4.2.1, "UDP usage".
	// Otherwise, DNS messages transmitted over Conn must adhere
	// to RFC 7766 section 5, "Transport Protocol Selection".
	// If nil, the default dialer is used.
	Dial func(ctx context.Context, network, address string) (net.Conn, error)

	// lookupGroup merges LookupIPAddr calls together for lookups for the same
	// host. The lookupGroup key is the LookupIPAddr.host argument.
	// The return values are ([]IPAddr, error).
	lookupGroup singleflight.Group

	CustomDNSConfig *SPDNSConfig
}

func (r *SPResolver) getLookupGroup() *singleflight.Group {
	return &r.lookupGroup
}

// ipVersion returns the provided network's IP version: '4', '6' or 0
// if network does not end in a '4' or '6' byte.
func ipVersion(network string) byte {
	if network == "" {
		return 0
	}
	n := network[len(network)-1]
	if n != '4' && n != '6' {
		n = 0
	}
	return n
}

// LookupHost looks up the given host using the local resolver.
// It returns a slice of that host's addresses.
func (r *SPResolver) LookupHost(ctx context.Context, host string) (addrs []string, err error) {
	// Make sure that no matter what we do later, host=="" is rejected.
	// parseIP, for example, does accept empty strings.
	if host == "" {
		return nil, &net.DNSError{Err: errNoSuchHost.Error(), Name: host, IsNotFound: true}
	}
	if ip, _ := parseIPZone(host); ip != nil {
		return []string{host}, nil
	}
	return r.lookupHost(ctx, host)
}

func (r *SPResolver) lookupHost(ctx context.Context, host string) (addrs []string, err error) {
	return r.GoLookupHost(ctx, host)
}

func (r *SPResolver) lookupIP(ctx context.Context, network, host string) (addrs []net.IPAddr, err error) {
	if true {
		return r.goLookupIP(ctx, network, host)
	}

	ips, _, err := r.goLookupIPCNAME(ctx, network, host)
	return ips, err
}

func (r *SPResolver) dial(ctx context.Context, network, server string) (net.Conn, error) {
	// Calling Dial here is scary -- we have to be sure not to
	// dial a name that will require a DNS lookup, or Dial will
	// call back here to translate it. The DNS config parser has
	// already checked that all the cfg.servers are IP
	// addresses, which Dial will use without a DNS lookup.
	var c net.Conn
	var err error
	if r != nil && r.Dial != nil {
		c, err = r.Dial(ctx, network, server)
	} else {
		var d net.Dialer
		c, err = d.DialContext(ctx, network, server)
	}
	if err != nil {
		return nil, mapErr(err)
	}
	return c, nil
}

// LookupIPAddr looks up host using the local resolver.
// It returns a slice of that host's IPv4 and IPv6 addresses.
func (r *SPResolver) LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	return r.lookupIPAddr(ctx, "ip", host)
}

// lookupIPAddr looks up host using the local resolver and particular network.
// It returns a slice of that host's IPv4 and IPv6 addresses.
func (r *SPResolver) lookupIPAddr(ctx context.Context, network, host string) ([]net.IPAddr, error) {
	// Make sure that no matter what we do later, host=="" is rejected.
	// parseIP, for example, does accept empty strings.
	if host == "" {
		return nil, &net.DNSError{Err: errNoSuchHost.Error(), Name: host, IsNotFound: true}
	}
	if ip, zone := parseIPZone(host); ip != nil {
		return []net.IPAddr{{IP: ip, Zone: zone}}, nil
	}

	// The underlying resolver func is lookupIP by default but it
	// can be overridden by tests. This is needed by net/http, so it
	// uses a context key instead of unexported variables.
	resolverFunc := r.lookupIP
	// if alt, _ := ctx.Value(nettrace.LookupIPAltResolverKey{}).(func(context.Context, string, string) ([]net.IPAddr, error)); alt != nil {
	// 	resolverFunc = alt
	// }

	// We don't want a cancellation of ctx to affect the
	// lookupGroup operation. Otherwise if our context gets
	// canceled it might cause an error to be returned to a lookup
	// using a completely different context. However we need to preserve
	// only the values in context. See Issue 28600.
	lookupGroupCtx, lookupGroupCancel := context.WithCancel(withUnexpiredValuesPreserved(ctx))

	lookupKey := network + "\000" + host
	dnsWaitGroup.Add(1)
	ch, called := r.getLookupGroup().DoChan(lookupKey, func() (interface{}, error) {
		defer dnsWaitGroup.Done()
		return testHookLookupIP(lookupGroupCtx, resolverFunc, network, host)
	})
	if !called {
		dnsWaitGroup.Done()
	}

	select {
	case <-ctx.Done():
		// Our context was canceled. If we are the only
		// goroutine looking up this key, then drop the key
		// from the lookupGroup and cancel the lookup.
		// If there are other goroutines looking up this key,
		// let the lookup continue uncanceled, and let later
		// lookups with the same key share the result.
		// See issues 8602, 20703, 22724.
		if r.getLookupGroup().ForgetUnshared(lookupKey) {
			lookupGroupCancel()
		} else {
			go func() {
				<-ch
				lookupGroupCancel()
			}()
		}
		err := mapErr(ctx.Err())

		return nil, err
	case r := <-ch:
		lookupGroupCancel()

		return lookupIPReturn(r.Val, r.Err, r.Shared)
	}
}

// onlyValuesCtx is a context that uses an underlying context
// for value lookup if the underlying context hasn't yet expired.
type onlyValuesCtx struct {
	context.Context
	lookupValues context.Context
}

var _ context.Context = (*onlyValuesCtx)(nil)

// Value performs a lookup if the original context hasn't expired.
func (ovc *onlyValuesCtx) Value(key interface{}) interface{} {
	select {
	case <-ovc.lookupValues.Done():
		return nil
	default:
		return ovc.lookupValues.Value(key)
	}
}

// withUnexpiredValuesPreserved returns a context.Context that only uses lookupCtx
// for its values, otherwise it is never canceled and has no deadline.
// If the lookup context expires, any looked up values will return nil.
// See Issue 28600.
func withUnexpiredValuesPreserved(lookupCtx context.Context) context.Context {
	return &onlyValuesCtx{Context: context.Background(), lookupValues: lookupCtx}
}

// lookupIPReturn turns the return values from singleflight.Do into
// the return values from LookupIP.
func lookupIPReturn(addrsi interface{}, err error, shared bool) ([]net.IPAddr, error) {
	if err != nil {
		return nil, err
	}
	addrs := addrsi.([]net.IPAddr)
	if shared {
		clone := make([]net.IPAddr, len(addrs))
		copy(clone, addrs)
		addrs = clone
	}
	return addrs, nil
}
