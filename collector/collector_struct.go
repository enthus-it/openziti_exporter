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

import (
	"github.com/go-kit/log"
	"github.com/openziti/ziti/ziti/cmd/api"
)

// LoginOptions are the flags for login commands
type LoginOptions struct {
	api.Options
	Username        string
	Password        string
	Host            string
	HostReady       string
	Token           string
	Logger          log.Logger
	CaCert          string
	ReadOnly        bool
	Yes             bool
	IgnoreConfig    bool
	ClientCert      string
	ClientKey       string
	ExtJwt          string
	IdentTypeFilter []string
}

type LoginSession struct {
	Data LoginData `json:"data"`
}

// LoginData represent the meaningful chracteristics of a Ziti Authentication endpoint
type LoginData struct {
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	ExpiresAt string `json:"expiresAt"`
	Token     string `json:"token"`
}
