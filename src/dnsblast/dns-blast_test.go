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
		testcase := testCases[i]
		t.Run(testcase.Name, func(tt *testing.T) {
			config := &Config{
				RootDomain:      testcase.RootDomain,
				Protocol:        testcase.Protocol,
				SeedDomains:     testcase.SeedDomains,
				Delay:           testcase.Delay,
				ParallelQueries: testcase.ParallelQueries,
			}

			tt.Logf("[%s] benchmark configuration: %+v", testcase.Name, config)

			blastContext, cancel := context.WithTimeout(context.Background(), testcase.Duration)
			defer cancel()

			err := Start(blastContext, config)
			if err != nil {
				tt.Errorf("failed to start the blaster: %s", err)
				return
			}

			time.Sleep(testcase.Duration + time.Second)
		})
	}
}

func TestGetSeedDomain(t *testing.T) {
	seedDomain := `example.com`
	generator, err := NewDistinctHeavyHitterGenerator([]string{seedDomain})
	if err != nil {
		t.Fatal(err)
	}
	count := 10
	for subdomain := range generator.Next() {
		resultSeedDomain := getSeedDomain(subdomain)
		if resultSeedDomain != seedDomain {
			t.Fatalf("Expect \"%s\", took \"%s\"\n", seedDomain, resultSeedDomain)
		}
		count--
		if count == 0 {
			break
		}
	}
}
