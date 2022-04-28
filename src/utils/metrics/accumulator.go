package metrics

// Accumulator for statistical metrics for use in a single job. Requires Flush()-ing to Reporter.
// Not concurrency-safe.
type Accumulator struct {
	jobID   string
	stats   [NumStats]map[string]uint64 // Array of metrics by Stat. Each metric is a map of uint64 values by target.
	metrics *Metrics
}

type dimensions struct {
	jobID  string
	target string
}

// Add n to the Accumulator Stat value. Returns self for chaining.
func (a *Accumulator) Add(target string, s Stat, n uint64) *Accumulator {
	a.stats[s][target] += n

	return a
}

// Inc increases Accumulator Stat value by 1. Returns self for chaining.
func (a *Accumulator) Inc(target string, s Stat) *Accumulator { return a.Add(target, s, 1) }

// Flush Accumulator contents to the Reporter.
func (a *Accumulator) Flush() {
	for stat := RequestsAttemptedStat; stat < NumStats; stat++ {
		for target, value := range a.stats[stat] {
			a.metrics[stat].Store(dimensions{jobID: a.jobID, target: target}, value)
		}
	}
}

// Clone a new, blank metrics Accumulator with the same Reporter as the original.
func (a *Accumulator) Clone(jobID string) *Accumulator {
	if a == nil {
		return nil
	}

	return newAccumulator(jobID, a.metrics)
}

func newAccumulator(jobID string, data *Metrics) *Accumulator {
	res := &Accumulator{
		jobID:   jobID,
		metrics: data,
	}

	for s := RequestsAttemptedStat; s < NumStats; s++ {
		res.stats[s] = make(map[string]uint64)
	}

	return res
}
