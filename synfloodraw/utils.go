package synfloodraw

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"regexp"
)

// getRandomPayload returns a byte slice to spoof ip packets with random payload in specified length
func getRandomPayload(length int) []byte {
	payload := make([]byte, length)
	rand.Read(payload)
	return payload
}

// getIps returns a string slice to spoof ip packets with dummy source ip addresses
func getIps() []string {
	ips := make([]string, 0)
	for i := 0; i < 20; i++ {
		ips = append(ips, fmt.Sprintf("%d.%d.%d.%d", rand.Intn(256), rand.Intn(256),
			rand.Intn(256), rand.Intn(256)))
	}

	return ips
}

// getPorts returns an int slice to spoof ip packets with dummy source ports
func getPorts() []int {
	ports := make([]int, 0)
	for i := 1024; i <= 65535; i++ {
		ports = append(ports, i)
	}

	return ports
}

// getMacAddrs returns a byte slice to spoof ip packets with dummy MAC addresses
func getMacAddrs() [][]byte {
	macAddrs := make([][]byte, 0)
	for i := 0; i <= 50; i++ {
		buf := make([]byte, 6)
		_, err := rand.Read(buf)
		if err != nil {
			fmt.Println("error:", err)
			continue
		}
		macAddrs = append(macAddrs, buf)
	}

	return macAddrs
}

// isDNS returns a boolean which indicates host parameter is a DNS record or not
func isDNS(host string) bool {
	var (
		res bool
		err error
	)

	if res, err = regexp.MatchString(DnsRegex, host); err != nil {
		log.Fatalf("a fatal error occured while matching provided --host with DNS regex: %s", err.Error())
	}

	return res
}

// isIP returns a boolean which indicates host parameter is an IP address or not
func isIP(host string) bool {
	var (
		res bool
		err error
	)

	if res, err = regexp.MatchString(IpRegex, host); err != nil {
		log.Fatalf("a fatal error occured while matching provided --host with IP regex: %s", err.Error())
	}

	return res
}

// resolveHost function gets a string and returns the ip address while deciding it is an ip address already or DNS record
func resolveHost(host string) (string, error) {
	if !isIP(host) && isDNS(host) {
		log.Printf("%s is a DNS record, making DNS lookup\n", host)
		ipRecords, err := net.DefaultResolver.LookupIP(context.Background(), "ip4", host)
		if err != nil {
			log.Printf("an error occured on dns lookup: %s\n", err.Error())
			return "", err
		}

		log.Printf("dns lookup succeeded, found %s for %s\n", ipRecords[0].String(), host)
		host = ipRecords[0].String()
	} else {
		log.Printf("%s is already an IP address, skipping DNS resolution\n", host)
	}

	return host, nil
}
