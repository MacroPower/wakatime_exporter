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
	leaderCollectorName = "leader"
	leaderSubsystem     = "leader"
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
