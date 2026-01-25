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

func TestNewStatsCollector(t *testing.T) {
	logger := log.NewNopLogger()
	inputs := CommonInputs{
		BaseURI:   url.URL{Scheme: "https", Host: "api.wakatime.com", Path: "/api/v1"},
		URI:       url.URL{Scheme: "https", Host: "api.wakatime.com", Path: "/api/v1/users/current"},
		Token:     "test-token",
		SSLVerify: true,
		Timeout:   5 * time.Second,
	}

	collector, err := NewStatsCollector(inputs, logger)
	if err != nil {
		t.Fatalf("Failed to create stats collector: %v", err)
	}

	if collector == nil {
		t.Fatal("Expected collector to be non-nil")
	}
}

func TestStatsCollectorUpdate(t *testing.T) {
	statsJSON := `{
		"data": {
			"best_day": {
				"date": "2025-01-20",
				"text": "8 hrs",
				"total_seconds": 28800.0
			},
			"categories": [
				{"name": "Coding", "total_seconds": 50000.0, "percent": 80.0},
				{"name": "ai coding", "total_seconds": 12500.0, "percent": 20.0}
			],
			"daily_average": 17857.21,
			"days_including_holidays": 7,
			"days_minus_holidays": 5,
			"editors": [
				{"name": "VS Code", "total_seconds": 62500.0, "percent": 100.0}
			],
			"is_up_to_date": true,
			"languages": [
				{"name": "Go", "total_seconds": 45000.0, "percent": 72.0},
				{"name": "Python", "total_seconds": 17500.0, "percent": 28.0}
			],
			"machines": [
				{"name": "MacBook", "machine_name_id": "abc-123", "total_seconds": 62500.0, "percent": 100.0}
			],
			"operating_systems": [
				{"name": "Mac", "total_seconds": 62500.0, "percent": 100.0}
			],
			"projects": [
				{"name": "wakatime_exporter", "total_seconds": 62500.0, "percent": 100.0}
			],
			"range": "last_7_days",
			"status": "ok",
			"total_seconds": 125000.5
		}
	}`

	logger := log.NewNopLogger()

	mockFetch := func(uri url.URL, subPath string, params url.Values) (io.ReadCloser, error) {
		// Verify the endpoint path contains stats/last_7_days
		if !strings.Contains(subPath, "stats") {
			t.Errorf("Expected subPath to contain 'stats', got: %s", subPath)
		}
		return io.NopCloser(strings.NewReader(statsJSON)), nil
	}

	collector := &statsCollector{
		total: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", "total_seconds"),
			"Total seconds for the stats range.",
			[]string{"range"}, nil,
		),
		dailyAverage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", "daily_average_seconds"),
			"Daily average seconds for the stats range.",
			[]string{"range"}, nil,
		),
		bestDay: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", "best_day_seconds"),
			"Best day total seconds for the stats range.",
			[]string{"range", "date"}, nil,
		),
		language: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", "language_seconds"),
			"Total seconds for each language in the stats range.",
			[]string{"range", "name"}, nil,
		),
		operatingSystem: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", "operating_system_seconds"),
			"Total seconds for each operating system in the stats range.",
			[]string{"range", "name"}, nil,
		),
		machine: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", "machine_seconds"),
			"Total seconds for each machine in the stats range.",
			[]string{"range", "name", "id"}, nil,
		),
		editor: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", "editor_seconds"),
			"Total seconds for each editor in the stats range.",
			[]string{"range", "name"}, nil,
		),
		project: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", "project_seconds"),
			"Total seconds for each project in the stats range.",
			[]string{"range", "name"}, nil,
		),
		category: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", "category_seconds"),
			"Total seconds for each category in the stats range.",
			[]string{"range", "name"}, nil,
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

	// Expected: 1 total + 1 dailyAvg + 1 bestDay + 2 languages + 1 OS + 1 machine + 1 editor + 1 project + 2 categories = 11
	expectedMetrics := 11
	if metricCount != expectedMetrics {
		t.Errorf("Expected %d metrics, got %d", expectedMetrics, metricCount)
	}
}

func TestStatsCollectorUpdateNotUpToDate(t *testing.T) {
	statsJSON := `{
		"data": {
			"is_up_to_date": false,
			"status": "calculating",
			"percent_calculated": 50,
			"total_seconds": 0
		}
	}`

	logger := log.NewNopLogger()

	mockFetch := func(uri url.URL, subPath string, params url.Values) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(statsJSON)), nil
	}

	collector := &statsCollector{
		total: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", "total_seconds"),
			"Total seconds for the stats range.",
			[]string{"range"}, nil,
		),
		dailyAverage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", "daily_average_seconds"),
			"Daily average seconds for the stats range.",
			[]string{"range"}, nil,
		),
		bestDay: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", "best_day_seconds"),
			"Best day total seconds for the stats range.",
			[]string{"range", "date"}, nil,
		),
		language: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", "language_seconds"),
			"Total seconds for each language in the stats range.",
			[]string{"range", "name"}, nil,
		),
		operatingSystem: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", "operating_system_seconds"),
			"Total seconds for each operating system in the stats range.",
			[]string{"range", "name"}, nil,
		),
		machine: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", "machine_seconds"),
			"Total seconds for each machine in the stats range.",
			[]string{"range", "name", "id"}, nil,
		),
		editor: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", "editor_seconds"),
			"Total seconds for each editor in the stats range.",
			[]string{"range", "name"}, nil,
		),
		project: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", "project_seconds"),
			"Total seconds for each project in the stats range.",
			[]string{"range", "name"}, nil,
		),
		category: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", "category_seconds"),
			"Total seconds for each category in the stats range.",
			[]string{"range", "name"}, nil,
		),
		uri:       url.URL{Scheme: "https", Host: "api.wakatime.com", Path: "/api/v1/users/current"},
		fetchStat: mockFetch,
		logger:    logger,
	}

	ch := make(chan prometheus.Metric, 20)

	// Should still work but log a warning
	err := collector.Update(ch)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	close(ch)
}

func TestStatsCollectorUpdateFetchError(t *testing.T) {
	logger := log.NewNopLogger()

	mockFetch := func(uri url.URL, subPath string, params url.Values) (io.ReadCloser, error) {
		return nil, io.ErrUnexpectedEOF
	}

	collector := &statsCollector{
		total: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", "total_seconds"),
			"Total seconds for the stats range.",
			[]string{"range"}, nil,
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
