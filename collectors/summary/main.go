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

package summary

import (
	"io"
	"net/url"
	"time"

	exporter "github.com/MacroPower/wakatime_exporter/lib"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "wakatime"
	subsystem = "summary"
	endpoint  = "summaries"
)

var (
	wakaMetrics = exporter.Metrics{
		"total":            exporter.NewWakaMetric("seconds_total", "Total seconds.", prometheus.CounterValue, nil, nil),
		"language":         exporter.NewWakaMetric("language_seconds_total", "Total seconds for each language.", prometheus.CounterValue, []string{"name"}, nil),
		"operating_system": exporter.NewWakaMetric("operating_system_seconds_total", "Total seconds for each operating system.", prometheus.CounterValue, []string{"name"}, nil),
		"machine":          exporter.NewWakaMetric("machine_seconds_total", "Total seconds for each machine.", prometheus.CounterValue, []string{"name", "id"}, nil),
		"editor":           exporter.NewWakaMetric("editor_seconds_total", "Total seconds for each editor.", prometheus.CounterValue, []string{"name"}, nil),
		"project":          exporter.NewWakaMetric("project_seconds_total", "Total seconds for each project.", prometheus.CounterValue, []string{"name"}, nil),
		"category":         exporter.NewWakaMetric("category_seconds_total", "Total seconds for each category.", prometheus.CounterValue, []string{"name"}, nil),
	}
)

// NewExporter creates the Summary exporter
func NewExporter(baseURI *url.URL, user string, token string, sslVerify bool, tzOffset time.Duration, timeout time.Duration, logger log.Logger) *exporterSummary {
	var fetchStat func(url.URL, string, string) (io.ReadCloser, error)
	fetchStat = exporter.FetchHTTP(token, sslVerify, timeout, logger)

	return &exporterSummary{
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
func (e *exporterSummary) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range wakaMetrics {
		ch <- m.Desc
	}

	ch <- e.up.Desc()
	ch <- e.totalScrapes.Desc()
	ch <- e.queryFailures.Desc()
}

// Collect all the metrics.
func (e *exporterSummary) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()

	err := e.scrape(ch)
	up := float64(1)
	if err != nil {
		up = float64(0)
		e.queryFailures.Inc()
		level.Error(e.logger).Log("msg", "Can't scrape wakatime", "subsystem", subsystem, "err", err)
	}
	e.up.Set(up)

	ch <- e.up
	ch <- e.totalScrapes
	ch <- e.queryFailures
}

func (e *exporterSummary) exportMetric(m exporter.MetricInfo, ch chan<- prometheus.Metric, value float64, labels ...string) {
	ch <- prometheus.MustNewConstMetric(m.Desc, m.Type, value, labels...)
}

func (e *exporterSummary) scrape(ch chan<- prometheus.Metric) error {
	level.Debug(e.logger).Log("msg", "Starting scrape")

	e.totalScrapes.Inc()

	dateUTC := exporter.GetDate(e.tzOffset)
	userURL := exporter.UserPath(e.URI, e.user)

	body, fetchErr := e.fetchStat(userURL, dateUTC, endpoint)
	defer body.Close()
	if fetchErr != nil {
		return fetchErr
	}

	summaryStats := wakatimeSummary{}
	err := exporter.ReadAndUnmarshal(body, &summaryStats)
	if err != nil {
		return err
	}

	for i, data := range summaryStats.Data {
		level.Info(e.logger).Log(
			"msg", "Collecting summary from Wakatime",
			"obj", i,
			"start", data.Range.Start.String(),
			"end", data.Range.End.String(),
			"tz", data.Range.Timezone,
			"text", data.Range.Text,
		)
	}

	resultLength := len(summaryStats.Data)
	if resultLength != 1 {
		level.Error(e.logger).Log("msg", "length of results is incorrect", "size", resultLength)
	}
	todaySummaryStats := summaryStats.Data[0]

	e.exportMetric(wakaMetrics["total"], ch, todaySummaryStats.GrandTotal.TotalSeconds)

	for _, lang := range todaySummaryStats.Languages {
		e.exportMetric(wakaMetrics["language"], ch, lang.TotalSeconds, lang.Name)
	}

	for _, os := range todaySummaryStats.OperatingSystems {
		e.exportMetric(wakaMetrics["operating_system"], ch, os.TotalSeconds, os.Name)
	}

	for _, machine := range todaySummaryStats.Machines {
		e.exportMetric(wakaMetrics["machine"], ch, machine.TotalSeconds, machine.Name, machine.MachineNameID)
	}

	for _, editor := range todaySummaryStats.Editors {
		e.exportMetric(wakaMetrics["editor"], ch, editor.TotalSeconds, editor.Name)
	}

	for _, project := range todaySummaryStats.Projects {
		e.exportMetric(wakaMetrics["project"], ch, project.TotalSeconds, project.Name)
	}

	for _, category := range todaySummaryStats.Categories {
		e.exportMetric(wakaMetrics["category"], ch, category.TotalSeconds, category.Name)
	}

	return nil
}
