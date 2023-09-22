// Copyright 2023 enthus GmbH
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

type FabricLinks struct {
	Data []FabricLink `json:"data"`
	Meta MetaData     `json:"meta"`
}

// FabricLink represent the meaningful chracteristics of a Ziti Fabric Link
// for this exporter
type FabricLink struct {
	Cost        float64 `json:"cost"`
	DestLatency float64 `json:"destLatency"`
	DestRouter  struct {
		Name string `json:"name"`
	} `json:"destRouter"`
	Down          bool    `json:"down"`
	ID            string  `json:"id"`
	Protocol      string  `json:"protocol"`
	SourceLatency float64 `json:"sourceLatency"`
	SourceRouter  struct {
		Name string `json:"name"`
	} `json:"sourceRouter"`
	State      string  `json:"state"`
	StaticCost float64 `json:"staticCost"`
}
