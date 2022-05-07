package customresolver

import (
	"context"
	"net"
	"time"

	"github.com/Arriven/db1000n/src/core/spdnsclient"
	"github.com/patrickmn/go-cache"
)

// We need reliable IP addresses
// just like we would've been in Russia/Belarus
var ReferenceDNSServersForHTTP = []string{
	// https://dns.yandex.com/
	"77.88.8.8:53",
	"77.88.8.1:53",
	"77.88.8.2:53",
	"77.88.8.3:53",
	"77.88.8.7:53",
	"77.88.8.8:53",
	"77.88.8.88:53",
}

var DnsCache *cache.Cache

var MasterResolver = &CustomResolver{
	FirstResolver: &spdnsclient.SPResolver{

		CustomDNSConfig: MakeDNSConfig(),
	},
	ParentResolver: net.DefaultResolver,
}

func MakeDNSConfig() (conf *spdnsclient.SPDNSConfig) {
	conf = &spdnsclient.SPDNSConfig{
		Ndots:    1,
		Timeout:  5 * time.Second,
		Attempts: 2,
	}
	conf.Servers = ReferenceDNSServersForHTTP

	return
}

type CustomResolver struct {
	FirstResolver  *spdnsclient.SPResolver
	ParentResolver *net.Resolver
}
type Resolver interface {
	LookupIPAddr(context.Context, string) (names []net.IPAddr, err error)
}

func (cr *CustomResolver) LookupIPAddr(ctx context.Context, host string) (names []net.IPAddr, err error) {
	if c, found := DnsCache.Get(host); found {
		return c.([]net.IPAddr), nil
	}
	names, err = cr.LookupIPAddrNoCache(ctx, host)
	if err == nil {
		DnsCache.SetDefault(host, names)
	}
	return
}
func (cr *CustomResolver) LookupIPAddrNoCache(ctx context.Context, host string) (names []net.IPAddr, err error) {
	names, err = cr.FirstResolver.LookupIPAddr(ctx, host)
	if err == nil {
		return
	}
	names, err = cr.ParentResolver.LookupIPAddr(ctx, host)
	return
}

func init() {
	DnsCache = cache.New(5*time.Minute, 10*time.Minute)
}
