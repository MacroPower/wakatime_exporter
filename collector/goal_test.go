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

func TestNewGoalCollector(t *testing.T) {
	logger := log.NewNopLogger()
	inputs := CommonInputs{
		BaseURI:   url.URL{Scheme: "https", Host: "api.wakatime.com", Path: "/api/v1"},
		URI:       url.URL{Scheme: "https", Host: "api.wakatime.com", Path: "/api/v1/users/current"},
		Token:     "test-token",
		SSLVerify: true,
		Timeout:   5 * time.Second,
	}

	collector, err := NewGoalCollector(inputs, logger)
	if err != nil {
		t.Fatalf("Failed to create goal collector: %v", err)
	}

	if collector == nil {
		t.Fatal("Expected collector to be non-nil")
	}
}

func TestGoalCollectorUpdate(t *testing.T) {
	goalJSON := `{
		"data": [
			{
				"id": "goal-123",
				"title": "Code 1 hr per day",
				"type": "coding",
				"delta": "day",
				"is_enabled": true,
				"ignore_zero_days": true,
				"is_inverse": false,
				"is_snoozed": false,
				"is_tweeting": false,
				"chart_data": [
					{
						"goal_seconds": 3600,
						"actual_seconds": 2500.5,
						"range": {
							"start": "2025-01-25T00:00:00Z",
							"end": "2025-01-25T23:59:59Z",
							"text": "Today"
						}
					}
				]
			},
			{
				"id": "goal-456",
				"title": "Code 6 hrs per week",
				"type": "coding",
				"delta": "week",
				"is_enabled": true,
				"ignore_zero_days": false,
				"is_inverse": false,
				"is_snoozed": false,
				"is_tweeting": true,
				"chart_data": [
					{
						"goal_seconds": 21600,
						"actual_seconds": 18000.0,
						"range": {
							"start": "2025-01-20T00:00:00Z",
							"end": "2025-01-26T23:59:59Z",
							"text": "This Week"
						}
					}
				]
			}
		],
		"total": 2,
		"total_pages": 1
	}`

	logger := log.NewNopLogger()

	mockFetch := func(uri url.URL, subPath string, params url.Values) (io.ReadCloser, error) {
		if subPath != "goals" {
			t.Errorf("Expected subPath 'goals', got: %s", subPath)
		}
		return io.NopCloser(strings.NewReader(goalJSON)), nil
	}

	collector := &goalCollector{
		goalThreshold: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, goalSubsystem, "threshold_seconds"),
			"The goal as set through the wakatime interface.",
			[]string{
				"name", "id", "type", "delta", "enabled",
				"ignore_zero_days", "inverse", "snoozed", "tweeting",
			},
			nil,
		),
		goalProgress: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, goalSubsystem, "progress_seconds"),
			"Progress towards the goal.",
			[]string{
				"name", "id", "type", "delta", "enabled",
				"ignore_zero_days", "inverse", "snoozed", "tweeting",
			},
			nil,
		),
		uri:       url.URL{Scheme: "https", Host: "api.wakatime.com", Path: "/api/v1/users/current"},
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

	// Expected: 2 goals * 2 metrics (threshold + progress) = 4
	expectedMetrics := 4
	if metricCount != expectedMetrics {
		t.Errorf("Expected %d metrics, got %d", expectedMetrics, metricCount)
	}
}

func TestGoalCollectorUpdateNoGoals(t *testing.T) {
	goalJSON := `{
		"data": [],
		"total": 0,
		"total_pages": 0
	}`

	logger := log.NewNopLogger()

	mockFetch := func(uri url.URL, subPath string, params url.Values) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(goalJSON)), nil
	}

	collector := &goalCollector{
		goalThreshold: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, goalSubsystem, "threshold_seconds"),
			"The goal as set through the wakatime interface.",
			[]string{
				"name", "id", "type", "delta", "enabled",
				"ignore_zero_days", "inverse", "snoozed", "tweeting",
			},
			nil,
		),
		goalProgress: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, goalSubsystem, "progress_seconds"),
			"Progress towards the goal.",
			[]string{
				"name", "id", "type", "delta", "enabled",
				"ignore_zero_days", "inverse", "snoozed", "tweeting",
			},
			nil,
		),
		uri:       url.URL{Scheme: "https", Host: "api.wakatime.com", Path: "/api/v1/users/current"},
		fetchStat: mockFetch,
		logger:    logger,
	}

	ch := make(chan prometheus.Metric, 10)

	err := collector.Update(ch)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	close(ch)

	// Should have no metrics when no goals
	metricCount := 0
	for range ch {
		metricCount++
	}

	if metricCount != 0 {
		t.Errorf("Expected 0 metrics, got %d", metricCount)
	}
}

func TestGoalCollectorUpdateFetchError(t *testing.T) {
	logger := log.NewNopLogger()

	mockFetch := func(uri url.URL, subPath string, params url.Values) (io.ReadCloser, error) {
		return nil, io.ErrUnexpectedEOF
	}

	collector := &goalCollector{
		goalThreshold: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, goalSubsystem, "threshold_seconds"),
			"The goal as set through the wakatime interface.",
			[]string{
				"name", "id", "type", "delta", "enabled",
				"ignore_zero_days", "inverse", "snoozed", "tweeting",
			},
			nil,
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
