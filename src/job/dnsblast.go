// MIT License

// Copyright (c) [2022] [Bohdan Ivashko (https://github.com/Arriven)]

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package job

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/Arriven/db1000n/src/core/dnsblast"
	"github.com/Arriven/db1000n/src/job/config"
)

type dnsBlastConfig struct {
	BasicJobConfig
	RootDomain      string
	Protocol        string // "udp", "tcp", "tcp-tls"
	SeedDomains     []string
	ParallelQueries int
}

func dnsBlastJob(ctx context.Context, logger *zap.Logger, globalConfig *GlobalConfig, args config.Args) (data any, err error) {
	jobConfig, err := getDNSBlastConfig(args, globalConfig)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup

	err = dnsblast.Start(ctx, logger, &wg, &dnsblast.Config{
		RootDomain:      jobConfig.RootDomain,
		Protocol:        jobConfig.Protocol,
		SeedDomains:     jobConfig.SeedDomains,
		ParallelQueries: jobConfig.ParallelQueries,
		Delay:           time.Duration(jobConfig.IntervalMs) * time.Millisecond,
		ClientID:        globalConfig.ClientID,
	})

	wg.Wait()

	return nil, err
}

func getDNSBlastConfig(args config.Args, globalConfig *GlobalConfig) (*dnsBlastConfig, error) {
	const (
		defaultParallelQueriesPerCycle = 5
		defaultIntervalMS              = 10
	)

	var jobConfig dnsBlastConfig
	if err := ParseConfig(&jobConfig, args, *globalConfig); err != nil {
		return nil, fmt.Errorf("failed to parse DNS Blast job configurations: %w", err)
	}

	//
	// Default settings and early misconfiguration prevention
	//

	switch {
	case len(jobConfig.RootDomain) == 0:
		return nil, errors.New("no root domain provided, consider adding it")
	case len(jobConfig.SeedDomains) == 0:
		return nil, errors.New("no seed domains provided, at least one is required")
	}

	switch jobConfig.Protocol {
	case dnsblast.UDPProtoName, dnsblast.TCPProtoName, dnsblast.TCPTLSProtoName:
		// All good
	case "":
		// Default to UDP
		jobConfig.Protocol = dnsblast.UDPProtoName
	default:
		return nil, fmt.Errorf("unrecognized DNS protocol %q, expected one of [%q, %q, %q]", jobConfig.Protocol,
			dnsblast.UDPProtoName, dnsblast.TCPProtoName, dnsblast.TCPTLSProtoName)
	}

	if jobConfig.ParallelQueries == 0 {
		jobConfig.ParallelQueries = defaultParallelQueriesPerCycle
	}

	if jobConfig.IntervalMs == 0 {
		jobConfig.IntervalMs = defaultIntervalMS
	}

	return &jobConfig, nil
}
