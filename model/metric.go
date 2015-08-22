// Copyright 2013 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

var separator = []byte{0}

// A Metric is similar to a LabelSet, but the key difference is that a Metric is
// a singleton and refers to one and only one stream of samples.
type Metric LabelSet

// Equal compares the metrics.
func (m Metric) Equal(o Metric) bool {
	return LabelSet(m).Equal(LabelSet(o))
}

// Before compares the metrics' underlying label sets.
func (m Metric) Before(o Metric) bool {
	return LabelSet(m).Before(LabelSet(o))
}

// Clone returns a copy of the Metric.
func (m Metric) Clone() Metric {
	clone := Metric{}
	for k, v := range m {
		clone[k] = v
	}
	return clone
}

func (m Metric) String() string {
	metricName, hasName := m[MetricNameLabel]
	numLabels := len(m) - 1
	if !hasName {
		numLabels = len(m)
	}
	labelStrings := make([]string, 0, numLabels)
	for label, value := range m {
		if label != MetricNameLabel {
			labelStrings = append(labelStrings, fmt.Sprintf("%s=%q", label, value))
		}
	}

	switch numLabels {
	case 0:
		if hasName {
			return string(metricName)
		}
		return "{}"
	default:
		sort.Strings(labelStrings)
		return fmt.Sprintf("%s{%s}", metricName, strings.Join(labelStrings, ", "))
	}
}

// Fingerprint returns a Metric's Fingerprint.
func (m Metric) Fingerprint() Fingerprint {
	return LabelSet(m).Fingerprint()
}

// FastFingerprint returns a Metric's Fingerprint calculated by a faster hashing
// algorithm, which is, however, more susceptible to hash collisions.
func (m Metric) FastFingerprint() Fingerprint {
	return LabelSet(m).FastFingerprint()
}

// COWMetric wraps a Metric to enable copy-on-write access patterns.
type COWMetric struct {
	Copied bool
	Metric Metric
}

// Set sets a label name in the wrapped Metric to a given value and copies the
// Metric initially, if it is not already a copy.
func (m *COWMetric) Set(ln LabelName, lv LabelValue) {
	m.doCOW()
	m.Metric[ln] = lv
}

// Del deletes a given label name from the wrapped Metric and copies the
// Metric initially, if it is not already a copy.
func (m *COWMetric) Del(ln LabelName) {
	m.doCOW()
	delete(m.Metric, ln)
}

// doCOW copies the underlying Metric if it is not already a copy.
func (m *COWMetric) doCOW() {
	if !m.Copied {
		m.Metric = m.Metric.Clone()
		m.Copied = true
	}
}

// String implements fmt.Stringer.
func (m COWMetric) String() string {
	return m.Metric.String()
}

// MarshalJSON implements json.Marshaler.
func (m COWMetric) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Metric)
}
