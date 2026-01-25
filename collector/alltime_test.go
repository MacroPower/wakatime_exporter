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

func TestNewAllTimeCollector(t *testing.T) {
	logger := log.NewNopLogger()
	inputs := CommonInputs{
		BaseURI:   url.URL{Scheme: "https", Host: "api.wakatime.com", Path: "/api/v1"},
		URI:       url.URL{Scheme: "https", Host: "api.wakatime.com", Path: "/api/v1/users/current"},
		Token:     "test-token",
		SSLVerify: true,
		Timeout:   5 * time.Second,
	}

	collector, err := NewAllTimeCollector(inputs, logger)
	if err != nil {
		t.Fatalf("Failed to create all-time collector: %v", err)
	}

	if collector == nil {
		t.Fatal("Expected collector to be non-nil")
	}
}

func TestAllTimeCollectorUpdate(t *testing.T) {
	tests := []struct {
		name           string
		jsonResponse   string
		expectError    bool
		expectedErrMsg string
		expectedMetric float64
	}{
		{
			name: "up to date data",
			jsonResponse: `{
				"data": {
					"is_up_to_date": true,
					"text": "100 hrs",
					"total_seconds": 360000.5
				}
			}`,
			expectError:    false,
			expectedMetric: 360000.5,
		},
		{
			name: "not up to date data",
			jsonResponse: `{
				"data": {
					"is_up_to_date": false,
					"text": "100 hrs",
					"total_seconds": 360000.5
				}
			}`,
			expectError:    true,
			expectedErrMsg: "not up to date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.NewNopLogger()

			mockFetch := func(uri url.URL, subPath string, params url.Values) (io.ReadCloser, error) {
				return io.NopCloser(strings.NewReader(tt.jsonResponse)), nil
			}

			collector := &alltimeCollector{
				total: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, allTimeSubsystem, "seconds_total"),
					"Total seconds (all time).",
					nil, nil,
				),
				uri:       url.URL{Scheme: "https", Host: "api.wakatime.com", Path: "/api/v1/users/current"},
				fetchStat: mockFetch,
				logger:    logger,
			}

			ch := make(chan prometheus.Metric, 10)

			err := collector.Update(ch)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.expectedErrMsg) {
					t.Errorf("Expected error containing '%s', got: %s", tt.expectedErrMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			close(ch)

			if !tt.expectError {
				metricCount := 0
				for range ch {
					metricCount++
				}
				if metricCount != 1 {
					t.Errorf("Expected 1 metric, got %d", metricCount)
				}
			}
		})
	}
}

func TestAllTimeCollectorUpdateFetchError(t *testing.T) {
	logger := log.NewNopLogger()

	mockFetch := func(uri url.URL, subPath string, params url.Values) (io.ReadCloser, error) {
		return nil, io.ErrUnexpectedEOF
	}

	collector := &alltimeCollector{
		total: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, allTimeSubsystem, "seconds_total"),
			"Total seconds (all time).",
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

func TestAllTimeCollectorUpdateInvalidJSON(t *testing.T) {
	logger := log.NewNopLogger()

	mockFetch := func(uri url.URL, subPath string, params url.Values) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(`{invalid json}`)), nil
	}

	collector := &alltimeCollector{
		total: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, allTimeSubsystem, "seconds_total"),
			"Total seconds (all time).",
			nil, nil,
		),
		uri:       url.URL{Scheme: "https", Host: "api.wakatime.com", Path: "/api/v1/users/current"},
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
