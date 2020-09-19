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
