/*
Wakatime Exporter for Prometheus
Copyright (C) 2020 Jacob Colvin (MacroPower)

This program is free software; you can redistribute it and/or
modify it under the terms of the GNU General Public License
as published by the Free Software Foundation; either version 2
of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program; if not, write to the Free Software
Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
*/

package lib

import (
	"io"
	"net/url"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

// Exporter is a struct for all collector exporters
type Exporter struct {
	URI                       *url.URL
	Endpoint, User, Subsystem string
	Mutex                     sync.RWMutex
	FetchStat                 func(url.URL, string, url.Values) (io.ReadCloser, error)

	DefaultMetrics MetricDefaults
	ExportMetric   func(m MetricInfo, ch chan<- prometheus.Metric, value float64, labels ...string)
	Logger         log.Logger
}

// Metrics maps all MetricInfo
type Metrics map[string]MetricInfo

// MetricInfo contains the metric Desc and Type
type MetricInfo struct {
	Desc *prometheus.Desc
	Type prometheus.ValueType
}

// MetricDefaults contains the default metrics exported by each collector
type MetricDefaults struct {
	Up            prometheus.Gauge
	TotalScrapes  prometheus.Counter
	QueryFailures prometheus.Counter
}
