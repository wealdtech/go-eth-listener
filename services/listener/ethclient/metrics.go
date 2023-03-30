// Copyright © 2023 Weald Technology Limited.
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

package ethclient

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/wealdtech/go-eth-listener/services/metrics"
)

var metricsNamespace = "eth_listener"

var failuresMetric prometheus.Counter

func registerMetrics(_ context.Context, monitor metrics.Service) error {
	if failuresMetric != nil {
		// Already registered.
		return nil
	}
	if monitor == nil {
		// No monitor.
		return nil
	}
	if monitor.Presenter() == "prometheus" {
		return registerPrometheusMetrics()
	}
	return nil
}

func registerPrometheusMetrics() error {
	failuresMetric = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: metricsNamespace,
		Subsystem: "ethclient",
		Name:      "failures_total",
		Help:      "The number of failures.",
	})
	return prometheus.Register(failuresMetric)
}

func monitorFailure() {
	if failuresMetric != nil {
		failuresMetric.Inc()
	}
}
