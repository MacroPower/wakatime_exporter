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
	"errors"
	"io"
	"net/url"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	allTimeCollector = "all-time"
	allTimeSubsystem = "alltime"
	allTimeEndpoint  = "all_time_since_today"
)

type alltimeCollector struct {
	total     *prometheus.Desc
	uri       url.URL
	fetchStat func(url.URL, string, url.Values) (io.ReadCloser, error)
	logger    log.Logger
}

func init() {
	registerCollector(allTimeCollector, defaultEnabled, NewAllTimeCollector)
}

// NewAllTimeCollector returns a new Collector exposing all-time stats.
func NewAllTimeCollector(in CommonInputs, logger log.Logger) (Collector, error) {
	return &alltimeCollector{
		total: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, allTimeSubsystem, "cumulative_seconds_total"),
			"Total seconds (all time).",
			nil, nil,
		),
		uri:       in.URI,
		fetchStat: FetchHTTP(in.Token, in.SSLVerify, in.Timeout, logger),
		logger:    logger,
	}, nil
}

func (c *alltimeCollector) Update(ch chan<- prometheus.Metric) error {
	params := url.Values{}
	params.Add("cache", "false")

	body, fetchErr := c.fetchStat(c.uri, allTimeEndpoint, params)
	if fetchErr != nil {
		return fetchErr
	}

	alltimeStats := wakatimeAlltime{}
	if err := ReadAndUnmarshal(body, &alltimeStats); err != nil {
		return err
	}

	defer body.Close()

	level.Info(c.logger).Log(
		"msg", "Collecting all-time from Wakatime",
		"IsUpToDate", alltimeStats.Data.IsUpToDate,
	)
	if alltimeStats.Data.IsUpToDate == true {
		ch <- prometheus.MustNewConstMetric(
			c.total,
			prometheus.CounterValue,
			alltimeStats.Data.TotalSeconds,
		)
	} else {
		return errors.New("skipped scrape of all-time metrics because they were not up to date")
	}

	return nil
}
