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
	"path"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const (
	statsCollectorName = "stats"
	statsEndpoint      = "stats"
)

var (
	statsRange = kingpin.Flag(
		"wakatime.stats-range",
		"Range for stats collector. Can be last_7_days, last_30_days, last_6_months, last_year, or all_time.",
	).Default("last_7_days").Envar("WAKA_STATS_RANGE").String()
)

type statsCollector struct {
	total           *prometheus.Desc
	dailyAverage    *prometheus.Desc
	bestDay         *prometheus.Desc
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
	registerCollector(statsCollectorName, defaultEnabled, NewStatsCollector)
}

// NewStatsCollector returns a new Collector exposing stats for a given range.
func NewStatsCollector(in CommonInputs, logger log.Logger) (Collector, error) {
	const subsystem = "stats"
	return &statsCollector{
		total: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "total_seconds"),
			"Total seconds for the stats range.",
			[]string{"range"}, nil,
		),
		dailyAverage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "daily_average_seconds"),
			"Daily average seconds for the stats range.",
			[]string{"range"}, nil,
		),
		bestDay: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "best_day_seconds"),
			"Best day total seconds for the stats range.",
			[]string{"range", "date"}, nil,
		),
		language: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "language_seconds"),
			"Total seconds for each language in the stats range.",
			[]string{"range", "name"}, nil,
		),
		operatingSystem: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "operating_system_seconds"),
			"Total seconds for each operating system in the stats range.",
			[]string{"range", "name"}, nil,
		),
		machine: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "machine_seconds"),
			"Total seconds for each machine in the stats range.",
			[]string{"range", "name", "id"}, nil,
		),
		editor: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "editor_seconds"),
			"Total seconds for each editor in the stats range.",
			[]string{"range", "name"}, nil,
		),
		project: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "project_seconds"),
			"Total seconds for each project in the stats range.",
			[]string{"range", "name"}, nil,
		),
		category: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "category_seconds"),
			"Total seconds for each category in the stats range.",
			[]string{"range", "name"}, nil,
		),
		uri:       in.URI,
		fetchStat: FetchHTTP(in.Token, in.SSLVerify, in.Timeout, logger),
		logger:    logger,
	}, nil
}

func (c *statsCollector) Update(ch chan<- prometheus.Metric) error {
	params := url.Values{}
	params.Add("cache", "false")

	// Build the endpoint path with the range
	endpointPath := path.Join(statsEndpoint, *statsRange)

	body, fetchErr := c.fetchStat(c.uri, endpointPath, params)
	if fetchErr != nil {
		return fetchErr
	}

	statsData := wakatimeStats{}
	if err := ReadAndUnmarshal(body, &statsData); err != nil {
		return err
	}

	defer body.Close()

	data := statsData.Data
	rangeLabel := *statsRange

	// Check if data is up to date
	if !data.IsUpToDate {
		level.Warn(c.logger).Log("msg", "Stats data is not up to date", "status", data.Status, "percent_calculated", data.PercentCalculated)
	}

	level.Info(c.logger).Log(
		"msg", "Collecting stats from Wakatime",
		"range", rangeLabel,
		"is_up_to_date", data.IsUpToDate,
		"total_seconds", data.TotalSeconds,
	)

	// Total seconds
	ch <- prometheus.MustNewConstMetric(
		c.total,
		prometheus.GaugeValue,
		data.TotalSeconds,
		rangeLabel,
	)

	// Daily average
	ch <- prometheus.MustNewConstMetric(
		c.dailyAverage,
		prometheus.GaugeValue,
		data.DailyAverage,
		rangeLabel,
	)

	// Best day
	if data.BestDay.Date != "" {
		ch <- prometheus.MustNewConstMetric(
			c.bestDay,
			prometheus.GaugeValue,
			data.BestDay.TotalSeconds,
			rangeLabel, data.BestDay.Date,
		)
	}

	// Languages
	for _, lang := range data.Languages {
		ch <- prometheus.MustNewConstMetric(
			c.language,
			prometheus.GaugeValue,
			lang.TotalSeconds,
			rangeLabel, lang.Name,
		)
	}

	// Operating Systems
	for _, os := range data.OperatingSystems {
		ch <- prometheus.MustNewConstMetric(
			c.operatingSystem,
			prometheus.GaugeValue,
			os.TotalSeconds,
			rangeLabel, os.Name,
		)
	}

	// Machines
	for _, machine := range data.Machines {
		ch <- prometheus.MustNewConstMetric(
			c.machine,
			prometheus.GaugeValue,
			machine.TotalSeconds,
			rangeLabel, machine.Name, machine.MachineNameID,
		)
	}

	// Editors
	for _, editor := range data.Editors {
		ch <- prometheus.MustNewConstMetric(
			c.editor,
			prometheus.GaugeValue,
			editor.TotalSeconds,
			rangeLabel, editor.Name,
		)
	}

	// Projects
	for _, project := range data.Projects {
		ch <- prometheus.MustNewConstMetric(
			c.project,
			prometheus.GaugeValue,
			project.TotalSeconds,
			rangeLabel, project.Name,
		)
	}

	// Categories (includes "ai coding")
	for _, category := range data.Categories {
		ch <- prometheus.MustNewConstMetric(
			c.category,
			prometheus.GaugeValue,
			category.TotalSeconds,
			rangeLabel, category.Name,
		)
	}

	return nil
}
