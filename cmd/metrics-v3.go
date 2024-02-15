// Copyright (c) 2015-2024 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/minio/minio/internal/logger"
	"github.com/minio/minio/internal/mcontext"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	poolIndex                                    = "pool_index"
	apiMetricNamespace           MetricNamespace = "minio_api"
	apiObjectMetricsNamespace    MetricNamespace = "minio_obj"
	processMetricsNamespace      MetricNamespace = "minio_process"
	replicationV3MetricNamespace MetricNamespace = "minio_replication"
)

var (
	// processCollector *minioProcessCollector
	//
	//	processMetricsGroup   []*MetricsGroup
	nodeReplicationMetricsGroupV3 []*MetricsGroup
	nodeReplCollectorV3           *minioNodeReplCollectorV3
)

func init() {
	nodeReplicationMetricsGroupV3 = []*MetricsGroup{
		getReplicationNodeMetrics(MetricsGroupOpts{dependGlobalObjectAPI: true, dependBucketTargetSys: true}, replicationV3MetricNamespace),
	}

	nodeReplCollectorV3 = newMinioReplCollectorNodeV3(nodeReplicationMetricsGroupV3)
}

type promLogger struct {
}

func (p promLogger) Println(v ...interface{}) {
	s := make([]string, 0, len(v))
	for _, val := range v {
		s = append(s, fmt.Sprintf("%v", val))
	}
	err := fmt.Errorf("metrics handler error: %v", strings.Join(s, " "))
	logger.LogIf(GlobalContext, err)
}

type v3MetricsHandler struct {
	registry *prometheus.Registry
	opts     promhttp.HandlerOpts
	authFn   func(http.Handler) http.Handler
}

func newV3MetricsHandler(authType prometheusAuthType) *v3MetricsHandler {
	registry := prometheus.NewRegistry()
	authFn := AuthMiddleware
	if prometheusAuthType(authType) == prometheusPublic {
		authFn = NoAuthMiddleware
	}
	return &v3MetricsHandler{
		registry: registry,
		opts: promhttp.HandlerOpts{
			ErrorLog:            promLogger{},
			ErrorHandling:       promhttp.HTTPErrorOnError,
			Registry:            registry,
			MaxRequestsInFlight: 2,
		},
		authFn: authFn,
	}
}

// handlerFor returns a http.Handler for the given prometheus.Collector(s) under
// the given handlerName.
func (h *v3MetricsHandler) handlerFor(handlerName string, cs ...prometheus.Collector) http.Handler {
	subRegistry := prometheus.NewRegistry()
	logger.CriticalIf(GlobalContext, h.registry.Register(subRegistry))
	for _, c := range cs {
		logger.CriticalIf(GlobalContext, subRegistry.Register(c))
	}

	promHandler := promhttp.HandlerFor(subRegistry, h.opts)

	// Add tracing to the prom. handler
	tracedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		tc, ok := r.Context().Value(mcontext.ContextTraceKey).(*mcontext.TraceCtxt)
		if ok {
			tc.FuncName = handlerName
			tc.ResponseRecorder.LogErrBody = true
		}

		promHandler.ServeHTTP(w, r)
	})

	// Add authentication
	return h.authFn(tracedHandler)

}

// minioNodeReplCollectorV3 is the Custom Collector
type minioNodeReplCollectorV3 struct {
	metricsGroups []*MetricsGroup
	desc          *prometheus.Desc
}

// Describe sends the super-set of all possible descriptors of metrics
func (c *minioNodeReplCollectorV3) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.desc
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *minioNodeReplCollectorV3) Collect(ch chan<- prometheus.Metric) {
	// Expose MinIO's version information
	minioVersionInfo.WithLabelValues(Version, CommitID).Set(1.0)

	populateAndPublish(c.metricsGroups, func(metric Metric) bool {
		labels, values := getOrderedLabelValueArrays(metric.VariableLabels)
		values = append(values, globalLocalNodeName)
		labels = append(labels, serverName)

		if metric.Description.Type == histogramMetric {
			if metric.Histogram == nil {
				return true
			}
			for k, v := range metric.Histogram {
				labels = append(labels, metric.HistogramBucketLabel)
				values = append(values, k)
				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc(
						prometheus.BuildFQName(string(metric.Description.Namespace),
							string(metric.Description.Subsystem),
							string(metric.Description.Name)),
						metric.Description.Help,
						labels,
						metric.StaticLabels,
					),
					prometheus.GaugeValue,
					float64(v),
					values...)
			}
			return true
		}

		metricType := prometheus.GaugeValue
		if metric.Description.Type == counterMetric {
			metricType = prometheus.CounterValue
		}
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(string(metric.Description.Namespace),
					string(metric.Description.Subsystem),
					string(metric.Description.Name)),
				metric.Description.Help,
				labels,
				metric.StaticLabels,
			),
			metricType,
			metric.Value,
			values...)
		return true
	})
}

// newMinioReplCollectorNodeV3 describes the collector
// and returns reference of minioCollector for version 3
// It creates the Prometheus Description which is used
// to define Metric and  help string
func newMinioReplCollectorNodeV3(metricsGroups []*MetricsGroup) *minioNodeReplCollectorV3 {
	return &minioNodeReplCollectorV3{
		metricsGroups: metricsGroups,
		desc:          prometheus.NewDesc("minio_stats", "Statistics exposed by MinIO server per node", nil, nil),
	}
}
