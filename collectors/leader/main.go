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

package leader

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/url"
	"time"

	exporter "github.com/MacroPower/wakatime_exporter/lib"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "wakatime"
	subsystem = "leader"
	endpoint  = "leaders"
)

var (
	wakaMetrics = exporter.Metrics{
		"rank": exporter.NewWakaMetric("rank", "Current rank of the user.", prometheus.GaugeValue, nil, nil),
	}
)

// NewExporter creates the Leader exporter
func NewExporter(baseURI *url.URL, user string, token string, sslVerify bool, tzOffset time.Duration, timeout time.Duration, logger log.Logger) *exporterLeader {
	var fetchStat func(url.URL, string, string) (io.ReadCloser, error)
	fetchStat = exporter.FetchHTTP(token, sslVerify, timeout, logger)

	return &exporterLeader{
		URI:       baseURI,
		endpoint:  endpoint,
		fetchStat: fetchStat,
		tzOffset:  tzOffset,
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "up",
			Help:      "Was the last scrape of wakatime successful.",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "exporter_scrapes_total",
			Help:      "Current total wakatime scrapes.",
		}),
		queryFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "exporter_query_failures_total",
			Help:      "Number of errors.",
		}),
		logger: logger,
	}
}

// Describe describes all the metrics ever exported by the wakatime exporter. It
// implements prometheus.Collector.
func (e *exporterLeader) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range wakaMetrics {
		ch <- m.Desc
	}

	ch <- e.up.Desc()
	ch <- e.totalScrapes.Desc()
	ch <- e.queryFailures.Desc()
}

// Collect all the metrics.
func (e *exporterLeader) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()

	err := e.scrape(ch)
	up := float64(1)
	if err != nil {
		up = float64(0)
		e.queryFailures.Inc()
		level.Error(e.logger).Log("msg", "Can't scrape wakatime", "err", err)
	}
	e.up.Set(up)

	ch <- e.up
	ch <- e.totalScrapes
	ch <- e.queryFailures
}

func (e *exporterLeader) exportMetric(m exporter.MetricInfo, ch chan<- prometheus.Metric, value float64, labels ...string) {
	ch <- prometheus.MustNewConstMetric(m.Desc, m.Type, value, labels...)
}

func (e *exporterLeader) scrape(ch chan<- prometheus.Metric) error {
	level.Debug(e.logger).Log("msg", "Starting scrape")

	e.totalScrapes.Inc()

	dateUTC := time.Now().UTC().Add(e.tzOffset).Format("2006-01-02")

	body, fetchErr := e.fetchStat(*e.URI, dateUTC, endpoint)
	if fetchErr != nil {
		return fetchErr
	}

	respBody, readErr := ioutil.ReadAll(body)
	if readErr != nil {
		return readErr
	}

	var closeErr error
	closeErr = body.Close()
	if closeErr != nil {
		return closeErr
	}

	var jsonErr error
	leaderStats := wakatimeLeader{}
	jsonErr = json.Unmarshal(respBody, &leaderStats)
	if jsonErr != nil {
		return jsonErr
	}

	level.Info(e.logger).Log(
		"msg", "Collecting rank from Wakatime",
		"page", leaderStats.Page,
		"updated", leaderStats.ModifiedAt,
	)
	currentRank := float64(leaderStats.CurrentUser.Rank)
	e.exportMetric(wakaMetrics["rank"], ch, currentRank)

	return nil
}
