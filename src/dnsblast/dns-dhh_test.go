package dnsblast

import (
	"testing"
)

func TestDistinctHeavyHitterGenerator(t *testing.T) {
	var (
		seedDomains = []string{
			"host.local",
			"laptop.local",
			"printer.local",
		}
	)

	dhhGenerator, err := NewDistinctHeavyHitterGenerator(seedDomains)
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

		if subDomainsGeneratedNumber += 1; subDomainsGeneratedNumber == dhhGeneratorBufferSize {
			dhhGenerator.Cancel()
		}
	}
}
