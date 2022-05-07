package customtcpdial

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/valyala/fasthttp"
)

// TCPDialer contains options to control a group of Dial calls.
type CustomTCPDialer struct {
	// Concurrency controls the maximum number of concurrent Dials
	// that can be performed using this object.
	// Setting this to 0 means unlimited.
	//
	// WARNING: This can only be changed before the first Dial.
	// Changes made after the first Dial will not affect anything.
	Concurrency int

	// Dial tickets controls how fast we can spam TCP SYN
	// For every net.Dial made we take a ticket
	// from the channel. If there was no ticket,
	// Dial just returns an error.
	DialTicketsC chan bool

	// LocalAddr is the local address to use when dialing an
	// address.
	// If nil, a local address is automatically chosen.
	LocalAddr *net.TCPAddr

	// This may be used to override DNS resolving policy, like this:
	// var dialer = &fasthttp.TCPDialer{
	// 	Resolver: &net.Resolver{
	// 		PreferGo:     true,
	// 		StrictErrors: false,
	// 		Dial: func (ctx context.Context, network, address string) (net.Conn, error) {
	// 			d := net.Dialer{}
	// 			return d.DialContext(ctx, "udp", "8.8.8.8:53")
	// 		},
	// 	},
	// }
	Resolver fasthttp.Resolver

	// DNSCacheDuration may be used to override the default DNS cache duration (DefaultDNSCacheDuration)
	DNSCacheDuration time.Duration

	tcpAddrsMap sync.Map

	concurrencyCh chan struct{}

	once sync.Once

	ParentDialer fasthttp.DialFunc
}

// Dial dials the given TCP addr using tcp4.
//
// This function has the following additional features comparing to net.Dial:
//
//   * It reduces load on DNS resolver by caching resolved TCP addressed
//     for DNSCacheDuration.
//   * It dials all the resolved TCP addresses in round-robin manner until
//     connection is established. This may be useful if certain addresses
//     are temporarily unreachable.
//   * It returns ErrDialTimeout if connection cannot be established during
//     DefaultDialTimeout seconds. Use DialTimeout for customizing dial timeout.
//
// This dialer is intended for custom code wrapping before passing
// to Client.Dial or HostClient.Dial.
//
// For instance, per-host counters and/or limits may be implemented
// by such wrappers.
//
// The addr passed to the function must contain port. Example addr values:
//
//     * foobar.baz:443
//     * foo.bar:80
//     * aaa.com:8080
func (d *CustomTCPDialer) Dial(addr string) (net.Conn, error) {
	return d.dial(addr, false, DefaultDialTimeout)
}

// DialTimeout dials the given TCP addr using tcp4 using the given timeout.
//
// This function has the following additional features comparing to net.Dial:
//
//   * It reduces load on DNS resolver by caching resolved TCP addressed
//     for DNSCacheDuration.
//   * It dials all the resolved TCP addresses in round-robin manner until
//     connection is established. This may be useful if certain addresses
//     are temporarily unreachable.
//
// This dialer is intended for custom code wrapping before passing
// to Client.Dial or HostClient.Dial.
//
// For instance, per-host counters and/or limits may be implemented
// by such wrappers.
//
// The addr passed to the function must contain port. Example addr values:
//
//     * foobar.baz:443
//     * foo.bar:80
//     * aaa.com:8080
func (d *CustomTCPDialer) DialTimeout(addr string, timeout time.Duration) (net.Conn, error) {
	return d.dial(addr, false, timeout)
}

// DialDualStack dials the given TCP addr using both tcp4 and tcp6.
//
// This function has the following additional features comparing to net.Dial:
//
//   * It reduces load on DNS resolver by caching resolved TCP addressed
//     for DNSCacheDuration.
//   * It dials all the resolved TCP addresses in round-robin manner until
//     connection is established. This may be useful if certain addresses
//     are temporarily unreachable.
//   * It returns ErrDialTimeout if connection cannot be established during
//     DefaultDialTimeout seconds. Use DialDualStackTimeout for custom dial
//     timeout.
//
// This dialer is intended for custom code wrapping before passing
// to Client.Dial or HostClient.Dial.
//
// For instance, per-host counters and/or limits may be implemented
// by such wrappers.
//
// The addr passed to the function must contain port. Example addr values:
//
//     * foobar.baz:443
//     * foo.bar:80
//     * aaa.com:8080
func (d *CustomTCPDialer) DialDualStack(addr string) (net.Conn, error) {
	return d.dial(addr, true, DefaultDialTimeout)
}

// DialDualStackTimeout dials the given TCP addr using both tcp4 and tcp6
// using the given timeout.
//
// This function has the following additional features comparing to net.Dial:
//
//   * It reduces load on DNS resolver by caching resolved TCP addressed
//     for DNSCacheDuration.
//   * It dials all the resolved TCP addresses in round-robin manner until
//     connection is established. This may be useful if certain addresses
//     are temporarily unreachable.
//
// This dialer is intended for custom code wrapping before passing
// to Client.Dial or HostClient.Dial.
//
// For instance, per-host counters and/or limits may be implemented
// by such wrappers.
//
// The addr passed to the function must contain port. Example addr values:
//
//     * foobar.baz:443
//     * foo.bar:80
//     * aaa.com:8080
func (d *CustomTCPDialer) DialDualStackTimeout(addr string, timeout time.Duration) (net.Conn, error) {
	return d.dial(addr, true, timeout)
}

var ErrTooFastDialSpam = errors.New("too fast TCP SYN (dial spam)")

func (d *CustomTCPDialer) dial(addr string, dualStack bool, timeout time.Duration) (net.Conn, error) {
	d.once.Do(func() {
		if d.Concurrency > 0 {
			d.concurrencyCh = make(chan struct{}, d.Concurrency)
		}

		if d.DNSCacheDuration == 0 {
			d.DNSCacheDuration = DefaultDNSCacheDuration
		}

		go d.tcpAddrsClean()
	})

	addrs, idx, err := d.getTCPAddrs(addr, dualStack)
	if err != nil {
		return nil, err
	}
	network := "tcp4"
	if dualStack {
		network = "tcp"
	}
	checkErr := CheckNonPublicTCPEndpoints(addrs)
	if checkErr != nil {
		return nil, errors.New("CustomTCPDialer: " + checkErr.Error())
	}

	var conn net.Conn
	n := uint32(len(addrs))
	deadline := time.Now().Add(timeout)
	for n > 0 {
		currentAddr := &addrs[idx%n]
		conn, err = d.tryDial(network, currentAddr, deadline, d.concurrencyCh)
		if err == nil {
			return conn, nil
		}
		if err == ErrDialTimeout {
			return nil, fmt.Errorf("CustomTCPDialer: %s timed out", currentAddr)
		}
		idx++
		n--
	}
	return nil, err
}

func (d *CustomTCPDialer) tryDial(network string, addr *net.TCPAddr, deadline time.Time, concurrencyCh chan struct{}) (net.Conn, error) {
	timeout := -time.Since(deadline)
	if timeout <= 0 {
		return nil, ErrDialTimeout
	}

	if concurrencyCh != nil {
		select {
		case concurrencyCh <- struct{}{}:
		default:
			tc := fasthttp.AcquireTimer(timeout)
			isTimeout := false
			select {
			case concurrencyCh <- struct{}{}:
			case <-tc.C:
				isTimeout = true
			}
			fasthttp.ReleaseTimer(tc)
			if isTimeout {
				return nil, ErrDialTimeout
			}
		}
		defer func() { <-concurrencyCh }()
	}
	ticketC := d.DialTicketsC
	if ticketC != nil {
		select {
		// either we catch the ticket instantly
		case <-ticketC:
		// or maybe let's wait until we have a green light
		case <-time.After(timeout / 2):
			// time passed, we didn't get a ticket :(
			return nil, ErrTooFastDialSpam
		}
	}

	dialer := d.ParentDialer
	if dialer == nil {
		panic("CustomTCPDialer.ParentDialer cannot be nil")
	}

	conn, err := dialer(addr.String())

	return conn, err
}

// ErrDialTimeout is returned when TCP dialing is timed out.
var ErrDialTimeout = errors.New("dialing to the given TCP address timed out")

// DefaultDialTimeout is timeout used by Dial and DialDualStack
// for establishing TCP connections.
const DefaultDialTimeout = 3 * time.Second

type tcpAddrEntry struct {
	addrs    []net.TCPAddr
	addrsIdx uint32

	pending           int32
	shouldResolveNext int32
	resolveTime       time.Time
}

// DefaultDNSCacheDuration is the duration for caching resolved TCP addresses
// by Dial* functions.
const DefaultDNSCacheDuration = time.Minute

func (d *CustomTCPDialer) tcpAddrsClean() {
	expireDuration := 2 * d.DNSCacheDuration
	for {
		time.Sleep(d.DNSCacheDuration / 4)
		t := time.Now()
		d.tcpAddrsMap.Range(func(k, v interface{}) bool {
			if e, ok := v.(*tcpAddrEntry); ok {
				elapsed := t.Sub(e.resolveTime)
				if atomic.LoadInt32(&e.shouldResolveNext) == 0 {
					if elapsed > d.DNSCacheDuration {
						atomic.SwapInt32(&e.shouldResolveNext, 1)
					}
				}
				if elapsed > expireDuration {
					d.tcpAddrsMap.Delete(k)
				}
			}

			return true
		})

	}
}

func (d *CustomTCPDialer) getTCPAddrs(addr string, dualStack bool) ([]net.TCPAddr, uint32, error) {
	item, exist := d.tcpAddrsMap.Load(addr)
	e, ok := item.(*tcpAddrEntry)
	if exist && ok && e != nil && atomic.SwapInt32(&e.shouldResolveNext, 0) == 1 {
		// Only let one goroutine re-resolve at a time.
		if atomic.SwapInt32(&e.pending, 1) == 0 {
			e = nil
		}
	}

	if e == nil {
		addrs, err := resolveTCPAddrs(addr, dualStack, d.Resolver)
		if err != nil {
			item, exist := d.tcpAddrsMap.Load(addr)
			e, ok = item.(*tcpAddrEntry)
			if exist && ok && e != nil {
				// Set pending to 0 so another goroutine can retry.
				atomic.StoreInt32(&e.pending, 0)
			}
			return nil, 0, err
		}

		e = &tcpAddrEntry{
			addrs:       addrs,
			resolveTime: time.Now(),
		}
		d.tcpAddrsMap.Store(addr, e)
	}

	idx := atomic.AddUint32(&e.addrsIdx, 1)
	return e.addrs, idx, nil
}

func resolveTCPAddrs(addr string, dualStack bool, resolver fasthttp.Resolver) ([]net.TCPAddr, error) {
	host, portS, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(portS)
	if err != nil {
		return nil, err
	}

	if resolver == nil {
		resolver = net.DefaultResolver
	}

	ctx := context.Background()
	ipaddrs, err := resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}

	n := len(ipaddrs)
	addrs := make([]net.TCPAddr, 0, n)
	for i := 0; i < n; i++ {
		ip := ipaddrs[i]
		if !dualStack && ip.IP.To4() == nil {
			continue
		}
		addrs = append(addrs, net.TCPAddr{
			IP:   ip.IP,
			Port: port,
			Zone: ip.Zone,
		})
	}
	if len(addrs) == 0 {
		return nil, errNoDNSEntries
	}
	return addrs, nil
}

var errNoDNSEntries = errors.New("couldn't find DNS entries for the given domain. Try using DialDualStack")
