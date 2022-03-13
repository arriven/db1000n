package dnsblast

import (
	"context"
	"testing"
)

func TestDistinctHeavyHitterGenerator(t *testing.T) {
	t.Parallel()

	seedDomains := []string{
		"host.local",
		"laptop.local",
		"printer.local",
	}

	dhhGenerator, err := NewDistinctHeavyHitterGenerator(context.Background(), seedDomains)
	if err != nil {
		t.Errorf("failed to create a DHH domain generator")

		return
	}

	subDomainsGeneratedNumber := 0

	for {
		newDomain, ok := <-dhhGenerator.Next()
		if !ok {
			break
		}

		t.Logf("new domain: %s", newDomain)

		if subDomainsGeneratedNumber++; subDomainsGeneratedNumber == dhhGeneratorBufferSize {
			dhhGenerator.Cancel()
		}
	}
}
