package metrics

import (
	"context"
	"sync"
	"time"
)

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

func (ms *MetricsStorage) NewWriter(ctx context.Context, name, jobID string) *MetricWriter {
	writer := &MetricWriter{ms: ms, jobID: jobID, name: name, value: 0}
	ticker := time.NewTicker(time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				writer.ms.Write(writer.name, writer.jobID, writer.value)
				writer.value = 0
			}
		}
	}()
	return writer
}

type MetricWriter struct {
	ms    *MetricsStorage
	jobID string
	name  string
	value int
}

func (w *MetricWriter) Add(value int) {
	w.value = w.value + value
}

func (w *MetricWriter) Set(value int) {
	w.value = value
}
