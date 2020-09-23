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
	"strconv"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	goalCollectorName = "goal"
	goalSubsystem     = "goal"
	goalEndpoint      = "goals"
)

type goalCollector struct {
	goalThreshold *prometheus.Desc
	goalProgress  *prometheus.Desc
	uri           url.URL
	fetchStat     func(url.URL, string, url.Values) (io.ReadCloser, error)
	logger        log.Logger
}

func init() {
	registerCollector(goalCollectorName, defaultEnabled, NewGoalCollector)
}

// NewGoalCollector returns a new Collector exposing all-time stats.
func NewGoalCollector(in CommonInputs, logger log.Logger) (Collector, error) {
	return &goalCollector{
		goalThreshold: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, goalSubsystem, "threshold_seconds"),
			"The goal as set through the wakatime interface.",
			[]string{
				"name", "id", "type", "delta", "enabled",
				"ignore_zero_days", "inverse", "snoozed", "tweeting",
			},
			nil,
		),
		goalProgress: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, goalSubsystem, "progress_seconds"),
			"Progress towards the goal.",
			[]string{
				"name", "id", "type", "delta", "enabled",
				"ignore_zero_days", "inverse", "snoozed", "tweeting",
			},
			nil,
		),
		uri:       in.URI,
		fetchStat: FetchHTTP(in.Token, in.SSLVerify, in.Timeout, logger),
		logger:    logger,
	}, nil
}

func (c *goalCollector) Update(ch chan<- prometheus.Metric) error {
	params := url.Values{}
	params.Add("cache", "false")

	body, fetchErr := c.fetchStat(c.uri, goalEndpoint, params)
	if fetchErr != nil {
		return fetchErr
	}

	goalStats := wakatimeGoal{}
	if err := ReadAndUnmarshal(body, &goalStats); err != nil {
		return err
	}

	defer body.Close()

	level.Info(c.logger).Log(
		"msg", "Collecting goals from Wakatime",
		"total", goalStats.Total,
		"pages", goalStats.TotalPages,
	)
	for i, data := range goalStats.Data {
		// the last element should be the most recent data
		currentChartData := data.ChartData[len(data.ChartData)-1]

		level.Info(c.logger).Log(
			"msg", "Collecting goal from Wakatime",
			"obj", i,
			"start", currentChartData.Range.Start,
			"end", currentChartData.Range.End,
			"text", currentChartData.Range.Text,
		)

		ch <- prometheus.MustNewConstMetric(
			c.goalThreshold,
			prometheus.GaugeValue,
			float64(currentChartData.GoalSeconds),
			data.Title,
			data.ID,
			data.Type,
			data.Delta,
			strconv.FormatBool(data.IsEnabled),
			strconv.FormatBool(data.IgnoreZeroDays),
			strconv.FormatBool(data.IsInverse),
			strconv.FormatBool(data.IsSnoozed),
			strconv.FormatBool(data.IsTweeting),
		)

		ch <- prometheus.MustNewConstMetric(
			c.goalProgress,
			prometheus.CounterValue,
			currentChartData.ActualSeconds,
			data.Title,
			data.ID,
			data.Type,
			data.Delta,
			strconv.FormatBool(data.IsEnabled),
			strconv.FormatBool(data.IgnoreZeroDays),
			strconv.FormatBool(data.IsInverse),
			strconv.FormatBool(data.IsSnoozed),
			strconv.FormatBool(data.IsTweeting),
		)
	}

	return nil
}
