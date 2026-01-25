/*
Copyright 2020 Jacob Colvin (MacroPower)
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package collector

import (
	"io"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

func TestNewSummaryCollector(t *testing.T) {
	logger := log.NewNopLogger()
	inputs := CommonInputs{
		BaseURI:   url.URL{Scheme: "https", Host: "api.wakatime.com", Path: "/api/v1"},
		URI:       url.URL{Scheme: "https", Host: "api.wakatime.com", Path: "/api/v1/users/current"},
		Token:     "test-token",
		SSLVerify: true,
		Timeout:   5 * time.Second,
	}

	collector, err := NewSummaryCollector(inputs, logger)
	if err != nil {
		t.Fatalf("Failed to create summary collector: %v", err)
	}

	if collector == nil {
		t.Fatal("Expected collector to be non-nil")
	}
}

func TestSummaryCollectorUpdate(t *testing.T) {
	summaryJSON := `{
		"data": [{
			"categories": [
				{"name": "Coding", "total_seconds": 3600.5, "percent": 80.0},
				{"name": "Debugging", "total_seconds": 900.25, "percent": 20.0}
			],
			"dependencies": [],
			"editors": [
				{"name": "VS Code", "total_seconds": 4500.75, "percent": 100.0}
			],
			"grand_total": {
				"digital": "1:15:00",
				"hours": 1,
				"minutes": 15,
				"text": "1 hr 15 mins",
				"total_seconds": 4500.75
			},
			"languages": [
				{"name": "Go", "total_seconds": 3000.0, "percent": 66.7},
				{"name": "Python", "total_seconds": 1500.75, "percent": 33.3}
			],
			"machines": [
				{"name": "MacBook", "machine_name_id": "abc-123", "total_seconds": 4500.75, "percent": 100.0}
			],
			"operating_systems": [
				{"name": "Mac", "total_seconds": 4500.75, "percent": 100.0}
			],
			"projects": [
				{"name": "wakatime_exporter", "total_seconds": 4500.75, "percent": 100.0}
			],
			"range": {
				"date": "2025-01-25",
				"end": "2025-01-25T23:59:59Z",
				"start": "2025-01-25T00:00:00Z",
				"text": "Today",
				"timezone": "America/New_York"
			}
		}],
		"end": "2025-01-25T23:59:59Z",
		"start": "2025-01-25T00:00:00Z"
	}`

	logger := log.NewNopLogger()

	// Create a mock fetch function
	mockFetch := func(uri url.URL, subPath string, params url.Values) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(summaryJSON)), nil
	}

	collector := &summaryCollector{
		total: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", summaryMetricName),
			"Total seconds.",
			nil, nil,
		),
		language: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "language", summaryMetricName),
			"Total seconds for each language.",
			[]string{"name"}, nil,
		),
		operatingSystem: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "operating_system", summaryMetricName),
			"Total seconds for each operating system.",
			[]string{"name"}, nil,
		),
		machine: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "machine", summaryMetricName),
			"Total seconds for each machine.",
			[]string{"name", "id"}, nil,
		),
		editor: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "editor", summaryMetricName),
			"Total seconds for each editor.",
			[]string{"name"}, nil,
		),
		project: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "project", summaryMetricName),
			"Total seconds for each project.",
			[]string{"name"}, nil,
		),
		category: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "category", summaryMetricName),
			"Total seconds for each category.",
			[]string{"name"}, nil,
		),
		uri:       url.URL{Scheme: "https", Host: "api.wakatime.com", Path: "/api/v1/users/current"},
		fetchStat: mockFetch,
		logger:    logger,
	}

	ch := make(chan prometheus.Metric, 20)

	err := collector.Update(ch)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	close(ch)

	// Count metrics
	metricCount := 0
	for range ch {
		metricCount++
	}

	// Expected: 1 total + 2 languages + 1 OS + 1 machine + 1 editor + 1 project + 2 categories = 9
	expectedMetrics := 9
	if metricCount != expectedMetrics {
		t.Errorf("Expected %d metrics, got %d", expectedMetrics, metricCount)
	}
}

func TestSummaryCollectorUpdateFetchError(t *testing.T) {
	logger := log.NewNopLogger()

	// Create a mock fetch function that returns an error
	mockFetch := func(uri url.URL, subPath string, params url.Values) (io.ReadCloser, error) {
		return nil, io.ErrUnexpectedEOF
	}

	collector := &summaryCollector{
		total: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", summaryMetricName),
			"Total seconds.",
			nil, nil,
		),
		uri:       url.URL{Scheme: "https", Host: "api.wakatime.com", Path: "/api/v1/users/current"},
		fetchStat: mockFetch,
		logger:    logger,
	}

	ch := make(chan prometheus.Metric, 10)

	err := collector.Update(ch)
	if err == nil {
		t.Error("Expected error but got none")
	}

	close(ch)
}
