// Copyright 2025 Tobias Hintze
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"context"
	"math"

	"github.com/paraopsde/go-x/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/thz/dimplex-modbus-exporter/pkg/modbus"
	"go.uber.org/zap"
)

type Collector struct {
	ctx              context.Context
	scrapes          int64
	modbusErrorCount int64
	modbus           *modbus.ModbusClient
}

type metricDefinition struct {
	register uint16
	name     string
	desc     string
	promDesc *prometheus.Desc
}

var (
	scrapeCountMetricDesc      *prometheus.Desc
	modbusErrorCountMetricDesc *prometheus.Desc

	// https://dimplex.atlassian.net/wiki/spaces/DW/pages/2873393288/NWPM+Modbus+TCP
	metricDefinitions []metricDefinition = []metricDefinition{
		{
			register: 1,
			name:     "temperature_outdoors",
			desc:     "temperature, outdoor sensor",
		},
		{
			register: 2,
			name:     "temperature_return_heating",
			desc:     "temperature, heating return",
		},
		{
			register: 53,
			name:     "temperature_heating_return_desired",
			desc:     "desired temperature, heating return",
		},
		{
			register: 3,
			name:     "temperature_domestic_hot_water",
			desc:     "temperature, domestic hot water",
		},
		{
			register: 58,
			name:     "temperature_domestic_hot_water_desired",
			desc:     "desired temperature, domestic hot water",
		},
		{
			register: 5,
			name:     "temperature_flow",
			desc:     "temperature, flow",
		},
		{
			register: 6,
			name:     "pressure_low",
			desc:     "pressure, low",
		},
		{
			register: 8,
			name:     "pressure_high",
			desc:     "pressure, high",
		},
		{
			register: 103,
			name:     "operating_status",
			desc:     "status message code: 2=heating 4=hot_water 10=defrost",
		},
	}
)

func init() {
	scrapeCountMetricDesc = prometheus.NewDesc("scrape_count", "number of times the exporter has been scraped", nil, nil)
	modbusErrorCountMetricDesc = prometheus.NewDesc("modbus_error_count", "number of times the modbus client observes an error", nil, nil)

	for idx := range metricDefinitions {
		mdef := &metricDefinitions[idx]
		mdef.promDesc = prometheus.NewDesc(mdef.name, mdef.desc, nil, nil)
	}
}

func lowPressure(nd float32) float32 {
	return ((((nd * 10) - 100) * 173) / 800) / 10
}
func highPressure(hd float32) float32 {
	return ((((hd * 10) - 100) * 345) / 800) / 10
}

func New(modbus *modbus.ModbusClient, ctx context.Context) *Collector {
	return &Collector{
		modbus: modbus,
		ctx:    ctx,
	}
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	log := util.CtxLogOrPanic(c.ctx)

	c.scrapes += 1

	ch <- prometheus.MustNewConstMetric(scrapeCountMetricDesc, prometheus.CounterValue, float64(c.scrapes))

	for idx := range metricDefinitions {
		mdef := &metricDefinitions[idx]
		switch mdef.name {
		case "operating_status":
			val, err := c.modbus.RequestInt16(c.ctx, mdef.register)
			if err != nil {
				log.Warn("failed to request data", zap.Error(err))
				c.modbusErrorCount += 1
				continue
			}
			ch <- prometheus.MustNewConstMetric(mdef.promDesc, prometheus.GaugeValue, float64(val))
		case "pressure_low":
			val, err := c.modbus.RequestPseudoFloat16(c.ctx, mdef.register)
			if err != nil {
				log.Warn("failed to request data", zap.Error(err))
				c.modbusErrorCount += 1
				continue
			}
			ch <- prometheus.MustNewConstMetric(mdef.promDesc, prometheus.GaugeValue, fixed(lowPressure(val), 1))
		case "pressure_high":
			val, err := c.modbus.RequestPseudoFloat16(c.ctx, mdef.register)
			if err != nil {
				log.Warn("failed to request data", zap.Error(err))
				c.modbusErrorCount += 1
				continue
			}
			ch <- prometheus.MustNewConstMetric(mdef.promDesc, prometheus.GaugeValue, fixed(highPressure(val), 1))
		default:
			val, err := c.modbus.RequestPseudoFloat16(c.ctx, mdef.register)
			if err != nil {
				log.Warn("failed to request data", zap.String("metric-definition-name", mdef.name), zap.Error(err))
				c.modbusErrorCount += 1
				continue
			}
			ch <- prometheus.MustNewConstMetric(mdef.promDesc, prometheus.GaugeValue, fixed(val, 2))
		}
	}
	ch <- prometheus.MustNewConstMetric(modbusErrorCountMetricDesc, prometheus.CounterValue, float64(c.modbusErrorCount))
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeCountMetricDesc
	for idx := range metricDefinitions {
		mdef := &metricDefinitions[idx]
		ch <- mdef.promDesc
	}
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}
func fixed(num float32, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(float64(num)*output)) / output
}
