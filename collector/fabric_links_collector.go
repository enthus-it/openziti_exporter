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

	jsoniter "github.com/json-iterator/go"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	nano2seconds     = 1e9
	fabricLinksSpace = "fabric_links"
)

type fabricLinksCollector struct {
	logger  log.Logger
	options *LoginOptions
}

func init() {
	registerCollector("fabric_links", defaultEnabled, newFabricLinksCollector)
}

// newFabricLinksCollector returns a new Collector exposing OpenZiti fabric links metrics.
func newFabricLinksCollector(logger log.Logger, options *LoginOptions) (Collector, error) {
	return &fabricLinksCollector{
		logger:  logger,
		options: options,
	}, nil
}

// Update pushes fabric links metrics onto ch
func (c *fabricLinksCollector) Update(ch chan<- prometheus.Metric) (err error) {
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

	fabricLinks, err := c.options.RunFabricLinks()
	if err != nil {
		return err
	}

	for i := range fabricLinks.Data {
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace,
					fabricLinksSpace, "status"),
				"Fabric Link status. (1: up, 0: down)",
				[]string{"destination", "source", "id"}, nil,
			), prometheus.GaugeValue,
			convertBool2Float(!fabricLinks.Data[i].Down),
			fabricLinks.Data[i].DestRouter.Name,
			fabricLinks.Data[i].SourceRouter.Name,
			fabricLinks.Data[i].ID,
		)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace,
					fabricLinksSpace, "source_latency_seconds"),
				"Fabric Link source latency.",
				[]string{"destination", "source", "id"}, nil,
			), prometheus.GaugeValue,
			fabricLinks.Data[i].SourceLatency/nano2seconds,
			fabricLinks.Data[i].DestRouter.Name,
			fabricLinks.Data[i].SourceRouter.Name,
			fabricLinks.Data[i].ID,
		)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace,
					fabricLinksSpace,
					"destination_latency_seconds"),
				"Fabric Link destination latency.",
				[]string{"destination", "source", "id"}, nil,
			), prometheus.GaugeValue,
			fabricLinks.Data[i].DestLatency/nano2seconds,
			fabricLinks.Data[i].DestRouter.Name,
			fabricLinks.Data[i].SourceRouter.Name,
			fabricLinks.Data[i].ID,
		)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace,
					fabricLinksSpace, "routing_cost"),
				"Fabric Link routing cost which depends of the strategy configured.",
				[]string{"destination", "source", "id"}, nil,
			), prometheus.GaugeValue,
			fabricLinks.Data[i].Cost,
			fabricLinks.Data[i].DestRouter.Name,
			fabricLinks.Data[i].SourceRouter.Name,
			fabricLinks.Data[i].ID,
		)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace,
					fabricLinksSpace, "routing_static_cost"),
				"Fabric Link routing static cost configured.",
				[]string{"destination", "source", "id"}, nil,
			), prometheus.GaugeValue,
			fabricLinks.Data[i].StaticCost,
			fabricLinks.Data[i].DestRouter.Name,
			fabricLinks.Data[i].SourceRouter.Name,
			fabricLinks.Data[i].ID,
		)
	}

	return nil
}

// RunFabricLinks implements this command
func (o *LoginOptions) RunFabricLinks() (FabricLinks, error) {
	var (
		limit                                     = 50
		offset                                    = 0
		fabricLinksStructTotal, fabricLinksStruct FabricLinks
		json                                      = jsoniter.ConfigCompatibleWithStandardLibrary
	)

	jsonBytes, err := controllerAPICall(o, "fabric", "/links", limit, offset)
	if err != nil {
		return fabricLinksStructTotal, err
	}

	err = json.Unmarshal(jsonBytes, &fabricLinksStruct)
	if err != nil {
		return fabricLinksStructTotal, err
	}

	fabricLinksStructTotal.Data = append(fabricLinksStructTotal.Data, fabricLinksStruct.Data...)

	totalFabricLinksCount := fabricLinksStruct.Meta.Pagination.TotalCount
	level.Debug(o.Logger).Log("msg", "Total Ziti Fabric Links found", "count", totalFabricLinksCount)

	for offset+limit < totalFabricLinksCount {
		offset += limit

		jsonBytes, err := controllerAPICall(o, "fabric", "/links", limit, offset)
		if err != nil {
			return fabricLinksStructTotal, err
		}

		err = json.Unmarshal(jsonBytes, &fabricLinksStruct)
		if err != nil {
			return fabricLinksStructTotal, err
		}

		fabricLinksStructTotal.Data = append(fabricLinksStructTotal.Data, fabricLinksStruct.Data...)
	}

	return fabricLinksStructTotal, err
}
