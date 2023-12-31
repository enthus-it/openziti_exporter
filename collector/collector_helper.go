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
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-kit/log/level"
	"github.com/openziti/ziti/ziti/util"
)

// controllerAPICall will return a API call response
// request.SetHeaderParam("zt-session", e.Token)
func controllerAPICall(o *LoginOptions, api, endpoint string, limit, offset int) ([]byte, error) {
	client := util.NewClient()
	timeout := o.Timeout
	verbose := o.Verbose
	hostReady := ""

	switch api {
	case "edge_management":
		hostReady = o.HostReadyEdgeManagementAPI
	case "fabric":
		hostReady = o.HostReadyFabricAPI
	default:
		return nil, fmt.Errorf("API not implemented %v", api)
	}

	cert := o.CaCert
	if cert != "" {
		client.SetRootCertificate(cert)
	}

	resp, err := client.
		SetTimeout(time.Duration(timeout)*time.Second).
		SetDebug(verbose).
		R().
		SetQueryParam("limit", strconv.Itoa(limit)).
		SetQueryParam("offset", strconv.Itoa(offset)).
		SetHeader("Content-Type", "application/json").
		SetHeader("zt-session", o.Token).
		Get(hostReady + endpoint)

	if err != nil {
		// reset login token to force a new login
		o.Token = ""
		return nil, fmt.Errorf("unable to authenticate to %v. Error: %v", hostReady, err)
	}

	if resp.StatusCode() != http.StatusOK {
		if resp.StatusCode() == http.StatusUnauthorized {
			// reset login token to force a new login
			o.Token = ""
		}

		return nil, fmt.Errorf("unable to authenticate to %v. Status code: %v, Server returned: %v", hostReady, resp.Status(), util.PrettyPrintResponse(resp))
	}

	return resp.Body(), nil
}

func (o *LoginOptions) WriteCert(id string, cert []byte) (string, error) {
	const rwx, rw = 0o700, 0o600

	cfgDir, err := util.ConfigDir()
	if err != nil {
		return "", err
	}

	certsDir := filepath.Join(cfgDir, "certs")
	if err = os.MkdirAll(certsDir, rwx); err != nil {
		return "", fmt.Errorf("unable to create ziti certs dir %v: %w", certsDir, err)
	}

	certFile := filepath.Join(certsDir, id)
	if err := os.WriteFile(certFile, cert, rw); err != nil {
		return "", err
	}

	level.Info(o.Logger).Log("msg", "server certificate chain written", "cert_file", certFile)

	return certFile, nil
}
