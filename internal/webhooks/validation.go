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

package webhooks

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/form3tech-oss/x-pdb/api/v1alpha1"
	xpdbv1alpha1 "github.com/form3tech-oss/x-pdb/api/v1alpha1"
	"github.com/form3tech-oss/x-pdb/internal/disruptionprobe"
	"github.com/form3tech-oss/x-pdb/internal/lock"
	"github.com/form3tech-oss/x-pdb/internal/metrics"
	"github.com/form3tech-oss/x-pdb/internal/pdb"
	"github.com/form3tech-oss/x-pdb/internal/preactivities"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	// NodeUnreachablePodReason
	// To avoid importing "k8s.io/kubernetes"
	// We're adding the NodeUnreachablePodReason
	// as a constant in this project.
	// This constant would be available at:
	// "k8s.io/kubernetes/pkg/util/node"
	NodeUnreachablePodReason = "NodeLost"

	PendingActivitiesDisruptionNotAllowedMessage = "Cannot disrupt pod has it has pending disruption pre-activities."
	XPDBDisruptionBudgetNotAllowedMessage        = "Cannot disrupt pod as it would violate the pod's xpdb disruption budget."
	XPDBDisruptionBudgetErrorMessage             = "Cannot disrupt pod as there was an error evaluating pod's xpdb disruption budget"
	XPDBDisruptionProbeNotAllowedMessage         = "Cannot disrupt pod as the pod's xpdb disruption probe didn't allow it."
	XPDBDisruptionProbeErrorMessage              = "Cannot disrupt pod as there was an error calling pod's xpdb disruption probe"
)

// PodValidationWebhook implements a admission webhook server that is used
// to intercept admission requests for Pod resources.
type PodValidationWebhook struct {
	decoder                admission.Decoder
	logger                 logr.Logger
	client                 client.Client
	recorder               record.EventRecorder
	pdbService             *pdb.Service
	lockService            *lock.Service
	disruptionProbeService *disruptionprobe.Service
	preactivitiesService   *preactivities.Service
	clusterID              string
	podID                  string
	dryRun                 bool
}

// NewPodValidationWebhook creates a new Pod validation webhook instance.
func NewPodValidationWebhook(
	client client.Client,
	logger logr.Logger,
	decoder admission.Decoder,
	recorder record.EventRecorder,
	clusterID string,
	podID string,
	dryRun bool,
	pdbService *pdb.Service,
	lockService *lock.Service,
	disruptionProbeService *disruptionprobe.Service,
	preactivitiesService *preactivities.Service,
) *PodValidationWebhook {
	return &PodValidationWebhook{
		client:                 client,
		logger:                 logger,
		decoder:                decoder,
		recorder:               recorder,
		pdbService:             pdbService,
		lockService:            lockService,
		disruptionProbeService: disruptionProbeService,
		preactivitiesService:   preactivitiesService,
		clusterID:              clusterID,
		podID:                  podID,
		dryRun:                 dryRun,
	}
}

// Handle decodes the admission request and verifies if the Pod can be disrupted.
// If not, it responds with a HTTP status code 429, similar how evictions are handled
// if a PDB is blocking the eviction.
func (h PodValidationWebhook) Handle(ctx context.Context, request admission.Request) admission.Response {
	pod, err := h.decodePod(ctx, request)
	if err != nil {
		return h.admissionResponse(true, "", nil)
	}

	// If pod was already deleted lets ignore this validation request
	if pod.ObjectMeta.DeletionTimestamp != nil {
		return h.admissionResponse(true, "", nil)
	}

	logger := h.logger.WithValues("pod", pod.Name, "namespace", pod.Namespace)

	logger.V(2).Info("received webhook request",
		"kind", request.RequestKind.Kind,
		"resource", request.RequestResource,
		"subresource", request.RequestSubResource,
		"operation", request.Operation,
		"userInfo.username", request.UserInfo.Username,
		"userInfo.extra", request.UserInfo.Extra,
		"dryRun", request.DryRun,
		"pod.status.phase", pod.Status.Phase,
		"pod.status.reason", pod.Status.Reason)

	// If pod is being evicted by tainteviction controller
	// then we must ignore this request, as it is a involuntary disruption.
	// see: https://github.com/kubernetes/kubernetes/blob/ccda2d6fd413d02b9df2dd76ca643dfb42d62aaf/pkg/controller/tainteviction/taint_eviction.go#L131-L150
	//
	// The DisruptionTarget condition indicates that the pod is about to be terminated due to a
	// disruption (such as preemption, eviction API or garbage-collection).
	if cond := getDisruptionTargetCondition(pod); cond != nil {
		logger.Info("ignoring pod: deleted due to disruption",
			"type", cond.Type,
			"status", cond.Status,
			"reason", cond.Reason,
			"message", cond.Message,
		)
		return h.admissionResponse(true, "", nil)
	}

	// relevant for pre-1.29 behaviour:
	// node-lifecycle-controller implemented the Pod eviction/deletion on its own
	// post-1.29 eviction is triggered via taint and deleted via tainteviction controller
	if podHasNodeLostReason(pod) {
		logger.Info("ignoring pod: has status.reason=NodeLost")
		return h.admissionResponse(true, "", nil)
	}

	// Handle preactivities feature
	canPodBeDisrupted, err := h.preactivitiesService.CanPodBeDisrupted(ctx, pod)
	if err != nil {
		return h.handleError(ctx, logger, nil, "", err, fmt.Sprintf("error verifying if pod had pending pre-activities: %s", err.Error()))
	}
	if !canPodBeDisrupted {
		return h.handleNotAllowedDisruption(ctx, logger, request, nil, pod, "", PendingActivitiesDisruptionNotAllowedMessage)
	}

	// Handle Multi-cluster pdb feature
	xpdbs, err := h.pdbService.GetXPdbsForPod(ctx, pod)
	if err != nil {
		return h.admissionResponse(false, fmt.Sprintf("could not get xpdbs for pod: %s", err.Error()), nil)
	}
	if len(xpdbs) == 0 {
		return h.admissionResponse(true, "", nil)
	}

	if len(xpdbs) > 1 {
		var xpdbNames []string
		for _, xpdb := range xpdbs {
			xpdbNames = append(xpdbNames, xpdb.Name)
		}
		h.recorder.Eventf(
			pod,
			corev1.EventTypeWarning,
			string(xpdbv1alpha1.XPDBEventReasonInvalidConfiguration),
			"invalid configuration: pod matches multiple XPDBs: %s",
			strings.Join(xpdbNames, ", "),
		)
		metrics.ObservePodMatchingMultipleXPDBs(pod.Namespace)
		logger.Error(nil, "pod matches multiple xpdbs")

		// When a pod matches multiple PDBs then this is a invalid configuration.
		// We want to match the same API behavior as kubernetes, that is to return a
		// HTTP 500 if that is the case.
		// see upstream Kubernetes docs:
		// https://kubernetes.io/docs/concepts/scheduling-eviction/api-eviction/#how-api-initiated-eviction-works
		return h.admissionResponse(false,
			"Cannot disrupt pod as it matched multiple xpdbs.",
			ptr.To(int32(http.StatusInternalServerError)))
	}

	xpdb := xpdbs[0]

	if xpdb.Spec.Suspend != nil && *xpdb.Spec.Suspend {
		return h.admissionResponse(true, "", nil)
	}

	leaseHolderIdentity := lock.CreateLeaseHolderIdentity(h.clusterID, h.podID, pod.Namespace, pod.Name)
	err = h.lockService.Lock(ctx, leaseHolderIdentity, xpdb.Namespace, &xpdb.Spec.Selector)
	if err != nil {
		logger.Error(
			err,
			"could obtain xpdb lock",
			"leaseHolderIdentity", leaseHolderIdentity,
		)
		metrics.ObserveLockError(pod.Namespace)
		return h.admissionResponse(
			false,
			"Cannot disrupt pod because xpdb couldn't obtain lock",
			nil)
	}

	canBeDisrupted, err := h.pdbService.CanPodBeDisrupted(ctx, pod, xpdb)
	if err != nil {
		return h.handleError(ctx, logger, xpdb, XPDBDisruptionBudgetErrorMessage, err, leaseHolderIdentity)
	}
	if !canBeDisrupted {
		return h.handleNotAllowedDisruption(ctx, logger, request, xpdb, pod, leaseHolderIdentity, XPDBDisruptionBudgetNotAllowedMessage)
	}

	// Handle disruption probe feature
	if xpdb.Spec.Probe != nil && (xpdb.Spec.Probe.Enabled == nil || *xpdb.Spec.Probe.Enabled) {
		canBeDisrupted, err := h.disruptionProbeService.CanPodBeDisrupted(ctx, pod, xpdb)
		if err != nil {
			return h.handleError(ctx, logger, xpdb, XPDBDisruptionProbeErrorMessage, err, leaseHolderIdentity)
		}
		if !canBeDisrupted {
			return h.handleNotAllowedDisruption(ctx, logger, request, xpdb, pod, leaseHolderIdentity, XPDBDisruptionProbeNotAllowedMessage)
		}
	}

	h.recorder.Eventf(xpdb, corev1.EventTypeNormal, string(xpdbv1alpha1.XPDBEventReasonAccepted), "attempted eviction of %s", pod.Name)

	// We leave the pdb in a locked state because admission-control response
	// is still in flight and the (potential) eviction hasn't been processed
	// yet by the kube-apiserver.
	return h.admissionResponse(true, "", nil)
}

func (h PodValidationWebhook) admissionResponse(allowed bool, message string, errorCode *int32) admission.Response {
	if h.dryRun {
		return admission.ValidationResponse(allowed, message)
	}
	if errorCode != nil {
		return admission.Errored(*errorCode, errors.New(message))
	}
	return admission.ValidationResponse(allowed, message)
}

func (h PodValidationWebhook) decodePod(ctx context.Context, request admission.Request) (*corev1.Pod, error) {
	var pod corev1.Pod
	var err error
	if len(request.OldObject.Raw) > 0 {
		err = h.decoder.DecodeRaw(request.OldObject, &pod)
	} else {
		err = h.client.Get(ctx, types.NamespacedName{
			Namespace: request.Namespace,
			Name:      request.Name,
		}, &pod)
	}
	return &pod, err
}

func (h PodValidationWebhook) handleError(
	ctx context.Context,
	logger logr.Logger,
	xpdb *v1alpha1.XPodDisruptionBudget,
	errorDescription string,
	err error,
	leaseHolderIdentity string,
) admission.Response {
	if xpdb != nil {
		logger.Error(err, "pod disruption check returned an error")
		unlockErr := h.lockService.Unlock(ctx, leaseHolderIdentity, xpdb.Namespace, &xpdb.Spec.Selector)
		if unlockErr != nil {
			logger.Error(unlockErr, "unable to release xpdb lock")
		}
	}

	return h.admissionResponse(
		false,
		fmt.Sprintf("%s: %s", errorDescription, err.Error()),
		nil)
}

func (h PodValidationWebhook) handleNotAllowedDisruption(
	ctx context.Context,
	logger logr.Logger,
	request admission.Request,
	xpdb *v1alpha1.XPodDisruptionBudget,
	pod *corev1.Pod,
	leaseHolderIdentity string,
	admissionResponseMessage string,
) admission.Response {
	if xpdb != nil {
		h.recorder.Eventf(xpdb, corev1.EventTypeNormal, string(xpdbv1alpha1.XPDBEventReasonBlocked), "attempted eviction of %s", pod.Name)
		metrics.ObserveEvictionRejected(xpdb.Namespace, request.Resource.Resource, request.SubResource, string(request.Operation))

		err := h.lockService.Unlock(ctx, leaseHolderIdentity, xpdb.Namespace, &xpdb.Spec.Selector)
		if err != nil {
			logger.Error(err, "unable to unlock xpdb")
		}
	}

	// We must return a HTTP 429 here so kubectl still behaves in the same way
	// as the regular PDB would.
	// see:
	// https://kubernetes.io/docs/concepts/scheduling-eviction/api-eviction/#how-api-initiated-eviction-works
	// https://github.com/kubernetes/kubectl/blob/acf4a09f2daede8fdbf65514ade9426db0367ed3/pkg/drain/drain.go#L318-L320
	return h.admissionResponse(
		false,
		admissionResponseMessage,
		ptr.To(int32(http.StatusTooManyRequests)))
}

func getDisruptionTargetCondition(po *corev1.Pod) *corev1.PodCondition {
	for i := range po.Status.Conditions {
		cond := po.Status.Conditions[i]
		if cond.Type == corev1.DisruptionTarget && cond.Status == corev1.ConditionTrue {
			return &cond
		}
	}
	return nil
}

func podHasNodeLostReason(po *corev1.Pod) bool {
	return po.Status.Reason == NodeUnreachablePodReason
}
