package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

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
)

var (
	wakaLabelNames = []string{"lang"}
)

func newWakaMetric(metricName string, docString string, t prometheus.ValueType, constLabels prometheus.Labels) metricInfo {
	return metricInfo{
		Desc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "exporter", metricName),
			docString,
			wakaLabelNames,
			constLabels,
		),
		Type: t,
	}
}

type metrics map[string]metricInfo

var (
	wakaMetrics = metrics{
		"test": newWakaMetric("test", "foobar", prometheus.CounterValue, nil),
	}

	wakaInfo = prometheus.NewDesc(prometheus.BuildFQName(namespace, "version", "info"), "Exporter version info.", []string{"release_date", "version"}, nil)
	wakaUp   = prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "up"), "Was the last scrape of wakatime successful.", nil, nil)
)

// Exporter collects stats from the given URI and exports them using
// the prometheus metrics package.
type Exporter struct {
	URI       string
	mutex     sync.RWMutex
	fetchStat func() (io.ReadCloser, error)

	up                          prometheus.Gauge
	totalScrapes, queryFailures prometheus.Counter
	logger                      log.Logger
}

type metricInfo struct {
	Desc *prometheus.Desc
	Type prometheus.ValueType
}

// Describe describes all the metrics ever exported by the wakatime exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range wakaMetrics {
		ch <- m.Desc
	}

	ch <- wakaInfo
	ch <- wakaUp
	ch <- e.totalScrapes.Desc()
	ch <- e.queryFailures.Desc()
}

func (e *Exporter) scrape(ch chan<- prometheus.Metric) (up float64) {
	e.totalScrapes.Inc()
	var err error

	body, err := e.fetchStat()

	// DISCARD THE BODY
	_ = body

	// AN EXAMPLE METRIC
	var value float64
	var valueInt int64

	valueInt = 1
	value = float64(valueInt)

	e.exportMetric(wakaMetrics["test"], ch, value, "python")

	if err != nil {
		e.queryFailures.Inc()
		level.Error(e.logger).Log("msg", "Can't scrape wakatime", "err", err)
		return 0
	}

	return 1
}

func (e *Exporter) exportMetric(m metricInfo, ch chan<- prometheus.Metric, value float64, labels ...string) {
	ch <- prometheus.MustNewConstMetric(m.Desc, m.Type, value, labels...)
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

func fetchHTTP(uri string, sslVerify bool, timeout time.Duration) func() (io.ReadCloser, error) {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: !sslVerify}}
	client := http.Client{
		Timeout:   timeout,
		Transport: tr,
	}

	return func() (io.ReadCloser, error) {
		resp, err := client.Get(uri)
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
func NewExporter(uri string, sslVerify bool, timeout time.Duration, logger log.Logger) (*Exporter, error) {
	_, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	var fetchStat func() (io.ReadCloser, error)
	fetchStat = fetchHTTP(uri, sslVerify, timeout)

	return &Exporter{
		URI:       uri,
		fetchStat: fetchStat,
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
		listenAddress = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9101").String()
		metricsPath   = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
		wakaTimeout   = kingpin.Flag("wakatime.timeout", "Timeout for trying to get stats from wakatime.").Default("5s").Duration()

		wakaScrapeURI = "https://localhost:9090"
		wakaSSLVerify = true
	)

	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print("wakatime_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promlog.New(promlogConfig)

	level.Info(logger).Log("msg", "Starting wakatime_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "context", version.BuildContext())

	exporter, err := NewExporter(wakaScrapeURI, wakaSSLVerify, *wakaTimeout, logger)
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
