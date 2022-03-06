package dnsblast

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"
)

func TestBlast(t *testing.T) {
	const (
		testDuration = 10 * time.Second

		testServer = "10.8.0.1"
		testPort   = DefaultDNSPort
		testProto  = UDPProtoName

		domainA = "cnn.com"
		domainB = "yahoo.com"
		domainC = "foxnews.com"

		testIterationDelay = 50 * time.Millisecond
		testParallelism    = 3
	)

	var (
		blastContext, cancel = context.WithTimeout(context.Background(), testDuration)
		config               = &Config{
			RootDomain: net.JoinHostPort(testServer, strconv.Itoa(testPort)),
			Protocol:   testProto,
			SeedDomains: []string{
				domainA,
				domainB,
				domainC,
			},
			Delay:           testIterationDelay,
			ParallelQueries: testParallelism,
		}
	)
	defer cancel()

	err := Start(blastContext, config)
	if err != nil {
		t.Errorf("failed to start the blaster: %s", err)
		return
	}

	time.Sleep(testDuration + time.Second)
}
