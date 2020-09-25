/*
Copyright 2020 Jacob Colvin (MacroPower)
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

package collector

import (
	"crypto/tls"
	"encoding/base64"
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
)

// Namespace defines the common namespace to be used by all metrics.
const namespace = "wakatime"

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
	sEnc := base64.StdEncoding.EncodeToString([]byte(token))
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
