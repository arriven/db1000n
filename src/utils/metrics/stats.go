package metrics

import (
	"sort"

	"go.uber.org/zap/zapcore"
)

type (
	// Stat is the type of statistical metrics.
	Stat int
	// Stats contains all metrics packed as an array.
	Stats [NumStats]uint64
	// MultiStats contains multiple Stats as a slice. Useful for collecting Stats from coroutines.
	MultiStats []Stats
	// PerTargetStats is a map of Stats per target.
	PerTargetStats map[string]Stats
)

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

// MarshalLogObject is required to log PerTargetStats objects to zap
func (ts PerTargetStats) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	for _, tgt := range ts.sortedTargets() {
		tgtStats := ts[tgt]

		if err := enc.AddObject(tgt, &tgtStats); err != nil {
			return err
		}
	}

	return nil
}

// MarshalLogObject is required to log Stats objects to zap
func (stats *Stats) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddUint64("requests_attempted", stats[RequestsAttemptedStat])
	enc.AddUint64("requests_sent", stats[RequestsSentStat])
	enc.AddUint64("responses_received", stats[ResponsesReceivedStat])
	enc.AddUint64("bytes_sent", stats[BytesSentStat])

	return nil
}
