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

type Routers struct {
	Data []Router `json:"data"`
	Meta MetaData `json:"meta"`
}

const (
	syncNew          = iota // connection accepted but no strategy actions have been taken
	syncQueued              // connection handed to a strategy and waiting for processing
	syncHelloTimeout        // sync failed due to a hello timeout, requeued for hello
	syncHello               // controller edge hello being sent
	syncHelloWait           // hello received from router and queued for processing
	syncResyncWait          // router requested a resync and queued for processing
	syncInProgress          // synchronization processing
	syncDone                // synchronization completed, router is now in maintenance updates
	syncUnknown             // state is unknown, edge router misbehaved, error state
	syncDisconnected        // strategy was disconnected before finishing, error state
	syncError               // sync failed due to an unexpected error
)

// Router represent the meaningful chracteristics of a Ziti Router
// for this exporter
type Router struct {
	Disabled          bool     `json:"disabled"`
	Hostname          string   `json:"hostname"`
	IsOnline          bool     `json:"isOnline"`
	Name              string   `json:"name"`
	NoTraversal       bool     `json:"noTraversal"`
	SyncStatus        string   `json:"syncStatus"`
	IsTunnelerEnabled bool     `json:"isTunnelerEnabled"`
	IsVerified        bool     `json:"isVerified"`
	RoleAttributes    []string `json:"roleAttributes"`
	VersionInfo       struct {
		Version string `json:"version"`
	} `json:"versionInfo"`
}
