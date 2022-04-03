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
	"fmt"
	"io"
	"sort"
	"sync"
	"text/tabwriter"
)

// Stat is the type of statistical metrics.
type Stat int

const (
	RequestsAttemptedStat Stat = iota
	RequestsSentStat
	ResponsesReceivedStat
	BytesSentStat
	BytesReceivedStat

	NumStats
)

// String representation of the Stat.
func (s Stat) String() string {
	if !s.valid() {
		return "<unknown>"
	}

	return [...]string{
		"requests attempted",
		"requests sent",
		"responses received",
		"bytes sent",
		"bytes received",
	}[s]
}

func (s Stat) valid() bool { return s >= RequestsAttemptedStat && s < NumStats }

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
	if r == nil || !s.valid() {
		return 0
	}

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

type PerTargetStats map[string][NumStats]uint64

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
	if r == nil {
		return nil
	}

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
func (r *Reporter) WriteSummary(target io.Writer) {
	w := tabwriter.NewWriter(target, 1, 1, 1, ' ', tabwriter.AlignRight)

	defer w.Flush()

	stats := r.SumAllStatsByTarget()
	if stats.sum(RequestsSentStat) == 0 {
		fmt.Fprintln(w, "[Error] No traffic generated. If you see this message a lot - contact admins")

		return
	}

	fmt.Fprintln(w, "\n\n!Атака проводиться успішно! Русскій воєнний корабль іди нахуй!")
	fmt.Fprintln(w, "!Attack is successful! Russian warship, go fuck yourself!")

	fmt.Fprintln(w, "\n --- Traffic stats ---")
	fmt.Fprintf(w, "|\tTarget\t|\tRequests attempted\t|\tRequests sent\t|\tResponses received\t|\tData sent\t|\tData received \t|\n")

	const BytesInMegabyte = 1024 * 1024
	var totals [NumStats]uint64

	for _, tgt := range stats.sortedTargets() {
		tgtStats := stats[tgt]

		fmt.Fprintf(w, "|\t%s\t|\t%d\t|\t%d\t|\t%d\t|\t%.2f MB\t|\t%.2f MB \t|\n", tgt,
			tgtStats[RequestsAttemptedStat],
			tgtStats[RequestsSentStat],
			tgtStats[ResponsesReceivedStat],
			float64(tgtStats[BytesSentStat])/BytesInMegabyte,
			float64(tgtStats[BytesReceivedStat])/BytesInMegabyte,
		)

		for s := range totals {
			totals[s] += tgtStats[s]
		}
	}

	fmt.Fprintln(w, "|\t---\t|\t---\t|\t---\t|\t---\t|\t---\t|\t--- \t|")
	fmt.Fprintf(w, "|\tTotal\t|\t%d\t|\t%d\t|\t%d\t|\t%.2f MB\t|\t%.2f MB \t|\n\n",
		totals[RequestsAttemptedStat],
		totals[RequestsSentStat],
		totals[ResponsesReceivedStat],
		float64(totals[BytesSentStat])/BytesInMegabyte,
		float64(totals[BytesReceivedStat])/BytesInMegabyte,
	)
}

// NewAccumulator returns a new metrics Accumulator for the Reporter.
func (r *Reporter) NewAccumulator(jobID string) *Accumulator {
	if r == nil {
		return nil
	}

	res := &Accumulator{
		jobID: jobID,
		r:     r,
	}

	for s := RequestsAttemptedStat; s < NumStats; s++ {
		res.metrics[s] = make(map[string]uint64)
	}

	return res
}

// Accumulator for statistical metrics for use in a single job. Requires Flush()-ing to Reporter.
// Not concurrency-safe.
type Accumulator struct {
	jobID   string
	metrics [NumStats]map[string]uint64 // Array of metrics by Stat. Each metric is a map of uint64 values by target.

	r *Reporter
}

// Add n to the Accumulator Stat value.
func (a *Accumulator) Add(s Stat, target string, n uint64) {
	if a == nil || !s.valid() {
		return
	}

	a.metrics[s][target] += n
}

// Inc increases Accumulator Stat value by 1.
func (a *Accumulator) Inc(s Stat, target string) { a.Add(s, target, 1) }

// Flush Accumulator contents to the Reporter.
func (a *Accumulator) Flush() {
	if a == nil {
		return
	}

	for stat := RequestsAttemptedStat; stat < NumStats; stat++ {
		for target, value := range a.metrics[stat] {
			a.r.metrics[stat].Store(dimensions{jobID: a.jobID, target: target}, value)
		}
	}
}

// NopWriter implements io.Writer interface to simply track how much data has to be serialized
type NopWriter struct{}

func (w NopWriter) Write(p []byte) (int, error) { return len(p), nil }
