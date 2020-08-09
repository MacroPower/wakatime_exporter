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
	"io"
	"net/url"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

type metrics map[string]metricInfo

// Exporter collects stats from the given URI and exports them using
// the prometheus metrics package.
type Exporter struct {
	URI       *url.URL
	user      string
	mutex     sync.RWMutex
	fetchStat func(url.URL, string, string) (io.ReadCloser, error)
	tzOffset  time.Duration

	up                          prometheus.Gauge
	totalScrapes, queryFailures prometheus.Counter
	logger                      log.Logger
}

type metricInfo struct {
	Desc *prometheus.Desc
	Type prometheus.ValueType
}

// WakatimeTime types for Wakatime
type WakatimeTime struct {
	Data struct {
		IsUpToDate   bool    `json:"is_up_to_date"`
		Text         string  `json:"text"`
		TotalSeconds float64 `json:"total_seconds"`
	} `json:"data"`
}

// WakatimeSummary types for Wakatime
type WakatimeSummary struct {
	Data []struct {
		Categories []struct {
			Digital      string  `json:"digital"`
			Hours        int     `json:"hours"`
			Minutes      int     `json:"minutes"`
			Name         string  `json:"name"`
			Percent      float64 `json:"percent"`
			Seconds      int     `json:"seconds"`
			Text         string  `json:"text"`
			TotalSeconds float64 `json:"total_seconds"`
		} `json:"categories"`
		Dependencies []struct {
			Digital      string  `json:"digital"`
			Hours        int     `json:"hours"`
			Minutes      int     `json:"minutes"`
			Name         string  `json:"name"`
			Percent      float64 `json:"percent"`
			Seconds      int     `json:"seconds"`
			Text         string  `json:"text"`
			TotalSeconds float64 `json:"total_seconds"`
		} `json:"dependencies"`
		Editors []struct {
			Digital      string  `json:"digital"`
			Hours        int     `json:"hours"`
			Minutes      int     `json:"minutes"`
			Name         string  `json:"name"`
			Percent      float64 `json:"percent"`
			Seconds      int     `json:"seconds"`
			Text         string  `json:"text"`
			TotalSeconds float64 `json:"total_seconds"`
		} `json:"editors"`
		GrandTotal struct {
			Digital      string  `json:"digital"`
			Hours        int     `json:"hours"`
			Minutes      int     `json:"minutes"`
			Text         string  `json:"text"`
			TotalSeconds float64 `json:"total_seconds"`
		} `json:"grand_total"`
		Languages []struct {
			Digital      string  `json:"digital"`
			Hours        int     `json:"hours"`
			Minutes      int     `json:"minutes"`
			Name         string  `json:"name"`
			Percent      float64 `json:"percent"`
			Seconds      int     `json:"seconds"`
			Text         string  `json:"text"`
			TotalSeconds float64 `json:"total_seconds"`
		} `json:"languages"`
		Machines []struct {
			Digital       string  `json:"digital"`
			Hours         int     `json:"hours"`
			MachineNameID string  `json:"machine_name_id"`
			Minutes       int     `json:"minutes"`
			Name          string  `json:"name"`
			Percent       float64 `json:"percent"`
			Seconds       int     `json:"seconds"`
			Text          string  `json:"text"`
			TotalSeconds  float64 `json:"total_seconds"`
		} `json:"machines"`
		OperatingSystems []struct {
			Digital      string  `json:"digital"`
			Hours        int     `json:"hours"`
			Minutes      int     `json:"minutes"`
			Name         string  `json:"name"`
			Percent      float64 `json:"percent"`
			Seconds      int     `json:"seconds"`
			Text         string  `json:"text"`
			TotalSeconds float64 `json:"total_seconds"`
		} `json:"operating_systems"`
		Projects []struct {
			Digital      string  `json:"digital"`
			Hours        int     `json:"hours"`
			Minutes      int     `json:"minutes"`
			Name         string  `json:"name"`
			Percent      float64 `json:"percent"`
			Seconds      int     `json:"seconds"`
			Text         string  `json:"text"`
			TotalSeconds float64 `json:"total_seconds"`
		} `json:"projects"`
		Range struct {
			Date     string    `json:"date"`
			End      time.Time `json:"end"`
			Start    time.Time `json:"start"`
			Text     string    `json:"text"`
			Timezone string    `json:"timezone"`
		} `json:"range"`
	} `json:"data"`
	End   time.Time `json:"end"`
	Start time.Time `json:"start"`
}

// WakatimeGoal types for Wakatime
type WakatimeGoal struct {
	Data []struct {
		AverageStatus string `json:"average_status"`
		ChartData     []struct {
			ActualSeconds     float64 `json:"actual_seconds"`
			ActualSecondsText string  `json:"actual_seconds_text"`
			GoalSeconds       int     `json:"goal_seconds"`
			GoalSecondsText   string  `json:"goal_seconds_text"`
			Range             struct {
				Date     string    `json:"date"`
				End      time.Time `json:"end"`
				Start    time.Time `json:"start"`
				Text     string    `json:"text"`
				Timezone string    `json:"timezone"`
			} `json:"range"`
			RangeStatus            string `json:"range_status"`
			RangeStatusReason      string `json:"range_status_reason"`
			RangeStatusReasonShort string `json:"range_status_reason_short"`
		} `json:"chart_data"`
		CreatedAt          time.Time     `json:"created_at"`
		CumulativeStatus   string        `json:"cumulative_status"`
		Delta              string        `json:"delta"`
		Editors            []interface{} `json:"editors"`
		ID                 string        `json:"id"`
		IgnoreDays         []interface{} `json:"ignore_days"`
		IgnoreZeroDays     bool          `json:"ignore_zero_days"`
		ImproveByPercent   interface{}   `json:"improve_by_percent"`
		IsCurrentUserOwner bool          `json:"is_current_user_owner"`
		IsEnabled          bool          `json:"is_enabled"`
		IsInverse          bool          `json:"is_inverse"`
		IsSnoozed          bool          `json:"is_snoozed"`
		IsTweeting         bool          `json:"is_tweeting"`
		Languages          []string      `json:"languages"`
		ModifiedAt         interface{}   `json:"modified_at"`
		Owner              struct {
			DisplayName string      `json:"display_name"`
			Email       interface{} `json:"email"`
			FullName    string      `json:"full_name"`
			ID          string      `json:"id"`
			Photo       string      `json:"photo"`
			Username    string      `json:"username"`
		} `json:"owner"`
		Projects    []interface{} `json:"projects"`
		RangeText   string        `json:"range_text"`
		Seconds     int           `json:"seconds"`
		SharedWith  []interface{} `json:"shared_with"`
		SnoozeUntil interface{}   `json:"snooze_until"`
		Status      string        `json:"status"`
		Subscribers []struct {
			DisplayName    string      `json:"display_name"`
			Email          interface{} `json:"email"`
			EmailFrequency string      `json:"email_frequency"`
			FullName       string      `json:"full_name"`
			UserID         string      `json:"user_id"`
			Username       string      `json:"username"`
		} `json:"subscribers"`
		Title string `json:"title"`
		Type  string `json:"type"`
	} `json:"data"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// WakatimeLeader types for Wakatime
type WakatimeLeader struct {
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
