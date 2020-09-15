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

package collector

import (
	"io"
	"net/url"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	summaryCollectorName = "summary"
	summarySubsystem     = "summary"
	summaryEndpoint      = "summaries"
)

type summaryCollector struct {
	total           *prometheus.Desc
	language        *prometheus.Desc
	operatingSystem *prometheus.Desc
	machine         *prometheus.Desc
	editor          *prometheus.Desc
	project         *prometheus.Desc
	category        *prometheus.Desc
	uri             url.URL
	fetchStat       func(url.URL, string, url.Values) (io.ReadCloser, error)
	logger          log.Logger
}

func init() {
	registerCollector(summaryCollectorName, defaultEnabled, NewSummaryCollector)
}

// NewSummaryCollector returns a new Collector exposing all-time stats.
func NewSummaryCollector(in CommonInputs, logger log.Logger) (Collector, error) {
	return &summaryCollector{
		total: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, summarySubsystem, "seconds_total"),
			"Total seconds.",
			nil, nil,
		),
		language: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, summarySubsystem, "language_seconds_total"),
			"Total seconds for each language.",
			[]string{"name"}, nil,
		),
		operatingSystem: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, summarySubsystem, "operating_system_seconds_total"),
			"Total seconds for each operating system.",
			[]string{"name"}, nil,
		),
		machine: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, summarySubsystem, "machine_seconds_total"),
			"Total seconds for each machine.",
			[]string{"name", "id"}, nil,
		),
		editor: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, summarySubsystem, "editor_seconds_total"),
			"Total seconds for each editor.",
			[]string{"name"}, nil,
		),
		project: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, summarySubsystem, "project_seconds_total"),
			"Total seconds for each project.",
			[]string{"name"}, nil,
		),
		category: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, summarySubsystem, "category_seconds_total"),
			"Total seconds for each category.",
			[]string{"name"}, nil,
		),
		uri:       in.URI,
		fetchStat: FetchHTTP(in.Token, in.SSLVerify, in.Timeout, logger),
		logger:    logger,
	}, nil
}

func (c *summaryCollector) Update(ch chan<- prometheus.Metric) error {
	params := url.Values{}
	params.Add("start", "today")
	params.Add("end", "today")
	params.Add("cache", "false")

	body, fetchErr := c.fetchStat(c.uri, summaryEndpoint, params)
	if fetchErr != nil {
		return fetchErr
	}

	summaryStats := wakatimeSummary{}
	if err := ReadAndUnmarshal(body, &summaryStats); err != nil {
		return err
	}

	defer body.Close()

	for i, data := range summaryStats.Data {
		level.Info(c.logger).Log(
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
		level.Error(c.logger).Log("msg", "length of results is incorrect", "size", resultLength)
	}
	todaySummaryStats := summaryStats.Data[0]

	ch <- prometheus.MustNewConstMetric(
		c.total,
		prometheus.CounterValue,
		todaySummaryStats.GrandTotal.TotalSeconds,
	)

	for _, lang := range todaySummaryStats.Languages {
		ch <- prometheus.MustNewConstMetric(
			c.language,
			prometheus.CounterValue,
			lang.TotalSeconds,
			lang.Name,
		)
	}

	for _, os := range todaySummaryStats.OperatingSystems {
		ch <- prometheus.MustNewConstMetric(
			c.operatingSystem,
			prometheus.CounterValue,
			os.TotalSeconds,
			os.Name,
		)
	}

	for _, machine := range todaySummaryStats.Machines {
		ch <- prometheus.MustNewConstMetric(
			c.machine,
			prometheus.CounterValue,
			machine.TotalSeconds,
			machine.Name, machine.MachineNameID,
		)
	}

	for _, editor := range todaySummaryStats.Editors {
		ch <- prometheus.MustNewConstMetric(
			c.editor,
			prometheus.CounterValue,
			editor.TotalSeconds,
			editor.Name,
		)
	}

	for _, project := range todaySummaryStats.Projects {
		ch <- prometheus.MustNewConstMetric(
			c.project,
			prometheus.CounterValue,
			project.TotalSeconds,
			project.Name,
		)
	}

	for _, category := range todaySummaryStats.Categories {
		ch <- prometheus.MustNewConstMetric(
			c.category,
			prometheus.CounterValue,
			category.TotalSeconds,
			category.Name,
		)
	}

	return nil
}
