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

type wakatimeDurations struct {
	Data []struct {
		Project        string  `json:"project"`
		Time           float64 `json:"time"`
		Duration       float64 `json:"duration"`
		Color          string  `json:"color"`
		Category       string  `json:"category"`
		Language       string  `json:"language"`
		AIAdditions    int     `json:"ai_additions"`
		AIDeletions    int     `json:"ai_deletions"`
		HumanAdditions int     `json:"human_additions"`
		HumanDeletions int     `json:"human_deletions"`
	} `json:"data"`
	Branches []struct {
		Name string `json:"name"`
	} `json:"branches"`
	Start    string `json:"start"`
	End      string `json:"end"`
	Timezone string `json:"timezone"`
}
