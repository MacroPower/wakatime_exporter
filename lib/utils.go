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

package lib

import (
	"crypto/tls"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "wakatime"
)

// NewWakaMetric creates a MetricInfo struct containing metric Desc and Type
func NewWakaMetric(metricName string, docString string, t prometheus.ValueType, variableLabels []string, constLabels prometheus.Labels) MetricInfo {
	return MetricInfo{
		Desc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", metricName),
			docString,
			variableLabels,
			constLabels,
		),
		Type: t,
	}
}

// BoolToBinary converts booleans to "0" or "1"
func BoolToBinary(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// FetchHTTP is a generic fetch method for Wakatime API endpoints
func FetchHTTP(token string, sslVerify bool, timeout time.Duration, logger log.Logger) func(uri url.URL, subPath string, params url.Values) (io.ReadCloser, error) {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: !sslVerify}}
	client := http.Client{
		Timeout:   timeout,
		Transport: tr,
	}
	sEnc := b64.StdEncoding.EncodeToString([]byte(token))
	return func(uri url.URL, subPath string, params url.Values) (io.ReadCloser, error) {

		uri.Path = path.Join(uri.Path, subPath)
		uri.RawQuery = params.Encode()
		url := uri.String()

		level.Info(logger).Log("msg", "Scraping Wakatime", "path", subPath, "url", url)

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

// ReadAndUnmarshal reads the JSON response body and unmarshals the response
func ReadAndUnmarshal(body io.ReadCloser, object interface{}) error {
	respBody, readErr := ioutil.ReadAll(body)

	if readErr != nil {
		return readErr
	}

	var jsonErr error
	jsonErr = json.Unmarshal(respBody, &object)
	if jsonErr != nil {
		return jsonErr
	}

	return nil
}

// UserPath appends the User path to a given URL
func UserPath(uri *url.URL, user string) url.URL {
	userURL := *uri
	userPath := path.Join(userURL.Path, "users", user)

	userURL.Path = userPath
	return userURL
}
