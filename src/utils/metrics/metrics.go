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

// Package metrics [everything related to metrics goes here]
package metrics

import (
	"context"
	"sync"
	"time"
)

// supported default metrics
const (
	Traffic          = "traffic"
	ProcessedTraffic = "processed_traffic"
)

// Storage is a general struct to store custom metrics
type Storage struct {
	trackers map[string]*metricTracker // map by metric type
}

type metricTracker struct {
	metrics sync.Map // map by job ID
}

// Default to allow global access for ease of use
// similar to http.DefaultClient and such
var Default = Storage{
	trackers: map[string]*metricTracker{
		Traffic:          {},
		ProcessedTraffic: {},
	},
}

func (ms *Storage) Write(name, jobID string, value uint64) {
	if tracker, ok := ms.trackers[name]; ok {
		tracker.metrics.Store(jobID, value)
	}
}

func (ms *Storage) Read(name string) uint64 {
	var sum uint64

	if tracker, ok := ms.trackers[name]; ok {
		tracker.metrics.Range(func(k, v interface{}) bool {
			if value, ok := v.(uint64); ok {
				sum += value
			}

			return true
		})
	}

	return sum
}

func (ms *Storage) ResetAll() {
	for k := range ms.trackers {
		ms.trackers[k] = &metricTracker{}
	}
}

// NewWriter creates a writer for accumulated writes to the storage
func (ms *Storage) NewWriter(name, jobID string) *Writer {
	return &Writer{ms: ms, jobID: jobID, name: name, value: 0}
}

// Writer is a helper to accumulate writes to a storage on a regular basis
type Writer struct {
	ms    *Storage
	jobID string
	name  string
	value uint64
}

// Add used to increase metric value by a specific amount
func (w *Writer) Add(value uint64) {
	w.value += value
}

// Set used to set metric to a specific value
func (w *Writer) Set(value uint64) {
	w.value = value
}

// Flush used to flush pending metrics updates to the storage
func (w *Writer) Flush() {
	w.ms.Write(w.name, w.jobID, w.value)
}

// Update updates writer with a set uint64erval
func (w *Writer) Update(ctx context.Context, uint64erval time.Duration) {
	ticker := time.NewTicker(uint64erval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.Flush()
		}
	}
}

// NopWriter implements io.Writer interface to simply track how much data has to be serialized
type NopWriter struct{}

func (w NopWriter) Write(p []byte) (n int, _ error) {
	return len(p), nil
}
