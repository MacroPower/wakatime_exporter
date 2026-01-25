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
	"time"
)

type wakatimeStats struct {
	Data struct {
		BestDay struct {
			CreatedAt    time.Time `json:"created_at"`
			Date         string    `json:"date"`
			ID           string    `json:"id"`
			ModifiedAt   time.Time `json:"modified_at"`
			Text         string    `json:"text"`
			TotalSeconds float64   `json:"total_seconds"`
		} `json:"best_day"`
		Categories []statsItem `json:"categories"`
		CreatedAt  time.Time   `json:"created_at"`
		DailyAverage float64   `json:"daily_average"`
		DailyAverageIncludingOtherLanguage float64 `json:"daily_average_including_other_language"`
		DaysIncludingHolidays int `json:"days_including_holidays"`
		DaysMinusHolidays     int `json:"days_minus_holidays"`
		Dependencies []statsItem `json:"dependencies"`
		Editors      []statsItem `json:"editors"`
		End          time.Time   `json:"end"`
		HumanReadableDailyAverage string `json:"human_readable_daily_average"`
		HumanReadableDailyAverageIncludingOtherLanguage string `json:"human_readable_daily_average_including_other_language"`
		HumanReadableRange       string `json:"human_readable_range"`
		HumanReadableTotal       string `json:"human_readable_total"`
		HumanReadableTotalIncludingOtherLanguage string `json:"human_readable_total_including_other_language"`
		ID               string    `json:"id"`
		IsAlreadyUpdating bool     `json:"is_already_updating"`
		IsCodingActivityVisible bool `json:"is_coding_activity_visible"`
		IsIncludingToday bool      `json:"is_including_today"`
		IsOtherUsageVisible bool   `json:"is_other_usage_visible"`
		IsStuck          bool      `json:"is_stuck"`
		IsUpToDate       bool      `json:"is_up_to_date"`
		Languages        []statsItem `json:"languages"`
		Machines         []statsItemWithID `json:"machines"`
		ModifiedAt       time.Time `json:"modified_at"`
		OperatingSystems []statsItem `json:"operating_systems"`
		PercentCalculated int      `json:"percent_calculated"`
		Projects         []statsItem `json:"projects"`
		Range            string    `json:"range"`
		Start            time.Time `json:"start"`
		Status           string    `json:"status"`
		Timeout          int       `json:"timeout"`
		Timezone         string    `json:"timezone"`
		TotalSeconds     float64   `json:"total_seconds"`
		TotalSecondsIncludingOtherLanguage float64 `json:"total_seconds_including_other_language"`
		UserID           string    `json:"user_id"`
		Username         string    `json:"username"`
		WritesOnly       bool      `json:"writes_only"`
	} `json:"data"`
}

type statsItem struct {
	Decimal      string  `json:"decimal"`
	Digital      string  `json:"digital"`
	Hours        int     `json:"hours"`
	Minutes      int     `json:"minutes"`
	Name         string  `json:"name"`
	Percent      float64 `json:"percent"`
	Seconds      int     `json:"seconds"`
	Text         string  `json:"text"`
	TotalSeconds float64 `json:"total_seconds"`
}

type statsItemWithID struct {
	Decimal       string  `json:"decimal"`
	Digital       string  `json:"digital"`
	Hours         int     `json:"hours"`
	MachineNameID string  `json:"machine_name_id"`
	Minutes       int     `json:"minutes"`
	Name          string  `json:"name"`
	Percent       float64 `json:"percent"`
	Seconds       int     `json:"seconds"`
	Text          string  `json:"text"`
	TotalSeconds  float64 `json:"total_seconds"`
}
