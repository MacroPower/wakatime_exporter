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
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
)

func TestBoolToBinary(t *testing.T) {
	tests := []struct {
		input    bool
		expected string
	}{
		{true, "1"},
		{false, "0"},
	}

	for _, tt := range tests {
		result := BoolToBinary(tt.input)
		if result != tt.expected {
			t.Errorf("BoolToBinary(%v) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}

func TestFetchHTTP(t *testing.T) {
	tests := []struct {
		name           string
		responseCode   int
		responseBody   string
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:         "successful request",
			responseCode: http.StatusOK,
			responseBody: `{"data": "test"}`,
			expectError:  false,
		},
		{
			name:           "server error",
			responseCode:   http.StatusInternalServerError,
			responseBody:   `{"error": "server error"}`,
			expectError:    true,
			expectedErrMsg: "HTTP status 500",
		},
		{
			name:           "unauthorized",
			responseCode:   http.StatusUnauthorized,
			responseBody:   `{"error": "unauthorized"}`,
			expectError:    true,
			expectedErrMsg: "HTTP status 401",
		},
		{
			name:           "forbidden",
			responseCode:   http.StatusForbidden,
			responseBody:   `{"error": "forbidden"}`,
			expectError:    true,
			expectedErrMsg: "HTTP status 403",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify authorization header is present
				auth := r.Header.Get("Authorization")
				if auth == "" {
					t.Error("Authorization header missing")
				}
				if !strings.HasPrefix(auth, "Basic ") {
					t.Errorf("Authorization header should start with 'Basic ', got: %s", auth)
				}

				w.WriteHeader(tt.responseCode)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			logger := log.NewNopLogger()
			fetchFunc := FetchHTTP("test-token", true, 5*time.Second, logger)

			serverURL, _ := url.Parse(server.URL)
			params := url.Values{}
			params.Add("cache", "false")

			body, err := fetchFunc(*serverURL, "/test", params)

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
				if body != nil {
					_ = body.Close()
				}
			}
		})
	}
}

func TestReadAndUnmarshal(t *testing.T) {
	tests := []struct {
		name        string
		jsonInput   string
		expectError bool
	}{
		{
			name:        "valid JSON",
			jsonInput:   `{"data": {"is_up_to_date": true, "total_seconds": 12345.67}}`,
			expectError: false,
		},
		{
			name:        "invalid JSON",
			jsonInput:   `{invalid json}`,
			expectError: true,
		},
		{
			name:        "empty JSON",
			jsonInput:   `{}`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := io.NopCloser(strings.NewReader(tt.jsonInput))

			var result wakatimeAlltime
			err := ReadAndUnmarshal(body, &result)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestFetchHTTPTimeout(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": "test"}`))
	}))
	defer server.Close()

	logger := log.NewNopLogger()
	// Use a very short timeout
	fetchFunc := FetchHTTP("test-token", true, 50*time.Millisecond, logger)

	serverURL, _ := url.Parse(server.URL)
	params := url.Values{}

	_, err := fetchFunc(*serverURL, "/test", params)

	if err == nil {
		t.Error("Expected timeout error but got none")
	}
}
