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

package goal

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/url"
	"path"
	"time"

	exporter "github.com/MacroPower/wakatime_exporter/lib"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "wakatime"
	subsystem = "goal"
	endpoint  = "goals"
)

var (
	wakaMetrics = exporter.Metrics{
		"goal":          exporter.NewWakaMetric("goal_seconds", "The goal.", prometheus.GaugeValue, []string{"name", "id", "type", "delta"}, nil),
		"goal_progress": exporter.NewWakaMetric("goal_progress_seconds_total", "Progress towards the goal.", prometheus.CounterValue, []string{"name", "id", "type", "delta"}, nil),
		"goal_info":     exporter.NewWakaMetric("goal_info", "Information about the goal.", prometheus.GaugeValue, []string{"name", "id", "ignore_zero_days", "is_enabled", "is_inverse", "is_snoozed", "is_tweeting"}, nil),
	}
)

// NewExporter creates the Summary exporter
func NewExporter(baseURI *url.URL, user string, token string, sslVerify bool, tzOffset time.Duration, timeout time.Duration, logger log.Logger) *exporterGoal {
	var fetchStat func(url.URL, string, string) (io.ReadCloser, error)
	fetchStat = exporter.FetchHTTP(token, sslVerify, timeout, logger)

	return &exporterGoal{
		URI:       baseURI,
		endpoint:  endpoint,
		user:      user,
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
func (e *exporterGoal) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range wakaMetrics {
		ch <- m.Desc
	}

	ch <- e.up.Desc()
	ch <- e.totalScrapes.Desc()
	ch <- e.queryFailures.Desc()
}

// Collect all the metrics.
func (e *exporterGoal) Collect(ch chan<- prometheus.Metric) {
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

func (e *exporterGoal) exportMetric(m exporter.MetricInfo, ch chan<- prometheus.Metric, value float64, labels ...string) {
	ch <- prometheus.MustNewConstMetric(m.Desc, m.Type, value, labels...)
}

func (e *exporterGoal) scrape(ch chan<- prometheus.Metric) error {
	level.Debug(e.logger).Log("msg", "Starting scrape")

	e.totalScrapes.Inc()

	dateUTC := time.Now().UTC().Add(e.tzOffset).Format("2006-01-02")
	userPath := path.Join(e.URI.Path, "users", e.user)
	userURL := *e.URI
	userURL.Path = userPath

	body, fetchErr := e.fetchStat(userURL, dateUTC, endpoint)
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
	goalStats := wakatimeGoal{}
	jsonErr = json.Unmarshal(respBody, &goalStats)
	if jsonErr != nil {
		return jsonErr
	}

	level.Info(e.logger).Log(
		"msg", "Collecting goals from Wakatime",
		"total", goalStats.Total,
		"pages", goalStats.TotalPages,
	)
	for _, data := range goalStats.Data {
		// the last element should be the most recent data
		currentChartData := data.ChartData[len(data.ChartData)-1]

		e.exportMetric(
			wakaMetrics["goal_progress"], ch, currentChartData.ActualSeconds,
			data.Title, data.ID, data.Type, data.Delta,
		)
		e.exportMetric(
			wakaMetrics["goal"], ch, float64(data.Seconds),
			data.Title, data.ID, data.Type, data.Delta,
		)
		e.exportMetric(
			wakaMetrics["goal_info"], ch, 1,
			data.Title, data.ID,
			exporter.BoolToBinary(data.IgnoreZeroDays),
			exporter.BoolToBinary(data.IsEnabled),
			exporter.BoolToBinary(data.IsInverse),
			exporter.BoolToBinary(data.IsSnoozed),
			exporter.BoolToBinary(data.IsTweeting),
		)
	}

	return nil
}
