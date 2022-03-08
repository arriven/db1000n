// MIT License

// Copyright (c) [2022] [Arriven (https://github.com/Arriven)]

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package metrics

import (
	"context"
	"github.com/Arriven/db1000n/src/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/push"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// common values for prometheus metrics
const (
	StatusLabel   = `status`
	StatusSuccess = `success`
	StatusFail    = `fail`
)

// DNS Blast related values and labels for prometheus metrics
const (
	DNSBlastRootDomainLabel = `root_domain`
	DNSBlastSeedDomainLabel = `seed_domain`
	DNSBlastProtocolLabel   = `protocol`
)

// HTTP related values and labels
const (
	HTTPDestinationHostLabel = `destination_host`
	HTTPMethodLabel          = `method`
)

// Packetgen related values and labels
const (
	PacketgenHostLabel        = `host`
	PacketgenDstHostPortLabel = `dst_host_port`
	PacketgenProtocolLabel    = `protocol`
)

// Slowloris related values and labels
const (
	SlowlorisAddressLabel  = `address`
	SlowlorisProtocolLabel = `protocol`
)

// Rawnet related values and labels
const (
	RawnetAddressLabel  = `address`
	RawnetProtocolLabel = `protocol`
)

// registered metrics
var (
	dnsBlastCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_blast_total",
			Help: "Number of dns queries",
		}, []string{DNSBlastRootDomainLabel, DNSBlastSeedDomainLabel, DNSBlastProtocolLabel, StatusLabel})
	httpCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_total",
			Help: "Number of http queries",
		}, []string{HTTPDestinationHostLabel, HTTPMethodLabel, StatusLabel})
	packetgenCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "packetgen_total",
			Help: "Number of packet generation transfers",
		}, []string{PacketgenHostLabel, PacketgenDstHostPortLabel, PacketgenProtocolLabel, StatusLabel})
	slowlorisCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "slowloris_total",
			Help: "Number of sent raw tcp/udp packets",
		}, []string{SlowlorisAddressLabel, SlowlorisProtocolLabel, StatusLabel})
	rawnetCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rawnet_total",
			Help: "Number of sent raw tcp/udp packets",
		}, []string{RawnetAddressLabel, RawnetProtocolLabel, StatusLabel})
)

func registerMetrics() {
	prometheus.MustRegister(dnsBlastCounter)
	prometheus.MustRegister(httpCounter)
	prometheus.MustRegister(packetgenCounter)
	prometheus.MustRegister(slowlorisCounter)
	prometheus.MustRegister(rawnetCounter)
}

// ValidatePrometheusPushGateways split value into list of comma separated values and validate that each value
// is valid URL
func ValidatePrometheusPushGateways(value string) bool {
	if len(value) == 0 {
		return true
	}
	listValues := strings.Split(value, ",")
	result := true
	for i, gatewayUrl := range listValues {
		_, err := url.Parse(gatewayUrl)
		if err != nil {
			log.Printf("Can't parse %dth (0-based) push gateway\n", i)
			result = false
		}
	}
	return result
}

// ExportPrometheusMetrics starts http server and export metrics at address <ip>:9090/metrics, also pushes metrics
// to gateways randomly
func ExportPrometheusMetrics(ctx context.Context, gateways string) {
	registerMetrics()

	http.Handle("/metrics", promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			// Opt into OpenMetrics to support exemplars.
			EnableOpenMetrics: true,
			// we don't expect that rendering metrics should take a lot of time
			// and needs long timeout
			Timeout: time.Second * 30,
		},
	))

	server := &http.Server{
		Addr:    "0.0.0.0:9090",
		Handler: nil,
	}
	go func(ctx context.Context, server *http.Server) {
		<-ctx.Done()
		server.Shutdown(ctx)
	}(ctx, server)

	go pushMetrics(ctx, strings.Split(gateways, ","))
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal()
	}
}

func pushMetrics(ctx context.Context, gateways []string) {
	// case of empty string splitted by empty delimiter
	if len(gateways) == 1 && gateways[0] == "" {
		return
	}
	jobName := utils.GetEnvStringDefault("PROMETHEUS_JOB_NAME", "default_push")

	gateway := gateways[rand.Int()%len(gateways)]
	tickerPeriodEnv := utils.GetEnvStringDefault("PROMETHEUS_PUSH_PERIOD", "1m")
	tickerPeriod, err := time.ParseDuration(tickerPeriodEnv)
	if err != nil {
		log.Println("Invalid value for <PROMETHEUS_PUSH_PERIOD> env variable. Read docs: https://pkg.go.dev/time#ParseDuration")
	}
	ticker := time.NewTicker(tickerPeriod)
	pusher := push.New(gateway, jobName).Gatherer(prometheus.DefaultGatherer)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := pusher.Push(); err != nil {
				log.Println("Can't push metrics to gateway, trying to change gateway")
				gateway = gateways[rand.Int()%len(gateways)]
			}
		}
	}
}

// IncDNSBlast increments counter of sent dns queries
func IncDNSBlast(rootDomain, seedDomain, protocol, status string) {
	dnsBlastCounter.With(prometheus.Labels{
		DNSBlastRootDomainLabel: rootDomain,
		DNSBlastSeedDomainLabel: seedDomain,
		DNSBlastProtocolLabel:   protocol,
		StatusLabel:             status}).Inc()
}

// IncHTTP increments counter of sent http queries
func IncHTTP(host, method, status string) {
	httpCounter.With(prometheus.Labels{
		HTTPMethodLabel:          method,
		HTTPDestinationHostLabel: host,
		StatusLabel:              status,
	}).Inc()
}

// IncPacketgen increments counter of sent raw packets
func IncPacketgen(host, host_port, protocol, status string) {
	packetgenCounter.With(prometheus.Labels{
		PacketgenHostLabel:        host,
		PacketgenDstHostPortLabel: host_port,
		PacketgenProtocolLabel:    protocol,
		StatusLabel:               status,
	}).Inc()
}

// IncSlowLoris increments counter of sent raw ethernet+ip+tcp/udp packets
func IncSlowLoris(address, protocol, status string) {
	slowlorisCounter.With(prometheus.Labels{
		SlowlorisAddressLabel:  address,
		SlowlorisProtocolLabel: protocol,
		StatusLabel:            status,
	}).Inc()
}

// IncRawnetTCP increments counter of sent raw tcp packets
func IncRawnetTCP(address, status string) {
	rawnetCounter.With(prometheus.Labels{
		RawnetAddressLabel:  address,
		RawnetProtocolLabel: "tcp",
		StatusLabel:         status,
	}).Inc()
}

// IncRawnetUDP increments counter of sent raw tcp packets
func IncRawnetUDP(address, status string) {
	rawnetCounter.With(prometheus.Labels{
		RawnetAddressLabel:  address,
		RawnetProtocolLabel: "udp",
		StatusLabel:         status,
	}).Inc()
}
