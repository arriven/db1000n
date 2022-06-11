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
	"strings"
	"sync"
)

type Metrics [NumStats]sync.Map // Array of metrics by Stat. Each metric is a map of uint64 values by dimensions.

// NewAccumulator returns a new metrics Accumulator for the Reporter.
func (m *Metrics) NewAccumulator(jobID string) *Accumulator {
	if m == nil {
		return nil
	}

	return newAccumulator(jobID, m)
}

// Calculates all targets and total stats
func (m *Metrics) SumAllStats(groupTargets bool) (stats PerTargetStats, totals Stats) {
	stats = m.sumAllStatsByTarget(groupTargets)

	for s := RequestsAttemptedStat; s < NumStats; s++ {
		totals[s] = m.Sum(s)
	}

	return
}

// Sum returns a total sum of metric s.
func (m *Metrics) Sum(s Stat) uint64 {
	var res uint64

	m[s].Range(func(_, v any) bool {
		value, ok := v.(uint64)
		if !ok {
			return true
		}

		res += value

		return true
	})

	return res
}

// Returns a total sum of all metrics by target.
func (m *Metrics) sumAllStatsByTarget(groupTargets bool) PerTargetStats {
	res := make(PerTargetStats)

	for s := RequestsAttemptedStat; s < NumStats; s++ {
		m[s].Range(func(k, v any) bool {
			d, ok := k.(dimensions)
			if !ok {
				return true
			}

			value, ok := v.(uint64)
			if !ok {
				return true
			}

			var target string
			if groupTargets {
				protocol, _, found := strings.Cut(d.target, "://")
				if found {
					target = protocol
				} else {
					target = "other"
				}
			} else {
				target = d.target
			}

			stats := res[target]
			stats[s] += value
			res[target] = stats

			return true
		})
	}

	return res
}
