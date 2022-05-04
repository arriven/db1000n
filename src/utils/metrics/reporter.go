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
	WriteSummary(*Metrics)
}

// ZapReporter

type ZapReporter struct {
	logger *zap.Logger
}

// NewZapReporter creates a new Reporter using a zap logger.
func NewZapReporter(logger *zap.Logger) Reporter {
	return &ZapReporter{logger: logger}
}

func (r *ZapReporter) WriteSummary(metrics *Metrics) {
	stats, totals := metrics.SumAllStats()

	r.logger.Info("stats", zap.Object("total", &totals), zap.Object("targets", stats))
}

// ConsoleReporter

type ConsoleReporter struct {
	target io.Writer
}

// NewConsoleReporter creates a new Reporter which outputs straight to the console
func NewConsoleReporter(target io.Writer) Reporter {
	return &ConsoleReporter{target: target}
}

func (r *ConsoleReporter) WriteSummary(metrics *Metrics) {
	writer := tabwriter.NewWriter(r.target, 1, 1, 1, ' ', tabwriter.AlignRight)
	stats, totals := metrics.SumAllStats()

	defer writer.Flush()

	// Print table's header
	fmt.Fprintln(writer, "\n --- Traffic stats ---")
	fmt.Fprintf(writer, "|\tTarget\t|\tRequests attempted\t|\tRequests sent\t|\tResponses received\t|\tData sent \t|\n")

	// Print all table rows
	for _, tgt := range stats.sortedTargets() {
		printStatsRow(writer, tgt, stats[tgt])
	}

	// Print table's footer
	fmt.Fprintln(writer, "|\t---\t|\t---\t|\t---\t|\t---\t|\t--- \t|")
	printStatsRow(writer, "Total", totals)
	fmt.Fprintln(writer)
}

func printStatsRow(writer *tabwriter.Writer, rowName string, stats Stats) {
	const BytesInMegabyte = 1024 * 1024

	fmt.Fprintf(writer, "|\t%s\t|\t%d\t|\t%d\t|\t%d\t|\t%.2f MB \t|\n", rowName,
		stats[RequestsAttemptedStat],
		stats[RequestsSentStat],
		stats[ResponsesReceivedStat],
		float64(stats[BytesSentStat])/BytesInMegabyte,
	)
}
