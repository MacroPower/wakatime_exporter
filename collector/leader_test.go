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

func TestNewLeaderCollector(t *testing.T) {
	logger := log.NewNopLogger()
	inputs := CommonInputs{
		BaseURI:   url.URL{Scheme: "https", Host: "api.wakatime.com", Path: "/api/v1"},
		URI:       url.URL{Scheme: "https", Host: "api.wakatime.com", Path: "/api/v1/users/current"},
		Token:     "test-token",
		SSLVerify: true,
		Timeout:   5 * time.Second,
	}

	collector, err := NewLeaderCollector(inputs, logger)
	if err != nil {
		t.Fatalf("Failed to create leader collector: %v", err)
	}

	if collector == nil {
		t.Fatal("Expected collector to be non-nil")
	}
}

func TestLeaderCollectorUpdate(t *testing.T) {
	leaderJSON := `{
		"current_user": {
			"rank": 2886,
			"running_total": {
				"daily_average": 14400,
				"human_readable_daily_average": "4 hrs",
				"human_readable_total": "28 hrs",
				"languages": [
					{"name": "Go", "total_seconds": 72000.0},
					{"name": "Python", "total_seconds": 28800.0}
				],
				"total_seconds": 100800.0
			},
			"user": {
				"full_name": "Test User",
				"id": "user-123",
				"username": "testuser"
			}
		},
		"data": [
			{
				"rank": 1,
				"running_total": {
					"total_seconds": 500000.0
				},
				"user": {
					"username": "topuser"
				}
			}
		],
		"language": null,
		"modified_at": "2025-01-25T12:00:00Z",
		"page": 1,
		"range": {
			"end_date": "2025-01-24",
			"end_text": "yesterday",
			"name": "last_7_days",
			"start_date": "2025-01-18",
			"start_text": "last week",
			"text": "Last 7 Days"
		},
		"timeout": 15,
		"total_pages": 100,
		"writes_only": false
	}`

	logger := log.NewNopLogger()

	mockFetch := func(uri url.URL, subPath string, params url.Values) (io.ReadCloser, error) {
		if subPath != "leaders" {
			t.Errorf("Expected subPath 'leaders', got: %s", subPath)
		}
		return io.NopCloser(strings.NewReader(leaderJSON)), nil
	}

	collector := &leaderCollector{
		rank: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, leaderSubsystem, "rank"),
			"Current rank of the user.",
			nil, nil,
		),
		uri:       url.URL{Scheme: "https", Host: "api.wakatime.com", Path: "/api/v1"},
		fetchStat: mockFetch,
		logger:    logger,
	}

	ch := make(chan prometheus.Metric, 10)

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

	// Expected: 1 rank metric
	expectedMetrics := 1
	if metricCount != expectedMetrics {
		t.Errorf("Expected %d metrics, got %d", expectedMetrics, metricCount)
	}
}

func TestLeaderCollectorUpdateNoCurrentUser(t *testing.T) {
	leaderJSON := `{
		"current_user": null,
		"data": [],
		"page": 1,
		"total_pages": 0
	}`

	logger := log.NewNopLogger()

	mockFetch := func(uri url.URL, subPath string, params url.Values) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(leaderJSON)), nil
	}

	collector := &leaderCollector{
		rank: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, leaderSubsystem, "rank"),
			"Current rank of the user.",
			nil, nil,
		),
		uri:       url.URL{Scheme: "https", Host: "api.wakatime.com", Path: "/api/v1"},
		fetchStat: mockFetch,
		logger:    logger,
	}

	ch := make(chan prometheus.Metric, 10)

	// This may or may not error depending on implementation
	// but should not panic
	_ = collector.Update(ch)

	close(ch)
}

func TestLeaderCollectorUpdateFetchError(t *testing.T) {
	logger := log.NewNopLogger()

	mockFetch := func(uri url.URL, subPath string, params url.Values) (io.ReadCloser, error) {
		return nil, io.ErrUnexpectedEOF
	}

	collector := &leaderCollector{
		rank: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, leaderSubsystem, "rank"),
			"Current rank of the user.",
			nil, nil,
		),
		uri:       url.URL{Scheme: "https", Host: "api.wakatime.com", Path: "/api/v1"},
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

func TestLeaderCollectorUpdateInvalidJSON(t *testing.T) {
	logger := log.NewNopLogger()

	mockFetch := func(uri url.URL, subPath string, params url.Values) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(`{invalid json}`)), nil
	}

	collector := &leaderCollector{
		rank: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, leaderSubsystem, "rank"),
			"Current rank of the user.",
			nil, nil,
		),
		uri:       url.URL{Scheme: "https", Host: "api.wakatime.com", Path: "/api/v1"},
		fetchStat: mockFetch,
		logger:    logger,
	}

	ch := make(chan prometheus.Metric, 10)

	err := collector.Update(ch)
	if err == nil {
		t.Error("Expected JSON parse error but got none")
	}

	close(ch)
}
