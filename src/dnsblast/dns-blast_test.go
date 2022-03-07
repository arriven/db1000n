package dnsblast

import (
	"context"
	"testing"
	"time"
)

func TestBlast(t *testing.T) {
	const (
		testDuration = 5 * time.Second

		testRootDomain = "example.com"
		seedDomainA    = "cnn.com"
		seedDomainB    = "yahoo.com"
		seedDomainC    = "foxnews.com"

		testIterationDelay = 50 * time.Millisecond
		testParallelism    = 3
	)

	type testCase struct {
		Name     string
		Duration time.Duration

		RootDomain      string
		Protocol        string
		SeedDomains     []string
		Delay           time.Duration
		ParallelQueries int
	}

	testCases := []testCase{
		{
			Name:       "Benchmark over UDP",
			Duration:   testDuration,
			RootDomain: testRootDomain,
			Protocol:   UDPProtoName,
			SeedDomains: []string{
				seedDomainA,
				seedDomainB,
				seedDomainC,
			},
			Delay:           testIterationDelay,
			ParallelQueries: testParallelism,
		},
		{
			Name:       "Benchmark over TCP",
			Duration:   testDuration,
			RootDomain: testRootDomain,
			Protocol:   TCPProtoName,
			SeedDomains: []string{
				seedDomainA,
				seedDomainB,
				seedDomainC,
			},
			Delay:           testIterationDelay,
			ParallelQueries: testParallelism,
		},
		{
			Name:       "Benchmark over TCP-TLS",
			Duration:   testDuration,
			RootDomain: testRootDomain,
			Protocol:   TCPTLSProtoName,
			SeedDomains: []string{
				seedDomainA,
				seedDomainB,
				seedDomainC,
			},
			Delay:           testIterationDelay,
			ParallelQueries: testParallelism,
		},
	}

	for i := range testCases {
		t.Run(testCases[i].Name, func(tt *testing.T) {
			config := &Config{
				RootDomain:      testCases[i].RootDomain,
				Protocol:        testCases[i].Protocol,
				SeedDomains:     testCases[i].SeedDomains,
				Delay:           testCases[i].Delay,
				ParallelQueries: testCases[i].ParallelQueries,
			}

			tt.Logf("[%s] benchmark configuration: %+v", testCases[i].Name, config)

			blastContext, cancel := context.WithTimeout(context.Background(), testCases[i].Duration)
			defer cancel()

			err := Start(blastContext, config)
			if err != nil {
				tt.Errorf("failed to start the blaster: %s", err)
				return
			}

			time.Sleep(testCases[i].Duration + time.Second)
		})
	}
}
