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
