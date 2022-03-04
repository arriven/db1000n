// MIT License

// Copyright (c) [2022] [Arriven (https://github.com/Arriven)]

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

package metrics

import (
	"context"
	"sync"
	"time"
)

type MetricsStorage struct {
	trackers map[string]*metricTracker
}

type metricTracker struct {
	metrics sync.Map
}

var Default MetricsStorage

func init() {
	Default = MetricsStorage{trackers: make(map[string]*metricTracker)}
	Default.trackers["traffic"] = &metricTracker{}
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
