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

// Package metrics collects and reports job metrics.
package metrics

import (
	"bytes"
	"sync"
	"testing"
)

func TestReporter_WriteSummary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		metrics [NumStats]map[dimensions]uint64
		want    string
	}{
		{
			name:    "empty",
			metrics: [NumStats]map[dimensions]uint64{nil, nil, nil, nil},
			want:    "[Error] No traffic generated. If you see this message a lot - contact admins\n",
		},
		{
			name: "zero requests sent",
			metrics: [NumStats]map[dimensions]uint64{
				{
					{"job1", "http://foobar.com"}: 10,
				}, {
					{"job1", "http://foobar.com"}: 0,
				}, {
					{"job1", "http://foobar.com"}: 0,
				}, {
					{"job1", "http://foobar.com"}: 0,
				},
			},
			want: "[Error] No traffic generated. If you see this message a lot - contact admins\n",
		},
		{
			name: "one successful job, one target",
			metrics: [NumStats]map[dimensions]uint64{
				{
					{"job1", "http://foobar.com"}: 100,
					{"job2", "http://foobar.com"}: 10,
				}, {
					{"job1", "http://foobar.com"}: 0,
					{"job2", "http://foobar.com"}: 10,
				}, {
					{"job1", "http://foobar.com"}: 0,
					{"job2", "http://foobar.com"}: 10,
				}, {
					{"job1", "http://foobar.com"}: 0,
					{"job2", "http://foobar.com"}: 1638400,
				},
			},
			want: `

!Атака проводиться успішно! Русскій воєнний корабль іди нахуй!
!Attack is successful! Russian warship, go fuck yourself!

 --- Traffic stats ---
 |            Target | Requests attempted | Requests sent | Responses received | Data sent |
 | http://foobar.com |                110 |            10 |                 10 |   1.56 MB |
 |               --- |                --- |           --- |                --- |       --- |
 |             Total |                110 |            10 |                 10 |   1.56 MB |

`,
		},
		{
			name: "two successful jobs, one target",
			metrics: [NumStats]map[dimensions]uint64{
				{
					{"job1", "http://foobar.com"}: 1000,
					{"job2", "http://foobar.com"}: 100,
				}, {
					{"job1", "http://foobar.com"}: 100,
					{"job2", "http://foobar.com"}: 10,
				}, {
					{"job1", "http://foobar.com"}: 10,
					{"job2", "http://foobar.com"}: 1,
				}, {
					{"job1", "http://foobar.com"}: 16384000,
					{"job2", "http://foobar.com"}: 1638400,
				},
			},
			want: `

!Атака проводиться успішно! Русскій воєнний корабль іди нахуй!
!Attack is successful! Russian warship, go fuck yourself!

 --- Traffic stats ---
 |            Target | Requests attempted | Requests sent | Responses received | Data sent |
 | http://foobar.com |               1100 |           110 |                 11 |  17.19 MB |
 |               --- |                --- |           --- |                --- |       --- |
 |             Total |               1100 |           110 |                 11 |  17.19 MB |

`,
		},
		{
			name: "two successful jobs, two targets",
			metrics: [NumStats]map[dimensions]uint64{
				{
					{"job1", "http://foobar.com"}: 1000,
					{"job2", "http://foobar.com"}: 100,
					{"job1", "tcp://foobar.com"}:  1000,
					{"job2", "tcp://foobar.com"}:  100,
				}, {
					{"job1", "http://foobar.com"}: 100,
					{"job2", "http://foobar.com"}: 10,
					{"job1", "tcp://foobar.com"}:  100,
					{"job2", "tcp://foobar.com"}:  10,
				}, {
					{"job1", "http://foobar.com"}: 10,
					{"job2", "http://foobar.com"}: 1,
					{"job1", "tcp://foobar.com"}:  10,
					{"job2", "tcp://foobar.com"}:  1,
				}, {
					{"job1", "http://foobar.com"}: 16384000,
					{"job2", "http://foobar.com"}: 1638400,
					{"job1", "tcp://foobar.com"}:  1638400,
					{"job2", "tcp://foobar.com"}:  163840,
				},
			},
			want: `

!Атака проводиться успішно! Русскій воєнний корабль іди нахуй!
!Attack is successful! Russian warship, go fuck yourself!

 --- Traffic stats ---
 |            Target | Requests attempted | Requests sent | Responses received | Data sent |
 | http://foobar.com |               1100 |           110 |                 11 |  17.19 MB |
 |  tcp://foobar.com |               1100 |           110 |                 11 |   1.72 MB |
 |               --- |                --- |           --- |                --- |       --- |
 |             Total |               2200 |           220 |                 22 |  18.91 MB |

`,
		},
	}

	for i := range tests {
		tt := tests[i]

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := &Reporter{}

			for s := range tt.metrics {
				r.metrics[s] = sync.Map{}

				for k, v := range tt.metrics[s] {
					r.metrics[s].Store(k, v)
				}
			}

			target := &bytes.Buffer{}
			r.WriteSummary(target)
			if got := target.String(); got != tt.want {
				t.Errorf("Reporter.WriteSummary() = %v, want %v", got, tt.want)
			}
		})
	}
}
