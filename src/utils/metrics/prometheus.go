// MIT License

// Copyright (c) [2022] [Bohdan Ivashko (https://github.com/Arriven)]

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
	"flag"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/Arriven/db1000n/src/utils"
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

// Client related values and labels
const (
	ClientIDLabel = `id`
	CountryLabel  = `country`
)

// registered metrics
var (
	dnsBlastCounter  *prometheus.CounterVec
	httpCounter      *prometheus.CounterVec
	packetgenCounter *prometheus.CounterVec
	slowlorisCounter *prometheus.CounterVec
	rawnetCounter    *prometheus.CounterVec
	clientCounter    *prometheus.CounterVec
)

// NewOptionsWithFlags returns metrics options initialized with command line flags.
func NewOptionsWithFlags() (prometheusOn *bool, prometheusListenAddress *string) {
	return flag.Bool("prometheus_on", utils.GetEnvBoolDefault("PROMETHEUS_ON", true),
			"Start metrics exporting via HTTP and pushing to gateways (specified via <prometheus_gateways>)"),
		flag.String("prometheus_listen", utils.GetEnvStringDefault("PROMETHEUS_LISTEN", ":9090"),
			"Address to listen on for metrics endpoint")
}

func InitOrFail(ctx context.Context, logger *zap.Logger, prometheusOn bool, prometheusListenAddress, clientID, country string) {
	if prometheusOn {
		Init(clientID, country)

		go ExportPrometheusMetrics(ctx, logger, clientID, prometheusListenAddress)
	}
}

// Init prometheus counters.
func Init(clientID, country string) {
	constLabels := prometheus.Labels{}
	if country != "" {
		constLabels[CountryLabel] = country
	}

	dnsBlastCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "db1000n_dns_blast_total",
			Help:        "Number of dns queries",
			ConstLabels: constLabels,
		}, []string{DNSBlastRootDomainLabel, DNSBlastSeedDomainLabel, DNSBlastProtocolLabel, StatusLabel})
	httpCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "db1000n_http_request_total",
			Help:        "Number of http queries",
			ConstLabels: constLabels,
		}, []string{HTTPDestinationHostLabel, HTTPMethodLabel, StatusLabel})
	packetgenCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "db1000n_packetgen_total",
			Help:        "Number of packet generation transfers",
			ConstLabels: constLabels,
		}, []string{PacketgenHostLabel, PacketgenDstHostPortLabel, PacketgenProtocolLabel, StatusLabel})
	slowlorisCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "db1000n_slowloris_total",
			Help:        "Number of sent raw tcp/udp packets",
			ConstLabels: constLabels,
		}, []string{SlowlorisAddressLabel, SlowlorisProtocolLabel, StatusLabel})
	rawnetCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "db1000n_rawnet_total",
			Help:        "Number of sent raw tcp/udp packets",
			ConstLabels: constLabels,
		}, []string{RawnetAddressLabel, RawnetProtocolLabel, StatusLabel})
	clientCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:        "db1000n_client_total",
		Help:        "Number of clients",
		ConstLabels: constLabels,
	}, []string{})
}

func registerMetrics() {
	prometheus.MustRegister(dnsBlastCounter)
	prometheus.MustRegister(httpCounter)
	prometheus.MustRegister(packetgenCounter)
	prometheus.MustRegister(slowlorisCounter)
	prometheus.MustRegister(rawnetCounter)
	prometheus.MustRegister(clientCounter)
}

// ExportPrometheusMetrics starts http server and export metrics at address <ip>:9090/metrics, also pushes metrics
// to gateways randomly
func ExportPrometheusMetrics(ctx context.Context, logger *zap.Logger, clientID, listen string) {
	registerMetrics()

	serveMetrics(ctx, logger, listen)
}

// IncDNSBlast increments counter of sent dns queries
func IncDNSBlast(rootDomain, seedDomain, protocol, status string) {
	if dnsBlastCounter == nil {
		return
	}

	dnsBlastCounter.With(prometheus.Labels{
		DNSBlastRootDomainLabel: rootDomain,
		DNSBlastSeedDomainLabel: seedDomain,
		DNSBlastProtocolLabel:   protocol,
		StatusLabel:             status,
	}).Inc()
}

// IncHTTP increments counter of sent http queries
func IncHTTP(host, method, status string) {
	if httpCounter == nil {
		return
	}

	httpCounter.With(prometheus.Labels{
		HTTPMethodLabel:          method,
		HTTPDestinationHostLabel: host,
		StatusLabel:              status,
	}).Inc()
}

// IncPacketgen increments counter of sent raw packets
func IncPacketgen(host, hostPort, protocol, status, id string) {
	if packetgenCounter == nil {
		return
	}

	packetgenCounter.With(prometheus.Labels{
		PacketgenHostLabel:        host,
		PacketgenDstHostPortLabel: hostPort,
		PacketgenProtocolLabel:    protocol,
		StatusLabel:               status,
		ClientIDLabel:             id,
	}).Inc()
}

// IncSlowLoris increments counter of sent raw ethernet+ip+tcp/udp packets
func IncSlowLoris(address, protocol, status string) {
	if slowlorisCounter == nil {
		return
	}

	slowlorisCounter.With(prometheus.Labels{
		SlowlorisAddressLabel:  address,
		SlowlorisProtocolLabel: protocol,
		StatusLabel:            status,
	}).Inc()
}

// IncRawnetTCP increments counter of sent raw tcp packets
func IncRawnetTCP(address, status string) {
	if rawnetCounter == nil {
		return
	}

	rawnetCounter.With(prometheus.Labels{
		RawnetAddressLabel:  address,
		RawnetProtocolLabel: "tcp",
		StatusLabel:         status,
	}).Inc()
}

// IncRawnetUDP increments counter of sent raw tcp packets
func IncRawnetUDP(address, status string) {
	if rawnetCounter == nil {
		return
	}

	rawnetCounter.With(prometheus.Labels{
		RawnetAddressLabel:  address,
		RawnetProtocolLabel: "udp",
		StatusLabel:         status,
	}).Inc()
}

// IncClient increments counter of calls from the current client ID
func IncClient() {
	if clientCounter == nil {
		return
	}

	clientCounter.With(prometheus.Labels{}).Inc()
}
