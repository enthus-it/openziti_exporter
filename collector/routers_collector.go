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
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	jsoniter "github.com/json-iterator/go"
	"github.com/prometheus/client_golang/prometheus"
)

type routersCollector struct {
	logger  log.Logger
	options *LoginOptions
}

const (
	routerSpace = "router"
)

func init() {
	registerCollector("routers", defaultEnabled, newRoutersCollector)
}

// newRoutersCollector returns a new Collector exposing OpenZiti Routers metrics.
func newRoutersCollector(logger log.Logger, options *LoginOptions) (Collector, error) {
	return &routersCollector{
		logger:  logger,
		options: options,
	}, nil
}

// Update pushes routers metrics onto ch
func (c *routersCollector) Update(ch chan<- prometheus.Metric) (err error) {
	// if not already logged, do the login.
	if c.options == nil {
		c.options, err = edgeAPILogin(c.logger)
		if err != nil {
			errString := fmt.Sprintf("%s", errors.Unwrap(err))
			zitiLoginErrors[errString]++

			return err
		}

		zitiLoginSuccess++
	} else if c.options.Token == "" {
		c.options, err = edgeAPILogin(c.logger)
		if err != nil {
			errString := fmt.Sprintf("%s", errors.Unwrap(err))
			zitiLoginErrors[errString]++

			return err
		}

		zitiLoginSuccess++
	}

	routers, err := c.options.RunRouters()
	if err != nil {
		return err
	}

	for i := range routers.Data {
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, routerSpace,
					"online"),
				"Router is currently online.",
				[]string{"hostname", "role_attributes", "version"}, nil,
			), prometheus.GaugeValue,
			convertBool2Float(routers.Data[i].IsOnline),
			routers.Data[i].Hostname,
			strings.Join(routers.Data[i].RoleAttributes, " "),
			routers.Data[i].VersionInfo.Version,
		)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, routerSpace,
					"enabled"),
				"Router is currently enabled.",
				[]string{"hostname", "role_attributes", "version"}, nil,
			), prometheus.GaugeValue,
			convertBool2Float(!routers.Data[i].Disabled),
			routers.Data[i].Hostname,
			strings.Join(routers.Data[i].RoleAttributes, " "),
			routers.Data[i].VersionInfo.Version,
		)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, routerSpace,
					"tunneler_enabled"),
				"Router as tunneler enabled.",
				[]string{"hostname", "role_attributes", "version"}, nil,
			), prometheus.GaugeValue,
			convertBool2Float(routers.Data[i].IsTunnelerEnabled),
			routers.Data[i].Hostname,
			strings.Join(routers.Data[i].RoleAttributes, " "),
			routers.Data[i].VersionInfo.Version,
		)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, routerSpace,
					"service_traversal_enabled"),
				"Router let services traverse through.",
				[]string{"hostname", "role_attributes", "version"}, nil,
			), prometheus.GaugeValue,
			convertBool2Float(!routers.Data[i].NoTraversal),
			routers.Data[i].Hostname,
			strings.Join(routers.Data[i].RoleAttributes, " "),
			routers.Data[i].VersionInfo.Version,
		)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, routerSpace,
					"sync_status"),
				"Router synchronization status.",
				[]string{"hostname", "role_attributes", "version", "sync_status"}, nil,
			), prometheus.GaugeValue,
			float64(routerStatus(routers.Data[i].SyncStatus)),
			routers.Data[i].Hostname,
			strings.Join(routers.Data[i].RoleAttributes, " "),
			routers.Data[i].VersionInfo.Version,
			routers.Data[i].SyncStatus,
		)
	}

	return nil
}

// RunRouters implements this command
func (o *LoginOptions) RunRouters() (Routers, error) {
	var (
		limit                           = 20
		offset                          = 0
		routerStructTotal, routerStruct Routers
		json                            = jsoniter.ConfigCompatibleWithStandardLibrary
	)

	jsonBytes, err := controllerAPICall(o, "edge_management", "/edge-routers", limit, offset)
	if err != nil {
		return routerStructTotal, err
	}

	err = json.Unmarshal(jsonBytes, &routerStruct)
	if err != nil {
		return routerStructTotal, err
	}

	routerStructTotal.Data = append(routerStructTotal.Data, routerStruct.Data...)

	totalRouterCount := routerStruct.Meta.Pagination.TotalCount
	level.Debug(o.Logger).Log("msg", "Total Ziti Routers found", "count", totalRouterCount)

	for offset+limit < totalRouterCount {
		offset += limit

		jsonBytes, err := controllerAPICall(o, "edge_management", "/edge-routers", limit, offset)
		if err != nil {
			return routerStructTotal, err
		}

		err = json.Unmarshal(jsonBytes, &routerStruct)
		if err != nil {
			return routerStructTotal, err
		}

		routerStructTotal.Data = append(routerStructTotal.Data, routerStruct.Data...)
	}

	return routerStructTotal, err
}

// routerStatus maps the router status with an integer value
func routerStatus(status string) int64 {
	switch status {
	case "SYNC_NEW":
		return syncNew
	case "SYNC_QUEUED":
		return syncQueued
	case "SYNC_HELLO_TIMEOUT":
		return syncHelloTimeout
	case "SYNC_HELLO":
		return syncHello
	case "SYNC_HELLO_WAIT":
		return syncHelloWait
	case "SYNC_RESYNC_WAIT":
		return syncResyncWait
	case "SYNC_IN_PROGRESS":
		return syncInProgress
	case "SYNC_DONE":
		return syncDone
	case "SYNC_UNKNOWN":
		return syncUnknown
	case "SYNC_DISCONNECTED":
		return syncDisconnected
	case "SYNC_ERROR":
		return syncError
	default:
		return math.MaxInt64
	}
}
