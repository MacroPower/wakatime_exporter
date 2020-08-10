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
	b64 "encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

func newWakaMetric(metricName string, docString string, t prometheus.ValueType, variableLabels []string, constLabels prometheus.Labels) metricInfo {
	return metricInfo{
		Desc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", metricName),
			docString,
			variableLabels,
			constLabels,
		),
		Type: t,
	}
}

func b2str(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func (e *Exporter) exportMetric(m metricInfo, ch chan<- prometheus.Metric, value float64, labels ...string) {
	ch <- prometheus.MustNewConstMetric(m.Desc, m.Type, value, labels...)
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
