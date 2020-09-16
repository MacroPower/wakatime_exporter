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
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/MacroPower/wakatime_exporter/collector"
)

// handler wraps an unfiltered http.Handler but uses a filtered handler,
// created on the fly, if filtering is requested. Create instances with
// newHandler.
type handler struct {
	unfilteredHandler http.Handler
	// exporterMetricsRegistry is a separate registry for the metrics about
	// the exporter itself.
	exporterMetricsRegistry *prometheus.Registry
	includeExporterMetrics  bool
	commonInputs            collector.CommonInputs
	logger                  log.Logger
}

func newHandler(commonInputs collector.CommonInputs, includeExporterMetrics bool, logger log.Logger) *handler {
	h := &handler{
		exporterMetricsRegistry: prometheus.NewRegistry(),
		includeExporterMetrics:  includeExporterMetrics,
		commonInputs:            commonInputs,
		logger:                  logger,
	}
	if h.includeExporterMetrics {
		h.exporterMetricsRegistry.MustRegister(
			prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
			prometheus.NewGoCollector(),
		)
	}
	if innerHandler, err := h.innerHandler(); err != nil {
		panic(fmt.Sprintf("Couldn't create metrics handler: %s", err))
	} else {
		h.unfilteredHandler = innerHandler
	}
	return h
}

// ServeHTTP implements http.Handler.
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filters := r.URL.Query()["collect[]"]
	level.Debug(h.logger).Log("msg", "collect query:", "filters", filters)

	if len(filters) == 0 {
		// No filters, use the prepared unfiltered handler.
		h.unfilteredHandler.ServeHTTP(w, r)
		return
	}
	// To serve filtered metrics, we create a filtering handler on the fly.
	filteredHandler, err := h.innerHandler(filters...)
	if err != nil {
		level.Warn(h.logger).Log("msg", "Couldn't create filtered metrics handler:", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Couldn't create filtered metrics handler: %s", err)))
		return
	}
	filteredHandler.ServeHTTP(w, r)
}

// innerHandler is used to create both the one unfiltered http.Handler to be
// wrapped by the outer handler and also the filtered handlers created on the
// fly. The former is accomplished by calling innerHandler without any arguments
// (in which case it will log all the collectors enabled via command-line
// flags).
func (h *handler) innerHandler(filters ...string) (http.Handler, error) {
	nc, err := collector.NewWakaCollector(h.commonInputs, h.logger, filters...)
	if err != nil {
		return nil, fmt.Errorf("couldn't create collector: %s", err)
	}

	// Only log the creation of an unfiltered handler, which should happen
	// only once upon startup.
	if len(filters) == 0 {
		level.Info(h.logger).Log("msg", "Enabled collectors")
		collectors := []string{}
		for n := range nc.Collectors {
			collectors = append(collectors, n)
		}
		sort.Strings(collectors)
		for _, c := range collectors {
			level.Info(h.logger).Log("collector", c)
		}
	}

	r := prometheus.NewRegistry()
	r.MustRegister(version.NewCollector("wakatime_exporter"))
	if err := r.Register(nc); err != nil {
		return nil, fmt.Errorf("couldn't register collector: %s", err)
	}
	handler := promhttp.HandlerFor(
		prometheus.Gatherers{h.exporterMetricsRegistry, r},
		promhttp.HandlerOpts{
			ErrorHandling: promhttp.ContinueOnError,
			Registry:      h.exporterMetricsRegistry,
		},
	)
	if h.includeExporterMetrics {
		// Note that we have to use h.exporterMetricsRegistry here to
		// use the same promhttp metrics for all expositions.
		handler = promhttp.InstrumentMetricHandler(
			h.exporterMetricsRegistry, handler,
		)
	}
	return handler, nil
}

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
