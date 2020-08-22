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
	"net/url"
	"time"

	exporter "github.com/MacroPower/wakatime_exporter/lib"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
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

// Exporter is the local definition of Exporter
type Exporter exporter.Exporter

// NewExporter creates the Goal exporter
func NewExporter(baseURI *url.URL, user string, token string, sslVerify bool, timeout time.Duration, logger log.Logger) *Exporter {
	fetchStat := exporter.FetchHTTP(token, sslVerify, timeout, logger)
	defaultMetrics := exporter.DefaultMetrics(subsystem)

	return &Exporter{
		URI:            baseURI,
		Endpoint:       endpoint,
		Subsystem:      subsystem,
		User:           user,
		FetchStat:      fetchStat,
		DefaultMetrics: defaultMetrics,
		ExportMetric:   exporter.ExportMetric,
		Logger:         logger,
	}
}

// Describe describes all the metrics ever exported by the wakatime exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range wakaMetrics {
		ch <- m.Desc
	}

	ch <- e.DefaultMetrics.Up.Desc()
	ch <- e.DefaultMetrics.TotalScrapes.Desc()
	ch <- e.DefaultMetrics.QueryFailures.Desc()
}

// Collect all the metrics.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.Mutex.Lock() // To protect metrics from concurrent collects.
	defer e.Mutex.Unlock()

	err := e.scrape(ch)
	up := float64(1)
	if err != nil {
		up = float64(0)
		e.DefaultMetrics.QueryFailures.Inc()
		level.Error(e.Logger).Log("msg", "Can't scrape wakatime", "subsystem", subsystem, "err", err)
	}
	e.DefaultMetrics.Up.Set(up)

	ch <- e.DefaultMetrics.Up
	ch <- e.DefaultMetrics.TotalScrapes
	ch <- e.DefaultMetrics.QueryFailures
}

func (e *Exporter) scrape(ch chan<- prometheus.Metric) error {
	level.Debug(e.Logger).Log("msg", "Starting scrape")

	e.DefaultMetrics.TotalScrapes.Inc()

	userURL := exporter.UserPath(e.URI, e.User)

	params := url.Values{}
	params.Add("cache", "false")

	body, fetchErr := e.FetchStat(userURL, endpoint, params)
	if fetchErr != nil {
		return fetchErr
	}

	goalStats := wakatimeGoal{}
	err := exporter.ReadAndUnmarshal(body, &goalStats)
	if err != nil {
		return err
	}

	defer body.Close()

	level.Info(e.Logger).Log(
		"msg", "Collecting goals from Wakatime",
		"total", goalStats.Total,
		"pages", goalStats.TotalPages,
	)
	for i, data := range goalStats.Data {
		// the last element should be the most recent data
		currentChartData := data.ChartData[len(data.ChartData)-1]

		level.Info(e.Logger).Log(
			"msg", "Collecting goal from Wakatime",
			"obj", i,
			"start", currentChartData.Range.Start,
			"end", currentChartData.Range.End,
			"text", currentChartData.Range.Text,
		)

		e.ExportMetric(
			wakaMetrics["goal_progress"], ch, currentChartData.ActualSeconds,
			data.Title, data.ID, data.Type, data.Delta,
		)
		e.ExportMetric(
			wakaMetrics["goal"], ch, float64(currentChartData.GoalSeconds),
			data.Title, data.ID, data.Type, data.Delta,
		)
		e.ExportMetric(
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
