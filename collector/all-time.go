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
