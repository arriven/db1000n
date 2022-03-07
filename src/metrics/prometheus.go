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
	"github.com/Arriven/db1000n/src/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/push"
	"log"
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

// registered metrics
var (
	dnsBlastCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_blast_total",
			Help: "Number of dns queries",
		}, []string{DNSBlastRootDomainLabel, DNSBlastSeedDomainLabel, DNSBlastProtocolLabel, StatusLabel})
)

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
	prometheus.MustRegister(dnsBlastCounter)

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
	if len(gateways) == 0 {
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
	pusher := push.New(gateway, jobName)
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

// IncDNSBlast increments counter of sent queries
func IncDNSBlast(rootDomain, seedDomain, protocol, status string) {
	dnsBlastCounter.With(prometheus.Labels{
		DNSBlastRootDomainLabel: rootDomain,
		DNSBlastSeedDomainLabel: seedDomain,
		DNSBlastProtocolLabel:   protocol,
		StatusLabel:             status}).Inc()
}
