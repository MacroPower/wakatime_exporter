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

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	leaderCollectorName = "leader"
	leaderSubsystem     = "leaderboard"
	leaderEndpoint      = "leaders"
)

type leaderCollector struct {
	rank      *prometheus.Desc
	uri       url.URL
	fetchStat func(url.URL, string, url.Values) (io.ReadCloser, error)
	logger    log.Logger
}

func init() {
	registerCollector(leaderCollectorName, defaultEnabled, NewLeaderCollector)
}

// NewLeaderCollector returns a new Collector exposing all-time stats.
func NewLeaderCollector(in CommonInputs, logger log.Logger) (Collector, error) {
	return &leaderCollector{
		rank: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, leaderSubsystem, "rank"),
			"Current rank of the user.",
			nil, nil,
		),
		uri:       in.BaseURI,
		fetchStat: FetchHTTP(in.Token, in.SSLVerify, in.Timeout, logger),
		logger:    logger,
	}, nil
}

func (c *leaderCollector) Update(ch chan<- prometheus.Metric) error {
	params := url.Values{}
	params.Add("cache", "false")

	body, fetchErr := c.fetchStat(c.uri, leaderEndpoint, params)
	if fetchErr != nil {
		return fetchErr
	}

	leaderStats := wakatimeLeader{}
	if err := ReadAndUnmarshal(body, &leaderStats); err != nil {
		return err
	}

	defer body.Close()

	level.Info(c.logger).Log(
		"msg", "Collecting rank from Wakatime",
		"page", leaderStats.Page,
		"updated", leaderStats.ModifiedAt,
	)

	ch <- prometheus.MustNewConstMetric(
		c.rank,
		prometheus.GaugeValue,
		float64(leaderStats.CurrentUser.Rank),
	)

	return nil
}
