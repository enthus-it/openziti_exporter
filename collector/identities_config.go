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
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"golang.org/x/exp/slices"
)

var (
	validIdentityTypes = []string{"default", "router"}
	zitiMgtAPI         = kingpin.Flag(
		"ziti.mgt.api", "Ziti Management API.",
	).Envar("ZITI_MGMT_API").Default("https://localhost:1281").String()
	zitiAdminPassword = kingpin.Flag(
		"ziti.admin.password", "Ziti Management Admin password.",
	).Short('p').Envar("ZITI_ADMIN_PASSWORD").Default("admin123").String()
	zitiAdminUsername = kingpin.Flag(
		"ziti.admin.username", "Ziti Management Admin username.",
	).Short('u').Envar("ZITI_ADMIN_USER").Default("admin").String()
	zitiIdentityTypes = kingpin.Flag(
		"ziti.identity.types", "Ziti Identity Types comma-separated filter.",
	).Envar("ZITI_IDENTITY_TYPES").Default(strings.Join(validIdentityTypes, ",")).String()
	zitiIdentityRoleAttributes = kingpin.Flag(
		"ziti.identity.role.attributes", "Ziti Identity Role Attributes comma-separated filter.",
	).Envar("ZITI_IDENTITY_ROLE_ATTRIBUTES").Default("").String()
)

// getIdentityTypesFilter validate Ziti Identity Types.
func getIdentityTypesFilter() ([]string, error) {
	var validFilterIdentType []string

	for _, value := range strings.Split(*zitiIdentityTypes, ",") {
		if slices.Contains(validIdentityTypes, value) {
			validFilterIdentType = append(validFilterIdentType, strings.ToLower(value))
		} else {
			return validFilterIdentType, fmt.Errorf("%v identity type not valid. Valid values are %v", value, strings.Join(validIdentityTypes, ","))
		}
	}

	return validFilterIdentType, nil
}
