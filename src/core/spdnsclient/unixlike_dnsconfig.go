// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package spdnsclient

import (
	"sync/atomic"
	"time"
)

type SPDNSConfig struct {
	Servers       []string      // server addresses (in host:port form) to use
	Search        []string      // rooted suffixes to append to local name
	Ndots         int           // number of dots in name to trigger absolute lookup
	Timeout       time.Duration // wait before giving up on a query, including retries
	Attempts      int           // lost packets before giving up on server
	Rotate        bool          // round robin among servers
	UnknownOpt    bool          // anything unknown was encountered
	Lookup        []string      // OpenBSD top-level database "lookup" order
	Err           error         // any error that occurs during open of resolv.conf
	Mtime         time.Time     // time of resolv.conf modification
	Soffset       uint32        // used by serverOffset
	SingleRequest bool          // use sequential A and AAAA queries instead of parallel queries
	UseTCP        bool          // force usage of TCP for DNS resolutions
}

// serverOffset returns an offset that can be used to determine
// indices of servers in c.servers when making queries.
// When the rotate option is enabled, this offset increases.
// Otherwise it is always 0.
func (c *SPDNSConfig) serverOffset() uint32 {
	if c.Rotate {
		return atomic.AddUint32(&c.Soffset, 1) - 1 // return 0 to start
	}
	return 0
}

// nameList returns a list of names for sequential DNS queries.
func (conf *SPDNSConfig) nameList(name string) []string {
	if avoidDNS(name) {
		return nil
	}

	// Check name length (see isDomainName).
	l := len(name)
	rooted := l > 0 && name[l-1] == '.'
	if l > 254 || l == 254 && rooted {
		return nil
	}

	// If name is rooted (trailing dot), try only that name.
	if rooted {
		return []string{name}
	}

	hasNdots := count(name, '.') >= conf.Ndots
	name += "."
	l++

	// Build list of search choices.
	names := make([]string, 0, 1+len(conf.Search))
	// If name has enough dots, try unsuffixed first.
	if hasNdots {
		names = append(names, name)
	}
	// Try suffixes that are not too long (see isDomainName).
	for _, suffix := range conf.Search {
		if l+len(suffix) <= 254 {
			names = append(names, name+suffix)
		}
	}
	// Try unsuffixed, if not tried first above.
	if !hasNdots {
		names = append(names, name)
	}
	return names
}

func ensureRooted(s string) string {
	if len(s) > 0 && s[len(s)-1] == '.' {
		return s
	}
	return s + "."
}
