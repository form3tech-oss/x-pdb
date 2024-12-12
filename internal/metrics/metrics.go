/*
Copyright 2024 Form3.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package metrics

import (
	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

// APICallResult represents an API call result.
type APICallResult string

const (
	// APICallResultSuccess represents a successful operation.
	APICallResultSuccess = "success"
	// APICallResultError represents a failed operation.
	APICallResultError = "error"
)

const (
	xpdbNamespace    = "xpdb"
	labelNamespace   = "namespace"
	labelResult      = "result"
	labelAPICall     = "api_call"
	labelResource    = "resource"
	labelSubresource = "subresource"
	labelOperation   = "operation"
)

var (
	evictionRejectedCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: xpdbNamespace,
		Name:      "pod_eviction_rejected",
		Help:      "Counter that represents the number of eviction which have been rejected through xpdb",
	}, []string{labelNamespace, labelResource, labelSubresource, labelOperation})

	podMatchingMultipleXPDBs = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: xpdbNamespace,
		Name:      "pod_matches_multiple_xpdbs",
		Help: "A eviction attempt for a pod has been observed which matches multiple XPDBs. " +
			"This is a invalid configuration and must be fixed",
	}, []string{labelNamespace})

	lockErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: xpdbNamespace,
		Name:      "lock_errors",
		Help:      "Counter that represents the number of errors when obtaining locks for xpdb.",
	}, []string{labelNamespace})

	GrpcClientMetrics = grpcprom.NewClientMetrics(
		grpcprom.WithClientHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets([]float64{0.01, 0.1, 0.3, 0.6, 1, 3, 5}),
		),
	)
)

// ObserveEvictionRejected increments the eviction rejected counter.
func ObserveEvictionRejected(namespace, resource, subresource, operation string) {
	evictionRejectedCounter.WithLabelValues(namespace, resource, subresource, operation).Inc()
}

// ObservePodMatchingMultipleXPDBs increments the pod matching multiple PDBs counter.
func ObservePodMatchingMultipleXPDBs(namespace string) {
	podMatchingMultipleXPDBs.WithLabelValues(namespace).Inc()
}

// ObserveLockError increments the locking errors counter.
func ObserveLockError(namespace string) {
	lockErrors.WithLabelValues(namespace).Inc()
}

func init() {
	metrics.Registry.MustRegister(podMatchingMultipleXPDBs)
	metrics.Registry.MustRegister(evictionRejectedCounter)
	metrics.Registry.MustRegister(lockErrors)
	metrics.Registry.MustRegister(GrpcClientMetrics)
}
