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
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"flag"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
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
func NewOptionsWithFlags() (prometheusOn *bool, prometheusPushGateways *string) {
	return flag.Bool("prometheus_on", utils.GetEnvBoolDefault("PROMETHEUS_ON", true),
			"Start metrics exporting via HTTP and pushing to gateways (specified via <prometheus_gateways>)"),
		flag.String("prometheus_gateways",
			utils.GetEnvStringDefault("PROMETHEUS_GATEWAYS", "https://178.62.78.144:9091,https://46.101.26.43:9091,https://178.62.33.149:9091"),
			"Comma separated list of prometheus push gateways")
}

func InitOrFail(ctx context.Context, logger *zap.Logger, prometheusOn bool, prometheusPushGateways, clientID, country string) {
	if !ValidatePrometheusPushGateways(prometheusPushGateways) {
		log.Fatal("Invalid value for --prometheus_gateways")
	}

	if prometheusOn {
		Init(clientID, country)

		go ExportPrometheusMetrics(ctx, logger, clientID, prometheusPushGateways)
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

// ValidatePrometheusPushGateways split value into list of comma separated values and validate that each value
// is valid URL
func ValidatePrometheusPushGateways(gatewayURLsCSV string) bool {
	if len(gatewayURLsCSV) == 0 {
		return true
	}

	for i, gatewayURL := range strings.Split(gatewayURLsCSV, ",") {
		if _, err := url.Parse(gatewayURL); err != nil {
			log.Printf("Can't parse %d-th (0-based) push gateway: %v", i, err)

			return false
		}
	}

	return true
}

// ExportPrometheusMetrics starts http server and export metrics at address <ip>:9090/metrics, also pushes metrics
// to gateways randomly
func ExportPrometheusMetrics(ctx context.Context, logger *zap.Logger, clientID, gateways string) {
	registerMetrics()

	if gateways != "" {
		go pushMetrics(ctx, logger, clientID, strings.Split(gateways, ","))
	}

	serveMetrics(ctx)
}

// BasicAuth client's credentials for push gateway encrypted with utils/crypto.go#EncryptionKeys[0] key
var BasicAuth = `YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+IHNjcnlwdCBpYWlSV1VBRWcweEt2NWdTd240a0JBIDE4CmlONnhLcURxWEVWdmFuU1Rh` +
	`SVl0dmplNGpLc0FqLzN5SE5neXdnM0xIMVUKLS0tIE1YdVNBVmk1NG9zNzRpQnh2R3U3MDBpWm5MNUxCb0hNeGxKTERGRDFRamMKJkpimmJGSDmx` +
	`BX2e38Z38EQZK7aq/W29YMbZKz/omNL0GPvurXZA6GTPmmlD/XZ+EjCkW6bKajIS9y9533tsn6MR8NMtFJoS+z7M9b/yd8YJR6fW069b2A==`

// getBasicAuth returns decrypted basic auth credential for push gateway
func getBasicAuth() (string, string, error) {
	encryptedData, err := base64.StdEncoding.DecodeString(BasicAuth)
	if err != nil {
		return "", "", err
	}

	decryptedData, err := utils.Decrypt(encryptedData)
	if err != nil {
		return "", "", err
	}

	const numBasicAuthParts = 2

	parts := bytes.Split(decryptedData, []byte{':'})
	if len(parts) != numBasicAuthParts {
		return "", "", errors.New("invalid basic auth credential format")
	}

	return string(parts[0]), string(parts[1]), nil
}

// PushGatewayCA variable to embed self-signed CA for TLS
var PushGatewayCA string

// getTLSConfig returns tls.Config with system root CAs and embedded CA if not empty
func getTLSConfig() (*tls.Config, error) {
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		rootCAs = x509.NewCertPool()
	}

	if PushGatewayCA != "" {
		decoded, err := base64.StdEncoding.DecodeString(PushGatewayCA)
		if err != nil {
			return nil, err
		}

		decrypted, err := utils.Decrypt(decoded)
		if err != nil {
			return nil, err
		}

		if ok := rootCAs.AppendCertsFromPEM(decrypted); !ok {
			return nil, errors.New("invalid embedded CA")
		}
	}

	return &tls.Config{
		RootCAs:    rootCAs,
		MinVersion: tls.VersionTLS12,
	}, nil
}

func pushMetrics(ctx context.Context, logger *zap.Logger, clientID string, gateways []string) {
	jobName := utils.GetEnvStringDefault("PROMETHEUS_JOB_NAME", "db1000n_default_add")
	gateway := gateways[rand.Intn(len(gateways))] //nolint:gosec // Cryptographically secure random not required
	tickerPeriod := utils.GetEnvDurationDefault("PROMETHEUS_PUSH_PERIOD", time.Minute)
	ticker := time.NewTicker(tickerPeriod)

	tlsConfig, err := getTLSConfig()
	if err != nil {
		logger.Debug("Can't get tls config", zap.Error(err))

		return
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	user, password, err := getBasicAuth()
	if err != nil {
		logger.Debug("Can't fetch basic auth credentials", zap.Error(err))

		return
	}

	pusher := setupPusher(push.New(gateway, jobName), clientID, httpClient, user, password)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := pusher.Add(); err != nil {
				logger.Debug("Can't push metrics to gateway, changing gateway", zap.Error(err))

				gateway = gateways[rand.Intn(len(gateways))] //nolint:gosec // Cryptographically secure random not required
				pusher = setupPusher(push.New(gateway, jobName), clientID, httpClient, user, password)
			}
		}
	}
}

func setupPusher(pusher *push.Pusher, clientID string, httpClient push.HTTPDoer, user, password string) *push.Pusher {
	return pusher.Gatherer(prometheus.DefaultGatherer).Grouping(ClientIDLabel, clientID).Client(httpClient).BasicAuth(user, password)
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
