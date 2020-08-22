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

package alltime

import (
	"errors"
	"net/url"
	"time"

	exporter "github.com/MacroPower/wakatime_exporter/lib"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	subsystem = "alltime"
	endpoint  = "all_time_since_today"
)

var (
	wakaMetrics = exporter.Metrics{
		"total": exporter.NewWakaMetric("cumulative_seconds_total", "Total seconds (all time).", prometheus.CounterValue, nil, nil),
	}
)

// Exporter is the local definition of Exporter
type Exporter exporter.Exporter

// NewExporter creates the AllTime exporter
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

	userURL := exporter.UserPath(e.URI, e.User)

	params := url.Values{}
	params.Add("cache", "false")

	body, fetchErr := e.FetchStat(userURL, endpoint, params)
	if fetchErr != nil {
		return fetchErr
	}

	alltimeStats := wakatimeAlltime{}
	err := exporter.ReadAndUnmarshal(body, &alltimeStats)
	if err != nil {
		return err
	}

	defer body.Close()

	level.Info(e.Logger).Log(
		"msg", "Collecting all time from Wakatime",
		"IsUpToDate", alltimeStats.Data.IsUpToDate,
	)
	if alltimeStats.Data.IsUpToDate == true {
		e.ExportMetric(wakaMetrics["total"], ch, alltimeStats.Data.TotalSeconds)
	} else {
		return errors.New("skipped scrape of alltime metrics because they were not up to date")
	}

	return nil
}
