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
	"net/url"
	"time"

	exporter "github.com/MacroPower/wakatime_exporter/lib"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
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

// Exporter is the local definition of Exporter
type Exporter exporter.Exporter

// NewExporter creates the Summary exporter
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

	params := url.Values{}
	params.Add("start", "today")
	params.Add("end", "today")
	params.Add("cache", "false")

	userURL := exporter.UserPath(e.URI, e.User)

	body, fetchErr := e.FetchStat(userURL, endpoint, params)
	if fetchErr != nil {
		return fetchErr
	}

	summaryStats := wakatimeSummary{}
	err := exporter.ReadAndUnmarshal(body, &summaryStats)
	if err != nil {
		return err
	}

	defer body.Close()

	for i, data := range summaryStats.Data {
		level.Info(e.Logger).Log(
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
		level.Error(e.Logger).Log("msg", "length of results is incorrect", "size", resultLength)
	}
	todaySummaryStats := summaryStats.Data[0]

	e.ExportMetric(wakaMetrics["total"], ch, todaySummaryStats.GrandTotal.TotalSeconds)

	for _, lang := range todaySummaryStats.Languages {
		e.ExportMetric(wakaMetrics["language"], ch, lang.TotalSeconds, lang.Name)
	}

	for _, os := range todaySummaryStats.OperatingSystems {
		e.ExportMetric(wakaMetrics["operating_system"], ch, os.TotalSeconds, os.Name)
	}

	for _, machine := range todaySummaryStats.Machines {
		e.ExportMetric(wakaMetrics["machine"], ch, machine.TotalSeconds, machine.Name, machine.MachineNameID)
	}

	for _, editor := range todaySummaryStats.Editors {
		e.ExportMetric(wakaMetrics["editor"], ch, editor.TotalSeconds, editor.Name)
	}

	for _, project := range todaySummaryStats.Projects {
		e.ExportMetric(wakaMetrics["project"], ch, project.TotalSeconds, project.Name)
	}

	for _, category := range todaySummaryStats.Categories {
		e.ExportMetric(wakaMetrics["category"], ch, category.TotalSeconds, category.Name)
	}

	return nil
}
