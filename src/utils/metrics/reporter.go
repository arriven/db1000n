package metrics

import (
	"bufio"
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
	logger       *zap.Logger
	groupTargets bool
	tracker      statsTracker
}

// NewZapReporter creates a new Reporter using a zap logger.
func NewZapReporter(logger *zap.Logger, groupTargets bool) Reporter {
	return &ZapReporter{logger: logger, groupTargets: groupTargets}
}

func (r *ZapReporter) WriteSummary(metrics *Metrics) {
	stats, totals, statsInterval, totalsInterval := r.tracker.sumStats(metrics, r.groupTargets)

	r.logger.Info("stats", zap.Object("total", &totals), zap.Object("targets", stats),
		zap.Object("total_since_last_report", &totalsInterval), zap.Object("targets_since_last_report", statsInterval))
}

// ConsoleReporter

type ConsoleReporter struct {
	target       *bufio.Writer
	groupTargets bool
	tracker      statsTracker
}

// NewConsoleReporter creates a new Reporter which outputs straight to the console
func NewConsoleReporter(target io.Writer, groupTargets bool) Reporter {
	return &ConsoleReporter{target: bufio.NewWriter(target), groupTargets: groupTargets}
}

func (r *ConsoleReporter) WriteSummary(metrics *Metrics) {
	writer := tabwriter.NewWriter(r.target, 1, 1, 1, ' ', tabwriter.AlignRight)

	r.writeSummaryTo(metrics, writer)

	// Important to flush the remains of bufio.Writer
	r.target.Flush()
}

func (r *ConsoleReporter) writeSummaryTo(metrics *Metrics, writer *tabwriter.Writer) {
	stats, totals, statsInterval, totalsInterval := r.tracker.sumStats(metrics, r.groupTargets)

	defer writer.Flush()

	// Print table's header
	fmt.Fprintln(writer, "\n --- Traffic stats ---")
	fmt.Fprintf(writer, "|\tTarget\t|\tRequests attempted\t|\tRequests sent\t|\tResponses received\t|\tData sent\t|\tData received \t|\n")

	// Print all table rows
	for _, tgt := range stats.sortedTargets() {
		printStatsRow(writer, tgt, stats[tgt], statsInterval[tgt])
	}

	// Print table's footer
	fmt.Fprintln(writer, "|\t---\t|\t---\t|\t---\t|\t---\t|\t---\t|\t--- \t|")
	printStatsRow(writer, "Total", totals, totalsInterval)
	fmt.Fprintln(writer)
}

func printStatsRow(writer *tabwriter.Writer, rowName string, stats Stats, diff Stats) {
	const BytesInMegabyte = 1024 * 1024

	fmt.Fprintf(writer, "|\t%s\t|\t%d/%d\t|\t%d/%d\t|\t%d/%d\t|\t%.2f MB/%.2f MB\t|\t%.2f MB/%.2f MB \t|\n", rowName,
		diff[RequestsAttemptedStat], stats[RequestsAttemptedStat],
		diff[RequestsSentStat], stats[RequestsSentStat],
		diff[ResponsesReceivedStat], stats[ResponsesReceivedStat],
		float64(diff[BytesSentStat])/BytesInMegabyte, float64(stats[BytesSentStat])/BytesInMegabyte,
		float64(diff[BytesReceivedStat])/BytesInMegabyte, float64(stats[BytesReceivedStat])/BytesInMegabyte,
	)
}

// statsTracker generalizes tracking stats changes between reports
type statsTracker struct {
	lastStats  PerTargetStats
	lastTotals Stats
}

func (st *statsTracker) sumStats(metrics *Metrics, groupTargets bool) (stats PerTargetStats, totals Stats, statsInterval PerTargetStats, totalsInterval Stats) {
	stats, totals = metrics.SumAllStats(groupTargets)
	statsInterval, totalsInterval = stats.Diff(st.lastStats), Diff(totals, st.lastTotals)
	st.lastStats, st.lastTotals = stats, totals

	return
}
