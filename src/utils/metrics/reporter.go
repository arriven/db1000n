package metrics

import (
	"fmt"
	"io"
	"text/tabwriter"

	"go.uber.org/zap"
)

// Reporter gathers metrics across jobs and reports them.
// Concurrency-safe.
type Reporter interface {
	// WriteSummary dumps Reporter contents into the target.
	WriteSummary()

	SumBytesSent() uint64
	ClearData()

	// NewAccumulator returns a new metrics Accumulator for the Reporter.
	NewAccumulator(jobID string) *Accumulator
}

// ZapReporter

type ZapReporter struct {
	data   ReporterData
	logger *zap.Logger
}

// NewZapReporter creates a new Reporter using a zap logger.
func NewZapReporter(logger *zap.Logger) Reporter {
	return &ZapReporter{logger: logger}
}

func (r *ZapReporter) WriteSummary() {
	stats, totals := r.data.SumAllStats()

	r.logger.Info("stats", zap.Object("total", &totals), zap.Object("targets", stats))
}

func (r *ZapReporter) NewAccumulator(jobID string) *Accumulator {
	if r == nil {
		return nil
	}

	return newAccumulator(jobID, &r.data)
}

func (r *ZapReporter) ClearData() {
	r.data.ClearData()
}

func (r *ZapReporter) SumBytesSent() uint64 {
	return r.data.SumBytesSent()
}

// ConsoleReporter

type ConsoleReporter struct {
	data   ReporterData
	target io.Writer
}

// NewConsoleReporter creates a new Reporter which outputs straight to the console
func NewConsoleReporter(target io.Writer) Reporter {
	return &ConsoleReporter{target: target}
}

func (r *ConsoleReporter) WriteSummary() {
	w := tabwriter.NewWriter(r.target, 1, 1, 1, ' ', tabwriter.AlignRight)

	defer w.Flush()

	stats, totals := r.data.SumAllStats()

	fmt.Fprintln(w, "\n --- Traffic stats ---")
	fmt.Fprintf(w, "|\tTarget\t|\tRequests attempted\t|\tRequests sent\t|\tResponses received\t|\tData sent \t|\n")

	const BytesInMegabyte = 1024 * 1024

	for _, tgt := range stats.sortedTargets() {
		tgtStats := stats[tgt]

		fmt.Fprintf(w, "|\t%s\t|\t%d\t|\t%d\t|\t%d\t|\t%.2f MB \t|\n", tgt,
			tgtStats[RequestsAttemptedStat],
			tgtStats[RequestsSentStat],
			tgtStats[ResponsesReceivedStat],
			float64(tgtStats[BytesSentStat])/BytesInMegabyte,
		)

		for s := range totals {
			totals[s] += tgtStats[s]
		}
	}

	fmt.Fprintln(w, "|\t---\t|\t---\t|\t---\t|\t---\t|\t--- \t|")
	fmt.Fprintf(w, "|\tTotal\t|\t%d\t|\t%d\t|\t%d\t|\t%.2f MB \t|\n\n",
		totals[RequestsAttemptedStat],
		totals[RequestsSentStat],
		totals[ResponsesReceivedStat],
		float64(totals[BytesSentStat])/BytesInMegabyte,
	)
}

func (r *ConsoleReporter) NewAccumulator(jobID string) *Accumulator {
	if r == nil {
		return nil
	}

	return newAccumulator(jobID, &r.data)
}

func (r *ConsoleReporter) ClearData() {
	r.data.ClearData()
}

func (r *ConsoleReporter) SumBytesSent() uint64 {
	return r.data.SumBytesSent()
}
