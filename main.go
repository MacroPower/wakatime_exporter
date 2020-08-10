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
	"net/http"
	"net/url"
	"os"

	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"

	alltime "github.com/MacroPower/wakatime_exporter/collectors/alltime"
	goal "github.com/MacroPower/wakatime_exporter/collectors/goal"
	leader "github.com/MacroPower/wakatime_exporter/collectors/leader"
	summary "github.com/MacroPower/wakatime_exporter/collectors/summary"
)

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

	wakaBaseURI, err := url.Parse(*wakaScrapeURI)
	if err != nil {
		level.Error(logger).Log("msg", "Error parsing URL", "err", err)
		os.Exit(1)
	}

	summaryExporter := summary.NewExporter(wakaBaseURI, *wakaUser, *wakaToken, *wakaSSLVerify, *wakaOffset, *wakaTimeout, logger)
	leaderExporter := leader.NewExporter(wakaBaseURI, *wakaUser, *wakaToken, *wakaSSLVerify, *wakaOffset, *wakaTimeout, logger)
	goalExporter := goal.NewExporter(wakaBaseURI, *wakaUser, *wakaToken, *wakaSSLVerify, *wakaOffset, *wakaTimeout, logger)
	alltimeExporter := alltime.NewExporter(wakaBaseURI, *wakaUser, *wakaToken, *wakaSSLVerify, *wakaOffset, *wakaTimeout, logger)

	prometheus.MustRegister(summaryExporter)
	prometheus.MustRegister(leaderExporter)
	prometheus.MustRegister(goalExporter)
	prometheus.MustRegister(alltimeExporter)
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
