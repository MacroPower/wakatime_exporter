/*
Copyright 2015 The Prometheus Authors
Modifications Copyright 2020 Jacob Colvin (MacroPower)
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

package main

import (
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/MacroPower/wakatime_exporter/collector"
)

// UserPath appends the User path to a given URL
func UserPath(uri *url.URL, user string) url.URL {
	userURL := *uri
	userPath := path.Join(userURL.Path, "users", user)

	userURL.Path = userPath
	return userURL
}

func main() {
	var (
		disableDefaultCollectors = kingpin.Flag(
			"collector.disable-defaults",
			"Set all collectors to disabled by default.",
		).Default("false").Envar("WAKA_DISABLE_DEFAULT_COLLECTORS").Bool()

		listenAddress = kingpin.Flag(
			"web.listen-address",
			"Address to listen on for web interface and metrics.",
		).Default(":9212").Envar("WAKA_LISTEN_ADDRESS").String()

		metricsPath = kingpin.Flag(
			"web.metrics-path",
			"Path under which to expose metrics.",
		).Default("/metrics").Envar("WAKA_METRICS_PATH").String()

		disableExporterMetrics = kingpin.Flag(
			"web.disable-exporter-metrics",
			"Exclude metrics about the exporter itself (promhttp_*, process_*, go_*).",
		).Default("false").Envar("WAKA_DISABLE_EXPORTER_METRICS").Bool()

		wakaScrapeURI = kingpin.Flag(
			"wakatime.scrape-uri",
			"Base path to query for Wakatime data.",
		).Default("https://wakatime.com/api/v1").Envar("WAKA_SCRAPE_URI").String()

		wakaUser = kingpin.Flag(
			"wakatime.user",
			"User to query for Wakatime data.",
		).Default("current").Envar("WAKA_USER").String()

		wakaToken = kingpin.Flag(
			"wakatime.api-key",
			"Token to use when getting stats from Wakatime.",
		).Required().Envar("WAKA_API_KEY").String()

		wakaTimeout = kingpin.Flag(
			"wakatime.timeout",
			"Timeout for trying to get stats from Wakatime.",
		).Default("5s").Envar("WAKA_TIMEOUT").Duration()

		wakaSSLVerify = kingpin.Flag(
			"wakatime.ssl-verify",
			"Flag that enables SSL certificate verification for the scrape URI.",
		).Default("true").Envar("WAKA_SSL_VERIFY").Bool()
	)

	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print("wakatime_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promlog.New(promlogConfig)

	if *disableDefaultCollectors {
		collector.DisableDefaultCollectors()
	}
	level.Info(logger).Log("msg", "Starting wakatime_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())

	wakaBaseURI, err := url.Parse(*wakaScrapeURI)
	if err != nil {
		level.Error(logger).Log("msg", "Error parsing URL", "err", err)
		os.Exit(1)
	}

	http.Handle(*metricsPath, newHandler(collector.CommonInputs{
		BaseURI:   *wakaBaseURI,
		URI:       UserPath(wakaBaseURI, *wakaUser),
		Token:     *wakaToken,
		SSLVerify: *wakaSSLVerify,
		Timeout:   *wakaTimeout,
	}, !*disableExporterMetrics, logger))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Wakatime Exporter</title></head>
			<body>
			<h1>Wakatime Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	level.Info(logger).Log("msg", "Listening on", "address", *listenAddress)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}
