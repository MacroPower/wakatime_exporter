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

package leader

import (
	"io"
	"net/url"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

type exporterLeader struct {
	URI       *url.URL
	endpoint  string
	mutex     sync.RWMutex
	fetchStat func(url.URL, string, string) (io.ReadCloser, error)
	tzOffset  time.Duration

	up                          prometheus.Gauge
	totalScrapes, queryFailures prometheus.Counter
	logger                      log.Logger
}

type wakatimeLeader struct {
	CurrentUser struct {
		Rank         int `json:"rank"`
		RunningTotal struct {
			DailyAverage              int    `json:"daily_average"`
			HumanReadableDailyAverage string `json:"human_readable_daily_average"`
			HumanReadableTotal        string `json:"human_readable_total"`
			Languages                 []struct {
				Name         string  `json:"name"`
				TotalSeconds float64 `json:"total_seconds"`
			} `json:"languages"`
			ModifiedAt   time.Time `json:"modified_at"`
			TotalSeconds float64   `json:"total_seconds"`
		} `json:"running_total"`
		User struct {
			FullName             string `json:"full_name"`
			HumanReadableWebsite string `json:"human_readable_website"`
			ID                   string `json:"id"`
			IsHireable           bool   `json:"is_hireable"`
			Location             string `json:"location"`
			Photo                string `json:"photo"`
			Username             string `json:"username"`
			Website              string `json:"website"`
		} `json:"user"`
	} `json:"current_user"`
	Data []struct {
		Rank         int `json:"rank"`
		RunningTotal struct {
			DailyAverage              int    `json:"daily_average"`
			HumanReadableDailyAverage string `json:"human_readable_daily_average"`
			HumanReadableTotal        string `json:"human_readable_total"`
			Languages                 []struct {
				Name         string  `json:"name"`
				TotalSeconds float64 `json:"total_seconds"`
			} `json:"languages"`
			TotalSeconds float64 `json:"total_seconds"`
		} `json:"running_total"`
		User struct {
			DisplayName          string `json:"display_name"`
			Email                string `json:"email"`
			FullName             string `json:"full_name"`
			HumanReadableWebsite string `json:"human_readable_website"`
			ID                   string `json:"id"`
			IsEmailPublic        bool   `json:"is_email_public"`
			IsHireable           bool   `json:"is_hireable"`
			Location             string `json:"location"`
			Photo                string `json:"photo"`
			PhotoPublic          bool   `json:"photo_public"`
			Username             string `json:"username"`
			Website              string `json:"website"`
		} `json:"user"`
	} `json:"data"`
	Language   interface{} `json:"language"`
	ModifiedAt time.Time   `json:"modified_at"`
	Page       int         `json:"page"`
	Range      struct {
		EndDate   string `json:"end_date"`
		EndText   string `json:"end_text"`
		Name      string `json:"name"`
		StartDate string `json:"start_date"`
		StartText string `json:"start_text"`
		Text      string `json:"text"`
	} `json:"range"`
	Timeout    int  `json:"timeout"`
	TotalPages int  `json:"total_pages"`
	WritesOnly bool `json:"writes_only"`
}
