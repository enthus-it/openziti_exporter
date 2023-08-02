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

type Identities struct {
	Data []Identity `json:"data"`
	Meta MetaData   `json:"meta"`
}

// Identity represent the meaningful chracteristics of a Ziti Identity
// for this exporter
type Identity struct {
	CreatedAt               string   `json:"createdAt"`
	UpdatedAt               string   `json:"updatedAt"`
	Disabled                bool     `json:"disabled"`
	HasAPISession           bool     `json:"hasApiSession"`
	HasEdgeRouterConnection bool     `json:"hasEdgeRouterConnection"`
	Name                    string   `json:"name"`
	RoleAttributes          []string `json:"roleAttributes"`
	SdkInfo                 struct {
		Type    string `json:"type"`
		Version string `json:"version"`
	} `json:"sdkInfo"`
	TypeID string `json:"typeId"`
}

// MetaData represent the pagination part of a Ziti Identity call
type MetaData struct {
	Pagination struct {
		TotalCount int `json:"totalCount"`
	} `json:"pagination"`
}
