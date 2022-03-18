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
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"

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

func InitMetrics(clientID, country string) {
	constLabels := prometheus.Labels{ClientIDLabel: clientID}
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
func ValidatePrometheusPushGateways(value string) bool {
	if len(value) == 0 {
		return true
	}

	listValues := strings.Split(value, ",")
	result := true

	for i, gatewayURL := range listValues {
		if _, err := url.Parse(gatewayURL); err != nil {
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

	if gateways != "" {
		go pushMetrics(ctx, strings.Split(gateways, ","))
	}

	serveMetrics(ctx)
}

// BasicAuth client's credentials for push gateway encrypted with utils/crypto.go#EncryptionKeys[0] key
var BasicAuth = `YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+IHNjcnlwdCBpYWlSV1VBRWcweEt2NWdTd240a0JBIDE4CmlONnhLcURxWEVWdmFuU1RhSVl0dmplNGpLc0FqLzN5SE5neXdnM0xIMVUKLS0tIE1YdVNBVmk1NG9zNzRpQnh2R3U3MDBpWm5MNUxCb0hNeGxKTERGRDFRamMKJkpimmJGSDmxBX2e38Z38EQZK7aq/W29YMbZKz/omNL0GPvurXZA6GTPmmlD/XZ+EjCkW6bKajIS9y9533tsn6MR8NMtFJoS+z7M9b/yd8YJR6fW069b2A==`

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

func pushMetrics(ctx context.Context, gateways []string) {
	jobName := utils.GetEnvStringDefault("PROMETHEUS_JOB_NAME", "default_push")
	gateway := gateways[rand.Intn(len(gateways))] //nolint:gosec // Cryptographically secure random not required
	tickerPeriod := utils.GetEnvDurationDefault("PROMETHEUS_PUSH_PERIOD", time.Minute)
	ticker := time.NewTicker(tickerPeriod)

	tlsConfig, err := getTLSConfig()
	if err != nil {
		log.Println("Can't get tls config")

		return
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	user, password, err := getBasicAuth()
	if err != nil {
		log.Println("Can't fetch basic auth credentials")

		return
	}

	pusher := push.New(gateway, jobName).Gatherer(prometheus.DefaultGatherer).Client(httpClient).BasicAuth(user, password)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := pusher.Push(); err != nil {
				log.Println("Can't push metrics to gateway, trying to change gateway")

				gateway = gateways[rand.Intn(len(gateways))] //nolint:gosec // Cryptographically secure random not required
				pusher = push.New(gateway, jobName).Gatherer(prometheus.DefaultGatherer).Client(httpClient).BasicAuth(user, password)
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
		StatusLabel:             status,
	}).Inc()
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
func IncPacketgen(host, hostPort, protocol, status, id string) {
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

func IncClient() {
	clientCounter.With(prometheus.Labels{}).Inc()
}
