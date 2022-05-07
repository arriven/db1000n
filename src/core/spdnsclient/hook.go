// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package spdnsclient

import (
	"context"
	"net"
)

var (
	testHookLookupIP = func(
		ctx context.Context,
		fn func(context.Context, string, string) ([]net.IPAddr, error),
		network string,
		host string,
	) ([]net.IPAddr, error) {
		return fn(ctx, network, host)
	}
)
