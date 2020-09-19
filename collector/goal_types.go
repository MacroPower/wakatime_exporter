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

type wakatimeGoal struct {
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
