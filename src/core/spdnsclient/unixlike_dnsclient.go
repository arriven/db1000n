package spdnsclient

import (
	"context"
	"errors"
	"io"
	"net"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

var (
	errLameReferral              = errors.New("[SP] lame referral")
	errCannotUnmarshalDNSMessage = errors.New("[SP] cannot unmarshal DNS message")
	errCannotMarshalDNSMessage   = errors.New("[SP] cannot marshal DNS message")
	errServerMisbehaving         = errors.New("[SP] server misbehaving")
	errInvalidDNSResponse        = errors.New("[SP] invalid DNS response")
	errNoAnswerFromDNSServer     = errors.New("[SP] no answer from DNS server")

	// errServerTemporarilyMisbehaving is like errServerMisbehaving, except
	// that when it gets translated to a DNSError, the IsTemporary field
	// gets set to true.
	errServerTemporarilyMisbehaving = errors.New("[SP] server misbehaving")
)

func newRequest(q dnsmessage.Question) (id uint16, udpReq, tcpReq []byte, err error) {
	id = uint16(randInt())
	b := dnsmessage.NewBuilder(make([]byte, 2, 514), dnsmessage.Header{ID: id, RecursionDesired: true})
	b.EnableCompression()
	if err := b.StartQuestions(); err != nil {
		return 0, nil, nil, err
	}
	if err := b.Question(q); err != nil {
		return 0, nil, nil, err
	}
	tcpReq, err = b.Finish()
	udpReq = tcpReq[2:]
	l := len(tcpReq) - 2
	tcpReq[0] = byte(l >> 8)
	tcpReq[1] = byte(l)
	return id, udpReq, tcpReq, err
}

func (r *SPResolver) strictErrors() bool { return r != nil && r.StrictErrors }

// goLookupHost is the native Go implementation of LookupHost.
// Used only if cgoLookupHost refuses to handle the request
// (that is, only if cgoLookupHost is the stub in cgo_stub.go).
// Normally we let cgo use the C library resolver instead of
// depending on our lookup code, so that Go and C get the same
// answers.
func (r *SPResolver) GoLookupHost(ctx context.Context, name string) (addrs []string, err error) {

	ips, _, err := r.goLookupIPCNAME(ctx, "ip", name)
	if err != nil {
		return
	}
	addrs = make([]string, 0, len(ips))
	for _, ip := range ips {
		addrs = append(addrs, ip.String())
	}
	return
}

// goLookupIP is the native Go implementation of LookupIP.
// The libc versions are in cgo_*.go.
func (r *SPResolver) goLookupIP(ctx context.Context, network, host string) (addrs []net.IPAddr, err error) {
	addrs, _, err = r.goLookupIPCNAME(ctx, network, host)
	return
}

func (r *SPResolver) goLookupIPCNAME(ctx context.Context, network, name string) (addrs []net.IPAddr, cname dnsmessage.Name, err error) {

	if !isDomainName(name) {
		// See comment in func lookup above about use of errNoSuchHost.
		return nil, dnsmessage.Name{}, &net.DNSError{Err: errNoSuchHost.Error(), Name: name, IsNotFound: true}
	}

	conf := r.CustomDNSConfig

	type result struct {
		p      dnsmessage.Parser
		server string
		error
	}
	lane := make(chan result, 1)
	qtypes := []dnsmessage.Type{dnsmessage.TypeA, dnsmessage.TypeAAAA}
	switch ipVersion(network) {
	case '4':
		qtypes = []dnsmessage.Type{dnsmessage.TypeA}
	case '6':
		qtypes = []dnsmessage.Type{dnsmessage.TypeAAAA}
	}
	var queryFn func(fqdn string, qtype dnsmessage.Type)
	var responseFn func(fqdn string, qtype dnsmessage.Type) result
	if conf.SingleRequest {
		queryFn = func(fqdn string, qtype dnsmessage.Type) {}
		responseFn = func(fqdn string, qtype dnsmessage.Type) result {
			dnsWaitGroup.Add(1)
			defer dnsWaitGroup.Done()
			p, server, err := r.tryOneName(ctx, conf, fqdn, qtype)
			return result{p, server, err}
		}
	} else {
		queryFn = func(fqdn string, qtype dnsmessage.Type) {
			dnsWaitGroup.Add(1)
			go func(qtype dnsmessage.Type) {
				p, server, err := r.tryOneName(ctx, conf, fqdn, qtype)
				lane <- result{p, server, err}
				dnsWaitGroup.Done()
			}(qtype)
		}

		responseFn = func(fqdn string, qtype dnsmessage.Type) result {
			return <-lane
		}
	}
	var lastErr error
	for _, fqdn := range conf.nameList(name) {
		for _, qtype := range qtypes {
			queryFn(fqdn, qtype)
		}
		hitStrictError := false
		for _, qtype := range qtypes {
			result := responseFn(fqdn, qtype)
			if result.error != nil {
				if nerr, ok := result.error.(net.Error); ok && nerr.Temporary() && r.strictErrors() {
					// This error will abort the nameList loop.
					hitStrictError = true
					lastErr = result.error
				} else if lastErr == nil || fqdn == name+"." {
					// Prefer error for original name.
					lastErr = result.error
				}
				continue
			}

			// Presotto says it's okay to assume that servers listed in
			// /etc/resolv.conf are recursive resolvers.
			//
			// We asked for recursion, so it should have included all the
			// answers we need in this one packet.
			//
			// Further, RFC 1035 section 4.3.1 says that "the recursive
			// response to a query will be... The answer to the query,
			// possibly preface by one or more CNAME RRs that specify
			// aliases encountered on the way to an answer."
			//
			// Therefore, we should be able to assume that we can ignore
			// CNAMEs and that the A and AAAA records we requested are
			// for the canonical name.

		loop:
			for {
				h, err := result.p.AnswerHeader()
				if err != nil && err != dnsmessage.ErrSectionDone {
					lastErr = &net.DNSError{
						Err:    "[SP] cannot marshal DNS message: " + err.Error(),
						Name:   name,
						Server: result.server,
					}
				}
				if err != nil {
					break
				}
				switch h.Type {
				case dnsmessage.TypeA:
					a, err := result.p.AResource()
					if err != nil {
						lastErr = &net.DNSError{
							Err:    "[SP] cannot marshal DNS message TypeA",
							Name:   name,
							Server: result.server,
						}
						break loop
					}
					addrs = append(addrs, net.IPAddr{IP: net.IP(a.A[:])})

				case dnsmessage.TypeAAAA:
					aaaa, err := result.p.AAAAResource()
					if err != nil {
						lastErr = &net.DNSError{
							Err:    "[SP] cannot marshal DNS message TypeAAAA",
							Name:   name,
							Server: result.server,
						}
						break loop
					}
					addrs = append(addrs, net.IPAddr{IP: net.IP(aaaa.AAAA[:])})

				default:
					if err := result.p.SkipAnswer(); err != nil {
						lastErr = &net.DNSError{
							Err:    "[SP] cannot marshal DNS message [default]",
							Name:   name,
							Server: result.server,
						}
						break loop
					}
					continue
				}
				if cname.Length == 0 && h.Name.Length != 0 {
					cname = h.Name
				}
			}
		}
		if hitStrictError {
			// If either family hit an error with StrictErrors enabled,
			// discard all addresses. This ensures that network flakiness
			// cannot turn a dualstack hostname IPv4/IPv6-only.
			addrs = nil
			break
		}
		if len(addrs) > 0 {
			break
		}
	}
	if lastErr, ok := lastErr.(*net.DNSError); ok {
		// Show original name passed to lookup, not suffixed one.
		// In general we might have tried many suffixes; showing
		// just one is misleading. See also golang.org/issue/6324.
		lastErr.Name = name
	}
	sortByRFC6724(addrs)
	if len(addrs) == 0 {
		if len(addrs) == 0 && lastErr != nil {
			return nil, dnsmessage.Name{}, lastErr
		}
	}
	return addrs, cname, nil
}

func (r *SPResolver) tryOneName(ctx context.Context, cfg *SPDNSConfig, name string, qtype dnsmessage.Type) (dnsmessage.Parser, string, error) {
	var lastErr error
	serverOffset := cfg.serverOffset()
	sLen := uint32(len(cfg.Servers))

	n, err := dnsmessage.NewName(name)
	if err != nil {
		return dnsmessage.Parser{}, "", errCannotMarshalDNSMessage
	}
	q := dnsmessage.Question{
		Name:  n,
		Type:  qtype,
		Class: dnsmessage.ClassINET,
	}

	for i := 0; i < cfg.Attempts; i++ {
		for j := uint32(0); j < sLen; j++ {
			server := cfg.Servers[(serverOffset+j)%sLen]

			p, h, err := r.exchange(ctx, server, q, cfg.Timeout, cfg.UseTCP)
			if err != nil {
				dnsErr := &net.DNSError{
					Err:    err.Error(),
					Name:   name,
					Server: server,
				}
				if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
					dnsErr.IsTimeout = true
				}
				// Set IsTemporary for socket-level errors. Note that this flag
				// may also be used to indicate a SERVFAIL response.
				if _, ok := err.(*net.OpError); ok {
					dnsErr.IsTemporary = true
				}
				lastErr = dnsErr
				continue
			}

			if err := checkHeader(&p, h); err != nil {
				dnsErr := &net.DNSError{
					Err:    err.Error(),
					Name:   name,
					Server: server,
				}
				if err == errServerTemporarilyMisbehaving {
					dnsErr.IsTemporary = true
				}
				if err == errNoSuchHost {
					// The name does not exist, so trying
					// another server won't help.

					dnsErr.IsNotFound = true
					return p, server, dnsErr
				}
				lastErr = dnsErr
				continue
			}

			err = skipToAnswer(&p, qtype)
			if err == nil {
				return p, server, nil
			}
			lastErr = &net.DNSError{
				Err:    err.Error(),
				Name:   name,
				Server: server,
			}
			if err == errNoSuchHost {
				// The name does not exist, so trying another
				// server won't help.

				lastErr.(*net.DNSError).IsNotFound = true
				return p, server, lastErr
			}
		}
	}
	return dnsmessage.Parser{}, "", lastErr
}

// exchange sends a query on the connection and hopes for a response.
func (r *SPResolver) exchange(ctx context.Context, server string, q dnsmessage.Question, timeout time.Duration, useTCP bool) (dnsmessage.Parser, dnsmessage.Header, error) {
	q.Class = dnsmessage.ClassINET
	id, udpReq, tcpReq, err := newRequest(q)
	if err != nil {
		return dnsmessage.Parser{}, dnsmessage.Header{}, errCannotMarshalDNSMessage
	}
	var networks []string
	if useTCP {
		networks = []string{"tcp"}
	} else {
		networks = []string{"udp", "tcp"}
	}
	for _, network := range networks {
		ctx, cancel := context.WithDeadline(ctx, time.Now().Add(timeout))
		defer cancel()

		c, err := r.dial(ctx, network, server)
		if err != nil {
			return dnsmessage.Parser{}, dnsmessage.Header{}, err
		}
		if d, ok := ctx.Deadline(); ok && !d.IsZero() {
			c.SetDeadline(d)
		}
		var p dnsmessage.Parser
		var h dnsmessage.Header
		if _, ok := c.(net.PacketConn); ok {
			p, h, err = dnsPacketRoundTrip(c, id, q, udpReq)
		} else {
			p, h, err = dnsStreamRoundTrip(c, id, q, tcpReq)
		}
		c.Close()
		if err != nil {
			return dnsmessage.Parser{}, dnsmessage.Header{}, mapErr(err)
		}
		if err := p.SkipQuestion(); err != dnsmessage.ErrSectionDone {
			return dnsmessage.Parser{}, dnsmessage.Header{}, errInvalidDNSResponse
		}
		if h.Truncated { // see RFC 5966
			continue
		}
		return p, h, nil
	}
	return dnsmessage.Parser{}, dnsmessage.Header{}, errNoAnswerFromDNSServer
}

func dnsPacketRoundTrip(c net.Conn, id uint16, query dnsmessage.Question, b []byte) (dnsmessage.Parser, dnsmessage.Header, error) {
	if _, err := c.Write(b); err != nil {
		return dnsmessage.Parser{}, dnsmessage.Header{}, err
	}

	b = make([]byte, 512) // see RFC 1035
	for {
		n, err := c.Read(b)
		if err != nil {
			return dnsmessage.Parser{}, dnsmessage.Header{}, err
		}
		var p dnsmessage.Parser
		// Ignore invalid responses as they may be malicious
		// forgery attempts. Instead continue waiting until
		// timeout. See golang.org/issue/13281.
		h, err := p.Start(b[:n])
		if err != nil {
			continue
		}
		q, err := p.Question()
		if err != nil || !checkResponse(id, query, h, q) {
			continue
		}
		return p, h, nil
	}
}

func dnsStreamRoundTrip(c net.Conn, id uint16, query dnsmessage.Question, b []byte) (dnsmessage.Parser, dnsmessage.Header, error) {
	if _, err := c.Write(b); err != nil {
		return dnsmessage.Parser{}, dnsmessage.Header{}, err
	}

	b = make([]byte, 1280) // 1280 is a reasonable initial size for IP over Ethernet, see RFC 4035
	if _, err := io.ReadFull(c, b[:2]); err != nil {
		return dnsmessage.Parser{}, dnsmessage.Header{}, err
	}
	l := int(b[0])<<8 | int(b[1])
	if l > len(b) {
		b = make([]byte, l)
	}
	n, err := io.ReadFull(c, b[:l])
	if err != nil {
		return dnsmessage.Parser{}, dnsmessage.Header{}, err
	}
	var p dnsmessage.Parser
	h, err := p.Start(b[:n])
	if err != nil {
		return dnsmessage.Parser{}, dnsmessage.Header{}, errCannotUnmarshalDNSMessage
	}
	q, err := p.Question()
	if err != nil {
		return dnsmessage.Parser{}, dnsmessage.Header{}, errCannotUnmarshalDNSMessage
	}
	if !checkResponse(id, query, h, q) {
		return dnsmessage.Parser{}, dnsmessage.Header{}, errInvalidDNSResponse
	}
	return p, h, nil
}

func checkResponse(reqID uint16, reqQues dnsmessage.Question, respHdr dnsmessage.Header, respQues dnsmessage.Question) bool {
	if !respHdr.Response {
		return false
	}
	if reqID != respHdr.ID {
		return false
	}
	if reqQues.Type != respQues.Type || reqQues.Class != respQues.Class || !equalASCIIName(reqQues.Name, respQues.Name) {
		return false
	}
	return true
}

func equalASCIIName(x, y dnsmessage.Name) bool {
	if x.Length != y.Length {
		return false
	}
	for i := 0; i < int(x.Length); i++ {
		a := x.Data[i]
		b := y.Data[i]
		if 'A' <= a && a <= 'Z' {
			a += 0x20
		}
		if 'A' <= b && b <= 'Z' {
			b += 0x20
		}
		if a != b {
			return false
		}
	}
	return true
}

// isDomainName checks if a string is a presentation-format domain name
// (currently restricted to hostname-compatible "preferred name" LDH labels and
// SRV-like "underscore labels"; see golang.org/issue/12421).
func isDomainName(s string) bool {
	// See RFC 1035, RFC 3696.
	// Presentation format has dots before every label except the first, and the
	// terminal empty label is optional here because we assume fully-qualified
	// (absolute) input. We must therefore reserve space for the first and last
	// labels' length octets in wire format, where they are necessary and the
	// maximum total length is 255.
	// So our _effective_ maximum is 253, but 254 is not rejected if the last
	// character is a dot.
	l := len(s)
	if l == 0 || l > 254 || l == 254 && s[l-1] != '.' {
		return false
	}

	last := byte('.')
	nonNumeric := false // true once we've seen a letter or hyphen
	partlen := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		default:
			return false
		case 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || c == '_':
			nonNumeric = true
			partlen++
		case '0' <= c && c <= '9':
			// fine
			partlen++
		case c == '-':
			// Byte before dash cannot be dot.
			if last == '.' {
				return false
			}
			partlen++
			nonNumeric = true
		case c == '.':
			// Byte before dot cannot be dot, dash.
			if last == '.' || last == '-' {
				return false
			}
			if partlen > 63 || partlen == 0 {
				return false
			}
			partlen = 0
		}
		last = c
	}
	if last == '-' || partlen > 63 {
		return false
	}

	return nonNumeric
}

// avoidDNS reports whether this is a hostname for which we should not
// use DNS. Currently this includes only .onion, per RFC 7686. See
// golang.org/issue/13705. Does not cover .local names (RFC 6762),
// see golang.org/issue/16739.
func avoidDNS(name string) bool {
	if name == "" {
		return true
	}
	if name[len(name)-1] == '.' {
		name = name[:len(name)-1]
	}
	return stringsHasSuffixFold(name, ".onion")
}

// checkHeader performs basic sanity checks on the header.
func checkHeader(p *dnsmessage.Parser, h dnsmessage.Header) error {
	if h.RCode == dnsmessage.RCodeNameError {
		return errNoSuchHost
	}

	_, err := p.AnswerHeader()
	if err != nil && err != dnsmessage.ErrSectionDone {
		return errCannotUnmarshalDNSMessage
	}

	// libresolv continues to the next server when it receives
	// an invalid referral response. See golang.org/issue/15434.
	if h.RCode == dnsmessage.RCodeSuccess && !h.Authoritative && !h.RecursionAvailable && err == dnsmessage.ErrSectionDone {
		return errLameReferral
	}

	if h.RCode != dnsmessage.RCodeSuccess && h.RCode != dnsmessage.RCodeNameError {
		// None of the error codes make sense
		// for the query we sent. If we didn't get
		// a name error and we didn't get success,
		// the server is behaving incorrectly or
		// having temporary trouble.
		if h.RCode == dnsmessage.RCodeServerFailure {
			return errServerTemporarilyMisbehaving
		}
		return errServerMisbehaving
	}

	return nil
}

func skipToAnswer(p *dnsmessage.Parser, qtype dnsmessage.Type) error {
	for {
		h, err := p.AnswerHeader()
		if err == dnsmessage.ErrSectionDone {
			return errNoSuchHost
		}
		if err != nil {
			return errCannotUnmarshalDNSMessage
		}
		if h.Type == qtype {
			return nil
		}
		if err := p.SkipAnswer(); err != nil {
			return errCannotUnmarshalDNSMessage
		}
	}
}
