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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// XPodDisruptionBudgetSpec defines the desired state of XPodDisruptionBudget.
type XPodDisruptionBudgetSpec struct {
	// A XPDB can be suspended, hence allowing pod deletion of pods matching this pdb.
	// This is intended to be used as a break-glass procedure
	// to allow engineers to take manual action on Pods which
	// must not be disrupted.
	// The suspension is configured on a per-cluster basis
	// and does only affect local pods. I.e. other clusters
	// that run x-pdb will not be able to evict pods if there
	// isn't enough disruption budget available globally.
	// To allow disruptions in other clusters one must set the `suspend` field to true
	// in those clusters as well.
	Suspend *bool `json:"suspend,omitempty"`

	// An eviction is allowed if at least "minAvailable" pods selected by
	// "selector" will still be available after the eviction, i.e. even in the
	// absence of the evicted pod.  So for example you can prevent all voluntary
	// evictions by specifying "100%".
	// +optional
	MinAvailable *intstr.IntOrString `json:"minAvailable,omitempty"`

	// Label query over pods whose evictions are managed by the disruption
	// budget.
	Selector metav1.LabelSelector `json:"selector,omitempty"`

	// An eviction is allowed if at most "maxUnavailable" pods selected by
	// "selector" are unavailable after the eviction, i.e. even in absence of
	// the evicted pod. For example, one can prevent all voluntary evictions
	// by specifying 0. This is a mutually exclusive setting with "minAvailable".
	// +optional
	MaxUnavailable *intstr.IntOrString `json:"maxUnavailable,omitempty"`

	// XPDB allows workload owners to define a disruption probe endpoint.
	// It might be helpful to probe internal state of some workloads like databases
	// to verify wether an eviction can happen or not.
	// Database Raft groups might become unavailable if a given pod
	// is disrupted. In these cases workload owners might want to
	// block disruptions to happen, even if all pods are ready.
	// +optional
	Probe *XPodDisruptionBudgetProbeSpec `json:"probe,omitempty"`
}

// XPodDisruptionBudgetProbeSpec allows workload owners to define a disruption probe endpoint.
type XPodDisruptionBudgetProbeSpec struct {
	// Specifies if the x-pdb will perform a call to the disruption probe endpoint.
	// Its enabled by default. It can be manually disabled in the case the disruption probe
	// is unavailable and a disruption needs to be performed.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// The endpoint that x-pdb will call when evaluating if a given pod can be disrupted.
	// This is a grpc endpoint of a service that should implement the contract defined in
	// protos/disruptionprobe/disruptionprobe.proto.
	Endpoint string `json:"endpoint,omitempty"`
}

// XPodDisruptionBudgetStatus defines the observed state of XPodDisruptionBudget.
type XPodDisruptionBudgetStatus struct{}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// XPodDisruptionBudget is the Schema for the xpoddisruptionbudgets API.
type XPodDisruptionBudget struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   XPodDisruptionBudgetSpec   `json:"spec,omitempty"`
	Status XPodDisruptionBudgetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// XPodDisruptionBudgetList contains a list of XPodDisruptionBudget.
type XPodDisruptionBudgetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []XPodDisruptionBudget `json:"items"`
}

func init() {
	SchemeBuilder.Register(&XPodDisruptionBudget{}, &XPodDisruptionBudgetList{})
}
