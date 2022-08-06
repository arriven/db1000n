package metrics

func NewStatsTracker(metrics *Metrics) *StatsTracker {
	return &StatsTracker{metrics: metrics}
}

// StatsTracker generalizes tracking stats changes between reports
type StatsTracker struct {
	lastStats  PerTargetStats
	lastTotals Stats
	metrics    *Metrics
}

func (st *StatsTracker) sumStats(groupTargets bool) (stats PerTargetStats, totals Stats, statsInterval PerTargetStats, totalsInterval Stats) {
	stats, totals = st.metrics.SumAllStats(groupTargets)
	statsInterval, totalsInterval = stats.Diff(st.lastStats), Diff(totals, st.lastTotals)
	st.lastStats, st.lastTotals = stats, totals

	return
}
