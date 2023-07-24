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
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/openziti/foundation/v2/term"
	"github.com/openziti/ziti/ziti/cmd/api"
	"github.com/openziti/ziti/ziti/cmd/common"
	"github.com/openziti/ziti/ziti/util"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/exp/slices"
)

type identitiesCollector struct {
	logger  log.Logger
	options *LoginOptions
}

func init() {
	registerCollector("identities", defaultEnabled, newIdentitiesCollector)
}

// newIdentitiesCollector returns a new Collector exposing OpenZiti Identities metrics.
func newIdentitiesCollector(logger log.Logger) (Collector, error) {
	return &identitiesCollector{
		logger: logger,
	}, nil
}

// edgeAPILogin returns a session token from edge/management/v1.
func edgeAPILogin(logger log.Logger) (*LoginOptions, error) {
	identityTypeFilter, err := getIdentityTypesFilter()
	if err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}

	options := &LoginOptions{
		Options: api.Options{
			CommonOptions:      common.CommonOptions{BatchMode: true},
			OutputJSONResponse: true,
		},
		Username:        *zitiAdminUsername,
		Password:        *zitiAdminPassword,
		Host:            *zitiMgtAPI,
		ReadOnly:        true,
		Logger:          logger,
		IdentTypeFilter: identityTypeFilter,
	}
	err = options.RunLogin()

	return options, err
}

// Update pushes identities metrics onto ch
func (c *identitiesCollector) Update(ch chan<- prometheus.Metric) (err error) {
	// if not already logged, do the login.
	if c.options == nil {
		c.options, err = edgeAPILogin(c.logger)
		if err != nil {
			return err
		}

		level.Debug(c.logger).Log("msg", "Login", "ztToken", c.options.Token)
	} else if c.options.Token == "" {
		c.options, err = edgeAPILogin(c.logger)
		if err != nil {
			return err
		}

		level.Debug(c.logger).Log("msg", "Login", "ztToken", c.options.Token)
	}

	identities, err := c.options.RunIdentities()
	if err != nil {
		return err
	}

	for i := range identities.Data {
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "", "identity_has_api_session"),
				"Identity has an API session active.",
				[]string{"name", "type"}, nil,
			), prometheus.GaugeValue,
			convertBool2Float(identities.Data[i].HasAPISession),
			identities.Data[i].Name,
			identities.Data[i].TypeID,
		)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "", "identity_has_edge_router_connection"),
				"Identity has an edge router connection active.",
				[]string{"name", "type"}, nil,
			), prometheus.GaugeValue,
			convertBool2Float(identities.Data[i].HasEdgeRouterConnection),
			identities.Data[i].Name,
			identities.Data[i].TypeID,
		)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "", "identity_last_update_timestamp_seconds"),
				"Identity last update timestamp.",
				[]string{"name", "type"}, nil,
			), prometheus.GaugeValue,
			convertRFC33339toUnix(identities.Data[i].UpdatedAt),
			identities.Data[i].Name,
			identities.Data[i].TypeID,
		)
	}

	return nil
}

// RunLogin implements this command
func (o *LoginOptions) RunLogin() error {
	host := o.Host
	if !strings.HasPrefix(host, "http") {
		host = "https://" + host
	}

	ctrlURL, err := url.Parse(host)
	if err != nil {
		return errors.Wrap(err, "invalid controller URL")
	}

	host = ctrlURL.Scheme + "://" + ctrlURL.Host

	if err := o.ConfigureCerts(host, ctrlURL); err != nil {
		return err
	}

	if o.CaCert != "" {
		if certAbs, err := filepath.Abs(o.CaCert); err == nil {
			o.CaCert = certAbs
		}
	}

	if ctrlURL.Path == "" {
		host = util.EdgeControllerGetManagementApiBasePath(host, o.CaCert)
	} else {
		host += ctrlURL.Path
	}

	o.HostReady = host
	body := "{}"

	if o.Token == "" && o.ClientCert == "" && o.ExtJwt == "" {
		for o.Username == "" {
			if o.Username, err = term.Prompt("Enter username: "); err != nil {
				return err
			}
		}

		if o.Password == "" {
			if o.Password, err = term.PromptPassword("Enter password: ", false); err != nil {
				return err
			}
		}

		container := gabs.New()
		_, _ = container.SetP(o.Username, "username")
		_, _ = container.SetP(o.Password, "password")

		body = container.String()
	}

	level.Debug(o.Logger).Log("msg", "Login", "options", o, "host", o.HostReady)
	jsonParsed, err := login(o, body)

	if err != nil {
		return err
	}

	if !jsonParsed.ExistsP("data.token") {
		return fmt.Errorf("no session token returned from login request to %v. Received: %v", o.HostReady, jsonParsed.String())
	}

	var ok bool
	o.Token, ok = jsonParsed.Path("data.token").Data().(string)

	if !ok {
		return fmt.Errorf("session token returned from login request to %v is not in the expected format. Received: %v", o.HostReady, jsonParsed.String())
	}

	if !o.OutputJSONResponse {
		level.Debug(o.Logger).Log("msg", "Token", "token", o.Token)
	}

	return err
}

// RunIdentities implements this command
func (o *LoginOptions) RunIdentities() (Identities, error) {
	var (
		limit                         = 20
		offset                        = 0
		identStructTotal, identStruct Identities
	)

	jsonParsed, err := identities(o, limit, offset)
	if err != nil {
		return identStructTotal, err
	}

	err = json.Unmarshal(jsonParsed.Bytes(), &identStruct)
	if err != nil {
		return identStructTotal, err
	}

	for i := range identStruct.Data {
		if slices.Contains(o.IdentTypeFilter, strings.ToLower(identStruct.Data[i].TypeID)) &&
			containsIdentRoleAttr(identStruct.Data[i].RoleAttributes) {
			identStructTotal.Data = append(identStructTotal.Data, identStruct.Data[i])
		}
	}

	totalIdentityCount, ok := jsonParsed.Path("meta.pagination.totalCount").Data().(float64)
	if !ok {
		return identStructTotal, fmt.Errorf("error returned from parsing totalidentitiescount %v. Received: %v", o.HostReady, jsonParsed.String())
	}

	level.Debug(o.Logger).Log("msg", "Total Ziti Identities found", "count", totalIdentityCount)

	for offset+limit < int(totalIdentityCount) {
		offset += limit

		jsonParsed, err := identities(o, limit, offset)
		if err != nil {
			return identStructTotal, err
		}

		err = json.Unmarshal(jsonParsed.Bytes(), &identStruct)
		if err != nil {
			return identStructTotal, err
		}

		for i := range identStruct.Data {
			if slices.Contains(o.IdentTypeFilter,
				strings.ToLower(identStruct.Data[i].TypeID)) &&
				containsIdentRoleAttr(identStruct.Data[i].RoleAttributes) {
				identStructTotal.Data = append(identStructTotal.Data, identStruct.Data[i])
			}
		}
	}

	return identStructTotal, err
}

// containsIdentRoleAttr compare Identity RoleAttributes slices with the one from command-line.
func containsIdentRoleAttr(roleAttr []string) bool {
	// if no filter was passed via the filter, return true
	if *zitiIdentityRoleAttributes == "" {
		return true
	} else if len(roleAttr) == 0 {
		return false
	}

	zitiIdentityRoleAttributesFilter := strings.Split(*zitiIdentityRoleAttributes, ",")
	for i := range roleAttr {
		if slices.Contains(zitiIdentityRoleAttributesFilter, roleAttr[i]) {
			return true
		}
	}

	return false
}

func (o *LoginOptions) ConfigureCerts(host string, ctrlURL *url.URL) error {
	isServerTrusted, err := util.IsServerTrusted(host)
	if err != nil {
		return err
	}

	if !isServerTrusted && o.CaCert == "" {
		wellKnownCerts, certs, err := util.GetWellKnownCerts(host)
		if err != nil {
			return errors.Wrapf(err, "unable to retrieve server certificate authority from %v", host)
		}

		certsTrusted, err := util.AreCertsTrusted(host, wellKnownCerts)
		if err != nil {
			return err
		}

		if !certsTrusted {
			return errors.New("server supplied certs not trusted by server, unable to continue")
		}

		savedCerts, certFile, err := util.ReadCert(ctrlURL.Hostname())
		if err != nil {
			return err
		}

		if savedCerts != nil {
			o.CaCert = certFile
			if !util.AreCertsSame(o, wellKnownCerts, savedCerts) {
				o.Printf("WARNING: server supplied certificate authority doesn't match cached certs at %v\n", certFile)

				replace := o.Yes
				if !replace {
					if replace, err = o.askYesNo("Replace cached certs [Y/N]: "); err != nil {
						return err
					}
				}

				if replace {
					_, err = util.WriteCert(o, ctrlURL.Hostname(), wellKnownCerts)
					if err != nil {
						return err
					}
				}
			}
		} else {
			o.Printf("Untrusted certificate authority retrieved from server\n")
			o.Println("Verified that server supplied certificates are trusted by server")
			o.Printf("Server supplied %v certificates\n", len(certs))
			importCerts := o.Yes
			if !importCerts {
				if importCerts, err = o.askYesNo("Trust server provided certificate authority [Y/N]: "); err != nil {
					return err
				}
			}
			if importCerts {
				o.CaCert, err = util.WriteCert(o, ctrlURL.Hostname(), wellKnownCerts)
				if err != nil {
					return err
				}
			} else {
				o.Println("WARNING: no certificate authority provided for server, continuing but login will likely fail")
			}
		}
	} else if isServerTrusted && o.CaCert != "" {
		override, err := o.askYesNo("Server certificate authority is already trusted. Are you sure you want to provide an additional CA [Y/N]: ")
		if err != nil {
			return err
		}
		if !override {
			o.CaCert = ""
		}
	}

	return nil
}

func (o *LoginOptions) askYesNo(prompt string) (bool, error) {
	filter := &yesNoFilter{}
	if _, err := o.ask(prompt, filter.Accept); err != nil {
		return false, err
	}

	return filter.result, nil
}

func (o *LoginOptions) ask(prompt string, f func(string) bool) (string, error) {
	for {
		val, err := term.Prompt(prompt)
		if err != nil {
			return "", err
		}

		val = strings.TrimSpace(val)
		if f(val) {
			return val, nil
		}

		o.Printf("Invalid input: %v\n", val)
	}
}

type yesNoFilter struct {
	result bool
}

func (filter *yesNoFilter) Accept(s string) bool {
	if strings.EqualFold("y", s) || strings.EqualFold("yes", s) {
		filter.result = true
		return true
	}

	if strings.EqualFold("n", s) || strings.EqualFold("no", s) {
		filter.result = false
		return true
	}

	return false
}

// EdgeControllerLogin will authenticate to the given Edge Controller
func login(o *LoginOptions, authentication string) (*gabs.Container, error) {
	client := util.NewClient()
	timeout := o.Timeout
	verbose := o.Verbose
	method := "password"

	cert := o.CaCert
	if cert != "" {
		client.SetRootCertificate(cert)
	}

	resp, err := client.
		SetTimeout(time.Duration(timeout)*time.Second).
		SetDebug(verbose).
		R().
		SetQueryParam("method", method).
		SetHeader("Content-Type", "application/json").
		SetBody(authentication).
		Post(o.HostReady + "/authenticate")

	if err != nil {
		return nil, fmt.Errorf("unable to authenticate to %v. Error: %v", o.HostReady, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("unable to authenticate to %v. Status code: %v, Server returned: %v", o.HostReady, resp.Status(), util.PrettyPrintResponse(resp))
	}

	jsonParsed, err := gabs.ParseJSON(resp.Body())
	if err != nil {
		return nil, fmt.Errorf("unable to parse response from %v. Server returned: %v", o.HostReady, resp.String())
	}

	return jsonParsed, nil
}

// EdgeControllerIdentities will return all available identities
// request.SetHeaderParam("zt-session", e.Token)
func identities(o *LoginOptions, limit, offset int) (*gabs.Container, error) {
	client := util.NewClient()
	timeout := o.Timeout
	verbose := o.Verbose

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
		Get(o.HostReady + "/identities")

	if err != nil {
		// reset login token to force a new login
		o.Token = ""
		return nil, fmt.Errorf("unable to authenticate to %v. Error: %v", o.HostReady, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("unable to authenticate to %v. Status code: %v, Server returned: %v", o.HostReady, resp.Status(), util.PrettyPrintResponse(resp))
	}

	jsonParsed, err := gabs.ParseJSON(resp.Body())
	if err != nil {
		return nil, fmt.Errorf("unable to parse response from %v. Server returned: %v", o.HostReady, resp.String())
	}

	return jsonParsed, nil
}
