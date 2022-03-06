// Package dnsblast [DNS Distinct Heavy Hitter, refer to the scientific article https://faculty.idc.ac.il/bremler/Papers/HotWeb_18.pdf]
package dnsblast

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"
)

const (
	dhhGeneratorBufferSize = 128
)

type DistinctHeavyHitterGenerator struct {
	buffer chan string

	randomizer           *rand.Rand
	randomizerDictionary []rune

	cancelOnce       sync.Once
	cancelGeneration func()
}

// NewDistinctHeavyHitterGenerator creates an endless stream of fake domain names
// using the random subdomain.
func NewDistinctHeavyHitterGenerator(seedDomains []string) (*DistinctHeavyHitterGenerator, error) {
	if len(seedDomains) == 0 {
		return nil, errors.New("no root base domain seeds provided")
	}

	generator := &DistinctHeavyHitterGenerator{
		buffer:               make(chan string, dhhGeneratorBufferSize),
		randomizer:           rand.New(rand.NewSource(time.Now().Unix())),
		randomizerDictionary: []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"),
	}

	ctx, cancel := context.WithCancel(context.Background())
	generator.cancelGeneration = cancel

	go generator.ignite(ctx, seedDomains)

	return generator, nil
}

func (rcv *DistinctHeavyHitterGenerator) Next() chan string {
	return rcv.buffer
}

func (rcv *DistinctHeavyHitterGenerator) Cancel() {
	rcv.cancelOnce.Do(func() {rcv.cancelGeneration()})
}

func (rcv *DistinctHeavyHitterGenerator) ignite(ctx context.Context, rootDomains []string) {
	totalSeeds := len(rootDomains)
	currentDomainIndex := 0

	for {
		select {
		case <-ctx.Done():
			close(rcv.buffer)
			return
		default:
			rcv.buffer <- rcv.generateSubdomain() + "." + rootDomains[currentDomainIndex] + "."

			currentDomainIndex += 1
			if currentDomainIndex == totalSeeds {
				currentDomainIndex = 0
			}
		}
	}
}

const (
	subdomainMinLength = 3
	subdomainMaxLength = 64
)

func (rcv *DistinctHeavyHitterGenerator) generateSubdomain() string {
	n := subdomainMinLength + rcv.randomizer.Intn(subdomainMaxLength-subdomainMinLength)
	b := make([]rune, n)
	for i := range b {
		b[i] = rcv.randomizerDictionary[rcv.randomizer.Intn(len(rcv.randomizerDictionary))]
	}
	return string(b)
}
