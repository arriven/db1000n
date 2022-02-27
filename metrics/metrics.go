package metrics

import "sync"

type MetricsStorage struct {
	trackers map[string]metricTracker
}

type metricTracker struct {
	metrics sync.Map
}

var Default MetricsStorage

func init() {
	Default = MetricsStorage{trackers: make(map[string]metricTracker)}
	Default.trackers["traffic"] = metricTracker{}
}

func (ms *MetricsStorage) Write(name, jobID string, value int) {
	if tracker, ok := ms.trackers[name]; ok {
		tracker.metrics.Store(jobID, value)
	}
}

func (ms *MetricsStorage) Read(name string) int {
	sum := 0
	if tracker, ok := ms.trackers[name]; ok {
		tracker.metrics.Range(func(k, v interface{}) bool {
			if value, ok := v.(int); ok {
				sum = sum + value
			}
			return true
		})
	}
	return sum
}
