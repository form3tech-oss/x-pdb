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

package pdb

import (
	"context"
	"time"

	xpdbv1alpha1 "github.com/form3tech-oss/x-pdb/api/v1alpha1"
	"github.com/form3tech-oss/x-pdb/internal/converters"
	stateclient "github.com/form3tech-oss/x-pdb/internal/state/client"
	statepb "github.com/form3tech-oss/x-pdb/pkg/protos/state"
	"github.com/go-logr/logr"
	"github.com/sourcegraph/conc/pool"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	remoteGetStateTimeout = 2 * time.Second
)

// Service implements the business logic
// to make a decision whether or not a Pod can
// be disrupted. It talks to external endpoints
// to make that decision.
type Service struct {
	logger          logr.Logger
	client          client.Client
	reader          client.Reader
	leaseNamespace  string
	scaleFinder     *ScaleFinder
	stateClientPool *stateclient.ClientPool
	remoteEndpoints []string
}

// NewService returns a new Service.
func NewService(
	logger logr.Logger,
	cli client.Client,
	rdr client.Reader,
	scaleFinder *ScaleFinder,
	stateClientPool *stateclient.ClientPool,
	leaseNamespace string,
	remoteEndpoints []string) *Service {
	return &Service{
		logger:          logger,
		client:          cli,
		reader:          rdr,
		leaseNamespace:  leaseNamespace,
		scaleFinder:     scaleFinder,
		stateClientPool: stateClientPool,
		remoteEndpoints: remoteEndpoints,
	}
}

// GetPodCounts returns the number of desired/actual healthy pods for the supplied namespace and selector.
func (s *Service) GetPodCounts(ctx context.Context, namespace string, selector *metav1.LabelSelector) (expectedCount, healthy int32, err error) {
	pods, err := s.getPodsMatchingSelector(ctx, namespace, selector)
	if err != nil {
		return expectedCount, healthy, err
	}

	expectedCount, _, err = s.scaleFinder.FindExpectedScale(ctx, pods)
	if err != nil {
		return expectedCount, healthy, err
	}

	healthy = countHealthyPods(pods)
	s.logger.V(1).Info("get-pod-counts", "expectedCount", expectedCount, "healthy", healthy)
	return expectedCount, healthy, nil
}

// CanXPdbBeDisrupted looks up both local and remote pods and calculates if a disruption would be acceptable.
// Note: You should use .Lock()/.Unlock() before using this func to ensure no other clusters are able to
// evict pods while we make the calculation and return the response back to the kube-apiserver.
//
// Note: It is still possible for pods to become unready or die unexpectedly while we do this calculation, hence
// the PDB would be disrupted.
func (s *Service) CanPodBeDisrupted(ctx context.Context, candidatePod *corev1.Pod, xpdb *xpdbv1alpha1.XPodDisruptionBudget) (bool, error) {
	var remoteExpectedCount, remoteHealthy int32
	var err error
	if len(s.remoteEndpoints) > 0 {
		remoteExpectedCount, remoteHealthy, err = s.getRemotePodCounts(ctx, xpdb.Namespace, &xpdb.Spec.Selector)
		if err != nil {
			s.logger.Error(err, "error getting remote pod counts", "namespace", xpdb.Namespace, "name", xpdb.Name)
			return false, err
		}
	}

	localExpectedCount, localHealthy, err := s.GetPodCounts(ctx, xpdb.Namespace, &xpdb.Spec.Selector)
	if err != nil {
		s.logger.Error(err, "error getting local pod counts", "namespace", xpdb.Namespace, "name", xpdb.Name)
		return false, err
	}

	totalHealthy := remoteHealthy + localHealthy
	totalExpectedCount := remoteExpectedCount + localExpectedCount
	s.logger.Info("xpdb aggregated remote state",
		"name", xpdb.Name,
		"namespace", xpdb.Namespace,
		"totalHealthy", totalHealthy,
		"totalExpectedCount", totalExpectedCount,
		"localHealthy", localHealthy,
		"localExpectedCount", localExpectedCount)

	return s.disruptionAllowed(xpdb, candidatePod, totalExpectedCount, totalHealthy)
}

// GetXPdbsForPod returns all XPDBs matching the particular pod.
func (s *Service) GetXPdbsForPod(ctx context.Context, pod *corev1.Pod) ([]*xpdbv1alpha1.XPodDisruptionBudget, error) {
	var items xpdbv1alpha1.XPodDisruptionBudgetList
	err := s.reader.List(ctx, &items, client.InNamespace(pod.Namespace))
	if err != nil {
		return nil, err
	}

	xpbds := make([]*xpdbv1alpha1.XPodDisruptionBudget, 0)
	for i := range items.Items {
		selector, err := metav1.LabelSelectorAsSelector(&items.Items[i].Spec.Selector)
		if err != nil {
			return nil, err
		}
		if !selector.Matches(labels.Set(pod.Labels)) {
			continue
		}
		xpbds = append(xpbds, &items.Items[i])
	}

	return xpbds, nil
}

func (s *Service) getPodsMatchingSelector(ctx context.Context, namespace string, selector *metav1.LabelSelector) ([]*corev1.Pod, error) {
	sel, err := metav1.LabelSelectorAsSelector(selector)
	if err != nil {
		return []*corev1.Pod{}, err
	}

	var podList corev1.PodList
	err = s.reader.List(ctx, &podList, client.InNamespace(namespace), client.MatchingLabelsSelector{Selector: sel})
	if err != nil {
		return []*corev1.Pod{}, err
	}

	pods := make([]*corev1.Pod, len(podList.Items))
	for i := range podList.Items {
		pods[i] = &podList.Items[i]
	}

	return pods, nil
}

func (s *Service) disruptionAllowed(xpdb *xpdbv1alpha1.XPodDisruptionBudget, candidatePod *corev1.Pod, expectedCount, healthyCount int32) (bool, error) {
	if xpdb == nil {
		return true, nil
	}
	var desiredHealthy int32
	if xpdb.Spec.MaxUnavailable != nil {
		maxUnavailable, err := intstr.GetScaledValueFromIntOrPercent(xpdb.Spec.MaxUnavailable, int(expectedCount), true)
		if err != nil {
			return false, err
		}
		desiredHealthy = expectedCount - int32(maxUnavailable)
	} else if xpdb.Spec.MinAvailable != nil {
		if xpdb.Spec.MinAvailable.Type == intstr.Int {
			desiredHealthy = xpdb.Spec.MinAvailable.IntVal
		} else if xpdb.Spec.MinAvailable.Type == intstr.String {
			var minAvailable int
			minAvailable, err := intstr.GetScaledValueFromIntOrPercent(xpdb.Spec.MinAvailable, int(expectedCount), true)
			if err != nil {
				return false, err
			}
			desiredHealthy = int32(minAvailable)
		}
	}

	// In the case the pod being deleted/evicted is not ready
	// we should account it as healthy to ensure it can be
	// deleted / evicted.
	candidatePodReady := IsPodReady(candidatePod)
	var healthyCompensation int32
	if !candidatePodReady {
		healthyCompensation = 1
	}

	allowed := healthyCount+healthyCompensation-1 >= desiredHealthy

	s.logger.Info("xpdb evaluation verdict",
		"xpdbName", xpdb.Name,
		"xpdbNamespace", xpdb.Namespace,
		"disruptionAllowed", allowed,
		"expectedCount", expectedCount,
		"healthyCount", healthyCount,
		"desiredHealthy", desiredHealthy,
		"podName", candidatePod.Name,
		"podReady", candidatePodReady)

	return allowed, nil
}

func (s *Service) getRemotePodCounts(ctx context.Context, namespace string, selector *metav1.LabelSelector) (remoteDesiredHealthy, remoteHealthy int32, err error) {
	if len(s.remoteEndpoints) == 0 {
		return
	}

	req := &statepb.GetStateRequest{
		Namespace:     namespace,
		LabelSelector: converters.ConvertLabelSelectorToState(selector),
	}

	p := pool.NewWithResults[*statepb.GetStateResponse]().
		WithErrors().
		WithMaxGoroutines(len(s.remoteEndpoints)).
		WithContext(ctx)

	for _, e := range s.remoteEndpoints {
		p.Go(func(ctx context.Context) (*statepb.GetStateResponse, error) {
			cli, err := s.stateClientPool.Get(e)
			if err != nil {
				return nil, err
			}

			cctx, cancel := context.WithTimeout(ctx, remoteGetStateTimeout)
			defer cancel()

			res, err := cli.GetState(cctx, req)
			if err != nil {
				s.logger.Error(err, "error obtaining remote state", "endpoint", e)
				return nil, err
			}
			s.logger.Info("xpdb remote count",
				"endpoint", e,
				"namespace", namespace,
				"selector", selector.String(),
				"desiredhealthy", res.DesiredHealthy,
				"healthy", res.Healthy,
			)
			return res, nil
		})
	}

	results, err := p.Wait()
	if err != nil {
		return 0, 0, err
	}

	for _, res := range results {
		remoteDesiredHealthy += res.DesiredHealthy
		remoteHealthy += res.Healthy
	}
	return
}

func countHealthyPods(pods []*corev1.Pod) (currentHealthy int32) {
	for _, pod := range pods {
		// Pod is being deleted.
		if pod.DeletionTimestamp != nil {
			continue
		}
		// Pod is expected to be deleted soon.
		if IsPodReady(pod) {
			currentHealthy++
		}
	}
	return
}
