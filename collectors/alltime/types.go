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
	"io"
	"net/url"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

type exporterAlltime struct {
	URI       *url.URL
	endpoint  string
	user      string
	mutex     sync.RWMutex
	fetchStat func(url.URL, string, string) (io.ReadCloser, error)
	tzOffset  time.Duration

	up                          prometheus.Gauge
	totalScrapes, queryFailures prometheus.Counter
	logger                      log.Logger
}

type wakatimeAlltime struct {
	Data struct {
		IsUpToDate   bool    `json:"is_up_to_date"`
		Text         string  `json:"text"`
		TotalSeconds float64 `json:"total_seconds"`
	} `json:"data"`
}
