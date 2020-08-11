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
	"io"
	"net/url"
	"time"

	exporter "github.com/MacroPower/wakatime_exporter/lib"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "wakatime"
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
func NewExporter(baseURI *url.URL, user string, token string, sslVerify bool, tzOffset time.Duration, timeout time.Duration, logger log.Logger) *Exporter {
	var fetchStat func(url.URL, string, string) (io.ReadCloser, error)
	fetchStat = exporter.FetchHTTP(token, sslVerify, timeout, logger)

	return &Exporter{
		URI:       baseURI,
		Endpoint:  endpoint,
		User:      user,
		FetchStat: fetchStat,
		TZOffset:  tzOffset,
		Up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "up",
			Help:      "Was the last scrape of wakatime successful.",
		}),
		TotalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "exporter_scrapes_total",
			Help:      "Current total wakatime scrapes.",
		}),
		QueryFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "exporter_query_failures_total",
			Help:      "Number of errors.",
		}),
		Logger: logger,
	}
}

// Describe describes all the metrics ever exported by the wakatime exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range wakaMetrics {
		ch <- m.Desc
	}

	ch <- e.Up.Desc()
	ch <- e.TotalScrapes.Desc()
	ch <- e.QueryFailures.Desc()
}

// Collect all the metrics.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.Mutex.Lock() // To protect metrics from concurrent collects.
	defer e.Mutex.Unlock()

	err := e.scrape(ch)
	up := float64(1)
	if err != nil {
		up = float64(0)
		e.QueryFailures.Inc()
		level.Error(e.Logger).Log("msg", "Can't scrape wakatime", "subsystem", subsystem, "err", err)
	}
	e.Up.Set(up)

	ch <- e.Up
	ch <- e.TotalScrapes
	ch <- e.QueryFailures
}

func (e *Exporter) exportMetric(m exporter.MetricInfo, ch chan<- prometheus.Metric, value float64, labels ...string) {
	ch <- prometheus.MustNewConstMetric(m.Desc, m.Type, value, labels...)
}

func (e *Exporter) scrape(ch chan<- prometheus.Metric) error {
	level.Debug(e.Logger).Log("msg", "Starting scrape")

	e.TotalScrapes.Inc()

	dateUTC := exporter.GetDate(e.TZOffset)
	body, fetchErr := e.FetchStat(*e.URI, dateUTC, endpoint)
	defer body.Close()
	if fetchErr != nil {
		return fetchErr
	}

	leaderStats := wakatimeLeader{}
	err := exporter.ReadAndUnmarshal(body, &leaderStats)
	if err != nil {
		return err
	}

	level.Info(e.Logger).Log(
		"msg", "Collecting rank from Wakatime",
		"page", leaderStats.Page,
		"updated", leaderStats.ModifiedAt,
	)
	currentRank := float64(leaderStats.CurrentUser.Rank)
	e.exportMetric(wakaMetrics["rank"], ch, currentRank)

	return nil
}
