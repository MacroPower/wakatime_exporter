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

package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	b64 "encoding/base64"
	"encoding/json"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	namespace = "wakatime"
	subsystem = "exporter"
)

var (
	wakaMetrics = metrics{
		"total":            newWakaMetric("seconds_total", "Total seconds.", prometheus.CounterValue, nil, nil),
		"cumulative_total": newWakaMetric("cumulative_seconds_total", "Total seconds (all time).", prometheus.CounterValue, nil, nil),
		"language":         newWakaMetric("language_seconds_total", "Total seconds for each language.", prometheus.CounterValue, []string{"name"}, nil),
		"operating_system": newWakaMetric("operating_system_seconds_total", "Total seconds for each operating system.", prometheus.CounterValue, []string{"name"}, nil),
		"machine":          newWakaMetric("machine_seconds_total", "Total seconds for each machine.", prometheus.CounterValue, []string{"name", "id"}, nil),
		"editor":           newWakaMetric("editor_seconds_total", "Total seconds for each editor.", prometheus.CounterValue, []string{"name"}, nil),
		"project":          newWakaMetric("project_seconds_total", "Total seconds for each project.", prometheus.CounterValue, []string{"name"}, nil),
		"category":         newWakaMetric("category_seconds_total", "Total seconds for each category.", prometheus.CounterValue, []string{"name"}, nil),
		"rank":             newWakaMetric("rank", "Current rank of the user.", prometheus.GaugeValue, nil, nil),
		"goal":             newWakaMetric("goal_seconds", "The goal.", prometheus.GaugeValue, []string{"name", "id", "type", "delta"}, nil),
		"goal_progress":    newWakaMetric("goal_progress_seconds_total", "Progress towards the goal.", prometheus.CounterValue, []string{"name", "id", "type", "delta"}, nil),
		"goal_info": newWakaMetric(
			"goal_info",
			"Information about the goal.",
			prometheus.GaugeValue,
			[]string{
				"name", "id", "ignore_zero_days",
				"is_enabled", "is_inverse", "is_snoozed", "is_tweeting",
			}, nil),
	}

	wakaUp = prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "up"), "Was the last scrape of wakatime successful.", nil, nil)
)

// Describe describes all the metrics ever exported by the wakatime exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range wakaMetrics {
		ch <- m.Desc
	}

	ch <- wakaUp
	ch <- e.totalScrapes.Desc()
	ch <- e.queryFailures.Desc()
}

// Collect all the metrics.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()

	up := e.scrape(ch)

	ch <- prometheus.MustNewConstMetric(wakaUp, prometheus.GaugeValue, up)
	ch <- e.totalScrapes
	ch <- e.queryFailures
}

func (e *Exporter) scrape(ch chan<- prometheus.Metric) (up float64) {
	level.Debug(e.logger).Log("msg", "Starting scrape")

	e.totalScrapes.Inc()

	dateUTC := time.Now().UTC().Add(e.tzOffset).Format("2006-01-02")
	userPath := path.Join(e.URI.Path, "users", e.user)
	userURL := *e.URI
	userURL.Path = userPath

	summariesBody, err := e.fetchStat(userURL, dateUTC, "summaries")
	goalsBody, err := e.fetchStat(userURL, dateUTC, "goals")
	leadersBody, err := e.fetchStat(*e.URI, dateUTC, "leaders")
	allTimeBody, err := e.fetchStat(userURL, dateUTC, "all_time_since_today")
	if err != nil {
		e.queryFailures.Inc()
		level.Error(e.logger).Log("msg", "Can't scrape wakatime", "err", err)
		return 0
	}

	respSummariesBody, readErr := ioutil.ReadAll(summariesBody)
	respGoalsBody, readErr := ioutil.ReadAll(goalsBody)
	respLeadersBody, readErr := ioutil.ReadAll(leadersBody)
	respAllTimeBody, readErr := ioutil.ReadAll(allTimeBody)
	if readErr != nil {
		e.queryFailures.Inc()
		level.Error(e.logger).Log("msg", "Can't read wakatime data", "err", readErr)
		return 0
	}

	var closeErr error
	closeErr = summariesBody.Close()
	closeErr = goalsBody.Close()
	closeErr = leadersBody.Close()
	closeErr = allTimeBody.Close()
	if closeErr != nil {
		e.queryFailures.Inc()
		level.Error(e.logger).Log("msg", "Can't close wakatime connection", "err", closeErr)
		return 0
	}

	var jsonErr error
	summaryStats := WakatimeSummary{}
	goalStats := WakatimeGoal{}
	leaderStats := WakatimeLeader{}
	totalStats := WakatimeTime{}
	jsonErr = json.Unmarshal(respSummariesBody, &summaryStats)
	jsonErr = json.Unmarshal(respGoalsBody, &goalStats)
	jsonErr = json.Unmarshal(respLeadersBody, &leaderStats)
	jsonErr = json.Unmarshal(respAllTimeBody, &totalStats)
	if jsonErr != nil {
		e.queryFailures.Inc()
		level.Error(e.logger).Log("msg", "Can't unmarshal wakatime data", "err", jsonErr)
		return 0
	}

	level.Info(e.logger).Log(
		"msg", "Collecting all time from Wakatime",
		"IsUpToDate", totalStats.Data.IsUpToDate,
	)
	if totalStats.Data.IsUpToDate == true {
		e.exportMetric(wakaMetrics["cumulative_total"], ch, totalStats.Data.TotalSeconds)
	}

	level.Info(e.logger).Log(
		"msg", "Collecting goals from Wakatime",
		"total", goalStats.Total,
		"pages", goalStats.TotalPages,
	)
	for _, data := range goalStats.Data {
		// the last element should be the most recent data
		currentChartData := data.ChartData[len(data.ChartData)-1]

		e.exportMetric(
			wakaMetrics["goal_progress"], ch, currentChartData.ActualSeconds,
			data.Title, data.ID, data.Type, data.Delta,
		)
		e.exportMetric(
			wakaMetrics["goal"], ch, float64(data.Seconds),
			data.Title, data.ID, data.Type, data.Delta,
		)
		e.exportMetric(
			wakaMetrics["goal_info"], ch, 1,
			data.Title, data.ID,
			b2str(data.IgnoreZeroDays), b2str(data.IsEnabled), b2str(data.IsInverse), b2str(data.IsSnoozed), b2str(data.IsTweeting),
		)
	}

	level.Info(e.logger).Log(
		"msg", "Collecting rank from Wakatime",
		"page", leaderStats.Page,
		"updated", leaderStats.ModifiedAt,
	)
	currentRank := float64(leaderStats.CurrentUser.Rank)
	e.exportMetric(wakaMetrics["rank"], ch, currentRank)

	for i, data := range summaryStats.Data {
		level.Info(e.logger).Log(
			"msg", "Collecting summary from Wakatime",
			"obj", i,
			"start", data.Range.Start.String(),
			"end", data.Range.End.String(),
			"tz", data.Range.Timezone,
			"text", data.Range.Text,
		)
	}

	resultLength := len(summaryStats.Data)
	if resultLength != 1 {
		level.Error(e.logger).Log("msg", "Length of results is incorrect", "size", resultLength)
	}
	todaySummaryStats := summaryStats.Data[0]

	e.exportMetric(wakaMetrics["total"], ch, todaySummaryStats.GrandTotal.TotalSeconds)

	for _, lang := range todaySummaryStats.Languages {
		e.exportMetric(wakaMetrics["language"], ch, lang.TotalSeconds, lang.Name)
	}

	for _, os := range todaySummaryStats.OperatingSystems {
		e.exportMetric(wakaMetrics["operating_system"], ch, os.TotalSeconds, os.Name)
	}

	for _, machine := range todaySummaryStats.Machines {
		e.exportMetric(wakaMetrics["machine"], ch, machine.TotalSeconds, machine.Name, machine.MachineNameID)
	}

	for _, editor := range todaySummaryStats.Editors {
		e.exportMetric(wakaMetrics["editor"], ch, editor.TotalSeconds, editor.Name)
	}

	for _, project := range todaySummaryStats.Projects {
		e.exportMetric(wakaMetrics["project"], ch, project.TotalSeconds, project.Name)
	}

	for _, category := range todaySummaryStats.Categories {
		e.exportMetric(wakaMetrics["category"], ch, category.TotalSeconds, category.Name)
	}

	level.Info(e.logger).Log("msg", "Finished scraping Wakatime", "start", summaryStats.Start.String(), "end", summaryStats.End.String())

	return 1
}

func fetchHTTP(token string, sslVerify bool, timeout time.Duration, logger log.Logger) func(uri url.URL, dateUTC string, subPath string) (io.ReadCloser, error) {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: !sslVerify}}
	client := http.Client{
		Timeout:   timeout,
		Transport: tr,
	}
	sEnc := b64.StdEncoding.EncodeToString([]byte(token))
	return func(uri url.URL, dateUTC string, subPath string) (io.ReadCloser, error) {
		level.Info(logger).Log("msg", "Scraping Wakatime", "date", dateUTC, "path", subPath)

		params := url.Values{}
		params.Add("start", dateUTC)
		params.Add("end", dateUTC)

		uri.Path = path.Join(uri.Path, subPath)
		uri.RawQuery = params.Encode()

		url := uri.String()

		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		req.Header = map[string][]string{
			"Authorization": {"Basic " + sEnc},
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
			resp.Body.Close()
			return nil, fmt.Errorf("HTTP status %d", resp.StatusCode)
		}
		return resp.Body, nil
	}
}

// NewExporter returns an initialized Exporter.
func NewExporter(uri string, user string, token string, sslVerify bool, tzOffset time.Duration, timeout time.Duration, logger log.Logger) (*Exporter, error) {
	wakaBaseURI, err := url.Parse(uri)
	if err != nil {
		level.Error(logger).Log("msg", "Malformed URL", "err", err.Error())
		return nil, err
	}

	var fetchStat func(url.URL, string, string) (io.ReadCloser, error)
	fetchStat = fetchHTTP(token, sslVerify, timeout, logger)

	return &Exporter{
		URI:       wakaBaseURI,
		user:      user,
		fetchStat: fetchStat,
		tzOffset:  tzOffset,
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "Was the last scrape of wakatime successful.",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_scrapes_total",
			Help:      "Current total wakatime scrapes.",
		}),
		queryFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_query_failures_total",
			Help:      "Number of errors.",
		}),
		logger: logger,
	}, nil
}

func main() {
	var (
		listenAddress = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9212").Envar("WAKA_LISTEN_ADDR").String()
		metricsPath   = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").Envar("WAKA_METRICS_PATH").String()
		wakaScrapeURI = kingpin.Flag("wakatime.scrape-uri", "Base path to query for Wakatime data.").Default("https://wakatime.com/api/v1").Envar("WAKA_SCRAPE_URI").String()
		wakaUser      = kingpin.Flag("wakatime.user", "User to query for Wakatime data.").Default("current").Envar("WAKA_USER").String()
		wakaToken     = kingpin.Flag("wakatime.api-key", "Token to use when getting stats from Wakatime.").Required().Envar("WAKA_API_KEY").String()
		wakaOffset    = kingpin.Flag("wakatime.t-offset", "Time offset (from UTC) for managing Wakatime dates.").Default("0s").Envar("WAKA_TIME_OFFSET").Duration()
		wakaTimeout   = kingpin.Flag("wakatime.timeout", "Timeout for trying to get stats from Wakatime.").Default("5s").Envar("WAKA_TIMEOUT").Duration()
		wakaSSLVerify = kingpin.Flag("wakatime.ssl-verify", "Flag that enables SSL certificate verification for the scrape URI.").Default("true").Envar("WAKA_SSL_VERIFY").Bool()
	)

	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print("wakatime_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promlog.New(promlogConfig)

	level.Info(logger).Log("msg", "Starting wakatime_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "context", version.BuildContext())

	exporter, err := NewExporter(*wakaScrapeURI, *wakaUser, *wakaToken, *wakaSSLVerify, *wakaOffset, *wakaTimeout, logger)
	if err != nil {
		level.Error(logger).Log("msg", "Error creating an exporter", "err", err)
		os.Exit(1)
	}
	prometheus.MustRegister(exporter)
	prometheus.MustRegister(version.NewCollector("wakatime_exporter"))

	level.Info(logger).Log("msg", "Listening on address", "address", *listenAddress)
	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Wakatime Exporter</title></head>
             <body>
             <h1>Wakatime Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}
