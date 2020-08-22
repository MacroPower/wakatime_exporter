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

package leader

import (
	"net/url"
	"time"

	exporter "github.com/MacroPower/wakatime_exporter/lib"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	subsystem = "leader"
	endpoint  = "leaders"
)

var (
	wakaMetrics = exporter.Metrics{
		"rank": exporter.NewWakaMetric("rank", "Current rank of the user.", prometheus.GaugeValue, nil, nil),
	}
)

// Exporter is the local definition of Exporter
type Exporter exporter.Exporter

// NewExporter creates the Leader exporter
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
	params.Add("cache", "false")

	body, fetchErr := e.FetchStat(*e.URI, endpoint, params)
	if fetchErr != nil {
		return fetchErr
	}

	leaderStats := wakatimeLeader{}
	err := exporter.ReadAndUnmarshal(body, &leaderStats)
	if err != nil {
		return err
	}

	defer body.Close()

	level.Info(e.Logger).Log(
		"msg", "Collecting rank from Wakatime",
		"page", leaderStats.Page,
		"updated", leaderStats.ModifiedAt,
	)
	currentRank := float64(leaderStats.CurrentUser.Rank)
	e.ExportMetric(wakaMetrics["rank"], ch, currentRank)

	return nil
}
