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
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	durationsCollectorName = "durations"
	durationsEndpoint      = "durations"
)

type durationsCollector struct {
	aiLinesAdded      *prometheus.Desc
	aiLinesDeleted    *prometheus.Desc
	humanLinesAdded   *prometheus.Desc
	humanLinesDeleted *prometheus.Desc
	totalDuration     *prometheus.Desc
	uri               url.URL
	fetchStat         func(url.URL, string, url.Values) (io.ReadCloser, error)
	logger            log.Logger
}

func init() {
	registerCollector(durationsCollectorName, defaultEnabled, NewDurationsCollector)
}

// NewDurationsCollector returns a new Collector exposing durations with AI coding metrics.
func NewDurationsCollector(in CommonInputs, logger log.Logger) (Collector, error) {
	const subsystem = "durations"
	return &durationsCollector{
		aiLinesAdded: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "ai_lines_added_total"),
			"Total lines added by AI coding assistants today.",
			nil, nil,
		),
		aiLinesDeleted: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "ai_lines_deleted_total"),
			"Total lines deleted by AI coding assistants today.",
			nil, nil,
		),
		humanLinesAdded: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "human_lines_added_total"),
			"Total lines added by human typing today.",
			nil, nil,
		),
		humanLinesDeleted: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "human_lines_deleted_total"),
			"Total lines deleted by human typing today.",
			nil, nil,
		),
		totalDuration: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "seconds_total"),
			"Total duration in seconds today.",
			nil, nil,
		),
		uri:       in.URI,
		fetchStat: FetchHTTP(in.Token, in.SSLVerify, in.Timeout, logger),
		logger:    logger,
	}, nil
}

func (c *durationsCollector) Update(ch chan<- prometheus.Metric) error {
	params := url.Values{}
	params.Add("date", time.Now().Format("2006-01-02"))
	params.Add("cache", "false")

	body, fetchErr := c.fetchStat(c.uri, durationsEndpoint, params)
	if fetchErr != nil {
		return fetchErr
	}

	durationsData := wakatimeDurations{}
	if err := ReadAndUnmarshal(body, &durationsData); err != nil {
		return err
	}

	defer func() { _ = body.Close() }()

	// Aggregate all duration data
	var totalAIAdded, totalAIDeleted, totalHumanAdded, totalHumanDeleted int
	var totalDuration float64

	for _, duration := range durationsData.Data {
		totalAIAdded += duration.AIAdditions
		totalAIDeleted += duration.AIDeletions
		totalHumanAdded += duration.HumanAdditions
		totalHumanDeleted += duration.HumanDeletions
		totalDuration += duration.Duration
	}

	_ = level.Info(c.logger).Log(
		"msg", "Collecting durations from Wakatime",
		"duration_count", len(durationsData.Data),
		"ai_lines_added", totalAIAdded,
		"human_lines_added", totalHumanAdded,
		"total_duration", totalDuration,
	)

	// AI lines added
	ch <- prometheus.MustNewConstMetric(
		c.aiLinesAdded,
		prometheus.CounterValue,
		float64(totalAIAdded),
	)

	// AI lines deleted
	ch <- prometheus.MustNewConstMetric(
		c.aiLinesDeleted,
		prometheus.CounterValue,
		float64(totalAIDeleted),
	)

	// Human lines added
	ch <- prometheus.MustNewConstMetric(
		c.humanLinesAdded,
		prometheus.CounterValue,
		float64(totalHumanAdded),
	)

	// Human lines deleted
	ch <- prometheus.MustNewConstMetric(
		c.humanLinesDeleted,
		prometheus.CounterValue,
		float64(totalHumanDeleted),
	)

	// Total duration
	ch <- prometheus.MustNewConstMetric(
		c.totalDuration,
		prometheus.CounterValue,
		totalDuration,
	)

	return nil
}
