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

// Package metrics collects and reports job metrics.
package metrics

import (
	"sort"
	"sync"

	"go.uber.org/zap"
)

// Stat is the type of statistical metrics.
type Stat int

const (
	RequestsAttemptedStat Stat = iota
	RequestsSentStat
	ResponsesReceivedStat
	BytesSentStat

	NumStats
)

// String representation of the Stat.
func (s Stat) String() string {
	return [...]string{
		"requests attempted",
		"requests sent",
		"responses received",
		"bytes sent",
	}[s]
}

type dimensions struct {
	jobID  string
	target string
}

// Reporter gathers metrics across jobs and reports them.
// Concurrency-safe.
type Reporter struct {
	metrics [NumStats]sync.Map // Array of metrics by Stat. Each metric is a map of uint64 values by dimensions.

	clientID string
}

// NewReporter creates a new Reporter.
func NewReporter(clientID string) *Reporter { return &Reporter{clientID: clientID} }

// Sum returns a total sum of metric s.
func (r *Reporter) Sum(s Stat) uint64 {
	var res uint64

	r.metrics[s].Range(func(_, v interface{}) bool {
		value, ok := v.(uint64)
		if !ok {
			return true
		}

		res += value

		return true
	})

	return res
}

type (
	// Stats contains all metrics packed as an array.
	Stats [NumStats]uint64
	// MultiStats contains multiple Stats as a slice. Useful for collecting Stats from coroutines.
	MultiStats []Stats
	// PerTargetStats is a map of Stats per target.
	PerTargetStats map[string]Stats
)

// NewStats returns a new Stats record.
func NewStats(requestsAttempted, requestsSent, responsesReceived, bytesSent uint64) Stats {
	return Stats{requestsAttempted, requestsSent, responsesReceived, bytesSent}
}

// Sum up all Stats into a total Stats record.
func (s MultiStats) Sum() Stats {
	var res Stats

	for i := range s {
		for j := RequestsAttemptedStat; j < NumStats; j++ {
			res[j] += s[i][j]
		}
	}

	return res
}

func (ts PerTargetStats) sum(s Stat) uint64 {
	var res uint64

	for _, v := range ts {
		res += v[s]
	}

	return res
}

func (ts PerTargetStats) sortedTargets() []string {
	res := make([]string, 0, len(ts))
	for k := range ts {
		res = append(res, k)
	}

	sort.Strings(res)

	return res
}

// SumAllStatsByTarget returns a total sum of all metrics by target.
func (r *Reporter) SumAllStatsByTarget() PerTargetStats {
	res := make(PerTargetStats)

	for s := RequestsAttemptedStat; s < NumStats; s++ {
		r.metrics[s].Range(func(k, v interface{}) bool {
			d, ok := k.(dimensions)
			if !ok {
				return true
			}

			value, ok := v.(uint64)
			if !ok {
				return true
			}

			stats := res[d.target]
			stats[s] += value
			res[d.target] = stats

			return true
		})
	}

	return res
}

// WriteSummary dumps Reporter contents into the target.
func (r *Reporter) WriteSummary(logger *zap.Logger) {
	stats := r.SumAllStatsByTarget()

	var totals Stats

	for _, tgt := range stats.sortedTargets() {
		tgtStats := stats[tgt]
		logger.Info("stats", zap.String("target", tgt),
			zap.Uint64("requests_attempted", tgtStats[RequestsAttemptedStat]),
			zap.Uint64("requests_sent", tgtStats[RequestsSentStat]),
			zap.Uint64("responses_received", tgtStats[ResponsesReceivedStat]),
			zap.Uint64("bytes_sent", tgtStats[BytesSentStat]))

		for s := range totals {
			totals[s] += tgtStats[s]
		}
	}

	logger.Info("stats", zap.String("target", "total"),
		zap.Uint64("requests_attempted", totals[RequestsAttemptedStat]),
		zap.Uint64("requests_sent", totals[RequestsSentStat]),
		zap.Uint64("responses_received", totals[ResponsesReceivedStat]),
		zap.Uint64("bytes_sent", totals[BytesSentStat]))
}

// NewAccumulator returns a new metrics Accumulator for the Reporter.
func (r *Reporter) NewAccumulator(jobID string) *Accumulator {
	if r == nil {
		return nil
	}

	return newAccumulator(jobID, r)
}

// Accumulator for statistical metrics for use in a single job. Requires Flush()-ing to Reporter.
// Not concurrency-safe.
type Accumulator struct {
	jobID   string
	metrics [NumStats]map[string]uint64 // Array of metrics by Stat. Each metric is a map of uint64 values by target.

	r *Reporter
}

// Add n to the Accumulator Stat value. Returns self for chaining.
func (a *Accumulator) Add(target string, s Stat, n uint64) *Accumulator {
	a.metrics[s][target] += n

	return a
}

// Inc increases Accumulator Stat value by 1. Returns self for chaining.
func (a *Accumulator) Inc(target string, s Stat) *Accumulator { return a.Add(target, s, 1) }

// AddStats to the Accumulator. Returns self for chaining.
func (a *Accumulator) AddStats(target string, s Stats) *Accumulator {
	for i := range a.metrics {
		a.metrics[i][target] += s[i]
	}

	return a
}

// Flush Accumulator contents to the Reporter.
func (a *Accumulator) Flush() {
	for stat := RequestsAttemptedStat; stat < NumStats; stat++ {
		for target, value := range a.metrics[stat] {
			a.r.metrics[stat].Store(dimensions{jobID: a.jobID, target: target}, value)
		}
	}
}

// Clone a new, blank metrics Accumulator with the same Reporter as the original.
func (a *Accumulator) Clone(jobID string) *Accumulator {
	if a == nil {
		return nil
	}

	return newAccumulator(jobID, a.r)
}

func newAccumulator(jobID string, r *Reporter) *Accumulator {
	res := &Accumulator{
		jobID: jobID,
		r:     r,
	}

	for s := RequestsAttemptedStat; s < NumStats; s++ {
		res.metrics[s] = make(map[string]uint64)
	}

	return res
}

// NopWriter implements io.Writer interface to simply track how much data has to be serialized
type NopWriter struct{}

func (w NopWriter) Write(p []byte) (int, error) { return len(p), nil }
