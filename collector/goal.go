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
	subsystem = "goal"
	endpoint  = "goals"
)

const (
	goalCollectorName = "goal"
	goalSubsystem     = "goal"
	goalEndpoint      = "goals"
)

type goalCollector struct {
	goalSeconds  *prometheus.Desc
	goalProgress *prometheus.Desc
	goalInfo     *prometheus.Desc
	uri          url.URL
	fetchStat    func(url.URL, string, url.Values) (io.ReadCloser, error)
	logger       log.Logger
}

func init() {
	registerCollector(goalCollectorName, defaultEnabled, NewGoalCollector)
}

// NewGoalCollector returns a new Collector exposing all-time stats.
func NewGoalCollector(in CommonInputs, logger log.Logger) (Collector, error) {
	return &goalCollector{
		goalSeconds: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, goalSubsystem, "goal_seconds"),
			"The Goal.",
			[]string{"name", "id", "type", "delta"}, nil,
		),
		goalProgress: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, goalSubsystem, "goal_progress"),
			"Progress towards the goal.",
			[]string{"name", "id", "type", "delta"}, nil,
		),
		goalInfo: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, goalSubsystem, "goal_info"),
			"Information about the goal.",
			[]string{"name", "id", "ignore_zero_days", "is_enabled", "is_inverse", "is_snoozed", "is_tweeting"}, nil,
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
			c.goalSeconds,
			prometheus.GaugeValue,
			float64(currentChartData.GoalSeconds),
			data.Title, data.ID, data.Type, data.Delta,
		)

		ch <- prometheus.MustNewConstMetric(
			c.goalProgress,
			prometheus.CounterValue,
			currentChartData.ActualSeconds,
			data.Title, data.ID, data.Type, data.Delta,
		)

		ch <- prometheus.MustNewConstMetric(
			c.goalInfo,
			prometheus.GaugeValue,
			float64(1),
			data.Title, data.ID,
			BoolToBinary(data.IgnoreZeroDays),
			BoolToBinary(data.IsEnabled),
			BoolToBinary(data.IsInverse),
			BoolToBinary(data.IsSnoozed),
			BoolToBinary(data.IsTweeting),
		)
	}

	return nil
}
